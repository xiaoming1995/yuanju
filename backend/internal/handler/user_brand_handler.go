package handler

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"yuanju/configs"
	"yuanju/internal/repository"
	"yuanju/pkg/ratelimit"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	maxLogoBytes       = 2 * 1024 * 1024 // 2MB
	logoSubDir         = "brand-logos"
	staticURLPrefix    = "/static/uploads"
	headerSampleLength = 12
)

var logoLimiter = ratelimit.New(10, time.Minute)

// BrandUpdateReq is the PUT /api/user/export-brand body.
type BrandUpdateReq struct {
	Title         string `json:"title"`
	FooterText    string `json:"footer_text"`
	WatermarkMode string `json:"watermark_mode"`
	WatermarkText string `json:"watermark_text"`
	LogoMode      string `json:"logo_mode"`
}

// BrandResponse is the GET / PUT response shape.
type BrandResponse struct {
	Title         string `json:"title"`
	FooterText    string `json:"footer_text"`
	LogoURL       string `json:"logo_url"`
	LogoMode      string `json:"logo_mode"`
	WatermarkMode string `json:"watermark_mode"`
	WatermarkText string `json:"watermark_text"`
}

// RequireUserID pulls user_id from gin context (set by middleware.Auth)
// and aborts 401 if missing. Splits Auth() (token verification) from user-id
// presence check so this handler suite can be unit-tested without a real JWT.
func RequireUserID(c *gin.Context) {
	v, ok := c.Get("user_id")
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未登录，请先登录"})
		return
	}
	// Note: this 401s gracefully if user_id isn't a string. The existing
	// auth pattern in other handlers (e.g. bazi_handler.go) does a panicking
	// cast — we chose graceful degradation because the brand UI flows through
	// this often. Keep both reads aligned if the JWT claim type ever changes.
	s, ok := v.(string)
	if !ok || s == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "无效的认证信息"})
		return
	}
	c.Set("user_id_str", s)
	c.Next()
}

func getUserIDStr(c *gin.Context) string {
	v, _ := c.Get("user_id_str")
	s, _ := v.(string)
	return s
}

func validateBrandUpdate(req BrandUpdateReq) error {
	if utf8.RuneCountInString(req.Title) > 20 {
		return errors.New("品牌标题不能超过 20 字符")
	}
	if utf8.RuneCountInString(req.FooterText) > 40 {
		return errors.New("底部文字不能超过 40 字符")
	}
	if utf8.RuneCountInString(req.WatermarkText) > 30 {
		return errors.New("水印文字不能超过 30 字符")
	}
	switch req.WatermarkMode {
	case "none", "bottom", "diagonal":
	default:
		return errors.New("水印模式不合法")
	}
	switch req.LogoMode {
	case "", "icon", "wordmark":
	default:
		return errors.New("logo 模式不合法")
	}
	for _, s := range []string{req.Title, req.FooterText, req.WatermarkText} {
		if strings.ContainsAny(s, `<>"'&`) {
			return errors.New("文本包含不允许的字符")
		}
	}
	return nil
}

// detectImageType returns "png" / "jpg" / "webp" / "" based on magic bytes.
// header must be at least 12 bytes; returns "" if shorter or no match.
func detectImageType(header []byte) string {
	if len(header) < headerSampleLength {
		return ""
	}
	// PNG: 89 50 4E 47 0D 0A 1A 0A
	if header[0] == 0x89 && header[1] == 0x50 && header[2] == 0x4E && header[3] == 0x47 &&
		header[4] == 0x0D && header[5] == 0x0A && header[6] == 0x1A && header[7] == 0x0A {
		return "png"
	}
	// JPEG: FF D8 FF
	if header[0] == 0xFF && header[1] == 0xD8 && header[2] == 0xFF {
		return "jpg"
	}
	// WebP: RIFF????WEBP
	if header[0] == 'R' && header[1] == 'I' && header[2] == 'F' && header[3] == 'F' &&
		header[8] == 'W' && header[9] == 'E' && header[10] == 'B' && header[11] == 'P' {
		return "webp"
	}
	return ""
}

