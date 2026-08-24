import { describe, expect, it } from 'vitest'
import {
  countPastEventsProgress,
  buildPastEventsExportSegments,
  filterPastEventsExportSegments,
  getPastEventsCurrentPeriod,
  getPastEventsFutureToggleLabel,
  getPastEventsYearNarrativeState,
  getPastEventsSegmentViewState,
  getPastEventsSignalLabel,
} from './pastEventsViewModel'

describe('getPastEventsSegmentViewState', () => {
  it('将生成中段标记为批语生成中且不暴露 AI 字眼', () => {
    const state = getPastEventsSegmentViewState({ loading: true, summary: '' })

    expect(state.state).toBe('generating')
    expect(state.label).toBe('批语生成中')
    expect(state.description).not.toContain('AI')
    expect(state.canGenerateAi).toBe(false)
  })

  it('将中断段标记为可重试', () => {
    const state = getPastEventsSegmentViewState({ status: 'interrupted', error: '等待响应超时，点击重试' })

    expect(state.state).toBe('interrupted')
    expect(state.canRetry).toBe(true)
    expect(state.description).toContain('等待响应超时')
  })

  it('将已有批语内容段标记为完成', () => {
    const state = getPastEventsSegmentViewState({ summary: '本步大运已有总结。', themes: ['事业'] })

    expect(state.state).toBe('generated')
    expect(state.label).toBe('批语已完成')
    expect(state.description).not.toContain('AI')
    expect(state.isComplete).toBe(true)
    expect(state.canRetry).toBe(false)
  })

  it('区分当前命盘缓存和刚生成内容，避免误认为复用旧命盘', () => {
    const cachedState = getPastEventsSegmentViewState({
      summary: '本步大运已有总结。',
      themes: ['事业'],
      cached: true,
    })
    const generatedState = getPastEventsSegmentViewState({
      summary: '本步大运已有总结。',
      themes: ['事业'],
      cached: false,
    })

    expect(cachedState.label).toBe('本次缓存')
    expect(cachedState.description).toContain('当前命盘')
    expect(cachedState.description).toContain('不复用同八字旧缓存')
    expect(generatedState.label).toBe('批语刚生成')
    expect(generatedState.description).toContain('当前命盘')
  })

  it('将未来折叠段标记为只展开年份信号', () => {
    const state = getPastEventsSegmentViewState({ folded: true, summary: '', themes: [] })

    expect(state.state).toBe('future_folded')
    expect(state.canRevealSignals).toBe(true)
    expect(state.canGenerateAi).toBe(false)
  })

  it('将未来已展开但未生成段标记为可生成批语', () => {
    const state = getPastEventsSegmentViewState({ folded: false, summary: '', themes: [] })

    expect(state.state).toBe('future_expanded_ungenerated')
    expect(state.description).not.toContain('AI')
    expect(state.canRevealSignals).toBe(false)
    expect(state.canGenerateAi).toBe(true)
  })

  it('用户可见状态文案不包含 AI', () => {
    const states = [
      getPastEventsSegmentViewState(undefined),
      getPastEventsSegmentViewState({ loading: true }),
      getPastEventsSegmentViewState({ status: 'interrupted' }),
      getPastEventsSegmentViewState({ summary: '已有总结' }),
      getPastEventsSegmentViewState({ folded: true }),
      getPastEventsSegmentViewState({ folded: false, summary: '', themes: [] }),
    ]

    for (const state of states) {
      expect(state.label).not.toContain('AI')
      expect(state.description).not.toContain('AI')
    }
  })
})

describe('buildPastEventsExportSegments', () => {
  const meta = [
    { index: 1, gan_zhi: '甲子', start_age: 0, end_age: 9, start_year: 1990, end_year: 1999 },
    { index: 2, gan_zhi: '乙丑', start_age: 10, end_age: 19, start_year: 2000, end_year: 2009 },
    { index: 3, gan_zhi: '丙寅', start_age: 20, end_age: 29, start_year: 2010, end_year: 2019 },
    { index: 4, gan_zhi: '丁卯', start_age: 30, end_age: 39, start_year: 2020, end_year: 2029 },
  ]
  const years = [
    { dayun_index: 1, year: 1990, age: 1, gan_zhi: '庚午', signals: ['学业_资源'], evidence_summary: ['证据1'], narrative: '算法正文' },
    { dayun_index: 1, year: 1991, age: 2, gan_zhi: '辛未', signals: ['学业_竞争'], evidence_summary: ['证据2'], narrative: '算法正文' },
    { dayun_index: 2, year: 2000, age: 10, gan_zhi: '庚辰', signals: ['事业'], evidence_summary: ['证据3'], narrative: '算法正文' },
  ]

  it('只导出已完成 summary 且有生成逐年正文的大运段', () => {
    const segments = buildPastEventsExportSegments(meta, years, {
      1: {
        summary: '第一步已生成。',
        themes: ['学业突破'],
        years: [
          { year: 1990, ganzhi: '庚午', narrative: '1990 生成正文。' },
          { year: 1991, ganzhi: '辛未', narrative: '' },
        ],
      },
      2: { loading: true, summary: '生成中不导出', years: [{ year: 2000, narrative: '不应导出' }] },
      3: { status: 'interrupted', summary: '中断不导出', years: [{ year: 2010, narrative: '不应导出' }] },
      4: { folded: true, summary: '', years: [] },
    })

    expect(segments).toHaveLength(1)
    expect(segments[0]).toMatchObject({
      dayun_index: 1,
      gan_zhi: '甲子',
      start_age: 0,
      end_age: 9,
      summary: '第一步已生成。',
    })
    expect(segments[0].years).toHaveLength(1)
    expect(segments[0].years[0]).toMatchObject({
      year: 1990,
      age: 1,
      gan_zhi: '庚午',
      narrative: '1990 生成正文。',
    })
  })

  it('默认不包含命理依据，开启后才带出信号和依据', () => {
    const summaries = {
      1: {
        summary: '第一步已生成。',
        themes: ['学业突破'],
        years: [{ year: 1990, ganzhi: '庚午', narrative: '1990 生成正文。' }],
      },
    }

    const concise = buildPastEventsExportSegments(meta, years, summaries)
    const full = buildPastEventsExportSegments(meta, years, summaries, true)

    expect(concise[0].years[0].signals).toBeUndefined()
    expect(concise[0].years[0].evidence_summary).toBeUndefined()
    expect(full[0].years[0].signals).toEqual(['学业_资源'])
    expect(full[0].years[0].evidence_summary).toEqual(['证据1'])
  })
})

