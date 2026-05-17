# 过往事件入口可见性 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Surface a "过往事件推算" entry card immediately below `DayunTimeline` so it is visible right after chart generation (no AI-report gating), with a disabled state for guests, and remove the redundant gated button at the bottom of `ResultPage`.

**Architecture:** Pure frontend change. Insert a new lightweight clickable card inside `<section className="dayun-section">` of `ResultPage.tsx`, after the `<DayunTimeline />` component. The card uses existing CSS variables and one new lucide-react icon (`History`). A new class `.past-events-entry` is added to `ResultPage.css` with hover / focus / disabled / mobile states. Two static-content tests in `frontend/tests/` enforce placement, behavior, and removal of the legacy button.

**Tech Stack:** React 19 + TypeScript + Vite, CSS variables (no UI framework), `node --test` with regex-on-source assertions for tests, lucide-react icons.

**Spec:** `docs/superpowers/specs/2026-05-17-past-events-entry-visibility-design.md`

---

## File Structure

| File | Responsibility | Change Type |
|------|----------------|-------------|
| `frontend/src/pages/ResultPage.tsx` | Render the page; mount new entry card; remove legacy button | Modify |
| `frontend/src/pages/ResultPage.css` | Visual styles for `.past-events-entry` | Modify |
| `frontend/tests/past-events-entry-visibility.test.mjs` | Static content tests enforcing placement & behavior | Create |

Zero backend, zero DDL, zero route changes.

---

## Task 0: Create feature branch

**Files:** (none — branch only)

- [ ] **Step 1: Verify working tree is clean and on main**

```bash
git status
git rev-parse --abbrev-ref HEAD
```

Expected:
- `nothing to commit, working tree clean`
- `main`

If dirty, stash or commit before proceeding.

- [ ] **Step 2: Create and check out the feature branch**

```bash
git checkout -b feat/past-events-entry-visibility
```

Expected: `Switched to a new branch 'feat/past-events-entry-visibility'`

- [ ] **Step 3: Verify base commit**

```bash
git log -1 --oneline
```

Expected first column SHA: `7761ef0` (`docs(specs): past-events entry visibility design`).

---

## Task 1: Write failing tests for entry visibility and legacy-button removal

**Files:**
- Create: `frontend/tests/past-events-entry-visibility.test.mjs`

- [ ] **Step 1: Create the test file with all assertions (RED)**

Create `frontend/tests/past-events-entry-visibility.test.mjs` with exactly this content:

```javascript
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import test from 'node:test'
import assert from 'node:assert/strict'

const root = new URL('..', import.meta.url).pathname

function read(path) {
  return readFileSync(join(root, path), 'utf8')
}

test('past-events entry is mounted inside dayun-section after DayunTimeline', () => {
  const page = read('src/pages/ResultPage.tsx')
  const dayunBlockMatch = page.match(/<section className="dayun-section">[\s\S]*?<\/section>/)
  assert.ok(dayunBlockMatch, 'dayun-section block not found')
  const block = dayunBlockMatch[0]
  assert.match(block, /<DayunTimeline[\s\S]*?\/>[\s\S]*?past-events-entry/)
})

test('past-events entry renders for logged-in users and disabled for guests', () => {
  const page = read('src/pages/ResultPage.tsx')
  assert.match(page, /past-events-entry/)
  assert.match(page, /isGuest/)
  assert.match(page, /登录后可查看/)
  assert.match(page, /展开每个大运段，看年份信号与白话批语/)
})

test('past-events entry navigates to /bazi/:chartId/past-events when enabled', () => {
  const page = read('src/pages/ResultPage.tsx')
  assert.match(
    page,
    /past-events-entry[\s\S]*?navigate\(`\/bazi\/\$\{targetId\}\/past-events`\)/,
  )
})

test('report-action-bar no longer contains a past-events navigation button', () => {
  const page = read('src/pages/ResultPage.tsx')
  const barMatch = page.match(/<div className="report-action-bar">[\s\S]*?<\/div>/)
  assert.ok(barMatch, 'report-action-bar block not found')
  assert.doesNotMatch(barMatch[0], /past-events/)
})

test('past-events-entry css defines hover, focus, and disabled states', () => {
  const css = read('src/pages/ResultPage.css')
  assert.match(css, /\.past-events-entry\s*\{/)
  assert.match(css, /\.past-events-entry:hover/)
  assert.match(css, /\.past-events-entry\.is-disabled/)
})
```

- [ ] **Step 2: Run the new test file and confirm all five tests fail**

```bash
cd frontend && node --test tests/past-events-entry-visibility.test.mjs
```

Expected:
- Test 1 fails (no `past-events-entry` in dayun-section)
- Test 2 fails (no `past-events-entry` token in file)
- Test 3 fails (no matching navigate call in entry block)
- Test 4 fails (`report-action-bar` currently *does* contain past-events button at `ResultPage.tsx:1130-1135`)
- Test 5 fails (no `.past-events-entry` class in CSS)
- Bottom summary should show `fail 5`.

