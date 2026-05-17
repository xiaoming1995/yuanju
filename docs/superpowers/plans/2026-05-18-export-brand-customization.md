# Export Brand Customization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let logged-in users upload their own logo, customize header title, footer text, and watermark; apply these to PNG (ShareCard) and PDF (PrintLayout) exports while preserving "缘聚命理" defaults for users who haven't configured anything.

**Architecture:** New `user_export_brand` table (1 row per user); 4 REST endpoints under `/api/user/export-brand` (auth-required); multipart logo upload with 3-layer validation (size / MIME / magic-bytes); in-process per-user token-bucket rate limit on logo upload; Gin Static for `/static/uploads/*` (public); frontend settings page; ShareCard + PrintLayout accept optional `brand` prop, fall back to "缘聚" defaults when absent.

**Tech Stack:** Go + Gin + lib/pq + `github.com/google/uuid` (already in go.mod via lunar-go), React 19 + TypeScript + Vite, `node --test` for static-regex frontend tests.

**Spec:** `docs/superpowers/specs/2026-05-18-export-brand-customization-design.md`

**Branch from:** `main` HEAD `eb6000d`. Cut `feat/export-brand-customization`.

**Repo layout:** repo root `/Users/liujiming/web/yuanju`. Frontend cwd `/Users/liujiming/web/yuanju/frontend`. Backend cwd `/Users/liujiming/web/yuanju/backend`. Spec path is from repo root.

---

## Test Strategy Note (Important)

Spec §10.1 lists 11 backend tests. Some of them (e.g. `TestBrandLogo_OverwriteDeletesOld`, `TestBrandPut_Valid`) require a live Postgres connection. The existing test suite in `backend/internal/handler/*_test.go` (see `compatibility_handler_test.go`) uses `httptest` with an **injected** `user_id` middleware and never touches the real DB. No DB test harness exists.

To stay consistent with that pattern AND achieve equivalent coverage, this plan extracts pure helpers (`validateBrandUpdate`, `detectImageType`, `Limiter.Allow`) and tests those — same 11-test count, all in-process, no infra needed. Repository and DB-roundtrip behavior is exercised end-to-end through manual acceptance in Task 15. This is a deliberate deviation from the spec's literal test list, justified by codebase convention.

---

## File Inventory

### Backend (Go)
- **Create**: `backend/pkg/ratelimit/inmem.go` (~50 lines)
- **Create**: `backend/pkg/ratelimit/inmem_test.go` (~40 lines)
- **Create**: `backend/internal/model/user_brand.go` (~25 lines)
- **Create**: `backend/internal/repository/user_brand_repository.go` (~100 lines)
- **Create**: `backend/internal/handler/user_brand_handler.go` (~250 lines)
- **Create**: `backend/internal/handler/user_brand_handler_test.go` (~150 lines)
- **Modify**: `backend/pkg/database/database.go` (add DDL after token_usage_logs)
- **Modify**: `backend/cmd/api/main.go` (register 4 routes + static + mkdir)
- **Modify**: `backend/configs/config.go` (add `UploadDir`)

### Frontend (React + TypeScript)
- **Create**: `frontend/src/pages/BrandSettingsPage.tsx` (~280 lines)
- **Create**: `frontend/src/pages/BrandSettingsPage.css` (~120 lines)
- **Create**: `frontend/src/components/BrandPreviewCard.tsx` (~90 lines)
- **Create**: `frontend/tests/brand-settings.test.mjs` (~50 lines)
- **Modify**: `frontend/src/lib/api.ts` (add types + brandAPI)
- **Modify**: `frontend/src/components/ShareCard.tsx` (add brand prop)
- **Modify**: `frontend/src/components/PrintLayout.tsx` (add brand prop)
- **Modify**: `frontend/src/pages/ResultPage.tsx` (fetch brand + pass props)
- **Modify**: `frontend/src/pages/ProfilePage.tsx` (add entry card)
- **Modify**: `frontend/src/App.tsx` (add `/settings/brand` route)
- **Modify**: `frontend/vite.config.ts` (proxy `/static`)

---

## Task 0: Cut Branch

**Files:** none

- [ ] **Step 1: Verify on main, clean tree**

```bash
git -C /Users/liujiming/web/yuanju status
git -C /Users/liujiming/web/yuanju log -1 --oneline
```

Expected: `On branch main` / `Your branch is up to date with 'origin/main'` / `eb6000d docs(specs): export brand customization...`

- [ ] **Step 2: Cut feature branch**

```bash
git -C /Users/liujiming/web/yuanju checkout -b feat/export-brand-customization
```

Expected: `Switched to a new branch 'feat/export-brand-customization'`

---

## Task 1: Rate Limiter — RED

**Files:**
- Create: `backend/pkg/ratelimit/inmem_test.go`

- [ ] **Step 1: Write the failing tests**

Create `backend/pkg/ratelimit/inmem_test.go`:

```go
package ratelimit

import (
	"testing"
	"time"
)

func TestLimiter_AllowsRateThenDenies(t *testing.T) {
	l := New(3, time.Minute)
	for i := 0; i < 3; i++ {
		if !l.Allow("user-1") {
			t.Fatalf("call %d should be allowed", i+1)
		}
	}
	if l.Allow("user-1") {
		t.Fatal("4th call should be denied")
	}
}

func TestLimiter_PerKeyIsolated(t *testing.T) {
	l := New(1, time.Minute)
	if !l.Allow("user-a") {
		t.Fatal("user-a first call should be allowed")
	}
	if !l.Allow("user-b") {
		t.Fatal("user-b first call should be allowed even when user-a is over limit")
	}
	if l.Allow("user-a") {
		t.Fatal("user-a second call should be denied")
	}
}

func TestLimiter_WindowResets(t *testing.T) {
	l := New(2, 50*time.Millisecond)
	l.Allow("user-1")
	l.Allow("user-1")
	if l.Allow("user-1") {
		t.Fatal("3rd call within window should be denied")
	}
	time.Sleep(70 * time.Millisecond)
	if !l.Allow("user-1") {
		t.Fatal("after window reset, call should be allowed")
	}
}
```

- [ ] **Step 2: Run tests — confirm RED**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./pkg/ratelimit/... -run TestLimiter
```

Expected: fails — `package ratelimit not found` or build errors (`New` and `Allow` undefined).

- [ ] **Step 3: Commit RED**

```bash
git -C /Users/liujiming/web/yuanju add backend/pkg/ratelimit/inmem_test.go
git -C /Users/liujiming/web/yuanju commit -m "test(ratelimit): per-user token bucket fails before impl (RED)"
```

---

## Task 2: Rate Limiter — GREEN

**Files:**
- Create: `backend/pkg/ratelimit/inmem.go`

- [ ] **Step 1: Write minimal implementation**

Create `backend/pkg/ratelimit/inmem.go`:

```go
// Package ratelimit provides an in-process per-key sliding-window rate limiter.
// For multi-instance deployments, swap in a Redis-backed implementation behind
// the same Allow(key) signature.
package ratelimit

import (
	"sync"
	"time"
)

type bucket struct {
	count int
	resetAt time.Time
}

type Limiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	rate    int
	window  time.Duration
}

func New(rate int, window time.Duration) *Limiter {
	return &Limiter{
		buckets: make(map[string]*bucket),
		rate:    rate,
		window:  window,
	}
}

func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	b, ok := l.buckets[key]
	if !ok || now.After(b.resetAt) {
		l.buckets[key] = &bucket{count: 1, resetAt: now.Add(l.window)}
		return true
	}
	if b.count >= l.rate {
		return false
	}
	b.count++
	return true
}
```

- [ ] **Step 2: Run tests — confirm GREEN**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./pkg/ratelimit/... -v
```

Expected: `--- PASS: TestLimiter_AllowsRateThenDenies`, `TestLimiter_PerKeyIsolated`, `TestLimiter_WindowResets`. 3 tests pass.

