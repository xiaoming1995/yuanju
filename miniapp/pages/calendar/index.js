const calendarData = require('../../data/calendar')
const { chart } = require('../../data/demo')
const { buildPersonalDay } = require('../../data/personal_day')
const { getJianChuInfo } = require('../../data/jianchu')
const navigation = require('../../utils/navigation_feedback')

const weekNames = ['星期日', '星期一', '星期二', '星期三', '星期四', '星期五', '星期六']

function keyOf(year, month, day) {
  return `${year}-${`${month}`.padStart(2, '0')}-${`${day}`.padStart(2, '0')}`
}

function monthView(year, month, selectedKey) {
  const firstWeekday = new Date(year, month - 1, 1).getDay()
  const daysInMonth = new Date(year, month, 0).getDate()
  const cellCount = Math.ceil((firstWeekday + daysInMonth) / 7) * 7
  const today = new Date()
  const todayKey = keyOf(today.getFullYear(), today.getMonth() + 1, today.getDate())
  const cells = []
  for (let index = 0; index < cellCount; index++) {
    const date = new Date(year, month - 1, index - firstWeekday + 1)
    const key = keyOf(date.getFullYear(), date.getMonth() + 1, date.getDate())
    const label = calendarData[key] || {}
    cells.push({
      key,
      day: date.getDate(),
      lunar: label.lunar || '演示',
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
  const today = new Date()
  return {
    date: key,
    weekday: weekNames[date.getDay()],
    lunarFull: record.lunarFull || '演示日期',
    yearGanZhi: record.yearGanZhi || '—',
    monthGanZhi: record.monthGanZhi || '—',
    dayGanZhi: record.dayGanZhi || '—',
    jianChu: record.jianChu || '',
    jianChuInfo: getJianChuInfo(record.jianChu),
    jieQi: record.jieQi || '',
    isToday: year === today.getFullYear() && month === today.getMonth() + 1 && day === today.getDate(),
  }
}

const initialYear = 2026
const initialMonth = 10
const initialKey = '2026-10-22'
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
    infoTab: 'date',
    weekShort: ['日', '一', '二', '三', '四', '五', '六'],
  },

  onShow() {
    navigation.pageVisible(this, 'pages/calendar/index')
    const dayPreviewEnabled = Boolean(getApp().globalData.previewDayProfileEnabled)
    const selectedKey = getApp().globalData.previewCalendarNext
    if (!selectedKey) {
      this.setData({ dayPreviewEnabled })
      return
    }
    getApp().globalData.previewCalendarNext = null
    const [year, month] = selectedKey.split('-').map(Number)
    const view = monthView(year, month, selectedKey)
    this.setData({ year, month, selectedKey, selected: selectedInfo(selectedKey), cells: view.cells, terms: view.terms, termSummary: view.termSummary, dayPreviewEnabled, personalDay: buildPersonalDay(selectedKey, calendarData[selectedKey], chart) })
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
    if (year < 2025 || year > 2027) {
      wx.showToast({ title: '演示日历范围为2025—2027年', icon: 'none' })
      return
    }
    const selectedDay = Math.min(Number(this.data.selectedKey.slice(-2)), new Date(year, month, 0).getDate())
    const selectedKey = keyOf(year, month, selectedDay)
    const view = monthView(year, month, selectedKey)
    this.setData({ year, month, selectedKey, selected: selectedInfo(selectedKey), cells: view.cells, terms: view.terms, termSummary: view.termSummary, personalDay: buildPersonalDay(selectedKey, calendarData[selectedKey], chart) })
  },

  pickDate(event) {
    const selectedKey = event.currentTarget.dataset.key
    const [year, month] = selectedKey.split('-').map(Number)
    if (year < 2025 || year > 2027) {
      wx.showToast({ title: '演示日历范围为2025—2027年', icon: 'none' })
      return
    }
    const view = monthView(year, month, selectedKey)
    this.setData({ year, month, selectedKey, selected: selectedInfo(selectedKey), cells: view.cells, terms: view.terms, termSummary: view.termSummary, personalDay: buildPersonalDay(selectedKey, calendarData[selectedKey], chart) })
  },

  pickToday() {
    const now = new Date()
    const year = now.getFullYear()
    if (year < 2025 || year > 2027) {
      wx.showToast({ title: '当前日期不在演示范围内', icon: 'none' })
      return
    }
    const month = now.getMonth() + 1
    const selectedKey = keyOf(year, month, now.getDate())
    const view = monthView(year, month, selectedKey)
    this.setData({ year, month, selectedKey, selected: selectedInfo(selectedKey), cells: view.cells, terms: view.terms, termSummary: view.termSummary, personalDay: buildPersonalDay(selectedKey, calendarData[selectedKey], chart) })
  },

  enablePersonalDay() {
    getApp().globalData.previewDayProfileEnabled = true
    this.setData({ dayPreviewEnabled: true })
  },

  openPersonalDay() {
    if (!this.data.dayPreviewEnabled) return
    navigation.navigate(this, { method: 'navigateTo', url: `/pages/personal-day/index?date=${this.data.selectedKey}`, key: 'personal' })
  },

  openMovingDate() {
    navigation.navigate(this, { method: 'navigateTo', url: '/pages/moving-date/index', key: 'moving' })
  },

  switchInfoTab(event) {
    this.setData({ infoTab: event.currentTarget.dataset.tab })
  },
})
