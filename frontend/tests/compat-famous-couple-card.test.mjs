import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'
const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (p) => readFileSync(resolve(root, p), 'utf8')

test('FamousCoupleCard component exists with three states', () => {
  const src = read('src/components/compatibility/FamousCoupleCard.tsx')
  assert.match(src, /export default function FamousCoupleCard/)
  assert.match(src, /famous_couple/)
  assert.match(src, /\.couple/)
  assert.match(src, /\.tagline/)
  assert.match(src, /\.reason/)
  assert.match(src, /揭晓你们的名人配对/)
  assert.match(src, /onGenerateReport/)
})

test('FamousCoupleCard hides on legacy report without famous_couple', () => {
  const src = read('src/components/compatibility/FamousCoupleCard.tsx')
  assert.match(src, /return null/)
})
