import './NextStepsAndAvoid.css'
import type { DecisionDashboardData } from '../../../lib/compatibilityDecision'

export default function NextStepsAndAvoid({ dashboard }: { dashboard: DecisionDashboardData }) {
  return (
    <details open className="compat-da-nextsteps">
      <summary className="compat-da-subsection-summary">
        <span className="compat-da-subsection-title">下一步 / 避免 / 核心矛盾</span>
        <span className="compat-da-subsection-hint">短期具体动作</span>
      </summary>
      <div className="compat-da-nextsteps__body">
        <div className="compat-da-nextsteps__group">
          <span>下一步验证</span>
          <p>{dashboard.nextAction}</p>
        </div>
        {dashboard.avoid.length > 0 && (
          <div className="compat-da-nextsteps__group">
            <span>短期避免</span>
            <ul>
              {dashboard.avoid.map(item => <li key={item}>{item}</li>)}
            </ul>
          </div>
        )}
        <div className="compat-da-nextsteps__group">
          <span>核心矛盾</span>
          <p>{dashboard.summary}</p>
        </div>
      </div>
    </details>
  )
}
