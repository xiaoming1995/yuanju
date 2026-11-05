package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
	"yuanju/internal/repository"
	"yuanju/internal/service"

	"github.com/gin-gonic/gin"
)

var todayQuestionNow = time.Now

func PostTodayQuestion(c *gin.Context) {
	var input struct {
		Question string `json:"question"`
	}
	if err := c.ShouldBindJSON(&input); err != nil || !service.ValidTodayQuestion(input.Question) {
		if service.UnsupportedTodayQuestion(input.Question) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "当前不支持投资买卖或医疗诊断类问题"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": service.ErrTodayQuestionInvalid.Error()})
		return
	}
	userID := personalUserID(c)
	chart, err := repository.GetPersonalChart(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取本人命盘失败"})
		return
	}
	if chart == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "请先设置本人命盘", "code": "personal_chart_unbound"})
		return
	}
	result, err := service.LoadOrCalculateResult(chart)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取本人命盘失败"})
		return
	}
	zone, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "日期服务暂不可用"})
		return
	}
	date := todayQuestionNow().In(zone).Format("2006-01-02")
	context, err := service.BuildTodayQuestionContext(date, chart.ID, input.Question, result)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "今日问事依据暂不可用"})
		return
	}
	contextJSON, err := json.Marshal(context)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "今日问事依据暂不可用"})
		return
	}
	if todayQuestionNow().In(zone).Format("2006-01-02") != date {
		c.JSON(http.StatusConflict, gin.H{"error": "北京时间日期已切换，请重新提问"})
		return
	}
	requestID, dailyLimit, err := repository.ReserveTodayQuestionRequest(userID, chart.ID, date, context.Category, context.Question, contextJSON)
	if err != nil {
		if errors.Is(err, repository.ErrTodayQuestionDailyLimit) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": fmt.Sprintf("今天免费问事次数与额外次数均已用完（每日免费 %d 次），北京时间明天 0 点重置免费次数", dailyLimit), "code": "today_question_daily_limit"})
			return
		}
		log.Printf("[TodayQuestion] reserve request failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "今日问事记录暂不可用"})
		return
	}
	answer, trace, generateErr := service.GenerateTodayQuestionWithTrace(context)
	completion := repository.TodayQuestionCompletion{
		AISystemPrompt: trace.SystemPrompt, AIUserPrompt: trace.UserPrompt, AIOutput: trace.RawOutput,
		ProviderID: trace.ProviderID, Model: trace.Model, DurationMs: trace.DurationMs,
		PromptTokens: trace.Usage.PromptTokens, CompletionTokens: trace.Usage.CompletionTokens,
		TotalTokens: trace.Usage.TotalTokens, ReasoningTokens: trace.Usage.ReasoningTokens,
		CacheHitTokens: trace.Usage.CacheHitTokens, CacheMissTokens: trace.Usage.CacheMissTokens,
		ErrorMsg: trace.ErrorMessage,
	}
	if trace.Usage.TotalTokens > 0 {
		completion.EstimatedCostCny = service.CalcCost(trace.Model, trace.Usage.PromptTokens, trace.Usage.CompletionTokens)
	}
	if generateErr != nil {
		completion.Status = "error"
	} else {
		completion.Status = "success"
		completion.AnswerData, _ = json.Marshal(answer)
	}
	if err := repository.CompleteTodayQuestionRequest(requestID, completion); err != nil {
		log.Printf("[TodayQuestion] complete request failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "今日问事记录暂不可用"})
		return
	}
	if generateErr != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": service.ErrTodayQuestionAI.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": answer})
}
