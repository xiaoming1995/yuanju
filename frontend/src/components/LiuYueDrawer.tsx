import { useState, useEffect, useCallback } from 'react'
import { baziAPI } from '../lib/api'
import type { LiuYueItem, LiuYueResponse } from '../lib/api'

interface LiuYueDrawerProps {
  open: boolean
  onClose: () => void
  initialYear: number
  dayGan: string
  liuNianGanZhi: string // 例如 "丙午"，用于标题
  chartId?: string
}

const GAN_WUXING: Record<string, string> = {
  甲: 'mu', 乙: 'mu', 丙: 'huo', 丁: 'huo', 戊: 'tu',
  己: 'tu', 庚: 'jin', 辛: 'jin', 壬: 'shui', 癸: 'shui',
}

// 格式化日期 YYYY-MM-DD → M/D
function fmtDate(d: string): string {
  const parts = d.split('-')
  if (parts.length !== 3) return d
  return `${parseInt(parts[1])}/${parseInt(parts[2])}`
}

export default function LiuYueDrawer({
  open,
  onClose,
  initialYear,
  dayGan,
  liuNianGanZhi,
  chartId,
}: LiuYueDrawerProps) {
  const [year, setYear] = useState(initialYear)
  const [data, setData] = useState<LiuYueResponse | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // 流年 AI 报告状态
  const [report, setReport] = useState<any>(null)
  const [reportLoading, setReportLoading] = useState(false)
  const [reportError, setReportError] = useState<string | null>(null)

  const fetchData = useCallback(async (y: number) => {
    setLoading(true)
    setError(null)
    try {
      const res = await baziAPI.fetchLiuYue(y, dayGan)
      setData(res.data)
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : '加载失败')
    } finally {
      setLoading(false)
    }
  }, [dayGan])

  // 每次打开或年份变化时重新请求
  useEffect(() => {
    if (open) {
      setYear(initialYear)
      fetchData(initialYear)
    }
  }, [open, initialYear, fetchData])

  useEffect(() => {
    if (open) fetchData(year)
    // 换年份时清空已有的报告
    setReport(null)
    setReportError(null)
  }, [year]) // eslint-disable-line react-hooks/exhaustive-deps

  const handleGenerateReport = async () => {
    if (!chartId) {
      setReportError('需登录且在专属排盘页方可生成AI报告')
      return
    }
    setReportLoading(true)
    setReportError(null)
    try {
      const { data } = await baziAPI.generateLiunianReport(chartId, year)
      setReport(data.report)
    } catch (e: any) {
      setReportError(e.message || '生成失败')
    } finally {
      setReportLoading(false)
    }
  }

  // 防止 body 滚动
  useEffect(() => {
    if (open) {
      document.body.style.overflow = 'hidden'
    } else {
      document.body.style.overflow = ''
    }
    return () => { document.body.style.overflow = '' }
  }, [open])

  if (!open) return null

  const items: LiuYueItem[] = data?.liu_yue ?? []
  const currentIndex = data?.current_month_index ?? -1

  return (
    <>
      {/* 遮罩层 */}
      <div
        onClick={onClose}
        style={{
          position: 'fixed', inset: 0,
          background: 'rgba(0,0,0,0.6)',
          backdropFilter: 'blur(4px)',
          zIndex: 1000,
          animation: 'fadeIn 0.2s ease',
        }}
      />

      {/* 抽屉主体 */}
      <div style={{
        position: 'fixed',
        top: 0,
        right: 0,
        bottom: 0,
        width: 'min(520px, 100vw)',
        background: 'var(--bg-card)',
        borderLeft: '1px solid var(--border-default)',
        zIndex: 1001,
        display: 'flex',
        flexDirection: 'column',
        animation: 'slideInRight 0.25s cubic-bezier(0.32,0.72,0,1)',
        boxShadow: '-8px 0 40px rgba(0,0,0,0.4)',
      }}>

        {/* 抽屉头部 */}
        <div style={{
          padding: '20px 24px 16px',
          borderBottom: '1px solid var(--border-subtle)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          flexShrink: 0,
        }}>
          <div>
            <div style={{ fontSize: 18, fontWeight: 700, color: 'var(--text-accent)', fontFamily: 'Noto Serif SC, serif' }}>
              {liuNianGanZhi}年 · 流月详情
            </div>
            <div style={{ fontSize: 12, color: 'var(--text-muted)', marginTop: 4 }}>
              十二月令 · 节气干支 · 十神
            </div>
          </div>

          <button
            onClick={onClose}
            style={{
              width: 32, height: 32, borderRadius: '50%',
              border: '1px solid var(--border-default)',
              background: 'var(--bg-elevated)',
              color: 'var(--text-secondary)',
              cursor: 'pointer',
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              fontSize: 16,
              transition: 'all 0.15s',
            }}
            onMouseEnter={e => (e.currentTarget.style.background = 'var(--bg-card-hover)')}
            onMouseLeave={e => (e.currentTarget.style.background = 'var(--bg-elevated)')}
          >
            ✕
          </button>
        </div>

        {/* 年份切换器 */}
        <div style={{
          padding: '14px 24px',
          borderBottom: '1px solid var(--border-subtle)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          gap: 20,
          flexShrink: 0,
        }}>
          <button
            onClick={() => setYear(y => y - 1)}
            style={navBtnStyle}
            onMouseEnter={e => (e.currentTarget.style.background = 'var(--bg-card-hover)')}
            onMouseLeave={e => (e.currentTarget.style.background = 'var(--bg-elevated)')}
          >
            ←
          </button>

          <div style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 20, fontWeight: 700, color: 'var(--text-primary)' }}>{year}</div>
            <div style={{ fontSize: 11, color: 'var(--text-muted)' }}>公历年</div>
          </div>

          <button
            onClick={() => setYear(y => y + 1)}
            style={navBtnStyle}
            onMouseEnter={e => (e.currentTarget.style.background = 'var(--bg-card-hover)')}
            onMouseLeave={e => (e.currentTarget.style.background = 'var(--bg-elevated)')}
          >
            →
          </button>
        </div>

        {/* 内容区 */}
        <div style={{ flex: 1, overflowY: 'auto', padding: '20px 24px' }}>

          {/* AI 流年批断区域 */}
          <div style={{
            background: 'rgba(201, 168, 76, 0.05)',
            border: '1px solid rgba(201, 168, 76, 0.2)',
            borderRadius: '12px',
            padding: '16px',
            marginBottom: '20px',
            position: 'relative',
            overflow: 'hidden'
          }}>
            <h3 style={{ margin: '0 0 12px 0', fontSize: '15px', color: 'var(--color-primary)', display: 'flex', alignItems: 'center', gap: '8px' }}>
              {year}年运势精批
            </h3>
            
            {!report && !reportLoading && (
              <div style={{ textAlign: 'center', padding: '10px 0' }}>
                <p style={{ fontSize: '13px', color: 'var(--text-secondary)', margin: '0 0 12px 0' }}>
                  结合原局喜忌，详细推演本年运势全景
                </p>
                <button
                  onClick={handleGenerateReport}
                  className="btn btn-primary btn-sm"
                  style={{ width: '100%' }}
                >
                  生成流年分析
                </button>
              </div>
            )}
            
            {reportLoading && (
              <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '10px', padding: '20px 0' }}>
                 <div className="spinner" style={{ width: 24, height: 24, borderColor: 'var(--color-primary) transparent var(--color-primary) transparent' }}></div>
                 <div style={{ fontSize: '13px', color: 'var(--text-secondary)' }}>正在结合原局深度测算中...</div>
              </div>
            )}
            
            {reportError && (
              <div style={{ fontSize: '13px', color: '#ff4d4f', textAlign: 'center', marginTop: '10px' }}>
                ⚠ {reportError}
              </div>
            )}

            {report && report.content_structured && (
              <div style={{ fontSize: '13.5px', lineHeight: 1.6, color: 'var(--text-primary)' }}>
                <div style={{ marginBottom: '16px' }}>
                  <div style={{ fontSize: '12px', color: 'var(--text-secondary)', marginBottom: '4px' }}>事业财运</div>
                  <div>{report.content_structured.career}</div>
                </div>
                <div style={{ marginBottom: '16px' }}>
                  <div style={{ fontSize: '12px', color: 'var(--text-secondary)', marginBottom: '4px' }}>感情桃花</div>
                  <div>{report.content_structured.romance}</div>
                </div>
                <div style={{ marginBottom: '16px' }}>
                  <div style={{ fontSize: '12px', color: 'var(--text-secondary)', marginBottom: '4px' }}>健康风险</div>
                  <div>{report.content_structured.health}</div>
                </div>
                <div style={{ background: 'rgba(255,255,255,0.05)', padding: '10px', borderRadius: '8px', borderLeft: '3px solid var(--color-primary)' }}>
                  <div style={{ fontSize: '12px', color: 'var(--color-primary)', marginBottom: '2px', fontWeight: 'bold' }}>年度锦囊</div>
                  <div style={{ fontStyle: 'italic' }}>{report.content_structured.advice}</div>
                </div>
              </div>
            )}
          </div>

          {/* 加载中：骨架屏 */}
          {loading && (
            <div style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(3, 1fr)',
              gap: 10,
            }}>
              {Array.from({ length: 12 }).map((_, i) => (
                <div key={i} style={{
                  height: 112,
                  borderRadius: 'var(--radius-sm)',
                  background: 'var(--bg-elevated)',
                  animation: `skeleton-pulse 1.4s ease-in-out ${(i % 3) * 0.1}s infinite`,
                }} />
              ))}
            </div>
          )}

          {/* 错误状态 */}
          {!loading && error && (
            <div style={{
              textAlign: 'center', padding: '40px 20px',
              color: 'var(--text-muted)',
            }}>
              <div style={{ fontSize: 32, marginBottom: 12 }}>⚠</div>
              <div style={{ marginBottom: 16 }}>{error}</div>
              <button
                onClick={() => fetchData(year)}
                style={{
                  padding: '8px 20px',
                  border: '1px solid var(--border-default)',
                  borderRadius: 'var(--radius-sm)',
                  background: 'var(--bg-elevated)',
                  color: 'var(--text-primary)',
                  cursor: 'pointer',
                  fontSize: 13,
                }}
              >
                重试
              </button>
            </div>
          )}

          {/* 流月网格 */}
          {!loading && !error && items.length > 0 && (
            <div style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(3, 1fr)',
              gap: 10,
            }}>
              {items.map(item => {
                const isCurrent = item.index === currentIndex
                const gan = item.gan_zhi.charAt(0)
                const zhi = item.gan_zhi.charAt(1)
                const wx = GAN_WUXING[gan] || 'jin'

                return (
                  <div
                    key={item.index}
                    style={{
                      position: 'relative',
                      padding: '12px 8px 10px',
                      background: isCurrent ? 'rgba(201,168,76,0.08)' : 'var(--bg-elevated)',
                      border: `1px solid ${isCurrent ? 'var(--border-accent)' : 'var(--border-subtle)'}`,
                      borderRadius: 'var(--radius-sm)',
                      display: 'flex',
                      flexDirection: 'column',
                      alignItems: 'center',
                      gap: 3,
                      boxShadow: isCurrent ? '0 0 12px rgba(201,168,76,0.12)' : 'none',
                      transition: 'all 0.15s',
                    }}
                  >
                    {/* 当前徽章 */}
                    {isCurrent && (
                      <div style={{
                        position: 'absolute', top: -9, left: '50%', transform: 'translateX(-50%)',
                        fontSize: 9, background: 'var(--wu-jin)', color: '#0d0f14',
                        padding: '1px 7px', borderRadius: 99, fontWeight: 700,
                        whiteSpace: 'nowrap', letterSpacing: 0.5,
                      }}>当前</div>
                    )}

                    {/* 节气名 */}
                    <div style={{ fontSize: 10, color: 'var(--text-accent)', fontWeight: 600, letterSpacing: 1 }}>
                      {item.jie_qi_name}
                    </div>

                    {/* 节气起止日期 */}
                    <div style={{ fontSize: 9, color: 'var(--text-muted)', lineHeight: 1.3, textAlign: 'center' }}>
                      {fmtDate(item.start_date)}—{fmtDate(item.end_date)}
                    </div>

                    {/* 月令名 */}
                    <div style={{ fontSize: 10, color: 'var(--text-secondary)', marginTop: 1 }}>
                      {item.month_name}
                    </div>

                    {/* 干支大字 */}
                    <div style={{ display: 'flex', gap: 1, fontFamily: 'Noto Serif SC, serif', fontWeight: 700, marginTop: 2 }}>
                      <span className={`wuxing-text-${wx}`} style={{ fontSize: '1.3rem' }}>{gan}</span>
                      <span style={{ fontSize: '1.3rem', color: 'var(--text-primary)' }}>{zhi}</span>
                    </div>

                    {/* 十神 */}
                    <div style={{ fontSize: 9, color: 'var(--text-secondary)', marginTop: 1 }}>
                      {item.gan_shishen} · {item.zhi_shishen}
                    </div>
                  </div>
                )
              })}
            </div>
          )}
        </div>
      </div>

      <style>{`
        @keyframes fadeIn {
          from { opacity: 0 }
          to   { opacity: 1 }
        }
        @keyframes slideInRight {
          from { transform: translateX(100%) }
          to   { transform: translateX(0) }
        }
        @keyframes skeleton-pulse {
          0%, 100% { opacity: 0.4 }
          50%       { opacity: 0.8 }
        }
        @media (max-width: 768px) {
          /* 移动端：底部全屏 sheet */
          .liuyue-drawer-body {
            top: 30% !important;
            right: 0 !important;
            bottom: 0 !important;
            left: 0 !important;
            width: 100% !important;
            border-radius: 16px 16px 0 0 !important;
            border-left: none !important;
            border-top: 1px solid var(--border-default) !important;
          }
        }
      `}</style>
    </>
  )
}

const navBtnStyle: React.CSSProperties = {
  width: 36, height: 36, borderRadius: 'var(--radius-sm)',
  border: '1px solid var(--border-default)',
  background: 'var(--bg-elevated)',
  color: 'var(--text-primary)',
  cursor: 'pointer',
  fontSize: 16,
  display: 'flex', alignItems: 'center', justifyContent: 'center',
  transition: 'background 0.15s',
}
