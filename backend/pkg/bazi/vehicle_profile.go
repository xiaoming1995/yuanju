package bazi

import (
	"fmt"
	"math"
	"strings"
)

const (
	RoadTypeHighway      = "highway"
	RoadTypeMainRoad     = "main_road"
	RoadTypeMountainRoad = "mountain_road"
	RoadTypeMuddyRoad    = "muddy_road"
	RoadTypeConstruction = "construction"
)

type ProfileEvidence struct {
	Source string `json:"source"`
	Label  string `json:"label"`
	Impact string `json:"impact"`
	Delta  int    `json:"delta"`
	Detail string `json:"detail"`
}

type VehicleProfile struct {
	Grade        string            `json:"grade"`
	GradeLabel   string            `json:"grade_label"`
	Score        int               `json:"score"`
	VehicleType  string            `json:"vehicle_type"`
	DrivingStyle string            `json:"driving_style,omitempty"`
	Summary      string            `json:"summary"`
	Tags         []string          `json:"tags"`
	Evidences    []ProfileEvidence `json:"evidences"`
}

type RoadPhase struct {
	Key     string `json:"key"`
	Label   string `json:"label"`
	Score   int    `json:"score"`
	Summary string `json:"summary"`
	Detail  string `json:"detail,omitempty"`
}

type DayunPhaseEvidence struct {
	Phase     string            `json:"phase"`
	Label     string            `json:"label"`
	Delta     int               `json:"delta"`
	Evidences []ProfileEvidence `json:"evidences"`
}

type DayunRoad struct {
	DayunIndex     int                  `json:"dayun_index"`
	GanZhi         string               `json:"gan_zhi"`
	Score          int                  `json:"score"`
	RoadType       string               `json:"road_type"`
	RoadLabel      string               `json:"road_label"`
	QianRoad       RoadPhase            `json:"qian_road"`
	HouRoad        RoadPhase            `json:"hou_road"`
	Summary        string               `json:"summary"`
	Tags           []string             `json:"tags"`
	Evidences      []ProfileEvidence    `json:"evidences"`
	PhaseEvidences []DayunPhaseEvidence `json:"phase_evidences,omitempty"`
}

func BuildVehicleRoadProfile(r *BaziResult) (*VehicleProfile, []DayunRoad) {
	if r == nil {
		return nil, nil
	}
	EnsureNatalAssessment(r)
	profile := buildVehicleProfile(r)
	roadmap := buildDayunRoadmap(r)
	return profile, roadmap
}

func buildVehicleProfile(r *BaziResult) *VehicleProfile {
	assessment := r.NatalAssessment
	if assessment == nil {
		assessment = AssessNatalStructure(r)
	}
	if assessment == nil {
		return nil
	}
	evidences := make([]ProfileEvidence, 0, len(assessment.Evidences))
	for _, item := range assessment.Evidences {
		evidences = append(evidences, ProfileEvidence{Source: item.Source, Label: item.Label, Impact: item.Impact, Delta: item.Delta, Detail: item.Detail})
	}
	grade, gradeLabel := assessment.Grade.Grade, assessment.Grade.Label
	vehicleType := resolveVehicleType(grade)
	drivingStyle := resolveDrivingStyle(r.MingGe)
	tags := []string{gradeLabel, vehicleType}
	if assessment.Pattern.FoundationTier == "high" {
		tags = append(tags, "日干调候成格·高格基础")
	}
	tags = append(tags, "主格结构"+patternQualityLabel(assessment.Pattern.Quality), "流通"+relationFlowLabel(assessment.Relations.Flow))
	if assessment.Fuyi.Yongshen != "" {
		tags = append(tags, "扶抑喜用"+assessment.Fuyi.Yongshen)
	}
	if len(assessment.YongshenAlignment.Elements) > 0 {
		tags = append(tags, "共同优先"+strings.Join(assessment.YongshenAlignment.Elements, "、"))
	}

	return &VehicleProfile{
		Grade:        grade,
		GradeLabel:   gradeLabel,
		Score:        assessment.Grade.Score,
		VehicleType:  vehicleType,
		DrivingStyle: drivingStyle,
		Summary:      vehicleSummaryFromAssessment(assessment, vehicleType),
		Tags:         compactStrings(tags, 5),
		Evidences:    evidences,
	}
}

