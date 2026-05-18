# Past-Events Narrative Evidence-Anchored Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Stop `RenderYearNarrative` from emitting paragraphs whose sentences came from polarity-count fallbacks or generic `default:` branches — hide the narrative paragraph entirely when fewer than 3 evidence-anchored sentences can be produced.

**Architecture:** Each of the 6 sentence builders in `event_narrative.go` is tightened so its fallback branches return `""` instead of generic prose. A new orchestrator counts non-empty sentences and returns `""` below threshold (`MinSentencesForNarrative = 3`). Frontend conditionally renders the narrative `<div>` only when `narrative !== ""`. No LLM, no DB, no API shape change.

**Tech Stack:** Go 1.21+ (`backend/pkg/bazi/`), `testing` package; React 19 + TypeScript (`frontend/src/pages/PastEventsPage.tsx`).

**Related spec:** `docs/superpowers/specs/2026-05-18-past-events-narrative-evidence-anchored-design.md`

---

## File Structure

**Modify (1 file, ~200 lines touched):**
- `backend/pkg/bazi/event_narrative.go`
  - Add `const MinSentencesForNarrative = 3` near top
  - Tighten 6 sentence builders: `yearToneSentence`, `triggerSourceSentence`, `domainDetailSentence`, `secondaryDetailSentence`, `tenGodNarrativeSentence`, `practicalStanceSentence`
  - Tighten 2 helpers: `richChangeSentence`, `richStudySentence` (drop their `default:` branches)
  - Rewrite `RenderYearNarrative` to count non-empty sentences and apply threshold; remove the "no meaningful signals → 兜底" branch

**Modify (1 file, ~150 lines added):**
- `backend/pkg/bazi/event_narrative_test.go`
  - Add 3 new test groups (hide threshold, screenshot regression, evidence-required contract)
  - Update 2 existing tests that assume current fallback behavior

**Modify (1 file, 3 lines touched):**
- `frontend/src/pages/PastEventsPage.tsx` (lines 417-419) — wrap narrative `<div>` in `{y.narrative && (...)}`

**Not touched:**
- `event_signals.go`, `event_narrative_leads.go`, `event_narrative_test.go`'s `TestRenderEvidenceSummary*`
- Backend handlers, services, repositories
- Database, API shape
- Dead code (`plainThemeSentence`, `changeSentence`, `schoolSentence`, `practicalReminder`) — out of scope

---

## Task 1: Tighten `yearToneSentence` — drop polarity-only branches

**Files:**
- Modify: `backend/pkg/bazi/event_narrative.go:115-149`
- Test: `backend/pkg/bazi/event_narrative_test.go` (append)

- [ ] **Step 1: Add the failing contract test**

Append to `backend/pkg/bazi/event_narrative_test.go`:

```go
func TestYearToneSentence_PolarityOnlyReturnsEmpty(t *testing.T) {
	cases := []struct {
		name    string
		signals []EventSignal
		primary EventSignal
	}{
		{
			name: "xiong>=2 ji>0 without hard primary",
			signals: []EventSignal{
				{Type: "综合变动", Polarity: PolarityXiong, Source: SourceZhuwei, Evidence: "节奏变化"},
				{Type: "综合变动", Polarity: PolarityXiong, Source: SourceZhuwei, Evidence: "节奏变化"},
				{Type: "综合变动", Polarity: PolarityJi, Source: SourceZhuwei, Evidence: "节奏变化"},
			},
			primary: EventSignal{Type: "综合变动", Polarity: PolarityXiong, Source: SourceZhuwei, Evidence: "节奏变化"},
		},
		{
			name: "all xiong without hard primary",
			signals: []EventSignal{
				{Type: "综合变动", Polarity: PolarityXiong, Source: SourceZhuwei, Evidence: "节奏变化"},
				{Type: "综合变动", Polarity: PolarityXiong, Source: SourceZhuwei, Evidence: "节奏变化"},
			},
			primary: EventSignal{Type: "综合变动", Polarity: PolarityXiong, Source: SourceZhuwei, Evidence: "节奏变化"},
		},
		{
			name: "all ji without hard primary",
			signals: []EventSignal{
				{Type: "综合变动", Polarity: PolarityJi, Source: SourceZhuwei, Evidence: "节奏变化"},
				{Type: "综合变动", Polarity: PolarityJi, Source: SourceZhuwei, Evidence: "节奏变化"},
			},
			primary: EventSignal{Type: "综合变动", Polarity: PolarityJi, Source: SourceZhuwei, Evidence: "节奏变化"},
		},
		{
			name:    "no signals",
			signals: nil,
			primary: EventSignal{Type: "综合变动", Polarity: PolarityNeutral, Source: SourceZhuwei, Evidence: "节奏变化"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := yearToneSentence(c.signals, c.primary); got != "" {
				t.Errorf("expected empty string, got %q", got)
			}
		})
	}
}

func TestYearToneSentence_HardSignalStillEmits(t *testing.T) {
	primary := EventSignal{
		Type:     "健康",
		Polarity: PolarityXiong,
		Source:   SourceZhuwei,
		Evidence: "流年地支午冲日支子，日柱受冲",
	}
	got := yearToneSentence([]EventSignal{primary}, primary)
	if got == "" {
		t.Fatal("expected non-empty hard-signal lead, got empty")
	}
}
```

- [ ] **Step 2: Run new tests to verify they fail**

```bash
cd backend && go test ./pkg/bazi -run 'TestYearToneSentence_PolarityOnlyReturnsEmpty' -v
```

Expected: FAIL — at least one subcase returns a non-empty polarity-count sentence (e.g., `"这一年有机会也有压力..."`).

- [ ] **Step 3: Tighten the implementation**

Replace `backend/pkg/bazi/event_narrative.go:115-149` with:

```go
func yearToneSentence(signals []EventSignal, primary EventSignal) string {
	if !isHardEventSignal(primary) {
		return ""
	}
	switch themeOf(primary.Type) {
	case "health":
		return healthLead(primary)
	case "change":
		return changeLead(primary)
	case "relationship":
		return relationshipLead(primary)
	default:
		return defaultHardLead(primary)
	}
}
```

