# 过往事件推算 · 年度批语改 AI 生成（合并大运调用）设计

> 日期：2026-05-19
> 前序：`docs/superpowers/specs/2026-05-18-past-events-narrative-evidence-anchored-design.md`
> 关联问题：截图反馈「年度批语相比大运总结明显薄」「丁亥 2007 等年份完全空白」

## 问题

`过往事件推算` 页面用户对比顶部 AI 大运总结（"申金穿害年柱亥水用神""庚寅年天克地冲、三合火局""壬辰、癸巳年官杀透干克身"）与下方算法生成的年度白话批语（"现实表现上同学比较..."）后发现：

1. **年度批语缺少具体干支事件描述**，相比大运总结明显薄
2. **部分年份完全空白**（如丁亥 2007，chips 显示"变动/竞争/贵人"但无文字）

经过 5 轮模板迭代（refactor → 收紧 default → 加 hasEvidenceAnchor → 砍 age<18 模板收尾 → 补 局势_重/大运合化 case），模板方向已收益递减：每修一个 case 引入新 case，截图问题、空白年份、过紧门控、跨年重复 —— 这是「工具用错了」的信号。

**模板的根本限制：** 无法写出"申金穿害年柱亥水用神"这种具体到干支位置 + 命理推断的句子。需要的不是更多模板变体，是命理推断能力。

## 决策

> 模板时代的失败是"说得平淡"，AI 时代的失败是"说得自信但说错"。后者对命理产品更危险，因此设计中含 3 道护栏。

1. **改造现有 `GenerateDayunSummariesStream`，让单次 AI 调用同时产出「大运总结 + 10 年卡片」**。调用次数不变（8-9 次/命盘），输出 token 约从 900 → 9000（10×），单份命盘成本约从 0.001 元 → 0.01 元（DeepSeek-V3 输出 1.1 元/M token）。
2. **每年卡片中等长度**：100-150 字，3-4 句，必须点名当年关键干支事件。
3. **`RenderYearNarrative` 整套模板逻辑废弃**，但 feature flag 控制（不立刻删代码）。
4. **AI 输出做"事实校验"** —— 引用的干支必须在算法侧 evidence 里追溯到。
5. **缓存 lazy migrate**：旧 `ai_dayun_summaries` 行没有 years 字段视为缓存失效、重生。

## 架构

### 当前数据流
```
Stage 1: GET /past-events/years/:chart_id
         → GeneratePastEventsYears (纯算法)
         → 返回 90 个年份 {年份, GanZhi, age, chips, narrative(模板), evidence}

Stage 2: SSE /past-events/dayun-summary-stream/:chart_id
         → GenerateDayunSummariesStream (AI × 9 段大运)
         → 每段返回 {themes, summary}
```

### 改造后
```
Stage 1: GET /past-events/years/:chart_id (不变)
         → GeneratePastEventsYears
         → 返回 90 个年份 {年份, GanZhi, age, chips, narrative="", evidence}

Stage 2: SSE /past-events/dayun-summary-stream/:chart_id (扩展输出)
         → GenerateDayunSummariesStream (复用现有调用)
         → 每段返回 {themes, summary, years: [{year, ganzhi, narrative}, ...]}
```

**保留的算法层职责**：信号检测、chips、命理依据、大运段信号 JSON 生成（喂给 AI）、TenGodPower 计算。

**砍掉的（feature flag 控制，4-6 周观察期后真删）**：`RenderYearNarrative` 整套（约 800 行）。

## Prompt 改造

现有 `report_service.go:1117-1136` 的 prompt 末尾输出要求扩展：

```
1. themes：2-4 个主题词（保持原规则）
2. summary：80-120 字，综合评述这 10 年整体走势
3. years：长度为 10 的数组，与算法信号 JSON 中年份顺序一一对应。每个元素：
   {"year": 数字年, "ganzhi": 干支, "narrative": "..."}

   narrative 撰写规则：
   - 100-150 字，3-4 句
   - 必须点名当年关键干支事件（如「丙火透干为食神」「流年地支冲日支」
     「白虎临运」「驿马合年支」「用神位受刑」「伏吟时柱」）
   - 结合极性写吉凶（吉应期写助力或机遇，凶应期写注意或代价）
   - 读书期年份（age<18）改写为学业/同学/家庭语义，不出现「事业/婚恋」等成人词
   - 若该年信号确实稀薄（无 hard event 信号、evidence 关键词都缺），
     narrative 可写 "" 表示该年无显著动象
   - 措辞与 summary 不重复，summary 概括十年，narrative 具体到当年

4. 严格输出 JSON：
   {"themes":[...],"summary":"...","years":[{...} ×10]}
```

