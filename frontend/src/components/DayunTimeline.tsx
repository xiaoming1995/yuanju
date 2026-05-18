import { useEffect, useState } from 'react'
import LiuYueDrawer from './LiuYueDrawer'
import { fetchShenshaAnnotations, type ShenshaAnnotation } from '../lib/api'
import { buildDayunOverview } from '../lib/dayunOverview'

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
  dayGan: string
  gender?: string
  pillarsLabel?: string
  chartId?: string
  yongshen?: string
  jishen?: string
  wuxing?: { mu: number; huo: number; tu: number; jin: number; shui: number }
  tiaohou?: { expected: string[]; tou: string[]; cang: string[]; text: string } | null
}

const GAN_WUXING: Record<string, string> = {
  甲: 'mu', 乙: 'mu', 丙: 'huo', 丁: 'huo', 戊: 'tu',
  己: 'tu', 庚: 'jin', 辛: 'jin', 壬: 'shui', 癸: 'shui',
}

const WUXING_LABEL: Record<string, string> = {
  mu: '木',
  huo: '火',
  tu: '土',
  jin: '金',
  shui: '水',
}


const SS_POLARITY: Record<string, { bg: string; color: string }> = {
  ji: { bg: 'rgba(76,175,80,0.15)', color: '#66bb6a' },
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

function getGenderLabel(gender?: string) {
  if (gender === 'male') return '男'
  if (gender === 'female') return '女'
  return gender || '未填'
}


export default function DayunTimeline({
  dayun, birthYear, startYunSolar, dayGan, gender, pillarsLabel, chartId,
  yongshen, jishen, wuxing, tiaohou,
}: DayunTimelineProps) {
  const currentYear = new Date().getFullYear()
  const displayDayun = dayun.slice(0, 10)
  const currentDayunIndex = displayDayun.findIndex(d => currentYear >= d.start_year && currentYear <= d.end_year)
  const defaultActiveIndex = currentDayunIndex !== -1 ? currentDayunIndex : 0
  const [activeIndex, setActiveIndex] = useState(defaultActiveIndex)
  const resolvedActiveIndex = displayDayun[activeIndex] ? activeIndex : defaultActiveIndex
  const activeDayun = displayDayun[resolvedActiveIndex]

  const [drawerOpen, setDrawerOpen] = useState(false)
  const [showExpert, setShowExpert] = useState(false)
  const [drawerYear, setDrawerYear] = useState(currentYear)
  const [drawerGanZhi, setDrawerGanZhi] = useState('')
  const [ssAnnotations, setSsAnnotations] = useState<Record<string, ShenshaAnnotation>>({})
  const [ssModalOpen, setSsModalOpen] = useState(false)
  const [ssModalName, setSsModalName] = useState('')

  useEffect(() => {
    fetchShenshaAnnotations().then(list => {
      const map: Record<string, ShenshaAnnotation> = {}
      for (const item of list) map[item.name] = item
      setSsAnnotations(map)
    }).catch(() => {})
  }, [])

  const handleSsClick = (name: string) => {
    setSsModalName(name)
    setSsModalOpen(true)
  }

  const getLiuNianClassName = (ln: LiuNianItem, isLnCurrent: boolean, isFocusYear: boolean) => [
    'liunian-card',
    isLnCurrent ? 'is-liunian-current' : '',
    ln.is_transition ? 'is-liunian-transition' : '',
    isFocusYear ? 'is-liunian-focus' : '',
  ].filter(Boolean).join(' ')

  return (
    <div className="dayun-design-shell">
      <div className="dayun-mobile-topbar" aria-label="大运时间轴移动端标题栏">
        <span className="dayun-mobile-back">‹</span>
        <strong>大运时间轴</strong>
        <span className="dayun-mobile-action">⌘</span>
      </div>

      <div className="dayun-design-panel">
        <div className="dayun-design-header">
          <div className="dayun-heading-row">
            <span className="dayun-heading-mark" />
            <h2 className="serif">大运时间轴</h2>
          </div>
          <div className="dayun-meta-row">
            <span>起运：{startYunSolar || `${birthYear}年后起运`}</span>
            <span>性别：{getGenderLabel(gender)}</span>
            {pillarsLabel && <span>命局：{pillarsLabel}</span>}
          </div>
        </div>

        <div className="dayun-timeline-container">
          <div className="dayun-overview-grid">
            {displayDayun.map((d, i) => {
              const isCurrent = currentYear >= d.start_year && currentYear <= d.end_year
              const isActive = i === resolvedActiveIndex
              const wx = GAN_WUXING[d.gan] || 'jin'
              return (
                <button
                  key={d.index}
                  type="button"
                  className={`dayun-step-card${isActive ? ' is-active' : ''}${isCurrent ? ' is-current' : ''}`}
                  onClick={() => setActiveIndex(i)}
                >
                  {isCurrent && <span className="dayun-current-badge">当前</span>}
                  <span className="dayun-step-index">{d.index}</span>
                  <span className="dayun-step-age">{d.start_age}岁-{d.start_age + 9}岁</span>
                  <span className="dayun-step-ganzhi">
                    <span className={`wuxing-text-${wx}`}>{d.gan}</span>
                    <span>{d.zhi}</span>
                  </span>
                  <span className="dayun-step-ten-god">{d.gan_shishen}</span>

                  {d.shen_sha && d.shen_sha.length > 0 && (
                    <span className="dayun-shensha-list">
                      {d.shen_sha.map((ss, si) => {
                        const pol = SS_POLARITY_MAP[ss] || 'zhong'
                        const sty = SS_POLARITY[pol] || SS_POLARITY.zhong
                        return (
                          <span
                            key={si}
                            className="dayun-shensha-tag"
                            onClick={(e) => { e.stopPropagation(); handleSsClick(ss) }}
                            style={{ background: sty.bg, color: sty.color }}
                            onMouseEnter={e => (e.currentTarget.style.transform = 'scale(1.08)')}
                            onMouseLeave={e => (e.currentTarget.style.transform = 'scale(1)')}
                          >{ss}</span>
                        )
                      })}
                    </span>
                  )}

                  <span className="dayun-year-range">
                    <span>{d.start_year}</span>
                    <span>-</span>
                    <span>{d.end_year}</span>
                  </span>
                  {isActive && <span className="dayun-step-caret" />}
                </button>
              )
            })}
          </div>

          {activeDayun && activeDayun.liu_nian && activeDayun.liu_nian.length > 0 && (
            <div className="liunian-panel animate-fade-up">
              <div className="liunian-panel-header">
                <div>
                  <div className="liunian-panel-title-row">
                    <h3>{activeDayun.gan}{activeDayun.zhi}大运流年</h3>
                  </div>
                  <p className="liunian-panel-subtitle">
                    {activeDayun.start_age}岁-{activeDayun.start_age + 9}岁（公历 {activeDayun.start_year}-{activeDayun.end_year}）
                  </p>
                </div>
                <div className="liunian-panel-legend" aria-label="流年标记说明">
                  <span><i className="legend-dot legend-dot--current" /> 当前年</span>
                  <span><i className="legend-dot legend-dot--transition" /> 交脱年</span>
                  <span><i className="legend-dot legend-dot--focus" /> 重点年</span>
                </div>
              </div>

              <div className="liunian-grid">
                {activeDayun.liu_nian.slice(0, 10).map((ln) => {
                  const lnGan = ln.gan_zhi.charAt(0)
                  const lnZhi = ln.gan_zhi.charAt(1)
                  const lnWx = GAN_WUXING[lnGan] || 'jin'
                  const isLnCurrent = currentYear === ln.year
                  const isDayunCurrent = currentYear >= activeDayun.start_year && currentYear <= activeDayun.end_year
                  const isTransitionYear = Boolean(ln.is_transition)
                  const isFocusYear = isLnCurrent || isTransitionYear

                  return (
                    <button
                      key={ln.year}
                      type="button"
                      className={getLiuNianClassName(ln, isDayunCurrent && isLnCurrent, isFocusYear)}
                      onClick={() => {
                        setDrawerYear(ln.year)
                        setDrawerGanZhi(ln.gan_zhi)
                        setDrawerOpen(true)
                      }}
                    >
                      <span className="liunian-card-badges">
                        {isDayunCurrent && isLnCurrent && <span className="liunian-current-badge">当前</span>}
                        {ln.is_transition && (
                          <span
                            className="liunian-transition-ribbon"
                            title={ln.trans_month && ln.trans_day ? `${ln.trans_month}月${ln.trans_day}日交脱` : '交脱年'}
                            aria-label={ln.trans_month && ln.trans_day ? `${ln.trans_month}月${ln.trans_day}日交脱` : '交脱年'}
                          >交脱</span>
                        )}
                        {isFocusYear && !isLnCurrent && !ln.is_transition && <span className="liunian-focus-badge">重点</span>}
                      </span>

                      <span className="liunian-card-topline">
                        <span>{ln.year}年</span>
                        <span>{ln.age}岁</span>
                      </span>
                      <span className="liunian-ganzhi">
                        <span className={`wuxing-text-${lnWx}`}>{lnGan}</span>
                        <span>{lnZhi}</span>
                      </span>
                      <span className="liunian-ten-god-line">
                        <strong>{ln.gan_shishen}</strong>
                        <span>{ln.zhi_shishen}</span>
                      </span>
                      <span className="liunian-card-divider" />
                      <span className="liunian-open-cue">
                        {ln.is_transition && ln.prev_dayun ? `承 ${ln.prev_dayun}` : '查看流月'}
                      </span>
                    </button>
                  )
                })}
              </div>
            </div>
          )}

          {activeDayun && (() => {
            const dayGanWx = GAN_WUXING[dayGan] ?? ''
            const dayGanWxCn = WUXING_LABEL[dayGanWx] ?? ''
            const overview = buildDayunOverview({
              dayun: {
                gan: activeDayun.gan,
                zhi: activeDayun.zhi,
                gan_shishen: activeDayun.gan_shishen,
                zhi_shishen: activeDayun.zhi_shishen,
                di_shi: activeDayun.di_shi,
              },
              yongshen: yongshen ?? '',
              jishen: jishen ?? '',
              wuxing: wuxing ?? { mu: 20, huo: 20, tu: 20, jin: 20, shui: 20 },
              dayGanWuxing: dayGanWxCn,
              tiaohou: tiaohou ?? null,
            })
            return (
              <div className="dayun-summary-strip">
                <div className="dayun-summary-copy">
                  <strong>大运总览</strong>
                  <span className="dayun-summary-body">
                    <span>{overview.proseLay}</span>
                    <button
                      type="button"
                      className="dayun-summary-toggle"
                      onClick={() => setShowExpert(s => !s)}
                    >
                      {showExpert ? '收起专业表述 ‹' : '查看专业表述 ›'}
                    </button>
                    {showExpert && (
                      <span className="dayun-summary-expert">{overview.prose}</span>
                    )}
                  </span>
                </div>
                <div className="dayun-summary-tags">
                  <span>十神主气：{activeDayun.gan_shishen}</span>
                  <span>五行主气：{WUXING_LABEL[GAN_WUXING[activeDayun.gan] || 'jin']}</span>
                  <span>趋势关键词：{overview.trendKeywords}</span>
                </div>
              </div>
            )
          })()}

          <p className="dayun-disclaimer">以上为中国传统命理分析，仅供参考，理性看待。</p>
        </div>
      </div>

      <LiuYueDrawer
        open={drawerOpen}
        onClose={() => setDrawerOpen(false)}
        initialYear={drawerYear}
        dayGan={dayGan}
        liuNianGanZhi={drawerGanZhi}
        chartId={chartId}
      />

      {ssModalOpen && (() => {
        const ann = ssAnnotations[ssModalName]
        const pol = SS_POLARITY_MAP[ssModalName] || 'zhong'
        const sty = SS_POLARITY[pol] || SS_POLARITY.zhong
        const polarityLabel = pol === 'ji' ? '吉神' : pol === 'xiong' ? '凶煞' : '中性'
        return (
          <div className="shensha-modal-backdrop" onClick={() => setSsModalOpen(false)}>
            <div
              className="shensha-modal-card"
              onClick={e => e.stopPropagation()}
              style={{ borderColor: `${sty.color}55`, boxShadow: `0 20px 60px rgba(0,0,0,0.5), 0 0 30px ${sty.color}15` }}
            >
              <div className="shensha-modal-header">
                <span style={{ color: sty.color }}>{ssModalName}</span>
                <small style={{ background: sty.bg, color: sty.color }}>{polarityLabel}</small>
              </div>
              <div className="shensha-modal-body">
                {ann?.description || '暂无此神煞的详细注解。'}
              </div>
              <div className="shensha-modal-footer">
                <button type="button" onClick={() => setSsModalOpen(false)}>关闭</button>
              </div>
            </div>
          </div>
        )
      })()}
    </div>
  )
}
