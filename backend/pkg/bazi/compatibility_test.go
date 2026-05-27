package bazi

import (
	"strings"
	"testing"
)

func makeCompatNatal(yearGZ, monthGZ, dayGZ, hourGZ, gender string) *BaziResult {
	yr := []rune(yearGZ)
	mr := []rune(monthGZ)
	dr := []rune(dayGZ)
	hr := []rune(hourGZ)
	return &BaziResult{
		YearGan: string(yr[0]), YearZhi: string(yr[1]),
		MonthGan: string(mr[0]), MonthZhi: string(mr[1]),
		DayGan: string(dr[0]), DayZhi: string(dr[1]),
		HourGan: string(hr[0]), HourZhi: string(hr[1]),
		YearGanWuxing:  ganWuxing[string(yr[0])],
		MonthGanWuxing: ganWuxing[string(mr[0])],
		DayGanWuxing:   ganWuxing[string(dr[0])],
		HourGanWuxing:  ganWuxing[string(hr[0])],
		YearZhiWuxing:  zhiWuxing[string(yr[1])],
		MonthZhiWuxing: zhiWuxing[string(mr[1])],
		DayZhiWuxing:   zhiWuxing[string(dr[1])],
		HourZhiWuxing:  zhiWuxing[string(hr[1])],
		YearHideGan:    []string{string(yr[0])},
		MonthHideGan:   []string{string(mr[0])},
		DayHideGan:     []string{string(dr[0])},
		HourHideGan:    []string{string(hr[0])},
		Gender:         gender,
	}
}

func hasCompatEvidence(items []CompatibilityEvidence, evType string, polarity CompatibilityPolarity) bool {
	for _, item := range items {
		if item.Type == evType && item.Polarity == polarity {
			return true
		}
	}
	return false
}

func TestAnalyzeCompatibility_ReturnsCoreShape(t *testing.T) {
	a := makeCompatNatal("甲子", "丙寅", "甲子", "丁卯", "male")
	b := makeCompatNatal("己丑", "戊辰", "己丑", "庚申", "female")

	got := AnalyzeCompatibility(a, b)

	if got.OverallLevel == "" {
		t.Fatal("expected overall level")
	}
	if got.DimensionScores.Attraction == 0 ||
		got.DimensionScores.Stability == 0 ||
		got.DimensionScores.Communication == 0 ||
		got.DimensionScores.Practicality == 0 {
		t.Fatalf("expected all four dimensions to be populated, got %+v", got.DimensionScores)
	}
	if len(got.Evidences) == 0 {
		t.Fatal("expected compatibility evidences")
	}
	if len(got.SummaryTags) == 0 {
		t.Fatal("expected summary tags for history preview")
	}
}

func TestAnalyzeCompatibility_DayZhiLiuHeAddsPositiveStabilityEvidence(t *testing.T) {
	a := makeCompatNatal("甲子", "丙寅", "甲子", "丁卯", "male")
	b := makeCompatNatal("己丑", "戊辰", "辛丑", "庚申", "female")

	got := AnalyzeCompatibility(a, b)

	if !hasCompatEvidence(got.Evidences, "夫妻宫六合", CompatibilityPositive) {
		t.Fatalf("expected positive day-branch liuhe evidence, got %+v", got.Evidences)
	}
	if got.DimensionScores.Stability < 60 {
		t.Fatalf("expected liuhe to keep stability at or above baseline, got %d", got.DimensionScores.Stability)
	}
}

func TestAnalyzeCompatibility_DayZhiChongCreatesNegativeStabilityEvidence(t *testing.T) {
	a := makeCompatNatal("甲子", "丙寅", "甲子", "丁卯", "male")
	b := makeCompatNatal("己午", "戊午", "辛午", "庚申", "female")

	got := AnalyzeCompatibility(a, b)

	if !hasCompatEvidence(got.Evidences, "夫妻宫六冲", CompatibilityNegative) {
		t.Fatalf("expected negative day-branch clash evidence, got %+v", got.Evidences)
	}
	if got.DimensionScores.Stability >= 60 {
		t.Fatalf("expected clash to reduce stability below baseline, got %d", got.DimensionScores.Stability)
	}
	if got.OverallLevel == CompatibilityHigh {
		t.Fatalf("expected clash case not to be high compatibility, got %s", got.OverallLevel)
	}
}

