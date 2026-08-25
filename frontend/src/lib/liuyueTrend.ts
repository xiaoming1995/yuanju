import type { LiuYueItem } from './api'

export type LiuYueTrendKey = 'overall' | 'career' | 'marriage' | 'wealth' | 'health'
export type LiuYueTrendLevel = 1 | 2 | 3

export interface LiuYueTrendPoint {
  index: number
  monthName: string
  startDate: string
  endDate: string
  ganZhi: string
  level: LiuYueTrendLevel
  label: string
  detail: string
  ganShishen: string
  zhiShishen: string
}

export interface LiuYueTrendSeries {
  key: LiuYueTrendKey
  title: string
  summary: string
  points: LiuYueTrendPoint[]
}

const TREND_CONFIGS: Array<{ key: LiuYueTrendKey; title: string }> = [
  { key: 'overall', title: '综合' },
  { key: 'career', title: '事业' },
  { key: 'marriage', title: '婚恋' },
  { key: 'wealth', title: '财运' },
  { key: 'health', title: '健康' },
]

function includesAny(value: string, targets: string[]) {
  return targets.some(target => value.includes(target))
}

function trendLabel(level: LiuYueTrendLevel) {
  if (level === 3) return '顺势'
  if (level === 1) return '留意'
  return '平稳'
}

function clampScore(score: number): LiuYueTrendLevel {
  if (score >= 1) return 3
  if (score <= -1) return 1
  return 2
}

function scoreDimension(key: Exclude<LiuYueTrendKey, 'overall'>, item: LiuYueItem, gender?: string) {
  const gods = `${item.gan_shishen || ''} ${item.zhi_shishen || ''}`
  let score = 0

  if (key === 'career') {
    if (includesAny(gods, ['正官', '七杀', '正印', '偏印'])) score += 1
    if (includesAny(gods, ['伤官', '劫财'])) score -= 1
  }

  if (key === 'marriage') {
    const spouseStars = gender === 'male'
      ? ['正财', '偏财']
      : gender === 'female'
        ? ['正官', '七杀']
        : ['正财', '偏财', '正官', '七杀']
    if (includesAny(gods, spouseStars)) score += 1
    if (includesAny(gods, ['比肩', '劫财', '伤官'])) score -= 1
  }

  if (key === 'wealth') {
    if (includesAny(gods, ['正财', '偏财', '食神', '伤官'])) score += 1
    if (includesAny(gods, ['比肩', '劫财'])) score -= 1
  }

  if (key === 'health') {
    if (includesAny(gods, ['正印', '偏印', '食神'])) score += 1
    if (includesAny(gods, ['七杀', '伤官', '劫财'])) score -= 1
  }

  return score
}

function scoreTrend(key: LiuYueTrendKey, item: LiuYueItem, gender?: string): LiuYueTrendLevel {
  if (key !== 'overall') return clampScore(scoreDimension(key, item, gender))

  const scores = [
    scoreDimension('career', item, gender),
    scoreDimension('marriage', item, gender),
    scoreDimension('wealth', item, gender),
    scoreDimension('health', item, gender),
  ]
  const total = scores.reduce((sum, score) => sum + score, 0)
  if (total >= 2) return 3
  if (total <= -2) return 1
  return 2
}

function buildPointDetail(key: LiuYueTrendKey, item: LiuYueItem, level: LiuYueTrendLevel) {
  const gods = [item.gan_shishen, item.zhi_shishen].filter(Boolean).join(' / ') || '十神待排'
  const prefix = `${item.month_name}${item.gan_zhi}，十神线索为${gods}。`
  const endings: Record<LiuYueTrendKey, Record<LiuYueTrendLevel, string>> = {
    overall: {
      3: '本月综合节奏偏顺，适合推进重要安排。',
      2: '本月综合节奏平稳，适合按原计划执行。',
      1: '本月综合波动偏多，宜放慢节奏并预留余地。',
    },
    career: {
      3: '事业侧更容易出现职责、资源或贵人机会。',
      2: '事业侧以稳定推进为主，适合守住主线。',
      1: '事业侧容易有压力、沟通成本或职责调整。',
    },
    marriage: {
      3: '关系侧更容易被激活，适合表达、沟通或推进关系。',
      2: '关系侧节奏平稳，适合自然相处。',
      1: '关系侧需要留意误解、拉扯或现实压力。',
    },
    wealth: {
      3: '财务与资源机会更明显，适合争取回报。',
      2: '财务侧以稳为主，适合整理现金流。',
      1: '财务侧容易有支出、分配或合作波动。',
    },
    health: {
      3: '身心恢复和节律管理较有利。',
      2: '健康侧以维持为主，注意规律即可。',
      1: '健康侧需要留意过劳、情绪压力与日常安全。',
    },
  }
  return `${prefix}${endings[key][level]}`
}

function summarize(points: LiuYueTrendPoint[]) {
  const good = points.filter(point => point.level === 3).length
  const watch = points.filter(point => point.level === 1).length
  if (good >= 5 && watch <= 2) return '顺势月份较多'
  if (watch >= 5) return '需多留意'
  if (good >= 4 && watch >= 4) return '机会夹波动'
  return '整体平稳'
}

export function buildLiuYueTrendSeries(items: LiuYueItem[], gender?: string): LiuYueTrendSeries[] {
  if (!items.length) return []

  return TREND_CONFIGS.map(config => {
    const points = items.map(item => {
      const level = scoreTrend(config.key, item, gender)
      return {
        index: item.index,
        monthName: item.month_name,
        startDate: item.start_date,
        endDate: item.end_date,
        ganZhi: item.gan_zhi,
        level,
        label: trendLabel(level),
        detail: buildPointDetail(config.key, item, level),
        ganShishen: item.gan_shishen,
        zhiShishen: item.zhi_shishen,
      }
    })

    return {
      ...config,
      summary: summarize(points),
      points,
    }
  })
}

export function liuyueTrendX(index: number, total: number) {
  if (total <= 1) return 48
  return 48 + (312 * index) / (total - 1)
}

export function liuyueTrendY(level: LiuYueTrendLevel) {
  if (level === 3) return 34
  if (level === 1) return 112
  return 73
}

export function buildLiuYueTrendPath(points: LiuYueTrendPoint[]) {
  return points
    .map((point, index) => `${index === 0 ? 'M' : 'L'} ${liuyueTrendX(index, points.length).toFixed(1)} ${liuyueTrendY(point.level)}`)
    .join(' ')
}
