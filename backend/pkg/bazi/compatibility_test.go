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

// 保留 Tasks 7-8 已写的两个 evidence / explanation 测试
func TestBuildEvidences_ZodiacLiuhe(t *testing.T) {
	a := makeCompatNatal("甲子", "丙寅", "壬午", "丁未", "male")
	b := makeCompatNatal("己丑", "戊辰", "庚申", "辛酉", "female")
	ev := buildCompatibilityEvidencesV3(a, b)
	found := false
	for _, item := range ev {
		if item.EvidenceKey == "zodiac_liuhe" {
			if item.Weight != 50 || item.Dimension != "zodiac" || item.Polarity != "positive" {
				t.Errorf("zodiac_liuhe: bad shape %+v", item)
			}
			found = true
		}
	}
	if !found {
		t.Error("expected zodiac_liuhe evidence for 子/丑 pair")
	}
}

func TestBuildEvidences_AllHits_Count6(t *testing.T) {
	a := makeCompatNatal("甲子", "甲子", "甲子", "甲子", "male")
	b := makeCompatNatal("乙丑", "乙丑", "乙丑", "乙丑", "female")
	ev := buildCompatibilityEvidencesV3(a, b)
	if len(ev) != 6 {
		t.Errorf("all-hit case: got %d evidences, want 6", len(ev))
	}
}

func TestBuildScoreExplanationsV3_FourEntries(t *testing.T) {
	a := makeCompatNatal("甲子", "丙寅", "壬午", "丁未", "male")
	b := makeCompatNatal("己丑", "戊辰", "庚申", "辛酉", "female")
	ev := buildCompatibilityEvidencesV3(a, b)
	exps := buildScoreExplanationsV3(a, b, ev)
	if len(exps) != 4 {
		t.Fatalf("expected exactly 4 explanations (one per module), got %d", len(exps))
	}
	dims := map[string]bool{}
	for _, e := range exps {
		dims[string(e.Dimension)] = true
		if e.NegativeFactor != "" || len(e.NegativeEvidenceKeys) != 0 {
			t.Errorf("v3 should never set negative factors, got %+v", e)
		}
	}
	for _, d := range []string{"zodiac", "nayin", "day_pillar", "eight_chars"} {
		if !dims[d] {
			t.Errorf("missing dimension %q", d)
		}
	}
}

func TestBuildScoreExplanationsV3_UnHitModule_HasSummary(t *testing.T) {
	a := makeCompatNatal("甲午", "丙寅", "壬午", "丁未", "male")
	b := makeCompatNatal("乙未", "戊辰", "庚申", "辛酉", "female")
	ev := buildCompatibilityEvidencesV3(a, b)
	exps := buildScoreExplanationsV3(a, b, ev)
	for _, e := range exps {
		if e.Summary == "" {
			t.Errorf("dimension %q has empty summary", e.Dimension)
		}
	}
}

// AnalyzeCompatibility 现在是 stub — 仅验证类型可用
func TestAnalyzeCompatibility_Stub_ReturnsTypesOnly(t *testing.T) {
	a := makeCompatNatal("甲子", "丙寅", "甲子", "丁卯", "male")
	b := makeCompatNatal("己丑", "戊辰", "己丑", "庚申", "female")
	got := AnalyzeCompatibility(a, b)
	// stub 阶段允许 zero 值；只检查不 panic 且类型已实例化
	_ = got.DimensionScores.Zodiac
	_ = got.OverallScore
	_ = got.OverallLevel
}

// 占位避免 unused import: strings 包将在 Task 14b 引入更多用途
var _ = strings.Contains

func TestBuildSummaryTagsV3_AllHits(t *testing.T) {
	got := buildSummaryTagsV3(CompatibilityDimensionScores{
		Zodiac: 50, Nayin: 20, DayPillar: 10, EightChars: 20,
	}, 100)
	if !containsString(got, "上吉合盘") {
		t.Errorf("expected 上吉合盘 tag, got %v", got)
	}
	if !containsString(got, "属相相合") {
		t.Errorf("expected 属相相合 tag, got %v", got)
	}
}

func TestBuildSummaryTagsV3_AllMiss(t *testing.T) {
	got := buildSummaryTagsV3(CompatibilityDimensionScores{}, 0)
	if !containsString(got, "合盘无加成") {
		t.Errorf("expected 合盘无加成 tag, got %v", got)
	}
}

func TestBuildSummaryTagsV3_MaxFour(t *testing.T) {
	got := buildSummaryTagsV3(CompatibilityDimensionScores{
		Zodiac: 50, Nayin: 20, DayPillar: 10, EightChars: 20,
	}, 100)
	if len(got) > 4 {
		t.Errorf("tags exceeded 4: %v", got)
	}
}

func containsString(slice []string, s string) bool {
	for _, x := range slice {
		if x == s {
			return true
		}
	}
	return false
}
