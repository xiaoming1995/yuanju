import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (path) => readFileSync(resolve(root, path), 'utf8')

test('claim evidence exposes the first item and previews collapsed items', () => {
  const page = read('src/pages/CompatibilityResultPage.tsx')
  const css = read('src/pages/CompatibilityResultPage.css')

  assert.match(page, /links\.map\(\(link,\s*index\)/)
  assert.match(page, /open=\{index === 0\}/)
  assert.match(page, /compatibility-claim-preview/)
  assert.match(page, /查看完整依据/)
  assert.match(page, /收起依据/)
  assert.match(css, /\.compatibility-claim-card\[open\]/)
  assert.match(css, /\.compatibility-claim-preview/)
})

test('deep report absent state is compact actionable and owns the single generation action', () => {
  const page = read('src/pages/CompatibilityResultPage.tsx')
  const css = read('src/pages/CompatibilityResultPage.css')

  const generateClicks = page.match(/onClick=\{onGenerateReport\}/g) || []
  assert.equal(generateClicks.length, 1)
  assert.match(page, /DeepReportPanel/)
  assert.match(page, /compatibility-ai-card--empty/)
  assert.match(page, /AI 深度解读会补充/)
  assert.match(page, /reportLoading \? '生成中' : '生成深度解读'/)
  assert.doesNotMatch(page, /<p className="compatibility-report-empty">尚未生成深度解读。<\/p>/)
  assert.match(css, /\.compatibility-ai-card--empty/)
  assert.match(css, /\.compatibility-report-action/)
  assert.match(css, /\.compatibility-report-state/)
})

test('deep report generated states keep structured and raw reading hierarchy', () => {
  const page = read('src/pages/CompatibilityResultPage.tsx')
  const css = read('src/pages/CompatibilityResultPage.css')

  assert.match(page, /structuredReport \?/)
  assert.match(page, /compatibility-report-section/)
  assert.match(page, /compatibility-report-raw/)
  assert.match(css, /\.compatibility-report-section/)
  assert.match(css, /\.compatibility-report-raw/)
})

test('professional details summary communicates available data before expansion', () => {
  const page = read('src/pages/CompatibilityResultPage.tsx')
  const css = read('src/pages/CompatibilityResultPage.css')

  assert.match(page, /<details className="compatibility-professional-details" id="compatibility-professional-details" open>/)
  assert.match(page, /compatibility-professional-summary-grid/)
  assert.match(page, /双方四柱/)
  assert.match(page, /五行摘要/)
  assert.match(page, /结构化依据/)
  assert.match(css, /\.compatibility-professional-summary-grid/)
  assert.match(css, /\.compatibility-professional-details\[open\]/)
})
