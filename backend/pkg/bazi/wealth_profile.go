package bazi

import (
	"fmt"
	"strings"
	"time"
)

const WealthProfileVersion = "v1"

type WealthWindowHint struct {
	Year       int      `json:"year"`
	DayunIndex int      `json:"dayun_index"`
	GanZhi     string   `json:"gan_zhi"`
	Level      string   `json:"level"`
	Label      string   `json:"label"`
	Summary    string   `json:"summary"`
	Evidences  []string `json:"evidences"`
}

type WealthProfile struct {
	Version     string            `json:"version"`
	Grade       string            `json:"grade"`
	GradeLabel  string            `json:"grade_label"`
	Score       int               `json:"score"`
	WealthType  string            `json:"wealth_type"`
	Summary     string            `json:"summary"`
	Tags        []string          `json:"tags"`
	RiskFlags   []string          `json:"risk_flags"`
	Evidences   []ProfileEvidence `json:"evidences"`
	CurrentHint *WealthWindowHint `json:"current_hint,omitempty"`
}

func currentGregorianYear() int {
	return time.Now().Year()
}

func EnsureWealthProfile(r *BaziResult, currentYear int) bool {
	if r == nil || r.DayGan == "" {
		return false
	}
	if currentYear == 0 {
		currentYear = currentGregorianYear()
	}
	if r.WealthProfile != nil && r.WealthProfile.Version == WealthProfileVersion {
		expectedHint := BuildCurrentWealthWindowHint(r, currentYear)
		if expectedHint == nil && r.WealthProfile.CurrentHint == nil {
			return false
		}
		if expectedHint != nil && r.WealthProfile.CurrentHint != nil && r.WealthProfile.CurrentHint.Year == currentYear {
			return false
		}
	}
	r.WealthProfile = BuildWealthProfile(r, currentYear)
	return r.WealthProfile != nil
}

func BuildWealthProfile(r *BaziResult, currentYear int) *WealthProfile {
	if r == nil || r.DayGan == "" {
		return nil
	}
	EnsureNatalAssessment(r)
	stats := natalTenGodStats(r)
	evidences := make([]ProfileEvidence, 0, 6)
	tags := make([]string, 0, 6)
	risks := make([]string, 0, 4)

	wealthVisible := stats["正财"].visible + stats["偏财"].visible
	wealthHidden := stats["正财"].hidden + stats["偏财"].hidden
	wealthRoot := maxInt(rootStrengthForTenGod(r, "正财"), rootStrengthForTenGod(r, "偏财"))
	wealthTotal := wealthVisible + wealthHidden

	score := 40
	wealthDelta := 0
	switch {
	case wealthVisible >= 2 && wealthRoot >= 4:
		wealthDelta = 24
		tags = append(tags, "财星透出有根")
	case wealthVisible >= 1 && wealthRoot > 0:
		wealthDelta = 18
		tags = append(tags, "财星透出有根")
	case wealthVisible >= 1:
		wealthDelta = 12
		tags = append(tags, "财星透出")
	case wealthHidden >= 2:
		wealthDelta = 10
		tags = append(tags, "财星藏支")
	case wealthHidden >= 1:
		wealthDelta = 6
		tags = append(tags, "财星微显")
	default:
		wealthDelta = -10
		risks = append(risks, "原局财星不显")
	}
	score += wealthDelta
	evidences = append(evidences, ProfileEvidence{
		Source: "财星显露",
		Label:  wealthVisibilityLabel(wealthVisible, wealthHidden, wealthRoot),
		Impact: impactLabel(wealthDelta),
		Delta:  wealthDelta,
		Detail: fmt.Sprintf("正财透干%d处、藏支%d处；偏财透干%d处、藏支%d处；最强财根%d。", stats["正财"].visible, stats["正财"].hidden, stats["偏财"].visible, stats["偏财"].hidden, wealthRoot),
	})

	carryDelta, carryLabel, carryRisk := wealthCarryingDelta(r, wealthTotal)
	score += carryDelta
	tags = append(tags, carryLabel)
	if carryRisk != "" {
		risks = append(risks, carryRisk)
	}
	evidences = append(evidences, ProfileEvidence{
		Source: "承载能力",
		Label:  carryLabel,
		Impact: impactLabel(carryDelta),
		Delta:  carryDelta,
		Detail: wealthCarryingDetail(r, wealthTotal),
	})

	chainDelta, chainTags, chainDetail := wealthChainDelta(r)
	score += chainDelta
	tags = append(tags, chainTags...)
	evidences = append(evidences, ProfileEvidence{
		Source: "生财链路",
		Label:  nonEmpty(strings.Join(chainTags, "、"), "链路平稳"),
		Impact: impactLabel(chainDelta),
		Delta:  chainDelta,
		Detail: chainDetail,
	})

	favorDelta, favorLabel, favorRisk := wealthFavorabilityDelta(r)
	score += favorDelta
	tags = append(tags, favorLabel)
	if favorRisk != "" {
		risks = append(risks, favorRisk)
	}
	evidences = append(evidences, ProfileEvidence{
		Source: "喜忌方向",
		Label:  favorLabel,
		Impact: impactLabel(favorDelta),
		Delta:  favorDelta,
		Detail: fmt.Sprintf("正财/偏财在本命喜忌十神中校验；当前置信度=%s。", nonEmpty(r.ShishenConfidence, "未定")),
	})

	riskDelta, riskTags, riskFlags, riskDetail := wealthRetentionRiskDelta(r, stats)
	score += riskDelta
	tags = append(tags, riskTags...)
	risks = append(risks, riskFlags...)
	evidences = append(evidences, ProfileEvidence{
		Source: "守财风险",
		Label:  nonEmpty(strings.Join(riskTags, "、"), "风险平稳"),
		Impact: impactLabel(riskDelta),
		Delta:  riskDelta,
		Detail: riskDetail,
	})

	score = clampInt(score, 0, 100)
	ceiling := wealthGradeCeiling(r, wealthTotal, risks)
	grade, gradeLabel := wealthGradeWithCeiling(score, ceiling)
	return &WealthProfile{
		Version:     WealthProfileVersion,
		Grade:       grade,
		GradeLabel:  gradeLabel,
		Score:       score,
		WealthType:  wealthTypeFromSignals(stats, r),
		Summary:     wealthSummary(grade, wealthTotal, risks),
		Tags:        compactStrings(tags, 5),
		RiskFlags:   compactStrings(risks, 4),
		Evidences:   evidences,
		CurrentHint: BuildCurrentWealthWindowHint(r, currentYear),
	}
}