describe('filterPastEventsExportSegments', () => {
  it('过滤空 summary、空年份正文，并按选项移除依据', () => {
    const segments = filterPastEventsExportSegments([
      {
        dayun_index: 1,
        gan_zhi: '甲子',
        start_year: 1990,
        end_year: 1999,
        themes: ['学业'],
        summary: ' 已缓存 ',
        years: [
          { year: 1990, age: 1, gan_zhi: '庚午', narrative: ' 生成正文 ', signals: ['学业_资源'], evidence_summary: ['证据'] },
          { year: 1991, age: 2, gan_zhi: '辛未', narrative: '' },
        ],
      },
      {
        dayun_index: 2,
        gan_zhi: '乙丑',
        summary: '',
        themes: [],
        years: [{ year: 2000, gan_zhi: '庚辰', narrative: '不应导出' }],
      },
    ])

    expect(segments).toHaveLength(1)
    expect(segments[0].summary).toBe('已缓存')
    expect(segments[0].years).toHaveLength(1)
    expect(segments[0].years[0].narrative).toBe('生成正文')
    expect(segments[0].years[0].signals).toBeUndefined()
  })
})

describe('countPastEventsProgress', () => {
  it('统计已生成、生成中、中断和未来待生成段', () => {
    expect(countPastEventsProgress({
      1: { summary: '已有总结', themes: ['事业'] },
      2: { loading: true },
      3: { status: 'interrupted', error: '失败' },
      4: { folded: true, summary: '' },
      5: { folded: false, summary: '', themes: [] },
    })).toEqual({
      generated: 1,
      generating: 1,
      interrupted: 1,
      futurePending: 2,
    })
  })
})

describe('getPastEventsSignalLabel', () => {
  it('显示夹拱信号标签', () => {
    expect(getPastEventsSignalLabel('夹拱')).toMatchObject({
      label: '夹拱',
    })
  })

  it('未知信号使用保守兜底标签而不是丢弃', () => {
    expect(getPastEventsSignalLabel('新的有效信号')).toMatchObject({
      label: '新的有效信号',
    })
  })
})

describe('getPastEventsCurrentPeriod', () => {
  it('找到当前大运和当前年份', () => {
    const current = getPastEventsCurrentPeriod(
      [
        { index: 1, gan_zhi: '甲子', start_year: 2000, end_year: 2009 },
        { index: 2, gan_zhi: '乙丑', start_year: 2010, end_year: 2019 },
      ],
      [
        { year: 2017, age: 27 },
        { year: 2020, age: 30 },
      ],
      2017,
    )

    expect(current.dayun?.index).toBe(2)
    expect(current.year?.age).toBe(27)
  })
})

describe('getPastEventsFutureToggleLabel', () => {
  it('用明确文案区分展开和收起年份信号', () => {
    expect(getPastEventsFutureToggleLabel(true)).toBe('展开年份信号')
    expect(getPastEventsFutureToggleLabel(false)).toBe('收起年份信号')
  })
})

describe('getPastEventsYearNarrativeState', () => {
  it('批语生成中时不提前展示年份算法正文', () => {
    const state = getPastEventsYearNarrativeState(
      {
        year: 2024,
        dayun_index: 1,
        narrative: '从命理依据看，这类冲动常落在亲密关系、居住状态和合作关系上。',
      },
      { status: 'loading', loading: true },
    )

    expect(state).toEqual({
      status: 'loading',
      text: '',
    })
  })

  it('生成中断时不使用年份算法正文兜底', () => {
    const state = getPastEventsYearNarrativeState(
      {
        year: 2024,
        dayun_index: 1,
        narrative: '从命理依据看，这类冲动常落在亲密关系、居住状态和合作关系上。',
      },
      { status: 'interrupted', error: '等待响应超时' },
    )

    expect(state).toEqual({
      status: 'empty',
      text: '',
    })
  })

  it('批语生成成功后优先使用生成的逐年正文', () => {
    const state = getPastEventsYearNarrativeState(
      {
        year: 2024,
        dayun_index: 1,
        narrative: '算法依据型兜底正文。',
      },
      {
        summary: '本段已完成。',
        years: [{ year: 2024, narrative: '生成后的逐年批语正文。' }],
      },
    )

    expect(state).toEqual({
      status: 'ready',
      text: '生成后的逐年批语正文。',
    })
  })

  it('批语完成但缺少对应逐年正文时才使用年份算法正文兜底', () => {
    const state = getPastEventsYearNarrativeState(
      {
        year: 2024,
        dayun_index: 1,
        narrative: '算法依据型兜底正文。',
      },
      {
        summary: '本段已完成。',
        years: [{ year: 2023, narrative: '其他年份正文。' }],
      },
    )

    expect(state).toEqual({
      status: 'ready',
      text: '算法依据型兜底正文。',
    })
  })
})
