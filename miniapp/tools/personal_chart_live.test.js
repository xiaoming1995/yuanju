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

test('bound user sees server daily data outside the demo calendar years', async () => {
  config.apiBaseUrl = 'https://api.test.example'
  const app = { globalData: { previewDayProfileEnabled: false } }
  const calls = []
  const wxApi = {
    getStorageSync: () => 'jwt',
    request(options) {
      calls.push(options.url)
      const data = options.url.includes('/personal-chart')
        ? { chart: { id: 'owned-chart', display_name: '自己的排盘' } }
        : options.url.includes('/calendar/day?')
          ? { data: { date: '2028-01-01', yi: ['祭祀'], ji: ['开市'] } }
          : { data: { date: '2028-01-01', day_gan_zhi: '甲子', status: 'mixed', label: '关系复杂', summary: '查看依据', evidence: [], activities: [] } }
      options.success({ statusCode: 200, data })
    },
  }
  const calendar = page('../pages/calendar/index', app, wxApi)
  await calendar.refreshPersonalState()
  assert.equal(calendar.data.personalChart.id, 'owned-chart')
  calendar.pickDate({ currentTarget: { dataset: { key: '2028-01-01' } } })
  await calendar.refreshRealDay()
  assert.equal(calendar.data.realDay.date, '2028-01-01')
  assert.equal(calendar.data.selected.dayGanZhi, '甲子')
  assert.ok(calls.some(url => url.endsWith('/personal-day?date=2028-01-01')))
})

test('authenticated unbound user does not see fixed sample as personal result', async () => {
  config.apiBaseUrl = 'https://api.test.example'
  const app = { globalData: { previewDayProfileEnabled: false } }
  const wxApi = {
    getStorageSync: () => 'jwt',
    request(options) { options.success({ statusCode: 200, data: { chart: null } }) },
  }
  const calendar = page('../pages/calendar/index', app, wxApi)
  await calendar.refreshPersonalState()
  assert.equal(calendar.data.personalChart, null)
  assert.equal(calendar.data.realDay, null)
  assert.equal(calendar.data.dayPreviewEnabled, false)
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
