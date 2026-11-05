const api = require('../../utils/api')
const fontSize = require('../../utils/font_size')
const navigation = require('../../utils/navigation_feedback')

function todayKey() { return new Date(Date.now() + 8 * 60 * 60 * 1000).toISOString().slice(0, 10) }
function dateLabel(key) { return `${Number(key.slice(5, 7))}月${Number(key.slice(8, 10))}日` }

Page({
  data: {
    question: '', date: todayKey(), dateLabel: dateLabel(todayKey()),
    chartId: '', state: 'loading', error: '', submitting: false, answer: null,
  },
  onLoad() { navigation.pageVisible(this, 'pages/today-question/index') },
  onReady() { navigation.pageReady(this, 'pages/today-question/index') },
  onShow() {
    fontSize.apply(this)
    navigation.pageVisible(this, 'pages/today-question/index')
    const date = todayKey()
    if (this.data.date !== date) this.setData({ date, dateLabel: dateLabel(date), answer: null, question: '' })
    this.loadBinding()
    clearTimeout(this._timer)
    const untilMidnight = 86400000 - ((Date.now() + 8 * 60 * 60 * 1000) % 86400000)
    this._timer = setTimeout(() => { const next = todayKey(); this.setData({ date: next, dateLabel: dateLabel(next), answer: null, question: '' }); this.loadBinding() }, untilMidnight + 1000)
    if (this._timer && typeof this._timer.unref === 'function') this._timer.unref()
  },
  onHide() { clearTimeout(this._timer); this._requestId = (this._requestId || 0) + 1; this.setData({ submitting: false }); navigation.pageHidden(this) },
  onUnload() { clearTimeout(this._timer); this._requestId = (this._requestId || 0) + 1; navigation.pageGone(this) },
  async loadBinding() {
    const requestId = (this._requestId || 0) + 1
    this._requestId = requestId
    const token = api.token()
    if (!token) { this.setData({ state: 'login', chartId: '', answer: null, submitting: false }); return }
    this.setData({ state: 'loading', error: '', submitting: false })
    try {
      const response = await api.personalChart()
      if (this._requestId !== requestId || api.token() !== token) return
      const chartId = response.chart && response.chart.id || ''
      if (chartId !== this.data.chartId) this.setData({ answer: null, question: '' })
      this.setData({ chartId, state: chartId ? 'ready' : 'unbound' })
    } catch (err) {
      if (this._requestId === requestId) this.setData({ state: api.token() ? 'error' : 'login', error: err.message, answer: null })
    }
  },
  onQuestionInput(event) { this.setData({ question: event.detail.value, answer: null }) },
  openSetup() { navigation.navigate(this, { method: 'switchTab', url: '/pages/mine/index', key: 'mine' }) },
  async submit() {
    if (this.data.submitting || this.data.state !== 'ready') return
    const question = (this.data.question || '').trim()
    if (Array.from(question).length < 5 || Array.from(question).length > 200) {
      wx.showToast({ title: '请填写 5—200 字的具体问题', icon: 'none' }); return
    }
    const token = api.token(), chartId = this.data.chartId, date = todayKey(), requestId = (this._requestId || 0) + 1
    this._requestId = requestId
    this.setData({ date, dateLabel: dateLabel(date), answer: null, error: '', submitting: true })
    try {
      const response = await api.todayQuestion({ question })
      if (this._requestId !== requestId || api.token() !== token || todayKey() !== date) return
      const binding = await api.personalChart()
      if (this._requestId !== requestId || api.token() !== token || todayKey() !== date) return
      if (!binding.chart || binding.chart.id !== chartId || !response.data || response.data.chart_id !== chartId || response.data.date !== date) {
        this.setData({ state: binding.chart ? 'ready' : 'unbound', chartId: binding.chart && binding.chart.id || '', answer: null, error: '日期或本人命盘已变化，请重新提问' })
        return
      }
      this.setData({ answer: response.data })
    } catch (err) {
      if (this._requestId !== requestId) return
      if (!api.token()) this.setData({ state: 'login', answer: null })
      else if (err.code === 'personal_chart_unbound') this.setData({ state: 'unbound', answer: null })
      this.setData({ error: err.message })
    } finally { if (this._requestId === requestId) this.setData({ submitting: false }) }
  },
})
