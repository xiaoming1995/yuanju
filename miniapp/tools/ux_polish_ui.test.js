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

test('the preview date picker produces an hour-precision date', () => {
  const app = loadApp()
  const page = loadPage('../pages/home/index', app)
  page.setData({ pickerValue: [90, 0, 0, 1] })
  page.confirmPicker()
  assert.equal(page.data.dateText, '1990-01-01 01:00')
})

test('long date groups reveal seven days at a time and reset for another activity', () => {
  const app = { globalData: { previewDayProfileEnabled: true } }
  const page = loadPage('../pages/moving-date/index', app)
  page.onLoad()
  page.setData({ startDate: '2026-10-22', endDate: '2026-12-22' })
  page.refreshResults()
  const longSection = page.data.sections.find(section => section.days.length > 7)
  assert.ok(longSection)
  assert.equal(longSection.visibleDays.length, 7)
  page.showMore({ currentTarget: { dataset: { section: longSection.id } } })
  assert.equal(page.data.sections.find(section => section.id === longSection.id).visibleDays.length, 14)
  while (page.data.sections.find(section => section.id === longSection.id).hasMore) {
    page.showMore({ currentTarget: { dataset: { section: longSection.id } } })
  }
  const expanded = page.data.sections.find(section => section.id === longSection.id)
  assert.equal(expanded.visibleDays.length, expanded.days.length)
  page.changeActivity({ currentTarget: { dataset: { activity: 'marriage' } } })
  assert.ok(page.data.sections.every(section => section.visibleDays.length <= 7))
  assert.equal(page.data.total, 62)
})

test('removing preview history updates the count and recent item on Mine', () => {
  const app = loadApp()
  const home = loadPage('../pages/home/index', app, {
    setNavigationBarTitle() {},
    showModal(options) { options.success({ confirm: true }) },
    showToast() {},
  })
  home.onShow()
  home.removeHistory({ currentTarget: { dataset: { id: 'self' } } })
  const mine = loadPage('../pages/mine/index', app, {
    switchTab() {},
    navigateTo() {},
    setNavigationBarTitle() {},
  })
  mine.onShow()
  assert.equal(home.data.history.length, 3)
  assert.equal(mine.data.chartCount, 3)
  assert.equal(mine.data.latestChart.id, 'child')
  assert.equal(mine.data.recentCharts[0].id, 'child')
  mine.openLatestChart()
  assert.equal(app.globalData.previewHistoryDetailId, 'child')
  assert.equal(app.globalData.previewHistoryNext, true)
  mine.onHide()
  home.onShow()
  home.onReady()
  assert.equal(home.data.detail.id, 'child')
})
