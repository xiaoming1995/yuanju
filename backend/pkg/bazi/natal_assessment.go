package bazi

import (
	"fmt"
	"math"
	"strings"
)

// NatalAssessmentVersion identifies the rules used to produce the original-natal
// assessment. Saved charts are lazily upgraded when this value changes.
const NatalAssessmentVersion = "v8"

type NatalEvidence struct {
	RuleID string `json:"rule_id"`
	Source string `json:"source"`
	Label  string `json:"label"`
	Impact string `json:"impact"`
	Delta  int    `json:"delta"`
	Detail string `json:"detail"`
}

type NatalClimateAssessment struct {
	Status           string `json:"status"`
	RequiredElements string `json:"required_elements,omitempty"`
	Score            int    `json:"score"`
	GradeCeiling     string `json:"grade_ceiling"`
}

// NatalDayStemTiaohouAssessment records the dictionary-based requirements for
// a day stem in its month command. It is distinct from seasonal heat/cold.
type NatalDayStemTiaohouAssessment struct {
	Status          string   `json:"status"`
	Formation       string   `json:"formation"`
	FoundationTier  string   `json:"foundation_tier"`
	FoundationScore int      `json:"foundation_score"`
	RequiredStems   []string `json:"required_stems"`
	VisibleStems    []string `json:"visible_stems"`
	HiddenStems     []string `json:"hidden_stems"`
	Score           int      `json:"score"`
	Detail          string   `json:"detail"`
}

// NatalThermalTiaohouAssessment describes the seasonal thermal condition and
// the availability tier of its remedy. Hidden support is always partial.
type NatalThermalTiaohouAssessment struct {
	Status           string   `json:"status"`
	Condition        string   `json:"condition"`
	RequiredElements string   `json:"required_elements,omitempty"`
	VisibleSupport   []string `json:"visible_support"`
	HiddenSupport    []string `json:"hidden_support"`
	Score            int      `json:"score"`
	GradeCeiling     string   `json:"grade_ceiling"`
	Detail           string   `json:"detail"`
}

type NatalTiaohouAssessment struct {
	DayStem NatalDayStemTiaohouAssessment `json:"day_stem"`
	Thermal NatalThermalTiaohouAssessment `json:"thermal"`
}

type NatalFuyiAssessment struct {
	DayMasterStrength string `json:"day_master_strength"`
	Yongshen          string `json:"yongshen"`
	Jishen            string `json:"jishen"`
	SupportLevel      string `json:"support_level"`
	Score             int    `json:"score"`
	StrengthScore     int    `json:"strength_score"`
	Evidence          string `json:"evidence"`
	GradeCeiling      string `json:"grade_ceiling"`
}

// NatalStemGuidanceItem explains a stem-level recommendation without changing
// the underlying element-level Fuyi conclusion.
type NatalStemGuidanceItem struct {
	Stem         string   `json:"stem"`
	Element      string   `json:"element"`
	TenGod       string   `json:"ten_god"`
	SourceLayers []string `json:"source_layers"`
	Detail       string   `json:"detail"`
}

// NatalStemGuidance keeps shared day-stem adjustment/Fuyi support distinct
// from Fuyi-only candidates and adjustment requirements that conflict with it.
type NatalStemGuidance struct {
	PrimaryFavorable   []NatalStemGuidanceItem `json:"primary_favorable"`
	SecondaryFavorable []NatalStemGuidanceItem `json:"secondary_favorable"`
	ConditioningOnly   []NatalStemGuidanceItem `json:"conditioning_only"`
	Adverse            []NatalStemGuidanceItem `json:"adverse"`
}

// NatalYongshenAlignment identifies elements selected independently by
// cold/heat thermal regulation and Fuyi. It is explanatory and never scored.
type NatalYongshenAlignment struct {
	Elements     []string `json:"elements"`
	SourceLayers []string `json:"source_layers"`
	Detail       string   `json:"detail,omitempty"`
}

type NatalPatternAssessment struct {
	Name             string   `json:"name"`
	Quality          string   `json:"quality"`
	Score            int      `json:"score"`
	FoundationSource string   `json:"foundation_source,omitempty"`
	FoundationLabel  string   `json:"foundation_label,omitempty"`
	FoundationTier   string   `json:"foundation_tier,omitempty"`
	Formations       []string `json:"formations"`
	Breaks           []string `json:"breaks"`
	SupportTenGods   []string `json:"support_ten_gods"`
	BreakTenGods     []string `json:"break_ten_gods"`
	CriticalBreak    bool     `json:"critical_break"`
}

type NatalRelationAssessment struct {
	Flow         string   `json:"flow"`
	Score        int      `json:"score"`
	Combinations []string `json:"combinations"`
	Disruptions  []string `json:"disruptions"`
}

type NatalCarryingAssessment struct {
	Level  string `json:"level"`
	Detail string `json:"detail"`
}

type NatalGradeAssessment struct {
	Score        int    `json:"score"`
	Grade        string `json:"grade"`
	Label        string `json:"label"`
	GradeCeiling string `json:"grade_ceiling"`
}

