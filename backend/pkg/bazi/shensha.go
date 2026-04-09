package bazi

import "strings"

// ShenShaPolarity 神煞极性表，供前端区分渲染
// "ji" = 吉, "xiong" = 凶, "zhong" = 中性（需结合格局判断）
var ShenShaPolarity = map[string]string{
	// ── 吉神 ──────────────────────────────────────────────────
	"天乙贵人": "ji",
	"太极贵人": "ji",
	"文昌贵人": "ji",
	"禄神":   "ji",
	"天德贵人": "ji",
	"月德贵人": "ji",
	"天德合":  "ji",
	"月德合":  "ji",
	"金舆贵人": "ji",
	"天喜星":  "ji",
	"天厨贵人": "ji",
	"国印贵人": "ji",
	"三奇贵人": "ji",
	"日德":   "ji",
	"将星":   "ji",
	"福星贵人": "ji",
	"天医":   "ji",

	// ── 凶煞 ──────────────────────────────────────────────────
	"羊刃":   "xiong",
	"飞刃":   "xiong",
	"劫煞":   "xiong",
	"亡神":   "xiong",
	"孤辰":   "xiong",
	"寡宿":   "xiong",
	"阴差阳错": "xiong",
	"魁罡":   "xiong",
	"十恶大败": "xiong",
	"天罗":   "xiong",
	"地网":   "xiong",

	// ── 中性（需结合格局判断）───────────────────────────────────
	"桃花": "zhong",
	"驿马": "zhong",
	"华盖": "zhong",
	"红艳": "zhong",
}

