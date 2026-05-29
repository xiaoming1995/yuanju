import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (p) => readFileSync(resolve(root, p), 'utf8')

test('PersonalityFit is a naked SECTION using section-level spacing tokens (not a card)', () => {
  const tsx = read('src/components/compatibility/deep-analysis/PersonalityFit.tsx')
  const css = read('src/components/compatibility/deep-analysis/PersonalityFit.css')

  // section element no longer carries the old card class alongside the section class
  assert.match(tsx, /className="compat-section-personality"/)
  assert.doesNotMatch(tsx, /compat-section-personality compat-da-personality/)

  // the section container adopts the same outer rules as the baseline sections
  const block = css.match(/\.compat-section-personality\s*\{[^}]*\}/s)
  assert.ok(block, '.compat-section-personality rule exists')
  assert.match(block[0], /scroll-margin-top:\s*var\(--sticky-h\)/)
  assert.match(block[0], /padding:\s*0 var\(--section-padding-mobile\)/)
  assert.match(block[0], /margin-bottom:\s*var\(--section-gap-mobile\)/)

  // no card skin (border-left was the card marker) anywhere in the personality styles
  assert.doesNotMatch(css, /border-left/)
})

test('DeepReportNarrative keeps its card but adopts section spacing tokens', () => {
  const css = read('src/components/compatibility/deep-analysis/DeepReportNarrative.css')
  const block = css.match(/\.compat-da-report\s*\{[^}]*\}/s)
  assert.ok(block, '.compat-da-report rule exists')
  // adopts section gap/padding so it no longer hugs the evidence drawer above it
  assert.match(block[0], /margin-top:\s*var\(--section-gap-mobile\)/)
  assert.match(block[0], /margin-left:\s*var\(--section-padding-mobile\)/)
  // still a card (附属/交互层)
  assert.match(block[0], /border-left/)
})

test('legacy ScoreOverview heading uses subsection font token, not the old section typography', () => {
  const tsx = read('src/components/compatibility/ScoreOverview.tsx')
  const css = read('src/components/compatibility/ScoreOverview.css')

  // migrated off the old global typography classes
  assert.doesNotMatch(tsx, /compatibility-section-title/)
  assert.doesNotMatch(tsx, /compatibility-section-header/)
  assert.doesNotMatch(tsx, /compatibility-section-desc/)
  assert.match(tsx, /compatibility-quick-score-title/)

  // subsection-scale title (fixes the heading-larger-than-parent inversion)
  const titleBlock = css.match(/\.compatibility-quick-score-title\s*\{[^}]*\}/s)
  assert.ok(titleBlock, '.compatibility-quick-score-title rule exists')
  assert.match(titleBlock[0], /font-size:\s*var\(--fs-subsection-title\)/)
})

test('verdict sticky header surface aligns to the section content column', () => {
  const css = read('src/components/compatibility/CompatibilityStickyHeader.css')
  const block = css.match(/\.compat-sticky-header\s*\{[^}]*\}/s)
  assert.ok(block, '.compat-sticky-header rule exists')
  // surface inset to the column via margin (like EvidenceDrawer), not full-bleed via padding
  assert.match(block[0], /margin:\s*0 var\(--section-padding-mobile\)/)
  // no hardcoded horizontal padding that made the bar overhang the content column
  assert.doesNotMatch(css, /padding:\s*0 16px/)
  assert.doesNotMatch(css, /padding:\s*0 24px/)
})

test('top export actions inset to the section content column', () => {
  const css = read('src/pages/CompatibilityResultPage.css')
  const block = css.match(/\.compat-export-actions\s*\{[^}]*\}/s)
  assert.ok(block, '.compat-export-actions rule exists')
  assert.match(block[0], /padding:\s*0 var\(--section-padding-mobile\)/)
})

test('the old dual typography system is fully removed from page CSS', () => {
  const css = read('src/pages/CompatibilityResultPage.css')
  assert.doesNotMatch(css, /\.compatibility-section-header/)
  assert.doesNotMatch(css, /\.compatibility-section-title/)
  assert.doesNotMatch(css, /\.compatibility-section-desc/)
})
