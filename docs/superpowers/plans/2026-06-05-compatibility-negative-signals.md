# 合盘负面信号如实披露 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让合盘报告在四柱存在冲/克/刑/害时如实说出来，不再出现「无合无冲」之类与事实相反的描述；不改动任何评分。

**Architecture:** 新增一个 `detectNegativeSignals` 检测函数，同位比较两人四柱（年-年、月-月、日-日、时-时），复用 `event_signals.go` 现成的冲/刑/害/克对照表，产出 `Polarity:"negative"` 的 `CompatibilityEvidence`。把这些证据接进 `AnalyzeCompatibility` 的 evidence 列表并填进 score-explanation 的 `NegativeFactor` 字段，再更新 prompt 模板：删掉「polarity 永远为 positive」这句假指令，新增「有负面信号必须如实点出」的约束。评分公式、总分、等级、duration、tags 一律不动。

**Tech Stack:** Go 1.25（后端 `pkg/bazi` 纯函数 + `pkg/prompt` 模板字符串），表驱动单元测试。

设计文档：`docs/superpowers/specs/2026-06-05-compatibility-negative-signals-design.md`

---

## 文件结构

- **新建** `backend/pkg/bazi/compatibility_negative.go` — 负面信号检测函数与私有判定辅助（冲/刑/害/克），单一职责。
- **新建** `backend/pkg/bazi/compatibility_negative_test.go` — 检测函数表驱动测试。
- **修改** `backend/pkg/bazi/compatibility.go` — 在 `AnalyzeCompatibility` 中把负面证据追加进 evidence 列表。
- **修改** `backend/pkg/bazi/compatibility_evidence.go` — `buildScoreExplanationsV3` 区分正/负证据，填充 `NegativeFactor`/`NegativeEvidenceKeys`。
- **修改** `backend/pkg/prompt/canonical_compatibility.go` — 删除假指令、新增负面披露约束、bump 版本号。

所有命令均在 `backend/` 目录下执行。

---

## Task 1: 负面信号检测函数

**Files:**
- Create: `backend/pkg/bazi/compatibility_negative.go`
- Test: `backend/pkg/bazi/compatibility_negative_test.go`

- [ ] **Step 1: 写失败测试**

创建 `backend/pkg/bazi/compatibility_negative_test.go`：

