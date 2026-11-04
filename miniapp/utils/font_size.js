const STORAGE_KEY = 'yj_miniapp_font_mode'
const MODES = ['follow', 'standard', 'large', 'extra']
const LABELS = ['跟随微信', '标准', '较大', '特大']
const FIXED_SCALES = { standard: 1, large: 1.15, extra: 1.3 }

function getMode() {
  try {
    const saved = wx.getStorageSync(STORAGE_KEY)
    return MODES.includes(saved) ? saved : 'follow'
  } catch (_) { return 'follow' }
}

function setMode(mode) {
  if (!MODES.includes(mode)) return false
  try { wx.setStorageSync(STORAGE_KEY, mode) } catch (_) { return false }
  return true
}

function wechatScale() {
  try {
    const info = typeof wx.getAppBaseInfo === 'function'
      ? wx.getAppBaseInfo()
      : typeof wx.getSystemInfoSync === 'function' ? wx.getSystemInfoSync() : null
    const size = Number(info && info.fontSizeSetting)
    if (!Number.isFinite(size) || size <= 0) return 1
    return Math.min(1.35, Math.max(1, size / 16))
  } catch (_) { return 1 }
}

function current() {
  const mode = getMode()
  const scale = mode === 'follow' ? wechatScale() : FIXED_SCALES[mode]
  return { mode, label: LABELS[MODES.indexOf(mode)], step: Math.round((scale - 1) * 20) }
}

function apply(page) {
  const setting = current()
  if (page.data.fontStep !== setting.step || page.data.fontMode !== setting.mode) {
    page.setData({ fontStep: setting.step, fontMode: setting.mode, fontModeLabel: setting.label })
  }
  return setting
}

module.exports = { MODES, LABELS, getMode, setMode, wechatScale, current, apply }
