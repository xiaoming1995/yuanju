import { useState, useEffect } from 'react'
import LiuYueDrawer from './LiuYueDrawer'
import { fetchShenshaAnnotations, type ShenshaAnnotation } from '../lib/api'

interface LiuNianItem {
  year: number
  age: number
  gan_zhi: string
  gan_shishen: string
  zhi_shishen: string
  is_transition?: boolean
  trans_month?: number
  trans_day?: number
  prev_dayun?: string
}

interface JinBuHuanResult {
  qian_level: string
  qian_desc: string
  hou_level: string
  hou_desc: string
  verse: string
}

interface DayunItem {
  index: number
  gan: string
  zhi: string
  start_age: number
  start_year: number
  end_year: number
  gan_shishen: string
  zhi_shishen: string
  di_shi: string
  shen_sha?: string[]
  jin_bu_huan?: JinBuHuanResult | null
  liu_nian: LiuNianItem[]
}

interface DayunTimelineProps {
  dayun: DayunItem[]
  birthYear: number
  startYunSolar: string
  dayGan: string // 命主日主天干，用于流月十神计算
  chartId?: string
}

const GAN_WUXING: Record<string, string> = {
  甲: 'mu', 乙: 'mu', 丙: 'huo', 丁: 'huo', 戊: 'tu',
  己: 'tu', 庚: 'jin', 辛: 'jin', 壬: 'shui', 癸: 'shui',
}

// 神煞极性配色
const SS_POLARITY: Record<string, { bg: string; color: string }> = {
  ji:    { bg: 'rgba(76,175,80,0.15)', color: '#66bb6a' },
  xiong: { bg: 'rgba(244,67,54,0.15)', color: '#ef5350' },
  zhong: { bg: 'rgba(255,193,7,0.12)', color: '#ffc107' },
}
const SS_POLARITY_MAP: Record<string, string> = {
  天乙贵人: 'ji', 太极贵人: 'ji', 文昌贵人: 'ji', 禄神: 'ji',
  天德贵人: 'ji', 月德贵人: 'ji', 天德合: 'ji', 月德合: 'ji',
  德秀贵人: 'ji', 金舆贵人: 'ji', 天喜: 'ji', 天厨贵人: 'ji',
  国印贵人: 'ji', 三奇贵人: 'ji', 日德: 'ji', 将星: 'ji',
  十灵日: 'ji', 词馆: 'ji', 福星贵人: 'ji', 天医: 'ji',
  羊刃: 'xiong', 飞刃: 'xiong', 劫煞: 'xiong', 亡神: 'xiong',
  孤辰: 'xiong', 寡宿: 'xiong', 阴差阳错: 'xiong', 魁罡: 'xiong',
  十恶大败: 'xiong', 天罗地网: 'xiong', 地网: 'xiong', 童子煞: 'xiong',
  灾煞: 'xiong', 流霞: 'xiong', 吊客: 'xiong', 墓门: 'xiong',
  桃花: 'zhong', 驿马: 'zhong', 华盖: 'zhong', 红艳: 'zhong',
}

