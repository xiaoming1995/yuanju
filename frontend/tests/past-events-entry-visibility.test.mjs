import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import test from 'node:test'
import assert from 'node:assert/strict'

const root = new URL('..', import.meta.url).pathname

function read(path) {
  return readFileSync(join(root, path), 'utf8')
}

test('past-events entry is mounted inside dayun-section after DayunTimeline', () => {
  const page = read('src/pages/ResultPage.tsx')
  const dayunBlockMatch = page.match(/<section[^>]*className="dayun-section">[\s\S]*?<\/section>/)
  assert.ok(dayunBlockMatch, 'dayun-section block not found')
  const block = dayunBlockMatch[0]
  assert.match(block, /<DayunTimeline[\s\S]*?\/>[\s\S]*?past-events-entry/)
})

test('past-events entry renders for logged-in users and disabled for guests', () => {
  const page = read('src/pages/ResultPage.tsx')
  assert.match(page, /past-events-entry/)
  assert.match(page, /isGuest/)
  assert.match(page, /登录后可查看/)
  assert.match(page, /展开每个大运段，看年份信号与白话批语/)
})

test('past-events entry navigates to /bazi/:chartId/past-events when enabled', () => {
  const page = read('src/pages/ResultPage.tsx')
  assert.match(
    page,
    /past-events-entry[\s\S]*?navigate\(`\/bazi\/\$\{targetId\}\/past-events`\)/,
  )
})

test('report-action-bar no longer contains a past-events navigation button', () => {
  const page = read('src/pages/ResultPage.tsx')
  const barMatch = page.match(/<div className="report-action-bar">[\s\S]*?<\/div>/)
  assert.ok(barMatch, 'report-action-bar block not found')
  assert.doesNotMatch(barMatch[0], /past-events/)
})

test('past-events-entry css defines hover, focus, and disabled states', () => {
  const css = read('src/pages/ResultPage.css')
  assert.match(css, /\.past-events-entry\s*\{/)
  assert.match(css, /\.past-events-entry:hover/)
  assert.match(css, /\.past-events-entry\.is-disabled/)
})
