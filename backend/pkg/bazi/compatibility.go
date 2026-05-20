package bazi

import (
	"fmt"
	"sort"
	"strings"
)

type CompatibilityLevel string

const (
	CompatibilityHigh   CompatibilityLevel = "high"
	CompatibilityMedium CompatibilityLevel = "medium"
	CompatibilityLow    CompatibilityLevel = "low"
)

type CompatibilityPolarity string

const (
	CompatibilityPositive CompatibilityPolarity = "positive"
	CompatibilityNegative CompatibilityPolarity = "negative"
	CompatibilityMixed    CompatibilityPolarity = "mixed"
	CompatibilityNeutral  CompatibilityPolarity = "neutral"
)

type CompatibilityDimension string

const (
	CompatibilityAttraction    CompatibilityDimension = "attraction"
	CompatibilityStability     CompatibilityDimension = "stability"
	CompatibilityCommunication CompatibilityDimension = "communication"
	CompatibilityPracticality  CompatibilityDimension = "practicality"
)

type CompatibilityDurationLevel string

const (
	CompatibilityDurationHigh   CompatibilityDurationLevel = "high"
	CompatibilityDurationMedium CompatibilityDurationLevel = "medium"
	CompatibilityDurationLow    CompatibilityDurationLevel = "low"
)

type CompatibilityEvidence struct {
	EvidenceKey    string                 `json:"evidence_key"`
	Dimension      CompatibilityDimension `json:"dimension"`
	Type           string                 `json:"type"`
	Polarity       CompatibilityPolarity  `json:"polarity"`
	Source         string                 `json:"source"`
	Perspective    string                 `json:"perspective,omitempty"`
	Actor          string                 `json:"actor,omitempty"`
	Target         string                 `json:"target,omitempty"`
	RelatedSources []string               `json:"related_sources,omitempty"`
	Title          string                 `json:"title"`
	Detail         string                 `json:"detail"`
	Weight         int                    `json:"weight"`
}

type CompatibilityFinding struct {
	Text         string   `json:"text"`
	EvidenceKeys []string `json:"evidence_keys"`
}

type CompatibilityRelationshipDiagnosis struct {
	RelationshipType string                 `json:"relationship_type"`
	Verdict          string                 `json:"verdict"`
	Summary          string                 `json:"summary"`
	TopFindings      []CompatibilityFinding `json:"top_findings"`
}

type CompatibilityDecisionAdvice struct {
	Recommendation string   `json:"recommendation"`
	Confidence     string   `json:"confidence"`
	Conditions     []string `json:"conditions"`
	DoNext         []string `json:"do_next"`
	Avoid          []string `json:"avoid"`
}

type CompatibilityStageRisk struct {
	Window       string   `json:"window"`
	RiskLevel    string   `json:"risk_level"`
	MainRisk     string   `json:"main_risk"`
	Trigger      string   `json:"trigger"`
	Advice       string   `json:"advice"`
	EvidenceKeys []string `json:"evidence_keys"`
}

type CompatibilityRelationshipStrategy struct {
	Communication string `json:"communication"`
	Conflict      string `json:"conflict"`
	Reality       string `json:"reality"`
	Boundary      string `json:"boundary"`
}

type CompatibilityClaimEvidenceLink struct {
	ClaimID      string   `json:"claim_id"`
	Claim        string   `json:"claim"`
	EvidenceKeys []string `json:"evidence_keys"`
	Reasoning    string   `json:"reasoning"`
	Caveat       string   `json:"caveat"`
}

type CompatibilityConsultingAssessment struct {
	RelationshipDiagnosis CompatibilityRelationshipDiagnosis `json:"relationship_diagnosis"`
	DecisionAdvice        CompatibilityDecisionAdvice        `json:"decision_advice"`
	StageRisks            []CompatibilityStageRisk           `json:"stage_risks"`
	RelationshipStrategy  CompatibilityRelationshipStrategy  `json:"relationship_strategy"`
	ClaimEvidenceLinks    []CompatibilityClaimEvidenceLink   `json:"claim_evidence_links"`
}

type CompatibilityDimensionScores struct {
	Attraction    int `json:"attraction"`
	Stability     int `json:"stability"`
	Communication int `json:"communication"`
	Practicality  int `json:"practicality"`
}

type CompatibilityDurationWindow struct {
	Level CompatibilityDurationLevel `json:"level"`
}

type CompatibilityDurationWindows struct {
	ThreeMonths  CompatibilityDurationWindow `json:"three_months"`
	OneYear      CompatibilityDurationWindow `json:"one_year"`
	TwoYearsPlus CompatibilityDurationWindow `json:"two_years_plus"`
}

type CompatibilityDurationAssessment struct {
	OverallBand string                       `json:"overall_band"`
	Windows     CompatibilityDurationWindows `json:"windows"`
	Summary     string                       `json:"summary"`
	Reasons     []string                     `json:"reasons"`
}

type CompatibilityScoreExplanation struct {
	Dimension            CompatibilityDimension `json:"dimension"`
	PositiveFactor       string                 `json:"positive_factor,omitempty"`
	NegativeFactor       string                 `json:"negative_factor,omitempty"`
	PositiveEvidenceKeys []string               `json:"positive_evidence_keys,omitempty"`
	NegativeEvidenceKeys []string               `json:"negative_evidence_keys,omitempty"`
	Summary              string                 `json:"summary"`
}

type CompatibilityAnalysis struct {
	OverallLevel         CompatibilityLevel                `json:"overall_level"`
	DimensionScores      CompatibilityDimensionScores      `json:"dimension_scores"`
	Evidences            []CompatibilityEvidence           `json:"evidences"`
	ScoreExplanations    []CompatibilityScoreExplanation   `json:"score_explanations"`
	SummaryTags          []string                          `json:"summary_tags"`
	DurationAssessment   CompatibilityDurationAssessment   `json:"duration_assessment"`
	ConsultingAssessment CompatibilityConsultingAssessment `json:"consulting_assessment"`
}

const (
	compatibilitySourceDayMaster    = "day_master"
	compatibilitySourceFiveElements = "five_elements"
	compatibilitySourceSpousePalace = "spouse_palace"
	compatibilitySourceSpouseStar   = "spouse_star"
	compatibilitySourceGanZhi       = "ganzhi"
	compatibilitySourceShensha      = "shensha"

	compatibilitySourceTenGodInteraction       = "ten_god_interaction"
	compatibilitySourceFavorableElementSupport = "favorable_element_support"
	compatibilitySourceGanZhiInteraction       = "ganzhi_interaction"
	compatibilitySourceRelationshipPattern     = "relationship_pattern"
	compatibilitySourceTimingContext           = "timing_context"

	compatibilityPerspectiveSelfToPartner = "self_to_partner"
	compatibilityPerspectivePartnerToSelf = "partner_to_self"
	compatibilityPerspectiveMutual        = "mutual"
	compatibilityActorSelf                = "self"
	compatibilityActorPartner             = "partner"
)

type compatibilityAnalysisBuilder struct {
	scores              CompatibilityDimensionScores
	evidences           []CompatibilityEvidence
	sourceContributions map[string]int
}

func newCompatibilityAnalysisBuilder() *compatibilityAnalysisBuilder {
	return &compatibilityAnalysisBuilder{
		scores: CompatibilityDimensionScores{
			Attraction:    60,
			Stability:     60,
			Communication: 60,
			Practicality:  60,
		},
		evidences:           make([]CompatibilityEvidence, 0, 8),
		sourceContributions: map[string]int{},
	}
}

