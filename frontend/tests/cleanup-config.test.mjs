import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import test from 'node:test'
import assert from 'node:assert/strict'

const root = new URL('..', import.meta.url).pathname

function read(path) {
  return readFileSync(join(root, path), 'utf8')
}

test('CleanupConfigPage exists and uses adminCleanupConfigAPI', () => {
  const src = read('src/pages/admin/CleanupConfigPage.tsx')
  assert.match(src, /adminCleanupConfigAPI/)
  assert.match(src, /retention_days/)
  assert.match(src, /run_hour/)
  assert.match(src, /enabled/)
})

test('adminApi exposes adminCleanupConfigAPI with get + update', () => {
  const src = read('src/lib/adminApi.ts')
  assert.match(src, /adminCleanupConfigAPI/)
  assert.match(src, /\/api\/admin\/cleanup-config/)
  assert.match(src, /\bget:\s*\(\)/)
  assert.match(src, /\bupdate:\s*\(data:/)
})

test('App.tsx registers /admin/cleanup-config route', () => {
  const src = read('src/App.tsx')
  assert.match(src, /path="cleanup-config"/)
  assert.match(src, /import CleanupConfigPage/)
})

test('AdminLayout nav includes 数据清理配置 entry', () => {
  const src = read('src/components/AdminLayout.tsx')
  assert.match(src, /\/admin\/cleanup-config/)
  assert.match(src, /数据清理配置/)
})

test('AlgoConfigPage PARAM_LABELS no longer contains cleanup_* keys', () => {
  const src = read('src/pages/admin/AlgoConfigPage.tsx')
  assert.doesNotMatch(src, /cleanup_enabled/)
  assert.doesNotMatch(src, /cleanup_retention_days/)
  assert.doesNotMatch(src, /cleanup_run_hour/)
})
