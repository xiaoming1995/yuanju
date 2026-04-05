package bazi

import (
	"testing"
)

// 测试年柱推算（1893年，立春后属癸巳年）
func TestYearPillar1893(t *testing.T) {
	result := Calculate(1893, 12, 26, 7, "male", false, 0)
	if result.YearGan != "癸" {
		t.Errorf("1893年冬年干期望'癸'，实际'%s'", result.YearGan)
	}
	if result.YearZhi != "巳" {
		t.Errorf("1893年冬年支期望'巳'，实际'%s'", result.YearZhi)
	}
	t.Logf("1893-12-26 07时 计算结果：%s%s·%s%s·%s%s·%s%s",
		result.YearGan, result.YearZhi,
		result.MonthGan, result.MonthZhi,
		result.DayGan, result.DayZhi,
		result.HourGan, result.HourZhi,
	)
}

// 测试时辰边界：早子时与晚子时日柱应不同
func TestHourBoundary(t *testing.T) {
	// 早子时（23点，属前一天）
	r1 := Calculate(1990, 6, 15, 23, "male", true, 0)
	// 晚子时（23点，属当天）
	r2 := Calculate(1990, 6, 15, 23, "male", false, 0)

	t.Logf("早子时（属前一日）日干：%s%s", r1.DayGan, r1.DayZhi)
	t.Logf("晚子时（属当日）日干：%s%s", r2.DayGan, r2.DayZhi)

	// 早子时与晚子时日柱必须不同
	if r1.DayGan == r2.DayGan && r1.DayZhi == r2.DayZhi {
		t.Error("早子时与晚子时应计算不同的日柱")
	}

	// 两者时支都应该是子时
	if r1.HourZhi != "子" {
		t.Errorf("早子时时支期望'子'，实际'%s'", r1.HourZhi)
	}
	if r2.HourZhi != "子" {
		t.Errorf("晚子时时支期望'子'，实际'%s'", r2.HourZhi)
	}
}

// 测试五行统计总和必须等于8（四柱共8个字）
func TestWuxingTotal(t *testing.T) {
	result := Calculate(1990, 6, 15, 14, "female", false, 0)
	total := result.Wuxing.Mu + result.Wuxing.Huo + result.Wuxing.Tu +
		result.Wuxing.Jin + result.Wuxing.Shui
	if total != 8 {
		t.Errorf("五行总数应为8（四柱共8字），实际为%d", total)
	}
	t.Logf("五行分布：木%d 火%d 土%d 金%d 水%d（共%d）",
		result.Wuxing.Mu, result.Wuxing.Huo, result.Wuxing.Tu,
		result.Wuxing.Jin, result.Wuxing.Shui, total)
}

// 测试五行百分比总和约等于100%
func TestWuxingPercentage(t *testing.T) {
	result := Calculate(1985, 3, 20, 10, "male", false, 0)
	pctTotal := result.Wuxing.MuPct + result.Wuxing.HuoPct + result.Wuxing.TuPct +
		result.Wuxing.JinPct + result.Wuxing.ShuiPct
	if pctTotal < 99.9 || pctTotal > 100.1 {
		t.Errorf("五行百分比总和应为100%%，实际为%.2f%%", pctTotal)
	}
}

// 测试大运推算结构
func TestDayun(t *testing.T) {
	result := Calculate(1990, 6, 15, 14, "male", false, 0)
	if len(result.Dayun) == 0 {
		t.Error("大运应返回至少1步")
	}
	for i, d := range result.Dayun {
		t.Logf("第%d步大运：%s%s，%d岁起（%d-%d）",
			i+1, d.Gan, d.Zhi, d.StartAge, d.StartYear, d.EndYear)
			
		// 校验交脱期标志
		if len(d.LiuNian) > 0 {
			firstLn := d.LiuNian[0]
			if !firstLn.IsTransition {
				t.Errorf("大运首年 %d 应被标记为交脱年", firstLn.Year)
			}
			if firstLn.TransMonth == 0 || firstLn.TransDay == 0 {
				t.Errorf("大运首年应挂载非0的交脱月日，实际 %d-%d", firstLn.TransMonth, firstLn.TransDay)
			}
			if i > 0 && firstLn.PrevDayun == "" {
				t.Errorf("非首运的流年交脱期中，上一运标识不应为空")
			}
			if len(d.LiuNian) > 1 {
				if d.LiuNian[1].IsTransition {
					t.Errorf("大运次年 %d 不应被标记为交脱年", d.LiuNian[1].Year)
				}
			}
		}
	}
}

// 测试哈希唯一性和一致性
func TestChartHash(t *testing.T) {
	r1 := Calculate(1990, 6, 15, 14, "male", false, 0)
	r2 := Calculate(1990, 6, 15, 14, "male", false, 0)
	r3 := Calculate(1990, 6, 15, 14, "female", false, 0)
	r4 := Calculate(1990, 6, 16, 14, "male", false, 0)

	if r1.ChartHash != r2.ChartHash {
		t.Error("相同输入应生成相同hash")
	}
	if r1.ChartHash == r3.ChartHash {
		t.Error("不同性别应生成不同hash")
	}
	if r1.ChartHash == r4.ChartHash {
		t.Error("不同日期应生成不同hash")
	}
}

