export type Polarity = 'xi' | 'ji' | 'zhong'
type Strength = 'wang' | 'ruo'
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
  wuxing: { mu: number; huo: number; tu: number; jin: number; shui: number }
  dayGanWuxing: string
  tiaohou?: {
    expected: string[]
    tou: string[]
    cang: string[]
    text: string
  } | null
}

export interface DayunOverviewOutput {
  prose: string
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

const HELP_MAP: Record<string, Array<'mu' | 'huo' | 'tu' | 'jin' | 'shui'>> = {
  木: ['mu', 'shui'],
  火: ['huo', 'mu'],
  土: ['tu', 'huo'],
  金: ['jin', 'tu'],
  水: ['shui', 'jin'],
}

const DI_SHI_BUCKET: Record<string, 'wang' | 'mid' | 'shuai'> = {
  帝旺: 'wang', 临官: 'wang', 长生: 'wang', 冠带: 'wang',
  沐浴: 'mid',  养: 'mid',     胎: 'mid',   墓: 'mid',
  衰: 'shuai',  病: 'shuai',   死: 'shuai', 绝: 'shuai',
}

const DI_SHI_LABEL: Record<'wang' | 'mid' | 'shuai', string> = {
  wang: '得位有力',
  mid: '态势中等',
  shuai: '气势减弱',
}

const BODY1: Record<string, Record<Strength, string>> = {
  比肩: { wang: '同行竞争分薄资源',     ruo: '兄弟朋友助身有力' },
  劫财: { wang: '损财争夺、合作伤利',   ruo: '同道分担、压力有人共担' },
  食神: { wang: '财源外吐、口腹之享',   ruo: '才华外泄、气力分散' },
  伤官: { wang: '才名突破、敢破规则',   ruo: '才华伤身、易招是非' },
  正财: { wang: '经营得利、稳定积累',   ruo: '财多身弱、力不从心' },
  偏财: { wang: '偏门机会、流动资金',   ruo: '财来财去、难以聚守' },
  正官: { wang: '事业晋升、责任加码',   ruo: '官杀压身、易受规则约束' },
  七杀: { wang: '立威破局、事业突破',   ruo: '身弱遭杀克身，压力与突发事件增多' },
  正印: { wang: '印重身旺反招迟滞',     ruo: '学习/贵人/资格类机会成形' },
  偏印: { wang: '转型旁门、思虑成局',   ruo: '灵感/研究/孤独感提升' },
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

function resolvePolarity(wuxing: string, yong: string, ji: string): Polarity {
  if (!wuxing) return 'zhong'
  if (yong && yong.includes(wuxing)) return 'xi'
  if (ji && ji.includes(wuxing)) return 'ji'
  return 'zhong'
}

function resolveDayStrength(
  wuxing: DayunOverviewInput['wuxing'] | undefined,
  dayGanWuxing: string,
): Strength {
  if (!wuxing) return 'ruo'
  const help = HELP_MAP[dayGanWuxing] ?? []
  const helpPct = help.reduce((s, k) => s + (wuxing[k] ?? 0), 0)
  return helpPct > 40 ? 'wang' : 'ruo'
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

function body2(diShi: string, gan: string, zhi: string, relation: Relation): string {
  const bucket = DI_SHI_BUCKET[diShi]
  if (!bucket) return ''
  let base = `${zhi}${diShi}${DI_SHI_LABEL[bucket]}`
  switch (relation) {
    case 'tongGen': base += `，${gan}通根${zhi}得力`; break
    case 'gaiTou':  base += `，但被${gan}盖头压制`; break
    case 'jieJiao': base += `，反被${zhi}截脚虚浮`; break
    case 'none': break
  }
  return base
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

export function buildDayunOverview(input: DayunOverviewInput): DayunOverviewOutput {
  const { dayun } = input
  const ganWx = GAN_WUXING_CN[dayun.gan]
  const zhiWx = ZHI_MAIN_WUXING[dayun.zhi]
  const inBody1 = !!BODY1[dayun.gan_shishen]
  const inBucket = !!DI_SHI_BUCKET[dayun.di_shi]

  if (!ganWx || !zhiWx || !inBody1 || !inBucket) {
    return {
      prose: FALLBACK_PROSE,
      trendKeywords: FALLBACK_KEYWORDS,
      ganPolarity: 'zhong',
      zhiPolarity: 'zhong',
    }
  }

  const ganPolarity = resolvePolarity(ganWx, input.yongshen, input.jishen)
  const zhiPolarity = resolvePolarity(zhiWx, input.yongshen, input.jishen)
  const baseStrength = resolveDayStrength(input.wuxing, input.dayGanWuxing)
  const dayStrength: Strength = ganPolarity === 'xi' ? 'wang' : baseStrength
  const relation = resolveGanZhiRelation(ganWx, zhiWx)
  const { fit, missingWx, matchedGan, coverGan } = resolveTiaohouFit(input, relation)

  const hasPolarity = !!input.yongshen || !!input.jishen
  const heading = hasPolarity
    ? `${dayun.gan}${dayun.zhi}运（${dayun.gan_shishen}为${POL_LABEL[ganPolarity]}·${dayun.zhi_shishen}为${POL_LABEL[zhiPolarity]}）：`
    : `${dayun.gan}${dayun.zhi}运：`

  const body1 = BODY1[dayun.gan_shishen][dayStrength]
  const body2text = body2(dayun.di_shi, dayun.gan, dayun.zhi, relation)
  const body3text = body3(fit, missingWx, matchedGan, coverGan)

  let prose = `${heading}${body1}；${body2text}。`
  if (body3text) prose += `${body3text}。`

  const trendEntry = TREND[dayun.gan_shishen]
  const trendKeywords = trendEntry
    ? (ganPolarity === 'ji' ? trendEntry.ji : trendEntry.xi)
    : FALLBACK_KEYWORDS

  return { prose, trendKeywords, ganPolarity, zhiPolarity }
}
