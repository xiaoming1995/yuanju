const assert = require('node:assert/strict')
const test = require('node:test')
const api = require('../utils/api')
const config = require('../config')

function fakeWx(responder) {
  const storage = {}
  const calls = []
  global.wx = {
    getStorageSync: key => storage[key],
    setStorageSync: (key, value) => { storage[key] = value },
    removeStorageSync: key => { delete storage[key] },
    login: options => options.success({ code: 'one-time-code' }),
    request: options => { calls.push(options); options.success(responder(options)) },
  }
  return { storage, calls }
}

test('WeChat login exchanges code through backend and stores only JWT', async () => {
  config.apiBaseUrl = 'https://api.test.example'
  const { storage, calls } = fakeWx(() => ({ statusCode: 200, data: { user: { id: 'u1' }, token: 'jwt1', is_new_user: true } }))
  const login = await api.wechatLogin()
  assert.equal(login.user.id, 'u1')
  assert.equal(login.isNewUser, true)
  assert.equal(storage.yj_token, 'jwt1')
  assert.equal(calls.length, 1)
  assert.equal(calls[0].url, 'https://api.test.example/api/auth/wechat/login')
  assert.deepEqual(calls[0].data, { code: 'one-time-code' })
})

test('failed web link keeps the miniapp session active', async () => {
  config.apiBaseUrl = 'https://api.test.example'
  const { storage, calls } = fakeWx(() => ({ statusCode: 409, data: { error: '微信身份已关联其他账号' } }))
  storage.yj_token = 'miniapp-jwt'
  await assert.rejects(api.linkWebAccount('old@example.com', 'password'), /微信身份已关联其他账号/)
  assert.equal(storage.yj_token, 'miniapp-jwt')
  assert.equal(calls[0].header.Authorization, 'Bearer miniapp-jwt')
})

test('wrong web password does not clear the miniapp session', async () => {
  config.apiBaseUrl = 'https://api.test.example'
  const { storage } = fakeWx(() => ({ statusCode: 422, data: { error: '邮箱或密码错误' } }))
  storage.yj_token = 'miniapp-jwt'
  await assert.rejects(api.linkWebAccount('old@example.com', 'wrong'), /邮箱或密码错误/)
  assert.equal(storage.yj_token, 'miniapp-jwt')
})

test('linking an existing web account replaces the miniapp JWT with web JWT', async () => {
  config.apiBaseUrl = 'https://api.test.example'
  const { storage, calls } = fakeWx(() => ({ statusCode: 200, data: { user: { id: 'web-user' }, token: 'web-jwt' } }))
  storage.yj_token = 'miniapp-jwt'
  const user = await api.linkWebAccount('old@example.com', 'password')
  assert.equal(user.id, 'web-user')
  assert.equal(storage.yj_token, 'web-jwt')
  assert.equal(calls[0].url, 'https://api.test.example/api/auth/wechat/link-web')
  assert.equal(calls[0].header.Authorization, 'Bearer miniapp-jwt')
  assert.deepEqual(calls[0].data, { email: 'old@example.com', password: 'password' })
})

test('selected WeChat avatar uploads with the current JWT', async () => {
  config.apiBaseUrl = 'https://api.test.example'
  const { storage } = fakeWx(() => ({ statusCode: 200, data: {} }))
  storage.yj_token = 'miniapp-jwt'
  let upload
  global.wx.uploadFile = options => {
    upload = options
    options.success({ statusCode: 200, data: JSON.stringify({ data: { avatar_url: '/static/uploads/user-avatars/picture.png' } }) })
  }
  const result = await api.uploadAvatar('wxfile://selected-avatar')
  assert.equal(upload.url, 'https://api.test.example/api/user/profile/avatar')
  assert.equal(upload.filePath, 'wxfile://selected-avatar')
  assert.equal(upload.header.Authorization, 'Bearer miniapp-jwt')
  assert.equal(api.assetUrl(result.data.avatar_url), 'https://api.test.example/static/uploads/user-avatars/picture.png')
})

test('HTTP 401 clears an expired token', async () => {
  config.apiBaseUrl = 'https://api.test.example'
  const { storage } = fakeWx(() => ({ statusCode: 401, data: { error: '登录已过期' } }))
  storage.yj_token = 'expired'
  await assert.rejects(api.profile(), /登录已过期/)
  assert.equal(storage.yj_token, undefined)
})

test('disabled account response clears the miniapp session', async () => {
  config.apiBaseUrl = 'https://api.test.example'
  const { storage } = fakeWx(() => ({ statusCode: 403, data: { error: '该账号已被禁用', code: 'account_disabled' } }))
  storage.yj_token = 'disabled-user-jwt'
  await assert.rejects(api.profile(), /已被禁用/)
  assert.equal(storage.yj_token, undefined)
})

test('developer tools can call the local HTTP backend', async () => {
  config.apiBaseUrl = 'http://127.0.0.1:9002'
  const { calls } = fakeWx(() => ({ statusCode: 200, data: { status: 'ok' } }))
  global.wx.getDeviceInfo = () => ({ platform: 'devtools' })
  global.wx.getSystemInfoSync = () => { throw new Error('deprecated API should not be called') }
  await api.request('GET', '/health-check')
  assert.equal(calls[0].url, 'http://127.0.0.1:9002/api/health-check')
})

