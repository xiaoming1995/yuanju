# Compatibility Consulting Report Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Upgrade compatibility readings from a score/result page into a structured consulting-style relationship report with diagnosis, decision advice, stage risks, strategy, and evidence-linked professional rationale.

**Architecture:** Extend the existing compatibility pipeline instead of replacing it. The bazi compatibility engine produces deterministic consulting assessment data and stable evidence keys; repository/service persist and expose it; the AI prompt can enrich the same JSON shape; the frontend renders a conclusion-first consulting page with evidence expansion and professional details.

**Tech Stack:** Go + Gin + PostgreSQL JSONB + `lib/pq`; React 18 + TypeScript + Vite; existing CSS variables; existing `node --test` frontend static tests and Go unit tests.

---

## File Structure

- Modify `backend/pkg/bazi/compatibility.go`: add stable evidence keys, consulting assessment structs, relationship diagnosis, decision advice, stage risk, strategy, and claim-evidence links.
- Modify `backend/pkg/bazi/compatibility_test.go`: cover consulting shape, evidence key uniqueness, and claim links.
- Create `backend/pkg/database/migrations/00007_add_compatibility_consulting.sql`: add `consulting_assessment JSONB` and `evidence_key TEXT`.
- Modify `backend/internal/model/compatibility.go`: add API/persistence types for consulting assessment and evidence key.
- Modify `backend/internal/repository/compatibility_repository.go`: write/read `consulting_assessment` and `evidence_key`; keep legacy rows safe.
- Modify `backend/internal/service/compatibility_service.go`: map bazi consulting data into model data, lazy-backfill old readings, include consulting JSON in prompt data, and parse new AI report shape.
- Modify `backend/internal/service/compatibility_service_test.go`: cover prompt data and lazy backfill.
- Modify `backend/internal/handler/compatibility_handler_test.go`: verify detail response includes consulting assessment and evidence keys.
- Modify `frontend/src/lib/api.ts`: add TypeScript types for consulting assessment and structured report extensions.
- Modify `frontend/src/pages/CompatibilityResultPage.tsx`: render diagnosis, decision advice, stage risks, strategy, and evidence-linked rationale ahead of existing score/professional sections.
- Modify `frontend/src/pages/CompatibilityResultPage.css`: add responsive consulting module styles.
- Modify `frontend/tests/compatibility-result-ux.test.mjs`: assert consulting sections, evidence expansion, and mobile-safe layout markers.

## Task 1: Deterministic Consulting Assessment in Bazi Engine

**Files:**
- Modify: `backend/pkg/bazi/compatibility.go`
- Test: `backend/pkg/bazi/compatibility_test.go`

- [ ] **Step 1: Write failing tests for consulting assessment**

Append these tests to `backend/pkg/bazi/compatibility_test.go`:

```go
func TestAnalyzeCompatibility_ReturnsConsultingAssessment(t *testing.T) {
	a := makeCompatNatal("甲子", "丙寅", "甲子", "丁卯", "male")
	b := makeCompatNatal("己午", "戊午", "己午", "戊申", "female")

	got := AnalyzeCompatibility(a, b)

	if got.ConsultingAssessment.RelationshipDiagnosis.RelationshipType == "" {
		t.Fatal("expected relationship type")
	}
	if got.ConsultingAssessment.RelationshipDiagnosis.Verdict == "" {
		t.Fatal("expected verdict")
	}
	if len(got.ConsultingAssessment.RelationshipDiagnosis.TopFindings) == 0 {
		t.Fatal("expected top findings")
	}
	if got.ConsultingAssessment.DecisionAdvice.Recommendation == "" {
		t.Fatal("expected decision recommendation")
	}
	if len(got.ConsultingAssessment.StageRisks) != 3 {
		t.Fatalf("expected three stage risks, got %d", len(got.ConsultingAssessment.StageRisks))
	}
	if got.ConsultingAssessment.RelationshipStrategy.Communication == "" {
		t.Fatal("expected communication strategy")
	}
	if len(got.ConsultingAssessment.ClaimEvidenceLinks) == 0 {
		t.Fatal("expected claim evidence links")
	}
}

func TestAnalyzeCompatibility_EvidenceKeysAreStableAndLinked(t *testing.T) {
	a := makeCompatNatal("甲子", "丙寅", "甲子", "丁卯", "male")
	b := makeCompatNatal("己午", "戊午", "己午", "戊申", "female")

	got := AnalyzeCompatibility(a, b)
	keys := map[string]bool{}
	for _, evidence := range got.Evidences {
		if evidence.EvidenceKey == "" {
			t.Fatalf("expected evidence key for %+v", evidence)
		}
		if keys[evidence.EvidenceKey] {
			t.Fatalf("duplicate evidence key %q", evidence.EvidenceKey)
		}
		keys[evidence.EvidenceKey] = true
	}
	for _, link := range got.ConsultingAssessment.ClaimEvidenceLinks {
		if link.ClaimID == "" || link.Claim == "" || link.Reasoning == "" {
			t.Fatalf("expected complete claim link, got %+v", link)
		}
		for _, key := range link.EvidenceKeys {
			if !keys[key] {
				t.Fatalf("claim link references missing evidence key %q", key)
			}
		}
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run:

```bash
go test ./pkg/bazi -run 'TestAnalyzeCompatibility_(ReturnsConsultingAssessment|EvidenceKeysAreStableAndLinked)'
```

Expected: FAIL with missing `ConsultingAssessment` and `EvidenceKey` fields.

- [ ] **Step 3: Add bazi consulting types and fields**

In `backend/pkg/bazi/compatibility.go`, extend `CompatibilityEvidence` and `CompatibilityAnalysis`, and add these types near the existing compatibility structs:

```go
type CompatibilityEvidence struct {
	EvidenceKey string                 `json:"evidence_key"`
	Dimension   CompatibilityDimension `json:"dimension"`
	Type        string                 `json:"type"`
	Polarity    CompatibilityPolarity  `json:"polarity"`
	Source      string                 `json:"source"`
	Title       string                 `json:"title"`
	Detail      string                 `json:"detail"`
	Weight      int                    `json:"weight"`
}

type CompatibilityFinding struct {
	Text         string   `json:"text"`
	EvidenceKeys []string `json:"evidence_keys"`
}

type CompatibilityRelationshipDiagnosis struct {
	RelationshipType string                 `json:"relationship_type"`
	Verdict          string                 `json:"verdict"`
	Summary          string                 `json:"summary"`
	TopFindings      []CompatibilityFinding `json:"top_findings"`
}

type CompatibilityDecisionAdvice struct {
	Recommendation string   `json:"recommendation"`
	Confidence     string   `json:"confidence"`
	Conditions     []string `json:"conditions"`
	DoNext         []string `json:"do_next"`
	Avoid          []string `json:"avoid"`
}

type CompatibilityStageRisk struct {
	Window       string   `json:"window"`
	RiskLevel    string   `json:"risk_level"`
	MainRisk     string   `json:"main_risk"`
	Trigger      string   `json:"trigger"`
	Advice       string   `json:"advice"`
	EvidenceKeys []string `json:"evidence_keys"`
}