export default function DayunTimeline({ dayun, startYunSolar, dayGan, chartId }: DayunTimelineProps) {
  const currentYear = new Date().getFullYear()
  
  // 找出当前年份所在的大运索引
  const currentDayunIndex = dayun.findIndex(d => currentYear >= d.start_year && currentYear <= d.end_year)
  const defaultActiveIndex = currentDayunIndex !== -1 ? currentDayunIndex : 0
  const [activeIndex, setActiveIndex] = useState(defaultActiveIndex)
  const resolvedActiveIndex = dayun[activeIndex] ? activeIndex : defaultActiveIndex

  // 流月抽屉状态
  const [drawerOpen, setDrawerOpen] = useState(false)
  const [drawerYear, setDrawerYear] = useState(currentYear)
  const [drawerGanZhi, setDrawerGanZhi] = useState('')

  const activeDayun = dayun[resolvedActiveIndex]
  // 神煞注解数据
  const [ssAnnotations, setSsAnnotations] = useState<Record<string, ShenshaAnnotation>>({})
  const [ssModalOpen, setSsModalOpen] = useState(false)
  const [ssModalName, setSsModalName] = useState('')

  useEffect(() => {
    fetchShenshaAnnotations().then(list => {
      const map: Record<string, ShenshaAnnotation> = {}
      for (const a of list) map[a.name] = a
      setSsAnnotations(map)
    }).catch(() => {})
  }, [])

  const handleSsClick = (name: string) => {
    setSsModalName(name)
    setSsModalOpen(true)
  }

  return (
    <div className="dayun-timeline-container">
      {startYunSolar && (
        <div style={{ textAlign: 'center', marginBottom: 16, color: 'var(--text-muted)', fontSize: 13, letterSpacing: 1 }}>
          精确交运时间：{startYunSolar}
        </div>
      )}
      
      <div style={{ overflowX: 'auto', paddingTop: 12, paddingBottom: 16, scrollBehavior: 'smooth' }}>
        <div style={{ display: 'flex', gap: 12, minWidth: 'max-content' }}>
          {dayun.map((d, i) => {
            const isCurrent = currentYear >= d.start_year && currentYear <= d.end_year
            const isActive = i === resolvedActiveIndex
            const wx = GAN_WUXING[d.gan] || 'jin'
            return (
              <div
                key={d.index}
                onClick={() => setActiveIndex(i)}
                style={{
                  display: 'flex',
                  flexDirection: 'column',
                  alignItems: 'center',
                  gap: 6,
                  padding: '16px 20px',
                  background: isActive ? 'var(--bg-card-hover)' : 'var(--bg-elevated)',
                  border: `1px solid ${isActive ? 'var(--border-accent)' : 'var(--border-subtle)'}`,
                  borderRadius: 'var(--radius-md)',
                  minWidth: 88,
                  position: 'relative',
                  cursor: 'pointer',
                  transition: 'all 0.2s',
                  boxShadow: isActive ? '0 0 15px rgba(201,168,76,0.1)' : 'none',
                }}
              >
                {isCurrent && (
                  <div style={{
                    position: 'absolute', top: -10, left: '50%', transform: 'translateX(-50%)',
                    fontSize: 10, background: 'var(--wu-jin)', color: '#0d0f14',
                    padding: '2px 8px', borderRadius: 99, fontWeight: 700, whiteSpace: 'nowrap',
                  }}>当前</div>
                )}
                
                <div style={{ fontSize: 11, color: 'var(--text-muted)' }}>{d.start_age}岁起</div>
                
                {/* 大运头顶的十神 */}
                <div style={{ fontSize: 10, color: 'var(--text-secondary)' }}>{d.gan_shishen}</div>
                
                <div style={{ display: 'flex', gap: 2, fontFamily: 'Noto Serif SC, serif', fontWeight: 700 }}>
                  <span className={`wuxing-text-${wx}`} style={{ fontSize: '1.4rem' }}>{d.gan}</span>
                  <span style={{ fontSize: '1.4rem', color: 'var(--text-primary)' }}>{d.zhi}</span>
                </div>
                
                {/* 大运脚底的十神及长生 */}
                <div style={{ display: 'flex', gap: 4, fontSize: 10, color: 'var(--text-secondary)' }}>
                  <span>{d.zhi_shishen}</span>
                  <span style={{color: 'var(--wu-jin)'}}>{d.di_shi}</span>
                </div>
                

                {/* 大运神煞标签 */}
                {d.shen_sha && d.shen_sha.length > 0 && (
                  <div style={{ display: 'flex', flexWrap: 'wrap', gap: 3, justifyContent: 'center', maxWidth: 90, marginTop: 2 }}>
                    {d.shen_sha.map((ss, si) => {
                      const pol = SS_POLARITY_MAP[ss] || 'zhong'
                      const sty = SS_POLARITY[pol] || SS_POLARITY.zhong
                      return (
                        <span
                          key={si}
                          onClick={(e) => { e.stopPropagation(); handleSsClick(ss) }}
                          style={{
                            fontSize: 9, padding: '1px 4px', borderRadius: 3,
                            background: sty.bg, color: sty.color,
                            lineHeight: 1.4, whiteSpace: 'nowrap',
                            cursor: 'pointer',
                            transition: 'transform 0.15s',
                          }}
                          onMouseEnter={e => (e.currentTarget.style.transform = 'scale(1.15)')}
                          onMouseLeave={e => (e.currentTarget.style.transform = 'scale(1)')}
                        >{ss}</span>
                      )
                    })}
                  </div>
                )}

                <div style={{ fontSize: 11, color: 'var(--text-muted)', textAlign: 'center', lineHeight: 1.5, marginTop: 4 }}>
                  {d.start_year}<br/>—<br/>{d.end_year}
                </div>
                
                {/* 选中态的下巴小箭头 */}
                {isActive && (
                  <div style={{
                    position: 'absolute', bottom: -8, left: '50%', transform: 'translateX(-50%) rotate(45deg)',
                    width: 14, height: 14, background: 'var(--bg-card-hover)',
                    borderRight: '1px solid var(--border-accent)',
                    borderBottom: '1px solid var(--border-accent)',
                  }}></div>
                )}
              </div>
            )
          })}
        </div>
      </div>

      {/* 流年网格面板 */}
      {activeDayun && activeDayun.liu_nian && activeDayun.liu_nian.length > 0 && (
        <div className="liunian-panel animate-fade-up" style={{
          marginTop: 8,
          padding: 20,
          background: 'rgba(20, 23, 32, 0.6)',
          backdropFilter: 'blur(10px)',
          border: '1px solid var(--border-default)',
          borderRadius: 'var(--radius-md)',
        }}>
          <div style={{ fontSize: 14, fontWeight: 600, color: 'var(--text-accent)', marginBottom: 16 }}>
            {activeDayun.gan}{activeDayun.zhi}大运流年
          </div>
          
          <div className="liunian-grid">
            {activeDayun.liu_nian.map((ln) => {
              const lnGan = ln.gan_zhi.charAt(0);
              const lnWx = GAN_WUXING[lnGan] || 'jin';
              const isLnCurrent = currentYear === ln.year;
              const isDayunCurrent = currentYear >= activeDayun.start_year && currentYear <= activeDayun.end_year;
              
              return (
                <div
                  key={ln.year}
                  onClick={() => {
                    setDrawerYear(ln.year)
                    setDrawerGanZhi(ln.gan_zhi)
                    setDrawerOpen(true)
                  }}
                  style={{
                    background: isLnCurrent ? 'rgba(201,168,76,0.1)' : 'var(--bg-elevated)',
                    border: `1px solid ${isLnCurrent ? 'var(--border-accent)' : 'var(--border-subtle)'}`,
                    borderRadius: 'var(--radius-sm)',
                    padding: ln.is_transition ? '0' : '12px 8px',
                    display: 'flex',
                    flexDirection: 'column',
                    alignItems: 'center',
                    position: 'relative',
                    cursor: 'pointer',
                    transition: 'all 0.15s',
                    overflow: 'hidden',
                  }}
                  onMouseEnter={e => {
                    e.currentTarget.style.border = '1px solid var(--border-accent)'
                    e.currentTarget.style.background = 'var(--bg-card-hover)'
                  }}
                  onMouseLeave={e => {
                    e.currentTarget.style.border = `1px solid ${isLnCurrent ? 'var(--border-accent)' : 'var(--border-subtle)'}`
                    e.currentTarget.style.background = isLnCurrent ? 'rgba(201,168,76,0.1)' : 'var(--bg-elevated)'
                  }}
                >
                  {isDayunCurrent && isLnCurrent && <div style={{width: 6, height: 6, borderRadius: '50%', backgroundColor: 'var(--wu-jin)', position: 'absolute', top: 6, right: 6, zIndex: 10}}></div>}
                  
                  {ln.is_transition ? (
                    <>
                      {/* 旧运区域 (上半部分) */}
                      <div style={{ width: '100%', padding: '8px 8px 6px', background: 'rgba(255,255,255,0.03)', display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
                        <div style={{ fontSize: 9, color: 'var(--text-muted)' }}>{ln.prev_dayun ? `[旧] ${ln.prev_dayun}` : '未起运'}</div>
                      </div>
                      
                      {/* 分割与交脱提示 */}
                      <div style={{ width: '100%', borderTop: '1px dashed var(--border-subtle)', position: 'relative', display: 'flex', justifyContent: 'center' }}>
                        <div style={{
                          position: 'absolute', top: -8, background: 'var(--bg-elevated)', padding: '0 4px', fontSize: 9,
                          color: 'var(--wu-jin)', borderRadius: 2
                        }}>
                          {ln.trans_month}月{ln.trans_day}日交脱
                        </div>
                      </div>
                      
                      {/* 主体流年区域 (下半部分) */}
                      <div style={{ width: '100%', padding: '12px 8px 8px', display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 4 }}>
                        <div style={{ fontSize: 10, color: 'var(--text-muted)' }}>{ln.year} ({ln.age}岁)</div>
                        <div style={{ fontSize: 10, color: 'var(--text-secondary)' }}>{ln.gan_shishen}</div>
                        <div style={{ fontFamily: 'Noto Serif SC, serif', fontWeight: 600, fontSize: '1.2rem', margin: '2px 0' }}>
                          <span className={`wuxing-text-${lnWx}`}>{lnGan}</span>
                          <span style={{ color: 'var(--text-primary)' }}>{ln.gan_zhi.charAt(1)}</span>
                        </div>
                        <div style={{ fontSize: 10, color: 'var(--text-secondary)' }}>{ln.zhi_shishen}</div>
                      </div>
                    </>
                  ) : (
                    <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 4 }}>
                      <div style={{ fontSize: 10, color: 'var(--text-muted)' }}>{ln.year} ({ln.age}岁)</div>
                      <div style={{ fontSize: 10, color: 'var(--text-secondary)' }}>{ln.gan_shishen}</div>
                      <div style={{ fontFamily: 'Noto Serif SC, serif', fontWeight: 600, fontSize: '1.2rem', margin: '2px 0' }}>
                        <span className={`wuxing-text-${lnWx}`}>{lnGan}</span>
                        <span style={{ color: 'var(--text-primary)' }}>{ln.gan_zhi.charAt(1)}</span>
                      </div>
                      <div style={{ fontSize: 10, color: 'var(--text-secondary)' }}>{ln.zhi_shishen}</div>
                    </div>
                  )}
                </div>
              )
            })}
          </div>
        </div>
      )}

      {/* 流月抽屉 */}
      <LiuYueDrawer
        open={drawerOpen}
        onClose={() => setDrawerOpen(false)}
        initialYear={drawerYear}
        dayGan={dayGan}
        liuNianGanZhi={drawerGanZhi}
        chartId={chartId}
      />

      {/* 神煞注解弹窗 */}
      {ssModalOpen && (() => {
        const ann = ssAnnotations[ssModalName]
        const pol = SS_POLARITY_MAP[ssModalName] || 'zhong'
        const sty = SS_POLARITY[pol] || SS_POLARITY.zhong
        const polarityLabel = pol === 'ji' ? '吉神' : pol === 'xiong' ? '凶煞' : '中性'
        return (
          <div
            onClick={() => setSsModalOpen(false)}
            style={{
              position: 'fixed', inset: 0, zIndex: 9999,
              background: 'rgba(0,0,0,0.65)',
              backdropFilter: 'blur(6px)',
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              padding: 20,
              animation: 'fadeIn 0.2s ease',
            }}
          >
            <div
              onClick={e => e.stopPropagation()}
              style={{
                background: 'var(--bg-card, #1a1d2e)',
                border: `1px solid ${sty.color}33`,
                borderRadius: 16,
                maxWidth: 440,
                width: '100%',
                maxHeight: '80vh',
                overflow: 'auto',
                padding: '28px 24px',
                boxShadow: `0 20px 60px rgba(0,0,0,0.5), 0 0 30px ${sty.color}15`,
                animation: 'slideUp 0.25s ease',
              }}
            >
              {/* 头部 */}
              <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 16 }}>
                <span style={{
                  fontSize: 28, fontFamily: 'Noto Serif SC, serif', fontWeight: 700,
                  color: sty.color,
                }}>{ssModalName}</span>
                <span style={{
                  fontSize: 11, padding: '3px 10px', borderRadius: 99,
                  background: sty.bg, color: sty.color, fontWeight: 600,
                }}>{polarityLabel}</span>
              </div>

              {/* 正文 */}
              <div style={{
                fontSize: 14, lineHeight: 1.9, color: 'var(--text-secondary, #b0b3c0)',
                whiteSpace: 'pre-wrap',
              }}>
                {ann?.description || '暂无此神煞的详细注解。'}
              </div>

              {/* 关闭按钮 */}
              <div style={{ textAlign: 'center', marginTop: 20 }}>
                <button
                  onClick={() => setSsModalOpen(false)}
                  style={{
                    background: 'transparent', border: `1px solid var(--border-default, #333)`,
                    color: 'var(--text-muted, #666)', padding: '8px 28px',
                    borderRadius: 8, cursor: 'pointer', fontSize: 13,
                  }}
                >关闭</button>
              </div>
            </div>
          </div>
        )
      })()}
    </div>
  )
}