- [ ] **Step 3: Commit GREEN**

```bash
git -C /Users/liujiming/web/yuanju add backend/pkg/ratelimit/inmem.go
git -C /Users/liujiming/web/yuanju commit -m "feat(ratelimit): in-process per-key sliding-window limiter"
```

---

## Task 3: Validation + Magic Bytes Helpers — RED

**Files:**
- Create: `backend/internal/handler/user_brand_handler_test.go`

- [ ] **Step 1: Write failing tests**

Create `backend/internal/handler/user_brand_handler_test.go`:

```go
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
	r.GET("/api/user/export-brand", requireUserID, GetExportBrand)
	req := httptest.NewRequest(http.MethodGet, "/api/user/export-brand", bytes.NewReader(nil))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for missing user_id, got %d", w.Code)
	}
}
```

- [ ] **Step 2: Run tests — confirm RED**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./internal/handler/... -run "TestValidateBrandUpdate|TestDetectImageType|TestBrandHandler_Unauthenticated"
```

Expected: build error — `BrandUpdateReq`, `validateBrandUpdate`, `detectImageType`, `requireUserID`, `GetExportBrand` all undefined.

- [ ] **Step 3: Commit RED**

```bash
git -C /Users/liujiming/web/yuanju add backend/internal/handler/user_brand_handler_test.go
git -C /Users/liujiming/web/yuanju commit -m "test(brand): validation + magic-bytes + auth-smoke tests (RED)"
```

---

## Task 4: DDL + Config + Model + Repository (Infrastructure, No Test)

This task wires up the data layer. No unit test — the layer is exercised end-to-end in manual acceptance (Task 15).

**Files:**
- Modify: `backend/configs/config.go`
- Modify: `backend/pkg/database/database.go`
- Create: `backend/internal/model/user_brand.go`
- Create: `backend/internal/repository/user_brand_repository.go`

- [ ] **Step 1: Add `UploadDir` to config**

Edit `backend/configs/config.go`:

Add field to the `Config` struct (after `AIPromptLog`):
```go
	// Upload
	UploadDir string
```

Add to the `AppConfig = Config{...}` literal in `Load()` (after `AIPromptLog:`):
```go
		UploadDir:          getEnv("UPLOAD_DIR", "./uploads"),
```

- [ ] **Step 2: Add DDL to database.go**

Edit `backend/pkg/database/database.go`. Find the last successful migration (the `token_usage_content` block). Add **after** that block, before `log.Println("✅ 数据库迁移完成")`:

```go
	// 增量迁移：user_export_brand 用户导出品牌定制
	brandMigration := `
CREATE TABLE IF NOT EXISTS user_export_brand (
    user_id        VARCHAR(36)  PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    title          VARCHAR(20)  NOT NULL DEFAULT '',
    footer_text    VARCHAR(40)  NOT NULL DEFAULT '',
    logo_path      VARCHAR(200) NOT NULL DEFAULT '',
    watermark_mode VARCHAR(16)  NOT NULL DEFAULT 'none',
    watermark_text VARCHAR(30)  NOT NULL DEFAULT '',
    updated_at     TIMESTAMP    NOT NULL DEFAULT NOW()
);`
	if _, err := DB.Exec(brandMigration); err != nil {
		log.Fatalf("增量迁移失败 (user_export_brand): %v", err)
	}
```

(Note: `users.id` is `VARCHAR(36)` matching existing schema. If `users.id` is actually `UUID` type, change `user_id` to `UUID` to match. Verify by grepping the users CREATE TABLE in this file.)

- [ ] **Step 3: Create model**

Create `backend/internal/model/user_brand.go`:

```go
package model

import "time"

// ExportBrand represents a user's customization of export PNG/PDF branding.
type ExportBrand struct {
	UserID        string    `json:"-"`
	Title         string    `json:"title"`
	FooterText    string    `json:"footer_text"`
	LogoPath      string    `json:"-"`         // relative path under upload dir; not exposed directly
	LogoURL       string    `json:"logo_url"`  // computed: /static/uploads/<LogoPath> when LogoPath != ""
	WatermarkMode string    `json:"watermark_mode"`
	WatermarkText string    `json:"watermark_text"`
	UpdatedAt     time.Time `json:"-"`
}
```

- [ ] **Step 4: Create repository**

Create `backend/internal/repository/user_brand_repository.go`:

```go
package repository

import (
	"database/sql"
	"yuanju/internal/model"
	"yuanju/pkg/database"
)

// GetExportBrand returns the user's brand row, or a zero-value struct
// with empty strings and WatermarkMode="none" if no row exists.
func GetExportBrand(userID string) (model.ExportBrand, error) {
	var b model.ExportBrand
	b.UserID = userID
	err := database.DB.QueryRow(`
		SELECT title, footer_text, logo_path, watermark_mode, watermark_text, updated_at
		FROM user_export_brand WHERE user_id = $1`, userID).Scan(
		&b.Title, &b.FooterText, &b.LogoPath, &b.WatermarkMode, &b.WatermarkText, &b.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		b.WatermarkMode = "none"
		return b, nil
	}
	return b, err
}

// UpsertExportBrandText writes title/footer/watermark fields. Does NOT touch logo_path.
func UpsertExportBrandText(userID, title, footerText, watermarkMode, watermarkText string) error {
	_, err := database.DB.Exec(`
		INSERT INTO user_export_brand (user_id, title, footer_text, watermark_mode, watermark_text)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id) DO UPDATE SET
			title = EXCLUDED.title,
			footer_text = EXCLUDED.footer_text,
			watermark_mode = EXCLUDED.watermark_mode,
			watermark_text = EXCLUDED.watermark_text,
			updated_at = NOW()`,
		userID, title, footerText, watermarkMode, watermarkText)
	return err
}

// UpdateExportBrandLogo sets logo_path, creating the row if it doesn't exist.
// Returns the old logo_path (empty if none), so caller can delete the old file.
func UpdateExportBrandLogo(userID, logoPath string) (oldPath string, err error) {
	err = database.DB.QueryRow(`
		INSERT INTO user_export_brand (user_id, logo_path)
		VALUES ($1, $2)
		ON CONFLICT (user_id) DO UPDATE SET
			logo_path = EXCLUDED.logo_path,
			updated_at = NOW()
		RETURNING (SELECT logo_path FROM user_export_brand WHERE user_id = $1)`,
		userID, logoPath).Scan(&oldPath)
	// First-time insert returns NULL for the subquery; treat as empty.
	if err == sql.ErrNoRows {
		return "", nil
	}
	return oldPath, err
}

// ClearExportBrandLogo sets logo_path to empty, returns previous value for file cleanup.
func ClearExportBrandLogo(userID string) (oldPath string, err error) {
	err = database.DB.QueryRow(`
		UPDATE user_export_brand SET logo_path = '', updated_at = NOW()
		WHERE user_id = $1
		RETURNING (SELECT logo_path FROM user_export_brand WHERE user_id = $1 AND logo_path != '')`,
		userID).Scan(&oldPath)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return oldPath, err
}
```

(Note: the `RETURNING (SELECT ...)` trick to fetch pre-update value relies on the subquery being evaluated before the UPDATE. If the implementer finds this doesn't work reliably with the postgres driver, fall back to a two-statement `SELECT` then `UPDATE` inside a transaction. Behavior must be: caller learns the old `logo_path` to delete the old file.)

- [ ] **Step 5: Build to verify**

```bash
cd /Users/liujiming/web/yuanju/backend && go build ./...
```

Expected: clean build, no errors.

- [ ] **Step 6: Commit**

```bash
git -C /Users/liujiming/web/yuanju add backend/configs/config.go backend/pkg/database/database.go backend/internal/model/user_brand.go backend/internal/repository/user_brand_repository.go
git -C /Users/liujiming/web/yuanju commit -m "feat(brand): DDL + model + repository + UploadDir config"
```

---

## Task 5: Handler + Route Wiring — GREEN

**Files:**
- Create: `backend/internal/handler/user_brand_handler.go`
- Modify: `backend/cmd/api/main.go`

- [ ] **Step 1: Create handler file**

Create `backend/internal/handler/user_brand_handler.go`:

```go
package handler

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"yuanju/configs"
	"yuanju/internal/repository"
	"yuanju/pkg/ratelimit"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	maxLogoBytes       = 2 * 1024 * 1024 // 2MB
	logoSubDir         = "brand-logos"
	staticURLPrefix    = "/static/uploads"
	headerSampleLength = 12
)

