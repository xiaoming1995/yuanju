# Compatibility Decision Dashboard Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rework the compatibility result page so relationship-decision users immediately see whether to continue, observe, or proceed cautiously, with risk, confidence, next action, and evidence available as supporting layers.

**Architecture:** Keep the existing backend and API contract. Add a small frontend decision-derivation module that turns existing `consulting_assessment`, `duration_assessment`, `overall_level`, and `evidences` into display-ready decision data, then refactor `CompatibilityResultPage.tsx` to render a decision dashboard, decision evidence summary, stage validation tasks, strategy, scores, AI deep reading, and folded professional details in that order.

**Tech Stack:** React 19 + TypeScript + Vite; existing Axios API types; CSS Variables; Node built-in `node --test` for static source tests; `npm run build` for TypeScript and production build verification.

---

## File Structure

- Create `frontend/src/lib/compatibilityDecision.ts`: pure helper functions for recommendation labels, confidence labels, fallback verdicts, max-risk extraction, finding extraction, and stage-risk fallback. This keeps UX decision logic testable and prevents `CompatibilityResultPage.tsx` from growing further.
- Create `frontend/tests/compatibility-decision-dashboard.test.mjs`: static tests that assert the helper module and page expose the required decision-dashboard structure. The project has no React test runner, so this uses Node's built-in test runner and source-level assertions.
- Modify `frontend/src/pages/CompatibilityResultPage.tsx`: replace scattered decision rendering with a `DecisionDashboardPanel`, `DecisionEvidenceSummary`, revised `StageRiskGrid`, updated AI copy, and reordered sections.
- Modify `frontend/src/pages/CompatibilityResultPage.css`: add responsive styles for the dashboard, metric cards, evidence summary, stage validation cards, and mobile-first layout.
- Verify with `node --test frontend/tests/compatibility-decision-dashboard.test.mjs`, `npm run build`, and a browser/mobile smoke check.

## Task 1: Add Decision Derivation Helpers

**Files:**
- Create: `frontend/src/lib/compatibilityDecision.ts`
- Test: `frontend/tests/compatibility-decision-dashboard.test.mjs`

- [ ] **Step 1: Write the failing static test**

Create `frontend/tests/compatibility-decision-dashboard.test.mjs` with this content:

```js
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { test } from 'node:test'

const helperPath = new URL('../src/lib/compatibilityDecision.ts', import.meta.url)
const pagePath = new URL('../src/pages/CompatibilityResultPage.tsx', import.meta.url)

function read(path) {
  return readFileSync(path, 'utf8')
}

test('compatibility decision helper exposes required derivation API', () => {
  const source = read(helperPath)

  assert.match(source, /export function buildDecisionDashboardData/)
  assert.match(source, /export function buildDecisionStageRisks/)
  assert.match(source, /export function buildDecisionFindings/)
  assert.match(source, /recommendationLabelMap/)
  assert.match(source, /confidenceLabelMap/)
  assert.match(source, /verdictFromOverallLevel/)
  assert.match(source, /maxRisk/)
  assert.match(source, /nextAction/)
})

test('compatibility result page renders decision-dashboard sections before scores and AI report', () => {
  const source = read(pagePath)

  const dashboard = source.indexOf('DecisionDashboardPanel')
  const evidence = source.indexOf('DecisionEvidenceSummary')
  const stages = source.indexOf('StageRiskGrid')
  const score = source.indexOf('ScoreOverview')
  const ai = source.indexOf('compatibility-ai-card')
  const professional = source.indexOf('compatibility-professional-details')

  assert.ok(dashboard > -1, 'decision dashboard component is rendered')
  assert.ok(evidence > dashboard, 'decision evidence summary follows dashboard')
  assert.ok(stages > evidence, 'stage validation follows decision evidence')
  assert.ok(score > stages, 'score overview follows decision sections')
  assert.ok(ai > score, 'AI deep reading follows score overview')
  assert.ok(professional > ai, 'professional details follow AI reading')
  assert.match(source, /生成深度解读/)
})

test('compatibility result page includes decision-dashboard CSS hooks', () => {
  const source = read(pagePath)

  assert.match(source, /compatibility-decision-dashboard/)
  assert.match(source, /compatibility-decision-metric-grid/)
  assert.match(source, /compatibility-decision-evidence-summary/)
  assert.match(source, /compatibility-stage-validation-grid/)
})
```