func (builder *compatibilityAnalysisBuilder) addEvidence(item CompatibilityEvidence) {
	item.Weight = builder.capEvidenceWeight(item)
	builder.evidences = append(builder.evidences, item)
	switch item.Dimension {
	case CompatibilityAttraction:
		builder.scores.Attraction = clampScore(builder.scores.Attraction + item.Weight)
	case CompatibilityStability:
		builder.scores.Stability = clampScore(builder.scores.Stability + item.Weight)
	case CompatibilityCommunication:
		builder.scores.Communication = clampScore(builder.scores.Communication + item.Weight)
	case CompatibilityPracticality:
		builder.scores.Practicality = clampScore(builder.scores.Practicality + item.Weight)
	}
}

func (builder *compatibilityAnalysisBuilder) capEvidenceWeight(item CompatibilityEvidence) int {
	if item.Weight == 0 || item.Source == "" || item.Dimension == "" {
		return item.Weight
	}
	sign := "positive"
	if item.Weight < 0 {
		sign = "negative"
	}
	key := compatibilityContributionKey(item.Source, item.Dimension, sign)
	cap := compatibilitySourceContributionCap(item.Source, item.Dimension)
	used := builder.sourceContributions[key]
	remaining := cap - used
	if remaining <= 0 {
		return 0
	}
	absWeight := absInt(item.Weight)
	if absWeight > remaining {
		absWeight = remaining
	}
	builder.sourceContributions[key] = used + absWeight
	if item.Weight < 0 {
		return -absWeight
	}
	return absWeight
}

func compatibilityContributionKey(source string, dimension CompatibilityDimension, sign string) string {
	return fmt.Sprintf("%s:%s:%s", source, dimension, sign)
}

func compatibilitySourceContributionCap(source string, dimension CompatibilityDimension) int {
	sourceCaps := map[string]int{
		compatibilitySourceDayMaster:               18,
		compatibilitySourceFiveElements:            16,
		compatibilitySourceSpousePalace:            28,
		compatibilitySourceSpouseStar:              20,
		compatibilitySourceGanZhi:                  16,
		compatibilitySourceShensha:                 10,
		compatibilitySourceTenGodInteraction:       18,
		compatibilitySourceFavorableElementSupport: 16,
		compatibilitySourceGanZhiInteraction:       22,
		compatibilitySourceRelationshipPattern:     12,
		compatibilitySourceTimingContext:           10,
	}
	cap, ok := sourceCaps[source]
	if !ok {
		return 12
	}
	if source == compatibilitySourceGanZhiInteraction && dimension == CompatibilityStability {
		return cap + 4
	}
	return cap
}

func AnalyzeCompatibility(a, b *BaziResult) CompatibilityAnalysis {
	builder := newCompatibilityAnalysisBuilder()

	buildDayMasterSignals(a, b, builder.addEvidence)
	buildFiveElementSignals(a, b, builder.addEvidence)
	buildFavorableElementSupportSignals(a, b, builder.addEvidence)
	buildSpousePalaceSignals(a, b, builder.addEvidence)
	buildSpouseStarSignals(a, b, builder.addEvidence)
	buildTenGodInteractionSignals(a, b, builder.addEvidence)
	buildGanZhiInteractionSignals(a, b, builder.addEvidence)
	buildGanZhiVolumeSignals(a, b, builder.addEvidence)
	buildShenshaSignals(a, b, builder.addEvidence)
	buildRelationshipPatternSignals(builder)

	sort.SliceStable(builder.evidences, func(i, j int) bool {
		wi, wj := absInt(builder.evidences[i].Weight), absInt(builder.evidences[j].Weight)
		if wi == wj {
			return builder.evidences[i].Type < builder.evidences[j].Type
		}
		return wi > wj
	})
	assignCompatibilityEvidenceKeys(builder.evidences)

	tags := buildCompatibilityTags(builder.scores, builder.evidences)
	scoreExplanations := buildCompatibilityScoreExplanations(builder.evidences)
	duration := buildDurationAssessment(builder.scores, builder.evidences)
	consulting := buildCompatibilityConsultingAssessment(builder.scores, builder.evidences, duration)

	return CompatibilityAnalysis{
		OverallLevel:         aggregateCompatibilityLevel(builder.scores, builder.evidences),
		DimensionScores:      builder.scores,
		Evidences:            builder.evidences,
		ScoreExplanations:    scoreExplanations,
		SummaryTags:          tags,
		DurationAssessment:   duration,
		ConsultingAssessment: consulting,
	}
}

type compatibilityEvidenceSink func(CompatibilityEvidence)

func buildDayMasterSignals(a, b *BaziResult, addEvidence compatibilityEvidenceSink) {
	switch relation := dayMasterRelation(a.DayGanWuxing, b.DayGanWuxing); relation {
	case "same":
		addEvidence(CompatibilityEvidence{
			Dimension: CompatibilityCommunication,
			Type:      "日主同气",
			Polarity:  CompatibilityPositive,
			Source:    compatibilitySourceDayMaster,
			Title:     "日主同气",
			Detail:    fmt.Sprintf("双方日主同属%s，底层气质相近，较容易理解彼此节奏。", a.DayGanWuxing),
			Weight:    8,
		})
	case "sheng":
		addEvidence(CompatibilityEvidence{
			Dimension: CompatibilityCommunication,
			Type:      "日主相生",
			Polarity:  CompatibilityPositive,
			Source:    compatibilitySourceDayMaster,
			Title:     "日主相生",
			Detail:    fmt.Sprintf("双方日主五行形成相生关系（%s / %s），沟通与情感流动更顺。", a.DayGanWuxing, b.DayGanWuxing),
			Weight:    10,
		})
		addEvidence(CompatibilityEvidence{
			Dimension: CompatibilityAttraction,
			Type:      "日主相生",
			Polarity:  CompatibilityPositive,
			Source:    compatibilitySourceDayMaster,
			Title:     "气机互引",
			Detail:    "双方气机能够互相牵引，初始吸引感更容易建立。",
			Weight:    6,
		})
	case "ke":
		addEvidence(CompatibilityEvidence{
			Dimension: CompatibilityCommunication,
			Type:      "日主相克",
			Polarity:  CompatibilityNegative,
			Source:    compatibilitySourceDayMaster,
			Title:     "日主相克",
			Detail:    fmt.Sprintf("双方日主五行互见克制（%s / %s），表达方式和主导感较容易顶撞。", a.DayGanWuxing, b.DayGanWuxing),
			Weight:    -12,
		})
		addEvidence(CompatibilityEvidence{
			Dimension: CompatibilityStability,
			Type:      "日主相克",
			Polarity:  CompatibilityNegative,
			Source:    compatibilitySourceDayMaster,
			Title:     "关系拉扯",
			Detail:    "若缺少其他缓冲结构，长期相处更容易形成拉扯。",
			Weight:    -8,
		})
	}
}

func buildFiveElementSignals(a, b *BaziResult, addEvidence compatibilityEvidenceSink) {
	if detail, ok, weight := fiveElementComplement(a, b); ok {
		addEvidence(CompatibilityEvidence{
			Dimension: CompatibilityPracticality,
			Type:      "五行互补",
			Polarity:  CompatibilityPositive,
			Source:    compatibilitySourceFiveElements,
			Title:     "五行互补",
			Detail:    detail,
			Weight:    weight,
		})
	} else if detail, ok, weight := fiveElementImbalance(a, b); ok {
		addEvidence(CompatibilityEvidence{
			Dimension: CompatibilityPracticality,
			Type:      "五行失衡",
			Polarity:  CompatibilityNegative,
			Source:    compatibilitySourceFiveElements,
			Title:     "五行失衡",
			Detail:    detail,
			Weight:    weight,
		})
	}
}

