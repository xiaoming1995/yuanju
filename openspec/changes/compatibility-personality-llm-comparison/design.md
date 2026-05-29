## Context

「双方性格画像与差异」目前是前端确定性引擎（`compatibilityPersonality.ts`）：每人抽 3 信号（日主五行 / 主导十神 / 旺衰）→ 5 维查表文案 + headline；两人信号配对 → 合点/冲突点。它常驻、免费、即时，但每人只取单一主导十神、文案为固定查表、无个性化叙事。

后端 AI 深度解读已有完整链路：
- prompt `canonical_compatibility.go` 接收双方**命盘摘要**（`compatibilityParticipantSummary`：四柱 + 日主 + 五行五数 + 用神 + 忌神，**不含十神/命格/旺衰**）、四模块分数、证据等，输出结构化 JSON。
- Go `CompatibilityStructuredReport`（`model/compatibility.go:181`）承接，解析方式**宽松**：`json.Unmarshal` 失败即跳过、成功则重新 marshal 落库（`compatibility_service.go:343`）。
- 报告 opt-in（用户点「生成深度解读」），渲染在页面底部 `DeepReportNarrative`。

用户已决策：性格**完全搬进 AI 报告由 LLM 生成**，删除确定性公式；性格渲染在 AI 解读块内，移除独立 SECTION 02；生成前页面不显示性格。

## Goals / Non-Goals

**Goals:**
- LLM 基于双方完整命盘生成每人 5 维性格画像 + headline + 合点/冲突点，作为结构化报告字段。
- 前端在 AI 解读块内渲染该字段；移除确定性 SECTION 02 与引擎。
- 旧报告（无该字段）安全降级，不报错。
- 精准删除，不误删仍被复用的辅助函数。

**Non-Goals:**
- 不改评分算法、证据生成、合盘维度（zodiac/nayin/day_pillar/eight_chars）。
- 不改报告的 opt-in 生成方式与页面位置（仍在底部）。
- 不为旧报告做数据回填/迁移。
- 不动合盘入口页的性格咨询预览、历史页的 matchType（保留）。

## Decisions

### D1：personality_comparison 结构对齐现有前端形状
新增字段结构沿用当前 `ParticipantPortrait` / `PersonalityPoint` 形状，使前端渲染是「同形状换数据源」而非重写：
```jsonc
"personality_comparison": {
  "self":    { "headline": "...", "dimensions": [ { "key": "expression|decision|intimacy|emotion|pressure", "label": "...", "detail": "..." } ] },
  "partner": { "headline": "...", "dimensions": [ ... ] },
  "fit_points":   [ { "title": "...", "detail": "..." } ],
  "clash_points": [ { "title": "...", "detail": "..." } ]
}
```
- 维度 key 固定为 5 个（表达/决策/亲密/情绪/压力），便于前端稳定渲染与降级。
- *备选*：自由文本段落——否决，丢失结构化、无法稳定分维度展示。

### D2：prompt 摘要扩充而非新增 prompt 输入块
在 `compatibilityParticipantSummary` 现有摘要字符串后追加 `十神=...；命格=...；旺衰=...`（数据取自已 Unmarshal 的 `BaziResult`）。
- 不新增独立 `{{.SelfChartShiShen}}` 占位符，减少 prompt 模板与 `CompatibilityPromptData` 字段改动面。
- *备选*：新增结构化 charts JSON 占位——改动更大，本期不需要。

### D3：复用宽松解析，旧报告零兼容成本
`CompatibilityStructuredReport` 加 `PersonalityComparison *CompatibilityPersonalityComparison`（指针，omitempty）。现有 `json.Unmarshal` 宽松解析：旧报告无该字段 → nil → 前端不渲染性格子块。无需迁移。

### D4：前端精准删除边界
- **删**：`buildParticipantPortrait`、`buildPersonalityContrast`、`buildPersonalityFitSummary` 及映射表 `EXPRESSION/DECISION/INTIMACY/EMOTION_WX/PRESSURE/GROUP_LABEL/GROUP_PERSONA/GROUP_PACE`、辅助 `dominantGroup/chartStrength/chartSignals/getDayWuxing/collectTenGods/fullDimensions/simplifiedDimensions/genericDimensions`、类型 `ParticipantPortrait/PersonalityDimension/PersonalityFitSummary`（若不再被任何保留函数引用）、常量 `TEN_GOD_GROUP/GAN_WUXING/SHENG/KE/GENERATED_BY/WUXING_KEY`。
- **保留**（仍被复用，consumer 已核实）：`getPersonalityMatchType`（历史页）、`buildPersonalityConsultationPreview`（入口页）、`getCompatibilityQuestionLabel`/`getCompatibilityStageLabel`。
- **解耦**：`buildPersonalityValidationPlan` 当前签名吃 `personality: PersonalityFitSummary` 并用 `personality.questionLabel`/`.matchType`。改为直接接收 `questionLabel`/`matchType`（或在结果页用 `getCompatibilityQuestionLabel` + `getPersonalityMatchType` 计算后传入），切断对已删 summary 的依赖。

### D5：渲染位置与降级
- 在 `DeepReportNarrative` 结构化分支内，于 summary/dimensions 附近新增性格子块（复用 `PersonalityFit.tsx` 的 PortraitCard / 合点冲突点列表 UI，改为吃 `structuredReport.personality_comparison`）。
- `personality_comparison` 缺失/为空 → 不渲染该子块。
- AI 报告**空态**文案补一句"含双方性格画像与差异"，提示生成后可见。
- 结果页移除 `<PersonalityFit>` 的 SECTION 02 挂载；`PersonalityFit.tsx`/`.css` 若被复用进 DeepReportNarrative 则保留并改造，否则删除。

## Risks / Trade-offs

- [LLM 输出结构不稳定 / 维度缺漏] → prompt 固定 5 维 key + 明确 JSON 规格；前端按维度数组渲染，缺维度则跳过；整体缺失则不渲染子块。
- [旧报告无性格字段] → D3 指针 + 宽松解析，nil 即降级。
- [用户感知倒退：性格不再常驻] → 已与用户确认接受；AI 报告空态明确提示性格在生成后出现。
- [误删仍被复用的函数] → D4 已用 grep 核实 consumer；tasks 含删除后全量 tsc/eslint/测试校验。
- [验证计划链路悬空] → D4 显式解耦 `buildPersonalityValidationPlan`。
- [prompt 同步测试] → 改 `canonical_compatibility.go` 需同步 `prompt/sync_test.go` 的快照/校验。

## Open Questions

- `PersonalityFit.tsx`/`.css` 是改造复用还是删除重写？倾向复用其 PortraitCard/列表子组件（UI 已成熟），仅替换数据源与外层挂载——实现时定。
- LLM 5 维 `label` 由 prompt 固定中文标签，还是前端按 key 映射固定标签？倾向前端按 key 映射（保证标签一致、防 LLM 漂移），prompt 只产 `detail`——design 倾向此法，tasks 标注。
