package bazi

import (
	"strings"
	"testing"
)

func TestTenGodGroupOfMapsExactTenGods(t *testing.T) {
	tests := []struct {
		name      string
		tenGod    string
		wantGroup string
		wantLabel string
	}{
		{name: "wealth", tenGod: "正财", wantGroup: TenGodGroupWealth, wantLabel: "财星"},
		{name: "official", tenGod: "七杀", wantGroup: TenGodGroupOfficial, wantLabel: "官杀"},
		{name: "seal", tenGod: "正印", wantGroup: TenGodGroupSeal, wantLabel: "印星"},
		{name: "output", tenGod: "伤官", wantGroup: TenGodGroupOutput, wantLabel: "食伤"},
		{name: "peer", tenGod: "劫财", wantGroup: TenGodGroupPeer, wantLabel: "比劫"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := TenGodGroupOf(tt.tenGod)
			if !ok {
				t.Fatalf("expected %s to map to a group", tt.tenGod)
			}
			if got.Group != tt.wantGroup || got.Label != tt.wantLabel {
				t.Fatalf("unexpected group for %s: %+v", tt.tenGod, got)
			}
		})
	}
}

func TestBuildYearTenGodPowerReinforcesSameDayunGroup(t *testing.T) {
	natal := makeNatal("甲子", "丁卯", "甲午", "戊子", "木火", "金水土")
	dayun := DayunItem{
		Gan:        "庚",
		Zhi:        "申",
		GanShiShen: GetShiShen(natal.DayGan, "庚"),
		ZhiShiShen: GetZhiShiShen(natal.DayGan, "申"),
	}
	liunian := LiuNianItem{
		GanZhi:     "辛酉",
		GanShiShen: GetShiShen(natal.DayGan, "辛"),
		ZhiShiShen: GetZhiShiShen(natal.DayGan, "酉"),
	}
	dayunPower := BuildDayunTenGodPower(natal, dayun)

	got := BuildYearTenGodPower(natal, dayun, liunian, YearSignalContext{DayunPhase: DayunPhaseGan}, dayunPower)

	if got.Group != TenGodGroupOfficial {
		t.Fatalf("expected official force group, got %+v", got)
	}
	if got.Score < 8 {
		t.Fatalf("expected same-group reinforcement to raise score, got %+v", got)
	}
	if !strings.Contains(got.Reason, "大运同类") {
		t.Fatalf("expected reinforcement reason, got %q", got.Reason)
	}
	if got.PlainTitle == "" || got.PlainText == "" {
		t.Fatalf("expected plain-language copy, got %+v", got)
	}
}

func TestBuildYearTenGodPowerPrefersLiunianForceOverDayunBackgroundTie(t *testing.T) {
	natal := makeNatal("甲子", "丁卯", "甲午", "戊子", "木火", "金水土")
	dayun := DayunItem{
		Gan:        "壬",
		Zhi:        "子",
		GanShiShen: GetShiShen(natal.DayGan, "壬"),
		ZhiShiShen: GetZhiShiShen(natal.DayGan, "子"),
	}
	liunian := LiuNianItem{
		GanZhi:     "甲辰",
		GanShiShen: GetShiShen(natal.DayGan, "甲"),
		ZhiShiShen: GetZhiShiShen(natal.DayGan, "辰"),
	}
	dayunPower := BuildDayunTenGodPower(natal, dayun)

	got := BuildYearTenGodPower(natal, dayun, liunian, YearSignalContext{DayunPhase: DayunPhaseGan}, dayunPower)

	if got.Group == TenGodGroupSeal {
		t.Fatalf("dayun background should not become the yearly dominant force by itself, got %+v", got)
	}
	if got.Group != TenGodGroupPeer {
		t.Fatalf("expected liunian stem force to stay dominant, got %+v", got)
	}
}

func TestTenGodPowerPlainTitleUsesEverydayLanguage(t *testing.T) {
	natal := makeNatal("甲子", "丁卯", "甲午", "戊子", "木火", "金水土")
	dayun := DayunItem{
		Gan:        "壬",
		Zhi:        "子",
		GanShiShen: GetShiShen(natal.DayGan, "壬"),
		ZhiShiShen: GetZhiShiShen(natal.DayGan, "子"),
	}

	got := BuildDayunTenGodPower(natal, dayun)

	for _, jargon := range []string{"印星", "财星", "官杀", "食伤", "比劫"} {
		if strings.Contains(got.PlainTitle, jargon) {
			t.Fatalf("plain title should avoid jargon %q, got %+v", jargon, got)
		}
	}
	if !strings.Contains(got.PlainTitle, "学习") && !strings.Contains(got.PlainTitle, "贵人") {
		t.Fatalf("expected seal force title to use everyday words, got %+v", got)
	}
}
