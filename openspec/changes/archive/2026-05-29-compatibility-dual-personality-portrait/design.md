## Context

合盘结果页已有「性格相处画像」区块（`PersonalityFit.tsx`），但其 self/partner 两侧内容由 `compatibilityPersonality.ts` 的 `participantPattern()` 生成，而该函数对两人喂入的是**同一份合盘四维分数**（`input.scores`），仅以 `dayGan` 字符串区分标题。因此当前并不存在真正的「各自性格」，更无对比。

可用数据：每个 `CompatibilityParticipant.chart_snapshot` 存的是 `bazi.Calculate()` 的完整 `BaziResult`（JSON 原样下发），含日主干支、各柱十神（`*_gan_shishen` / `*_zhi_shishen`）、五行分布 `wuxing`、用神/忌神、`ming_ge`（格局）、`ten_god_relation` 矩阵等。前端 TS 类型 `CompatibilityChartSnapshot` 目前仅声明了四柱 + 性别 + 五行，未声明这些字段，但数据已在网络层。

## Goals / Non-Goals

**Goals:**
- 用每个人**各自的命盘结构**确定性地推导个人性格画像（不依赖 LLM、不依赖深度报告是否生成）。
- 每人输出 5 维度画像 + 一句「日主 + 主导十神」定性。
- 推导两人「自然合的地方 / 容易冲突的地方」差异对照。
- 在历史数据 `chart_snapshot` 字段缺失时优雅降级，不报错、不空白。

**Non-Goals:**
- 不改后端、不改 API、不做数据库迁移。
- 不调整合盘评分算法或证据生成。
- 不做 LLM 个性化增强（后端 prompt + `personality_comparison` 结构化字段）——列为二期。
- 不替代单人八字报告的完整性格解读；本区块服务于「合盘对比」语境，保持克制。

## Decisions

### D1：性格来源——前端读各自 `chart_snapshot`，确定性推导
复用已下发的完整 `BaziResult`，在前端按命理规则映射。
- 备选：后端 LLM 生成 → 更个性化但需改 prompt/schema/迁移，且要等深度报告生成才有内容。**推迟到二期**，本期用确定性兜底，与现有 `hasReport` 渐进增强模式一致。

### D2：三因子映射模型
个人画像由三个结构信号组合决定，而非单因子，避免模板化：
1. **日主五行**（`day_gan` + 其五行）→ 基础气质底色（甲木进取、壬水机变、戊土厚重……）。
2. **主导十神**→ 行为驱动与风格。将十神归并为 5 类：比劫（比肩/劫财）、食伤（食神/伤官）、财（正财/偏财）、官杀（正官/七杀）、印（正印/偏印）。主导类别由各柱十神出现频次（天干十神 + 地支藏干十神）统计取最高；可优先复用 `ming_ge`（格局）作为主导信号，缺失时回退到频次统计。
3. **旺衰**（粗粒度强/中/弱）→ 同一性格的「表达强度」。`BaziResult` 无标量旺衰字段，由日主五行得分相对其余五行的占比 + 印比是否当令粗略判定；仅分 强/弱/均衡 三档。

### D3：5 维度画像的字段化输出
画像不是一段话，而是 5 条结构化维度，便于在 UI 对齐渲染与未来 LLM 增强：
- `表达/沟通方式`：食伤 vs 官印主导 → 外放直接 / 克制内敛。
- `决策与节奏`：比劫·七杀 → 果断快；正官·印 → 稳谨慎；财 → 务实权衡。
- `亲密里的核心需求`：印 → 安全感被照顾；财·食伤 → 被回应/新鲜感；官 → 被认可/承诺；比劫 → 平等与空间。
- `情绪反应`：日主五行 + 伤官/七杀强度 → 外显起伏 / 内化沉淀。
- `压力下的样子`：旺 → 硬扛对抗；弱 → 回避求援；均衡 → 先观望再应对。
另出一句 `headline` 定性：「日主X + 主导十神Y」一句话概括。

### D4：差异对照算法（互补 vs 对冲）
基于两人结构信号比对，输出「自然合的地方」与「容易冲突的地方」两组要点：
- **互补/合**：一方印旺（需被照顾）配另一方食伤/财旺（愿回应）；两日主五行相生；节奏档位接近或一快一稳形成承接。
- **冲突/对冲**：两人同为比劫/七杀主导（争强）；两日主五行相克；节奏档位相反（一急一缓）；两人皆旺（硬碰硬）。
每组各取至多 2–3 条，措辞对齐现有 `compatibilityPersonality.ts` 文案语气。

### D5：缺字段降级
历史 `chart_snapshot` 可能只含基础四柱。引擎按可用信号逐级降级：有十神/格局 → 完整画像；仅有日主 + 五行 → 用日主五行 + 五行旺衰出简化画像；连五行都无 → 退回当前基于合盘分数的通用描述（保底不空白）。

### D6：类型与渲染落点
- 在 `frontend/src/lib/api.ts` 补全 `CompatibilityChartSnapshot`：增加各柱十神、藏干十神、`yongshen`/`jishen`、`ming_ge` 等引擎所需字段（可选属性，兼容旧数据）。
- 在 `compatibilityPersonality.ts` 新增映射引擎与 `buildParticipantPortrait()` / `buildPersonalityContrast()`，重写 `buildPersonalityFitSummary` 中 self/partner 的来源，删除两边同源的旧 `participantPattern`。
- `PersonalityFit.tsx` 渲染 A 画像 / B 画像 / 差异对照三块。

### D7：画像/对照与合盘分数版本解耦（实现期补充）
现状：后端所有新合盘产出 `analysis_version = "v3.1"`，而 `CompatibilityResultPage` 仅在 `legacyScores` 存在（非 V3）时构建 `personalitySummary`，且 `SectionDeepAnalysis` 用 `{personalitySummary && <PersonalityFit/>}` 门控——导致 V3（即当前全部新数据）下「性格相处画像」整块不渲染。
决策：双方画像与差异对照**只依赖 `chart_snapshot`，与四维分数无关**，故必须与分数版本解耦，V3 与 legacy 下都渲染。
- `buildPersonalityFitSummary` 的 `scores` 改为**可选**；缺失时跳过依赖分数的 `matchType`/分数派生 fit-clash，改用 `buildPersonalityContrast` 的命盘派生结果，`matchType` 取中性默认值。
- self/partner 输入扩展为携带各自完整 `chart_snapshot`（不再只有 `dayGan`）。
- 结果页**无条件**构建并渲染画像/对照（只要存在参与者命盘）。
- 范围克制：`buildPersonalityValidationPlan` / ActionPlan 维持原 `legacyScores` 门控不变，避免把行动计划区块的可见性卷入本次改动。

## Risks / Trade-offs

- [经典映射流于模板、千篇一律] → 用三因子组合 + 分档文案，而非单因子；保留指向证据的链接增强可信度。
- [主导十神判定口径与后端引擎不一致] → 优先复用 `ming_ge` 与 `ten_god_relation` 等引擎既有产物，不自创权重；频次统计仅作回退。
- [历史数据 `chart_snapshot` 字段缺失导致崩溃] → D5 逐级降级，所有新字段在 TS 中为可选，引擎对 undefined 做兜底。
- [本区块与单人八字报告性格解读语义重叠] → 限定 5 维度 + 合盘对比语境，不展开单人完整命理叙事。
- [二期 LLM 增强可能与确定性文案口径冲突] → 维度字段化设计预留增强位，二期由 LLM 覆写 detail 而非另起结构。
