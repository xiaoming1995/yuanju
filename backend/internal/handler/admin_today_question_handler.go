package handler

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
	"yuanju/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func AdminGetTodayQuestionSettings(c *gin.Context) {
	limit, err := repository.GetTodayQuestionDailyLimit()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取每日问事次数失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"daily_limit": limit}})
}

func AdminUpdateTodayQuestionSettings(c *gin.Context) {
	var input struct {
		DailyLimit *int `json:"daily_limit"`
	}
	if err := c.ShouldBindJSON(&input); err != nil || input.DailyLimit == nil || *input.DailyLimit < 1 || *input.DailyLimit > repository.MaxTodayQuestionDailyLimit {
		c.JSON(http.StatusBadRequest, gin.H{"error": "每日问事次数须为 1—100 的整数"})
		return
	}
	if err := repository.SetTodayQuestionDailyLimit(*input.DailyLimit); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存每日问事次数失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"daily_limit": *input.DailyLimit}})
}

func AdminListTodayQuestions(c *gin.Context) {
	page, pageErr := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, limitErr := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if pageErr != nil || limitErr != nil || page < 1 || page > 100000 || limit < 1 || limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "分页参数无效"})
		return
	}
	from, to := c.Query("from"), c.Query("to")
	for _, value := range []string{from, to} {
		if value != "" {
			parsed, err := time.Parse("2006-01-02", value)
			if err != nil || parsed.Format("2006-01-02") != value {
				c.JSON(http.StatusBadRequest, gin.H{"error": "日期格式应为 YYYY-MM-DD"})
				return
			}
		}
	}
	if from != "" && to != "" && from > to {
		c.JSON(http.StatusBadRequest, gin.H{"error": "起始日期不能晚于结束日期"})
		return
	}
	status := c.Query("status")
	if status != "" && status != "processing" && status != "success" && status != "insufficient" && status != "error" && status != "legacy" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "状态参数无效"})
		return
	}
	user := strings.TrimSpace(c.Query("user"))
	if utf8.RuneCountInString(user) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户搜索内容过长"})
		return
	}
	total, items, err := repository.ListTodayQuestionRequests(from, to, status, user, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取今日问事记录失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"total": total, "items": items, "page": page, "limit": limit}})
}

func AdminGetTodayQuestion(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "记录 ID 无效"})
		return
	}
	item, err := repository.GetTodayQuestionRequest(id)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "今日问事记录不存在"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取今日问事记录失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": item})
}