**留给 AI 的自由度**：命理师风格、干支术语口径、案例措辞。
**强制约束**：JSON 结构、输出长度区间、读书期重映射、years 数组长度 = 算法侧年份数。

## 缓存与迁移

### Schema 改动

新增 migration `pkg/database/migrations/00002_dayun_summaries_years.sql`：

```sql
ALTER TABLE ai_dayun_summaries ADD COLUMN years JSONB;
COMMENT ON COLUMN ai_dayun_summaries.years IS '10 个年份卡片 [{year,ganzhi,narrative}, ...]，AI dayun 调用同时产出';
```

`ALTER TABLE ADD COLUMN` 在 PG 是 metadata-only 操作，不锁表。

### 读时机
- `years IS NULL` → 视为旧版缓存，触发 AI 重生
- `years IS NOT NULL` → 命中缓存返回

**lazy migrate 而非 backfill**：按用户访问分摊，避免一次性 token 爆发。

### Repository 改动
- `model.AIDayunSummary` 加 `Years json.RawMessage` 字段
- `repository.GetDayunSummary` 多读一列、扫到 `*json.RawMessage`
- `repository.UpsertDayunSummary` 多写一列
- `cachedDayunSummaryToStreamItem` 把 years 一起 unmarshal 推到前端

## 前端

`PastEventsPage.tsx` 数据流：

```tsx
interface DayunSummary {
  themes: string[]
  summary: string
  years?: Array<{year: number, ganzhi: string, narrative: string}>  // 新增
  loading?: boolean
  error?: string
}

interface YearEvent {
  // narrative 字段保留以保持向后兼容，但永远为 ""
  year, age, gan_zhi, dayun_index, dayun_phase, ten_god_power,
  signals, evidence_summary
}
```

渲染顺序：
1. **首加载** → Stage 1 立即返回全部 90 年的干支/年龄/chips/命理依据。**年卡片不渲染 narrative 段**。
2. **Stage 2 SSE 流** → 每段大运到位时同步推 themes + summary + years。前端按 dayun_index + year 索引把 narrative 补上对应卡片。
3. **某段大运还在生成中** → 该段大运下属年卡片显示"本段批语正在生成…"小灰字。
4. **AI 失败** → 错误展示沿用现有机制，年卡片就是没批语段。

合并查询：
```tsx
const yearNarrative = (year: YearEvent): string | undefined => {
  const ds = summaries[year.dayun_index]
  if (!ds || ds.loading || !ds.years) return undefined
  return ds.years.find(y => y.year === year.year)?.narrative || ''
}
```

## 三道护栏

### 护栏 1：AI 输出事实校验

后端收到 AI 输出后**逐年扫一遍 narrative 字符串**，做下列校验：

- **干支位置追溯**：narrative 里出现的干支组合（年柱/月柱/日柱/时柱、流年、大运）必须在该年算法侧 evidence 字符串里出现过
- **命理关键词追溯**：narrative 里出现的「用神位/忌神位/伏吟/反吟/合化/驿马/白虎/桃花」等关键词必须能在该年某个 EventSignal evidence 里追溯到
- **不匹配处理**：找到不能追溯的术语 → 丢弃这一年的 narrative，记日志 `narrative_validation_failed`，不存缓存的这一年

实现位置：`internal/service/report_service.go` 新增 `validateYearNarrative(narrative string, signals []EventSignal) (valid bool, reason string)` 函数。

**这是把"AI 自信说错"风险按下来的唯一办法。**

### 护栏 2：Feature flag

加入 `algo_config.year_narrative_mode: "ai" | "template"`，默认 `"ai"`。

实现位置：
- `pkg/bazi/algo_config.go` 加字段
- `report_service.go` 读 config，`mode == "template"` 时走旧路径（保留 RenderYearNarrative 整套不删）
- admin UI（`AlgoConfigPage`）暴露切换

`event_narrative*.go` 文件**不立刻删**。4-6 周观察期后确认 AI 模式稳定再清。

**这是真出 AI 质量问题时的一键回滚。**

### 护栏 3：Token 成本打点

`ai_dayun_summaries` 写入路径已经记录 `token_usage_log`（`module=dayun`）。增加：

