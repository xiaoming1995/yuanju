import './ActionPlan7d30d.css'
import type { CompatibilityDurationAssessment, CompatibilityStageRisk } from '../../../lib/api'
import type { PersonalityValidationPlan } from '../../../lib/compatibilityPersonality'

const durationLevelText: Record<string, string> = {
  high: '偏高',
  medium: '中等',
  low: '偏低',
}

const durationLevelClass: Record<string, string> = {
  high: 'compatibility-duration-task-item--high',
  medium: 'compatibility-duration-task-item--medium',
  low: 'compatibility-duration-task-item--low',
}

const stageWindowText: Record<string, string> = {
  three_months: '3 个月',
  one_year: '1 年',
  two_years_plus: '2 年以上',
}

function StageRiskGrid({ risks }: { risks: CompatibilityStageRisk[] }) {
  return (
    <div className="compatibility-stage-validation-grid">
      {risks.map(risk => (
        <div key={risk.window} className="card compatibility-stage-validation-card">
          <div className="compatibility-stage-window">{stageWindowText[risk.window] || risk.window}</div>
          <div className="compatibility-stage-label">要验证什么</div>
          <div className="serif compatibility-stage-risk">{risk.main_risk}</div>
          <p><span>触发场景：</span>{risk.trigger}</p>
          <div className="compatibility-stage-advice">{risk.advice}</div>
        </div>
      ))}
    </div>
  )
}

function DurationTaskSummary({ assessment }: { assessment: CompatibilityDurationAssessment }) {
  const windows = [
    { key: 'three_months', label: '3 个月', level: assessment.windows.three_months.level },
    { key: 'one_year', label: '1 年', level: assessment.windows.one_year.level },
    { key: 'two_years_plus', label: '2 年以上', level: assessment.windows.two_years_plus.level },
  ]

  return (
    <div className="compatibility-duration-task">
      <div className="compatibility-duration-task-grid">
        {windows.map(window => (
          <div
            key={window.key}
            className={`compatibility-duration-task-item ${durationLevelClass[window.level] || ''}`.trim()}
          >
            <span>{window.label}</span>
            <strong className="serif">{durationLevelText[window.level] || window.level}</strong>
          </div>
        ))}
      </div>
      {assessment.summary && <p>{assessment.summary}</p>}
      {assessment.reasons.length > 0 && (
        <div className="compatibility-duration-reasons">
          {assessment.reasons.map(reason => (
            <div key={reason} className="compatibility-duration-reason">{reason}</div>
          ))}
        </div>
      )}
    </div>
  )
}

export default function ActionPlan7d30d({
  plan,
  risks,
  assessment,
}: {
  plan: PersonalityValidationPlan | null | undefined
  risks: CompatibilityStageRisk[]
  assessment: CompatibilityDurationAssessment
}) {
  if (!plan) {
    return (
      <details open className="compat-da-actionplan">
        <summary className="compat-da-subsection-summary">
          <span className="compat-da-subsection-title">阶段风险与时段</span>
          <span className="compat-da-subsection-hint">分阶段查看主要风险点和时段强弱</span>
        </summary>
        <div className="compat-da-actionplan__body">
          <StageRiskGrid risks={risks} />
          <DurationTaskSummary assessment={assessment} />
        </div>
      </details>
    )
  }

  const groups: Array<{ title: string; items: string[]; anchor?: string }> = [
    plan.shortTerm,
    plan.mediumTerm,
    plan.avoid,
  ]

  return (
    <details open className="compat-da-actionplan">
      <summary className="compat-da-subsection-summary">
        <span className="compat-da-subsection-title">性格验证计划</span>
        <span className="compat-da-subsection-hint">7 天观察 / 30 天复盘</span>
      </summary>
      <div className="compat-da-actionplan__body">
        <div className="compatibility-validation-plan-grid">
          {groups.map(group => (
            <div key={group.title} className="card compatibility-validation-plan-card">
              <div className="compatibility-validation-plan-title">{group.title}</div>
              {group.items.map(item => <p key={item}>{item}</p>)}
              {group.anchor && <a href={group.anchor}>查看阶段验证</a>}
            </div>
          ))}
        </div>
        <p className="compatibility-validation-plan-note">{plan.supportNote}</p>
        <div className="compatibility-validation-detail">
          <div className="compatibility-validation-detail-heading">阶段风险明细</div>
          <StageRiskGrid risks={risks} />
          <DurationTaskSummary assessment={assessment} />
        </div>
      </div>
    </details>
  )
}
