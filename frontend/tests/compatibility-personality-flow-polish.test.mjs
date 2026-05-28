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

test('compatibility result new layout renders deep analysis after verdict section', () => {
  const page = read('src/pages/CompatibilityResultPage.tsx')
  const css = read('src/pages/CompatibilityResultPage.css')
  const personalityFit = read('src/components/compatibility/deep-analysis/PersonalityFit.tsx')
  const actionPlan = read('src/components/compatibility/deep-analysis/ActionPlan7d30d.tsx')

  // new layout has no legacy inline panels
  assert.doesNotMatch(page, /<DecisionDashboardPanel/)
  assert.doesNotMatch(page, /<ResultReadingMap/)
  assert.doesNotMatch(page, /<PersonalityFitPanel/)
  assert.doesNotMatch(page, /<PersonalityValidationPlanPanel/)

  // section order: Verdict (with scores) → DeepAnalysis (with personality)
  const verdict = page.indexOf('<SectionVerdict')
  const deep = page.indexOf('<SectionDeepAnalysis')
  assert.ok(verdict > -1, 'SectionVerdict is present')
  assert.ok(deep > verdict, 'SectionDeepAnalysis follows SectionVerdict')

  // personality content lives in extracted components
  assert.match(personalityFit, /matchTypeDescription/)
  // personality-fit styles live in PersonalityFit.css after CSS split (T22)
  const personalityFitCss = read('src/components/compatibility/deep-analysis/PersonalityFit.css')
  assert.match(personalityFitCss, /compat-da-personality/)
})

test('compatibility validation groups stage risks under the validation plan in ActionPlan7d30d', () => {
  const actionPlan = read('src/components/compatibility/deep-analysis/ActionPlan7d30d.tsx')

  const validationComponent = actionPlan.indexOf('function ActionPlan7d30d')
  const stageGrid = actionPlan.indexOf('<StageRiskGrid')
  const durationSummary = actionPlan.indexOf('<DurationTaskSummary')
  const standaloneHeading = actionPlan.indexOf('接下来要验证什么')

  assert.ok(stageGrid > validationComponent, 'stage risks are rendered inside ActionPlan7d30d')
  assert.ok(durationSummary > validationComponent, 'duration summary is inside ActionPlan7d30d')
  assert.equal(standaloneHeading, -1, 'old duplicate validation heading is removed')
  assert.match(actionPlan, /阶段风险明细/)
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