- [ ] **Step 2: Run the test to verify it fails**

Run from repo root:

```bash
node --test frontend/tests/compatibility-decision-dashboard.test.mjs
```

Expected: FAIL with `ENOENT` for `frontend/src/lib/compatibilityDecision.ts` and missing page component markers.

- [ ] **Step 3: Create the decision helper module**

Create `frontend/src/lib/compatibilityDecision.ts` with this content:

```ts
import type {
  CompatibilityClaimEvidenceLink,
  CompatibilityDecisionAdvice,
  CompatibilityDurationAssessment,
  CompatibilityEvidence,
  CompatibilityFinding,
  CompatibilityRelationshipDiagnosis,
  CompatibilityStageRisk,
} from './api'

export interface DecisionFinding {
  text: string
  evidenceKeys: string[]
}

export interface DecisionDashboardData {
  verdict: string
  relationshipType: string
  summary: string
  recommendationLabel: string
  confidenceLabel: string
  maxRisk: string
  nextAction: string
  avoid: string[]
  findings: DecisionFinding[]
}

export const recommendationLabelMap: Record<string, string> = {
  continue: '可推进',
  observe: '继续观察',
  caution: '谨慎投入',
  pause: '暂缓投入',
}

export const confidenceLabelMap: Record<string, string> = {
  high: '高',
  medium: '中',
  low: '低',
}

const levelRecommendationMap: Record<string, string> = {
  high: '可推进',
  medium: '继续观察',
  low: '谨慎投入',
}

const levelVerdictMap: Record<string, string> = {
  high: '有继续推进的基础，仍需验证现实节奏',
  medium: '建议继续观察，不宜过早重投入',
  low: '建议谨慎投入，先处理边界与现实压力',
}

const durationWindowLabel: Record<string, string> = {
  three_months: '3 个月',
  one_year: '1 年',
  two_years_plus: '2 年以上',
}

const durationRiskText: Record<string, string> = {
  high: '阶段承压偏高',
  medium: '阶段风险中等',
  low: '阶段压力较低',
}

function compact(values: Array<string | undefined | null>) {
  return values.map(value => value?.trim()).filter((value): value is string => Boolean(value))
}

function firstNonEmpty(values: Array<string | undefined | null>, fallback: string) {
  return compact(values)[0] || fallback
}

export function verdictFromOverallLevel(overallLevel: string) {
  return levelVerdictMap[overallLevel] || levelVerdictMap.medium
}

export function recommendationLabel(recommendation: string | undefined, overallLevel: string) {
  if (recommendation && recommendationLabelMap[recommendation]) {
    return recommendationLabelMap[recommendation]
  }
  return levelRecommendationMap[overallLevel] || levelRecommendationMap.medium
}

export function confidenceLabel(confidence: string | undefined) {
  return confidenceLabelMap[confidence || ''] || confidenceLabelMap.medium
}

export function buildDecisionStageRisks(
  stageRisks: CompatibilityStageRisk[] | undefined,
  duration: CompatibilityDurationAssessment,
): CompatibilityStageRisk[] {
  const safeRisks = Array.isArray(stageRisks) ? stageRisks.filter(risk => risk && risk.window) : []
  if (safeRisks.length > 0) return safeRisks

  const windows = [
    { window: 'three_months', level: duration.windows?.three_months?.level },
    { window: 'one_year', level: duration.windows?.one_year?.level },
    { window: 'two_years_plus', level: duration.windows?.two_years_plus?.level },
  ]

  return windows.map(item => ({
    window: item.window,
    risk_level: item.level || 'medium',
    main_risk: `${durationWindowLabel[item.window]}${durationRiskText[item.level || 'medium']}`,
    trigger: '关系推进、沟通频率或现实安排发生变化时',
    advice: '先观察冲突后的修复能力，再决定是否增加投入。',
    evidence_keys: [],
  }))
}

export function buildDecisionFindings(
  diagnosis: CompatibilityRelationshipDiagnosis | undefined,
  evidences: CompatibilityEvidence[],
): DecisionFinding[] {
  const topFindings = Array.isArray(diagnosis?.top_findings)
    ? diagnosis.top_findings.filter(Boolean).slice(0, 3)
    : []

  if (topFindings.length > 0) {
    return topFindings.map((finding: CompatibilityFinding) => ({
      text: finding.text,
      evidenceKeys: finding.evidence_keys || [],
    })).filter(finding => finding.text)
  }

  const positive = evidences.find(evidence => evidence.polarity === 'positive')
  const negative = evidences.find(evidence => evidence.polarity === 'negative')
  return [positive, negative]
    .filter((evidence): evidence is CompatibilityEvidence => Boolean(evidence))
    .slice(0, 3)
    .map(evidence => ({
      text: `${evidence.title}：${evidence.detail}`,
      evidenceKeys: compact([evidence.evidence_key || evidence.id]),
    }))
}

export function buildDecisionDashboardData({
  diagnosis,
  advice,
  stageRisks,
  duration,
  evidences,
  overallLevel,
}: {
  diagnosis?: CompatibilityRelationshipDiagnosis
  advice?: CompatibilityDecisionAdvice
  stageRisks?: CompatibilityStageRisk[]
  duration: CompatibilityDurationAssessment
  evidences: CompatibilityEvidence[]
  overallLevel: string
}): DecisionDashboardData {
  const resolvedStageRisks = buildDecisionStageRisks(stageRisks, duration)
  const negativeEvidence = evidences.find(evidence => evidence.polarity === 'negative')
  const maxRisk = firstNonEmpty(
    [
      resolvedStageRisks[0]?.main_risk,
      negativeEvidence ? `${negativeEvidence.title}：${negativeEvidence.detail}` : '',
    ],
    '短期先验证沟通节奏和现实安排是否稳定。',
  )
  const nextAction = firstNonEmpty(
    Array.isArray(advice?.do_next) ? advice?.do_next : [],
    resolvedStageRisks[0]?.advice || '未来 1-2 个月先观察冲突后的修复能力。',
  )

  return {
    verdict: firstNonEmpty([diagnosis?.verdict], verdictFromOverallLevel(overallLevel)),
    relationshipType: firstNonEmpty([diagnosis?.relationship_type], '关系需要结合现实节奏观察'),
    summary: firstNonEmpty([diagnosis?.summary], '这段关系需要把短期吸引和长期稳定分开验证。'),
    recommendationLabel: recommendationLabel(advice?.recommendation, overallLevel),
    confidenceLabel: confidenceLabel(advice?.confidence),
    maxRisk,
    nextAction,
    avoid: compact(Array.isArray(advice?.avoid) ? advice?.avoid : []).slice(0, 2),
    findings: buildDecisionFindings(diagnosis, evidences),
  }
}

export function hasLinkedEvidence(
  links: CompatibilityClaimEvidenceLink[] | undefined,
  evidenceKeys: string[],
) {
  if (!Array.isArray(links) || links.length === 0 || evidenceKeys.length === 0) return false
  return links.some(link => link.evidence_keys?.some(key => evidenceKeys.includes(key)))
}
```

