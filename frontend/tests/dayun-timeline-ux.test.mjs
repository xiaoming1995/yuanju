import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import test from 'node:test'
import assert from 'node:assert/strict'

const root = new URL('..', import.meta.url).pathname

function read(path) {
  return readFileSync(join(root, path), 'utf8')
}

test('dayun timeline uses a responsive grid instead of a horizontal scroll strip', () => {
  const component = read('src/components/DayunTimeline.tsx')

  assert.match(component, /className="dayun-design-shell"/)
  assert.match(component, /className="dayun-design-panel"/)
  assert.match(component, /className="dayun-design-header"/)
  assert.match(component, /className="dayun-meta-row"/)
  assert.match(component, /className="dayun-overview-grid"/)
  assert.match(component, /className=\{`dayun-step-card/)
  assert.doesNotMatch(component, /overflowX:\s*'auto'/)
  assert.doesNotMatch(component, /minWidth:\s*'max-content'/)
})

test('result page lets the dayun component own the replicated mockup shell', () => {
  const page = read('src/pages/ResultPage.tsx')

  assert.match(page, /<section className="dayun-section">/)
  assert.doesNotMatch(page, /<div className="dayun-section card">/)
  assert.doesNotMatch(page, /<h2 className="section-title serif">大运时间轴<\/h2>\s*<DayunTimeline/)
  assert.match(page, /gender=\{result\.gender\}/)
  assert.match(page, /pillarsLabel=\{dayunPillarsLabel\}/)
})

test('liunian detail cards expose readable states for current, transition, and focus years', () => {
  const component = read('src/components/DayunTimeline.tsx')

  assert.match(component, /is-liunian-current/)
  assert.match(component, /is-liunian-transition/)
  assert.match(component, /is-liunian-focus/)
  assert.match(component, /className="liunian-card-topline"/)
  assert.match(component, /className="liunian-transition-ribbon"/)
  assert.match(component, /重点/)
  assert.doesNotMatch(component, /style=\{\{\s*background:\s*isLnCurrent/)
})

test('dayun timeline css keeps desktop dayun cards in one row while preserving mobile wrapping', () => {
  const css = read('src/pages/ResultPage.css')

  assert.match(css, /\.dayun-design-shell\s*\{[\s\S]*border-radius:\s*var\(--radius-lg\);/)
  assert.match(css, /\.dayun-design-panel\s*\{[\s\S]*padding:\s*24px 26px 18px;/)
  assert.match(css, /\.dayun-design-header\s*\{[\s\S]*margin-bottom:\s*22px;/)
  assert.match(css, /\.dayun-meta-row\s*\{[\s\S]*display:\s*flex;/)
  assert.match(css, /\.dayun-overview-grid\s*\{[\s\S]*display:\s*grid;[\s\S]*grid-template-columns:\s*repeat\(10,\s*minmax\(0,\s*1fr\)\);/)
  assert.match(css, /\.dayun-step-card\s*\{[\s\S]*min-width:\s*0;/)
  assert.match(css, /\.dayun-step-card\s*\{[\s\S]*min-height:\s*156px;/)
  assert.match(css, /@media \(max-width: 900px\)\s*\{[\s\S]*\.dayun-overview-grid\s*\{[\s\S]*grid-template-columns:\s*repeat\(5,\s*minmax\(0,\s*1fr\)\);/)
  assert.match(css, /@media \(max-width: 640px\)\s*\{[\s\S]*\.dayun-overview-grid\s*\{[\s\S]*grid-template-columns:\s*repeat\(2,\s*minmax\(0,\s*1fr\)\);/)
})

test('liunian detail css provides card hierarchy and state highlights', () => {
  const css = read('src/pages/ResultPage.css')

  assert.match(css, /\.liunian-panel\s*\{[\s\S]*background:\s*rgba\(20,\s*23,\s*32,\s*0\.6\);/)
  assert.match(css, /\.liunian-card\s*\{[\s\S]*display:\s*grid;/)
  assert.match(css, /\.liunian-card\s*\{[\s\S]*min-width:\s*0;/)
  assert.match(css, /\.liunian-card\.is-liunian-current/)
  assert.match(css, /\.liunian-card\.is-liunian-transition/)
  assert.match(css, /\.liunian-card\.is-liunian-focus/)
  assert.match(css, /\.liunian-transition-ribbon/)
})

test('liunian detail matches the compact timeline mockup with ten cards on desktop', () => {
  const component = read('src/components/DayunTimeline.tsx')
  const css = read('src/pages/ResultPage.css')

  assert.match(component, /className="liunian-panel-title-row"/)
  assert.match(component, /className="liunian-card-badges"/)
  assert.match(component, /className="liunian-card-divider"/)
  assert.match(component, /className="liunian-open-cue"/)
  assert.doesNotMatch(component, /getTenGodInsight/)
  assert.doesNotMatch(component, /className="liunian-insight"/)

  assert.match(css, /\.liunian-grid\s*\{[\s\S]*grid-template-columns:\s*repeat\(10,\s*minmax\(0,\s*1fr\)\);/)
  assert.match(css, /\.liunian-card\s*\{[\s\S]*justify-items:\s*center;[\s\S]*min-height:\s*180px;/)
  assert.match(css, /\.liunian-card-divider\s*\{[\s\S]*width:\s*72%;/)
  assert.match(css, /@media \(max-width: 640px\)\s*\{[\s\S]*\.liunian-grid\s*\{[\s\S]*grid-template-columns:\s*repeat\(2,\s*minmax\(0,\s*1fr\)\)/)
})

test('dayun timeline includes the mockup summary strip and mobile controls', () => {
  const component = read('src/components/DayunTimeline.tsx')
  const css = read('src/pages/ResultPage.css')

  assert.match(component, /className="dayun-mobile-topbar"/)
  assert.match(component, /className="dayun-summary-strip"/)
  assert.match(component, /className="dayun-summary-copy"/)
  assert.match(component, /className="dayun-summary-tags"/)
  assert.match(component, /className="dayun-disclaimer"/)
  assert.match(component, /getTrendKeywords/)

  assert.match(css, /\.dayun-summary-strip\s*\{[\s\S]*display:\s*grid;/)
  assert.match(css, /\.dayun-summary-tags\s*\{[\s\S]*display:\s*flex;/)
  assert.match(css, /\.dayun-mobile-topbar\s*\{[\s\S]*display:\s*none;/)
  assert.match(css, /@media \(max-width: 640px\)\s*\{[\s\S]*\.dayun-mobile-topbar\s*\{[\s\S]*display:\s*flex;/)
})

test('dayun timeline keeps backend dayun data unchanged and only styles the received periods', () => {
  const component = read('src/components/DayunTimeline.tsx')

  assert.match(component, /const displayDayun = dayun\.slice\(0,\s*10\)/)
  assert.doesNotMatch(component, /buildFallbackDayun/)
  assert.doesNotMatch(component, /getYearGanZhi/)
  assert.doesNotMatch(component, /GAN_YINYANG/)
  assert.doesNotMatch(component, /ZHI_MAIN_GAN/)
})

test('dayun timeline preserves all backend shensha labels instead of truncating them', () => {
  const component = read('src/components/DayunTimeline.tsx')

  assert.match(component, /d\.shen_sha\.map\(\(ss,\s*si\)/)
  assert.doesNotMatch(component, /d\.shen_sha\.slice\(0,\s*2\)/)
})

test('dayun card shensha labels do not stretch the compact mockup cards', () => {
  const css = read('src/pages/ResultPage.css')

  assert.match(css, /\.dayun-shensha-list\s*\{[\s\S]*position:\s*absolute;/)
  assert.match(css, /\.dayun-shensha-list\s*\{[\s\S]*max-height:\s*0;/)
  assert.match(css, /\.dayun-step-card:hover \.dayun-shensha-list/)
  assert.match(css, /\.dayun-step-card:focus-visible \.dayun-shensha-list/)
})

test('liunian markers match the mockup dot and corner-ribbon treatment', () => {
  const component = read('src/components/DayunTimeline.tsx')
  const css = read('src/pages/ResultPage.css')

  assert.match(component, /className="liunian-transition-ribbon"[^>]*>交脱<\/span>/)
  assert.match(css, /\.liunian-current-badge\s*\{[\s\S]*width:\s*10px;[\s\S]*height:\s*10px;/)
  assert.match(css, /\.liunian-current-badge\s*\{[\s\S]*text-indent:\s*-999px;/)
  assert.match(css, /\.liunian-transition-ribbon\s*\{[\s\S]*transform:\s*rotate\(45deg\)/)
  assert.match(css, /\.liunian-focus-badge\s*\{[\s\S]*border-radius:\s*0 0 4px 4px;/)
})