func TestAnalyzeCompatibility_SpouseStarResonanceBoostsAttraction(t *testing.T) {
	// 甲日男命以土为配偶星；对方多土，应该增强吸引力。
	a := makeCompatNatal("甲子", "丙寅", "甲子", "丁卯", "male")
	b := makeCompatNatal("己丑", "戊辰", "己未", "戊戌", "female")

	got := AnalyzeCompatibility(a, b)

	if !hasCompatEvidence(got.Evidences, "配偶星呼应", CompatibilityPositive) {
		t.Fatalf("expected spouse-star resonance evidence, got %+v", got.Evidences)
	}
	if got.DimensionScores.Attraction <= 60 {
		t.Fatalf("expected spouse-star resonance to boost attraction above baseline, got %d", got.DimensionScores.Attraction)
	}
}

func TestAnalyzeCompatibility_ReturnsDurationAssessmentShape(t *testing.T) {
	a := makeCompatNatal("甲子", "丙寅", "甲子", "丁卯", "male")
	b := makeCompatNatal("己丑", "戊辰", "己丑", "庚申", "female")

	got := AnalyzeCompatibility(a, b)

	if got.DurationAssessment.OverallBand == "" {
		t.Fatal("expected duration overall band")
	}
	if got.DurationAssessment.Windows.ThreeMonths.Level == "" {
		t.Fatal("expected three-month window")
	}
	if got.DurationAssessment.Windows.OneYear.Level == "" {
		t.Fatal("expected one-year window")
	}
	if got.DurationAssessment.Windows.TwoYearsPlus.Level == "" {
		t.Fatal("expected two-years-plus window")
	}
	if len(got.DurationAssessment.Reasons) == 0 {
		t.Fatal("expected duration reasons")
	}
}

func TestAnalyzeCompatibility_DurationCanBeStrongShortTermButWeakLongTerm(t *testing.T) {
	a := makeCompatNatal("甲子", "丙寅", "甲子", "丁卯", "male")
	b := makeCompatNatal("己午", "戊午", "己午", "戊申", "female")

	got := AnalyzeCompatibility(a, b)

	if got.DurationAssessment.Windows.ThreeMonths.Level == "" || got.DurationAssessment.Windows.TwoYearsPlus.Level == "" {
		t.Fatal("expected both short and long windows")
	}
	if got.DurationAssessment.Windows.ThreeMonths.Level == got.DurationAssessment.Windows.TwoYearsPlus.Level {
		t.Fatalf("expected staged difference, got %+v", got.DurationAssessment.Windows)
	}
}

func TestAnalyzeCompatibility_ReturnsConsultingAssessment(t *testing.T) {
	a := makeCompatNatal("甲子", "丙寅", "甲子", "丁卯", "male")
	b := makeCompatNatal("己午", "戊午", "己午", "戊申", "female")

	got := AnalyzeCompatibility(a, b)

	if got.ConsultingAssessment.RelationshipDiagnosis.RelationshipType == "" {
		t.Fatal("expected relationship type")
	}
	if got.ConsultingAssessment.RelationshipDiagnosis.Verdict == "" {
		t.Fatal("expected verdict")
	}
	if len(got.ConsultingAssessment.RelationshipDiagnosis.TopFindings) == 0 {
		t.Fatal("expected top findings")
	}
	if got.ConsultingAssessment.DecisionAdvice.Recommendation == "" {
		t.Fatal("expected decision recommendation")
	}
	if len(got.ConsultingAssessment.StageRisks) != 3 {
		t.Fatalf("expected three stage risks, got %d", len(got.ConsultingAssessment.StageRisks))
	}
	if got.ConsultingAssessment.RelationshipStrategy.Communication == "" {
		t.Fatal("expected communication strategy")
	}
	if len(got.ConsultingAssessment.ClaimEvidenceLinks) == 0 {
		t.Fatal("expected claim evidence links")
	}
}

