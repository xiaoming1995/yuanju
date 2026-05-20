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
  assert.match(css, /@media \(max-width: 768px\)[\s\S]*\.compatibility-decision-hero\s*\{[^}]*padding:\s*20px;/s)
})

test('compatibility result page defines consulting report sections', () => {
  const page = read('src/pages/CompatibilityResultPage.tsx')
  assert.match(page, /StageRiskGrid/)
  assert.match(page, /RelationshipStrategyPanel/)
  assert.match(page, /EvidenceLinkedClaims/)
})

test('compatibility result page exposes a single report generation action', () => {
  const page = read('src/pages/CompatibilityResultPage.tsx')
  const generateClicks = page.match(/onClick=\{onGenerateReport\}/g) || []
  assert.equal(generateClicks.length, 1)
  assert.match(page, /onGenerateReport=\{handleGenerateReport\}/)
})

test('compatibility result page uses decision-first consulting hierarchy', () => {
  const page = read('src/pages/CompatibilityResultPage.tsx')
  assert.match(page, /DecisionHeroPanel/)
  assert.match(page, /relationshipStageText/)
  assert.match(page, /primaryQuestionText/)
  assert.match(page, /关系背景/)
  assert.match(page, /核心矛盾/)
  assert.match(page, /下一步/)
  assert.doesNotMatch(page, /compatibility-hero-card/)
  const decisionIndex = page.indexOf('<DecisionHeroPanel')
  const stageRiskIndex = page.indexOf('<StageRiskGrid')
  const scoreIndex = page.indexOf('<ScoreOverview')
  const evidenceIndex = page.indexOf('<EvidenceLinkedClaims')
  assert.ok(decisionIndex > -1, 'decision hero should render')
  assert.ok(stageRiskIndex > decisionIndex, 'stage risks should render after decision hero')
  assert.ok(scoreIndex > stageRiskIndex, 'scores should render after stage risks')
  assert.ok(evidenceIndex > scoreIndex, 'professional evidence should render after scores')
  assert.match(page, /onGenerateReport=\{handleGenerateReport\}/)
  assert.match(page, /hasReport=\{Boolean\(detail\.latest_report\)\}/)
})

test('compatibility result page explains scores and stages as user questions', () => {
  const page = read('src/pages/CompatibilityResultPage.tsx')
  assert.match(page, /会不会互相吸引/)
  assert.match(page, /能不能长期稳定/)
  assert.match(page, /吵架后能不能修复/)
  assert.match(page, /现实条件能不能落地/)
  assert.match(page, /阶段任务/)
  assert.match(page, /风险触发/)
})

test('compatibility result page renders question-aware report focus', () => {
  const page = read('src/pages/CompatibilityResultPage.tsx')
  assert.match(page, /QuestionFocusPanel/)
  assert.match(page, /question_focus/)
  assert.match(page, /boundary_conditions/)
})
