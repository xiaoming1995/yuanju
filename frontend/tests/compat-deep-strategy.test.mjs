import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'
const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (p) => readFileSync(resolve(root, p), 'utf8')

test('RelationshipStrategy subsection exists', () => {
  const src = read('src/components/compatibility/deep-analysis/RelationshipStrategy.tsx')
  assert.match(src, /export default function RelationshipStrategy/)
  assert.match(src, /<details open/)
})

test('page no longer defines RelationshipStrategyPanel inline', () => {
  const page = read('src/pages/CompatibilityResultPage.tsx')
  assert.doesNotMatch(page, /^function RelationshipStrategyPanel\(/m)
})
