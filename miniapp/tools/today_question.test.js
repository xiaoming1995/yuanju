const assert = require('node:assert/strict')
const test = require('node:test')
const fs = require('node:fs')
const config = require('../config')

function loadPage(wxApi) {
  let definition
  global.Page = value => { definition = value }
  global.wx = wxApi
  global.getApp = () => ({ globalData: {} })
  delete require.cache[require.resolve('../pages/today-question/index')]
  require('../pages/today-question/index')
  return { ...definition, data: { ...definition.data }, setData(patch) { this.data = { ...this.data, ...patch } } }
}

test('home inserts today question between date hero and almanac', () => {
  const home = fs.readFileSync(require.resolve('../pages/home/index.wxml'), 'utf8')
  const question = home.indexOf('today-question-card')
  assert.ok(question > home.indexOf('today-hero'))
  assert.ok(question < home.indexOf('today-almanac-card'))
  assert.match(home, /bindtap="openTodayQuestion"/)
})

test('today question shows only direct question input', () => {
  const page = fs.readFileSync(require.resolve('../pages/today-question/index.wxml'), 'utf8')
  assert.match(page, /今天想问什么？/)
  assert.match(page, /class="question-input"/)
  assert.doesNotMatch(page, /question-categories|selectCategory/)
})

test('today question uses current binding and server date', async () => {
  config.apiBaseUrl = 'https://api.test.example'
  const date = new Date(Date.now() + 8 * 60 * 60 * 1000).toISOString().slice(0, 10)
  const calls = []
  const page = loadPage({
    getStorageSync: () => 'jwt', showToast() {},
    request(options) {
      calls.push(options)
      const data = options.url.endsWith('/personal-chart') ? { chart: { id: 'mine' } }
        : { data: { date, chart_id: 'mine', level: '信号交错', summary: '有支持也有需要留意的信号', support: [], caution: [], advice: [], unknown: [], boundary: '文化参考' } }
      options.success({ statusCode: 200, data })
    },
  })
  page.onShow()
  await new Promise(resolve => setImmediate(resolve))
  assert.equal(page.data.state, 'ready')
  page.onQuestionInput({ detail: { value: '今天下午二面，担心临场表达' } })
  await page.submit()
  assert.equal(page.data.answer.chart_id, 'mine')
  assert.equal(page.data.answer.date, date)
  const sent = calls.find(call => call.url.endsWith('/today-question'))
  assert.deepEqual(sent.data, { question: '今天下午二面，担心临场表达' })
  assert.equal(sent.data.date, undefined)
  assert.equal(sent.data.chart_id, undefined)
  page.onHide()
})

test('today question discards answer after binding changes', async () => {
  config.apiBaseUrl = 'https://api.test.example'
  const date = new Date(Date.now() + 8 * 60 * 60 * 1000).toISOString().slice(0, 10)
  let chartId = 'old'
  const page = loadPage({
    getStorageSync: () => 'jwt', showToast() {},
    request(options) {
      if (options.url.endsWith('/today-question')) {
        chartId = 'new'
        options.success({ statusCode: 200, data: { data: { date, chart_id: 'old', level: '偏有利' } } })
      } else options.success({ statusCode: 200, data: { chart: { id: chartId } } })
    },
  })
  page.onShow()
  await new Promise(resolve => setImmediate(resolve))
  page.onQuestionInput({ detail: { value: '今天下午参加重要面试' } })
  await page.submit()
  assert.equal(page.data.answer, null)
  assert.equal(page.data.chartId, 'new')
  assert.match(page.data.error, /已变化/)
  page.onHide()
})

test('unbound today question never calls AI', async () => {
  config.apiBaseUrl = 'https://api.test.example'
  const calls = []
  const page = loadPage({
    getStorageSync: () => 'jwt',
    request(options) { calls.push(options.url); options.success({ statusCode: 200, data: { chart: null } }) },
  })
  page.onShow()
  await new Promise(resolve => setImmediate(resolve))
  assert.equal(page.data.state, 'unbound')
  await page.submit()
  assert.ok(calls.every(url => url.endsWith('/personal-chart')))
  page.onHide()
})

test('long question is rejected and an old-day response is discarded', async () => {
  config.apiBaseUrl = 'https://api.test.example'
  const originalNow = Date.now
  Date.now = () => Date.parse('2026-11-05T14:00:00Z')
  let pending
  let postCount = 0
  const page = loadPage({
    getStorageSync: () => 'jwt', showToast() {},
    request(options) {
      if (options.url.endsWith('/today-question')) { postCount++; pending = options }
      else options.success({ statusCode: 200, data: { chart: { id: 'mine' } } })
    },
  })
  try {
    page.onShow()
    await new Promise(resolve => setImmediate(resolve))
    page.onQuestionInput({ detail: { value: '长'.repeat(201) } })
    await page.submit()
    assert.equal(postCount, 0)
    page.onQuestionInput({ detail: { value: '今天有一场重要面试' } })
    const submission = page.submit()
    assert.equal(postCount, 1)
    Date.now = () => Date.parse('2026-11-05T16:01:00Z')
    page.onShow()
    pending.success({ statusCode: 200, data: { data: { date: '2026-11-05', chart_id: 'mine', level: '偏有利' } } })
    await submission
    assert.equal(page.data.answer, null)
    assert.equal(page.data.date, '2026-11-06')
  } finally { page.onHide(); Date.now = originalNow }
})

test('AI failure keeps the question and offers a retry state', async () => {
  config.apiBaseUrl = 'https://api.test.example'
  const page = loadPage({
    getStorageSync: () => 'jwt', showToast() {},
    request(options) {
      if (options.url.endsWith('/today-question')) options.success({ statusCode: 502, data: { error: '今日问事暂时无法生成，请稍后重试' } })
      else options.success({ statusCode: 200, data: { chart: { id: 'mine' } } })
    },
  })
  page.onShow()
  await new Promise(resolve => setImmediate(resolve))
  page.onQuestionInput({ detail: { value: '今天下午二面，担心临场表达' } })
  await page.submit()
  assert.equal(page.data.state, 'ready')
  assert.equal(page.data.answer, null)
  assert.equal(page.data.question, '今天下午二面，担心临场表达')
  assert.match(page.data.error, /稍后重试/)
  assert.equal(page.data.submitting, false)
  page.onHide()
})

test('daily quota response keeps the question and shows reset time', async () => {
  config.apiBaseUrl = 'https://api.test.example'
  const page = loadPage({
    getStorageSync: () => 'jwt', showToast() {},
    request(options) {
      if (options.url.endsWith('/today-question')) options.success({ statusCode: 429, data: { code: 'today_question_daily_limit', error: '今天免费问事次数与额外次数均已用完（每日免费 1 次），北京时间明天 0 点重置免费次数' } })
      else options.success({ statusCode: 200, data: { chart: { id: 'mine' } } })
    },
  })
  page.onShow()
  await new Promise(resolve => setImmediate(resolve))
  page.onQuestionInput({ detail: { value: '今天想约心仪的人见面' } })
  await page.submit()
  assert.equal(page.data.question, '今天想约心仪的人见面')
  assert.equal(page.data.answer, null)
  assert.match(page.data.error, /明天 0 点重置/)
  page.onHide()
})
