package bazi

import (
	"reflect"
	"testing"
)

func TestBuildGongJiaAdjacentSameGanGap(t *testing.T) {
	result := &BaziResult{
		YearGan:  "甲",
		YearZhi:  "子",
		MonthGan: "甲",
		MonthZhi: "寅",
		DayGan:   "丙",
		DayZhi:   "午",
		HourGan:  "戊",
		HourZhi:  "申",
	}

	items := BuildGongJia(result)

	if len(items) != 1 {
		t.Fatalf("BuildGongJia() len = %d, want 1: %#v", len(items), items)
	}
	got := items[0]
	if got.Source != "year_month" {
		t.Errorf("Source = %q, want year_month", got.Source)
	}
	if got.SameGan != "甲" {
		t.Errorf("SameGan = %q, want 甲", got.SameGan)
	}
	if got.VirtualZhi != "丑" {
		t.Errorf("VirtualZhi = %q, want 丑", got.VirtualZhi)
	}
	if !reflect.DeepEqual(got.HideGan, []string{"己", "癸", "辛"}) {
		t.Errorf("HideGan = %#v, want [己 癸 辛]", got.HideGan)
	}
}

func TestBuildGongJiaReverseBranchOrder(t *testing.T) {
	result := &BaziResult{
		YearGan:  "甲",
		YearZhi:  "寅",
		MonthGan: "甲",
		MonthZhi: "子",
		DayGan:   "丙",
		DayZhi:   "午",
		HourGan:  "戊",
		HourZhi:  "申",
	}

	items := BuildGongJia(result)

	if len(items) != 1 {
		t.Fatalf("BuildGongJia() len = %d, want 1: %#v", len(items), items)
	}
	if items[0].VirtualZhi != "丑" {
		t.Errorf("VirtualZhi = %q, want 丑", items[0].VirtualZhi)
	}
}

func TestBuildGongJiaAddsBranchBasedShenSha(t *testing.T) {
	result := &BaziResult{
		YearGan:  "甲",
		YearZhi:  "子",
		MonthGan: "甲",
		MonthZhi: "寅",
		DayGan:   "庚",
		DayZhi:   "午",
		HourGan:  "戊",
		HourZhi:  "申",
	}

	items := BuildGongJia(result)

	if len(items) != 1 {
		t.Fatalf("BuildGongJia() len = %d, want 1: %#v", len(items), items)
	}
	if items[0].VirtualZhi != "丑" {
		t.Fatalf("VirtualZhi = %q, want 丑", items[0].VirtualZhi)
	}
	if !containsGongJiaString(items[0].ShenSha, "天乙贵人") {
		t.Fatalf("ShenSha = %#v, want 天乙贵人", items[0].ShenSha)
	}
}

func TestVirtualBranchShenshaDoesNotCreateFullPillarGanZhi(t *testing.T) {
	result := &BaziResult{
		YearGan:  "甲",
		YearZhi:  "子",
		MonthGan: "甲",
		MonthZhi: "寅",
		DayGan:   "庚",
		DayZhi:   "午",
		HourGan:  "戊",
		HourZhi:  "申",
	}

	items := BuildGongJia(result)

	if len(items) != 1 {
		t.Fatalf("BuildGongJia() len = %d, want 1: %#v", len(items), items)
	}
	for _, name := range []string{"阴差阳错", "魁罡"} {
		if containsGongJiaString(items[0].ShenSha, name) {
			t.Fatalf("ShenSha = %#v, should not include full-pillar shensha %s", items[0].ShenSha, name)
		}
	}
}

