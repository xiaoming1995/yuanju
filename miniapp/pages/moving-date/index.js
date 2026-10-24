const calendarData = require('../../data/calendar')
const { chart } = require('../../data/demo')
const { ACTIVITIES, defaultDateRange, analyzeActivityRange } = require('../../data/moving_date')
const navigation = require('../../utils/navigation_feedback')
const renderFeedback = require('../../utils/render_feedback')
const INITIAL_VISIBLE_DAYS = 7

function withVisibleDays(sections) {
  return sections.map(section => ({
    ...section,
    visibleDays: section.days.slice(0, INITIAL_VISIBLE_DAYS),
    hasMore: section.days.length > INITIAL_VISIBLE_DAYS,
  }))
}

Page({
  data: {
    enabled: false,
    activityId: 'moving',
    activity: ACTIVITIES[0],
    activityOptions: ACTIVITIES.map(item => ({ id: item.id, label: item.label, selected: item.id === 'moving' })),
    startDate: '',
    endDate: '',
    minDate: '2025-01-01',
    maxDate: '2027-12-31',
    rangeError: '',
    sections: [],
    total: 0,
    priorityCount: 0,
    resultUpdating: false,
    resultRenderError: false,
    chartDate: chart.input.dateTime,
  },

  onLoad() {
    navigation.pageVisible(this, 'pages/moving-date/index')
    const range = defaultDateRange()
    const enabled = Boolean(getApp().globalData.previewDayProfileEnabled)
    this.refreshResults({ enabled, startDate: range.startDate, endDate: range.endDate })
  },

  onShow() {
    navigation.pageVisible(this, 'pages/moving-date/index')
    const enabled = Boolean(getApp().globalData.previewDayProfileEnabled)
    if (enabled !== this.data.enabled) {
      this.refreshResults({ enabled })
    }
  },

  onReady() {
    navigation.pageReady(this, 'pages/moving-date/index')
  },

  onHide() {
    navigation.pageHidden(this)
    renderFeedback.cancelPage(this)
  },

  onUnload() {
    navigation.pageGone(this)
    renderFeedback.cancelPage(this)
  },

  startPreview() {
    getApp().globalData.previewDayProfileEnabled = true
    this.refreshResults({ enabled: true })
  },

  changeStartDate(event) {
    this.refreshResults({ startDate: event.detail.value })
  },

  changeEndDate(event) {
    this.refreshResults({ endDate: event.detail.value })
  },

  changeActivity(event) {
    const activity = ACTIVITIES.find(item => item.id === event.currentTarget.dataset.activity)
    if (!activity || activity.id === this.data.activityId) return
    this.refreshResults({
      activityId: activity.id,
      activity,
      activityOptions: ACTIVITIES.map(item => ({ id: item.id, label: item.label, selected: item.id === activity.id })),
    })
  },

  refreshResults(overrides = {}) {
    const next = { ...this.data, ...overrides }
    if (!next.enabled) {
      renderFeedback.apply(this, { ...overrides, rangeError: '', sections: [], total: 0, priorityCount: 0 })
      return
    }
    if (!next.startDate && !next.endDate) {
      renderFeedback.apply(this, {
        ...overrides,
        rangeError: '演示日历范围为2025—2027年，请选择范围内的日期。',
        sections: [], total: 0, priorityCount: 0,
      })
      return
    }
    const result = analyzeActivityRange(next.activityId, next.startDate, next.endDate, calendarData, chart)
    renderFeedback.apply(this, {
      ...overrides,
      rangeError: result.error,
      sections: result.error ? [] : withVisibleDays(result.sections),
      total: result.total,
      priorityCount: result.priorityCount,
    })
  },

  retryResults() {
    this.refreshResults()
  },

  showMore(event) {
    const id = event.currentTarget.dataset.section
    this.setData({
      sections: this.data.sections.map(section => {
        if (section.id !== id) return section
        const count = Math.min(section.days.length, section.visibleDays.length + INITIAL_VISIBLE_DAYS)
        return { ...section, visibleDays: section.days.slice(0, count), hasMore: count < section.days.length }
      }),
    })
  },

  openDay(event) {
    const date = event.currentTarget.dataset.date
    if (!/^\d{4}-\d{2}-\d{2}$/.test(date)) return
    navigation.navigate(this, { method: 'navigateTo', url: `/pages/personal-day/index?date=${date}&activity=${this.data.activityId}`, key: date })
  },
})
