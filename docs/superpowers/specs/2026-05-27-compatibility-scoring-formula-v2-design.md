# 合盘评分公式 v2 设计

**Date:** 2026-05-27
**Status:** Approved (brainstorming → ready for implementation plan)
**Scope:** backend/pkg/bazi/compatibility.go 全体重写 + frontend/DB/prompt 同步

## 1. 背景与目的

合盘模块当前实现位于 `backend/pkg/bazi/compatibility.go`（1510 行，违反项目 500 行硬约束），采用「11 类信号源 × 4 维度（吸引/稳定/沟通/现实）× evidence 加权 + 来源贡献封顶」的复杂模型。该模型存在以下问题：

- 日柱六合/六冲/刑害在「夫妻宫」与「干支互动」两个 source 下双重计分，封顶机制不互通
- 十神互动只看 DayGan×DayGan，信号过粗；不分性别
- 五行失衡触发门槛过严（双方共同 ==0），含藏干的盘面几乎不触发
- 喜神/忌神判定依赖字符串 `Contains`，脆弱
- 评分基线 60 + 加权偏移的语义难以向用户解释
- 单文件 1510 行违反工程约束

本次重做将引擎替换为**用户给出的传统命理 100 分制公式**，结构上从「证据加权」转为「纯加分制」，从「4 维度（心理学化）」转为「4 模块（命理学化）」。

## 2. 锁定的决策（来自 brainstorming 会话）

下列每一条均经用户在分节澄清中明确确认，不再回滚：

1. **完全替换**当前 4 维度分系统；不保留并存
2. **纯加分制**：忌类（六冲/六害/相刑/自刑/相克）一律 0 分，不扣分
3. **不引入喜神/忌神**：现有 `buildFavorableElementSupportSignals` 整段废弃
4. **4 组成分字段命名**：`zodiac` / `nayin` / `day_pillar` / `eight_chars`
5. **附加结构全部保留**（`evidences` / `score_explanations` / `summary_tags` / `duration_assessment` / `consulting_assessment`），仅重接驱动数据源
6. **历史记录不动**：以 `analysis_version` 字段区分；旧记录走旧渲染路径，新查询走新算法

## 3. 算法规范

### 3.1 评分网格

| 模块 | 锚定 | 满分 | 命中条件 | 加分 |
|---|---|---|---|---|
| 合属相 | 年支 × 年支 | 50 | 六合 OR 三合（含半三合） | +50 |
| 合纳音 | 年柱纳音五行 × 年柱纳音五行 | 20 | 相生 OR 相同 | +20 |
| 合日柱 | 日柱 × 日柱 | 10 | 日支合 + (干合 OR 干生) | +10 |
|  |  |  | 日支合 + (干同 / 干克 / 干无关) | +5 |
|  |  |  | 日支不合 | 0 |
| 合八字 | 年柱对、月柱对、时柱对 | 20 | 每柱独立按合日柱规则得 0/5/10，三柱和 × 2/3（整数四舍五入） | — |
| **总分** | — | **100** | 直接相加 | — |

### 3.2 关键定义

**支合**（共享判定）：两个地支属于 六合 集合 `{子丑, 寅亥, 卯戌, 辰酉, 巳申, 午未}` 之一，OR 两个地支属于同一三合局 `{申子辰, 亥卯未, 巳酉丑, 寅午戌}` 之一且不相等（即两人合盘可能凑出的半三合）。

**干合**：天干五合 `{甲己, 乙庚, 丙辛, 丁壬, 戊癸}` 之一。

**干生**：两天干五行存在相生关系（甲乙木 ↔ 丙丁火 ↔ 戊己土 ↔ 庚辛金 ↔ 壬癸水）。

**干同**：两天干五行相同（不要求字符相同，例如「甲」与「乙」均属木，算干同）。

**纳音相生 / 相同 / 相克**：依据《六十甲子纳音表》将两人年柱（如 甲子 → 海中金）映射到 五行 之一，再判断两 五行 间关系。

### 3.3 总分分档（`overall_level`）

```
≥ 80  → high
60–79 → medium
< 60  → low
```

### 3.4 八字模块的整数运算

