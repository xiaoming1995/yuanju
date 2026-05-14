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

func TestRenderYearNarrative_TenGodPowerEnrichesGenericYear(t *testing.T) {
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

	got := RenderYearNarrative(ys)
	if !strings.Contains(got, "官杀偏旺") || !strings.Contains(got, "规则") {
		t.Fatalf("expected ten-god force to enrich generic year, got: %s", got)
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
	if runeLen(got) < 120 {
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
