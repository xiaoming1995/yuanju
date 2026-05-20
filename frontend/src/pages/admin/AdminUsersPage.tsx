import { useCallback, useEffect, useState } from 'react'
import { adminRegistrationSettingsAPI, adminStatsAPI, adminUsersAPI } from '../../lib/adminApi'

interface User {
  id: string
  email: string
  nickname: string
  source: 'self_registered' | 'admin_created' | string
  created_at: string
  chart_count: number
}

const initialForm = { email: '', nickname: '', password: '' }

export default function AdminUsersPage() {
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

  const load = useCallback((q: string) => {
    setLoading(true)
    adminStatsAPI.users(1, q)
      .then(r => { setUsers(r.data.users || []); setTotal(r.data.total || 0) })
      .finally(() => setLoading(false))
  }, [])

  const loadSettings = () => {
    adminRegistrationSettingsAPI.get()
      .then(r => setRegistrationEnabled(r.data.registration_enabled))
      .catch(() => setSettingsMsg('注册设置读取失败'))
  }

  useEffect(() => {
    load('')
    loadSettings()
  }, [load])

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    load(query)
  }

  const toggleRegistration = async () => {
    const next = !registration_enabled
    setRegistrationSaving(true)
    setSettingsMsg('')
    try {
      const res = await adminRegistrationSettingsAPI.update({ registration_enabled: next })
      setRegistrationEnabled(res.data.registration_enabled)
      setSettingsMsg('已保存')
    } catch (e: unknown) {
      setSettingsMsg(e instanceof Error ? e.message : '保存失败')
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
      load(query)
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : '创建失败')
    } finally {
      setSaving(false)
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
            onClick={() => { setQuery(''); load('') }}>清除</button>
        )}
      </form>

      {loading ? <div className="admin-loading">加载中...</div> : (
        <div className="admin-card" style={{ padding: 0, overflow: 'hidden' }}>
          <table className="admin-table">
            <thead>
              <tr><th>邮箱</th><th>昵称</th><th>来源</th><th>命盘数</th><th>注册时间</th></tr>
            </thead>
            <tbody>
              {!users?.length && (
                <tr><td colSpan={5} style={{ textAlign: 'center', color: '#666', padding: 40 }}>
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
                  <td style={{ color: '#666', fontSize: 13 }}>
                    {new Date(u.created_at).toLocaleDateString('zh-CN')}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

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
    </div>
  )
}