- [ ] **Step 3: Commit the failing test**

```bash
git add frontend/tests/past-events-entry-visibility.test.mjs
git commit -m "test(result): past-events entry visibility (RED)"
```

---

## Task 2: Implement entry card + CSS so visibility tests pass

**Files:**
- Modify: `frontend/src/pages/ResultPage.tsx:3` (add `History` to lucide import)
- Modify: `frontend/src/pages/ResultPage.tsx:884-894` (insert entry card inside `dayun-section`, after `<DayunTimeline />`)
- Modify: `frontend/src/pages/ResultPage.css` (append `.past-events-entry` styles at end of file)

- [ ] **Step 1: Add `History` to the lucide-react import**

Find at `frontend/src/pages/ResultPage.tsx:3`:

```tsx
import { Diamond, X } from 'lucide-react'
```

Replace with:

```tsx
import { Diamond, X, History } from 'lucide-react'
```

- [ ] **Step 2: Insert the entry card inside `dayun-section`, after `<DayunTimeline />`**

Find at `frontend/src/pages/ResultPage.tsx:884-894`:

```tsx
            {/* 大运时间轴 */}
            <section className="dayun-section">
              <DayunTimeline
                dayun={result.dayun}
                birthYear={result.birth_year}
                startYunSolar={result.start_yun_solar}
                dayGan={result.day_gan || ''}
                gender={result.gender}
                pillarsLabel={dayunPillarsLabel}
                chartId={targetId}
              />
            </section>
```

Replace with:

```tsx
            {/* 大运时间轴 */}
            <section className="dayun-section">
              <DayunTimeline
                dayun={result.dayun}
                birthYear={result.birth_year}
                startYunSolar={result.start_yun_solar}
                dayGan={result.day_gan || ''}
                gender={result.gender}
                pillarsLabel={dayunPillarsLabel}
                chartId={targetId}
              />
              {(isGuest || targetId) && (
                <button
                  type="button"
                  className={`past-events-entry${isGuest ? ' is-disabled' : ''}`}
                  onClick={isGuest || !targetId ? undefined : () => navigate(`/bazi/${targetId}/past-events`)}
                  disabled={isGuest || !targetId}
                  aria-label="过往事件推算"
                >
                  <History className="past-events-entry-icon" size={22} aria-hidden="true" />
                  <span className="past-events-entry-body">
                    <span className="past-events-entry-title">过往事件推算</span>
                    <span className="past-events-entry-sub">
                      {isGuest ? '登录后可查看' : '展开每个大运段，看年份信号与白话批语'}
                    </span>
                  </span>
                  {!isGuest && (
                    <span className="past-events-entry-cta" aria-hidden="true">继续 →</span>
                  )}
                </button>
              )}
            </section>
```

- [ ] **Step 3: Append `.past-events-entry` styles to `ResultPage.css`**

Append to the end of `frontend/src/pages/ResultPage.css`:

```css

/* ── 过往事件推算入口 ─────────────────────────── */
.past-events-entry {
  display: flex;
  align-items: center;
  gap: 14px;
  width: 100%;
  margin-top: 16px;
  padding: 14px 18px;
  border-radius: var(--radius-md);
  background: var(--bg-card);
  border: 1px solid var(--border-default);
  cursor: pointer;
  text-align: left;
  font: inherit;
  color: inherit;
  transition: border-color 0.15s ease, background 0.15s ease;
}

.past-events-entry:hover:not(.is-disabled) {
  border-color: var(--border-accent);
  background: var(--bg-card-hover);
}

.past-events-entry:focus-visible {
  outline: 2px solid var(--wu-jin);
  outline-offset: 2px;
}

.past-events-entry.is-disabled {
  cursor: not-allowed;
  opacity: 0.55;
}

.past-events-entry-icon {
  flex: 0 0 auto;
  color: var(--wu-jin);
}

.past-events-entry-body {
  display: flex;
  flex-direction: column;
  gap: 4px;
  flex: 1 1 auto;
  min-width: 0;
}

.past-events-entry-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
}

.past-events-entry-sub {
  font-size: 13px;
  color: var(--text-secondary);
}

.past-events-entry-cta {
  flex: 0 0 auto;
  font-size: 13px;
  color: var(--wu-jin);
  font-weight: 500;
}

@media (max-width: 480px) {
  .past-events-entry {
    padding: 12px 14px;
    gap: 10px;
  }
  .past-events-entry-cta {
    display: none;
  }
}
```

- [ ] **Step 4: Run tests — expect 4 of 5 passing (legacy-button removal still fails)**

```bash
cd frontend && node --test tests/past-events-entry-visibility.test.mjs
```

Expected: `fail 1`, `pass 4`. The only failing test should be "report-action-bar no longer contains a past-events navigation button" because the legacy button is still in place — that gets removed in Task 3.

- [ ] **Step 5: Verify ResultPage still builds (TypeScript check)**

