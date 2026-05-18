# Logo Wordmark Mode Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 给 `user_export_brand` 加 `logo_mode` 字段（`icon` / `wordmark`），让用户可以上传横版 wordmark 取代 PDF/PNG 导出产物中的文字标题；后端 + 前端 + Crop UI + 三个渲染组件全部按 mode 分支。

**Architecture:** 后端走"现有 schema 增量 ALTER + handler/repo 增加一个文本字段"轻量路线。前端 LogoCropModal 增加预设比例 chip 行（2:1 / 3:1 / 4:1，默认 3:1）— 这是对 OpenSpec spec R3 "free crop with clamp" 的务实偏离，由用户在 explore 中拍板（react-easy-crop 不支持真正的 free aspect，切库代价过大）。ShareCard / PrintLayout / BrandPreviewCard 在 `brand.logo_mode === 'wordmark' && brand.logo_url` 同时成立时切换布局：logo 图替换文字标题。

**Tech Stack:** Go + lib/pq（无 ORM 直接 SQL）、Gin handler、PostgreSQL 增量 ALTER；React 19 + TypeScript + Vite + CSS Variables（无 UI 框架）；react-easy-crop（已用）；Node `node:test` 静态正则测试 + Go testing 单测。

**Repo paths:**
- Repo root: `/Users/liujiming/web/yuanju`
- Frontend cwd: `/Users/liujiming/web/yuanju/frontend`
- Backend cwd: `/Users/liujiming/web/yuanju/backend`
- Git 命令使用 `git -C /Users/liujiming/web/yuanju ...`
- 当前分支：`feat/export-brand-customization`，起始 HEAD：`d82a927`

**OpenSpec 引用:**
- Change folder: `openspec/changes/logo-wordmark-mode/`
- Capability spec: `openspec/changes/logo-wordmark-mode/specs/logo-wordmark-mode/spec.md`（6 Requirements / 14 Scenarios — 本 plan 任务对照覆盖）

**对 spec 的务实偏离（用户已确认走 B 方案）：**
- spec.md R3 "free crop within [1.5, 6]" 实现为**预设档 2:1 / 3:1 / 4:1**（默认 3:1）
- 选 chip 即固定 aspect，不会出现"用户拖到 7:1 → snap 回 6:1"的清晰场景，因为用户压根选不到 6:1
- spec.md R3 第 2-3 个 Scenario（"drag past cap → snap"）的语义改成"用户只能从预设档选 → 越界不可能发生"
- 影响：max-width 上限从 768 收紧到 512（4 × 128 = 512），文件更小
- 后续如需更细粒度比例（2.5:1、3.2:1 等），切 `react-image-crop` 是清晰的下一步增量

---

## File Structure

| File | Role | Change |
|------|------|--------|
| `backend/pkg/database/database.go` | DDL | 追加 ALTER TABLE 增量迁移 `logo_mode` 列 |
| `backend/internal/model/user_brand.go` | Model | `ExportBrand` 加 `LogoMode` 字段 |
| `backend/internal/repository/user_brand_repository.go` | Repo | `GetExportBrand` SELECT/Scan、`UpsertExportBrandText` INSERT/UPDATE 都加 `logo_mode` |
| `backend/internal/handler/user_brand_handler.go` | Handler | `BrandUpdateReq` + `BrandResponse` 加 `LogoMode`；`validateBrandUpdate` 校验；`UpdateExportBrand`/`buildBrandResponse` 传递 |
| `backend/internal/handler/user_brand_handler_test.go` | Test | 加 `validateBrandUpdate` 关于 `logo_mode` 的单测 |
| `frontend/src/lib/api.ts` | API types | `ExportBrand` + `BrandUpdateInput` 加 `logo_mode: 'icon' \| 'wordmark'` |
| `frontend/src/components/LogoCropModal.tsx` | Crop modal | 加 `mode` prop、预设 chip 行、输出尺寸分支 |
| `frontend/src/pages/BrandSettingsPage.tsx` | Settings UI | 加 Logo 模式 radio、传 mode 到 modal、计算 aspect 不匹配警告、`dirty` 加 mode 差 |
| `frontend/src/pages/BrandSettingsPage.css` | Settings CSS | 加 `.brand-warning` + `.brand-mode-hint` + `.logo-crop-aspect-chip*` |
| `frontend/src/components/LogoCropModal.css` | Modal CSS | 加 `.logo-crop-aspect-chips` 横排 chip 容器 |
| `frontend/src/components/ShareCard.tsx` | PNG render | 顶部品牌栏按 mode 分支 |
| `frontend/src/components/PrintLayout.tsx` | PDF render | 每页页眉 + 封面 banner 按 mode 分支 |
| `frontend/src/components/BrandPreviewCard.tsx` | Preview render | 顶部按 mode 分支（与 ShareCard 同形态） |
| `frontend/tests/brand-settings.test.mjs` | Tests | 加 6 个静态正则断言：6 个文件中 logo_mode 引用 |

---

## Task 0: 确认基线

**Files:** 无修改

- [ ] **Step 1: 确认分支与 HEAD**

```bash
git -C /Users/liujiming/web/yuanju status
git -C /Users/liujiming/web/yuanju log --oneline -1
```

Expected:
- Branch: `feat/export-brand-customization`
- 工作树干净
- HEAD: `d82a927 docs(openspec): propose logo-wordmark-mode change`

- [ ] **Step 2: sanity-check backend 现状**

```bash
grep -n 'user_export_brand' /Users/liujiming/web/yuanju/backend/pkg/database/database.go | head -3
grep -n 'LogoMode\|logo_mode' /Users/liujiming/web/yuanju/backend/internal/model/user_brand.go
```

Expected:
- `database.go` 含 `CREATE TABLE IF NOT EXISTS user_export_brand` 块
- `user_brand.go` **没有** `LogoMode` 字段（grep 应空）

如果 grep `LogoMode` 已经命中，说明有人提前改过，停下来 BLOCKED。

---

## Task 1: RED — backend validateBrandUpdate 加 logo_mode 测试

**Files:**
- Modify: `backend/internal/handler/user_brand_handler_test.go`

- [ ] **Step 1: 看一下现有测试函数找好插入点**

```bash
grep -n 'TestValidateBrandUpdate\|func Test' /Users/liujiming/web/yuanju/backend/internal/handler/user_brand_handler_test.go
```

Expected: 输出现有 `Test*` 函数列表。本任务在末尾追加新函数。

- [ ] **Step 2: 在文件末尾追加这段 Go 测试**

```go

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
```

