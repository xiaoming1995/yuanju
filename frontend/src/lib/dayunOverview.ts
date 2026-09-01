export type Polarity = 'xi' | 'ji' | 'zhong'
type Relation = 'tongGen' | 'gaiTou' | 'jieJiao' | 'none'
type Fit = 'buZu' | 'weiDaoWei' | 'weiJi' | 'skip'

export interface DayunOverviewInput {
  dayun: {
    gan: string
    zhi: string
    gan_shishen: string
    zhi_shishen: string
    di_shi: string
  }
  yongshen: string
  jishen: string
  natalDayMasterStrength?: string
  tiaohou?: {
    expected: string[]
    tou: string[]
    cang: string[]
    text: string
  } | null
}

export interface DayunOverviewOutput {
  prose: string
  proseLay: string
  trendKeywords: string
  ganPolarity: Polarity
  zhiPolarity: Polarity
}

const FALLBACK_PROSE = '选择一段大运后查看该十年流年节奏。'
const FALLBACK_KEYWORDS = '节奏 · 观察 · 平衡'

const GAN_WUXING_CN: Record<string, string> = {
  甲: '木', 乙: '木', 丙: '火', 丁: '火', 戊: '土',
  己: '土', 庚: '金', 辛: '金', 壬: '水', 癸: '水',
}

const ZHI_MAIN_WUXING: Record<string, string> = {
  子: '水', 丑: '土', 寅: '木', 卯: '木',
  辰: '土', 巳: '火', 午: '火', 未: '土',
  申: '金', 酉: '金', 戌: '土', 亥: '水',
}

const ZHI_MAIN_GAN: Record<string, string> = {
  子: '癸', 丑: '己', 寅: '甲', 卯: '乙',
  辰: '戊', 巳: '丙', 午: '丁', 未: '己',
  申: '庚', 酉: '辛', 戌: '戊', 亥: '壬',
}

const K_GRAPH: Record<string, string> = {
  木: '土', 土: '水', 水: '火', 火: '金', 金: '木',
}

const DI_SHI_BUCKET: Record<string, 'wang' | 'mid' | 'shuai'> = {
  帝旺: 'wang', 临官: 'wang', 长生: 'wang', 冠带: 'wang',
  沐浴: 'mid',  养: 'mid',     胎: 'mid',   墓: 'mid',
  衰: 'shuai',  病: 'shuai',   死: 'shuai', 绝: 'shuai',
}

const TREND: Record<string, { xi: string; ji: string }> = {
  比肩: { xi: '同道 · 自立 · 稳进', ji: '分薄 · 竞争 · 节制' },
  劫财: { xi: '合伙 · 协力 · 取舍', ji: '损财 · 争夺 · 化解' },
  食神: { xi: '表达 · 享受 · 作品', ji: '泄气 · 分心 · 节用' },
  伤官: { xi: '突破 · 才名 · 创意', ji: '是非 · 锋芒 · 收敛' },
  正财: { xi: '经营 · 责任 · 积累', ji: '负重 · 守财 · 量力' },
  偏财: { xi: '机会 · 流动 · 人脉', ji: '财去 · 投机 · 谨慎' },
  正官: { xi: '事业 · 晋升 · 成就', ji: '约束 · 规矩 · 顺应' },
  七杀: { xi: '突破 · 立威 · 决断', ji: '压力 · 守势 · 化解' },
  正印: { xi: '学习 · 贵人 · 资质', ji: '迟滞 · 内耗 · 取舍' },
  偏印: { xi: '研究 · 灵感 · 转型', ji: '孤独 · 怀疑 · 沉淀' },
}

const POL_LABEL: Record<Polarity, string> = { xi: '喜', ji: '忌', zhong: '中' }