- [ ] **Step 4: Run the helper test again**

Run:

```bash
node --test frontend/tests/compatibility-decision-dashboard.test.mjs
```

Expected: FAIL only on page markers, because the helper module now exists but the page has not been refactored.

- [ ] **Step 5: Commit the helper module and failing/partial test**

Run:

```bash
git add frontend/src/lib/compatibilityDecision.ts frontend/tests/compatibility-decision-dashboard.test.mjs
git commit -m "test: cover compatibility decision dashboard structure"
```

Expected: commit succeeds. It is acceptable that the new test still fails at this checkpoint because the page refactor is not implemented yet.

## Task 2: Refactor Result Page Into Decision Flow

**Files:**
- Modify: `frontend/src/pages/CompatibilityResultPage.tsx`
- Test: `frontend/tests/compatibility-decision-dashboard.test.mjs`

- [ ] **Step 1: Import decision helpers**

In `frontend/src/pages/CompatibilityResultPage.tsx`, add this import after the existing API import:

```ts
import {
  buildDecisionDashboardData,
  buildDecisionStageRisks,
  hasLinkedEvidence,
  type DecisionDashboardData,
  type DecisionFinding,
} from '../lib/compatibilityDecision'
```

- [ ] **Step 2: Replace `DecisionHeroPanel` with `DecisionDashboardPanel`**

