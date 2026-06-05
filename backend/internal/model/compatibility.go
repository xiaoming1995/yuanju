package model

import (
	"encoding/json"
	"time"
)

type CompatibilityBirthProfile struct {
	Year         int    `json:"year"`
	Month        int    `json:"month"`
	Day          int    `json:"day"`
	Hour         int    `json:"hour"`
	Gender       string `json:"gender"`
	CalendarType string `json:"calendar_type"`
	IsLeapMonth  bool   `json:"is_leap_month"`
}

type CompatibilityContext struct {
	RelationshipStage string `json:"relationship_stage"`
	PrimaryQuestion   string `json:"primary_question"`
}

type CompatibilityDisplayNames struct {
	Self    string
	Partner string
}

type CompatibilityDimensionScores struct {
	Zodiac     int `json:"zodiac"`
	Nayin      int `json:"nayin"`
	DayPillar  int `json:"day_pillar"`
	EightChars int `json:"eight_chars"`
}

type CompatibilityDurationWindow struct {
	Level string `json:"level"`
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

type CompatibilityEvidence struct {
	ID             string    `json:"id"`
	ReadingID      string    `json:"reading_id"`
	EvidenceKey    string    `json:"evidence_key"`
	Dimension      string    `json:"dimension"`
	Type           string    `json:"type"`
	Polarity       string    `json:"polarity"`
	Source         string    `json:"source"`
	Perspective    string    `json:"perspective,omitempty"`
	Actor          string    `json:"actor,omitempty"`
	Target         string    `json:"target,omitempty"`
	RelatedSources []string  `json:"related_sources,omitempty"`
	Title          string    `json:"title"`
	Detail         string    `json:"detail"`
	Weight         int       `json:"weight"`
	CreatedAt      time.Time `json:"created_at"`
}

type CompatibilityReading struct {
	ID                   string                            `json:"id"`
	UserID               string                            `json:"user_id"`
	RelationshipStage    string                            `json:"relationship_stage"`
	PrimaryQuestion      string                            `json:"primary_question"`
	OverallScore         int                               `json:"overall_score"`
	OverallLevel         string                            `json:"overall_level"`
	DimensionScores      CompatibilityDimensionScores      `json:"dimension_scores"`
	ScoreExplanations    []CompatibilityScoreExplanation   `json:"score_explanations"`
	DurationAssessment   CompatibilityDurationAssessment   `json:"duration_assessment"`
	ConsultingAssessment CompatibilityConsultingAssessment `json:"consulting_assessment"`
	SummaryTags          []string                          `json:"summary_tags"`
	AnalysisVersion      string                            `json:"analysis_version"`
	CreatedAt            time.Time                         `json:"created_at"`
	UpdatedAt            time.Time                         `json:"updated_at"`
}

type CompatibilityParticipant struct {
	ID            string                    `json:"id"`
	ReadingID     string                    `json:"reading_id"`
	Role          string                    `json:"role"`
	DisplayName   string                    `json:"display_name"`
	BirthProfile  CompatibilityBirthProfile `json:"birth_profile"`
	ChartHash     string                    `json:"chart_hash"`
	ChartSnapshot *json.RawMessage          `json:"chart_snapshot,omitempty"`
	CreatedAt     time.Time                 `json:"created_at"`
}

type AICompatibilityReport struct {
	ID                string           `json:"id"`
	ReadingID         string           `json:"reading_id"`
	Content           string           `json:"content"`
	ContentStructured *json.RawMessage `json:"content_structured,omitempty"`
	Model             string           `json:"model"`
	CreatedAt         time.Time        `json:"created_at"`
}

type CompatibilityDimensionNarrative struct {
	Key     string `json:"key"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type CompatibilityQuestionFocus struct {
	Title              string   `json:"title"`
	Judgment           string   `json:"judgment"`
	KeyChecks          []string `json:"key_checks"`
	BoundaryConditions []string `json:"boundary_conditions"`
}

type CompatibilityPersonalityDimension struct {
	Key    string `json:"key"`
	Detail string `json:"detail"`
}

type CompatibilityPersonalityPortrait struct {
	Headline   string                              `json:"headline"`
	Dimensions []CompatibilityPersonalityDimension `json:"dimensions"`
}

type CompatibilityPersonalityPoint struct {
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

type CompatibilityPersonalityComparison struct {
	Self        CompatibilityPersonalityPortrait `json:"self"`
	Partner     CompatibilityPersonalityPortrait `json:"partner"`
	FitPoints   []CompatibilityPersonalityPoint  `json:"fit_points"`
	ClashPoints []CompatibilityPersonalityPoint  `json:"clash_points"`
}

type CompatibilitySpousePalaceSide struct {
	IdealPortrait string   `json:"ideal_portrait"`
	MatchLevel    string   `json:"match_level"`
	FitPoints     []string `json:"fit_points"`
	GapPoints     []string `json:"gap_points"`
	EvidenceKeys  []string `json:"evidence_keys"`
}

type CompatibilitySpousePalaceMatch struct {
	Self    CompatibilitySpousePalaceSide `json:"self"`
	Partner CompatibilitySpousePalaceSide `json:"partner"`
	Summary string                        `json:"summary"`
}

type CompatibilityStructuredReport struct {
	Summary               string                              `json:"summary"`
	FamousCouple          *CompatibilityFamousCouple          `json:"famous_couple,omitempty"`
	QuestionFocus         CompatibilityQuestionFocus          `json:"question_focus"`
	PersonalityComparison *CompatibilityPersonalityComparison `json:"personality_comparison,omitempty"`
	SpousePalaceMatch     *CompatibilitySpousePalaceMatch     `json:"spouse_palace_match,omitempty"`
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
	Couple  string `json:"couple"`  // 这对 CP 的名字，例如「梁山伯与祝英台」
	Tagline string `json:"tagline"` // 一句话点出关系气质
	Reason  string `json:"reason"`  // 1–2 句大白话，扣住报告里已有的信号
}

type CompatibilityPromptData struct {
	SelfLabel              string
	PartnerLabel           string
	RelationshipStage      string
	RelationshipStageLabel string
	PrimaryQuestion        string
	PrimaryQuestionLabel   string
	QuestionGuidance       string
	SelfChartSummary       string
	PartnerChartSummary    string
	ScoresJSON             string
	ScoreExplanationsJSON  string
	DurationJSON           string
	ConsultingJSON         string
	EvidencesJSON          string
	EvidenceGroupsJSON     string
	SummaryTags            string
}

type CompatibilityDetail struct {
	Reading      *CompatibilityReading      `json:"reading"`
	Participants []CompatibilityParticipant `json:"participants"`
	Evidences    []CompatibilityEvidence    `json:"evidences"`
	LatestReport *AICompatibilityReport     `json:"latest_report,omitempty"`
}

type CompatibilityHistoryItem struct {
	ID                string                       `json:"id"`
	RelationshipStage string                       `json:"relationship_stage"`
	PrimaryQuestion   string                       `json:"primary_question"`
	OverallScore      int                          `json:"overall_score"`
	OverallLevel      string                       `json:"overall_level"`
	DimensionScores   CompatibilityDimensionScores `json:"dimension_scores"`
	SummaryTags       []string                     `json:"summary_tags"`
	SelfName          string                       `json:"self_name"`
	PartnerName       string                       `json:"partner_name"`
	CreatedAt         time.Time                    `json:"created_at"`
}
