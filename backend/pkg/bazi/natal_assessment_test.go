package bazi

import (
	"strings"
	"testing"
)

func assessmentFixture(mingGe string, gans ...string) *BaziResult {
	r := &BaziResult{
		YearGan: "乙", YearZhi: "寅",
		MonthGan: "丙", MonthZhi: "辰",
		DayGan: "甲", DayZhi: "午",
		HourGan: "戊", HourZhi: "申",
		YearHideGan:  []string{"甲", "丙", "戊"},
		MonthHideGan: []string{"戊", "乙", "癸"},
		DayHideGan:   []string{"丁", "己"},
		HourHideGan:  []string{"庚", "壬", "戊"},
		MingGe:       mingGe,
		Wuxing:       WuxingStats{Mu: 2, Huo: 2, Tu: 2, Jin: 1, Shui: 1, Total: 8},
		Yongshen:     "水木",
		Jishen:       "金土",
	}
	if len(gans) > 0 {
		r.YearGan = gans[0]
	}
	if len(gans) > 1 {
		r.MonthGan = gans[1]
	}
	if len(gans) > 2 {
		r.HourGan = gans[2]
	}
	return r
}

func TestNatalAssessmentUsesFuyiWithoutOverwritingLegacyUseGod(t *testing.T) {
	r := assessmentFixture("伤官格", "丁", "癸", "己")
	legacyYongshen := r.Yongshen
	assessment := AssessNatalStructure(r)
	if assessment == nil || assessment.Version != NatalAssessmentVersion {
		t.Fatalf("expected current assessment, got %+v", assessment)
	}
	if assessment.Tiaohou.Thermal.Status != "seasonal_resolved" {
		t.Fatalf("expected seasonal thermal assessment, got %+v", assessment.Tiaohou.Thermal)
	}
	if assessment.Fuyi.Evidence == "" {
		t.Fatal("expected global Fuyi evidence")
	}
	if r.Yongshen != legacyYongshen {
		t.Fatalf("assessment must preserve legacy Yongshen, got %q want %q", r.Yongshen, legacyYongshen)
	}
	if !contains(assessment.Pattern.Formations, "伤官配印") {
		t.Fatalf("expected 伤官配印 formation, got %+v", assessment.Pattern)
	}
}

func TestNatalAssessmentRecordsUnremediedPatternBreak(t *testing.T) {
	r := assessmentFixture("伤官格", "丁", "辛", "己")
	r.YearZhi = "酉" // 辛金通根，满足关键破格的透干有根条件。
	r.MonthHideGan = []string{"戊", "乙"}
	r.HourHideGan = []string{"庚", "戊"}
	assessment := AssessNatalStructure(r)
	if !contains(assessment.Pattern.Breaks, "伤官见官") || !assessment.Pattern.CriticalBreak {
		t.Fatalf("expected critical 伤官见官, got %+v", assessment.Pattern)
	}
	if assessment.Grade.GradeCeiling != "C" {
		t.Fatalf("critical pattern break should cap at C, got %+v", assessment.Grade)
	}
}

func TestNatalAssessmentKeepsHiddenFoodGodAsPatternContext(t *testing.T) {
	r := Calculate(1995, 10, 12, 12, "male", false, 0, "solar", false)
	assessment := r.NatalAssessment
	if assessment == nil {
		t.Fatal("expected natal assessment")
	}
	if !contains(assessment.Tiaohou.DayStem.RequiredStems, "甲") || !contains(assessment.Tiaohou.DayStem.RequiredStems, "壬") {
		t.Fatalf("expected Bing-Xu day-stem Tiaohou requirements 甲壬, got %+v", assessment.Tiaohou.DayStem)
	}
	if !contains(assessment.Tiaohou.DayStem.VisibleStems, "甲") || !contains(assessment.Tiaohou.DayStem.HiddenStems, "壬") {
		t.Fatalf("expected 甲 visible and 壬 hidden, got %+v", assessment.Tiaohou.DayStem)
	}
	if assessment.Tiaohou.DayStem.Status != "resolved" || assessment.Tiaohou.DayStem.Score != 12 || assessment.Tiaohou.DayStem.Formation != "formed" || assessment.Tiaohou.DayStem.FoundationTier != "high" || assessment.Tiaohou.DayStem.FoundationScore != 24 {
		t.Fatalf("expected day-stem Tiaohou formed/high foundation for 甲透壬藏, got %+v", assessment.Tiaohou.DayStem)
	}
	if assessment.Pattern.FoundationSource != "日干调候成格" || assessment.Pattern.FoundationLabel != "高格基础" || assessment.Pattern.FoundationTier != "high" {
		t.Fatalf("expected high day-stem foundation to remain distinct from primary pattern quality, got %+v", assessment.Pattern)
	}
	if assessment.Grade.Score != 69 || assessment.Grade.Grade != "B" {
		t.Fatalf("expected high foundation to lift chart to B/69 under retained Fuyi cap, got %+v", assessment.Grade)
	}
	foundFormationEvidence := false
	for _, evidence := range assessment.Evidences {
		if evidence.RuleID == "day-stem-tiaohou-formed" && evidence.Source == "日干调候成格" && evidence.Label == "高格基础" && evidence.Delta == 24 {
			foundFormationEvidence = true
		}
	}
	if !foundFormationEvidence {
		t.Fatalf("expected dedicated day-stem formation evidence, got %+v", assessment.Evidences)
	}
	if contains(assessment.Pattern.Breaks, "枭神夺食") || contains(assessment.Pattern.Formations, "印比相扶") {
		t.Fatalf("hidden-only Food God and Fuyi support must not score as pattern interaction: %+v", assessment.Pattern)
	}
	if assessment.Tiaohou.Thermal.Status != "seasonal_partial" || !contains(assessment.Tiaohou.Thermal.HiddenSupport, "壬") {
		t.Fatalf("expected hidden water as partial thermal support, got %+v", assessment.Tiaohou.Thermal)
	}
}

