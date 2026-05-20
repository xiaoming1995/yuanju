package bazi

import (
	"strings"
	"testing"
)

func TestRenderYearNarrative_UsesPlainLanguageWithoutTechnicalTerms(t *testing.T) {
	ys := YearSignals{
		Year:        2024,
		Age:         32,
		GanZhi:      "甲辰",
		DayunGanZhi: "庚午",
		Signals: []EventSignal{
			{
				Type:     "事业",
				Evidence: "流年地支辰冲月柱壬戌（提纲），易有行业/职位变动",
				Polarity: PolarityXiong,
				Source:   SourceZhuwei,
			},
			{
				Type:     "财运_得",
				Evidence: "甲透干为偏财，但财星五行为命主忌神，财来财去/破耗",
				Polarity: PolarityXiong,
				Source:   SourceZhuwei,
			},
		},
	}

	got := RenderYearNarrative(ys)
	for _, term := range []string{"流年地支", "月柱", "提纲", "透干", "偏财", "财星", "忌神"} {
		if strings.Contains(got, term) {
			t.Fatalf("narrative leaked technical term %q: %s", term, got)
		}
	}
	if !strings.Contains(got, "工作") && !strings.Contains(got, "事业") {
		t.Fatalf("expected plain career wording, got: %s", got)
	}
	if !strings.Contains(got, "钱") && !strings.Contains(got, "财务") {
		t.Fatalf("expected plain money wording, got: %s", got)
	}
}

func TestRenderYearNarrative_YoungAgeUsesSchoolAndPersonalityWording(t *testing.T) {
	ys := YearSignals{
		Year:   2010,
		Age:    14,
		GanZhi: "庚寅",
		Signals: []EventSignal{
			{
				Type:     TypeXueYeYaLi,
				Evidence: "庚透干为七杀，少年期官星临运，学业上有规则约束或重大考核",
				Polarity: PolarityNeutral,
				Source:   SourceZhuwei,
			},
			{
				Type:     TypeXingGePanNi,
				Evidence: "流年地支寅冲日支申（自我宫位），少年期情绪波动",
				Polarity: PolarityXiong,
				Source:   SourceZhuwei,
			},
		},
	}

	got := RenderYearNarrative(ys)
	if strings.Contains(got, "事业") || strings.Contains(got, "财运") || strings.Contains(got, "婚恋") {
		t.Fatalf("young-age narrative used adult wording: %s", got)
	}
	if !strings.Contains(got, "学习") && !strings.Contains(got, "学业") {
		t.Fatalf("expected school wording, got: %s", got)
	}
}

func TestRenderYearNarrative_AdjacentYoungYearsDoNotRepeatGenericChangeOpening(t *testing.T) {
	years := []YearSignals{
		{
			Year:   2004,
			Age:    9,
			GanZhi: "甲申",
			Signals: []EventSignal{
				{Type: "综合变动", Evidence: "流年地支申合年支巳（祖荫/根基），家族/根基方面易有正向事件", Polarity: PolarityJi, Source: SourceZhuwei},
				{Type: TypeXueYeYaLi, Evidence: "甲透干为七杀，少年期官星临运，学业上有规则约束或重大考核", Polarity: PolarityNeutral, Source: SourceZhuwei},
			},
		},
		{
			Year:   2005,
			Age:    10,
			GanZhi: "乙酉",
			Signals: []EventSignal{
				{Type: "综合变动", Evidence: "流年地支酉合时柱丁辰（子女/晚景宫）", Polarity: PolarityJi, Source: SourceZhuwei},
				{Type: TypeXingGeQingYi, Evidence: "流年地支酉为桃花星临命，少年期人缘旺 / 同窗喜事多", Polarity: PolarityNeutral, Source: SourceZhuwei},
			},
		},
		{
			Year:   2006,
			Age:    11,
			GanZhi: "丙戌",
			Signals: []EventSignal{
				{Type: "综合变动", Evidence: "流年地支戌落日柱旬空（戌亥空），事件虚而不实/过而不留", Polarity: PolarityNeutral, Source: SourceKongwang},
				{Type: "健康", Evidence: "流年天干丙（火）克制日干庚（金），日主元气受损，需注意身体健康", Polarity: PolarityXiong, Source: SourceZhuwei},
			},
		},
	}

	openings := map[string]bool{}
	for _, ys := range years {
		narrative := RenderYearNarrative(ys)
		// Empty narrative is OK under the new contract — it means the year
		// has no evidence-anchored sentences to show. Only enforce
		// uniqueness on years that actually render.
		if narrative == "" {
			continue
		}
		if strings.Contains(narrative, "变化感会比较强") {
			t.Fatalf("young-age narrative used generic repeated change opening: %s", narrative)
		}
		opening := firstSentence(narrative)
		if openings[opening] {
			t.Fatalf("repeated opening sentence %q for narrative: %s", opening, narrative)
		}
		openings[opening] = true
	}
}

func TestRenderYearNarrative_StrongChangeStillDominates(t *testing.T) {
	ys := YearSignals{
		Year:   2012,
		Age:    17,
		GanZhi: "壬辰",
		Signals: []EventSignal{
			{Type: TypeXueYeGuiRen, Evidence: "壬透干为正印，少年期印星护身，得师长指点", Polarity: PolarityJi, Source: SourceZhuwei},
			{Type: "伏吟", Evidence: "流年壬辰伏吟日柱壬辰，主同类事件重现/旧事重提", Polarity: PolarityXiong, Source: SourceFuyin},
		},
	}

	got := RenderYearNarrative(ys)
	if !strings.Contains(got, "旧事") && !strings.Contains(got, "反复") && !strings.Contains(got, "重复") {
		t.Fatalf("expected strong change wording for fuyin, got: %s", got)
	}
	if strings.Contains(got, "伏吟") {
		t.Fatalf("narrative leaked technical fuyin term: %s", got)
	}
}

