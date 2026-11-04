const fontSize = require('../../utils/font_size')
const { toResult } = require('../../utils/chart_view')
const { toHistory } = require('../../utils/chart_view')
const api = require('../../utils/api')
const navigation = require('../../utils/navigation_feedback')

Page({
  data: {
    selectedDate: '1990-01-01 01:00',
    selectedNote: '',
    chart: null,
    pillars: [],
    basicRows: [],
    loading: true,
    loadError: '',
  },

  async onLoad() {
    navigation.pageVisible(this, 'pages/result/index')
    return this.loadDetail()
  },

  async loadDetail() {
    const loadSeq = this._loadSeq = (this._loadSeq || 0) + 1
    const currentToken = api.token()
    if (!currentToken) { this.showUnavailable('请先登录并配置本人命盘'); return }
    this.setData({ loading: true, loadError: '', chart: null, pillars: [], basicRows: [] })
    try {
      const response = await api.personalChartDetail()
      if (api.token() !== currentToken || this._navigationGone || this._hidden || loadSeq !== this._loadSeq) return
      const history = toHistory(response.chart)
      const chart = toResult(response.result, { dateText: history.date, gender: response.chart.gender, note: history.name })
      this.setData({
        selectedDate: history.date,
        selectedNote: history.name,
        chart,
        pillars: chart.pillars.map(p => ({ label: p.label.slice(0, 1), gan: p.gan, zhi: p.zhi, ganTone: p.ganTone, zhiTone: p.zhiTone })),
        basicRows: chart.basicRows,
        loading: false,
      })
    } catch (err) {
      if (api.token() !== currentToken || this._navigationGone || this._hidden || loadSeq !== this._loadSeq) return
      this.showUnavailable(err.code === 'personal_chart_unbound' ? '请先配置本人命盘' : err.message)
    }
  },

  showUnavailable(message) {
    this.setData({ loading: false, loadError: message, chart: null, pillars: [], basicRows: [] })
  },

  onShow() {
    fontSize.apply(this)
    this._hidden = false
    navigation.pageVisible(this, 'pages/result/index')
    if (this._shownOnce) this.loadDetail()
    this._shownOnce = true
  },

  onReady() {
    navigation.pageReady(this, 'pages/result/index')
  },

  onHide() {
    this._hidden = true
    this._loadSeq = (this._loadSeq || 0) + 1
    this.setData({ chart: null, pillars: [], basicRows: [] })
    navigation.pageHidden(this)
  },

  onUnload() {
    navigation.pageGone(this)
  },

  returnToMine() {
    navigation.navigate(this, {
      method: 'switchTab', url: '/pages/mine/index', key: 'back',
    })
  },
})
