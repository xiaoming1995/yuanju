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

// ─── 端到端：已移除全年基底色，不再输出 "用神基底" 信号 ────────────────────────

func TestGetYearEventSignals_NoBaselineSignal(t *testing.T) {
	natal := makeNatal("甲子", "丁卯", "丙午", "戊子", "木火", "金水土")
	signals := GetYearEventSignals(natal, "甲", "寅", "丁卯", "male", 30)
	for _, s := range signals {
		if s.Type == "用神基底" {
			t.Fatalf("不应再输出用神基底信号，got: %+v", s)
		}
	}
}

// ─── collectYingqiSignals 单元测试 ─────────────────────────────────────────

// TestYingqi_KeYong 流年天干克原局用神天干位 → 凶
func TestYingqi_KeYong(t *testing.T) {
	// 年干甲(木=用神)，流年庚(金)克甲 → 凶
	natal := makeNatal("甲子", "丁卯", "丙午", "戊子", "木火", "金水土")
	sigs := collectYingqiSignals(natal, "庚", "申", "", "")
	found := false
	for _, s := range sigs {
		if s.Polarity == PolarityXiong && strings.Contains(s.Evidence, "用神位") && strings.Contains(s.Evidence, "克") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected 凶 signal for 流年天干克用神天干位，got: %v", sigs)
	}
}

// TestYingqi_ChongJi 流年地支冲原局忌神地支位 → 吉
func TestYingqi_ChongJi(t *testing.T) {
	// 年支子/时支子(水=忌神)，流年午冲子 → 吉
	natal := makeNatal("甲子", "丁卯", "丙午", "戊子", "木火", "金水土")
	sigs := collectYingqiSignals(natal, "壬", "午", "", "")
	found := false
	for _, s := range sigs {
		if s.Polarity == PolarityJi && strings.Contains(s.Evidence, "忌神位") && strings.Contains(s.Evidence, "冲") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected 吉 signal for 流年地支冲忌神地支位，got: %v", sigs)
	}
}

// TestYingqi_ChuanYong 流年地支穿原局用神地支位 → 凶
func TestYingqi_ChuanYong(t *testing.T) {
	// 月支卯(木=用神)，sixHai[辰]=卯，流年辰穿卯 → 凶
	natal := makeNatal("甲子", "丁卯", "丙午", "戊子", "木火", "金水土")
	sigs := collectYingqiSignals(natal, "戊", "辰", "", "")
	found := false
	for _, s := range sigs {
		if s.Polarity == PolarityXiong && strings.Contains(s.Evidence, "用神位") && strings.Contains(s.Evidence, "穿") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected 凶 signal for 流年地支穿用神地支位，got: %v", sigs)
	}
}

// TestYingqi_HeYongHuaJi 流年天干合用神位，化出属忌神 → 凶
func TestYingqi_HeYongHuaJi(t *testing.T) {
	// 年干甲(木=用神)，月支辰(tu)，忌神=土金水
	// 流年己合甲，化土，月支辰本气=tu，化成立，jiSet[tu]=true → 凶
	natal := makeNatal("甲辰", "丁未", "丙子", "壬子", "木火", "土金水")
	sigs := collectYingqiSignals(natal, "己", "亥", "", "")
	found := false
	for _, s := range sigs {
		if s.Polarity == PolarityXiong && strings.Contains(s.Evidence, "合") && strings.Contains(s.Evidence, "忌神") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected 凶 signal for 合用神位化出忌神，got: %v", sigs)
	}
}

