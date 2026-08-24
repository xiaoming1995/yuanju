import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { BookOpen, User, Calendar, RefreshCw } from 'lucide-react'
import { adminChartsAPI } from '../../lib/adminApi'

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

const genderLabel = (g: string) => g === 'male' ? '男' : '女'

export default function AdminChartsPage() {
  const [charts, setCharts] = useState<ChartRecord[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(true)
  const [q, setQ] = useState('')
  const [from, setFrom] = useState('')
  const [to, setTo] = useState('')
  const pageSize = 20

  const fetchCharts = async (pageNum: number) => {
    try {
      setLoading(true)
      const res = await adminChartsAPI.list(pageNum, pageSize, q, from, to)
      setCharts(res.data?.data || [])
      setTotal(res.data?.total || 0)
    } catch (err: unknown) {
      console.error('获取起盘流水失败:', err)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchCharts(page)
  }, [page]) // eslint-disable-line react-hooks/exhaustive-deps

  const totalPages = Math.ceil((total || 0) / pageSize) || 1

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <h1 className="admin-page-title" style={{ display: 'flex', alignItems: 'center', gap: 8, margin: 0 }}>
          <BookOpen size={24} /> 全站起盘明细
        </h1>
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

      <div className="admin-card">
        <div style={{ marginBottom: 16, fontSize: 13, color: 'var(--admin-text-muted)' }}>
          记录平台上每一次八字排盘动作（包含注册用户与游客），共 {total} 条记录。
        </div>

        <form className="admin-search-bar" onSubmit={e => { e.preventDefault(); setPage(1); fetchCharts(1) }}>
          <input className="admin-search-input" value={q} onChange={e => setQ(e.target.value)} placeholder="按邮箱搜索..." />
          <input type="date" className="admin-search-input" value={from} onChange={e => setFrom(e.target.value)} title="排盘起始日期" />
          <input type="date" className="admin-search-input" value={to} onChange={e => setTo(e.target.value)} title="排盘截止日期" />
          <button type="submit" className="admin-btn admin-btn-primary">搜索</button>
          {(q || from || to) && (
            <button type="button" className="admin-btn admin-btn-ghost"
              onClick={() => { setQ(''); setFrom(''); setTo(''); setPage(1); adminChartsAPI.list(1, pageSize).then(res => { setCharts(res.data?.data || []); setTotal(res.data?.total || 0) }) }}>清除</button>
          )}
        </form>

        {loading ? (
          <div className="admin-loading">加载中...</div>
        ) : charts.length === 0 ? (
          <div style={{ textAlign: 'center', color: 'var(--admin-text-faint)', padding: '40px 0' }}>
            <BookOpen size={48} color="var(--admin-border-strong)" style={{ margin: '0 auto 16px' }} />
            <p>暂无起盘记录</p>
          </div>
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
