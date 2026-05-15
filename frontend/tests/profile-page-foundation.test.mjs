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
  assert.match(page, /profile-feature-status/)
  assert.doesNotMatch(page, /\/pay|\/payment|createOrder|支付成功/)
})

test('profile page provides a continuation workbench', () => {
  const page = read('src/pages/ProfilePage.tsx')
  const css = read('src/pages/ProfilePage.css')
  assert.match(page, /profile-workbench/)
  assert.match(page, /continueTarget/)
  assert.match(page, /继续上次分析/)
  assert.match(page, /最近命盘/)
  assert.match(page, /最近合盘/)
  assert.match(page, /to=\{continueTarget\.href\}/)
  assert.match(css, /\.profile-workbench/)
})

test('profile stats link to existing archive destinations', () => {
  const page = read('src/pages/ProfilePage.tsx')
  assert.match(page, /to: '\/history'/)
  assert.match(page, /to: '\/compatibility\/history'/)
  assert.match(page, /profile-stat-card--link/)
  assert.match(page, /profile-stat-card--static/)
})

test('profile page reserves mobile bottom navigation safe area', () => {
  const css = read('src/index.css')
  const page = read('src/pages/ProfilePage.tsx')
  assert.match(page, /className="profile-page container page"/)
  assert.doesNotMatch(page, /page-container/)
  assert.match(
    css,
    /\.page\s*\{[^}]*padding-bottom:\s*calc\(140px \+ env\(safe-area-inset-bottom\)\);/s,
  )
  assert.match(
    css,
    /\.page\s*\{[^}]*padding-top:\s*96px;/s,
  )
  assert.match(
    css,
    /@media \(max-width: 640px\)[\s\S]*\.page\s*\{[^}]*padding-bottom:\s*calc\(150px \+ env\(safe-area-inset-bottom\)\)\s*!important;/s,
  )
  assert.match(
    css,
    /@media \(max-width: 640px\)[\s\S]*\.page\s*\{[^}]*padding-top:\s*84px;/s,
  )
})
