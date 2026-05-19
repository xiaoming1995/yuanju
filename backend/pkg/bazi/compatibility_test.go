package bazi

import (
	"strings"
	"testing"
)

func makeCompatNatal(yearGZ, monthGZ, dayGZ, hourGZ, gender string) *BaziResult {
	yr := []rune(yearGZ)
	mr := []rune(monthGZ)
	dr := []rune(dayGZ)
	hr := []rune(hourGZ)
	return &BaziResult{
		YearGan: string(yr[0]), YearZhi: string(yr[1]),
		MonthGan: string(mr[0]), MonthZhi: string(mr[1]),
		DayGan: string(dr[0]), DayZhi: string(dr[1]),
		HourGan: string(hr[0]), HourZhi: string(hr[1]),
		YearGanWuxing:  ganWuxing[string(yr[0])],
		MonthGanWuxing: ganWuxing[string(mr[0])],
		DayGanWuxing:   ganWuxing[string(dr[0])],
		HourGanWuxing:  ganWuxing[string(hr[0])],
		YearZhiWuxing:  zhiWuxing[string(yr[1])],
		MonthZhiWuxing: zhiWuxing[string(mr[1])],
		DayZhiWuxing:   zhiWuxing[string(dr[1])],
		HourZhiWuxing:  zhiWuxing[string(hr[1])],
		YearHideGan:    []string{string(yr[0])},
		MonthHideGan:   []string{string(mr[0])},
		DayHideGan:     []string{string(dr[0])},
		HourHideGan:    []string{string(hr[0])},
		Gender:         gender,
	}
}

func hasCompatEvidence(items []CompatibilityEvidence, evType string, polarity CompatibilityPolarity) bool {
	for _, item := range items {
		if item.Type == evType && item.Polarity == polarity {
			return true
		}
	}
	return false
}

func TestAnalyzeCompatibility_ReturnsCoreShape(t *testing.T) {
	a := makeCompatNatal("甲子", "丙寅", "甲子", "丁卯", "male")
	b := makeCompatNatal("己丑", "戊辰", "己丑", "庚申", "female")

	got := AnalyzeCompatibility(a, b)

	if got.OverallLevel == "" {
		t.Fatal("expected overall level")
	}
	if got.DimensionScores.Attraction == 0 ||
		got.DimensionScores.Stability == 0 ||
		got.DimensionScores.Communication == 0 ||
		got.DimensionScores.Practicality == 0 {
		t.Fatalf("expected all four dimensions to be populated, got %+v", got.DimensionScores)
	}
	if len(got.Evidences) == 0 {
		t.Fatal("expected compatibility evidences")
	}
	if len(got.SummaryTags) == 0 {
		t.Fatal("expected summary tags for history preview")
	}
}

func TestAnalyzeCompatibility_DayZhiLiuHeAddsPositiveStabilityEvidence(t *testing.T) {
	a := makeCompatNatal("甲子", "丙寅", "甲子", "丁卯", "male")
	b := makeCompatNatal("己丑", "戊辰", "辛丑", "庚申", "female")

	got := AnalyzeCompatibility(a, b)

	if !hasCompatEvidence(got.Evidences, "夫妻宫六合", CompatibilityPositive) {
		t.Fatalf("expected positive day-branch liuhe evidence, got %+v", got.Evidences)
	}
	if got.DimensionScores.Stability < 60 {
		t.Fatalf("expected liuhe to keep stability at or above baseline, got %d", got.DimensionScores.Stability)
	}
}

func TestAnalyzeCompatibility_DayZhiChongCreatesNegativeStabilityEvidence(t *testing.T) {
	a := makeCompatNatal("甲子", "丙寅", "甲子", "丁卯", "male")
	b := makeCompatNatal("己午", "戊午", "辛午", "庚申", "female")

	got := AnalyzeCompatibility(a, b)

	if !hasCompatEvidence(got.Evidences, "夫妻宫六冲", CompatibilityNegative) {
		t.Fatalf("expected negative day-branch clash evidence, got %+v", got.Evidences)
	}
	if got.DimensionScores.Stability >= 60 {
		t.Fatalf("expected clash to reduce stability below baseline, got %d", got.DimensionScores.Stability)
	}
	if got.OverallLevel == CompatibilityHigh {
		t.Fatalf("expected clash case not to be high compatibility, got %s", got.OverallLevel)
	}
}