// NatalAssessment is the versioned, deterministic context used by the vehicle
// metaphor and Dayun road calculation. Legacy Yongshen/Jishen stay untouched.
type NatalAssessment struct {
	Version string                 `json:"version"`
	Tiaohou NatalTiaohouAssessment `json:"tiaohou"`
	// Climate is retained for existing consumers and mirrors Tiaohou.Thermal.
	Climate           NatalClimateAssessment  `json:"climate"`
	Fuyi              NatalFuyiAssessment     `json:"fuyi"`
	YongshenAlignment NatalYongshenAlignment  `json:"yongshen_alignment"`
	StemGuidance      *NatalStemGuidance      `json:"stem_guidance,omitempty"`
	Pattern           NatalPatternAssessment  `json:"pattern"`
	Relations         NatalRelationAssessment `json:"relations"`
	Carrying          NatalCarryingAssessment `json:"carrying"`
	Evidences         []NatalEvidence         `json:"evidences"`
	Grade             NatalGradeAssessment    `json:"grade"`
}

// EnsureNatalAssessment attaches a current assessment only when the chart does
// not already contain one from the current ruleset.
func EnsureNatalAssessment(r *BaziResult) bool {
	if r == nil || r.DayGan == "" {
		return false
	}
	if r.NatalAssessment != nil && r.NatalAssessment.Version == NatalAssessmentVersion {
		return false
	}
	r.NatalAssessment = AssessNatalStructure(r)
	return true
}

// AssessNatalStructure applies the fixed order: urgent climate gate, Fuyi base,
// pattern formation/break, then natal flow and auxiliary risk.
func AssessNatalStructure(r *BaziResult) *NatalAssessment {
	if r == nil || r.DayGan == "" {
		return nil
	}
	if r.MingGe == "" && r.YearZhi != "" && r.MonthZhi != "" && r.DayZhi != "" && r.HourZhi != "" {
		r.MingGe, r.MingGeDesc = DetectMingGe(r)
	}

	evidences := make([]NatalEvidence, 0, 8)
	tiaohou, tiaohouEvidence := assessNatalTiaohou(r)
	climate := climateFromThermal(tiaohou.Thermal)
	evidences = append(evidences, tiaohouEvidence...)

	globalFuyi := AssessFuyiStrength(r)
	fuyiYongshen, fuyiJishen := globalFuyi.Yongshen, globalFuyi.Jishen
	strength, strengthScore, strengthDetail := globalFuyi.Level, globalFuyi.Score, globalFuyi.Reason
	fuyiScore, fuyiLabel, fuyiDetail, fuyiCeiling := vehicleFuyiScore(r, fuyiYongshen, fuyiJishen)
	fuyi := NatalFuyiAssessment{
		DayMasterStrength: strength,
		Yongshen:          fuyiYongshen,
		Jishen:            fuyiJishen,
		SupportLevel:      fuyiLabel,
		Score:             fuyiScore,
		StrengthScore:     strengthScore,
		Evidence:          strengthDetail,
		GradeCeiling:      fuyiCeiling,
	}
	alignment := assessSharedPriorityYongshen(tiaohou.Thermal, fuyi)
	stemGuidance := assessNatalStemGuidance(r, tiaohou.DayStem, fuyi)
	fuyiDelta := int(math.Round(float64(fuyiScore) * 25 / 30))
	evidences = append(evidences, NatalEvidence{
		RuleID: "fuyi-support", Source: "扶抑用神", Label: fuyiLabel,
		Impact: impactLabel(fuyiDelta), Delta: fuyiDelta,
		Detail: fmt.Sprintf("%s 日主%s。", fuyiDetail, strengthLevelLabel(strength)) + strengthDetail,
	})

	pattern, patternEvidence := assessNatalPattern(r)
	applyDayStemFoundation(&pattern, tiaohou.DayStem)
	evidences = append(evidences, patternEvidence...)
	relations, relationEvidence := assessNatalRelations(r, fuyiYongshen, fuyiJishen)
	evidences = append(evidences, relationEvidence...)

	riskDelta, riskLabel, riskDetail := vehicleNatalModifierDelta(r)
	riskScore := clampInt(5+riskDelta, 0, 10)
	evidences = append(evidences, NatalEvidence{
		RuleID: "auxiliary-risk", Source: "辅助风险", Label: riskLabel,
		Impact: impactLabel(riskDelta), Delta: riskDelta, Detail: riskDetail,
	})

	score := clampInt(climate.Score+tiaohou.DayStem.FoundationScore+fuyiDelta+pattern.Score+relations.Score+riskScore, 0, 100)
	ceiling := stricterVehicleGrade(climate.GradeCeiling, fuyi.GradeCeiling)
	if pattern.CriticalBreak {
		ceiling = stricterVehicleGrade(ceiling, "C")
	}
	grade, label := vehicleGradeWithCeiling(score, ceiling)
	if grade != vehicleGradeOnly(score) {
		evidences = append(evidences, NatalEvidence{
			RuleID: "grade-ceiling", Source: "等级上限", Label: grade + "级上限", Impact: "限制", Delta: 0,
			Detail: natalCeilingDetail(climate.GradeCeiling, fuyi.GradeCeiling, pattern.CriticalBreak, vehicleGradeOnly(score), grade),
		})
	}

	carrying := NatalCarryingAssessment{Level: natalCarryingLevel(strength, fuyiScore), Detail: "日主" + strengthLevelLabel(strength) + "，" + fuyiLabel}
	return &NatalAssessment{
		Version: NatalAssessmentVersion,
		Tiaohou: tiaohou, Climate: climate, Fuyi: fuyi, YongshenAlignment: alignment, StemGuidance: stemGuidance, Pattern: pattern, Relations: relations, Carrying: carrying,
		Evidences: evidences,
		Grade:     NatalGradeAssessment{Score: score, Grade: grade, Label: label, GradeCeiling: ceiling},
	}
}