type CompatibilityRelationshipStrategy struct {
	Communication string `json:"communication"`
	Conflict      string `json:"conflict"`
	Reality       string `json:"reality"`
	Boundary      string `json:"boundary"`
}

type CompatibilityClaimEvidenceLink struct {
	ClaimID      string   `json:"claim_id"`
	Claim        string   `json:"claim"`
	EvidenceKeys []string `json:"evidence_keys"`
	Reasoning    string   `json:"reasoning"`
	Caveat       string   `json:"caveat"`
}

type CompatibilityConsultingAssessment struct {
	RelationshipDiagnosis CompatibilityRelationshipDiagnosis `json:"relationship_diagnosis"`
	DecisionAdvice        CompatibilityDecisionAdvice        `json:"decision_advice"`
	StageRisks            []CompatibilityStageRisk           `json:"stage_risks"`
	RelationshipStrategy  CompatibilityRelationshipStrategy  `json:"relationship_strategy"`
	ClaimEvidenceLinks    []CompatibilityClaimEvidenceLink   `json:"claim_evidence_links"`
}

type CompatibilityAnalysis struct {
	OverallLevel         CompatibilityLevel                `json:"overall_level"`
	DimensionScores      CompatibilityDimensionScores      `json:"dimension_scores"`
	Evidences            []CompatibilityEvidence           `json:"evidences"`
	SummaryTags          []string                          `json:"summary_tags"`
	DurationAssessment   CompatibilityDurationAssessment   `json:"duration_assessment"`
	ConsultingAssessment CompatibilityConsultingAssessment `json:"consulting_assessment"`
}
```

- [ ] **Step 4: Assign stable evidence keys and build consulting assessment**

Change `addEvidence` inside `AnalyzeCompatibility` so every evidence gets a stable key:

```go
	addEvidence := func(item CompatibilityEvidence) {
		item.EvidenceKey = buildCompatibilityEvidenceKey(item, len(evidences))
		evidences = append(evidences, item)
		switch item.Dimension {
		case CompatibilityAttraction:
			scores.Attraction = clampScore(scores.Attraction + item.Weight)
		case CompatibilityStability:
			scores.Stability = clampScore(scores.Stability + item.Weight)
		case CompatibilityCommunication:
			scores.Communication = clampScore(scores.Communication + item.Weight)
		case CompatibilityPracticality:
			scores.Practicality = clampScore(scores.Practicality + item.Weight)
		}
	}
```

After duration is built, return consulting assessment:

```go
	consulting := buildCompatibilityConsultingAssessment(scores, evidences, duration)

	return CompatibilityAnalysis{
		OverallLevel:         aggregateCompatibilityLevel(scores, evidences),
		DimensionScores:      scores,
		Evidences:            evidences,
		SummaryTags:          tags,
		DurationAssessment:   duration,
		ConsultingAssessment: consulting,
	}
```

Add these helper functions near `buildDurationAssessment`:

```go
func buildCompatibilityEvidenceKey(item CompatibilityEvidence, index int) string {
	return fmt.Sprintf("%s_%s_%s_%02d", item.Source, item.Dimension, item.Type, index)
}

func buildCompatibilityConsultingAssessment(scores CompatibilityDimensionScores, evidences []CompatibilityEvidence, duration CompatibilityDurationAssessment) CompatibilityConsultingAssessment {
	negativeKeys := topEvidenceKeys(evidences, CompatibilityNegative, 2)
	positiveKeys := topEvidenceKeys(evidences, CompatibilityPositive, 2)
	primaryKeys := append([]string{}, positiveKeys...)
	primaryKeys = append(primaryKeys, negativeKeys...)
	if len(primaryKeys) == 0 && len(evidences) > 0 {
		primaryKeys = []string{evidences[0].EvidenceKey}
	}

	relationshipType := "均衡观察型"
	if scores.Attraction >= 70 && scores.Stability < 60 {
		relationshipType = "短期吸引强、长期承压型"
	} else if scores.Stability >= 70 && scores.Practicality >= 65 {
		relationshipType = "稳定经营型"
	} else if scores.Attraction >= 70 && scores.Communication >= 65 {
		relationshipType = "高吸引互动型"
	} else if scores.Practicality < 55 || scores.Stability < 55 {
		relationshipType = "高磨合成本型"
	}

	recommendation := "observe"
	verdict := "建议谨慎观察"
	if scores.Stability >= 68 && scores.Practicality >= 62 && len(negativeKeys) == 0 {
		recommendation = "continue"
		verdict = "适合继续推进"
	} else if scores.Stability < 52 || scores.Practicality < 50 {
		recommendation = "caution"
		verdict = "不宜过早重投入"
	}

	confidence := "medium"
	if absInt(scores.Attraction-scores.Stability) >= 22 || len(evidences) >= 5 {
		confidence = "high"
	}

	topFindings := []CompatibilityFinding{
		{Text: "关系优势与风险需要分开判断，不能只看短期吸引。", EvidenceKeys: primaryKeys},
	}
	if scores.Attraction >= 70 {
		topFindings = append(topFindings, CompatibilityFinding{Text: "双方存在较明显的靠近感和吸引支点。", EvidenceKeys: positiveKeys})
	}
	if scores.Stability < 60 || scores.Practicality < 60 {
		topFindings = append(topFindings, CompatibilityFinding{Text: "长期稳定更依赖沟通规则和现实安排。", EvidenceKeys: negativeKeys})
	}
	if len(topFindings) > 3 {
		topFindings = topFindings[:3]
	}

	return CompatibilityConsultingAssessment{
		RelationshipDiagnosis: CompatibilityRelationshipDiagnosis{
			RelationshipType: relationshipType,
			Verdict:          verdict,
			Summary:          compatibilityConsultingSummary(relationshipType, recommendation),
			TopFindings:      topFindings,
		},
		DecisionAdvice: CompatibilityDecisionAdvice{
			Recommendation: recommendation,
			Confidence:     confidence,
			Conditions:     []string{"先观察冲突后的修复能力", "把现实安排和投入节奏说清楚"},
			DoNext:         []string{"用一到两个月验证沟通节奏是否稳定", "把容易争执的问题具体化处理"},
			Avoid:          []string{"用短期吸引替代长期判断", "在关系规则未稳定前过早绑定重大决定"},
		},
		StageRisks: []CompatibilityStageRisk{
			buildCompatibilityStageRisk("three_months", duration.Windows.ThreeMonths.Level, "初期热度与节奏差异", "一方推进过快或回应不稳定时", "先约定沟通频率和边界。", primaryKeys),
			buildCompatibilityStageRisk("one_year", duration.Windows.OneYear.Level, "现实磨合和冲突修复", "生活安排、承诺节奏或家庭压力进入关系时", "把分歧拆成具体事项，不用情绪判断关系本身。", primaryKeys),
			buildCompatibilityStageRisk("two_years_plus", duration.Windows.TwoYearsPlus.Level, "长期稳定和责任承接", "长期规划、责任分工和资源投入需要落地时", "建立可持续的责任分工和共同计划。", primaryKeys),
		},
		RelationshipStrategy: CompatibilityRelationshipStrategy{
			Communication: "重要议题用明确约定替代情绪试探。",
			Conflict:      "争执时先暂停升级，再回到具体事件和责任分工。",
			Reality:       "长期计划需要拆成可验证的小步骤。",
			Boundary:      "初期保留个人节奏，避免过快形成单方依赖。",
		},
		ClaimEvidenceLinks: []CompatibilityClaimEvidenceLink{
			{
				ClaimID:      "relationship_main_judgement",
				Claim:        verdict,
				EvidenceKeys: primaryKeys,
				Reasoning:    "主要判断来自吸引、稳定、沟通和现实磨合四类证据的合并结果。",
				Caveat:       "合盘表达的是关系倾向，现实选择和相处方式会改变结果表现。",
			},
		},
	}
}

