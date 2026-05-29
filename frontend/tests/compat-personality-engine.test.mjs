import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (p) => readFileSync(resolve(root, p), 'utf8')

// Node 23 strips types; the engine's only imports are type-only (erased at runtime).
const engine = await import('../src/lib/compatibilityPersonality.ts')
const { buildParticipantPortrait, buildPersonalityContrast, buildPersonalityFitSummary } = engine

const dim = (portrait, key) => portrait.dimensions.find(d => d.key === key)

// --- 6.1 不同命盘 → 不同画像 -----------------------------------------------
test('不同主导十神产出不同画像', () => {
  const shishang = buildParticipantPortrait({ day_gan: '甲', day_gan_wuxing: '木', ming_ge: '食神格' }, 'A', 'self')
  const guansha = buildParticipantPortrait({ day_gan: '庚', day_gan_wuxing: '金', ming_ge: '正官格' }, 'B', 'partner')
  assert.equal(shishang.hasStructuredData, true)
  assert.equal(guansha.hasStructuredData, true)
  assert.equal(shishang.dimensions.length, 5)
  assert.notEqual(shishang.headline, guansha.headline)
  assert.notEqual(dim(shishang, 'expression').detail, dim(guansha, 'expression').detail)
})

test('旺衰不同产出不同的“压力下的样子”', () => {
  const strong = buildParticipantPortrait(
    { day_gan: '甲', day_gan_wuxing: '木', ming_ge: '比肩格', wuxing: { mu: 5, huo: 1, tu: 1, jin: 0, shui: 3 } },
    'A', 'self',
  )
  const weak = buildParticipantPortrait(
    { day_gan: '甲', day_gan_wuxing: '木', ming_ge: '比肩格', wuxing: { mu: 1, huo: 4, tu: 3, jin: 2, shui: 0 } },
    'B', 'partner',
  )
  assert.notEqual(dim(strong, 'pressure').detail, dim(weak, 'pressure').detail)
})

// --- 6.2 缺字段降级 ---------------------------------------------------------
test('仅有基础四柱(日主)时输出简化画像且不报错', () => {
  const p = buildParticipantPortrait({ day_gan: '丙' }, 'A', 'self')
  assert.equal(p.hasStructuredData, false)
  assert.match(p.headline, /日主丙/)
  assert.ok(p.dimensions.length >= 1)
})

test('命盘几乎为空时回退为通用画像且不报错', () => {
  const p = buildParticipantPortrait({}, '对方', 'partner')
  assert.equal(p.hasStructuredData, false)
  assert.ok(p.dimensions.length >= 1)
})

// --- 6.3 差异对照 -----------------------------------------------------------
test('印旺×食伤旺 → 判为“照顾与被照顾”合点', () => {
  const yin = { day_gan: '甲', day_gan_wuxing: '木', ming_ge: '正印格' }
  const shishang = { day_gan: '丙', day_gan_wuxing: '火', ming_ge: '伤官格' }
  const { fitPoints } = buildPersonalityContrast(yin, shishang, 'A', 'B')
  assert.ok(fitPoints.some(p => p.title.includes('照顾')))
})

test('双比劫主导 → 判为“都强势”冲突点', () => {
  const a = { day_gan: '甲', day_gan_wuxing: '木', ming_ge: '比肩格' }
  const b = { day_gan: '乙', day_gan_wuxing: '木', ming_ge: '劫财格' }
  const { clashPoints } = buildPersonalityContrast(a, b, 'A', 'B')
  assert.ok(clashPoints.some(p => p.title.includes('强势')))
})

test('日主五行相克 → 判为“相克易顶撞”冲突点', () => {
  const wood = { day_gan: '甲', day_gan_wuxing: '木', ming_ge: '食神格' }
  const earth = { day_gan: '戊', day_gan_wuxing: '土', ming_ge: '正财格' } // 木克土
  const { clashPoints } = buildPersonalityContrast(wood, earth, 'A', 'B')
  assert.ok(clashPoints.some(p => p.title.includes('相克')))
})

// --- 6.4 与分数版本解耦 / hasReport 为否仍渲染 -------------------------------
test('无 legacy 分数(V3)时仍构建双画像与差异对照', () => {
  const summary = buildPersonalityFitSummary({
    self: { name: '甲', chart: { day_gan: '甲', day_gan_wuxing: '木', ming_ge: '食神格' } },
    partner: { name: '乙', chart: { day_gan: '庚', day_gan_wuxing: '金', ming_ge: '正官格' } },
    hasReport: false,
  })
  assert.equal(summary.matchType, '待磨合观察型')
  assert.equal(summary.selfPortrait.hasStructuredData, true)
  assert.equal(summary.partnerPortrait.hasStructuredData, true)
  assert.ok(summary.fitPoints.length >= 1)
  assert.ok(summary.clashPoints.length >= 1)
})

// --- 静态：渲染门控已解除，PersonalityFit 在页面层无条件渲染 -----------------
test('PersonalityFit 在页面层无条件渲染（不再被门控、不在 SectionDeepAnalysis 内）', () => {
  const page = read('src/pages/CompatibilityResultPage.tsx')
  const deep = read('src/components/compatibility/SectionDeepAnalysis.tsx')
  // 无条件渲染：页面直接渲染，无 `personalitySummary &&` 门控
  assert.doesNotMatch(page, /\{personalitySummary &&/)
  assert.match(page, /<PersonalityFit summary=\{personalitySummary\} \/>/)
  // 已从深度分析容器移出
  assert.doesNotMatch(deep, /PersonalityFit/)
})

test('PersonalityFit 渲染双画像与差异对照', () => {
  const src = read('src/components/compatibility/deep-analysis/PersonalityFit.tsx')
  assert.match(src, /summary\.selfPortrait/)
  assert.match(src, /summary\.partnerPortrait/)
  assert.match(src, /自然合的地方/)
  assert.match(src, /容易冲突的地方/)
})
