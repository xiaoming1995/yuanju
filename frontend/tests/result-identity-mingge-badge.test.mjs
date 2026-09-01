import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import test from 'node:test'
import assert from 'node:assert/strict'

const root = new URL('..', import.meta.url).pathname

function read(path) {
  return readFileSync(join(root, path), 'utf8')
}

test('命格说明以紧凑按钮展示且不被摘要高度拉伸', () => {
  const page = read('src/pages/ResultPage.tsx')
  const css = read('src/pages/ResultPage.css')

  assert.match(page, /<button\s+type="button"\s+className="mingge-badge"/)
  assert.match(page, /className="mingge-badge-label">格局<\/span>/)
  assert.match(page, /aria-haspopup="dialog"/)
  assert.match(css, /\.result-tags\s*\{[\s\S]*align-items:\s*flex-start;/)
  assert.match(css, /\.result-tags \.mingge-badge\s*\{[\s\S]*align-self:\s*flex-start;/)
  assert.match(css, /\.result-tags \.mingge-badge\s*\{[\s\S]*border-radius:\s*6px;/)
})
