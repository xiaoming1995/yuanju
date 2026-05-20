import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { test } from 'node:test'

const helperPath = new URL('../src/lib/compatibilityDecision.ts', import.meta.url)
const pagePath = new URL('../src/pages/CompatibilityResultPage.tsx', import.meta.url)

function read(path) {
  return readFileSync(path, 'utf8')
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

test('compatibility result page renders decision-dashboard sections before scores and AI report', () => {
  const source = read(pagePath)

  const dashboard = source.indexOf('DecisionDashboardPanel')
  const evidence = source.indexOf('DecisionEvidenceSummary')
  const stages = source.indexOf('StageRiskGrid')
  const score = source.indexOf('ScoreOverview')
  const ai = source.indexOf('compatibility-ai-card')
  const professional = source.indexOf('compatibility-professional-details')

  assert.ok(dashboard > -1, 'decision dashboard component is rendered')
  assert.ok(evidence > dashboard, 'decision evidence summary follows dashboard')
  assert.ok(stages > evidence, 'stage validation follows decision evidence')
  assert.ok(score > stages, 'score overview follows decision sections')
  assert.ok(ai > score, 'AI deep reading follows score overview')
  assert.ok(professional > ai, 'professional details follow AI reading')
  assert.match(source, /生成深度解读/)
})

test('compatibility result page includes decision-dashboard CSS hooks', () => {
  const source = read(pagePath)

  assert.match(source, /compatibility-decision-dashboard/)
  assert.match(source, /compatibility-decision-metric-grid/)
  assert.match(source, /compatibility-decision-evidence-summary/)
  assert.match(source, /compatibility-stage-validation-grid/)
})
