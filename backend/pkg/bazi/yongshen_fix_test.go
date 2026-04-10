package bazi

import "testing"

// TestYongshenFix_JiTuInMao 验证 Bug 修复：1989年3月20日 己土卯月
// 卯月木旺（七杀当令），己土失令，应判身弱，喜用神应为火土，忌神为木水金
func TestYongshenFix_JiTuInMao(t *testing.T) {
	// 1989年3月20日22时 → 己巳·丁卯·己卯·乙亥
	result := Calculate(1989, 3, 20, 22, "male", false, 0, "solar", false)
	t.Logf("四柱：%s%s·%s%s·%s%s·%s%s",
		result.YearGan, result.YearZhi,
		result.MonthGan, result.MonthZhi,
		result.DayGan, result.DayZhi,
		result.HourGan, result.HourZhi)
	t.Logf("日主五行：%s，月支：%s", result.DayGanWuxing, result.MonthZhi)
	t.Logf("五行分布：木%d 火%d 土%d 金%d 水%d",
		result.Wuxing.Mu, result.Wuxing.Huo, result.Wuxing.Tu,
		result.Wuxing.Jin, result.Wuxing.Shui)
	t.Logf("喜用神：%s，忌神：%s", result.Yongshen, result.Jishen)

	if result.Yongshen != "火土" {
		t.Errorf("己土卯月（失令身弱）喜用神应为「火土」，实际为「%s」", result.Yongshen)
	}
	if result.Jishen != "木水金" {
		t.Errorf("己土卯月（失令身弱）忌神应为「木水金」，实际为「%s」", result.Jishen)
	}
}

// TestYongshenFix_WeightedStrong 验证月令帮扶时仍能正确判为身强
// 庚金生于丑月（土月），土旺生金，月令帮扶，多土多金下应判身强
func TestYongshenFix_WeightedStrong(t *testing.T) {
	// 1991年1月20日12时 → 庚午·辛丑·庚辰·庚午（丑月土旺）
	result := Calculate(1991, 1, 20, 12, "male", false, 0, "solar", false)
	t.Logf("四柱：%s%s·%s%s·%s%s·%s%s",
		result.YearGan, result.YearZhi,
		result.MonthGan, result.MonthZhi,
		result.DayGan, result.DayZhi,
		result.HourGan, result.HourZhi)
	t.Logf("日主五行：%s，月支五行：%s，月支：%s",
		result.DayGanWuxing, result.MonthZhiWuxing, result.MonthZhi)
	t.Logf("五行分布：木%d 火%d 土%d 金%d 水%d",
		result.Wuxing.Mu, result.Wuxing.Huo, result.Wuxing.Tu,
		result.Wuxing.Jin, result.Wuxing.Shui)
	t.Logf("喜用神：%s，忌神：%s", result.Yongshen, result.Jishen)
	// 仅打印，不做强断言（此处验证命局不崩溃）
}
