import React, { useEffect, useState } from 'react'
import { BookOpen, User, Calendar, RefreshCw, Trash2 } from 'lucide-react'
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
    personality?: string
    career?: string
    romance?: string
    health?: string
  }
  created_at: string
}

interface AdminLiunianReport {
  id: string
  chart_id: string
  target_year: number
  dayun_ganzhi: string
  content_structured: any
  model: string
  created_at: string
}

const genderLabel = (g: string) => g === 'male' ? '男' : '女'

export default function AdminChartsPage() {
  const [charts, setCharts] = useState<ChartRecord[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(true)
  const [expandedId, setExpandedId] = useState<string | null>(null)
  const [liunianReports, setLiunianReports] = useState<Record<string, AdminLiunianReport[]>>({})
  const [liunianLoading, setLiunianLoading] = useState<Record<string, boolean>>({})
  const pageSize = 20

  const fetchCharts = async (pageNum: number) => {
    try {
      setLoading(true)
      const res = await adminChartsAPI.list(pageNum, pageSize)
      setCharts(res.data?.data || [])
      setTotal(res.data?.total || 0)
    } catch (err: any) {
      console.error('获取起盘流水失败:', err)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchCharts(page)
  }, [page])

  useEffect(() => {
    if (expandedId && !liunianReports[expandedId]) {
      setLiunianLoading(prev => ({ ...prev, [expandedId]: true }))
      adminChartsAPI.getLiunianReports(expandedId)
        .then(res => setLiunianReports(prev => ({ ...prev, [expandedId]: res.data?.data || [] })))
        .catch(err => console.error(err))
        .finally(() => setLiunianLoading(prev => ({ ...prev, [expandedId]: false })))
    }
  }, [expandedId]) // eslint-disable-line react-hooks/exhaustive-deps

  const totalPages = Math.ceil((total || 0) / pageSize) || 1

  const handleDeleteLiunian = async (id: string, chartId: string) => {
    if (!window.confirm('确定要清除该年份的流年预测缓存吗？清除后前端用户将可以重新生成该年的运势。')) return;
    try {
      await adminChartsAPI.deleteLiunianReport(id);
      // Remove it from the local state
      setLiunianReports(prev => ({
        ...prev,
        [chartId]: (prev[chartId] || []).filter(r => r.id !== id)
      }));
    } catch (err: any) {
      alert('清除失败: ' + (err.response?.data?.error || err.message));
    }
  }

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
            padding: '6px 14px', borderRadius: 8, border: '1px solid #444',
            background: '#2a2a3a', color: '#ccc', cursor: 'pointer', fontSize: 13
          }}
        >
          <RefreshCw size={14} /> 刷新
        </button>
      </div>

      <div className="admin-card">
        <div style={{ marginBottom: 16, fontSize: 13, color: '#888' }}>
          记录平台上每一次八字排盘动作（包含注册用户与游客），共 {total} 条记录。
        </div>

        {loading ? (
          <div className="admin-loading">加载中...</div>
        ) : charts.length === 0 ? (
          <div style={{ textAlign: 'center', color: '#555', padding: '40px 0' }}>
            <BookOpen size={48} color="#333" style={{ margin: '0 auto 16px' }} />
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
                <React.Fragment key={chart.id}>
                  <tr style={{ background: expandedId === chart.id ? 'rgba(255,255,255,0.02)' : 'transparent' }}>
                    <td>
                      <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                        <User size={14} color="#888" />
                        {chart.user_email ? (
                          <span style={{ color: '#ccc' }}>{chart.user_email}</span>
                        ) : (
                          <span style={{ color: '#666', fontStyle: 'italic' }}>游客</span>
                        )}
                      </div>
                    </td>
                    <td style={{ fontSize: 12, color: '#888' }}>
                      {new Date(chart.created_at).toLocaleString('zh-CN')}
                    </td>
                    <td>
                      <div style={{ color: '#ccc', marginBottom: 4 }}>{genderLabel(chart.gender)}命</div>
                      <div style={{ fontSize: 11, color: '#666', display: 'flex', alignItems: 'center', gap: 4 }}>
                        <Calendar size={11} /> {chart.birth_year}年{chart.birth_month}月{chart.birth_day}日
                      </div>
                    </td>
                    <td>
                      <span style={{
                        display: 'inline-block', padding: '2px 8px', borderRadius: 4,
                        background: 'rgba(211,47,47,0.15)', color: '#ff6b6b', fontSize: 12, fontWeight: 600
                      }}>
                        {chart.year_gan}{chart.year_zhi}年
                      </span>
                    </td>
                    <td>
                      <button
                        onClick={() => setExpandedId(expandedId === chart.id ? null : chart.id)}
                        style={{
                          padding: '4px 12px', fontSize: 12, background: expandedId === chart.id ? '#a78bfa' : '#2a2a3a',
                          border: expandedId === chart.id ? 'none' : '1px solid #444',
                          borderRadius: 6, color: expandedId === chart.id ? '#fff' : '#ccc', cursor: 'pointer'
                        }}
                      >
                        {expandedId === chart.id ? '收起详情' : '查看详情'}
                      </button>
                    </td>
                  </tr>
                  
                  {/* 展开的详情面板 */}
                  {expandedId === chart.id && (
                    <tr style={{ background: 'rgba(0,0,0,0.2)' }}>
                      <td colSpan={5} style={{ padding: '20px 24px', borderLeft: '3px solid #a78bfa' }}>
                        
                        <div style={{ display: 'flex', gap: 40 }}>
                          {/* 四柱排盘区 */}
                          <div>
                            <div style={{ fontSize: 13, color: '#888', marginBottom: 12 }}>命局四柱：</div>
                            <div style={{ display: 'flex', gap: 16 }}>
                              {[
                                { label: '年柱', gan: chart.year_gan, zhi: chart.year_zhi },
                                { label: '月柱', gan: chart.month_gan, zhi: chart.month_zhi },
                                { label: '日柱', gan: chart.day_gan, zhi: chart.day_zhi },
                                { label: '时柱', gan: chart.hour_gan, zhi: chart.hour_zhi },
                              ].map((col, idx) => (
                                <div key={idx} style={{ textAlign: 'center', background: '#222', padding: '8px 16px', borderRadius: 8, border: '1px solid #333' }}>
                                  <div style={{ fontSize: 11, color: '#666', marginBottom: 8 }}>{col.label}</div>
                                  <div style={{ fontSize: 18, fontWeight: 600, color: '#ccc', marginBottom: 4 }}>{col.gan}</div>
                                  <div style={{ fontSize: 18, fontWeight: 600, color: '#ccc' }}>{col.zhi}</div>
                                </div>
                              ))}
                            </div>
                          </div>

                          {/* 喜用神与AI分析区 */}
                          <div style={{ flex: 1 }}>
                            <div style={{ fontSize: 13, color: '#888', marginBottom: 12 }}>算法用忌与 AI 分析：</div>
                            {chart.yongshen ? (
                              <div style={{ background: '#222', padding: '12px 16px', borderRadius: 8, border: '1px solid #333', marginBottom: 12 }}>
                                <div style={{ fontSize: 13, display: 'flex', gap: 16 }}>
                                  <span style={{ color: '#ff6b6b' }}><strong>喜用神：</strong> {chart.yongshen}</span>
                                  <span style={{ color: '#66aa66' }}><strong>忌神：</strong> {chart.jishen}</span>
                                </div>
                              </div>
                            ) : null}
                            
                            {chart.ai_result_structured ? (
                               <div style={{ background: 'rgba(167, 139, 250, 0.05)', padding: 16, borderRadius: 8, border: '1px solid rgba(167, 139, 250, 0.2)', marginBottom: 16 }}>
                                 <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
                                   {chart.ai_result_structured.personality && (
                                     <div style={{ background: 'rgba(0,0,0,0.2)', padding: '10px 12px', borderRadius: 6 }}>
                                       <div style={{ fontSize: 11, color: '#a78bfa', marginBottom: 6, fontWeight: 600 }}>性格特质</div>
                                       <div style={{ fontSize: 12, color: '#ccc', lineHeight: 1.6 }}>{chart.ai_result_structured.personality}</div>
                                     </div>
                                   )}
                                   {chart.ai_result_structured.career && (
                                     <div style={{ background: 'rgba(0,0,0,0.2)', padding: '10px 12px', borderRadius: 6 }}>
                                       <div style={{ fontSize: 11, color: '#fbbf24', marginBottom: 6, fontWeight: 600 }}>事业财运</div>
                                       <div style={{ fontSize: 12, color: '#ccc', lineHeight: 1.6 }}>{chart.ai_result_structured.career}</div>
                                     </div>
                                   )}
                                   {chart.ai_result_structured.romance && (
                                     <div style={{ background: 'rgba(0,0,0,0.2)', padding: '10px 12px', borderRadius: 6 }}>
                                       <div style={{ fontSize: 11, color: '#f472b6', marginBottom: 6, fontWeight: 600 }}>感情婚姻</div>
                                       <div style={{ fontSize: 12, color: '#ccc', lineHeight: 1.6 }}>{chart.ai_result_structured.romance}</div>
                                     </div>
                                   )}
                                   {chart.ai_result_structured.health && (
                                     <div style={{ background: 'rgba(0,0,0,0.2)', padding: '10px 12px', borderRadius: 6 }}>
                                       <div style={{ fontSize: 11, color: '#34d399', marginBottom: 6, fontWeight: 600 }}>健康运势</div>
                                       <div style={{ fontSize: 12, color: '#ccc', lineHeight: 1.6 }}>{chart.ai_result_structured.health}</div>
                                     </div>
                                   )}
                                 </div>
                               </div>
                            ) : chart.ai_result ? (
                              <div style={{ background: '#222', padding: 16, borderRadius: 8, border: '1px dashed #555', color: '#888', fontSize: 12, marginBottom: 16, fontStyle: 'italic' }}>
                                此命盘 AI 报告为旧格式，无结构化内容可展示。
                              </div>
                            ) : (
                               <div style={{ background: '#222', padding: 16, borderRadius: 8, border: '1px dashed #444', color: '#666', fontSize: 13, marginBottom: 16 }}>
                                 此命盘尚未生成 AI 原局报告。
                               </div>
                            )}

                            {/* 流年历年报告库 */}
                            <div style={{ fontSize: 13, color: '#888', marginBottom: 12 }}>流年批断记录 (共 {(liunianReports[chart.id] || []).length} 条)：</div>
                            {liunianLoading[chart.id] && <div style={{ fontSize: 12, color: '#666' }}>加载中...</div>}
                            
                            {!liunianLoading[chart.id] && (liunianReports[chart.id] || []).length === 0 && (
                              <div style={{ fontSize: 12, color: '#666', fontStyle: 'italic' }}>无流年生成记录</div>
                            )}

                            {!liunianLoading[chart.id] && (liunianReports[chart.id] || []).map((lr) => (
                              <div key={lr.id} style={{ background: '#222', border: '1px solid #333', borderRadius: 8, marginBottom: 12, overflow: 'hidden' }}>
                                <div style={{ background: '#2a2a3a', padding: '8px 12px', display: 'flex', justifyContent: 'space-between', alignItems: 'center', borderBottom: '1px solid #333' }}>
                                  <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                                    <span style={{ color: '#fff', fontWeight: 'bold', fontSize: 14 }}>{lr.target_year} 年</span>
                                    <span style={{ fontSize: 12, color: '#a78bfa', background: 'rgba(167,139,250,0.1)', padding: '2px 6px', borderRadius: 4 }}>大运: {lr.dayun_ganzhi}</span>
                                  </div>
                                  <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                                    <span style={{ fontSize: 11, color: '#888' }}>{lr.model}</span>
                                    <span style={{ fontSize: 11, color: '#666' }}>{new Date(lr.created_at).toLocaleString('zh-CN')}</span>
                                    <button 
                                      onClick={() => handleDeleteLiunian(lr.id, lr.chart_id)}
                                      title="清除此年流年报告缓存"
                                      style={{ background: 'transparent', border: 'none', color: '#ff6b6b', cursor: 'pointer', display: 'flex', alignItems: 'center' }}
                                    >
                                      <Trash2 size={14} />
                                    </button>
                                  </div>
                                </div>
                                <div style={{ padding: 12, fontSize: 13, color: '#ccc', lineHeight: 1.5, maxHeight: 150, overflowY: 'auto' }}>
                                  {lr.content_structured ? (
                                    <>
                                      <div style={{ marginBottom: 8 }}><strong style={{ color: '#888' }}>事业财运:</strong> {lr.content_structured.career}</div>
                                      <div style={{ marginBottom: 8 }}><strong style={{ color: '#888' }}>感情桃花:</strong> {lr.content_structured.romance}</div>
                                      <div style={{ marginBottom: 8 }}><strong style={{ color: '#888' }}>健康风险:</strong> {lr.content_structured.health}</div>
                                      <div><strong style={{ color: '#a78bfa' }}>年度锦囊:</strong> {lr.content_structured.advice}</div>
                                    </>
                                  ) : (
                                    <div style={{ color: '#ff6b6b' }}>数据异常 (无结构化内容)</div>
                                  )}
                                </div>
                              </div>
                            ))}
                          </div>
                        </div>

                      </td>
                    </tr>
                  )}
                </React.Fragment>
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
                background: page === 1 ? '#1a1a2e' : '#2a2a3a',
                color: page === 1 ? '#555' : '#ccc', cursor: page === 1 ? 'not-allowed' : 'pointer'
              }}
            >
              上一页
            </button>
            <span style={{ lineHeight: '32px', fontSize: 13, color: '#666', margin: '0 8px' }}>
              第 {page} / {totalPages} 页
            </span>
            <button
              disabled={page >= totalPages}
              onClick={() => setPage(p => p + 1)}
              style={{
                padding: '6px 14px', borderRadius: 8, border: 'none',
                background: page >= totalPages ? '#1a1a2e' : '#2a2a3a',
                color: page >= totalPages ? '#555' : '#ccc', cursor: page >= totalPages ? 'not-allowed' : 'pointer'
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