func TestAnalyzeCompatibility_EvidenceKeysAreStableAndLinked(t *testing.T) {
	a := makeCompatNatal("甲子", "丙寅", "甲子", "丁卯", "male")
	b := makeCompatNatal("己午", "戊午", "己午", "戊申", "female")

	got := AnalyzeCompatibility(a, b)
	keys := map[string]bool{}
	for _, evidence := range got.Evidences {
		if evidence.EvidenceKey == "" {
			t.Fatalf("expected evidence key for %+v", evidence)
		}
		if keys[evidence.EvidenceKey] {
			t.Fatalf("duplicate evidence key %q", evidence.EvidenceKey)
		}
		keys[evidence.EvidenceKey] = true
	}
	for _, finding := range got.ConsultingAssessment.RelationshipDiagnosis.TopFindings {
		for _, key := range finding.EvidenceKeys {
			if !keys[key] {
				t.Fatalf("top finding references missing evidence key %q", key)
			}
		}
	}
	for _, risk := range got.ConsultingAssessment.StageRisks {
		for _, key := range risk.EvidenceKeys {
			if !keys[key] {
				t.Fatalf("stage risk references missing evidence key %q", key)
			}
		}
	}
	for _, link := range got.ConsultingAssessment.ClaimEvidenceLinks {
		if link.ClaimID == "" || link.Claim == "" || link.Reasoning == "" {
			t.Fatalf("expected complete claim link, got %+v", link)
		}
		for _, key := range link.EvidenceKeys {
			if !keys[key] {
				t.Fatalf("claim link references missing evidence key %q", key)
			}
		}
	}
}

func TestAnalyzeCompatibility_EvidenceKeysAreASCIIAndSemantic(t *testing.T) {
	a := makeCompatNatal("甲子", "丙寅", "甲子", "丁卯", "male")
	b := makeCompatNatal("己午", "戊午", "己午", "戊申", "female")

	got := AnalyzeCompatibility(a, b)

	for _, evidence := range got.Evidences {
		if !isASCII(evidence.EvidenceKey) {
			t.Fatalf("expected ASCII evidence key, got %q", evidence.EvidenceKey)
		}
		if strings.Contains(evidence.EvidenceKey, evidence.Type) || strings.Contains(evidence.EvidenceKey, evidence.Title) {
			t.Fatalf("expected evidence key not to contain display text, got key %q for %+v", evidence.EvidenceKey, evidence)
		}
	}
}

func TestAnalyzeCompatibility_LegacySignalFixturesKeepCoreEvidence(t *testing.T) {
	tests := []struct {
		name         string
		a            *BaziResult
		b            *BaziResult
		wantEvidence []CompatibilityEvidence
	}{
		{
			name: "high affinity liuhe and spouse star",
			a:    makeCompatNatal("甲子", "丙寅", "甲子", "丁卯", "male"),
			b:    makeCompatNatal("己丑", "戊辰", "丁丑", "庚申", "female"),
			wantEvidence: []CompatibilityEvidence{
				{Type: "夫妻宫六合", Dimension: CompatibilityStability, Polarity: CompatibilityPositive, Source: "spouse_palace", Weight: 18},
				{Type: "配偶星呼应", Dimension: CompatibilityAttraction, Polarity: CompatibilityPositive, Source: "spouse_star", Weight: 14},
			},
		},
		{
			name: "medium mixed same day master",
			a:    makeCompatNatal("甲子", "丙寅", "甲子", "丁卯", "male"),
			b:    makeCompatNatal("甲申", "戊辰", "甲申", "庚申", "female"),
			wantEvidence: []CompatibilityEvidence{
				{Type: "日主同气", Dimension: CompatibilityCommunication, Polarity: CompatibilityPositive, Source: "day_master", Weight: 8},
			},
		},
		{
			name: "low pressure day branch clash",
			a:    makeCompatNatal("甲子", "丙寅", "甲子", "丁卯", "male"),
			b:    makeCompatNatal("己午", "戊午", "己午", "戊申", "female"),
			wantEvidence: []CompatibilityEvidence{
				{Type: "夫妻宫六冲", Dimension: CompatibilityStability, Polarity: CompatibilityNegative, Source: "spouse_palace", Weight: -18},
				{Type: "干支冲克偏多", Dimension: CompatibilityPracticality, Polarity: CompatibilityNegative, Source: "ganzhi", Weight: -10},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AnalyzeCompatibility(tt.a, tt.b)
			for _, want := range tt.wantEvidence {
				assertCompatibilityEvidence(t, got.Evidences, want)
			}
		})
	}
}

