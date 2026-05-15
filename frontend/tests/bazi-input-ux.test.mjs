import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (path) => readFileSync(resolve(root, path), 'utf8')

test('birth profile form exposes confirmation summary and leap month text', () => {
  const form = read('src/components/BirthProfileForm.tsx')
  assert.match(form, /formatBirthProfileSummary/)
  assert.match(form, /闰\$\{value\.month\}月|闰.*月/)
  assert.match(form, /按北京时间排盘/)
  assert.match(form, /birth-profile-summary/)
})

test('zi hour is disambiguated with explicit mutually exclusive options', () => {
  const form = read('src/components/BirthProfileForm.tsx')
  assert.match(form, /ziHourMode/)
  assert.match(form, /23:00-23:59/)
  assert.match(form, /00:00-00:59/)
  assert.match(form, /按前一日/)
  assert.match(form, /按当日/)
  assert.doesNotMatch(form, /checkbox-label[\s\S]*早子时/)
})

test('homepage moves precision controls behind advanced calibration without changing payload mapping', () => {
  const home = read('src/pages/HomePage.tsx')
  assert.match(home, /showAdvancedCalibration/)
  assert.match(home, /高级校准/)
  assert.match(home, /longitude:\s*PROVINCE_LONGITUDE\[province\]\s*\|\|\s*0/)
  assert.match(home, /is_early_zishi:\s*isZishi\s*\?\s*isEarlyZishi\s*:\s*false/)
})

test('mobile css supports single-column birth input and visible submit flow', () => {
  const birthCss = read('src/components/BirthProfileForm.css')
  const homeCss = read('src/pages/HomePage.css')
  assert.match(birthCss, /\.birth-profile-basic/)
  assert.match(birthCss, /\.birth-date-fields/)
  assert.match(birthCss, /min-height:\s*44px/)
  assert.match(homeCss, /advanced-calibration/)
  assert.match(homeCss, /@media \(max-width: 768px\)/)
})

test('mobile compact layout keeps primary controls dense and month day side by side', () => {
  const form = read('src/components/BirthProfileForm.tsx')
  const css = read('src/components/BirthProfileForm.css')
  assert.match(form, /birth-profile-primary-grid/)
  assert.match(form, /formatYearOption/)
  assert.match(form, /岁/)
  assert.match(css, /\.birth-profile-primary-grid/)
  assert.match(css, /\.birth-date-fields\s*\{[^}]*grid-template-columns:\s*2fr 1fr 1fr;/s)
  assert.match(css, /@media \(max-width: 768px\)[\s\S]*\.birth-date-fields\s*\{[^}]*grid-template-columns:\s*1fr 1fr;/s)
  assert.match(css, /@media \(max-width: 768px\)[\s\S]*\.birth-year-field\s*\{[^}]*grid-column:\s*1 \/ -1;/s)
})

test('compatibility page uses the improved birth input experience for both participants', () => {
  const page = read('src/pages/CompatibilityPage.tsx')
  assert.match(page, /showSummary/)
  assert.match(page, /我的生辰/)
  assert.match(page, /对方的生辰/)
})

test('mobile bottom navigation remains available on birth input routes', () => {
  const nav = read('src/components/BottomNav.tsx')
  const css = read('src/components/BottomNav.css')
  assert.doesNotMatch(nav, /isInputRoute/)
  assert.doesNotMatch(nav, /bottom-nav--input-route/)
  assert.doesNotMatch(css, /\.bottom-nav--input-route/)
  assert.match(read('src/pages/HomePage.css'), /\.home-page\.page\s*\{[^}]*padding-bottom:\s*calc\(140px \+ env\(safe-area-inset-bottom\)\)\s*!important;/s)
})

test('mobile homepage hero is reduced so the form appears sooner', () => {
  const css = read('src/pages/HomePage.css')
  assert.match(css, /@media \(max-width: 768px\)[\s\S]*\.hero-badge\s*\{[^}]*display:\s*none;/s)
  assert.match(css, /@media \(max-width: 768px\)[\s\S]*\.hero-title\s*\{[^}]*font-size:\s*28px;/s)
  assert.match(css, /@media \(max-width: 768px\)[\s\S]*\.hero-desc\s*\{[^}]*font-size:\s*13px;/s)
})

test('compatibility page has mobile tabs while preserving desktop panels', () => {
  const page = read('src/pages/CompatibilityPage.tsx')
  const css = read('src/pages/CompatibilityPage.css')
  assert.match(page, /activeProfile/)
  assert.match(page, /compatibility-mobile-tabs/)
  assert.match(page, /compatibility-profile-panel--active/)
  assert.match(css, /\.compatibility-mobile-tabs\s*\{[^}]*display:\s*none;/s)
  assert.match(css, /@media \(max-width: 768px\)[\s\S]*\.compatibility-mobile-tabs\s*\{[^}]*display:\s*grid;/s)
  assert.match(css, /@media \(max-width: 768px\)[\s\S]*\.compatibility-profile-panel\s*\{[^}]*display:\s*none;/s)
  assert.match(css, /@media \(max-width: 768px\)[\s\S]*\.compatibility-profile-panel--active\s*\{[^}]*display:\s*block;/s)
})
