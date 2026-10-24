// The miniapp currently uses one fixed chart and local calendar labels. This
// small rule layer is a UI preview, not a complete daily selection algorithm.
const stemElement = {
  甲: '木', 乙: '木', 丙: '火', 丁: '火', 戊: '土', 己: '土',
  庚: '金', 辛: '金', 壬: '水', 癸: '水',
}

const branchElement = {
  寅: '木', 卯: '木', 巳: '火', 午: '火', 辰: '土', 戌: '土',
  丑: '土', 未: '土', 申: '金', 酉: '金', 亥: '水', 子: '水',
}

const oppositeBranch = {
  子: '午', 午: '子', 丑: '未', 未: '丑', 寅: '申', 申: '寅',
  卯: '酉', 酉: '卯', 辰: '戌', 戌: '辰', 巳: '亥', 亥: '巳',
}

function direction(element, favorable, adverse) {
  if (favorable.includes(element)) return 'favorable'
  if (adverse.includes(element)) return 'adverse'
  return 'neutral'
}

function explain(part, symbol, element, directionValue) {
  const relationship = directionValue === 'favorable'
    ? '落入示例命盘的扶抑喜用'
    : directionValue === 'adverse'
      ? '落入示例命盘的扶抑忌神'
      : '未落入当前示例喜忌方向'
  return `${part}「${symbol}」属${element}，${relationship}${part === '日支主气' ? '；这里只按主气五行展示' : ''}。`
}

function insufficient(date, dayGanZhi) {
  return {
    date,
    dayGanZhi: dayGanZhi || '—',
    status: 'insufficient',
    label: '依据不足',
    summary: '该日期或示例命盘资料不足，暂不展示个人关系。',
    evidence: [],
    activities: [],
  }
}

function buildPersonalDay(date, calendarRecord, chart) {
  const dayGanZhi = calendarRecord && calendarRecord.dayGanZhi
  const profile = chart && chart.personalDay
  const gan = typeof dayGanZhi === 'string' ? dayGanZhi[0] : ''
  const zhi = typeof dayGanZhi === 'string' ? dayGanZhi[1] : ''
  const favorable = profile && profile.fuyiYongshen
  const adverse = profile && profile.fuyiJishen
  if (!stemElement[gan] || !branchElement[zhi] || !profile ||
      typeof favorable !== 'string' || !/^[木火土金水]+$/.test(favorable) ||
      typeof adverse !== 'string' || !/^[木火土金水]+$/.test(adverse) ||
      [...favorable].some(element => adverse.includes(element)) ||
      !branchElement[profile.dayBranch]) {
    return insufficient(date, dayGanZhi)
  }

  const ganDirection = direction(stemElement[gan], favorable, adverse)
  const zhiDirection = direction(branchElement[zhi], favorable, adverse)
  const clash = oppositeBranch[zhi] === profile.dayBranch
  const evidence = [
    explain('日干', gan, stemElement[gan], ganDirection),
    explain('日支主气', zhi, branchElement[zhi], zhiDirection),
  ]
  if (clash) {
    evidence.push(`所选日支「${zhi}」与示例命盘日支「${profile.dayBranch}」相冲，重要安排需要另行核对。`)
  }

  let status = 'neutral'
  if (clash || (ganDirection === 'adverse' && zhiDirection === 'adverse')) {
    status = 'caution'
  } else if (ganDirection === 'favorable' && zhiDirection === 'favorable') {
    status = 'supportive'
  } else if (ganDirection !== 'neutral' || zhiDirection !== 'neutral') {
    status = 'mixed'
  }

  const labels = {
    supportive: '偏合',
    mixed: '关系复杂',
    caution: '需留意',
    neutral: '方向中性',
  }
  const summaries = {
    supportive: '所选日干与日支主气均与示例命盘的扶抑喜用同向，可作为日程规划时的一项参考。',
    mixed: '所选日干、日支主气的方向并不一致，适合先看具体依据，再安排重要事项。',
    caution: '所选日与示例命盘存在忌神或日支相冲信号，重要事项建议多核对现实条件。',
    neutral: '当前简化规则没有明显方向，不能据此判断事项是否适宜。',
  }

  const moving = clash
    ? { tip: '先核对安排', detail: '日支与示例命盘日支相冲；若计划搬家，可先核对人员、时间和实际搬运安排。' }
    : status === 'supportive'
      ? { tip: '可先规划', detail: '示例流日关系偏合，可先规划搬家流程；这不是完整择日结果。' }
      : { tip: '保留备选', detail: '当前关系不足以单独确定搬家日期，可结合家人、时间与实际条件选择。' }
  const finance = status === 'caution'
    ? { tip: '注意风险控制', detail: '示例流日有需留意的信号；涉及投资时仍应独立评估产品、预算与风险。' }
    : status === 'supportive'
      ? { tip: '适合梳理计划', detail: '可用这一天整理预算与计划；示例干支关系不构成投资买卖依据。' }
      : { tip: '先核对预算', detail: '可先检查预算与风险承受能力；示例干支关系不构成投资买卖依据。' }

  return {
    date,
    dayGanZhi,
    status,
    label: labels[status],
    summary: summaries[status],
    evidence,
    activities: [
      { id: 'moving', title: '搬家', tip: moving.tip, detail: moving.detail },
      { id: 'finance', title: '财务安排', tip: finance.tip, detail: finance.detail },
    ],
  }
}

module.exports = { buildPersonalDay }
