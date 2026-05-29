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
			Detail:      fmt.Sprintf("你俩的属相（年支 %s / %s）构成了「六合」——这是属相搭配里最讨喜、最顺的一种。落到相处上，就是你们见面容易互相来电、自来熟，很多事不用刻意经营就能对上眼。这种天生的亲近感，会让你们在磨合期少很多无谓的摩擦。不过它管的是「合不合得来」，关系能走多远，长久还得看两个人愿不愿意一起用心经营。", a.YearZhi, b.YearZhi),
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
			Detail:      fmt.Sprintf("你俩的属相（年支 %s / %s）同属「%s 三合」——这是一种气场很合拍的属相组合。相处时你们更容易步调一致、想到一块去，遇事也常本能地站在同一边。比起针锋相对，你俩更像天然的同盟，这对关系的稳定是实打实的加分。当然，合得来不等于不用沟通，重要的事还是要摊开说清楚。", a.YearZhi, b.YearZhi, group),
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
			Detail:      fmt.Sprintf("你俩的属相（年支 %s / %s）五行都属%s，命理里叫「双生」——本质上是同一类能量。这意味着你们性子和节奏比较像，容易理解彼此，有种「同类」的天然亲切感。相处起来不太需要费力解释自己，对方往往一点就通。要留意的是，太像有时也会少了点互补，碰到同一类短板时容易一起卡住。", a.YearZhi, b.YearZhi, wx),
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
			Detail:      fmt.Sprintf("你俩的属相（年支 %s / %s）构成「%s 生 %s」的五行相生——一方天然能滋养、托举另一方。相处里常表现为一个愿意付出、一个被照顾，关系有种顺其自然的承接感。这种「你帮我、我托你」的流动，是长期相处很舒服的底子。只要别让付出长期单方向倾斜，这份相生就能一直顺下去。", a.YearZhi, b.YearZhi, wxA, wxB),
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
			Detail:      fmt.Sprintf("你俩的「纳音」五行是 %s 与 %s 相生——纳音说的是两个人骨子里的底色气质。相生意味着这两种底色能互相滋养，相处时情绪和资源都流动得比较顺，不太会互相消耗。日子久了你们会发现，跟对方在一起更像「回血」而不是「耗电」。这是一段关系里很难得、也很值钱的底层契合。", wxA, wxB),
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
			Detail:      fmt.Sprintf("你俩的「纳音」五行同为 %s——纳音是两个人骨子里的底色气质，同气说明你们本质上是一类人。这种同频会让你们天然懂彼此的在意和顾虑，很多感受不用说出口对方也能体会，默契感比一般人强不少。唯一要提醒的是，太同步时也容易一起钻牛角尖，偶尔需要有一个人先跳出来踩刹车。", wxA),
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
					"日柱（%s%s / %s%s）是命盘里最贴近婚恋、最代表「枕边人」的一根柱子，你俩在这里咬合得很到位——地支相合、天干也彼此呼应。这说明在亲密关系的核心地带，你们有天然的契合和稳定结构。相处中容易有那种「找对了人」的踏实感，亲密和信任都建立得比较顺。这是合盘里分量很重的一个好信号，值得珍惜。",
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
				"日柱（%s%s / %s%s）代表两个人最贴近婚恋的核心，你俩的地支在这里是相合的——亲密关系有不错的底子。只是天干层面没能再进一步互相加成（只是相同、相克或不相干），所以这份契合算「够好」但还没到顶配。日常相处大方向是合的，偶尔在细节和默契上需要多一点磨合。把沟通做扎实，这段亲密就能稳稳地往上走。",
				a.DayGan, a.DayZhi, b.DayGan, b.DayZhi,
			),
			Weight: 5,
		}}
	}
	if branchSameElement(a.DayZhi, b.DayZhi) || branchShengElement(a.DayZhi, b.DayZhi) {
		return []CompatibilityEvidence{{
			EvidenceKey: "day_pillar_safe",
			Dimension:   "day_pillar",
			Type:        "日柱安慰分",
			Polarity:    "positive",
			Source:      "day_pillar",
			Title:       "日柱安慰分",
			Detail: fmt.Sprintf(
				"日柱（%s%s / %s%s）是两个人最贴近婚恋的核心。你俩的日支虽然没有直接相合，但五行上是相同或相生的，所以亲密层还是留了一丝天然的亲近感。这说明你们不是格格不入，底子里有可以亲近的余地，只是要比「天生一对」那种多花点心思去经营。把相处的节奏和沟通磨顺，这点微弱的亲近感是能养大的，别因为起步平淡就轻易否定它。",
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
				"%s（%s%s / %s%s）也合上了——这是你们在生活外围（家世背景、日常相处、未来安排这些围绕婚恋的方面）的一处天然契合。它不像日柱那样直接管亲密核心，但能在周边给关系搭把手，让你们在现实层面更容易对得上。这类外围的合拍越多，往后一起过日子越省心。",
				p.label, p.ganA, p.zhiA, p.ganB, p.zhiB,
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
