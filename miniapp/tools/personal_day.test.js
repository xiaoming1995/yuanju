const assert = require('node:assert/strict')
const test = require('node:test')

const { buildPersonalDay } = require('../data/personal_day')

const chart = {
  personalDay: {
    dayBranch: '寅',
    fuyiYongshen: '木火',
    fuyiJishen: '水金土',
  },
}

test('favorable stem and branch produce a dated explanation and planning reminders', () => {
  const result = buildPersonalDay('2026-10-22', { dayGanZhi: '甲午' }, chart)

  assert.equal(result.date, '2026-10-22')
  assert.equal(result.dayGanZhi, '甲午')
  assert.equal(result.status, 'supportive')
  assert.equal(result.label, '偏合')
  assert.ok(result.evidence.some(item => item.includes('甲') && item.includes('扶抑喜用')))
  assert.ok(result.evidence.some(item => item.includes('午') && item.includes('主气')))
  assert.deepEqual(result.activities.map(item => item.id), ['moving', 'finance'])
  assert.equal(result.activities[0].title, '搬家')
  assert.equal(result.activities[1].title, '财务安排')
  assert.doesNotMatch(JSON.stringify(result.activities), /买入|卖出|保证收益|宜搬家/)
})

test('a day branch clash is visible and changes the relationship status', () => {
  const result = buildPersonalDay('2026-10-23', { dayGanZhi: '甲申' }, chart)

  assert.equal(result.status, 'caution')
  assert.ok(result.evidence.some(item => item.includes('申') && item.includes('寅') && item.includes('相冲')))
  assert.match(result.activities[0].detail, /时间|安排/)
})

test('mixed directions do not become a favorable activity verdict', () => {
  const result = buildPersonalDay('2026-10-24', { dayGanZhi: '壬午' }, chart)

  assert.equal(result.status, 'mixed')
  assert.equal(result.label, '关系复杂')
  assert.doesNotMatch(JSON.stringify(result.activities), /宜投资|适合投资/)
})

test('missing calendar or complete fuyi direction yields no suitability conclusion', () => {
  const missingDate = buildPersonalDay('2028-01-01', {}, chart)
  const missingChart = buildPersonalDay('2026-10-22', { dayGanZhi: '甲午' }, {
    personalDay: { dayBranch: '寅', fuyiYongshen: '木火' },
  })
  const contradictoryChart = buildPersonalDay('2026-10-22', { dayGanZhi: '甲午' }, {
    personalDay: { dayBranch: '寅', fuyiYongshen: '木火', fuyiJishen: '木水' },
  })

  for (const result of [missingDate, missingChart, contradictoryChart]) {
    assert.equal(result.status, 'insufficient')
    assert.equal(result.label, '依据不足')
    assert.deepEqual(result.activities, [])
  }
})