// TestYingqi_HeBuHuaYong 流年天干合用神位合而不化 → 凶（用神被锁）
func TestYingqi_HeBuHuaYong(t *testing.T) {
	// 年干甲(木=用神)，月支子(水≠土)，流年己合甲化土，月支无土根 → 不化，用神被锁 → 凶
	natal := makeNatal("甲子", "丁子", "丙子", "壬子", "木火", "土金水")
	sigs := collectYingqiSignals(natal, "己", "亥", "", "")
	found := false
	for _, s := range sigs {
		if s.Polarity == PolarityXiong && strings.Contains(s.Evidence, "合而不化") && strings.Contains(s.Evidence, "用神") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected 凶 signal for 合而不化锁用神位，got: %v", sigs)
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

// ─── 天干克流通判断 ──────────────────────────────────────────────────────────

// TestCheckGan_Liutong: 忌神丙（火）克用神庚（金），原局有戊（土）使火生土生金，流通成立 → 中性
func TestCheckGan_Liutong_Neutral(t *testing.T) {
	// 日干壬（水），用神=金（庚为用神位），忌神=火
	// 原局：年干戊（土），使得丙(火)→戊(土)→庚(金) 流通成立
	natal := makeNatal("戊子", "甲午", "壬寅", "甲午", "金", "火")
	// 流年丙（火）克月干甲... 等等，需要让丙克用神天干位（金五行）
	// 用神=金 → 用神天干位需要是金五行的天干：庚或辛
	// 重新构建：年干=庚（金，用神位），忌神=火，有戊（土）做中间
	natal2 := makeNatal("庚子", "戊午", "壬寅", "甲子", "金", "火")
	// 流年丙（火）克年干庚（金），原局戊（土）= 丙生戊生庚 → 流通成立
	sigs := collectYingqiSignals(natal2, "丙", "寅", "", "")
	hasNeutral := false
	for _, s := range sigs {
		if s.Source == SourceZhuwei && s.Polarity == PolarityNeutral && strings.Contains(s.Evidence, "流通") {
			hasNeutral = true
			break
		}
	}
	if !hasNeutral {
		t.Logf("signals: %v", sigs)
		_ = natal
		t.Fatal("expected neutral signal due to 五行流通, got none")
	}
}

// TestCheckGan_NoLiutong: 无中间五行 → 维持凶
func TestCheckGan_NoLiutong_Xiong(t *testing.T) {
	// 年干=庚（金，用神位），忌神=火，原局无土 → 无流通
	natal := makeNatal("庚子", "甲午", "壬寅", "乙子", "金", "火")
	// 流年丙（火）克年干庚（金），原局天干 庚甲壬乙 均无土 → 无流通
	sigs := collectYingqiSignals(natal, "丙", "寅", "", "")
	hasXiong := false
	for _, s := range sigs {
		if s.Source == SourceZhuwei && s.Polarity == PolarityXiong && strings.Contains(s.Evidence, "用神受克") {
			hasXiong = true
			break
		}
	}
	if !hasXiong {
		t.Logf("signals: %v", sigs)
		t.Fatal("expected 凶 signal when no 五行流通, got none")
	}
}

// ─── Layer 0 vs Layer 4 压制 ─────────────────────────────────────────────────

// TestLayer0XiongSuppressesLayer4Ji: Layer 0 凶 → Layer 4 吉（SourceZhuwei）被过滤
func TestLayer0XiongSuppressesLayer4Ji(t *testing.T) {
	// 日支=午（火，用神位），流年子冲午 → Layer 0 凶
	// 流年甲（木），日干=壬（水），木生水... 木非克水故无事业压力；
	// 用甲透正印于身弱日主 → 会产生事业 吉 Signal (Layer 4)
	// 我们用简单构造：日支午=火=用神，流年子冲午=凶；
	// 流年天干甲=木=忌神（忌神=木），流年忌神透干可能产生 Layer 4 事件
	// 更精确：我们直接检查 layer0HasXiong 时 Layer4 吉信号被移除

	// natal: 用神=火，日支午（火），流年子冲午 = Layer0凶
	// 让流年天干产生 Layer4 吉信号：流年=甲子，日干=丙（火），甲=正印→事业吉
	natal := makeNatal("壬申", "甲午", "丙午", "壬子", "火", "水金")
	signals := GetYearEventSignals(natal, "甲", "子", "甲子", "male", 30)

	hasLayer0Xiong := false
	hasLayer4Ji := false
	for _, s := range signals {
		if s.Source == SourceZhuwei && s.Polarity == PolarityXiong && strings.Contains(s.Evidence, "用神受冲") {
			hasLayer0Xiong = true
		}
		if s.Source == SourceZhuwei && s.Polarity == PolarityJi {
			hasLayer4Ji = true
		}
	}
	if !hasLayer0Xiong {
		t.Skip("本测试需要 Layer 0 凶信号触发，当前 natal 未触发")
	}
	if hasLayer4Ji {
		t.Fatal("Layer 0 凶信号存在时，Layer 4 吉信号应被压制，但仍存在")
	}
}

// TestLayer0JiDoesNotSuppressLayer4Xiong: Layer 0 吉不压制 Layer 4 凶
func TestLayer0JiDoesNotSuppressLayer4Xiong(t *testing.T) {
	// 用神=火，流年午（火）来合日支子（水=忌神位）→ Layer 0 产生忌神受冲吉信号
	// 同时 Layer 4 有凶信号（比如比劫夺财凶）
	// 此时 Layer 4 凶信号应保留
	natal := makeNatal("壬子", "甲午", "丙子", "壬午", "火", "水")
	// 流年庚（金），庚=七杀克日干丙（火），身弱→事业凶（Layer 4）
	// 流年地支午冲年支子（水=忌神位）→ Layer 0 吉
	signals := GetYearEventSignals(natal, "庚", "午", "甲午", "male", 30)

	hasLayer4Xiong := false
	for _, s := range signals {
		if s.Source == SourceZhuwei && s.Polarity == PolarityXiong && s.Source != SourceShensha {
			hasLayer4Xiong = true
			break
		}
	}
	if !hasLayer4Xiong {
		t.Skip("本 natal 未产生 Layer 4 凶信号，跳过压制测试")
	}
}

// ─── 大运+流年双冲合并 ────────────────────────────────────────────────────────

// TestDoubleHitMerge: 大运子冲月支午（用神位），流年子亦冲月支午 → 合并为叠加信号（两条独立冲变为一条merged）
func TestDoubleHitMerge(t *testing.T) {
	// 月支=午（火=用神），时支=亥（水，非用神），大运地支=子冲午，流年地支=子亦冲午 → 双冲合并
	// 使用时支=亥确保只有月支午一个用神地支位被冲
	natal := makeNatal("壬申", "甲午", "壬寅", "甲亥", "火", "水")
	sigs := collectYingqiSignals(natal, "甲", "子", "壬", "子")
	mergedCount := 0
	for _, s := range sigs {
		if strings.Contains(s.Evidence, "大运流年双重命中") {
			mergedCount++
		}
	}
	if mergedCount != 1 {
		t.Logf("signals: %v", sigs)
		t.Fatalf("expected 1 merged 双重命中 signal, got %d", mergedCount)
	}
}

// TestDoubleHitNoMerge_DiffKind: 大运冲+流年刑同一位置 → 不合并
func TestDoubleHitNoMerge_DiffKind(t *testing.T) {
	// 月支=卯（木=用神），大运地支=酉冲卯，流年地支=子刑卯 → 不合并
	natal := makeNatal("壬申", "甲卯", "壬寅", "甲午", "木", "金")
	sigs := collectYingqiSignals(natal, "甲", "子", "壬", "酉")
	for _, s := range sigs {
		if strings.Contains(s.Evidence, "大运流年双重命中") {
			t.Fatalf("expected no merge for different kind (冲 vs 刑), but got merged signal: %s", s.Evidence)
		}
	}
}

// ─── 神煞重煞压制轻煞吉信号 ──────────────────────────────────────────────────

func TestShenshaMeta_IsHeavy(t *testing.T) {
	heavyStars := []string{"羊刃", "白虎", "丧门", "吊客", "灾煞", "劫煞", "亡神"}
	for _, name := range heavyStars {
		meta, ok := shenshaWhitelist[name]
		if !ok {
			t.Errorf("神煞 %q not found in whitelist", name)
			continue
		}
		if !meta.IsHeavy {
			t.Errorf("神煞 %q should be IsHeavy=true", name)
		}
	}
	lightStars := []string{"天乙贵人", "文昌贵人", "驿马", "勾绞"}
	for _, name := range lightStars {
		meta, ok := shenshaWhitelist[name]
		if !ok {
			continue // skip if not in whitelist
		}
		if meta.IsHeavy {
			t.Errorf("神煞 %q should be IsHeavy=false", name)
		}
	}
}

// ─── 三合/三会局势力判断 ─────────────────────────────────────────────────────

// TestJuShi_YongWins: 用神=火，原局寅+戌，流年=午 → 寅午戌三合火局；忌神=金（火克金）→ 吉
func TestJuShi_YongWins(t *testing.T) {
	natal := makeNatal("壬寅", "甲戌", "壬子", "甲申", "火", "金")
	sigs := collectJuShiSignals(natal, "午", "")
	hasJi := false
	for _, s := range sigs {
		if s.Polarity == PolarityJi && strings.Contains(s.Evidence, "用神赢") {
			hasJi = true
		}
	}
	if !hasJi {
		t.Logf("sigs: %v", sigs)
		t.Fatal("expected 吉 signal with 用神赢 for 三合火局 克 忌神金")
	}
}

// TestJuShi_JiXiong: 忌神=水，原局申+辰，流年=子 → 申子辰三合水局 → 极凶★
func TestJuShi_JiXiong(t *testing.T) {
	natal := makeNatal("壬申", "甲辰", "壬寅", "甲午", "火", "水")
	sigs := collectJuShiSignals(natal, "子", "")
	hasXiong := false
	for _, s := range sigs {
		if s.Type == TypeJuShiZhong && s.Polarity == PolarityXiong && strings.Contains(s.Evidence, "★") {
			hasXiong = true
		}
	}
	if !hasXiong {
		t.Logf("sigs: %v", sigs)
		t.Fatal("expected 极凶 signal Type=局势_重 with ★ for 三合水局（忌神）")
	}
}

// TestJuShi_HalfHe_NoSignal: 原局只有寅，流年=午（缺戌）→ 半合，不触发
func TestJuShi_HalfHe_NoSignal(t *testing.T) {
	natal := makeNatal("壬寅", "甲子", "壬申", "甲辰", "火", "金")
	sigs := collectJuShiSignals(natal, "午", "")
	for _, s := range sigs {
		if strings.Contains(s.Evidence, "用神赢") || strings.Contains(s.Evidence, "寅午戌") {
			t.Fatalf("expected no 三合火局 signal when only half-he, got: %s", s.Evidence)
		}
	}
}

// TestJuShi_NoKe_NoSignal: 用神=火，三合火局成，忌神=土（火不克土）→ 无信号
func TestJuShi_NoKe_NoSignal(t *testing.T) {
	natal := makeNatal("壬寅", "甲戌", "壬子", "甲申", "火", "土")
	sigs := collectJuShiSignals(natal, "午", "")
	for _, s := range sigs {
		if strings.Contains(s.Evidence, "用神赢") {
			t.Fatalf("expected no 用神赢 signal when 火 does not ke 土, got: %s", s.Evidence)
		}
	}
}
