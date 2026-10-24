const assert = require('node:assert/strict')
const test = require('node:test')

const { chart } = require('../data/demo')

test('the preview chart is one complete backend-derived example', () => {
  assert.ok(chart, 'a fixed chart example is available')
  assert.equal(chart.input.dateTime, '1990-01-01 01:00')
  assert.equal(chart.input.gender, 'male')
  assert.deepEqual(chart.pillars.map(pillar => pillar.gan + pillar.zhi), [
    '己巳', '丙子', '丙寅', '己丑',
  ])

  for (const pillar of chart.pillars) {
    for (const field of ['ganWuxing', 'zhiWuxing', 'ganShiShen', 'xingYun', 'diShi', 'xunKong', 'naYin']) {
      assert.equal(typeof pillar[field], 'string', `${pillar.label}: ${field}`)
      assert.ok(pillar[field], `${pillar.label}: ${field} has a value`)
    }
    for (const field of ['hideGan', 'zhiShiShen', 'shenSha']) {
      assert.ok(Array.isArray(pillar[field]), `${pillar.label}: ${field} is a list`)
    }
    assert.equal(pillar.hideGan.length, pillar.zhiShiShen.length)
  }

  assert.deepEqual(chart.wuxingRows.map(row => row.label), ['木', '火', '土', '金', '水'])
  assert.equal(chart.wuxingRows.reduce((total, row) => total + row.count, 0), chart.wuxingTotal)
  assert.equal(chart.personalDay.dayBranch, chart.pillars[2].zhi)
  assert.ok(chart.personalDay.fuyiYongshen)
  assert.ok(chart.personalDay.fuyiJishen)
  assert.notEqual(chart.personalDay.fuyiYongshen, chart.personalDay.fuyiJishen)
  assert.ok(chart.mingGe, 'the example includes its engine-detected pattern')
  assert.ok(chart.mingGeDesc, 'the example includes the pattern explanation')
  assert.ok(chart.stemGuidance, 'the example includes the web overview stem guidance')
  const guidance = chart.stemGuidance
  for (const field of ['primary_favorable', 'secondary_favorable', 'adverse', 'conditioning_only']) {
    assert.ok(Array.isArray(guidance[field]), `${field} is an engine-generated list`)
  }
  assert.ok(
    guidance.primary_favorable.length + guidance.secondary_favorable.length + guidance.adverse.length + guidance.conditioning_only.length > 0,
    'the example has a visible stem-level conclusion',
  )
})