- [ ] **Step 3: 跑测试，确认它 *无法编译*（RED）**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./internal/handler/ -run TestValidateBrandUpdate_LogoMode 2>&1 | head -20
```

Expected: 编译错误 — `BrandUpdateReq` 上没有 `LogoMode` 字段。错误形如 `unknown field 'LogoMode' in struct literal of type handler.BrandUpdateReq`。

如果意外编译过了，停下来 — 说明结构体已经有这个字段，跟 Task 0 sanity-check 矛盾。

- [ ] **Step 4: 提交 RED**

```bash
git -C /Users/liujiming/web/yuanju add backend/internal/handler/user_brand_handler_test.go
git -C /Users/liujiming/web/yuanju -c commit.gpgsign=false commit -m "test(brand-handler): logo_mode validation cases (RED, compile-fails)"
```

---

## Task 2: GREEN — handler 加 LogoMode 字段 + validation

**Files:**
- Modify: `backend/internal/handler/user_brand_handler.go`

- [ ] **Step 1: 在 BrandUpdateReq 结构体末尾加 LogoMode 字段**

找到（约 30 行附近）：

```go
type BrandUpdateReq struct {
	Title         string `json:"title"`
	FooterText    string `json:"footer_text"`
	WatermarkMode string `json:"watermark_mode"`
	WatermarkText string `json:"watermark_text"`
}
```

替换为：

```go
type BrandUpdateReq struct {
	Title         string `json:"title"`
	FooterText    string `json:"footer_text"`
	WatermarkMode string `json:"watermark_mode"`
	WatermarkText string `json:"watermark_text"`
	LogoMode      string `json:"logo_mode"`
}
```

- [ ] **Step 2: 在 BrandResponse 结构体加 LogoMode 字段**

找到（约 38 行附近）：

```go
type BrandResponse struct {
	Title         string `json:"title"`
	FooterText    string `json:"footer_text"`
	LogoURL       string `json:"logo_url"`
	WatermarkMode string `json:"watermark_mode"`
	WatermarkText string `json:"watermark_text"`
}
```

替换为：

```go
type BrandResponse struct {
	Title         string `json:"title"`
	FooterText    string `json:"footer_text"`
	LogoURL       string `json:"logo_url"`
	LogoMode      string `json:"logo_mode"`
	WatermarkMode string `json:"watermark_mode"`
	WatermarkText string `json:"watermark_text"`
}
```

- [ ] **Step 3: 在 validateBrandUpdate 加 logo_mode switch**

找到（约 74 行附近，validateBrandUpdate 函数体内 `switch req.WatermarkMode` 之后、return 之前）：

```go
	switch req.WatermarkMode {
	case "none", "bottom", "diagonal":
	default:
		return errors.New("水印模式不合法")
	}
```

在它**正下方**插入：

```go
	switch req.LogoMode {
	case "", "icon", "wordmark":
	default:
		return errors.New("logo 模式不合法")
	}
```

- [ ] **Step 4: UpdateExportBrand 默认空 logo_mode 为 icon、把字段传给 repo**

找到 `UpdateExportBrand` 函数（约 154 行附近）。在 `if req.WatermarkMode == "" { req.WatermarkMode = "none" }` **下面**追加：

```go
	if req.LogoMode == "" {
		req.LogoMode = "icon"
	}
```

然后修改 `UpsertExportBrandText` 调用（同函数内，约 171 行）。当前：

```go
	if err := repository.UpsertExportBrandText(userID, req.Title, req.FooterText, req.WatermarkMode, req.WatermarkText); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存品牌设置失败"})
		return
	}
```

改为：

```go
	if err := repository.UpsertExportBrandText(userID, req.Title, req.FooterText, req.WatermarkMode, req.WatermarkText, req.LogoMode); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存品牌设置失败"})
		return
	}
```

（注意：repo 函数签名我们在 Task 5 改；现在这里先按新签名写，go 编译会暂时挂 — 任务依赖图是: Task 2 编译挂 → Task 5 GREEN，因此 Task 2 的"绿"判定要延后到 Task 5 完成后。）

- [ ] **Step 5: buildBrandResponse emit LogoMode**

找到（约 129 行）：

```go
	return BrandResponse{
		Title:         b.Title,
		FooterText:    b.FooterText,
		LogoURL:       logoURLFromPath(b.LogoPath),
		WatermarkMode: b.WatermarkMode,
		WatermarkText: b.WatermarkText,
	}, nil
```

改为：

```go
	return BrandResponse{
		Title:         b.Title,
		FooterText:    b.FooterText,
		LogoURL:       logoURLFromPath(b.LogoPath),
		LogoMode:      b.LogoMode,
		WatermarkMode: b.WatermarkMode,
		WatermarkText: b.WatermarkText,
	}, nil
```

- [ ] **Step 6: 此时 *不要* 跑测试 — model + repo 还没改完。直接提交。**

```bash
git -C /Users/liujiming/web/yuanju add backend/internal/handler/user_brand_handler.go
git -C /Users/liujiming/web/yuanju -c commit.gpgsign=false commit -m "feat(brand-handler): wire logo_mode into request/response and validation"
```

---

## Task 3: DB schema 增量 ALTER

**Files:**
- Modify: `backend/pkg/database/database.go`

- [ ] **Step 1: 在 brandMigration `Exec` 之后追加新的增量 ALTER**

找到（约 873 行附近）：

```go
	if _, err := DB.Exec(brandMigration); err != nil {
		log.Fatalf("增量迁移失败 (user_export_brand): %v", err)
	}

	log.Println("✅ 数据库迁移完成")
}
```

在 `if _, err := DB.Exec(brandMigration); err != nil { ... }` 块之后、`log.Println("✅ 数据库迁移完成")` 之前插入：

```go

	// 增量迁移：user_export_brand.logo_mode（icon = 方形 1:1，wordmark = 横版）
	brandLogoModeMigration := `
ALTER TABLE user_export_brand
    ADD COLUMN IF NOT EXISTS logo_mode VARCHAR(16) NOT NULL DEFAULT 'icon' CHECK (logo_mode IN ('icon', 'wordmark'));`
	if _, err := DB.Exec(brandLogoModeMigration); err != nil {
		log.Fatalf("增量迁移失败 (user_export_brand.logo_mode): %v", err)
	}
```

- [ ] **Step 2: 提交（不重启 DB，下一任务统一在 Task 7 验证）**

```bash
git -C /Users/liujiming/web/yuanju add backend/pkg/database/database.go
git -C /Users/liujiming/web/yuanju -c commit.gpgsign=false commit -m "feat(db): add user_export_brand.logo_mode column (icon|wordmark)"
```

---

## Task 4: Model 加 LogoMode 字段

**Files:**
- Modify: `backend/internal/model/user_brand.go`

- [ ] **Step 1: 替换整个 ExportBrand 结构体**

当前内容：

```go
package model

import "time"

// ExportBrand represents a user's customization of export PNG/PDF branding.
type ExportBrand struct {
	UserID        string    `json:"-"`
	Title         string    `json:"title"`
	FooterText    string    `json:"footer_text"`
	LogoPath      string    `json:"-"`
	LogoURL       string    `json:"logo_url"`
	WatermarkMode string    `json:"watermark_mode"`
	WatermarkText string    `json:"watermark_text"`
	UpdatedAt     time.Time `json:"-"`
}
```

替换为：

```go
package model

import "time"