func TestCompatibilitySignalHelpersDefineStableDepthSources(t *testing.T) {
	if compatibilitySourceTenGodInteraction != "ten_god_interaction" {
		t.Fatalf("unexpected ten-god source %q", compatibilitySourceTenGodInteraction)
	}
	if compatibilitySourceFavorableElementSupport != "favorable_element_support" {
		t.Fatalf("unexpected favorable-element source %q", compatibilitySourceFavorableElementSupport)
	}
	if compatibilitySourceRelationshipPattern != "relationship_pattern" {
		t.Fatalf("unexpected relationship-pattern source %q", compatibilitySourceRelationshipPattern)
	}
	if cap := compatibilitySourceContributionCap(compatibilitySourceGanZhi, CompatibilityStability); cap <= 0 {
		t.Fatalf("expected positive source cap for gan-zhi stability, got %d", cap)
	}
}

func TestCompatibilityEvidenceCanCarryDirectionalMetadata(t *testing.T) {
	item := CompatibilityEvidence{
		Perspective: compatibilityPerspectiveSelfToPartner,
		Actor:       compatibilityActorSelf,
		Target:      compatibilityActorPartner,
	}
	if item.Perspective != "self_to_partner" || item.Actor != "self" || item.Target != "partner" {
		t.Fatalf("unexpected directional metadata: %+v", item)
	}
}

func TestAnalyzeCompatibility_AddsDirectionalTenGodInteractionSignals(t *testing.T) {
	tests := []struct {
		name       string
		partnerDay string
		wantType   string
		wantTitle  string
		wantPol    CompatibilityPolarity
	}{
		{name: "seal support", partnerDay: "壬子", wantType: "十神互动-印星", wantTitle: "支持与照拂感", wantPol: CompatibilityPositive},
		{name: "official pressure", partnerDay: "庚申", wantType: "十神互动-官杀", wantTitle: "责任与压力感", wantPol: CompatibilityMixed},
		{name: "wealth attraction", partnerDay: "戊辰", wantType: "十神互动-财星", wantTitle: "现实吸引与投入感", wantPol: CompatibilityPositive},
		{name: "output expression", partnerDay: "丙午", wantType: "十神互动-食伤", wantTitle: "表达与被激发感", wantPol: CompatibilityPositive},
		{name: "peer competition", partnerDay: "乙卯", wantType: "十神互动-比劫", wantTitle: "同频与竞争感", wantPol: CompatibilityMixed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			self := makeCompatNatal("甲子", "丙寅", "甲子", "丁卯", "male")
			partner := makeCompatNatal("己丑", "戊辰", tt.partnerDay, "庚申", "female")

			got := AnalyzeCompatibility(self, partner)

			evidence := findCompatibilityEvidence(got.Evidences, tt.wantType, compatibilitySourceTenGodInteraction)
			if evidence == nil {
				t.Fatalf("expected %s evidence in %+v", tt.wantType, got.Evidences)
			}
			if evidence.Title != tt.wantTitle {
				t.Fatalf("title = %q, want %q", evidence.Title, tt.wantTitle)
			}
			if evidence.Polarity != tt.wantPol {
				t.Fatalf("polarity = %q, want %q", evidence.Polarity, tt.wantPol)
			}
			if evidence.Perspective != compatibilityPerspectiveSelfToPartner ||
				evidence.Actor != compatibilityActorSelf ||
				evidence.Target != compatibilityActorPartner {
				t.Fatalf("expected self-to-partner metadata, got %+v", evidence)
			}
		})
	}
}

