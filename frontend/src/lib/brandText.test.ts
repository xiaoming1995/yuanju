import { describe, expect, it } from 'vitest'
import type { ExportBrand } from './api'
import { resolveFooter, showDiagonalWatermark } from './brandText'

const baseBrand: ExportBrand = {
  title: '缘聚命理',
  footer_text: '',
  logo_url: '',
  logo_mode: 'icon',
  watermark_mode: 'none',
  watermark_text: '',
}

describe('resolveFooter', () => {
  it('无品牌配置时返回 fallback', () => {
    expect(resolveFooter(null, 'yuanju.com')).toBe('yuanju.com')
    expect(resolveFooter(undefined, 'yuanju.com')).toBe('yuanju.com')
  })

  it('bottom 模式且两个文案都有时用 · 拼接', () => {
    const brand = { ...baseBrand, watermark_mode: 'bottom' as const, watermark_text: '微信 abc', footer_text: '张师傅' }
    expect(resolveFooter(brand, 'yuanju.com')).toBe('张师傅 · 微信 abc')
  })

  it('bottom 模式只有水印文案时只显示水印文案', () => {
    const brand = { ...baseBrand, watermark_mode: 'bottom' as const, watermark_text: '微信 abc' }
    expect(resolveFooter(brand, 'yuanju.com')).toBe('微信 abc')
  })

  it('非 bottom 模式时 footer_text 优先，为空则 fallback', () => {
    expect(resolveFooter({ ...baseBrand, footer_text: '张师傅' }, 'yuanju.com')).toBe('张师傅')
    expect(resolveFooter(baseBrand, 'yuanju.com')).toBe('yuanju.com')
  })
})

describe('showDiagonalWatermark', () => {
  it('diagonal 模式且有文案时显示', () => {
    expect(showDiagonalWatermark({ ...baseBrand, watermark_mode: 'diagonal', watermark_text: '内部资料' })).toBe(true)
  })

  it('diagonal 模式但无文案不显示', () => {
    expect(showDiagonalWatermark({ ...baseBrand, watermark_mode: 'diagonal' })).toBe(false)
  })

  it('非 diagonal 模式不显示', () => {
    expect(showDiagonalWatermark({ ...baseBrand, watermark_mode: 'bottom', watermark_text: '内部资料' })).toBe(false)
    expect(showDiagonalWatermark(null)).toBe(false)
  })
})
