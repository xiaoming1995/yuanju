import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'
const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (p) => readFileSync(resolve(root, p), 'utf8')

test('NextStepsAndAvoid subsection exists, shows nextAction/avoid/coreContradiction', () => {
  const src = read('src/components/compatibility/deep-analysis/NextStepsAndAvoid.tsx')
  assert.match(src, /export default function NextStepsAndAvoid/)
  assert.match(src, /<details open/)
  assert.match(src, /nextAction/)
  assert.match(src, /avoid/)
  assert.match(src, /summary/)
})
