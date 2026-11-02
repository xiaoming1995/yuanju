const { toResult } = require('../../utils/chart_view')
const navigation = require('../../utils/navigation_feedback')

Page({
  data: {
    selectedDate: '1990-01-01 01:00',
    selectedNote: '',
    chart: null,
    pillars: [],
    basicRows: [],
  },

  onLoad() {
    navigation.pageVisible(this, 'pages/result/index')
    const query = getApp().globalData.activeResult
    if (query && query.result) {
      const chart = toResult(query.result, query)
      this.setData({
        selectedDate: query.dateText,
        selectedNote: query.note,
        chart,
        pillars: chart.pillars.map(p => ({ label: p.label.slice(0, 1), gan: p.gan, zhi: p.zhi, ganTone: p.ganTone, zhiTone: p.zhiTone })),
        basicRows: chart.basicRows,
      })
    } else {
      wx.showToast({ title: '干支数据已失效，请重新查询', icon: 'none' })
      setTimeout(() => wx.navigateBack({ delta: 1 }), 800)
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
