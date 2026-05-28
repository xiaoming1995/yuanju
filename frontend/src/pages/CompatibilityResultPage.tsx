import { useCallback, useEffect, useRef, useState, type ReactNode } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { HeartHandshake } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'
import {
  compatibilityAPI,
  isV3DimensionScores,
  type CompatibilityClaimEvidenceLink,
  type CompatibilityDetail,
  type CompatibilityDimensionScoresLegacy,
  type CompatibilityDimensionScoresV3,
  type CompatibilityDurationAssessment,
  type CompatibilityEvidence,
  type CompatibilityQuestionFocus,
  type CompatibilityRelationshipStrategy,
  type CompatibilityStageRisk,
  type CompatibilityStructuredReport,
} from '../lib/api'
import {
  buildDecisionDashboardData,
  buildDecisionStageRisks,
  hasLinkedEvidence,
  type DecisionDashboardData,
  type DecisionFinding,
} from '../lib/compatibilityDecision'
import {
  buildPersonalityFitSummary,
  buildPersonalityValidationPlan,
  type PersonalityFitSummary,
  type PersonalityPoint,
  type PersonalityValidationPlan,
} from '../lib/compatibilityPersonality'
import { toBlob, toPng } from 'html-to-image'
import jsPDF from 'jspdf'
import html2canvas from 'html2canvas'
import { brandAPI, type ExportBrand } from '../lib/api'
import CompatibilityShareCard from '../components/CompatibilityShareCard'
import CompatibilityPrintLayout from '../components/CompatibilityPrintLayout'
import SectionBasicCharts from '../components/compatibility/SectionBasicCharts'
import SectionVerdict from '../components/compatibility/SectionVerdict'
import SectionDeepAnalysis from '../components/compatibility/SectionDeepAnalysis'
import ParticipantSummaryCard from '../components/compatibility/ParticipantSummaryCard'
import './CompatibilityResultPage.css'

// Feature flag: 控制重构期间新结构 vs 旧 11 段结构。批次 4 完成时删除。
const ENABLE_NEW_LAYOUT = false

const dimensionText: Record<string, string> = {
  attraction: '会不会互相吸引？',
  stability: '能不能长期稳定？',
  communication: '吵架后能不能修复？',
  practicality: '现实条件能不能落地？',
  zodiac: '合属相',
  nayin: '合纳音',
  day_pillar: '合日柱',
  eight_chars: '合八字',
}

const dimensionHint: Record<keyof CompatibilityDimensionScoresLegacy, string> = {
  attraction: '初期靠近感与彼此牵引',
  stability: '长期承接与持续投入',
  communication: '冲突后的理解和修复',
  practicality: '现实安排、责任和节奏',
}

const polarityText: Record<string, string> = {
  positive: '正向',
  negative: '风险',
  mixed: '复杂',
  neutral: '中性',
}

const polarityColor: Record<string, string> = {
  positive: '#66bb6a',
  negative: '#ef5350',
  mixed: '#ffb74d',
  neutral: 'var(--text-muted)',
}

const evidenceSourceText: Record<string, string> = {
  day_master: '日主关系',
  five_elements: '五行结构',
  spouse_palace: '夫妻宫',
  spouse_star: '配偶星',
  ganzhi: '冲克总量',
  shensha: '神煞辅助',
  ten_god_interaction: '十神互动',
  favorable_element_support: '喜忌互补',
  ganzhi_interaction: '干支合冲刑害',
  relationship_pattern: '关系模式',
  timing_context: '阶段时机',
  zodiac: '合属相',
  nayin: '合纳音',
  day_pillar: '合日柱',
  eight_chars: '合八字',
}

const perspectiveText: Record<string, string> = {
  self_to_partner: '我看对方',
  partner_to_self: '对方看我',
  mutual: '双方互见',
}

const durationLevelText: Record<string, string> = {
  high: '偏高',
  medium: '中等',
  low: '偏低',
}

const stageWindowText: Record<string, string> = {
  three_months: '3 个月',
  one_year: '1 年',
  two_years_plus: '2 年以上',
}

const relationshipStageText: Record<string, string> = {
  ambiguous: '暧昧中',
  dating: '恋爱中',
  long_distance: '异地中',
  reconciliation: '分手/复合中',
  marriage_or_engagement: '谈婚论嫁',
  crush: '单恋/暗恋',
  general: '综合关系判断',
}

const primaryQuestionText: Record<string, string> = {
  continue_investment: '值不值得继续投入',
  marriage_suitability: '适不适合结婚',
  recurring_conflict: '为什么反复拉扯',
  reconciliation_potential: '复合有没有意义',
  long_term_stability: '长期能不能稳定',
  relationship_strategy: '怎么相处更顺',
  general: '综合关系判断',
}

