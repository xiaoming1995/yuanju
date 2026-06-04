# 合盘「名人配对类比」Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在合盘结果页顶部展示一对名人/经典情侣类比（如「梁山伯与祝英台」），由 LLM 随深度解读一并生成，贴合该盘真实关系动态。

**Architecture:** 在结构化报告 JSON schema 新增 `famous_couple` 字段（LLM 产出），同步到 Go 结构体（否则持久化 round-trip 会丢字段）与 TS 类型；前端新增顶部 `FamousCoupleCard`，三态展示（未生成钩子 / 已生成类比卡 / 旧报告无字段则隐藏）。不动评分引擎、create 路径、数据库迁移、分享图。

**Tech Stack:** Go 1.25（后端 + `go test`）、React 19 + TypeScript + Vite（前端，`node --test` 源码断言测试）。

**设计依据:** `docs/superpowers/specs/2026-06-04-compatibility-famous-couple-analogy-design.md`

---

## 运维须知（非代码任务，但必须知会）

`prompt.SyncCanonical` 是 **insert-only**：DB 已存在 `compatibility` prompt 行时为 noop（`backend/pkg/prompt/sync.go:68`），且运行时优先用 DB 里的 prompt（`compatibility_service.go:318-327`）。
- **新库 / 本地无该行**：启动时自动 seed 新 canonical 内容，`famous_couple` 直接生效。
- **已有部署（DB 已存在该行）**：drift 状态会变为 outdated（canonicalHash 改变），需管理员在后台「AI Prompt」页把 compatibility 提示词更新到最新出厂版，新 schema 才会到达 LLM。实现完成后在交付说明里提示用户执行这一步。

---

## 文件结构

| 文件 | 责任 | 动作 |
|---|---|---|
| `backend/internal/model/compatibility.go` | 结构化报告类型 | 新增 `CompatibilityFamousCouple` + `FamousCouple` 字段 |
| `backend/internal/model/compatibility_famous_couple_test.go` | round-trip 回归测试 | 新建 |
| `backend/pkg/prompt/canonical_compatibility.go` | LLM prompt schema + 约束 | 修改 |
| `backend/pkg/prompt/canonical_compatibility_famous_couple_test.go` | prompt 内容断言 | 新建 |
| `frontend/src/lib/api.ts` | TS 结构化报告类型 | 修改 `CompatibilityStructuredReport` 接口 |
| `frontend/src/components/compatibility/FamousCoupleCard.tsx` | 顶部类比卡组件 | 新建 |
| `frontend/src/components/compatibility/FamousCoupleCard.css` | 样式 | 新建 |
| `frontend/tests/compat-famous-couple-card.test.mjs` | 组件源码断言 | 新建 |
| `frontend/src/pages/CompatibilityResultPage.tsx` | 挂载类比卡 | 修改 |
| `frontend/tests/compat-result-famous-couple.test.mjs` | 挂载断言 | 新建 |

---

## Task 1: Go 结构体新增 famous_couple（载重项）

**Files:**
- Modify: `backend/internal/model/compatibility.go:203-216`
- Test: `backend/internal/model/compatibility_famous_couple_test.go`（新建）

**为什么先做这个：** `GenerateCompatibilityReport` 会把 LLM JSON `Unmarshal` 进 `CompatibilityStructuredReport` 再 `Marshal` 回去持久化（`compatibility_service.go:358-371`）。不在结构体里的字段会在这次 round-trip 中被静默丢弃。先用测试锁死「famous_couple 能存活 round-trip」。

- [ ] **Step 1: 写失败测试**

新建 `backend/internal/model/compatibility_famous_couple_test.go`：

