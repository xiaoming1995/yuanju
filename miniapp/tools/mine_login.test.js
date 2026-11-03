const assert = require('node:assert/strict')
const test = require('node:test')

test('WeChat login does not require a web account', async () => {
  const api = require('../utils/api')
  const previous = api.wechatLogin
  api.wechatLogin = async () => ({ user: { id: 'miniapp-user' }, isNewUser: false })
  let definition
  const toasts = []
  global.Page = value => { definition = value }
  global.wx = { showToast: options => toasts.push(options) }
  try {
    delete require.cache[require.resolve('../pages/mine/index')]
    require('../pages/mine/index')
    const page = { ...definition, data: { ...definition.data }, setData(patch) { Object.assign(this.data, patch) } }
    page.loadProfile = async () => true
    await page.openLogin()
    assert.equal(toasts.find(item => item.title === '登录成功')?.icon, 'success')
    assert.equal(page.data.loggingIn, false)
  } finally {
    api.wechatLogin = previous
  }
})

test('web binding is chosen after miniapp login', async () => {
  const api = require('../utils/api')
  const previous = api.linkWebAccount
  let received
  api.linkWebAccount = async (email, password) => { received = [email, password]; return { id: 'web-user' } }
  let definition
  global.Page = value => { definition = value }
  global.wx = { showToast() {} }
  try {
    delete require.cache[require.resolve('../pages/mine/index')]
    require('../pages/mine/index')
    const page = { ...definition, data: { ...definition.data }, setData(patch) { Object.assign(this.data, patch) } }
    page.setData({ loggedIn: true, user: { id: 'miniapp-user', email: '' }, email: 'web@example.com', password: 'secret' })
    page.toggleWebBinding()
    assert.equal(page.data.bindingVisible, true)
    page.setData({ password: 'secret' })
    page.loadProfile = async () => true
    await page.bindWebAccount()
    assert.deepEqual(received, ['web@example.com', 'secret'])
    assert.equal(page.data.bindingVisible, false)
    assert.equal(page.data.password, '')
  } finally {
    api.linkWebAccount = previous
  }
})

test('a failed profile refresh does not claim login succeeded', async () => {
  const api = require('../utils/api')
  const previous = api.wechatLogin
  api.wechatLogin = async () => ({ user: { id: 'web-user' }, isNewUser: false })
  let definition
  const toasts = []
  global.Page = value => { definition = value }
  global.wx = { showToast: options => toasts.push(options), showModal() {} }
  try {
    delete require.cache[require.resolve('../pages/mine/index')]
    require('../pages/mine/index')
    const page = { ...definition, data: { ...definition.data }, setData(patch) { Object.assign(this.data, patch) } }
    page.loadProfile = async () => false
    await page.openLogin()
    assert.equal(toasts.filter(item => item.title === '登录成功').length, 0)
  } finally {
    api.wechatLogin = previous
  }
})

test('account header opens the separate profile page without changing profile data', async () => {
  const api = require('../utils/api')
  const navigation = require('../utils/navigation_feedback')
  const previous = { token: api.token, profile: api.profile, personalChart: api.personalChart,
    assetUrl: api.assetUrl, navigate: navigation.navigate }
  const calls = []
  api.token = () => 'miniapp-jwt'
  api.profile = async () => ({ data: { user: { id: 'u1', source: 'miniapp_wechat', nickname: '微信用户', avatar_url: '' },
    stats: { chart_count: 0 }, recent_charts: [] } })
  api.assetUrl = value => value
  api.personalChart = async () => ({ chart: null })
  navigation.navigate = (page, options) => calls.push(options)
  let definition
  global.Page = value => { definition = value }
  global.wx = { showToast() {} }
  try {
    delete require.cache[require.resolve('../pages/mine/index')]
    require('../pages/mine/index')
    const page = { ...definition, data: { ...definition.data }, setData(patch) { Object.assign(this.data, patch) } }
    await page.loadProfile()
    assert.deepEqual(calls, [])
    page.openProfile({ type: 'tap' })
    assert.equal(calls[0].method, 'navigateTo')
    assert.equal(calls[0].url, '/pages/profile/index')
    assert.equal(page.data.user.nickname, '微信用户')
  } finally {
    api.token = previous.token
    api.profile = previous.profile
    api.personalChart = previous.personalChart
    api.assetUrl = previous.assetUrl
    navigation.navigate = previous.navigate
  }
})

test('only new WeChat registration opens the profile page in onboarding mode', async () => {
  const api = require('../utils/api')
  const navigation = require('../utils/navigation_feedback')
  const previous = { wechatLogin: api.wechatLogin, navigate: navigation.navigate }
  const calls = []
  let isNewUser = true
  api.wechatLogin = async () => ({ isNewUser })
  navigation.navigate = (page, options) => calls.push(options)
  let definition
  global.Page = value => { definition = value }
  global.wx = { showToast() {} }
  try {
    delete require.cache[require.resolve('../pages/mine/index')]
    require('../pages/mine/index')
    const page = { ...definition, data: { ...definition.data }, setData(patch) { Object.assign(this.data, patch) } }
    page.loadProfile = async () => {
      page.setData({ loggedIn: true, user: { id: 'u1', source: 'miniapp_wechat' } })
      return true
    }
    await page.openLogin()
    assert.equal(calls.length, 1)
    assert.equal(calls[0].url, '/pages/profile/index?onboarding=1')
    isNewUser = false
    await page.openLogin()
    assert.equal(calls.length, 1)
    assert.equal(page.data.loggedIn, true)
  } finally {
    api.wechatLogin = previous.wechatLogin
    navigation.navigate = previous.navigate
  }
})

