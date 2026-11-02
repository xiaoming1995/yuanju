package handler

import (
	"errors"
	"net/http"
	"yuanju/internal/repository"
	"yuanju/internal/service"

	"github.com/gin-gonic/gin"
)

func WeChatLogin(c *gin.Context) {
	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "微信登录凭证必填"})
		return
	}
	user, token, isNewUser, err := service.LoginWithWeChatCode(c.Request.Context(), req.Code, service.ExchangeWeChatCode)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrWeChatCodeInvalid):
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrWeChatConfigMissing):
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "微信登录尚未配置"})
		case errors.Is(err, service.ErrWeChatRegistrationDisabled):
			c.JSON(http.StatusForbidden, gin.H{"error": "小程序新用户注册暂未开放", "code": "miniapp_registration_disabled"})
		case errors.Is(err, service.ErrWeChatAccountDisabled):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error(), "code": "account_disabled"})
		default:
			c.JSON(http.StatusBadGateway, gin.H{"error": "微信登录失败，请重试"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user, "token": token, "is_new_user": isNewUser})
}

func WeChatBind(c *gin.Context) {
	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "微信登录凭证必填"})
		return
	}
	userID, _ := c.Get("user_id")
	user, err := service.BindWeChatCode(c.Request.Context(), userID.(string), req.Code, service.ExchangeWeChatCode)
	if err != nil {
		if errors.Is(err, service.ErrWeChatIdentityConflict) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, service.ErrWeChatCodeInvalid) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, service.ErrWeChatConfigMissing) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "微信登录尚未配置"})
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{"error": "绑定微信失败，请重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}

func WeChatLinkWebAccount(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "请填写网页账号邮箱和密码"})
		return
	}
	userID, _ := c.Get("user_id")
	user, token, err := service.LinkWebAccount(c.Request.Context(), userID.(string), req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrWeChatIdentityConflict), errors.Is(err, service.ErrWeChatMergeConflict):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case err.Error() == "邮箱或密码错误":
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		case err.Error() == "该账号已被禁用":
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "绑定网页账号失败，请重试"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user, "token": token})
}

func MiniappRegistrationSettings(c *gin.Context) {
	enabled, err := repository.GetBoolSetting(repository.SettingMiniappRegistrationEnabled, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取注册设置失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"miniapp_registration_enabled": enabled})
}

func AdminGetMiniappRegistrationSetting(c *gin.Context) {
	MiniappRegistrationSettings(c)
}

func AdminUpdateMiniappRegistrationSetting(c *gin.Context) {
	var req struct {
		Enabled *bool `json:"miniapp_registration_enabled" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Enabled == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "miniapp_registration_enabled 必填"})
		return
	}
	if err := repository.SetBoolSetting(repository.SettingMiniappRegistrationEnabled, *req.Enabled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存注册设置失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"miniapp_registration_enabled": *req.Enabled})
}
