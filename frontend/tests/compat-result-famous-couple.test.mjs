import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'
const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (p) => readFileSync(resolve(root, p), 'utf8')

test('result page mounts FamousCoupleCard near the top', () => {
  const src = read('src/pages/CompatibilityResultPage.tsx')
  assert.match(src, /import FamousCoupleCard from '..\/components\/compatibility\/FamousCoupleCard'/)
  assert.match(src, /<FamousCoupleCard/)
  assert.match(src, /famousCouple=\{structuredReport\?\.famous_couple\}/)
  assert.match(src, /onGenerateReport=\{handleGenerateReport\}/)
  // 顶部：必须在 SectionVerdict 之前出现
  const coupleIdx = src.indexOf('<FamousCoupleCard')
  const verdictIdx = src.indexOf('<SectionVerdict')
  assert.ok(coupleIdx > -1 && verdictIdx > -1 && coupleIdx < verdictIdx, 'FamousCoupleCard must render before SectionVerdict')
})
