package bazi

import "testing"

func TestBranchCompatible_Liuhe(t *testing.T) {
	pairs := [][2]string{
		{"子", "丑"}, {"寅", "亥"}, {"卯", "戌"},
		{"辰", "酉"}, {"巳", "申"}, {"午", "未"},
	}
	for _, p := range pairs {
		if !branchCompatible(p[0], p[1]) {
			t.Errorf("liuhe %s/%s should be compatible", p[0], p[1])
		}
		if !branchCompatible(p[1], p[0]) {
			t.Errorf("liuhe %s/%s (reverse) should be compatible", p[1], p[0])
		}
	}
}

func TestBranchCompatible_Sanhe(t *testing.T) {
	for _, p := range [][2]string{{"申", "子"}, {"子", "辰"}, {"申", "辰"}} {
		if !branchCompatible(p[0], p[1]) {
			t.Errorf("sanhe %s/%s should be compatible", p[0], p[1])
		}
	}
}

func TestBranchCompatible_SameBranch_NotCompatible(t *testing.T) {
	for _, b := range []string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"} {
		if branchCompatible(b, b) {
			t.Errorf("same branch %s/%s should NOT be compatible (sanhe requires different)", b, b)
		}
	}
}

func TestBranchCompatible_ChongHaiXing_NotCompatible(t *testing.T) {
	cases := [][2]string{
		{"子", "午"}, {"丑", "未"}, {"寅", "申"},
		{"子", "未"}, {"丑", "午"},
		{"子", "卯"},
	}
	for _, p := range cases {
		if branchCompatible(p[0], p[1]) {
			t.Errorf("chong/hai/xing %s/%s should NOT be compatible", p[0], p[1])
		}
	}
}

func TestBranchCompatible_Empty_NotCompatible(t *testing.T) {
	if branchCompatible("", "子") || branchCompatible("子", "") {
		t.Error("empty branch should not be compatible")
	}
}

func TestBranchCompatibilityKind(t *testing.T) {
	cases := []struct {
		a, b string
		want string
	}{
		{"子", "丑", "liuhe"},
		{"申", "子", "sanhe"},
		{"子", "午", ""},
		{"子", "子", ""},
		{"", "子", ""},
	}
	for _, c := range cases {
		if got := branchCompatibilityKind(c.a, c.b); got != c.want {
			t.Errorf("branchCompatibilityKind(%q,%q) = %q, want %q", c.a, c.b, got, c.want)
		}
	}
}

func TestSanheGroupName(t *testing.T) {
	cases := []struct {
		a, b string
		want string
	}{
		{"申", "子", "申子辰"},
		{"亥", "卯", "亥卯未"},
		{"巳", "酉", "巳酉丑"},
		{"寅", "午", "寅午戌"},
		{"子", "丑", ""}, // liuhe，不是 sanhe
		{"子", "午", ""}, // chong
		{"子", "子", ""}, // 同支
	}
	for _, c := range cases {
		if got := sanheGroupName(c.a, c.b); got != c.want {
			t.Errorf("sanheGroupName(%q,%q) = %q, want %q", c.a, c.b, got, c.want)
		}
	}
}

func TestScoreZodiac_Liuhe_Returns50(t *testing.T) {
	if got := scoreZodiac("子", "丑"); got != 50 {
		t.Errorf("子丑 liuhe: got %d, want 50", got)
	}
}

func TestScoreZodiac_Sanhe_Returns50(t *testing.T) {
	if got := scoreZodiac("申", "子"); got != 50 {
		t.Errorf("申子 半三合: got %d, want 50", got)
	}
	if got := scoreZodiac("子", "辰"); got != 50 {
		t.Errorf("子辰 半三合: got %d, want 50", got)
	}
}

