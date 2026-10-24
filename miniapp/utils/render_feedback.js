const DELAY_MS = 200
const TIMEOUT_MS = 8000

function createRenderFeedback({ setTimeout: schedule = setTimeout, clearTimeout: cancel = clearTimeout } = {}) {
  const pending = new WeakMap()

  function stop(page) {
    const state = pending.get(page)
    if (!state) return
    cancel(state.delayTimer)
    cancel(state.timeoutTimer)
    pending.delete(page)
  }

  function apply(page, patch) {
    const wasUpdating = Boolean(page.data.resultUpdating)
    stop(page)
    const state = { shown: wasUpdating, delayTimer: null, timeoutTimer: null }
    pending.set(page, state)
    state.delayTimer = schedule(() => {
      if (pending.get(page) !== state) return
      state.shown = true
      page.setData({ resultUpdating: true })
    }, DELAY_MS)
    state.timeoutTimer = schedule(() => {
      if (pending.get(page) !== state) return
      stop(page)
      page.setData({ resultUpdating: false, resultRenderError: true })
    }, TIMEOUT_MS)
    page.setData({ ...patch, resultUpdating: wasUpdating, resultRenderError: false }, () => {
      if (pending.get(page) !== state) return
      stop(page)
      if (state.shown) page.setData({ resultUpdating: false })
    })
  }

  function cancelPage(page) {
    stop(page)
    if (page.data.resultUpdating) page.setData({ resultUpdating: false })
  }

  return { apply, cancelPage }
}

module.exports = createRenderFeedback()
module.exports.createRenderFeedback = createRenderFeedback
