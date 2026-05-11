package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
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
		// 同步写入完整 BaziResult 快照，让下游全部跳过 lunar-go 重算
		if resultJSON, mErr := json.Marshal(result); mErr == nil {
			_ = repository.SaveChartResultJSON(chartID, resultJSON)
		}
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

	// 2. 加载命盘 BaziResult 快照（result_json 优先；老命盘懒加载并写回 DB）
	result, err := service.LoadOrCalculateResult(chart)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. 生成 AI 报告（若针对此 chartID 曾跑过则 0s 内自动命中缓存）
	log.Printf("[AI Report] 开始生成报告 chart_id=%s user=%s", chart.ID, userIDStr)
	report, err := service.GenerateAIReport(chart.ID, result, &userIDStr)
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
	t0 := time.Now()
	chartID := c.Param("chart_id")
	if chartID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的排盘ID"})
		return
	}

	log.Printf("[Stream T+%dms] 开始处理 chart_id=%s", time.Since(t0).Milliseconds(), chartID)

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

	log.Printf("[Stream T+%dms] 加载命盘快照（result_json 优先）", time.Since(t0).Milliseconds())
	result, err := service.LoadOrCalculateResult(chart)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Printf("[Stream T+%dms] 命盘加载完成", time.Since(t0).Milliseconds())

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")
	c.Header("X-Accel-Buffering", "no")

	c.Writer.Flush()
	log.Printf("[Stream T+%dms] SSE headers 已发送，开始调用 AI", time.Since(t0).Milliseconds())

	chunkCount := 0
	err = service.GenerateAIReportStream(chart.ID, result, &userIDStr, func(chunk string) error {
		chunkCount++
		if chunkCount == 1 {
			log.Printf("[Stream T+%dms] ✅ 收到第 1 个 chunk (长度=%d)", time.Since(t0).Milliseconds(), len(chunk))
		} else if chunkCount%50 == 0 {
			log.Printf("[Stream T+%dms] 已接收 %d 个 chunks", time.Since(t0).Milliseconds(), chunkCount)
		}
		jsonBytes, _ := json.Marshal(map[string]string{"chunk": chunk})
		fmt.Fprintf(c.Writer, "event: message\ndata: %s\n\n", string(jsonBytes))
		c.Writer.Flush()
		return nil
	}, func() error {
		// 推理模型进入思考阶段时通知前端
		log.Printf("[Stream T+%dms] 🧠 发送 thinking 事件到前端", time.Since(t0).Milliseconds())
		fmt.Fprintf(c.Writer, "event: thinking\ndata: {}\n\n")
		c.Writer.Flush()
		return nil
	})

	log.Printf("[Stream T+%dms] 流式生成结束，共 %d 个 chunks, err=%v", time.Since(t0).Milliseconds(), chunkCount, err)

	if err != nil {
		fmt.Fprintf(c.Writer, "event: error\ndata: %s\n\n", err.Error())
		c.Writer.Flush()
	} else {
		fmt.Fprintf(c.Writer, "event: done\ndata: [DONE]\n\n")
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
	report, err := service.GenerateLiunianReport(chart.ID, req.TargetYear, &userIDStr)
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

	// 加载完整 BaziResult 快照（result_json 优先，老命盘懒加载）
	result, err := service.LoadOrCalculateResult(chart)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"chart":  chart,
		"result": result,
		"report": report,
	})
}

// HandlePastEventsYears 即时返回算法+模板生成的所有年份叙述（无 AI，毫秒级）
func HandlePastEventsYears(c *gin.Context) {
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
	if chart.UserID == nil || *chart.UserID != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权操作此命盘"})
		return
	}
	resp, err := service.GeneratePastEventsYears(chartID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// HandleDayunSummariesStream 按大运分段流式生成 AI 总结（SSE，每段独立缓存）
func HandleDayunSummariesStream(c *gin.Context) {
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
	if chart.UserID == nil || *chart.UserID != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权操作此命盘"})
		return
	}

	userIDStr := userID.(string)
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")
	c.Header("X-Accel-Buffering", "no")
	c.Writer.Flush()

	err = service.GenerateDayunSummariesStream(chartID, &userIDStr, func(item service.DayunSummaryStreamItem) error {
		bytes, _ := json.Marshal(item)
		fmt.Fprintf(c.Writer, "event: dayun\ndata: %s\n\n", string(bytes))
		c.Writer.Flush()
		return nil
	})
	if err != nil {
		fmt.Fprintf(c.Writer, "event: error\ndata: %s\n\n", err.Error())
	} else {
		fmt.Fprintf(c.Writer, "event: done\ndata: [DONE]\n\n")
	}
	c.Writer.Flush()
}

// HandlePastEventsStream 流式生成过往年份事件推算（需登录）
func HandlePastEventsStream(c *gin.Context) {
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
	if chart.UserID == nil || *chart.UserID != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权操作此命盘"})
		return
	}
	userIDStr := userID.(string)

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")
	c.Header("X-Accel-Buffering", "no")
	c.Writer.Flush()

	err = service.GeneratePastEventsStream(chartID, &userIDStr, func(chunk string) error {
		jsonBytes, _ := json.Marshal(map[string]string{"chunk": chunk})
		fmt.Fprintf(c.Writer, "event: message\ndata: %s\n\n", string(jsonBytes))
		c.Writer.Flush()
		return nil
	}, func() error {
		fmt.Fprintf(c.Writer, "event: thinking\ndata: {}\n\n")
		c.Writer.Flush()
		return nil
	})

	if err != nil {
		fmt.Fprintf(c.Writer, "event: error\ndata: %s\n\n", err.Error())
	} else {
		fmt.Fprintf(c.Writer, "event: done\ndata: [DONE]\n\n")
	}
	c.Writer.Flush()
}
