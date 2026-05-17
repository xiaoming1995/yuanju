# Hide 用神/忌神 from Export Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove every 用神 / 忌神 display from `ShareCard` (image export) and `PrintLayout` (PDF export), including the props that fed them and the matching glossary entries — without changing any on-page UI in `ResultPage`.

**Architecture:** Pure deletion across three frontend files. Component prop interfaces, destructuring, JSX blocks, glossary entries, and the caller's prop-passes must be removed in one atomic commit — partial deletion breaks TypeScript (caller passes a prop the callee no longer declares, or callee declares a required prop the caller no longer passes). A static-content `node --test` mjs file guards the deletions.

**Tech Stack:** React 19 + TypeScript + Vite. Tests via `node --test` with `readFileSync` + regex on source files (existing project pattern, see `frontend/tests/dayun-timeline-ux.test.mjs`).

**Spec:** `docs/superpowers/specs/2026-05-17-hide-yongshen-jishen-from-export-design.md`

---

## File Structure

| File | Responsibility | Change |
|------|----------------|--------|
| `frontend/src/components/ShareCard.tsx` | Share-image renderer; remove 喜用神/忌神 badges + corresponding props | Modify |
| `frontend/src/components/PrintLayout.tsx` | PDF renderer; remove header 喜用神/忌神 badges + 用神/忌神 glossary entries + props | Modify |
| `frontend/src/pages/ResultPage.tsx` | Mounts ShareCard & PrintLayout; remove the `yongshen` / `jishen` prop passes | Modify |
| `frontend/tests/yongshen-jishen-hidden-in-export.test.mjs` | Static-content test guarding the deletions | Create |

No backend, no DDL, no routes, no other components affected.

---

## Task 0: Create feature branch

**Files:** (none — branch only)

- [ ] **Step 1: Verify clean working tree on main at the spec commit**

```bash
git -C /Users/liujiming/web/yuanju status --short
git -C /Users/liujiming/web/yuanju rev-parse --abbrev-ref HEAD
git -C /Users/liujiming/web/yuanju log -1 --oneline
```

Expected:
- Status empty (clean)
- Branch: `main`
- HEAD: `e1120bc docs(specs): hide yongshen/jishen from share image and PDF export`

- [ ] **Step 2: Create feature branch**

```bash
git -C /Users/liujiming/web/yuanju checkout -b feat/hide-yongshen-from-export
```

Expected: `Switched to a new branch 'feat/hide-yongshen-from-export'`

---

## Task 1: Write failing tests (RED)

**Files:**
- Create: `frontend/tests/yongshen-jishen-hidden-in-export.test.mjs`

- [ ] **Step 1: Create the test file with all assertions**

Create `frontend/tests/yongshen-jishen-hidden-in-export.test.mjs` with exactly this content:

```javascript
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import test from 'node:test'
import assert from 'node:assert/strict'

const root = new URL('..', import.meta.url).pathname

function read(path) {
  return readFileSync(join(root, path), 'utf8')
}

test('ShareCard source does not contain 喜用神/忌神 badge text', () => {
  const src = read('src/components/ShareCard.tsx')
  assert.doesNotMatch(src, /喜用神：/)
  assert.doesNotMatch(src, /忌神：/)
})

test('ShareCard props interface no longer declares yongshen/jishen', () => {
  const src = read('src/components/ShareCard.tsx')
  assert.doesNotMatch(src, /yongshen:\s*string/)
  assert.doesNotMatch(src, /jishen:\s*string/)
})

test('PrintLayout source does not contain 喜用神/忌神 badge text', () => {
  const src = read('src/components/PrintLayout.tsx')
  assert.doesNotMatch(src, /喜用神：/)
  assert.doesNotMatch(src, /忌 ?神：/)
})

test('PrintLayout glossary no longer lists 用神 or 忌神 terms', () => {
  const src = read('src/components/PrintLayout.tsx')
  assert.doesNotMatch(src, /term:\s*'用神'/)
  assert.doesNotMatch(src, /term:\s*'忌神'/)
})

test('PrintLayout props interface no longer declares yongshen/jishen', () => {
  const src = read('src/components/PrintLayout.tsx')
  assert.doesNotMatch(src, /yongshen:\s*string/)
  assert.doesNotMatch(src, /jishen:\s*string/)
})

test('ResultPage no longer passes yongshen/jishen to ShareCard or PrintLayout', () => {
  const src = read('src/pages/ResultPage.tsx')
  const shareCardMatch = src.match(/<ShareCard\b[\s\S]*?\/>/)
  assert.ok(shareCardMatch, 'ShareCard mount not found in ResultPage')
  assert.doesNotMatch(shareCardMatch[0], /yongshen=/)
  assert.doesNotMatch(shareCardMatch[0], /jishen=/)

  const printLayoutMatch = src.match(/<PrintLayout\b[\s\S]*?\/>/)
  assert.ok(printLayoutMatch, 'PrintLayout mount not found in ResultPage')
  assert.doesNotMatch(printLayoutMatch[0], /yongshen=/)
  assert.doesNotMatch(printLayoutMatch[0], /jishen=/)
})
```

