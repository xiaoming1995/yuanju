import type {
  CompatibilityClaimEvidenceLink,
  CompatibilityDecisionAdvice,
  CompatibilityDurationAssessment,
  CompatibilityEvidence,
  CompatibilityFinding,
  CompatibilityRelationshipDiagnosis,
  CompatibilityStageRisk,
} from './api'

export interface DecisionFinding {
  text: string
  evidenceKeys: string[]
}

export interface DecisionDashboardData {
  verdict: string
  relationshipType: string
  summary: string
  recommendationLabel: string
  confidenceLabel: string
  maxRisk: string
  nextAction: string
  avoid: string[]
  findings: DecisionFinding[]
}

export const recommendationLabelMap: Record<string, string> = {
  continue: '可推进',
  observe: '继续观察',
  caution: '谨慎投入',
  pause: '暂缓投入',
}

export const confidenceLabelMap: Record<string, string> = {
  high: '高',
  medium: '中',
  low: '低',
}

const levelRecommendationMap: Record<string, string> = {
  high: '可推进',
  medium: '继续观察',
  low: '谨慎投入',
}

const levelVerdictMap: Record<string, string> = {
  high: '有继续推进的基础，仍需验证现实节奏',
  medium: '建议继续观察，不宜过早重投入',
  low: '建议谨慎投入，先处理边界与现实压力',
}

const durationWindowLabel: Record<string, string> = {
  three_months: '3 个月',
  one_year: '1 年',
  two_years_plus: '2 年以上',
}

const durationRiskText: Record<string, string> = {
  high: '阶段承压偏高',
  medium: '阶段风险中等',
  low: '阶段压力较低',
}

function compact(values: Array<string | undefined | null>) {
  return values.map(value => value?.trim()).filter((value): value is string => Boolean(value))
}

function firstNonEmpty(values: Array<string | undefined | null>, fallback: string) {
  return compact(values)[0] || fallback
}

export function verdictFromOverallLevel(overallLevel: string) {
  return levelVerdictMap[overallLevel] || levelVerdictMap.medium
}

export function recommendationLabel(recommendation: string | undefined, overallLevel: string) {
  if (recommendation && recommendationLabelMap[recommendation]) {
    return recommendationLabelMap[recommendation]
  }
  return levelRecommendationMap[overallLevel] || levelRecommendationMap.medium
}

export function confidenceLabel(confidence: string | undefined) {
  return confidenceLabelMap[confidence || ''] || confidenceLabelMap.medium
}

export function buildDecisionStageRisks(
  stageRisks: CompatibilityStageRisk[] | undefined,
  duration: CompatibilityDurationAssessment,
): CompatibilityStageRisk[] {
  const safeRisks = Array.isArray(stageRisks) ? stageRisks.filter(risk => risk && risk.window) : []
  if (safeRisks.length > 0) return safeRisks

  const windows = [
    { window: 'three_months', level: duration.windows?.three_months?.level },
    { window: 'one_year', level: duration.windows?.one_year?.level },
    { window: 'two_years_plus', level: duration.windows?.two_years_plus?.level },
  ] as const

  return windows.map(item => ({
    window: item.window,
    risk_level: item.level || 'medium',
    main_risk: `${durationWindowLabel[item.window]}${durationRiskText[item.level || 'medium']}`,
    trigger: '关系推进、沟通频率或现实安排发生变化时',
    advice: '先观察冲突后的修复能力，再决定是否增加投入。',
    evidence_keys: [],
  }))
}

export function buildDecisionFindings(
  diagnosis: CompatibilityRelationshipDiagnosis | undefined,
  evidences: CompatibilityEvidence[],
): DecisionFinding[] {
  const topFindings = Array.isArray(diagnosis?.top_findings)
    ? diagnosis.top_findings.filter(Boolean).slice(0, 3)
    : []

  if (topFindings.length > 0) {
    return topFindings.map((finding: CompatibilityFinding) => ({
      text: finding.text,
      evidenceKeys: finding.evidence_keys || [],
    })).filter(finding => finding.text)
  }

  const positive = evidences.find(evidence => evidence.polarity === 'positive')
  const negative = evidences.find(evidence => evidence.polarity === 'negative')
  return [positive, negative]
    .filter((evidence): evidence is CompatibilityEvidence => Boolean(evidence))
    .slice(0, 3)
    .map(evidence => ({
      text: `${evidence.title}：${evidence.detail}`,
      evidenceKeys: compact([evidence.evidence_key || evidence.id]),
    }))
}

export function buildDecisionDashboardData({
  diagnosis,
  advice,
  stageRisks,
  duration,
  evidences,
  overallLevel,
}: {
  diagnosis?: CompatibilityRelationshipDiagnosis
  advice?: CompatibilityDecisionAdvice
  stageRisks?: CompatibilityStageRisk[]
  duration: CompatibilityDurationAssessment
  evidences: CompatibilityEvidence[]
  overallLevel: string
}): DecisionDashboardData {
  const resolvedStageRisks = buildDecisionStageRisks(stageRisks, duration)
  const negativeEvidence = evidences.find(evidence => evidence.polarity === 'negative')
  const maxRisk = firstNonEmpty(
    [
      resolvedStageRisks[0]?.main_risk,
      negativeEvidence ? `${negativeEvidence.title}：${negativeEvidence.detail}` : '',
    ],
    '短期先验证沟通节奏和现实安排是否稳定。',
  )
  const nextAction = firstNonEmpty(
    Array.isArray(advice?.do_next) ? advice?.do_next : [],
    resolvedStageRisks[0]?.advice || '未来 1-2 个月先观察冲突后的修复能力。',
  )

  return {
    verdict: firstNonEmpty([diagnosis?.verdict], verdictFromOverallLevel(overallLevel)),
    relationshipType: firstNonEmpty([diagnosis?.relationship_type], '关系需要结合现实节奏观察'),
    summary: firstNonEmpty([diagnosis?.summary], '这段关系需要把短期吸引和长期稳定分开验证。'),
    recommendationLabel: recommendationLabel(advice?.recommendation, overallLevel),
    confidenceLabel: confidenceLabel(advice?.confidence),
    maxRisk,
    nextAction,
    avoid: compact(Array.isArray(advice?.avoid) ? advice?.avoid : []).slice(0, 2),
    findings: buildDecisionFindings(diagnosis, evidences),
  }
}

export function hasLinkedEvidence(
  links: CompatibilityClaimEvidenceLink[] | undefined,
  evidenceKeys: string[],
) {
  if (!Array.isArray(links) || links.length === 0 || evidenceKeys.length === 0) return false
  return links.some(link => link.evidence_keys?.some(key => evidenceKeys.includes(key)))
}
