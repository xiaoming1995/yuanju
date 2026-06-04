import './DeepReportNarrative.css'
import FamousCoupleCard from '../FamousCoupleCard'
import PersonalityComparison from './PersonalityComparison'
import RelationshipStrategy from './RelationshipStrategy'
import type {
  CompatibilityStructuredReport,
  CompatibilityQuestionFocus,
  CompatibilityRelationshipStrategy,
} from '../../../lib/api'

type Props = {
  hasReport: boolean
  structuredReport?: CompatibilityStructuredReport | null
  relationshipStrategy?: CompatibilityRelationshipStrategy
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
  relationshipStrategy,
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
        <span className="compat-da-subsection-title">AI 深度解读</span>
        <span className="compat-da-subsection-hint">{hasReport ? '性格画像 · 叙事 · 维度 · 风险 · 建议' : '可选扩展，点击下方按钮生成'}</span>
      </summary>
      <div className={`compat-da-report__body ${reportStateClass}`}>
        {!hasReport && (
          <button className="btn btn-primary compatibility-report-action" onClick={onGenerateReport} disabled={reportLoading}>
            {reportLoading ? '生成中…' : '生成深度解读'}
          </button>
        )}

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
            <FamousCoupleCard famousCouple={structuredReport.famous_couple} />
            <QuestionFocusPanel focus={structuredReport.question_focus} />
            <div className="compatibility-report-section">
              <div className="serif compatibility-report-title">总体判断</div>
              <p className="compatibility-report-summary">{structuredReport.summary}</p>
            </div>
            <PersonalityComparison comparison={structuredReport.personality_comparison} />
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
            {relationshipStrategy && (
              <RelationshipStrategy strategy={relationshipStrategy} />
            )}
          </div>
        ) : rawContent ? (
          <div className="compatibility-report-raw">{rawContent}</div>
        ) : (
          <div className="compatibility-report-empty">
            <p>AI 深度解读会基于双方命盘生成「双方性格画像与差异」，以及完整的关系叙事、风险解释和相处建议。</p>
            <div className="compatibility-report-empty-grid">
              <span>双方性格画像与差异</span>
              <span>关系叙事</span>
              <span>相处建议</span>
            </div>
          </div>
        )}
      </div>
    </details>
  )
}
