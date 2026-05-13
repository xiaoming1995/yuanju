import { useEffect, useRef, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { ChevronDown, ChevronLeft, Loader2 } from 'lucide-react'
import { baziAPI } from '../lib/api'
import { useAuth } from '../contexts/AuthContext'

interface YearEvent {
  year: number
  age: number
  gan_zhi: string
  dayun_gan_zhi: string
  dayun_index: number
  signals: string[]
  narrative: string
  evidence_summary?: string[]
}

interface DayunMeta {
  index: number
  gan_zhi: string
  start_age: number
  end_age: number
  start_year: number
  end_year: number
}

interface DayunSummary {
  themes: string[]
  summary: string
  loading?: boolean
  error?: string
}

const SIGNAL_LABEL: Record<string, { label: string; color: string }> = {
  '婚恋_合': { label: '婚恋↑', color: 'var(--wu-huo)' },
  '婚恋_冲': { label: '婚恋↓', color: '#e77' },
  '婚恋_变': { label: '婚恋变', color: 'var(--wu-tu)' },
  '事业':    { label: '事业', color: 'var(--wu-mu)' },
  '财运_得': { label: '财运↑', color: 'var(--wu-jin)' },
  '财运_损': { label: '财运↓', color: '#888' },
  '健康':    { label: '健康↓', color: '#e77' },
  '迁变':    { label: '迁变', color: 'var(--wu-shui)' },
  '伏吟':    { label: '伏吟', color: '#e77' },
  '反吟':    { label: '反吟', color: '#e77' },
  '大运合化': { label: '合化', color: 'var(--wu-tu)' },
  '喜神临运': { label: '喜神', color: 'var(--wu-jin)' },
  '综合变动': { label: '变动', color: 'var(--wu-shui)' },
  // 读书期专属（age < 18 由后端自动重映射）
  '学业_资源': { label: '学业↑', color: 'var(--wu-mu)' },
  '学业_竞争': { label: '竞争', color: '#888' },
  '学业_压力': { label: '压力↓', color: '#e77' },
  '学业_贵人': { label: '贵人', color: 'var(--wu-mu)' },
  '学业_才艺': { label: '才艺', color: 'var(--wu-mu)' },
  '性格_情谊': { label: '情谊', color: 'var(--wu-tu)' },
  '性格_叛逆': { label: '叛逆', color: '#e77' },
}

const WUXING_GAN: Record<string, string> = {
  '甲': 'mu', '乙': 'mu', '丙': 'huo', '丁': 'huo',
  '戊': 'tu', '己': 'tu', '庚': 'jin', '辛': 'jin',
  '壬': 'shui', '癸': 'shui',
}

const currentYear = new Date().getFullYear()

export default function PastEventsPage() {
  const { chartId } = useParams<{ chartId: string }>()
  const navigate = useNavigate()
  const { user, isLoading } = useAuth()
  const [yearsLoaded, setYearsLoaded] = useState(false)
  const [yearsError, setYearsError] = useState('')
  const [events, setEvents] = useState<YearEvent[]>([])
  const [dayunMeta, setDayunMeta] = useState<DayunMeta[]>([])
  const [summaries, setSummaries] = useState<Record<number, DayunSummary>>({})
  const [expandedEvidence, setExpandedEvidence] = useState<Record<string, boolean>>({})
  const [streamDone, setStreamDone] = useState(false)
  const [streamError, setStreamError] = useState('')
  const inflightRef = useRef(false)

  useEffect(() => {
    if (isLoading) return
    if (!user) {
      navigate('/login')
      return
    }
    if (!chartId) return
    loadAll()
  }, [chartId, user, isLoading])

  async function loadAll() {
    if (inflightRef.current) return
    inflightRef.current = true
    setYearsLoaded(false)
    setYearsError('')
    setStreamDone(false)
    setStreamError('')
    setExpandedEvidence({})

    // Stage 1: 即时拿所有年份（毫秒级）
    try {
      const resp = await baziAPI.fetchPastEventsYears(chartId!)
      const data = resp.data
      setEvents(data.years || [])
      setDayunMeta(data.dayun_meta || [])
      // 初始化各大运 summary 占位（含远期大运，全量 AI 生成）
      const init: Record<number, DayunSummary> = {}
      for (const dm of data.dayun_meta || []) {
        init[dm.index] = { themes: [], summary: '', loading: true }
      }
      setSummaries(init)
      setYearsLoaded(true)
    } catch (e: any) {
      setYearsError(e?.message || '年份加载失败')
      inflightRef.current = false
      return
    }

    // Stage 2: 后台流式拉大运 AI 总结
    baziAPI.streamDayunSummaries(
      chartId!,
      (item) => {
        setSummaries((prev) => {
          const next = { ...prev }
          if (item.error) {
            next[item.dayun_index] = { themes: [], summary: '', error: item.error, loading: false }
          } else {
            next[item.dayun_index] = {
              themes: item.themes || [],
              summary: item.summary || '',
              loading: false,
            }
          }
          return next
        })
      },
      (err) => {
        setStreamError(err)
        inflightRef.current = false
      },
      () => {
        setStreamDone(true)
        inflightRef.current = false
      },
    )
  }

  // 按 dayun_index 分组（保持原顺序）
  const grouped: Array<{ meta: DayunMeta; years: YearEvent[] }> = dayunMeta.map((dm) => ({
    meta: dm,
    years: events.filter((y) => y.dayun_index === dm.index),
  })).filter((g) => g.years.length > 0)

  return (
    <div style={{ minHeight: '100vh', background: 'var(--bg-base)', paddingBottom: 60 }}>
      {/* 顶部导航 */}
      <div style={{
        display: 'flex', alignItems: 'center', gap: 12,
        padding: '16px 20px',
        background: 'var(--bg-card)',
        borderBottom: '1px solid var(--border-subtle)',
        position: 'sticky', top: 0, zIndex: 100,
      }}>
        <button
          onClick={() => navigate(-1)}
          style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'var(--text-secondary)', padding: 4 }}
        >
          <ChevronLeft size={20} />
        </button>
        <div>
          <div style={{ fontFamily: 'Noto Serif SC, serif', fontSize: '1rem', color: 'var(--text-primary)', fontWeight: 600 }}>
            过往事件推算
          </div>
          <div style={{ fontSize: '0.7rem', color: 'var(--text-muted)', marginTop: 2 }}>
            {!yearsLoaded ? '正在加载年份时间轴……' :
             streamDone ? '已完成，所有大运总结已生成' :
             '年份已就绪 · 大运总结正在后台生成'}
          </div>
        </div>
      </div>

      <div style={{ maxWidth: 700, margin: '0 auto', padding: '24px 16px' }}>
        {!yearsLoaded && !yearsError && (
          <div style={{ textAlign: 'center', padding: '60px 0' }}>
            <Loader2 size={32} style={{ color: 'var(--wu-jin)', animation: 'spin 1s linear infinite', marginBottom: 16 }} />
            <div style={{ color: 'var(--text-secondary)', fontSize: '0.9rem' }}>正在加载……</div>
          </div>
        )}

        {yearsError && (
          <div style={{ textAlign: 'center', padding: '40px 0' }}>
            <div style={{ color: '#e77', marginBottom: 16 }}>{yearsError}</div>
            <button
              onClick={loadAll}
              style={{
                background: 'var(--wu-jin)', color: '#000', border: 'none',
                borderRadius: 8, padding: '10px 24px', cursor: 'pointer', fontWeight: 600,
              }}
            >重新加载</button>
          </div>
        )}

        {yearsLoaded && events.length === 0 && (
          <div style={{ textAlign: 'center', padding: '40px 0', color: 'var(--text-muted)' }}>
            暂无过往年份数据
          </div>
        )}

        {yearsLoaded && events.length > 0 && (
          <div>
            <div style={{ color: 'var(--text-muted)', fontSize: '0.75rem', marginBottom: 20, textAlign: 'center' }}>
              共推算 {events.length} 个年份 · 算法即时生成 · 大运总结后台生成
            </div>

            {grouped.map(({ meta, years }) => {
              const dyGan = meta.gan_zhi[0] || ''
              const dyWx = WUXING_GAN[dyGan] || 'tu'
              const dySum = summaries[meta.index]
              return (
                <div key={meta.index} style={{ marginBottom: 32 }}>
                  {/* 大运标题 */}
                  <div style={{
                    display: 'flex', alignItems: 'center', gap: 8,
                    marginBottom: dySum ? 10 : 12,
                    paddingBottom: dySum ? 0 : 8,
                    borderBottom: dySum ? 'none' : '1px solid var(--border-subtle)',
                  }}>
                    <div style={{
                      fontFamily: 'Noto Serif SC, serif',
                      fontSize: '1.1rem',
                      fontWeight: 700,
                      color: `var(--wu-${dyWx})`,
                    }}>{meta.gan_zhi}</div>
                    <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
                      大运 {meta.start_age}-{meta.end_age}岁
                    </div>
                  </div>

                  {/* 大运整体总结块 */}
                  {dySum && (
                    <div style={{
                      background: `color-mix(in srgb, var(--wu-${dyWx}) 8%, var(--bg-card))`,
                      border: `1px solid color-mix(in srgb, var(--wu-${dyWx}) 25%, transparent)`,
                      borderRadius: 10,
                      padding: '12px 14px',
                      marginBottom: 12,
                    }}>
                      {dySum.loading && (
                        <div style={{ display: 'flex', alignItems: 'center', gap: 8, color: 'var(--text-muted)', fontSize: '0.78rem' }}>
                          <Loader2 size={14} style={{ animation: 'spin 1s linear infinite' }} />
                          正在生成本段大运总结……
                        </div>
                      )}
                      {dySum.error && (
                        <div style={{ color: '#e77', fontSize: '0.78rem' }}>
                          本段总结生成失败：{dySum.error}
                        </div>
                      )}
                      {!dySum.loading && !dySum.error && dySum.themes.length > 0 && (
                        <>
                          <div style={{ display: 'flex', flexWrap: 'wrap', gap: 6, marginBottom: 8 }}>
                            {dySum.themes.map((theme) => {
                              const hasUp = theme.includes('↑')
                              const hasDown = theme.includes('↓')
                              const direction = hasUp ? '↑' : hasDown ? '↓' : null
                              const text = direction ? theme.replace(direction, '').trim() : theme
                              const dirColor = hasUp ? '#66bb6a' : hasDown ? '#ef5350' : null
                              const borderColor = dirColor ?? `var(--wu-${dyWx})`
                              return (
                                <span
                                  key={theme}
                                  style={{
                                    display: 'inline-flex',
                                    alignItems: 'center',
                                    gap: 2,
                                    fontSize: '0.68rem',
                                    padding: '2px 7px',
                                    borderRadius: 4,
                                    border: `1px solid ${borderColor}`,
                                    color: `var(--wu-${dyWx})`,
                                    whiteSpace: 'nowrap',
                                  }}
                                >
                                  {text}
                                  {direction && (
                                    <span style={{ fontSize: '0.85rem', fontWeight: 700, color: dirColor!, lineHeight: 1 }}>{direction}</span>
                                  )}
                                </span>
                              )
                            })}
                          </div>
                          <div style={{ color: 'var(--text-secondary)', fontSize: '0.83rem', lineHeight: 1.75 }}>
                            {dySum.summary}
                          </div>
                        </>
                      )}
                    </div>
                  )}

                  {/* 年份列表 */}
                  <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
                    {years.map((y) => {
                      const gan = y.gan_zhi?.[0] || ''
                      const wx = WUXING_GAN[gan] || 'tu'
                      const hasSignals = y.signals && y.signals.length > 0
                      const isFuture = y.year > currentYear
                      const evidenceKey = `${meta.index}-${y.year}`
                      const hasEvidence = Boolean(y.evidence_summary?.length)
                      const evidenceOpen = Boolean(expandedEvidence[evidenceKey])
                      return (
                        <div
                          key={y.year}
                          style={{
                            position: 'relative',
                            background: 'var(--bg-card)',
                            borderRadius: 12,
                            padding: '14px 16px',
                            opacity: isFuture ? 0.75 : 1,
                            border: hasSignals
                              ? `1px ${isFuture ? 'dashed' : 'solid'} color-mix(in srgb, var(--wu-${wx}) 40%, transparent)`
                              : `1px ${isFuture ? 'dashed' : 'solid'} var(--border-subtle)`,
                          }}
                        >
                          {isFuture && (
                            <span style={{
                              position: 'absolute',
                              top: 8,
                              right: 10,
                              fontSize: '0.6rem',
                              padding: '1px 5px',
                              borderRadius: 3,
                              background: 'color-mix(in srgb, var(--wu-shui) 15%, transparent)',
                              border: '1px dashed var(--wu-shui)',
                              color: 'var(--wu-shui)',
                              letterSpacing: '0.05em',
                            }}>未来</span>
                          )}
                          <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8, flexWrap: 'wrap' }}>
                            <span style={{
                              fontFamily: 'Noto Serif SC, serif',
                              fontWeight: 700,
                              fontSize: '1.05rem',
                              color: `var(--wu-${wx})`,
                            }}>{y.gan_zhi}</span>
                            <span style={{ color: 'var(--text-muted)', fontSize: '0.8rem' }}>
                              {y.year}年 · {y.age}岁
                            </span>
                            {y.signals?.map((sig) => {
                              const meta = SIGNAL_LABEL[sig]
                              if (!meta) return null
                              return (
                                <span
                                  key={sig}
                                  style={{
                                    fontSize: '0.65rem',
                                    padding: '2px 6px',
                                    borderRadius: 4,
                                    border: `1px solid ${meta.color}`,
                                    color: meta.color,
                                    whiteSpace: 'nowrap',
                                  }}
                                >{meta.label}</span>
                              )
                            })}
                          </div>
                          <div style={{ color: 'var(--text-secondary)', fontSize: '0.85rem', lineHeight: 1.7 }}>
                            {y.narrative}
                          </div>
                          {hasEvidence && (
                            <div style={{ marginTop: 10 }}>
                              <button
                                type="button"
                                onClick={() => setExpandedEvidence((prev) => ({
                                  ...prev,
                                  [evidenceKey]: !prev[evidenceKey],
                                }))}
                                style={{
                                  display: 'inline-flex',
                                  alignItems: 'center',
                                  gap: 4,
                                  border: 'none',
                                  background: 'transparent',
                                  color: 'var(--text-muted)',
                                  fontSize: '0.72rem',
                                  padding: '2px 0',
                                  cursor: 'pointer',
                                }}
                                aria-expanded={evidenceOpen}
                              >
                                <ChevronDown
                                  size={14}
                                  style={{
                                    transform: evidenceOpen ? 'rotate(180deg)' : 'rotate(0deg)',
                                    transition: 'transform 0.16s ease',
                                  }}
                                />
                                命理依据
                              </button>
                              {evidenceOpen && (
                                <ul style={{
                                  margin: '8px 0 0',
                                  paddingLeft: 18,
                                  color: 'var(--text-muted)',
                                  fontSize: '0.72rem',
                                  lineHeight: 1.65,
                                }}>
                                  {y.evidence_summary!.map((ev, idx) => (
                                    <li key={`${evidenceKey}-${idx}`}>{ev}</li>
                                  ))}
                                </ul>
                              )}
                            </div>
                          )}
                        </div>
                      )
                    })}
                  </div>
                </div>
              )
            })}

            {streamError && (
              <div style={{
                marginTop: 16, padding: '10px 14px',
                background: 'color-mix(in srgb, #e77 12%, transparent)',
                border: '1px solid #e77',
                borderRadius: 8,
                fontSize: '0.78rem',
                color: '#e77',
              }}>
                大运总结生成中断：{streamError}
                <button
                  onClick={loadAll}
                  style={{
                    marginLeft: 12,
                    background: 'none', border: '1px solid #e77',
                    borderRadius: 4, padding: '2px 10px',
                    color: '#e77', cursor: 'pointer', fontSize: '0.72rem',
                  }}
                >重试</button>
              </div>
            )}

            <div style={{
              marginTop: 24, padding: '12px 16px',
              background: 'var(--bg-elevated)',
              borderRadius: 8,
              fontSize: '0.72rem',
              color: 'var(--text-muted)',
              lineHeight: 1.6,
            }}>
              本推算内容仅供参考，不构成任何决策建议。
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