// ExportBrand represents a user's customization of export PNG/PDF branding.
type ExportBrand struct {
	UserID        string    `json:"-"`
	Title         string    `json:"title"`
	FooterText    string    `json:"footer_text"`
	LogoPath      string    `json:"-"`
	LogoURL       string    `json:"logo_url"`
	LogoMode      string    `json:"logo_mode"`
	WatermarkMode string    `json:"watermark_mode"`
	WatermarkText string    `json:"watermark_text"`
	UpdatedAt     time.Time `json:"-"`
}
```

- [ ] **Step 2: 提交**

```bash
git -C /Users/liujiming/web/yuanju add backend/internal/model/user_brand.go
git -C /Users/liujiming/web/yuanju -c commit.gpgsign=false commit -m "feat(brand-model): add LogoMode field to ExportBrand"
```

---

## Task 5: Repository 扩 Get + Upsert

**Files:**
- Modify: `backend/internal/repository/user_brand_repository.go`

- [ ] **Step 1: 替换 GetExportBrand**

当前 `GetExportBrand` 函数（约 11 行）：

```go
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
```

替换为：

```go
func GetExportBrand(userID string) (model.ExportBrand, error) {
	var b model.ExportBrand
	b.UserID = userID
	err := database.DB.QueryRow(`
		SELECT title, footer_text, logo_path, logo_mode, watermark_mode, watermark_text, updated_at
		FROM user_export_brand WHERE user_id = $1`, userID).Scan(
		&b.Title, &b.FooterText, &b.LogoPath, &b.LogoMode, &b.WatermarkMode, &b.WatermarkText, &b.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		b.WatermarkMode = "none"
		b.LogoMode = "icon"
		return b, nil
	}
	return b, err
}
```

- [ ] **Step 2: 替换 UpsertExportBrandText**

当前函数（约 27 行）：

```go
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
```

替换为：

```go
func UpsertExportBrandText(userID, title, footerText, watermarkMode, watermarkText, logoMode string) error {
	_, err := database.DB.Exec(`
		INSERT INTO user_export_brand (user_id, title, footer_text, watermark_mode, watermark_text, logo_mode)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id) DO UPDATE SET
			title = EXCLUDED.title,
			footer_text = EXCLUDED.footer_text,
			watermark_mode = EXCLUDED.watermark_mode,
			watermark_text = EXCLUDED.watermark_text,
			logo_mode = EXCLUDED.logo_mode,
			updated_at = NOW()`,
		userID, title, footerText, watermarkMode, watermarkText, logoMode)
	return err
}
```

`UpdateExportBrandLogo` / `ClearExportBrandLogo` 都不需要改 — 它们只动 `logo_path`。

- [ ] **Step 3: 提交**

```bash
git -C /Users/liujiming/web/yuanju add backend/internal/repository/user_brand_repository.go
git -C /Users/liujiming/web/yuanju -c commit.gpgsign=false commit -m "feat(brand-repo): persist logo_mode in get/upsert"
```

---

## Task 6: 后端测试 + build 验证

**Files:** 无修改

- [ ] **Step 1: 跑 handler 单测**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./internal/handler/ -run TestValidateBrandUpdate -v 2>&1 | tail -25
```

Expected: 全部 PASS。包括我们 Task 1 加的 `TestValidateBrandUpdate_LogoMode` 6 个子用例。

- [ ] **Step 2: 跑全套后端测试**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./... 2>&1 | tail -20
```

Expected: 全部 PASS，无 FAIL。

- [ ] **Step 3: 后端编译**

```bash
cd /Users/liujiming/web/yuanju/backend && go build ./...
```

Expected: 干净，无错。

如果任何一步失败，回头检查是否漏改 `UpsertExportBrandText` 调用方（应该只有 `UpdateExportBrand` 这一处）。

---

## Task 7: 前端 API types 加 logo_mode

**Files:**
- Modify: `frontend/src/lib/api.ts`

- [ ] **Step 1: 替换 ExportBrand 与 BrandUpdateInput 接口**

当前（约 107–120 行）：

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
```

替换为：

```ts
// ======= Export Brand =======
export interface ExportBrand {
  title: string
  footer_text: string
  logo_url: string
  logo_mode: 'icon' | 'wordmark'
  watermark_mode: 'none' | 'bottom' | 'diagonal'
  watermark_text: string
}

export interface BrandUpdateInput {
  title: string
  footer_text: string
  logo_mode: 'icon' | 'wordmark'
  watermark_mode: 'none' | 'bottom' | 'diagonal'
  watermark_text: string
}
```

- [ ] **Step 2: 提交**

```bash
git -C /Users/liujiming/web/yuanju add frontend/src/lib/api.ts
git -C /Users/liujiming/web/yuanju -c commit.gpgsign=false commit -m "feat(brand-api): add logo_mode to ExportBrand and BrandUpdateInput"
```

---

## Task 8: RED — 前端 logo_mode plumbing 静态正则测试

**Files:**
- Modify: `frontend/tests/brand-settings.test.mjs`

- [ ] **Step 1: 在文件末尾追加 6 个断言**

打开 `frontend/tests/brand-settings.test.mjs`，在最后一行 `})` 之后追加：

```javascript

test('ExportBrand interface includes logo_mode union', async () => {
  const apiText = await readFile(resolve(REPO_ROOT, 'src/lib/api.ts'), 'utf8')
  assert.match(
    apiText,
    /logo_mode:\s*'icon'\s*\|\s*'wordmark'/,
    'src/lib/api.ts ExportBrand 接口缺少 logo_mode union',
  )
})

test('LogoCropModal accepts mode prop and branches on wordmark', async () => {
  const text = await readFile(
    resolve(REPO_ROOT, 'src/components/LogoCropModal.tsx'),
    'utf8',
  )
  assert.match(text, /mode:\s*'icon'\s*\|\s*'wordmark'/, 'mode prop 类型未声明')
  assert.match(text, /mode === 'wordmark'/, '缺少 wordmark 分支')
})

test('BrandSettingsPage references draft.logo_mode', async () => {
  const text = await readFile(
    resolve(REPO_ROOT, 'src/pages/BrandSettingsPage.tsx'),
    'utf8',
  )
  assert.match(text, /draft\.logo_mode/, 'BrandSettingsPage 未把 logo_mode 接入 draft')
})

test('ShareCard branches on brand.logo_mode === wordmark', async () => {
  const text = await readFile(
    resolve(REPO_ROOT, 'src/components/ShareCard.tsx'),
    'utf8',
  )
  assert.match(text, /logo_mode === 'wordmark'/, 'ShareCard 缺少 wordmark 分支')
})

test('PrintLayout branches on brand.logo_mode === wordmark', async () => {
  const text = await readFile(
    resolve(REPO_ROOT, 'src/components/PrintLayout.tsx'),
    'utf8',
  )
  assert.match(text, /logo_mode === 'wordmark'/, 'PrintLayout 缺少 wordmark 分支')
})

test('BrandPreviewCard branches on brand.logo_mode === wordmark', async () => {
  const text = await readFile(
    resolve(REPO_ROOT, 'src/components/BrandPreviewCard.tsx'),
    'utf8',
  )
  assert.match(text, /logo_mode === 'wordmark'/, 'BrandPreviewCard 缺少 wordmark 分支')
})
```