var logoLimiter = ratelimit.New(10, time.Minute)

// BrandUpdateReq is the PUT /api/user/export-brand body.
type BrandUpdateReq struct {
	Title         string `json:"title"`
	FooterText    string `json:"footer_text"`
	WatermarkMode string `json:"watermark_mode"`
	WatermarkText string `json:"watermark_text"`
}

// BrandResponse is the GET / PUT response shape.
type BrandResponse struct {
	Title         string `json:"title"`
	FooterText    string `json:"footer_text"`
	LogoURL       string `json:"logo_url"`
	WatermarkMode string `json:"watermark_mode"`
	WatermarkText string `json:"watermark_text"`
}

// requireUserID is middleware-equivalent: pulls user_id from gin context (set by middleware.Auth)
// and aborts 401 if missing. Splits Auth() (token verification) from user-id presence check
// so this handler suite can be tested without a real JWT.
func requireUserID(c *gin.Context) {
	v, ok := c.Get("user_id")
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未登录，请先登录"})
		return
	}
	s, ok := v.(string)
	if !ok || s == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "无效的认证信息"})
		return
	}
	c.Set("user_id_str", s)
	c.Next()
}

func getUserIDStr(c *gin.Context) string {
	v, _ := c.Get("user_id_str")
	s, _ := v.(string)
	return s
}

func validateBrandUpdate(req BrandUpdateReq) error {
	if utf8.RuneCountInString(req.Title) > 20 {
		return errors.New("品牌标题不能超过 20 字符")
	}
	if utf8.RuneCountInString(req.FooterText) > 40 {
		return errors.New("底部文字不能超过 40 字符")
	}
	if utf8.RuneCountInString(req.WatermarkText) > 30 {
		return errors.New("水印文字不能超过 30 字符")
	}
	switch req.WatermarkMode {
	case "none", "bottom", "diagonal":
	default:
		return errors.New("水印模式不合法")
	}
	for _, s := range []string{req.Title, req.FooterText, req.WatermarkText} {
		if strings.ContainsAny(s, `<>"'&`) {
			return errors.New("文本包含不允许的字符")
		}
	}
	return nil
}

// detectImageType returns "png" / "jpg" / "webp" / "" based on magic bytes.
// header must be at least 12 bytes; returns "" if shorter or no match.
func detectImageType(header []byte) string {
	if len(header) < headerSampleLength {
		return ""
	}
	// PNG: 89 50 4E 47 0D 0A 1A 0A
	if header[0] == 0x89 && header[1] == 0x50 && header[2] == 0x4E && header[3] == 0x47 &&
		header[4] == 0x0D && header[5] == 0x0A && header[6] == 0x1A && header[7] == 0x0A {
		return "png"
	}
	// JPEG: FF D8 FF
	if header[0] == 0xFF && header[1] == 0xD8 && header[2] == 0xFF {
		return "jpg"
	}
	// WebP: RIFF????WEBP
	if header[0] == 'R' && header[1] == 'I' && header[2] == 'F' && header[3] == 'F' &&
		header[8] == 'W' && header[9] == 'E' && header[10] == 'B' && header[11] == 'P' {
		return "webp"
	}
	return ""
}

func logoURLFromPath(p string) string {
	if p == "" {
		return ""
	}
	return staticURLPrefix + "/" + p
}

// GET /api/user/export-brand
func GetExportBrand(c *gin.Context) {
	userID := getUserIDStr(c)
	b, err := repository.GetExportBrand(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取品牌设置失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": BrandResponse{
		Title:         b.Title,
		FooterText:    b.FooterText,
		LogoURL:       logoURLFromPath(b.LogoPath),
		WatermarkMode: b.WatermarkMode,
		WatermarkText: b.WatermarkText,
	}})
}

// PUT /api/user/export-brand
func UpdateExportBrand(c *gin.Context) {
	userID := getUserIDStr(c)
	var req BrandUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求体格式错误"})
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	req.FooterText = strings.TrimSpace(req.FooterText)
	req.WatermarkText = strings.TrimSpace(req.WatermarkText)
	if req.WatermarkMode == "" {
		req.WatermarkMode = "none"
	}
	if err := validateBrandUpdate(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := repository.UpsertExportBrandText(userID, req.Title, req.FooterText, req.WatermarkMode, req.WatermarkText); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存品牌设置失败"})
		return
	}
	GetExportBrand(c) // return refreshed state via the same response shape
}

// POST /api/user/export-brand/logo
func UploadExportBrandLogo(c *gin.Context) {
	userID := getUserIDStr(c)
	if !logoLimiter.Allow(userID) {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "上传过于频繁，请稍后重试"})
		return
	}
	if c.Request.ContentLength > maxLogoBytes {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "logo 文件不能超过 2MB"})
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 file 字段"})
		return
	}
	if file.Size > maxLogoBytes {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "logo 文件不能超过 2MB"})
		return
	}
	declaredCT := file.Header.Get("Content-Type")
	switch declaredCT {
	case "image/png", "image/jpeg", "image/webp":
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "仅支持 PNG / JPG / WebP 格式"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法读取上传文件"})
		return
	}
	defer src.Close()
	header := make([]byte, headerSampleLength)
	if _, err := src.Read(header); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无法读取文件头"})
		return
	}
	ext := detectImageType(header)
	if ext == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不是有效的图片文件"})
		return
	}
	// Cross-check declared MIME against detected magic bytes.
	expectedCT := map[string]string{"png": "image/png", "jpg": "image/jpeg", "webp": "image/webp"}[ext]
	if declaredCT != expectedCT {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件内容与声明的类型不一致"})
		return
	}

	relPath := filepath.Join(logoSubDir, uuid.NewString()+"."+ext)
	absPath := filepath.Join(configs.AppConfig.UploadDir, relPath)
	if err := c.SaveUploadedFile(file, absPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败"})
		return
	}

	oldPath, dbErr := repository.UpdateExportBrandLogo(userID, relPath)
	if dbErr != nil {
		_ = os.Remove(absPath) // don't leave orphan
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新品牌设置失败"})
		return
	}
	if oldPath != "" && oldPath != relPath {
		_ = os.Remove(filepath.Join(configs.AppConfig.UploadDir, oldPath))
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"logo_url": logoURLFromPath(relPath)}})
}

// DELETE /api/user/export-brand/logo
func DeleteExportBrandLogo(c *gin.Context) {
	userID := getUserIDStr(c)
	oldPath, err := repository.ClearExportBrandLogo(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "清除 logo 失败"})
		return
	}
	if oldPath != "" {
		_ = os.Remove(filepath.Join(configs.AppConfig.UploadDir, oldPath))
	}
	c.Status(http.StatusNoContent)
}

// Make fmt import used (path joining variant if needed by linter)
var _ = fmt.Sprintf
```

(Remove the unused `fmt` import + the `var _ = fmt.Sprintf` if go vet flags it. The implementer should let `goimports` clean unused imports.)

- [ ] **Step 2: Wire routes + static + mkdir in main.go**

Edit `backend/cmd/api/main.go`.

Find the existing `user := api.Group("/user", middleware.Auth())` block (around line 87). Replace it with:

```go
		user := api.Group("/user", middleware.Auth())
		{
			user.GET("/profile", handler.GetUserProfile)
			user.GET("/export-brand", handler.RequireUserID, handler.GetExportBrand)
			user.PUT("/export-brand", handler.RequireUserID, handler.UpdateExportBrand)
			user.POST("/export-brand/logo", handler.RequireUserID, handler.UploadExportBrandLogo)
			user.DELETE("/export-brand/logo", handler.RequireUserID, handler.DeleteExportBrandLogo)
		}
