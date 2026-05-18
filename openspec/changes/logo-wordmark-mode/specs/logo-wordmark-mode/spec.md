## ADDED Requirements

### Requirement: Brand record persists a logo display mode
The brand record SHALL include a `logo_mode` field with values `icon` (default) or `wordmark`, and this field SHALL be settable independently of the logo file.

#### Scenario: New users default to icon mode
- **WHEN** a user opens the brand settings page for the first time
- **THEN** the form shows `Logo 模式 = 图标` selected
- **AND** the persisted brand record has `logo_mode = 'icon'`

#### Scenario: User toggles to wordmark mode and saves
- **WHEN** a logged-in user changes the Logo 模式 radio from 图标 to 商标 and clicks 保存
- **THEN** the API request body includes `logo_mode: "wordmark"`
- **AND** subsequent `GET /api/user/export-brand` returns `logo_mode: "wordmark"`

#### Scenario: Invalid mode is rejected
- **WHEN** the API receives `logo_mode: "square"` (or any value other than `icon`, `wordmark`, or empty string)
- **THEN** the response is HTTP 400
- **AND** no DB write occurs

### Requirement: Icon mode preserves 1:1 crop behavior
When `logo_mode = icon`, the upload pipeline SHALL preserve the existing behavior introduced by `logo-upload-crop`: a 1:1 aspect lock in the cropper and a 256×256 PNG output.

#### Scenario: User uploads in icon mode
- **WHEN** the brand's draft `logo_mode = 'icon'` and the user selects a file
- **THEN** `LogoCropModal` opens with `aspect={1}` on the `<Cropper>`
- **AND** confirming the crop produces a 256-pixel-square PNG

#### Scenario: Pan and zoom remain available
- **WHEN** the cropper is open in icon mode
- **THEN** drag-to-pan and pinch/wheel-to-zoom behave identically to the previous logo-upload-crop behavior

### Requirement: Wordmark mode unlocks free aspect within clamped range
When `logo_mode = wordmark`, the cropper SHALL accept any aspect ratio within the closed interval `[1.5, 6]` and SHALL output a PNG of fixed height 128 pixels with width following the chosen aspect ratio, capped at 768 pixels wide.

#### Scenario: User uploads a 2.5:1 wordmark
- **WHEN** the brand's draft `logo_mode = 'wordmark'` and the user drags the crop region to roughly 2.5:1
- **THEN** `LogoCropModal` does NOT snap or override the aspect
- **AND** confirming produces a PNG of approximately 320×128 pixels (2.5 × 128)

#### Scenario: User tries to drag past the upper aspect cap
- **WHEN** the user drags the crop region to an aspect wider than 6:1 in wordmark mode
- **THEN** the cropper's `onCropComplete` interceptor clamps the area's aspect to exactly 6:1
- **AND** the confirmed PNG is at most 768×128 pixels

#### Scenario: User tries to drag below the lower aspect cap
- **WHEN** the user drags the crop region to an aspect narrower than 1.5:1 in wordmark mode
- **THEN** the cropper's `onCropComplete` interceptor clamps the area's aspect to exactly 1.5:1
- **AND** the confirmed PNG is at least 192×128 pixels

### Requirement: Wordmark mode replaces text titles in export products
When `logo_mode = wordmark` AND `logo_url` is non-empty, every export-rendering component SHALL omit the text title and render the wordmark image in its place.

#### Scenario: PDF per-page header in wordmark mode
- **WHEN** the user exports a PDF and their brand record has `logo_mode = 'wordmark'` and `logo_url` set
- **THEN** each printed page's top header shows the wordmark image (height ~6 mm) on the left and the date·gender info on the right
- **AND** the text title `<brand.title>` or `命 理 命 书` is NOT shown

#### Scenario: PDF cover banner in wordmark mode
- **WHEN** the user exports a PDF in wordmark mode
- **THEN** the first-page cover replaces the 28-pixel text title block with the wordmark image at `max-height: 40mm; max-width: 120mm`
- **AND** the small kicker line is hidden

#### Scenario: PNG share card in wordmark mode
- **WHEN** the user clicks 保存分享图 in wordmark mode
- **THEN** the ShareCard's top brand bar shows the wordmark image centered (max-height 48 px, max-width 320 px) instead of the text title
- **AND** the 40×40 absolute-positioned left logo slot is hidden

#### Scenario: brand.title is preserved but unused in wordmark mode
- **WHEN** a user with `logo_mode = 'wordmark'` had previously set `brand.title = '缘聚命理'`
- **THEN** the value remains in the DB and remains in the input field on the settings page
- **AND** export products do NOT render that text anywhere

### Requirement: Wordmark mode falls back to text title when no logo is uploaded
When `logo_mode = wordmark` AND `logo_url` is empty, the system SHALL behave as if `logo_mode = icon` for rendering purposes.

#### Scenario: Wordmark mode with no logo uploaded
- **WHEN** the user has selected 商标 mode in settings but has not uploaded a logo
- **THEN** the PDF and PNG export products render the text title `<brand.title>` (or default) — i.e., the same as icon mode with no logo

### Requirement: Switching modes preserves existing logo but warns on aspect mismatch
The system SHALL NOT delete the user's existing logo file when they switch `logo_mode`. The settings page SHALL show a soft warning when the existing logo's aspect ratio does not match the newly selected mode.

#### Scenario: Switching from icon to wordmark with a square logo present
- **WHEN** the user has a 256×256 logo and toggles the mode radio from 图标 to 商标
- **THEN** the existing logo is NOT cleared
- **AND** an inline warning text appears next to the logo preview: `建议重新上传符合当前模式的 logo`
- **AND** the user may still save the form with the mismatched logo (it will render letterboxed in wordmark slots until replaced)

#### Scenario: Switching from wordmark to icon with a wide wordmark present
- **WHEN** the user has a 384×128 wordmark and toggles the mode radio from 商标 to 图标
- **THEN** the existing wordmark is NOT cleared
- **AND** the same inline warning appears
- **AND** the user may still save (the wordmark renders inside the square 40×40 / 18×18 slots via `object-fit: contain` until they re-upload)
