import './SectionDeepAnalysis.css'
import ActionPlan7d30d from './deep-analysis/ActionPlan7d30d'
import RelationshipStrategy from './deep-analysis/RelationshipStrategy'
import NextStepsAndAvoid from './deep-analysis/NextStepsAndAvoid'
import type {
  CompatibilityDurationAssessment,
  CompatibilityRelationshipStrategy,
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
  relationshipStrategy?: CompatibilityRelationshipStrategy
  dashboard: DecisionDashboardData
}

export default function SectionDeepAnalysis({
  personalityValidationPlan, decisionStageRisks, durationAssessment,
  relationshipStrategy,
  dashboard,
}: Props) {
  return (
    <section id="compat-section-deep-analysis" className="compat-section-deep-analysis">
      <div className="compat-section-deep-analysis__head">
        <div className="compat-section-kicker">SECTION 03</div>
        <h2 className="serif compat-section-title">深度分析</h2>
      </div>
      <div className="compat-section-deep-analysis__stack">
        {relationshipStrategy && <RelationshipStrategy strategy={relationshipStrategy} />}
        <ActionPlan7d30d plan={personalityValidationPlan} risks={decisionStageRisks} assessment={durationAssessment} />
        <NextStepsAndAvoid dashboard={dashboard} />
      </div>
    </section>
  )
}