function isDurationLevel(value: unknown): value is 'high' | 'medium' | 'low' {
  return value === 'high' || value === 'medium' || value === 'low'
}

function normalizeDurationAssessment(
  primary: CompatibilityDurationAssessment | null | undefined,
  fallback: CompatibilityDurationAssessment
): CompatibilityDurationAssessment {
  const hasPrimaryValue = Boolean(primary && (
    (typeof primary.summary === 'string' && primary.summary.trim()) ||
    (Array.isArray(primary.reasons) && primary.reasons.length > 0) ||
    isDurationLevel(primary.windows?.three_months?.level) ||
    isDurationLevel(primary.windows?.one_year?.level) ||
    isDurationLevel(primary.windows?.two_years_plus?.level)
  ))

  const source = hasPrimaryValue ? primary! : fallback

  return {
    overall_band: source.overall_band || fallback.overall_band,
    summary: source.summary?.trim() || fallback.summary,
    reasons: Array.isArray(source.reasons) ? source.reasons.filter(Boolean) : fallback.reasons,
    windows: {
      three_months: {
        level: isDurationLevel(source.windows?.three_months?.level) ? source.windows.three_months.level : fallback.windows.three_months.level,
      },
      one_year: {
        level: isDurationLevel(source.windows?.one_year?.level) ? source.windows.one_year.level : fallback.windows.one_year.level,
      },
      two_years_plus: {
        level: isDurationLevel(source.windows?.two_years_plus?.level) ? source.windows.two_years_plus.level : fallback.windows.two_years_plus.level,
      },
    },
  }
}

function normalizeConsultingAssessment(detail: CompatibilityDetail) {
  const report = detail.latest_report?.content_structured
  const base = detail.reading.consulting_assessment
  return {
    relationship_diagnosis: report?.relationship_diagnosis || base?.relationship_diagnosis,
    decision_advice: report?.decision_advice || base?.decision_advice,
    stage_risks: report?.stage_risks?.length ? report.stage_risks : base?.stage_risks || [],
    relationship_strategy: report?.relationship_strategy || base?.relationship_strategy,
    claim_evidence_links: report?.claim_evidence_links?.length ? report.claim_evidence_links : base?.claim_evidence_links || [],
  }
}

function clampScore(value: number) {
  return Math.max(0, Math.min(100, Math.round(value)))
}

function scoreTone(value: number) {
  if (value >= 78) return 'high'
  if (value >= 62) return 'medium'
  return 'low'
}

function getDimensionItems(scores: CompatibilityDimensionScoresLegacy) {
  return ([
    ['attraction', scores.attraction],
    ['stability', scores.stability],
    ['communication', scores.communication],
    ['practicality', scores.practicality],
  ] as Array<[keyof CompatibilityDimensionScoresLegacy, number]>).map(([key, value]) => ({
    key,
    label: dimensionText[key],
    hint: dimensionHint[key],
    value: clampScore(value),
    tone: scoreTone(value),
  }))
}

function EvidenceCard({ evidence }: { evidence: CompatibilityEvidence }) {
  const badgeColor = polarityColor[evidence.polarity] || 'var(--text-muted)'

  return (
    <div className="card compatibility-evidence-card">
      <div className="compatibility-evidence-header">
        <div className="serif compatibility-evidence-title">{evidence.title}</div>
        <div className="compatibility-evidence-badges">
          <span
            className="compatibility-evidence-badge"
          >
            {dimensionText[evidence.dimension] || evidence.dimension}
          </span>
          {evidence.perspective && (
            <span className="compatibility-evidence-badge">
              {perspectiveText[evidence.perspective] || evidence.perspective}
            </span>
          )}
          <span
            className="compatibility-evidence-badge"
            style={{
              border: `1px solid ${badgeColor}33`,
              color: badgeColor,
              background: `${badgeColor}14`,
            }}
          >
            {polarityText[evidence.polarity] || evidence.polarity}
          </span>
        </div>
      </div>
      <div className="compatibility-evidence-detail">{evidence.detail}</div>
      {Array.isArray(evidence.related_sources) && evidence.related_sources.length > 0 && (
        <div className="compatibility-evidence-related">
          关联：{evidence.related_sources.map(source => evidenceSourceText[source] || source).join(' / ')}
        </div>
      )}
    </div>
  )
}

function groupEvidenceBySource(evidences: CompatibilityEvidence[]) {
  const groups = new Map<string, CompatibilityEvidence[]>()
  evidences.forEach(evidence => {
    const key = evidence.source || 'unknown'
    const items = groups.get(key) || []
    items.push(evidence)
    groups.set(key, items)
  })
  return Array.from(groups.entries())
    .filter(([, items]) => items.length > 0)
    .sort(([a], [b]) => (evidenceSourceText[a] || a).localeCompare(evidenceSourceText[b] || b, 'zh-Hans-CN'))
}

