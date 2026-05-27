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
	if branchSameElement(a.YearZhi, b.YearZhi) {
		wx := wxPinyin2CN[zhiWuxing[a.YearZhi]]
		return []CompatibilityEvidence{{
			EvidenceKey: "zodiac_same_element",
			Dimension:   "zodiac",
			Type:        "年支同行",
			Polarity:    "positive",
			Source:      "zodiac",
			Title:       "年支同行",
			Detail:      fmt.Sprintf("双方年支 %s/%s 同属 %s 行（双生），属相层有亲近感。", a.YearZhi, b.YearZhi, wx),
			Weight:      30,
		}}
	}
	if branchShengElement(a.YearZhi, b.YearZhi) {
		wxA := wxPinyin2CN[zhiWuxing[a.YearZhi]]
		wxB := wxPinyin2CN[zhiWuxing[b.YearZhi]]
		return []CompatibilityEvidence{{
			EvidenceKey: "zodiac_sheng",
			Dimension:   "zodiac",
			Type:        "年支相生",
			Polarity:    "positive",
			Source:      "zodiac",
			Title:       "年支相生",
			Detail:      fmt.Sprintf("双方年支 %s/%s 构成 %s/%s 五行相生，属相层有顺承之意。", a.YearZhi, b.YearZhi, wxA, wxB),
			Weight:      20,
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
	if branchCompatible(a.DayZhi, b.DayZhi) {
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
			Type:        "日柱次吉",
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
	if branchSameElement(a.DayZhi, b.DayZhi) || branchShengElement(a.DayZhi, b.DayZhi) {
		return []CompatibilityEvidence{{
			EvidenceKey: "day_pillar_safe",
			Dimension:   "day_pillar",
			Type:        "日柱安慰",
			Polarity:    "positive",
			Source:      "day_pillar",
			Title:       "日柱安慰分",
			Detail: fmt.Sprintf(
				"日柱 %s%s/%s%s 地支虽不合，但五行相同或相生，亲密层有微弱亲近感。",
				a.DayGan, a.DayZhi, b.DayGan, b.DayZhi,
			),
			Weight: 3,
		}}
	}
	return nil
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
	tierByScore := map[int]struct {
		key   string
		label string
	}{
		10: {"upper", "上档"},
		5:  {"lower", "下档"},
		3:  {"safe", "安慰分"},
	}
	for _, p := range pillars {
		s := scoreDayPillar(p.ganA, p.zhiA, p.ganB, p.zhiB)
		t, ok := tierByScore[s]
		if !ok {
			continue
		}
		out = append(out, CompatibilityEvidence{
			EvidenceKey: "eight_chars_" + p.name + "_" + t.key,
			Dimension:   "eight_chars",
			Type:        p.label + "对" + t.label,
			Polarity:    "positive",
			Source:      "eight_chars",
			Title:       p.label + "对" + t.label,
			Detail: fmt.Sprintf(
				"%s %s%s/%s%s 命中%s（贡献 %d）。",
				p.label, p.ganA, p.zhiA, p.ganB, p.zhiB, t.label, s,
			),
			Weight: s,
		})
	}
	return out
}

// buildScoreExplanationsV3 按 4 模块各出一条解释（zodiac/nayin/day_pillar/eight_chars）。
// 纯加分制下 NegativeFactor / NegativeEvidenceKeys 永远为空。
func buildScoreExplanationsV3(a, b *BaziResult, evidences []CompatibilityEvidence) []CompatibilityScoreExplanation {
	dimensions := []string{"zodiac", "nayin", "day_pillar", "eight_chars"}
	out := make([]CompatibilityScoreExplanation, 0, 4)
	for _, dim := range dimensions {
		hit := findEvidenceByDimension(evidences, dim)
		exp := CompatibilityScoreExplanation{Dimension: dim}
		if hit != nil {
			exp.PositiveFactor = hit.Title
			exp.PositiveEvidenceKeys = []string{hit.EvidenceKey}
		}
		exp.Summary = scoreExplanationSummaryV3(dim, hit, a, b)
		out = append(out, exp)
	}
	return out
}

func findEvidenceByDimension(evidences []CompatibilityEvidence, dim string) *CompatibilityEvidence {
	for i := range evidences {
		if string(evidences[i].Dimension) == dim {
			return &evidences[i]
		}
	}
	return nil
}

func scoreExplanationSummaryV3(dim string, hit *CompatibilityEvidence, a, b *BaziResult) string {
	switch dim {
	case "zodiac":
		if hit == nil {
			return fmt.Sprintf("双方年支 %s/%s 无六合 / 三合 / 同行 / 相生，属相层无加成。", a.YearZhi, b.YearZhi)
		}
		switch hit.EvidenceKey {
		case "zodiac_liuhe":
			return fmt.Sprintf("双方属相 %s/%s 构成六合，关系基础线吸引力强。", a.YearZhi, b.YearZhi)
		case "zodiac_sanhe":
			return fmt.Sprintf("双方属相 %s/%s 同属 %s 三合局，气场协同。",
				a.YearZhi, b.YearZhi, sanheGroupName(a.YearZhi, b.YearZhi))
		case "zodiac_same_element":
			return fmt.Sprintf("双方年支 %s/%s 五行同行（双生），属相层有亲近感。", a.YearZhi, b.YearZhi)
		case "zodiac_sheng":
			return fmt.Sprintf("双方年支 %s/%s 五行相生，属相层有顺承之意。", a.YearZhi, b.YearZhi)
		}
		return ""
	case "nayin":
		wxA := nayinElement(a.YearGan + a.YearZhi)
		wxB := nayinElement(b.YearGan + b.YearZhi)
		if hit == nil {
			return fmt.Sprintf("%s 与 %s 纳音五行相克，纳音层无加分。", wxA, wxB)
		}
		if hit.EvidenceKey == "nayin_sheng" {
			return fmt.Sprintf("%s 与 %s 纳音五行相生，资源 / 情绪流动顺。", wxA, wxB)
		}
		return fmt.Sprintf("双方纳音同为 %s，本质同气。", wxA)
	case "day_pillar":
		if hit == nil {
			return fmt.Sprintf("日柱 %s%s/%s%s 地支不合且五行无亲，亲密层无加成。",
				a.DayGan, a.DayZhi, b.DayGan, b.DayZhi)
		}
		switch hit.EvidenceKey {
		case "day_pillar_upper":
			return fmt.Sprintf("日柱 %s%s/%s%s 地支合且天干五合 / 相生，亲密层结构稳。",
				a.DayGan, a.DayZhi, b.DayGan, b.DayZhi)
		case "day_pillar_lower":
			return fmt.Sprintf("日柱 %s%s/%s%s 地支合，天干仅相同 / 克 / 无关，亲密层有基础但未达上吉。",
				a.DayGan, a.DayZhi, b.DayGan, b.DayZhi)
		case "day_pillar_safe":
			return fmt.Sprintf("日柱 %s%s/%s%s 地支虽不合，但五行同行或相生，亲密层有微弱亲近感。",
				a.DayGan, a.DayZhi, b.DayGan, b.DayZhi)
		}
		return ""
	case "eight_chars":
		// 八字模块可能 0–3 柱命中，单条 evidence 无法表达命中数，故直接重新计算柱位命中。
		return eightCharsSummary(a, b)
	}
	return ""
}

// eightCharsSummary 单独计算 八字 命中柱数并生成 summary。
func eightCharsSummary(a, b *BaziResult) string {
	type p struct{ name, label, ga, za, gb, zb string }
	pillars := []p{
		{"year", "年柱", a.YearGan, a.YearZhi, b.YearGan, b.YearZhi},
		{"month", "月柱", a.MonthGan, a.MonthZhi, b.MonthGan, b.MonthZhi},
		{"hour", "时柱", a.HourGan, a.HourZhi, b.HourGan, b.HourZhi},
	}
	hits := 0
	var soloLabel string
	for _, pp := range pillars {
		if scoreDayPillar(pp.ga, pp.za, pp.gb, pp.zb) > 0 {
			hits++
			soloLabel = pp.label
		}
	}
	switch hits {
	case 0:
		return "年 / 月 / 时三柱均无合，外围层无加成。"
	case 1:
		return fmt.Sprintf("三柱中仅 %s 合，外围层支撑薄弱。", soloLabel)
	default:
		return fmt.Sprintf("年 / 月 / 时三柱中有 %d 柱合，外围层有支撑。", hits)
	}
}

// buildSummaryTagsV3 按 design §4.4 规则生成最多 4 条 tag。
// 总分级 tag（上吉合盘 / 合盘无加成）作为整体定性，优先放在最前；
// 之后按模块顺序追加，最终截断为 4 条。
func buildSummaryTagsV3(scores CompatibilityDimensionScores, total int) []string {
	tags := make([]string, 0, 4)
	if total >= 80 {
		tags = append(tags, "上吉合盘")
	}
	if total < 60 && scores.Zodiac == 0 && scores.Nayin == 0 &&
		scores.DayPillar == 0 && scores.EightChars == 0 {
		tags = append(tags, "合盘无加成")
	}
	if scores.Zodiac >= 50 {
		tags = append(tags, "属相相合")
	}
	if scores.Nayin >= 20 {
		tags = append(tags, "纳音同气")
	}
	if scores.DayPillar == 10 {
		tags = append(tags, "日柱上吉")
	} else if scores.DayPillar == 5 {
		tags = append(tags, "日柱次吉")
	}
	if scores.EightChars >= 14 {
		tags = append(tags, "八字承接好")
	}
	if len(tags) > 4 {
		tags = tags[:4]
	}
	return tags
}
