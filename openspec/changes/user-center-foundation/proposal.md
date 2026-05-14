## Why

缘聚目前只有登录、注册、历史记录等分散入口，还没有真正的用户个人中心。后续充值系统、点数账户、PDF 模板个性化都需要一个统一的账户页作为承载入口，因此应先建设个人中心基础能力，避免支付和模板定制直接散落在各页面中。

## What Changes

- 新增普通用户个人中心页面 `/profile`，作为“我的”入口。
- 新增用户中心聚合接口，返回用户基本信息、命盘数量、AI 报告数量、合盘记录数量、最近记录摘要等账户概览数据。
- 导航调整：登录用户在顶部和底部导航可进入个人中心；历史记录作为个人中心中的核心入口之一保留。
- 个人中心展示后续能力入口占位：充值/点数、PDF 模板设置，但本变更不实现支付渠道、不实现点数扣费、不实现 PDF 模板引擎。
- 保持现有认证、历史记录、结果页和 PDF 导出功能兼容。

## Capabilities

### New Capabilities

- `user-center-foundation`: 普通用户个人中心基础能力，包括账户概览、数据统计、最近记录入口、后续商业化模块入口占位和导航集成。

### Modified Capabilities

- None.

## Impact

- **后端**：新增用户中心 handler/service/repository 聚合查询；新增 `/api/user/profile` 或同等鉴权接口。
- **前端**：新增 `ProfilePage`、API 类型与请求方法；更新 `App.tsx`、`Navbar`、`BottomNav` 的个人中心入口。
- **数据库**：优先复用现有 `users`、`bazi_charts`、`ai_reports`、`compatibility_readings` 等表；本变更不要求新增充值或 PDF 模板表。
- **后续依赖**：充值系统与 PDF 模板定制可以在该页面和接口结构之上扩展。
