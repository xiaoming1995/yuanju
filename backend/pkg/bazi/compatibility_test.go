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

func TestClassifyRelationshipType_AllBranches(t *testing.T) {
	cases := []struct {
		total, zodiac, dayPillar, eightChars int
		want                                  string
	}{
		{85, 50, 10, 20, "高契合型"},
		{75, 50, 10, 0, "亲密层稳固型"},
		{75, 50, 0, 0, "属相吸引型"},
		{55, 0, 10, 0, "亲密外围支撑型"},
		{55, 0, 0, 14, "亲密外围支撑型"},
		{30, 0, 0, 0, "合盘无加成"},
	}
	for _, tc := range cases {
		got := classifyRelationshipTypeV3(tc.total, CompatibilityDimensionScores{
			Zodiac: tc.zodiac, DayPillar: tc.dayPillar, EightChars: tc.eightChars,
		})
		if got != tc.want {
			t.Errorf("total=%d zodiac=%d day=%d 8chars=%d → got %q, want %q",
				tc.total, tc.zodiac, tc.dayPillar, tc.eightChars, got, tc.want)
		}
	}
}

func TestBuildDecisionAdviceV3_AllBranches(t *testing.T) {
	cases := []struct {
		total, hitsCount int
		recommendation   string
		verdict          string
		confidence       string
	}{
		{85, 4, "continue", "适合继续推进", "high"},
		{70, 2, "observe", "建议谨慎观察", "medium"},
		{50, 1, "caution", "不宜过早重投入", "medium"},
		{40, 0, "caution", "不宜过早重投入", "low"},
	}
	for _, tc := range cases {
		adv := buildDecisionAdviceV3(tc.total, tc.hitsCount)
		if adv.Recommendation != tc.recommendation || adv.Verdict != tc.verdict || adv.Confidence != tc.confidence {
			t.Errorf("total=%d hits=%d: got rec=%q verdict=%q conf=%q, want %q/%q/%q",
				tc.total, tc.hitsCount,
				adv.Recommendation, adv.Verdict, adv.Confidence,
				tc.recommendation, tc.verdict, tc.confidence)
		}
	}
}

func TestBuildDurationAssessmentV3_Branches(t *testing.T) {
	cases := []struct {
		name                                 string
		zodiac, nayin, dayPillar, eightChars int
		wantShort, wantMid, wantLong         CompatibilityDurationLevel
	}{
		{"all high",
			50, 20, 10, 20,
			CompatibilityDurationHigh, CompatibilityDurationHigh, CompatibilityDurationHigh},
		{"zodiac+nayin only",
			50, 20, 0, 0,
			CompatibilityDurationHigh, CompatibilityDurationLow, CompatibilityDurationLow},
		{"day_pillar lower with zodiac",
			50, 0, 5, 0,
			CompatibilityDurationMedium, CompatibilityDurationHigh, CompatibilityDurationLow},
		{"eight_chars strong only",
			0, 0, 0, 17,
			CompatibilityDurationLow, CompatibilityDurationLow, CompatibilityDurationMedium},
		{"all miss",
			0, 0, 0, 0,
			CompatibilityDurationLow, CompatibilityDurationLow, CompatibilityDurationLow},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			da := buildDurationAssessmentV3(CompatibilityDimensionScores{
				Zodiac: tc.zodiac, Nayin: tc.nayin,
				DayPillar: tc.dayPillar, EightChars: tc.eightChars,
			})
			if da.Windows.ThreeMonths.Level != tc.wantShort {
				t.Errorf("short: got %q, want %q", da.Windows.ThreeMonths.Level, tc.wantShort)
			}
			if da.Windows.OneYear.Level != tc.wantMid {
				t.Errorf("mid: got %q, want %q", da.Windows.OneYear.Level, tc.wantMid)
			}
			if da.Windows.TwoYearsPlus.Level != tc.wantLong {
				t.Errorf("long: got %q, want %q", da.Windows.TwoYearsPlus.Level, tc.wantLong)
			}
		})
	}
}

