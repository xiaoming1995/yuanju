// Package bazi - 经典命理用神算法（按命主公式严格实现）
//
// 公式: 调候为急、扶抑优先
//   - 原局至寒（亥子丑月 且 透干无丙丁、地支无巳午）→ 调候用火
//   - 原局至热（巳午未月 且 透干无壬癸、地支无亥子）→ 调候用水
//   - 冷热适度 → 走扶抑 (月支+2加权 helpPct > 40% = 身强)
//
// 用神粒度: 主从级 (PrimaryGan 透干优先 + WuxingSet 大粒度兜底)
//
// 仅用于过往事件推算路径 (GenerateDayunSummariesStream)，
// 不修改 yongshen.go 原算法、不写回数据库、不影响其他 AI 报告。
package bazi

import "fmt"

// ClassicalYongshenStrategy 用神推算路径标签
const (
	ClassicalStrategyTiaohouCold = "tiaohou_cold" // 至寒，调候为急，用神=火
	ClassicalStrategyTiaohouHot  = "tiaohou_hot"  // 至热，调候为急，用神=水
	ClassicalStrategyFuyiStrong  = "fuyi_strong"  // 冷热适度 + 身强，用神=克泄耗
	ClassicalStrategyFuyiWeak    = "fuyi_weak"    // 冷热适度 + 身弱，用神=生扶
)

// ClassicalYongshenResult 经典用神算法输出
type ClassicalYongshenResult struct {
	Strategy string // ClassicalStrategy* 之一
	IsStrong bool   // 身强弱（仅 fuyi_* 时有效）

	PrimaryGan string   // 主用神天干 (透干优先, 如 "丙")
	WuxingSet  string   // 用神五行集 (如 "金土火")
	JishenSet  string   // 忌神五行集 (如 "水木")
	JishenGans []string // 忌神天干列表 (如 ["甲","乙","壬","癸"])

	Reason string // 自然语言说明
}

// ComputeClassicalYongshen 按命主公式推算用神
//
// 流程:
//
//	Step 1: 月令派寒热判定
//	Step 2: 冷热适度时, 月支+2加权扶抑判断身强弱
//	Step 3: 用神五行集 (身强=克泄耗 / 身弱=生扶 / 至寒=火 / 至热=水)
//	Step 4: 透干优先选 primary_gan
//
// 不调用 yongshen.go 的 inferYongshenWithTiaohouPriority，
// 独立路径，输出仅用于覆盖 collect* 函数读取的 natal.Yongshen/Jishen。
func ComputeClassicalYongshen(natal *BaziResult) ClassicalYongshenResult {
	if natal == nil {
		return ClassicalYongshenResult{}
	}

	natalGans := []string{natal.YearGan, natal.MonthGan, natal.DayGan, natal.HourGan}
	natalZhis := []string{natal.YearZhi, natal.MonthZhi, natal.DayZhi, natal.HourZhi}

	// Step 1: 寒热判定
	if isExtremeCold(natal.MonthZhi, natalGans, natalZhis) {
		return buildTiaohouResult(natal, "cold")
	}
	if isExtremeHot(natal.MonthZhi, natalGans, natalZhis) {
		return buildTiaohouResult(natal, "hot")
	}

	// Step 2+3: 扶抑
	dayGanWx := ganWuxing[natal.DayGan]
	monthZhiWx := zhiWuxing[natal.MonthZhi]

	helpElements, opposeElements := getHelpOpposeElements(dayGanWx)
	if helpElements == "" {
		return ClassicalYongshenResult{
			Reason: fmt.Sprintf("无法识别日主五行: dayGan=%s", natal.DayGan),
		}
	}

	isStrong := calcIsStrongByMonthWeight(dayGanWx, monthZhiWx, natal.Wuxing)

	var wuxingSet, jishenSet, strategy, reason string
	if isStrong {
		wuxingSet = opposeElements
		jishenSet = helpElements
		strategy = ClassicalStrategyFuyiStrong
		reason = fmt.Sprintf("身强 (月令%s+权重计算)，扶抑取克泄耗: 用神=%s 忌神=%s",
			monthZhiWx, wuxingSet, jishenSet)
	} else {
		wuxingSet = helpElements
		jishenSet = opposeElements
		strategy = ClassicalStrategyFuyiWeak
		reason = fmt.Sprintf("身弱 (月令%s+权重计算)，扶抑取生扶: 用神=%s 忌神=%s",
			monthZhiWx, wuxingSet, jishenSet)
	}

	primaryGan := selectPrimaryGan(natal, wuxingSet)
	jishenGans := collectJishenGans(jishenSet)

	return ClassicalYongshenResult{
		Strategy:   strategy,
		IsStrong:   isStrong,
		PrimaryGan: primaryGan,
		WuxingSet:  wuxingSet,
		JishenSet:  jishenSet,
		JishenGans: jishenGans,
		Reason:     reason + fmt.Sprintf("，primary_gan=%s", primaryGan),
	}
}

// isExtremeCold 月令派至寒判定
// 标准: 月支为亥/子/丑 且 原局透干无丙丁、地支无巳午
func isExtremeCold(monthZhi string, natalGans, natalZhis []string) bool {
	if monthZhi != "亥" && monthZhi != "子" && monthZhi != "丑" {
		return false
	}
	for _, g := range natalGans {
		if g == "丙" || g == "丁" {
			return false
		}
	}
	for _, z := range natalZhis {
		if z == "巳" || z == "午" {
			return false
		}
	}
	return true
}