func TestNatalStemGuidance1995SeparatesPriorityAndConflict(t *testing.T) {
	r := Calculate(1995, 10, 12, 12, "male", false, 0, "solar", false)
	assessment := r.NatalAssessment
	if assessment == nil || assessment.StemGuidance == nil {
		t.Fatalf("expected stem guidance, got %+v", assessment)
	}
	guidance := assessment.StemGuidance
	if got := stemGuidanceStems(guidance.PrimaryFavorable); !sameStrings(got, []string{"壬"}) {
		t.Fatalf("expected 壬 as the only primary favorable stem, got %+v", guidance.PrimaryFavorable)
	}
	primary := guidance.PrimaryFavorable[0]
	if primary.Element != "水" || primary.TenGod != "七杀" || !sameStrings(primary.SourceLayers, []string{"日干调候", "扶抑"}) {
		t.Fatalf("expected 壬水七杀 with shared sources, got %+v", primary)
	}
	if got := stemGuidanceStems(guidance.SecondaryFavorable); !sameStrings(got, []string{"癸", "庚", "辛"}) {
		t.Fatalf("expected non-duplicated Fuyi-only stems, got %+v", guidance.SecondaryFavorable)
	}
	if got := stemGuidanceStems(guidance.ConditioningOnly); !sameStrings(got, []string{"甲"}) {
		t.Fatalf("expected 甲 as adjustment-only, got %+v", guidance.ConditioningOnly)
	}
	if !strings.Contains(guidance.ConditioningOnly[0].Detail, "不作为后天通用喜神") {
		t.Fatalf("expected explicit adjustment/Fuyi conflict detail, got %q", guidance.ConditioningOnly[0].Detail)
	}
	if got := stemGuidanceStems(guidance.Adverse); !sameStrings(got, []string{"乙", "丙", "丁"}) {
		t.Fatalf("expected adverse stems without duplicate 甲, got %+v", guidance.Adverse)
	}

	var renWu *DayunRoad
	for i := range r.DayunRoadmap {
		if r.DayunRoadmap[i].GanZhi == "壬午" {
			renWu = &r.DayunRoadmap[i]
			break
		}
	}
	if renWu == nil || renWu.Score != 69 || renWu.RoadType != "mountain_road" {
		t.Fatalf("stem guidance must not change the established 壬午 road result, got %+v", renWu)
	}
}

func TestNatalStemGuidanceDoesNotInventFuyiDirection(t *testing.T) {
	guidance := assessNatalStemGuidance(
		&BaziResult{DayGan: "丙"},
		NatalDayStemTiaohouAssessment{RequiredStems: []string{"甲", "壬"}},
		NatalFuyiAssessment{},
	)
	if len(guidance.PrimaryFavorable) != 0 || len(guidance.SecondaryFavorable) != 0 || len(guidance.Adverse) != 0 {
		t.Fatalf("neutral Fuyi must not invent stem direction: %+v", guidance)
	}
	if got := stemGuidanceStems(guidance.ConditioningOnly); !sameStrings(got, []string{"甲", "壬"}) {
		t.Fatalf("day-stem references should remain available, got %+v", guidance.ConditioningOnly)
	}
	for _, item := range guidance.ConditioningOnly {
		if !strings.Contains(item.Detail, "当前扶抑喜忌未定") {
			t.Fatalf("expected neutral Fuyi context, got %+v", item)
		}
	}
}