func assessNatalTiaohou(r *BaziResult) (NatalTiaohouAssessment, []NatalEvidence) {
	dayStem := assessDayStemTiaohou(r)
	thermal, thermalEvidence := assessThermalTiaohou(r)
	evidences := []NatalEvidence{thermalEvidence}
	if len(dayStem.RequiredStems) > 0 {
		ruleID, source, label, delta := "day-stem-tiaohou", "日干调候", dayStemStatusLabel(dayStem.Status), dayStem.Score
		if dayStem.Formation == "formed" {
			ruleID, source, label, delta = "day-stem-tiaohou-formed", "日干调候成格", "高格基础", dayStem.FoundationScore
		}
		impact := impactLabel(delta)
		if delta == 0 {
			impact = "中性"
		}
		evidences = append(evidences, NatalEvidence{
			RuleID: ruleID, Source: source, Label: label,
			Impact: impact, Delta: delta, Detail: dayStem.Detail,
		})
	}
	return NatalTiaohouAssessment{DayStem: dayStem, Thermal: thermal}, evidences
}

func assessDayStemTiaohou(r *BaziResult) NatalDayStemTiaohouAssessment {
	result := NatalDayStemTiaohouAssessment{RequiredStems: []string{}, VisibleStems: []string{}, HiddenStems: []string{}}
	if r == nil {
		return result
	}
	tiaohou := r.Tiaohou
	if tiaohou == nil {
		tiaohou = calcTiaohou(r)
	}
	if tiaohou == nil || len(tiaohou.Expected) == 0 {
		result.Status, result.Formation, result.FoundationTier = "unavailable", "unavailable", "none"
		result.Detail = "当前日干与月令未命中可用的日干调候字典条目。"
		return result
	}
	result.RequiredStems = compactStrings(append(result.RequiredStems, tiaohou.Expected...), 0)
	result.VisibleStems = matchingRequiredStems(result.RequiredStems, tiaohou.Tou)
	result.HiddenStems = matchingRequiredStems(result.RequiredStems, tiaohou.Cang)
	matched := map[string]bool{}
	for _, stem := range append(append([]string{}, result.VisibleStems...), result.HiddenStems...) {
		matched[stem] = true
	}
	allRequiredMatched := true
	for _, stem := range result.RequiredStems {
		if !matched[stem] {
			allRequiredMatched = false
			break
		}
	}
	switch {
	case allRequiredMatched && len(result.VisibleStems) > 0 && len(result.HiddenStems) > 0:
		result.Status, result.Formation, result.FoundationTier = "resolved", "formed", "high"
		result.Score, result.FoundationScore = 12, 24
	case len(matched) > 0:
		result.Status, result.Formation, result.FoundationTier = "partial", "partial", "normal"
		result.Score, result.FoundationScore = 6, 6
	default:
		result.Status, result.Formation, result.FoundationTier = "missing", "unformed", "none"
		result.Score, result.FoundationScore = 0, 0
	}
	parts := make([]string, 0, 2)
	if len(result.VisibleStems) > 0 {
		parts = append(parts, strings.Join(result.VisibleStems, "、")+"透干")
	}
	if len(result.HiddenStems) > 0 {
		parts = append(parts, strings.Join(result.HiddenStems, "、")+"藏支")
	}
	if len(parts) == 0 {
		parts = append(parts, "所需天干尚未显现")
	}
	if result.Formation == "formed" {
		result.Detail = fmt.Sprintf("日干%s生于%s月，日干调候取%s；%s，天透地藏成格，列为高格基础，计%d分。", r.DayGan, r.MonthZhi, strings.Join(result.RequiredStems, "、"), strings.Join(parts, "；"), result.FoundationScore)
	} else {
		result.Detail = fmt.Sprintf("日干%s生于%s月，日干调候取%s；%s，计%d分。", r.DayGan, r.MonthZhi, strings.Join(result.RequiredStems, "、"), strings.Join(parts, "；"), result.FoundationScore)
	}
	return result
}

func matchingRequiredStems(required, candidates []string) []string {
	requiredSet := map[string]bool{}
	for _, stem := range required {
		requiredSet[stem] = true
	}
	matched := make([]string, 0, len(candidates))
	for _, stem := range candidates {
		if requiredSet[stem] {
			matched = append(matched, stem)
		}
	}
	return compactStrings(matched, 0)
}

