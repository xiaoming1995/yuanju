package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	"yuanju/configs"
	"yuanju/internal/model"
	"yuanju/internal/repository"
)

var (
	ErrWeChatRegistrationDisabled = errors.New("小程序新用户注册暂未开放")
	ErrWeChatAccountDisabled      = errors.New("该账号已被禁用")
	ErrWeChatIdentityConflict     = repository.ErrWeChatIdentityConflict
	ErrWeChatMergeConflict        = repository.ErrWeChatMergeConflict
	ErrWeChatCodeInvalid          = errors.New("微信登录凭证无效或已过期")
	ErrWeChatUnavailable          = errors.New("微信登录服务暂不可用")
	ErrWeChatConfigMissing        = errors.New("微信小程序服务端凭证未配置")
)

type WeChatCodeExchangeFunc func(context.Context, string, string) (string, error)

var wechatHTTPClient = &http.Client{Timeout: 5 * time.Second}

func ExchangeWeChatCode(ctx context.Context, appID, code string) (string, error) {
	secret := configs.AppConfig.WeChatMiniAppSecret
	if appID == "" || secret == "" {
		return "", ErrWeChatConfigMissing
	}
	query := url.Values{
		"appid":      {appID},
		"secret":     {secret},
		"js_code":    {code},
		"grant_type": {"authorization_code"},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.weixin.qq.com/sns/jscode2session?"+query.Encode(), nil)
	if err != nil {
		return "", err
	}
	resp, err := wechatHTTPClient.Do(req)
	if err != nil {
		return "", ErrWeChatUnavailable
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w：状态 %d", ErrWeChatUnavailable, resp.StatusCode)
	}
	var result struct {
		OpenID  string `json:"openid"`
		ErrCode int    `json:"errcode"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 4096)).Decode(&result); err != nil {
		return "", ErrWeChatUnavailable
	}
	if result.ErrCode != 0 || result.OpenID == "" {
		return "", ErrWeChatCodeInvalid
	}
	return result.OpenID, nil
}

func LoginWithWeChatCode(ctx context.Context, code string, exchange WeChatCodeExchangeFunc) (*model.User, string, bool, error) {
	appID := configs.AppConfig.WeChatMiniAppID
	if appID == "" {
		return nil, "", false, ErrWeChatConfigMissing
	}
	if strings.TrimSpace(code) == "" || len(code) > 1024 {
		return nil, "", false, ErrWeChatCodeInvalid
	}
	openID, err := exchange(ctx, appID, code)
	if err != nil {
		return nil, "", false, err
	}
	user, err := repository.GetUserByWeChatIdentity(appID, openID)
	if err != nil {
		return nil, "", false, err
	}
	isNewUser := false
	if user == nil {
		enabled, err := repository.GetBoolSetting(repository.SettingMiniappRegistrationEnabled, true)
		if err != nil {
			return nil, "", false, err
		}
		if !enabled {
			return nil, "", false, ErrWeChatRegistrationDisabled
		}
		user, isNewUser, err = repository.CreateUserWithWeChatIdentity(ctx, appID, openID)
		if err != nil {
			return nil, "", false, err
		}
	}
	if user == nil {
		return nil, "", false, errors.New("微信登录账号不存在")
	}
	if user.DisabledAt != nil {
		return nil, "", false, ErrWeChatAccountDisabled
	}
	token, err := generateJWTForChannel(user, "miniapp")
	if err != nil {
		return nil, "", false, err
	}
	if err := repository.RecordUserActivity(ctx, user.ID, "login", "miniapp", nil); err != nil {
		return nil, "", false, err
	}
	return user, token, isNewUser, nil
}

func BindWeChatCode(ctx context.Context, userID, code string, exchange WeChatCodeExchangeFunc) (*model.User, error) {
	appID := configs.AppConfig.WeChatMiniAppID
	if appID == "" {
		return nil, ErrWeChatConfigMissing
	}
	if strings.TrimSpace(code) == "" || len(code) > 1024 {
		return nil, ErrWeChatCodeInvalid
	}
	user, err := repository.GetUserByID(userID)
	if err != nil || user == nil {
		return nil, errors.New("账号不存在")
	}
	if user.DisabledAt != nil {
		return nil, ErrWeChatAccountDisabled
	}
	openID, err := exchange(ctx, appID, code)
	if err != nil {
		return nil, err
	}
	if err := repository.BindWeChatIdentity(ctx, userID, appID, openID); err != nil {
		return nil, err
	}
	return user, nil
}

// LinkWebAccount requires an authenticated miniapp session and valid web
// credentials. The web user remains the account owner after the records merge.
func LinkWebAccount(ctx context.Context, miniappUserID, email, password string) (*model.User, string, error) {
	webUser, _, err := Login(LoginInput{Email: email, Password: password})
	if err != nil {
		return nil, "", err
	}
	if miniappUserID != webUser.ID {
		if err := repository.MergeMiniappIntoWebUser(ctx, miniappUserID, webUser.ID, configs.AppConfig.WeChatMiniAppID); err != nil {
			return nil, "", err
		}
	}
	if miniappUserID == webUser.ID {
		if err := repository.RecordUserActivity(ctx, webUser.ID, "account_linked", "miniapp", nil); err != nil {
			return nil, "", err
		}
	}
	token, err := generateJWTForChannel(webUser, "miniapp")
	return webUser, token, err
}
