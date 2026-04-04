package handler

import (
	"net/http"
	"yuanju/internal/model"
	"yuanju/internal/repository"
	"yuanju/internal/service"

	"github.com/gin-gonic/gin"
)

// AdminListCelebrities获取全量名人录
func AdminListCelebrities(c *gin.Context) {
	celebs, err := repository.ListCelebrities(false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取名人录失败"})
		return
	}
	if celebs == nil {
		celebs = make([]model.CelebrityRecord, 0) // 防止返回 null
	}
	c.JSON(http.StatusOK, gin.H{"data": celebs})
}

// AdminCreateCelebrity 创建一条名人数据
func AdminCreateCelebrity(c *gin.Context) {
	var body struct {
		Name   string `json:"name" binding:"required"`
		Gender string `json:"gender"`
		Traits string `json:"traits"`
		Career string `json:"career"`
		Active bool   `json:"active"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数格式错误"})
		return
	}

	celeb, err := repository.CreateCelebrity(body.Name, body.Gender, body.Traits, body.Career, body.Active)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建名人记录失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": celeb})
}

// AdminUpdateCelebrity 编辑/更新名人记录
func AdminUpdateCelebrity(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 ID"})
		return
	}

	var body struct {
		Name   string `json:"name" binding:"required"`
		Gender string `json:"gender"`
		Traits string `json:"traits"`
		Career string `json:"career"`
		Active bool   `json:"active"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数格式错误"})
		return
	}

	err := repository.UpdateCelebrity(id, body.Name, body.Gender, body.Traits, body.Career, body.Active)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": "更新成功"})
}

// AdminDeleteCelebrity 删除名人记录
func AdminDeleteCelebrity(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 ID"})
		return
	}

	err := repository.DeleteCelebrity(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": "删除成功"})
}

// AdminAIGenerateCelebrities 批量 AI 自动收集名人数据
func AdminAIGenerateCelebrities(c *gin.Context) {
	var req struct {
		Topic string `json:"topic" binding:"required"`
		Count int    `json:"count" binding:"required,min=1,max=20"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误，需要提供 topic 和 count(1-20)"})
		return
	}

	records, err := service.GenerateCelebrities(req.Topic, req.Count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var inserted []model.CelebrityRecord
	for _, rec := range records {
		created, err := repository.CreateCelebrity(rec.Name, rec.Gender, rec.Traits, rec.Career, true)
		if err == nil && created != nil {
			inserted = append(inserted, *created)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"inserted_count": len(inserted),
			"records":        inserted,
		},
	})
}
