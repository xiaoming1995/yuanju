package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	tcpg "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"yuanju/configs"
	"yuanju/internal/middleware"
	"yuanju/internal/model"
	"yuanju/internal/repository"
	"yuanju/internal/service"
	"yuanju/pkg/database"
)

func TestAdminUserActivityHTTPFlow(t *testing.T) {
	ctx := context.Background()
	pg, err := tcpg.Run(ctx, "postgres:16-alpine", tcpg.WithDatabase("yuanju_activity_test"), tcpg.WithUsername("test"), tcpg.WithPassword("test"), testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(60*time.Second)))
	if err != nil {
		t.Fatal(err)
	}
	defer pg.Terminate(ctx)
	dsn, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	oldDB := database.DB
	database.DB = db
	defer func() { database.DB = oldDB }()
	oldConfig := configs.AppConfig
	defer func() { configs.AppConfig = oldConfig }()
	configs.AppConfig.JWTSecret = "activity-user-secret"
	configs.AppConfig.AdminJWTSecret = "activity-admin-secret"
	if _, err := database.Migrate(database.ModeApply); err != nil {
		t.Fatal(err)
	}
	user, _, err := service.Register(service.RegisterInput{Email: "activity-http@example.com", Password: "password123"})
	if err != nil {
		t.Fatal(err)
	}
	var adminID string
	if err := db.QueryRow(`INSERT INTO admins (email,password_hash,name) VALUES ('activity-http-admin@example.com','hash','Test Admin') RETURNING id`).Scan(&adminID); err != nil {
		t.Fatal(err)
	}

	r := gin.New()
	r.POST("/api/auth/login", Login)
	r.POST("/api/bazi/calculate", func(c *gin.Context) { c.Set("user_id", user.ID); Calculate(c) })
	admin := r.Group("/api/admin", middleware.AdminAuth())
	admin.GET("/users", AdminGetUsers)
	admin.GET("/users/:id", AdminGetUserDetail)
	admin.GET("/users/:id/events", AdminGetUserEvents)
	admin.PUT("/users/:id/disable", AdminSetUserDisabled)
	admin.GET("/charts", AdminListCharts)
	request := func(method, path, body, token string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		if body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}
	if w := request(http.MethodPost, "/api/auth/login", `{"email":"activity-http@example.com","password":"wrong"}`, ""); w.Code != http.StatusUnauthorized {
		t.Fatalf("failed login %d", w.Code)
	}
	if w := request(http.MethodPost, "/api/auth/login", `{"email":"activity-http@example.com","password":"password123"}`, ""); w.Code != http.StatusOK {
		t.Fatalf("login %d %s", w.Code, w.Body.String())
	}
	chartResponse := request(http.MethodPost, "/api/bazi/calculate", `{"year":1990,"month":1,"day":1,"hour":12,"gender":"male"}`, "")
	var createdChart struct {
		ChartID string `json:"chart_id"`
	}
	if chartResponse.Code != http.StatusOK || json.Unmarshal(chartResponse.Body.Bytes(), &createdChart) != nil || createdChart.ChartID == "" {
		t.Fatalf("calculate: %d %s", chartResponse.Code, chartResponse.Body.String())
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"iss": "yuanju-admin", "sub": adminID, "email": "activity-http-admin@example.com", "exp": time.Now().Add(time.Hour).Unix()}).SignedString([]byte(configs.AppConfig.AdminJWTSecret))
	if err != nil {
		t.Fatal(err)
	}
	var miniappID string
	if err := db.QueryRow(`INSERT INTO users (source) VALUES ('miniapp_wechat') RETURNING id`).Scan(&miniappID); err != nil {
		t.Fatal(err)
	}
	miniChart, err := repository.CreateChartWithChannel(&model.BaziChart{UserID: &miniappID, BirthYear: 1991, BirthMonth: 2, BirthDay: 2, BirthHour: 2, Gender: "female", YearGan: "辛", YearZhi: "未", MonthGan: "庚", MonthZhi: "寅", DayGan: "甲", DayZhi: "子", HourGan: "乙", HourZhi: "丑", ChartHash: "miniapp-history-chart"}, "miniapp")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO bazi_charts (birth_year,birth_month,birth_day,birth_hour,gender,year_gan,year_zhi,month_gan,month_zhi,day_gan,day_zhi,hour_gan,hour_zhi) VALUES (1992,3,3,3,'male','壬','申','癸','卯','丙','午','丁','巳')`); err != nil {
		t.Fatal(err)
	}
	if w := request(http.MethodGet, "/api/admin/users/"+user.ID+"/events", "", ""); w.Code != http.StatusUnauthorized {
		t.Fatalf("unguarded events %d", w.Code)
	}
	if w := request(http.MethodGet, "/api/admin/charts?user_id="+user.ID, "", ""); w.Code != http.StatusUnauthorized {
		t.Fatalf("unguarded charts %d", w.Code)
	}
	var scoped struct {
		Total int `json:"total"`
		Data  []struct {
			ChartID   *string `json:"chart_id"`
			QueriedAt string  `json:"queried_at"`
			Deleted   bool    `json:"deleted"`
		} `json:"data"`
	}
	w := request(http.MethodGet, "/api/admin/charts?user_id="+user.ID, "", token)
	if w.Code != http.StatusOK || json.Unmarshal(w.Body.Bytes(), &scoped) != nil || scoped.Total != 1 || len(scoped.Data) != 1 || scoped.Data[0].ChartID == nil || *scoped.Data[0].ChartID != createdChart.ChartID || scoped.Data[0].Deleted || scoped.Data[0].QueriedAt == "" {
		t.Fatalf("scoped user charts: %d %s", w.Code, w.Body.String())
	}
	w = request(http.MethodGet, "/api/admin/charts?user_id="+miniappID, "", token)
	if w.Code != http.StatusOK || json.Unmarshal(w.Body.Bytes(), &scoped) != nil || scoped.Total != 1 || len(scoped.Data) != 1 || scoped.Data[0].ChartID == nil || *scoped.Data[0].ChartID != miniChart.ID || scoped.Data[0].Deleted {
		t.Fatalf("scoped miniapp charts: %d %s", w.Code, w.Body.String())
	}
	w = request(http.MethodGet, "/api/admin/charts?user_id="+user.ID+"&from=2999-01-01", "", token)
	if w.Code != http.StatusOK || json.Unmarshal(w.Body.Bytes(), &scoped) != nil || scoped.Total != 0 || len(scoped.Data) != 0 {
		t.Fatalf("scoped date charts: %d %s", w.Code, w.Body.String())
	}
	w = request(http.MethodGet, "/api/admin/charts?user_id=00000000-0000-0000-0000-000000000001", "", token)
	if w.Code != http.StatusOK || json.Unmarshal(w.Body.Bytes(), &scoped) != nil || scoped.Total != 0 || len(scoped.Data) != 0 {
		t.Fatalf("missing user charts: %d %s", w.Code, w.Body.String())
	}
	w = request(http.MethodGet, "/api/admin/charts", "", token)
	if w.Code != http.StatusOK || json.Unmarshal(w.Body.Bytes(), &scoped) != nil || scoped.Total != 3 {
		t.Fatalf("all charts: %d %s", w.Code, w.Body.String())
	}
	w = request(http.MethodGet, "/api/admin/charts?user_id=bad-id", "", token)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("invalid user ID: %d %s", w.Code, w.Body.String())
	}
	w = request(http.MethodGet, "/api/admin/users?activity=active7&sort=last_login_desc", "", token)
	if w.Code != http.StatusOK {
		t.Fatalf("list %d %s", w.Code, w.Body.String())
	}
	var list struct {
		Users []struct {
			ID              string `json:"id"`
			Channel         string `json:"last_login_channel"`
			ChartQueryCount int64  `json:"chart_query_count"`
		} `json:"users"`
		Total int `json:"total"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &list); err != nil || list.Total != 2 {
		t.Fatalf("list %+v err=%v", list, err)
	}
	foundWeb := false
	for _, item := range list.Users {
		if item.ID == user.ID && item.Channel == "web" && item.ChartQueryCount == 1 {
			foundWeb = true
		}
	}
	if !foundWeb {
		t.Fatalf("web user missing from active list: %+v", list)
	}
	w = request(http.MethodGet, "/api/admin/users/"+user.ID, "", token)
	var detail struct {
		Data struct {
			ChartQueryCount int64 `json:"chart_query_count"`
			ChartCount      int   `json:"chart_count"`
		} `json:"data"`
	}
	if w.Code != http.StatusOK || json.Unmarshal(w.Body.Bytes(), &detail) != nil || detail.Data.ChartQueryCount != 1 || detail.Data.ChartCount != 1 {
		t.Fatalf("detail count: %d %s", w.Code, w.Body.String())
	}
	for i := 0; i < 11; i++ {
		if _, err := db.Exec(`INSERT INTO users (email,password_hash) VALUES ($1,'hash')`, "pagination-"+strconv.Itoa(i)+"@example.com"); err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct{ page, size, count int }{{1, 10, 10}, {2, 10, 3}, {1, 20, 13}} {
		w = request(http.MethodGet, "/api/admin/users?page="+strconv.Itoa(tc.page)+"&pageSize="+strconv.Itoa(tc.size), "", token)
		var paged struct {
			Users    []json.RawMessage `json:"users"`
			Total    int               `json:"total"`
			Page     int               `json:"page"`
			PageSize int               `json:"page_size"`
		}
		if w.Code != http.StatusOK || json.Unmarshal(w.Body.Bytes(), &paged) != nil || paged.Total != 13 || paged.Page != tc.page || paged.PageSize != tc.size || len(paged.Users) != tc.count {
			t.Fatalf("pagination page=%d size=%d: %d %s", tc.page, tc.size, w.Code, w.Body.String())
		}
	}
	w = request(http.MethodGet, "/api/admin/users/"+user.ID+"/events", "", token)
	if w.Code != http.StatusOK {
		t.Fatalf("events %d %s", w.Code, w.Body.String())
	}
	var events struct {
		Total  int `json:"total"`
		Events []struct {
			Type string `json:"event_type"`
		} `json:"events"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &events); err != nil || events.Total != 3 || events.Events[0].Type != "chart_created" {
		t.Fatalf("events %+v err=%v", events, err)
	}
	w = request(http.MethodPut, "/api/admin/users/"+user.ID+"/disable", `{"disabled":true}`, token)
	if w.Code != http.StatusOK {
		t.Fatalf("disable %d %s", w.Code, w.Body.String())
	}
	w = request(http.MethodGet, "/api/admin/users/"+user.ID+"/events", "", token)
	if !strings.Contains(w.Body.String(), `"action":"disabled"`) {
		t.Fatalf("audit missing: %s", w.Body.String())
	}
	if _, err := repository.DeleteChart(createdChart.ChartID); err != nil {
		t.Fatal(err)
	}
	w = request(http.MethodGet, "/api/admin/charts?user_id="+user.ID, "", token)
	if w.Code != http.StatusOK || json.Unmarshal(w.Body.Bytes(), &scoped) != nil || scoped.Total != 1 || len(scoped.Data) != 1 || !scoped.Data[0].Deleted || scoped.Data[0].ChartID != nil || scoped.Data[0].QueriedAt == "" || strings.Contains(w.Body.String(), "birth_year") {
		t.Fatalf("deleted query history: %d %s", w.Code, w.Body.String())
	}
	if _, err := db.Exec(`DROP TABLE user_activity_events`); err != nil {
		t.Fatal(err)
	}
	w = request(http.MethodPost, "/api/bazi/calculate", `{"year":1990,"month":1,"day":1,"hour":12,"gender":"male"}`, "")
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("failed save must reject chart query: %d %s", w.Code, w.Body.String())
	}
	if err := db.QueryRow(`SELECT chart_query_count FROM users WHERE id=$1`, user.ID).Scan(&detail.Data.ChartQueryCount); err != nil || detail.Data.ChartQueryCount != 1 {
		t.Fatalf("failed query changed count: %d %v", detail.Data.ChartQueryCount, err)
	}
}
