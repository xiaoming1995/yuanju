import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'
import { mobileQaChecks, mobileQaPages, mobileQaViewports } from './mobilePageQaMatrix.mjs'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (path) => readFileSync(resolve(root, path), 'utf8')

test('mobile QA matrix covers the core user-facing routes', () => {
  const routes = mobileQaPages.map(page => page.route)
  assert.deepEqual(routes, [
    '/',
    '/profile',
    '/history',
    '/compatibility',
    '/compatibility/history',
    '/history/:id',
    '/compatibility/:id',
  ])
})

test('mobile QA matrix defines required mobile viewports and checks', () => {
  assert.deepEqual(mobileQaViewports, [
    { width: 390, height: 844, label: 'iPhone 12/13 portrait' },
    { width: 375, height: 812, label: 'iPhone X/11 Pro portrait' },
    { width: 360, height: 740, label: 'Small Android portrait' },
  ])
  assert.deepEqual(mobileQaChecks, [
    'top-nav-clearance',
    'bottom-nav-clearance',
    'no-horizontal-overflow',
    'primary-content-visible',
  ])
})

test('mobile QA matrix routes are wired with Navbar and BottomNav', () => {
  const app = read('src/App.tsx')
  for (const page of mobileQaPages) {
    const routePattern = new RegExp(
      `path="${page.route.replaceAll('/', '\\/').replace(':id', ':id')}"[\\s\\S]*?<Navbar \\/>[\\s\\S]*?<BottomNav \\/>[\\s\\S]*?<${page.componentName} \\/>`,
    )
    assert.match(app, routePattern, `${page.route} should use Navbar, BottomNav, and ${page.componentName}`)
  }
})

test('mobile QA matrix pages use the shared page shell', () => {
  for (const page of mobileQaPages) {
    const source = read(page.sourcePath)
    assert.match(source, page.pageShellPattern, `${page.componentName} should render the shared page shell`)
  }
})
