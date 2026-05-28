import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (path) => readFileSync(resolve(root, path), 'utf8')

test('claim evidence exposes the first item and previews collapsed items in EvidenceDrawer', () => {
  const evidenceDrawer = read('src/components/compatibility/EvidenceDrawer.tsx')
  // Claim card styles live in EvidenceDrawer.css after CSS split (T22)
  const css = read('src/components/compatibility/EvidenceDrawer.css')

  assert.match(evidenceDrawer, /links\.map\(\(link,\s*index\)/)
  assert.match(evidenceDrawer, /open=\{index === 0\}/)
  assert.match(evidenceDrawer, /compatibility-claim-preview/)
  assert.match(evidenceDrawer, /查看完整依据/)
  assert.match(evidenceDrawer, /收起依据/)
  assert.match(css, /\.compatibility-claim-card\[open\]/)
  assert.match(css, /\.compatibility-claim-preview/)
})

test('deep report absent state is compact actionable and owns the single generation action', () => {
  const deepReport = read('src/components/compatibility/deep-analysis/DeepReportNarrative.tsx')
  // AI card / report styles live in DeepReportNarrative.css after CSS split (T22)
  const css = read('src/components/compatibility/deep-analysis/DeepReportNarrative.css')

  // DeepReportNarrative owns the single onClick for report generation
  const generateClicks = deepReport.match(/onClick=\{onGenerateReport\}/g) || []
  assert.equal(generateClicks.length, 1)
  assert.match(deepReport, /compatibility-ai-card--empty/)
  assert.match(deepReport, /AI 深度解读会补充/)
  assert.match(deepReport, /reportLoading \? '生成中' : '生成深度解读'/)
  assert.doesNotMatch(deepReport, /<p className="compatibility-report-empty">尚未生成深度解读。<\/p>/)
  assert.match(css, /\.compatibility-ai-card--empty/)
  assert.match(css, /\.compatibility-report-action/)
  assert.match(css, /\.compatibility-report-state/)
})

test('deep report generated states keep structured and raw reading hierarchy in DeepReportNarrative', () => {
  const deepReport = read('src/components/compatibility/deep-analysis/DeepReportNarrative.tsx')
  // Report section/raw styles live in DeepReportNarrative.css after CSS split (T22)
  const css = read('src/components/compatibility/deep-analysis/DeepReportNarrative.css')

  assert.match(deepReport, /structuredReport \?/)
  assert.match(deepReport, /compatibility-report-section/)
  assert.match(deepReport, /compatibility-report-raw/)
  assert.match(css, /\.compatibility-report-section/)
  assert.match(css, /\.compatibility-report-raw/)
})

test('professional details summary communicates available data in EvidenceDrawer', () => {
  const evidenceDrawer = read('src/components/compatibility/EvidenceDrawer.tsx')
  // Evidence drawer styles live in EvidenceDrawer.css after CSS split (T22)
  const css = read('src/components/compatibility/EvidenceDrawer.css')

  // The evidence drawer uses a <details open> pattern for expandable sections
  assert.match(evidenceDrawer, /compat-evidence-drawer/)
  assert.match(evidenceDrawer, /结构化证据组/)
  assert.match(css, /\.compat-evidence-drawer/)
  assert.match(css, /\.compat-evidence-drawer__body/)
})
