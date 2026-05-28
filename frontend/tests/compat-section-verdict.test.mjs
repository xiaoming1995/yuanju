import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (path) => readFileSync(resolve(root, path), 'utf8')

test('SectionVerdict accepts dashboard / findings / scores props and renders both v3 and legacy branches', () => {
  const src = read('src/components/compatibility/SectionVerdict.tsx')
  assert.match(src, /export default function SectionVerdict/)
  assert.match(src, /dashboard/)
  assert.match(src, /findings/)
  assert.match(src, /isV3/)
  assert.match(src, /v3Scores/)
  assert.match(src, /legacyScores/)
  assert.match(src, /ScoreOverviewV3/)
  assert.match(src, /ScoreOverview/)
  assert.match(src, /compat-section-verdict/)
})

test('SectionVerdict CSS uses 2-col layout on desktop', () => {
  const css = read('src/components/compatibility/SectionVerdict.css')
  assert.match(css, /@media \(min-width: 1024px\)[\s\S]*compat-section-verdict__columns[\s\S]*grid-template-columns/)
})
