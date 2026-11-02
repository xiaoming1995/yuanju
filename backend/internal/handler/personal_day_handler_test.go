package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"yuanju/internal/middleware"

	"github.com/gin-gonic/gin"
)

func personalHandlerRouter(userID string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	group := r.Group("/api/user", func(c *gin.Context) { c.Set("user_id", userID); c.Next() })
	group.GET("/personal-chart", GetPersonalChart)
	group.PUT("/personal-chart", SetPersonalChart)
	group.DELETE("/personal-chart", ClearPersonalChart)
	group.GET("/personal-day", GetPersonalDay)
	return r
}

func personalRequest(r *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	out := httptest.NewRecorder()
	r.ServeHTTP(out, req)
	return out
}

func TestCalendarDayPublicYiJi(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/calendar/day", GetCalendarDay)
	got := personalRequest(r, http.MethodGet, "/api/calendar/day?date=2026-11-02", "")
	if got.Code != http.StatusOK || !bytes.Contains(got.Body.Bytes(), []byte(`"lunar_full":"九月廿四"`)) || !bytes.Contains(got.Body.Bytes(), []byte(`"yi":["祭祀","沐浴","解除","破屋","坏垣","馀事勿取"]`)) || !bytes.Contains(got.Body.Bytes(), []byte(`"ji":["开市","嫁娶"]`)) {
		t.Fatalf("public full calendar labels: %d %s", got.Code, got.Body.String())
	}
	if invalid := personalRequest(r, http.MethodGet, "/api/calendar/day?date=2026-02-30", ""); invalid.Code != http.StatusBadRequest {
		t.Fatalf("invalid date: %d %s", invalid.Code, invalid.Body.String())
	}
}

func TestCalendarRangeIsCompleteAndBounded(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/calendar/range", GetCalendarRange)
	got := personalRequest(r, http.MethodGet, "/api/calendar/range?start=2028-01-01&end=2028-01-03", "")
	if got.Code != http.StatusOK {
		t.Fatalf("range: %d %s", got.Code, got.Body.String())
	}
	var body struct {
		Data struct {
			Days []struct {
				Date      string   `json:"date"`
				Yi        []string `json:"yi"`
				Ji        []string `json:"ji"`
				DayGanZhi string   `json:"day_gan_zhi"`
			} `json:"days"`
		} `json:"data"`
	}
	if err := json.Unmarshal(got.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Data.Days) != 3 || body.Data.Days[0].Date != "2028-01-01" || body.Data.Days[2].Date != "2028-01-03" {
		t.Fatalf("incomplete range: %+v", body.Data.Days)
	}
	for _, day := range body.Data.Days {
		if day.Yi == nil || day.Ji == nil || day.DayGanZhi == "" {
			t.Fatalf("missing date fields: %+v", day)
		}
	}
	for _, path := range []string{
		"?start=2026-02-30&end=2026-03-01", "?start=2026-11-03&end=2026-11-02",
		"?start=2026-01-01&end=2026-04-01", "?start=1899-12-31&end=1900-01-01",
		"?start=2100-12-31&end=2101-01-01",
	} {
		result := personalRequest(r, http.MethodGet, "/api/calendar/range"+path, "")
		if result.Code != http.StatusBadRequest || bytes.Contains(result.Body.Bytes(), []byte(`"days"`)) {
			t.Fatalf("invalid range %s: %d %s", path, result.Code, result.Body.String())
		}
	}
	boundary := personalRequest(r, http.MethodGet, "/api/calendar/range?start=2026-11-02&end=2027-01-30", "")
	if boundary.Code != http.StatusOK {
		t.Fatalf("90-day boundary: %d %s", boundary.Code, boundary.Body.String())
	}
}

func TestPersonalEndpointsRequireOwnedChartAndSession(t *testing.T) {
	db := withArticleHandlerDB(t)
	var owner, other, chartID string
	for _, id := range []*string{&owner, &other} {
		if err := db.QueryRow(`INSERT INTO users (email,password_hash) VALUES (gen_random_uuid()::text || '@example.com','hash') RETURNING id`).Scan(id); err != nil {
			t.Fatal(err)
		}
	}
	// A complete chart is produced through the same authenticated calculation
	// handler used by the miniapp; creation alone must not set personal binding.
	calc := gin.New()
	calc.POST("/calc", func(c *gin.Context) { c.Set("user_id", owner); Calculate(c) })
	response := personalRequest(calc, http.MethodPost, "/calc", `{"year":1990,"month":1,"day":1,"hour":1,"gender":"male"}`)
	if response.Code != http.StatusOK {
		t.Fatalf("calculate: %d %s", response.Code, response.Body.String())
	}
	var calculation struct {
		ChartID string `json:"chart_id"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &calculation); err != nil {
		t.Fatal(err)
	}
	chartID = calculation.ChartID
	if chartID == "" {
		t.Fatal("chart not saved")
	}

	ownerAPI, otherAPI := personalHandlerRouter(owner), personalHandlerRouter(other)
	if got := personalRequest(ownerAPI, http.MethodGet, "/api/user/personal-chart", ""); got.Code != http.StatusOK || !bytes.Contains(got.Body.Bytes(), []byte(`"chart":null`)) {
		t.Fatalf("new chart must not auto-bind: %d %s", got.Code, got.Body.String())
	}
	if got := personalRequest(otherAPI, http.MethodPut, "/api/user/personal-chart", `{"chart_id":"`+chartID+`"}`); got.Code != http.StatusNotFound {
		t.Fatalf("cross-user bind: %d %s", got.Code, got.Body.String())
	}
	if got := personalRequest(ownerAPI, http.MethodPut, "/api/user/personal-chart", `{"chart_id":"`+chartID+`"}`); got.Code != http.StatusOK {
		t.Fatalf("owner bind: %d %s", got.Code, got.Body.String())
	}
	if got := personalRequest(otherAPI, http.MethodGet, "/api/user/personal-day?date=2026-11-01", ""); got.Code != http.StatusConflict {
		t.Fatalf("other user should remain unbound: %d %s", got.Code, got.Body.String())
	}
	if got := personalRequest(ownerAPI, http.MethodGet, "/api/user/personal-day?date=2026-11-01", ""); got.Code != http.StatusOK || !bytes.Contains(got.Body.Bytes(), []byte(chartID)) {
		t.Fatalf("owner daily result: %d %s", got.Code, got.Body.String())
	}
	if got := personalRequest(ownerAPI, http.MethodGet, "/api/user/personal-day?date=2026-02-30", ""); got.Code != http.StatusBadRequest {
		t.Fatalf("invalid date: %d %s", got.Code, got.Body.String())
	}
	protected := gin.New()
	protected.GET("/api/user/personal-day", middleware.Auth(), GetPersonalDay)
	if got := personalRequest(protected, http.MethodGet, "/api/user/personal-day?date=2026-11-01", ""); got.Code != http.StatusUnauthorized {
		t.Fatalf("missing JWT: %d %s", got.Code, got.Body.String())
	}
}
