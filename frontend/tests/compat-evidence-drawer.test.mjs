import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'
const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (p) => readFileSync(resolve(root, p), 'utf8')

test('EvidenceDrawer is a single details-open block', () => {
  const src = read('src/components/compatibility/EvidenceDrawer.tsx')
  assert.match(src, /export default function EvidenceDrawer/)
  assert.match(src, /<details open/)
  assert.match(src, /compat-evidence-drawer/)
  assert.match(src, /关键判断依据/)
  assert.match(src, /命盘细节/)
})

test('page no longer defines EvidenceLinkedClaims / ProfessionalEvidenceGroups / EvidenceCard inline', () => {
  const page = read('src/pages/CompatibilityResultPage.tsx')
  assert.doesNotMatch(page, /^function EvidenceLinkedClaims\(/m)
  assert.doesNotMatch(page, /^function ProfessionalEvidenceGroups\(/m)
  assert.doesNotMatch(page, /^function EvidenceCard\(/m)
})
