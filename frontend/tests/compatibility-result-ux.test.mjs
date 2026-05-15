import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (path) => readFileSync(resolve(root, path), 'utf8')

test('compatibility result page exposes conclusion-first mobile sections', () => {
  const page = read('src/pages/CompatibilityResultPage.tsx')
  assert.match(page, /compatibility-result-page/)
  assert.match(page, /compatibility-quick-score/)
  assert.match(page, /compatibility-insight-grid/)
  assert.match(page, /compatibility-professional-details/)
  assert.match(page, /关键风险/)
  assert.match(page, /行动建议/)
})

test('compatibility result css uses mobile score rows and bottom nav safe area', () => {
  const css = read('src/pages/CompatibilityResultPage.css')
  assert.match(css, /\.compatibility-result-page\s*\{[^}]*padding-bottom:\s*calc\(140px \+ env\(safe-area-inset-bottom\)\)/s)
  assert.match(css, /\.compatibility-quick-score-bar/)
  assert.match(css, /\.compatibility-quick-score-fill/)
  assert.match(css, /@media \(max-width: 768px\)[\s\S]*\.compatibility-professional-details\s*\{[^}]*margin-top:/s)
  assert.match(css, /@media \(max-width: 768px\)[\s\S]*\.compatibility-hero-card\s*\{[^}]*padding:\s*20px;/s)
})
