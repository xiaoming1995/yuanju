## Context

`ai_requests_log` 表已存在并持续写入数据（每次 AI 调用无论成败均记录），字段包含：`id, chart_id, provider_id, model, duration_ms, status, error_msg, created_at`。现有 Admin 面板仅通过 `GET /api/admin/stats/ai` 提供按 Provider 聚合的成功率统计，缺乏单条请求粒度的查看能力。

## Goals / Non-Goals

**Goals:**
- 管理员能按时间线浏览 AI 调用日志明细（最新在前）
- 管理员能按状态（success/error）筛选日志
- 管理员能看到每条日志的完整错误信息（如 429 过载、超时等）
- 管理员能看到近 7 天的调用趋势和平均耗时

**Non-Goals:**
- 不做实时日志推送（WebSocket/SSE），采用刷新查询即可
- 不做日志导出（CSV/Excel）
- 不做告警规则配置（如连续失败自动通知）

## Decisions

### 1. 后端 API 设计

新增两个接口，挂载在现有的 Admin 路由组下：

| 接口 | 用途 |
|------|------|
| `GET /api/admin/ai-logs?page=1&status=error` | 分页查询日志明细，支持 status 筛选 |
| `GET /api/admin/ai-logs/summary` | 近 7 天趋势 + 平均耗时 + 错误类型分布 |

**分页策略**：每页 20 条，按 `created_at DESC` 排序。返回 `total` 字段支持前端分页器。

**为什么不复用 `/api/admin/stats/ai`？** 现有接口是按 Provider 分组的聚合统计，与日志明细的数据形态完全不同，分开更清晰。

### 2. 前端页面设计

新增 `AdminAILogsPage.tsx`，包含：
- 顶部统计卡片（总调用数、成功率、平均耗时）
- 筛选栏（全部/成功/失败）
- 日志列表表格（时间、Provider、模型、耗时、状态、错误信息）
- 错误详情通过点击行展开查看（避免表格过宽）
- 底部分页器

路由：`/admin/ai-logs`，在 Admin 侧边栏新增入口。

### 3. Repository 层

在 `admin_repository.go` 新增：
- `ListAIRequestLogs(page, pageSize int, statusFilter string)` — 分页查询，JOIN `llm_providers` 取 Provider 名称
- `GetAILogsSummary()` — 近 7 天按天分组的调用统计

## Risks / Trade-offs

- **[数据量增长]** → 日志表会随使用量增加而增长。当前阶段无需分区或归档，未来可按月归档旧数据。
- **[查询性能]** → `created_at` 和 `provider_id` 上已有索引（`idx_ai_requests_log_created` / `idx_ai_requests_log_provider`），分页查询性能有保障。
