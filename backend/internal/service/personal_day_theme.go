package service

import (
	"fmt"
	"yuanju/pkg/bazi"
)

// Themes are traditional planning prompts. A match is not an event prediction.
func tianYiDay(stem, branch string) bool {
	switch stem {
	case "甲", "戊", "庚":
		return branch == "丑" || branch == "未"
	case "乙", "己":
		return branch == "子" || branch == "申"
	case "丙", "丁":
		return branch == "亥" || branch == "酉"
	case "壬", "癸":
		return branch == "卯" || branch == "巳"
	case "辛":
		return branch == "午" || branch == "寅"
	}
	return false
}

func taoHuaDay(base, branch string) bool {
	switch base {
	case "申", "子", "辰":
		return branch == "酉"
	case "寅", "午", "戌":
		return branch == "卯"
	case "亥", "卯", "未":
		return branch == "子"
	case "巳", "酉", "丑":
		return branch == "午"
	}
	return false
}

func dayBranchLiuHe(base, branch string) bool {
	pairs := map[string]string{"子": "丑", "丑": "子", "寅": "亥", "亥": "寅", "卯": "戌", "戌": "卯", "辰": "酉", "酉": "辰", "巳": "申", "申": "巳", "午": "未", "未": "午"}
	return pairs[base] == branch && branch != ""
}

