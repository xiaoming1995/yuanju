import { useState, Fragment } from 'react'
import { BarChart2 } from 'lucide-react'
import { adminTokenUsageAPI } from '../../lib/adminApi'

interface SummaryRow {
  user_id: string
  email: string
  nickname: string
  model: string
  request_count: number
  prompt_tokens: number
  completion_tokens: number
  total_tokens: number
  estimated_cost_cny: number
}

interface DetailRow {
  id: string
  call_type: string
  model: string
  prompt_tokens: number
  completion_tokens: number
  total_tokens: number
  reasoning_tokens: number
  cache_hit_tokens: number
  cache_miss_tokens: number
  estimated_cost_cny: number
  created_at: string
}

interface DetailData {
  total: number
  items: DetailRow[]
}

interface ContentModal {
  id: string
  loading: boolean
  inputContent: string
  outputContent: string
}

function fmt(n: number) {
  return n.toLocaleString('zh-CN')
}

function todayStr() {
  return new Date().toISOString().slice(0, 10)
}

function firstOfMonthStr() {
  const d = new Date()
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-01`
}

interface UserGroup {
  userID: string
  email: string
  nickname: string
  rows: SummaryRow[]
  totalRequestCount: number
  totalPrompt: number
  totalCompletion: number
  totalTokens: number
  totalCost: number
}

function buildGroups(summary: SummaryRow[]): UserGroup[] {
  const order: string[] = []
  const map = new Map<string, UserGroup>()
  for (const r of summary) {
    if (!map.has(r.user_id)) {
      order.push(r.user_id)
      map.set(r.user_id, {
        userID: r.user_id,
        email: r.email,
        nickname: r.nickname,
        rows: [],
        totalRequestCount: 0,
        totalPrompt: 0,
        totalCompletion: 0,
        totalTokens: 0,
        totalCost: 0,
      })
    }
    const g = map.get(r.user_id)!
    g.rows.push(r)
    g.totalRequestCount += r.request_count
    g.totalPrompt += r.prompt_tokens
    g.totalCompletion += r.completion_tokens
    g.totalTokens += r.total_tokens
    g.totalCost += r.estimated_cost_cny
  }
  return order.map(id => map.get(id)!)
}

export default function TokenUsagePage() {
  const [from, setFrom] = useState(firstOfMonthStr())
  const [to, setTo] = useState(todayStr())
  const [summary, setSummary] = useState<SummaryRow[]>([])
  const [loading, setLoading] = useState(false)
  const [queried, setQueried] = useState(false)

  const [drawerUser, setDrawerUser] = useState<SummaryRow | null>(null)
  const [drawerModel, setDrawerModel] = useState('')
  const [detail, setDetail] = useState<DetailData | null>(null)
  const [detailPage, setDetailPage] = useState(1)
  const [detailLoading, setDetailLoading] = useState(false)
  const detailLimit = 20
  const [contentModal, setContentModal] = useState<ContentModal | null>(null)

  const handleQuery = async () => {
    setLoading(true)
    try {
      const res = await adminTokenUsageAPI.summary(from, to)
      setSummary(res.data || [])
      setQueried(true)
    } finally {
      setLoading(false)
    }
  }

  const openDetail = async (row: SummaryRow, page = 1, model = drawerModel) => {
    setDrawerUser(row)
    setDetailPage(page)
    setDetailLoading(true)
    try {
      const res = await adminTokenUsageAPI.detail(row.user_id, from, to, page, detailLimit, model)
      setDetail(res.data)
    } finally {
      setDetailLoading(false)
    }
  }

  const closeDrawer = () => {
    setDrawerUser(null)
    setDetail(null)
    setDrawerModel('')
  }

  const openContent = async (id: string) => {
    setContentModal({ id, loading: true, inputContent: '', outputContent: '' })
    try {
      const res = await adminTokenUsageAPI.content(id)
      setContentModal({ id, loading: false, inputContent: res.data.input_content, outputContent: res.data.output_content })
    } catch {
      setContentModal(prev => prev ? { ...prev, loading: false, inputContent: '加载失败', outputContent: '' } : null)
    }
  }

  const callTypeLabel: Record<string, string> = {
    report: '原局报告',
    report_stream: '流式报告',
    liunian: '流年',
    dayun: '大运',
    celebrity: '名人生成',
    compatibility: '合盘',
  }

  const detailTotalPages = detail ? Math.ceil(detail.total / detailLimit) : 1
  const groups = buildGroups(summary)

  const userModels = drawerUser
    ? summary.filter(r => r.user_id === drawerUser.user_id).map(r => r.model)
    : []

  return (
    <div>
      <h1 className="admin-page-title" style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
        <BarChart2 size={24} /> Token 用量统计
      </h1>

      {/* 筛选栏 */}
      <div className="admin-card" style={{ marginBottom: 24, display: 'flex', gap: 12, alignItems: 'center', flexWrap: 'wrap' }}>
        <label style={{ color: '#aaa', fontSize: 14 }}>开始日期</label>
        <input
          type="date"
          value={from}
          onChange={e => setFrom(e.target.value)}
          style={{ background: '#1a1a2e', color: '#e0e0e0', border: '1px solid #333', borderRadius: 6, padding: '6px 10px' }}
        />
        <label style={{ color: '#aaa', fontSize: 14 }}>结束日期</label>
        <input
          type="date"
          value={to}
          onChange={e => setTo(e.target.value)}
          style={{ background: '#1a1a2e', color: '#e0e0e0', border: '1px solid #333', borderRadius: 6, padding: '6px 10px' }}
        />
        <button
          className="admin-btn"
          onClick={handleQuery}
          disabled={loading}
          style={{ minWidth: 80 }}
        >
          {loading ? '查询中…' : '查询'}
        </button>
      </div>

      {/* 汇总表格 */}
      {queried && (
        <div className="admin-card">
          {summary.length === 0 ? (
            <div style={{ color: '#888', textAlign: 'center', padding: 32 }}>该时间段内无 token 消耗记录</div>
          ) : (
            <>
              <table className="admin-table" style={{ width: '100%' }}>
                <thead>
                  <tr>
                    <th>用户邮箱</th>
                    <th>昵称</th>
                    <th>模型</th>
                    <th style={{ textAlign: 'right' }}>请求次数</th>
                    <th style={{ textAlign: 'right' }}>输入 tokens</th>
                    <th style={{ textAlign: 'right' }}>输出 tokens</th>
                    <th style={{ textAlign: 'right' }}>总 tokens</th>
                    <th style={{ textAlign: 'right' }}>预估费用</th>
                    <th>操作</th>
                  </tr>
                </thead>
                <tbody>
                  {groups.map((group, gi) => (
                    <Fragment key={group.userID}>
                      {/* spacer between groups */}
                      {gi > 0 && (
                        <tr style={{ height: 4, background: '#0d0d1a' }}>
                          <td colSpan={9} style={{ padding: 0 }} />
                        </tr>
                      )}
                      {/* per-model rows */}
                      {group.rows.map((row, ri) => (
                        <tr key={row.model}>
                          <td style={{ color: '#e0e0e0' }}>{ri === 0 ? row.email : ''}</td>
                          <td>{ri === 0 ? (row.nickname || '—') : ''}</td>
                          <td style={{ fontSize: 12, color: '#aaa' }}>{row.model}</td>
                          <td style={{ textAlign: 'right' }}>{fmt(row.request_count)}</td>
                          <td style={{ textAlign: 'right' }}>{fmt(row.prompt_tokens)}</td>
                          <td style={{ textAlign: 'right' }}>{fmt(row.completion_tokens)}</td>
                          <td style={{ textAlign: 'right', color: '#a78bfa' }}>{fmt(row.total_tokens)}</td>
                          <td style={{ textAlign: 'right', color: '#f59e0b', fontSize: 12 }}>
                            ¥ {row.estimated_cost_cny.toFixed(4)}
                          </td>
                          <td />
                        </tr>
                      ))}
                      {/* subtotal row */}
                      <tr style={{ background: '#1e1e40' }}>
                        <td style={{ color: '#888' }} />
                        <td style={{ color: '#888' }} />
                        <td style={{ fontWeight: 700, color: '#e0e0e0', fontSize: 13 }}>合计</td>
                        <td style={{ textAlign: 'right', fontWeight: 700 }}>{fmt(group.totalRequestCount)}</td>
                        <td style={{ textAlign: 'right', fontWeight: 700 }}>{fmt(group.totalPrompt)}</td>
                        <td style={{ textAlign: 'right', fontWeight: 700 }}>{fmt(group.totalCompletion)}</td>
                        <td style={{ textAlign: 'right', fontWeight: 700, color: '#a78bfa' }}>{fmt(group.totalTokens)}</td>
                        <td style={{ textAlign: 'right', fontWeight: 700, color: '#f59e0b' }}>
                          ¥ {group.totalCost.toFixed(4)}
                        </td>
                        <td>
                          <button
                            className="admin-btn"
                            style={{ padding: '4px 12px', fontSize: 13 }}
                            onClick={() => {
                              setDrawerModel('')
                              openDetail(group.rows[0], 1, '')
                            }}
                          >
                            明细
                          </button>
                        </td>
                      </tr>
                    </Fragment>
                  ))}
                </tbody>
              </table>
              <div style={{ fontSize: 12, color: '#555', marginTop: 8, padding: '0 4px' }}>
                * 基于当前 algo_config 单价估算，仅供参考
              </div>
            </>
          )}
        </div>
      )}

      {/* 明细抽屉 */}
      {drawerUser && (
        <div
          style={{
            position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.5)', zIndex: 1000,
            display: 'flex', justifyContent: 'flex-end',
          }}
          onClick={closeDrawer}
        >
          <div
            style={{
              width: 600, maxWidth: '95vw', background: '#12122a', height: '100%',
              overflowY: 'auto', padding: 24, boxShadow: '-4px 0 20px rgba(0,0,0,0.4)',
            }}
            onClick={e => e.stopPropagation()}
          >
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 12 }}>
              <h2 style={{ fontSize: 16, fontWeight: 700, color: '#e0e0e0' }}>
                {drawerUser.email} 的调用明细
              </h2>
              <button className="admin-btn" onClick={closeDrawer} style={{ padding: '4px 12px' }}>关闭</button>
            </div>

            {/* 模型筛选 tabs */}
            <div style={{ display: 'flex', gap: 6, flexWrap: 'wrap', marginBottom: 16 }}>
              {['', ...userModels].map(m => (
                <button
                  key={m === '' ? '__all__' : m}
                  onClick={() => {
                    setDrawerModel(m)
                    openDetail(drawerUser, 1, m)
                  }}
                  style={{
                    background: drawerModel === m ? '#a78bfa' : 'transparent',
                    color: drawerModel === m ? '#fff' : '#888',
                    border: '1px solid #333',
                    borderRadius: 4, padding: '4px 12px', fontSize: 12, cursor: 'pointer',
                  }}
                >
                  {m === '' ? '全部' : m}
                </button>
              ))}
            </div>

            {detailLoading ? (
              <div className="admin-loading">加载中…</div>
            ) : detail && detail.items.length === 0 ? (
              <div style={{ color: '#888', textAlign: 'center', padding: 32 }}>无记录</div>
            ) : detail ? (
              <>
                <table className="admin-table" style={{ width: '100%', marginBottom: 16 }}>
                  <thead>
                    <tr>
                      <th>时间</th>
                      <th>类型</th>
                      <th>模型</th>
                      <th style={{ textAlign: 'right' }}>输入</th>
                      <th style={{ textAlign: 'right' }}>输出</th>
                      <th style={{ textAlign: 'right' }}>推理</th>
                      <th style={{ textAlign: 'right' }}>缓存命中</th>
                      <th style={{ textAlign: 'right' }}>费用</th>
                      <th style={{ textAlign: 'right' }}>总计</th>
                      <th></th>
                    </tr>
                  </thead>
                  <tbody>
                    {detail.items.map(item => (
                      <tr key={item.id}>
                        <td style={{ fontSize: 12, color: '#aaa' }}>
                          {new Date(item.created_at).toLocaleString('zh-CN', { hour12: false })}
                        </td>
                        <td>{callTypeLabel[item.call_type] ?? item.call_type}</td>
                        <td style={{ fontSize: 12, color: '#aaa' }}>{item.model}</td>
                        <td style={{ textAlign: 'right' }}>{fmt(item.prompt_tokens)}</td>
                        <td style={{ textAlign: 'right' }}>{fmt(item.completion_tokens)}</td>
                        <td style={{ textAlign: 'right', color: item.reasoning_tokens > 0 ? '#94a3b8' : '#555' }}>
                          {item.reasoning_tokens > 0 ? fmt(item.reasoning_tokens) : '—'}
                        </td>
                        <td style={{ textAlign: 'right', color: item.cache_hit_tokens > 0 ? '#4ade80' : '#555' }}>
                          {item.cache_hit_tokens > 0 ? fmt(item.cache_hit_tokens) : '—'}
                        </td>
                        <td style={{ textAlign: 'right', color: '#f59e0b', fontSize: 12 }}>
                          ¥ {item.estimated_cost_cny.toFixed(4)}
                        </td>
                        <td style={{ textAlign: 'right', fontWeight: 700, color: '#a78bfa' }}>{fmt(item.total_tokens)}</td>
                        <td>
                          <button
                            className="admin-btn"
                            style={{ padding: '2px 10px', fontSize: 12 }}
                            onClick={() => openContent(item.id)}
                          >
                            查看
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>

                {/* 分页 */}
                {detailTotalPages > 1 && (
                  <div style={{ display: 'flex', gap: 8, justifyContent: 'center' }}>
                    <button
                      className="admin-btn"
                      disabled={detailPage <= 1}
                      onClick={() => openDetail(drawerUser, detailPage - 1)}
                      style={{ padding: '4px 12px' }}
                    >
                      上一页
                    </button>
                    <span style={{ color: '#888', lineHeight: '32px', fontSize: 13 }}>
                      {detailPage} / {detailTotalPages}（共 {detail.total} 条）
                    </span>
                    <button
                      className="admin-btn"
                      disabled={detailPage >= detailTotalPages}
                      onClick={() => openDetail(drawerUser, detailPage + 1)}
                      style={{ padding: '4px 12px' }}
                    >
                      下一页
                    </button>
                  </div>
                )}
              </>
            ) : null}
          </div>
        </div>
      )}

      {/* 内容查看 Modal */}
      {contentModal && (
        <div
          onClick={() => setContentModal(null)}
          style={{
            position: 'fixed', inset: 0,
            background: 'rgba(0,0,0,0.7)',
            zIndex: 2000,
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            padding: 24,
          }}
        >
          <div
            onClick={e => e.stopPropagation()}
            style={{
              background: '#1a1f2e',
              border: '1px solid rgba(255,255,255,0.1)',
              borderRadius: 12,
              width: '90vw', maxWidth: 1100,
              maxHeight: '85vh',
              display: 'flex', flexDirection: 'column',
              overflow: 'hidden',
            }}
          >
            <div style={{ padding: '16px 20px', borderBottom: '1px solid rgba(255,255,255,0.08)', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <span style={{ fontWeight: 700, color: '#e8e4d8' }}>调用内容详情</span>
              <button className="admin-btn" onClick={() => setContentModal(null)}>关闭</button>
            </div>
            {contentModal.loading ? (
              <div className="admin-loading" style={{ padding: 40 }}>加载中…</div>
            ) : (
              <div style={{ display: 'flex', flex: 1, overflow: 'hidden', gap: 1, background: 'rgba(255,255,255,0.05)' }}>
                <div style={{ flex: 1, display: 'flex', flexDirection: 'column', overflow: 'hidden', background: '#1a1f2e' }}>
                  <div style={{ padding: '10px 16px', fontSize: 12, color: '#9e9a8a', borderBottom: '1px solid rgba(255,255,255,0.06)', fontWeight: 600 }}>
                    输入（Prompt）
                  </div>
                  <pre style={{
                    flex: 1, overflowY: 'auto', margin: 0,
                    padding: '12px 16px',
                    fontSize: 12, color: '#c8c0b0',
                    lineHeight: 1.7, whiteSpace: 'pre-wrap', wordBreak: 'break-word',
                    fontFamily: 'monospace',
                  }}>
                    {contentModal.inputContent || '（无内容）'}
                  </pre>
                </div>
                <div style={{ flex: 1, display: 'flex', flexDirection: 'column', overflow: 'hidden', background: '#1a1f2e' }}>
                  <div style={{ padding: '10px 16px', fontSize: 12, color: '#9e9a8a', borderBottom: '1px solid rgba(255,255,255,0.06)', fontWeight: 600 }}>
                    输出（Response）
                  </div>
                  <pre style={{
                    flex: 1, overflowY: 'auto', margin: 0,
                    padding: '12px 16px',
                    fontSize: 12, color: '#c8c0b0',
                    lineHeight: 1.7, whiteSpace: 'pre-wrap', wordBreak: 'break-word',
                    fontFamily: 'monospace',
                  }}>
                    {contentModal.outputContent || '（无内容）'}
                  </pre>
                </div>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
