import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import test from 'node:test'
import assert from 'node:assert/strict'

const root = new URL('..', import.meta.url).pathname

function read(path) {
  return readFileSync(join(root, path), 'utf8')
}

test('大运总览为十二地支维护独立五行映射', () => {
  const timeline = read('src/components/DayunTimeline.tsx')

  assert.match(timeline, /const ZHI_WUXING: Record<string, string>/)
  for (const zhi of ['子', '丑', '寅', '卯', '辰', '巳', '午', '未', '申', '酉', '戌', '亥']) {
    assert.match(timeline, new RegExp(`${zhi}: '\\w+'`))
  }
  assert.match(timeline, /const zhiWuxing = ZHI_WUXING\[d\.zhi\]/)
  assert.match(timeline, /wuxing-text-\$\{zhiWuxing\}/)
})

test('大运总览同时呈现干支十神并安全忽略缺失值', () => {
  const timeline = read('src/components/DayunTimeline.tsx')

  assert.match(timeline, /d\.gan_shishen \|\| d\.zhi_shishen/)
  assert.match(timeline, /d\.gan_shishen && <span>干·\{d\.gan_shishen\}<\/span>/)
  assert.match(timeline, /d\.zhi_shishen && <span>支·\{d\.zhi_shishen\}<\/span>/)
})

test('大运卡片不再将地支强制设为默认文字色', () => {
  const css = read('src/pages/ResultPage.css')

  assert.doesNotMatch(css, /\.dayun-step-ganzhi span:last-child\s*\{\s*color:/)
  assert.match(css, /\.dayun-step-ten-god\s*\{[\s\S]*flex-wrap:\s*wrap;/)
  assert.match(css, /\.dayun-step-ten-god span\s*\{[\s\S]*white-space:\s*nowrap;/)
})
