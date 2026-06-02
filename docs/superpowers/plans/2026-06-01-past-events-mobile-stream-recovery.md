# Past Events Mobile Stream Recovery Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix mobile background/foreground stream interruptions so past-events dayun summaries and section narratives never stay in endless loading.

**Architecture:** Keep the fix frontend-only. Add terminal-state protection to `streamDayunSummaries`, then make `PastEventsPage` track per-dayun generation attempts and recover stale loading entries when the browser returns from the background.

**Tech Stack:** React 19, TypeScript, Vite, browser `fetch` streaming, `AbortController`, `node:test` static wiring tests, existing CSS variables.

---

## File Structure

- Modify `frontend/src/lib/api.ts`: add stream timeout/abort protection and a mutually terminal `onDone`/`onError` guard to `streamDayunSummaries`.
- Modify `frontend/src/pages/PastEventsPage.tsx`: add per-dayun generation metadata, interrupted state, lifecycle listeners, retry-on-return logic, and retry UI copy.
- Create `frontend/tests/past-events-mobile-stream-recovery.test.mjs`: static wiring tests that enforce the new stream and page recovery guards.

The existing code keeps page logic in `PastEventsPage.tsx`; this plan does not split the file because the change is tightly coupled to current page state and existing test style is source-based.

---

### Task 1: Add Failing Recovery Wiring Tests

**Files:**
- Create: `frontend/tests/past-events-mobile-stream-recovery.test.mjs`
- Inspect: `frontend/src/lib/api.ts`
- Inspect: `frontend/src/pages/PastEventsPage.tsx`

- [ ] **Step 1: Create the failing test file**

Create `frontend/tests/past-events-mobile-stream-recovery.test.mjs` with this content:

```js
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import test from 'node:test'
import assert from 'node:assert/strict'

const root = new URL('..', import.meta.url).pathname
const read = (p) => readFileSync(join(root, p), 'utf8')

test('streamDayunSummaries has terminal timeout and abort protection', () => {
  const api = read('src/lib/api.ts')
  assert.match(api, /DAYUN_STREAM_INACTIVITY_TIMEOUT_MS\s*=\s*45_000/)
  assert.match(api, /interface\s+StreamDayunSummariesOptions/)
  assert.match(api, /AbortController/)
  assert.match(api, /resetInactivityTimer/)
  assert.match(api, /safeOnError/)
  assert.match(api, /controller\.abort\(\)/)
  assert.match(api, /clearTimeout\(inactivityTimer\)/)
})

test('PastEventsPage tracks generation metadata for loading dayun summaries', () => {
  const page = read('src/pages/PastEventsPage.tsx')
  assert.match(page, /DAYUN_GENERATION_STALE_MS\s*=\s*10_000/)
  assert.match(page, /DAYUN_GENERATION_MAX_AUTO_RETRIES\s*=\s*1/)
  assert.match(page, /type\s+DayunGenerationSource\s*=\s*'initial'\s*\|\s*'manual'\s*\|\s*'recovery'/)
  assert.match(page, /generation\?:\s*DayunGenerationMeta/)
  assert.match(page, /status\?:\s*'loading'\s*\|\s*'interrupted'/)
})

test('PastEventsPage listens for mobile return signals and recovers stale loading summaries', () => {
  const page = read('src/pages/PastEventsPage.tsx')
  assert.match(page, /recoverStaleDayunSummaries/)
  assert.match(page, /document\.addEventListener\('visibilitychange'/)
  assert.match(page, /window\.addEventListener\('pageshow'/)
  assert.match(page, /window\.addEventListener\('focus'/)
  assert.match(page, /window\.addEventListener\('online'/)
  assert.match(page, /document\.visibilityState\s*===\s*'visible'/)
  assert.match(page, /handleGenerateSegment\(index,\s*'recovery'\)/)
})

test('interrupted dayun generation renders retry copy instead of endless loading', () => {
  const page = read('src/pages/PastEventsPage.tsx')
  assert.match(page, /生成中断，点击重试/)
  assert.match(page, /生成失败，请重试/)
  assert.match(page, /dySum\.status\s*===\s*'interrupted'/)
  assert.match(page, /handleGenerateSegment\(meta\.index,\s*'manual'\)/)
})
```