func buildFavorableElementSupportSignals(a, b *BaziResult, addEvidence compatibilityEvidenceSink) {
	buildDirectionalFavorableElementSignal(a, b, compatibilityPerspectiveSelfToPartner, compatibilityActorSelf, compatibilityActorPartner, addEvidence)
	buildDirectionalFavorableElementSignal(b, a, compatibilityPerspectivePartnerToSelf, compatibilityActorPartner, compatibilityActorSelf, addEvidence)
}

func buildDirectionalFavorableElementSignal(observer, target *BaziResult, perspective, actor, targetRole string, addEvidence compatibilityEvidenceSink) {
	if observer == nil || target == nil {
		return
	}
	targetStats := compatibilityStatsCN(target)
	supports := suppliedElementsFromPhrase(observer.Yongshen, targetStats)
	pressures := suppliedElementsFromPhrase(observer.Jishen, targetStats)
	if len(supports) > 0 {
		addEvidence(CompatibilityEvidence{
			Dimension:   CompatibilityPracticality,
			Type:        "五行支持-喜用补足",
			Polarity:    CompatibilityPositive,
			Source:      compatibilitySourceFavorableElementSupport,
			Perspective: perspective,
			Actor:       actor,
			Target:      targetRole,
			Title:       "补足倾向",
			Detail:      fmt.Sprintf("对方盘面明显带有%s，能补到你命局中的喜用倾向，关系中更容易形成支持与承接。", strings.Join(supports, "、")),
			Weight:      8,
		})
	}
	if len(pressures) > 0 {
		addEvidence(CompatibilityEvidence{
			Dimension:   CompatibilityPracticality,
			Type:        "五行支持-忌神加压",
			Polarity:    CompatibilityNegative,
			Source:      compatibilitySourceFavorableElementSupport,
			Perspective: perspective,
			Actor:       actor,
			Target:      targetRole,
			Title:       "压力倾向",
			Detail:      fmt.Sprintf("对方盘面明显带有%s，容易触发你命局中的忌神压力，现实磨合中需留意同类议题被放大。", strings.Join(pressures, "、")),
			Weight:      -8,
		})
	}
	if observer.Yongshen != "" || observer.Jishen != "" || len(supports) > 0 || len(pressures) > 0 {
		return
	}
	observerStats := compatibilityStatsCN(observer)
	missing := missingElements(observerStats)
	fallbackSupports := suppliedMissingElements(missing, targetStats)
	if len(fallbackSupports) == 0 {
		return
	}
	addEvidence(CompatibilityEvidence{
		Dimension:   CompatibilityPracticality,
		Type:        "五行支持-结构互补",
		Polarity:    CompatibilityPositive,
		Source:      compatibilitySourceFavorableElementSupport,
		Perspective: perspective,
		Actor:       actor,
		Target:      targetRole,
		Title:       "结构互补倾向",
		Detail:      fmt.Sprintf("未使用确定用神结论，仅从五行结构看，对方盘面带有你较少见的%s，存在互补倾向。", strings.Join(fallbackSupports, "、")),
		Weight:      5,
	})
}

func suppliedElementsFromPhrase(phrase string, supplierStats map[string]int) []string {
	if phrase == "" {
		return nil
	}
	out := []string{}
	for _, wx := range []string{"木", "火", "土", "金", "水"} {
		if strings.Contains(phrase, wx) && supplierStats[wx] > 0 {
			out = append(out, wx)
		}
	}
	return out
}

func suppliedMissingElements(missing []string, supplierStats map[string]int) []string {
	out := []string{}
	for _, wx := range missing {
		if supplierStats[wx] > 0 {
			out = append(out, wx)
		}
	}
	return out
}

func buildSpousePalaceSignals(a, b *BaziResult, addEvidence compatibilityEvidenceSink) {
	if sixHe[a.DayZhi] == b.DayZhi {
		addEvidence(CompatibilityEvidence{
			Dimension: CompatibilityStability,
			Type:      "夫妻宫六合",
			Polarity:  CompatibilityPositive,
			Source:    compatibilitySourceSpousePalace,
			Title:     "夫妻宫六合",
			Detail:    fmt.Sprintf("双方日支 %s / %s 构成六合，婚恋靠近感和关系推进意愿更强。", a.DayZhi, b.DayZhi),
			Weight:    18,
		})
		addEvidence(CompatibilityEvidence{
			Dimension: CompatibilityAttraction,
			Type:      "夫妻宫六合",
			Polarity:  CompatibilityPositive,
			Source:    compatibilitySourceSpousePalace,
			Title:     "亲近感强",
			Detail:    "日支六合通常会放大彼此愿意靠近与接纳的倾向。",
			Weight:    8,
		})
	}
	if sixChong[a.DayZhi] == b.DayZhi {
		addEvidence(CompatibilityEvidence{
			Dimension: CompatibilityStability,
			Type:      "夫妻宫六冲",
			Polarity:  CompatibilityNegative,
			Source:    compatibilitySourceSpousePalace,
			Title:     "夫妻宫六冲",
			Detail:    fmt.Sprintf("双方日支 %s / %s 形成六冲，亲密关系中的节奏和需求容易正面碰撞。", a.DayZhi, b.DayZhi),
			Weight:    -18,
		})
		addEvidence(CompatibilityEvidence{
			Dimension: CompatibilityCommunication,
			Type:      "夫妻宫六冲",
			Polarity:  CompatibilityNegative,
			Source:    compatibilitySourceSpousePalace,
			Title:     "情绪冲撞",
			Detail:    "日支六冲往往让相处中的情绪起伏更大，争执成本更高。",
			Weight:    -8,
		})
	}
	if sixXing[a.DayZhi] == b.DayZhi || sixXing[b.DayZhi] == a.DayZhi {
		addEvidence(CompatibilityEvidence{
			Dimension: CompatibilityStability,
			Type:      "夫妻宫刑害",
			Polarity:  CompatibilityNegative,
			Source:    compatibilitySourceSpousePalace,
			Title:     "夫妻宫刑害",
			Detail:    fmt.Sprintf("双方日支 %s / %s 存在刑害结构，关系中更容易积累别扭和隐性损耗。", a.DayZhi, b.DayZhi),
			Weight:    -10,
		})
	}
}

func buildSpouseStarSignals(a, b *BaziResult, addEvidence compatibilityEvidenceSink) {
	if spouseDetail, ok := spouseStarResonance(a, b); ok {
		addEvidence(CompatibilityEvidence{
			Dimension: CompatibilityAttraction,
			Type:      "配偶星呼应",
			Polarity:  CompatibilityPositive,
			Source:    compatibilitySourceSpouseStar,
			Title:     "配偶星呼应",
			Detail:    spouseDetail,
			Weight:    14,
		})
		addEvidence(CompatibilityEvidence{
			Dimension: CompatibilityStability,
			Type:      "配偶星呼应",
			Polarity:  CompatibilityPositive,
			Source:    compatibilitySourceSpouseStar,
			Title:     "婚恋指向明确",
			Detail:    "一方命盘中的婚恋星被对方显著触发，关系更容易朝伴侣关系理解。",
			Weight:    6,
		})
	}
}