func buildDayunRoadmap(r *BaziResult) []DayunRoad {
	if len(r.Dayun) == 0 {
		return nil
	}
	EnsureNatalAssessment(r)
	out := make([]DayunRoad, 0, len(r.Dayun))
	for _, dy := range r.Dayun {
		score := 50
		evidences := make([]ProfileEvidence, 0, 5)
		frontEvidences := make([]ProfileEvidence, 0, 3)
		backEvidences := make([]ProfileEvidence, 0, 5)

		qian := phaseFromJinBuHuan("front", dy.JinBuHuan)
		hou := phaseFromJinBuHuan("back", dy.JinBuHuan)
		jbhDelta := int(math.Round((float64(qian.Score+hou.Score)/2.0 - 50.0) * 0.4))
		score += jbhDelta
		evidences = append(evidences, ProfileEvidence{
			Source: "金不换",
			Label:  qian.Label + "/" + hou.Label,
			Impact: impactLabel(jbhDelta),
			Delta:  jbhDelta,
			Detail: strings.TrimSpace(qian.Summary + " " + hou.Summary),
		})

		frontElement, backElement := dayunElementPhaseEvidences(r, r.NatalAssessment, dy)
		frontEvidences = append(frontEvidences, frontElement)
		backEvidences = append(backEvidences, backElement)
		wxDelta := frontElement.Delta + backElement.Delta
		score += wxDelta
		evidences = append(evidences, aggregatePhaseEvidence("大运五行", dy.Gan+dy.Zhi, frontElement, backElement))

		frontPattern, backPattern := dayunPatternPhaseEvidences(r, r.NatalAssessment, dy)
		frontEvidences = append(frontEvidences, frontPattern)
		backEvidences = append(backEvidences, backPattern)
		patternDelta := frontPattern.Delta + backPattern.Delta
		score += patternDelta
		evidences = append(evidences, aggregatePhaseEvidence("格局作用", aggregatePatternLabel(frontPattern, backPattern), frontPattern, backPattern))

		frontTenGod, backTenGod := dayunTenGodPhaseEvidences(r, dy)
		frontEvidences = append(frontEvidences, frontTenGod)
		backEvidences = append(backEvidences, backTenGod)
		tenGodDelta := frontTenGod.Delta + backTenGod.Delta
		score += tenGodDelta
		evidences = append(evidences, aggregatePhaseEvidence("大运十神", dy.GanShiShen+"/"+dy.ZhiShiShen, frontTenGod, backTenGod))

		diShiDelta := diShiRoadDelta(dy.DiShi)
		score += diShiDelta
		diShiEvidence := ProfileEvidence{
			Source: "十二长生",
			Label:  nonEmpty(dy.DiShi, "无"),
			Impact: impactLabel(diShiDelta),
			Delta:  diShiDelta,
			Detail: "大运地支对应日主十二长生为" + nonEmpty(dy.DiShi, "无"),
		}
		backEvidences = append(backEvidences, diShiEvidence)
		evidences = append(evidences, diShiEvidence)

		modDelta, modDetail := dayunModifierDelta(dy)
		score += modDelta
		modifierEvidence := ProfileEvidence{
			Source: "神煞修正",
			Label:  modifierLabel(modDelta),
			Impact: impactLabel(modDelta),
			Delta:  modDelta,
			Detail: modDetail,
		}
		backEvidences = append(backEvidences, modifierEvidence)
		evidences = append(evidences, modifierEvidence)

		score = clampInt(score, 0, 100)
		roadType, roadLabel := roadTypeForScore(score)
		tags := compactStrings([]string{
			roadLabel,
			"前五年" + qian.Label,
			"后五年" + hou.Label,
			dy.GanShiShen,
		}, 5)

		out = append(out, DayunRoad{
			DayunIndex: dy.Index,
			GanZhi:     dy.Gan + dy.Zhi,
			Score:      score,
			RoadType:   roadType,
			RoadLabel:  roadLabel,
			QianRoad:   qian,
			HouRoad:    hou,
			Summary:    dayunRoadSummary(dy, roadLabel, qian, hou),
			Tags:       tags,
			Evidences:  evidences,
			PhaseEvidences: []DayunPhaseEvidence{
				newDayunPhaseEvidence("front", "前五年 · 天干主事", frontEvidences),
				newDayunPhaseEvidence("back", "后五年 · 地支主事", backEvidences),
			},
		})
	}
	return out
}

