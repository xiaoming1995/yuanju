import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import test from 'node:test'
import assert from 'node:assert/strict'

const root = new URL('..', import.meta.url).pathname
const read = (p) => readFileSync(join(root, p), 'utf8')

test('PromptSettings shows version badges in three states', () => {
  const page = read('src/pages/admin/PromptSettings.tsx')
  // 三态徽标文案存在
  assert.match(page, /历史遗留/)
  assert.match(page, /已自定义（基于出厂/)
  assert.match(page, /已是出厂版/)
  // PromptRecord 接口暴露 version/is_customized
  assert.match(page, /version:\s*string/)
  assert.match(page, /is_customized:\s*boolean/)
})

test('PromptSettings exposes reset button and confirmation modal', () => {
  const page = read('src/pages/admin/PromptSettings.tsx')
  // 重置按钮文本
  assert.match(page, /重置为系统默认/)
  // 二次确认 modal 文本（不可撤销提示）
  assert.match(page, /此操作不可撤销/)
  // 调用了 reset API
  assert.match(page, /resetToCanonical/)
})

test('adminApi exposes resetToCanonical posting to the reset endpoint', () => {
  const api = read('src/lib/adminApi.ts')
  assert.match(api, /resetToCanonical:\s*\(module/)
  assert.match(api, /\/api\/admin\/prompts\/\$\{module\}\/reset/)
})
