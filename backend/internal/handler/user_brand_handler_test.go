package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestValidateBrandUpdate_Valid(t *testing.T) {
	if err := validateBrandUpdate(BrandUpdateReq{
		Title:         "清雨堂",
		FooterText:    "清雨堂 · 命理咨询",
		WatermarkMode: "diagonal",
		WatermarkText: "仅供参考",
	}); err != nil {
		t.Fatalf("expected valid, got %v", err)
	}
}

func TestValidateBrandUpdate_TitleTooLong(t *testing.T) {
	twentyOne := "一二三四五六七八九十一二三四五六七八九十一"
	if err := validateBrandUpdate(BrandUpdateReq{Title: twentyOne, WatermarkMode: "none"}); err == nil {
		t.Fatal("expected error for 21-rune title")
	}
}

func TestValidateBrandUpdate_FooterTooLong(t *testing.T) {
	fortyOne := ""
	for i := 0; i < 41; i++ {
		fortyOne += "字"
	}
	if err := validateBrandUpdate(BrandUpdateReq{FooterText: fortyOne, WatermarkMode: "none"}); err == nil {
		t.Fatal("expected error for 41-rune footer")
	}
}

func TestValidateBrandUpdate_InvalidMode(t *testing.T) {
	if err := validateBrandUpdate(BrandUpdateReq{WatermarkMode: "evil"}); err == nil {
		t.Fatal("expected error for invalid mode")
	}
}

func TestValidateBrandUpdate_UnsafeChars(t *testing.T) {
	for _, ch := range []string{"<", ">", "\"", "'", "&"} {
		if err := validateBrandUpdate(BrandUpdateReq{Title: "x" + ch + "y", WatermarkMode: "none"}); err == nil {
			t.Fatalf("expected error for unsafe char %q in title", ch)
		}
	}
}

func TestDetectImageType_PNG(t *testing.T) {
	header := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0, 0, 0, 0}
	if got := detectImageType(header); got != "png" {
		t.Fatalf("expected png, got %q", got)
	}
}

func TestDetectImageType_JPEG(t *testing.T) {
	header := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0, 0, 0, 0, 0, 0, 0, 0}
	if got := detectImageType(header); got != "jpg" {
		t.Fatalf("expected jpg, got %q", got)
	}
}

func TestDetectImageType_WebP(t *testing.T) {
	header := []byte{'R', 'I', 'F', 'F', 0, 0, 0, 0, 'W', 'E', 'B', 'P'}
	if got := detectImageType(header); got != "webp" {
		t.Fatalf("expected webp, got %q", got)
	}
}

func TestDetectImageType_RejectText(t *testing.T) {
	header := []byte("plain text content xx")
	if got := detectImageType(header); got != "" {
		t.Fatalf("expected empty for non-image, got %q", got)
	}
}

func TestDetectImageType_RejectShortBuffer(t *testing.T) {
	header := []byte{0x89, 0x50, 0x4E}
	if got := detectImageType(header); got != "" {
		t.Fatalf("expected empty for short buffer, got %q", got)
	}
}

func TestBrandHandler_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/user/export-brand", RequireUserID, GetExportBrand)
	req := httptest.NewRequest(http.MethodGet, "/api/user/export-brand", bytes.NewReader(nil))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for missing user_id, got %d", w.Code)
	}
}