func buildPersonalDayThemes(chart *bazi.BaziResult, ctx personalDayActivityContext) []PersonalDayGuidance {
	themes := []PersonalDayGuidance{}
	if !ctx.valid || chart == nil {
		return themes
	}

	if tianYiDay(chart.DayGan, ctx.dayZhi) || tianYiDay(chart.YearGan, ctx.dayZhi) {
		item := newGuidance("noble", "人际往来")
		item.Level, item.Label, item.SourceLabel = "candidate", "宜", "本人命盘"
		item.Sources = []string{"personal_chart"}
		item.Summary = "适合主动沟通、请教或推进协作。"
		item.Actions = []string{"把需要请教的问题提前列清楚，再联系合适的人。"}
		if tianYiDay(chart.DayGan, ctx.dayZhi) {
			item.Evidence = append(item.Evidence, fmt.Sprintf("本人日干%s与当日日支%s符合天乙贵人规则。", chart.DayGan, ctx.dayZhi))
		}
		if tianYiDay(chart.YearGan, ctx.dayZhi) {
			item.Evidence = append(item.Evidence, fmt.Sprintf("本人年干%s与当日日支%s符合天乙贵人规则。", chart.YearGan, ctx.dayZhi))
		}
		if ctx.zhiDirection == "adverse" || ctx.conflictElements[ctx.zhiElement] || ctx.clash {
			item.Level, item.Label = "review", "慎"
			item.Summary = "有贵人参考信号，也有个人关系冲突；协作安排宜多核对。"
			item.Actions = []string{"联系他人前先整理问题，重要决定仍核对实际条件。"}
			item.Evidence = append(item.Evidence, "当日日支另有忌神、调候分歧或与本人日支相冲的信号。")
		}
		themes = append(themes, item)
	} else {
		item := newGuidance("noble", "人际往来")
		item.Level, item.Label, item.SourceLabel = "neutral", "平", "本人命盘"
		item.Sources = []string{"personal_chart"}
		item.Summary = "今日未见特别的贵人信号，交流协作照常安排。"
		item.Actions = []string{"有需要可主动请教或联系他人，提前准备好要讨论的问题。"}
		item.Evidence = []string{fmt.Sprintf("本人年干%s、日干%s与当日日支%s未命中天乙贵人规则。", chart.YearGan, chart.DayGan, ctx.dayZhi)}
		themes = append(themes, item)
	}

	finance := domainGuidance(ctx, "finance", "财星参考")
	if finance.Level == "candidate" || finance.Level == "review" {
		if finance.Level == "candidate" {
			finance.Summary = "可安排投资研究与财务规划；不代表买卖时机。"
		}
	} else {
		finance.Level, finance.Label = "neutral", "平"
		finance.Summary = "今日未见明确的财务主题倾向，按计划做研究与预算核对。"
		if len(finance.Evidence) == 0 {
			finance.Evidence = []string{fmt.Sprintf("当日日干%s对应%s、日支%s主气对应%s；未触发财务主题的明确喜忌。", ctx.dayGan, ctx.ganGod, ctx.dayZhi, ctx.zhiGod)}
		}
	}
	finance.Actions = []string{"核对已有计划、预算和风险，不因单日参考改变投资决定。"}
	themes = append(themes, finance)

	if taoHuaDay(chart.YearZhi, ctx.dayZhi) || taoHuaDay(chart.DayZhi, ctx.dayZhi) || dayBranchLiuHe(chart.DayZhi, ctx.dayZhi) {
		item := newGuidance("romance", "桃花缘分")
		item.Level, item.Label, item.SourceLabel = "candidate", "宜", "本人命盘"
		item.Sources = []string{"personal_chart"}
		item.Summary = "可考虑安排见面或社交，保持自然交流。"
		item.Actions = []string{"联系想见的朋友或安排约会，提前确认双方时间与意愿。"}
		if taoHuaDay(chart.YearZhi, ctx.dayZhi) {
			item.Evidence = append(item.Evidence, fmt.Sprintf("本人年支%s对应桃花支%s。", chart.YearZhi, ctx.dayZhi))
		}
		if taoHuaDay(chart.DayZhi, ctx.dayZhi) {
			item.Evidence = append(item.Evidence, fmt.Sprintf("本人日支%s对应桃花支%s。", chart.DayZhi, ctx.dayZhi))
		}
		if dayBranchLiuHe(chart.DayZhi, ctx.dayZhi) {
			item.Evidence = append(item.Evidence, fmt.Sprintf("当日日支%s与本人日支%s六合，可作见面交流的传统参考。", ctx.dayZhi, chart.DayZhi))
		}
		if ctx.clash || ctx.zhiDirection == "adverse" || ctx.conflictElements[ctx.zhiElement] {
			item.Level, item.Label = "review", "慎"
			item.Summary = "有社交参考信号，也有个人关系冲突；见面安排宜留弹性。"
			item.Actions = []string{"提前沟通时间与地点，留出调整安排的余地。"}
			if ctx.clash {
				item.Evidence = append(item.Evidence, fmt.Sprintf("当日日支%s与本人日支%s相冲。", ctx.dayZhi, chart.DayZhi))
			}
			if ctx.zhiDirection == "adverse" {
				item.Evidence = append(item.Evidence, "当日日支主气落入本人扶抑忌神。")
			}
			if ctx.conflictElements[ctx.zhiElement] {
				item.Evidence = append(item.Evidence, "当日日支主气存在扶抑与调候方向分歧。")
			}
		}
		themes = append(themes, item)
	} else {
		item := newGuidance("romance", "桃花缘分")
		item.Level, item.Label, item.SourceLabel = "neutral", "平", "本人命盘"
		item.Sources = []string{"personal_chart"}
		item.Summary = "今日未见特别的桃花或相合信号，社交见面按实际安排。"
		item.Actions = []string{"可按彼此时间联系朋友或安排约会，保持自然沟通。"}
		item.Evidence = []string{fmt.Sprintf("当日日支%s未命中本人年支%s、日支%s的桃花规则，也未与本人日支六合。", ctx.dayZhi, chart.YearZhi, chart.DayZhi)}
		themes = append(themes, item)
	}

	if ctx.ganDirection == "adverse" && ctx.zhiDirection == "adverse" {
		item := newGuidance("wellbeing", "作息节奏")
		item.Level, item.Label, item.SourceLabel = "review", "慎", "本人命盘"
		item.Sources = []string{"personal_chart"}
		item.Summary = "今日干支均落入忌神，作息和活动安排宜从容一些。"
		item.Actions = []string{"为连续工作留出休息时间，避免把活动安排得过满。", "按自己的体感调整活动强度，保持平常作息。"}
		item.Evidence = append(item.Evidence, "当日日干和日支主气均落入本人扶抑忌神。")
		themes = append(themes, item)
	} else {
		item := newGuidance("wellbeing", "作息节奏")
		item.Level, item.Label, item.SourceLabel = "neutral", "平", "本人命盘"
		item.Sources = []string{"personal_chart"}
		item.Summary = "未触发双忌神身心提醒，按平常节奏安排作息。"
		item.Actions = []string{"连续工作或活动时，按需要留出休息时间。"}
		item.Evidence = []string{"当日日干和日支主气未同时落入本人扶抑忌神。"}
		themes = append(themes, item)
	}
	return themes
}
