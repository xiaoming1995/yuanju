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

func TestDetectMingGe_ProfessionalRules(t *testing.T) {
	tests := []struct {
		name   string
		result *BaziResult
		wantGe string
	}{
		// 规则 3：月支中气根强，单候选立格
		{
			name: "C1 男 1995-10-12 11时 → 偏印格",
			result: &BaziResult{
				YearGan: "乙", YearZhi: "亥",
				MonthGan: "丙", MonthZhi: "戌",
				DayGan: "丙", DayZhi: "子",
				HourGan: "甲", HourZhi: "午",
			},
			wantGe: "偏印格",
		},
		{
			name: "C2 男 1996-02-08 20时 → 伤官格",
			result: &BaziResult{
				YearGan: "丙", YearZhi: "子",
				MonthGan: "庚", MonthZhi: "寅",
				DayGan: "乙", DayZhi: "亥",
				HourGan: "丙", HourZhi: "戌",
			},
			wantGe: "伤官格",
		},
		// 规则 4：食伤同透取伤（C3 辛无根但强制立）
		{
			name: "C3 女 1991-02-07 16时 → 伤官格（食伤同透）",
			result: &BaziResult{
				YearGan: "辛", YearZhi: "未",
				MonthGan: "庚", MonthZhi: "寅",
				DayGan: "戊", DayZhi: "申",
				HourGan: "庚", HourZhi: "申",
			},
			wantGe: "伤官格",
		},
		// 规则 3：单候选 + 余气根
		{
			name: "C4 女 1997-12-01 12时 → 偏财格",
			result: &BaziResult{
				YearGan: "丁", YearZhi: "丑",
				MonthGan: "辛", MonthZhi: "亥",
				DayGan: "丁", DayZhi: "丑",
				HourGan: "丙", HourZhi: "午",
			},
			wantGe: "偏财格",
		},
		// 规则 5：单候选无根 → 杂气格
		{
			name: "C5 女 1988-01-18 04时 → 杂气格",
			result: &BaziResult{
				YearGan: "丁", YearZhi: "卯",
				MonthGan: "癸", MonthZhi: "丑",
				DayGan: "壬", DayZhi: "申",
				HourGan: "壬", HourZhi: "寅",
			},
			wantGe: "杂气格",
		},
		{
			name: "C6 女 1996-11-08 08时 → 杂气格",
			result: &BaziResult{
				YearGan: "丙", YearZhi: "子",
				MonthGan: "己", MonthZhi: "亥",
				DayGan: "己", DayZhi: "酉",
				HourGan: "戊", HourZhi: "辰",
			},
			wantGe: "杂气格",
		},
		// 规则 3：多候选按根强度选
		{
			name: "C7 女 1991-12-30 10时 → 正财格",
			result: &BaziResult{
				YearGan: "辛", YearZhi: "未",
				MonthGan: "庚", MonthZhi: "子",
				DayGan: "甲", DayZhi: "戌",
				HourGan: "己", HourZhi: "巳",
			},
			wantGe: "正财格",
		},
		// 规则 5：候选都无根 + 地支气势非财 → 杂气格
		{
			name: "C8 男 1996-12-16 22时 → 杂气格（地支全水 但非财）",
			result: &BaziResult{
				YearGan: "丙", YearZhi: "子",
				MonthGan: "庚", MonthZhi: "子",
				DayGan: "丁", DayZhi: "亥",
				HourGan: "辛", HourZhi: "亥",
			},
			wantGe: "杂气格",
		},
		// 规则 3：他支中气根 > 他支余气根
		{
			name: "C9 男 1995-01-23 16时 → 偏印格",
			result: &BaziResult{
				YearGan: "甲", YearZhi: "戌",
				MonthGan: "丁", MonthZhi: "丑",
				DayGan: "甲", DayZhi: "寅",
				HourGan: "壬", HourZhi: "申",
			},
			wantGe: "偏印格",
		},
		// 规则 3：月支中气根 > 他支中气根
		{
			name: "C10 女 1993-01-16 14时 → 七杀格",
			result: &BaziResult{
				YearGan: "壬", YearZhi: "申",
				MonthGan: "癸", MonthZhi: "丑",
				DayGan: "丁", DayZhi: "酉",
				HourGan: "丁", HourZhi: "未",
			},
			wantGe: "七杀格",
		},
		// 规则 3：月支中气根 > 他支主气根 > 无根
		{
			name: "C11 男 2015-02-02 18时40分 → 偏财格",
			result: &BaziResult{
				YearGan: "甲", YearZhi: "午",
				MonthGan: "丁", MonthZhi: "丑",
				DayGan: "己", DayZhi: "酉",
				HourGan: "癸", HourZhi: "酉",
			},
			wantGe: "偏财格",
		},
		// 规则 6：候选都无根 + 地支气势全土 = 乙日财 → 月支主气戊配乙阴 异性 → 正财格
		{
			name: "C12 男 1985-04-26 19时46分 → 正财格（地支气势全土）",
			result: &BaziResult{
				YearGan: "乙", YearZhi: "丑",
				MonthGan: "庚", MonthZhi: "辰",
				DayGan: "乙", DayZhi: "未",
				HourGan: "丙", HourZhi: "戌",
			},
			wantGe: "正财格",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotGe, _ := DetectMingGe(tt.result)
			if gotGe != tt.wantGe {
				t.Errorf("DetectMingGe got = %q, want %q", gotGe, tt.wantGe)
			}
		})
	}
}