func TestScoreZodiac_NoHit_Returns0(t *testing.T) {
	// v3.1：相克/相冲（不同行）/相穿/相害/自刑 → 0
	// 注：子卯 虽然相刑，但水生木（相生）→ v3.1 按相生 20 计，已移至 _Sheng_Returns20 覆盖
	cases := [][2]string{
		{"子", "午"}, // 六冲（水克火）
		{"子", "未"}, // 六害（水土相克）
		{"子", "子"}, // 同支（自刑）
	}
	for _, p := range cases {
		if got := scoreZodiac(p[0], p[1]); got != 0 {
			t.Errorf("scoreZodiac(%q,%q) = %d, want 0", p[0], p[1], got)
		}
	}
}

func TestScoreZodiac_Empty_Returns0(t *testing.T) {
	if scoreZodiac("", "子") != 0 || scoreZodiac("子", "") != 0 {
		t.Error("empty branch should score 0")
	}
}

func TestScoreNayin_Sheng_Returns20(t *testing.T) {
	if got := scoreNayin("甲子", "庚午"); got != 20 {
		t.Errorf("甲子(金)+庚午(土) 相生: got %d, want 20", got)
	}
}

func TestScoreNayin_Same_Returns20(t *testing.T) {
	if got := scoreNayin("甲子", "乙丑"); got != 20 {
		t.Errorf("甲子(金)+乙丑(金) 同金: got %d, want 20", got)
	}
}

func TestScoreNayin_Ke_Returns0(t *testing.T) {
	if got := scoreNayin("甲子", "戊辰"); got != 0 {
		t.Errorf("甲子(金)+戊辰(木) 相克: got %d, want 0", got)
	}
}

func TestScoreNayin_Empty_Returns0(t *testing.T) {
	if scoreNayin("", "甲子") != 0 || scoreNayin("甲子", "") != 0 || scoreNayin("XX", "YY") != 0 {
		t.Error("empty / unknown ganzhi should score 0")
	}
}

func TestScoreDayPillar_UpperTier_GanHe(t *testing.T) {
	got := scoreDayPillar("甲", "子", "己", "丑")
	if got != 10 {
		t.Errorf("甲子/己丑 上档: got %d, want 10", got)
	}
}

func TestScoreDayPillar_UpperTier_GanSheng(t *testing.T) {
	got := scoreDayPillar("甲", "子", "丁", "丑")
	if got != 10 {
		t.Errorf("甲子/丁丑 上档: got %d, want 10", got)
	}
}

func TestScoreDayPillar_LowerTier_GanTong(t *testing.T) {
	got := scoreDayPillar("甲", "子", "乙", "丑")
	if got != 5 {
		t.Errorf("甲子/乙丑 下档(干同): got %d, want 5", got)
	}
}

func TestScoreDayPillar_LowerTier_GanKe(t *testing.T) {
	got := scoreDayPillar("甲", "子", "戊", "丑")
	if got != 5 {
		t.Errorf("甲子/戊丑 下档(干克): got %d, want 5", got)
	}
}

func TestScoreDayPillar_ZhiNotCompatible_Returns0(t *testing.T) {
	got := scoreDayPillar("甲", "子", "己", "未")
	if got != 0 {
		t.Errorf("甲子/己未 支不合: got %d, want 0", got)
	}
}

func TestScoreDayPillar_SanheZhi(t *testing.T) {
	got := scoreDayPillar("甲", "子", "己", "辰")
	if got != 10 {
		t.Errorf("甲子/己辰 上档(三合支): got %d, want 10", got)
	}
}

func TestScoreEightChars_AllUpper_Returns20(t *testing.T) {
	got := scoreEightChars(
		"甲", "子", "己", "丑",
		"甲", "子", "己", "丑",
		"甲", "子", "己", "丑",
	)
	if got != 20 {
		t.Errorf("三柱全上档: got %d, want 20", got)
	}
}