func TestRenderYearNarrative_DayunPhaseDoesNotDominateSpecificTheme(t *testing.T) {
	ys := YearSignals{
		Year:        2028,
		Age:         32,
		GanZhi:      "戊申",
		DayunGanZhi: "丁申",
		YearInDayun: 7,
		DayunPhase:  DayunPhaseZhi,
		Signals: []EventSignal{
			{Type: TypeDayunPhase, Evidence: "大运第7年进入地支主事阶段：申为金不换忌神地支，后5年运势不利。", Polarity: PolarityXiong, Source: SourceDayunPhase},
			{Type: "健康", Evidence: "流年地支申冲日支寅，日柱受冲，体力精神有下滑风险", Polarity: PolarityXiong, Source: SourceZhuwei},
		},
	}

	got := RenderYearNarrative(ys)
	opening := firstSentence(got)
	if !strings.Contains(opening, "健康") && !strings.Contains(opening, "身体") && !strings.Contains(opening, "出行安全") && !strings.Contains(opening, "作息") {
		t.Fatalf("specific health theme should dominate opening, got: %s", got)
	}
	if strings.Contains(opening, "大运") || strings.Contains(opening, "前5年") || strings.Contains(opening, "后5年") {
		t.Fatalf("dayun phase should not dominate or leak technical wording, got: %s", got)
	}
}

func TestRenderYearNarrative_DayunPhaseStaysOutOfYearlyBody(t *testing.T) {
	ys := YearSignals{
		Year:        2028,
		Age:         32,
		GanZhi:      "戊申",
		DayunGanZhi: "丁申",
		YearInDayun: 7,
		DayunPhase:  DayunPhaseZhi,
		Signals: []EventSignal{
			{Type: "迁变", Evidence: "流年地支申为驿马星，主奔波变动、出行迁移或职位调动", Polarity: PolarityNeutral, Source: SourceZhuwei},
			{Type: TypeDayunPhase, Evidence: "大运第7年进入地支主事阶段：申为金不换忌神地支，后5年运势不利。", Polarity: PolarityXiong, Source: SourceDayunPhase},
		},
	}

	got := RenderYearNarrative(ys)
	for _, term := range []string{"本步大运", "前五年", "后五年", "天干主事", "地支主事", "金不换"} {
		if strings.Contains(got, term) {
			t.Fatalf("yearly body should not repeat dayun phase term %q: %s", term, got)
		}
	}
}

func TestRenderYearNarrative_DayunPhaseEvidenceDoesNotForceVisibleBodyText(t *testing.T) {
	ys := YearSignals{
		Year:        2009,
		Age:         14,
		GanZhi:      "己丑",
		DayunGanZhi: "辛卯",
		YearInDayun: 6,
		DayunPhase:  DayunPhaseZhi,
		Signals: []EventSignal{
			{Type: "健康", Evidence: "流年天干己（土）克制日干癸（水），日主元气受损，需注意身体健康", Polarity: PolarityXiong, Source: SourceZhuwei},
			{Type: TypeXueYeZiYuan, Evidence: "己透干为偏财，少年期家庭资源或学习投入受关注", Polarity: PolarityNeutral, Source: SourceZhuwei},
			{Type: TypeDayunPhase, Evidence: "大运第6年进入地支主事阶段：卯不在金不换喜忌地支之列，后5年中平论之。", Polarity: PolarityNeutral, Source: SourceDayunPhase},
		},
	}

	got := RenderYearNarrative(ys)
	for _, term := range []string{"本步大运", "前五年", "后五年", "天干主事", "地支主事", "金不换"} {
		if strings.Contains(got, term) {
			t.Fatalf("yearly body should stay focused on yearly events, leaked %q: %s", term, got)
		}
	}
}

func TestRenderEvidenceSummary_SelectsTechnicalEvidence(t *testing.T) {
	ys := YearSignals{
		Year:   2025,
		Age:    33,
		GanZhi: "乙巳",
		Signals: []EventSignal{
			{Type: "事业", Evidence: "流年地支巳冲月柱亥（提纲），易有行业/职位变动", Polarity: PolarityXiong, Source: SourceZhuwei},
			{Type: "健康", Evidence: "白虎临运，主孝服、突发伤痛或意外", Polarity: PolarityXiong, Source: SourceShensha},
			{Type: "迁变", Evidence: "流年地支巳为驿马星，主奔波变动、出行迁移", Polarity: PolarityNeutral, Source: SourceZhuwei},
		},
	}

	got := RenderEvidenceSummary(ys)
	if len(got) == 0 {
		t.Fatal("expected evidence summary")
	}
	if !strings.Contains(strings.Join(got, "；"), "流年地支") {
		t.Fatalf("expected technical evidence to be preserved, got: %#v", got)
	}
	if len(got) > 5 {
		t.Fatalf("expected at most 5 evidence items, got %d: %#v", len(got), got)
	}
}