func vehicleDayMasterScore(level string) int {
	switch level {
	case "neutral":
		return 12
	case "strong", "weak":
		return 8
	case "vstrong", "vweak":
		return 2
	default:
		return 5
	}
}

func vehicleMinggeScore(mingge string) int {
	if strings.TrimSpace(mingge) == "" {
		return 0
	}
	if mingge == "杂气格" {
		return 6
	}
	return 15
}

func vehicleMinggeDetail(mingge string) string {
	if strings.TrimSpace(mingge) == "" {
		return "暂无系统定格结果，结构清晰度按保守分处理。"
	}
	if mingge == "杂气格" {
		return "系统定格为杂气格，代表结构复杂，并非简单吉凶。"
	}
	return "系统定格为" + mingge + "，原局主线较容易被识别。"
}

func vehicleNatalModifierDelta(r *BaziResult) (int, string, string) {
	all := append(append(append(append([]string{}, r.YearShenSha...), r.MonthShenSha...), r.DayShenSha...), r.HourShenSha...)
	good, watch := 0, 0
	for _, name := range all {
		if isHelpfulShensha(name) {
			good++
		}
		if isHarshShensha(name) {
			watch++
		}
	}
	delta := clampInt(good-watch*3, -10, 5)
	label := "修正平稳"
	if delta > 0 {
		label = "助力较多"
	} else if delta < 0 {
		label = "阻力较多"
	}
	return delta, label, fmt.Sprintf("吉性神煞%d项，压力神煞%d项，按辅助层小幅修正。", good, watch)
}

// vehicleUrgentTiaohouScore handles only the cold/hot conditions that the
// classical rule identifies as urgent. Ordinary charts are not promoted for
// merely matching a Tiaohou dictionary entry.
func vehicleUrgentTiaohouScore(r *BaziResult, classical ClassicalYongshenResult) (int, string, string, string) {
	if classical.Strategy != ClassicalStrategyTiaohouCold && classical.Strategy != ClassicalStrategyTiaohouHot {
		return 24, "调候不急", "原局不属至寒或至热，调候无急病，进入扶抑为主的判断。", "S"
	}

	tou, cang := natalElementSupport(r, classical.WuxingSet)
	condition := "至寒"
	if classical.Strategy == ClassicalStrategyTiaohouHot {
		condition = "至热"
	}
	switch {
	case tou > 0:
		return 35, "调候急需透干", fmt.Sprintf("原局%s，急需%s，现有%d处透干可用。", condition, classical.WuxingSet, tou), "S"
	case cang > 0:
		return 18, "调候急需藏支", fmt.Sprintf("原局%s，急需%s，仅%d处藏支可用，仍待引动。", condition, classical.WuxingSet, cang), "A"
	default:
		return 0, "调候急需未解", fmt.Sprintf("原局%s，急需%s，透干和藏支均未见可用力量。", condition, classical.WuxingSet), "C"
	}
}

func vehicleFuyiScore(r *BaziResult, yongshen, jishen string) (int, string, string, string) {
	if yongshen == "" {
		return 0, "扶抑未定", "无法根据日主和四柱推导扶抑用神，按无可用扶抑支持处理。", "B"
	}

	yongTou, yongCang := natalElementSupport(r, yongshen)
	jiTou, jiCang := natalElementSupport(r, jishen)
	score := 0
	switch {
	case yongTou >= 2:
		score = 26
	case yongTou == 1:
		score = 20
	case yongCang >= 2:
		score = 14
	case yongCang == 1:
		score = 8
	}
	if jiTou == 0 {
		score += 4
	} else {
		score -= clampInt(jiTou*3, 0, 8)
	}
	score = clampInt(score, 0, 30)
	detail := fmt.Sprintf("扶抑喜用%s：透干%d处、藏支%d处；扶抑忌神%s：透干%d处、藏支%d处。", yongshen, yongTou, yongCang, nonEmpty(jishen, "未定"), jiTou, jiCang)
	switch {
	case score >= 24:
		return score, "扶抑用神得力", detail + " 用神外显且忌神未形成明显压制。", "S"
	case score >= 16:
		return score, "扶抑用神可用", detail + " 用神已有可用支撑，仍需注意忌神干扰。", "S"
	case score >= 8:
		return score, "扶抑用神有根待引", detail + " 用神主要藏于支中，发挥更依赖行运引动。", "S"
	default:
		return score, "扶抑用神无力", detail + " 用神缺少有效支撑或受忌神压制，限制基础盘上限。", "B"
	}
}

