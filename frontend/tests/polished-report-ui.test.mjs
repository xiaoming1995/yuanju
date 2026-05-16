import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import test from 'node:test'
import assert from 'node:assert/strict'

const root = new URL('..', import.meta.url).pathname
const read = (p) => readFileSync(join(root, p), 'utf8')

test('PolishedPanel 三态：empty / has-input / has-report 都渲染', () => {
  const c = read('src/components/PolishedPanel.tsx')
  assert.match(c, /className="polished-panel"/)
  assert.match(c, /polished-empty-state/)
  assert.match(c, /polished-input-area/)
  assert.match(c, /polished-content-area/)
})

test('PolishedPanel 包含 user_situation 输入 + 字数提示 + 提交按钮', () => {
  const c = read('src/components/PolishedPanel.tsx')
  assert.match(c, /<textarea/)
  assert.match(c, /maxLength=\{?300/)
  assert.match(c, /生成润色版|重新润色/)
})

test('PolishedPanel 渲染 5 章节复用 chapter 列表逻辑', () => {
  const c = read('src/components/PolishedPanel.tsx')
  assert.match(c, /chapters/)
  assert.match(c, /cleanReportText/)
  assert.match(c, /splitParagraphs/)
})

test('PolishedPanel.css 定义 tab + panel + input 基础样式', () => {
  const css = read('src/components/PolishedPanel.css')
  assert.match(css, /\.polished-panel\b/)
  assert.match(css, /\.polished-input-area\b/)
  assert.match(css, /\.polished-content-area\b/)
})

test('ResultPage 装配 PolishedPanel + tab 切换', () => {
  const page = read('src/pages/ResultPage.tsx')
  assert.match(page, /import PolishedPanel from/)
  assert.match(page, /reportTab/)
  assert.match(page, /<PolishedPanel/)
})
