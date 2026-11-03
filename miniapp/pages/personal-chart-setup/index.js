const api = require('../../utils/api')
const navigation = require('../../utils/navigation_feedback')

const hours = Array.from({ length: 24 }, (_, hour) => `${`${hour}`.padStart(2, '0')}:00`)

Page({
  data: { loading: true, loadError: '', saving: false, hasPersonalChart: false,
    birthDate: '1990-01-01', hourIndex: 0, gender: 'male', hours },
  onLoad() { this.loadPersonalChart() },
  onShow() { navigation.pageVisible(this, 'pages/personal-chart-setup/index') },
  onReady() { navigation.pageReady(this, 'pages/personal-chart-setup/index') },
  onHide() { navigation.pageHidden(this) },
  onUnload() { navigation.pageGone(this) },
  async loadPersonalChart() {
    const currentToken = api.token()
    if (!currentToken) { this.setData({ loading: false, loadError: '请先登录' }); return }
    this.setData({ loading: true, loadError: '' })
    try {
      const response = await api.personalChart()
      if (api.token() !== currentToken || this._navigationGone) return
      const chart = response.chart
      this.setData({ loading: false, hasPersonalChart: Boolean(chart),
        birthDate: chart ? `${chart.birth_year}-${`${chart.birth_month}`.padStart(2, '0')}-${`${chart.birth_day}`.padStart(2, '0')}` : '1990-01-01',
        hourIndex: chart ? chart.birth_hour : 0, gender: chart ? chart.gender : 'male' })
    } catch (err) {
      if (api.token() === currentToken && !this._navigationGone) this.setData({ loading: false, loadError: err.message })
    }
  },
  changeDate(event) { this.setData({ birthDate: event.detail.value }) },
  changeHour(event) { this.setData({ hourIndex: Number(event.detail.value) }) },
  changeGender(event) { this.setData({ gender: event.currentTarget.dataset.gender }) },
  async save() {
    if (this.data.loading || this.data.saving) return
    const currentToken = api.token()
    if (!currentToken) { wx.showToast({ title: '请先登录', icon: 'none' }); return }
    const [year, month, day] = this.data.birthDate.split('-').map(Number)
    this.setData({ saving: true })
    try {
      const response = await api.savePersonalChart({ year, month, day, hour: this.data.hourIndex,
        gender: this.data.gender, calendar_type: 'solar' })
      if (api.token() !== currentToken || this._navigationGone) return
      if (!response.chart || !response.chart.id) throw new Error('本人命盘保存失败，请重试')
      getApp().globalData.personalChart = response.chart
      wx.showToast({ title: '本人命盘已保存', icon: 'success' })
      navigation.navigate(this, { method: 'navigateBack', delta: 1, target: '/pages/mine/index', key: 'back',
        fallback: { method: 'switchTab', url: '/pages/mine/index' } })
    } catch (err) {
      if (api.token() === currentToken && !this._navigationGone) wx.showToast({ title: err.message, icon: 'none' })
    } finally {
      if (!this._navigationGone) this.setData({ saving: false })
    }
  },
})
