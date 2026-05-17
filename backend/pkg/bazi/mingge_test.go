package bazi

import "testing"

func TestLinGuanZhi(t *testing.T) {
	cases := map[string]string{
		"甲": "寅", "乙": "卯", "丙": "巳", "丁": "午",
		"戊": "巳", "己": "午", "庚": "申", "辛": "酉",
		"壬": "亥", "癸": "子",
	}
	for gan, want := range cases {
		if got := linGuanZhi[gan]; got != want {
			t.Errorf("linGuanZhi[%q] = %q, want %q", gan, got, want)
		}
	}
}

func TestDiWangZhi_OnlyYangGans(t *testing.T) {
	cases := map[string]string{
		"甲": "卯", "丙": "午", "戊": "午", "庚": "酉", "壬": "子",
	}
	for gan, want := range cases {
		if got := diWangZhi[gan]; got != want {
			t.Errorf("diWangZhi[%q] = %q, want %q", gan, got, want)
		}
	}
	for _, yin := range []string{"乙", "丁", "己", "辛", "癸"} {
		if _, ok := diWangZhi[yin]; ok {
			t.Errorf("diWangZhi should not contain 阴干 %q", yin)
		}
	}
}

func TestRootStrength_MonthBranchMainQi(t *testing.T) {
	// 月支寅, 主气甲 → 甲在月支主气, 根强度 = 6
	r := &BaziResult{
		YearGan: "辛", YearZhi: "酉",
		MonthGan: "丙", MonthZhi: "寅",
		DayGan: "庚", DayZhi: "子",
		HourGan: "辛", HourZhi: "酉",
	}
	if got := rootStrength("甲", r); got != 6 {
		t.Errorf("rootStrength(甲) = %d, want 6 (月支主气)", got)
	}
	if got := rootStrength("丙", r); got != 5 {
		t.Errorf("rootStrength(丙) = %d, want 5 (月支中气)", got)
	}
	if got := rootStrength("戊", r); got != 4 {
		t.Errorf("rootStrength(戊) = %d, want 4 (月支余气)", got)
	}
	if got := rootStrength("辛", r); got != 3 {
		t.Errorf("rootStrength(辛) = %d, want 3 (他支主气)", got)
	}
	if got := rootStrength("癸", r); got != 3 {
		t.Errorf("rootStrength(癸) = %d, want 3", got)
	}
	if got := rootStrength("庚", r); got != 0 {
		t.Errorf("rootStrength(庚) = %d, want 0 (无根)", got)
	}
}

func TestRootStrength_OtherBranchMidAndRemainder(t *testing.T) {
	// 年支辰 (主气戊, 中气乙, 余气癸) → 乙=他支中气=2, 癸=他支余气=1
	// 时支午 (主气丁, 中气己) → 己=他支中气=2
	r := &BaziResult{
		YearGan: "戊", YearZhi: "辰",
		MonthGan: "丙", MonthZhi: "寅",
		DayGan: "甲", DayZhi: "子",
		HourGan: "丁", HourZhi: "午",
	}
	// 乙 在 辰中气 (他支中气=2). 寅其它主中余? 寅=[甲,丙,戊], 没乙. 子=[癸], 没. 午=[丁,己], 没.
	if got := rootStrength("乙", r); got != 2 {
		t.Errorf("rootStrength(乙) = %d, want 2 (他支中气)", got)
	}
	// 癸 在 子 主气 (日支主气=3) 和 辰 余气 (他支余气=1). 取最强=3.
	if got := rootStrength("癸", r); got != 3 {
		t.Errorf("rootStrength(癸) = %d, want 3 (取日支主气最强)", got)
	}
	// 己 在 午中气 (时支=他支中气=2). 寅/子/辰都不藏己.
	if got := rootStrength("己", r); got != 2 {
		t.Errorf("rootStrength(己) = %d, want 2 (午中气=他支中气)", got)
	}
}

func TestRootStrength_MonthBranchWinsOverOtherBranch(t *testing.T) {
	// 月支寅 (主气甲), 年支寅 (主气甲) → 甲 同时在月支和他支扎主气根, 应取月支根=6
	r := &BaziResult{
		YearGan: "辛", YearZhi: "寅",
		MonthGan: "甲", MonthZhi: "寅",
		DayGan: "丙", DayZhi: "子",
		HourGan: "辛", HourZhi: "酉",
	}
	if got := rootStrength("甲", r); got != 6 {
		t.Errorf("rootStrength(甲) = %d, want 6 (月支主气 wins over 他支主气)", got)
	}
}
