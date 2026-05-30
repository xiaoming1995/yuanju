package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"yuanju/configs"
	"yuanju/internal/middleware"

	"github.com/gin-gonic/gin"
)

func TestAdminListCompatReadingsRequiresAdminAuth(t *testing.T) {
	configs.AppConfig.AdminJWTSecret = "test-admin-secret"
	r := gin.New()
	admin := r.Group("/api/admin", middleware.AdminAuth())
	admin.GET("/compatibility/readings", AdminListCompatReadings)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/compatibility/readings", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for unauthenticated compat list, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAdminGetCompatReadingDetailRequiresAdminAuth(t *testing.T) {
	configs.AppConfig.AdminJWTSecret = "test-admin-secret"
	r := gin.New()
	admin := r.Group("/api/admin", middleware.AdminAuth())
	admin.GET("/compatibility/readings/:id", AdminGetCompatReadingDetail)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/compatibility/readings/abc", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for unauthenticated compat detail, got %d: %s", w.Code, w.Body.String())
	}
}
