# Gongjia Virtual Branch Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add first-version 原局夹拱虚支 support as a neutral signal layer for Bazi charts, reports, and event signals.

**Architecture:** Add a focused backend `gongjia` unit in `backend/pkg/bazi` that detects adjacent natal pillar virtual branches and enriches them with hidden stems, ten gods, and branch-based shensha. Wire the result into `BaziResult`, cached snapshot backfill, neutral year-event signals, report prompt context, and a small frontend result-page panel without changing core wuxing, yongshen, mingge, or strength calculations.

**Tech Stack:** Go + existing `backend/pkg/bazi` package, Gin service snapshot loading, React 18 + TypeScript + CSS variables, existing Go tests and frontend Node text/DOM-style tests.

---

## File Map

- Create: `backend/pkg/bazi/gongjia.go`
  - Owns `GongJiaItem`, detection, enrichment, branch-based shensha for virtual branches, and snapshot backfill helper.
- Create: `backend/pkg/bazi/gongjia_test.go`
  - Unit tests for adjacent-only detection, reverse branch order, enrichment, and no core-field mutation.
- Modify: `backend/pkg/bazi/engine.go`
  - Add `GongJia []GongJiaItem json:"gong_jia,omitempty"` to `BaziResult`.
  - Call `EnsureGongJia(res)` after ten-god relation setup.
- Modify: `backend/internal/service/report_service.go`
  - Backfill old snapshots with `EnsureGongJia`.
  - Add compact `[原局夹拱]` prompt context.
- Modify: `backend/internal/service/report_service_test.go`
  - Assert prompt includes gongjia context and does not say it changes core yongshen/mingge.
- Modify: `backend/pkg/bazi/event_signals.go`
  - Add `SourceGongJia`.
  - Add neutral event signals for 大运/流年冲合刑害夹支 and 三合/三会 involving a virtual branch.
- Modify: `backend/pkg/bazi/event_signals_test.go`
  - Unit tests for neutral gongjia trigger signals.
- Create: `frontend/src/components/GongJiaPanel.tsx`
  - Result-page panel for structured `gong_jia`.
- Create: `frontend/src/components/GongJiaPanel.css`
  - Compact panel styling consistent with existing result-page cards.
- Modify: `frontend/src/pages/ResultPage.tsx`
  - Add `GongJiaItem` type, import panel, render under the basic chart grid when present.
- Create: `frontend/tests/gongjia-panel.test.mjs`
  - Static/render-oriented checks for the new component and result-page integration.

## Task 1: Backend Data Model And Adjacent Detection

**Files:**
- Create: `backend/pkg/bazi/gongjia.go`
- Create: `backend/pkg/bazi/gongjia_test.go`
- Modify: `backend/pkg/bazi/engine.go`

- [ ] **Step 1: Write failing detection tests**

Add `backend/pkg/bazi/gongjia_test.go`:

```go
package bazi

import "testing"

func TestBuildGongJiaAdjacentSameGanGap(t *testing.T) {
	r := &BaziResult{
		YearGan: "甲", YearZhi: "子",
		MonthGan: "甲", MonthZhi: "寅",
		DayGan: "丙", DayZhi: "午",
		HourGan: "戊", HourZhi: "申",
	}

	got := BuildGongJia(r)
	if len(got) != 1 {
		t.Fatalf("BuildGongJia count = %d, want 1: %+v", len(got), got)
	}
	item := got[0]
	if item.Source != "year_month" || item.SameGan != "甲" || item.VirtualZhi != "丑" {
		t.Fatalf("unexpected item: %+v", item)
	}
	if len(item.HideGan) != 3 || item.HideGan[0] != "己" || item.HideGan[1] != "癸" || item.HideGan[2] != "辛" {
		t.Fatalf("hide gan = %+v, want [己 癸 辛]", item.HideGan)
	}
}

func TestBuildGongJiaReverseBranchOrder(t *testing.T) {
	r := &BaziResult{
		YearGan: "甲", YearZhi: "寅",
		MonthGan: "甲", MonthZhi: "子",
		DayGan: "丙", DayZhi: "午",
		HourGan: "戊", HourZhi: "申",
	}

	got := BuildGongJia(r)
	if len(got) != 1 || got[0].VirtualZhi != "丑" {
		t.Fatalf("reverse order should still clip 丑, got %+v", got)
	}
}

func TestBuildGongJiaAdjacentOnly(t *testing.T) {
	r := &BaziResult{
		YearGan: "甲", YearZhi: "子",
		MonthGan: "乙", MonthZhi: "午",
		DayGan: "甲", DayZhi: "寅",
		HourGan: "丙", HourZhi: "申",
	}

	got := BuildGongJia(r)
	if len(got) != 0 {
		t.Fatalf("non-adjacent year/day pair must not clip, got %+v", got)
	}
}

func TestBuildGongJiaRejectsDifferentGanOrNonGapBranches(t *testing.T) {
	cases := []struct {
		name string
		r    *BaziResult
	}{
		{
			name: "different gan",
			r: &BaziResult{
				YearGan: "甲", YearZhi: "子",
				MonthGan: "乙", MonthZhi: "寅",
			},
		},
		{
			name: "not one branch between",
			r: &BaziResult{
				YearGan: "甲", YearZhi: "子",
				MonthGan: "甲", MonthZhi: "卯",
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := BuildGongJia(tc.r); len(got) != 0 {
				t.Fatalf("BuildGongJia() = %+v, want none", got)
			}
		})
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run:

```bash
cd backend && go test ./pkg/bazi -run 'TestBuildGongJia' -count=1
```

Expected: FAIL with `undefined: BuildGongJia`.

- [ ] **Step 3: Add the data model and detection implementation**

Create `backend/pkg/bazi/gongjia.go`:

```go
package bazi

