package bazi

import (
	"strings"
	"testing"
)

// makeNatal 构造一个最小的 BaziResult 测试夹具
func makeNatal(yearGZ, monthGZ, dayGZ, hourGZ, yongshen, jishen string) *BaziResult {
	yr := []rune(yearGZ)
	mr := []rune(monthGZ)
	dr := []rune(dayGZ)
	hr := []rune(hourGZ)
	return &BaziResult{
		YearGan: string(yr[0]), YearZhi: string(yr[1]),
		MonthGan: string(mr[0]), MonthZhi: string(mr[1]),
		DayGan: string(dr[0]), DayZhi: string(dr[1]),
		HourGan: string(hr[0]), HourZhi: string(hr[1]),
		Yongshen: yongshen,
		Jishen:   jishen,
	}
}

// ─── 用神/忌神基底色 ─────────────────────────────────────────────────────────

func TestYongshenBaseline_Yong(t *testing.T) {
	natal := makeNatal("甲子", "丁卯", "丙午", "戊子", "木火", "金水土")
	pol, ev := getYongshenBaseline(natal, "甲") // 甲=木=用神
	if pol != PolarityJi {
		t.Fatalf("expected PolarityJi, got %q", pol)
	}
	if !strings.Contains(ev, "用神") {
		t.Fatalf("expected evidence to mention 用神, got %q", ev)
	}
}

func TestYongshenBaseline_Ji(t *testing.T) {
	natal := makeNatal("甲子", "丁卯", "丙午", "戊子", "木火", "金水土")
	pol, ev := getYongshenBaseline(natal, "庚") // 庚=金=忌神
	if pol != PolarityXiong {
		t.Fatalf("expected PolarityXiong, got %q", pol)
	}
	if !strings.Contains(ev, "忌神") {
		t.Fatalf("expected evidence to mention 忌神, got %q", ev)
	}
}

func TestYongshenBaseline_Neutral(t *testing.T) {
	natal := makeNatal("甲子", "丁卯", "丙午", "戊子", "木", "金")
	pol, _ := getYongshenBaseline(natal, "丙") // 丙=火，既非木也非金
	if pol != PolarityNeutral {
		t.Fatalf("expected PolarityNeutral, got %q", pol)
	}
}

func TestYongshenBaseline_Missing(t *testing.T) {
	natal := makeNatal("甲子", "丁卯", "丙午", "戊子", "", "")
	pol, ev := getYongshenBaseline(natal, "甲")
	if pol != "" || ev != "" {
		t.Fatalf("expected empty when both yongshen/jishen missing, got pol=%q ev=%q", pol, ev)
	}
}

// ─── 加权身强弱评分 ──────────────────────────────────────────────────────────

func TestStrengthLevel_VStrong(t *testing.T) {
	// 丙日主，月支午（同气），其他全是火/木（生扶）
	natal := makeNatal("甲寅", "甲午", "丙寅", "甲午", "", "")
	level, score, _ := dayMasterStrengthLevel(natal)
	if level != "vstrong" {
		t.Fatalf("expected vstrong, got %s (score=%d)", level, score)
	}
}

func TestStrengthLevel_VWeak(t *testing.T) {
	// 丙日主，月支子（克我），其他全是水/金/土（克泄）
	natal := makeNatal("壬子", "壬子", "丙子", "戊戌", "", "")
	level, score, _ := dayMasterStrengthLevel(natal)
	if level != "vweak" && level != "weak" {
		t.Fatalf("expected weak/vweak, got %s (score=%d)", level, score)
	}
}

func TestStrengthLevel_Neutral(t *testing.T) {
	// 月支克我，其他得地，预期落在中和
	natal := makeNatal("甲寅", "壬子", "丙寅", "甲午", "", "")
	level, _, _ := dayMasterStrengthLevel(natal)
	if level != "neutral" && level != "weak" && level != "strong" {
		t.Fatalf("expected one of neutral/weak/strong, got %s", level)
	}
}

// ─── 三会、三刑、伏吟、反吟 ─────────────────────────────────────────────────

