package bazi

import (
	"slices"
	"testing"
)

// GetPillarsShenSha 参数顺序：yg, yz, mg, mz, dg, dz, hg, hz
// 返回 [4][]string，索引 0=年 1=月 2=日 3=时

func TestPillars_SangMen(t *testing.T) {
	// 年支子 → 丧门=寅；月支寅 → 月柱(索引1)命中
	r := GetPillarsShenSha("甲", "子", "丙", "寅", "戊", "辰", "庚", "午")
	if !slices.Contains(r[1], "丧门") {
		t.Fatalf("月柱应含丧门，实际 %v", r[1])
	}
}

func TestPillars_PiMa(t *testing.T) {
	// 年支子 → 披麻=酉；时支酉 → 时柱(索引3)命中
	r := GetPillarsShenSha("甲", "子", "丁", "卯", "己", "巳", "癸", "酉")
	if !slices.Contains(r[3], "披麻") {
		t.Fatalf("时柱应含披麻，实际 %v", r[3])
	}
}

func TestPillars_SanQiu_SpringHit(t *testing.T) {
	// 月支寅(春) → 三丘=辛丑；年柱=辛丑(非月柱) → 索引0命中
	r := GetPillarsShenSha("辛", "丑", "庚", "寅", "戊", "戌", "甲", "午")
	if !slices.Contains(r[0], "三丘") {
		t.Fatalf("年柱应含三丘，实际 %v", r[0])
	}
}

func TestPillars_SanQiu_StrictGanZhi(t *testing.T) {
	// 月支寅(春) → 三丘=辛丑；年柱=己丑(地支对、天干错) → 不命中（B 方案严格性）
	r := GetPillarsShenSha("己", "丑", "庚", "寅", "戊", "戌", "甲", "午")
	if slices.Contains(r[0], "三丘") {
		t.Fatalf("己丑≠辛丑，年柱不应含三丘，实际 %v", r[0])
	}
}

func TestPillars_WuMu_SeasonMonthHitNonMonthPillar(t *testing.T) {
	// 月支辰(季月) → 五墓=戊辰；日柱=戊辰(非月柱) → 索引2命中
	r := GetPillarsShenSha("甲", "子", "庚", "辰", "戊", "辰", "丙", "午")
	if !slices.Contains(r[2], "五墓") {
		t.Fatalf("日柱应含五墓，实际 %v", r[2])
	}
}

func TestPillars_WuMu_ExcludeMonthSelfMatch(t *testing.T) {
	// 月支辰(季月) → 五墓=戊辰；月柱本身=戊辰，但月柱为基准柱，应排除
	r := GetPillarsShenSha("甲", "子", "戊", "辰", "乙", "酉", "丙", "午")
	if slices.Contains(r[1], "五墓") {
		t.Fatalf("月柱自命中应被排除，实际 %v", r[1])
	}
}

func TestPillars_WuMu_OrdinaryMonth(t *testing.T) {
	// 月支寅(正月) → 五墓=乙未；日柱=乙未 → 索引2命中
	r := GetPillarsShenSha("甲", "子", "丙", "寅", "乙", "未", "庚", "午")
	if !slices.Contains(r[2], "五墓") {
		t.Fatalf("日柱应含五墓，实际 %v", r[2])
	}
}

// GetDayunShenSha 参数顺序：yearGan, yearZhi, monthZhi, dayGan, dayZhi, dayunGan, dayunZhi

func TestDayun_SangMen(t *testing.T) {
	// 年支子 → 丧门=寅；大运支=寅 → 命中
	r := GetDayunShenSha("甲", "子", "寅", "戊", "辰", "丙", "寅")
	if !slices.Contains(r, "丧门") {
		t.Fatalf("大运应含丧门，实际 %v", r)
	}
}

func TestDayun_PiMa(t *testing.T) {
	// 年支子 → 披麻=酉；大运支=酉 → 命中
	r := GetDayunShenSha("甲", "子", "寅", "戊", "辰", "乙", "酉")
	if !slices.Contains(r, "披麻") {
		t.Fatalf("大运应含披麻，实际 %v", r)
	}
}

func TestDayun_SanQiu_Hit(t *testing.T) {
	// 月支寅(春) → 三丘=辛丑；大运=辛丑 → 命中
	r := GetDayunShenSha("甲", "子", "寅", "戊", "辰", "辛", "丑")
	if !slices.Contains(r, "三丘") {
		t.Fatalf("大运应含三丘，实际 %v", r)
	}
}

func TestDayun_SanQiu_StrictGanZhi(t *testing.T) {
	// 月支寅(春) → 三丘=辛丑；大运=己丑(地支对、天干错) → 不命中
	r := GetDayunShenSha("甲", "子", "寅", "戊", "辰", "己", "丑")
	if slices.Contains(r, "三丘") {
		t.Fatalf("己丑≠辛丑，大运不应含三丘，实际 %v", r)
	}
}

func TestDayun_WuMu_SeasonMonth(t *testing.T) {
	// 月支辰(季月) → 五墓=戊辰；大运=戊辰 → 命中（大运无月柱排除）
	r := GetDayunShenSha("甲", "子", "辰", "戊", "酉", "戊", "辰")
	if !slices.Contains(r, "五墓") {
		t.Fatalf("大运应含五墓，实际 %v", r)
	}
}
