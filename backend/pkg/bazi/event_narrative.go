package bazi

import "strings"

// RenderYearNarrative 根据 EventSignal 列表生成面向用户的白话批语。
// 底层 Evidence 保留给 RenderEvidenceSummary，不直接暴露在默认正文中。
func RenderYearNarrative(ys YearSignals) string {
	if len(ys.Signals) == 0 {
		return "本年命理信号较弱，运势相对平稳，无明显重大变动。"
	}

	primary, ok := pickDominantSignal(ys.Signals, "")
	if !ok {
		return ys.GanZhi + "年整体动象不强，适合按部就班推进，保持稳定节奏。"
	}
	secondary, hasSecondary := pickDominantSignal(ys.Signals, themeOf(primary.Type))

	var parts []string
	parts = append(parts, ys.GanZhi+"年，"+plainThemeSentence(primary, true))
	if hasSecondary {
		parts = append(parts, plainThemeSentence(secondary, false))
	}
	parts = append(parts, practicalReminder(ys.Signals))

	return strings.Join(parts, "")
}

// RenderEvidenceSummary 提取专业用户可展开查看的命理依据。
func RenderEvidenceSummary(ys YearSignals) []string {
	if len(ys.Signals) == 0 {
		return nil
	}
	pools := [][]EventSignal{
		filterEvidenceSignals(ys.Signals, PolarityXiong, false),
		filterEvidenceSignals(ys.Signals, PolarityJi, false),
		filterEvidenceSignals(ys.Signals, PolarityXiong, true),
		filterEvidenceSignals(ys.Signals, PolarityNeutral, false),
		filterEvidenceSignals(ys.Signals, PolarityNeutral, true),
		filterEvidenceSignals(ys.Signals, PolarityJi, true),
	}
	out := make([]string, 0, 5)
	seen := map[string]bool{}
	for _, pool := range pools {
		for _, s := range pool {
			ev := compactEvidence(s.Evidence)
			if ev == "" || seen[ev] {
				continue
			}
			seen[ev] = true
			out = append(out, ev)
			if len(out) >= 5 {
				return out
			}
		}
	}
	return out
}

func pickDominantSignal(signals []EventSignal, excludeTheme string) (EventSignal, bool) {
	var best EventSignal
	found := false
	bestRank := 999
	bestPol := 999
	for _, s := range signals {
		if s.Type == "用神基底" {
			continue
		}
		theme := themeOf(s.Type)
		if theme == "" || theme == excludeTheme {
			continue
		}
		rank := themeRank(theme, s)
		pol := polarityRank(s.Polarity)
		if !found || rank < bestRank || (rank == bestRank && pol < bestPol) {
			best = s
			found = true
			bestRank = rank
			bestPol = pol
		}
	}
	return best, found
}

func themeOf(typ string) string {
	switch typ {
	case "综合变动", "伏吟", "反吟", "大运合化", TypeJuShiZhong:
		return "change"
	case "健康":
		return "health"
	case "婚恋_合", "婚恋_冲", "婚恋_变", TypeXingGeQingYi, TypeXingGePanNi:
		return "relationship"
	case "事业", TypeXueYeYaLi, TypeXueYeGuiRen, TypeXueYeCaiYi, TypeXueYeJingZheng:
		return "career"
	case "财运_得", "财运_损", TypeXueYeZiYuan:
		return "money"
	case "迁变":
		return "movement"
	case "喜神临运":
		return "support"
	default:
		return ""
	}
}

func themeRank(theme string, sig EventSignal) int {
	switch theme {
	case "change":
		return 1
	case "health":
		return 2
	case "relationship":
		return 3
	case "career":
		return 4
	case "money":
		return 5
	case "movement":
		return 6
	case "support":
		return 7
	default:
		return 99
	}
}

func polarityRank(p string) int {
	switch p {
	case PolarityXiong:
		return 1
	case PolarityJi:
		return 2
	default:
		return 3
	}
}

