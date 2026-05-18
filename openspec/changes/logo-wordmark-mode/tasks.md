## 1. Backend — schema

- [ ] 1.1 Add `logo_mode TEXT NOT NULL DEFAULT 'icon' CHECK (logo_mode IN ('icon', 'wordmark'))` to `user_export_brand` in `backend/pkg/database/database.go`, gated by `ALTER TABLE ... ADD COLUMN IF NOT EXISTS` to keep the migration idempotent.
- [ ] 1.2 Restart the dev DB and verify `psql ... -c "\d user_export_brand"` shows the new column with the default.

## 2. Backend — model + repository

- [ ] 2.1 In `backend/internal/model/user_brand.go`, add `LogoMode string` to the `ExportBrand` struct with `json:"logo_mode"` tag.
- [ ] 2.2 In `backend/internal/repository/user_brand_repository.go`:
  - Extend `Get` to scan `logo_mode`.
  - Extend `Upsert` and `UpdateLogo` to write `logo_mode`.
- [ ] 2.3 Add a Go unit test (`user_brand_repository_test.go` if not present, otherwise extend existing) that calls Upsert with `LogoMode='wordmark'` and verifies a subsequent Get returns it.

## 3. Backend — handler

- [ ] 3.1 In `backend/internal/handler/user_brand_handler.go`:
  - Extend `validateBrandUpdate` to validate `LogoMode in {'icon', 'wordmark', ''}` (empty string → treat as `icon` for backward compat).
  - Extend `buildBrandResponse` to emit `LogoMode`.
  - Extend `UpdateExportBrand` to accept and persist the field.
- [ ] 3.2 Add unit tests for `validateBrandUpdate` covering: valid `icon`, valid `wordmark`, empty (allowed), invalid `'square'` (rejected with 400).

## 4. Frontend — API types

- [ ] 4.1 In `frontend/src/lib/api.ts`, extend the `ExportBrand` interface with `logo_mode: 'icon' | 'wordmark'`.

## 5. Frontend — LogoCropModal mode support

- [ ] 5.1 Extend the `Props` interface of `LogoCropModal` with `mode: 'icon' | 'wordmark'`.
- [ ] 5.2 In `LogoCropModal.tsx`:
  - When `mode === 'icon'`: pass `aspect={1}` to `<Cropper>`, output `OUTPUT_SIZE = 256` (square).
  - When `mode === 'wordmark'`: pass `aspect={3}` as the *starting* aspect, do NOT lock; in `onCropComplete`, if `area.width / area.height < 1.5` or `> 6`, replace area with the clamped value before storing.
  - In `cropToBlob`: for `wordmark` mode, set `canvas.height = 128`, `canvas.width = Math.round(128 * (areaPx.width / areaPx.height))` (clamped via `Math.min(canvas.width, 768)` then re-clamped to `Math.max(canvas.width, 192)`).
  - Update the small `<small className="logo-crop-note">` line to reflect the mode: `mode === 'icon' ? '动图（GIF / 动 WebP）将仅保留第一帧。输出 256×256 PNG。' : '动图（GIF / 动 WebP）将仅保留第一帧。横版裁剪比例 1.5:1 ~ 6:1，输出高 128 PNG。'`
- [ ] 5.3 Static-regex test in `frontend/tests/brand-settings.test.mjs` (extend, do not create new) asserting the new `mode` prop is destructured.

## 6. Frontend — BrandSettings UI

- [ ] 6.1 In `BrandSettingsPage.tsx`:
  - Add `logo_mode: 'icon'` to `DEFAULT_BRAND`.
  - Render a radio control labeled `Logo 模式` inside the 顶部品牌 section, above the existing `品牌标题` input.
  - The radio's `value` reflects `draft.logo_mode`; selecting toggles it.
  - Pass `mode={draft.logo_mode}` to `<LogoCropModal>`.
  - Helper note under the radio: `图标模式：方形 logo + 文字标题。商标模式：横版 logo 取代文字标题。`
  - Below the radio, when `dirty` includes the mode switch AND `serverState.logo_url`, show the soft warning: `<div className="brand-warning">建议重新上传符合当前模式的 logo</div>`.