```bash
cd frontend && npx tsc -b --noEmit
```

Expected: no errors. If TypeScript complains about the `History` import (icon not exported), inspect `node_modules/lucide-react` exports — the symbol should exist.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/pages/ResultPage.tsx frontend/src/pages/ResultPage.css
git commit -m "feat(result): mount past-events entry card below DayunTimeline"
```

---

## Task 3: Remove the legacy past-events button from `report-action-bar`

**Files:**
- Modify: `frontend/src/pages/ResultPage.tsx:1130-1135` (delete the gated `<button>` block)

- [ ] **Step 1: Remove the legacy button**

Find at `frontend/src/pages/ResultPage.tsx:1130-1135`:

```tsx
              {user && targetId && (
                <button
                  className="btn btn-ghost report-action-highlight"
                  onClick={() => navigate(`/bazi/${targetId}/past-events`)}
                >过往事件</button>
              )}
```

Delete this entire block (6 lines). The surrounding `{user && <a href="/history" ...>}` (line 1129) and the next `<button ... onClick={handleExportPDF}>` (line 1136) should remain unchanged and now sit adjacent.

After deletion, the `report-action-bar` block should look like:

```tsx
          {report && (
            <div className="report-action-bar">
              <button className="btn btn-ghost" onClick={() => navigate('/')}>重新起盘</button>
              {user && <a href="/history" className="btn btn-ghost">查看历史</a>}
              <button
                className="btn btn-ghost"
                onClick={handleExportPDF}
                disabled={exportingPDF}
              >
                {exportingPDF ? '生成中...' : '导出 PDF'}
              </button>
            </div>
          )}
```

- [ ] **Step 2: Run tests — expect all five passing**

```bash
cd frontend && node --test tests/past-events-entry-visibility.test.mjs
```

Expected: `fail 0`, `pass 5`.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/pages/ResultPage.tsx
git commit -m "feat(result): remove legacy past-events button from action bar"
```

---

## Task 4: Full verification — full test suite, lint, build

**Files:** (none — verification only)

- [ ] **Step 1: Run the entire frontend test directory to catch any regression**

```bash
cd frontend && node --test tests/*.test.mjs
```

Expected: zero failures. The new test file should appear in output with 5 passes; pre-existing files should match their prior pass counts. If any pre-existing test fails, inspect — most likely the JSX rearrangement broke a regex in another test (e.g., a test that matched the bottom button by string presence). Diagnose root cause and fix the test or revert local change.

- [ ] **Step 2: Lint the frontend**

```bash
cd frontend && npm run lint
```

Expected: no errors. If lint flags the `History` import as unused, ensure step 1 of Task 2 was applied and the JSX in `dayun-section` uses `<History ... />`.

- [ ] **Step 3: Production build sanity check**

```bash
cd frontend && npm run build
```

Expected: build succeeds; no TypeScript errors.

- [ ] **Step 4: Smoke-check the result-layout-polish test still matches if it references `report-action-bar`**

```bash
cd frontend && grep -n "report-action-bar\|过往事件\|past-events" tests/result-layout-polish.test.mjs
```

If matches appear: read the test and assess whether the legacy-button removal invalidated any assertion. Fix tests that referenced the old layout (delete the now-stale assertion, do NOT re-add the button). If no matches, skip.

- [ ] **Step 5: Final commit if any test was adjusted in Step 4**

If a pre-existing test was edited:

```bash
git add frontend/tests/<adjusted-file>.test.mjs
git commit -m "test(result): drop stale past-events assertion in <file>"
```

If nothing was edited, skip this step.

---

## Verification Against Spec

| Spec requirement | Implemented in |
|------------------|----------------|
| §4.1 — Mount entry inside `dayun-section` after `DayunTimeline` | Task 2 Step 2 |
| §4.2 — Lightweight clickable card with icon, title, sub, CTA | Task 2 Steps 2-3 |
| §4.3 — Guest disabled with "登录后可查看"; logged-in clickable; no `targetId` not rendered | Task 2 Step 2 (`(isGuest || targetId)` gate; `disabled={isGuest || !targetId}`) |
| §4.4 — Delete legacy bottom button | Task 3 Step 1 |
| §5 — Files touched: `ResultPage.tsx`, `ResultPage.css`, new test mjs | Tasks 1-3 |
| §6.1 — Anonymous user sees disabled card with "登录后可查看" | Task 1 test 2 + Task 2 Step 2 |
| §6.2 — Logged-in user clicks and navigates | Task 1 test 3 + Task 2 Step 2 |
| §6.3 — Bottom bar no longer has past-events | Task 1 test 4 + Task 3 |
| §6.4 — Keyboard reachable via `<button>` with `disabled` | Task 2 Step 2 uses native `<button disabled>` (Tab/Enter/Space behave natively; disabled buttons skip focus per browser default) |
| §6.5 — Responsive ≤480px | Task 2 Step 3 (media query collapses CTA, tightens padding) |

All spec requirements have a backing task.