func BuildCurrentWealthWindowHint(r *BaziResult, currentYear int) *WealthWindowHint {
	if r == nil || currentYear == 0 || len(r.Dayun) == 0 {
		return nil
	}
	var dy *DayunItem
	for i := range r.Dayun {
		if currentYear >= r.Dayun[i].StartYear && currentYear <= r.Dayun[i].EndYear {
			dy = &r.Dayun[i]
			break
		}
	}
	if dy == nil {
		return nil
	}
	evidences := make([]string, 0, 4)
	score := 0
	if isWealthTenGod(dy.GanShiShen) {
		score += 2
		evidences = append(evidences, "大运天干"+dy.Gan+"为"+dy.GanShiShen)
	}
	if isWealthTenGod(dy.ZhiShiShen) {
		score += 2
		evidences = append(evidences, "大运地支"+dy.Zhi+"主气为"+dy.ZhiShiShen)
	}
	power := BuildDayunTenGodPower(r, *dy)
	if power.Group == TenGodGroupWealth {
		score += 2
		evidences = append(evidences, power.PlainTitle)
		if power.Polarity == TenGodPolarityPressure {
			score--
			evidences = append(evidences, "财务资源伴随压力")
		}
	}
	if road := findDayunRoadForHint(r, dy.Index); road != nil {
		if road.Score >= 70 {
			score++
			evidences = append(evidences, "当前大运路况"+road.RoadLabel)
		} else if road.Score < 55 {
			score--
			evidences = append(evidences, "当前大运路况"+road.RoadLabel)
		}
	}
	if score <= 0 && len(evidences) == 0 {
		return nil
	}
	level, label := wealthHintLevel(score)
	return &WealthWindowHint{
		Year:       currentYear,
		DayunIndex: dy.Index,
		GanZhi:     dy.Gan + dy.Zhi,
		Level:      level,
		Label:      label,
		Summary:    wealthHintSummary(level, dy.Gan+dy.Zhi),
		Evidences:  compactStrings(evidences, 4),
	}
}

