# Compat Chart Picker UX Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Upgrade the compatibility page's "从命盘档案选择" modal picker so users can identify and select the right chart in under 3 seconds, even with 10+ saved charts.

**Architecture:** Surgical upgrade inside the existing modal. Extract shared chart-label pure functions into a new util module (`frontend/src/lib/chartLabel.ts`), migrate HistoryPage to use it, then add picker-local state (`pickerSearch`, `pickerGenderFilter`), a `useMemo`-derived list with `isSelectedHere` / `isUsedByOtherSide` flags, a new JSX layout (toolbar + listbox + role-aware empty states), focused autoFocus on the search input, and new CSS for toolbar/badges plus a mobile full-screen media query at ≤640px.

**Tech Stack:** React 19 + TypeScript (strict) + Vite + React Router 7. No frontend test framework — verification via `npm run build` (which runs `tsc -b` + `vite build`), `npm run lint`, and manual UI testing on `npm run dev`. Spec: `docs/superpowers/specs/2026-05-28-compat-chart-picker-ux-design.md`.

---

## File Structure

| File | Purpose | Action |
|------|---------|--------|
| `frontend/src/lib/chartLabel.ts` | Pure helpers: `genderText`, `chartFallbackName`, `chartDisplayName`, `formatPillars`, `formatBirth`. Imported by HistoryPage + CompatibilityPage. | CREATE |
| `frontend/src/pages/HistoryPage.tsx` | Replace local helper definitions (lines 10-29) with imports from chartLabel. No behavior change. | MODIFY |
| `frontend/src/pages/CompatibilityPage.tsx` | Picker upgrade: state, useMemo, new JSX, autoFocus retargeting. | MODIFY |
| `frontend/src/pages/CompatibilityPage.css` | Tighten legacy `.item span/small` selectors, add toolbar/badges/pillars styles, add 640px full-screen media query. | MODIFY |

---

## Task 1: Extract `chartLabel.ts` shared util

**Files:**
- Create: `frontend/src/lib/chartLabel.ts`

- [ ] **Step 1: Create the util file**

Create `frontend/src/lib/chartLabel.ts` with this exact content:

```ts
import type { BaziHistoryChart } from './api'

export function genderText(gender: string): string {
  return gender === 'female' ? '女命' : '男命'
}

export function chartFallbackName(chart: BaziHistoryChart): string {
  return `${genderText(chart.gender)} · ${chart.birth_year}年${chart.birth_month}月${chart.birth_day}日`
}

export function chartDisplayName(chart: BaziHistoryChart): string {
  return chart.display_name?.trim() || chartFallbackName(chart)
}

export function formatPillars(chart: BaziHistoryChart): string {
  return `${chart.year_gan}${chart.year_zhi} · ${chart.month_gan}${chart.month_zhi} · ${chart.day_gan}${chart.day_zhi} · ${chart.hour_gan}${chart.hour_zhi}`
}

export function formatBirth(chart: BaziHistoryChart): string {
  return `${chart.birth_year}年${chart.birth_month}月${chart.birth_day}日 ${chart.birth_hour}时`
}
```

- [ ] **Step 2: Verify the type import resolves**

Run from `/Users/liujiming/web/yuanju/frontend`:
```
npx tsc --noEmit -p .
```
Expected: exit code 0. If `BaziHistoryChart` import path differs, open `frontend/src/lib/api.ts` and grep `export.*BaziHistoryChart` to confirm the named export exists — it does, used by CompatibilityPage and HistoryPage.

- [ ] **Step 3: Commit**

