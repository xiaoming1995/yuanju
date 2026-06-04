import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'
const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (p) => readFileSync(resolve(root, p), 'utf8')

test('share card renders famous_couple after the score', () => {
  const src = read('src/components/CompatibilityShareCard.tsx')
  assert.match(src, /const famousCouple = structured\?\.famous_couple/)
  assert.match(src, /compat-share-couple/)
  assert.match(src, /famousCouple\.couple/)
  assert.match(src, /famousCouple\.tagline/)
  assert.match(src, /famousCouple\.reason/)
  // 位置：在综合契合度分数之后
  const scoreIdx = src.indexOf('compat-share-score-value')
  const coupleIdx = src.indexOf('compat-share-couple')
  assert.ok(scoreIdx > -1 && coupleIdx > scoreIdx, 'famous_couple block must come after the score')
})

test('print layout renders famous_couple atop 命理解读', () => {
  const src = read('src/components/CompatibilityPrintLayout.tsx')
  assert.match(src, /structured\.famous_couple/)
  assert.match(src, /名人配对类比/)
  assert.match(src, /compat-print-couple-name/)
  // 位置：在命理解读章节标题之后、总体 summary 之前
  const headingIdx = src.indexOf('六、命理解读')
  const coupleIdx = src.indexOf('名人配对类比')
  const summaryIdx = src.indexOf('compat-print-summary')
  assert.ok(headingIdx > -1 && coupleIdx > headingIdx && coupleIdx < summaryIdx,
    'famous_couple must render after 命理解读 heading and before the summary')
})
