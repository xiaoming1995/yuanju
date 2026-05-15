import { useId } from 'react'
import { LunarMonth, LunarYear } from 'lunar-javascript'
import type { BirthProfileFormValue, ZiHourMode } from './birthProfile'
import './BirthProfileForm.css'

interface BirthProfileFormProps {
  title?: string
  value: BirthProfileFormValue
  onChange: (next: BirthProfileFormValue) => void
  showSummary?: boolean
  summaryCalibrationText?: string
  ziHourMode?: ZiHourMode
  onZiHourModeChange?: (mode: ZiHourMode) => void
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

function formatYearOption(year: number, currentYear: number) {
  const age = currentYear - year
  return age > 0 ? `${year} 年（${age}岁）` : `${year} 年`
}

function normalizeDate(next: BirthProfileFormValue) {
  const maxDay = getMaxDay(next)
  return {
    ...next,
    day: next.day > maxDay ? 1 : next.day,
  }
}

function getShichenLabel(hour: number) {
  return SHICHEN.find(item => item.value === hour)?.label.split('（')[0] || `${hour}时`
}

function formatBirthProfileSummary(
  value: BirthProfileFormValue,
  calibrationText = '按北京时间排盘',
  ziHourMode: ZiHourMode = 'late',
) {
  const genderText = value.gender === 'male' ? '男命' : '女命'
  const calendarText = value.calendarType === 'solar' ? '公历' : '农历'
  const monthText = value.calendarType === 'lunar' && value.isLeapMonth ? `闰${value.month}月` : `${value.month}月`
  const timeText = value.hour === 0
    ? (ziHourMode === 'early' ? '子时 23:00-23:59' : '子时 00:00-00:59')
    : getShichenLabel(value.hour)

  return `${genderText} · ${calendarText} ${value.year}年${monthText}${value.day}日 · ${timeText} · ${calibrationText}`
}

export default function BirthProfileForm({
  title,
  value,
  onChange,
  showSummary = false,
  summaryCalibrationText = '按北京时间排盘',
  ziHourMode = 'late',
  onZiHourModeChange,
}: BirthProfileFormProps) {
  const formId = useId()
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
    <div className="birth-profile-form">
      {title && <h2 className="serif birth-profile-title">{title}</h2>}

      <div className="birth-profile-basic">
        <div className="birth-profile-primary-grid">
          <div className="birth-profile-fieldset">
            <div className="form-label">性别</div>
            <div className="birth-profile-segmented">
              {(['male', 'female'] as const).map(g => (
                <button
                  key={g}
                  type="button"
                  className={`birth-profile-option ${value.gender === g ? 'active' : ''}`}
                  onClick={() => update({ gender: g })}
                >
                  {g === 'male' ? '♂ 男命' : '♀ 女命'}
                </button>
              ))}
            </div>
          </div>

          <div className="birth-profile-fieldset">
            <div className="form-label">历法</div>
            <div className="birth-profile-segmented">
              {(['solar', 'lunar'] as const).map(ct => (
                <button
                  key={ct}
                  type="button"
                  className={`birth-profile-option ${value.calendarType === ct ? 'active' : ''}`}
                  onClick={() => toggleCalendarType(ct)}
                >
                  {ct === 'solar' ? '公历' : '农历'}
                </button>
              ))}
            </div>
          </div>
        </div>

        <div className="birth-profile-fieldset">
          <div className="form-label">出生日期</div>
          <div className="birth-date-fields">
            <div className="form-group birth-year-field">
              <select
                id={`${formId}-year`}
                className="form-select"
                value={value.year}
                aria-label="出生年份"
                onChange={e => handleDateChange({ year: Number(e.target.value) })}
              >
                {years.map(year => (
                  <option key={year} value={year}>
                    {formatYearOption(year, currentYear)}
                  </option>
                ))}
              </select>
            </div>

            <div className="form-group">
              <select
                id={`${formId}-month`}
                className="form-select"
                value={`${value.month}-${value.isLeapMonth}`}
                aria-label="出生月份"
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
              <select
                id={`${formId}-day`}
                className="form-select"
                value={value.day}
                aria-label="出生日期"
                onChange={e => update({ day: Number(e.target.value) })}
              >
                {days.map(day => (
                  <option key={day} value={day}>
                    {day} 日
                  </option>
                ))}
              </select>
            </div>
          </div>
        </div>

        <div className="birth-profile-fieldset">
          <label className="form-label" htmlFor={`${formId}-hour`}>出生时辰</label>
          <select
            id={`${formId}-hour`}
            className="form-select"
            value={value.hour}
            onChange={e => update({ hour: Number(e.target.value) })}
          >
            {SHICHEN.map(item => (
              <option key={item.value} value={item.value}>
                {item.label}
              </option>
            ))}
          </select>
        </div>

        {value.hour === 0 && onZiHourModeChange && (
          <div className="birth-profile-fieldset zi-hour-group">
            <div className="form-label">子时细分</div>
            <div className="zi-hour-options">
              <button
                type="button"
                className={`zi-hour-option ${ziHourMode === 'early' ? 'active' : ''}`}
                onClick={() => onZiHourModeChange('early')}
              >
                <span>23:00-23:59</span>
                <strong>按前一日</strong>
              </button>
              <button
                type="button"
                className={`zi-hour-option ${ziHourMode === 'late' ? 'active' : ''}`}
                onClick={() => onZiHourModeChange('late')}
              >
                <span>00:00-00:59</span>
                <strong>按当日</strong>
              </button>
            </div>
          </div>
        )}
      </div>

      {showSummary && (
        <div className="birth-profile-summary" aria-live="polite">
          <span className="birth-profile-summary-label">已选</span>
          <span>{formatBirthProfileSummary(value, summaryCalibrationText, ziHourMode)}</span>
        </div>
      )}
    </div>
  )
}