func assessThermalTiaohou(r *BaziResult) (NatalThermalTiaohouAssessment, NatalEvidence) {
	condition, required, urgent := thermalRequirement(r)
	if required == "" {
		return NatalThermalTiaohouAssessment{
			Status: "non_urgent", Condition: "平和", VisibleSupport: []string{}, HiddenSupport: []string{}, Score: 20, GradeCeiling: "S",
			Detail: "寒热调候未见急需，原局基础以扶抑用神为主线。",
		}, NatalEvidence{RuleID: "thermal-tiaohou", Source: "寒热调候", Label: "寒热无急", Impact: "中性", Delta: 0, Detail: "寒热调候未见急需，原局基础以扶抑用神为主线。"}
	}
	visible, hidden := natalElementSupportStems(r, required)
	result := NatalThermalTiaohouAssessment{
		Condition: condition, RequiredElements: required, VisibleSupport: visible, HiddenSupport: hidden,
	}
	label, impact, delta := "", "中性", 0
	switch {
	case len(visible) > 0:
		if urgent {
			result.Status, result.Score, result.GradeCeiling = "urgent_resolved", 30, "S"
			label, impact, delta = "寒热急需透干", "加分", 10
		} else {
			result.Status, result.Score, result.GradeCeiling = "seasonal_resolved", 22, "S"
			label, impact, delta = "季候已有透干", "加分", 2
		}
	case len(hidden) > 0:
		if urgent {
			result.Status, result.Score, result.GradeCeiling = "urgent_partial", 12, "A"
			label, impact, delta = "寒热急需藏支", "减分", -8
		} else {
			result.Status, result.Score, result.GradeCeiling = "seasonal_partial", 18, "S"
			label, impact, delta = "季候藏支可用", "减分", -2
		}
	default:
		if urgent {
			result.Status, result.Score, result.GradeCeiling = "urgent_unresolved", 0, "C"
			label, impact, delta = "寒热急需未解", "减分", -20
		} else {
			result.Status, result.Score, result.GradeCeiling = "seasonal_unresolved", 14, "S"
			label, impact, delta = "季候调剂不足", "减分", -6
		}
	}
	support := "透干和藏支均未见"
	if len(visible) > 0 {
		support = strings.Join(visible, "、") + "透干"
	} else if len(hidden) > 0 {
		support = strings.Join(hidden, "、") + "藏支，只作部分支持，仍待行运引动"
	}
	result.Detail = fmt.Sprintf("原局%s，取%s调剂；%s。", condition, required, support)
	return result, NatalEvidence{RuleID: "thermal-tiaohou", Source: "寒热调候", Label: label, Impact: impact, Delta: delta, Detail: result.Detail}
}

func thermalRequirement(r *BaziResult) (condition, required string, urgent bool) {
	if r == nil {
		return "", "", false
	}
	switch r.MonthZhi {
	case "寅":
		return "初春偏寒", "火", false
	case "亥", "子", "丑":
		return "至寒", "火", true
	case "巳", "午", "未":
		return "至热", "水", true
	case "戌":
		return "偏燥", "水", false
	case "辰":
		return "偏湿", "火", false
	default:
		return "", "", false
	}
}

func assessSharedPriorityYongshen(thermal NatalThermalTiaohouAssessment, fuyi NatalFuyiAssessment) NatalYongshenAlignment {
	alignment := NatalYongshenAlignment{Elements: []string{}, SourceLayers: []string{}}
	if thermal.RequiredElements == "" || fuyi.Yongshen == "" {
		return alignment
	}
	for _, element := range []string{"木", "火", "土", "金", "水"} {
		if strings.Contains(thermal.RequiredElements, element) && strings.Contains(fuyi.Yongshen, element) {
			alignment.Elements = append(alignment.Elements, element)
		}
	}
	if len(alignment.Elements) == 0 {
		return alignment
	}
	alignment.SourceLayers = []string{"寒热调候", "扶抑"}
	alignment.Detail = fmt.Sprintf("共同优先用神：%s（寒热调候 + 扶抑）。完整扶抑喜用仍为%s。", strings.Join(alignment.Elements, "、"), fuyi.Yongshen)
	return alignment
}

var stemGuidanceByElement = map[string][]string{
	"木": {"甲", "乙"},
	"火": {"丙", "丁"},
	"土": {"戊", "己"},
	"金": {"庚", "辛"},
	"水": {"壬", "癸"},
}

