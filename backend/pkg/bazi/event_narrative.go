package bazi

import "strings"

// RenderYearNarrative 根据 EventSignal 列表生成面向用户的白话批语。
// 底层 Evidence 保留给 RenderEvidenceSummary，不直接暴露在默认正文中。
func RenderYearNarrative(ys YearSignals) string {
	if len(ys.Signals) == 0 {
		return "本年命理信号较弱，运势相对平稳，无明显重大变动。"
	}

	primary, ok := pickDominantSignal(ys.Signals, "", ys.Age)
	if !ok {
		return ys.GanZhi + "年整体动象不强，适合按部就班推进，保持稳定节奏。"
	}
	secondary, hasSecondary := pickDominantSignal(ys.Signals, themeOf(primary.Type), ys.Age)

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

func pickDominantSignal(signals []EventSignal, excludeTheme string, age int) (EventSignal, bool) {
	var best EventSignal
	found := false
	bestRank := 999
	bestPol := 999
	isYoung := age > 0 && age < YoungAgeCutoff
	for _, s := range signals {
		if s.Type == "用神基底" {
			continue
		}
		theme := themeOf(s.Type)
		if theme == "" || theme == excludeTheme {
			continue
		}
		rank := themeRank(theme, s, isYoung)
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

func themeRank(theme string, sig EventSignal, isYoung bool) int {
	if isYoung {
		switch {
		case isSchoolType(sig.Type):
			return 1
		case theme == "relationship":
			return 2
		case theme == "health":
			return 3
		case theme == "money":
			return 4
		case theme == "change" && isStrongChangeSignal(sig):
			return 5
		case theme == "movement":
			return 6
		case theme == "change":
			return 7
		case theme == "support":
			return 8
		default:
			return 99
		}
	}
	switch theme {
	case "change":
		if isStrongChangeSignal(sig) {
			return 1
		}
		return 7
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
		return 8
	default:
		return 99
	}
}

func isSchoolType(typ string) bool {
	switch typ {
	case TypeXueYeYaLi, TypeXueYeGuiRen, TypeXueYeCaiYi, TypeXueYeJingZheng:
		return true
	default:
		return false
	}
}

func isStrongChangeSignal(sig EventSignal) bool {
	switch sig.Type {
	case "伏吟", "反吟", "大运合化", TypeJuShiZhong:
		return true
	}
	for _, marker := range []string{"大运流年双重命中", "力度倍增", "重大事件高发"} {
		if strings.Contains(sig.Evidence, marker) {
			return true
		}
	}
	return false
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
		return prefix + changeSentence(sig)
	case "health":
		if strings.Contains(sig.Evidence, "睡眠") || strings.Contains(sig.Evidence, "精神") || strings.Contains(sig.Evidence, "自刑") {
			return prefix + "作息和情绪状态需要多照看，压力大时更要保证休息。"
		}
		if strings.Contains(sig.Evidence, "冲") || strings.Contains(sig.Evidence, "意外") || strings.Contains(sig.Evidence, "白虎") {
			return prefix + "健康和出行安全需要多留心，避免冒险、磕碰和过度劳累。"
		}
		return prefix + "身体状态和日常作息需要更稳一些，别把小问题拖成大问题。"
	case "relationship":
		if sig.Type == TypeXingGePanNi || strings.Contains(sig.Evidence, "情绪") || strings.Contains(sig.Evidence, "家庭") {
			return prefix + "情绪和家庭沟通容易起波动，越是着急表达，越要放慢语气。"
		}
		if strings.HasPrefix(sig.Type, "性格_") || strings.Contains(sig.Evidence, "同窗") || strings.Contains(sig.Evidence, "人缘") {
			return prefix + "同学朋友之间更有存在感，既有靠近的机会，也要注意分寸。"
		}
		if sig.Polarity == PolarityXiong || sig.Type == "婚恋_冲" || sig.Type == "婚恋_变" {
			return prefix + "感情和人际关系容易被触动，适合把话说清楚，不宜急着做重决定。"
		}
		return prefix + "感情和人缘有被带动的迹象，主动沟通会更容易打开局面。"
	case "career":
		if strings.HasPrefix(sig.Type, "学业_") {
			return prefix + schoolSentence(sig)
		}
		if sig.Polarity == PolarityXiong {
			return prefix + "工作方向、职责或合作关系容易出现调整，需要重新适应节奏。"
		}
		return prefix + "工作上容易出现助力或机会，适合主动承担、争取表现。"
	case "money":
		if sig.Type == TypeXueYeZiYuan {
			return prefix + "家庭资源或学习投入会更受关注，适合把条件用在真正有帮助的地方。"
		}
		if sig.Polarity == PolarityXiong || sig.Type == "财运_损" {
			return prefix + "钱财机会看似增加，但也容易伴随支出和压力，宜稳健安排。"
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

func changeSentence(sig EventSignal) string {
	if sig.Type == "伏吟" {
		return "旧事或同类问题容易反复出现，适合把过去没处理完的事情重新梳理。"
	}
	if sig.Type == "反吟" {
		return "变化力度会比平时更明显，遇到突发调整时先稳住节奏。"
	}
	if sig.Type == "大运合化" {
		return "这段阶段性的方向感会被牵动，想法和选择容易出现转向。"
	}
	if sig.Type == TypeJuShiZhong {
		return "整体压力会被放大，重要决定宜慢一点，先避开高风险选择。"
	}
	if strings.Contains(sig.Evidence, "大运流年双重命中") || strings.Contains(sig.Evidence, "力度倍增") || strings.Contains(sig.Evidence, "重大事件高发") {
		return "这一年外部节奏推得更急，容易集中出现转折点，先稳住再行动。"
	}
	if sig.Source == SourceKongwang || strings.Contains(sig.Evidence, "空") || strings.Contains(sig.Evidence, "虚而不实") {
		return "计划感会比较强，但落地感未必稳定，重要安排要多确认细节。"
	}
	if strings.Contains(sig.Evidence, "月柱") || strings.Contains(sig.Evidence, "提纲") {
		return "学习方向、班级环境或老师要求容易调整，需要重新适应节奏。"
	}
	if strings.Contains(sig.Evidence, "日支") || strings.Contains(sig.Evidence, "自我宫位") {
		return "情绪、人际和家庭沟通更容易起波动，遇事先别急着顶上去。"
	}
	if strings.Contains(sig.Evidence, "驿马") || strings.Contains(sig.Evidence, "奔波") || strings.Contains(sig.Evidence, "迁移") {
		return "出行、搬动或环境变化会增加，提前安排会更稳。"
	}
	if strings.Contains(sig.Evidence, "大运地支") && strings.Contains(sig.Evidence, "流年地支") {
		return "外部节奏和环境要求更容易变化，适合顺势调整安排。"
	}
	if sig.Polarity == PolarityJi {
		return "事情容易被推动起来，适合顺势整理方向，把该推进的安排往前推一推。"
	}
	return "事情会有起伏和变动，先看清局面，再决定要不要推进。"
}

func schoolSentence(sig EventSignal) string {
	switch sig.Type {
	case TypeXueYeYaLi:
		if sig.Polarity == PolarityXiong {
			return "学习压力会更明显，考试、规则或师长要求容易让人紧绷。"
		}
		return "学习上会遇到规则和要求，适合把基础打牢，不要临时应付。"
	case TypeXueYeGuiRen:
		if sig.Polarity == PolarityJi {
			return "学习上容易得到师长帮助或方法启发，适合稳住节奏继续积累。"
		}
		return "学习方法和师长关系需要多磨合，有问题及时请教会更稳。"
	case TypeXueYeCaiYi:
		return "兴趣、表达和才艺会更有存在感，但要避免分心过度。"
	case TypeXueYeJingZheng:
		if sig.Polarity == PolarityJi {
			return "同伴和团队支持会更明显，适合借助集体力量完成目标。"
		}
		return "同学之间竞争感会增强，比较心重时更要回到自己的节奏。"
	default:
		return "学习和日常规则会更受关注，按节奏积累比急着突破更重要。"
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
