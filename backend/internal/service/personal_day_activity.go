package service

import (
	"container/list"
	"fmt"
)

// PersonalDayGuidance is a limited planning reference, not an event prediction.
type PersonalDayGuidance struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Level       string   `json:"level"`
	Label       string   `json:"label"`
	Summary     string   `json:"summary"`
	Sources     []string `json:"sources"`
	SourceLabel string   `json:"source_label"`
	Evidence    []string `json:"evidence"`
}

type personalDayActivityContext struct {
	dayGan, dayZhi             string
	ganElement, zhiElement     string
	ganDirection, zhiDirection string
	ganGod, zhiGod             string
	yi, ji                     *list.List
	valid, clash               bool
	conflictElements           map[string]bool
}

func dayListContains(items *list.List, target string) bool {
	if items == nil {
		return false
	}
	for item := items.Front(); item != nil; item = item.Next() {
		if label, ok := item.Value.(string); ok && label == target {
			return true
		}
	}
	return false
}

func dayListLabels(items *list.List) []string {
	labels := []string{}
	if items == nil {
		return labels
	}
	for item := items.Front(); item != nil; item = item.Next() {
		if label, ok := item.Value.(string); ok && label != "" {
			labels = append(labels, label)
		}
	}
	return labels
}

func newGuidance(id, title string) PersonalDayGuidance {
	return PersonalDayGuidance{ID: id, Title: title, Level: "insufficient", Label: "未列", Sources: []string{}, Evidence: []string{}}
}

func travelGuidance(ctx personalDayActivityContext) PersonalDayGuidance {
	result := newGuidance("travel", "旅行与重要出行")
	result.Sources = []string{"calendar_yiji"}
	result.SourceLabel = "通用日历"
	yi := dayListContains(ctx.yi, "出行")
	ji := dayListContains(ctx.ji, "出行")
	generalJi := dayListContains(ctx.ji, "诸事不宜")
	if yi {
		result.Evidence = append(result.Evidence, "通用日历列出「宜出行」。")
	}
	if ji {
		result.Evidence = append(result.Evidence, "通用日历列出「忌出行」。")
	}
	if generalJi {
		result.Evidence = append(result.Evidence, "通用日历列出「诸事不宜」。")
	}
	switch {
	case ji || generalJi:
		result.Level, result.Label = "review", "忌"
		result.Summary = "通用日历忌出行，重要行程宜另行核对。"
	case yi:
		result.Level, result.Label = "candidate", "宜"
		result.Summary = "仅为通用日历宜出行，重要行程仍需核对。"
	default:
		result.Summary = "通用日历未列出行宜忌。"
		result.Evidence = append(result.Evidence, "通用日历未列出出行相关宜忌标签。")
	}
	if ctx.valid && ctx.clash {
		result.Sources = append(result.Sources, "personal_chart")
		result.SourceLabel = "通用日历 · 本人命盘"
		result.Level, result.Label = "review", "慎"
		result.Evidence = append(result.Evidence, fmt.Sprintf("当日日支「%s」与本人命盘日支相冲。", ctx.dayZhi))
		result.Summary = "本人日支相冲，重要行程宜核对。"
		if yi {
			result.Summary = "通用日历虽宜出行，本人日支相冲；重要行程宜核对。"
		}
	}
	return result
}

func guidanceDomain(god, id string) bool {
	switch id {
	case "finance":
		return god == "正财" || god == "偏财" || god == "比肩" || god == "劫财"
	case "work":
		return god == "正官" || god == "七杀"
	case "communication":
		return god == "食神" || god == "伤官"
	}
	return false
}

func domainGuidance(ctx personalDayActivityContext, id, title string) PersonalDayGuidance {
	result := newGuidance(id, title)
	result.Sources = []string{"personal_chart"}
	result.SourceLabel = "本人命盘"
	if !ctx.valid {
		result.Summary = "命盘资料不足，暂未列此项宜忌。"
		return result
	}
	type signal struct{ symbol, element, god, direction string }
	signals := []signal{}
	if guidanceDomain(ctx.ganGod, id) {
		signals = append(signals, signal{ctx.dayGan, ctx.ganElement, ctx.ganGod, ctx.ganDirection})
	}
	if guidanceDomain(ctx.zhiGod, id) {
		signals = append(signals, signal{ctx.dayZhi, ctx.zhiElement, ctx.zhiGod, ctx.zhiDirection})
	}
	if len(signals) == 0 {
		result.Summary = "今日未列此项宜忌。"
		return result
	}
	allFavorable := true
	relevantConflict := false
	directionalCount := 0
	for _, item := range signals {
		result.Evidence = append(result.Evidence, fmt.Sprintf("「%s」属%s，对日主为%s；扶抑方向：%s。", item.symbol, item.element, item.god, directionLabel(item.direction)))
		if item.direction != "favorable" {
			allFavorable = false
		}
		if item.direction != "neutral" {
			directionalCount++
		}
		if ctx.conflictElements[item.element] {
			relevantConflict = true
		}
	}
	if relevantConflict {
		result.Evidence = append(result.Evidence, "扶抑与调候对此五行方向不同。")
	}
	if directionalCount == 0 && !relevantConflict {
		result.Level, result.Label = "neutral", "平"
		result.Summary = "此项未见明显喜忌倾向。"
		return result
	}
	if relevantConflict || !allFavorable {
		result.Level, result.Label = "review", "慎"
		switch id {
		case "finance":
			result.Summary = "财务安排宜先核对预算、合同与风险。"
		case "work":
			result.Summary = "工作安排宜核对责任与时间。"
		default:
			result.Summary = "沟通交付宜检查内容与时间。"
		}
		return result
	}
	result.Level, result.Label = "candidate", "宜"
	switch id {
	case "finance":
		result.Summary = "宜做投资研究、预算与风险核对；不代表买卖时机。"
	case "work":
		result.Summary = "宜安排工作事务。"
	default:
		result.Summary = "宜安排沟通与交付。"
	}
	return result
}

func buildPersonalDayGuidance(ctx personalDayActivityContext) []PersonalDayGuidance {
	return []PersonalDayGuidance{
		travelGuidance(ctx),
		domainGuidance(ctx, "finance", "投资研究与财务规划"),
		domainGuidance(ctx, "work", "工作事务"),
		domainGuidance(ctx, "communication", "沟通交付"),
	}
}
