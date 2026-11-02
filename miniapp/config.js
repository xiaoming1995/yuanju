// Local testing: the simulator uses loopback; the phone uses the computer's LAN IP.
// Replace the LAN IP when it changes; use a registered HTTPS API origin for release.
// No secrets belong in this file.
const lanApiBaseUrl = 'http://192.168.101.24:9002'
let apiBaseUrl = lanApiBaseUrl
try {
  if (typeof wx !== 'undefined' && wx.getDeviceInfo && wx.getDeviceInfo().platform === 'devtools') {
    apiBaseUrl = 'http://127.0.0.1:9002'
  }
} catch (_) {
  // Device detection is optional; keep the LAN address when it is unavailable.
}

module.exports = { apiBaseUrl }
