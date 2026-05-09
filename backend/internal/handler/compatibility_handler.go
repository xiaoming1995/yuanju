package handler

import (
	"net/http"
	"strconv"
	"yuanju/internal/model"
	"yuanju/internal/service"

	"github.com/gin-gonic/gin"
)

type CompatibilityProfileInput struct {
	Year         int    `json:"year" binding:"required,min=1900,max=2100"`
	Month        int    `json:"month" binding:"required,min=1,max=12"`
	Day          int    `json:"day" binding:"required,min=1,max=31"`
	Hour         int    `json:"hour" binding:"required,min=0,max=23"`
	Gender       string `json:"gender" binding:"required,oneof=male female"`
	CalendarType string `json:"calendar_type" binding:"omitempty,oneof=solar lunar"`
	IsLeapMonth  bool   `json:"is_leap_month"`
}

type CreateCompatibilityReadingRequest struct {
	Self    CompatibilityProfileInput `json:"self" binding:"required"`
	Partner CompatibilityProfileInput `json:"partner" binding:"required"`
}

func CreateCompatibilityReading(c *gin.Context) {
	var req CreateCompatibilityReadingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "请检查双方生辰信息：" + err.Error()})
		return
	}
	userID, _ := c.Get("user_id")
	detail, err := service.CreateCompatibilityReading(
		userID.(string),
		model.CompatibilityBirthProfile(req.Self),
		model.CompatibilityBirthProfile(req.Partner),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": detail})
}

func GetCompatibilityHistory(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	limit := 20
	offset := (page - 1) * limit
	userID, _ := c.Get("user_id")
	items, err := service.GetCompatibilityHistoryForUser(userID.(string), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取合盘历史失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

func GetCompatibilityDetail(c *gin.Context) {
	readingID := c.Param("id")
	userID, _ := c.Get("user_id")
	detail, err := service.GetCompatibilityDetailForUser(readingID, userID.(string))
	if err != nil {
		if err.Error() == "forbidden" {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权查看此合盘记录"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取合盘详情失败"})
		return
	}
	if detail == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到合盘记录"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": detail})
}

func GenerateCompatibilityReport(c *gin.Context) {
	readingID := c.Param("id")
	userID, _ := c.Get("user_id")
	report, err := service.GenerateCompatibilityReport(readingID, userID.(string))
	if err != nil {
		if err.Error() == "forbidden" {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权操作此合盘记录"})
			return
		}
		if err.Error() == "未找到合盘记录" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": report})
}
