const assert = require('node:assert/strict')
const test = require('node:test')
const fs = require('node:fs')
const path = require('node:path')

function page(modulePath, app, wxApi) {
  let definition
  global.Page = value => { definition = value }
  global.getApp = () => app
  global.wx = wxApi
  delete require.cache[require.resolve(modulePath)]
  require(modulePath)
  return { ...definition, data: { ...definition.data }, setData(patch) { Object.assign(this.data, patch) } }
}

test('Home removes arbitrary birth form and keeps the calendar', () => {
  const wxml = fs.readFileSync(path.join(__dirname, '../pages/home/index.wxml'), 'utf8')
  assert.doesNotMatch(wxml, /出生信息排盘|开始排盘|openPicker|date-picker/)
  assert.match(wxml, /今日干支|calendarPreview/)
})

test('Mine setup saves own birth details and switches the binding', async () => {
  const api = require('../utils/api')
  const navigation = require('../utils/navigation_feedback')
  const previous = { token: api.token, personalChart: api.personalChart,
    savePersonalChart: api.savePersonalChart, navigate: navigation.navigate }
  const app = { globalData: {} }
  const saved = []
  const visits = []
  api.token = () => 'user-token'
  api.personalChart = async () => ({ chart: null })
  api.savePersonalChart = async input => { saved.push(input); return { chart: { id: 'owned' } } }
  navigation.navigate = (_page, options) => visits.push(options)
  try {
    const setup = page('../pages/personal-chart-setup/index', app, { showToast() {} })
    await setup.loadPersonalChart()
    setup.changeDate({ detail: { value: '1995-10-12' } })
    setup.changeHour({ detail: { value: 12 } })
    setup.changeGender({ currentTarget: { dataset: { gender: 'female' } } })
    await setup.save()
    assert.deepEqual(saved, [{ year: 1995, month: 10, day: 12, hour: 12,
      gender: 'female', calendar_type: 'solar' }])
    assert.equal(app.globalData.personalChart.id, 'owned')
    assert.equal(visits[0].method, 'navigateBack')
  } finally { Object.assign(api, previous); navigation.navigate = previous.navigate }
})

test('unbound result and non-bound history do not display full detail', async () => {
  const api = require('../utils/api')
  const previous = { token: api.token, personalChartDetail: api.personalChartDetail }
  const app = { globalData: {} }
  const toasts = []
  api.token = () => 'user-token'
  api.personalChartDetail = async () => { const error = new Error('请先配置本人命盘'); error.code = 'personal_chart_unbound'; throw error }
  const wxApi = { showToast(options) { toasts.push(options.title) }, setNavigationBarTitle() {} }
  try {
    const result = page('../pages/result/index', app, wxApi)
    await result.onLoad()
    assert.equal(result.data.chart, null)
    assert.equal(result.data.loadError, '请先配置本人命盘')
    const home = page('../pages/home/index', app, wxApi)
    home.setData({ personalChartId: 'owned' })
    await home.openHistoryDetail({ currentTarget: { dataset: { id: 'other' } } })
    assert.equal(home.data.detail, null)
    assert.match(toasts[0], /本人命盘/)
  } finally { Object.assign(api, previous) }
})
