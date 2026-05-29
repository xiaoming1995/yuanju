import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { runInNewContext } from 'node:vm'
import test from 'node:test'
import ts from 'typescript'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (path) => readFileSync(resolve(root, path), 'utf8')

function loadPersonalityHelpers() {
  const source = read('src/lib/compatibilityPersonality.ts')
  const output = ts.transpileModule(source, {
    compilerOptions: {
      module: ts.ModuleKind.CommonJS,
      target: ts.ScriptTarget.ES2020,
    },
  }).outputText
  const exports = {}
  runInNewContext(output, { exports })
  return exports
}

test('compatibility personality helper exposes fit and validation APIs', () => {
  const source = read('src/lib/compatibilityPersonality.ts')

  assert.match(source, /export function buildPersonalityFitSummary/)
  assert.match(source, /export function buildPersonalityValidationPlan/)
  assert.match(source, /export function getPersonalityMatchType/)
  assert.match(source, /export function getCompatibilityQuestionLabel/)
})

test('personality fit helper renders fallback and evidence-linked fit details without AI report', () => {
  const { buildPersonalityFitSummary } = loadPersonalityHelpers()
  const summary = buildPersonalityFitSummary({
    scores: { attraction: 82, stability: 58, communication: 46, practicality: 64 },
    evidences: [
      {
        evidence_key: 'comm-risk',
        dimension: 'communication',
        polarity: 'negative',
        title: '沟通节奏容易错位',
        detail: '一方推进快，一方需要更多缓冲。',
      },
      {
        evidence_key: 'attraction-good',
        dimension: 'attraction',
        polarity: 'positive',
        title: '吸引力较强',
        detail: '靠近感明显。',
      },
    ],
    relationshipStage: 'dating',
    primaryQuestion: 'recurring_conflict',
    selfName: '我',
    partnerName: '对方',
    hasReport: false,
  })

  assert.equal(summary.questionLabel, '为什么反复拉扯')
  assert.match(summary.matchType, /高吸引高消耗型|反复拉扯型/)
  assert.ok(summary.fitPoints.length > 0)
  assert.ok(summary.clashPoints.length > 0)
  assert.ok(summary.evidenceTargets.length > 0)
  assert.match(summary.reportNote, /深度解读/)
})

test('personality validation plan uses conditional 7-day and 30-day observation language', () => {
  const { buildPersonalityFitSummary, buildPersonalityValidationPlan } = loadPersonalityHelpers()
  const personality = buildPersonalityFitSummary({
    scores: { attraction: 72, stability: 61, communication: 48, practicality: 57 },
    evidences: [],
    relationshipStage: 'ambiguous',
    primaryQuestion: 'continue_investment',
    selfName: '我',
    partnerName: '对方',
    hasReport: false,
  })
  const plan = buildPersonalityValidationPlan({
    personality,
    advice: {
      recommendation: 'observe',
      confidence: 'medium',
      conditions: ['确认对方是否能稳定回应'],
      do_next: ['先观察互动节奏'],
      avoid: ['不要过早加码投入'],
    },
    stageRisks: [
      {
        window: 'three_months',
        risk_level: 'medium',
        main_risk: '沟通节奏不一致',
        trigger: '推进关系时',
        advice: '先观察修复能力。',
        evidence_keys: ['comm-risk'],
      },
    ],
    hasEvidence: true,
  })

  assert.match(plan.shortTerm.title, /7 天/)
  assert.match(plan.mediumTerm.title, /30 天/)
  assert.match(plan.shortTerm.items.join(' '), /观察|确认|验证/)
  assert.match(plan.mediumTerm.items.join(' '), /观察|确认|验证/)
  assert.match(plan.avoid.items.join(' '), /不要|避免|暂缓/)
  assert.doesNotMatch(JSON.stringify(plan), /一定|必然|结婚日期|分手日期/)
})

