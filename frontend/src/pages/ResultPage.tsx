import { useLocation, useParams, useNavigate } from 'react-router-dom'
import { Fragment, useEffect, useRef, useState } from 'react'
import { Diamond, X, History } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'
import { authAPI, baziAPI, brandAPI, fetchShenshaAnnotations } from '../lib/api'
import type { AIReport, ShenshaAnnotation, StructuredReport, PolishedReport, ExportBrand, CalculateInput, MingGeHistoricalFigure } from '../lib/api'
import { cleanLegacyReportContent, cleanReportText, hasMeaningfulReportContent, splitParagraphs } from '../lib/reportText'
import { buildAuthPath } from '../lib/authRedirect'
import { createPendingBaziJourney, savePendingJourney } from '../lib/pendingJourney'
import { buildBaziResultRoute } from '../lib/resultRoute'
import WuxingRadar from '../components/WuxingRadar'
import DayunTimeline from '../components/DayunTimeline'
import YongshenBadge from '../components/YongshenBadge'
import MingpanAvatar from '../components/MingpanAvatar'
import TiaohouCard from '../components/TiaohouCard'
import ThermalTiaohouCard from '../components/ThermalTiaohouCard'
import ShareCard from '../components/ShareCard'
import PrintLayout from '../components/PrintLayout'
import PolishedPanel from '../components/PolishedPanel'
import GongJiaPanel, { type GongJiaItem } from '../components/GongJiaPanel'
import { SegmentedTabs } from '../components/ui/SegmentedTabs'
import { useToast } from '../components/ui/useToast'
import { filterPastEventsExportSegments, type PastEventsExportReadySegment } from '../lib/pastEventsViewModel'
import { buildDayunTrendSeries, buildTrendNote, buildTrendPath, trendX, trendY } from '../lib/dayunTrend'
import './ResultPage.css'
import './WealthProfile.css'

// 特性开关 (Feature Flags)
const ENABLE_MINGPAN_AVATAR = false // 暂时隐藏专属命理头像模块

const WUXING_MAP: Record<string, string> = {
  '木': 'mu', '火': 'huo', '土': 'tu', '金': 'jin', '水': 'shui'
}

// 神煞极性表（与后端 ShenShaPolarity 保持同步）
// ji = 吉神（金色），xiong = 凶煞（红色），zhong = 中性（灰色）
const SHENSHA_POLARITY: Record<string, string> = {
  // 吉神
  '天乙贵人': 'ji', '太极贵人': 'ji', '文昌贵人': 'ji', '禄神': 'ji',
  '天德贵人': 'ji', '月德贵人': 'ji', '天德合': 'ji', '月德合': 'ji',
  '德秀贵人': 'ji', '金舆贵人': 'ji', '天喜': 'ji', '天厨贵人': 'ji',
  '国印贵人': 'ji', '三奇贵人': 'ji', '日德': 'ji', '将星': 'ji', '福星贵人': 'ji', '天医': 'ji',
  '十灵日': 'ji', '词馆': 'ji',
  // 凶煞
  '羊刃': 'xiong', '飞刃': 'xiong', '劫煞': 'xiong', '亡神': 'xiong',
  '孤辰': 'xiong', '寡宿': 'xiong', '阴差阳错': 'xiong',
  '魁罡': 'xiong', '十恶大败': 'xiong', '天罗地网': 'xiong', '地网': 'xiong', '灾煞': 'xiong', '童子煞': 'xiong',
  '流霞': 'xiong', '吊客': 'xiong', '墓门': 'xiong',
  // 中性
  '桃花': 'zhong', '驿马': 'zhong', '华盖': 'zhong', '红艳': 'zhong',
}

const LOADING_STEPS = [
  '提取四柱大运神煞...',
  '结合真太阳时精确校准星运...',
  '依据典籍推断月令与格局...',
  '深度分析抓取全局调候用神...',
  '正在汇总专属呈现命局详析...'
]

const REPORT_TERMS = [
  { term: '用神', desc: '命局中最需要扶助或调节的关键五行。' },
  { term: '忌神', desc: '容易加重失衡，需要节制或避开的五行。' },
  { term: '格局', desc: '月令与十神形成的命局主结构。' },
  { term: '大运', desc: '每十年一段的人生阶段性趋势。' },
]

const GAN_WUXING: Record<string, string> = {
  '甲': '木', '乙': '木',
  '丙': '火', '丁': '火',
  '戊': '土', '己': '土',
  '庚': '金', '辛': '金',
  '壬': '水', '癸': '水',
}

const GAN_YINYANG: Record<string, number> = {
  '甲': 1, '乙': 0,
  '丙': 1, '丁': 0,
  '戊': 1, '己': 0,
  '庚': 1, '辛': 0,
  '壬': 1, '癸': 0,
}

const WUXING_SHENG: Record<string, string> = {
  '木': '火', '火': '土', '土': '金', '金': '水', '水': '木',
}

const WUXING_KE: Record<string, string> = {
  '木': '土', '火': '金', '土': '水', '金': '木', '水': '火',
}

const TEN_GOD_META: Record<string, { relation: string; group: string; group_label: string; summary: string }> = {
  '比肩': { relation: '同我', group: 'peer', group_label: '比劫', summary: '自我、同类、独立意识与同辈协作。' },
  '劫财': { relation: '同我', group: 'peer', group_label: '比劫', summary: '同辈竞争、资源分配、行动冲劲与合作博弈。' },
  '食神': { relation: '我生', group: 'output', group_label: '食伤', summary: '表达、才艺、稳定输出、口福与享受。' },
  '伤官': { relation: '我生', group: 'output', group_label: '食伤', summary: '创意表达、突破规则、才华外放与锋芒。' },
  '正财': { relation: '我克', group: 'wealth', group_label: '财星', summary: '稳定财富、务实经营、责任感与现实积累。' },
  '偏财': { relation: '我克', group: 'wealth', group_label: '财星', summary: '机会资源、流动财富、人情经营与商业嗅觉。' },
  '正官': { relation: '克我', group: 'official', group_label: '官杀', summary: '规则、职位、责任、名誉与秩序感。' },
  '七杀': { relation: '克我', group: 'official', group_label: '官杀', summary: '外部压力、竞争、规则挑战与行动魄力。' },
  '正印': { relation: '生我', group: 'seal', group_label: '印星', summary: '学习、贵人、保护、资质与正统资源。' },
  '偏印': { relation: '生我', group: 'seal', group_label: '印星', summary: '灵感、研究、特殊资源、独特思维与保护。' },
}

function buildReportDigestItems(structured: StructuredReport, result: BaziResult) {
  const firstChapter = structured.chapters?.[0]
  const adviceChapter = structured.chapters?.find((c) =>
    /建议|总结|总论|方向|展望|策略|布局/.test(c.title),
  )
  const yongshen = structured.yongshen || result.yongshen || ''
  const jishen = structured.jishen || result.jishen || ''
  const fallbackAdvice = yongshen && jishen
    ? `优先围绕「${yongshen}」方向布局，对「${jishen}」相关领域更克制谨慎。`
    : '先读摘要，再展开与当前问题最相关的章节。'

  return [
    {
      label: '总体判断',
      value: cleanReportText(structured.analysis?.summary || firstChapter?.brief) || '已生成完整命理解读，可继续查看各章节。',
    },
    {
      label: '喜用重点',
      value: `${yongshen || '待判定'}：优先观察能补足命局平衡的方向。`,
    },
    {
      label: '主要风险',
      value: `${jishen || '待判定'}：相关五行过旺或失衡时，需要在选择与节奏上更谨慎。`,
    },
    {
      label: '行动建议',
      value: cleanReportText(structured.analysis?.advice) || cleanReportText(adviceChapter?.brief) || fallbackAdvice,
    },
  ]
}


interface TenGodDayMaster {
  gan: string
  wuxing: string
  label: string
}

interface TenGodRelationItem {
  pillar: string
  pillar_label: string
  gan: string
  wuxing: string
  ten_god: string
  group?: string
  group_label?: string
  relation: string
  summary: string
}

interface TenGodHiddenStemItem {
  gan: string
  wuxing: string
  ten_god: string
  group?: string
  group_label?: string
  relation: string
  summary: string
}

interface TenGodHiddenStemGroup {
  pillar: string
  pillar_label: string
  branch: string
  items: TenGodHiddenStemItem[]
}

interface TenGodRelationMatrix {
  day_master: TenGodDayMaster
  heavenly_stems: TenGodRelationItem[]
  hidden_stems: TenGodHiddenStemGroup[]
}

interface ProfileEvidence {
  source: string
  label: string
  impact: string
  delta: number
  detail: string
}

interface VehicleProfile {
  grade: string
  grade_label: string
  score: number
  vehicle_type: string
  driving_style?: string
  summary: string
  tags: string[]
  evidences: ProfileEvidence[]
}

interface WealthWindowHint {
  year: number
  dayun_index: number
  gan_zhi: string
  level: string
  label: string
  summary: string
  evidences: string[]
}

interface WealthProfile {
  version?: string
  grade: string
  grade_label: string
  score: number
  wealth_type: string
  summary: string
  tags: string[]
  risk_flags: string[]
  evidences: ProfileEvidence[]
  current_hint?: WealthWindowHint
}

interface RoadPhase {
  key: string
  label: string
  score: number
  summary: string
  detail?: string
}

interface DayunPhaseEvidence {
  phase: 'front' | 'back'
  label: string
  delta: number
  evidences: ProfileEvidence[]
}

interface NatalStemGuidanceItem {
  stem: string
  element: string
  ten_god: string
  source_layers: string[]
  detail: string
}

interface NatalStemGuidance {
  primary_favorable: NatalStemGuidanceItem[]
  secondary_favorable: NatalStemGuidanceItem[]
  conditioning_only: NatalStemGuidanceItem[]
  adverse: NatalStemGuidanceItem[]
}

interface StemLevelYongshenSummary {
  primary?: NatalStemGuidanceItem
  usable: NatalStemGuidanceItem[]
  adverse: NatalStemGuidanceItem[]
  conditioning_reference: NatalStemGuidanceItem[]
}

interface DayunRoad {
  dayun_index: number
  gan_zhi: string
  score: number
  road_type: string
  road_label: string
  qian_road: RoadPhase
  hou_road: RoadPhase
  summary: string
  tags: string[]
  evidences: ProfileEvidence[]
  phase_evidences?: DayunPhaseEvidence[]
}

interface NatalAssessment {
  version: string
  climate: { status: string; required_elements?: string; score: number; grade_ceiling: string }
  tiaohou?: {
    day_stem: { status: string; formation?: string; foundation_tier?: string; foundation_score?: number; required_stems: string[]; visible_stems: string[]; hidden_stems: string[]; score: number; detail: string }
    thermal: { status: string; condition: string; required_elements?: string; visible_support: string[]; hidden_support: string[]; detail: string }
  }
  fuyi: { day_master_strength: string; yongshen: string; jishen: string; support_level: string; score: number; strength_score?: number; evidence?: string }
  yongshen_alignment?: { elements: string[]; source_layers: string[]; detail?: string }
  stem_guidance?: NatalStemGuidance
  pattern: { name: string; quality: string; foundation_source?: string; foundation_label?: string; foundation_tier?: string; formations: string[]; breaks: string[] }
  relations: { flow: string; combinations: string[]; disruptions: string[] }
  grade: { score: number; grade: string; label: string; grade_ceiling: string }
  evidences: ProfileEvidence[]
}

interface BaziResult {
  display_name?: string
  year_gan: string; year_zhi: string
  month_gan: string; month_zhi: string
  day_gan: string; day_zhi: string
  hour_gan: string; hour_zhi: string
  year_gan_wuxing: string; year_zhi_wuxing: string
  month_gan_wuxing: string; month_zhi_wuxing: string
  day_gan_wuxing: string; day_zhi_wuxing: string
  hour_gan_wuxing: string; hour_zhi_wuxing: string
  // 藏干
  year_hide_gan: string[]; month_hide_gan: string[]
  day_hide_gan: string[]; hour_hide_gan: string[]
  
  // 十神和长生
  year_gan_shishen: string; month_gan_shishen: string; day_gan_shishen: string; hour_gan_shishen: string;
  year_zhi_shishen: string[]; month_zhi_shishen: string[]; day_zhi_shishen: string[]; hour_zhi_shishen: string[];
  year_di_shi: string; month_di_shi: string; day_di_shi: string; hour_di_shi: string;
  year_xing_yun: string; month_xing_yun: string; day_xing_yun: string; hour_xing_yun: string;
  year_xun_kong: string; month_xun_kong: string; day_xun_kong: string; hour_xun_kong: string;
  year_shen_sha: string[]; month_shen_sha: string[]; day_shen_sha: string[]; hour_shen_sha: string[];
  // 纳音
  year_na_yin: string; month_na_yin: string
  day_na_yin: string; hour_na_yin: string
  // 真太阳时
  true_solar_hour: number; true_solar_minute: number
  wuxing: { mu: number; huo: number; tu: number; jin: number; shui: number }
  yongshen: string; jishen: string
  tiaohou?: {
    expected: string[]
    tou: string[]
    cang: string[]
    text: string
  }
  // 交运时间
  start_yun_solar: string;
  dayun: Array<{
    index: number;
    gan: string;
    zhi: string;
    start_age: number;
    start_year: number;
    end_year: number;
    gan_shishen: string;
    zhi_shishen: string;
    di_shi: string;
    jin_bu_huan?: { qian_level: string; qian_desc: string; hou_level: string; hou_desc: string; verse: string } | null;
    liu_nian: Array<{
      year: number;
      age: number;
      gan_zhi: string;
      gan_shishen: string;
      zhi_shishen: string;
    }>;
  }>
  birth_year: number; birth_month: number; birth_day: number; birth_hour: number; gender: string
  // 命格
  ming_ge?: string
  ming_ge_desc?: string
  ten_god_relation?: TenGodRelationMatrix
  gong_jia?: GongJiaItem[]
  natal_assessment?: NatalAssessment
  vehicle_profile?: VehicleProfile
  wealth_profile?: WealthProfile
  dayun_roadmap?: DayunRoad[]
}

