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
