import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (path) => readFileSync(resolve(root, path), 'utf8')

test('chart history page renders as an archive with switcher stats and card actions', () => {
  const page = read('src/pages/HistoryPage.tsx')
  assert.match(page, /history-archive-hero/)
  assert.match(page, /archive-switcher/)
  assert.match(page, /history-stat-grid/)
  assert.match(page, /history-record-card/)
  assert.match(page, /查看命盘/)
  assert.match(page, /合盘档案/)
})

test('chart history css includes archive layout and mobile bottom safe area', () => {
  const css = read('src/pages/HistoryPage.css')
  assert.match(css, /\.history-page\.page\s*\{[^}]*padding-bottom:\s*calc\(140px \+ env\(safe-area-inset-bottom\)\)\s*!important;/s)
  assert.match(css, /\.archive-switcher/)
  assert.match(css, /\.history-stat-grid/)
  assert.match(css, /\.history-record-card/)
  assert.match(css, /@media \(max-width: 768px\)[\s\S]*\.history-stat-grid\s*\{[^}]*grid-template-columns:\s*1fr;/s)
})

test('compatibility history uses archive styling without inline layout-heavy cards', () => {
  const page = read('src/pages/CompatibilityHistoryPage.tsx')
  assert.match(page, /CompatibilityHistoryPage.css/)
  assert.match(page, /compatibility-history-page/)
  assert.match(page, /archive-switcher/)
  assert.match(page, /compatibility-history-card/)
  assert.match(page, /查看合盘/)
  assert.match(page, /命盘档案/)
  assert.doesNotMatch(page, /style=\{\{/)
})

test('compatibility history css supports archive cards and mobile safe area', () => {
  const css = read('src/pages/CompatibilityHistoryPage.css')
  assert.match(css, /\.compatibility-history-page\.page\s*\{[^}]*padding-bottom:\s*calc\(140px \+ env\(safe-area-inset-bottom\)\)\s*!important;/s)
  assert.match(css, /\.compatibility-history-card/)
  assert.match(css, /\.compatibility-score-list/)
  assert.match(css, /@media \(max-width: 768px\)[\s\S]*\.compatibility-score-list\s*\{[^}]*grid-template-columns:\s*1fr;/s)
})