func TestSanhuiTriggered_MuLocal(t *testing.T) {
	// 流年寅，原局已含卯辰 → 触发寅卯辰木会
	ok, wx := isSanhuiTriggered("寅", []string{"卯", "辰", "酉"})
	if !ok || wx != "mu" {
		t.Fatalf("expected sanhui mu triggered, got ok=%v wx=%q", ok, wx)
	}
}

func TestSanxingTriggered_YinSiShen(t *testing.T) {
	// 流年申，原局已有寅巳 → 凑齐三刑
	ok, kind := isSanxingTriggered("申", []string{"寅", "巳", "丑"})
	if !ok || kind != "寅巳申" {
		t.Fatalf("expected sanxing 寅巳申, got ok=%v kind=%q", ok, kind)
	}
}

func TestFuyin(t *testing.T) {
	if !isFuyin("丙", "午", "丙", "午") {
		t.Fatal("expected fuyin true")
	}
	if isFuyin("丙", "午", "丁", "午") {
		t.Fatal("expected fuyin false (different gan)")
	}
}

func TestFanyin(t *testing.T) {
	// 庚申 反吟 甲寅: 庚克甲(金克木) + 申冲寅
	if !isFanyin("庚", "申", "甲", "寅") {
		t.Fatal("expected fanyin true for 庚申 vs 甲寅")
	}
	// 丙午 vs 庚申：火克金 + 午冲... 午不冲申 → 不构成反吟
	if isFanyin("丙", "午", "庚", "申") {
		t.Fatal("expected fanyin false")
	}
}

// ─── 旬空 ───────────────────────────────────────────────────────────────────

func TestXunkong_JiaZi(t *testing.T) {
	z1, z2 := getXunkong("甲", "子")
	if z1 != "戌" || z2 != "亥" {
		t.Fatalf("甲子日柱旬空应为戌亥，got %s%s", z1, z2)
	}
}

func TestXunkong_GuiHai(t *testing.T) {
	z1, z2 := getXunkong("癸", "亥")
	if z1 != "子" || z2 != "丑" {
		t.Fatalf("癸亥日柱旬空应为子丑，got %s%s", z1, z2)
	}
}

func TestXunkong_Unknown(t *testing.T) {
	z1, z2 := getXunkong("X", "Y")
	if z1 != "" || z2 != "" {
		t.Fatalf("unknown day pillar should return empty, got %s%s", z1, z2)
	}
}

// ─── 大运合化 ───────────────────────────────────────────────────────────────

func TestDetectDayunHuahe_DingRen(t *testing.T) {
	// 丁日主走壬运，月支寅（木根） → 化神成立（合化木）
	// 原局其他天干无金（金克木 = 反克），构造无金的命局
	natal := makeNatal("甲寅", "丙寅", "丁卯", "乙巳", "", "")
	st := detectDayunHuahe(natal, "壬", "寅")
	if !st.Triggered || st.HuashenWx != "mu" {
		t.Fatalf("expected huahe triggered with mu, got triggered=%v wx=%q", st.Triggered, st.HuashenWx)
	}
}

func TestDetectDayunHuahe_HuaheFails(t *testing.T) {
	// 丁日主走壬运，月支酉（金 = 反克木），合而不化
	natal := makeNatal("庚午", "辛酉", "丁未", "甲辰", "", "")
	st := detectDayunHuahe(natal, "壬", "申")
	if !st.Combined || st.Triggered {
		t.Fatalf("expected combined but not triggered, got combined=%v triggered=%v", st.Combined, st.Triggered)
	}
}

func TestDetectDayunHuahe_NoCombination(t *testing.T) {
	// 丙日主走甲运，无五合关系
	natal := makeNatal("甲寅", "丙寅", "丙午", "甲午", "", "")
	st := detectDayunHuahe(natal, "甲", "寅")
	if st.Combined {
		t.Fatalf("expected no combination, got combined=%v", st.Combined)
	}
}

// ─── 神煞接入：调用不报错且白名单生效 ────────────────────────────────────────

func TestShenshaSignal_Tianyi(t *testing.T) {
	sig, ok := shenshaSignal("天乙贵人", PolarityNeutral)
	if !ok {
		t.Fatal("expected 天乙贵人 in whitelist")
	}
	if sig.Polarity != PolarityJi || sig.Source != SourceShensha {
		t.Fatalf("unexpected sig: %+v", sig)
	}
}

