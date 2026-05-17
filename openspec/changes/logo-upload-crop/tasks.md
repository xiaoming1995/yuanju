## 1. Tests

- [ ] 1.1 Add static-regex test asserting `LogoCropModal` component file exists and exports a default React component.
- [ ] 1.2 Add static-regex test asserting `BrandSettingsPage` no longer calls `brandAPI.uploadLogo` directly from `onLogoChange` — modal flow is the only upload path.

## 2. Dependency

- [ ] 2.1 Add `react-easy-crop` to `frontend/package.json` and run `npm install`.

## 3. Cropper component

- [ ] 3.1 Create `frontend/src/components/LogoCropModal.tsx` with props `{ sourceDataUrl: string; open: boolean; onConfirm: (file: File) => void; onCancel: () => void }`.
- [ ] 3.2 Inside the modal, render `<Cropper image={sourceDataUrl} aspect={1} crop zoom onCropChange onZoomChange onCropComplete={(_, areaPx) => setArea(areaPx)} />`.
- [ ] 3.3 Implement `confirm` handler that draws the cropped region onto a 256×256 `<canvas>` via `ctx.drawImage(img, x, y, w, h, 0, 0, 256, 256)`, then `canvas.toBlob(blob, 'image/png')`, wraps blob as `new File([blob], 'logo.png', { type: 'image/png' })`, and calls `onConfirm(file)`.
- [ ] 3.4 Implement `cancel` handler that calls `onCancel()` without state changes.
- [ ] 3.5 Add `frontend/src/components/LogoCropModal.css` with backdrop, modal frame, footer buttons matching the brand-settings style (tan border, gold accent).

## 4. Pre-emptive downscale

- [ ] 4.1 In `LogoCropModal`, when receiving `sourceDataUrl`, load into an `<img>`, check `naturalWidth/Height`; if either > 1600, draw onto an offscreen canvas scaled to fit 1600×1600 and feed the resulting dataURL to `Cropper` instead. Otherwise pass through.

## 5. BrandSettingsPage integration

- [ ] 5.1 Add state `const [cropSourceUrl, setCropSourceUrl] = useState<string | null>(null)` to `BrandSettingsPage.tsx`.
- [ ] 5.2 Modify `onLogoChange`: instead of calling `brandAPI.uploadLogo(file)`, run client-side validation (≤ 2MB, MIME in {png, jpeg, webp}), then `FileReader.readAsDataURL(file)` and `setCropSourceUrl(reader.result)`.
- [ ] 5.3 Add `<LogoCropModal sourceDataUrl={cropSourceUrl ?? ''} open={!!cropSourceUrl} onConfirm={handleCropConfirm} onCancel={() => setCropSourceUrl(null)} />` to the page.
- [ ] 5.4 Implement `handleCropConfirm(file: File)`: closes modal (`setCropSourceUrl(null)`), sets `uploading=true`, calls `brandAPI.uploadLogo(file)`, updates `serverState.logo_url` on success, sets error on failure.

## 6. Animated-image note

- [ ] 6.1 Add a small `<small>` text in `LogoCropModal` body: "动图（GIF/动 WebP）将仅保留第一帧".

## 7. Verification

- [ ] 7.1 Run `cd frontend && node --test tests/brand-settings.test.mjs` — all assertions pass.
- [ ] 7.2 Run `cd frontend && npm run build` — clean build, no new TypeScript errors.
- [ ] 7.3 Run `cd frontend && npm run lint` — no new lint errors.
- [ ] 7.4 Manual smoke: in the running app, upload a 1000×200 banner JPG, drag the 1:1 crop region to the brand-text portion, confirm, verify the resulting saved logo is a square crop of that region.
- [ ] 7.5 Manual smoke: upload a 100×100 small PNG with transparency, confirm, verify the saved PNG still has transparency (open in image viewer or check `<img>` against checkered background).
