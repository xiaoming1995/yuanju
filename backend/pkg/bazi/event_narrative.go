package bazi

import "strings"

// MinSentencesForNarrative is the minimum number of non-empty
// evidence-anchored sentences required before RenderYearNarrative
// returns a paragraph. Below this threshold the narrative is hidden
// (returns ""), and the frontend renders only the signal chips and
// evidence summary for that year.
//
// Threshold = 2 (not 3): after dropping polarity-only fallbacks and
// the age<18 generic closer, a typical "thin but real" year produces
// 2 evidence-anchored sentences (domain + secondary, or yearTone +
// domain). Requiring 3 would over-hide. Truly empty years still
// surface 0 sentences and stay hidden.
const MinSentencesForNarrative = 2

// structuralEvidenceKeywords are the substrings inside EventSignal.Evidence
// that the narrative engine treats as concrete, citable anchors. Used by
// hasEvidenceAnchor to gate secondaryDetailSentence.
var structuralEvidenceKeywords = []string{
	"冲", "刑", "空", "用神", "忌神", "驿马", "月柱", "提纲", "日支", "自我宫位", "大运流年双重命中", "意外", "白虎",
}

// hasEvidenceAnchor returns true when sig carries a specific differentiator
// that a sentence can cite: a hard-event Source, an allowed Type, or a
// recognized keyword inside Evidence. Used to gate secondary-detail prose
// so it does not pad narratives with generic theme wording.
func hasEvidenceAnchor(sig EventSignal) bool {
	if isHardEventSignal(sig) {
		return true
	}
	switch sig.Type {
	case "伏吟", "反吟", "大运合化", TypeJuShiZhong:
		return true
	}
	if strings.HasPrefix(sig.Type, "学业_") || strings.HasPrefix(sig.Type, "性格_") || strings.HasPrefix(sig.Type, "婚恋_") {
		return true
	}
	for _, k := range structuralEvidenceKeywords {
		if strings.Contains(sig.Evidence, k) {
			return true
		}
	}
	return false
}

// RenderYearNarrative 根据 EventSignal 列表生成面向用户的白话批语。
// 底层 Evidence 保留给 RenderEvidenceSummary，不直接暴露在默认正文中。
//
// 当能产出的"有 evidence 支撑"的句子数少于 MinSentencesForNarrative 时
// 返回空串，前端会跳过 narrative 段落，只渲染徽标和命理依据。
func RenderYearNarrative(ys YearSignals) string {
	primary, ok := pickDominantSignal(ys.Signals, "", ys.Age)
	if !ok {
		return ""
	}
	secondary, hasSecondary := pickDominantSignal(ys.Signals, themeOf(primary.Type), ys.Age)

	sentences := make([]string, 0, 6)
	if s := yearToneSentence(ys.Signals, primary); s != "" {
		sentences = append(sentences, s)
	}
	if s := triggerSourceSentence(primary, ys.Age); s != "" {
		sentences = append(sentences, s)
	}
	if s := domainDetailSentence(primary, secondary, hasSecondary, ys.Age); s != "" {
		sentences = append(sentences, s)
	}
	if hasSecondary {
		if s := secondaryDetailSentence(secondary, ys.Age); s != "" {
			sentences = append(sentences, s)
		}
	}
	if s := tenGodNarrativeSentence(ys.TenGodPower, primary, secondary, hasSecondary); s != "" {
		sentences = append(sentences, s)
	}
	if s := practicalStanceSentence(ys.Signals, primary, ys.Age); s != "" {
		sentences = append(sentences, s)
	}

	if len(sentences) < MinSentencesForNarrative {
		return ""
	}
	return joinNarrativeParts(append([]string{ys.GanZhi + "年，" + sentences[0]}, sentences[1:]...))
}

