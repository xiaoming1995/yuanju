const api = require('../../utils/api')
const { toHistory } = require('../../utils/chart_view')
const calendarData = require('../../data/calendar')
const navigation = require('../../utils/navigation_feedback')

const years = Array.from({ length: 131 }, (_, index) => 1900 + index)
const months = Array.from({ length: 12 }, (_, index) => `${index + 1}`.padStart(2, '0'))
const hours = Array.from({ length: 24 }, (_, index) => `${index}`.padStart(2, '0'))

function daysFor(year, month) {
  return Array.from({ length: new Date(year, month, 0).getDate() }, (_, index) => `${index + 1}`.padStart(2, '0'))
}

function dateLabel(year, month, day, hour) {
  return `${year}-${`${month}`.padStart(2, '0')}-${`${day}`.padStart(2, '0')} ${`${hour}`.padStart(2, '0')}:00`
}

function calendarPreview() {
  const now = new Date()
  const currentKey = `${now.getFullYear()}-${`${now.getMonth() + 1}`.padStart(2, '0')}-${`${now.getDate()}`.padStart(2, '0')}`
  const isToday = Boolean(calendarData[currentKey])
  const key = isToday ? currentKey : '2026-10-22'
  const record = calendarData[key]
  const [year, month, day] = key.split('-')
  return {
    key,
    title: isToday ? '今日干支' : '干支历示例',
    date: `${year}年${Number(month)}月${Number(day)}日`,
    lunar: record.lunarFull,
    yearGanZhi: record.yearGanZhi,
    monthGanZhi: record.monthGanZhi,
    dayGanZhi: record.dayGanZhi,
  }
}

Page({
  data: {
    mode: 'query',
    dateText: '1990-01-01 01:00',
    note: '',
    gender: 'male',
    calculating: false,
    history: [],
    recentHistory: null,
    page: 1,
    hasMore: false,
    historyLoading: false,
    personalChartId: '',
    calendarPreview: calendarPreview(),
    showPicker: false,
    detail: null,
    years,
    months,
    days: daysFor(1990, 1),
    hours,
    pickerValue: [90, 0, 0, 1],
  },

  onShow() {
    navigation.pageVisible(this, 'pages/home/index')
    const globalData = getApp().globalData
    this.setData({
      calendarPreview: calendarPreview(),
    })
    if (globalData.previewQueryNext) {
      globalData.previewQueryNext = false
      globalData.previewHistoryNext = false
      globalData.previewHistoryDetailId = null
      this.setData({ mode: 'query' })
    } else if (globalData.previewHistoryNext) {
      globalData.previewHistoryNext = false
      this.setData({ mode: 'history' })
      if (globalData.previewHistoryDetailId) {
        const id = globalData.previewHistoryDetailId
        globalData.previewHistoryDetailId = null
        this.openHistoryDetail({ currentTarget: { dataset: { id } } })
      }
    }
    this.loadHistory(1)
    this.loadPersonalChart()
    wx.setNavigationBarTitle({ title: this.data.mode === 'history' ? '查询历史' : '干支排盘' })
  },

  onReady() {
    navigation.pageReady(this, 'pages/home/index')
  },

  onHide() {
    navigation.pageHidden(this)
  },

  onUnload() {
    navigation.pageGone(this)
  },

  switchMode(event) {
    const mode = event.currentTarget.dataset.mode
    this.setData({ mode })
    if (mode === 'history') this.loadHistory(1)
    wx.setNavigationBarTitle({ title: mode === 'history' ? '查询历史' : '干支排盘' })
  },

  onNoteInput(event) {
    this.setData({ note: event.detail.value })
  },

  setGender(event) { this.setData({ gender: event.currentTarget.dataset.gender }) },

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
      this.setData({ history: next, recentHistory: next[0] || null, page, hasMore: next.length < response.total })
    } catch (err) {
      if (!api.token()) this.setData({ history: [], recentHistory: null, detail: null, hasMore: false })
      wx.showToast({ title: err.message, icon: 'none' })
    } finally {
      this.setData({ historyLoading: false })
    }
  },

  async loadPersonalChart() {
    if (!api.token()) { this.setData({ personalChartId: '' }); return }
    const currentToken = api.token()
    this.setData({ personalChartId: '' })
    try {
      const response = await api.personalChart()
      if (api.token() !== currentToken) return
      const chart = response.chart || null
      getApp().globalData.personalChart = chart
      this.setData({ personalChartId: chart ? chart.id : '' })
    } catch (err) {
      this.setData({ personalChartId: '' })
      wx.showToast({ title: err.message, icon: 'none' })
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
          this.setData({ personalChartId: '' }); wx.showToast({ title: '已解绑', icon: 'none' }) }
        catch (err) { wx.showToast({ title: err.message, icon: 'none' }) }
      } })
  },

  loadMoreHistory() { if (this.data.hasMore && !this.data.historyLoading) this.loadHistory(this.data.page + 1) },

  openPicker() {
    this.setData({ showPicker: true })
  },

  closePicker() {
    this.setData({ showPicker: false })
  },

  onPickerChange(event) {
    const value = event.detail.value.slice()
    const year = years[value[0]]
    const month = value[1] + 1
    const days = daysFor(year, month)
    if (value[2] >= days.length) value[2] = days.length - 1
    this.setData({ pickerValue: value, days })
  },

  confirmPicker() {
    const [yearIndex, monthIndex, dayIndex, hourIndex] = this.data.pickerValue
    const dateText = dateLabel(years[yearIndex], monthIndex + 1, dayIndex + 1, hourIndex)
    this.setData({ dateText, showPicker: false })
  },

  async openDemoResult() {
    if (this.data.navigationKey || this.data.calculating) return
    if (!api.token()) { wx.showToast({ title: '请先在「我的」页面登录', icon: 'none' }); return }
    const match = this.data.dateText.match(/^(\d{4})-(\d{2})-(\d{2}) (\d{2}):00$/)
    if (!match) { wx.showToast({ title: '请检查出生日期', icon: 'none' }); return }
    const [, year, month, day, hour] = match
    const input = { year: +year, month: +month, day: +day, hour: +hour,
      gender: this.data.gender, calendar_type: 'solar', display_name: this.data.note }
    this.setData({ calculating: true })
    try {
      const response = await api.calculate(input)
      if (!response.chart_id) throw new Error('干支保存失败，请重试')
      getApp().globalData.activeResult = { result: response.result, chartId: response.chart_id,
        dateText: this.data.dateText, gender: this.data.gender, note: this.data.note }
      navigation.navigate(this, { method: 'navigateTo', url: '/pages/result/index', key: 'result' })
    } catch (err) {
      if (!api.token()) this.setData({ detail: null, history: [], recentHistory: null, hasMore: false })
      wx.showToast({ title: err.message, icon: 'none' })
    }
    finally { this.setData({ calculating: false }) }
  },

  openCalendar() {
    if (this.data.navigationKey) return
    getApp().globalData.previewCalendarNext = this.data.calendarPreview.key
    navigation.navigate(this, { method: 'switchTab', url: '/pages/calendar/index', key: 'calendar' })
  },

  openHistory() {
    this.setData({ mode: 'history' })
    this.loadHistory(1)
    wx.setNavigationBarTitle({ title: '查询历史' })
  },

  async openHistoryDetail(event) {
    try {
      const response = await api.detail(event.currentTarget.dataset.id)
      const detail = toHistory(response.chart)
      this.setData({ detail })
      getApp().globalData.activeResult = { result: response.result, chartId: detail.id,
        dateText: detail.date, gender: response.chart.gender, note: detail.name }
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
