import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (p) => readFileSync(resolve(root, p), 'utf8')

// 生成标准六十甲子
const GAN = ['甲', '乙', '丙', '丁', '戊', '己', '庚', '辛', '壬', '癸']
const ZHI = ['子', '丑', '寅', '卯', '辰', '巳', '午', '未', '申', '酉', '戌', '亥']
const SIXTY = Array.from({ length: 60 }, (_, i) => GAN[i % 10] + ZHI[i % 12])

test('data file exposes the lookup type and getter', () => {
  const src = read('src/lib/dayPillarPortraits.ts')
  assert.match(src, /export type DayPillarPortrait/)
  assert.match(src, /export function getDayPillarPortrait\(dayGan: string, dayZhi: string\)/)
})

test('present entries have non-empty tag and text', () => {
  const src = read('src/lib/dayPillarPortraits.ts')
  // 不允许空串占位
  assert.doesNotMatch(src, /tag:\s*''/)
  assert.doesNotMatch(src, /text:\s*''/)
})

test('all 60 jiazi keys are present (completeness gate — green only after Task 4)', () => {
  const src = read('src/lib/dayPillarPortraits.ts')
  const missing = SIXTY.filter((gz) => !new RegExp(`(?:^|[^甲-癸])${gz}:`).test(src))
  assert.deepEqual(missing, [], `缺失干支: ${missing.join(' ')}`)
})

test('day-pillar layer stays decoupled from the personality engine', () => {
  const lib = read('src/lib/dayPillarPortraits.ts')
  assert.doesNotMatch(lib, /compatibilityPersonality/)
  const comp = read('src/components/compatibility/DayPillarPortrait.tsx')
  assert.doesNotMatch(comp, /compatibilityPersonality/)
  assert.match(comp, /getDayPillarPortrait/)
})

test('screen / PDF / share-image all render via the shared lookup', () => {
  assert.match(read('src/pages/CompatibilityResultPage.tsx'), /<DayPillarPortrait/)
  assert.match(read('src/components/CompatibilityPrintLayout.tsx'), /getDayPillarPortrait/)
  assert.match(read('src/components/CompatibilityShareCard.tsx'), /getDayPillarPortrait/)
})
