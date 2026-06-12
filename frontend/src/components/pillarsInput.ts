export interface PillarsFormValue {
  yearPillar: string
  monthPillar: string
  dayPillar: string
  hourPillar: string
  gender: 'male' | 'female'
  minYear: number
  maxYear: number
}

export type PillarKey = 'yearPillar' | 'monthPillar' | 'dayPillar' | 'hourPillar'

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