func natalElementSupport(r *BaziResult, wuxingSet string) (tou, cang int) {
	if r == nil || wuxingSet == "" {
		return 0, 0
	}
	for _, gan := range []string{r.YearGan, r.MonthGan, r.HourGan} {
		if wuxingSetContains(wuxingSet, ganWuxing[gan]) {
			tou++
		}
	}
	for _, hideGans := range [][]string{r.YearHideGan, r.MonthHideGan, r.DayHideGan, r.HourHideGan} {
		for _, gan := range hideGans {
			if wuxingSetContains(wuxingSet, ganWuxing[gan]) {
				cang++
			}
		}
	}
	return tou, cang
}

func wuxingSetContains(wuxingSet, wuxingPinyin string) bool {
	wuxingCN, ok := wxPinyin2CN[wuxingPinyin]
	return ok && strings.Contains(wuxingSet, wuxingCN)
}

func vehicleGradeWithCeiling(score int, ceiling string) (string, string) {
	grade, _ := vehicleGrade(score)
	grade = stricterVehicleGrade(grade, ceiling)
	_, label := vehicleGrade(gradeMinimumScore(grade))
	return grade, label
}

func vehicleGradeOnly(score int) string {
	grade, _ := vehicleGrade(score)
	return grade
}

func stricterVehicleGrade(a, b string) string {
	if gradeRank(a) <= gradeRank(b) {
		return a
	}
	return b
}

func gradeRank(grade string) int {
	switch grade {
	case "S":
		return 5
	case "A":
		return 4
	case "B":
		return 3
	case "C":
		return 2
	default:
		return 1
	}
}

func gradeMinimumScore(grade string) int {
	switch grade {
	case "S":
		return 85
	case "A":
		return 70
	case "B":
		return 50
	case "C":
		return 30
	default:
		return 0
	}
}

func vehicleGradeCeilingReason(ceiling, tiaohouCeiling, fuyiCeiling string) string {
	if ceiling == tiaohouCeiling && ceiling != "S" {
		return "调候为急的硬性上限生效"
	}
	if ceiling == fuyiCeiling && ceiling != "S" {
		return "扶抑用神的硬性上限生效"
	}
	return "基础盘硬性上限生效"
}

func phaseFromJinBuHuan(part string, jbh *JinBuHuanResult) RoadPhase {
	level := "平"
	detail := "金不换未提供此阶段评级，按平稳路况处理。"
	if jbh != nil {
		if part == "front" {
			level = nonEmpty(jbh.QianLevel, "平")
			detail = nonEmpty(jbh.QianDesc, detail)
		} else {
			level = nonEmpty(jbh.HouLevel, "平")
			detail = nonEmpty(jbh.HouDesc, detail)
		}
	}
	score := 58
	switch level {
	case "吉":
		score = 82
	case "凶":
		score = 35
	}
	roadType, label := roadTypeForScore(score)
	phaseName := "前五年"
	if part == "back" {
		phaseName = "后五年"
	}
	return RoadPhase{
		Key:     roadType,
		Label:   label,
		Score:   score,
		Summary: phaseName + "：" + label,
		Detail:  detail,
	}
}

func newDayunPhaseEvidence(phase, label string, evidences []ProfileEvidence) DayunPhaseEvidence {
	delta := 0
	for _, evidence := range evidences {
		delta += evidence.Delta
	}
	return DayunPhaseEvidence{Phase: phase, Label: label, Delta: delta, Evidences: evidences}
}

func aggregatePhaseEvidence(source, label string, front, back ProfileEvidence) ProfileEvidence {
	delta := front.Delta + back.Delta
	return ProfileEvidence{
		Source: source,
		Label:  label,
		Impact: impactLabel(delta),
		Delta:  delta,
		Detail: fmt.Sprintf("前五年：%s（%+d）；后五年：%s（%+d）；十年合计%+d。", front.Detail, front.Delta, back.Detail, back.Delta, delta),
	}
}

