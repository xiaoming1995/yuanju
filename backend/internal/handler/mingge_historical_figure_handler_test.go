package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"yuanju/internal/model"

	"github.com/gin-gonic/gin"
)

func TestGetMingGeHistoricalFiguresLimitsPublicResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	original := listPublishedMingGeHistoricalFigures
	t.Cleanup(func() { listPublishedMingGeHistoricalFigures = original })
	listPublishedMingGeHistoricalFigures = func(mingGe string) ([]model.MingGeHistoricalFigure, error) {
		if mingGe != "伤官格" {
			t.Fatalf("got ming_ge %q", mingGe)
		}
		return []model.MingGeHistoricalFigure{{FigureName: "甲"}, {FigureName: "乙"}, {FigureName: "丙"}}, nil
	}
	router := gin.New()
	router.GET("/references/:ming_ge", GetMingGeHistoricalFigures)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/references/%E4%BC%A4%E5%AE%98%E6%A0%BC", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if strings.Count(recorder.Body.String(), "figure_name") != 2 {
		t.Fatalf("public response should cap at two entries, got %s", recorder.Body.String())
	}
}

func TestAdminCreateMingGeHistoricalFigureExplainsIncompleteDayun(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/references", AdminCreateMingGeHistoricalFigure)
	body := `{"ming_ge":"伤官格","figure_name":"测试人物","era":"测试时代","identity":"测试身份","historical_memory":"测试历史印记","source_title":"测试资料","source_url":"https://example.com/source","show_dayun":true}`
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/references", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected incomplete Dayun request to be rejected, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "展示大运") {
		t.Fatalf("expected actionable Dayun validation message, got %s", recorder.Body.String())
	}
}
