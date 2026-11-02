// Development uses the local backend; trial and release use the public HTTPS origin.
// Update the LAN IP when the development computer's address changes.
// No secrets belong in this file.
const productionApiBaseUrl = 'https://www.guanshilu.cn'
const lanApiBaseUrl = 'http://192.168.101.24:9002'
let apiBaseUrl = productionApiBaseUrl
let envVersion = ''
let platform = ''
try {
  if (typeof wx !== 'undefined' && wx.getAccountInfoSync) {
    envVersion = wx.getAccountInfoSync().miniProgram.envVersion
  }
} catch (_) {
  // Unknown environments keep the production URL unless running in developer tools.
}
try {
  if (typeof wx !== 'undefined' && wx.getDeviceInfo) {
    platform = wx.getDeviceInfo().platform
  }
} catch (_) {
  // Device detection is optional.
}
if (envVersion === 'develop' || (!envVersion && platform === 'devtools')) {
  apiBaseUrl = platform === 'devtools' ? 'http://127.0.0.1:9002' : lanApiBaseUrl
}

module.exports = { apiBaseUrl }