func aggregatePatternLabel(front, back ProfileEvidence) string {
	labels := compactStrings([]string{front.Label, back.Label}, 0)
	if len(labels) == 0 {
		return "未直接作用主格"
	}
	return strings.Join(labels, " / ")
}

func dayunElementPhaseEvidences(r *BaziResult, assessment *NatalAssessment, dy DayunItem) (ProfileEvidence, ProfileEvidence) {
	yongshen, jishen := r.Yongshen, r.Jishen
	urgentElement := ""
	if assessment != nil {
		yongshen, jishen = assessment.Fuyi.Yongshen, assessment.Fuyi.Jishen
		thermal := assessment.Tiaohou.Thermal
		if thermal.Status == "" {
			thermal = NatalThermalTiaohouAssessment{
				Status: assessment.Climate.Status, RequiredElements: assessment.Climate.RequiredElements,
			}
		}
		if strings.HasPrefix(thermal.Status, "urgent_") {
			urgentElement = thermal.RequiredElements
		}
	}
	return dayunElementPhaseEvidence("天干", dy.Gan, wxPinyin2CN[ganWuxing[dy.Gan]], yongshen, jishen, urgentElement),
		dayunElementPhaseEvidence("地支", dy.Zhi, wxPinyin2CN[zhiWuxing[dy.Zhi]], yongshen, jishen, urgentElement)
}

func dayunElementPhaseEvidence(position, stemOrBranch, wuxing, yongshen, jishen, urgentElement string) ProfileEvidence {
	detail := position + stemOrBranch + "五行无法识别，按中性处理。"
	delta := 0
	if wuxing != "" {
		switch {
		case urgentElement != "" && strings.Contains(urgentElement, wuxing):
			delta = 8
			detail = position + stemOrBranch + "补调候急需" + wuxing
		case yongshen != "" && strings.Contains(yongshen, wuxing):
			delta = 7
			detail = position + stemOrBranch + "为扶抑喜用" + wuxing
		case jishen != "" && strings.Contains(jishen, wuxing):
			delta = -7
			detail = position + stemOrBranch + "为扶抑忌" + wuxing
		default:
			detail = position + stemOrBranch + "为中性" + wuxing
		}
	}
	return ProfileEvidence{Source: "大运五行", Label: stemOrBranch, Impact: impactLabel(delta), Delta: delta, Detail: detail}
}

func dayunPatternPhaseEvidences(r *BaziResult, assessment *NatalAssessment, dy DayunItem) (ProfileEvidence, ProfileEvidence) {
	front := dayunPatternEvidenceForStars(assessment, []string{dy.GanShiShen})
	back := dayunPatternEvidenceForStars(assessment, []string{dy.ZhiShiShen})
	if delta, detail, matched := dayunKillPrintDelta(r, assessment, dy); matched {
		front.Delta += delta
		front.Impact = impactLabel(front.Delta)
		front.Label = "杀印相生"
		front.Detail = detail
	} else if detail != "" {
		if dy.GanShiShen == "七杀" {
			front.Detail += "；" + detail
		}
		if dy.ZhiShiShen == "七杀" {
			back.Detail += "；" + detail
		}
	}
	return front, back
}