// GetPillarsShenSha 计算四柱神煞（全面扩充版）
//
// 参数：年干 年支 月干 月支 日干 日支 时干 时支
// 返回：[4][]string，索引 0=年 1=月 2=日 3=时
func GetPillarsShenSha(yg, yz, mg, mz, dg, dz, hg, hz string) [4][]string {
	var result [4][]string
	for i := range result {
		result[i] = make([]string, 0)
	}

	zhis := [4]string{yz, mz, dz, hz}    // 四柱地支
	gans := [4]string{yg, mg, dg, hg}    // 四柱天干
	ganZhis := [4]string{yg + yz, mg + mz, dg + dz, hg + hz} // 四柱干支组合

	addIf := func(idx int, cond bool, name string) {
		if cond {
			result[idx] = append(result[idx], name)
		}
	}

	// ══════════════════════════════════════════════════════════
	// 第一组：日干 → 四柱地支
	// ══════════════════════════════════════════════════════════
	for i, z := range zhis {

		// 天乙贵人（甲戊庚牛羊，乙己鼠猴乡，丙丁猪鸡位，壬癸兔蛇藏，六辛逢马虎）
		addIf(i,
			(strings.Contains("甲戊庚", dg) && strings.Contains("丑未", z)) ||
				(strings.Contains("乙己", dg) && strings.Contains("子申", z)) ||
				(strings.Contains("丙丁", dg) && strings.Contains("亥酉", z)) ||
				(strings.Contains("壬癸", dg) && strings.Contains("卯巳", z)) ||
				(dg == "辛" && strings.Contains("午寅", z)),
			"天乙贵人")

		// 太极贵人（甲乙见子午，丙丁见卯酉，戊己见辰戌丑未，庚辛见寅亥，壬癸见巳申）
		addIf(i,
			(strings.Contains("甲乙", dg) && strings.Contains("子午", z)) ||
				(strings.Contains("丙丁", dg) && strings.Contains("卯酉", z)) ||
				(strings.Contains("戊己", dg) && strings.Contains("辰戌丑未", z)) ||
				(strings.Contains("庚辛", dg) && strings.Contains("寅亥", z)) ||
				(strings.Contains("壬癸", dg) && strings.Contains("巳申", z)),
			"太极贵人")

		// 文昌贵人（甲巳 乙午 丙戊申 丁己酉 庚亥 辛子 壬寅 癸卯）
		addIf(i,
			(dg == "甲" && z == "巳") || (dg == "乙" && z == "午") ||
				(strings.Contains("丙戊", dg) && z == "申") || (strings.Contains("丁己", dg) && z == "酉") ||
				(dg == "庚" && z == "亥") || (dg == "辛" && z == "子") ||
				(dg == "壬" && z == "寅") || (dg == "癸" && z == "卯"),
			"文昌贵人")

		// 禄神（日干临官之地）
		addIf(i,
			(dg == "甲" && z == "寅") || (dg == "乙" && z == "卯") ||
				(strings.Contains("丙戊", dg) && z == "巳") || (strings.Contains("丁己", dg) && z == "午") ||
				(dg == "庚" && z == "申") || (dg == "辛" && z == "酉") ||
				(dg == "壬" && z == "亥") || (dg == "癸" && z == "子"),
			"禄神")

		// 羊刃（禄前一位）
		addIf(i,
			(dg == "甲" && z == "卯") || (dg == "乙" && z == "辰") ||
				(strings.Contains("丙戊", dg) && z == "午") || (strings.Contains("丁己", dg) && z == "未") ||
				(dg == "庚" && z == "酉") || (dg == "辛" && z == "戌") ||
				(dg == "壬" && z == "子") || (dg == "癸" && z == "丑"),
			"羊刃")

		// 飞刃（羊刃对冲）
		addIf(i,
			(dg == "甲" && z == "酉") || (dg == "乙" && z == "戌") ||
				(strings.Contains("丙戊", dg) && z == "子") || (strings.Contains("丁己", dg) && z == "丑") ||
				(dg == "庚" && z == "卯") || (dg == "辛" && z == "辰") ||
				(dg == "壬" && z == "午") || (dg == "癸" && z == "未"),
			"飞刃")

		// 金舆贵人（日干之帝旺前一位）
		addIf(i,
			(dg == "甲" && z == "辰") || (dg == "乙" && z == "巳") ||
				(dg == "丙" && z == "未") || (dg == "丁" && z == "申") ||
				(dg == "戊" && z == "未") || (dg == "己" && z == "申") ||
				(dg == "庚" && z == "戌") || (dg == "辛" && z == "亥") ||
				(dg == "壬" && z == "丑") || (dg == "癸" && z == "寅"),
			"金舆贵人")

		// 红艳（以艳遇、情感色彩为主）
		addIf(i,
			(strings.Contains("甲乙", dg) && z == "午") ||
				(dg == "丙" && z == "寅") || (dg == "丁" && z == "未") ||
				(strings.Contains("戊己", dg) && z == "辰") ||
				(dg == "庚" && z == "戌") || (dg == "辛" && z == "酉") ||
				(dg == "壬" && z == "子") || (dg == "癸" && z == "申"),
			"红艳")

		// 天厨贵人（食神得禄，主富贵福寿）
		addIf(i,
			(dg == "甲" && z == "巳") || (dg == "乙" && z == "午") ||
				(dg == "丙" && z == "寅") || (dg == "丁" && z == "酉") ||
				(dg == "戊" && z == "申") || (dg == "己" && z == "酉") ||
				(dg == "庚" && z == "亥") || (dg == "辛" && z == "子") ||
				(dg == "壬" && z == "寅") || (dg == "癸" && z == "卯"),
			"天厨贵人")

		// 国印贵人（日干查各柱地支）
		addIf(i,
			(dg == "甲" && z == "戌") || (dg == "乙" && z == "亥") ||
				(dg == "丙" && z == "丑") || (dg == "丁" && z == "寅") ||
				(dg == "戊" && z == "丑") || (dg == "己" && z == "寅") ||
				(dg == "庚" && z == "辰") || (dg == "辛" && z == "巳") ||
				(dg == "壬" && z == "未") || (dg == "癸" && z == "申"),
			"国印贵人")
	}

	// ══════════════════════════════════════════════════════════
	// 第二组：月支 → 四柱天干（天德、月德系列）
	// ══════════════════════════════════════════════════════════

	// 天德贵人对应表（月支 → 吉神，天干型 or 地支型混合）
	// 正寅→丁 二卯→申(地支) 三辰→壬 四巳→辛 五午→亥(地支) 六未→甲
	// 七申→癸 八酉→寅(地支) 九戌→丙 十亥→乙 十一子→庚 十二丑→己
	tiandeDryGan := map[string]string{
		"寅": "丁", "辰": "壬", "巳": "辛",
		"未": "甲", "申": "癸", "戌": "丙",
		"亥": "乙", "子": "庚", "丑": "己",
	}
	tiandeDryZhi := map[string]string{
		"卯": "申", "午": "亥", "酉": "寅",
	}
	// 天德合（六合对应天干型天德）
	tiandeheDryGan := map[string]string{
		"寅": "壬", // 丁合壬
		"辰": "丁", // 壬合丁
		"巳": "丙", // 辛合丙
		"未": "己", // 甲合己
		"申": "戊", // 癸合戊
		"戌": "辛", // 丙合辛
		"亥": "庚", // 乙合庚
		"子": "乙", // 庚合乙
		"丑": "甲", // 己合甲
	}
	// 月德贵人（月支三合组 → 特定天干）
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
	// 月德合
	yuedeheMap := map[string]string{"丙": "辛", "壬": "丁", "甲": "己", "庚": "乙"}

	yuedeTg := yuedeTiangan(mz)
	yuedeheTg := yuedeheMap[yuedeTg]
	tiandeTianganVal := tiandeDryGan[mz]  // 天干型天德
	tiandeZhiVal := tiandeDryZhi[mz]      // 地支型天德
	tiandeheTianganVal := tiandeheDryGan[mz] // 天德合

	for i, g := range gans {
		// 天德贵人（天干型月支：当柱天干 == 天德干）
		if tiandeTianganVal != "" {
			addIf(i, g == tiandeTianganVal, "天德贵人")
		}
		// 天德合（天干型月支对应）
		if tiandeheTianganVal != "" {
			addIf(i, g == tiandeheTianganVal, "天德合")
		}
		// 月德贵人
		if yuedeTg != "" {
			addIf(i, g == yuedeTg, "月德贵人")
		}
		// 月德合
		if yuedeheTg != "" {
			addIf(i, g == yuedeheTg, "月德合")
		}
	}
	// 天德贵人（地支型月支：当柱地支 == 天德支）
	if tiandeZhiVal != "" {
		for i, z := range zhis {
			addIf(i, z == tiandeZhiVal, "天德贵人")
		}
	}

	// ══════════════════════════════════════════════════════════
	// 第三组：年支 OR 日支 → 四柱地支
	// ══════════════════════════════════════════════════════════
	for i, z := range zhis {
		isTaohua := func(base, check string) bool {
			return (strings.Contains("申子辰", base) && check == "酉") ||
				(strings.Contains("寅午戌", base) && check == "卯") ||
				(strings.Contains("亥卯未", base) && check == "子") ||
				(strings.Contains("巳酉丑", base) && check == "午")
		}
		addIf(i, isTaohua(dz, z) || isTaohua(yz, z), "桃花")

		isYima := func(base, check string) bool {
			return (strings.Contains("申子辰", base) && check == "寅") ||
				(strings.Contains("寅午戌", base) && check == "申") ||
				(strings.Contains("亥卯未", base) && check == "巳") ||
				(strings.Contains("巳酉丑", base) && check == "亥")
		}
		addIf(i, isYima(dz, z) || isYima(yz, z), "驿马")

		isHuagai := func(base, check string) bool {
			return (strings.Contains("申子辰", base) && check == "辰") ||
				(strings.Contains("寅午戌", base) && check == "戌") ||
				(strings.Contains("亥卯未", base) && check == "未") ||
				(strings.Contains("巳酉丑", base) && check == "丑")
		}
		addIf(i, isHuagai(dz, z) || isHuagai(yz, z), "华盖")

		isJiangxing := func(base, check string) bool {
			return (strings.Contains("申子辰", base) && check == "子") ||
				(strings.Contains("寅午戌", base) && check == "午") ||
				(strings.Contains("亥卯未", base) && check == "卯") ||
				(strings.Contains("巳酉丑", base) && check == "酉")
		}
		addIf(i, isJiangxing(dz, z) || isJiangxing(yz, z), "将星")

		isJiesha := func(base, check string) bool {
			return (strings.Contains("申子辰", base) && check == "巳") ||
				(strings.Contains("寅午戌", base) && check == "亥") ||
				(strings.Contains("亥卯未", base) && check == "申") ||
				(strings.Contains("巳酉丑", base) && check == "寅")
		}
		addIf(i, isJiesha(dz, z) || isJiesha(yz, z), "劫煞")

		isWangshen := func(base, check string) bool {
			return (strings.Contains("申子辰", base) && check == "亥") ||
				(strings.Contains("寅午戌", base) && check == "巳") ||
				(strings.Contains("亥卯未", base) && check == "寅") ||
				(strings.Contains("巳酉丑", base) && check == "申")
		}
		addIf(i, isWangshen(dz, z) || isWangshen(yz, z), "亡神")
	}

	// ══════════════════════════════════════════════════════════
	// 第四组：年支 → 四柱地支
	// ══════════════════════════════════════════════════════════
	for i, z := range zhis {
		// 孤辰（年支查）
		isGuchen := func(base, check string) bool {
			return (strings.Contains("亥子丑", base) && check == "寅") ||
				(strings.Contains("寅卯辰", base) && check == "巳") ||
				(strings.Contains("巳午未", base) && check == "申") ||
				(strings.Contains("申酉戌", base) && check == "亥")
		}
		addIf(i, isGuchen(yz, z), "孤辰")

		// 寡宿（年支查）
		isGuasu := func(base, check string) bool {
			return (strings.Contains("亥子丑", base) && check == "戌") ||
				(strings.Contains("寅卯辰", base) && check == "丑") ||
				(strings.Contains("巳午未", base) && check == "辰") ||
				(strings.Contains("申酉戌", base) && check == "未")
		}
		addIf(i, isGuasu(yz, z), "寡宿")

		// 天喜星（年支三合局 → 三合末位的冲位前一个）
		isTianxi := func(base, check string) bool {
			return (strings.Contains("申子辰", base) && check == "戌") ||
				(strings.Contains("寅午戌", base) && check == "辰") ||
				(strings.Contains("亥卯未", base) && check == "丑") ||
				(strings.Contains("巳酉丑", base) && check == "未")
		}
		addIf(i, isTianxi(yz, z), "天喜星")
	}

	// 福星贵人（年干 → 四柱地支）
	fuxingMap := map[string][]string{
		"甲": {"寅"},
		"乙": {"丑", "子"},
		"丙": {"子", "亥"},
		"丁": {"酉"},
		"戊": {"申"},
		"己": {"未", "午"},
		"庚": {"巳"},
		"辛": {"辰"},
		"壬": {"卯", "寅"},
		"癸": {"丑"},
	}
	if fxZhis, ok := fuxingMap[yg]; ok {
		for i, z := range zhis {
			for _, fxz := range fxZhis {
				addIf(i, z == fxz, "福星贵人")
			}
		}
	}

	// ══════════════════════════════════════════════════════════
	// 第五组：月支 → 四柱地支（天医）
	// ══════════════════════════════════════════════════════════
	// 天医 = 月支前一辰（地支序列往前退一位）
	zhiOrder := []string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}
	zhiIndex := func(z string) int {
		for idx, v := range zhiOrder {
			if v == z {
				return idx
			}
		}
		return -1
	}
	mzIdx := zhiIndex(mz)
	if mzIdx >= 0 {
		tianYiZhi := zhiOrder[(mzIdx+11)%12] // 前一位（+11 mod 12 等于 -1）
		for i, z := range zhis {
			addIf(i, z == tianYiZhi, "天医")
		}
	}

	// ══════════════════════════════════════════════════════════
	// 第六组：干支自柱（干支组合对）
	// ══════════════════════════════════════════════════════════
	for i, gz := range ganZhis {
		// 阴差阳错
		yinchaSet := map[string]bool{
			"丙子": true, "丁丑": true, "戊寅": true, "辛卯": true, "壬辰": true, "癸巳": true,
			"丙午": true, "丁未": true, "戊申": true, "辛酉": true, "壬戌": true, "癸亥": true,
		}
		addIf(i, yinchaSet[gz], "阴差阳错")

		// 日德（甲辰 戊戌）
		addIf(i, gz == "甲辰" || gz == "戊戌", "日德")

		// 魁罡（庚辰 庚戌 壬辰 壬戌，部分流派加戊戌）
		kuiGangSet := map[string]bool{"庚辰": true, "庚戌": true, "壬辰": true, "壬戌": true, "戊戌": true}
		addIf(i, kuiGangSet[gz], "魁罡")

		// 十恶大败
		shiESet := map[string]bool{
			"甲辰": true, "乙巳": true, "丙申": true, "丁亥": true, "戊戌": true,
			"己丑": true, "庚辰": true, "辛巳": true, "壬申": true, "癸亥": true,
		}
		addIf(i, shiESet[gz], "十恶大败")
	}

	// ══════════════════════════════════════════════════════════
	// 第七组：三奇贵人（四柱天干组合）
	// ══════════════════════════════════════════════════════════
	allGans := yg + mg + dg + hg
	hasTianSanqi := strings.Contains(allGans, "甲") && strings.Contains(allGans, "戊") && strings.Contains(allGans, "庚")
	hasDiSanqi := strings.Contains(allGans, "乙") && strings.Contains(allGans, "丙") && strings.Contains(allGans, "丁")
	hasRenSanqi := strings.Contains(allGans, "壬") && strings.Contains(allGans, "癸") && strings.Contains(allGans, "辛")
	if hasTianSanqi || hasDiSanqi || hasRenSanqi {
		for i, g := range gans {
			var partOf bool
			if hasTianSanqi {
				partOf = partOf || strings.Contains("甲戊庚", g)
			}
			if hasDiSanqi {
				partOf = partOf || strings.Contains("乙丙丁", g)
			}
			if hasRenSanqi {
				partOf = partOf || strings.Contains("壬癸辛", g)
			}
			addIf(i, partOf, "三奇贵人")
		}
	}

	// ══════════════════════════════════════════════════════════
	// 第八组：天罗地网（全盘地支扫描，修复 Bug）
	// ══════════════════════════════════════════════════════════
	// 天罗：命局中同时存在「戌」和「亥」
	// 地网：命局中同时存在「辰」和「巳」
	allZhis := yz + mz + dz + hz
	hasTianLuo := strings.Contains(allZhis, "戌") && strings.Contains(allZhis, "亥")
	hasDiWang := strings.Contains(allZhis, "辰") && strings.Contains(allZhis, "巳")
	for i, z := range zhis {
		if hasTianLuo {
			addIf(i, z == "戌" || z == "亥", "天罗")
		}
		if hasDiWang {
			addIf(i, z == "辰" || z == "巳", "地网")
		}
	}

	return result
}
