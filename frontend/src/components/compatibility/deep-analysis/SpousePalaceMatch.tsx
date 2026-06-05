import './SpousePalaceMatch.css'
import type {
  CompatibilitySpousePalaceMatch,
  CompatibilitySpousePalaceSide,
} from '../../../lib/api'

// 契合度 label 由前端按 level 固定映射（防 LLM 标签漂移）
const MATCH_LEVEL_TEXT: Record<string, string> = {
  high: '高',
  medium: '中',
  low: '低',
}

function SideCard({ name, side }: { name: string; side?: CompatibilitySpousePalaceSide }) {
  if (!side) return null
  const fit = Array.isArray(side.fit_points) ? side.fit_points.filter(Boolean) : []
  const gap = Array.isArray(side.gap_points) ? side.gap_points.filter(Boolean) : []
  const levelText = MATCH_LEVEL_TEXT[side.match_level]
  return (
    <div className="spouse-match-card">
      <div className="spouse-match-card-head">
        <span className="spouse-match-card-title">{name}理想的另一半</span>
        {levelText && (
          <span className={`spouse-match-badge spouse-match-badge--${side.match_level}`}>契合 {levelText}</span>
        )}
      </div>
      {side.ideal_portrait && <p className="spouse-match-portrait">{side.ideal_portrait}</p>}
      {fit.length > 0 && (
        <div className="spouse-match-list">
          <div className="spouse-match-list-title">对上了</div>
          {fit.map((text, index) => (
            <p key={`fit-${index}`} className="spouse-match-point">✓ {text}</p>
          ))}
        </div>
      )}
      {gap.length > 0 && (
        <div className="spouse-match-list">
          <div className="spouse-match-list-title">有差距</div>
          {gap.map((text, index) => (
            <p key={`gap-${index}`} className="spouse-match-point">✗ {text}</p>
          ))}
        </div>
      )}
    </div>
  )
}

export default function SpousePalaceMatch({
  match,
  selfName,
  partnerName,
}: {
  match?: CompatibilitySpousePalaceMatch | null
  selfName: string
  partnerName: string
}) {
  if (!match || (!match.self && !match.partner)) return null
  return (
    <div className="compatibility-report-section spouse-match">
      <div className="serif compatibility-report-title">夫妻宫匹配</div>
      <div className="spouse-match-grid">
        <SideCard name={selfName} side={match.self} />
        <SideCard name={partnerName} side={match.partner} />
      </div>
      {match.summary && <p className="spouse-match-summary">{match.summary}</p>}
    </div>
  )
}
