import type {
  CompatibilityDecisionAdvice,
  CompatibilityDimensionScoresLegacy,
  CompatibilityDurationAssessment,
  CompatibilityStageRisk,
} from './api'

type RelationshipStage = string | null | undefined
type PrimaryQuestion = string | null | undefined

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

export function getCompatibilityQuestionLabel(question: PrimaryQuestion) {
  return questionLabels[question || 'general'] || questionLabels.general
}

export function getCompatibilityStageLabel(stage: RelationshipStage) {
  return stageLabels[stage || 'general'] || stageLabels.general
}

function clamp(value: number | undefined) {
  return Math.max(0, Math.min(100, Math.round(value || 0)))
}

export function getPersonalityMatchType(
  scores: CompatibilityDimensionScoresLegacy,
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

export function buildPersonalityValidationPlan({
  questionLabel,
  matchType,
  advice,
  stageRisks,
  duration,
  hasEvidence,
}: {
  questionLabel: string
  matchType: string
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
        `观察双方是否能围绕「${questionLabel}」稳定回应，而不是只靠一时情绪推进。`,
        shortRisk?.advice || doNext[0] || '确认聊天节奏、见面安排和情绪回应是否能自然承接。',
      ],
      anchor: shortRisk ? '#compatibility-stage-validation' : undefined,
    },
    mediumTerm: {
      title: '30 天验证',
      items: [
        mediumRisk?.main_risk ? `验证「${mediumRisk.main_risk}」是否反复出现。` : `确认${matchType}是否在真实相处里仍然成立。`,
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
      : '当前先用关系阶段和双方画像做观察计划，深度解读可继续补充依据。',
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
    description: `${stageLabel}阶段会优先回答「${questionLabel}」，并拆成双方各自性格、合的地方、冲突点和验证建议。`,
    bullets: [
      '双方各自的表达、决策、亲密需求与情绪反应',
      '自然契合点与容易反复拉扯的地方',
      '7 天观察、30 天验证和短期避免事项',
    ],
  }
}