- [ ] **Step 2: Run the test file and confirm all six tests fail**

```bash
cd /Users/liujiming/web/yuanju/frontend && node --test tests/yongshen-jishen-hidden-in-export.test.mjs
```

Expected: `fail 6`, `pass 0`. Each test fails because:
- ShareCard.tsx still has `喜用神：` `忌神：` badge text and the `yongshen: string` / `jishen: string` fields
- PrintLayout.tsx has the same badges plus glossary entries
- ResultPage.tsx still passes `yongshen={...} jishen={...}` to both mounts

- [ ] **Step 3: Commit the failing tests**

```bash
git -C /Users/liujiming/web/yuanju add frontend/tests/yongshen-jishen-hidden-in-export.test.mjs
git -C /Users/liujiming/web/yuanju commit -m "test(export): yongshen/jishen hidden in share image and PDF (RED)"
```

---

## Task 2: Strip 用神/忌神 atomically from ShareCard, PrintLayout, and ResultPage

**Files:**
- Modify: `frontend/src/components/ShareCard.tsx` (props interface, destructuring, badge JSX block)
- Modify: `frontend/src/components/PrintLayout.tsx` (props interface, destructuring, header badge JSX block, glossary entries)
- Modify: `frontend/src/pages/ResultPage.tsx` (ShareCard mount, PrintLayout mount)

**Why atomic:** Removing the props from a component without simultaneously removing the prop-pass at the caller (or vice-versa) breaks TypeScript. All edits ship in a single commit.

- [ ] **Step 1: Delete the yongshen/jishen fields from `ShareCardProps`**

In `frontend/src/components/ShareCard.tsx`, find lines 79-80:

```tsx
  yongshen: string
  jishen: string
```

Delete those two lines. The surrounding interface block should now have `hourGanWx: string; hourZhiWx: string` (line 78) directly followed by `structured: StructuredReport | null` (was line 81, now 79).

- [ ] **Step 2: Remove `yongshen, jishen` from the destructuring inside `ShareCard`**

In the same file, find line 90:

```tsx
    yongshen, jishen, structured,
```

Replace with:

```tsx
    structured,
```

The destructuring block keeps the same overall shape — only the two names are dropped from the trailing line.

- [ ] **Step 3: Delete the entire 喜用神 / 忌神 JSX block**

In the same file, find lines 203-237 (the block that begins with the comment `{/* ┌ 喜用神 / 忌神 ── */}` and ends with the matching `)}`):

```tsx
      {/* ┌ 喜用神 / 忌神 ── */}
      {(yongshen || jishen) && (
        <>
          <div style={{
            padding: '14px 24px',
            display: 'flex',
            gap: 12,
            justifyContent: 'center',
            flexWrap: 'wrap',
            background: '#fdf8f0',
            borderBottom: '1px solid #e8dcc8',
          }}>
            {yongshen && (
              <span style={{
                fontSize: 13, padding: '5px 16px', borderRadius: 20,
                background: 'rgba(74,124,89,0.1)',
                color: '#3d6b4f', border: '1px solid rgba(74,124,89,0.35)',
                fontFamily: '"Noto Sans SC", sans-serif',
              }}>
                喜用神：{yongshen}
              </span>
            )}
            {jishen && (
              <span style={{
                fontSize: 13, padding: '5px 16px', borderRadius: 20,
                background: 'rgba(192,57,43,0.07)',
                color: '#8b2c1e', border: '1px solid rgba(192,57,43,0.25)',
                fontFamily: '"Noto Sans SC", sans-serif',
              }}>
                忌神：{jishen}
              </span>
            )}
          </div>
        </>
      )}
```