import "fmt"

type GongJiaItem struct {
	Source       string   `json:"source"`
	SourceLabels []string `json:"source_labels"`
	SameGan      string   `json:"same_gan"`
	SourceZhis   []string `json:"source_zhis"`
	VirtualZhi   string   `json:"virtual_zhi"`
	HideGan      []string `json:"hide_gan"`
	ShiShen      []string `json:"shishen"`
	ShenSha      []string `json:"shensha"`
	Meaning      string   `json:"meaning"`
}

type gongJiaPair struct {
	source       string
	sourceLabels []string
	leftGan      string
	leftZhi      string
	rightGan     string
	rightZhi     string
}

var zhiOrderForGongJia = []string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}

var zhiIndexForGongJia = map[string]int{
	"子": 0, "丑": 1, "寅": 2, "卯": 3, "辰": 4, "巳": 5,
	"午": 6, "未": 7, "申": 8, "酉": 9, "戌": 10, "亥": 11,
}

func BuildGongJia(r *BaziResult) []GongJiaItem {
	if r == nil {
		return nil
	}
	pairs := []gongJiaPair{
		{source: "year_month", sourceLabels: []string{"年柱", "月柱"}, leftGan: r.YearGan, leftZhi: r.YearZhi, rightGan: r.MonthGan, rightZhi: r.MonthZhi},
		{source: "month_day", sourceLabels: []string{"月柱", "日柱"}, leftGan: r.MonthGan, leftZhi: r.MonthZhi, rightGan: r.DayGan, rightZhi: r.DayZhi},
		{source: "day_hour", sourceLabels: []string{"日柱", "时柱"}, leftGan: r.DayGan, leftZhi: r.DayZhi, rightGan: r.HourGan, rightZhi: r.HourZhi},
	}

	out := make([]GongJiaItem, 0, 3)
	for _, p := range pairs {
		if p.leftGan == "" || p.leftGan != p.rightGan {
			continue
		}
		virtualZhi, ok := clippedZhiBetween(p.leftZhi, p.rightZhi)
		if !ok {
			continue
		}
		hideGan := append([]string{}, zhiHideGanFull[virtualZhi]...)
		shiShen := make([]string, 0, len(hideGan))
		for _, g := range hideGan {
			if ss := GetShiShen(r.DayGan, g); ss != "" {
				shiShen = append(shiShen, ss)
			}
		}
		out = append(out, GongJiaItem{
			Source:       p.source,
			SourceLabels: append([]string{}, p.sourceLabels...),
			SameGan:      p.leftGan,
			SourceZhis:   []string{p.leftZhi, p.rightZhi},
			VirtualZhi:   virtualZhi,
			HideGan:      hideGan,
			ShiShen:      shiShen,
			Meaning:      fmt.Sprintf("原局%s与%s天干同为%s，地支%s%s夹%s，%s为暗藏虚支。", p.sourceLabels[0], p.sourceLabels[1], p.leftGan, p.leftZhi, p.rightZhi, virtualZhi, virtualZhi),
		})
	}
	return out
}

func clippedZhiBetween(a, b string) (string, bool) {
	ai, aOK := zhiIndexForGongJia[a]
	bi, bOK := zhiIndexForGongJia[b]
	if !aOK || !bOK {
		return "", false
	}
	if (ai+2)%len(zhiOrderForGongJia) == bi {
		return zhiOrderForGongJia[(ai+1)%len(zhiOrderForGongJia)], true
	}
	if (bi+2)%len(zhiOrderForGongJia) == ai {
		return zhiOrderForGongJia[(bi+1)%len(zhiOrderForGongJia)], true
	}
	return "", false
}

func EnsureGongJia(r *BaziResult) bool {
	if r == nil || len(r.GongJia) > 0 {
		return false
	}
	r.GongJia = BuildGongJia(r)
	return len(r.GongJia) > 0
}
```

Modify `backend/pkg/bazi/engine.go`:

```go
// Add inside BaziResult after TenGodRelation:
GongJia []GongJiaItem `json:"gong_jia,omitempty"`
```

Then add after `EnsureTenGodRelation(res)`:

```go
EnsureGongJia(res)
```

- [ ] **Step 4: Run detection tests**

Run:

```bash
cd backend && go test ./pkg/bazi -run 'TestBuildGongJia' -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit Task 1**

