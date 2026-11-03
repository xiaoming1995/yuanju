const assert = require('node:assert/strict')
const test = require('node:test')
const config = require('../config')

function page(path, app, wxApi) {
  let definition
  global.Page = value => { definition = value }
  global.getApp = () => app
  global.wx = wxApi
  delete require.cache[require.resolve(path)]
  require(path)
  return { ...definition, data: { ...definition.data }, setData(patch) { this.data = { ...this.data, ...patch } } }
}

test('home shows same real today for general and personal data outside demo years', async () => {
  config.apiBaseUrl = 'https://api.test.example'
  const previousNow = Date.now
  Date.now = () => Date.parse('2027-12-31T16:30:00Z')
  try {
  const app = { globalData: { previewDayProfileEnabled: false } }
  const calls = []
  const wxApi = {
    getStorageSync: () => 'jwt',
    request(options) {
      calls.push(options.url)
      const data = options.url.includes('/personal-chart')
        ? { chart: { id: 'owned-chart', display_name: '自己的排盘' } }
        : options.url.includes('/calendar/day?')
          ? { data: { date: '2028-01-01', day_gan_zhi: '甲子', yi: ['祭祀'], ji: ['开市'] } }
          : { data: { date: '2028-01-01', day_gan_zhi: '甲子', status: 'mixed', label: '关系复杂', summary: '查看依据', evidence: [], activities: [] } }
      options.success({ statusCode: 200, data })
    },
  }
  const home = page('../pages/home/index', app, wxApi)
  home.refreshToday()
  await new Promise(resolve => setImmediate(resolve))
  await new Promise(resolve => setImmediate(resolve))
  assert.equal(home.data.calendarPreview.key, '2028-01-01')
  assert.equal(home.data.calendarDay.date, '2028-01-01')
  assert.equal(home.data.personalDay.date, '2028-01-01')
  assert.equal(home.data.calendarPreview.dayGanZhi, '甲子')
  assert.ok(calls.some(url => url.endsWith('/personal-day?date=2028-01-01')))
  } finally { Date.now = previousNow }
})

test('home keeps unbound state separate from public almanac', async () => {
  config.apiBaseUrl = 'https://api.test.example'
  const app = { globalData: { previewDayProfileEnabled: false } }
  const wxApi = {
    getStorageSync: () => 'jwt',
    request(options) {
      const date = options.url.split('date=')[1]
      options.success({ statusCode: 200, data: options.url.endsWith('/personal-chart')
        ? { chart: null } : { data: { date, yi: ['祭祀'], ji: ['开市'] } } })
    },
  }
  const home = page('../pages/home/index', app, wxApi)
  home.refreshToday()
  await new Promise(resolve => setImmediate(resolve))
  assert.deepEqual(home.data.calendarDay.yi, ['祭祀'])
  assert.equal(home.data.personalState, 'unbound')
  assert.equal(home.data.personalDay, null)
})

test('home and detail show the same personal items, directions, and luck', async () => {
  config.apiBaseUrl = 'https://api.test.example'
  const date = new Date(Date.now() + 8 * 60 * 60 * 1000).toISOString().slice(0, 10)
  const day = {
    date, day_gan: '壬', day_zhi: '午', day_gan_zhi: '壬午',
    stem_element: '水', branch_element: '火', stem_direction: 'favorable', branch_direction: 'adverse',
    status: 'mixed', label: '关系复杂', summary: '日干与日支主气方向相反，需分别看依据。',
    activity_guidance: [
      { id: 'travel', title: '旅行与重要出行', label: '慎', level: 'review', source_label: '通用日历 · 本人命盘', sources: ['calendar_yiji', 'personal_chart'], summary: '重要行程宜核对。' },
      { id: 'finance', title: '投资研究与财务规划', label: '慎', level: 'review', source_label: '本人命盘', sources: ['personal_chart'], summary: '先核对预算。' },
      { id: 'work', title: '工作事务', label: '宜', level: 'candidate', source_label: '本人命盘', sources: ['personal_chart'], summary: '宜安排工作事务。' },
    ],
    luck: { available: true, level: 'qualified', element: '水', direction: '北方', color: '蓝黑色系', reason: '今日水属喜用，但另一干支五行属忌神，仅为局部参考。' },
    evidence: [], activities: [],
  }
  const wxApi = {
    getStorageSync: () => 'jwt',
    request(options) {
      const data = options.url.endsWith('/personal-chart') ? { chart: { id: 'mine' } }
        : options.url.includes('/calendar/day?') ? { data: { date, yi: ['出行'], ji: [] } } : { data: day }
      options.success({ statusCode: 200, data })
    },
  }
  const app = { globalData: {} }
  const home = page('../pages/home/index', app, wxApi)
  home.refreshToday()
  const detail = page('../pages/personal-day/index', app, wxApi)
  await detail.loadRealDay(date)
  await new Promise(resolve => setImmediate(resolve))
  assert.deepEqual(home.data.personalGuidance, detail.data.personalHints)
  assert.equal(home.data.personalGuidance.length, 3)
  assert.deepEqual(home.data.personalDirections, detail.data.realDirections)
  assert.deepEqual(home.data.personalDay.luck, detail.data.realDay.luck)
  assert.equal(home.data.personalDay.luck.level, 'qualified')
  assert.equal(home.data.personalDay.summary, detail.data.realDay.summary)
})

