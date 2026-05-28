# 合盘结果页排版重构 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把合盘结果页从 11 段纵向堆叠重构成「sticky 摘要 + 段 ① 双方基础盘 + 段 ② 是否合 + 段 ③ AI 深度（5 子段） + ▼ 证据抽屉」的 5 大块结构，对齐已有的 ShareCard / PDF 排版。

**Architecture:** 把现 1289 行 inline 的 `CompatibilityResultPage.tsx` 中的 panel 逐个抽到 `components/compatibility/` 下独立组件；在 `.compatibility-result-page` 作用域新增 14 个 CSS 变量；用本地常量 `ENABLE_NEW_LAYOUT` 做 feature flag 保证 4 批次每批可独立 commit + 可独立验证。所有现存数据归一化（`buildDecisionDashboardData` / `buildPersonalityFitSummary` 等）保持不变。

**Tech Stack:** React 18 + TypeScript + Vite + 纯 CSS Variables（无 UI 框架）。测试用 `node:test` + `typescript` 转译，断言基于源码 grep 匹配（与项目现有测试一致）。

**Reference:** `docs/superpowers/specs/2026-05-29-compat-result-page-restructure-design.md`

**Test command pattern:** `cd frontend && node --test tests/<file>.test.mjs`

**Build command:** `cd frontend && npm run build` （tsc 严格类型 + Vite 构建，每个任务末尾运行以验证类型）

---

## 批次 1 · 设计 token + 三段骨架（最低风险）

### Task 1: 新增 14 个 CSS 变量到 `.compatibility-result-page` 作用域

**Files:**
- Test: `frontend/tests/compat-layout-tokens.test.mjs` (create)
- Modify: `frontend/src/pages/CompatibilityResultPage.css:1-7`

- [ ] **Step 1: Write the failing test**

Create `frontend/tests/compat-layout-tokens.test.mjs`:

