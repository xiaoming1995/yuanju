package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"yuanju/configs"
	"yuanju/internal/middleware"

	"github.com/gin-gonic/gin"
)

func newUserProfileRouter() *gin.Engine {
	configs.AppConfig.JWTSecret = "test-secret"
	r := gin.New()
	r.GET("/api/user/profile", middleware.Auth(), GetUserProfile)
	return r
}

func TestUserProfileRequiresAuthentication(t *testing.T) {
	r := newUserProfileRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/user/profile", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized profile access to return 401, got %d: %s", w.Code, w.Body.String())
	}
}
