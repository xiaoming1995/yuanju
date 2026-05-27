import test from 'node:test'
import assert from 'node:assert/strict'
import fs from 'node:fs'

const read = path => fs.readFileSync(new URL(`../${path}`, import.meta.url), 'utf8')

test('API exposes chart display names and update endpoint', () => {
  const api = read('src/lib/api.ts')
  assert.match(api, /display_name\??:\s*string/)
  assert.match(api, /updateHistoryDisplayName:\s*\(id:\s*string,\s*displayName:\s*string\)/)
  assert.match(api, /\/api\/bazi\/history\/\$\{id\}\/display-name/)
})

test('compatibility create payload supports participant display names', () => {
  const api = read('src/lib/api.ts')
  assert.match(api, /self_display_name\??:\s*string/)
  assert.match(api, /partner_display_name\??:\s*string/)
})

test('birth profile helper converts saved charts into compatibility form values', () => {
  const helper = read('src/components/birthProfile.ts')
  assert.match(helper, /export function chartToBirthProfile/)
  assert.match(helper, /birth_year/)
  assert.match(helper, /calendarType:\s*chart\.calendar_type\s*\|\|\s*'solar'/)
  assert.match(helper, /isLeapMonth:\s*Boolean\(chart\.is_leap_month\)/)
})

test('history page supports chart naming and compatibility launch role choice', () => {
  const page = read('src/pages/HistoryPage.tsx')
  const css = read('src/pages/HistoryPage.css')
  assert.match(page, /editingChartId/)
  assert.match(page, /handleSaveDisplayName/)
  assert.match(page, /compatibilityRoleChart/)
  assert.match(page, /作为我/)
  assert.match(page, /作为对方/)
  assert.match(page, /\/compatibility\?importChart=\$\{compatibilityRoleChart\.id\}&role=/)
  assert.match(css, /history-display-name/)
  assert.match(css, /history-role-dialog/)
})

test('history page guards nested actions and role dialog keyboard focus', () => {
  const page = read('src/pages/HistoryPage.tsx')
  assert.match(page, /isInteractiveTarget/)
  assert.match(page, /target\.closest\('button, input, select, textarea, a'\)/)
  assert.match(page, /roleDialogRef/)
  assert.match(page, /roleDialogRef\.current\?\.focus\(\)/)
  assert.match(page, /event\.key === 'Escape'/)
  assert.match(page, /previous\?\.focus\(\)/)
  assert.match(page, /tabIndex=\{-1\}/)
})
