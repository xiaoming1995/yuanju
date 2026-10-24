const DELAY_MS = 200
const TIMEOUT_MS = 8000

function routeOf(url) {
  return String(url || '').split('?')[0].replace(/^\/+/, '')
}

function createNavigationFeedback({
  getWx = () => wx,
  setTimeout: schedule = setTimeout,
  clearTimeout: cancel = clearTimeout,
} = {}) {
  let pending = null

  function setPageData(page, patch) {
    if (page && !page._navigationGone && typeof page.setData === 'function') page.setData(patch)
  }

  function finish(error) {
    if (!pending) return
    const task = pending
    pending = null
    cancel(task.delayTimer)
    cancel(task.timeoutTimer)
    setPageData(task.source, { navigationKey: '', navigationVisible: false })
    if (task.targetPage !== task.source) setPageData(task.targetPage, { navigationVisible: false })
    if (error && typeof task.wxApi.showToast === 'function') {
      task.wxApi.showToast({ title: '打开失败，请重试', icon: 'none' })
    }
  }

  function navigate(source, { method, url, delta, target, key, fallback } = {}) {
    if (pending) return false
    const task = {
      source,
      target: routeOf(target || url),
      key: key || 'page',
      visible: false,
      left: false,
      targetPage: null,
      fallbackUsed: false,
      wxApi: getWx(),
    }
    pending = task
    source._navigationGone = false
    setPageData(source, { navigationKey: task.key, navigationVisible: false })
    task.delayTimer = schedule(() => {
      if (pending !== task) return
      task.visible = true
      setPageData(task.targetPage || task.source, { navigationVisible: true })
    }, DELAY_MS)
    task.timeoutTimer = schedule(() => {
      if (pending === task) finish(true)
    }, TIMEOUT_MS)

    function invoke(nextMethod, nextOptions) {
      const args = nextMethod === 'navigateBack'
        ? { delta: nextOptions.delta || 1 }
        : { url: nextOptions.url }
      args.fail = () => {
        if (pending !== task) return
        if (fallback && !task.fallbackUsed) {
          task.fallbackUsed = true
          invoke(fallback.method, fallback)
          return
        }
        finish(true)
      }
      try {
        task.wxApi[nextMethod](args)
      } catch {
        args.fail()
      }
    }

    invoke(method, { url, delta })
    return true
  }

  function pageVisible(page, route) {
    page._navigationGone = false
    if (!pending) return
    if (pending.source === page && pending.left) {
      finish(false)
      return
    }
    if (pending.target !== routeOf(route)) return
    pending.targetPage = page
    if (page._navigationReady) {
      finish(false)
      return
    }
    if (pending.visible) {
      setPageData(pending.source, { navigationVisible: false })
      setPageData(page, { navigationVisible: true })
    }
  }

  function pageReady(page, route) {
    page._navigationReady = true
    if (pending && pending.target === routeOf(route)) finish(false)
  }

  function pageHidden(page) {
    if (pending && pending.source === page) pending.left = true
  }

  function pageGone(page) {
    page._navigationGone = true
    if (pending && pending.targetPage === page) finish(false)
  }

  return { navigate, pageVisible, pageReady, pageHidden, pageGone }
}

module.exports = createNavigationFeedback()
module.exports.createNavigationFeedback = createNavigationFeedback
