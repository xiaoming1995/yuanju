const calendarData = require('../../data/calendar')
const api = require('../../utils/api')
const { chart } = require('../../data/demo')
const { buildPersonalDay } = require('../../data/personal_day')
const { getJianChuInfo } = require('../../data/jianchu')
const { personalHints } = require('../../utils/personal_day_view')
const navigation = require('../../utils/navigation_feedback')

const weekNames = ['星期日', '星期一', '星期二', '星期三', '星期四', '星期五', '星期六']

function keyOf(year, month, day) {
  return `${year}-${`${month}`.padStart(2, '0')}-${`${day}`.padStart(2, '0')}`
}

function chinaTodayKey() {
  return new Date(Date.now() + 8 * 60 * 60 * 1000).toISOString().slice(0, 10)
}

function monthView(year, month, selectedKey) {
  const firstWeekday = new Date(year, month - 1, 1).getDay()
  const daysInMonth = new Date(year, month, 0).getDate()
  const cellCount = Math.ceil((firstWeekday + daysInMonth) / 7) * 7
  const todayKey = chinaTodayKey()
  const cells = []
  for (let index = 0; index < cellCount; index++) {
    const date = new Date(year, month - 1, index - firstWeekday + 1)
    const key = keyOf(date.getFullYear(), date.getMonth() + 1, date.getDate())
    const label = calendarData[key] || {}
    cells.push({
      key,
      day: date.getDate(),
      lunar: label.lunar || '—',
      jieQi: label.jieQi || '',
      otherMonth: date.getMonth() + 1 !== month,
      selected: key === selectedKey,
      today: key === todayKey,
    })
  }
  const terms = []
  for (let day = 1; day <= daysInMonth; day++) {
    const key = keyOf(year, month, day)
    if (calendarData[key] && calendarData[key].jieQi) {
      const date = new Date(year, month - 1, day)
      terms.push({
        key,
        name: calendarData[key].jieQi,
        day,
        weekday: weekNames[date.getDay()],
        lunarFull: calendarData[key].lunarFull,
      })
    }
  }
  return { cells, terms, termSummary: terms.map(term => term.name).join(' · ') || '暂无节气' }
}

function selectedInfo(key) {
  const [year, month, day] = key.split('-').map(Number)
  const record = calendarData[key] || {}
  const date = new Date(year, month - 1, day)
  return {
    date: key,
    weekday: weekNames[date.getDay()],
    lunarFull: record.lunarFull || '暂无本地农历资料',
    yearGanZhi: record.yearGanZhi || '—',
    monthGanZhi: record.monthGanZhi || '—',
    dayGanZhi: record.dayGanZhi || '—',
    jianChu: record.jianChu || '',
    jianChuInfo: getJianChuInfo(record.jianChu),
    jieQi: record.jieQi || '',
    isToday: key === chinaTodayKey(),
  }
}

const initialKey = chinaTodayKey()
const [initialYear, initialMonth] = initialKey.split('-').map(Number)
const initialView = monthView(initialYear, initialMonth, initialKey)

