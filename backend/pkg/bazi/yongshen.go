package bazi

// YongshenStatus 取值常量
const (
	YongshenStatusTiaohouHit          = "tiaohou_hit"
	YongshenStatusTiaohouMissFallback = "tiaohou_miss_fallback_fuyi"
	YongshenStatusTiaohouDictMissing  = "tiaohou_dict_missing"
	YongshenStatusFuyi                = "fuyi"
)

// collectNatalGans 收集原局四柱透干（保持出现顺序，未去重）
func collectNatalGans(yearGan, monthGan, dayGan, hourGan string) []string {
	out := make([]string, 0, 4)
	for _, g := range []string{yearGan, monthGan, dayGan, hourGan} {
		if g != "" {
			out = append(out, g)
		}
	}
	return out
}

// collectNatalHideGans 合并四支藏干为单一切片（保持出现顺序，未去重）
func collectNatalHideGans(yearHide, monthHide, dayHide, hourHide []string) []string {
	total := len(yearHide) + len(monthHide) + len(dayHide) + len(hourHide)
	out := make([]string, 0, total)
	for _, group := range [][]string{yearHide, monthHide, dayHide, hourHide} {
		for _, g := range group {
			if g != "" {
				out = append(out, g)
			}
		}
	}
	return out
}

// intersectGans 求交集，返回命中与缺位天干列表（保持 needs 顺序，去重）
func intersectGans(needs, available []string) (hit, miss []string) {
	avail := make(map[string]bool, len(available))
	for _, g := range available {
		avail[g] = true
	}
	seen := make(map[string]bool, len(needs))
	for _, g := range needs {
		if seen[g] {
			continue
		}
		seen[g] = true
		if avail[g] {
			hit = append(hit, g)
		} else {
			miss = append(miss, g)
		}
	}
	return hit, miss
}

// gansToWuxingSet 将命中天干列表去重转为五行中文集字符串（如 ["丙","癸"] → "火水"）
// 顺序：按天干列表首次出现的五行顺序
func gansToWuxingSet(gans []string) string {
	seen := make(map[string]bool, 5)
	var out []byte
	for _, g := range gans {
		wxPin, ok := ganWuxing[g]
		if !ok {
			continue
		}
		wxCN, ok := wxPinyin2CN[wxPin]
		if !ok {
			continue
		}
		if seen[wxCN] {
			continue
		}
		seen[wxCN] = true
		out = append(out, []byte(wxCN)...)
	}
	return string(out)
}

// wuxingSetToJishen 由 yongshen 五行集派生 jishen 五行集
// 规则：jishen = ∪(克 yongshen) ∪ ∪(泄 yongshen，即 yongshen 所生)，去重并排除 yongshen 自身五行
func wuxingSetToJishen(yongshenSet string) string {
	cn2pin := map[string]string{"木": "mu", "火": "huo", "土": "tu", "金": "jin", "水": "shui"}
	yongPin := make(map[string]bool, 5)
	for _, ch := range yongshenSet {
		cn := string(ch)
		if pin, ok := cn2pin[cn]; ok {
			yongPin[pin] = true
		}
	}
	if len(yongPin) == 0 {
		return ""
	}
	jiPin := make(map[string]bool, 5)
	for yong := range yongPin {
		// 克者：找 K 使 wxKe[K] = yong
		for k, target := range wxKe {
			if target == yong {
				jiPin[k] = true
			}
		}
		// 泄者：yong 所生
		if xie, ok := wxSheng[yong]; ok {
			jiPin[xie] = true
		}
	}
	// 排除 yongshen 自身五行
	for yong := range yongPin {
		delete(jiPin, yong)
	}
	// 按固定顺序输出（木火土金水），与 yongshenSet 输出风格一致
	order := []string{"mu", "huo", "tu", "jin", "shui"}
	pin2cn := map[string]string{"mu": "木", "huo": "火", "tu": "土", "jin": "金", "shui": "水"}
	var out []byte
	for _, p := range order {
		if jiPin[p] {
			out = append(out, []byte(pin2cn[p])...)
		}
	}
	return string(out)
}

// inferYongshenWithTiaohouPriority 主入口：t0 调候字典优先，t1 月令权重扶抑 fallback
//
// 返回值：
//   - yongshen, jishen：五行中文集合字符串（兼容下游 strings.Contains 匹配）
//   - status：YongshenStatusTiaohouHit / TiaohouMissFallback / TiaohouDictMissing / Fuyi
//   - hitGans：t0 命中的具体调候用神天干列表（fallback 时为空）
//   - missGans：t0 缺位的调候用神天干列表（命中时为空）
//
// dayGan、monthZhi 用于查字典；gans/hideGans 为原局透+藏天干集合（用于命中检测）
// dayGanWx、monthZhiWx、stats 仅在 fallback 时传给 calcWeightedYongshen
func inferYongshenWithTiaohouPriority(
	dayGan, monthZhi string,
	gans, hideGans []string,
	dayGanWx, monthZhiWx string,
	stats WuxingStats,
) (yongshen, jishen, status string, hitGans, missGans []string) {
	tiaohouNeeds := LookupTiaohouYongshen(dayGan, monthZhi)
	if len(tiaohouNeeds) == 0 {
		// 字典缺失（理论 120 条全覆盖，此为防御性分支）
		ys, js := calcWeightedYongshen(dayGanWx, monthZhiWx, stats)
		return ys, js, YongshenStatusTiaohouDictMissing, nil, nil
	}

	available := append([]string{}, gans...)
	available = append(available, hideGans...)
	hit, miss := intersectGans(tiaohouNeeds, available)

	if len(hit) > 0 {
		ys := gansToWuxingSet(hit)
		js := wuxingSetToJishen(ys)
		return ys, js, YongshenStatusTiaohouHit, hit, miss
	}

	// 调候用神天干在原局完全缺位 → 回退至月令权重扶抑
	ys, js := calcWeightedYongshen(dayGanWx, monthZhiWx, stats)
	return ys, js, YongshenStatusTiaohouMissFallback, nil, miss
}
