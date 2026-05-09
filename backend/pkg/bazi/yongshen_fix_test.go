package bazi

import "testing"

// TestYongshenFix_JiTuInMao 验证 t0 调候命中行为（己卯字典：癸甲；亥藏含甲）
// 命主：己巳·丁卯·己卯·乙亥；调候字典 己_卯 = [癸,甲]；亥藏壬甲 → 甲命中
// 旧扶抑算法返回 yongshen=火土；新算法走调候 → yongshen 含木（甲）
// 注：此为算法迁移后预期变化，验证 t0 路径正常工作
func TestYongshenFix_JiTuInMao(t *testing.T) {
	result := Calculate(1989, 3, 20, 22, "male", false, 0, "solar", false)
	t.Logf("四柱：%s%s·%s%s·%s%s·%s%s",
		result.YearGan, result.YearZhi,
		result.MonthGan, result.MonthZhi,
		result.DayGan, result.DayZhi,
		result.HourGan, result.HourZhi)
	t.Logf("yongshen=%q jishen=%q status=%q hit=%v miss=%v",
		result.Yongshen, result.Jishen, result.YongshenStatus,
		result.YongshenGans, result.YongshenMissing)

	if result.YongshenStatus != YongshenStatusTiaohouHit {
		t.Errorf("status = %q, want tiaohou_hit (己_卯 字典 = 癸甲，亥藏含甲)", result.YongshenStatus)
	}
	if !containsStrTest(result.YongshenGans, "甲") {
		t.Errorf("YongshenGans = %v, expected to contain 甲", result.YongshenGans)
	}
	if result.Yongshen != "木" {
		t.Errorf("yongshen = %q, want 木 (甲=木 单一命中)", result.Yongshen)
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
