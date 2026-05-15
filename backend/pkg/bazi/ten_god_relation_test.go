package bazi

import (
	"strings"
	"testing"
)

func TestCalculateBuildsDayMasterTenGodRelation(t *testing.T) {
	result := Calculate(1996, 2, 8, 20, "male", false, 0, "solar", false)
	if result.TenGodRelation == nil {
		t.Fatal("TenGodRelation should be populated")
	}

	relation := result.TenGodRelation
	if relation.DayMaster.Gan != result.DayGan {
		t.Fatalf("day master gan = %q, want %q", relation.DayMaster.Gan, result.DayGan)
	}
	if relation.DayMaster.Label != result.DayGan+result.DayGanWuxing {
		t.Fatalf("day master label = %q, want %q", relation.DayMaster.Label, result.DayGan+result.DayGanWuxing)
	}
	if len(relation.HeavenlyStems) != 4 {
		t.Fatalf("heavenly stem relations = %d, want 4", len(relation.HeavenlyStems))
	}

	dayStem := findStemRelation(t, relation.HeavenlyStems, "day")
	if dayStem.TenGod != "日主 / 日元" {
		t.Fatalf("day stem ten god = %q, want 日主 / 日元", dayStem.TenGod)
	}
	if dayStem.Relation != "命主自身" {
		t.Fatalf("day stem relation = %q, want 命主自身", dayStem.Relation)
	}
	if !strings.Contains(dayStem.Summary, "参照点") {
		t.Fatalf("day stem summary should explain reference point, got %q", dayStem.Summary)
	}

	yearStem := findStemRelation(t, relation.HeavenlyStems, "year")
	if yearStem.TenGod != result.YearGanShiShen {
		t.Fatalf("year stem ten god = %q, want %q", yearStem.TenGod, result.YearGanShiShen)
	}
	if yearStem.Relation == "" || yearStem.Summary == "" || yearStem.GroupLabel == "" {
		t.Fatalf("year stem should include relation, summary, and group label: %+v", yearStem)
	}
}

func TestTenGodRelationPairsHiddenStemsWithTheirOwnTenGods(t *testing.T) {
	result := Calculate(1996, 2, 8, 20, "male", false, 0, "solar", false)
	if result.TenGodRelation == nil {
		t.Fatal("TenGodRelation should be populated")
	}

	for _, group := range result.TenGodRelation.HiddenStems {
		if len(group.Items) == 0 {
			t.Fatalf("%s should include hidden stem relation items", group.PillarLabel)
		}
		for _, item := range group.Items {
			want := GetShiShen(result.DayGan, item.Gan)
			if item.TenGod != want {
				t.Fatalf("%s hidden stem %s ten god = %q, want %q", group.PillarLabel, item.Gan, item.TenGod, want)
			}
			if item.Relation == "" || item.Summary == "" {
				t.Fatalf("%s hidden stem %s should include relation and summary: %+v", group.PillarLabel, item.Gan, item)
			}
		}
	}
}

func TestEnsureTenGodRelationDerivesLegacySnapshots(t *testing.T) {
	legacy := &BaziResult{
		YearGan:           "壬",
		YearZhi:           "申",
		MonthGan:          "甲",
		MonthZhi:          "辰",
		DayGan:            "丙",
		DayZhi:            "戌",
		HourGan:           "戊",
		HourZhi:           "子",
		YearGanWuxing:     "水",
		MonthGanWuxing:    "木",
		DayGanWuxing:      "火",
		HourGanWuxing:     "土",
		YearGanShiShen:    "七杀",
		MonthGanShiShen:   "偏印",
		DayGanShiShen:     "比肩",
		HourGanShiShen:    "食神",
		YearHideGan:       []string{"庚", "壬", "戊"},
		YearZhiShiShen:    []string{"偏财"},
		MonthHideGan:      []string{"戊", "乙", "癸"},
		MonthZhiShiShen:   []string{"食神", "正印", "正官"},
		DayHideGan:        []string{"戊", "辛", "丁"},
		DayZhiShiShen:     []string{"食神", "正财", "劫财"},
		HourHideGan:       []string{"癸"},
		HourZhiShiShen:    []string{"正官"},
		TenGodRelation:    nil,
	}

	EnsureTenGodRelation(legacy)
	if legacy.TenGodRelation == nil {
		t.Fatal("legacy snapshot should derive TenGodRelation")
	}
	yearHidden := findHiddenGroup(t, legacy.TenGodRelation.HiddenStems, "year")
	if len(yearHidden.Items) != 3 {
		t.Fatalf("year hidden items = %d, want 3 derived from hidden stems", len(yearHidden.Items))
	}
	if yearHidden.Items[1].Gan != "壬" || yearHidden.Items[1].TenGod != "七杀" {
		t.Fatalf("year hidden second item should derive its own ten god, got %+v", yearHidden.Items[1])
	}
}

func findStemRelation(t *testing.T, items []TenGodStemRelation, pillar string) TenGodStemRelation {
	t.Helper()
	for _, item := range items {
		if item.Pillar == pillar {
			return item
		}
	}
	t.Fatalf("missing %s stem relation in %+v", pillar, items)
	return TenGodStemRelation{}
}

func findHiddenGroup(t *testing.T, items []TenGodHiddenStemGroup, pillar string) TenGodHiddenStemGroup {
	t.Helper()
	for _, item := range items {
		if item.Pillar == pillar {
			return item
		}
	}
	t.Fatalf("missing %s hidden group in %+v", pillar, items)
	return TenGodHiddenStemGroup{}
}
