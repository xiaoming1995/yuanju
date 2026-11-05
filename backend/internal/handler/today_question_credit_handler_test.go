package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
	"yuanju/configs"
	"yuanju/internal/middleware"
	"yuanju/internal/repository"
	"yuanju/pkg/database"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestTodayQuestionCreditGrantAndConsumption(t *testing.T) {
	_ = withArticleHandlerDB(t)
	if err := repository.SetTodayQuestionDailyLimit(1); err != nil {
		t.Fatal(err)
	}
	var userID, adminID string
	if err := database.DB.QueryRow(`INSERT INTO users (email,password_hash) VALUES ('credit-user@example.com','hash') RETURNING id`).Scan(&userID); err != nil {
		t.Fatal(err)
	}
	if err := database.DB.QueryRow(`INSERT INTO admins (email,password_hash,name) VALUES ('credit-admin@example.com','hash','Credit Admin') RETURNING id`).Scan(&adminID); err != nil {
		t.Fatal(err)
	}
	adminAPI := gin.New()
	admin := adminAPI.Group("/api/admin", middleware.AdminAuth())
	admin.GET("/users/:id/today-question-credits", AdminGetTodayQuestionCredits)
	admin.POST("/users/:id/today-question-credits/grants", AdminGrantTodayQuestionCredits)
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": "yuanju-admin", "sub": adminID, "exp": time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte(configs.AppConfig.AdminJWTSecret))
	if err != nil {
		t.Fatal(err)
	}
	path := "/api/admin/users/" + userID + "/today-question-credits"
	call := func(method, path, body, auth string) *httptest.ResponseRecorder {
		t.Helper()
		request := httptest.NewRequest(method, path, strings.NewReader(body))
		if auth != "" {
			request.Header.Set("Authorization", "Bearer "+auth)
		}
		response := httptest.NewRecorder()
		adminAPI.ServeHTTP(response, request)
		return response
	}
	if denied := call(http.MethodGet, path, "", ""); denied.Code != http.StatusUnauthorized {
		t.Fatalf("unprotected read: %d", denied.Code)
	}
	if denied := call(http.MethodPost, path+"/grants", `{}`, ""); denied.Code != http.StatusUnauthorized {
		t.Fatalf("unprotected grant: %d", denied.Code)
	}
	key := uuid.NewString()
	grantBody := `{"amount":3,"reason":"测试补偿","idempotency_key":"` + key + `"}`
	for _, body := range []string{
		`{"amount":0,"reason":"测试","idempotency_key":"` + uuid.NewString() + `"}`,
		`{"amount":1001,"reason":"测试","idempotency_key":"` + uuid.NewString() + `"}`,
		`{"amount":1,"reason":"","idempotency_key":"` + uuid.NewString() + `"}`,
		`{"amount":1.5,"reason":"测试","idempotency_key":"` + uuid.NewString() + `"}`,
	} {
		if invalid := call(http.MethodPost, path+"/grants", body, token); invalid.Code != http.StatusBadRequest {
			t.Fatalf("invalid grant %s: %d %s", body, invalid.Code, invalid.Body.String())
		}
	}
	if missing := call(http.MethodPost, "/api/admin/users/"+uuid.NewString()+"/today-question-credits/grants", grantBody, token); missing.Code != http.StatusNotFound {
		t.Fatalf("missing user: %d %s", missing.Code, missing.Body.String())
	}
	if grant := call(http.MethodPost, path+"/grants", grantBody, token); grant.Code != http.StatusOK || !strings.Contains(grant.Body.String(), `"balance_after":3`) {
		t.Fatalf("grant: %d %s", grant.Code, grant.Body.String())
	}
	if retry := call(http.MethodPost, path+"/grants", grantBody, token); retry.Code != http.StatusOK || !strings.Contains(retry.Body.String(), `"balance_after":3`) {
		t.Fatalf("idempotent retry: %d %s", retry.Code, retry.Body.String())
	}
	if conflict := call(http.MethodPost, path+"/grants", `{"amount":2,"reason":"测试补偿","idempotency_key":"`+key+`"}`, token); conflict.Code != http.StatusConflict {
		t.Fatalf("key conflict: %d %s", conflict.Code, conflict.Body.String())
	}
	account := func() repository.TodayQuestionCreditAccount {
		t.Helper()
		response := call(http.MethodGet, path, "", token)
		if response.Code != http.StatusOK {
			t.Fatalf("account: %d %s", response.Code, response.Body.String())
		}
		var body struct {
			Data repository.TodayQuestionCreditAccount `json:"data"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		return body.Data
	}
	if got := account(); got.Balance != 3 || len(got.Entries) != 1 || got.Entries[0].AdminID != adminID {
		t.Fatalf("grant audit: %+v", got)
	}
	reserve := func(date string) (string, error) {
		id, _, err := repository.ReserveTodayQuestionRequest(userID, uuid.NewString(), date, "general", "今天有一件事情", []byte(`{"evidence":[]}`))
		return id, err
	}
	complete := func(id, status string) {
		t.Helper()
		if err := repository.CompleteTodayQuestionRequest(id, repository.TodayQuestionCompletion{Status: status}); err != nil {
			t.Fatal(err)
		}
	}
	freeID, err := reserve("2026-11-05")
	if err != nil {
		t.Fatal(err)
	}
	complete(freeID, "insufficient")
	if got := account().Balance; got != 3 {
		t.Fatalf("free request spent credits: %d", got)
	}
	creditID, err := reserve("2026-11-05")
	if err != nil {
		t.Fatal(err)
	}
	complete(creditID, "success")
	if got := account().Balance; got != 2 {
		t.Fatalf("credit spend: %d", got)
	}
	failID, err := reserve("2026-11-05")
	if err != nil {
		t.Fatal(err)
	}
	complete(failID, "error")
	if got := account().Balance; got != 2 {
		t.Fatalf("credit refund: %d", got)
	}
	if err := repository.CompleteTodayQuestionRequest(failID, repository.TodayQuestionCompletion{Status: "error"}); err == nil {
		t.Fatal("duplicate completion succeeded")
	}
	nextFreeID, err := reserve("2026-11-06")
	if err != nil {
		t.Fatal(err)
	}
	complete(nextFreeID, "success")
	if got := account().Balance; got != 2 {
		t.Fatalf("next-day free spent credits: %d", got)
	}

	// Use one credit, then race two submissions for the last remaining credit.
	usedID, err := reserve("2026-11-05")
	if err != nil {
		t.Fatal(err)
	}
	complete(usedID, "success")
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for range 2 {
		wg.Add(1)
		go func() { defer wg.Done(); _, err := reserve("2026-11-06"); results <- err }()
	}
	wg.Wait()
	close(results)
	accepted, rejected := 0, 0
	for result := range results {
		if result == nil {
			accepted++
		} else if errors.Is(result, repository.ErrTodayQuestionDailyLimit) {
			rejected++
		} else {
			t.Fatal(result)
		}
	}
	if accepted != 1 || rejected != 1 || account().Balance != 0 {
		t.Fatalf("concurrent credit use: accepted=%d rejected=%d balance=%d", accepted, rejected, account().Balance)
	}
	if _, err := reserve("2026-11-06"); !errors.Is(err, repository.ErrTodayQuestionDailyLimit) {
		t.Fatalf("exhausted credits: %v", err)
	}
	if extra := call(http.MethodPost, path+"/grants", `{"amount":1,"reason":"再次补偿","idempotency_key":"`+uuid.NewString()+`"}`, token); extra.Code != http.StatusOK {
		t.Fatalf("second grant: %d %s", extra.Code, extra.Body.String())
	}
	staleID, err := reserve("2026-11-06")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := database.DB.Exec(`UPDATE today_question_requests SET created_at=NOW()-INTERVAL '11 minutes' WHERE id=$1`, staleID); err != nil {
		t.Fatal(err)
	}
	// A later-day request releases the old credit before using its fresh free quota.
	laterID, err := reserve("2026-11-07")
	if err != nil {
		t.Fatal(err)
	}
	complete(laterID, "success")
	if got := account().Balance; got != 1 {
		t.Fatalf("cross-day stale refund: %d", got)
	}
	var refunds int
	if err := database.DB.QueryRow(`SELECT count(*) FROM today_question_credit_entries WHERE request_id=$1 AND kind='refund'`, staleID).Scan(&refunds); err != nil || refunds != 1 {
		t.Fatalf("stale refund count: %d %v", refunds, err)
	}
}