func topEvidenceKeys(evidences []CompatibilityEvidence, polarity CompatibilityPolarity, limit int) []string {
	keys := []string{}
	for _, item := range evidences {
		if item.Polarity == polarity && item.EvidenceKey != "" {
			keys = append(keys, item.EvidenceKey)
			if len(keys) == limit {
				return keys
			}
		}
	}
	return keys
}

func compatibilityConsultingSummary(relationshipType, recommendation string) string {
	switch recommendation {
	case "continue":
		return "这组关系具备继续推进的基础，但仍需要把优势落到稳定沟通和现实安排中。"
	case "caution":
		return "这组关系不宜只凭短期感受快速投入，长期稳定需要先通过现实相处验证。"
	default:
		return fmt.Sprintf("%s需要边推进边观察，重点看冲突修复和现实节奏是否能对齐。", relationshipType)
	}
}

func buildCompatibilityStageRisk(window string, level CompatibilityDurationLevel, mainRisk, trigger, advice string, evidenceKeys []string) CompatibilityStageRisk {
	return CompatibilityStageRisk{
		Window:       window,
		RiskLevel:    string(level),
		MainRisk:     mainRisk,
		Trigger:      trigger,
		Advice:       advice,
		EvidenceKeys: evidenceKeys,
	}
}
```

- [ ] **Step 5: Run bazi tests**

Run:

```bash
go test ./pkg/bazi
```

Expected: PASS.

- [ ] **Step 6: Commit Task 1**

```bash
git add backend/pkg/bazi/compatibility.go backend/pkg/bazi/compatibility_test.go
git commit -m "feat(compatibility): add consulting assessment engine"
```

## Task 2: Persist Consulting Assessment and Evidence Keys

**Files:**
- Create: `backend/pkg/database/migrations/00007_add_compatibility_consulting.sql`
- Modify: `backend/internal/model/compatibility.go`
- Modify: `backend/internal/repository/compatibility_repository.go`
- Test: existing repository coverage if present; otherwise service/handler tests in later tasks verify persistence indirectly.

- [ ] **Step 1: Add migration**

Create `backend/pkg/database/migrations/00007_add_compatibility_consulting.sql`:

```sql
ALTER TABLE compatibility_readings
  ADD COLUMN IF NOT EXISTS consulting_assessment JSONB NOT NULL DEFAULT '{}'::jsonb;

ALTER TABLE compatibility_evidences
  ADD COLUMN IF NOT EXISTS evidence_key TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_compatibility_evidences_key
  ON compatibility_evidences(reading_id, evidence_key);
```

- [ ] **Step 2: Add model fields and types**

In `backend/internal/model/compatibility.go`, add the same JSON-facing consulting structs as Task 1 using `string` fields for API values, then extend existing structs:

```go
type CompatibilityFinding struct {
	Text         string   `json:"text"`
	EvidenceKeys []string `json:"evidence_keys"`
}

type CompatibilityRelationshipDiagnosis struct {
	RelationshipType string                 `json:"relationship_type"`
	Verdict          string                 `json:"verdict"`
	Summary          string                 `json:"summary"`
	TopFindings      []CompatibilityFinding `json:"top_findings"`
}

type CompatibilityDecisionAdvice struct {
	Recommendation string   `json:"recommendation"`
	Confidence     string   `json:"confidence"`
	Conditions     []string `json:"conditions"`
	DoNext         []string `json:"do_next"`
	Avoid          []string `json:"avoid"`
}

type CompatibilityStageRisk struct {
	Window       string   `json:"window"`
	RiskLevel    string   `json:"risk_level"`
	MainRisk     string   `json:"main_risk"`
	Trigger      string   `json:"trigger"`
	Advice       string   `json:"advice"`
	EvidenceKeys []string `json:"evidence_keys"`
}

type CompatibilityRelationshipStrategy struct {
	Communication string `json:"communication"`
	Conflict      string `json:"conflict"`
	Reality       string `json:"reality"`
	Boundary      string `json:"boundary"`
}

type CompatibilityClaimEvidenceLink struct {
	ClaimID      string   `json:"claim_id"`
	Claim        string   `json:"claim"`
	EvidenceKeys []string `json:"evidence_keys"`
	Reasoning    string   `json:"reasoning"`
	Caveat       string   `json:"caveat"`
}

type CompatibilityConsultingAssessment struct {
	RelationshipDiagnosis CompatibilityRelationshipDiagnosis `json:"relationship_diagnosis"`
	DecisionAdvice        CompatibilityDecisionAdvice        `json:"decision_advice"`
	StageRisks            []CompatibilityStageRisk           `json:"stage_risks"`
	RelationshipStrategy  CompatibilityRelationshipStrategy  `json:"relationship_strategy"`
	ClaimEvidenceLinks    []CompatibilityClaimEvidenceLink   `json:"claim_evidence_links"`
}
```

Add fields:

```go
type CompatibilityEvidence struct {
	ID          string    `json:"id"`
	ReadingID   string    `json:"reading_id"`
	EvidenceKey string    `json:"evidence_key"`
	Dimension   string    `json:"dimension"`
	Type        string    `json:"type"`
	Polarity    string    `json:"polarity"`
	Source      string    `json:"source"`
	Title       string    `json:"title"`
	Detail      string    `json:"detail"`
	Weight      int       `json:"weight"`
	CreatedAt   time.Time `json:"created_at"`
}