```go
package bazi

import (
	"sort"
	"strings"
	"testing"
)

// 用四柱干支构造一个最小 BaziResult（只填检测需要的字段）。
func makeChart(yg, yz, mg, mz, dg, dz, hg, hz string) *BaziResult {
	return &BaziResult{
		YearGan: yg, YearZhi: yz,
		MonthGan: mg, MonthZhi: mz,
		DayGan: dg, DayZhi: dz,
		HourGan: hg, HourZhi: hz,
	}
}

// 收集检测结果里的 EvidenceKey，排序后便于断言。
func negKeys(evs []CompatibilityEvidence) []string {
	keys := make([]string, 0, len(evs))
	for _, e := range evs {
		keys = append(keys, e.EvidenceKey)
	}
	sort.Strings(keys)
	return keys
}

func TestDetectNegativeSignals(t *testing.T) {
	cases := []struct {
		name     string
		a, b     *BaziResult
		wantKeys []string
	}{
		{
			// 真实触发案例：A 乙亥日 / B 己巳日 → 日柱 巳亥相冲 + 乙克己
			name:     "day pillar 巳亥冲 + 乙克己",
			a:        makeChart("丙", "子", "庚", "寅", "乙", "亥", "丙", "戌"),
			b:        makeChart("乙", "亥", "己", "丑", "己", "巳", "壬", "申"),
			wantKeys: []string{"neg_day_chong", "neg_day_gan_ke"},
		},
		{
			// 日柱地支相刑：寅 刑 巳（无礼之刑）
			name:     "day pillar 寅巳相刑",
			a:        makeChart("甲", "辰", "甲", "辰", "甲", "寅", "甲", "辰"),
			b:        makeChart("甲", "辰", "甲", "辰", "甲", "巳", "甲", "辰"),
			wantKeys: []string{"neg_day_xing"},
		},
		{
			// 日柱地支自刑：午 午
			name:     "day pillar 午午自刑",
			a:        makeChart("甲", "辰", "甲", "辰", "甲", "午", "甲", "辰"),
			b:        makeChart("甲", "辰", "甲", "辰", "甲", "午", "甲", "辰"),
			wantKeys: []string{"neg_day_xing"},
		},
		{
			// 月柱地支相害（穿）：子 未
			name:     "month pillar 子未相害",
			a:        makeChart("甲", "辰", "甲", "子", "甲", "辰", "甲", "辰"),
			b:        makeChart("甲", "辰", "甲", "未", "甲", "辰", "甲", "辰"),
			wantKeys: []string{"neg_month_hai"},
		},
		{
			// 全部为合/无关，无负面信号
			name:     "no negatives",
			a:        makeChart("甲", "子", "甲", "子", "甲", "子", "甲", "子"),
			b:        makeChart("甲", "子", "甲", "子", "甲", "子", "甲", "子"),
			wantKeys: []string{},
		},
		{
			name:     "nil safe",
			a:        nil,
			b:        makeChart("甲", "子", "甲", "子", "甲", "子", "甲", "子"),
			wantKeys: []string{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := negKeys(detectNegativeSignals(tc.a, tc.b))
			if strings.Join(got, ",") != strings.Join(tc.wantKeys, ",") {
				t.Errorf("keys = %v, want %v", got, tc.wantKeys)
			}
		})
	}
}

func TestDetectNegativeSignalsArePolarityNegative(t *testing.T) {
	a := makeChart("丙", "子", "庚", "寅", "乙", "亥", "丙", "戌")
	b := makeChart("乙", "亥", "己", "丑", "己", "巳", "壬", "申")
	evs := detectNegativeSignals(a, b)
	if len(evs) == 0 {
		t.Fatal("expected negative evidences, got none")
	}
	for _, e := range evs {
		if e.Polarity != "negative" {
			t.Errorf("evidence %s polarity = %q, want negative", e.EvidenceKey, e.Polarity)
		}
		if e.Dimension == "" || e.Title == "" || e.Detail == "" {
			t.Errorf("evidence %s missing fields: %+v", e.EvidenceKey, e)
		}
	}
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./pkg/bazi/ -run 'TestDetectNegativeSignals' -v`
Expected: 编译失败 / `undefined: detectNegativeSignals`

- [ ] **Step 3: 实现检测函数**

创建 `backend/pkg/bazi/compatibility_negative.go`：