func buildTenGodInteractionSignals(a, b *BaziResult, addEvidence compatibilityEvidenceSink) {
	buildDirectionalTenGodInteractionSignal(a, b, compatibilityPerspectiveSelfToPartner, compatibilityActorSelf, compatibilityActorPartner, addEvidence)
	buildDirectionalTenGodInteractionSignal(b, a, compatibilityPerspectivePartnerToSelf, compatibilityActorPartner, compatibilityActorSelf, addEvidence)
}

func buildDirectionalTenGodInteractionSignal(observer, target *BaziResult, perspective, actor, targetRole string, addEvidence compatibilityEvidenceSink) {
	if observer == nil || target == nil || observer.DayGan == "" || target.DayGan == "" {
		return
	}
	tenGod := GetShiShen(observer.DayGan, target.DayGan)
	info, ok := TenGodGroupOf(tenGod)
	if !ok {
		return
	}
	dimension, polarity, title, weight := tenGodCompatibilityMeaning(info.Group)
	if title == "" {
		return
	}
	addEvidence(CompatibilityEvidence{
		Dimension:   dimension,
		Type:        "十神互动-" + info.Label,
		Polarity:    polarity,
		Source:      compatibilitySourceTenGodInteraction,
		Perspective: perspective,
		Actor:       actor,
		Target:      targetRole,
		Title:       title,
		Detail:      fmt.Sprintf("以%s日主观之，对方日主%s落为%s，关系中容易呈现%s。", observer.DayGan, target.DayGan, tenGod, tenGodCompatibilityPlainText(info.Group)),
		Weight:      weight,
	})
}

func tenGodCompatibilityMeaning(group string) (CompatibilityDimension, CompatibilityPolarity, string, int) {
	switch group {
	case TenGodGroupSeal:
		return CompatibilityCommunication, CompatibilityPositive, "支持与照拂感", 6
	case TenGodGroupOfficial:
		return CompatibilityStability, CompatibilityMixed, "责任与压力感", 5
	case TenGodGroupWealth:
		return CompatibilityAttraction, CompatibilityPositive, "现实吸引与投入感", 7
	case TenGodGroupOutput:
		return CompatibilityCommunication, CompatibilityPositive, "表达与被激发感", 6
	case TenGodGroupPeer:
		return CompatibilityCommunication, CompatibilityMixed, "同频与竞争感", 4
	default:
		return "", CompatibilityNeutral, "", 0
	}
}

func tenGodCompatibilityPlainText(group string) string {
	switch group {
	case TenGodGroupSeal:
		return "被理解、被支持、被照顾的感受"
	case TenGodGroupOfficial:
		return "责任、规则、承诺或被要求的压力"
	case TenGodGroupWealth:
		return "现实投入、资源牵引和伴侣吸引"
	case TenGodGroupOutput:
		return "表达欲、分享欲和情绪外放"
	case TenGodGroupPeer:
		return "相似节奏、同频感，也伴随比较和竞争"
	default:
		return "关系互动倾向"
	}
}

func buildGanZhiVolumeSignals(a, b *BaziResult, addEvidence compatibilityEvidenceSink) {
	if zhiClashCount := branchClashCount(a, b); zhiClashCount >= 2 {
		addEvidence(CompatibilityEvidence{
			Dimension: CompatibilityPracticality,
			Type:      "干支冲克偏多",
			Polarity:  CompatibilityNegative,
			Source:    compatibilitySourceGanZhi,
			Title:     "冲克偏多",
			Detail:    fmt.Sprintf("双方盘面存在 %d 处以上明显地支冲克，现实层面的磨合成本会更高。", zhiClashCount),
			Weight:    -10,
		})
	}
}

type compatibilityPillarPoint struct {
	Name   string
	Label  string
	Gan    string
	Zhi    string
	Weight int
}

func buildGanZhiInteractionSignals(a, b *BaziResult, addEvidence compatibilityEvidenceSink) {
	left := compatibilityPillars(a)
	right := compatibilityPillars(b)
	pairs := make([][2]compatibilityPillarPoint, 0, len(left)*len(right))
	for _, lp := range left {
		for _, rp := range right {
			pairs = append(pairs, [2]compatibilityPillarPoint{lp, rp})
		}
	}
	sort.SliceStable(pairs, func(i, j int) bool {
		return compatibilityInteractionWeight(pairs[i][0], pairs[i][1]) > compatibilityInteractionWeight(pairs[j][0], pairs[j][1])
	})
	for _, pair := range pairs {
		buildStemInteractionSignal(pair[0], pair[1], addEvidence)
		buildBranchInteractionSignal(pair[0], pair[1], addEvidence)
	}
}

func compatibilityPillars(r *BaziResult) []compatibilityPillarPoint {
	return []compatibilityPillarPoint{
		{Name: "year", Label: "年柱", Gan: r.YearGan, Zhi: r.YearZhi, Weight: 3},
		{Name: "month", Label: "月柱", Gan: r.MonthGan, Zhi: r.MonthZhi, Weight: 7},
		{Name: "day", Label: "日柱", Gan: r.DayGan, Zhi: r.DayZhi, Weight: 12},
		{Name: "hour", Label: "时柱", Gan: r.HourGan, Zhi: r.HourZhi, Weight: 5},
	}
}

func buildStemInteractionSignal(left, right compatibilityPillarPoint, addEvidence compatibilityEvidenceSink) {
	if left.Gan == "" || right.Gan == "" {
		return
	}
	hua, ok := ganWuhe[[2]string{left.Gan, right.Gan}]
	if !ok {
		return
	}
	weight := compatibilityInteractionWeight(left, right)
	addEvidence(CompatibilityEvidence{
		Dimension: CompatibilityAttraction,
		Type:      "干支互动-天干五合",
		Polarity:  CompatibilityPositive,
		Source:    compatibilitySourceGanZhiInteraction,
		Title:     "天干相合",
		Detail:    fmt.Sprintf("双方%s%s与%s%s构成天干五合，化%s，容易带来靠近、协商或彼此牵引。", left.Label, left.Gan, right.Label, right.Gan, wxPinyin2CN[hua]),
		Weight:    maxInt(4, weight-2),
	})
}

