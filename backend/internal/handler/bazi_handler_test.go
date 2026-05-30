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
