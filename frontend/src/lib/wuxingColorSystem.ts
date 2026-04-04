/**
 * 五行视觉映射规则集
 * 将木/火/土/金/水映射为完整的视觉属性，作为所有视觉化功能的底层数据源
 */

export type WuxingKey = '木' | '火' | '土' | '金' | '水'

export interface WuxingProfile {
  /** 主色（hex） */
  color: string
  /** 深色版本，用于渐变结尾或强调 */
  darkColor: string
  /** 浅色/透明版本，用于背景 */
  lightColor: string
  /** SVG 纹样形状标识 */
  shape: 'tall' | 'sharp' | 'square' | 'round' | 'wave'
  /** 方位 */
  direction: string
  /** 季节 */
  season: string
  /** 幸运数字数组 */
  luckyNumbers: number[]
  /** 质感关键词 */
  material: string
  /** 五行描述 */
  description: string
  /** 代表 Emoji */
  emoji: string
  /** CSS 变量名（对应 index.css 中已有的五行色彩变量） */
  cssVar: string
  /** CSS 类名后缀（对应 wuxing-badge 样式） */
  cssClass: 'mu' | 'huo' | 'tu' | 'jin' | 'shui'
}

export const WUXING_MAP: Record<WuxingKey, WuxingProfile> = {
  木: {
    color: '#4caf7d',
    darkColor: '#2e7d52',
    lightColor: 'rgba(76, 175, 125, 0.15)',
    shape: 'tall',
    direction: '东方',
    season: '春',
    luckyNumbers: [3, 8],
    material: '木质、竹制',
    description: '木主生长、仁慈、条达，性格温和富有生命力，适合东方朝阳之地',
    emoji: '🌿',
    cssVar: '--wu-mu',
    cssClass: 'mu',
  },
  火: {
    color: '#e05c4b',
    darkColor: '#b03a2b',
    lightColor: 'rgba(224, 92, 75, 0.15)',
    shape: 'sharp',
    direction: '南方',
    season: '夏',
    luckyNumbers: [2, 7],
    material: '棉麻、皮质',
    description: '火主礼仪、热情、光明，性格外向充满活力，适合南方温暖之地',
    emoji: '🔥',
    cssVar: '--wu-huo',
    cssClass: 'huo',
  },
  土: {
    color: '#c17f3e',
    darkColor: '#8b5a2b',
    lightColor: 'rgba(193, 127, 62, 0.15)',
    shape: 'square',
    direction: '中央',
    season: '四季末',
    luckyNumbers: [5, 10],
    material: '陶瓷、石材',
    description: '土主信义、厚重、中庸，性格稳重踏实，中央四方皆宜',
    emoji: '🌍',
    cssVar: '--wu-tu',
    cssClass: 'tu',
  },
  金: {
    color: '#c9a84c',
    darkColor: '#a8872a',
    lightColor: 'rgba(201, 168, 76, 0.15)',
    shape: 'round',
    direction: '西方',
    season: '秋',
    luckyNumbers: [4, 9],
    material: '金属、矿石',
    description: '金主义气、刚毅、肃杀，性格果敢有原则，适合西方金秋之地',
    emoji: '⚡',
    cssVar: '--wu-jin',
    cssClass: 'jin',
  },
  水: {
    color: '#5b9bd5',
    darkColor: '#2c6ea3',
    lightColor: 'rgba(91, 155, 213, 0.15)',
    shape: 'wave',
    direction: '北方',
    season: '冬',
    luckyNumbers: [1, 6],
    material: '玻璃、水晶',
    description: '水主智慧、灵动、隐忍，性格聪慧善变，适合北方寒冬之地',
    emoji: '💧',
    cssVar: '--wu-shui',
    cssClass: 'shui',
  },
}

const VALID_WUXING = new Set<string>(['木', '火', '土', '金', '水'])

/**
 * 从喜用神/忌神字符串中提取有效五行字符列表
 * @param str 如 "木火"、"金水土"、"水"、"" 等
 * @returns 过滤后的五行字符数组，如 ["木", "火"]
 */
export function parseWuxingList(str: string): WuxingKey[] {
  if (!str) return []
  const result: WuxingKey[] = []
  for (const char of str) {
    if (VALID_WUXING.has(char) && !result.includes(char as WuxingKey)) {
      result.push(char as WuxingKey)
    }
  }
  return result
}

/**
 * 合并多个五行的幸运数字（去重）
 */
export function mergeLuckyNumbers(wuxingList: WuxingKey[]): number[] {
  const nums = new Set<number>()
  for (const wx of wuxingList) {
    for (const n of WUXING_MAP[wx].luckyNumbers) {
      nums.add(n)
    }
  }
  return Array.from(nums).sort((a, b) => a - b)
}

/**
 * 合并多个五行的幸运方位（去重）
 */
export function mergeLuckyDirections(wuxingList: WuxingKey[]): string[] {
  const seen = new Set<string>()
  const result: string[] = []
  for (const wx of wuxingList) {
    const dir = WUXING_MAP[wx].direction
    if (!seen.has(dir)) {
      seen.add(dir)
      result.push(dir)
    }
  }
  return result
}

/**
 * 合并多个五行的幸运季节（去重）
 */
export function mergeLuckySeasons(wuxingList: WuxingKey[]): string[] {
  const seen = new Set<string>()
  const result: string[] = []
  for (const wx of wuxingList) {
    const season = WUXING_MAP[wx].season
    if (!seen.has(season)) {
      seen.add(season)
      result.push(season)
    }
  }
  return result
}
