import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'
const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (p) => readFileSync(resolve(root, p), 'utf8')

test('FamousCoupleCard renders the analogy fields', () => {
  const src = read('src/components/compatibility/FamousCoupleCard.tsx')
  assert.match(src, /export default function FamousCoupleCard/)
  assert.match(src, /famousCouple/)
  assert.match(src, /\.couple/)
  assert.match(src, /\.tagline/)
  assert.match(src, /\.reason/)
})

test('FamousCoupleCard renders nothing when famous_couple is absent', () => {
  const src = read('src/components/compatibility/FamousCoupleCard.tsx')
  assert.match(src, /if \(!famousCouple\) return null/)
})
