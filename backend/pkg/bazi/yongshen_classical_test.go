package bazi

import (
	"reflect"
	"testing"
)

// ──────────────────────────────────────────────────────────────────────
// ComputeClassicalYongshen 测试
//
// 严格遵循用户命理公式：
//   1. 寒热判定（月令派）
//      • 至寒 = 亥子丑月 且 原局无丙丁巳午
//      • 至热 = 巳午未月 且 原局无壬癸亥子
//   2. 扶抑（calcWeightedYongshen 月支+2加权，helpPct > 40% = 身强）
//   3. 透干优先选 primary_gan
// ──────────────────────────────────────────────────────────────────────

// 主测试盘：1996-02-08 20:00 男
// 四柱：丙子 庚寅 乙亥 丙戌
// 应得 strategy=fuyi_strong, IsStrong=true, WuxingSet="金土火", JishenSet="水木", PrimaryGan="丙"
func TestComputeClassicalYongshen_身强寅月乙木(t *testing.T) {
	natal := &BaziResult{
		YearGan: "丙", YearZhi: "子",
		MonthGan: "庚", MonthZhi: "寅",
		DayGan: "乙", DayZhi: "亥",
		HourGan: "丙", HourZhi: "戌",
		YearHideGan:  []string{"癸"},      // 子藏癸
		MonthHideGan: []string{"甲", "丙", "戊"}, // 寅藏甲丙戊
		DayHideGan:   []string{"壬", "甲"}, // 亥藏壬甲
		HourHideGan:  []string{"戊", "辛", "丁"}, // 戌藏戊辛丁
		Wuxing: WuxingStats{
			Mu: 2, Huo: 2, Tu: 1, Jin: 1, Shui: 2, Total: 8,
		},
	}

	got := ComputeClassicalYongshen(natal)

	if got.Strategy != "fuyi_strong" {
		t.Errorf("Strategy = %q, want fuyi_strong", got.Strategy)
	}
	if !got.IsStrong {
		t.Errorf("IsStrong = false, want true (乙木寅月建禄+亥子双印=身强)")
	}
	if got.WuxingSet != "金土火" {
		t.Errorf("WuxingSet = %q, want 金土火 (身强→克泄耗)", got.WuxingSet)
	}
	if got.JishenSet != "水木" {
		t.Errorf("JishenSet = %q, want 水木", got.JishenSet)
	}
	if got.PrimaryGan != "丙" {
		t.Errorf("PrimaryGan = %q, want 丙 (火土金候选中丙最早透干)", got.PrimaryGan)
	}
}

// 至寒盘：子月乙木，原局无丙丁巳午
// 期望走 tiaohou_cold 路径，WuxingSet="火"
func TestComputeClassicalYongshen_至寒走调候(t *testing.T) {
	natal := &BaziResult{
		YearGan: "庚", YearZhi: "申",
		MonthGan: "戊", MonthZhi: "子",
		DayGan: "乙", DayZhi: "亥",
		HourGan: "癸", HourZhi: "未",
		YearHideGan:  []string{"庚", "壬", "戊"},
		MonthHideGan: []string{"癸"},
		DayHideGan:   []string{"壬", "甲"},
		HourHideGan:  []string{"己", "乙", "丁"}, // 未含丁——这是个边界情形，原局藏丁
		Wuxing: WuxingStats{
			Mu: 1, Huo: 0, Tu: 2, Jin: 1, Shui: 3, Total: 7,
		},
	}

	// 注：此盘虽然月支=子，但时支未藏丁火 → 严格按"月令派"规则的字面定义
	// 原局透干（年月日时四个天干位）无丙丁，地支位无巳午
	// 但藏干中有丁（未藏丁）。月令派古法严格只看透干和地支本气
	got := ComputeClassicalYongshen(natal)

	if got.Strategy != "tiaohou_cold" {
		t.Errorf("Strategy = %q, want tiaohou_cold", got.Strategy)
	}
	if got.WuxingSet != "火" {
		t.Errorf("WuxingSet = %q, want 火 (至寒用火暖局)", got.WuxingSet)
	}
	if got.JishenSet != "水" {
		t.Errorf("JishenSet = %q, want 水", got.JishenSet)
	}
}

// 至热盘：午月丙火，原局无壬癸亥子
func TestComputeClassicalYongshen_至热走调候(t *testing.T) {
	natal := &BaziResult{
		YearGan: "戊", YearZhi: "午",
		MonthGan: "戊", MonthZhi: "午",
		DayGan: "丙", DayZhi: "寅",
		HourGan: "甲", HourZhi: "午",
		YearHideGan:  []string{"丁", "己"},
		MonthHideGan: []string{"丁", "己"},
		DayHideGan:   []string{"甲", "丙", "戊"},
		HourHideGan:  []string{"丁", "己"},
		Wuxing: WuxingStats{
			Mu: 2, Huo: 4, Tu: 2, Jin: 0, Shui: 0, Total: 8,
		},
	}

	got := ComputeClassicalYongshen(natal)

	if got.Strategy != "tiaohou_hot" {
		t.Errorf("Strategy = %q, want tiaohou_hot", got.Strategy)
	}
	if got.WuxingSet != "水" {
		t.Errorf("WuxingSet = %q, want 水 (至热用水降温)", got.WuxingSet)
	}
}