func TestRenderYearNarrative_TenGodPowerDoesNotRescueGenericYear(t *testing.T) {
	// Old behavior: a 10-god power title appended a "...可作为理解这一年事件
	// 走向的背景力量。" wrap, padding generic years into a visible paragraph.
	// New behavior (per 2026-05-18 spec): un-anchored years stay hidden
	// regardless of 10-god power, so the algorithm doesn't fill silence
	// with generic prose.
	ys := YearSignals{
		Year:   2024,
		Age:    29,
		GanZhi: "甲辰",
		TenGodPower: TenGodPowerProfile{
			PlainTitle: "官杀偏旺",
			PlainText:  "规则、考核、责任和外部压力更明显，宜稳住节奏。",
			Reason:     "流年天干为七杀",
		},
		Signals: []EventSignal{
			{Type: "综合变动", Evidence: "流年节奏变化", Polarity: PolarityNeutral, Source: SourceZhuwei},
		},
	}
	if got := RenderYearNarrative(ys); got != "" {
		t.Errorf("expected hidden narrative for un-anchored year with 10-god power; got %q", got)
	}
}

func TestRenderYearNarrative_HardEvidenceNotOverriddenByTenGodPower(t *testing.T) {
	ys := YearSignals{
		Year:   2026,
		Age:    31,
		GanZhi: "丙午",
		TenGodPower: TenGodPowerProfile{
			PlainTitle: "财星偏旺",
			PlainText:  "钱财、资源、合作回报和现实机会更容易被看见。",
		},
		Signals: []EventSignal{
			{Type: "健康", Evidence: "流年地支午冲日支子，日柱受冲，体力精神有下滑风险", Polarity: PolarityXiong, Source: SourceZhuwei},
		},
	}

	got := RenderYearNarrative(ys)
	opening := firstSentence(got)
	if strings.Contains(opening, "财星") {
		t.Fatalf("hard health evidence should dominate opening, got: %s", got)
	}
	if !strings.Contains(opening, "身体") && !strings.Contains(opening, "健康") && !strings.Contains(opening, "作息") {
		t.Fatalf("expected health wording to remain dominant, got: %s", got)
	}
}

func TestRenderYearNarrative_RichSignalYearHasMediumDetail(t *testing.T) {
	ys := YearSignals{
		Year:   2025,
		Age:    30,
		GanZhi: "乙巳",
		TenGodPower: TenGodPowerProfile{
			PlainTitle: "钱财压力明显",
			PlainText:  "钱财、资源和现实事务更突出，也容易伴随支出压力。",
		},
		Signals: []EventSignal{
			{Type: "事业", Evidence: "流年地支巳冲月柱亥（提纲），易有行业/职位变动", Polarity: PolarityXiong, Source: SourceZhuwei},
			{Type: "财运_损", Evidence: "乙透干为偏财，财星为忌神，容易财来财去", Polarity: PolarityXiong, Source: SourceZhuwei},
		},
	}

	got := RenderYearNarrative(ys)
	if runeLen(got) < 120 {
		t.Fatalf("expected richer medium-detail narrative, got %d chars: %s", runeLen(got), got)
	}
	for _, want := range []string{"乙巳年", "工作", "钱财", "取舍"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected narrative to include %q, got: %s", want, got)
		}
	}
}

func TestRenderYearNarrative_WeakSignalYearDoesNotUseFiller(t *testing.T) {
	ys := YearSignals{
		Year:   2022,
		Age:    27,
		GanZhi: "壬寅",
	}

	got := RenderYearNarrative(ys)
	if runeLen(got) > 90 {
		t.Fatalf("weak-signal year should stay concise, got %d chars: %s", runeLen(got), got)
	}
	for _, filler := range []string{"年度定调", "触发来源", "现实落点", "十神力量"} {
		if strings.Contains(got, filler) {
			t.Fatalf("weak-signal narrative used structural filler %q: %s", filler, got)
		}
	}
}

func TestRenderYearNarrative_RichHardEvidenceLeadsWithPracticalMeaning(t *testing.T) {
	ys := YearSignals{
		Year:   2026,
		Age:    31,
		GanZhi: "丙午",
		TenGodPower: TenGodPowerProfile{
			PlainTitle: "钱财资源明显",
			PlainText:  "钱财、资源、合作回报和现实机会更容易被看见。",
		},
		Signals: []EventSignal{
			{Type: "健康", Evidence: "流年地支午冲日支子，日柱受冲，体力精神有下滑风险", Polarity: PolarityXiong, Source: SourceZhuwei},
			{Type: "财运_得", Evidence: "丙透干为正财，现实资源被带动", Polarity: PolarityJi, Source: SourceZhuwei},
		},
	}

	got := RenderYearNarrative(ys)
	opening := firstSentence(got)
	if !strings.Contains(opening, "健康") && !strings.Contains(opening, "身体") && !strings.Contains(opening, "出行安全") {
		t.Fatalf("hard evidence should lead with practical health meaning, got: %s", got)
	}
	if strings.Contains(opening, "钱财") || strings.Contains(opening, "资源") {
		t.Fatalf("money force should not lead hard-health year, got: %s", got)
	}
	if runeLen(got) < 120 {
		t.Fatalf("hard-evidence year should still be detailed, got %d chars: %s", runeLen(got), got)
	}
}

