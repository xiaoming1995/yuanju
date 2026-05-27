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
