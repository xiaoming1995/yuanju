package service

import (
	"encoding/json"
	"strings"
	"testing"
	"yuanju/internal/model"
	"yuanju/pkg/bazi"
	"yuanju/pkg/prompt"
)

func makeCompatibilitySnapshot(displayName, gender string) *json.RawMessage {
	snapshot, _ := json.Marshal(bazi.BaziResult{
		YearGan: "甲", YearZhi: "子",
		MonthGan: "丙", MonthZhi: "寅",
		DayGan: "甲", DayZhi: "子",
		HourGan: "丁", HourZhi: "卯",
		Gender:   gender,
		Wuxing:   bazi.WuxingStats{Mu: 2, Huo: 2, Tu: 1, Jin: 1, Shui: 2},
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

func TestBuildCompatibilityPromptData_EmbedsConsultingAssessment(t *testing.T) {
	detail := &model.CompatibilityDetail{
		Reading: &model.CompatibilityReading{
			DimensionScores: model.CompatibilityDimensionScores{
				Attraction:    78,
				Stability:     54,
				Communication: 66,
				Practicality:  48,
			},
			DurationAssessment: model.CompatibilityDurationAssessment{
				OverallBand: "medium_term",
				Windows: model.CompatibilityDurationWindows{
					ThreeMonths:  model.CompatibilityDurationWindow{Level: "high"},
					OneYear:      model.CompatibilityDurationWindow{Level: "medium"},
					TwoYearsPlus: model.CompatibilityDurationWindow{Level: "low"},
				},
				Summary: "前期吸引强，但长期承压。",
				Reasons: []string{"夫妻宫冲克明显"},
			},
			ConsultingAssessment: model.CompatibilityConsultingAssessment{
				RelationshipDiagnosis: model.CompatibilityRelationshipDiagnosis{
					RelationshipType: "短期吸引强、长期承压型",
					Verdict:          "建议谨慎观察",
					Summary:          "先观察冲突修复能力。",
					TopFindings: []model.CompatibilityFinding{
						{Text: "吸引与稳定分化", EvidenceKeys: []string{"spouse_palace_stability_spouse_palace_chong"}},
					},
				},
				DecisionAdvice: model.CompatibilityDecisionAdvice{Recommendation: "observe", Confidence: "medium"},
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
	if !strings.Contains(got.ConsultingJSON, `"relationship_type":"短期吸引强、长期承压型"`) {
		t.Fatalf("expected consulting json in prompt data, got %s", got.ConsultingJSON)
	}
}

func TestNormalizeCompatibilityContext_DefaultsAndFallbacks(t *testing.T) {
	got := normalizeCompatibilityContext(model.CompatibilityContext{})
	if got.RelationshipStage != "general" {
		t.Fatalf("expected default relationship stage general, got %q", got.RelationshipStage)
	}
	if got.PrimaryQuestion != "general" {
		t.Fatalf("expected default primary question general, got %q", got.PrimaryQuestion)
	}

	got = normalizeCompatibilityContext(model.CompatibilityContext{
		RelationshipStage: "reconciliation",
		PrimaryQuestion:   "marriage_suitability",
	})
	if got.RelationshipStage != "reconciliation" || got.PrimaryQuestion != "marriage_suitability" {
		t.Fatalf("expected valid context to be preserved, got %+v", got)
	}

	got = normalizeCompatibilityContext(model.CompatibilityContext{
		RelationshipStage: "unknown-stage",
		PrimaryQuestion:   "unknown-question",
	})
	if got.RelationshipStage != "general" || got.PrimaryQuestion != "general" {
		t.Fatalf("expected unknown context to fall back to general, got %+v", got)
	}
}

func TestBuildCompatibilityPromptData_EmbedsRelationshipContext(t *testing.T) {
	detail := &model.CompatibilityDetail{
		Reading: &model.CompatibilityReading{
			RelationshipStage: "reconciliation",
			PrimaryQuestion:   "reconciliation_potential",
			DimensionScores: model.CompatibilityDimensionScores{
				Attraction:    78,
				Stability:     54,
				Communication: 66,
				Practicality:  48,
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
	if got.RelationshipStage != "reconciliation" {
		t.Fatalf("expected relationship stage in prompt data, got %q", got.RelationshipStage)
	}
	if got.PrimaryQuestion != "reconciliation_potential" {
		t.Fatalf("expected primary question in prompt data, got %q", got.PrimaryQuestion)
	}
	if got.RelationshipStageLabel == "" || got.PrimaryQuestionLabel == "" || got.QuestionGuidance == "" {
		t.Fatalf("expected context labels and guidance, got %+v", got)
	}
}

func TestBuildCompatibilityPromptData_EmbedsDepthEvidenceAndScoreExplanations(t *testing.T) {
	detail := &model.CompatibilityDetail{
		Reading: &model.CompatibilityReading{
			DimensionScores: model.CompatibilityDimensionScores{Attraction: 72, Stability: 58, Communication: 64, Practicality: 66},
			ScoreExplanations: []model.CompatibilityScoreExplanation{
				{
					Dimension:            "stability",
					PositiveFactor:       "夫妻宫六合",
					NegativeFactor:       "地支相冲",
					PositiveEvidenceKeys: []string{"positive_key"},
					NegativeEvidenceKeys: []string{"negative_key"},
					Summary:              "稳定度同时有支撑和压力。",
				},
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
		Evidences: []model.CompatibilityEvidence{
			{
				EvidenceKey:    "ten_god_key",
				Dimension:      "communication",
				Type:           "十神互动-印星",
				Polarity:       "positive",
				Source:         "ten_god_interaction",
				Perspective:    "self_to_partner",
				Actor:          "self",
				Target:         "partner",
				RelatedSources: []string{"day_master"},
				Title:          "支持与照拂感",
				Detail:         "directional evidence",
				Weight:         6,
			},
		},
	}

	got, err := buildCompatibilityPromptData(detail)
	if err != nil {
		t.Fatalf("build prompt data: %v", err)
	}
	if !strings.Contains(got.ScoreExplanationsJSON, "稳定度同时有支撑和压力") {
		t.Fatalf("expected score explanations in prompt data, got %s", got.ScoreExplanationsJSON)
	}
	if !strings.Contains(got.EvidenceGroupsJSON, "ten_god_interaction") {
		t.Fatalf("expected grouped depth evidence in prompt data, got %s", got.EvidenceGroupsJSON)
	}
	if !strings.Contains(got.EvidencesJSON, "self_to_partner") {
		t.Fatalf("expected directional metadata in evidences JSON, got %s", got.EvidencesJSON)
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

func TestEnsureCompatibilityConsultingAssessment_BackfillsMissingConsulting(t *testing.T) {
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

func TestEnsureCompatibilityEvidenceKeys_BackfillsMissingEvidenceKeys(t *testing.T) {
	selfSnapshot := makeCompatibilitySnapshot("我", "male")
	partnerSnapshot := makeCompatibilitySnapshot("对方", "female")
	var selfResult, partnerResult bazi.BaziResult
	if err := json.Unmarshal(*selfSnapshot, &selfResult); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(*partnerSnapshot, &partnerResult); err != nil {
		t.Fatal(err)
	}
	analysis := bazi.AnalyzeCompatibility(&selfResult, &partnerResult)
	if len(analysis.Evidences) == 0 {
		t.Fatal("expected generated evidence")
	}
	generated := analysis.Evidences[0]
	detail := &model.CompatibilityDetail{
		Reading: &model.CompatibilityReading{},
		Participants: []model.CompatibilityParticipant{
			{
				Role:          "self",
				DisplayName:   "我",
				ChartSnapshot: selfSnapshot,
			},
			{
				Role:          "partner",
				DisplayName:   "对方",
				ChartSnapshot: partnerSnapshot,
			},
		},
		Evidences: []model.CompatibilityEvidence{
			{
				ID:        "ev-1",
				Dimension: string(generated.Dimension),
				Type:      generated.Type,
				Polarity:  string(generated.Polarity),
				Source:    generated.Source,
				Title:     generated.Title,
				Detail:    generated.Detail,
				Weight:    generated.Weight,
			},
		},
	}

	changed, err := ensureCompatibilityEvidenceKeys(detail)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("expected evidence key backfill to report a change")
	}
	if detail.Evidences[0].EvidenceKey != generated.EvidenceKey {
		t.Fatalf("expected %q, got %q", generated.EvidenceKey, detail.Evidences[0].EvidenceKey)
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

func TestCompatibilityCanonical_ContainsTemplateVars(t *testing.T) {
	fb := prompt.MustGet("compatibility").Content
	for _, v := range []string{
		"{{.SelfLabel}}",
		"{{.PartnerLabel}}",
		"{{.RelationshipStageLabel}}",
		"{{.PrimaryQuestionLabel}}",
		"{{.QuestionGuidance}}",
		"{{.ScoresJSON}}",
		"{{.ScoreExplanationsJSON}}",
		"{{.DurationJSON}}",
		"{{.ConsultingJSON}}",
		"{{.EvidenceGroupsJSON}}",
	} {
		if !strings.Contains(fb, v) {
			t.Errorf("expected fallback prompt to contain %q", v)
		}
	}
	if !strings.Contains(fb, `"question_focus"`) {
		t.Errorf("expected fallback prompt to require question_focus output")
	}
	if !strings.Contains(fb, "不得输出具体结婚、分手、复合、出轨、怀孕等确定事件日期") {
		t.Fatal("expected fallback prompt to prohibit deterministic event dates")
	}
}

func TestCompatibilityCanonical_DefinesQuestionSpecificBranches(t *testing.T) {
	fb := prompt.MustGet("compatibility").Content
	for _, want := range []string{
		"primary_question = reconciliation_potential",
		"是否建议复合",
		"primary_question = marriage_suitability",
		"谈婚前必须确认",
		"primary_question = continue_investment",
		"是否继续投入",
	} {
		if !strings.Contains(fb, want) {
			t.Fatalf("expected question-specific prompt branch %q", want)
		}
	}
}
