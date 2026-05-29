import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'
const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (p) => readFileSync(resolve(root, p), 'utf8')

test('personality is rendered by PersonalityComparison inside the AI report', () => {
  const src = read('src/components/compatibility/deep-analysis/PersonalityComparison.tsx')
  assert.match(src, /export default function PersonalityComparison/)
  assert.match(src, /comparison\.self/)
  assert.match(src, /comparison\.partner/)
  // no standalone personality SECTION mounted on the result page anymore
  const page = read('src/pages/CompatibilityResultPage.tsx')
  assert.doesNotMatch(page, /PersonalityFit/)
})

test('page no longer defines PersonalityFitPanel inline', () => {
  const page = read('src/pages/CompatibilityResultPage.tsx')
  assert.doesNotMatch(page, /^function PersonalityFitPanel\(/m)
})