const VEHICLE_GRADE_GUIDE = [
  { grade: 'S', label: '上格配置', vehicle: '超跑级座驾', summary: '扶抑、格局与流通协同。', detail: '调候急需已解或并不急，扶抑用神得力，格局制化和原局流通也能承接。' },
  { grade: 'A', label: '中上格配置', vehicle: '高性能车级座驾', summary: '主线成立，局部仍有瑕疵。', detail: '扶抑可用，格局大体成立；用神力量、制化或流通仍有局部限制。' },
  { grade: 'B', label: '中格配置', vehicle: '标准轿车级座驾', summary: '原局可用，短板明确。', detail: '扶抑有支撑，但格局仅部分成立或流通有限，顺运时更容易发挥。' },
  { grade: 'C', label: '中下格配置', vehicle: '实用 MPV 级座驾', summary: '关键条件至少一项不足。', detail: '调候急需未解、扶抑支撑不足，或格局出现关键破损，基础发挥更依赖大运补足。' },
  { grade: 'D', label: '下格配置', vehicle: '基础代步单车级', summary: '原局需要优先补救短板。', detail: '调候急需与扶抑不足同时存在，且结构承载薄弱，更需要后天调整和顺运支持。' },
] as const

const ROAD_GUIDE = [
  { type: 'highway', label: '高速路', summary: '外部支持更集中，适合推进重点目标。' },
  { type: 'main_road', label: '城市主路', summary: '总体较顺畅，按节奏执行更容易见效。' },
  { type: 'mountain_road', label: '山路', summary: '有空间也有弯道，更重选择与节奏。' },
  { type: 'muddy_road', label: '泥路', summary: '阻力较多，适合稳住、调整并减少冒进。' },
  { type: 'construction', label: '施工路段', summary: '变化或修整较强，适合先处理基础问题。' },
] as const

const VEHICLE_GRADE_TAGS = new Set([
  ...VEHICLE_GRADE_GUIDE.map(item => item.label),
  '顶配', '高配', '实用', '偏科', '需调校',
])

// Older saved charts stored a Ming Ge-derived vehicle type. The final grade is
// now the sole source of the primary vehicle class, including for old snapshots.
const LEGACY_VEHICLE_TYPE_TAGS = new Set([
  '基础通勤型',
  '高扭矩越野型',
  '稳定商务型',
  '灵感跑车型',
  '资源运营型',
  '重载工程型',
  '均衡通勤型',
])

function getShiShen(dayGan: string, targetGan: string) {
  const dayWx = GAN_WUXING[dayGan]
  const targetWx = GAN_WUXING[targetGan]
  if (!dayWx || !targetWx) return ''

  const sameYinyang = GAN_YINYANG[dayGan] === GAN_YINYANG[targetGan]
  if (dayWx === targetWx) return sameYinyang ? '比肩' : '劫财'
  if (WUXING_SHENG[dayWx] === targetWx) return sameYinyang ? '食神' : '伤官'
  if (WUXING_KE[dayWx] === targetWx) return sameYinyang ? '偏财' : '正财'
  if (WUXING_KE[targetWx] === dayWx) return sameYinyang ? '七杀' : '正官'
  if (WUXING_SHENG[targetWx] === dayWx) return sameYinyang ? '偏印' : '正印'
  return ''
}

function relationFromTenGod(tenGod: string) {
  return TEN_GOD_META[tenGod] || { relation: '', group: '', group_label: '', summary: '' }
}

function tenGodSummary(tenGod: string) {
  return relationFromTenGod(tenGod).summary
}

function buildStemRelation(
  dayGan: string,
  pillar: string,
  pillarLabel: string,
  gan: string,
  wuxing: string,
  tenGod: string,
  isDayMaster = false,
): TenGodRelationItem {
  if (isDayMaster) {
    return {
      pillar,
      pillar_label: pillarLabel,
      gan,
      wuxing: wuxing || GAN_WUXING[gan] || '',
      ten_god: '日主 / 日元',
      relation: '命主自身',
      summary: '这是命盘的参照点，其他十神都以此天干为中心推导。',
    }
  }
  const resolvedTenGod = tenGod || getShiShen(dayGan, gan)
  const meta = relationFromTenGod(resolvedTenGod)
  return {
    pillar,
    pillar_label: pillarLabel,
    gan,
    wuxing: wuxing || GAN_WUXING[gan] || '',
    ten_god: resolvedTenGod,
    group: meta.group,
    group_label: meta.group_label,
    relation: meta.relation,
    summary: tenGodSummary(resolvedTenGod),
  }
}

function buildHiddenStemGroup(
  dayGan: string,
  pillar: string,
  pillarLabel: string,
  branch: string,
  hiddenGans: string[],
): TenGodHiddenStemGroup {
  return {
    pillar,
    pillar_label: pillarLabel,
    branch,
    items: hiddenGans
      .map((gan) => {
        const tenGod = getShiShen(dayGan, gan)
        const meta = relationFromTenGod(tenGod)
        return {
          gan,
          wuxing: GAN_WUXING[gan] || '',
          ten_god: tenGod,
          group: meta.group,
          group_label: meta.group_label,
          relation: meta.relation,
          summary: tenGodSummary(tenGod),
        }
      })
      .filter(item => item.ten_god),
  }
}

function buildTenGodRelationMatrix(result: BaziResult): TenGodRelationMatrix {
  return {
    day_master: {
      gan: result.day_gan,
      wuxing: result.day_gan_wuxing || GAN_WUXING[result.day_gan] || '',
      label: `${result.day_gan}${result.day_gan_wuxing || GAN_WUXING[result.day_gan] || ''}`,
    },
    heavenly_stems: [
      buildStemRelation(result.day_gan, 'year', '年干', result.year_gan, result.year_gan_wuxing, result.year_gan_shishen),
      buildStemRelation(result.day_gan, 'month', '月干', result.month_gan, result.month_gan_wuxing, result.month_gan_shishen),
      buildStemRelation(result.day_gan, 'day', '日干', result.day_gan, result.day_gan_wuxing, result.day_gan_shishen, true),
      buildStemRelation(result.day_gan, 'hour', '时干', result.hour_gan, result.hour_gan_wuxing, result.hour_gan_shishen),
    ],
    hidden_stems: [
      buildHiddenStemGroup(result.day_gan, 'year', '年支', result.year_zhi, result.year_hide_gan || []),
      buildHiddenStemGroup(result.day_gan, 'month', '月支', result.month_zhi, result.month_hide_gan || []),
      buildHiddenStemGroup(result.day_gan, 'day', '日支', result.day_zhi, result.day_hide_gan || []),
      buildHiddenStemGroup(result.day_gan, 'hour', '时支', result.hour_zhi, result.hour_hide_gan || []),
    ],
  }
}

type ReadingMode = 'simple' | 'professional'
type OverviewModalKind = 'grade' | 'road' | 'vehicle-evidence' | 'road-evidence' | 'wealth-evidence'
type DayunPeriod = BaziResult['dayun'][number]

const WUXING_LABELS: Record<keyof BaziResult['wuxing'], string> = {
  mu: '木',
  huo: '火',
  tu: '土',
  jin: '金',
  shui: '水',
}

const WUXING_KEYS = ['mu', 'huo', 'tu', 'jin', 'shui'] as const

