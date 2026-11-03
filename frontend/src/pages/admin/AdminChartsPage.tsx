import { useLayoutEffect, useRef, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { BookOpen, User, Calendar, RefreshCw } from 'lucide-react'
import { adminChartsAPI } from '../../lib/adminApi'
import { AdminRequestState } from '../../components/AdminRequestState'

interface ChartRecord {
  id: string
  user_id?: string
  user_email?: string
  birth_year: number
  birth_month: number
  birth_day: number
  birth_hour: number
  gender: string
  year_gan: string
  year_zhi: string
  month_gan: string
  month_zhi: string
  day_gan: string
  day_zhi: string
  hour_gan: string
  hour_zhi: string
  yongshen: string
  jishen: string
  ai_result?: string
  ai_result_structured?: {
    // 新版 chapters 格式
    chapters?: Array<{ title: string; brief: string; detail?: string }>
    analysis?: { logic?: string; summary?: string }
    yongshen?: string
    jishen?: string
    // 旧版平铺格式（降级兼容）
    personality?: string
    career?: string
    romance?: string
    health?: string
  }
  created_at: string
}

interface ChartQueryRecord {
  chart_id?: string
  queried_at: string
  deleted: boolean
}

const genderLabel = (g: string) => g === 'male' ? '男' : '女'

export default function AdminChartsPage() {
  const [searchParams] = useSearchParams()
  const userID = searchParams.get('user_id') || ''
  const [charts, setCharts] = useState<ChartRecord[]>([])
  const [queries, setQueries] = useState<ChartQueryRecord[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const requestSerial = useRef(0)
  const previousUserID = useRef(userID)
  const [q, setQ] = useState('')
  const [from, setFrom] = useState('')
  const [to, setTo] = useState('')
  const pageSize = 20

  const fetchCharts = async (pageNum: number, filters = { q, from, to }, scopedUserID = userID) => {
    const serial = ++requestSerial.current
    try {
      setLoading(true)
      setError('')
      const res = await adminChartsAPI.list(pageNum, pageSize, filters.q, filters.from, filters.to, scopedUserID)
      if (serial !== requestSerial.current) return
      if (scopedUserID) setQueries(res.data?.data || [])
      else setCharts(res.data?.data || [])
      setTotal(res.data?.total || 0)
    } catch (err: unknown) {
      if (serial === requestSerial.current) setError(err instanceof Error ? err.message : '起盘明细加载失败')
    } finally {
      if (serial === requestSerial.current) setLoading(false)
    }
  }

  useLayoutEffect(() => {
    if (previousUserID.current !== userID) {
      previousUserID.current = userID
      requestSerial.current += 1
      setCharts([])
      setQueries([])
      setTotal(0)
      setLoading(true)
      setError('')
      setQ('')
      setFrom('')
      setTo('')
      if (page !== 1) { setPage(1); return }
      void fetchCharts(1, { q: '', from: '', to: '' }, userID)
      return
    }
    void fetchCharts(page)
  }, [page, userID]) // eslint-disable-line react-hooks/exhaustive-deps

  const totalPages = Math.ceil((total || 0) / pageSize) || 1
  const hasRecords = userID ? queries.length > 0 : charts.length > 0

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: 12, marginBottom: 24 }}>
        <h1 className="admin-page-title" style={{ display: 'flex', alignItems: 'center', gap: 8, margin: 0 }}>
          <BookOpen size={24} /> {userID ? '用户起盘记录' : '全站起盘明细'}
        </h1>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          {userID && <Link to="/admin/charts" className="admin-btn admin-btn-ghost" style={{ display: 'inline-flex', alignItems: 'center', textDecoration: 'none' }}>查看全站</Link>}
          <button
            onClick={() => fetchCharts(page)}
            style={{
              display: 'flex', alignItems: 'center', gap: 6,
              padding: '6px 14px', borderRadius: 8, border: '1px solid var(--admin-border-strong)',
              background: 'var(--admin-bg-soft)', color: 'var(--admin-text-secondary)', cursor: 'pointer', fontSize: 13
            }}
          >
            <RefreshCw size={14} /> 刷新
          </button>
        </div>
      </div>

      <div className="admin-card" aria-busy={loading}>
        <div style={{ marginBottom: 16, fontSize: 13, color: 'var(--admin-text-muted)' }}>
          {userID
            ? <>用户 ID <code style={{ overflowWrap: 'anywhere' }}>{userID}</code> 的查盘记录，共 {total} 条；已删除的记录仅保留查询时间和状态，迁移前无法追溯的旧记录可能缺失。</>
            : <>平台当前保存的干支排盘记录（包含注册用户与游客），共 {total} 条。</>}
        </div>

        <form className="admin-search-bar" onSubmit={e => { e.preventDefault(); if (page === 1) void fetchCharts(1); else setPage(1) }}>
          {!userID && <input className="admin-search-input" value={q} onChange={e => setQ(e.target.value)} placeholder="按邮箱搜索..." />}
          <input type="date" className="admin-search-input" value={from} onChange={e => setFrom(e.target.value)} title="排盘起始日期" />
          <input type="date" className="admin-search-input" value={to} onChange={e => setTo(e.target.value)} title="排盘截止日期" />
          <button type="submit" className="admin-btn admin-btn-primary">搜索</button>
          {(q || from || to) && (
            <button type="button" className="admin-btn admin-btn-ghost"
              onClick={() => { setQ(''); setFrom(''); setTo(''); if (page === 1) void fetchCharts(1, { q: '', from: '', to: '' }); else setPage(1) }}>清除</button>
          )}
        </form>

        {loading && !hasRecords ? (
          <AdminRequestState state="loading" message="正在加载起盘明细…" />
        ) : error && !hasRecords ? (
          <AdminRequestState state="error" message={error} onRetry={() => void fetchCharts(page)} />
        ) : !hasRecords ? (
          <AdminRequestState state="empty" message="当前条件下暂无起盘记录" />
        ) : (
          <>
          {loading && <AdminRequestState state="updating" message="正在更新起盘明细，当前显示上一结果…" />}
          {error && <AdminRequestState state="error" message={`${error}；当前结果未更新`} onRetry={() => void fetchCharts(page)} />}
          {userID ? (
          <table className="admin-table">
            <thead><tr><th>查询时间</th><th>状态</th><th>操作</th></tr></thead>
            <tbody>
              {queries.map((record, index) => (
                <tr key={record.chart_id || `${record.queried_at}-${index}`}>
                  <td>{new Date(record.queried_at).toLocaleString('zh-CN')}</td>
                  <td>{record.deleted ? <span style={{ color: 'var(--admin-text-faint)' }}>已删除</span> : <span>现存</span>}</td>
                  <td>{record.deleted || !record.chart_id
                    ? <span style={{ color: 'var(--admin-text-faint)' }}>命盘内容已删除</span>
                    : <Link to={`/admin/charts/${record.chart_id}?user_id=${encodeURIComponent(userID)}`} className="admin-btn admin-btn-ghost" style={{ display: 'inline-block', padding: '5px 12px', textDecoration: 'none' }}>查看详情</Link>}</td>
                </tr>
              ))}
            </tbody>
          </table>
          ) : (
          <table className="admin-table">
            <thead>
              <tr>
                <th>排盘用户</th>
                <th>排盘时间</th>
                <th>测算命主</th>
                <th>简述</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
	              {charts.map(chart => (
	                <tr key={chart.id}>
                    <td>
                      <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                        <User size={14} color="var(--admin-text-muted)" />
                        {chart.user_email ? (
                          <span style={{ color: 'var(--admin-text-secondary)' }}>{chart.user_email}</span>
                        ) : chart.user_id ? (
                          <span style={{ color: 'var(--admin-text-secondary)' }}>微信用户 · {chart.user_id.slice(0, 8)}</span>
                        ) : (
                          <span style={{ color: 'var(--admin-text-faint)', fontStyle: 'italic' }}>游客</span>
                        )}
                      </div>
                    </td>
                    <td style={{ fontSize: 12, color: 'var(--admin-text-muted)' }}>
                      {new Date(chart.created_at).toLocaleString('zh-CN')}
                    </td>
                    <td>
                      <div style={{ color: 'var(--admin-text-secondary)', marginBottom: 4 }}>{genderLabel(chart.gender)}命</div>
                      <div style={{ fontSize: 11, color: 'var(--admin-text-faint)', display: 'flex', alignItems: 'center', gap: 4 }}>
                        <Calendar size={11} /> {chart.birth_year}年{chart.birth_month}月{chart.birth_day}日
                      </div>
                    </td>
                    <td>
                      <span style={{
                        display: 'inline-block', padding: '2px 8px', borderRadius: 4,
                        background: 'rgba(var(--admin-danger-rgb), 0.15)', color: 'var(--admin-danger)', fontSize: 12, fontWeight: 600
                      }}>
                        {chart.year_gan}{chart.year_zhi}年
                      </span>
                    </td>
                    <td>
	                      <Link to={`/admin/charts/${chart.id}`} className="admin-btn admin-btn-ghost" style={{ display: 'inline-block', padding: '5px 12px', textDecoration: 'none' }}>
	                        查看详情
	                      </Link>
                    </td>
	                  </tr>
	              ))}
            </tbody>
          </table>
          )}
          </>
        )}

        {/* 分页器 */}
        {totalPages > 1 && (
          <div style={{ display: 'flex', justifyContent: 'center', gap: 8, marginTop: 20 }}>
            <button
              disabled={page === 1}
              onClick={() => setPage(p => p - 1)}
              style={{
                padding: '6px 14px', borderRadius: 8, border: 'none',
                background: page === 1 ? 'var(--admin-bg-muted)' : 'var(--admin-bg-soft)',
                color: page === 1 ? 'var(--admin-text-faint)' : 'var(--admin-text-secondary)', cursor: page === 1 ? 'not-allowed' : 'pointer'
              }}
            >
              上一页
            </button>
            <span style={{ lineHeight: '32px', fontSize: 13, color: 'var(--admin-text-faint)', margin: '0 8px' }}>
              第 {page} / {totalPages} 页
            </span>
            <button
              disabled={page >= totalPages}
              onClick={() => setPage(p => p + 1)}
              style={{
                padding: '6px 14px', borderRadius: 8, border: 'none',
                background: page >= totalPages ? 'var(--admin-bg-muted)' : 'var(--admin-bg-soft)',
                color: page >= totalPages ? 'var(--admin-text-faint)' : 'var(--admin-text-secondary)', cursor: page >= totalPages ? 'not-allowed' : 'pointer'
              }}
            >
              下一页
            </button>
          </div>
        )}
      </div>
    </div>
  )
}
