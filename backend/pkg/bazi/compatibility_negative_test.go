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
