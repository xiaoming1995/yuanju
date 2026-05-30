import './DayPillarPortrait.css'
import type { CompatibilityParticipant } from '../../lib/api'
import { getDayPillarPortrait } from '../../lib/dayPillarPortraits'

function PortraitCard({ participant }: { participant?: CompatibilityParticipant | null }) {
  const snap = participant?.chart_snapshot
  const portrait = snap ? getDayPillarPortrait(snap.day_gan, snap.day_zhi) : undefined
  if (!snap || !portrait) return null
  return (
    <div className="day-pillar-card">
      <div className="day-pillar-card-head">
        <span className="day-pillar-name">{participant?.display_name || '—'}</span>
        <span className="day-pillar-ganzhi">{snap.day_gan}{snap.day_zhi}日</span>
      </div>
      <div className="day-pillar-tag">{portrait.tag}</div>
      <p className="day-pillar-text">{portrait.text}</p>
    </div>
  )
}

export default function DayPillarPortrait({
  self,
  partner,
}: {
  self?: CompatibilityParticipant | null
  partner?: CompatibilityParticipant | null
}) {
  const has = (p?: CompatibilityParticipant | null) =>
    Boolean(p?.chart_snapshot && getDayPillarPortrait(p.chart_snapshot.day_gan, p.chart_snapshot.day_zhi))
  if (!has(self) && !has(partner)) return null
  return (
    <section className="day-pillar-portrait">
      <div className="day-pillar-portrait-title serif">日柱 · 本命速写</div>
      <div className="day-pillar-grid">
        <PortraitCard participant={self} />
        <PortraitCard participant={partner} />
      </div>
    </section>
  )
}
