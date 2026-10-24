const assert = require('node:assert/strict')
const test = require('node:test')

const calendar = require('../data/calendar')
const { getJianChuInfo, jianChuNames } = require('../data/jianchu')

test('every preview calendar date has one of the twelve values from lunar-go', () => {
  const days = Object.values(calendar)
  assert.equal(days.length, 1095)
  assert.deepEqual(jianChuNames, ['建', '除', '满', '平', '定', '执', '破', '危', '成', '收', '开', '闭'])
  assert.ok(days.every(day => jianChuNames.includes(day.jianChu)))
  assert.deepEqual(new Set(days.map(day => day.jianChu)), new Set(jianChuNames))
  assert.equal(calendar['2026-10-22'].jianChu, '危')
  assert.equal(calendar['2026-10-23'].jianChu, '成')
})

test('each value has readable copy and unknown input has no invented meaning', () => {
  for (const name of jianChuNames) {
    const info = getJianChuInfo(name)
    assert.equal(info.label, `${name}日`)
    assert.ok(info.meaning.length >= 4)
  }
  assert.equal(getJianChuInfo('未知').label, '暂无建除数据')
  assert.equal(getJianChuInfo('').meaning, '')
})
