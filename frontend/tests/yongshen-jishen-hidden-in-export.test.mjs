import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import test from 'node:test'
import assert from 'node:assert/strict'

const root = new URL('..', import.meta.url).pathname

function read(path) {
  return readFileSync(join(root, path), 'utf8')
}

test('ShareCard source does not contain 喜用神/忌神 badge text', () => {
  const src = read('src/components/ShareCard.tsx')
  assert.doesNotMatch(src, /喜用神：/)
  assert.doesNotMatch(src, /忌神：/)
})

test('ShareCard props interface no longer declares yongshen/jishen', () => {
  const src = read('src/components/ShareCard.tsx')
  assert.doesNotMatch(src, /yongshen:\s*string/)
  assert.doesNotMatch(src, /jishen:\s*string/)
})

test('PrintLayout source does not contain 喜用神/忌神 badge text', () => {
  const src = read('src/components/PrintLayout.tsx')
  assert.doesNotMatch(src, /喜用神：/)
  assert.doesNotMatch(src, /忌 ?神：/)
})

test('PrintLayout glossary no longer lists 用神 or 忌神 terms', () => {
  const src = read('src/components/PrintLayout.tsx')
  assert.doesNotMatch(src, /term:\s*'用神'/)
  assert.doesNotMatch(src, /term:\s*'忌神'/)
})

test('PrintLayout props interface no longer declares yongshen/jishen', () => {
  const src = read('src/components/PrintLayout.tsx')
  assert.doesNotMatch(src, /yongshen:\s*string/)
  assert.doesNotMatch(src, /jishen:\s*string/)
})

test('ResultPage no longer passes yongshen/jishen to ShareCard or PrintLayout', () => {
  const src = read('src/pages/ResultPage.tsx')
  const shareCardMatch = src.match(/<ShareCard\b[\s\S]*?\/>/)
  assert.ok(shareCardMatch, 'ShareCard mount not found in ResultPage')
  assert.doesNotMatch(shareCardMatch[0], /yongshen=/)
  assert.doesNotMatch(shareCardMatch[0], /jishen=/)

  const printLayoutMatch = src.match(/<PrintLayout\b[\s\S]*?\/>/)
  assert.ok(printLayoutMatch, 'PrintLayout mount not found in ResultPage')
  assert.doesNotMatch(printLayoutMatch[0], /yongshen=/)
  assert.doesNotMatch(printLayoutMatch[0], /jishen=/)
})