- [ ] **Step 2: Run the new test and verify it fails**

Run:

```bash
node --test frontend/tests/past-events-mobile-stream-recovery.test.mjs
```

Expected: FAIL because the constants, metadata fields, lifecycle listeners, and timeout guards do not exist yet.

- [ ] **Step 3: Confirm no existing past-events test is broken before implementation**

Run:

```bash
node --test frontend/tests/past-events-progressive-generation.test.mjs
```

Expected: PASS. If this fails before editing source code, inspect the current dirty workspace before continuing.

---

### Task 2: Add Stream Terminal-State Protection

**Files:**
- Modify: `frontend/src/lib/api.ts`
- Test: `frontend/tests/past-events-mobile-stream-recovery.test.mjs`

- [ ] **Step 1: Add stream timeout types and constant near `errorMessage`**

Add this near the top of `frontend/src/lib/api.ts`, after `errorMessage`:

```ts
const DAYUN_STREAM_INACTIVITY_TIMEOUT_MS = 45_000
const DAYUN_STREAM_INTERRUPTED_MESSAGE = '生成中断，点击重试'

interface StreamDayunSummariesOptions {
  inactivityTimeoutMs?: number
  signal?: AbortSignal
}
```

- [ ] **Step 2: Replace `streamDayunSummaries` with terminal-safe implementation**

Replace the existing `streamDayunSummaries` function body and signature with this version:

```ts
streamDayunSummaries: async (
  chartId: string,
  onItem: (item: {
    dayun_index: number
    gan_zhi: string
    themes: string[]
    summary: string
    years?: { year: number; ganzhi: string; narrative: string }[]
    cached: boolean
    error?: string
  }) => void,
  onError: (err: string) => void,
  onDone: () => void,
  dayunIndexes?: number[],
  options: StreamDayunSummariesOptions = {},
) => {
  const token = localStorage.getItem('yj_token')
  const baseURL = import.meta.env.VITE_API_URL || ''
  const controller = new AbortController()
  const timeoutMs = options.inactivityTimeoutMs ?? DAYUN_STREAM_INACTIVITY_TIMEOUT_MS
  let inactivityTimer: ReturnType<typeof setTimeout> | undefined
  let settled = false
  let timedOut = false

  const cleanup = () => {
    if (inactivityTimer) {
      clearTimeout(inactivityTimer)
      inactivityTimer = undefined
    }
    options.signal?.removeEventListener('abort', abortFromParent)
  }

  const safeOnDone = () => {
    if (settled) return
    settled = true
    cleanup()
    onDone()
  }

  const safeOnError = (err: string) => {
    if (settled) return
    settled = true
    cleanup()
    onError(err)
  }

  const abortFromParent = () => {
    controller.abort()
  }

  const resetInactivityTimer = () => {
    if (inactivityTimer) clearTimeout(inactivityTimer)
    inactivityTimer = setTimeout(() => {
      timedOut = true
      controller.abort()
    }, timeoutMs)
  }

  try {
    if (options.signal?.aborted) {
      safeOnError(DAYUN_STREAM_INTERRUPTED_MESSAGE)
      return
    }
    options.signal?.addEventListener('abort', abortFromParent, { once: true })
    resetInactivityTimer()

    const response = await fetch(`${baseURL}/api/bazi/past-events/dayun-summary-stream/${chartId}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', ...(token ? { 'Authorization': `Bearer ${token}` } : {}) },
      body: dayunIndexes && dayunIndexes.length > 0
        ? JSON.stringify({ dayun_indexes: dayunIndexes })
        : undefined,
      signal: controller.signal,
    })
    if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`)
    const reader = response.body?.getReader()
    if (!reader) throw new Error('No reader available')
    const decoder = new TextDecoder()
    let buffer = ''
    let pendingError = false
    while (true) {
      const { done, value } = await reader.read()
      if (done) break
      resetInactivityTimer()
      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split('\n')
      buffer = lines.pop() || ''
      for (const line of lines) {
        if (line.startsWith('event: error')) {
          pendingError = true
        } else if (line.startsWith('event: done')) {
          safeOnDone()
          return
        } else if (line.startsWith('data: ')) {
          const data = line.slice(6)
          if (data === '[DONE]') {
            safeOnDone()
            return
          }
          if (pendingError) {
            safeOnError(data)
            pendingError = false
            return
          }
          try {
            const parsed = JSON.parse(data)
            onItem(parsed)
          } catch {
            // Ignore malformed SSE data and continue reading the stream.
          }
        }
      }
    }
    safeOnDone()
  } catch (err: unknown) {
    const interrupted = timedOut || (err instanceof DOMException && err.name === 'AbortError')
    safeOnError(interrupted ? DAYUN_STREAM_INTERRUPTED_MESSAGE : errorMessage(err))
  }
},
```