func logoURLFromPath(p string) string {
	if p == "" {
		return ""
	}
	return staticURLPrefix + "/" + p
}

// buildBrandResponse loads the user's brand row and shapes it for the API response.
// Used by both GetExportBrand and UpdateExportBrand to avoid re-entering a handler.
func buildBrandResponse(userID string) (BrandResponse, error) {
	b, err := repository.GetExportBrand(userID)
	if err != nil {
		return BrandResponse{}, err
	}
	return BrandResponse{
		Title:         b.Title,
		FooterText:    b.FooterText,
		LogoURL:       logoURLFromPath(b.LogoPath),
		LogoMode:      b.LogoMode,
		WatermarkMode: b.WatermarkMode,
		WatermarkText: b.WatermarkText,
	}, nil
}

// GetExportBrand handles GET /api/user/export-brand.
func GetExportBrand(c *gin.Context) {
	resp, err := buildBrandResponse(getUserIDStr(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取品牌设置失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// UpdateExportBrand handles PUT /api/user/export-brand.
func UpdateExportBrand(c *gin.Context) {
	userID := getUserIDStr(c)
	var req BrandUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求体格式错误"})
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	req.FooterText = strings.TrimSpace(req.FooterText)
	req.WatermarkText = strings.TrimSpace(req.WatermarkText)
	if req.WatermarkMode == "" {
		req.WatermarkMode = "none"
	}
	if req.LogoMode == "" {
		req.LogoMode = "icon"
	}
	if err := validateBrandUpdate(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := repository.UpsertExportBrandText(userID, req.Title, req.FooterText, req.WatermarkMode, req.WatermarkText, req.LogoMode); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存品牌设置失败"})
		return
	}
	resp, err := buildBrandResponse(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取品牌设置失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// UploadExportBrandLogo handles POST /api/user/export-brand/logo.
func UploadExportBrandLogo(c *gin.Context) {
	userID := getUserIDStr(c)
	if !logoLimiter.Allow(userID) {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "上传过于频繁，请稍后重试"})
		return
	}
	if c.Request.ContentLength > maxLogoBytes {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "logo 文件不能超过 2MB"})
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 file 字段"})
		return
	}
	if file.Size > maxLogoBytes {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "logo 文件不能超过 2MB"})
		return
	}
	declaredCT := file.Header.Get("Content-Type")
	switch declaredCT {
	case "image/png", "image/jpeg", "image/webp":
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "仅支持 PNG / JPG / WebP 格式"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法读取上传文件"})
		return
	}
	defer src.Close()
	header := make([]byte, headerSampleLength)
	if _, err := src.Read(header); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无法读取文件头"})
		return
	}
	ext := detectImageType(header)
	if ext == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不是有效的图片文件"})
		return
	}
	expectedCT := map[string]string{"png": "image/png", "jpg": "image/jpeg", "webp": "image/webp"}[ext]
	if declaredCT != expectedCT {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件内容与声明的类型不一致"})
		return
	}

	relPath := filepath.Join(logoSubDir, uuid.NewString()+"."+ext)
	absPath := filepath.Join(configs.AppConfig.UploadDir, relPath)
	if err := c.SaveUploadedFile(file, absPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败"})
		return
	}

	oldPath, dbErr := repository.UpdateExportBrandLogo(userID, relPath)
	if dbErr != nil {
		_ = os.Remove(absPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新品牌设置失败"})
		return
	}
	if oldPath != "" && oldPath != relPath {
		_ = os.Remove(filepath.Join(configs.AppConfig.UploadDir, oldPath))
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"logo_url": logoURLFromPath(relPath)}})
}

// DeleteExportBrandLogo handles DELETE /api/user/export-brand/logo.
func DeleteExportBrandLogo(c *gin.Context) {
	userID := getUserIDStr(c)
	oldPath, err := repository.ClearExportBrandLogo(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "清除 logo 失败"})
		return
	}
	if oldPath != "" {
		_ = os.Remove(filepath.Join(configs.AppConfig.UploadDir, oldPath))
	}
	c.Status(http.StatusNoContent)
}