func assessNatalStemGuidance(r *BaziResult, dayStem NatalDayStemTiaohouAssessment, fuyi NatalFuyiAssessment) *NatalStemGuidance {
	guidance := &NatalStemGuidance{
		PrimaryFavorable:   []NatalStemGuidanceItem{},
		SecondaryFavorable: []NatalStemGuidanceItem{},
		ConditioningOnly:   []NatalStemGuidanceItem{},
		Adverse:            []NatalStemGuidanceItem{},
	}
	if r == nil {
		return guidance
	}

	// A stem-level Fuyi direction requires both sides of the conclusion. Do not
	// invent favorable or adverse stems for a neutral or incomplete assessment.
	hasFuyiDirection := fuyi.Yongshen != "" && fuyi.Jishen != ""
	primaryStems := map[string]bool{}
	conditioningStems := map[string]bool{}
	for _, stem := range dayStem.RequiredStems {
		element := stemGuidanceElement(stem)
		if hasFuyiDirection && wuxingSetContains(fuyi.Yongshen, ganWuxing[stem]) {
			guidance.PrimaryFavorable = append(guidance.PrimaryFavorable, newNatalStemGuidanceItem(
				r, stem, []string{"日干调候", "扶抑"},
				fmt.Sprintf("日干调候取%s，且%s属于扶抑喜用，列为天干优先。", stem, element),
			))
			primaryStems[stem] = true
			continue
		}

		detail := fmt.Sprintf("日干调候取%s，用于说明原局调候结构。", stem)
		if hasFuyiDirection && wuxingSetContains(fuyi.Jishen, ganWuxing[stem]) {
			detail = fmt.Sprintf("日干调候取%s，但%s属于扶抑忌神；用于说明原局调候结构，不作为后天通用喜神。", stem, element)
		} else if !hasFuyiDirection {
			detail = fmt.Sprintf("日干调候取%s；当前扶抑喜忌未定，不推导后天通用喜忌。", stem)
		}
		guidance.ConditioningOnly = append(guidance.ConditioningOnly, newNatalStemGuidanceItem(
			r, stem, []string{"日干调候"}, detail,
		))
		conditioningStems[stem] = true
	}

	if !hasFuyiDirection {
		return guidance
	}

	for _, stem := range stemGuidanceStemsForElements(fuyi.Yongshen) {
		if primaryStems[stem] {
			continue
		}
		element := stemGuidanceElement(stem)
		guidance.SecondaryFavorable = append(guidance.SecondaryFavorable, newNatalStemGuidanceItem(
			r, stem, []string{"扶抑"}, fmt.Sprintf("%s属于扶抑喜用，作为天干级可用方向。", element),
		))
	}
	for _, stem := range stemGuidanceStemsForElements(fuyi.Jishen) {
		if conditioningStems[stem] {
			continue
		}
		element := stemGuidanceElement(stem)
		guidance.Adverse = append(guidance.Adverse, newNatalStemGuidanceItem(
			r, stem, []string{"扶抑"}, fmt.Sprintf("%s属于扶抑忌神，后天遇此天干宜结合全局谨慎判断。", element),
		))
	}
	return guidance
}

func stemGuidanceStemsForElements(elements string) []string {
	stems := make([]string, 0, len(elements)*2)
	seen := map[string]bool{}
	for _, element := range elements {
		for _, stem := range stemGuidanceByElement[string(element)] {
			if !seen[stem] {
				stems = append(stems, stem)
				seen[stem] = true
			}
		}
	}
	return stems
}

func stemGuidanceElement(stem string) string {
	return wxPinyin2CN[ganWuxing[stem]]
}

func newNatalStemGuidanceItem(r *BaziResult, stem string, sourceLayers []string, detail string) NatalStemGuidanceItem {
	tenGod := ""
	if r != nil && r.DayGan != "" {
		tenGod = GetShiShen(r.DayGan, stem)
	}
	return NatalStemGuidanceItem{
		Stem:         stem,
		Element:      stemGuidanceElement(stem),
		TenGod:       tenGod,
		SourceLayers: append([]string{}, sourceLayers...),
		Detail:       detail,
	}
}

func climateFromThermal(thermal NatalThermalTiaohouAssessment) NatalClimateAssessment {
	return NatalClimateAssessment{Status: thermal.Status, RequiredElements: thermal.RequiredElements, Score: thermal.Score, GradeCeiling: thermal.GradeCeiling}
}

func natalElementSupportStems(r *BaziResult, required string) (visible, hidden []string) {
	if r == nil || required == "" {
		return []string{}, []string{}
	}
	for _, gan := range []string{r.YearGan, r.MonthGan, r.HourGan} {
		if wuxingSetContains(required, ganWuxing[gan]) {
			visible = append(visible, gan)
		}
	}
	for _, group := range [][]string{r.YearHideGan, r.MonthHideGan, r.DayHideGan, r.HourHideGan} {
		for _, gan := range group {
			if wuxingSetContains(required, ganWuxing[gan]) {
				hidden = append(hidden, gan)
			}
		}
	}
	return compactStrings(visible, 0), compactStrings(hidden, 0)
}

func dayStemStatusLabel(status string) string {
	switch status {
	case "resolved":
		return "日干调候天透地藏"
	case "partial":
		return "日干调候已有依据"
	case "missing":
		return "日干调候待补"
	default:
		return "日干调候待查"
	}
}

func applyDayStemFoundation(pattern *NatalPatternAssessment, dayStem NatalDayStemTiaohouAssessment) {
	if pattern == nil || dayStem.Formation != "formed" {
		return
	}
	pattern.FoundationSource = "日干调候成格"
	pattern.FoundationLabel = "高格基础"
	pattern.FoundationTier = "high"
}