```

(Note: the `requireUserID` function in the handler file is package-private — to use it as middleware from `main.go`, rename it to `RequireUserID` (exported). Update `user_brand_handler.go` and the test file accordingly.)

Add imports (after `"yuanju/pkg/database"`):

```go
	"os"
	"path/filepath"
```

After `database.Migrate()` and before `seed.SeedLLMProviders()`, add:

```go
	// 确保 logo 上传目录存在
	if err := os.MkdirAll(filepath.Join(configs.AppConfig.UploadDir, "brand-logos"), 0755); err != nil {
		log.Fatalf("创建上传目录失败: %v", err)
	}
```

After `r.Use(middleware.CORS())`, before the `r.GET("/health", ...)` line, add:

```go
	// 公开静态文件托管（用户上传的品牌 logo）
	r.Static("/static/uploads", configs.AppConfig.UploadDir)
```

- [ ] **Step 3: Rename `requireUserID` → `RequireUserID` in handler + test**

In `backend/internal/handler/user_brand_handler.go` change `func requireUserID(c *gin.Context)` to `func RequireUserID(c *gin.Context)`.

In `backend/internal/handler/user_brand_handler_test.go` change `requireUserID` → `RequireUserID` in `TestBrandHandler_Unauthenticated`.

- [ ] **Step 4: Verify `google/uuid` is in go.mod**

```bash
cd /Users/liujiming/web/yuanju/backend && grep "google/uuid" go.mod
```

If absent, run `go get github.com/google/uuid` then `go mod tidy`.

- [ ] **Step 5: Run all backend tests — confirm GREEN**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./...
```

Expected: all green including:
- `TestLimiter_*` (3 tests in ratelimit)
- `TestValidateBrandUpdate_*` (5 tests)
- `TestDetectImageType_*` (5 tests)
- `TestBrandHandler_Unauthenticated` (1 test)
- All pre-existing tests still pass

Total new tests: 11 backend.

- [ ] **Step 6: Run server locally to smoke-test build**

```bash
cd /Users/liujiming/web/yuanju/backend && go build ./cmd/api && rm api
```

Expected: clean build.

- [ ] **Step 7: Commit**

```bash
git -C /Users/liujiming/web/yuanju add backend/internal/handler/user_brand_handler.go backend/internal/handler/user_brand_handler_test.go backend/cmd/api/main.go
git -C /Users/liujiming/web/yuanju commit -m "feat(brand): handler + route wiring + static mount"
```

---

## Task 6: Frontend Static Tests — RED

**Files:**
- Create: `frontend/tests/brand-settings.test.mjs`

- [ ] **Step 1: Write failing tests**

Create `frontend/tests/brand-settings.test.mjs`:

```js
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import test from 'node:test'
import assert from 'node:assert/strict'

const root = new URL('..', import.meta.url).pathname

function read(path) {
  return readFileSync(join(root, path), 'utf8')
}

test('ShareCard accepts optional brand prop', () => {
  const src = read('src/components/ShareCard.tsx')
  assert.match(src, /brand\?:\s*ExportBrand/)
})

test('ShareCard preserves 缘聚 命 理 default when brand title is empty', () => {
  const src = read('src/components/ShareCard.tsx')
  assert.match(src, /缘 聚 命 理/)
  assert.match(src, /brand\?\.title\s*\|\|\s*['"]缘 聚 命 理['"]/)
})

test('PrintLayout accepts optional brand prop', () => {
  const src = read('src/components/PrintLayout.tsx')
  assert.match(src, /brand\?:\s*ExportBrand/)
})

test('vite dev proxy includes /static for backend static files', () => {
  const src = read('vite.config.ts')
  assert.match(src, /['"]\/static['"]/)
  assert.match(src, /localhost:9002/)
})
```

- [ ] **Step 2: Run tests — confirm RED**

```bash
cd /Users/liujiming/web/yuanju/frontend && node --test tests/brand-settings.test.mjs
```

Expected: 4 tests fail — `ShareCard.tsx` doesn't contain `brand?: ExportBrand`, `PrintLayout.tsx` doesn't either, `vite.config.ts` doesn't have `/static` proxy.

- [ ] **Step 3: Commit RED**

```bash
git -C /Users/liujiming/web/yuanju add frontend/tests/brand-settings.test.mjs
git -C /Users/liujiming/web/yuanju commit -m "test(brand): static-regex tests for ShareCard/PrintLayout/vite (RED)"
```

---

## Task 7: Frontend Types + API Client + Vite Proxy

**Files:**
- Modify: `frontend/src/lib/api.ts`
- Modify: `frontend/vite.config.ts`

- [ ] **Step 1: Add ExportBrand type and brandAPI to api.ts**

Edit `frontend/src/lib/api.ts`. After the existing `userAPI` / `authAPI` definitions, add:

```ts
// ======= Export Brand =======
export interface ExportBrand {
  title: string
  footer_text: string
  logo_url: string
  watermark_mode: 'none' | 'bottom' | 'diagonal'
  watermark_text: string
}

export interface BrandUpdateInput {
  title: string
  footer_text: string
  watermark_mode: 'none' | 'bottom' | 'diagonal'
  watermark_text: string
}

export const brandAPI = {
  get: () => api.get<{ data: ExportBrand }>('/api/user/export-brand'),
  update: (body: BrandUpdateInput) =>
    api.put<{ data: ExportBrand }>('/api/user/export-brand', body),
  uploadLogo: (file: File) => {
    const fd = new FormData()
    fd.append('file', file)
    return api.post<{ data: { logo_url: string } }>('/api/user/export-brand/logo', fd, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
  },
  deleteLogo: () => api.delete('/api/user/export-brand/logo'),
}
```

- [ ] **Step 2: Add `/static` proxy to vite.config.ts**

Edit `frontend/vite.config.ts`. Inside the `proxy` object (after the `/api` block), add:

```ts
      '/static': {
        target: 'http://localhost:9002',
        changeOrigin: true,
      },
```

- [ ] **Step 3: Run vite proxy test — confirm GREEN for that test**

```bash
cd /Users/liujiming/web/yuanju/frontend && node --test tests/brand-settings.test.mjs
```

Expected: `vite dev proxy includes /static for backend static files` now PASSES. The other 3 still fail (ShareCard / PrintLayout untouched).

- [ ] **Step 4: Commit**

```bash
git -C /Users/liujiming/web/yuanju add frontend/src/lib/api.ts frontend/vite.config.ts
git -C /Users/liujiming/web/yuanju commit -m "feat(brand): ExportBrand type + brandAPI + vite /static proxy"
```

---

## Task 8: BrandPreviewCard Component

**Files:**
- Create: `frontend/src/components/BrandPreviewCard.tsx`

- [ ] **Step 1: Create component**

Create `frontend/src/components/BrandPreviewCard.tsx`:

```tsx
import type { ExportBrand } from '../lib/api'

interface Props {
  brand: ExportBrand
}

/**
 * Lightweight preview of the export card with current brand settings.
 * Mocks four-pillar area with gray placeholder bars; only purpose is to
 * give the user a live visual of how title/logo/footer/watermark render.
 */
export default function BrandPreviewCard({ brand }: Props) {
  const title = brand.title || '缘 聚 命 理'
  const footer = footerRightText(brand)
  const showDiagonal = brand.watermark_mode === 'diagonal' && brand.watermark_text.length > 0

  return (
    <div style={{
      position: 'relative',
      width: 320,
      background: '#fdf9f2',
      fontFamily: '"Noto Serif SC", serif',
      overflow: 'hidden',
      borderRadius: 8,
      border: '1px solid #e0cca0',
    }}>
      {/* Header */}
      <div style={{
        position: 'relative',
        background: 'linear-gradient(135deg, #2d1f14 0%, #4a3020 50%, #3a2416 100%)',
        padding: '16px 20px 14px',
        textAlign: 'center',
      }}>
        {brand.logo_url && (
          <img
            src={brand.logo_url}
            alt=""
            style={{
              position: 'absolute', left: 16, top: '50%', transform: 'translateY(-50%)',
              width: 32, height: 32, objectFit: 'contain',
            }}
          />
        )}
        <div style={{ color: '#e8c97c', fontSize: 16, letterSpacing: 5, fontWeight: 700 }}>{title}</div>
        <div style={{ color: '#c4a06a', fontSize: 10, marginTop: 4 }}>预览（占位）</div>
      </div>
      {/* Body: gray placeholders */}
      <div style={{ padding: '18px 16px', display: 'flex', gap: 6 }}>
        {[0, 1, 2, 3].map(i => (
          <div key={i} style={{ flex: 1, height: 60, background: '#e8dcc8', borderRadius: 4 }} />
        ))}
      </div>
      {/* Footer */}
      <div style={{
        background: 'linear-gradient(135deg, #2d1f14 0%, #4a3020 100%)',
        padding: '10px 16px',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        fontSize: 10,
      }}>
        <span style={{ color: '#9a7a5a' }}>仅供参考，不作决策依据</span>
        <span style={{ color: '#e8c97c' }}>{footer}</span>
      </div>
      {/* Diagonal watermark overlay */}
      {showDiagonal && (
        <div style={{
          position: 'absolute', inset: 0, pointerEvents: 'none',
          overflow: 'hidden', zIndex: 1,
        }}>
          <div style={{
            position: 'absolute',
            top: '-30%', left: '-30%', right: '-30%', bottom: '-30%',
            transform: 'rotate(-30deg)',
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fill, 140px)',
            gap: '40px 30px',
            opacity: 0.08,
            color: '#000',
            fontSize: 12,
            whiteSpace: 'nowrap',
          }}>
            {Array.from({ length: 40 }).map((_, i) => (
              <span key={i}>{brand.watermark_text}</span>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}

/**
 * Footer right-side text resolution. Mirror of ShareCard/PrintLayout logic
 * so preview matches actual export.
 */
function footerRightText(brand: ExportBrand): string {
  if (brand.watermark_mode === 'bottom' && brand.watermark_text && brand.footer_text) {
    return `${brand.footer_text} · ${brand.watermark_text}`
  }
  if (brand.watermark_mode === 'bottom' && brand.watermark_text) {
    return brand.watermark_text
  }
  return brand.footer_text || 'yuanju.com'
}
```

- [ ] **Step 2: Verify it builds (no test in this step)**

```bash
cd /Users/liujiming/web/yuanju/frontend && npx tsc --noEmit
```

Expected: clean type check.

- [ ] **Step 3: Commit**

```bash
git -C /Users/liujiming/web/yuanju add frontend/src/components/BrandPreviewCard.tsx
git -C /Users/liujiming/web/yuanju commit -m "feat(brand): BrandPreviewCard live mockup component"
```

---

## Task 9: BrandSettingsPage + CSS + Route + Entry

**Files:**
- Create: `frontend/src/pages/BrandSettingsPage.tsx`
- Create: `frontend/src/pages/BrandSettingsPage.css`
- Modify: `frontend/src/App.tsx`
- Modify: `frontend/src/pages/ProfilePage.tsx`

- [ ] **Step 1: Create settings page**

Create `frontend/src/pages/BrandSettingsPage.tsx`:

```tsx
import { useEffect, useRef, useState } from 'react'
import { Link, Navigate, useNavigate } from 'react-router-dom'
import { ArrowLeft } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'
import { brandAPI } from '../lib/api'
import type { ExportBrand } from '../lib/api'
import BrandPreviewCard from '../components/BrandPreviewCard'
import './BrandSettingsPage.css'

const DEFAULT_BRAND: ExportBrand = {
  title: '',
  footer_text: '',
  logo_url: '',
  watermark_mode: 'none',
  watermark_text: '',
}

const MAX_TITLE = 20
const MAX_FOOTER = 40
const MAX_WATERMARK = 30

export default function BrandSettingsPage() {
  const { user, isLoading: authLoading } = useAuth()
  const navigate = useNavigate()
  const [serverState, setServerState] = useState<ExportBrand>(DEFAULT_BRAND)
  const [draft, setDraft] = useState<ExportBrand>(DEFAULT_BRAND)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [uploading, setUploading] = useState(false)
  const [error, setError] = useState('')
  const fileInputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    if (!user) return
    brandAPI.get()
      .then(r => {
        setServerState(r.data.data)
        setDraft(r.data.data)
      })
      .catch(err => setError(err.message || '加载品牌设置失败'))
      .finally(() => setLoading(false))
  }, [user])

  if (!authLoading && !user) return <Navigate to="/login" replace />
  if (authLoading || loading) {
    return <main className="brand-page container page"><div className="brand-loading">加载中...</div></main>
  }

  const dirty =
    draft.title !== serverState.title ||
    draft.footer_text !== serverState.footer_text ||
    draft.watermark_mode !== serverState.watermark_mode ||
    draft.watermark_text !== serverState.watermark_text

  async function onSave() {
    setSaving(true)
    setError('')
    try {
      const r = await brandAPI.update({
        title: draft.title,
        footer_text: draft.footer_text,
        watermark_mode: draft.watermark_mode,
        watermark_text: draft.watermark_text,
      })
      setServerState(r.data.data)
      setDraft(r.data.data)
    } catch (e) {
      setError(e instanceof Error ? e.message : '保存失败')
    } finally {
      setSaving(false)
    }
  }

  async function onLogoChange(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    e.target.value = '' // allow re-selecting same file
    if (!file) return
    if (file.size > 2 * 1024 * 1024) {
      setError('logo 文件不能超过 2MB')
      return
    }
    if (!['image/png', 'image/jpeg', 'image/webp'].includes(file.type)) {
      setError('仅支持 PNG / JPG / WebP 格式')
      return
    }
    setUploading(true)
    setError('')
    try {
      const r = await brandAPI.uploadLogo(file)
      const next = { ...serverState, logo_url: r.data.data.logo_url }
      setServerState(next)
      setDraft(d => ({ ...d, logo_url: r.data.data.logo_url }))
    } catch (e) {
      setError(e instanceof Error ? e.message : '上传失败')
    } finally {
      setUploading(false)
    }
  }

  async function onLogoDelete() {
    if (!serverState.logo_url) return
    if (!window.confirm('确定要删除当前 logo 吗？')) return
    try {
      await brandAPI.deleteLogo()
      const next = { ...serverState, logo_url: '' }
      setServerState(next)
      setDraft(d => ({ ...d, logo_url: '' }))
    } catch (e) {
      setError(e instanceof Error ? e.message : '删除失败')
    }
  }

  async function onReset() {
    if (!window.confirm('重置为默认设置？已上传的 logo 也会被删除。')) return
    try {
      await brandAPI.update({ title: '', footer_text: '', watermark_mode: 'none', watermark_text: '' })
      if (serverState.logo_url) await brandAPI.deleteLogo()
      const fresh = await brandAPI.get()
      setServerState(fresh.data.data)
      setDraft(fresh.data.data)
    } catch (e) {
      setError(e instanceof Error ? e.message : '重置失败')
    }
  }

  return (
    <main className="brand-page container page">
      <header className="brand-page-header">
        <button className="brand-back" onClick={() => navigate(-1)}>
          <ArrowLeft size={16} /> 返回
        </button>
        <h1>导出品牌设置</h1>
      </header>

      {error && <div className="brand-error">{error}</div>}
      {dirty && <div className="brand-unsaved">有未保存的修改</div>}

      <section className="brand-section">
        <h2>顶部品牌</h2>
        <label className="brand-field">
          <span>品牌标题</span>
          <input
            type="text"
            maxLength={MAX_TITLE}
            value={draft.title}
            onChange={e => setDraft(d => ({ ...d, title: e.target.value }))}
            placeholder="留空则使用默认 "缘聚 命 理""
          />
          <small>{draft.title.length} / {MAX_TITLE}</small>
        </label>

        <div className="brand-logo-row">
          <div className="brand-logo-preview">
            {serverState.logo_url ? (
              <img src={serverState.logo_url} alt="当前 logo" />
            ) : (
              <span className="brand-logo-empty">未上传</span>
            )}
          </div>
          <div className="brand-logo-actions">
            <button
              type="button"
              onClick={() => fileInputRef.current?.click()}
              disabled={uploading}
            >
              {uploading ? '上传中...' : (serverState.logo_url ? '更换' : '上传')}
            </button>
            {serverState.logo_url && (
              <button type="button" onClick={onLogoDelete} disabled={uploading}>删除</button>
            )}
            <small>PNG / JPG / WebP，≤ 2MB</small>
          </div>
          <input
            ref={fileInputRef}
            type="file"
            accept="image/png,image/jpeg,image/webp"
            style={{ display: 'none' }}
            onChange={onLogoChange}
          />
        </div>
      </section>

      <section className="brand-section">
        <h2>底部品牌</h2>
        <label className="brand-field">
          <span>底部文字</span>
          <input
            type="text"
            maxLength={MAX_FOOTER}
            value={draft.footer_text}
            onChange={e => setDraft(d => ({ ...d, footer_text: e.target.value }))}
            placeholder="留空则使用默认 "yuanju.com""
          />
          <small>{draft.footer_text.length} / {MAX_FOOTER}</small>
        </label>
      </section>

      <section className="brand-section">
        <h2>水印</h2>
        <div className="brand-radio-row">
          {([
            ['none', '无水印'],
            ['bottom', '底部文字水印'],
            ['diagonal', '满页对角水印'],
          ] as const).map(([val, label]) => (
            <label key={val}>
              <input
                type="radio"
                name="watermark_mode"
                value={val}
                checked={draft.watermark_mode === val}
                onChange={() => setDraft(d => ({ ...d, watermark_mode: val }))}
              />
              {label}
            </label>
          ))}
        </div>
        <label className="brand-field">
          <span>水印文字</span>
          <input
            type="text"
            maxLength={MAX_WATERMARK}
            value={draft.watermark_text}
            onChange={e => setDraft(d => ({ ...d, watermark_text: e.target.value }))}
            disabled={draft.watermark_mode === 'none'}
            placeholder={draft.watermark_mode === 'none' ? '请先选择水印模式' : '建议填写'}
          />
          <small>{draft.watermark_text.length} / {MAX_WATERMARK}</small>
        </label>
      </section>

      <section className="brand-section brand-preview-section">
        <h2>预览</h2>
        <BrandPreviewCard brand={draft} />
      </section>

      <div className="brand-footer-actions">
        <button type="button" className="btn btn-ghost" onClick={onReset}>重置默认</button>
        <button
          type="button"
          className="btn btn-primary"
          onClick={onSave}
          disabled={!dirty || saving}
        >
          {saving ? '保存中...' : '保存'}
        </button>
      </div>

      <p className="brand-tip">
        提示：设置不影响 ResultPage 网页本身展示，仅作用于"保存图片"与"导出 PDF"的产物。
      </p>

      <Link to="/profile" className="brand-bottom-link">返回个人中心</Link>
    </main>
  )
}
```