function ProfessionalEvidenceGroups({ evidences }: { evidences: CompatibilityEvidence[] }) {
  const groups = groupEvidenceBySource(evidences)
  if (groups.length === 0) {
    return <p className="compatibility-report-empty">暂无结构化依据。</p>
  }

  return (
    <div className="compatibility-evidence-groups">
      {groups.map(([source, items]) => (
        <section key={source} className="compatibility-evidence-group">
          <div className="compatibility-evidence-group-header">
            <div className="serif compatibility-evidence-group-title">{evidenceSourceText[source] || source}</div>
            <div className="compatibility-evidence-group-count">{items.length} 条</div>
          </div>
          <div className="compatibility-evidence-grid">
            {items.map(evidence => <EvidenceCard key={evidence.id || evidence.evidence_key} evidence={evidence} />)}
          </div>
        </section>
      ))}
    </div>
  )
}

const dimensionHintV3: Record<keyof CompatibilityDimensionScoresV3, string> = {
  zodiac: '属相（年支）：六合/三合 50、五行同（双生）30、五行相生 20',
  nayin: '纳音五行：相生/相同 命中即满分 20',
  day_pillar: '日柱（亲密层）：支六合/三合 + 干合/生 10、支六合/三合 5、支五行同/相生 3',
  eight_chars: '年/月/时三柱：按日柱规则评分，最高 20',
}

const dimensionLabelV3: Record<keyof CompatibilityDimensionScoresV3, string> = {
  zodiac: '合属相',
  nayin: '合纳音',
  day_pillar: '合日柱',
  eight_chars: '合八字',
}

const dimensionMaxV3: Record<keyof CompatibilityDimensionScoresV3, number> = {
  zodiac: 50,
  nayin: 20,
  day_pillar: 10,
  eight_chars: 20,
}

function ScoreOverviewV3({
  scores,
  overallScore,
  overallLevel,
}: {
  scores: CompatibilityDimensionScoresV3
  overallScore: number
  overallLevel: 'high' | 'medium' | 'low'
}) {
  const keys: Array<keyof CompatibilityDimensionScoresV3> = [
    'zodiac',
    'nayin',
    'day_pillar',
    'eight_chars',
  ]
  return (
    <section className="compat-score-v3">
      <header className="compat-score-v3__header">
        <span className="compat-score-v3__total">{overallScore}</span>
        <span className="compat-score-v3__unit">/100</span>
        <span className={`compat-score-v3__badge compat-score-v3__badge--${overallLevel}`}>
          {overallLevel === 'high' ? '上吉' : overallLevel === 'medium' ? '中' : '低'}
        </span>
      </header>
      <ul className="compat-score-v3__modules">
        {keys.map((key) => {
          const value = scores[key]
          const max = dimensionMaxV3[key]
          return (
            <li key={key} className="compat-score-v3__module">
              <div className="compat-score-v3__module-row">
                <span className="compat-score-v3__module-label">{dimensionLabelV3[key]}</span>
                <span className="compat-score-v3__module-value">
                  {value}<span className="compat-score-v3__module-max">/{max}</span>
                </span>
              </div>
              <div className="compat-score-v3__module-hint">{dimensionHintV3[key]}</div>
              <div className="compat-score-v3__module-bar">
                <div
                  className="compat-score-v3__module-bar-fill"
                  style={{ width: `${(value / max) * 100}%` }}
                />
              </div>
            </li>
          )
        })}
      </ul>
    </section>
  )
}

