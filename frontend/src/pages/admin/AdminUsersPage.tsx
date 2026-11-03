import { useCallback, useEffect, useRef, useState } from 'react'
import { Link } from 'react-router-dom'
import { adminActionErrorMessage, adminMiniappRegistrationSettingsAPI, adminRegistrationSettingsAPI, adminStatsAPI, adminUsersAPI } from '../../lib/adminApi'
import { ConfirmDialog } from '../../components/ui/ConfirmDialog'
import { StatusBadge } from '../../components/ui/StatusBadge'
import { useToast } from '../../components/ui/useToast'
import { AdminUserPagination } from './AdminUserPagination'
import { AdminRequestState } from '../../components/AdminRequestState'
import './AdminUserActivity.css'

interface User {
  id: string
  email: string
  nickname: string
  source: 'self_registered' | 'admin_created' | string
  created_at: string
  chart_count: number
  chart_query_count: number
  disabled_at?: string | null
  compat_count?: number
  last_login_at?: string | null
  last_login_channel?: string
  last_active_at?: string | null
}

interface UserDetail extends User { ai_requests_this_month: number }
interface ActivityEvent { id: string; event_type: string; channel: string; object_id?: string | null; created_at: string }
interface AdminAuditEvent { id: string; admin_id: string; admin_name: string; action: string; created_at: string }
const eventLabels: Record<string, string> = {
  register: '注册账号', login: '登录', account_linked: '绑定网页账号', chart_created: '创建排盘',
  compatibility_created: '创建合盘', bazi_report_generated: '生成八字报告',
  compatibility_report_generated: '生成合盘报告', nickname_updated: '修改昵称', avatar_updated: '更新头像',
}
const auditLabels: Record<string, string> = { created: '后台创建账号', disabled: '禁用账号', enabled: '解禁账号', password_reset: '重置密码' }
const formatTime = (value?: string | null) => value ? new Date(value).toLocaleString('zh-CN', { hour12: false }) : '暂无记录'
const formatDate = (value: string) => new Date(value).toLocaleDateString('zh-CN')
const formatClock = (value: string) => new Date(value).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit', hour12: false })
const defaultFilters = { source: '', status: '', activity: '', sort: 'created_desc' }

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
  const [appliedQuery, setAppliedQuery] = useState('')
  const [filters, setFilters] = useState(defaultFilters)
  const [selectedID, setSelectedID] = useState<string | null>(null)
  const [detail, setDetail] = useState<UserDetail | null>(null)
  const [events, setEvents] = useState<ActivityEvent[]>([])
  const [adminEvents, setAdminEvents] = useState<AdminAuditEvent[]>([])
  const [eventsPage, setEventsPage] = useState(1)
  const [eventsTotal, setEventsTotal] = useState(0)
  const [detailLoading, setDetailLoading] = useState(false)
  const [detailError, setDetailError] = useState('')
  const [detailRevision, setDetailRevision] = useState(0)
  const [loading, setLoading] = useState(true)
  const [listError, setListError] = useState('')
  const requestSerial = useRef(0)
  const [registration_enabled, setRegistrationEnabled] = useState(false)
  const [settingsLoading, setSettingsLoading] = useState(true)
  const [registrationSaving, setRegistrationSaving] = useState(false)
  const [miniappRegistrationEnabled, setMiniappRegistrationEnabled] = useState(false)
  const [miniappRegistrationSaving, setMiniappRegistrationSaving] = useState(false)
  const [settingsMsg, setSettingsMsg] = useState('')
  const [showModal, setShowModal] = useState(false)
  const [form, setForm] = useState(initialForm)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [pendingAction, setPendingAction] = useState<PendingAction>(null)
  const [actionSaving, setActionSaving] = useState(false)

  const load = useCallback((q: string, pageNum: number, nextFilters = filters) => {
    const serial = ++requestSerial.current
    setLoading(true)
    setListError('')
    adminStatsAPI.users(pageNum, q, nextFilters, pageSize)
      .then(r => {
        if (serial !== requestSerial.current) return
        const nextTotal = r.data.total || 0
        setTotal(nextTotal)
        const lastPage = Math.max(1, Math.ceil(nextTotal / pageSize))
        if (pageNum > lastPage) { setPage(lastPage); return }
        setUsers(r.data.users || [])
      })
      .catch((e: unknown) => { if (serial === requestSerial.current) setListError(e instanceof Error ? e.message : '用户列表加载失败') })
      .finally(() => { if (serial === requestSerial.current) setLoading(false) })
  }, [filters, pageSize])

  const loadSettings = () => {
    setSettingsLoading(true)
    setSettingsMsg('')
    const web = adminRegistrationSettingsAPI.get()
      .then(r => setRegistrationEnabled(r.data.registration_enabled))
      .catch(() => setSettingsMsg('注册设置读取失败'))
    const miniapp = adminMiniappRegistrationSettingsAPI.get()
      .then(r => setMiniappRegistrationEnabled(r.data.miniapp_registration_enabled))
      .catch(() => setSettingsMsg('小程序注册设置读取失败'))
    void Promise.allSettled([web, miniapp]).then(() => setSettingsLoading(false))
  }

  useEffect(() => {
    loadSettings()
  }, [])

  useEffect(() => {
    load(appliedQuery, page, filters)
  }, [page, appliedQuery, filters, load])

  useEffect(() => {
    if (!selectedID) return
    let active = true
    setDetailLoading(true)
    setDetailError('')
    Promise.all([adminUsersAPI.detail(selectedID), adminUsersAPI.events(selectedID, eventsPage)])
      .then(([userResponse, eventResponse]) => {
        if (!active) return
        setDetail(userResponse.data.data)
        setEvents(eventResponse.data.events || [])
        setAdminEvents(eventResponse.data.admin_events || [])
        setEventsTotal(eventResponse.data.total || 0)
      })
      .catch((e: unknown) => { if (active) setDetailError(e instanceof Error ? e.message : '用户详情加载失败') })
      .finally(() => { if (active) setDetailLoading(false) })
    return () => { active = false }
  }, [selectedID, eventsPage, detailRevision, showToast])

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    setPage(1); setAppliedQuery(query)
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
      setSettingsMsg(adminActionErrorMessage(e))
      showToast(adminActionErrorMessage(e), 'error')
    } finally {
      setRegistrationSaving(false)
    }
  }

  const toggleMiniappRegistration = async () => {
    setMiniappRegistrationSaving(true)
    setSettingsMsg('')
    try {
      const res = await adminMiniappRegistrationSettingsAPI.update({ miniapp_registration_enabled: !miniappRegistrationEnabled })
      setMiniappRegistrationEnabled(res.data.miniapp_registration_enabled)
      setSettingsMsg('已保存')
      showToast(res.data.miniapp_registration_enabled ? '已开启小程序新用户注册' : '已关闭小程序新用户注册', 'success')
    } catch (e: unknown) {
      setSettingsMsg(adminActionErrorMessage(e))
    } finally {
      setMiniappRegistrationSaving(false)
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
      load(appliedQuery, page)
    } catch (e: unknown) {
      setError(adminActionErrorMessage(e))
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
      setDetailRevision(n => n + 1)
      showToast('已重置，请通过安全渠道告知用户新密码。', 'success')
    } catch (e: unknown) {
      setResetErr(adminActionErrorMessage(e))
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
      if (pendingAction.type === 'delete' && selectedID === pendingAction.user.id) setSelectedID(null)
      setDetailRevision(n => n + 1)
      load(appliedQuery, page)
    } catch (e: unknown) {
      showToast(adminActionErrorMessage(e), 'error')
    } finally {
      setActionSaving(false)
    }
  }

  const sourceLabel = (source: string) => source === 'admin_created' ? '后台创建' : source === 'miniapp_wechat' ? '微信小程序' : '网页注册'
  const userLabel = (user: User) => user.email || `${user.nickname || '微信用户'}（${user.id.slice(0, 8)}）`

  return (
    <div className="admin-users-page">
      <div className="admin-users-heading">
        <div><h1 className="admin-page-title">用户管理</h1><p>查看登录、活跃和关键行为记录</p></div>
        <div className="admin-users-heading-actions"><span>共 {total} 名用户</span><button className="admin-btn admin-btn-primary" onClick={openCreate}>创建用户</button></div>
      </div>

      <section className="admin-card admin-user-settings" aria-label="注册设置" aria-busy={settingsLoading || registrationSaving || miniappRegistrationSaving}>
        <div className="admin-user-settings-heading"><strong>注册设置</strong>{settingsMsg && <span role="status" className={settingsMsg === '已保存' ? 'is-success' : 'is-error'}>{settingsMsg}</span>}</div>
        {settingsLoading && <AdminRequestState state="updating" message="正在读取注册设置…" />}
        {!settingsLoading && settingsMsg.includes('读取失败') && <button className="admin-btn admin-btn-ghost" onClick={loadSettings}>重试读取</button>}
        <div className="admin-user-setting">
          <div><strong>网页邮箱注册</strong><p>{registration_enabled ? '访客可自行注册邮箱账号' : '仅管理员可创建邮箱账号'}</p></div>
          <div className="admin-user-setting-control"><span className={registration_enabled ? 'is-enabled' : ''}>{settingsLoading ? '读取中' : registration_enabled ? '已开启' : '已关闭'}</span><button className="admin-btn admin-btn-ghost" onClick={toggleRegistration} disabled={settingsLoading || registrationSaving || settingsMsg.includes('读取失败')}>
            {registrationSaving ? '保存中...' : registration_enabled ? '关闭注册' : '开启注册'}
          </button></div>
        </div>
        <div className="admin-user-setting">
          <div><strong>小程序微信注册</strong><p>控制首次微信建号；已绑定账号仍可登录</p></div>
          <div className="admin-user-setting-control"><span className={miniappRegistrationEnabled ? 'is-enabled' : ''}>{settingsLoading ? '读取中' : miniappRegistrationEnabled ? '已开启' : '已关闭'}</span><button className="admin-btn admin-btn-ghost" onClick={toggleMiniappRegistration} disabled={settingsLoading || miniappRegistrationSaving || settingsMsg.includes('读取失败')}>
            {miniappRegistrationSaving ? '保存中...' : miniappRegistrationEnabled ? '关闭注册' : '开启注册'}
          </button></div>
        </div>
      </section>

      <section className="admin-user-toolbar" aria-label="搜索和筛选用户">
        <form onSubmit={handleSearch} className="admin-search-bar">
          <input id="admin-user-search" className="admin-search-input" value={query}
            onChange={e => setQuery(e.target.value)} placeholder="搜索邮箱、昵称或用户 ID" />
          <button type="submit" className="admin-btn admin-btn-primary">搜索</button>
          {query && <button type="button" className="admin-btn admin-btn-ghost"
            onClick={() => { setQuery(''); setAppliedQuery(''); setPage(1) }}>清除</button>}
        </form>
        <div className="admin-user-filters" aria-label="用户筛选">
          <label><span>来源</span><select value={filters.source} onChange={e => { setFilters(f => ({ ...f, source: e.target.value })); setPage(1) }}>
            <option value="">全部来源</option><option value="self_registered">网页注册</option><option value="admin_created">后台创建</option><option value="miniapp_wechat">微信小程序</option>
          </select></label>
          <label><span>状态</span><select value={filters.status} onChange={e => { setFilters(f => ({ ...f, status: e.target.value })); setPage(1) }}>
            <option value="">全部状态</option><option value="enabled">正常</option><option value="disabled">已禁用</option>
          </select></label>
          <label><span>活跃情况</span><select value={filters.activity} onChange={e => { setFilters(f => ({ ...f, activity: e.target.value })); setPage(1) }}>
            <option value="">全部用户</option><option value="never">从未登录</option><option value="active7">近 7 天活跃</option><option value="active30">近 30 天活跃</option><option value="inactive30">30 天未活跃</option>
          </select></label>
          <label><span>排序</span><select value={filters.sort} onChange={e => { setFilters(f => ({ ...f, sort: e.target.value })); setPage(1) }}>
            <option value="created_desc">最近注册</option><option value="last_login_desc">最近登录</option><option value="last_active_desc">最近活跃</option>
          </select></label>
        </div>
      </section>

      {loading && !users.length && !listError ? <AdminRequestState state="loading" message="正在加载用户列表…" /> : (
        <div className="admin-card admin-user-list-card" id="admin-user-list" aria-busy={loading}>
          {loading && <AdminRequestState state="updating" message="正在更新用户列表，当前显示上一结果…" />}
          {listError && <AdminRequestState state="error" message={`${listError}；当前结果未更新`} onRetry={() => load(appliedQuery, page, filters)} />}
          <div className="admin-user-list-heading">
            <div><strong>用户记录</strong><span>点击用户名或累计查盘次数查看命盘记录；点击“用户详情”查看活动与管理操作</span></div>
            <AdminUserPagination page={page} pageSize={pageSize} total={total} onPageChange={setPage}
              onPageSizeChange={next => { setPage(1); setPageSize(next) }} />
          </div>
          <p className="admin-user-mobile-hint">左右滑动表格，查看登录与活跃信息</p>
          <div className="admin-user-table-wrap">
          <table className="admin-table">
            <thead>
              <tr><th>用户</th><th>来源</th><th>累计查盘</th><th>状态</th><th>注册时间</th><th>最后登录</th><th>最近活跃</th><th>操作</th></tr>
            </thead>
            <tbody>
              {!users?.length && (
                <tr><td colSpan={8} className="admin-user-empty-row">
                  没有符合条件的用户
                </td></tr>
              )}
              {users?.map(u => (
                <tr key={u.id}>
                  <td className="admin-user-identity"><Link className="admin-user-link" title={`查看 ${u.email || u.nickname || u.id} 的命盘记录`} to={`/admin/charts?user_id=${encodeURIComponent(u.id)}`}>{u.email || `未绑定邮箱 · ${u.id.slice(0, 8)}`}</Link><span>{u.nickname || '未设置昵称'}</span></td>
                  <td><span className={`admin-user-source ${u.source === 'miniapp_wechat' ? 'is-miniapp' : u.source === 'admin_created' ? 'is-admin' : ''}`}>{sourceLabel(u.source)}</span></td>
                  <td>
                    <Link className={`admin-user-query-link${u.chart_query_count === 0 ? ' is-empty' : ''}`}
                      to={`/admin/charts?user_id=${encodeURIComponent(u.id)}`}
                      aria-label={`查看 ${u.email || u.nickname || u.id} 的 ${u.chart_query_count} 次命盘查询记录`}>
                      <span>{u.chart_query_count}</span><span aria-hidden="true">›</span>
                    </Link>
                  </td>
                  <td>
                    {u.disabled_at
                      ? <StatusBadge tone="danger">已禁用</StatusBadge>
                      : <StatusBadge tone="success">正常</StatusBadge>}
                  </td>
                  <td className="admin-user-date">
                    {formatDate(u.created_at)}
                  </td>
                  <td className="admin-user-time">{u.last_login_at ? <><strong title={formatTime(u.last_login_at)}>{formatDate(u.last_login_at)}</strong><small>{formatClock(u.last_login_at)} · {u.last_login_channel === 'miniapp' ? '微信小程序' : '网页'}</small></> : <span className="admin-user-no-data">从未登录</span>}</td>
                  <td className="admin-user-time">{u.last_active_at ? <><strong title={formatTime(u.last_active_at)}>{formatDate(u.last_active_at)}</strong><small>{formatClock(u.last_active_at)}</small></> : <span className="admin-user-no-data">暂无记录</span>}</td>
                  <td className="admin-user-row-action">
                    <button className="admin-btn admin-btn-ghost" onClick={() => { setDetail(null); setSelectedID(u.id); setEventsPage(1) }}>用户详情</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          </div>
          <AdminUserPagination page={page} pageSize={pageSize} total={total}
            onPageChange={next => { setPage(next); document.getElementById('admin-user-list')?.scrollIntoView({ behavior: 'smooth', block: 'start' }) }} />
        </div>
      )}

      {selectedID && (
        <div className="admin-user-detail-backdrop" onClick={e => { if (e.target === e.currentTarget) setSelectedID(null) }}>
          <section className="admin-user-detail" role="dialog" aria-modal="true" aria-label="用户详情">
            <header className="admin-user-detail-header">
              <h2>用户详情</h2><button className="admin-btn admin-btn-ghost" onClick={() => setSelectedID(null)}>关闭</button>
            </header>
            {detailLoading && !detail ? <AdminRequestState state="loading" message="正在加载用户详情…" /> : detailError && !detail ? <AdminRequestState state="error" message={detailError} onRetry={() => setDetailRevision(n => n + 1)} /> : detail && <>
              <div className="admin-user-summary">
                <div><small>用户</small><strong>{detail.email || detail.nickname || detail.id}</strong></div>
                <div><small>来源</small><strong>{sourceLabel(detail.source)}</strong></div>
                <div><small>最后登录</small><strong>{detail.last_login_at ? `${formatTime(detail.last_login_at)} · ${detail.last_login_channel === 'miniapp' ? '微信小程序' : '网页'}` : '从未登录'}</strong></div>
                <div><small>最近活跃</small><strong>{formatTime(detail.last_active_at)}</strong></div>
                <div><small>累计查盘次数</small><strong>{detail.chart_query_count}</strong></div>
                <div><small>现存排盘 / 合盘</small><strong>{detail.chart_count} / {detail.compat_count || 0}</strong></div>
                <div><small>本月 AI 调用</small><strong>{detail.ai_requests_this_month}</strong></div>
              </div>
              <p className="admin-user-retention-note">计数启用时按现存排盘记录回填；此前已删除的排盘无法追溯。</p>
              <div className="admin-user-detail-actions">
                <Link className="admin-btn admin-btn-primary" to={`/admin/charts?user_id=${encodeURIComponent(detail.id)}`}>查看命盘记录</Link>
                {detail.email && <button className="admin-btn admin-btn-ghost" onClick={() => openReset(detail)}>重置密码</button>}
                <button className="admin-btn admin-btn-ghost" onClick={() => toggleDisabled(detail)}>{detail.disabled_at ? '解禁用户' : '禁用用户'}</button>
                <button className="admin-btn admin-btn-ghost" style={{ color: 'var(--admin-danger)' }} onClick={() => removeUser(detail)}>删除用户</button>
              </div>
              <h3>活动记录</h3>
              <p className="admin-user-retention-note">仅展示近 90 天的活动明细；最后登录和最近活跃时间会继续保留。</p>
              {events.length === 0 ? <p className="admin-user-empty">暂无记录</p> : <ol className="admin-user-events">
                {events.map(event => <li key={event.id}><span>{eventLabels[event.event_type] || event.event_type}</span><small>{event.channel === 'miniapp' ? '微信小程序' : '网页'} · {formatTime(event.created_at)}</small></li>)}
              </ol>}
              {eventsTotal > 20 && <div className="admin-user-event-pages">
                <button className="admin-btn admin-btn-ghost" disabled={eventsPage <= 1} onClick={() => setEventsPage(p => p - 1)}>上一页</button>
                <span>第 {eventsPage} / {Math.ceil(eventsTotal / 20)} 页</span>
                <button className="admin-btn admin-btn-ghost" disabled={eventsPage >= Math.ceil(eventsTotal / 20)} onClick={() => setEventsPage(p => p + 1)}>下一页</button>
              </div>}
              <h3>管理员操作</h3>
              {adminEvents.length === 0 ? <p className="admin-user-empty">暂无记录</p> : <ol className="admin-user-events">
                {adminEvents.map(event => <li key={event.id}><span>{auditLabels[event.action] || event.action}</span><small>管理员 {event.admin_name || event.admin_id.slice(0, 8)} · {formatTime(event.created_at)}</small></li>)}
              </ol>}
            </>}
          </section>
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
            <div style={{ color: 'var(--admin-text-muted)', fontSize: 12, marginTop: 10 }}>
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
            ? `确认删除用户 ${userLabel(pendingAction.user)}？将连带删除其 ${pendingAction.user.compat_count || 0} 条合盘记录（不可恢复），其干支记录将转为游客记录保留。`
            : pendingAction
              ? pendingAction.user.disabled_at
                ? `确认解禁用户 ${userLabel(pendingAction.user)}？`
                : `确认禁用用户 ${userLabel(pendingAction.user)}？禁用后该用户将无法再次登录。`
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