func TestScoreEightChars_AllLower_Returns10(t *testing.T) {
	got := scoreEightChars(
		"甲", "子", "乙", "丑",
		"甲", "子", "乙", "丑",
		"甲", "子", "乙", "丑",
	)
	if got != 10 {
		t.Errorf("三柱全下档: got %d, want 10", got)
	}
}

func TestScoreEightChars_OneUpperOnly_Returns7(t *testing.T) {
	got := scoreEightChars(
		"甲", "子", "己", "丑",
		"甲", "午", "甲", "午",
		"甲", "午", "甲", "午",
	)
	if got != 7 {
		t.Errorf("一柱上档 其余不合: got %d, want 7", got)
	}
}

func TestScoreEightChars_NothingHits_Returns0(t *testing.T) {
	got := scoreEightChars(
		"甲", "午", "甲", "午",
		"甲", "午", "甲", "午",
		"甲", "午", "甲", "午",
	)
	if got != 0 {
		t.Errorf("全不命中: got %d, want 0", got)
	}
}

func TestScoreEightChars_RoundingTable(t *testing.T) {
	cases := []struct {
		sum  int
		want int
	}{{0, 0}, {5, 3}, {10, 7}, {15, 10}, {20, 13}, {25, 17}, {30, 20}}
	for _, tc := range cases {
		if got := normalizeEightCharsSum(tc.sum); got != tc.want {
			t.Errorf("normalizeEightCharsSum(%d) = %d, want %d", tc.sum, got, tc.want)
		}
	}
}

func TestBranchSameElement_TrueCases(t *testing.T) {
	cases := [][2]string{
		{"亥", "子"}, {"子", "亥"}, // 水
		{"寅", "卯"}, {"卯", "寅"}, // 木
		{"巳", "午"}, {"午", "巳"}, // 火
		{"申", "酉"}, {"酉", "申"}, // 金
		{"辰", "戌"}, {"丑", "未"}, {"辰", "丑"}, {"戌", "未"}, // 土
	}
	for _, p := range cases {
		if !branchSameElement(p[0], p[1]) {
			t.Errorf("branchSameElement(%q,%q) = false, want true", p[0], p[1])
		}
	}
}

func TestBranchSameElement_FalseCases(t *testing.T) {
	cases := [][2]string{
		{"子", "子"},   // 同支
		{"子", "寅"},   // 水生木（不同行）
		{"子", "丑"},   // 不同行（水/土）
		{"", "子"}, {"子", ""},
	}
	for _, p := range cases {
		if branchSameElement(p[0], p[1]) {
			t.Errorf("branchSameElement(%q,%q) = true, want false", p[0], p[1])
		}
	}
}

func TestBranchShengElement_TrueCases(t *testing.T) {
	cases := [][2]string{
		{"子", "寅"}, {"寅", "子"}, // 水生木
		{"寅", "巳"}, {"巳", "寅"}, // 木生火
		{"巳", "辰"}, {"辰", "巳"}, // 火生土
		{"辰", "申"}, {"申", "辰"}, // 土生金
		{"申", "亥"}, {"亥", "申"}, // 金生水
	}
	for _, p := range cases {
		if !branchShengElement(p[0], p[1]) {
			t.Errorf("branchShengElement(%q,%q) = false, want true", p[0], p[1])
		}
	}
}

func TestBranchShengElement_FalseCases(t *testing.T) {
	cases := [][2]string{
		{"子", "子"},   // 同支
		{"亥", "子"},   // 同行水（不是相生）
		{"子", "未"},   // 水土相克
		{"子", "辰"},   // 水土相克
		{"子", "午"},   // 水火相克
		{"", "子"}, {"子", ""},
	}
	for _, p := range cases {
		if branchShengElement(p[0], p[1]) {
			t.Errorf("branchShengElement(%q,%q) = true, want false", p[0], p[1])
		}
	}
}

