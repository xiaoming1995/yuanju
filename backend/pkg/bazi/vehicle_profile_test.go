package bazi

import (
	"strings"
	"testing"
)

func findRoadEvidence(evidences []ProfileEvidence, source string) *ProfileEvidence {
	for i := range evidences {
		if evidences[i].Source == source {
			return &evidences[i]
		}
	}
	return nil
}

func findRoadPhase(road DayunRoad, phase string) *DayunPhaseEvidence {
	for i := range road.PhaseEvidences {
		if road.PhaseEvidences[i].Phase == phase {
			return &road.PhaseEvidences[i]
		}
	}
	return nil
}

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
	if want := resolveVehicleType(result.VehicleProfile.Grade); result.VehicleProfile.VehicleType != want {
		t.Fatalf("vehicle type should follow final grade, got %q want %q", result.VehicleProfile.VehicleType, want)
	}
	if result.MingGe != "" && result.VehicleProfile.DrivingStyle == "" {
		t.Fatalf("Ming Ge %q should expose a separate professional driving style", result.MingGe)
	}
	if len(result.VehicleProfile.Evidences) < 4 {
		t.Fatalf("vehicle profile should expose evidence, got %d item(s)", len(result.VehicleProfile.Evidences))
	}
	if !contains(result.VehicleProfile.Tags, "日干调候成格·高格基础") {
		t.Fatalf("vehicle tags should distinguish the high day-stem foundation: %+v", result.VehicleProfile.Tags)
	}
	if !contains(result.VehicleProfile.Tags, "主格结构部分成立") {
		t.Fatalf("vehicle tags should retain distinct primary-pattern quality: %+v", result.VehicleProfile.Tags)
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
		if len(road.PhaseEvidences) != 2 {
			t.Fatalf("road should expose front/back phase evidence: %+v", road)
		}
	}
}

func TestDayunRoad1995RenWuKeepsOpposingPhaseEvidence(t *testing.T) {
	r := Calculate(1995, 10, 12, 12, "male", false, 0, "solar", false)
	var road *DayunRoad
	for i := range r.DayunRoadmap {
		if r.DayunRoadmap[i].GanZhi == "壬午" {
			road = &r.DayunRoadmap[i]
			break
		}
	}
	if road == nil {
		t.Fatalf("expected 壬午 road in 1995 fixture: %+v", r.DayunRoadmap)
	}

	front := findRoadPhase(*road, "front")
	back := findRoadPhase(*road, "back")
	if front == nil || back == nil {
		t.Fatalf("expected front/back phase evidence: %+v", road.PhaseEvidences)
	}
	frontElement := findRoadEvidence(front.Evidences, "大运五行")
	backElement := findRoadEvidence(back.Evidences, "大运五行")
	if frontElement == nil || frontElement.Label != "壬" || frontElement.Delta <= 0 {
		t.Fatalf("expected positive 壬 front element evidence: %+v", front.Evidences)
	}
	if backElement == nil || backElement.Label != "午" || backElement.Delta >= 0 {
		t.Fatalf("expected negative 午 back element evidence: %+v", back.Evidences)
	}
	if aggregate := findRoadEvidence(road.Evidences, "大运五行"); aggregate == nil || aggregate.Delta != frontElement.Delta+backElement.Delta {
		t.Fatalf("element aggregate must equal phase sum: aggregate=%+v front=%+v back=%+v", aggregate, frontElement, backElement)
	}

	frontTenGod := findRoadEvidence(front.Evidences, "大运十神")
	backTenGod := findRoadEvidence(back.Evidences, "大运十神")
	if frontTenGod == nil || frontTenGod.Label != "七杀" || frontTenGod.Delta <= 0 {
		t.Fatalf("expected positive 七杀 front Ten God evidence: %+v", front.Evidences)
	}
	if backTenGod == nil || backTenGod.Label != "劫财" || backTenGod.Delta >= 0 {
		t.Fatalf("expected negative 劫财 back Ten God evidence: %+v", back.Evidences)
	}

	frontPattern := findRoadEvidence(front.Evidences, "格局作用")
	backPattern := findRoadEvidence(back.Evidences, "格局作用")
	if frontPattern == nil || backPattern == nil || frontPattern.Label != "杀印相生" || frontPattern.Delta <= 0 {
		t.Fatalf("expected verified 杀印相生 in front phase: %+v", front.Evidences)
	}
	if aggregate := findRoadEvidence(road.Evidences, "格局作用"); aggregate == nil || aggregate.Delta != frontPattern.Delta+backPattern.Delta {
		t.Fatalf("pattern aggregate must equal phase sum: aggregate=%+v front=%+v back=%+v", aggregate, frontPattern, back.Evidences)
	}
}