- [ ] **Step 2: Create CSS**

Create `frontend/src/pages/BrandSettingsPage.css`:

```css
.brand-page {
  max-width: 720px;
  margin: 0 auto;
  padding-bottom: 80px;
}
.brand-page-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 24px;
}
.brand-page-header h1 {
  font-size: 20px;
  margin: 0;
  color: var(--text-primary, #2a1a0a);
}
.brand-back {
  background: none;
  border: 1px solid var(--border, #e0cca0);
  padding: 6px 12px;
  border-radius: 6px;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  color: var(--text-muted, #5a3a1a);
}
.brand-loading {
  padding: 40px;
  text-align: center;
  color: var(--text-muted);
}
.brand-error {
  background: #fdf0f0;
  border: 1px solid #c0392b;
  color: #c0392b;
  padding: 10px 14px;
  border-radius: 6px;
  margin-bottom: 16px;
  font-size: 13px;
}
.brand-unsaved {
  background: #fffbe6;
  border: 1px solid #d4b896;
  color: #7a5c2e;
  padding: 8px 14px;
  border-radius: 6px;
  margin-bottom: 16px;
  font-size: 13px;
}
.brand-section {
  background: #fdf9f2;
  border: 1px solid var(--border, #e0cca0);
  border-radius: 8px;
  padding: 18px 20px;
  margin-bottom: 18px;
}
.brand-section h2 {
  font-size: 14px;
  font-weight: 700;
  color: var(--text-primary, #2a1a0a);
  margin: 0 0 14px;
  letter-spacing: 1px;
}
.brand-field {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-bottom: 12px;
}
.brand-field span {
  font-size: 13px;
  color: var(--text-muted, #5a3a1a);
}
.brand-field input[type="text"] {
  padding: 8px 12px;
  border: 1px solid var(--border, #e0cca0);
  border-radius: 6px;
  font-size: 14px;
  background: #fff;
}
.brand-field input[type="text"]:disabled {
  background: #f5efe3;
  color: #999;
}
.brand-field small {
  align-self: flex-end;
  font-size: 11px;
  color: #999;
}
.brand-logo-row {
  display: flex;
  align-items: flex-start;
  gap: 16px;
  margin-top: 8px;
}
.brand-logo-preview {
  width: 80px;
  height: 80px;
  border: 1px dashed var(--border, #e0cca0);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #fff;
  overflow: hidden;
}
.brand-logo-preview img {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}
.brand-logo-empty {
  font-size: 11px;
  color: #aaa;
}
.brand-logo-actions {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.brand-logo-actions button {
  background: #fdf9f2;
  border: 1px solid var(--border, #e0cca0);
  border-radius: 6px;
  padding: 6px 16px;
  cursor: pointer;
  font-size: 13px;
}
.brand-logo-actions small {
  font-size: 11px;
  color: #aaa;
}
.brand-radio-row {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  margin-bottom: 12px;
  font-size: 13px;
  color: var(--text-primary, #2a1a0a);
}
.brand-radio-row label {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
}
.brand-preview-section {
  display: flex;
  flex-direction: column;
  align-items: center;
}
.brand-footer-actions {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  margin-top: 8px;
}
.brand-tip {
  font-size: 12px;
  color: #aaa;
  text-align: center;
  margin-top: 16px;
}
.brand-bottom-link {
  display: block;
  text-align: center;
  margin-top: 24px;
  color: var(--text-muted);
  font-size: 13px;
}
```

- [ ] **Step 3: Add route to App.tsx**

Edit `frontend/src/App.tsx`. Add import (after `import ProfilePage`):

```tsx
import BrandSettingsPage from './pages/BrandSettingsPage'
```

In `<Routes>` block, add (after the `/profile` route):

```tsx
<Route path="/settings/brand" element={<><Navbar /><BottomNav /><BrandSettingsPage /></>} />
```

- [ ] **Step 4: Add entry to ProfilePage**

Edit `frontend/src/pages/ProfilePage.tsx`. Find the section where existing entry cards are rendered (stats / continue / recent). Add a new entry card linking to `/settings/brand` consistent with existing card styling. Suggested insertion: a new section after recent charts.

