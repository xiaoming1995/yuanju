import { useCallback, useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { HeartHandshake } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'
import {
  compatibilityAPI,
  type CompatibilityChartSnapshot,
  type CompatibilityClaimEvidenceLink,
  type CompatibilityDecisionAdvice,
  type CompatibilityDetail,
  type CompatibilityDimensionScores,
  type CompatibilityDurationAssessment,
  type CompatibilityEvidence,
  type CompatibilityParticipant,
  type CompatibilityQuestionFocus,
  type CompatibilityRelationshipDiagnosis,
  type CompatibilityRelationshipStrategy,
  type CompatibilityStageRisk,
} from '../lib/api'
import './CompatibilityResultPage.css'

const levelText: Record<string, string> = {
  high: '契合度高',
  medium: '可发展，但需要磨合',
  low: '吸引与稳定存在明显矛盾',
}

const dimensionText: Record<string, string> = {
  attraction: '会不会互相吸引？',
  stability: '能不能长期稳定？',
  communication: '吵架后能不能修复？',
  practicality: '现实条件能不能落地？',
}

const dimensionHint: Record<keyof CompatibilityDimensionScores, string> = {
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

const wuxingLabel = [
  { key: 'mu', label: '木', className: 'wuxing-mu' },
  { key: 'huo', label: '火', className: 'wuxing-huo' },
  { key: 'tu', label: '土', className: 'wuxing-tu' },
  { key: 'jin', label: '金', className: 'wuxing-jin' },
  { key: 'shui', label: '水', className: 'wuxing-shui' },
] as const

function formatBirthText(snapshot?: CompatibilityChartSnapshot | null, fallback?: CompatibilityParticipant['birth_profile']) {
  if (snapshot) {
    return `${snapshot.birth_year}年${snapshot.birth_month}月${snapshot.birth_day}日 ${snapshot.birth_hour}时`
  }
  if (fallback) {
    return `${fallback.year}年${fallback.month}月${fallback.day}日 ${fallback.hour}时`
  }
  return '出生信息缺失'
}

function genderText(snapshot?: CompatibilityChartSnapshot | null, fallback?: CompatibilityParticipant['birth_profile']) {
  const value = snapshot?.gender || fallback?.gender
  return value === 'female' ? '女命' : '男命'
}

function getPillars(snapshot?: CompatibilityChartSnapshot | null) {
  if (!snapshot) return []
  return [
    { label: '年柱', value: `${snapshot.year_gan}${snapshot.year_zhi}` },
    { label: '月柱', value: `${snapshot.month_gan}${snapshot.month_zhi}` },
    { label: '日柱', value: `${snapshot.day_gan}${snapshot.day_zhi}` },
    { label: '时柱', value: `${snapshot.hour_gan}${snapshot.hour_zhi}` },
  ]
}

function getWuxingItems(snapshot?: CompatibilityChartSnapshot | null) {
  return wuxingLabel.map(item => ({
    ...item,
    value: snapshot?.wuxing?.[item.key] ?? 0,
  }))
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

function getDimensionItems(scores: CompatibilityDimensionScores) {
  return ([
    ['attraction', scores.attraction],
    ['stability', scores.stability],
    ['communication', scores.communication],
    ['practicality', scores.practicality],
  ] as Array<[keyof CompatibilityDimensionScores, number]>).map(([key, value]) => ({
    key,
    label: dimensionText[key],
    hint: dimensionHint[key],
    value: clampScore(value),
    tone: scoreTone(value),
  }))
}

function fallbackAdvice(level: string) {
  if (level === 'high') {
    return '这组关系有继续推进的基础，建议把优势落到稳定沟通、现实安排和共同节奏上。'
  }
  if (level === 'low') {
    return '这组关系需要先处理边界、沟通方式和现实压力，再判断是否适合长期投入。'
  }
  return '这组关系可以继续观察，但要把吸引感和现实磨合分开看，先建立稳定的沟通规则。'
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
    </div>
  )
}

function ScoreOverview({ scores }: { scores: CompatibilityDimensionScores }) {
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

function InsightPanel({
  risks,
  advice,
  hasStructuredReport,
}: {
  risks: string[]
  advice: string
  hasStructuredReport: boolean
}) {
  return (
    <div className="compatibility-insight-grid">
      <div className="card compatibility-insight-card">
        <div className="compatibility-insight-kicker">关键风险</div>
        {risks.length > 0 ? (
          <ul className="compatibility-insight-list">
            {risks.slice(0, 3).map(risk => <li key={risk}>{risk}</li>)}
          </ul>
        ) : (
          <p className="compatibility-insight-empty">当前结构中没有特别突出的单点风险，仍建议结合现实相处节奏判断。</p>
        )}
      </div>
      <div className="card compatibility-insight-card compatibility-insight-card--advice">
        <div className="compatibility-insight-kicker">行动建议</div>
        <p className="compatibility-insight-advice">{advice}</p>
        {!hasStructuredReport && (
          <p className="compatibility-insight-note">生成完整解读后，会补充更具体的沟通与长期关系建议。</p>
        )}
      </div>
    </div>
  )
}

function DecisionHeroPanel({
  detail,
  diagnosis,
  advice,
  selfName,
  partnerName,
  overallLevel,
  summaryTags,
  hasReport,
  reportLoading,
  onGenerateReport,
}: {
  detail: CompatibilityDetail
  diagnosis?: CompatibilityRelationshipDiagnosis
  advice?: CompatibilityDecisionAdvice
  selfName: string
  partnerName: string
  overallLevel: string
  summaryTags: string[]
  hasReport: boolean
  reportLoading: boolean
  onGenerateReport: () => void
}) {
  const reading = detail.reading
  const nextActions = Array.isArray(advice?.do_next) ? advice.do_next.filter(Boolean).slice(0, 3) : []
  const avoid = Array.isArray(advice?.avoid) ? advice.avoid.filter(Boolean).slice(0, 2) : []
  const findings = Array.isArray(diagnosis?.top_findings) ? diagnosis.top_findings.filter(Boolean).slice(0, 3) : []
  const stageLabel = relationshipStageText[reading.relationship_stage] || relationshipStageText.general
  const questionLabel = primaryQuestionText[reading.primary_question] || primaryQuestionText.general

  return (
    <div className="card compatibility-decision-hero">
      <div className="compatibility-decision-header">
        <HeartHandshake size={24} />
        <h1 className="serif compatibility-decision-title">{selfName} × {partnerName}</h1>
      </div>
      <div className="compatibility-context-heading">关系背景</div>
      <div className="compatibility-decision-context">
        <span>{stageLabel}</span>
        <span>{questionLabel}</span>
        <span>{levelText[overallLevel] || overallLevel}</span>
      </div>
      <div className="compatibility-consulting-kicker">关系判断</div>
      <h2 className="serif compatibility-decision-headline">{diagnosis?.verdict || '建议结合现实相处继续观察'}</h2>
      {diagnosis?.relationship_type && (
        <div className="compatibility-decision-type">{diagnosis.relationship_type}</div>
      )}
      <div className="compatibility-core-contradiction">
        <span>核心矛盾</span>
        <p>{diagnosis?.summary || '这段关系需要把短期吸引和长期稳定分开验证。'}</p>
      </div>
      {findings.length > 0 && (
        <div className="compatibility-finding-list">
          {findings.map(finding => (
            <div key={finding.text} className="compatibility-finding-item">{finding.text}</div>
          ))}
        </div>
      )}
      <div className="compatibility-decision-action-grid">
        <AdviceList title="下一步" items={nextActions} />
        <AdviceList title="避免" items={avoid} />
      </div>
      {summaryTags.length > 0 && (
        <div className="compatibility-tag-row">
          {summaryTags.map(tag => (
            <span key={tag} className="compatibility-tag">{tag}</span>
          ))}
        </div>
      )}
      {!hasReport && (
        <div className="compatibility-decision-report-action">
          <button className="btn btn-primary" onClick={onGenerateReport} disabled={reportLoading}>
            {reportLoading ? '生成中...' : '生成完整解读'}
          </button>
        </div>
      )}
    </div>
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

function StageRiskGrid({ risks }: { risks: CompatibilityStageRisk[] }) {
  return (
    <div className="compatibility-stage-grid">
      {risks.map(risk => (
        <div key={risk.window} className="card compatibility-stage-card">
          <div className="compatibility-stage-window">{stageWindowText[risk.window] || risk.window}</div>
          <div className="compatibility-stage-label">阶段任务</div>
          <div className="serif compatibility-stage-risk">{risk.main_risk}</div>
          <p><span>风险触发：</span>{risk.trigger}</p>
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

  return (
    <div className="compatibility-claim-list">
      {links.map(link => (
        <details key={link.claim_id || link.claim} className="card compatibility-claim-card">
          <summary>
            <span className="serif">{link.claim}</span>
            <span>查看依据</span>
          </summary>
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

function ParticipantSummaryCard({
  participant,
}: {
  participant: CompatibilityParticipant
}) {
  const snapshot = participant.chart_snapshot || null
  const pillars = getPillars(snapshot)
  const wuxingItems = getWuxingItems(snapshot)

  return (
    <div className="card compatibility-person-card">
      <div className="compatibility-person-header">
        <div>
          <div className="compatibility-person-name serif">{participant.display_name}</div>
          <div className="compatibility-person-meta">
            <span>{genderText(snapshot, participant.birth_profile)}</span>
            <span>{formatBirthText(snapshot, participant.birth_profile)}</span>
          </div>
        </div>
        {snapshot?.day_gan && (
          <div className="compatibility-day-master">
            <span className="compatibility-day-master-label">日主</span>
            <span className="compatibility-day-master-value serif">{snapshot.day_gan}</span>
          </div>
        )}
      </div>

      {pillars.length > 0 && (
        <div className="compatibility-pillar-grid">
          {pillars.map(pillar => (
            <div key={pillar.label} className="compatibility-pillar-cell">
              <div className="compatibility-pillar-label">{pillar.label}</div>
              <div className="compatibility-pillar-value serif">{pillar.value}</div>
            </div>
          ))}
        </div>
      )}

      <div className="compatibility-wuxing-title">五行概览</div>
      <div className="compatibility-wuxing-grid">
        {wuxingItems.map(item => (
          <div key={item.key} className="compatibility-wuxing-item">
            <span className={`wuxing-badge ${item.className}`}>{item.label}</span>
            <span className="compatibility-wuxing-value">{item.value}</span>
          </div>
        ))}
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
  const structuredReport = detail.latest_report?.content_structured
  const durationAssessment = normalizeDurationAssessment(structuredReport?.duration_assessment, reading.duration_assessment)
  const summaryTags = Array.isArray(reading.summary_tags) ? reading.summary_tags : []
  const reportDimensions = Array.isArray(structuredReport?.dimensions) ? structuredReport.dimensions : []
  const reportRisks = Array.isArray(structuredReport?.risks) ? structuredReport.risks.filter(Boolean) : []
  const fallbackRisks = detail.evidences
    .filter(evidence => evidence.polarity === 'negative')
    .map(evidence => `${evidence.title}：${evidence.detail}`)
  const insightRisks = reportRisks.length > 0 ? reportRisks : fallbackRisks
  const insightAdvice = structuredReport?.advice?.trim() || fallbackAdvice(reading.overall_level)
  const consulting = normalizeConsultingAssessment(detail)

  return (
    <div className="page compatibility-result-page">
      <div className="container compatibility-result-container">
        <DecisionHeroPanel
          detail={detail}
          diagnosis={consulting.relationship_diagnosis}
          advice={consulting.decision_advice}
          selfName={selfP?.display_name || '我'}
          partnerName={partnerP?.display_name || '对方'}
          overallLevel={reading.overall_level}
          summaryTags={summaryTags}
          hasReport={Boolean(detail.latest_report)}
          reportLoading={reportLoading}
          onGenerateReport={handleGenerateReport}
        />

        {consulting.stage_risks.length > 0 && (
          <div className="compatibility-section">
            <div className="compatibility-section-header">
              <h2 className="serif compatibility-section-title">阶段任务与风险</h2>
              <p className="compatibility-section-desc">按关系推进阶段看任务、触发点和行动建议。</p>
            </div>
            <StageRiskGrid risks={consulting.stage_risks} />
            <DurationTaskSummary assessment={durationAssessment} />
          </div>
        )}

        {consulting.relationship_strategy && (
          <RelationshipStrategyPanel strategy={consulting.relationship_strategy} />
        )}

        <ScoreOverview scores={reading.dimension_scores} />

        <div className="compatibility-section">
          <div className="compatibility-section-header compatibility-section-header--stacked">
            <h2 className="serif compatibility-section-title">关系洞察</h2>
            <p className="compatibility-section-desc">把风险和建议单独提出来，避免被专业术语淹没。</p>
          </div>
          <InsightPanel
            risks={insightRisks}
            advice={insightAdvice}
            hasStructuredReport={Boolean(structuredReport)}
          />
        </div>

        {consulting.claim_evidence_links.length > 0 && (
          <div className="compatibility-section">
            <div className="compatibility-section-header">
              <h2 className="serif compatibility-section-title">关键判断依据</h2>
              <p className="compatibility-section-desc">每条咨询判断都可以回看对应命理证据。</p>
            </div>
            <EvidenceLinkedClaims links={consulting.claim_evidence_links} evidences={detail.evidences} />
          </div>
        )}

        <details className="compatibility-professional-details">
          <summary className="compatibility-professional-summary">
            <span className="serif">专业命盘细节</span>
            <span>四柱、五行与结构化依据</span>
          </summary>
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
              <div className="compatibility-evidence-grid">
                {detail.evidences.map(evidence => <EvidenceCard key={evidence.id} evidence={evidence} />)}
              </div>
            </div>
          </div>
        </details>

        <div className="card compatibility-ai-card">
          <div className="compatibility-ai-header">
            <h2 className="serif compatibility-section-title">合盘解读</h2>
          </div>

          {error && <p style={{ color: '#e77' }}>{error}</p>}

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
          ) : detail.latest_report ? (
            <div className="compatibility-report-raw">{detail.latest_report.content}</div>
          ) : (
            <p className="compatibility-report-empty">尚未生成合盘解读。</p>
          )}
        </div>
      </div>
    </div>
  )
}