func buildBranchInteractionSignal(left, right compatibilityPillarPoint, addEvidence compatibilityEvidenceSink) {
	if left.Zhi == "" || right.Zhi == "" {
		return
	}
	weight := compatibilityInteractionWeight(left, right)
	switch {
	case sixHe[left.Zhi] == right.Zhi:
		addEvidence(CompatibilityEvidence{
			Dimension: CompatibilityStability,
			Type:      "干支互动-地支六合",
			Polarity:  CompatibilityPositive,
			Source:    compatibilitySourceGanZhiInteraction,
			Title:     "地支相合",
			Detail:    fmt.Sprintf("双方%s%s与%s%s形成六合，能增加关系中的靠近感和协调空间。", left.Label, left.Zhi, right.Label, right.Zhi),
			Weight:    weight,
		})
	case sixChong[left.Zhi] == right.Zhi:
		addEvidence(CompatibilityEvidence{
			Dimension: CompatibilityStability,
			Type:      "干支互动-地支六冲",
			Polarity:  CompatibilityNegative,
			Source:    compatibilitySourceGanZhiInteraction,
			Title:     "地支相冲",
			Detail:    fmt.Sprintf("双方%s%s与%s%s形成六冲，容易在对应生活层面出现节奏碰撞。", left.Label, left.Zhi, right.Label, right.Zhi),
			Weight:    -weight,
		})
	case sixXing[left.Zhi] == right.Zhi || sixXing[right.Zhi] == left.Zhi || sixHai[left.Zhi] == right.Zhi:
		addEvidence(CompatibilityEvidence{
			Dimension: CompatibilityCommunication,
			Type:      "干支互动-地支刑害",
			Polarity:  CompatibilityNegative,
			Source:    compatibilitySourceGanZhiInteraction,
			Title:     "刑害暗耗",
			Detail:    fmt.Sprintf("双方%s%s与%s%s存在刑害结构，容易形成误读、别扭或隐性消耗。", left.Label, left.Zhi, right.Label, right.Zhi),
			Weight:    -maxInt(4, weight-3),
		})
	case sameJuGroup(left.Zhi, right.Zhi, sanheGroups):
		addEvidence(CompatibilityEvidence{
			Dimension: CompatibilityAttraction,
			Type:      "干支互动-地支半合",
			Polarity:  CompatibilityPositive,
			Source:    compatibilitySourceGanZhiInteraction,
			Title:     "地支半合",
			Detail:    fmt.Sprintf("双方%s%s与%s%s落在同一三合结构，容易形成共同兴趣或情绪牵引。", left.Label, left.Zhi, right.Label, right.Zhi),
			Weight:    maxInt(3, weight/2),
		})
	case sameJuGroup(left.Zhi, right.Zhi, sanhuiGroups):
		addEvidence(CompatibilityEvidence{
			Dimension: CompatibilityPracticality,
			Type:      "干支互动-地支半会",
			Polarity:  CompatibilityMixed,
			Source:    compatibilitySourceGanZhiInteraction,
			Title:     "地支半会",
			Detail:    fmt.Sprintf("双方%s%s与%s%s落在同一三会结构，容易在现实节奏或外部环境上互相带动。", left.Label, left.Zhi, right.Label, right.Zhi),
			Weight:    maxInt(3, weight/2),
		})
	}
}

func compatibilityInteractionWeight(left, right compatibilityPillarPoint) int {
	if left.Name == "day" && right.Name == "day" {
		return 12
	}
	return maxInt(3, (left.Weight+right.Weight)/2)
}

func sameJuGroup(left, right string, groups [4][3]string) bool {
	if left == right {
		return false
	}
	for _, group := range groups {
		hasLeft, hasRight := false, false
		for _, zhi := range group {
			if zhi == left {
				hasLeft = true
			}
			if zhi == right {
				hasRight = true
			}
		}
		if hasLeft && hasRight {
			return true
		}
	}
	return false
}

func buildShenshaSignals(a, b *BaziResult, addEvidence compatibilityEvidenceSink) {
	if hasRomanceShensha(a) || hasRomanceShensha(b) {
		addEvidence(CompatibilityEvidence{
			Dimension: CompatibilityAttraction,
			Type:      "桃花助缘",
			Polarity:  CompatibilityPositive,
			Source:    compatibilitySourceShensha,
			Title:     "桃花助缘",
			Detail:    "盘中带有桃花/天喜/红艳类神煞，能增强关系中的浪漫和被吸引感。",
			Weight:    6,
		})
	}
	if hasLonelyShensha(a) || hasLonelyShensha(b) {
		addEvidence(CompatibilityEvidence{
			Dimension: CompatibilityStability,
			Type:      "孤寡错位",
			Polarity:  CompatibilityNegative,
			Source:    compatibilitySourceShensha,
			Title:     "孤寡错位",
			Detail:    "盘中带有孤辰/寡宿类神煞，容易放大距离感、误读或情绪疏离。",
			Weight:    -6,
		})
	}
}

func buildRelationshipPatternSignals(builder *compatibilityAnalysisBuilder) {
	negativeCommunication := countCompatibilityEvidence(builder.evidences, func(item CompatibilityEvidence) bool {
		return item.Polarity == CompatibilityNegative &&
			(item.Dimension == CompatibilityCommunication || item.Dimension == CompatibilityStability)
	})
	positivePracticality := countCompatibilityEvidence(builder.evidences, func(item CompatibilityEvidence) bool {
		return item.Polarity == CompatibilityPositive &&
			(item.Dimension == CompatibilityPracticality || item.Source == compatibilitySourceFavorableElementSupport)
	})
	positiveAttraction := countCompatibilityEvidence(builder.evidences, func(item CompatibilityEvidence) bool {
		return item.Polarity == CompatibilityPositive && item.Dimension == CompatibilityAttraction
	})
	negativeStability := countCompatibilityEvidence(builder.evidences, func(item CompatibilityEvidence) bool {
		return item.Polarity == CompatibilityNegative && item.Dimension == CompatibilityStability
	})

	if negativeCommunication >= 2 {
		builder.addEvidence(CompatibilityEvidence{
			Dimension:      CompatibilityCommunication,
			Type:           "关系模式-冲突触发",
			Polarity:       CompatibilityNegative,
			Source:         compatibilitySourceRelationshipPattern,
			Title:          "冲突触发点清晰",
			Detail:         "多组稳定或沟通证据同时指向节奏碰撞，关系中需要把冲突触发点提前具体化。",
			RelatedSources: relatedCompatibilitySources(builder.evidences, CompatibilityNegative),
			Weight:         -6,
		})
	}
	if positiveAttraction >= 3 && negativeStability >= 1 {
		builder.addEvidence(CompatibilityEvidence{
			Dimension:      CompatibilityStability,
			Type:           "关系模式-高吸引高波动",
			Polarity:       CompatibilityMixed,
			Source:         compatibilitySourceRelationshipPattern,
			Title:          "吸引强但波动也强",
			Detail:         "吸引证据较多，但稳定压力也同步存在，适合把热度和长期承接分开判断。",
			RelatedSources: relatedCompatibilitySources(builder.evidences, CompatibilityMixed),
			Weight:         -4,
		})
	}
	if positivePracticality >= 2 {
		builder.addEvidence(CompatibilityEvidence{
			Dimension:      CompatibilityPracticality,
			Type:           "关系模式-现实承接",
			Polarity:       CompatibilityPositive,
			Source:         compatibilitySourceRelationshipPattern,
			Title:          "现实承接点较多",
			Detail:         "互补或支持证据较多，关系更容易找到现实层面的分工、照应或共同推进点。",
			RelatedSources: relatedCompatibilitySources(builder.evidences, CompatibilityPositive),
			Weight:         5,
		})
	}
}

func countCompatibilityEvidence(evidences []CompatibilityEvidence, match func(CompatibilityEvidence) bool) int {
	count := 0
	for _, item := range evidences {
		if match(item) {
			count++
		}
	}
	return count
}

func relatedCompatibilitySources(evidences []CompatibilityEvidence, polarity CompatibilityPolarity) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, item := range evidences {
		if item.Source == "" || item.Source == compatibilitySourceRelationshipPattern {
			continue
		}
		if polarity != CompatibilityMixed && item.Polarity != polarity {
			continue
		}
		if seen[item.Source] {
			continue
		}
		seen[item.Source] = true
		out = append(out, item.Source)
		if len(out) == 4 {
			return out
		}
	}
	return out
}

func assignCompatibilityEvidenceKeys(evidences []CompatibilityEvidence) {
	counts := map[string]int{}
	for i := range evidences {
		baseKey := buildCompatibilityEvidenceKey(evidences[i])
		counts[baseKey]++
		if counts[baseKey] == 1 {
			evidences[i].EvidenceKey = baseKey
			continue
		}
		evidences[i].EvidenceKey = fmt.Sprintf("%s_%d", baseKey, counts[baseKey])
	}
}

