import './SectionDeepAnalysis.css'
import ActionPlan7d30d from './deep-analysis/ActionPlan7d30d'
import NextStepsAndAvoid from './deep-analysis/NextStepsAndAvoid'
import type {
  CompatibilityDurationAssessment,
  CompatibilityStageRisk,
} from '../../lib/api'
import type {
  DecisionDashboardData,
} from '../../lib/compatibilityDecision'
import type {
  PersonalityValidationPlan,
} from '../../lib/compatibilityPersonality'

type Props = {
  personalityValidationPlan: PersonalityValidationPlan | null
  decisionStageRisks: CompatibilityStageRisk[]
  durationAssessment: CompatibilityDurationAssessment
  dashboard: DecisionDashboardData
}

export default function SectionDeepAnalysis({
  personalityValidationPlan, decisionStageRisks, durationAssessment,
  dashboard,
}: Props) {
  return (
    <section id="compat-section-deep-analysis" className="compat-section-deep-analysis">
      <div className="compat-section-deep-analysis__head">
        <div className="compat-section-kicker">SECTION 03</div>
        <h2 className="serif compat-section-title">深度分析</h2>
      </div>
      <div className="compat-section-deep-analysis__stack">
        <ActionPlan7d30d plan={personalityValidationPlan} risks={decisionStageRisks} assessment={durationAssessment} />
        <NextStepsAndAvoid dashboard={dashboard} />
      </div>
    </section>
  )
}