Page({
  data: {
    year: initialYear,
    month: initialMonth,
    selectedKey: initialKey,
    selected: selectedInfo(initialKey),
    cells: initialView.cells,
    terms: initialView.terms,
    termSummary: initialView.termSummary,
    dayPreviewEnabled: false,
    personalDay: buildPersonalDay(initialKey, calendarData[initialKey], chart),
    personalChart: null,
    calendarDay: null,
    calendarLoading: false,
    calendarError: '',
    realDay: null,
    personalHints: [],
    personalHint: null,
    personalHintRemaining: 0,
    realLoading: false,
    realError: '',
    selectedByUser: false,
    infoTab: 'date',
    weekShort: ['日', '一', '二', '三', '四', '五', '六'],
  },

  onShow() {
    navigation.pageVisible(this, 'pages/calendar/index')
    const dayPreviewEnabled = Boolean(getApp().globalData.previewDayProfileEnabled)
    const selectedKey = getApp().globalData.previewCalendarNext
    if (!selectedKey) {
      if (!this.data.selectedByUser && this.data.selectedKey !== chinaTodayKey()) {
        this.showDate(chinaTodayKey(), false)
      }
      this.setData({ dayPreviewEnabled })
      this.refreshCalendarDay()
      this.refreshPersonalState()
      return
    }
    getApp().globalData.previewCalendarNext = null
    this.showDate(selectedKey, true)
    this.setData({ dayPreviewEnabled })
    this.refreshCalendarDay()
    this.refreshPersonalState()
  },

  showDate(selectedKey, selectedByUser = true) {
    this._dailyRequestId = (this._dailyRequestId || 0) + 1
    this._calendarRequestId = (this._calendarRequestId || 0) + 1
    const [year, month] = selectedKey.split('-').map(Number)
    const view = monthView(year, month, selectedKey)
    this.setData({ year, month, selectedKey, selected: selectedInfo(selectedKey), cells: view.cells,
      terms: view.terms, termSummary: view.termSummary, selectedByUser,
      personalDay: buildPersonalDay(selectedKey, calendarData[selectedKey], chart),
      calendarDay: null,
      calendarError: '', realDay: null, personalHints: [], personalHint: null,
      personalHintRemaining: 0, realError: '' })
  },

  async refreshCalendarDay() {
    const date = this.data.selectedKey
    const requestId = (this._calendarRequestId || 0) + 1
    this._calendarRequestId = requestId
    this.setData({ calendarLoading: true, calendarError: '', calendarDay: null })
    try {
      const response = await api.calendarDay(date)
      if (this.data.selectedKey !== date || this._calendarRequestId !== requestId) return
      const day = response.data
      if (!day || day.date !== date || !Array.isArray(day.yi) || !Array.isArray(day.ji)) {
        throw new Error('黄历宜忌数据不完整')
      }
      this.setData({ calendarDay: day, selected: { ...this.data.selected,
        lunarFull: day.lunar_full || this.data.selected.lunarFull,
        yearGanZhi: day.year_gan_zhi || this.data.selected.yearGanZhi,
        monthGanZhi: day.month_gan_zhi || this.data.selected.monthGanZhi,
        dayGanZhi: day.day_gan_zhi || this.data.selected.dayGanZhi } })
    } catch (err) {
      if (this.data.selectedKey === date && this._calendarRequestId === requestId) this.setData({ calendarError: err.message })
    } finally { if (this._calendarRequestId === requestId) this.setData({ calendarLoading: false }) }
  },

  async refreshPersonalState() {
    if (typeof wx.getStorageSync !== 'function' || !api.token()) {
      this._dailyRequestId = (this._dailyRequestId || 0) + 1
      this.setData({ personalChart: null, realDay: null, personalHints: [], personalHint: null,
        personalHintRemaining: 0, realError: '', realLoading: false })
      return
    }
    const currentToken = api.token()
    this.setData({ realLoading: true, realError: '', realDay: null, personalHints: [],
      personalHint: null, personalHintRemaining: 0, personalChart: null })
    try {
      const response = await api.personalChart()
      if (api.token() !== currentToken) return
      const personalChart = response.chart || null
      this.setData({ personalChart })
      if (personalChart) await this.refreshRealDay()
    } catch (err) {
      if (api.token() === currentToken) this.setData({ realError: err.message, personalChart: null,
        realDay: null, personalHints: [], personalHint: null, personalHintRemaining: 0 })
    } finally { this.setData({ realLoading: false }) }
  },

  async refreshRealDay() {
    if (!this.data.personalChart || !api.token()) return
    const date = this.data.selectedKey
    const currentToken = api.token()
    const requestId = (this._dailyRequestId || 0) + 1
    this._dailyRequestId = requestId
    this.setData({ realLoading: true, realError: '', realDay: null, personalHints: [],
      personalHint: null, personalHintRemaining: 0 })
    try {
      const response = await api.personalDay(date)
      if (api.token() !== currentToken || this.data.selectedKey !== date || this._dailyRequestId !== requestId) return
      const day = response.data
      const hints = personalHints(day)
      this.setData({ realDay: day, personalHints: hints, personalHint: hints[0] || null,
        personalHintRemaining: Math.max(0, hints.length - 1), selected: { ...this.data.selected,
        lunarFull: day.lunar_full || this.data.selected.lunarFull,
        yearGanZhi: day.year_gan_zhi || this.data.selected.yearGanZhi,
        monthGanZhi: day.month_gan_zhi || this.data.selected.monthGanZhi,
        dayGanZhi: day.day_gan_zhi || this.data.selected.dayGanZhi } })
    } catch (err) {
      if (api.token() !== currentToken || this.data.selectedKey !== date || this._dailyRequestId !== requestId) return
      if (err.code === 'personal_chart_unbound') this.setData({ personalChart: null })
      this.setData({ realError: err.message, realDay: null, personalHints: [],
        personalHint: null, personalHintRemaining: 0 })
    } finally { if (this._dailyRequestId === requestId) this.setData({ realLoading: false }) }
  },

  onReady() {
    navigation.pageReady(this, 'pages/calendar/index')
  },

  onHide() {
    navigation.pageHidden(this)
  },

  onUnload() {
    navigation.pageGone(this)
  },

  changeMonth(event) {
    const delta = Number(event.currentTarget.dataset.delta)
    const next = new Date(this.data.year, this.data.month - 1 + delta, 1)
    const year = next.getFullYear()
    const month = next.getMonth() + 1
    if (year < 1900 || year > 2100) {
      wx.showToast({ title: '请选择1900—2100年的日期', icon: 'none' })
      return
    }
    const selectedDay = Math.min(Number(this.data.selectedKey.slice(-2)), new Date(year, month, 0).getDate())
    const selectedKey = keyOf(year, month, selectedDay)
    this.showDate(selectedKey)
    this.refreshCalendarDay()
    this.refreshRealDay()
  },

  pickDate(event) {
    const selectedKey = event.currentTarget.dataset.key
    const [year, month] = selectedKey.split('-').map(Number)
    if (year < 1900 || year > 2100) {
      wx.showToast({ title: '请选择1900—2100年的日期', icon: 'none' })
      return
    }
    this.showDate(selectedKey)
    this.refreshCalendarDay()
    this.refreshRealDay()
  },

  pickToday() {
    const todayKey = chinaTodayKey()
    const [year, month] = todayKey.split('-').map(Number)
    if (year < 1900 || year > 2100) {
      wx.showToast({ title: '当前日期不在支持范围内', icon: 'none' })
      return
    }
    this.showDate(todayKey)
    this.refreshCalendarDay()
    this.refreshRealDay()
  },

  enablePersonalDay() {
    getApp().globalData.previewDayProfileEnabled = true
    this.setData({ dayPreviewEnabled: true })
  },

  openPersonalDay() {
    if (!this.data.dayPreviewEnabled && !this.data.personalChart) return
    navigation.navigate(this, { method: 'navigateTo', url: `/pages/personal-day/index?date=${this.data.selectedKey}`, key: 'personal' })
  },

  openBinding() {
    getApp().globalData.previewHistoryNext = true
    navigation.navigate(this, { method: 'switchTab', url: '/pages/home/index', key: 'history' })
  },

  openMovingDate() {
    navigation.navigate(this, { method: 'navigateTo', url: '/pages/moving-date/index', key: 'moving' })
  },

  switchInfoTab(event) {
    this.setData({ infoTab: event.currentTarget.dataset.tab })
  },
})