type CompatibilityReading struct {
	ID                   string                            `json:"id"`
	UserID               string                            `json:"user_id"`
	OverallLevel         string                            `json:"overall_level"`
	DimensionScores      CompatibilityDimensionScores      `json:"dimension_scores"`
	DurationAssessment   CompatibilityDurationAssessment   `json:"duration_assessment"`
	ConsultingAssessment CompatibilityConsultingAssessment `json:"consulting_assessment"`
	SummaryTags          []string                          `json:"summary_tags"`
	AnalysisVersion      string                            `json:"analysis_version"`
	CreatedAt            time.Time                         `json:"created_at"`
	UpdatedAt            time.Time                         `json:"updated_at"`
}
```

- [ ] **Step 3: Update repository write/read paths**

Change `CreateCompatibilityReading` signature:

```go
func CreateCompatibilityReading(userID, overallLevel string, scores model.CompatibilityDimensionScores, duration model.CompatibilityDurationAssessment, consulting model.CompatibilityConsultingAssessment, summaryTags []string, analysisVersion string) (*model.CompatibilityReading, error) {
	scoresJSON, _ := json.Marshal(scores)
	durationJSON, _ := json.Marshal(duration)
	consultingJSON, _ := json.Marshal(consulting)
	tagsJSON, _ := json.Marshal(summaryTags)
```

Update its SQL to include `consulting_assessment`, and scan `rawConsulting`:

```sql
INSERT INTO compatibility_readings (user_id, overall_level, dimension_scores, duration_assessment, consulting_assessment, summary_tags, analysis_version)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, user_id, overall_level, dimension_scores, duration_assessment, consulting_assessment, summary_tags, analysis_version, created_at, updated_at
```

Update scan variables and unmarshal:

```go
var rawScores, rawDuration, rawConsulting, rawTags []byte
err := database.DB.QueryRow(
	`INSERT INTO compatibility_readings (user_id, overall_level, dimension_scores, duration_assessment, consulting_assessment, summary_tags, analysis_version)
	 VALUES ($1, $2, $3, $4, $5, $6, $7)
	 RETURNING id, user_id, overall_level, dimension_scores, duration_assessment, consulting_assessment, summary_tags, analysis_version, created_at, updated_at`,
	userID, overallLevel, scoresJSON, durationJSON, consultingJSON, tagsJSON, analysisVersion,
).Scan(&r.ID, &r.UserID, &r.OverallLevel, &rawScores, &rawDuration, &rawConsulting, &rawTags, &r.AnalysisVersion, &r.CreatedAt, &r.UpdatedAt)
_ = json.Unmarshal(rawConsulting, &r.ConsultingAssessment)
```

Change `CreateCompatibilityEvidence` SQL to include `evidence_key`:

```sql
INSERT INTO compatibility_evidences (reading_id, evidence_key, dimension, type, polarity, source, title, detail, weight)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, reading_id, evidence_key, dimension, type, polarity, source, title, detail, weight, created_at
```

Update `GetCompatibilityReadingByID` and `GetCompatibilityEvidences` SELECT/SCAN statements to read the new fields.

- [ ] **Step 4: Add update helper for lazy backfill**

Add this repository function:

```go
func UpdateCompatibilityConsultingAssessment(readingID string, consulting model.CompatibilityConsultingAssessment) error {
	consultingJSON, _ := json.Marshal(consulting)
	_, err := database.DB.Exec(
		`UPDATE compatibility_readings
		 SET consulting_assessment = $2, updated_at = NOW()
		 WHERE id = $1`,
		readingID, consultingJSON,
	)
	return err
}
```

- [ ] **Step 5: Run backend compile tests for touched packages**

Run:

```bash
go test ./internal/model ./internal/repository ./pkg/database
```

Expected: PASS or “no test files” for packages without tests. If migration numbering conflicts with a newly added migration in the working tree, rename this migration to the next available number and update this plan checkbox with the chosen file name before committing.

- [ ] **Step 6: Commit Task 2**

```bash
git add backend/pkg/database/migrations/00007_add_compatibility_consulting.sql backend/internal/model/compatibility.go backend/internal/repository/compatibility_repository.go
git commit -m "feat(compatibility): persist consulting assessment"
```

## Task 3: Service Mapping, Prompt Data, and AI Structured Shape

**Files:**
- Modify: `backend/internal/service/compatibility_service.go`
- Modify: `backend/internal/service/compatibility_service_test.go`

- [ ] **Step 1: Write failing service tests**

Append these tests to `backend/internal/service/compatibility_service_test.go`:

```go
func TestBuildCompatibilityPromptData_EmbedsConsultingAssessment(t *testing.T) {
	detail := &model.CompatibilityDetail{
		Reading: &model.CompatibilityReading{
			DimensionScores: model.CompatibilityDimensionScores{Attraction: 78, Stability: 54, Communication: 66, Practicality: 48},
			DurationAssessment: model.CompatibilityDurationAssessment{
				OverallBand: "medium_term",
				Windows: model.CompatibilityDurationWindows{
					ThreeMonths: model.CompatibilityDurationWindow{Level: "high"},
					OneYear: model.CompatibilityDurationWindow{Level: "medium"},
					TwoYearsPlus: model.CompatibilityDurationWindow{Level: "low"},
				},
				Summary: "前期吸引强，但长期承压。",
				Reasons: []string{"夫妻宫冲克明显"},
			},
			ConsultingAssessment: model.CompatibilityConsultingAssessment{
				RelationshipDiagnosis: model.CompatibilityRelationshipDiagnosis{
					RelationshipType: "短期吸引强、长期承压型",
					Verdict: "建议谨慎观察",
					Summary: "先观察冲突修复能力。",
					TopFindings: []model.CompatibilityFinding{{Text: "吸引与稳定分化", EvidenceKeys: []string{"spouse_palace_clash"}}},
				},
				DecisionAdvice: model.CompatibilityDecisionAdvice{Recommendation: "observe", Confidence: "medium"},
			},
		},
		Participants: []model.CompatibilityParticipant{
			{Role: "self", DisplayName: "我", ChartSnapshot: makeCompatibilitySnapshot("我", "male")},
			{Role: "partner", DisplayName: "对方", ChartSnapshot: makeCompatibilitySnapshot("对方", "female")},
		},
	}

	got, err := buildCompatibilityPromptData(detail)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got.ConsultingJSON, `"relationship_type":"短期吸引强、长期承压型"`) {
		t.Fatalf("expected consulting json in prompt data, got %s", got.ConsultingJSON)
	}
}

func TestEnsureCompatibilityConsultingAssessment_BackfillsMissingConsulting(t *testing.T) {
	detail := &model.CompatibilityDetail{
		Reading: &model.CompatibilityReading{
			DimensionScores: model.CompatibilityDimensionScores{Attraction: 72, Stability: 58, Communication: 61, Practicality: 55},
		},
		Participants: []model.CompatibilityParticipant{
			{Role: "self", DisplayName: "我", ChartSnapshot: makeCompatibilitySnapshot("我", "male")},
			{Role: "partner", DisplayName: "对方", ChartSnapshot: makeCompatibilitySnapshot("对方", "female")},
		},
	}

	changed, err := ensureCompatibilityConsultingAssessment(detail)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("expected consulting backfill to report a change")
	}
	if detail.Reading.ConsultingAssessment.RelationshipDiagnosis.RelationshipType == "" {
		t.Fatal("expected relationship diagnosis to be backfilled")
	}
}
```

- [ ] **Step 2: Run service tests to verify they fail**

Run:

```bash
go test ./internal/service -run 'Test(BuildCompatibilityPromptData_EmbedsConsultingAssessment|EnsureCompatibilityConsultingAssessment_BackfillsMissingConsulting)'
```

Expected: FAIL with missing `ConsultingJSON` and `ensureCompatibilityConsultingAssessment`.

- [ ] **Step 3: Add model prompt/report fields**

In `backend/internal/model/compatibility.go`, extend prompt data and structured report:

```go
type CompatibilityStructuredReport struct {
	Summary                 string                            `json:"summary"`
	Dimensions              []CompatibilityDimensionNarrative `json:"dimensions"`
	DurationAssessment      CompatibilityDurationAssessment   `json:"duration_assessment"`
	RelationshipDiagnosis   CompatibilityRelationshipDiagnosis `json:"relationship_diagnosis"`
	DecisionAdvice          CompatibilityDecisionAdvice        `json:"decision_advice"`
	StageRisks              []CompatibilityStageRisk           `json:"stage_risks"`
	RelationshipStrategy    CompatibilityRelationshipStrategy  `json:"relationship_strategy"`
	ClaimEvidenceLinks      []CompatibilityClaimEvidenceLink   `json:"claim_evidence_links"`
	Risks                   []string                          `json:"risks"`
	Advice                  string                            `json:"advice"`
}

