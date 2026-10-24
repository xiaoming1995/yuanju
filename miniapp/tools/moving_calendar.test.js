const assert = require('node:assert/strict')
const test = require('node:test')

const calendar = require('../data/calendar')
const labels = ['入宅', '移徙', '嫁娶', '开市', '出行', '动土']

test('every preview date carries only selected original Yi and Ji labels', () => {
  const keys = Object.keys(calendar)
  assert.equal(keys.length, 1095)
  assert.equal(keys[0], '2025-01-01')
  assert.equal(keys.at(-1), '2027-12-31')
  for (const day of Object.values(calendar)) {
    assert.ok(Array.isArray(day.activityYi))
    assert.ok(Array.isArray(day.activityJi))
    assert.ok(day.activityYi.every(label => labels.includes(label)))
    assert.ok(day.activityJi.every(label => [...labels, '诸事不宜'].includes(label)))
  }
  for (const label of labels) {
    assert.ok(Object.values(calendar).some(day => day.activityYi.includes(label) || day.activityJi.includes(label)), label)
  }
  assert.deepEqual(calendar['2026-10-22'].activityJi.filter(label => ['入宅', '移徙'].includes(label)), ['入宅'])
  assert.deepEqual(calendar['2026-10-23'].activityYi.filter(label => ['入宅', '移徙'].includes(label)), ['入宅', '移徙'])
})