三柱独立打分 ∈ {0, 5, 10}，最高总和 30。归一化到 [0, 20] 采用 `(sum * 2 + 1) / 3`（整数四舍五入）：

| sum | (sum*2+1)/3 |
|---|---|
| 0 | 0 |
| 5 | 3 |
| 10 | 7 |
| 15 | 10 |
| 20 | 13 |
| 25 | 17 |
| 30 | 20 |

## 4. 输出 Schema

### 4.1 类型变更（`backend/internal/model/compatibility.go`）

```go
// 旧
type CompatibilityDimensionScores struct {
    Attraction    int `json:"attraction"`
    Stability     int `json:"stability"`
    Communication int `json:"communication"`
    Practicality  int `json:"practicality"`
}

// 新（v2 schema）
type CompatibilityDimensionScores struct {
    Zodiac     int `json:"zodiac"`       // 合属相 0–50
    Nayin      int `json:"nayin"`        // 合纳音 0–20
    DayPillar  int `json:"day_pillar"`   // 合日柱 0–10
    EightChars int `json:"eight_chars"`  // 合八字 0–20
}
```

`CompatibilityAnalysis` 新增顶层字段：

```go
type CompatibilityAnalysis struct {
    OverallScore         int                               `json:"overall_score"`    // 新增 0–100
    OverallLevel         CompatibilityLevel                `json:"overall_level"`
    DimensionScores      CompatibilityDimensionScores      `json:"dimension_scores"` // 字段名改
    Evidences            []CompatibilityEvidence           `json:"evidences"`
    ScoreExplanations    []CompatibilityScoreExplanation   `json:"score_explanations"`
    SummaryTags          []string                          `json:"summary_tags"`
    DurationAssessment   CompatibilityDurationAssessment   `json:"duration_assessment"`
    ConsultingAssessment CompatibilityConsultingAssessment `json:"consulting_assessment"`
}
```

`CompatibilityReading.AnalysisVersion` 新写入固定为 `"v2"`。

### 4.2 evidence 列表

最多 6 条（1 属相 + 1 纳音 + 1 日柱 + 最多 3 八字），最少 0 条。

| evidence_key | 触发 | weight |
|---|---|---|
| `zodiac_liuhe` | 年支六合 | 50 |
| `zodiac_sanhe` | 年支三合（含半三合） | 50 |
| `nayin_sheng` | 纳音五行相生 | 20 |
| `nayin_same` | 纳音五行相同 | 20 |
| `day_pillar_upper` | 日支合 + 干合/干生 | 10 |
| `day_pillar_lower` | 日支合 + 干同/克/无关 | 5 |
| `eight_chars_year_upper` | 年柱对 上档 | 10 |
| `eight_chars_year_lower` | 年柱对 下档 | 5 |
| `eight_chars_month_upper` | 月柱对 上档 | 10 |
| `eight_chars_month_lower` | 月柱对 下档 | 5 |
| `eight_chars_hour_upper` | 时柱对 上档 | 10 |
| `eight_chars_hour_lower` | 时柱对 下档 | 5 |

**字段约束**：
- `Polarity` 永远 `"positive"`
- `Source` 即 `Dimension`，取值 `"zodiac" | "nayin" | "day_pillar" | "eight_chars"`
- `Perspective / Actor / Target / RelatedSources` 保留 schema，新算法**永不填充**

### 4.3 score_explanations

按 4 模块各出一条，共 4 条。`Dimension` 取模块名；`NegativeFactor / NegativeEvidenceKeys` 永远为空（纯加分制）。Summary 文案按模块 × 命中状态从下面模板表生成：

