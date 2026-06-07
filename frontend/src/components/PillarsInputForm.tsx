import { useId } from 'react'

export interface PillarsFormValue {
  yearPillar: string
  monthPillar: string
  dayPillar: string
  hourPillar: string
  gender: 'male' | 'female'
  minYear: number
  maxYear: number
}

const GAN = '甲乙丙丁戊己庚辛壬癸'
const ZHI = '子丑寅卯辰巳午未申酉戌亥'

// 60 甲子（干支同阴阳配对）：甲子、乙丑 … 癸亥
export const JIAZI: string[] = Array.from({ length: 60 }, (_, i) => GAN[i % 10] + ZHI[i % 12])

export const initialPillarsValue = (gender: 'male' | 'female'): PillarsFormValue => ({
  yearPillar: '甲子',
  monthPillar: '丙寅',
  dayPillar: '甲子',
  hourPillar: '甲子',
  gender,
  minYear: 1900,
  maxYear: 2030,
})

interface PillarsInputFormProps {
  value: PillarsFormValue
  onChange: (next: PillarsFormValue) => void
}

const PILLAR_FIELDS: Array<{ key: keyof PillarsFormValue; label: string }> = [
  { key: 'yearPillar', label: '年柱' },
  { key: 'monthPillar', label: '月柱' },
  { key: 'dayPillar', label: '日柱' },
  { key: 'hourPillar', label: '时柱' },
]

export default function PillarsInputForm({ value, onChange }: PillarsInputFormProps) {
  const formId = useId()
  const update = (patch: Partial<PillarsFormValue>) => onChange({ ...value, ...patch })

  return (
    <div className="pillars-input-form">
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
        <div className="form-label">
          四柱八字
          <span className="field-note">从排盘图或已知八字里逐柱选择，不知道出生时间也能起盘</span>
        </div>
        <div className="pillars-grid">
          {PILLAR_FIELDS.map(field => (
            <div className="form-group" key={field.key}>
              <select
                id={`${formId}-${field.key}`}
                className="form-select"
                aria-label={field.label}
                value={value[field.key] as string}
                onChange={e => update({ [field.key]: e.target.value } as Partial<PillarsFormValue>)}
              >
                {JIAZI.map(gz => (
                  <option key={gz} value={gz}>{field.label.slice(0, 1)}：{gz}</option>
                ))}
              </select>
            </div>
          ))}
        </div>
      </div>

      <div className="birth-profile-fieldset">
        <div className="form-label">
          大致年代（可选）
          <span className="field-note">缩小反查范围、减少候选；不确定就留默认</span>
        </div>
        <div className="pillars-range">
          <select
            className="form-select"
            aria-label="起始年份"
            value={value.minYear}
            onChange={e => update({ minYear: Number(e.target.value) })}
          >
            {Array.from({ length: 14 }, (_, i) => 1900 + i * 10).map(y => (
              <option key={y} value={y}>{y} 年起</option>
            ))}
          </select>
          <span className="pillars-range-sep">—</span>
          <select
            className="form-select"
            aria-label="结束年份"
            value={value.maxYear}
            onChange={e => update({ maxYear: Number(e.target.value) })}
          >
            {Array.from({ length: 14 }, (_, i) => 1900 + i * 10).concat([2030]).map(y => (
              <option key={y} value={y}>{y} 年止</option>
            ))}
          </select>
        </div>
      </div>
    </div>
  )
}
