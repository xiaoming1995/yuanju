package bazi

import (
	"fmt"
	"strings"
)

// FuyiStrengthAssessment is the common regular-strength conclusion for a
// complete natal chart. Tiaohou is intentionally evaluated separately.
type FuyiStrengthAssessment struct {
	Level    string
	Score    int
	Yongshen string
	Jishen   string
	Reason   string
}

const (
	monthCommandSameElementScore     = 12
	monthCommandGeneratesDayScore    = 6
	monthCommandDrainsDayScore       = -6
	monthCommandControlsDayScore     = -8
	monthCommandControlledByDayScore = -6
)

type fuyiPillarWeight struct {
	label         string
	gan           string
	zhi           string
	hidden        []string
	ganWeight     int
	hiddenWeights []int
	rootWeight    int
}

// AssessFuyiStrength evaluates regular Fuyi strength from the complete natal
// chart. Visible stems, hidden-stem order, month command, roots and the
// Jianlu/Yangren structure are all part of the same calculation.
func AssessFuyiStrength(natal *BaziResult) FuyiStrengthAssessment {
	if natal == nil || natal.DayGan == "" {
		return FuyiStrengthAssessment{Level: "neutral"}
	}

	dayWx := ganWuxing[natal.DayGan]
	if dayWx == "" {
		return FuyiStrengthAssessment{Level: "neutral", Reason: "日主五行无法识别"}
	}

	pillars := []fuyiPillarWeight{
		{"年柱", natal.YearGan, natal.YearZhi, fuyiHiddenGans(natal.YearHideGan, natal.YearZhi), 4, []int{6, 3, 1}, 2},
		{"月柱", natal.MonthGan, natal.MonthZhi, fuyiHiddenGans(natal.MonthHideGan, natal.MonthZhi), 6, []int{9, 4, 2}, 4},
		{"日支", "", natal.DayZhi, fuyiHiddenGans(natal.DayHideGan, natal.DayZhi), 0, []int{5, 3, 1}, 3},
		{"时柱", natal.HourGan, natal.HourZhi, fuyiHiddenGans(natal.HourHideGan, natal.HourZhi), 4, []int{6, 3, 1}, 2},
	}

	score := 0
	details := make([]string, 0, 16)
	add := func(label, gan string, weight int) {
		if gan == "" || weight == 0 {
			return
		}
		delta := fuyiElementDelta(dayWx, ganWuxing[gan], weight)
		if delta == 0 {
			return
		}
		score += delta
		details = append(details, fmt.Sprintf("%s%s%s%+d", label, gan, GetShiShen(natal.DayGan, gan), delta))
	}

	for _, pillar := range pillars {
		add(pillar.label+"透", pillar.gan, pillar.ganWeight)
		hasRoot := false
		for index, gan := range pillar.hidden {
			weight := pillar.hiddenWeights[len(pillar.hiddenWeights)-1]
			if index < len(pillar.hiddenWeights) {
				weight = pillar.hiddenWeights[index]
			}
			add(pillar.label+hiddenStemPosition(index), gan, weight)
			if ganWuxing[gan] == dayWx {
				hasRoot = true
			}
		}
		if hasRoot && pillar.rootWeight > 0 {
			score += pillar.rootWeight
			details = append(details, fmt.Sprintf("%s日主有根+%d", pillar.label, pillar.rootWeight))
		}
		if linGuanZhi[natal.DayGan] == pillar.zhi {
			score += 6
			details = append(details, pillar.label+"临官+6")
		}
		if isYangRen(natal.DayGan, pillar.zhi) {
			score += 12
			details = append(details, pillar.label+"羊刃+12")
		}
	}

	monthCommandDelta, monthCommandDetail := fuyiMonthCommandDelta(dayWx, zhiWuxing[natal.MonthZhi])
	if monthCommandDetail != "" {
		score += monthCommandDelta
		details = append(details, monthCommandDetail)
	}

	level := fuyiStrengthLevel(score)
	yongshen, jishen := fuyiElementsForLevel(dayWx, level)
	return FuyiStrengthAssessment{
		Level:    level,
		Score:    score,
		Yongshen: yongshen,
		Jishen:   jishen,
		Reason:   fmt.Sprintf("全局扶抑评分%d：%s", score, strings.Join(details, "；")),
	}
}

