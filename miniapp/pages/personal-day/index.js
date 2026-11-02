const calendarData = require('../../data/calendar')
const api = require('../../utils/api')
const { chart } = require('../../data/demo')
const { buildPersonalDay } = require('../../data/personal_day')
const { getJianChuInfo, jianChuNames } = require('../../data/jianchu')
const { ACTIVITIES, classifyActivityDate } = require('../../data/moving_date')
const navigation = require('../../utils/navigation_feedback')
const { personalHints } = require('../../utils/personal_day_view')

const previewDate = '2026-10-22'

function realDirectionRows(day) {
  const labels = { favorable: '喜用', adverse: '忌神', neutral: '未落入喜忌' }
  if (!day || !labels[day.stem_direction] || !labels[day.branch_direction]) return []
  return [
    { name: `日干${day.day_gan}（${day.stem_element}）`, direction: labels[day.stem_direction] },
    { name: `日支${day.day_zhi}（${day.branch_element}主气）`, direction: labels[day.branch_direction] },
  ]
}

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
  data: {
    enabled: false,
    chartDate: chart.input.dateTime,
    chartDay: `${chart.pillars[2].gan}${chart.pillars[2].zhi}`,
    favorable: chart.personalDay.fuyiYongshen,
    adverse: chart.personalDay.fuyiJishen,
    jianChuNames,
    realMode: false,
    realDay: null,
    personalHints: [],
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
      enabled: Boolean(getApp().globalData.previewDayProfileEnabled),
      ...dateDetails(date, options && options.activity),
    })
    if (typeof wx.getStorageSync === 'function' && api.token()) this.loadRealDay(date)
  },

  async loadRealDay(date) {
    const currentToken = api.token()
    this.setData({ realMode: true, realLoading: true, realDay: null, personalHints: [], realDirections: [], showActivityEvidence: false, realError: '' })
    try {
      const response = await api.personalChart()
      if (api.token() !== currentToken) return
      if (!response.chart) {
        this.setData({ realLoading: false, realMode: !getApp().globalData.previewDayProfileEnabled })
        return
      }
      const dayResponse = await api.personalDay(date)
      if (api.token() !== currentToken) return
      this.setData({ realDay: dayResponse.data, personalHints: personalHints(dayResponse.data), realDirections: realDirectionRows(dayResponse.data) })
    } catch (err) {
      if (api.token() === currentToken) this.setData({ realError: err.message })
    } finally { if (api.token() === currentToken) this.setData({ realLoading: false }) }
  },

  openBinding() {
    getApp().globalData.previewHistoryNext = true
    navigation.navigate(this, { method: 'switchTab', url: '/pages/home/index', key: 'history' })
  },

  useDemo() { this.setData({ realMode: false }) },

  toggleActivityEvidence() { this.setData({ showActivityEvidence: !this.data.showActivityEvidence }) },

  onShow() {
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
