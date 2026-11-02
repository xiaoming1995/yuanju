const assert = require('node:assert/strict')
const test = require('node:test')
const api = require('../utils/api')
const navigation = require('../utils/navigation_feedback')

async function withProfile(options, run) {
  const previous = { token: api.token, profile: api.profile, assetUrl: api.assetUrl,
    avatarDisplayUrl: api.avatarDisplayUrl,
    updateNickname: api.updateNickname, uploadAvatar: api.uploadAvatar, navigate: navigation.navigate }
  const user = { id: 'u1', nickname: '缘友582731', avatar_url: '', ...options.user }
  const saved = [], routes = [], toasts = [], leaveWarnings = []
  api.token = () => 'test-token'
  api.profile = async () => ({ data: { user } })
  api.assetUrl = value => value
  api.avatarDisplayUrl = async value => value || ''
  api.updateNickname = async value => { saved.push(['name', value]); user.nickname = value }
  api.uploadAvatar = async value => {
    saved.push(['avatar', value])
    user.avatar_url = '/static/uploads/user-avatars/saved.png'
    return { data: { avatar_url: user.avatar_url } }
  }
  navigation.navigate = (page, route) => routes.push(route)
  global.wx = { showToast: value => toasts.push(value),
    enableAlertBeforeUnload: () => leaveWarnings.push(true), disableAlertBeforeUnload: () => leaveWarnings.push(false) }
  let definition
  global.Page = value => { definition = value }
  delete require.cache[require.resolve('../pages/profile/index')]
  require('../pages/profile/index')
  const page = { ...definition, data: { ...definition.data }, setData(patch, callback) { Object.assign(this.data, patch); if (callback) callback() } }
  try {
    if (options.beforeLoad) options.beforeLoad()
    await page.onLoad(options.onboarding ? { onboarding: '1' } : {})
    await run({ page, user, saved, routes, toasts, leaveWarnings })
  } finally {
    page.onUnload()
    Object.assign(api, { token: previous.token, profile: previous.profile, assetUrl: previous.assetUrl,
      avatarDisplayUrl: previous.avatarDisplayUrl,
      updateNickname: previous.updateNickname, uploadAvatar: previous.uploadAvatar })
    navigation.navigate = previous.navigate
  }
}

test('profile loads the stored nickname and saves only the selected avatar', async () => {
  await withProfile({ user: { nickname: '微信用户', avatar_url: '/current.png' } }, async ({ page, saved, routes }) => {
    assert.equal(page.data.profileNickname, '微信用户')
    assert.equal(page.data.profileAvatarPreviewUrl, '/current.png')
    await page.onChooseAvatar({ detail: { avatarUrl: { invalid: true } } })
    assert.deepEqual(saved, [])
    await page.onChooseAvatar({ detail: { avatarUrl: 'wxfile://selected' } })
    assert.deepEqual(saved, [['avatar', 'wxfile://selected']])
    assert.equal(page.data.avatarStatus, 'saved')
    assert.equal(page.data.profileAvatarPath, '')
    assert.equal(page.data.profileAvatarPreviewUrl, '/static/uploads/user-avatars/saved.png')
    assert.deepEqual(routes, [])
  })
})

function enterNickname(page, value) {
  page.onProfileNicknameInput({ detail: { value } })
  page.onNicknameBlur({ detail: { value } })
}

test('nickname saves after review, skips unchanged input and stays on the page', async () => {
  await withProfile({}, async ({ page, saved, routes, leaveWarnings }) => {
    page.onProfileNicknameInput({ detail: { value: '新昵称' } })
    assert.deepEqual(saved, [])
    page.onNicknameBlur({ detail: { value: '新昵称' } })
    assert.equal(page.data.nicknameStatus, 'reviewing')
    assert.deepEqual(saved, [])
    await page.onProfileNicknameReview({ detail: { pass: true, timeout: false } })
    assert.deepEqual(saved, [['name', '新昵称']])
    assert.equal(page.data.nicknameStatus, '')
    assert.deepEqual(leaveWarnings, [true, false])
    enterNickname(page, '新昵称')
    await page.onProfileNicknameReview({ detail: { pass: true } })
    assert.equal(saved.length, 1)
    assert.deepEqual(routes, [])
  })
})