func TestScoreZodiac_SameElement_Returns30(t *testing.T) {
	// 五行同（双生）：上档(六合/三合)不命中 → 中档 30
	cases := [][2]string{
		{"亥", "子"}, // 水
		{"寅", "卯"}, // 木
		{"巳", "午"}, // 火
		{"申", "酉"}, // 金
		{"辰", "戌"}, // 土，同时也是六冲——按 v3.1 优先级判中档 30
		{"丑", "未"}, // 土，同时也是六冲——同上
	}
	for _, p := range cases {
		if got := scoreZodiac(p[0], p[1]); got != 30 {
			t.Errorf("scoreZodiac(%q,%q) = %d, want 30", p[0], p[1], got)
		}
	}
}

func TestScoreZodiac_Sheng_Returns20(t *testing.T) {
	// 五行相生：上档(六合/三合) + 中档(同行) 均不命中 → 下档 20
	// 注：辰申同属申子辰三合，用戌申（土生金、无三合/六合关系）替代
	// 子卯 虽相刑，但水生木，v3.1 纯加分制取正档 20
	cases := [][2]string{
		{"子", "寅"}, // 水生木
		{"子", "卯"}, // 水生木（相刑，但正面取优）
		{"寅", "巳"}, // 木生火
		{"巳", "辰"}, // 火生土
		{"戌", "申"}, // 土生金（辰申三合，故改用戌申）
		{"申", "亥"}, // 金生水
	}
	for _, p := range cases {
		if got := scoreZodiac(p[0], p[1]); got != 20 {
			t.Errorf("scoreZodiac(%q,%q) = %d, want 20", p[0], p[1], got)
		}
	}
}

func TestScoreZodiac_Priority_LiuheBeatsSheng(t *testing.T) {
	// 验证 cascade 优先级：同时构成 六合 与 五行相生 时，必须命中上档 50（不掉落到下档 20）。
	// （备注：六合 / 三合 与 同行(双生) 无法同时命中——所有 liuhe/sanhe 对的两支五行都不同行——所以
	// 无法用单一 input 测试 liuhe-vs-same-element 的优先级；priority guard 在 liuhe-vs-sheng 上才有实测价值。）
	cases := [][2]string{
		{"寅", "亥"}, // 六合 + 亥水生寅木 → 50
		{"午", "未"}, // 六合 + 午火生未土 → 50
	}
	for _, p := range cases {
		if got := scoreZodiac(p[0], p[1]); got != 50 {
			t.Errorf("scoreZodiac(%q,%q) = %d, want 50 (liuhe 上档优先于 sheng 下档)", p[0], p[1], got)
		}
	}
}

func TestScoreDayPillar_LowerTier3_SameElement(t *testing.T) {
	// 日支双生（亥子同水），干任意 → 3
	if got := scoreDayPillar("甲", "亥", "丙", "子"); got != 3 {
		t.Errorf("甲亥/丙子 双生日支: got %d, want 3", got)
	}
	// 干同（甲乙同木）也不影响下档分
	if got := scoreDayPillar("甲", "亥", "乙", "子"); got != 3 {
		t.Errorf("甲亥/乙子 双生日支(干同): got %d, want 3", got)
	}
}

func TestScoreDayPillar_LowerTier3_Sheng(t *testing.T) {
	// 日支五行相生（子→寅 水生木），干任意 → 3
	if got := scoreDayPillar("甲", "子", "丙", "寅"); got != 3 {
		t.Errorf("甲子/丙寅 水生木日支: got %d, want 3", got)
	}
}

func TestScoreDayPillar_Ke_Returns0(t *testing.T) {
	// 日支相克 / 相冲 / 相害 → 0
	if got := scoreDayPillar("甲", "子", "戊", "未"); got != 0 {
		t.Errorf("甲子/戊未 日支相害: got %d, want 0", got)
	}
	if got := scoreDayPillar("甲", "子", "戊", "戌"); got != 0 {
		t.Errorf("甲子/戊戌 日支(水土相克): got %d, want 0", got)
	}
}
