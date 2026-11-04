const api = require('../../utils/api')
const { ACTIVITIES, defaultDateRange, validateDateRange, analyzeActivityRange } = require('../../data/real_date_selection')
const navigation = require('../../utils/navigation_feedback')
const { showAlmanacTerm } = require('../../data/almanac_terms')
const INITIAL_VISIBLE_DAYS = 5

function withVisibleDays(sections) {
  return sections.map(section => ({
    ...section, visibleDays: section.days.slice(0, INITIAL_VISIBLE_DAYS),
    hasMore: section.days.length > INITIAL_VISIBLE_DAYS,
  }))
}

Page({
  showAlmanacTerm,
  data: {
    activityId: 'moving', activity: ACTIVITIES[0],
    activityOptions: ACTIVITIES.map(item => ({ id: item.id, label: item.label, selected: item.id === 'moving' })),
    startDate: '', endDate: '', minDate: '1900-01-01', maxDate: '2100-12-31',
    rangeError: '', resultError: '', resultUpdating: false, sections: [], total: 0, candidateCount: 0,
    customRangeVisible: false, activeSectionId: 'yi', expandedDate: '', personalChart: null,
  },

  onLoad() {
    navigation.pageVisible(this, 'pages/moving-date/index')
    const range = defaultDateRange()
    this.setData(range)
    return this.refreshResults()
  },

  onShow() {
    navigation.pageVisible(this, 'pages/moving-date/index')
    return this.refreshBinding()
  },

  onReady() { navigation.pageReady(this, 'pages/moving-date/index') },
  onHide() { navigation.pageHidden(this) },
  onUnload() { this._requestId = (this._requestId || 0) + 1; this._bindingId = (this._bindingId || 0) + 1; navigation.pageGone(this) },

  async refreshBinding() {
    const id = this._bindingId = (this._bindingId || 0) + 1
    if (typeof wx.getStorageSync !== 'function' || !api.token()) {
      this.setData({ personalChart: null })
      return
    }
    const token = api.token()
    this.setData({ personalChart: null })
    try {
      const response = await api.personalChart()
      if (this._bindingId === id && api.token() === token) this.setData({ personalChart: response.chart || null })
    } catch (_) {
      if (this._bindingId === id) this.setData({ personalChart: null })
    }
  },

  changeStartDate(event) { return this.refreshResults({ startDate: event.detail.value }) },
  changeEndDate(event) { return this.refreshResults({ endDate: event.detail.value }) },
  toggleCustomRange() { this.setData({ customRangeVisible: !this.data.customRangeVisible }) },

  changeSection(event) {
    const id = event.currentTarget.dataset.section
    if (!this.data.sections.some(section => section.id === id) || id === this.data.activeSectionId) return
    this.setData({ activeSectionId: id, expandedDate: '' })
  },

  changeActivity(event) {
    const activity = ACTIVITIES.find(item => item.id === event.currentTarget.dataset.activity)
    if (!activity || activity.id === this.data.activityId) return
    return this.refreshResults({ activityId: activity.id, activity,
      activityOptions: ACTIVITIES.map(item => ({ id: item.id, label: item.label, selected: item.id === activity.id })) })
  },

  async refreshResults(overrides = {}) {
    const next = { ...this.data, ...overrides }
    const requestId = this._requestId = (this._requestId || 0) + 1
    const rangeError = validateDateRange(next.startDate, next.endDate)
    this.setData({ ...overrides, rangeError, resultError: '', resultUpdating: !rangeError,
      activeSectionId: 'yi',
      sections: [], total: 0, candidateCount: 0, expandedDate: '' })
    if (rangeError) return
    try {
      const response = await api.calendarRange(next.startDate, next.endDate)
      if (this._requestId !== requestId) return
      const result = analyzeActivityRange(next.activityId, next.startDate, next.endDate, response.data && response.data.days)
      this.setData({ sections: withVisibleDays(result.sections), total: result.total,
        candidateCount: result.candidateCount, resultUpdating: false })
    } catch (err) {
      if (this._requestId === requestId) this.setData({ resultError: err.message || '日期资料暂不可用', resultUpdating: false,
        sections: [], total: 0, candidateCount: 0 })
    }
  },

  retryResults() { return this.refreshResults() },

  showMore(event) {
    const id = event.currentTarget.dataset.section
    this.setData({ sections: this.data.sections.map(section => {
      if (section.id !== id) return section
      const count = Math.min(section.days.length, section.visibleDays.length + INITIAL_VISIBLE_DAYS)
      return { ...section, visibleDays: section.days.slice(0, count), hasMore: count < section.days.length }
    }) })
  },

  openDay(event) {
    const date = event.currentTarget.dataset.date
    const day = this.data.sections.flatMap(section => section.days).find(item => item.date === date)
    if (!day) return
    this.setData({ expandedDate: this.data.expandedDate === date ? '' : date })
  },

  openPersonalDay(event) {
    if (!this.data.personalChart) return
    const date = event.currentTarget.dataset.date
    if (date !== this.data.expandedDate) return
    navigation.navigate(this, { method: 'navigateTo', url: `/pages/personal-day/index?date=${date}`, key: date })
  },
})
