package service

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"yuanju/pkg/bazi"
)

func dailyChart(dayGan, dayZhi string) *bazi.BaziResult {
	return &bazi.BaziResult{
		DayGan: dayGan, DayZhi: dayZhi,
		NatalAssessment: &bazi.NatalAssessment{Fuyi: bazi.NatalFuyiAssessment{Yongshen: "水金", Jishen: "火"}},
	}
}

func guidanceByID(day *PersonalDayResult, id string) *PersonalDayGuidance {
	for i := range day.ActivityGuidance {
		if day.ActivityGuidance[i].ID == id {
			return &day.ActivityGuidance[i]
		}
	}
	return nil
}

func TestPersonalDayDirectionsAndSpecificGuidance(t *testing.T) {
	chart := dailyChart("丙", "子")
	chart.NatalAssessment.Fuyi = bazi.NatalFuyiAssessment{Yongshen: "水金", Jishen: "火木"}
	day, err := BuildPersonalDay("2026-11-02", chart)
	if err != nil {
		t.Fatal(err)
	}
	if day.DayGanZhi != "庚辰" || day.Status != "undetermined" || day.Label != "日干偏合" || !strings.Contains(day.Summary, "庚（金）属喜用") || !strings.Contains(day.Summary, "辰（土主气）未落入喜忌") || day.StemElement != "金" || day.StemDirection != "favorable" || day.BranchElement != "土" || day.BranchDirection != "neutral" {
		t.Fatalf("unexpected directional result: %+v", day)
	}
	if got := strings.Join(day.CalendarYi, "、"); got != "祭祀、沐浴、解除、破屋、坏垣、馀事勿取" {
		t.Fatalf("full calendar Yi missing: %s", got)
	}
	if got := strings.Join(day.CalendarJi, "、"); got != "开市、嫁娶" {
		t.Fatalf("full calendar Ji missing: %s", got)
	}
	travel := guidanceByID(day, "travel")
	if travel == nil || travel.Level != "insufficient" || travel.Label != "未列" || travel.SourceLabel != "通用日历" || strings.Contains(travel.Summary, "有利") {
		t.Fatalf("unexpected travel guidance: %+v", travel)
	}
	finance := guidanceByID(day, "finance")
	if finance == nil || finance.Level != "candidate" || finance.Label != "宜" || finance.SourceLabel != "本人命盘" || len(finance.Evidence) == 0 {
		t.Fatalf("missing finance evidence: %+v", finance)
	}
	if work, communication := guidanceByID(day, "work"), guidanceByID(day, "communication"); work == nil || work.Level != "insufficient" || work.Label != "未列" || communication == nil || communication.Level != "neutral" || communication.Label != "平" {
		t.Fatalf("traditional labels must distinguish unlisted from neutral: work=%+v communication=%+v", work, communication)
	}
	for _, item := range day.ActivityGuidance {
		if strings.Contains(item.Summary+strings.Join(item.Evidence, ""), "适合买入") || strings.Contains(item.Summary, "稳赚") || strings.Contains(item.Summary, "建仓") {
			t.Fatalf("guidance must not provide a trading signal: %+v", item)
		}
	}
	encoded, err := json.Marshal(day)
	if err != nil || !strings.Contains(string(encoded), `"stem_direction":"favorable"`) || !strings.Contains(string(encoded), `"activity_guidance"`) {
		t.Fatalf("new API fields missing: %s %v", encoded, err)
	}
}

func TestPersonalDayGuidanceKeepsCalendarAndPersonalSourcesSeparate(t *testing.T) {
	chart := dailyChart("甲", "子")
	chart.NatalAssessment.Fuyi = bazi.NatalFuyiAssessment{Yongshen: "木火", Jishen: "金"}
	day, err := BuildPersonalDay("2026-11-16", chart)
	if err != nil {
		t.Fatal(err)
	}
	travel := guidanceByID(day, "travel")
	if day.Status != "caution" || !strings.Contains(day.Summary, "相冲") || travel == nil || travel.Level != "review" || len(travel.Sources) != 2 || !strings.Contains(travel.Summary, "宜出行") || !strings.Contains(travel.Summary, "相冲") {
		t.Fatalf("calendar Yi and personal clash should remain separate: %+v %+v", day, travel)
	}
	noClash := dailyChart("甲", "寅")
	noClash.NatalAssessment.Fuyi = chart.NatalAssessment.Fuyi
	yiDay, err := BuildPersonalDay("2026-11-16", noClash)
	if err != nil {
		t.Fatal(err)
	}
	yiTravel := guidanceByID(yiDay, "travel")
	if yiTravel == nil || yiTravel.Level != "candidate" || yiTravel.Label != "宜" || len(yiTravel.Sources) != 1 || !strings.Contains(yiTravel.Summary, "仅为通用日历") {
		t.Fatalf("generic Yi must not become a personal benefit claim: %+v", yiTravel)
	}
	jiDay, err := BuildPersonalDay("2026-01-07", noClash)
	if err != nil {
		t.Fatal(err)
	}
	jiTravel := guidanceByID(jiDay, "travel")
	if jiTravel == nil || jiTravel.Level != "review" || jiTravel.Label != "忌" || !strings.Contains(strings.Join(jiTravel.Evidence, ""), "忌出行") {
		t.Fatalf("missing generic Ji: %+v", jiTravel)
	}
	outside, err := BuildPersonalDay("2028-01-01", noClash)
	if err != nil || len(outside.ActivityGuidance) != 4 {
		t.Fatalf("real guidance must not stop at bundled calendar years: %+v %v", outside, err)
	}
}

