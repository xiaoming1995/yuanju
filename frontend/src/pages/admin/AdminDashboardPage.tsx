import { useEffect, useState } from 'react'
import { Users, Compass, Bot, Wrench, Trash2, LayoutDashboard } from 'lucide-react'
import { adminStatsAPI, adminReportAPI } from '../../lib/adminApi'
import { adminActionErrorMessage } from '../../lib/adminApi'
import { AdminRequestState } from '../../components/AdminRequestState'

interface Stats {
  total_users: number; today_users: number
  total_charts: number; today_charts: number
  total_ai_requests: number; today_ai_requests: number
}

export default function AdminDashboardPage() {
  const [stats, setStats] = useState<Stats | null>(null)
  const [loading, setLoading] = useState(true)
  const [loadError, setLoadError] = useState('')
  const [revision, setRevision] = useState(0)
  const [clearing, setClearing] = useState(false)
  const [clearResult, setClearResult] = useState<{ type: 'success' | 'error'; msg: string } | null>(null)

  useEffect(() => {
    let active = true
    adminStatsAPI.overview().then(r => { if (active) setStats(r.data) })
      .catch((e: unknown) => { if (active) setLoadError(e instanceof Error ? e.message : '加载失败') })
      .finally(() => { if (active) setLoading(false) })
    return () => { active = false }
  }, [revision])

  const handleClearCache = async () => {
    if (!window.confirm('确定要清空所有 AI 报告缓存吗？\n\n清空后，用户下次访问时将重新生成报告。')) return
    setClearing(true)
    setClearResult(null)
    try {
      const res = await adminReportAPI.clearAll()
      const data = res.data as { deleted: number; message: string }
      setClearResult({ type: 'success', msg: `${data.message}，共清除 ${data.deleted} 条记录` })
    } catch (e: unknown) {
      const msg = adminActionErrorMessage(e)
      setClearResult({ type: 'error', msg })
    } finally {
      setClearing(false)
    }
  }

  const cards = stats ? [
    { label: '总用户数', value: stats.total_users, sub: `今日 +${stats.today_users}`, icon: <Users size={18} /> },
    { label: '总起盘数', value: stats.total_charts, sub: `今日 +${stats.today_charts}`, icon: <Compass size={18} /> },
    { label: 'AI 调用总数', value: stats.total_ai_requests, sub: `今日 +${stats.today_ai_requests}`, icon: <Bot size={18} /> },
  ] : []

  return (
    <div>
      <div className="admin-dashboard-heading">
        <h1 className="admin-page-title"><LayoutDashboard size={24} /> 数据概览</h1>
        <p>查看平台运行情况与近期使用数据</p>
      </div>
      {loading && !stats ? (
        <AdminRequestState state="loading" message="正在加载数据概览…" />
      ) : loadError ? (
        <AdminRequestState state="error" message={loadError} onRetry={() => { setLoading(true); setLoadError(''); setRevision(n => n + 1) }} />
      ) : (
        <div className="admin-stats-grid">
          {cards.map(c => (
            <div key={c.label} className="admin-stat-card">
              <div className="admin-stat-top"><div className="admin-stat-label">{c.label}</div><span className="admin-stat-icon">{c.icon}</span></div>
              <div className="admin-stat-value">{c.value.toLocaleString('zh-CN')}</div>
              <div className="admin-stat-sub">{c.sub}</div>
            </div>
          ))}
        </div>
      )}

      {/* 系统操作区 */}
      <div className="admin-card admin-system-card" style={{ marginTop: 24 }}>
        <h2 style={{ fontSize: 16, fontWeight: 700, marginBottom: 16, color: 'var(--admin-text-secondary)', display: 'flex', alignItems: 'center', gap: 6 }}>
          <Wrench size={18} /> 系统操作
        </h2>
        <div style={{ display: 'flex', alignItems: 'center', gap: 16, flexWrap: 'wrap' }}>
          <div>
            <div style={{ fontSize: 14, color: 'var(--admin-text-secondary)', marginBottom: 6 }}>AI 报告缓存</div>
            <div style={{ fontSize: 12, color: 'var(--admin-text-faint)', marginBottom: 12 }}>
              清空后用户下次点击"AI 分析"将重新调用 LLM 生成报告
            </div>
            <button
              onClick={handleClearCache}
              disabled={clearing}
              style={{
                padding: '8px 20px',
                background: clearing ? 'var(--admin-bg-muted)' : 'var(--admin-danger)',
                color: '#fffdf8',
                border: 'none',
                borderRadius: 8,
                cursor: clearing ? 'not-allowed' : 'pointer',
                fontSize: 14,
                fontWeight: 600,
                transition: 'all 0.2s',
              }}
            >
              <div style={{ display: 'flex', alignItems: 'center', gap: 6, justifyContent: 'center' }}>
                {clearing ? <Trash2 size={16} /> : <Trash2 size={16} />}
                {clearing ? '清除中...' : '清空全部报告缓存'}
              </div>
            </button>
          </div>
          {clearResult && (
            <div style={{
              padding: '10px 16px',
              background: clearResult.type === 'success' ? 'rgba(var(--admin-success-rgb), 0.1)' : 'rgba(var(--admin-danger-rgb), 0.1)',
              border: `1px solid ${clearResult.type === 'success' ? 'var(--admin-success)' : 'var(--admin-danger)'}`,
              borderRadius: 8,
              fontSize: 13,
              color: clearResult.type === 'success' ? 'var(--admin-success)' : 'var(--admin-danger)',
              display: 'flex',
              alignItems: 'center',
              gap: 6
            }}>
              <span className={`status-dot ${clearResult.type}`}></span>
              {clearResult.msg}
            </div>
          )}
        </div>
      </div>

      <AIStatsSection />
    </div>
  )
}

function AIStatsSection() {
  const [aiStats, setAiStats] = useState<{ by_provider: Array<{ provider: string; total: number; success_rate: number }> } | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [revision, setRevision] = useState(0)

  useEffect(() => {
    let active = true
    adminStatsAPI.ai().then(r => { if (active) setAiStats(r.data) })
      .catch((e: unknown) => { if (active) setError(e instanceof Error ? e.message : '加载失败') })
      .finally(() => { if (active) setLoading(false) })
    return () => { active = false }
  }, [revision])

  if (loading) return <AdminRequestState state="loading" message="正在加载 AI 调用分布…" />
  if (error) return <AdminRequestState state="error" message={error} onRetry={() => { setLoading(true); setError(''); setRevision(n => n + 1) }} />
  if (!aiStats?.by_provider?.length) return <AdminRequestState state="empty" message="暂无 AI 调用分布数据" />

  return (
    <div className="admin-card">
      <h2 style={{ fontSize: 16, fontWeight: 700, marginBottom: 16, color: 'var(--admin-text-secondary)' }}>AI 调用分布</h2>
      <table className="admin-table">
        <thead><tr><th>Provider</th><th>调用次数</th><th>成功率</th></tr></thead>
        <tbody>
          {aiStats.by_provider.map(p => (
            <tr key={p.provider}>
              <td>{p.provider || '未知'}</td>
              <td>{p.total}</td>
              <td>{p.success_rate.toFixed(1)}%</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
