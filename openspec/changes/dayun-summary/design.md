## Context

现有 `GeneratePastEventsStream` 通过一次 AI 调用生成所有过往流年的叙述，结果以 `{years: [...]}` JSON 格式缓存在 `ai_past_events.content_structured`（JSONB）。前端 `PastEventsPage` 按 `dayun_gan_zhi` 字段将年份分组展示，大运组头部只有干支名称。

Prompt 模板存在数据库 `prompt_templates` 表（`name='past_events'`），可通过 Admin UI 更新，无需重新部署即可调整。

## Goals / Non-Goals

**Goals:**
- AI 在同一次调用中额外返回 `dayun_summaries[]`，每个大运一条总结
- 总结包含：`gan_zhi`（大运干支）、`themes[]`（主题关键词，2-4个短词）、`summary`（叙述段，80-120字）
- 前端在每个大运组顶部渲染总结块，主题词用彩色标签展示

**Non-Goals:**
- 不新增 API 端点或数据库表
- 不为大运总结单独缓存（与年份数据共用同一条 `ai_past_events` 记录）
- 不提供对历史缓存的自动升级（旧缓存不含 `dayun_summaries`，前端降级处理）

## Decisions

### 1. 在同一 AI 响应中生成 dayun_summaries，而非单独调用

**选择**：扩展现有 JSON 输出结构，让 AI 一次返回 `years[]` + `dayun_summaries[]`。

**理由**：AI 已读取全部年份数据，有足够上下文推断每段大运的整体走势；拆成两次调用会增加延迟和 token 消耗，且两份结果需要协调缓存。

**备选**：为每个大运单独请求 AI。拒绝，因为延迟成本高（N 次 SSE 流），且实现复杂度大幅上升。

### 2. Prompt 模板中传入大运列表（{{.DayunList}}）

**选择**：在 Service 层将 `result.Dayun` 序列化为简洁的文字列表（格式：`大运 甲子 12-22岁 [官星 正印]`），注入 Prompt 模板变量 `{{.DayunList}}`。

**理由**：AI 需要知道"哪些年属于哪段大运"以及"每段大运的天干十神"才能写出有质量的总结，仅靠年份数据中的 `dayun_gan_zhi` 字段信息不够充分。

**备选**：让 AI 从 `years[]` 中自行推断大运分组。拒绝，因为容易遗漏边界年份（换运年），且不知道大运干支的十神信息。

### 3. 旧缓存降级处理（无 dayun_summaries 时静默忽略）

**选择**：前端解析时，若 `dayun_summaries` 字段不存在则不渲染总结块，大运组只显示原有标题。

**理由**：已有用户的缓存不含此字段，强制失效缓存会导致所有用户重新触发 AI 生成，成本高。静默降级保证向后兼容，用户主动"重新推算"时会拿到新格式。

## Risks / Trade-offs

- [Prompt 变长导致 AI 输出质量下降] → 在 Prompt 中明确要求 `dayun_summaries` 长度（每条 summary ≤ 120 字），并给出 JSON schema 示例
- [AI 返回的 themes 词汇不一致] → Prompt 中提供参考词汇列表（从现有信号类型中衍生），但允许 AI 自由扩展
- [换运边界年份归属模糊] → DayunList 传入 `start_age` / `end_year`，Prompt 中说明按起止年份划分
