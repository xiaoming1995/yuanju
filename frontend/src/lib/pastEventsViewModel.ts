export interface PastEventsDayunSummaryLike {
  themes?: string[]
  summary?: string
  years?: PastEventsYearNarrativeEntryLike[]
  loading?: boolean
  status?: 'loading' | 'interrupted'
  error?: string
  folded?: boolean
  cached?: boolean
}

export interface PastEventsYearNarrativeEntryLike {
  year: number
  ganzhi?: string
  gan_zhi?: string
  narrative?: string
}

export type PastEventsSegmentState =
  | 'algorithm_ready'
  | 'generating'
  | 'generated'
  | 'future_folded'
  | 'future_expanded_ungenerated'
  | 'interrupted'

export interface PastEventsSegmentViewState {
  state: PastEventsSegmentState
  label: string
  description: string
  canRevealSignals: boolean
  canGenerateAi: boolean
  canRetry: boolean
  isComplete: boolean
}

export interface PastEventsProgressCounts {
  generated: number
  generating: number
  interrupted: number
  futurePending: number
}

export interface PastEventsDayunMetaLike {
  index: number
  gan_zhi?: string
  start_age?: number
  end_age?: number
  start_year: number
  end_year: number
}

export interface PastEventsYearLike {
  year: number
  age?: number
}

export interface PastEventsYearNarrativeLike extends PastEventsYearLike {
  dayun_index: number
  gan_zhi?: string
  narrative?: string
  signals?: string[]
  evidence_summary?: string[]
  year_in_dayun?: number
  dayun_phase?: string
}

export interface PastEventsExportReadyYear {
  year: number
  age?: number
  gan_zhi: string
  narrative: string
  signals?: string[]
  evidence_summary?: string[]
  year_in_dayun?: number
  dayun_phase?: string
}

export interface PastEventsExportReadySegment {
  dayun_index: number
  gan_zhi: string
  start_age?: number
  end_age?: number
  start_year?: number
  end_year?: number
  themes: string[]
  summary: string
  years: PastEventsExportReadyYear[]
}

export type PastEventsYearNarrativeState = {
  text: string
  status: 'loading' | 'ready' | 'empty'
}

export function getPastEventsCurrentPeriod<TDayun extends PastEventsDayunMetaLike, TYear extends PastEventsYearLike>(
  dayunMeta: TDayun[],
  years: TYear[],
  year: number,
): { dayun?: TDayun; year?: TYear } {
  return {
    dayun: dayunMeta.find((meta) => year >= meta.start_year && year <= meta.end_year),
    year: years.find((item) => item.year === year),
  }
}

export function getPastEventsFutureToggleLabel(folded: boolean): string {
  return folded ? '展开年份信号' : '收起年份信号'
}

export function getPastEventsYearNarrativeState(
  year: PastEventsYearNarrativeLike,
  summary?: PastEventsDayunSummaryLike,
): PastEventsYearNarrativeState {
  if (!summary || summary.loading || summary.status === 'loading') {
    return { text: '', status: 'loading' }
  }

  if (!summary.error) {
    const generated = summary.years?.find((entry) => entry.year === year.year)
    if (generated?.narrative) {
      return { text: generated.narrative, status: 'ready' }
    }

    const hasCompletedAi = Boolean(summary.summary || summary.years?.length || summary.themes?.length)
    if (hasCompletedAi && year.narrative) {
      return { text: year.narrative, status: 'ready' }
    }
  }

  return { text: '', status: 'empty' }
}

export const PAST_EVENTS_SIGNAL_LABELS: Record<string, { label: string; color: string }> = {
  '婚恋_合': { label: '婚恋↑', color: 'var(--wu-huo)' },
  '婚恋_冲': { label: '婚恋↓', color: 'var(--status-danger)' },
  '婚恋_变': { label: '婚恋变', color: 'var(--wu-tu)' },
  '事业': { label: '事业', color: 'var(--wu-mu)' },
  '财运_得': { label: '财运↑', color: 'var(--wu-jin)' },
  '财运_损': { label: '财运↓', color: '#888' },
  '健康': { label: '健康↓', color: 'var(--status-danger)' },
  '迁变': { label: '迁变', color: 'var(--wu-shui)' },
  '伏吟': { label: '伏吟', color: 'var(--status-danger)' },
  '反吟': { label: '反吟', color: 'var(--status-danger)' },
  '大运合化': { label: '合化', color: 'var(--wu-tu)' },
  '喜神临运': { label: '喜神', color: 'var(--wu-jin)' },
  '综合变动': { label: '变动', color: 'var(--wu-shui)' },
  '局势_重': { label: '成局', color: 'var(--status-danger)' },
  '夹拱': { label: '夹拱', color: 'var(--wu-tu)' },
  '学业_资源': { label: '学业↑', color: 'var(--wu-mu)' },
  '学业_竞争': { label: '竞争', color: '#888' },
  '学业_压力': { label: '压力↓', color: 'var(--status-danger)' },
  '学业_贵人': { label: '贵人', color: 'var(--wu-mu)' },
  '学业_才艺': { label: '才艺', color: 'var(--wu-mu)' },
  '性格_情谊': { label: '情谊', color: 'var(--wu-tu)' },
  '性格_叛逆': { label: '叛逆', color: 'var(--status-danger)' },
}

