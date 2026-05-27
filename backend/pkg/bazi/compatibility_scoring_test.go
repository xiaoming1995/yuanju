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
	cases := [][2]string{
		{"子", "午"}, // 六冲
		{"子", "未"}, // 六害
		{"子", "卯"}, // 相刑
		{"子", "子"}, // 同支（自刑）
		{"寅", "卯"}, // 双生（五行同），不命中
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
