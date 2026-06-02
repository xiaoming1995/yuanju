import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import test from 'node:test'
import assert from 'node:assert/strict'

const root = new URL('..', import.meta.url).pathname
const read = (p) => readFileSync(join(root, p), 'utf8')

// covers tasks.md 10.2 / 10.3 / 11.2 of optimize-past-events-token-cost

test('past-events page initializes folded=true for future dayuns', () => {
  const page = read('src/pages/PastEventsPage.tsx')
  // 未来段的 start_year 大于 currentYear → folded
  assert.match(page, /isFuture\s*=\s*dm\.start_year\s*>\s*currentYear/)
  assert.match(page, /isFuture[\s\S]{1,300}folded:\s*true/)
})

test('folded dayun segment renders expand button and no chips', () => {
  const page = read('src/pages/PastEventsPage.tsx')
  // 折叠态切换按钮
  assert.match(page, /展开 ▼|收起 ▲/)
  // 折叠时显示提示（"未来大运段"），不显示 chips/year cards
  assert.match(page, /\{dySum\?\.folded\s*&&/)
  assert.match(page, /未来大运段，点击/)
  // 年份列表和整体总结被包在 !folded 条件里
  assert.match(page, /!dySum\?\.folded\s*&&/)
})

test('expanded uncached future dayun shows the "生成本段 AI 批语" button', () => {
  const page = read('src/pages/PastEventsPage.tsx')
  // 当 folded=false 且无 years 且无 loading/error 时显示生成按钮
  assert.match(page, /dySum\?\.folded === false/)
  assert.match(page, /生成本段 AI 批语/)
  // 按钮要 wire 到 handleGenerateSegment
  assert.match(page, /onClick=\{\(\)\s*=>\s*handleGenerateSegment\(meta\.index,\s*'manual'\)\}/)
})

test('handleExpand only flips folded state — no network call', () => {
  const page = read('src/pages/PastEventsPage.tsx')
  // 提取 handleExpand 函数体（到下一个 const declaration）
  const m = page.match(/const handleExpand[\s\S]*?\}, \[\]\)/)
  assert.ok(m, 'handleExpand not found')
  const body = m[0]
  // 不应包含任何 fetch / streamDayunSummaries / api 调用
  assert.ok(!/fetch\s*\(/.test(body), `handleExpand should not call fetch; body: ${body}`)
  assert.ok(!/streamDayunSummaries/.test(body), `handleExpand should not stream API; body: ${body}`)
  assert.ok(!/baziAPI\./.test(body), `handleExpand should not call baziAPI; body: ${body}`)
  // 必须调用 setSummaries 翻 folded
  assert.match(body, /setSummaries/)
  assert.match(body, /folded:\s*false/)
})

test('handleGenerateSegment passes single dayunIndex to streamDayunSummaries', () => {
  const page = read('src/pages/PastEventsPage.tsx')
  const m = page.match(/const handleGenerateSegment[\s\S]*?\}, \[beginDayunGeneration,\s*chartId,\s*markDayunInterrupted,\s*writeDayunSummary\]\)/)
  assert.ok(m, 'handleGenerateSegment not found')
  const body = m[0]
  // 必须把单段 [dayunIndex] 传给 stream
  assert.match(body, /streamDayunSummaries/)
  assert.match(body, /\[dayunIndex\]/)
})

test('streamDayunSummaries API accepts optional dayunIndexes parameter', () => {
  const api = read('src/lib/api.ts')
  // 函数签名末尾应有 dayunIndexes?: number[]
  assert.match(api, /dayunIndexes\?:\s*number\[\]/)
  // body 编码逻辑：dayunIndexes 非空时 JSON 化为 { dayun_indexes }
  assert.match(api, /dayun_indexes:\s*dayunIndexes/)
})