test('initial registration saves each field independently of the generated nickname', async () => {
  await withProfile({ onboarding: true }, async ({ page, saved, routes }) => {
    assert.equal(page.data.profileNickname, '缘友582731')
    await page.onChooseAvatar({ detail: { avatarUrl: 'wxfile://selected' } })
    assert.deepEqual(saved, [['avatar', 'wxfile://selected']])
    enterNickname(page, '自己的昵称')
    await page.onProfileNicknameReview({ detail: { pass: true } })
    assert.deepEqual(saved.at(-1), ['name', '自己的昵称'])
    assert.deepEqual(routes, [])
  })
})

test('failed avatar and nickname saves retry independently and preserve drafts', async () => {
  await withProfile({}, async ({ page, saved, routes }) => {
    api.uploadAvatar = async () => { throw new Error('上传失败') }
    await page.onChooseAvatar({ detail: { avatarUrl: 'wxfile://selected' } })
    assert.equal(page.data.avatarStatus, 'error')
    assert.equal(page.data.profileAvatarPath, 'wxfile://selected')
    enterNickname(page, '新的昵称')
    await page.onProfileNicknameReview({ detail: { pass: true } })
    assert.deepEqual(saved, [['name', '新的昵称']])
    api.uploadAvatar = async value => saved.push(['avatar', value])
    await page.saveAvatar()
    assert.deepEqual(saved.at(-1), ['avatar', 'wxfile://selected'])
    api.updateNickname = async () => { throw new Error('网络错误') }
    enterNickname(page, '另一个昵称')
    await page.onProfileNicknameReview({ detail: { pass: true } })
    assert.equal(page.data.nicknameStatus, 'error')
    assert.equal(page.data.profileNickname, '另一个昵称')
    api.updateNickname = async value => saved.push(['name', value])
    await page.retryNickname()
    assert.equal(page.data.nicknameStatus, '')
    assert.deepEqual(saved, [['name', '新的昵称'], ['avatar', 'wxfile://selected'], ['name', '另一个昵称']])
    assert.deepEqual(routes, [])
  })
})

test('a saving avatar does not block nickname saving and repeated avatar events are ignored', async () => {
  await withProfile({}, async ({ page, saved }) => {
    let finishAvatar
    api.uploadAvatar = () => new Promise(resolve => { finishAvatar = resolve })
    const uploading = page.onChooseAvatar({ detail: { avatarUrl: 'wxfile://selected' } })
    await page.onChooseAvatar({ detail: { avatarUrl: 'wxfile://duplicate' } })
    assert.equal(page.data.profileAvatarPath, 'wxfile://selected')
    enterNickname(page, '新昵称')
    await page.onProfileNicknameReview({ detail: { pass: true } })
    assert.deepEqual(saved, [['name', '新昵称']])
    assert.equal(page.data.avatarStatus, 'saving')
    finishAvatar()
    await uploading
    assert.equal(page.data.avatarStatus, 'saved')
  })
})

test('failed or timed out nickname review never submits and can be retried', async () => {
  await withProfile({}, async ({ page, saved }) => {
    enterNickname(page, '不通过的昵称')
    await page.onProfileNicknameReview({ detail: { pass: false, timeout: false } })
    assert.equal(page.data.profileNickname, '不通过的昵称')
    assert.deepEqual(saved, [])
    enterNickname(page, '检查超时的昵称')
    await page.onProfileNicknameReview({ detail: { timeout: true } })
    assert.equal(page.data.nicknameStatus, 'reviewError')
    assert.equal(page.data.profileNickname, '检查超时的昵称')
    page.retryNickname()
    assert.equal(page.data.nicknameFocus, true)
    page.onNicknameBlur({ detail: { value: '检查超时的昵称' } })
    await page.onProfileNicknameReview({ detail: { pass: true } })
    assert.deepEqual(saved, [['name', '检查超时的昵称']])
    enterNickname(page, '尚未审核的昵称')
    page.onUnload()
    await page.onProfileNicknameReview({ detail: { pass: true } })
    assert.equal(saved.length, 1)
  })
})