// RenderEvidenceSummary 提取专业用户可展开查看的命理依据。
func RenderEvidenceSummary(ys YearSignals) []string {
	if len(ys.Signals) == 0 && ys.TenGodPower.Reason == "" {
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
	if ys.TenGodPower.Reason != "" && len(out) < 5 {
		out = append(out, "十神力量："+ys.TenGodPower.Reason)
	}
	return out
}

func tenGodContextSentence(power TenGodPowerProfile) string {
	if power.PlainTitle == "" || power.PlainText == "" {
		return ""
	}
	return power.PlainTitle + "，" + strings.TrimSuffix(power.PlainText, "。") + "。"
}

func meaningfulSignals(signals []EventSignal) []EventSignal {
	out := make([]EventSignal, 0, len(signals))
	for _, s := range signals {
		if s.Type == "用神基底" || s.Type == TypeDayunPhase {
			continue
		}
		if themeOf(s.Type) == "" {
			continue
		}
		out = append(out, s)
	}
	return out
}

func joinNarrativeParts(parts []string) string {
	var out []string
	seen := map[string]bool{}
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if !strings.HasSuffix(p, "。") {
			p += "。"
		}
		if seen[p] {
			continue
		}
		seen[p] = true
		out = append(out, p)
	}
	return strings.Join(out, "")
}

func yearToneSentence(signals []EventSignal, primary EventSignal) string {
	if !isHardEventSignal(primary) {
		return ""
	}
	switch themeOf(primary.Type) {
	case "health":
		return healthLead(primary)
	case "change":
		return changeLead(primary)
	case "relationship":
		return relationshipLead(primary)
	default:
		return defaultHardLead(primary)
	}
}

func triggerSourceSentence(sig EventSignal, age int) string {
	ev := sig.Evidence
	switch {
	case strings.Contains(ev, "冲"):
		if age > 0 && age < YoungAgeCutoff {
			return "触发点多落在日常环境、同学关系或家庭沟通上，容易因为节奏变化而需要重新适应。"
		}
		return "触发点多来自关系、环境或职责边界的碰撞，容易出现调整、分歧或临时改变。"
	case strings.Contains(ev, "刑"):
		return "触发点偏向内耗和反复，事情未必一下爆发，却容易在细节中消耗心力。"
	case sig.Source == SourceKongwang || strings.Contains(ev, "空") || strings.Contains(ev, "虚而不实"):
		return "触发点带有不稳定感，计划看似成形，但落地前要多确认时间、承诺和细节。"
	case sig.Type == "伏吟":
		return "触发点来自同类主题反复，过去没有处理干净的事容易再次出现。"
	case sig.Type == "反吟":
		return "触发点来自明显反复和冲突，环境变化或突发安排会更容易打乱原计划。"
	case sig.Type == TypeJuShiZhong || strings.Contains(ev, "局") || strings.Contains(ev, "势力"):
		return "触发点来自整体局势被放大，单一选择可能牵动多个层面的压力。"
	case strings.Contains(ev, "用神") || strings.Contains(ev, "喜神"):
		return "触发点带有支持性质，贵人、资源或更合适的方向会在关键处出现。"
	case strings.Contains(ev, "忌神"):
		return "触发点带有消耗性质，越是看似诱人的机会，越要先判断代价。"
	case strings.Contains(ev, "驿马") || strings.Contains(ev, "奔波") || strings.Contains(ev, "迁移"):
		return "触发点落在移动和环境变化上，出行、搬动、换环境或事务奔波会增加。"
	default:
		return ""
	}
}

