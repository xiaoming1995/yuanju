import { LunarMonth, LunarYear } from 'lunar-javascript'
import './BirthProfileForm.css'

export interface BirthProfileFormValue {
  year: number
  month: number
  day: number
  hour: number
  gender: 'male' | 'female'
  calendarType: 'solar' | 'lunar'
  isLeapMonth: boolean
}

interface BirthProfileFormProps {
  title?: string
  value: BirthProfileFormValue
  onChange: (next: BirthProfileFormValue) => void
}

const SHICHEN = [
  { value: 0, label: '子时（23:00 - 00:59）' },
  { value: 2, label: '丑时（01:00 - 02:59）' },
  { value: 4, label: '寅时（03:00 - 04:59）' },
  { value: 6, label: '卯时（05:00 - 06:59）' },
  { value: 8, label: '辰时（07:00 - 08:59）' },
  { value: 10, label: '巳时（09:00 - 10:59）' },
  { value: 12, label: '午时（11:00 - 12:59）' },
  { value: 14, label: '未时（13:00 - 14:59）' },
  { value: 16, label: '申时（15:00 - 16:59）' },
  { value: 18, label: '酉时（17:00 - 18:59）' },
  { value: 20, label: '戌时（19:00 - 20:59）' },
  { value: 22, label: '亥时（21:00 - 22:59）' },
]

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

function getSolarDaysInMonth(year: number, month: number) {
  return new Date(year, month, 0).getDate()
}

function getMaxDay(value: BirthProfileFormValue) {
  if (value.calendarType === 'solar') {
    return getSolarDaysInMonth(value.year, value.month)
  }
  return LunarMonth.fromYm(value.year, value.isLeapMonth ? -value.month : value.month)?.getDayCount() || 30
}

function buildLunarMonthOptions(year: number) {
  const options: Array<{ value: number; isLeap: boolean; label: string }> = []
  const leapMonth = LunarYear.fromYear(year).getLeapMonth()
  for (let i = 1; i <= 12; i++) {
    options.push({ value: i, isLeap: false, label: `${i} 月` })
    if (i === leapMonth) {
      options.push({ value: i, isLeap: true, label: `闰 ${i} 月` })
    }
  }
  return options
}

function normalizeDate(next: BirthProfileFormValue) {
  const maxDay = getMaxDay(next)
  return {
    ...next,
    day: next.day > maxDay ? 1 : next.day,
  }
}

export default function BirthProfileForm({ title, value, onChange }: BirthProfileFormProps) {
  const currentYear = new Date().getFullYear()
  const years = Array.from({ length: 120 }, (_, i) => currentYear - i)
  const monthOptions = value.calendarType === 'solar'
    ? Array.from({ length: 12 }, (_, i) => ({ value: i + 1, isLeap: false, label: `${i + 1} 月` }))
    : buildLunarMonthOptions(value.year)
  const days = Array.from({ length: getMaxDay(value) }, (_, i) => i + 1)

  const update = (patch: Partial<BirthProfileFormValue>) => {
    onChange(normalizeDate({ ...value, ...patch }))
  }

  const handleDateChange = (patch: { year?: number; month?: number; isLeapMonth?: boolean }) => {
    const nextYear = patch.year ?? value.year
    const nextMonth = patch.month ?? value.month
    let nextIsLeapMonth = patch.isLeapMonth ?? value.isLeapMonth

    if (value.calendarType === 'lunar') {
      const leapMonth = LunarYear.fromYear(nextYear).getLeapMonth()
      if (nextIsLeapMonth && nextMonth !== leapMonth) {
        nextIsLeapMonth = false
      }
    } else {
      nextIsLeapMonth = false
    }

    update({
      year: nextYear,
      month: nextMonth,
      isLeapMonth: nextIsLeapMonth,
    })
  }

  const toggleCalendarType = (calendarType: 'solar' | 'lunar') => {
    const nextValue = normalizeDate({
      ...value,
      calendarType,
      isLeapMonth: false,
    })
    onChange(nextValue)
  }

  return (
    <div>
      {title && <h2 className="serif birth-profile-title">{title}</h2>}

      <div className="gender-selector">
        {(['male', 'female'] as const).map(g => (
          <button
            key={g}
            type="button"
            className={`gender-btn ${value.gender === g ? 'active' : ''}`}
            onClick={() => update({ gender: g })}
          >
            {g === 'male' ? '♂ 男命' : '♀ 女命'}
          </button>
        ))}
      </div>

      <div className="gender-selector birth-profile-calendar-toggle">
        {(['solar', 'lunar'] as const).map(ct => (
          <button
            key={ct}
            type="button"
            className={`gender-btn ${value.calendarType === ct ? 'active' : ''}`}
            onClick={() => toggleCalendarType(ct)}
          >
            {ct === 'solar' ? '公历' : '农历'}
          </button>
        ))}
      </div>

      <div className="date-row">
        <div className="form-group">
          <label className="form-label">出生年份</label>
          <select
            className="form-select"
            value={value.year}
            onChange={e => handleDateChange({ year: Number(e.target.value) })}
          >
            {years.map(year => (
              <option key={year} value={year}>
                {year} 年
              </option>
            ))}
          </select>
        </div>

        <div className="form-group">
          <label className="form-label">月</label>
          <select
            className="form-select"
            value={`${value.month}-${value.isLeapMonth}`}
            onChange={e => {
              const [month, isLeapMonth] = e.target.value.split('-')
              handleDateChange({
                month: Number(month),
                isLeapMonth: isLeapMonth === 'true',
              })
            }}
          >
            {monthOptions.map(option => (
              <option key={`${option.value}-${option.isLeap}`} value={`${option.value}-${option.isLeap}`}>
                {option.label}
              </option>
            ))}
          </select>
        </div>

        <div className="form-group">
          <label className="form-label">日</label>
          <select className="form-select" value={value.day} onChange={e => update({ day: Number(e.target.value) })}>
            {days.map(day => (
              <option key={day} value={day}>
                {day} 日
              </option>
            ))}
          </select>
        </div>
      </div>

      <div className="form-group">
        <label className="form-label">出生时辰</label>
        <select className="form-select" value={value.hour} onChange={e => update({ hour: Number(e.target.value) })}>
          {SHICHEN.map(item => (
            <option key={item.value} value={item.value}>
              {item.label}
            </option>
          ))}
        </select>
      </div>
    </div>
  )
}