const HEADING_TONE: Record<string, Record<Polarity, string>> = {
  比肩: { xi: '同道协作的十年',     ji: '竞争分薄的十年',     zhong: '节奏中性的十年' },
  劫财: { xi: '合作分担的十年',     ji: '损财争夺的十年',     zhong: '节奏中性的十年' },
  食神: { xi: '表达与享受的十年',   ji: '易分心耗神的十年',   zhong: '节奏中性的十年' },
  伤官: { xi: '突破创新的十年',     ji: '锋芒易招是非的十年', zhong: '节奏中性的十年' },
  正财: { xi: '稳健积累的十年',     ji: '量力守财的十年',     zhong: '节奏中性的十年' },
  偏财: { xi: '机会与人脉的十年',   ji: '财来财去的十年',     zhong: '节奏中性的十年' },
  正官: { xi: '事业晋升的十年',     ji: '规则约束多的十年',   zhong: '节奏中性的十年' },
  七杀: { xi: '适合主动出击的十年', ji: '压力偏大的十年',     zhong: '节奏中性的十年' },
  正印: { xi: '学习与贵人的十年',   ji: '易内耗的十年',       zhong: '节奏中性的十年' },
  偏印: { xi: '研究与转型的十年',   ji: '易孤独沉淀的十年',   zhong: '节奏中性的十年' },
}

const WUXING_MEANING: Record<string, string> = {
  木: '生发 / 条理',
  火: '热情 / 行动',
  土: '稳重 / 物质',
  金: '决断 / 收敛',
  水: '柔韧 / 智慧',
}

function resolvePolarity(wuxing: string, yong: string, ji: string): Polarity {
  if (!wuxing) return 'zhong'
  if (yong && yong.includes(wuxing)) return 'xi'
  if (ji && ji.includes(wuxing)) return 'ji'
  return 'zhong'
}

function resolveGanZhiRelation(ganWx: string, zhiWx: string): Relation {
  if (!ganWx || !zhiWx) return 'none'
  if (ganWx === zhiWx) return 'tongGen'
  if (K_GRAPH[ganWx] === zhiWx) return 'gaiTou'
  if (K_GRAPH[zhiWx] === ganWx) return 'jieJiao'
  return 'none'
}

function resolveTiaohouFit(
  input: DayunOverviewInput,
  relation: Relation,
): { fit: Fit; missingWx?: string; matchedGan?: string; coverGan?: string } {
  const t = input.tiaohou
  if (!t || !t.expected || t.expected.length === 0) return { fit: 'skip' }

  const have = new Set([...(t.tou ?? []), ...(t.cang ?? [])])
  const missingGans = t.expected.filter(g => !have.has(g))
  if (missingGans.length === 0) return { fit: 'skip' }

  const missingWxSet = new Set(
    missingGans.map(g => GAN_WUXING_CN[g]).filter(Boolean),
  )
  if (missingWxSet.size === 0) return { fit: 'skip' }

  const ganWx = GAN_WUXING_CN[input.dayun.gan]
  const zhiMainWx = ZHI_MAIN_WUXING[input.dayun.zhi]
  const ganMatches = missingWxSet.has(ganWx)
  const zhiMatches = missingWxSet.has(zhiMainWx)

  if (!ganMatches && !zhiMatches) {
    return { fit: 'weiJi', missingWx: GAN_WUXING_CN[missingGans[0]] }
  }

  let matchedGan: string
  let matchedWx: string
  if (ganMatches) {
    matchedGan = input.dayun.gan
    matchedWx = ganWx
  } else {
    matchedGan = ZHI_MAIN_GAN[input.dayun.zhi] ?? input.dayun.zhi
    matchedWx = zhiMainWx
  }

  if (relation === 'gaiTou') {
    return { fit: 'weiDaoWei', missingWx: matchedWx, coverGan: input.dayun.gan }
  }
  return { fit: 'buZu', missingWx: matchedWx, matchedGan }
}

const TEN_GOD_ROLE: Record<string, string> = {
  比肩: '同辈协作与资源分配', 劫财: '合作竞争与资源取舍',
  食神: '表达、作品与稳定输出', 伤官: '创新、表达与突破',
  正财: '经营、责任与积累', 偏财: '机会、人脉与流动资源',
  正官: '规则、职位与责任', 七杀: '执行、规制与决断',
  正印: '学习、资质与支持', 偏印: '研究、资源与思虑',
}

