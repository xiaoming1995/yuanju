const assert = require('node:assert/strict')
const test = require('node:test')
require('../config').apiBaseUrl = 'https://api.test.example'

function loadPage(path, wxApi, app = { globalData: {} }) {
  let definition
  global.Page = value => { definition = value }
  global.getApp = () => app
  global.wx = wxApi
  delete require.cache[require.resolve(path)]
  require(path)
  return {
    ...definition, data: { ...definition.data },
    setData(patch, callback) { this.data = { ...this.data, ...patch }; if (callback) callback() },
  }
}

function daysFromUrl(url) {
  const params = new URL(url).searchParams
  const start = new Date(`${params.get('start')}T00:00:00Z`)
  const end = new Date(`${params.get('end')}T00:00:00Z`)
  const days = []
  for (let date = start; date <= end; date = new Date(date.getTime() + 86400000)) {
    days.push({ date: date.toISOString().slice(0, 10), lunar_full: '十月初一', day_gan_zhi: '甲子',
      yi: ['入宅', '嫁娶', '出行'], ji: ['开市'] })
  }
  return days
}

function environment({ token = '', chart = null, delayed = false, failure = false } = {}) {
  const calls = []
  const pending = []
  const wxApi = {
    getStorageSync() { return token },
    navigateTo(options) { calls.push(options.url) },
    request(options) {
      if (delayed) { pending.push(options); return }
      if (failure && options.url.includes('/calendar/range')) {
        options.fail(); return
      }
      const data = options.url.includes('/calendar/range') ? { data: { days: daysFromUrl(options.url) } }
        : options.url.includes('/personal-chart') ? { chart } : {}
      options.success({ statusCode: 200, data })
    },
  }
  return { wxApi, calls, pending }
}

test('calendar opens the dedicated date selection page', () => {
  const { wxApi, calls } = environment()
  const calendar = loadPage('../pages/calendar/index', wxApi)
  calendar.openMovingDate()
  assert.deepEqual(calls, ['/pages/moving-date/index'])
  const target = loadPage('../pages/moving-date/index', wxApi)
  target.onLoad()
  target.onReady()
})

test('unbound users see real general dates without a sample switch', async () => {
  const { wxApi } = environment()
  const page = loadPage('../pages/moving-date/index', wxApi)
  await page.onLoad()
  assert.equal(page.data.total, 30)
  assert.equal(page.data.candidateCount, 30)
  assert.equal(page.data.personalChart, null)
  assert.equal(typeof page.startPreview, 'undefined')
  assert.equal(page.data.sections[0].days[0].allYi[0], '入宅')
  assert.equal(page.data.customRangeVisible, false)
  assert.equal(page.data.activeSectionId, 'yi')
  assert.equal(page.data.sections[0].visibleDays.length, 5)
})

test('custom range can be collapsed without changing the selected dates', async () => {
  const { wxApi } = environment()
  const page = loadPage('../pages/moving-date/index', wxApi)
  await page.onLoad()
  page.toggleCustomRange()
  assert.equal(page.data.customRangeVisible, true)
  await page.changeEndDate({ detail: { value: page.data.startDate } })
  assert.equal(page.data.total, 1)
  page.toggleCustomRange()
  assert.equal(page.data.customRangeVisible, false)
  assert.equal(page.data.endDate, page.data.startDate)
  assert.equal(page.data.total, 1)
})

test('result tabs switch existing groups without a request and clear the open date', async () => {
  const { wxApi } = environment()
  let requests = 0
  const originalRequest = wxApi.request
  wxApi.request = options => { requests += 1; originalRequest(options) }
  const page = loadPage('../pages/moving-date/index', wxApi)
  await page.onLoad()
  await page.changeActivity({ currentTarget: { dataset: { activity: 'opening' } } })
  assert.equal(page.data.activeSectionId, 'yi')
  assert.equal(page.data.candidateCount, 0)
  assert.equal(page.data.sections.find(section => section.id === 'ji').days.length, 30)
  const beforeSwitch = requests
  page.changeSection({ currentTarget: { dataset: { section: 'ji' } } })
  assert.equal(requests, beforeSwitch)
  assert.equal(page.data.activeSectionId, 'ji')
  const date = page.data.sections.find(section => section.id === 'ji').days[0].date
  page.openDay({ currentTarget: { dataset: { date } } })
  assert.equal(page.data.expandedDate, date)
  page.changeSection({ currentTarget: { dataset: { section: 'unlisted' } } })
  assert.equal(page.data.expandedDate, '')
  assert.equal(requests, beforeSwitch)
})

