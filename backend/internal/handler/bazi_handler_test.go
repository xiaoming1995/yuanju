package handler

import (
	"bytes"
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

func performUpdateHistoryDisplayNameRequest(t *testing.T, body string) *httptest.ResponseRecorder {
	t.Helper()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PATCH("/history/:id/display-name", UpdateHistoryDisplayName)

	req := httptest.NewRequest(http.MethodPatch, "/history/chart-1/display-name", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)
	return recorder
}
