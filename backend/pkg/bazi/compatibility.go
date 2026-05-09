package bazi

import (
	"fmt"
	"sort"
)

type CompatibilityLevel string

const (
	CompatibilityHigh   CompatibilityLevel = "high"
	CompatibilityMedium CompatibilityLevel = "medium"
	CompatibilityLow    CompatibilityLevel = "low"
)

type CompatibilityPolarity string

const (
	CompatibilityPositive CompatibilityPolarity = "positive"
	CompatibilityNegative CompatibilityPolarity = "negative"
	CompatibilityMixed    CompatibilityPolarity = "mixed"
	CompatibilityNeutral  CompatibilityPolarity = "neutral"
)

type CompatibilityDimension string

const (
	CompatibilityAttraction    CompatibilityDimension = "attraction"
	CompatibilityStability     CompatibilityDimension = "stability"
	CompatibilityCommunication CompatibilityDimension = "communication"
	CompatibilityPracticality  CompatibilityDimension = "practicality"
)

type CompatibilityEvidence struct {
	Dimension CompatibilityDimension `json:"dimension"`
	Type      string                 `json:"type"`
	Polarity  CompatibilityPolarity  `json:"polarity"`
	Source    string                 `json:"source"`
	Title     string                 `json:"title"`
	Detail    string                 `json:"detail"`
	Weight    int                    `json:"weight"`
}

type CompatibilityDimensionScores struct {
	Attraction    int `json:"attraction"`
	Stability     int `json:"stability"`
	Communication int `json:"communication"`
	Practicality  int `json:"practicality"`
}

type CompatibilityAnalysis struct {
	OverallLevel    CompatibilityLevel           `json:"overall_level"`
	DimensionScores CompatibilityDimensionScores `json:"dimension_scores"`
	Evidences       []CompatibilityEvidence      `json:"evidences"`
	SummaryTags     []string                     `json:"summary_tags"`
}