func isWealthTenGod(tenGod string) bool {
	return tenGod == "正财" || tenGod == "偏财"
}

func wealthVisibilityLabel(visible, hidden, root int) string {
	switch {
	case visible > 0 && root > 0:
		return "财星透出有根"
	case visible > 0:
		return "财星透出"
	case hidden > 0:
		return "财星藏支"
	default:
		return "财星不显"
	}
}

func wealthCarryingDelta(r *BaziResult, wealthTotal int) (int, string, string) {
	level, _, _ := dayMasterStrengthLevel(r)
	switch level {
	case "neutral":
		return 12, "承载平衡", ""
	case "strong":
		return 10, "承载有力", ""
	case "vstrong":
		return 2, "身旺需流通", "身旺财弱需见疏泄"
	case "weak":
		if wealthTotal >= 3 {
			return -14, "财重身弱", "财多身弱"
		}
		return -4, "承载偏弱", "承载偏弱"
	case "vweak":
		return -20, "承载薄弱", "身弱难承财"
	default:
		return 0, "承载待定", ""
	}
}

func wealthCarryingDetail(r *BaziResult, wealthTotal int) string {
	level, score, detail := dayMasterStrengthLevel(r)
	return fmt.Sprintf("日主强弱=%s（评分%d），财星合计%d处。%s", strengthLevelLabel(level), score, wealthTotal, detail)
}

func wealthChainDelta(r *BaziResult) (int, []string, string) {
	stats := natalTenGodStats(r)
	tags := []string{}
	delta := 0
	if (stats.hasVisible("食神") || stats.hasVisible("伤官") || stats.hasHidden("食神", "伤官")) && stats.hasAny("正财", "偏财") {
		delta += 10
		tags = append(tags, "食伤生财")
	}
	if stats.hasAny("正财", "偏财") && stats.hasAny("正官", "七杀") {
		delta += 7
		tags = append(tags, "财官相生")
	}
	if r.NatalAssessment != nil {
		if r.NatalAssessment.Pattern.Name == "正财格" || r.NatalAssessment.Pattern.Name == "偏财格" {
			switch r.NatalAssessment.Pattern.Quality {
			case "formed", "usable":
				delta += 8
				tags = append(tags, r.NatalAssessment.Pattern.Name+"可用")
			case "partial":
				delta += 3
				tags = append(tags, r.NatalAssessment.Pattern.Name+"部分成立")
			case "broken":
				delta -= 8
				tags = append(tags, r.NatalAssessment.Pattern.Name+"受损")
			}
		}
	}
	if len(tags) == 0 {
		return 0, tags, "原局未见明确食伤生财、财官相生或财格成败信号，按中性处理。"
	}
	return clampInt(delta, -12, 24), tags, "识别到" + strings.Join(tags, "、") + "，作为财富资源流通与承接链路。"
}

func wealthFavorabilityDelta(r *BaziResult) (int, string, string) {
	if r.ShishenConfidence == "soft" {
		return 0, "喜忌十神置信度偏软", ""
	}
	fav := contains(r.FavorableShishen, "正财") || contains(r.FavorableShishen, "偏财")
	adv := contains(r.AdverseShishen, "正财") || contains(r.AdverseShishen, "偏财")
	switch {
	case fav && !adv:
		return 12, "财星为喜", ""
	case adv && !fav:
		return -14, "财星为忌", "财星为忌"
	case fav && adv:
		return 0, "财星喜忌交错", "财星喜忌交错"
	default:
		return 0, "财星喜忌未定", ""
	}
}

func wealthRetentionRiskDelta(r *BaziResult, stats tenGodStats) (int, []string, []string, string) {
	delta := 0
	tags := []string{}
	risks := []string{}
	if stats.hasVisibleAndRooted(r, "比肩", "劫财") && stats.hasAny("正财", "偏财") {
		delta -= 12
		tags = append(tags, "比劫夺财需防")
		risks = append(risks, "同辈竞争或合作分利")
	}
	if r.NatalAssessment != nil {
		if contains(r.NatalAssessment.Pattern.Breaks, "财多身弱") {
			delta -= 16
			tags = append(tags, "财多身弱")
			risks = append(risks, "财务压力偏重")
		}
		if contains(r.NatalAssessment.Pattern.Breaks, "比劫夺财") {
			delta -= 10
			tags = append(tags, "比劫夺财")
			risks = append(risks, "资源分散")
		}
		if contains(r.NatalAssessment.Pattern.Breaks, "财坏印") {
			delta -= 8
			tags = append(tags, "财坏印")
			risks = append(risks, "资源压过稳定支撑")
		}
	}
	if len(tags) == 0 {
		return 4, []string{"守财风险平稳"}, risks, "未见明确比劫夺财、财多身弱或财坏印等硬风险。"
	}
	return clampInt(delta, -28, 4), tags, risks, "识别到" + strings.Join(tags, "、") + "，财富结构需重视支出、分利或承载问题。"
}