test('Mine shows the explicitly bound chart rather than the newest history record', async () => {
  const api = require('../utils/api')
  const navigation = require('../utils/navigation_feedback')
  const previous = { token: api.token, profile: api.profile, personalChart: api.personalChart, navigate: navigation.navigate }
  const chart = (id, year) => ({ id, birth_year: year, birth_month: 1, birth_day: 1, birth_hour: 12,
    year_gan: '甲', year_zhi: '子', month_gan: '乙', month_zhi: '丑',
    day_gan: '庚', day_zhi: '午', hour_gan: '丁', hour_zhi: '卯' })
  const visits = []
  const app = { globalData: { previewHistoryNext: false, previewHistoryDetailId: null } }
  global.getApp = () => app
  global.wx = { showToast() {} }
  api.token = () => 'current-user'
  api.profile = async () => ({ data: { user: { id: 'u1', nickname: '用户', avatar_url: '' },
    stats: { chart_count: 2 }, recent_charts: [chart('newest', 1996), chart('selected', 1995)] } })
  api.personalChart = async () => ({ chart: chart('selected', 1995) })
  navigation.navigate = (_page, options) => visits.push(options)
  let definition
  global.Page = value => { definition = value }
  try {
    delete require.cache[require.resolve('../pages/mine/index')]
    require('../pages/mine/index')
    const page = { ...definition, data: { ...definition.data }, setData(patch) { Object.assign(this.data, patch) } }
    assert.equal(await page.loadProfile(), true)
    assert.equal(page.data.chartCount, 2)
    assert.equal('recentCharts' in page.data, false)
    assert.equal(page.data.personalChart.id, 'selected')
    page.openPersonalChart()
    assert.equal(visits[0].url, '/pages/result/index')
    api.personalChart = async () => ({ chart: null })
    await page.loadProfile()
    assert.equal(page.data.personalChart, null)
    assert.equal(page.data.chartCount, 2)
  } finally {
    Object.assign(api, { token: previous.token, profile: previous.profile, personalChart: previous.personalChart })
    navigation.navigate = previous.navigate
  }
})

test('failed profile refresh hides stale account data and can retry', async () => {
  const api = require('../utils/api')
  const previous = { token: api.token, profile: api.profile, personalChart: api.personalChart }
  let fail = false
  api.token = () => 'current-user'
  api.profile = async () => {
    if (fail) throw new Error('网络暂不可用')
    return { data: { user: { id: 'u1', nickname: '小明', avatar_url: '' }, stats: { chart_count: 2 }, recent_charts: [] } }
  }
  api.personalChart = async () => ({ chart: null })
  let definition
  global.Page = value => { definition = value }
  global.wx = { showToast() {} }
  try {
    delete require.cache[require.resolve('../pages/mine/index')]
    require('../pages/mine/index')
    const page = { ...definition, data: { ...definition.data }, setData(patch) { Object.assign(this.data, patch) } }
    assert.equal(await page.loadProfile(), true)
    fail = true
    assert.equal(await page.loadProfile(), false)
    assert.equal(page.data.user, null)
    assert.equal(page.data.chartCount, 0)
    assert.equal('recentCharts' in page.data, false)
    assert.equal(page.data.profileError, '网络暂不可用')
    fail = false
    assert.equal(await page.loadProfile(), true)
    assert.equal(page.data.profileError, '')
    assert.equal(page.data.user.nickname, '小明')
  } finally {
    Object.assign(api, previous)
  }
})

test('logout clears visible account data and the web binding form', () => {
  const api = require('../utils/api')
  const previous = api.logout
  let loggedOut = false
  api.logout = () => { loggedOut = true }
  let definition
  global.Page = value => { definition = value }
  global.wx = { showToast() {} }
  try {
    delete require.cache[require.resolve('../pages/mine/index')]
    require('../pages/mine/index')
    const page = { ...definition, data: { ...definition.data }, setData(patch) { Object.assign(this.data, patch) } }
    page.setData({ loggedIn: true, user: { id: 'u1' }, chartCount: 2,
      personalChart: { id: 'c1' }, bindingVisible: true, email: 'web@example.com', password: 'secret' })
    page.logoutPreview()
    assert.equal(loggedOut, true)
    assert.equal(page.data.loggedIn, false)
    assert.equal(page.data.user, null)
    assert.equal(page.data.chartCount, 0)
    assert.equal(page.data.personalChart, null)
    assert.equal(page.data.bindingVisible, false)
    assert.equal(page.data.password, '')
  } finally {
    api.logout = previous
  }
})
