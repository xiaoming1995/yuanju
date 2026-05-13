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

func TestRenderYearNarrative_AdverseDayunPhaseCanBeSecondaryCaution(t *testing.T) {
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
	if !strings.Contains(got, "底色") && !strings.Contains(got, "保守") && !strings.Contains(got, "压力") {
		t.Fatalf("expected adverse phase caution as secondary context, got: %s", got)
	}
	if strings.Contains(got, "金不换") || strings.Contains(got, "后5年") {
		t.Fatalf("narrative leaked technical phase evidence: %s", got)
	}
}

func TestRenderYearNarrative_AlwaysShowsDayunPhaseWhenPresent(t *testing.T) {
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
	if !strings.Contains(got, "后五年") || !strings.Contains(got, "地支") {
		t.Fatalf("expected visible late dayun phase wording, got: %s", got)
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

func firstSentence(s string) string {
	idx := strings.Index(s, "。")
	if idx < 0 {
		return s
	}
	return s[:idx+len("。")]
}
