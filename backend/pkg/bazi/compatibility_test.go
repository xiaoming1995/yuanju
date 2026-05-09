package bazi

import "testing"

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
