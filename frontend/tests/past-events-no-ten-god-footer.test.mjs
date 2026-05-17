import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import test from 'node:test'
import assert from 'node:assert/strict'

const root = new URL('..', import.meta.url).pathname

function read(path) {
  return readFileSync(join(root, path), 'utf8')
}

test('PastEventsPage no longer renders 年度力量 footer line', () => {
  const src = read('src/pages/PastEventsPage.tsx')
  assert.doesNotMatch(src, /年度力量：/)
})

test('PastEventsPage no longer references ten_god_power.plain_title in JSX', () => {
  const src = read('src/pages/PastEventsPage.tsx')
  // 命中模式：`y.ten_god_power?.plain_title` 不应在 JSX 渲染条件中出现
  assert.doesNotMatch(src, /ten_god_power\?\.plain_title/)
})
