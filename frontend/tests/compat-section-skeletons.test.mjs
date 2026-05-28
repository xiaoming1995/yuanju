import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (path) => readFileSync(resolve(root, path), 'utf8')

test('SectionBasicCharts exports default React component', () => {
  const src = read('src/components/compatibility/SectionBasicCharts.tsx')
  assert.match(src, /export default function SectionBasicCharts/)
  assert.match(src, /compat-section-basic-charts/)
})

test('SectionVerdict exports default React component', () => {
  const src = read('src/components/compatibility/SectionVerdict.tsx')
  assert.match(src, /export default function SectionVerdict/)
  assert.match(src, /compat-section-verdict/)
})

test('SectionDeepAnalysis exports default React component', () => {
  const src = read('src/components/compatibility/SectionDeepAnalysis.tsx')
  assert.match(src, /export default function SectionDeepAnalysis/)
  assert.match(src, /compat-section-deep-analysis/)
})

test('SectionBasicCharts renders two ParticipantSummaryCard side-by-side', () => {
  const src = read('src/components/compatibility/SectionBasicCharts.tsx')
  assert.match(src, /import ParticipantSummaryCard from '\.\/ParticipantSummaryCard'/)
  assert.match(src, /ParticipantSummaryCard/)
  assert.match(src, /self/)
  assert.match(src, /partner/)
})

test('SectionBasicCharts CSS uses 2-col grid on desktop', () => {
  const css = read('src/components/compatibility/SectionBasicCharts.css')
  assert.match(css, /@media \(min-width: 641px\)[\s\S]*compat-section-basic-charts__grid[\s\S]*grid-template-columns:\s*repeat\(2/)
})