```go
package bazi

import "fmt"

// detectNegativeSignals 同位比较两人四柱（年-年/月-月/日-日/时-时），
// 检出地支冲/刑/害与天干相克，产出 polarity="negative" 的 evidence。
// 与正向评分口径一致（同名柱比较），不影响任何分数。日柱权重最高。
func detectNegativeSignals(a, b *BaziResult) []CompatibilityEvidence {
	if a == nil || b == nil {
		return nil
	}
	pillars := []negPillar{
		{name: "day", label: "日柱", dimension: "day_pillar", weight: 10,
			ganA: a.DayGan, zhiA: a.DayZhi, ganB: b.DayGan, zhiB: b.DayZhi},
		{name: "month", label: "月柱", dimension: "eight_chars", weight: 6,
			ganA: a.MonthGan, zhiA: a.MonthZhi, ganB: b.MonthGan, zhiB: b.MonthZhi},
		{name: "year", label: "年柱", dimension: "zodiac", weight: 5,
			ganA: a.YearGan, zhiA: a.YearZhi, ganB: b.YearGan, zhiB: b.YearZhi},
		{name: "hour", label: "时柱", dimension: "eight_chars", weight: 4,
			ganA: a.HourGan, zhiA: a.HourZhi, ganB: b.HourGan, zhiB: b.HourZhi},
	}
	out := make([]CompatibilityEvidence, 0, 4)
	for _, p := range pillars {
		out = append(out, pillarNegatives(p)...)
	}
	return out
}

// negPillar 描述一柱的同位比较输入与落位元数据。
type negPillar struct {
	name      string // year/month/day/hour（用于 evidence_key）
	label     string // 年柱/月柱/日柱/时柱（用于文案）
	dimension string // 落入报告分节：year→zodiac，month/hour→eight_chars，day→day_pillar
	weight    int
	ganA, zhiA, ganB, zhiB string
}

func pillarNegatives(p negPillar) []CompatibilityEvidence {
	var out []CompatibilityEvidence
	if branchChong(p.zhiA, p.zhiB) {
		out = append(out, negEvidence(p, "chong", "地支相冲", fmt.Sprintf(
			"%s地支 %s 与 %s 相冲——这是两股直接对撞的力量，落到关系里就是这块容易顶牛、各执一端，需要主动让一步才不至于僵住。",
			p.label, p.zhiA, p.zhiB)))
	}
	if branchXing(p.zhiA, p.zhiB) {
		out = append(out, negEvidence(p, "xing", "地支相刑", fmt.Sprintf(
			"%s地支 %s 与 %s 相刑——刑主纠缠、暗耗，容易反复在同一件事上磨人，要警惕翻旧账式的内耗。",
			p.label, p.zhiA, p.zhiB)))
	}
	if branchHai(p.zhiA, p.zhiB) {
		out = append(out, negEvidence(p, "hai", "地支相害", fmt.Sprintf(
			"%s地支 %s 与 %s 相害（穿）——害主暗里别扭、好心办坏事，容易因误解积小怨，要把话说开别憋着。",
			p.label, p.zhiA, p.zhiB)))
	}
	if ke, actor, target := ganKe(p.ganA, p.ganB); ke {
		out = append(out, negEvidence(p, "gan_ke", "天干相克", fmt.Sprintf(
			"%s天干 %s 克 %s（%s克%s）——一方在气势上压住另一方，相处容易一强一弱、一个主导一个迁就，时间长了被压的一方会憋屈。",
			p.label, actor, target, wxPinyin2CN[ganWuxing[actor]], wxPinyin2CN[ganWuxing[target]])))
	}
	return out
}

func negEvidence(p negPillar, kind, typeLabel, detail string) CompatibilityEvidence {
	return CompatibilityEvidence{
		EvidenceKey: "neg_" + p.name + "_" + kind,
		Dimension:   p.dimension,
		Type:        p.label + typeLabel,
		Polarity:    "negative",
		Source:      p.dimension,
		Title:       p.label + typeLabel,
		Detail:      detail,
		Weight:      p.weight,
	}
}

// branchChong 判定两地支是否相冲（六冲）。
func branchChong(x, y string) bool {
	if x == "" || y == "" {
		return false
	}
	return sixChong[x] == y
}

// branchXing 判定两地支是否相刑：异支查 sixXing 双向；同支查自刑。
func branchXing(x, y string) bool {
	if x == "" || y == "" {
		return false
	}
	if x == y {
		return selfXing[x]
	}
	return sixXing[x] == y || sixXing[y] == x
}

// branchHai 判定两地支是否相害（穿）。
func branchHai(x, y string) bool {
	if x == "" || y == "" {
		return false
	}
	return sixHai[x] == y
}

// ganKe 判定两天干是否五行相克，返回 (是否相克, 克方, 被克方)。
// 涵盖天干相冲（庚冲甲等本质即金克木）。
func ganKe(x, y string) (bool, string, string) {
	if x == "" || y == "" {
		return false, "", ""
	}
	wx, wy := ganWuxing[x], ganWuxing[y]
	if wx == "" || wy == "" {
		return false, "", ""
	}
	if wxKe[wx] == wy {
		return true, x, y
	}
	if wxKe[wy] == wx {
		return true, y, x
	}
	return false, "", ""
}
```

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./pkg/bazi/ -run 'TestDetectNegativeSignals' -v`
Expected: PASS（两个测试函数全绿）

- [ ] **Step 5: 提交**

```bash
git add pkg/bazi/compatibility_negative.go pkg/bazi/compatibility_negative_test.go
git commit -m "feat(bazi): detect compatibility negative signals (冲/克/刑/害)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 2: 接入 AnalyzeCompatibility 与 score-explanation