func stemGuidanceStems(items []NatalStemGuidanceItem) []string {
	stems := make([]string, 0, len(items))
	for _, item := range items {
		stems = append(stems, item.Stem)
	}
	return stems
}

func sameStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

func TestDayStemTiaohouScoresVisibleAndHiddenSupport(t *testing.T) {
	visibleOnly := assessDayStemTiaohou(&BaziResult{Tiaohou: &TiaohouResult{Expected: []string{"甲"}, Tou: []string{"甲"}}})
	if visibleOnly.Status != "partial" || visibleOnly.Score != 6 || visibleOnly.Formation != "partial" || visibleOnly.FoundationScore != 6 {
		t.Fatalf("visible-only day-stem Tiaohou should be partial/6, got %+v", visibleOnly)
	}

	hiddenOnly := assessDayStemTiaohou(&BaziResult{Tiaohou: &TiaohouResult{Expected: []string{"甲"}, Cang: []string{"甲"}}})
	if hiddenOnly.Status != "partial" || hiddenOnly.Score != 6 || hiddenOnly.Formation != "partial" || hiddenOnly.FoundationScore != 6 {
		t.Fatalf("hidden-only day-stem Tiaohou should be partial/6, got %+v", hiddenOnly)
	}
}

func TestDayStemTiaohouFormationRequiresExactCoverageAcrossHeavenAndEarth(t *testing.T) {
	cases := []struct {
		name      string
		tiaohou   *TiaohouResult
		formation string
		tier      string
	}{
		{
			name:      "one required stem appears in both locations",
			tiaohou:   &TiaohouResult{Expected: []string{"甲"}, Tou: []string{"甲"}, Cang: []string{"甲"}},
			formation: "formed", tier: "high",
		},
		{
			name:      "two required stems can be split across locations",
			tiaohou:   &TiaohouResult{Expected: []string{"甲", "壬"}, Tou: []string{"甲"}, Cang: []string{"壬"}},
			formation: "formed", tier: "high",
		},
		{
			name:      "unrelated hidden stem cannot cover missing required stem",
			tiaohou:   &TiaohouResult{Expected: []string{"甲", "壬"}, Tou: []string{"甲", "丙"}, Cang: []string{"丙"}},
			formation: "partial", tier: "normal",
		},
		{
			name:      "all coverage in one location remains partial",
			tiaohou:   &TiaohouResult{Expected: []string{"甲", "壬"}, Tou: []string{"甲", "壬"}},
			formation: "partial", tier: "normal",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assessment := assessDayStemTiaohou(&BaziResult{Tiaohou: tc.tiaohou})
			if assessment.Formation != tc.formation || assessment.FoundationTier != tc.tier {
				t.Fatalf("formation=%s/%s, want %s/%s: %+v", assessment.Formation, assessment.FoundationTier, tc.formation, tc.tier, assessment)
			}
		})
	}
}

func TestNatalAssessmentRequiresVisibleRootedFoodGodForOwlSeizingFood(t *testing.T) {
	r := &BaziResult{
		YearGan: "甲", YearZhi: "寅", // 偏印透干且通根
		MonthGan: "戊", MonthZhi: "戌", // 食神透干且通根
		DayGan: "丙", DayZhi: "午",
		HourGan: "乙", HourZhi: "卯",
		YearHideGan: []string{"甲", "丙", "戊"}, MonthHideGan: []string{"戊", "辛", "丁"},
		DayHideGan: []string{"丁", "己"}, HourHideGan: []string{"乙"},
		MingGe: "偏印格",
	}
	assessment := AssessNatalStructure(r)
	if !contains(assessment.Pattern.Breaks, "枭神夺食") {
		t.Fatalf("visible and rooted Food God should trigger Owl Seizing Food: %+v", assessment.Pattern)
	}
}

func TestNatalAssessmentTreatsHiddenThermalRemedyAsPartial(t *testing.T) {
	r := &BaziResult{
		YearGan: "甲", YearZhi: "申",
		MonthGan: "乙", MonthZhi: "子",
		DayGan: "甲", DayZhi: "辰",
		HourGan: "庚", HourZhi: "丑",
		HourHideGan: []string{"丙"},
		MingGe:      "正官格",
	}
	assessment := AssessNatalStructure(r)
	thermal := assessment.Tiaohou.Thermal
	if thermal.Status != "urgent_partial" || thermal.GradeCeiling != "A" || !contains(thermal.HiddenSupport, "丙") {
		t.Fatalf("hidden thermal remedy must remain partial, got %+v", thermal)
	}
}