```go
package model

import (
	"encoding/json"
	"testing"
)

// LLM 返回的 JSON 经 Unmarshal→Marshal 持久化后，famous_couple 必须存活。
func TestCompatibilityStructuredReport_FamousCoupleSurvivesRoundTrip(t *testing.T) {
	raw := `{
		"summary": "x",
		"famous_couple": {
			"couple": "梁山伯与祝英台",
			"tagline": "一见倾心，却被现实层层阻隔",
			"reason": "你们吸引力来得快而强，但长期更受现实安排牵制。"
		},
		"risks": [],
		"advice": "y"
	}`

	var report CompatibilityStructuredReport
	if err := json.Unmarshal([]byte(raw), &report); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if report.FamousCouple == nil {
		t.Fatal("FamousCouple was dropped on unmarshal")
	}
	if report.FamousCouple.Couple != "梁山伯与祝英台" {
		t.Errorf("Couple = %q, want 梁山伯与祝英台", report.FamousCouple.Couple)
	}

	out, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var again CompatibilityStructuredReport
	if err := json.Unmarshal(out, &again); err != nil {
		t.Fatalf("re-unmarshal failed: %v", err)
	}
	if again.FamousCouple == nil || again.FamousCouple.Tagline != "一见倾心，却被现实层层阻隔" {
		t.Errorf("famous_couple lost across round-trip: %+v", again.FamousCouple)
	}
}

// 旧报告没有 famous_couple 时，字段应为 nil 且 marshal 时省略（omitempty）。
func TestCompatibilityStructuredReport_FamousCoupleOmittedWhenAbsent(t *testing.T) {
	var report CompatibilityStructuredReport
	if err := json.Unmarshal([]byte(`{"summary":"x","risks":[],"advice":"y"}`), &report); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if report.FamousCouple != nil {
		t.Fatalf("expected nil FamousCouple, got %+v", report.FamousCouple)
	}
	out, _ := json.Marshal(report)
	if string(out) == "" {
		t.Fatal("marshal produced empty output")
	}
	if containsKey(out, "famous_couple") {
		t.Errorf("famous_couple should be omitted when nil, got: %s", out)
	}
}

func containsKey(b []byte, key string) bool {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(b, &m); err != nil {
		return false
	}
	_, ok := m[key]
	return ok
}
```

- [ ] **Step 2: 运行测试，确认失败**

Run: `cd backend && go test ./internal/model/ -run FamousCouple -v`
Expected: 编译失败 / FAIL —— `report.FamousCouple undefined`（结构体还没有该字段）。

- [ ] **Step 3: 加结构体字段（最小实现）**

在 `backend/internal/model/compatibility.go` 的 `CompatibilityStructuredReport` 里，`Summary` 行之后插入一行字段，并在该结构体定义下方新增子结构体：

```go
type CompatibilityStructuredReport struct {
	Summary               string                              `json:"summary"`
	FamousCouple          *CompatibilityFamousCouple          `json:"famous_couple,omitempty"`
	QuestionFocus         CompatibilityQuestionFocus          `json:"question_focus"`
	PersonalityComparison *CompatibilityPersonalityComparison `json:"personality_comparison,omitempty"`
	Dimensions            []CompatibilityDimensionNarrative   `json:"dimensions"`
	DurationAssessment    CompatibilityDurationAssessment     `json:"duration_assessment"`
	RelationshipDiagnosis CompatibilityRelationshipDiagnosis  `json:"relationship_diagnosis"`
	DecisionAdvice        CompatibilityDecisionAdvice         `json:"decision_advice"`
	StageRisks            []CompatibilityStageRisk            `json:"stage_risks"`
	RelationshipStrategy  CompatibilityRelationshipStrategy   `json:"relationship_strategy"`
	ClaimEvidenceLinks    []CompatibilityClaimEvidenceLink    `json:"claim_evidence_links"`
	Risks                 []string                            `json:"risks"`
	Advice                string                              `json:"advice"`
}

// CompatibilityFamousCouple 是 LLM 给这对关系挑的名人/经典情侣类比，
// 反映关系真实动态（可甜可虐），随深度解读生成、随报告 JSON 持久化。
type CompatibilityFamousCouple struct {
	Couple  string `json:"couple"`            // 这对 CP 的名字，例如「梁山伯与祝英台」
	Tagline string `json:"tagline"`           // 一句话点出关系气质
	Reason  string `json:"reason"`            // 1–2 句大白话，扣住报告里已有的信号
}
```

- [ ] **Step 4: 运行测试，确认通过**

