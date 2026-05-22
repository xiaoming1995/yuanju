import type {
  CompatibilityDecisionAdvice,
  CompatibilityDimensionScores,
  CompatibilityDurationAssessment,
  CompatibilityEvidence,
  CompatibilityRelationshipDiagnosis,
  CompatibilityStageRisk,
} from './api'

type RelationshipStage = string | null | undefined
type PrimaryQuestion = string | null | undefined

interface PersonalityParticipantInput {
  name?: string | null
  dayGan?: string | null
}

interface PersonalityFitInput {
  scores: CompatibilityDimensionScores
  evidences?: Array<Partial<CompatibilityEvidence>>
  relationshipDiagnosis?: CompatibilityRelationshipDiagnosis | null
  relationshipStage?: RelationshipStage
  primaryQuestion?: PrimaryQuestion
  self?: PersonalityParticipantInput | null
  partner?: PersonalityParticipantInput | null
  selfName?: string
  partnerName?: string
  hasReport?: boolean
}

export interface PersonalityPoint {
  title: string
  detail: string
  evidenceKey?: string
  dimension?: keyof CompatibilityDimensionScores | string
}

export interface PersonalityFitSummary {
  questionLabel: string
  stageLabel: string
  matchType: string
  matchTypeDescription: string
  headline: string
  summary: string
  selfPattern: PersonalityPoint
  partnerPattern: PersonalityPoint
  fitPoints: PersonalityPoint[]
  clashPoints: PersonalityPoint[]
  communicationGuidance: PersonalityPoint[]
  evidenceTargets: string[]
  reportNote: string
}

export interface PersonalityValidationPlan {
  shortTerm: {
    title: string
    items: string[]
    anchor?: string
  }
  mediumTerm: {
    title: string
    items: string[]
    anchor?: string
  }
  avoid: {
    title: string
    items: string[]
  }
  supportNote: string
}

export interface PersonalityConsultationPreview {
  title: string
  description: string
  bullets: string[]
}

const questionLabels: Record<string, string> = {
  continue_investment: '值不值得继续投入',
  marriage_suitability: '适不适合结婚',
  recurring_conflict: '为什么反复拉扯',
  reconciliation_potential: '复合有没有意义',
  long_term_stability: '长期能不能稳定',
  relationship_strategy: '怎么相处更顺',
  general: '性格合不合',
}

const stageLabels: Record<string, string> = {
  ambiguous: '暧昧中',
  dating: '恋爱中',
  long_distance: '异地中',
  reconciliation: '分手/复合中',
  marriage_or_engagement: '谈婚论嫁',
  crush: '单恋/暗恋',
  general: '综合关系判断',
}

const dimensionLabels: Record<string, string> = {
  attraction: '吸引力',
  stability: '稳定度',
  communication: '沟通协同',
  practicality: '现实磨合',
}

export function getCompatibilityQuestionLabel(question: PrimaryQuestion) {
  return questionLabels[question || 'general'] || questionLabels.general
}

export function getCompatibilityStageLabel(stage: RelationshipStage) {
  return stageLabels[stage || 'general'] || stageLabels.general
}

function clamp(value: number | undefined) {
  return Math.max(0, Math.min(100, Math.round(value || 0)))
}

function average(values: number[]) {
  if (values.length === 0) return 0
  return values.reduce((sum, value) => sum + value, 0) / values.length
}

export function getPersonalityMatchType(
  scores: CompatibilityDimensionScores,
  primaryQuestion?: PrimaryQuestion,
  relationshipStage?: RelationshipStage
) {
  const attraction = clamp(scores.attraction)
  const stability = clamp(scores.stability)
  const communication = clamp(scores.communication)
  const practicality = clamp(scores.practicality)

  if (attraction >= 76 && (communication < 58 || stability < 58)) return '高吸引高消耗型'
  if (primaryQuestion === 'recurring_conflict' || communication < 52) return '反复拉扯型'
  if (relationshipStage === 'long_distance' || practicality < 52) return '现实压力型'
  if (stability >= 70 && communication >= 62 && practicality >= 62) return '稳定互补型'
  return '慢热磨合型'
}

export function getPersonalityMatchTypeDescription(matchType: string) {
  const descriptions: Record<string, string> = {
    稳定互补型: '彼此节奏相对稳定，适合用长期承接和现实配合继续确认。',
    高吸引高消耗型: '吸引感强，但沟通节奏和情绪消耗需要先被看见。',
    慢热磨合型: '不一定一开始很强烈，适合通过持续互动慢慢确认默契。',
    反复拉扯型: '容易靠近又容易误解，关键在冲突后能否修复。',
    现实压力型: '不是没有吸引，而是距离、责任或节奏会更考验关系。',
  }
  return descriptions[matchType] || '这类关系需要先观察真实互动，再决定投入强度。'
}

