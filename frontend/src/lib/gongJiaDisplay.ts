export interface GongJiaSourceLike {
  source?: string
  source_labels?: string[]
}

const GONGJIA_SOURCE_LABELS: Record<string, string> = {
  year_month: '年月夹拱',
  month_day: '月日夹拱',
  day_hour: '日时夹拱',
}

export function gongJiaSourceDisplay(item: GongJiaSourceLike): string {
  const mapped = item.source ? GONGJIA_SOURCE_LABELS[item.source] : ''
  if (mapped) return mapped

  const compactLabels = item.source_labels
    ?.map(label => label.replace(/柱$/, '').trim())
    .filter(Boolean)
    .join('')

  return compactLabels ? `${compactLabels}夹拱` : '夹拱'
}
