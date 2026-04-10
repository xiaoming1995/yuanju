package bazi

// DayunJixiong 大运综合吉凶评价（优先级规则驱动）
type DayunJixiong struct {
	Level  string `json:"level"`  // 好运/霉运/平常运
	Basis  string `json:"basis"`  // 调候/扶抑
	Reason string `json:"reason"` // 说明文字
}

// mergeDayunWuxing 将大运天干/地支五行并入原局五行，返回合局统计
// 合局代表命主在该大运期间实际受到的五行环境
func mergeDayunWuxing(original WuxingStats, dayunGanWx, dayunZhiWx string) WuxingStats {
	m := original
	m.Total += 2
	for _, wx := range []string{dayunGanWx, dayunZhiWx} {
		switch wx {
		case "木":
			m.Mu++
		case "火":
			m.Huo++
		case "土":
			m.Tu++
		case "金":
			m.Jin++
		case "水":
			m.Shui++
		}
	}
	t := float64(m.Total)
	m.MuPct = float64(m.Mu) / t * 100
	m.HuoPct = float64(m.Huo) / t * 100
	m.TuPct = float64(m.Tu) / t * 100
	m.JinPct = float64(m.Jin) / t * 100
	m.ShuiPct = float64(m.Shui) / t * 100
	return m
}

// CalcDayunJixiong 按优先级规则评判大运吉凶
//
// 公式作用于原局+大运合局：先将大运干支五行并入原局，再对合局状态应用规则。
//
// 优先级：
//  1. 合局至寒/至热 → 调候（火/水）最急，超越扶抑
//  2. 合局冷热适度 → 扶抑优先（身强则抑，身弱则扶）
//
// 注：刑冲克合应期判断需完整柱信息，暂由 LLM 层处理，待后续迭代实现。
func CalcDayunJixiong(dayunGanWx, dayunZhiWx, dayGanWx, monthZhiWx string, originalWuxing WuxingStats) *DayunJixiong {
	// 合局：原局八字 + 大运天干 + 大运地支
	merged := mergeDayunWuxing(originalWuxing, dayunGanWx, dayunZhiWx)

	// ── 规则一：合局极寒/极热 → 调候优先 ────────────────────────────
	cfg := GetAlgoConfig()
	// 极寒：合局无火 且 寒性（水+金）≥ jiHanMin
	isJiHan := merged.Huo == 0 && (merged.Shui+merged.Jin) >= cfg.JiHanMin
	// 极热：合局无水 且 暖性（火+木）≥ jiReMin
	isJiRe := merged.Shui == 0 && (merged.Huo+merged.Mu) >= cfg.JiReMin

	if isJiHan {
		// 合局至寒，需要火来调候；大运若带火则已起调候作用（已并入合局）
		if dayunGanWx == "火" || dayunZhiWx == "火" {
			return &DayunJixiong{Level: "好运", Basis: "调候", Reason: "合局至寒，大运透火入局，调候得解，吉。"}
		}
		return &DayunJixiong{Level: "霉运", Basis: "调候", Reason: "合局至寒，大运无火，寒极不化，凶。"}
	}

	if isJiRe {
		// 合局至热，需要水来调候
		if dayunGanWx == "水" || dayunZhiWx == "水" {
			return &DayunJixiong{Level: "好运", Basis: "调候", Reason: "合局至热，大运透水入局，调候得济，吉。"}
		}
		return &DayunJixiong{Level: "霉运", Basis: "调候", Reason: "合局至热，大运无水，热极不降，凶。"}
	}

	// ── 规则二：合局冷热适度 → 扶抑优先 ─────────────────────────────
	// 对合局重新推断用神/忌神（反映大运期间的实际强弱）
	yongshen, jishen := calcWeightedYongshen(dayGanWx, monthZhiWx, merged)

	score := 0
	for _, wx := range []string{dayunGanWx, dayunZhiWx} {
		if yongshen != "" && containsRune(yongshen, wx) {
			score++
		}
		if jishen != "" && containsRune(jishen, wx) {
			score--
		}
	}

	switch {
	case score > 0:
		return &DayunJixiong{Level: "好运", Basis: "扶抑", Reason: "大运五行入局后合用神，顺势利导，吉。"}
	case score < 0:
		return &DayunJixiong{Level: "霉运", Basis: "扶抑", Reason: "大运五行入局后触忌神，逆势阻滞，凶。"}
	default:
		return &DayunJixiong{Level: "平常运", Basis: "-", Reason: "大运五行入局后用忌相抵，平常论之。"}
	}
}

// containsRune 判断五行字符串（如 "水木"）是否包含某五行单字
func containsRune(wuxingGroup, wx string) bool {
	for _, r := range wuxingGroup {
		if string(r) == wx {
			return true
		}
	}
	return false
}
