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
  require('../config').apiBaseUrl = 'https://api.test.example'
  const app = { globalData: {} }
  const calls = []
  const storage = { yj_token: 'test-token' }
  const wxApi = {
    navigateTo(options) { calls.push({ method: 'navigateTo', options }) },
    switchTab(options) { calls.push({ method: 'switchTab', options }) },
    navigateBack(options) { calls.push({ method: 'navigateBack', options }) },
    setNavigationBarTitle() {},
    showToast() {},
    getStorageSync(key) { return storage[key] },
    removeStorageSync(key) { delete storage[key] },
    request(options) {
      const data = options.url.includes('/bazi/calculate')
        ? { chart_id: 'chart-1', result: { year_gan: '甲', year_zhi: '子', month_gan: '乙', month_zhi: '丑', day_gan: '丙', day_zhi: '寅', hour_gan: '丁', hour_zhi: '卯', wuxing: { total: 8 } } }
        : options.url.includes('/bazi/history/')
          ? { chart: { id: 'chart-1', birth_year: 2000, birth_month: 6, birth_day: 1, birth_hour: 8, year_gan: '甲', year_zhi: '子', month_gan: '乙', month_zhi: '丑', day_gan: '丙', day_zhi: '寅', hour_gan: '丁', hour_zhi: '卯' }, result: {} }
          : { charts: [], total: 0 }
      options.success({ statusCode: 200, data })
    },
  }
  return { app, calls, wxApi }
}

