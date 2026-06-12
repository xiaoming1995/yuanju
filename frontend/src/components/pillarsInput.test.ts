import { describe, expect, it } from 'vitest'
import { parseEightChars } from './pillarsInput'

describe('parseEightChars', () => {
  it('解析连写的 8 个干支字', () => {
    expect(parseEightChars('甲子丙寅丁丑丙午')).toEqual({
      yearPillar: '甲子',
      monthPillar: '丙寅',
      dayPillar: '丁丑',
      hourPillar: '丙午',
    })
  })

  it('允许任意空白分隔', () => {
    expect(parseEightChars('甲子 丙寅　丁丑\t丙午')).toEqual({
      yearPillar: '甲子',
      monthPillar: '丙寅',
      dayPillar: '丁丑',
      hourPillar: '丙午',
    })
  })

  it('长度不是 8 个字时返回 null', () => {
    expect(parseEightChars('甲子丙寅丁丑')).toBeNull()
    expect(parseEightChars('甲子丙寅丁丑丙午甲')).toBeNull()
    expect(parseEightChars('')).toBeNull()
  })

  it('偶数位不是天干时返回 null', () => {
    expect(parseEightChars('子甲丙寅丁丑丙午')).toBeNull()
  })

  it('奇数位不是地支时返回 null', () => {
    expect(parseEightChars('甲甲丙寅丁丑丙午')).toBeNull()
  })

  it('混入非干支字符返回 null', () => {
    expect(parseEightChars('甲子丙寅丁丑丙X')).toBeNull()
  })
})