func TestRenderYearNarrative_DifferentSignalsDoNotReuseTenGodSentenceAsMainBody(t *testing.T) {
	power := TenGodPowerProfile{
		PlainTitle: "学习贵人明显",
		PlainText:  "学习、证书、师长贵人和保护性资源更容易出现。",
	}
	years := []YearSignals{
		{
			Year:        2024,
			Age:         29,
			GanZhi:      "甲辰",
			TenGodPower: power,
			Signals: []EventSignal{
				{Type: "婚恋_冲", Evidence: "流年地支辰冲日支戌，感情和家庭沟通受触动", Polarity: PolarityXiong, Source: SourceZhuwei},
			},
		},
		{
			Year:        2025,
			Age:         30,
			GanZhi:      "乙巳",
			TenGodPower: power,
			Signals: []EventSignal{
				{Type: "财运_损", Evidence: "乙透干为偏财，财星为忌神，容易财来财去", Polarity: PolarityXiong, Source: SourceZhuwei},
			},
		},
	}

	bodies := map[string]bool{}
	for _, ys := range years {
		got := RenderYearNarrative(ys)
		if strings.HasPrefix(strings.TrimPrefix(got, ys.GanZhi+"年，"), "学习贵人明显") {
			t.Fatalf("narrative should not rely on standalone ten-god sentence first, got: %s", got)
		}
		if bodies[got] {
			t.Fatalf("different signal years produced identical body: %s", got)
		}
		bodies[got] = true
	}
}

func TestRenderYearNarrative_RichYoungAgeUsesConcreteSchoolContext(t *testing.T) {
	ys := YearSignals{
		Year:   2011,
		Age:    16,
		GanZhi: "辛卯",
		TenGodPower: TenGodPowerProfile{
			PlainTitle: "学习贵人明显",
			PlainText:  "学习、证书、师长贵人和保护性资源更容易出现。",
		},
		Signals: []EventSignal{
			{Type: TypeXueYeGuiRen, Evidence: "辛透干为正印，少年期印星护身，得师长指点", Polarity: PolarityJi, Source: SourceZhuwei},
			{Type: TypeXingGePanNi, Evidence: "流年地支卯冲日支酉（自我宫位），少年期情绪波动", Polarity: PolarityXiong, Source: SourceZhuwei},
		},
	}

	got := RenderYearNarrative(ys)
	for _, bad := range []string{"事业", "财运", "婚恋", "职位"} {
		if strings.Contains(got, bad) {
			t.Fatalf("young-age rich narrative used adult wording %q: %s", bad, got)
		}
	}
	for _, want := range []string{"学习", "师长", "同学", "情绪"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected school-age detail %q, got: %s", want, got)
		}
	}
	// Length floor was 120 when age<18 narratives carried a templated closer.
	// After dropping the closer (it repeated identically across years for any
	// shared theme), child-age rich-signal years emit ~2-3 sentences. 70 is
	// the empirical lower bound for the domain+secondary pair this fixture
	// produces.
	if runeLen(got) < 70 {
		t.Fatalf("expected richer school-age narrative, got %d chars: %s", runeLen(got), got)
	}
}

func TestRenderYearNarrative_1996ChartRecentYearsStayDetailedAndDistinct(t *testing.T) {
	result := Calculate(1996, 2, 8, 20, "male", false, 0, "solar", false)
	years := GetAllYearSignals(result, "male", 2026, 0)
	targets := map[int]string{}
	for _, ys := range years {
		if ys.Year < 2024 || ys.Year > 2026 {
			continue
		}
		narrative := RenderYearNarrative(ys)
		if runeLen(narrative) < 100 {
			t.Fatalf("expected detailed narrative for %d, got %d chars: %s", ys.Year, runeLen(narrative), narrative)
		}
		targets[ys.Year] = narrative
	}

	if len(targets) != 3 {
		t.Fatalf("expected 2024-2026 narratives, got years: %#v", targets)
	}
	if targets[2024] == targets[2025] || targets[2025] == targets[2026] || targets[2024] == targets[2026] {
		t.Fatalf("recent year narratives should be distinct, got: %#v", targets)
	}
}

func TestRenderYearNarrative_1996RecentYearsAvoidRepeatedToneAndStance(t *testing.T) {
	result := Calculate(1996, 2, 8, 20, "male", false, 0, "solar", false)
	years := GetAllYearSignals(result, "male", 2026, 0)
	openings := map[string]int{}
	closings := map[string]int{}

	for _, ys := range years {
		if ys.Year < 2024 || ys.Year > 2026 {
			continue
		}
		narrative := RenderYearNarrative(ys)
		openings[firstSentence(narrative)]++
		closings[lastSentence(narrative)]++
	}

	for opening, count := range openings {
		if count > 1 {
			t.Fatalf("recent years repeated opening %q", opening)
		}
	}
	for closing, count := range closings {
		if count > 1 {
			t.Fatalf("recent years repeated closing %q", closing)
		}
	}
}

func firstSentence(s string) string {
	idx := strings.Index(s, "。")
	if idx < 0 {
		return s
	}
	return s[:idx+len("。")]
}

func lastSentence(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "。")
	idx := strings.LastIndex(s, "。")
	if idx < 0 {
		return s + "。"
	}
	return s[idx+len("。"):] + "。"
}

func runeLen(s string) int {
	return len([]rune(s))
}

