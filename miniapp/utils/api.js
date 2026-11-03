const config = require('../config')
const TOKEN_KEY = 'yj_token'

function token() { return wx.getStorageSync(TOKEN_KEY) || '' }

function allowLocalDevHTTP(base) {
  if (!/^http:\/\/(?:127\.0\.0\.1|localhost)(?::\d+)?$/.test(base)) return false
  try {
    const device = typeof wx.getDeviceInfo === 'function' ? wx.getDeviceInfo()
      : typeof wx.getSystemInfoSync === 'function' ? wx.getSystemInfoSync() : null
    return Boolean(device && device.platform === 'devtools')
  } catch (_) {
    return false
  }
}

function allowLanHTTP(base) {
  return /^http:\/\/(?:10\.\d{1,3}\.\d{1,3}\.\d{1,3}|172\.(?:1[6-9]|2\d|3[01])\.\d{1,3}\.\d{1,3}|192\.168\.\d{1,3}\.\d{1,3})(?::\d+)?$/.test(base)
}

function allowHTTP(base) {
  return allowLocalDevHTTP(base) || allowLanHTTP(base)
}

function request(method, path, data, authenticated = false) {
  const base = (config.apiBaseUrl || '').replace(/\/$/, '')
  if (!/^https:\/\//.test(base) && !allowHTTP(base)) {
    return Promise.reject(new Error('请配置 HTTPS 后端地址或同一局域网的 HTTP IP；127.0.0.1 仅限开发者工具模拟器'))
  }
  const currentToken = token()
  if (authenticated && !currentToken) return Promise.reject(new Error('请先登录'))
  return new Promise((resolve, reject) => {
    wx.request({
      url: `${base}/api${path}`, method, data,
      header: { 'content-type': 'application/json', ...(currentToken ? { Authorization: `Bearer ${currentToken}` } : {}) },
      success(response) {
        if (response.statusCode >= 200 && response.statusCode < 300) return resolve(response.data)
        const sessionRejected = response.statusCode === 401 ||
          (response.statusCode === 403 && response.data && response.data.code === 'account_disabled')
        if (sessionRejected && currentToken && token() === currentToken) {
          wx.removeStorageSync(TOKEN_KEY)
          if (typeof getApp === 'function') {
            getApp().globalData.activeResult = null
            getApp().globalData.personalChart = null
          }
        }
        const err = new Error(response.data && response.data.error || `请求失败（${response.statusCode}）`)
        err.statusCode = response.statusCode
        err.code = response.data && response.data.code
        reject(err)
      },
      fail() { reject(new Error('网络连接失败，请检查网络和服务器地址')) },
    })
  })
}

function wxCode() {
  return new Promise((resolve, reject) => wx.login({
    success: ({ code }) => code ? resolve(code) : reject(new Error('未取得微信登录凭证')),
    fail: () => reject(new Error('微信登录失败，请重试')),
  }))
}

async function wechatLogin() {
  const response = await request('POST', '/auth/wechat/login', { code: await wxCode() })
  wx.setStorageSync(TOKEN_KEY, response.token)
  if (typeof getApp === 'function') getApp().globalData.personalChart = null
  return { user: response.user, isNewUser: response.is_new_user === true }
}

async function linkWebAccount(email, password) {
  const response = await request('POST', '/auth/wechat/link-web', { email, password }, true)
  wx.setStorageSync(TOKEN_KEY, response.token)
  if (typeof getApp === 'function') getApp().globalData.personalChart = null
  return response.user
}

function assetUrl(path) {
  if (!path || !path.startsWith('/static/uploads/user-avatars/')) return ''
  return (config.apiBaseUrl || '').replace(/\/$/, '') + path
}

function officialAccountQRUrl(path) {
  if (!/^\/static\/uploads\/official-account-qr\/[0-9a-fA-F-]{36}\.(?:png|jpg|webp)$/.test(path || '')) return ''
  return (config.apiBaseUrl || '').replace(/\/$/, '') + path
}

