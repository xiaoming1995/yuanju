import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'
import ts from 'typescript'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = path => readFileSync(resolve(root, path), 'utf8')

async function importTs(path) {
  const source = readFileSync(resolve(root, path), 'utf8')
  const output = ts.transpileModule(source, {
    compilerOptions: {
      module: ts.ModuleKind.ES2022,
      target: ts.ScriptTarget.ES2022,
      importsNotUsedAsValues: ts.ImportsNotUsedAsValues.Remove,
    },
  }).outputText
  const encoded = Buffer.from(output, 'utf8').toString('base64')
  return import(`data:text/javascript;base64,${encoded}`)
}

function memoryStorage() {
  const map = new Map()
  return {
    getItem: key => map.get(key) ?? null,
    setItem: (key, value) => map.set(key, value),
    removeItem: key => map.delete(key),
  }
}

const validInput = {
  year: 1992,
  month: 8,
  day: 15,
  hour: 10,
  gender: 'female',
  calendar_type: 'solar',
}

test('pending journey stores validates and clears guest bazi context', async () => {
  const pending = await importTs('src/lib/pendingJourney.ts')
  const storage = memoryStorage()
  const journey = pending.createPendingBaziJourney({
    input: validInput,
    anonymousChartId: 'anon-chart-1',
    displayLabel: '我',
    intent: 'generate_report',
    returnPath: '/result',
    now: 1000,
  })

  pending.savePendingJourney(journey, storage)

  const restored = pending.readPendingJourney(storage, 2000)
  assert.equal(restored.type, 'bazi')
  assert.equal(restored.intent, 'generate_report')
  assert.deepEqual(restored.input, validInput)

  pending.clearPendingJourney(storage)
  assert.equal(pending.readPendingJourney(storage, 3000), null)
})

test('pending journey rejects stale or malformed payloads and removes them', async () => {
  const pending = await importTs('src/lib/pendingJourney.ts')
  const storage = memoryStorage()
  const stale = pending.createPendingBaziJourney({
    input: validInput,
    now: 1000,
  })

  pending.savePendingJourney(stale, storage)
  assert.equal(pending.readPendingJourney(storage, 1000 + pending.PENDING_JOURNEY_TTL_MS + 1), null)
  assert.equal(storage.getItem(pending.PENDING_JOURNEY_STORAGE_KEY), null)

  storage.setItem(pending.PENDING_JOURNEY_STORAGE_KEY, JSON.stringify({ type: 'bazi', input: { year: 1899 } }))
  assert.equal(pending.readPendingJourney(storage, 2000), null)
  assert.equal(storage.getItem(pending.PENDING_JOURNEY_STORAGE_KEY), null)
})

test('safe post-auth targets reject open redirects admin routes and unknown paths', async () => {
  const redirects = await importTs('src/lib/authRedirect.ts')

  assert.equal(redirects.getSafePostAuthTarget('/history?page=2'), '/history?page=2')
  assert.equal(redirects.getSafePostAuthTarget('/compatibility/history'), '/compatibility/history')
  assert.equal(redirects.getSafePostAuthTarget('https://evil.example/history'), '/profile')
  assert.equal(redirects.getSafePostAuthTarget('//evil.example/history'), '/profile')
  assert.equal(redirects.getSafePostAuthTarget('/admin'), '/profile')
  assert.equal(redirects.getSafePostAuthTarget('/admin/users'), '/profile')
  assert.equal(redirects.getSafePostAuthTarget('/unknown'), '/profile')
})

test('auth links preserve only safe next targets', async () => {
  const redirects = await importTs('src/lib/authRedirect.ts')

  assert.equal(redirects.buildAuthPath('/login', '/history'), '/login?next=%2Fhistory')
  assert.equal(redirects.buildAuthPath('/register', '/compatibility/history'), '/register?next=%2Fcompatibility%2Fhistory')
  assert.equal(redirects.buildAuthPath('/login', 'https://evil.example'), '/login')
})

test('guest bazi result auth flow stores pending journey and auth pages restore it', () => {
  const resultPage = read('src/pages/ResultPage.tsx')
  const loginPage = read('src/pages/LoginPage.tsx')
  const registerPage = read('src/pages/RegisterPage.tsx')

  assert.match(resultPage, /savePendingJourney\(createPendingBaziJourney/)
  assert.match(resultPage, /pendingIntent\s*!==\s*'generate_report'/)
  assert.match(loginPage, /readPendingJourney\(\)/)
  assert.match(loginPage, /baziAPI\.calculate\(pendingJourney\.input\)/)
  assert.match(loginPage, /clearPendingJourney\(\)/)
  assert.match(loginPage, /pendingIntent:\s*pendingJourney\.intent/)
  assert.match(registerPage, /readPendingJourney\(\)/)
  assert.match(registerPage, /baziAPI\.calculate\(pendingJourney\.input\)/)
  assert.match(registerPage, /clearPendingJourney\(\)/)
  assert.match(registerPage, /pendingIntent:\s*pendingJourney\.intent/)
})

test('profile continuation compares bazi and compatibility timestamps', () => {
  const page = read('src/pages/ProfilePage.tsx')

  assert.match(page, /latestChartTime\s*=\s*latestChart\s*\?\s*new Date\(latestChart\.created_at\)\.getTime\(\)/)
  assert.match(page, /latestCompatibilityTime\s*=\s*latestCompatibility\s*\?\s*new Date\(latestCompatibility\.created_at\)\.getTime\(\)/)
  assert.match(page, /latestCompatibilityTime\s*>\s*latestChartTime/)
  assert.match(page, /label:\s*'新的分析'/)
  assert.match(page, /href:\s*'\/'/)
})

test('mobile bottom navigation exposes archive destination with anonymous return target', () => {
  const nav = read('src/components/BottomNav.tsx')

  assert.match(nav, /History/)
  assert.match(nav, /<span>记录<\/span>/)
  assert.match(nav, /user\s*\?\s*'\/history'\s*:\s*buildAuthPath\('\/login',\s*'\/history'\)/)
  assert.match(nav, /location\.pathname\.startsWith\('\/history'\)/)
})

test('user-facing result and archive flows avoid native dialogs', () => {
  const paths = [
    'src/pages/ResultPage.tsx',
    'src/pages/CompatibilityResultPage.tsx',
    'src/pages/HistoryPage.tsx',
    'src/pages/CompatibilityHistoryPage.tsx',
  ]

  for (const path of paths) {
    const src = read(path)
    assert.doesNotMatch(src, /\balert\(/, `${path} should not use alert()`)
    assert.doesNotMatch(src, /\bconfirm\(/, `${path} should not use confirm()`)
  }

  assert.match(read('src/pages/ResultPage.tsx'), /useToast/)
  assert.match(read('src/pages/CompatibilityResultPage.tsx'), /useToast/)
  assert.match(read('src/pages/HistoryPage.tsx'), /ConfirmDialog/)
  assert.match(read('src/pages/CompatibilityHistoryPage.tsx'), /ConfirmDialog/)
  assert.match(read('src/pages/ResultPage.tsx'), /report-retry-panel/)
})