test('a native timeout reported as pass=false preserves the nickname and supports same-value retry', async () => {
  await withProfile({}, async ({ page, saved }) => {
    enterNickname(page, '保留我的输入')
    // 3.17.5 compares a bare error string with a prefixed message and can
    // report a native upload timeout using these misleading boolean flags.
    await page.onProfileNicknameReview({ detail: { pass: false, timeout: false } })
    assert.equal(page.data.profileNickname, '保留我的输入')
    assert.equal(page.data.nicknameInputVisible, true)
    assert.equal(page.data.nicknameStatus, 'reviewError')
    assert.match(page.data.nicknameMessage, /暂未确认/)
    assert.doesNotMatch(page.data.nicknameMessage, /未通过/)
    assert.deepEqual(saved, [])
    page.retryNickname()
    assert.equal(page.data.profileNickname, '保留我的输入')
    assert.equal(page.data.nicknameFocus, true)
    page.onNicknameBlur({ detail: { value: '保留我的输入' } })
    await page.onProfileNicknameReview({ detail: { pass: true, timeout: false } })
    assert.deepEqual(saved, [['name', '保留我的输入']])
  })
})

test('an incomplete review result never authorizes saving and a timeout error is identified when available', async () => {
  await withProfile({}, async ({ page, saved }) => {
    enterNickname(page, '保留昵称')
    await page.onProfileNicknameReview({ detail: {} })
    assert.deepEqual(saved, [])
    assert.equal(page.data.profileNickname, '保留昵称')
    page.retryNickname()
    page.onNicknameBlur({ detail: { value: '保留昵称' } })
    await page.onProfileNicknameReview({ detail: { pass: false, timeout: false, errMsg: 'reviewNickname:fail upload nickname timeout' } })
    assert.match(page.data.nicknameMessage, /超时/)
    assert.equal(page.data.profileNickname, '保留昵称')
    assert.deepEqual(saved, [])
  })
})

test('profile load can retry and missing or replaced sessions cannot submit changes', async () => {
  await withProfile({ beforeLoad() { api.profile = async () => { throw new Error('网络错误') } } }, async ({ page, saved }) => {
    assert.equal(page.data.loadError, '网络错误')
    await page.onChooseAvatar({ detail: { avatarUrl: 'wxfile://selected' } })
    assert.deepEqual(saved, [])
    api.profile = async () => ({ data: { user: { nickname: '小明' } } })
    await page.loadProfile()
    assert.equal(page.data.ready, true)
    assert.equal(page.data.loadError, '')
    api.token = () => 'another-account-token'
    await page.onChooseAvatar({ detail: { avatarUrl: 'wxfile://selected' } })
    assert.equal(page.data.loginRequired, true)
    assert.deepEqual(saved, [])
  })
  await withProfile({ beforeLoad() { api.token = () => '' } }, async ({ page, saved }) => {
    assert.equal(page.data.loginRequired, true)
    await page.onChooseAvatar({ detail: { avatarUrl: 'wxfile://selected' } })
    assert.deepEqual(saved, [])
  })
})

test('an upload finishing after return refreshes the visible Mine page', async () => {
  const previousPages = global.getCurrentPages
  try {
    await withProfile({}, async ({ page }) => {
      let finishUpload, refreshes = 0
      api.uploadAvatar = () => new Promise(resolve => { finishUpload = resolve })
      const uploading = page.onChooseAvatar({ detail: { avatarUrl: 'wxfile://selected' } })
      page.onUnload()
      global.getCurrentPages = () => [{ route: 'pages/mine/index', loadProfile() { refreshes++ } }]
      finishUpload()
      await uploading
      assert.equal(refreshes, 1)
    })
  } finally { global.getCurrentPages = previousPages }
})
