package handler

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"yuanju/configs"
	"yuanju/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const officialAccountQRDir = "official-account-qr"
const maxOfficialAccountQRBytes = 2 << 20

func GetPublicOfficialAccountGuide(c *gin.Context) {
	guide, err := repository.GetOfficialAccountGuide()
	if err != nil {
		log.Printf("read official account guide: %v", err)
		c.JSON(http.StatusOK, gin.H{"data": gin.H{"enabled": false}})
		return
	}
	if !guide.Ready() {
		c.JSON(http.StatusOK, gin.H{"data": gin.H{"enabled": false}})
		return
	}
	info, err := os.Stat(officialAccountQRFile(guide.QRPath))
	if err != nil || !info.Mode().IsRegular() {
		c.JSON(http.StatusOK, gin.H{"data": gin.H{"enabled": false}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": guide})
}

func AdminGetOfficialAccountGuide(c *gin.Context) {
	guide, err := repository.GetOfficialAccountGuide()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取公众号引导配置失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": guide})
}

func officialAccountQRFile(path string) string {
	return filepath.Join(configs.AppConfig.UploadDir, officialAccountQRDir, filepath.Base(path))
}

func AdminSaveOfficialAccountGuide(c *gin.Context) {
	var proposed repository.OfficialAccountGuide
	if err := c.ShouldBindJSON(&proposed); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "配置格式无效"})
		return
	}
	guide, err := repository.ValidateOfficialAccountGuide(proposed)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if guide.QRPath != "" {
		info, err := os.Stat(officialAccountQRFile(guide.QRPath))
		if err != nil || !info.Mode().IsRegular() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请上传有效的公众号二维码"})
			return
		}
	}
	previous, err := repository.GetOfficialAccountGuide()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取现有配置失败"})
		return
	}
	guide, err = repository.SaveOfficialAccountGuide(guide)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存公众号引导配置失败"})
		return
	}
	if previous.QRPath != "" && previous.QRPath != guide.QRPath {
		// A replaced image is no longer referenced by the persisted configuration.
		_ = os.Remove(officialAccountQRFile(previous.QRPath))
	}
	pruneUnusedOfficialAccountQR(guide.QRPath)
	c.JSON(http.StatusOK, gin.H{"data": guide})
}

func AdminUploadOfficialAccountQR(c *gin.Context) {
	if c.Request.ContentLength > maxOfficialAccountQRBytes+(1<<20) {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "二维码图片不能超过 2MB"})
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 file 字段"})
		return
	}
	if file.Size > maxOfficialAccountQRBytes {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "二维码图片不能超过 2MB"})
		return
	}
	declaredType := file.Header.Get("Content-Type")
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无法读取图片"})
		return
	}
	defer src.Close()
	header := make([]byte, headerSampleLength)
	if _, err := io.ReadFull(src, header); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "图片内容无效"})
		return
	}
	ext := detectImageType(header)
	validType := map[string]string{"png": "image/png", "jpg": "image/jpeg", "webp": "image/webp"}[ext]
	if validType == "" || declaredType != validType {
		c.JSON(http.StatusBadRequest, gin.H{"error": "仅支持内容与格式一致的 PNG、JPG 或 WebP 图片"})
		return
	}
	dir := filepath.Join(configs.AppConfig.UploadDir, officialAccountQRDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建二维码目录失败"})
		return
	}
	path := repository.OfficialAccountQRPrefix + uuid.NewString() + "." + ext
	if err := c.SaveUploadedFile(file, officialAccountQRFile(path)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存二维码失败"})
		return
	}
	guide, err := repository.GetOfficialAccountGuide()
	if err == nil {
		pruneUnusedOfficialAccountQR(guide.QRPath)
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"qr_path": path}})
}

func pruneUnusedOfficialAccountQR(activePath string) {
	dir := filepath.Join(configs.AppConfig.UploadDir, officialAccountQRDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	cutoff := time.Now().Add(-24 * time.Hour)
	for _, entry := range entries {
		if entry.IsDir() || strings.ContainsAny(entry.Name(), `/\\`) || repository.OfficialAccountQRPrefix+entry.Name() == activePath {
			continue
		}
		info, err := entry.Info()
		if err == nil && info.ModTime().Before(cutoff) {
			_ = os.Remove(filepath.Join(dir, entry.Name()))
		}
	}
}
