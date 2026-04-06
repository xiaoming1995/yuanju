## Why

管理后台「全站起盘明细」页面存在多个数据展示缺陷：`ai_reports` 表无唯一约束导致 JOIN 产生重复行，`ai_result` 字段直接暴露原始 JSON 字符串，以及流年记录加载时 loading 状态全局污染。这些问题直接影响管理员对数据的判断准确性，需立即修复。

## What Changes

- **修复 SQL 重复行**：`ListBaziCharts` 查询改为 `LEFT JOIN LATERAL`（取最新一条报告），或用子查询 `ORDER BY created_at DESC LIMIT 1` 避免笛卡尔积
- **修复 AI 报告展示**：管理详情面板中，`ai_result` 展示从"原始 JSON 文本"改为解析 `content_structured` 字段后结构化呈现（性格/事业/感情/健康）
- **修复 loading 状态**：将全局 `liunianLoading` 改为按 `chartId` 隔离的 Map 状态，避免快速切换时的竞争条件
- **后端补充 `content_structured` 字段**：`ListBaziCharts` SQL 新增 JOIN `ai_reports.content_structured` 字段供前端解析

## Capabilities

### New Capabilities
- `admin-charts-display`：管理后台起盘明细数据展示规范，涵盖重复数据去除规则、AI 报告结构化显示格式、以及流年加载状态管理

### Modified Capabilities
<!-- 无 spec 层行为变化，仅修复实现缺陷 -->

## Impact

- `backend/internal/repository/admin_repository.go`：修改 `ListBaziCharts` SQL 查询逻辑
- `backend/internal/model/admin.go`：`AdminChartRecord` 新增 `AIResultStructured` 字段
- `frontend/src/pages/admin/AdminChartsPage.tsx`：修复 loading 状态 + 改善 AI 报告展示
