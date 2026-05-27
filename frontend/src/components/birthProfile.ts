export interface BirthProfileFormValue {
  year: number
  month: number
  day: number
  hour: number
  gender: 'male' | 'female'
  calendarType: 'solar' | 'lunar'
  isLeapMonth: boolean
}

export type ZiHourMode = 'early' | 'late'

export interface BirthProfileImportSource {
  chartId: string
  displayName: string
  profile: BirthProfileFormValue
}

export interface BirthProfileChartLike {
  id: string
  birth_year: number
  birth_month: number
  birth_day: number
  birth_hour: number
  gender: string
  display_name?: string
  calendar_type?: 'solar' | 'lunar'
  is_leap_month?: boolean
}

export function initialBirthProfile(gender: 'male' | 'female'): BirthProfileFormValue {
  return {
    year: new Date().getFullYear() - 25,
    month: 1,
    day: 1,
    hour: 12,
    gender,
    calendarType: 'solar',
    isLeapMonth: false,
  }
}

export function chartToBirthProfile(chart: BirthProfileChartLike): BirthProfileFormValue {
  return {
    year: chart.birth_year,
    month: chart.birth_month,
    day: chart.birth_day,
    hour: chart.birth_hour,
    gender: chart.gender === 'female' ? 'female' : 'male',
    calendarType: chart.calendar_type || 'solar',
    isLeapMonth: Boolean(chart.is_leap_month),
  }
}
