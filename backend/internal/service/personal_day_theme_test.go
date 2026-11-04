package service

import (
	"encoding/json"
	"strings"
	"testing"
	"yuanju/pkg/bazi"
)

func themeByID(items []PersonalDayGuidance, id string) *PersonalDayGuidance {
	for i := range items {
		if items[i].ID == id {
			return &items[i]
		}
	}
	return nil
}

func TestPersonalDayThemesSpecificSignalsAndBoundaries(t *testing.T) {
	chart := &bazi.BaziResult{YearGan: "乙", DayGan: "乙", YearZhi: "子", DayZhi: "卯"}
	ctx := personalDayActivityContext{
		dayGan: "辛", dayZhi: "酉", valid: true, clash: true,
		ganGod: "正财", ganElement: "金", ganDirection: "adverse",
		zhiDirection: "neutral", conflictElements: map[string]bool{},
	}
	items := buildPersonalDayThemes(chart, ctx)
	if len(items) != 4 || items[0].ID != "noble" || items[1].ID != "finance" || items[2].ID != "romance" || items[3].ID != "wellbeing" {
		t.Fatalf("want four ordered themes; got %+v", items)
	}
	finance, romance, health, noble := themeByID(items, "finance"), themeByID(items, "romance"), themeByID(items, "wellbeing"), themeByID(items, "noble")
	if finance == nil || finance.Label != "慎" || romance == nil || romance.Label != "慎" || health == nil || health.Label != "平" || noble == nil || noble.Label != "平" {
		t.Fatalf("unexpected theme levels: %+v", items)
	}
	if strings.Contains(strings.Join(health.Evidence, ""), "相冲") || strings.Contains(health.Summary+strings.Join(health.Actions, ""), "路口") {
		t.Fatalf("clash-only health guidance must not reuse travel content: %+v", health)
	}
	if !strings.Contains(strings.Join(romance.Evidence, ""), "桃花") || !strings.Contains(strings.Join(romance.Evidence, ""), "相冲") {
		t.Fatalf("romance must retain both signals: %+v", romance)
	}
	for _, item := range items {
		if len(item.Evidence) == 0 || len(item.Actions) == 0 || item.SourceLabel != "本人命盘" {
			t.Fatalf("theme lacks evidence or action: %+v", item)
		}
		copy := item.Summary + strings.Join(item.Actions, "")
		for _, forbidden := range []string{"买入", "卖出", "收益", "疾病", "意外", "车祸"} {
			if strings.Contains(copy, forbidden) {
				t.Fatalf("forbidden prediction in %+v", item)
			}
		}
	}
}

func TestPersonalDayHealthOnlyReviewsDoubleAdverse(t *testing.T) {
	chart := &bazi.BaziResult{YearGan: "甲", DayGan: "甲", YearZhi: "寅", DayZhi: "子"}
	ctx := personalDayActivityContext{valid: true, dayZhi: "午", clash: true, ganDirection: "favorable", zhiDirection: "adverse"}
	if item := themeByID(buildPersonalDayThemes(chart, ctx), "wellbeing"); item == nil || item.Label != "平" {
		t.Fatalf("day-branch clash alone must not mark health as caution: %+v", item)
	}
	ctx.clash = false
	ctx.ganDirection = "adverse"
	if item := themeByID(buildPersonalDayThemes(chart, ctx), "wellbeing"); item == nil || item.Label != "慎" || !strings.Contains(strings.Join(item.Evidence, ""), "忌神") || strings.Contains(item.Summary+strings.Join(item.Actions, ""), "过路口") {
		t.Fatalf("double adverse health reminder must stay about rest and activity: %+v", item)
	}
}

func TestPersonalDayThemesNobleAndNoSignal(t *testing.T) {
	chart := &bazi.BaziResult{YearGan: "甲", DayGan: "甲", YearZhi: "寅", DayZhi: "寅"}
	base := personalDayActivityContext{valid: true, dayZhi: "未", ganDirection: "neutral", zhiDirection: "neutral"}
	got := buildPersonalDayThemes(chart, base)
	if len(got) != 4 || got[0].ID != "noble" || got[0].Label != "宜" || !strings.Contains(strings.Join(got[0].Evidence, ""), "天乙贵人") {
		t.Fatalf("unexpected noble theme: %+v", got)
	}
	base.zhiDirection = "adverse"
	got = buildPersonalDayThemes(chart, base)
	if noble := themeByID(got, "noble"); noble == nil || noble.Label != "慎" {
		t.Fatalf("conflicting branch direction must downgrade noble reference: %+v", got)
	}
	base.zhiDirection = "neutral"
	base.dayZhi = "子"
	if got := buildPersonalDayThemes(chart, base); len(got) != 4 || got[0].Label != "平" || got[1].Label != "平" || got[2].Label != "平" || got[3].Label != "平" {
		t.Fatalf("no special signal should return four checked neutral themes: %+v", got)
	}
	base.valid = false
	if got := buildPersonalDayThemes(chart, base); len(got) != 0 {
		t.Fatalf("invalid chart should return no theme: %+v", got)
	}
}

func TestPersonalDayThemesLiuHeIsSocialReference(t *testing.T) {
	chart := &bazi.BaziResult{YearGan: "乙", DayGan: "乙", YearZhi: "寅", DayZhi: "子"}
	ctx := personalDayActivityContext{valid: true, dayZhi: "丑", ganDirection: "neutral", zhiDirection: "neutral"}
	got := buildPersonalDayThemes(chart, ctx)
	if len(got) != 4 || got[2].ID != "romance" || !strings.Contains(strings.Join(got[2].Evidence, ""), "六合") {
		t.Fatalf("day-branch liuhe should be a social reference: %+v", got)
	}
}

func TestPersonalDayThemesIncludedInDailyResult(t *testing.T) {
	chart := dailyChart("丙", "子")
	chart.YearGan, chart.YearZhi = "甲", "申"
	day, err := BuildPersonalDay("2026-11-02", chart)
	if err != nil || len(day.ThemeGuidance) != 4 || themeByID(day.ThemeGuidance, "finance") == nil {
		t.Fatalf("daily result should include new themes: %+v %v", day, err)
	}
	encoded, err := json.Marshal(day)
	if err != nil || !strings.Contains(string(encoded), `"theme_guidance"`) || !strings.Contains(string(encoded), `"activity_guidance"`) {
		t.Fatalf("new and old fields must coexist: %s %v", encoded, err)
	}
}
