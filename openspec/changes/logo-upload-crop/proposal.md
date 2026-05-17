## Why

Brand logo upload (导出品牌设置) currently accepts any image and renders it through CSS `object-fit: contain` in a 40×40 box. Non-square logos (banner-shaped brand marks, screenshots, photos) shrink to a fraction of the box with large whitespace, producing a poor visual in exported PNG and PDF. Users have no way to choose which part of their image becomes the displayed logo.

## What Changes

- Open a cropping modal after the user selects a logo file but before uploading.
- Use `react-easy-crop` to provide a fixed 1:1 crop region with drag-to-reposition and pinch/scroll zoom.
- On confirm, render the chosen region to a 256×256 PNG via Canvas API, then POST that result to the existing logo endpoint.
- Cancel discards the file selection; the previously stored logo (if any) is unchanged.
- Keep the change frontend-only; the backend upload endpoint, DB schema, and rendering components are not modified.

## Capabilities

### New Capabilities
- `logo-upload-crop`: Brand logo uploads run through a client-side 1:1 cropping modal that produces a fixed 256×256 PNG, ensuring exported PNG/PDF show a consistently sized logo regardless of the source image's aspect ratio.

### Modified Capabilities
- None. The export brand customization feature already lives in commits `eb6000d..ca2ed5b` on `feat/export-brand-customization`; this proposal layers a UX refinement on top of it without changing its data contract.

## Impact

- Affected frontend files:
  - `frontend/package.json` — adds `react-easy-crop` (~8KB gzip).
  - `frontend/src/components/LogoCropModal.tsx` — new modal component.
  - `frontend/src/components/LogoCropModal.css` — new styles.
  - `frontend/src/pages/BrandSettingsPage.tsx` — modify `onLogoChange` to open the modal instead of uploading directly.
  - `frontend/tests/brand-settings.test.mjs` — add 2 static assertions for the modal.
- No backend, DDL, API contract, or rendering-component changes.
- Reverting is a clean removal of the new files plus restoring `onLogoChange` to the direct-upload code path.
