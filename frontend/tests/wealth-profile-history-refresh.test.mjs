import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import test from 'node:test'
import assert from 'node:assert/strict'

const root = new URL('..', import.meta.url).pathname
const resultPage = readFileSync(join(root, 'src/pages/ResultPage.tsx'), 'utf8')
const wealthStyles = readFileSync(join(root, 'src/pages/WealthProfile.css'), 'utf8')

test('historical result pages refresh saved state for wealth-profile backfill', () => {
  assert.match(resultPage, /历史详情始终以服务端结果为准/)
  assert.match(resultPage, /if \(id\) \{\s*baziAPI\.getHistoryDetail\(id\)/)
  assert.doesNotMatch(resultPage, /if \(id && !result\) \{\s*baziAPI\.getHistoryDetail\(id\)/)
})

test('wealth profile has a compact card and professional evidence entry', () => {
  assert.match(resultPage, /interface WealthProfile/)
  assert.match(resultPage, /wealth_profile\?: WealthProfile/)
  assert.match(resultPage, /className="result-wealth-card result-summary-card result-summary-card--wealth"/)
  assert.match(resultPage, /查看财富依据/)
  assert.match(resultPage, /overviewModal === 'wealth-evidence'/)
  assert.match(wealthStyles, /\.result-summary-card--wealth/)
  assert.match(wealthStyles, /\.result-wealth-current-hint/)
  assert.match(wealthStyles, /\.result-wealth-risk-list/)
})
