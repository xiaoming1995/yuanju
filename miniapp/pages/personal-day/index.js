const calendarData = require('../../data/calendar')
const { chart } = require('../../data/demo')
const { buildPersonalDay } = require('../../data/personal_day')
const { getJianChuInfo, jianChuNames } = require('../../data/jianchu')
const { ACTIVITIES, classifyActivityDate } = require('../../data/moving_date')
const navigation = require('../../utils/navigation_feedback')

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
  data: {
    enabled: false,
    chartDate: chart.input.dateTime,
    chartDay: `${chart.pillars[2].gan}${chart.pillars[2].zhi}`,
    favorable: chart.personalDay.fuyiYongshen,
    adverse: chart.personalDay.fuyiJishen,
    jianChuNames,
    ...dateDetails(previewDate),
  },

  onLoad(options) {
    navigation.pageVisible(this, 'pages/personal-day/index')
    const date = options && options.date ? options.date : previewDate
    this.setData({
      enabled: Boolean(getApp().globalData.previewDayProfileEnabled),
      ...dateDetails(date, options && options.activity),
    })
  },

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
