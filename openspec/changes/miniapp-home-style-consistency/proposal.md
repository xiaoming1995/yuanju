## Why

小程序首页已采用暖白纸张、浅金细线和轻量卡片，但从首页进入今日详情、择日、日历及排盘结果时，深色主视觉、冷灰表面和不同的标题样式造成明显跳变。统一这些页面的视觉规则，可以让跨页阅读更连贯，同时保留各项数据的原有含义。

## What Changes

- 以当前首页每日参考区域为视觉基线，统一其他小程序页面的标题层级、卡片表面、边框与阴影、文字色阶、区块间距和操作强调。
- 优先调整今日完整依据与择日页的深色顶部、排盘结果页的冷灰及紫色装饰；协调日历页和“我的”相关页面的局部差异。
- 保留宜忌、五行、状态级别等有业务含义的颜色；保留页面已有的真实数据、加载／错误／空状态和导航操作。
- 通过微信开发者工具或真机检查常见与窄屏尺寸，以及长文案、不同登录和绑定状态。

## Capabilities

### New Capabilities

- `miniapp-visual-consistency`: 定义小程序跨页面视觉连续性、重点与普通内容的层级，以及业务状态颜色和可读性要求。

### Modified Capabilities

无。现有 `miniapp-paper-background` 已规定共享纸张背景；本变更补充页面内容层的统一规则。

## Impact

- 涉及 `miniapp/pages/calendar/`、`personal-day/`、`moving-date/`、`result/`、`mine/`、`profile/`、`personal-chart-setup/` 的局部 WXML/WXSS；仅在确有重复规则时整理 `miniapp/app.wxss`。
- 首页当前 WXML/WXSS 用作参照，历史态差异可在同一轮校准；不修改后端 API、排盘与历法逻辑、数据库或依赖。
- 现有 `miniapp-home-visual-refinement` 仍有窄屏视觉验收项，实施时以实际工作区和最新首页画面为准，避免覆盖其未完成验证。
