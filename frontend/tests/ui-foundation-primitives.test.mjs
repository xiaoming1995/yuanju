import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (path) => readFileSync(resolve(root, path), 'utf8')

test('global css exposes semantic UX foundation tokens', () => {
  const css = read('src/index.css')
  const tokens = [
    '--surface-base',
    '--surface-panel',
    '--surface-muted',
    '--text-heading',
    '--text-body',
    '--text-subtle',
    '--status-success',
    '--status-warning',
    '--status-danger',
    '--space-page-x',
    '--space-section',
    '--space-card',
    '--radius-control',
    '--radius-panel',
  ]

  for (const token of tokens) {
    assert.match(css, new RegExp(`${token}\\s*:`), `missing token ${token}`)
  }
})

test('shared ui primitives export stable components and class hooks', () => {
  const files = {
    'src/components/ui/Button.tsx': ['Button', 'ui-button'],
    'src/components/ui/SegmentedTabs.tsx': ['SegmentedTabs', 'ui-segmented-tabs'],
    'src/components/ui/StatusBadge.tsx': ['StatusBadge', 'ui-status-badge'],
    'src/components/ui/EmptyState.tsx': ['EmptyState', 'ui-empty-state'],
    'src/components/ui/ConfirmDialog.tsx': ['ConfirmDialog', 'ui-confirm-dialog'],
    'src/components/ui/Toast.tsx': ['ToastProvider', 'ui-toast'],
  }

  for (const [path, [exportName, classHook]] of Object.entries(files)) {
    const src = read(path)
    assert.match(src, new RegExp(`export (function|const) ${exportName}`), `${path} should export ${exportName}`)
    assert.match(src, new RegExp(classHook), `${path} should include ${classHook}`)
  }
})

test('shared ui primitives stay business-domain neutral', () => {
  const paths = [
    'src/components/ui/Button.tsx',
    'src/components/ui/SegmentedTabs.tsx',
    'src/components/ui/StatusBadge.tsx',
    'src/components/ui/EmptyState.tsx',
    'src/components/ui/ConfirmDialog.tsx',
    'src/components/ui/Toast.tsx',
  ]

  for (const path of paths) {
    const src = read(path)
    assert.doesNotMatch(src, /bazi|compatibility|admin|AuthContext|adminApi|baziAPI/i, `${path} should not depend on app domains`)
  }
})
