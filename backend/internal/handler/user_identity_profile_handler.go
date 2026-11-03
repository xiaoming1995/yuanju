package handler

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
	"yuanju/configs"
	"yuanju/internal/repository"
	"yuanju/pkg/ratelimit"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	maxAvatarBytes = 2 * 1024 * 1024
	avatarSubDir   = "user-avatars"
)

var avatarLimiter = ratelimit.New(10, time.Minute)

func UpdateProfileNickname(c *gin.Context) {
	var req struct {
		Nickname string `json:"nickname"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求体格式错误"})
		return
	}
	nickname := strings.TrimSpace(req.Nickname)
	if nickname == "" || utf8.RuneCountInString(nickname) > 32 || strings.ContainsFunc(nickname, unicode.IsControl) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "昵称应为 1 至 32 个有效字符"})
		return
	}
	userID := getUserIDStr(c)
	if err := repository.UpdateUserNickname(c.Request.Context(), userID, nickname, userActivityChannel(c)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存昵称失败"})
		return
	}
	user, err := repository.GetUserByID(userID)
	if err != nil || user == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取个人资料失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": user})
}

func UploadProfileAvatar(c *gin.Context) {
	userID := getUserIDStr(c)
	if !avatarLimiter.Allow(userID) {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "上传过于频繁，请稍后重试"})
		return
	}
	if c.Request.ContentLength > maxAvatarBytes+64*1024 {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "头像不能超过 2MB"})
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxAvatarBytes+64*1024)
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择头像图片"})
		return
	}
	if file.Size == 0 || file.Size > maxAvatarBytes {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "头像不能超过 2MB"})
		return
	}
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无法读取头像"})
		return
	}
	header := make([]byte, headerSampleLength)
	_, err = io.ReadFull(src, header)
	src.Close()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "头像文件无效"})
		return
	}
	ext := detectImageType(header)
	if ext == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "仅支持 PNG、JPG 或 WebP 头像"})
		return
	}
	relPath := filepath.Join(avatarSubDir, uuid.NewString()+"."+ext)
	absPath := filepath.Join(configs.AppConfig.UploadDir, relPath)
	if err := c.SaveUploadedFile(file, absPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存头像失败"})
		return
	}
	previous, err := repository.UpdateUserAvatar(c.Request.Context(), userID, relPath, userActivityChannel(c))
	if err != nil {
		_ = os.Remove(absPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新头像失败"})
		return
	}
	if previous != "" && previous != relPath {
		_ = os.Remove(filepath.Join(configs.AppConfig.UploadDir, previous))
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"avatar_url": "/static/uploads/" + relPath}})
}
