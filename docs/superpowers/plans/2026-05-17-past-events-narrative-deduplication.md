# Past-Events Narrative Deduplication Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让相邻年开头不再逐字相同（信号驱动选词），并删除卡内底部 `年度力量：` 行（与主段内容重复）。

**Architecture:** 后端新增 `event_narrative_leads.go` 文件容纳 4 个 lead helper，按 primary 信号子类型/极性/Source/Evidence 选前导句；`event_narrative.go::yearToneSentence` 内 `isHardEventSignal` 分支改为分派到新 helper。前端删除 `PastEventsPage.tsx` 内 `年度力量：` JSX 块。零核心算法变更、零 API/DDL/缓存变更。

**Tech Stack:** Go 1.x（后端），React 19 + TypeScript（前端），`node --test` 用于前端静态正则断言。

---

## File Structure

| 文件 | 责任 | 操作 |
|------|------|------|
| `backend/pkg/bazi/event_narrative_leads.go` | 4 个 lead helper：`changeLead` / `healthLead` / `relationshipLead` / `defaultHardLead`，按信号字段选前导句 | 新建 |
| `backend/pkg/bazi/event_narrative_leads_test.go` | 5 个 Test_* 函数，验证各 lead 的分支差异化 + 兼容回归 | 新建 |
| `backend/pkg/bazi/event_narrative.go` | `yearToneSentence` 内 `isHardEventSignal` 分支由"一对一返回字符串"改为分派到 4 个 lead helper | 修改 125-135 行 |
| `frontend/src/pages/PastEventsPage.tsx` | 删除 `年度力量：` JSX 块（10 行） | 修改 420-429 行 |
| `frontend/tests/past-events-no-ten-god-footer.test.mjs` | 静态正则断言：源码不再含 `年度力量：` | 新建 |

---

## Task 0: Setup Feature Branch

**Files:** none yet

- [ ] **Step 1: Confirm clean working tree on main**

Run: `git -C /Users/liujiming/web/yuanju status`
Expected: `On branch main / Your branch is up to date with 'origin/main'. / nothing to commit, working tree clean`

- [ ] **Step 2: Cut feature branch from main**

Run: `git -C /Users/liujiming/web/yuanju checkout -b feat/past-events-dedup-narrative`
Expected: `Switched to a new branch 'feat/past-events-dedup-narrative'`

---

## Task 1: Backend RED — Failing Tests for Lead Helpers

**Files:**
- Create: `backend/pkg/bazi/event_narrative_leads_test.go`

- [ ] **Step 1: Write the failing test file**

Write `backend/pkg/bazi/event_narrative_leads_test.go` with this exact content:

```go
package bazi

import (
	"strings"
	"testing"
)

func assertAllDistinct(t *testing.T, name string, outs []string) {
	t.Helper()
	seen := map[string]int{}
	for i, s := range outs {
		if s == "" {
			t.Fatalf("%s output #%d is empty", name, i)
		}
		if prev, ok := seen[s]; ok {
			t.Fatalf("%s output #%d duplicates #%d: %q", name, i, prev, s)
		}
		seen[s] = i
	}
}

func TestChangeLeadDistinctBranches(t *testing.T) {
	inputs := []EventSignal{
		{Type: "伏吟", Polarity: PolarityXiong, Source: SourceFuyin},
		{Type: "反吟", Polarity: PolarityXiong, Source: SourceXing},
		{Type: "大运合化", Polarity: PolarityXiong, Source: SourceHehua},
		{Type: TypeJuShiZhong, Polarity: PolarityXiong},
		{Type: "综合变动", Polarity: PolarityXiong, Source: SourceXing, Evidence: "受刑"},
	}
	outs := make([]string, len(inputs))
	for i, in := range inputs {
		outs[i] = changeLead(in)
	}
	assertAllDistinct(t, "changeLead", outs)
}

func TestChangeLeadLegacyFallbackReachable(t *testing.T) {
	sig := EventSignal{
		Type:     "综合变动",
		Polarity: PolarityXiong,
		Source:   "",
		Evidence: "月柱受冲",
	}
	got := changeLead(sig)
	want := "这一年的变动感比较强，旧问题或突发调整容易被推到眼前"
	if got != want {
		t.Fatalf("changeLead legacy fallback = %q, want %q", got, want)
	}
}

func TestHealthLeadThreeBranches(t *testing.T) {
	inputs := []EventSignal{
		{Type: "健康", Polarity: PolarityXiong, Evidence: "白虎临运"},
		{Type: "健康", Polarity: PolarityXiong, Evidence: "羊刃临运"},
		{Type: "健康", Polarity: PolarityJi, Evidence: "天医临运"},
	}
	outs := make([]string, len(inputs))
	for i, in := range inputs {
		outs[i] = healthLead(in)
	}
	assertAllDistinct(t, "healthLead", outs)
}

func TestRelationshipLeadDistinctBranches(t *testing.T) {
	inputs := []EventSignal{
		{Type: "婚恋_合", Polarity: PolarityJi},
		{Type: "婚恋_冲", Polarity: PolarityXiong},
		{Type: "婚恋_变", Polarity: PolarityXiong},
		{Type: TypeXingGeQingYi, Polarity: PolarityJi},
		{Type: TypeXingGePanNi, Polarity: PolarityXiong},
	}
	outs := make([]string, len(inputs))
	for i, in := range inputs {
		outs[i] = relationshipLead(in)
	}
	assertAllDistinct(t, "relationshipLead", outs)
}

func TestDefaultHardLeadFourSources(t *testing.T) {
	inputs := []EventSignal{
		{Source: SourceKongwang, Polarity: PolarityXiong},
		{Source: SourceXing, Polarity: PolarityXiong},
		{Source: SourceFuyin, Polarity: PolarityXiong},
		{Source: SourceHehua, Polarity: PolarityXiong},
	}
	outs := make([]string, len(inputs))
	for i, in := range inputs {
		outs[i] = defaultHardLead(in)
	}
	assertAllDistinct(t, "defaultHardLead", outs)
}

// 防回归：旧的 "这一年的变动感比较强" 字符串必须仍被 changeLead 兜底分支保留
func TestLegacyChangeLeadStringStillPresent(t *testing.T) {
	src := EventSignal{Type: "综合变动", Polarity: PolarityXiong}
	if !strings.Contains(changeLead(src), "这一年的变动感比较强") {
		t.Fatalf("legacy change lead string must remain in fallback branch")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/liujiming/web/yuanju/backend && go test ./pkg/bazi/ -run "TestChangeLead|TestHealthLead|TestRelationshipLead|TestDefaultHardLead|TestLegacyChange" -v`
Expected: compile error mentioning undefined `changeLead` / `healthLead` / `relationshipLead` / `defaultHardLead`

- [ ] **Step 3: Commit RED test**

```bash
git -C /Users/liujiming/web/yuanju add backend/pkg/bazi/event_narrative_leads_test.go
git -C /Users/liujiming/web/yuanju commit -m "$(cat <<'EOF'
test(bazi): event_narrative lead helpers cover signal-driven branches (RED)

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 2: Backend GREEN — Implement Lead Helpers + Wire into yearToneSentence

**Files:**
- Create: `backend/pkg/bazi/event_narrative_leads.go`
- Modify: `backend/pkg/bazi/event_narrative.go:125-135`

- [ ] **Step 1: Create lead helper file**

Write `backend/pkg/bazi/event_narrative_leads.go` with this exact content:

```go
package bazi

import "strings"

// changeLead 按 primary 信号子类型/极性/Source 选择 change 主题前导句，避免跨年模板碰撞。
func changeLead(p EventSignal) string {
	switch p.Type {
	case "伏吟":
		return "这一年旧事容易卷土重来，过去搁置的处理可能重新冒头"
	case "反吟":
		return "这一年节奏变化会比较突兀，环境、计划或关系都可能临时倒挂"
	case "大运合化":
		return "这一年大运能量被牵动重组，方向感会有一次明显的调整"
	case TypeJuShiZhong:
		return "这一年整体力量容易被放大，一个选择容易牵动多条线"
	}
	if p.Polarity == PolarityXiong {
		if p.Source == SourceXing || strings.Contains(p.Evidence, "刑") {
			return "这一年容易在反复和细节里消耗，问题未必爆发却拖出余波"
		}
		return "这一年的变动感比较强，旧问题或突发调整容易被推到眼前"
	}
	if p.Polarity == PolarityJi {
		return "这一年变动中带着调整空间，主动顺势比被动应对更省力"
	}
	return "这一年节奏不算稳定，但调整中容易找到新方向"
}

// healthLead 按 Evidence/Polarity 选择 health 主题前导句。
func healthLead(p EventSignal) string {
	if strings.Contains(p.Evidence, "冲") || strings.Contains(p.Evidence, "白虎") {
		return "这一年身体和安全节奏需要被前置考虑，意外性消耗要避免"
	}
	if p.Polarity == PolarityXiong {
		return "健康、安全或日常节奏是这一年的主线，压力点会比较直接"
	}
	return "这一年身心提醒会更频繁，作息节律值得重新校准"
}

