package bazi

import "strings"

type GongJiaItem struct {
	Source       string   `json:"source"`
	SourceLabels []string `json:"source_labels"`
	SameGan      string   `json:"same_gan"`
	SourceZhis   []string `json:"source_zhis"`
	VirtualZhi   string   `json:"virtual_zhi"`
	HideGan      []string `json:"hide_gan"`
	ShiShen      []string `json:"shishen"`
	ShenSha      []string `json:"shensha"`
	Meaning      string   `json:"meaning"`
}

type gongJiaPair struct {
	source string
	labels []string
	ganA   string
	zhiA   string
	ganB   string
	zhiB   string
}

func BuildGongJia(r *BaziResult) []GongJiaItem {
	if r == nil || r.DayGan == "" {
		return nil
	}

	pairs := []gongJiaPair{
		{source: "year_month", labels: []string{"年柱", "月柱"}, ganA: r.YearGan, zhiA: r.YearZhi, ganB: r.MonthGan, zhiB: r.MonthZhi},
		{source: "month_day", labels: []string{"月柱", "日柱"}, ganA: r.MonthGan, zhiA: r.MonthZhi, ganB: r.DayGan, zhiB: r.DayZhi},
		{source: "day_hour", labels: []string{"日柱", "时柱"}, ganA: r.DayGan, zhiA: r.DayZhi, ganB: r.HourGan, zhiB: r.HourZhi},
	}

	items := make([]GongJiaItem, 0, len(pairs))
	for _, pair := range pairs {
		if pair.ganA == "" || pair.ganA != pair.ganB {
			continue
		}
		virtualZhi, ok := clippedZhiBetween(pair.zhiA, pair.zhiB)
		if !ok {
			continue
		}
		hideGan := append([]string(nil), zhiHideGanFull[virtualZhi]...)
		shiShen := make([]string, 0, len(hideGan))
		for _, gan := range hideGan {
			shiShen = append(shiShen, GetShiShen(r.DayGan, gan))
		}
		shensha := GetGongJiaShenSha(r.YearGan, r.YearZhi, r.MonthZhi, r.DayGan, r.DayZhi, virtualZhi)
		items = append(items, GongJiaItem{
			Source:       pair.source,
			SourceLabels: append([]string(nil), pair.labels...),
			SameGan:      pair.ganA,
			SourceZhis:   []string{pair.zhiA, pair.zhiB},
			VirtualZhi:   virtualZhi,
			HideGan:      hideGan,
			ShiShen:      shiShen,
			ShenSha:      shensha,
			Meaning:      "原局相邻两柱同干，地支隔一位夹出虚支；仅作为中性信号层参考，不参与五行强弱与用忌计算。",
		})
	}

	return items
}

func GetGongJiaShenSha(yearGan, yearZhi, monthZhi, dayGan, dayZhi, virtualZhi string) []string {
	var result []string
	seen := map[string]bool{}
	add := func(cond bool, name string) {
		if cond && !seen[name] {
			result = append(result, name)
			seen[name] = true
		}
	}

	add(gongJiaTianYi(dayGan, virtualZhi) || gongJiaTianYi(yearGan, virtualZhi), "天乙贵人")
	add(gongJiaWenChang(dayGan, virtualZhi) || gongJiaWenChang(yearGan, virtualZhi), "文昌贵人")
	add(isTaohuaBase(yearZhi, virtualZhi) || isTaohuaBase(dayZhi, virtualZhi), "桃花")
	add(isYimaBase(yearZhi, virtualZhi) || isYimaBase(dayZhi, virtualZhi), "驿马")
	add(gongJiaHuagai(yearZhi, virtualZhi) || gongJiaHuagai(dayZhi, virtualZhi), "华盖")
	add(gongJiaJiangxing(yearZhi, virtualZhi) || gongJiaJiangxing(dayZhi, virtualZhi), "将星")
	add(gongJiaJiesha(yearZhi, virtualZhi) || gongJiaJiesha(dayZhi, virtualZhi), "劫煞")
	add(gongJiaZaisha(yearZhi, virtualZhi) || gongJiaZaisha(dayZhi, virtualZhi), "灾煞")

	return result
}

func gongJiaTianYi(stem, branch string) bool {
	return (strings.Contains("甲戊庚", stem) && strings.Contains("丑未", branch)) ||
		(strings.Contains("乙己", stem) && strings.Contains("子申", branch)) ||
		(strings.Contains("丙丁", stem) && strings.Contains("亥酉", branch)) ||
		(strings.Contains("壬癸", stem) && strings.Contains("卯巳", branch)) ||
		(stem == "辛" && strings.Contains("午寅", branch))
}

func gongJiaWenChang(stem, branch string) bool {
	return (stem == "甲" && branch == "巳") || (stem == "乙" && branch == "午") ||
		(strings.Contains("丙戊", stem) && branch == "申") || (strings.Contains("丁己", stem) && branch == "酉") ||
		(stem == "庚" && branch == "亥") || (stem == "辛" && branch == "子") ||
		(stem == "壬" && branch == "寅") || (stem == "癸" && branch == "卯")
}

func gongJiaHuagai(base, check string) bool {
	return (strings.Contains("申子辰", base) && check == "辰") ||
		(strings.Contains("寅午戌", base) && check == "戌") ||
		(strings.Contains("亥卯未", base) && check == "未") ||
		(strings.Contains("巳酉丑", base) && check == "丑")
}

func gongJiaJiangxing(base, check string) bool {
	return (strings.Contains("申子辰", base) && check == "子") ||
		(strings.Contains("寅午戌", base) && check == "午") ||
		(strings.Contains("亥卯未", base) && check == "卯") ||
		(strings.Contains("巳酉丑", base) && check == "酉")
}

func gongJiaJiesha(base, check string) bool {
	return (strings.Contains("申子辰", base) && check == "巳") ||
		(strings.Contains("寅午戌", base) && check == "亥") ||
		(strings.Contains("亥卯未", base) && check == "申") ||
		(strings.Contains("巳酉丑", base) && check == "寅")
}

func gongJiaZaisha(base, check string) bool {
	return (strings.Contains("申子辰", base) && check == "午") ||
		(strings.Contains("寅午戌", base) && check == "子") ||
		(strings.Contains("亥卯未", base) && check == "酉") ||
		(strings.Contains("巳酉丑", base) && check == "卯")
}

func clippedZhiBetween(a, b string) (string, bool) {
	aIndex, aOK := zhiIndex(a)
	bIndex, bOK := zhiIndex(b)
	if !aOK || !bOK {
		return "", false
	}
	if (aIndex+2)%len(Zhi) == bIndex {
		return Zhi[(aIndex+1)%len(Zhi)], true
	}
	if (bIndex+2)%len(Zhi) == aIndex {
		return Zhi[(bIndex+1)%len(Zhi)], true
	}
	return "", false
}

func EnsureGongJia(r *BaziResult) bool {
	if r == nil || len(r.GongJia) > 0 {
		return false
	}
	r.GongJia = BuildGongJia(r)
	return len(r.GongJia) > 0
}

func zhiIndex(zhi string) (int, bool) {
	for i, item := range Zhi {
		if item == zhi {
			return i, true
		}
	}
	return 0, false
}
