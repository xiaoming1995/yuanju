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
