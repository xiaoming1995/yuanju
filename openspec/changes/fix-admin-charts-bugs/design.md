## Context

管理后台「全站起盘明细」功能存在三个已确认的实现缺陷：

1. **SQL 重复行**：`ListBaziCharts` 使用 `LEFT JOIN ai_reports`，而 `ai_reports` 表对 `chart_id` 没有唯一约束（`database.go` 第65-71行），用户每次重新生成 AI 报告都会新增一行，导致该 JOIN 产生笛卡尔积，同一命盘在管理列表中出现多次。

2. **AI 报告明文展示**：`AdminChartRecord.AIResult` 对应的是 `ai_reports.content`（原始 JSON 字符串），前端用 `whiteSpace: pre-wrap` 直接渲染原始文本，对管理员无意义。实际上已有 `content_structured` (JSONB) 字段存储解析后的结构化内容，应优先使用。

3. **loading 竞态**：`AdminChartsPage` 中 `liunianLoading` 为全局 boolean，快速切换展开行时多个异步请求并发，loading UI 结果不确定。

## Goals / Non-Goals

**Goals:**
- 修复 `ListBaziCharts` SQL，保证每个命盘最多返回一行（取最新报告）
- 后端查询补充 `content_structured` 字段，模型对应新增 `AIResultStructured` 字段
- 前端起盘明细详情面板改为结构化渲染 AI 报告（性格/事业/感情/健康）
- 前端流年 loading 状态改为按 `chartId` 隔离

**Non-Goals:**
- 不修改 `ai_reports` 表唯一约束（清缓存+重生成是合法设计，保留历史多版本报告）
- 不改变分页行为（total 基于 `COUNT(*) FROM bazi_charts` 不变）
- 不涉及其他管理页面

## Decisions

### 决策 1：SQL 用子查询 LATERAL 还是 `DISTINCT ON`？

**选择**：使用 PostgreSQL `DISTINCT ON (c.id)` + `ORDER BY c.id, a.created_at DESC`

**理由**：
- `LATERAL` 语义清晰但在某些版本写法复杂
- `DISTINCT ON` 是 PostgreSQL 惯用法，一行 SQL 变更，副作用最小
- 子查询 `(SELECT content_structured FROM ai_reports WHERE chart_id=c.id ORDER BY created_at DESC LIMIT 1)` 更可读，选此方案作为备选

**最终采用**：**相关子查询**方式，在 SELECT 中内嵌 `(SELECT ... LIMIT 1)` 取最新报告，保持 FROM 子句简洁，避免 DISTINCT ON 改变排序语义。

```sql
SELECT
  c.id, c.user_id, u.email as user_email,
  c.birth_year, ...,
  COALESCE(c.yongshen, ''), COALESCE(c.jishen, ''),
  (SELECT content FROM ai_reports WHERE chart_id=c.id ORDER BY created_at DESC LIMIT 1) as ai_result,
  (SELECT content_structured FROM ai_reports WHERE chart_id=c.id ORDER BY created_at DESC LIMIT 1) as ai_result_structured,
  c.created_at
FROM bazi_charts c
LEFT JOIN users u ON c.user_id = u.id
ORDER BY c.created_at DESC
LIMIT $1 OFFSET $2
```

### 决策 2：`content_structured` 如何传递到前端？

**选择**：后端 `AdminChartRecord` 新增 `AIResultStructured *json.RawMessage` 字段，序列化为 JSON 原样传给前端。

**理由**：前端 TypeScript 已有对应类型结构（`career/romance/health/advice`），直接复用即可，无需后端再解析一次。

### 决策 3：前端 loading 状态隔离方案

**选择**：将 `liunianLoading: boolean` 改为 `liunianLoading: Record<string, boolean>`，key 为 chartId。

## Risks / Trade-offs

- **子查询性能**：每行执行两个相关子查询，数据量大时性能退化。缓解：`ai_reports(chart_id)` 已有 index（`idx_ai_reports_chart_id`），影响可控。
- **`content_structured` 为 null 时的降级**：部分历史报告可能无结构化内容，前端需做空值处理，降级显示原始 `ai_result` 文本。

## Migration Plan

纯代码修改，无 DB Schema 变更，无需迁移脚本。直接部署即可回滚（代码回退）。
