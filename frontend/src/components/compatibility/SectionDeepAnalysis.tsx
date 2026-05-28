import './SectionDeepAnalysis.css'
import PersonalityFit from './deep-analysis/PersonalityFit'
import ActionPlan7d30d from './deep-analysis/ActionPlan7d30d'
import RelationshipStrategy from './deep-analysis/RelationshipStrategy'
import NextStepsAndAvoid from './deep-analysis/NextStepsAndAvoid'
import DeepReportNarrative from './deep-analysis/DeepReportNarrative'
import type {
  CompatibilityDurationAssessment,
  CompatibilityStageRisk,
  CompatibilityRelationshipStrategy,
  CompatibilityStructuredReport,
} from '../../lib/api'
import type {
  DecisionDashboardData,
} from '../../lib/compatibilityDecision'
import type {
  PersonalityFitSummary,
  PersonalityValidationPlan,
} from '../../lib/compatibilityPersonality'

type Props = {
  personalitySummary: PersonalityFitSummary | null
  personalityValidationPlan: PersonalityValidationPlan | null
  decisionStageRisks: CompatibilityStageRisk[]
  durationAssessment: CompatibilityDurationAssessment
  relationshipStrategy?: CompatibilityRelationshipStrategy
  dashboard: DecisionDashboardData
  deepReport: {
    hasReport: boolean
    structuredReport?: CompatibilityStructuredReport | null
    reportDimensions: CompatibilityStructuredReport['dimensions']
    reportRisks: string[]
    rawContent?: string
    error: string
    reportLoading: boolean
    onGenerateReport: () => void
  }
}

export default function SectionDeepAnalysis({
  personalitySummary, personalityValidationPlan, decisionStageRisks, durationAssessment,
  relationshipStrategy, dashboard, deepReport,
}: Props) {
  return (
    <section id="compat-section-deep-analysis" className="compat-section-deep-analysis">
      <div className="compat-section-deep-analysis__head">
        <div className="compat-section-kicker">SECTION 03</div>
        <h2 className="serif compat-section-title">AI 深度分析</h2>
      </div>
      <div className="compat-section-deep-analysis__stack">
        {personalitySummary && <PersonalityFit summary={personalitySummary} />}
        <ActionPlan7d30d plan={personalityValidationPlan} risks={decisionStageRisks} assessment={durationAssessment} />
        {relationshipStrategy && <RelationshipStrategy strategy={relationshipStrategy} />}
        <NextStepsAndAvoid dashboard={dashboard} />
        <DeepReportNarrative {...deepReport} />
      </div>
    </section>
  )
}