func TestAnalyzeCompatibility_SpouseStarResonanceBoostsAttraction(t *testing.T) {
	// 甲日男命以土为配偶星；对方多土，应该增强吸引力。
	a := makeCompatNatal("甲子", "丙寅", "甲子", "丁卯", "male")
	b := makeCompatNatal("己丑", "戊辰", "己未", "戊戌", "female")

	got := AnalyzeCompatibility(a, b)

	if !hasCompatEvidence(got.Evidences, "配偶星呼应", CompatibilityPositive) {
		t.Fatalf("expected spouse-star resonance evidence, got %+v", got.Evidences)
	}
	if got.DimensionScores.Attraction <= 60 {
		t.Fatalf("expected spouse-star resonance to boost attraction above baseline, got %d", got.DimensionScores.Attraction)
	}
}

func TestAnalyzeCompatibility_ReturnsDurationAssessmentShape(t *testing.T) {
	a := makeCompatNatal("甲子", "丙寅", "甲子", "丁卯", "male")
	b := makeCompatNatal("己丑", "戊辰", "己丑", "庚申", "female")

	got := AnalyzeCompatibility(a, b)

	if got.DurationAssessment.OverallBand == "" {
		t.Fatal("expected duration overall band")
	}
	if got.DurationAssessment.Windows.ThreeMonths.Level == "" {
		t.Fatal("expected three-month window")
	}
	if got.DurationAssessment.Windows.OneYear.Level == "" {
		t.Fatal("expected one-year window")
	}
	if got.DurationAssessment.Windows.TwoYearsPlus.Level == "" {
		t.Fatal("expected two-years-plus window")
	}
	if len(got.DurationAssessment.Reasons) == 0 {
		t.Fatal("expected duration reasons")
	}
}

func TestAnalyzeCompatibility_DurationCanBeStrongShortTermButWeakLongTerm(t *testing.T) {
	a := makeCompatNatal("甲子", "丙寅", "甲子", "丁卯", "male")
	b := makeCompatNatal("己午", "戊午", "己午", "戊申", "female")

	got := AnalyzeCompatibility(a, b)

	if got.DurationAssessment.Windows.ThreeMonths.Level == "" || got.DurationAssessment.Windows.TwoYearsPlus.Level == "" {
		t.Fatal("expected both short and long windows")
	}
	if got.DurationAssessment.Windows.ThreeMonths.Level == got.DurationAssessment.Windows.TwoYearsPlus.Level {
		t.Fatalf("expected staged difference, got %+v", got.DurationAssessment.Windows)
	}
}

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
	for _, finding := range got.ConsultingAssessment.RelationshipDiagnosis.TopFindings {
		for _, key := range finding.EvidenceKeys {
			if !keys[key] {
				t.Fatalf("top finding references missing evidence key %q", key)
			}
		}
	}
	for _, risk := range got.ConsultingAssessment.StageRisks {
		for _, key := range risk.EvidenceKeys {
			if !keys[key] {
				t.Fatalf("stage risk references missing evidence key %q", key)
			}
		}
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

func TestAnalyzeCompatibility_EvidenceKeysAreASCIIAndSemantic(t *testing.T) {
	a := makeCompatNatal("甲子", "丙寅", "甲子", "丁卯", "male")
	b := makeCompatNatal("己午", "戊午", "己午", "戊申", "female")

	got := AnalyzeCompatibility(a, b)

	for _, evidence := range got.Evidences {
		if !isASCII(evidence.EvidenceKey) {
			t.Fatalf("expected ASCII evidence key, got %q", evidence.EvidenceKey)
		}
		if strings.Contains(evidence.EvidenceKey, evidence.Type) || strings.Contains(evidence.EvidenceKey, evidence.Title) {
			t.Fatalf("expected evidence key not to contain display text, got key %q for %+v", evidence.EvidenceKey, evidence)
		}
	}
}

func TestBuildCompatibilityConsultingAssessment_WithoutEvidenceHasNoClaimEvidenceLinks(t *testing.T) {
	got := buildCompatibilityConsultingAssessment(
		CompatibilityDimensionScores{Attraction: 60, Stability: 60, Communication: 60, Practicality: 60},
		nil,
		CompatibilityDurationAssessment{},
	)

	if len(got.ClaimEvidenceLinks) != 0 {
		t.Fatalf("expected no claim evidence links without evidence, got %+v", got.ClaimEvidenceLinks)
	}
}

func isASCII(value string) bool {
	for _, r := range value {
		if r > 127 {
			return false
		}
	}
	return true
}