Replace the existing `DecisionHeroPanel` function with this component:

```tsx
function DecisionDashboardPanel({
  selfName,
  partnerName,
  reading,
  dashboard,
  hasReport,
  reportLoading,
  onGenerateReport,
}: {
  selfName: string
  partnerName: string
  reading: CompatibilityDetail['reading']
  dashboard: DecisionDashboardData
  hasReport: boolean
  reportLoading: boolean
  onGenerateReport: () => void
}) {
  const stageLabel = relationshipStageText[reading.relationship_stage] || relationshipStageText.general
  const questionLabel = primaryQuestionText[reading.primary_question] || primaryQuestionText.general

  return (
    <section className="card compatibility-decision-dashboard">
      <div className="compatibility-decision-header">
        <HeartHandshake size={24} />
        <h1 className="serif compatibility-decision-title">{selfName} × {partnerName}</h1>
      </div>

      <div className="compatibility-decision-context">
        <span>{stageLabel}</span>
        <span>{questionLabel}</span>
      </div>

      <div className="compatibility-consulting-kicker">关系决策</div>
      <h2 className="serif compatibility-decision-headline">{dashboard.verdict}</h2>
      <div className="compatibility-decision-type">{dashboard.relationshipType}</div>

      <div className="compatibility-decision-metric-grid">
        <div className="compatibility-decision-metric">
          <span>投入建议</span>
          <strong>{dashboard.recommendationLabel}</strong>
        </div>
        <div className="compatibility-decision-metric">
          <span>判断信心</span>
          <strong>{dashboard.confidenceLabel}</strong>
        </div>
        <div className="compatibility-decision-metric compatibility-decision-metric--wide">
          <span>最大风险</span>
          <strong>{dashboard.maxRisk}</strong>
        </div>
      </div>

      <div className="compatibility-next-action">
        <span>下一步验证</span>
        <p>{dashboard.nextAction}</p>
      </div>

      {dashboard.avoid.length > 0 && (
        <div className="compatibility-avoid-list">
          <span>短期避免</span>
          <div>
            {dashboard.avoid.map(item => <em key={item}>{item}</em>)}
          </div>
        </div>
      )}

      <div className="compatibility-core-contradiction">
        <span>核心矛盾</span>
        <p>{dashboard.summary}</p>
      </div>

      {!hasReport && (
        <div className="compatibility-decision-report-action">
          <button className="btn btn-primary" onClick={onGenerateReport} disabled={reportLoading}>
            {reportLoading ? '生成中' : '生成深度解读'}
          </button>
        </div>
      )}
    </section>
  )
}
```

- [ ] **Step 3: Add decision evidence summary component**

Add this component below `AdviceList`:

```tsx
function DecisionEvidenceSummary({
  findings,
  links,
}: {
  findings: DecisionFinding[]
  links: CompatibilityClaimEvidenceLink[]
}) {
  if (findings.length === 0) return null

  return (
    <section className="compatibility-section compatibility-decision-evidence-summary">
      <div className="compatibility-section-header">
        <h2 className="serif compatibility-section-title">为什么这么判断</h2>
        <p className="compatibility-section-desc">先看白话依据，专业命理证据可继续下钻。</p>
      </div>
      <div className="compatibility-decision-evidence-grid">
        {findings.map((finding, index) => (
          <div key={`${finding.text}-${index}`} className="card compatibility-decision-evidence-card">
            <span className="compatibility-decision-evidence-index">{index + 1}</span>
            <p>{finding.text}</p>
            {hasLinkedEvidence(links, finding.evidenceKeys) && (
              <a href="#compatibility-claim-evidence" className="compatibility-evidence-link">查看依据</a>
            )}
          </div>
        ))}
      </div>
    </section>
  )
}
```

- [ ] **Step 4: Revise `StageRiskGrid` copy and classes**

Replace `StageRiskGrid` with:

```tsx
function StageRiskGrid({ risks }: { risks: CompatibilityStageRisk[] }) {
  return (
    <div className="compatibility-stage-validation-grid">
      {risks.map(risk => (
        <div key={risk.window} className="card compatibility-stage-validation-card">
          <div className="compatibility-stage-window">{stageWindowText[risk.window] || risk.window}</div>
          <div className="compatibility-stage-label">要验证什么</div>
          <div className="serif compatibility-stage-risk">{risk.main_risk}</div>
          <p><span>触发场景：</span>{risk.trigger}</p>
          <div className="compatibility-stage-advice">{risk.advice}</div>
        </div>
      ))}
    </div>
  )
}
```

- [ ] **Step 5: Build dashboard data in page render**

In the main component, after `const consulting = normalizeConsultingAssessment(detail)`, add:

```ts
  const decisionStageRisks = buildDecisionStageRisks(consulting.stage_risks, durationAssessment)
  const decisionDashboard = buildDecisionDashboardData({
    diagnosis: consulting.relationship_diagnosis,
    advice: consulting.decision_advice,
    stageRisks: consulting.stage_risks,
    duration: durationAssessment,
    evidences: detail.evidences,
    overallLevel: reading.overall_level,
  })
```

- [ ] **Step 6: Reorder page sections**

In the JSX return, replace the current top-level sequence from `DecisionHeroPanel` through professional details with this order:

```tsx
        <DecisionDashboardPanel
          reading={reading}
          dashboard={decisionDashboard}
          selfName={selfP?.display_name || '我'}
          partnerName={partnerP?.display_name || '对方'}
          hasReport={Boolean(detail.latest_report)}
          reportLoading={reportLoading}
          onGenerateReport={handleGenerateReport}
        />

        <DecisionEvidenceSummary
          findings={decisionDashboard.findings}
          links={consulting.claim_evidence_links}
        />

        <div className="compatibility-section">
          <div className="compatibility-section-header">
            <h2 className="serif compatibility-section-title">接下来要验证什么</h2>
            <p className="compatibility-section-desc">按关系推进阶段看风险、触发点和验证动作。</p>
          </div>
          <StageRiskGrid risks={decisionStageRisks} />
          <DurationTaskSummary assessment={durationAssessment} />
        </div>

        {consulting.relationship_strategy && (
          <RelationshipStrategyPanel strategy={consulting.relationship_strategy} />
        )}

        <ScoreOverview scores={reading.dimension_scores} />

        {consulting.claim_evidence_links.length > 0 && (
          <div id="compatibility-claim-evidence" className="compatibility-section">
            <div className="compatibility-section-header">
              <h2 className="serif compatibility-section-title">关键判断依据</h2>
              <p className="compatibility-section-desc">每条咨询判断都可以回看对应命理证据。</p>
            </div>
            <EvidenceLinkedClaims links={consulting.claim_evidence_links} evidences={detail.evidences} />
          </div>
        )}
```

Then move the existing `<div className="card compatibility-ai-card">` block so it appears before the existing `<details className="compatibility-professional-details">` block. Keep the current AI report rendering logic inside the moved block.