Run: `cd backend && go test ./internal/model/ -run FamousCouple -v`
Expected: PASS（两个测试都过）。

- [ ] **Step 5: 提交**

```bash
cd /Users/liujiming/web/yuanju
git add backend/internal/model/compatibility.go backend/internal/model/compatibility_famous_couple_test.go
git commit -m "feat(compat): add famous_couple field to structured report struct"
```

---

## Task 2: canonical prompt 新增 famous_couple schema 与约束

**Files:**
- Modify: `backend/pkg/prompt/canonical_compatibility.go`
- Test: `backend/pkg/prompt/canonical_compatibility_famous_couple_test.go`（新建）

- [ ] **Step 1: 写失败测试**

新建 `backend/pkg/prompt/canonical_compatibility_famous_couple_test.go`：

```go
package prompt

import (
	"strings"
	"testing"
)

func TestCompatibilityPromptIncludesFamousCouple(t *testing.T) {
	def := MustGet("compatibility")
	// schema 字段
	for _, want := range []string{
		`"famous_couple"`,
		`"couple"`,
		`"tagline"`,
		`"reason"`,
	} {
		if !strings.Contains(def.Content, want) {
			t.Errorf("compatibility prompt missing famous_couple schema token %q", want)
		}
	}
	// 约束段：必须反映真实动态、可悲剧、得体不越线
	for _, want := range []string{
		"名人类比约束",
		"真实动态",
	} {
		if !strings.Contains(def.Content, want) {
			t.Errorf("compatibility prompt missing famous_couple constraint %q", want)
		}
	}
}
```

- [ ] **Step 2: 运行测试，确认失败**

Run: `cd backend && go test ./pkg/prompt/ -run FamousCouple -v`
Expected: FAIL —— prompt 内容还不含 `famous_couple` / `名人类比约束`。

- [ ] **Step 3: 修改 prompt 内容（最小实现）**

3a. 在 `backend/pkg/prompt/canonical_compatibility.go` 的「表达约束（面向普通用户，务必遵守）」段落**之后**、「关系经营策略·沟通」段落**之前**，插入约束段：

```
名人类比约束（famous_couple，务必遵守）：
- 给这对关系挑一对广为人知的名人/经典情侣（真实或传说皆可，自由联想），用来类比「你们这对」的气质。
- 必须反映这对关系的真实动态（综合分、四模块分、缘分时长、性格 fit/clash）：数据偏负向时可以是苦情/悲剧 CP（如梁祝、牛郎织女），不要一律浪漫圆满。
- couple 给名字；tagline 一句话点出关系气质；reason 1–2 句大白话，落到「你们 / 相处」，引用报告里已有的具体信号，用条件语气、不下绝对命运断语。
- 必须得体、不越线、不出现不合适或冒犯性的配对。
```

3b. 在输出 JSON schema 里，`"summary"` 行之后插入 `famous_couple` 段（紧跟 summary，与结构体字段顺序一致）：

```
  "summary": "总体判断，必须基于输入证据，不使用绝对断语",
  "famous_couple": {
    "couple": "这对关系最贴切的名人/经典情侣名字，例如：梁山伯与祝英台",
    "tagline": "一句话点出关系气质，例如：一见倾心，却被现实层层阻隔",
    "reason": "1–2 句大白话，扣住报告里已有的信号，说清为什么是这对"
  },
```

- [ ] **Step 4: 运行测试，确认通过**

Run: `cd backend && go test ./pkg/prompt/ -run FamousCouple -v`
Expected: PASS。

- [ ] **Step 5: 跑 prompt 包全量回归（确认没破坏 drift / 既有断言）**

Run: `cd backend && go test ./pkg/prompt/...`
Expected: ok。`canonical_test.go` 的 version 断言仍通过（version 未改，仅追加内容；drift_test 用动态 hash，不受影响）。

- [ ] **Step 6: 提交**

```bash
cd /Users/liujiming/web/yuanju
git add backend/pkg/prompt/canonical_compatibility.go backend/pkg/prompt/canonical_compatibility_famous_couple_test.go
git commit -m "feat(compat): add famous_couple schema and constraints to compatibility prompt"
```

