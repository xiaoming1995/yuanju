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
