const assert = require('node:assert/strict')
const test = require('node:test')

const { basicRows } = require('../data/demo')

test('basic chart rows follow the web result order and preserve four pillar values', () => {
  assert.ok(basicRows, 'basic chart rows are available')
  assert.deepEqual(basicRows.map(row => row.label), [
    '主星', '天干', '地支', '藏干', '副星', '星运', '自坐', '空亡', '纳音', '神煞',
  ])
  assert.ok(basicRows.every(row => row.cells.length === 4))
  assert.deepEqual(basicRows[0].cells.map(cell => cell.text), ['伤官', '比肩', '日主', '伤官'])
  assert.deepEqual(basicRows[3].cells[0].items, ['丙', '庚', '戊'])
  assert.deepEqual(basicRows[4].cells[0].items, ['比肩', '偏财', '食神'])
})
