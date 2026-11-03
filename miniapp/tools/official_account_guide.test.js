const assert = require('node:assert/strict')
const test = require('node:test')

function minePage() {
  let definition
  global.Page = value => { definition = value }
  delete require.cache[require.resolve('../pages/mine/index')]
  require('../pages/mine/index')
  return { ...definition, data: { ...definition.data }, setData(patch) { Object.assign(this.data, patch) } }
}

const active = { enabled: true, name: '缘聚', wechat_id: 'yuanju', description: '内容更新',
  qr_path: '/static/uploads/official-account-qr/550e8400-e29b-41d4-a716-446655440000.png' }

test('Mine loads the public guide for guests independently of profile', async () => {
  const api = require('../utils/api')
  const navigation = require('../utils/navigation_feedback')
  const old = { guide: api.officialAccountGuide, url: api.officialAccountQRUrl, visible: navigation.pageVisible }
  api.officialAccountGuide = async () => ({ data: active })
  api.officialAccountQRUrl = path => 'https://api.example' + path
  navigation.pageVisible = () => {}
  try {
    const page = minePage()
    page.loadProfile = () => new Promise(() => {})
    page.onShow()
    await new Promise(resolve => setImmediate(resolve))
    assert.equal(page.data.officialAccountGuide.name, '缘聚')
    assert.equal(page.data.officialAccountGuide.qrUrl, 'https://api.example' + active.qr_path)
  } finally { api.officialAccountGuide = old.guide; api.officialAccountQRUrl = old.url; navigation.pageVisible = old.visible }
})

test('disabled, failed and stale responses remove old guide', async () => {
  const api = require('../utils/api')
  const old = { guide: api.officialAccountGuide, url: api.officialAccountQRUrl }
  api.officialAccountQRUrl = path => 'https://api.example' + path
  try {
    const page = minePage()
    api.officialAccountGuide = async () => ({ data: active })
    await page.loadOfficialAccountGuide()
    assert.equal(page.data.officialAccountGuide.name, '缘聚')
    api.officialAccountGuide = async () => ({ data: { enabled: false } })
    await page.loadOfficialAccountGuide()
    assert.equal(page.data.officialAccountGuide, null)
    api.officialAccountGuide = async () => { throw new Error('offline') }
    await page.loadOfficialAccountGuide()
    assert.equal(page.data.officialAccountGuide, null)

    let resolveSlow
    api.officialAccountGuide = () => new Promise(resolve => { resolveSlow = resolve })
    const slow = page.loadOfficialAccountGuide()
    api.officialAccountGuide = async () => ({ data: active })
    await page.loadOfficialAccountGuide()
    resolveSlow({ data: { enabled: false } })
    await slow
    assert.equal(page.data.officialAccountGuide.name, '缘聚')
  } finally { api.officialAccountGuide = old.guide; api.officialAccountQRUrl = old.url }
})

test('QR sheet previews image, copies search ID and survives image failure', async () => {
  const api = require('../utils/api')
  const old = { guide: api.officialAccountGuide, url: api.officialAccountQRUrl }
  api.officialAccountGuide = async () => ({ data: active })
  api.officialAccountQRUrl = path => 'https://api.example' + path
  const calls = { previews: [], copies: [] }
  global.wx = { previewImage: options => calls.previews.push(options), setClipboardData: options => calls.copies.push(options.data), showToast() {} }
  try {
    const page = minePage()
    await page.loadOfficialAccountGuide()
    page.openOfficialAccountGuide()
    assert.equal(page.data.officialAccountOpen, true)
    page.previewOfficialAccountQR()
    assert.equal(calls.previews[0].current, 'https://api.example' + active.qr_path)
    page.officialAccountImageError()
    assert.equal(page.data.officialAccountQRFailed, true)
    page.previewOfficialAccountQR()
    assert.equal(calls.previews.length, 1)
    page.copyOfficialAccountID()
    assert.deepEqual(calls.copies, ['yuanju'])
    page.closeOfficialAccountGuide()
    assert.equal(page.data.officialAccountOpen, false)
  } finally { api.officialAccountGuide = old.guide; api.officialAccountQRUrl = old.url }
})

test('unsafe QR paths are rejected before rendering', async () => {
  const api = require('../utils/api')
  const previous = api.officialAccountQRUrl
  try {
    assert.equal(api.officialAccountQRUrl('https://bad.example/qr.png'), '')
    assert.equal(api.officialAccountQRUrl('/static/uploads/official-account-qr/../user-avatars/abc.png'), '')
  } finally { api.officialAccountQRUrl = previous }
})
