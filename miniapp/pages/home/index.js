const { history } = require('../../data/demo')
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
    history: history.map(item => ({ ...item })),
    recentHistory: history[0] || null,
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
    const currentHistory = globalData.previewHistory || history
    this.setData({
      calendarPreview: calendarPreview(),
      history: currentHistory,
      recentHistory: currentHistory[0] || null,
    })
    if (globalData.previewHistoryNext) {
      globalData.previewHistoryNext = false
      const detail = currentHistory.find(item => item.id === globalData.previewHistoryDetailId) || null
      globalData.previewHistoryDetailId = null
      this.setData({ mode: 'history', detail })
    }
    wx.setNavigationBarTitle({ title: this.data.mode === 'history' ? '查询历史' : '干支演示' })
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
    wx.setNavigationBarTitle({ title: mode === 'history' ? '查询历史' : '干支演示' })
  },

  onNoteInput(event) {
    this.setData({ note: event.detail.value })
  },

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

  openDemoResult() {
    if (this.data.navigationKey) return
    getApp().globalData.previewQuery = {
      dateText: this.data.dateText,
      note: this.data.note,
    }
    navigation.navigate(this, { method: 'navigateTo', url: '/pages/result/index', key: 'result' })
  },

  openCalendar() {
    if (this.data.navigationKey) return
    getApp().globalData.previewCalendarNext = this.data.calendarPreview.key
    navigation.navigate(this, { method: 'switchTab', url: '/pages/calendar/index', key: 'calendar' })
  },

  openHistory() {
    this.setData({ mode: 'history' })
    wx.setNavigationBarTitle({ title: '查询历史' })
  },

  openHistoryDetail(event) {
    const detail = this.data.history.find(item => item.id === event.currentTarget.dataset.id)
    if (detail) this.setData({ detail })
  },

  closeHistoryDetail() {
    this.setData({ detail: null })
  },

  removeHistory(event) {
    const id = event.currentTarget.dataset.id
    wx.showModal({
      title: '移除演示记录',
      content: '仅从当前演示列表移除，重新打开小程序会恢复；不会删除真实数据。',
      confirmColor: '#c38c24',
      success: ({ confirm }) => {
        if (!confirm) return
        const nextHistory = this.data.history.filter(item => item.id !== id)
        getApp().globalData.previewHistory = nextHistory
        this.setData({ history: nextHistory, recentHistory: nextHistory[0] || null })
        wx.showToast({ title: '演示记录已移除', icon: 'none' })
      },
    })
  },
})