func assessNatalPattern(r *BaziResult) (NatalPatternAssessment, []NatalEvidence) {
	name := nonEmpty(r.MingGe, "未定格")
	stats := natalTenGodStats(r)
	pattern := NatalPatternAssessment{Name: name, Formations: []string{}, Breaks: []string{}, SupportTenGods: []string{}, BreakTenGods: []string{}}
	evidences := make([]NatalEvidence, 0, 4)
	star := patternPrimaryTenGod(name)
	if star == "" {
		pattern.Quality = "undecided"
		evidences = append(evidences, NatalEvidence{RuleID: "pattern-base", Source: "格局成败", Label: name, Impact: "中性", Detail: "未能确认可评分的主格，结构按保守待定处理。"})
		return pattern, evidences
	}

	present := stats[star]
	if present.visible > 0 {
		pattern.Score += 9
	}
	if present.hidden > 0 {
		pattern.Score += 5
	}
	if rootStrengthForTenGod(r, star) >= 4 {
		pattern.Score += 5
	}
	if pattern.Score == 0 {
		evidences = append(evidences, NatalEvidence{RuleID: "pattern-base", Source: "格局成败", Label: name + "主星未显", Impact: "减分", Delta: -8, Detail: "系统取到格名，但主格十神未见透出、通根支撑，不能按成格计分。"})
	} else {
		evidences = append(evidences, NatalEvidence{RuleID: "pattern-base", Source: "格局成败", Label: name + "主星有据", Impact: "加分", Delta: pattern.Score, Detail: fmt.Sprintf("主格十神%s透干%d处、藏支%d处，按实际显露和通根评分。", star, present.visible, present.hidden)})
	}

	addFormation := func(rule, detail string, support []string) {
		pattern.Score += 9
		pattern.Formations = append(pattern.Formations, rule)
		pattern.SupportTenGods = compactStrings(append(pattern.SupportTenGods, support...), 0)
		evidences = append(evidences, NatalEvidence{RuleID: "pattern-formation", Source: "格局制化", Label: rule, Impact: "加分", Delta: 9, Detail: detail})
	}
	addBreak := func(rule, detail string, breakGods []string, critical bool) {
		pattern.Score -= 12
		pattern.Breaks = append(pattern.Breaks, rule)
		pattern.BreakTenGods = compactStrings(append(pattern.BreakTenGods, breakGods...), 0)
		if critical {
			pattern.CriticalBreak = true
		}
		evidences = append(evidences, NatalEvidence{RuleID: "pattern-break", Source: "格局成败", Label: rule, Impact: "减分", Delta: -12, Detail: detail})
	}
	hasVisible := func(names ...string) bool { return stats.hasVisible(names...) }
	hasVisibleAndRooted := func(names ...string) bool { return stats.hasVisibleAndRooted(r, names...) }
	primaryVisibleAndRooted := hasVisibleAndRooted(star)

	switch name {
	case "伤官格":
		if primaryVisibleAndRooted && hasVisible("正印", "偏印") {
			addFormation("伤官配印", "伤官见印，才华有印星收束和转化。", []string{"正印", "偏印"})
		}
		if primaryVisibleAndRooted && hasVisible("正财", "偏财") {
			addFormation("伤官生财", "伤官见财，输出有财星承接。", []string{"正财", "偏财"})
		}
		if primaryVisibleAndRooted && hasVisibleAndRooted("正官") && !hasVisible("正印", "偏印") {
			addBreak("伤官见官", "伤官与正官同见而无印星通关，主线容易冲突。", []string{"正官"}, true)
		}
	case "食神格":
		if primaryVisibleAndRooted && hasVisible("正财", "偏财") {
			addFormation("食神生财", "食神见财，才艺和产出有承接。", []string{"正财", "偏财"})
		}
		if primaryVisibleAndRooted && hasVisibleAndRooted("偏印") {
			addBreak("枭神夺食", "食神透干且有根，又见透干有根的偏印，输出与资源的衔接需留意。", []string{"偏印"}, !hasVisible("正财", "偏财"))
		}
	case "七杀格":
		if primaryVisibleAndRooted && hasVisible("食神") {
			addFormation("食神制杀", "七杀有食神制约，压力可被转化为执行力。", []string{"食神"})
		}
		if primaryVisibleAndRooted && hasVisible("正印", "偏印") {
			addFormation("杀印相生", "七杀见印，压力有印星承接和化解。", []string{"正印", "偏印"})
		}
		if primaryVisibleAndRooted && !hasVisible("食神", "正印", "偏印") && pattern.Score < 15 {
			addBreak("杀重无制", "七杀未见食神或印星制化，且主星基础不足。", []string{"七杀"}, true)
		}
	case "正官格":
		if primaryVisibleAndRooted && hasVisible("正印", "偏印") {
			addFormation("官印相生", "正官见印，规则与承载形成连结。", []string{"正印", "偏印"})
		}
		if primaryVisibleAndRooted && hasVisible("正财", "偏财") {
			addFormation("财官相生", "财星见官，资源能承接规则与责任。", []string{"正财", "偏财"})
		}
		if primaryVisibleAndRooted && hasVisibleAndRooted("伤官") && !hasVisible("正印", "偏印") {
			addBreak("伤官见官", "正官与伤官同见而无印星通关，规则主线容易受扰。", []string{"伤官"}, true)
		}
	case "正印格":
		if primaryVisibleAndRooted && hasVisible("正官", "七杀") {
			addFormation("官印相生", "官杀见印，外部压力可转为资质和支持。", []string{"正官", "七杀"})
		}
		if primaryVisibleAndRooted && hasVisibleAndRooted("正财", "偏财") && !hasVisible("比肩", "劫财") {
			addBreak("财坏印", "财星明显而少见比劫护印，印星承载需谨慎。", []string{"正财", "偏财"}, false)
		}
	case "偏印格":
		if primaryVisibleAndRooted && hasVisibleAndRooted("食神") {
			addBreak("枭神夺食", "偏印格见食神透干且有根，输出与资源的衔接需留意。", []string{"食神"}, !hasVisible("正财", "偏财"))
		} else if stats.hasHidden("食神") {
			evidences = append(evidences, NatalEvidence{RuleID: "pattern-context", Source: "格局观察", Label: "食神藏支", Impact: "中性", Delta: 0, Detail: "食神仅见于藏支，保留为结构背景，不按枭神夺食扣分。"})
		}
	case "正财格", "偏财格":
		if primaryVisibleAndRooted && hasVisible("正官", "七杀") {
			addFormation("财官相生", "财星见官杀，资源有方向和约束。", []string{"正官", "七杀"})
		}
		if primaryVisibleAndRooted && hasVisible("食神", "伤官") {
			addFormation("食伤生财", "食伤见财，产出有资源承接。", []string{"食神", "伤官"})
		}
		if primaryVisibleAndRooted && (rDayMasterWeak(r) || rDayMasterVeryWeak(r)) && present.total() >= 2 {
			addBreak("财多身弱", "财星较多而日主偏弱，资源承载需量力。", []string{"正财", "偏财"}, true)
		}
		if primaryVisibleAndRooted && hasVisibleAndRooted("比肩", "劫财") && !hasVisible("正官", "七杀") {
			addBreak("比劫夺财", "财星见比劫而少官杀约束，资源分散风险增加。", []string{"比肩", "劫财"}, false)
		}
	case "建禄格", "月禄格", "月刃格":
		if primaryVisibleAndRooted && hasVisible("正官", "七杀") {
			addFormation("官杀制身", "比劫旺而见官杀，动力有约束和出口。", []string{"正官", "七杀"})
		}
		if primaryVisibleAndRooted && hasVisible("食神", "伤官") {
			addFormation("食伤泄秀", "比劫旺而见食伤，力量有表达和消耗出口。", []string{"食神", "伤官"})
		}
		if primaryVisibleAndRooted && !hasVisible("正官", "七杀", "食神", "伤官", "正财", "偏财") {
			addBreak("比劫偏重", "比劫为主而少见疏泄、财官，力量容易壅滞。", []string{"比肩", "劫财"}, name == "月刃格")
		}
	}

	pattern.Score = clampInt(pattern.Score, 0, 30)
	switch {
	case pattern.Score >= 25:
		pattern.Quality = "formed"
	case pattern.Score >= 16:
		pattern.Quality = "usable"
	case pattern.Score >= 8:
		pattern.Quality = "partial"
	default:
		pattern.Quality = "broken"
	}
	return pattern, evidences
}