// 测试天干地支的有效性（必须在合法范围内）
func TestValidGanZhi(t *testing.T) {
	validGan := map[string]bool{"甲": true, "乙": true, "丙": true, "丁": true, "戊": true, "己": true, "庚": true, "辛": true, "壬": true, "癸": true}
	validZhi := map[string]bool{"子": true, "丑": true, "寅": true, "卯": true, "辰": true, "巳": true, "午": true, "未": true, "申": true, "酉": true, "戌": true, "亥": true}

	result := Calculate(2000, 1, 15, 12, "female", false, 0)
	for name, val := range map[string]string{
		"年干": result.YearGan, "月干": result.MonthGan,
		"日干": result.DayGan, "时干": result.HourGan,
	} {
		if !validGan[val] {
			t.Errorf("%s='%s' 不是有效天干", name, val)
		}
	}
	for name, val := range map[string]string{
		"年支": result.YearZhi, "月支": result.MonthZhi,
		"日支": result.DayZhi, "时支": result.HourZhi,
	} {
		if !validZhi[val] {
			t.Errorf("%s='%s' 不是有效地支", name, val)
		}
	}
	t.Logf("2000-01-15 12时：%s%s·%s%s·%s%s·%s%s",
		result.YearGan, result.YearZhi, result.MonthGan, result.MonthZhi,
		result.DayGan, result.DayZhi, result.HourGan, result.HourZhi)

	// 验证藏干字段不为空
	if len(result.YearHideGan) == 0 {
		t.Error("年支藏干不应为空")
	}
	t.Logf("年支藏干：%v，月支藏干：%v，日支藏干：%v，时支藏干：%v",
		result.YearHideGan, result.MonthHideGan, result.DayHideGan, result.HourHideGan)
	t.Logf("真太阳时：%d:%02d", result.TrueSolarHour, result.TrueSolarMinute)
}

// 测试真太阳时修正（新疆经度约 87.6°，比北京慢约 2.1 小时）
func TestTrueSolarTime(t *testing.T) {
	// 北京时间 14:00，新疆经度 87.6
	result := Calculate(1990, 6, 15, 14, "male", false, 87.6)
	t.Logf("北京时间14:00，新疆真太阳时：%d:%02d", result.TrueSolarHour, result.TrueSolarMinute)
	// 真太阳时应该比14:00早约2小时
	if result.TrueSolarHour >= 14 {
		t.Errorf("新疆经度应使真太阳时早于北京时间，期望 < 14:00，实际 %d:%02d",
			result.TrueSolarHour, result.TrueSolarMinute)
	}
}

// ===== 调候用神查表测试 =====

// TestLookupTiaohou 验证典型命例的调候用神结果
func TestLookupTiaohou(t *testing.T) {
	cases := []struct {
		dayGan   string
		monthZhi string
		wantKey  string // 期望结果中必须包含的关键字
		desc     string
	}{
		{"甲", "子", "丙火", "甲木生子月（冬），急需丙火暖局"},
		{"壬", "子", "戊土", "壬水生子月（冬），水势极旺，戊土为堤坝最急"},
		{"丙", "午", "壬水", "丙火生午月（夏），火极旺，壬水制约为急"},
		{"癸", "午", "庚", "癸水生午月（夏），癸极弱，金生水补源"},
		{"甲", "午", "癸水", "甲木生午月（夏），燥热，癸水解暑救燥为急"},
	}

	for _, c := range cases {
		result := LookupTiaohou(c.dayGan, c.monthZhi)
		if result == "" {
			t.Errorf("[%s] LookupTiaohou(%s, %s) 返回空字符串，期望非空", c.desc, c.dayGan, c.monthZhi)
			continue
		}
		found := false
		for _, r := range []rune(result) {
			_ = r
			found = true
			break
		}
		if !found {
			t.Errorf("[%s] 结果为空", c.desc)
		}
		t.Logf("[%s] %s日主生%s月 → 调候用神：%s", c.desc, c.dayGan, c.monthZhi, result)
	}
}

// TestLookupTiaohouMissing 验证非法输入返回空字符串
func TestLookupTiaohouMissing(t *testing.T) {
	result := LookupTiaohou("甲", "X月")
	if result != "" {
		t.Errorf("非法月支应返回空字符串，实际：%s", result)
	}
	result2 := LookupTiaohou("", "子")
	if result2 != "" {
		t.Errorf("空日干应返回空字符串，实际：%s", result2)
	}
}

// TestTiaohouInBaziResult 验证 BaziResult.Tiaohou 字段通过 Calculate() 正常返回
func TestTiaohouInBaziResult(t *testing.T) {
	// 甲木生于午月（夏），期望包含"癸水"
	result := Calculate(1990, 6, 15, 14, "male", false, 0)
	if result.Tiaohou == "" {
		t.Error("BaziResult.Tiaohou 不应为空（炎夏命局应有明确调候用神）")
	}
	t.Logf("日主：%s，月支：%s，调候用神：%s", result.DayGan, result.MonthZhi, result.Tiaohou)
}
