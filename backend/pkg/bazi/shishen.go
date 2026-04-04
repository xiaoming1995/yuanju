package bazi

// GetShiShen 计算天干对天干的十神
func GetShiShen(dayGan, targetGan string) string {
	// util.LunarUtil 提供了十神的换算表
	// 为了确保安全，我们先检查 util 里有没有，如果没有我们也可以手动补
	// 其实 lunar-go 有一个静态变量 util.LunarUtil.SHI_SHEN_GAN[dayGan+targetGan]
	// 但这是私有的或不存在，我们可以本地实现一个硬编码映射以保证速度和稳定性。

	wuXingMap := map[string]string{
		"甲": "mu", "乙": "mu",
		"丙": "huo", "丁": "huo",
		"戊": "tu", "己": "tu",
		"庚": "jin", "辛": "jin",
		"壬": "shui", "癸": "shui",
	}
	yinYangMap := map[string]int{
		"甲": 1, "乙": 0,
		"丙": 1, "丁": 0,
		"戊": 1, "己": 0,
		"庚": 1, "辛": 0,
		"壬": 1, "癸": 0,
	}

	dayWx := wuXingMap[dayGan]
	targetWx := wuXingMap[targetGan]
	dayYy := yinYangMap[dayGan]
	targetYy := yinYangMap[targetGan]

	sameYy := dayYy == targetYy

	if dayWx == targetWx {
		if sameYy { return "比肩" } else { return "劫财" }
	}

	// 泄（我生）
	if (dayWx == "mu" && targetWx == "huo") ||
		(dayWx == "huo" && targetWx == "tu") ||
		(dayWx == "tu" && targetWx == "jin") ||
		(dayWx == "jin" && targetWx == "shui") ||
		(dayWx == "shui" && targetWx == "mu") {
		if sameYy { return "食神" } else { return "伤官" }
	}

	// 耗（我克）
	if (dayWx == "mu" && targetWx == "tu") ||
		(dayWx == "huo" && targetWx == "jin") ||
		(dayWx == "tu" && targetWx == "shui") ||
		(dayWx == "jin" && targetWx == "mu") ||
		(dayWx == "shui" && targetWx == "huo") {
		if sameYy { return "偏财" } else { return "正财" }
	}

	// 杀（克我）
	if (targetWx == "mu" && dayWx == "tu") ||
		(targetWx == "huo" && dayWx == "jin") ||
		(targetWx == "tu" && dayWx == "shui") ||
		(targetWx == "jin" && dayWx == "mu") ||
		(targetWx == "shui" && dayWx == "huo") {
		if sameYy { return "七杀" } else { return "正官" }
	}

	// 生（生我）
	if (targetWx == "mu" && dayWx == "huo") ||
		(targetWx == "huo" && dayWx == "tu") ||
		(targetWx == "tu" && dayWx == "jin") ||
		(targetWx == "jin" && dayWx == "shui") ||
		(targetWx == "shui" && dayWx == "mu") {
		if sameYy { return "偏印" } else { return "正印" }
	}

	return ""
}

// GetZhiShiShen 取地支主气对日主的十神
func GetZhiShiShen(dayGan, zhi string) string {
	zhiMainGan := map[string]string{
		"子": "癸", "丑": "己", "寅": "甲", "卯": "乙",
		"辰": "戊", "巳": "丙", "午": "丁", "未": "己",
		"申": "庚", "酉": "辛", "戌": "戊", "亥": "壬",
	}
	mainGan, ok := zhiMainGan[zhi]
	if !ok {
		return ""
	}
	return GetShiShen(dayGan, mainGan)
}

// GetDiShi 获取日干在某地支的十二长生（地势）状态
func GetDiShi(dayGan, zhi string) string {
	dishiArr := []string{"长生", "沐浴", "冠带", "临官", "帝旺", "衰", "病", "死", "墓", "绝", "胎", "养"}
	zhiArr := []string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}

	// 阳干：甲丙戊庚壬 顺推
	// 阴干：乙丁己辛癸 逆推
	startIdx := map[string]int{
		"甲": 11, // 亥长生
		"丙": 2,  // 寅长生
		"戊": 2,  // 寅长生
		"庚": 5,  // 巳长生
		"壬": 8,  // 申长生
		"乙": 6,  // 午长生
		"丁": 9,  // 酉长生
		"己": 9,  // 酉长生
		"辛": 0,  // 子长生
		"癸": 3,  // 卯长生
	}

	isYang := map[string]bool{
		"甲": true, "丙": true, "戊": true, "庚": true, "壬": true,
		"乙": false, "丁": false, "己": false, "辛": false, "癸": false,
	}

	start, ok1 := startIdx[dayGan]
	yang, ok2 := isYang[dayGan]
	if !ok1 || !ok2 {
		return ""
	}

	targetIdx := -1
	for i, z := range zhiArr {
		if z == zhi {
			targetIdx = i
			break
		}
	}
	if targetIdx == -1 {
		return ""
	}

	offset := 0
	if yang {
		offset = (targetIdx - start + 12) % 12
	} else {
		offset = (start - targetIdx + 12) % 12
	}

	return dishiArr[offset]
}