func assessNatalRelations(r *BaziResult, yongshen, jishen string) (NatalRelationAssessment, []NatalEvidence) {
	result := NatalRelationAssessment{Combinations: []string{}, Disruptions: []string{}}
	evidences := make([]NatalEvidence, 0, 2)
	elements := natalElementSet(r)
	bestChain := 0
	for start := range elements {
		length := 1
		for next := wxSheng[start]; next != "" && elements[next] && length < 5; next = wxSheng[next] {
			length++
		}
		if length > bestChain {
			bestChain = length
		}
	}
	switch {
	case bestChain >= 3:
		result.Flow, result.Score = "smooth", 10
	case bestChain == 2:
		result.Flow, result.Score = "partial", 6
	default:
		result.Flow, result.Score = "blocked", 2
	}
	evidences = append(evidences, NatalEvidence{RuleID: "element-flow", Source: "原局流通", Label: map[string]string{"smooth": "相生链较顺", "partial": "相生链部分连结", "blocked": "相生链不足"}[result.Flow], Impact: impactLabel(result.Score - 6), Delta: result.Score - 6, Detail: fmt.Sprintf("原局可识别的最长五行相生链为%d段；流通仅作结构修正，不替代调候与扶抑。", bestChain)})

	zhis := []string{r.YearZhi, r.MonthZhi, r.DayZhi, r.HourZhi}
	if combined := detectSanHeHui(zhis); combined != "" {
		result.Combinations = append(result.Combinations, "三合三会"+combined)
		if strings.Contains(yongshen, combined) {
			result.Score += 2
		} else if strings.Contains(jishen, combined) {
			result.Score -= 2
		}
	}
	disruption := 0
	for i := 0; i < len(zhis); i++ {
		for j := i + 1; j < len(zhis); j++ {
			a, b := zhis[i], zhis[j]
			if sixHe[a] == b {
				result.Combinations = append(result.Combinations, a+b+"六合")
			}
			if sixChong[a] == b {
				result.Disruptions = append(result.Disruptions, a+b+"冲")
				disruption += 2
			}
			if sixHai[a] == b {
				result.Disruptions = append(result.Disruptions, a+b+"害")
				disruption++
			}
			if sixXing[a] == b || sixXing[b] == a {
				result.Disruptions = append(result.Disruptions, a+b+"刑")
				disruption++
			}
		}
	}
	if disruption > 0 {
		penalty := clampInt(disruption, 0, 4)
		result.Score -= penalty
		evidences = append(evidences, NatalEvidence{RuleID: "branch-relations", Source: "支干关系", Label: "冲刑害修正", Impact: "减分", Delta: -penalty, Detail: "原局见" + strings.Join(compactStrings(result.Disruptions, 0), "、") + "，按关系层保守修正，不单独决定吉凶。"})
	}
	result.Score = clampInt(result.Score, 0, 15)
	return result, evidences
}