| 模块 | 命中 | Summary 模板 |
|---|---|---|
| zodiac | 六合 | "双方属相 {a}/{b} 构成六合，关系基础线吸引力强。" |
| zodiac | 三合 | "双方属相 {a}/{b} 同属 {三合局名}，气场协同。" |
| zodiac | 未命中 | "双方属相 {a}/{b} 无六合/三合，关系无属相层级加成。" |
| nayin | 相生 | "{自纳音} 与 {对方纳音} 五行相生，资源/情绪流动顺。" |
| nayin | 相同 | "双方纳音同为 {五行}，本质同气。" |
| nayin | 相克 | "{自纳音} 与 {对方纳音} 五行相克，纳音层无加分。" |
| day_pillar | 上档 | "日柱 {a}/{b}：地支 {六合/三合} 且天干 {五合/相生}，亲密层结构稳。" |
| day_pillar | 下档 | "日柱 {a}/{b}：地支 {六合/三合}，天干仅同/克/无关，亲密层有基础但未达上吉。" |
| day_pillar | 未命中 | "日柱 {a}/{b} 地支不合，亲密层无加成。" |
| eight_chars | 命中 ≥ 2 柱 | "年/月/时三柱中有 {n} 柱合，外围层有支撑。" |
| eight_chars | 仅 1 柱 | "三柱中仅 {柱位名} 合，外围层支撑薄弱。" |
| eight_chars | 0 柱 | "年/月/时三柱均无合，外围层无加成。" |

`PositiveFactor` 字段填该模块命中的 evidence 的 `Title`；未命中时为空字符串。

### 4.4 summary_tags

阈值规则（最多 4 条）：

- 合属相命中 → `属相相合`
- 合纳音命中 → `纳音同气`
- 合日柱 == 10 → `日柱上吉`
- 合日柱 == 5 → `日柱次吉`
- 合八字 ≥ 14 → `八字承接好`
- 总分 ≥ 80 → `上吉合盘`
- 总分 < 60 且 4 模块全 0 → `合盘无加成`

## 5. consulting_assessment / duration_assessment / stage_risks 重接

### 5.1 relationship_type 分类

**按下列优先级顺序逐条匹配，命中第一条即返回**（短路求值）：

1. `total >= 80` → `"高契合型"`
2. `zodiac == 50 AND day_pillar >= 5` → `"亲密层稳固型"`
3. `zodiac == 50` → `"属相吸引型"`
4. `day_pillar >= 5 OR eight_chars >= 14` → `"亲密外围支撑型"`
5. 否则 → `"合盘无加成"`

### 5.2 decision_advice

| 总分 | recommendation | verdict |
|---|---|---|
| ≥ 80 | `continue` | "适合继续推进" |
| 60–79 | `observe` | "建议谨慎观察" |
| < 60 | `caution` | "不宜过早重投入" |

**confidence**：
- 命中模块数 ≥ 3 → `high`
- 命中模块数 1–2 → `medium`
- 命中模块数 0 → `low`

**Conditions / DoNext / Avoid**：按 recommendation 三档切换 3 套模板（4 句话每套），文案保留现行表达。

### 5.3 duration_assessment

| 窗口 | high 阈值 | medium 阈值 | low |
|---|---|---|---|
| 3 个月 | zodiac == 50 AND nayin == 20 | zodiac == 50 OR nayin == 20 | 都未命中 |
| 1 年 | day_pillar == 10 OR (day_pillar == 5 AND zodiac == 50) | day_pillar >= 5 | day_pillar == 0 |
| 2 年+ | eight_chars >= 14 AND day_pillar >= 5 | eight_chars >= 7 OR day_pillar == 10 | 否则 |

`overall_band` 由长期窗口决定：
- 长期 high → `long_term`
- 长期 medium → `medium_term`
- 长期 low → `short_term`

`Summary` 按 (短期, 长期) 组合给 4 套模板（同现行 case 数量）。`Reasons` 取前 3 条命中 evidence 的 `Title: Detail`。

### 5.4 stage_risks

保留 3 个窗口结构。`RiskLevel` 直接复用 §5.3 的三档。`MainRisk / Trigger / Advice` 用 (窗口 × level) = 9 套硬编码模板字符串——**具体文案由实现 plan 阶段产出**（参照现行 `buildCompatibilityStageRisk` 的现有文案风格，主语和触发条件改成新 4 模块语境）。`EvidenceKeys` 按窗口挂相关模块的 key：
- 3 个月 → 取 zodiac / nayin 的 evidence_key
- 1 年 → 取 day_pillar 的 evidence_key
- 2 年+ → 取 eight_chars 的 evidence_key

### 5.5 relationship_strategy

4 句 `Communication / Conflict / Reality / Boundary` 按 recommendation 三档切换 3 套模板共 12 句——**具体文案由实现 plan 阶段产出**（参照现行 `CompatibilityRelationshipStrategy` 现有文案风格调整措辞）。

