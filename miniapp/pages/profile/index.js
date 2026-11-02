const api = require('../../utils/api')
const navigation = require('../../utils/navigation_feedback')

Page({
  data: {
    loading: true, ready: false, loadError: '', loginRequired: false, profileOnboarding: false,
    profileNickname: '', profileAvatarPath: '', profileAvatarPreviewUrl: '',
    avatarStatus: '', nicknameStatus: '', nicknameMessage: '', nicknameFocus: false, nicknameInputVisible: true,
  },
  onLoad(options) {
    this.setData({ profileOnboarding: Boolean(options && options.onboarding === '1') })
    return this.loadProfile()
  },
  onShow() {
    this._profileHidden = false
    navigation.pageVisible(this, 'pages/profile/index')
    this.updateLeaveWarning()
  },
  onReady() { navigation.pageReady(this, 'pages/profile/index') },
  onHide() { this._profileHidden = true; navigation.pageHidden(this) },
  onUnload() {
    clearTimeout(this._reviewTimer)
    if (this._leaveWarning && typeof wx.disableAlertBeforeUnload === 'function') wx.disableAlertBeforeUnload({})
    navigation.pageGone(this)
  },
  async loadProfile() {
    const currentToken = api.token()
    this.setData({ loading: true, ready: false, loadError: '', loginRequired: false })
    try {
      if (!currentToken) throw new Error('请先登录后再修改资料')
      const response = await api.profile()
      if (this._navigationGone) return
      if (api.token() !== currentToken) throw new Error('登录状态已变更，请返回后重试')
      const user = response.data.user
      this._sessionToken = currentToken
      this._savedNickname = user.nickname || ''
      const avatarRequestId = this._avatarRequestId = (this._avatarRequestId || 0) + 1
      const avatarUrl = api.assetUrl(user.avatar_url) || ''
      this.setData({ ready: true, profileNickname: this._savedNickname,
        profileAvatarPreviewUrl: avatarUrl.startsWith('http://') ? '' : avatarUrl })
      api.avatarDisplayUrl(user.avatar_url).then(displayUrl => {
        if (!this._navigationGone && api.token() === currentToken &&
          this._avatarRequestId === avatarRequestId && !this.data.profileAvatarPath) {
          this.setData({ profileAvatarPreviewUrl: displayUrl })
        }
      }, () => {})
    } catch (err) {
      if (!this._navigationGone) this.setData({ loadError: err.message, loginRequired: !api.token() })
    } finally {
      if (!this._navigationGone) this.setData({ loading: false })
    }
  },
  canEdit() {
    if (!this.data.ready || this._navigationGone) return false
    if (api.token() !== this._sessionToken) {
      this.setData({ ready: false, loadError: '登录状态已变更，请返回后重新登录', loginRequired: true })
      this.updateLeaveWarning()
      return false
    }
    return true
  },
  onChooseAvatar(event) {
    if (!this.canEdit() || this.data.avatarStatus === 'saving') return
    const path = event.detail && event.detail.avatarUrl
    if (typeof path !== 'string' || !path) return
    this._avatarRequestId = (this._avatarRequestId || 0) + 1
    this.setData({ profileAvatarPath: path, profileAvatarPreviewUrl: path })
    return this.saveAvatar()
  },
  async saveAvatar() {
    if (!this.canEdit() || this.data.avatarStatus === 'saving' || !this.data.profileAvatarPath) return
    const path = this.data.profileAvatarPath
    this.setData({ avatarStatus: 'saving' })
    this.updateLeaveWarning()
    try {
      const response = await api.uploadAvatar(path)
      this.refreshMineIfVisible()
      if (this._navigationGone || api.token() !== this._sessionToken) return
      const avatarPath = response && response.data && response.data.avatar_url
      const avatarRequestId = this._avatarRequestId
      const savedAvatarUrl = api.assetUrl(avatarPath)
      this.setData({ profileAvatarPath: '', avatarStatus: 'saved',
        profileAvatarPreviewUrl: savedAvatarUrl && !savedAvatarUrl.startsWith('http://')
          ? savedAvatarUrl : this.data.profileAvatarPreviewUrl })
      if (avatarPath) api.avatarDisplayUrl(avatarPath).then(displayUrl => {
        if (!this._navigationGone && api.token() === this._sessionToken &&
          this._avatarRequestId === avatarRequestId && !this.data.profileAvatarPath) {
          this.setData({ profileAvatarPreviewUrl: displayUrl })
        }
      }, () => {})
    } catch (err) {
      this.showSaveError('avatar', err)
    } finally {
      if (!this._navigationGone) this.updateLeaveWarning()
    }
  },
  onProfileNicknameInput(event) {
    if (this._resettingNicknameInput || !this.canEdit() || ['reviewing', 'saving'].includes(this.data.nicknameStatus)) return this.data.profileNickname
    this._approvedNickname = ''
    this.setData({ profileNickname: event.detail.value, nicknameStatus: '', nicknameMessage: '' })
    this.updateLeaveWarning()
  },
  onNicknameFocus() { this.setData({ nicknameFocus: true }) },
  onNicknameBlur(event) {
    if (this._resettingNicknameInput) return
    this.setData({ nicknameFocus: false })
    if (!this.canEdit() || ['reviewing', 'saving'].includes(this.data.nicknameStatus)) return
    const nickname = (event.detail.value || '').trim()
    this.setData({ profileNickname: nickname })
    if (!nickname) {
      this.setData({ profileNickname: this._savedNickname, nicknameStatus: 'reviewError', nicknameMessage: '昵称不能为空' })
      this.updateLeaveWarning()
      return
    }
    if (nickname === this._savedNickname) {
      this.setData({ nicknameStatus: '', nicknameMessage: '' })
      this.updateLeaveWarning()
      return
    }
    this._reviewNickname = nickname
    this.setData({ nicknameStatus: 'reviewing', nicknameMessage: '' })
    this.updateLeaveWarning()
    clearTimeout(this._reviewTimer)
    this._reviewTimer = setTimeout(() => this.onProfileNicknameReview({ detail: { timeout: true } }), 10000)
  },
  onProfileNicknameReview(event) {
    if (!this.canEdit() || !this._reviewNickname) return
    clearTimeout(this._reviewTimer)
    const nickname = this._reviewNickname
    this._reviewNickname = ''
    const { pass, timeout, errMsg } = event.detail || {}
    const timedOut = timeout === true || /timeout|timed out/i.test(errMsg || '')
    if (pass !== true || timedOut) {
      this._approvedNickname = ''
      // Some native component versions report service timeouts as pass=false,
      // timeout=false and clear the input. Keep the draft without claiming rejection.
      this.setData({ profileNickname: nickname, nicknameStatus: 'reviewError',
        nicknameMessage: timedOut ? '微信昵称检查超时，请重试' : '微信暂未确认昵称，请重试或更换' })
      this.resetNicknameInput(false)
      this.updateLeaveWarning()
      return
    }
    this._approvedNickname = nickname
    return this.saveNickname()
  },
  async saveNickname() {
    if (!this.canEdit() || this.data.nicknameStatus === 'saving') return
    const nickname = this._approvedNickname
    if (!nickname || nickname !== this.data.profileNickname.trim() || nickname === this._savedNickname) return
    this.setData({ nicknameStatus: 'saving', nicknameMessage: '' })
    this.updateLeaveWarning()
    try {
      await api.updateNickname(nickname)
      this.refreshMineIfVisible()
      if (this._navigationGone || api.token() !== this._sessionToken) return
      this._savedNickname = nickname
      this.setData({ nicknameStatus: '' })
    } catch (err) {
      this.showSaveError('nickname', err)
    } finally {
      if (!this._navigationGone) this.updateLeaveWarning()
    }
  },
  retryNickname() {
    if (this.data.nicknameStatus === 'error') return this.saveNickname()
    this.setData({ nicknameStatus: '', nicknameMessage: '' })
    this.resetNicknameInput(true)
  },
  resetNicknameInput(focus) {
    // Recreate the native component so a cleared value is restored and the same
    // nickname can trigger a fresh review after a service failure.
    this._resettingNicknameInput = true
    this.setData({ nicknameInputVisible: false, nicknameFocus: false }, () => {
      if (!this._navigationGone) this.setData({ nicknameInputVisible: true, nicknameFocus: focus }, () => {
        this._resettingNicknameInput = false
      })
    })
  },
  showSaveError(field, err) {
    if (this._navigationGone) return
    this.setData({ [`${field}Status`]: 'error' })
    if (!api.token()) this.setData({ ready: false, loadError: '登录已失效，请返回后重新登录', loginRequired: true })
    wx.showToast({ title: err.message, icon: 'none', duration: 3000 })
  },
  updateLeaveWarning() {
    if (this._navigationGone || this._profileHidden) return
    const pending = Boolean(this.data.ready && (this.data.profileAvatarPath ||
      this.data.profileNickname.trim() !== (this._savedNickname || '')))
    if (pending === Boolean(this._leaveWarning)) return
    this._leaveWarning = pending
    if (pending && typeof wx.enableAlertBeforeUnload === 'function') {
      wx.enableAlertBeforeUnload({ message: '部分修改尚未保存，确定离开吗？' })
    } else if (!pending && typeof wx.disableAlertBeforeUnload === 'function') wx.disableAlertBeforeUnload({})
  },
  refreshMineIfVisible() {
    if (typeof getCurrentPages !== 'function') return
    const pages = getCurrentPages()
    const current = pages[pages.length - 1]
    if (current && current.route === 'pages/mine/index' && typeof current.loadProfile === 'function') current.loadProfile()
  },
  closeProfile() {
    navigation.navigate(this, { method: 'navigateBack', delta: 1, target: '/pages/mine/index', key: 'back',
      fallback: { method: 'switchTab', url: '/pages/mine/index' } })
  },
})