type CompatibilityPromptData struct {
	SelfLabel           string
	PartnerLabel        string
	SelfChartSummary    string
	PartnerChartSummary string
	ScoresJSON          string
	DurationJSON        string
	ConsultingJSON      string
	EvidencesJSON       string
	SummaryTags         string
}
```

- [ ] **Step 4: Map bazi consulting assessment into service model**

In `CreateCompatibilityReading`, build a model consulting value and pass it into `repository.CreateCompatibilityReading`:

```go
consulting := mapCompatibilityConsultingAssessment(analysis.ConsultingAssessment)
reading, err := repository.CreateCompatibilityReading(
	userID,
	string(analysis.OverallLevel),
	model.CompatibilityDimensionScores{
		Attraction: analysis.DimensionScores.Attraction,
		Stability: analysis.DimensionScores.Stability,
		Communication: analysis.DimensionScores.Communication,
		Practicality: analysis.DimensionScores.Practicality,
	},
	model.CompatibilityDurationAssessment{
		OverallBand: analysis.DurationAssessment.OverallBand,
		Windows: model.CompatibilityDurationWindows{
			ThreeMonths: model.CompatibilityDurationWindow{Level: string(analysis.DurationAssessment.Windows.ThreeMonths.Level)},
			OneYear: model.CompatibilityDurationWindow{Level: string(analysis.DurationAssessment.Windows.OneYear.Level)},
			TwoYearsPlus: model.CompatibilityDurationWindow{Level: string(analysis.DurationAssessment.Windows.TwoYearsPlus.Level)},
		},
		Summary: analysis.DurationAssessment.Summary,
		Reasons: analysis.DurationAssessment.Reasons,
	},
	consulting,
	analysis.SummaryTags,
	compatibilityAnalysisVersion,
)
```

When creating evidences, map `EvidenceKey`:

```go
EvidenceKey: item.EvidenceKey,
```

Add mapper helpers:

```go
func mapCompatibilityConsultingAssessment(in bazi.CompatibilityConsultingAssessment) model.CompatibilityConsultingAssessment {
	return model.CompatibilityConsultingAssessment{
		RelationshipDiagnosis: model.CompatibilityRelationshipDiagnosis{
			RelationshipType: in.RelationshipDiagnosis.RelationshipType,
			Verdict: in.RelationshipDiagnosis.Verdict,
			Summary: in.RelationshipDiagnosis.Summary,
			TopFindings: mapCompatibilityFindings(in.RelationshipDiagnosis.TopFindings),
		},
		DecisionAdvice: model.CompatibilityDecisionAdvice{
			Recommendation: in.DecisionAdvice.Recommendation,
			Confidence: in.DecisionAdvice.Confidence,
			Conditions: in.DecisionAdvice.Conditions,
			DoNext: in.DecisionAdvice.DoNext,
			Avoid: in.DecisionAdvice.Avoid,
		},
		StageRisks: mapCompatibilityStageRisks(in.StageRisks),
		RelationshipStrategy: model.CompatibilityRelationshipStrategy{
			Communication: in.RelationshipStrategy.Communication,
			Conflict: in.RelationshipStrategy.Conflict,
			Reality: in.RelationshipStrategy.Reality,
			Boundary: in.RelationshipStrategy.Boundary,
		},
		ClaimEvidenceLinks: mapCompatibilityClaimLinks(in.ClaimEvidenceLinks),
	}
}
```

Add these mapper helpers below `mapCompatibilityConsultingAssessment`:

```go
func mapCompatibilityFindings(in []bazi.CompatibilityFinding) []model.CompatibilityFinding {
	out := make([]model.CompatibilityFinding, 0, len(in))
	for _, item := range in {
		out = append(out, model.CompatibilityFinding{
			Text:         item.Text,
			EvidenceKeys: item.EvidenceKeys,
		})
	}
	return out
}

func mapCompatibilityStageRisks(in []bazi.CompatibilityStageRisk) []model.CompatibilityStageRisk {
	out := make([]model.CompatibilityStageRisk, 0, len(in))
	for _, item := range in {
		out = append(out, model.CompatibilityStageRisk{
			Window:       item.Window,
			RiskLevel:    item.RiskLevel,
			MainRisk:     item.MainRisk,
			Trigger:      item.Trigger,
			Advice:       item.Advice,
			EvidenceKeys: item.EvidenceKeys,
		})
	}
	return out
}

func mapCompatibilityClaimLinks(in []bazi.CompatibilityClaimEvidenceLink) []model.CompatibilityClaimEvidenceLink {
	out := make([]model.CompatibilityClaimEvidenceLink, 0, len(in))
	for _, item := range in {
		out = append(out, model.CompatibilityClaimEvidenceLink{
			ClaimID:      item.ClaimID,
			Claim:        item.Claim,
			EvidenceKeys: item.EvidenceKeys,
			Reasoning:    item.Reasoning,
			Caveat:       item.Caveat,
		})
	}
	return out
}
```

- [ ] **Step 5: Add lazy consulting backfill**

Add:

```go
func ensureCompatibilityConsultingAssessment(detail *model.CompatibilityDetail) (bool, error) {
	if detail == nil || detail.Reading == nil {
		return false, nil
	}
	if detail.Reading.ConsultingAssessment.RelationshipDiagnosis.RelationshipType != "" {
		return false, nil
	}
	selfResult, partnerResult, err := compatibilityResultsFromDetail(detail)
	if err != nil {
		return false, err
	}
	analysis := bazi.AnalyzeCompatibility(selfResult, partnerResult)
	detail.Reading.ConsultingAssessment = mapCompatibilityConsultingAssessment(analysis.ConsultingAssessment)
	return true, nil
}

