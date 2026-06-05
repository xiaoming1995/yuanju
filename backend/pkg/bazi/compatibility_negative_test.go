package bazi

import (
	"sort"
	"strings"
	"testing"
)

// 用四柱干支构造一个最小 BaziResult（只填检测需要的字段）。
func makeChart(yg, yz, mg, mz, dg, dz, hg, hz string) *BaziResult {
	return &BaziResult{
		YearGan: yg, YearZhi: yz,
		MonthGan: mg, MonthZhi: mz,
		DayGan: dg, DayZhi: dz,
		HourGan: hg, HourZhi: hz,
	}
}

// 收集检测结果里的 EvidenceKey，排序后便于断言。
func negKeys(evs []CompatibilityEvidence) []string {
	keys := make([]string, 0, len(evs))
	for _, e := range evs {
		keys = append(keys, e.EvidenceKey)
	}
	sort.Strings(keys)
	return keys
}

func TestDetectNegativeSignals(t *testing.T) {
	cases := []struct {
		name     string
		a, b     *BaziResult
		wantKeys []string
	}{
		{
			// A 乙亥日 / B 己巳日 → 日柱 巳亥相冲 + 乙(木)克己(土)；其余柱无冲克刑害
			name:     "day pillar 巳亥冲 + 乙克己",
			a:        makeChart("丙", "子", "庚", "寅", "乙", "亥", "丙", "戌"),
			b:        makeChart("乙", "亥", "己", "丑", "己", "巳", "甲", "申"),
			wantKeys: []string{"neg_day_chong", "neg_day_gan_ke"},
		},
		{
			// 日柱地支相刑：丑 刑 戌（无恩之刑），非六害亦非六冲；其余柱用子子（安全填充）
			name:     "day pillar 丑戌相刑",
			a:        makeChart("甲", "子", "甲", "子", "甲", "丑", "甲", "子"),
			b:        makeChart("甲", "子", "甲", "子", "甲", "戌", "甲", "子"),
			wantKeys: []string{"neg_day_xing"},
		},
		{
			// 日柱地支自刑：午 午；其余柱用子子（安全填充）
			name:     "day pillar 午午自刑",
			a:        makeChart("甲", "子", "甲", "子", "甲", "午", "甲", "子"),
			b:        makeChart("甲", "子", "甲", "子", "甲", "午", "甲", "子"),
			wantKeys: []string{"neg_day_xing"},
		},
		{
			// 月柱地支相害（穿）：子 未；其余柱用子子（安全填充）
			name:     "month pillar 子未相害",
			a:        makeChart("甲", "子", "甲", "子", "甲", "子", "甲", "子"),
			b:        makeChart("甲", "子", "甲", "未", "甲", "子", "甲", "子"),
			wantKeys: []string{"neg_month_hai"},
		},
		{
			// 全部为合/无关，无负面信号
			name:     "no negatives",
			a:        makeChart("甲", "子", "甲", "子", "甲", "子", "甲", "子"),
			b:        makeChart("甲", "子", "甲", "子", "甲", "子", "甲", "子"),
			wantKeys: []string{},
		},
		{
			name:     "nil safe",
			a:        nil,
			b:        makeChart("甲", "子", "甲", "子", "甲", "子", "甲", "子"),
			wantKeys: []string{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := negKeys(detectNegativeSignals(tc.a, tc.b))
			if strings.Join(got, ",") != strings.Join(tc.wantKeys, ",") {
				t.Errorf("keys = %v, want %v", got, tc.wantKeys)
			}
		})
	}
}

func TestDetectNegativeSignalsArePolarityNegative(t *testing.T) {
	a := makeChart("丙", "子", "庚", "寅", "乙", "亥", "丙", "戌")
	b := makeChart("乙", "亥", "己", "丑", "己", "巳", "壬", "申")
	evs := detectNegativeSignals(a, b)
	if len(evs) == 0 {
		t.Fatal("expected negative evidences, got none")
	}
	for _, e := range evs {
		if e.Polarity != "negative" {
			t.Errorf("evidence %s polarity = %q, want negative", e.EvidenceKey, e.Polarity)
		}
		if e.Dimension == "" || e.Title == "" || e.Detail == "" {
			t.Errorf("evidence %s missing fields: %+v", e.EvidenceKey, e)
		}
	}
}

// 集成断言：触发案例的负面信号进入 Evidences，且评分丝毫未动（总分仍 34）。
func TestAnalyzeCompatibilitySurfacesNegativesWithoutScoreChange(t *testing.T) {
	a := makeChart("丙", "子", "庚", "寅", "乙", "亥", "丙", "戌")
	b := makeChart("乙", "亥", "己", "丑", "己", "巳", "壬", "申")
	res := AnalyzeCompatibility(a, b)

	// 评分未被触碰
	if res.OverallScore != 34 {
		t.Errorf("OverallScore = %d, want 34 (评分不应被负面信号改动)", res.OverallScore)
	}
	if res.OverallLevel != CompatibilityLow {
		t.Errorf("OverallLevel = %q, want low", res.OverallLevel)
	}
	if res.DimensionScores.DayPillar != 0 {
		t.Errorf("DayPillar score = %d, want 0", res.DimensionScores.DayPillar)
	}

	// 负面证据已进入列表
	var hasChong, hasKe bool
	for _, e := range res.Evidences {
		if e.EvidenceKey == "neg_day_chong" {
			hasChong = true
		}
		if e.EvidenceKey == "neg_day_gan_ke" {
			hasKe = true
		}
	}
	if !hasChong || !hasKe {
		t.Errorf("expected neg_day_chong & neg_day_gan_ke in Evidences, got %+v", res.Evidences)
	}

	// score_explanation 的 day_pillar 条目填了负面因子
	var dayExp *CompatibilityScoreExplanation
	for i := range res.ScoreExplanations {
		if res.ScoreExplanations[i].Dimension == "day_pillar" {
			dayExp = &res.ScoreExplanations[i]
		}
	}
	if dayExp == nil {
		t.Fatal("missing day_pillar score explanation")
	}
	if dayExp.NegativeFactor == "" || len(dayExp.NegativeEvidenceKeys) == 0 {
		t.Errorf("day_pillar explanation missing negatives: %+v", dayExp)
	}
}
