package bazi

import (
	"reflect"
	"testing"
)

func TestBuildGongJiaAdjacentSameGanGap(t *testing.T) {
	result := &BaziResult{
		YearGan:  "甲",
		YearZhi:  "子",
		MonthGan: "甲",
		MonthZhi: "寅",
		DayGan:   "丙",
		DayZhi:   "午",
		HourGan:  "戊",
		HourZhi:  "申",
	}

	items := BuildGongJia(result)

	if len(items) != 1 {
		t.Fatalf("BuildGongJia() len = %d, want 1: %#v", len(items), items)
	}
	got := items[0]
	if got.Source != "year_month" {
		t.Errorf("Source = %q, want year_month", got.Source)
	}
	if got.SameGan != "甲" {
		t.Errorf("SameGan = %q, want 甲", got.SameGan)
	}
	if got.VirtualZhi != "丑" {
		t.Errorf("VirtualZhi = %q, want 丑", got.VirtualZhi)
	}
	if !reflect.DeepEqual(got.HideGan, []string{"己", "癸", "辛"}) {
		t.Errorf("HideGan = %#v, want [己 癸 辛]", got.HideGan)
	}
}

func TestBuildGongJiaReverseBranchOrder(t *testing.T) {
	result := &BaziResult{
		YearGan:  "甲",
		YearZhi:  "寅",
		MonthGan: "甲",
		MonthZhi: "子",
		DayGan:   "丙",
		DayZhi:   "午",
		HourGan:  "戊",
		HourZhi:  "申",
	}

	items := BuildGongJia(result)

	if len(items) != 1 {
		t.Fatalf("BuildGongJia() len = %d, want 1: %#v", len(items), items)
	}
	if items[0].VirtualZhi != "丑" {
		t.Errorf("VirtualZhi = %q, want 丑", items[0].VirtualZhi)
	}
}

func TestBuildGongJiaAdjacentOnly(t *testing.T) {
	result := &BaziResult{
		YearGan:  "甲",
		YearZhi:  "子",
		MonthGan: "乙",
		MonthZhi: "卯",
		DayGan:   "甲",
		DayZhi:   "寅",
		HourGan:  "戊",
		HourZhi:  "申",
	}

	items := BuildGongJia(result)

	if len(items) != 0 {
		t.Fatalf("BuildGongJia() len = %d, want 0: %#v", len(items), items)
	}
}

func TestBuildGongJiaRejectsDifferentGanOrNonGapBranches(t *testing.T) {
	tests := []struct {
		name   string
		result *BaziResult
	}{
		{
			name: "different stems",
			result: &BaziResult{
				YearGan:  "甲",
				YearZhi:  "子",
				MonthGan: "乙",
				MonthZhi: "寅",
				DayGan:   "丙",
				DayZhi:   "午",
				HourGan:  "戊",
				HourZhi:  "申",
			},
		},
		{
			name: "non gap branches",
			result: &BaziResult{
				YearGan:  "甲",
				YearZhi:  "子",
				MonthGan: "甲",
				MonthZhi: "卯",
				DayGan:   "丙",
				DayZhi:   "午",
				HourGan:  "戊",
				HourZhi:  "申",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items := BuildGongJia(tt.result)
			if len(items) != 0 {
				t.Fatalf("BuildGongJia() len = %d, want 0: %#v", len(items), items)
			}
		})
	}
}
