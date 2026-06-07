import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (path) => readFileSync(resolve(root, path), 'utf8')

test('result page top uses the simple pillar header, not the oversized hero card', () => {
  const page = read('src/pages/ResultPage.tsx')

  assert.match(page, /result-header/)
  assert.match(page, /result-pillars/)
  assert.match(page, /result-tags/)
  // The UX-refactor hero/action-bar were reverted; guard against reintroduction.
  assert.doesNotMatch(page, /ResultHeroSummary/)
  assert.doesNotMatch(page, /ResultActionBar/)
})

test('result page header appears before professional detail and report sections', () => {
  const page = read('src/pages/ResultPage.tsx')

  const header = page.indexOf('result-header')
  const detail = page.indexOf('professional-view')
  const report = page.indexOf('report-section')

  assert.ok(header > -1, 'header should be rendered')
  assert.ok(detail > -1, 'professional detail should remain rendered')
  assert.ok(report > -1, 'report section should remain rendered')
  assert.ok(header < detail, 'header should appear before professional detail')
  assert.ok(header < report, 'header should appear before report section')
})

test('result page exposes stable segmented navigation targets', () => {
  const page = read('src/pages/ResultPage.tsx')
  const css = read('src/pages/ResultPage.css')

  for (const id of [
    'result-section-overview',
    'result-section-chart',
    'result-section-yongshen',
    'result-section-dayun',
    'result-section-ai',
  ]) {
    assert.match(page, new RegExp(`id="${id}"`), `missing ${id}`)
  }

  assert.match(page, /SegmentedTabs/)
  assert.match(css, /\.result-segment-nav/)
  assert.match(css, /overflow-x:\s*auto/)
})

test('result missing report state still offers AI generation action', () => {
  const page = read('src/pages/ResultPage.tsx')
  assert.match(page, /id="generate-ai-report"/)
})
