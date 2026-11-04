const fontSize = require('../../utils/font_size')
const calendarData = require('../../data/calendar')
const api = require('../../utils/api')
const { chart } = require('../../data/demo')
const { buildPersonalDay } = require('../../data/personal_day')
const { getJianChuInfo, jianChuNames } = require('../../data/jianchu')
const { ACTIVITIES, classifyActivityDate } = require('../../data/moving_date')
const navigation = require('../../utils/navigation_feedback')
const { personalThemes, personalDirectionRows } = require('../../utils/personal_day_view')
const { showAlmanacTerm } = require('../../data/almanac_terms')

const previewDate = '2026-10-22'

function dateDetails(date, activityId) {
  const record = calendarData[date] || {}
  const personalDay = buildPersonalDay(date, record, chart)
  const activity = ACTIVITIES.find(item => item.id === activityId)
  return {
    dateLabel: /^\d{4}-\d{2}-\d{2}$/.test(date) ? `${Number(date.slice(0, 4))}年${Number(date.slice(5, 7))}月${Number(date.slice(8, 10))}日` : date,
    lunarFull: record.lunarFull || '暂无农历标签',
    yearGanZhi: record.yearGanZhi || '—',
    monthGanZhi: record.monthGanZhi || '—',
    dayGanZhi: record.dayGanZhi || '—',
    jieQi: record.jieQi || '',
    jianChuInfo: getJianChuInfo(record.jianChu),
    personalDay,
    selection: activity ? {
      activityLabel: activity.label,
      note: activity.note,
      ...classifyActivityDate(activity.id, date, record, personalDay),
    } : null,
  }
}

Page({
  showAlmanacTerm,
  data: {
    enabled: false,
    chartDate: chart.input.dateTime,
    chartDay: `${chart.pillars[2].gan}${chart.pillars[2].zhi}`,
    favorable: chart.personalDay.fuyiYongshen,
    adverse: chart.personalDay.fuyiJishen,
    jianChuNames,
    realMode: true,
    calendarDay: null,
    calendarLoading: false,
    calendarError: '',
    realDay: null,
    personalThemes: [],
    realDirections: [],
    showActivityEvidence: false,
    realLoading: false,
    realError: '',
    ...dateDetails(previewDate),
  },

  onLoad(options) {
    navigation.pageVisible(this, 'pages/personal-day/index')
    const date = options && options.date ? options.date : previewDate
    this.setData({
      realMode: true,
      enabled: Boolean(getApp().globalData.previewDayProfileEnabled),
      ...dateDetails(date, options && options.activity),
    })
    this.loadCalendarDay(date)
    if (typeof wx.getStorageSync === 'function' && api.token()) this.loadRealDay(date)
  },

  async loadCalendarDay(date) {
    this.setData({ calendarDay: null, calendarLoading: true, calendarError: '' })
    try {
      const response = await api.calendarDay(date)
      const day = response.data
      if (!day || day.date !== date || !Array.isArray(day.yi) || !Array.isArray(day.ji)) throw new Error('日期资料不完整')
      this.setData({ calendarDay: day, lunarFull: day.lunar_full || this.data.lunarFull,
        yearGanZhi: day.year_gan_zhi || this.data.yearGanZhi,
        monthGanZhi: day.month_gan_zhi || this.data.monthGanZhi,
        dayGanZhi: day.day_gan_zhi || this.data.dayGanZhi })
    } catch (err) {
      this.setData({ calendarError: err.message })
    } finally { this.setData({ calendarLoading: false }) }
  },

  async loadRealDay(date) {
    const currentToken = api.token()
    this.setData({ realMode: true, realLoading: true, realDay: null, personalThemes: [], realDirections: [], showActivityEvidence: false, realError: '' })
    try {
      const response = await api.personalChart()
      if (api.token() !== currentToken) return
      if (!response.chart) {
        this.setData({ realLoading: false, realMode: true })
        return
      }
      const dayResponse = await api.personalDay(date)
      if (api.token() !== currentToken) return
      this.setData({ realDay: dayResponse.data, personalThemes: personalThemes(dayResponse.data), realDirections: personalDirectionRows(dayResponse.data) })
    } catch (err) {
      if (api.token() === currentToken) this.setData({ realError: err.message })
    } finally { if (api.token() === currentToken) this.setData({ realLoading: false }) }
  },

  openBinding() {
    navigation.navigate(this, { method: 'switchTab', url: '/pages/mine/index', key: 'mine' })
  },

  useDemo() { this.setData({ realMode: false }) },

  toggleActivityEvidence() { this.setData({ showActivityEvidence: !this.data.showActivityEvidence }) },

  onShow() {
    fontSize.apply(this)
    navigation.pageVisible(this, 'pages/personal-day/index')
  },

  onReady() {
    navigation.pageReady(this, 'pages/personal-day/index')
  },

  onHide() {
    navigation.pageHidden(this)
  },

  onUnload() {
    navigation.pageGone(this)
  },

  startPreview() {
    getApp().globalData.previewDayProfileEnabled = true
    this.setData({ enabled: true })
  },

  endPreview() {
    if (this.data.navigationKey) return
    getApp().globalData.previewDayProfileEnabled = false
    navigation.navigate(this, {
      method: 'navigateBack', delta: 1, target: '/pages/calendar/index', key: 'back',
      fallback: { method: 'switchTab', url: '/pages/calendar/index' },
    })
  },
})