- [ ] 6.2 In `BrandSettingsPage.css`, add `.brand-warning` styling using dark-theme tokens (background `rgba(212,184,150,0.10)`, border `var(--border-accent)`, color `var(--text-accent)`).
- [ ] 6.3 Update `dirty` derivation in `BrandSettingsPage.tsx` to include `draft.logo_mode !== serverState.logo_mode`.

## 7. Frontend — PrintLayout branch

- [ ] 7.1 In `PrintLayout.tsx`:
  - Compute `isWordmark = brand?.logo_mode === 'wordmark' && !!brand?.logo_url`.
  - Per-page header: when `isWordmark`, render the logo image at `height: 6mm; width: auto; max-width: 80mm; object-fit: contain` and OMIT the text title span. Otherwise current icon+title layout.
  - Cover banner: when `isWordmark`, replace the existing `<div style={{ fontSize: 28, ... }}>` text block with `<img src={brand.logo_url} style={{ maxHeight: '40mm', maxWidth: '120mm', objectFit: 'contain' }} />`. Hide the small kicker too.
  - Otherwise current cover layout (which already honors `brand.title`).
- [ ] 7.2 In `ResultPage.css`, add print-only CSS for `.print-page-header-wordmark` (used as the wrapper class) with `display: inline-flex; align-items: center;`.

## 8. Frontend — ShareCard branch

- [ ] 8.1 In `ShareCard.tsx`:
  - Compute `isWordmark = brand?.logo_mode === 'wordmark' && !!brand?.logo_url`.
  - Top brand bar: when `isWordmark`, replace the entire centered title block with a centered `<img>` set to `maxHeight: 48px, maxWidth: 320px, objectFit: 'contain'`. Drop the small kicker / pinyin. Hide the absolute-positioned 40×40 left logo (no longer needed; the wordmark IS the brand).
  - Otherwise current layout (which already shows `resolvedTitle`).

## 9. Frontend — BrandPreviewCard branch

- [ ] 9.1 In `BrandPreviewCard.tsx`, mirror ShareCard's logic: when `brand.logo_mode === 'wordmark'` and `logo_url` set, render the wordmark in the title slot instead of text. Use a smaller scale appropriate for the preview card.

## 10. Static-regex tests

- [ ] 10.1 Extend `frontend/tests/brand-settings.test.mjs` with assertions:
  - `BrandSettingsPage.tsx` references `draft.logo_mode`
  - `LogoCropModal.tsx` accepts `mode` prop and branches on `mode === 'wordmark'`
  - `ShareCard.tsx` references `brand.logo_mode`
  - `PrintLayout.tsx` references `brand.logo_mode`

## 11. Build verification

- [ ] 11.1 `cd frontend && npm run lint` — clean
- [ ] 11.2 `cd frontend && npm run build` — clean (TypeScript compiles, Vite builds)
- [ ] 11.3 `cd backend && go test ./...` — passing including new repository and handler tests
- [ ] 11.4 `cd frontend && node --test tests/brand-settings.test.mjs` — all assertions pass
- [ ] 11.5 `cd frontend && node --test tests/brand-settings-dark-theme.test.mjs` — still passes (regression guard)

## 12. Manual acceptance (smoke)

- [ ] 12.1 Rebuild and deploy: `docker compose up -d --build backend frontend`
- [ ] 12.2 Visit `/settings/brand`:
  - Mode radio defaults to 图标; switching to 商标 shows the helper note
  - In 图标 mode, click 上传 → crop modal opens with 1:1 lock → confirm → 256×256 PNG uploaded
  - Switch to 商标 mode → soft warning appears next to existing square logo
  - Click 更换 → crop modal opens WITHOUT 1:1 lock → drag to 3:1 → confirm → wordmark PNG uploaded
- [ ] 12.3 Open a recent chart's report and export PDF:
  - In 图标 mode: per-page header shows `[□] <brand.title || '命 理 命 书'>     date·gender` (matches earlier fix)
  - In 商标 mode: per-page header shows `[wordmark image]     date·gender` (no text title)
  - In 商标 mode: cover banner shows the wordmark image instead of the big 命理命书 text
- [ ] 12.4 Click 保存分享图 (PNG export):
  - In 图标 mode: top brand bar shows logo + text title (as before)
  - In 商标 mode: top brand bar shows just the wordmark image centered
- [ ] 12.5 Toggle mode back to 图标 while wordmark logo is set → soft warning appears → square slot tries to render wide wordmark with object-fit:contain → user understands they need to re-upload
