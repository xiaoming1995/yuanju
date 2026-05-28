import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (path) => readFileSync(resolve(root, path), 'utf8')

test('CompatibilityStickyHeader exports component with required props', () => {
  const src = read('src/components/compatibility/CompatibilityStickyHeader.tsx')
  assert.match(src, /export default function CompatibilityStickyHeader/)
  assert.match(src, /selfName/)
  assert.match(src, /partnerName/)
  assert.match(src, /overallScore/)
  assert.match(src, /verdict/)
  assert.match(src, /compat-sticky-header/)
})

test('CompatibilityStickyHeader CSS sets position sticky and var-driven height', () => {
  const css = read('src/components/compatibility/CompatibilityStickyHeader.css')
  assert.match(css, /\.compat-sticky-header\s*\{[\s\S]*?position:\s*sticky/)
  assert.match(css, /\.compat-sticky-header\s*\{[\s\S]*?top:\s*0/)
  assert.match(css, /\.compat-sticky-header\s*\{[\s\S]*?height:\s*var\(--sticky-h\)/)
  assert.match(css, /@media \(min-width: 1024px\)[\s\S]*\.compat-sticky-header\s*\{[\s\S]*?height:\s*var\(--sticky-h-desktop\)/)
})
