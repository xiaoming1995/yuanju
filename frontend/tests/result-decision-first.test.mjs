import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (path) => readFileSync(resolve(root, path), 'utf8')

test('result page has decision-first hero and action components', () => {
  const page = read('src/pages/ResultPage.tsx')
  const hero = read('src/components/result/ResultHeroSummary.tsx')
  const actionBar = read('src/components/result/ResultActionBar.tsx')

  assert.match(page, /ResultHeroSummary/)
  assert.match(page, /ResultActionBar/)
  assert.match(hero, /result-hero-summary/)
  assert.match(hero, /result-core-conclusion/)
  assert.match(actionBar, /result-action-bar/)
  assert.match(actionBar, /result-mobile-primary-action/)
})

test('result hero appears before professional detail content', () => {
  const page = read('src/pages/ResultPage.tsx')

  const hero = page.indexOf('<ResultHeroSummary')
  const detail = page.indexOf('professional-view')
  const report = page.indexOf('report-section')

  assert.ok(hero > -1, 'hero should be rendered')
  assert.ok(detail > -1, 'professional detail should remain rendered')
  assert.ok(report > -1, 'report section should remain rendered')
  assert.ok(hero < detail, 'hero should appear before professional detail')
  assert.ok(hero < report, 'hero should appear before report section')
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

test('result missing report state still offers deterministic summary and AI generation action', () => {
  const hero = read('src/components/result/ResultHeroSummary.tsx')
  const page = read('src/pages/ResultPage.tsx')

  assert.match(hero, /buildResultCoreConclusion/)
  assert.match(hero, /待生成 AI 解读/)
  assert.match(page, /id="generate-ai-report"/)
})