- [ ] **Step 3: Run the stream wiring test**

Run:

```bash
node --test frontend/tests/past-events-mobile-stream-recovery.test.mjs
```

Expected: still FAIL, but the failures should now be only for `PastEventsPage.tsx` recovery wiring.

- [ ] **Step 4: Run TypeScript build after API change**

Run:

```bash
cd frontend && npm run build
```

Expected: PASS. If TypeScript reports `DOMException` or timer typing issues, keep the same behavior and adjust the type expression only.

---

### Task 3: Add Page Recovery State and Lifecycle Hooks

**Files:**
- Modify: `frontend/src/pages/PastEventsPage.tsx`
- Test: `frontend/tests/past-events-mobile-stream-recovery.test.mjs`
- Test: `frontend/tests/past-events-progressive-generation.test.mjs`

- [ ] **Step 1: Add generation types and constants after `DayunSummary`**

Add this block after the `DayunSummary` interface:

```ts
type DayunGenerationSource = 'initial' | 'manual' | 'recovery'

interface DayunGenerationMeta {
  startedAt: number
  attempt: number
  source: DayunGenerationSource
}

const DAYUN_GENERATION_STALE_MS = 10_000
const DAYUN_GENERATION_MAX_AUTO_RETRIES = 1
const DAYUN_GENERATION_INTERRUPTED_COPY = '生成中断，点击重试'
const DAYUN_GENERATION_FAILED_COPY = '生成失败，请重试'
```

Extend `DayunSummary` with these fields:

```ts
  status?: 'loading' | 'interrupted'
  generation?: DayunGenerationMeta
```

- [ ] **Step 2: Add latest summary ref after `inflightRef`**

Add this after `const inflightRef = useRef(false)`:

```ts
  const summariesRef = useRef<Record<number, DayunSummary>>({})
```

Add this effect after the state declarations:

```ts
  useEffect(() => {
    summariesRef.current = summaries
  }, [summaries])
```

- [ ] **Step 3: Add helper functions before `loadAll`**

Add these callbacks before `const loadAll = useCallback(...)`:

```ts
  const beginDayunGeneration = useCallback((dayunIndex: number, source: DayunGenerationSource) => {
    setSummaries((prev) => {
      const existing = prev[dayunIndex]
      const previousAttempt = existing?.generation?.attempt ?? 0
      const attempt = source === 'recovery' ? previousAttempt + 1 : previousAttempt
      return {
        ...prev,
        [dayunIndex]: {
          ...existing,
          themes: existing?.themes || [],
          summary: existing?.summary || '',
          loading: true,
          status: 'loading',
          error: undefined,
          folded: false,
          generation: {
            startedAt: Date.now(),
            attempt,
            source,
          },
        },
      }
    })
  }, [])

  const markDayunInterrupted = useCallback((dayunIndex: number, message = DAYUN_GENERATION_INTERRUPTED_COPY) => {
    setSummaries((prev) => ({
      ...prev,
      [dayunIndex]: {
        ...prev[dayunIndex],
        themes: prev[dayunIndex]?.themes || [],
        summary: prev[dayunIndex]?.summary || '',
        loading: false,
        status: 'interrupted',
        error: message,
        folded: false,
      },
    }))
  }, [])

  const markLoadingDayunsInterrupted = useCallback((message = DAYUN_GENERATION_INTERRUPTED_COPY) => {
    setSummaries((prev) => {
      const next = { ...prev }
      for (const [key, summary] of Object.entries(prev)) {
        if (summary.loading) {
          next[Number(key)] = {
            ...summary,
            loading: false,
            status: 'interrupted',
            error: message,
            folded: false,
          }
        }
      }
      return next
    })
  }, [])
```

