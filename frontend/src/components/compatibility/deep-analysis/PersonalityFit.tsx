import './PersonalityFit.css'
import type { PersonalityFitSummary, PersonalityPoint } from '../../../lib/compatibilityPersonality'

function PersonalityPointList({ title, points }: { title: string; points: PersonalityPoint[] }) {
  return (
    <div className="compatibility-personality-list">
      <div className="compatibility-personality-list-title">{title}</div>
      {points.map(point => (
        <div key={`${title}-${point.title}-${point.evidenceKey || point.dimension || point.detail}`} className="compatibility-personality-point">
          <strong>{point.title}</strong>
          <p>{point.detail}</p>
          {point.evidenceKey && <a href="#compatibility-claim-evidence">查看性格依据</a>}
        </div>
      ))}
    </div>
  )
}

export default function PersonalityFit({ summary }: { summary: PersonalityFitSummary }) {
  return (
    <details open className="compat-da-personality" id="compatibility-personality-fit">
      <summary className="compat-da-subsection-summary">
        <span className="compat-da-subsection-title">性格相处画像</span>
        <span className="compat-da-subsection-hint">{summary.matchTypeDescription}</span>
      </summary>

      <div className="compat-da-personality__body">
        <h3 className="serif compat-da-personality__headline">{summary.headline}</h3>
        <p className="compat-da-personality__meta">当前问题：{summary.questionLabel} · 关系阶段：{summary.stageLabel}</p>
        <p className="compatibility-personality-summary">{summary.summary}</p>

        <div className="compatibility-personality-pattern-grid">
          <div className="compatibility-personality-pattern">
            <span>{summary.selfPattern.title}</span>
            <p>{summary.selfPattern.detail}</p>
          </div>
          <div className="compatibility-personality-pattern">
            <span>{summary.partnerPattern.title}</span>
            <p>{summary.partnerPattern.detail}</p>
          </div>
        </div>

        <div className="compatibility-personality-grid">
          <PersonalityPointList title="自然合的地方" points={summary.fitPoints} />
          <PersonalityPointList title="容易冲突的地方" points={summary.clashPoints} />
          <PersonalityPointList title="沟通建议" points={summary.communicationGuidance} />
        </div>

        <div className="compatibility-personality-note">
          <span>{summary.reportNote}</span>
          {summary.evidenceTargets.length > 0 && <a href="#compatibility-claim-evidence">查看性格判断依据</a>}
        </div>
      </div>
    </details>
  )
}