### 5.6 claim_evidence_links

`relationship_main_judgement` 的 `EvidenceKeys` 取所有命中模块的 evidence_key（≤ 6 条）。Reasoning / Caveat 保留。

## 6. AI Prompt 改造

文件：`backend/pkg/prompt/canonical_compatibility.go`

**改动点**：

1. `ScoresJSON` 字段名换为 `{zodiac, nayin, day_pillar, eight_chars}`，新增顶层 `overall_score`
2. Prompt 模板里「吸引力 / 稳定度 / 沟通修复 / 现实磨合」措辞 → 「合属相 / 合纳音 / 合日柱 / 合八字」
3. 增加 6–8 行**评分规则说明段**，让 LLM 理解 4 个数字代表的命理学含义与上下界
4. `EvidenceGroupsJSON` 的分组键从 11 种 source 缩为 4 模块
5. 描述 evidence 语境时去掉"正面/负面"，统一为「命中 / 未命中」

`CompatibilityPromptData` 结构保留（字段名不变，仅内容 JSON 改）。

## 7. Frontend 改造

### 7.1 类型（`frontend/src/lib/api.ts`）

```ts
type CompatibilityDimensionScores = {
  zodiac: number; nayin: number;
  day_pillar: number; eight_chars: number;
}

type CompatibilityReading = {
  // ...
  overall_score: number;                  // 新增
  analysis_version: 'v1' | 'v2';
  dimension_scores: CompatibilityDimensionScores;
}
```

### 7.2 渲染分支（`CompatibilityResultPage.tsx`）

- 读 `reading.analysis_version`：
  - `v2` → 走新组件 `ScoreOverviewV2`，按 4 模块渲染
  - `v1` 或缺失 → **保留**现行 `ScoreOverview` + dimension/evidence/consulting 渲染路径作为兼容分支
- 顶部摘要新增「总分」大数字 + level 徽章（仅 v2 路径）

### 7.3 dimensionHint 文案

```ts
const dimensionHintV2 = {
  zodiac:      "属相（年支）层：六合/三合 命中即满分 50",
  nayin:       "纳音五行：相生/相同 命中即满分 20",
  day_pillar:  "日柱（亲密层）：支合 + 干合/生 满分 10",
  eight_chars: "年/月/时三柱：外围承接，最高 20",
}
```

### 7.4 历史列表（`CompatibilityHistoryPage.tsx`）

mini-renderer 也按 version 分两路；v2 路径显示总分大数字与 level。

## 8. DB / 历史兼容

- **不做 DDL 结构改动**：`dimension_scores` 是 JSONB，键名变化不需要 ALTER TABLE
- **新 migration** `00011_compatibility_v2_analysis.sql`：仅写 `COMMENT ON COLUMN ... analysis_version` 标注 v1/v2 含义
- **新写入**：`compatibility_repository.go` INSERT 时硬编码 `analysis_version = 'v2'`
- **旧记录**：保留原状，前端按 version 切渲染
- **不做** lazy-recompute、不做批量 backfill、不做删除

## 9. 拆分后的文件布局

`backend/pkg/bazi/` 下：

```
compatibility.go               # 类型 + AnalyzeCompatibility 主入口（~120 行）
compatibility_scoring.go       # 4 模块打分函数（~300 行）
compatibility_nayin.go         # 六十甲子纳音表 + 纳音五行（~100 行）
compatibility_evidence.go      # 命中→evidence + score_explanations（~150 行）
compatibility_assessment.go    # consulting / duration / strategy（~200 行）
compatibility_test.go          # 重写测试（~400 行）
```

每个文件均在项目 500 行硬约束之下。

## 10. 测试策略

- **打分函数单元测试**：4 模块各 5–8 个 case
  - 合属相：六合命中 / 三合（半三合）命中 / 既不合也不冲 / 六冲（应得 0）
  - 合纳音：相生 / 相同 / 相克
  - 合日柱：上档（干合+支合）/ 上档（干生+支合）/ 下档（干同+支合）/ 下档（干克+支合）/ 干合但支不合（应得 0）
  - 合八字：3 柱全上档 / 仅 1 柱命中 / 全 0 / 归一化边界
