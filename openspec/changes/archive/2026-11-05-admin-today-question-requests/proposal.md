## Why

“今日问事”已经产生真实 AI 请求，但通用调用日志没有问题、依据、完整输入输出和逐次 Token 用量，且依据不足的请求没有 AI 用量记录。管理员无法核对某次回答和成本，也无法排查失败。

## What Changes

- 记录每次已绑定用户的今日问事请求、问题、分析日期、命盘、状态和最终回答；AI 调用时另存实际输入、原始输出、模型、耗时及 Token 用量。
- 增加仅管理员可访问的分页列表与单条详情 API，支持按日期、状态和用户搜索。
- 增加后台独立“今日问事”页面，列表看请求和消耗，详情按需展开命理上下文、AI 输入输出及最终回答。
- 失败、依据不足和未产生 Token 的请求也可核查，敏感正文不出现在列表响应中。

## Capabilities

### New Capabilities

- `admin-today-question-requests`: 管理员审查今日问事请求、AI 输入输出与 Token 消耗。

### Modified Capabilities

无。

## Impact

新增 PostgreSQL 迁移与后端 repository、管理员 API；调整今日问事服务和 handler 的审计写入；增加 React 管理页面、路由、侧边栏与 API 封装。复用现有管理员 JWT、Provider 与定价逻辑，不增加第三方依赖。
