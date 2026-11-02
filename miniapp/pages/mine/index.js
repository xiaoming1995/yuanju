const api = require('../../utils/api')
const { toHistory } = require('../../utils/chart_view')
const navigation = require('../../utils/navigation_feedback')

Page({
  data: {
    loggedIn: false, loggingIn: false, user: null,
    email: '', password: '', bindingVisible: false, chartCount: 0, recentCharts: [],
    personalChart: null, personalChartError: '', personalChartLoading: false,
    userAvatarUrl: '',
  },
  onShow() { navigation.pageVisible(this, 'pages/mine/index'); return this.loadProfile() },
  onReady() { navigation.pageReady(this, 'pages/mine/index') },
  onHide() { navigation.pageHidden(this) },
  onUnload() { navigation.pageGone(this) },

  async loadProfile() {
    const loadSeq = this._profileLoadSeq = (this._profileLoadSeq || 0) + 1
    const currentToken = api.token()
    if (!currentToken) { this.setData({ loggedIn: false, user: null, userAvatarUrl: '',
      chartCount: 0, recentCharts: [], personalChart: null, personalChartError: '', personalChartLoading: false }); return false }
    this.setData({ personalChart: null, personalChartError: '', personalChartLoading: true })
    try {
      const bindingRequest = api.personalChart().then(value => ({ value }), error => ({ error }))
      const response = await api.profile()
      if (api.token() !== currentToken || this._profileLoadSeq !== loadSeq) return false
      const profile = response.data
      const history = (profile.recent_charts || []).map(toHistory)
      const user = profile.user
      const avatarUrl = api.assetUrl(user.avatar_url) || ''
      this.setData({ loggedIn: true, user: profile.user, chartCount: profile.stats.chart_count,
        recentCharts: history.slice(0, 2),
        userAvatarUrl: avatarUrl.startsWith('http://') ? '' : avatarUrl })
      api.avatarDisplayUrl(user.avatar_url).then(displayUrl => {
        if (api.token() === currentToken && this._profileLoadSeq === loadSeq) this.setData({ userAvatarUrl: displayUrl })
      }, () => {})
      const binding = await bindingRequest
      if (api.token() !== currentToken || this._profileLoadSeq !== loadSeq) return false
      this.setData({
        personalChart: binding.value && binding.value.chart ? toHistory(binding.value.chart) : null,
        personalChartError: binding.error ? binding.error.message : '',
        personalChartLoading: false })
      return true
    } catch (err) {
      if (this._profileLoadSeq !== loadSeq) return false
      if (!api.token()) this.setData({ loggedIn: false, user: null, userAvatarUrl: '',
        chartCount: 0, recentCharts: [], personalChart: null, personalChartError: '', personalChartLoading: false })
      else this.setData({ personalChartLoading: false, personalChartError: err.message })
      wx.showToast({ title: err.message, icon: 'none' })
      return false
    }
  },
  onEmailInput(event) { this.setData({ email: event.detail.value }) },
  onPasswordInput(event) { this.setData({ password: event.detail.value }) },
  async openLogin() {
    if (this.data.loggingIn) return
    this.setData({ loggingIn: true })
    try {
      const login = await api.wechatLogin()
      if (await this.loadProfile()) {
        if (login.isNewUser && this.data.user && this.data.user.source === 'miniapp_wechat') {
          this.openProfile(true)
        } else wx.showToast({ title: '登录成功', icon: 'success' })
      }
    } catch (err) {
      wx.showToast({ title: err.message, icon: 'none', duration: 3000 })
    }
    finally { this.setData({ loggingIn: false }) }
  },
  toggleWebBinding() { this.setData({ bindingVisible: !this.data.bindingVisible, password: '' }) },
  async bindWebAccount() {
    if (this.data.loggingIn) return
    if (!this.data.email || !this.data.password) { wx.showToast({ title: '请填写网页账号和密码', icon: 'none' }); return }
    this.setData({ loggingIn: true })
    try {
      await api.linkWebAccount(this.data.email, this.data.password)
      this.setData({ email: '', password: '', bindingVisible: false })
      if (await this.loadProfile()) wx.showToast({ title: '网页账号绑定成功', icon: 'success' })
    } catch (err) { wx.showToast({ title: err.message, icon: 'none', duration: 3000 }) }
    finally { this.setData({ loggingIn: false }) }
  },
  openProfile(onboarding) {
    if (!this.data.user || this.data.navigationKey) return
    navigation.navigate(this, { method: 'navigateTo',
      url: `/pages/profile/index${onboarding === true ? '?onboarding=1' : ''}`, key: 'profile' })
  },
  openPersonalChart() { if (this.data.personalChart) this.openRecentChart({ currentTarget: { dataset: { id: this.data.personalChart.id } } }) },
  openPersonalBanner() { if (this.data.personalChartError) return this.loadProfile(); this.openHistory() },
  openRecentChart(event) {
    if (this.data.navigationKey) return
    getApp().globalData.previewHistoryDetailId = event.currentTarget.dataset.id
    this.openHistory()
  },
  logoutPreview() {
    api.logout()
    this.setData({ loggedIn: false, user: null, userAvatarUrl: '', chartCount: 0, recentCharts: [], personalChart: null, personalChartError: '', personalChartLoading: false })
    wx.showToast({ title: '已退出登录', icon: 'none' })
  },
  openHistory() {
    getApp().globalData.previewHistoryNext = true
    navigation.navigate(this, { method: 'switchTab', url: '/pages/home/index', key: 'history' })
  },
  openTool(event) { wx.showToast({ title: `${event.currentTarget.dataset.title}尚未接入`, icon: 'none' }) },
  openPrivacy() { wx.showModal({ title: '隐私说明', content: '登录及干支记录会提交到配置的后端服务。请在正式发布前提供完整隐私政策。', showCancel: false }) },
  clearDemo() { wx.showModal({ title: '演示数据', content: '月历标签与择日参考仍使用随包示例；主动绑定本人命盘后，每日干支参考由后端计算。', showCancel: false }) },
  openAbout() { wx.showModal({ title: '关于缘聚', content: '提供干支排盘、日历宜忌、个人参考与历史记录。', showCancel: false }) },
})
