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
	Grade       string            `json:"grade"`
	GradeLabel  string            `json:"grade_label"`
	Score       int               `json:"score"`
	VehicleType string            `json:"vehicle_type"`
	Summary     string            `json:"summary"`
	Tags        []string          `json:"tags"`
	Evidences   []ProfileEvidence `json:"evidences"`
}

type RoadPhase struct {
	Key     string `json:"key"`
	Label   string `json:"label"`
	Score   int    `json:"score"`
	Summary string `json:"summary"`
	Detail  string `json:"detail,omitempty"`
}

type DayunRoad struct {
	DayunIndex int               `json:"dayun_index"`
	GanZhi     string            `json:"gan_zhi"`
	Score      int               `json:"score"`
	RoadType   string            `json:"road_type"`
	RoadLabel  string            `json:"road_label"`
	QianRoad   RoadPhase         `json:"qian_road"`
	HouRoad    RoadPhase         `json:"hou_road"`
	Summary    string            `json:"summary"`
	Tags       []string          `json:"tags"`
	Evidences  []ProfileEvidence `json:"evidences"`
}

func BuildVehicleRoadProfile(r *BaziResult) (*VehicleProfile, []DayunRoad) {
	if r == nil {
		return nil, nil
	}
	profile := buildVehicleProfile(r)
	roadmap := buildDayunRoadmap(r)
	return profile, roadmap
}

func buildVehicleProfile(r *BaziResult) *VehicleProfile {
	score := 50
	evidences := make([]ProfileEvidence, 0, 6)

	strengthLevel, strengthScore, strengthDetail := dayMasterStrengthLevel(r)
	strengthDelta := vehicleStrengthDelta(strengthLevel)
	score += strengthDelta
	evidences = append(evidences, ProfileEvidence{
		Source: "日主强弱",
		Label:  strengthLevelLabel(strengthLevel),
		Impact: impactLabel(strengthDelta),
		Delta:  strengthDelta,
		Detail: fmt.Sprintf("身强弱%s，算法%s", strengthLevelLabel(strengthLevel), strengthDetail),
	})

	tiaohouDelta, tiaohouLabel, tiaohouDetail := vehicleTiaohouDelta(r.Tiaohou)
	score += tiaohouDelta
	evidences = append(evidences, ProfileEvidence{
		Source: "调候",
		Label:  tiaohouLabel,
		Impact: impactLabel(tiaohouDelta),
		Delta:  tiaohouDelta,
		Detail: tiaohouDetail,
	})

	minggeDelta := vehicleMinggeDelta(r.MingGe)
	score += minggeDelta
	evidences = append(evidences, ProfileEvidence{
		Source: "命格",
		Label:  nonEmpty(r.MingGe, "未定格"),
		Impact: impactLabel(minggeDelta),
		Delta:  minggeDelta,
		Detail: vehicleMinggeDetail(r.MingGe),
	})

	confDelta := vehicleConfidenceDelta(r.ShishenConfidence, r.Yongshen, r.Jishen)
	score += confDelta
	evidences = append(evidences, ProfileEvidence{
		Source: "喜忌十神",
		Label:  shishenConfidenceLabel(r.ShishenConfidence),
		Impact: impactLabel(confDelta),
		Delta:  confDelta,
		Detail: fmt.Sprintf("喜用%s，忌%s，十神置信度%s", nonEmpty(r.Yongshen, "待判定"), nonEmpty(r.Jishen, "待判定"), shishenConfidenceLabel(r.ShishenConfidence)),
	})

	riskDelta, riskLabel, riskDetail := vehicleNatalModifierDelta(r)
	score += riskDelta
	evidences = append(evidences, ProfileEvidence{
		Source: "原局修正",
		Label:  riskLabel,
		Impact: impactLabel(riskDelta),
		Delta:  riskDelta,
		Detail: riskDetail,
	})

	score = clampInt(score, 0, 100)
	grade, gradeLabel := vehicleGrade(score)
	vehicleType := resolveVehicleType(r, strengthLevel, strengthScore)
	tags := []string{gradeLabel, vehicleType, strengthLevelLabel(strengthLevel)}
	if r.MingGe != "" {
		tags = append(tags, r.MingGe)
	}
	if r.Yongshen != "" {
		tags = append(tags, "喜用"+r.Yongshen)
	}

	return &VehicleProfile{
		Grade:       grade,
		GradeLabel:  gradeLabel,
		Score:       score,
		VehicleType: vehicleType,
		Summary:     vehicleSummary(grade, vehicleType, r.Yongshen),
		Tags:        compactStrings(tags, 5),
		Evidences:   evidences,
	}
}

