package bazi

import (
	"strings"
	"testing"
)

func findWealthEvidence(evidences []ProfileEvidence, source string) *ProfileEvidence {
	for i := range evidences {
		if evidences[i].Source == source {
			return &evidences[i]
		}
	}
	return nil
}

func TestWealthProfileOnCalculate(t *testing.T) {
	result := Calculate(1995, 10, 12, 12, "male", false, 0, "solar", false)
	if result.WealthProfile == nil {
		t.Fatal("WealthProfile should be attached to BaziResult")
	}
	profile := result.WealthProfile
	if profile.Version != WealthProfileVersion {
		t.Fatalf("wealth profile version = %q, want %q", profile.Version, WealthProfileVersion)
	}
	if profile.Score < 0 || profile.Score > 100 {
		t.Fatalf("wealth score should be within 0-100, got %d", profile.Score)
	}
	if profile.Grade == "" || profile.GradeLabel == "" || profile.WealthType == "" {
		t.Fatalf("wealth profile should expose grade, label, and type: %+v", profile)
	}
	if len(profile.Evidences) < 5 {
		t.Fatalf("wealth profile should expose evidence, got %d item(s)", len(profile.Evidences))
	}
	for _, source := range []string{"财星显露", "承载能力", "生财链路", "喜忌方向", "守财风险"} {
		if findWealthEvidence(profile.Evidences, source) == nil {
			t.Fatalf("missing wealth evidence source %q: %+v", source, profile.Evidences)
		}
	}
	for _, forbidden := range []string{"资产", "收入", "保证发财", "投资建议"} {
		if strings.Contains(profile.Summary, forbidden) {
			t.Fatalf("wealth summary should stay structural, found %q in %q", forbidden, profile.Summary)
		}
	}
}

func TestWealthProfileWeakCarryingCapsGrade(t *testing.T) {
	result := &BaziResult{
		YearGan: "戊", YearZhi: "辰",
		MonthGan: "己", MonthZhi: "丑",
		DayGan: "甲", DayZhi: "戌",
		HourGan: "戊", HourZhi: "未",
		YearHideGan:  []string{"戊", "乙", "癸"},
		MonthHideGan: []string{"己", "癸", "辛"},
		DayHideGan:   []string{"戊", "辛", "丁"},
		HourHideGan:  []string{"己", "丁", "乙"},
	}
	fuyi := AssessFuyiStrength(result)
	result.FavorableShishen, result.AdverseShishen, result.ShishenConfidence = BuildFavorableShishen(result.DayGan, fuyi.Yongshen, fuyi.Jishen, fuyi.Level)

	profile := BuildWealthProfile(result, 2026)
	if profile == nil {
		t.Fatal("expected wealth profile")
	}
	if gradeRank(profile.Grade) > gradeRank("B") {
		t.Fatalf("weak wealth-heavy chart should be capped at B, got %+v", profile)
	}
	if !contains(profile.RiskFlags, "财多身弱") && !contains(profile.RiskFlags, "身弱难承财") {
		t.Fatalf("expected weak carrying risk flag, got %+v", profile.RiskFlags)
	}
}

func TestWealthProfileChainsAndAdverseRisks(t *testing.T) {
	result := Calculate(1997, 12, 1, 12, "female", false, 0, "solar", false)
	profile := result.WealthProfile
	if profile == nil {
		t.Fatal("expected wealth profile")
	}
	if findWealthEvidence(profile.Evidences, "生财链路") == nil {
		t.Fatalf("expected chain evidence: %+v", profile.Evidences)
	}
	if profile.Grade == "" || profile.Summary == "" {
		t.Fatalf("expected readable wealth profile: %+v", profile)
	}
}

func TestCurrentWealthWindowHintDoesNotChangeNatalGrade(t *testing.T) {
	result := Calculate(1995, 10, 12, 12, "male", false, 0, "solar", false)
	profile := BuildWealthProfile(result, 2026)
	if profile == nil {
		t.Fatal("expected wealth profile")
	}
	grade := profile.Grade
	profile.CurrentHint = BuildCurrentWealthWindowHint(result, 2030)
	if profile.Grade != grade {
		t.Fatalf("current hint must not change natal grade: before=%s after=%s", grade, profile.Grade)
	}
	if profile.CurrentHint != nil && profile.CurrentHint.Year != 2030 {
		t.Fatalf("current hint should record evaluated year, got %+v", profile.CurrentHint)
	}
}
