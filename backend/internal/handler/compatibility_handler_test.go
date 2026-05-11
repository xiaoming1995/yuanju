package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newCompatibilityRouter() *gin.Engine {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", "test-user-id")
		c.Next()
	})
	r.POST("/compatibility", CreateCompatibilityReading)
	r.GET("/compatibility/history", GetCompatibilityHistory)
	r.GET("/compatibility/:id", GetCompatibilityDetail)
	r.POST("/compatibility/:id/report", GenerateCompatibilityReport)
	return r
}

func TestCreateCompatibilityReading_MissingFields(t *testing.T) {
	r := newCompatibilityRouter()

	cases := []struct {
		name    string
		body    map[string]any
		wantCode int
	}{
		{
			name:    "empty body",
			body:    map[string]any{},
			wantCode: http.StatusUnprocessableEntity,
		},
		{
			name: "missing partner",
			body: map[string]any{
				"self": map[string]any{
					"year": 1990, "month": 1, "day": 1, "hour": 0, "gender": "male",
				},
			},
			wantCode: http.StatusUnprocessableEntity,
		},
		{
			name: "invalid gender",
			body: map[string]any{
				"self": map[string]any{
					"year": 1990, "month": 1, "day": 1, "hour": 0, "gender": "unknown",
				},
				"partner": map[string]any{
					"year": 1992, "month": 6, "day": 15, "hour": 12, "gender": "female",
				},
			},
			wantCode: http.StatusUnprocessableEntity,
		},
		{
			name: "year out of range",
			body: map[string]any{
				"self": map[string]any{
					"year": 1800, "month": 1, "day": 1, "hour": 0, "gender": "male",
				},
				"partner": map[string]any{
					"year": 1992, "month": 6, "day": 15, "hour": 12, "gender": "female",
				},
			},
			wantCode: http.StatusUnprocessableEntity,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.body)
			req := httptest.NewRequest(http.MethodPost, "/compatibility", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tc.wantCode {
				t.Errorf("expected %d, got %d: %s", tc.wantCode, w.Code, w.Body.String())
			}
		})
	}
}