func plainThemeSentence(sig EventSignal, primary bool) string {
	prefix := ""
	if !primary {
		prefix = "同时，"
	}
	switch themeOf(sig.Type) {
	case "change":
		if sig.Polarity == PolarityJi {
			return prefix + "事情容易被推动起来，适合顺势整理方向，把该推进的安排往前推一推。"
		}
		return prefix + "变化感会比较强，旧问题或新安排容易集中出现，生活节奏可能比平时更紧。"
	case "health":
		return prefix + "健康、睡眠和出行安全需要多留心，避免冒险、熬夜和硬扛。"
	case "relationship":
		if sig.Type == TypeXingGePanNi {
			return prefix + "情绪和家庭沟通容易起波动，越是着急表达，越要放慢语气。"
		}
		if strings.HasPrefix(sig.Type, "性格_") {
			return prefix + "人际关系会更有存在感，同学朋友之间既有靠近，也要注意分寸。"
		}
		if sig.Polarity == PolarityXiong || sig.Type == "婚恋_冲" || sig.Type == "婚恋_变" {
			return prefix + "感情和人际关系容易被触动，适合把话说清楚，不宜急着做重决定。"
		}
		return prefix + "感情和人缘有被带动的迹象，主动沟通会更容易打开局面。"
	case "career":
		if strings.HasPrefix(sig.Type, "学业_") {
			if sig.Polarity == PolarityXiong {
				return prefix + "学习压力会更明显，考试、规则或师长要求容易让人紧绷。"
			}
			return prefix + "学习上容易得到帮助或看到新的兴趣方向，适合稳住节奏继续积累。"
		}
		if sig.Polarity == PolarityXiong {
			return prefix + "工作方向、职责或合作关系容易出现调整，需要重新适应节奏。"
		}
		return prefix + "工作上容易出现助力或机会，适合主动承担、争取表现。"
	case "money":
		if sig.Polarity == PolarityXiong || sig.Type == "财运_损" {
			return prefix + "钱财机会看似增加，但也容易伴随支出和压力，宜稳健安排。"
		}
		if sig.Type == TypeXueYeZiYuan {
			return prefix + "家庭资源或学习投入会更受关注，适合把条件用在真正有帮助的地方。"
		}
		return prefix + "财务和资源面有可把握的空间，但仍要量力而行。"
	case "movement":
		return prefix + "出行、搬动、岗位或生活环境有变化机会，提前规划会更从容。"
	case "support":
		return prefix + "外部助力会比平时明显，适合借势推进重要事项。"
	default:
		return prefix + "整体运势有起伏，宜多观察、少冒进。"
	}
}

func practicalReminder(signals []EventSignal) string {
	xiong, ji := 0, 0
	for _, s := range signals {
		switch s.Polarity {
		case PolarityXiong:
			xiong++
		case PolarityJi:
			ji++
		}
	}
	switch {
	case xiong >= 2 && ji > 0:
		return "这一年有机会也有压力，取舍要清楚，先稳住再求突破。"
	case xiong >= 2:
		return "这一年宜保守谨慎，少做冲动决定，把风险降下来更重要。"
	case ji >= 2:
		return "这一年可以主动一些，把握机会，但仍要保持节制。"
	case xiong > 0:
		return "遇事不必急着硬碰硬，留出余地会更稳。"
	default:
		return "整体可以按计划推进，稳中求进即可。"
	}
}

func filterEvidenceSignals(signals []EventSignal, polarity string, shensha bool) []EventSignal {
	var out []EventSignal
	for _, s := range signals {
		if s.Evidence == "" || s.Type == "用神基底" {
			continue
		}
		if s.Polarity != polarity {
			continue
		}
		if (s.Source == SourceShensha) != shensha {
			continue
		}
		out = append(out, s)
	}
	return out
}

func compactEvidence(ev string) string {
	ev = strings.TrimSpace(ev)
	ev = strings.ReplaceAll(ev, "（受空亡影响，力度减半）", "")
	ev = strings.ReplaceAll(ev, "（本年有重煞，此信号仅作参考）", "")
	return ev
}

// ExtractYearSignalTypes 提取 signal types（用于前端徽标显示），去除"用神基底"等内部 type
func ExtractYearSignalTypes(ys YearSignals) []string {
	hide := map[string]bool{"用神基底": true, "综合变动": false}
	out := []string{}
	seen := map[string]bool{}
	for _, s := range ys.Signals {
		if hide[s.Type] {
			continue
		}
		if seen[s.Type] {
			continue
		}
		seen[s.Type] = true
		out = append(out, s.Type)
	}
	return out
}
