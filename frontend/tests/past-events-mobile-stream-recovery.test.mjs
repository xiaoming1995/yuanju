import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import test from 'node:test'
import assert from 'node:assert/strict'

const root = new URL('..', import.meta.url).pathname
const read = (p) => readFileSync(join(root, p), 'utf8')

const streamDayunSummariesSource = () => {
  const api = read('src/lib/api.ts')
  const start = api.indexOf('streamDayunSummaries: async (')
  const end = api.indexOf('streamPastEvents:', start)
  assert.notEqual(start, -1, 'streamDayunSummaries implementation not found')
  assert.notEqual(end, -1, 'streamPastEvents boundary not found')
  return api.slice(start, end)
}

test('streamDayunSummaries has terminal timeout and abort protection', () => {
  const api = read('src/lib/api.ts')
  const stream = streamDayunSummariesSource()
  assert.match(api, /DAYUN_STREAM_INACTIVITY_TIMEOUT_MS\s*=\s*45_000/)
  assert.match(api, /interface\s+StreamDayunSummariesOptions/)
  assert.match(stream, /AbortController/)
  assert.match(stream, /resetInactivityTimer/)
  assert.match(stream, /let\s+settled\s*=\s*false/)
  assert.match(stream, /const\s+safeOnDone\s*=\s*\(\)\s*=>\s*\{[\s\S]*?if\s*\(settled\)\s*return[\s\S]*?settled\s*=\s*true[\s\S]*?onDone\(\)/)
  assert.match(stream, /const\s+safeOnError\s*=\s*\(err:\s*string\)\s*=>\s*\{[\s\S]*?if\s*\(settled\)\s*return[\s\S]*?settled\s*=\s*true[\s\S]*?onError\(err\)/)
  assert.match(stream, /controller\.abort\(\)/)
  assert.match(stream, /clearTimeout\(inactivityTimer\)/)
})

test('PastEventsPage tracks generation metadata for loading dayun summaries', () => {
  const page = read('src/pages/PastEventsPage.tsx')
  assert.match(page, /DAYUN_GENERATION_STALE_MS\s*=\s*10_000/)
  assert.match(page, /DAYUN_GENERATION_MAX_AUTO_RETRIES\s*=\s*1/)
  assert.match(page, /type\s+DayunGenerationSource\s*=\s*'initial'\s*\|\s*'manual'\s*\|\s*'recovery'/)
  assert.match(page, /generation\?:\s*DayunGenerationMeta/)
  assert.match(page, /status\?:\s*'loading'\s*\|\s*'interrupted'/)
  assert.match(page, /requestId:\s*number/)
  assert.match(page, /generationRequestSeqRef\s*=\s*useRef\(0\)/)
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

test('PastEventsPage fences stale stream callbacks by generation request id', () => {
  const page = read('src/pages/PastEventsPage.tsx')
  assert.match(page, /const requestId\s*=\s*\+\+generationRequestSeqRef\.current/)
  assert.match(page, /const requestId\s*=\s*beginDayunGeneration\(dayunIndex,\s*source\)/)
  assert.match(page, /current\?\.generation\?\.requestId\s*!==\s*requestId/)
  assert.match(page, /initialGenerationRequestIds\[item\.dayun_index\]/)
  assert.match(page, /current\?\.generation\?\.requestId\s*!==\s*expectedRequestId/)
  assert.match(page, /markLoadingDayunsInterrupted\(err \|\| DAYUN_GENERATION_INTERRUPTED_COPY,\s*'initial'\)/)
})

test('PastEventsPage synchronously updates summary ref during generation and interruption', () => {
  const page = read('src/pages/PastEventsPage.tsx')
  assert.match(page, /summariesRef\.current\s*=\s*\{[\s\S]*?\.\.\.summariesRef\.current[\s\S]*?\[dayunIndex\]:\s*nextSummary[\s\S]*?\}/)
  assert.match(page, /summariesRef\.current\s*=\s*next/)
})

test('PastEventsPage derives header stream status from summary state', () => {
  const page = read('src/pages/PastEventsPage.tsx')
  assert.match(page, /const hasLoadingSummary\s*=/)
  assert.match(page, /const hasInterruptedSummary\s*=/)
  assert.match(page, /hasLoadingSummary\s*\?\s*'年份已就绪 · 大运总结正在后台生成'/)
  assert.match(page, /hasInterruptedSummary\s*\?\s*'部分大运总结生成中断，可点击重试'/)
})