Run:

```bash
git add backend/pkg/bazi/gongjia.go backend/pkg/bazi/gongjia_test.go backend/pkg/bazi/engine.go
git commit -m "feat(bazi): detect gongjia virtual branches"
```

## Task 2: Virtual-Branch Shensha Enrichment

**Files:**
- Modify: `backend/pkg/bazi/gongjia.go`
- Modify: `backend/pkg/bazi/gongjia_test.go`

- [ ] **Step 1: Write failing shensha tests**

Append to `backend/pkg/bazi/gongjia_test.go`:

```go
func TestBuildGongJiaAddsBranchBasedShensha(t *testing.T) {
	r := &BaziResult{
		YearGan: "甲", YearZhi: "子",
		MonthGan: "甲", MonthZhi: "寅",
		DayGan: "庚", DayZhi: "午",
		HourGan: "戊", HourZhi: "申",
	}

	got := BuildGongJia(r)
	if len(got) != 1 {
		t.Fatalf("BuildGongJia count = %d, want 1", len(got))
	}
	if !containsStr(got[0].ShenSha, "天乙贵人") {
		t.Fatalf("virtual 丑 should inherit branch-based 天乙贵人 from 甲/庚, got %+v", got[0].ShenSha)
	}
}

func TestVirtualBranchShenshaDoesNotCreateFullPillarGanZhi(t *testing.T) {
	r := &BaziResult{
		YearGan: "甲", YearZhi: "子",
		MonthGan: "甲", MonthZhi: "寅",
		DayGan: "庚", DayZhi: "午",
		HourGan: "戊", HourZhi: "申",
	}

	got := BuildGongJia(r)
	if len(got) != 1 {
		t.Fatalf("BuildGongJia count = %d, want 1", len(got))
	}
	if containsStr(got[0].ShenSha, "阴差阳错") || containsStr(got[0].ShenSha, "魁罡") {
		t.Fatalf("virtual branch must not be treated as a full pillar, got %+v", got[0].ShenSha)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run:

```bash
cd backend && go test ./pkg/bazi -run 'Test.*GongJia.*ShenSha|TestVirtualBranchShensha' -count=1
```

Expected: FAIL because `ShenSha` is empty.

- [ ] **Step 3: Add virtual branch shensha helper**

In `backend/pkg/bazi/gongjia.go`, add `BuildGongJia` assignment:

```go
shensha := GetGongJiaShenSha(r.YearGan, r.YearZhi, r.MonthZhi, r.DayGan, r.DayZhi, virtualZhi)
```

Then set:

```go
ShenSha: shensha,
```

Add the helper functions:

```go
func GetGongJiaShenSha(yearGan, yearZhi, monthZhi, dayGan, dayZhi, virtualZhi string) []string {
	var out []string
	add := func(cond bool, name string) {
		if cond && !containsStr(out, name) {
			out = append(out, name)
		}
	}

	tianYi := func(stem, branch string) bool {
		return (containsAnyGan(stem, "甲戊庚") && containsAnyZhi(branch, "丑未")) ||
			(containsAnyGan(stem, "乙己") && containsAnyZhi(branch, "子申")) ||
			(containsAnyGan(stem, "丙丁") && containsAnyZhi(branch, "亥酉")) ||
			(containsAnyGan(stem, "壬癸") && containsAnyZhi(branch, "卯巳")) ||
			(stem == "辛" && containsAnyZhi(branch, "午寅"))
	}
	add(tianYi(dayGan, virtualZhi) || tianYi(yearGan, virtualZhi), "天乙贵人")

	wenChang := func(stem, branch string) bool {
		return (stem == "甲" && branch == "巳") || (stem == "乙" && branch == "午") ||
			(containsAnyGan(stem, "丙戊") && branch == "申") || (containsAnyGan(stem, "丁己") && branch == "酉") ||
			(stem == "庚" && branch == "亥") || (stem == "辛" && branch == "子") ||
			(stem == "壬" && branch == "寅") || (stem == "癸" && branch == "卯")
	}
	add(wenChang(dayGan, virtualZhi) || wenChang(yearGan, virtualZhi), "文昌贵人")

	add(isTaohuaBase(yearZhi, virtualZhi) || isTaohuaBase(dayZhi, virtualZhi), "桃花")
	add(isYimaBase(yearZhi, virtualZhi) || isYimaBase(dayZhi, virtualZhi), "驿马")
	add(isHuagaiBase(yearZhi, virtualZhi) || isHuagaiBase(dayZhi, virtualZhi), "华盖")
	add(isJiangxingBase(yearZhi, virtualZhi) || isJiangxingBase(dayZhi, virtualZhi), "将星")
	add(isJieshaBase(yearZhi, virtualZhi) || isJieshaBase(dayZhi, virtualZhi), "劫煞")
	add(isZaishaBase(yearZhi, virtualZhi) || isZaishaBase(dayZhi, virtualZhi), "灾煞")

	return out
}