Delete all 35 lines. The line immediately before (currently `</div>` closing `{yearGan}{yearZhi} · ... · {hourGan}{hourZhi}` block, around 201) should now be directly followed by the next comment `{/* ┌ 命局格局分析（专业模式）── */}` (currently around 239).

- [ ] **Step 4: Delete the yongshen/jishen field from `PrintLayoutProps`**

In `frontend/src/components/PrintLayout.tsx`, find line 66:

```tsx
  yongshen: string; jishen: string
```

Delete this single line entirely. The line before it (`birthYear: ... gender: string` on line 65) should now be directly followed by `mingGe?: string` (currently line 67).

- [ ] **Step 5: Remove `yongshen, jishen` from `PrintLayout`'s destructured parameters**

In the same file, find line 102:

```tsx
  yongshen, jishen, mingGe, mingGeDesc, pillars, dayun, structured, shenshaMap,
```

Replace with:

```tsx
  mingGe, mingGeDesc, pillars, dayun, structured, shenshaMap,
```

- [ ] **Step 6: Delete the PDF header 喜用神/忌神 badges block**

In the same file, find lines 207-220:

```tsx
        {(yongshen || jishen) && (
          <div style={{ marginTop: 10, display: 'flex', justifyContent: 'center', gap: 16 }}>
            {yongshen && (
              <span style={{ fontSize: 11, color: '#2d5a3d', background: '#f0f7f2', padding: '3px 12px', borderRadius: 2, border: '1px solid #b5d6c3' }}>
                喜用神：{yongshen}
              </span>
            )}
            {jishen && (
              <span style={{ fontSize: 11, color: '#8b1a1a', background: '#fdf0f0', padding: '3px 12px', borderRadius: 2, border: '1px solid #f5c6c6' }}>
                忌 神：{jishen}
              </span>
            )}
          </div>
        )}
```

Delete all 14 lines. The line before (currently line 206, closing `</div>` of the date row inside the header) should be directly followed by the closing `</div>` of the header wrapper (currently line 221).

- [ ] **Step 7: Delete the 用神 and 忌神 glossary entries**

In the same file, find lines 653-654 inside the `{[ ... ]}` array literal of the 术语释义 section:

```tsx
            { term: '用神', desc: '命局中最需要扶助或调节的关键五行。' },
            { term: '忌神', desc: '容易加重失衡、需要节制或避开的五行。' },
```

Delete both lines. The array now starts directly with `{ term: '日主', ... }` (was line 655). The remaining 6 entries (日主, 十神, 调候, 格局, 大运, 流年) stay unchanged.

- [ ] **Step 8: Remove the prop-pass to `<ShareCard ...>` in ResultPage**

In `frontend/src/pages/ResultPage.tsx`, find lines 1179-1180:

```tsx
          yongshen={result.yongshen || ''}
          jishen={result.jishen || ''}
```

Delete those two lines. The block above (`hourGanWx={result.hour_gan_wuxing} hourZhiWx={result.hour_zhi_wuxing}` on 1178) should now be directly followed by `structured={report?.content_structured ?? null}` (was 1181).

- [ ] **Step 9: Remove the prop-pass to `<PrintLayout ...>` in ResultPage**

In the same file, find lines 1271-1272:

```tsx
            yongshen={result.yongshen || ''}
            jishen={result.jishen || ''}
```

Delete those two lines. The block above (`gender={result.gender}` on 1270) should now be directly followed by `mingGe={result.ming_ge || ''}` (was 1273).

- [ ] **Step 10: Run the visibility test file — expect all 6 tests pass**

```bash
cd /Users/liujiming/web/yuanju/frontend && node --test tests/yongshen-jishen-hidden-in-export.test.mjs
```

Expected: `fail 0`, `pass 6`.