```js
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (path) => readFileSync(resolve(root, path), 'utf8')

test('compatibility result page declares 14 layout tokens in page scope', () => {
  const css = read('src/pages/CompatibilityResultPage.css')
  const tokens = [
    '--section-gap-mobile',
    '--section-gap-desktop',
    '--section-padding-mobile',
    '--section-padding-desktop',
    '--subsection-gap',
    '--fs-section-kicker',
    '--fs-section-title',
    '--fs-section-title-desktop',
    '--fs-subsection-title',
    '--fs-body',
    '--fs-caption',
    '--sticky-h',
    '--sticky-h-desktop',
    '--container-max',
  ]
  for (const token of tokens) {
    assert.match(css, new RegExp(`\\.compatibility-result-page[^{]*\\{[^}]*${token.replace(/-/g, '\\-')}\\s*:`, 's'),
      `expected ${token} declared inside .compatibility-result-page block`)
  }
})
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd frontend && node --test tests/compat-layout-tokens.test.mjs
```

Expected: 14 assertion failures, "expected --section-gap-mobile declared inside .compatibility-result-page block".

- [ ] **Step 3: Add tokens to CSS**

Modify `frontend/src/pages/CompatibilityResultPage.css` — replace the opening block (lines 1-7) with:

```css
.compatibility-result-page {
  /* 间距节奏 */
  --section-gap-mobile: 24px;
  --section-gap-desktop: 40px;
  --section-padding-mobile: 16px;
  --section-padding-desktop: 24px;
  --subsection-gap: 16px;

  /* 字号层级 */
  --fs-section-kicker: 11px;
  --fs-section-title: 22px;
  --fs-section-title-desktop: 26px;
  --fs-subsection-title: 15px;
  --fs-body: 14px;
  --fs-caption: 12px;

  /* sticky 与栅格 */
  --sticky-h: 48px;
  --sticky-h-desktop: 56px;
  --container-max: 900px;

  padding-bottom: calc(140px + env(safe-area-inset-bottom));
}

.compatibility-result-page.page {
  padding-bottom: calc(140px + env(safe-area-inset-bottom)) !important;
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
cd frontend && node --test tests/compat-layout-tokens.test.mjs
```

Expected: PASS (1 test, 0 failures).

- [ ] **Step 5: Run full test suite to confirm no regression**

```bash
cd frontend && node --test tests/
```

Expected: all existing tests PASS.

- [ ] **Step 6: Commit**

```bash
cd frontend && git add tests/compat-layout-tokens.test.mjs src/pages/CompatibilityResultPage.css
git commit -m "feat(compat-result): add 14 layout tokens to page scope"
```

---

### Task 2: 创建 `components/compatibility/` 目录 + feature flag 常量

**Files:**
- Create: `frontend/src/components/compatibility/.gitkeep`
- Modify: `frontend/src/pages/CompatibilityResultPage.tsx:1-20` （顶部 imports 之后插入常量）

- [ ] **Step 1: Create directory placeholder**

```bash
mkdir -p frontend/src/components/compatibility/deep-analysis
touch frontend/src/components/compatibility/.gitkeep
```

- [ ] **Step 2: Add feature flag constant near top of page file**

Modify `frontend/src/pages/CompatibilityResultPage.tsx` — find the line after the final `import` statement (around line 20) and insert:

```ts
// Feature flag: 控制重构期间新结构 vs 旧 11 段结构。批次 4 完成时删除。
const ENABLE_NEW_LAYOUT = false
```

- [ ] **Step 3: Type-check**

```bash
cd frontend && npm run build
```

Expected: build succeeds.

- [ ] **Step 4: Commit**

```bash
cd frontend && git add src/components/compatibility/.gitkeep src/pages/CompatibilityResultPage.tsx
git commit -m "scaffold(compat-result): create components dir + ENABLE_NEW_LAYOUT flag"
```

---

### Task 3: 创建 3 个空容器组件（skeleton）

**Files:**
- Create: `frontend/src/components/compatibility/SectionBasicCharts.tsx`
- Create: `frontend/src/components/compatibility/SectionVerdict.tsx`
- Create: `frontend/src/components/compatibility/SectionDeepAnalysis.tsx`
- Test: `frontend/tests/compat-section-skeletons.test.mjs` (create)

- [ ] **Step 1: Write the failing test**

Create `frontend/tests/compat-section-skeletons.test.mjs`:

```js
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (path) => readFileSync(resolve(root, path), 'utf8')

test('SectionBasicCharts exports default React component', () => {
  const src = read('src/components/compatibility/SectionBasicCharts.tsx')
  assert.match(src, /export default function SectionBasicCharts/)
  assert.match(src, /compat-section-basic-charts/)
})

test('SectionVerdict exports default React component', () => {
  const src = read('src/components/compatibility/SectionVerdict.tsx')
  assert.match(src, /export default function SectionVerdict/)
  assert.match(src, /compat-section-verdict/)
})

test('SectionDeepAnalysis exports default React component', () => {
  const src = read('src/components/compatibility/SectionDeepAnalysis.tsx')
  assert.match(src, /export default function SectionDeepAnalysis/)
  assert.match(src, /compat-section-deep-analysis/)
})
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd frontend && node --test tests/compat-section-skeletons.test.mjs
```

Expected: 3 ENOENT failures.

- [ ] **Step 3: Create `SectionBasicCharts.tsx`**

```tsx
export default function SectionBasicCharts() {
  return (
    <section id="compat-section-basic-charts" className="compat-section-basic-charts">
      {/* 批次 2 实现 */}
    </section>
  )
}
```

- [ ] **Step 4: Create `SectionVerdict.tsx`**

```tsx
export default function SectionVerdict() {
  return (
    <section id="compat-section-verdict" className="compat-section-verdict">
      {/* 批次 3 实现 */}
    </section>
  )
}
```

- [ ] **Step 5: Create `SectionDeepAnalysis.tsx`**

```tsx
export default function SectionDeepAnalysis() {
  return (
    <section id="compat-section-deep-analysis" className="compat-section-deep-analysis">
      {/* 批次 4 实现 */}
    </section>
  )
}
```

- [ ] **Step 6: Run skeleton test, type-check, full suite**

```bash
cd frontend && node --test tests/compat-section-skeletons.test.mjs && npm run build && node --test tests/
```

Expected: skeleton test PASS, build PASS, all other tests PASS.

- [ ] **Step 7: Commit**

```bash
cd frontend && git add src/components/compatibility/Section*.tsx tests/compat-section-skeletons.test.mjs
git commit -m "scaffold(compat-result): empty Section{BasicCharts,Verdict,DeepAnalysis} containers"
```

---

### Task 4: 在主页面挂载新 3 段（隐藏在 flag 下，并不影响旧渲染）

**Files:**
- Modify: `frontend/src/pages/CompatibilityResultPage.tsx:1099-1227` （JSX return 区段）

- [ ] **Step 1: Import the new sections at the top of the file**

Modify `frontend/src/pages/CompatibilityResultPage.tsx` — add to the existing import group (after the lucide-react / lib imports):

```ts
import SectionBasicCharts from '../components/compatibility/SectionBasicCharts'
import SectionVerdict from '../components/compatibility/SectionVerdict'
import SectionDeepAnalysis from '../components/compatibility/SectionDeepAnalysis'
```

- [ ] **Step 2: Insert new sections (flag-gated) right after `<div className="compat-export-actions">...</div>` closes (around line 1124)**

Find `</div>` at end of `compat-export-actions` block, then insert AFTER it:

```tsx
{ENABLE_NEW_LAYOUT && (
  <>
    <SectionBasicCharts />
    <SectionVerdict />
    <SectionDeepAnalysis />
  </>
)}
```

- [ ] **Step 3: Type-check + dev sanity**

```bash
cd frontend && npm run build && node --test tests/
```

Expected: build PASS, all tests PASS, page renders identical (flag = false).

- [ ] **Step 4: Commit**

```bash
cd frontend && git add src/pages/CompatibilityResultPage.tsx
git commit -m "wire(compat-result): mount empty new-layout sections behind ENABLE_NEW_LAYOUT flag"
```

---

## 批次 2 · 段 ① 双方基础盘 + Sticky 摘要

### Task 5: 抽出 `ParticipantSummaryCard.tsx` 到独立文件

**Files:**
- Create: `frontend/src/components/compatibility/ParticipantSummaryCard.tsx`
- Modify: `frontend/src/pages/CompatibilityResultPage.tsx:839-888` （删除 inline 定义） + imports
- Test: `frontend/tests/compat-participant-card.test.mjs` (create)

- [ ] **Step 1: Write the failing test**

Create `frontend/tests/compat-participant-card.test.mjs`:

```js
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (path) => readFileSync(resolve(root, path), 'utf8')

test('ParticipantSummaryCard lives in components/compatibility', () => {
  const src = read('src/components/compatibility/ParticipantSummaryCard.tsx')
  assert.match(src, /export default function ParticipantSummaryCard/)
  assert.match(src, /compatibility-person-card/)
  assert.match(src, /compatibility-pillar-grid/)
  assert.match(src, /compatibility-wuxing-grid/)
})

test('page no longer defines ParticipantSummaryCard inline', () => {
  const page = read('src/pages/CompatibilityResultPage.tsx')
  assert.doesNotMatch(page, /^function ParticipantSummaryCard\(/m)
  assert.match(page, /import ParticipantSummaryCard from '\.\.\/components\/compatibility\/ParticipantSummaryCard'/)
})
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd frontend && node --test tests/compat-participant-card.test.mjs
```

Expected: ENOENT for the new file + assertion fail on `import ...ParticipantSummaryCard`.

- [ ] **Step 3: Create `ParticipantSummaryCard.tsx` — copy verbatim from page**

Open `frontend/src/pages/CompatibilityResultPage.tsx`. Copy these line ranges:
- 139-147: `function formatBirthText(...)`
- 149-152: `function genderText(...)`
- 154-162: `function getPillars(...)`
- 164-169: `function getWuxingItems(...)`
- 839-888: `function ParticipantSummaryCard(...)`

Create `frontend/src/components/compatibility/ParticipantSummaryCard.tsx`. Paste the 5 functions in order. Then make two changes:

1. Add imports at the top:

```ts
import type { CompatibilityParticipant, CompatibilityChartSnapshot } from '../../lib/api'
```

If `getWuxingItems` references a `wuxingLabel` constant (page line 165 calls `wuxingLabel.map(...)`), find where `wuxingLabel` is declared in `CompatibilityResultPage.tsx` (likely top-of-file or imported from another lib). If it's a local `const`, copy it to this new file. If it's imported, add the same import here.

```bash
cd frontend && grep -n "wuxingLabel\b" src/pages/CompatibilityResultPage.tsx | head -3
```

2. Change `function ParticipantSummaryCard` to `export default function ParticipantSummaryCard` (drop the `function ParticipantSummaryCard(...) { ... }` signature into the standard default-export form).

3. Sanity check after paste:

```bash
cd frontend && grep -c "compatibility-person-\|compatibility-pillar-\|compatibility-wuxing-\|compatibility-day-master" src/components/compatibility/ParticipantSummaryCard.tsx
```

Expected: ≥ 8 (matches the original count in page).

> **Don't** rewrite field names from memory. Real `formatBirthText` uses `fallback.year / month / day / hour` (not `fallback.birth_year / birth_month / ...`); real `getWuxingItems` uses a `wuxingLabel.map(...)`, not a hardcoded array. Paste exactly.

- [ ] **Step 4: Add import in page + delete the inline definition**

Modify `frontend/src/pages/CompatibilityResultPage.tsx`:
- Add to imports: `import ParticipantSummaryCard from '../components/compatibility/ParticipantSummaryCard'`
- Delete lines 839-888 (the entire inline `function ParticipantSummaryCard(...)` block)
- Also delete inline helpers `formatBirthText` (lines 139-147), `genderText` (149-152), `getPillars` (154-162), `getWuxingItems` (164-169) — they are now exclusively used by ParticipantSummaryCard

(After deletion, verify the file still uses any of these helpers elsewhere. As of writing, they are only used by ParticipantSummaryCard. If grep finds other usages, leave them in place.)

```bash
cd frontend && grep -n "formatBirthText\|genderText\|getPillars\|getWuxingItems" src/pages/CompatibilityResultPage.tsx
```

Expected: empty (all 4 helpers no longer used in page after this task).

- [ ] **Step 5: Type-check + tests**

```bash
cd frontend && npm run build && node --test tests/
```

Expected: build PASS, new participant-card test PASS, all existing tests PASS.

- [ ] **Step 6: Commit**

```bash
cd frontend && git add src/components/compatibility/ParticipantSummaryCard.tsx src/pages/CompatibilityResultPage.tsx tests/compat-participant-card.test.mjs
git commit -m "refactor(compat-result): extract ParticipantSummaryCard to its own file"
```

---

### Task 6: 创建 `CompatibilityStickyHeader.tsx` + CSS

**Files:**
- Create: `frontend/src/components/compatibility/CompatibilityStickyHeader.tsx`
- Create: `frontend/src/components/compatibility/CompatibilityStickyHeader.css`
- Test: `frontend/tests/compat-sticky-header.test.mjs` (create)

- [ ] **Step 1: Write the failing test**

Create `frontend/tests/compat-sticky-header.test.mjs`:

```js
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (path) => readFileSync(resolve(root, path), 'utf8')

test('CompatibilityStickyHeader exports component with required props', () => {
  const src = read('src/components/compatibility/CompatibilityStickyHeader.tsx')
  assert.match(src, /export default function CompatibilityStickyHeader/)
  assert.match(src, /selfName/)
  assert.match(src, /partnerName/)
  assert.match(src, /overallScore/)
  assert.match(src, /verdict/)
  assert.match(src, /compat-sticky-header/)
})

test('CompatibilityStickyHeader CSS sets position sticky and var-driven height', () => {
  const css = read('src/components/compatibility/CompatibilityStickyHeader.css')
  assert.match(css, /\.compat-sticky-header\s*\{[\s\S]*?position:\s*sticky/)
  assert.match(css, /\.compat-sticky-header\s*\{[\s\S]*?top:\s*0/)
  assert.match(css, /\.compat-sticky-header\s*\{[\s\S]*?height:\s*var\(--sticky-h\)/)
  assert.match(css, /@media \(min-width: 1024px\)[\s\S]*\.compat-sticky-header\s*\{[\s\S]*?height:\s*var\(--sticky-h-desktop\)/)
})
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd frontend && node --test tests/compat-sticky-header.test.mjs
```

Expected: ENOENT for both new files.

- [ ] **Step 3: Create `CompatibilityStickyHeader.tsx`**

```tsx
import './CompatibilityStickyHeader.css'

type Props = {
  selfName: string
  partnerName: string
  overallScore: number
  verdict: string
}

export default function CompatibilityStickyHeader({ selfName, partnerName, overallScore, verdict }: Props) {
  return (
    <header className="compat-sticky-header" aria-label="合盘摘要">
      <div className="compat-sticky-header__left">
        <span className="compat-sticky-header__names">{selfName} × {partnerName}</span>
        <a
          href="#compat-section-verdict"
          className="compat-sticky-header__verdict"
          aria-label="跳到判断详情"
        >
          {verdict}
        </a>
      </div>
      <div className="compat-sticky-header__right">
        <span className="compat-sticky-header__score serif">{overallScore}</span>
        <span className="compat-sticky-header__score-max">/100</span>
      </div>
    </header>
  )
}
```

- [ ] **Step 4: Create `CompatibilityStickyHeader.css`**

```css
.compat-sticky-header {
  position: sticky;
  top: 0;
  z-index: 50;
  height: var(--sticky-h);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 0 16px;
  background: var(--bg-surface);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  border-bottom: 1px solid var(--border-subtle);
  margin-bottom: var(--section-gap-mobile);
}

.compat-sticky-header__left {
  display: flex;
  align-items: baseline;
  gap: 10px;
  min-width: 0;
  flex: 1;
}

.compat-sticky-header__names {
  color: var(--wu-jin);
  font-weight: 600;
  font-size: 14px;
  white-space: nowrap;
}

.compat-sticky-header__verdict {
  color: var(--text-secondary);
  font-size: 13px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  text-decoration: none;
  pointer-events: none;       /* mobile: 不让用户点击跳转 */
  cursor: default;
}

.compat-sticky-header__right {
  display: flex;
  align-items: baseline;
  gap: 2px;
  color: var(--wu-jin);
}

.compat-sticky-header__score {
  font-size: 20px;
  line-height: 1;
}

.compat-sticky-header__score-max {
  font-size: 12px;
  color: var(--text-muted);
}

@media (min-width: 1024px) {
  .compat-sticky-header {
    height: var(--sticky-h-desktop);
    padding: 0 24px;
    margin-bottom: var(--section-gap-desktop);
  }
  .compat-sticky-header__verdict {
    pointer-events: auto;
    cursor: pointer;
  }
  .compat-sticky-header__verdict:hover {
    color: var(--text-primary);
  }
  .compat-sticky-header__score {
    font-size: 24px;
  }
}
```

- [ ] **Step 5: Run tests**

```bash
cd frontend && node --test tests/compat-sticky-header.test.mjs && npm run build
```

Expected: test PASS, build PASS.

- [ ] **Step 6: Commit**

```bash
cd frontend && git add src/components/compatibility/CompatibilityStickyHeader.{tsx,css} tests/compat-sticky-header.test.mjs
git commit -m "feat(compat-result): CompatibilityStickyHeader component (mobile/desktop heights)"
```

---

### Task 7: 实现 `SectionBasicCharts.tsx` 渲染双方 ParticipantSummaryCard

**Files:**
- Modify: `frontend/src/components/compatibility/SectionBasicCharts.tsx`
- Create: `frontend/src/components/compatibility/SectionBasicCharts.css`
- Test: extend `frontend/tests/compat-section-skeletons.test.mjs`

- [ ] **Step 1: Extend the test (add new cases)**

Append to `frontend/tests/compat-section-skeletons.test.mjs`:

```js
test('SectionBasicCharts renders two ParticipantSummaryCard side-by-side', () => {
  const src = read('src/components/compatibility/SectionBasicCharts.tsx')
  assert.match(src, /import ParticipantSummaryCard from '\.\/ParticipantSummaryCard'/)
  assert.match(src, /participants/)
  assert.match(src, /self/)
  assert.match(src, /partner/)
})

test('SectionBasicCharts CSS uses 2-col grid on desktop', () => {
  const css = read('src/components/compatibility/SectionBasicCharts.css')
  assert.match(css, /@media \(min-width: 641px\)[\s\S]*compat-section-basic-charts__grid[\s\S]*grid-template-columns:\s*repeat\(2/)
})
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd frontend && node --test tests/compat-section-skeletons.test.mjs
```

Expected: 2 new failures.

- [ ] **Step 3: Replace `SectionBasicCharts.tsx` content**

```tsx
import './SectionBasicCharts.css'
import ParticipantSummaryCard from './ParticipantSummaryCard'
import type { CompatibilityParticipant } from '../../lib/api'

type Props = {
  self?: CompatibilityParticipant | null
  partner?: CompatibilityParticipant | null
}

export default function SectionBasicCharts({ self, partner }: Props) {
  return (
    <section id="compat-section-basic-charts" className="compat-section-basic-charts">
      <div className="compat-section-basic-charts__head">
        <div className="compat-section-kicker">SECTION 01</div>
        <h2 className="serif compat-section-title">双方基础盘</h2>
      </div>
      <div className="compat-section-basic-charts__grid">
        {self && <ParticipantSummaryCard participant={self} />}
        {partner && <ParticipantSummaryCard participant={partner} />}
      </div>
    </section>
  )
}
```

- [ ] **Step 4: Create `SectionBasicCharts.css`**

```css
.compat-section-basic-charts {
  scroll-margin-top: var(--sticky-h);
  padding: 0 var(--section-padding-mobile);
  margin-bottom: var(--section-gap-mobile);
}

.compat-section-basic-charts__head {
  margin-bottom: 12px;
}

.compat-section-kicker {
  font-size: var(--fs-section-kicker);
  letter-spacing: 2px;
  text-transform: uppercase;
  color: var(--wu-jin);
  margin-bottom: 4px;
}

.compat-section-title {
  font-size: var(--fs-section-title);
  margin: 0;
  color: var(--text-primary);
}

.compat-section-basic-charts__grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: var(--subsection-gap);
}

@media (min-width: 641px) {
  .compat-section-basic-charts__grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (min-width: 1024px) {
  .compat-section-basic-charts {
    scroll-margin-top: var(--sticky-h-desktop);
    padding: 0 var(--section-padding-desktop);
    margin-bottom: var(--section-gap-desktop);
  }
  .compat-section-title {
    font-size: var(--fs-section-title-desktop);
  }
}
```

- [ ] **Step 5: Tests + build**

```bash
cd frontend && node --test tests/compat-section-skeletons.test.mjs && npm run build
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
cd frontend && git add src/components/compatibility/SectionBasicCharts.{tsx,css} tests/compat-section-skeletons.test.mjs
git commit -m "feat(compat-result): SectionBasicCharts renders two ParticipantSummaryCard"
```

---

### Task 8: 把 sticky + 段 ① 接到主页面（flag = true）

**Files:**
- Modify: `frontend/src/pages/CompatibilityResultPage.tsx`

- [ ] **Step 1: Add CompatibilityStickyHeader import + update flag block**

Modify `frontend/src/pages/CompatibilityResultPage.tsx`:

- Add to imports:

```ts
import CompatibilityStickyHeader from '../components/compatibility/CompatibilityStickyHeader'
```

- Change the flag declaration line:

```ts
const ENABLE_NEW_LAYOUT = true
```

- Replace the prior flag-gated empty `<SectionBasicCharts/>` block (added in Task 4) — find the existing `{ENABLE_NEW_LAYOUT && (<><SectionBasicCharts/><SectionVerdict/><SectionDeepAnalysis/></>)}` block and replace with:

```tsx
{ENABLE_NEW_LAYOUT && (
  <>
    <CompatibilityStickyHeader
      selfName={selfP?.display_name || '我'}
      partnerName={partnerP?.display_name || '对方'}
      overallScore={reading.overall_score}
      verdict={decisionDashboard.verdict}
    />
    <SectionBasicCharts self={selfP || null} partner={partnerP || null} />
    <SectionVerdict />
    <SectionDeepAnalysis />
  </>
)}
```

- [ ] **Step 2: Hide旧的 "专业命盘细节" 折叠区里的"双方命盘摘要"（避免与段 ① 重复渲染）**

Modify `frontend/src/pages/CompatibilityResultPage.tsx` — find the existing `<details className="compatibility-professional-details"` block (currently lines ~1196-1226). Inside the `compatibility-professional-body`, find the `<div className="compatibility-section">` that renders `<ParticipantSummaryCard>` (around lines 1207-1216) and wrap it with the flag:

```tsx
{!ENABLE_NEW_LAYOUT && (
  <div className="compatibility-section">
    <div className="compatibility-section-header">
      <h2 className="serif compatibility-section-title">双方命盘摘要</h2>
      <p className="compatibility-section-desc">确认双方四柱与命盘核心信息。</p>
    </div>
    <div className="compatibility-summary-grid">
      {selfP && <ParticipantSummaryCard participant={selfP} />}
      {partnerP && <ParticipantSummaryCard participant={partnerP} />}
    </div>
  </div>
)}
```

- [ ] **Step 3: Type-check + run all tests**

```bash
cd frontend && npm run build && node --test tests/
```

Expected: build PASS. Existing tests for "professional details" still pass because the `<details>` wrapper is unchanged.

- [ ] **Step 4: Manual smoke check**

Run `npm run dev`, open `/compatibility/result/<an existing reading id>` in mobile-sized browser window:

- [ ] sticky 顶栏可见，显示「姓名 × 姓名 · verdict · 分数」
- [ ] 双方基础盘出现在 sticky 下方（段 ①）
- [ ] 滚到底部"专业命盘细节"折叠区，确认**没有**再次渲染"双方命盘摘要"
- [ ] 其它旧段（决策仪表盘 / 阅读地图 / ... AI 报告）依然正常显示
- [ ] 调宽窗口到 ≥ 1024px：双盘并排，sticky 高度变高

- [ ] **Step 5: Commit**

```bash
cd frontend && git add src/pages/CompatibilityResultPage.tsx
git commit -m "wire(compat-result): turn on sticky header + Section ① basic charts"
```

---

### Task 9: 删除 `ResultReadingMap`（被 sticky 取代）

**Files:**
- Modify: `frontend/src/pages/CompatibilityResultPage.tsx:531-547` （删除 inline 函数定义）+ 调用点

- [ ] **Step 1: Delete inline function + its usage**

Modify `frontend/src/pages/CompatibilityResultPage.tsx`:

1. Locate the inline function `function ResultReadingMap()` around line 531 — delete the entire function definition (16 lines).
2. Locate the call site `<ResultReadingMap />` in the JSX (around line 1132) — delete the line.

- [ ] **Step 2: Grep to confirm removal**

```bash
cd frontend && grep -n "ResultReadingMap\|compatibility-result-map" src/pages/CompatibilityResultPage.tsx
```

Expected: empty.

- [ ] **Step 3: Remove orphan CSS (optional cleanup — only if grep shows no other usage)**

```bash
cd frontend && grep -rn "compatibility-result-map" src/
```

If empty, delete the corresponding CSS block in `frontend/src/pages/CompatibilityResultPage.css` (look for `.compatibility-result-map` selector and its descendants — typically a small block).

- [ ] **Step 4: Type-check + tests**

```bash
cd frontend && npm run build && node --test tests/
```

Expected: build PASS, no test failures (the removed function wasn't asserted on by name in any existing test — verified by `grep -n ResultReadingMap tests/`).

- [ ] **Step 5: Commit**

```bash
cd frontend && git add src/pages/CompatibilityResultPage.tsx src/pages/CompatibilityResultPage.css
git commit -m "refactor(compat-result): delete ResultReadingMap (sticky header takes over)"
```

---

### Task 10: 所有 `<details>` 改为 `<details open>`（默认展开）

**Files:**
- Modify: `frontend/src/pages/CompatibilityResultPage.tsx`

- [ ] **Step 1: List all `<details>` in scope**

```bash
cd frontend && grep -n "<details" src/pages/CompatibilityResultPage.tsx
```

Expected: 1 hit at the `compatibility-professional-details` block (~line 1196). (More may be added by later tasks — those will be created with `open` from the start.)

- [ ] **Step 2: Confirm the existing instance has `open` attribute**

Read `frontend/src/pages/CompatibilityResultPage.tsx` around line 1196. The file already has `<details className="compatibility-professional-details" id="compatibility-professional-details" open>` — verify the `open` attribute is present. If missing, add it. If present, no change needed.

- [ ] **Step 3: Update test to enforce `open` everywhere**

Append to `frontend/tests/compatibility-result-ux.test.mjs`:

```js
test('all <details> elements in compatibility result page are open by default', () => {
  const page = read('src/pages/CompatibilityResultPage.tsx')
  const opens = page.match(/<details\b[^>]*>/g) || []
  for (const tag of opens) {
    assert.match(tag, /\bopen\b/, `expected ${tag} to include the open attribute`)
  }
})
```

- [ ] **Step 4: Run test + commit**

```bash
cd frontend && node --test tests/compatibility-result-ux.test.mjs && npm run build
cd frontend && git add src/pages/CompatibilityResultPage.tsx tests/compatibility-result-ux.test.mjs
git commit -m "fix(compat-result): ensure all <details> default open"
```

---

## 批次 3 · 段 ② 是否合

### Task 11: 实现 `SectionVerdict.tsx` + `SectionVerdict.css`（verdict + 总分 + 维度条 + findings）

**Files:**
- Modify: `frontend/src/components/compatibility/SectionVerdict.tsx`
- Create: `frontend/src/components/compatibility/SectionVerdict.css`
- Test: `frontend/tests/compat-section-verdict.test.mjs` (create)

- [ ] **Step 1: Write the failing test**

Create `frontend/tests/compat-section-verdict.test.mjs`:

```js
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (path) => readFileSync(resolve(root, path), 'utf8')

test('SectionVerdict accepts dashboard / findings / scores props and renders both v3 and legacy branches', () => {
  const src = read('src/components/compatibility/SectionVerdict.tsx')
  assert.match(src, /export default function SectionVerdict/)
  assert.match(src, /dashboard/)
  assert.match(src, /findings/)
  assert.match(src, /isV3/)
  assert.match(src, /v3Scores/)
  assert.match(src, /legacyScores/)
  assert.match(src, /ScoreOverviewV3/)
  assert.match(src, /ScoreOverview/)
  assert.match(src, /compat-section-verdict/)
})

test('SectionVerdict CSS uses 2-col layout on desktop', () => {
  const css = read('src/components/compatibility/SectionVerdict.css')
  assert.match(css, /@media \(min-width: 1024px\)[\s\S]*compat-section-verdict__columns[\s\S]*grid-template-columns/)
})
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd frontend && node --test tests/compat-section-verdict.test.mjs
```

Expected: assertion failures.

- [ ] **Step 3: Implement `SectionVerdict.tsx`** — re-export the existing `ScoreOverviewV3` / `ScoreOverview` / `DecisionEvidenceSummary` from page

Plan to extract those 3 sub-components to a shared internal helper module first. To avoid widening this task's scope, we copy them inline into SectionVerdict temporarily and clean up in Task 13.

```tsx
import './SectionVerdict.css'
import type {
  CompatibilityDimensionScoresLegacy,
  CompatibilityDimensionScoresV3,
} from '../../lib/api'
import type { DecisionFinding, DecisionDashboardData } from '../../lib/compatibilityDecision'

type Props = {
  dashboard: DecisionDashboardData
  isV3: boolean
  v3Scores: CompatibilityDimensionScoresV3 | null
  legacyScores: CompatibilityDimensionScoresLegacy | null
  overallScore: number
  overallLevel: string
  findings: DecisionFinding[]
}

export default function SectionVerdict({
  dashboard, isV3, v3Scores, legacyScores, overallScore, overallLevel, findings,
}: Props) {
  return (
    <section id="compat-section-verdict" className="compat-section-verdict">
      <div className="compat-section-verdict__head">
        <div className="compat-section-kicker">SECTION 02</div>
        <h2 className="serif compat-section-title">是否合</h2>
      </div>

      <div className="compat-section-verdict__columns">
        <div className="compat-section-verdict__main">
          <div className="compat-section-verdict__verdict-line serif">{dashboard.verdict}</div>
          <div className="compat-section-verdict__type">{dashboard.relationshipType}</div>

          {isV3 && v3Scores && (
            <InlineScoreOverviewV3 scores={v3Scores} overallScore={overallScore} overallLevel={overallLevel} />
          )}
          {!isV3 && legacyScores && (
            <InlineScoreOverviewLegacy scores={legacyScores} overallScore={overallScore} />
          )}
        </div>

        <div className="compat-section-verdict__findings">
          <div className="compat-section-verdict__findings-title">为什么这么判断</div>
          {findings.length === 0 ? (
            <p className="compat-section-verdict__findings-empty">暂无结构化判断要点。</p>
          ) : (
            <ol className="compat-section-verdict__findings-list">
              {findings.slice(0, 5).map((f, i) => (
                <li key={`${f.text}-${i}`}>
                  <span className="compat-section-verdict__findings-index">{i + 1}</span>
                  <span>{f.text}</span>
                </li>
              ))}
            </ol>
          )}
        </div>
      </div>
    </section>
  )
}

// 临时占位组件：依赖 Task 12 把 ScoreOverview 抽出到独立文件。在 Task 12 完成之前，
// SectionVerdict 不会被任何调用方实际渲染（Task 13 才会接到主页面），所以下面的 throw
// 不会在生产代码路径里触发。Task 12 会用真实的 import 替换这两个占位。
function InlineScoreOverviewV3(_props: { scores: CompatibilityDimensionScoresV3; overallScore: number; overallLevel: string }) {
  throw new Error('SectionVerdict requires Task 12 (extract ScoreOverview module) to be completed first')
}
function InlineScoreOverviewLegacy(_props: { scores: CompatibilityDimensionScoresLegacy; overallScore: number }) {
  throw new Error('SectionVerdict requires Task 12 (extract ScoreOverview module) to be completed first')
}
```

> ⚠️ This file deliberately throws — Task 12 must precede Task 13 (wiring). The throw is intentional, ensures dependency ordering, and is removed by Task 12.

- [ ] **Step 4: Implement `SectionVerdict.css`**

```css
.compat-section-verdict {
  scroll-margin-top: var(--sticky-h);
  padding: 0 var(--section-padding-mobile);
  margin-bottom: var(--section-gap-mobile);
}

.compat-section-verdict__head {
  margin-bottom: 12px;
}

.compat-section-verdict__columns {
  display: grid;
  grid-template-columns: 1fr;
  gap: var(--subsection-gap);
}

.compat-section-verdict__main,
.compat-section-verdict__findings {
  background: var(--bg-card);
  border-radius: var(--radius-md);
  padding: 16px;
  border-left: 3px solid var(--wu-jin);
}

.compat-section-verdict__verdict-line {
  font-size: 20px;
  color: var(--text-primary);
  line-height: 1.5;
  margin-bottom: 6px;
}

.compat-section-verdict__type {
  font-size: var(--fs-caption);
  color: var(--text-muted);
  margin-bottom: 16px;
}

.compat-section-verdict__findings-title {
  font-size: var(--fs-subsection-title);
  color: var(--text-primary);
  margin-bottom: 10px;
}

.compat-section-verdict__findings-empty {
  color: var(--text-muted);
  font-size: var(--fs-body);
  margin: 0;
}

.compat-section-verdict__findings-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: grid;
  gap: 10px;
}

.compat-section-verdict__findings-list li {
  display: grid;
  grid-template-columns: 24px 1fr;
  gap: 10px;
  align-items: start;
  font-size: var(--fs-body);
  color: var(--text-secondary);
  line-height: 1.7;
}

.compat-section-verdict__findings-index {
  color: var(--wu-jin);
  font-weight: 600;
  text-align: center;
}

@media (min-width: 1024px) {
  .compat-section-verdict {
    scroll-margin-top: var(--sticky-h-desktop);
    padding: 0 var(--section-padding-desktop);
    margin-bottom: var(--section-gap-desktop);
  }
  .compat-section-verdict__columns {
    grid-template-columns: 1.2fr 1fr;
  }
}
```

- [ ] **Step 5: Tests + build**

```bash
cd frontend && node --test tests/compat-section-verdict.test.mjs && npm run build
```

Expected: source-pattern test PASS; build PASS (the runtime `throw` doesn't fire because nothing renders this component yet — flag wiring happens in Task 13).

- [ ] **Step 6: Commit**

```bash
cd frontend && git add src/components/compatibility/SectionVerdict.{tsx,css} tests/compat-section-verdict.test.mjs
git commit -m "feat(compat-result): SectionVerdict scaffolding (inline score placeholders, throws until Task 12)"
```

---

### Task 12: 抽出 `ScoreOverviewV3` 和 `ScoreOverview` 到独立文件

**Files:**
- Create: `frontend/src/components/compatibility/ScoreOverview.tsx`
- Modify: `frontend/src/pages/CompatibilityResultPage.tsx:341-417` （删除两个 inline 定义）
- Modify: `frontend/src/components/compatibility/SectionVerdict.tsx` （替换 Inline* 为真正的 import）
- Test: `frontend/tests/compat-score-overview.test.mjs` (create)

- [ ] **Step 1: Write the failing test**

Create `frontend/tests/compat-score-overview.test.mjs`:

```js
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (path) => readFileSync(resolve(root, path), 'utf8')

test('ScoreOverview module exports both V3 and legacy components', () => {
  const src = read('src/components/compatibility/ScoreOverview.tsx')
  assert.match(src, /export function ScoreOverviewV3/)
  assert.match(src, /export function ScoreOverview/)
  assert.match(src, /compat-score-v3/)
  assert.match(src, /compatibility-quick-score/)
})

test('page no longer defines ScoreOverviewV3 nor ScoreOverview inline', () => {
  const page = read('src/pages/CompatibilityResultPage.tsx')
  assert.doesNotMatch(page, /^function ScoreOverviewV3\(/m)
  assert.doesNotMatch(page, /^function ScoreOverview\(/m)
  assert.match(page, /import \{ ScoreOverviewV3, ScoreOverview \} from '\.\.\/components\/compatibility\/ScoreOverview'/)
})

test('SectionVerdict imports the extracted ScoreOverview pair', () => {
  const src = read('src/components/compatibility/SectionVerdict.tsx')
  assert.match(src, /import \{ ScoreOverviewV3, ScoreOverview \} from '\.\/ScoreOverview'/)
  assert.doesNotMatch(src, /InlineScoreOverview/)
  assert.doesNotMatch(src, /throw new Error/)
})
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd frontend && node --test tests/compat-score-overview.test.mjs
```

Expected: ENOENT for the new file + 2 more assertion failures.

- [ ] **Step 3: Create `frontend/src/components/compatibility/ScoreOverview.tsx`**

Copy the inline code from `frontend/src/pages/CompatibilityResultPage.tsx` **verbatim**, lines:

- 219-227: `clampScore`, `scoreTone` (helpers)
- 229-242: `getDimensionItems` (legacy helper; uses `dimensionText[key]` and `dimensionHint[key]` maps imported from `../../lib/compatibilityPersonality` or similar — verify the original file's imports and copy any required map declarations along)
- 320-339: `dimensionHintV3`, `dimensionLabelV3`, `dimensionMaxV3` (V3 maps)
- 341-390: `ScoreOverviewV3` (the V3 component body)
- 392-417: `ScoreOverview` (legacy component body)

In the new `ScoreOverview.tsx`:

1. Top imports:

```ts
import type {
  CompatibilityDimensionScoresLegacy,
  CompatibilityDimensionScoresV3,
} from '../../lib/api'
// 如果原 page 文件中 `dimensionText` / `dimensionHint` 是从 `../lib/compatibilityPersonality` 或别处导入的，本文件 import 同样的路径。否则把原 map 定义一并复制到本文件。
```

2. Both `ScoreOverviewV3` and `ScoreOverview` functions **must be exported as `export function`** (not default) — because Task 11's `SectionVerdict.tsx` already imports them as `import { ScoreOverviewV3, ScoreOverview } from './ScoreOverview'`.

3. **Do not rewrite the body.** The V3 dimension keys are `zodiac`, `nayin`, `day_pillar`, `eight_chars` (NOT `intimacy/durability/practicality/growth` — those would be a hallucinated rewrite). The legacy keys are `attraction`, `stability`, `communication`, `practicality`. The V3 level badge text is `上吉 / 中 / 低`. Match the original exactly.

4. Quick sanity check after pasting:

```bash
cd frontend && diff <(grep -A 100 "^function ScoreOverviewV3" src/pages/CompatibilityResultPage.tsx | head -50) <(grep -A 100 "^export function ScoreOverviewV3" src/components/compatibility/ScoreOverview.tsx | head -50 | sed 's/export //')
```

Expected: no diff (only the `export` keyword differs, which `sed` strips for comparison).

> 注意：该文件保留了 `dimensionLabelV3` / `dimensionHintV3` / `dimensionMaxV3` / `clampScore` / `scoreTone` / `getDimensionItems` 等辅助函数。`dimensionText` / `dimensionHint` 这两个 map（被 `getDimensionItems` 用到）若在 page 中本就是 `import` 进来的，新文件保持同样 import；若是本地定义则一并搬过来。下一步删除 inline 时会处理 page 端的死代码。

- [ ] **Step 4: Replace page's inline `ScoreOverviewV3`/`ScoreOverview` and their helpers with the import**

Modify `frontend/src/pages/CompatibilityResultPage.tsx`:

1. Add import:
```ts
import { ScoreOverviewV3, ScoreOverview } from '../components/compatibility/ScoreOverview'
```
2. Delete:
   - Lines 219-242: `clampScore`, `scoreTone`, `getDimensionItems`, `dimensionLabelV3`, `dimensionHintV3` (verify via grep that no other site uses them — `grep -n "clampScore\|scoreTone\|getDimensionItems\|dimensionLabelV3\|dimensionHintV3" src/pages/CompatibilityResultPage.tsx`; only ScoreOverview-related lines should remain after this delete).
   - Lines 341-390: `function ScoreOverviewV3(...)`.
   - Lines 392-417: `function ScoreOverview(...)`.

- [ ] **Step 5: Replace `SectionVerdict.tsx`'s Inline* placeholders with real imports**

Modify `frontend/src/components/compatibility/SectionVerdict.tsx`:

- Add `import { ScoreOverviewV3, ScoreOverview } from './ScoreOverview'`
- Delete the two `InlineScoreOverview*` placeholder function definitions (the ones that `throw new Error(...)`).
- Replace usage `<InlineScoreOverviewV3 ... />` → `<ScoreOverviewV3 ... />`
- Replace `<InlineScoreOverviewLegacy ... />` → `<ScoreOverview ... />`

- [ ] **Step 6: Run all tests + build**

```bash
cd frontend && node --test tests/ && npm run build
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
cd frontend && git add src/components/compatibility/ScoreOverview.tsx src/components/compatibility/SectionVerdict.tsx src/pages/CompatibilityResultPage.tsx tests/compat-score-overview.test.mjs
git commit -m "refactor(compat-result): extract ScoreOverview{V3} module; wire into SectionVerdict"
```

---

### Task 13: 把段 ② 接到主页面 + 把 verdict 接到 sticky

**Files:**
- Modify: `frontend/src/pages/CompatibilityResultPage.tsx`

- [ ] **Step 1: Replace the empty `<SectionVerdict />` with full prop wiring**

In the existing flag block (added in Task 4 + Task 8), replace `<SectionVerdict />` with:

```tsx
<SectionVerdict
  dashboard={decisionDashboard}
  isV3={isV3}
  v3Scores={v3Scores}
  legacyScores={legacyScores}
  overallScore={reading.overall_score}
  overallLevel={reading.overall_level}
  findings={decisionDashboard.findings}
/>
```

- [ ] **Step 2: Wrap old "决策仪表盘 / 阅读地图（已删） / 为什么这么判断 / 量化分数 / 关键依据" sections with `!ENABLE_NEW_LAYOUT`**

The duplicated rendering should be hidden when flag is on. Find these JSX blocks in the page render (around lines 1125-1180):

- `<DecisionDashboardPanel ... />` (~1125)
- `<DecisionEvidenceSummary ... />` (~1134)
- `<div id="compatibility-score-evidence">...</div>` containing `<ScoreOverviewV3>` or `<ScoreOverview>` and `EvidenceLinkedClaims` (~1161-1181)

Wrap all three with `{!ENABLE_NEW_LAYOUT && (<>...</>)}`. The remaining flag-off path keeps the old behavior.

Example (illustrative — apply to all 3 blocks):

```tsx
{!ENABLE_NEW_LAYOUT && (
  <DecisionDashboardPanel
    reading={reading}
    dashboard={decisionDashboard}
    selfName={selfP?.display_name || '我'}
    partnerName={partnerP?.display_name || '对方'}
  />
)}
```

- [ ] **Step 3: Build + run all tests**

```bash
cd frontend && npm run build && node --test tests/
```

Expected: build PASS. ⚠️ Existing `compatibility-decision-dashboard.test.mjs` may start failing because its assertion `assert.ok(dashboard > -1)` may pass (we kept the panel under flag-off) but the relative ordering changed. **Inspect the failure** — if it asserts source order, those assertions still pass because both the new and old paths are present in source. If something fails, defer to Task 22 (test update batch).

- [ ] **Step 4: Manual smoke**

`npm run dev`, open the page:
- [ ] 段 ② 在 sticky 和段 ① 之后，显示 verdict + 总分 + 维度条 + findings
- [ ] 旧的"决策仪表盘 / 为什么这么判断 / 量化分数"块**不再可见**
- [ ] 旧的"性格相处画像 / 7-30 天验证 / 关系策略 / AI 报告 / 专业命盘细节"**依然可见**（这些归批次 4 处理）
- [ ] v3 用户和 legacy 用户都可正确渲染分数

- [ ] **Step 5: Commit**

```bash
cd frontend && git add src/pages/CompatibilityResultPage.tsx
git commit -m "wire(compat-result): Section ② receives dashboard/findings/scores; hide legacy verdict path"
```

---

## 批次 4 · 段 ③ AI 深度 + 证据抽屉 + 清理

### Task 14: 抽出 `PersonalityFit.tsx`（段 ③ 子段 1）

**Files:**
- Create: `frontend/src/components/compatibility/deep-analysis/PersonalityFit.tsx`
- Create: `frontend/src/components/compatibility/deep-analysis/PersonalityFit.css`
- Modify: `frontend/src/pages/CompatibilityResultPage.tsx:563-600` （删 inline）
- Test: `frontend/tests/compat-deep-personality.test.mjs` (create)

- [ ] **Step 1: Write the failing test**

```js
// frontend/tests/compat-deep-personality.test.mjs
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'
const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (p) => readFileSync(resolve(root, p), 'utf8')

test('PersonalityFit deep-analysis subsection exists', () => {
  const src = read('src/components/compatibility/deep-analysis/PersonalityFit.tsx')
  assert.match(src, /export default function PersonalityFit/)
  assert.match(src, /<details open/)
  assert.match(src, /compat-da-personality/)
})

test('page no longer defines PersonalityFitPanel inline', () => {
  const page = read('src/pages/CompatibilityResultPage.tsx')
  assert.doesNotMatch(page, /^function PersonalityFitPanel\(/m)
})
```

- [ ] **Step 2: Run test, expect fails**

```bash
cd frontend && node --test tests/compat-deep-personality.test.mjs
```

- [ ] **Step 3: Create `PersonalityFit.tsx`** — verbatim copy of page lines 548-562 + 563-600, then wrap.

Open `frontend/src/pages/CompatibilityResultPage.tsx`. Copy:
- Lines 548-562: `function PersonalityPointList(...)` (helper)
- Lines 563-600: `function PersonalityFitPanel(...)` (main body)

Create `frontend/src/components/compatibility/deep-analysis/PersonalityFit.tsx`. Paste both, then:

1. Add top imports:

```ts
import './PersonalityFit.css'
import type { PersonalityFitSummary } from '../../../lib/compatibilityPersonality'
// Plus whatever type PersonalityPointList's `points` parameter uses (likely PersonalityPoint
// from the same module). Verify by checking page imports.
```

2. Rename `function PersonalityFitPanel(...)` → `export default function PersonalityFit(...)`.

3. Wrap the returned `<section>` with `<details open>` so it can be collapsed (per spec §3.7). Move the `<div className="compatibility-section-header compatibility-section-header--stacked">` content into `<summary>` for a clean toggle UI. Keep all real field references (`summary.headline`, `summary.matchTypeDescription`, `summary.questionLabel`, `summary.stageLabel`, `summary.summary`, `summary.selfPattern.{title,detail}`, `summary.partnerPattern.{title,detail}`, `summary.fitPoints`, `summary.clashPoints`, `summary.communicationGuidance`, `summary.reportNote`, `summary.evidenceTargets`) unchanged.

4. CSS classes can stay as-is (`.compatibility-personality-fit`, `.compatibility-personality-pattern`, etc.) — Task 22 (CSS split) decides whether to rename or just move them to `PersonalityFit.css`. For this step, keep classes unchanged so the existing styles still apply.

5. Sanity check:

```bash
cd frontend && grep -c "summary\.\(headline\|matchTypeDescription\|questionLabel\|stageLabel\|selfPattern\|partnerPattern\|fitPoints\|clashPoints\|communicationGuidance\|reportNote\|evidenceTargets\)" src/components/compatibility/deep-analysis/PersonalityFit.tsx
```

Expected: ≥ 11 (one for each summary field used in the original).

> **Don't** invent fields like `summary.interactionMode` or `summary.tension` — those don't exist on `PersonalityFitSummary`. Paste exactly from source.

- [ ] **Step 4: Create `PersonalityFit.css`** — port over the original styles (search `frontend/src/pages/CompatibilityResultPage.css` for `.compatibility-personality-` blocks and reuse them under `.compat-da-personality__*` namespace).

```css
.compat-da-personality { padding: 16px; background: var(--bg-card); border-radius: var(--radius-md); border-left: 3px solid var(--wu-jin); }
.compat-da-subsection-summary { display: flex; flex-wrap: wrap; align-items: baseline; gap: 8px; cursor: pointer; list-style: none; }
.compat-da-subsection-summary::-webkit-details-marker { display: none; }
.compat-da-subsection-title { font-size: var(--fs-subsection-title); color: var(--text-primary); }
.compat-da-subsection-hint { font-size: var(--fs-caption); color: var(--text-muted); }
.compat-da-personality__body { margin-top: 12px; display: grid; gap: 14px; }
.compat-da-personality__summary { font-size: var(--fs-body); color: var(--text-secondary); line-height: 1.7; margin: 0; }
.compat-da-personality__pattern-grid { display: grid; grid-template-columns: 1fr; gap: 12px; }
.compat-da-personality__pattern { background: var(--bg-base); padding: 12px; border-radius: var(--radius-sm); }
.compat-da-personality__pattern span { font-size: var(--fs-caption); color: var(--text-muted); }
.compat-da-personality__pattern p { font-size: var(--fs-body); color: var(--text-secondary); margin: 6px 0 0; line-height: 1.6; }
.compat-da-personality__grid { display: grid; grid-template-columns: 1fr; gap: 12px; }
.compat-da-personality__list { padding: 12px; background: var(--bg-base); border-radius: var(--radius-sm); }
.compat-da-personality__list-title { font-size: var(--fs-caption); color: var(--text-muted); margin-bottom: 8px; }
.compat-da-personality__point { border-top: 1px solid var(--border-subtle); padding-top: 8px; margin-top: 8px; }
.compat-da-personality__point:first-child { border-top: 0; padding-top: 0; margin-top: 0; }
.compat-da-personality__point strong { font-size: var(--fs-body); color: var(--text-primary); }
.compat-da-personality__point p { font-size: var(--fs-caption); color: var(--text-secondary); margin: 4px 0 0; line-height: 1.6; }
@media (min-width: 768px) {
  .compat-da-personality__pattern-grid,
  .compat-da-personality__grid { grid-template-columns: 1fr 1fr; }
}
```

- [ ] **Step 5: Delete `PersonalityFitPanel` and `PersonalityPointList` inline from page**

Delete `function PersonalityFitPanel(...)` (lines 563-600) and `function PersonalityPointList(...)` (lines 548-562) in `frontend/src/pages/CompatibilityResultPage.tsx`.

⚠️ DO NOT yet remove the JSX call `<PersonalityFitPanel ... />`. We'll do that in Task 19 when wiring `SectionDeepAnalysis`. To prevent a build break in the interim, **at the same time**, change the JSX call site to use a temporary import alias:

```ts
// At top of page imports:
import PersonalityFit from '../components/compatibility/deep-analysis/PersonalityFit'

// In JSX, replace <PersonalityFitPanel summary={personalitySummary} /> with:
<PersonalityFit summary={personalitySummary} />
```

(The new component name differs and exports default — keep flow valid.)

- [ ] **Step 6: Tests + build**

```bash
cd frontend && node --test tests/compat-deep-personality.test.mjs && npm run build
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
cd frontend && git add src/components/compatibility/deep-analysis/PersonalityFit.{tsx,css} src/pages/CompatibilityResultPage.tsx tests/compat-deep-personality.test.mjs
git commit -m "refactor(compat-result): extract PersonalityFit deep-analysis subsection"
```

---

### Task 15: 抽出 `ActionPlan7d30d.tsx`（段 ③ 子段 2，合并验证 + 阶段风险 + 时段任务）

**Files:**
- Create: `frontend/src/components/compatibility/deep-analysis/ActionPlan7d30d.tsx`
- Create: `frontend/src/components/compatibility/deep-analysis/ActionPlan7d30d.css`
- Modify: `frontend/src/pages/CompatibilityResultPage.tsx:602-685` （删 inline）
- Test: `frontend/tests/compat-deep-actionplan.test.mjs` (create)

- [ ] **Step 1: Write failing test**

```js
// frontend/tests/compat-deep-actionplan.test.mjs
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'
const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (p) => readFileSync(resolve(root, p), 'utf8')

test('ActionPlan7d30d subsection exists with details-open', () => {
  const src = read('src/components/compatibility/deep-analysis/ActionPlan7d30d.tsx')
  assert.match(src, /export default function ActionPlan7d30d/)
  assert.match(src, /<details open/)
  assert.match(src, /compat-da-actionplan/)
})

test('page no longer defines PersonalityValidationPlanPanel / StageRiskGrid / DurationTaskSummary inline', () => {
  const page = read('src/pages/CompatibilityResultPage.tsx')
  assert.doesNotMatch(page, /^function PersonalityValidationPlanPanel\(/m)
  assert.doesNotMatch(page, /^function StageRiskGrid\(/m)
  assert.doesNotMatch(page, /^function DurationTaskSummary\(/m)
})
```

- [ ] **Step 2: Run test, expect fails**

```bash
cd frontend && node --test tests/compat-deep-actionplan.test.mjs
```

- [ ] **Step 3: Create `ActionPlan7d30d.tsx`** — verbatim copy of 3 functions, then merge into one `<details open>`.

Open `frontend/src/pages/CompatibilityResultPage.tsx`. Copy:
- Lines 602-639: `function PersonalityValidationPlanPanel(...)` — takes `{ plan, children }`
- Lines 641-655: `function StageRiskGrid(...)` — takes `{ risks }`
- Lines 657-684: `function DurationTaskSummary(...)` — takes `{ assessment }`

Create `frontend/src/components/compatibility/deep-analysis/ActionPlan7d30d.tsx`:

1. Top imports:

```ts
import './ActionPlan7d30d.css'
import type { CompatibilityDurationAssessment, CompatibilityStageRisk } from '../../../lib/api'
import type { PersonalityValidationPlan } from '../../../lib/compatibilityPersonality'
```

2. Paste `StageRiskGrid` and `DurationTaskSummary` as private (non-exported) helpers in this file. Their bodies depend on `stageWindowText` and `durationLevelText` constants — `grep -n "stageWindowText\|durationLevelText" src/pages/CompatibilityResultPage.tsx` to find their declarations and either copy or re-import them.

3. Paste `PersonalityValidationPlanPanel` body. Note its real signature:
   - Takes `{ plan: PersonalityValidationPlan, children: ReactNode }`
   - Iterates `plan.shortTerm`, `plan.mediumTerm`, `plan.avoid` — each is an object `{ title: string, items: string[], anchor?: string }` (NOT a flat array of strings).
   - Renders `plan.supportNote` at the bottom.

4. Rename `PersonalityValidationPlanPanel` to `export default function ActionPlan7d30d` and change props to accept:

```ts
type Props = {
  plan: PersonalityValidationPlan | null
  risks: CompatibilityStageRisk[]
  assessment: CompatibilityDurationAssessment
}
```

5. The new component composition (replacing the original `{children}` injection point): render `<StageRiskGrid risks={risks} />` + `<DurationTaskSummary assessment={assessment} />` inside the `compatibility-validation-detail` block. If `plan` is null, fall back to just rendering the risk grid + duration summary without the plan wrapper.

6. Wrap the entire returned `<section>` with `<details open>` and add a `<summary>` line:

```tsx
<details open className="compat-da-actionplan">
  <summary className="compat-da-subsection-summary">
    <span className="compat-da-subsection-title">性格验证计划 / 7-30 天</span>
    <span className="compat-da-subsection-hint">分阶段查看主要风险点和时段强弱</span>
  </summary>
  {/* 原 PersonalityValidationPlanPanel + StageRiskGrid + DurationTaskSummary 的合并 JSX */}
</details>
```

7. Sanity check:

```bash
cd frontend && grep -c "plan\.\(shortTerm\|mediumTerm\|avoid\|supportNote\)\|assessment\.windows\.\(three_months\|one_year\|two_years_plus\)" src/components/compatibility/deep-analysis/ActionPlan7d30d.tsx
```

Expected: ≥ 7.

> **Don't** invent `plan.shortTerm.map(item => <li>{item}</li>)` — `plan.shortTerm` is an object `{title, items, anchor?}`, not an array of strings. Real code does `plan.shortTerm.items.map(...)`. Paste from source to avoid this trap.

- [ ] **Step 4: Create `ActionPlan7d30d.css`** — adapt original styles for `.compatibility-validation-plan`, `.compatibility-stage-validation-grid`, etc.

```css
.compat-da-actionplan { padding: 16px; background: var(--bg-card); border-radius: var(--radius-md); border-left: 3px solid var(--wu-jin); }
.compat-da-actionplan__body { margin-top: 12px; display: grid; gap: 14px; }
.compat-da-actionplan__plan { display: grid; grid-template-columns: 1fr; gap: 12px; }
.compat-da-actionplan__group { background: var(--bg-base); padding: 12px; border-radius: var(--radius-sm); }
.compat-da-actionplan__group-title { font-size: var(--fs-caption); color: var(--text-muted); margin-bottom: 8px; }
.compat-da-actionplan__group ul { margin: 0; padding-left: 18px; }
.compat-da-actionplan__group li { font-size: var(--fs-body); color: var(--text-secondary); line-height: 1.7; }
.compat-da-actionplan__risks { display: grid; grid-template-columns: 1fr; gap: 10px; }
.compat-da-actionplan__risk-card { background: var(--bg-base); padding: 12px; border-radius: var(--radius-sm); border-top: 1px solid var(--border-subtle); }
.compat-da-actionplan__risk-stage { font-size: var(--fs-caption); color: var(--text-muted); }
.compat-da-actionplan__risk-text { font-size: var(--fs-body); color: var(--text-primary); margin-top: 4px; }
.compat-da-actionplan__risk-advice { font-size: var(--fs-caption); color: var(--text-secondary); margin-top: 6px; }
.compat-da-actionplan__duration { background: var(--bg-base); padding: 12px; border-radius: var(--radius-sm); }
.compat-da-actionplan__duration-summary { font-size: var(--fs-body); color: var(--text-secondary); line-height: 1.7; }
.compat-da-actionplan__duration ul { margin-top: 8px; padding-left: 18px; }
.compat-da-actionplan__duration li { font-size: var(--fs-caption); color: var(--text-muted); }
@media (min-width: 768px) {
  .compat-da-actionplan__plan { grid-template-columns: 1fr 1fr; }
  .compat-da-actionplan__risks { grid-template-columns: 1fr 1fr; }
}
```

- [ ] **Step 5: Delete 3 inline functions in page + update JSX call sites**

Modify `frontend/src/pages/CompatibilityResultPage.tsx`:

1. Delete `function PersonalityValidationPlanPanel(...)` (lines 602-640).
2. Delete `function StageRiskGrid(...)` (lines 641-656).
3. Delete `function DurationTaskSummary(...)` (lines 657-685).
4. Add import: `import ActionPlan7d30d from '../components/compatibility/deep-analysis/ActionPlan7d30d'`
5. Replace the entire JSX block currently rendering the validation panel (the `{personalityValidationPlan ? (<PersonalityValidationPlanPanel ...>...</PersonalityValidationPlanPanel>) : (<section ...>...</section>)}` block, around lines 1141-1155) with:

```tsx
<ActionPlan7d30d
  plan={personalityValidationPlan}
  risks={decisionStageRisks}
  assessment={durationAssessment}
/>
```

- [ ] **Step 6: Tests + build**

```bash
cd frontend && node --test tests/compat-deep-actionplan.test.mjs && npm run build
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
cd frontend && git add src/components/compatibility/deep-analysis/ActionPlan7d30d.{tsx,css} src/pages/CompatibilityResultPage.tsx tests/compat-deep-actionplan.test.mjs
git commit -m "refactor(compat-result): extract ActionPlan7d30d (merge plan+risks+duration)"
```

---

### Task 16: 抽出 `RelationshipStrategy.tsx`（段 ③ 子段 3）

**Files:**
- Create: `frontend/src/components/compatibility/deep-analysis/RelationshipStrategy.tsx`
- Create: `frontend/src/components/compatibility/deep-analysis/RelationshipStrategy.css`
- Modify: `frontend/src/pages/CompatibilityResultPage.tsx:686-699` （删 inline）
- Test: `frontend/tests/compat-deep-strategy.test.mjs` (create)

- [ ] **Step 1: Write the failing test**

```js
// frontend/tests/compat-deep-strategy.test.mjs
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'
const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (p) => readFileSync(resolve(root, p), 'utf8')

test('RelationshipStrategy subsection exists', () => {
  const src = read('src/components/compatibility/deep-analysis/RelationshipStrategy.tsx')
  assert.match(src, /export default function RelationshipStrategy/)
  assert.match(src, /<details open/)
})

test('page no longer defines RelationshipStrategyPanel inline', () => {
  const page = read('src/pages/CompatibilityResultPage.tsx')
  assert.doesNotMatch(page, /^function RelationshipStrategyPanel\(/m)
})
```

- [ ] **Step 2: Run test, expect fails**

```bash
cd frontend && node --test tests/compat-deep-strategy.test.mjs
```

- [ ] **Step 3: Create `RelationshipStrategy.tsx`** — verbatim copy of page lines 686-698.

Open `frontend/src/pages/CompatibilityResultPage.tsx`. Copy `function RelationshipStrategyPanel(...)` (lines 686-698) verbatim. It uses 4 string fields: `strategy.communication`, `strategy.conflict`, `strategy.reality`, `strategy.boundary` (NOT `strategy.summary` or `strategy.bullets` — those don't exist on the type). It also uses an inline helper `AdviceList` (page lines 486-499).

Create `frontend/src/components/compatibility/deep-analysis/RelationshipStrategy.tsx`:

1. Top imports:

```ts
import './RelationshipStrategy.css'
import type { CompatibilityRelationshipStrategy } from '../../../lib/api'
```

2. Paste `AdviceList` (page lines 486-499) as a private helper above the default export.

3. Paste `RelationshipStrategyPanel` body, rename to `export default function RelationshipStrategy`.

4. Wrap returned JSX with `<details open className="compat-da-strategy">` + `<summary>`. Keep the inner `compatibility-strategy-grid` + four `<AdviceList title="..." items={...} />` calls verbatim.

5. Sanity check:

```bash
cd frontend && grep -c "strategy\.\(communication\|conflict\|reality\|boundary\)" src/components/compatibility/deep-analysis/RelationshipStrategy.tsx
```

Expected: 4.

> **Don't** invent `strategy.summary` or `strategy.bullets[]` — paste real fields.

- [ ] **Step 4: Create `RelationshipStrategy.css`**

```css
.compat-da-strategy { padding: 16px; background: var(--bg-card); border-radius: var(--radius-md); border-left: 3px solid var(--wu-jin); }
.compat-da-strategy__body { margin-top: 12px; display: grid; gap: 10px; }
.compat-da-strategy__summary { font-size: var(--fs-body); color: var(--text-secondary); line-height: 1.7; margin: 0; }
.compat-da-strategy__list { margin: 0; padding-left: 18px; }
.compat-da-strategy__list li { font-size: var(--fs-body); color: var(--text-secondary); line-height: 1.7; }
```

- [ ] **Step 5: Delete inline `RelationshipStrategyPanel` + replace call site**

Modify page:
1. Delete `function RelationshipStrategyPanel(...)` (lines 686-699).
2. Add import: `import RelationshipStrategy from '../components/compatibility/deep-analysis/RelationshipStrategy'`
3. Replace the JSX call site `<RelationshipStrategyPanel strategy={consulting.relationship_strategy} />` (around line 1158) with `<RelationshipStrategy strategy={consulting.relationship_strategy} />`. Keep the surrounding `{consulting.relationship_strategy && ...}` conditional intact.

- [ ] **Step 6: Tests + build + commit**

```bash
cd frontend && node --test tests/compat-deep-strategy.test.mjs && npm run build
cd frontend && git add src/components/compatibility/deep-analysis/RelationshipStrategy.{tsx,css} src/pages/CompatibilityResultPage.tsx tests/compat-deep-strategy.test.mjs
git commit -m "refactor(compat-result): extract RelationshipStrategy deep-analysis subsection"
```

---

### Task 17: 创建 `NextStepsAndAvoid.tsx`（段 ③ 子段 4，从 DecisionDashboardPanel 残留拆出）

**Files:**
- Create: `frontend/src/components/compatibility/deep-analysis/NextStepsAndAvoid.tsx`
- Create: `frontend/src/components/compatibility/deep-analysis/NextStepsAndAvoid.css`
- Test: `frontend/tests/compat-deep-nextsteps.test.mjs` (create)

- [ ] **Step 1: Write the failing test**

```js
// frontend/tests/compat-deep-nextsteps.test.mjs
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'
const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (p) => readFileSync(resolve(root, p), 'utf8')

test('NextStepsAndAvoid subsection exists, shows nextAction/avoid/coreContradiction', () => {
  const src = read('src/components/compatibility/deep-analysis/NextStepsAndAvoid.tsx')
  assert.match(src, /export default function NextStepsAndAvoid/)
  assert.match(src, /<details open/)
  assert.match(src, /nextAction/)
  assert.match(src, /avoid/)
  assert.match(src, /summary/)
})
```

- [ ] **Step 2: Run test, expect fails**

```bash
cd frontend && node --test tests/compat-deep-nextsteps.test.mjs
```

- [ ] **Step 3: Create `NextStepsAndAvoid.tsx`** — take "下一步验证 / 短期避免 / 核心矛盾" from `DecisionDashboardPanel` body.

```tsx
import './NextStepsAndAvoid.css'
import type { DecisionDashboardData } from '../../../lib/compatibilityDecision'

export default function NextStepsAndAvoid({ dashboard }: { dashboard: DecisionDashboardData }) {
  return (
    <details open className="compat-da-nextsteps">
      <summary className="compat-da-subsection-summary">
        <span className="compat-da-subsection-title">下一步 / 避免 / 核心矛盾</span>
        <span className="compat-da-subsection-hint">短期具体动作</span>
      </summary>
      <div className="compat-da-nextsteps__body">
        <div className="compat-da-nextsteps__group">
          <span>下一步验证</span>
          <p>{dashboard.nextAction}</p>
        </div>
        {dashboard.avoid.length > 0 && (
          <div className="compat-da-nextsteps__group">
            <span>短期避免</span>
            <ul>
              {dashboard.avoid.map(item => <li key={item}>{item}</li>)}
            </ul>
          </div>
        )}
        <div className="compat-da-nextsteps__group">
          <span>核心矛盾</span>
          <p>{dashboard.summary}</p>
        </div>
      </div>
    </details>
  )
}
```

- [ ] **Step 4: Create `NextStepsAndAvoid.css`**

```css
.compat-da-nextsteps { padding: 16px; background: var(--bg-card); border-radius: var(--radius-md); border-left: 3px solid var(--wu-jin); }
.compat-da-nextsteps__body { margin-top: 12px; display: grid; gap: 12px; }
.compat-da-nextsteps__group { background: var(--bg-base); padding: 12px; border-radius: var(--radius-sm); }
.compat-da-nextsteps__group span { font-size: var(--fs-caption); color: var(--text-muted); display: block; margin-bottom: 6px; }
.compat-da-nextsteps__group p { font-size: var(--fs-body); color: var(--text-secondary); margin: 0; line-height: 1.7; }
.compat-da-nextsteps__group ul { margin: 0; padding-left: 18px; }
.compat-da-nextsteps__group li { font-size: var(--fs-body); color: var(--text-secondary); line-height: 1.7; }
```

- [ ] **Step 5: Tests + build + commit**

```bash
cd frontend && node --test tests/compat-deep-nextsteps.test.mjs && npm run build
cd frontend && git add src/components/compatibility/deep-analysis/NextStepsAndAvoid.{tsx,css} tests/compat-deep-nextsteps.test.mjs
git commit -m "feat(compat-result): NextStepsAndAvoid deep-analysis subsection"
```

---

### Task 18: 抽出 `DeepReportNarrative.tsx`（段 ③ 子段 5，从 `DeepReportPanel` 改名）

**Files:**
- Create: `frontend/src/components/compatibility/deep-analysis/DeepReportNarrative.tsx`
- Create: `frontend/src/components/compatibility/deep-analysis/DeepReportNarrative.css`
- Modify: `frontend/src/pages/CompatibilityResultPage.tsx:736-820` （删 inline `DeepReportPanel`）
- Test: `frontend/tests/compat-deep-report.test.mjs` (create)

- [ ] **Step 1: Write the failing test**

```js
// frontend/tests/compat-deep-report.test.mjs
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'
const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (p) => readFileSync(resolve(root, p), 'utf8')

test('DeepReportNarrative subsection exists', () => {
  const src = read('src/components/compatibility/deep-analysis/DeepReportNarrative.tsx')
  assert.match(src, /export default function DeepReportNarrative/)
  assert.match(src, /<details open/)
  assert.match(src, /onGenerateReport/)
})

test('page no longer defines DeepReportPanel inline', () => {
  const page = read('src/pages/CompatibilityResultPage.tsx')
  assert.doesNotMatch(page, /^function DeepReportPanel\(/m)
})
```

- [ ] **Step 2: Run test, expect fails**

```bash
cd frontend && node --test tests/compat-deep-report.test.mjs
```

- [ ] **Step 3: Create `DeepReportNarrative.tsx`** — copy `DeepReportPanel` (page lines 736-820) **verbatim** + its helper `QuestionFocusPanel` (page lines 822-837). Wrap the entire JSX inside a `<details open>` and add the standard subsection summary.

Procedure:

1. Open `frontend/src/pages/CompatibilityResultPage.tsx`. Select lines 736-820 (`function DeepReportPanel ...`) and lines 822-837 (`function QuestionFocusPanel ...`). Copy both functions.

2. Create `frontend/src/components/compatibility/deep-analysis/DeepReportNarrative.tsx` with the following structure:

```tsx
import './DeepReportNarrative.css'
import type {
  CompatibilityStructuredReport,
  CompatibilityQuestionFocus,
} from '../../../lib/api'

type Props = {
  hasReport: boolean
  structuredReport?: CompatibilityStructuredReport | null
  reportDimensions: CompatibilityStructuredReport['dimensions']
  reportRisks: string[]
  rawContent?: string
  error: string
  reportLoading: boolean
  onGenerateReport: () => void
}

// 粘贴自 frontend/src/pages/CompatibilityResultPage.tsx 行 822-837，逐字不改：
function QuestionFocusPanel({ focus }: { focus?: CompatibilityQuestionFocus }) {
  /* ←把 page 行 822-837 完整粘贴进来← */
}

// 粘贴自 frontend/src/pages/CompatibilityResultPage.tsx 行 757-819 的 return 内部 JSX。
// 注意：原 DeepReportPanel 没有 <details> 包裹；新版本要外加一层 <details open>。
export default function DeepReportNarrative({
  hasReport, structuredReport, reportDimensions, reportRisks, rawContent, error, reportLoading, onGenerateReport,
}: Props) {
  const reportStateClass = hasReport ? 'compatibility-ai-card--generated' : 'compatibility-ai-card--empty'
  return (
    <details open className="compat-da-report">
      <summary className="compat-da-subsection-summary">
        <span className="compat-da-subsection-title">AI 长文叙事</span>
        <span className="compat-da-subsection-hint">完整解读</span>
      </summary>
      <div className={`compat-da-report__body ${reportStateClass}`}>
        {/* ↓ 把 page 行 759-817 的 JSX（compatibility-ai-header / 错误状态 / 加载状态 / structured 分支 / rawContent 分支 / empty 分支）原样粘贴 ↓ */}
      </div>
    </details>
  )
}
```

3. **Don't rewrite or simplify** any conditional branches (`error`, `reportLoading`, `structuredReport`, `rawContent`, empty fallback). Copy lines 759-817 verbatim into the `compat-da-report__body` div.

4. **CSS class renaming is OPTIONAL** for this step — leave `.compatibility-ai-*` and `.compatibility-report-*` class names unchanged in JSX for now. Task 22 (CSS split) decides whether to rename or move them to `DeepReportNarrative.css` as-is.

5. After paste, verify the file does **not** drop any branch:

```bash
cd frontend && grep -c "compatibility-report-" src/components/compatibility/deep-analysis/DeepReportNarrative.tsx
```

Expected: count ≥ count in original page (compare with `grep -c "compatibility-report-" src/pages/CompatibilityResultPage.tsx | head -1` before deletion).

- [ ] **Step 4: Create `DeepReportNarrative.css`**

Take the original `.compatibility-ai-*` blocks in `frontend/src/pages/CompatibilityResultPage.css` (likely lines 700-900 — search with `grep -n "compatibility-ai-" src/pages/CompatibilityResultPage.css`), copy them, and replace `compatibility-ai-` prefix with `compat-da-report__`. Add the standard subsection wrapper:

```css
.compat-da-report { padding: 16px; background: var(--bg-card); border-radius: var(--radius-md); border-left: 3px solid var(--wu-jin); }
.compat-da-report__body { margin-top: 12px; }
/* ...其余 .compat-da-report__* 直接来自原 .compatibility-ai-* 块 */
```

- [ ] **Step 5: Delete `DeepReportPanel` + `QuestionFocusPanel` + update call site**

Modify page:
1. Delete `function DeepReportPanel(...)` (lines 736-820).
2. Delete `function QuestionFocusPanel(...)` (lines 822-837) — it was only used inside `DeepReportPanel`, now lives inside `DeepReportNarrative.tsx`.
3. Add import: `import DeepReportNarrative from '../components/compatibility/deep-analysis/DeepReportNarrative'`
4. Replace the JSX `<DeepReportPanel ... />` block (currently around lines 1183-1194 inside `<div className="card compatibility-ai-card">`) with `<DeepReportNarrative ... />`. **Also remove** the wrapping `<div className="card compatibility-ai-card">` — the new subsection provides its own card.
5. Sanity check no other consumer of QuestionFocusPanel exists:

```bash
cd frontend && grep -n "QuestionFocusPanel" src/
```

Expected: only matches inside `src/components/compatibility/deep-analysis/DeepReportNarrative.tsx`.

- [ ] **Step 6: Tests + build**

```bash
cd frontend && node --test tests/compat-deep-report.test.mjs && npm run build
```

Expected: PASS. If build fails on missing structured types in `DeepReportNarrative.tsx` import, ensure the types `CompatibilityStructuredReport`, `CompatibilityDimensionReport` are exported from `frontend/src/lib/api.ts` (they likely already are — verify by grepping).

- [ ] **Step 7: Commit**

```bash
cd frontend && git add src/components/compatibility/deep-analysis/DeepReportNarrative.{tsx,css} src/pages/CompatibilityResultPage.tsx tests/compat-deep-report.test.mjs
git commit -m "refactor(compat-result): extract DeepReportNarrative deep-analysis subsection"
```

---

### Task 19: 实现 `SectionDeepAnalysis.tsx` 容器并接入主页面

**Files:**
- Modify: `frontend/src/components/compatibility/SectionDeepAnalysis.tsx`
- Create: `frontend/src/components/compatibility/SectionDeepAnalysis.css`
- Modify: `frontend/src/pages/CompatibilityResultPage.tsx`
- Test: extend `frontend/tests/compat-section-skeletons.test.mjs`

- [ ] **Step 1: Extend the test**

Append to `frontend/tests/compat-section-skeletons.test.mjs`:

```js
test('SectionDeepAnalysis composes 5 subsection components', () => {
  const src = read('src/components/compatibility/SectionDeepAnalysis.tsx')
  assert.match(src, /import PersonalityFit from '\.\/deep-analysis\/PersonalityFit'/)
  assert.match(src, /import ActionPlan7d30d from '\.\/deep-analysis\/ActionPlan7d30d'/)
  assert.match(src, /import RelationshipStrategy from '\.\/deep-analysis\/RelationshipStrategy'/)
  assert.match(src, /import NextStepsAndAvoid from '\.\/deep-analysis\/NextStepsAndAvoid'/)
  assert.match(src, /import DeepReportNarrative from '\.\/deep-analysis\/DeepReportNarrative'/)
})
```

- [ ] **Step 2: Run test, expect fails**

```bash
cd frontend && node --test tests/compat-section-skeletons.test.mjs
```

- [ ] **Step 3: Replace `SectionDeepAnalysis.tsx`**

```tsx
import './SectionDeepAnalysis.css'
import PersonalityFit from './deep-analysis/PersonalityFit'
import ActionPlan7d30d from './deep-analysis/ActionPlan7d30d'
import RelationshipStrategy from './deep-analysis/RelationshipStrategy'
import NextStepsAndAvoid from './deep-analysis/NextStepsAndAvoid'
import DeepReportNarrative from './deep-analysis/DeepReportNarrative'
import type {
  CompatibilityDurationAssessment,
  CompatibilityStageRisk,
  CompatibilityRelationshipStrategy,
  CompatibilityStructuredReport,
  CompatibilityDimensionReport,
} from '../../lib/api'
import type {
  DecisionDashboardData,
} from '../../lib/compatibilityDecision'
import type {
  PersonalityFitSummary,
  PersonalityValidationPlan,
} from '../../lib/compatibilityPersonality'

type Props = {
  personalitySummary: PersonalityFitSummary | null
  personalityValidationPlan: PersonalityValidationPlan | null
  decisionStageRisks: CompatibilityStageRisk[]
  durationAssessment: CompatibilityDurationAssessment
  relationshipStrategy?: CompatibilityRelationshipStrategy
  dashboard: DecisionDashboardData
  deepReport: {
    hasReport: boolean
    structuredReport: CompatibilityStructuredReport | null | undefined
    reportDimensions: CompatibilityDimensionReport[]
    reportRisks: string[]
    rawContent?: string
    error?: string | null
    reportLoading: boolean
    onGenerateReport: () => void
  }
}

export default function SectionDeepAnalysis({
  personalitySummary, personalityValidationPlan, decisionStageRisks, durationAssessment,
  relationshipStrategy, dashboard, deepReport,
}: Props) {
  return (
    <section id="compat-section-deep-analysis" className="compat-section-deep-analysis">
      <div className="compat-section-deep-analysis__head">
        <div className="compat-section-kicker">SECTION 03</div>
        <h2 className="serif compat-section-title">AI 深度分析</h2>
      </div>
      <div className="compat-section-deep-analysis__stack">
        {personalitySummary && <PersonalityFit summary={personalitySummary} />}
        <ActionPlan7d30d plan={personalityValidationPlan} risks={decisionStageRisks} assessment={durationAssessment} />
        {relationshipStrategy && <RelationshipStrategy strategy={relationshipStrategy} />}
        <NextStepsAndAvoid dashboard={dashboard} />
        <DeepReportNarrative {...deepReport} />
      </div>
    </section>
  )
}
```

- [ ] **Step 4: Create `SectionDeepAnalysis.css`**

```css
.compat-section-deep-analysis {
  scroll-margin-top: var(--sticky-h);
  padding: 0 var(--section-padding-mobile);
  margin-bottom: var(--section-gap-mobile);
}
.compat-section-deep-analysis__head { margin-bottom: 12px; }
.compat-section-deep-analysis__stack { display: grid; gap: var(--subsection-gap); }
@media (min-width: 1024px) {
  .compat-section-deep-analysis {
    scroll-margin-top: var(--sticky-h-desktop);
    padding: 0 var(--section-padding-desktop);
    margin-bottom: var(--section-gap-desktop);
    max-width: 720px;
    margin-left: auto;
    margin-right: auto;
  }
}
```

- [ ] **Step 5: Wire SectionDeepAnalysis into main page (still under flag)**

In `frontend/src/pages/CompatibilityResultPage.tsx`, replace the existing `<SectionDeepAnalysis />` placeholder (added in Task 4) with:

```tsx
<SectionDeepAnalysis
  personalitySummary={personalitySummary}
  personalityValidationPlan={personalityValidationPlan}
  decisionStageRisks={decisionStageRisks}
  durationAssessment={durationAssessment}
  relationshipStrategy={consulting.relationship_strategy}
  dashboard={decisionDashboard}
  deepReport={{
    hasReport: Boolean(detail.latest_report),
    structuredReport,
    reportDimensions,
    reportRisks,
    rawContent: detail.latest_report?.content,
    error,
    reportLoading,
    onGenerateReport: handleGenerateReport,
  }}
/>
```

- [ ] **Step 6: Hide old AI subsection blocks under flag**

Find the JSX blocks that already call `<PersonalityFit />` / `<ActionPlan7d30d />` / `<RelationshipStrategy />` / `<DeepReportNarrative />` (they were inserted by Tasks 14-18 to replace the old inline panels — at the same call-site of the page-level JSX, around lines 1139-1194). Wrap each of those call sites with `{!ENABLE_NEW_LAYOUT && ...}` so they don't double-render.

Example:

```tsx
{!ENABLE_NEW_LAYOUT && personalitySummary && <PersonalityFit summary={personalitySummary} />}
```

Repeat for `ActionPlan7d30d`, `RelationshipStrategy`, `DeepReportNarrative`.

- [ ] **Step 7: Build + smoke**

```bash
cd frontend && npm run build && node --test tests/
```

Manual:
- [ ] Page renders sticky + Section ① + Section ② + Section ③（5 子段全展开）
- [ ] 各子段标题点击可折叠/展开
- [ ] v3 + legacy 用户都能看到 Section ②
- [ ] 删除掉的旧 inline panels 不再出现

- [ ] **Step 8: Commit**

```bash
cd frontend && git add src/components/compatibility/SectionDeepAnalysis.{tsx,css} src/pages/CompatibilityResultPage.tsx tests/compat-section-skeletons.test.mjs
git commit -m "wire(compat-result): SectionDeepAnalysis composes 5 subsections; hide legacy AI path"
```

---

### Task 20: 抽出 `EvidenceDrawer.tsx`（底部抽屉）

**Files:**
- Create: `frontend/src/components/compatibility/EvidenceDrawer.tsx`
- Create: `frontend/src/components/compatibility/EvidenceDrawer.css`
- Modify: `frontend/src/pages/CompatibilityResultPage.tsx` （删 inline `EvidenceLinkedClaims`, `ProfessionalEvidenceGroups`, `EvidenceCard`, `groupEvidenceBySource`）
- Test: `frontend/tests/compat-evidence-drawer.test.mjs` (create)

- [ ] **Step 1: Write the failing test**

```js
// frontend/tests/compat-evidence-drawer.test.mjs
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'
const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (p) => readFileSync(resolve(root, p), 'utf8')

test('EvidenceDrawer is a single details-open block', () => {
  const src = read('src/components/compatibility/EvidenceDrawer.tsx')
  assert.match(src, /export default function EvidenceDrawer/)
  assert.match(src, /<details open/)
  assert.match(src, /compat-evidence-drawer/)
  assert.match(src, /关键判断依据/)
  assert.match(src, /命盘细节/)
})

test('page no longer defines EvidenceLinkedClaims / ProfessionalEvidenceGroups / EvidenceCard inline', () => {
  const page = read('src/pages/CompatibilityResultPage.tsx')
  assert.doesNotMatch(page, /^function EvidenceLinkedClaims\(/m)
  assert.doesNotMatch(page, /^function ProfessionalEvidenceGroups\(/m)
  assert.doesNotMatch(page, /^function EvidenceCard\(/m)
})
```

- [ ] **Step 2: Run test, expect fails**

```bash
cd frontend && node --test tests/compat-evidence-drawer.test.mjs
```

- [ ] **Step 3: Create `EvidenceDrawer.tsx`** — verbatim copy of 4 inline functions, wrapped in `<details open>`.

Procedure:

1. Open `frontend/src/pages/CompatibilityResultPage.tsx`. From that file:
   - Copy lines 244-282 (`function EvidenceCard ...`)
   - Copy lines 284-295 (`function groupEvidenceBySource ...`)
   - Copy lines 297-318 (`function ProfessionalEvidenceGroups ...`)
   - Copy lines 700-734 (`function EvidenceLinkedClaims ...`)

2. Create `frontend/src/components/compatibility/EvidenceDrawer.tsx` with this skeleton, then **paste the 4 functions verbatim** into the marked slots:

```tsx
import './EvidenceDrawer.css'
import type {
  CompatibilityEvidence,
  CompatibilityClaimEvidenceLink,
} from '../../lib/api'

// 这两个 record map 在原 page 中是顶层 const（如 polarityColor / dimensionText / perspectiveText /
// polarityText / evidenceSourceText）。它们要么是 import，要么是 const declaration——
// 在 page 中 grep 它们的定义位置，照搬到本文件，或保持原 import 路径。
// 示例（具体以原 page 为准）：
// import { polarityColor, dimensionText, perspectiveText, polarityText, evidenceSourceText } from '../../lib/compatibilityCopy'
//
// 验证命令：
// cd frontend && grep -n "polarityColor\|dimensionText\|perspectiveText\|polarityText\|evidenceSourceText" src/pages/CompatibilityResultPage.tsx | head -10

/* === 粘贴自 page 行 244-282 === */
function EvidenceCard({ evidence }: { evidence: CompatibilityEvidence }) {
  /* 原样粘贴 page 行 245-281 的函数体 */
}

/* === 粘贴自 page 行 284-295 === */
function groupEvidenceBySource(evidences: CompatibilityEvidence[]) {
  /* 原样粘贴 page 行 285-294 的函数体 */
}

/* === 粘贴自 page 行 297-318 === */
function ProfessionalEvidenceGroups({ evidences }: { evidences: CompatibilityEvidence[] }) {
  /* 原样粘贴 page 行 298-317 的函数体 */
}

/* === 粘贴自 page 行 700-734 === */
function EvidenceLinkedClaims({
  links,
  evidences,
}: {
  links: CompatibilityClaimEvidenceLink[]
  evidences: CompatibilityEvidence[]
}) {
  /* 原样粘贴 page 行 707-733 的函数体 */
}

type Props = {
  evidences: CompatibilityEvidence[]
  claimEvidenceLinks: CompatibilityClaimEvidenceLink[]
}

export default function EvidenceDrawer({ evidences, claimEvidenceLinks }: Props) {
  return (
    <details open className="compat-evidence-drawer">
      <summary className="compat-evidence-drawer__summary">
        <span className="compat-evidence-drawer__title serif">命理证据 / 命盘细节</span>
        <span className="compat-evidence-drawer__hint">关键判断依据 + 结构化证据组</span>
      </summary>
      <div className="compat-evidence-drawer__body">
        {claimEvidenceLinks.length > 0 && (
          <div className="compat-evidence-drawer__group">
            <h3 className="compat-evidence-drawer__group-title">关键判断依据</h3>
            <EvidenceLinkedClaims links={claimEvidenceLinks} evidences={evidences} />
          </div>
        )}
        <div className="compat-evidence-drawer__group">
          <h3 className="compat-evidence-drawer__group-title">结构化证据组</h3>
          <ProfessionalEvidenceGroups evidences={evidences} />
        </div>
      </div>
    </details>
  )
}
```

3. Post-paste sanity check:

```bash
cd frontend && grep -c "compatibility-evidence-" src/components/compatibility/EvidenceDrawer.tsx
```

Expected: ≥ 8 (the page used 8+ different `.compatibility-evidence-*` className strings; all 8 must be present in the copied function bodies).

4. Build:

```bash
cd frontend && npm run build
```

Expected: PASS. If any imports are missing (e.g., `polarityColor`), the build will tell you. Add the missing imports.

- [ ] **Step 4: Create `EvidenceDrawer.css`**

```css
.compat-evidence-drawer {
  margin-top: var(--section-gap-mobile);
  padding: 16px;
  background: var(--bg-card);
  border-radius: var(--radius-md);
  border-left: 3px solid var(--wu-jin);
}
.compat-evidence-drawer__summary {
  display: flex; flex-wrap: wrap; align-items: baseline; gap: 8px;
  cursor: pointer; list-style: none;
}
.compat-evidence-drawer__summary::-webkit-details-marker { display: none; }
.compat-evidence-drawer__title { font-size: 18px; color: var(--text-primary); }
.compat-evidence-drawer__hint { font-size: var(--fs-caption); color: var(--text-muted); }
.compat-evidence-drawer__body { margin-top: 12px; display: grid; gap: var(--subsection-gap); }
.compat-evidence-drawer__group-title { font-size: var(--fs-subsection-title); color: var(--text-primary); margin: 0 0 8px; }
@media (min-width: 1024px) {
  .compat-evidence-drawer { margin-top: var(--section-gap-desktop); padding: 24px; }
}
```

- [ ] **Step 5: Delete page's inline `EvidenceCard` / `groupEvidenceBySource` / `ProfessionalEvidenceGroups` / `EvidenceLinkedClaims` + their CSS dependencies**

In `frontend/src/pages/CompatibilityResultPage.tsx`:
1. Delete `EvidenceCard` function (lines 244-282).
2. Delete `groupEvidenceBySource` function (lines 284-295).
3. Delete `ProfessionalEvidenceGroups` function (lines 297-339).
4. Delete `EvidenceLinkedClaims` function (lines 700-735).
5. Add import: `import EvidenceDrawer from '../components/compatibility/EvidenceDrawer'`

(CSS for `.compatibility-evidence-*` stays in `CompatibilityResultPage.css` for now — those classes are still referenced by the moved code in `EvidenceDrawer.tsx`. Cleanup happens in Task 23.)

- [ ] **Step 6: Wire EvidenceDrawer into page (under flag)**

In the flag block after `<SectionDeepAnalysis ... />`, add:

```tsx
<EvidenceDrawer
  evidences={detail.evidences}
  claimEvidenceLinks={consulting.claim_evidence_links}
/>
```

Also wrap the existing `<details className="compatibility-professional-details">` block (around lines 1196-1226) with `!ENABLE_NEW_LAYOUT`:

```tsx
{!ENABLE_NEW_LAYOUT && (
  <details className="compatibility-professional-details" id="compatibility-professional-details" open>
    {/* ... existing content ... */}
  </details>
)}
```

- [ ] **Step 7: Tests + build + smoke**

```bash
cd frontend && node --test tests/compat-evidence-drawer.test.mjs && npm run build && node --test tests/
```

Manual:
- [ ] 抽屉显示在页面底部，标题"命理证据 / 命盘细节"
- [ ] 默认展开
- [ ] 里面包含"关键判断依据"和"结构化证据组"

- [ ] **Step 8: Commit**

```bash
cd frontend && git add src/components/compatibility/EvidenceDrawer.{tsx,css} src/pages/CompatibilityResultPage.tsx tests/compat-evidence-drawer.test.mjs
git commit -m "refactor(compat-result): extract EvidenceDrawer (merges 4 inline evidence helpers)"
```

---

### Task 21: 移除 feature flag + 删除旧 11 段渲染路径

**Files:**
- Modify: `frontend/src/pages/CompatibilityResultPage.tsx`
- Test: update `frontend/tests/compatibility-decision-dashboard.test.mjs` (the source-order test will need a rewrite)
- Test: update `frontend/tests/compatibility-result-ux.test.mjs`

- [ ] **Step 1: Search for all `ENABLE_NEW_LAYOUT` usage**

```bash
cd frontend && grep -n "ENABLE_NEW_LAYOUT" src/pages/CompatibilityResultPage.tsx
```

Expected: 1 constant declaration + several `{ENABLE_NEW_LAYOUT && (...)}` and `{!ENABLE_NEW_LAYOUT && (...)}` blocks.

- [ ] **Step 2: For each `{ENABLE_NEW_LAYOUT && (X)}`, unwrap to `X`**

For each occurrence, remove the conditional wrapper. The "true path" stays; the "false path" gets deleted entirely.

- [ ] **Step 3: For each `{!ENABLE_NEW_LAYOUT && (X)}`, delete the entire block (and the inline components it referenced that aren't used elsewhere)**

This means deleting from the page:
- `<DecisionDashboardPanel ... />` JSX call
- `<DecisionEvidenceSummary ... />` JSX call
- `<div id="compatibility-score-evidence">...</div>` block (old score + claim evidence path)
- The inline call sites of `<PersonalityFit ... />` / `<ActionPlan7d30d ... />` / `<RelationshipStrategy ... />` / `<DeepReportNarrative ... />` that were inside `!ENABLE_NEW_LAYOUT` blocks
- The old `<details className="compatibility-professional-details">` block (the new EvidenceDrawer + Section ① replaces it)

- [ ] **Step 4: Delete remaining inline components no longer referenced**

```bash
cd frontend && grep -n "DecisionDashboardPanel\|DecisionEvidenceSummary\|AdviceList" src/pages/CompatibilityResultPage.tsx
```

If `DecisionDashboardPanel`'s only references are now the inline `function` definition itself, delete:
- `function DecisionDashboardPanel(...)` (page lines 419-484)
- `function DecisionEvidenceSummary(...)` (page lines 501-530)
- `function AdviceList(...)` (page lines 486-499) — only used by old code path
- `function QuestionFocusPanel(...)` (page lines 822-837) — verify with grep first

For each, grep before deleting to ensure no other call sites remain.

- [ ] **Step 5: Delete the feature flag constant**

```ts
// const ENABLE_NEW_LAYOUT = true  ← delete this line
```

- [ ] **Step 6: Rewrite the broken source-order test**

Replace assertions in `frontend/tests/compatibility-decision-dashboard.test.mjs`:

The test currently asserts source order of `<DecisionDashboardPanel>` ... `<DeepReportPanel>`. Replace those assertions with the new layout's source order:

```js
test('compatibility result page renders new layout sections in correct order', () => {
  const source = read(pagePath)

  const sticky = source.indexOf('<CompatibilityStickyHeader')
  const basic = source.indexOf('<SectionBasicCharts')
  const verdict = source.indexOf('<SectionVerdict')
  const deep = source.indexOf('<SectionDeepAnalysis')
  const drawer = source.indexOf('<EvidenceDrawer')

  assert.ok(sticky > -1, 'sticky header rendered')
  assert.ok(basic > sticky, 'basic charts follows sticky')
  assert.ok(verdict > basic, 'verdict follows basic charts')
  assert.ok(deep > verdict, 'deep analysis follows verdict')
  assert.ok(drawer > deep, 'evidence drawer follows deep analysis')
})

test('compatibility result page no longer renders legacy decision dashboard inline', () => {
  const source = read(pagePath)
  assert.doesNotMatch(source, /<DecisionDashboardPanel/)
  assert.doesNotMatch(source, /<DecisionEvidenceSummary/)
  assert.doesNotMatch(source, /<ResultReadingMap/)
})
```

Remove the old "renders decision-dashboard sections before scores and AI report" test as it's no longer applicable.

The "compatibility result page includes decision-dashboard CSS hooks" test should also be rewritten or removed — those CSS hooks live in extracted components now.

- [ ] **Step 7: Update `compatibility-result-ux.test.mjs`**

In `frontend/tests/compatibility-result-ux.test.mjs`, the test "compatibility result page exposes conclusion-first mobile sections" asserts `compatibility-quick-score` and `compatibility-decision-metric-grid` exist. Those classes are now in extracted components, not in the page itself. Rewrite the assertions to point at the new structure:

```js
test('compatibility result page renders sticky + 3 sections + drawer', () => {
  const page = read('src/pages/CompatibilityResultPage.tsx')
  assert.match(page, /CompatibilityStickyHeader/)
  assert.match(page, /SectionBasicCharts/)
  assert.match(page, /SectionVerdict/)
  assert.match(page, /SectionDeepAnalysis/)
  assert.match(page, /EvidenceDrawer/)
})
```

For tests inside `compatibility-result-ux.test.mjs` that grep for "最大风险" or "下一步验证" — those strings have moved into `NextStepsAndAvoid.tsx`. Update file path in the read:

```js
test('NextStepsAndAvoid mentions verification action and short-term avoid', () => {
  const src = read('src/components/compatibility/deep-analysis/NextStepsAndAvoid.tsx')
  assert.match(src, /下一步验证/)
  assert.match(src, /短期避免/)
})
```

Remove or rewrite any test that depends on `.compatibility-decision-hero` or `.compatibility-personality-fit` — these classes are obsolete.

- [ ] **Step 8: Run full test suite + build**

```bash
cd frontend && npm run build && node --test tests/
```

Expected: ALL tests PASS. If any test fails because it references a removed class/string, update or remove that assertion. **Do not silence failures with `.skip` — fix the test.**

- [ ] **Step 9: Manual smoke**

`npm run dev`, open the result page on mobile + desktop. Verify:
- [ ] Sticky + Section ① + Section ② + Section ③ + Drawer in order, no duplicate panels
- [ ] All `<details>` default open
- [ ] sticky verdict link is clickable on desktop, jumps to Section ②, section title not hidden by sticky
- [ ] BottomNav still visible, no content cut-off
- [ ] 分享图片按钮可点 → 弹出 ShareCard，与改前像素级一致（拍照对比批次 1 前的截图）
- [ ] 导出 PDF 按钮可点 → PDF 与改前像素级一致

- [ ] **Step 10: Commit**

```bash
cd frontend && git add src/pages/CompatibilityResultPage.tsx tests/compatibility-decision-dashboard.test.mjs tests/compatibility-result-ux.test.mjs
git commit -m "chore(compat-result): remove ENABLE_NEW_LAYOUT flag; delete legacy 11-section path"
```

---

### Task 22: 拆分 `CompatibilityResultPage.css` — 把每个 panel 的样式移到对应组件 `.css`

**Files:**
- Modify: `frontend/src/pages/CompatibilityResultPage.css`
- Modify: 各 `frontend/src/components/compatibility/**/*.css`

- [ ] **Step 1: Identify orphan CSS classes (no longer used)**

```bash
cd frontend && grep -E "^\.[a-zA-Z]" src/pages/CompatibilityResultPage.css | awk '{print $1}' | sort -u > /tmp/page-classes.txt
grep -rh "className=" src/ | grep -oE '[a-zA-Z][a-zA-Z0-9_-]+' | sort -u > /tmp/used-classes.txt
comm -23 /tmp/page-classes.txt /tmp/used-classes.txt | head -50
```

Expected: a list of class names that exist in page CSS but aren't used anywhere — these are orphans.

- [ ] **Step 2: Delete orphan blocks**

For each class in the orphan list, find and delete the CSS block in `frontend/src/pages/CompatibilityResultPage.css`. Common orphans expected: `.compatibility-decision-dashboard`, `.compatibility-decision-header`, `.compatibility-decision-context`, `.compatibility-decision-headline`, `.compatibility-decision-type`, `.compatibility-decision-metric-*`, `.compatibility-next-action`, `.compatibility-avoid-list`, `.compatibility-core-contradiction`, `.compatibility-section-*`, `.compatibility-section-header`, `.compatibility-personality-fit*` (the ones unused), `.compatibility-stage-validation-*`, `.compatibility-relationship-strategy-*`, `.compatibility-result-map`, `.compatibility-decision-evidence-*`, `.compatibility-claim-evidence-*`, `.compatibility-ai-card`.

- [ ] **Step 3: Move CSS for `compatibility-evidence-*` classes to `EvidenceDrawer.css`**

`grep -n "compatibility-evidence-" src/pages/CompatibilityResultPage.css` → relocate those blocks into `frontend/src/components/compatibility/EvidenceDrawer.css`. Delete them from the page CSS.

- [ ] **Step 4: Move CSS for `compatibility-person-*`, `compatibility-pillar-*`, `compatibility-wuxing-*`, `compatibility-day-master*`, `wuxing-badge` to `ParticipantSummaryCard.css`**

These are needed by `ParticipantSummaryCard.tsx`. Create `frontend/src/components/compatibility/ParticipantSummaryCard.css` (it doesn't exist yet — Task 5 only created the `.tsx`). Add `import './ParticipantSummaryCard.css'` at the top of `ParticipantSummaryCard.tsx`. Then `grep -n "compatibility-person-\|compatibility-pillar-\|compatibility-wuxing-\|compatibility-day-master\|wuxing-badge" src/pages/CompatibilityResultPage.css` to find the blocks, cut them, and paste into the new CSS file.

- [ ] **Step 5: Move CSS for `compat-score-v3*`, `compatibility-quick-score*` to `ScoreOverview.css`**

Create `frontend/src/components/compatibility/ScoreOverview.css` and import it from `ScoreOverview.tsx`. Move the blocks.

- [ ] **Step 6: Move CSS for `compatibility-ai-*` (any remaining) to `DeepReportNarrative.css`**

Task 18 may have already renamed some classes to `compat-da-report__*`. Any leftover `.compatibility-ai-*` should move to `DeepReportNarrative.css` (or be renamed at the source).

- [ ] **Step 7: Move CSS for `compat-export-actions` — keep in main page CSS**

This block stays in `frontend/src/pages/CompatibilityResultPage.css` (it's a page-level concern).

- [ ] **Step 8: Build + tests + manual visual smoke**

```bash
cd frontend && npm run build && node --test tests/
```

Manual:
- [ ] Visually compare each subsection with before-CSS-move screenshot — no style regression
- [ ] Sticky still works
- [ ] Section ① card styling intact
- [ ] Section ② layout intact
- [ ] Section ③ subsections intact
- [ ] Evidence drawer intact

- [ ] **Step 9: Commit**

```bash
cd frontend && git add src/pages/CompatibilityResultPage.css src/components/compatibility/
git commit -m "refactor(compat-result): split CSS per-component; delete orphan blocks from page CSS"
```

---

### Task 23: 验证文件行数达标 + final QA

**Files:**
- N/A (verification only)

- [ ] **Step 1: Verify TSX line counts**

```bash
cd frontend && wc -l src/pages/CompatibilityResultPage.tsx src/components/compatibility/**/*.tsx
```

Expected:
- `src/pages/CompatibilityResultPage.tsx` < 500 行
- 任一新组件文件 < 250 行

If any file exceeds, look for further extraction opportunities (e.g., if `DeepReportNarrative.tsx` is too long, extract its structured-dimension sub-renderer to a child component).

- [ ] **Step 2: Verify CSS no longer mixes 4 grey layers**

```bash
cd frontend && grep -n "bg-card-hover\|bg-elevated" src/pages/CompatibilityResultPage.css src/components/compatibility/**/*.css
```

Expected: empty.

- [ ] **Step 3: Manual QA checklist (run on a real reading detail page)**

Run `npm run dev`. Open `/compatibility/result/<existing reading id>` in mobile-sized window AND in desktop-sized window:

- [ ] 手机（≤ 414px）：sticky 始终可见；三段单列；段 ① 单列；段 ② 单列；段 ③ 子段单列
- [ ] 桌面（≥ 1024px）：sticky 高度变大；容器 max-width 900px 居中；段 ① 双列；段 ② verdict+score / findings 双列
- [ ] sticky 摘要点击 verdict 链接 → 滚动到段 ②，段标题不被 sticky 遮挡
- [ ] 所有 `<details>` 默认全展开
- [ ] v3 用户：sticky 显示 `overall_score`，段 ② 显示 `ScoreOverviewV3`
- [ ] legacy 用户：sticky 显示 `overall_score`，段 ② 显示 `ScoreOverview`
- [ ] 打开"分享图片" → 弹出 ShareCard，与改前像素级一致
- [ ] 打开"导出 PDF" → PDF 输出与改前像素级一致
- [ ] 移动端底部 BottomNav 不被遮挡；证据抽屉滚到底可见

- [ ] **Step 4: 截图归档（可选）**

把手机+桌面的最终页面截图 + 分享卡截图 + PDF 首页截图保存为 `docs/superpowers/specs/2026-05-29-compat-result-page-restructure-design.qa-screenshots/` 下的图片。这不是必须，但便于后续 PR 评审。

- [ ] **Step 5: Final commit (if any cleanup happened)**

```bash
cd frontend && git status
# 如果有未提交的清理改动：
git add . && git commit -m "chore(compat-result): final cleanup after layout restructure"
```

---

## 自审 / Self-Review Checklist

**Spec coverage:**

- [x] Spec §3.1 信息架构 → Tasks 4 / 8 / 13 / 19 / 20（每段 sticky/①/②/③/drawer 都对应至少一个 Task）
- [x] Spec §3.2 组件树 → Tasks 3 / 5 / 6 / 7 / 11 / 12 / 14 / 15 / 16 / 17 / 18 / 19 / 20
- [x] Spec §3.3 数据流（不改 props 接口）→ 各 Task 都明确说"现有 helper 不动"
- [x] Spec §3.4 Sticky 摘要规则 → Task 6（CSS、aria、移动/桌面）
- [x] Spec §3.5 响应式栅格 → Task 1（tokens）+ 各 section CSS 的 `@media`
- [x] Spec §3.6 视觉 token + 节奏 → Task 1
- [x] Spec §3.7 折叠默认展开 → Task 10（既有）+ Tasks 14-18 + Task 20（新建时即用 `<details open>`）
- [x] Spec §3.8 导出一致性 → Task 23 手动 QA Step 3 验证像素级一致
- [x] Spec §3.9 测试策略 → 每个 Task 都有源码 grep 测试；Task 23 含手动 QA
- [x] Spec §4 4 批次 → 计划用"批次 1-4"分组对应

**Placeholder scan:** 无 "TBD" / "TODO" / "实现稍后" 留在步骤里。Tasks 12 / 18 / 20 三处大块代码（ScoreOverview / DeepReportNarrative / EvidenceDrawer）采用"原样粘贴 page 行 X-Y"的指令模式而不是把数百行源码塞进 plan，这是有意的——既避免 plan 文档臃肿，也避免我对照源码手抄时出错（V3 维度键就是这么差点写错的）。每个"原样粘贴"指令都带具体行号 + 粘贴后的 sanity check 命令（grep 计数比对）。

**Type consistency:** 
- `DecisionDashboardData` (来自 `compatibilityDecision.ts`) — 各 Task 一致使用
- `PersonalityFitSummary` / `PersonalityValidationPlan` — 一致
- `CompatibilityStructuredReport` / `CompatibilityDimensionReport` — 一致
- 新组件 props 接口在创建时定义，下游 Task 复用时 import 自同一文件 — 一致

---

## 执行交接

Plan complete and saved to `docs/superpowers/plans/2026-05-29-compat-result-page-restructure.md`. Two execution options:

**1. Subagent-Driven (recommended)** — 每个 Task 派一个 fresh subagent；任务间两阶段评审；适合本计划这种 23 个串行 Task、依赖链清晰的场景。

**2. Inline Execution** — 在当前会话里按 Task 顺序连续执行；过程中可以在每个批次后停下来检查（批次 1 → 2 → 3 → 4 之间设 checkpoint）。

哪种？
