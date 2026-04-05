## Why

当前大运模块仅支持「大运 → 流年」两层时间精度，用户无法查看流月层级的干支与十神信息，也无法精确定位当前所处的节气月份，导致命理分析时间维度不完整，实用性受限。

## What Changes

- 新增 `POST /api/bazi/liu-yue` 接口：传入流年公历年 + 日主天干，返回 12 个流月的干支、十神、节气名及节气起止日期
- 前端流年格点击后弹出抽屉（Drawer），展示该流年的 12 个流月网格
- 流月格内容：干支、天干十神、地支十神、节气名、起止日期
- 自动高亮当前所处流月（服务端基于今日节气判断，返回 `current_month_index`）
- 抽屉内提供年份切换器，允许用户查询任意年份的流月（默认当前年）

## Capabilities

### New Capabilities
- `liuyue-query`: 按需查询任意流年的 12 个流月干支、十神与节气起止日期的接口能力
- `liuyue-drawer-ui`: 前端流年格点击弹出流月抽屉、当前流月高亮、年份切换的交互能力

### Modified Capabilities
- `bazi-advanced-data`: 大运数据层扩展——新增流月查询 API，现有 /calculate 响应不变

## Impact

- **后端**：`internal/handler/bazi_handler.go` 新增路由处理函数、`pkg/bazi/engine.go` 新增流月计算函数（调用 `LiuNian.GetLiuYue()` + 节气日期查表）
- **前端**：新增 `LiuYueDrawer.tsx` 组件，修改 `DayunTimeline.tsx` 为流年格绑定点击事件
- **API**：新增公开接口 `POST /api/bazi/liu-yue`（无需登录）
- **依赖**：`lunar-go v1.4.6` 已原生支持 `GetLiuYue()`，无需升级
