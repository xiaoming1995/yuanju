import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'
const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (p) => readFileSync(resolve(root, p), 'utf8')

test('FamousCoupleCard lives inside the AI deep report, not at page top', () => {
  const deep = read('src/components/compatibility/deep-analysis/DeepReportNarrative.tsx')
  // 名人类比作为 AI 深度解读的开场，渲染在「总体判断」之前
  assert.match(deep, /import FamousCoupleCard from '..\/FamousCoupleCard'/)
  assert.match(deep, /<FamousCoupleCard famousCouple=\{structuredReport\.famous_couple\} \/>/)
  const coupleIdx = deep.indexOf('<FamousCoupleCard')
  const summaryIdx = deep.indexOf('总体判断')
  assert.ok(coupleIdx > -1 && summaryIdx > -1 && coupleIdx < summaryIdx, 'FamousCoupleCard must render before 总体判断')

  // 页面顶部不再单独挂载这张卡
  const page = read('src/pages/CompatibilityResultPage.tsx')
  assert.doesNotMatch(page, /FamousCoupleCard/)
})