test('compatibility entry puts personality consultation before birth profile forms', () => {
  const page = read('src/pages/CompatibilityPage.tsx')
  const css = read('src/pages/CompatibilityPage.css')

  const intro = page.indexOf('compatibility-personality-consultation')
  const preview = page.indexOf('compatibility-personality-preview')
  const forms = page.indexOf('compatibility-forms')
  const payload = page.indexOf('relationship_stage: relationshipStage')

  assert.ok(intro > -1, 'personality consultation section exists')
  assert.ok(preview > intro, 'preview follows consultation controls')
  assert.ok(forms > preview, 'birth forms follow personality preview')
  assert.ok(payload > -1, 'existing payload keeps relationship context')
  assert.match(page, /性格合不合/)
  assert.match(css, /compatibility-personality-consultation/)
  assert.match(css, /compatibility-personality-preview/)
})

test('compatibility result shows personality fit before scores in extracted components', () => {
  const page = read('src/pages/CompatibilityResultPage.tsx')
  const css = read('src/pages/CompatibilityResultPage.css')
  const personalityFit = read('src/components/compatibility/deep-analysis/PersonalityFit.tsx')
  const actionPlan = read('src/components/compatibility/deep-analysis/ActionPlan7d30d.tsx')
  const sectionDeep = read('src/components/compatibility/SectionDeepAnalysis.tsx')

  // narrative layout: PersonalityFit (性格) → SectionVerdict (分数) → SectionDeepAnalysis
  const personality = page.indexOf('<PersonalityFit')
  const verdict = page.indexOf('<SectionVerdict')
  const deep = page.indexOf('<SectionDeepAnalysis')
  assert.ok(personality > -1 && personality < verdict, 'PersonalityFit renders before verdict (scores)')
  assert.ok(verdict > -1, 'SectionVerdict renders scores')
  assert.ok(deep > verdict, 'SectionDeepAnalysis follows verdict (scores)')

  // personality is now its own top-level SECTION, no longer inside SectionDeepAnalysis
  assert.doesNotMatch(sectionDeep, /PersonalityFit/)
  // deep analysis inner order: 关系策略 → 阶段风险 → 下一步（AI 深度解读已移出到页面末尾）
  const strategy = sectionDeep.indexOf('<RelationshipStrategy')
  const action = sectionDeep.indexOf('<ActionPlan7d30d')
  const next = sectionDeep.indexOf('<NextStepsAndAvoid')
  assert.ok(strategy > -1 && strategy < action && action < next, 'deep analysis sub-blocks ordered 策略→风险→下一步')
  assert.doesNotMatch(sectionDeep, /DeepReportNarrative/)
  // AI 深度解读拎出深度分析，作为最后一环排在 EvidenceDrawer 之后
  const drawer = page.indexOf('<EvidenceDrawer')
  const report = page.indexOf('<DeepReportNarrative')
  assert.ok(drawer > deep, 'EvidenceDrawer follows deep analysis')
  assert.ok(report > drawer, 'AI 深度解读 is the last block, after EvidenceDrawer')

  // content lives in extracted components / helper
  assert.match(personalityFit, /双方性格画像与差异/)
  // 7-day / 30-day titles are generated by the personality helper
  const personalityHelper = read('src/lib/compatibilityPersonality.ts')
  assert.match(personalityHelper, /7 天观察/)
  assert.match(personalityHelper, /30 天验证/)
  // Personality and validation styles live in component CSS files after CSS split (T22)
  const personalityFitCss = read('src/components/compatibility/deep-analysis/PersonalityFit.css')
  const actionPlanCss = read('src/components/compatibility/deep-analysis/ActionPlan7d30d.css')
  assert.match(personalityFitCss, /compat-da-personality/)
  assert.match(actionPlanCss, /compatibility-validation-plan-grid/)
})

test('compatibility history highlights personality match type and continuation action', () => {
  const page = read('src/pages/CompatibilityHistoryPage.tsx')
  const css = read('src/pages/CompatibilityHistoryPage.css')

  assert.match(page, /getPersonalityMatchType/)
  assert.match(page, /compatibility-history-personality/)
  assert.match(page, /性格匹配/)
  assert.match(page, /继续看性格合盘|生成深度解读|查看性格合盘/)
  assert.match(css, /compatibility-history-personality/)
  assert.match(css, /compatibility-history-continuation/)
})