- [ ] **Step 4: Mark initial summaries with generation metadata**

In `loadAll`, replace this initial non-future summary:

```ts
: { themes: [], summary: '', loading: true }
```

with:

```ts
: {
    themes: [],
    summary: '',
    loading: true,
    status: 'loading',
    generation: { startedAt: Date.now(), attempt: 0, source: 'initial' },
  }
```

- [ ] **Step 5: Clear loading metadata on successful stream items**

In both `streamDayunSummaries` item handlers, ensure success writes include:

```ts
loading: false,
status: undefined,
generation: undefined,
folded: false,
```

For item-level errors, use:

```ts
next[item.dayun_index] = {
  ...next[item.dayun_index],
  themes: [],
  summary: '',
  error: item.error,
  loading: false,
  status: 'interrupted',
  folded: false,
}
```

- [ ] **Step 6: Convert initial stream errors into interrupted summaries**

In the initial `streamDayunSummaries` error callback inside `loadAll`, change it to:

```ts
(err) => {
  setStreamError(err)
  inflightRef.current = false
  markLoadingDayunsInterrupted(err || DAYUN_GENERATION_INTERRUPTED_COPY)
},
```

Add `markLoadingDayunsInterrupted` to the `loadAll` dependency array.

- [ ] **Step 7: Update manual generation signature and terminal states**

Change the function declaration:

```ts
const handleGenerateSegment = useCallback((dayunIndex: number, source: DayunGenerationSource = 'manual') => {
```

Replace its initial `setSummaries` call with:

```ts
beginDayunGeneration(dayunIndex, source)
```

In its error callback, write:

```ts
(err) => {
  markDayunInterrupted(dayunIndex, err || DAYUN_GENERATION_FAILED_COPY)
},
```

Update the dependency array to:

```ts
}, [beginDayunGeneration, chartId, markDayunInterrupted])
```

- [ ] **Step 8: Add stale recovery callback**

Add this callback after `handleGenerateSegment`:

```ts
  const recoverStaleDayunSummaries = useCallback(() => {
    const now = Date.now()
    const staleIndexes = Object.entries(summariesRef.current)
      .filter(([, summary]) => {
        if (!summary.loading || !summary.generation) return false
        return now - summary.generation.startedAt >= DAYUN_GENERATION_STALE_MS
      })
      .map(([index]) => Number(index))

    if (staleIndexes.length === 0) return
    inflightRef.current = false

    for (const index of staleIndexes) {
      const summary = summariesRef.current[index]
      const attempt = summary?.generation?.attempt ?? 0
      if (attempt < DAYUN_GENERATION_MAX_AUTO_RETRIES) {
        handleGenerateSegment(index, 'recovery')
      } else {
        markDayunInterrupted(index)
      }
    }
  }, [handleGenerateSegment, markDayunInterrupted])
```

- [ ] **Step 9: Add mobile return listeners**

Add this effect after `recoverStaleDayunSummaries`:

```ts
  useEffect(() => {
    const recoverIfVisible = () => {
      if (document.visibilityState === 'visible') {
        recoverStaleDayunSummaries()
      }
    }

    document.addEventListener('visibilitychange', recoverIfVisible)
    window.addEventListener('pageshow', recoverStaleDayunSummaries)
    window.addEventListener('focus', recoverStaleDayunSummaries)
    window.addEventListener('online', recoverStaleDayunSummaries)
    return () => {
      document.removeEventListener('visibilitychange', recoverIfVisible)
      window.removeEventListener('pageshow', recoverStaleDayunSummaries)
      window.removeEventListener('focus', recoverStaleDayunSummaries)
      window.removeEventListener('online', recoverStaleDayunSummaries)
    }
  }, [recoverStaleDayunSummaries])
```

