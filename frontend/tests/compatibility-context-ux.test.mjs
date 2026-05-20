import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (path) => readFileSync(resolve(root, path), 'utf8')

test('compatibility API supports relationship context in create payload', () => {
  const api = read('src/lib/api.ts')
  assert.match(api, /relationship_stage/)
  assert.match(api, /primary_question/)
  assert.match(api, /CompatibilityRelationshipStage/)
  assert.match(api, /CompatibilityPrimaryQuestion/)
})

test('compatibility input page collects relationship context before submit', () => {
  const page = read('src/pages/CompatibilityPage.tsx')
  assert.match(page, /relationshipStage/)
  assert.match(page, /primaryQuestion/)
  assert.match(page, /relationship_stage:\s*relationshipStage/)
  assert.match(page, /primary_question:\s*primaryQuestion/)
  assert.match(page, /关系背景/)
  assert.match(page, /你们目前是什么关系/)
  assert.match(page, /你最想知道什么/)
})

test('compatibility history page shows relationship context labels with fallback', () => {
  const page = read('src/pages/CompatibilityHistoryPage.tsx')
  assert.match(page, /relationshipStageText/)
  assert.match(page, /primaryQuestionText/)
  assert.match(page, /compatibility-history-context/)
  assert.match(page, /关系背景/)
  assert.match(page, /综合关系判断/)
  assert.match(page, /item\.relationship_stage/)
  assert.match(page, /item\.primary_question/)
})
