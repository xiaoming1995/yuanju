## Why

当前「双方性格画像与差异」是前端确定性公式（`compatibilityPersonality.ts` 的查表引擎：日主五行 × 主导十神 × 旺衰 → 固定文案）。它即时、免费、常驻，但有结构性局限：每人只取**单一主导十神**、5 维全是**查表模板文案**、缺乏针对具体两人关系的个性化叙事。LLM 能基于双方完整命盘生成更丰富、更贴合的性格画像与差异——这正是当初 `compatibility-dual-personality-portrait` 明确推迟的「二期」。本次将性格画像与差异改为由 AI 深度解读用 LLM 生成。

## What Changes

- **后端**：
  - 扩充 `compatibilityParticipantSummary`，在 prompt 摘要中补出每人**十神（各柱）/命格/旺衰**（数据已在 `chart_snapshot` 的完整 `BaziResult`，仅未序列化）。
  - `canonical_compatibility` prompt 增加 `personality_comparison` 输出规格：每人 `headline` + 5 维（表达沟通 / 决策节奏 / 亲密核心需求 / 情绪反应 / 压力下样子）+ 双方 `fit_points` / `clash_points`；附护栏（必须基于命盘、不下绝对断语）。
  - Go 侧 `CompatibilityStructuredReport` 增加 `personality_comparison` 结构体 + 解析与缺失兜底。
- **前端**：
  - `CompatibilityStructuredReport` TS 类型增加 `personality_comparison`。
  - 性格画像与差异**移入页面底部「AI 深度解读」块内**作为子块渲染；生成报告前页面不再显示性格，AI 报告空态提示"含双方性格画像与差异"。
  - **移除独立 SECTION 02**「双方性格画像与差异」及 `PersonalityFit` 组件在结果页的挂载。
  - **精准删除**确定性画像/对照引擎：`buildParticipantPortrait`、`buildPersonalityContrast`、`buildPersonalityFitSummary` 及其映射表（`EXPRESSION`/`DECISION`/`INTIMACY`/`EMOTION_WX`/`PRESSURE`/`GROUP_*` 等）与相关类型；保留仍被复用的 `getPersonalityMatchType`（历史页）、`buildPersonalityConsultationPreview`（合盘入口）、`getCompatibilityQuestionLabel`/`getCompatibilityStageLabel`。
  - 解耦 `buildPersonalityValidationPlan` 对已删 `PersonalityFitSummary` 的依赖（改为直接取 questionLabel/matchType）。
- **BREAKING（用户可见）**：性格画像不再常驻、不再免费秒出——只有生成 AI 深度解读后才出现。

## Capabilities

### New Capabilities
- `compatibility-llm-personality-comparison`: 由 AI 深度解读用 LLM 基于双方完整命盘生成「双方性格画像与差异」（每人 5 维画像 + headline + 合点/冲突点），作为结构化报告的一部分，渲染在 AI 解读块内。

### Modified Capabilities
<!-- 确定性版「双方性格画像」由未归档变更 compatibility-dual-personality-portrait 引入，其 spec 尚未进入 openspec/specs/，故本次不以 MODIFIED delta 形式改写；以新 capability 取代，并在 tasks 中删除其前端实现。 -->

## Impact

- 后端：`backend/internal/service/compatibility_service.go`（摘要扩充 + 结构化解析）、`backend/pkg/prompt/canonical_compatibility.go`（prompt 规格）、`backend/internal/model`（`CompatibilityStructuredReport` 结构体）；可能涉及 prompt 同步测试（`backend/pkg/prompt/sync_test.go`）。
- 前端：`frontend/src/lib/api.ts`（类型）、`frontend/src/lib/compatibilityPersonality.ts`（删引擎/解耦）、`frontend/src/components/compatibility/deep-analysis/DeepReportNarrative.tsx`（渲染性格子块）、`frontend/src/pages/CompatibilityResultPage.tsx`（移除 SECTION 02 挂载、调整 personalitySummary/validationPlan 接线）、`PersonalityFit.tsx`/`.css`（结果页不再使用 → 视情况删除或保留给他处）。
- 数据/迁移：无 schema 迁移；历史已生成报告不含 `personality_comparison` → 前端按缺失兜底（不渲染性格子块或提示重新生成）。
- 成本/延迟：每次查看性格 = 一次 LLM 调用（用户已确认接受）。
