package handler

import (
	"errors"
	"net/http"
	"yuanju/internal/repository"
	"yuanju/internal/service"

	"github.com/gin-gonic/gin"
)

func Register(c *gin.Context) {
	var input service.RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "请检查输入信息：" + err.Error()})
		return
	}

	user, token, err := service.Register(input)
	if err != nil {
		if err.Error() == "该邮箱已被注册" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, service.ErrRegistrationDisabled) {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "注册失败，请稍后重试"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user":  user,
		"token": token,
	})
}

func RegistrationSettings(c *gin.Context) {
	enabled, err := repository.GetBoolSetting(repository.SettingRegistrationEnabled, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取注册设置失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"registration_enabled": enabled})
}

func Login(c *gin.Context) {
	var input service.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "请检查输入信息"})
		return
	}

	user, token, err := service.Login(input)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":  user,
		"token": token,
	})
}

func Me(c *gin.Context) {
	userID, _ := c.Get("user_id")
	user, err := repository.GetUserByID(userID.(string))
	if err != nil || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}
