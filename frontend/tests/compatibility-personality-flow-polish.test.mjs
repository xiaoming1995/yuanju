import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { runInNewContext } from 'node:vm'
import test from 'node:test'
import ts from 'typescript'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (path) => readFileSync(resolve(root, path), 'utf8')

function loadPersonalityHelpers() {
  const source = read('src/lib/compatibilityPersonality.ts')
  const output = ts.transpileModule(source, {
    compilerOptions: {
      module: ts.ModuleKind.CommonJS,
      target: ts.ScriptTarget.ES2020,
    },
  }).outputText
  const exports = {}
  runInNewContext(output, { exports })
  return exports
}

test('compatibility entry exposes compact consultation progress before birth forms', () => {
  const page = read('src/pages/CompatibilityPage.tsx')
  const css = read('src/pages/CompatibilityPage.css')

  const progress = page.indexOf('compatibility-step-progress')
  const consultation = page.indexOf('compatibility-personality-consultation')
  const forms = page.indexOf('compatibility-forms')

  assert.ok(progress > -1, 'entry progress exists')
  assert.ok(progress < forms, 'progress appears before birth forms')
  assert.ok(consultation < forms, 'consultation controls appear before birth forms')
  assert.match(page, /第 1 步/)
  assert.match(page, /第 2 步/)
  assert.match(page, /第 3 步/)
  assert.match(page, /birthProfileProgressLabel/)
  assert.match(css, /compatibility-step-progress/)
  assert.match(css, /compatibility-step-item/)
  assert.match(css, /@media \(max-width: 768px\)[\s\S]*\.compatibility-step-progress/s)
})

test('personality helper exposes descriptions for match types', () => {
  const { getPersonalityMatchTypeDescription } = loadPersonalityHelpers()

  assert.equal(typeof getPersonalityMatchTypeDescription, 'function')
  assert.match(getPersonalityMatchTypeDescription('高吸引高消耗型'), /吸引|消耗|沟通/)
  assert.match(getPersonalityMatchTypeDescription('稳定互补型'), /稳定|互补|承接/)
})

test('compatibility result has reading map and self-explanatory personality section', () => {
  const page = read('src/pages/CompatibilityResultPage.tsx')
  const css = read('src/pages/CompatibilityResultPage.css')

  const dashboard = page.indexOf('<DecisionDashboardPanel')
  const map = page.indexOf('<ResultReadingMap')
  const personality = page.indexOf('<PersonalityFitPanel')
  const validation = page.indexOf('<PersonalityValidationPlanPanel')
  const score = page.indexOf('<ScoreOverview')

  assert.ok(map > dashboard, 'reading map follows decision dashboard')
  assert.ok(personality > map, 'personality section follows reading map')
  assert.ok(validation > personality, 'validation follows personality section')
  assert.ok(score > validation, 'scores remain below validation')
  assert.match(page, /id="compatibility-personality-fit"/)
  assert.match(page, /id="compatibility-conflict-validation"/)
  assert.match(page, /id="compatibility-score-evidence"/)
  assert.match(page, /id="compatibility-professional-details"/)
  assert.match(page, /compatibility-result-map/)
  assert.match(page, /matchTypeDescription/)
  assert.match(css, /compatibility-result-map/)
  assert.match(css, /compatibility-personality-fit--polished/)
})

test('compatibility validation groups stage risks under the validation plan', () => {
  const page = read('src/pages/CompatibilityResultPage.tsx')

  const validationComponent = page.indexOf('function PersonalityValidationPlanPanel')
  const stageGrid = page.indexOf('<StageRiskGrid')
  const durationSummary = page.indexOf('<DurationTaskSummary')
  const standaloneHeading = page.indexOf('接下来要验证什么')

  assert.ok(stageGrid > validationComponent, 'stage risks are rendered by/after validation panel definition')
  assert.ok(durationSummary > validationComponent, 'duration summary is grouped with validation panel')
  assert.equal(standaloneHeading, -1, 'old duplicate validation heading is removed')
  assert.match(page, /阶段风险明细/)
})

test('compatibility history prioritizes personality and de-emphasizes scores', () => {
  const page = read('src/pages/CompatibilityHistoryPage.tsx')
  const css = read('src/pages/CompatibilityHistoryPage.css')

  const personality = page.indexOf('compatibility-history-personality')
  const context = page.indexOf('compatibility-history-context')
  const scores = page.indexOf('compatibility-history-score-summary')
  const continuation = page.indexOf('compatibility-history-continuation')

  assert.ok(personality > -1, 'personality summary exists')
  assert.ok(context > personality, 'context follows personality summary')
  assert.ok(scores > context, 'scores are below context')
  assert.ok(continuation > scores, 'continuation follows secondary scores')
  assert.doesNotMatch(page, /compatibility-score-list/)
  assert.match(page, /分数参考/)
  assert.match(css, /compatibility-history-score-summary/)
  assert.match(css, /compatibility-history-continuation/)
})
