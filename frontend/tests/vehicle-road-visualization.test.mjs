import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import test from 'node:test'
import assert from 'node:assert/strict'

const root = new URL('..', import.meta.url).pathname
function read(path) {
  return readFileSync(join(root, path), 'utf8')
}

test('ResultPage declares and renders vehicle-road fields', () => {
  const src = read('src/pages/ResultPage.tsx')
  assert.match(src, /interface VehicleProfile/)
  assert.match(src, /interface DayunRoad/)
  assert.match(src, /vehicle_profile\?: VehicleProfile/)
  assert.match(src, /dayun_roadmap\?: DayunRoad\[\]/)
  assert.match(src, /命盘座驾/)
  assert.match(src, /当前路况/)
  assert.match(src, /result\.vehicle_profile &&/)
  assert.match(src, /findDayunRoad\(result\.dayun_roadmap,\s*currentDayun\)/)
})

test('ResultPage exposes professional evidence and fallback-safe rendering', () => {
  const src = read('src/pages/ResultPage.tsx')
  assert.match(src, /readingMode === 'professional'[\s\S]+?result\.vehicle_profile\.evidences/)
  assert.match(src, /const currentRoadEvidences = currentDayunRoad\?\.evidences \?\? \[\]/)
  assert.match(src, /readingMode === 'professional'[\s\S]+?currentRoadEvidences\.length > 0/)
  assert.match(src, /currentDayunRoad\?\.summary \|\|/)
  assert.match(src, /formatEvidenceDelta/)
})

test('ResultPage explains grade and road labels in ordinary-user language', () => {
  const src = read('src/pages/ResultPage.tsx')
  assert.match(src, /const VEHICLE_GRADE_GUIDE = \[/)
  assert.match(src, /协同型配置/)
  assert.match(src, /稳健型配置/)
  assert.match(src, /实用型配置/)
  assert.match(src, /特性型配置/)
  assert.match(src, /调校型配置/)
  assert.match(src, /const ROAD_GUIDE = \[/)
  assert.match(src, /大运路况说明/)
  assert.match(src, /车代表命盘的基础配置与驾驭难度/)
  assert.match(src, /currentRoadGuide\.summary/)
  assert.match(src, /VEHICLE_GRADE_TAGS/)
})

test('ResultPage renders the full grade scale without requiring expansion', () => {
  const src = read('src/pages/ResultPage.tsx')
  assert.match(src, /className="result-grade-guide"/)
  assert.match(src, />等级说明</)
  assert.match(src, /S 到 D 衡量的是命盘配置完整度与驾驭难度/)
  assert.match(src, /VEHICLE_GRADE_GUIDE\.map/)
  assert.match(src, /item\.grade === result\.vehicle_profile\?\.grade \? 'is-current'/)
  assert.match(src, /className="result-road-guide"/)
  assert.match(src, />大运路况说明</)
})

test('DayunTimeline receives roadmap and renders road labels and phases', () => {
  const src = read('src/components/DayunTimeline.tsx')
  assert.match(src, /dayunRoadmap\?: DayunRoad\[\]/)
  assert.match(src, /roadByDayunIndex/)
  assert.match(src, /dayun-road-badge/)
  assert.match(src, /activeRoad\.qian_road\.label/)
  assert.match(src, /activeRoad\.hou_road\.label/)
})

test('Result CSS includes compact vehicle and road treatments', () => {
  const css = read('src/pages/ResultPage.css')
  assert.match(css, /\.result-vehicle-card/)
  assert.match(css, /\.result-vehicle-meter/)
  assert.match(css, /\.result-road-phase-row/)
  assert.match(css, /\.result-profile-evidence/)
  assert.match(css, /\.result-grade-guide/)
  assert.match(css, /\.result-grade-guide li\.is-current/)
  assert.match(css, /\.result-road-guide/)
  assert.match(css, /\.result-profile-plain-note/)
  assert.match(css, /\.dayun-road-badge/)
  assert.match(css, /\.dayun-road-phase-strip/)
})
