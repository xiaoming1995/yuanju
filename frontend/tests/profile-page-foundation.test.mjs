import assert from 'node:assert/strict'
import { existsSync, readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (path) => readFileSync(resolve(root, path), 'utf8')

test('profile page route and API are wired', () => {
  assert.ok(existsSync(resolve(root, 'src/pages/ProfilePage.tsx')), 'ProfilePage should exist')
  assert.match(read('src/App.tsx'), /path="\/profile"/)
  assert.match(read('src/lib/api.ts'), /userAPI[\s\S]*profile:\s*\(\)\s*=>\s*api\.get\('\/api\/user\/profile'\)/)
})

test('logged-in navigation points 我的 to profile page', () => {
  assert.match(read('src/components/BottomNav.tsx'), /to="\/profile"[\s\S]*<span>我的<\/span>/)
  assert.match(read('src/components/Navbar.tsx'), /to="\/profile"[\s\S]*我的/)
})

test('profile page reserves inactive recharge and PDF template entrypoints', () => {
  const page = read('src/pages/ProfilePage.tsx')
  assert.match(page, /充值|点数/)
  assert.match(page, /PDF.*模板|模板.*PDF/)
  assert.match(page, /即将开放/)
  assert.doesNotMatch(page, /\/pay|\/payment|createOrder|支付成功/)
})
