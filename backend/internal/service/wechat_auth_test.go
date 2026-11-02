package service

import (
	"context"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"testing"
	"yuanju/configs"
	"yuanju/internal/repository"
	"yuanju/pkg/database"
)

type wechatRoundTripFunc func(*http.Request) (*http.Response, error)

func (fn wechatRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return fn(req) }

func TestExchangeWeChatCodeUsesServerSecretAndReturnsOnlyOpenID(t *testing.T) {
	oldClient := wechatHTTPClient
	oldSecret := configs.AppConfig.WeChatMiniAppSecret
	t.Cleanup(func() { wechatHTTPClient = oldClient; configs.AppConfig.WeChatMiniAppSecret = oldSecret })
	configs.AppConfig.WeChatMiniAppSecret = "server-only-secret"
	wechatHTTPClient = &http.Client{Transport: wechatRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Host != "api.weixin.qq.com" || req.URL.Query().Get("secret") != "server-only-secret" || req.URL.Query().Get("js_code") != "one-time-code" {
			t.Errorf("unexpected WeChat request: %s", req.URL.Host+req.URL.Path)
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"openid":"openid-1","session_key":"private"}`))}, nil
	})}
	openID, err := ExchangeWeChatCode(context.Background(), "wx-app", "one-time-code")
	if err != nil || openID != "openid-1" {
		t.Fatalf("openid=%q err=%v", openID, err)
	}
	wechatHTTPClient = &http.Client{Transport: wechatRoundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"errcode":40029}`))}, nil
	})}
	if _, err := ExchangeWeChatCode(context.Background(), "wx-app", "expired"); !errors.Is(err, ErrWeChatCodeInvalid) {
		t.Fatalf("expired code err=%v", err)
	}
}

func testWeChatExchange(_ context.Context, _, code string) (string, error) {
	if code == "bad" {
		return "", errors.New("invalid code")
	}
	return "openid-" + code, nil
}

