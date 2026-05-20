package handler

import (
	"log"
	"net/http"
	"yuanju/internal/repository"
	"yuanju/pkg/prompt"

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

	if !prompt.Has(module) {
		c.JSON(http.StatusNotFound, gin.H{"error": "未知模块: " + module})
		return
	}

	err := repository.UpdatePrompt(module, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新 Prompt 失败: " + err.Error()})
		return
	}

	if err := repository.SetCustomized(module, true); err != nil {
		log.Printf("[admin-prompt] module=%s set_customized failed: %v", module, err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

// ResetPromptToCanonical 强制把指定模块回到 canonical 注册表当前版本，
// 并清除 is_customized 标记。供 admin UI"重置为系统默认"按钮调用。
func ResetPromptToCanonical(c *gin.Context) {
	module := c.Param("module")

	def, ok := prompt.Lookup(module)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "unknown module: " + module})
		return
	}

	if err := repository.ResetToCanonical(module, def.Version, def.Content, def.Hash); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "重置失败: " + err.Error()})
		return
	}

	updated, err := repository.GetPromptByModule(module)
	if err != nil || updated == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "重置后查询失败"})
		return
	}
	c.JSON(http.StatusOK, updated)
}