func AnalyzeCompatibility(a, b *BaziResult) CompatibilityAnalysis {
	scores := CompatibilityDimensionScores{
		Attraction:    60,
		Stability:     60,
		Communication: 60,
		Practicality:  60,
	}
	evidences := make([]CompatibilityEvidence, 0, 8)
	tags := make([]string, 0, 4)

	addEvidence := func(item CompatibilityEvidence) {
		evidences = append(evidences, item)
		switch item.Dimension {
		case CompatibilityAttraction:
			scores.Attraction = clampScore(scores.Attraction + item.Weight)
		case CompatibilityStability:
			scores.Stability = clampScore(scores.Stability + item.Weight)
		case CompatibilityCommunication:
			scores.Communication = clampScore(scores.Communication + item.Weight)
		case CompatibilityPracticality:
			scores.Practicality = clampScore(scores.Practicality + item.Weight)
		}
	}

	// 1. 日主关系
	switch relation := dayMasterRelation(a.DayGanWuxing, b.DayGanWuxing); relation {
	case "same":
		addEvidence(CompatibilityEvidence{
			Dimension: CompatibilityCommunication,
			Type:      "日主同气",
			Polarity:  CompatibilityPositive,
			Source:    "day_master",
			Title:     "日主同气",
			Detail:    fmt.Sprintf("双方日主同属%s，底层气质相近，较容易理解彼此节奏。", a.DayGanWuxing),
			Weight:    8,
		})
	case "sheng":
		addEvidence(CompatibilityEvidence{
			Dimension: CompatibilityCommunication,
			Type:      "日主相生",
			Polarity:  CompatibilityPositive,
			Source:    "day_master",
			Title:     "日主相生",
			Detail:    fmt.Sprintf("双方日主五行形成相生关系（%s / %s），沟通与情感流动更顺。", a.DayGanWuxing, b.DayGanWuxing),
			Weight:    10,
		})
		addEvidence(CompatibilityEvidence{
			Dimension: CompatibilityAttraction,
			Type:      "日主相生",
			Polarity:  CompatibilityPositive,
			Source:    "day_master",
			Title:     "气机互引",
			Detail:    "双方气机能够互相牵引，初始吸引感更容易建立。",
			Weight:    6,
		})
	case "ke":
		addEvidence(CompatibilityEvidence{
			Dimension: CompatibilityCommunication,
			Type:      "日主相克",
			Polarity:  CompatibilityNegative,
			Source:    "day_master",
			Title:     "日主相克",
			Detail:    fmt.Sprintf("双方日主五行互见克制（%s / %s），表达方式和主导感较容易顶撞。", a.DayGanWuxing, b.DayGanWuxing),
			Weight:    -12,
		})
		addEvidence(CompatibilityEvidence{
			Dimension: CompatibilityStability,
			Type:      "日主相克",
			Polarity:  CompatibilityNegative,
			Source:    "day_master",
			Title:     "关系拉扯",
			Detail:    "若缺少其他缓冲结构，长期相处更容易形成拉扯。",
			Weight:    -8,
		})
	}

	// 2. 五行互补 / 偏枯
	if detail, ok, weight := fiveElementComplement(a, b); ok {
		addEvidence(CompatibilityEvidence{
			Dimension: CompatibilityPracticality,
			Type:      "五行互补",
			Polarity:  CompatibilityPositive,
			Source:    "five_elements",
			Title:     "五行互补",
			Detail:    detail,
			Weight:    weight,
		})
	} else if detail, ok, weight := fiveElementImbalance(a, b); ok {
		addEvidence(CompatibilityEvidence{
			Dimension: CompatibilityPracticality,
			Type:      "五行失衡",
			Polarity:  CompatibilityNegative,
			Source:    "five_elements",
			Title:     "五行失衡",
			Detail:    detail,
			Weight:    weight,
		})
	}

	// 3. 夫妻宫互动（日支）
	if sixHe[a.DayZhi] == b.DayZhi {
		addEvidence(CompatibilityEvidence{
			Dimension: CompatibilityStability,
			Type:      "夫妻宫六合",
			Polarity:  CompatibilityPositive,
			Source:    "spouse_palace",
			Title:     "夫妻宫六合",
			Detail:    fmt.Sprintf("双方日支 %s / %s 构成六合，婚恋靠近感和关系推进意愿更强。", a.DayZhi, b.DayZhi),
			Weight:    18,
		})
		addEvidence(CompatibilityEvidence{
			Dimension: CompatibilityAttraction,
			Type:      "夫妻宫六合",
			Polarity:  CompatibilityPositive,
			Source:    "spouse_palace",
			Title:     "亲近感强",
			Detail:    "日支六合通常会放大彼此愿意靠近与接纳的倾向。",
			Weight:    8,
		})
	}
	if sixChong[a.DayZhi] == b.DayZhi {
		addEvidence(CompatibilityEvidence{
			Dimension: CompatibilityStability,
			Type:      "夫妻宫六冲",
			Polarity:  CompatibilityNegative,
			Source:    "spouse_palace",
			Title:     "夫妻宫六冲",
			Detail:    fmt.Sprintf("双方日支 %s / %s 形成六冲，亲密关系中的节奏和需求容易正面碰撞。", a.DayZhi, b.DayZhi),
			Weight:    -18,
		})
		addEvidence(CompatibilityEvidence{
			Dimension: CompatibilityCommunication,
			Type:      "夫妻宫六冲",
			Polarity:  CompatibilityNegative,
			Source:    "spouse_palace",
			Title:     "情绪冲撞",
			Detail:    "日支六冲往往让相处中的情绪起伏更大，争执成本更高。",
			Weight:    -8,
		})
	}
	if sixXing[a.DayZhi] == b.DayZhi || sixXing[b.DayZhi] == a.DayZhi {
		addEvidence(CompatibilityEvidence{
			Dimension: CompatibilityStability,
			Type:      "夫妻宫刑害",
			Polarity:  CompatibilityNegative,
			Source:    "spouse_palace",
			Title:     "夫妻宫刑害",
			Detail:    fmt.Sprintf("双方日支 %s / %s 存在刑害结构，关系中更容易积累别扭和隐性损耗。", a.DayZhi, b.DayZhi),
			Weight:    -10,
		})
	}

	// 4. 配偶星呼应
	if spouseDetail, ok := spouseStarResonance(a, b); ok {
		addEvidence(CompatibilityEvidence{
			Dimension: CompatibilityAttraction,
			Type:      "配偶星呼应",
			Polarity:  CompatibilityPositive,
			Source:    "spouse_star",
			Title:     "配偶星呼应",
			Detail:    spouseDetail,
			Weight:    14,
		})
		addEvidence(CompatibilityEvidence{
			Dimension: CompatibilityStability,
			Type:      "配偶星呼应",
			Polarity:  CompatibilityPositive,
			Source:    "spouse_star",
			Title:     "婚恋指向明确",
			Detail:    "一方命盘中的婚恋星被对方显著触发，关系更容易朝伴侣关系理解。",
			Weight:    6,
		})
	}

	// 5. 干支合冲总量
	if zhiClashCount := branchClashCount(a, b); zhiClashCount >= 2 {
		addEvidence(CompatibilityEvidence{
			Dimension: CompatibilityPracticality,
			Type:      "干支冲克偏多",
			Polarity:  CompatibilityNegative,
			Source:    "ganzhi",
			Title:     "冲克偏多",
			Detail:    fmt.Sprintf("双方盘面存在 %d 处以上明显地支冲克，现实层面的磨合成本会更高。", zhiClashCount),
			Weight:    -10,
		})
	}

	// 6. 神煞辅助
	if hasRomanceShensha(a) || hasRomanceShensha(b) {
		addEvidence(CompatibilityEvidence{
			Dimension: CompatibilityAttraction,
			Type:      "桃花助缘",
			Polarity:  CompatibilityPositive,
			Source:    "shensha",
			Title:     "桃花助缘",
			Detail:    "盘中带有桃花/天喜/红艳类神煞，能增强关系中的浪漫和被吸引感。",
			Weight:    6,
		})
	}
	if hasLonelyShensha(a) || hasLonelyShensha(b) {
		addEvidence(CompatibilityEvidence{
			Dimension: CompatibilityStability,
			Type:      "孤寡错位",
			Polarity:  CompatibilityNegative,
			Source:    "shensha",
			Title:     "孤寡错位",
			Detail:    "盘中带有孤辰/寡宿类神煞，容易放大距离感、误读或情绪疏离。",
			Weight:    -6,
		})
	}

	sort.SliceStable(evidences, func(i, j int) bool {
		wi, wj := absInt(evidences[i].Weight), absInt(evidences[j].Weight)
		if wi == wj {
			return evidences[i].Type < evidences[j].Type
		}
		return wi > wj
	})

	tags = buildCompatibilityTags(scores, evidences)

	return CompatibilityAnalysis{
		OverallLevel:    aggregateCompatibilityLevel(scores, evidences),
		DimensionScores: scores,
		Evidences:       evidences,
		SummaryTags:     tags,
	}
}

