# Compatibility Duration Assessment Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extend the existing compatibility feature so each reading also returns and renders a staged relationship-duration assessment for `3个月`、`1年`、`2年以上`.

**Architecture:** Keep the existing two-layer compatibility flow intact: `bazi.Calculate()` still produces two natal charts, `AnalyzeCompatibility()` still owns the static four-dimension result, and a new duration layer derives staged maintenance potential from the static signals plus lightweight time-window heuristics. Persist the new structure under the compatibility reading resource, thread it through the report prompt context, and render it in the result page before the AI narrative.

**Tech Stack:** Go 1.25, Gin, PostgreSQL JSONB, React 19 + TypeScript + Vite, existing compatibility repository/service/page structure.

---

### Task 1: Add failing backend tests for duration assessment

**Files:**
- Modify: `backend/pkg/bazi/compatibility_test.go`
- Test: `backend/pkg/bazi/compatibility_test.go`

- [ ] **Step 1: Write failing duration tests**

Add tests that lock down the new contract before implementation:

```go
func TestAnalyzeCompatibility_ReturnsDurationAssessmentShape(t *testing.T) {
	a := makeCompatNatal("甲子", "丙寅", "甲子", "丁卯", "male")
	b := makeCompatNatal("己丑", "戊辰", "己丑", "庚申", "female")

	got := AnalyzeCompatibility(a, b)

	if got.DurationAssessment.OverallBand == "" {
		t.Fatal("expected duration overall band")
	}
	if got.DurationAssessment.Windows.ThreeMonths.Level == "" {
		t.Fatal("expected three-month window")
	}
	if got.DurationAssessment.Windows.OneYear.Level == "" {
		t.Fatal("expected one-year window")
	}
	if got.DurationAssessment.Windows.TwoYearsPlus.Level == "" {
		t.Fatal("expected two-years-plus window")
	}
	if len(got.DurationAssessment.Reasons) == 0 {
		t.Fatal("expected duration reasons")
	}
}

func TestAnalyzeCompatibility_DurationCanBeStrongShortTermButWeakLongTerm(t *testing.T) {
	a := makeCompatNatal("甲子", "丙寅", "甲子", "丁卯", "male")
	b := makeCompatNatal("庚午", "壬午", "辛午", "庚申", "female")

	got := AnalyzeCompatibility(a, b)

	if got.DurationAssessment.Windows.ThreeMonths.Level == "" || got.DurationAssessment.Windows.TwoYearsPlus.Level == "" {
		t.Fatal("expected both short and long windows")
	}
	if got.DurationAssessment.Windows.ThreeMonths.Level == got.DurationAssessment.Windows.TwoYearsPlus.Level {
		t.Fatalf("expected staged difference, got %+v", got.DurationAssessment.Windows)
	}
}
```

- [ ] **Step 2: Run the backend compatibility tests and verify RED**

Run:

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./pkg/bazi -run Compatibility -v
```

Expected: FAIL because `CompatibilityAnalysis` has no `DurationAssessment` fields yet.

- [ ] **Step 3: Record the expected public shape**

Before implementation, confirm these names are the source of truth:

```go
type CompatibilityDurationAssessment struct {
	OverallBand string
	Windows     CompatibilityDurationWindows
	Summary     string
	Reasons     []string
}

type CompatibilityDurationWindows struct {
	ThreeMonths  CompatibilityDurationWindow
	OneYear      CompatibilityDurationWindow
	TwoYearsPlus CompatibilityDurationWindow
}

type CompatibilityDurationWindow struct {
	Level string
}
```

- [ ] **Step 4: Keep the tests failing until the production shape exists**

Do not edit assertions to match current code. The production code must grow to satisfy these names.

### Task 2: Implement duration assessment in the compatibility engine

**Files:**
- Modify: `backend/pkg/bazi/compatibility.go`
- Modify: `backend/pkg/bazi/compatibility_test.go`
- Test: `backend/pkg/bazi/compatibility_test.go`

- [ ] **Step 1: Add the duration types to the compatibility analysis result**

Extend the existing model in `compatibility.go`:

```go
type CompatibilityDurationLevel string

const (
	CompatibilityDurationHigh   CompatibilityDurationLevel = "high"
	CompatibilityDurationMedium CompatibilityDurationLevel = "medium"
	CompatibilityDurationLow    CompatibilityDurationLevel = "low"
)

type CompatibilityDurationWindow struct {
	Level CompatibilityDurationLevel `json:"level"`
}

type CompatibilityDurationWindows struct {
	ThreeMonths  CompatibilityDurationWindow `json:"three_months"`
	OneYear      CompatibilityDurationWindow `json:"one_year"`
	TwoYearsPlus CompatibilityDurationWindow `json:"two_years_plus"`
}