(The unused `signals` parameter stays for now — Task 8 won't touch the signature; future cleanup can remove it.)

- [ ] **Step 4: Run new tests to verify they pass**

```bash
cd backend && go test ./pkg/bazi -run 'TestYearToneSentence' -v
```

Expected: PASS for both new tests.

- [ ] **Step 5: Run full bazi package to spot existing-test breakage**

```bash
cd backend && go test ./pkg/bazi
```

Expected output: at least the following existing tests are likely to fail because they assume polarity-count opener exists:
- `TestRenderYearNarrative_RichSignalYearHasMediumDetail` may complain about "取舍" not in narrative (yearTone removed). Check whether "取舍" still appears in `practicalStanceSentence` output for that input (it does — see line 683 "稳责任边界和取舍" ✓), so should still pass.
- `TestRenderYearNarrative_AdjacentYoungYearsDoNotRepeatGenericChangeOpening` may have empty narratives now (will be reconciled in Task 8).

Do **not** fix existing tests yet; that's Task 8's job. As long as the new `TestYearToneSentence_*` tests pass, this task is done.

- [ ] **Step 6: Commit**

```bash
cd backend && git add pkg/bazi/event_narrative.go pkg/bazi/event_narrative_test.go
git -c commit.gpgsign=false commit -m "$(cat <<'EOF'
refactor(narrative): yearToneSentence drops polarity-only branches

Only emit a tone opener when primary signal is hard-event (kongwang /
xing / fuyin / hehua / 用神位 / 忌神位 / 受冲 / 受刑 / 力度倍增 /
大运流年双重命中). Polarity counts alone no longer trigger generic
"这一年有机会也有压力" type openers.

EOF
)"
```

---

## Task 2: Tighten `triggerSourceSentence` — drop default branch

**Files:**
- Modify: `backend/pkg/bazi/event_narrative.go:151-178`
- Test: `backend/pkg/bazi/event_narrative_test.go` (append)

- [ ] **Step 1: Add the failing contract test**

Append to `backend/pkg/bazi/event_narrative_test.go`:

```go
func TestTriggerSourceSentence_NoKeywordReturnsEmpty(t *testing.T) {
	cases := []struct {
		name string
		sig  EventSignal
		age  int
	}{
		{
			name: "no keyword in evidence",
			sig:  EventSignal{Type: "综合变动", Evidence: "节奏一般变化", Source: SourceZhuwei},
			age:  30,
		},
		{
			name: "empty evidence and neutral type",
			sig:  EventSignal{Type: "综合变动", Evidence: "", Source: SourceZhuwei},
			age:  15,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := triggerSourceSentence(c.sig, c.age); got != "" {
				t.Errorf("expected empty string, got %q", got)
			}
		})
	}
}

func TestTriggerSourceSentence_KeywordStillEmits(t *testing.T) {
	cases := []EventSignal{
		{Type: "事业", Evidence: "流年地支辰冲月柱壬戌", Source: SourceZhuwei},
		{Type: "综合变动", Evidence: "落空亡，虚而不实", Source: SourceKongwang},
		{Type: "伏吟", Evidence: "流年壬辰伏吟日柱壬辰", Source: SourceFuyin},
	}
	for i, c := range cases {
		if got := triggerSourceSentence(c, 30); got == "" {
			t.Errorf("case %d: expected non-empty, got empty for %+v", i, c)
		}
	}
}
```

- [ ] **Step 2: Run new tests to verify they fail**

```bash
cd backend && go test ./pkg/bazi -run 'TestTriggerSourceSentence_NoKeywordReturnsEmpty' -v
```

Expected: FAIL — current `default:` branch emits "触发点来自这一年的主导信号..."

- [ ] **Step 3: Tighten the implementation**

In `backend/pkg/bazi/event_narrative.go:151-178`, change the final `default:` branch from emitting a sentence to returning `""`:

```go
	default:
		return ""
	}
}
```

Concretely, replace:

```go
	default:
		return "触发点来自这一年的主导信号，事情不会只停留在想法层面，容易落实到具体安排。"
	}
}
```

with:

```go
	default:
		return ""
	}
}
```

- [ ] **Step 4: Run new tests to verify they pass**

```bash
cd backend && go test ./pkg/bazi -run 'TestTriggerSourceSentence' -v
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
cd backend && git add pkg/bazi/event_narrative.go pkg/bazi/event_narrative_test.go
git -c commit.gpgsign=false commit -m "$(cat <<'EOF'
refactor(narrative): triggerSourceSentence drops generic default

Only emit a trigger sentence when evidence contains a recognized
keyword (冲/刑/空/伏吟/反吟/局/势力/用神/喜神/忌神/驿马/奔波/迁移)
or signal source/type matches. Generic "触发点来自这一年的主导信号"
fallback removed.

EOF
)"
```

---

## Task 3: Tighten `domainDetailSentence` + drop `richChangeSentence` / `richStudySentence` defaults

**Files:**
- Modify: `backend/pkg/bazi/event_narrative.go:180-289`
- Test: `backend/pkg/bazi/event_narrative_test.go` (append)

- [ ] **Step 1: Add the failing contract test**

Append to `backend/pkg/bazi/event_narrative_test.go`:

```go
func TestDomainDetailSentence_UnknownThemeReturnsEmpty(t *testing.T) {
	primary := EventSignal{Type: "未知类型", Evidence: "无关键词", Source: SourceZhuwei}
	if got := domainDetailSentence(primary, EventSignal{}, false, 30); got != "" {
		t.Errorf("expected empty for unknown theme, got %q", got)
	}
}

func TestRichChangeSentence_NoAnchorReturnsEmpty(t *testing.T) {
	sig := EventSignal{Type: "综合变动", Evidence: "节奏一般变化", Source: SourceZhuwei, Polarity: PolarityNeutral}
	if got := richChangeSentence(sig); got != "" {
		t.Errorf("expected empty for un-anchored change signal, got %q", got)
	}
}

func TestRichStudySentence_UnknownStudyTypeReturnsEmpty(t *testing.T) {
	primary := EventSignal{Type: "事业", Evidence: "无关键词", Source: SourceZhuwei}
	if got := richStudySentence(primary, EventSignal{}, false); got != "" {
		t.Errorf("expected empty for unknown study type, got %q", got)
	}
}
```

- [ ] **Step 2: Run new tests to verify they fail**

```bash
cd backend && go test ./pkg/bazi -run 'TestDomainDetailSentence_UnknownThemeReturnsEmpty|TestRichChangeSentence_NoAnchorReturnsEmpty|TestRichStudySentence_UnknownStudyTypeReturnsEmpty' -v
```

Expected: FAIL — all three currently emit `default:` fallbacks.

- [ ] **Step 3: Drop default branches in three places**

In `backend/pkg/bazi/event_narrative.go:218-220`, replace the `default:` branch of `domainDetailSentence`:

```go
	default:
		return "现实表现上，日常安排会出现新的侧重点，适合多观察变化，再决定推进顺序。"
	}
}
```

with:

```go
	default:
		return ""
	}
}
```

In `backend/pkg/bazi/event_narrative.go:286-288`, replace the `default:` branch of `richChangeSentence`:

```go
	default:
		return "现实表现上，事情会被推动起来，适合顺势整理方向，把该确认的安排先确认清楚。"
	}
}
```

with:

```go
	default:
		return ""
	}
}
```

In `backend/pkg/bazi/event_narrative.go:267-269`, replace the `default:` branch of `richStudySentence`:

```go
	default:
		return "现实表现上，学习、日常规则和个人状态会更受关注，按节奏积累比急着突破更重要。"
	}
}
```

with:

```go
	default:
		return ""
	}
}
```

- [ ] **Step 4: Run new tests to verify they pass**

```bash
cd backend && go test ./pkg/bazi -run 'TestDomainDetailSentence_UnknownThemeReturnsEmpty|TestRichChangeSentence_NoAnchorReturnsEmpty|TestRichStudySentence_UnknownStudyTypeReturnsEmpty' -v
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
cd backend && git add pkg/bazi/event_narrative.go pkg/bazi/event_narrative_test.go
git -c commit.gpgsign=false commit -m "$(cat <<'EOF'
refactor(narrative): drop default branches in domain/change/study sentences

domainDetailSentence, richChangeSentence, richStudySentence each had a
generic default branch that emitted vague prose when no theme/keyword
matched. All three now return "" when the input doesn't carry specific
evidence to anchor on.

EOF
)"
```

---

## Task 4: Tighten `secondaryDetailSentence` — gate on secondary's own evidence anchor

**Files:**
- Modify: `backend/pkg/bazi/event_narrative.go:223-252`
- Test: `backend/pkg/bazi/event_narrative_test.go` (append)

- [ ] **Step 1: Add the failing contract test**

Append to `backend/pkg/bazi/event_narrative_test.go`:

```go
func TestSecondaryDetailSentence_UnanchoredSignalReturnsEmpty(t *testing.T) {
	cases := []struct {
		name string
		sig  EventSignal
	}{
		{
			name: "vague 综合变动 with no keyword",
			sig:  EventSignal{Type: "综合变动", Evidence: "节奏一般变化", Source: SourceZhuwei, Polarity: PolarityNeutral},
		},
		{
			name: "vague 喜神临运 with no anchor keyword",
			sig:  EventSignal{Type: "喜神临运", Evidence: "印星生身", Source: SourceZhuwei, Polarity: PolarityJi},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := secondaryDetailSentence(c.sig, 30); got != "" {
				t.Errorf("expected empty, got %q", got)
			}
		})
	}
}

func TestSecondaryDetailSentence_AnchoredSignalStillEmits(t *testing.T) {
	cases := []EventSignal{
		// (a) hard event signal
		{Type: "健康", Evidence: "流年地支午冲日支子，日柱受冲", Source: SourceZhuwei, Polarity: PolarityXiong},
		// (b) evidence keyword
		{Type: "财运_得", Evidence: "财星为忌神，破耗", Source: SourceZhuwei, Polarity: PolarityXiong},
		// (c) signal type in allowed set
		{Type: "伏吟", Evidence: "伏吟", Source: SourceFuyin, Polarity: PolarityXiong},
		{Type: TypeXueYeYaLi, Evidence: "学业要求", Source: SourceZhuwei, Polarity: PolarityNeutral},
		{Type: "婚恋_冲", Evidence: "婚恋冲", Source: SourceZhuwei, Polarity: PolarityXiong},
	}
	for i, c := range cases {
		if got := secondaryDetailSentence(c, 30); got == "" {
			t.Errorf("case %d (Type=%s): expected non-empty, got empty", i, c.Type)
		}
	}
}
```

- [ ] **Step 2: Run new tests to verify they fail**

```bash
cd backend && go test ./pkg/bazi -run 'TestSecondaryDetailSentence_UnanchoredSignalReturnsEmpty' -v
```

Expected: FAIL — current `secondaryDetailSentence` returns text for any known theme regardless of evidence.

- [ ] **Step 3: Add the anchor gate**

In `backend/pkg/bazi/event_narrative.go` near the top of the file (after the existing `package bazi` and imports, before `RenderYearNarrative`), add a new helper:

```go
// hasEvidenceAnchor returns true when sig carries a specific differentiator
// that a sentence can cite: a hard-event Source, an allowed Type, or a
// recognized keyword inside Evidence. Used to gate secondary-detail prose
// so it does not pad narratives with generic theme wording.
func hasEvidenceAnchor(sig EventSignal) bool {
	if isHardEventSignal(sig) {
		return true
	}
	switch sig.Type {
	case "伏吟", "反吟", "大运合化", TypeJuShiZhong:
		return true
	}
	if strings.HasPrefix(sig.Type, "学业_") || strings.HasPrefix(sig.Type, "性格_") || strings.HasPrefix(sig.Type, "婚恋_") {
		return true
	}
	keywords := []string{"冲", "刑", "空", "用神", "忌神", "驿马", "月柱", "提纲", "日支", "自我宫位", "大运流年双重命中", "意外", "白虎"}
	for _, k := range keywords {
		if strings.Contains(sig.Evidence, k) {
			return true
		}
	}
	return false
}
```

Then in `backend/pkg/bazi/event_narrative.go:223-224`, add the gate as the first line of `secondaryDetailSentence`:

```go
func secondaryDetailSentence(sig EventSignal, age int) string {
	if !hasEvidenceAnchor(sig) {
		return ""
	}
	theme := themeOf(sig.Type)
	switch theme {
```

- [ ] **Step 4: Run new tests to verify they pass**

```bash
cd backend && go test ./pkg/bazi -run 'TestSecondaryDetailSentence' -v
```

Expected: both new tests PASS.

- [ ] **Step 5: Commit**

```bash
cd backend && git add pkg/bazi/event_narrative.go pkg/bazi/event_narrative_test.go
git -c commit.gpgsign=false commit -m "$(cat <<'EOF'
refactor(narrative): secondaryDetailSentence requires evidence anchor

Introduce hasEvidenceAnchor() and gate secondaryDetailSentence on it.
A secondary signal must be a hard event, have an allowed Type, or
carry a recognized evidence keyword to be eligible for "同时,..." prose;
otherwise it stays out of the narrative entirely.

EOF
)"
```

---

## Task 5: Tighten `tenGodNarrativeSentence` — drop generic background fallback

**Files:**
- Modify: `backend/pkg/bazi/event_narrative.go:291-303`
- Test: `backend/pkg/bazi/event_narrative_test.go` (append)

- [ ] **Step 1: Add the failing contract test**

Append to `backend/pkg/bazi/event_narrative_test.go`:

```go
func TestTenGodNarrativeSentence_NoGroupAlignmentReturnsEmpty(t *testing.T) {
	power := TenGodPowerProfile{
		PlainTitle: "官杀偏旺",
		PlainText:  "规则、责任和外部压力更明显",
		Group:      "", // no group → tenGodGroupTheme returns ""
	}
	primary := EventSignal{Type: "综合变动", Evidence: "节奏变化", Source: SourceZhuwei}
	if got := tenGodNarrativeSentence(power, primary, EventSignal{}, false); got != "" {
		t.Errorf("expected empty when 10-god group has no theme alignment, got %q", got)
	}
}

func TestTenGodNarrativeSentence_GroupAlignedStillEmits(t *testing.T) {
	power := TenGodPowerProfile{
		PlainTitle: "财星偏旺",
		PlainText:  "钱财、资源、合作回报更明显",
		Group:      TenGodGroupWealth, // wealth → money theme
	}
	primary := EventSignal{Type: "财运_得", Evidence: "财来财去", Source: SourceZhuwei}
	if got := tenGodNarrativeSentence(power, primary, EventSignal{}, false); got == "" {
		t.Error("expected non-empty when 10-god group aligns with primary theme")
	}
}
```

- [ ] **Step 2: Run new tests to verify they fail**

```bash
cd backend && go test ./pkg/bazi -run 'TestTenGodNarrativeSentence' -v
```

Expected: `TestTenGodNarrativeSentence_NoGroupAlignmentReturnsEmpty` FAILS — current code's third return is the generic "可作为理解这一年事件走向的背景力量" fallback.

- [ ] **Step 3: Drop the generic fallback**

In `backend/pkg/bazi/event_narrative.go:291-303`, replace:

```go
func tenGodNarrativeSentence(power TenGodPowerProfile, primary EventSignal, secondary EventSignal, hasSecondary bool) string {
	if power.PlainTitle == "" || power.PlainText == "" {
		return ""
	}
	groupTheme := tenGodGroupTheme(power.Group)
	if isHardEventSignal(primary) && (groupTheme == "" || groupTheme == themeOf(primary.Type)) {
		return ""
	}
	if groupTheme != "" && (groupTheme == themeOf(primary.Type) || (hasSecondary && groupTheme == themeOf(secondary.Type))) {
		return "这股年度力量会把" + tenGodPlainDomain(power.Group, power.Polarity) + "推到台前，处理得好可以成为助力，处理得急则容易变成压力。"
	}
	return power.PlainTitle + "，" + strings.TrimSuffix(power.PlainText, "。") + "，可作为理解这一年事件走向的背景力量。"
}
```

with:

```go
func tenGodNarrativeSentence(power TenGodPowerProfile, primary EventSignal, secondary EventSignal, hasSecondary bool) string {
	if power.PlainTitle == "" || power.PlainText == "" {
		return ""
	}
	groupTheme := tenGodGroupTheme(power.Group)
	if groupTheme == "" {
		return ""
	}
	if isHardEventSignal(primary) && groupTheme == themeOf(primary.Type) {
		return ""
	}
	if groupTheme == themeOf(primary.Type) || (hasSecondary && groupTheme == themeOf(secondary.Type)) {
		return "这股年度力量会把" + tenGodPlainDomain(power.Group, power.Polarity) + "推到台前，处理得好可以成为助力，处理得急则容易变成压力。"
	}
	return ""
}
```

- [ ] **Step 4: Run new tests to verify they pass**

```bash
cd backend && go test ./pkg/bazi -run 'TestTenGodNarrativeSentence' -v
```

Expected: both new tests PASS.

- [ ] **Step 5: Commit**

```bash
cd backend && git add pkg/bazi/event_narrative.go pkg/bazi/event_narrative_test.go
git -c commit.gpgsign=false commit -m "$(cat <<'EOF'
refactor(narrative): tenGodNarrativeSentence requires group-theme alignment

Drop the generic "可作为理解这一年事件走向的背景力量" wrap that fired
whenever a 10-god power existed but didn't align with any year signal.
The sentence now only emits when the 10-god group theme matches the
primary or secondary signal's theme — otherwise stays out of narrative.

EOF
)"
```

---

## Task 6: Tighten `practicalStanceSentence` — drop polarity-only fallback

**Files:**
- Modify: `backend/pkg/bazi/event_narrative.go:645-694`
- Test: `backend/pkg/bazi/event_narrative_test.go` (append)

- [ ] **Step 1: Add the failing contract test**

Append to `backend/pkg/bazi/event_narrative_test.go`:

```go
func TestPracticalStanceSentence_UnknownThemeReturnsEmpty(t *testing.T) {
	primary := EventSignal{Type: "未知类型", Polarity: PolarityXiong, Source: SourceZhuwei}
	if got := practicalStanceSentence([]EventSignal{primary}, primary, 30); got != "" {
		t.Errorf("expected empty for unknown theme, got %q", got)
	}
}
```

- [ ] **Step 2: Run the new test to verify it fails**

```bash
cd backend && go test ./pkg/bazi -run 'TestPracticalStanceSentence_UnknownThemeReturnsEmpty' -v
```

Expected: FAIL — current code falls through to `return practicalReminder(signals)`, which emits polarity-count prose.

- [ ] **Step 3: Drop the trailing polarity-only fallback**

In `backend/pkg/bazi/event_narrative.go:693`, replace:

```go
	return practicalReminder(signals)
}
```

with:

```go
	return ""
}
```

- [ ] **Step 4: Run the new test to verify it passes**

```bash
cd backend && go test ./pkg/bazi -run 'TestPracticalStanceSentence_UnknownThemeReturnsEmpty' -v
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
cd backend && git add pkg/bazi/event_narrative.go pkg/bazi/event_narrative_test.go
git -c commit.gpgsign=false commit -m "$(cat <<'EOF'
refactor(narrative): practicalStanceSentence drops polarity-only fallback

The terminal `return practicalReminder(signals)` was a polarity-count
escape hatch — it fired whenever the dominant signal's theme wasn't
in the known set. Removed; unknown themes now contribute no closer
sentence to narrative.

EOF
)"
```

---

## Task 7: Add `MinSentencesForNarrative` const + rewrite `RenderYearNarrative`

**Files:**
- Modify: `backend/pkg/bazi/event_narrative.go:1-38`
- Test: `backend/pkg/bazi/event_narrative_test.go` (append)

- [ ] **Step 1: Add the failing threshold test**

Append to `backend/pkg/bazi/event_narrative_test.go`:

```go
func TestRenderYearNarrative_HiddenWhenBelowThreshold(t *testing.T) {
	// Two signals, both un-anchored — every builder returns "".
	ys := YearSignals{
		Year:   2010,
		Age:    11,
		GanZhi: "庚寅",
		Signals: []EventSignal{
			{Type: "综合变动", Evidence: "节奏一般变化", Source: SourceZhuwei, Polarity: PolarityNeutral},
			{Type: "综合变动", Evidence: "另一个变化", Source: SourceZhuwei, Polarity: PolarityNeutral},
		},
	}
	if got := RenderYearNarrative(ys); got != "" {
		t.Errorf("expected empty narrative below threshold, got %q", got)
	}
}

func TestRenderYearNarrative_NoSignalsReturnsEmpty(t *testing.T) {
	// No meaningful signals at all — old code emitted a tengod context fallback
	// or "本年命理信号较弱" stub; new code returns "".
	ys := YearSignals{Year: 2022, Age: 27, GanZhi: "壬寅"}
	if got := RenderYearNarrative(ys); got != "" {
		t.Errorf("expected empty narrative for no-signals year, got %q", got)
	}
}

func TestRenderYearNarrative_MeetsThresholdWhenAnchored(t *testing.T) {
	// Hard health signal: yearTone (healthLead), trigger (冲), domain (health
	// with 冲 keyword), practical (health) — at least 3 anchored sentences.
	ys := YearSignals{
		Year:   2026,
		Age:    31,
		GanZhi: "丙午",
		Signals: []EventSignal{
			{Type: "健康", Evidence: "流年地支午冲日支子，日柱受冲，体力精神有下滑风险", Polarity: PolarityXiong, Source: SourceZhuwei},
		},
	}
	got := RenderYearNarrative(ys)
	if got == "" {
		t.Fatal("expected narrative to render for hard health signal year")
	}
	if !strings.HasPrefix(got, "丙午年，") {
		t.Errorf("expected narrative to start with GanZhi prefix, got: %s", got)
	}
}
```

- [ ] **Step 2: Run new tests to verify they fail**

```bash
cd backend && go test ./pkg/bazi -run 'TestRenderYearNarrative_HiddenWhenBelowThreshold|TestRenderYearNarrative_NoSignalsReturnsEmpty' -v
```

Expected: FAIL — current `RenderYearNarrative` emits the "本年命理信号较弱..." or "整体动象不强..." fallback strings.

- [ ] **Step 3: Add the threshold constant**

In `backend/pkg/bazi/event_narrative.go` immediately after `package bazi` and `import "strings"` (around line 4), add:

```go
// MinSentencesForNarrative is the minimum number of non-empty
// evidence-anchored sentences required before RenderYearNarrative
// returns a paragraph. Below this threshold the narrative is hidden
// (returns ""), and the frontend renders only the signal chips and
// evidence summary for that year.
const MinSentencesForNarrative = 3
```

- [ ] **Step 4: Rewrite `RenderYearNarrative`**

In `backend/pkg/bazi/event_narrative.go:7-38`, replace:

```go
// RenderYearNarrative 根据 EventSignal 列表生成面向用户的白话批语。
// 底层 Evidence 保留给 RenderEvidenceSummary，不直接暴露在默认正文中。
func RenderYearNarrative(ys YearSignals) string {
	if len(meaningfulSignals(ys.Signals)) == 0 {
		if s := tenGodContextSentence(ys.TenGodPower); s != "" {
			return ys.GanZhi + "年，" + s + "整体节奏不必急，适合顺着这股力量稳步安排。"
		}
		return "本年命理信号较弱，运势相对平稳，无明显重大变动。"
	}

	primary, ok := pickDominantSignal(ys.Signals, "", ys.Age)
	if !ok {
		if s := tenGodContextSentence(ys.TenGodPower); s != "" {
			return ys.GanZhi + "年，" + s + "整体动象不算强，但方向感会更清楚。"
		}
		return ys.GanZhi + "年整体动象不强，适合按部就班推进，保持稳定节奏。"
	}
	secondary, hasSecondary := pickDominantSignal(ys.Signals, themeOf(primary.Type), ys.Age)

	parts := []string{
		ys.GanZhi + "年，" + yearToneSentence(ys.Signals, primary),
		triggerSourceSentence(primary, ys.Age),
		domainDetailSentence(primary, secondary, hasSecondary, ys.Age),
	}
	if hasSecondary {
		parts = append(parts, secondaryDetailSentence(secondary, ys.Age))
	}
	if s := tenGodNarrativeSentence(ys.TenGodPower, primary, secondary, hasSecondary); s != "" {
		parts = append(parts, s)
	}
	parts = append(parts, practicalStanceSentence(ys.Signals, primary, ys.Age))

	return joinNarrativeParts(parts)
}
```

with:

```go
// RenderYearNarrative 根据 EventSignal 列表生成面向用户的白话批语。
// 底层 Evidence 保留给 RenderEvidenceSummary，不直接暴露在默认正文中。
//
// 当能产出的"有 evidence 支撑"的句子数少于 MinSentencesForNarrative 时
// 返回空串，前端会跳过 narrative 段落，只渲染徽标和命理依据。
func RenderYearNarrative(ys YearSignals) string {
	primary, ok := pickDominantSignal(ys.Signals, "", ys.Age)
	if !ok {
		return ""
	}
	secondary, hasSecondary := pickDominantSignal(ys.Signals, themeOf(primary.Type), ys.Age)

	sentences := make([]string, 0, 6)
	if s := yearToneSentence(ys.Signals, primary); s != "" {
		sentences = append(sentences, s)
	}
	if s := triggerSourceSentence(primary, ys.Age); s != "" {
		sentences = append(sentences, s)
	}
	if s := domainDetailSentence(primary, secondary, hasSecondary, ys.Age); s != "" {
		sentences = append(sentences, s)
	}
	if hasSecondary {
		if s := secondaryDetailSentence(secondary, ys.Age); s != "" {
			sentences = append(sentences, s)
		}
	}
	if s := tenGodNarrativeSentence(ys.TenGodPower, primary, secondary, hasSecondary); s != "" {
		sentences = append(sentences, s)
	}
	if s := practicalStanceSentence(ys.Signals, primary, ys.Age); s != "" {
		sentences = append(sentences, s)
	}

	if len(sentences) < MinSentencesForNarrative {
		return ""
	}
	return joinNarrativeParts(append([]string{ys.GanZhi + "年，" + sentences[0]}, sentences[1:]...))
}
```

(The GanZhi prefix attaches to the first surviving sentence rather than always being its own slot — that way "丙午年，触发点..." reads naturally when yearTone is empty.)

- [ ] **Step 5: Run new tests to verify they pass**

```bash
cd backend && go test ./pkg/bazi -run 'TestRenderYearNarrative_HiddenWhenBelowThreshold|TestRenderYearNarrative_NoSignalsReturnsEmpty|TestRenderYearNarrative_MeetsThresholdWhenAnchored' -v
```

Expected: all three PASS.

- [ ] **Step 6: Run full bazi package to see existing-test damage**

```bash
cd backend && go test ./pkg/bazi
```

Expected: a small number of existing tests now fail. Note them — Task 8 fixes them. The expected failing set is:
- `TestRenderYearNarrative_AdjacentYoungYearsDoNotRepeatGenericChangeOpening` (some years return empty, breaking uniqueness assertion)
- `TestRenderYearNarrative_TenGodPowerEnrichesGenericYear` (asserts a fallback that no longer fires)

Other tests should still pass — verify and write down anything unexpected.

- [ ] **Step 7: Commit**

```bash
cd backend && git add pkg/bazi/event_narrative.go pkg/bazi/event_narrative_test.go
git -c commit.gpgsign=false commit -m "$(cat <<'EOF'
feat(narrative): hide RenderYearNarrative below MinSentencesForNarrative

Add MinSentencesForNarrative const (=3). Restructure RenderYearNarrative
to collect non-empty sentences from the six tightened builders and
return "" when fewer than 3 remain. Remove the two top-level fallback
branches ("本年命理信号较弱" / "整体动象不强") — user preference is to
hide narrative entirely when the algorithm has no specific evidence to
cite, rather than emit generic prose.

GanZhi prefix now attaches to the first surviving sentence so phrases
like "丙午年，触发点..." read naturally when yearTone is hidden.

EOF
)"
```

---

## Task 8: Reconcile existing tests broken by the new contract

**Files:**
- Modify: `backend/pkg/bazi/event_narrative_test.go:74-117` and `event_narrative_test.go:232-251`

- [ ] **Step 1: Update `TestRenderYearNarrative_AdjacentYoungYearsDoNotRepeatGenericChangeOpening`**

In `backend/pkg/bazi/event_narrative_test.go:74-117`, replace the existing function body (lines 105-116) — the part that loops and asserts unique openings — to accept empty narratives:

```go
	openings := map[string]bool{}
	for _, ys := range years {
		narrative := RenderYearNarrative(ys)
		// Empty narrative is OK under the new contract — it means the year
		// has no evidence-anchored sentences to show. Only enforce
		// uniqueness on years that actually render.
		if narrative == "" {
			continue
		}
		if strings.Contains(narrative, "变化感会比较强") {
			t.Fatalf("young-age narrative used generic repeated change opening: %s", narrative)
		}
		opening := firstSentence(narrative)
		if openings[opening] {
			t.Fatalf("repeated opening sentence %q for narrative: %s", opening, narrative)
		}
		openings[opening] = true
	}
```

- [ ] **Step 2: Update `TestRenderYearNarrative_TenGodPowerEnrichesGenericYear`**

The test at `backend/pkg/bazi/event_narrative_test.go:232-251` asserts the old "background force" fallback fires when only `综合变动` is present. Under the new contract this case must hide the narrative entirely. Rewrite the test to assert hiding (this becomes a regression test for the spec's user preference "宁可不显示"):

Replace the function `TestRenderYearNarrative_TenGodPowerEnrichesGenericYear` (entire function, lines 232-251) with:

```go
func TestRenderYearNarrative_TenGodPowerDoesNotRescueGenericYear(t *testing.T) {
	// Old behavior: a 10-god power title appended a "...可作为理解这一年事件
	// 走向的背景力量。" wrap, padding generic years into a visible paragraph.
	// New behavior (per 2026-05-18 spec): un-anchored years stay hidden
	// regardless of 10-god power, so the algorithm doesn't fill silence
	// with generic prose.
	ys := YearSignals{
		Year:   2024,
		Age:    29,
		GanZhi: "甲辰",
		TenGodPower: TenGodPowerProfile{
			PlainTitle: "官杀偏旺",
			PlainText:  "规则、考核、责任和外部压力更明显，宜稳住节奏。",
			Reason:     "流年天干为七杀",
		},
		Signals: []EventSignal{
			{Type: "综合变动", Evidence: "流年节奏变化", Polarity: PolarityNeutral, Source: SourceZhuwei},
		},
	}
	if got := RenderYearNarrative(ys); got != "" {
		t.Errorf("expected hidden narrative for un-anchored year with 10-god power; got %q", got)
	}
}
```

- [ ] **Step 3: Run full bazi package — should be green**

```bash
cd backend && go test ./pkg/bazi
```

Expected: all tests PASS. If any other tests fail unexpectedly, investigate before continuing — the symptom is likely a test that relied on the old polarity-count fallback. Fix by either (a) updating the test to expect the new behavior, or (b) tightening the test input to anchor on real evidence. Do NOT relax the new contract.

- [ ] **Step 4: Commit**

```bash
cd backend && git add pkg/bazi/event_narrative_test.go
git -c commit.gpgsign=false commit -m "$(cat <<'EOF'
test(narrative): reconcile two existing tests with evidence-anchored contract

- AdjacentYoungYearsDoNotRepeatGenericChangeOpening: skip empty
  narratives in the uniqueness assertion (empty is the new "no specific
  evidence" signal, not a repeat).
- TenGodPowerEnrichesGenericYear → renamed to
  TenGodPowerDoesNotRescueGenericYear, now asserts that an un-anchored
  year stays hidden even when a 10-god power has been computed.

EOF
)"
```

---

## Task 9: Add screenshot regression test

**Files:**
- Modify: `backend/pkg/bazi/event_narrative_test.go` (append)

- [ ] **Step 1: Add the regression test**

Append to `backend/pkg/bazi/event_narrative_test.go`:

```go
func TestRenderYearNarrative_ScreenshotRegression_RepetitiveOpenerHidden(t *testing.T) {
	// Reproduces the 2026-05-18 screenshot: three adjacent child-age years
	// (乙酉 2005 / 丙戌 2006 / 丁亥 2007) where the old template emitted
	// "这一年有机会也有压力，事情会同时出现可争取和需取舍的一面" as the
	// opener for ALL THREE. Under the evidence-anchored contract, none of
	// them carry hard signals — all three narratives should be hidden.
	years := []YearSignals{
		{
			Year:   2005,
			Age:    10,
			GanZhi: "乙酉",
			Signals: []EventSignal{
				{Type: "综合变动", Evidence: "流年地支酉与原局子时刑（空亡相邻）", Polarity: PolarityXiong, Source: SourceKongwang},
				{Type: TypeXueYeJingZheng, Evidence: "乙木为命主比劫，少年期同学比较增强", Polarity: PolarityNeutral, Source: SourceZhuwei},
				{Type: TypeXingGeQingYi, Evidence: "流年地支酉为桃花星临命", Polarity: PolarityNeutral, Source: SourceZhuwei},
			},
		},
		{
			Year:   2006,
			Age:    11,
			GanZhi: "丙戌",
			Signals: []EventSignal{
				{Type: "综合变动", Evidence: "流年节奏一般变化", Polarity: PolarityNeutral, Source: SourceZhuwei},
				{Type: TypeXueYeGuiRen, Evidence: "丙透干为食神，少年期表达能力突出", Polarity: PolarityJi, Source: SourceZhuwei},
				{Type: TypeXingGeQingYi, Evidence: "流年地支戌合卯木", Polarity: PolarityNeutral, Source: SourceZhuwei},
				{Type: "健康", Evidence: "流年节奏微调", Polarity: PolarityXiong, Source: SourceZhuwei},
			},
		},
		{
			Year:   2007,
			Age:    12,
			GanZhi: "丁亥",
			Signals: []EventSignal{
				{Type: "综合变动", Evidence: "流年节奏一般变化", Polarity: PolarityXiong, Source: SourceZhuwei},
				{Type: TypeXueYeGuiRen, Evidence: "丁透干印星，少年期师长缘", Polarity: PolarityJi, Source: SourceZhuwei},
				{Type: "健康", Evidence: "流年微调", Polarity: PolarityXiong, Source: SourceZhuwei},
			},
		},
	}

	bannedFillers := []string{
		"这一年有机会也有压力",
		"触发点来自这一年的主导信号",
		"这一年最要紧的",
		"本年命理信号较弱",
		"可作为理解这一年事件走向的背景力量",
	}

	openings := map[string]string{} // opening → ganzhi that emitted it
	for _, ys := range years {
		narrative := RenderYearNarrative(ys)
		for _, banned := range bannedFillers {
			if strings.Contains(narrative, banned) {
				t.Fatalf("%s narrative contains banned filler %q: %s", ys.GanZhi, banned, narrative)
			}
		}
		if narrative == "" {
			continue // hidden cards are fine — we want that
		}
		opening := firstSentence(narrative)
		if prev, seen := openings[opening]; seen {
			t.Fatalf("repeated opening %q across years %s and %s", opening, prev, ys.GanZhi)
		}
		openings[opening] = ys.GanZhi
	}
}
```

- [ ] **Step 2: Run the regression test**

```bash
cd backend && go test ./pkg/bazi -run 'TestRenderYearNarrative_ScreenshotRegression_RepetitiveOpenerHidden' -v
```

Expected: PASS. (All three years should be hidden, so the loop body's "if narrative == "" continue" branch is taken every time, and the banned-filler scan trivially passes for empty strings.)

- [ ] **Step 3: Run full bazi package — final green check**

```bash
cd backend && go test ./pkg/bazi
```

Expected: ALL tests PASS.

- [ ] **Step 4: Commit**

```bash
cd backend && git add pkg/bazi/event_narrative_test.go
git -c commit.gpgsign=false commit -m "$(cat <<'EOF'
test(narrative): pin 2026-05-18 screenshot regression

Reproduces the three-consecutive-years case (乙酉/丙戌/丁亥, ages 10-12)
where the old template repeated "这一年有机会也有压力..." as the opener.
Asserts: no banned filler strings leak; openings of any rendered years
are mutually distinct.

EOF
)"
```

---

## Task 10: Frontend — conditional render of narrative `<div>`

**Files:**
- Modify: `frontend/src/pages/PastEventsPage.tsx:417-419`

- [ ] **Step 1: Apply the conditional render**

In `frontend/src/pages/PastEventsPage.tsx:417-419`, replace:

```tsx
                          <div style={{ color: 'var(--text-secondary)', fontSize: '0.85rem', lineHeight: 1.7 }}>
                            {y.narrative}
                          </div>
```

with:

```tsx
                          {y.narrative && (
                            <div style={{ color: 'var(--text-secondary)', fontSize: '0.85rem', lineHeight: 1.7 }}>
                              {y.narrative}
                            </div>
                          )}
```

- [ ] **Step 2: Type-check the frontend**

```bash
cd frontend && npm run build
```

Expected: build succeeds, no TypeScript errors. The `y.narrative` type is `string` (see `PastEventsPage.tsx:17`), so the truthy check correctly treats empty string as "hide".

- [ ] **Step 3: Smoke the frontend lint**

```bash
cd frontend && npm run lint
```

Expected: 0 new lint errors. (If ESLint complains about `&&` short-circuit returning `string` instead of `boolean`, that's a pre-existing pattern in this codebase — leave as-is.)

- [ ] **Step 4: Commit**

```bash
git add frontend/src/pages/PastEventsPage.tsx
git -c commit.gpgsign=false commit -m "$(cat <<'EOF'
feat(past-events): hide narrative div when backend returns empty string

Backend RenderYearNarrative now returns "" for years where it cannot
produce >=3 evidence-anchored sentences. Frontend skips rendering the
narrative <div> in those cases — the year card keeps its GanZhi badge,
signal chips, and 命理依据 expander.

EOF
)"
```

---

## Task 11: Manual chart verification

**Files:** none modified — verification only.

- [ ] **Step 1: Boot the stack**

```bash
cd /Users/liujiming/web/yuanju && ./scripts/docker-compose-up.sh
```

Wait for the compose stack to be ready (postgres + backend on :9002 + frontend nginx on :5200). Expected output ends with all three services in "running" state.

- [ ] **Step 2: Open a child-age-heavy chart in browser**

In a logged-in browser session at `http://localhost:5200`, open any chart that includes ages 8-17. Click "过往事件推算" entry. If you don't have a usable chart, create one with birth datetime `1996-02-08 20:00` (this is the fixture the prior `fix-past-events-repetitive-narrative` change used).

- [ ] **Step 3: Visual check — pre/post comparison**

Open DevTools → Network → call `POST /api/bazi/past-events/years/<chart_id>` once via the page, inspect the JSON response. Confirm:
- **Some `years[].narrative` values are `""`** (these are the hidden cards). Expect roughly 30-60% of child-age years (≤17岁) to have empty narrative, depending on the chart's signal density.
- **Years that DO have narrative** start with `<GanZhi>年，` followed by a sentence that visibly cites a specific evidence keyword (冲, 刑, 空, 用神, 忌神, 月柱, 日支, 驿马, etc.) or signal type (健康, 伏吟, 学业_*, etc.).
- **No two consecutive rendered years share an opening sentence.** Scroll through the rendered timeline and visually spot-check.
- **Cards with empty narrative still show** the GanZhi badge, year/age line, signal chips, and "命理依据 ▾" expander.

- [ ] **Step 4: Confirm scope unchanged**

In the same page:
- 大运 group headers still show the AI-generated themes + summary
- "命理依据 ▾" expander still expands and shows raw evidence rows
- Signal chips (婚恋↑ / 健康↓ / 学业↑ / 喜神 / etc.) still appear above the narrative when relevant

If any of these are broken, revert the relevant task's commit and investigate.

- [ ] **Step 5: Document the verification**

No commit; this is a verification step. Note the chart ID(s) used and rough percentage of hidden cards in the PR description when the work goes for review.

---

## Self-Review

- **Spec coverage:** Walked through each section of `2026-05-18-past-events-narrative-evidence-anchored-design.md`:
  - § "决策" 1-5 → Tasks 1-7, 10 cover all five decisions
  - § "Evidence-anchored 定义" table → Tasks 1-6 implement each row, Task 4 introduces `hasEvidenceAnchor()`
  - § "编排器" pseudocode → Task 7 implements verbatim
  - § "前端" → Task 10
  - § "测试" 1-3 + manual checklist → Tasks 2-7 each add the contract test (item 3); Task 7 + Task 9 add hide threshold (item 1) and screenshot regression (item 2); Task 11 covers the manual checklist
  - § "范围边界" → "Not touched" section in File Structure aligns
  - § "风险" → no task needed; risks are operational notes, not implementation steps

- **Placeholder scan:** No "TBD", "TODO", "fill in", "appropriate error handling", or "similar to Task N" appears. Every code block is complete. Every command is executable as written.

- **Type / signature consistency:** `hasEvidenceAnchor(sig EventSignal) bool` introduced in Task 4 is referenced only in Task 4. `MinSentencesForNarrative` introduced in Task 7 step 3 is referenced only in Task 7 step 4. No drift across tasks. Existing function signatures (`yearToneSentence`, `triggerSourceSentence`, etc.) are preserved — only bodies change.

- **Order soundness:** Tasks 1-6 each tighten one builder with its own contract test that passes in isolation; existing tests may flicker but are only reconciled in Task 8 after the orchestrator changes land (Task 7). This is intentional — having one consolidated reconciliation task avoids whack-a-mole.
