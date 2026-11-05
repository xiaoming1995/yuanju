package service

import (
	"fmt"
	"strings"
	"yuanju/pkg/bazi"
)

// addQuestionRelationEvidence compares an established layer with the user's day
// pillar. These are cultural timing references, not predictions of an event.
func addQuestionRelationEvidence(add func(string, string, string, string), layer, ganZhi string, chart *bazi.BaziResult, category string) {
	parts := []rune(ganZhi)
	if len(parts) != 2 || chart == nil {
		return
	}
	gan, zhi := string(parts[0]), string(parts[1])
	if dayStemElement[gan] == "" || dayBranchElement[zhi] == "" || dayStemElement[chart.DayGan] == "" || dayBranchElement[chart.DayZhi] == "" {
		return
	}
	const source = "本人命盘·冲合参考"
	matched := false
	combination := func(kind, incoming, natal, element string) {
		matched = true
		direction := questionRelationDirection(chart, element)
		label := "喜忌方向未定"
		if direction == "favorable" {
			label, direction = "本人喜用", "candidate"
		} else if direction == "adverse" {
			label, direction = "本人忌神", "review"
		}
		boundary := "不代表所问事项的实际结果"
		if category == "finance" {
			boundary = "不代表投资收益或适合买卖"
		}
		add(layer, source, direction, fmt.Sprintf("%s%s与本人日柱%s构成%s；%s属%s，%s。仅作整体安排参考，合不等于吉，%s。", layer, incoming, natal, kind, incoming, element, label, boundary))
	}
	if bazi.GanFiveHe(gan, chart.DayGan) {
		combination("天干五合", gan, chart.DayGan, dayStemElement[gan])
	}
	if bazi.ZhiSixHe(zhi, chart.DayZhi) {
		combination("地支六合", zhi, chart.DayZhi, dayBranchElement[zhi])
	}
	if bazi.ZhiSixChong(zhi, chart.DayZhi) {
		matched = true
		tip := "重要安排宜核对现实条件并留有余地"
		if category == "finance" {
			tip = "投资研究宜复核资金安排、流动性与风险承受力，忌仓促决定"
		}
		add(layer, source, "review", fmt.Sprintf("%s支%s与本人日支%s构成地支六冲；%s。相冲仅作整体安排参考，不代表所问事项失败或实际亏损。", layer, zhi, chart.DayZhi, tip))
	}
	if layer == "流日" && !matched {
		add(layer, source, "neutral", fmt.Sprintf("流日%s与本人日柱%s%s未形成天干五合、地支六合或六冲；无冲合不能据此判断所问事项吉凶。", ganZhi, chart.DayGan, chart.DayZhi))
	}
}

func questionRelationDirection(chart *bazi.BaziResult, element string) string {
	if chart.NatalAssessment == nil {
		return "neutral"
	}
	fuyi := chart.NatalAssessment.Fuyi
	if !validElementSet(fuyi.Yongshen) || !validElementSet(fuyi.Jishen) {
		return "neutral"
	}
	for _, e := range fuyi.Yongshen {
		if strings.ContainsRune(fuyi.Jishen, e) {
			return "neutral"
		}
	}
	return personalDayDirection(element, fuyi.Yongshen, fuyi.Jishen)
}

func todayQuestionIsIndirect(evidence []TodayQuestionEvidence) bool {
	hasReference, hasDirect := false, false
	for _, e := range evidence {
		if e.Direction != "candidate" && e.Direction != "favorable" && e.Direction != "review" && e.Direction != "adverse" {
			continue
		}
		if e.Source == "本人命盘" {
			hasDirect = true
		}
		if e.Source == "本人命盘·整体参考" || e.Source == "本人命盘·冲合参考" {
			hasReference = true
		}
	}
	return hasReference && !hasDirect
}