- admin 报表：按天聚合 `token_usage_log` 中 `module='dayun'` 的写入次数、平均输出 token、累计估算成本
- 报表位置：`AdminTokenUsagePage`（已有）加一个新视图 / 滤镜

**上线第一周每天看一次。如果发现单日成本 > 预算 → 切回 template 模式查问题。**

## 出错处理

| 失败模式 | 检测 | 兜底 |
|---|---|---|
| AI 调用网络失败 | `aiErr != nil` | 前端显示"本段总结生成失败"，年卡片无 narrative |
| JSON 解析失败 | `json.Unmarshal` 错 | 同上 |
| years 数组长度 ≠ 算法侧 | 后端校验 `len(parsed.years) != len(inputYears)` | 同上（整段算失败） |
| years 数组年份/干支不对齐 | 后端校验 ganzhi 字段匹配 | 同上 |
| 单年 narrative 校验失败（护栏 1） | `validateYearNarrative` 返回 false | 只丢弃该年的 narrative，其他年正常存。日志记录 |
| AI 写读书期成人词汇 | prompt 强约束，目前不做后处理 | 观察期发现高频再加正则过滤 |

## 范围边界

### 不做
- 信号检测算法 `event_signals.go` —— 维持原样
- chips（`ExtractYearSignalTypes`、前端 SIGNAL_LABEL 表）—— 不动
- 命理依据 `RenderEvidenceSummary` —— 不动
- TenGodPower 计算 —— 不动
- 大运段 themes 字段规则 —— 不动
- API 路径 —— 不变
- 老的 `GeneratePastEventsStream`（ai_past_events 表）—— 暂不动，单独立题清理
- backfill 旧缓存 —— lazy migrate 取代

### 改动文件
后端（约 7 文件）：
- `pkg/database/migrations/00002_dayun_summaries_years.sql`（新）
- `pkg/bazi/algo_config.go`（feature flag 字段）
- `internal/model/admin.go`（`AIDayunSummary` 加 Years 字段）
- `internal/repository/dayun_summary_repository.go`（读写新列）
- `internal/service/report_service.go`（prompt 改造、years 解析、`validateYearNarrative`、feature flag 分支）
- `internal/handler/bazi_handler.go`（SSE 推送元素加 years）
- `internal/handler/algo_config_handler.go`（admin 暴露 feature flag）

前端（2 文件）：
- `frontend/src/pages/PastEventsPage.tsx`（DayunSummary 接口加 years、渲染合并、loading 状态）
- `frontend/src/pages/admin/AlgoConfigPage.tsx`（feature flag UI）

测试：
- 删除 `event_narrative_test.go` 大部分模板测试（保留 `TestRenderEvidenceSummary*`）
- 新增 `validateYearNarrative` 单元测试
- 新增 prompt 渲染测试（assert JSON schema 输出结构正确）

## 风险与权衡

| 风险 | 影响 | 缓解 |
|---|---|---|
| AI 信心十足说错（"用神位受刑"但实际没这信号）| 用户错信 | 护栏 1：事实校验 |
| AI 服务降智/降速/降价 | 功能整体退化 | 护栏 2：feature flag 回滚 |
| Token 成本失控 | 钱包问题 | 护栏 3：成本打点 |
| 输出非确定性 | 同命盘前后看到的不一样 | 缓存（同命盘命中后稳定）+ admin 可手动重生 |
| 回归测试覆盖丢失 | bug 不易发现 | 改测 prompt 渲染 + JSON schema 而不是具体措辞 |
| 命理师审稿模式从"审规则"变"审 narrative" | 审稿成本上升 | 接受。命理师评估的是几个代表命盘的输出质量，不是穷尽规则 |
| 4-6 周后真删模板代码时遗漏 | 死代码 | 写好 task 清单，删除时跑一遍 grep 确认无引用 |
| `evidence_summary` 命理依据可解释性变得更重要 | 用户验证 narrative 真伪靠它 | 不动 RenderEvidenceSummary，保持原样 |

## 接力

进入 `superpowers:writing-plans`，把本设计拆解为 step-by-step 实施计划，含：
- migration 文件 + repository 改动
- algo_config feature flag 字段
- prompt 改造 + JSON 解析 + 长度/对齐校验
- validateYearNarrative 函数 + 单元测试
- SSE 推送元素扩展
- 前端 DayunSummary 接口扩展 + 渲染合并 + loading 状态
- admin 配置 UI 暴露
- token 成本报表（如果不在范围内则单独立题）
