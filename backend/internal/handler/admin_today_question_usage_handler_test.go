package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"yuanju/internal/repository"
	"yuanju/pkg/database"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestAdminUserListTodayQuestionUsage(t *testing.T) {
	withArticleHandlerDB(t)
	if err := repository.SetTodayQuestionDailyLimit(2); err != nil {
		t.Fatal(err)
	}
	var userID, otherID, adminID string
	for _, target := range []*string{&userID, &otherID} {
		if err := database.DB.QueryRow(`INSERT INTO users (email,password_hash) VALUES ($1,'hash') RETURNING id`, uuid.NewString()+"@example.com").Scan(target); err != nil {
			t.Fatal(err)
		}
	}
	if err := database.DB.QueryRow(`INSERT INTO admins (email,password_hash,name) VALUES ('usage-admin@example.com','hash','Usage Admin') RETURNING id`).Scan(&adminID); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.GrantTodayQuestionCredits(userID, adminID, 3, "测试用量", uuid.NewString()); err != nil {
		t.Fatal(err)
	}
	beijing, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}
	today := time.Now().In(beijing).Format("2006-01-02")
	yesterday := time.Now().In(beijing).AddDate(0, 0, -1).Format("2006-01-02")
	reserve := func(date string) string {
		t.Helper()
		id, _, err := repository.ReserveTodayQuestionRequest(userID, uuid.NewString(), date, "general", "今天有一件事", []byte(`{"evidence":[]}`))
		if err != nil {
			t.Fatal(err)
		}
		return id
	}
	complete := func(id, status string) {
		t.Helper()
		if err := repository.CompleteTodayQuestionRequest(id, repository.TodayQuestionCompletion{Status: status}); err != nil {
			t.Fatal(err)
		}
	}
	complete(reserve(today), "success")
	freeProcessingID := reserve(today)
	complete(reserve(today), "insufficient")
	complete(reserve(today), "error")
	complete(reserve(yesterday), "success")

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/admin/users", AdminGetUsers)
	list := func(query string) (int, []repository.AdminUserRow) {
		t.Helper()
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/admin/users?q="+query, nil))
		if response.Code != http.StatusOK {
			t.Fatalf("list status=%d body=%s", response.Code, response.Body.String())
		}
		var body struct {
			Users []repository.AdminUserRow `json:"users"`
			Total int                       `json:"total"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		return body.Total, body.Users
	}
	if total, users := list(userID); total != 1 || len(users) != 1 || users[0].ID != userID {
		t.Fatalf("filtered users: total=%d users=%+v", total, users)
	} else if got := users[0].TodayQuestion; got.Date != today || got.DailyLimit != 2 || got.Used != 2 || got.FreeUsed != 1 || got.Processing != 1 || got.FreeProcessing != 1 || got.CreditBalance != 2 || got.Remaining != 2 {
		t.Fatalf("today usage: %+v", got)
	}
	if total, users := list(otherID); total != 1 || len(users) != 1 || users[0].TodayQuestion.Used != 0 || users[0].TodayQuestion.Remaining != 2 {
		t.Fatalf("other user's usage: total=%d users=%+v", total, users)
	}
	creditProcessingID := reserve(today)
	if _, err := database.DB.Exec(`UPDATE today_question_requests SET created_at=NOW()-INTERVAL '11 minutes' WHERE id IN ($1,$2)`, freeProcessingID, creditProcessingID); err != nil {
		t.Fatal(err)
	}
	if _, users := list(userID); users[0].TodayQuestion.Processing != 0 || users[0].TodayQuestion.CreditBalance != 2 || users[0].TodayQuestion.Remaining != 3 {
		t.Fatalf("stale reservation not returned: %+v", users[0].TodayQuestion)
	}
	if err := repository.SetTodayQuestionDailyLimit(3); err != nil {
		t.Fatal(err)
	}
	if _, users := list(userID); users[0].TodayQuestion.DailyLimit != 3 || users[0].TodayQuestion.Remaining != 4 {
		t.Fatalf("new daily limit: %+v", users[0].TodayQuestion)
	}
}
