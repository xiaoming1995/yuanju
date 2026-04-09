package handler

import (
	"net/http"
	"strings"
	"unicode/utf8"
	"yuanju/internal/repository"

	"github.com/gin-gonic/gin"
)

var validTianGan = map[string]bool{
	"甲": true, "乙": true, "丙": true, "丁": true, "戊": true,
	"己": true, "庚": true, "辛": true, "壬": true, "癸": true,
}

var validDiZhi = map[string]bool{
	"子": true, "丑": true, "寅": true, "卯": true, "辰": true, "巳": true,
	"午": true, "未": true, "申": true, "酉": true, "戌": true, "亥": true,
}

// AdminGetAlgoTiaohou GET /api/admin/algo-tiaohou?day_gan=
func AdminGetAlgoTiaohou(c *gin.Context) {
	dayGan := c.Query("day_gan")
	if dayGan != "" && !validTianGan[dayGan] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的日干字符"})
		return
	}

	rows, err := repository.GetAllAlgoTiaohou(dayGan)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rows)
}

// AdminUpdateAlgoTiaohou PUT /api/admin/algo-tiaohou/:day_gan/:month_zhi
func AdminUpdateAlgoTiaohou(c *gin.Context) {
	dayGan := c.Param("day_gan")
	monthZhi := c.Param("month_zhi")

	if !validTianGan[dayGan] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的日干字符: " + dayGan})
		return
	}
	if !validDiZhi[monthZhi] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的月支字符: " + monthZhi})
		return
	}

	var body struct {
		XiElements string `json:"xi_elements" binding:"required"`
		Text       string `json:"text"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 校验 xi_elements：每个逗号分隔的部分必须是单个天干字符
	for _, elem := range strings.Split(body.XiElements, ",") {
		elem = strings.TrimSpace(elem)
		if utf8.RuneCountInString(elem) != 1 || !validTianGan[elem] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "xi_elements 每项须为单个天干字符，非法值: " + elem})
			return
		}
	}

	if err := repository.UpsertAlgoTiaohou(dayGan, monthZhi, body.XiElements, body.Text); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

// AdminDeleteAlgoTiaohou DELETE /api/admin/algo-tiaohou/:day_gan/:month_zhi
// 删除自定义规则后，该条目将回归硬编码默认值（需 reload）
func AdminDeleteAlgoTiaohou(c *gin.Context) {
	dayGan := c.Param("day_gan")
	monthZhi := c.Param("month_zhi")

	if !validTianGan[dayGan] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的日干字符: " + dayGan})
		return
	}
	if !validDiZhi[monthZhi] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的月支字符: " + monthZhi})
		return
	}

	if err := repository.DeleteAlgoTiaohou(dayGan, monthZhi); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "已删除，请调用 reload 使更改生效"})
}