function scorePoint(dimension: keyof CompatibilityDimensionScores, value: number): PersonalityPoint {
  const label = dimensionLabels[dimension]
  if (dimension === 'attraction') {
    return {
      title: `${label}较强`,
      detail: value >= 70 ? '两个人靠近感明显，容易先被对方的状态牵引。' : '吸引感需要靠持续互动慢慢累积。',
      dimension,
    }
  }
  if (dimension === 'communication') {
    return {
      title: `${label}${value >= 62 ? '有修复空间' : '需要重点验证'}`,
      detail: value >= 62 ? '出现误会时有机会通过解释、放慢节奏来修复。' : '容易在表达速度、情绪回应或冲突修复上反复错位。',
      dimension,
    }
  }
  if (dimension === 'practicality') {
    return {
      title: `${label}${value >= 62 ? '可落地' : '承压明显'}`,
      detail: value >= 62 ? '现实安排不是最大阻力，适合用具体计划验证关系承接度。' : '现实节奏、距离、责任分工或投入比例容易带来压力。',
      dimension,
    }
  }
  return {
    title: `${label}${value >= 62 ? '能承接' : '需观察'}`,
    detail: value >= 62 ? '长期相处有一定承接力，可以观察稳定回应是否持续。' : '长期稳定性仍要看真实互动中的责任感和情绪稳定。',
    dimension,
  }
}

function pointFromEvidence(evidence: Partial<CompatibilityEvidence>): PersonalityPoint {
  return {
    title: evidence.title || scorePoint((evidence.dimension || 'communication') as keyof CompatibilityDimensionScores, 50).title,
    detail: evidence.detail || '这条依据提示双方互动中存在可观察的相处模式。',
    evidenceKey: evidence.evidence_key || evidence.id,
    dimension: evidence.dimension,
  }
}

function compactPoints(points: PersonalityPoint[], fallback: PersonalityPoint, max = 3) {
  const filtered = points.filter(point => point.title || point.detail)
  return (filtered.length > 0 ? filtered : [fallback]).slice(0, max)
}

function participantPattern(name: string, role: 'self' | 'partner', scores: CompatibilityDimensionScores, dayGan?: string | null): PersonalityPoint {
  const communication = clamp(scores.communication)
  const stability = clamp(scores.stability)
  const dayMaster = dayGan ? `日主${dayGan}` : '命盘结构'
  const title = role === 'self' ? `${name}的相处模式` : `${name}的关系需求`
  const detail = communication >= stability
    ? `${dayMaster}提示这方更适合把需求说清楚，在关系里需要被理解和及时回应。`
    : `${dayMaster}提示这方更看重稳定承接，在关系里需要确认对方是否持续可靠。`
  return { title, detail }
}

function getEvidenceTargets(points: PersonalityPoint[]) {
  return Array.from(new Set(points.map(point => point.evidenceKey).filter((key): key is string => Boolean(key))))
}

export function buildPersonalityFitSummary(input: PersonalityFitInput): PersonalityFitSummary {
  const evidences = Array.isArray(input.evidences) ? input.evidences : []
  const positiveEvidence = evidences
    .filter(evidence => evidence.polarity === 'positive' || evidence.polarity === 'neutral')
    .map(pointFromEvidence)
  const pressureEvidence = evidences
    .filter(evidence => evidence.polarity === 'negative' || evidence.polarity === 'mixed')
    .map(pointFromEvidence)
  const highDimension = ([
    ['attraction', input.scores.attraction],
    ['stability', input.scores.stability],
    ['communication', input.scores.communication],
    ['practicality', input.scores.practicality],
  ] as Array<[keyof CompatibilityDimensionScores, number]>).sort((a, b) => b[1] - a[1])[0]
  const lowDimension = ([
    ['attraction', input.scores.attraction],
    ['stability', input.scores.stability],
    ['communication', input.scores.communication],
    ['practicality', input.scores.practicality],
  ] as Array<[keyof CompatibilityDimensionScores, number]>).sort((a, b) => a[1] - b[1])[0]

  const questionLabel = getCompatibilityQuestionLabel(input.primaryQuestion)
  const stageLabel = getCompatibilityStageLabel(input.relationshipStage)
  const matchType = getPersonalityMatchType(input.scores, input.primaryQuestion, input.relationshipStage)
  const selfName = input.self?.name || input.selfName || '我'
  const partnerName = input.partner?.name || input.partnerName || '对方'
  const fitPoints = compactPoints(
    positiveEvidence,
    scorePoint(highDimension[0], highDimension[1])
  )
  const clashPoints = compactPoints(
    pressureEvidence,
    scorePoint(lowDimension[0], lowDimension[1])
  )
  const allPointTargets = [...fitPoints, ...clashPoints]
  const overallScore = Math.round(average([
    clamp(input.scores.attraction),
    clamp(input.scores.stability),
    clamp(input.scores.communication),
    clamp(input.scores.practicality),
  ]))
  const diagnosisSummary = input.relationshipDiagnosis?.summary?.trim()

  return {
    questionLabel,
    stageLabel,
    matchType,
    matchTypeDescription: getPersonalityMatchTypeDescription(matchType),
    headline: `性格合不合：${matchType}`,
    summary: diagnosisSummary || `从四维结构看，这段关系属于${matchType}，整体匹配约为 ${overallScore} 分。先看双方相处节奏是否能稳定回应，再决定投入强度。`,
    selfPattern: participantPattern(selfName, 'self', input.scores, input.self?.dayGan),
    partnerPattern: participantPattern(partnerName, 'partner', input.scores, input.partner?.dayGan),
    fitPoints,
    clashPoints,
    communicationGuidance: [
      scorePoint('communication', input.scores.communication),
      {
        title: '沟通建议',
        detail: clamp(input.scores.communication) >= 62
          ? '适合把需求、边界和下一步计划说具体，用真实互动继续确认默契。'
          : '先降低情绪密度，遇到分歧时确认事实和需求，再讨论投入或承诺。',
        dimension: 'communication',
      },
    ],
    evidenceTargets: getEvidenceTargets(allPointTargets),
    reportNote: input.hasReport
      ? '深度解读已生成，可结合下方报告细看性格判断的细节。'
      : '当前性格判断来自结构化合盘数据；生成深度解读后，可进一步细化双方相处模式。',
  }
}

