package bazi

// SpouseStarSignal 描述一个人命盘中「配偶星 + 夫妻宫」的结构化信号，
// 供 LLM 推导「TA 命里理想/容易吸引的另一半画像」。本结构只陈述客观事实，不做性格解读。
type SpouseStarSignal struct {
	Available              bool     // 性别可识别（能定配偶星类别）；false → 上层跳过本节
	Present                bool     // 命盘中存在配偶星
	Category               string   // "财星"(男) / "官杀"(女)；Available=false 时为空
	StarNames              []string // 命中出现的具体十神：正财/偏财 或 正官/七杀（去重保序）
	Positions              []string // 配偶星位置，如 "月干(透)"、"日支(藏)"
	Visible                bool     // 配偶星是否透于天干
	InSpousePalace         bool     // 配偶星是否坐日支（夫妻宫）
	DayBranchHiddenShiShen []string // 日支藏干各自的十神（夫妻宫画像主料）
}

// DetectSpouseStarSignal 按性别定位配偶星（男看财星、女看官杀），扫描四柱天干（透）
// 与地支藏干（藏）的十神，并附上日支藏干十神。性别缺失或盘异常时返回 Available=false。
func DetectSpouseStarSignal(r *BaziResult) SpouseStarSignal {
	if r == nil || r.DayGan == "" {
		return SpouseStarSignal{}
	}

	var category string
	targets := map[string]bool{}
	switch r.Gender {
	case "male":
		category = "财星"
		targets["正财"] = true
		targets["偏财"] = true
	case "female":
		category = "官杀"
		targets["正官"] = true
		targets["七杀"] = true
	default:
		return SpouseStarSignal{Available: false}
	}

	sig := SpouseStarSignal{Available: true, Category: category}
	sig.DayBranchHiddenShiShen = append([]string(nil), r.DayZhiShiShen...)

	seen := map[string]bool{}
	addStar := func(name string) {
		if !seen[name] {
			seen[name] = true
			sig.StarNames = append(sig.StarNames, name)
		}
	}

	// 透干：四柱天干十神
	ganPillars := []struct {
		label string
		ss    string
	}{
		{"年干", r.YearGanShiShen},
		{"月干", r.MonthGanShiShen},
		{"日干", r.DayGanShiShen},
		{"时干", r.HourGanShiShen},
	}
	for _, p := range ganPillars {
		if targets[p.ss] {
			sig.Present = true
			sig.Visible = true
			sig.Positions = append(sig.Positions, p.label+"(透)")
			addStar(p.ss)
		}
	}

	// 藏干：四柱地支藏干十神
	zhiPillars := []struct {
		label    string
		isDayZhi bool
		ss       []string
	}{
		{"年支", false, r.YearZhiShiShen},
		{"月支", false, r.MonthZhiShiShen},
		{"日支", true, r.DayZhiShiShen},
		{"时支", false, r.HourZhiShiShen},
	}
	for _, p := range zhiPillars {
		for _, s := range p.ss {
			if targets[s] {
				sig.Present = true
				sig.Positions = append(sig.Positions, p.label+"(藏)")
				if p.isDayZhi {
					sig.InSpousePalace = true
				}
				addStar(s)
			}
		}
	}

	return sig
}