func containsAnyGan(stem, chars string) bool {
	return stem != "" && strings.Contains(chars, stem)
}

func containsAnyZhi(branch, chars string) bool {
	return branch != "" && strings.Contains(chars, branch)
}

func isHuagaiBase(base, check string) bool {
	return (strings.Contains("申子辰", base) && check == "辰") ||
		(strings.Contains("寅午戌", base) && check == "戌") ||
		(strings.Contains("亥卯未", base) && check == "未") ||
		(strings.Contains("巳酉丑", base) && check == "丑")
}

func isJiangxingBase(base, check string) bool {
	return (strings.Contains("申子辰", base) && check == "子") ||
		(strings.Contains("寅午戌", base) && check == "午") ||
		(strings.Contains("亥卯未", base) && check == "卯") ||
		(strings.Contains("巳酉丑", base) && check == "酉")
}

func isJieshaBase(base, check string) bool {
	return (strings.Contains("申子辰", base) && check == "巳") ||
		(strings.Contains("寅午戌", base) && check == "亥") ||
		(strings.Contains("亥卯未", base) && check == "申") ||
		(strings.Contains("巳酉丑", base) && check == "寅")
}

func isZaishaBase(base, check string) bool {
	return (strings.Contains("申子辰", base) && check == "午") ||
		(strings.Contains("寅午戌", base) && check == "子") ||
		(strings.Contains("亥卯未", base) && check == "酉") ||
		(strings.Contains("巳酉丑", base) && check == "卯")
}
```

Add `strings` to the imports:

```go
import (
	"fmt"
	"strings"
)
```

- [ ] **Step 4: Run shensha tests**

Run:

```bash
cd backend && go test ./pkg/bazi -run 'Test.*GongJia.*ShenSha|TestVirtualBranchShensha|TestBuildGongJia' -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit Task 2**

Run:

```bash
git add backend/pkg/bazi/gongjia.go backend/pkg/bazi/gongjia_test.go
git commit -m "feat(bazi): enrich gongjia virtual shensha"
```

## Task 3: Snapshot Backfill And Prompt Context

**Files:**
- Modify: `backend/internal/service/report_service.go`
- Modify: `backend/internal/service/report_service_test.go`

- [ ] **Step 1: Write failing prompt/backfill tests**

Append to `backend/internal/service/report_service_test.go`:

```go
func TestBuildBaziPromptIncludesGongJiaContext(t *testing.T) {
	result := &bazi.BaziResult{
		YearGan: "甲", YearZhi: "子",
		MonthGan: "甲", MonthZhi: "寅",
		DayGan: "庚", DayZhi: "午",
		HourGan: "戊", HourZhi: "申",
		YearGanWuxing: "木", YearZhiWuxing: "水",
		MonthGanWuxing: "木", MonthZhiWuxing: "木",
		DayGanWuxing: "金", DayZhiWuxing: "火",
		HourGanWuxing: "土", HourZhiWuxing: "金",
		YearHideGan: []string{"癸"},
		MonthHideGan: []string{"甲", "丙", "戊"},
		DayHideGan: []string{"丁", "己"},
		HourHideGan: []string{"庚", "壬", "戊"},
		YearNaYin: "海中金", MonthNaYin: "大溪水", DayNaYin: "路旁土", HourNaYin: "大驿土",
		Wuxing: bazi.WuxingStats{Mu: 2, Huo: 1, Tu: 1, Jin: 2, Shui: 1},
		Gender: "female",
		MingGe: "七杀格",
		MingGeDesc: "测试格局",
	}
	bazi.EnsureGongJia(result)

	prompt := buildBaziPrompt(result)
	for _, want := range []string{"[原局夹拱]", "年月夹丑", "暗藏虚支", "不改原局五行、用神或命格", "天乙贵人"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q\n%s", want, prompt)
		}
	}
}
```