- [ ] **Step 11: TypeScript build check**

```bash
cd /Users/liujiming/web/yuanju/frontend && npx tsc -b --noEmit
```

Expected: no errors. If TypeScript reports `Property 'yongshen' is missing in type ...` or `Property 'yongshen' does not exist on type ...`, that means one of Steps 1-9 was missed. Re-check every step.

- [ ] **Step 12: Commit**

```bash
git -C /Users/liujiming/web/yuanju add frontend/src/components/ShareCard.tsx frontend/src/components/PrintLayout.tsx frontend/src/pages/ResultPage.tsx
git -C /Users/liujiming/web/yuanju commit -m "feat(export): hide yongshen/jishen from share image and PDF"
```

---

## Task 3: Full verification — full test suite, lint, build

**Files:** (none — verification only)

- [ ] **Step 1: Run the entire frontend test directory to catch any regression**

```bash
cd /Users/liujiming/web/yuanju/frontend && node --test tests/*.test.mjs
```

Expected: zero failures. The newly added file should appear with 6 passes; pre-existing files should match prior pass counts.

If any pre-existing test fails, inspect — most likely a test elsewhere matched `喜用神：` / `忌神：` / `用神` / `忌神` strings as part of a wider regex. If a stale assertion relied on the deleted content, drop that assertion (do NOT re-add the badges or glossary entries).

- [ ] **Step 2: Lint**

```bash
cd /Users/liujiming/web/yuanju/frontend && npm run lint
```

Expected: no NEW errors beyond the pre-existing `PrintLayout.tsx:336,342 no-irregular-whitespace` errors (these are unrelated to this PR — they were present on `main` before the branch was cut).

If lint flags an "unused variable" warning on `yongshen` or `jishen` somewhere, that means a destructure or import wasn't fully cleaned in Task 2. Find and remove.

- [ ] **Step 3: Production build sanity check**

```bash
cd /Users/liujiming/web/yuanju/frontend && npm run build
```

Expected: build succeeds; no TypeScript errors.

- [ ] **Step 4: Final commit only if Step 1 required test edits**

If a pre-existing test in `tests/*.test.mjs` had to be adjusted in Step 1:

```bash
git -C /Users/liujiming/web/yuanju add frontend/tests/<adjusted-file>.test.mjs
git -C /Users/liujiming/web/yuanju commit -m "test(<area>): drop stale yongshen/jishen assertion"
```

Otherwise, skip.

---

## Verification Against Spec

| Spec requirement | Implemented in |
|------------------|----------------|
| §4 table row 1 — delete ShareCard badge JSX block | Task 2 Step 3 |
| §4 table row 2 — delete ShareCard `yongshen/jishen` interface fields | Task 2 Step 1 |
| §4 table row 3 — remove from ShareCard destructure | Task 2 Step 2 |
| §4 table row 4 — delete PrintLayout header badge JSX block | Task 2 Step 6 |
| §4 table row 5 — delete PrintLayout glossary entries | Task 2 Step 7 |
| §4 table row 6 — delete PrintLayout interface field | Task 2 Step 4 |
| §4 table row 7 — remove from PrintLayout destructure | Task 2 Step 5 |
| §4 table row 8 — remove prop-pass to ShareCard in ResultPage | Task 2 Step 8 |
| §4 table row 9 — remove prop-pass to PrintLayout in ResultPage | Task 2 Step 9 |
| §5.1 — share image has no 喜用神/忌神 | Test in Task 1 (ShareCard text check) + Task 2 Step 3 |
| §5.2 — PDF header has no badges; glossary keeps 6 terms | Tests in Task 1 + Task 2 Steps 6, 7 |
| §5.3 — ResultPage UI unchanged | Only ShareCard/PrintLayout call sites edited in ResultPage; on-page badges/YongshenBadge/MingpanAvatar untouched (Task 2 Steps 8-9 only delete two lines in each mount) |
| §5.4 — TypeScript compiles, ESLint clean | Task 2 Step 11 + Task 3 Steps 2-3 |
| §5.5 — new mjs test 5/5 (in fact 6/6) green | Tasks 1 + 2 Step 10 |

All spec requirements have a backing task.
