const assert = require('node:assert/strict')
const test = require('node:test')
const fonts = require('../utils/font_size')

test('follows WeChat setting and bounds extreme values', () => {
  let saved
  global.wx = {
    getStorageSync: () => saved,
    getAppBaseInfo: () => ({ fontSizeSetting: 20 }),
    setStorageSync: (_, value) => { saved = value },
  }
  assert.deepEqual(fonts.current(), { mode: 'follow', label: '跟随微信', step: 5 })
  wx.getAppBaseInfo = () => ({ fontSizeSetting: 50 })
  assert.equal(fonts.current().step, 7)
  wx.getAppBaseInfo = () => ({ fontSizeSetting: 12 })
  assert.equal(fonts.current().step, 0)
  assert.equal(fonts.setMode('large'), true)
  assert.equal(fonts.current().step, 3)
  assert.equal(fonts.setMode('invalid'), false)
  assert.equal(fonts.current().mode, 'large')
})

test('falls back when WeChat font information is unavailable', () => {
  global.wx = { getStorageSync: () => 'follow', getAppBaseInfo: () => { throw Error('unavailable') } }
  assert.equal(fonts.current().step, 0)
  const page = { data: {}, setData(value) { Object.assign(this.data, value) } }
  fonts.apply(page)
  assert.equal(page.data.fontModeLabel, '跟随微信')
  assert.equal(page.data.fontStep, 0)
})

test('Mine choice applies immediately and persists without login', () => {
  let saved = ''
  let definition
  global.wx = {
    getStorageSync: () => saved,
    setStorageSync: (_, value) => { saved = value },
    getAppBaseInfo: () => ({ fontSizeSetting: 20 }),
    showActionSheet: ({ itemList, success }) => {
      assert.deepEqual(itemList, ['跟随微信 ✓', '标准', '较大', '特大'])
      success({ tapIndex: 3 })
    },
  }
  global.Page = value => { definition = value }
  delete require.cache[require.resolve('../pages/mine/index')]
  require('../pages/mine/index')
  const page = { ...definition, data: { ...definition.data }, setData(value) { Object.assign(this.data, value) } }
  fonts.apply(page)
  page.chooseFontSize()
  assert.equal(saved, 'extra')
  assert.equal(page.data.fontStep, 6)
  assert.equal(page.data.fontModeLabel, '特大')
  assert.equal(fonts.current().mode, 'extra')
})
