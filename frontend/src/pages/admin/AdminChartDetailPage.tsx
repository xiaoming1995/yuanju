import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { ArrowLeft, BookOpen, Calendar, RefreshCw, Trash2, User } from 'lucide-react'
import { adminChartsAPI } from '../../lib/adminApi'
import './AdminChartDetailPage.css'

interface ChartRecord {
  id: string
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
    chapters?: Array<{ title: string; brief: string; detail?: string }>
    analysis?: { summary?: string }
    personality?: string
    career?: string
    romance?: string
    health?: string
  }
  created_at: string
}

interface AdminLiunianReport {
  id: string
  target_year: number
  dayun_ganzhi: string
  content_structured?: {
    career?: string
    romance?: string
    health?: string
    advice?: string
  } | null
  model: string
  created_at: string
}

interface AdminPastEventsYear {
  year: number
  ganzhi?: string
  gan_zhi?: string
  narrative?: string
}

interface AdminPastEventsRecord {
  id: string
  dayun_index: number
  dayun_ganzhi: string
  themes?: string[] | null
  summary: string
  years?: AdminPastEventsYear[] | null
  model: string
  algorithm_version: string
  created_at: string
}

const genderLabel = (g: string) => g === 'male' ? '男' : '女'

function adminErrorMessage(err: unknown, fallback = '操作失败') {
  return err instanceof Error ? err.message : fallback
}

function formatDate(input: string) {
  return new Date(input).toLocaleString('zh-CN')
}

function EmptyState({ children }: { children: string }) {
  return <div className="admin-chart-detail-empty">{children}</div>
}

function renderStructuredReport(chart: ChartRecord) {
  const s = chart.ai_result_structured
  const chapters = s?.chapters?.length
    ? s.chapters
    : [
        s?.personality ? { title: '性格特质', brief: s.personality } : null,
        s?.career ? { title: '事业财运', brief: s.career } : null,
        s?.romance ? { title: '感情婚姻', brief: s.romance } : null,
        s?.health ? { title: '健康运势', brief: s.health } : null,
      ].filter(Boolean) as Array<{ title: string; brief: string; detail?: string }>

  if (chapters.length > 0) {
    return (
      <div className="admin-chart-report">
        {s?.analysis?.summary && <div className="admin-chart-report-summary">{s.analysis.summary}</div>}
        <div className="admin-chart-report-grid">
          {chapters.map((ch, idx) => (
            <article key={`${ch.title}-${idx}`} className="admin-chart-report-card">
              <div className="admin-chart-report-title">{ch.title}</div>
              <div className="admin-chart-report-body">{ch.detail || ch.brief}</div>
            </article>
          ))}
        </div>
      </div>
    )
  }
  if (chart.ai_result) return <EmptyState>此命盘 AI 报告为旧格式，无结构化内容可展示。</EmptyState>
  return <EmptyState>此命盘尚未生成 AI 原局报告。</EmptyState>
}