func TestGetGongJiaShenShaIgnoresEmptyInputs(t *testing.T) {
	tests := []struct {
		name       string
		yearGan    string
		yearZhi    string
		monthZhi   string
		dayGan     string
		dayZhi     string
		virtualZhi string
	}{
		{
			name:       "empty bases with tianyi branch",
			virtualZhi: "丑",
		},
		{
			name:       "empty virtual branch with valid year stem",
			yearGan:    "甲",
			virtualZhi: "",
		},
		{
			name:       "empty bases with wenchang branch",
			virtualZhi: "申",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetGongJiaShenSha(tt.yearGan, tt.yearZhi, tt.monthZhi, tt.dayGan, tt.dayZhi, tt.virtualZhi)
			if len(got) != 0 {
				t.Fatalf("GetGongJiaShenSha() = %#v, want empty", got)
			}
		})
	}
}

func TestGetGongJiaShenShaRequiresNonEmptyBranchBases(t *testing.T) {
	for _, virtualZhi := range []string{"酉", "午", "辰", "子", "巳", "卯"} {
		t.Run("empty bases "+virtualZhi, func(t *testing.T) {
			got := GetGongJiaShenSha("", "", "", "", "", virtualZhi)
			if len(got) != 0 {
				t.Fatalf("GetGongJiaShenSha() = %#v, want empty", got)
			}
		})
	}

	got := GetGongJiaShenSha("", "子", "", "", "", "酉")
	if !containsGongJiaString(got, "桃花") {
		t.Fatalf("GetGongJiaShenSha() = %#v, want 桃花 from valid non-empty yearZhi base", got)
	}
}

func TestBuildGongJiaAdjacentOnly(t *testing.T) {
	result := &BaziResult{
		YearGan:  "甲",
		YearZhi:  "子",
		MonthGan: "乙",
		MonthZhi: "卯",
		DayGan:   "甲",
		DayZhi:   "寅",
		HourGan:  "戊",
		HourZhi:  "申",
	}

	items := BuildGongJia(result)

	if len(items) != 0 {
		t.Fatalf("BuildGongJia() len = %d, want 0: %#v", len(items), items)
	}
}

func TestBuildGongJiaRejectsDifferentGanOrNonGapBranches(t *testing.T) {
	tests := []struct {
		name   string
		result *BaziResult
	}{
		{
			name: "different stems",
			result: &BaziResult{
				YearGan:  "甲",
				YearZhi:  "子",
				MonthGan: "乙",
				MonthZhi: "寅",
				DayGan:   "丙",
				DayZhi:   "午",
				HourGan:  "戊",
				HourZhi:  "申",
			},
		},
		{
			name: "non gap branches",
			result: &BaziResult{
				YearGan:  "甲",
				YearZhi:  "子",
				MonthGan: "甲",
				MonthZhi: "卯",
				DayGan:   "丙",
				DayZhi:   "午",
				HourGan:  "戊",
				HourZhi:  "申",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items := BuildGongJia(tt.result)
			if len(items) != 0 {
				t.Fatalf("BuildGongJia() len = %d, want 0: %#v", len(items), items)
			}
		})
	}
}

func TestEnsureGongJiaBackfillsExplicitEmptySlice(t *testing.T) {
	result := &BaziResult{
		YearGan:  "甲",
		YearZhi:  "子",
		MonthGan: "甲",
		MonthZhi: "寅",
		DayGan:   "丙",
		DayZhi:   "午",
		HourGan:  "戊",
		HourZhi:  "申",
		GongJia:  []GongJiaItem{},
	}

	changed := EnsureGongJia(result)

	if !changed {
		t.Fatal("EnsureGongJia() changed = false, want true")
	}
	if len(result.GongJia) != 1 {
		t.Fatalf("GongJia len = %d, want 1: %#v", len(result.GongJia), result.GongJia)
	}
	if result.GongJia[0].VirtualZhi != "丑" {
		t.Errorf("VirtualZhi = %q, want 丑", result.GongJia[0].VirtualZhi)
	}
}

