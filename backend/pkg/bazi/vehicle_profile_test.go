package bazi

import "testing"

func TestVehicleRoadProfileOnCalculate(t *testing.T) {
	result := Calculate(1995, 10, 12, 12, "male", false, 0, "solar", false)
	if result.VehicleProfile == nil {
		t.Fatal("VehicleProfile should be attached to BaziResult")
	}
	if result.VehicleProfile.Score < 0 || result.VehicleProfile.Score > 100 {
		t.Fatalf("vehicle score should be within 0-100, got %d", result.VehicleProfile.Score)
	}
	if result.VehicleProfile.Grade == "" {
		t.Fatal("vehicle grade should not be empty")
	}
	if result.VehicleProfile.VehicleType == "" {
		t.Fatal("vehicle type should not be empty")
	}
	if len(result.VehicleProfile.Evidences) < 4 {
		t.Fatalf("vehicle profile should expose evidence, got %d item(s)", len(result.VehicleProfile.Evidences))
	}
	if len(result.DayunRoadmap) == 0 {
		t.Fatal("DayunRoadmap should be attached to BaziResult")
	}
	if len(result.DayunRoadmap) != len(result.Dayun) {
		t.Fatalf("dayun roadmap should match dayun length, got %d vs %d", len(result.DayunRoadmap), len(result.Dayun))
	}
	for i, road := range result.DayunRoadmap {
		if road.DayunIndex != result.Dayun[i].Index {
			t.Fatalf("road index should match dayun index at %d, got %d vs %d", i, road.DayunIndex, result.Dayun[i].Index)
		}
		if road.Score < 0 || road.Score > 100 {
			t.Fatalf("road score should be within 0-100, got %d", road.Score)
		}
		if road.RoadType == "" || road.RoadLabel == "" {
			t.Fatalf("road should include type and label: %+v", road)
		}
		if road.QianRoad.Label == "" || road.HouRoad.Label == "" {
			t.Fatalf("road should include front/back phases: %+v", road)
		}
		if len(road.Evidences) == 0 {
			t.Fatalf("road should expose evidence: %+v", road)
		}
	}
}

func TestVehicleGradeThresholds(t *testing.T) {
	cases := []struct {
		score int
		grade string
		label string
	}{
		{95, "S", "协同型配置"},
		{90, "S", "协同型配置"},
		{89, "A", "稳健型配置"},
		{75, "A", "稳健型配置"},
		{74, "B", "实用型配置"},
		{60, "B", "实用型配置"},
		{59, "C", "特性型配置"},
		{45, "C", "特性型配置"},
		{44, "D", "调校型配置"},
	}
	for _, tc := range cases {
		grade, label := vehicleGrade(tc.score)
		if grade != tc.grade || label != tc.label {
			t.Fatalf("vehicleGrade(%d)=%s/%s, want %s/%s", tc.score, grade, label, tc.grade, tc.label)
		}
	}
}

func TestRoadTypeThresholds(t *testing.T) {
	cases := []struct {
		score int
		key   string
		label string
	}{
		{90, RoadTypeHighway, "高速路"},
		{85, RoadTypeHighway, "高速路"},
		{84, RoadTypeMainRoad, "城市主路"},
		{70, RoadTypeMainRoad, "城市主路"},
		{69, RoadTypeMountainRoad, "山路"},
		{55, RoadTypeMountainRoad, "山路"},
		{54, RoadTypeMuddyRoad, "泥路"},
		{40, RoadTypeMuddyRoad, "泥路"},
		{39, RoadTypeConstruction, "施工路段"},
	}
	for _, tc := range cases {
		key, label := roadTypeForScore(tc.score)
		if key != tc.key || label != tc.label {
			t.Fatalf("roadTypeForScore(%d)=%s/%s, want %s/%s", tc.score, key, label, tc.key, tc.label)
		}
	}
}

func TestVehicleRoadProfileFallbackWithoutOptionalSignals(t *testing.T) {
	result := &BaziResult{
		YearGan: "甲", YearZhi: "子",
		MonthGan: "丙", MonthZhi: "寅",
		DayGan: "戊", DayZhi: "辰",
		HourGan: "庚", HourZhi: "申",
		DayGanWuxing: "土",
		Wuxing:       WuxingStats{Mu: 2, Huo: 1, Tu: 2, Jin: 2, Shui: 1, Total: 8},
		Dayun: []DayunItem{
			{Index: 1, Gan: "辛", Zhi: "酉", GanShiShen: "伤官", ZhiShiShen: "伤官", DiShi: "死"},
		},
	}
	profile, roads := BuildVehicleRoadProfile(result)
	if profile == nil {
		t.Fatal("profile should still be generated without optional signals")
	}
	if profile.Score < 0 || profile.Score > 100 {
		t.Fatalf("fallback vehicle score out of bounds: %d", profile.Score)
	}
	if len(roads) != 1 {
		t.Fatalf("expected one road item, got %d", len(roads))
	}
	if roads[0].QianRoad.Label == "" || roads[0].HouRoad.Label == "" {
		t.Fatalf("fallback road phases should be present: %+v", roads[0])
	}
}

func TestDayunRoadRepresentativeCases(t *testing.T) {
	base := &BaziResult{
		DayGan:            "甲",
		Yongshen:          "火水",
		Jishen:            "金土",
		FavorableShishen:  []string{"食神", "正印"},
		AdverseShishen:    []string{"七杀", "偏财"},
		ShishenConfidence: "hard",
		Dayun: []DayunItem{
			{
				Index: 1, Gan: "丙", Zhi: "子",
				GanShiShen: "食神", ZhiShiShen: "正印", DiShi: "长生",
				JinBuHuan: &JinBuHuanResult{QianLevel: "吉", QianDesc: "丙为调候喜用天干。", HouLevel: "吉", HouDesc: "子为金不换喜用地支。"},
			},
			{
				Index: 2, Gan: "戊", Zhi: "申",
				GanShiShen: "偏财", ZhiShiShen: "七杀", DiShi: "绝",
				JinBuHuan: &JinBuHuanResult{QianLevel: "凶", QianDesc: "戊为调候忌神天干。", HouLevel: "凶", HouDesc: "申为金不换忌神地支。"},
			},
		},
	}
	roads := buildDayunRoadmap(base)
	if len(roads) != 2 {
		t.Fatalf("expected two road items, got %d", len(roads))
	}
	if roads[0].Score <= roads[1].Score {
		t.Fatalf("favorable road should score higher than adverse road: %+v vs %+v", roads[0], roads[1])
	}
	if roads[0].RoadType != RoadTypeHighway && roads[0].RoadType != RoadTypeMainRoad {
		t.Fatalf("favorable road should be highway/main road, got %s", roads[0].RoadType)
	}
	if roads[1].RoadType != RoadTypeConstruction && roads[1].RoadType != RoadTypeMuddyRoad {
		t.Fatalf("adverse road should be construction/muddy road, got %s", roads[1].RoadType)
	}
}
