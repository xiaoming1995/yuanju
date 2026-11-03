const api = require('../../utils/api')
const { toHistory } = require('../../utils/chart_view')
const navigation = require('../../utils/navigation_feedback')

Page({
  data: {
    loggedIn: false, loggingIn: false, profileLoading: false, profileError: '', user: null,
    email: '', password: '', bindingVisible: false, chartCount: 0,
    personalChart: null, personalChartError: '', personalChartLoading: false,
    userAvatarUrl: '',
    officialAccountGuide: null, officialAccountOpen: false, officialAccountQRFailed: false,
  },
  onShow() {
    navigation.pageVisible(this, 'pages/mine/index')
    this.loadOfficialAccountGuide()
    return this.loadProfile()
  },
  onReady() { navigation.pageReady(this, 'pages/mine/index') },
  onHide() { navigation.pageHidden(this) },
  onUnload() { navigation.pageGone(this) },

  async loadOfficialAccountGuide() {
    const seq = this._guideLoadSeq = (this._guideLoadSeq || 0) + 1
    this.setData({ officialAccountGuide: null, officialAccountOpen: false, officialAccountQRFailed: false })
    try {
      const response = await api.officialAccountGuide()
      if (seq !== this._guideLoadSeq) return false
      const guide = response && response.data
      const qrUrl = api.officialAccountQRUrl(guide && guide.qr_path)
      if (!guide || !guide.enabled || !guide.name || !guide.wechat_id || !guide.description || !qrUrl) return false
      this.setData({ officialAccountGuide: { ...guide, qrUrl } })
      return true
    } catch (_) { return false }
  },
  openOfficialAccountGuide() {
    if (this.data.officialAccountGuide) this.setData({ officialAccountOpen: true, officialAccountQRFailed: false })
  },
  closeOfficialAccountGuide() { this.setData({ officialAccountOpen: false }) },
  stopOfficialAccountTap() {},
  officialAccountImageError() { this.setData({ officialAccountQRFailed: true }) },
  previewOfficialAccountQR() {
    const guide = this.data.officialAccountGuide
    if (!guide || this.data.officialAccountQRFailed) return
    wx.previewImage({ current: guide.qrUrl, urls: [guide.qrUrl] })
  },
  copyOfficialAccountID() {
    const guide = this.data.officialAccountGuide
    if (!guide) return
    wx.setClipboardData({ data: guide.wechat_id, fail: () => wx.showToast({ title: '复制失败，请重试', icon: 'none' }) })
  },

  async loadProfile() {
    const loadSeq = this._profileLoadSeq = (this._profileLoadSeq || 0) + 1
    const currentToken = api.token()
    if (!currentToken) { this.setData({ loggedIn: false, profileLoading: false, profileError: '', user: null, userAvatarUrl: '',
      chartCount: 0, personalChart: null, personalChartError: '', personalChartLoading: false }); return false }
    this.setData({ profileLoading: !this.data.loggedIn, profileError: '', personalChart: null, personalChartError: '', personalChartLoading: true })
    try {
      const bindingRequest = api.personalChart().then(value => ({ value }), error => ({ error }))
      const response = await api.profile()
      if (api.token() !== currentToken || this._profileLoadSeq !== loadSeq) return false
      const profile = response.data
      const user = profile.user
      const avatarUrl = api.assetUrl(user.avatar_url) || ''
      this.setData({ loggedIn: true, profileLoading: false, profileError: '', user: profile.user, chartCount: profile.stats.chart_count,
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
      const stillAuthenticated = api.token() === currentToken
      if (!stillAuthenticated && api.token()) return false
      this.setData({ loggedIn: false, profileLoading: false, profileError: stillAuthenticated ? (err.message || '个人资料读取失败') : '',
        user: null, userAvatarUrl: '', chartCount: 0, personalChart: null,
        personalChartError: '', personalChartLoading: false })
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
  openPersonalChart() {
    if (!this.data.personalChart) return
    navigation.navigate(this, { method: 'navigateTo', url: '/pages/result/index', key: 'result' })
  },
  openPersonalBanner() {
    if (this.data.personalChartError) return this.loadProfile()
    if (!this.data.chartCount) return this.openCreateChart()
    this.openHistory()
  },
  logoutPreview() {
    api.logout()
    this._profileLoadSeq = (this._profileLoadSeq || 0) + 1
    this.setData({ loggedIn: false, profileLoading: false, profileError: '', user: null, userAvatarUrl: '',
      chartCount: 0, personalChart: null, personalChartError: '', personalChartLoading: false,
      email: '', password: '', bindingVisible: false })
    wx.showToast({ title: '已退出登录', icon: 'none' })
  },
  openHistory() {
    getApp().globalData.previewHistoryNext = true
    navigation.navigate(this, { method: 'switchTab', url: '/pages/home/index', key: 'history' })
  },
  openCreateChart() {
    navigation.navigate(this, { method: 'navigateTo', url: '/pages/personal-chart-setup/index', key: 'setup' })
  },
  openPrivacy() { wx.showModal({ title: '隐私说明', content: '登录及干支记录会提交到配置的后端服务。正式隐私政策尚未在小程序内提供。', showCancel: false }) },
  openAbout() { wx.showModal({ title: '关于缘聚', content: '提供本人命盘、日历宜忌与个人参考。', showCancel: false }) },
})
