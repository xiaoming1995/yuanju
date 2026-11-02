const assert = require('node:assert/strict')
const test = require('node:test')
const { ACTIVITIES, defaultDateRange, validateDateRange, classifyActivityDate, analyzeActivityRange } = require('../data/real_date_selection')

const record = (date, yi = [], ji = []) => ({ date, lunar_full: '十月初一', day_gan_zhi: '甲子', yi, ji })

test('all five activities use only their original calendar labels', () => {
  const labels = { moving: '入宅', marriage: '嫁娶', opening: '开市', travel: '出行', groundbreaking: '动土' }
  assert.equal(ACTIVITIES.length, 5)
  for (const activity of ACTIVITIES) {
    const day = classifyActivityDate(activity.id, record('2028-01-01', [labels[activity.id]], []))
    assert.equal(day.group, 'yi')
    assert.deepEqual(day.yi, [labels[activity.id]])
  }
})

test('a matching Ji or 诸事不宜 overrides Yi and both original labels remain visible', () => {
  const both = classifyActivityDate('moving', record('2028-01-01', ['入宅'], ['移徙']))
  assert.equal(both.group, 'ji')
  assert.deepEqual(both.yi, ['入宅'])
  assert.deepEqual(both.ji, ['移徙'])
  assert.equal(classifyActivityDate('marriage', record('2028-01-01', ['嫁娶'], ['诸事不宜'])).group, 'ji')
  assert.equal(classifyActivityDate('travel', record('2028-01-01', ['入宅'], [])).group, 'unlisted')
})

test('China date default and range validation no longer use the bundled years', () => {
  assert.deepEqual(defaultDateRange(Date.parse('2027-12-31T16:00:00Z')), { startDate: '2028-01-01', endDate: '2028-01-30' })
  assert.equal(validateDateRange('2028-01-01', '2028-01-30'), '')
  assert.match(validateDateRange('2028-02-30', '2028-03-01'), /有效/)
  assert.match(validateDateRange('2028-01-02', '2028-01-01'), /起日/)
  assert.match(validateDateRange('2028-01-01', '2028-04-01'), /90/)
})

test('range requires every date and complete Yi/Ji arrays', () => {
  const days = [record('2028-01-01', ['入宅']), record('2028-01-02', [], ['入宅'])]
  const result = analyzeActivityRange('moving', '2028-01-01', '2028-01-02', days)
  assert.equal(result.total, 2)
  assert.equal(result.candidateCount, 1)
  assert.deepEqual(result.sections.map(section => section.days.length), [1, 1, 0])
  assert.throws(() => analyzeActivityRange('moving', '2028-01-01', '2028-01-02', days.slice(0, 1)), /不完整/)
  assert.throws(() => analyzeActivityRange('moving', '2028-01-01', '2028-01-02', [days[1], days[0]]), /不完整/)
  assert.throws(() => analyzeActivityRange('moving', '2028-01-01', '2028-01-02', [days[0], { date: '2028-01-02', yi: [] }]), /不完整/)
})
