import { useEffect, useRef, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { ChevronLeft, Loader2 } from 'lucide-react'
import { baziAPI } from '../lib/api'
import { useAuth } from '../contexts/AuthContext'

interface YearEvent {
  year: number
  age: number
  gan_zhi: string
  dayun_gan_zhi: string
  signals: string[]
  narrative: string
}

interface DayunSummary {
  gan_zhi: string
  themes: string[]
  summary: string
}

const SIGNAL_LABEL: Record<string, { label: string; color: string }> = {
  '婚恋':   { label: '婚恋', color: 'var(--wu-huo)' },
  '事业':   { label: '事业', color: 'var(--wu-mu)' },
  '财运_得': { label: '财运↑', color: 'var(--wu-jin)' },
  '财运_损': { label: '财运↓', color: '#888' },
  '健康':   { label: '健康', color: '#e77' },
  '迁变':   { label: '迁变', color: 'var(--wu-shui)' },
}

const WUXING_GAN: Record<string, string> = {
  '甲': 'mu', '乙': 'mu', '丙': 'huo', '丁': 'huo',
  '戊': 'tu', '己': 'tu', '庚': 'jin', '辛': 'jin',
  '壬': 'shui', '癸': 'shui',
}

export default function PastEventsPage() {
  const { chartId } = useParams<{ chartId: string }>()
  const navigate = useNavigate()
  const { user } = useAuth()
  const [status, setStatus] = useState<'idle' | 'loading' | 'done' | 'error'>('idle')
  const [events, setEvents] = useState<YearEvent[]>([])
  const [summaries, setSummaries] = useState<DayunSummary[]>([])
  const [errorMsg, setErrorMsg] = useState('')
  const [thinking, setThinking] = useState(false)
  const rawRef = useRef('')

  useEffect(() => {
    if (!user) {
      navigate('/login')
      return
    }
    if (!chartId) return
    startStream()
  }, [chartId])

  function startStream() {
    rawRef.current = ''
    setStatus('loading')
    setEvents([])
    setSummaries([])
    setErrorMsg('')
    setThinking(false)

    baziAPI.streamPastEvents(
      chartId!,
      (chunk) => {
        rawRef.current += chunk
        setThinking(false)
      },
      (err) => {
        setErrorMsg(err)
        setStatus('error')
      },
      () => {
        parseAndSet(rawRef.current)
        setStatus('done')
      },
      () => setThinking(true),
    )
  }

  function parseAndSet(raw: string) {
    try {
      const parsed = JSON.parse(raw.trim())
      setEvents(parsed.years ?? [])
      setSummaries(parsed.dayun_summaries ?? [])
    } catch {
      setErrorMsg('解析推算结果失败，请重试')
      setStatus('error')
    }
  }

  // 按大运分组
  const grouped = events.reduce<Record<string, YearEvent[]>>((acc, y) => {
    const key = y.dayun_gan_zhi || '起运前'
    if (!acc[key]) acc[key] = []
    acc[key].push(y)
    return acc
  }, {})

  // 大运总结索引，key = gan_zhi
  const summaryMap = summaries.reduce<Record<string, DayunSummary>>((acc, s) => {
    acc[s.gan_zhi] = s
    return acc
  }, {})

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
            基于命理算法 · AI 组织语言
          </div>
        </div>
      </div>

      <div style={{ maxWidth: 700, margin: '0 auto', padding: '24px 16px' }}>
        {/* 加载中状态 */}
        {status === 'loading' && (
          <div style={{ textAlign: 'center', padding: '60px 0' }}>
            <Loader2 size={32} style={{ color: 'var(--wu-jin)', animation: 'spin 1s linear infinite', marginBottom: 16 }} />
            <div style={{ color: 'var(--text-secondary)', fontSize: '0.9rem' }}>
              {thinking ? '🧠 深度推理中……' : '正在推算过往年份命理事件……'}
            </div>
            <div style={{ color: 'var(--text-muted)', fontSize: '0.75rem', marginTop: 8 }}>
              首次生成约需 20-40 秒，之后从缓存直接读取
            </div>
          </div>
        )}

        {/* 错误状态 */}
        {status === 'error' && (
          <div style={{ textAlign: 'center', padding: '40px 0' }}>
            <div style={{ color: '#e77', marginBottom: 16 }}>{errorMsg || '推算失败，请重试'}</div>
            <button
              onClick={startStream}
              style={{
                background: 'var(--wu-jin)', color: '#000', border: 'none',
                borderRadius: 8, padding: '10px 24px', cursor: 'pointer', fontWeight: 600,
              }}
            >重新推算</button>
          </div>
        )}

        {/* 结果：按大运分组时间轴 */}
        {status === 'done' && events.length > 0 && (
          <div>
            <div style={{ color: 'var(--text-muted)', fontSize: '0.75rem', marginBottom: 20, textAlign: 'center' }}>
              共推算 {events.length} 个年份 · 点击年份卡片查看详情
            </div>
            {Object.entries(grouped).map(([dayunGz, years]) => {
              const dyGan = dayunGz.length >= 1 ? dayunGz[0] : ''
              const dyWx = WUXING_GAN[dyGan] || 'tu'
              const dySum = summaryMap[dayunGz]
              return (
                <div key={dayunGz} style={{ marginBottom: 32 }}>
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
                    }}>{dayunGz}</div>
                    <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>大运</div>
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
                      {/* 主题标签行 */}
                      <div style={{ display: 'flex', flexWrap: 'wrap', gap: 6, marginBottom: 8 }}>
                        {dySum.themes.map((theme) => {
                          const meta = SIGNAL_LABEL[theme]
                          const color = meta ? meta.color : 'var(--text-muted)'
                          return (
                            <span
                              key={theme}
                              style={{
                                fontSize: '0.68rem',
                                padding: '2px 7px',
                                borderRadius: 4,
                                border: `1px solid ${color}`,
                                color,
                                whiteSpace: 'nowrap',
                              }}
                            >{meta ? meta.label : theme}</span>
                          )
                        })}
                      </div>
                      {/* 叙述段 */}
                      <div style={{
                        color: 'var(--text-secondary)',
                        fontSize: '0.83rem',
                        lineHeight: 1.75,
                      }}>
                        {dySum.summary}
                      </div>
                    </div>
                  )}

                  {/* 年份列表 */}
                  <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
                    {years.map((y) => {
                      const gan = y.gan_zhi?.[0] || ''
                      const wx = WUXING_GAN[gan] || 'tu'
                      const hasSignals = y.signals && y.signals.length > 0
                      return (
                        <div
                          key={y.year}
                          style={{
                            background: 'var(--bg-card)',
                            borderRadius: 12,
                            padding: '14px 16px',
                            border: hasSignals
                              ? `1px solid color-mix(in srgb, var(--wu-${wx}) 40%, transparent)`
                              : '1px solid var(--border-subtle)',
                          }}
                        >
                          {/* 年份行 */}
                          <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8 }}>
                            <span style={{
                              fontFamily: 'Noto Serif SC, serif',
                              fontWeight: 700,
                              fontSize: '1.05rem',
                              color: `var(--wu-${wx})`,
                            }}>{y.gan_zhi}</span>
                            <span style={{ color: 'var(--text-muted)', fontSize: '0.8rem' }}>
                              {y.year}年 · {y.age}岁
                            </span>
                            {/* 信号标签 */}
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
                          {/* 叙述文字 */}
                          <div style={{
                            color: 'var(--text-secondary)',
                            fontSize: '0.85rem',
                            lineHeight: 1.7,
                          }}>
                            {y.narrative}
                          </div>
                        </div>
                      )
                    })}
                  </div>
                </div>
              )
            })}

            {/* 免责 */}
            <div style={{
              marginTop: 24, padding: '12px 16px',
              background: 'var(--bg-elevated)',
              borderRadius: 8,
              fontSize: '0.72rem',
              color: 'var(--text-muted)',
              lineHeight: 1.6,
            }}>
              本推算基于八字命理算法与 AI 语言生成，仅供参考，不构成任何决策建议。
            </div>
          </div>
        )}

        {status === 'done' && events.length === 0 && (
          <div style={{ textAlign: 'center', padding: '40px 0', color: 'var(--text-muted)' }}>
            暂无过往年份数据
          </div>
        )}
      </div>
    </div>
  )
}