**Files:**
- Modify: `backend/pkg/bazi/compatibility.go:155`（`evidences := buildCompatibilityEvidencesV3(a, b)` 之后追加）
- Modify: `backend/pkg/bazi/compatibility_evidence.go:204-225`（`buildScoreExplanationsV3` + 辅助函数）
- Test: `backend/pkg/bazi/compatibility_negative_test.go`（追加集成测试）

- [ ] **Step 1: 写失败测试**

在 `backend/pkg/bazi/compatibility_negative_test.go` 末尾追加：

```go
// 集成断言：触发案例的负面信号进入 Evidences，且评分丝毫未动（总分仍 34）。
func TestAnalyzeCompatibilitySurfacesNegativesWithoutScoreChange(t *testing.T) {
	a := makeChart("丙", "子", "庚", "寅", "乙", "亥", "丙", "戌")
	b := makeChart("乙", "亥", "己", "丑", "己", "巳", "壬", "申")
	res := AnalyzeCompatibility(a, b)

	// 评分未被触碰
	if res.OverallScore != 34 {
		t.Errorf("OverallScore = %d, want 34 (评分不应被负面信号改动)", res.OverallScore)
	}
	if res.OverallLevel != CompatibilityLow {
		t.Errorf("OverallLevel = %q, want low", res.OverallLevel)
	}
	if res.DimensionScores.DayPillar != 0 {
		t.Errorf("DayPillar score = %d, want 0", res.DimensionScores.DayPillar)
	}

	// 负面证据已进入列表
	var hasChong, hasKe bool
	for _, e := range res.Evidences {
		if e.EvidenceKey == "neg_day_chong" {
			hasChong = true
		}
		if e.EvidenceKey == "neg_day_gan_ke" {
			hasKe = true
		}
	}
	if !hasChong || !hasKe {
		t.Errorf("expected neg_day_chong & neg_day_gan_ke in Evidences, got %+v", res.Evidences)
	}

	// score_explanation 的 day_pillar 条目填了负面因子
	var dayExp *CompatibilityScoreExplanation
	for i := range res.ScoreExplanations {
		if res.ScoreExplanations[i].Dimension == "day_pillar" {
			dayExp = &res.ScoreExplanations[i]
		}
	}
	if dayExp == nil {
		t.Fatal("missing day_pillar score explanation")
	}
	if dayExp.NegativeFactor == "" || len(dayExp.NegativeEvidenceKeys) == 0 {
		t.Errorf("day_pillar explanation missing negatives: %+v", dayExp)
	}
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./pkg/bazi/ -run 'TestAnalyzeCompatibilitySurfacesNegatives' -v`
Expected: FAIL —`neg_day_chong` 不在 Evidences 中，且 `NegativeFactor` 为空（因为尚未接线）

- [ ] **Step 3a: 把负面证据接进 AnalyzeCompatibility**

修改 `backend/pkg/bazi/compatibility.go`，把这一行：

```go
	evidences := buildCompatibilityEvidencesV3(a, b)
```

改为：

```go
	evidences := buildCompatibilityEvidencesV3(a, b)
	evidences = append(evidences, detectNegativeSignals(a, b)...)
```

- [ ] **Step 3b: 让 score-explanation 区分正/负证据**

修改 `backend/pkg/bazi/compatibility_evidence.go`。

(1) 顶部 import 由 `import "fmt"` 改为：

```go
import (
	"fmt"
	"strings"
)
```

(2) 把 `buildScoreExplanationsV3` 整个函数体替换为：