func buildCompatibilityEvidenceKey(item CompatibilityEvidence) string {
	return fmt.Sprintf("%s_%s_%s", item.Source, item.Dimension, compatibilityEvidenceTypeSlug(item.Type))
}

func compatibilityEvidenceTypeSlug(typeName string) string {
	switch typeName {
	case "日主同气":
		return "day_master_same"
	case "日主相生":
		return "day_master_generating"
	case "日主相克":
		return "day_master_controlling"
	case "五行互补":
		return "five_element_complement"
	case "五行失衡":
		return "five_element_imbalance"
	case "夫妻宫六合":
		return "spouse_palace_liuhe"
	case "夫妻宫六冲":
		return "spouse_palace_chong"
	case "夫妻宫刑害":
		return "spouse_palace_xing_hai"
	case "配偶星呼应":
		return "spouse_star_resonance"
	case "干支冲克偏多":
		return "branch_clash_heavy"
	case "桃花助缘":
		return "romance_shensha"
	case "孤寡错位":
		return "lonely_shensha"
	case "十神互动-财星":
		return "ten_god_wealth"
	case "十神互动-官杀":
		return "ten_god_official"
	case "十神互动-印星":
		return "ten_god_seal"
	case "十神互动-食伤":
		return "ten_god_output"
	case "十神互动-比劫":
		return "ten_god_peer"
	case "五行支持-喜用补足":
		return "favorable_element_support"
	case "五行支持-忌神加压":
		return "favorable_element_pressure"
	case "五行支持-结构互补":
		return "favorable_element_structural"
	case "干支互动-天干五合":
		return "ganzhi_stem_combo"
	case "干支互动-地支六合":
		return "ganzhi_branch_liuhe"
	case "干支互动-地支六冲":
		return "ganzhi_branch_chong"
	case "干支互动-地支刑害":
		return "ganzhi_branch_xing_hai"
	case "干支互动-地支半合":
		return "ganzhi_branch_half_combo"
	case "干支互动-地支半会":
		return "ganzhi_branch_half_meeting"
	case "关系模式-冲突触发":
		return "relationship_pattern_conflict_trigger"
	case "关系模式-高吸引高波动":
		return "relationship_pattern_attraction_volatility"
	case "关系模式-现实承接":
		return "relationship_pattern_reality_support"
	default:
		return "unknown"
	}
}

func buildCompatibilityConsultingAssessment(scores CompatibilityDimensionScores, evidences []CompatibilityEvidence, duration CompatibilityDurationAssessment) CompatibilityConsultingAssessment {
	negativeKeys := topEvidenceKeys(evidences, CompatibilityNegative, 2)
	positiveKeys := topEvidenceKeys(evidences, CompatibilityPositive, 2)
	primaryKeys := append([]string{}, positiveKeys...)
	primaryKeys = append(primaryKeys, negativeKeys...)
	if len(primaryKeys) == 0 && len(evidences) > 0 && evidences[0].EvidenceKey != "" {
		primaryKeys = []string{evidences[0].EvidenceKey}
	}

	relationshipType := "均衡观察型"
	if scores.Attraction >= 70 && scores.Stability < 60 {
		relationshipType = "短期吸引强、长期承压型"
	} else if scores.Stability >= 70 && scores.Practicality >= 65 {
		relationshipType = "稳定经营型"
	} else if scores.Attraction >= 70 && scores.Communication >= 65 {
		relationshipType = "高吸引互动型"
	} else if scores.Practicality < 55 || scores.Stability < 55 {
		relationshipType = "高磨合成本型"
	}

	recommendation := "observe"
	verdict := "建议谨慎观察"
	if scores.Stability >= 68 && scores.Practicality >= 62 && len(negativeKeys) == 0 {
		recommendation = "continue"
		verdict = "适合继续推进"
	} else if scores.Stability < 52 || scores.Practicality < 50 {
		recommendation = "caution"
		verdict = "不宜过早重投入"
	}

	confidence := "medium"
	if absInt(scores.Attraction-scores.Stability) >= 22 || len(evidences) >= 5 {
		confidence = "high"
	}

	topFindings := []CompatibilityFinding{
		{Text: "关系优势与风险需要分开判断，不能只看短期吸引。", EvidenceKeys: primaryKeys},
	}
	if scores.Attraction >= 70 {
		topFindings = append(topFindings, CompatibilityFinding{Text: "双方存在较明显的靠近感和吸引支点。", EvidenceKeys: positiveKeys})
	}
	if scores.Stability < 60 || scores.Practicality < 60 {
		topFindings = append(topFindings, CompatibilityFinding{Text: "长期稳定更依赖沟通规则和现实安排。", EvidenceKeys: negativeKeys})
	}
	if len(topFindings) > 3 {
		topFindings = topFindings[:3]
	}

	claimEvidenceLinks := []CompatibilityClaimEvidenceLink{}
	if hasNonEmptyCompatibilityEvidenceKey(primaryKeys) {
		claimEvidenceLinks = append(claimEvidenceLinks, CompatibilityClaimEvidenceLink{
			ClaimID:      "relationship_main_judgement",
			Claim:        verdict,
			EvidenceKeys: primaryKeys,
			Reasoning:    "主要判断来自吸引、稳定、沟通和现实磨合四类证据的合并结果。",
			Caveat:       "合盘表达的是关系倾向，现实选择和相处方式会改变结果表现。",
		})
	}

	return CompatibilityConsultingAssessment{
		RelationshipDiagnosis: CompatibilityRelationshipDiagnosis{
			RelationshipType: relationshipType,
			Verdict:          verdict,
			Summary:          compatibilityConsultingSummary(relationshipType, recommendation),
			TopFindings:      topFindings,
		},
		DecisionAdvice: CompatibilityDecisionAdvice{
			Recommendation: recommendation,
			Confidence:     confidence,
			Conditions:     []string{"先观察冲突后的修复能力", "把现实安排和投入节奏说清楚"},
			DoNext:         []string{"用一到两个月验证沟通节奏是否稳定", "把容易争执的问题具体化处理"},
			Avoid:          []string{"用短期吸引替代长期判断", "在关系规则未稳定前过早绑定重大决定"},
		},
		StageRisks: []CompatibilityStageRisk{
			buildCompatibilityStageRisk("three_months", duration.Windows.ThreeMonths.Level, "初期热度与节奏差异", "一方推进过快或回应不稳定时", "先约定沟通频率和边界。", primaryKeys),
			buildCompatibilityStageRisk("one_year", duration.Windows.OneYear.Level, "现实磨合和冲突修复", "生活安排、承诺节奏或家庭压力进入关系时", "把分歧拆成具体事项，不用情绪判断关系本身。", primaryKeys),
			buildCompatibilityStageRisk("two_years_plus", duration.Windows.TwoYearsPlus.Level, "长期稳定和责任承接", "长期规划、责任分工和资源投入需要落地时", "建立可持续的责任分工和共同计划。", primaryKeys),
		},
		RelationshipStrategy: CompatibilityRelationshipStrategy{
			Communication: "重要议题用明确约定替代情绪试探。",
			Conflict:      "争执时先暂停升级，再回到具体事件和责任分工。",
			Reality:       "长期计划需要拆成可验证的小步骤。",
			Boundary:      "初期保留个人节奏，避免过快形成单方依赖。",
		},
		ClaimEvidenceLinks: claimEvidenceLinks,
	}
}

