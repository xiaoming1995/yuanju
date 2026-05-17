import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import test from 'node:test'
import assert from 'node:assert/strict'

const root = new URL('..', import.meta.url).pathname

function read(path) {
  return readFileSync(join(root, path), 'utf8')
}

test('ShareCard accepts optional brand prop', () => {
  const src = read('src/components/ShareCard.tsx')
  assert.match(src, /brand\?:\s*ExportBrand/)
})

test('ShareCard preserves 缘聚 命 理 default when brand title is empty', () => {
  const src = read('src/components/ShareCard.tsx')
  assert.match(src, /缘 聚 命 理/)
  assert.match(src, /brand\?\.title\s*\|\|\s*['"]缘 聚 命 理['"]/)
})

test('PrintLayout accepts optional brand prop', () => {
  const src = read('src/components/PrintLayout.tsx')
  assert.match(src, /brand\?:\s*ExportBrand/)
})

test('vite dev proxy includes /static for backend static files', () => {
  const src = read('vite.config.ts')
  assert.match(src, /['"]\/static['"]/)
  assert.match(src, /localhost:9002/)
})
