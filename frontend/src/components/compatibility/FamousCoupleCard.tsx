import type { CompatibilityStructuredReport } from '../../lib/api'
import './FamousCoupleCard.css'

type Props = {
  hasReport: boolean
  famousCouple: CompatibilityStructuredReport['famous_couple']
  reportLoading: boolean
  onGenerateReport: () => void
}

export default function FamousCoupleCard({ hasReport, famousCouple, reportLoading, onGenerateReport }: Props) {
  // 状态三：已有报告但旧版没有 famous_couple → 隐藏，不误导
  if (hasReport && !famousCouple) {
    return null
  }

  // 状态一：还没生成深度解读 → 钩子占位 + 生成入口
  if (!hasReport) {
    return (
      <div className="famous-couple-card famous-couple-card--teaser">
        <div className="famous-couple-card__teaser-text">✨ 生成深度解读，揭晓你们的名人配对</div>
        <button
          type="button"
          className="btn btn-primary"
          onClick={onGenerateReport}
          disabled={reportLoading}
        >
          {reportLoading ? '生成中' : '生成深度解读'}
        </button>
      </div>
    )
  }

  // 状态二：已生成且有类比
  return (
    <div className="famous-couple-card famous-couple-card--filled">
      <div className="famous-couple-card__kicker">你们这对，像</div>
      <div className="serif famous-couple-card__couple">{famousCouple!.couple}</div>
      {famousCouple!.tagline && <div className="famous-couple-card__tagline">{famousCouple!.tagline}</div>}
      {famousCouple!.reason && <p className="famous-couple-card__reason">{famousCouple!.reason}</p>}
    </div>
  )
}