- [ ] **Step 10: Update interrupted UI rendering**

Inside the dayun summary block, add interrupted handling before the generic error branch:

```tsx
{dySum.status === 'interrupted' && (
  <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 10, color: '#e7c766', fontSize: '0.78rem' }}>
    <span>{dySum.error || DAYUN_GENERATION_INTERRUPTED_COPY}</span>
    <button
      type="button"
      onClick={() => handleGenerateSegment(meta.index, 'manual')}
      style={{
        border: '1px solid color-mix(in srgb, var(--wu-jin) 45%, transparent)',
        background: 'transparent',
        color: 'var(--wu-jin)',
        borderRadius: 6,
        padding: '4px 8px',
        cursor: 'pointer',
        fontSize: '0.72rem',
      }}
    >
      重试
    </button>
  </div>
)}
```

Change the existing error branch condition from:

```tsx
{dySum.error && (
```

to:

```tsx
{dySum.error && dySum.status !== 'interrupted' && (
```

Change the ready branch condition from:

```tsx
{!dySum.loading && !dySum.error && dySum.themes.length > 0 && (
```

to:

```tsx
{!dySum.loading && dySum.status !== 'interrupted' && !dySum.error && dySum.themes.length > 0 && (
```

- [ ] **Step 11: Update manual generate button call**

Change:

```tsx
onClick={() => handleGenerateSegment(meta.index)}
```

to:

```tsx
onClick={() => handleGenerateSegment(meta.index, 'manual')}
```

- [ ] **Step 12: Run page wiring tests**

Run:

```bash
node --test frontend/tests/past-events-mobile-stream-recovery.test.mjs
```

Expected: PASS.

- [ ] **Step 13: Run existing progressive generation tests**

Run:

```bash
node --test frontend/tests/past-events-progressive-generation.test.mjs
```

Expected: PASS.

---

### Task 4: Final Verification

**Files:**
- Verify: `frontend/src/lib/api.ts`
- Verify: `frontend/src/pages/PastEventsPage.tsx`
- Verify: `frontend/tests/past-events-mobile-stream-recovery.test.mjs`
- Verify: `frontend/tests/past-events-progressive-generation.test.mjs`

- [ ] **Step 1: Run lint**

Run:

```bash
cd frontend && npm run lint
```

Expected: PASS with no ESLint errors.

- [ ] **Step 2: Run production build**

Run:

```bash
cd frontend && npm run build
```

Expected: PASS with TypeScript and Vite build complete.

- [ ] **Step 3: Review changed files**

Run:

```bash
git diff -- frontend/src/lib/api.ts frontend/src/pages/PastEventsPage.tsx frontend/tests/past-events-mobile-stream-recovery.test.mjs frontend/tests/past-events-progressive-generation.test.mjs
```

Expected: diff only contains stream terminal-state protection, page recovery logic, retry UI, and the new test file.

- [ ] **Step 4: Commit implementation**

Run:

```bash
git add frontend/src/lib/api.ts frontend/src/pages/PastEventsPage.tsx frontend/tests/past-events-mobile-stream-recovery.test.mjs
git commit -m "fix: recover interrupted past events streams"
```

Expected: one implementation commit. Do not stage unrelated existing UX audit or result-page changes.

---

## Self-Review

- Spec coverage: page-level recovery is covered in Task 3; stream terminal-state protection is covered in Task 2; retry UI and testing are covered in Tasks 1, 3, and 4.
- Placeholder scan: the plan contains concrete file paths, commands, code snippets, and expected outputs.
- Type consistency: `DayunGenerationSource`, `DayunGenerationMeta`, `DayunSummary.status`, and `StreamDayunSummariesOptions` are named consistently across tasks.
- Scope check: the plan is frontend-only and does not alter backend APIs, prompts, bazi algorithms, or layout direction.