// isExtremeHot 月令派至热判定
// 标准: 月支为巳/午/未 且 原局透干无壬癸、地支无亥子
func isExtremeHot(monthZhi string, natalGans, natalZhis []string) bool {
	if monthZhi != "巳" && monthZhi != "午" && monthZhi != "未" {
		return false
	}
	for _, g := range natalGans {
		if g == "壬" || g == "癸" {
			return false
		}
	}
	for _, z := range natalZhis {
		if z == "亥" || z == "子" {
			return false
		}
	}
	return true
}

// buildTiaohouResult 调候为急路径: 至寒用火, 至热用水
func buildTiaohouResult(natal *BaziResult, kind string) ClassicalYongshenResult {
	var wuxingSet, jishenSet, strategy, label string
	switch kind {
	case "cold":
		wuxingSet, jishenSet, strategy, label = "火", "水", ClassicalStrategyTiaohouCold, "寒"
	case "hot":
		wuxingSet, jishenSet, strategy, label = "水", "火", ClassicalStrategyTiaohouHot, "热"
	default:
		return ClassicalYongshenResult{}
	}
	primaryGan := selectPrimaryGan(natal, wuxingSet)
	return ClassicalYongshenResult{
		Strategy:   strategy,
		PrimaryGan: primaryGan,
		WuxingSet:  wuxingSet,
		JishenSet:  jishenSet,
		JishenGans: collectJishenGans(jishenSet),
		Reason:     fmt.Sprintf("至%s, 调候为急, 用神=%s, primary_gan=%s", label, wuxingSet, primaryGan),
	}
}

// getHelpOpposeElements 返回日主同党+生身 vs 克泄耗五行集字符串
// 复用 yongshen.go::fuyiElementMap
func getHelpOpposeElements(dayGanWx string) (help, oppose string) {
	elems, ok := fuyiElementMap[dayGanWx]
	if !ok {
		return "", ""
	}
	return elems[0], elems[1]
}

// calcIsStrongByMonthWeight 月支五行 +2 加权, helpPct > 40% = 身强
// 复用 calcWeightedYongshen 同一阈值，但只返回 bool（不返回用神五行集）
func calcIsStrongByMonthWeight(dayGanWx, monthZhiWx string, stats WuxingStats) bool {
	ws := stats
	ws.Total = 10
	switch monthZhiWx {
	case "mu":
		ws.Mu += 2
	case "huo":
		ws.Huo += 2
	case "tu":
		ws.Tu += 2
	case "jin":
		ws.Jin += 2
	case "shui":
		ws.Shui += 2
	}
	t := float64(ws.Total)
	muPct := float64(ws.Mu) / t * 100
	huoPct := float64(ws.Huo) / t * 100
	tuPct := float64(ws.Tu) / t * 100
	jinPct := float64(ws.Jin) / t * 100
	shuiPct := float64(ws.Shui) / t * 100

	var helpPct float64
	switch dayGanWx {
	case "mu":
		helpPct = muPct + shuiPct
	case "huo":
		helpPct = huoPct + muPct
	case "tu":
		helpPct = tuPct + huoPct
	case "jin":
		helpPct = jinPct + tuPct
	case "shui":
		helpPct = shuiPct + jinPct
	}
	return helpPct > 40.0
}

// selectPrimaryGan 在用神五行集中选 primary_gan
// 优先级: 透干 (按 甲乙丙丁戊己庚辛壬癸 顺序) > 藏干 > 候选首位
func selectPrimaryGan(natal *BaziResult, wuxingSet string) string {
	if wuxingSet == "" || natal == nil {
		return ""
	}
	cn2pin := map[string]string{"木": "mu", "火": "huo", "土": "tu", "金": "jin", "水": "shui"}
	targetPins := map[string]bool{}
	for _, ch := range wuxingSet {
		if p, ok := cn2pin[string(ch)]; ok {
			targetPins[p] = true
		}
	}

	allGans := []string{"甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"}
	var candidates []string
	for _, g := range allGans {
		if targetPins[ganWuxing[g]] {
			candidates = append(candidates, g)
		}
	}
	if len(candidates) == 0 {
		return ""
	}

	// 优先: 透干 (年/月/日/时 任一)
	natalGans := []string{natal.YearGan, natal.MonthGan, natal.DayGan, natal.HourGan}
	for _, g := range candidates {
		for _, ng := range natalGans {
			if g == ng {
				return g
			}
		}
	}
	// 次优: 藏干
	allHide := [][]string{natal.YearHideGan, natal.MonthHideGan, natal.DayHideGan, natal.HourHideGan}
	for _, g := range candidates {
		for _, group := range allHide {
			for _, hg := range group {
				if g == hg {
					return g
				}
			}
		}
	}
	// 末选: 候选首位
	return candidates[0]
}

// collectJishenGans 由忌神五行集生成忌神天干列表 (甲乙丙...癸 顺序)
func collectJishenGans(jishenSet string) []string {
	cn2pin := map[string]string{"木": "mu", "火": "huo", "土": "tu", "金": "jin", "水": "shui"}
	targetPins := map[string]bool{}
	for _, ch := range jishenSet {
		if p, ok := cn2pin[string(ch)]; ok {
			targetPins[p] = true
		}
	}
	allGans := []string{"甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"}
	var out []string
	for _, g := range allGans {
		if targetPins[ganWuxing[g]] {
			out = append(out, g)
		}
	}
	return out
}