注意：不要重复定义 `__dirname` / `REPO_ROOT` 等顶部已有的常量；只追加 `test(...)` 块。

- [ ] **Step 2: 跑测试，确认 6 个全部 FAIL（RED）**

```bash
cd /Users/liujiming/web/yuanju/frontend && node --test tests/brand-settings.test.mjs 2>&1 | tail -40
```

Expected:
- `ExportBrand interface includes logo_mode union` → 应已 PASS（Task 7 已加）
- 另外 5 个 → FAIL（5 个文件都还没改）

如果 `ExportBrand` 那条没 PASS，回头检查 Task 7 是否正确执行。

- [ ] **Step 3: 提交 RED**

```bash
git -C /Users/liujiming/web/yuanju add frontend/tests/brand-settings.test.mjs
git -C /Users/liujiming/web/yuanju -c commit.gpgsign=false commit -m "test(brand-settings): wordmark mode plumbing assertions (5 RED)"
```

---

## Task 9: GREEN — LogoCropModal 加 mode prop + 预设 chip + 输出尺寸分支

**Files:**
- Modify: `frontend/src/components/LogoCropModal.tsx`
- Modify: `frontend/src/components/LogoCropModal.css`

- [ ] **Step 1: 替换 LogoCropModal.tsx 整文件**

```tsx
import { useCallback, useEffect, useState } from 'react'
import Cropper from 'react-easy-crop'
import './LogoCropModal.css'

interface Props {
  sourceDataUrl: string
  open: boolean
  mode: 'icon' | 'wordmark'
  onConfirm: (file: File) => void
  onCancel: () => void
}

interface PixelArea {
  x: number
  y: number
  width: number
  height: number
}

const ICON_OUTPUT_SIZE = 256
const WORDMARK_OUTPUT_HEIGHT = 128
const WORDMARK_ASPECT_PRESETS = [2, 3, 4] as const
const DEFAULT_WORDMARK_ASPECT = 3
const MAX_SOURCE_LONG_AXIS = 1600

export default function LogoCropModal({ sourceDataUrl, open, mode, onConfirm, onCancel }: Props) {
  const [imgUrl, setImgUrl] = useState('')
  const [crop, setCrop] = useState({ x: 0, y: 0 })
  const [zoom, setZoom] = useState(1)
  const [areaPx, setAreaPx] = useState<PixelArea | null>(null)
  const [processing, setProcessing] = useState(false)
  const [wordmarkAspect, setWordmarkAspect] = useState<number>(DEFAULT_WORDMARK_ASPECT)

  const cropperAspect = mode === 'icon' ? 1 : wordmarkAspect

  useEffect(() => {
    let cancelled = false
    if (!sourceDataUrl) {
      setImgUrl('')
      return
    }
    setCrop({ x: 0, y: 0 })
    setZoom(1)
    setAreaPx(null)
    downscaleIfLarge(sourceDataUrl)
      .then(url => { if (!cancelled) setImgUrl(url) })
      .catch(() => { if (!cancelled) setImgUrl(sourceDataUrl) })
    return () => { cancelled = true }
  }, [sourceDataUrl])

  // 切换比例 chip 时重置 crop 位置，让用户立刻看到新比例下的预览
  useEffect(() => {
    setCrop({ x: 0, y: 0 })
    setZoom(1)
    setAreaPx(null)
  }, [cropperAspect])

  const onCropComplete = useCallback((_: unknown, area: PixelArea) => {
    setAreaPx(area)
  }, [])

  if (!open) return null

  async function handleConfirm() {
    if (!areaPx || !imgUrl) return
    setProcessing(true)
    try {
      let outW: number, outH: number
      if (mode === 'icon') {
        outW = ICON_OUTPUT_SIZE
        outH = ICON_OUTPUT_SIZE
      } else {
        outH = WORDMARK_OUTPUT_HEIGHT
        outW = Math.round(outH * wordmarkAspect)
      }
      const blob = await cropToBlob(imgUrl, areaPx, outW, outH)
      const file = new File([blob], 'logo.png', { type: 'image/png' })
      onConfirm(file)
    } catch (err) {
      console.error('crop failed', err)
    } finally {
      setProcessing(false)
    }
  }

  return (
    <div className="logo-crop-overlay" onClick={onCancel} role="dialog" aria-modal="true">
      <div className="logo-crop-modal" onClick={e => e.stopPropagation()}>
        <h3 className="logo-crop-title">
          {mode === 'icon' ? '调整 logo 裁剪区域' : '调整商标裁剪区域'}
        </h3>

        {mode === 'wordmark' && (
          <div className="logo-crop-aspect-chips" role="radiogroup" aria-label="商标比例">
            {WORDMARK_ASPECT_PRESETS.map(a => (
              <button
                key={a}
                type="button"
                role="radio"
                aria-checked={wordmarkAspect === a}
                className={`logo-crop-aspect-chip${wordmarkAspect === a ? ' is-active' : ''}`}
                onClick={() => setWordmarkAspect(a)}
              >
                {a}:1
              </button>
            ))}
          </div>
        )}

        <div className="logo-crop-canvas-area">
          {imgUrl ? (
            <Cropper
              image={imgUrl}
              crop={crop}
              zoom={zoom}
              aspect={cropperAspect}
              cropShape="rect"
              showGrid
              onCropChange={setCrop}
              onZoomChange={setZoom}
              onCropComplete={onCropComplete}
            />
          ) : (
            <div className="logo-crop-loading">加载中...</div>
          )}
        </div>

        <div className="logo-crop-zoom-row">
          <span>缩放</span>
          <input
            type="range"
            min={1}
            max={3}
            step={0.01}
            value={zoom}
            onChange={e => setZoom(Number(e.target.value))}
          />
        </div>

        <small className="logo-crop-note">
          {mode === 'icon'
            ? '动图（GIF / 动 WebP）将仅保留第一帧。输出 256×256 PNG。'
            : '动图（GIF / 动 WebP）将仅保留第一帧。输出高 128 PNG，宽随比例（最多 512）。'}
        </small>

        <div className="logo-crop-actions">
          <button type="button" className="btn btn-ghost" onClick={onCancel} disabled={processing}>
            取消
          </button>
          <button
            type="button"
            className="btn btn-primary"
            onClick={handleConfirm}
            disabled={processing || !areaPx || !imgUrl}
          >
            {processing ? '处理中...' : '确认裁剪'}
          </button>
        </div>
      </div>
    </div>
  )
}

function loadImage(src: string): Promise<HTMLImageElement> {
  return new Promise((resolve, reject) => {
    const img = new Image()
    img.onload = () => resolve(img)
    img.onerror = () => reject(new Error('图片加载失败'))
    img.src = src
  })
}

async function downscaleIfLarge(dataUrl: string): Promise<string> {
  const img = await loadImage(dataUrl)
  const longest = Math.max(img.naturalWidth, img.naturalHeight)
  if (longest <= MAX_SOURCE_LONG_AXIS) return dataUrl
  const scale = MAX_SOURCE_LONG_AXIS / longest
  const canvas = document.createElement('canvas')
  canvas.width = Math.round(img.naturalWidth * scale)
  canvas.height = Math.round(img.naturalHeight * scale)
  const ctx = canvas.getContext('2d')
  if (!ctx) return dataUrl
  ctx.drawImage(img, 0, 0, canvas.width, canvas.height)
  return canvas.toDataURL('image/png')
}

async function cropToBlob(src: string, area: PixelArea, outW: number, outH: number): Promise<Blob> {
  const img = await loadImage(src)
  const canvas = document.createElement('canvas')
  canvas.width = outW
  canvas.height = outH
  const ctx = canvas.getContext('2d')
  if (!ctx) throw new Error('canvas 2d 不可用')
  ctx.drawImage(
    img,
    area.x, area.y, area.width, area.height,
    0, 0, outW, outH,
  )
  return new Promise((resolve, reject) => {
    canvas.toBlob(
      blob => blob ? resolve(blob) : reject(new Error('toBlob 失败')),
      'image/png',
    )
  })
}
```

