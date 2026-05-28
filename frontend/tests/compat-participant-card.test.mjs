import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (path) => readFileSync(resolve(root, path), 'utf8')

test('ParticipantSummaryCard lives in components/compatibility', () => {
  const src = read('src/components/compatibility/ParticipantSummaryCard.tsx')
  assert.match(src, /export default function ParticipantSummaryCard/)
  assert.match(src, /compatibility-person-card/)
  assert.match(src, /compatibility-pillar-grid/)
  assert.match(src, /compatibility-wuxing-grid/)
})

test('page no longer defines ParticipantSummaryCard inline', () => {
  const page = read('src/pages/CompatibilityResultPage.tsx')
  assert.doesNotMatch(page, /^function ParticipantSummaryCard\(/m)
  // ParticipantSummaryCard is now used inside SectionBasicCharts, not imported directly by the page
  const basicCharts = read('src/components/compatibility/SectionBasicCharts.tsx')
  assert.match(basicCharts, /import ParticipantSummaryCard from/)
})
