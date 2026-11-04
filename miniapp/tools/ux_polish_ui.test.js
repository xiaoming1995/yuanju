const assert = require('node:assert/strict')
const test = require('node:test')

function loadApp() {
  let app
  global.App = value => { app = value }
  delete require.cache[require.resolve('../app')]
  require('../app')
  return app
}

function loadPage(path, app, wxApi = {}) {
  let definition
  global.Page = value => { definition = value }
  global.getApp = () => app
  global.wx = wxApi
  delete require.cache[require.resolve(path)]
  require(path)
  return {
    ...definition,
    data: { ...definition.data },
    setData(patch, callback) {
      this.data = { ...this.data, ...patch }
      if (callback) callback()
    },
  }
}

test('personal chart setup keeps hour precision', () => {
  const app = loadApp()
  const page = loadPage('../pages/personal-chart-setup/index', app)
  page.changeDate({ detail: { value: '1990-01-01' } })
  page.changeHour({ detail: { value: 1 } })
  assert.equal(page.data.birthDate, '1990-01-01')
  assert.equal(page.data.hours[page.data.hourIndex], '01:00')
})

test('long date groups reveal five days at a time and reset for another activity', async () => {
  require('../config').apiBaseUrl = 'https://api.test.example'
  const app = { globalData: { previewDayProfileEnabled: true } }
  const page = loadPage('../pages/moving-date/index', app, { getStorageSync() { return '' }, request(options) {
    const params = new URL(options.url).searchParams
    const days = []
    for (let time = Date.parse(`${params.get('start')}T00:00:00Z`); time <= Date.parse(`${params.get('end')}T00:00:00Z`); time += 86400000) {
      days.push({ date: new Date(time).toISOString().slice(0, 10), lunar_full: '十月初一', day_gan_zhi: '甲子', yi: ['入宅'], ji: [] })
    }
    options.success({ statusCode: 200, data: { data: { days } } })
  } })
  await page.onLoad()
  page.setData({ startDate: '2026-10-22', endDate: '2026-12-22' })
  await page.refreshResults()
  const longSection = page.data.sections.find(section => section.days.length > 5)
  assert.ok(longSection)
  assert.equal(longSection.visibleDays.length, 5)
  page.showMore({ currentTarget: { dataset: { section: longSection.id } } })
  assert.equal(page.data.sections.find(section => section.id === longSection.id).visibleDays.length, 10)
  while (page.data.sections.find(section => section.id === longSection.id).hasMore) {
    page.showMore({ currentTarget: { dataset: { section: longSection.id } } })
  }
  const expanded = page.data.sections.find(section => section.id === longSection.id)
  assert.equal(expanded.visibleDays.length, expanded.days.length)
  await page.changeActivity({ currentTarget: { dataset: { activity: 'marriage' } } })
  assert.ok(page.data.sections.every(section => section.visibleDays.length <= 5))
  assert.equal(page.data.activeSectionId, 'yi')
  assert.equal(page.data.total, 62)
})

test('removing real history calls backend and refreshes list', async () => {
  require('../config').apiBaseUrl = 'https://api.test.example'
  const app = loadApp()
  const calls = []
  const chart = id => ({ id, birth_year: 2000, birth_month: 1, birth_day: 1, birth_hour: 1,
    year_gan: '甲', year_zhi: '子', month_gan: '乙', month_zhi: '丑', day_gan: '丙', day_zhi: '寅', hour_gan: '丁', hour_zhi: '卯' })
  let charts = [chart('first'), chart('second')]
  const home = loadPage('../pages/home/index', app, {
    setNavigationBarTitle() {},
    showModal(options) { options.success({ confirm: true }) },
    showToast() {},
    getStorageSync() { return 'test-token' },
    request(options) {
      calls.push(`${options.method} ${options.url}`)
      if (options.method === 'DELETE') { charts = charts.filter(item => item.id !== 'first'); options.success({ statusCode: 200, data: {} }) }
      else options.success({ statusCode: 200, data: { charts, total: charts.length } })
    },
  })
  await home.loadHistory(1)
  assert.equal(home.data.history.length, 2)
  home.removeHistory({ currentTarget: { dataset: { id: 'first' } } })
  await new Promise(resolve => setImmediate(resolve))
  assert.equal(home.data.history.length, 1)
  assert.equal(home.data.recentHistory, null)
  assert.ok(calls.some(call => call.includes('DELETE') && call.includes('/bazi/history/first')))
})