func TestNatalAssessmentUrgentClimateCapsGrade(t *testing.T) {
	r := &BaziResult{
		YearGan: "甲", YearZhi: "申",
		MonthGan: "乙", MonthZhi: "子",
		DayGan: "甲", DayZhi: "辰",
		HourGan: "庚", HourZhi: "丑",
		MingGe: "正官格",
		Wuxing: WuxingStats{Mu: 2, Huo: 0, Tu: 2, Jin: 2, Shui: 2, Total: 8},
	}
	assessment := AssessNatalStructure(r)
	if assessment.Climate.Status != "urgent_unresolved" || assessment.Grade.GradeCeiling != "C" {
		t.Fatalf("expected unresolved climate C cap, got climate=%+v grade=%+v", assessment.Climate, assessment.Grade)
	}
}

func TestNatalAssessmentEarlySpringAlignsThermalAndFuyiFire(t *testing.T) {
	// 1996-02-08 戌时：丙子·庚寅·乙亥·丙戌。
	r := Calculate(1996, 2, 8, 20, "male", false, 0, "solar", false)
	assessment := r.NatalAssessment
	if assessment == nil {
		t.Fatal("expected natal assessment")
	}
	thermal := assessment.Tiaohou.Thermal
	if thermal.Condition != "初春偏寒" || thermal.RequiredElements != "火" || thermal.Status != "seasonal_resolved" {
		t.Fatalf("expected early-spring fire thermal resolution, got %+v", thermal)
	}
	if !contains(thermal.VisibleSupport, "丙") {
		t.Fatalf("expected visible Bing fire support, got %+v", thermal.VisibleSupport)
	}
	if assessment.Fuyi.DayMasterStrength != "vstrong" || assessment.Fuyi.Yongshen != "金土火" || !strings.Contains(assessment.Fuyi.Evidence, "月令木得令+12") {
		t.Fatalf("expected month-command Fuyi strength and complete yongshen, got %+v", assessment.Fuyi)
	}
	alignment := assessment.YongshenAlignment
	if !contains(alignment.Elements, "火") || !contains(alignment.SourceLayers, "寒热调候") || !contains(alignment.SourceLayers, "扶抑") {
		t.Fatalf("expected shared fire priority, got %+v", alignment)
	}
	if !strings.Contains(alignment.Detail, "共同优先用神：火") {
		t.Fatalf("expected shared priority explanation, got %q", alignment.Detail)
	}
}

func TestNatalAssessmentDoesNotAlignOneLayerOnly(t *testing.T) {
	alignment := assessSharedPriorityYongshen(
		NatalThermalTiaohouAssessment{RequiredElements: "火"},
		NatalFuyiAssessment{Yongshen: "水木"},
	)
	if len(alignment.Elements) != 0 || len(alignment.SourceLayers) != 0 {
		t.Fatalf("thermal fire without matching Fuyi fire must not align, got %+v", alignment)
	}
}

func TestSharedPriorityDoesNotCreateAnAdditionalGradeEvidence(t *testing.T) {
	r := Calculate(1996, 2, 8, 20, "male", false, 0, "solar", false)
	assessment := r.NatalAssessment
	if assessment == nil || len(assessment.YongshenAlignment.Elements) == 0 {
		t.Fatalf("expected aligned target fixture, got %+v", assessment)
	}
	for _, evidence := range assessment.Evidences {
		if evidence.RuleID == "shared-priority-yongshen" || evidence.Source == "共同优先用神" {
			t.Fatalf("shared priority is explanatory and must not add grade evidence: %+v", assessment.Evidences)
		}
	}
}

func TestDayunRoadAddsPatternInteractionEvidence(t *testing.T) {
	r := assessmentFixture("伤官格", "丁", "癸", "己")
	r.Dayun = []DayunItem{{Index: 1, Gan: "癸", Zhi: "亥", GanShiShen: "正印", ZhiShiShen: "偏印", DiShi: "长生"}}
	profile, roads := BuildVehicleRoadProfile(r)
	if profile == nil || len(roads) != 1 {
		t.Fatalf("expected vehicle profile and one road, got profile=%+v roads=%+v", profile, roads)
	}
	found := false
	for _, evidence := range roads[0].Evidences {
		if evidence.Source == "格局作用" && evidence.Label == "助格" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected Dayun to expose positive pattern interaction: %+v", roads[0].Evidences)
	}
}
