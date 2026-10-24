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

test('calendar opens the dedicated date selection page', () => {
  const visits = []
  const app = { globalData: {} }
  const wxApi = {
    navigateTo(value) { visits.push(value.url) },
  }
  const page = loadPage('../pages/calendar/index', app, wxApi)
  page.openMovingDate()
  assert.deepEqual(visits, ['/pages/moving-date/index'])
  const target = loadPage('../pages/moving-date/index', app, wxApi)
  target.onLoad()
  target.onReady()
})

test('date selection page asks to enable the fixed sample before showing personal candidates', () => {
  const app = { globalData: { previewDayProfileEnabled: false } }
  const page = loadPage('../pages/moving-date/index', app, {})
  page.onLoad()
  assert.equal(page.data.enabled, false)
  assert.equal(page.data.total, 0)
  page.startPreview()
  assert.equal(app.globalData.previewDayProfileEnabled, true)
  assert.equal(page.data.enabled, true)
  assert.ok(page.data.total > 0)
})

test('changing activity recomputes results without losing the date range', () => {
  const app = { globalData: { previewDayProfileEnabled: true } }
  const page = loadPage('../pages/moving-date/index', app, {})
  page.onLoad()
  page.setData({ startDate: '2026-10-22', endDate: '2026-10-24' })
  page.refreshResults()
  assert.equal(page.data.activityId, 'moving')
  page.changeActivity({ currentTarget: { dataset: { activity: 'marriage' } } })
  assert.equal(page.data.activityId, 'marriage')
  assert.equal(page.data.activity.label, '结婚')
  assert.equal(page.data.startDate, '2026-10-22')
  assert.equal(page.data.endDate, '2026-10-24')
  assert.equal(page.data.total, 3)
  assert.equal(page.data.activityOptions.filter(item => item.selected).length, 1)
  page.onShow()
  assert.equal(page.data.activityId, 'marriage')
})

test('activity and displayed results update in the same view patch', () => {
  const app = { globalData: { previewDayProfileEnabled: true } }
  const page = loadPage('../pages/moving-date/index', app, {})
  page.onLoad()
  page.setData({ startDate: '2026-10-22', endDate: '2026-10-24' })
  page.refreshResults()
  const oldSections = page.data.sections
  const snapshots = []
  const originalSetData = page.setData
  page.setData = function (patch, callback) {
    originalSetData.call(this, patch, callback)
    snapshots.push({ activityId: this.data.activityId, sections: this.data.sections })
  }
  page.changeActivity({ currentTarget: { dataset: { activity: 'marriage' } } })
  assert.notDeepEqual(page.data.sections, oldSections)
  assert.ok(snapshots.every(snapshot => snapshot.activityId !== 'marriage' || snapshot.sections === page.data.sections))
})

test('invalid range clears old results rather than retaining a partial recommendation', () => {
  const app = { globalData: { previewDayProfileEnabled: true } }
  const page = loadPage('../pages/moving-date/index', app, {})
  page.onLoad()
  page.setData({ startDate: '2026-10-23', endDate: '2026-10-24' })
  page.refreshResults()
  assert.equal(page.data.total, 2)
  page.changeStartDate({ detail: { value: '2026-10-25' } })
  assert.match(page.data.rangeError, /起日/)
  assert.equal(page.data.total, 0)
  assert.ok(page.data.sections.every(section => section.days.length === 0))
})

test('an unavailable default date keeps the supported range visible after enabling the preview', () => {
  const app = { globalData: { previewDayProfileEnabled: false } }
  const page = loadPage('../pages/moving-date/index', app, {})
  page.onLoad()
  page.setData({ startDate: '', endDate: '', rangeError: '演示日历范围为2025—2027年，请选择范围内的日期。' })
  page.startPreview()
  assert.match(page.data.rangeError, /2025—2027/)
  assert.equal(page.data.total, 0)
})

test('no priority candidate still leaves the conflicting day available', () => {
  const app = { globalData: { previewDayProfileEnabled: true } }
  const page = loadPage('../pages/moving-date/index', app, {})
  page.onLoad()
  page.setData({ startDate: '2026-10-22', endDate: '2026-10-22' })
  page.refreshResults()
  assert.equal(page.data.priorityCount, 0)
  assert.equal(page.data.total, 1)
  assert.equal(page.data.sections.find(section => section.id === 'conflict').days[0].date, '2026-10-22')
})

test('opening the daily detail and returning retains the selected activity and range', () => {
  const visits = []
  const app = { globalData: { previewDayProfileEnabled: true } }
  const wxApi = {
    navigateTo(value) { visits.push(value.url) },
  }
  const page = loadPage('../pages/moving-date/index', app, wxApi)
  page.onLoad()
  page.setData({ startDate: '2026-10-23', endDate: '2026-10-24' })
  page.refreshResults()
  page.changeActivity({ currentTarget: { dataset: { activity: 'travel' } } })
  page.openDay({ currentTarget: { dataset: { date: '2026-10-23' } } })
  page.onShow()
  assert.deepEqual(visits, ['/pages/personal-day/index?date=2026-10-23&activity=travel'])
  assert.equal(page.data.startDate, '2026-10-23')
  assert.equal(page.data.endDate, '2026-10-24')
  assert.equal(page.data.activityId, 'travel')
  assert.equal(page.data.total, 2)
  const target = loadPage('../pages/personal-day/index', app, wxApi)
  target.onLoad({ date: '2026-10-23', activity: 'travel' })
  target.onReady()
})