export function getPastEventsSegmentViewState(summary?: PastEventsDayunSummaryLike): PastEventsSegmentViewState {
  if (!summary) {
    return {
      state: 'algorithm_ready',
      label: '算法信号已就绪',
      description: '年份信号已生成，等待大运批语状态。',
      canRevealSignals: false,
      canGenerateAi: false,
      canRetry: false,
      isComplete: false,
    }
  }

  if (summary.loading || summary.status === 'loading') {
    return {
      state: 'generating',
      label: '批语生成中',
      description: '年份信号已就绪，本段大运批语正在生成。',
      canRevealSignals: false,
      canGenerateAi: false,
      canRetry: false,
      isComplete: false,
    }
  }

  if (summary.status === 'interrupted' || summary.error) {
    return {
      state: 'interrupted',
      label: '生成中断，可重试',
      description: summary.error || '本段大运批语生成中断，可单独重试。',
      canRevealSignals: false,
      canGenerateAi: false,
      canRetry: true,
      isComplete: false,
    }
  }

  if (summary.summary || summary.years?.length || summary.themes?.length) {
    if (summary.cached === true) {
      return {
        state: 'generated',
        label: '本次缓存',
        description: '本段批语来自当前命盘已生成内容，不复用同八字旧缓存。',
        canRevealSignals: false,
        canGenerateAi: false,
        canRetry: false,
        isComplete: true,
      }
    }
    if (summary.cached === false) {
      return {
        state: 'generated',
        label: '批语刚生成',
        description: '本段批语已为当前命盘重新生成，并保存为本次命盘缓存。',
        canRevealSignals: false,
        canGenerateAi: false,
        canRetry: false,
        isComplete: true,
      }
    }
    return {
      state: 'generated',
      label: '批语已完成',
      description: '本段大运批语已生成或已从缓存读取。',
      canRevealSignals: false,
      canGenerateAi: false,
      canRetry: false,
      isComplete: true,
    }
  }

  if (summary.folded) {
    return {
      state: 'future_folded',
      label: '未来大运，待生成批语',
      description: '可先展开查看算法年份信号，不会生成批语。',
      canRevealSignals: true,
      canGenerateAi: false,
      canRetry: false,
      isComplete: false,
    }
  }

  return {
    state: 'future_expanded_ungenerated',
    label: '年份信号已展开',
    description: '当前只展示算法年份信号，可按需生成本段批语。',
    canRevealSignals: false,
    canGenerateAi: true,
    canRetry: false,
    isComplete: false,
  }
}

export function countPastEventsProgress(
  summaries: Record<number, PastEventsDayunSummaryLike>,
): PastEventsProgressCounts {
  return Object.values(summaries).reduce<PastEventsProgressCounts>((counts, summary) => {
    const viewState = getPastEventsSegmentViewState(summary)
    if (viewState.state === 'generated') counts.generated += 1
    if (viewState.state === 'generating') counts.generating += 1
    if (viewState.state === 'interrupted') counts.interrupted += 1
    if (viewState.state === 'future_folded' || viewState.state === 'future_expanded_ungenerated') {
      counts.futurePending += 1
    }
    return counts
  }, {
    generated: 0,
    generating: 0,
    interrupted: 0,
    futurePending: 0,
  })
}

export function buildPastEventsExportSegments(
  dayunMeta: PastEventsDayunMetaLike[],
  years: PastEventsYearNarrativeLike[],
  summaries: Record<number, PastEventsDayunSummaryLike>,
  includeEvidence = false,
): PastEventsExportReadySegment[] {
  const segments: PastEventsExportReadySegment[] = []
  for (const meta of dayunMeta) {
    const summary = summaries[meta.index]
    if (
      !summary ||
      summary.loading ||
      summary.status === 'loading' ||
      summary.status === 'interrupted' ||
      summary.error ||
      !summary.summary?.trim()
    ) {
      continue
    }
    const generatedByYear = new Map<number, PastEventsYearNarrativeEntryLike>()
    summary.years?.forEach((entry) => {
      if (entry.narrative?.trim()) generatedByYear.set(entry.year, entry)
    })
    const exportYears: PastEventsExportReadyYear[] = []
    for (const year of years) {
      if (year.dayun_index !== meta.index) continue
      const generated = generatedByYear.get(year.year)
      const narrative = generated?.narrative?.trim()
      if (!generated || !narrative) continue
      exportYears.push({
        year: year.year,
        age: year.age,
        gan_zhi: generated.ganzhi || generated.gan_zhi || year.gan_zhi || '',
        narrative,
        signals: includeEvidence ? year.signals : undefined,
        evidence_summary: includeEvidence ? year.evidence_summary : undefined,
        year_in_dayun: year.year_in_dayun,
        dayun_phase: year.dayun_phase,
      })
    }

    if (exportYears.length > 0) {
      segments.push({
        dayun_index: meta.index,
        gan_zhi: meta.gan_zhi || '',
        start_age: meta.start_age,
        end_age: meta.end_age,
        start_year: meta.start_year,
        end_year: meta.end_year,
        themes: summary.themes || [],
        summary: summary.summary.trim(),
        years: exportYears,
      })
    }
  }
  return segments
}

export function filterPastEventsExportSegments(
  segments: PastEventsExportReadySegment[],
  includeEvidence = false,
): PastEventsExportReadySegment[] {
  const exportable: PastEventsExportReadySegment[] = []
  for (const segment of segments) {
    const summary = segment.summary?.trim()
    if (!summary) continue
    const years: PastEventsExportReadyYear[] = []
    for (const year of segment.years || []) {
      const narrative = year.narrative?.trim()
      if (!narrative) continue
      years.push({
        ...year,
        narrative,
        signals: includeEvidence ? year.signals : undefined,
        evidence_summary: includeEvidence ? year.evidence_summary : undefined,
      })
    }
    if (years.length > 0) {
      exportable.push({
        ...segment,
        themes: segment.themes || [],
        summary,
        years,
      })
    }
  }
  return exportable
}

export function getPastEventsSignalLabel(signal: string) {
  return PAST_EVENTS_SIGNAL_LABELS[signal] || { label: signal, color: 'var(--text-muted)' }
}
