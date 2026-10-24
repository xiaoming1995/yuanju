const assert = require('node:assert/strict')
const test = require('node:test')

function loadPage(path, app, wxApi) {
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

function environment() {
  const app = { globalData: {} }
  const calls = []
  const wxApi = {
    navigateTo(options) { calls.push({ method: 'navigateTo', options }) },
    switchTab(options) { calls.push({ method: 'switchTab', options }) },
    navigateBack(options) { calls.push({ method: 'navigateBack', options }) },
    setNavigationBarTitle() {},
    showToast() {},
  }
  return { app, calls, wxApi }
}

test('home result and back navigation acknowledge only one tap and restore the form', () => {
  const { app, calls, wxApi } = environment()
  const home = loadPage('../pages/home/index', app, wxApi)
  home.onShow()
  home.onReady()
  home.setData({ dateText: '2000-06-01 08:00', note: '自己的命盘' })

  home.openDemoResult()
  home.openDemoResult()
  assert.equal(calls.filter(call => call.method === 'navigateTo').length, 1)
  assert.equal(home.data.navigationKey, 'result')
  home.onHide()

  const result = loadPage('../pages/result/index', app, wxApi)
  result.onLoad()
  result.onShow()
  result.onReady()
  assert.equal(home.data.navigationKey, '')
  assert.equal(result.data.selectedDate, '2000-06-01 08:00')
  assert.equal(result.data.selectedNote, '自己的命盘')

  result.returnToQuery()
  result.returnToQuery()
  assert.equal(calls.filter(call => call.method === 'navigateBack').length, 1)
  home.onShow()
  assert.equal(result.data.navigationKey, '')
  assert.equal(home.data.dateText, '2000-06-01 08:00')
})

test('calendar and date selection entries use the same navigation feedback', () => {
  const { app, calls, wxApi } = environment()
  const calendar = loadPage('../pages/calendar/index', app, wxApi)
  calendar.onShow()
  calendar.onReady()
  calendar.openMovingDate()
  calendar.openMovingDate()
  assert.equal(calls.length, 1)
  assert.equal(calendar.data.navigationKey, 'moving')
  calendar.onHide()

  const moving = loadPage('../pages/moving-date/index', app, wxApi)
  moving.onLoad()
  moving.onShow()
  moving.onReady()
  assert.equal(calendar.data.navigationKey, '')

  const event = { currentTarget: { dataset: { date: '2026-10-22' } } }
  moving.openDay(event)
  moving.openDay(event)
  assert.equal(calls.length, 2)
  const detail = loadPage('../pages/personal-day/index', app, wxApi)
  detail.onLoad({ date: '2026-10-22', activity: 'moving' })
  detail.onShow()
  detail.onReady()
  assert.equal(moving.data.navigationKey, '')
})

test('mine to history tab clears loading when the cached home tab appears', () => {
  const { app, calls, wxApi } = environment()
  const home = loadPage('../pages/home/index', app, wxApi)
  home.onShow()
  home.onReady()
  const mine = loadPage('../pages/mine/index', app, wxApi)
  mine.onShow()
  mine.onReady()
  mine.openHistory()
  mine.openHistory()
  assert.equal(calls.filter(call => call.method === 'switchTab').length, 1)
  assert.equal(mine.data.navigationKey, 'history')
  mine.onHide()
  home.onShow()
  assert.equal(mine.data.navigationKey, '')
  assert.equal(home.data.mode, 'history')
})

test('a second history card tap does not replace the first pending destination', () => {
  const { app, calls, wxApi } = environment()
  const mine = loadPage('../pages/mine/index', app, wxApi)
  mine.onShow()
  const [first, second] = mine.data.recentCharts
  mine.openRecentChart({ currentTarget: { dataset: { id: first.id } } })
  mine.openRecentChart({ currentTarget: { dataset: { id: second.id } } })
  assert.equal(calls.length, 1)
  assert.equal(app.globalData.previewHistoryDetailId, first.id)
  const home = loadPage('../pages/home/index', app, wxApi)
  home.onShow()
  home.onReady()
})
