const { history } = require('../../data/demo')
const navigation = require('../../utils/navigation_feedback')

Page({
  data: {
    previewLoggedIn: false,
    demoUser: { nickname: '演示用户', email: 'demo@example.com' },
    chartCount: history.length,
    latestChart: history[0],
    recentCharts: history.slice(0, 2),
  },

  onShow() {
    navigation.pageVisible(this, 'pages/mine/index')
    const currentHistory = getApp().globalData.previewHistory || history
    this.setData({
      chartCount: currentHistory.length,
      latestChart: currentHistory[0] || null,
      recentCharts: currentHistory.slice(0, 2),
    })
  },

  onReady() {
    navigation.pageReady(this, 'pages/mine/index')
  },

  onHide() {
    navigation.pageHidden(this)
  },

  onUnload() {
    navigation.pageGone(this)
  },

  openLogin() {
    this.setData({ previewLoggedIn: true })
    wx.showToast({ title: '已进入登录态演示', icon: 'none' })
  },

  openProfile() {
    wx.showModal({
      title: '演示账号',
      content: '这是登录后的界面预览，昵称、邮箱和统计均为示例，尚未连接真实账号。',
      showCancel: false,
      confirmColor: '#c88e22',
    })
  },

  openLatestChart() {
    if (this.data.navigationKey) return
    if (!this.data.latestChart) return
    getApp().globalData.previewHistoryDetailId = this.data.latestChart.id
    this.openHistory()
  },

  openRecentChart(event) {
    if (this.data.navigationKey) return
    getApp().globalData.previewHistoryDetailId = event.currentTarget.dataset.id
    this.openHistory()
  },

  logoutPreview() {
    this.setData({ previewLoggedIn: false })
    wx.pageScrollTo({ scrollTop: 0, duration: 200 })
    wx.showToast({ title: '已退出登录态演示', icon: 'none' })
  },

  openHistory() {
    if (this.data.navigationKey) return
    getApp().globalData.previewHistoryNext = true
    navigation.navigate(this, { method: 'switchTab', url: '/pages/home/index', key: 'history' })
  },

  openTool(event) {
    const title = event.currentTarget.dataset.title
    wx.showToast({ title: `${title}尚未接入，当前为界面演示`, icon: 'none', duration: 2100 })
  },

  openPrivacy() {
    wx.showModal({
      title: '隐私说明 · 原型',
      content: '本原型使用内置演示数据，不连接账号服务，也不提交生辰信息。正式版需单独提供完整隐私政策。',
      showCancel: false,
      confirmColor: '#c88e22',
    })
  },

  clearDemo() {
    wx.showModal({
      title: '本地演示数据',
      content: '本原型未保存个人数据。历史卡片只在当前页面临时变化，重新打开会恢复示例。',
      showCancel: false,
      confirmColor: '#c88e22',
    })
  },

  openAbout() {
    wx.showModal({
      title: '关于缘聚小程序原型',
      content: '根据界面展示稿制作的前端视觉原型。排盘、账号、日历同步与分享功能尚未接入。',
      showCancel: false,
      confirmColor: '#c88e22',
    })
  },
})
