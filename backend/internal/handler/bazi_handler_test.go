package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"yuanju/internal/model"
	"yuanju/internal/service"
	"yuanju/pkg/bazi"
)

func TestNormalizeChartDisplayName(t *testing.T) {
	name, err := normalizeChartDisplayName("  小王  ")
	if err != nil {
		t.Fatal(err)
	}
	if name != "小王" {
		t.Fatalf("expected trimmed name, got %q", name)
	}
}

func TestNormalizeChartDisplayName_AllowsEmpty(t *testing.T) {
	name, err := normalizeChartDisplayName("   ")
	if err != nil {
		t.Fatal(err)
	}
	if name != "" {
		t.Fatalf("expected empty name, got %q", name)
	}
}

func TestNormalizeChartDisplayName_RejectsLongName(t *testing.T) {
	_, err := normalizeChartDisplayName(strings.Repeat("命", 21))
	if err == nil {
		t.Fatal("expected long name to be rejected")
	}
	if !strings.Contains(err.Error(), "20") {
		t.Fatalf("expected error to mention length limit, got %v", err)
	}
}

func TestUpdateChartDisplayNameRequest_RejectsMissingDisplayName(t *testing.T) {
	recorder := performUpdateHistoryDisplayNameRequest(t, `{}`)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "display_name") {
		t.Fatalf("expected error to mention display_name, got %s", recorder.Body.String())
	}
}

func TestUpdateChartDisplayNameRequest_RejectsNullDisplayName(t *testing.T) {
	recorder := performUpdateHistoryDisplayNameRequest(t, `{"display_name":null}`)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "display_name") {
		t.Fatalf("expected error to mention display_name, got %s", recorder.Body.String())
	}
}

func TestUpdateChartDisplayNameRequest_RejectsMalformedChartID(t *testing.T) {
	recorder := performUpdateHistoryDisplayNameRequestAtPath(t, "/history/not-a-uuid/display-name", `{"display_name":"小王"}`)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "无效的命盘ID") {
		t.Fatalf("expected invalid chart id error, got %s", recorder.Body.String())
	}
}

func TestDeleteHistory_RejectsMalformedChartID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set("user_id", "user-1") })
	router.DELETE("/history/:id", DeleteHistory)

	req := httptest.NewRequest(http.MethodDelete, "/history/not-a-uuid", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "无效的命盘ID") {
		t.Fatalf("expected invalid chart id error, got %s", recorder.Body.String())
	}
}

func TestHandlePastEventsExport_ReturnsReadOnlyExportDataForOwnedChart(t *testing.T) {
	gin.SetMode(gin.TestMode)
	owner := "user-1"
	origGetChart := getBaziChartByIDForPastEventsExport
	origGenerate := generatePastEventsExportForChartHandler
	getBaziChartByIDForPastEventsExport = func(chartID string) (*model.BaziChart, error) {
		if chartID != "chart-1" {
			t.Fatalf("unexpected chart id %q", chartID)
		}
		return &model.BaziChart{ID: chartID, UserID: &owner}, nil
	}
	generateCalled := false
	generatePastEventsExportForChartHandler = func(chart *model.BaziChart) (*service.PastEventsExportResponse, error) {
		generateCalled = true
		return &service.PastEventsExportResponse{
			Chart: service.PastEventsExportChart{ID: chart.ID},
			Segments: []service.PastEventsExportSegment{
				{DayunIndex: 1, GanZhi: "甲子", Summary: "已生成", Years: []service.PastEventsExportYear{{Year: 2000, GanZhi: "庚辰", Narrative: "已生成年份批语"}}},
			},
			Generated: "cached-dayun-summaries",
		}, nil
	}
	t.Cleanup(func() {
		getBaziChartByIDForPastEventsExport = origGetChart
		generatePastEventsExportForChartHandler = origGenerate
	})

	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set("user_id", owner) })
	router.GET("/past-events/export/:chart_id", HandlePastEventsExport)
	req := httptest.NewRequest(http.MethodGet, "/past-events/export/chart-1", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if !generateCalled {
		t.Fatalf("expected read-only export service to be called")
	}
	if !strings.Contains(recorder.Body.String(), "cached-dayun-summaries") {
		t.Fatalf("expected cached export response, got %s", recorder.Body.String())
	}
}

func TestHandlePastEventsExport_RejectsForeignChartWithoutGenerating(t *testing.T) {
	gin.SetMode(gin.TestMode)
	owner := "user-1"
	other := "user-2"
	origGetChart := getBaziChartByIDForPastEventsExport
	origGenerate := generatePastEventsExportForChartHandler
	getBaziChartByIDForPastEventsExport = func(chartID string) (*model.BaziChart, error) {
		return &model.BaziChart{ID: chartID, UserID: &other}, nil
	}
	generatePastEventsExportForChartHandler = func(chart *model.BaziChart) (*service.PastEventsExportResponse, error) {
		t.Fatalf("must not generate or aggregate export data for foreign chart")
		return nil, nil
	}
	t.Cleanup(func() {
		getBaziChartByIDForPastEventsExport = origGetChart
		generatePastEventsExportForChartHandler = origGenerate
	})

	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set("user_id", owner) })
	router.GET("/past-events/export/:chart_id", HandlePastEventsExport)
	req := httptest.NewRequest(http.MethodGet, "/past-events/export/chart-1", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestGetLiunianReport_ReturnsCachedReportForOwnedChart(t *testing.T) {
	gin.SetMode(gin.TestMode)
	owner := "user-1"
	origGetChart := getBaziChartByIDForLiunianReport
	origGetReport := getLiunianReportForHandler
	raw := json.RawMessage(`{"career":"已缓存"}`)
	getBaziChartByIDForLiunianReport = func(chartID string) (*model.BaziChart, error) {
		if chartID != "chart-1" {
			t.Fatalf("unexpected chart id %q", chartID)
		}
		return &model.BaziChart{ID: chartID, UserID: &owner}, nil
	}
	getLiunianReportForHandler = func(chartID string, targetYear int) (*model.AILiunianReport, error) {
		if chartID != "chart-1" || targetYear != 2026 {
			t.Fatalf("unexpected cache key chart=%s year=%d", chartID, targetYear)
		}
		return &model.AILiunianReport{
			ID:                "liunian-1",
			ChartID:           chartID,
			TargetYear:        targetYear,
			ContentStructured: &raw,
		}, nil
	}
	t.Cleanup(func() {
		getBaziChartByIDForLiunianReport = origGetChart
		getLiunianReportForHandler = origGetReport
	})

	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set("user_id", owner) })
	router.GET("/liunian-report/:chart_id", GetLiunianReport)
	req := httptest.NewRequest(http.MethodGet, "/liunian-report/chart-1?target_year=2026", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"cached":true`) {
		t.Fatalf("expected cached response, got %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"career":"已缓存"`) {
		t.Fatalf("expected cached content, got %s", recorder.Body.String())
	}
}

