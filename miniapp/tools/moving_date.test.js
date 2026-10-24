const assert = require('node:assert/strict')
const test = require('node:test')

const { chart } = require('../data/demo')
const calendar = require('../data/calendar')
const { ACTIVITIES, classifyActivityDate, defaultDateRange, analyzeActivityRange } = require('../data/moving_date')

const date = '2026-10-23'
const record = { lunarFull: '九月十四', dayGanZhi: '庚午', jianChu: '成', activityYi: ['入宅'], activityJi: [] }

test('five activities map to their own original Yi and Ji labels', () => {
  assert.deepEqual(ACTIVITIES.map(item => [item.id, item.tags]), [
    ['moving', ['入宅', '移徙']], ['marriage', ['嫁娶']], ['opening', ['开市']],
    ['travel', ['出行']], ['groundbreaking', ['动土']],
  ])
  for (const activity of ACTIVITIES) {
    const matched = classifyActivityDate(activity.id, date, { ...record, activityYi: [activity.tags[0]] }, { status: 'supportive', label: '偏合' })
    assert.equal(matched.group, 'priority', activity.id)
    assert.deepEqual(matched.yi, [activity.tags[0]])
  }
})

test('related Ji overrides Yi; unrelated activity labels do not affect the selected activity', () => {
  const mixed = classifyActivityDate('moving', date, { ...record, activityJi: ['移徙', '嫁娶'] }, { status: 'supportive', label: '偏合' })
  assert.equal(mixed.group, 'conflict')
  assert.deepEqual(mixed.ji, ['移徙'])
  const unrelated = classifyActivityDate('marriage', date, record, { status: 'supportive', label: '偏合' })
  assert.equal(unrelated.group, 'insufficient')
  assert.deepEqual(unrelated.yi, [])
})

test('the raw Ji label 诸事不宜 conflicts with every activity', () => {
  for (const activity of ACTIVITIES) {
    const result = classifyActivityDate(activity.id, date, { ...record, activityYi: [activity.tags[0]], activityJi: ['诸事不宜'] }, { status: 'supportive', label: '偏合' })
    assert.equal(result.group, 'conflict', activity.id)
    assert.deepEqual(result.ji, ['诸事不宜'])
  }
})

test('mixed and caution personal relations need review; missing evidence is insufficient', () => {
  for (const status of ['mixed', 'caution']) {
    assert.equal(classifyActivityDate('moving', date, record, { status, label: '需核对' }).group, 'review')
  }
  assert.equal(classifyActivityDate('moving', date, undefined, { status: 'supportive' }).group, 'insufficient')
  assert.equal(classifyActivityDate('moving', date, { ...record, activityJi: undefined }, { status: 'supportive' }).group, 'insufficient')
  assert.equal(classifyActivityDate('moving', date, { ...record, activityYi: [] }, { status: 'supportive' }).group, 'insufficient')
  assert.equal(classifyActivityDate('moving', date, record, { status: 'insufficient', label: '依据不足' }).group, 'insufficient')
})

test('default range and range validation remain bounded', () => {
  assert.deepEqual(defaultDateRange(new Date(2026, 9, 23)), { startDate: '2026-10-23', endDate: '2026-11-21', error: '' })
  assert.deepEqual(defaultDateRange(new Date(2027, 11, 20)), { startDate: '2027-12-20', endDate: '2027-12-31', error: '' })
  assert.match(defaultDateRange(new Date(2028, 0, 1)).error, /2025—2027/)
  assert.match(analyzeActivityRange('moving', '2026-10-24', '2026-10-23', {}, chart).error, /起日/)
  assert.match(analyzeActivityRange('moving', '2026-01-01', '2026-04-01', {}, chart).error, /90/)
  assert.match(analyzeActivityRange('moving', '2024-12-31', '2025-01-01', {}, chart).error, /2025—2027/)
  assert.match(analyzeActivityRange('moving', '2026-02-30', '2026-03-01', {}, chart).error, /日期/)
  assert.match(analyzeActivityRange('unknown', date, date, {}, chart).error, /事项/)
})

test('range keeps every date visible and filters by selected activity', () => {
  const data = {
    '2026-10-22': { ...record, activityYi: ['嫁娶'], activityJi: ['入宅'] },
    '2026-10-23': { ...record, activityYi: [], activityJi: [] },
  }
  const moving = analyzeActivityRange('moving', '2026-10-22', '2026-10-23', data, chart)
  const marriage = analyzeActivityRange('marriage', '2026-10-22', '2026-10-23', data, chart)
  assert.equal(moving.total, 2)
  assert.equal(marriage.total, 2)
  assert.equal(moving.sections.find(section => section.id === 'conflict').days[0].date, '2026-10-22')
  assert.notEqual(marriage.sections.find(section => section.id === 'conflict').days[0]?.date, '2026-10-22')
  assert.deepEqual(marriage.sections.flatMap(section => section.days).find(day => day.date === '2026-10-22').yi, ['嫁娶'])
})

test('real preview data still offers an explained moving candidate', () => {
  const result = analyzeActivityRange('moving', '2026-11-16', '2026-11-16', calendar, chart)
  assert.equal(result.priorityCount, 1)
  assert.deepEqual(result.sections[0].days[0].yi, ['移徙', '入宅'])
  assert.equal(result.sections[0].days[0].personalLabel, '偏合')
})
