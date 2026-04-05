package handler

import (
	"net/http"
	"yuanju/internal/repository"

	"github.com/gin-gonic/gin"
)

// GetPrompts 获取所有配置的 Prompts
func GetPrompts(c *gin.Context) {
	prompts, err := repository.GetAllPrompts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取 Prompt 失败"})
		return
	}
	c.JSON(http.StatusOK, prompts)
}

// UpdatePrompt 更新特定的 Prompt 模板
func UpdatePrompt(c *gin.Context) {
	module := c.Param("module")
	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求格式"})
		return
	}

	err := repository.UpdatePrompt(module, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新 Prompt 失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}
