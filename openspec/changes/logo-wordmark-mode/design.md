## Context

The export brand customization feature, layered with the `logo-upload-crop` change, currently produces a uniform pipeline:

```
source img → 1:1 crop in modal → 256×256 PNG → uploaded → object-fit:contain in square display slot
```

Square display slots: ShareCard 40×40, PrintLayout per-page header 18×18, PrintLayout cover (no logo today after recent fix), BrandPreviewCard 32×32.

This design assumes the user's brand is iconic (square or circular). For wordmark brands ("我的工作室", "缘聚命理", purely textual brand marks), the user's image gets either cropped (losing text) or letterboxed (rendered tiny inside whitespace).

Field interviews would not produce this insight in a 30-minute call; the design problem becomes visible only when you watch a user try to upload their actual wordmark and see the result.

## Goals / Non-Goals

**Goals:**
- Support two upload paths in one model: `icon` and `wordmark`
- Free aspect within a sensible range for wordmark to prevent extreme banners
- Preserve all existing icon-mode behavior bit-for-bit (no regression for current users)
- Mode is part of the persisted brand record (survives session, drives both crop and render)
- Wordmark mode hides text title entirely — the image is the brand mark

**Non-Goals:**
- Multiple logos per user (e.g., one icon + one wordmark)
- Per-page or per-export logo variants
- Server-side rendering of mode-dependent compositions
- Animated wordmarks
- Vector (SVG) upload — Canvas-based PNG output is the only path
- Multi-logo "kit" with primary/secondary/footer variants

## Decisions

1. **Mode field: text enum with CHECK constraint.**
   - Column: `logo_mode TEXT NOT NULL DEFAULT 'icon' CHECK (logo_mode IN ('icon', 'wordmark'))`
   - Matches existing project convention (see `watermark_mode` column on the same table — also text enum with CHECK)
   - Migration is additive and idempotent; existing rows default to `icon`

2. **Aspect constraints.**
   - Icon: `aspect = 1`, exactly as today. Cropper component receives `aspect={1}`.
   - Wordmark: `aspect = undefined` passed to Cropper (free), but a `onCropComplete` interceptor clamps the reported area to `1.5 ≤ width/height ≤ 6`. If the user drags beyond, the reported area snaps back to the nearest valid extreme.
   - Alternative considered: enforce only at upload time. Rejected because the user wouldn't see the clamp until after confirming — bad feedback loop.
   - Alternative considered: preset chips (2:1, 3:1, 4:1). Rejected because it doesn't accommodate the long tail (2.5:1, 3.2:1) of real wordmarks.

3. **Output sizing per mode.**
   - Icon: 256×256 PNG (unchanged).
   - Wordmark: fixed height 128 px, width follows aspect, capped at 768 px wide (matching the 6:1 max aspect at height 128). Output via `canvas.toBlob('image/png')`.
   - Why 128 px height: PDF print uses logo at ~24-30 px tall at 150 DPI → ~50-60 px raster equivalent. 128 px gives 2× retina headroom without bloating file size.
   - Why max 768 wide: keeps single file under ~200 KB even for high-contrast PNGs.

4. **Rendering branch table.**

   | Location | icon mode | wordmark mode |
   |----------|-----------|---------------|
   | ShareCard top brand bar (PNG) | 40×40 square logo on left + centered text title | Centered wordmark image replacing text title; logo image scales to fit bar height (~48 px) |
   | PrintLayout per-page header (PDF) | 18×18 square logo + text title beside | Wordmark image only, height auto to header bar (~5mm); date/gender stays right |
   | PrintLayout cover banner (PDF) | Big text title (28 px) | Big wordmark image (max-height 40 mm, max-width 120 mm) |
   | BrandPreviewCard (settings preview) | Square 32×32 + text title | Wordmark image replacing text title |

5. **Mode toggle UX placement.**
   - Inside the existing 顶部品牌 section, above the Logo upload row.
   - Single-line radio: `Logo 模式：[ ] 图标 [ ] 商标`
   - Helper text: "图标模式：方形 logo + 文字标题。商标模式：横版 logo 取代文字标题。"

