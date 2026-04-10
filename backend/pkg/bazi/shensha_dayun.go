package bazi

import "strings"

// GetDayunShenSha 计算大运干支所对应的神煞
// 推算基准与原局相同（年干/年支/月支/日干/日支不变），检查目标为大运的天干和地支。
// 仅日柱专属神煞（阴差阳错/日德/魁罡/十恶大败/飞刃/十灵日）不适用于大运。
func GetDayunShenSha(yearGan, yearZhi, monthZhi, dayGan, dayZhi, dayunGan, dayunZhi string) []string {
	var result []string
	add := func(cond bool, name string) {
		if cond {
			result = append(result, name)
		}
	}

	dg := dayGan
	yg := yearGan
	yz := yearZhi
	mz := monthZhi
	dz := dayZhi
	g := dayunGan // 大运天干
	z := dayunZhi // 大运地支

	// ── 第一组：日干/年干 → 大运地支 ──

	// 太极贵人（年干+日干双基准）
	add(
		(strings.Contains("甲乙", yg) && strings.Contains("子午", z)) ||
			(strings.Contains("丙丁", yg) && strings.Contains("卯酉", z)) ||
			(strings.Contains("戊己", yg) && strings.Contains("辰戌丑未", z)) ||
			(strings.Contains("庚辛", yg) && strings.Contains("寅亥", z)) ||
			(strings.Contains("壬癸", yg) && strings.Contains("巳申", z)) ||
			(strings.Contains("甲乙", dg) && strings.Contains("子午", z)) ||
			(strings.Contains("丙丁", dg) && strings.Contains("卯酉", z)) ||
			(strings.Contains("戊己", dg) && strings.Contains("辰戌丑未", z)) ||
			(strings.Contains("庚辛", dg) && strings.Contains("寅亥", z)) ||
			(strings.Contains("壬癸", dg) && strings.Contains("巳申", z)),
		"太极贵人")

	// 天乙贵人（日干+年干双基准）
	tianYi := func(stem, branch string) bool {
		return (strings.Contains("甲戊庚", stem) && strings.Contains("丑未", branch)) ||
			(strings.Contains("乙己", stem) && strings.Contains("子申", branch)) ||
			(strings.Contains("丙丁", stem) && strings.Contains("亥酉", branch)) ||
			(strings.Contains("壬癸", stem) && strings.Contains("卯巳", branch)) ||
			(stem == "辛" && strings.Contains("午寅", branch))
	}
	add(tianYi(dg, z) || tianYi(yg, z), "天乙贵人")

	// 文昌贵人（日干+年干双基准）
	wenChang := func(stem, branch string) bool {
		return (stem == "甲" && branch == "巳") || (stem == "乙" && branch == "午") ||
			(strings.Contains("丙戊", stem) && branch == "申") || (strings.Contains("丁己", stem) && branch == "酉") ||
			(stem == "庚" && branch == "亥") || (stem == "辛" && branch == "子") ||
			(stem == "壬" && branch == "寅") || (stem == "癸" && branch == "卯")
	}
	add(wenChang(dg, z) || wenChang(yg, z), "文昌贵人")

	// 禄神（日干基准）
	add(
		(dg == "甲" && z == "寅") || (dg == "乙" && z == "卯") ||
			(strings.Contains("丙戊", dg) && z == "巳") || (strings.Contains("丁己", dg) && z == "午") ||
			(dg == "庚" && z == "申") || (dg == "辛" && z == "酉") ||
			(dg == "壬" && z == "亥") || (dg == "癸" && z == "子"),
		"禄神")

	// 羊刃（日干基准）
	add(
		(dg == "甲" && z == "卯") || (dg == "乙" && z == "寅") ||
			(strings.Contains("丙戊", dg) && z == "午") || (strings.Contains("丁己", dg) && z == "未") ||
			(dg == "庚" && z == "酉") || (dg == "辛" && z == "戌") ||
			(dg == "壬" && z == "子") || (dg == "癸" && z == "丑"),
		"羊刃")

	// 金舆贵人（日干基准）
	add(
		(dg == "甲" && z == "辰") || (dg == "乙" && z == "巳") ||
			(dg == "丙" && z == "未") || (dg == "丁" && z == "申") ||
			(dg == "戊" && z == "未") || (dg == "己" && z == "申") ||
			(dg == "庚" && z == "戌") || (dg == "辛" && z == "亥") ||
			(dg == "壬" && z == "丑") || (dg == "癸" && z == "寅"),
		"金舆贵人")

	// 红艳（日干基准）
	add(
		(strings.Contains("甲乙", dg) && z == "午") ||
			(dg == "丙" && z == "寅") || (dg == "丁" && z == "未") ||
			(strings.Contains("戊己", dg) && z == "辰") ||
			(dg == "庚" && z == "戌") || (dg == "辛" && z == "酉") ||
			(dg == "壬" && z == "子") || (dg == "癸" && z == "申"),
		"红艳")

	// 天厨贵人（日干+年干双基准）
	tianChu := func(stem, branch string) bool {
		return (strings.Contains("甲丙", stem) && branch == "巳") ||
			(strings.Contains("乙丁", stem) && branch == "午") ||
			(stem == "戊" && branch == "申") || (stem == "己" && branch == "酉") ||
			(stem == "庚" && branch == "亥") || (stem == "辛" && branch == "子") ||
			(stem == "壬" && branch == "寅") || (stem == "癸" && branch == "卯")
	}
	add(tianChu(dg, z) || tianChu(yg, z), "天厨贵人")

	// 国印贵人（日干查大运地支 + 大运天干自查）
	guoYinMap := map[string]string{
		"甲": "戌", "乙": "亥", "丙": "丑", "丁": "寅",
		"戊": "丑", "己": "寅", "庚": "辰", "辛": "巳",
		"壬": "未", "癸": "申",
	}
	add(z == guoYinMap[dg] || z == guoYinMap[g], "国印贵人")

	// 词馆（日干基准）
	ciguanMap := map[string][]string{
		"甲": {"寅", "戌"}, "乙": {"卯", "亥"},
		"丙": {"巳", "丑"}, "丁": {"巳", "丑"},
		"戊": {"午", "辰"}, "己": {"午", "辰"},
		"庚": {"申", "子"}, "辛": {"申", "子"},
		"壬": {"酉", "未"}, "癸": {"酉", "未"},
	}
	if cgZhis, ok := ciguanMap[dg]; ok {
		for _, cgz := range cgZhis {
			if z == cgz {
				add(true, "词馆")
				break
			}
		}
	}

	// 墓门（大运天干克大运地支）
	ganWx := ""
	switch {
	case strings.Contains("甲乙", g):
		ganWx = "木"
	case strings.Contains("丙丁", g):
		ganWx = "火"
	case strings.Contains("戊己", g):
		ganWx = "土"
	case strings.Contains("庚辛", g):
		ganWx = "金"
	case strings.Contains("壬癸", g):
		ganWx = "水"
	}
	zhiWx := ""
	switch {
	case strings.Contains("寅卯", z):
		zhiWx = "木"
	case strings.Contains("巳午", z):
		zhiWx = "火"
	case strings.Contains("辰戌丑未", z):
		zhiWx = "土"
	case strings.Contains("申酉", z):
		zhiWx = "金"
	case strings.Contains("亥子", z):
		zhiWx = "水"
	}
	add(
		(ganWx == "金" && zhiWx == "木") || (ganWx == "木" && zhiWx == "土") ||
			(ganWx == "土" && zhiWx == "水") || (ganWx == "水" && zhiWx == "火") ||
			(ganWx == "火" && zhiWx == "金"),
		"墓门")

	// ── 第二组：月支 → 大运天干（天德/月德系列）──

	tiandeDryGan := map[string]string{
		"寅": "丁", "辰": "壬", "巳": "辛",
		"未": "甲", "申": "癸", "戌": "丙",
		"亥": "乙", "子": "庚", "丑": "己",
	}
	tiandeDryZhi := map[string]string{"卯": "申", "午": "亥", "酉": "寅"}
	tiandeheDryGan := map[string]string{
		"寅": "壬", "辰": "丁", "巳": "丙",
		"未": "己", "申": "戊", "戌": "辛",
		"亥": "庚", "子": "乙", "丑": "甲",
	}
	tiandeheDryZhi := map[string]string{"卯": "巳", "午": "寅", "酉": "亥"}

	yuedeTiangan := func(monthZhi string) string {
		switch {
		case strings.Contains("寅午戌", monthZhi):
			return "丙"
		case strings.Contains("申子辰", monthZhi):
			return "壬"
		case strings.Contains("亥卯未", monthZhi):
			return "甲"
		case strings.Contains("巳酉丑", monthZhi):
			return "庚"
		}
		return ""
	}
	yuedeheMap := map[string]string{"丙": "辛", "壬": "丁", "甲": "己", "庚": "乙"}

	hasTiande := false
	if tg := tiandeDryGan[mz]; tg != "" && g == tg {
		add(true, "天德贵人")
		hasTiande = true
	}
	if tz := tiandeDryZhi[mz]; tz != "" && z == tz {
		add(true, "天德贵人")
		hasTiande = true
	}
	if thg := tiandeheDryGan[mz]; thg != "" && g == thg {
		add(true, "天德合")
	}
	if thz := tiandeheDryZhi[mz]; thz != "" && z == thz {
		add(true, "天德合")
	}

	hasYuede := false
	if ytg := yuedeTiangan(mz); ytg != "" && g == ytg {
		add(true, "月德贵人")
		hasYuede = true
	}
	if yhtg := yuedeheMap[yuedeTiangan(mz)]; yhtg != "" && g == yhtg {
		add(true, "月德合")
	}

	// 德秀贵人
	if hasTiande || hasYuede {
		add(true, "德秀贵人")
	}

	// ── 第三组：年支/日支 → 大运地支 ──

	// 桃花
	isTaohua := func(base, check string) bool {
		return (strings.Contains("申子辰", base) && check == "酉") ||
			(strings.Contains("寅午戌", base) && check == "卯") ||
			(strings.Contains("亥卯未", base) && check == "子") ||
			(strings.Contains("巳酉丑", base) && check == "午")
	}
	add(isTaohua(yz, z) || isTaohua(dz, z), "桃花")

	// 驿马
	isYima := func(base, check string) bool {
		return (strings.Contains("申子辰", base) && check == "寅") ||
			(strings.Contains("寅午戌", base) && check == "申") ||
			(strings.Contains("亥卯未", base) && check == "巳") ||
			(strings.Contains("巳酉丑", base) && check == "亥")
	}
	add(isYima(yz, z) || isYima(dz, z), "驿马")

	// 华盖
	isHuagai := func(base, check string) bool {
		return (strings.Contains("申子辰", base) && check == "辰") ||
			(strings.Contains("寅午戌", base) && check == "戌") ||
			(strings.Contains("亥卯未", base) && check == "未") ||
			(strings.Contains("巳酉丑", base) && check == "丑")
	}
	add(isHuagai(yz, z) || isHuagai(dz, z), "华盖")

	// 将星
	isJiangxing := func(base, check string) bool {
		return (strings.Contains("申子辰", base) && check == "子") ||
			(strings.Contains("寅午戌", base) && check == "午") ||
			(strings.Contains("亥卯未", base) && check == "卯") ||
			(strings.Contains("巳酉丑", base) && check == "酉")
	}
	add(isJiangxing(yz, z) || isJiangxing(dz, z), "将星")

	// 劫煞
	isJiesha := func(base, check string) bool {
		return (strings.Contains("申子辰", base) && check == "巳") ||
			(strings.Contains("寅午戌", base) && check == "亥") ||
			(strings.Contains("亥卯未", base) && check == "申") ||
			(strings.Contains("巳酉丑", base) && check == "寅")
	}
	add(isJiesha(yz, z) || isJiesha(dz, z), "劫煞")

	// 亡神
	isWangshen := func(base, check string) bool {
		return (strings.Contains("申子辰", base) && check == "亥") ||
			(strings.Contains("寅午戌", base) && check == "巳") ||
			(strings.Contains("亥卯未", base) && check == "寅") ||
			(strings.Contains("巳酉丑", base) && check == "申")
	}
	add(isWangshen(yz, z) || isWangshen(dz, z), "亡神")

	// ── 第四组：年支 → 大运地支 ──

	// 孤辰
	add(
		(strings.Contains("亥子丑", yz) && z == "寅") ||
			(strings.Contains("寅卯辰", yz) && z == "巳") ||
			(strings.Contains("巳午未", yz) && z == "申") ||
			(strings.Contains("申酉戌", yz) && z == "亥"),
		"孤辰")

	// 寡宿
	add(
		(strings.Contains("亥子丑", yz) && z == "戌") ||
			(strings.Contains("寅卯辰", yz) && z == "丑") ||
			(strings.Contains("巳午未", yz) && z == "辰") ||
			(strings.Contains("申酉戌", yz) && z == "未"),
		"寡宿")

	// 天喜
	tianxiMap := map[string]string{
		"子": "酉", "丑": "申", "寅": "未", "卯": "午",
		"辰": "巳", "巳": "辰", "午": "卯", "未": "寅",
		"申": "丑", "酉": "子", "戌": "亥", "亥": "戌",
	}
	if txz, ok := tianxiMap[yz]; ok {
		add(z == txz, "天喜")
	}

	// 灾煞
	zaishaTianzhi := func(base string) string {
		switch {
		case strings.Contains("申子辰", base):
			return "午"
		case strings.Contains("寅午戌", base):
			return "子"
		case strings.Contains("亥卯未", base):
			return "酉"
		case strings.Contains("巳酉丑", base):
			return "卯"
		}
		return ""
	}
	if zaizhi := zaishaTianzhi(yz); zaizhi != "" {
		add(z == zaizhi, "灾煞")
	}

	// 流霞（年干 → 大运地支）
	liuxiaMap := map[string]string{
		"甲": "申", "乙": "酉", "丙": "戌", "丁": "亥",
		"戊": "子", "己": "丑", "庚": "寅", "辛": "卯",
		"壬": "辰", "癸": "巳",
	}
	if lxz, ok := liuxiaMap[yg]; ok {
		add(z == lxz, "流霞")
	}

	// 吊客（年支 → 大运地支）
	diaokeMap := map[string]string{
		"子": "戌", "丑": "亥", "寅": "子", "卯": "丑",
		"辰": "寅", "巳": "卯", "午": "辰", "未": "巳",
		"申": "午", "酉": "未", "戌": "申", "亥": "酉",
	}
	if dkz, ok := diaokeMap[yz]; ok {
		add(z == dkz, "吊客")
	}

	// 福星贵人（年干查大运地支 + 大运天干自查）
	yearFuxingMap := map[string][]string{
		"甲": {"寅"}, "乙": {"丑", "子"}, "丙": {"子"},
		"丁": {"酉"}, "戊": {"申"}, "己": {"未", "午"},
		"庚": {"午"}, "辛": {"辰"}, "壬": {"卯", "寅"}, "癸": {"丑"},
	}
	ownFuxingMap := map[string][]string{
		"甲": {"寅"}, "乙": {"子", "丑"}, "丙": {"子", "亥"},
		"丁": {"酉"}, "戊": {"申"}, "己": {"未", "午"},
		"庚": {"午", "寅"}, "辛": {"辰"}, "壬": {"卯", "寅"}, "癸": {"丑"},
	}
	fuxingHit := false
	if fxZhis, ok := yearFuxingMap[yg]; ok {
		for _, fxz := range fxZhis {
			if z == fxz {
				add(true, "福星贵人")
				fuxingHit = true
				break
			}
		}
	}
	if !fuxingHit {
		if fxZhis, ok := ownFuxingMap[g]; ok {
			for _, fxz := range fxZhis {
				if z == fxz {
					add(true, "福星贵人")
					break
				}
			}
		}
	}

	// 天医（月支前一辰 → 大运地支）
	zhiOrder := []string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}
	mzIdx := -1
	for idx, v := range zhiOrder {
		if v == mz {
			mzIdx = idx
			break
		}
	}
	if mzIdx >= 0 {
		tianYiZhi := zhiOrder[(mzIdx+11)%12]
		add(z == tianYiZhi, "天医")
	}

	return result
}
