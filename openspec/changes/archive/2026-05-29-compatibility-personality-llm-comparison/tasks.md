## 1. 后端：prompt 摘要扩充

- [x] 1.1 在 `compatibilityParticipantSummary`（`compatibility_service.go`）的摘要字符串后追加 `十神=...；命格=...；旺衰=...`，数据取自已 Unmarshal 的 `BaziResult`（十神各柱、`ming_ge`、旺衰）。确认 `BaziResult` 含这些字段；旺衰若无现成字段，用与前端一致的口径（比劫+印五行占比）或命盘已有的旺衰判断。
- [x] 1.2 验证：构造一条 reading，打印/断言生成的 `SelfChartSummary` 含十神/命格/旺衰。

## 2. 后端：prompt 输出规格 + 护栏

- [x] 2.1 在 `canonical_compatibility.go` 的输出 JSON 规格中新增 `personality_comparison`：`self`/`partner` 各 `headline` + `dimensions`（5 维，key 固定 expression/decision/intimacy/emotion/pressure），`fit_points`/`clash_points`（title+detail）。
- [x] 2.2 prompt 文案加护栏：性格判断必须基于命盘十神/五行/旺衰，不下绝对断语、不输出确定事件日期；5 维各一句、克制直断。
- [x] 2.3 同步 `backend/pkg/prompt/sync_test.go`（及任何 canonical 快照/校验测试），使其通过。

## 3. 后端：结构体 + 解析

- [x] 3.1 `model/compatibility.go` 增加 `CompatibilityPersonalityComparison`（self/partner: headline + dimensions[]；fit_points[]；clash_points[]）与维度/点结构体；在 `CompatibilityStructuredReport` 加 `PersonalityComparison *CompatibilityPersonalityComparison json:"personality_comparison,omitempty"`。
- [x] 3.2 确认现有宽松解析（`compatibility_service.go:343` 的 `json.Unmarshal` + 重新 marshal 落库）自动覆盖新字段；缺失即 nil。无需新增解析分支。
- [x] 3.3 后端测试：含/不含 personality_comparison 的 LLM 输出都能正确解析（含字段→落库、缺字段→nil 不报错）。

## 4. 前端：类型 + 渲染

- [x] 4.1 `api.ts` 的 `CompatibilityStructuredReport` 增加可选 `personality_comparison`（self/partner: headline + dimensions[{key,label,detail}]；fit_points[]；clash_points[]）。
- [x] 4.2 在 `DeepReportNarrative.tsx` 结构化分支内新增性格子块：复用 `PersonalityFit.tsx` 的 PortraitCard 与合点/冲突点列表 UI，数据源改为 `structuredReport.personality_comparison`；5 维 `label` 由前端按 key 映射固定标签（防 LLM 漂移），`detail` 取自 LLM。
- [x] 4.3 `personality_comparison` 缺失/空 → 不渲染该子块。
- [x] 4.4 AI 报告空态（`compatibility-report-empty`）文案补"含双方性格画像与差异"。

## 5. 前端：移除 SECTION 02 + 删除确定性引擎

- [x] 5.1 `CompatibilityResultPage.tsx`：移除 `<PersonalityFit>` 的 SECTION 02 挂载与其 `buildPersonalityFitSummary` 调用/import。
- [x] 5.2 解耦 `buildPersonalityValidationPlan`：改为接收 `questionLabel`/`matchType`（在结果页用 `getCompatibilityQuestionLabel` + `getPersonalityMatchType` 计算后传入），切断对已删 `PersonalityFitSummary` 的依赖。
- [x] 5.3 删除 `compatibilityPersonality.ts` 中的确定性引擎：`buildParticipantPortrait`/`buildPersonalityContrast`/`buildPersonalityFitSummary`、映射表（`EXPRESSION`/`DECISION`/`INTIMACY`/`EMOTION_WX`/`PRESSURE`/`GROUP_*`）、辅助函数（`dominantGroup`/`chartStrength`/`chartSignals`/`getDayWuxing`/`collectTenGods`/`fullDimensions`/`simplifiedDimensions`/`genericDimensions`）、不再被引用的类型与五行常量。保留 `getPersonalityMatchType`/`buildPersonalityConsultationPreview`/`getCompatibilityQuestionLabel`/`getCompatibilityStageLabel`。
- [x] 5.4 `PersonalityFit.tsx`/`.css`：若 4.2 复用其子组件则保留改造为 DeepReportNarrative 的内部组件，否则删除；同步更新结果页相关 SECTION 编号（若仍以 01/02/03 标号，性格移除后重排为连续编号）。

## 6. 验证

- [x] 6.1 后端 `go build` + 相关单测（service/prompt）通过。
- [x] 6.2 前端 `tsc -b` + eslint 无错误、无孤儿引用（对应 spec「删除后无悬空引用」）。
- [x] 6.3 更新/修正受影响的前端测试：移除断言确定性画像/SECTION 02 的用例；新增断言 personality_comparison 类型与 DeepReportNarrative 渲染；保留函数（matchType/咨询预览）测试仍通过。
- [~] 6.4 真实链路验证：后端 prompt/结构体/解析由 Go 单测覆盖（含/缺字段、摘要含十神/命格/旺衰）；前端渲染由 Vite SSR 用 mock LLM 数据验证（5 维 + headline + 合点/冲突点，key→固定 label，缺失→不渲染）。**待用户在浏览器点「生成深度解读」做完整 LLM 往返确认**（需 API key + 一次 LLM 调用）。
- [x] 6.5 全量前端测试对比基线无新增失败。