// 不属至寒（亥子丑月但原局有火字）
func TestComputeClassicalYongshen_冬月有火不走调候(t *testing.T) {
	// 子月乙木 + 时干透丙 → 不再"至寒"
	natal := &BaziResult{
		YearGan: "庚", YearZhi: "申",
		MonthGan: "戊", MonthZhi: "子",
		DayGan: "乙", DayZhi: "卯",
		HourGan: "丙", HourZhi: "戌", // 时干丙
		YearHideGan:  []string{"庚", "壬", "戊"},
		MonthHideGan: []string{"癸"},
		DayHideGan:   []string{"乙"},
		HourHideGan:  []string{"戊", "辛", "丁"},
		Wuxing: WuxingStats{
			Mu: 2, Huo: 1, Tu: 2, Jin: 1, Shui: 1, Total: 7,
		},
	}

	got := ComputeClassicalYongshen(natal)

	if got.Strategy == "tiaohou_cold" {
		t.Errorf("Strategy = tiaohou_cold, want 非调候路径 (透干有丙就不再至寒)")
	}
}

// 透干优先级测试：用神五行集多选时，按 甲乙丙丁戊…癸 顺序选第一个透干
func TestSelectPrimaryGan_透干优先(t *testing.T) {
	natal := &BaziResult{
		YearGan:  "庚", // 金
		MonthGan: "丙", // 火
		DayGan:   "甲", // 木
		HourGan:  "戊", // 土
	}

	// 用神集="金土火" → 候选: 丙 丁 戊 己 庚 辛
	//   透干: 丙(火) ✓  戊(土) ✓  庚(金) ✓
	//   甲乙丙丁顺序 → 丙 优先
	got := selectPrimaryGan(natal, "金土火")
	if got != "丙" {
		t.Errorf("selectPrimaryGan = %q, want 丙 (透干中按顺序第一)", got)
	}
}

func TestSelectPrimaryGan_无透干用藏干(t *testing.T) {
	natal := &BaziResult{
		YearGan: "庚", MonthGan: "戊", DayGan: "甲", HourGan: "壬", // 透干全无火
		YearHideGan:  []string{"丁"},
		MonthHideGan: []string{"甲"},
		DayHideGan:   []string{},
		HourHideGan:  []string{},
	}
	// 用神集="火" → 候选: 丙丁
	//   透干: 无
	//   藏干: 丁 (年支藏)
	got := selectPrimaryGan(natal, "火")
	if got != "丁" {
		t.Errorf("selectPrimaryGan = %q, want 丁 (无透→取藏)", got)
	}
}

func TestSelectPrimaryGan_无透干无藏干用候选首位(t *testing.T) {
	natal := &BaziResult{
		YearGan: "庚", MonthGan: "壬", DayGan: "甲", HourGan: "乙",
		YearHideGan:  []string{},
		MonthHideGan: []string{},
		DayHideGan:   []string{},
		HourHideGan:  []string{},
	}
	got := selectPrimaryGan(natal, "火")
	// 甲乙丙丁顺序 → 丙
	if got != "丙" {
		t.Errorf("selectPrimaryGan = %q, want 丙 (全无→按顺序首位)", got)
	}
}

// 忌神天干列表派生测试
func TestCollectJishenGans(t *testing.T) {
	got := collectJishenGans("水木")
	want := []string{"甲", "乙", "壬", "癸"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("collectJishenGans(水木) = %v, want %v", got, want)
	}
}

// 寒热判定单元测试
func TestIsExtremeCold_命中(t *testing.T) {
	gans := []string{"庚", "戊", "乙", "癸"}
	zhis := []string{"申", "子", "亥", "未"}
	if !isExtremeCold("子", gans, zhis) {
		t.Errorf("isExtremeCold 应返回 true (子月+无丙丁巳午)")
	}
}

func TestIsExtremeCold_有火不命中(t *testing.T) {
	gans := []string{"庚", "丙", "乙", "癸"} // 月干丙
	zhis := []string{"申", "子", "亥", "未"}
	if isExtremeCold("子", gans, zhis) {
		t.Errorf("isExtremeCold 应返回 false (透干有丙)")
	}
}

func TestIsExtremeHot_命中(t *testing.T) {
	gans := []string{"戊", "戊", "丙", "甲"}
	zhis := []string{"午", "午", "寅", "午"}
	if !isExtremeHot("午", gans, zhis) {
		t.Errorf("isExtremeHot 应返回 true (午月+无壬癸亥子)")
	}
}