func wealthGradeCeiling(r *BaziResult, wealthTotal int, risks []string) string {
	level, _, _ := dayMasterStrengthLevel(r)
	if level == "vweak" && wealthTotal > 0 {
		return "C"
	}
	if contains(risks, "财多身弱") || contains(risks, "身弱难承财") || contains(risks, "财星为忌") {
		return "B"
	}
	return "S"
}

func wealthGradeWithCeiling(score int, ceiling string) (string, string) {
	grade, label := wealthGrade(score)
	grade = stricterVehicleGrade(grade, ceiling)
	_, label = wealthGrade(gradeMinimumScore(grade))
	return grade, label
}

func wealthGrade(score int) (string, string) {
	switch {
	case score >= 85:
		return "S", "财富结构通达"
	case score >= 70:
		return "A", "财富承接较好"
	case score >= 50:
		return "B", "财富线索可用"
	case score >= 30:
		return "C", "财富波动偏多"
	default:
		return "D", "财富承接薄弱"
	}
}

func wealthTypeFromSignals(stats tenGodStats, r *BaziResult) string {
	switch {
	case r != nil && r.NatalAssessment != nil && (r.NatalAssessment.Pattern.Name == "正财格" || r.NatalAssessment.Pattern.Name == "偏财格"):
		return r.NatalAssessment.Pattern.Name + "资源型"
	case stats["正财"].total() >= stats["偏财"].total() && stats["正财"].total() > 0:
		return "稳定经营型"
	case stats["偏财"].total() > 0:
		return "机会资源型"
	case stats.hasAny("食神", "伤官"):
		return "输出变现型"
	default:
		return "资源待引型"
	}
}

func wealthSummary(grade string, wealthTotal int, risks []string) string {
	base := map[string]string{
		"S": "钱财资源线索清晰，承载与流通配合较好，适合以长期经营和风险控制来放大优势。",
		"A": "钱财资源承接较好，原局有可用财富线索，发挥程度仍要看节奏与行运引动。",
		"B": "财富线索可用但并非一路通达，适合先看清资源来源、承载能力和支出边界。",
		"C": "财富结构波动偏多，钱财机会与压力容易并见，宜以稳守、分散风险和节奏控制为主。",
		"D": "财富承接偏薄，原局更需要后天能力、稳定现金流和顺运补足。",
	}[grade]
	if wealthTotal == 0 {
		base = "原局财星不显，财富主题更依赖后天能力、行业选择和行运引动。"
	}
	if len(risks) > 0 {
		base += " 需留意：" + strings.Join(compactStrings(risks, 2), "、") + "。"
	}
	return base
}

func findDayunRoadForHint(r *BaziResult, dayunIndex int) *DayunRoad {
	for i := range r.DayunRoadmap {
		if r.DayunRoadmap[i].DayunIndex == dayunIndex {
			return &r.DayunRoadmap[i]
		}
	}
	return nil
}

func wealthHintLevel(score int) (string, string) {
	switch {
	case score >= 5:
		return "strong", "当前钱财资源较明显"
	case score >= 3:
		return "medium", "当前钱财资源被带动"
	default:
		return "watch", "当前钱财主题需观察"
	}
}

func wealthHintSummary(level, ganZhi string) string {
	switch level {
	case "strong":
		return ganZhi + "大运钱财资源主题较明显，但仍须结合路况与风险边界，不等于现实收益保证。"
	case "medium":
		return ganZhi + "大运会带动钱财资源议题，适合看承接方式和节奏。"
	default:
		return ganZhi + "大运有钱财相关信号，也伴随压力或阻力，宜谨慎观察。"
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
