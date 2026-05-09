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