const NATAL_STRENGTH_LABEL: Record<string, { expert: string; lay: string }> = {
  vstrong: { expert: '原局身极强', lay: '原局底子偏强' },
  strong: { expert: '原局身强', lay: '原局底子偏强' },
  neutral: { expert: '原局中和', lay: '原局底子较均衡' },
  weak: { expert: '原局身弱', lay: '原局底子偏弱' },
  vweak: { expert: '原局身极弱', lay: '原局底子偏弱' },
}

function fuyiRoleText(
  stemOrBranch: string,
  wuxing: string,
  tenGod: string,
  polarity: Polarity,
): string {
  const role = TEN_GOD_ROLE[tenGod] ?? '阶段作用'
  if (polarity === 'xi') return `${stemOrBranch}${wuxing}为扶抑喜用，${tenGod}主${role}`
  if (polarity === 'ji') return `${stemOrBranch}${wuxing}为扶抑忌，${tenGod}主${role}`
  return `${stemOrBranch}${wuxing}为扶抑中性，${tenGod}主${role}`
}

function fuyiRoleLayText(
  stemOrBranch: string,
  tenGod: string,
  polarity: Polarity,
): string {
  const role = TEN_GOD_ROLE[tenGod] ?? '阶段作用'
  if (polarity === 'xi') return `${stemOrBranch}带来的${role}可借力`
  if (polarity === 'ji') return `${stemOrBranch}带来的${role}需要节制`
  return `${stemOrBranch}带来${role}的中性影响`
}

function branchStageText(zhi: string, diShi: string, polarity: Polarity): string {
  const bucket = DI_SHI_BUCKET[diShi]
  if (!bucket) return ''
  const stage = `${zhi}${diShi}`
  if (polarity === 'xi') {
    if (bucket === 'wang') return `${stage}，喜用力量更易发挥`
    if (bucket === 'mid') return `${stage}，喜用作用平稳`
    return `${stage}，喜用力量较弱`
  }
  if (polarity === 'ji') {
    if (bucket === 'wang') return `${stage}，忌神力量较显`
    if (bucket === 'mid') return `${stage}，忌神作用仍在`
    return `${stage}，忌神力量受限`
  }
  if (bucket === 'wang') return `${stage}，地支作用较明显，仍以原局扶抑喜忌为准`
  if (bucket === 'mid') return `${stage}，地支作用平稳，仍以原局扶抑喜忌为准`
  return `${stage}，地支作用较弱，仍以原局扶抑喜忌为准`
}

function branchStageLayText(zhi: string, diShi: string, polarity: Polarity): string {
  const bucket = DI_SHI_BUCKET[diShi]
  if (!bucket) return ''
  const stage = `${zhi}${diShi}`
  if (polarity === 'xi') {
    if (bucket === 'wang') return `${stage}，这股助力更明显`
    if (bucket === 'mid') return `${stage}，这股助力较平稳`
    return `${stage}，这股助力较弱`
  }
  if (polarity === 'ji') {
    if (bucket === 'wang') return `${stage}，这部分压力较显`
    if (bucket === 'mid') return `${stage}，这部分压力仍在`
    return `${stage}，这部分压力受限`
  }
  return bucket === 'wang' ? `${stage}，影响较明显` : `${stage}，影响较平稳`
}

function body3(
  fit: Fit,
  missingWx?: string,
  matchedGan?: string,
  coverGan?: string,
): string {
  switch (fit) {
    case 'buZu':
      return `${matchedGan ?? ''}${missingWx ?? ''}透出，正补足命局所缺调候`
    case 'weiDaoWei':
      return `命局所需${missingWx ?? ''}虽现于运中，却被${coverGan ?? ''}压制，调候未到位`
    case 'weiJi':
      return `命局所缺${missingWx ?? ''}未在此运补足，需外接调候助力`
    case 'skip':
      return ''
  }
}