func dayunPatternEvidenceForStars(assessment *NatalAssessment, stars []string) ProfileEvidence {
	if assessment == nil || assessment.Pattern.Name == "" || assessment.Pattern.Name == "未定格" {
		return ProfileEvidence{Source: "格局作用", Label: "格局待定", Impact: "中性", Detail: "原局未形成可匹配的格局作用规则，按中性处理。"}
	}
	stars = compactStrings(stars, 0)
	matchedSupport := intersectStrings(stars, assessment.Pattern.SupportTenGods)
	matchedBreak := intersectStrings(stars, assessment.Pattern.BreakTenGods)
	delta := len(matchedSupport)*7 - len(matchedBreak)*8
	switch {
	case len(matchedSupport) > 0 && len(matchedBreak) == 0:
		return ProfileEvidence{Source: "格局作用", Label: "助格", Impact: impactLabel(delta), Delta: clampInt(delta, -16, 16), Detail: fmt.Sprintf("大运十神%s呼应%s的%s条件。", strings.Join(matchedSupport, "、"), assessment.Pattern.Name, strings.Join(assessment.Pattern.Formations, "、"))}
	case len(matchedBreak) > 0 && len(matchedSupport) == 0:
		return ProfileEvidence{Source: "格局作用", Label: "加重格局风险", Impact: impactLabel(delta), Delta: clampInt(delta, -16, 16), Detail: fmt.Sprintf("大运十神%s触及%s的%s风险。", strings.Join(matchedBreak, "、"), assessment.Pattern.Name, strings.Join(assessment.Pattern.Breaks, "、"))}
	case len(matchedSupport) > 0:
		return ProfileEvidence{Source: "格局作用", Label: "格局作用交错", Impact: "中性", Detail: fmt.Sprintf("大运同时见助格%s和风险%s，按相互抵消处理。", strings.Join(matchedSupport, "、"), strings.Join(matchedBreak, "、"))}
	default:
		return ProfileEvidence{Source: "格局作用", Label: "未直接作用主格", Impact: "中性", Detail: fmt.Sprintf("大运十神%s未直接匹配%s已确认的制化或风险条件。", strings.Join(stars, "、"), assessment.Pattern.Name)}
	}
}

func dayunKillPrintDelta(r *BaziResult, assessment *NatalAssessment, dy DayunItem) (int, string, bool) {
	if assessment == nil || assessment.Pattern.Name != "偏印格" {
		return 0, "", false
	}
	if dy.GanShiShen != "七杀" {
		if dy.ZhiShiShen == "七杀" {
			return 0, "七杀仅见大运地支，不按前五年杀印相生计分。", false
		}
		return 0, "", false
	}
	if contains(assessment.Pattern.Formations, "杀印相生") {
		return 0, "原局已计入杀印相生，不重复计分。", false
	}
	stats := natalTenGodStats(r)
	if !stats.hasVisibleAndRooted(r, "偏印") {
		return 0, "偏印未透干有根，不能按杀印相生计分。", false
	}
	dayunWuxing := ganWuxing[dy.Gan]
	for _, gan := range []string{r.YearGan, r.MonthGan, r.HourGan} {
		if GetShiShen(r.DayGan, gan) != "偏印" || rootStrength(gan, r) <= 0 {
			continue
		}
		if wxSheng[dayunWuxing] == ganWuxing[gan] {
			return 7, fmt.Sprintf("大运天干%s为七杀，其%s生原局透干有根的偏印%s，构成杀印相生。", dy.Gan, wxPinyin2CN[dayunWuxing], gan), true
		}
	}
	return 0, "七杀与原局偏印五行未形成相生链，不按杀印相生计分。", false
}

func intersectStrings(left, right []string) []string {
	set := map[string]bool{}
	for _, item := range right {
		set[item] = true
	}
	matched := make([]string, 0, len(left))
	for _, item := range left {
		if set[item] {
			matched = append(matched, item)
		}
	}
	return compactStrings(matched, 0)
}

func dayunTenGodPhaseEvidences(r *BaziResult, dy DayunItem) (ProfileEvidence, ProfileEvidence) {
	if r.ShishenConfidence == "soft" {
		detail := "本命喜忌十神置信度偏软，十神不参与强修正。"
		return ProfileEvidence{Source: "大运十神", Label: nonEmpty(dy.GanShiShen, "无"), Impact: "中性", Detail: detail},
			ProfileEvidence{Source: "大运十神", Label: nonEmpty(dy.ZhiShiShen, "无"), Impact: "中性", Detail: detail}
	}
	return dayunTenGodPhaseEvidence(r, dy.GanShiShen), dayunTenGodPhaseEvidence(r, dy.ZhiShiShen)
}

func dayunTenGodPhaseEvidence(r *BaziResult, tenGod string) ProfileEvidence {
	if tenGod == "" {
		return ProfileEvidence{Source: "大运十神", Label: "无", Impact: "中性", Detail: "大运十神缺失，按中性处理。"}
	}
	rawDelta := 0
	detail := tenGod + "未列入明确喜忌十神"
	switch {
	case contains(r.FavorableShishen, tenGod):
		rawDelta = 5
		detail = tenGod + "为本命偏喜十神"
	case contains(r.AdverseShishen, tenGod):
		rawDelta = -5
		detail = tenGod + "为本命偏忌十神"
	}
	delta := rawDelta
	if r.ShishenConfidence == "hard" {
		delta *= 2
	} else {
		delta = int(math.Round(float64(delta) * 1.4))
	}
	return ProfileEvidence{Source: "大运十神", Label: tenGod, Impact: impactLabel(delta), Delta: clampInt(delta, -20, 20), Detail: detail}
}

