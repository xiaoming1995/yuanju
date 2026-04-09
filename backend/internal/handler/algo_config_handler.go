package handler

import (
	"net/http"
	"strconv"
	"yuanju/internal/repository"
	"yuanju/internal/service"
	"yuanju/pkg/bazi"

	"github.com/gin-gonic/gin"
)

// AdminGetAlgoConfig GET /api/admin/algo-config
// 返回全部算法参数，标注来源（db/default）
func AdminGetAlgoConfig(c *gin.Context) {
	rows, err := repository.GetAllAlgoConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 构造返回结构（默认值作为 fallback 展示）
	defaults := map[string]string{
		"jixiong_jiHan_min":   strconv.Itoa(bazi.DefaultJiHanMin),
		"jixiong_jiRe_min":    strconv.Itoa(bazi.DefaultJiReMin),
		"jixiong_shenQiang_pct": strconv.FormatFloat(bazi.DefaultShenQiangPct, 'f', 1, 64),
	}

	dbMap := make(map[string]repository.AlgoConfigRow, len(rows))
	for _, r := range rows {
		dbMap[r.Key] = r
	}

	type item struct {
		Key         string `json:"key"`
		Value       string `json:"value"`
		Description string `json:"description"`
		Source      string `json:"source"` // "db" or "default"
	}

	var result []item
	for key, defVal := range defaults {
		if row, ok := dbMap[key]; ok {
			result = append(result, item{Key: key, Value: row.Value, Description: row.Description, Source: "db"})
		} else {
			result = append(result, item{Key: key, Value: defVal, Description: "", Source: "default"})
		}
	}

	c.JSON(http.StatusOK, result)
}

// AdminUpdateAlgoConfig PUT /api/admin/algo-config/:key
func AdminUpdateAlgoConfig(c *gin.Context) {
	key := c.Param("key")

	validKeys := map[string]bool{
		"jixiong_jiHan_min":     true,
		"jixiong_jiRe_min":      true,
		"jixiong_shenQiang_pct": true,
	}
	if !validKeys[key] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未知参数键: " + key})
		return
	}

	var body struct {
		Value       string `json:"value" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 数值格式校验
	switch key {
	case "jixiong_jiHan_min", "jixiong_jiRe_min":
		if _, err := strconv.Atoi(body.Value); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数值必须为整数"})
			return
		}
	case "jixiong_shenQiang_pct":
		if _, err := strconv.ParseFloat(body.Value, 64); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数值必须为数字"})
			return
		}
	}

	if err := repository.UpsertAlgoConfig(key, body.Value, body.Description); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

// AdminReloadAlgoConfig POST /api/admin/algo-config/reload
func AdminReloadAlgoConfig(c *gin.Context) {
	if err := service.ReloadAlgoConfig(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "重载失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "算法配置已重载"})
}