export default function AdminChartDetailPage() {
  const { chartId } = useParams()
  const [chart, setChart] = useState<ChartRecord | null>(null)
  const [pastEvents, setPastEvents] = useState<AdminPastEventsRecord[]>([])
  const [liunianReports, setLiunianReports] = useState<AdminLiunianReport[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const loadAll = async () => {
    if (!chartId) return
    try {
      setLoading(true)
      setError('')
      const [chartRes, pastRes, liunianRes] = await Promise.all([
        adminChartsAPI.detail(chartId),
        adminChartsAPI.getPastEventsRecords(chartId),
        adminChartsAPI.getLiunianReports(chartId),
      ])
      setChart(chartRes.data?.data || null)
      setPastEvents(pastRes.data?.data || [])
      setLiunianReports(liunianRes.data?.data || [])
    } catch (err: unknown) {
      setError(adminErrorMessage(err, '加载详情失败'))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadAll()
  }, [chartId]) // eslint-disable-line react-hooks/exhaustive-deps

  const handleDeleteLiunian = async (id: string) => {
    if (!window.confirm('确定要清除该年份的流年预测缓存吗？清除后前端用户将可以重新生成该年的运势。')) return
    try {
      await adminChartsAPI.deleteLiunianReport(id)
      setLiunianReports(prev => prev.filter(r => r.id !== id))
    } catch (err: unknown) {
      alert('清除失败: ' + adminErrorMessage(err, '清除失败'))
    }
  }

  if (loading) return <div className="admin-loading">加载中...</div>

  if (error || !chart) {
    return (
      <div className="admin-chart-detail">
        <Link to="/admin/charts" className="admin-chart-detail-back"><ArrowLeft size={14} /> 返回起盘明细</Link>
        <div className="admin-card admin-chart-detail-error">{error || '命盘不存在'}</div>
      </div>
    )
  }

  const pillars = [
    { label: '年柱', gan: chart.year_gan, zhi: chart.year_zhi },
    { label: '月柱', gan: chart.month_gan, zhi: chart.month_zhi },
    { label: '日柱', gan: chart.day_gan, zhi: chart.day_zhi },
    { label: '时柱', gan: chart.hour_gan, zhi: chart.hour_zhi },
  ]

  return (
    <div className="admin-chart-detail">
      <header className="admin-chart-detail-header">
        <div>
          <Link to="/admin/charts" className="admin-chart-detail-back"><ArrowLeft size={14} /> 返回起盘明细</Link>
          <h1 className="admin-page-title admin-chart-detail-title">
            <BookOpen size={24} /> 命盘详情
          </h1>
        </div>
        <button className="admin-btn admin-btn-ghost admin-chart-detail-refresh" onClick={loadAll}>
          <RefreshCw size={14} /> 刷新
        </button>
      </header>

      <section className="admin-chart-hero">
        <div className="admin-chart-profile">
          <div className="admin-chart-profile-top">
            <span className="admin-chart-profile-avatar"><User size={18} /></span>
            <div>
              <div className="admin-chart-profile-name">{chart.user_email || '游客命盘'}</div>
              <div className="admin-chart-profile-meta">{genderLabel(chart.gender)}命 · {formatDate(chart.created_at)}</div>
            </div>
          </div>
          <div className="admin-chart-profile-birth">
            <Calendar size={15} />
            {chart.birth_year}年{chart.birth_month}月{chart.birth_day}日 {chart.birth_hour}时
          </div>
          <div className="admin-chart-pill-row">
            <span className="admin-chart-pill admin-chart-pill-good">喜用神：{chart.yongshen || '未记录'}</span>
            <span className="admin-chart-pill admin-chart-pill-risk">忌神：{chart.jishen || '未记录'}</span>
          </div>
        </div>

        <div className="admin-chart-pillar-grid" aria-label="命局四柱">
          {pillars.map(col => (
            <div key={col.label} className="admin-chart-pillar">
              <div className="admin-chart-pillar-label">{col.label}</div>
              <div className="admin-chart-pillar-char">{col.gan}</div>
              <div className="admin-chart-pillar-char">{col.zhi}</div>
            </div>
          ))}
        </div>
      </section>

      <section className="admin-card admin-chart-section">
        <div className="admin-chart-section-head">
          <h2>AI 原局分析</h2>
        </div>
        {renderStructuredReport(chart)}
      </section>

      <section className="admin-card admin-chart-section">
        <div className="admin-chart-section-head">
          <h2>过往推算记录</h2>
          <span>{pastEvents.length} 段大运</span>
        </div>
        {pastEvents.length === 0 && <EmptyState>无过往推算生成记录</EmptyState>}
        <div className="admin-dayun-list">
          {pastEvents.map(record => (
            <details key={record.id} className="admin-dayun-record" open>
              <summary className="admin-dayun-record-summary">
                <div className="admin-dayun-record-main">
                  <strong>第 {record.dayun_index} 运 · {record.dayun_ganzhi || '未记录干支'}</strong>
                  <div className="admin-dayun-record-tags">
                    {(record.themes || []).map(theme => <span key={theme}>{theme}</span>)}
                  </div>
                </div>
                <div className="admin-dayun-record-meta">
                  <span>{record.model || 'unknown'}</span>
                  <span>{record.algorithm_version || 'v1'}</span>
                  <span>{formatDate(record.created_at)}</span>
                </div>
              </summary>
              <div className="admin-dayun-record-body">
                {record.summary && <p className="admin-dayun-summary-text">{record.summary}</p>}
                {(record.years || []).length > 0 ? (
                  <div className="admin-year-grid">
                    {(record.years || []).map(year => (
                      <article key={`${record.id}-${year.year}`} className="admin-year-card">
                        <div className="admin-year-card-title">{year.year} 年 {year.ganzhi || year.gan_zhi || ''}</div>
                        <p>{year.narrative || '无逐年叙述'}</p>
                      </article>
                    ))}
                  </div>
                ) : (
                  <EmptyState>该大运段没有逐年 narrative 缓存</EmptyState>
                )}
              </div>
            </details>
          ))}
        </div>
      </section>

      <section className="admin-card admin-chart-section">
        <div className="admin-chart-section-head">
          <h2>流年批断记录</h2>
          <span>{liunianReports.length} 条</span>
        </div>
        {liunianReports.length === 0 && <EmptyState>无流年生成记录</EmptyState>}
        <div className="admin-liunian-list">
          {liunianReports.map(lr => (
            <article key={lr.id} className="admin-liunian-card">
              <header className="admin-liunian-card-head">
                <div>
                  <strong>{lr.target_year} 年</strong>
                  <span>大运：{lr.dayun_ganzhi || '未记录'}</span>
                </div>
                <div className="admin-liunian-card-meta">
                  <span>{lr.model}</span>
                  <span>{formatDate(lr.created_at)}</span>
                  <button onClick={() => handleDeleteLiunian(lr.id)} title="清除此年流年报告缓存">
                    <Trash2 size={14} />
                  </button>
                </div>
              </header>
              <div className="admin-liunian-card-body">
                {lr.content_structured ? (
                  <>
                    <p><strong>事业财运</strong>{lr.content_structured.career}</p>
                    <p><strong>感情桃花</strong>{lr.content_structured.romance}</p>
                    <p><strong>健康风险</strong>{lr.content_structured.health}</p>
                    <p><strong>年度锦囊</strong>{lr.content_structured.advice}</p>
                  </>
                ) : (
                  <EmptyState>数据异常，无结构化内容</EmptyState>
                )}
              </div>
            </article>
          ))}
        </div>
      </section>
    </div>
  )
}
