package handler

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"yuanju/configs"
	"yuanju/internal/middleware"
	"yuanju/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func officialAccountTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/miniapp/official-account-guide", GetPublicOfficialAccountGuide)
	r.Static("/static/uploads/official-account-qr", filepath.Join(configs.AppConfig.UploadDir, "official-account-qr"))
	admin := r.Group("/api/admin", middleware.AdminAuth())
	admin.GET("/settings/official-account-guide", AdminGetOfficialAccountGuide)
	admin.PUT("/settings/official-account-guide", AdminSaveOfficialAccountGuide)
	admin.POST("/settings/official-account-guide/qr", AdminUploadOfficialAccountQR)
	return r
}

func officialAccountToken(t *testing.T) string {
	t.Helper()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": "yuanju-admin", "sub": "admin-test", "exp": time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte(configs.AppConfig.AdminJWTSecret))
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func officialAccountRequest(r *gin.Engine, method, path, body, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func officialAccountUpload(r *gin.Engine, t *testing.T, imageBytes []byte, contentType, token string) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	part, err := mw.CreateFormFile("file", "guide.png")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(imageBytes); err != nil {
		t.Fatal(err)
	}
	if err := mw.Close(); err != nil {
		t.Fatal(err)
	}
	// multipart.CreateFormFile defaults to application/octet-stream; set the part type explicitly.
	raw := bytes.Replace(body.Bytes(), []byte("application/octet-stream"), []byte(contentType), 1)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/settings/official-account-guide/qr", bytes.NewReader(raw))
	req.Header.Set("Content-Type", mw.FormDataContentType())
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestOfficialAccountGuideAdminToMiniapp(t *testing.T) {
	withArticleHandlerDB(t)
	previousDir := configs.AppConfig.UploadDir
	configs.AppConfig.UploadDir = t.TempDir()
	t.Cleanup(func() { configs.AppConfig.UploadDir = previousDir })
	if err := os.MkdirAll(filepath.Join(configs.AppConfig.UploadDir, "official-account-qr"), 0755); err != nil {
		t.Fatal(err)
	}
	r := officialAccountTestRouter()
	token := officialAccountToken(t)

	if got := officialAccountRequest(r, "GET", "/api/miniapp/official-account-guide", "", ""); got.Code != 200 || !strings.Contains(got.Body.String(), `"enabled":false`) {
		t.Fatalf("default: %d %s", got.Code, got.Body.String())
	}
	if got := officialAccountRequest(r, "PUT", "/api/admin/settings/official-account-guide", `{}`, ""); got.Code != 401 {
		t.Fatalf("unauthorized save: %d", got.Code)
	}
	if got := officialAccountUpload(r, t, []byte("not an image"), "image/png", token); got.Code != 400 {
		t.Fatalf("fake image: %d %s", got.Code, got.Body.String())
	}
	if got := officialAccountUpload(r, t, []byte("not an image"), "image/png", ""); got.Code != 401 {
		t.Fatalf("unauthorized upload: %d", got.Code)
	}
	if got := officialAccountUpload(r, t, bytes.Repeat([]byte("x"), 2*1024*1024+1), "image/png", token); got.Code != 413 {
		t.Fatalf("oversize: %d %s", got.Code, got.Body.String())
	}
	if got := officialAccountRequest(r, "PUT", "/api/admin/settings/official-account-guide", `{"enabled":true,"name":"缘聚"}`, token); got.Code != 400 {
		t.Fatalf("incomplete: %d %s", got.Code, got.Body.String())
	}

	img := image.NewRGBA(image.Rect(0, 0, 320, 320))
	img.Set(0, 0, color.Black)
	var pngBody bytes.Buffer
	if err := png.Encode(&pngBody, img); err != nil {
		t.Fatal(err)
	}
	upload := officialAccountUpload(r, t, pngBody.Bytes(), "image/png", token)
	if upload.Code != 200 {
		t.Fatalf("upload: %d %s", upload.Code, upload.Body.String())
	}
	var uploaded struct {
		Data struct {
			QRPath string `json:"qr_path"`
		} `json:"data"`
	}
	if err := json.Unmarshal(upload.Body.Bytes(), &uploaded); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(uploaded.Data.QRPath, repository.OfficialAccountQRPrefix) {
		t.Fatalf("path: %q", uploaded.Data.QRPath)
	}
	if got := officialAccountRequest(r, "PUT", "/api/admin/settings/official-account-guide", `{"enabled":true,"name":"缘聚","wechat_id":"yuanju","description":"内容更新","qr_path":"https://bad.example/qr.png"}`, token); got.Code != 400 {
		t.Fatalf("external URL: %d", got.Code)
	}
	data, _ := json.Marshal(repository.OfficialAccountGuide{Enabled: true, Name: "缘聚", WechatID: "yuanju", Description: "内容更新", QRPath: uploaded.Data.QRPath})
	if got := officialAccountRequest(r, "PUT", "/api/admin/settings/official-account-guide", string(data), token); got.Code != 200 {
		t.Fatalf("save: %d %s", got.Code, got.Body.String())
	}
	if got := officialAccountRequest(r, "GET", "/api/miniapp/official-account-guide", "", ""); got.Code != 200 || !strings.Contains(got.Body.String(), `"wechat_id":"yuanju"`) {
		t.Fatalf("public: %d %s", got.Code, got.Body.String())
	}
	if got := officialAccountRequest(r, "GET", uploaded.Data.QRPath, "", ""); got.Code != 200 {
		t.Fatalf("image: %d", got.Code)
	}

	disabled, _ := json.Marshal(repository.OfficialAccountGuide{Name: "缘聚", WechatID: "yuanju", Description: "内容更新", QRPath: uploaded.Data.QRPath})
	if got := officialAccountRequest(r, "PUT", "/api/admin/settings/official-account-guide", string(disabled), token); got.Code != 200 {
		t.Fatalf("disable: %d %s", got.Code, got.Body.String())
	}
	if got := officialAccountRequest(r, "GET", "/api/miniapp/official-account-guide", "", ""); got.Code != 200 || strings.Contains(got.Body.String(), `"wechat_id"`) {
		t.Fatalf("disabled public: %d %s", got.Code, got.Body.String())
	}
	if got := officialAccountRequest(r, "PUT", "/api/admin/settings/official-account-guide", string(data), token); got.Code != 200 {
		t.Fatalf("re-enable: %d %s", got.Code, got.Body.String())
	}
	replacement := officialAccountUpload(r, t, pngBody.Bytes(), "image/png", token)
	if replacement.Code != 200 {
		t.Fatalf("replacement upload: %d %s", replacement.Code, replacement.Body.String())
	}
	var replacementData struct {
		Data struct {
			QRPath string `json:"qr_path"`
		} `json:"data"`
	}
	if err := json.Unmarshal(replacement.Body.Bytes(), &replacementData); err != nil {
		t.Fatal(err)
	}
	if got := officialAccountRequest(r, "PUT", "/api/admin/settings/official-account-guide", `{"enabled":true,"name":"","wechat_id":"yuanju"}`, token); got.Code != 400 {
		t.Fatalf("invalid replacement: %d", got.Code)
	}
	if got := officialAccountRequest(r, "GET", uploaded.Data.QRPath, "", ""); got.Code != 200 {
		t.Fatalf("old image lost before save: %d", got.Code)
	}
	updated, _ := json.Marshal(repository.OfficialAccountGuide{Enabled: true, Name: "缘聚", WechatID: "yuanju", Description: "内容更新", QRPath: replacementData.Data.QRPath})
	if got := officialAccountRequest(r, "PUT", "/api/admin/settings/official-account-guide", string(updated), token); got.Code != 200 {
		t.Fatalf("replace: %d %s", got.Code, got.Body.String())
	}
	if got := officialAccountRequest(r, "GET", uploaded.Data.QRPath, "", ""); got.Code != 404 {
		t.Fatalf("replaced image still public: %d", got.Code)
	}
	if got := officialAccountRequest(r, "GET", replacementData.Data.QRPath, "", ""); got.Code != 200 {
		t.Fatalf("new image unavailable: %d", got.Code)
	}
	if err := os.Remove(officialAccountQRFile(replacementData.Data.QRPath)); err != nil {
		t.Fatal(err)
	}
	if got := officialAccountRequest(r, "GET", "/api/miniapp/official-account-guide", "", ""); got.Code != 200 || strings.Contains(got.Body.String(), `"wechat_id"`) {
		t.Fatalf("missing image exposed: %d %s", got.Code, got.Body.String())
	}
}

func TestOfficialAccountQRPrunesOnlyOldUnusedFiles(t *testing.T) {
	previousDir := configs.AppConfig.UploadDir
	configs.AppConfig.UploadDir = t.TempDir()
	t.Cleanup(func() { configs.AppConfig.UploadDir = previousDir })
	dir := filepath.Join(configs.AppConfig.UploadDir, officialAccountQRDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	active := repository.OfficialAccountQRPrefix + "550e8400-e29b-41d4-a716-446655440000.png"
	oldUnused := "550e8400-e29b-41d4-a716-446655440001.png"
	newUnused := "550e8400-e29b-41d4-a716-446655440002.png"
	for _, name := range []string{filepath.Base(active), oldUnused, newUnused} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("image"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	old := time.Now().Add(-25 * time.Hour)
	for _, name := range []string{filepath.Base(active), oldUnused} {
		if err := os.Chtimes(filepath.Join(dir, name), old, old); err != nil {
			t.Fatal(err)
		}
	}
	pruneUnusedOfficialAccountQR(active)
	if _, err := os.Stat(filepath.Join(dir, filepath.Base(active))); err != nil {
		t.Fatalf("active image removed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, newUnused)); err != nil {
		t.Fatalf("recent draft removed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, oldUnused)); !os.IsNotExist(err) {
		t.Fatalf("old unused image kept: %v", err)
	}
}
