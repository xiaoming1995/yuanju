## Why

目前 AI 调用日志虽然已入库（`ai_requests_log` 表），但 Admin 面板只能看到按 Provider 聚合的成功率统计，无法查看单条请求的时间线、耗时、具体错误信息。当 Provider 出现 429 过载、超时等问题时，管理员无法快速定位故障原因和影响范围。

## What Changes

- 新增后端 API `GET /api/admin/ai-logs`，支持分页查询 AI 调用日志明细（时间、Provider、模型、耗时、状态、错误信息）
- 新增后端 API `GET /api/admin/ai-logs/stats`，提供更丰富的统计维度（近 7 天趋势、平均耗时、错误分布）
- 新增 Admin 前端"AI 调用日志"页面，包含：
  - 日志明细列表（支持按状态筛选、按时间排序、分页）
  - 错误详情展开查看
  - 调用趋势图表（近 7 天成功/失败数量）
  - 平均响应耗时统计

## Capabilities

### New Capabilities
- `admin-ai-log-viewer`: Admin 面板的 AI 调用日志查看与分析功能，包括明细列表 API、统计 API 和前端页面

### Modified Capabilities
（无需修改现有能力规格，日志写入逻辑已完善）

## Impact

- **后端**：`internal/handler/admin_handler.go` 新增 2 个 handler；`internal/repository/admin_repository.go` 新增查询函数；`cmd/api/main.go` 注册新路由
- **前端**：`src/pages/admin/` 新增 AI 日志页面组件；`src/lib/adminApi.ts` 新增 API 调用
- **数据库**：无需 schema 变更，`ai_requests_log` 表结构已满足需求