func TestAnalyzeCompatibility_AddsFavorableElementSupportSignals(t *testing.T) {
	self := makeCompatNatal("甲寅", "乙卯", "甲寅", "乙卯", "male")
	self.Yongshen = "水"
	partner := makeCompatNatal("壬子", "癸亥", "壬子", "癸亥", "female")

	got := AnalyzeCompatibility(self, partner)

	evidence := findCompatibilityEvidence(got.Evidences, "五行支持-喜用补足", compatibilitySourceFavorableElementSupport)
	if evidence == nil {
		t.Fatalf("expected favorable support evidence in %+v", got.Evidences)
	}
	if evidence.Polarity != CompatibilityPositive || evidence.Title != "补足倾向" {
		t.Fatalf("unexpected support evidence %+v", evidence)
	}
	if evidence.Perspective != compatibilityPerspectiveSelfToPartner {
		t.Fatalf("expected self-to-partner perspective, got %+v", evidence)
	}
}

func TestAnalyzeCompatibility_AddsFavorableElementPressureSignals(t *testing.T) {
	self := makeCompatNatal("甲寅", "乙卯", "甲寅", "乙卯", "male")
	self.Jishen = "金"
	partner := makeCompatNatal("庚申", "辛酉", "庚申", "辛酉", "female")

	got := AnalyzeCompatibility(self, partner)

	evidence := findCompatibilityEvidence(got.Evidences, "五行支持-忌神加压", compatibilitySourceFavorableElementSupport)
	if evidence == nil {
		t.Fatalf("expected pressure evidence in %+v", got.Evidences)
	}
	if evidence.Polarity != CompatibilityNegative || evidence.Title != "压力倾向" {
		t.Fatalf("unexpected pressure evidence %+v", evidence)
	}
}

func TestAnalyzeCompatibility_FavorableElementFallbackUsesTendencyLanguage(t *testing.T) {
	self := makeCompatNatal("甲寅", "乙卯", "甲寅", "乙卯", "male")
	partner := makeCompatNatal("壬子", "癸亥", "壬子", "癸亥", "female")

	got := AnalyzeCompatibility(self, partner)

	evidence := findCompatibilityEvidence(got.Evidences, "五行支持-结构互补", compatibilitySourceFavorableElementSupport)
	if evidence == nil {
		t.Fatalf("expected fallback support evidence in %+v", got.Evidences)
	}
	if evidence.Title != "结构互补倾向" {
		t.Fatalf("unexpected fallback title %q", evidence.Title)
	}
	if strings.Contains(evidence.Detail, "你的用神") || strings.Contains(evidence.Detail, " definitive") {
		t.Fatalf("fallback evidence overstates precision: %s", evidence.Detail)
	}
}

func TestAnalyzeCompatibility_AddsExpandedGanZhiInteractionSignals(t *testing.T) {
	self := makeCompatNatal("甲子", "丙寅", "甲子", "丁卯", "male")
	partner := makeCompatNatal("己丑", "戊辰", "己丑", "庚申", "female")

	got := AnalyzeCompatibility(self, partner)

	stem := findCompatibilityEvidence(got.Evidences, "干支互动-天干五合", compatibilitySourceGanZhiInteraction)
	if stem == nil {
		t.Fatalf("expected stem-combination evidence in %+v", got.Evidences)
	}
	if stem.Polarity != CompatibilityPositive || stem.Title != "天干相合" {
		t.Fatalf("unexpected stem evidence %+v", stem)
	}
	branch := findCompatibilityEvidence(got.Evidences, "干支互动-地支六合", compatibilitySourceGanZhiInteraction)
	if branch == nil {
		t.Fatalf("expected branch liuhe evidence in %+v", got.Evidences)
	}
	if branch.Dimension != CompatibilityStability || branch.Weight <= 0 {
		t.Fatalf("unexpected branch evidence %+v", branch)
	}
}

