import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const tsx = readFileSync(resolve(root, 'src/pages/ResultPage.tsx'), 'utf8')
const css = readFileSync(resolve(root, 'src/pages/ResultPage.css'), 'utf8')

test('structured bazi reports default to full interpretation mode', () => {
  assert.match(
    tsx,
    /useState<'brief' \| 'detail'>\('detail'\)/,
    'structured reports should open in the detailed interpretation mode',
  )
})

test('report mode labels distinguish quick conclusions from full interpretation', () => {
  assert.match(tsx, />精简版<\/button>/)
  assert.match(tsx, />完整解读<\/button>/)
})

test('report body typography supports long-form Chinese reading', () => {
  assert.match(css, /\.report-block-content\s*\{[^}]*font-size:\s*16px;/s)
  assert.match(css, /\.report-block-content\s*\{[^}]*line-height:\s*1\.9;/s)
  assert.match(css, /\.report-content\s*\{[^}]*font-size:\s*16px;/s)
})
