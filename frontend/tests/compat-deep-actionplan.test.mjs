import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'
const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (p) => readFileSync(resolve(root, p), 'utf8')

test('ActionPlan7d30d subsection exists with details-open', () => {
  const src = read('src/components/compatibility/deep-analysis/ActionPlan7d30d.tsx')
  assert.match(src, /export default function ActionPlan7d30d/)
  assert.match(src, /<details open/)
  assert.match(src, /compat-da-actionplan/)
})

test('page no longer defines PersonalityValidationPlanPanel / StageRiskGrid / DurationTaskSummary inline', () => {
  const page = read('src/pages/CompatibilityResultPage.tsx')
  assert.doesNotMatch(page, /^function PersonalityValidationPlanPanel\(/m)
  assert.doesNotMatch(page, /^function StageRiskGrid\(/m)
  assert.doesNotMatch(page, /^function DurationTaskSummary\(/m)
})
