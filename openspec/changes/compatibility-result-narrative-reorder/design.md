## Context

合盘结果页 `CompatibilityResultPage` 当前渲染顺序：StickyHeader（吸顶）→ `SectionBasicCharts`(SECTION 01 双方基础盘) → `SectionVerdict`(SECTION 02 是否合/分数) → `SectionDeepAnalysis`(SECTION 03 深度分析) → `EvidenceDrawer`(关键依据)。

`SectionDeepAnalysis` 是一个容器，内部按序渲染 5 个 `<details>` 折叠子块：`PersonalityFit`(双方性格画像与差异) → `ActionPlan7d30d`(阶段风险与时段) → `RelationshipStrategy`(关系经营策略, 条件渲染) → `NextStepsAndAvoid`(下一步/避免) → `DeepReportNarrative`(AI 深度解读)。

各顶级 SECTION 用统一头部：`<div class="compat-section-kicker">SECTION 0N</div>` + `<h2 class="compat-section-title">标题</h2>`。性格块目前不是顶级 SECTION，而是 `compat-da-subsection`（`<details open>` + `compat-da-subsection-summary`）。

`personalitySummary` 已在页面层用 `buildPersonalityFitSummary` 构建，作为 prop 传入 `SectionDeepAnalysis`。

## Goals / Non-Goals

**Goals:**
- 结果页主体按叙事弧线重排：认识两人 → 揭晓合不合 → 怎么相处 → 想深挖。
- 性格画像与差异升为独立顶级 SECTION，排在基础盘与分数之间。
- SECTION 重新编号 01/02/03/04；「深度分析」内部子块改为 策略→风险→下一步→AI。

**Non-Goals:**
- 不改 StickyHeader（仍吸顶显示总分+结论）。
- 不拆分 `PersonalityFit`（各自画像与差异对照仍同处一块）。
- 不改各模块内部内容、数据来源、评分算法、AI 报告生成。
- 不动后端、API、数据库。

## Decisions

### D1：重排在页面层完成，模块整块移动
顺序调整在 `CompatibilityResultPage` 的渲染树与 `SectionDeepAnalysis` 的子块排列上完成，模块内部不动。
- `SectionDeepAnalysis` 不再渲染 `PersonalityFit`，其 `personalitySummary` prop 移除（或保留但不用——优先移除以免留孤儿 prop）。
- 「深度分析」内部子块顺序改为：`RelationshipStrategy` → `ActionPlan7d30d` → `NextStepsAndAvoid` → `DeepReportNarrative`。

### D2：性格块如何升为顶级 SECTION（推荐方案）
性格块当前是 `<details>` 子块，要与其它 SECTION 头部风格一致。
- **推荐**：新增轻量包装 `SectionPersonality`（或在页面层直接包一层 `<section>`），提供 `SECTION 02` kicker + 「性格画像与差异」标题，内部复用现有 `PersonalityFit` 的主体内容。为避免 SECTION 标题与 `PersonalityFit` 自身 `<details>` summary（也叫「双方性格画像与差异」）**标题重复**，升级后该块以常开 section 呈现，移除/弱化其内层折叠 summary。
- 备选：保留 `PersonalityFit` 原 `<details>` 不变，仅在外层加 SECTION 头——会出现标题重复，不推荐。

### D3：SECTION 编号
统一改 kicker 文案：`SectionBasicCharts`=01、性格=02、`SectionVerdict`=03、`SectionDeepAnalysis`=04。编号是各组件内的静态文案，逐处改。

### D4：锚点导航同步
若结果页顶部存在快速跳转锚点（指向性格/依据等 `id`），`id` 本身不变（按内容锚定），但导航项的**顺序/编号文案**需与新结构对齐。实现时先确认是否存在该导航，存在则同步。

## Risks / Trade-offs

- [性格升级为 SECTION 后标题与内层 `<details>` summary 重复] → D2 推荐方案：常开 section + 去掉内层 summary 标题。
- [叙事上"揭晓"被提前——StickyHeader 已把分数吸顶，正文却把分数放到 SECTION 03] → 本次按用户选择（温和版）不动 StickyHeader；如后续觉得剧透影响体验，可另开 change 处理吸顶内容。
- [SECTION 编号散落在多个组件，漏改导致编号跳号] → 实现时全量 grep `SECTION 0` 校验 01–04 连续无重复。
- [锚点导航顺序与新结构不一致] → D4 先确认其存在性再同步；静态 UX 测试若断言旧顺序需一并更新。
