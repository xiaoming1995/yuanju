# Compatibility Scoring Formula v3 — Design

详见外部完整设计文档：[`docs/superpowers/specs/2026-05-27-compatibility-scoring-formula-v2-design.md`](../../../docs/superpowers/specs/2026-05-27-compatibility-scoring-formula-v2-design.md)。本文档为 OpenSpec change 的提炼版本。

## Why this design
见 [`proposal.md`](./proposal.md)。

## 算法规范

### 评分网格

| 模块 | 锚定 | 满分 | 命中条件 |
|---|---|---|---|
| 合属相 (zodiac) | 年支 × 年支 | 50 | 六合 OR 三合（含半三合）|
| 合纳音 (nayin) | 年柱纳音五行 × 年柱纳音五行 | 20 | 相生 OR 相同 |
| 合日柱 (day_pillar) | 日柱 × 日柱 | 10 | 支合 + 干合/干生 → 10；支合 + 其他 → 5；支不合 → 0 |
| 合八字 (eight_chars) | 年/月/时三柱对 | 20 | 每柱按合日柱规则 0/5/10，三柱和 × 2/3（整数四舍五入） |

总分 = 4 模块直接相加 ∈ [0, 100]。

### 总分分档（OverallLevel）

- ≥ 80 → high
- 60–79 → medium
- < 60 → low

### 八字归一化整数运算

`normalizeEightCharsSum(sum)` = `(sum*2 + 1) / 3`（整数除法，四舍五入）：

| sum | 归一化 |
|---|---|
| 0 | 0 |
| 5 | 3 |
| 10 | 7 |
| 15 | 10 |
| 20 | 13 |
| 25 | 17 |
| 30 | 20 |

## 输出 Schema

`CompatibilityDimensionScores`：

```go
type CompatibilityDimensionScores struct {
    Zodiac     int `json:"zodiac"`       // 0–50
    Nayin      int `json:"nayin"`        // 0–20
    DayPillar  int `json:"day_pillar"`   // 0–10
    EightChars int `json:"eight_chars"`  // 0–20
}
```

`CompatibilityAnalysis` 新增：
- `OverallScore int json:"overall_score"` —— 0–100

## consulting / duration / strategy 改写

- `relationshipType`：5 档优先级链（高契合型 / 亲密层稳固型 / 属相吸引型 / 亲密外围支撑型 / 合盘无加成）
- `decision_advice`：3 档（continue / observe / caution）+ 3 档 confidence（high / medium / low），按 total + hits 计算
- `duration_assessment`：3 窗口（3 个月 / 1 年 / 2 年+）独立阈值
- `stage_risks`：9 套模板（3 窗口 × 3 level）
- `relationship_strategy`：3 档 × 4 句模板
- `claim_evidence_links`：单条 `relationship_main_judgement` 链接到所有命中 evidence_key

## 兼容策略

- DB 字段命名：JSONB key 全部换为新名（v3 写入）
- 旧记录（analysis_version='v1'/'v2'）：保留不动，前端按 version 切渲染分支
- 新写入：`analysis_version = 'v3'`
- migration 00012：仅添加 `overall_score INTEGER NOT NULL DEFAULT 0` 列 + COMMENT