func TestDayunTenGodPhaseWeightingDoesNotCancelBeforeScaling(t *testing.T) {
	r := &BaziResult{
		FavorableShishen:  []string{"七杀"},
		AdverseShishen:    []string{"劫财"},
		ShishenConfidence: "medium",
	}
	front, back := dayunTenGodPhaseEvidences(r, DayunItem{GanShiShen: "七杀", ZhiShiShen: "劫财"})
	if front.Delta != 7 || back.Delta != -7 {
		t.Fatalf("medium confidence should weight phases separately, got front=%+v back=%+v", front, back)
	}
	if front.Delta+back.Delta != 0 {
		t.Fatalf("opposing weighted phase evidence should still aggregate to zero, got %d", front.Delta+back.Delta)
	}
}

func TestDayunKillPrintRequiresVisibleRootedNatalPrint(t *testing.T) {
	r := Calculate(1995, 10, 12, 12, "male", false, 0, "solar", false)
	r.HourGan = "戊" // Keep 偏印 in hidden stems only.
	r.NatalAssessment.Pattern.Name = "偏印格"
	front, _ := dayunPatternPhaseEvidences(r, r.NatalAssessment, DayunItem{Gan: "壬", GanShiShen: "七杀"})
	if front.Delta != 0 || front.Label == "杀印相生" {
		t.Fatalf("hidden-only 偏印 must not form 杀印相生: %+v", front)
	}
	if !hasSubstring(front.Detail, "偏印未透干有根") {
		t.Fatalf("expected explicit hidden-only reason, got %q", front.Detail)
	}
}

func hasSubstring(value, want string) bool {
	for i := 0; i+len(want) <= len(value); i++ {
		if value[i:i+len(want)] == want {
			return true
		}
	}
	return false
}

func TestVehicleGradeThresholds(t *testing.T) {
	cases := []struct {
		score int
		grade string
		label string
	}{
		{100, "S", "上格配置"},
		{85, "S", "上格配置"},
		{84, "A", "中上格配置"},
		{70, "A", "中上格配置"},
		{69, "B", "中格配置"},
		{50, "B", "中格配置"},
		{49, "C", "中下格配置"},
		{30, "C", "中下格配置"},
		{29, "D", "下格配置"},
	}
	for _, tc := range cases {
		grade, label := vehicleGrade(tc.score)
		if grade != tc.grade || label != tc.label {
			t.Fatalf("vehicleGrade(%d)=%s/%s, want %s/%s", tc.score, grade, label, tc.grade, tc.label)
		}
	}
}

func TestVehicleTypeFollowsFinalGrade(t *testing.T) {
	cases := []struct {
		grade string
		want  string
	}{
		{"S", "超跑级座驾"},
		{"A", "高性能车级座驾"},
		{"B", "标准轿车级座驾"},
		{"C", "实用 MPV 级座驾"},
		{"D", "基础代步单车级"},
	}
	for _, tc := range cases {
		if got := resolveVehicleType(tc.grade); got != tc.want {
			t.Fatalf("resolveVehicleType(%s)=%q, want %q", tc.grade, got, tc.want)
		}
	}

	for _, mingge := range []string{"七杀格", "正官格", "食神格", "杂气格"} {
		if got := resolveVehicleType("B"); got != "标准轿车级座驾" {
			t.Fatalf("grade B with %s resolved to %q, want 标准轿车级座驾", mingge, got)
		}
	}
}

