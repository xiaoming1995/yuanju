package handler

import (
	"database/sql"
	"net/http"
	"yuanju/internal/repository"
	"yuanju/pkg/database"

	"github.com/gin-gonic/gin"
)

// GetShenshaAnnotations 公开接口：返回全部神煞注解
// GET /api/shensha/annotations
func GetShenshaAnnotations(c *gin.Context) {
	annotations, err := repository.GetAllShenshaAnnotations(database.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取神煞注解失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": annotations})
}

// AdminUpdateShenshaAnnotation Admin 接口：更新指定神煞的描述文案
// PUT /api/admin/shensha-annotations/:name
func AdminUpdateShenshaAnnotation(c *gin.Context) {
	name := c.Param("name")

	var body struct {
		Description string `json:"description" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "description 字段不能为空"})
		return
	}

	err := repository.UpdateShenshaAnnotation(database.DB, name, body.Description)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "神煞不存在: " + name})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "神煞注解已更新", "name": name})
}
