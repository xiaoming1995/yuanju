import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (path) => readFileSync(resolve(root, path), 'utf8')

test('gongjia panel presents virtual branches as hidden signals', () => {
  const panel = read('src/components/GongJiaPanel.tsx')
  assert.match(panel, /原局夹拱/)
  assert.match(panel, /暗藏虚支/)
  assert.match(panel, /不改原局五行与用神/)
  assert.match(panel, /拱神煞/)
})

test('result page wires gongjia panel without changing four pillars', () => {
  const page = read('src/pages/ResultPage.tsx')
  assert.match(page, /import GongJiaPanel,\s*\{\s*type GongJiaItem\s*\}/)
  assert.match(page, /gong_jia\?: GongJiaItem\[\]/)
  assert.match(page, /<GongJiaPanel[\s\S]*items=\{result\.gong_jia \|\| \[\]\}/)

  const pillarsMatch = page.match(/const pillars = \[[\s\S]*?\n  \]/)
  assert.ok(pillarsMatch, 'pillars array should be defined')
  assert.doesNotMatch(pillarsMatch[0], /gong_jia/, 'gong_jia must not be rendered as a fifth pillar')
})

test('gongjia shensha annotations use semantic clickable tags', () => {
  const panel = read('src/components/GongJiaPanel.tsx')
  assert.match(panel, /<button[\s\S]*type="button"/)
  assert.doesNotMatch(panel, /<span\b[^>]*\bonClick=/, 'clickable shensha tags should not be spans')
})

test('gongjia panel owns minimal shensha tag styling', () => {
  const css = read('src/components/GongJiaPanel.css')
  assert.match(css, /\.shensha-tag\s*\{/)
})