func dayMasterRelation(aWx, bWx string) string {
	if aWx == "" || bWx == "" {
		return ""
	}
	if aWx == bWx {
		return "same"
	}
	if wxSheng[aWx] == bWx || wxSheng[bWx] == aWx {
		return "sheng"
	}
	if wxKe[aWx] == bWx || wxKe[bWx] == aWx {
		return "ke"
	}
	return ""
}

func fiveElementComplement(a, b *BaziResult) (string, bool, int) {
	aStats := compatibilityStats(a)
	bStats := compatibilityStats(b)
	missingA := missingElements(aStats)
	missingB := missingElements(bStats)
	if len(missingA) == 0 || len(missingB) == 0 {
		return "", false, 0
	}
	helpA, helpB := 0, 0
	for _, wx := range missingA {
		if bStats[wx] > 0 {
			helpA++
		}
	}
	for _, wx := range missingB {
		if aStats[wx] > 0 {
			helpB++
		}
	}
	if helpA == 0 && helpB == 0 {
		return "", false, 0
	}
	weight := 8
	if helpA > 0 && helpB > 0 {
		weight = 12
	}
	return fmt.Sprintf("双方五行短板存在互补关系：A 缺 %v、B 可补；B 缺 %v、A 可补。", missingA, missingB), true, weight
}

func fiveElementImbalance(a, b *BaziResult) (string, bool, int) {
	aStats := compatibilityStats(a)
	bStats := compatibilityStats(b)
	sharedMissing := make([]string, 0, 2)
	for _, wx := range []string{"木", "火", "土", "金", "水"} {
		if aStats[wx] == 0 && bStats[wx] == 0 {
			sharedMissing = append(sharedMissing, wx)
		}
	}
	if len(sharedMissing) > 0 {
		return fmt.Sprintf("双方共同缺少 %v，现实协同上容易出现同类短板叠加。", sharedMissing), true, -10
	}
	return "", false, 0
}

func spouseStarResonance(self, partner *BaziResult) (string, bool) {
	targets := []string{}
	if self.Gender == "male" {
		targets = []string{"正财", "偏财"}
	} else {
		targets = []string{"正官", "七杀"}
	}
	for _, gan := range collectCompatibilityGans(partner) {
		shishen := GetShiShen(self.DayGan, gan)
		for _, target := range targets {
			if shishen == target {
				return fmt.Sprintf("以%s日主观之，对方命盘中的 %s 对应你的%s，婚恋指向较明确。", self.DayGan, gan, target), true
			}
		}
	}
	return "", false
}

