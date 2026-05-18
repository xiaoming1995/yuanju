package handler

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strings"
	"testing"
	"time"

	"yuanju/pkg/ratelimit"

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

// TestBrandLogo_FakeMime exercises the upload handler end-to-end through
// multipart parsing: declared image/png with non-image body should 400 at
// the magic-bytes layer, before any DB call.
func TestBrandLogo_FakeMime(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", "00000000-0000-0000-0000-000000000001")
		c.Next()
	})
	r.POST("/upload", RequireUserID, UploadExportBrandLogo)

	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)
	h := textproto.MIMEHeader{}
	h.Set("Content-Disposition", `form-data; name="file"; filename="logo.png"`)
	h.Set("Content-Type", "image/png")
	part, _ := mw.CreatePart(h)
	part.Write([]byte("this is not a real png, just text content that fills 12+ bytes"))
	mw.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for fake mime, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "图片") && !strings.Contains(w.Body.String(), "类型") {
		t.Fatalf("expected error body to mention image/type, got %q", w.Body.String())
	}
}

// TestBrandLogo_RateLimit exercises the per-user rate limiter on the upload
// endpoint. Eleven rapid uploads from the same user must yield exactly one 429
// among them. Uses a *fresh* limiter instance via package var override so
// state from prior tests in the same binary doesn't pollute this test.
func TestBrandLogo_RateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Override the package-global limiter for this test with a fresh instance.
	// (Save and restore to avoid leaking state to other tests.)
	saved := logoLimiter
	logoLimiter = ratelimit.New(10, time.Minute)
	defer func() { logoLimiter = saved }()

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", "00000000-0000-0000-0000-000000000002")
		c.Next()
	})
	r.POST("/upload", RequireUserID, UploadExportBrandLogo)

	rateLimited := 0
	for i := 0; i < 11; i++ {
		// Empty body — but the limiter check runs BEFORE FormFile parsing.
		req := httptest.NewRequest(http.MethodPost, "/upload", bytes.NewReader(nil))
		req.Header.Set("Content-Type", "multipart/form-data; boundary=x")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code == http.StatusTooManyRequests {
			rateLimited++
		}
	}
	if rateLimited != 1 {
		t.Fatalf("expected exactly 1 of 11 calls to be 429, got %d", rateLimited)
	}
}

func TestValidateBrandUpdate_LogoMode(t *testing.T) {
	cases := []struct {
		name    string
		mode    string
		wantErr bool
	}{
		{"empty allowed (defaults to icon)", "", false},
		{"icon valid", "icon", false},
		{"wordmark valid", "wordmark", false},
		{"invalid square rejected", "square", true},
		{"invalid mixed-case rejected", "Icon", true},
		{"invalid arbitrary rejected", "xyz", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateBrandUpdate(BrandUpdateReq{
				WatermarkMode: "none",
				LogoMode:      tc.mode,
			})
			if (err != nil) != tc.wantErr {
				t.Fatalf("logo_mode=%q wantErr=%v gotErr=%v", tc.mode, tc.wantErr, err)
			}
		})
	}
}
