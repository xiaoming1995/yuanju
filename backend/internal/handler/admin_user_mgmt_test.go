package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"yuanju/configs"
	"yuanju/internal/middleware"

	"github.com/gin-gonic/gin"
)

func TestAdminResetUserPasswordRequiresAdminAuth(t *testing.T) {
	configs.AppConfig.AdminJWTSecret = "test-admin-secret"
	r := gin.New()
	admin := r.Group("/api/admin", middleware.AdminAuth())
	admin.POST("/users/:id/reset-password", AdminResetUserPassword)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/users/abc/reset-password", strings.NewReader(`{"password":"password123"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAdminSetUserDisabledRequiresAdminAuth(t *testing.T) {
	configs.AppConfig.AdminJWTSecret = "test-admin-secret"
	r := gin.New()
	admin := r.Group("/api/admin", middleware.AdminAuth())
	admin.PUT("/users/:id/disable", AdminSetUserDisabled)

	req := httptest.NewRequest(http.MethodPut, "/api/admin/users/abc/disable", strings.NewReader(`{"disabled":true}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAdminDeleteUserRequiresAdminAuth(t *testing.T) {
	configs.AppConfig.AdminJWTSecret = "test-admin-secret"
	r := gin.New()
	admin := r.Group("/api/admin", middleware.AdminAuth())
	admin.DELETE("/users/:id", AdminDeleteUser)

	req := httptest.NewRequest(http.MethodDelete, "/api/admin/users/abc", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAdminUserActivityEndpointsRequireAdminAuth(t *testing.T) {
	configs.AppConfig.AdminJWTSecret = "test-admin-secret"
	r := gin.New()
	admin := r.Group("/api/admin", middleware.AdminAuth())
	admin.GET("/users/:id", AdminGetUserDetail)
	admin.GET("/users/:id/events", AdminGetUserEvents)
	for _, path := range []string{"/api/admin/users/00000000-0000-0000-0000-000000000001", "/api/admin/users/00000000-0000-0000-0000-000000000001/events"} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("%s: expected 401, got %d", path, w.Code)
		}
	}
}
