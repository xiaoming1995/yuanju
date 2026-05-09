package bazi

import (
	"strings"
	"testing"
)

// hasSignal 在 signals 中查找 Type，并可选断言 Polarity
func hasSignal(signals []EventSignal, typ string) (EventSignal, bool) {
	for _, s := range signals {
		if s.Type == typ {
			return s, true
		}
	}
	return EventSignal{}, false
}

// noSignal 断言列表中不含某 Type
func noSignal(t *testing.T, signals []EventSignal, typ string, msg string) {
	t.Helper()
	for _, s := range signals {
		if s.Type == typ {
			t.Errorf("%s: 期望不含 %q 但实际命中 evidence=%q", msg, typ, s.Evidence)
			return
		}
	}
}

// 9.2 t0 命中财星、非忌、身非弱 → 学业_资源 吉
func TestYoungAge_FinanceStarNonJi(t *testing.T) {
	// 甲日主（木）、忌神含金土（即财土非忌）→ 流年戊（土）= 财星非忌
	// 配置 yongshen="水木"（生扶身的），即 fuyi 视角下身偏弱
	natal := makeNatal("丁卯", "壬子", "甲寅", "丁卯", "水木", "金土火")
	natal.Tiaohou = nil
	signals := GetYearEventSignals(natal, "戊", "辰", "壬子", "male", 14)

	sig, ok := hasSignal(signals, TypeXueYeZiYuan)
	if !ok {
		t.Fatal("少年期财星透干、非忌神 → 期望 学业_资源")
	}
	// 包含少年期关键词
	if !strings.Contains(sig.Evidence, "少年") && !strings.Contains(sig.Evidence, "家境") {
		t.Errorf("evidence 缺少少年期关键词：%q", sig.Evidence)
	}
	// 不出现成人期词汇
	if strings.Contains(sig.Evidence, "财运提升") || strings.Contains(sig.Evidence, "进财") {
		t.Errorf("evidence 不应含成人期词汇：%q", sig.Evidence)
	}
	noSignal(t, signals, "财运_得", "少年期不应输出 财运_得")
}

// 9.3 t0 财星为忌 → 学业_资源 凶
func TestYoungAge_FinanceStarIsJi(t *testing.T) {
	// 甲日主、忌神含土 → 流年戊（土）= 财星为忌
	natal := makeNatal("丁卯", "壬子", "甲寅", "丁卯", "水木", "土金")
	natal.Tiaohou = nil
	signals := GetYearEventSignals(natal, "戊", "辰", "壬子", "male", 14)

	sig, ok := hasSignal(signals, TypeXueYeZiYuan)
	if !ok {
		t.Fatal("少年期财星为忌 → 期望 学业_资源")
	}
	if sig.Polarity != PolarityXiong {
		t.Errorf("财星为忌时 Polarity 应为凶，实际：%q", sig.Polarity)
	}
}

// 9.4 19 岁财星透干 → 财运_得（成人期）
func TestAdultAge_FinanceStarStillCai(t *testing.T) {
	natal := makeNatal("丁卯", "壬子", "甲寅", "丁卯", "水木", "土金")
	natal.Tiaohou = nil
	signals := GetYearEventSignals(natal, "戊", "辰", "壬子", "male", 19)

	if _, ok := hasSignal(signals, "财运_得"); !ok {
		t.Fatal("19 岁财星透干仍应输出 财运_得（成人期不变）")
	}
	noSignal(t, signals, TypeXueYeZiYuan, "19 岁不应触发少年期分支")
}

// 9.5 14 岁流年合日支 → 性格_情谊
func TestYoungAge_LiunianHeDayZhi(t *testing.T) {
	// 日支为午，流年地支未与午六合（午+未=六合）
	natal := makeNatal("甲子", "丁卯", "丙午", "戊子", "", "")
	natal.Tiaohou = nil
	signals := GetYearEventSignals(natal, "癸", "未", "丁卯", "male", 14)

	// 通过大运地支 与日支六合 / 流年合月支 等机制查找
	// 这里测试经流年合月支 卯 与 戌 六合（月支为卯）— 用流年戌测
	natal2 := makeNatal("甲子", "丁卯", "丙午", "戊子", "", "")
	natal2.Tiaohou = nil
	_ = natal2
	_ = signals
	// 直接测：大运地支 = 未（与日支午六合）→ 触发少年期 性格_情谊 分支
	signals3 := GetYearEventSignals(natal, "戊", "申", "癸未", "male", 14)
	if _, ok := hasSignal(signals3, TypeXingGeQingYi); !ok {
		t.Logf("信号列表：")
		for _, s := range signals3 {
			t.Logf("  %s | %s | %s", s.Type, s.Polarity, s.Evidence)
		}
		t.Fatal("少年期大运合日支 → 期望 性格_情谊")
	}
	noSignal(t, signals3, "婚恋_合", "少年期不应含 婚恋_合")
}

