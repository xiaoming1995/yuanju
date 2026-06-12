import { describe, expect, it } from 'vitest'
import type { CompatibilityDimensionScoresLegacy } from './api'
import {
  buildPersonalityValidationPlan,
  getCompatibilityQuestionLabel,
  getCompatibilityStageLabel,
  getPersonalityMatchType,
} from './compatibilityPersonality'

describe('getCompatibilityQuestionLabel', () => {
  it('已知问题返回对应文案', () => {
    expect(getCompatibilityQuestionLabel('marriage_suitability')).toBe('适不适合结婚')
  })
  it('未知或空值回退到 general', () => {
    expect(getCompatibilityQuestionLabel('unknown_key')).toBe('性格合不合')
    expect(getCompatibilityQuestionLabel(null)).toBe('性格合不合')
    expect(getCompatibilityQuestionLabel(undefined)).toBe('性格合不合')
  })
})

describe('getCompatibilityStageLabel', () => {
  it('已知阶段返回对应文案', () => {
    expect(getCompatibilityStageLabel('long_distance')).toBe('异地中')
  })
  it('未知或空值回退到 general', () => {
    expect(getCompatibilityStageLabel('whatever')).toBe('综合关系判断')
    expect(getCompatibilityStageLabel(null)).toBe('综合关系判断')
  })
})

describe('getPersonalityMatchType', () => {
  it('高吸引但沟通或稳定低 → 高吸引高消耗型', () => {
    expect(getPersonalityMatchType({ attraction: 80, stability: 70, communication: 55, practicality: 70 })).toBe('高吸引高消耗型')
  })
  it('主问题为反复拉扯或沟通过低 → 反复拉扯型', () => {
    expect(getPersonalityMatchType({ attraction: 60, stability: 70, communication: 50, practicality: 70 })).toBe('反复拉扯型')
    expect(getPersonalityMatchType({ attraction: 60, stability: 70, communication: 70, practicality: 70 }, 'recurring_conflict')).toBe('反复拉扯型')
  })
  it('异地或现实分过低 → 现实压力型', () => {
    expect(getPersonalityMatchType({ attraction: 60, stability: 70, communication: 70, practicality: 50 })).toBe('现实压力型')
    expect(getPersonalityMatchType({ attraction: 60, stability: 70, communication: 70, practicality: 70 }, undefined, 'long_distance')).toBe('现实压力型')
  })
  it('稳定/沟通/现实都达标 → 稳定互补型', () => {
    expect(getPersonalityMatchType({ attraction: 60, stability: 75, communication: 65, practicality: 65 })).toBe('稳定互补型')
  })
  it('其余情况 → 慢热磨合型', () => {
    expect(getPersonalityMatchType({ attraction: 60, stability: 60, communication: 60, practicality: 60 })).toBe('慢热磨合型')
  })
  it('缺失分数按 0 处理且不越界', () => {
    expect(getPersonalityMatchType({} as CompatibilityDimensionScoresLegacy)).toBe('反复拉扯型')
  })
})

describe('buildPersonalityValidationPlan', () => {
  it('无任何依据时给出默认观察计划', () => {
    const plan = buildPersonalityValidationPlan({
      questionLabel: '性格合不合',
      matchType: '慢热磨合型',
      hasEvidence: false,
    })
    expect(plan.shortTerm.items).toHaveLength(2)
    expect(plan.shortTerm.items[0]).toContain('性格合不合')
    expect(plan.mediumTerm.items[0]).toContain('慢热磨合型')
    expect(plan.shortTerm.anchor).toBeUndefined()
    expect(plan.avoid.items.length).toBeLessThanOrEqual(3)
    expect(plan.supportNote).toContain('深度解读可继续补充依据')
  })

  it('有阶段风险时引用风险并带锚点', () => {
    const plan = buildPersonalityValidationPlan({
      questionLabel: '值不值得继续投入',
      matchType: '高吸引高消耗型',
      stageRisks: [
        { window: 'three_months', risk_level: 'high', main_risk: '情绪化冷战', trigger: '回避沟通', advice: '约定冲突后 24 小时内复盘', evidence_keys: [] },
        { window: 'one_year', risk_level: 'medium', main_risk: '现实规划分歧', trigger: '异地', advice: '明确同城时间表', evidence_keys: [] },
      ],
      hasEvidence: true,
    })
    expect(plan.shortTerm.items[1]).toBe('约定冲突后 24 小时内复盘')
    expect(plan.mediumTerm.items[0]).toContain('现实规划分歧')
    expect(plan.shortTerm.anchor).toBe('#compatibility-stage-validation')
    expect(plan.supportNote).toContain('阶段风险')
  })

  it('avoid 项去重并最多 3 条', () => {
    const plan = buildPersonalityValidationPlan({
      questionLabel: '性格合不合',
      matchType: '慢热磨合型',
      advice: {
        recommendation: 'observe',
        confidence: 'medium',
        do_next: [],
        avoid: ['不要把短期吸引直接等同于长期稳定。', '避免翻旧账', '避免冷处理', '避免赌气消失'],
        conditions: [],
      },
    })
    expect(plan.avoid.items).toHaveLength(3)
    expect(new Set(plan.avoid.items).size).toBe(3)
  })
})
