import { useId, useState } from 'react'

export interface PillarsFormValue {
  yearPillar: string
  monthPillar: string
  dayPillar: string
  hourPillar: string
  gender: 'male' | 'female'
  minYear: number
  maxYear: number
}

type PillarKey = 'yearPillar' | 'monthPillar' | 'dayPillar' | 'hourPillar'

export const GAN = '甲乙丙丁戊己庚辛壬癸'
export const ZHI = '子丑寅卯辰巳午未申酉戌亥'

export const initialPillarsValue = (gender: 'male' | 'female'): PillarsFormValue => ({
  yearPillar: '甲子',
  monthPillar: '丙寅',
  dayPillar: '甲子',
  hourPillar: '甲子',
  gender,
  minYear: 1900,
  maxYear: 2030,
})

// 解析连写 / 带空格的 8 个干支字（如「甲子丙寅丁丑丙午」或「甲子 丙寅 丁丑 丙午」）。
// 偶数位须为天干、奇数位须为地支；成功返回四柱，否则返回 null。
export function parseEightChars(raw: string): Pick<PillarsFormValue, PillarKey> | null {
  const chars = [...raw.replace(/\s+/g, '')]
  if (chars.length !== 8) return null
  for (let i = 0; i < 8; i++) {
    const ok = i % 2 === 0 ? GAN.includes(chars[i]) : ZHI.includes(chars[i])
    if (!ok) return null
  }
  return {
    yearPillar: chars[0] + chars[1],
    monthPillar: chars[2] + chars[3],
    dayPillar: chars[4] + chars[5],
    hourPillar: chars[6] + chars[7],
  }
}

interface PillarsInputFormProps {
  value: PillarsFormValue
  onChange: (next: PillarsFormValue) => void
}

const PILLARS: Array<{ key: PillarKey; label: string }> = [
  { key: 'yearPillar', label: '年柱' },
  { key: 'monthPillar', label: '月柱' },
  { key: 'dayPillar', label: '日柱' },
  { key: 'hourPillar', label: '时柱' },
]

export default function PillarsInputForm({ value, onChange }: PillarsInputFormProps) {
  const formId = useId()
  const [quick, setQuick] = useState('')
  const [quickError, setQuickError] = useState('')

  const update = (patch: Partial<PillarsFormValue>) => onChange({ ...value, ...patch })

  const setGan = (key: PillarKey, gan: string) =>
    update({ [key]: gan + value[key][1] } as Partial<PillarsFormValue>)
  const setZhi = (key: PillarKey, zhi: string) =>
    update({ [key]: value[key][0] + zhi } as Partial<PillarsFormValue>)

  const handleQuick = (raw: string) => {
    setQuick(raw)
    if (raw.trim() === '') {
      setQuickError('')
      return
    }
    const parsed = parseEightChars(raw)
    if (parsed) {
      setQuickError('')
      update(parsed)
    } else {
      setQuickError('请输入 8 个干支字，如：甲子 丙寅 丁丑 丙午')
    }
  }

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
        <label className="form-label" htmlFor={`${formId}-quick`}>
          快速填入八字（选填）
          <span className="field-note">直接打字或粘贴 8 个字，自动拆到下方四柱；也可在下方逐字选择</span>
        </label>
        <input
          id={`${formId}-quick`}
          className="form-input"
          type="text"
          value={quick}
          onChange={e => handleQuick(e.target.value)}
          placeholder="如：甲子 丙寅 丁丑 丙午"
        />
        {quickError && <p className="form-error">{quickError}</p>}
      </div>

      <div className="birth-profile-fieldset">
        <div className="form-label">
          四柱八字
          <span className="field-note">点开每个格子选一个字：上为天干、下为地支</span>
        </div>
        <div className="pillars-grid">
          {PILLARS.map(p => {
            const gan = value[p.key][0]
            const zhi = value[p.key][1]
            return (
              <div className="pillar-col" key={p.key}>
                <div className="pillar-col-label">{p.label}</div>
                <select
                  className="form-select"
                  aria-label={`${p.label}天干`}
                  value={gan}
                  onChange={e => setGan(p.key, e.target.value)}
                >
                  {[...GAN].map(g => (
                    <option key={g} value={g}>{g}</option>
                  ))}
                </select>
                <select
                  className="form-select"
                  aria-label={`${p.label}地支`}
                  value={zhi}
                  onChange={e => setZhi(p.key, e.target.value)}
                >
                  {[...ZHI].map(z => (
                    <option key={z} value={z}>{z}</option>
                  ))}
                </select>
              </div>
            )
          })}
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
