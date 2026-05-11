package service

import (
	"encoding/json"
	"strings"
	"testing"
	"yuanju/internal/model"
	"yuanju/pkg/bazi"
)

func makeCompatibilitySnapshot(displayName, gender string) *json.RawMessage {
	snapshot, _ := json.Marshal(bazi.BaziResult{
		YearGan: "甲", YearZhi: "子",
		MonthGan: "丙", MonthZhi: "寅",
		DayGan: "甲", DayZhi: "子",
		HourGan: "丁", HourZhi: "卯",
		Gender: gender,
		Wuxing: bazi.WuxingStats{Mu: 2, Huo: 2, Tu: 1, Jin: 1, Shui: 2},
		Yongshen: "火",
		Jishen:   "金",
	})
	raw := json.RawMessage(snapshot)
	_ = displayName
	return &raw
}

func TestBuildCompatibilityPromptData_EmbedsDurationAssessment(t *testing.T) {
	detail := &model.CompatibilityDetail{
		Reading: &model.CompatibilityReading{
			DimensionScores: model.CompatibilityDimensionScores{
				Attraction:    78,
				Stability:     54,
				Communication: 66,
				Practicality:  48,
			},
			SummaryTags: []string{"吸引力强", "关系波动"},
			DurationAssessment: model.CompatibilityDurationAssessment{
				OverallBand: "medium_term",
				Windows: model.CompatibilityDurationWindows{
					ThreeMonths:  model.CompatibilityDurationWindow{Level: "high"},
					OneYear:      model.CompatibilityDurationWindow{Level: "medium"},
					TwoYearsPlus: model.CompatibilityDurationWindow{Level: "low"},
				},
				Summary: "前期吸引强，但长期承压。",
				Reasons: []string{"夫妻宫冲克明显", "配偶星呼应强"},
			},
		},
		Participants: []model.CompatibilityParticipant{
			{
				Role:          "self",
				DisplayName:   "我",
				ChartSnapshot: makeCompatibilitySnapshot("我", "male"),
			},
			{
				Role:          "partner",
				DisplayName:   "对方",
				ChartSnapshot: makeCompatibilitySnapshot("对方", "female"),
			},
		},
	}

	got, err := buildCompatibilityPromptData(detail)
	if err != nil {
		t.Fatal(err)
	}
	if got.DurationJSON == "" {
		t.Fatal("expected duration json")
	}
	if !strings.Contains(got.DurationJSON, `"overall_band":"medium_term"`) {
		t.Fatalf("expected duration overall band in prompt data, got %s", got.DurationJSON)
	}
}

func TestEnsureCompatibilityDurationAssessment_BackfillsMissingDuration(t *testing.T) {
	detail := &model.CompatibilityDetail{
		Reading: &model.CompatibilityReading{
			DimensionScores: model.CompatibilityDimensionScores{
				Attraction:    72,
				Stability:     58,
				Communication: 61,
				Practicality:  55,
			},
		},
		Participants: []model.CompatibilityParticipant{
			{
				Role:          "self",
				DisplayName:   "我",
				ChartSnapshot: makeCompatibilitySnapshot("我", "male"),
			},
			{
				Role:          "partner",
				DisplayName:   "对方",
				ChartSnapshot: makeCompatibilitySnapshot("对方", "female"),
			},
		},
	}

	changed, err := ensureCompatibilityDurationAssessment(detail)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("expected lazy duration backfill to report a change")
	}
	if detail.Reading.DurationAssessment.OverallBand == "" {
		t.Fatal("expected duration overall band to be backfilled")
	}
	if detail.Reading.DurationAssessment.Windows.ThreeMonths.Level == "" {
		t.Fatal("expected duration windows to be backfilled")
	}
}

func TestNormalizeCompatibilityProfile_DefaultsSolar(t *testing.T) {
	p := model.CompatibilityBirthProfile{Year: 1990, Month: 1, Day: 1, Hour: 0, Gender: "male"}
	got := normalizeCompatibilityProfile(p)
	if got.CalendarType != "solar" {
		t.Errorf("expected solar, got %q", got.CalendarType)
	}
}

func TestNormalizeCompatibilityProfile_PreservesExisting(t *testing.T) {
	p := model.CompatibilityBirthProfile{Year: 1990, Month: 1, Day: 1, Hour: 0, Gender: "female", CalendarType: "lunar"}
	got := normalizeCompatibilityProfile(p)
	if got.CalendarType != "lunar" {
		t.Errorf("expected lunar, got %q", got.CalendarType)
	}
}

func TestCompatibilityParticipantSummary_MissingSnapshot(t *testing.T) {
	p := &model.CompatibilityParticipant{DisplayName: "张三", ChartSnapshot: nil}
	_, err := compatibilityParticipantSummary(p)
	if err == nil {
		t.Fatal("expected error for missing chart_snapshot")
	}
}

func TestCompatibilityParticipantSummary_ValidSnapshot(t *testing.T) {
	p := &model.CompatibilityParticipant{
		DisplayName:   "李四",
		ChartSnapshot: makeCompatibilitySnapshot("李四", "male"),
	}
	summary, err := compatibilityParticipantSummary(p)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(summary, "李四") {
		t.Errorf("expected summary to contain display name, got: %s", summary)
	}
	if !strings.Contains(summary, "日主=") {
		t.Errorf("expected summary to contain 日主, got: %s", summary)
	}
}

func TestCompatibilityPromptFallback_ContainsTemplateVars(t *testing.T) {
	fb := compatibilityPromptFallback()
	for _, v := range []string{"{{.SelfLabel}}", "{{.PartnerLabel}}", "{{.ScoresJSON}}", "{{.DurationJSON}}"} {
		if !strings.Contains(fb, v) {
			t.Errorf("expected fallback prompt to contain %q", v)
		}
	}
}
