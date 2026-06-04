import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (path) => readFileSync(resolve(root, path), 'utf8')

test('compatibility result page renders new layout sections', () => {
  const page = read('src/pages/CompatibilityResultPage.tsx')
  const scoreModule = read('src/components/compatibility/ScoreOverview.tsx')
  const sectionVerdict = read('src/components/compatibility/SectionVerdict.tsx')
  const nextSteps = read('src/components/compatibility/deep-analysis/NextStepsAndAvoid.tsx')

  assert.match(page, /compatibility-result-page/)
  assert.match(scoreModule, /compatibility-quick-score/)
  // new layout uses SectionVerdict which contains the verdict and scores
  assert.match(sectionVerdict, /compat-section-verdict/)
  // next steps and avoid live in extracted component
  assert.match(nextSteps, /最大风险|下一步验证/)
})

test('compatibility result css uses mobile score rows and bottom nav safe area', () => {
  const css = read('src/pages/CompatibilityResultPage.css')
  // Quick-score styles live in ScoreOverview.css after CSS split (T22)
  const scoreCss = read('src/components/compatibility/ScoreOverview.css')
  assert.match(css, /\.compatibility-result-page\s*\{[^}]*padding-bottom:\s*calc\(140px \+ env\(safe-area-inset-bottom\)\)/s)
  assert.match(scoreCss, /\.compatibility-quick-score-bar/)
  assert.match(scoreCss, /\.compatibility-quick-score-fill/)
})

test('compatibility result page defines consulting report sections in extracted components', () => {
  const actionPlan = read('src/components/compatibility/deep-analysis/ActionPlan7d30d.tsx')
  const relationshipStrategy = read('src/components/compatibility/deep-analysis/RelationshipStrategy.tsx')
  const evidenceDrawer = read('src/components/compatibility/EvidenceDrawer.tsx')

  assert.match(actionPlan, /StageRiskGrid/)
  assert.match(relationshipStrategy, /RelationshipStrategy/)
  assert.match(evidenceDrawer, /EvidenceLinkedClaims/)
})

test('compatibility result page groups professional depth evidence in EvidenceDrawer', () => {
  const evidenceDrawer = read('src/components/compatibility/EvidenceDrawer.tsx')
  const css = read('src/pages/CompatibilityResultPage.css')

  assert.match(evidenceDrawer, /ProfessionalEvidenceGroups/)
  assert.match(evidenceDrawer, /groupEvidenceBySource/)
  assert.match(evidenceDrawer, /ten_god_interaction:\s*'十神互动'/)
  assert.match(evidenceDrawer, /ganzhi_interaction:\s*'干支合冲刑害'/)
  assert.match(evidenceDrawer, /favorable_element_support:\s*'喜忌互补'/)
  assert.match(evidenceDrawer, /relationship_pattern:\s*'关系模式'/)
  assert.match(evidenceDrawer, /groups\.length === 0/)
  assert.match(evidenceDrawer, /evidence\.related_sources/)
  // Evidence group styles live in EvidenceDrawer.css after CSS split (T22)
  const evidenceCss = read('src/components/compatibility/EvidenceDrawer.css')
  assert.match(evidenceCss, /\.compatibility-evidence-groups/)
  assert.match(evidenceCss, /\.compatibility-evidence-group-header/)
})

test('compatibility result page wires a single report generation handler to DeepReportNarrative', () => {
  const page = read('src/pages/CompatibilityResultPage.tsx')
  const deepReport = read('src/components/compatibility/deep-analysis/DeepReportNarrative.tsx')

  // page reuses the single handleGenerateReport handler in two generate entrypoints:
  // the standalone DeepReportNarrative and the top FamousCoupleCard teaser
  const handlerWires = page.match(/onGenerateReport=\{handleGenerateReport\}/g) || []
  assert.equal(handlerWires.length, 2)
  const generateClicks = deepReport.match(/onClick=\{onGenerateReport\}/g) || []
  assert.equal(generateClicks.length, 1)
})

test('compatibility result page uses decision-first consulting hierarchy', () => {
  const page = read('src/pages/CompatibilityResultPage.tsx')

  // legacy inline components are gone
  assert.doesNotMatch(page, /<DecisionDashboardPanel/)
  assert.doesNotMatch(page, /compatibility-hero-card/)

  // new layout sections are present
  assert.match(page, /<SectionVerdict/)
  assert.match(page, /<SectionDeepAnalysis/)

  // section order: StickyHeader → BasicCharts → Verdict → DeepAnalysis → EvidenceDrawer → DeepReportNarrative
  const sticky = page.indexOf('<CompatibilityStickyHeader')
  const verdict = page.indexOf('<SectionVerdict')
  const deep = page.indexOf('<SectionDeepAnalysis')
  const drawer = page.indexOf('<EvidenceDrawer')
  const report = page.indexOf('<DeepReportNarrative')

  assert.ok(sticky > -1, 'sticky header should render')
  assert.ok(verdict > sticky, 'verdict should render after sticky header')
  assert.ok(deep > verdict, 'deep analysis should render after verdict')
  assert.ok(drawer > deep, 'evidence drawer should render after deep analysis')
  assert.ok(report > drawer, 'AI 深度解读 renders last, after evidence drawer')

  // report props now passed as JSX attributes on the standalone DeepReportNarrative
  assert.match(page, /onGenerateReport=\{handleGenerateReport\}/)
  assert.match(page, /hasReport=\{Boolean\(detail\.latest_report\)\}/)
})

test('compatibility result page explains scores and stages as user questions in extracted components', () => {
  const scoreModule = read('src/components/compatibility/ScoreOverview.tsx')
  const evidenceDrawer = read('src/components/compatibility/EvidenceDrawer.tsx')
  const actionPlan = read('src/components/compatibility/deep-analysis/ActionPlan7d30d.tsx')

  assert.match(scoreModule, /会不会互相吸引/)
  assert.match(scoreModule, /能不能长期稳定/)
  assert.match(evidenceDrawer, /吵架后能不能修复/)
  assert.match(scoreModule, /现实条件能不能落地/)
  assert.match(actionPlan, /要验证什么/)
  assert.match(actionPlan, /触发场景/)
})

test('compatibility result page renders question-aware report focus in DeepReportNarrative', () => {
  const deepReport = read('src/components/compatibility/deep-analysis/DeepReportNarrative.tsx')

  assert.match(deepReport, /QuestionFocusPanel/)
  assert.match(deepReport, /question_focus/)
  assert.match(deepReport, /boundary_conditions/)
})

test('all <details> elements in compatibility result page are open by default', () => {
  const page = read('src/pages/CompatibilityResultPage.tsx')
  const opens = page.match(/<details\b[^>]*>/g) || []
  for (const tag of opens) {
    assert.match(tag, /\bopen\b/, `expected ${tag} to include the open attribute`)
  }
})
