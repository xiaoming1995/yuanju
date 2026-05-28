import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'
const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (p) => readFileSync(resolve(root, p), 'utf8')

test('DeepReportNarrative subsection exists', () => {
  const src = read('src/components/compatibility/deep-analysis/DeepReportNarrative.tsx')
  assert.match(src, /export default function DeepReportNarrative/)
  assert.match(src, /<details open/)
  assert.match(src, /onGenerateReport/)
})

test('page no longer defines DeepReportPanel inline', () => {
  const page = read('src/pages/CompatibilityResultPage.tsx')
  assert.doesNotMatch(page, /^function DeepReportPanel\(/m)
})