6. **Mode switching with mismatched logo: soft warning, no auto-clear.**
   - When the user switches mode AND `serverState.logo_url` exists AND the existing logo's recorded aspect doesn't match the new mode (icon → wordmark with a square logo, or wordmark → icon with a wide logo), show an inline `<div className="brand-warning">建议重新上传符合当前模式的 logo</div>` between the radio and the logo preview row.
   - Do NOT clear the existing logo automatically. Users may have legitimate reasons to use a square logo in wordmark mode for a few minutes while preparing the new file.
   - Alternative considered: force-clear on mode switch. Rejected because it destroys data on a misclick.

7. **Title suppression in wordmark mode.**
   - In wordmark mode, `brand.title` is read but ignored by all three rendering components. The wordmark image is the brand mark.
   - The `BrandSettingsPage` form still shows the `品牌标题` input enabled — it's just visually subordinated with a small helper note "(商标模式下不会显示)".
   - Rationale: don't delete the user's saved title on mode switch; they can flip back and the title is still there.

8. **Validation belt and braces.**
   - Frontend: cropper clamps aspect during interaction.
   - Frontend before upload: a final validation in `cropToBlob` rejects out-of-range aspects (defensive; should not fire if interaction clamping works).
   - Backend: `validateBrandUpdate` accepts `logo_mode` only as `'icon'` or `'wordmark'`; any other string yields 400.
   - Backend upload endpoint does NOT enforce aspect (it accepts whatever the cropper produced). Rationale: frontend is authoritative for aspect; adding image-decoding to the backend is non-trivial and out of scope.

## Risks / Trade-offs

- **Risk:** A user uploads a perfect square icon, then switches to wordmark mode and sees their square logo rendered tiny against a wide PDF header. They blame the app.
  - **Mitigation:** Soft warning on mode switch; helper text under the mode radio describes the behavior.

- **Risk:** Old DB rows have no `logo_mode` value if the migration runs before any user touches their brand record.
  - **Mitigation:** `DEFAULT 'icon'` on the column; `INSERT ON CONFLICT UPDATE` upserts already include all columns or fall back to default.

- **Risk:** The Cropper library may not gracefully handle `aspect={undefined}` (some versions of react-easy-crop quirk on this).
  - **Mitigation:** During implementation, verify with a spike. If buggy, use `aspect={3}` as a starting value and let user drag freely (the aspect prop in react-easy-crop is a *starting* aspect — without lock, the user can change it via crop drag handles).

- **Risk:** PNG file size for a 6:1 768×128 wordmark with transparency can hit 300 KB+.
  - **Mitigation:** Acceptable. The 2 MB backend cap absorbs this; rate-limit (3/hr/user) prevents abuse.

- **Risk:** Two PDF cover banner heights (text vs image) cause page-1 layout to jump between users.
  - **Mitigation:** Render the cover banner in a fixed-height container; both text and image variants render within the same vertical box, with the image variant centered within it.

- **Risk:** A user with no logo in wordmark mode would see an empty space where the wordmark goes.
  - **Mitigation:** Treat `mode = wordmark AND logo_url = ''` as "fall back to text title" (same as icon mode). Wordmark mode only kicks in when a logo is actually uploaded.

## Trade-offs not taken

- **Preset aspect chips (2:1 / 3:1 / 4:1).** Considered for clarity. Rejected because wordmark aspect ratios cluster around 2.5–4:1 in practice and a preset would penalize 3.2:1 wordmarks. Free-with-clamp is more forgiving.

- **Backend image decoding.** Would let the server enforce aspect server-side. Rejected because the Go backend has no image library on the brand path today and adding one (e.g. `image/png` for decode, `disintegration/imaging` for analysis) introduces new failure modes (broken PNG headers, EXIF reading) without solving a real problem — the frontend is the only entrypoint and is authoritative.

- **Multi-mode per render target.** A theoretical "show wordmark on cover but icon in header" feature. Rejected — too many knobs for a feature that should be simple.

- **Auto-detect mode from upload aspect.** Mentioned in earlier exploration. Rejected because it's unpredictable and surprises users who carefully cropped a 1.7:1 image expecting icon mode.

- **Migration to convert existing 256×256 icons to wordmark mode.** No conversion happens. Existing users stay in icon mode unless they explicitly switch.
