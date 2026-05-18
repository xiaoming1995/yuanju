## ADDED Requirements

### Requirement: Logo upload runs through a cropping modal
Brand logo selection in the export brand settings page SHALL open a cropping modal before sending the image to the server.

#### Scenario: User selects a logo file
- **WHEN** a logged-in user selects a PNG/JPG/WebP file from the brand settings logo input
- **THEN** the settings page opens a cropping modal showing the selected image with a 1:1 crop region overlay
- **AND** no network request to the logo upload endpoint is made yet

#### Scenario: User cancels the crop
- **WHEN** the cropping modal is open and the user clicks "取消"
- **THEN** the modal closes
- **AND** the previously stored logo (if any) is unchanged
- **AND** no upload request is sent

### Requirement: Cropping is constrained to 1:1 with pan and zoom
The cropping modal SHALL fix the crop aspect ratio at 1:1 and allow the user to pan and zoom to position the crop region.

#### Scenario: Pan the crop region
- **WHEN** the cropping modal is open
- **THEN** the user can drag (mouse or touch) to move the crop region across the source image

#### Scenario: Zoom the crop region
- **WHEN** the cropping modal is open
- **THEN** the user can change zoom via mouse wheel, pinch gesture, or a zoom control

#### Scenario: Aspect remains 1:1
- **WHEN** any pan or zoom interaction completes
- **THEN** the crop region's width-to-height ratio remains exactly 1:1

### Requirement: Confirmed crop produces a 256×256 PNG
On confirm, the cropping modal SHALL produce a 256×256-pixel PNG from the selected region and upload it via the existing logo endpoint.

#### Scenario: Confirm a crop on a large source
- **WHEN** the source image is larger than 256 pixels on the short axis and the user clicks "确认"
- **THEN** the modal draws the cropped region scaled to fit 256×256 onto a canvas
- **AND** the canvas is exported as PNG via `canvas.toBlob('image/png')`
- **AND** the resulting Blob is wrapped as a `File('logo.png', 'image/png')` and POSTed to `/api/user/export-brand/logo`

#### Scenario: Confirm a crop on a small source
- **WHEN** the source image is smaller than 256 pixels on the short axis
- **THEN** the canvas still outputs a 256×256 PNG (the cropped region is upscaled)
- **AND** the upload proceeds normally

#### Scenario: Transparent PNG preserved
- **WHEN** the source image is a PNG with alpha-channel transparency and the user confirms
- **THEN** the uploaded PNG retains the alpha channel in the cropped region

### Requirement: Large source images are downscaled before cropping
The cropping modal SHALL preemptively downscale source images larger than 1600 pixels on the long axis before passing them to the cropper, to avoid mobile memory pressure.

#### Scenario: Very large image uploaded
- **WHEN** the user selects an image whose longest dimension exceeds 1600 pixels
- **THEN** the modal scales the image to fit within 1600×1600 (preserving aspect) before displaying it in the cropper
- **AND** the user sees the same crop region UX as for smaller images

### Requirement: Re-cropping requires re-uploading the source file
The system SHALL NOT store the user's original uncropped image. Adjusting the crop region requires the user to select the source file again.

#### Scenario: User wants to adjust crop after saving
- **WHEN** a logo already exists for the user and they want to change the crop region
- **THEN** they re-select the source file via the "更换" button
- **AND** the cropping modal opens with the freshly selected file (not the stored cropped logo)

### Requirement: Animated images collapse to the first frame
Cropping animated images (GIF, animated WebP) SHALL result in a static PNG capturing only the first frame.

#### Scenario: User selects an animated GIF
- **WHEN** the user selects an animated image and confirms a crop
- **THEN** the uploaded PNG is the first-frame crop, with no animation preserved
- **AND** the modal displays a notice "动图（GIF/动 WebP）将仅保留第一帧"