- [ ] **Step 2: 在 LogoCropModal.css 末尾追加 chip 行样式**

```css

.logo-crop-aspect-chips {
  display: flex;
  gap: 8px;
  margin-bottom: 10px;
  flex-wrap: wrap;
}

.logo-crop-aspect-chip {
  padding: 4px 12px;
  font-size: 12px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-default);
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
  transition: border-color 0.18s, color 0.18s, background 0.18s;
}

.logo-crop-aspect-chip:hover {
  border-color: var(--border-accent);
  color: var(--text-primary);
}

.logo-crop-aspect-chip.is-active {
  border-color: var(--text-accent);
  color: var(--text-accent);
  background: rgba(201, 168, 76, 0.10);
}
```

- [ ] **Step 3: 跑测试，确认 LogoCropModal 那条 PASS**

```bash
cd /Users/liujiming/web/yuanju/frontend && node --test tests/brand-settings.test.mjs 2>&1 | grep -E '(✔|✖|LogoCropModal)' | head
```

Expected: `LogoCropModal accepts mode prop and branches on wordmark` 现在 ✔ PASS。其它 4 个 wordmark 测试仍 FAIL。

- [ ] **Step 4: 提交**

```bash
git -C /Users/liujiming/web/yuanju add frontend/src/components/LogoCropModal.tsx frontend/src/components/LogoCropModal.css
git -C /Users/liujiming/web/yuanju -c commit.gpgsign=false commit -m "feat(logo-crop-modal): icon/wordmark mode with aspect chips (2:1/3:1/4:1)"
```

---

## Task 10: BrandSettingsPage 加模式 radio + draft.logo_mode + 警告 + 传 mode 到 modal

**Files:**
- Modify: `frontend/src/pages/BrandSettingsPage.tsx`

- [ ] **Step 1: 把 DEFAULT_BRAND 中加 logo_mode**

找到（约 11 行附近）：

```tsx
const DEFAULT_BRAND: ExportBrand = {
  title: '',
  footer_text: '',
  logo_url: '',
  watermark_mode: 'none',
  watermark_text: '',
}
```

替换为：

```tsx
const DEFAULT_BRAND: ExportBrand = {
  title: '',
  footer_text: '',
  logo_url: '',
  logo_mode: 'icon',
  watermark_mode: 'none',
  watermark_text: '',
}
```

- [ ] **Step 2: 在 BrandSettingsPage 组件函数体内、`dirty` 计算上方添加 logo aspect 检测**

找到（约 64 行附近，`if (authLoading || loading) {...}` 之后、`const dirty = ...` 之前）插入：

```tsx
  const [logoAspect, setLogoAspect] = useState<number | null>(null)

  useEffect(() => {
    if (!serverState.logo_url) {
      setLogoAspect(null)
      return
    }
    const img = new Image()
    img.onload = () => setLogoAspect(img.naturalWidth / img.naturalHeight)
    img.onerror = () => setLogoAspect(null)
    img.src = serverState.logo_url
  }, [serverState.logo_url])

  const aspectMismatch = logoAspect != null && (
    (draft.logo_mode === 'icon' && Math.abs(logoAspect - 1) > 0.1) ||
    (draft.logo_mode === 'wordmark' && logoAspect < 1.5)
  )

```

- [ ] **Step 3: 把 dirty 计算扩展进 logo_mode**

找到（约 66 行附近）：

```tsx
  const dirty =
    draft.title !== serverState.title ||
    draft.footer_text !== serverState.footer_text ||
    draft.watermark_mode !== serverState.watermark_mode ||
    draft.watermark_text !== serverState.watermark_text
```

替换为：

```tsx
  const dirty =
    draft.title !== serverState.title ||
    draft.footer_text !== serverState.footer_text ||
    draft.logo_mode !== serverState.logo_mode ||
    draft.watermark_mode !== serverState.watermark_mode ||
    draft.watermark_text !== serverState.watermark_text
```

- [ ] **Step 4: onSave 函数把 logo_mode 一起送出去**

找到（约 72 行附近）：

```tsx
    try {
      const r = await brandAPI.update({
        title: draft.title,
        footer_text: draft.footer_text,
        watermark_mode: draft.watermark_mode,
        watermark_text: draft.watermark_text,
      })
```

替换为：

```tsx
    try {
      const r = await brandAPI.update({
        title: draft.title,
        footer_text: draft.footer_text,
        logo_mode: draft.logo_mode,
        watermark_mode: draft.watermark_mode,
        watermark_text: draft.watermark_text,
      })
```

注意：onReset 里也有一个 update 调用，也要改：

```tsx
      await brandAPI.update({ title: '', footer_text: '', watermark_mode: 'none', watermark_text: '' })
```

替换为：

```tsx
      await brandAPI.update({ title: '', footer_text: '', logo_mode: 'icon', watermark_mode: 'none', watermark_text: '' })
```

- [ ] **Step 5: 在「顶部品牌」section 头部加 mode radio**

找到（约 172 行附近）：

```tsx
      <section className="brand-section">
        <h2>顶部品牌</h2>
        <label className="brand-field">
          <span>品牌标题</span>
```

在 `<h2>顶部品牌</h2>` 之后、`<label className="brand-field">` 之前插入：

```tsx
        <div className="brand-radio-row" role="radiogroup" aria-label="Logo 模式">
          <label>
            <input
              type="radio"
              name="logo_mode"
              value="icon"
              checked={draft.logo_mode === 'icon'}
              onChange={() => setDraft(d => ({ ...d, logo_mode: 'icon' }))}
            />
            图标（方形 1:1）
          </label>
          <label>
            <input
              type="radio"
              name="logo_mode"
              value="wordmark"
              checked={draft.logo_mode === 'wordmark'}
              onChange={() => setDraft(d => ({ ...d, logo_mode: 'wordmark' }))}
            />
            商标（横版）
          </label>
        </div>
        <small className="brand-mode-hint">
          图标：方形 logo + 文字标题。商标：横版 logo 取代文字标题。
        </small>

```