export function buildPersonalityValidationPlan({
  personality,
  advice,
  stageRisks,
  duration,
  hasEvidence,
}: {
  personality: PersonalityFitSummary
  advice?: CompatibilityDecisionAdvice | null
  stageRisks?: CompatibilityStageRisk[]
  duration?: CompatibilityDurationAssessment | null
  hasEvidence?: boolean
}): PersonalityValidationPlan {
  const shortRisk = stageRisks?.find(risk => risk.window === 'three_months') || stageRisks?.[0]
  const mediumRisk = stageRisks?.find(risk => risk.window === 'one_year') || stageRisks?.[1] || shortRisk
  const doNext = Array.isArray(advice?.do_next) ? advice.do_next.filter(Boolean) : []
  const avoid = Array.isArray(advice?.avoid) ? advice.avoid.filter(Boolean) : []
  const conditions = Array.isArray(advice?.conditions) ? advice.conditions.filter(Boolean) : []
  const durationHint = duration?.summary?.trim()

  return {
    shortTerm: {
      title: '7 天观察',
      items: [
        `观察双方是否能围绕「${personality.questionLabel}」稳定回应，而不是只靠一时情绪推进。`,
        shortRisk?.advice || doNext[0] || '确认聊天节奏、见面安排和情绪回应是否能自然承接。',
      ],
      anchor: shortRisk ? '#compatibility-stage-validation' : undefined,
    },
    mediumTerm: {
      title: '30 天验证',
      items: [
        mediumRisk?.main_risk ? `验证「${mediumRisk.main_risk}」是否反复出现。` : `确认${personality.matchType}是否在真实相处里仍然成立。`,
        conditions[0] || durationHint || '观察冲突后能否修复，以及现实安排是否有人主动承接。',
      ],
      anchor: mediumRisk ? '#compatibility-stage-validation' : undefined,
    },
    avoid: {
      title: '暂缓/避免',
      items: compactTextItems([
        ...avoid,
        '不要把短期吸引直接等同于长期稳定。',
        '避免在还没有验证沟通修复前过早加码投入。',
      ], 3),
    },
    supportNote: hasEvidence
      ? '这些观察点可回到阶段风险、分数和关键依据继续核对。'
      : '当前先用关系阶段和四维分数做观察计划，深度解读可继续补充依据。',
  }
}

function compactTextItems(items: string[], max: number) {
  return Array.from(new Set(items.filter(Boolean))).slice(0, max)
}

export function buildPersonalityConsultationPreview(
  relationshipStage: RelationshipStage,
  primaryQuestion: PrimaryQuestion
): PersonalityConsultationPreview {
  const stageLabel = getCompatibilityStageLabel(relationshipStage)
  const questionLabel = getCompatibilityQuestionLabel(primaryQuestion)
  return {
    title: `先看性格合不合，再看要不要继续`,
    description: `${stageLabel}阶段会优先回答「${questionLabel}」，并拆成双方相处模式、合的地方、冲突点和验证建议。`,
    bullets: [
      '双方在关系里的需求、压力反应和沟通节奏',
      '自然契合点与容易反复拉扯的地方',
      '7 天观察、30 天验证和短期避免事项',
    ],
  }
}
