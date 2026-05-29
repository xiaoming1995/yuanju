## 1. 前端类型补全

- [x] 1.1 在 `frontend/src/lib/api.ts` 扩展 `CompatibilityChartSnapshot`，新增引擎所需的可选字段：各柱天干十神（`*_gan_shishen`）、各柱地支藏干十神（`*_zhi_shishen`）、`yongshen`/`jishen`、`ming_ge`/`ming_ge_desc`、`ten_god_relation`（按需）。所有新增字段标记为可选以兼容历史数据。
- [x] 1.2 确认 `CompatibilityChartSnapshot` 与后端 `BaziResult` 的字段名一一对应（snake_case），不引入前端独有命名。

## 2. 经典映射引擎（compatibilityPersonality.ts）

- [x] 2.1 实现「主导十神」判定：优先用 `ming_ge`/`ten_god_relation`，缺失时按各柱十神（含藏干）频次统计，归并为 比劫/食伤/财/官杀/印 五类。
- [x] 2.2 实现「粗粒度旺衰」判定（强/弱/均衡）：由日主五行得分相对其余五行占比 + 印比当令粗略推导。
- [x] 2.3 实现「日主五行 → 基础气质」映射表（10 日主或 5 行底色）。
- [x] 2.4 实现 `buildParticipantPortrait(chart)`：组合日主五行 × 主导十神 × 旺衰，输出 5 条维度（表达沟通/决策节奏/亲密核心需求/情绪反应/压力下样子）+ 一句 headline 定性。
- [x] 2.5 实现缺字段降级路径：仅日主+五行 → 简化画像；连五行都无 → 退回基于合盘分数的通用描述（保底不空白）。

## 3. 差异对照推导

- [x] 3.1 实现 `buildPersonalityContrast(selfChart, partnerChart)`：基于十神互补性、日主五行生克、节奏档位差、双方旺衰，输出「自然合的地方」与「容易冲突的地方」两组，每组至多 3 条。

## 4. 接入与清理

- [x] 4.1 重写 `buildPersonalityFitSummary`：`scores` 改为可选；`selfPattern`/`partnerPattern` 替换为 `selfPortrait`/`partnerPortrait`（分别调用 `buildParticipantPortrait`，各自命盘）；fit/clash 改用 `buildPersonalityContrast` 的命盘派生结果；`matchType` 在缺分数时取中性默认。输入扩展为携带各自完整 `chart_snapshot`。
- [x] 4.2 删除两边同源的旧 `participantPattern` 实现及其不再使用的引用（仅清理由本次改动产生的孤儿）。
- [x] 4.3 更新 `PersonalityFitSummary` 类型与相关接口以承载双画像（5 维度）与差异对照。
- [x] 4.4 解耦渲染门控：`CompatibilityResultPage.tsx` 无条件构建画像/对照（不再以 `legacyScores` 门控），`scores` 仅在 legacy 时传入；`buildPersonalityValidationPlan`/ActionPlan 维持原 `legacyScores` 门控不变。

## 5. 渲染（PersonalityFit.tsx）

- [x] 5.1 在 `PersonalityFit.tsx` 渲染「A 画像 / B 画像」两块，每块展示 headline + 5 条维度。
- [x] 5.2 渲染「差异对照」块：自然合的地方 / 容易冲突的地方。
- [x] 5.3 在 `PersonalityFit.css` 补充双画像与差异对照的样式，对齐现有 compat-da 区块风格与移动端布局。
- [x] 5.4 在 `SectionDeepAnalysis.tsx` 解除 `{personalitySummary && ...}` 门控，使 V3 下也渲染 `PersonalityFit`；确认每位参与者的 `chart_snapshot` 正确传入数据源。

## 6. 验证

- [x] 6.1 在 `frontend/tests/` 补充映射引擎单测：不同日主/主导十神/旺衰组合产出不同画像；两人不同命盘产出不同输出。
- [x] 6.2 补充降级测试：`chart_snapshot` 仅含基础四柱时不报错、输出简化画像。
- [x] 6.3 补充对照测试：印旺×食伤旺判为合点；双比劫/七杀或日主相克判为冲突点。
- [x] 6.4 在 `hasReport` 为否时验证双画像与差异对照仍正常展示。
- [x] 6.5 运行前端 lint/test，确认无类型错误与回归。
