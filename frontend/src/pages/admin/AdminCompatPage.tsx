import React, { useEffect, useState } from 'react'
import { Heart, User, RefreshCw } from 'lucide-react'
import { adminCompatAPI } from '../../lib/adminApi'

interface CompatListItem {
  id: string
  user_email?: string | null
  self_name: string
  partner_name: string
  overall_score: number
  overall_level: string
  relationship_stage: string
  primary_question: string
  analysis_version: string
  created_at: string
}

interface CompatDetail {
  reading: {
    overall_score: number
    overall_level: string
    relationship_stage: string
    primary_question: string
    analysis_version: string
    dimension_scores: Record<string, unknown>
    duration_assessment: { summary?: string; overall_band?: string }
    consulting_assessment: Record<string, unknown>
    summary_tags: string[]
  }
  participants: Array<{
    role: string
    display_name: string
    birth_profile: { year: number; month: number; day: number; hour: number; gender: string }
    chart_snapshot?: {
      year_gan?: string; year_zhi?: string; month_gan?: string; month_zhi?: string
      day_gan?: string; day_zhi?: string; hour_gan?: string; hour_zhi?: string
    } | null
  }>
  evidences: Array<{ id: string; dimension: string; polarity: string; title: string; detail: string }>
  latest_report?: { content: string; model: string; created_at: string } | null
}

const levelLabel = (l: string) => l === 'high' ? '高' : l === 'low' ? '低' : '中'
const stageLabel: Record<string, string> = {
  ambiguous: '暧昧', dating: '热恋', long_distance: '异地', reconciliation: '复合',
  marriage_or_engagement: '婚姻/订婚', crush: '单恋',
}

