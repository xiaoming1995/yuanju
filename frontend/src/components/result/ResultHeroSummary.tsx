import '../ui/StatusBadge.css'
import './ResultHeroSummary.css'

interface ResultPillarSummary {
  label: string
  gan: string
  zhi: string
  ganWx?: string
  zhiWx?: string
}

interface ResultHeroSummaryProps {
  birthYear: number
  birthMonth: number
  birthDay: number
  birthHour: number
  gender: string
  pillars: ResultPillarSummary[]
  yongshen?: string
  jishen?: string
  mingGe?: string
  reportReady?: boolean
  reportLoading?: boolean
  onMingGeClick?: () => void
}

function buildResultCoreConclusion({
  dayGan,
  dayZhi,
  yongshen,
  jishen,
  reportReady,
}: {
  dayGan?: string
  dayZhi?: string
  yongshen?: string
  jishen?: string
  reportReady?: boolean
}) {
  if (reportReady) {
    return 'AI 解读已生成，建议先读摘要结论，再展开专业依据。'
  }
  const dayPillar = dayGan && dayZhi ? `${dayGan}${dayZhi}` : '日柱'
  if (yongshen && jishen) {
    return `此盘以${dayPillar}为命主参照，先抓「${yongshen}」为调衡重点，对「${jishen}」相关倾向保持克制。`
  }
  return `${dayPillar}已排出，待生成 AI 解读后可获得完整性格、事业、感情与健康判断。`
}

export function ResultHeroSummary({
  birthYear,
  birthMonth,
  birthDay,
  birthHour,
  gender,
  pillars,
  yongshen,
  jishen,
  mingGe,
  reportReady = false,
  reportLoading = false,
  onMingGeClick,
}: ResultHeroSummaryProps) {
  const dayPillar = pillars.find((pillar) => pillar.label === '日柱')
  const conclusion = buildResultCoreConclusion({
    dayGan: dayPillar?.gan,
    dayZhi: dayPillar?.zhi,
    yongshen,
    jishen,
    reportReady,
  })

  return (
    <section className="result-hero-summary animate-fade-up" aria-labelledby="result-hero-title">
      <div className="result-hero-summary__main">
        <p className="result-hero-summary__eyebrow">排盘结果总览</p>
        <h1 id="result-hero-title" className="result-hero-summary__title serif">
          先看结论，再看依据
        </h1>
        <p className="result-core-conclusion">{conclusion}</p>
        <div className="result-hero-summary__meta">
          <span>{birthYear}年{birthMonth}月{birthDay}日 {birthHour}时</span>
          <span>{gender === 'male' ? '男命' : '女命'}</span>
          <span>{reportLoading ? 'AI 解读生成中' : reportReady ? 'AI 解读已生成' : '待生成 AI 解读'}</span>
        </div>
        <div className="result-hero-summary__badges">
          <span className="ui-status-badge ui-status-badge--warning">喜用：{yongshen || '待判定'}</span>
          <span className="ui-status-badge ui-status-badge--danger">忌：{jishen || '待判定'}</span>
          {mingGe && (
            <button type="button" className="result-hero-summary__mingge" onClick={onMingGeClick}>
              {mingGe}
            </button>
          )}
        </div>
      </div>

      <div className="result-hero-summary__pillars" aria-label="四柱总览">
        {pillars.map((pillar) => (
          <div key={pillar.label} className={`result-hero-summary__pillar${pillar.label === '日柱' ? ' is-day-pillar' : ''}`}>
            <span>{pillar.label}</span>
            <strong className="serif">{pillar.gan}{pillar.zhi}</strong>
            <em>{pillar.ganWx || '-'} / {pillar.zhiWx || '-'}</em>
          </div>
        ))}
      </div>
    </section>
  )
}