func TestWeChatLoginKeepsEmailUsersAndUsesOneIdentity(t *testing.T) {
	withAuthTestDB(t)
	configs.AppConfig.WeChatMiniAppID = "wx-test-app"
	if err := repository.SetBoolSetting(repository.SettingMiniappRegistrationEnabled, false); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := LoginWithWeChatCode(context.Background(), "new", testWeChatExchange); !errors.Is(err, ErrWeChatRegistrationDisabled) {
		t.Fatalf("disabled miniapp registration err=%v", err)
	}
	if err := repository.SetBoolSetting(repository.SettingMiniappRegistrationEnabled, true); err != nil {
		t.Fatal(err)
	}

	old, _, err := Register(RegisterInput{Email: "old@example.com", Password: "password123"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := BindWeChatCode(context.Background(), old.ID, "old", testWeChatExchange); err != nil {
		t.Fatal(err)
	}
	bound, token, isNew, err := LoginWithWeChatCode(context.Background(), "old", testWeChatExchange)
	if err != nil || token == "" || bound.ID != old.ID || isNew {
		t.Fatalf("bound login user=%+v token=%q err=%v", bound, token, err)
	}
	web, _, err := Login(LoginInput{Email: "old@example.com", Password: "password123"})
	if err != nil || web.ID != old.ID {
		t.Fatalf("web login user=%+v err=%v", web, err)
	}
	var chartID string
	if err := database.DB.QueryRow(`INSERT INTO bazi_charts (user_id, birth_year, birth_month, birth_day, birth_hour,
		year_gan, year_zhi, month_gan, month_zhi, day_gan, day_zhi, hour_gan, hour_zhi, yongshen, jishen, chart_hash)
		VALUES ($1, 1990, 1, 1, 1, '己', '巳', '丙', '子', '丙', '寅', '己', '丑', '', '', 'wechat-shared-chart')
		RETURNING id`, old.ID).Scan(&chartID); err != nil {
		t.Fatal(err)
	}
	sharedCharts, err := repository.GetChartsByUserID(bound.ID, 20, 0)
	if err != nil || len(sharedCharts) != 1 || sharedCharts[0].ID != chartID {
		t.Fatalf("bound WeChat user did not see web history: charts=%+v err=%v", sharedCharts, err)
	}

	created, _, isNew, err := LoginWithWeChatCode(context.Background(), "new", testWeChatExchange)
	if err != nil || created.Email != "" || created.Source != "miniapp_wechat" || !isNew {
		t.Fatalf("new miniapp user=%+v err=%v", created, err)
	}
	if !regexp.MustCompile(`^缘友[0-9]{6}$`).MatchString(created.Nickname) {
		t.Fatalf("unexpected generated nickname %q", created.Nickname)
	}
	returned, _, isNew, err := LoginWithWeChatCode(context.Background(), "new", testWeChatExchange)
	if err != nil || isNew || returned.ID != created.ID || returned.Nickname != created.Nickname {
		t.Fatalf("returning miniapp user=%+v isNew=%v err=%v", returned, isNew, err)
	}
	if err := repository.UpdateUserNickname(context.Background(), created.ID, "我的新昵称"); err != nil {
		t.Fatal(err)
	}
	returned, _, _, err = LoginWithWeChatCode(context.Background(), "new", testWeChatExchange)
	if err != nil || returned.Nickname != "我的新昵称" {
		t.Fatalf("returning login reset nickname: user=%+v err=%v", returned, err)
	}
	var email, hash *string
	if err := database.DB.QueryRow(`SELECT email, password_hash FROM users WHERE id=$1`, created.ID).Scan(&email, &hash); err != nil {
		t.Fatal(err)
	}
	if email != nil || hash != nil {
		t.Fatalf("miniapp user must have no synthetic credentials, email=%v hash=%v", email, hash)
	}
	if _, err := BindWeChatCode(context.Background(), old.ID, "new", testWeChatExchange); !errors.Is(err, ErrWeChatIdentityConflict) {
		t.Fatalf("conflicting bind err=%v", err)
	}
	if err := repository.SetUserDisabled(created.ID, true); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := LoginWithWeChatCode(context.Background(), "new", testWeChatExchange); !errors.Is(err, ErrWeChatAccountDisabled) {
		t.Fatalf("disabled login err=%v", err)
	}
}

func TestConcurrentWeChatLoginCreatesOneUser(t *testing.T) {
	withAuthTestDB(t)
	configs.AppConfig.WeChatMiniAppID = "wx-test-app"
	if err := repository.SetBoolSetting(repository.SettingMiniappRegistrationEnabled, true); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	ids := make(chan string, 8)
	newUsers := make(chan bool, 8)
	errs := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			user, _, isNewUser, err := LoginWithWeChatCode(context.Background(), "race", testWeChatExchange)
			if err != nil {
				errs <- err
				return
			}
			ids <- user.ID
			newUsers <- isNewUser
		}()
	}
	wg.Wait()
	close(ids)
	close(newUsers)
	close(errs)
	for err := range errs {
		t.Fatalf("concurrent login: %v", err)
	}
	first := ""
	for id := range ids {
		if first == "" {
			first = id
		} else if id != first {
			t.Fatalf("duplicate users %s and %s", first, id)
		}
	}
	if first == "" {
		t.Fatal("no successful login")
	}
	createdCount := 0
	for isNewUser := range newUsers {
		if isNewUser {
			createdCount++
		}
	}
	if createdCount != 1 {
		t.Fatalf("new user flag count=%d, want 1", createdCount)
	}
}

