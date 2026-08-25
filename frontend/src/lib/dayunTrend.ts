export type TrendKey = 'career' | 'marriage' | 'wealth' | 'health'
export type TrendLevel = 1 | 2 | 3

export interface LiuNianTrendInput {
  year: number
  age?: number
  gan_zhi: string
  gan_shishen?: string
  zhi_shishen?: string
}

export interface DayunTrendInput {
  liu_nian?: LiuNianTrendInput[]
}

export interface DayunTrendPoint {
  year: number
  age?: number
  ganZhi: string
  level: TrendLevel
  label: string
  detail: string
}

export interface DayunTrendSeries {
  key: TrendKey
  title: string
  summary: string
  points: DayunTrendPoint[]
}

function trendLabel(level: TrendLevel) {
  if (level === 3) return '顺势'
  if (level === 1) return '留意'
  return '平稳'
}

function clampTrendLevel(score: number): TrendLevel {
  if (score >= 1) return 3
  if (score <= -1) return 1
  return 2
}

function includesAny(value: string, targets: string[]) {
  return targets.some(target => value.includes(target))
}

function scoreLiuNianTrend(key: TrendKey, item: LiuNianTrendInput, gender: string): TrendLevel {
  const gods = `${item.gan_shishen || ''} ${item.zhi_shishen || ''}`
  let score = 0

  if (key === 'career') {
    if (includesAny(gods, ['正官', '七杀', '正印', '偏印'])) score += 1
    if (includesAny(gods, ['伤官', '劫财'])) score -= 1
  }

  if (key === 'marriage') {
    const spouseStars = gender === 'female' ? ['正官', '七杀'] : ['正财', '偏财']
    if (includesAny(gods, spouseStars)) score += 1
    if (includesAny(gods, ['比肩', '劫财', '伤官'])) score -= 1
  }

  if (key === 'wealth') {
    if (includesAny(gods, ['正财', '偏财', '食神', '伤官'])) score += 1
    if (includesAny(gods, ['劫财', '比肩'])) score -= 1
  }

  if (key === 'health') {
    if (includesAny(gods, ['正印', '偏印', '食神'])) score += 1
    if (includesAny(gods, ['七杀', '伤官', '劫财'])) score -= 1
  }

  return clampTrendLevel(score)
}

function buildTrendDetail(key: TrendKey, item: LiuNianTrendInput, level: TrendLevel) {
  const gods = [item.gan_shishen, item.zhi_shishen].filter(Boolean).join(' / ') || '流年信号待排'
  const prefix = `${item.year}年${item.gan_zhi}，十神线索为${gods}。`
  const ending: Record<TrendKey, Record<TrendLevel, string>> = {
    career: {
      3: '事业侧更容易出现职责、平台或贵人机会，适合主动争取。',
      2: '事业侧以平稳推进为主，适合守住主线与节奏。',
      1: '事业侧容易出现压力或调整，适合提前管理合作与职责边界。',
    },
    marriage: {
      3: '关系侧更容易被激活，适合沟通、推进或修复承诺。',
      2: '关系侧节奏平稳，适合稳定表达与自然推进。',
      1: '关系侧需要留意误解、拉扯或现实安排带来的压力。',
    },
    wealth: {
      3: '财务与资源机会更明显，适合争取收入、合作或回报。',
      2: '财务侧以稳为主，适合整理现金流与控制投入。',
      1: '财务侧容易有支出、分配或合作波动，宜守不宜冒进。',
    },
    health: {
      3: '身心恢复和节律管理较有利，适合建立长期习惯。',
      2: '健康侧以维持为主，注意规律作息即可。',
      1: '健康侧需要留意过劳、情绪压力与日常安全。',
    },
  }
  return `${prefix}${ending[key][level]}`
}

function summarizeTrend(points: DayunTrendPoint[]) {
  const good = points.filter(point => point.level === 3).length
  const watch = points.filter(point => point.level === 1).length
  if (good >= 4 && watch <= 2) return '顺势较多'
  if (watch >= 4) return '需多留意'
  if (good >= 3 && watch >= 3) return '机会夹波动'
  return '整体平稳'
}

export function buildDayunTrendSeries(dayun: DayunTrendInput | null | undefined, gender: string): DayunTrendSeries[] {
  if (!dayun?.liu_nian?.length) return []
  const configs: Array<{ key: TrendKey; title: string }> = [
    { key: 'career', title: '事业' },
    { key: 'marriage', title: '婚恋' },
    { key: 'wealth', title: '财运' },
    { key: 'health', title: '健康' },
  ]

  return configs.map(config => {
    const points = dayun.liu_nian!.slice(0, 10).map(item => {
      const level = scoreLiuNianTrend(config.key, item, gender)
      return {
        year: item.year,
        age: item.age,
        ganZhi: item.gan_zhi,
        level,
        label: trendLabel(level),
        detail: buildTrendDetail(config.key, item, level),
      }
    })
    return {
      ...config,
      summary: summarizeTrend(points),
      points,
    }
  })
}

export function trendX(index: number, total: number) {
  if (total <= 1) return 54
  return 54 + (340 * index) / (total - 1)
}

export function trendY(level: TrendLevel) {
  if (level === 3) return 28
  if (level === 1) return 112
  return 70
}

export function buildTrendPath(points: DayunTrendPoint[]) {
  return points
    .map((point, index) => `${index === 0 ? 'M' : 'L'} ${trendX(index, points.length).toFixed(1)} ${trendY(point.level)}`)
    .join(' ')
}

export function buildTrendNote(series: DayunTrendSeries) {
  const goodYears = series.points.filter(point => point.level === 3).map(point => point.year).slice(0, 3)
  const watchYears = series.points.filter(point => point.level === 1).map(point => point.year).slice(0, 3)
  if (goodYears.length && watchYears.length) {
    return `顺势年份：${goodYears.join('、')}；需留意：${watchYears.join('、')}。`
  }
  if (goodYears.length) return `顺势年份：${goodYears.join('、')}，适合主动推进。`
  if (watchYears.length) return `需留意年份：${watchYears.join('、')}，建议提前放缓节奏。`
  return '整体以平稳推进为主，重点看当前年份与交运节点。'
}