test('almanac day uses a public date endpoint', async () => {
  config.apiBaseUrl = 'https://api.test.example'
  const { calls } = fakeWx(() => ({ statusCode: 200, data: { data: { date: '2026-11-02', yi: ['祭祀'], ji: ['开市'] } } }))
  const result = await api.calendarDay('2026-11-02')
  assert.equal(calls[0].url, 'https://api.test.example/api/calendar/day?date=2026-11-02')
  assert.equal(calls[0].header.Authorization, undefined)
  assert.deepEqual(result.data.ji, ['开市'])
})

test('date selection reads a public bounded range', async () => {
  config.apiBaseUrl = 'https://api.test.example'
  const { calls } = fakeWx(() => ({ statusCode: 200, data: { data: { days: [] } } }))
  await api.calendarRange('2028-01-01', '2028-01-30')
  assert.equal(calls[0].url, 'https://api.test.example/api/calendar/range?start=2028-01-01&end=2028-01-30')
  assert.equal(calls[0].header.Authorization, undefined)
})

test('phone can call a private LAN HTTP backend and upload an avatar', async () => {
  config.apiBaseUrl = 'http://192.168.101.24:9002'
  const { calls, storage } = fakeWx(() => ({ statusCode: 200, data: { status: 'ok' } }))
  global.wx.getDeviceInfo = () => ({ platform: 'ios' })
  storage.yj_token = 'miniapp-jwt'
  await api.request('GET', '/auth/me')
  assert.equal(calls[0].url, 'http://192.168.101.24:9002/api/auth/me')
  let upload
  global.wx.uploadFile = options => {
    upload = options
    options.success({ statusCode: 200, data: '{}' })
  }
  await api.uploadAvatar('wxfile://avatar')
  assert.equal(upload.url, 'http://192.168.101.24:9002/api/user/profile/avatar')
})

test('phone reads a LAN avatar through request and displays a local image file', async () => {
  config.apiBaseUrl = 'http://192.168.101.24:9002'
  const pixels = new Uint8Array([1, 2, 3]).buffer
  let requested, written
  global.wx = {
    env: { USER_DATA_PATH: 'wxfile://user' },
    request(options) {
      requested = options
      options.success({ statusCode: 200, data: pixels })
    },
    getFileSystemManager: () => ({ writeFile(options) {
      written = options
      options.success({})
    } }),
  }
  const path = await api.avatarDisplayUrl('/static/uploads/user-avatars/picture.jpg')
  assert.equal(requested.url, 'http://192.168.101.24:9002/static/uploads/user-avatars/picture.jpg')
  assert.equal(requested.responseType, 'arraybuffer')
  assert.equal(written.data, pixels)
  assert.equal(path, 'wxfile://user/avatar-picture.jpg')
})

test('local HTTP is rejected outside developer tools and public HTTP is always rejected', async () => {
  config.apiBaseUrl = 'http://127.0.0.1:9002'
  const { calls } = fakeWx(() => ({ statusCode: 200, data: {} }))
  global.wx.getDeviceInfo = () => ({ platform: 'ios' })
  global.wx.getSystemInfoSync = () => ({ platform: 'devtools' })
  await assert.rejects(api.request('GET', '/auth/me'), /HTTPS/)
  global.wx.getDeviceInfo = () => ({ platform: 'devtools' })
  config.apiBaseUrl = 'http://example.com'
  await assert.rejects(api.request('GET', '/auth/me'), /HTTPS/)
  config.apiBaseUrl = 'http://0.0.0.0:9002'
  await assert.rejects(api.request('GET', '/auth/me'), /HTTPS/)
  assert.equal(calls.length, 0)
})

test('legacy device detection remains limited to local developer tools', async () => {
  config.apiBaseUrl = 'http://127.0.0.1:9002'
  const { calls } = fakeWx(() => ({ statusCode: 200, data: {} }))
  global.wx.getSystemInfoSync = () => ({ platform: 'devtools' })
  await api.request('GET', '/health-check')
  global.wx.getSystemInfoSync = () => ({ platform: 'android' })
  await assert.rejects(api.request('GET', '/health-check'), /HTTPS/)
  assert.equal(calls.length, 1)
})

test('login error retains status and code so the page can explain registration control', async () => {
  config.apiBaseUrl = 'https://api.test.example'
  fakeWx(() => ({ statusCode: 403, data: { error: '小程序新用户注册暂未开放', code: 'miniapp_registration_disabled' } }))
  await assert.rejects(api.wechatLogin(), err =>
    err.message === '小程序新用户注册暂未开放' && err.statusCode === 403 && err.code === 'miniapp_registration_disabled')
})

test('personal chart and selected day calls use the authenticated account', async () => {
  config.apiBaseUrl = 'https://api.test.example'
  const { storage, calls } = fakeWx(() => ({ statusCode: 200, data: { chart: { id: 'chart-1' }, data: { status: 'caution' } } }))
  storage.yj_token = 'current-user'
  await api.personalChart()
  await api.setPersonalChart('chart-1')
  await api.personalDay('2028-01-01')
  await api.clearPersonalChart()
  assert.deepEqual(calls.map(call => `${call.method} ${call.url.slice('https://api.test.example'.length)}`), [
    'GET /api/user/personal-chart', 'PUT /api/user/personal-chart',
    'GET /api/user/personal-day?date=2028-01-01', 'DELETE /api/user/personal-chart',
  ])
  assert.deepEqual(calls[1].data, { chart_id: 'chart-1' })
  assert.ok(calls.every(call => call.header.Authorization === 'Bearer current-user'))
})
