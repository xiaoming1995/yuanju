package handler

import (
	"net/http"
	"yuanju/internal/service"

	"github.com/gin-gonic/gin"
)

func GetUserProfile(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录，请先登录"})
		return
	}
	userIDStr, ok := userID.(string)
	if !ok || userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的认证信息"})
		return
	}

	profile, err := service.GetUserProfileOverview(userIDStr)
	if err != nil {
		if err.Error() == "用户不存在" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取个人中心失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": profile})
}