func TestEnsureGongJiaPreservesNonEmptyExistingSlice(t *testing.T) {
	existing := []GongJiaItem{{
		Source:     "manual",
		VirtualZhi: "辰",
	}}
	result := &BaziResult{
		YearGan:  "甲",
		YearZhi:  "子",
		MonthGan: "甲",
		MonthZhi: "寅",
		DayGan:   "丙",
		DayZhi:   "午",
		HourGan:  "戊",
		HourZhi:  "申",
		GongJia:  existing,
	}

	changed := EnsureGongJia(result)

	if changed {
		t.Fatal("EnsureGongJia() changed = true, want false")
	}
	if !reflect.DeepEqual(result.GongJia, existing) {
		t.Fatalf("GongJia = %#v, want preserved %#v", result.GongJia, existing)
	}
}

func TestCalculateBuildsGongJia(t *testing.T) {
	result := Calculate(1980, 1, 1, 21, "male", false, 120, "solar", false)

	if len(result.GongJia) == 0 {
		t.Fatalf("Calculate() GongJia len = 0, want non-empty; pillars=%s%s %s%s %s%s %s%s",
			result.YearGan, result.YearZhi,
			result.MonthGan, result.MonthZhi,
			result.DayGan, result.DayZhi,
			result.HourGan, result.HourZhi,
		)
	}
	got := result.GongJia[0]
	if got.Source != "day_hour" {
		t.Errorf("Source = %q, want day_hour", got.Source)
	}
	if got.SameGan != "癸" {
		t.Errorf("SameGan = %q, want 癸", got.SameGan)
	}
	if got.VirtualZhi != "戌" {
		t.Errorf("VirtualZhi = %q, want 戌", got.VirtualZhi)
	}
}

func TestEnsureGongJiaDoesNotMutateCoreFields(t *testing.T) {
	result := &BaziResult{
		YearGan:    "甲",
		YearZhi:    "子",
		MonthGan:   "甲",
		MonthZhi:   "寅",
		DayGan:     "丙",
		DayZhi:     "午",
		HourGan:    "戊",
		HourZhi:    "申",
		Wuxing:     WuxingStats{Mu: 1, Huo: 2, Tu: 3, Jin: 1, Shui: 1, Total: 8, MuPct: 12.5, HuoPct: 25, TuPct: 37.5, JinPct: 12.5, ShuiPct: 12.5},
		Yongshen:   "木火",
		Jishen:     "金水",
		Tiaohou:    &TiaohouResult{Expected: []string{"壬", "庚"}, Tou: []string{"壬"}, Cang: []string{"庚"}, Text: "fixture"},
		MingGe:     "食神格",
		MingGeDesc: "fixture desc",
		GongJia:    []GongJiaItem{},
	}
	wantWuxing := result.Wuxing
	wantYongshen := result.Yongshen
	wantJishen := result.Jishen
	wantTiaohou := TiaohouResult{
		Expected: append([]string(nil), result.Tiaohou.Expected...),
		Tou:      append([]string(nil), result.Tiaohou.Tou...),
		Cang:     append([]string(nil), result.Tiaohou.Cang...),
		Text:     result.Tiaohou.Text,
	}
	wantMingGe := result.MingGe
	wantMingGeDesc := result.MingGeDesc

	EnsureGongJia(result)

	if result.Wuxing != wantWuxing {
		t.Errorf("Wuxing mutated: got %#v, want %#v", result.Wuxing, wantWuxing)
	}
	if result.Yongshen != wantYongshen {
		t.Errorf("Yongshen mutated: got %q, want %q", result.Yongshen, wantYongshen)
	}
	if result.Jishen != wantJishen {
		t.Errorf("Jishen mutated: got %q, want %q", result.Jishen, wantJishen)
	}
	if !reflect.DeepEqual(*result.Tiaohou, wantTiaohou) {
		t.Errorf("Tiaohou mutated: got %#v, want %#v", *result.Tiaohou, wantTiaohou)
	}
	if result.MingGe != wantMingGe {
		t.Errorf("MingGe mutated: got %q, want %q", result.MingGe, wantMingGe)
	}
	if result.MingGeDesc != wantMingGeDesc {
		t.Errorf("MingGeDesc mutated: got %q, want %q", result.MingGeDesc, wantMingGeDesc)
	}
}

func containsGongJiaString(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
