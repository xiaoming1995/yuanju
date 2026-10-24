const { buildPersonalDay } = require('./personal_day')
const { getJianChuInfo } = require('./jianchu')

const FIRST_DATE = '2025-01-01'
const LAST_DATE = '2027-12-31'
const DAY_MS = 24 * 60 * 60 * 1000
const ACTIVITIES = [
  { id: 'moving', label: '搬家', tags: ['入宅', '移徙'], note: '对应传统标签「入宅」「移徙」；具体搬迁类型仍需核对。' },
  { id: 'marriage', label: '结婚', tags: ['嫁娶'], note: '仅按传统「嫁娶」标签演示，不包含订婚。' },
  { id: 'opening', label: '开业', tags: ['开市'], note: '仅按传统「开市」标签演示，不代表交易或投资时机。' },
  { id: 'travel', label: '出行', tags: ['出行'], note: '仅按传统「出行」标签演示，实际行程与安全安排另行核对。' },
  { id: 'groundbreaking', label: '动土', tags: ['动土'], note: '仅按传统「动土」标签演示，不包含其他施工事项。' },
]
const groupDefinitions = [
  { id: 'priority', label: '通用宜与示例关系同向' },
  { id: 'review', label: '有通用宜，仍需核对' },
  { id: 'conflict', label: '有通用宜忌冲突' },
  { id: 'insufficient', label: '依据不足' },
]

function dateKey(timestamp) {
  return new Date(timestamp).toISOString().slice(0, 10)
}

function parseDate(value) {
  if (typeof value !== 'string' || !/^\d{4}-\d{2}-\d{2}$/.test(value)) return null
  const [year, month, day] = value.split('-').map(Number)
  const timestamp = Date.UTC(year, month - 1, day)
  return dateKey(timestamp) === value ? timestamp : null
}

function defaultDateRange(now = new Date()) {
  const today = `${now.getFullYear()}-${`${now.getMonth() + 1}`.padStart(2, '0')}-${`${now.getDate()}`.padStart(2, '0')}`
  const timestamp = parseDate(today)
  if (timestamp === null || today < FIRST_DATE || today > LAST_DATE) {
    return { startDate: '', endDate: '', error: '演示日历范围为2025—2027年，请选择范围内的日期。' }
  }
  const endTimestamp = Math.min(timestamp + 29 * DAY_MS, parseDate(LAST_DATE))
  return { startDate: today, endDate: dateKey(endTimestamp), error: '' }
}

function validateDateRange(startDate, endDate) {
  const start = parseDate(startDate)
  const end = parseDate(endDate)
  if (start === null || end === null) return '请选择有效的起止日期。'
  if (startDate < FIRST_DATE || endDate > LAST_DATE) return '演示日历范围为2025—2027年。'
  if (start > end) return '起日不能晚于止日。'
  if ((end - start) / DAY_MS + 1 > 90) return '一次最多查看90个日历日。'
  return ''
}

function baseDay(activity, date, record, personalDay) {
  const tags = activity ? activity.tags : []
  const yi = record && Array.isArray(record.activityYi)
    ? record.activityYi.filter(label => tags.includes(label)) : []
  const ji = record && Array.isArray(record.activityJi)
    ? record.activityJi.filter(label => tags.includes(label) || label === '诸事不宜') : []
  return {
    date,
    lunarFull: record && record.lunarFull || '暂无农历标签',
    dayGanZhi: record && record.dayGanZhi || '—',
    jianChu: getJianChuInfo(record && record.jianChu).label,
    yi,
    ji,
    yiText: yi.length ? yi.join('、') : '无相关标签',
    jiText: ji.length ? ji.join('、') : '无相关标签',
    personalLabel: personalDay && personalDay.label || '依据不足',
    personalStatus: personalDay && personalDay.status || 'insufficient',
    personalSummary: personalDay && personalDay.summary || '暂无个人关系数据。',
  }
}

function classifyActivityDate(activityId, date, record, personalDay) {
  const activity = ACTIVITIES.find(item => item.id === activityId)
  const result = baseDay(activity, date, record, personalDay)
  const validYi = ACTIVITIES.flatMap(item => item.tags)
  if (!activity || !record || !Array.isArray(record.activityYi) || !Array.isArray(record.activityJi) ||
      !record.activityYi.every(label => validYi.includes(label)) ||
      !record.activityJi.every(label => validYi.includes(label) || label === '诸事不宜')) {
    return { ...result, group: 'insufficient', reason: `日期或${activity ? activity.label : '事项'}宜忌数据不完整，无法形成候选判断。` }
  }
  if (result.ji.length) {
    const personalNote = result.personalStatus === 'insufficient' ? '；个人关系也缺少依据' : ''
    return { ...result, group: 'conflict', reason: `通用忌包含${result.ji.join('、')}${personalNote}，先核对${activity.label}安排。` }
  }
  if (!result.yi.length) {
    return { ...result, group: 'insufficient', reason: `通用宜忌未列出${activity.tags.join('或')}，不能只凭建除推断${activity.label}是否适宜。` }
  }
  if (!personalDay || personalDay.status === 'insufficient' ||
      !['supportive', 'mixed', 'caution', 'neutral'].includes(personalDay.status)) {
    return { ...result, group: 'insufficient', reason: `通用宜有${activity.label}相关标签，但示例命盘的个人关系依据不足。` }
  }
  if (personalDay.status === 'supportive') {
    return { ...result, group: 'priority', reason: `通用宜包含${result.yi.join('、')}，示例命盘通用日干支关系为偏合；仍需核对实际安排。` }
  }
  return { ...result, group: 'review', reason: `通用宜包含${result.yi.join('、')}，示例命盘通用关系为${result.personalLabel}，需要继续核对。` }
}

function analyzeActivityRange(activityId, startDate, endDate, calendar, chart) {
  const error = ACTIVITIES.some(item => item.id === activityId) ? validateDateRange(startDate, endDate) : '请选择有效事项。'
  const sections = groupDefinitions.map(group => ({ ...group, days: [] }))
  if (error) return { error, sections, total: 0, priorityCount: 0 }

  const end = parseDate(endDate)
  for (let timestamp = parseDate(startDate); timestamp <= end; timestamp += DAY_MS) {
    const date = dateKey(timestamp)
    const record = calendar && calendar[date]
    const personalDay = buildPersonalDay(date, record, chart)
    const day = classifyActivityDate(activityId, date, record, personalDay)
    sections.find(section => section.id === day.group).days.push(day)
  }
  return {
    error: '',
    sections,
    total: sections.reduce((sum, section) => sum + section.days.length, 0),
    priorityCount: sections[0].days.length,
  }
}

module.exports = { ACTIVITIES, classifyActivityDate, defaultDateRange, validateDateRange, analyzeActivityRange }