function body3Lay(
  fit: Fit,
  missingWx?: string,
  coverGan?: string,
): string {
  const meaning = missingWx ? WUXING_MEANING[missingWx] ?? '' : ''
  const wxLabel = missingWx ? `${missingWx}气（${meaning}）` : ''
  switch (fit) {
    case 'buZu':
      return `命局缺的${wxLabel}在这十年补上，体感会比较顺`
    case 'weiDaoWei':
      return `命局缺的${wxLabel}虽现于运中，但被${coverGan ?? ''}压制，效果打折扣`
    case 'weiJi':
      return `命局缺的${wxLabel}这十年没补上，需要主动从外界补给`
    case 'skip':
      return ''
  }
}

export function buildDayunOverview(input: DayunOverviewInput): DayunOverviewOutput {
  const { dayun } = input
  const ganWx = GAN_WUXING_CN[dayun.gan]
  const zhiWx = ZHI_MAIN_WUXING[dayun.zhi]
  const inBucket = !!DI_SHI_BUCKET[dayun.di_shi]

  if (!ganWx || !zhiWx || !inBucket) {
    return {
      prose: FALLBACK_PROSE,
      proseLay: FALLBACK_PROSE,
      trendKeywords: FALLBACK_KEYWORDS,
      ganPolarity: 'zhong',
      zhiPolarity: 'zhong',
    }
  }

  const ganPolarity = resolvePolarity(ganWx, input.yongshen, input.jishen)
  const zhiPolarity = resolvePolarity(zhiWx, input.yongshen, input.jishen)
  const relation = resolveGanZhiRelation(ganWx, zhiWx)
  const { fit, missingWx, matchedGan, coverGan } = resolveTiaohouFit(input, relation)

  const hasPolarity = !!input.yongshen || !!input.jishen
  const heading = hasPolarity
    ? `${dayun.gan}${dayun.zhi}运（${dayun.gan_shishen}为${POL_LABEL[ganPolarity]}·${dayun.zhi_shishen}为${POL_LABEL[zhiPolarity]}）：`
    : `${dayun.gan}${dayun.zhi}运：`

  const natalStrength = NATAL_STRENGTH_LABEL[input.natalDayMasterStrength ?? '']
  const ganText = fuyiRoleText(dayun.gan, ganWx, dayun.gan_shishen, ganPolarity)
  const zhiText = fuyiRoleText(dayun.zhi, zhiWx, dayun.zhi_shishen, zhiPolarity)
  const stageText = branchStageText(dayun.zhi, dayun.di_shi, zhiPolarity)
  const body3text = body3(fit, missingWx, matchedGan, coverGan)

  let prose = `${heading}${natalStrength?.expert ? `${natalStrength.expert}；` : ''}${ganText}；${zhiText}`
  if (stageText) prose += `；${stageText}`
  prose += '。'
  if (body3text) prose += `${body3text}。`

  const tone = HEADING_TONE[dayun.gan_shishen]?.[ganPolarity] ?? '节奏中性的十年'
  const headingLay = hasPolarity
    ? `${dayun.gan}${dayun.zhi}运（${tone}）：`
    : `${dayun.gan}${dayun.zhi}运：`
  const body3LayText = body3Lay(fit, missingWx, coverGan)
  const ganLayText = fuyiRoleLayText(dayun.gan, dayun.gan_shishen, ganPolarity)
  const zhiLayText = fuyiRoleLayText(dayun.zhi, dayun.zhi_shishen, zhiPolarity)
  const stageLayText = branchStageLayText(dayun.zhi, dayun.di_shi, zhiPolarity)

  let proseLay = `${headingLay}${natalStrength?.lay ? `${natalStrength.lay}；` : ''}${ganLayText}；${zhiLayText}`
  if (stageLayText) proseLay += `；${stageLayText}`
  proseLay += '。'
  if (body3LayText) proseLay += `${body3LayText}。`

  const trendEntry = TREND[dayun.gan_shishen]
  const trendKeywords = trendEntry
    ? (ganPolarity === 'ji' ? trendEntry.ji : trendEntry.xi)
    : FALLBACK_KEYWORDS

  return { prose, proseLay, trendKeywords, ganPolarity, zhiPolarity }
}
