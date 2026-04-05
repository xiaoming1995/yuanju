import { useState, useEffect } from 'react'
import LiuYueDrawer from './LiuYueDrawer'

interface LiuNianItem {
  year: number
  age: number
  gan_zhi: string
  gan_shishen: string
  zhi_shishen: string
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
  liu_nian: LiuNianItem[]
}

interface DayunTimelineProps {
  dayun: DayunItem[]
  birthYear: number
  startYunSolar: string
  dayGan: string // 命主日主天干，用于流月十神计算
}

const GAN_WUXING: Record<string, string> = {
  甲: 'mu', 乙: 'mu', 丙: 'huo', 丁: 'huo', 戊: 'tu',
  己: 'tu', 庚: 'jin', 辛: 'jin', 壬: 'shui', 癸: 'shui',
}

export default function DayunTimeline({ dayun, startYunSolar, dayGan }: DayunTimelineProps) {
  const currentYear = new Date().getFullYear()
  
  // 找出当前年份所在的大运索引
  const initialActiveIndex = dayun.findIndex(d => currentYear >= d.start_year && currentYear <= d.end_year)
  const [activeIndex, setActiveIndex] = useState(initialActiveIndex !== -1 ? initialActiveIndex : 0)

  // 流月抽屉状态
  const [drawerOpen, setDrawerOpen] = useState(false)
  const [drawerYear, setDrawerYear] = useState(currentYear)
  const [drawerGanZhi, setDrawerGanZhi] = useState('')

  useEffect(() => {
    const idx = dayun.findIndex(d => currentYear >= d.start_year && currentYear <= d.end_year)
    setActiveIndex(idx !== -1 ? idx : 0)
  }, [dayun, currentYear])

  const activeDayun = dayun[activeIndex]

  return (
    <div className="dayun-timeline-container">
      {startYunSolar && (
        <div style={{ textAlign: 'center', marginBottom: 16, color: 'var(--text-muted)', fontSize: 13, letterSpacing: 1 }}>
          ✦ 精确交运时间：{startYunSolar} ✦
        </div>
      )}
      
      <div style={{ overflowX: 'auto', paddingTop: 12, paddingBottom: 16, scrollBehavior: 'smooth' }}>
        <div style={{ display: 'flex', gap: 12, minWidth: 'max-content' }}>
          {dayun.map((d, i) => {
            const isCurrent = currentYear >= d.start_year && currentYear <= d.end_year
            const isActive = i === activeIndex
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
          <div style={{ fontSize: 14, fontWeight: 600, color: 'var(--text-accent)', marginBottom: 16, display: 'flex', alignItems: 'center', gap: 8 }}>
            <span>✦</span>
            <span>{activeDayun.gan}{activeDayun.zhi}大运流年</span>
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
                    // 点击流年格弹出流月抽屉
                    setDrawerYear(ln.year)
                    setDrawerGanZhi(ln.gan_zhi)
                    setDrawerOpen(true)
                  }}
                  style={{
                    background: isLnCurrent ? 'rgba(201,168,76,0.1)' : 'var(--bg-elevated)',
                    border: `1px solid ${isLnCurrent ? 'var(--border-accent)' : 'var(--border-subtle)'}`,
                    borderRadius: 'var(--radius-sm)',
                    padding: '12px 8px',
                    display: 'flex',
                    flexDirection: 'column',
                    alignItems: 'center',
                    gap: 4,
                    position: 'relative',
                    cursor: 'pointer',
                    transition: 'all 0.15s',
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
                  {isDayunCurrent && isLnCurrent && <div style={{width: 6, height: 6, borderRadius: '50%', backgroundColor: 'var(--wu-jin)', position: 'absolute', top: 6, right: 6}}></div>}
                  <div style={{ fontSize: 10, color: 'var(--text-muted)' }}>{ln.year} ({ln.age}岁)</div>
                  <div style={{ fontSize: 10, color: 'var(--text-secondary)' }}>{ln.gan_shishen}</div>
                  <div style={{ fontFamily: 'Noto Serif SC, serif', fontWeight: 600, fontSize: '1.2rem', margin: '2px 0' }}>
                    <span className={`wuxing-text-${lnWx}`}>{lnGan}</span>
                    <span style={{ color: 'var(--text-primary)' }}>{ln.gan_zhi.charAt(1)}</span>
                  </div>
                  <div style={{ fontSize: 10, color: 'var(--text-secondary)' }}>{ln.zhi_shishen}</div>
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
      />
    </div>
  )
}
