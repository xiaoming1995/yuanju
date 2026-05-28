import './SectionBasicCharts.css'
import ParticipantSummaryCard from './ParticipantSummaryCard'
import type { CompatibilityParticipant } from '../../lib/api'

type Props = {
  self?: CompatibilityParticipant | null
  partner?: CompatibilityParticipant | null
}

export default function SectionBasicCharts({ self, partner }: Props) {
  return (
    <section id="compat-section-basic-charts" className="compat-section-basic-charts">
      <div className="compat-section-basic-charts__head">
        <div className="compat-section-kicker">SECTION 01</div>
        <h2 className="serif compat-section-title">双方基础盘</h2>
      </div>
      <div className="compat-section-basic-charts__grid">
        {self && <ParticipantSummaryCard participant={self} />}
        {partner && <ParticipantSummaryCard participant={partner} />}
      </div>
    </section>
  )
}