- **集成测试**：8–10 个完整合盘 case，断言 `OverallScore` / `OverallLevel` / 4 模块分 / 命中 evidence_keys
- **回归测试**：`compatibility_test.go` 整体重写。原 `TestAnalyzeCompatibility_ReturnsCoreShape` 保留但断言新字段
- **不做**：性能测试（O(1)）；frontend e2e（项目无 e2e 框架）

## 11. OpenSpec 归档

现行两份 spec 在新算法下语义被推翻：

- `openspec/specs/compatibility-explainable-compatibility-scoring/spec.md`
- `openspec/specs/compatibility-depth-signal-engine/spec.md`

处理：在新 change `openspec/changes/compatibility-scoring-formula-v2/` 的 `proposal.md` 中列明待废弃，实现完成后通过 `/opsx-archive` 归档。新建 change 目录包含：

```
openspec/changes/compatibility-scoring-formula-v2/
  proposal.md
  design.md            # 本文件的精简版
  tasks.md             # 由 /opsx-apply 工作流消费
  specs/
    compatibility-scoring-formula/spec.md   # 新算法的 Requirement 定义
```

## 12. 影响文件清单（估算改动量）

| 文件 | 改动 |
|---|---|
| `backend/pkg/bazi/compatibility.go` | 1510 → ~120 行（仅入口） |
| `backend/pkg/bazi/compatibility_scoring.go` | **新增** ~300 行 |
| `backend/pkg/bazi/compatibility_nayin.go` | **新增** ~100 行 |
| `backend/pkg/bazi/compatibility_evidence.go` | **新增** ~150 行 |
| `backend/pkg/bazi/compatibility_assessment.go` | **新增** ~200 行 |
| `backend/pkg/bazi/compatibility_test.go` | 634 → ~400 行重写 |
| `backend/internal/model/compatibility.go` | DimensionScores 字段重命名 + OverallScore 新增 |
| `backend/internal/repository/compatibility_repository.go` | INSERT 标 `analysis_version='v2'` |
| `backend/internal/service/compatibility_service.go` | I/O 字段同步 |
| `backend/pkg/prompt/canonical_compatibility.go` | prompt 模板重写 |
| `backend/pkg/database/migrations/00011_compatibility_v2_analysis.sql` | 新增（仅 COMMENT） |
| `frontend/src/lib/api.ts` | 类型重定义 + overall_score |
| `frontend/src/pages/CompatibilityResultPage.tsx` | version 分两路渲染；ScoreOverviewV2 |
| `frontend/src/pages/CompatibilityHistoryPage.tsx` | 列表按 version 分两路 |
| `openspec/changes/compatibility-scoring-formula-v2/` | 新增整套 change 目录 |

## 13. 范围之外（明确不做）

- 不引入「喜神/忌神」逻辑（已确认）
- 不引入「双生」（五行相生/相同）单独得分（已确认）
- 不计算「神煞」「十神互动」「夫妻宫单独项」「干支互动多柱组合」（统一归并到 4 模块）
- 不对旧记录做批量重算或迁移
- 不调整 OpenSpec 已归档的 `compatibility-depth-signals` change
- 不修改 `bazi-compatibility-match` 章节相关的 input form / participant schema
- 不引入性能优化（O(1) 算法）
- 不写 e2e / 视觉回归测试

## 14. 风险与缓解

| 风险 | 缓解 |
|---|---|
| 旧记录在前端历史页与新记录混排，视觉割裂 | version 分支渲染，旧记录卡片保留旧样式；新记录卡片显示总分大数字 |
| LLM prompt 改后 AI 报告语言风格变化 | 在 prompt 评分规则段落明确解释 4 模块语义；保留 prompt 其余结构 |
| 纯加分制下大量盘面得分集中在 0–30 区间，过早 caution | 总分阈值（≥80 high / <60 low）在实测后可调，仅需改 1 处常量 |
| 新算法测试覆盖不足 | 集成测试至少覆盖：六合上吉、三合上吉、全冲（应得 0）、纳音相克 + 其余命中、日柱上下档边界 |
| 纳音表硬编码 60 条易写错 | 单元测试覆盖全部 60 个 ganzhi，与 `lunar-go` 提供的 `GetYearNaYin()` 输出对齐校验 |
