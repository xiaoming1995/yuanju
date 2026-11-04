## Why

小程序“我的”页的本人命盘卡片把默认名称、四柱和两处操作排成近似普通资料卡，视觉重点不够明确。用户已确认以原局页的暖白、深棕和浅金色调，改为一张更简洁的四柱摘要卡。

## What Changes

- 已绑定状态以天干、地支上下排列的四列展示年、月、日、时柱，并用浅色底突出日柱。
- 保留“本人命盘”、查看详情、出生时间和独立的修改资料入口；默认“我的干支”不再重复显示，用户自定义名称仍可识别。
- 保持未绑定、读取中和失败状态，以及现有导航、绑定和修改流程。

## Capabilities

### New Capabilities

- `miniapp-personal-chart-card-presentation`: “我的”页本人命盘卡片的四柱层级、原局色调与不同状态下的可用性。

### Modified Capabilities

无。

## Impact

仅涉及 `miniapp/pages/mine/index.wxml`、`index.wxss`，必要时调整卡片数据映射及对应小程序检查。不改动后端、数据库、其他页面或依赖。
