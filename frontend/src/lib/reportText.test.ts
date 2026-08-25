import { describe, expect, it } from 'vitest'
import { cleanLegacyReportContent, cleanReportText, hasMeaningfulReportContent, splitParagraphs } from './reportText'

describe('cleanReportText', () => {
  it('空值返回空串', () => {
    expect(cleanReportText(undefined)).toBe('')
    expect(cleanReportText(null)).toBe('')
    expect(cleanReportText('')).toBe('')
  })

  it('去掉 ** 与 __ 加粗符号', () => {
    expect(cleanReportText('**日主**偏强，__喜__金水')).toBe('日主偏强，喜金水')
  })

  it('去掉贴着文字的单 * 斜体符号', () => {
    expect(cleanReportText('*财星*透出')).toBe('财星透出')
  })

  it('不误删两侧带空格的数学星号', () => {
    expect(cleanReportText('2 * 3 = 6')).toBe('2 * 3 = 6')
  })

  it('去掉首尾空白', () => {
    expect(cleanReportText('  正文  ')).toBe('正文')
  })
})

describe('splitParagraphs', () => {
  it('按连续空行拆段并过滤空段', () => {
    expect(splitParagraphs('第一段\n\n第二段\n\n\n第三段')).toEqual(['第一段', '第二段', '第三段'])
  })

  it('单段不拆', () => {
    expect(splitParagraphs('只有一段\n仍是同一段')).toEqual(['只有一段\n仍是同一段'])
  })

  it('空值返回空数组', () => {
    expect(splitParagraphs(undefined)).toEqual([])
    expect(splitParagraphs('')).toEqual([])
  })
})

describe('cleanLegacyReportContent', () => {
  it('过滤旧报告里的分隔线与免责声明', () => {
    expect(cleanLegacyReportContent('正文\n\n---\n*本报告由 AI 辅助生成，内容仅供参考，不构成任何决策建议。*')).toBe('正文')
  })

  it('只剩分隔线和免责声明时视为无正文', () => {
    const content = '---\n*本报告由 AI 辅助生成，内容仅供参考，不构成任何决策建议。*'

    expect(cleanLegacyReportContent(content)).toBe('')
    expect(hasMeaningfulReportContent(content)).toBe(false)
  })

  it('只剩章节标题时视为无正文', () => {
    expect(hasMeaningfulReportContent('【命局概要】\n\n---')).toBe(false)
  })
})
