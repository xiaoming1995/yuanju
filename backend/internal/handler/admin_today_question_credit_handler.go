package handler

import (
	"errors"
	"net/http"
	"strings"
	"unicode/utf8"
	"yuanju/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func AdminGetTodayQuestionCredits(c *gin.Context) {
	userID := c.Param("id")
	if _, err := uuid.Parse(userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户 ID 无效"})
		return
	}
	account, err := repository.GetTodayQuestionCreditAccount(userID)
	if errors.Is(err, repository.ErrAdminUserMissing) {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取额外次数失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": account})
}

func AdminGrantTodayQuestionCredits(c *gin.Context) {
	userID := c.Param("id")
	if _, err := uuid.Parse(userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户 ID 无效"})
		return
	}
	var input struct {
		Amount         int    `json:"amount"`
		Reason         string `json:"reason"`
		IdempotencyKey string `json:"idempotency_key"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "发放参数无效"})
		return
	}
	input.Reason = strings.TrimSpace(input.Reason)
	if input.Amount < 1 || input.Amount > 1000 || input.Reason == "" || utf8.RuneCountInString(input.Reason) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "增加次数须为 1—1000 的整数，原因须为 1—200 字"})
		return
	}
	if _, err := uuid.Parse(input.IdempotencyKey); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求标识无效"})
		return
	}
	adminID := c.GetString("admin_id")
	if _, err := uuid.Parse(adminID); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "管理员身份无效"})
		return
	}
	entry, err := repository.GrantTodayQuestionCredits(userID, adminID, input.Amount, input.Reason, input.IdempotencyKey)
	if errors.Is(err, repository.ErrAdminUserMissing) {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}
	if errors.Is(err, repository.ErrTodayQuestionCreditKeyConflict) {
		c.JSON(http.StatusConflict, gin.H{"error": "请求标识已用于其他发放，请重新打开弹框"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "增加次数失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": entry})
}
