import type { CompatibilityStructuredReport } from '../../lib/api'
import './FamousCoupleCard.css'

type Props = {
  famousCouple: CompatibilityStructuredReport['famous_couple']
}

// 名人配对类比，作为 AI 深度解读的开场，与其它 LLM 内容放在一起。
export default function FamousCoupleCard({ famousCouple }: Props) {
  if (!famousCouple) return null

  return (
    <div className="famous-couple-card famous-couple-card--filled">
      <div className="famous-couple-card__kicker">你们这对，像</div>
      <div className="serif famous-couple-card__couple">{famousCouple.couple}</div>
      {famousCouple.tagline && <div className="famous-couple-card__tagline">{famousCouple.tagline}</div>}
      {famousCouple.reason && <p className="famous-couple-card__reason">{famousCouple.reason}</p>}
    </div>
  )
}