func compatibilityResultsFromDetail(detail *model.CompatibilityDetail) (*bazi.BaziResult, *bazi.BaziResult, error) {
	var selfResult, partnerResult *bazi.BaziResult
	for i := range detail.Participants {
		p := &detail.Participants[i]
		result, err := compatibilityParticipantResult(p)
		if err != nil {
			return nil, nil, err
		}
		if p.Role == "self" {
			selfResult = result
		} else if p.Role == "partner" {
			partnerResult = result
		}
	}
	if selfResult == nil || partnerResult == nil {
		return nil, nil, fmt.Errorf("合盘参与者信息不完整")
	}
	return selfResult, partnerResult, nil
}
```

Use it in `GetCompatibilityDetailForUser` and `GenerateCompatibilityReport`; if changed, call `repository.UpdateCompatibilityConsultingAssessment`.

- [ ] **Step 6: Add consulting JSON to prompt data and fallback prompt**

In `buildCompatibilityPromptData`, marshal consulting:

```go
consultingJSON, _ := json.Marshal(detail.Reading.ConsultingAssessment)
```

Return it:

```go
ConsultingJSON: string(consultingJSON),
```

In `compatibilityPromptFallback`, add:

```text
咨询型结构化诊断（JSON）：
{{.ConsultingJSON}}
```

Set the output JSON instruction to this exact shape:

```json
{
  "summary": "总体判断，必须基于输入证据，不使用绝对断语",
  "relationship_diagnosis": {
    "relationship_type": "短期吸引强、长期承压型",
    "verdict": "建议谨慎观察",
    "summary": "双方初期靠近感较强，但长期稳定更依赖沟通节奏和现实安排是否能对齐。",
    "top_findings": [
      {
        "text": "吸引力有明显支点，但稳定维度存在拉扯。",
        "evidence_keys": ["spouse_palace_stability_夫妻宫六冲_02"]
      }
    ]
  },
  "decision_advice": {
    "recommendation": "observe",
    "confidence": "medium",
    "conditions": ["先建立稳定沟通规则"],
    "do_next": ["用一到两个月观察冲突后的修复能力"],
    "avoid": ["用短期吸引感替代长期判断"]
  },
  "stage_risks": [
    {
      "window": "three_months",
      "risk_level": "medium",
      "main_risk": "热度高但节奏不一致",
      "trigger": "一方推进过快、另一方需要空间时",
      "advice": "先约定沟通频率和边界，不急于做长期承诺",
      "evidence_keys": ["day_master_communication_日主相克_00"]
    }
  ],
  "relationship_strategy": {
    "communication": "重要议题用明确约定替代情绪试探。",
    "conflict": "争执时先暂停升级，再回到具体事件和责任分工。",
    "reality": "长期计划需要拆成可验证的小步骤。",
    "boundary": "初期保留个人节奏，避免过快形成单方依赖。"
  },
  "claim_evidence_links": [
    {
      "claim_id": "long_term_pressure",
      "claim": "长期关系需要额外经营稳定感。",
      "evidence_keys": ["spouse_palace_stability_夫妻宫六冲_02"],
      "reasoning": "夫妻宫冲动和现实磨合信号叠加时，关系更容易在长期安排中反复消耗。",
      "caveat": "若双方能建立清晰沟通规则，负向信号的影响会被削弱。"
    }
  ],
  "dimensions": [
    { "key": "attraction", "title": "吸引力", "content": "基于证据的维度解释" },
    { "key": "stability", "title": "稳定度", "content": "基于证据的维度解释" },
    { "key": "communication", "title": "沟通协同", "content": "基于证据的维度解释" },
    { "key": "practicality", "title": "现实磨合", "content": "基于证据的维度解释" }
  ],
  "duration_assessment": {
    "overall_band": "medium_term",
    "summary": "阶段性维持判断",
    "reasons": ["只引用输入中已有的阶段原因"],
    "windows": {
      "three_months": { "level": "high" },
      "one_year": { "level": "medium" },
      "two_years_plus": { "level": "low" }
    }
  },
  "risks": ["基于证据的风险点"],
  "advice": "综合建议"
}
```

- [ ] **Step 7: Run service tests**

Run:

```bash
go test ./internal/service -run Compatibility
```

Expected: PASS.

- [ ] **Step 8: Commit Task 3**

```bash
git add backend/internal/model/compatibility.go backend/internal/service/compatibility_service.go backend/internal/service/compatibility_service_test.go
git commit -m "feat(compatibility): add consulting report service shape"
```

## Task 4: Handler Response Coverage

**Files:**
- Modify: `backend/internal/handler/compatibility_handler_test.go`

- [ ] **Step 1: Add response shape JSON tag assertion**

Add this import to `backend/internal/handler/compatibility_handler_test.go`:

```go
import "yuanju/internal/model"
```

Then add this test. It avoids DB setup and protects the JSON response contract used by the handler:

```go
func TestCompatibilityDetailJSON_IncludesConsultingShape(t *testing.T) {
	detail := model.CompatibilityDetail{
		Reading: &model.CompatibilityReading{
			ConsultingAssessment: model.CompatibilityConsultingAssessment{
				RelationshipDiagnosis: model.CompatibilityRelationshipDiagnosis{
					RelationshipType: "短期吸引强、长期承压型",
					Verdict: "建议谨慎观察",
				},
			},
		},
		Evidences: []model.CompatibilityEvidence{
			{EvidenceKey: "spouse_palace_stability_夫妻宫六冲_02", Title: "夫妻宫六冲"},
		},
	}
	raw, err := json.Marshal(detail)
	if err != nil {
		t.Fatal(err)
	}
	body := string(raw)
	if !strings.Contains(body, `"consulting_assessment"`) {
		t.Fatalf("expected consulting assessment in json: %s", body)
	}
	if !strings.Contains(body, `"relationship_diagnosis"`) {
		t.Fatalf("expected relationship diagnosis in json: %s", body)
	}
	if !strings.Contains(body, `"evidence_key"`) {
		t.Fatalf("expected evidence key in json: %s", body)
	}
}
```

- [ ] **Step 2: Run handler test to verify or fix setup**

Run:

```bash
go test ./internal/handler -run TestCompatibilityDetailJSON_IncludesConsultingShape
```

Expected: PASS after Task 2/3.

- [ ] **Step 3: Commit Task 4**

```bash
git add backend/internal/handler/compatibility_handler_test.go
git commit -m "test(compatibility): cover consulting response shape"
```

## Task 5: Frontend API Types and Fallback Normalizers

**Files:**
- Modify: `frontend/src/lib/api.ts`
- Modify: `frontend/src/pages/CompatibilityResultPage.tsx`

- [ ] **Step 1: Extend TypeScript API types**

In `frontend/src/lib/api.ts`, add:

```ts
export interface CompatibilityFinding {
  text: string
  evidence_keys: string[]
}

export interface CompatibilityRelationshipDiagnosis {
  relationship_type: string
  verdict: string
  summary: string
  top_findings: CompatibilityFinding[]
}

export interface CompatibilityDecisionAdvice {
  recommendation: 'continue' | 'observe' | 'caution' | string
  confidence: 'high' | 'medium' | 'low' | string
  conditions: string[]
  do_next: string[]
  avoid: string[]
}

export interface CompatibilityStageRisk {
  window: 'three_months' | 'one_year' | 'two_years_plus' | string
  risk_level: 'high' | 'medium' | 'low' | string
  main_risk: string
  trigger: string
  advice: string
  evidence_keys: string[]
}

