import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import test from 'node:test'
import assert from 'node:assert/strict'

const root = new URL('..', import.meta.url).pathname
function read(path) {
  return readFileSync(join(root, path), 'utf8')
}

// ─── 静态接线测试 ───────────────────────────────────────────────────────

test('DayunTimeline imports buildDayunOverview from lib', () => {
  const src = read('src/components/DayunTimeline.tsx')
  assert.match(src, /import\s+\{\s*buildDayunOverview/)
})

test('DayunTimeline no longer hardcodes 宜先看节奏', () => {
  const src = read('src/components/DayunTimeline.tsx')
  assert.doesNotMatch(src, /宜先看节奏/)
})

test('ResultPage passes yongshen/jishen/wuxing/tiaohou to DayunTimeline', () => {
  const src = read('src/pages/ResultPage.tsx')
  assert.match(src, /<DayunTimeline[\s\S]+?yongshen=/)
  assert.match(src, /<DayunTimeline[\s\S]+?jishen=/)
  assert.match(src, /<DayunTimeline[\s\S]+?wuxing=/)
  assert.match(src, /<DayunTimeline[\s\S]+?tiaohou=/)
})

// ─── 行为单测: buildDayunOverview ──────────────────────────────────────
// 通过 dynamic import 加载 .ts 源（依赖 --experimental-strip-types 启动）

const { buildDayunOverview } = await import('../src/lib/dayunOverview.ts')

test('截图原例: 壬午身弱忌杀盖头缺火', () => {
  const o = buildDayunOverview({
    dayun: { gan: '壬', zhi: '午', gan_shishen: '七杀', zhi_shishen: '劫财', di_shi: '帝旺' },
    yongshen: '丙火',
    jishen: '壬癸水',
    wuxing: { mu: 10, huo: 15, tu: 10, jin: 25, shui: 40 },
    dayGanWuxing: '火',
    tiaohou: { expected: ['丙', '丁'], tou: [], cang: [], text: '冬月需丙丁火调候' },
  })
  assert.match(o.prose, /七杀为忌/)
  assert.match(o.prose, /身弱遭杀克身/)
  assert.match(o.prose, /但被壬盖头压制/)
  assert.match(o.prose, /调候未到位/)
  assert.equal(o.ganPolarity, 'ji')
})

test('身旺喜杀: 七杀 chip 取 xi', () => {
  const o = buildDayunOverview({
    dayun: { gan: '丙', zhi: '寅', gan_shishen: '七杀', zhi_shishen: '偏印', di_shi: '长生' },
    yongshen: '丙火',
    jishen: '壬癸水',
    wuxing: { mu: 15, huo: 25, tu: 20, jin: 30, shui: 10 },
    dayGanWuxing: '金',
    tiaohou: null,
  })
  assert.match(o.prose, /立威破局/)
  assert.match(o.trendKeywords, /突破/)
  assert.equal(o.ganPolarity, 'xi')
})

test('通根: 甲寅 (七杀)', () => {
  const o = buildDayunOverview({
    dayun: { gan: '甲', zhi: '寅', gan_shishen: '七杀', zhi_shishen: '七杀', di_shi: '临官' },
    yongshen: '',
    jishen: '',
    wuxing: { mu: 20, huo: 20, tu: 20, jin: 20, shui: 20 },
    dayGanWuxing: '土',
    tiaohou: null,
  })
  assert.match(o.prose, /甲通根寅得力/)
})

test('截脚: 丙子 (正官)', () => {
  const o = buildDayunOverview({
    dayun: { gan: '丙', zhi: '子', gan_shishen: '正官', zhi_shishen: '正官', di_shi: '胎' },
    yongshen: '',
    jishen: '',
    wuxing: { mu: 20, huo: 20, tu: 20, jin: 20, shui: 20 },
    dayGanWuxing: '土',
    tiaohou: null,
  })
  assert.match(o.prose, /反被子截脚虚浮/)
})

test('调候补足: 缺火, 丁未运不被克', () => {
  const o = buildDayunOverview({
    dayun: { gan: '丁', zhi: '未', gan_shishen: '正印', zhi_shishen: '伤官', di_shi: '冠带' },
    yongshen: '丙火',
    jishen: '',
    wuxing: { mu: 30, huo: 5, tu: 30, jin: 20, shui: 15 },
    dayGanWuxing: '金',
    tiaohou: { expected: ['丙', '丁'], tou: [], cang: [], text: '' },
  })
  assert.match(o.prose, /正补足命局所缺调候/)
})

test('调候未及: 缺火, 庚申运完全无火', () => {
  const o = buildDayunOverview({
    dayun: { gan: '庚', zhi: '申', gan_shishen: '比肩', zhi_shishen: '比肩', di_shi: '临官' },
    yongshen: '',
    jishen: '',
    wuxing: { mu: 30, huo: 5, tu: 30, jin: 20, shui: 15 },
    dayGanWuxing: '金',
    tiaohou: { expected: ['丙', '丁'], tou: [], cang: [], text: '' },
  })
  assert.match(o.prose, /未在此运补足/)
})

test('yongshen/jishen 都空时标题简化, BODY1 仍出', () => {
  const o = buildDayunOverview({
    dayun: { gan: '壬', zhi: '午', gan_shishen: '七杀', zhi_shishen: '劫财', di_shi: '帝旺' },
    yongshen: '',
    jishen: '',
    wuxing: { mu: 10, huo: 15, tu: 10, jin: 25, shui: 40 },
    dayGanWuxing: '火',
    tiaohou: null,
  })
  assert.doesNotMatch(o.prose, /（七杀为/)
  assert.match(o.prose, /^壬午运：/)
  assert.match(o.prose, /身弱遭杀克身/)
})

test('字典 miss → fallback', () => {
  const o = buildDayunOverview({
    dayun: { gan: 'X', zhi: 'Y', gan_shishen: '不存在', zhi_shishen: '不存在', di_shi: 'Z' },
    yongshen: '',
    jishen: '',
    wuxing: { mu: 20, huo: 20, tu: 20, jin: 20, shui: 20 },
    dayGanWuxing: '木',
    tiaohou: null,
  })
  assert.equal(o.prose, '选择一段大运后查看该十年流年节奏。')
  assert.equal(o.proseLay, '选择一段大运后查看该十年流年节奏。')
  assert.equal(o.trendKeywords, '节奏 · 观察 · 平衡')
})

// ─── proseLay 大白话版 ─────────────────────────────────────────────────

test('proseLay: 壬午身弱忌杀盖头缺火 — 大白话不出现专业术语', () => {
  const o = buildDayunOverview({
    dayun: { gan: '壬', zhi: '午', gan_shishen: '七杀', zhi_shishen: '劫财', di_shi: '帝旺' },
    yongshen: '丙火',
    jishen: '壬癸水',
    wuxing: { mu: 10, huo: 15, tu: 10, jin: 25, shui: 40 },
    dayGanWuxing: '火',
    tiaohou: { expected: ['丙', '丁'], tou: [], cang: [], text: '冬月需丙丁火调候' },
  })
  assert.match(o.proseLay, /压力偏大的十年/)
  assert.match(o.proseLay, /外部压力较大、突发事件多/)
  assert.match(o.proseLay, /地支有支撑但被天干压住/)
  assert.match(o.proseLay, /火气（热情 \/ 行动）/)
  assert.match(o.proseLay, /效果打折扣/)
  assert.doesNotMatch(o.proseLay, /七杀|盖头|帝旺|调候/)
})

test('proseLay: 喜杀身旺通根 — heading 用 xi 基调', () => {
  const o = buildDayunOverview({
    dayun: { gan: '甲', zhi: '寅', gan_shishen: '七杀', zhi_shishen: '七杀', di_shi: '临官' },
    yongshen: '甲乙木',
    jishen: '',
    wuxing: { mu: 20, huo: 20, tu: 30, jin: 15, shui: 15 },
    dayGanWuxing: '土',
    tiaohou: null,
  })
  assert.match(o.proseLay, /适合主动出击的十年/)
  assert.match(o.proseLay, /适合主动出击、立威破局/)
  assert.match(o.proseLay, /天干和地支力量一致/)
})

test('proseLay: 调候补足 — 用"补上，体感会比较顺"', () => {
  const o = buildDayunOverview({
    dayun: { gan: '丁', zhi: '未', gan_shishen: '正印', zhi_shishen: '伤官', di_shi: '冠带' },
    yongshen: '丙火',
    jishen: '',
    wuxing: { mu: 30, huo: 5, tu: 30, jin: 20, shui: 15 },
    dayGanWuxing: '金',
    tiaohou: { expected: ['丙', '丁'], tou: [], cang: [], text: '' },
  })
  assert.match(o.proseLay, /学习与贵人的十年/)
  assert.match(o.proseLay, /火气（热情 \/ 行动）/)
  assert.match(o.proseLay, /在这十年补上，体感会比较顺/)
})

test('proseLay: yongshen/jishen 空时去掉括号基调', () => {
  const o = buildDayunOverview({
    dayun: { gan: '壬', zhi: '午', gan_shishen: '七杀', zhi_shishen: '劫财', di_shi: '帝旺' },
    yongshen: '',
    jishen: '',
    wuxing: { mu: 10, huo: 15, tu: 10, jin: 25, shui: 40 },
    dayGanWuxing: '火',
    tiaohou: null,
  })
  assert.match(o.proseLay, /^壬午运：/)
  assert.doesNotMatch(o.proseLay, /（.*的十年/)
})

// ─── UI 接线 ───────────────────────────────────────────────────────────

test('DayunTimeline renders proseLay as default and has toggle to expert', () => {
  const src = read('src/components/DayunTimeline.tsx')
  assert.match(src, /overview\.proseLay/)
  assert.match(src, /查看专业表述/)
  assert.match(src, /dayun-summary-toggle/)
})