func TestYearToneSentence_PolarityOnlyReturnsEmpty(t *testing.T) {
	cases := []struct {
		name    string
		signals []EventSignal
		primary EventSignal
	}{
		{
			name: "xiong>=2 ji>0 without hard primary",
			signals: []EventSignal{
				{Type: "综合变动", Polarity: PolarityXiong, Source: SourceZhuwei, Evidence: "节奏变化"},
				{Type: "综合变动", Polarity: PolarityXiong, Source: SourceZhuwei, Evidence: "节奏变化"},
				{Type: "综合变动", Polarity: PolarityJi, Source: SourceZhuwei, Evidence: "节奏变化"},
			},
			primary: EventSignal{Type: "综合变动", Polarity: PolarityXiong, Source: SourceZhuwei, Evidence: "节奏变化"},
		},
		{
			name: "all xiong without hard primary",
			signals: []EventSignal{
				{Type: "综合变动", Polarity: PolarityXiong, Source: SourceZhuwei, Evidence: "节奏变化"},
				{Type: "综合变动", Polarity: PolarityXiong, Source: SourceZhuwei, Evidence: "节奏变化"},
			},
			primary: EventSignal{Type: "综合变动", Polarity: PolarityXiong, Source: SourceZhuwei, Evidence: "节奏变化"},
		},
		{
			name: "all ji without hard primary",
			signals: []EventSignal{
				{Type: "综合变动", Polarity: PolarityJi, Source: SourceZhuwei, Evidence: "节奏变化"},
				{Type: "综合变动", Polarity: PolarityJi, Source: SourceZhuwei, Evidence: "节奏变化"},
			},
			primary: EventSignal{Type: "综合变动", Polarity: PolarityJi, Source: SourceZhuwei, Evidence: "节奏变化"},
		},
		{
			name:    "no signals",
			signals: nil,
			primary: EventSignal{Type: "综合变动", Polarity: PolarityNeutral, Source: SourceZhuwei, Evidence: "节奏变化"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := yearToneSentence(c.signals, c.primary); got != "" {
				t.Errorf("expected empty string, got %q", got)
			}
		})
	}
}

func TestYearToneSentence_HardSignalStillEmits(t *testing.T) {
	primary := EventSignal{
		Type:     "健康",
		Polarity: PolarityXiong,
		Source:   SourceZhuwei,
		Evidence: "流年地支午冲日支子，日柱受冲",
	}
	got := yearToneSentence([]EventSignal{primary}, primary)
	if got == "" {
		t.Fatal("expected non-empty hard-signal lead, got empty")
	}
}

func TestTriggerSourceSentence_NoKeywordReturnsEmpty(t *testing.T) {
	cases := []struct {
		name string
		sig  EventSignal
		age  int
	}{
		{
			name: "no keyword in evidence",
			sig:  EventSignal{Type: "综合变动", Evidence: "节奏一般变化", Source: SourceZhuwei},
			age:  30,
		},
		{
			name: "empty evidence and neutral type",
			sig:  EventSignal{Type: "综合变动", Evidence: "", Source: SourceZhuwei},
			age:  15,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := triggerSourceSentence(c.sig, c.age); got != "" {
				t.Errorf("expected empty string, got %q", got)
			}
		})
	}
}

func TestTriggerSourceSentence_KeywordStillEmits(t *testing.T) {
	cases := []EventSignal{
		{Type: "事业", Evidence: "流年地支辰冲月柱壬戌", Source: SourceZhuwei},
		{Type: "综合变动", Evidence: "落空亡，虚而不实", Source: SourceKongwang},
		{Type: "伏吟", Evidence: "流年壬辰伏吟日柱壬辰", Source: SourceFuyin},
	}
	for i, c := range cases {
		if got := triggerSourceSentence(c, 30); got == "" {
			t.Errorf("case %d: expected non-empty, got empty for %+v", i, c)
		}
	}
}

func TestDomainDetailSentence_UnknownThemeReturnsEmpty(t *testing.T) {
	primary := EventSignal{Type: "未知类型", Evidence: "无关键词", Source: SourceZhuwei}
	if got := domainDetailSentence(primary, EventSignal{}, false, 30); got != "" {
		t.Errorf("expected empty for unknown theme, got %q", got)
	}
}

func TestRichChangeSentence_NoAnchorReturnsEmpty(t *testing.T) {
	sig := EventSignal{Type: "综合变动", Evidence: "节奏一般变化", Source: SourceZhuwei, Polarity: PolarityNeutral}
	if got := richChangeSentence(sig); got != "" {
		t.Errorf("expected empty for un-anchored change signal, got %q", got)
	}
}

func TestRichStudySentence_UnknownStudyTypeReturnsEmpty(t *testing.T) {
	primary := EventSignal{Type: "事业", Evidence: "无关键词", Source: SourceZhuwei}
	if got := richStudySentence(primary, EventSignal{}, false); got != "" {
		t.Errorf("expected empty for unknown study type, got %q", got)
	}
}

func TestSecondaryDetailSentence_UnanchoredSignalReturnsEmpty(t *testing.T) {
	cases := []struct {
		name string
		sig  EventSignal
	}{
		{
			name: "vague 综合变动 with no keyword",
			sig:  EventSignal{Type: "综合变动", Evidence: "节奏一般变化", Source: SourceZhuwei, Polarity: PolarityNeutral},
		},
		{
			name: "vague 喜神临运 with no anchor keyword",
			sig:  EventSignal{Type: "喜神临运", Evidence: "印星生身", Source: SourceZhuwei, Polarity: PolarityJi},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := secondaryDetailSentence(c.sig, 30); got != "" {
				t.Errorf("expected empty, got %q", got)
			}
		})
	}
}