```go
// buildScoreExplanationsV3 按 4 模块各出一条解释（zodiac/nayin/day_pillar/eight_chars）。
// PositiveFactor 取该维度首条 positive 证据；NegativeFactor 汇总该维度全部 negative 证据。
func buildScoreExplanationsV3(a, b *BaziResult, evidences []CompatibilityEvidence) []CompatibilityScoreExplanation {
	dimensions := []string{"zodiac", "nayin", "day_pillar", "eight_chars"}
	out := make([]CompatibilityScoreExplanation, 0, 4)
	for _, dim := range dimensions {
		hit := findPositiveEvidenceByDimension(evidences, dim)
		exp := CompatibilityScoreExplanation{Dimension: dim}
		if hit != nil {
			exp.PositiveFactor = hit.Title
			exp.PositiveEvidenceKeys = []string{hit.EvidenceKey}
		}
		if negs := findNegativeEvidencesByDimension(evidences, dim); len(negs) > 0 {
			titles := make([]string, 0, len(negs))
			keys := make([]string, 0, len(negs))
			for i := range negs {
				titles = append(titles, negs[i].Title)
				keys = append(keys, negs[i].EvidenceKey)
			}
			exp.NegativeFactor = strings.Join(titles, "、")
			exp.NegativeEvidenceKeys = keys
		}
		exp.Summary = scoreExplanationSummaryV3(dim, hit, a, b)
		out = append(out, exp)
	}
	return out
}
```

(3) 把原 `findEvidenceByDimension` 函数替换为下面两个函数（按 polarity 过滤）：

```go
func findPositiveEvidenceByDimension(evidences []CompatibilityEvidence, dim string) *CompatibilityEvidence {
	for i := range evidences {
		if evidences[i].Dimension == dim && evidences[i].Polarity != "negative" {
			return &evidences[i]
		}
	}
	return nil
}

func findNegativeEvidencesByDimension(evidences []CompatibilityEvidence, dim string) []CompatibilityEvidence {
	var out []CompatibilityEvidence
	for i := range evidences {
		if evidences[i].Dimension == dim && evidences[i].Polarity == "negative" {
			out = append(out, evidences[i])
		}
	}
	return out
}
```

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./pkg/bazi/ -run 'TestAnalyzeCompatibility' -v`
Expected: PASS

- [ ] **Step 5: 跑整包回归，确认评分/既有快照未破**

Run: `go test ./pkg/bazi/...`
Expected: ok（既有正向证据、评分、duration 等测试全绿）

- [ ] **Step 6: 提交**

```bash
git add pkg/bazi/compatibility.go pkg/bazi/compatibility_evidence.go pkg/bazi/compatibility_negative_test.go
git commit -m "feat(bazi): surface negative signals in compatibility evidence & explanations

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 3: 更新 prompt 模板（删假指令 + 负面披露约束）

**Files:**
- Modify: `backend/pkg/prompt/canonical_compatibility.go`（init 版本号 + 模板字符串第 32/34/37/54 行附近 + 证据约束段）

> 说明：`drift_test.go` / `sync_test.go` / `canonical_compatibility_famous_couple_test.go` 均从当前注册内容**动态**取 Content/Hash 或只检查 token 是否存在，不含写死的快照哈希；下面的改动只要不删 famous_couple 相关 token 就不会破坏它们。改 Content 后 bump 版本号是让管理员侧自定义文案被标记为可同步的既有约定。

- [ ] **Step 1: bump 版本号**

修改 `backend/pkg/prompt/canonical_compatibility.go` 的 `init`，把：

```go
		Version:     "v3.1-question-aware-4",
```

改为：

```go
		Version:     "v3.1-question-aware-5",
```

- [ ] **Step 2: 修订评分规则说明（删除假指令、补负面口径）**

在模板字符串中，把第 32 行：

```
- zodiac（合属相，0–50，v3.1 三级）：年支六合/三合 = 50；五行同行（双生）= 30；五行相生 = 20；相克/相冲/相穿 = 0。
```

改为：

```
- zodiac（合属相，0–50，v3.1 三级）：年支六合/三合 = 50；五行同行（双生）= 30；五行相生 = 20；相克/相冲/相穿 = 0（注意：0 分不代表「无关系」，可能恰恰是相冲/相克，需结合 negative 证据判断）。
```

把第 34 行：

