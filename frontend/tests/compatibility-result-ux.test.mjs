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

test('compatibility result page defines consulting report sections', () => {
  const page = read('src/pages/CompatibilityResultPage.tsx')
  assert.match(page, /ConsultingOverview/)
  assert.match(page, /DecisionAdvicePanel/)
  assert.match(page, /StageRiskGrid/)
  assert.match(page, /RelationshipStrategyPanel/)
  assert.match(page, /EvidenceLinkedClaims/)
})

test('compatibility result page groups professional depth evidence', () => {
  const page = read('src/pages/CompatibilityResultPage.tsx')
  const css = read('src/pages/CompatibilityResultPage.css')
  assert.match(page, /ProfessionalEvidenceGroups/)
  assert.match(page, /groupEvidenceBySource/)
  assert.match(page, /ten_god_interaction:\s*'十神互动'/)
  assert.match(page, /ganzhi_interaction:\s*'干支合冲刑害'/)
  assert.match(page, /favorable_element_support:\s*'喜忌互补'/)
  assert.match(page, /relationship_pattern:\s*'关系模式'/)
  assert.match(page, /groups\.length === 0/)
  assert.match(page, /evidence\.related_sources/)
  assert.match(css, /\.compatibility-evidence-groups/)
  assert.match(css, /\.compatibility-evidence-group-header/)
})

test('compatibility result page exposes a single report generation action', () => {
  const page = read('src/pages/CompatibilityResultPage.tsx')
  const generateClicks = page.match(/onClick=\{handleGenerateReport\}/g) || []
  assert.equal(generateClicks.length, 1)
})
