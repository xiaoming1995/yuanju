const api = require('../../utils/api')
const { toHistory } = require('../../utils/chart_view')
const { personalHints, personalDirectionRows } = require('../../utils/personal_day_view')
const navigation = require('../../utils/navigation_feedback')

function chinaTodayKey() {
  return new Date(Date.now() + 8 * 60 * 60 * 1000).toISOString().slice(0, 10)
}

function calendarPreview() {
  const key = chinaTodayKey()
  const [year, month, day] = key.split('-')
  const weekdays = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
  return {
    key,
    date: `${year}年${Number(month)}月${Number(day)}日 · ${weekdays[new Date(Date.UTC(Number(year), Number(month) - 1, Number(day))).getUTCDay()]}`,
    lunar: '',
    yearGanZhi: '—',
    monthGanZhi: '—',
    dayGanZhi: '—',
  }
}

Page({
  data: {
    mode: 'home',
    history: [],
    recentHistory: null,
    page: 1,
    hasMore: false,
    historyLoading: false,
    personalChartId: '',
    personalChartLoading: false,
    personalChartError: '',
    calendarPreview: calendarPreview(),
    calendarDay: null,
    calendarLoading: false,
    calendarError: '',
    personalDay: null,
    personalGuidance: [],
    personalDirections: [],
    personalLoading: false,
    personalError: '',
    personalState: 'unbound',
    detail: null,
  },

  onShow() {
    navigation.pageVisible(this, 'pages/home/index')
    const globalData = getApp().globalData
    if (globalData.previewHistoryNext) {
      globalData.previewHistoryNext = false
      this.setData({ mode: 'history' })
    }
    this.refreshToday()
    this.scheduleTodayRefresh()
    if (this.data.mode === 'history') {
      this.loadHistory(1)
      this.loadPersonalChart()
    }
    wx.setNavigationBarTitle({ title: this.data.mode === 'history' ? '我的排盘' : '' })
  },

  onReady() {
    navigation.pageReady(this, 'pages/home/index')
  },

  onHide() {
    clearTimeout(this._todayTimer)
    this._todayRequestId = (this._todayRequestId || 0) + 1
    navigation.pageHidden(this)
  },

  onUnload() {
    clearTimeout(this._todayTimer)
    this._todayRequestId = (this._todayRequestId || 0) + 1
    navigation.pageGone(this)
  },

  refreshToday() {
    const preview = calendarPreview()
    const date = preview.key
    const requestId = (this._todayRequestId || 0) + 1
    this._todayRequestId = requestId
    this.setData({ calendarPreview: preview, calendarDay: null, calendarLoading: true, calendarError: '',
      personalDay: null, personalGuidance: [], personalDirections: [], personalLoading: Boolean(api.token()),
      personalError: '', personalState: api.token() ? 'loading' : 'unbound' })
    this.loadTodayCalendar(date, requestId)
    if (api.token()) this.loadTodayPersonal(date, requestId, api.token())
  },

  scheduleTodayRefresh() {
    clearTimeout(this._todayTimer)
    const untilMidnight = 86400000 - ((Date.now() + 8 * 60 * 60 * 1000) % 86400000)
    this._todayTimer = setTimeout(() => { this.refreshToday(); this.scheduleTodayRefresh() }, untilMidnight + 1000)
    if (this._todayTimer && typeof this._todayTimer.unref === 'function') this._todayTimer.unref()
  },

  async loadTodayCalendar(date, requestId) {
    try {
      const response = await api.calendarDay(date)
      if (this._todayRequestId !== requestId || chinaTodayKey() !== date) return
      const day = response.data
      if (!day || day.date !== date || !Array.isArray(day.yi) || !Array.isArray(day.ji)) throw new Error('今日宜忌数据不完整')
      this.setData({ calendarDay: day, calendarPreview: { ...this.data.calendarPreview,
        lunar: day.lunar_full || this.data.calendarPreview.lunar,
        yearGanZhi: day.year_gan_zhi || '—', monthGanZhi: day.month_gan_zhi || '—',
        dayGanZhi: day.day_gan_zhi || '—' } })
    } catch (err) {
      if (this._todayRequestId === requestId) this.setData({ calendarError: err.message })
    } finally { if (this._todayRequestId === requestId) this.setData({ calendarLoading: false }) }
  },

  async loadTodayPersonal(date, requestId, token) {
    try {
      const response = await api.personalChart()
      if (this._todayRequestId !== requestId || api.token() !== token || chinaTodayKey() !== date) return
      if (!response.chart) { this.setData({ personalState: 'unbound' }); return }
      const dayResponse = await api.personalDay(date)
      if (this._todayRequestId !== requestId || api.token() !== token || chinaTodayKey() !== date) return
      const day = dayResponse.data
      if (!day || day.date !== date) throw new Error('个人每日资料不完整')
      this.setData({ personalDay: day, personalGuidance: personalHints(day),
        personalDirections: personalDirectionRows(day), personalState: 'ready' })
    } catch (err) {
      if (this._todayRequestId !== requestId) return
      if (api.token() !== token) { this.setData({ personalState: 'unbound', personalDay: null }); return }
      this.setData({ personalError: err.message, personalState: err.code === 'personal_chart_unbound' ? 'unbound' : 'error' })
    } finally { if (this._todayRequestId === requestId) this.setData({ personalLoading: false }) }
  },

  switchMode(event) {
    const mode = event.currentTarget.dataset.mode
    this.setData({ mode })
    if (mode === 'history') this.loadHistory(1)
    wx.setNavigationBarTitle({ title: mode === 'history' ? '我的排盘' : '' })
  },

  async loadHistory(page) {
    const currentToken = api.token()
    if (!currentToken) {
      this.setData({ history: [], recentHistory: null, hasMore: false })
      return
    }
    this.setData({ historyLoading: true })
    try {
      const response = await api.history(page)
      if (api.token() !== currentToken) return
      const next = page === 1 ? response.charts.map(toHistory) : this.data.history.concat(response.charts.map(toHistory))
      this.setData({ history: next, page, hasMore: next.length < response.total })
    } catch (err) {
      if (!api.token()) this.setData({ history: [], recentHistory: null, detail: null, hasMore: false })
      wx.showToast({ title: err.message, icon: 'none' })
    } finally {
      this.setData({ historyLoading: false })
    }
  },

  async loadPersonalChart() {
    if (!api.token()) { this.setData({ personalChartId: '', recentHistory: null, personalChartLoading: false, personalChartError: '' }); return }
    const currentToken = api.token()
    this.setData({ personalChartLoading: true, personalChartError: '' })
    try {
      const response = await api.personalChart()
      if (api.token() !== currentToken) return
      const chart = response.chart || null
      getApp().globalData.personalChart = chart
      this.setData({ personalChartId: chart ? chart.id : '', recentHistory: chart ? toHistory(chart) : null, personalChartError: '' })
    } catch (err) {
      if (api.token() !== currentToken) return
      this.setData({ personalChartError: err.message })
      wx.showToast({ title: err.message, icon: 'none' })
    } finally {
      if (api.token() === currentToken) this.setData({ personalChartLoading: false })
    }
  },

  bindPersonalChart(event) {
    const chartId = event.currentTarget.dataset.id
    if (!chartId) return
    wx.showModal({ title: '设为本人命盘', content: '请确认这份排盘使用的是你自己的出生资料。设置后可在日历查看每日干支参考。',
      success: async ({ confirm }) => {
        if (!confirm) return
        try {
          await api.setPersonalChart(chartId)
          await this.loadPersonalChart()
          wx.showToast({ title: '已绑定本人命盘', icon: 'success' })
        } catch (err) { wx.showToast({ title: err.message, icon: 'none' }) }
      } })
  },

  unbindPersonalChart() {
    wx.showModal({ title: '解绑本人命盘', content: '解绑后将停止显示个人每日参考，历史排盘不会删除。',
      success: async ({ confirm }) => {
        if (!confirm) return
        try { await api.clearPersonalChart(); getApp().globalData.personalChart = null
          this.setData({ personalChartId: '', recentHistory: null }); wx.showToast({ title: '已解绑', icon: 'none' }) }
        catch (err) { wx.showToast({ title: err.message, icon: 'none' }) }
      } })
  },

  loadMoreHistory() { if (this.data.hasMore && !this.data.historyLoading) this.loadHistory(this.data.page + 1) },

  openPersonalSetup() {
    navigation.navigate(this, { method: 'switchTab', url: '/pages/mine/index', key: 'mine' })
  },

  openCalendar() {
    if (this.data.navigationKey) return
    getApp().globalData.previewCalendarNext = this.data.calendarPreview.key
    navigation.navigate(this, { method: 'switchTab', url: '/pages/calendar/index', key: 'calendar' })
  },

  openTodayDetail() {
    navigation.navigate(this, { method: 'navigateTo', url: `/pages/personal-day/index?date=${this.data.calendarPreview.key}`, key: 'personal' })
  },

  openHistory() {
    this.setData({ mode: 'history' })
    this.loadHistory(1)
    wx.setNavigationBarTitle({ title: '我的排盘' })
  },

  async openHistoryDetail(event) {
    if (!this.data.personalChartId || event.currentTarget.dataset.id !== this.data.personalChartId) {
      wx.showToast({ title: '请先设为本人命盘', icon: 'none' })
      return
    }
    try {
      const response = await api.personalChartDetail()
      if (!response.chart || response.chart.id !== event.currentTarget.dataset.id) {
        this.loadPersonalChart()
        wx.showToast({ title: '本人命盘已更新，请重试', icon: 'none' })
        return
      }
      const detail = toHistory(response.chart)
      this.setData({ detail })
    } catch (err) {
      if (!api.token()) this.setData({ detail: null, history: [], recentHistory: null, hasMore: false })
      wx.showToast({ title: err.message, icon: 'none' })
    }
  },

  openDetailResult() {
    this.setData({ detail: null })
    navigation.navigate(this, { method: 'navigateTo', url: '/pages/result/index', key: 'result' })
  },

  closeHistoryDetail() {
    this.setData({ detail: null })
  },

  removeHistory(event) {
    const id = event.currentTarget.dataset.id
    wx.showModal({
      title: '删除干支',
      content: '确认删除这份干支记录？此操作无法撤销。',
      confirmColor: '#c38c24',
      success: async ({ confirm }) => {
        if (!confirm) return
        try { await api.remove(id); await this.loadHistory(1); await this.loadPersonalChart(); wx.showToast({ title: '已删除', icon: 'none' }) }
        catch (err) { wx.showToast({ title: err.message, icon: 'none' }) }
      },
    })
  },
})