func TestPersonalDayStatusCasesAndMissingChart(t *testing.T) {
	tests := []struct{ name, favorable, adverse, chartBranch, thermal, want, label string }{
		{"both favorable", "金土", "火", "子", "", "supportive", "偏合"},
		{"both adverse", "水", "金土", "子", "", "caution", "需留意"},
		{"opposite", "金", "土", "子", "", "mixed", "关系复杂"},
		{"stem favorable branch neutral", "金", "火", "子", "", "undetermined", "日干偏合"},
		{"branch favorable stem neutral", "土", "火", "子", "", "undetermined", "日支偏合"},
		{"stem adverse branch neutral", "水", "金", "子", "", "undetermined", "日干需留意"},
		{"branch adverse stem neutral", "水", "土", "子", "", "undetermined", "日支需留意"},
		{"both neutral", "水", "火", "子", "", "undetermined", "均无明显喜忌"},
		{"climate conflict", "水", "金", "子", "金", "mixed", "关系复杂"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			chart := dailyChart("丙", tc.chartBranch)
			chart.NatalAssessment.Fuyi = bazi.NatalFuyiAssessment{Yongshen: tc.favorable, Jishen: tc.adverse}
			chart.NatalAssessment.Tiaohou.Thermal.RequiredElements = tc.thermal
			day, err := BuildPersonalDay("2026-11-02", chart)
			if err != nil || day.Status != tc.want || day.Label != tc.label {
				t.Fatalf("unexpected status: %+v %v", day, err)
			}
		})
	}
	missing, err := BuildPersonalDay("2026-11-02", &bazi.BaziResult{DayGan: "丙", DayZhi: "子"})
	if err != nil || missing.Status != "insufficient" || missing.StemDirection != "" || guidanceByID(missing, "finance").Level != "insufficient" {
		t.Fatalf("missing chart should not get personal guidance: %+v %v", missing, err)
	}
}

func TestPersonalDayMatchesStaticCalendarAndDistinguishesTenGods(t *testing.T) {
	metal, err := BuildPersonalDay("2026-10-22", dailyChart("庚", "子"))
	if err != nil {
		t.Fatal(err)
	}
	if metal.DayGanZhi != "己巳" || metal.YearGanZhi != "丙午" || metal.MonthGanZhi != "戊戌" || metal.LunarFull != "九月十三" {
		t.Fatalf("date labels differ from static calendar: %+v", metal)
	}
	fireMetal, err := BuildPersonalDay("2026-10-23", dailyChart("庚", "子"))
	if err != nil {
		t.Fatal(err)
	}
	fireWater, err := BuildPersonalDay("2026-10-23", dailyChart("壬", "子"))
	if err != nil {
		t.Fatal(err)
	}
	if fireMetal.DayGanZhi != fireWater.DayGanZhi || strings.Join(fireMetal.Evidence, " ") == strings.Join(fireWater.Evidence, " ") {
		t.Fatalf("different day masters must have different evidence: %+v %+v", fireMetal, fireWater)
	}
	if strings.Contains(strings.Join(fireMetal.Evidence, " ")+fireMetal.Summary, "破财") || strings.Contains(strings.Join(fireWater.Evidence, " ")+fireWater.Summary, "破财") {
		t.Fatal("daily result must not claim certain financial loss")
	}
}

func TestDoubleFireDayDoesNotBecomeCertainFinancialLoss(t *testing.T) {
	metal, err := BuildPersonalDay("2025-02-06", dailyChart("庚", "辰"))
	if err != nil {
		t.Fatal(err)
	}
	water, err := BuildPersonalDay("2025-02-06", dailyChart("壬", "辰"))
	if err != nil {
		t.Fatal(err)
	}
	if metal.DayGanZhi != "丙午" || metal.Status != "caution" || water.Status != "caution" {
		t.Fatalf("unexpected double-fire direction: %+v %+v", metal, water)
	}
	for _, activity := range metal.Activities {
		if activity.ID == "finance" {
			t.Fatalf("metal day master fire should not be financial domain: %+v", metal.Activities)
		}
	}
	foundFinance := false
	for _, activity := range water.Activities {
		if activity.ID == "finance" {
			foundFinance = true
		}
	}
	if !foundFinance {
		t.Fatalf("water day master fire should have financial context: %+v", water.Activities)
	}
	if strings.Contains(water.Summary, "破财") || strings.Contains(metal.Summary, "破财") {
		t.Fatal("certain event claim")
	}
}

func TestPersonalDayInsufficientAndInvalidDate(t *testing.T) {
	day, err := BuildPersonalDay("2028-01-01", &bazi.BaziResult{DayGan: "庚", DayZhi: "子"})
	if err != nil || day.Status != "insufficient" || len(day.Activities) != 0 {
		t.Fatalf("unexpected missing-data result: %+v %v", day, err)
	}
	for _, date := range []string{"2026-02-30", "1899-12-31", "2101-01-01", ""} {
		if _, err := BuildPersonalDay(date, dailyChart("庚", "子")); !errors.Is(err, ErrInvalidPersonalDayDate) {
			t.Fatalf("%q: %v", date, err)
		}
	}
}
