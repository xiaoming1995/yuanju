const { chart, pillars, basicRows } = require('../../data/demo')
const navigation = require('../../utils/navigation_feedback')

Page({
  data: {
    selectedDate: '1990-01-01 01:00',
    selectedNote: '',
    chart,
    pillars,
    basicRows,
  },

  onLoad() {
    navigation.pageVisible(this, 'pages/result/index')
    const query = getApp().globalData.previewQuery
    if (query) {
      this.setData({
        selectedDate: query.dateText,
        selectedNote: query.note,
      })
    }
  },

  onShow() {
    navigation.pageVisible(this, 'pages/result/index')
  },

  onReady() {
    navigation.pageReady(this, 'pages/result/index')
  },

  onHide() {
    navigation.pageHidden(this)
  },

  onUnload() {
    navigation.pageGone(this)
  },

  returnToQuery() {
    navigation.navigate(this, {
      method: 'navigateBack', delta: 1, target: '/pages/home/index', key: 'back',
      fallback: { method: 'switchTab', url: '/pages/home/index' },
    })
  },
})