func TestGenerateLiunianReport_PassesForceRegenerateFlag(t *testing.T) {
	gin.SetMode(gin.TestMode)
	owner := "user-1"
	origGetChart := getBaziChartByIDForLiunianReport
	origGenerate := generateLiunianReportForHandler
	raw := json.RawMessage(`{"career":"重新生成"}`)
	getBaziChartByIDForLiunianReport = func(chartID string) (*model.BaziChart, error) {
		return &model.BaziChart{ID: chartID, UserID: &owner}, nil
	}
	generateLiunianReportForHandler = func(chartID string, targetYear int, userID *string, forceRegenerate bool) (*model.AILiunianReport, bool, error) {
		if chartID != "chart-1" || targetYear != 2026 {
			t.Fatalf("unexpected request chart=%s year=%d", chartID, targetYear)
		}
		if userID == nil || *userID != owner {
			t.Fatalf("unexpected user id: %v", userID)
		}
		if !forceRegenerate {
			t.Fatal("expected forceRegenerate=true")
		}
		return &model.AILiunianReport{
			ID:                "liunian-2",
			ChartID:           chartID,
			TargetYear:        targetYear,
			ContentStructured: &raw,
		}, false, nil
	}
	t.Cleanup(func() {
		getBaziChartByIDForLiunianReport = origGetChart
		generateLiunianReportForHandler = origGenerate
	})

	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set("user_id", owner) })
	router.POST("/liunian-report/:chart_id", GenerateLiunianReport)
	req := httptest.NewRequest(http.MethodPost, "/liunian-report/chart-1", bytes.NewBufferString(`{"target_year":2026,"force_regenerate":true}`))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"cached":false`) {
		t.Fatalf("expected generated response, got %s", recorder.Body.String())
	}
}

func performUpdateHistoryDisplayNameRequest(t *testing.T, body string) *httptest.ResponseRecorder {
	return performUpdateHistoryDisplayNameRequestAtPath(t, "/history/chart-1/display-name", body)
}

func performUpdateHistoryDisplayNameRequestAtPath(t *testing.T, path, body string) *httptest.ResponseRecorder {
	t.Helper()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PATCH("/history/:id/display-name", UpdateHistoryDisplayName)

	req := httptest.NewRequest(http.MethodPatch, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)
	return recorder
}

func TestCalculateInput_BindsDisplayName(t *testing.T) {
	var input CalculateInput
	body := `{"year":2001,"month":1,"day":1,"hour":12,"gender":"male","display_name":"小王"}`
	if err := json.Unmarshal([]byte(body), &input); err != nil {
		t.Fatal(err)
	}
	if input.DisplayName != "小王" {
		t.Fatalf("expected DisplayName=小王, got %q", input.DisplayName)
	}
}

func TestCalculate_RejectsLongDisplayName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/calculate", Calculate)

	longName := strings.Repeat("命", 21)
	body := fmt.Sprintf(
		`{"year":2001,"month":1,"day":1,"hour":12,"gender":"male","display_name":%q}`,
		longName,
	)

	req := httptest.NewRequest(http.MethodPost, "/calculate", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "20") {
		t.Fatalf("expected error to mention length limit, got %s", recorder.Body.String())
	}
}

func TestResolvePillars_ReturnsCandidates(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/resolve-pillars", ResolvePillars)

	body := `{"year_pillar":"甲子","month_pillar":"丙寅","day_pillar":"丁丑","hour_pillar":"丙午"}`
	req := httptest.NewRequest(http.MethodPost, "/resolve-pillars", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp struct {
		Candidates []bazi.Candidate `json:"candidates"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp.Candidates) == 0 {
		t.Fatalf("expected at least one candidate")
	}
}

func TestResolvePillars_EmptyOnNoMatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/resolve-pillars", ResolvePillars)

	body := `{"year_pillar":"甲子","month_pillar":"甲子","day_pillar":"甲子","hour_pillar":"甲子"}`
	req := httptest.NewRequest(http.MethodPost, "/resolve-pillars", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp struct {
		Candidates []bazi.Candidate `json:"candidates"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp.Candidates) != 0 {
		t.Errorf("expected empty candidates, got %v", resp.Candidates)
	}
}

func TestResolvePillars_RejectsMissingPillar(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/resolve-pillars", ResolvePillars)

	// 缺少 hour_pillar，binding:"required" 应返回 422
	body := `{"year_pillar":"甲子","month_pillar":"丙寅","day_pillar":"丁丑"}`
	req := httptest.NewRequest(http.MethodPost, "/resolve-pillars", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 for missing pillar, got %d", w.Code)
	}
}
