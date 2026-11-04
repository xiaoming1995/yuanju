const assert = require('node:assert/strict')
const test = require('node:test')
const { toResult, toHistory } = require('../utils/chart_view')

test('result view renders the returned birth chart rather than the fixed preview', () => {
  const result = {
    year_gan: '甲', year_zhi: '子', month_gan: '乙', month_zhi: '丑',
    day_gan: '丙', day_zhi: '寅', hour_gan: '丁', hour_zhi: '卯',
    year_gan_wuxing: '木', wuxing: { mu: 3, huo: 2, tu: 1, jin: 1, shui: 1, total: 8 },
    ming_ge: '测试命格', natal_assessment: { fuyi: { yongshen: '木', jishen: '金' } },
  }
  const view = toResult(result, { dateText: '2001-02-03 04:00', gender: 'female' })
  assert.equal(view.input.dateTime, '2001-02-03 04:00')
  assert.equal(view.input.gender, 'female')
  assert.equal(view.pillars[0].gan, '甲')
  assert.equal(view.pillars[0].ganTone, 'wood')
  assert.equal(view.pillars[3].zhi, '卯')
  assert.equal(view.basicRows[1].cells[0].text, '甲')
  assert.equal(view.wuxingRows[0].count, 3)
  assert.equal(view.personalDay.fuyiYongshen, '木')
})

test('history view keeps server chart ID and entered date', () => {
  const view = toHistory({ id: 'server-id', display_name: '孩子', birth_year: 2018,
    birth_month: 5, birth_day: 12, birth_hour: 10,
    year_gan: '戊', year_zhi: '戌', month_gan: '丁', month_zhi: '巳',
    day_gan: '壬', day_zhi: '辰', hour_gan: '乙', hour_zhi: '巳' })
  assert.equal(view.id, 'server-id')
  assert.equal(view.name, '孩子')
  assert.equal(view.date, '2018-05-12 10:00')
  assert.equal(view.columns[0].gan, '戊')
  assert.equal(view.columns[0].zhi, '戌')
  assert.equal(view.columns[0].text, '戊戌')
  assert.deepEqual(view.columns.map(column => column.key), ['year', 'month', 'day', 'hour'])
})