If `report_service_test.go` does not import `strings` or `yuanju/pkg/bazi`, add them.

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
cd backend && go test ./internal/service -run TestBuildBaziPromptIncludesGongJiaContext -count=1
```

Expected: FAIL because prompt does not include `[原局夹拱]`.

- [ ] **Step 3: Add prompt formatting and snapshot backfill**

In `backend/internal/service/report_service.go`, add:

```go
func formatGongJiaSummary(result *bazi.BaziResult) string {
	if result == nil || len(result.GongJia) == 0 {
		return ""
	}
	var lines []string
	for _, item := range result.GongJia {
		source := strings.Join(item.SourceLabels, "")
		hideGan := strings.Join(item.HideGan, "")
		shiShen := strings.Join(item.ShiShen, "、")
		shensha := "无"
		if len(item.ShenSha) > 0 {
			shensha = strings.Join(item.ShenSha, "、")
		}
		lines = append(lines, fmt.Sprintf("%s夹%s：来源地支%s，藏干%s，对日主十神=%s，拱神煞=%s。此为暗藏虚支，只作暗线与应期参考，不改原局五行、用神或命格。",
			source, item.VirtualZhi, strings.Join(item.SourceZhis, "/"), hideGan, shiShen, shensha))
	}
	return "\n[原局夹拱]\n" + strings.Join(lines, "\n") + "\n"
}
```

In `LoadOrCalculateResult`, after `bazi.EnsureTenGodRelation(&cached)`, add:

```go
if bazi.EnsureGongJia(&cached) {
	backfilled = true
}
```

Move `backfilled := false` so it is declared before both `EnsureGongJia` and `ShishenConfidence` backfills:

```go
backfilled := false
bazi.EnsureTenGodRelation(&cached)
if bazi.EnsureGongJia(&cached) {
	backfilled = true
}
```

In `buildBaziPrompt`, before `prompt :=`, add:

```go
gongJiaStr := formatGongJiaSummary(r)
```

Then insert `gongJiaStr` after the `[神煞]` block and before `[大运序列]`:

```go
fmt.Sprintf("[十二长生]...\n\n"+
	"[旬空-空亡]...\n\n"+
	"[神煞]\n年柱：%s | 月柱：%s | 日柱：%s | 时柱：%s\n",
	...
) +
gongJiaStr +
fmt.Sprintf("\n[大运序列]\n%s", dayunStr)
```

- [ ] **Step 4: Run prompt tests**

Run:

```bash
cd backend && go test ./internal/service -run 'TestBuildBaziPromptIncludesGongJiaContext|TestBuildPolishPrompt' -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit Task 3**

Run:

```bash
git add backend/internal/service/report_service.go backend/internal/service/report_service_test.go
git commit -m "feat(report): include gongjia prompt context"
```

## Task 4: Neutral Gongjia Event Signals

**Files:**
- Modify: `backend/pkg/bazi/event_signals.go`
- Modify: `backend/pkg/bazi/event_signals_test.go`

- [ ] **Step 1: Write failing event-signal tests**

Append to `backend/pkg/bazi/event_signals_test.go`:

```go
func TestGetYearEventSignals_GongJiaChongIsNeutral(t *testing.T) {
	natal := makeNatal("甲子", "甲寅", "庚午", "戊申", "火", "水")
	EnsureGongJia(natal)

	signals := GetYearEventSignals(natal, "己", "未", "戊辰", "female", 30)
	sig, ok := hasSignal(signals, "夹拱")
	if !ok {
		t.Fatalf("expected gongjia signal, got %+v", signals)
	}
	if sig.Polarity != PolarityNeutral || sig.Source != SourceGongJia {
		t.Fatalf("gongjia signal polarity/source = %s/%s, want 中性/夹拱: %+v", sig.Polarity, sig.Source, sig)
	}
	if !strings.Contains(sig.Evidence, "流年未冲原局年月夹支丑") {
		t.Fatalf("unexpected evidence: %s", sig.Evidence)
	}
}

func TestGetYearEventSignals_GongJiaSanheIsNeutral(t *testing.T) {
	natal := makeNatal("甲子", "甲寅", "庚巳", "戊申", "火", "水")
	EnsureGongJia(natal)

	signals := GetYearEventSignals(natal, "辛", "酉", "戊辰", "female", 30)
	sig, ok := hasSignal(signals, "夹拱")
	if !ok {
		t.Fatalf("expected gongjia sanhe signal, got %+v", signals)
	}
	if sig.Polarity != PolarityNeutral || sig.Source != SourceGongJia {
		t.Fatalf("gongjia signal polarity/source = %s/%s, want 中性/夹拱: %+v", sig.Polarity, sig.Source, sig)
	}
	if !strings.Contains(sig.Evidence, "夹支丑参与巳酉丑三合金局") {
		t.Fatalf("unexpected evidence: %s", sig.Evidence)
	}
}
```

If `event_signals_test.go` does not import `strings`, add it.

- [ ] **Step 2: Run tests to verify they fail**

Run:

```bash
cd backend && go test ./pkg/bazi -run 'TestGetYearEventSignals_GongJia' -count=1
```

Expected: FAIL because no `夹拱` event signal is produced.

- [ ] **Step 3: Add neutral collection helper**

In `backend/pkg/bazi/event_signals.go`, add source constant:

```go
SourceGongJia = "夹拱"
```

Add helper:

```go
func collectGongJiaSignals(natal *BaziResult, lnZhi, dyZhi string) []EventSignal {
	if natal == nil || len(natal.GongJia) == 0 {
		return nil
	}
	var sigs []EventSignal
	add := func(evidence string) {
		if evidence == "" {
			return
		}
		sigs = append(sigs, EventSignal{
			Type:     "夹拱",
			Evidence: evidence,
			Polarity: PolarityNeutral,
			Source:   SourceGongJia,
		})
	}
	for _, gj := range natal.GongJia {
		label := gongJiaSourceShortLabel(gj)
		for _, incoming := range []struct {
			prefix string
			zhi    string
		}{{"流年", lnZhi}, {"大运", dyZhi}} {
			if incoming.zhi == "" {
				continue
			}
			if c, ok := sixChong[incoming.zhi]; ok && c == gj.VirtualZhi {
				add(fmt.Sprintf("%s%s冲原局%s夹支%s，暗藏虚支被冲动，主意料之外但有迹可循的人事变化。", incoming.prefix, incoming.zhi, label, gj.VirtualZhi))
			}
			if h, ok := sixHe[incoming.zhi]; ok && h == gj.VirtualZhi {
				add(fmt.Sprintf("%s%s合原局%s夹支%s，夹支被合动，暗线关系、资源或贵人信号更容易显化。", incoming.prefix, incoming.zhi, label, gj.VirtualZhi))
			}
			if x, ok := sixXing[incoming.zhi]; ok && x == gj.VirtualZhi {
				add(fmt.Sprintf("%s%s刑原局%s夹支%s，暗藏虚支受刑，相关暗线压力被引动。", incoming.prefix, incoming.zhi, label, gj.VirtualZhi))
			}
			if selfXing[incoming.zhi] && incoming.zhi == gj.VirtualZhi {
				add(fmt.Sprintf("%s%s与原局%s夹支%s自刑，暗藏虚支自刑被引动。", incoming.prefix, incoming.zhi, label, gj.VirtualZhi))
			}
			if hai, ok := sixHai[incoming.zhi]; ok && hai == gj.VirtualZhi {
				add(fmt.Sprintf("%s%s害原局%s夹支%s，暗藏虚支受害，暗线人事容易浮出。", incoming.prefix, incoming.zhi, label, gj.VirtualZhi))
			}
			add(gongJiaJuEvidence(natal, gj, incoming.prefix, incoming.zhi))
		}
	}
	return sigs
}

func gongJiaSourceShortLabel(item GongJiaItem) string {
	switch item.Source {
	case "year_month":
		return "年月"
	case "month_day":
		return "月日"
	case "day_hour":
		return "日时"
	default:
		return strings.Join(item.SourceLabels, "")
	}
}

func gongJiaJuEvidence(natal *BaziResult, item GongJiaItem, prefix, incomingZhi string) string {
	if incomingZhi == "" {
		return ""
	}
	realZhi := []string{natal.YearZhi, natal.MonthZhi, natal.DayZhi, natal.HourZhi}
	for _, g := range allJuGroups {
		if !containsStr(g.branches[:], item.VirtualZhi) || !containsStr(g.branches[:], incomingZhi) {
			continue
		}
		needReal := false
		for _, z := range g.branches {
			if z != item.VirtualZhi && z != incomingZhi && containsStr(realZhi, z) {
				needReal = true
				break
			}
		}
		if !needReal {
			continue
		}
		juName := g.branches[0] + g.branches[1] + g.branches[2]
		return fmt.Sprintf("%s%s与原局真实地支、夹支%s参与%s%s%s局，夹支参与成局，暗藏力量被引出。",
			prefix, incomingZhi, item.VirtualZhi, juName, g.kind, wxPinyin2CN[g.wx])
	}
	return ""
}
```

If Go cannot pass `g.branches[:]` to `containsStr` because `g.branches` is an array field, assign it first:

```go
branches := g.branches[:]
```

In `GetYearEventSignalsWithContext`, after `collectJuShiSignals`, append:

```go
layer0Sigs = append(layer0Sigs, collectGongJiaSignals(natal, lnZhi, dyZhi)...)
```

- [ ] **Step 4: Run gongjia event tests**

Run:

```bash
cd backend && go test ./pkg/bazi -run 'TestGetYearEventSignals_GongJia' -count=1
```

Expected: PASS.

- [ ] **Step 5: Run nearby event tests**

Run:

```bash
cd backend && go test ./pkg/bazi -run 'TestGetYearEventSignals|Test.*JuShi|Test.*Yingqi' -count=1
```

Expected: PASS.

- [ ] **Step 6: Commit Task 4**

Run:

```bash
git add backend/pkg/bazi/event_signals.go backend/pkg/bazi/event_signals_test.go
git commit -m "feat(bazi): add neutral gongjia event signals"
```

## Task 5: Frontend Gongjia Panel

**Files:**
- Create: `frontend/src/components/GongJiaPanel.tsx`
- Create: `frontend/src/components/GongJiaPanel.css`
- Modify: `frontend/src/pages/ResultPage.tsx`
- Create: `frontend/tests/gongjia-panel.test.mjs`

