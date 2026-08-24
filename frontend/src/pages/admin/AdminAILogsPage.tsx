import { useEffect, useState } from 'react'
import { Activity, CheckCircle, Clock, TrendingUp, List, XCircle } from 'lucide-react'
import { adminAILogsAPI } from '../../lib/adminApi'

interface AILog {
  id: string
  chart_id?: string
  provider_id?: string
  provider_name?: string
  model: string
  duration_ms: number
  status: 'success' | 'error'
  error_msg?: string
  created_at: string
}

interface DayStat {
  date: string
  total: number
  success_count: number
  error_count: number
  avg_duration_ms: number
}

export default function AdminAILogsPage() {
  const [logs, setLogs] = useState<AILog[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [statusFilter, setStatusFilter] = useState('')
  const [loading, setLoading] = useState(true)
  const [summary, setSummary] = useState<DayStat[]>([])
  const [expandedId, setExpandedId] = useState<string | null>(null)

  const pageSize = 20

  useEffect(() => {
    adminAILogsAPI.list(page, statusFilter)
      .then(r => {
        setLogs(r.data.logs || [])
        setTotal(r.data.total || 0)
      })
      .finally(() => setLoading(false))
  }, [page, statusFilter])

  useEffect(() => {
    adminAILogsAPI.summary().then(r => setSummary(r.data.summary || []))
  }, [])

  const totalAll = summary.reduce((s, d) => s + d.total, 0)
  const totalSuccess = summary.reduce((s, d) => s + d.success_count, 0)
  const successRate = totalAll > 0 ? ((totalSuccess / totalAll) * 100).toFixed(1) : '–'
  const avgDuration = summary.length > 0
    ? (summary.reduce((s, d) => s + d.avg_duration_ms, 0) / summary.length).toFixed(0)
    : '–'

  const totalPages = Math.ceil(total / pageSize)

  const handleFilter = (f: string) => {
    setStatusFilter(f)
    setPage(1)
  }

  return (
    <div>
      <h1 className="admin-page-title" style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
        <List size={24} /> AI 调用日志
      </h1>

      {/* 统计卡片 */}
      <div className="admin-stats-grid" style={{ marginBottom: 24 }}>
        <div className="admin-stat-card">
          <div style={{ marginBottom: 12, display: 'flex' }}><Activity size={28} color="var(--admin-accent)" /></div>
          <div className="admin-stat-value">{totalAll}</div>
          <div className="admin-stat-label">近7天总调用</div>
        </div>
        <div className="admin-stat-card">
          <div style={{ marginBottom: 12, display: 'flex' }}><CheckCircle size={28} color="var(--admin-success)" /></div>
          <div className="admin-stat-value">{successRate}%</div>
          <div className="admin-stat-label">近7天成功率</div>
        </div>
        <div className="admin-stat-card">
          <div style={{ marginBottom: 12, display: 'flex' }}><Clock size={28} color="var(--admin-warning)" /></div>
          <div className="admin-stat-value">{avgDuration}ms</div>
          <div className="admin-stat-label">日均响应耗时</div>
        </div>
      </div>

      {/* 近7天趋势 */}
      {summary.length > 0 && (
        <div className="admin-card" style={{ marginBottom: 24 }}>
          <h2 style={{ fontSize: 16, fontWeight: 700, marginBottom: 16, color: 'var(--admin-text-secondary)', display: 'flex', alignItems: 'center', gap: 6 }}>
            <TrendingUp size={18} /> 近7天趋势
          </h2>
          <div style={{ display: 'flex', gap: 8, alignItems: 'flex-end', height: 80 }}>
            {summary.map(d => {
              const maxTotal = Math.max(...summary.map(s => s.total), 1)
              const height = Math.max((d.total / maxTotal) * 60, d.total > 0 ? 6 : 2)
              return (
                <div key={d.date} style={{ flex: 1, textAlign: 'center' }}>
                  <div style={{ fontSize: 10, color: 'var(--admin-text-faint)', marginBottom: 4 }}>{d.total}</div>
                  <div style={{
                    height,
                    background: d.error_count > 0
                      ? `linear-gradient(to top, var(--admin-danger) ${Math.round((d.error_count / Math.max(d.total, 1)) * 100)}%, var(--admin-success) 0%)`
                      : 'var(--admin-success)',
                    borderRadius: 3,
                    minHeight: 2,
                  }} />
                  <div style={{ fontSize: 9, color: 'var(--admin-text-faint)', marginTop: 4 }}>
                    {d.date.slice(5)}
                  </div>
                </div>
              )
            })}
          </div>
          <div style={{ display: 'flex', gap: 16, marginTop: 8, fontSize: 11, color: 'var(--admin-text-faint)' }}>
            <span style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
              <span style={{ width: 10, height: 10, background: 'var(--admin-success)', borderRadius: 2, display: 'inline-block' }} />
              成功
            </span>
            <span style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
              <span style={{ width: 10, height: 10, background: 'var(--admin-danger)', borderRadius: 2, display: 'inline-block' }} />
              失败
            </span>
          </div>
        </div>
      )}

      {/* 筛选栏 */}
      <div className="admin-card">
        <div style={{ display: 'flex', gap: 8, marginBottom: 16 }}>
          {[
            { val: '', label: '全部', icon: null },
            { val: 'success', label: '成功', icon: <CheckCircle size={14} /> },
            { val: 'error', label: '失败', icon: <XCircle size={14} /> }
          ].map(({ val, label, icon }) => (
            <button
              key={val}
              onClick={() => handleFilter(val)}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 6,
                padding: '6px 16px',
                borderRadius: 20,
                border: 'none',
                cursor: 'pointer',
                fontSize: 13,
                fontWeight: 600,
                background: statusFilter === val ? 'var(--admin-accent)' : 'var(--admin-bg-soft)',
                color: statusFilter === val ? '#fffdf8' : 'var(--admin-text-muted)',
                transition: 'all 0.2s',
              }}
            >
              {icon} {label}
            </button>
          ))}
          <span style={{ marginLeft: 'auto', fontSize: 13, color: 'var(--admin-text-faint)', lineHeight: '32px' }}>
            共 {total} 条记录
          </span>
        </div>

        {/* 日志表格 */}
        {loading ? (
          <div className="admin-loading">加载中...</div>
        ) : logs.length === 0 ? (
          <div style={{ textAlign: 'center', color: 'var(--admin-text-faint)', padding: '40px 0' }}>暂无记录</div>
        ) : (
          <table className="admin-table">
            <thead>
              <tr>
                <th>时间</th>
                <th>Provider</th>
                <th>模型</th>
                <th>耗时</th>
                <th>状态</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              {logs.map(log => (
                <>
                  <tr key={log.id} style={{ cursor: log.error_msg ? 'pointer' : 'default' }}>
                    <td style={{ fontSize: 12, color: 'var(--admin-text-muted)' }}>
                      {new Date(log.created_at).toLocaleString('zh-CN')}
                    </td>
                    <td>{log.provider_name || '–'}</td>
                    <td style={{ fontSize: 12, color: 'var(--admin-text-secondary)' }}>{log.model || '–'}</td>
                    <td style={{ fontFamily: 'monospace' }}>{log.duration_ms.toLocaleString()}ms</td>
                    <td>
                      <span style={{
                        display: 'inline-block',
                        padding: '2px 10px',
                        borderRadius: 12,
                        fontSize: 12,
                        fontWeight: 600,
                        background: log.status === 'success' ? 'rgba(var(--admin-success-rgb), 0.15)' : 'rgba(var(--admin-danger-rgb), 0.15)',
                        color: log.status === 'success' ? 'var(--admin-success)' : 'var(--admin-danger)',
                      }}>
                        {log.status === 'success' ? '成功' : '失败'}
                      </span>
                    </td>
                    <td>
                      {log.error_msg && (
                        <button
                          onClick={() => setExpandedId(expandedId === log.id ? null : log.id)}
                          style={{
                            padding: '3px 10px',
                            fontSize: 11,
                            background: 'var(--admin-bg-soft)',
                            border: '1px solid var(--admin-border-strong)',
                            borderRadius: 6,
                            color: 'var(--admin-text-secondary)',
                            cursor: 'pointer',
                          }}
                        >
                          {expandedId === log.id ? '收起' : '查看错误'}
                        </button>
                      )}
                    </td>
                  </tr>
                  {expandedId === log.id && log.error_msg && (
                    <tr key={`${log.id}-detail`}>
                      <td colSpan={6} style={{
                        background: 'rgba(var(--admin-danger-rgb), 0.05)',
                        padding: '12px 16px',
                        borderLeft: '3px solid var(--admin-danger)',
                      }}>
                        <div style={{ fontSize: 12, color: 'var(--admin-danger)', fontFamily: 'monospace', whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>
                          {log.error_msg}
                        </div>
                      </td>
                    </tr>
                  )}
                </>
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
                color: page === 1 ? 'var(--admin-border-strong)' : 'var(--admin-text-secondary)', cursor: page === 1 ? 'default' : 'pointer',
              }}
            >← 上一页</button>
            <span style={{ lineHeight: '34px', fontSize: 13, color: 'var(--admin-text-faint)' }}>
              {page} / {totalPages}
            </span>
            <button
              disabled={page === totalPages}
              onClick={() => setPage(p => p + 1)}
              style={{
                padding: '6px 14px', borderRadius: 8, border: 'none',
                background: page === totalPages ? 'var(--admin-bg-muted)' : 'var(--admin-bg-soft)',
                color: page === totalPages ? 'var(--admin-border-strong)' : 'var(--admin-text-secondary)', cursor: page === totalPages ? 'default' : 'pointer',
              }}
            >下一页 →</button>
          </div>
        )}
      </div>
    </div>
  )
}
