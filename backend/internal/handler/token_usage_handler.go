package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"yuanju/internal/repository"
	"yuanju/internal/service"
)

// AdminGetTokenUsageSummary GET /api/admin/token-usage/summary?from=YYYY-MM-DD&to=YYYY-MM-DD
func AdminGetTokenUsageSummary(c *gin.Context) {
	from, to := parseDateRange(c)
	rows, err := repository.GetTokenUsageSummary(from, to, service.CalcCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if rows == nil {
		rows = []repository.TokenUsageSummaryRow{}
	}
	c.JSON(http.StatusOK, rows)
}

// AdminGetTokenUsageDetail GET /api/admin/token-usage/detail?user_id=xxx&from=&to=&page=1&limit=20
func AdminGetTokenUsageDetail(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id 必填"})
		return
	}
	from, to := parseDateRange(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	model := c.DefaultQuery("model", "")
	total, items, err := repository.GetTokenUsageDetail(userID, from, to, page, limit, model, service.CalcCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if items == nil {
		items = []repository.TokenUsageDetailRow{}
	}
	c.JSON(http.StatusOK, gin.H{
		"total": total,
		"items": items,
	})
}

// parseDateRange 解析 from/to 查询参数，默认当月第一天至今日
func parseDateRange(c *gin.Context) (from, to time.Time) {
	now := time.Now()
	defaultFrom := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	defaultTo := now

	fromStr := c.DefaultQuery("from", defaultFrom.Format("2006-01-02"))
	toStr := c.DefaultQuery("to", defaultTo.Format("2006-01-02"))

	from, _ = time.Parse("2006-01-02", fromStr)
	to, _ = time.Parse("2006-01-02", toStr)

	if from.IsZero() {
		from = defaultFrom
	}
	if to.IsZero() {
		to = defaultTo
	}
	return from, to
}
