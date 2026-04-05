package bazi

import "strings"

// GetPillarsShenSha 计算四柱中的基础核心神煞
// 输入 parameters:
// yearG, yearZ, monthG, monthZ, dayG, dayZ, hourG, hourZ
// 返回值：每柱对应的神煞数组
func GetPillarsShenSha(yg, yz, mg, mz, dg, dz, hg, hz string) [4][]string {
	var result [4][]string
	for i := range result {
		result[i] = make([]string, 0)
	}

	zhis := [4]string{yz, mz, dz, hz}
	ganZhis := [4]string{yg + yz, mg + mz, dg + dz, hg + hz}

	addIf := func(idx int, condition bool, name string) {
		if condition {
			result[idx] = append(result[idx], name)
		}
	}

	// 1. 基于日干查地支
	for i, z := range zhis {
		// 天乙贵人 (甲戊庚牛羊, 乙己鼠猴乡, 丙丁猪鸡位, 壬癸兔蛇藏, 六辛逢马虎)
		addIf(i,
			(strings.Contains("甲戊庚", dg) && strings.Contains("丑未", z)) ||
				(strings.Contains("乙己", dg) && strings.Contains("子申", z)) ||
				(strings.Contains("丙丁", dg) && strings.Contains("亥酉", z)) ||
				(strings.Contains("壬癸", dg) && strings.Contains("卯巳", z)) ||
				(dg == "辛" && strings.Contains("午寅", z)),
			"天乙贵人")

		// 太极贵人 (甲乙见子午, 丙丁见卯酉, 戊己见辰戌丑未, 庚辛见寅亥, 壬癸见巳申)
		addIf(i,
			(strings.Contains("甲乙", dg) && strings.Contains("子午", z)) ||
				(strings.Contains("丙丁", dg) && strings.Contains("卯酉", z)) ||
				(strings.Contains("戊己", dg) && strings.Contains("辰戌丑未", z)) ||
				(strings.Contains("庚辛", dg) && strings.Contains("寅亥", z)) ||
				(strings.Contains("壬癸", dg) && strings.Contains("巳申", z)),
			"太极贵人")

		// 文昌贵人 (甲巳乙午丙戊申, 丁己酉庚亥辛子, 壬寅癸卯)
		addIf(i,
			(dg == "甲" && z == "巳") || (dg == "乙" && z == "午") ||
				(strings.Contains("丙戊", dg) && z == "申") || (strings.Contains("丁己", dg) && z == "酉") ||
				(dg == "庚" && z == "亥") || (dg == "辛" && z == "子") ||
				(dg == "壬" && z == "寅") || (dg == "癸" && z == "卯"),
			"文昌贵人")

		// 禄神 (甲寅乙卯丙戊巳...)
		addIf(i,
			(dg == "甲" && z == "寅") || (dg == "乙" && z == "卯") ||
				(strings.Contains("丙戊", dg) && z == "巳") || (strings.Contains("丁己", dg) && z == "午") ||
				(dg == "庚" && z == "申") || (dg == "辛" && z == "酉") ||
				(dg == "壬" && z == "亥") || (dg == "癸" && z == "子"),
			"禄神")

		// 羊刃 (禄前一位)
		addIf(i,
			(dg == "甲" && z == "卯") || (dg == "乙" && z == "辰") ||
				(strings.Contains("丙戊", dg) && z == "午") || (strings.Contains("丁己", dg) && z == "未") ||
				(dg == "庚" && z == "酉") || (dg == "辛" && z == "戌") ||
				(dg == "壬" && z == "子") || (dg == "癸" && z == "丑"),
			"羊刃")

		// 天厨贵人 (甲丁, 乙丙, 丙子...) 等规则较繁复，此处简化略过
	}

	// 2. 基于日支或年支查地支 (以日支为主，年支为辅，这里如果两者其一匹配就算)
	for i, z := range zhis {
		isTaohua := func(baseZhi, checkZhi string) bool {
			return (strings.Contains("申子辰", baseZhi) && checkZhi == "酉") ||
				(strings.Contains("寅午戌", baseZhi) && checkZhi == "卯") ||
				(strings.Contains("亥卯未", baseZhi) && checkZhi == "子") ||
				(strings.Contains("巳酉丑", baseZhi) && checkZhi == "午")
		}
		addIf(i, isTaohua(dz, z) || isTaohua(yz, z), "桃花")

		isYima := func(baseZhi, checkZhi string) bool {
			return (strings.Contains("申子辰", baseZhi) && checkZhi == "寅") ||
				(strings.Contains("寅午戌", baseZhi) && checkZhi == "申") ||
				(strings.Contains("亥卯未", baseZhi) && checkZhi == "巳") ||
				(strings.Contains("巳酉丑", baseZhi) && checkZhi == "亥")
		}
		addIf(i, isYima(dz, z) || isYima(yz, z), "驿马")

		isHuagai := func(baseZhi, checkZhi string) bool {
			return (strings.Contains("申子辰", baseZhi) && checkZhi == "辰") ||
				(strings.Contains("寅午戌", baseZhi) && checkZhi == "戌") ||
				(strings.Contains("亥卯未", baseZhi) && checkZhi == "未") ||
				(strings.Contains("巳酉丑", baseZhi) && checkZhi == "丑")
		}
		addIf(i, isHuagai(dz, z) || isHuagai(yz, z), "华盖")

		isJiangxing := func(baseZhi, checkZhi string) bool {
			return (strings.Contains("申子辰", baseZhi) && checkZhi == "子") ||
				(strings.Contains("寅午戌", baseZhi) && checkZhi == "午") ||
				(strings.Contains("亥卯未", baseZhi) && checkZhi == "卯") ||
				(strings.Contains("巳酉丑", baseZhi) && checkZhi == "酉")
		}
		addIf(i, isJiangxing(dz, z) || isJiangxing(yz, z), "将星")

		isJiesha := func(baseZhi, checkZhi string) bool {
			return (strings.Contains("申子辰", baseZhi) && checkZhi == "巳") ||
				(strings.Contains("寅午戌", baseZhi) && checkZhi == "亥") ||
				(strings.Contains("亥卯未", baseZhi) && checkZhi == "申") ||
				(strings.Contains("巳酉丑", baseZhi) && checkZhi == "寅")
		}
		addIf(i, isJiesha(dz, z) || isJiesha(yz, z), "劫煞")

		isWangshen := func(baseZhi, checkZhi string) bool {
			return (strings.Contains("申子辰", baseZhi) && checkZhi == "亥") ||
				(strings.Contains("寅午戌", baseZhi) && checkZhi == "巳") ||
				(strings.Contains("亥卯未", baseZhi) && checkZhi == "寅") ||
				(strings.Contains("巳酉丑", baseZhi) && checkZhi == "申")
		}
		addIf(i, isWangshen(dz, z) || isWangshen(yz, z), "亡神")
	}

	// 3. 孤辰寡宿 (基于年支查)
	for i, z := range zhis {
		isGuchen := func(baseZhi, checkZhi string) bool {
			return (strings.Contains("亥子丑", baseZhi) && checkZhi == "寅") ||
				(strings.Contains("寅卯辰", baseZhi) && checkZhi == "巳") ||
				(strings.Contains("巳午未", baseZhi) && checkZhi == "申") ||
				(strings.Contains("申酉戌", baseZhi) && checkZhi == "亥")
		}
		addIf(i, isGuchen(yz, z), "孤辰")

		isGuasu := func(baseZhi, checkZhi string) bool {
			return (strings.Contains("亥子丑", baseZhi) && checkZhi == "戌") ||
				(strings.Contains("寅卯辰", baseZhi) && checkZhi == "丑") ||
				(strings.Contains("巳午未", baseZhi) && checkZhi == "辰") ||
				(strings.Contains("申酉戌", baseZhi) && checkZhi == "未")
		}
		addIf(i, isGuasu(yz, z), "寡宿")
	}

	// 4. 干支自柱神煞 (如阴差阳错、天罗地网等)
	for i, gz := range ganZhis {
		addIf(i, strings.Contains("丙子,丁丑,戊寅,辛卯,壬辰,癸巳,丙午,丁未,戊申,辛酉,壬戌,癸亥", gz), "阴差阳错")
		addIf(i, strings.Contains("甲辰,戊戌", gz), "日德")
		addIf(i, strings.Contains("庚辰,庚戌,壬辰,壬戌,戊戌", gz), "魁罡")
		addIf(i, strings.Contains("甲辰,乙巳,丙申,丁亥,戊戌,己丑,庚辰,辛巳,壬申,癸亥", gz), "十恶大败")

		// 国印贵人 (甲见戌, 乙见亥, 丙见丑, 丁见寅, 戊见丑, 己见寅, 庚见辰, 辛见巳, 壬见未, 癸见申)  (以日干查)
		z := zhis[i]
		addIf(i,
			(dg == "甲" && z == "戌") || (dg == "乙" && z == "亥") ||
				(dg == "丙" && z == "丑") || (dg == "丁" && z == "寅") ||
				(dg == "戊" && z == "丑") || (dg == "己" && z == "寅") ||
				(dg == "庚" && z == "辰") || (dg == "辛" && z == "巳") ||
				(dg == "壬" && z == "未") || (dg == "癸" && z == "申"),
			"国印贵人")
	}

	// 天罗地网: 辰为天罗, 戌为地网; 戌见亥为天罗, 辰见巳为地网 (简化: 命中见辰戌巳亥多见者多视为罗网)
	// 在古典中：男怕天罗，女怕地网...
	for i, z := range zhis {
		if (yz == "戌" || yz == "亥") && ((z == "戌" || z == "亥") && z != yz) {
			addIf(i, true, "天罗")
		}
		if (yz == "辰" || yz == "巳") && ((z == "辰" || z == "巳") && z != yz) {
			addIf(i, true, "地网")
		}
	}

	return result
}
