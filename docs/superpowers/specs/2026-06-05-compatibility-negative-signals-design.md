# 合盘「负面信号如实披露」设计文档

日期：2026-06-05
状态：已通过 brainstorming 评审，待用户复核

## 1. 目标

让合盘报告在存在**冲 / 克 / 刑 / 害**时直接如实说出来，不再回避或说反话。

**触发问题（真实案例）**：两个八字 1996-02-08 20时（乙亥日）与 1996-02-02 16时（己巳日），日柱实为**巳亥相冲 + 乙克己（天克地冲）**。但当前报告「合日柱」分节写的是「地支无合无冲且五行不亲」——与事实相反。

**根因**：评分算法只计算「合 / 同行 / 相生」，完全不计算冲克刑害。负面关系在整条链路（评分→证据→说明→大模型 prompt→报告）全程缺席；prompt 模板更明确写死「polarity 永远为 positive」，等于在指令层面教大模型说假话。

用户明确诉求：**「相冲能直接说出来」**——只要描述如实，本次**不改评分**。

## 2. 关键决策（brainstorming 已确认）

| 决策点 | 结论 |
|---|---|
| 信号范围 | 冲 / 克 / 刑 / 害**全套**（地支冲、地支刑+自刑、地支害/穿、天干相克） |
| 是否影响分数 | **否**。方案 A：只如实描述，`DimensionScores`/总分/`overallLevel`/`duration`/tags 全部不动 |
| 评分惩罚 | 留作**下一个独立改动**（v3.1 评分体系联动复杂，不和描述层混改） |
| 对照表 | 复用 `event_signals.go` 现成表，**不新造**（CLAUDE.md §5.1 复用优先） |
| 同柱多关系 | 每类各出一条，不去重合并（合并属评分层，留给下次） |

## 3. 架构与改动点（3 处）

### 3.1 新增检测函数 `pkg/bazi/compatibility_negative.go`

```go
func detectNegativeSignals(a, b *BaziResult) []CompatibilityEvidence
```

扫描两人四柱，复用现成对照表：

| 信号 | 数据源（event_signals.go） |
|---|---|
| 地支相冲 | `sixChong` |
| 地支相刑 / 自刑 | `sixXing` + `selfXing` |
| 地支相害（穿） | `sixHai` |
| 天干相克 | `ganWuxing` + `wxKe` |

产出 `CompatibilityEvidence`：
- `Polarity: "negative"`
- 复用现有字段 `Source`（哪柱）、`Actor`/`Target`（谁冲/克谁）、`Type`（如「日柱地支相冲」）、`Title`/`Detail`
- `Weight`：**日柱（夫妻宫）命中给最高**，月柱次之，年/时再次，便于报告优先讲（对齐 `event_signals.go:409` `pillarWeightLabel` 既有「日柱最重」口径）
- 文案尺度参考现有 evidence 的大白话风格（术语后跟一句解释）

边界：空盘 / 缺柱安全返回空 slice（对齐 `AnalyzeCompatibility` 既有 nil 防护）。

### 3.2 接入数据流（`pkg/bazi/compatibility.go`）

- `AnalyzeCompatibility` 中 `evidences` 末尾追加 `detectNegativeSignals(a, b)`（约 `compatibility.go:155`）。
- `buildScoreExplanationsV3` 把负面命中填进**已预留但一直为空**的 `NegativeFactor` / `NegativeEvidenceKeys` 字段。
- **不改动**：`DimensionScores`、`total`、`overallLevelFromScoreV3`、`buildDurationAssessmentV3`、`buildSummaryTagsV3`、`countHitsV3`、`buildConsultingAssessmentV3` 的分数逻辑。

### 3.3 更新 prompt 模板 `pkg/prompt/canonical_compatibility.go`

- **删除**已成假话的说明：第 37 行「所有 evidence 的 polarity 均为 positive」、第 54 行「polarity 永远为 positive」、「纯加分不产 negative evidence」等。
- **新增指令**：当 evidence 含 `polarity:"negative"` 时，**必须**在对应分节如实点出冲/克/刑/害，**禁止说「无冲」之类与事实相反的话**；并按既有规则（术语后跟大白话）解释它对关系的含义。
- 同步修订第 32/34 行口径：让模型理解「0 分」背后可能藏着冲克，而非「无任何关系」。
- canonical prompt 改动后需同步：`canonical_compatibility_famous_couple_test.go` 黄金快照、`drift_test.go`/`sync_test.go`（纳入实现计划的验证步骤）。

## 4. 测试

`pkg/bazi/compatibility_negative_test.go`（表驱动）：
- **核心用例**：乙亥 / 己巳 → 断言检出「日柱地支相冲（巳亥）」+「日柱天干相克（乙克己）」两条 negative evidence。
- **回归断言**：同一对八字 `OverallScore == 34`、`DimensionScores` 与改动前一致——证明评分未被触碰。
- 覆盖刑（如寅巳）、自刑（如午午）、害（如子未）各一例。
- 无负面信号的盘 → 返回空，且现有 positive 测试不受影响。

prompt 改动：更新并通过 `drift_test.go` / `sync_test.go` / 名人配对黄金测试。

## 5. 成功标准

1. `go test ./...` 全绿。
2. 跑触发案例生成深度报告，「合日柱」分节**明确写出「巳亥相冲、乙克己」**，不再出现「无合无冲」。
3. 该案例总分仍为 34，等级仍为 low（评分未动）。

## 6. 明确不做（YAGNI）

- 不引入扣分 / 封顶 / 等级惩罚（下次独立改动）。
- 不做同柱多关系的合并 / 叠加。
- 不新增对照表。
- 不动前端展示组件（报告文字由 LLM 生成，前端按现有结构渲染即可）。
