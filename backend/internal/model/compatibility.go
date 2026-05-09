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

type CompatibilityDimensionScores struct {
	Attraction    int `json:"attraction"`
	Stability     int `json:"stability"`
	Communication int `json:"communication"`
	Practicality  int `json:"practicality"`
}

type CompatibilityEvidence struct {
	ID        string    `json:"id"`
	ReadingID string    `json:"reading_id"`
	Dimension string    `json:"dimension"`
	Type      string    `json:"type"`
	Polarity  string    `json:"polarity"`
	Source    string    `json:"source"`
	Title     string    `json:"title"`
	Detail    string    `json:"detail"`
	Weight    int       `json:"weight"`
	CreatedAt time.Time `json:"created_at"`
}

type CompatibilityReading struct {
	ID              string                       `json:"id"`
	UserID          string                       `json:"user_id"`
	OverallLevel    string                       `json:"overall_level"`
	DimensionScores CompatibilityDimensionScores `json:"dimension_scores"`
	SummaryTags     []string                     `json:"summary_tags"`
	AnalysisVersion string                       `json:"analysis_version"`
	CreatedAt       time.Time                    `json:"created_at"`
	UpdatedAt       time.Time                    `json:"updated_at"`
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

type CompatibilityStructuredReport struct {
	Summary    string                            `json:"summary"`
	Dimensions []CompatibilityDimensionNarrative `json:"dimensions"`
	Risks      []string                          `json:"risks"`
	Advice     string                            `json:"advice"`
}

type CompatibilityPromptData struct {
	SelfLabel           string
	PartnerLabel        string
	SelfChartSummary    string
	PartnerChartSummary string
	ScoresJSON          string
	EvidencesJSON       string
	SummaryTags         string
}

type CompatibilityDetail struct {
	Reading      *CompatibilityReading      `json:"reading"`
	Participants []CompatibilityParticipant `json:"participants"`
	Evidences    []CompatibilityEvidence    `json:"evidences"`
	LatestReport *AICompatibilityReport     `json:"latest_report,omitempty"`
}

type CompatibilityHistoryItem struct {
	ID              string                       `json:"id"`
	OverallLevel    string                       `json:"overall_level"`
	DimensionScores CompatibilityDimensionScores `json:"dimension_scores"`
	SummaryTags     []string                     `json:"summary_tags"`
	SelfName        string                       `json:"self_name"`
	PartnerName     string                       `json:"partner_name"`
	CreatedAt       time.Time                    `json:"created_at"`
}