// relationshipLead 按 primary.Type 选择 relationship 主题前导句。
func relationshipLead(p EventSignal) string {
	switch p.Type {
	case "婚恋_合":
		return "这一年人际或感情的靠近感增强，关系节奏容易加快"
	case "婚恋_冲":
		return "这一年关系、距离和承诺容易被检验，节奏可能出现明显波动"
	case "婚恋_变":
		return "这一年情感或合作的方向容易调整，分寸感和边界都会被试探"
	case TypeXingGeQingYi:
		return "这一年情绪表达和人际反应会更外露，主动沟通比闷着推进有效"
	case TypeXingGePanNi:
		return "这一年个性主张容易和外部要求碰撞，关键节点上要稳住态度"
	}
	return "人际、感情或家庭沟通是这一年的主线，情绪触发会比较明显"
}

// defaultHardLead 按 primary.Source 选择硬事件兜底前导句。
func defaultHardLead(p EventSignal) string {
	switch p.Source {
	case SourceKongwang:
		return "这一年带着虚而不实的不稳定感，承诺和计划要多确认细节"
	case SourceXing:
		return "这一年有内耗反复的影子，事情未必爆发但容易消耗心力"
	case SourceFuyin:
		return "这一年旧主题反复回头，过去没处理完的事情会再被推上来"
	case SourceHehua:
		return "这一年大运能量被牵动，方向上的关键节点会比预想更明显"
	}
	return "这一年不是完全平稳的年份，关键事件会比平时更容易显形"
}
```

- [ ] **Step 2: Modify yearToneSentence to dispatch to lead helpers**

In `backend/pkg/bazi/event_narrative.go`, replace lines 125-135 (the `if isHardEventSignal(primary) { switch themeOf(primary.Type) { ... } }` block) with:

```go
	if isHardEventSignal(primary) {
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

Use Edit tool. Match exactly:

```
	if isHardEventSignal(primary) {
		switch themeOf(primary.Type) {
		case "health":
			return "健康、安全或日常节奏是这一年的主线，压力点会比较直接"
		case "change":
			return "这一年的变动感比较强，旧问题或突发调整容易被推到眼前"
		case "relationship":
			return "人际、感情或家庭沟通是这一年的主线，情绪触发会比较明显"
		default:
			return "这一年不是完全平稳的年份，关键事件会比平时更容易显形"
		}
	}
```

Replace with:

```
	if isHardEventSignal(primary) {
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

The polarity 5 分支（lines 137-148 in the original file) must remain unchanged.

- [ ] **Step 3: Run the new lead tests to verify GREEN**

Run: `cd /Users/liujiming/web/yuanju/backend && go test ./pkg/bazi/ -run "TestChangeLead|TestHealthLead|TestRelationshipLead|TestDefaultHardLead|TestLegacyChange" -v`
Expected: all 6 tests PASS (5 distinctness tests + 1 legacy fallback test)

- [ ] **Step 4: Run full bazi test suite to confirm no regression**

Run: `cd /Users/liujiming/web/yuanju/backend && go test ./pkg/bazi/...`
Expected: all tests PASS, no FAIL output

- [ ] **Step 5: Commit GREEN**

```bash
git -C /Users/liujiming/web/yuanju add backend/pkg/bazi/event_narrative_leads.go backend/pkg/bazi/event_narrative.go
git -C /Users/liujiming/web/yuanju commit -m "$(cat <<'EOF'
feat(bazi): dispatch year-tone leads by signal type/source for adjacent-year variation

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 3: Frontend RED — Failing Test for Footer Removal

**Files:**
- Create: `frontend/tests/past-events-no-ten-god-footer.test.mjs`

- [ ] **Step 1: Write the failing test file**

Write `frontend/tests/past-events-no-ten-god-footer.test.mjs` with this exact content:

```javascript
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import test from 'node:test'
import assert from 'node:assert/strict'

const root = new URL('..', import.meta.url).pathname

function read(path) {
  return readFileSync(join(root, path), 'utf8')
}

test('PastEventsPage no longer renders 年度力量 footer line', () => {
  const src = read('src/pages/PastEventsPage.tsx')
  assert.doesNotMatch(src, /年度力量：/)
})

test('PastEventsPage no longer references ten_god_power.plain_title in JSX', () => {
  const src = read('src/pages/PastEventsPage.tsx')
  // 命中模式：`y.ten_god_power?.plain_title` 不应在 JSX 渲染条件中出现
  assert.doesNotMatch(src, /ten_god_power\?\.plain_title/)
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/liujiming/web/yuanju/frontend && node --test tests/past-events-no-ten-god-footer.test.mjs`
Expected: both tests FAIL with messages about regex matching `年度力量：` and `ten_god_power?.plain_title` in PastEventsPage.tsx

- [ ] **Step 3: Commit RED test**

```bash
git -C /Users/liujiming/web/yuanju add frontend/tests/past-events-no-ten-god-footer.test.mjs
git -C /Users/liujiming/web/yuanju commit -m "$(cat <<'EOF'
test(past-events): assert 年度力量 footer removed from PastEventsPage (RED)

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 4: Frontend GREEN — Remove 年度力量 Footer Block

**Files:**
- Modify: `frontend/src/pages/PastEventsPage.tsx:420-429`

- [ ] **Step 1: Delete the 10-line JSX block**

In `frontend/src/pages/PastEventsPage.tsx`, remove lines 420-429 (the `{y.ten_god_power?.plain_title && (...)}` conditional block).

Use Edit tool. Match exactly:

```
                          {y.ten_god_power?.plain_title && (
                            <div style={{
                              marginTop: 8,
                              color: 'var(--text-muted)',
                              fontSize: '0.74rem',
                              lineHeight: 1.55,
                            }}>
                              年度力量：{y.ten_god_power.plain_title} - {y.ten_god_power.plain_text}
                            </div>
                          )}
```

Replace with empty string (i.e., delete the block entirely; the lines around it — `{y.narrative}` div above and `{hasEvidence && (...)}` below — stay intact, separated by their existing indentation).

- [ ] **Step 2: Run the frontend test to verify GREEN**

Run: `cd /Users/liujiming/web/yuanju/frontend && node --test tests/past-events-no-ten-god-footer.test.mjs`
Expected: both tests PASS

- [ ] **Step 3: Run build to confirm no TypeScript regression**

Run: `cd /Users/liujiming/web/yuanju/frontend && npm run build`
Expected: build completes successfully, no TS errors. (Note: `ten_god_power` field on the API type stays; we only stopped *rendering* it, so TS does not break.)

- [ ] **Step 4: Run lint**

Run: `cd /Users/liujiming/web/yuanju/frontend && npm run lint`
Expected: no new errors introduced by this change. (Pre-existing PrintLayout.tsx `no-irregular-whitespace` errors at lines 321/327 are out of scope — ignore those.)

- [ ] **Step 5: Commit GREEN**

```bash
git -C /Users/liujiming/web/yuanju add frontend/src/pages/PastEventsPage.tsx
git -C /Users/liujiming/web/yuanju commit -m "$(cat <<'EOF'
feat(past-events): drop duplicate 年度力量 footer from year card

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 5: Final Validation

**Files:** none modified

- [ ] **Step 1: Run full backend test suite**

Run: `cd /Users/liujiming/web/yuanju/backend && go test ./...`
Expected: PASS across all packages.

- [ ] **Step 2: Run full frontend test suite**

Run: `cd /Users/liujiming/web/yuanju/frontend && node --test tests/`
Expected: all `.test.mjs` files pass (including the new `past-events-no-ten-god-footer.test.mjs` and existing tests).

- [ ] **Step 3: Confirm clean branch state**

Run: `git -C /Users/liujiming/web/yuanju status`
Expected: `On branch feat/past-events-dedup-narrative / nothing to commit, working tree clean`

- [ ] **Step 4: Confirm commit chain**

Run: `git -C /Users/liujiming/web/yuanju log main..HEAD --oneline`
Expected output (4 commits, newest first):
```
<sha> feat(past-events): drop duplicate 年度力量 footer from year card
<sha> test(past-events): assert 年度力量 footer removed from PastEventsPage (RED)
<sha> feat(bazi): dispatch year-tone leads by signal type/source for adjacent-year variation
<sha> test(bazi): event_narrative lead helpers cover signal-driven branches (RED)
```

---

## Notes for Reviewer

- **Algorithm unchanged.** Signal generation (`event_signals.go`), dominant signal picking (`pickDominantSignal`), theme mapping (`themeOf`), hardness predicate (`isHardEventSignal`), polarity rank — all untouched. Only the prose template selection inside `yearToneSentence` changed.
- **API contract unchanged.** `ten_god_power` still flows through `PastEventsYearItem`; frontend just doesn't render the duplicate footer.
- **Cache unaffected.** Year narrative is rendered fresh per request (see `report_service.go:1035`), not cached. `ai_past_events` table caches AI 大运总结, which references but does not verbatim-copy 年 narrative, so existing cached summaries remain coherent.
- **Polarity 5-branch fallback untouched.** Years that don't hit `isHardEventSignal` fall into the existing 5 polarity branches at `event_narrative.go:137-148`. Those remain single-string per branch — out of scope per spec section 3.