test('home result and back navigation acknowledge only one tap and restore the form', async () => {
  const { app, calls, wxApi } = environment()
  const home = loadPage('../pages/home/index', app, wxApi)
  home.onShow()
  home.onReady()
  home.setData({ dateText: '2000-06-01 08:00', note: '自己的干支' })

  await Promise.all([home.openDemoResult(), home.openDemoResult()])
  assert.equal(calls.filter(call => call.method === 'navigateTo').length, 1)
  assert.equal(home.data.navigationKey, 'result')
  home.onHide()

  const result = loadPage('../pages/result/index', app, wxApi)
  result.onLoad()
  result.onShow()
  result.onReady()
  assert.equal(home.data.navigationKey, '')
  assert.equal(result.data.selectedDate, '2000-06-01 08:00')
  assert.equal(result.data.selectedNote, '自己的干支')

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

  moving.setData({ personalChart: { id: 'chart-1' }, sections: [{ id: 'yi', days: [{ date: '2026-10-22' }] }] })
  const event = { currentTarget: { dataset: { date: '2026-10-22' } } }
  moving.openDay(event)
  assert.equal(moving.data.expandedDate, '2026-10-22')
  moving.openPersonalDay(event)
  moving.openPersonalDay(event)
  assert.equal(calls.length, 2)
  const detail = loadPage('../pages/personal-day/index', app, wxApi)
  detail.onLoad({ date: '2026-10-22' })
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

test('empty Mine record action opens the query form even when Home was showing history', () => {
  const { app, calls, wxApi } = environment()
  const home = loadPage('../pages/home/index', app, wxApi)
  home.setData({ mode: 'history' })
  const mine = loadPage('../pages/mine/index', app, wxApi)
  mine.openCreateChart()
  assert.equal(calls.at(-1).method, 'switchTab')
  mine.onHide()
  home.onShow()
  home.onReady()
  assert.equal(home.data.mode, 'query')
  home.onHide()
})

test('unbound Mine chart directs users without records to the query form', () => {
  const { app, calls, wxApi } = environment()
  const home = loadPage('../pages/home/index', app, wxApi)
  home.setData({ mode: 'history' })
  const mine = loadPage('../pages/mine/index', app, wxApi)
  mine.setData({ chartCount: 0, personalChartError: '' })
  mine.openPersonalBanner()
  assert.equal(calls.at(-1).method, 'switchTab')
  mine.onHide()
  home.onShow()
  home.onReady()
  assert.equal(home.data.mode, 'query')
  home.onHide()
})

test('a second history card tap does not replace the first pending destination', () => {
  const { app, calls, wxApi } = environment()
  const mine = loadPage('../pages/mine/index', app, wxApi)
  mine.setData({ recentCharts: [{ id: 'first' }, { id: 'second' }] })
  const [first, second] = mine.data.recentCharts
  mine.openRecentChart({ currentTarget: { dataset: { id: first.id } } })
  mine.openRecentChart({ currentTarget: { dataset: { id: second.id } } })
  assert.equal(calls.length, 1)
  assert.equal(app.globalData.previewHistoryDetailId, first.id)
  const home = loadPage('../pages/home/index', app, wxApi)
  home.onShow()
  home.onReady()
})

test('profile navigation prevents repeated opening and refreshes Mine after automatic saving and return', async () => {
  const { app, calls, wxApi } = environment()
  const user = { id: 'u1', nickname: '小明', avatar_url: '', source: 'miniapp_wechat' }
  wxApi.request = options => {
    if (options.method === 'PUT') user.nickname = options.data.nickname
    options.success({ statusCode: 200, data: { data: { user: { ...user }, stats: { chart_count: 0 }, recent_charts: [] } } })
  }
  const mine = loadPage('../pages/mine/index', app, wxApi)
  await mine.onShow()
  mine.onReady()
  mine.openProfile({ type: 'tap' })
  mine.openProfile({ type: 'tap' })
  assert.equal(calls.length, 1)
  assert.equal(calls[0].options.url, '/pages/profile/index')
  mine.onHide()

  const profile = loadPage('../pages/profile/index', app, wxApi)
  await profile.onLoad({})
  profile.onShow()
  profile.onReady()
  assert.equal(mine.data.navigationKey, '')
  assert.equal(profile.data.profileNickname, '小明')
  profile.onProfileNicknameInput({ detail: { value: '新昵称' } })
  profile.onNicknameBlur({ detail: { value: '新昵称' } })
  await profile.onProfileNicknameReview({ detail: { pass: true } })
  assert.equal(calls.length, 1)
  wxApi.navigateBack({ delta: 1 })
  assert.equal(calls.at(-1).method, 'navigateBack')
  profile.onHide()
  await mine.onShow()
  profile.onUnload()
  assert.equal(mine.data.user.nickname, '新昵称')
  assert.equal(mine.data.navigationKey, '')
})

test('Mine updates saved profile before a slow personal chart request finishes', async () => {
  const { app, wxApi } = environment()
  const user = { id: 'u1', nickname: '旧昵称', avatar_url: '', source: 'miniapp_wechat' }
  let finishChart
  wxApi.request = options => {
    if (options.url.includes('/user/personal-chart')) {
      finishChart = () => options.success({ statusCode: 200, data: { chart: null } })
      return
    }
    options.success({ statusCode: 200, data: { data: { user: { ...user }, stats: { chart_count: 0 }, recent_charts: [] } } })
  }
  const mine = loadPage('../pages/mine/index', app, wxApi)
  const initialLoad = mine.onShow()
  await Promise.resolve()
  finishChart()
  await initialLoad

  user.nickname = '新昵称'
  user.avatar_url = '/static/uploads/user-avatars/new.png'
  const returnLoad = mine.onShow()
  await Promise.resolve()
  assert.equal(mine.data.user.nickname, '新昵称')
  assert.equal(mine.data.userAvatarUrl, 'https://api.test.example/static/uploads/user-avatars/new.png')
  assert.equal(mine.data.personalChartLoading, true)
  finishChart()
  await returnLoad
})

test('Mine displays a LAN avatar from the miniapp local file', async () => {
  const { app, wxApi } = environment()
  require('../config').apiBaseUrl = 'http://192.168.101.24:9002'
  wxApi.env = { USER_DATA_PATH: 'wxfile://user' }
  wxApi.getFileSystemManager = () => ({ writeFile(options) { options.success({}) } })
  wxApi.request = options => {
    if (options.url.includes('/static/uploads/user-avatars/')) {
      options.success({ statusCode: 200, data: new Uint8Array([1, 2, 3]).buffer })
    } else if (options.url.includes('/user/personal-chart')) {
      options.success({ statusCode: 200, data: { chart: null } })
    } else {
      options.success({ statusCode: 200, data: { data: { user: {
        id: 'u1', nickname: '小明', avatar_url: '/static/uploads/user-avatars/picture.jpg',
      }, stats: { chart_count: 0 }, recent_charts: [] } } })
    }
  }
  const mine = loadPage('../pages/mine/index', app, wxApi)
  await mine.onShow()
  await Promise.resolve()
  assert.equal(mine.data.userAvatarUrl, 'wxfile://user/avatar-picture.jpg')
})