func TestAnalyzeCompatibility_GanZhiInteractionWeightsPrioritizeDayPillar(t *testing.T) {
	dayPair := AnalyzeCompatibility(
		makeCompatNatal("甲寅", "丙寅", "甲子", "丁卯", "male"),
		makeCompatNatal("己申", "戊辰", "己午", "庚申", "female"),
	)
	yearPair := AnalyzeCompatibility(
		makeCompatNatal("甲子", "丙寅", "甲寅", "丁卯", "male"),
		makeCompatNatal("己午", "戊辰", "己亥", "庚申", "female"),
	)

	dayClash := findCompatibilityEvidence(dayPair.Evidences, "干支互动-地支六冲", compatibilitySourceGanZhiInteraction)
	yearClash := findCompatibilityEvidence(yearPair.Evidences, "干支互动-地支六冲", compatibilitySourceGanZhiInteraction)
	if dayClash == nil || yearClash == nil {
		t.Fatalf("expected both day and year clash evidence, day=%+v year=%+v", dayPair.Evidences, yearPair.Evidences)
	}
	if absInt(dayClash.Weight) <= absInt(yearClash.Weight) {
		t.Fatalf("expected day clash weight %d to outrank year clash weight %d", dayClash.Weight, yearClash.Weight)
	}
}

func TestAnalyzeCompatibility_GanZhiInteractionCanReturnMixedEvidence(t *testing.T) {
	self := makeCompatNatal("甲子", "丙寅", "甲子", "丁卯", "male")
	partner := makeCompatNatal("己午", "戊午", "己丑", "庚申", "female")

	got := AnalyzeCompatibility(self, partner)

	if findCompatibilityEvidence(got.Evidences, "干支互动-地支六合", compatibilitySourceGanZhiInteraction) == nil {
		t.Fatalf("expected positive branch-combination evidence in %+v", got.Evidences)
	}
	if findCompatibilityEvidence(got.Evidences, "干支互动-地支六冲", compatibilitySourceGanZhiInteraction) == nil {
		t.Fatalf("expected negative branch-clash evidence in %+v", got.Evidences)
	}
}

func TestAnalyzeCompatibility_AddsRelationshipPatternSignals(t *testing.T) {
	self := makeCompatNatal("甲子", "丙寅", "甲子", "丁卯", "male")
	partner := makeCompatNatal("己午", "戊午", "己午", "庚申", "female")

	got := AnalyzeCompatibility(self, partner)

	evidence := findCompatibilityEvidence(got.Evidences, "关系模式-冲突触发", compatibilitySourceRelationshipPattern)
	if evidence == nil {
		t.Fatalf("expected conflict pattern evidence in %+v", got.Evidences)
	}
	if evidence.Weight >= 0 || absInt(evidence.Weight) > compatibilitySourceContributionCap(compatibilitySourceRelationshipPattern, evidence.Dimension) {
		t.Fatalf("unexpected bounded pattern weight %+v", evidence)
	}
	if len(evidence.RelatedSources) == 0 {
		t.Fatalf("expected pattern evidence to reference related source families: %+v", evidence)
	}
}

func TestAnalyzeCompatibility_RelationshipPatternCanSummarizeRealitySupport(t *testing.T) {
	self := makeCompatNatal("甲寅", "乙卯", "甲寅", "乙卯", "male")
	self.Yongshen = "水"
	partner := makeCompatNatal("壬子", "癸亥", "壬子", "癸亥", "female")

	got := AnalyzeCompatibility(self, partner)

	evidence := findCompatibilityEvidence(got.Evidences, "关系模式-现实承接", compatibilitySourceRelationshipPattern)
	if evidence == nil {
		t.Fatalf("expected reality-support pattern evidence in %+v", got.Evidences)
	}
	if evidence.Polarity != CompatibilityPositive || evidence.Weight <= 0 {
		t.Fatalf("unexpected reality pattern evidence %+v", evidence)
	}
}

