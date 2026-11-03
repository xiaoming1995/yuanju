package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"
	"yuanju/internal/model"
	"yuanju/internal/repository"
	"yuanju/internal/service"
	"yuanju/pkg/bazi"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SavePersonalChart creates a fresh chart snapshot and switches the personal
// binding atomically. This endpoint is used by the miniapp's Mine setup flow.
func SavePersonalChart(c *gin.Context) {
	var input CalculateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "请检查本人出生资料"})
		return
	}
	birthDate := time.Date(input.Year, time.Month(input.Month), input.Day, 0, 0, 0, 0, time.UTC)
	if birthDate.Year() != input.Year || int(birthDate.Month()) != input.Month || birthDate.Day() != input.Day {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "出生日期无效"})
		return
	}
	if input.CalendarType == "" {
		input.CalendarType = "solar"
	}
	if input.CalendarType != "solar" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "本人命盘请填写公历出生日期"})
		return
	}
	name, err := normalizeChartDisplayName(input.DisplayName)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	result := bazi.Calculate(input.Year, input.Month, input.Day, input.Hour, input.Gender,
		input.IsEarlyZishi, input.Longitude, input.CalendarType, input.IsLeapMonth)
	userID := personalUserID(c)
	chart := &model.BaziChart{UserID: &userID, BirthYear: input.Year, BirthMonth: input.Month,
		BirthDay: input.Day, BirthHour: input.Hour, Gender: input.Gender, DisplayName: name,
		YearGan: result.YearGan, YearZhi: result.YearZhi, MonthGan: result.MonthGan, MonthZhi: result.MonthZhi,
		DayGan: result.DayGan, DayZhi: result.DayZhi, HourGan: result.HourGan, HourZhi: result.HourZhi,
		Wuxing: result.Wuxing, Dayun: result.Dayun, Yongshen: result.Yongshen, Jishen: result.Jishen,
		ChartHash: result.ChartHash, CalendarType: input.CalendarType, IsLeapMonth: input.IsLeapMonth}
	saved, err := repository.CreatePersonalChartWithChannel(chart, userActivityChannel(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "本人命盘保存失败，请稍后重试"})
		return
	}
	if snapshot, err := json.Marshal(result); err == nil {
		_ = repository.SaveChartResultJSON(saved.ID, snapshot)
	}
	c.JSON(http.StatusOK, gin.H{"chart": saved, "result": result})
}

func GetPersonalChartDetail(c *gin.Context) {
	chart, err := repository.GetPersonalChart(personalUserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取本人命盘失败"})
		return
	}
	if chart == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "请先配置本人命盘", "code": "personal_chart_unbound"})
		return
	}
	result, err := service.LoadOrCalculateResult(chart)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取本人命盘详情失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"chart": chart, "result": result})
}

func personalUserID(c *gin.Context) string {
	value, _ := c.Get("user_id")
	id, _ := value.(string)
	return id
}

func GetPersonalChart(c *gin.Context) {
	chart, err := repository.GetPersonalChart(personalUserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取本人命盘失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"chart": chart})
}

func SetPersonalChart(c *gin.Context) {
	var input struct {
		ChartID string `json:"chart_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择一份排盘"})
		return
	}
	if _, err := uuid.Parse(input.ChartID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "排盘ID无效"})
		return
	}
	if err := repository.SetPersonalChart(personalUserID(c), input.ChartID); err != nil {
		if errors.Is(err, repository.ErrPersonalChartNotOwned) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "绑定本人命盘失败"})
		return
	}
	GetPersonalChart(c)
}

func ClearPersonalChart(c *gin.Context) {
	if err := repository.ClearPersonalChart(personalUserID(c)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "解绑本人命盘失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"chart": nil})
}

func GetPersonalDay(c *gin.Context) {
	chart, err := repository.GetPersonalChart(personalUserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取本人命盘失败"})
		return
	}
	if chart == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "请先绑定本人命盘", "code": "personal_chart_unbound"})
		return
	}
	result, err := service.LoadOrCalculateResult(chart)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取命盘失败"})
		return
	}
	day, err := service.BuildPersonalDay(c.Query("date"), result)
	if errors.Is(err, service.ErrInvalidPersonalDayDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "计算每日干支失败"})
		return
	}
	day.ChartID, day.ChartName = chart.ID, chart.DisplayName
	if day.ChartName == "" {
		day.ChartName = "本人命盘"
	}
	c.JSON(http.StatusOK, gin.H{"data": day})
}

func GetCalendarDay(c *gin.Context) {
	day, err := service.BuildPersonalDay(c.Query("date"), nil)
	if errors.Is(err, service.ErrInvalidPersonalDayDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取黄历宜忌失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{
		"date": day.Date, "lunar_full": day.LunarFull,
		"year_gan_zhi": day.YearGanZhi, "month_gan_zhi": day.MonthGanZhi, "day_gan_zhi": day.DayGanZhi,
		"yi": day.CalendarYi, "ji": day.CalendarJi,
	}})
}

// GetCalendarRange returns a complete, bounded set of civil-day almanac data.
func GetCalendarRange(c *gin.Context) {
	zone, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "日期服务暂不可用"})
		return
	}
	parse := func(value string) (time.Time, bool) {
		day, err := time.ParseInLocation("2006-01-02", value, zone)
		return day, err == nil && day.Format("2006-01-02") == value && day.Year() >= 1900 && day.Year() <= 2100
	}
	start, startOK := parse(c.Query("start"))
	end, endOK := parse(c.Query("end"))
	if !startOK || !endOK {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择1900—2100年内的有效起止日期"})
		return
	}
	if start.After(end) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "起日不能晚于止日"})
		return
	}
	if end.After(start.AddDate(0, 0, 89)) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "一次最多查看90个日历日"})
		return
	}
	days := make([]gin.H, 0, 90)
	for date := start; !date.After(end); date = date.AddDate(0, 0, 1) {
		day, err := service.BuildPersonalDay(date.Format("2006-01-02"), nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "读取日期资料失败"})
			return
		}
		days = append(days, gin.H{
			"date": day.Date, "lunar_full": day.LunarFull, "day_gan_zhi": day.DayGanZhi,
			"yi": day.CalendarYi, "ji": day.CalendarJi,
		})
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"days": days}})
}
