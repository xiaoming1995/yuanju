import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import test from 'node:test'
import assert from 'node:assert/strict'

const root = new URL('..', import.meta.url).pathname
function read(path) {
  return readFileSync(join(root, path), 'utf8')
}

const { buildDayunOverview } = await import('../src/lib/dayunOverview.ts')

test('DayunTimeline passes natal Fuyi strength to the overview builder', () => {
  const timeline = read('src/components/DayunTimeline.tsx')
  const resultPage = read('src/pages/ResultPage.tsx')

  assert.match(timeline, /natalDayMasterStrength/)
  assert.doesNotMatch(timeline, /resolveDayStrength|dayGanWuxing/)
  assert.match(resultPage, /<DayunTimeline[\s\S]+?natalDayMasterStrength=/)
})

test('overview does not infer Day Master strength from five-element percentages', () => {
  const overview = read('src/lib/dayunOverview.ts')
  assert.doesNotMatch(overview, /resolveDayStrength|HELP_MAP|dayGanWuxing/)
  assert.doesNotMatch(overview, /身弱遭杀克身/)
})

test('natal-strong 壬日主的戊申运 separates favorable stem and adverse longsheng branch', () => {
  const o = buildDayunOverview({
    dayun: { gan: '戊', zhi: '申', gan_shishen: '七杀', zhi_shishen: '偏印', di_shi: '长生' },
    yongshen: '土火木',
    jishen: '金水',
    natalDayMasterStrength: 'vstrong',
    tiaohou: null,
  })

  assert.match(o.prose, /原局身极强/)
  assert.match(o.prose, /戊土为扶抑喜用/)
  assert.match(o.prose, /申金为扶抑忌/)
  assert.match(o.prose, /申长生，忌神力量较显/)
  assert.doesNotMatch(o.prose, /身弱遭杀克身/)
  assert.equal(o.ganPolarity, 'xi')
  assert.equal(o.zhiPolarity, 'ji')
})

test('Dayun branch longsheng amplifies favorable force without changing natal context', () => {
  const o = buildDayunOverview({
    dayun: { gan: '甲', zhi: '寅', gan_shishen: '七杀', zhi_shishen: '偏印', di_shi: '临官' },
    yongshen: '木',
    jishen: '金',
    natalDayMasterStrength: 'strong',
    tiaohou: null,
  })

  assert.match(o.prose, /原局身强/)
  assert.match(o.prose, /寅木为扶抑喜用/)
  assert.match(o.prose, /寅临官，喜用力量更易发挥/)
})

test('weak-stage adverse branch is described as constrained rather than favorable', () => {
  const o = buildDayunOverview({
    dayun: { gan: '戊', zhi: '申', gan_shishen: '七杀', zhi_shishen: '偏印', di_shi: '绝' },
    yongshen: '土',
    jishen: '金',
    natalDayMasterStrength: 'strong',
    tiaohou: null,
  })

  assert.match(o.prose, /申金为扶抑忌/)
  assert.match(o.prose, /申绝，忌神力量受限/)
  assert.doesNotMatch(o.prose, /喜用力量更易发挥/)
})

test('neutral branch receives no directional verdict from its Twelve Growth Stage', () => {
  const o = buildDayunOverview({
    dayun: { gan: '丙', zhi: '辰', gan_shishen: '食神', zhi_shishen: '偏财', di_shi: '墓' },
    yongshen: '火',
    jishen: '金',
    natalDayMasterStrength: 'neutral',
    tiaohou: null,
  })

  assert.match(o.prose, /辰土为扶抑中性/)
  assert.match(o.prose, /仍以原局扶抑喜忌为准/)
})

test('Dayun overview retains Tiaohou supplement wording', () => {
  const o = buildDayunOverview({
    dayun: { gan: '丁', zhi: '未', gan_shishen: '正印', zhi_shishen: '伤官', di_shi: '冠带' },
    yongshen: '丙火',
    jishen: '',
    natalDayMasterStrength: 'weak',
    tiaohou: { expected: ['丙', '丁'], tou: [], cang: [], text: '' },
  })
  assert.match(o.prose, /正补足命局所缺调候/)
})

test('legacy payload without natal assessment remains neutral and does not infer strength', () => {
  const o = buildDayunOverview({
    dayun: { gan: '壬', zhi: '午', gan_shishen: '七杀', zhi_shishen: '劫财', di_shi: '帝旺' },
    yongshen: '',
    jishen: '',
    tiaohou: null,
  })

  assert.match(o.prose, /^壬午运：/)
  assert.doesNotMatch(o.prose, /原局身强|原局身弱|身弱遭杀克身/)
})

test('dictionary miss uses the existing fallback', () => {
  const o = buildDayunOverview({
    dayun: { gan: 'X', zhi: 'Y', gan_shishen: '不存在', zhi_shishen: '不存在', di_shi: 'Z' },
    yongshen: '',
    jishen: '',
    tiaohou: null,
  })
  assert.equal(o.prose, '选择一段大运后查看该十年流年节奏。')
})

test('ordinary Dayun summary remains the default UI while expert copy is available', () => {
  const src = read('src/components/DayunTimeline.tsx')
  assert.match(src, /overview\.proseLay/)
  assert.match(src, /查看专业表述/)
  assert.match(src, /dayun-summary-toggle/)
})
