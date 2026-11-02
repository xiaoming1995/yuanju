const { ACTIVITIES: originalActivities } = require('./moving_date')

const DAY_MS = 24 * 60 * 60 * 1000
const MIN_DATE = '1900-01-01'
const MAX_DATE = '2100-12-31'
const notes = {
  moving: '按通用宜忌中的「入宅」「移徙」筛选；具体搬迁类型仍需核对。',
  marriage: '按通用宜忌中的「嫁娶」筛选；不包含订婚。',
  opening: '按通用宜忌中的「开市」筛选；不代表交易或投资时机。',
  travel: '按通用宜忌中的「出行」筛选；重要行程与安全安排仍需核对。',
  groundbreaking: '按通用宜忌中的「动土」筛选；不包含其他施工事项。',
}
const ACTIVITIES = originalActivities.map(item => ({ ...item, note: notes[item.id] }))
const GROUPS = [
  { id: 'yi', label: '有相关宜' },
  { id: 'ji', label: '有相关忌' },
  { id: 'unlisted', label: '未列相关标签' },
]

function dateKey(timestamp) { return new Date(timestamp).toISOString().slice(0, 10) }

function parseDate(value) {
  if (typeof value !== 'string' || !/^\d{4}-\d{2}-\d{2}$/.test(value)) return null
  const [year, month, day] = value.split('-').map(Number)
  const timestamp = Date.UTC(year, month - 1, day)
  return dateKey(timestamp) === value ? timestamp : null
}

function chinaTodayKey(now = Date.now()) { return dateKey(now + 8 * 60 * 60 * 1000) }

function defaultDateRange(now = Date.now()) {
  const today = chinaTodayKey(now)
  if (today < MIN_DATE || today > MAX_DATE) return { startDate: '', endDate: '' }
  return { startDate: today, endDate: dateKey(Math.min(parseDate(today) + 29 * DAY_MS, parseDate(MAX_DATE))) }
}

function validateDateRange(startDate, endDate) {
  const start = parseDate(startDate)
  const end = parseDate(endDate)
  if (start === null || end === null || startDate < MIN_DATE || startDate > MAX_DATE ||
    endDate < MIN_DATE || endDate > MAX_DATE) return '请选择1900—2100年内的有效起止日期。'
  if (start > end) return '起日不能晚于止日。'
  if ((end - start) / DAY_MS >= 90) return '一次最多查看90个日历日。'
  return ''
}

function classifyActivityDate(activityId, record) {
  const activity = ACTIVITIES.find(item => item.id === activityId)
  if (!activity || !record || !Array.isArray(record.yi) || !Array.isArray(record.ji)) throw new Error('日期宜忌资料不完整')
  const yi = record.yi.filter(label => activity.tags.includes(label))
  const ji = record.ji.filter(label => activity.tags.includes(label) || label === '诸事不宜')
  const group = ji.length ? 'ji' : yi.length ? 'yi' : 'unlisted'
  const reason = group === 'ji' ? `通用忌包含${ji.join('、')}，请核对${activity.label}安排。`
    : group === 'yi' ? `通用宜包含${yi.join('、')}；仍需核对时辰、参与人和实际安排。`
      : `通用宜忌未列${activity.tags.join('或')}，不据此推断${activity.label}是否适宜。`
  return {
    date: record.date, lunarFull: record.lunar_full || '农历资料暂缺', dayGanZhi: record.day_gan_zhi || '—',
    yi, ji, yiText: yi.length ? yi.join('、') : '无相关标签',
    jiText: ji.length ? ji.join('、') : '无相关标签',
    allYi: record.yi, allJi: record.ji, group, reason,
  }
}

function analyzeActivityRange(activityId, startDate, endDate, days) {
  const error = validateDateRange(startDate, endDate)
  if (error) throw new Error(error)
  if (!ACTIVITIES.some(item => item.id === activityId)) throw new Error('请选择有效事项。')
  const expected = (parseDate(endDate) - parseDate(startDate)) / DAY_MS + 1
  if (!Array.isArray(days) || days.length !== expected) throw new Error('日期资料不完整，请重试。')
  const sections = GROUPS.map(group => ({ ...group, days: [] }))
  days.forEach((record, index) => {
    if (!record || record.date !== dateKey(parseDate(startDate) + index * DAY_MS) ||
      !record.lunar_full || !record.day_gan_zhi ||
      !Array.isArray(record.yi) || !Array.isArray(record.ji) ||
      !record.yi.every(label => typeof label === 'string') ||
      !record.ji.every(label => typeof label === 'string')) throw new Error('日期资料不完整，请重试。')
    const day = classifyActivityDate(activityId, record)
    sections.find(section => section.id === day.group).days.push(day)
  })
  return { sections, total: days.length, candidateCount: sections[0].days.length }
}

module.exports = { ACTIVITIES, GROUPS, chinaTodayKey, defaultDateRange, validateDateRange, classifyActivityDate, analyzeActivityRange }
