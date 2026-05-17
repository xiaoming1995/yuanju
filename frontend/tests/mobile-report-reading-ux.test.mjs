import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (path) => readFileSync(resolve(root, path), 'utf8')

test('result page exposes summary-first report reading structure', () => {
  const page = read('src/pages/ResultPage.tsx')
  assert.match(page, /buildReportDigestItems/)
  assert.match(page, /report-digest-card/)
  assert.match(page, /report-term-glossary/)
  assert.match(page, /report-chapter-list/)
  assert.match(page, /report-chapter-detail/)
  assert.match(page, /report-action-bar/)
})

test('result page report actions are class based and include follow-up paths', () => {
  const page = read('src/pages/ResultPage.tsx')
  assert.match(page, /report-header-actions/)
  assert.match(page, /查看历史/)
  assert.match(page, /过往事件/)
  assert.match(page, /重新起盘/)
  assert.doesNotMatch(page, /report-section-header[\s\S]{0,240}style=\{\{/)
})

test('result page only exposes PDF report export after AI interpretation exists', () => {
  const page = read('src/pages/ResultPage.tsx')
  assert.match(
    page,
    /\{report && \(\s*<>[\s\S]*id="export-report-btn"[\s\S]*导出[\s\S]*PDF[\s\S]*<\/>\s*\)\}/,
    'header PDF export should be gated by an existing AI report',
  )
  assert.match(
    page,
    /\{report && \(\s*<div className="report-action-bar">[\s\S]*导出[\s\S]*PDF[\s\S]*<\/div>\s*\)\}/,
    'PDF report actions should be gated by an existing AI report',
  )
  // PrintLayout 现在也允许在润色版存在时 mount（report 不存在但 polishedReport 存在的边缘场景）
  assert.match(
    page,
    /\{\(report \|\| polishedReport\) && \([\s\S]*<PrintLayout[\s\S]*structured=\{printStructured\}/,
    'print/PDF template should mount when either original report or polished report exists',
  )
})

test('result page css supports mobile report reading and safe area', () => {
  const css = read('src/pages/ResultPage.css')
  assert.match(css, /\.result-page\.page\s*\{[^}]*padding-bottom:\s*calc\(140px \+ env\(safe-area-inset-bottom\)\)\s*!important;/s)
  assert.match(css, /\.report-digest-card/)
  assert.match(css, /\.report-term-glossary/)
  assert.match(css, /\.report-chapter-detail/)
  assert.match(css, /\.report-action-bar/)
  assert.match(css, /@media \(max-width: 640px\)[\s\S]*\.report-action-bar\s*\{[^}]*grid-template-columns:\s*1fr;/s)
})
