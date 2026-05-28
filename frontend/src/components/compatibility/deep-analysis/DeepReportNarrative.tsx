import './DeepReportNarrative.css'
import type {
  CompatibilityStructuredReport,
  CompatibilityQuestionFocus,
} from '../../../lib/api'

type Props = {
  hasReport: boolean
  structuredReport?: CompatibilityStructuredReport | null
  reportDimensions: CompatibilityStructuredReport['dimensions']
  reportRisks: string[]
  rawContent?: string
  error: string
  reportLoading: boolean
  onGenerateReport: () => void
}

function AdviceList({ title, items }: { title: string; items: string[] }) {
  const safeItems = Array.isArray(items) ? items.filter(Boolean) : []

  return (
    <div className="compatibility-advice-list">
      <div className="compatibility-advice-title">{title}</div>
      {safeItems.length > 0 ? (
        safeItems.map(item => <div key={item}>{item}</div>)
      ) : (
        <div>暂无明确建议</div>
      )}
    </div>
  )
}

function QuestionFocusPanel({ focus }: { focus?: CompatibilityQuestionFocus }) {
  if (!focus || (!focus.title && !focus.judgment)) return null
  const checks = Array.isArray(focus.key_checks) ? focus.key_checks.filter(Boolean) : []
  const boundaryConditions = Array.isArray(focus.boundary_conditions) ? focus.boundary_conditions.filter(Boolean) : []

  return (
    <div className="compatibility-question-focus">
      <div className="compatibility-consulting-kicker">{focus.title || '问题焦点'}</div>
      {focus.judgment && <p>{focus.judgment}</p>}
      <div className="compatibility-question-focus-grid">
        <AdviceList title="需要验证" items={checks} />
        <AdviceList title="边界条件" items={boundaryConditions} />
      </div>
    </div>
  )
}

export default function DeepReportNarrative({
  hasReport,
  structuredReport,
  reportDimensions,
  reportRisks,
  rawContent,
  error,
  reportLoading,
  onGenerateReport,
}: Props) {
  const reportStateClass = hasReport ? 'compatibility-ai-card--generated' : 'compatibility-ai-card--empty'
  return (
    <details open className="compat-da-report">
      <summary className="compat-da-subsection-summary">
        <span className="compat-da-subsection-title">AI 长文叙事</span>
        <span className="compat-da-subsection-hint">完整解读</span>
      </summary>
      <div className={`compat-da-report__body ${reportStateClass}`}>
        <div className="compatibility-ai-header">
          <div>
            <div className="compatibility-consulting-kicker">可选扩展</div>
            <h2 className="serif compatibility-section-title">深度解读</h2>
          </div>
          {!hasReport && (
            <button className="btn btn-primary compatibility-report-action" onClick={onGenerateReport} disabled={reportLoading}>
              {reportLoading ? '生成中' : '生成深度解读'}
            </button>
          )}
        </div>

        {error && (
          <div className="compatibility-report-state compatibility-report-state--error">
            {error}
          </div>
        )}

        {reportLoading && (
          <div className="compatibility-report-state">
            正在生成 AI 深度解读，请稍候。
          </div>
        )}

        {structuredReport ? (
          <div className="compatibility-report-content">
            <QuestionFocusPanel focus={structuredReport.question_focus} />
            <p className="compatibility-report-summary">{structuredReport.summary}</p>
            {reportDimensions.map(item => (
              <div key={item.key} className="compatibility-report-section">
                <div className="serif compatibility-report-title">{item.title}</div>
                <div className="compatibility-report-text">{item.content}</div>
              </div>
            ))}
            {reportRisks.length > 0 && (
              <div className="compatibility-report-section">
                <div className="serif compatibility-report-title">风险点</div>
                <ul className="compatibility-report-list">
                  {reportRisks.map(risk => <li key={risk}>{risk}</li>)}
                </ul>
              </div>
            )}
            <div className="compatibility-report-section">
              <div className="serif compatibility-report-title">建议</div>
              <div className="compatibility-report-text">{structuredReport.advice}</div>
            </div>
          </div>
        ) : rawContent ? (
          <div className="compatibility-report-raw">{rawContent}</div>
        ) : (
          <div className="compatibility-report-empty">
            <p>当前合盘结果已包含性格画像、冲突验证和关键依据。AI 深度解读会补充更完整的关系叙事、风险解释和相处建议。</p>
            <div className="compatibility-report-empty-grid">
              <span>关系叙事</span>
              <span>冲突解释</span>
              <span>相处建议</span>
            </div>
          </div>
        )}
      </div>
    </details>
  )
}