```
- day_pillar（合日柱，0–10，v3.1 四级）：日支六合/三合 + 干合/相生 = 10；日支六合/三合 = 5；日支五行同/相生 = 3；日支相克/相冲 = 0。
```

改为：

```
- day_pillar（合日柱，0–10，v3.1 四级）：日支六合/三合 + 干合/相生 = 10；日支六合/三合 = 5；日支五行同/相生 = 3；日支相克/相冲 = 0（0 分可能是日柱相冲/相刑，须如实点出，禁止说「无冲」）。
```

把第 37 行：

```
- 本算法采用「纯加分制」，所有 evidence 的 polarity 均为 positive；不命中的模块得 0 分，不产生 evidence。
```

改为：

```
- evidence 的 polarity 有 positive（合/同行/相生等正面信号）与 negative（冲/克/刑/害等负面信号）两类。正面信号参与加分；负面信号不参与本版评分，但**必须如实写进报告**，不得忽略或回避。
```

- [ ] **Step 3: 修订证据来源「注」行**

把第 54 行：

```
注：所有 evidence 来源仅四种（zodiac / nayin / day_pillar / eight_chars），polarity 永远为 positive。
```

改为：

```
注：evidence 的 source 为 zodiac / nayin / day_pillar / eight_chars；polarity 为 positive 或 negative。negative 证据（如日柱地支相冲、天干相克）代表真实的冲克刑害，必须在对应分节如实呈现。
```

- [ ] **Step 4: 在「证据约束」段补一条负面披露硬约束**

找到第 60 行附近的「证据约束：」块（首条为「- 所有主要判断必须引用 evidence_key。」），在该块内新增一条：

```
- 凡输入 evidence 中存在 polarity="negative" 的项，必须在其所属维度分节（按 dimension：day_pillar→合日柱、zodiac→合属相、eight_chars→合八字）如实指出对应的冲/克/刑/害，并用一句大白话解释它对关系的实际影响；**严禁出现与负面证据相矛盾的描述（如证据为日柱相冲却写「无冲」「无合无冲」）**。
```

- [ ] **Step 5: 编译并跑 prompt 包测试**

Run: `go test ./pkg/prompt/...`
Expected: ok（drift / sync / famous_couple 全绿）

- [ ] **Step 6: 跑全后端测试**

Run: `go test ./...`
Expected: ok

- [ ] **Step 7: 提交**

```bash
git add pkg/prompt/canonical_compatibility.go
git commit -m "feat(prompt): require compatibility report to disclose negative signals truthfully

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## 验收（人工 / 端到端）

实现完成后，按设计文档 §5 成功标准验证：

1. `go test ./...` 全绿。
2. 用触发案例（A 1996-02-08 20时、B 1996-02-02 16时）走一次「生成深度解读」，确认「合日柱」分节明确写出「巳亥相冲」「乙克己」，不再出现「无合无冲」。
3. 该案例总分仍为 34、等级仍 low。

> 注：第 2 步依赖真实 LLM 调用，属人工抽检，不纳入自动化测试。

---

## Self-Review 结论

- **Spec 覆盖**：§3.1 检测函数→Task 1；§3.2 接线+NegativeFactor→Task 2；§3.3 prompt→Task 3；§4 测试分散在各 Task 的 TDD 步骤；§5 验收→末段。无遗漏。
- **占位符**：无 TBD/TODO，所有代码步骤含完整代码。
- **类型一致性**：`detectNegativeSignals`、`negPillar`、`negEvidence`、`branchChong/branchXing/branchHai/ganKe`、`findPositiveEvidenceByDimension`、`findNegativeEvidencesByDimension` 在定义与调用处命名一致；`CompatibilityEvidence` 字段（EvidenceKey/Dimension/Polarity/Source/Type/Title/Detail/Weight）与 `compatibility.go` 现有结构体一致；`CompatibilityScoreExplanation.NegativeFactor/NegativeEvidenceKeys` 为现有字段。
- **复用**：`sixChong/sixXing/selfXing/sixHai/wxKe/ganWuxing/wxPinyin2CN` 全部复用 `event_signals.go`，未新造表。
