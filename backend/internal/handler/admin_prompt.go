package handler

import (
	"net/http"
	"yuanju/internal/model"
	"yuanju/internal/repository"
	"yuanju/pkg/prompt"

	"github.com/gin-gonic/gin"
)

// promptWithDrift 在 model.AIPrompt 之上附加读取时计算的漂移信息。
type promptWithDrift struct {
	model.AIPrompt
	CanonicalVersion string `json:"canonical_version"`
	DriftStatus      string `json:"drift_status"`
}

// GetPrompts 获取所有配置的 Prompts（含相对出厂版的漂移状态）
func GetPrompts(c *gin.Context) {
	prompts, err := repository.GetAllPrompts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取 Prompt 失败"})
		return
	}
	out := make([]promptWithDrift, 0, len(prompts))
	for _, p := range prompts {
		item := promptWithDrift{AIPrompt: p}
		if def, ok := prompt.Lookup(p.Module); ok {
			item.CanonicalVersion = def.Version
		}
		item.DriftStatus = prompt.DriftStatus(p.Module, p.Content, p.CanonicalHash)
		out = append(out, item)
	}
	c.JSON(http.StatusOK, out)
}

// UpdatePrompt 更新特定的 Prompt 模板（保存维护版）。
// 分支点对齐当前出厂版；据内容是否偏离出厂版决定 is_customized。
func UpdatePrompt(c *gin.Context) {
	module := c.Param("module")
	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求格式"})
		return
	}

	def, ok := prompt.Lookup(module)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "未知模块: " + module})
		return
	}

	isCustomized := prompt.HashContent(req.Content) != def.Hash
	if err := repository.UpdateMaintained(module, req.Content, def.Version, def.Hash, isCustomized); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新 Prompt 失败: " + err.Error()})
		return
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

// GetPromptCanonical 返回指定模块的出厂版（代码 canonical）内容，供后台「采用出厂新版」载入编辑器。
func GetPromptCanonical(c *gin.Context) {
	module := c.Param("module")
	def, ok := prompt.Lookup(module)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "unknown module: " + module})
		return
	}
	c.JSON(http.StatusOK, gin.H{"version": def.Version, "content": def.Content})
}