func TestCompatibilityBuilderCapsRepeatedSourceContribution(t *testing.T) {
	builder := newCompatibilityAnalysisBuilder()
	builder.addEvidence(CompatibilityEvidence{
		Dimension: CompatibilityStability,
		Type:      "干支互动-地支六冲",
		Polarity:  CompatibilityNegative,
		Source:    compatibilitySourceGanZhiInteraction,
		Title:     "first",
		Weight:    -20,
	})
	builder.addEvidence(CompatibilityEvidence{
		Dimension: CompatibilityStability,
		Type:      "干支互动-地支六冲",
		Polarity:  CompatibilityNegative,
		Source:    compatibilitySourceGanZhiInteraction,
		Title:     "second",
		Weight:    -20,
	})

	cap := compatibilitySourceContributionCap(compatibilitySourceGanZhiInteraction, CompatibilityStability)
	if gotDrop := 60 - builder.scores.Stability; gotDrop != cap {
		t.Fatalf("score drop = %d, want capped drop %d", gotDrop, cap)
	}
	totalWeight := 0
	for _, item := range builder.evidences {
		totalWeight += item.Weight
	}
	if absInt(totalWeight) != cap {
		t.Fatalf("evidence total weight = %d, want capped %d", totalWeight, cap)
	}
}

func TestAnalyzeCompatibility_ReturnsScoreExplanations(t *testing.T) {
	got := AnalyzeCompatibility(
		makeCompatNatal("甲子", "丙寅", "甲子", "丁卯", "male"),
		makeCompatNatal("己午", "戊午", "己丑", "庚申", "female"),
	)

	explanation := findCompatibilityScoreExplanation(got.ScoreExplanations, CompatibilityStability)
	if explanation == nil {
		t.Fatalf("expected stability explanation in %+v", got.ScoreExplanations)
	}
	if explanation.PositiveFactor == "" || explanation.NegativeFactor == "" {
		t.Fatalf("expected mixed stability factors, got %+v", explanation)
	}
	if len(explanation.PositiveEvidenceKeys) == 0 || len(explanation.NegativeEvidenceKeys) == 0 {
		t.Fatalf("expected evidence keys in explanation %+v", explanation)
	}
}

func TestBuildCompatibilityScoreExplanations_UsesNeutralLanguageForLimitedEvidence(t *testing.T) {
	got := buildCompatibilityScoreExplanations(nil)
	explanation := findCompatibilityScoreExplanation(got, CompatibilityAttraction)
	if explanation == nil {
		t.Fatalf("expected attraction explanation")
	}
	if !strings.Contains(explanation.Summary, "证据有限") {
		t.Fatalf("expected limited-evidence summary, got %+v", explanation)
	}
}

func findCompatibilityScoreExplanation(items []CompatibilityScoreExplanation, dimension CompatibilityDimension) *CompatibilityScoreExplanation {
	for i := range items {
		if items[i].Dimension == dimension {
			return &items[i]
		}
	}
	return nil
}

func findCompatibilityEvidence(items []CompatibilityEvidence, evType, source string) *CompatibilityEvidence {
	for i := range items {
		if items[i].Type == evType && items[i].Source == source {
			return &items[i]
		}
	}
	return nil
}

func assertCompatibilityEvidence(t *testing.T, got []CompatibilityEvidence, want CompatibilityEvidence) {
	t.Helper()
	for _, item := range got {
		if item.Type == want.Type &&
			item.Dimension == want.Dimension &&
			item.Polarity == want.Polarity &&
			item.Source == want.Source &&
			item.Weight == want.Weight {
			if item.EvidenceKey == "" {
				t.Fatalf("matched evidence %s but evidence key is empty", want.Type)
			}
			return
		}
	}
	t.Fatalf("missing evidence %+v in %+v", want, got)
}

func TestBuildCompatibilityConsultingAssessment_WithoutEvidenceHasNoClaimEvidenceLinks(t *testing.T) {
	got := buildCompatibilityConsultingAssessment(
		CompatibilityDimensionScores{Attraction: 60, Stability: 60, Communication: 60, Practicality: 60},
		nil,
		CompatibilityDurationAssessment{},
	)

	if len(got.ClaimEvidenceLinks) != 0 {
		t.Fatalf("expected no claim evidence links without evidence, got %+v", got.ClaimEvidenceLinks)
	}
}

