package handler

import (
	"log"
	"net/http"
	"strconv"
	"yuanju/internal/model"
	"yuanju/internal/repository"
	"yuanju/internal/service"
	"yuanju/pkg/bazi"

	"github.com/gin-gonic/gin"
)

type CalculateInput struct {
	Year         int     `json:"year" binding:"required,min=1900,max=2100"`
	Month        int     `json:"month" binding:"required,min=1,max=12"`
	Day          int     `json:"day" binding:"required,min=1,max=31"`
	Hour         int     `json:"hour" binding:"min=0,max=23"`
	Gender       string  `json:"gender" binding:"required,oneof=male female"`
	IsEarlyZishi bool    `json:"is_early_zishi"`
	Longitude    float64 `json:"longitude"`                                           // 出生地经度，用于真太阳时修正，0 表示不修正
	CalendarType string  `json:"calendar_type" binding:"omitempty,oneof=solar lunar"` // solar: 公历, lunar: 农历
	IsLeapMonth  bool    `json:"is_leap_month"`                                       // 是否为闰月
}

// Calculate 计算八字（无需登录，但若是已登录用户起盘，则自动落库保存历史）
func Calculate(c *gin.Context) {
	var input CalculateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "请检查生辰信息：" + err.Error()})
		return
	}

	if input.CalendarType == "" {
		input.CalendarType = "solar" // 默认兼容公历
	}

	result := bazi.Calculate(input.Year, input.Month, input.Day, input.Hour,
		input.Gender, input.IsEarlyZishi, input.Longitude, input.CalendarType, input.IsLeapMonth)

	var chartID string
	var ptrUserID *string

	// 如果上下文被 OptionalAuth 注入了 user_id，则绑定归属
	if userID, exists := c.Get("user_id"); exists {
		userIDStr, ok := userID.(string)
		if ok && userIDStr != "" {
			ptrUserID = &userIDStr
		}
	}

	// 无论是否登录（游客为 nil），均强制起盘落库，以便 Admin 观测真实流量
	chart := &model.BaziChart{
		UserID:     ptrUserID,
		BirthYear:  input.Year,
		BirthMonth: input.Month,
		BirthDay:   input.Day,
		BirthHour:  input.Hour,
		Gender:     input.Gender,
		YearGan:    result.YearGan, YearZhi: result.YearZhi,
		MonthGan: result.MonthGan, MonthZhi: result.MonthZhi,
		DayGan: result.DayGan, DayZhi: result.DayZhi,
		HourGan: result.HourGan, HourZhi: result.HourZhi,
		Wuxing:       result.Wuxing,
		Dayun:        result.Dayun,
		Yongshen:     result.Yongshen,
		Jishen:       result.Jishen,
		ChartHash:    result.ChartHash,
		CalendarType: input.CalendarType,
		IsLeapMonth:  input.IsLeapMonth,
	}

	// 静默落库，出错不影响排盘响应
	if savedChart, err := repository.CreateChart(chart); err == nil && savedChart != nil {
		chartID = savedChart.ID
	}

	c.JSON(http.StatusOK, gin.H{
		"result":   result,
		"chart_id": chartID,
	})
}

// GenerateReport 生成 AI 报告（需登录）
func GenerateReport(c *gin.Context) {
	chartID := c.Param("chart_id")
	if chartID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的排盘ID"})
		return
	}

	// 1. 获取命盘并验证归属
	chart, err := repository.GetChartByID(chartID)
	if err != nil || chart == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到指定命盘记录"})
		return
	}

	userID, _ := c.Get("user_id")
	userIDStr := userID.(string)

	if chart.UserID == nil || *chart.UserID != userIDStr {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权操作此命盘"})
		return
	}

	// 2. 将数据重构回溯至 BaziResult 提供给 AI Prompt （此时临时缺省早子与真太阳时间，但已够用）
	result := bazi.Calculate(
		chart.BirthYear, chart.BirthMonth, chart.BirthDay, chart.BirthHour,
		chart.Gender, false, 0, chart.CalendarType, chart.IsLeapMonth)

	// 3. 生成 AI 报告（若针对此 chartID 曾跑过则 0s 内自动命中缓存）
	log.Printf("[AI Report] 开始生成报告 chart_id=%s user=%s", chart.ID, userIDStr)
	report, err := service.GenerateAIReport(chart.ID, result)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 4. AI 分析后如果推断出了最新鲜的用神，则更新回溯
	if result.Yongshen != "" {
		chart.Yongshen = result.Yongshen
		chart.Jishen = result.Jishen
	}

	c.JSON(http.StatusOK, gin.H{
		"chart":  chart,
		"result": result,
		"report": report,
	})
}