func buildDayunRoadmap(r *BaziResult) []DayunRoad {
	if len(r.Dayun) == 0 {
		return nil
	}
	out := make([]DayunRoad, 0, len(r.Dayun))
	for _, dy := range r.Dayun {
		score := 50
		evidences := make([]ProfileEvidence, 0, 5)

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

		wxDelta, wxDetail := dayunElementDelta(r, dy)
		score += wxDelta
		evidences = append(evidences, ProfileEvidence{
			Source: "大运五行",
			Label:  dy.Gan + dy.Zhi,
			Impact: impactLabel(wxDelta),
			Delta:  wxDelta,
			Detail: wxDetail,
		})

		tenGodDelta, tenGodDetail := dayunTenGodDelta(r, dy)
		score += tenGodDelta
		evidences = append(evidences, ProfileEvidence{
			Source: "大运十神",
			Label:  dy.GanShiShen + "/" + dy.ZhiShiShen,
			Impact: impactLabel(tenGodDelta),
			Delta:  tenGodDelta,
			Detail: tenGodDetail,
		})

		diShiDelta := diShiRoadDelta(dy.DiShi)
		score += diShiDelta
		evidences = append(evidences, ProfileEvidence{
			Source: "十二长生",
			Label:  nonEmpty(dy.DiShi, "无"),
			Impact: impactLabel(diShiDelta),
			Delta:  diShiDelta,
			Detail: "大运地支对应日主十二长生为" + nonEmpty(dy.DiShi, "无"),
		})

		modDelta, modDetail := dayunModifierDelta(dy)
		score += modDelta
		evidences = append(evidences, ProfileEvidence{
			Source: "神煞修正",
			Label:  modifierLabel(modDelta),
			Impact: impactLabel(modDelta),
			Delta:  modDelta,
			Detail: modDetail,
		})

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
		})
	}
	return out
}

func vehicleStrengthDelta(level string) int {
	switch level {
	case "neutral":
		return 18
	case "strong", "weak":
		return 11
	case "vstrong", "vweak":
		return 4
	default:
		return 8
	}
}

func vehicleTiaohouDelta(t *TiaohouResult) (int, string, string) {
	if t == nil || len(t.Expected) == 0 {
		return 5, "调候未命中字典", "暂无明确调候规则，按中性略保守处理。"
	}
	tou := len(t.Tou)
	cang := len(t.Cang)
	expected := len(t.Expected)
	switch {
	case tou >= expected:
		return 22, "调候透干齐全", fmt.Sprintf("理论所需%s，透干%s。", strings.Join(t.Expected, "、"), strings.Join(t.Tou, "、"))
	case tou > 0:
		return 16, "调候部分透干", fmt.Sprintf("理论所需%s，透干%s，仍有部分需行运引动。", strings.Join(t.Expected, "、"), strings.Join(t.Tou, "、"))
	case cang > 0:
		return 10, "调候藏支待引", fmt.Sprintf("理论所需%s，藏于%s，力量需大运引出。", strings.Join(t.Expected, "、"), strings.Join(t.Cang, "、"))
	default:
		return 2, "调候缺位", fmt.Sprintf("理论所需%s，原局未见明显透藏。", strings.Join(t.Expected, "、"))
	}
}