func domainDetailSentence(primary EventSignal, secondary EventSignal, hasSecondary bool, age int) string {
	theme := themeOf(primary.Type)
	switch theme {
	case "career":
		if age > 0 && age < YoungAgeCutoff || strings.HasPrefix(primary.Type, "学业_") {
			return richStudySentence(primary, secondary, hasSecondary)
		}
		if primary.Polarity == PolarityXiong {
			return "现实表现上，工作方向、责任分配或合作关系容易出现调整，既要适应新要求，也要避免把压力全部扛在自己身上。"
		}
		return "现实表现上，工作、职责或个人表现会更受关注，适合主动承担能看见成果的事项，但节奏仍要稳。"
	case "money":
		if age > 0 && age < YoungAgeCutoff || primary.Type == TypeXueYeZiYuan {
			return "现实表现上，家庭资源、学习投入或生活条件更受关注，适合把支持用在真正能提升自己的地方。"
		}
		if primary.Polarity == PolarityXiong || primary.Type == "财运_损" {
			return "现实表现上，钱财机会和支出压力会同时出现，容易有投入、置办、合作分账或预算失衡的问题。"
		}
		return "现实表现上，资源、收入、合作回报或实际利益更容易被看见，适合稳健争取，不宜贪快。"
	case "relationship":
		if age > 0 && age < YoungAgeCutoff {
			return "现实表现上，同学、朋友和家庭互动会更有存在感，既有靠近和被关注的机会，也容易因情绪或分寸引起波动。"
		}
		if primary.Polarity == PolarityXiong || primary.Type == "婚恋_冲" || primary.Type == "婚恋_变" {
			return "现实表现上，感情、人际或家庭沟通容易被触动，适合先把话说清楚，再决定关系和合作边界。"
		}
		return "现实表现上，人缘、感情互动和合作氛围会被带动，主动沟通更容易打开局面。"
	case "health":
		if strings.Contains(primary.Evidence, "冲") || strings.Contains(primary.Evidence, "意外") || strings.Contains(primary.Evidence, "白虎") {
			return "现实表现上，健康和出行安全需要放在前面，运动、交通、熬夜和过度劳累都要留出缓冲。"
		}
		return "现实表现上，身体状态、作息和情绪承压更明显，小问题宜早处理，不要拖到影响日常节奏。"
	case "movement":
		return "现实表现上，出行、搬动、岗位、学校或生活环境有变化机会，提前规划路线和时间会更从容。"
	case "support":
		return "现实表现上，外部助力、长辈提携或贵人资源更容易出现，适合借势推进重要安排。"
	case "change":
		return richChangeSentence(primary)
	default:
		return ""
	}
}

func secondaryDetailSentence(sig EventSignal, age int) string {
	if !hasEvidenceAnchor(sig) {
		return ""
	}
	theme := themeOf(sig.Type)
	switch theme {
	case "career":
		if age > 0 && age < YoungAgeCutoff || strings.HasPrefix(sig.Type, "学业_") {
			return "同时，学习方法、师长要求或同伴比较也会被带动，稳住基础比临时冲刺更重要。"
		}
		return "同时，工作表现和责任分配也会被带动，越是任务变多，越要分清主次。"
	case "money":
		if age > 0 && age < YoungAgeCutoff || sig.Type == TypeXueYeZiYuan {
			return "同时，家庭资源或学习投入会成为辅助主题，适合把条件用在真正有效的地方。"
		}
		return "同时，钱财和资源安排也要更清楚，尤其要避免边推进边增加无谓支出。"
	case "relationship":
		if age > 0 && age < YoungAgeCutoff {
			return "同时，同学朋友之间的距离会变得敏感，既要表达自己，也要注意语气和分寸。"
		}
		return "同时，人际和感情沟通会影响判断，急着表态前最好先确认彼此期待。"
	case "health":
		return "同时，作息、身体反应和安全边界不能忽视，压力大时更要保证休息。"
	case "movement":
		return "同时，出行、搬动或环境变化会增加，提前安排会减少临时被动。"
	case "support":
		return "同时，外部助力可以借用，但关键决定仍要回到自己的节奏。"
	case "change":
		s := richChangeSentence(sig)
		if s == "" {
			return ""
		}
		return "同时，" + s
	default:
		return ""
	}
}

func richStudySentence(primary EventSignal, secondary EventSignal, hasSecondary bool) string {
	switch primary.Type {
	case TypeXueYeYaLi:
		return "现实表现上，学习规则、考试要求或师长期待会更明显，适合把基础打牢，别靠临时应付硬撑。"
	case TypeXueYeGuiRen:
		if hasSecondary && themeOf(secondary.Type) == "relationship" {
			return "现实表现上，学习上容易得到师长帮助或方法启发，但同学关系和家庭情绪也会被带动，不能只顾成绩而忽略沟通。"
		}
		return "现实表现上，学习方法、证书考试或师长帮助更容易出现，适合稳住节奏，把能积累的内容扎实做起来。"
	case TypeXueYeCaiYi:
		return "现实表现上，兴趣、表达和才艺表现会更突出，适合展示能力，但要防止分心过多影响主线学习。"
	case TypeXueYeJingZheng:
		return "现实表现上，同学比较、团队协作和竞争感会增强，适合借助集体力量，但不要被比较心带乱节奏。"
	default:
		return ""
	}
}

