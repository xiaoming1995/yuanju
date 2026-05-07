## Why

过往事件推算页面（PastEventsPage）按大运分组展示年份卡片，但每个大运组只有一行标题，缺乏对这10年周期的整体评述。用户需要先逐年阅读才能形成对某段大运的宏观感知，体验割裂。

## What Changes

- 在"过往事件推算"AI Prompt 模板中增加对 `dayun_summaries[]` 的生成要求
- 扩展后端 AI 响应 JSON 结构，新增 `dayun_summaries` 数组（含 `gan_zhi`、`themes`、`summary` 字段）
- 前端 PastEventsPage 解析 `dayun_summaries`，在每个大运组顶部渲染主题标签 + 叙述段

## Capabilities

### New Capabilities

- `dayun-summary`: 大运整体总结 — 每个大运组顶部展示 AI 生成的主题关键词标签（如「事业↑」「贵人」）和一段 80-120 字的整体运势叙述

### Modified Capabilities

- `past-year-events`: AI 响应 JSON 结构扩展（新增 `dayun_summaries` 字段），Prompt 模板更新，前端解析逻辑更新

## Impact

- `backend/pkg/seed/seed.go`：更新 `past_events` Prompt 模板，增加 `dayun_summaries` 生成指令
- `backend/internal/service/report_service.go`：`GeneratePastEventsStream` 构建 Prompt 时传入大运列表数据（`{{.DayunList}}`）
- `frontend/src/pages/PastEventsPage.tsx`：新增 `DayunSummary` 类型，解析 `dayun_summaries`，渲染大运总结块
- 不涉及数据库 schema 变更（`content_structured` JSONB 兼容扩展）
- 不涉及新 API 端点
