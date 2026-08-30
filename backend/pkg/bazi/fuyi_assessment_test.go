package bazi

import (
	"strings"
	"testing"
)

func TestAssessFuyiStrength_1995BingFireUsesWaterAndMetal(t *testing.T) {
	// 1995-10-12 午时：乙亥 丙戌 丙子 甲午。
	// 透出正偏印与比肩，午中劫财且为丙火羊刃；尽管亥子官杀、戌土食伤存在，
	// 完整原局仍应判为身极强，扶抑用水金。
	natal := &BaziResult{
		YearGan: "乙", YearZhi: "亥",
		MonthGan: "丙", MonthZhi: "戌",
		DayGan: "丙", DayZhi: "子",
		HourGan: "甲", HourZhi: "午",
		YearHideGan:  []string{"壬", "甲"},
		MonthHideGan: []string{"戊", "辛", "丁"},
		DayHideGan:   []string{"癸"},
		HourHideGan:  []string{"丁", "己"},
	}

	assessment := AssessFuyiStrength(natal)
	if assessment.Level != "vstrong" {
		t.Fatalf("expected vstrong, got %s (score=%d, reason=%s)", assessment.Level, assessment.Score, assessment.Reason)
	}
	if assessment.Yongshen != "水金" {
		t.Fatalf("expected global Fuyi yongshen 水金, got %q", assessment.Yongshen)
	}
	if assessment.Jishen != "木火" {
		t.Fatalf("expected global Fuyi jishen 木火, got %q", assessment.Jishen)
	}
	if !strings.Contains(assessment.Reason, "羊刃+12") {
		t.Fatalf("expected Yangren evidence, got %s", assessment.Reason)
	}
}

func TestNatalAssessmentUsesGlobalFuyiInsteadOfLegacyTiaohou(t *testing.T) {
	natal := &BaziResult{
		YearGan: "乙", YearZhi: "亥",
		MonthGan: "丙", MonthZhi: "戌",
		DayGan: "丙", DayZhi: "子",
		HourGan: "甲", HourZhi: "午",
		YearHideGan:  []string{"壬", "甲"},
		MonthHideGan: []string{"戊", "辛", "丁"},
		DayHideGan:   []string{"癸"},
		HourHideGan:  []string{"丁", "己"},
		Yongshen:     "木水", // legacy Tiaohou result remains a distinct field
		Jishen:       "火土金",
	}

	assessment := AssessNatalStructure(natal)
	if assessment.Fuyi.DayMasterStrength != "vstrong" || assessment.Fuyi.Yongshen != "水金" {
		t.Fatalf("expected global Fuyi assessment, got %+v", assessment.Fuyi)
	}
	if natal.Yongshen != "木水" {
		t.Fatalf("legacy Tiaohou field must remain distinct, got %q", natal.Yongshen)
	}
}

func TestEnsureNatalAssessmentRebuildsEarlierFuyiSnapshot(t *testing.T) {
	natal := &BaziResult{
		YearGan: "乙", YearZhi: "亥",
		MonthGan: "丙", MonthZhi: "戌",
		DayGan: "丙", DayZhi: "子",
		HourGan: "甲", HourZhi: "午",
		YearHideGan:  []string{"壬", "甲"},
		MonthHideGan: []string{"戊", "辛", "丁"},
		DayHideGan:   []string{"癸"},
		HourHideGan:  []string{"丁", "己"},
		NatalAssessment: &NatalAssessment{
			Version: "v6",
			Fuyi:    NatalFuyiAssessment{Yongshen: "木火"},
		},
	}

	if !EnsureNatalAssessment(natal) {
		t.Fatal("expected stale assessment to be rebuilt")
	}
	if natal.NatalAssessment.Version != NatalAssessmentVersion || natal.NatalAssessment.Fuyi.Yongshen != "水金" {
		t.Fatalf("expected current global Fuyi snapshot, got %+v", natal.NatalAssessment)
	}
}

func TestFuyiMonthCommandRelationshipEvidence(t *testing.T) {
	cases := []struct {
		name       string
		dayGan     string
		monthZhi   string
		wantDelta  int
		wantDetail string
	}{
		{name: "same element gains command", dayGan: "甲", monthZhi: "寅", wantDelta: monthCommandSameElementScore, wantDetail: "月令木得令"},
		{name: "month command generates day", dayGan: "丙", monthZhi: "寅", wantDelta: monthCommandGeneratesDayScore, wantDetail: "月令木生身"},
		{name: "month command drains day", dayGan: "甲", monthZhi: "巳", wantDelta: monthCommandDrainsDayScore, wantDetail: "月令火泄身"},
		{name: "month command controls day", dayGan: "丙", monthZhi: "子", wantDelta: monthCommandControlsDayScore, wantDetail: "月令水克身"},
		{name: "day controls month command", dayGan: "甲", monthZhi: "辰", wantDelta: monthCommandControlledByDayScore, wantDetail: "月令土受制"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			delta, detail := fuyiMonthCommandDelta(ganWuxing[tc.dayGan], zhiWuxing[tc.monthZhi])
			if delta != tc.wantDelta || !strings.Contains(detail, tc.wantDetail) {
				t.Fatalf("month command = %d/%q, want %d containing %q", delta, detail, tc.wantDelta, tc.wantDetail)
			}
		})
	}
}
