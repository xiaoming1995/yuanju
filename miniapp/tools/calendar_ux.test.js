const assert = require('node:assert/strict')
const fs = require('node:fs')
const test = require('node:test')
const config = require('../config')

function loadPage(app, wxApi = {}) {
  let definition
  global.Page = value => { definition = value }
  global.getApp = () => app
  global.wx = wxApi
  const path = '../pages/calendar/index'
  delete require.cache[require.resolve(path)]
  require(path)
  return { ...definition, data: { ...definition.data }, setData(patch) { this.data = { ...this.data, ...patch } } }
}

test('calendar starts on China today, preserves a chosen date, and returns to today', () => {
  const now = Date.now
  Date.now = () => Date.parse('2026-11-01T16:30:00Z')
  try {
    const app = { globalData: { previewDayProfileEnabled: false } }
    const page = loadPage(app)
    page.onShow()
    assert.equal(page.data.selectedKey, '2026-11-02')
    assert.equal(page.data.cells.find(cell => cell.key === '2026-11-02').today, true)
    assert.equal(page.data.cells.find(cell => cell.key === '2026-11-02').selected, true)

    page.pickDate({ currentTarget: { dataset: { key: '2026-11-10' } } })
    page.onShow()
    assert.equal(page.data.selectedKey, '2026-11-10')
    assert.equal(page.data.cells.find(cell => cell.key === '2026-11-02').today, true)
    assert.equal(page.data.cells.find(cell => cell.key === '2026-11-02').selected, false)

    app.globalData.previewCalendarNext = '2026-10-24'
    page.onShow()
    assert.equal(page.data.selectedKey, '2026-10-24')
    assert.equal(app.globalData.previewCalendarNext, null)
    page.pickToday()
    assert.equal(page.data.selectedKey, '2026-11-02')
  } finally { Date.now = now }
})

test('calendar outside the bundled years keeps civil dates without inventing lunar data', () => {
  const now = Date.now
  Date.now = () => Date.parse('2028-01-01T01:00:00Z')
  try {
    const page = loadPage({ globalData: {} })
    page.onShow()
    assert.equal(page.data.selectedKey, '2028-01-01')
    assert.equal(page.data.selected.lunarFull, '暂无本地农历资料')
    assert.equal(page.data.selected.jianChu, '')
    assert.equal(page.data.cells.find(cell => cell.key === '2028-01-01').lunar, '—')
    page.changeMonth({ currentTarget: { dataset: { delta: -1 } } })
    assert.equal(page.data.selectedKey, '2027-12-01')
    page.pickToday()
    assert.equal(page.data.selectedKey, '2028-01-01')
  } finally { Date.now = now }
})

test('four, five, and six week months retain seven complete date columns', () => {
  const page = loadPage({ globalData: {} })
  for (const [date, count] of [['2026-02-01', 28], ['2026-11-01', 35], ['2026-05-01', 42]]) {
    page.showDate(date)
    assert.equal(page.data.cells.length, count)
    assert.equal(page.data.cells.length % 7, 0)
    assert.equal(page.data.cells.filter(cell => cell.selected).length, 1)
  }
})

test('public almanac retains complete Yi and Ji without login and ignores an older date response', async () => {
  config.apiBaseUrl = 'https://api.test.example'
  const pending = []
  const page = loadPage({ globalData: {} }, {
    getStorageSync: () => '',
    request(options) { pending.push(options) },
  })
  page.onShow()
  assert.equal(pending.length, 1)
  const first = pending[0]
  page.pickDate({ currentTarget: { dataset: { key: '2026-11-02' } } })
  const second = pending[1]
  second.success({ statusCode: 200, data: { data: { date: '2026-11-02', yi: ['祭祀', '沐浴', '解除', '破屋', '坏垣', '馀事勿取'], ji: ['开市', '嫁娶'] } } })
  await new Promise(resolve => setImmediate(resolve))
  first.success({ statusCode: 200, data: { data: { date: '2026-11-01', yi: ['旧宜'], ji: ['旧忌'] } } })
  await new Promise(resolve => setImmediate(resolve))
  assert.deepEqual(page.data.calendarDay.yi, ['祭祀', '沐浴', '解除', '破屋', '坏垣', '馀事勿取'])
  assert.deepEqual(page.data.calendarDay.ji, ['开市', '嫁娶'])
  page.pickDate({ currentTarget: { dataset: { key: '2026-11-03' } } })
  assert.equal(page.data.calendarDay, null)
})

test('calendar omits the selected-date card and keeps its detail link in date information', () => {
  const calendar = fs.readFileSync(require.resolve('../pages/calendar/index.wxml'), 'utf8')
  const detail = fs.readFileSync(require.resolve('../pages/personal-day/index.wxml'), 'utf8')
  assert.doesNotMatch(calendar, /selected-card/)
  assert.match(calendar, /date-detail-link[^>]*bindtap="openSelectedDay"/)
  assert.doesNotMatch(calendar, /wx:for="\{\{calendarDay\.(yi|ji)\}\}"/)
  assert.match(detail, /wx:for="\{\{calendarDay\.yi\}\}"/)
  assert.match(detail, /wx:for="\{\{calendarDay\.ji\}\}"/)
})

test('calendar distinguishes empty, loading, and failed almanac data', async () => {
  config.apiBaseUrl = 'https://api.test.example'
  const pending = []
  const page = loadPage({ globalData: {} }, {
    getStorageSync: () => '',
    request(options) { pending.push(options) },
  })
  const empty = page.refreshCalendarDay()
  assert.equal(page.data.calendarLoading, true)
  assert.equal(page.data.calendarDay, null)
  pending[0].success({ statusCode: 200, data: { data: { date: page.data.selectedKey, yi: [], ji: [] } } })
  await empty
  assert.equal(page.data.calendarLoading, false)
  assert.deepEqual(page.data.calendarDay.yi, [])
  assert.deepEqual(page.data.calendarDay.ji, [])
  const failed = page.refreshCalendarDay()
  pending[1].fail({ errMsg: '网络失败' })
  await failed
  assert.equal(page.data.calendarDay, null)
  assert.ok(page.data.calendarError)
})