func TestSecondaryDetailSentence_AnchoredSignalStillEmits(t *testing.T) {
	cases := []EventSignal{
		// (a) hard event signal
		{Type: "健康", Evidence: "流年地支午冲日支子，日柱受冲", Source: SourceZhuwei, Polarity: PolarityXiong},
		// (b) evidence keyword
		{Type: "财运_得", Evidence: "财星为忌神，破耗", Source: SourceZhuwei, Polarity: PolarityXiong},
		// (c) signal type in allowed set
		{Type: "伏吟", Evidence: "伏吟", Source: SourceFuyin, Polarity: PolarityXiong},
		{Type: TypeXueYeYaLi, Evidence: "学业要求", Source: SourceZhuwei, Polarity: PolarityNeutral},
		{Type: "婚恋_冲", Evidence: "婚恋冲", Source: SourceZhuwei, Polarity: PolarityXiong},
	}
	for i, c := range cases {
		if got := secondaryDetailSentence(c, 30); got == "" {
			t.Errorf("case %d (Type=%s): expected non-empty, got empty", i, c.Type)
		}
	}
}

func TestSecondaryDetailSentence_UnanchoredChangeDoesNotEmitBarePrefix(t *testing.T) {
	// Regression: after Task 3 dropped richChangeSentence's default branch,
	// secondaryDetailSentence's "change" case used to produce "同时，"
	// (just the prefix). Must return "" instead so joinNarrativeParts can
	// drop it cleanly.
	sig := EventSignal{Type: "综合变动", Evidence: "财来财去为忌神", Source: SourceZhuwei, Polarity: PolarityXiong}
	got := secondaryDetailSentence(sig, 30)
	if got == "同时，" || got == "同时，。" {
		t.Fatalf("bare prefix emitted: %q", got)
	}
	// Allowed: either truly empty, or a fully-formed sentence — never just the prefix.
	if got != "" && !strings.Contains(got, "现实表现上") && !strings.Contains(got, "钱财") && !strings.Contains(got, "资源") && !strings.Contains(got, "情绪") && !strings.Contains(got, "出行") && !strings.Contains(got, "外部") && !strings.Contains(got, "作息") && !strings.Contains(got, "感情") && !strings.Contains(got, "工作") && !strings.Contains(got, "学习") && !strings.Contains(got, "同学") {
		t.Errorf("non-empty result lacks meaningful body: %q", got)
	}
}

func TestTenGodNarrativeSentence_NoGroupAlignmentReturnsEmpty(t *testing.T) {
	power := TenGodPowerProfile{
		PlainTitle: "官杀偏旺",
		PlainText:  "规则、责任和外部压力更明显",
		Group:      "", // no group → tenGodGroupTheme returns ""
	}
	primary := EventSignal{Type: "综合变动", Evidence: "节奏变化", Source: SourceZhuwei}
	if got := tenGodNarrativeSentence(power, primary, EventSignal{}, false); got != "" {
		t.Errorf("expected empty when 10-god group has no theme alignment, got %q", got)
	}
}

func TestTenGodNarrativeSentence_GroupAlignedStillEmits(t *testing.T) {
	power := TenGodPowerProfile{
		PlainTitle: "财星偏旺",
		PlainText:  "钱财、资源、合作回报更明显",
		Group:      TenGodGroupWealth, // wealth → money theme
	}
	primary := EventSignal{Type: "财运_得", Evidence: "财来财去", Source: SourceZhuwei}
	if got := tenGodNarrativeSentence(power, primary, EventSignal{}, false); got == "" {
		t.Error("expected non-empty when 10-god group aligns with primary theme")
	}
}

func TestPracticalStanceSentence_UnknownThemeReturnsEmpty(t *testing.T) {
	primary := EventSignal{Type: "未知类型", Polarity: PolarityXiong, Source: SourceZhuwei}
	if got := practicalStanceSentence([]EventSignal{primary}, primary, 30); got != "" {
		t.Errorf("expected empty for unknown theme, got %q", got)
	}
}

func TestRenderYearNarrative_HiddenWhenBelowThreshold(t *testing.T) {
	// Two signals, both un-anchored — every builder returns "".
	ys := YearSignals{
		Year:   2010,
		Age:    11,
		GanZhi: "庚寅",
		Signals: []EventSignal{
			{Type: "综合变动", Evidence: "节奏一般变化", Source: SourceZhuwei, Polarity: PolarityNeutral},
			{Type: "综合变动", Evidence: "另一个变化", Source: SourceZhuwei, Polarity: PolarityNeutral},
		},
	}
	if got := RenderYearNarrative(ys); got != "" {
		t.Errorf("expected empty narrative below threshold, got %q", got)
	}
}

func TestRenderYearNarrative_NoSignalsReturnsEmpty(t *testing.T) {
	// No meaningful signals at all — old code emitted a tengod context fallback
	// or "本年命理信号较弱" stub; new code returns "".
	ys := YearSignals{Year: 2022, Age: 27, GanZhi: "壬寅"}
	if got := RenderYearNarrative(ys); got != "" {
		t.Errorf("expected empty narrative for no-signals year, got %q", got)
	}
}

