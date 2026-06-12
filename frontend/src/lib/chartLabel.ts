import type { BaziHistoryChart } from './api'

export function genderText(gender: string): string {
  return gender === 'female' ? '女命' : '男命'
}

export function formatDate(value: string): string {
  if (!value) return '-'
  return new Date(value).toLocaleDateString('zh-CN')
}

export function chartFallbackName(chart: BaziHistoryChart): string {
  return `${genderText(chart.gender)} · ${chart.birth_year}年${chart.birth_month}月${chart.birth_day}日`
}

export function chartDisplayName(chart: BaziHistoryChart): string {
  return chart.display_name?.trim() || chartFallbackName(chart)
}

export function formatPillars(chart: BaziHistoryChart): string {
  return `${chart.year_gan}${chart.year_zhi} · ${chart.month_gan}${chart.month_zhi} · ${chart.day_gan}${chart.day_zhi} · ${chart.hour_gan}${chart.hour_zhi}`
}

export function formatBirth(chart: BaziHistoryChart): string {
  return `${chart.birth_year}年${chart.birth_month}月${chart.birth_day}日 ${chart.birth_hour}时`
}
