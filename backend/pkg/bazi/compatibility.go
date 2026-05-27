package bazi

// Compatibility scoring v3 — 100-point classical formula
// (zodiac 50 / nayin 20 / day_pillar 10 / eight_chars 20).
// Design: docs/superpowers/specs/2026-05-27-compatibility-scoring-formula-v2-design.md

type CompatibilityLevel string

const (
	CompatibilityHigh   CompatibilityLevel = "high"
	CompatibilityMedium CompatibilityLevel = "medium"
	CompatibilityLow    CompatibilityLevel = "low"
)

type CompatibilityDurationLevel string

const (
	CompatibilityDurationHigh   CompatibilityDurationLevel = "high"
	CompatibilityDurationMedium CompatibilityDurationLevel = "medium"
	CompatibilityDurationLow    CompatibilityDurationLevel = "low"
)

type CompatibilityDimensionScores struct {
	Zodiac     int `json:"zodiac"`
	Nayin      int `json:"nayin"`
	DayPillar  int `json:"day_pillar"`
	EightChars int `json:"eight_chars"`
}

type CompatibilityEvidence struct {
	EvidenceKey    string   `json:"evidence_key"`
	Dimension      string   `json:"dimension"`
	Type           string   `json:"type"`
	Polarity       string   `json:"polarity"`
	Source         string   `json:"source"`
	Perspective    string   `json:"perspective,omitempty"`
	Actor          string   `json:"actor,omitempty"`
	Target         string   `json:"target,omitempty"`
	RelatedSources []string `json:"related_sources,omitempty"`
	Title          string   `json:"title"`
	Detail         string   `json:"detail"`
	Weight         int      `json:"weight"`
}

type CompatibilityScoreExplanation struct {
	Dimension            string   `json:"dimension"`
	PositiveFactor       string   `json:"positive_factor,omitempty"`
	NegativeFactor       string   `json:"negative_factor,omitempty"`
	PositiveEvidenceKeys []string `json:"positive_evidence_keys,omitempty"`
	NegativeEvidenceKeys []string `json:"negative_evidence_keys,omitempty"`
	Summary              string   `json:"summary"`
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

type CompatibilityAnalysis struct {
	OverallScore         int                               `json:"overall_score"`
	OverallLevel         CompatibilityLevel                `json:"overall_level"`
	DimensionScores      CompatibilityDimensionScores      `json:"dimension_scores"`
	Evidences            []CompatibilityEvidence           `json:"evidences"`
	ScoreExplanations    []CompatibilityScoreExplanation   `json:"score_explanations"`
	SummaryTags          []string                          `json:"summary_tags"`
	DurationAssessment   CompatibilityDurationAssessment   `json:"duration_assessment"`
	ConsultingAssessment CompatibilityConsultingAssessment `json:"consulting_assessment"`
}

// AnalyzeCompatibility 是合盘评分 v3 的公开入口。
// 计算 4 模块得分（合属相 / 合纳音 / 合日柱 / 合八字），汇总到总分（0–100），
// 并产出 evidence / score_explanations / summary_tags / duration / consulting 全套结构。
func AnalyzeCompatibility(a, b *BaziResult) CompatibilityAnalysis {
	if a == nil || b == nil {
		return CompatibilityAnalysis{
			OverallLevel:    CompatibilityLow,
			DimensionScores: CompatibilityDimensionScores{},
		}
	}
	scores := CompatibilityDimensionScores{
		Zodiac:    scoreZodiac(a.YearZhi, b.YearZhi),
		Nayin:     scoreNayin(a.YearGan+a.YearZhi, b.YearGan+b.YearZhi),
		DayPillar: scoreDayPillar(a.DayGan, a.DayZhi, b.DayGan, b.DayZhi),
		EightChars: scoreEightChars(
			a.YearGan, a.YearZhi, b.YearGan, b.YearZhi,
			a.MonthGan, a.MonthZhi, b.MonthGan, b.MonthZhi,
			a.HourGan, a.HourZhi, b.HourGan, b.HourZhi,
		),
	}
	total := scores.Zodiac + scores.Nayin + scores.DayPillar + scores.EightChars
	evidences := buildCompatibilityEvidencesV3(a, b)
	explanations := buildScoreExplanationsV3(a, b, evidences)
	tags := buildSummaryTagsV3(scores, total)
	duration := buildDurationAssessmentV3(scores)
	duration.Reasons = durationReasonsFromEvidence(evidences)
	hits := countHitsV3(scores)
	consulting := buildConsultingAssessmentV3(total, hits, scores, evidences, duration)
	return CompatibilityAnalysis{
		OverallScore:         total,
		OverallLevel:         overallLevelFromScoreV3(total),
		DimensionScores:      scores,
		Evidences:            evidences,
		ScoreExplanations:    explanations,
		SummaryTags:          tags,
		DurationAssessment:   duration,
		ConsultingAssessment: consulting,
	}
}

// overallLevelFromScoreV3 把 0–100 总分映射到 high / medium / low 三档。
func overallLevelFromScoreV3(total int) CompatibilityLevel {
	switch {
	case total >= 80:
		return CompatibilityHigh
	case total >= 60:
		return CompatibilityMedium
	default:
		return CompatibilityLow
	}
}

// countHitsV3 统计 4 模块中 score > 0 的个数（0–4）。
func countHitsV3(s CompatibilityDimensionScores) int {
	n := 0
	if s.Zodiac > 0 {
		n++
	}
	if s.Nayin > 0 {
		n++
	}
	if s.DayPillar > 0 {
		n++
	}
	if s.EightChars > 0 {
		n++
	}
	return n
}
