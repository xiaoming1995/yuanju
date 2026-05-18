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

test('LogoCropModal exists and locks aspect to 1 in icon mode', () => {
  const src = read('src/components/LogoCropModal.tsx')
  assert.match(src, /import\s+Cropper\s+from\s+['"]react-easy-crop['"]/)
  // icon mode resolves aspect to 1; wordmark uses preset chip
  assert.match(src, /mode === 'icon' \? 1 : wordmarkAspect/)
  assert.match(src, /toBlob\(/)
})

test('BrandSettingsPage routes logo upload through LogoCropModal', () => {
  const src = read('src/pages/BrandSettingsPage.tsx')
  assert.match(src, /import\s+LogoCropModal\s+from\s+['"]\.\.\/components\/LogoCropModal['"]/)
  assert.match(src, /<LogoCropModal[\s\S]+?sourceDataUrl=/)
})

test('ExportBrand interface includes logo_mode union', () => {
  const apiText = read('src/lib/api.ts')
  assert.match(
    apiText,
    /logo_mode:\s*'icon'\s*\|\s*'wordmark'/,
    'src/lib/api.ts ExportBrand 接口缺少 logo_mode union',
  )
})

test('LogoCropModal accepts mode prop and branches on wordmark', () => {
  const text = read('src/components/LogoCropModal.tsx')
  assert.match(text, /mode:\s*'icon'\s*\|\s*'wordmark'/, 'mode prop 类型未声明')
  assert.match(text, /mode === 'wordmark'/, '缺少 wordmark 分支')
})

test('BrandSettingsPage references draft.logo_mode', () => {
  const text = read('src/pages/BrandSettingsPage.tsx')
  assert.match(text, /draft\.logo_mode/, 'BrandSettingsPage 未把 logo_mode 接入 draft')
})

test('ShareCard branches on brand.logo_mode === wordmark', () => {
  const text = read('src/components/ShareCard.tsx')
  assert.match(text, /logo_mode === 'wordmark'/, 'ShareCard 缺少 wordmark 分支')
})

test('PrintLayout branches on brand.logo_mode === wordmark', () => {
  const text = read('src/components/PrintLayout.tsx')
  assert.match(text, /logo_mode === 'wordmark'/, 'PrintLayout 缺少 wordmark 分支')
})

test('BrandPreviewCard branches on brand.logo_mode === wordmark', () => {
  const text = read('src/components/BrandPreviewCard.tsx')
  assert.match(text, /logo_mode === 'wordmark'/, 'BrandPreviewCard 缺少 wordmark 分支')
})