```bash
cd /Users/liujiming/web/yuanju
git add frontend/src/lib/chartLabel.ts
git commit -m "$(cat <<'EOF'
feat(chart-label): extract shared chart label helpers

Pure helpers (genderText / chartFallbackName / chartDisplayName /
formatPillars / formatBirth) consolidated into a single module so
HistoryPage and the upcoming compatibility-picker upgrade can share
the same fallback/formatting logic.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 2: Migrate HistoryPage to use `chartLabel.ts`

**Files:**
- Modify: `frontend/src/pages/HistoryPage.tsx:1-29`

- [ ] **Step 1: Replace the helper definitions with an import**

Open `frontend/src/pages/HistoryPage.tsx`. Replace lines 1-29 (the imports block plus the 4 local helper functions `formatDate`, `genderText`, `formatPillars`, `chartFallbackName`, `chartDisplayName`) so the top of the file reads:

```tsx
import { useEffect, useRef, useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { CalendarDays, Compass, HeartHandshake, Sparkles } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'
import { baziAPI, type BaziHistoryChart } from '../lib/api'
import { chartDisplayName, chartFallbackName, formatPillars, genderText } from '../lib/chartLabel'
import './HistoryPage.css'

type Chart = BaziHistoryChart

function formatDate(value: string) {
  if (!value) return '-'
  return new Date(value).toLocaleDateString('zh-CN')
}

function isInteractiveTarget(target: EventTarget | null) {
  return target instanceof HTMLElement && Boolean(target.closest('button, input, select, textarea, a'))
}
```

Note: `formatDate` stays local (it's HistoryPage-specific and not on the shared util). `isInteractiveTarget` stays local. The body of the component (line 35 onward) does not change — `genderText(c.gender)`, `formatPillars(c)`, `chartDisplayName(c)`, `chartFallbackName(c)` calls now resolve to the imported versions.

- [ ] **Step 2: Verify HistoryPage still type-checks and builds**

Run from `/Users/liujiming/web/yuanju/frontend`:
```
npm run build
```
Expected: exit code 0. Build output shows transformed modules including `chartLabel.ts`.

- [ ] **Step 3: Manual smoke — HistoryPage still renders**

Run from `/Users/liujiming/web/yuanju/frontend`:
```
npm run dev
```
Open `http://localhost:5173/history` while logged in. Confirm:
- Each card shows the correct display name (or fallback `女命 · 1995年2月2日` form)
- Pillars line shows `庚午 · 庚寅 · 庚子 · 丙戌` etc.
- Edit-name flow still works (no regression)

Stop the dev server after verification.

- [ ] **Step 4: Commit**

```bash
cd /Users/liujiming/web/yuanju
git add frontend/src/pages/HistoryPage.tsx
git commit -m "$(cat <<'EOF'
refactor(history): use shared chartLabel helpers

No behavior change — HistoryPage's local genderText / chartFallbackName /
chartDisplayName / formatPillars are replaced by imports from
lib/chartLabel.ts to prepare for the same helpers being consumed by
the compat picker.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 3: Upgrade the compatibility chart picker

**Files:**
- Modify: `frontend/src/pages/CompatibilityPage.tsx` (imports, state, `openChartPicker`, autoFocus effect, picker JSX block)

- [ ] **Step 1: Add new imports**

Open `frontend/src/pages/CompatibilityPage.tsx`. Update line 1 (the React import) and add new imports below the existing block:

Change line 1 from:
```ts
import { useCallback, useEffect, useRef, useState } from 'react'
```
to:
```ts
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
```

Then add a new import line below the existing `birthProfile` import (after line 19):
```ts
import { Link } from 'react-router-dom'
import { chartDisplayName, formatBirth, formatPillars, genderText } from '../lib/chartLabel'
```

Note: `useNavigate` and `useSearchParams` are already imported from `react-router-dom` on line 2; the new `Link` import joins that line. Final form of line 2:
```ts
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
```

- [ ] **Step 2: Add picker-local state**

Inside the `CompatibilityPage` component, locate the existing state declarations (lines 73-79). Immediately after `const [error, setError] = useState('')`, add:

```ts
const [pickerSearch, setPickerSearch] = useState('')
const [pickerGenderFilter, setPickerGenderFilter] = useState<'all' | 'male' | 'female'>('all')
```

- [ ] **Step 3: Update `openChartPicker` to set role-aware filter default**

Locate `openChartPicker` (currently CompatibilityPage.tsx:184-188). Replace its body so it resets picker-local state when opening:

```ts
const openChartPicker = async (role: 'self' | 'partner') => {
  if (isLoading) return
  setPickerSearch('')
  if (role === 'partner') {
    setPickerGenderFilter(selfProfile.gender === 'male' ? 'female' : 'male')
  } else {
    setPickerGenderFilter(selfProfile.gender === 'male' ? 'male' : 'female')
  }
  setPickerRole(role)
  await loadHistoryCharts()
}
```

Note: `selfProfile.gender` is always defined (initialized via `initialBirthProfile('male')` on line 68 — its type is `'male' | 'female'`).

- [ ] **Step 4: Add `pickerCharts` derived list**

Immediately after the `openChartPicker` function (before `handleSubmit`), insert:

```ts
const pickerCharts = useMemo(() => {
  const sameSideId = pickerRole === 'self' ? selfImportSource?.chartId : partnerImportSource?.chartId
  const otherSideId = pickerRole === 'self' ? partnerImportSource?.chartId : selfImportSource?.chartId
  const q = pickerSearch.trim().toLowerCase()

  return historyCharts
    .filter(c => pickerGenderFilter === 'all' || c.gender === pickerGenderFilter)
    .filter(c => !q || (c.display_name || '').toLowerCase().includes(q))
    .map(c => ({
      chart: c,
      isSelectedHere: c.id === sameSideId,
      isUsedByOtherSide: c.id === otherSideId,
    }))
}, [historyCharts, pickerGenderFilter, pickerSearch, pickerRole, selfImportSource, partnerImportSource])
```

- [ ] **Step 5: Retarget the panel autoFocus to focus the search input**

Locate the existing picker focus effect (currently CompatibilityPage.tsx:147-159). The `chartPickerRef` currently focuses the panel. We'll keep `chartPickerRef` on the panel but add a new ref for the search input and focus it instead:

Add this ref next to `chartPickerRef` (right after line 67 where `chartPickerRef` is declared):
```ts
const pickerSearchRef = useRef<HTMLInputElement | null>(null)
```

Then replace the picker focus effect (lines 147-159) with:
```tsx
useEffect(() => {
  if (!pickerRole) return
  const previous = document.activeElement instanceof HTMLElement ? document.activeElement : null
  pickerSearchRef.current?.focus()
  const handleKeyDown = (event: KeyboardEvent) => {
    if (event.key === 'Escape') setPickerRole(null)
  }
  document.addEventListener('keydown', handleKeyDown)
  return () => {
    document.removeEventListener('keydown', handleKeyDown)
    previous?.focus()
  }
}, [pickerRole])
```

- [ ] **Step 6: Replace the picker JSX**

Locate the picker JSX block (currently CompatibilityPage.tsx:347-380, the entire `{pickerRole && ( ... )}` expression). Replace it wholesale with:

```tsx
{pickerRole && (
  <div className="compatibility-chart-picker" role="dialog" aria-modal="true" aria-label="选择命盘档案">
    <div className="compatibility-chart-picker-panel" ref={chartPickerRef} tabIndex={-1}>
      <div className="compatibility-chart-picker-head">
        <strong>选择命盘档案</strong>
        <button type="button" className="btn btn-ghost" onClick={() => setPickerRole(null)}>
          关闭
        </button>
      </div>
      <div className="compatibility-chart-picker-toolbar">
        <input
          ref={pickerSearchRef}
          type="search"
          className="compatibility-chart-picker-search"
          placeholder="搜索命盘称呼"
          value={pickerSearch}
          onChange={event => setPickerSearch(event.target.value)}
          aria-label="搜索命盘称呼"
        />
        <div className="compatibility-chart-picker-filter" role="tablist" aria-label="性别筛选">
          {(['all', 'male', 'female'] as const).map(g => (
            <button
              key={g}
              type="button"
              role="tab"
              aria-selected={pickerGenderFilter === g}
              onClick={() => setPickerGenderFilter(g)}
            >
              {g === 'all' ? '全部' : g === 'male' ? '男命' : '女命'}
            </button>
          ))}
        </div>
      </div>
      {pickerCharts.length === 0 ? (
        historyCharts.length === 0 ? (
          <div className="compatibility-chart-picker-empty">
            <p>还没有命盘档案</p>
            <Link to="/" className="btn btn-primary compatibility-chart-picker-empty-cta">
              立即新建命盘
            </Link>
          </div>
        ) : (
          <div className="compatibility-chart-picker-empty">
            <p>没有匹配的命盘</p>
            <button
              type="button"
              className="btn btn-ghost compatibility-chart-picker-empty-cta"
              onClick={() => { setPickerSearch(''); setPickerGenderFilter('all') }}
            >
              清除筛选
            </button>
          </div>
        )
      ) : (
        <div className="compatibility-chart-picker-list" role="listbox">
          {pickerCharts.map(({ chart, isSelectedHere, isUsedByOtherSide }) => (
            <button
              key={chart.id}
              type="button"
              role="option"
              aria-selected={isSelectedHere}
              className="compatibility-chart-picker-item"
              onClick={() => {
                applyImportedChart(pickerRole, chart)
                setPickerRole(null)
              }}
            >
              <div className="compatibility-chart-picker-item-head">
                <span className="compatibility-chart-picker-item-name">
                  {chartDisplayName(chart)}
                </span>
                <span className="compatibility-chart-picker-item-badges">
                  <span className="compatibility-chart-picker-badge compatibility-chart-picker-badge-gender">
                    {genderText(chart.gender)}
                  </span>
                  {isSelectedHere && (
                    <span className="compatibility-chart-picker-badge compatibility-chart-picker-badge-selected">
                      已选中
                    </span>
                  )}
                  {!isSelectedHere && isUsedByOtherSide && (
                    <span className="compatibility-chart-picker-badge compatibility-chart-picker-badge-conflict">
                      已作为{pickerRole === 'self' ? '对方' : '我'}
                    </span>
                  )}
                </span>
              </div>
              <div className="compatibility-chart-picker-item-birth">{formatBirth(chart)}</div>
              <div className="compatibility-chart-picker-item-pillars serif">{formatPillars(chart)}</div>
            </button>
          ))}
        </div>
      )}
    </div>
  </div>
)}
```

- [ ] **Step 7: Verify the TSX builds and lints**

Run from `/Users/liujiming/web/yuanju/frontend`:
```
npm run build
```
Expected: exit code 0.

Then:
```
npm run lint
```
Expected: no errors specific to CompatibilityPage.tsx. Pre-existing warnings in other files are not blockers — confirm only by diffing against `git diff HEAD~1 -- frontend/src/pages/CompatibilityPage.tsx`.

- [ ] **Step 8: Commit**

```bash
cd /Users/liujiming/web/yuanju
git add frontend/src/pages/CompatibilityPage.tsx
git commit -m "$(cat <<'EOF'
feat(compat-picker): search + filter + selected/conflict badges

Picker upgrade in CompatibilityPage:
- Add pickerSearch / pickerGenderFilter local state
- Role-aware filter default (partner -> opposite gender of self)
- useMemo derived list with isSelectedHere / isUsedByOtherSide flags
- New JSX: toolbar (search + gender chips), listbox with badges and
  full pillars line, dedicated empty-archive and empty-filter states
- Search input autoFocus on open (replacing panel focus)
- Use shared chartLabel helpers for name/birth/pillars formatting

CSS for new classes lands in the next commit; layout may look
unstyled in this intermediate state.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 4: CSS for the upgraded picker

**Files:**
- Modify: `frontend/src/pages/CompatibilityPage.css` (replace 2 legacy item-child rules, append new picker styles + 640px media query)

- [ ] **Step 1: Tighten legacy `.item span` / `.item small` selectors**

The existing rules at lines 254-262 currently apply to *any* span/small inside an item, which would catch the new `.compatibility-chart-picker-badge` spans. Replace those two rules:

Find this block:
```css
.compatibility-chart-picker-item span {
  font-weight: 600;
  line-height: 1.4;
}

.compatibility-chart-picker-item small {
  color: var(--text-muted);
  font-size: 12px;
}
```

Replace with:
```css
.compatibility-chart-picker-item-name {
  font-weight: 600;
  line-height: 1.4;
  color: var(--text-primary);
}

.compatibility-chart-picker-item-birth {
  color: var(--text-muted);
  font-size: 12px;
}
```

- [ ] **Step 2: Append toolbar, badges, pillars, and empty-state styles**

Find the existing `.compatibility-chart-picker-empty` rule (currently at lines 264-268):
```css
.compatibility-chart-picker-empty {
  padding: 28px 16px;
  color: var(--text-muted);
  text-align: center;
}
```

Replace it with this expanded block (so the empty container can host the CTA button vertically), then append all the new rules:

```css
.compatibility-chart-picker-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 14px;
  padding: 36px 16px;
  color: var(--text-muted);
  text-align: center;
}

.compatibility-chart-picker-empty-cta {
  min-width: 160px;
}

.compatibility-chart-picker-toolbar {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border-subtle);
}

.compatibility-chart-picker-search {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.04);
  color: var(--text-primary);
  font-family: var(--font-sans);
  font-size: 14px;
}

.compatibility-chart-picker-search:focus {
  outline: none;
  border-color: rgba(201, 168, 76, 0.6);
}

.compatibility-chart-picker-filter {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.compatibility-chart-picker-filter button {
  padding: 6px 14px;
  border: 1px solid var(--border-subtle);
  border-radius: 999px;
  background: transparent;
  color: var(--text-muted);
  font-family: var(--font-sans);
  font-size: 13px;
  cursor: pointer;
}

.compatibility-chart-picker-filter button[aria-selected="true"] {
  border-color: rgba(201, 168, 76, 0.6);
  background: rgba(201, 168, 76, 0.12);
  color: var(--text-primary);
}

.compatibility-chart-picker-item-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.compatibility-chart-picker-item-badges {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.compatibility-chart-picker-badge {
  padding: 2px 8px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 500;
  line-height: 1.5;
}

.compatibility-chart-picker-badge-gender {
  border: 1px solid var(--border-subtle);
  background: rgba(255, 255, 255, 0.04);
  color: var(--text-muted);
}

.compatibility-chart-picker-badge-selected {
  border: 1px solid rgba(201, 168, 76, 0.6);
  background: rgba(201, 168, 76, 0.15);
  color: rgba(201, 168, 76, 1);
}

.compatibility-chart-picker-badge-conflict {
  border: 1px solid rgba(255, 165, 0, 0.4);
  background: rgba(255, 165, 0, 0.1);
  color: rgba(255, 165, 0, 0.9);
}

.compatibility-chart-picker-item-pillars {
  color: var(--text-muted);
  font-size: 12px;
  letter-spacing: 0.4px;
}
```

- [ ] **Step 3: Append the ≤640px mobile full-screen media query**

At the end of `frontend/src/pages/CompatibilityPage.css` (after the closing `}` of the existing `@media (max-width: 768px)` block, which is currently the final line of the file), append a new media query:

```css
@media (max-width: 640px) {
  .compatibility-chart-picker {
    align-items: stretch;
    padding: 0;
  }

  .compatibility-chart-picker-panel {
    width: 100%;
    height: 100vh;
    max-height: 100vh;
    border-radius: 0;
    border: none;
    display: flex;
    flex-direction: column;
  }

  .compatibility-chart-picker-head,
  .compatibility-chart-picker-toolbar {
    flex: 0 0 auto;
  }

  .compatibility-chart-picker-list {
    flex: 1 1 auto;
    max-height: none;
  }
}
```

Note: this is in addition to the existing `@media (max-width: 768px)` block at lines 270-393 — do not remove or alter that block (it carries layout rules for the rest of the page).

- [ ] **Step 4: Verify the build still succeeds**

Run from `/Users/liujiming/web/yuanju/frontend`:
```
npm run build
```
Expected: exit code 0.

- [ ] **Step 5: Commit**

```bash
cd /Users/liujiming/web/yuanju
git add frontend/src/pages/CompatibilityPage.css
git commit -m "$(cat <<'EOF'
style(compat-picker): toolbar, badges, mobile full-screen

Adds CSS for the picker toolbar (search input + gender filter chips),
the three badge variants (gender / selected / conflict), the pillars
sub-line, the empty-state CTA layout, and a 640px breakpoint that
turns the picker into a flex-column full-screen panel on phones.

Also tightens the legacy .item span / .item small selectors to
specific name/birth classes so the new badge spans don't inherit
600 font-weight.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 5: Manual smoke against spec §6 validation criteria

**Files:** none modified — this task is verification only.

- [ ] **Step 1: Start the dev server**

Run from `/Users/liujiming/web/yuanju/frontend`:
```
npm run dev
```
Wait for `VITE ready` line. Open `http://localhost:5173` in a browser; log in if not already.

- [ ] **Step 2: Walk through spec §6 verification points 1-9**

For each, mark a "✅" or "❌ <note>" inline below as you go. The full criteria list is in `docs/superpowers/specs/2026-05-28-compat-chart-picker-ux-design.md` §6.

1. **Unnamed chart legibility** — pick a chart without a display_name in the picker. Title should read `男命 · 1996年2月8日` form, NOT `庚午 庚子`. [ ]
2. **Gender filter functional** — toggle `全部 / 男命 / 女命` chips; list narrows accordingly. [ ]
3. **Partner default to opposite gender** — fill self profile as male; open partner picker; gender filter should default to `女命`. Repeat with female self → defaults to `男命`. [ ]
4. **Already-selected badge** — import chart A as self; reopen self picker; chart A shows `已选中`. [ ]
5. **Conflict badge** — with A imported as self, open partner picker; chart A shows `已作为我`. Click A — it imports as partner; self side still references A. [ ]
6. **Empty archive CTA** — sign in as a new account (or temporarily revoke charts) and open the picker; should see `还没有命盘档案` + `[立即新建命盘]` button linking to `/`. (If creating a fresh account is too disruptive, mock the empty state by editing `pickerCharts` to `[]` in DevTools React tools — but only as a fallback.) [ ]
7. **Empty filter clear** — with charts present, type `xxxx` in search → `没有匹配的命盘` + `[清除筛选]`; click button → list returns to full. [ ]
8. **Mobile full-screen** — open Chrome DevTools, switch to iPhone SE (375px); reopen picker; panel covers full viewport, head and toolbar visible without scroll, list scrolls underneath. [ ]
9. **Esc closes, focus returns** — open picker (search input gets focus), press Esc, picker closes, focus returns to the "从命盘档案选择" trigger button. [ ]

- [ ] **Step 3: Stop the dev server and commit the plan completion**

After all 9 are ✅, stop the dev server (`Ctrl-C`). If any are ❌, do NOT proceed — re-open the affected task, fix, re-verify.

When all green:
```bash
cd /Users/liujiming/web/yuanju
git commit --allow-empty -m "$(cat <<'EOF'
verify(compat-picker): manual smoke against spec §6 passed

All 9 verification points from
docs/superpowers/specs/2026-05-28-compat-chart-picker-ux-design.md §6
walked through on local dev server. Picker now meets the 3-second
selection goal.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Self-Review Notes (from plan author)

1. **Spec coverage** — every spec §3 sub-section and §6 validation point maps to a task:
   - §3.2 shared util → T1
   - §3.3 picker state → T3 step 2
   - §3.4 useMemo → T3 step 4
   - §3.5 JSX → T3 step 6
   - §3.6 selected/conflict badges → T3 step 6 + T4 step 2
   - §3.7 CSS → T4 (all steps)
   - §3.8 A11y (autoFocus + listbox/option) → T3 steps 5-6
   - §6 validation → T5
2. **Placeholder scan** — no TBD/TODO/"handle edge cases"; every step has exact code or exact command.
3. **Type consistency** — `chartDisplayName / chartFallbackName / formatPillars / genderText / formatBirth` defined in T1, imported with same names in T2 and T3. `pickerSearch / pickerGenderFilter / pickerCharts` introduced in T3 with consistent identifiers across all steps.