test('changing activity recomputes general results while keeping range', async () => {
  const { wxApi } = environment()
  const page = loadPage('../pages/moving-date/index', wxApi)
  await page.onLoad()
  const range = [page.data.startDate, page.data.endDate]
  await page.changeActivity({ currentTarget: { dataset: { activity: 'opening' } } })
  assert.deepEqual([page.data.startDate, page.data.endDate], range)
  assert.equal(page.data.activityId, 'opening')
  assert.equal(page.data.candidateCount, 0)
  assert.equal(page.data.sections.find(section => section.id === 'ji').days.length, 30)
  assert.equal(page.data.activityOptions.filter(item => item.selected).length, 1)
})

test('invalid range clears candidates and later service failure has a retry path', async () => {
  const { wxApi } = environment()
  const page = loadPage('../pages/moving-date/index', wxApi)
  await page.onLoad()
  const invalidStart = new Date(`${page.data.endDate}T00:00:00Z`)
  invalidStart.setUTCDate(invalidStart.getUTCDate() + 1)
  await page.changeStartDate({ detail: { value: invalidStart.toISOString().slice(0, 10) } })
  assert.match(page.data.rangeError, /起日/)
  assert.equal(page.data.total, 0)
  const oldRequest = wxApi.request
  wxApi.request = options => options.fail()
  await page.changeStartDate({ detail: { value: page.data.endDate } })
  assert.match(page.data.resultError, /网络/)
  assert.equal(page.data.total, 0)
  wxApi.request = oldRequest
  await page.retryResults()
  assert.equal(page.data.resultError, '')
  assert.equal(page.data.total, 1)
})

test('an older response cannot replace a newer range or activity', async () => {
  const { wxApi, pending } = environment({ delayed: true })
  const page = loadPage('../pages/moving-date/index', wxApi)
  page.onLoad()
  const first = pending[0]
  page.setData({ startDate: '2028-01-01', endDate: '2028-01-02' })
  const newer = page.changeActivity({ currentTarget: { dataset: { activity: 'marriage' } } })
  pending[1].success({ statusCode: 200, data: { data: { days: daysFromUrl(pending[1].url) } } })
  await newer
  first.success({ statusCode: 200, data: { data: { days: daysFromUrl(first.url) } } })
  await Promise.resolve()
  assert.equal(page.data.activityId, 'marriage')
  assert.equal(page.data.total, 2)
  assert.equal(page.data.sections[0].days[0].date, '2028-01-01')
})

test('date detail uses the same full labels and bound user can open personal day', async () => {
  const { wxApi, calls } = environment({ token: 'token', chart: { id: 'chart-1' } })
  const page = loadPage('../pages/moving-date/index', wxApi)
  await page.onLoad()
  await page.onShow()
  const date = page.data.sections[0].days[0].date
  page.openDay({ currentTarget: { dataset: { date } } })
  assert.equal(page.data.expandedDate, date)
  assert.deepEqual(page.data.sections[0].days[0].allYi, ['入宅', '嫁娶', '出行'])
  page.openPersonalDay({ currentTarget: { dataset: { date } } })
  assert.deepEqual(calls, [`/pages/personal-day/index?date=${date}`])
  const detail = loadPage('../pages/personal-day/index', wxApi)
  detail.onLoad({ date })
  detail.onReady()
  page.onShow()
  assert.equal(page.data.expandedDate, date)
  assert.equal(page.data.total, 30)
})
