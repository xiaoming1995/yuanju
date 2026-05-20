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

func TestAdminCreateUserRequiresAdminAuth(t *testing.T) {
	configs.AppConfig.AdminJWTSecret = "test-admin-secret"
	r := gin.New()
	admin := r.Group("/api/admin", middleware.AdminAuth())
	admin.POST("/users", AdminCreateUser)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/users", strings.NewReader(`{
		"email":"new@example.com",
		"password":"password123"
	}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for unauthenticated admin user creation, got %d: %s", w.Code, w.Body.String())
	}
}
