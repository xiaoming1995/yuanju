import { useEffect, useState } from 'react'
import { adminStatsAPI, adminReportAPI } from '../../lib/adminApi'

interface Stats {
  total_users: number; today_users: number
  total_charts: number; today_charts: number
  total_ai_requests: number; today_ai_requests: number
}

export default function AdminDashboardPage() {
  const [stats, setStats] = useState<Stats | null>(null)
  const [loading, setLoading] = useState(true)
  const [clearing, setClearing] = useState(false)
  const [clearResult, setClearResult] = useState<{ type: 'success' | 'error'; msg: string } | null>(null)

  useEffect(() => {
    adminStatsAPI.overview().then(r => setStats(r.data)).finally(() => setLoading(false))
  }, [])

  const handleClearCache = async () => {
    if (!window.confirm('⚠️ 确定要清空所有 AI 报告缓存吗？\n\n清空后，用户下次访问时将重新生成报告。')) return
    setClearing(true)
    setClearResult(null)
    try {
      const res = await adminReportAPI.clearAll()
      const data = res.data as { deleted: number; message: string }
      setClearResult({ type: 'success', msg: `✅ ${data.message}，共清除 ${data.deleted} 条记录` })
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : '操作失败'
      setClearResult({ type: 'error', msg: `❌ ${msg}` })
    } finally {
      setClearing(false)
    }
  }

  const cards = stats ? [
    { label: '总用户数', value: stats.total_users, sub: `今日 +${stats.today_users}`, icon: '👥' },
    { label: '总命盘数', value: stats.total_charts, sub: `今日 +${stats.today_charts}`, icon: '☯' },
    { label: 'AI 调用总数', value: stats.total_ai_requests, sub: `今日 +${stats.today_ai_requests}`, icon: '🤖' },
  ] : []

  return (
    <div>
      <h1 className="admin-page-title">📊 数据概览</h1>
      {loading ? (
        <div className="admin-loading">加载中...</div>
      ) : (
        <div className="admin-stats-grid">
          {cards.map(c => (
            <div key={c.label} className="admin-stat-card">
              <div style={{ fontSize: 28, marginBottom: 8 }}>{c.icon}</div>
              <div className="admin-stat-value">{c.value}</div>
              <div className="admin-stat-label">{c.label}</div>
              <div className="admin-stat-sub">{c.sub}</div>
            </div>
          ))}
        </div>
      )}

      {/* 系统操作区 */}
      <div className="admin-card" style={{ marginTop: 24 }}>
        <h2 style={{ fontSize: 16, fontWeight: 700, marginBottom: 16, color: '#ccc' }}>🛠️ 系统操作</h2>
        <div style={{ display: 'flex', alignItems: 'center', gap: 16, flexWrap: 'wrap' }}>
          <div>
            <div style={{ fontSize: 14, color: '#aaa', marginBottom: 6 }}>AI 报告缓存</div>
            <div style={{ fontSize: 12, color: '#666', marginBottom: 12 }}>
              清空后用户下次点击"AI 分析"将重新调用 LLM 生成报告
            </div>
            <button
              onClick={handleClearCache}
              disabled={clearing}
              style={{
                padding: '8px 20px',
                background: clearing ? '#333' : '#c0392b',
                color: '#fff',
                border: 'none',
                borderRadius: 8,
                cursor: clearing ? 'not-allowed' : 'pointer',
                fontSize: 14,
                fontWeight: 600,
                transition: 'all 0.2s',
              }}
            >
              {clearing ? '清除中...' : '🗑️ 清空全部报告缓存'}
            </button>
          </div>
          {clearResult && (
            <div style={{
              padding: '10px 16px',
              background: clearResult.type === 'success' ? 'rgba(46,204,113,0.1)' : 'rgba(231,76,60,0.1)',
              border: `1px solid ${clearResult.type === 'success' ? '#2ecc71' : '#e74c3c'}`,
              borderRadius: 8,
              fontSize: 13,
              color: clearResult.type === 'success' ? '#2ecc71' : '#e74c3c',
            }}>
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

  useEffect(() => {
    adminStatsAPI.ai().then(r => setAiStats(r.data))
  }, [])

  if (!aiStats || !aiStats.by_provider?.length) return null

  return (
    <div className="admin-card">
      <h2 style={{ fontSize: 16, fontWeight: 700, marginBottom: 16, color: '#ccc' }}>AI 调用分布</h2>
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