---

## Task 3: 前端 TS 类型新增 famous_couple

**Files:**
- Modify: `frontend/src/lib/api.ts:445-458`

- [ ] **Step 1: 修改接口**

在 `frontend/src/lib/api.ts` 的 `CompatibilityStructuredReport` 接口里，`summary` 行之后加一行可选字段，并在该接口**上方**新增类型：

```ts
export interface CompatibilityFamousCouple {
  couple: string
  tagline: string
  reason: string
}

export interface CompatibilityStructuredReport {
  summary: string
  famous_couple?: CompatibilityFamousCouple | null
  question_focus?: CompatibilityQuestionFocus
  personality_comparison?: CompatibilityPersonalityComparison | null
  dimensions: Array<{ key: string; title: string; content: string }>
  duration_assessment: CompatibilityDurationAssessment
  relationship_diagnosis?: CompatibilityRelationshipDiagnosis
  decision_advice?: CompatibilityDecisionAdvice
  stage_risks?: CompatibilityStageRisk[]
  relationship_strategy?: CompatibilityRelationshipStrategy
  claim_evidence_links?: CompatibilityClaimEvidenceLink[]
  risks: string[]
  advice: string
}
```

- [ ] **Step 2: 类型检查通过**

Run: `cd frontend && npx tsc -b`
Expected: 无错误（仅新增可选字段，不影响既有用法）。

- [ ] **Step 3: 提交**

```bash
cd /Users/liujiming/web/yuanju
git add frontend/src/lib/api.ts
git commit -m "feat(compat): add famous_couple to CompatibilityStructuredReport type"
```

---

## Task 4: 前端 FamousCoupleCard 组件（三态）

**Files:**
- Create: `frontend/src/components/compatibility/FamousCoupleCard.tsx`
- Create: `frontend/src/components/compatibility/FamousCoupleCard.css`
- Test: `frontend/tests/compat-famous-couple-card.test.mjs`（新建）

- [ ] **Step 1: 写失败测试（源码断言，匹配现有 .mjs 风格）**

新建 `frontend/tests/compat-famous-couple-card.test.mjs`：

```js
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'
const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (p) => readFileSync(resolve(root, p), 'utf8')

test('FamousCoupleCard component exists with three states', () => {
  const src = read('src/components/compatibility/FamousCoupleCard.tsx')
  assert.match(src, /export default function FamousCoupleCard/)
  // 已生成且有类比：展示 couple / tagline / reason
  assert.match(src, /famous_couple/)
  assert.match(src, /\.couple/)
  assert.match(src, /\.tagline/)
  assert.match(src, /\.reason/)
  // 未生成：钩子文案 + 生成入口
  assert.match(src, /揭晓你们的名人配对/)
  assert.match(src, /onGenerateReport/)
})

test('FamousCoupleCard hides on legacy report without famous_couple', () => {
  const src = read('src/components/compatibility/FamousCoupleCard.tsx')
  // 已有报告但没有 famous_couple 时返回 null（隐藏，不误导）
  assert.match(src, /return null/)
})
```

- [ ] **Step 2: 运行测试，确认失败**

Run: `cd frontend && node --test tests/compat-famous-couple-card.test.mjs`
Expected: FAIL —— 文件不存在 / ENOENT。

- [ ] **Step 3: 写组件**

新建 `frontend/src/components/compatibility/FamousCoupleCard.tsx`：

