const chart = require('./demo_chart')
const pillars = chart.pillars.map(pillar => ({
  label: pillar.label.slice(0, 1),
  gan: pillar.gan,
  zhi: pillar.zhi,
  ganTone: pillar.ganTone,
  zhiTone: pillar.zhiTone,
}))

const basicRowFields = [
  { label: '主星', field: 'ganShiShen', kind: 'text' },
  { label: '天干', field: 'gan', kind: 'main', wuxing: 'ganWuxing', tone: 'ganTone' },
  { label: '地支', field: 'zhi', kind: 'main', wuxing: 'zhiWuxing', tone: 'zhiTone' },
  { label: '藏干', field: 'hideGan', kind: 'list' },
  { label: '副星', field: 'zhiShiShen', kind: 'list' },
  { label: '星运', field: 'xingYun', kind: 'text' },
  { label: '自坐', field: 'diShi', kind: 'text' },
  { label: '空亡', field: 'xunKong', kind: 'text' },
  { label: '纳音', field: 'naYin', kind: 'text' },
  { label: '神煞', field: 'shenSha', kind: 'tags' },
]

const basicRows = basicRowFields.map(row => ({
  label: row.label,
  kind: row.kind,
  cells: chart.pillars.map(pillar => {
    const value = pillar[row.field]
    return {
      pillar: pillar.label,
      text: Array.isArray(value) ? value.join(' ') : value,
      items: Array.isArray(value) ? value : [],
      wuxing: row.wuxing ? pillar[row.wuxing] : '',
      tone: row.tone ? pillar[row.tone] : '',
    }
  }),
}))

const history = [
  {
    id: 'self', name: '我的命盘', date: '1990-01-01 01:00',
    columns: [
      { text: '己巳', tone: 'earth' }, { text: '丙子', tone: 'water' },
      { text: '丙寅', tone: 'fire' }, { text: '己丑', tone: 'earth' },
    ],
  },
  {
    id: 'child', name: '孩子', date: '2018-05-12 10:30',
    columns: [
      { text: '戊戌', tone: 'earth' }, { text: '丁巳', tone: 'fire' },
      { text: '壬辰', tone: 'water' }, { text: '乙巳', tone: 'earth' },
    ],
  },
  {
    id: 'friend', name: '朋友', date: '2000-08-25 16:00',
    columns: [
      { text: '庚辰', tone: 'earth' }, { text: '甲申', tone: 'water' },
      { text: '丙午', tone: 'fire' }, { text: '丙申', tone: 'earth' },
    ],
  },
  {
    id: 'test', name: '测试', date: '2023-03-01 09:20',
    columns: [
      { text: '癸卯', tone: 'wood' }, { text: '甲寅', tone: 'water' },
      { text: '己亥', tone: 'earth' }, { text: '己巳', tone: 'fire' },
    ],
  },
]

module.exports = { chart, pillars, basicRows, history }
