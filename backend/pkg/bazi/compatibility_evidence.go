package bazi

import "fmt"

// buildCompatibilityEvidencesV3 把 4 模块的命中按设计文档 §4.2 转为 evidence 列表。
// 仅产出"positive"性 evidence（纯加分制无 negative）；4 模块最多 6 条（1+1+1+3）。
func buildCompatibilityEvidencesV3(a, b *BaziResult) []CompatibilityEvidence {
	out := make([]CompatibilityEvidence, 0, 6)
	out = append(out, zodiacEvidence(a, b)...)
	out = append(out, nayinEvidence(a, b)...)
	out = append(out, dayPillarEvidence(a, b)...)
	out = append(out, eightCharsEvidence(a, b)...)
	return out
}

func zodiacEvidence(a, b *BaziResult) []CompatibilityEvidence {
	kind := branchCompatibilityKind(a.YearZhi, b.YearZhi)
	switch kind {
	case "liuhe":
		return []CompatibilityEvidence{{
			EvidenceKey: "zodiac_liuhe",
			Dimension:   "zodiac",
			Type:        "年支六合",
			Polarity:    "positive",
			Source:      "zodiac",
			Title:       "年支六合",
			Detail:      fmt.Sprintf("双方年支 %s/%s 构成六合，属相基础线吸引力强。", a.YearZhi, b.YearZhi),
			Weight:      50,
		}}
	case "sanhe":
		group := sanheGroupName(a.YearZhi, b.YearZhi)
		return []CompatibilityEvidence{{
			EvidenceKey: "zodiac_sanhe",
			Dimension:   "zodiac",
			Type:        "年支三合",
			Polarity:    "positive",
			Source:      "zodiac",
			Title:       "年支三合",
			Detail:      fmt.Sprintf("双方年支 %s/%s 同属 %s 三合局，气场协同。", a.YearZhi, b.YearZhi, group),
			Weight:      50,
		}}
	}
	return nil
}

func nayinEvidence(a, b *BaziResult) []CompatibilityEvidence {
	gzA := a.YearGan + a.YearZhi
	gzB := b.YearGan + b.YearZhi
	wxA := nayinElement(gzA)
	wxB := nayinElement(gzB)
	switch nayinRelation(wxA, wxB) {
	case "sheng":
		return []CompatibilityEvidence{{
			EvidenceKey: "nayin_sheng",
			Dimension:   "nayin",
			Type:        "纳音相生",
			Polarity:    "positive",
			Source:      "nayin",
			Title:       "纳音相生",
			Detail:      fmt.Sprintf("%s 与 %s 纳音五行相生，资源 / 情绪流动顺。", wxA, wxB),
			Weight:      20,
		}}
	case "same":
		return []CompatibilityEvidence{{
			EvidenceKey: "nayin_same",
			Dimension:   "nayin",
			Type:        "纳音相同",
			Polarity:    "positive",
			Source:      "nayin",
			Title:       "纳音同气",
			Detail:      fmt.Sprintf("双方纳音同为 %s，本质同气。", wxA),
			Weight:      20,
		}}
	}
	return nil
}

func dayPillarEvidence(a, b *BaziResult) []CompatibilityEvidence {
	if !branchCompatible(a.DayZhi, b.DayZhi) {
		return nil
	}
	if ganUpperTier(a.DayGan, b.DayGan) {
		return []CompatibilityEvidence{{
			EvidenceKey: "day_pillar_upper",
			Dimension:   "day_pillar",
			Type:        "日柱上档",
			Polarity:    "positive",
			Source:      "day_pillar",
			Title:       "日柱上档",
			Detail: fmt.Sprintf(
				"日柱 %s%s/%s%s 地支合且天干强化（五合 / 相生），亲密层结构稳。",
				a.DayGan, a.DayZhi, b.DayGan, b.DayZhi,
			),
			Weight: 10,
		}}
	}
	return []CompatibilityEvidence{{
		EvidenceKey: "day_pillar_lower",
		Dimension:   "day_pillar",
		Type:        "日柱下档",
		Polarity:    "positive",
		Source:      "day_pillar",
		Title:       "日柱次吉",
		Detail: fmt.Sprintf(
			"日柱 %s%s/%s%s 地支合，天干仅相同 / 克 / 无关，亲密层有基础但未达上吉。",
			a.DayGan, a.DayZhi, b.DayGan, b.DayZhi,
		),
		Weight: 5,
	}}
}

func eightCharsEvidence(a, b *BaziResult) []CompatibilityEvidence {
	out := make([]CompatibilityEvidence, 0, 3)
	type pillar struct {
		name  string
		label string
		ganA  string
		zhiA  string
		ganB  string
		zhiB  string
	}
	pillars := []pillar{
		{"year", "年柱", a.YearGan, a.YearZhi, b.YearGan, b.YearZhi},
		{"month", "月柱", a.MonthGan, a.MonthZhi, b.MonthGan, b.MonthZhi},
		{"hour", "时柱", a.HourGan, a.HourZhi, b.HourGan, b.HourZhi},
	}
	tierLabelMap := map[string]string{"upper": "上档", "lower": "下档"}
	for _, p := range pillars {
		s := scoreDayPillar(p.ganA, p.zhiA, p.ganB, p.zhiB)
		if s == 0 {
			continue
		}
		tier := "lower"
		if s == 10 {
			tier = "upper"
		}
		tierLabel := tierLabelMap[tier]
		out = append(out, CompatibilityEvidence{
			EvidenceKey: "eight_chars_" + p.name + "_" + tier,
			Dimension:   "eight_chars",
			Type:        p.label + "对" + tierLabel,
			Polarity:    "positive",
			Source:      "eight_chars",
			Title:       p.label + "对" + tierLabel,
			Detail: fmt.Sprintf(
				"%s %s%s/%s%s 命中%s（贡献 %d）。",
				p.label, p.ganA, p.zhiA, p.ganB, p.zhiB, tierLabel, s,
			),
			Weight: s,
		})
	}
	return out
}
