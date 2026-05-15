import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (path) => readFileSync(resolve(root, path), 'utf8')

test('result page defines ten god relation data and legacy fallback derivation', () => {
  const page = read('src/pages/ResultPage.tsx')
  assert.match(page, /interface TenGodRelationMatrix/)
  assert.match(page, /ten_god_relation\?: TenGodRelationMatrix/)
  assert.match(page, /buildTenGodRelationMatrix/)
  assert.match(page, /日主 \/ 日元/)
  assert.match(page, /命盘的参照点/)
})

test('result page renders day-master and heavenly-stem relation cards', () => {
  const page = read('src/pages/ResultPage.tsx')
  assert.match(page, /ten-god-relation-section/)
  assert.match(page, /命主十神关系/)
  assert.match(page, /命主日元/)
  assert.match(page, /ten-god-stem-grid/)
  assert.match(page, /relation\.heavenly_stems\.map/)
})

test('basic bazi chart appears before ten god explanation module', () => {
  const page = read('src/pages/ResultPage.tsx')
  const chartIndex = page.indexOf('基本排盘')
  const relationIndex = page.indexOf('命主十神关系')
  assert.ok(chartIndex > -1, 'basic chart heading should exist')
  assert.ok(relationIndex > -1, 'ten god relation heading should exist')
  assert.ok(chartIndex < relationIndex, '基本排盘 should appear before 命主十神关系')
})

test('result page renders hidden-stem relation cards with concise explanations', () => {
  const page = read('src/pages/ResultPage.tsx')
  assert.match(page, /地支藏干关系/)
  assert.match(page, /ten-god-hidden-grid/)
  assert.match(page, /group\.items\.map/)
  assert.match(page, /tenGodSummary/)
  assert.match(page, /七杀.*外部压力/)
})

test('ten god relation css supports mobile cards without horizontal overflow', () => {
  const css = read('src/pages/ResultPage.css')
  assert.match(css, /\.ten-god-relation-section/)
  assert.match(css, /\.ten-god-stem-grid/)
  assert.match(css, /\.ten-god-hidden-grid/)
  assert.match(css, /@media \(max-width: 640px\)[\s\S]*\.ten-god-stem-grid\s*\{[^}]*grid-template-columns:\s*1fr;/s)
  assert.match(css, /@media \(max-width: 640px\)[\s\S]*\.ten-god-hidden-grid\s*\{[^}]*grid-template-columns:\s*1fr;/s)
})