test('home ignores an earlier personal response after the account changes', async () => {
  config.apiBaseUrl = 'https://api.test.example'
  let token = 'old-jwt'
  const pending = []
  const home = page('../pages/home/index', { globalData: {} }, {
    getStorageSync: () => token,
    request(options) {
      if (options.url.includes('/calendar/day?')) {
        const date = options.url.split('date=')[1]
        options.success({ statusCode: 200, data: { data: { date, yi: [], ji: [] } } })
      } else pending.push(options)
    },
  })
  home.refreshToday()
  assert.equal(pending.length, 1)
  token = ''
  home.refreshToday()
  pending[0].success({ statusCode: 200, data: { chart: { id: 'old-chart' } } })
  await new Promise(resolve => setImmediate(resolve))
  assert.equal(home.data.personalDay, null)
  assert.equal(home.data.personalState, 'unbound')
})

test('unbound date detail reads public real almanac without personal API', async () => {
  config.apiBaseUrl = 'https://api.test.example'
  const calls = []
  const detail = page('../pages/personal-day/index', { globalData: {} }, {
    getStorageSync: () => '',
    request(options) {
      calls.push(options.url)
      options.success({ statusCode: 200, data: { data: {
        date: '2026-11-02', lunar_full: '九月廿四', year_gan_zhi: '丙午', month_gan_zhi: '戊戌',
        day_gan_zhi: '庚辰', yi: ['祭祀', '沐浴'], ji: ['开市'],
      } } })
    },
  })
  detail.onLoad({ date: '2026-11-02' })
  await new Promise(resolve => setImmediate(resolve))
  assert.equal(detail.data.realMode, true)
  assert.deepEqual(detail.data.calendarDay.yi, ['祭祀', '沐浴'])
  assert.deepEqual(detail.data.calendarDay.ji, ['开市'])
  assert.ok(calls.every(url => url.includes('/calendar/day?')))
})

test('real daily detail keeps activity evidence and directional fields together', async () => {
  config.apiBaseUrl = 'https://api.test.example'
  const realDay = {
    date: '2026-11-02', day_gan: '庚', day_zhi: '辰', day_gan_zhi: '庚辰',
    stem_element: '金', branch_element: '土', stem_direction: 'favorable', branch_direction: 'neutral',
    status: 'undetermined', label: '日干偏合', summary: '日干庚（金）属喜用；日支辰（土主气）未落入喜忌。',
    calendar_yi: ['祭祀', '沐浴'], calendar_ji: ['开市'],
    activity_guidance: [
      { id: 'travel', level: 'insufficient', sources: ['calendar_yiji'], source_label: '通用日历', evidence: ['无出行标签'] },
      { id: 'finance', level: 'candidate', sources: ['personal_chart'], source_label: '本人命盘', evidence: ['喜用同向'] },
      { id: 'work', level: 'insufficient', sources: ['personal_chart'], source_label: '本人命盘', evidence: [] },
    ],
    evidence: [], activities: [],
  }
  const wxApi = {
    getStorageSync: () => 'jwt',
    request(options) {
      const data = options.url.endsWith('/personal-chart') ? { chart: { id: 'mine' } } : { data: realDay }
      options.success({ statusCode: 200, data })
    },
  }
  const detail = page('../pages/personal-day/index', { globalData: { previewDayProfileEnabled: false } }, wxApi)
  await detail.loadRealDay('2026-11-02')
  assert.equal(detail.data.realDay.activity_guidance[0].source_label, '通用日历')
  assert.deepEqual(detail.data.realDay.calendar_yi, ['祭祀', '沐浴'])
  assert.deepEqual(detail.data.personalHints.map(item => item.id), ['finance'])
  assert.deepEqual(detail.data.realDirections, [
    { name: '日干庚（金）', direction: '喜用' },
    { name: '日支辰（土主气）', direction: '未落入喜忌' },
  ])
  assert.equal(detail.data.realDay.label, '日干偏合')
  assert.equal(detail.data.showActivityEvidence, false)
  detail.toggleActivityEvidence()
  assert.equal(detail.data.showActivityEvidence, true)
})

test('history binding is explicit and can be cleared without deleting the chart', async () => {
  config.apiBaseUrl = 'https://api.test.example'
  const app = { globalData: { activeResult: null, personalChart: null } }
  let bound = false
  const calls = []
  const wxApi = {
    getStorageSync: () => 'jwt',
    showModal(options) { options.success({ confirm: true }) },
    showToast() {},
    request(options) {
      calls.push(`${options.method} ${options.url}`)
      if (options.method === 'PUT') bound = true
      if (options.method === 'DELETE') bound = false
      options.success({ statusCode: 200, data: { chart: bound ? { id: 'mine' } : null } })
    },
  }
  const home = page('../pages/home/index', app, wxApi)
  assert.equal(home.data.personalChartId, '')
  home.bindPersonalChart({ currentTarget: { dataset: { id: 'mine' } } })
  await new Promise(resolve => setImmediate(resolve))
  assert.equal(home.data.personalChartId, 'mine')
  home.unbindPersonalChart()
  await new Promise(resolve => setImmediate(resolve))
  assert.equal(home.data.personalChartId, '')
  assert.ok(calls.some(call => call.includes('PUT https://api.test.example/api/user/personal-chart')))
  assert.ok(calls.some(call => call.includes('DELETE https://api.test.example/api/user/personal-chart')))
  assert.ok(!calls.some(call => call.includes('/bazi/history/')))
})