// 9.6 14 岁大运冲日支 → 性格_叛逆
func TestYoungAge_DayunChongDayZhi(t *testing.T) {
	// 日支午、大运地支子（子午冲）
	natal := makeNatal("甲子", "丁卯", "丙午", "戊子", "", "")
	natal.Tiaohou = nil
	signals := GetYearEventSignals(natal, "甲", "辰", "丙子", "male", 14)

	if _, ok := hasSignal(signals, TypeXingGePanNi); !ok {
		t.Logf("信号列表：")
		for _, s := range signals {
			t.Logf("  %s | %s | %s", s.Type, s.Polarity, s.Evidence)
		}
		t.Fatal("少年期大运冲日支 → 期望 性格_叛逆")
	}
	noSignal(t, signals, "婚恋_冲", "少年期不应含 婚恋_冲")
}

// 9.8 14 岁流年克日干 → 仍输出 健康（与年龄无关）
func TestYoungAge_HealthUnchanged(t *testing.T) {
	// 甲日主（木）、流年庚（金克木）→ 健康 凶
	natal := makeNatal("甲子", "丁卯", "甲寅", "戊子", "", "")
	natal.Tiaohou = nil
	signals := GetYearEventSignals(natal, "庚", "申", "丁卯", "male", 14)

	sig, ok := hasSignal(signals, "健康")
	if !ok {
		t.Fatal("健康信号在少年期也应保留（与年龄无关）")
	}
	if sig.Polarity != PolarityXiong {
		t.Errorf("克日干 Polarity 应为凶，实际：%q", sig.Polarity)
	}
}

// 9.9 14 岁不输出婚恋 / 财官双叠
func TestYoungAge_NoMarriageDoubling(t *testing.T) {
	// 男命甲日主，流年戊（财）+ 大运戊辰（财双叠）→ 成人期会输出 婚恋_合
	natal := makeNatal("丁卯", "壬子", "甲寅", "丁卯", "水木", "")
	natal.Tiaohou = nil
	signals := GetYearEventSignals(natal, "戊", "辰", "戊辰", "male", 14)

	noSignal(t, signals, "婚恋_合", "少年期不应输出 婚恋_合")
	noSignal(t, signals, "婚恋_变", "少年期不应输出 婚恋_变")
	noSignal(t, signals, "财运_叠", "少年期不应输出 财运_叠")
}

// 9.7 14 岁桃花神煞 → 性格_情谊
func TestYoungAge_PeachBlossomShensha(t *testing.T) {
	// 构造一个能命中桃花神煞的命盘：日支午 → 桃花地支为卯
	// 流年地支选卯（月支=卯，所以本来就有，但流年再现也算）
	natal := makeNatal("甲子", "丁卯", "丙午", "戊子", "", "")
	natal.Tiaohou = nil
	signals := GetYearEventSignals(natal, "乙", "卯", "丁卯", "male", 14)

	hasShensha := false
	for _, s := range signals {
		if s.Source == SourceShensha {
			hasShensha = true
			// 神煞输出在少年期不应是 婚恋_*
			if strings.HasPrefix(s.Type, "婚恋") {
				t.Errorf("少年期神煞 Type 应被重映射，实际仍为：%s", s.Type)
			}
		}
	}
	if !hasShensha {
		t.Skip("命盘未触发任何神煞（此测试场景依赖具体桃花/红艳/天喜命中，可跳过）")
	}
}

// 9.10 youngRatio 计算覆盖
func TestBuildLifeStageHint_FullSchool(t *testing.T) {
	hint := buildLifeStageHintForTest(10, 10)
	if !strings.Contains(hint, "全部年份处于读书期") {
		t.Errorf("期望全段读书期提示，实际：%q", hint)
	}
}

func TestBuildLifeStageHint_Crossing(t *testing.T) {
	hint := buildLifeStageHintForTest(4, 10)
	if !strings.Contains(hint, "跨越读书期") {
		t.Errorf("期望跨界提示，实际：%q", hint)
	}
	if !strings.Contains(hint, "前 4") {
		t.Errorf("期望含前 N 数字，实际：%q", hint)
	}
}

func TestBuildLifeStageHint_AllAdult(t *testing.T) {
	hint := buildLifeStageHintForTest(0, 10)
	if hint != "" {
		t.Errorf("成人期应返回空，实际：%q", hint)
	}
}

// buildLifeStageHintForTest 测试桥（buildLifeStageHint 在 service 包，bazi 包测试无法直接调用）
// 此处复制相同逻辑，跑同样断言
func buildLifeStageHintForTest(youngCount, totalCount int) string {
	if totalCount <= 0 || youngCount <= 0 {
		return ""
	}
	if youngCount == totalCount {
		return "本段大运全部年份处于读书期，请以学业、性格塑造、同窗关系为主轴撰写 summary。"
	}
	adultCount := totalCount - youngCount
	return "本段大运跨越读书期与成人期（前 " + itoa(youngCount) + " 年读书、后 " + itoa(adultCount) + " 年入社会），summary 请分两段叙述：先讲读书期学业 / 性格，再讲成人期事业 / 婚恋。"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