export default function AdminCompatPage() {
  const [items, setItems] = useState<CompatListItem[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(true)
  const [expandedId, setExpandedId] = useState<string | null>(null)
  const [details, setDetails] = useState<Record<string, CompatDetail>>({})
  const [detailLoading, setDetailLoading] = useState<Record<string, boolean>>({})
  const pageSize = 20

  const fetchList = async (pageNum: number) => {
    try {
      setLoading(true)
      const res = await adminCompatAPI.list(pageNum, pageSize)
      setItems(res.data?.data || [])
      setTotal(res.data?.total || 0)
    } catch (err) {
      console.error('获取合盘明细失败:', err)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { fetchList(page) }, [page])

  useEffect(() => {
    if (expandedId && !details[expandedId]) {
      setDetailLoading(prev => ({ ...prev, [expandedId]: true }))
      adminCompatAPI.detail(expandedId)
        .then(res => setDetails(prev => ({ ...prev, [expandedId]: res.data?.data })))
        .catch(err => console.error(err))
        .finally(() => setDetailLoading(prev => ({ ...prev, [expandedId]: false })))
    }
  }, [expandedId]) // eslint-disable-line react-hooks/exhaustive-deps

  const totalPages = Math.ceil((total || 0) / pageSize) || 1

  const pillars = (s?: CompatDetail['participants'][0]['chart_snapshot']) => [
    { label: '年', gan: s?.year_gan, zhi: s?.year_zhi },
    { label: '月', gan: s?.month_gan, zhi: s?.month_zhi },
    { label: '日', gan: s?.day_gan, zhi: s?.day_zhi },
    { label: '时', gan: s?.hour_gan, zhi: s?.hour_zhi },
  ]

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <h1 className="admin-page-title" style={{ display: 'flex', alignItems: 'center', gap: 8, margin: 0 }}>
          <Heart size={24} /> 全站合盘明细
        </h1>
        <button onClick={() => fetchList(page)} style={{ display: 'flex', alignItems: 'center', gap: 6, padding: '6px 14px', borderRadius: 8, border: '1px solid #444', background: '#2a2a3a', color: '#ccc', cursor: 'pointer', fontSize: 13 }}>
          <RefreshCw size={14} /> 刷新
        </button>
      </div>

      <div className="admin-card">
        <div style={{ marginBottom: 16, fontSize: 13, color: '#888' }}>
          记录平台上每一次合盘测算，共 {total} 条记录。
        </div>

        {loading ? (
          <div className="admin-loading">加载中...</div>
        ) : items.length === 0 ? (
          <div style={{ textAlign: 'center', color: '#555', padding: '40px 0' }}>
            <Heart size={48} color="#333" style={{ margin: '0 auto 16px' }} />
            <p>暂无合盘记录</p>
          </div>
        ) : (
          <table className="admin-table">
            <thead>
              <tr><th>排盘用户</th><th>排盘时间</th><th>双方命主</th><th>总分/等级</th><th>关系阶段</th><th>操作</th></tr>
            </thead>
            <tbody>
              {items.map(it => (
                <React.Fragment key={it.id}>
                  <tr style={{ background: expandedId === it.id ? 'rgba(255,255,255,0.02)' : 'transparent' }}>
                    <td>
                      <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                        <User size={14} color="#888" />
                        {it.user_email
                          ? <span style={{ color: '#ccc' }}>{it.user_email}</span>
                          : <span style={{ color: '#666', fontStyle: 'italic' }}>未知</span>}
                      </div>
                    </td>
                    <td style={{ fontSize: 12, color: '#888' }}>{new Date(it.created_at).toLocaleString('zh-CN')}</td>
                    <td style={{ color: '#ccc' }}>{it.self_name || '—'} × {it.partner_name || '—'}</td>
                    <td>
                      <span style={{ color: '#f472b6', fontWeight: 600 }}>{it.overall_score}</span>
                      <span style={{ color: '#888', fontSize: 12 }}> / {levelLabel(it.overall_level)}</span>
                    </td>
                    <td style={{ fontSize: 12, color: '#aaa' }}>{stageLabel[it.relationship_stage] || it.relationship_stage}</td>
                    <td>
                      <button onClick={() => setExpandedId(expandedId === it.id ? null : it.id)} style={{ padding: '4px 12px', fontSize: 12, background: expandedId === it.id ? '#a78bfa' : '#2a2a3a', border: expandedId === it.id ? 'none' : '1px solid #444', borderRadius: 6, color: expandedId === it.id ? '#fff' : '#ccc', cursor: 'pointer' }}>
                        {expandedId === it.id ? '收起详情' : '查看详情'}
                      </button>
                    </td>
                  </tr>

                  {expandedId === it.id && (
                    <tr style={{ background: 'rgba(0,0,0,0.2)' }}>
                      <td colSpan={6} style={{ padding: '20px 24px', borderLeft: '3px solid #f472b6' }}>
                        {detailLoading[it.id] && <div style={{ fontSize: 12, color: '#666' }}>加载中...</div>}
                        {!detailLoading[it.id] && details[it.id] && (() => {
                          const d = details[it.id]
                          return (
                            <div>
                              {/* 双方四柱 */}
                              <div style={{ display: 'flex', gap: 40, flexWrap: 'wrap', marginBottom: 20 }}>
                                {d.participants.map((p, pi) => (
                                  <div key={pi}>
                                    <div style={{ fontSize: 13, color: '#888', marginBottom: 12 }}>
                                      {p.role === 'self' ? '命主' : '对方'}：{p.display_name || '—'}（{p.birth_profile.gender === 'male' ? '男' : '女'} {p.birth_profile.year}-{p.birth_profile.month}-{p.birth_profile.day}）
                                    </div>
                                    <div style={{ display: 'flex', gap: 12 }}>
                                      {pillars(p.chart_snapshot).map((col, ci) => (
                                        <div key={ci} style={{ textAlign: 'center', background: '#222', padding: '8px 14px', borderRadius: 8, border: '1px solid #333' }}>
                                          <div style={{ fontSize: 11, color: '#666', marginBottom: 8 }}>{col.label}柱</div>
                                          <div style={{ fontSize: 16, fontWeight: 600, color: '#ccc' }}>{col.gan || '—'}</div>
                                          <div style={{ fontSize: 16, fontWeight: 600, color: '#ccc' }}>{col.zhi || ''}</div>
                                        </div>
                                      ))}
                                    </div>
                                  </div>
                                ))}
                              </div>

                              {/* 期限评估 */}
                              {d.reading.duration_assessment?.summary && (
                                <div style={{ background: '#222', padding: '12px 16px', borderRadius: 8, border: '1px solid #333', marginBottom: 16 }}>
                                  <div style={{ fontSize: 13, color: '#888', marginBottom: 6 }}>期限评估</div>
                                  <div style={{ fontSize: 13, color: '#ccc', lineHeight: 1.7 }}>{d.reading.duration_assessment.summary}</div>
                                </div>
                              )}

                              {/* 证据列表 */}
                              <div style={{ fontSize: 13, color: '#888', marginBottom: 8 }}>证据 ({d.evidences.length} 条)</div>
                              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 8, marginBottom: 16 }}>
                                {d.evidences.map(ev => (
                                  <div key={ev.id} style={{ background: 'rgba(0,0,0,0.2)', padding: '8px 12px', borderRadius: 6 }}>
                                    <div style={{ fontSize: 11, color: ev.polarity === 'positive' ? '#34d399' : ev.polarity === 'negative' ? '#ff6b6b' : '#aaa', marginBottom: 4, fontWeight: 600 }}>
                                      [{ev.dimension}] {ev.title}
                                    </div>
                                    <div style={{ fontSize: 12, color: '#ccc', lineHeight: 1.6 }}>{ev.detail}</div>
                                  </div>
                                ))}
                              </div>

                              {/* AI 报告 */}
                              <div style={{ fontSize: 13, color: '#888', marginBottom: 8 }}>AI 报告</div>
                              {d.latest_report
                                ? <div style={{ background: 'rgba(244,114,182,0.05)', padding: 16, borderRadius: 8, border: '1px solid rgba(244,114,182,0.2)', fontSize: 12, color: '#ccc', lineHeight: 1.8, whiteSpace: 'pre-wrap', maxHeight: 300, overflowY: 'auto' }}>{d.latest_report.content}</div>
                                : <div style={{ background: '#222', padding: 16, borderRadius: 8, border: '1px dashed #444', color: '#666', fontSize: 13 }}>此合盘尚未生成 AI 报告。</div>}
                            </div>
                          )
                        })()}
                      </td>
                    </tr>
                  )}
                </React.Fragment>
              ))}
            </tbody>
          </table>
        )}

        {totalPages > 1 && (
          <div style={{ display: 'flex', justifyContent: 'center', gap: 8, marginTop: 20 }}>
            <button disabled={page === 1} onClick={() => setPage(p => p - 1)} style={{ padding: '6px 14px', borderRadius: 8, border: 'none', background: page === 1 ? '#1a1a2e' : '#2a2a3a', color: page === 1 ? '#555' : '#ccc', cursor: page === 1 ? 'not-allowed' : 'pointer' }}>上一页</button>
            <span style={{ lineHeight: '32px', fontSize: 13, color: '#666', margin: '0 8px' }}>第 {page} / {totalPages} 页</span>
            <button disabled={page >= totalPages} onClick={() => setPage(p => p + 1)} style={{ padding: '6px 14px', borderRadius: 8, border: 'none', background: page >= totalPages ? '#1a1a2e' : '#2a2a3a', color: page >= totalPages ? '#555' : '#ccc', cursor: page >= totalPages ? 'not-allowed' : 'pointer' }}>下一页</button>
          </div>
        )}
      </div>
    </div>
  )
}
