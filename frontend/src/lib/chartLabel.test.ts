import { describe, expect, it } from 'vitest'
import type { BaziHistoryChart } from './api'
import { chartDisplayName, chartFallbackName, formatBirth, formatPillars, genderText } from './chartLabel'

const chart: BaziHistoryChart = {
  id: 'c1',
  birth_year: 1990,
  birth_month: 6,
  birth_day: 15,
  birth_hour: 10,
  gender: 'male',
  year_gan: '庚', year_zhi: '午',
  month_gan: '壬', month_zhi: '午',
  day_gan: '甲', day_zhi: '子',
  hour_gan: '己', hour_zhi: '巳',
  created_at: '2026-01-01T00:00:00Z',
}

describe('genderText', () => {
  it('female 显示女命，其余显示男命', () => {
    expect(genderText('female')).toBe('女命')
    expect(genderText('male')).toBe('男命')
    expect(genderText('')).toBe('男命')
  })
})

describe('chartFallbackName', () => {
  it('拼出性别 + 出生年月日', () => {
    expect(chartFallbackName(chart)).toBe('男命 · 1990年6月15日')
  })
})

describe('chartDisplayName', () => {
  it('有称呼时优先用称呼', () => {
    expect(chartDisplayName({ ...chart, display_name: '阿明' })).toBe('阿明')
  })
  it('称呼为空白或缺失时回退到默认名', () => {
    expect(chartDisplayName({ ...chart, display_name: '   ' })).toBe('男命 · 1990年6月15日')
    expect(chartDisplayName(chart)).toBe('男命 · 1990年6月15日')
  })
})

describe('formatPillars', () => {
  it('四柱以 · 分隔', () => {
    expect(formatPillars(chart)).toBe('庚午 · 壬午 · 甲子 · 己巳')
  })
})

describe('formatBirth', () => {
  it('拼出年月日时', () => {
    expect(formatBirth(chart)).toBe('1990年6月15日 10时')
  })
})
