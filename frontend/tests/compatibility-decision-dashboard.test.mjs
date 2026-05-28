import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { test } from 'node:test'
import { runInNewContext } from 'node:vm'
import ts from 'typescript'

const helperPath = new URL('../src/lib/compatibilityDecision.ts', import.meta.url)
const pagePath = new URL('../src/pages/CompatibilityResultPage.tsx', import.meta.url)
const cssPath = new URL('../src/pages/CompatibilityResultPage.css', import.meta.url)

function read(path) {
  return readFileSync(path, 'utf8')
}

function loadDecisionHelpers() {
  const source = read(helperPath)
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

test('compatibility decision helper exposes required derivation API', () => {
  const source = read(helperPath)

  assert.match(source, /export function buildDecisionDashboardData/)
  assert.match(source, /export function buildDecisionStageRisks/)
  assert.match(source, /export function buildDecisionFindings/)
  assert.match(source, /recommendationLabelMap/)
  assert.match(source, /confidenceLabelMap/)
  assert.match(source, /verdictFromOverallLevel/)
  assert.match(source, /export function hasLinkedEvidence/)
  assert.match(source, /maxRisk/)
  assert.match(source, /nextAction/)
})

test('compatibility decision dashboard chooses highest stage risk as max risk', () => {
  const { buildDecisionDashboardData } = loadDecisionHelpers()
  const dashboard = buildDecisionDashboardData({
    duration: {
      overall_band: 'medium',
      summary: '',
      reasons: [],
      windows: {
        three_months: { level: 'low' },
        one_year: { level: 'medium' },
        two_years_plus: { level: 'high' },
      },
    },
    evidences: [],
    overallLevel: 'medium',
    stageRisks: [
      {
        window: 'three_months',
        risk_level: 'low',
        main_risk: '短期节奏稳定',
        trigger: '短期推进时',
        advice: '保持观察。',
        evidence_keys: [],
      },
      {
        window: 'two_years_plus',
        risk_level: 'high',
        main_risk: '长期现实承压明显',
        trigger: '长期规划落地时',
        advice: '先验证责任分工。',
        evidence_keys: [],
      },
    ],
  })

  assert.equal(dashboard.maxRisk, '长期现实承压明显')
})

test('compatibility result page renders new layout sections in correct order', () => {
  const source = read(pagePath)

  const sticky = source.indexOf('<CompatibilityStickyHeader')
  const basic = source.indexOf('<SectionBasicCharts')
  const verdict = source.indexOf('<SectionVerdict')
  const deep = source.indexOf('<SectionDeepAnalysis')
  const drawer = source.indexOf('<EvidenceDrawer')

  assert.ok(sticky > -1, 'CompatibilityStickyHeader is rendered')
  assert.ok(basic > sticky, 'SectionBasicCharts follows sticky header')
  assert.ok(verdict > basic, 'SectionVerdict follows basic charts')
  assert.ok(deep > verdict, 'SectionDeepAnalysis follows verdict')
  assert.ok(drawer > deep, 'EvidenceDrawer follows deep analysis')
})

test('compatibility result page no longer renders legacy decision dashboard inline', () => {
  const source = read(pagePath)

  assert.doesNotMatch(source, /<DecisionDashboardPanel/)
  assert.doesNotMatch(source, /<DecisionEvidenceSummary/)
  assert.doesNotMatch(source, /<ResultReadingMap/)
})

test('compatibility result page includes decision-dashboard CSS hooks in extracted components', () => {
  const actionPlan = read(new URL('../src/components/compatibility/deep-analysis/ActionPlan7d30d.tsx', import.meta.url))
  const sectionVerdict = read(new URL('../src/components/compatibility/SectionVerdict.tsx', import.meta.url))

  assert.match(actionPlan, /compatibility-stage-validation-grid/)
  assert.match(sectionVerdict, /compat-section-verdict/)
})

test('compatibility result page styles stage validation card paragraphs', () => {
  const source = read(cssPath)

  assert.match(source, /\.compatibility-stage-validation-card p\s*\{/)
  assert.match(source, /\.compatibility-stage-validation-card p span\s*\{/)
})