func TestBuildCompatibilityConsultingAssessment_EmptyEvidenceKeyHasNoClaimEvidenceLinks(t *testing.T) {
	got := buildCompatibilityConsultingAssessment(
		CompatibilityDimensionScores{Attraction: 60, Stability: 60, Communication: 60, Practicality: 60},
		[]CompatibilityEvidence{
			{
				Dimension: CompatibilityStability,
				Type:      "夫妻宫六冲",
				Polarity:  CompatibilityNegative,
				Source:    "spouse_palace",
				Title:     "夫妻宫六冲",
				Detail:    "empty-key regression fixture",
				Weight:    -18,
			},
		},
		CompatibilityDurationAssessment{},
	)

	if len(got.ClaimEvidenceLinks) != 0 {
		t.Fatalf("expected no claim evidence links for empty evidence key, got %+v", got.ClaimEvidenceLinks)
	}
	for _, link := range got.ClaimEvidenceLinks {
		for _, key := range link.EvidenceKeys {
			if key == "" {
				t.Fatalf("claim link references empty evidence key: %+v", link)
			}
		}
	}
}

func isASCII(value string) bool {
	for _, r := range value {
		if r > 127 {
			return false
		}
	}
	return true
}

func TestBuildEvidences_ZodiacLiuhe(t *testing.T) {
	a := makeCompatNatal("甲子", "丙寅", "壬午", "丁未", "male")
	b := makeCompatNatal("己丑", "戊辰", "庚申", "辛酉", "female")
	ev := buildCompatibilityEvidencesV3(a, b)
	found := false
	for _, item := range ev {
		if item.EvidenceKey == "zodiac_liuhe" {
			if item.Weight != 50 || item.Dimension != "zodiac" || item.Polarity != "positive" {
				t.Errorf("zodiac_liuhe: bad shape %+v", item)
			}
			found = true
		}
	}
	if !found {
		t.Error("expected zodiac_liuhe evidence for 子/丑 pair")
	}
}

func TestBuildEvidences_AllHits_Count6(t *testing.T) {
	// 构造一组让 4 模块全部命中、且 eight_chars 三柱全命中的盘。
	// 甲子/乙丑：年支 子/丑 六合（zodiac），纳音同为「金」（nayin same），
	// 日柱地支六合 + 干五行同行 → 下档命中（day_pillar），三柱同理（eight_chars × 3）。
	a := makeCompatNatal("甲子", "甲子", "甲子", "甲子", "male")
	b := makeCompatNatal("乙丑", "乙丑", "乙丑", "乙丑", "female")
	ev := buildCompatibilityEvidencesV3(a, b)
	// zodiac + nayin + day_pillar + eight_chars(year/month/hour) = 1+1+1+3 = 6
	if len(ev) != 6 {
		t.Errorf("all-hit case: got %d evidences, want 6", len(ev))
	}
}

func TestBuildScoreExplanationsV3_FourEntries(t *testing.T) {
	a := makeCompatNatal("甲子", "丙寅", "壬午", "丁未", "male")
	b := makeCompatNatal("己丑", "戊辰", "庚申", "辛酉", "female")
	ev := buildCompatibilityEvidencesV3(a, b)
	exps := buildScoreExplanationsV3(a, b, ev)
	if len(exps) != 4 {
		t.Fatalf("expected exactly 4 explanations (one per module), got %d", len(exps))
	}
	dims := map[string]bool{}
	for _, e := range exps {
		dims[string(e.Dimension)] = true
		if e.NegativeFactor != "" || len(e.NegativeEvidenceKeys) != 0 {
			t.Errorf("v3 should never set negative factors, got %+v", e)
		}
	}
	for _, d := range []string{"zodiac", "nayin", "day_pillar", "eight_chars"} {
		if !dims[d] {
			t.Errorf("missing dimension %q", d)
		}
	}
}

func TestBuildScoreExplanationsV3_UnHitModule_HasSummary(t *testing.T) {
	a := makeCompatNatal("甲午", "丙寅", "壬午", "丁未", "male")
	b := makeCompatNatal("乙未", "戊辰", "庚申", "辛酉", "female") // 午未六合，但其他模块可能不命中
	ev := buildCompatibilityEvidencesV3(a, b)
	exps := buildScoreExplanationsV3(a, b, ev)
	for _, e := range exps {
		if e.Summary == "" {
			t.Errorf("dimension %q has empty summary", e.Dimension)
		}
	}
}