- [ ] **Step 7: Remove the old `InsightPanel` section**

Delete the render block titled `关系洞察` and delete the now-unused `InsightPanel` function. Keep `fallbackAdvice` only if another section still uses it; otherwise delete `fallbackAdvice`, `reportRisks`, `fallbackRisks`, `insightRisks`, and `insightAdvice`.

- [ ] **Step 8: Run the static test**

Run:

```bash
node --test frontend/tests/compatibility-decision-dashboard.test.mjs
```

Expected: PASS.

- [ ] **Step 9: Commit the page refactor**

Run:

```bash
git add frontend/src/pages/CompatibilityResultPage.tsx frontend/tests/compatibility-decision-dashboard.test.mjs
git commit -m "feat: prioritize compatibility decision dashboard"
```

Expected: commit succeeds.

## Task 3: Add Responsive Dashboard Styles

**Files:**
- Modify: `frontend/src/pages/CompatibilityResultPage.css`
- Test: `frontend/tests/compatibility-decision-dashboard.test.mjs`

- [ ] **Step 1: Add dashboard styles**

Append this CSS to `frontend/src/pages/CompatibilityResultPage.css` near the existing decision/consulting styles:

```css
.compatibility-decision-dashboard {
  padding: 24px;
  margin-bottom: 20px;
  border-color: rgba(201, 168, 76, 0.28);
  background:
    linear-gradient(180deg, rgba(201, 168, 76, 0.1), rgba(255, 255, 255, 0.03) 44%),
    rgba(18, 23, 36, 0.94);
}

.compatibility-decision-metric-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
  margin: 18px 0;
}

.compatibility-decision-metric {
  min-width: 0;
  padding: 14px;
  border: 1px solid var(--border-subtle);
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.04);
}

.compatibility-decision-metric span,
.compatibility-next-action span,
.compatibility-avoid-list > span {
  display: block;
  margin-bottom: 7px;
  color: var(--text-muted);
  font-size: 12px;
}

.compatibility-decision-metric strong {
  display: block;
  color: var(--wu-jin);
  font-size: 18px;
  line-height: 1.45;
}

.compatibility-next-action {
  padding: 16px;
  border-radius: 14px;
  border: 1px solid rgba(201, 168, 76, 0.22);
  background: rgba(201, 168, 76, 0.08);
}

.compatibility-next-action p {
  margin: 0;
  color: var(--text-primary);
  line-height: 1.75;
}

.compatibility-avoid-list {
  margin-top: 12px;
  padding: 14px;
  border-radius: 12px;
  background: rgba(239, 83, 80, 0.07);
  border: 1px solid rgba(239, 83, 80, 0.18);
}

.compatibility-avoid-list div {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.compatibility-avoid-list em {
  padding: 5px 9px;
  border-radius: 999px;
  color: #ffb4ab;
  font-size: 12px;
  font-style: normal;
  background: rgba(239, 83, 80, 0.12);
}

.compatibility-decision-evidence-summary {
  margin-bottom: 20px;
}

.compatibility-decision-evidence-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.compatibility-decision-evidence-card {
  position: relative;
  padding: 18px;
}

.compatibility-decision-evidence-index {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  margin-bottom: 10px;
  border-radius: 999px;
  color: var(--bg-primary);
  background: var(--wu-jin);
  font-size: 12px;
  font-weight: 700;
}

.compatibility-decision-evidence-card p {
  margin: 0;
  color: var(--text-secondary);
  line-height: 1.75;
}

.compatibility-evidence-link {
  display: inline-flex;
  margin-top: 12px;
  color: var(--wu-jin);
  font-size: 13px;
  text-decoration: none;
}

.compatibility-stage-validation-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.compatibility-stage-validation-card {
  padding: 18px;
}
```

- [ ] **Step 2: Add mobile styles**

Append this media query near the end of `CompatibilityResultPage.css`, or merge into the existing mobile media query if one already covers this page:

```css
@media (max-width: 720px) {
  .compatibility-decision-dashboard {
    padding: 18px;
  }

  .compatibility-decision-headline {
    font-size: 24px;
    line-height: 1.35;
  }

  .compatibility-decision-metric-grid,
  .compatibility-decision-evidence-grid,
  .compatibility-stage-validation-grid {
    grid-template-columns: 1fr;
  }

  .compatibility-decision-metric {
    padding: 12px;
  }

  .compatibility-decision-metric strong {
    font-size: 16px;
  }

  .compatibility-next-action,
  .compatibility-avoid-list {
    padding: 13px;
  }
}
```

- [ ] **Step 3: Run static test**

Run:

```bash
node --test frontend/tests/compatibility-decision-dashboard.test.mjs
```

Expected: PASS.

- [ ] **Step 4: Commit styles**

Run:

```bash
git add frontend/src/pages/CompatibilityResultPage.css
git commit -m "style: polish compatibility decision dashboard"
```

Expected: commit succeeds.

## Task 4: Build Verification and Browser Smoke Check

**Files:**
- Verify: `frontend/src/lib/compatibilityDecision.ts`
- Verify: `frontend/src/pages/CompatibilityResultPage.tsx`
- Verify: `frontend/src/pages/CompatibilityResultPage.css`
- Verify: `frontend/tests/compatibility-decision-dashboard.test.mjs`

- [ ] **Step 1: Run static test**

Run from repo root:

```bash
node --test frontend/tests/compatibility-decision-dashboard.test.mjs
```

Expected: PASS with three passing subtests.

- [ ] **Step 2: Run frontend build**

Run:

```bash
cd /Users/liujiming/web/yuanju/frontend && npm run build
```

Expected: PASS. Vite chunk-size warnings are acceptable; TypeScript errors are not.

- [ ] **Step 3: Start frontend dev server**

Run:

```bash
cd /Users/liujiming/web/yuanju/frontend && npm run dev
```

Expected: Vite starts and prints a local URL, usually `http://localhost:5173/`.

- [ ] **Step 4: Browser smoke check on compatibility result page**

Use the in-app browser at the Vite URL. Log in if needed, open or create a compatibility reading, and verify:

```text
Top section is the decision dashboard.
Dashboard shows verdict, 投入建议, 判断信心, 最大风险, 下一步验证.
The button says 生成深度解读 when no AI report exists.
“为什么这么判断” appears before stage risks.
“接下来要验证什么” appears before 四维分数.
AI 合盘解读 appears before 专业命盘细节.
Mobile width does not overlap bottom navigation and key text does not overflow.
```

- [ ] **Step 5: Stop dev server**

Stop the Vite process with Ctrl-C in the terminal session.

- [ ] **Step 6: Final commit if verification required code/style tweaks**

If verification required changes, run:

```bash
git add frontend/src/lib/compatibilityDecision.ts frontend/src/pages/CompatibilityResultPage.tsx frontend/src/pages/CompatibilityResultPage.css frontend/tests/compatibility-decision-dashboard.test.mjs
git commit -m "fix: verify compatibility decision dashboard"
```

Expected: commit succeeds only if there were follow-up fixes. If no files changed, skip this commit.

## Self-Review

- Spec coverage: The plan covers the decision dashboard, decision evidence summary, stage validation tasks, strategy ordering, score demotion, AI deep-reading copy, professional details ordering, fallback derivation, and mobile verification.
- Scope: The plan does not touch backend algorithms, AI prompt text, multi-object comparison, relationship tracking, or date predictions.
- Placeholder scan: No task contains open-ended placeholder instructions; each code-changing step names exact files and provides concrete snippets.
- Type consistency: Helper names used by the page and static tests are consistent: `buildDecisionDashboardData`, `buildDecisionStageRisks`, `buildDecisionFindings`, `hasLinkedEvidence`, `DecisionDashboardData`, and `DecisionFinding`.