func diShiRoadDelta(diShi string) int {
	switch diShi {
	case "帝旺", "临官", "长生", "冠带":
		return 7
	case "沐浴", "养", "胎", "墓":
		return 2
	case "衰", "病", "死", "绝":
		return -6
	default:
		return 0
	}
}

func dayunModifierDelta(dy DayunItem) (int, string) {
	good, watch := 0, 0
	for _, name := range dy.ShenSha {
		if isHelpfulShensha(name) {
			good++
		}
		if isHarshShensha(name) {
			watch++
		}
	}
	delta := clampInt(good*2-watch*2, -5, 5)
	if len(dy.ShenSha) == 0 {
		return 0, "大运未见明显神煞修正。"
	}
	return delta, fmt.Sprintf("大运神煞：%s。吉性%d项，压力%d项。", strings.Join(dy.ShenSha, "、"), good, watch)
}

func vehicleGrade(score int) (string, string) {
	switch {
	case score >= 85:
		return "S", "上格配置"
	case score >= 70:
		return "A", "中上格配置"
	case score >= 50:
		return "B", "中格配置"
	case score >= 30:
		return "C", "中下格配置"
	default:
		return "D", "下格配置"
	}
}

func roadTypeForScore(score int) (string, string) {
	switch {
	case score >= 85:
		return RoadTypeHighway, "高速路"
	case score >= 70:
		return RoadTypeMainRoad, "城市主路"
	case score >= 55:
		return RoadTypeMountainRoad, "山路"
	case score >= 40:
		return RoadTypeMuddyRoad, "泥路"
	default:
		return RoadTypeConstruction, "施工路段"
	}
}

func resolveVehicleType(grade string) string {
	switch grade {
	case "S":
		return "超跑级座驾"
	case "A":
		return "高性能车级座驾"
	case "B":
		return "标准轿车级座驾"
	case "C":
		return "实用 MPV 级座驾"
	default:
		return "基础代步单车级"
	}
}

func resolveDrivingStyle(mingge string) string {
	switch mingge {
	case "七杀格", "月刃格":
		return "动力输出偏强，操控和风险控制要求更高。"
	case "食神格", "伤官格":
		return "响应偏灵活，适合表达、创造和输出型节奏。"
	case "正财格", "偏财格":
		return "资源调度和持续运营倾向更明显。"
	case "正官格", "正印格":
		return "稳定巡航取向更明显，重视规则、耐久和节奏。"
	case "建禄格", "月禄格":
		return "自驱和持续推进倾向更明显，需注意动力分配。"
	case "偏印格":
		return "调校感更强，适合研究和非标准路径。"
	case "杂气格":
		return "多种力量并行，驾驶时更需要精细调校。"
	default:
		return ""
	}
}

func vehicleSummary(grade, vehicleType, yongshen string) string {
	prefix := fmt.Sprintf("%s级%s，", grade, vehicleType)
	switch grade {
	case "S":
		prefix += "调候和扶抑均有稳定支撑，原局基础承载力较强"
	case "A":
		prefix += "调候和扶抑主线成立，原局基础较稳"
	case "B":
		prefix += "原局可用但短板明确，顺运时更容易发挥"
	case "C":
		prefix += "调候或扶抑存在关键不足，发挥更依赖大运补足"
	default:
		prefix += "调候急需或扶抑支撑不足，原局高度依赖大运补救"
	}
	if yongshen != "" {
		return prefix + "；遇到" + yongshen + "相关运势更容易提速。"
	}
	return prefix + "。"
}