func richChangeSentence(sig EventSignal) string {
	switch {
	case sig.Type == "伏吟":
		return "现实表现上，旧事、旧关系或类似问题容易再度出现，适合趁机梳理，而不是继续拖延。"
	case sig.Type == "反吟":
		return "现实表现上，变化会比平时更突然，环境、计划或关系节奏可能需要快速调整。"
	case sig.Source == SourceKongwang || strings.Contains(sig.Evidence, "空") || strings.Contains(sig.Evidence, "虚而不实"):
		return "现实表现上，想法和计划会比较多，但真正落地未必稳定，合同、承诺和时间安排要多确认。"
	case strings.Contains(sig.Evidence, "月柱") || strings.Contains(sig.Evidence, "提纲"):
		return "现实表现上，学习方向、工作环境、班级团队或上级要求容易调整，需要重新适应节奏。"
	case strings.Contains(sig.Evidence, "日支") || strings.Contains(sig.Evidence, "自我宫位"):
		return "现实表现上，情绪、人际和亲密关系更容易被触动，遇事先别急着顶上去。"
	case strings.Contains(sig.Evidence, "驿马") || strings.Contains(sig.Evidence, "奔波") || strings.Contains(sig.Evidence, "迁移"):
		return "现实表现上，出行、搬动、换环境或奔波事务会增加，提前安排会更稳。"
	default:
		return ""
	}
}

func tenGodNarrativeSentence(power TenGodPowerProfile, primary EventSignal, secondary EventSignal, hasSecondary bool) string {
	if power.PlainTitle == "" || power.PlainText == "" {
		return ""
	}
	groupTheme := tenGodGroupTheme(power.Group)
	if groupTheme == "" {
		return ""
	}
	if isHardEventSignal(primary) && groupTheme == themeOf(primary.Type) {
		return ""
	}
	if groupTheme == themeOf(primary.Type) || (hasSecondary && groupTheme == themeOf(secondary.Type)) {
		return "这股年度力量会把" + tenGodPlainDomain(power.Group, power.Polarity) + "推到台前，处理得好可以成为助力，处理得急则容易变成压力。"
	}
	return ""
}

func tenGodGroupTheme(group string) string {
	switch group {
	case TenGodGroupWealth:
		return "money"
	case TenGodGroupOfficial:
		return "career"
	case TenGodGroupSeal:
		return "support"
	case TenGodGroupOutput:
		return "change"
	case TenGodGroupPeer:
		return "relationship"
	default:
		return ""
	}
}

func tenGodPlainDomain(group, polarity string) string {
	switch group {
	case TenGodGroupWealth:
		return "钱财、资源和现实承诺"
	case TenGodGroupOfficial:
		return "规则、责任和外部要求"
	case TenGodGroupSeal:
		return "学习、师长和保护性资源"
	case TenGodGroupOutput:
		return "表达、才华和个人输出"
	case TenGodGroupPeer:
		if polarity == TenGodPolarityPressure {
			return "同辈竞争、合作分利和比较心"
		}
		return "朋友、团队和同伴助力"
	default:
		return "年度主导力量"
	}
}

func isHardEventSignal(sig EventSignal) bool {
	if sig.Source == SourceKongwang || sig.Source == SourceXing || sig.Source == SourceFuyin || sig.Source == SourceHehua {
		return true
	}
	return strings.Contains(sig.Evidence, "用神位") ||
		strings.Contains(sig.Evidence, "忌神位") ||
		strings.Contains(sig.Evidence, "大运流年双重命中") ||
		strings.Contains(sig.Evidence, "受冲") ||
		strings.Contains(sig.Evidence, "受刑") ||
		strings.Contains(sig.Evidence, "力度倍增")
}

