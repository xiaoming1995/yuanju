package bazi

import (
	"fmt"
	"sort"
	"strings"
)

const (
	TenGodGroupWealth   = "wealth"
	TenGodGroupOfficial = "official"
	TenGodGroupSeal     = "seal"
	TenGodGroupOutput   = "output"
	TenGodGroupPeer     = "peer"

	TenGodPolaritySupport  = "support"
	TenGodPolarityPressure = "pressure"
	TenGodPolarityMixed    = "mixed"
)

type TenGodGroupInfo struct {
	Group string
	Label string
}

type TenGodPowerProfile struct {
	Dominant   string `json:"dominant"`
	Group      string `json:"group"`
	GroupLabel string `json:"group_label"`
	Strength   string `json:"strength"`
	Polarity   string `json:"polarity"`
	PlainTitle string `json:"plain_title"`
	PlainText  string `json:"plain_text"`
	Score      int    `json:"score"`
	Reason     string `json:"reason,omitempty"`
}

type tenGodCandidate struct {
	TenGod string
	Group  string
	Label  string
	Score  int
	Reason []string
}

func TenGodGroupOf(tenGod string) (TenGodGroupInfo, bool) {
	switch tenGod {
	case "正财", "偏财":
		return TenGodGroupInfo{Group: TenGodGroupWealth, Label: "财星"}, true
	case "正官", "七杀":
		return TenGodGroupInfo{Group: TenGodGroupOfficial, Label: "官杀"}, true
	case "正印", "偏印":
		return TenGodGroupInfo{Group: TenGodGroupSeal, Label: "印星"}, true
	case "食神", "伤官":
		return TenGodGroupInfo{Group: TenGodGroupOutput, Label: "食伤"}, true
	case "比肩", "劫财":
		return TenGodGroupInfo{Group: TenGodGroupPeer, Label: "比劫"}, true
	default:
		return TenGodGroupInfo{}, false
	}
}

func BuildDayunTenGodPower(natal *BaziResult, dy DayunItem) TenGodPowerProfile {
	candidates := map[string]*tenGodCandidate{}
	addTenGodScore(candidates, dy.GanShiShen, 4, "大运天干"+dy.Gan+"为"+dy.GanShiShen)
	addTenGodScore(candidates, dy.ZhiShiShen, 3, "大运地支"+dy.Zhi+"主气为"+dy.ZhiShiShen)
	return buildTenGodProfile(natal, candidates, "")
}

func BuildYearTenGodPower(natal *BaziResult, dy DayunItem, ln LiuNianItem, ctx YearSignalContext, dayunPower TenGodPowerProfile) TenGodPowerProfile {
	candidates := map[string]*tenGodCandidate{}
	addTenGodScore(candidates, ln.GanShiShen, 4, "流年天干为"+ln.GanShiShen)
	addTenGodScore(candidates, ln.ZhiShiShen, 2, "流年地支主气为"+ln.ZhiShiShen)

	if ctx.DayunPhase == DayunPhaseGan {
		boostExistingTenGodGroup(candidates, dy.GanShiShen, 1, "大运前五年同类背景加力", 1)
	} else if ctx.DayunPhase == DayunPhaseZhi {
		boostExistingTenGodGroup(candidates, dy.ZhiShiShen, 1, "大运后五年同类背景加力", 1)
	}

	if dayunPower.Group != "" {
		for _, c := range candidates {
			if c.Group == dayunPower.Group {
				c.Score += 2
				c.Reason = append(c.Reason, "大运同类力量叠加")
				if c.TenGod == dayunPower.Dominant {
					c.Score++
					c.Reason = append(c.Reason, "大运同一十神再临")
				}
			}
		}
	}

	return buildTenGodProfile(natal, candidates, "")
}

func boostExistingTenGodGroup(candidates map[string]*tenGodCandidate, tenGod string, score int, reason string, exactBonus int) {
	info, ok := TenGodGroupOf(tenGod)
	if !ok || score == 0 {
		return
	}
	for _, c := range candidates {
		if c.Group != info.Group {
			continue
		}
		c.Score += score
		if reason != "" {
			c.Reason = append(c.Reason, reason)
		}
		if exactBonus > 0 && c.TenGod == tenGod {
			c.Score += exactBonus
			c.Reason = append(c.Reason, "大运同一十神呼应")
		}
	}
}

func addTenGodScore(candidates map[string]*tenGodCandidate, tenGod string, score int, reason string) {
	info, ok := TenGodGroupOf(tenGod)
	if !ok || score == 0 {
		return
	}
	c, ok := candidates[tenGod]
	if !ok {
		c = &tenGodCandidate{TenGod: tenGod, Group: info.Group, Label: info.Label}
		candidates[tenGod] = c
	}
	c.Score += score
	if reason != "" {
		c.Reason = append(c.Reason, reason)
	}
}

