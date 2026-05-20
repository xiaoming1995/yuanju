import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (path) => readFileSync(resolve(root, path), 'utf8')

test('ordinary API exposes public registration settings endpoint', () => {
  const src = read('src/lib/api.ts')
  assert.match(src, /registrationSettings:\s*\(\)\s*=>\s*api\.get\('\/api\/auth\/registration-settings'\)/)
  assert.match(src, /registration_enabled/)
})

test('admin API exposes create user and registration settings controls', () => {
  const src = read('src/lib/adminApi.ts')
  assert.match(src, /adminUsersAPI/)
  assert.match(src, /create:\s*\(data:/)
  assert.match(src, /adminApi\.post\('\/api\/admin\/users'/)
  assert.match(src, /adminRegistrationSettingsAPI/)
  assert.match(src, /\/api\/admin\/settings\/registration/)
})

test('admin users page includes registration toggle and create user form', () => {
  const src = read('src/pages/admin/AdminUsersPage.tsx')
  assert.match(src, /公开注册/)
  assert.match(src, /registration_enabled/)
  assert.match(src, /创建用户/)
  assert.match(src, /初始密码/)
  assert.match(src, /admin_created|后台创建/)
})

test('public registration entrypoints respect registration availability', () => {
  assert.match(read('src/components/Navbar.tsx'), /registration_enabled/)
  assert.match(read('src/pages/RegisterPage.tsx'), /暂未开放公开注册/)
  assert.match(read('src/pages/RegisterPage.tsx'), /registrationSettings/)
  assert.match(read('src/pages/ResultPage.tsx'), /registration_enabled/)
})
