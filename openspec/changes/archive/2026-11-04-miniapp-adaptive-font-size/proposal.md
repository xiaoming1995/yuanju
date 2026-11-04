## Why

小程序多处正文和说明文字偏小，且固定字号无法响应用户在微信中的字体大小偏好。用户需要更清晰的默认文字，并能在“我的”页自行调整。

## What Changes

- 提高小程序常用正文、说明和操作文字的默认字号，保留标题与数据的清晰层级。
- 默认跟随微信字体大小设置，并在页面重新展示时更新。
- 在“我的 → 设置与帮助”增加字体大小选择：跟随微信、标准、较大、特大；手动选择在本机保存。
- 适配放大字体后的日历格、命盘、卡片、按钮和长文本布局。

## Capabilities

### New Capabilities

- `miniapp-adaptive-typography`: 小程序字体大小跟随、手动覆盖、持久化及页面可读性。

### Modified Capabilities

无。

## Impact

影响 `miniapp/app.js`、`miniapp/app.wxss`、`miniapp/pages/` 下的页面样式与“我的”页交互，以及对应的前端检查。不变更后端、数据库和依赖。