func TestBuildStageRisksV3_AllWindowsHaveLevelAndKeys(t *testing.T) {
	evidences := []CompatibilityEvidence{
		{EvidenceKey: "zodiac_liuhe", Dimension: "zodiac"},
		{EvidenceKey: "nayin_sheng", Dimension: "nayin"},
		{EvidenceKey: "day_pillar_upper", Dimension: "day_pillar"},
		{EvidenceKey: "eight_chars_year_upper", Dimension: "eight_chars"},
	}
	duration := CompatibilityDurationAssessment{
		Windows: CompatibilityDurationWindows{
			ThreeMonths:  CompatibilityDurationWindow{Level: CompatibilityDurationHigh},
			OneYear:      CompatibilityDurationWindow{Level: CompatibilityDurationMedium},
			TwoYearsPlus: CompatibilityDurationWindow{Level: CompatibilityDurationLow},
		},
	}
	risks := buildStageRisksV3(duration, evidences)
	if len(risks) != 3 {
		t.Fatalf("expected 3 windows, got %d", len(risks))
	}
	for _, r := range risks {
		if r.RiskLevel == "" || r.MainRisk == "" || r.Trigger == "" || r.Advice == "" {
			t.Errorf("incomplete risk: %+v", r)
		}
	}
	if !containsString(risks[0].EvidenceKeys, "zodiac_liuhe") &&
		!containsString(risks[0].EvidenceKeys, "nayin_sheng") {
		t.Error("3-month window should reference zodiac/nayin evidence")
	}
	if !containsString(risks[1].EvidenceKeys, "day_pillar_upper") {
		t.Error("1-year window should reference day_pillar evidence")
	}
	if !containsString(risks[2].EvidenceKeys, "eight_chars_year_upper") {
		t.Error("2-year+ window should reference eight_chars evidence")
	}
}

func TestBuildRelationshipStrategyV3_ThreeTiers(t *testing.T) {
	for _, rec := range []string{"continue", "observe", "caution"} {
		s := buildRelationshipStrategyV3(rec)
		if s.Communication == "" || s.Conflict == "" || s.Reality == "" || s.Boundary == "" {
			t.Errorf("recommendation %q produced empty strategy: %+v", rec, s)
		}
	}
}

func TestBuildClaimEvidenceLinksV3_FromHits(t *testing.T) {
	evidences := []CompatibilityEvidence{
		{EvidenceKey: "zodiac_liuhe"},
		{EvidenceKey: "day_pillar_upper"},
	}
	links := buildClaimEvidenceLinksV3("适合继续推进", evidences)
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
	if links[0].ClaimID != "relationship_main_judgement" {
		t.Errorf("bad ClaimID: %q", links[0].ClaimID)
	}
	if len(links[0].EvidenceKeys) != 2 {
		t.Errorf("expected 2 evidence keys, got %v", links[0].EvidenceKeys)
	}
}

func TestBuildClaimEvidenceLinksV3_NoEvidence_EmptyResult(t *testing.T) {
	links := buildClaimEvidenceLinksV3("不宜过早重投入", nil)
	if len(links) != 0 {
		t.Errorf("no evidence should produce no link, got %d", len(links))
	}
}

func TestBuildConsultingAssessmentV3_Integration(t *testing.T) {
	scores := CompatibilityDimensionScores{Zodiac: 50, Nayin: 20, DayPillar: 10, EightChars: 17}
	total := 97
	hits := 4
	evidences := []CompatibilityEvidence{
		{EvidenceKey: "zodiac_liuhe", Dimension: "zodiac"},
		{EvidenceKey: "nayin_sheng", Dimension: "nayin"},
		{EvidenceKey: "day_pillar_upper", Dimension: "day_pillar"},
		{EvidenceKey: "eight_chars_year_upper", Dimension: "eight_chars"},
	}
	duration := buildDurationAssessmentV3(scores)
	got := buildConsultingAssessmentV3(total, hits, scores, evidences, duration)
	if got.RelationshipDiagnosis.RelationshipType != "高契合型" {
		t.Errorf("bad relationship type: %q", got.RelationshipDiagnosis.RelationshipType)
	}
	if got.DecisionAdvice.Recommendation != "continue" {
		t.Errorf("bad recommendation: %q", got.DecisionAdvice.Recommendation)
	}
	if len(got.StageRisks) != 3 {
		t.Errorf("expected 3 stage risks, got %d", len(got.StageRisks))
	}
	if got.RelationshipStrategy.Communication == "" {
		t.Error("missing strategy")
	}
}
