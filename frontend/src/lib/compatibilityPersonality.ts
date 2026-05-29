import type {
  CompatibilityChartSnapshot,
  CompatibilityDecisionAdvice,
  CompatibilityDimensionScoresLegacy,
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
  chart?: CompatibilityChartSnapshot | null
}

interface PersonalityFitInput {
  scores?: CompatibilityDimensionScoresLegacy | null
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
  dimension?: keyof CompatibilityDimensionScoresLegacy | string
}

export interface PersonalityDimension {
  key: 'expression' | 'decision' | 'intimacy' | 'emotion' | 'pressure' | 'overview'
  label: string
  detail: string
}

export interface ParticipantPortrait {
  name: string
  role: 'self' | 'partner'
  headline: string
  dimensions: PersonalityDimension[]
  hasStructuredData: boolean
}

export interface PersonalityFitSummary {
  questionLabel: string
  stageLabel: string
  matchType: string
  matchTypeDescription: string
  headline: string
  summary: string
  selfPortrait: ParticipantPortrait
  partnerPortrait: ParticipantPortrait
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

// ---------------------------------------------------------------------------
// 经典映射引擎：各自命盘 → 个人性格画像 / 双方差异对照（确定性，不依赖合盘分数）
// ---------------------------------------------------------------------------

type TenGodGroup = 'bijie' | 'shishang' | 'cai' | 'guansha' | 'yin'
type ChartStrength = 'strong' | 'weak' | 'balanced'
type Pace = 'fast' | 'mid' | 'steady' | 'slow'
type Wuxing = '木' | '火' | '土' | '金' | '水'

const TEN_GOD_GROUP: Record<string, TenGodGroup> = {
  比肩: 'bijie', 劫财: 'bijie',
  食神: 'shishang', 伤官: 'shishang',
  正财: 'cai', 偏财: 'cai',
  正官: 'guansha', 七杀: 'guansha', 偏官: 'guansha',
  正印: 'yin', 偏印: 'yin', 枭神: 'yin',
}

const GAN_WUXING: Record<string, Wuxing> = {
  甲: '木', 乙: '木', 丙: '火', 丁: '火', 戊: '土',
  己: '土', 庚: '金', 辛: '金', 壬: '水', 癸: '水',
}

// 五行生：key 生 value；五行克：key 克 value
const SHENG: Record<Wuxing, Wuxing> = { 木: '火', 火: '土', 土: '金', 金: '水', 水: '木' }
const KE: Record<Wuxing, Wuxing> = { 木: '土', 土: '水', 水: '火', 火: '金', 金: '木' }
// 生我者（印）：value 生 key
const GENERATED_BY: Record<Wuxing, Wuxing> = { 木: '水', 火: '木', 土: '火', 金: '土', 水: '金' }
const WUXING_KEY: Record<Wuxing, 'mu' | 'huo' | 'tu' | 'jin' | 'shui'> = {
  木: 'mu', 火: 'huo', 土: 'tu', 金: 'jin', 水: 'shui',
}

const GROUP_LABEL: Record<TenGodGroup, string> = {
  bijie: '比劫', shishang: '食伤', cai: '财星', guansha: '官杀', yin: '印星',
}
const GROUP_PERSONA: Record<TenGodGroup, string> = {
  bijie: '独立好胜', shishang: '才情外放', cai: '务实经营', guansha: '自律守序', yin: '稳重内省',
}
const GROUP_PACE: Record<TenGodGroup, Pace> = {
  bijie: 'fast', shishang: 'fast', cai: 'mid', guansha: 'steady', yin: 'slow',
}

const EXPRESSION: Record<TenGodGroup, string> = {
  bijie: '表达直率有主见，不太绕弯，但容易先入为主。',
  shishang: '表达欲强、直接外放，想法和感受容易主动说出来。',
  cai: '沟通务实、直奔主题，更关注结果而非情绪铺陈。',
  guansha: '表达克制讲分寸，在意场合与对错，话不轻易出口。',
  yin: '偏内敛温和，倾向先消化再开口，需要被耐心倾听。',
}
const DECISION: Record<TenGodGroup, string> = {
  bijie: '决策果断、行动力强，认定了就推进，节奏偏快。',
  shishang: '决策灵活跟感觉走，点子多但容易临时改主意。',
  cai: '决策务实看性价比，会盘算投入产出再定节奏。',
  guansha: '决策自律守规则，权衡责任与后果，节奏偏稳。',
  yin: '决策谨慎慢热，倾向想清楚、求稳妥后再动。',
}
const INTIMACY: Record<TenGodGroup, string> = {
  bijie: '在亲密里需要平等和各自空间，看重被尊重而非被管。',
  shishang: '需要新鲜感和情绪共鸣，看重对方能接住自己的表达。',
  cai: '需要被实在地回应和经营，看重对方的投入与陪伴。',
  guansha: '需要被认可与承诺感，看重关系的确定性和分寸。',
  yin: '在亲密里最需要安全感和被照顾，看重对方稳定可靠。',
}
const EMOTION_WX: Record<Wuxing, string> = {
  木: '情绪温和但有韧劲，受触动时会持续较真。',
  火: '情绪来得快、表露明显，开心或不快都写在脸上。',
  土: '情绪沉稳、不易被带动，闷起来不爱多说。',
  金: '情绪克制、偏冷静理性，习惯压下情绪讲道理。',
  水: '情绪内化、想得多不易外显，容易自己消化。',
}
const PRESSURE: Record<ChartStrength, string> = {
  strong: '压力下倾向硬扛、正面应对，不轻易示弱。',
  weak: '压力下容易回避或退缩，需要外部支持和确认。',
  balanced: '压力下会先观望评估，再决定进退。',
}

function getDayWuxing(chart: CompatibilityChartSnapshot): Wuxing | '' {
  const fromField = chart.day_gan_wuxing as Wuxing | undefined
  if (fromField && WUXING_KEY[fromField]) return fromField
  return GAN_WUXING[chart.day_gan] || ''
}

function collectTenGods(chart: CompatibilityChartSnapshot): string[] {
  const gans = [chart.year_gan_shishen, chart.month_gan_shishen, chart.day_gan_shishen, chart.hour_gan_shishen]
  const zhis = [chart.year_zhi_shishen, chart.month_zhi_shishen, chart.day_zhi_shishen, chart.hour_zhi_shishen]
  const out: string[] = []
  gans.forEach(g => { if (g) out.push(g) })
  zhis.forEach(arr => { if (Array.isArray(arr)) arr.forEach(s => { if (s) out.push(s) }) })
  return out
}

// 主导十神：优先用命格（ming_ge），否则按各柱十神频次统计；归并为五类
function dominantGroup(chart: CompatibilityChartSnapshot): TenGodGroup | null {
  if (chart.ming_ge) {
    for (const key of Object.keys(TEN_GOD_GROUP)) {
      if (chart.ming_ge.includes(key)) return TEN_GOD_GROUP[key]
    }
  }
  const counts: Record<TenGodGroup, number> = { bijie: 0, shishang: 0, cai: 0, guansha: 0, yin: 0 }
  let total = 0
  collectTenGods(chart).forEach(name => {
    const group = TEN_GOD_GROUP[name]
    if (group) { counts[group] += 1; total += 1 }
  })
  if (total === 0) return null
  const priority: TenGodGroup[] = ['guansha', 'yin', 'cai', 'shishang', 'bijie']
  return priority.reduce((best, g) => (counts[g] > counts[best] ? g : best), priority[0])
}

// 粗粒度旺衰：日主同类（比劫）+ 生我（印）五行占比相对总量
function chartStrength(chart: CompatibilityChartSnapshot): ChartStrength {
  const wx = getDayWuxing(chart)
  const w = chart.wuxing
  if (!wx || !w) return 'balanced'
  const total = w.mu + w.huo + w.tu + w.jin + w.shui
  if (total <= 0) return 'balanced'
  const support = w[WUXING_KEY[wx]] + w[WUXING_KEY[GENERATED_BY[wx]]]
  const ratio = support / total
  if (ratio >= 0.5) return 'strong'
  if (ratio <= 0.3) return 'weak'
  return 'balanced'
}

function genericDimensions(): PersonalityDimension[] {
  return [{
    key: 'overview',
    label: '相处倾向',
    detail: '当前命盘信息不足，待补充命盘细节后可细化性格画像。',
  }]
}

function simplifiedDimensions(wx: Wuxing | '', strength: ChartStrength): PersonalityDimension[] {
  return [
    {
      key: 'emotion',
      label: '情绪反应',
      detail: wx ? EMOTION_WX[wx] : '情绪表现需结合更完整命盘判断。',
    },
    { key: 'pressure', label: '压力下的样子', detail: PRESSURE[strength] },
    {
      key: 'overview',
      label: '相处倾向',
      detail: '命盘十神信息有限，先用日主五行给出基础底色，生成深度报告后可进一步细化。',
    },
  ]
}

function fullDimensions(group: TenGodGroup, wx: Wuxing | '', strength: ChartStrength): PersonalityDimension[] {
  return [
    { key: 'expression', label: '表达 / 沟通', detail: EXPRESSION[group] },
    { key: 'decision', label: '决策与节奏', detail: DECISION[group] },
    { key: 'intimacy', label: '亲密里的核心需求', detail: INTIMACY[group] },
    {
      key: 'emotion',
      label: '情绪反应',
      detail: wx ? EMOTION_WX[wx] : '情绪表现需结合更完整命盘判断。',
    },
    { key: 'pressure', label: '压力下的样子', detail: PRESSURE[strength] },
  ]
}

export function buildParticipantPortrait(
  chart: CompatibilityChartSnapshot | null | undefined,
  name: string,
  role: 'self' | 'partner'
): ParticipantPortrait {
  const dayGan = chart?.day_gan || ''
  if (!chart || !dayGan) {
    return { name, role, headline: `${name}的性格画像`, dimensions: genericDimensions(), hasStructuredData: false }
  }
  const wx = getDayWuxing(chart)
  const group = dominantGroup(chart)
  const strength = chartStrength(chart)
  if (!group) {
    return {
      name,
      role,
      headline: `${name}：日主${dayGan}${wx ? `（${wx}）` : ''}`,
      dimensions: simplifiedDimensions(wx, strength),
      hasStructuredData: false,
    }
  }
  return {
    name,
    role,
    headline: `${name}：日主${dayGan}（${wx}）· ${GROUP_LABEL[group]}主导，${GROUP_PERSONA[group]}`,
    dimensions: fullDimensions(group, wx, strength),
    hasStructuredData: true,
  }
}

interface ChartSignals {
  wx: Wuxing | ''
  group: TenGodGroup | null
  strength: ChartStrength
  pace: Pace
}

function chartSignals(chart: CompatibilityChartSnapshot | null | undefined): ChartSignals | null {
  if (!chart || !chart.day_gan) return null
  const group = dominantGroup(chart)
  return {
    wx: getDayWuxing(chart),
    group,
    strength: chartStrength(chart),
    pace: group ? GROUP_PACE[group] : 'mid',
  }
}

const FAST_PACE: Pace[] = ['fast']
const SLOW_PACE: Pace[] = ['steady', 'slow']

export function buildPersonalityContrast(
  selfChart: CompatibilityChartSnapshot | null | undefined,
  partnerChart: CompatibilityChartSnapshot | null | undefined,
  selfName: string,
  partnerName: string
): { fitPoints: PersonalityPoint[]; clashPoints: PersonalityPoint[] } {
  const a = chartSignals(selfChart)
  const b = chartSignals(partnerChart)
  const fit: PersonalityPoint[] = []
  const clash: PersonalityPoint[] = []

  if (a && b) {
    // —— 自然合的地方 ——
    const careGiver = (g: TenGodGroup | null) => g === 'shishang' || g === 'cai'
    if ((a.group === 'yin' && careGiver(b.group)) || (b.group === 'yin' && careGiver(a.group))) {
      fit.push({ title: '照顾与被照顾自然咬合', detail: '一方更需要被照顾、被回应，另一方习惯主动给予和经营，需求刚好对得上。' })
    }
    if (a.wx && b.wx && (SHENG[a.wx] === b.wx || SHENG[b.wx] === a.wx)) {
      const src = SHENG[a.wx] === b.wx ? { n: selfName, w: a.wx, m: partnerName, mw: b.wx } : { n: partnerName, w: b.wx, m: selfName, mw: a.wx }
      fit.push({ title: '五行相生有滋养', detail: `${src.n}的${src.w}生${src.m}的${src.mw}，能量上一方能托住另一方，相处有自然顺位。` })
    }
    const aFast = FAST_PACE.includes(a.pace)
    const bSlow = SLOW_PACE.includes(b.pace)
    const bFast = FAST_PACE.includes(b.pace)
    const aSlow = SLOW_PACE.includes(a.pace)
    if ((aFast && bSlow) || (bFast && aSlow)) {
      fit.push({ title: '一推一稳能承接', detail: '一个偏推进、一个偏把稳，节奏一急一缓反而容易互相补位。' })
    }
    if (a.group === 'yin' && b.group === 'yin') {
      fit.push({ title: '都偏稳重内敛', detail: '两人都喜欢安静、求稳，相处摩擦少、节奏接近。' })
    }
  }

  if (a && b) {
    // —— 容易冲突的地方 ——
    const aggressive = (g: TenGodGroup | null) => g === 'bijie' || g === 'guansha'
    if (aggressive(a.group) && aggressive(b.group)) {
      clash.push({ title: '都强势、爱争主导', detail: '两人都偏强势好胜，容易在谁说了算上较劲、互不相让。' })
    }
    if (a.wx && b.wx && (KE[a.wx] === b.wx || KE[b.wx] === a.wx)) {
      const src = KE[a.wx] === b.wx ? { n: selfName, w: a.wx, m: partnerName, mw: b.wx } : { n: partnerName, w: b.wx, m: selfName, mw: a.wx }
      clash.push({ title: '五行相克易顶撞', detail: `${src.n}的${src.w}克${src.m}的${src.mw}，价值观或做法上容易直接顶上。` })
    }
    const aFast = FAST_PACE.includes(a.pace)
    const bSlow = SLOW_PACE.includes(b.pace)
    const bFast = FAST_PACE.includes(b.pace)
    const aSlow = SLOW_PACE.includes(a.pace)
    if ((aFast && bSlow) || (bFast && aSlow)) {
      clash.push({ title: '一快一慢易摩擦', detail: '一个想快、一个要慢，节奏差异容易积累出催促与拖延的摩擦。' })
    }
    if (a.strength === 'strong' && b.strength === 'strong') {
      clash.push({ title: '都旺、都不爱示弱', detail: '两人都偏旺，冲突时容易硬碰硬，谁也不先退一步。' })
    }
    const outward = (g: TenGodGroup | null) => g === 'shishang'
    const inward = (g: TenGodGroup | null) => g === 'yin' || g === 'guansha'
    if ((outward(a.group) && inward(b.group)) || (outward(b.group) && inward(a.group))) {
      clash.push({ title: '表达冷热不一致', detail: '一个表达直接外放、一个习惯收着，容易一个嫌冷、一个嫌冲。' })
    }
  }

  if (fit.length === 0) {
    fit.push({ title: '先看真实互动', detail: '从命盘结构看没有特别突出的天然契合点，更要靠真实相处去确认默契。' })
  }
  if (clash.length === 0) {
    clash.push({ title: '冲突点不突出', detail: '命盘结构上没有明显对冲，但仍要观察现实节奏和情绪回应是否对得上。' })
  }

  return { fitPoints: fit.slice(0, 3), clashPoints: clash.slice(0, 3) }
}

function buildCommunicationGuidance(
  scores: CompatibilityDimensionScoresLegacy | null,
  clashPoints: PersonalityPoint[]
): PersonalityPoint[] {
  if (scores) {
    const communication = clamp(scores.communication)
    return [
      {
        title: communication >= 62 ? '沟通有修复空间' : '沟通需重点验证',
        detail: communication >= 62
          ? '出现误会时有机会通过解释、放慢节奏来修复。'
          : '容易在表达速度、情绪回应或冲突修复上反复错位。',
        dimension: 'communication',
      },
      {
        title: '沟通建议',
        detail: communication >= 62
          ? '适合把需求、边界和下一步计划说具体，用真实互动继续确认默契。'
          : '先降低情绪密度，遇到分歧时确认事实和需求，再讨论投入或承诺。',
        dimension: 'communication',
      },
    ]
  }
  const topClash = clashPoints[0]
  return [
    {
      title: '沟通建议',
      detail: topClash
        ? `针对「${topClash.title}」，遇到分歧时先确认彼此的事实和需求，再讨论怎么做。`
        : '把需求、边界和下一步计划说具体，用真实互动慢慢确认默契。',
    },
  ]
}

function getEvidenceTargets(evidences: Array<Partial<CompatibilityEvidence>>) {
  return Array.from(new Set(
    evidences
      .map(e => e.evidence_key || e.id)
      .filter((key): key is string => Boolean(key))
  ))
}

export function buildPersonalityFitSummary(input: PersonalityFitInput): PersonalityFitSummary {
  const evidences = Array.isArray(input.evidences) ? input.evidences : []
  const scores = input.scores || null
  const selfName = input.self?.name || input.selfName || '我'
  const partnerName = input.partner?.name || input.partnerName || '对方'
  const selfChart = input.self?.chart || null
  const partnerChart = input.partner?.chart || null

  const questionLabel = getCompatibilityQuestionLabel(input.primaryQuestion)
  const stageLabel = getCompatibilityStageLabel(input.relationshipStage)
  const matchType = scores
    ? getPersonalityMatchType(scores, input.primaryQuestion, input.relationshipStage)
    : '待磨合观察型'
  const matchTypeDescription = scores
    ? getPersonalityMatchTypeDescription(matchType)
    : '先看双方各自性格与差异点，再决定怎么相处、要不要继续投入。'

  const selfPortrait = buildParticipantPortrait(selfChart, selfName, 'self')
  const partnerPortrait = buildParticipantPortrait(partnerChart, partnerName, 'partner')
  const contrast = buildPersonalityContrast(selfChart, partnerChart, selfName, partnerName)

  const diagnosisSummary = input.relationshipDiagnosis?.summary?.trim()
  const overallScore = scores
    ? Math.round(average([
        clamp(scores.attraction),
        clamp(scores.stability),
        clamp(scores.communication),
        clamp(scores.practicality),
      ]))
    : null

  const headline = scores
    ? `性格合不合：${matchType}`
    : `${selfName} × ${partnerName}：双方性格画像与差异`
  const summary = diagnosisSummary || (scores
    ? `从四维结构看，这段关系属于${matchType}，整体匹配约为 ${overallScore} 分。先看双方相处节奏是否能稳定回应，再决定投入强度。`
    : `下面先分别刻画 ${selfName} 与 ${partnerName} 的性格，再点出两人自然合的地方与容易冲突的地方。`)

  return {
    questionLabel,
    stageLabel,
    matchType,
    matchTypeDescription,
    headline,
    summary,
    selfPortrait,
    partnerPortrait,
    fitPoints: contrast.fitPoints,
    clashPoints: contrast.clashPoints,
    communicationGuidance: buildCommunicationGuidance(scores, contrast.clashPoints),
    evidenceTargets: getEvidenceTargets(evidences),
    reportNote: input.hasReport
      ? '深度解读已生成，可结合下方报告细看性格判断的细节。'
      : '当前性格画像来自双方命盘的结构化数据；生成深度解读后，可进一步细化双方相处模式。',
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
