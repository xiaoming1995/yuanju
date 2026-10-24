const assert = require('node:assert/strict')
const test = require('node:test')

let createNavigationFeedback
try {
  ({ createNavigationFeedback } = require('../utils/navigation_feedback'))
} catch {
  createNavigationFeedback = undefined
}

function fakeClock() {
  let now = 0
  let nextId = 1
  const jobs = new Map()
  return {
    setTimeout(fn, delay) {
      const id = nextId++
      jobs.set(id, { at: now + delay, fn })
      return id
    },
    clearTimeout(id) { jobs.delete(id) },
    advance(ms) {
      const until = now + ms
      while (true) {
        const due = [...jobs.entries()].filter(([, job]) => job.at <= until).sort((a, b) => a[1].at - b[1].at)[0]
        if (!due) break
        now = due[1].at
        jobs.delete(due[0])
        due[1].fn()
      }
      now = until
    },
  }
}

function page() {
  return {
    data: {},
    setData(patch) { Object.assign(this.data, patch) },
  }
}

function setup() {
  assert.equal(typeof createNavigationFeedback, 'function')
  const clock = fakeClock()
  const calls = []
  const toasts = []
  const wxApi = {
    navigateTo(options) { calls.push({ method: 'navigateTo', options }) },
    switchTab(options) { calls.push({ method: 'switchTab', options }) },
    navigateBack(options) { calls.push({ method: 'navigateBack', options }) },
    showToast(options) { toasts.push(options) },
  }
  const navigation = createNavigationFeedback({
    getWx: () => wxApi,
    setTimeout: clock.setTimeout,
    clearTimeout: clock.clearTimeout,
  })
  return { navigation, clock, calls, toasts }
}

test('quick navigation gives immediate busy feedback but no delayed loading flash', () => {
  const { navigation, clock, calls } = setup()
  const source = page()
  const target = page()

  assert.equal(navigation.navigate(source, { method: 'navigateTo', url: '/pages/result/index', key: 'result' }), true)
  assert.equal(source.data.navigationKey, 'result')
  assert.equal(navigation.navigate(source, { method: 'navigateTo', url: '/pages/result/index', key: 'result' }), false)
  assert.equal(calls.length, 1)

  navigation.pageVisible(target, 'pages/result/index')
  navigation.pageReady(target, 'pages/result/index')
  clock.advance(300)
  assert.equal(source.data.navigationKey, '')
  assert.notEqual(source.data.navigationVisible, true)
  assert.notEqual(target.data.navigationVisible, true)
})

test('slow navigation moves the loading label to its target and clears it when ready', () => {
  const { navigation, clock } = setup()
  const source = page()
  const target = page()
  navigation.navigate(source, { method: 'switchTab', url: '/pages/calendar/index', key: 'calendar' })
  navigation.pageHidden(source)
  navigation.pageVisible(target, 'pages/calendar/index')

  clock.advance(199)
  assert.notEqual(target.data.navigationVisible, true)
  clock.advance(1)
  assert.equal(target.data.navigationVisible, true)
  navigation.pageReady(target, 'pages/calendar/index')
  assert.equal(target.data.navigationVisible, false)
  assert.equal(source.data.navigationKey, '')
})

test('failed navigation and timeout clear busy state and allow retry', () => {
  const { navigation, clock, calls, toasts } = setup()
  const source = page()
  navigation.navigate(source, { method: 'navigateTo', url: '/pages/result/index', key: 'result' })
  calls[0].options.fail({ errMsg: 'navigateTo:fail' })
  assert.equal(source.data.navigationKey, '')
  assert.equal(toasts.length, 1)

  assert.equal(navigation.navigate(source, { method: 'navigateTo', url: '/pages/result/index', key: 'result' }), true)
  clock.advance(8000)
  assert.equal(source.data.navigationKey, '')
  assert.equal(source.data.navigationVisible, false)
  assert.equal(toasts.length, 2)
})

test('returning to the source before destination readiness clears stale navigation', () => {
  const { navigation, clock } = setup()
  const source = page()
  navigation.navigate(source, { method: 'navigateTo', url: '/pages/result/index', key: 'result' })
  navigation.pageHidden(source)
  navigation.pageVisible(source, 'pages/home/index')
  clock.advance(300)
  assert.equal(source.data.navigationKey, '')
  assert.notEqual(source.data.navigationVisible, true)
})

test('back navigation uses its fallback tab when native back fails', () => {
  const { navigation, calls } = setup()
  const source = page()
  const target = page()
  navigation.navigate(source, {
    method: 'navigateBack', delta: 1, target: '/pages/home/index', key: 'back',
    fallback: { method: 'switchTab', url: '/pages/home/index' },
  })
  calls[0].options.fail({ errMsg: 'navigateBack:fail' })
  assert.equal(calls[1].method, 'switchTab')
  navigation.pageVisible(target, 'pages/home/index')
  navigation.pageReady(target, 'pages/home/index')
  assert.equal(source.data.navigationKey, '')
})
