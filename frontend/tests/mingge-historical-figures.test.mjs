import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import test from 'node:test'
import assert from 'node:assert/strict'

const root = new URL('..', import.meta.url).pathname
const read = (path) => readFileSync(join(root, path), 'utf8')

test('ResultPage loads and renders historical references only for a Ming Ge', () => {
  const source = read('src/pages/ResultPage.tsx')
  assert.match(source, /const \[historicalFigures, setHistoricalFigures\]/)
  assert.match(source, /if \(!mingGe\) \{[\s\S]+setHistoricalFigures\(\[\]\)/)
  assert.match(source, /baziAPI\.getMingGeHistoricalFigures\(mingGe\)/)
  assert.match(source, /result\.ming_ge && historicalFigures\.length > 0/)
  assert.match(source, /古人映照/)
  assert.match(source, /古人映照用于理解格局可能的发挥方向/)
})

test('professional mode reveals only supplied source, turning point and authorized Dayun', () => {
  const source = read('src/pages/ResultPage.tsx')
  assert.match(source, /readingMode === 'professional'/)
  assert.match(source, /figure\.source_title/)
  assert.match(source, /figure\.turning_point/)
  assert.match(source, /figure\.show_dayun && figure\.dayun_period && figure\.dayun_explanation/)
  assert.match(source, /catch\(\(\) => \{[\s\S]+setHistoricalFigures\(\[\]\)/)
})

test('API client and admin route keep the curated library separate from celebrity records', () => {
  const api = read('src/lib/api.ts')
  const adminApi = read('src/lib/adminApi.ts')
  const app = read('src/App.tsx')
  const admin = read('src/pages/admin/AdminMingGeHistoricalFiguresPage.tsx')
  assert.match(api, /interface MingGeHistoricalFigure/)
  assert.match(api, /getMingGeHistoricalFigures/)
  assert.match(adminApi, /adminMingGeHistoricalFiguresAPI/)
  assert.match(app, /mingge-historical-figures/)
  assert.match(admin, /不使用名人库 AI 自动生成内容/)
  assert.doesNotMatch(admin, /generateAI/)
})

test('result CSS keeps the reference section responsive and unframed', () => {
  const css = read('src/pages/ResultPage.css')
  assert.match(css, /\.result-historical-figures \{/)
  assert.match(css, /\.result-historical-figures-list \{/)
  assert.match(css, /\.result-historical-boundary/)
  assert.match(css, /\.result-historical-figure:last-child/)
})