Look for a closing `</section>` near the end of the rendered tree before `</main>` and add:

```tsx
<section className="profile-panel">
  <Link to="/settings/brand" className="profile-action-row">
    <span>导出品牌设置</span>
    <span className="profile-action-hint">自定义导出图片/PDF 的 logo、标题与水印</span>
    <ArrowRight size={16} />
  </Link>
</section>
```

(If `ArrowRight` is not already imported, add to the lucide-react import line. If `profile-action-row` class doesn't exist, the implementer should pick a class that matches the existing entry card styling in ProfilePage.css — inspect ProfilePage.css for an existing pattern. If unclear, use inline style with `style={{ display: 'flex', justifyContent: 'space-between', padding: 16, color: 'inherit', textDecoration: 'none' }}`.)

- [ ] **Step 5: Build + typecheck**

```bash
cd /Users/liujiming/web/yuanju/frontend && npm run build
```

Expected: build succeeds.

- [ ] **Step 6: Commit**

```bash
git -C /Users/liujiming/web/yuanju add frontend/src/pages/BrandSettingsPage.tsx frontend/src/pages/BrandSettingsPage.css frontend/src/App.tsx frontend/src/pages/ProfilePage.tsx
git -C /Users/liujiming/web/yuanju commit -m "feat(brand): settings page + /settings/brand route + profile entry"
```

---

## Task 10: ShareCard Accepts Brand Prop

**Files:**
- Modify: `frontend/src/components/ShareCard.tsx`

- [ ] **Step 1: Import ExportBrand type**

At top of `ShareCard.tsx`, add to imports:

```tsx
import type { StructuredReport, ExportBrand } from '../lib/api'
```

- [ ] **Step 2: Add brand prop to interface**

In `ShareCardProps`, add at end:

```tsx
  brand?: ExportBrand | null
```

- [ ] **Step 3: Add brand to destructuring**

Find `const { ... } = props` block and add `brand` to it.

- [ ] **Step 4: Compute resolved values at top of component (right after `const chapterDefs = [...]`)**

```tsx
  const resolvedTitle = brand?.title || '缘 聚 命 理'
  const resolvedFooter = (() => {
    if (!brand) return 'yuanju.com'
    if (brand.watermark_mode === 'bottom' && brand.watermark_text && brand.footer_text) {
      return `${brand.footer_text} · ${brand.watermark_text}`
    }
    if (brand.watermark_mode === 'bottom' && brand.watermark_text) {
      return brand.watermark_text
    }
    return brand.footer_text || 'yuanju.com'
  })()
  const showDiagonalMark = brand?.watermark_mode === 'diagonal' && (brand?.watermark_text?.length ?? 0) > 0
```

- [ ] **Step 5: Add `position: 'relative'` to root div**

The root `<div ref={ref} style={{...}}>` needs `position: 'relative'` added to the style object (so the watermark overlay can absolute-position over it).

- [ ] **Step 6: Render logo + title override in header**

The current header (around L120-144) renders a centered `<div>缘 聚 命 理</div>`. Wrap the header section to add an optional logo and use `resolvedTitle`:

Original gradient header `<div>` — change the outer div to add `position: 'relative'` to its style. Inside, before the existing centered title div, conditionally render the logo:

```tsx
        {brand?.logo_url && (
          <img
            src={brand.logo_url}
            alt=""
            crossOrigin="anonymous"
            style={{
              position: 'absolute',
              left: 24,
              top: '50%',
              transform: 'translateY(-50%)',
              width: 40,
              height: 40,
              objectFit: 'contain',
            }}
          />
        )}
```

And change the centered title text to use `resolvedTitle`:

```tsx
        <div style={{
          color: '#e8c97c',
          fontSize: 20,
          letterSpacing: 6,
          fontWeight: 700,
          textAlign: 'center',
          fontFamily: '"Noto Serif SC", serif',
          marginBottom: 6,
        }}>
          {resolvedTitle}
        </div>
```

(Make sure the gradient `<div>` containing logo+title has `position: 'relative'` so the absolute-positioned logo aligns to it.)

- [ ] **Step 7: Replace footer right-side text**

In the footer block (around L284-289), change:

```tsx
        <span style={{
          fontSize: 12, color: '#e8c97c', letterSpacing: 1,
          fontFamily: '"Noto Serif SC", serif',
        }}>
          yuanju.com
        </span>
```

To use `resolvedFooter`:

```tsx
        <span style={{
          fontSize: 12, color: '#e8c97c', letterSpacing: 1,
          fontFamily: '"Noto Serif SC", serif',
        }}>
          {resolvedFooter}
        </span>
```

- [ ] **Step 8: Add diagonal watermark overlay at the end of the root div (just before `</div>` of `<div ref={ref}>`)**

```tsx
      {showDiagonalMark && brand && (
        <div style={{
          position: 'absolute', inset: 0, pointerEvents: 'none',
          overflow: 'hidden', zIndex: 1,
        }}>
          <div style={{
            position: 'absolute',
            top: '-30%', left: '-30%', right: '-30%', bottom: '-30%',
            transform: 'rotate(-30deg)',
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fill, 180px)',
            gap: '60px 40px',
            opacity: 0.06,
            color: '#000',
            fontSize: 14,
            fontFamily: '"Noto Sans SC", sans-serif',
            whiteSpace: 'nowrap',
          }}>
            {Array.from({ length: 60 }).map((_, i) => (
              <span key={i}>{brand.watermark_text}</span>
            ))}
          </div>
        </div>
      )}
```

- [ ] **Step 9: Run frontend tests — confirm 2 of the 4 now pass**

```bash
cd /Users/liujiming/web/yuanju/frontend && node --test tests/brand-settings.test.mjs
```

Expected:
- ✅ `ShareCard accepts optional brand prop`
- ✅ `ShareCard preserves 缘聚 命 理 default when brand title is empty`
- ❌ `PrintLayout accepts optional brand prop` (still red — fix in Task 11)
- ✅ `vite dev proxy includes /static for backend static files`

- [ ] **Step 10: Commit**

```bash
git -C /Users/liujiming/web/yuanju add frontend/src/components/ShareCard.tsx
git -C /Users/liujiming/web/yuanju commit -m "feat(brand): ShareCard accepts brand prop, supports logo/title/footer/diagonal watermark"
```

---

## Task 11: PrintLayout Accepts Brand Prop

**Files:**
- Modify: `frontend/src/components/PrintLayout.tsx`

Apply the same 9 transformations as Task 10 to PrintLayout (logo top-left in header, `resolvedTitle` replaces hardcoded header brand text, `resolvedFooter` replaces hardcoded footer URL/disclaimer right-side, diagonal watermark overlay over entire rendered area, root div gets `position: 'relative'`).

- [ ] **Step 1: Add import**

```tsx
import type { ExportBrand } from '../lib/api'
```

- [ ] **Step 2: Add to PrintLayoutProps interface**

```tsx
  brand?: ExportBrand | null
```

- [ ] **Step 3: Add `brand` to function destructuring**

In the `export default function PrintLayout({ ... })` signature, add `brand`.

- [ ] **Step 4: Compute resolved values inside the function (near top, after existing const declarations)**

```tsx
  const resolvedTitle = brand?.title || '缘聚命理'
  const resolvedFooter = (() => {
    if (!brand) return 'yuanju.com'
    if (brand.watermark_mode === 'bottom' && brand.watermark_text && brand.footer_text) {
      return `${brand.footer_text} · ${brand.watermark_text}`
    }
    if (brand.watermark_mode === 'bottom' && brand.watermark_text) {
      return brand.watermark_text
    }
    return brand.footer_text || 'yuanju.com'
  })()
  const showDiagonalMark = brand?.watermark_mode === 'diagonal' && (brand?.watermark_text?.length ?? 0) > 0
```

- [ ] **Step 5: Locate and modify PrintLayout's header**

Find PrintLayout's header section that displays the brand. The implementer should:
- Search for occurrences of `缘聚` or hardcoded brand strings in `PrintLayout.tsx`
- Wrap header div in `position: 'relative'` if not already
- Conditionally render `<img>` for `brand?.logo_url` positioned top-left 40×40
- Replace the hardcoded brand text with `resolvedTitle`
- Replace any hardcoded `yuanju.com` in footer area with `resolvedFooter`

The exact existing line numbers were ~207-220 (header badges) and ~653-654 (terminology) per the historical refactor; but lines may have shifted. Search for hardcoded strings before editing.

- [ ] **Step 6: Add `position: 'relative'` to the root div and append diagonal watermark overlay**

Find the outermost `<div>` returned by PrintLayout, ensure its style contains `position: 'relative'`. Just before the closing `</div>` of that root, add the same diagonal watermark block as in ShareCard Task 10 Step 8.

- [ ] **Step 7: Build + run tests**

```bash
cd /Users/liujiming/web/yuanju/frontend && node --test tests/brand-settings.test.mjs && npm run build
```

Expected: All 4 brand-settings tests PASS. Build succeeds.

- [ ] **Step 8: Commit**

```bash
git -C /Users/liujiming/web/yuanju add frontend/src/components/PrintLayout.tsx
git -C /Users/liujiming/web/yuanju commit -m "feat(brand): PrintLayout accepts brand prop, supports logo/title/footer/diagonal watermark"
```

---

## Task 12: ResultPage Wires Brand to Both Exports

**Files:**
- Modify: `frontend/src/pages/ResultPage.tsx`

- [ ] **Step 1: Import**

Add to imports near other api imports:

```tsx
import { brandAPI } from '../lib/api'
import type { ExportBrand } from '../lib/api'
```

- [ ] **Step 2: Add brand state + fetch**

Inside the `ResultPage` function, near other `useState` declarations, add:

```tsx
  const [brand, setBrand] = useState<ExportBrand | null>(null)
```

After existing useEffect blocks (anywhere among them), add:

```tsx
  useEffect(() => {
    if (!user) return
    brandAPI.get()
      .then(r => setBrand(r.data.data))
      .catch(() => setBrand(null))
  }, [user])
```

- [ ] **Step 3: Pass `brand` prop to ShareCard**

Find `<ShareCard ref={shareCardRef} ... />` (around line 1164). Add a new prop line:

```tsx
          brand={brand}
```

- [ ] **Step 4: Pass `brand` prop to PrintLayout**

Find `<PrintLayout ... />` (around line 1263). Add inside the props:

```tsx
            brand={brand}
```

- [ ] **Step 5: Build + lint**

```bash
cd /Users/liujiming/web/yuanju/frontend && npm run build && npm run lint
```

Expected: clean build, no new lint errors.

- [ ] **Step 6: Commit**

```bash
git -C /Users/liujiming/web/yuanju add frontend/src/pages/ResultPage.tsx
git -C /Users/liujiming/web/yuanju commit -m "feat(brand): ResultPage fetches brand and forwards to ShareCard/PrintLayout"
```

---

## Task 13: Final Validation

- [ ] **Step 1: All backend tests pass**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./...
```

Expected: 0 failures. New tests include 3 ratelimit + 5 validate + 5 detectImageType + 1 unauthenticated = 14. (Spec said 11; we have 14 actually because pure helpers split better than the original integration plan. Acceptable — more coverage same scope.)

Update plan note: Final test count is 14 backend + 4 frontend = 18.

- [ ] **Step 2: All frontend tests pass**

```bash
cd /Users/liujiming/web/yuanju/frontend && for f in tests/*.test.mjs; do node --test "$f" || echo "FAIL: $f"; done
```

Expected: 0 FAIL lines.

- [ ] **Step 3: Frontend build + lint**

```bash
cd /Users/liujiming/web/yuanju/frontend && npm run build && npm run lint
```

Expected: clean.

- [ ] **Step 4: Manual smoke test (local stack)**

If docker is up: hit `http://localhost:5200/settings/brand` after logging in. Fill title="清雨堂", upload a small PNG, pick mode="diagonal", text="清雨堂", save. Go to a result page, click "保存分享图" — verify exported PNG has user title (no "缘聚"), logo top-left, diagonal watermark, "yuanju.com" gone. Click "导出 PDF" — verify same on PDF.

(If no local stack: skip and rely on test gate.)

- [ ] **Step 5: Verify all spec acceptance criteria**

Re-read `docs/superpowers/specs/2026-05-18-export-brand-customization-design.md` §11 (Acceptance). Confirm each item is observably true. Each item not satisfied → return to relevant task.

- [ ] **Step 6: Push branch**

```bash
git -C /Users/liujiming/web/yuanju push -u origin feat/export-brand-customization
```

- [ ] **Step 7: Open PR (or hand off to user)**

Report to user: branch ready at `feat/export-brand-customization` with N commits, all tests passing, ready for review/merge.

---

## Spec Coverage Audit

| Spec section | Task | Note |
|---|---|---|
| §4.1 DDL `user_export_brand` | Task 4 | DDL added to `database.Migrate()` |
| §4.2 Upload link | Tasks 4, 5 | mkdir in main.go, Static mount, handler with magic-bytes |
| §4.3 Export data flow | Task 12 | ResultPage fetches + passes to ShareCard/PrintLayout |
| §5 File inventory | Tasks 1-12 | 16 files; all covered |
| §6.1 GET endpoint | Task 5 | `GetExportBrand` returns empty for no-row |
| §6.2 PUT endpoint | Task 5 | `UpdateExportBrand` validates + upserts |
| §6.3 POST logo endpoint | Task 5 | 3-layer validation + magic bytes + rate limit |
| §6.4 DELETE logo | Task 5 | `DeleteExportBrandLogo` |
| §6.5 Static mount | Task 5 | `r.Static("/static/uploads", ...)` |
| §7 Settings page UX | Task 9 | `BrandSettingsPage.tsx` + CSS |
| §8 Rendering changes | Tasks 10, 11 | ShareCard + PrintLayout brand prop + diagonal watermark |
| §9 Security | Tasks 2 (rate), 5 (magic bytes + validation) | All 3 layers implemented |
| §10 Tests | Tasks 1, 3, 6 | Adjusted from spec's 11 to 14 pure-function tests + 4 frontend (see Test Strategy Note) |
| §11 Acceptance | Task 13 | Manual + automated verification |

---

## Self-Review Checklist (Done by Plan Author)

**Placeholder scan:** Searched for "TBD", "TODO", "fill in", "etc." — none found.

**Type consistency:**
- Backend `ExportBrand` (`internal/model/user_brand.go`) uses snake_case JSON tags matching frontend `ExportBrand` (`lib/api.ts`). ✓
- `BrandUpdateReq` (handler) matches `BrandUpdateInput` (frontend). ✓
- `validateBrandUpdate`, `detectImageType`, `Limiter.Allow`, `RequireUserID` — all named consistently across test (Task 1, 3) and impl (Task 2, 5). ✓
- `requireUserID` → `RequireUserID` rename in Task 5 Step 3 is explicit, won't drift.

**Test naming:**
- Task 1 tests: `TestLimiter_*` × 3
- Task 3 tests: `TestValidateBrandUpdate_*` × 5, `TestDetectImageType_*` × 5, `TestBrandHandler_Unauthenticated` × 1 = 11 test names actually defined
- Spec said 11 backend tests; plan delivers 14 (counted as 11 in tests listed but 3 ratelimit also count). Final count 14 — noted in Task 13.

**Spec deviations:**
- Test design swapped from integration to pure-helper unit tests (justified in Test Strategy Note).
- `requireUserID` middleware helper added (not in spec; needed to make handler-level auth-smoke test feasible without JWT).

No other deviations.