function avatarDisplayUrl(path) {
  const url = assetUrl(path)
  const base = (config.apiBaseUrl || '').replace(/\/$/, '')
  if (!url || !allowLanHTTP(base)) return Promise.resolve(url)
  const filename = path.split('/').pop()
  if (!/^[a-zA-Z0-9-]+\.(?:png|jpg|webp)$/.test(filename)) return Promise.resolve('')
  return new Promise((resolve, reject) => wx.request({
    url, method: 'GET', responseType: 'arraybuffer',
    success(response) {
      if (response.statusCode !== 200 || !(response.data instanceof ArrayBuffer)) {
        reject(new Error('读取头像失败'))
        return
      }
      const filePath = `${wx.env.USER_DATA_PATH}/avatar-${filename}`
      wx.getFileSystemManager().writeFile({
        filePath, data: response.data,
        success() { resolve(filePath) },
        fail() { reject(new Error('保存头像预览失败')) },
      })
    },
    fail() { reject(new Error('读取头像失败')) },
  }))
}

function uploadAvatar(filePath) {
  const base = (config.apiBaseUrl || '').replace(/\/$/, '')
  if (!/^https:\/\//.test(base) && !allowHTTP(base)) {
    return Promise.reject(new Error('请配置 HTTPS 后端地址或同一局域网的 HTTP IP；127.0.0.1 仅限开发者工具模拟器'))
  }
  const currentToken = token()
  if (!currentToken) return Promise.reject(new Error('请先登录'))
  return new Promise((resolve, reject) => wx.uploadFile({
    url: `${base}/api/user/profile/avatar`, filePath, name: 'file',
    header: { Authorization: `Bearer ${currentToken}` },
    success(response) {
      let body
      try { body = JSON.parse(response.data) } catch (_) { body = {} }
      if (response.statusCode >= 200 && response.statusCode < 300) return resolve(body)
      if (response.statusCode === 401 && token() === currentToken) wx.removeStorageSync(TOKEN_KEY)
      reject(new Error(body.error || `头像上传失败（${response.statusCode}）`))
    },
    fail() { reject(new Error('头像上传失败，请检查网络连接')) },
  }))
}

module.exports = {
  token, request, wechatLogin, linkWebAccount, assetUrl, avatarDisplayUrl, uploadAvatar, officialAccountQRUrl,
  officialAccountGuide: () => request('GET', '/miniapp/official-account-guide'),
  me: () => request('GET', '/auth/me', null, true),
  profile: () => request('GET', '/user/profile', null, true),
  updateNickname: nickname => request('PUT', '/user/profile/nickname', { nickname }, true),
  calculate: input => request('POST', '/bazi/calculate', input, true),
  history: page => request('GET', `/bazi/history?page=${page}`, null, true),
  detail: id => request('GET', `/bazi/history/${encodeURIComponent(id)}`, null, true),
  remove: id => request('DELETE', `/bazi/history/${encodeURIComponent(id)}`, null, true),
  personalChart: () => request('GET', '/user/personal-chart', null, true),
  setPersonalChart: chartId => request('PUT', '/user/personal-chart', { chart_id: chartId }, true),
  clearPersonalChart: () => request('DELETE', '/user/personal-chart', null, true),
  personalDay: date => request('GET', `/user/personal-day?date=${encodeURIComponent(date)}`, null, true),
  calendarDay: date => request('GET', `/calendar/day?date=${encodeURIComponent(date)}`),
  calendarRange: (start, end) => request('GET', `/calendar/range?start=${encodeURIComponent(start)}&end=${encodeURIComponent(end)}`),
  logout: () => {
    wx.removeStorageSync(TOKEN_KEY)
    if (typeof getApp === 'function') {
      getApp().globalData.activeResult = null
      getApp().globalData.personalChart = null
    }
  },
}
