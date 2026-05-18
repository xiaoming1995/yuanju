import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import test from 'node:test'
import assert from 'node:assert/strict'

const root = new URL('..', import.meta.url).pathname

function read(path) {
  return readFileSync(join(root, path), 'utf8')
}

test('AlgoConfigPage exposes cleanup_enabled label', () => {
  const src = read('src/pages/admin/AlgoConfigPage.tsx')
  assert.match(src, /cleanup_enabled:/)
})

test('AlgoConfigPage exposes cleanup_retention_days label', () => {
  const src = read('src/pages/admin/AlgoConfigPage.tsx')
  assert.match(src, /cleanup_retention_days:/)
})

test('AlgoConfigPage exposes cleanup_run_hour label', () => {
  const src = read('src/pages/admin/AlgoConfigPage.tsx')
  assert.match(src, /cleanup_run_hour:/)
})