func TestDrivingStyleIsSeparateFromPrimaryVehicleType(t *testing.T) {
	if got := resolveDrivingStyle("七杀格"); got == "" {
		t.Fatal("七杀格 should expose a professional driving style")
	}
	if got := resolveDrivingStyle("杂气格"); got == "" {
		t.Fatal("杂气格 should expose a professional driving style")
	}
	if got := resolveDrivingStyle(""); got != "" {
		t.Fatalf("missing Ming Ge should not invent a driving style, got %q", got)
	}
	if resolveVehicleType("A") != "高性能车级座驾" {
		t.Fatal("Ming Ge-derived driving styles must not replace the grade-derived vehicle type")
	}
}

func TestVehicleTierTiaohouAndFuyiCeilings(t *testing.T) {
	cold := &BaziResult{
		YearGan: "甲", YearZhi: "申",
		MonthGan: "乙", MonthZhi: "子",
		DayGan: "甲", DayZhi: "辰",
		HourGan: "庚", HourZhi: "丑",
	}
	classical := ComputeClassicalYongshen(cold)
	if classical.Strategy != ClassicalStrategyTiaohouCold {
		t.Fatalf("expected cold urgency, got %s", classical.Strategy)
	}
	if score, label, _, ceiling := vehicleUrgentTiaohouScore(cold, classical); score != 0 || label != "调候急需未解" || ceiling != "C" {
		t.Fatalf("unresolved cold Tiaohou = %d/%s/%s, want 0/调候急需未解/C", score, label, ceiling)
	}

	cold.HourHideGan = []string{"丙"}
	if score, label, _, ceiling := vehicleUrgentTiaohouScore(cold, classical); score != 18 || label != "调候急需藏支" || ceiling != "A" {
		t.Fatalf("hidden cold Tiaohou = %d/%s/%s, want 18/调候急需藏支/A", score, label, ceiling)
	}

	if grade, _ := vehicleGradeWithCeiling(95, "C"); grade != "C" {
		t.Fatalf("unresolved Tiaohou must cap 95 points at C, got %s", grade)
	}
	if grade, _ := vehicleGradeWithCeiling(95, "B"); grade != "B" {
		t.Fatalf("unusable Fuyi must cap 95 points at B, got %s", grade)
	}
}