type tenGodCount struct{ visible, hidden int }

func (c tenGodCount) total() int { return c.visible + c.hidden }

type tenGodStats map[string]tenGodCount

func (s tenGodStats) hasAny(names ...string) bool {
	for _, name := range names {
		if s[name].total() > 0 {
			return true
		}
	}
	return false
}

func (s tenGodStats) hasVisible(names ...string) bool {
	for _, name := range names {
		if s[name].visible > 0 {
			return true
		}
	}
	return false
}

func (s tenGodStats) hasHidden(names ...string) bool {
	for _, name := range names {
		if s[name].hidden > 0 {
			return true
		}
	}
	return false
}

func (s tenGodStats) hasVisibleAndRooted(r *BaziResult, names ...string) bool {
	for _, name := range names {
		if s[name].visible > 0 && rootStrengthForTenGod(r, name) > 0 {
			return true
		}
	}
	return false
}

func natalTenGodStats(r *BaziResult) tenGodStats {
	stats := tenGodStats{}
	for _, gan := range []string{r.YearGan, r.MonthGan, r.HourGan} {
		if ss := GetShiShen(r.DayGan, gan); ss != "" {
			count := stats[ss]
			count.visible++
			stats[ss] = count
		}
	}
	for _, group := range [][]string{r.YearHideGan, r.MonthHideGan, r.DayHideGan, r.HourHideGan} {
		for _, gan := range group {
			if ss := GetShiShen(r.DayGan, gan); ss != "" {
				count := stats[ss]
				count.hidden++
				stats[ss] = count
			}
		}
	}
	return stats
}

func patternPrimaryTenGod(name string) string {
	switch name {
	case "建禄格", "月禄格":
		return "比肩"
	case "月刃格":
		return "劫财"
	case "杂气格", "":
		return ""
	default:
		return strings.TrimSuffix(name, "格")
	}
}

func rootStrengthForTenGod(r *BaziResult, tenGod string) int {
	best := 0
	for _, gan := range Gan {
		if GetShiShen(r.DayGan, gan) == tenGod {
			if strength := rootStrength(gan, r); strength > best {
				best = strength
			}
		}
	}
	return best
}

func natalElementSet(r *BaziResult) map[string]bool {
	set := map[string]bool{}
	for _, gan := range []string{r.YearGan, r.MonthGan, r.DayGan, r.HourGan} {
		if wx := ganWuxing[gan]; wx != "" {
			set[wx] = true
		}
	}
	for _, group := range [][]string{r.YearHideGan, r.MonthHideGan, r.DayHideGan, r.HourHideGan} {
		for _, gan := range group {
			if wx := ganWuxing[gan]; wx != "" {
				set[wx] = true
			}
		}
	}
	return set
}

func rDayMasterWeak(r *BaziResult) bool {
	level, _, _ := dayMasterStrengthLevel(r)
	return level == "weak"
}
func rDayMasterVeryWeak(r *BaziResult) bool {
	level, _, _ := dayMasterStrengthLevel(r)
	return level == "vweak"
}

func natalCarryingLevel(strength string, fuyiScore int) string {
	if fuyiScore < 8 || strength == "vweak" {
		return "不足"
	}
	if fuyiScore < 16 || strength == "vstrong" {
		return "有限"
	}
	return "可承"
}

func natalCeilingDetail(climateCeiling, fuyiCeiling string, criticalBreak bool, rawGrade, finalGrade string) string {
	reason := "原局结构硬性上限"
	if climateCeiling != "S" {
		reason = "调候急需上限"
	} else if fuyiCeiling != "S" {
		reason = "扶抑用神上限"
	} else if criticalBreak {
		reason = "格局关键破损上限"
	}
	return fmt.Sprintf("%s生效；原始分数对应%s级，按保守门槛限制为%s级。", reason, rawGrade, finalGrade)
}
