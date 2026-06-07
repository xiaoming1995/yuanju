import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (path) => readFileSync(resolve(root, path), 'utf8')

test('pillars input form renders four ganzhi dropdowns and gender', () => {
  const form = read('src/components/PillarsInputForm.tsx')
  assert.match(form, /年柱/)
  assert.match(form, /月柱/)
  assert.match(form, /日柱/)
  assert.match(form, /时柱/)
  assert.match(form, /JIAZI/)
  assert.match(form, /男命/)
})

test('JIAZI builds 60 sexagenary combinations', () => {
  const form = read('src/components/PillarsInputForm.tsx')
  assert.match(form, /length:\s*60/)
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
