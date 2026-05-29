## 1. 现状确认

- [x] 1.1 grep 结果页与各 Section 组件，确认当前 SECTION 编号文案位置（`compat-section-kicker`），以及是否存在顶部锚点导航（指向性格/依据等 `id`）。
- [x] 1.2 确认 `PersonalityFit` 的内层 `<details>`/`compat-da-subsection-summary` 结构，规划升级为 SECTION 后如何避免标题重复。

## 2. 性格块升为独立 SECTION

- [x] 2.1 新增轻量 `SectionPersonality`（或在页面层包 `<section>`），提供 `SECTION 02` kicker + 「性格画像与差异」标题，内部复用 `PersonalityFit` 主体内容；按 D2 移除/弱化内层折叠 summary 标题以避免重复。
- [x] 2.2 在 `CompatibilityResultPage` 中，将该性格 SECTION 渲染在 `SectionBasicCharts` 与 `SectionVerdict` 之间，沿用页面已构建的 `personalitySummary`。

## 3. 深度分析容器调整

- [x] 3.1 `SectionDeepAnalysis` 移除 `PersonalityFit` 渲染及其 `personalitySummary` prop（清理由本次改动产生的孤儿 prop/类型）。
- [x] 3.2 调整 `SectionDeepAnalysis` 内部子块顺序为：`RelationshipStrategy` → `ActionPlan7d30d` → `NextStepsAndAvoid`（保持各自条件渲染逻辑）。
- [x] 3.3 将 `DeepReportNarrative`（AI 深度解读）从 `SectionDeepAnalysis` 拎出，在页面层渲染于 `EvidenceDrawer` 之后，作为全页最后一环；相应 props 由对象 prop 改为 JSX 属性直传。

## 4. SECTION 编号与导航

- [x] 4.1 更新各 SECTION 的 kicker 编号：基础盘=01、性格=02、是否合=03、深度分析=04，全量 grep `SECTION 0` 校验连续无重复。
- [x] 4.2 若存在顶部锚点导航，按新顺序/编号同步导航项文案（`id` 不变）。

## 5. 验证

- [x] 5.1 更新/新增静态 UX 测试：断言主体顺序为 基础盘 → 性格 → 是否合 → 深度分析 → 依据；断言深度分析内部为 策略→风险→下一步→AI；断言深度分析不再含性格块。
- [x] 5.2 修正任何断言旧顺序/旧编号的既有测试（仅改受本次重排影响的断言）。
- [x] 5.3 运行 `tsc -b` 与 `eslint`，确认无类型错误/孤儿引用。
- [x] 5.4 在运行的页面上肉眼确认：性格 SECTION 在分数前、编号 01–04 连续、深度分析子块新顺序、无标题重复。