export interface CompatibilityRelationshipStrategy {
  communication: string
  conflict: string
  reality: string
  boundary: string
}

export interface CompatibilityClaimEvidenceLink {
  claim_id: string
  claim: string
  evidence_keys: string[]
  reasoning: string
  caveat: string
}

export interface CompatibilityConsultingAssessment {
  relationship_diagnosis: CompatibilityRelationshipDiagnosis
  decision_advice: CompatibilityDecisionAdvice
  stage_risks: CompatibilityStageRisk[]
  relationship_strategy: CompatibilityRelationshipStrategy
  claim_evidence_links: CompatibilityClaimEvidenceLink[]
}
```

Extend:

```ts
export interface CompatibilityEvidence {
  id: string
  reading_id: string
  evidence_key: string
  dimension: 'attraction' | 'stability' | 'communication' | 'practicality'
  type: string
  polarity: 'positive' | 'negative' | 'mixed' | 'neutral'
  source: string
  title: string
  detail: string
  weight: number
  created_at: string
}

export interface CompatibilityReading {
  id: string
  user_id: string
  overall_level: 'high' | 'medium' | 'low'
  dimension_scores: CompatibilityDimensionScores
  duration_assessment: CompatibilityDurationAssessment
  consulting_assessment: CompatibilityConsultingAssessment
  summary_tags: string[]
  analysis_version: string
  created_at: string
  updated_at: string
}

export interface CompatibilityStructuredReport {
  summary: string
  dimensions: Array<{ key: string; title: string; content: string }>
  duration_assessment: CompatibilityDurationAssessment
  relationship_diagnosis?: CompatibilityRelationshipDiagnosis
  decision_advice?: CompatibilityDecisionAdvice
  stage_risks?: CompatibilityStageRisk[]
  relationship_strategy?: CompatibilityRelationshipStrategy
  claim_evidence_links?: CompatibilityClaimEvidenceLink[]
}
```

- [ ] **Step 2: Add frontend normalizer tests**

Append to `frontend/tests/compatibility-result-ux.test.mjs`:

```js
test('compatibility result page defines consulting report sections', () => {
  const page = read('src/pages/CompatibilityResultPage.tsx')
  assert.match(page, /ConsultingOverview/)
  assert.match(page, /DecisionAdvicePanel/)
  assert.match(page, /StageRiskGrid/)
  assert.match(page, /RelationshipStrategyPanel/)
  assert.match(page, /EvidenceLinkedClaims/)
})
```

- [ ] **Step 3: Run frontend test to verify it fails**

Run:

```bash
cd frontend && node --test tests/compatibility-result-ux.test.mjs
```

Expected: FAIL because components are not defined yet.

- [ ] **Step 4: Commit Task 5 types/tests**

```bash
git add frontend/src/lib/api.ts frontend/tests/compatibility-result-ux.test.mjs
git commit -m "test(compatibility): define consulting frontend contract"
```

## Task 6: Render Consulting Sections on Result Page

**Files:**
- Modify: `frontend/src/pages/CompatibilityResultPage.tsx`
- Modify: `frontend/src/pages/CompatibilityResultPage.css`
- Test: `frontend/tests/compatibility-result-ux.test.mjs`

- [ ] **Step 1: Add labels and fallback helpers**

In `CompatibilityResultPage.tsx`, add mappings:

```ts
const recommendationText: Record<string, string> = {
  continue: '适合继续推进',
  observe: '建议谨慎观察',
  caution: '不宜过早重投入',
}

const confidenceText: Record<string, string> = {
  high: '判断较明确',
  medium: '需要结合相处验证',
  low: '信息仍需观察',
}

const stageWindowText: Record<string, string> = {
  three_months: '3 个月',
  one_year: '1 年',
  two_years_plus: '2 年以上',
}
```

Add fallback normalizer:

```ts
function normalizeConsultingAssessment(detail: CompatibilityDetail) {
  const report = detail.latest_report?.content_structured
  const base = detail.reading.consulting_assessment
  return {
    relationship_diagnosis: report?.relationship_diagnosis || base?.relationship_diagnosis,
    decision_advice: report?.decision_advice || base?.decision_advice,
    stage_risks: report?.stage_risks?.length ? report.stage_risks : base?.stage_risks || [],
    relationship_strategy: report?.relationship_strategy || base?.relationship_strategy,
    claim_evidence_links: report?.claim_evidence_links?.length ? report.claim_evidence_links : base?.claim_evidence_links || [],
  }
}
```

- [ ] **Step 2: Add consulting components**

Add focused components above `export default function CompatibilityResultPage()`:

```tsx
function ConsultingOverview({ diagnosis }: { diagnosis: CompatibilityRelationshipDiagnosis }) {
  return (
    <div className="card compatibility-consulting-overview">
      <div className="compatibility-consulting-kicker">关系诊断</div>
      <h2 className="serif compatibility-consulting-title">{diagnosis.relationship_type || '关系观察型'}</h2>
      <div className="compatibility-consulting-verdict">{diagnosis.verdict || '建议结合现实相处继续观察'}</div>
      <p className="compatibility-consulting-summary">{diagnosis.summary}</p>
      <div className="compatibility-finding-list">
        {(diagnosis.top_findings || []).slice(0, 3).map((finding) => (
          <div key={finding.text} className="compatibility-finding-item">{finding.text}</div>
        ))}
      </div>
    </div>
  )
}

function DecisionAdvicePanel({ advice }: { advice: CompatibilityDecisionAdvice }) {
  return (
    <div className="card compatibility-decision-card">
      <div className="compatibility-consulting-kicker">决策建议</div>
      <div className="serif compatibility-decision-main">{recommendationText[advice.recommendation] || advice.recommendation}</div>
      <div className="compatibility-decision-confidence">{confidenceText[advice.confidence] || advice.confidence}</div>
      <div className="compatibility-advice-columns">
        <AdviceList title="继续前提" items={advice.conditions} />
        <AdviceList title="下一步" items={advice.do_next} />
        <AdviceList title="避免" items={advice.avoid} />
      </div>
    </div>
  )
}

function AdviceList({ title, items }: { title: string; items: string[] }) {
  return (
    <div className="compatibility-advice-list">
      <div className="compatibility-advice-title">{title}</div>
      {(items || []).length > 0 ? items.map(item => <div key={item}>{item}</div>) : <div>暂无明确建议</div>}
    </div>
  )
}

function StageRiskGrid({ risks }: { risks: CompatibilityStageRisk[] }) {
  return (
    <div className="compatibility-stage-grid">
      {risks.map(risk => (
        <div key={risk.window} className="card compatibility-stage-card">
          <div className="compatibility-stage-window">{stageWindowText[risk.window] || risk.window}</div>
          <div className="serif compatibility-stage-risk">{risk.main_risk}</div>
          <p>{risk.trigger}</p>
          <div className="compatibility-stage-advice">{risk.advice}</div>
        </div>
      ))}
    </div>
  )
}