function ScoreOverview({ scores }: { scores: CompatibilityDimensionScoresLegacy }) {
  return (
    <div className="card compatibility-quick-score">
      <div className="compatibility-section-header compatibility-section-header--stacked">
        <h2 className="serif compatibility-section-title">关系速览</h2>
        <p className="compatibility-section-desc">先看四个关键维度的强弱，再展开专业依据。</p>
      </div>
      <div className="compatibility-quick-score-list">
        {getDimensionItems(scores).map(item => (
          <div key={item.key} className={`compatibility-quick-score-row compatibility-quick-score-row--${item.tone}`}>
            <div className="compatibility-quick-score-copy">
              <div className="compatibility-quick-score-label">{item.label}</div>
              <div className="compatibility-quick-score-hint">{item.hint}</div>
            </div>
            <div className="compatibility-quick-score-meter">
              <div className="compatibility-quick-score-value serif">{item.value}</div>
              <div className="compatibility-quick-score-bar" aria-hidden="true">
                <div className="compatibility-quick-score-fill" style={{ width: `${item.value}%` }} />
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

function DecisionDashboardPanel({
  selfName,
  partnerName,
  reading,
  dashboard,
}: {
  selfName: string
  partnerName: string
  reading: CompatibilityDetail['reading']
  dashboard: DecisionDashboardData
}) {
  const stageLabel = relationshipStageText[reading.relationship_stage] || relationshipStageText.general
  const questionLabel = primaryQuestionText[reading.primary_question] || primaryQuestionText.general

  return (
    <section className="card compatibility-decision-dashboard">
      <div className="compatibility-decision-header">
        <HeartHandshake size={24} />
        <h1 className="serif compatibility-decision-title">{selfName} × {partnerName}</h1>
      </div>

      <div className="compatibility-decision-context">
        <span>{stageLabel}</span>
        <span>{questionLabel}</span>
      </div>

      <div className="compatibility-consulting-kicker">关系决策</div>
      <h2 className="serif compatibility-decision-headline">{dashboard.verdict}</h2>
      <div className="compatibility-decision-type">{dashboard.relationshipType}</div>

      <div className="compatibility-decision-metric-grid">
        <div className="compatibility-decision-metric">
          <span>投入建议</span>
          <strong>{dashboard.recommendationLabel}</strong>
        </div>
        <div className="compatibility-decision-metric">
          <span>判断信心</span>
          <strong>{dashboard.confidenceLabel}</strong>
        </div>
        <div className="compatibility-decision-metric compatibility-decision-metric--wide">
          <span>最大风险</span>
          <strong>{dashboard.maxRisk}</strong>
        </div>
      </div>

      <div className="compatibility-next-action">
        <span>下一步验证</span>
        <p>{dashboard.nextAction}</p>
      </div>

      {dashboard.avoid.length > 0 && (
        <div className="compatibility-avoid-list">
          <span>短期避免</span>
          <div>
            {dashboard.avoid.map(item => <em key={item}>{item}</em>)}
          </div>
        </div>
      )}

      <div className="compatibility-core-contradiction">
        <span>核心矛盾</span>
        <p>{dashboard.summary}</p>
      </div>
    </section>
  )
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

function DecisionEvidenceSummary({
  findings,
  links,
}: {
  findings: DecisionFinding[]
  links: CompatibilityClaimEvidenceLink[]
}) {
  if (findings.length === 0) return null

  return (
    <section className="compatibility-section compatibility-decision-evidence-summary">
      <div className="compatibility-section-header">
        <h2 className="serif compatibility-section-title">为什么这么判断</h2>
        <p className="compatibility-section-desc">先看白话依据，专业命理证据可继续下钻。</p>
      </div>
      <div className="compatibility-decision-evidence-grid">
        {findings.map((finding, index) => (
          <div key={`${finding.text}-${index}`} className="card compatibility-decision-evidence-card">
            <span className="compatibility-decision-evidence-index">{index + 1}</span>
            <p>{finding.text}</p>
            {hasLinkedEvidence(links, finding.evidenceKeys) && (
              <a href="#compatibility-claim-evidence" className="compatibility-evidence-link">查看依据</a>
            )}
          </div>
        ))}
      </div>
    </section>
  )
}

function ResultReadingMap() {
  const items = [
    { href: '#compatibility-personality-fit', label: '性格合不合' },
    { href: '#compatibility-conflict-validation', label: '冲突/验证' },
    { href: '#compatibility-score-evidence', label: '分数依据' },
    { href: '#compatibility-professional-details', label: '专业细节' },
  ]

  return (
    <nav className="compatibility-result-map" aria-label="合盘结果阅读顺序">
      {items.map(item => (
        <a key={item.href} href={item.href}>{item.label}</a>
      ))}
    </nav>
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
          {point.evidenceKey && <a href="#compatibility-claim-evidence">查看性格依据</a>}
        </div>
      ))}
    </div>
  )
}

function PersonalityFitPanel({ summary }: { summary: PersonalityFitSummary }) {
  return (
    <section id="compatibility-personality-fit" className="card compatibility-personality-fit compatibility-personality-fit--polished">
      <div className="compatibility-section-header compatibility-section-header--stacked">
        <div className="compatibility-consulting-kicker">性格相处画像</div>
        <h2 className="serif compatibility-section-title">{summary.headline}</h2>
        <p className="compatibility-personality-type-desc">{summary.matchTypeDescription}</p>
        <p className="compatibility-section-desc">
          当前问题：{summary.questionLabel} · 关系阶段：{summary.stageLabel}
        </p>
      </div>

      <p className="compatibility-personality-summary">{summary.summary}</p>

      <div className="compatibility-personality-pattern-grid">
        <div className="compatibility-personality-pattern">
          <span>{summary.selfPattern.title}</span>
          <p>{summary.selfPattern.detail}</p>
        </div>
        <div className="compatibility-personality-pattern">
          <span>{summary.partnerPattern.title}</span>
          <p>{summary.partnerPattern.detail}</p>
        </div>
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
    </section>
  )
}

function PersonalityValidationPlanPanel({
  plan,
  children,
}: {
  plan: PersonalityValidationPlan
  children: ReactNode
}) {
  const groups: Array<{ title: string; items: string[]; anchor?: string }> = [
    plan.shortTerm,
    plan.mediumTerm,
    plan.avoid,
  ]

  return (
    <section id="compatibility-conflict-validation" className="compatibility-section compatibility-validation-plan" aria-label="7 天观察与30 天验证">
      <div className="compatibility-section-header">
        <div>
          <h2 className="serif compatibility-section-title">性格验证计划</h2>
          <p className="compatibility-section-desc">用短期观察确认性格判断，不把一时吸引直接当成结论。</p>
        </div>
      </div>
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
        {children}
      </div>
    </section>
  )
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
          <div key={window.key} className="compatibility-duration-task-item">
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

function RelationshipStrategyPanel({ strategy }: { strategy: CompatibilityRelationshipStrategy }) {
  return (
    <div className="card compatibility-strategy-card">
      <div className="compatibility-consulting-kicker">关系经营策略</div>
      <div className="compatibility-strategy-grid">
        <AdviceList title="沟通" items={[strategy.communication].filter(Boolean)} />
        <AdviceList title="冲突" items={[strategy.conflict].filter(Boolean)} />
        <AdviceList title="现实" items={[strategy.reality].filter(Boolean)} />
        <AdviceList title="边界" items={[strategy.boundary].filter(Boolean)} />
      </div>
    </div>
  )
}

function EvidenceLinkedClaims({
  links,
  evidences,
}: {
  links: CompatibilityClaimEvidenceLink[]
  evidences: CompatibilityEvidence[]
}) {
  const byKey = new Map(evidences.map(evidence => [evidence.evidence_key || evidence.id, evidence]))
  const previewText = (text: string) => text.length > 72 ? `${text.slice(0, 72)}...` : text

  return (
    <div className="compatibility-claim-list">
      {links.map((link, index) => (
        <details key={link.claim_id || link.claim} className="card compatibility-claim-card" open={index === 0}>
          <summary>
            <span className="serif">{link.claim}</span>
            <span className="compatibility-claim-toggle">
              <span className="compatibility-claim-toggle-open">收起依据</span>
              <span className="compatibility-claim-toggle-closed">查看完整依据</span>
            </span>
          </summary>
          <div className="compatibility-claim-preview">{previewText(link.reasoning)}</div>
          <p>{link.reasoning}</p>
          {link.caveat && <p className="compatibility-claim-caveat">{link.caveat}</p>}
          <div className="compatibility-claim-evidence">
            {(link.evidence_keys || []).map(key => {
              const evidence = byKey.get(key)
              return evidence ? <EvidenceCard key={key} evidence={evidence} /> : null
            })}
          </div>
        </details>
      ))}
    </div>
  )
}

function DeepReportPanel({
  hasReport,
  structuredReport,
  reportDimensions,
  reportRisks,
  rawContent,
  error,
  reportLoading,
  onGenerateReport,
}: {
  structuredReport?: CompatibilityStructuredReport | null
  reportDimensions: CompatibilityStructuredReport['dimensions']
  reportRisks: string[]
  rawContent?: string
  error: string
  reportLoading: boolean
  onGenerateReport: () => void
  hasReport: boolean
}) {
  const reportStateClass = hasReport ? 'compatibility-ai-card--generated' : 'compatibility-ai-card--empty'

  return (
    <div className={reportStateClass}>
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

export default function CompatibilityResultPage() {
  const { id } = useParams()
  const { user, isLoading } = useAuth()
  const navigate = useNavigate()
  const [detail, setDetail] = useState<CompatibilityDetail | null>(null)
  const [loading, setLoading] = useState(true)
  const [reportLoading, setReportLoading] = useState(false)
  const [error, setError] = useState('')
  const [brand, setBrand] = useState<ExportBrand | null>(null)
  const [shareModalOpen, setShareModalOpen] = useState(false)
  const [savingImage, setSavingImage] = useState(false)
  const [exportingPDF, setExportingPDF] = useState(false)
  const shareCardRef = useRef<HTMLDivElement>(null)
  const shareModalCloseBtnRef = useRef<HTMLButtonElement>(null)
  const shareTriggerBtnRef = useRef<HTMLButtonElement>(null)
  const prevShareModalOpenRef = useRef(false)
  const isMobileDevice = /iPhone|iPad|iPod|Android/i.test(navigator.userAgent)
  const isIOS = /iPhone|iPad|iPod/i.test(navigator.userAgent)

  const load = useCallback(async () => {
    if (!id) return
    const res = await compatibilityAPI.getDetail(id)
    setDetail(res.data.data)
  }, [id])

  useEffect(() => {
    if (isLoading) {
      return
    }
    if (!user) {
      navigate('/login')
      return
    }
    load()
      .catch((err: unknown) => setError(err instanceof Error ? err.message : '加载失败'))
      .finally(() => setLoading(false))
  }, [user, isLoading, navigate, load])

  useEffect(() => {
    if (!user) return
    brandAPI.get()
      .then(r => setBrand(r.data.data))
      .catch(() => setBrand(null))
  }, [user])

  useEffect(() => {
    if (shareModalOpen) {
      shareModalCloseBtnRef.current?.focus()
    } else if (prevShareModalOpenRef.current) {
      shareTriggerBtnRef.current?.focus()
    }
    prevShareModalOpenRef.current = shareModalOpen
  }, [shareModalOpen])

  const handleGenerateReport = async () => {
    if (!id) return
    setReportLoading(true)
    setError('')
    try {
      await compatibilityAPI.generateReport(id)
      await load()
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : '生成合盘解读失败')
    } finally {
      setReportLoading(false)
    }
  }

  if (loading || isLoading) {
    return <div className="page"><div className="container" style={{ paddingTop: 40 }}>加载中...</div></div>
  }
  if (!detail) {
    return <div className="page"><div className="container" style={{ paddingTop: 40 }}>未找到合盘记录</div></div>
  }

  const reading = detail.reading
  const selfP = detail.participants.find(p => p.role === 'self')
  const partnerP = detail.participants.find(p => p.role === 'partner')

  const handleSaveImage = async () => {
    if (!shareCardRef.current) return
    setSavingImage(true)
    try {
      await document.fonts.ready
      const selfName = selfP?.display_name || '我'
      const partnerName = partnerP?.display_name || '伴侣'
      const fileName = `缘聚合盘-${selfName}-${partnerName}.png`

      if (isIOS) {
        const blob = await toBlob(shareCardRef.current, { quality: 0.98, pixelRatio: 3, cacheBust: true })
        if (!blob) throw new Error('生成图片失败')
        const file = new File([blob], fileName, { type: 'image/png' })
        if (navigator.canShare && navigator.canShare({ files: [file] })) {
          await navigator.share({
            files: [file],
            title: '缘聚合盘 · 命理合参',
            text: `${selfName} × ${partnerName} 综合契合度 ${detail?.reading.overall_score ?? ''} 分`,
          })
        } else {
          const objectUrl = URL.createObjectURL(blob)
          Object.assign(document.createElement('a'), { href: objectUrl, download: fileName }).click()
          setTimeout(() => URL.revokeObjectURL(objectUrl), 5000)
        }
      } else if (isMobileDevice) {
        const blob = await toBlob(shareCardRef.current, { quality: 0.98, pixelRatio: 3, cacheBust: true })
        if (!blob) throw new Error('生成图片失败')
        const objectUrl = URL.createObjectURL(blob)
        Object.assign(document.createElement('a'), { href: objectUrl, download: fileName }).click()
        setTimeout(() => URL.revokeObjectURL(objectUrl), 5000)
      } else {
        const dataUrl = await toPng(shareCardRef.current, { quality: 0.98, pixelRatio: 2, cacheBust: true })
        Object.assign(document.createElement('a'), { href: dataUrl, download: fileName }).click()
      }
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : ''
      if (!msg.includes('AbortError') && !msg.includes('cancel')) {
        alert('生成图片失败，请稍后重试')
      }
    } finally {
      setSavingImage(false)
    }
  }

  const handleExportPDF = async () => {
    if (!detail?.latest_report?.content_structured) return
    if (!isMobileDevice) {
      window.print()
      return
    }
    const el = document.querySelector('.compat-print-layout') as HTMLElement | null
    if (!el) return
    setExportingPDF(true)
    const prevDisplay = el.style.display
    try {
      await document.fonts.ready
      el.style.display = 'block'
      const canvas = await html2canvas(el, { scale: 2, useCORS: true, logging: false })
      el.style.display = prevDisplay
      const imgData = canvas.toDataURL('image/jpeg', 0.92)
      const pdf = new jsPDF({ orientation: 'portrait', unit: 'mm', format: 'a4' })
      const pageW = pdf.internal.pageSize.getWidth()
      const pageH = pdf.internal.pageSize.getHeight()
      const imgH = (canvas.height * pageW) / canvas.width
      let remaining = imgH
      let offset = 0
      pdf.addImage(imgData, 'JPEG', 0, offset, pageW, imgH)
      remaining -= pageH
      while (remaining > 0) {
        offset -= pageH
        pdf.addPage()
        pdf.addImage(imgData, 'JPEG', 0, offset, pageW, imgH)
        remaining -= pageH
      }
      const selfName = selfP?.display_name || '我'
      const partnerName = partnerP?.display_name || '伴侣'
      pdf.save(`缘聚合盘-${selfName}-${partnerName}.pdf`)
    } catch {
      alert('生成 PDF 失败，请稍后重试')
    } finally {
      el.style.display = prevDisplay
      setExportingPDF(false)
    }
  }

  const structuredReport = detail.latest_report?.content_structured
  const durationAssessment = normalizeDurationAssessment(structuredReport?.duration_assessment, reading.duration_assessment)
  const reportDimensions = Array.isArray(structuredReport?.dimensions) ? structuredReport.dimensions : []
  const reportRisks = Array.isArray(structuredReport?.risks) ? structuredReport.risks.filter(Boolean) : []
  const consulting = normalizeConsultingAssessment(detail)
  const decisionStageRisks = buildDecisionStageRisks(consulting.stage_risks, durationAssessment)
  const decisionDashboard = buildDecisionDashboardData({
    diagnosis: consulting.relationship_diagnosis,
    advice: consulting.decision_advice,
    stageRisks: consulting.stage_risks,
    duration: durationAssessment,
    evidences: detail.evidences,
    overallLevel: reading.overall_level,
  })
  const isV3 = (reading.analysis_version === 'v3' || reading.analysis_version === 'v3.1') && isV3DimensionScores(reading.dimension_scores)
  const legacyScores = isV3 ? null : (reading.dimension_scores as CompatibilityDimensionScoresLegacy)
  const v3Scores = isV3 ? (reading.dimension_scores as CompatibilityDimensionScoresV3) : null
  const personalitySummary = legacyScores
    ? buildPersonalityFitSummary({
        scores: legacyScores,
        evidences: detail.evidences,
        relationshipDiagnosis: consulting.relationship_diagnosis,
        relationshipStage: reading.relationship_stage,
        primaryQuestion: reading.primary_question,
        self: {
          name: selfP?.display_name,
          dayGan: selfP?.chart_snapshot?.day_gan,
        },
        partner: {
          name: partnerP?.display_name,
          dayGan: partnerP?.chart_snapshot?.day_gan,
        },
        hasReport: Boolean(detail.latest_report),
      })
    : null
  const personalityValidationPlan = personalitySummary
    ? buildPersonalityValidationPlan({
        personality: personalitySummary,
        advice: consulting.decision_advice,
        stageRisks: consulting.stage_risks,
        duration: durationAssessment,
        hasEvidence: detail.evidences.length > 0 || consulting.claim_evidence_links.length > 0,
      })
    : null

  return (
    <>
    <div className="page compatibility-result-page">
      <div className="container compatibility-result-container">
        <div className="compat-export-actions">
          <button
            type="button"
            ref={shareTriggerBtnRef}
            className="btn btn-secondary"
            disabled={!structuredReport}
            onClick={() => setShareModalOpen(true)}
            title={!structuredReport ? '请先生成命理解读' : ''}
          >
            分享图片
          </button>
          <button
            type="button"
            className="btn btn-primary"
            disabled={!structuredReport || exportingPDF}
            onClick={handleExportPDF}
            title={!structuredReport ? '请先生成命理解读' : ''}
          >
            {exportingPDF ? '生成中…' : '导出 PDF'}
          </button>
        </div>

        {ENABLE_NEW_LAYOUT && (
          <>
            <SectionBasicCharts />
            <SectionVerdict />
            <SectionDeepAnalysis />
          </>
        )}

        <DecisionDashboardPanel
          reading={reading}
          dashboard={decisionDashboard}
          selfName={selfP?.display_name || '我'}
          partnerName={partnerP?.display_name || '对方'}
        />

        <ResultReadingMap />

        <DecisionEvidenceSummary
          findings={decisionDashboard.findings}
          links={consulting.claim_evidence_links}
        />

        {personalitySummary && <PersonalityFitPanel summary={personalitySummary} />}

        {personalityValidationPlan ? (
          <PersonalityValidationPlanPanel plan={personalityValidationPlan}>
            <StageRiskGrid risks={decisionStageRisks} />
            <DurationTaskSummary assessment={durationAssessment} />
          </PersonalityValidationPlanPanel>
        ) : (
          <section id="compatibility-conflict-validation" className="compatibility-section">
            <div className="compatibility-section-header">
              <h2 className="serif compatibility-section-title">阶段风险与时段</h2>
              <p className="compatibility-section-desc">分阶段查看主要风险点和时段强弱。</p>
            </div>
            <StageRiskGrid risks={decisionStageRisks} />
            <DurationTaskSummary assessment={durationAssessment} />
          </section>
        )}

        {consulting.relationship_strategy && (
          <RelationshipStrategyPanel strategy={consulting.relationship_strategy} />
        )}

        <div id="compatibility-score-evidence" className="compatibility-section-anchor">
          {v3Scores ? (
            <ScoreOverviewV3
              scores={v3Scores}
              overallScore={reading.overall_score}
              overallLevel={reading.overall_level}
            />
          ) : (
            <ScoreOverview scores={legacyScores as CompatibilityDimensionScoresLegacy} />
          )}

          {consulting.claim_evidence_links.length > 0 && (
            <div id="compatibility-claim-evidence" className="compatibility-section">
              <div className="compatibility-section-header">
                <h2 className="serif compatibility-section-title">关键判断依据</h2>
                <p className="compatibility-section-desc">每条咨询判断都可以回看对应命理证据。</p>
              </div>
              <EvidenceLinkedClaims links={consulting.claim_evidence_links} evidences={detail.evidences} />
            </div>
          )}
        </div>

        <div className="card compatibility-ai-card">
          <DeepReportPanel
            hasReport={Boolean(detail.latest_report)}
            structuredReport={structuredReport}
            reportDimensions={reportDimensions}
            reportRisks={reportRisks}
            rawContent={detail.latest_report?.content}
            error={error}
            reportLoading={reportLoading}
            onGenerateReport={handleGenerateReport}
          />
        </div>

        <details className="compatibility-professional-details" id="compatibility-professional-details" open>
          <summary className="compatibility-professional-summary">
            <span className="serif">专业命盘细节</span>
            <span>四柱、五行与结构化依据</span>
          </summary>
          <div className="compatibility-professional-summary-grid">
            <span>双方四柱</span>
            <span>五行摘要</span>
            <span>结构化依据</span>
          </div>
          <div className="compatibility-professional-body">
            <div className="compatibility-section">
              <div className="compatibility-section-header">
                <h2 className="serif compatibility-section-title">双方命盘摘要</h2>
                <p className="compatibility-section-desc">确认双方四柱与命盘核心信息。</p>
              </div>
              <div className="compatibility-summary-grid">
                {selfP && <ParticipantSummaryCard participant={selfP} />}
                {partnerP && <ParticipantSummaryCard participant={partnerP} />}
              </div>
            </div>

            <div className="compatibility-section">
              <div className="compatibility-section-header">
                <h2 className="serif compatibility-section-title">关键依据</h2>
                <p className="compatibility-section-desc">这些结构化证据是合盘结论的主要命理依据。</p>
              </div>
              <ProfessionalEvidenceGroups evidences={detail.evidences} />
            </div>
          </div>
        </details>
      </div>

      {shareModalOpen && structuredReport && (
        <div
          className="compat-share-modal"
          role="dialog"
          aria-modal="true"
          aria-label="分享图片预览"
          tabIndex={-1}
          onClick={() => setShareModalOpen(false)}
          onKeyDown={(e) => { if (e.key === 'Escape') setShareModalOpen(false) }}
        >
          <div className="compat-share-modal-panel" onClick={e => e.stopPropagation()}>
            <header className="compat-share-modal-head">
              <h3>分享图片预览</h3>
              <button
                type="button"
                className="compat-share-modal-close"
                ref={shareModalCloseBtnRef}
                onClick={() => setShareModalOpen(false)}
                aria-label="关闭"
              >×</button>
            </header>
            <div className="compat-share-modal-preview">
              <CompatibilityShareCard
                ref={shareCardRef}
                reading={reading}
                participants={detail.participants}
                evidences={detail.evidences}
                decision={decisionDashboard}
                stageRisks={decisionStageRisks}
                structured={structuredReport ?? null}
                brand={brand}
              />
            </div>
            <p className="compat-share-modal-hint">导出的图片为完整版 · 预览可上下滚动</p>
            <footer className="compat-share-modal-footer">
              <button
                type="button"
                className="btn btn-primary"
                onClick={handleSaveImage}
                disabled={savingImage}
              >
                {savingImage ? '生成中…' : isIOS ? '保存 / 分享' : '保存到本地'}
              </button>
            </footer>
          </div>
        </div>
      )}

    </div>
    <CompatibilityPrintLayout
      reading={reading}
      participants={detail.participants}
      evidences={detail.evidences}
      decision={decisionDashboard}
      stageRisks={decisionStageRisks}
      structured={structuredReport ?? null}
      brand={brand}
    />
    </>
  )
}
