import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'
const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (p) => readFileSync(resolve(root, p), 'utf8')

test('PersonalityFit renders as a top-level SECTION', () => {
  const src = read('src/components/compatibility/deep-analysis/PersonalityFit.tsx')
  assert.match(src, /export default function PersonalityFit/)
  assert.match(src, /<section/)
  assert.match(src, /compat-da-personality/)
})

test('page no longer defines PersonalityFitPanel inline', () => {
  const page = read('src/pages/CompatibilityResultPage.tsx')
  assert.doesNotMatch(page, /^function PersonalityFitPanel\(/m)
})
