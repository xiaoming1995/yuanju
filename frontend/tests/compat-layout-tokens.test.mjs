import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (path) => readFileSync(resolve(root, path), 'utf8')

test('compatibility result page declares 14 layout tokens in page scope', () => {
  const css = read('src/pages/CompatibilityResultPage.css')
  const tokens = [
    '--section-gap-mobile',
    '--section-gap-desktop',
    '--section-padding-mobile',
    '--section-padding-desktop',
    '--subsection-gap',
    '--fs-section-kicker',
    '--fs-section-title',
    '--fs-section-title-desktop',
    '--fs-subsection-title',
    '--fs-body',
    '--fs-caption',
    '--sticky-h',
    '--sticky-h-desktop',
    '--container-max',
  ]
  for (const token of tokens) {
    assert.match(css, new RegExp(`\\.compatibility-result-page[^{]*\\{[^}]*${token.replace(/-/g, '\\-')}\\s*:`, 's'),
      `expected ${token} declared inside .compatibility-result-page block`)
  }
})