func TestShenshaSignal_Yangren(t *testing.T) {
	sig, ok := shenshaSignal("羊刃", PolarityNeutral)
	if !ok || sig.Polarity != PolarityXiong || sig.Type != "健康" {
		t.Fatalf("unexpected sig: %+v", sig)
	}
}

func TestShenshaSignal_NotInWhitelist(t *testing.T) {
	if _, ok := shenshaSignal("不存在的神煞", ""); ok {
		t.Fatal("expected unknown shensha to return false")
	}
}

// ─── 端到端：财=忌神时财运信号 polarity 翻为凶 ─────────────────────────────

func TestGetYearEventSignals_FinanceIsJishen(t *testing.T) {
	// 甲日主（木），金为官，土为财；忌神含土 → 流年戊（土=财）应为凶
	natal := makeNatal("丁卯", "壬子", "甲寅", "丁卯", "水木", "土金")
	natal.Tiaohou = nil // 屏蔽调候喜神干扰
	signals := GetYearEventSignals(natal, "戊", "辰", "壬子", "male", 30)
	found := false
	for _, s := range signals {
		if s.Type == "财运_得" && s.Polarity == PolarityXiong {
			found = true
			if !strings.Contains(s.Evidence, "忌神") {
				t.Errorf("evidence should mention 忌神: %q", s.Evidence)
			}
			break
		}
	}
	if !found {
		t.Fatal("expected 财运_得 signal with PolarityXiong when 财=忌神")
	}
}

// ─── 端到端：用神基底色信号被注入 ───────────────────────────────────────────

func TestGetYearEventSignals_BaselineSignal(t *testing.T) {
	natal := makeNatal("甲子", "丁卯", "丙午", "戊子", "木火", "金水土")
	signals := GetYearEventSignals(natal, "甲", "寅", "丁卯", "male", 30)
	hasBaseline := false
	for _, s := range signals {
		if s.Type == "用神基底" && s.Source == SourceYongshen {
			hasBaseline = true
			break
		}
	}
	if !hasBaseline {
		t.Fatal("expected 用神基底 signal when yongshen/jishen present")
	}
}

// ─── 端到端：流年与月支冲 → 事业信号 ───────────────────────────────────────

func TestGetYearEventSignals_LnChongMonthZhi(t *testing.T) {
	// 月支寅 → 流年地支申冲月支
	natal := makeNatal("甲子", "丙寅", "丁卯", "庚子", "", "")
	signals := GetYearEventSignals(natal, "庚", "申", "丁卯", "male", 30)
	found := false
	for _, s := range signals {
		if s.Type == "事业" && strings.Contains(s.Evidence, "冲月柱") {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected 事业 signal with 冲月柱 evidence")
	}
}

// ─── 端到端：流年伏吟日柱 ───────────────────────────────────────────────────

func TestGetYearEventSignals_FuyinDayPillar(t *testing.T) {
	natal := makeNatal("甲子", "丙寅", "丁卯", "庚戌", "", "")
	signals := GetYearEventSignals(natal, "丁", "卯", "戊辰", "male", 30)
	found := false
	for _, s := range signals {
		if s.Type == "伏吟" && strings.Contains(s.Evidence, "日柱") {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected 伏吟 signal for 流年 = 日柱")
	}
}

// ─── 端到端：流年落日柱旬空 ─────────────────────────────────────────────────

func TestGetYearEventSignals_KongwangDownweight(t *testing.T) {
	// 甲子日柱 → 旬空戌亥；流年戊戌 → 戌空亡
	natal := makeNatal("丁卯", "丁未", "甲子", "戊辰", "", "")
	signals := GetYearEventSignals(natal, "戊", "戌", "丁未", "male", 30)
	hasKongwangSig := false
	for _, s := range signals {
		if s.Source == SourceKongwang && strings.Contains(s.Evidence, "空") {
			hasKongwangSig = true
			break
		}
	}
	if !hasKongwangSig {
		t.Fatal("expected 空亡 signal when 流年地支 falls in 旬空")
	}
}
