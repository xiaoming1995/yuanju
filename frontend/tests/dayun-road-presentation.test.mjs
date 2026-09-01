import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import test from 'node:test'
import assert from 'node:assert/strict'

const root = new URL('..', import.meta.url).pathname
function read(path) {
  return readFileSync(join(root, path), 'utf8')
}

const {
  DAYUN_ROAD_SCOPE_EXPLANATION,
  getDayunPhasePrompt,
  getDayunPhaseRoadLabel,
  getDayunPhaseTheme,
  getJinBuHuanRating,
} = await import('../src/lib/dayunRoadPresentation.ts')

const period = { gan: '癸', zhi: '巳', start_year: 2024, end_year: 2033 }

test('phase prompts preserve Jin Bu Huan ratings and their five-year time scopes', () => {
  const front = getDayunPhasePrompt(period, { score: 35, detail: '癸为调候忌神天干。' }, 'front')
  const back = getDayunPhasePrompt(period, { score: 35, detail: '巳为金不换忌神地支。' }, 'back')

  assert.deepEqual(
    { phaseLabel: front?.phaseLabel, timeRange: front?.timeRange, roadLabel: front?.roadLabel, theme: front?.theme, governingLabel: front?.governingLabel, rating: front?.rating },
    { phaseLabel: '前五年', timeRange: '2024-2028', roadLabel: '施工路段', theme: '修整蓄力', governingLabel: '天干癸主事', rating: '凶' },
  )
  assert.deepEqual(
    { phaseLabel: back?.phaseLabel, timeRange: back?.timeRange, roadLabel: back?.roadLabel, theme: back?.theme, governingLabel: back?.governingLabel, rating: back?.rating },
    { phaseLabel: '后五年', timeRange: '2029-2033', roadLabel: '施工路段', theme: '修整蓄力', governingLabel: '地支巳主事', rating: '凶' },
  )
})

test('phase road labels provide themes and retain old-data fallbacks', () => {
  assert.equal(getJinBuHuanRating({ label: '施工路段' }), '凶')
  assert.equal(getJinBuHuanRating({ label: '山路' }), '平')
  assert.equal(getJinBuHuanRating({ label: '高速路' }), '吉')
  assert.equal(getDayunPhaseRoadLabel({ score: 35 }), '施工路段')
  assert.equal(getDayunPhaseRoadLabel({ score: 58 }), '山路')
  assert.equal(getDayunPhaseTheme('施工路段'), '修整蓄力')
  assert.equal(getDayunPhaseTheme('城市主路'), '稳步落地')
  assert.match(DAYUN_ROAD_SCOPE_EXPLANATION, /十年综合路况/)
  assert.match(DAYUN_ROAD_SCOPE_EXPLANATION, /金不换/)
})

test('result and timeline lead phase guidance with road and theme', () => {
  const resultPage = read('src/pages/ResultPage.tsx')
  const timeline = read('src/components/DayunTimeline.tsx')

  assert.match(resultPage, /十年综合路况/)
  assert.match(resultPage, /前后五年阶段指引/)
  assert.match(resultPage, /主题：\{prompt\.theme\}/)
  assert.match(resultPage, /查看路况与阶段提示说明/)
  assert.match(resultPage, /DAYUN_ROAD_SCOPE_EXPLANATION/)
  assert.doesNotMatch(resultPage, /前五年：\{currentDayunRoad\.qian_road\.label\}/)

  assert.match(timeline, /十年综合路况/)
  assert.match(timeline, /前后五年阶段指引/)
  assert.match(timeline, /主题：\{prompt\.theme\}/)
  assert.match(timeline, /getDayunPhasePrompt/)
  assert.doesNotMatch(timeline, /前后五年：\{activeRoad\.qian_road\.label\}/)
})