function scrollToResultSection(id: string) {
  document.getElementById(id)?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

function getCurrentDayun(dayun: DayunPeriod[] = []) {
  const currentYear = new Date().getFullYear()
  return dayun.find(item => currentYear >= item.start_year && currentYear <= item.end_year) || dayun[0] || null
}

function getCurrentDayunIndex(dayun: DayunPeriod[] = []) {
  const currentYear = new Date().getFullYear()
  const current = dayun.find(item => currentYear >= item.start_year && currentYear <= item.end_year)
  return current?.index ?? dayun[0]?.index ?? null
}

function findDayunRoad(roadmap: DayunRoad[] | undefined, dayun: DayunPeriod | null) {
  if (!roadmap?.length || !dayun) return null
  return roadmap.find(item => item.dayun_index === dayun.index) ?? null
}

function formatEvidenceDelta(delta: number) {
  if (delta > 0) return `+${delta}`
  return String(delta)
}

function findVehicleGradeGuide(grade: string | undefined) {
  return VEHICLE_GRADE_GUIDE.find(item => item.grade === grade)
}

function findRoadGuide(roadType: string | undefined) {
  return ROAD_GUIDE.find(item => item.type === roadType)
}

function getWuxingExtremes(wuxing: BaziResult['wuxing']) {
  const items = WUXING_KEYS
    .map((key) => ({ key, label: WUXING_LABELS[key], value: Number(wuxing[key] ?? 0) }))
    .sort((a, b) => b.value - a.value)
  return {
    strongest: items[0],
    weakest: items[items.length - 1],
  }
}

function getFuyiYongshen(result: BaziResult) {
  return result.natal_assessment?.fuyi.yongshen || result.yongshen || ''
}

function getFuyiJishen(result: BaziResult) {
  return result.natal_assessment?.fuyi.jishen || result.jishen || ''
}

function buildStemLevelYongshenSummary(result: BaziResult): StemLevelYongshenSummary | null {
  const guidance = result.natal_assessment?.stem_guidance
  if (!guidance) return null

  const primary = guidance.primary_favorable[0]
  const usable = guidance.secondary_favorable
  const adverse = guidance.adverse
  if (primary || usable.length > 0 || adverse.length > 0) {
    return { primary, usable, adverse, conditioning_reference: [] }
  }
  if (guidance.conditioning_only.length > 0) {
    return { usable: [], adverse: [], conditioning_reference: guidance.conditioning_only }
  }
  return null
}

function formatStemGuidanceItems(items: NatalStemGuidanceItem[]) {
  return items.map(item => `${item.stem}${item.element}`).join('、')
}

function StemGuidanceInlineList({ items }: { items: NatalStemGuidanceItem[] }) {
  return (
    <span className="result-stem-summary-list">
      {items.map((item, index) => (
        <Fragment key={item.stem}>
          {index > 0 && '、'}
          <strong className={`wuxing-text-${WUXING_MAP[item.element] || 'jin'}`}>{item.stem}{item.element}</strong>
        </Fragment>
      ))}
    </span>
  )
}

function StemLevelYongshenSummaryPanel({ summary }: { summary: StemLevelYongshenSummary }) {
  return (
    <div className="result-stem-summary" aria-label="天干级喜忌摘要">
      {summary.primary && (
        <div className="result-stem-summary-row result-stem-summary-row--primary">
          <span className="result-stem-summary-label">首取</span>
          <strong className={`wuxing-text-${WUXING_MAP[summary.primary.element] || 'jin'}`}>
            {summary.primary.stem}{summary.primary.element}
          </strong>
          {summary.primary.ten_god && <em>{summary.primary.ten_god}</em>}
          <small>{summary.primary.source_layers.join(' + ')}</small>
        </div>
      )}
      {summary.usable.length > 0 && (
        <div className="result-stem-summary-row">
          <span className="result-stem-summary-label">扶抑可用</span>
          <StemGuidanceInlineList items={summary.usable} />
        </div>
      )}
      {summary.adverse.length > 0 && (
        <div className="result-stem-summary-row result-stem-summary-row--adverse">
          <span className="result-stem-summary-label">扶抑慎用</span>
          <StemGuidanceInlineList items={summary.adverse} />
        </div>
      )}
      {summary.conditioning_reference.length > 0 && (
        <div className="result-stem-summary-row result-stem-summary-row--conditioning">
          <span className="result-stem-summary-label">调候参考</span>
          <StemGuidanceInlineList items={summary.conditioning_reference} />
          <small>原局调候结构，不等同后天通用喜神</small>
        </div>
      )}
    </div>
  )
}

function buildChartKeywords(result: BaziResult, relation: TenGodRelationMatrix) {
  const groups = relation.heavenly_stems
    .map(item => item.group_label)
    .filter((label): label is string => Boolean(label))
  const uniqueGroups = Array.from(new Set(groups)).slice(0, 2)
  const stemSummary = buildStemLevelYongshenSummary(result)
  if (stemSummary) {
    return [
      result.ming_ge,
      stemSummary.primary ? `首取${stemSummary.primary.stem}${stemSummary.primary.element}` : '',
      stemSummary.usable.length > 0 ? `扶抑可用${formatStemGuidanceItems(stemSummary.usable)}` : '',
      stemSummary.adverse.length > 0 ? `扶抑慎用${formatStemGuidanceItems(stemSummary.adverse)}` : '',
      ...uniqueGroups,
    ].filter(Boolean).slice(0, 5)
  }
  const yongshen = getFuyiYongshen(result)
  const jishen = getFuyiJishen(result)
  return [
    result.ming_ge,
    yongshen ? `扶抑喜${yongshen}` : '',
    jishen ? `扶抑忌${jishen}` : '',
    ...uniqueGroups,
  ].filter(Boolean).slice(0, 5)
}

function buildChartVerdict(result: BaziResult, relation: TenGodRelationMatrix) {
  const dayMaster = relation.day_master.label || `${result.day_gan}${result.day_gan_wuxing || ''}`
  const stemSummary = buildStemLevelYongshenSummary(result)
  const minggeText = result.ming_ge ? `，格局线索为「${result.ming_ge}」` : ''
  if (stemSummary) {
    const yongshenText = stemSummary.primary
      ? `首取${stemSummary.primary.stem}${stemSummary.primary.element}，扶抑可用${formatStemGuidanceItems(stemSummary.usable) || '待结合全局判断'}${stemSummary.adverse.length > 0 ? `，慎用${formatStemGuidanceItems(stemSummary.adverse)}` : ''}`
      : `扶抑可用${formatStemGuidanceItems(stemSummary.usable) || '待结合全局判断'}${stemSummary.adverse.length > 0 ? `，慎用${formatStemGuidanceItems(stemSummary.adverse)}` : ''}`
    return `日主${dayMaster}为命局参照，${yongshenText}${minggeText}。建议先看当前大运与用神依据，再进入专业命盘和完整命理解读。`
  }
  const yongshen = getFuyiYongshen(result)
  const jishen = getFuyiJishen(result)
  const yongshenText = yongshen
    ? `扶抑喜用${yongshen}${jishen ? `，忌${jishen}` : ''}`
    : '喜忌待结合解读确认'
  return `日主${dayMaster}为命局参照，${yongshenText}${minggeText}。建议先看当前大运与用神依据，再进入专业命盘和完整命理解读。`
}

function buildYongshenEvidence(result: BaziResult, relation: TenGodRelationMatrix) {
  const { strongest, weakest } = getWuxingExtremes(result.wuxing)
  const tenGodGroups = Array.from(new Set(
    relation.heavenly_stems
      .map(item => item.group_label)
      .filter((label): label is string => Boolean(label)),
  ))
  const tiaohouExpected = result.tiaohou?.expected?.filter(Boolean) || []
  const tiaohouHit = [...(result.tiaohou?.tou || []), ...(result.tiaohou?.cang || [])].filter(Boolean)
  const primaryStem = result.natal_assessment?.stem_guidance?.primary_favorable?.[0]

  const evidences = [
    {
      title: '日主与月令',
      detail: `以${result.day_gan}${result.day_zhi}日柱为参照，结合${result.month_zhi}月令判断命局底色。`,
    },
    {
      title: '五行分布',
      detail: strongest && weakest
        ? `${strongest.label}相对更显，${weakest.label}相对不足，用神展示用于帮助观察命局平衡。`
        : '结合五行分布观察命局偏性与补益方向。',
    },
    {
      title: '调候线索',
      detail: tiaohouExpected.length > 0
        ? `调候优先参考${tiaohouExpected.join('、')}${tiaohouHit.length ? `，命盘中可见${tiaohouHit.join('、')}` : '，命盘显现程度需结合藏干与透干观察'}。`
        : '当前命盘未展示明确调候命中项，优先结合五行与十神结构阅读。',
    },
    {
      title: '十神结构',
      detail: tenGodGroups.length > 0
        ? `天干关系中可见${tenGodGroups.join('、')}等线索，专业模式可继续查看每柱关系。`
        : '专业模式可查看天干与藏干相对日主的十神关系。',
    },
  ]
  if (primaryStem) {
    evidences.splice(3, 0, {
      title: '天干优先',
      detail: `${primaryStem.stem}${primaryStem.element}为共同优先，兼具${primaryStem.source_layers.join('与')}方向，对应${primaryStem.ten_god}；其余天干需结合扶抑与调候分别理解。`,
    })
  }
  return evidences
}

function StemGuidancePanel({ guidance }: { guidance?: NatalStemGuidance }) {
  if (!guidance) return null
  const groups = [
    { key: 'primary', title: '主喜', note: '日干调候与扶抑共同支持', items: guidance.primary_favorable, tone: 'primary' },
    { key: 'secondary', title: '扶抑可用', note: '符合扶抑方向', items: guidance.secondary_favorable, tone: 'secondary' },
    { key: 'conditioning', title: '调候结构', note: '原局调候参考，不等同后天通用喜神', items: guidance.conditioning_only, tone: 'conditioning' },
    { key: 'adverse', title: '扶抑忌神', note: '后天遇此天干需结合全局谨慎判断', items: guidance.adverse, tone: 'adverse' },
  ] as const
  const hasGuidance = groups.some(group => group.items.length > 0)
  if (!hasGuidance) return null

  return (
    <section className="stem-guidance-panel" aria-labelledby="stem-guidance-title">
      <div className="stem-guidance-heading">
        <div>
          <span>天干级喜忌</span>
          <h3 id="stem-guidance-title" className="serif">天干优先</h3>
        </div>
        <p>五行说明方向，天干说明具体落点。</p>
      </div>
      <div className="stem-guidance-groups">
        {groups.filter(group => group.items.length > 0).map(group => (
          <section key={group.key} className={`stem-guidance-group stem-guidance-group--${group.tone}`}>
            <div className="stem-guidance-group-heading">
              <strong>{group.title}</strong>
              <span>{group.note}</span>
            </div>
            <ul>
              {group.items.map(item => (
                <li key={`${group.key}-${item.stem}`}>
                  <div className="stem-guidance-stem">
                    <strong className={`wuxing-text-${WUXING_MAP[item.element] || 'jin'}`}>{item.stem}{item.element}</strong>
                    {item.ten_god && <span>{item.ten_god}</span>}
                    <em>{item.source_layers.join(' + ')}</em>
                  </div>
                  <p>{item.detail}</p>
                </li>
              ))}
            </ul>
          </section>
        ))}
      </div>
    </section>
  )
}


export default function ResultPage() {
  const location = useLocation()
  const { id } = useParams()
  const navigate = useNavigate()
  const { user } = useAuth()
  const { showToast } = useToast()

  const [result, setResult] = useState<BaziResult | null>(location.state?.result || null)
  const [report, setReport] = useState<AIReport | null>(location.state?.report || null)
  const [isGuest] = useState(location.state?.isGuest ?? !user)
  const [registration_enabled, setRegistrationEnabled] = useState(false)
  const [loading, setLoading] = useState(!result && !!id)
  const [reportMode, setReportMode] = useState<'brief' | 'detail'>('detail')
  const [readingMode, setReadingMode] = useState<ReadingMode>('simple')
  const [overviewModal, setOverviewModal] = useState<OverviewModalKind | null>(null)
  const [trendDayunIndex, setTrendDayunIndex] = useState<number | null>(null)
  const [activeTrendYear, setActiveTrendYear] = useState<number | null>(null)
  const [reportTab, setReportTab] = useState<'original' | 'polished'>('original')
  const reportTabRowRef = useRef<HTMLDivElement | null>(null)
  const overviewModalDialogRef = useRef<HTMLDivElement | null>(null)
  const overviewModalTriggerRef = useRef<HTMLButtonElement | null>(null)

  const openOverviewModal = (kind: OverviewModalKind, trigger: HTMLButtonElement) => {
    overviewModalTriggerRef.current = trigger
    setOverviewModal(kind)
  }

  const closeOverviewModal = () => setOverviewModal(null)

  useEffect(() => {
    if (!overviewModal) return

    const previousOverflow = document.body.style.overflow
    const focusDialog = window.requestAnimationFrame(() => overviewModalDialogRef.current?.focus())
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        event.preventDefault()
        closeOverviewModal()
      }
    }

    document.body.style.overflow = 'hidden'
    window.addEventListener('keydown', handleKeyDown)

    return () => {
      window.cancelAnimationFrame(focusDialog)
      document.body.style.overflow = previousOverflow
      window.removeEventListener('keydown', handleKeyDown)
      overviewModalTriggerRef.current?.focus()
    }
  }, [overviewModal])

  // 切换原版/润色版后把视口拉回报告区顶部，避免停留在另一版的中部位置
  const switchReportTab = (tab: 'original' | 'polished') => {
    setReportTab(tab)
    requestAnimationFrame(() => {
      reportTabRowRef.current?.scrollIntoView({ behavior: 'smooth', block: 'start' })
    })
  }
  const [polishedReport, setPolishedReport] = useState<PolishedReport | null>(null)
  const [polishing, setPolishing] = useState(false)
  const [polishError, setPolishError] = useState<string | null>(null)
  const [savingImage, setSavingImage] = useState(false)
  const [exportingPDF, setExportingPDF] = useState(false)
  const [chartDisplayNameDraft, setChartDisplayNameDraft] = useState(location.state?.result?.display_name || '')
  const [chartDisplayNameError, setChartDisplayNameError] = useState('')
  const shareCardRef = useRef<HTMLDivElement>(null)
  const pendingIntentConsumedRef = useRef(false)
  const pendingInput = location.state?.input as CalculateInput | undefined
  // 页面当前命盘 id：历史 URL、起盘 state 都会汇入这里。
  const targetId = id || location.state?.chartId

  useEffect(() => {
    setTrendDayunIndex(null)
    setActiveTrendYear(null)
  }, [targetId, result?.birth_year, result?.birth_month, result?.birth_day, result?.birth_hour, result?.gender])

  useEffect(() => {
    if (activeTrendYear == null) return
    const previousOverflow = document.body.style.overflow
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        setActiveTrendYear(null)
      }
    }

    document.body.style.overflow = 'hidden'
    window.addEventListener('keydown', handleKeyDown)

    return () => {
      document.body.style.overflow = previousOverflow
      window.removeEventListener('keydown', handleKeyDown)
    }
  }, [activeTrendYear])

  // 神煞注解状态
  const [shenshaMap, setShenshaMap] = useState<Map<string, ShenshaAnnotation>>(new Map())
  const [activeAnnotation, setActiveAnnotation] = useState<ShenshaAnnotation | null>(null)
  const [activeMingGe, setActiveMingGe] = useState<{ name: string; desc: string } | null>(null)
  const [historicalFigures, setHistoricalFigures] = useState<MingGeHistoricalFigure[]>([])
  const hoverTimer = useRef<ReturnType<typeof setTimeout> | null>(null)

  // 导出品牌定制
  const [brand, setBrand] = useState<ExportBrand | null>(null)
  const [pastEventsExportSegments, setPastEventsExportSegments] = useState<PastEventsExportReadySegment[]>([])

  // 预加载神煞注解
  useEffect(() => {
    fetchShenshaAnnotations()
      .then(list => {
        const map = new Map<string, ShenshaAnnotation>()
        list.forEach(a => map.set(a.name, a))
        setShenshaMap(map)
      })
      .catch(err => {
        // 注解加载失败不影响主功能，留痕便于线上排查
        console.warn('[shensha] 注解加载失败', err)
      })
  }, [])

  useEffect(() => {
    const mingGe = result?.ming_ge?.trim()
    if (!mingGe) {
      setHistoricalFigures([])
      return
    }

    let cancelled = false
    baziAPI.getMingGeHistoricalFigures(mingGe)
      .then((response) => {
        if (!cancelled) setHistoricalFigures(response.data.data || [])
      })
      .catch(() => {
        // 古人映照是补充阅读，读取失败不能影响排盘主结果。
        if (!cancelled) setHistoricalFigures([])
      })

    return () => { cancelled = true }
  }, [result?.ming_ge])

  useEffect(() => {
    if (!isGuest) return
    authAPI.registrationSettings()
      .then(r => setRegistrationEnabled(r.data.registration_enabled))
      .catch(() => setRegistrationEnabled(true))
  }, [isGuest])

  // 加载用户导出品牌定制
  useEffect(() => {
    if (!user) return
    brandAPI.get()
      .then(r => setBrand(r.data.data))
      .catch(() => setBrand(null))
  }, [user])

  const handleSaveImage = async () => {
    if (!shareCardRef.current) return
    if (!report?.content_structured && !hasMeaningfulReportContent(report?.content)) {
      showToast('当前报告没有可保存的正文，请先重新生成命理解读', 'error')
      return
    }
    setSavingImage(true)

    const isMobile = /iPhone|iPad|iPod|Android/i.test(navigator.userAgent)
    const isIOS = /iPhone|iPad|iPod/i.test(navigator.userAgent)

    try {
      await loadPastEventsExportForPrint()
      await waitForPrintLayoutUpdate()

      // 导出库较大，点击时才加载
      const { toPng, toBlob } = await import('html-to-image')
      await document.fonts.ready

      if (isIOS) {
        // iOS 最佳方案：Web Share API + File Blob
        // 调起系统原生分享面板，用户可直接选“存储图像”保存到相册
        const blob = await toBlob(shareCardRef.current, {
          quality: 0.98,
          pixelRatio: 3,
          cacheBust: true,
        })
        if (!blob) throw new Error('生成图片失败')

        const fileName = `缘聚命理-${result?.year_gan ?? ''}年${result?.month_gan ?? ''}月.png`
        const file = new File([blob], fileName, { type: 'image/png' })

        if (navigator.canShare && navigator.canShare({ files: [file] })) {
          // 支持 Web Share API（iOS 15+ Safari 全款支持）
          await navigator.share({
            files: [file],
            title: `缘聚命理 · 八字命理报告`,
            text: `我的八字命理：${result?.year_gan ?? ''}${result?.year_zhi ?? ''}年`,
          })
        } else {
          // 退化到 Blob Object URL——比 base64 更靠谱
          const objectUrl = URL.createObjectURL(blob)
          Object.assign(document.createElement('a'), {
            href: objectUrl, download: fileName,
          }).click()
          setTimeout(() => URL.revokeObjectURL(objectUrl), 5000)
        }
      } else if (isMobile) {
        // Android：直接下载
        const blob = await toBlob(shareCardRef.current, {
          quality: 0.98, pixelRatio: 3, cacheBust: true,
        })
        if (!blob) throw new Error('生成图片失败')
        const objectUrl = URL.createObjectURL(blob)
        const fileName = `缘聚命理-${result?.year_gan ?? ''}年${result?.month_gan ?? ''}月.png`
        Object.assign(document.createElement('a'), {
          href: objectUrl, download: fileName,
        }).click()
        setTimeout(() => URL.revokeObjectURL(objectUrl), 5000)
      } else {
        // 桌面端：toPng + 下载
        const dataUrl = await toPng(shareCardRef.current, {
          quality: 0.98, pixelRatio: 2, cacheBust: true,
        })
        const link = document.createElement('a')
        link.download = `缘聚命理-${result?.year_gan ?? ''}年${result?.month_gan ?? ''}月.png`
        link.href = dataUrl
        link.click()
      }
    } catch (err: unknown) {
      // 用户主动取消分享不算错误
      const msg = err instanceof Error ? err.message : ''
      if (!msg.includes('AbortError') && !msg.includes('cancel')) {
        showToast('生成图片失败，请稍后重试', 'error')
      }
    } finally {
      setSavingImage(false)
    }
  }

  const isMobileDevice = /iPhone|iPad|iPod|Android/i.test(navigator.userAgent)

  const loadPastEventsExportForPrint = async () => {
    if (!targetId) {
      setPastEventsExportSegments([])
      return
    }
    try {
      const resp = await baziAPI.fetchPastEventsExport(targetId)
      const normalized = resp.data.segments.map((segment) => ({
        ...segment,
        themes: segment.themes || [],
      }))
      setPastEventsExportSegments(filterPastEventsExportSegments(normalized, false))
    } catch {
      setPastEventsExportSegments([])
    }
  }

  const waitForPrintLayoutUpdate = () => new Promise<void>((resolve) => {
    requestAnimationFrame(() => requestAnimationFrame(() => resolve()))
  })

  const handleExportPDF = async () => {
    // 导出相关提示用 toast 就近反馈，不写进 AI 解读区的 reportError
    if (reportTab === 'polished' && !polishedReport) {
      showToast('请先生成润色版命理解读，再导出 PDF 报告', 'error')
      return
    }
    if (reportTab !== 'polished' && !report) {
      showToast('请先生成命理解读，再导出 PDF 报告', 'error')
      return
    }
    if (reportTab !== 'polished' && !report?.content_structured && !hasMeaningfulReportContent(report?.content)) {
      showToast('当前报告没有可导出的正文，请先重新生成命理解读', 'error')
      return
    }
    setExportingPDF(true)
    let el: HTMLElement | null = null
    let prevDisplay = ''
    try {
      await loadPastEventsExportForPrint()
      await waitForPrintLayoutUpdate()
      if (!isMobileDevice) {
        window.print()
        return
      }
      // 移动端：用 html2canvas + jsPDF 生成 PDF 文件下载
      el = document.querySelector('.print-only') as HTMLElement | null
      if (!el) return
      prevDisplay = el.style.display
      // 导出库较大，点击时才加载
      const [{ default: html2canvas }, { default: jsPDF }] = await Promise.all([
        import('html2canvas'),
        import('jspdf'),
      ])
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
      const fileName = `缘聚命理-命书.pdf`
      pdf.save(fileName)
    } catch {
      showToast('生成 PDF 失败，请稍后重试', 'error')
    } finally {
      if (el) el.style.display = prevDisplay
      setExportingPDF(false)
    }
  }

  const persistGuestJourney = (intent: 'view_result' | 'generate_report' = 'generate_report') => {
    if (!isGuest || !pendingInput) return
    savePendingJourney(createPendingBaziJourney({
      input: pendingInput,
      anonymousChartId: targetId,
      displayLabel: result?.display_name || undefined,
      intent,
      returnPath: '/result',
    }))
  }

  // AI 解读状态
  const [reportLoading, setReportLoading] = useState(false)
  const [isStreaming, setIsStreaming] = useState(false)
  const [isThinking, setIsThinking] = useState(false)
  const [thinkingSeconds, setThinkingSeconds] = useState(0)
  const [streamingText, setStreamingText] = useState('')
  const [reportError, setReportError] = useState('')
  const [loadingStepIndex, setLoadingStepIndex] = useState(0)

  useEffect(() => {
    let timer: number
    if (reportLoading) {
      setLoadingStepIndex(0)
      timer = window.setInterval(() => {
        setLoadingStepIndex(prev => {
          if (prev < LOADING_STEPS.length - 1) return prev + 1
          return prev
        })
      }, 4000)
    } else {
      setLoadingStepIndex(0)
    }
    return () => window.clearInterval(timer)
  }, [reportLoading])

  // 推理计时器
  useEffect(() => {
    let timer: number
    if (isThinking) {
      setThinkingSeconds(0)
      timer = window.setInterval(() => {
        setThinkingSeconds(prev => prev + 1)
      }, 1000)
    }
    return () => window.clearInterval(timer)
  }, [isThinking])

  useEffect(() => {
    setChartDisplayNameDraft(result?.display_name || '')
  }, [result?.display_name])

  // 加载润色版报告
  useEffect(() => {
    if (!targetId || !user) return
    baziAPI.getPolishedReport(targetId)
      .then(res => setPolishedReport(res.data.polished_report || null))
      .catch(() => setPolishedReport(null))
  }, [targetId, user])

  // 历史详情始终以服务端结果为准，避免浏览器历史状态保留旧快照而跳过字段补齐。
  useEffect(() => {
    if (id) {
      baziAPI.getHistoryDetail(id)
        .then(res => {
          setResult(res.data.result || res.data.chart || null)
          setReport(res.data.report || null)
        })
        .catch(() => navigate('/history'))
        .finally(() => setLoading(false))
    }
  }, [id]) // eslint-disable-line react-hooks/exhaustive-deps

  // 点击"生成 AI 解读"按钮
  const handleGenerateReport = async () => {
    if (reportLoading || isStreaming || isThinking) return
    let reportTargetId = targetId
    if (!reportTargetId && pendingInput && !isGuest) {
      setReportLoading(true)
      setReportError('')
      try {
        const res = await baziAPI.calculate(pendingInput)
        reportTargetId = res.data.chart_id
        setResult(res.data.result || result)
        navigate(buildBaziResultRoute(reportTargetId, false) + window.location.hash, {
          replace: true,
          state: {
            result: res.data.result || result,
            chartId: reportTargetId,
            input: pendingInput,
            isGuest: false,
          },
        })
      } catch (err: unknown) {
        setReportError(err instanceof Error ? `保存命盘失败：${err.message}` : '保存命盘失败，请重新起盘后再生成解读。')
        setReportLoading(false)
        return
      }
    }
    if (!reportTargetId) {
      setReportError('未找到可保存的命盘记录，请返回首页重新起盘后再生成解读。')
      return
    }
    setReportLoading(true)
    setIsStreaming(false)
    setIsThinking(false)
    setStreamingText('')
    setReportError('')

    let currentText = ''
    let isFirstByte = true
    await baziAPI.generateReportStream(
      reportTargetId,
      (text) => {
        if (isFirstByte) {
          setReportLoading(false)
          setIsThinking(false)
          setIsStreaming(true)
          isFirstByte = false
        }
        currentText += text
        setStreamingText(currentText)
      },
      (err) => {
        setReportError(err)
        setIsStreaming(false)
        setIsThinking(false)
        setReportLoading(false)
      },
      () => {
        // 流结束：先保持 isStreaming=true 避免闪烁，等拉取完结构化数据后再统一切换
        baziAPI.getHistoryDetail(reportTargetId).then(res => {
          setResult(res.data.result || res.data.chart || null)
          setReport(res.data.report || null)
        }).catch(err => {
          console.error('Failed to fetch finished report', err)
        }).finally(() => {
          setIsStreaming(false)
          setIsThinking(false)
          setReportLoading(false)
        })
      },
      () => {
        // 推理模型进入思考阶段
        setIsThinking(true)
        setReportLoading(false) // 关闭普通 loading，显示 thinking UI
      }
    )
  }

  useEffect(() => {
    if (pendingIntentConsumedRef.current) return
    if (location.state?.pendingIntent !== 'generate_report') return
    if (!targetId || isGuest || report || reportLoading || isStreaming || isThinking) return
    pendingIntentConsumedRef.current = true
    handleGenerateReport()
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [location.state?.pendingIntent, targetId, isGuest, report, reportLoading, isStreaming, isThinking])

  const handleSaveChartDisplayName = async () => {
    if (!targetId || isGuest) return
    const nextName = chartDisplayNameDraft.trim()
    if (Array.from(nextName).length > 20) {
      setChartDisplayNameError('称呼不能超过20个字符')
      return
    }
    try {
      const res = await baziAPI.updateHistoryDisplayName(targetId, nextName)
      const savedName = res.data.data.display_name
      setResult(prev => prev ? { ...prev, display_name: savedName } : prev)
      setChartDisplayNameDraft(savedName)
      setChartDisplayNameError('')
    } catch (err: unknown) {
      setChartDisplayNameError(err instanceof Error ? err.message : '保存称呼失败')
    }
  }

  const submitPolish = async (userSituation: string) => {
    if (!targetId) return
    setPolishing(true)
    setPolishError(null)
    try {
      const res = await baziAPI.generatePolishedReport(targetId, userSituation)
      setPolishedReport(res.data.polished_report)
    } catch (e: unknown) {
      const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error || (e instanceof Error ? e.message : '润色失败，请稍后重试')
      setPolishError(msg)
    } finally {
      setPolishing(false)
    }
  }

  if (loading) return <LoadingSkeleton />
  if (!result) return <div className="page container" style={{ paddingTop: 120 }}>数据加载失败</div>

  const pillars = [
    { label: '年柱', gan: result.year_gan, zhi: result.year_zhi, ganWx: result.year_gan_wuxing, zhiWx: result.year_zhi_wuxing, hideGan: result.year_hide_gan || [], naYin: result.year_na_yin || '', ganShiShen: result.year_gan_shishen, zhiShiShen: result.year_zhi_shishen || [], diShi: result.year_di_shi, xingYun: result.year_xing_yun, xunKong: result.year_xun_kong, shenSha: result.year_shen_sha || [] },
    { label: '月柱', gan: result.month_gan, zhi: result.month_zhi, ganWx: result.month_gan_wuxing, zhiWx: result.month_zhi_wuxing, hideGan: result.month_hide_gan || [], naYin: result.month_na_yin || '', ganShiShen: result.month_gan_shishen, zhiShiShen: result.month_zhi_shishen || [], diShi: result.month_di_shi, xingYun: result.month_xing_yun, xunKong: result.month_xun_kong, shenSha: result.month_shen_sha || [] },
    { label: '日柱', gan: result.day_gan, zhi: result.day_zhi, ganWx: result.day_gan_wuxing, zhiWx: result.day_zhi_wuxing, hideGan: result.day_hide_gan || [], naYin: result.day_na_yin || '', ganShiShen: result.day_gan_shishen, zhiShiShen: result.day_zhi_shishen || [], diShi: result.day_di_shi, xingYun: result.day_xing_yun, xunKong: result.day_xun_kong, shenSha: result.day_shen_sha || [] },
    { label: '时柱', gan: result.hour_gan, zhi: result.hour_zhi, ganWx: result.hour_gan_wuxing, zhiWx: result.hour_zhi_wuxing, hideGan: result.hour_hide_gan || [], naYin: result.hour_na_yin || '', ganShiShen: result.hour_gan_shishen, zhiShiShen: result.hour_zhi_shishen || [], diShi: result.hour_di_shi, xingYun: result.hour_xing_yun, xunKong: result.hour_xun_kong, shenSha: result.hour_shen_sha || [] },
  ]
  const relation = result.ten_god_relation ?? buildTenGodRelationMatrix(result)
  const hiddenStemGroups = relation.hidden_stems.filter(group => group.items.length > 0)
  const dayPillarCellClass = (index: number) => index === 2 ? ' is-day-pillar-cell' : ''
  const dayunPillarsLabel = `${result.year_gan}${result.year_zhi} ${result.month_gan}${result.month_zhi} ${result.day_gan}${result.day_zhi} ${result.hour_gan}${result.hour_zhi}`
  const currentYear = new Date().getFullYear()
  const currentDayun = getCurrentDayun(result.dayun)
  const currentLiuNian = currentDayun?.liu_nian?.find(item => item.year === currentYear)
  const currentDayunIndex = getCurrentDayunIndex(result.dayun)
  const selectedDayunIndex = trendDayunIndex ?? currentDayunIndex
  const selectedDayun = result.dayun.find(item => item.index === selectedDayunIndex) || currentDayun || result.dayun[0] || null
  const currentDayunRoad = findDayunRoad(result.dayun_roadmap, currentDayun)
  const currentRoadEvidences = currentDayunRoad?.evidences ?? []
  const currentRoadPhaseEvidences = currentDayunRoad?.phase_evidences ?? []
  const hasCurrentRoadEvidence = currentRoadEvidences.length > 0 || currentRoadPhaseEvidences.length > 0
  const vehicleEvidences = result.vehicle_profile?.evidences ?? []
  const vehicleDrivingStyle = result.vehicle_profile?.driving_style
  const wealthProfile = result.wealth_profile
  const wealthEvidences = wealthProfile?.evidences ?? []
  const wealthTags = wealthProfile?.tags ?? []
  const wealthRisks = wealthProfile?.risk_flags ?? []
  const wealthCurrentHint = wealthProfile?.current_hint?.year === currentYear ? wealthProfile.current_hint : null
  const stemGuidance = result.natal_assessment?.stem_guidance
  const stemLevelYongshenSummary = buildStemLevelYongshenSummary(result)
  const vehicleGradeGuide = findVehicleGradeGuide(result.vehicle_profile?.grade)
  const currentRoadGuide = findRoadGuide(currentDayunRoad?.road_type)
  const overviewModalTitle = overviewModal === 'grade'
    ? '等级说明'
    : overviewModal === 'road'
      ? '大运路况说明'
      : overviewModal === 'vehicle-evidence'
        ? '座驾依据'
        : overviewModal === 'road-evidence'
          ? '路况依据'
          : overviewModal === 'wealth-evidence'
            ? '财富依据'
            : ''
  const overviewModalKicker = overviewModal === 'grade' || overviewModal === 'vehicle-evidence'
    ? '命盘座驾'
    : overviewModal === 'wealth-evidence'
      ? '财富结构'
      : '当前路况'
  const vehicleType = vehicleGradeGuide?.vehicle || result.vehicle_profile?.vehicle_type || ''
  const vehicleProfileTags = result.vehicle_profile?.tags?.filter(tag => (
    !VEHICLE_GRADE_TAGS.has(tag)
    && !LEGACY_VEHICLE_TYPE_TAGS.has(tag)
    && tag !== vehicleType
    && tag !== result.ming_ge
  )) ?? []
  const selectedDayunTrendSeries = buildDayunTrendSeries(selectedDayun, result.gender)
  const isSelectedCurrentDayun = Boolean(selectedDayun && currentDayun && selectedDayun.index === currentDayun.index)
  const selectedDayunTitle = selectedDayun ? `${selectedDayun.gan}${selectedDayun.zhi}大运` : '大运趋势'
  const selectedDayunPeriod = selectedDayun
    ? `${selectedDayun.start_age} - ${selectedDayun.start_age + 9} 岁 · 公历 ${selectedDayun.start_year} - ${selectedDayun.end_year}`
    : ''
  const activeTrendYearItems = activeTrendYear == null
    ? []
    : selectedDayunTrendSeries.flatMap(series => {
      const point = series.points.find(item => item.year === activeTrendYear)
      return point ? [{ series, point }] : []
    })
  const activeTrendYearMeta = activeTrendYearItems[0]?.point
  const activeTrendYearSummary = activeTrendYearItems.length
    ? (() => {
      const good = activeTrendYearItems.filter(item => item.point.level === 3).length
      const watch = activeTrendYearItems.filter(item => item.point.level === 1).length
      if (good >= 2 && watch >= 2) return '机会与压力并存，适合分清主次后推进。'
      if (good >= 3) return '顺势维度较多，适合主动推进关键事务。'
      if (watch >= 3) return '需留意维度较多，建议放缓节奏并先稳风险。'
      if (watch >= 1) return '整体有推进空间，但部分事项需要提前留意。'
      return '整体较平稳，适合按既定节奏推进。'
    })()
    : ''
  const chartKeywords = buildChartKeywords(result, relation)
  const chartVerdict = buildChartVerdict(result, relation)
  const yongshenEvidence = buildYongshenEvidence(result, relation)
  const fuyiYongshen = getFuyiYongshen(result)
  const fuyiJishen = getFuyiJishen(result)
  const resultQuickSummary = stemLevelYongshenSummary
    ? `日主${result.day_gan}${result.day_zhi}为命局参照，${stemLevelYongshenSummary.primary ? `首取${stemLevelYongshenSummary.primary.stem}${stemLevelYongshenSummary.primary.element}` : '以扶抑方向为主'}${stemLevelYongshenSummary.usable.length > 0 ? `，扶抑可用${formatStemGuidanceItems(stemLevelYongshenSummary.usable)}` : ''}${stemLevelYongshenSummary.adverse.length > 0 ? `，慎用${formatStemGuidanceItems(stemLevelYongshenSummary.adverse)}` : ''}。先确认四柱主盘与命局结构，再阅读命理解读和大运走势。`
    : fuyiYongshen || fuyiJishen
      ? `日主${result.day_gan}${result.day_zhi}为命局参照，扶抑喜用${fuyiYongshen || '待生成'}${fuyiJishen ? `，忌${fuyiJishen}` : ''}。先确认四柱主盘与命局结构，再阅读命理解读和大运走势。`
    : `日主${result.day_gan}${result.day_zhi}为命局参照。先确认四柱主盘，再查看五行结构、大运走势与命理解读。`

  const structured = report?.content_structured ?? null
  const reportDigestItems = structured ? buildReportDigestItems(structured, result) : []
  // 旧报告降级：解析纯文字 content
  const legacyReportContent = structured ? '' : cleanLegacyReportContent(report?.content || '')
  const reportSections = structured ? [] : parseReport(report?.content || '')
  const hasLegacyReportContent = hasMeaningfulReportContent(legacyReportContent)
  const hasRenderableOriginalReport = Boolean(structured || reportSections.length > 0 || hasLegacyReportContent)
  const canExportActiveReport = reportTab === 'polished' ? Boolean(polishedReport) : hasRenderableOriginalReport
  const resultSegments = readingMode === 'professional'
    ? [
      { id: 'result-section-overview', label: '总览' },
      { id: 'result-section-chart', label: '命盘' },
      { id: 'result-section-yongshen', label: '用神' },
      { id: 'result-section-dayun', label: '大运' },
      { id: 'result-section-ai', label: 'AI 解读' },
    ]
    : [
      { id: 'result-section-overview', label: '总览' },
      { id: 'result-section-current', label: '当前阶段' },
      { id: 'result-section-trend', label: '趋势' },
      { id: 'result-section-ai', label: 'AI 解读' },
    ]
  const historicalFiguresSection = result.ming_ge && historicalFigures.length > 0 ? (
    <section className="result-historical-figures" aria-labelledby="historical-figures-title">
      <div className="result-historical-figures-heading">
        <div>
          <span>命格参考</span>
          <h2 id="historical-figures-title" className="serif">{result.ming_ge} · 古人映照</h2>
        </div>
      </div>
      <div className="result-historical-figures-list">
        {historicalFigures.map((figure) => (
          <article key={figure.id} className="result-historical-figure">
            <div className="result-historical-figure-title">
              <h3 className="serif">{figure.figure_name}</h3>
              <span>{figure.era} · {figure.identity}</span>
            </div>
            <p>{figure.historical_memory}</p>
            {readingMode === 'professional' && (
              <div className="result-historical-evidence">
                <p><strong>历史资料：</strong><a href={figure.source_url} target="_blank" rel="noreferrer">{figure.source_title}</a></p>
                {figure.turning_point && <p><strong>人生转折：</strong>{figure.turning_point}{figure.turning_point_year ? `（${figure.turning_point_year}）` : ''}</p>}
                {figure.show_dayun && figure.dayun_period && figure.dayun_explanation && (
                  <p><strong>大运呼应：</strong>{figure.dayun_period}，{figure.dayun_explanation}</p>
                )}
              </div>
            )}
          </article>
        ))}
      </div>
      <p className="result-historical-boundary">古人映照用于理解格局可能的发挥方向，不代表你与其相似，也不构成个人结果预测。</p>
    </section>
  ) : null

  return (
    <>
      <div className="result-page page screen-only">
        <div className="container">

        <section id="result-section-overview" className="result-overview-section">
          <div className="result-header animate-fade-up">
            <div className="result-birth-info">
              {result.birth_year}年{result.birth_month}月{result.birth_day}日 {result.birth_hour}时
              &nbsp;·&nbsp;{result.gender === 'male' ? '男命' : '女命'}
            </div>
            <h1 className="result-pillars serif">
              {pillars.map(p => `${p.gan}${p.zhi}`).join('·')}
            </h1>
            <div className="result-tags">
              {stemLevelYongshenSummary ? (
                <StemLevelYongshenSummaryPanel summary={stemLevelYongshenSummary} />
              ) : (
                <>
                  <span className={`wuxing-badge ${fuyiYongshen ? 'wuxing-' + (WUXING_MAP[fuyiYongshen.charAt(0)] || 'jin') : 'wuxing-unknown'}`}>
                    扶抑喜用：{fuyiYongshen || (reportLoading ? '测算中...' : '待生成')}
                  </span>
                  <span className={`wuxing-badge ${fuyiJishen ? 'wuxing-' + (WUXING_MAP[fuyiJishen.charAt(0)] || 'huo') : 'wuxing-unknown'}`}>
                    扶抑忌：{fuyiJishen || (reportLoading ? '测算中...' : '待生成')}
                  </span>
                </>
              )}
              {result.ming_ge && (
                <span
                  className="mingge-badge"
                  onClick={() => setActiveMingGe({ name: result.ming_ge!, desc: result.ming_ge_desc || '' })}
                  title="点击查看格局说明"
                >
                  {result.ming_ge}
                </span>
              )}
            </div>
            <p className="result-quick-summary">{resultQuickSummary}</p>
          </div>

          <div className="result-reading-mode" aria-label="原局阅读模式">
            <button
              type="button"
              className={readingMode === 'simple' ? 'is-active' : ''}
              onClick={() => setReadingMode('simple')}
            >
              小白模式
            </button>
            <button
              type="button"
              className={readingMode === 'professional' ? 'is-active' : ''}
              onClick={() => setReadingMode('professional')}
            >
              专业模式
            </button>
          </div>

          <article className="result-verdict-card">
            <div className="result-product-card-head">
              <div>
                <span className="result-product-kicker">命盘总评</span>
                <h2 className="serif">先看结论</h2>
              </div>
              {fuyiYongshen && <span className="result-product-pill">扶抑喜用：{fuyiYongshen}</span>}
            </div>
            <p>{chartVerdict}</p>
            {chartKeywords.length > 0 && (
              <div className="result-keyword-row">
                {chartKeywords.map(keyword => (
                  <span key={keyword}>{keyword}</span>
                ))}
              </div>
            )}
            <button
              type="button"
              className="result-inline-link"
              onClick={() => readingMode === 'professional' ? scrollToResultSection('result-section-yongshen') : scrollToResultSection('result-section-evidence')}
            >
              查看判断依据
            </button>
          </article>

          <div className={`result-summary-grid${result.vehicle_profile ? '' : ' is-single-panel'}`}>
            {result.vehicle_profile && (
              <article className="result-vehicle-card result-summary-card result-summary-card--vehicle">
                <div className="result-product-card-head">
                  <div>
                    <span className="result-product-kicker">命盘座驾</span>
                    <h2 className="serif">
                      {vehicleGradeGuide?.label || result.vehicle_profile.grade_label} · {vehicleType}
                    </h2>
                  </div>
                  <span className="result-product-pill">{result.vehicle_profile.grade} 级</span>
                </div>
                <div className="result-vehicle-meter" aria-label={`命盘座驾分 ${result.vehicle_profile.score}`}>
                  <span style={{ width: `${Math.max(0, Math.min(100, result.vehicle_profile.score))}%` }} />
                </div>
                {vehicleGradeGuide && (
                  <p className="result-profile-plain-note">基础盘定位：{vehicleGradeGuide.summary}</p>
                )}
                {result.natal_assessment?.pattern.foundation_tier === 'high' && (
                  <section className="result-pattern-foundation" aria-label="格局基础说明">
                    <strong>{result.natal_assessment.pattern.foundation_source || '日干调候成格'} · {result.natal_assessment.pattern.foundation_label || '高格基础'}</strong>
                    <span>日干调候用神天透地藏，构成原局的高格基础；主格结构与制化配合仍另行判断。</span>
                  </section>
                )}
                {result.natal_assessment?.pattern && (
                  <p className="result-pattern-structure">
                    主格结构：{result.natal_assessment.pattern.name} · {result.natal_assessment.pattern.quality === 'formed' ? '成格' : result.natal_assessment.pattern.quality === 'usable' ? '可用' : result.natal_assessment.pattern.quality === 'partial' ? '部分成立' : result.natal_assessment.pattern.quality === 'broken' ? '受损' : '待定'}
                    {result.natal_assessment.pattern.formations.length > 0 && `；制化配合：${result.natal_assessment.pattern.formations.join('、')}`}
                    {result.natal_assessment.pattern.breaks.length > 0 && `；需留意：${result.natal_assessment.pattern.breaks.join('、')}`}
                  </p>
                )}
                {result.natal_assessment?.yongshen_alignment?.elements.length ? (
                  <p className="result-shared-yongshen">
                    <strong>共同优先用神：</strong>{result.natal_assessment.yongshen_alignment.elements.join('、')}
                    <span>同时属于寒热调候与扶抑；不替代完整扶抑喜用。</span>
                  </p>
                ) : null}
                <p>{result.vehicle_profile.summary}</p>
                {vehicleProfileTags.length > 0 && (
                  <div className="result-keyword-row">
                    {vehicleProfileTags.map(tag => <span key={tag}>{tag}</span>)}
                  </div>
                )}
                <div className="result-overview-actions">
                  <button
                    type="button"
                    className="result-overview-modal-trigger"
                    onClick={event => openOverviewModal('grade', event.currentTarget)}
                  >
                    等级说明
                  </button>
                  {readingMode === 'professional' && vehicleEvidences.length > 0 && (
                    <button
                      type="button"
                      className="result-overview-modal-trigger"
                      onClick={event => openOverviewModal('vehicle-evidence', event.currentTarget)}
                    >
                      查看座驾依据
                    </button>
                  )}
                </div>
              </article>
            )}

            {wealthProfile && (
              <article className="result-wealth-card result-summary-card result-summary-card--wealth">
                <div className="result-product-card-head">
                  <div>
                    <span className="result-product-kicker">财富结构</span>
                    <h2 className="serif">
                      {wealthProfile.grade_label} · {wealthProfile.wealth_type}
                    </h2>
                  </div>
                  <span className="result-product-pill">{wealthProfile.grade} 级</span>
                </div>
                <div className="result-vehicle-meter result-wealth-meter" aria-label={`财富结构分 ${wealthProfile.score}`}>
                  <span style={{ width: `${Math.max(0, Math.min(100, wealthProfile.score))}%` }} />
                </div>
                <p>{wealthProfile.summary}</p>
                {wealthCurrentHint && (
                  <section className={`result-wealth-current-hint is-${wealthCurrentHint.level}`} aria-label="当前财富窗口">
                    <span>{wealthCurrentHint.year} 年 · {wealthCurrentHint.gan_zhi}大运</span>
                    <strong>{wealthCurrentHint.label}</strong>
                    <p>{wealthCurrentHint.summary}</p>
                  </section>
                )}
                {wealthTags.length > 0 && (
                  <div className="result-keyword-row">
                    {wealthTags.map(tag => <span key={tag}>{tag}</span>)}
                  </div>
                )}
                {wealthRisks.length > 0 && (
                  <div className="result-wealth-risk-list" aria-label="财富结构风险提示">
                    {wealthRisks.map(flag => <span key={flag}>{flag}</span>)}
                  </div>
                )}
                {readingMode === 'professional' && wealthEvidences.length > 0 && (
                  <div className="result-overview-actions">
                    <button
                      type="button"
                      className="result-overview-modal-trigger"
                      onClick={event => openOverviewModal('wealth-evidence', event.currentTarget)}
                    >
                      查看财富依据
                    </button>
                  </div>
                )}
              </article>
            )}

            <article id="result-section-current" className="result-current-card result-summary-card result-summary-card--road">
              <div className="result-product-card-head">
                <div>
                  <span className="result-product-kicker">{currentDayunRoad ? '当前路况' : '当前阶段'}</span>
                  <h2 className="serif">
                    {currentDayunRoad
                      ? `${currentDayunRoad.gan_zhi}大运 · ${currentDayunRoad.road_label}`
                      : currentDayun ? `${currentDayun.gan}${currentDayun.zhi}大运` : '大运待确认'}
                  </h2>
                </div>
                <span className="result-product-pill">当前</span>
              </div>
              {currentDayun ? (
                <>
                  <p>{currentDayunRoad?.summary || `${currentDayun.start_age} - ${currentDayun.start_age + 9} 岁 · 公历 ${currentDayun.start_year} - ${currentDayun.end_year}`}</p>
                  {currentDayunRoad && currentRoadGuide && (
                    <p className="result-profile-plain-note"><strong>{currentRoadGuide.label}</strong>：{currentRoadGuide.summary}</p>
                  )}
                  {currentDayunRoad && (
                    <div className="result-road-phase-row">
                      <span>前五年：{currentDayunRoad.qian_road.label}</span>
                      <span>后五年：{currentDayunRoad.hou_road.label}</span>
                    </div>
                  )}
                  <div className="result-current-year">
                    <span>{currentYear} 年</span>
                    <strong>{currentLiuNian?.gan_zhi || '流年待排'}</strong>
                    <em>{currentLiuNian ? `${currentLiuNian.gan_shishen} / ${currentLiuNian.zhi_shishen}` : '进入大运区查看逐年走势'}</em>
                  </div>
                </>
              ) : (
                <p>当前命盘暂未返回大运数据，可先阅读命盘结构与 AI 解读。</p>
              )}
              {currentDayunRoad && (
                <div className="result-overview-actions">
                  <button
                    type="button"
                    className="result-overview-modal-trigger"
                    onClick={event => openOverviewModal('road', event.currentTarget)}
                  >
                    大运路况说明
                  </button>
                  {readingMode === 'professional' && hasCurrentRoadEvidence && (
                    <button
                      type="button"
                      className="result-overview-modal-trigger"
                      onClick={event => openOverviewModal('road-evidence', event.currentTarget)}
                    >
                      查看路况依据
                    </button>
                  )}
                </div>
              )}
              <button
                type="button"
                className="result-inline-link"
                onClick={() => targetId && !isGuest ? navigate(`/bazi/${targetId}/past-events`) : scrollToResultSection('result-section-ai')}
              >
                {targetId && !isGuest ? '查看过往年运回看' : '登录后查看年运回看'}
              </button>
            </article>
          </div>

          {targetId && !isGuest && (
            <div className="chart-archive-tools result-utility-bar">
              <div className="chart-archive-name">
                <label htmlFor="result-chart-display-name">命盘称呼</label>
                <input
                  id="result-chart-display-name"
                  value={chartDisplayNameDraft}
                  onChange={(event) => setChartDisplayNameDraft(event.target.value)}
                  maxLength={20}
                  placeholder={`${result.birth_year}年${result.birth_month}月${result.birth_day}日`}
                />
                <button type="button" className="btn btn-secondary" onClick={handleSaveChartDisplayName}>
                  保存称呼
                </button>
              </div>
              {chartDisplayNameError && <div className="chart-archive-error">{chartDisplayNameError}</div>}
              <div className="chart-archive-compatibility">
                <span>用此命盘发起合盘</span>
                <button type="button" className="btn btn-ghost" onClick={() => navigate(`/compatibility?importChart=${targetId}&role=self`)}>
                  作为我
                </button>
                <button type="button" className="btn btn-ghost" onClick={() => navigate(`/compatibility?importChart=${targetId}&role=partner`)}>
                  作为对方
                </button>
              </div>
            </div>
          )}

          {selectedDayun && selectedDayunTrendSeries.length > 0 && (
            <section id="result-section-trend" className="result-trend-panel" aria-labelledby="result-trend-title">
              <div className="result-section-heading">
                <div>
                  <span className="result-section-kicker">趋势图</span>
                  <h2 id="result-trend-title" className="section-title serif">大运十年趋势</h2>
                </div>
                <p>
                  {selectedDayunTitle} · {selectedDayunPeriod}
                </p>
              </div>

              <div className="result-trend-dayun-switcher" aria-label="切换大运趋势">
                {result.dayun.map(item => {
                  const isCurrent = currentDayun?.index === item.index
                  const isActive = selectedDayun.index === item.index
                  return (
                    <button
                      key={item.index}
                      type="button"
                      className={isActive ? 'is-active' : ''}
                      onClick={() => {
                        setTrendDayunIndex(item.index)
                        setActiveTrendYear(null)
                      }}
                    >
                      <strong>{item.gan}{item.zhi}</strong>
                      <span>{item.start_age}-{item.start_age + 9}岁</span>
                      {isCurrent && <em>当前</em>}
                    </button>
                  )
                })}
              </div>

              <div className="result-trend-legend" aria-label="趋势等级">
                <span><i className="result-trend-dot result-trend-dot--good" />顺势</span>
                <span><i className="result-trend-dot result-trend-dot--flat" />平稳</span>
                <span><i className="result-trend-dot result-trend-dot--watch" />留意</span>
              </div>

              <div className="result-trend-grid">
                {selectedDayunTrendSeries.map(series => (
                    <article key={series.key} className={`result-trend-card result-trend-card--${series.key}`}>
                      <div className="result-trend-card-head">
                        <h3><span />{series.title}</h3>
                        <em>{series.summary}</em>
                      </div>
                      <svg className="result-trend-chart" viewBox="0 0 420 150" role="img" aria-label={`${series.title}十年趋势点线图，点击年份节点查看依据`}>
                        <line className="result-trend-grid-line" x1="54" y1="28" x2="394" y2="28" />
                        <line className="result-trend-grid-line" x1="54" y1="70" x2="394" y2="70" />
                        <line className="result-trend-grid-line" x1="54" y1="112" x2="394" y2="112" />
                        <text className="result-trend-axis-label" x="16" y="32">顺势</text>
                        <text className="result-trend-axis-label" x="16" y="74">平稳</text>
                        <text className="result-trend-axis-label" x="16" y="116">留意</text>
                        <path className="result-trend-path" d={buildTrendPath(series.points)} />
                        {series.points.map((point, index) => {
                          const x = trendX(index, series.points.length)
                          const y = trendY(point.level)
                          const isCurrent = isSelectedCurrentDayun && point.year === currentYear
                          const isSelected = activeTrendYear === point.year
                          return (
                            <g
                              key={`${series.key}-${point.year}`}
                              role="button"
                              tabIndex={0}
                              className={[isCurrent ? 'is-current' : '', isSelected ? 'is-selected' : ''].filter(Boolean).join(' ')}
                              aria-label={`${point.year}年${series.title}${point.label}，点击查看依据`}
                              onClick={() => setActiveTrendYear(point.year)}
                              onKeyDown={event => {
                                if (event.key === 'Enter' || event.key === ' ') {
                                  event.preventDefault()
                                  setActiveTrendYear(point.year)
                                }
                              }}
                            >
                              {isCurrent && <circle className="result-trend-focus-ring" cx={x} cy={y} r="13" />}
                              {isSelected && <circle className="result-trend-selected-ring" cx={x} cy={y} r="15" />}
                              <circle
                                className={`result-trend-point result-trend-point--level-${point.level}`}
                                cx={x}
                                cy={y}
                                r="5.5"
                              />
                              <title>{point.year}年 · {point.ganZhi} · {point.label}：{point.detail}</title>
                            </g>
                          )
                        })}
                        {series.points.length > 0 && (
                          <>
                            <text className="result-trend-year-label" x={trendX(0, series.points.length)} y="140" textAnchor="middle">
                              {series.points[0].year}
                            </text>
                            <text className="result-trend-year-label" x={trendX(Math.floor((series.points.length - 1) / 2), series.points.length)} y="140" textAnchor="middle">
                              {series.points[Math.floor((series.points.length - 1) / 2)].year}
                            </text>
                            <text className="result-trend-year-label" x={trendX(series.points.length - 1, series.points.length)} y="140" textAnchor="middle">
                              {series.points[series.points.length - 1].year}
                            </text>
                          </>
                        )}
                      </svg>
                      <p>{buildTrendNote(series)}</p>
                    </article>
                ))}
              </div>

              <div className="result-trend-note" role="note">
                <span>趋势说明</span>
                <p>
                  曲线表示本段大运内部的相对趋势，不代表不同大运的绝对层级。同样上行，坏运可能只是从 50 到 100，好运可能是从 500 到 1000；请结合命局喜忌、大运整体强弱与 AI 解读判断。
                </p>
              </div>

              {activeTrendYearMeta && (
                <div
                  className="result-trend-modal-overlay"
                  onClick={() => setActiveTrendYear(null)}
                >
                  <div
                    className="result-trend-modal"
                    role="dialog"
                    aria-modal="true"
                    aria-labelledby="result-trend-modal-title"
                    onClick={event => event.stopPropagation()}
                  >
                    <div className="result-trend-year-detail-head">
                      <div>
                        <span>年度依据</span>
                        <h3 id="result-trend-modal-title">{activeTrendYearMeta.year}年 · {activeTrendYearMeta.ganZhi}</h3>
                        <p>{selectedDayunTitle} · {selectedDayunPeriod}</p>
                      </div>
                      <div className="result-trend-modal-actions">
                        {activeTrendYearMeta.age && <em>{activeTrendYearMeta.age}岁</em>}
                        <button
                          type="button"
                          className="result-trend-modal-close"
                          onClick={() => setActiveTrendYear(null)}
                          aria-label="关闭年度依据"
                        >
                          <X size={18} />
                        </button>
                      </div>
                    </div>
                    <p className="result-trend-year-summary">{activeTrendYearSummary}</p>
                    <div className="result-trend-year-detail-grid">
                      {activeTrendYearItems.map(({ series, point }) => (
                        <article key={`${series.key}-${point.year}`} className={`result-trend-year-item result-trend-year-item--${series.key}`}>
                          <div className="result-trend-year-item-head">
                            <strong>{series.title}</strong>
                            <span className={`result-trend-level result-trend-level--${point.level}`}>{point.label}</span>
                          </div>
                          <dl>
                            <div>
                              <dt>十神依据</dt>
                              <dd>{point.ganShishen || '待排'} / {point.zhiShishen || '待排'}</dd>
                            </div>
                            <div>
                              <dt>判断说明</dt>
                              <dd>{point.detail}</dd>
                            </div>
                            <div>
                              <dt>行动建议</dt>
                              <dd>{point.action}</dd>
                            </div>
                          </dl>
                        </article>
                      ))}
                    </div>
                  </div>
                </div>
              )}
            </section>
          )}

          {historicalFiguresSection}

          {readingMode === 'simple' && (
            <section id="result-section-evidence" className="result-evidence-panel" aria-labelledby="simple-evidence-title">
              <div className="result-section-heading">
                <span className="result-section-kicker">判断依据</span>
                <h2 id="simple-evidence-title" className="section-title serif">为什么这样看</h2>
                <p>这里不重新计算算法，只把现有命盘数据整理成更容易理解的依据链。</p>
              </div>
              <div className="result-evidence-grid">
                {yongshenEvidence.map((item, index) => (
                  <article key={item.title} className="result-evidence-item">
                    <span>{index + 1}</span>
                    <div>
                      <h3>{item.title}</h3>
                      <p>{item.detail}</p>
                    </div>
                  </article>
                ))}
              </div>
              <button type="button" className="result-professional-cta" onClick={() => setReadingMode('professional')}>
                展开专业命盘
              </button>
            </section>
          )}

          <div className="result-segment-nav">
            <SegmentedTabs items={resultSegments} ariaLabel="结果页分段导航" />
          </div>
        </section>

        {/* 命盘详情 */}
        {readingMode === 'professional' && (
        <section id="result-section-chart" className="professional-view animate-fade-up">

            {/* 四柱数据网格 (Professional Data Grid) */}
            <div className="pillars-section card bazi-primary-panel">
              <div className="result-panel-heading">
                <div>
                  <span className="result-panel-kicker">排盘核心</span>
                  <h2 className="section-title serif">基本排盘</h2>
                </div>
                <p>先确认四柱主盘，再向下查看十神、五行与调候结构。</p>
              </div>
              <div className="bazi-data-grid bazi-data-grid--primary">
                
                {/* 标尺列1：列头 */}
                <div className="grid-cell row-label">日期</div>
                {pillars.map((p, i) => <div key={i} className={`grid-cell col-header${i === 2 ? ' is-day-pillar' : ''}${dayPillarCellClass(i)}`}>{p.label}</div>)}

                {/* 主星行 */}
                <div className="grid-cell row-label">主星</div>
                {pillars.map((p, i) => <div key={i} className={`grid-cell top-shishen text-muted${dayPillarCellClass(i)}`}>{p.ganShiShen}</div>)}

                {/* 天干行 */}
                <div className="grid-cell row-label">天干</div>
                {pillars.map((p, i) => (
                  <div key={i} className={`grid-cell main-char gan wuxing-text-${WUXING_MAP[p.ganWx] || 'jin'}${dayPillarCellClass(i)}`}>
                    <span>{p.gan}</span><span className="wx-tag">{p.ganWx}</span>
                  </div>
                ))}

                {/* 地支行 */}
                <div className="grid-cell row-label">地支</div>
                {pillars.map((p, i) => (
                  <div key={i} className={`grid-cell main-char zhi wuxing-text-${WUXING_MAP[p.zhiWx] || 'jin'}${dayPillarCellClass(i)}`}>
                    <span>{p.zhi}</span><span className="wx-tag">{p.zhiWx}</span>
                  </div>
                ))}

                {/* 藏干行 */}
                <div className="grid-cell row-label">藏干</div>
                {pillars.map((p, i) => (
                  <div key={i} className={`grid-cell hide-gan-cell${dayPillarCellClass(i)}`}>
                    {p.hideGan.map((g, idx) => (
                       <div key={idx} className={`hg-row wuxing-text-${WUXING_MAP[GAN_WUXING[g]] || 'shui'}`} style={{ color: 'var(--text-color)' }}>{g}</div>
                    ))}
                  </div>
                ))}

                {/* 副星行 */}
                <div className="grid-cell row-label">副星</div>
                {pillars.map((p, i) => (
                  <div key={i} className={`grid-cell hide-gan-cell text-muted${dayPillarCellClass(i)}`}>
                    {p.zhiShiShen.map((ss, idx) => <div key={idx} className="hg-row">{ss}</div>)}
                  </div>
                ))}

                {/* 星运行 */}
                <div className="grid-cell row-label">星运</div>
                {pillars.map((p, i) => <div key={i} className={`grid-cell text-muted${dayPillarCellClass(i)}`}>{p.xingYun || p.diShi}</div>)}

                {/* 自坐行 */}
                <div className="grid-cell row-label">自坐</div>
                {pillars.map((p, i) => <div key={i} className={`grid-cell text-muted${dayPillarCellClass(i)}`}>{p.diShi}</div>)}

                {/* 空亡行 */}
                <div className="grid-cell row-label">空亡</div>
                {pillars.map((p, i) => <div key={i} className={`grid-cell text-muted${dayPillarCellClass(i)}`}>{p.xunKong}</div>)}

                {/* 纳音行 */}
                <div className="grid-cell row-label">纳音</div>
                {pillars.map((p, i) => <div key={i} className={`grid-cell text-muted nayin${dayPillarCellClass(i)}`}>{p.naYin}</div>)}

                {/* 神煞行 */}
                <div className="grid-cell row-label shensha-label">神煞</div>
                {pillars.map((p, i) => (
                  <div key={i} className={`grid-cell shensha-cell${dayPillarCellClass(i)}`}>
                    {p.shenSha.map((sh, idx) => {
                      const polarity = SHENSHA_POLARITY[sh] || 'zhong'
                      const hasAnnotation = shenshaMap.has(sh)
                      return (
                        <span
                          key={idx}
                          className={`shensha-tag shensha-tag--${polarity}${hasAnnotation ? ' shensha-tag--clickable' : ''}`}
                          onClick={() => {
                            const ann = shenshaMap.get(sh)
                            if (ann) setActiveAnnotation(ann)
                          }}
                          onMouseEnter={() => {
                            if (!hasAnnotation) return
                            hoverTimer.current = setTimeout(() => {
                              const ann = shenshaMap.get(sh)
                              if (ann) setActiveAnnotation(ann)
                            }, 300)
                          }}
                          onMouseLeave={() => {
                            if (hoverTimer.current) clearTimeout(hoverTimer.current)
                          }}
                        >{sh}</span>
                      )
                    })}
                  </div>
                ))}

              </div>
            </div>

            <GongJiaPanel
              items={result.gong_jia || []}
              hasShenshaAnnotation={(name) => shenshaMap.has(name)}
              shenshaPolarity={(name) => SHENSHA_POLARITY[name] || 'zhong'}
              onShenshaClick={(name) => {
                const ann = shenshaMap.get(name)
                if (!ann) return
                setActiveAnnotation({
                  ...ann,
                  description: `拱神煞：${ann.description}`,
                })
              }}
            />

            <section className="ten-god-relation-section" aria-labelledby="ten-god-relation-title">
              <div className="ten-god-relation-header">
                <div>
                  <h2 id="ten-god-relation-title" className="section-title serif">命主十神关系</h2>
                  <p>命主日元：<strong>{relation.day_master.label}</strong></p>
                </div>
                <span>以日干为参照点，查看其他天干与藏干对应的十神。</span>
              </div>

              <div className="ten-god-stem-grid">
                {relation.heavenly_stems.map((item) => (
                  <article key={item.pillar} className={`ten-god-relation-card ${item.pillar === 'day' ? 'is-day-master' : ''}`}>
                    <div className="ten-god-card-topline">
                      <span>{item.pillar_label}</span>
                      {item.group_label && <em>{item.group_label}</em>}
                    </div>
                    <div className="ten-god-card-main">
                      <strong className={`wuxing-text-${WUXING_MAP[item.wuxing] || 'jin'}`}>{item.gan}</strong>
                      <span>{item.ten_god}</span>
                    </div>
                    <p>{item.relation} · {item.summary}</p>
                  </article>
                ))}
              </div>

              {hiddenStemGroups.length > 0 && (
                <div className="ten-god-hidden-block">
                  <h3>地支藏干关系</h3>
                  <div className="ten-god-hidden-grid">
                    {hiddenStemGroups.map((group) => (
                      <article key={group.pillar} className="ten-god-hidden-card">
                        <div className="ten-god-hidden-title">
                          <span>{group.pillar_label}</span>
                          <strong>{group.branch}</strong>
                        </div>
                        <div className="ten-god-hidden-list">
                          {group.items.map((item) => (
                            <div key={`${group.pillar}-${item.gan}`} className="ten-god-hidden-item">
                              <span className={`wuxing-text-${WUXING_MAP[item.wuxing] || 'jin'}`}>{item.gan}</span>
                              <strong>{item.ten_god}</strong>
                              <em>{item.relation}</em>
                              <small>{item.summary}</small>
                            </div>
                          ))}
                        </div>
                      </article>
                    ))}
                  </div>
                </div>
              )}
            </section>

            <section id="result-section-yongshen" className="result-structure-section" aria-labelledby="structure-title">
              <div className="result-section-heading">
                <span className="result-section-kicker">结构判断</span>
                <h2 id="structure-title" className="section-title serif">命局结构</h2>
                <p>扶抑、日干调候与寒热调候分别呈现，避免把不同判断混为一个结论。</p>
              </div>

              <div className="result-structure-grid">
                <div className="structure-card structure-card--wuxing card">
                  <h3 className="structure-card-title serif">五行分布</h3>
                  <WuxingRadar wuxing={result.wuxing} />
                </div>

                <div className="structure-card structure-card--yongshen">
                  <YongshenBadge yongshen={fuyiYongshen} jishen={fuyiJishen} />
                </div>

                {stemGuidance && (
                  <div className="structure-card structure-card--stem-guidance">
                    <StemGuidancePanel guidance={stemGuidance} />
                  </div>
                )}

                {result.tiaohou && (
                  <div className="structure-card structure-card--tiaohou card">
                    <TiaohouCard
                      dayGan={result.day_gan}
                      monthZhi={result.month_zhi}
                      score={result.natal_assessment?.tiaohou?.day_stem.score}
                      formation={result.natal_assessment?.tiaohou?.day_stem.formation}
                      foundationTier={result.natal_assessment?.tiaohou?.day_stem.foundation_tier}
                      foundationScore={result.natal_assessment?.tiaohou?.day_stem.foundation_score}
                      tiaohou={result.tiaohou}
                    />
                  </div>
                )}

                {result.natal_assessment?.tiaohou?.thermal && (
                  <div className="structure-card structure-card--tiaohou card">
                    <ThermalTiaohouCard thermal={result.natal_assessment.tiaohou.thermal} />
                  </div>
                )}
              </div>
            </section>


            {/* 命理专属头像 (Feature Flag 控制) */}
            {ENABLE_MINGPAN_AVATAR && (
              <div className="mingpan-avatar-section card">
                <h2 className="section-title serif">专属命理头像</h2>
                <p className="section-desc">根据你的喜用神五行，程序化生成专属命元图腾</p>
                <MingpanAvatar
                  yongshen={result.yongshen || ''}
                  jishen={result.jishen || ''}
                  dayGan={result.day_gan || ''}
                />
              </div>
            )}

            {/* 大运时间轴 */}
            <section id="result-section-dayun" className="dayun-section">
              <DayunTimeline
                dayun={result.dayun}
                birthYear={result.birth_year}
                startYunSolar={result.start_yun_solar}
                dayGan={result.day_gan || ''}
                gender={result.gender}
                pillarsLabel={dayunPillarsLabel}
                chartId={targetId}
                yongshen={fuyiYongshen}
                jishen={fuyiJishen}
                wuxing={result.wuxing}
                tiaohou={result.tiaohou ?? null}
                dayunRoadmap={result.dayun_roadmap}
              />
              {(isGuest || targetId) && (
                <button
                  type="button"
                  className={`past-events-entry${isGuest ? ' is-disabled' : ''}`}
                  onClick={isGuest || !targetId ? undefined : () => navigate(`/bazi/${targetId}/past-events`)}
                  disabled={isGuest || !targetId}
                  aria-label="过往年运回看"
                >
                  <History className="past-events-entry-icon" size={22} aria-hidden="true" />
                  <span className="past-events-entry-body">
                    <span className="past-events-entry-title">过往年运回看</span>
                    <span className="past-events-entry-sub">
                      {isGuest ? '登录后可查看年运回看' : '查看大运分段、年份信号与 AI 批语'}
                    </span>
                  </span>
                  {!isGuest && (
                    <span className="past-events-entry-cta" aria-hidden="true">进入回看 →</span>
                  )}
                </button>
              )}
            </section>
          </section>
        )}

        {/* AI 解读区域 */}
        <section id="result-section-ai" className="report-section card animate-fade-up">
          <div className="report-section-header">
            <h2 className="section-title serif">命理解读</h2>
            <div className="report-header-actions">
              {report && (
                <>
                  <button
                    id="save-card-btn"
                    className="btn btn-ghost btn-sm"
                    onClick={handleSaveImage}
                    disabled={savingImage || !hasRenderableOriginalReport}
                  >
                    {savingImage ? '生成中...' : '保存分享图'}
                  </button>
                  <button
                    id="export-report-btn"
                    className="btn btn-ghost btn-sm"
                    onClick={handleExportPDF}
                    disabled={exportingPDF || !canExportActiveReport}
                  >
                    {exportingPDF ? '生成中...' : (reportTab === 'polished' && polishedReport ? '导出润色版 PDF' : '导出 PDF')}
                  </button>
                </>
              )}
            </div>
          </div>

          {report && (
            <div className="report-tab-row" ref={reportTabRowRef}>
              <button
                className={`report-tab${reportTab === 'original' ? ' is-active' : ''}`}
                onClick={() => switchReportTab('original')}
              >
                原版
              </button>
              <button
                className={`report-tab${reportTab === 'polished' ? ' is-active' : ''}`}
                onClick={() => switchReportTab('polished')}
              >
                润色版
              </button>
            </div>
          )}

          {/* 已有报告 */}
          {reportTab === 'original' && report && (
            <div className="report-sections">
              {/* 精简/专业切换按钮（仅新格式报告显示） */}
              {structured && (
                <>
                  <div className="report-digest-card">
                    <div className="report-digest-heading">
                      <span>阅读摘要</span>
                      <strong className="serif">{structured.yongshen || result.yongshen || '命局'}为线索</strong>
                    </div>
                    <div className="report-digest-grid">
                      {reportDigestItems.map(item => (
                        <div key={item.label} className="report-digest-item">
                          <span>{item.label}</span>
                          <p>{item.value}</p>
                        </div>
                      ))}
                    </div>
                  </div>

                  <div className="report-mode-switcher">
                    <button
                      className={`mode-btn${reportMode === 'brief' ? ' active' : ''}`}
                      onClick={() => setReportMode('brief')}
                    >精简版</button>
                    <button
                      className={`mode-btn${reportMode === 'detail' ? ' active' : ''}`}
                      onClick={() => setReportMode('detail')}
                    >完整解读</button>
                  </div>

                  <div className="report-term-glossary" aria-label="命理术语解释">
                    {REPORT_TERMS.map(item => (
                      <div key={item.term} className="report-term-item">
                        <strong>{item.term}</strong>
                        <span>{item.desc}</span>
                      </div>
                    ))}
                  </div>
                </>
              )}

              {/* 新格式：结构化渲染 */}
              {structured ? (
                <>
                  {reportMode === 'detail' && structured.analysis?.logic && (
                    <div className="report-block report-analysis">
                      <h3 className="report-block-title serif"><Diamond size={16} className="title-diamond-icon" /> 命局分析总览</h3>
                      <div className="report-block-content">
                        {cleanReportText(structured.analysis.logic)
                          .split(/\n{2,}/)
                          .filter(Boolean)
                          .map((para, idx) => <p key={idx}>{para}</p>)}
                      </div>
                    </div>
                  )}
                  {reportMode === 'brief' && structured.analysis?.summary && (
                    <div className="report-summary">
                      <span>{cleanReportText(structured.analysis.summary)}</span>
                    </div>
                  )}
                  <div className="report-chapter-list">
                    {(structured.chapters || []).map((ch, i) => {
                      const raw = reportMode === 'brief' ? ch.brief : ch.detail
                      const paragraphs = cleanReportText(raw)
                        .split(/\n{2,}/)
                        .map(p => p.trim())
                        .filter(Boolean)
                      return (
                        <details key={i} className="report-chapter-detail" open>
                          <summary>
                            <span className="serif">【{ch.title}】</span>
                            <em>{cleanReportText(ch.brief)}</em>
                          </summary>
                          <div className="report-block-content">
                            {paragraphs.length > 0
                              ? paragraphs.map((para, idx) => <p key={idx}>{para}</p>)
                              : <p>{cleanReportText(raw)}</p>}
                          </div>
                        </details>
                      )
                    })}
                  </div>
                </>
              ) : (
                /* 旧格式：降级渲染 */
                reportSections.length > 0 ? reportSections.map((sec, i) => (
                  <div key={i} className="report-block">
                    <h3 className="report-block-title serif">{sec.title}</h3>
                    <div className="report-block-content">
                      {splitParagraphs(sec.content).map((para, idx) => <p key={idx}>{para}</p>)}
                    </div>
                  </div>
                )) : hasLegacyReportContent ? (
                  <div className="report-content">
                    {splitParagraphs(legacyReportContent).map((para, idx) => <p key={idx}>{para}</p>)}
                  </div>
                ) : (
                  <div className="report-empty-state">
                    <strong>当前历史记录没有可展示的命理解读</strong>
                    <p>这条报告只保存了分隔线或免责声明，建议重新生成一版完整解读。</p>
                    <button
                      type="button"
                      className="btn btn-primary btn-sm"
                      onClick={handleGenerateReport}
                      disabled={reportLoading || isStreaming || isThinking}
                    >
                      重新生成命理解读
                    </button>
                  </div>
                )
              )}
              {(structured || reportSections.length > 0 || hasLegacyReportContent) && (
                <p className="report-disclaimer">
                  本报告内容仅供参考，不构成任何决策建议。
                </p>
              )}
            </div>
          )}

          {reportTab === 'original' && (
            <>
              {/* 流式生成中 */}
              {isStreaming && (
                <div className="report-sections animate-fade-in">
                  <div className="report-content" style={{ whiteSpace: 'pre-wrap', fontFamily: 'monospace', lineHeight: 1.8 }}>
                    {streamingText}
                    <span className="cursor-blink">|</span>
                  </div>
                </div>
              )}

              {/* 推理模型正在思考 */}
              {isThinking && !isStreaming && (
                <div className="ai-loading-container animate-fade-in">
                  <div className="ai-loading-icon">
                    <div className="spinner"></div>
                  </div>
                  <div className="ai-loading-step">
                    <div className="ai-loading-text">
                      正在深度推演中...  已思考 {thinkingSeconds} 秒
                    </div>
                  </div>
                </div>
              )}

              {/* 初始加载等待动画（SSE连接建立前） */}
              {reportLoading && !isStreaming && !isThinking && (
                <div className="ai-loading-container animate-fade-in">
                  <div className="ai-loading-icon">
                    <div className="spinner"></div>
                  </div>
                  <div className="ai-loading-step">
                    <div key={loadingStepIndex} className="ai-loading-text">
                      {LOADING_STEPS[loadingStepIndex]}
                    </div>
                  </div>
                </div>
              )}

              {/* 报错 */}
              {reportError && !reportLoading && !isStreaming && (
                <div className="report-retry-panel">
                  <p className="form-error">{reportError}</p>
                  <button
                    type="button"
                    className="btn btn-secondary btn-sm"
                    onClick={handleGenerateReport}
                    disabled={reportLoading || isStreaming || isThinking}
                  >
                    重试生成
                  </button>
                </div>
              )}

              {reportError && streamingText && !report && !reportLoading && !isStreaming && (
                <div className="report-sections animate-fade-in">
                  <div className="report-content" style={{ whiteSpace: 'pre-wrap', fontFamily: 'monospace', lineHeight: 1.8 }}>
                    {streamingText}
                  </div>
                </div>
              )}

              {/* 未生成：显示按钮或引导 */}
              {!report && !reportLoading && !isStreaming && !isThinking && (
                <>
                  {!isGuest ? (
                    <div className="report-cta">
                      <p className="report-cta-desc">
                        点击下方按钮，生成性格、感情、事业、健康四维解读
                      </p>
                      <button
                        id="generate-ai-report"
                        className="btn btn-primary"
                        onClick={handleGenerateReport}
                      >
                        生成命理解读
                      </button>
                    </div>
                  ) : (
                    <div className="guest-banner">
                      <span>登录后可获得完整解读报告，并保存命盘记录</span>
                      {registration_enabled
                        ? (
                          <a
                            href={buildAuthPath('/register', '/result')}
                            className="btn btn-primary btn-sm"
                            onClick={() => persistGuestJourney('generate_report')}
                          >
                            立即注册
                          </a>
                        )
                        : (
                          <a
                            href={buildAuthPath('/login', '/result')}
                            className="btn btn-primary btn-sm"
                            onClick={() => persistGuestJourney('generate_report')}
                          >
                            登录账号
                          </a>
                        )}
                    </div>
                  )}
                </>
              )}
            </>
          )}

          {reportTab === 'polished' && (
            <PolishedPanel
              polishedReport={polishedReport}
              hasOriginalReport={!!report}
              loading={polishing}
              errorMsg={polishError}
              onSubmit={submitPolish}
            />
          )}

          {report && (
            <div className="report-action-bar">
              <button className="btn btn-ghost" onClick={() => navigate('/')}>重新起盘</button>
              {user && <a href="/history" className="btn btn-ghost">查看历史</a>}
              <button
                className="btn btn-ghost"
                onClick={handleExportPDF}
                disabled={exportingPDF || !canExportActiveReport}
              >
                {exportingPDF ? '生成中...' : '导出 PDF'}
              </button>
            </div>
          )}
        </section>
      </div>

      {overviewModal && (
        <div
          className="result-overview-modal"
          onMouseDown={event => {
            if (event.target === event.currentTarget) closeOverviewModal()
          }}
        >
          <div
            ref={overviewModalDialogRef}
            className="result-overview-modal__dialog"
            role="dialog"
            aria-modal="true"
            aria-labelledby="result-overview-modal-title"
            tabIndex={-1}
          >
            <div className="result-overview-modal__header">
              <div>
                <span>{overviewModalKicker}</span>
                <h3 id="result-overview-modal-title" className="serif">{overviewModalTitle}</h3>
              </div>
              <button
                type="button"
                className="result-overview-modal__close"
                onClick={closeOverviewModal}
                aria-label={`关闭${overviewModalTitle}`}
                title="关闭"
              >
                <X size={18} />
              </button>
            </div>

            <div className="result-overview-modal__body">
              {overviewModal === 'grade' && (
                <div className="result-grade-guide-body">
                  <p>S 到 D 区分原局基础层次：调候急需优先；无急需时以扶抑为基线，再看日干调候成格、主格结构、制化与流通；不等同于人的高低或人生结局。</p>
                  <ul className="result-grade-guide">
                    {VEHICLE_GRADE_GUIDE.map(item => (
                      <li key={item.grade} className={item.grade === result.vehicle_profile?.grade ? 'is-current' : ''}>
                        <strong>{item.grade} · {item.label} · {item.vehicle}</strong>
                        <span>{item.detail}</span>
                      </li>
                    ))}
                  </ul>
                </div>
              )}

              {overviewModal === 'road' && (
                <div className="result-road-guide-body">
                  <p>车代表命盘的基础配置与驾驭难度，路代表每十年大运带来的外部支持与阻力。车好不等于一路顺，路顺也能让普通配置发挥得更好。</p>
                  <ul className="result-road-guide">
                    {ROAD_GUIDE.map(item => (
                      <li key={item.type} className={item.type === currentDayunRoad?.road_type ? 'is-current' : ''}>
                        <strong>{item.label}</strong>
                        <span>{item.summary}</span>
                      </li>
                    ))}
                  </ul>
                </div>
              )}

              {overviewModal === 'vehicle-evidence' && vehicleEvidences.length > 0 && (
                <div className="result-profile-evidence-body">
                  {vehicleDrivingStyle && (
                    <p className="result-profile-driving-style">
                      <strong>驾驶特性</strong>
                      <span>{vehicleDrivingStyle}</span>
                      <em>由命格辅助说明，不参与基础盘等级或主车型判定。</em>
                    </p>
                  )}
                  <ul className="result-profile-evidence">
                    {vehicleEvidences.map((item, index) => (
                      <li key={`${item.source}-${index}`}>
                        <strong>{item.source}</strong>
                        <span>{item.label} · {item.impact} {formatEvidenceDelta(item.delta)}</span>
                        <em>{item.detail}</em>
                      </li>
                    ))}
                  </ul>
                </div>
              )}

              {overviewModal === 'wealth-evidence' && wealthEvidences.length > 0 && (
                <div className="result-profile-evidence-body">
                  <p className="result-profile-plain-note">
                    财富结构只说明原局钱财资源的显露、承载、流通与守成风险，不等于现实资产、收入规模或投资建议。
                  </p>
                  {wealthRisks.length > 0 && (
                    <div className="result-wealth-risk-list result-wealth-risk-list--modal" aria-label="财富结构风险">
                      {wealthRisks.map(flag => <span key={flag}>{flag}</span>)}
                    </div>
                  )}
                  <ul className="result-profile-evidence">
                    {wealthEvidences.map((item, index) => (
                      <li key={`${item.source}-${index}`}>
                        <strong>{item.source}</strong>
                        <span>{item.label} · {item.impact} {formatEvidenceDelta(item.delta)}</span>
                        <em>{item.detail}</em>
                      </li>
                    ))}
                  </ul>
                  {wealthCurrentHint && (
                    <section className="result-wealth-current-hint result-wealth-current-hint--modal" aria-label="当前财富窗口说明">
                      <span>{wealthCurrentHint.year} 年 · {wealthCurrentHint.gan_zhi}大运</span>
                      <strong>{wealthCurrentHint.label}</strong>
                      <p>{wealthCurrentHint.summary}</p>
                      {wealthCurrentHint.evidences.length > 0 && <em>依据：{wealthCurrentHint.evidences.join('、')}</em>}
                    </section>
                  )}
                </div>
              )}

              {overviewModal === 'road-evidence' && hasCurrentRoadEvidence && (
                <div className="result-profile-evidence-body">
                  {currentRoadPhaseEvidences.length > 0 ? (
                    <>
                      {currentRoadEvidences.length > 0 && (
                        <section className="result-road-evidence-aggregate" aria-labelledby="road-evidence-aggregate-title">
                          <h4 id="road-evidence-aggregate-title">十年合计</h4>
                          <ul className="result-profile-evidence">
                            {currentRoadEvidences.map((item, index) => (
                              <li key={`${item.source}-${index}`}>
                                <strong>{item.source}</strong>
                                <span>{item.label} · {item.impact} {formatEvidenceDelta(item.delta)}</span>
                                <em>{item.detail}</em>
                              </li>
                            ))}
                          </ul>
                        </section>
                      )}
                      <div className="result-road-evidence-phases">
                        {currentRoadPhaseEvidences.map((phase) => (
                          <section key={phase.phase} className="result-road-evidence-phase" aria-labelledby={`road-evidence-${phase.phase}`}>
                            <div className="result-road-evidence-phase-heading">
                              <h4 id={`road-evidence-${phase.phase}`}>{phase.label}</h4>
                              <span>阶段合计 {formatEvidenceDelta(phase.delta)}</span>
                            </div>
                            <ul className="result-profile-evidence">
                              {phase.evidences.map((item, index) => (
                                <li key={`${phase.phase}-${item.source}-${index}`}>
                                  <strong>{item.source}</strong>
                                  <span>{item.label} · {item.impact} {formatEvidenceDelta(item.delta)}</span>
                                  <em>{item.detail}</em>
                                </li>
                              ))}
                            </ul>
                          </section>
                        ))}
                      </div>
                    </>
                  ) : (
                    <ul className="result-profile-evidence">
                      {currentRoadEvidences.map((item, index) => (
                        <li key={`${item.source}-${index}`}>
                          <strong>{item.source}</strong>
                          <span>{item.label} · {item.impact} {formatEvidenceDelta(item.delta)}</span>
                          <em>{item.detail}</em>
                        </li>
                      ))}
                    </ul>
                  )}
                </div>
              )}
            </div>
          </div>
        </div>
      )}

      {/* 隐藏的分享卡片（用于生成图片，不可见） */}
      <div style={{ position: 'fixed', top: -9999, left: -9999, zIndex: -1, pointerEvents: 'none' }}>
        <ShareCard
          ref={shareCardRef}
          birthYear={result.birth_year}
          birthMonth={result.birth_month}
          birthDay={result.birth_day}
          birthHour={result.birth_hour}
          gender={result.gender}
          yearGan={result.year_gan} yearZhi={result.year_zhi}
          monthGan={result.month_gan} monthZhi={result.month_zhi}
          dayGan={result.day_gan} dayZhi={result.day_zhi}
          hourGan={result.hour_gan} hourZhi={result.hour_zhi}
          yearGanWx={result.year_gan_wuxing} yearZhiWx={result.year_zhi_wuxing}
          monthGanWx={result.month_gan_wuxing} monthZhiWx={result.month_zhi_wuxing}
          dayGanWx={result.day_gan_wuxing} dayZhiWx={result.day_zhi_wuxing}
          hourGanWx={result.hour_gan_wuxing} hourZhiWx={result.hour_zhi_wuxing}
          structured={report?.content_structured ?? null}
          brand={brand}
          pastEventsExportSegments={pastEventsExportSegments}
          dayunTrendSeries={selectedDayunTrendSeries}
          dayunTrendLabel={selectedDayunTitle}
          dayunTrendPeriod={selectedDayunPeriod}
        />
      </div>

      {/* 神煞注解浮层卡片 */}
      {activeAnnotation && (
        <div
          className="shensha-modal-overlay"
          onClick={() => setActiveAnnotation(null)}
        >
          <div
            className="shensha-modal-card"
            onClick={e => e.stopPropagation()}
          >
            <div className="shensha-modal-header">
              <div className="shensha-modal-title">
                <span className={`shensha-modal-dot shensha-modal-dot--${activeAnnotation.polarity}`} />
                <span className="shensha-modal-name">{activeAnnotation.name}</span>
                <span className={`shensha-modal-badge shensha-modal-badge--${activeAnnotation.polarity}`}>
                  {activeAnnotation.polarity === 'ji' ? '吉神' : activeAnnotation.polarity === 'xiong' ? '凶煞' : '中性'}
                </span>
              </div>
              <button
                className="shensha-modal-close"
                onClick={() => setActiveAnnotation(null)}
                aria-label="关闭"
              >
                <X size={18} />
              </button>
            </div>
            <div className="shensha-modal-divider" />
            <div className="shensha-modal-body">
              {activeAnnotation.category && (
                <span style={{ fontSize: 11, color: 'var(--wu-shui)', background: 'rgba(91,155,213,0.12)', borderRadius: 4, padding: '2px 8px', marginBottom: 8, display: 'inline-block' }}>
                  {activeAnnotation.category}
                </span>
              )}
              {activeAnnotation.short_desc && (
                <p style={{ color: 'var(--text-secondary)', fontSize: 13, margin: '6px 0 10px', fontStyle: 'italic' }}>{activeAnnotation.short_desc}</p>
              )}
              <p className="shensha-modal-description">{activeAnnotation.description}</p>
            </div>
          </div>
        </div>
      )}

      {/* 命格说明 Modal */}
      {activeMingGe && (
        <div
          className="shensha-modal-overlay"
          onClick={() => setActiveMingGe(null)}
        >
          <div
            className="shensha-modal-card"
            onClick={e => e.stopPropagation()}
          >
            <div className="shensha-modal-header">
              <div className="shensha-modal-title">
                <span className="mingge-modal-dot" />
                <span className="shensha-modal-name">{activeMingGe.name}</span>
                <span className="mingge-modal-badge">格局</span>
              </div>
              <button
                className="shensha-modal-close"
                onClick={() => setActiveMingGe(null)}
                aria-label="关闭"
              >
                <X size={18} />
              </button>
            </div>
            <div className="shensha-modal-divider" />
            <div className="shensha-modal-body">
              <p className="shensha-modal-description">{activeMingGe.desc}</p>
            </div>
          </div>
        </div>
      )}
      </div>
      {(report || polishedReport) && (() => {
        const isPolishedExport = reportTab === 'polished' && polishedReport
        const printStructured = isPolishedExport && polishedReport.content_structured
          ? polishedReport.content_structured
          : structured
        return (
          <PrintLayout
            birthYear={result.birth_year}
            birthMonth={result.birth_month}
            birthDay={result.birth_day}
            birthHour={result.birth_hour}
            gender={result.gender}
            mingGe={result.ming_ge || ''}
            mingGeDesc={result.ming_ge_desc || ''}
            pillars={pillars}
            dayun={result.dayun}
            structured={printStructured}
            shenshaMap={shenshaMap}
            tenGodRelation={relation}
            polishedUserSituation={isPolishedExport ? polishedReport.user_situation : undefined}
            brand={brand}
            pastEventsExportSegments={pastEventsExportSegments}
            dayunTrendSeries={selectedDayunTrendSeries}
            dayunTrendLabel={selectedDayunTitle}
            dayunTrendPeriod={selectedDayunPeriod}
          />
        )
      })()}
    </>
  )
}

function parseReport(content: string) {
  const sections: { title: string; content: string }[] = []
  const normalized = cleanLegacyReportContent(content)
  const matches = normalized.matchAll(/【(.+?)】\n?([\s\S]*?)(?=【|$)/g)
  for (const m of matches) {
    const sectionContent = cleanLegacyReportContent(m[2])
    if (hasMeaningfulReportContent(sectionContent)) {
      sections.push({ title: `【${m[1]}】`, content: sectionContent })
    }
  }
  return sections
}

function LoadingSkeleton() {
  return (
    <div className="result-page page">
      <div className="container">
        <div style={{ paddingTop: 40 }}>
          <div className="skeleton" style={{ height: 32, width: 300, marginBottom: 16 }} />
          <div className="skeleton" style={{ height: 48, width: 400, marginBottom: 32 }} />
          <div className="skeleton" style={{ height: 300, borderRadius: 20 }} />
        </div>
      </div>
    </div>
  )
}