// fuyiMonthCommandDelta makes the seasonal relationship explicit instead of
// leaving it implicit in the month branch's hidden-stem subtotal.
func fuyiMonthCommandDelta(dayWx, monthWx string) (int, string) {
	if dayWx == "" || monthWx == "" {
		return 0, ""
	}
	monthLabel := wxPinyin2CN[monthWx]
	switch {
	case dayWx == monthWx:
		return monthCommandSameElementScore, fmt.Sprintf("月令%s得令+%d", monthLabel, monthCommandSameElementScore)
	case wxSheng[monthWx] == dayWx:
		return monthCommandGeneratesDayScore, fmt.Sprintf("月令%s生身+%d", monthLabel, monthCommandGeneratesDayScore)
	case wxSheng[dayWx] == monthWx:
		return monthCommandDrainsDayScore, fmt.Sprintf("月令%s泄身%d", monthLabel, monthCommandDrainsDayScore)
	case wxKe[monthWx] == dayWx:
		return monthCommandControlsDayScore, fmt.Sprintf("月令%s克身%d", monthLabel, monthCommandControlsDayScore)
	case wxKe[dayWx] == monthWx:
		return monthCommandControlledByDayScore, fmt.Sprintf("月令%s受制%d", monthLabel, monthCommandControlledByDayScore)
	default:
		return 0, ""
	}
}

func fuyiHiddenGans(gans []string, zhi string) []string {
	if len(gans) > 0 {
		return gans
	}
	return zhiHideGanFull[zhi]
}

func hiddenStemPosition(index int) string {
	switch index {
	case 0:
		return "本气"
	case 1:
		return "中气"
	default:
		return "余气"
	}
}

func fuyiElementDelta(dayWx, element string, weight int) int {
	switch {
	case element == dayWx || wxSheng[element] == dayWx:
		return weight
	case wxSheng[dayWx] == element:
		return -weight
	case wxKe[element] == dayWx:
		return -weight
	case wxKe[dayWx] == element:
		return -weight
	default:
		return 0
	}
}

func fuyiStrengthLevel(score int) string {
	thresholds := GetAlgoConfig().ShenStrengthThresholds
	if thresholds.VStrong == 0 && thresholds.Strong == 0 && thresholds.Weak == 0 && thresholds.VWeak == 0 {
		thresholds = DefaultShenStrengthThresholds
	}
	switch {
	case score >= thresholds.VStrong:
		return "vstrong"
	case score >= thresholds.Strong:
		return "strong"
	case score <= thresholds.VWeak:
		return "vweak"
	case score <= thresholds.Weak:
		return "weak"
	default:
		return "neutral"
	}
}

func isYangRen(dayGan, zhi string) bool {
	if yangGans[dayGan] {
		return diWangZhi[dayGan] == zhi
	}
	return map[string]string{
		"乙": "辰", "丁": "未", "己": "未", "辛": "戌", "癸": "丑",
	}[dayGan] == zhi
}

func fuyiElementsForLevel(dayWx, level string) (yongshen, jishen string) {
	elements, ok := fuyiElementMap[dayWx]
	if !ok {
		return "", ""
	}
	switch level {
	case "vstrong", "strong":
		// Strong fire is controlled by water. Metal is the supporting source for
		// that control; dry earth is not presented as a general Fuyi remedy.
		if dayWx == "huo" {
			return "水金", "木火"
		}
		return elements[1], elements[0]
	case "vweak", "weak":
		return elements[0], elements[1]
	default:
		return "", ""
	}
}