func TestVehicleFuyiSupportAndConfidence(t *testing.T) {
	base := &BaziResult{
		YearGan: "庚", YearZhi: "申",
		MonthGan: "己", MonthZhi: "丑",
		DayGan: "甲", DayZhi: "辰",
		HourGan: "辛", HourZhi: "戌",
		YearHideGan:  []string{"庚"},
		MonthHideGan: []string{"己"},
		DayHideGan:   []string{"戊"},
		HourHideGan:  []string{"辛"},
		Wuxing:       WuxingStats{Mu: 1, Huo: 0, Tu: 3, Jin: 3, Shui: 1, Total: 8},
	}
	if score, label, _, ceiling := vehicleFuyiScore(base, "金土火", "水木"); score < 24 || label != "扶抑用神得力" || ceiling != "S" {
		t.Fatalf("supported Fuyi = %d/%s/%s, want >=24/扶抑用神得力/S", score, label, ceiling)
	}
	if score, label, _, ceiling := vehicleFuyiScore(base, "水木", "金土火"); score != 0 || label != "扶抑用神无力" || ceiling != "B" {
		t.Fatalf("unsupported Fuyi = %d/%s/%s, want 0/扶抑用神无力/B", score, label, ceiling)
	}

	base.ShishenConfidence = "hard"
	hard, _ := BuildVehicleRoadProfile(base)
	base.ShishenConfidence = "soft"
	soft, _ := BuildVehicleRoadProfile(base)
	if hard.Score != soft.Score || hard.Grade != soft.Grade {
		t.Fatalf("Shishen confidence must not affect natal tier: hard=%d/%s soft=%d/%s", hard.Score, hard.Grade, soft.Score, soft.Grade)
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

func TestDayunRoadSummarySeparatesCompositeRoadFromJinBuHuanPhases(t *testing.T) {
	summary := dayunRoadSummary(
		DayunItem{Gan: "癸", Zhi: "巳"},
		"泥路",
		RoadPhase{Label: "施工路段", Score: 35},
		RoadPhase{Label: "施工路段", Score: 35},
	)

	for _, want := range []string{
		"十年综合路况为泥路",
		"金不换阶段提示",
		"前五年由天干癸主事，评级凶",
		"后五年由地支巳主事，评级凶",
	} {
		if !strings.Contains(summary, want) {
			t.Fatalf("summary should contain %q, got %q", want, summary)
		}
	}
	if strings.Contains(summary, "整体路况施工路段") {
		t.Fatalf("summary must not present phase labels as the overall road condition: %q", summary)
	}
}

func TestDayunRoadKeepsNatalStrengthAndAttributesDiShiToBranchPolarity(t *testing.T) {
	// 1987-12-09 卯时：丁卯 壬子 壬辰 癸卯。壬水得令、透比劫且有根，原局身强。
	r := Calculate(1987, 12, 9, 6, "male", false, 0, "solar", false)
	if r.NatalAssessment == nil {
		t.Fatal("expected natal assessment")
	}
	if got := r.NatalAssessment.Fuyi.DayMasterStrength; got != "strong" && got != "vstrong" {
		t.Fatalf("expected natal strong assessment, got %q: %+v", got, r.NatalAssessment.Fuyi)
	}
	natalStrength := r.NatalAssessment.Fuyi.DayMasterStrength

	var road *DayunRoad
	for i := range r.DayunRoadmap {
		if r.DayunRoadmap[i].GanZhi == "戊申" {
			road = &r.DayunRoadmap[i]
			break
		}
	}
	if road == nil {
		t.Fatalf("expected 戊申 Dayun in fixture: %+v", r.DayunRoadmap)
	}
	if r.NatalAssessment.Fuyi.DayMasterStrength != natalStrength {
		t.Fatalf("Dayun evaluation must not change natal strength: before=%s after=%s", natalStrength, r.NatalAssessment.Fuyi.DayMasterStrength)
	}

	back := findRoadPhase(*road, "back")
	if back == nil {
		t.Fatalf("expected back phase evidence: %+v", road.PhaseEvidences)
	}
	diShi := findRoadEvidence(back.Evidences, "十二长生")
	if diShi == nil || diShi.Delta >= 0 {
		t.Fatalf("申金为扶抑忌且长生应加强忌神，不应给正分: %+v", diShi)
	}
	if !strings.Contains(diShi.Detail, "扶抑忌金") || !strings.Contains(diShi.Detail, "长生") {
		t.Fatalf("expected branch polarity and stage detail, got %q", diShi.Detail)
	}
}

func TestDayunBranchDiShiIntensityFollowsFuyiPolarity(t *testing.T) {
	assessment := &NatalAssessment{Fuyi: NatalFuyiAssessment{Yongshen: "水", Jishen: "金"}}
	cases := []struct {
		name      string
		dayun     DayunItem
		wantDelta int
		wantText  string
	}{
		{name: "vigorous favorable branch", dayun: DayunItem{Zhi: "子", DiShi: "长生"}, wantDelta: 5, wantText: "扶抑喜用水"},
		{name: "vigorous adverse branch", dayun: DayunItem{Zhi: "申", DiShi: "长生"}, wantDelta: -5, wantText: "扶抑忌金"},
		{name: "weak adverse branch is suppressed", dayun: DayunItem{Zhi: "申", DiShi: "绝"}, wantDelta: 0, wantText: "力量受限"},
		{name: "neutral branch has no directional score", dayun: DayunItem{Zhi: "辰", DiShi: "长生"}, wantDelta: 0, wantText: "按原局扶抑喜忌判断"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			evidence := dayunBranchDiShiEvidence(&BaziResult{}, assessment, tc.dayun)
			if evidence.Delta != tc.wantDelta {
				t.Fatalf("delta=%d, want %d: %+v", evidence.Delta, tc.wantDelta, evidence)
			}
			if !strings.Contains(evidence.Detail, tc.wantText) {
				t.Fatalf("detail %q should contain %q", evidence.Detail, tc.wantText)
			}
		})
	}
}
