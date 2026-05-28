import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (path) => readFileSync(resolve(root, path), 'utf8')

test('ScoreOverview module exports both V3 and legacy components', () => {
  const src = read('src/components/compatibility/ScoreOverview.tsx')
  assert.match(src, /export function ScoreOverviewV3/)
  assert.match(src, /export function ScoreOverview/)
  assert.match(src, /compat-score-v3/)
  assert.match(src, /compatibility-quick-score/)
})

test('page no longer defines ScoreOverviewV3 nor ScoreOverview inline', () => {
  const page = read('src/pages/CompatibilityResultPage.tsx')
  assert.doesNotMatch(page, /^function ScoreOverviewV3\(/m)
  assert.doesNotMatch(page, /^function ScoreOverview\(/m)
  // ScoreOverview pair is now used inside SectionVerdict, not imported directly by the page
  const sectionVerdict = read('src/components/compatibility/SectionVerdict.tsx')
  assert.match(sectionVerdict, /import \{ ScoreOverviewV3, ScoreOverview \} from/)
})

test('SectionVerdict imports the extracted ScoreOverview pair', () => {
  const src = read('src/components/compatibility/SectionVerdict.tsx')
  assert.match(src, /import \{ ScoreOverviewV3, ScoreOverview \} from '\.\/ScoreOverview'/)
  assert.doesNotMatch(src, /InlineScoreOverview/)
  assert.doesNotMatch(src, /throw new Error/)
})