func buildCompatibilityScoreExplanations(evidences []CompatibilityEvidence) []CompatibilityScoreExplanation {
	dimensions := []CompatibilityDimension{
		CompatibilityAttraction,
		CompatibilityStability,
		CompatibilityCommunication,
		CompatibilityPracticality,
	}
	out := make([]CompatibilityScoreExplanation, 0, len(dimensions))
	for _, dimension := range dimensions {
		positive := strongestCompatibilityEvidence(evidences, dimension, CompatibilityPositive)
		negative := strongestCompatibilityEvidence(evidences, dimension, CompatibilityNegative)
		explanation := CompatibilityScoreExplanation{Dimension: dimension}
		if positive != nil {
			explanation.PositiveFactor = positive.Title
			explanation.PositiveEvidenceKeys = []string{positive.EvidenceKey}
		}
		if negative != nil {
			explanation.NegativeFactor = negative.Title
			explanation.NegativeEvidenceKeys = []string{negative.EvidenceKey}
		}
		switch {
		case positive != nil && negative != nil:
			explanation.Summary = fmt.Sprintf("%s同时受%s支撑、受%s牵制，需要看实际相处能否化解压力。", compatibilityDimensionLabel(dimension), positive.Title, negative.Title)
		case positive != nil:
			explanation.Summary = fmt.Sprintf("%s主要受%s支撑。", compatibilityDimensionLabel(dimension), positive.Title)
		case negative != nil:
			explanation.Summary = fmt.Sprintf("%s主要受%s牵制。", compatibilityDimensionLabel(dimension), negative.Title)
		default:
			explanation.Summary = fmt.Sprintf("%s证据有限，宜结合其他维度保守判断。", compatibilityDimensionLabel(dimension))
		}
		out = append(out, explanation)
	}
	return out
}

func strongestCompatibilityEvidence(evidences []CompatibilityEvidence, dimension CompatibilityDimension, polarity CompatibilityPolarity) *CompatibilityEvidence {
	var best *CompatibilityEvidence
	for i := range evidences {
		item := &evidences[i]
		if item.Dimension != dimension || item.Polarity != polarity || item.Weight == 0 {
			continue
		}
		if best == nil || absInt(item.Weight) > absInt(best.Weight) {
			best = item
		}
	}
	return best
}

func compatibilityDimensionLabel(dimension CompatibilityDimension) string {
	switch dimension {
	case CompatibilityAttraction:
		return "吸引力"
	case CompatibilityStability:
		return "稳定度"
	case CompatibilityCommunication:
		return "沟通修复"
	case CompatibilityPracticality:
		return "现实磨合"
	default:
		return "该维度"
	}
}

func hasNonEmptyCompatibilityEvidenceKey(keys []string) bool {
	for _, key := range keys {
		if key != "" {
			return true
		}
	}
	return false
}

func topEvidenceKeys(evidences []CompatibilityEvidence, polarity CompatibilityPolarity, limit int) []string {
	keys := []string{}
	for _, item := range evidences {
		if item.Polarity == polarity && item.EvidenceKey != "" {
			keys = append(keys, item.EvidenceKey)
			if len(keys) == limit {
				return keys
			}
		}
	}
	return keys
}

func compatibilityConsultingSummary(relationshipType, recommendation string) string {
	switch recommendation {
	case "continue":
		return "这组关系具备继续推进的基础，但仍需要把优势落到稳定沟通和现实安排中。"
	case "caution":
		return "这组关系不宜只凭短期感受快速投入，长期稳定需要先通过现实相处验证。"
	default:
		return fmt.Sprintf("%s需要边推进边观察，重点看冲突修复和现实节奏是否能对齐。", relationshipType)
	}
}

func buildCompatibilityStageRisk(window string, level CompatibilityDurationLevel, mainRisk, trigger, advice string, evidenceKeys []string) CompatibilityStageRisk {
	return CompatibilityStageRisk{
		Window:       window,
		RiskLevel:    string(level),
		MainRisk:     mainRisk,
		Trigger:      trigger,
		Advice:       advice,
		EvidenceKeys: evidenceKeys,
	}
}

func buildDurationAssessment(scores CompatibilityDimensionScores, evidences []CompatibilityEvidence) CompatibilityDurationAssessment {
	shortScore := scores.Attraction*2 + scores.Communication/2
	midScore := scores.Attraction/2 + scores.Stability + scores.Communication/2 + scores.Practicality/2
	longScore := scores.Stability + scores.Practicality*2

	for _, item := range evidences {
		switch item.Source {
		case "spouse_palace":
			shortScore += item.Weight / 3
			midScore += item.Weight
			longScore += item.Weight
		case "spouse_star", "shensha":
			shortScore += item.Weight
			midScore += item.Weight / 2
		case "ganzhi", "five_elements":
			midScore += item.Weight / 2
			longScore += item.Weight
		case "day_master":
			shortScore += item.Weight / 2
			midScore += item.Weight / 2
		}
	}

	shortScore = clampDurationScore(shortScore / 2)
	midScore = clampDurationScore(midScore / 2)
	longScore = clampDurationScore(longScore / 2)

	return CompatibilityDurationAssessment{
		OverallBand: durationBand(longScore),
		Windows: CompatibilityDurationWindows{
			ThreeMonths:  CompatibilityDurationWindow{Level: durationLevel(shortScore)},
			OneYear:      CompatibilityDurationWindow{Level: durationLevel(midScore)},
			TwoYearsPlus: CompatibilityDurationWindow{Level: durationLevel(longScore)},
		},
		Summary: durationSummary(shortScore, midScore, longScore),
		Reasons: durationReasons(evidences),
	}
}

func dayMasterRelation(aWx, bWx string) string {
	if aWx == "" || bWx == "" {
		return ""
	}
	if aWx == bWx {
		return "same"
	}
	if wxSheng[aWx] == bWx || wxSheng[bWx] == aWx {
		return "sheng"
	}
	if wxKe[aWx] == bWx || wxKe[bWx] == aWx {
		return "ke"
	}
	return ""
}

func fiveElementComplement(a, b *BaziResult) (string, bool, int) {
	aStats := compatibilityStats(a)
	bStats := compatibilityStats(b)
	missingA := missingElements(aStats)
	missingB := missingElements(bStats)
	if len(missingA) == 0 || len(missingB) == 0 {
		return "", false, 0
	}
	helpA, helpB := 0, 0
	for _, wx := range missingA {
		if bStats[wx] > 0 {
			helpA++
		}
	}
	for _, wx := range missingB {
		if aStats[wx] > 0 {
			helpB++
		}
	}
	if helpA == 0 && helpB == 0 {
		return "", false, 0
	}
	weight := 8
	if helpA > 0 && helpB > 0 {
		weight = 12
	}
	return fmt.Sprintf("双方五行短板存在互补关系：A 缺 %v、B 可补；B 缺 %v、A 可补。", missingA, missingB), true, weight
}

func fiveElementImbalance(a, b *BaziResult) (string, bool, int) {
	aStats := compatibilityStats(a)
	bStats := compatibilityStats(b)
	sharedMissing := make([]string, 0, 2)
	for _, wx := range []string{"木", "火", "土", "金", "水"} {
		if aStats[wx] == 0 && bStats[wx] == 0 {
			sharedMissing = append(sharedMissing, wx)
		}
	}
	if len(sharedMissing) > 0 {
		return fmt.Sprintf("双方共同缺少 %v，现实协同上容易出现同类短板叠加。", sharedMissing), true, -10
	}
	return "", false, 0
}

