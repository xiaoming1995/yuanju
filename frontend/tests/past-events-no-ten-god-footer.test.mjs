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

test('PastEventsPage no longer references y.ten_god_power.plain_title in JSX', () => {
  const src = read('src/pages/PastEventsPage.tsx')
  // 命中模式：per-year `y.ten_god_power?.plain_title` 不应再在 JSX 中出现
  // （注意：大运段头的 `meta.ten_god_power?.plain_title` 是不同语境，不在本次删除范围）
  assert.doesNotMatch(src, /y\.ten_god_power\?\.plain_title/)
})