func vehicleMinggeDelta(mingge string) int {
	if strings.TrimSpace(mingge) == "" {
		return 6
	}
	if mingge == "杂气格" {
		return 9
	}
	return 16
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

func vehicleConfidenceDelta(confidence, yongshen, jishen string) int {
	if yongshen == "" && jishen == "" {
		return 2
	}
	switch confidence {
	case "hard":
		return 14
	case "medium":
		return 10
	case "soft":
		return 6
	default:
		return 8
	}
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
	delta := clampInt(good*2-watch*3, -8, 8)
	label := "修正平稳"
	if delta > 0 {
		label = "助力较多"
	} else if delta < 0 {
		label = "阻力较多"
	}
	return delta, label, fmt.Sprintf("吉性神煞%d项，压力神煞%d项，按辅助层小幅修正。", good, watch)
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

func dayunElementDelta(r *BaziResult, dy DayunItem) (int, string) {
	ganCN := wxPinyin2CN[ganWuxing[dy.Gan]]
	zhiCN := wxPinyin2CN[zhiWuxing[dy.Zhi]]
	delta := 0
	parts := make([]string, 0, 2)
	for _, item := range []struct {
		label string
		wx    string
	}{
		{"天干" + dy.Gan, ganCN},
		{"地支" + dy.Zhi, zhiCN},
	} {
		if item.wx == "" {
			continue
		}
		switch {
		case r.Yongshen != "" && strings.Contains(r.Yongshen, item.wx):
			delta += 7
			parts = append(parts, item.label+"为喜用"+item.wx)
		case r.Jishen != "" && strings.Contains(r.Jishen, item.wx):
			delta -= 7
			parts = append(parts, item.label+"为忌"+item.wx)
		default:
			parts = append(parts, item.label+"为中性"+item.wx)
		}
	}
	if len(parts) == 0 {
		return 0, "大运五行无法识别，按中性处理。"
	}
	return clampInt(delta, -13, 13), strings.Join(parts, "；")
}

func dayunTenGodDelta(r *BaziResult, dy DayunItem) (int, string) {
	if r.ShishenConfidence == "soft" {
		return 0, "本命喜忌十神置信度偏软，十神不参与强修正。"
	}
	delta := 0
	parts := make([]string, 0, 2)
	for _, tenGod := range []string{dy.GanShiShen, dy.ZhiShiShen} {
		if tenGod == "" {
			continue
		}
		switch {
		case contains(r.FavorableShishen, tenGod):
			delta += 5
			parts = append(parts, tenGod+"为本命偏喜十神")
		case contains(r.AdverseShishen, tenGod):
			delta -= 5
			parts = append(parts, tenGod+"为本命偏忌十神")
		default:
			parts = append(parts, tenGod+"未列入明确喜忌十神")
		}
	}
	if r.ShishenConfidence == "hard" {
		delta *= 2
	} else {
		delta = int(math.Round(float64(delta) * 1.4))
	}
	if len(parts) == 0 {
		return 0, "大运十神缺失，按中性处理。"
	}
	return clampInt(delta, -20, 20), strings.Join(parts, "；")
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
	case score >= 90:
		return "S", "协同型配置"
	case score >= 75:
		return "A", "稳健型配置"
	case score >= 60:
		return "B", "实用型配置"
	case score >= 45:
		return "C", "特性型配置"
	default:
		return "D", "调校型配置"
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

func resolveVehicleType(r *BaziResult, strengthLevel string, strengthScore int) string {
	if r == nil {
		return "基础通勤型"
	}
	switch {
	case r.MingGe == "七杀格" || r.MingGe == "月刃格":
		return "高扭矩越野型"
	case r.MingGe == "正官格" || r.MingGe == "正印格":
		return "稳定商务型"
	case r.MingGe == "食神格" || r.MingGe == "伤官格":
		return "灵感跑车型"
	case r.MingGe == "正财格" || r.MingGe == "偏财格":
		return "资源运营型"
	case strings.HasPrefix(strengthLevel, "v") || math.Abs(float64(strengthScore)) >= 12:
		return "重载工程型"
	default:
		return "均衡通勤型"
	}
}

func vehicleSummary(grade, vehicleType, yongshen string) string {
	prefix := fmt.Sprintf("%s级%s，", grade, vehicleType)
	switch grade {
	case "S", "A":
		prefix += "原局主线较清楚，配置完整度高"
	case "B":
		prefix += "原局可用性稳定，适合借助顺运发挥"
	case "C":
		prefix += "原局优势和短板都明显，需要看路况驾驶"
	default:
		prefix += "原局更依赖后天调校和大运补足"
	}
	if yongshen != "" {
		return prefix + "；遇到" + yongshen + "相关运势更容易提速。"
	}
	return prefix + "。"
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
