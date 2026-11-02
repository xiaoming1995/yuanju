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
    setData(patch) { this.data = { ...this.data, ...patch } },
  }
}

test('calendar preview follows the selected day and opens its detail route', () => {
  const app = { globalData: { previewDayProfileEnabled: false } }
  const visits = []
  const wxApi = {
    navigateTo(value) { visits.push(value.url) },
  }
  const page = loadPage('../pages/calendar/index', app, wxApi)

  assert.equal(page.data.dayPreviewEnabled, false)
  page.enablePersonalDay()
  assert.equal(app.globalData.previewDayProfileEnabled, true)
  assert.equal(page.data.personalDay.date, new Date(Date.now() + 8 * 60 * 60 * 1000).toISOString().slice(0, 10))

  page.pickDate({ currentTarget: { dataset: { key: '2026-10-24' } } })
  assert.equal(page.data.personalDay.date, '2026-10-24')
  assert.equal(page.data.selected.jianChu, '收')
  assert.equal(page.data.selected.jianChuInfo.label, '收日')
  page.openPersonalDay()
  assert.deepEqual(visits, ['/pages/personal-day/index?date=2026-10-24'])
  const target = loadPage('../pages/personal-day/index', app, wxApi)
  target.onLoad({ date: '2026-10-24' })
  target.onReady()
})

test('returning from detail after ending preview restores the invitation', () => {
  const app = { globalData: { previewDayProfileEnabled: true } }
  const page = loadPage('../pages/calendar/index', app, {})

  page.onShow()
  assert.equal(page.data.dayPreviewEnabled, true)
  app.globalData.previewDayProfileEnabled = false
  page.onShow()
  assert.equal(page.data.dayPreviewEnabled, false)
})

test('detail reads the selected date and ending demo clears the process state', () => {
  const app = { globalData: { previewDayProfileEnabled: true } }
  let returned = 0
  const wxApi = {
    navigateBack() { returned += 1 },
  }
  const page = loadPage('../pages/personal-day/index', app, wxApi)

  page.onLoad({ date: '2026-10-24' })
  assert.equal(page.data.personalDay.date, '2026-10-24')
  assert.equal(page.data.personalDay.dayGanZhi, '辛未')
  assert.equal(page.data.jianChuInfo.label, '收日')
  page.endPreview()
  assert.equal(app.globalData.previewDayProfileEnabled, false)
  assert.equal(returned, 1)
  const target = loadPage('../pages/calendar/index', app, wxApi)
  target.onShow()
  target.onReady()
})

test('detail opened from date selection shows the selected activity evidence', () => {
  const app = { globalData: { previewDayProfileEnabled: true } }
  const page = loadPage('../pages/personal-day/index', app, {})
  page.onLoad({ date: '2026-10-23', activity: 'marriage' })
  assert.equal(page.data.selection.activityLabel, '结婚')
  assert.deepEqual(page.data.selection.yi, ['嫁娶'])
  assert.deepEqual(page.data.selection.ji, [])
  assert.equal(page.data.selection.date, page.data.personalDay.date)
  page.onLoad({ date: '2026-10-23' })
  assert.equal(page.data.selection, null)
})
