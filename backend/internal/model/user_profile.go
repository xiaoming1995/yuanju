package model

import "time"

type UserProfileStats struct {
	ChartCount         int `json:"chart_count"`
	AIReportCount      int `json:"ai_report_count"`
	CompatibilityCount int `json:"compatibility_count"`
}

type UserProfileChartSummary struct {
	ID         string    `json:"id"`
	BirthYear  int       `json:"birth_year"`
	BirthMonth int       `json:"birth_month"`
	BirthDay   int       `json:"birth_day"`
	BirthHour  int       `json:"birth_hour"`
	Gender     string    `json:"gender"`
	YearGan    string    `json:"year_gan"`
	YearZhi    string    `json:"year_zhi"`
	MonthGan   string    `json:"month_gan"`
	MonthZhi   string    `json:"month_zhi"`
	DayGan     string    `json:"day_gan"`
	DayZhi     string    `json:"day_zhi"`
	HourGan    string    `json:"hour_gan"`
	HourZhi    string    `json:"hour_zhi"`
	Yongshen   string    `json:"yongshen"`
	Jishen     string    `json:"jishen"`
	CreatedAt  time.Time `json:"created_at"`
}

type UserProfileCompatibilitySummary struct {
	ID           string    `json:"id"`
	OverallLevel string    `json:"overall_level"`
	SelfName     string    `json:"self_name"`
	PartnerName  string    `json:"partner_name"`
	SummaryTags  []string  `json:"summary_tags"`
	CreatedAt    time.Time `json:"created_at"`
}

type UserProfileFeatureEntry struct {
	Key         string `json:"key"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

type UserProfileOverview struct {
	User                *User                             `json:"user"`
	Stats               UserProfileStats                  `json:"stats"`
	RecentCharts        []UserProfileChartSummary         `json:"recent_charts"`
	RecentCompatibility []UserProfileCompatibilitySummary `json:"recent_compatibility"`
	Features            []UserProfileFeatureEntry         `json:"features"`
}
