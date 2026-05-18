## Why

Today's logo upload pipeline (introduced in `logo-upload-crop`) forces every uploaded image through a 1:1 crop into a 256×256 PNG. This serves users whose brand is a square or circular icon, but penalizes users whose brand is a horizontal **wordmark** (text logo like "缘聚命理"、"DIOR"、"我的工作室"). Their wordmark either gets cropped, or they have to manually add whitespace top/bottom in the cropper, after which it renders as a tiny strip inside the 18×18 / 40×40 square display slot — visually small and surrounded by whitespace.

Brand identity is the whole point of the export brand customization feature. A user who uploads a wordmark is signaling "my brand IS the logo, replace the default 命理命书 with this". The current pipeline can't honor that.

## What Changes

- Introduce a `logo_mode` field on `user_export_brand` with values `icon` (default) and `wordmark`.
- Add a "Logo 模式：图标 / 商标" radio in the `/settings/brand` page, inside the 顶部品牌 section.
- `LogoCropModal` reads the draft mode and adjusts crop behavior:
  - `icon` mode: keep current 1:1 aspect lock, 256×256 PNG output
  - `wordmark` mode: free crop with aspect clamped to `[1.5, 6]`, output `height = 128, width = auto, max = 768×128 PNG`
- `PrintLayout`, `ShareCard`, `BrandPreviewCard` branch their layout by mode:
  - `icon` mode: current behavior (square logo slot beside the title text)
  - `wordmark` mode: logo image **replaces** the text title in both the per-page PDF header AND the cover banner big-title block; `brand.title` is ignored in this mode
- Switching modes with a mismatched existing logo (e.g., square logo while switching to wordmark mode) shows a soft warning prompting re-upload, but does NOT clear the existing logo.

## Capabilities

### New Capabilities
- `logo-wordmark-mode`: Brand logos can be uploaded and rendered in either icon (1:1) or wordmark (horizontal free crop, capped aspect) mode. Wordmark mode replaces the text title in export products; icon mode preserves the title alongside the logo.

### Modified Capabilities
- `logo-upload-crop`: The 1:1 aspect lock is no longer absolute. It now applies only when the user's `logo_mode = icon` (the default). When `logo_mode = wordmark`, the cropper allows free aspect within `[1.5, 6]`.
- `export-brand-customization`: The rendering rule "show `brand.title` text in header / cover" now branches by `logo_mode`. In wordmark mode, the title text is suppressed and the logo image takes its place.

## Impact

**Backend:**
- `backend/pkg/database/database.go` — add `logo_mode TEXT NOT NULL DEFAULT 'icon'` column with `CHECK (logo_mode IN ('icon', 'wordmark'))` to `user_export_brand`
- `backend/internal/model/user_brand.go` — add `LogoMode string` field to `ExportBrand` struct
- `backend/internal/repository/user_brand_repository.go` — extend `Upsert` / `Get` for new column
- `backend/internal/handler/user_brand_handler.go` — validate `logo_mode` in `UpdateExportBrand`; emit it in the GET response

**Frontend:**
- `frontend/src/lib/api.ts` — add `logo_mode` to `ExportBrand` TypeScript interface
- `frontend/src/pages/BrandSettingsPage.tsx` — add mode radio control + pass draft mode into `LogoCropModal`
- `frontend/src/components/LogoCropModal.tsx` — accept `mode: 'icon' | 'wordmark'` prop; gate `aspect` and clamp logic; switch output size
- `frontend/src/components/ShareCard.tsx` — render wordmark image full-width replacing title text when `brand.logo_mode === 'wordmark'`
- `frontend/src/components/PrintLayout.tsx` — same branching for both per-page header and cover banner
- `frontend/src/components/BrandPreviewCard.tsx` — same branching so the preview reflects mode
- `frontend/tests/brand-settings.test.mjs` — add static-regex assertions for mode prop wiring

**Database migration:** New nullable column with default; existing rows auto-populate as `icon`. No data backfill needed.

**Reverse course on prior decision:** This change formally relaxes the "aspect ratio: 1:1 only" decision logged in `openspec/changes/logo-upload-crop/design.md` decision #2. The original spec's wording will continue to apply for users in icon mode; wordmark mode is a strict opt-in via the new field.