```tsx
import type { CompatibilityStructuredReport } from '../../lib/api'
import './FamousCoupleCard.css'

type Props = {
  hasReport: boolean
  famousCouple: CompatibilityStructuredReport['famous_couple']
  reportLoading: boolean
  onGenerateReport: () => void
}

export default function FamousCoupleCard({ hasReport, famousCouple, reportLoading, onGenerateReport }: Props) {
  // 状态三：已有报告但旧版没有 famous_couple → 隐藏，不误导
  if (hasReport && !famousCouple) {
    return null
  }

  // 状态一：还没生成深度解读 → 钩子占位 + 生成入口
  if (!hasReport) {
    return (
      <div className="famous-couple-card famous-couple-card--teaser">
        <div className="famous-couple-card__teaser-text">✨ 生成深度解读，揭晓你们的名人配对</div>
        <button
          type="button"
          className="btn btn-primary famous-couple-card__cta"
          onClick={onGenerateReport}
          disabled={reportLoading}
        >
          {reportLoading ? '生成中' : '生成深度解读'}
        </button>
      </div>
    )
  }

  // 状态二：已生成且有类比
  return (
    <div className="famous-couple-card famous-couple-card--filled">
      <div className="famous-couple-card__kicker">你们这对，像</div>
      <div className="serif famous-couple-card__couple">{famousCouple!.couple}</div>
      {famousCouple!.tagline && <div className="famous-couple-card__tagline">{famousCouple!.tagline}</div>}
      {famousCouple!.reason && <p className="famous-couple-card__reason">{famousCouple!.reason}</p>}
    </div>
  )
}
```

新建 `frontend/src/components/compatibility/FamousCoupleCard.css`（克制、与结果页其它卡片视觉一致；具体变量名沿用项目既有 token，若不存在则用字面值）：

```css
.famous-couple-card {
  border-radius: 16px;
  padding: 20px 24px;
  margin: 16px 0;
  text-align: center;
  background: linear-gradient(135deg, rgba(180, 120, 200, 0.12), rgba(120, 140, 210, 0.12));
  border: 1px solid rgba(150, 130, 200, 0.25);
}

.famous-couple-card--teaser {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

.famous-couple-card__teaser-text {
  font-size: 15px;
  opacity: 0.85;
}

.famous-couple-card__kicker {
  font-size: 13px;
  letter-spacing: 0.1em;
  opacity: 0.7;
  margin-bottom: 6px;
}

.famous-couple-card__couple {
  font-size: 26px;
  font-weight: 700;
  line-height: 1.3;
}

.famous-couple-card__tagline {
  font-size: 15px;
  margin-top: 6px;
  opacity: 0.9;
}

.famous-couple-card__reason {
  font-size: 14px;
  line-height: 1.7;
  margin-top: 12px;
  opacity: 0.85;
}
```

- [ ] **Step 4: 运行测试，确认通过**

Run: `cd frontend && node --test tests/compat-famous-couple-card.test.mjs`
Expected: PASS（两个测试都过）。

- [ ] **Step 5: 提交**

```bash
cd /Users/liujiming/web/yuanju
git add frontend/src/components/compatibility/FamousCoupleCard.tsx frontend/src/components/compatibility/FamousCoupleCard.css frontend/tests/compat-famous-couple-card.test.mjs
git commit -m "feat(compat): add FamousCoupleCard component with three states"
```

---

## Task 5: 在结果页顶部挂载 FamousCoupleCard

**Files:**
- Modify: `frontend/src/pages/CompatibilityResultPage.tsx`（import 区 + 渲染区 `CompatibilityStickyHeader` 之后）
- Test: `frontend/tests/compat-result-famous-couple.test.mjs`（新建）

- [ ] **Step 1: 写失败测试**

新建 `frontend/tests/compat-result-famous-couple.test.mjs`：

```js
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'
const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (p) => readFileSync(resolve(root, p), 'utf8')

test('result page mounts FamousCoupleCard near the top', () => {
  const src = read('src/pages/CompatibilityResultPage.tsx')
  assert.match(src, /import FamousCoupleCard from '..\/components\/compatibility\/FamousCoupleCard'/)
  assert.match(src, /<FamousCoupleCard/)
  assert.match(src, /famousCouple=\{structuredReport\?\.famous_couple\}/)
  assert.match(src, /onGenerateReport=\{handleGenerateReport\}/)
  // 顶部：必须在 SectionVerdict 之前出现
  const coupleIdx = src.indexOf('<FamousCoupleCard')
  const verdictIdx = src.indexOf('<SectionVerdict')
  assert.ok(coupleIdx > -1 && verdictIdx > -1 && coupleIdx < verdictIdx, 'FamousCoupleCard must render before SectionVerdict')
})
```