- [ ] **Step 6: 在 logo 预览行下加 aspectMismatch 警告**

找到（约 213 行附近，`</div>` 关闭 `.brand-logo-row` 的位置）。`.brand-logo-row` 整块结束后（同一 section 内）插入：

```tsx
        {aspectMismatch && (
          <div className="brand-warning">建议重新上传符合当前模式的 logo</div>
        )}
```

具体定位：在 `.brand-logo-row` 关闭 `</div>` 之后、`</section>` 之前。

- [ ] **Step 7: 把 mode 透传给 LogoCropModal**

找到文件末尾附近（约 289 行）：

```tsx
      <LogoCropModal
        sourceDataUrl={cropSourceUrl ?? ''}
        open={!!cropSourceUrl}
        onConfirm={handleCropConfirm}
        onCancel={() => setCropSourceUrl(null)}
      />
```

替换为：

```tsx
      <LogoCropModal
        sourceDataUrl={cropSourceUrl ?? ''}
        open={!!cropSourceUrl}
        mode={draft.logo_mode}
        onConfirm={handleCropConfirm}
        onCancel={() => setCropSourceUrl(null)}
      />
```

- [ ] **Step 8: 跑测试，确认 BrandSettingsPage 那条 PASS**

```bash
cd /Users/liujiming/web/yuanju/frontend && node --test tests/brand-settings.test.mjs 2>&1 | grep -E '(✔|✖|BrandSettingsPage)' | head
```

Expected: `BrandSettingsPage references draft.logo_mode` 现在 ✔。

- [ ] **Step 9: 提交**

```bash
git -C /Users/liujiming/web/yuanju add frontend/src/pages/BrandSettingsPage.tsx
git -C /Users/liujiming/web/yuanju -c commit.gpgsign=false commit -m "feat(brand-settings): add logo mode radio, wire draft, aspect mismatch warning"
```

---

## Task 11: BrandSettingsPage CSS 加 .brand-warning + .brand-mode-hint

**Files:**
- Modify: `frontend/src/pages/BrandSettingsPage.css`

- [ ] **Step 1: 在文件末尾追加**

```css

.brand-warning {
  background: rgba(212, 184, 150, 0.10);
  border: 1px solid var(--border-accent);
  color: var(--text-accent);
  padding: 8px 14px;
  border-radius: var(--radius-sm);
  margin-top: 8px;
  font-size: 12px;
}

.brand-mode-hint {
  display: block;
  font-size: 11px;
  color: var(--text-muted);
  margin: -4px 0 12px;
}
```

- [ ] **Step 2: 提交**

```bash
git -C /Users/liujiming/web/yuanju add frontend/src/pages/BrandSettingsPage.css
git -C /Users/liujiming/web/yuanju -c commit.gpgsign=false commit -m "style(brand-settings): add .brand-warning and .brand-mode-hint"
```

---

## Task 12: ShareCard 按 mode 分支

**Files:**
- Modify: `frontend/src/components/ShareCard.tsx`

- [ ] **Step 1: 在 ShareCard 函数体内、`return` 之前加 isWordmark 计算**

找到（约 111 行附近）：

```tsx
  const resolvedTitle = brand?.title || '缘 聚 命 理'
  const resolvedFooter = resolveFooter(brand, 'yuanju.com')
  const showDiagonalMark = showDiagonalWatermark(brand)
```

在它之后追加一行：

```tsx
  const isWordmark = brand?.logo_mode === 'wordmark' && !!brand?.logo_url
```

- [ ] **Step 2: 顶部品牌栏整段替换**

找到（约 128–169 行附近，`<div style={{ background: 'linear-gradient(135deg, #2d1f14 ...`）。这整个顶部 div（含 logo img、resolvedTitle div、date·gender div）替换为：

```tsx
      {/* ┌ 顶部品牌栏 ── */}
      <div style={{
        background: 'linear-gradient(135deg, #2d1f14 0%, #4a3020 50%, #3a2416 100%)',
        padding: '20px 24px 18px',
        position: 'relative',
      }}>
        {isWordmark ? (
          <img
            src={brand!.logo_url}
            alt=""
            crossOrigin="anonymous"
            style={{
              display: 'block',
              margin: '0 auto 6px',
              maxHeight: 48,
              maxWidth: 320,
              objectFit: 'contain',
            }}
          />
        ) : (
          <>
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
          </>
        )}
        <div style={{
          color: '#c4a06a',
          fontSize: 12,
          letterSpacing: 1,
          textAlign: 'center',
          fontFamily: '"Noto Sans SC", sans-serif',
        }}>
          {birthYear}年{birthMonth}月{birthDay}日&nbsp;{birthHour}时 · {gender === 'male' ? '男命' : '女命'}
        </div>
      </div>
```

- [ ] **Step 3: 跑测试，确认 ShareCard 那条 PASS**

```bash
cd /Users/liujiming/web/yuanju/frontend && node --test tests/brand-settings.test.mjs 2>&1 | grep -E '(✔|✖|ShareCard)' | head
```

Expected: `ShareCard branches on brand.logo_mode === wordmark` 现在 ✔。

- [ ] **Step 4: 提交**

```bash
git -C /Users/liujiming/web/yuanju add frontend/src/components/ShareCard.tsx
git -C /Users/liujiming/web/yuanju -c commit.gpgsign=false commit -m "feat(share-card): branch top brand bar by logo_mode"
```

---

## Task 13: PrintLayout 按 mode 分支（per-page header + cover banner）

**Files:**
- Modify: `frontend/src/components/PrintLayout.tsx`
- Modify: `frontend/src/pages/ResultPage.css`

- [ ] **Step 1: 在 PrintLayout 中加 isWordmark**

找到（约 119–122 行）：

```tsx
  const customTitle = brand?.title?.trim() || ''
  const coverTitle = customTitle || '命 理 命 书'
  const headerTitle = customTitle || '命 理 命 书'
  const resolvedFooter = resolveFooter(brand, '缘 聚 命 理')
  const showDiagonalMark = showDiagonalWatermark(brand)
```

在它之后追加：

```tsx
  const isWordmark = brand?.logo_mode === 'wordmark' && !!brand?.logo_url
```

- [ ] **Step 2: 替换每页页眉**

找到（约 175–195 行）：

```tsx
              <div className="print-page-header">
                <span className="print-page-header-left">
                  {brand?.logo_url && (
                    <img
                      src={brand.logo_url}
                      alt=""
                      crossOrigin="anonymous"
                      className="print-page-header-logo"
                    />
                  )}
                  <span className="print-page-header-center">{headerTitle}</span>
                </span>
                <span className="print-page-header-info">
                  {birthYear}年{birthMonth}月{birthDay}日&nbsp;·&nbsp;{gender === 'male' ? '男命' : '女命'}
                </span>
              </div>
```