func TestRenderYearNarrative_MeetsThresholdWhenAnchored(t *testing.T) {
	// Hard health signal: yearTone (healthLead), trigger (冲), domain (health
	// with 冲 keyword), practical (health) — at least 3 anchored sentences.
	ys := YearSignals{
		Year:   2026,
		Age:    31,
		GanZhi: "丙午",
		Signals: []EventSignal{
			{Type: "健康", Evidence: "流年地支午冲日支子，日柱受冲，体力精神有下滑风险", Polarity: PolarityXiong, Source: SourceZhuwei},
		},
	}
	got := RenderYearNarrative(ys)
	if got == "" {
		t.Fatal("expected narrative to render for hard health signal year")
	}
	if !strings.HasPrefix(got, "丙午年，") {
		t.Errorf("expected narrative to start with GanZhi prefix, got: %s", got)
	}
}

func TestRenderYearNarrative_AdultSoftShenshaRenders(t *testing.T) {
	// Regression: user reported (2026-05-19) that adult years with multiple
	// shensha-derived signal chips (健康/婚恋_合/迁变/喜神临运) rendered as
	// empty cards. Root cause was an over-strict pre-flight gate that
	// required at least one signal to carry a structural evidence keyword
	// (冲/刑/空/etc.). Soft shensha signals like 天医/桃花/华盖/喜神临运
	// described real events but their Evidence strings lacked those keywords.
	// After dropping the pre-flight gate, such years should render via the
	// MinSentencesForNarrative=3 threshold path.
	ys := YearSignals{
		Year:   2022,
		Age:    28,
		GanZhi: "壬寅",
		Signals: []EventSignal{
			{Type: "健康", Evidence: "天医临运，主疾病减损、医疗顺遂", Polarity: PolarityJi, Source: SourceShensha},
			{Type: "婚恋_合", Evidence: "桃花临运，人缘异性缘旺", Polarity: PolarityNeutral, Source: SourceShensha},
			{Type: "迁变", Evidence: "华盖临运，主清高、宗教、艺术、孤独", Polarity: PolarityNeutral, Source: SourceShensha},
			{Type: "喜神临运", Evidence: "壬为调候喜神，全局运势有明显助力", Polarity: PolarityJi, Source: SourceYongshen},
		},
	}
	got := RenderYearNarrative(ys)
	if got == "" {
		t.Fatal("expected narrative to render for adult year with multiple shensha signals; got empty")
	}
	if !strings.HasPrefix(got, "壬寅年，") {
		t.Errorf("expected GanZhi prefix, got: %s", got)
	}
}

func TestRenderYearNarrative_ScreenshotRegression_RepetitiveOpenerHidden(t *testing.T) {
	// Reproduces the 2026-05-18 screenshot: three adjacent child-age years
	// (乙酉 2005 / 丙戌 2006 / 丁亥 2007) where the old template emitted
	// "这一年有机会也有压力，事情会同时出现可争取和需取舍的一面" as the
	// opener for ALL THREE. Under the evidence-anchored contract, none of
	// them carry hard signals — all three narratives should be hidden.
	years := []YearSignals{
		{
			Year:   2005,
			Age:    10,
			GanZhi: "乙酉",
			Signals: []EventSignal{
				{Type: "综合变动", Evidence: "流年地支酉与原局子时刑（空亡相邻）", Polarity: PolarityXiong, Source: SourceKongwang},
				{Type: TypeXueYeJingZheng, Evidence: "乙木为命主比劫，少年期同学比较增强", Polarity: PolarityNeutral, Source: SourceZhuwei},
				{Type: TypeXingGeQingYi, Evidence: "流年地支酉为桃花星临命", Polarity: PolarityNeutral, Source: SourceZhuwei},
			},
		},
		{
			Year:   2006,
			Age:    11,
			GanZhi: "丙戌",
			Signals: []EventSignal{
				{Type: "综合变动", Evidence: "流年节奏一般变化", Polarity: PolarityNeutral, Source: SourceZhuwei},
				{Type: TypeXueYeGuiRen, Evidence: "丙透干为食神，少年期表达能力突出", Polarity: PolarityJi, Source: SourceZhuwei},
				{Type: TypeXingGeQingYi, Evidence: "流年地支戌合卯木", Polarity: PolarityNeutral, Source: SourceZhuwei},
				{Type: "健康", Evidence: "流年节奏微调", Polarity: PolarityXiong, Source: SourceZhuwei},
			},
		},
		{
			Year:   2007,
			Age:    12,
			GanZhi: "丁亥",
			Signals: []EventSignal{
				{Type: "综合变动", Evidence: "流年节奏一般变化", Polarity: PolarityXiong, Source: SourceZhuwei},
				{Type: TypeXueYeGuiRen, Evidence: "丁透干印星，少年期师长缘", Polarity: PolarityJi, Source: SourceZhuwei},
				{Type: "健康", Evidence: "流年微调", Polarity: PolarityXiong, Source: SourceZhuwei},
			},
		},
	}

	bannedFillers := []string{
		"这一年有机会也有压力",
		"触发点来自这一年的主导信号",
		"这一年最要紧的",
		"本年命理信号较弱",
		"可作为理解这一年事件走向的背景力量",
	}

	openings := map[string]string{} // opening → ganzhi that emitted it
	for _, ys := range years {
		narrative := RenderYearNarrative(ys)
		for _, banned := range bannedFillers {
			if strings.Contains(narrative, banned) {
				t.Fatalf("%s narrative contains banned filler %q: %s", ys.GanZhi, banned, narrative)
			}
		}
		if narrative == "" {
			continue // hidden cards are fine — we want that
		}
		opening := firstSentence(narrative)
		if prev, seen := openings[opening]; seen {
			t.Fatalf("repeated opening %q across years %s and %s", opening, prev, ys.GanZhi)
		}
		openings[opening] = ys.GanZhi
	}
}