// GenerateReportStream 流式生成 AI 报告（需登录）
func GenerateReportStream(c *gin.Context) {
	chartID := c.Param("chart_id")
	if chartID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的排盘ID"})
		return
	}

	chart, err := repository.GetChartByID(chartID)
	if err != nil || chart == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到指定命盘记录"})
		return
	}

	userID, _ := c.Get("user_id")
	userIDStr := userID.(string)

	if chart.UserID == nil || *chart.UserID != userIDStr {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权操作此命盘"})
		return
	}

	result := bazi.Calculate(
		chart.BirthYear, chart.BirthMonth, chart.BirthDay, chart.BirthHour,
		chart.Gender, false, 0, chart.CalendarType, chart.IsLeapMonth)

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	// 通知前端要开始推流了
	c.Writer.Flush()

	log.Printf("[AI Report Stream] 开始流式生成报告 chart_id=%s user=%s", chart.ID, userIDStr)
	err = service.GenerateAIReportStream(chart.ID, result, func(chunk string) error {
		c.SSEvent("message", chunk)
		c.Writer.Flush()
		return nil
	})

	if err != nil {
		c.SSEvent("error", err.Error())
		c.Writer.Flush()
	} else {
		c.SSEvent("done", "[DONE]")
		c.Writer.Flush()
	}
}

// GenerateLiunianReport 生成 AI 流年精批报告（需登录）
func GenerateLiunianReport(c *gin.Context) {
	chartID := c.Param("chart_id")
	if chartID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的排盘ID"})
		return
	}

	var req struct {
		TargetYear int `json:"target_year" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "需要 target_year 参数"})
		return
	}

	// 1. 获取命盘并验证归属
	chart, err := repository.GetChartByID(chartID)
	if err != nil || chart == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到指定命盘记录"})
		return
	}

	userID, _ := c.Get("user_id")
	userIDStr := userID.(string)

	if chart.UserID == nil || *chart.UserID != userIDStr {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权操作此命盘"})
		return
	}

	log.Printf("[AI Liunian Report] 开始生成流年报告 chart_id=%s year=%d", chart.ID, req.TargetYear)
	report, err := service.GenerateLiunianReport(chart.ID, req.TargetYear)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"report": report,
	})
}

// GetHistory 获取历史记录列表
func GetHistory(c *gin.Context) {
	userID, _ := c.Get("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit := 20
	offset := (page - 1) * limit

	charts, err := repository.GetChartsByUserID(userID.(string), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取历史记录失败"})
		return
	}

	if charts == nil {
		charts = []*model.BaziChart{}
	}

	c.JSON(http.StatusOK, gin.H{
		"charts": charts,
		"page":   page,
		"limit":  limit,
	})
}

// LiuYueInput 流月查询请求体
type LiuYueInput struct {
	LiuNianYear int    `json:"liu_nian_year" binding:"required,min=1900,max=2200"`
	DayGan      string `json:"day_gan" binding:"required"`
}

// HandleLiuYue 查询指定流年的 12 个流月数据（无需登录）
func HandleLiuYue(c *gin.Context) {
	var input LiuYueInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误：" + err.Error()})
		return
	}

	items, currentIndex, err := bazi.CalcLiuYue(input.LiuNianYear, input.DayGan)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"liu_yue":             items,
		"current_month_index": currentIndex,
	})
}

// GetHistoryDetail 获取历史记录详情
func GetHistoryDetail(c *gin.Context) {
	userID, _ := c.Get("user_id")
	chartID := c.Param("id")

	chart, err := repository.GetChartByID(chartID)
	if err != nil || chart == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "记录不存在"})
		return
	}

	// 越权访问检查
	if chart.UserID == nil || *chart.UserID != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权访问此记录"})
		return
	}

	report, _ := repository.GetReportByChartID(chartID)

	// 重新计算完整 BaziResult（DB 仅存精简字段，result 含十神/藏干/纳音/神煞等）
	// longitude 和 is_early_zishi 未持久化，使用默认值（0 和 false）
	result := bazi.Calculate(
		chart.BirthYear, chart.BirthMonth, chart.BirthDay, chart.BirthHour,
		chart.Gender, false, 0, chart.CalendarType, chart.IsLeapMonth)

	c.JSON(http.StatusOK, gin.H{
		"chart":  chart,
		"result": result,
		"report": report,
	})
}