type CompatibilityDurationAssessment struct {
	OverallBand string                       `json:"overall_band"`
	Windows     CompatibilityDurationWindows `json:"windows"`
	Summary     string                       `json:"summary"`
	Reasons     []string                     `json:"reasons"`
}
```

- [ ] **Step 2: Implement the minimal duration scoring heuristic**

Add a helper that derives staged levels from existing scores and evidence polarity:

```go
func buildDurationAssessment(scores CompatibilityDimensionScores, evidences []CompatibilityEvidence) CompatibilityDurationAssessment {
	shortScore := scores.Attraction + scores.Communication/2
	midScore := scores.Stability + scores.Communication/2 + scores.Practicality/2
	longScore := scores.Stability + scores.Practicality

	for _, item := range evidences {
		switch item.Source {
		case "spouse_palace":
			shortScore += item.Weight / 2
			midScore += item.Weight
			longScore += item.Weight
		case "spouse_star", "shensha":
			shortScore += item.Weight
			midScore += item.Weight / 2
		case "ganzhi", "five_elements":
			midScore += item.Weight / 2
			longScore += item.Weight
		}
	}

	return CompatibilityDurationAssessment{
		OverallBand: durationBand(longScore),
		Windows: CompatibilityDurationWindows{
			ThreeMonths:  CompatibilityDurationWindow{Level: durationLevel(shortScore)},
			OneYear:      CompatibilityDurationWindow{Level: durationLevel(midScore)},
			TwoYearsPlus: CompatibilityDurationWindow{Level: durationLevel(longScore)},
		},
		Summary: durationSummary(shortScore, midScore, longScore),
		Reasons: durationReasons(evidences),
	}
}
```

- [ ] **Step 3: Wire duration assessment into `AnalyzeCompatibility()`**

After static scores and tags are finalized:

```go
duration := buildDurationAssessment(scores, evidences)

return CompatibilityAnalysis{
	OverallLevel:       level,
	DimensionScores:    scores,
	Evidences:          evidences,
	SummaryTags:        tags,
	DurationAssessment: duration,
}
```

- [ ] **Step 4: Run the compatibility tests and verify GREEN**

Run:

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./pkg/bazi -run Compatibility -v
```

Expected: PASS, including the new staged-duration tests.

- [ ] **Step 5: Refactor only if the tests stay green**

If helper extraction is needed, keep it local to `compatibility.go` and do not change the public JSON field names introduced above.

### Task 3: Persist and expose duration assessment through model, repository, service, and report context

**Files:**
- Modify: `backend/internal/model/compatibility.go`
- Modify: `backend/internal/repository/compatibility_repository.go`
- Modify: `backend/internal/service/compatibility_service.go`
- Modify: `backend/internal/handler/compatibility_handler.go`
- Modify: `backend/pkg/database/database.go`
- Test: `backend/internal/service/compatibility_service_test.go` (create if missing)

- [ ] **Step 1: Write a failing service-level test for the persisted shape**

Add a focused test that creates a reading and expects `duration_assessment` in the detail payload:

```go
func TestCreateCompatibilityReading_PopulatesDurationAssessment(t *testing.T) {
	detail, err := CreateCompatibilityReading(userID, selfProfile, partnerProfile)
	if err != nil {
		t.Fatal(err)
	}
	if detail.Reading.DurationAssessment.OverallBand == "" {
		t.Fatal("expected duration assessment on reading")
	}
}
```

- [ ] **Step 2: Add the new persisted field to the reading model**

Extend `CompatibilityReading`:

```go
type CompatibilityReading struct {
	// ...
	DurationAssessment CompatibilityDurationAssessment `json:"duration_assessment"`
}
```

Mirror the same structure in the prompt/report types:

```go
type CompatibilityStructuredReport struct {
	Summary    string                            `json:"summary"`
	Dimensions []CompatibilityDimensionNarrative `json:"dimensions"`
	Duration   CompatibilityDurationAssessment   `json:"duration_assessment"`
	Risks      []string                          `json:"risks"`
	Advice     string                            `json:"advice"`
}
```

- [ ] **Step 3: Persist `duration_assessment` as JSONB on `compatibility_readings`**

If the table is missing the column, add a safe migration in `backend/pkg/database/database.go`:

```sql
ALTER TABLE compatibility_readings
ADD COLUMN IF NOT EXISTS duration_assessment JSONB NOT NULL DEFAULT '{}'::jsonb;
```

Update repository create/read paths:

```go
durationJSON, _ := json.Marshal(durationAssessment)

INSERT INTO compatibility_readings (..., duration_assessment, ...)
VALUES (..., $N, ...)
RETURNING ..., duration_assessment, ...
```

and on scans:

```go
var rawDuration []byte
_ = json.Unmarshal(rawDuration, &r.DurationAssessment)
```

- [ ] **Step 4: Pass duration assessment from engine -> repository -> detail**

Update `CreateCompatibilityReading()` to send:

```go
reading, err := repository.CreateCompatibilityReading(
	userID,
	string(analysis.OverallLevel),
	model.CompatibilityDimensionScores{...},
	analysis.DurationAssessment,
	analysis.SummaryTags,
	compatibilityAnalysisVersion,
)
```

- [ ] **Step 5: Add duration context to the compatibility prompt fallback**

Extend prompt data:

```go
type CompatibilityPromptData struct {
	// ...
	DurationJSON string
}
```

and include it in the fallback prompt:

```text
缘分时长评估（JSON）：
{{.DurationJSON}}
```

Require the JSON report output to include:

```json
"duration_assessment": {
  "summary": "...",
  "windows": {
    "three_months": { "level": "high" },
    "one_year": { "level": "medium" },
    "two_years_plus": { "level": "low" }
  }
}
```

- [ ] **Step 6: Run backend verification**

Run:

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./...
```

Expected: PASS with the new duration field flowing through the service/repository boundary.

### Task 4: Render duration assessment on the compatibility result page

**Files:**
- Modify: `frontend/src/lib/api.ts`
- Modify: `frontend/src/pages/CompatibilityResultPage.tsx`
- Modify: `frontend/src/pages/CompatibilityResultPage.css`
- Test: `frontend/src/pages/CompatibilityResultPage.test.tsx` (create only if a frontend test harness is added)

- [ ] **Step 1: Extend the frontend compatibility types**

Add matching interfaces in `api.ts`:

```ts
export interface CompatibilityDurationWindow {
  level: 'high' | 'medium' | 'low'
}

export interface CompatibilityDurationAssessment {
  overall_band: 'short_term' | 'medium_term' | 'long_term'
  summary: string
  reasons: string[]
  windows: {
    three_months: CompatibilityDurationWindow
    one_year: CompatibilityDurationWindow
    two_years_plus: CompatibilityDurationWindow
  }
}
```

and add it under `CompatibilityReading` plus `CompatibilityStructuredReport`.

- [ ] **Step 2: Render a dedicated duration section before evidences**

Insert a new section in `CompatibilityResultPage.tsx` between the score grid and evidence list:

```tsx
<div className="compatibility-section">
  <div className="compatibility-section-header">
    <h2 className="serif compatibility-section-title">缘分时长评估</h2>
    <p className="compatibility-section-desc">不是判断某一天结束，而是看这段关系在不同阶段的维持潜力。</p>
  </div>
  <div className="compatibility-duration-grid">
    {[
      ['3个月', reading.duration_assessment.windows.three_months.level],
      ['1年', reading.duration_assessment.windows.one_year.level],
      ['2年以上', reading.duration_assessment.windows.two_years_plus.level],
    ].map(([label, level]) => (
      <div key={label} className="card compatibility-duration-card">
        <div className="compatibility-duration-label">{label}</div>
        <div className="serif compatibility-duration-value">{durationLevelText[level as string]}</div>
      </div>
    ))}
  </div>
  <p className="compatibility-duration-summary">
    {detail.latest_report?.content_structured?.duration_assessment?.summary || reading.duration_assessment.summary}
  </p>
</div>
```

- [ ] **Step 3: Add lightweight styling for the new block**

Add CSS only for the new section:

```css
.compatibility-duration-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.compatibility-duration-card {
  padding: 18px;
}

.compatibility-duration-label {
  font-size: 13px;
  color: var(--text-muted);
  margin-bottom: 6px;
}

.compatibility-duration-value {
  font-size: 24px;
}

.compatibility-duration-summary {
  margin: 12px 0 0;
  color: var(--text-secondary);
  line-height: 1.8;
}
```

- [ ] **Step 4: Verify the frontend build**

Run:

```bash
cd /Users/liujiming/web/yuanju/frontend && npm run build
```

Expected: PASS. If TypeScript reports missing `duration_assessment`, fix the typing surface rather than using `any`.

### Task 5: Update OpenSpec task tracking and final verification

**Files:**
- Modify: `openspec/changes/bazi-compatibility-match/tasks.md`

- [ ] **Step 1: Mark completed duration tasks as they land**

Flip these items only after the corresponding code and tests are green:

```md
- [x] 3.8 ...
- [x] 3.9 ...
- [x] 3.10 ...
- [x] 4.6 ...
- [x] 4.7 ...
- [x] 6.6 ...
- [x] 6.7 ...
- [x] 7.6 ...
```

- [ ] **Step 2: Run focused verification before broad verification**

Run:

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./pkg/bazi -run Compatibility -v
cd /Users/liujiming/web/yuanju/frontend && npm run build
```

Expected: both succeed.

- [ ] **Step 3: Run broad verification**

Run:

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./...
```

Expected: PASS for the whole backend test suite.

- [ ] **Step 4: Summarize residual risk honestly**

If frontend interaction tests are still absent, record that as a gap instead of claiming full test coverage.
