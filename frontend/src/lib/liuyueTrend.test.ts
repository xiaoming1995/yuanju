import { describe, expect, it } from 'vitest'
import type { LiuYueItem } from './api'
import { buildLiuYueTrendPath, buildLiuYueTrendSeries, liuyueTrendX } from './liuyueTrend'

function item(index: number, patch: Partial<LiuYueItem>): LiuYueItem {
  return {
    index,
    month_name: `${index + 1}月`,
    gan_zhi: '甲子',
    gan_shishen: '正印',
    zhi_shishen: '正官',
    jie_qi_name: '立春',
    start_date: '2026-02-04',
    end_date: '2026-03-04',
    ...patch,
  }
}

describe('buildLiuYueTrendSeries', () => {
  it('空流月返回空趋势', () => {
    expect(buildLiuYueTrendSeries([])).toEqual([])
  })

  it('按十神生成五个趋势维度', () => {
    const series = buildLiuYueTrendSeries([
      item(0, { gan_shishen: '正官', zhi_shishen: '正印' }),
      item(1, { gan_shishen: '劫财', zhi_shishen: '伤官' }),
    ])

    expect(series.map(s => s.key)).toEqual(['overall', 'career', 'marriage', 'wealth', 'health'])
    expect(series[0].points).toHaveLength(2)
    expect(series[0].points[0].startDate).toBe('2026-02-04')
  })

  it('男命婚恋以财星为顺势线索', () => {
    const marriage = buildLiuYueTrendSeries([
      item(0, { gan_shishen: '正财', zhi_shishen: '偏财' }),
    ], 'male').find(series => series.key === 'marriage')

    expect(marriage?.points[0].level).toBe(3)
  })

  it('女命婚恋以官杀为顺势线索', () => {
    const marriage = buildLiuYueTrendSeries([
      item(0, { gan_shishen: '正官', zhi_shishen: '七杀' }),
    ], 'female').find(series => series.key === 'marriage')

    expect(marriage?.points[0].level).toBe(3)
  })
})

describe('liuyue trend geometry', () => {
  it('单点时返回稳定 x 坐标', () => {
    expect(liuyueTrendX(0, 1)).toBe(48)
  })

  it('生成 SVG path', () => {
    const series = buildLiuYueTrendSeries([
      item(0, { gan_shishen: '正官' }),
      item(1, { gan_shishen: '劫财', zhi_shishen: '伤官' }),
    ])

    expect(buildLiuYueTrendPath(series[0].points)).toContain('M 48.0')
    expect(buildLiuYueTrendPath(series[0].points)).toContain('L 360.0')
  })
})
