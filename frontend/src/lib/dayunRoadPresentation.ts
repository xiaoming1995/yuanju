export interface DayunRoadPhaseLike {
  key?: string
  label?: string
  score?: number
  summary?: string
  detail?: string
}

export interface DayunPeriodLike {
  gan: string
  zhi: string
  start_year: number
  end_year: number
}

export type JinBuHuanRating = '吉' | '平' | '凶'

export interface DayunPhasePrompt {
  phaseLabel: '前五年' | '后五年'
  timeRange: string
  roadLabel: string
  theme: string
  governingLabel: string
  rating: JinBuHuanRating
  ratingTone: 'favorable' | 'neutral' | 'adverse'
  detail: string
}

const LEGACY_PHASE_RATING: Record<string, JinBuHuanRating> = {
  高速路: '吉',
  城市主路: '吉',
  山路: '平',
  泥路: '平',
  施工路段: '凶',
}

const PHASE_ROAD_THEME: Record<string, string> = {
  高速路: '快速推进',
  城市主路: '稳步落地',
  山路: '选择与节奏',
  泥路: '稳住调整',
  施工路段: '修整蓄力',
}

function getRoadLabelFromScore(score?: number): string | null {
  if (typeof score !== 'number') return null
  if (score >= 80) return '高速路'
  if (score >= 65) return '城市主路'
  if (score >= 55) return '山路'
  if (score >= 40) return '泥路'
  return '施工路段'
}

function getFallbackRoadLabel(rating: JinBuHuanRating): string {
  if (rating === '吉') return '城市主路'
  if (rating === '凶') return '施工路段'
  return '山路'
}

export function getDayunPhaseRoadLabel(phase?: DayunRoadPhaseLike | null): string {
  const existingLabel = phase?.label?.trim()
  if (existingLabel) return existingLabel
  return getRoadLabelFromScore(phase?.score) || getFallbackRoadLabel(getJinBuHuanRating(phase))
}

export function getDayunPhaseTheme(roadLabel: string): string {
  return PHASE_ROAD_THEME[roadLabel] || '顺势调整'
}

export function getJinBuHuanRating(phase?: DayunRoadPhaseLike | null): JinBuHuanRating {
  if (typeof phase?.score === 'number') {
    if (phase.score >= 70) return '吉'
    if (phase.score < 55) return '凶'
    return '平'
  }
  return LEGACY_PHASE_RATING[phase?.label || ''] || '平'
}

export function getDayunPhasePrompt(
  period: DayunPeriodLike,
  phase: DayunRoadPhaseLike | null | undefined,
  position: 'front' | 'back',
): DayunPhasePrompt | null {
  if (!phase) return null

  const isFront = position === 'front'
  const startYear = isFront ? period.start_year : Math.min(period.start_year + 5, period.end_year)
  const endYear = isFront ? Math.min(period.start_year + 4, period.end_year) : period.end_year
  const rating = getJinBuHuanRating(phase)
  const roadLabel = getDayunPhaseRoadLabel(phase)
  const phaseLabel = isFront ? '前五年' : '后五年'
  const governingLabel = isFront ? `天干${period.gan}主事` : `地支${period.zhi}主事`

  return {
    phaseLabel,
    timeRange: `${startYear}-${endYear}`,
    roadLabel,
    theme: getDayunPhaseTheme(roadLabel),
    governingLabel,
    rating,
    ratingTone: rating === '吉' ? 'favorable' : rating === '凶' ? 'adverse' : 'neutral',
    detail: phase.detail || phase.summary || `${phaseLabel}按金不换提示${rating}。`,
  }
}

export const DAYUN_ROAD_SCOPE_EXPLANATION = '十年综合路况综合金不换、扶抑五行、格局、十神、十二长生与神煞等信息；前后五年阶段路况则以金不换分别判断大运天干与地支，并给出对应阶段主题。'
