const assert = require('node:assert/strict')
const test = require('node:test')

let createRenderFeedback
try {
  ({ createRenderFeedback } = require('../utils/render_feedback'))
} catch {
  createRenderFeedback = undefined
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
      now += ms
      for (const [id, job] of [...jobs]) {
        if (job.at <= now) { jobs.delete(id); job.fn() }
      }
    },
  }
}

function delayedPage() {
  const callbacks = []
  return {
    data: {}, callbacks,
    setData(patch, callback) {
      Object.assign(this.data, patch)
      if (callback) callbacks.push(callback)
    },
  }
}

function setup() {
  assert.equal(typeof createRenderFeedback, 'function')
  const clock = fakeClock()
  return { feedback: createRenderFeedback({ setTimeout: clock.setTimeout, clearTimeout: clock.clearTimeout }), clock }
}

test('fast result render does not flash an updating indicator', () => {
  const { feedback, clock } = setup()
  const page = delayedPage()
  feedback.apply(page, { total: 30 })
  page.callbacks.shift()()
  clock.advance(250)
  assert.equal(page.data.total, 30)
  assert.notEqual(page.data.resultUpdating, true)
})

test('slow result render shows a local indicator until the view update finishes', () => {
  const { feedback, clock } = setup()
  const page = delayedPage()
  feedback.apply(page, { total: 90 })
  clock.advance(199)
  assert.notEqual(page.data.resultUpdating, true)
  clock.advance(1)
  assert.equal(page.data.resultUpdating, true)
  page.callbacks.shift()()
  assert.equal(page.data.resultUpdating, false)
})

test('an older render callback cannot clear a newer pending result', () => {
  const { feedback, clock } = setup()
  const page = delayedPage()
  feedback.apply(page, { total: 30 })
  feedback.apply(page, { total: 31 })
  page.callbacks.shift()()
  clock.advance(200)
  assert.equal(page.data.resultUpdating, true)
  page.callbacks.shift()()
  assert.equal(page.data.resultUpdating, false)
  assert.equal(page.data.total, 31)
})

test('a stalled view update stops loading and offers a retry state', () => {
  const { feedback, clock } = setup()
  const page = delayedPage()
  feedback.apply(page, { total: 90 })
  clock.advance(8000)
  assert.equal(page.data.resultUpdating, false)
  assert.equal(page.data.resultRenderError, true)
})

test('leaving the page cancels pending result feedback', () => {
  const { feedback, clock } = setup()
  const page = delayedPage()
  feedback.apply(page, { total: 90 })
  clock.advance(200)
  assert.equal(page.data.resultUpdating, true)
  feedback.cancelPage(page)
  clock.advance(8000)
  assert.equal(page.data.resultUpdating, false)
  assert.notEqual(page.data.resultRenderError, true)
})
