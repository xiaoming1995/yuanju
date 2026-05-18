import { test } from 'node:test'
import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import { fileURLToPath } from 'node:url'
import { dirname, resolve } from 'node:path'

const __dirname = dirname(fileURLToPath(import.meta.url))
const REPO_ROOT = resolve(__dirname, '..')

const LEGACY_LIGHT_HEXES = [
  '#fdf9f2',
  '#e0cca0',
  '#2a1a0a',
  '#5a3a1a',
]

const CSS_FILES_TO_CHECK = [
  'src/pages/BrandSettingsPage.css',
  'src/components/LogoCropModal.css',
]

for (const relPath of CSS_FILES_TO_CHECK) {
  test(`${relPath} 不再包含原 light-theme 硬编码色`, async () => {
    const text = (await readFile(resolve(REPO_ROOT, relPath), 'utf8')).toLowerCase()
    for (const hex of LEGACY_LIGHT_HEXES) {
      assert.ok(
        !text.includes(hex.toLowerCase()),
        `${relPath} 仍包含浅色硬编码 ${hex}，应改用 var(--*) token`,
      )
    }
  })
}

test('BrandSettingsPage.tsx logo 上传/删除按钮使用全局 .btn-ghost-sm', async () => {
  const text = await readFile(
    resolve(REPO_ROOT, 'src/pages/BrandSettingsPage.tsx'),
    'utf8',
  )
  assert.match(
    text,
    /className="btn btn-ghost btn-sm"/,
    'BrandSettingsPage.tsx 的 logo 按钮应使用 className="btn btn-ghost btn-sm"',
  )
})

test('LogoCropModal.tsx 取消/确认按钮使用全局 .btn 类', async () => {
  const text = await readFile(
    resolve(REPO_ROOT, 'src/components/LogoCropModal.tsx'),
    'utf8',
  )
  assert.match(
    text,
    /className="btn btn-ghost"/,
    'LogoCropModal.tsx 的取消按钮应使用 className="btn btn-ghost"',
  )
  assert.match(
    text,
    /className="btn btn-primary"/,
    'LogoCropModal.tsx 的确认按钮应使用 className="btn btn-primary"',
  )
  assert.doesNotMatch(
    text,
    /className="logo-crop-btn-(ghost|primary)"/,
    'LogoCropModal.tsx 不应再使用本地 logo-crop-btn-* 类（已废弃）',
  )
})
