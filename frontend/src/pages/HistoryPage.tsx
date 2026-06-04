import { useEffect, useRef, useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { CalendarDays, Compass, HeartHandshake, Sparkles, Trash2 } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'
import { baziAPI, type BaziHistoryChart } from '../lib/api'
import { chartDisplayName, chartFallbackName, formatPillars, genderText } from '../lib/chartLabel'
import { Button } from '../components/ui/Button'
import { ConfirmDialog } from '../components/ui/ConfirmDialog'
import './HistoryPage.css'

type Chart = BaziHistoryChart

function formatDate(value: string) {
  if (!value) return '-'
  return new Date(value).toLocaleDateString('zh-CN')
}

function isInteractiveTarget(target: EventTarget | null) {
  return target instanceof HTMLElement && Boolean(target.closest('button, input, select, textarea, a'))
}

export default function HistoryPage() {
  const { user, isLoading } = useAuth()
  const navigate = useNavigate()
  const roleDialogRef = useRef<HTMLDivElement | null>(null)
  const [charts, setCharts] = useState<Chart[]>([])
  const [loading, setLoading] = useState(true)
  const [editingChartId, setEditingChartId] = useState<string | null>(null)
  const [displayNameDraft, setDisplayNameDraft] = useState('')
  const [displayNameError, setDisplayNameError] = useState('')
  const [compatibilityRoleChart, setCompatibilityRoleChart] = useState<Chart | null>(null)
  const [deletingChart, setDeletingChart] = useState<Chart | null>(null)
  const [deleteError, setDeleteError] = useState('')
  const [deleteLoading, setDeleteLoading] = useState(false)

  useEffect(() => {
    if (isLoading) return
    if (!user) { navigate('/login'); return }
    baziAPI.getHistory()
      .then(res => setCharts(res.data.charts || []))
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [user, isLoading, navigate])

  useEffect(() => {
    if (!compatibilityRoleChart) return
    const previous = document.activeElement instanceof HTMLElement ? document.activeElement : null
    roleDialogRef.current?.focus()
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') setCompatibilityRoleChart(null)
    }
    document.addEventListener('keydown', handleKeyDown)
    return () => {
      document.removeEventListener('keydown', handleKeyDown)
      previous?.focus()
    }
  }, [compatibilityRoleChart])

  const latestChart = charts[0]
  const maleCount = charts.filter(chart => chart.gender === 'male').length
  const femaleCount = charts.filter(chart => chart.gender === 'female').length
  const stats = [
    { label: '命盘记录', value: charts.length ? `${charts.length}` : '0', icon: CalendarDays },
    { label: '最近起盘', value: latestChart ? formatDate(latestChart.created_at) : '-', icon: Sparkles },
    { label: '男女命', value: `${maleCount}/${femaleCount}`, icon: Compass },
  ]

  const startEditDisplayName = (chart: Chart) => {
    setEditingChartId(chart.id)
    setDisplayNameDraft(chart.display_name || '')
    setDisplayNameError('')
  }

  const handleSaveDisplayName = async (chart: Chart) => {
    const nextName = displayNameDraft.trim()
    if (Array.from(nextName).length > 20) {
      setDisplayNameError('称呼不能超过20个字符')
      return
    }
    try {
      const res = await baziAPI.updateHistoryDisplayName(chart.id, nextName)
      const savedName = res.data.data.display_name
      setCharts(prev => prev.map(item => item.id === chart.id ? { ...item, display_name: savedName } : item))
      setEditingChartId(null)
      setDisplayNameDraft('')
      setDisplayNameError('')
    } catch (err: unknown) {
      setDisplayNameError(err instanceof Error ? err.message : '保存称呼失败')
    }
  }

  const launchCompatibility = (role: 'self' | 'partner') => {
    if (!compatibilityRoleChart) return
    navigate(`/compatibility?importChart=${compatibilityRoleChart.id}&role=${role}`)
  }

  const handleConfirmDelete = async () => {
    if (!deletingChart) return
    setDeleteLoading(true)
    setDeleteError('')
    try {
      await baziAPI.deleteHistory(deletingChart.id)
      setCharts(prev => prev.filter(item => item.id !== deletingChart.id))
      setDeletingChart(null)
    } catch (err: unknown) {
      setDeleteError(err instanceof Error ? err.message : '删除失败，请重试')
    } finally {
      setDeleteLoading(false)
    }
  }

  if (loading || isLoading) return (
    <div className="history-page page">
      <div className="container">
        <div className="history-loading">
          {[1,2,3].map(i => <div key={i} className="skeleton history-skeleton" />)}
        </div>
      </div>
    </div>
  )

  return (
    <div className="history-page page">
      <div className="container">
        <section className="history-archive-hero card animate-fade-up">
          <div>
            <p className="history-kicker">命盘档案</p>
            <h1 className="history-title serif">我的命盘档案</h1>
            <p className="history-desc">已保存 {charts.length} 份命盘，回看四柱、出生信息和后续分析入口。</p>
          </div>
          <Link to="/" className="btn btn-primary">新建命盘</Link>
        </section>

        <nav className="archive-switcher animate-fade-up" aria-label="档案类型">
          <Link to="/history" className="archive-switcher-item archive-switcher-item--active">
            <Compass size={17} />
            <span>命盘档案</span>
          </Link>
          <Link to="/compatibility/history" className="archive-switcher-item">
            <HeartHandshake size={17} />
            <span>合盘档案</span>
          </Link>
        </nav>

        <section className="history-stat-grid animate-fade-up">
          {stats.map(item => {
            const Icon = item.icon
            return (
              <div className="history-stat-card" key={item.label}>
                <Icon size={18} />
                <span>{item.label}</span>
                <strong>{item.value}</strong>
              </div>
            )
          })}
        </section>

        <div className="history-section-title animate-fade-up">
          <div>
            <h2 className="serif">最近命盘</h2>
            <p>按保存时间排列，点击可查看完整命盘。</p>
          </div>
        </div>

        {charts.length === 0 ? (
          <div className="history-empty card animate-fade-up">
            <div className="empty-icon"><Compass size={48} /></div>
            <p className="empty-title serif">还没有命盘记录</p>
            <p className="empty-desc">起盘后登录即可自动保存</p>
            <Link to="/" className="btn btn-primary">立即起盘</Link>
          </div>
        ) : (
          <div className="history-list animate-fade-up">
            {charts.map(c => (
              <article
                key={c.id}
                className="history-record-card card"
                role="link"
                tabIndex={0}
                onClick={(event) => {
                  if (isInteractiveTarget(event.target)) return
                  navigate(`/history/${c.id}`)
                }}
                onKeyDown={(event) => {
                  if (isInteractiveTarget(event.target)) return
                  if (event.key === 'Enter' || event.key === ' ') {
                    event.preventDefault()
                    navigate(`/history/${c.id}`)
                  }
                }}
              >
                <div className="history-record-main">
                  <div>
                    {editingChartId === c.id ? (
                      <div
                        className="history-name-editor"
                        onClick={(event) => event.stopPropagation()}
                        onKeyDown={(event) => event.stopPropagation()}
                      >
                        <input
                          value={displayNameDraft}
                          onChange={(event) => setDisplayNameDraft(event.target.value)}
                          maxLength={20}
                          placeholder={chartFallbackName(c)}
                          aria-label="命盘称呼"
                        />
                        <button
                          type="button"
                          className="history-inline-action"
                          onClick={() => handleSaveDisplayName(c)}
                        >
                          保存
                        </button>
                        <button
                          type="button"
                          className="history-inline-action"
                          onClick={() => {
                            setEditingChartId(null)
                            setDisplayNameDraft('')
                            setDisplayNameError('')
                          }}
                        >
                          取消
                        </button>
                        {displayNameError ? <span className="history-name-error">{displayNameError}</span> : null}
                      </div>
                    ) : (
                      <div className="history-display-name">
                        <span>{chartDisplayName(c)}</span>
                        <button
                          type="button"
                          className="history-inline-action"
                          onClick={(event) => {
                            event.stopPropagation()
                            startEditDisplayName(c)
                          }}
                        >
                          编辑称呼
                        </button>
                      </div>
                    )}
                    <div className="history-pillars serif">{formatPillars(c)}</div>
                    <div className="history-meta">
                      <span>{c.birth_year}年{c.birth_month}月{c.birth_day}日 {c.birth_hour}时</span>
                      <span>{genderText(c.gender)}</span>
                    </div>
                  </div>
                  <div
                    className="history-record-actions"
                    onKeyDown={(event) => event.stopPropagation()}
                  >
                    <button
                      type="button"
                      className="history-record-action-button"
                      onClick={(event) => {
                        event.stopPropagation()
                        setCompatibilityRoleChart(c)
                      }}
                    >
                      用此命盘合盘
                    </button>
                    <button
                      type="button"
                      className="history-record-delete-button"
                      onClick={(event) => {
                        event.stopPropagation()
                        setDeleteError('')
                        setDeletingChart(c)
                      }}
                    >
                      <Trash2 size={14} />
                      删除
                    </button>
                    <Button
                      href={`/bazi/${c.id}/past-events`}
                      variant="ghost"
                      size="sm"
                      onClick={(event) => event.stopPropagation()}
                    >
                      查看过往事件
                    </Button>
                    <span className="history-record-action">查看结果</span>
                  </div>
                </div>
                <div className="history-record-footer">
                  <span>保存于 {formatDate(c.created_at)}</span>
                  <span>四柱档案</span>
                </div>
              </article>
            ))}
          </div>
        )}

        {compatibilityRoleChart ? (
          <div
            className="history-role-dialog"
            role="dialog"
            aria-modal="true"
            aria-label="选择合盘身份"
            onClick={() => setCompatibilityRoleChart(null)}
          >
            <div
              ref={roleDialogRef}
              className="history-role-dialog-panel"
              tabIndex={-1}
              onClick={(event) => event.stopPropagation()}
            >
              <p className="history-kicker">选择合盘身份</p>
              <h2 className="serif">{chartDisplayName(compatibilityRoleChart)}</h2>
              <p>请选择这份命盘在合盘中的角色。</p>
              <div className="history-role-actions">
                <button type="button" className="btn btn-primary" onClick={() => launchCompatibility('self')}>
                  作为我
                </button>
                <button type="button" className="btn btn-secondary" onClick={() => launchCompatibility('partner')}>
                  作为对方
                </button>
              </div>
              <button type="button" className="history-inline-action" onClick={() => setCompatibilityRoleChart(null)}>
                取消
              </button>
            </div>
          </div>
        ) : null}

        <ConfirmDialog
          open={Boolean(deletingChart)}
          title={deletingChart ? `删除 ${chartDisplayName(deletingChart)}` : '删除命盘'}
          description={deleteError ? <>删除后该命盘及其 AI 报告将永久消失，无法恢复。<br />{deleteError}</> : '删除后该命盘及其 AI 报告将永久消失，无法恢复。'}
          confirmText="删除"
          danger
          pending={deleteLoading}
          onConfirm={handleConfirmDelete}
          onCancel={() => { if (!deleteLoading) setDeletingChart(null) }}
        />
      </div>
    </div>
  )
}