func TestEmailLessMiniappUserCanBeReadByProfileAndAdmin(t *testing.T) {
	withAuthTestDB(t)
	configs.AppConfig.WeChatMiniAppID = "wx-test-app"
	if err := repository.SetBoolSetting(repository.SettingMiniappRegistrationEnabled, true); err != nil {
		t.Fatal(err)
	}
	user, _, _, err := LoginWithWeChatCode(context.Background(), "profile", testWeChatExchange)
	if err != nil {
		t.Fatal(err)
	}
	profile, err := GetUserProfileOverview(user.ID)
	if err != nil || profile.User.Email != "" || profile.User.ID != user.ID {
		t.Fatalf("profile=%+v err=%v", profile, err)
	}
	users, total, err := repository.ListAdminUsers(1, 20, user.Nickname)
	if err != nil || total != 1 || len(users) != 1 || users[0].ID != user.ID {
		t.Fatalf("admin users=%+v total=%d err=%v", users, total, err)
	}
	if err := ResetUserPassword(user.ID, "newpassword123"); !errors.Is(err, ErrEmailCredentialsRequired) {
		t.Fatalf("email-less password reset err=%v", err)
	}
	if err := repository.UpdateUserNickname(context.Background(), user.ID, "我的微信昵称"); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.UpdateUserAvatar(context.Background(), user.ID, "user-avatars/test-avatar.png"); err != nil {
		t.Fatal(err)
	}
	profile, err = GetUserProfileOverview(user.ID)
	if err != nil || profile == nil || profile.User.Nickname != "我的微信昵称" || profile.User.AvatarURL != "/static/uploads/user-avatars/test-avatar.png" {
		t.Fatalf("saved miniapp profile=%+v err=%v", profile, err)
	}
	if _, err := database.DB.Exec(`INSERT INTO compatibility_readings
		(user_id, overall_level, dimension_scores) VALUES ($1, 'neutral', '{}'::jsonb)`, user.ID); err != nil {
		t.Fatal(err)
	}
	compat, totalCompat, err := repository.AdminListCompatibilityReadings(1, 20)
	if err != nil || totalCompat != 1 || len(compat) != 1 || compat[0].UserID == nil || *compat[0].UserID != user.ID || compat[0].UserEmail != nil {
		t.Fatalf("email-less compatibility row=%+v total=%d err=%v", compat, totalCompat, err)
	}
}

