import { describe, expect, it } from 'vitest'
import { gongJiaSourceDisplay } from './gongJiaDisplay'

describe('gongJiaSourceDisplay', () => {
  it('将后端夹拱来源枚举显示为中文标签', () => {
    expect(gongJiaSourceDisplay({ source: 'year_month' })).toBe('年月夹拱')
    expect(gongJiaSourceDisplay({ source: 'month_day' })).toBe('月日夹拱')
    expect(gongJiaSourceDisplay({ source: 'day_hour' })).toBe('日时夹拱')
  })

  it('未知来源优先使用 source_labels 兜底，避免露出枚举值', () => {
    expect(gongJiaSourceDisplay({
      source: 'unknown_pair',
      source_labels: ['月柱', '日柱'],
    })).toBe('月日夹拱')
  })

  it('缺失来源信息时使用保守文案', () => {
    expect(gongJiaSourceDisplay({ source: 'unknown_pair' })).toBe('夹拱')
  })
})
