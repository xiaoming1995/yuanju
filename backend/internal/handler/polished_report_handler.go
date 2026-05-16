package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"yuanju/internal/repository"
	"yuanju/internal/service"
)

// generatePolishedReportRequest POST body
type generatePolishedReportRequest struct {
	UserSituation string `json:"user_situation" binding:"required"`
}

// GenerateAndSavePolishedReport
// POST /api/bazi/polished-report/:chart_id
func GenerateAndSavePolishedReport(c *gin.Context) {
	chartID := c.Param("chart_id")
	if chartID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的排盘ID"})
		return
	}

	var body generatePolishedReportRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请填写当前情况描述"})
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

	result, err := service.LoadOrCalculateResult(chart)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[Polish] 请求 chart_id=%s user=%s", chart.ID, userIDStr)
	report, err := service.PolishReport(chart.ID, result, body.UserSituation)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"polished_report": report})
}

// GetPolishedReport
// GET /api/bazi/polished-report/:chart_id
func GetPolishedReport(c *gin.Context) {
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

	report, err := repository.GetPolishedByChartID(chart.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"polished_report": report})
}