func buildTenGodProfile(natal *BaziResult, candidates map[string]*tenGodCandidate, fallback string) TenGodPowerProfile {
	if len(candidates) == 0 {
		return TenGodPowerProfile{}
	}
	list := make([]*tenGodCandidate, 0, len(candidates))
	for _, c := range candidates {
		list = append(list, c)
	}
	sort.SliceStable(list, func(i, j int) bool {
		if list[i].Score == list[j].Score {
			return list[i].TenGod < list[j].TenGod
		}
		return list[i].Score > list[j].Score
	})
	best := list[0]
	if fallback != "" && best.TenGod == "" {
		best.TenGod = fallback
	}

	polarity := tenGodPolarity(natal, best.Group)
	strength := tenGodStrength(best.Score)
	title := tenGodPlainTitle(best.Group, polarity, strength)
	return TenGodPowerProfile{
		Dominant:   best.TenGod,
		Group:      best.Group,
		GroupLabel: best.Label,
		Strength:   strength,
		Polarity:   polarity,
		PlainTitle: title,
		PlainText:  tenGodPlainText(best.Group, polarity),
		Score:      best.Score,
		Reason:     strings.Join(best.Reason, "；"),
	}
}

func tenGodStrength(score int) string {
	switch {
	case score >= 10:
		return "very_strong"
	case score >= 7:
		return "strong"
	case score >= 4:
		return "medium"
	default:
		return "weak"
	}
}

func tenGodStrengthLabel(strength string) string {
	switch strength {
	case "very_strong":
		return "极旺"
	case "strong":
		return "偏旺"
	case "medium":
		return "有力"
	default:
		return "微显"
	}
}

func tenGodPlainTitle(group, polarity, strength string) string {
	intensity := tenGodEverydayIntensity(strength)
	switch group {
	case TenGodGroupWealth:
		if polarity == TenGodPolarityPressure {
			return "钱财压力" + intensity
		}
		return "钱财资源" + intensity
	case TenGodGroupOfficial:
		if polarity == TenGodPolarityPressure {
			return "规则压力" + intensity
		}
		return "责任机会" + intensity
	case TenGodGroupSeal:
		return "学习贵人" + intensity
	case TenGodGroupOutput:
		if polarity == TenGodPolarityPressure {
			return "表达消耗" + intensity
		}
		return "才华表达" + intensity
	case TenGodGroupPeer:
		if polarity == TenGodPolarityPressure {
			return "同辈竞争" + intensity
		}
		return "同伴助力" + intensity
	default:
		return "年度力量" + intensity
	}
}

func tenGodEverydayIntensity(strength string) string {
	switch strength {
	case "very_strong":
		return "很强"
	case "strong":
		return "明显"
	case "medium":
		return "较明显"
	default:
		return "初显"
	}
}

func tenGodPolarity(natal *BaziResult, group string) string {
	strength := dayMasterStrength(natal)
	switch group {
	case TenGodGroupSeal:
		return TenGodPolaritySupport
	case TenGodGroupOfficial:
		if strength == "weak" {
			return TenGodPolarityPressure
		}
		if strength == "strong" {
			return TenGodPolaritySupport
		}
	case TenGodGroupOutput:
		if strength == "weak" {
			return TenGodPolarityPressure
		}
		if strength == "strong" {
			return TenGodPolaritySupport
		}
	case TenGodGroupPeer:
		if strength == "weak" {
			return TenGodPolaritySupport
		}
		if strength == "strong" {
			return TenGodPolarityPressure
		}
	case TenGodGroupWealth:
		if strength == "weak" {
			return TenGodPolarityPressure
		}
	}
	return TenGodPolarityMixed
}

func tenGodPlainText(group, polarity string) string {
	switch group {
	case TenGodGroupWealth:
		if polarity == TenGodPolarityPressure {
			return "钱财、资源和现实事务更突出，也容易伴随支出压力。"
		}
		return "钱财、资源、合作回报和现实机会更容易被看见。"
	case TenGodGroupOfficial:
		if polarity == TenGodPolarityPressure {
			return "规则、考核、责任和外部压力更明显，宜稳住节奏。"
		}
		return "规则、职位、考试和责任感更明显，适合稳步争取认可。"
	case TenGodGroupSeal:
		return "学习、证书、师长贵人和保护性资源更容易出现。"
	case TenGodGroupOutput:
		if polarity == TenGodPolarityPressure {
			return "表达、才艺和想法变多，但也容易分心或消耗精力。"
		}
		return "表达、创作、才华输出和主动表现的机会更明显。"
	case TenGodGroupPeer:
		if polarity == TenGodPolarityPressure {
			return "同辈竞争、合作分利和人际摩擦会更明显。"
		}
		return "同伴、团队、朋友和协作助力更容易被带动。"
	default:
		return fmt.Sprintf("%s力量较明显。", group)
	}
}
