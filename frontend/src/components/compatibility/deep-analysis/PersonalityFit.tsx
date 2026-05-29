import './PersonalityFit.css'
import type { ParticipantPortrait, PersonalityFitSummary, PersonalityPoint } from '../../../lib/compatibilityPersonality'

function PortraitCard({ portrait }: { portrait: ParticipantPortrait }) {
  return (
    <div className="compatibility-portrait-card">
      <div className="compatibility-portrait-headline">{portrait.headline}</div>
      <dl className="compatibility-portrait-dims">
        {portrait.dimensions.map(dim => (
          <div key={dim.key} className="compatibility-portrait-dim">
            <dt>{dim.label}</dt>
            <dd>{dim.detail}</dd>
          </div>
        ))}
      </dl>
    </div>
  )
}

function PersonalityPointList({ title, points }: { title: string; points: PersonalityPoint[] }) {
  return (
    <div className="compatibility-personality-list">
      <div className="compatibility-personality-list-title">{title}</div>
      {points.map(point => (
        <div key={`${title}-${point.title}-${point.evidenceKey || point.dimension || point.detail}`} className="compatibility-personality-point">
          <strong>{point.title}</strong>
          <p>{point.detail}</p>
        </div>
      ))}
    </div>
  )
}

export default function PersonalityFit({ summary }: { summary: PersonalityFitSummary }) {
  return (
    <section id="compatibility-personality-fit" className="compat-section-personality">
      <div className="compat-section-personality__head">
        <div className="compat-section-kicker">SECTION 02</div>
        <h2 className="serif compat-section-title">双方性格画像与差异</h2>
      </div>
      <p className="compat-da-personality__type-hint">{summary.matchTypeDescription}</p>

      <div className="compat-da-personality__body">
        <h3 className="serif compat-da-personality__headline">{summary.headline}</h3>
        <p className="compat-da-personality__meta">当前问题：{summary.questionLabel} · 关系阶段：{summary.stageLabel}</p>
        <p className="compatibility-personality-summary">{summary.summary}</p>

        <div className="compatibility-portrait-grid">
          <PortraitCard portrait={summary.selfPortrait} />
          <PortraitCard portrait={summary.partnerPortrait} />
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
    </section>
  )
}
