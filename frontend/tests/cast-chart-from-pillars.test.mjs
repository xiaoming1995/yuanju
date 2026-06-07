import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (path) => readFileSync(resolve(root, path), 'utf8')

test('pillars input form renders per-character gan/zhi selectors and gender', () => {
  const form = read('src/components/PillarsInputForm.tsx')
  assert.match(form, /年柱/)
  assert.match(form, /月柱/)
  assert.match(form, /日柱/)
  assert.match(form, /时柱/)
  assert.match(form, /天干/)
  assert.match(form, /地支/)
  assert.match(form, /export const GAN/)
  assert.match(form, /export const ZHI/)
  assert.match(form, /男命/)
})

test('pillars form supports typing the eight characters directly', () => {
  const form = read('src/components/PillarsInputForm.tsx')
  assert.match(form, /parseEightChars/)
  assert.match(form, /快速填入/)
  assert.match(form, /length !== 8/)
})

test('homepage offers birth/pillars input mode toggle', () => {
  const home = read('src/pages/HomePage.tsx')
  assert.match(home, /inputMode/)
  assert.match(home, /按生辰/)
  assert.match(home, /按八字/)
  assert.match(home, /PillarsInputForm/)
})

test('homepage resolves pillars then calculates, handling 0 / many candidates', () => {
  const home = read('src/pages/HomePage.tsx')
  assert.match(home, /resolvePillars/)
  assert.match(home, /找不到对应的真实日期/)
  assert.match(home, /candidates/)
  assert.match(home, /ref_age|参考年龄|岁/)
})

test('homepage gives a zi-hour-specific hint when no candidate matches', () => {
  const home = read('src/pages/HomePage.tsx')
  assert.match(home, /endsWith\('子'\)/)
  assert.match(home, /早\/晚子时|早子时|晚子时/)
})