func branchClashCount(a, b *BaziResult) int {
	az := []string{a.YearZhi, a.MonthZhi, a.DayZhi, a.HourZhi}
	bz := []string{b.YearZhi, b.MonthZhi, b.DayZhi, b.HourZhi}
	count := 0
	for _, left := range az {
		for _, right := range bz {
			if sixChong[left] == right {
				count++
			}
		}
	}
	return count
}

func hasRomanceShensha(r *BaziResult) bool {
	for _, item := range appendShensha(nil, r) {
		if item == "桃花" || item == "天喜" || item == "红艳" {
			return true
		}
	}
	return false
}

func hasLonelyShensha(r *BaziResult) bool {
	for _, item := range appendShensha(nil, r) {
		if item == "孤辰" || item == "寡宿" {
			return true
		}
	}
	return false
}

func appendShensha(dst []string, r *BaziResult) []string {
	dst = append(dst, r.YearShenSha...)
	dst = append(dst, r.MonthShenSha...)
	dst = append(dst, r.DayShenSha...)
	dst = append(dst, r.HourShenSha...)
	return dst
}

func compatibilityStats(r *BaziResult) map[string]int {
	if r.Wuxing.Total > 0 {
		return map[string]int{
			"木": r.Wuxing.Mu,
			"火": r.Wuxing.Huo,
			"土": r.Wuxing.Tu,
			"金": r.Wuxing.Jin,
			"水": r.Wuxing.Shui,
		}
	}
	stats := map[string]int{"木": 0, "火": 0, "土": 0, "金": 0, "水": 0}
	for _, wx := range []string{
		r.YearGanWuxing, r.MonthGanWuxing, r.DayGanWuxing, r.HourGanWuxing,
		r.YearZhiWuxing, r.MonthZhiWuxing, r.DayZhiWuxing, r.HourZhiWuxing,
	} {
		if wx != "" {
			stats[wx]++
		}
	}
	return stats
}

func missingElements(stats map[string]int) []string {
	out := make([]string, 0, 2)
	for _, wx := range []string{"木", "火", "土", "金", "水"} {
		if stats[wx] == 0 {
			out = append(out, wx)
		}
	}
	return out
}

func collectCompatibilityGans(r *BaziResult) []string {
	gans := []string{r.YearGan, r.MonthGan, r.DayGan, r.HourGan}
	gans = append(gans, r.YearHideGan...)
	gans = append(gans, r.MonthHideGan...)
	gans = append(gans, r.DayHideGan...)
	gans = append(gans, r.HourHideGan...)
	return gans
}

func aggregateCompatibilityLevel(scores CompatibilityDimensionScores, evidences []CompatibilityEvidence) CompatibilityLevel {
	avg := (scores.Attraction + scores.Stability + scores.Communication + scores.Practicality) / 4
	negativeHeavy := 0
	for _, item := range evidences {
		if item.Polarity == CompatibilityNegative && item.Weight <= -12 {
			negativeHeavy++
		}
	}
	if scores.Stability < 50 || (avg < 58 && negativeHeavy > 0) {
		return CompatibilityLow
	}
	if avg >= 70 && scores.Stability >= 65 && negativeHeavy == 0 {
		return CompatibilityHigh
	}
	return CompatibilityMedium
}

func buildCompatibilityTags(scores CompatibilityDimensionScores, evidences []CompatibilityEvidence) []string {
	tags := make([]string, 0, 4)
	if scores.Attraction >= 70 {
		tags = append(tags, "吸引力强")
	}
	if scores.Stability >= 70 {
		tags = append(tags, "稳定度高")
	}
	if scores.Communication >= 70 {
		tags = append(tags, "沟通顺畅")
	}
	if scores.Practicality >= 70 {
		tags = append(tags, "现实互补")
	}
	if scores.Stability < 55 {
		tags = append(tags, "关系波动")
	}
	if scores.Practicality < 55 {
		tags = append(tags, "磨合成本高")
	}
	if len(tags) == 0 && len(evidences) > 0 {
		tags = append(tags, evidences[0].Title)
	}
	if len(tags) > 4 {
		tags = tags[:4]
	}
	return tags
}

func clampScore(v int) int {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