function RelationshipStrategyPanel({ strategy }: { strategy: CompatibilityRelationshipStrategy }) {
  return (
    <div className="card compatibility-strategy-card">
      <div className="compatibility-consulting-kicker">关系经营策略</div>
      <div className="compatibility-strategy-grid">
        <AdviceList title="沟通" items={[strategy.communication].filter(Boolean)} />
        <AdviceList title="冲突" items={[strategy.conflict].filter(Boolean)} />
        <AdviceList title="现实" items={[strategy.reality].filter(Boolean)} />
        <AdviceList title="边界" items={[strategy.boundary].filter(Boolean)} />
      </div>
    </div>
  )
}

function EvidenceLinkedClaims({
  links,
  evidences,
}: {
  links: CompatibilityClaimEvidenceLink[]
  evidences: CompatibilityEvidence[]
}) {
  const byKey = new Map(evidences.map(evidence => [evidence.evidence_key || evidence.id, evidence]))
  return (
    <div className="compatibility-claim-list">
      {links.map(link => (
        <details key={link.claim_id || link.claim} className="card compatibility-claim-card">
          <summary>
            <span className="serif">{link.claim}</span>
            <span>查看依据</span>
          </summary>
          <p>{link.reasoning}</p>
          {link.caveat && <p className="compatibility-claim-caveat">{link.caveat}</p>}
          <div className="compatibility-claim-evidence">
            {(link.evidence_keys || []).map(key => {
              const evidence = byKey.get(key)
              return evidence ? <EvidenceCard key={key} evidence={evidence} /> : null
            })}
          </div>
        </details>
      ))}
    </div>
  )
}
```

- [ ] **Step 3: Insert consulting sections into page**

After `summaryTags` and before `ScoreOverview`, derive:

```ts
const consulting = normalizeConsultingAssessment(detail)
```

Render:

```tsx
{consulting.relationship_diagnosis && (
  <ConsultingOverview diagnosis={consulting.relationship_diagnosis} />
)}
{consulting.decision_advice && (
  <DecisionAdvicePanel advice={consulting.decision_advice} />
)}
{consulting.stage_risks.length > 0 && (
  <div className="compatibility-section">
    <div className="compatibility-section-header">
      <h2 className="serif compatibility-section-title">阶段风险预警</h2>
      <p className="compatibility-section-desc">按关系推进阶段看主要风险和处理建议。</p>
    </div>
    <StageRiskGrid risks={consulting.stage_risks} />
  </div>
)}
{consulting.relationship_strategy && (
  <RelationshipStrategyPanel strategy={consulting.relationship_strategy} />
)}
{consulting.claim_evidence_links.length > 0 && (
  <div className="compatibility-section">
    <div className="compatibility-section-header">
      <h2 className="serif compatibility-section-title">关键判断依据</h2>
      <p className="compatibility-section-desc">每条咨询判断都可以回看对应命理证据。</p>
    </div>
    <EvidenceLinkedClaims links={consulting.claim_evidence_links} evidences={detail.evidences} />
  </div>
)}
```

- [ ] **Step 4: Add CSS**

Add to `CompatibilityResultPage.css`:

```css
.compatibility-consulting-overview,
.compatibility-decision-card,
.compatibility-strategy-card {
  padding: 24px;
}

.compatibility-consulting-kicker,
.compatibility-advice-title,
.compatibility-stage-window {
  color: var(--text-muted);
  font-size: 13px;
  letter-spacing: 0;
}

.compatibility-consulting-title,
.compatibility-decision-main {
  margin: 8px 0;
  font-size: 28px;
}

.compatibility-consulting-verdict {
  display: inline-flex;
  width: fit-content;
  padding: 6px 10px;
  border-radius: 8px;
  color: var(--accent);
  background: rgba(201, 162, 93, 0.12);
  border: 1px solid rgba(201, 162, 93, 0.24);
}

.compatibility-finding-list,
.compatibility-advice-columns,
.compatibility-strategy-grid,
.compatibility-stage-grid {
  display: grid;
  gap: 12px;
}

.compatibility-advice-columns,
.compatibility-strategy-grid,
.compatibility-stage-grid {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.compatibility-finding-item,
.compatibility-advice-list,
.compatibility-stage-card,
.compatibility-claim-card {
  padding: 14px;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.08);
}

.compatibility-claim-card summary {
  cursor: pointer;
  display: flex;
  justify-content: space-between;
  gap: 16px;
}

.compatibility-claim-evidence {
  display: grid;
  gap: 12px;
  margin-top: 12px;
}

@media (max-width: 768px) {
  .compatibility-consulting-title,
  .compatibility-decision-main {
    font-size: 23px;
  }

  .compatibility-advice-columns,
  .compatibility-strategy-grid,
  .compatibility-stage-grid {
    grid-template-columns: 1fr;
  }
}
```

- [ ] **Step 5: Run frontend compatibility UX test**

Run:

```bash
cd frontend && node --test tests/compatibility-result-ux.test.mjs
```

Expected: PASS.

- [ ] **Step 6: Commit Task 6**

```bash
git add frontend/src/pages/CompatibilityResultPage.tsx frontend/src/pages/CompatibilityResultPage.css frontend/tests/compatibility-result-ux.test.mjs
git commit -m "feat(compatibility): render consulting report sections"
```

## Task 7: Final Verification

**Files:**
- No new files; verify integrated behavior.

- [ ] **Step 1: Run focused backend tests**

Run:

```bash
go test ./pkg/bazi ./internal/service ./internal/handler
```

Expected: PASS.

- [ ] **Step 2: Run frontend tests**

Run:

```bash
cd frontend && node --test tests/compatibility-result-ux.test.mjs tests/mobile-page-qa-matrix.test.mjs
```

Expected: PASS.

- [ ] **Step 3: Run full build if dependencies are available**

Run:

```bash
cd frontend && npm run build
```

Expected: Vite build succeeds. If dependencies are missing, run `npm install` only with explicit user approval because it may require network access.

- [ ] **Step 4: Manual smoke test with local app**

Run backend/frontend using the repository’s normal development path, then create a compatibility reading and verify:

- Result page shows relationship diagnosis before scores.
- Decision advice appears without generating AI report.
- Stage risks show 3 months, 1 year, and 2 years plus.
- Key judgement details expand and show evidence cards.
- Professional details still show both participants and evidence list.
- Generating AI report keeps the same consulting structure and can enrich text without removing algorithm fallback.

- [ ] **Step 5: Commit verification notes if docs changed**

Only run this if a verification note or follow-up doc is added:

```bash
git add docs/superpowers/plans/2026-05-20-compatibility-consulting-report.md
git commit -m "docs(plan): compatibility consulting report verification notes"
```

## Self-Review Notes

- Spec coverage: relationship diagnosis, decision advice, stage risks, relationship strategy, evidence links, professional evidence expansion, AI fallback, and no absolute date predictions are each covered by tasks.
- Persistence coverage: new JSONB field stores algorithm fallback; report JSON can override/enrich it; old records are handled with lazy backfill.
- Frontend coverage: conclusion-first rendering is covered by static UX tests and manual smoke test.
- Scope control: no object comparison, relationship diary, social sharing, or practitioner workflow is included.