替换为：

```tsx
              <div className="print-page-header">
                <span className="print-page-header-left">
                  {isWordmark ? (
                    <img
                      src={brand!.logo_url}
                      alt=""
                      crossOrigin="anonymous"
                      className="print-page-header-wordmark"
                    />
                  ) : (
                    <>
                      {brand?.logo_url && (
                        <img
                          src={brand.logo_url}
                          alt=""
                          crossOrigin="anonymous"
                          className="print-page-header-logo"
                        />
                      )}
                      <span className="print-page-header-center">{headerTitle}</span>
                    </>
                  )}
                </span>
                <span className="print-page-header-info">
                  {birthYear}年{birthMonth}月{birthDay}日&nbsp;·&nbsp;{gender === 'male' ? '男命' : '女命'}
                </span>
              </div>
```

- [ ] **Step 3: 替换封面 banner**

找到（约 214–242 行附近）。当前从 `{/* ── 封面头部 ── */}` 起的整块（带 borderBottom 渐变线、kicker、coverTitle、生日行）。我们只改"kicker + coverTitle"那两段。

找到（约 220–240 行）：

```tsx
        {!customTitle && (
          <div style={{ fontSize: 9, letterSpacing: 6, color: '#999', marginBottom: 6 }}>
            YUAN JU MING LI
          </div>
        )}
        <div
          style={{
            fontSize: 28,
            fontWeight: 900,
            letterSpacing: customTitle ? (customTitle.length > 6 ? 2 : 6) : 10,
            color: darkBrown,
            marginBottom: 8,
          }}
        >
          {coverTitle}
        </div>
```

替换为：

```tsx
        {isWordmark ? (
          <img
            src={brand!.logo_url}
            alt=""
            crossOrigin="anonymous"
            style={{
              display: 'block',
              margin: '0 auto 8px',
              maxHeight: '40mm',
              maxWidth: '120mm',
              objectFit: 'contain',
            }}
          />
        ) : (
          <>
            {!customTitle && (
              <div style={{ fontSize: 9, letterSpacing: 6, color: '#999', marginBottom: 6 }}>
                YUAN JU MING LI
              </div>
            )}
            <div
              style={{
                fontSize: 28,
                fontWeight: 900,
                letterSpacing: customTitle ? (customTitle.length > 6 ? 2 : 6) : 10,
                color: darkBrown,
                marginBottom: 8,
              }}
            >
              {coverTitle}
            </div>
          </>
        )}
```

- [ ] **Step 4: 加 .print-page-header-wordmark CSS**

在 `frontend/src/pages/ResultPage.css` 中找到 `.print-page-header-logo` 块（约 1988 行附近，已在前一个改动里加过）。在它的下一个 `}` 后追加：

```css
  .print-page-header-wordmark {
    height: 6mm;
    width: auto;
    max-width: 80mm;
    object-fit: contain;
    display: inline-block;
  }
```

注意：这一段必须在 `@media print { ... }` 块内（与 `.print-page-header-logo` 同级），不然每页样式不会生效。

- [ ] **Step 5: 跑测试，确认 PrintLayout 那条 PASS**

```bash
cd /Users/liujiming/web/yuanju/frontend && node --test tests/brand-settings.test.mjs 2>&1 | grep -E '(✔|✖|PrintLayout)' | head
```

Expected: `PrintLayout branches on brand.logo_mode === wordmark` 现在 ✔。

- [ ] **Step 6: 提交**

```bash
git -C /Users/liujiming/web/yuanju add frontend/src/components/PrintLayout.tsx frontend/src/pages/ResultPage.css
git -C /Users/liujiming/web/yuanju -c commit.gpgsign=false commit -m "feat(print-layout): branch per-page header and cover by logo_mode"
```

---

## Task 14: BrandPreviewCard 按 mode 分支

**Files:**
- Modify: `frontend/src/components/BrandPreviewCard.tsx`

- [ ] **Step 1: 在 BrandPreviewCard 函数体内、return 之前加 isWordmark**

找到（约 13–16 行）：

```tsx
export default function BrandPreviewCard({ brand }: Props) {
  const title = brand.title || '缘 聚 命 理'
  const footer = resolveFooter(brand, 'yuanju.com')
  const showDiagonal = showDiagonalWatermark(brand)
```

在 `showDiagonal` 那行之后追加：

```tsx
  const isWordmark = brand.logo_mode === 'wordmark' && !!brand.logo_url
```

- [ ] **Step 2: 替换 Header 块**

找到（约 28–46 行）：

```tsx
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
```

替换为：

```tsx
      {/* Header */}
      <div style={{
        position: 'relative',
        background: 'linear-gradient(135deg, #2d1f14 0%, #4a3020 50%, #3a2416 100%)',
        padding: '16px 20px 14px',
        textAlign: 'center',
      }}>
        {isWordmark ? (
          <img
            src={brand.logo_url}
            alt=""
            style={{
              display: 'block',
              margin: '0 auto',
              maxHeight: 32,
              maxWidth: 240,
              objectFit: 'contain',
            }}
          />
        ) : (
          <>
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
          </>
        )}
        <div style={{ color: '#c4a06a', fontSize: 10, marginTop: 4 }}>预览（占位）</div>
      </div>
```

- [ ] **Step 3: 跑测试，确认 BrandPreviewCard 那条 PASS**

```bash
cd /Users/liujiming/web/yuanju/frontend && node --test tests/brand-settings.test.mjs 2>&1 | grep -E '(✔|✖|BrandPreviewCard)' | head
```

Expected: `BrandPreviewCard branches on brand.logo_mode === wordmark` 现在 ✔。

- [ ] **Step 4: 提交**

```bash
git -C /Users/liujiming/web/yuanju add frontend/src/components/BrandPreviewCard.tsx
git -C /Users/liujiming/web/yuanju -c commit.gpgsign=false commit -m "feat(brand-preview-card): branch header by logo_mode"
```

---

## Task 15: 全套 lint / build / tests 验证

**Files:** 无修改

