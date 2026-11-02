package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"yuanju/internal/middleware"

	"github.com/gin-gonic/gin"
)

func TestWeChatBindingRequiresOrdinaryUserToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/auth/wechat/bind", middleware.Auth(), WeChatBind)
	request := httptest.NewRequest(http.MethodPost, "/api/auth/wechat/bind", strings.NewReader(`{"code":"one-time-code"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated bind status=%d, want 401", response.Code)
	}
}

func TestWeChatWebLinkRequiresOrdinaryUserToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/auth/wechat/link-web", middleware.Auth(), WeChatLinkWebAccount)
	request := httptest.NewRequest(http.MethodPost, "/api/auth/wechat/link-web", strings.NewReader(`{"email":"web@example.com","password":"password123"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated link status=%d, want 401", response.Code)
	}
}