func pickDominantSignal(signals []EventSignal, excludeTheme string, age int) (EventSignal, bool) {
	var best EventSignal
	found := false
	bestRank := 999
	bestPol := 999
	isYoung := age > 0 && age < YoungAgeCutoff
	for _, s := range signals {
		if s.Type == "用神基底" || s.Type == TypeDayunPhase {
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
	case TypeDayunPhase:
		return "phase"
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
		case theme == "phase":
			return 9
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
	case "phase":
		return 9
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
	case "phase":
		return prefix + dayunPhaseSentence(sig)
	default:
		return prefix + "整体运势有起伏，宜多观察、少冒进。"
	}
}

func dayunPhaseSentence(sig EventSignal) string {
	period := "这步大运"
	focus := "阶段节奏"
	if strings.Contains(sig.Evidence, "天干主事") || strings.Contains(sig.Evidence, "前5年") || strings.Contains(sig.Evidence, "前五年") {
		period = "本步大运前五年"
		focus = "天干主事"
	} else if strings.Contains(sig.Evidence, "地支主事") || strings.Contains(sig.Evidence, "后5年") || strings.Contains(sig.Evidence, "后五年") {
		period = "本步大运后五年"
		focus = "地支主事"
	}
	switch sig.Polarity {
	case PolarityJi:
		return period + "进入" + focus + "，整体底色偏顺，适合稳稳承接机会，不必急于冒进。"
	case PolarityXiong:
		return period + "进入" + focus + "，整体底色偏紧，安排上宜保守一些，先控制压力和风险。"
	default:
		return period + "进入" + focus + "，整体底色中平，关键仍在于把日常节奏守稳。"
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
		return "这一年宜保守谨慎，取舍要清楚，少做冲动决定，把风险降下来更重要。"
	case ji >= 2:
		return "这一年可以主动一些，把握机会，但仍要保持节制。"
	case xiong > 0:
		return "遇事不必急着硬碰硬，留出余地会更稳。"
	default:
		return "整体可以按计划推进，稳中求进即可。"
	}
}

func practicalStanceSentence(signals []EventSignal, primary EventSignal, age int) string {
	// Child-age years deliberately do not emit a "practical stance" closer.
	// The previous age<18 templates ("这一年最要紧的是稳住学习节奏..." etc.)
	// were theme-anchored only and guaranteed to repeat across years for any
	// recurring theme, padding child-age narratives with identical closers.
	// Dropping the closer means age<18 years lose one sentence; combined with
	// the MinSentencesForNarrative=3 threshold this naturally hides narratives
	// for years whose remaining evidence-anchored content is thin.
	if age > 0 && age < YoungAgeCutoff {
		return ""
	}
	xiong, ji := 0, 0
	for _, s := range signals {
		switch s.Polarity {
		case PolarityXiong:
			xiong++
		case PolarityJi:
			ji++
		}
	}
	theme := themeOf(primary.Type)
	switch theme {
	case "relationship":
		if xiong > 0 {
			return "这一年处理关系要先立边界，再谈推进；话说清楚，比急着表态更重要。"
		}
		return "这一年可以主动经营关系，但仍要把节奏掌握在自己手里。"
	case "health":
		return "这一年不宜硬扛，休息、安全和情绪管理要排在求快求成之前。"
	case "money":
		if xiong > 0 {
			return "这一年钱财上宜先控预算，再看机会；能延后确认的支出不要急着拍板。"
		}
		return "这一年资源机会可以争取，但要先算清投入和回报。"
	case "career":
		if xiong > 0 {
			return "这一年工作上宜先稳责任边界和取舍，再争表现；少接模糊任务，避免压力外溢。"
		}
		return "这一年适合主动承担可见成果的事务，但仍要分清主次。"
	case "movement":
		return "这一年凡涉及出行、搬动或环境调整，都宜提前确认细节，给自己留缓冲。"
	case "change":
		return "这一年先确认事实，再调整方向；越是变化多，越要避免凭一时情绪下结论。"
	case "support":
		return "这一年可以借助外力，但关键选择仍要自己拿稳，不宜完全依赖他人安排。"
	}
	return ""
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
	hide := map[string]bool{"用神基底": true, TypeDayunPhase: true, "综合变动": false}
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