func vehicleSummaryFromAssessment(assessment *NatalAssessment, vehicleType string) string {
	if assessment == nil {
		return "原局结构评估待生成。"
	}
	prefix := fmt.Sprintf("%s级%s，", assessment.Grade.Grade, vehicleType)
	thermal := assessment.Tiaohou.Thermal
	if thermal.Status == "" {
		thermal.Status = assessment.Climate.Status
	}
	if thermal.Status == "urgent_unresolved" {
		prefix += "调候急需未解，基础发挥先看补足短板"
	} else if assessment.Pattern.FoundationTier == "high" {
		prefix += "日干调候天透地藏成格，原局具高格基础"
	} else if assessment.Pattern.CriticalBreak {
		prefix += "扶抑可作基础，但格局有关键破损需要优先处理"
	} else {
		prefix += "扶抑为基础，主格结构、制化配合与原局流通共同决定承载层次"
	}
	if assessment.Pattern.FoundationTier == "high" && assessment.Pattern.CriticalBreak {
		prefix += "，但主格仍有关键破损需要优先处理"
	} else if assessment.Pattern.FoundationTier == "high" && assessment.Fuyi.SupportLevel != "扶抑用神得力" {
		prefix += "，但" + assessment.Fuyi.SupportLevel + "，承载仍需看扶抑补足"
	}
	if assessment.Fuyi.Yongshen != "" {
		if len(assessment.YongshenAlignment.Elements) > 0 {
			return prefix + "；" + assessment.YongshenAlignment.Detail + "遇到" + assessment.Fuyi.Yongshen + "相关运势更容易获得支持。"
		}
		return prefix + "；遇到" + assessment.Fuyi.Yongshen + "相关运势更容易获得支持。"
	}
	return prefix + "。"
}

func patternQualityLabel(quality string) string {
	switch quality {
	case "formed":
		return "成"
	case "usable":
		return "可用"
	case "partial":
		return "部分成立"
	case "broken":
		return "受损"
	default:
		return "待定"
	}
}

func relationFlowLabel(flow string) string {
	switch flow {
	case "smooth":
		return "较顺"
	case "partial":
		return "部分连结"

	default:
		return "不足"
	}
}

func dayunRoadSummary(dy DayunItem, roadLabel string, qian, hou RoadPhase) string {
	transition := qian.Label
	if qian.Label != hou.Label {
		transition = qian.Label + "接" + hou.Label
	}
	return fmt.Sprintf("%s%s大运为%s，整体路况%s；前五年看天干%s，后五年看地支%s。",
		dy.Gan, dy.Zhi, roadLabel, transition, dy.Gan, dy.Zhi)
}

func strengthLevelLabel(level string) string {
	switch level {
	case "vstrong":
		return "极旺"
	case "strong":
		return "偏旺"
	case "neutral":
		return "中和"
	case "weak":
		return "偏弱"
	case "vweak":
		return "极弱"
	default:
		return "未明"
	}
}

func shishenConfidenceLabel(confidence string) string {
	switch confidence {
	case "hard":
		return "明确"
	case "medium":
		return "中等"
	case "soft":
		return "偏软"
	default:
		return "未明"
	}
}

func impactLabel(delta int) string {
	switch {
	case delta > 0:
		return "加分"
	case delta < 0:
		return "减分"
	default:
		return "中性"
	}
}

func modifierLabel(delta int) string {
	switch {
	case delta > 0:
		return "助力"
	case delta < 0:
		return "阻力"
	default:
		return "平稳"
	}
}

func isHelpfulShensha(name string) bool {
	switch name {
	case "天乙贵人", "太极贵人", "文昌贵人", "禄神", "天德贵人", "月德贵人", "天德合", "月德合", "德秀贵人", "金舆贵人", "天喜", "天厨贵人", "国印贵人", "三奇贵人", "日德", "将星", "十灵日", "词馆", "福星贵人", "天医":
		return true
	default:
		return false
	}
}

func isHarshShensha(name string) bool {
	switch name {
	case "羊刃", "飞刃", "劫煞", "亡神", "孤辰", "寡宿", "阴差阳错", "魁罡", "十恶大败", "天罗地网", "地网", "童子煞", "灾煞", "流霞", "吊客", "墓门":
		return true
	default:
		return false
	}
}

func compactStrings(items []string, limit int) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" || seen[item] {
			continue
		}
		out = append(out, item)
		seen[item] = true
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	if out == nil {
		return []string{}
	}
	return out
}

func nonEmpty(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func clampInt(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
