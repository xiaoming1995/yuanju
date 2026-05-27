package bazi

// findSanheGroup 在 sanheGroups 中查找同时包含 a 与 b 的三合组。
// 命中返回该组（如 ["申","子","辰"]）与 true；同支、空字符串或未命中返回零值与 false。
// 该函数集中保管唯一的 sanhe 查找循环，供 branchCompatibilityKind / sanheGroupName 复用。
func findSanheGroup(a, b string) ([3]string, bool) {
	if a == "" || b == "" || a == b {
		return [3]string{}, false
	}
	for _, group := range sanheGroups {
		hasA, hasB := false, false
		for _, z := range group {
			if z == a {
				hasA = true
			}
			if z == b {
				hasB = true
			}
		}
		if hasA && hasB {
			return group, true
		}
	}
	return [3]string{}, false
}

// branchCompatibilityKind 返回支合的类型，用于 evidence 文案区分：
//
//	"liuhe" - 六合
//	"sanhe" - 三合（半三合）
//	""      - 不合
//
// 同支 (a == b) 与空字符串均返回 ""。
func branchCompatibilityKind(a, b string) string {
	if a == "" || b == "" || a == b {
		return ""
	}
	if sixHe[a] == b {
		return "liuhe"
	}
	if _, ok := findSanheGroup(a, b); ok {
		return "sanhe"
	}
	return ""
}

// branchCompatible 判定两个地支是否构成「支合」：
//   - 六合（子丑/寅亥/卯戌/辰酉/巳申/午未）
//   - 三合（申子辰/亥卯未/巳酉丑/寅午戌 中任选两支不相等者，即「半三合」）
//
// 同支返回 false（自刑/单纯重复均不算合）；空字符串返回 false。
func branchCompatible(a, b string) bool {
	return branchCompatibilityKind(a, b) != ""
}

// sanheGroupName 返回包含 a 与 b 的三合局名（如 "申子辰" 水局），未命中返回空。
func sanheGroupName(a, b string) string {
	group, ok := findSanheGroup(a, b)
	if !ok {
		return ""
	}
	return group[0] + group[1] + group[2]
}

// scoreZodiac 计算「合属相」模块得分（满分 50）。
// 输入为两人年支；命中六合或三合（含半三合）即得 50，否则 0。
func scoreZodiac(yearZhiA, yearZhiB string) int {
	if branchCompatible(yearZhiA, yearZhiB) {
		return 50
	}
	return 0
}

// scoreNayin 计算「合纳音」模块得分（满分 20）。
// 输入为两人年柱干支字符串（如 "甲子"）；
// 纳音五行 相生 或 相同 即得 20，相克或无法识别返回 0。
func scoreNayin(yearGanZhiA, yearGanZhiB string) int {
	wxA := nayinElement(yearGanZhiA)
	wxB := nayinElement(yearGanZhiB)
	switch nayinRelation(wxA, wxB) {
	case "sheng", "same":
		return 20
	default:
		return 0
	}
}

// scoreDayPillar 计算「合日柱」模块得分（满分 10）。
// 上档 (10)：日支合 + (干五合 OR 干五行相生)
// 下档 (5) ：日支合 + (干五行相同 OR 干相克 OR 干无关)
// 不命中(0)：日支不合，无论干如何
func scoreDayPillar(dayGanA, dayZhiA, dayGanB, dayZhiB string) int {
	if !branchCompatible(dayZhiA, dayZhiB) {
		return 0
	}
	if ganUpperTier(dayGanA, dayGanB) {
		return 10
	}
	return 5
}

// ganUpperTier 判定两天干是否构成「合日柱上档」的强化条件：
//   - 天干五合（甲己/乙庚/丙辛/丁壬/戊癸）
//   - 天干五行相生
//
// 干相同 / 干相克 / 干无关 一律返回 false（落到下档 5 分）。
func ganUpperTier(a, b string) bool {
	if a == "" || b == "" {
		return false
	}
	if _, ok := ganWuhe[[2]string{a, b}]; ok {
		return true
	}
	wxA := ganWuxing[a]
	wxB := ganWuxing[b]
	if wxA == "" || wxB == "" {
		return false
	}
	if wxA == wxB {
		return false // 同行 → 下档（不算上档）
	}
	if wxSheng[wxA] == wxB || wxSheng[wxB] == wxA {
		return true
	}
	return false
}

// scoreEightChars 计算「合八字」模块得分（满分 20）。
// 输入为年/月/时三柱（不含日柱）双方的干支。每柱独立按 scoreDayPillar 规则得 0/5/10。
// 三柱总和 ∈ [0,30]，归一化到 [0,20]：(sum*2 + 1) / 3（整数四舍五入）。
func scoreEightChars(
	yearGanA, yearZhiA, yearGanB, yearZhiB string,
	monthGanA, monthZhiA, monthGanB, monthZhiB string,
	hourGanA, hourZhiA, hourGanB, hourZhiB string,
) int {
	y := scoreDayPillar(yearGanA, yearZhiA, yearGanB, yearZhiB)
	m := scoreDayPillar(monthGanA, monthZhiA, monthGanB, monthZhiB)
	h := scoreDayPillar(hourGanA, hourZhiA, hourGanB, hourZhiB)
	return normalizeEightCharsSum(y + m + h)
}

// normalizeEightCharsSum 把三柱总和（0..30）归一化到 [0,20]，整数四舍五入。
func normalizeEightCharsSum(sum int) int {
	if sum <= 0 {
		return 0
	}
	if sum >= 30 {
		return 20
	}
	return (sum*2 + 1) / 3
}

// branchSameElement 判定两个地支非空、不相等、且五行相同（"双生"）。
// 复用 event_signals.go 的 zhiWuxing 映射。
func branchSameElement(a, b string) bool {
	if a == "" || b == "" || a == b {
		return false
	}
	wxA, wxB := zhiWuxing[a], zhiWuxing[b]
	if wxA == "" || wxB == "" {
		return false
	}
	return wxA == wxB
}

// branchShengElement 判定两个地支非空、不相等、且五行存在 相生 关系（任一方向）。
func branchShengElement(a, b string) bool {
	if a == "" || b == "" || a == b {
		return false
	}
	wxA, wxB := zhiWuxing[a], zhiWuxing[b]
	if wxA == "" || wxB == "" {
		return false
	}
	if wxA == wxB {
		return false
	}
	return wxSheng[wxA] == wxB || wxSheng[wxB] == wxA
}
