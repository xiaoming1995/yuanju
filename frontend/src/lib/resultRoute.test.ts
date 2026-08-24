import { describe, expect, it } from 'vitest'
import { buildBaziResultRoute } from './resultRoute'

describe('buildBaziResultRoute', () => {
  it('登录用户带 chart id 时进入可刷新恢复的历史详情页', () => {
    expect(buildBaziResultRoute('chart-123', false)).toBe('/history/chart-123')
  })

  it('游客保留临时结果页', () => {
    expect(buildBaziResultRoute('anonymous-chart', true)).toBe('/result')
  })

  it('没有 chart id 时保守回到临时结果页', () => {
    expect(buildBaziResultRoute(undefined, false)).toBe('/result')
  })
})