- [ ] **Step 1: Write failing frontend test**

Create `frontend/tests/gongjia-panel.test.mjs`:

```js
import fs from 'node:fs'
import path from 'node:path'
import test from 'node:test'
import assert from 'node:assert/strict'

const root = path.resolve(import.meta.dirname, '..')

test('GongJiaPanel component labels virtual branch as hidden signal', () => {
  const file = fs.readFileSync(path.join(root, 'src/components/GongJiaPanel.tsx'), 'utf8')
  assert.match(file, /原局夹拱/)
  assert.match(file, /暗藏虚支/)
  assert.match(file, /不改原局五行与用神/)
  assert.match(file, /拱神煞/)
})

test('ResultPage renders GongJiaPanel without treating it as a fifth pillar', () => {
  const file = fs.readFileSync(path.join(root, 'src/pages/ResultPage.tsx'), 'utf8')
  assert.match(file, /GongJiaPanel/)
  assert.match(file, /gong_jia\?: GongJiaItem\[\]/)
  assert.doesNotMatch(file, /pillars\s*=\s*\[[\s\S]*gong_jia/)
})
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
cd frontend && node --test tests/gongjia-panel.test.mjs
```

Expected: FAIL because `GongJiaPanel.tsx` does not exist.

- [ ] **Step 3: Add component**

Create `frontend/src/components/GongJiaPanel.tsx`:

```tsx
import './GongJiaPanel.css'

export interface GongJiaItem {
  source: 'year_month' | 'month_day' | 'day_hour' | string
  source_labels?: string[]
  same_gan: string
  source_zhis?: string[]
  virtual_zhi: string
  hide_gan?: string[]
  shishen?: string[]
  shensha?: string[]
  meaning?: string
}

interface GongJiaPanelProps {
  items?: GongJiaItem[]
  onShenshaClick?: (name: string) => void
  hasShenshaAnnotation?: (name: string) => boolean
  shenshaPolarity?: (name: string) => string
}

export default function GongJiaPanel({
  items = [],
  onShenshaClick,
  hasShenshaAnnotation,
  shenshaPolarity,
}: GongJiaPanelProps) {
  if (!items.length) return null

  return (
    <section className="gongjia-panel card" aria-labelledby="gongjia-title">
      <div className="gongjia-panel__header">
        <div>
          <span className="result-panel-kicker">暗藏信号</span>
          <h2 id="gongjia-title" className="section-title serif">原局夹拱</h2>
        </div>
        <p>夹出之支为暗藏虚支，不改原局五行与用神，仅参与应期引动。</p>
      </div>

      <div className="gongjia-panel__grid">
        {items.map((item, index) => {
          const sourceLabel = item.source_labels?.join(' 与 ') || sourceLabelFallback(item.source)
          const zhis = item.source_zhis?.join(' / ') || ''
          return (
            <article className="gongjia-card" key={`${item.source}-${item.virtual_zhi}-${index}`}>
              <div className="gongjia-card__topline">
                <span>{sourceLabel}</span>
                <em>同干 {item.same_gan}</em>
              </div>
              <div className="gongjia-card__main">
                <span>{zhis}</span>
                <strong>夹出 {item.virtual_zhi}</strong>
              </div>
              <p>{item.meaning || `${sourceLabel}夹出${item.virtual_zhi}，${item.virtual_zhi}为暗藏虚支。`}</p>

              <MetaRow label="藏干" values={item.hide_gan} />
              <MetaRow label="十神" values={item.shishen} />

              {item.shensha && item.shensha.length > 0 && (
                <div className="gongjia-meta-row">
                  <span>拱神煞</span>
                  <div className="gongjia-tags">
                    {item.shensha.map(name => {
                      const clickable = hasShenshaAnnotation?.(name) ?? false
                      const polarity = shenshaPolarity?.(name) || 'zhong'
                      return (
                        <button
                          key={name}
                          type="button"
                          className={`shensha-tag shensha-tag--${polarity}${clickable ? ' shensha-tag--clickable' : ''}`}
                          onClick={() => clickable && onShenshaClick?.(name)}
                        >
                          {name}
                        </button>
                      )
                    })}
                  </div>
                </div>
              )}
            </article>
          )
        })}
      </div>
    </section>
  )
}

function MetaRow({ label, values = [] }: { label: string; values?: string[] }) {
  if (!values.length) return null
  return (
    <div className="gongjia-meta-row">
      <span>{label}</span>
      <div className="gongjia-tags">
        {values.map(value => <em key={value}>{value}</em>)}
      </div>
    </div>
  )
}

function sourceLabelFallback(source: string) {
  switch (source) {
    case 'year_month': return '年柱 与 月柱'
    case 'month_day': return '月柱 与 日柱'
    case 'day_hour': return '日柱 与 时柱'
    default: return '相邻两柱'
  }
}
```

Create `frontend/src/components/GongJiaPanel.css`:

```css
.gongjia-panel {
  margin-top: 16px;
  padding: 18px;
}

.gongjia-panel__header {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  align-items: flex-start;
  margin-bottom: 14px;
}

.gongjia-panel__header p {
  max-width: 420px;
  margin: 0;
  color: var(--text-muted);
  font-size: 14px;
  line-height: 1.7;
}

.gongjia-panel__grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 12px;
}

.gongjia-card {
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
  padding: 14px;
  background: var(--bg-elevated);
}

.gongjia-card__topline,
.gongjia-card__main,
.gongjia-meta-row {
  display: flex;
  justify-content: space-between;
  gap: 10px;
  align-items: center;
}

.gongjia-card__topline {
  color: var(--text-muted);
  font-size: 13px;
}

.gongjia-card__topline em {
  font-style: normal;
  color: var(--primary);
}

.gongjia-card__main {
  margin-top: 8px;
  font-size: 15px;
}

.gongjia-card__main strong {
  font-size: 22px;
}

.gongjia-card p {
  margin: 10px 0 12px;
  color: var(--text-muted);
  line-height: 1.7;
}

.gongjia-meta-row {
  align-items: flex-start;
  padding-top: 8px;
  border-top: 1px dashed var(--border-subtle);
  margin-top: 8px;
}

.gongjia-meta-row > span {
  color: var(--text-muted);
  font-size: 13px;
  white-space: nowrap;
}

.gongjia-tags {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 6px;
}

.gongjia-tags em {
  font-style: normal;
  border-radius: 999px;
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  padding: 2px 8px;
  font-size: 13px;
}

@media (max-width: 640px) {
  .gongjia-panel__header {
    display: block;
  }

  .gongjia-panel__header p {
    margin-top: 8px;
  }
}
```

- [ ] **Step 4: Wire ResultPage**

Modify imports in `frontend/src/pages/ResultPage.tsx`:

```tsx
import GongJiaPanel, { type GongJiaItem } from '../components/GongJiaPanel'
```

Add to `BaziResult`:

```tsx
gong_jia?: GongJiaItem[]
```

After the basic chart grid card and before `ten-god-relation-section`, render:

```tsx
<GongJiaPanel
  items={result.gong_jia || []}
  hasShenshaAnnotation={(name) => shenshaMap.has(name)}
  shenshaPolarity={(name) => SHENSHA_POLARITY[name] || 'zhong'}
  onShenshaClick={(name) => {
    const ann = shenshaMap.get(name)
    if (ann) setActiveAnnotation({ ...ann, description: `拱神煞：${ann.description}` })
  }}
/>
```

- [ ] **Step 5: Run frontend test**

Run:

```bash
cd frontend && node --test tests/gongjia-panel.test.mjs
```

Expected: PASS.

- [ ] **Step 6: Commit Task 5**

Run:

```bash
git add frontend/src/components/GongJiaPanel.tsx frontend/src/components/GongJiaPanel.css frontend/src/pages/ResultPage.tsx frontend/tests/gongjia-panel.test.mjs
git commit -m "feat(frontend): show gongjia virtual branches"
```

## Task 6: Final Integration Tests

**Files:**
- Modify only if tests expose a bug in files already touched by Tasks 1-5.

- [ ] **Step 1: Run focused backend package tests**

Run:

```bash
cd backend && go test ./pkg/bazi ./internal/service
```

Expected: PASS.

- [ ] **Step 2: Run frontend focused tests**

Run:

```bash
cd frontend && node --test tests/gongjia-panel.test.mjs tests/ten-god-relation-ux.test.mjs tests/result-page-readability.test.mjs
```

Expected: PASS.

- [ ] **Step 3: Run broader project verification**

Run:

```bash
cd backend && go test ./pkg/bazi ./internal/service ./internal/handler
```

Expected: PASS. If repository-dependent handler tests skip due to no database, record the skip lines in the final handoff.

Run:

```bash
cd frontend && npm run build
```

Expected: PASS with Vite build output and no TypeScript errors.

- [ ] **Step 4: Commit any test-fix follow-up**

If Step 1-3 required fixes, commit them:

```bash
git add backend/pkg/bazi backend/internal/service frontend/src frontend/tests
git commit -m "test: verify gongjia integration"
```

If no files changed, do not create an empty commit.

## Task 7: Completion Handoff

**Files:**
- No file changes unless execution revealed documentation drift.

- [ ] **Step 1: Confirm final status**

Run:

```bash
git status --short
```

Expected: clean worktree or only intentional uncommitted changes explicitly listed in the handoff.

- [ ] **Step 2: Summarize implemented behavior**

Final handoff must mention:

- `gong_jia` is generated only from adjacent natal pillars.
- The virtual branch does not alter wuxing, yongshen, mingge, tiaohou, or strength.
- Event signals from gongjia are neutral.
- Frontend displays gongjia as an暗藏信号 and not as a fifth pillar.
- Exact tests and build commands run.