- [ ] **Step 1: 后端测试 + build**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./... 2>&1 | tail -15
cd /Users/liujiming/web/yuanju/backend && go build ./...
```

Expected: 测试全 PASS；build 干净。

- [ ] **Step 2: 前端新静态测试 + 既有静态测试 + 深色化测试**

```bash
cd /Users/liujiming/web/yuanju/frontend && node --test tests/brand-settings.test.mjs 2>&1 | tail -20
cd /Users/liujiming/web/yuanju/frontend && node --test tests/brand-settings-dark-theme.test.mjs 2>&1 | tail -15
```

Expected:
- `brand-settings.test.mjs`：6 原断言 + 6 新断言 = 12 PASS
- `brand-settings-dark-theme.test.mjs`：4/4 PASS（回归不破）

- [ ] **Step 3: 前端 lint**

```bash
cd /Users/liujiming/web/yuanju/frontend && npm run lint 2>&1 | tail -15
```

Expected: 干净，或仅有此前的 PrintLayout `no-irregular-whitespace` 2 条**预存在**错误（与本 PR 无关）。

- [ ] **Step 4: 前端 build**

```bash
cd /Users/liujiming/web/yuanju/frontend && npm run build 2>&1 | tail -15
```

Expected: TypeScript 通过；Vite build 产物输出。

如果 build 报 TypeScript 错误（典型：`Type ... is missing the following properties from type ExportBrand: logo_mode`），回头看是否有其他地方构造 `ExportBrand`/`BrandUpdateInput` 没有更新（grep `ExportBrand` 和 `BrandUpdateInput` 全工程）。

- [ ] **Step 5: （仅验证步骤都通过的话）什么也不提交，确认工作树干净**

```bash
git -C /Users/liujiming/web/yuanju status
```

Expected: `nothing to commit, working tree clean`，并且本 PR 共 **14 个新 commit**（Task 1-14 各 1 个）。

---

## Task 16: 手动视觉验收（交给用户）

**Files:** 无修改

- [ ] **Step 1: 重建容器**

```bash
docker compose -f /Users/liujiming/web/yuanju/docker-compose.yml up -d --build backend frontend
```

Expected: backend 容器会跑增量迁移；frontend 容器重建后 serve 新代码。

- [ ] **Step 2: 后端数据库验证**

```bash
docker compose -f /Users/liujiming/web/yuanju/docker-compose.yml exec postgres psql -U postgres -d yuanju -c "\d user_export_brand" | head -20
```

Expected: 输出含 `logo_mode | character varying(16) | not null` 一行，且 CHECK 约束 `logo_mode = ANY (ARRAY['icon'::..., 'wordmark'::...])`。

- [ ] **Step 3: 浏览器逐项核对**

打开 `http://localhost:3000/settings/brand`（需登录）。依次核对：

1. **Logo 模式 radio** 出现在「顶部品牌」section 顶端，默认选中"图标"；切到"商标"helper text 顺读
2. **图标模式上传**：点上传 → crop modal 标题 "调整 logo 裁剪区域" → 1:1 锁定 → 确认 → 256×256 PNG 上传成功
3. **切到商标模式**：现有方形 logo 旁出现金色"建议重新上传符合当前模式的 logo"警告
4. **商标模式上传**：点更换 → crop modal 标题 "调整商标裁剪区域" → 顶部出现 `[2:1] [3:1] [4:1]` chip 行，默认 3:1 高亮 → 切 2:1 → 切 4:1 → 比例随之变 → 确认 → ~384×128 PNG 上传
5. **商标 + PDF 导出**：打开任一 chart 的 ResultPage → 导出 PDF → 第 1 页**封面没有"命 理 命 书"文字，是横版 logo 居中** → 翻第 2、3 页，**每页左上角是 wordmark logo + date·gender**，没有任何文字标题
6. **商标 + PNG 导出**：点保存分享图 → 输出 PNG 的顶部深色品牌栏是 wordmark logo 居中，下方是 date·gender，无文字标题
7. **回到图标模式**：选回 icon → 立刻看到现有 wordmark logo 还在，但旁边出现警告（因为现在是 icon 模式但 logo 是宽的）→ 重新上传方形 → 警告消失
8. **图标 + PDF**：仍是 `[□ icon] <brand.title || '命 理 命 书'> ... date·gender`（已是之前修好的样式）

如果任何一项不符，回头检查对应 Task 的 commit。

---

## Self-Review Notes

**Spec 覆盖核对：**

| Spec Requirement | 对应任务 |
|------------------|----------|
| R1 Brand record persists logo mode | Task 3（schema）+ Task 4（model）+ Task 5（repo） |
| R1 Scenario: New users default to icon | Task 5 GetExportBrand `b.LogoMode = "icon"` 在 sql.ErrNoRows 分支 |
| R1 Scenario: User toggles to wordmark and saves | Task 2 UpdateExportBrand + Task 5 Upsert + Task 10 onSave |
| R1 Scenario: Invalid mode rejected | Task 1 RED + Task 2 validation switch |
| R2 Icon mode preserves 1:1 | Task 9 `cropperAspect = mode === 'icon' ? 1 : wordmarkAspect`，OUTPUT_SIZE 256 |
| R2 Scenario: Pan and zoom available | Task 9 Cropper 仍支持（zoom slider + drag） |
| R3 Wordmark mode aspect — **务实偏离为预设档** | Task 9 chip 行 `WORDMARK_ASPECT_PRESETS = [2, 3, 4]` |
| R3 Scenario: 2.5:1 wordmark | **不再适用**（用户选预设而非任意拖拽，文档中已声明） |
| R3 Scenario: drag past upper / lower cap | **不再适用**（预设档无越界） |
| R4 Wordmark replaces text titles | Task 12 ShareCard + Task 13 PrintLayout + Task 14 BrandPreviewCard |
| R4 Scenario: PDF per-page header | Task 13 Step 2 |
| R4 Scenario: PDF cover banner | Task 13 Step 3 |
| R4 Scenario: PNG share card | Task 12 Step 2 |
| R4 Scenario: brand.title preserved but unused | Task 12-14 各组件忽略 title 当 isWordmark；Task 10 form 不动 title input |
| R5 Wordmark with no logo falls back to text | Task 12-14 各组件 `isWordmark = ... && !!brand.logo_url`，空 logo → 走 icon 分支 |
| R6 Switching mode preserves logo + warns | Task 10 Step 2-3（aspectMismatch + 不清空逻辑） |
| R6 Scenario: icon→wordmark with square logo | Task 10 Step 2 `logoAspect ≈ 1 && draft=wordmark` 触发警告 |
| R6 Scenario: wordmark→icon with wide logo | Task 10 Step 2 `logoAspect ≥ 1.5 && draft=icon` 触发警告 |

**Placeholder 扫描：** 已扫描全文。所有任务都附完整代码 / 命令 / 期望输出。无 "TBD" / "TODO" / "similar to Task N" / 未定义类型。

**Type / 命名一致性：**
- `LogoMode` Go 字段（Task 2/4/5）↔ `logo_mode` JSON tag（Task 2/4）↔ `logo_mode` TS interface 键（Task 7）↔ `draft.logo_mode` / `brand.logo_mode` 引用（Task 10/12/13/14） ✓
- `UpsertExportBrandText` 签名（Task 5）含 6 个参数，与 handler 调用（Task 2 Step 4）参数列表一一对应 ✓
- `Props` interface 中 `mode: 'icon' | 'wordmark'`（Task 9）↔ `<LogoCropModal mode={draft.logo_mode}>`（Task 10 Step 7） ✓
- `isWordmark` 变量名在 ShareCard / PrintLayout / BrandPreviewCard 三处保持一致 ✓
- CSS class `.print-page-header-wordmark`（Task 13 Step 4）↔ JSX className（Task 13 Step 2） ✓
- 静态测试断言（Task 8）↔ 实际实现的字符串模式 ✓
