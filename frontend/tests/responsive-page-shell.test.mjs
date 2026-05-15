import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (path) => readFileSync(resolve(root, path), 'utf8')

test('global page shell reserves fixed top and bottom navigation areas', () => {
  const css = read('src/index.css')
  assert.match(
    css,
    /\.page\s*\{[^}]*padding-top:\s*96px;[^}]*padding-bottom:\s*calc\(140px \+ env\(safe-area-inset-bottom\)\);/s,
  )
  assert.match(
    css,
    /@media \(max-width: 640px\)[\s\S]*\.page\s*\{[^}]*padding-top:\s*84px;[^}]*padding-bottom:\s*calc\(150px \+ env\(safe-area-inset-bottom\)\)\s*!important;/s,
  )
})

test('profile page uses shared page shell in every render state', () => {
  const page = read('src/pages/ProfilePage.tsx')
  const profilePageMatches = page.match(/className="profile-page container page"/g) || []
  assert.equal(profilePageMatches.length, 4)
  assert.doesNotMatch(page, /page-container/)
})
