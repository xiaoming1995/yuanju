import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (p) => readFileSync(resolve(root, p), 'utf8')

test('api type exposes personality_comparison on the structured report', () => {
  const api = read('src/lib/api.ts')
  assert.match(api, /export interface CompatibilityPersonalityComparison/)
  assert.match(api, /personality_comparison\?: CompatibilityPersonalityComparison \| null/)
  assert.match(api, /fit_points: CompatibilityPersonalityPoint\[\]/)
  assert.match(api, /clash_points: CompatibilityPersonalityPoint\[\]/)
})

test('PersonalityComparison renders LLM portraits with frontend-fixed dimension labels', () => {
  const src = read('src/components/compatibility/deep-analysis/PersonalityComparison.tsx')
  // 5 维 label 由前端按 key 固定映射（防 LLM 漂移）
  assert.match(src, /expression: '表达 \/ 沟通'/)
  assert.match(src, /decision: '决策与节奏'/)
  assert.match(src, /intimacy: '亲密里的核心需求'/)
  assert.match(src, /emotion: '情绪反应'/)
  assert.match(src, /pressure: '压力下的样子'/)
  // 渲染双方画像 + 合点/冲突点
  assert.match(src, /comparison\.self/)
  assert.match(src, /comparison\.partner/)
  assert.match(src, /自然合的地方/)
  assert.match(src, /容易冲突的地方/)
  // 缺失/空 → 不渲染（返回 null）
  assert.match(src, /if \(!comparison.*return null/)
})

test('DeepReportNarrative renders personality comparison inside the structured report', () => {
  const src = read('src/components/compatibility/deep-analysis/DeepReportNarrative.tsx')
  assert.match(src, /import PersonalityComparison from '\.\/PersonalityComparison'/)
  assert.match(src, /<PersonalityComparison comparison=\{structuredReport\.personality_comparison\} \/>/)
  // 空态提示性格画像将在生成后出现
  assert.match(src, /双方性格画像与差异/)
})

test('result page no longer mounts a deterministic personality SECTION', () => {
  const page = read('src/pages/CompatibilityResultPage.tsx')
  assert.doesNotMatch(page, /PersonalityFit/)
  assert.doesNotMatch(page, /buildPersonalityFitSummary/)
})

test('deterministic personality engine is removed; reused helpers are kept', () => {
  const lib = read('src/lib/compatibilityPersonality.ts')
  // removed engine
  assert.doesNotMatch(lib, /buildParticipantPortrait/)
  assert.doesNotMatch(lib, /buildPersonalityContrast/)
  assert.doesNotMatch(lib, /buildPersonalityFitSummary/)
  assert.doesNotMatch(lib, /getPersonalityMatchTypeDescription/)
  // kept (still consumed by history page / entry page / result page)
  assert.match(lib, /export function getPersonalityMatchType/)
  assert.match(lib, /export function buildPersonalityValidationPlan/)
  assert.match(lib, /export function buildPersonalityConsultationPreview/)
  assert.match(lib, /export function getCompatibilityQuestionLabel/)
  // validation plan decoupled from the deleted summary: takes questionLabel + matchType
  assert.match(lib, /questionLabel: string/)
  assert.match(lib, /matchType: string/)
})
