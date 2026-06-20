package bazi

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
		items = append(items, GongJiaItem{
			Source:       pair.source,
			SourceLabels: append([]string(nil), pair.labels...),
			SameGan:      pair.ganA,
			SourceZhis:   []string{pair.zhiA, pair.zhiB},
			VirtualZhi:   virtualZhi,
			HideGan:      hideGan,
			ShiShen:      shiShen,
			ShenSha:      []string{},
			Meaning:      "原局相邻两柱同干，地支隔一位夹出虚支；仅作为中性信号层参考，不参与五行强弱与用忌计算。",
		})
	}

	return items
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
	if r == nil || r.GongJia != nil {
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