func spouseStarResonance(self, partner *BaziResult) (string, bool) {
	targets := []string{}
	if self.Gender == "male" {
		targets = []string{"正财", "偏财"}
	} else {
		targets = []string{"正官", "七杀"}
	}
	for _, gan := range collectCompatibilityGans(partner) {
		shishen := GetShiShen(self.DayGan, gan)
		for _, target := range targets {
			if shishen == target {
				return fmt.Sprintf("以%s日主观之，对方命盘中的 %s 对应你的%s，婚恋指向较明确。", self.DayGan, gan, target), true
			}
		}
	}
	return "", false
}

func branchClashCount(a, b *BaziResult) int {
	az := []string{a.YearZhi, a.MonthZhi, a.DayZhi, a.HourZhi}
	bz := []string{b.YearZhi, b.MonthZhi, b.DayZhi, b.HourZhi}
	count := 0
	for _, left := range az {
		for _, right := range bz {
			if sixChong[left] == right {
				count++
			}
		}
	}
	return count
}

func hasRomanceShensha(r *BaziResult) bool {
	for _, item := range appendShensha(nil, r) {
		if item == "桃花" || item == "天喜" || item == "红艳" {
			return true
		}
	}
	return false
}

func hasLonelyShensha(r *BaziResult) bool {
	for _, item := range appendShensha(nil, r) {
		if item == "孤辰" || item == "寡宿" {
			return true
		}
	}
	return false
}

func appendShensha(dst []string, r *BaziResult) []string {
	dst = append(dst, r.YearShenSha...)
	dst = append(dst, r.MonthShenSha...)
	dst = append(dst, r.DayShenSha...)
	dst = append(dst, r.HourShenSha...)
	return dst
}

func compatibilityStats(r *BaziResult) map[string]int {
	if r.Wuxing.Total > 0 {
		return map[string]int{
			"木": r.Wuxing.Mu,
			"火": r.Wuxing.Huo,
			"土": r.Wuxing.Tu,
			"金": r.Wuxing.Jin,
			"水": r.Wuxing.Shui,
		}
	}
	stats := map[string]int{"木": 0, "火": 0, "土": 0, "金": 0, "水": 0}
	for _, wx := range []string{
		r.YearGanWuxing, r.MonthGanWuxing, r.DayGanWuxing, r.HourGanWuxing,
		r.YearZhiWuxing, r.MonthZhiWuxing, r.DayZhiWuxing, r.HourZhiWuxing,
	} {
		if wx != "" {
			stats[wx]++
		}
	}
	return stats
}

func compatibilityStatsCN(r *BaziResult) map[string]int {
	if r.Wuxing.Total > 0 {
		return map[string]int{
			"木": r.Wuxing.Mu,
			"火": r.Wuxing.Huo,
			"土": r.Wuxing.Tu,
			"金": r.Wuxing.Jin,
			"水": r.Wuxing.Shui,
		}
	}
	stats := map[string]int{"木": 0, "火": 0, "土": 0, "金": 0, "水": 0}
	for _, wx := range []string{
		r.YearGanWuxing, r.MonthGanWuxing, r.DayGanWuxing, r.HourGanWuxing,
		r.YearZhiWuxing, r.MonthZhiWuxing, r.DayZhiWuxing, r.HourZhiWuxing,
	} {
		if cn, ok := wxPinyin2CN[wx]; ok {
			stats[cn]++
			continue
		}
		if _, ok := stats[wx]; ok {
			stats[wx]++
		}
	}
	return stats
}

func missingElements(stats map[string]int) []string {
	out := make([]string, 0, 2)
	for _, wx := range []string{"木", "火", "土", "金", "水"} {
		if stats[wx] == 0 {
			out = append(out, wx)
		}
	}
	return out
}

func collectCompatibilityGans(r *BaziResult) []string {
	gans := []string{r.YearGan, r.MonthGan, r.DayGan, r.HourGan}
	gans = append(gans, r.YearHideGan...)
	gans = append(gans, r.MonthHideGan...)
	gans = append(gans, r.DayHideGan...)
	gans = append(gans, r.HourHideGan...)
	return gans
}

func aggregateCompatibilityLevel(scores CompatibilityDimensionScores, evidences []CompatibilityEvidence) CompatibilityLevel {
	avg := (scores.Attraction + scores.Stability + scores.Communication + scores.Practicality) / 4
	negativeHeavy := 0
	for _, item := range evidences {
		if item.Polarity == CompatibilityNegative && item.Weight <= -12 {
			negativeHeavy++
		}
	}
	if scores.Stability < 50 || (avg < 58 && negativeHeavy > 0) {
		return CompatibilityLow
	}
	if avg >= 70 && scores.Stability >= 65 && negativeHeavy == 0 {
		return CompatibilityHigh
	}
	return CompatibilityMedium
}

func buildCompatibilityTags(scores CompatibilityDimensionScores, evidences []CompatibilityEvidence) []string {
	tags := make([]string, 0, 4)
	if scores.Attraction >= 70 {
		tags = append(tags, "吸引力强")
	}
	if scores.Stability >= 70 {
		tags = append(tags, "稳定度高")
	}
	if scores.Communication >= 70 {
		tags = append(tags, "沟通顺畅")
	}
	if scores.Practicality >= 70 {
		tags = append(tags, "现实互补")
	}
	if scores.Stability < 55 {
		tags = append(tags, "关系波动")
	}
	if scores.Practicality < 55 {
		tags = append(tags, "磨合成本高")
	}
	if len(tags) == 0 && len(evidences) > 0 {
		tags = append(tags, evidences[0].Title)
	}
	if len(tags) > 4 {
		tags = tags[:4]
	}
	return tags
}

func durationLevel(score int) CompatibilityDurationLevel {
	switch {
	case score >= 68:
		return CompatibilityDurationHigh
	case score >= 52:
		return CompatibilityDurationMedium
	default:
		return CompatibilityDurationLow
	}
}

func durationBand(longScore int) string {
	switch durationLevel(longScore) {
	case CompatibilityDurationHigh:
		return "long_term"
	case CompatibilityDurationMedium:
		return "medium_term"
	default:
		return "short_term"
	}
}

func durationSummary(shortScore, midScore, longScore int) string {
	shortLevel := durationLevel(shortScore)
	longLevel := durationLevel(longScore)
	switch {
	case shortLevel == CompatibilityDurationHigh && longLevel == CompatibilityDurationLow:
		return "前期吸引和推进感较强，但长期承压明显，关系更像先热后难。"
	case shortLevel == CompatibilityDurationHigh && longLevel == CompatibilityDurationHigh:
		return "从短期靠近到长期承接都相对顺，关系更有持续发展的空间。"
	case longLevel == CompatibilityDurationHigh:
		return "关系不只靠一时吸引，进入长期后仍有承接和稳定的支点。"
	default:
		return "这段关系的维持性更依赖阶段中的现实磨合，而不只是最初的感觉。"
	}
}

func durationReasons(evidences []CompatibilityEvidence) []string {
	reasons := make([]string, 0, 3)
	for _, item := range evidences {
		switch item.Type {
		case "夫妻宫六合", "夫妻宫六冲", "夫妻宫刑害", "配偶星呼应", "五行互补", "五行失衡", "桃花助缘", "孤寡错位":
			reasons = append(reasons, item.Title+"："+item.Detail)
		}
		if len(reasons) == 3 {
			return reasons
		}
	}
	if len(reasons) == 0 {
		reasons = append(reasons, "当前盘面更适合结合四维分数综合判断阶段维持性。")
	}
	return reasons
}

func clampScore(v int) int {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}

func clampDurationScore(v int) int {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
