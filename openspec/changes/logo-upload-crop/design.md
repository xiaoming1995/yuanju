## Context

The `feat/export-brand-customization` branch already ships a working logo upload (multipart POST to `/api/user/export-brand/logo` with 3-layer validation + rate limit). Logos are stored server-side as `<UploadDir>/brand-logos/<uuid>.<ext>` and rendered by `ShareCard`, `PrintLayout`, and `BrandPreviewCard` via CSS `object-fit: contain` inside 32-40px square boxes.

`object-fit: contain` preserves aspect ratio but shrinks non-square images to the box's short axis, leaving whitespace on the long axis. A 1000×200 banner logo therefore renders as a tiny 40×8 strip with 32px of white padding above and below — a visible quality regression compared with users who upload pre-cropped square logos.

This change shifts the burden from "user must pre-crop in Photoshop" to "app provides a crop UI before upload" without changing what the server stores.

## Goals / Non-Goals

**Goals:**
- Every uploaded logo arrives at the server as a 256×256 PNG (1:1 aspect, fixed resolution).
- User has visual control over which region of the source image becomes the logo (pan + zoom).
- Touch-friendly (the brand-settings page is reachable from the mobile profile entry).
- Zero backend changes; existing rate limit, magic-bytes check, and 2MB cap continue to apply.

**Non-Goals:**
- Multiple aspect ratios (1:1, 16:9, free crop). Locked to 1:1 for MVP — adding more ratios requires `ShareCard`/`PrintLayout`/`BrandPreviewCard` to accept non-square logo containers, which is a larger refactor.
- Server-side image processing (e.g., Go `disintegration/imaging`). Kept client-side to avoid the dependency and CPU cost.
- Preserving the user's original uncropped upload for later re-cropping. Re-cropping requires the user to re-select the source file.
- Animated image support — GIF/animated WebP collapse to a single frame after Canvas capture, which is fine for static logos but should be noted.
- Multi-resolution output (40 + 200 + 400 variants). Single 256×256 PNG is enough for display + PDF print at ~150dpi.

## Decisions

1. **Library: `react-easy-crop`.**
   - 8KB gzip, simpler API than `react-image-crop`, built-in pinch zoom for touch.
   - Alternative considered: hand-rolled Canvas. Rejected because pan + pinch zoom on mobile is non-trivial and would balloon component code.

2. **Aspect ratio: 1:1 only.**
   - Matches the 40×40 / 32×32 display boxes in all three rendering components.
   - Avoids a major refactor of those components to accept variable-aspect logos.

3. **Output resolution: 256×256 PNG.**
   - Retina-friendly (covers 40px display ~6×).
   - PDF print at ~150dpi is clear without forcing a 400×400 file size penalty.
   - PNG (not JPEG) preserves transparency for logos with alpha channels.

4. **No re-crop without re-upload.**
   - Server stores only the cropped 256×256 PNG. To change the crop, user re-selects the source file.
   - Adding `logo_original_path` alongside `logo_path` would triple the data model complexity (DDL change, dual delete, dual rate limit) for an infrequent operation.

5. **Crop happens client-side before the network call.**
   - Source file is read via `FileReader.readAsDataURL`, fed to `<img>`, then `react-easy-crop` reports `{x, y, width, height}` in source-image pixels.
   - On confirm, a 256×256 `<canvas>` is created, `ctx.drawImage(img, cropX, cropY, cropW, cropH, 0, 0, 256, 256)`, then `canvas.toBlob(blob, 'image/png')` becomes the upload payload.
   - The Blob is wrapped as a `File('logo.png', ...)` and passed to the existing `brandAPI.uploadLogo(file)` — no API change.

6. **Pre-emptive downscale for very large source images.**
   - Images > 1600px on the long axis are first downscaled to fit 1600×1600 before being shown in the cropper. Prevents mobile OOM on 4000×4000 photo uploads.
   - Implemented with `createImageBitmap` + offscreen canvas where supported, with a plain `<img>` fallback.

## Risks / Trade-offs

- **Risk:** First-time users hit an unexpected modal between "click upload" and "save".
  - **Mitigation:** Modal opens immediately on file selection. Title text states "调整 logo 裁剪区域". Cancel returns to the settings page with no state changes.

- **Risk:** Animated GIF/WebP uploads silently become static (only first frame captured).
  - **Mitigation:** Static logos are the common case. Add a small note in the modal: "动图将仅保留第一帧". Not blocking.

- **Risk:** Some browsers fail `canvas.toBlob` for images with cross-origin headers.
  - **Mitigation:** Source comes from `FileReader.readAsDataURL` which produces a same-origin data URL — no CORS issue. Tested behavior preserved.

- **Risk:** `react-easy-crop` adds a dependency. If the project later decides to standardize on a different image-manipulation library, this is one more thing to migrate.
  - **Mitigation:** The library surface used (`<Cropper aspect crop onCropChange onCropComplete zoom onZoomChange>` and `getCroppedImg`-style helper) is small; replacement effort is ~1 day.

- **Risk:** Output PNG always re-encodes (even if source was already 1:1 256×256 PNG).
  - **Mitigation:** Acceptable — re-encoding cost is small, and the contract "server always gets 256×256 PNG" is worth the simplicity.

## Trade-offs not taken

- **Server-side processing.** Pros: simpler frontend. Cons: requires a Go imaging library (`disintegration/imaging` or similar), increases backend CPU per upload, requires writing crop coordinates in the API. Rejected for client-side because the UI needs interactive crop preview anyway, and once the canvas is producing pixels there's no reason to send the original.
- **Multiple aspect ratios.** Pros: banner logos can stay banner-shaped. Cons: rendering components need conditional layout. Deferred to a later proposal if user demand emerges.
- **Preserving original.** Pros: re-crop without re-upload. Cons: doubles storage, doubles cleanup logic. Deferred — re-upload is acceptable for an infrequent settings operation.
