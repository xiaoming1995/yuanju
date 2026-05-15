import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import test from 'node:test'
import assert from 'node:assert/strict'

const root = new URL('..', import.meta.url).pathname

function read(path) {
  return readFileSync(join(root, path), 'utf8')
}

test('result page presents the basic chart as the primary panel before relation and structure sections', () => {
  const page = read('src/pages/ResultPage.tsx')

  assert.match(page, /bazi-primary-panel/)
  assert.match(page, /bazi-data-grid--primary/)

  const chart = page.indexOf('基本排盘')
  const relation = page.indexOf('命主十神关系')
  const structure = page.indexOf('命局结构')

  assert.ok(chart > -1, 'basic chart heading should exist')
  assert.ok(relation > -1, 'ten god relation heading should exist')
  assert.ok(structure > -1, 'structure heading should exist')
  assert.ok(chart < relation, 'basic chart should be shown before ten god relation')
  assert.ok(relation < structure, 'ten god relation should be shown before structure summary')
})

test('result page groups wuxing, yongshen, and tiaohou into one structure section', () => {
  const page = read('src/pages/ResultPage.tsx')

  assert.match(page, /result-structure-section/)
  assert.match(page, /result-structure-grid/)
  assert.match(page, /structure-card--wuxing/)
  assert.match(page, /structure-card--yongshen/)
  assert.match(page, /structure-card--tiaohou/)
  assert.doesNotMatch(page, /className="wuxing-section card"/)
  assert.doesNotMatch(page, /className="yongshen-section">\s*<h2/)
})

test('result page css includes calmer primary chart and mobile structure layout', () => {
  const css = read('src/pages/ResultPage.css')

  assert.match(css, /\.professional-view\s*\{[\s\S]*grid-template-columns:\s*minmax\(0,\s*1fr\);[\s\S]*gap:\s*24px;/)
  assert.match(css, /\.professional-view\s*>\s*\*\s*\{[\s\S]*min-width:\s*0;/)
  assert.match(css, /\.bazi-primary-panel/)
  assert.match(css, /\.bazi-data-grid--primary/)
  assert.match(css, /\.result-structure-section/)
  assert.match(css, /\.result-structure-grid/)
  assert.match(css, /@media \(max-width: 640px\)\s*\{[\s\S]*\.result-structure-grid\s*\{[\s\S]*grid-template-columns:\s*1fr;/)
})

test('result page highlights the day pillar without a full-column tinted fill', () => {
  const css = read('src/pages/ResultPage.css')

  assert.match(css, /\.bazi-data-grid--primary \.grid-cell\.is-day-pillar-cell/)
  assert.doesNotMatch(css, /\.bazi-data-grid--primary \.grid-cell:nth-child\(5n \+ 4\)\s*\{[\s\S]*background:/)
})

test('hidden stems use their own five-element color mapping instead of placeholders', () => {
  const page = read('src/pages/ResultPage.tsx')

  assert.doesNotMatch(page, /WUXING_MAP\['TODO'\]/)
  assert.match(page, /GAN_WUXING\[g\]/)
})