func TestLinkWebAccountMergesMiniappHistoryIntoExistingWebUser(t *testing.T) {
	withAuthTestDB(t)
	configs.AppConfig.WeChatMiniAppID = "wx-test-app"
	web, _, err := Register(RegisterInput{Email: "merge@example.com", Password: "password123"})
	if err != nil {
		t.Fatal(err)
	}
	miniapp, _, _, err := LoginWithWeChatCode(context.Background(), "merge", testWeChatExchange)
	if err != nil {
		t.Fatal(err)
	}
	if miniapp.ID == web.ID {
		t.Fatal("miniapp login must create a separate user before linking")
	}
	if _, _, err := LinkWebAccount(context.Background(), miniapp.ID, web.Email, "wrong"); err == nil {
		t.Fatal("incorrect web password should not merge")
	}
	insertChart := func(userID, hash string) {
		t.Helper()
		_, err := database.DB.Exec(`INSERT INTO bazi_charts
		 (user_id, birth_year, birth_month, birth_day, birth_hour, gender,
		 year_gan, year_zhi, month_gan, month_zhi, day_gan, day_zhi, hour_gan, hour_zhi,
		 yongshen, jishen, chart_hash)
		 VALUES ($1, 1990, 1, 1, 1, 'male', '己', '巳', '丙', '子', '丙', '寅', '己', '丑', '', '', $2)`, userID, hash)
		if err != nil {
			t.Fatal(err)
		}
	}
	insertChart(web.ID, "web-chart")
	insertChart(miniapp.ID, "miniapp-chart")
	if _, err := repository.UpdateUserAvatar(context.Background(), miniapp.ID, "user-avatars/merge-avatar.png"); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{web.ID, miniapp.ID} {
		if _, err := database.DB.Exec(`INSERT INTO compatibility_readings
			(user_id, overall_level, dimension_scores) VALUES ($1, 'neutral', '{}'::jsonb)`, id); err != nil {
			t.Fatal(err)
		}
	}
	for _, id := range []string{web.ID, miniapp.ID} {
		_, err := database.DB.Exec(`INSERT INTO token_usage_logs_monthly
		 (user_id, model, year_month, call_count, total_tokens)
		 VALUES ($1, 'model-a', '2026-10', 1, 10)`, id)
		if err != nil {
			t.Fatal(err)
		}
	}
	linked, token, err := LinkWebAccount(context.Background(), miniapp.ID, web.Email, "password123")
	if err != nil || linked.ID != web.ID || token == "" {
		t.Fatalf("link user=%+v tokenPresent=%t err=%v", linked, token != "", err)
	}
	charts, err := repository.GetChartsByUserID(web.ID, 20, 0)
	if err != nil || len(charts) != 2 {
		t.Fatalf("merged charts=%d err=%v", len(charts), err)
	}
	var compatibilityCount int
	if err := database.DB.QueryRow(`SELECT COUNT(*) FROM compatibility_readings WHERE user_id=$1`, web.ID).Scan(&compatibilityCount); err != nil || compatibilityCount != 2 {
		t.Fatalf("merged compatibility=%d err=%v", compatibilityCount, err)
	}
	var calls, tokens int
	if err := database.DB.QueryRow(`SELECT call_count, total_tokens FROM token_usage_logs_monthly
		 WHERE user_id=$1 AND model='model-a' AND year_month='2026-10'`, web.ID).Scan(&calls, &tokens); err != nil || calls != 2 || tokens != 20 {
		t.Fatalf("merged token usage calls=%d tokens=%d err=%v", calls, tokens, err)
	}
	if old, err := repository.GetUserByID(miniapp.ID); err != nil || old != nil {
		t.Fatalf("old miniapp user=%+v err=%v", old, err)
	}
	if profile, err := repository.GetUserByID(web.ID); err != nil || profile.AvatarURL != "/static/uploads/user-avatars/merge-avatar.png" {
		t.Fatalf("merged avatar=%+v err=%v", profile, err)
	}
	wechatUser, _, _, err := LoginWithWeChatCode(context.Background(), "merge", testWeChatExchange)
	if err != nil || wechatUser.ID != web.ID {
		t.Fatalf("later WeChat login=%+v err=%v", wechatUser, err)
	}
	webUser, _, err := Login(LoginInput{Email: web.Email, Password: "password123"})
	if err != nil || webUser.ID != web.ID {
		t.Fatalf("later web login=%+v err=%v", webUser, err)
	}
}

func TestLinkWebAccountConflictLeavesMiniappHistoryUntouched(t *testing.T) {
	withAuthTestDB(t)
	configs.AppConfig.WeChatMiniAppID = "wx-test-app"
	web, _, err := Register(RegisterInput{Email: "occupied@example.com", Password: "password123"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := BindWeChatCode(context.Background(), web.ID, "already-bound", testWeChatExchange); err != nil {
		t.Fatal(err)
	}
	miniapp, _, _, err := LoginWithWeChatCode(context.Background(), "other", testWeChatExchange)
	if err != nil {
		t.Fatal(err)
	}
	var chartID string
	if err := database.DB.QueryRow(`INSERT INTO bazi_charts
		 (user_id, birth_year, birth_month, birth_day, birth_hour, gender)
		 VALUES ($1, 1990, 1, 1, 1, 'male') RETURNING id`, miniapp.ID).Scan(&chartID); err != nil {
		t.Fatal(err)
	}
	if _, _, err := LinkWebAccount(context.Background(), miniapp.ID, web.Email, "password123"); !errors.Is(err, ErrWeChatIdentityConflict) {
		t.Fatalf("link conflict err=%v", err)
	}
	var owner string
	if err := database.DB.QueryRow(`SELECT user_id FROM bazi_charts WHERE id=$1`, chartID).Scan(&owner); err != nil || owner != miniapp.ID {
		t.Fatalf("chart owner=%q err=%v", owner, err)
	}
	wechatUser, _, _, err := LoginWithWeChatCode(context.Background(), "other", testWeChatExchange)
	if err != nil || wechatUser.ID != miniapp.ID {
		t.Fatalf("WeChat account=%+v err=%v", wechatUser, err)
	}
}