func TestRenderYearNarrativeWithFallback_NoSignalsReturnsNonEmpty(t *testing.T) {
	// 0 signals — RenderYearNarrative 返 ""，Fallback 必须兜底
	ys := YearSignals{Year: 2022, Age: 27, GanZhi: "壬寅", DayunGanZhi: "辛丑"}
	got := RenderYearNarrativeWithFallback(ys)
	if got == "" {
		t.Fatal("expected non-empty fallback for no-signals year")
	}
}

func TestRenderYearNarrativeWithFallback_AnchoredYearDelegatesToOriginal(t *testing.T) {
	// 有真实信号的年应原样返回 RenderYearNarrative 的输出（不调兜底）
	ys := YearSignals{
		Year:   2026,
		Age:    31,
		GanZhi: "丙午",
		Signals: []EventSignal{
			{Type: "健康", Evidence: "流年地支午冲日支子，日柱受冲，体力精神有下滑风险", Polarity: PolarityXiong, Source: SourceZhuwei},
		},
	}
	want := RenderYearNarrative(ys)
	if want == "" {
		t.Fatal("test fixture invalid: RenderYearNarrative returned empty for anchored year")
	}
	got := RenderYearNarrativeWithFallback(ys)
	if got != want {
		t.Errorf("expected wrapper to return original narrative; got %q want %q", got, want)
	}
}

func TestMakeMinimalFallback_PolarityJi(t *testing.T) {
	ys := YearSignals{Year: 2020, Age: 25, GanZhi: "庚子", DayunGanZhi: "甲寅",
		Signals: []EventSignal{{Type: "用神基底", Source: SourceYongshen, Polarity: PolarityJi}}}
	got := makeMinimalFallback(ys)
	if !strings.Contains(got, "向吉") {
		t.Errorf("ji polarity should mention '向吉'; got %q", got)
	}
	if !strings.Contains(got, "庚子") {
		t.Errorf("expected ganzhi in output; got %q", got)
	}
}

func TestMakeMinimalFallback_PolarityXiong(t *testing.T) {
	ys := YearSignals{Year: 2021, Age: 26, GanZhi: "辛丑", DayunGanZhi: "甲寅",
		Signals: []EventSignal{{Type: "用神基底", Source: SourceYongshen, Polarity: PolarityXiong}}}
	got := makeMinimalFallback(ys)
	if !strings.Contains(got, "偏凶") {
		t.Errorf("xiong polarity should mention '偏凶'; got %q", got)
	}
}

func TestMakeMinimalFallback_NeutralWithDayun(t *testing.T) {
	ys := YearSignals{Year: 2022, Age: 27, GanZhi: "壬寅", DayunGanZhi: "甲寅",
		Signals: []EventSignal{{Type: "用神基底", Source: SourceYongshen, Polarity: PolarityNeutral}}}
	got := makeMinimalFallback(ys)
	if !strings.Contains(got, "按本段大运甲寅方向延展") {
		t.Errorf("neutral polarity with dayun should reference dayun ganzhi; got %q", got)
	}
}

func TestMakeMinimalFallback_NoBasisNoDayun(t *testing.T) {
	// 没有 SourceYongshen 信号、也没有 DayunGanZhi
	ys := YearSignals{Year: 2023, Age: 28, GanZhi: "癸卯"}
	got := makeMinimalFallback(ys)
	if !strings.Contains(got, "无明显波动") {
		t.Errorf("absent dayun should fall to 无明显波动 phrasing; got %q", got)
	}
	if got == "" {
		t.Fatal("must not return empty")
	}
}

func TestMakeMinimalFallback_KeywordSafe(t *testing.T) {
	// 兜底文案的任何 polarity 分支都禁止触发 ValidateYearNarrative 的 28 个关键词。
	forbidden := []string{
		"用神位", "忌神位", "喜神位",
		"伏吟", "反吟", "大运合化", "三会", "三合",
		"受冲", "受刑", "双重命中", "力度倍增",
		"驿马", "桃花", "华盖", "白虎", "丧门", "吊客", "灾煞", "流霞",
		"天医", "天喜", "天乙", "天德", "月德", "文昌", "太极", "福星",
		"红艳", "孤辰", "寡宿", "羊刃", "亡神", "劫煞", "披麻", "咸池",
		"勾绞", "国印",
	}
	cases := []YearSignals{
		{Year: 2020, Age: 25, GanZhi: "庚子", DayunGanZhi: "甲寅",
			Signals: []EventSignal{{Type: "用神基底", Source: SourceYongshen, Polarity: PolarityJi}}},
		{Year: 2021, Age: 26, GanZhi: "辛丑", DayunGanZhi: "甲寅",
			Signals: []EventSignal{{Type: "用神基底", Source: SourceYongshen, Polarity: PolarityXiong}}},
		{Year: 2022, Age: 27, GanZhi: "壬寅", DayunGanZhi: "甲寅",
			Signals: []EventSignal{{Type: "用神基底", Source: SourceYongshen, Polarity: PolarityNeutral}}},
		{Year: 2023, Age: 28, GanZhi: "癸卯"},
	}
	for _, ys := range cases {
		got := makeMinimalFallback(ys)
		for _, kw := range forbidden {
			if strings.Contains(got, kw) {
				t.Errorf("fallback for %s contains forbidden keyword %q: %q", ys.GanZhi, kw, got)
			}
		}
	}
}