- [ ] **Step 2: 运行测试，确认失败**

Run: `cd frontend && node --test tests/compat-result-famous-couple.test.mjs`
Expected: FAIL —— 页面还没 import / 渲染 FamousCoupleCard。

- [ ] **Step 3: 修改页面**

3a. 在 `frontend/src/pages/CompatibilityResultPage.tsx` import 区，`SectionDeepAnalysis` import 行附近加：

```tsx
import FamousCoupleCard from '../components/compatibility/FamousCoupleCard'
```

3b. 在渲染区，`<CompatibilityStickyHeader ... />` 之后、`<SectionBasicCharts ... />` 之前插入（顶部、醒目）：

```tsx
        <CompatibilityStickyHeader
          selfName={selfP?.display_name || '我'}
          partnerName={partnerP?.display_name || '对方'}
          overallScore={reading.overall_score}
          verdict={decisionDashboard.verdict}
        />
        <FamousCoupleCard
          hasReport={Boolean(detail.latest_report)}
          famousCouple={structuredReport?.famous_couple}
          reportLoading={reportLoading}
          onGenerateReport={handleGenerateReport}
        />
        <SectionBasicCharts self={selfP || null} partner={partnerP || null} />
```

注：`structuredReport`、`reportLoading`、`handleGenerateReport`、`detail` 均已在该组件作用域内定义（见现有 `CompatibilityResultPage.tsx:269 / 110 / 158 / 108`），无需新增 state。

- [ ] **Step 4: 运行测试 + 类型检查，确认通过**

Run: `cd frontend && node --test tests/compat-result-famous-couple.test.mjs && npx tsc -b`
Expected: 测试 PASS；tsc 无错误。

- [ ] **Step 5: 提交**

```bash
cd /Users/liujiming/web/yuanju
git add frontend/src/pages/CompatibilityResultPage.tsx frontend/tests/compat-result-famous-couple.test.mjs
git commit -m "feat(compat): mount FamousCoupleCard at top of result page"
```

---

## Task 6: 全量验证

- [ ] **Step 1: 后端全量测试**

Run: `cd backend && go build ./... && go test ./...`
Expected: build 无错误；所有包测试 ok（特别是 `internal/model`、`pkg/prompt`、`internal/service`、`internal/handler`）。

- [ ] **Step 2: 前端全量测试 + 构建**

Run: `cd frontend && npm run lint && npx tsc -b && node --test tests/*.test.mjs`
Expected: lint 通过；tsc 无错误；所有 .mjs 测试 PASS。

- [ ] **Step 3: 人工冒烟（可选但推荐）**

1. 启动应用，做一次合盘 → 顶部应显示「✨ 生成深度解读，揭晓你们的名人配对」钩子卡。
2. 点「生成深度解读」→ 顶部钩子卡替换为名人类比卡（couple + tagline + reason），reason 扣住该盘信号。
3. 刷新页面 → 类比卡内容不变（随报告持久化）。
4. （已有部署）若钩子卡生成后类比为空：检查后台「AI Prompt」compatibility 是否已更新到最新出厂版（见顶部「运维须知」）。

- [ ] **Step 4: 无新增提交（本任务仅验证）；如人工冒烟发现问题，回到对应 Task 修复**

---

## Self-Review 记录

- **Spec 覆盖**：§3 数据结构 → Task 1（Go）/ Task 3（TS）；§4 prompt 约束 → Task 2；§5 三态前端卡 → Task 4 + Task 5；§6 改动清单全覆盖；§7 成功标准 → Task 6 验证。分享图/PDF（§8 范围外）未建任务 ✓。
- **Placeholder 扫描**：无 TBD/TODO；所有代码步骤给出完整代码。
- **类型一致性**：`CompatibilityFamousCouple`（Go `Couple/Tagline/Reason` ↔ JSON/TS `couple/tagline/reason`）、组件 props（`hasReport/famousCouple/reportLoading/onGenerateReport`）在 Task 4 定义、Task 5 调用一致；`famous_couple` 字段名在 Go tag / TS / prompt schema / 测试断言中统一。
