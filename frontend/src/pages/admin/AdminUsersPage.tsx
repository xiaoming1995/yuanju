import { useCallback, useEffect, useState } from 'react'
import { adminRegistrationSettingsAPI, adminStatsAPI, adminUsersAPI } from '../../lib/adminApi'
import { ConfirmDialog } from '../../components/ui/ConfirmDialog'
import { StatusBadge } from '../../components/ui/StatusBadge'
import { useToast } from '../../components/ui/useToast'

interface User {
  id: string
  email: string
  nickname: string
  source: 'self_registered' | 'admin_created' | string
  created_at: string
  chart_count: number
  disabled_at?: string | null
  compat_count?: number
}

const initialForm = { email: '', nickname: '', password: '' }

type PendingAction =
  | { type: 'disable'; user: User }
  | { type: 'delete'; user: User }
  | null

export default function AdminUsersPage() {
  const { showToast } = useToast()
  const [users, setUsers] = useState<User[]>([])
  const [total, setTotal] = useState(0)
  const [query, setQuery] = useState('')
  const [loading, setLoading] = useState(true)
  const [registration_enabled, setRegistrationEnabled] = useState(false)
  const [registrationSaving, setRegistrationSaving] = useState(false)
  const [settingsMsg, setSettingsMsg] = useState('')
  const [showModal, setShowModal] = useState(false)
  const [form, setForm] = useState(initialForm)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  const [page, setPage] = useState(1)
  const [pendingAction, setPendingAction] = useState<PendingAction>(null)
  const [actionSaving, setActionSaving] = useState(false)

  const load = useCallback((q: string, pageNum: number) => {
    setLoading(true)
    adminStatsAPI.users(pageNum, q)
      .then(r => { setUsers(r.data.users || []); setTotal(r.data.total || 0) })
      .finally(() => setLoading(false))
  }, [])

  const loadSettings = () => {
    adminRegistrationSettingsAPI.get()
      .then(r => setRegistrationEnabled(r.data.registration_enabled))
      .catch(() => setSettingsMsg('注册设置读取失败'))
  }

  useEffect(() => {
    loadSettings()
  }, [])

  useEffect(() => {
    load(query, page)
  }, [page]) // eslint-disable-line react-hooks/exhaustive-deps

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    setPage(1); load(query, 1)
  }

  const toggleRegistration = async () => {
    const next = !registration_enabled
    setRegistrationSaving(true)
    setSettingsMsg('')
    try {
      const res = await adminRegistrationSettingsAPI.update({ registration_enabled: next })
      setRegistrationEnabled(res.data.registration_enabled)
      setSettingsMsg('已保存')
      showToast(res.data.registration_enabled ? '已开启公开注册' : '已关闭公开注册', 'success')
    } catch (e: unknown) {
      setSettingsMsg(e instanceof Error ? e.message : '保存失败')
      showToast(e instanceof Error ? e.message : '保存失败', 'error')
    } finally {
      setRegistrationSaving(false)
    }
  }

  const openCreate = () => {
    setForm(initialForm)
    setError('')
    setShowModal(true)
  }

  const handleCreate = async () => {
    setError('')
    if (!form.email || !form.password) {
      setError('请填写邮箱和初始密码')
      return
    }
    if (form.password.length < 8) {
      setError('初始密码至少需要8位')
      return
    }
    setSaving(true)
    try {
      await adminUsersAPI.create(form)
      setShowModal(false)
      setForm(initialForm)
      showToast('用户已创建', 'success')
      load(query, page)
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : '创建失败')
    } finally {
      setSaving(false)
    }
  }

  const [resetTarget, setResetTarget] = useState<User | null>(null)
  const [resetPwd, setResetPwd] = useState('')
  const [resetErr, setResetErr] = useState('')
  const [resetSaving, setResetSaving] = useState(false)

  const openReset = (u: User) => { setResetTarget(u); setResetPwd(''); setResetErr('') }

  const submitReset = async () => {
    if (resetPwd.length < 8) { setResetErr('新密码至少需要8位'); return }
    setResetSaving(true)
    try {
      await adminUsersAPI.resetPassword(resetTarget!.id, resetPwd)
      setResetTarget(null)
      showToast('已重置，请通过安全渠道告知用户新密码。', 'success')
    } catch (e: unknown) {
      setResetErr(e instanceof Error ? e.message : '重置失败')
    } finally {
      setResetSaving(false)
    }
  }

  const toggleDisabled = async (u: User) => {
    setPendingAction({ type: 'disable', user: u })
  }

  const removeUser = async (u: User) => {
    setPendingAction({ type: 'delete', user: u })
  }

  const confirmPendingAction = async () => {
    if (!pendingAction) return
    setActionSaving(true)
    try {
      if (pendingAction.type === 'disable') {
        const next = !pendingAction.user.disabled_at
        await adminUsersAPI.setDisabled(pendingAction.user.id, next)
        showToast(next ? '用户已禁用' : '用户已解禁', 'success')
      } else {
        await adminUsersAPI.remove(pendingAction.user.id)
        showToast('用户已删除', 'success')
      }
      setPendingAction(null)
      load(query, page)
    } catch (e: unknown) {
      showToast(e instanceof Error ? e.message : '操作失败', 'error')
    } finally {
      setActionSaving(false)
    }
  }

  const sourceLabel = (source: string) => source === 'admin_created' ? '后台创建' : '公开注册'

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <h1 className="admin-page-title" style={{ margin: 0 }}>用户列表</h1>
        <span style={{ color: '#888', fontSize: 14 }}>共 {total} 名用户</span>
      </div>

      <div className="admin-card" style={{ marginBottom: 20 }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', gap: 16, alignItems: 'center', flexWrap: 'wrap' }}>
          <div>
            <div style={{ color: '#e8e8e8', fontWeight: 700, marginBottom: 4 }}>公开注册</div>
            <div style={{ color: '#888', fontSize: 13 }}>
              当前状态：{registration_enabled ? '允许访客自行注册' : '关闭公开注册，仅后台可创建用户'}
            </div>
          </div>
          <div style={{ display: 'flex', gap: 12, alignItems: 'center' }}>
            {settingsMsg && <span style={{ color: settingsMsg === '已保存' ? '#22c55e' : '#ef4444', fontSize: 13 }}>{settingsMsg}</span>}
            <button className="admin-btn admin-btn-ghost" onClick={toggleRegistration} disabled={registrationSaving}>
              {registrationSaving ? '保存中...' : registration_enabled ? '关闭公开注册' : '开启公开注册'}
            </button>
            <button className="admin-btn admin-btn-primary" onClick={openCreate}>创建用户</button>
          </div>
        </div>
      </div>

      <form onSubmit={handleSearch} className="admin-search-bar">
        <input
          id="admin-user-search"
          className="admin-search-input"
          value={query}
          onChange={e => setQuery(e.target.value)}
          placeholder="搜索邮箱..."
        />
        <button type="submit" className="admin-btn admin-btn-primary">搜索</button>
        {query && (
          <button type="button" className="admin-btn admin-btn-ghost"
            onClick={() => { setQuery(''); setPage(1); load('', 1) }}>清除</button>
        )}
      </form>

      {loading ? <div className="admin-loading">加载中...</div> : (
        <div className="admin-card" style={{ padding: 0, overflow: 'hidden' }}>
          <table className="admin-table">
            <thead>
              <tr><th>邮箱</th><th>昵称</th><th>来源</th><th>命盘数</th><th>状态</th><th>注册时间</th><th>操作</th></tr>
            </thead>
            <tbody>
              {!users?.length && (
                <tr><td colSpan={7} style={{ textAlign: 'center', color: '#666', padding: 40 }}>
                  暂无用户数据
                </td></tr>
              )}
              {users?.map(u => (
                <tr key={u.id}>
                  <td style={{ color: '#e8e8e8' }}>{u.email}</td>
                  <td style={{ color: '#aaa' }}>{u.nickname || '—'}</td>
                  <td style={{ color: u.source === 'admin_created' ? '#c9a227' : '#888', fontSize: 13 }}>
                    {sourceLabel(u.source)}
                  </td>
                  <td>
                    <span style={{
                      background: u.chart_count > 0 ? 'rgba(201,162,39,0.15)' : 'transparent',
                      color: u.chart_count > 0 ? '#c9a227' : '#666',
                      padding: '2px 8px', borderRadius: 12, fontSize: 13
                    }}>{u.chart_count}</span>
                  </td>
                  <td>
                    {u.disabled_at
                      ? <StatusBadge tone="danger">已禁用</StatusBadge>
                      : <StatusBadge tone="success">正常</StatusBadge>}
                  </td>
                  <td style={{ color: '#666', fontSize: 13 }}>
                    {new Date(u.created_at).toLocaleDateString('zh-CN')}
                  </td>
                  <td style={{ whiteSpace: 'nowrap' }}>
                    <button className="admin-btn admin-btn-ghost" style={{ marginRight: 6 }}
                      onClick={() => openReset(u)}>重置密码</button>
                    <button className="admin-btn admin-btn-ghost" style={{ marginRight: 6 }}
                      onClick={() => toggleDisabled(u)}>{u.disabled_at ? '解禁' : '禁用'}</button>
                    <button className="admin-btn admin-btn-ghost" style={{ color: '#ff6b6b' }}
                      onClick={() => removeUser(u)}>删除</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {(() => {
        const totalPages = Math.ceil((total || 0) / 20) || 1
        if (totalPages <= 1) return null
        return (
          <div style={{ display: 'flex', justifyContent: 'center', gap: 8, marginTop: 20 }}>
            <button disabled={page === 1} onClick={() => setPage(p => p - 1)} style={{ padding: '6px 14px', borderRadius: 8, border: 'none', background: page === 1 ? '#1a1a2e' : '#2a2a3a', color: page === 1 ? '#555' : '#ccc', cursor: page === 1 ? 'not-allowed' : 'pointer' }}>上一页</button>
            <span style={{ lineHeight: '32px', fontSize: 13, color: '#666', margin: '0 8px' }}>第 {page} / {totalPages} 页</span>
            <button disabled={page >= totalPages} onClick={() => setPage(p => p + 1)} style={{ padding: '6px 14px', borderRadius: 8, border: 'none', background: page >= totalPages ? '#1a1a2e' : '#2a2a3a', color: page >= totalPages ? '#555' : '#ccc', cursor: page >= totalPages ? 'not-allowed' : 'pointer' }}>下一页</button>
          </div>
        )
      })()}

      {showModal && (
        <div className="admin-modal-overlay" onClick={e => e.target === e.currentTarget && setShowModal(false)}>
          <div className="admin-modal">
            <div className="admin-modal-title">创建用户</div>
            {error && <div className="admin-error">{error}</div>}
            <div className="admin-form-group" style={{ marginBottom: 16 }}>
              <label className="admin-form-label">邮箱</label>
              <input className="admin-form-input" type="email" value={form.email}
                onChange={e => setForm(f => ({ ...f, email: e.target.value }))} />
            </div>
            <div className="admin-form-group" style={{ marginBottom: 16 }}>
              <label className="admin-form-label">昵称（选填）</label>
              <input className="admin-form-input" value={form.nickname}
                onChange={e => setForm(f => ({ ...f, nickname: e.target.value }))} />
            </div>
            <div className="admin-form-group">
              <label className="admin-form-label">初始密码</label>
              <input className="admin-form-input" type="password" value={form.password}
                onChange={e => setForm(f => ({ ...f, password: e.target.value }))} />
            </div>
            <div style={{ color: '#888', fontSize: 12, marginTop: 10 }}>
              请通过安全渠道告知用户初始密码。
            </div>
            <div className="admin-modal-actions">
              <button className="admin-btn admin-btn-ghost" onClick={() => setShowModal(false)}>取消</button>
              <button className="admin-btn admin-btn-primary" onClick={handleCreate} disabled={saving}>
                {saving ? '创建中...' : '创建'}
              </button>
            </div>
          </div>
        </div>
      )}

      {resetTarget && (
        <div className="admin-modal-overlay" onClick={e => e.target === e.currentTarget && setResetTarget(null)}>
          <div className="admin-modal">
            <div className="admin-modal-title">重置密码 — {resetTarget.email}</div>
            {resetErr && <div className="admin-error">{resetErr}</div>}
            <div className="admin-form-group">
              <label className="admin-form-label">新密码（至少8位）</label>
              <input className="admin-form-input" type="password" value={resetPwd}
                onChange={e => setResetPwd(e.target.value)} />
            </div>
            <div className="admin-modal-actions">
              <button className="admin-btn admin-btn-ghost" onClick={() => setResetTarget(null)}>取消</button>
              <button className="admin-btn admin-btn-primary" onClick={submitReset} disabled={resetSaving}>
                {resetSaving ? '重置中...' : '确认重置'}
              </button>
            </div>
          </div>
        </div>
      )}

      <ConfirmDialog
        open={!!pendingAction}
        title={pendingAction?.type === 'delete' ? '删除用户' : pendingAction?.user.disabled_at ? '解禁用户' : '禁用用户'}
        description={
          pendingAction?.type === 'delete'
            ? `确认删除用户 ${pendingAction.user.email}？将连带删除其 ${pendingAction.user.compat_count || 0} 条合盘记录（不可恢复），其八字命盘将转为游客记录保留。`
            : pendingAction
              ? pendingAction.user.disabled_at
                ? `确认解禁用户 ${pendingAction.user.email}？`
                : `确认禁用用户 ${pendingAction.user.email}？禁用后该用户将无法再次登录，已签发的登录令牌在到期前仍有效。`
              : ''
        }
        confirmText={pendingAction?.type === 'delete' ? '删除' : pendingAction?.user.disabled_at ? '解禁' : '禁用'}
        danger={pendingAction?.type === 'delete' || !pendingAction?.user.disabled_at}
        pending={actionSaving}
        onCancel={() => setPendingAction(null)}
        onConfirm={confirmPendingAction}
      />
    </div>
  )
}
