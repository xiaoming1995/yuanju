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
	"德秀贵人": "ji",
	"金舆贵人": "ji",
	"天喜":   "ji",
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
	"灾煞":   "xiong",
	// ── 中性（需结合格局判断）───────────────────────────────────
	"桃花": "zhong",
	"驿马": "zhong",
	"华盖": "zhong",
	"红艳": "zhong",
}

// GetPillarsShenSha 计算四柱神煞（精校版 v3）
//
// 本版主要修复（对照问真八字校验）：
// 1. 太极贵人：改用年干推算（非日干）
// 2. 天厨贵人：丙→巳, 丁→午（食神得禄法，原错误为丙→寅, 丁→酉）
// 3. 国印贵人：日干查全柱 + 各柱本干自查（双重）
// 4. 德秀贵人：持有天德/月德的柱均得德秀；月柱作为来源也得德秀
// 5. 天罗地网：戌亥均标天罗，辰巳均标地网
// 6. 阴差阳错：仅查日柱
// 7. 华盖/将星等：基准柱自身不参与（避免年/日支自我命中）
// 8. 福星贵人：庚→午（庚马，原错误为庚→巳）
// 9. 灾煞：新增（申子辰→午, 寅午戌→子, 亥卯未→酉, 巳酉丑→卯）
//
// 参数：年干 年支 月干 月支 日干 日支 时干 时支
// 返回：[4][]string，索引 0=年 1=月 2=日 3=时
func GetPillarsShenSha(yg, yz, mg, mz, dg, dz, hg, hz string) [4][]string {
	var result [4][]string
	for i := range result {
		result[i] = make([]string, 0)
	}

	zhis := [4]string{yz, mz, dz, hz}   // 四柱地支
	gans := [4]string{yg, mg, dg, hg}   // 四柱天干
	ganZhis := [4]string{yg + yz, mg + mz, dg + dz, hg + hz}

	addIf := func(idx int, cond bool, name string) {
		if cond {
			result[idx] = append(result[idx], name)
		}
	}

	// ══════════════════════════════════════════════════════════
	// 第一组-A：年干 → 四柱地支（太极贵人）
	// Fix: 太极贵人以年干为基准，非日干
	// ══════════════════════════════════════════════════════════
	for i, z := range zhis {
		addIf(i,
			(strings.Contains("甲乙", yg) && strings.Contains("子午", z)) ||
				(strings.Contains("丙丁", yg) && strings.Contains("卯酉", z)) ||
				(strings.Contains("戊己", yg) && strings.Contains("辰戌丑未", z)) ||
				(strings.Contains("庚辛", yg) && strings.Contains("寅亥", z)) ||
				(strings.Contains("壬癸", yg) && strings.Contains("巳申", z)),
			"太极贵人")
	}

	// ══════════════════════════════════════════════════════════
	// 第一组-B：日干 → 四柱地支（主要贵人/凶煞）
	// ══════════════════════════════════════════════════════════

	// 国印贵人检查表（在循环外预计算）
	// Fix: 日干查全柱 + 各柱本干自查（双重），解决月柱漏标问题
	guoYinMap := map[string]string{
		"甲": "戌", "乙": "亥", "丙": "丑", "丁": "寅",
		"戊": "丑", "己": "寅", "庚": "辰", "辛": "巳",
		"壬": "未", "癸": "申",
	}
	dayGIZhi := guoYinMap[dg] // 日干对应的国印地支（空字符串则该干无国印）

	for i, z := range zhis {
		// 天乙贵人（甲戊庚牛羊，乙己鼠猴乡，丙丁猪鸡位，壬癸兔蛇藏，六辛逢马虎）
		addIf(i,
			(strings.Contains("甲戊庚", dg) && strings.Contains("丑未", z)) ||
				(strings.Contains("乙己", dg) && strings.Contains("子申", z)) ||
				(strings.Contains("丙丁", dg) && strings.Contains("亥酉", z)) ||
				(strings.Contains("壬癸", dg) && strings.Contains("卯巳", z)) ||
				(dg == "辛" && strings.Contains("午寅", z)),
			"天乙贵人")

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

		// 金舆贵人（日干帝旺前一位）
		addIf(i,
			(dg == "甲" && z == "辰") || (dg == "乙" && z == "巳") ||
				(dg == "丙" && z == "未") || (dg == "丁" && z == "申") ||
				(dg == "戊" && z == "未") || (dg == "己" && z == "申") ||
				(dg == "庚" && z == "戌") || (dg == "辛" && z == "亥") ||
				(dg == "壬" && z == "丑") || (dg == "癸" && z == "寅"),
			"金舆贵人")

		// 红艳（情感色彩）
		addIf(i,
			(strings.Contains("甲乙", dg) && z == "午") ||
				(dg == "丙" && z == "寅") || (dg == "丁" && z == "未") ||
				(strings.Contains("戊己", dg) && z == "辰") ||
				(dg == "庚" && z == "戌") || (dg == "辛" && z == "酉") ||
				(dg == "壬" && z == "子") || (dg == "癸" && z == "申"),
			"红艳")

		// 天厨贵人（食神得禄法）
		// Fix: 丙食神=戊→禄巳, 丁食神=己→禄午（原错误：丙→寅, 丁→酉）
		// 甲→巳(食神丙禄), 乙/丁→午(食神己禄), 丙→巳(食神戊禄), 戊→申, 己→酉
		// 庚→亥, 辛→子, 壬→寅, 癸→卯
		addIf(i,
			(strings.Contains("甲丙", dg) && z == "巳") ||
				(strings.Contains("乙丁", dg) && z == "午") ||
				(dg == "戊" && z == "申") || (dg == "己" && z == "酉") ||
				(dg == "庚" && z == "亥") || (dg == "辛" && z == "子") ||
				(dg == "壬" && z == "寅") || (dg == "癸" && z == "卯"),
			"天厨贵人")

		// 国印贵人（双重检查）
		// Fix: 日干查全柱（dayGIZhi）+ 各柱本干自查（ownGIZhi），取 OR
		ownGIZhi := guoYinMap[gans[i]]
		addIf(i, z == dayGIZhi || z == ownGIZhi, "国印贵人")
	}

	// ══════════════════════════════════════════════════════════
	// 第二组：月支 → 天干/地支（天德、月德、德秀系列）
	// ══════════════════════════════════════════════════════════

	// 天德贵人对应表（干型月支）
	tiandeDryGan := map[string]string{
		"寅": "丁", "辰": "壬", "巳": "辛",
		"未": "甲", "申": "癸", "戌": "丙",
		"亥": "乙", "子": "庚", "丑": "己",
	}
	// 天德贵人对应表（支型月支：卯月天德在申, 午月在亥, 酉月在寅）
	tiandeDryZhi := map[string]string{
		"卯": "申", "午": "亥", "酉": "寅",
	}
	// 天德合（天干型月支的六合干）
	tiandeheDryGan := map[string]string{
		"寅": "壬", "辰": "丁", "巳": "丙",
		"未": "己", "申": "戊", "戌": "辛",
		"亥": "庚", "子": "乙", "丑": "甲",
	}
	// 天德合（支型月支的六合支：申合巳, 亥合寅, 寅合亥）
	tiandeheDryZhi := map[string]string{
		"卯": "巳", "午": "寅", "酉": "亥",
	}
	// 月德贵人（三合火局寅午戌→丙, 水局申子辰→壬, 木局亥卯未→甲, 金局巳酉丑→庚）
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
	tiandeTg := tiandeDryGan[mz]
	tiandeZhi := tiandeDryZhi[mz]
	tiandeheTg := tiandeheDryGan[mz]
	tiandeheZ := tiandeheDryZhi[mz]

	// 追踪各柱是否持有天德/月德（用于德秀贵人）
	pillarHasTiande := [4]bool{}
	pillarHasYuede := [4]bool{}

	for i, g := range gans {
		if tiandeTg != "" && g == tiandeTg {
			addIf(i, true, "天德贵人")
			pillarHasTiande[i] = true
		}
		if tiandeheTg != "" {
			addIf(i, g == tiandeheTg, "天德合")
		}
		if yuedeTg != "" && g == yuedeTg {
			addIf(i, true, "月德贵人")
			pillarHasYuede[i] = true
		}
		if yuedeheTg != "" {
			addIf(i, g == yuedeheTg, "月德合")
		}
	}
	// 天德贵人（地支型）
	if tiandeZhi != "" {
		for i, z := range zhis {
			if z == tiandeZhi {
				addIf(i, true, "天德贵人")
				pillarHasTiande[i] = true
			}
		}
	}
	// 天德合（地支型）
	if tiandeheZ != "" {
		for i, z := range zhis {
			addIf(i, z == tiandeheZ, "天德合")
		}
	}

	// 德秀贵人
	// Fix: 任何持有天德/月德的柱均得德秀；命局同时有天德和月德时，月柱亦得德秀（作为来源）
	hasTiandeInChart, hasYuedeInChart := false, false
	for i := 0; i < 4; i++ {
		if pillarHasTiande[i] {
			hasTiandeInChart = true
		}
		if pillarHasYuede[i] {
			hasYuedeInChart = true
		}
	}
	for i := 0; i < 4; i++ {
		if pillarHasTiande[i] || pillarHasYuede[i] {
			addIf(i, true, "德秀贵人")
		}
	}
	// 月柱作为来源：命局同时存在天德和月德，月柱本身不含任何一者时，补标德秀
	if hasTiandeInChart && hasYuedeInChart && !pillarHasTiande[1] && !pillarHasYuede[1] {
		addIf(1, true, "德秀贵人")
	}

	// ══════════════════════════════════════════════════════════
	// 第三组：年支 OR 日支 → 四柱地支
	// Fix: 基准柱自身不参与（避免年/日支自我命中，如年支辰自标华盖）
	// ══════════════════════════════════════════════════════════
	for i, z := range zhis {
		// fromYz: 年支为基准时，排除年柱(i==0)自标
		// fromDz: 日支为基准时，排除日柱(i==2)自标
		fromYz := func(fn func(string, string) bool) bool { return i != 0 && fn(yz, z) }
		fromDz := func(fn func(string, string) bool) bool { return i != 2 && fn(dz, z) }

		isTaohua := func(base, check string) bool {
			return (strings.Contains("申子辰", base) && check == "酉") ||
				(strings.Contains("寅午戌", base) && check == "卯") ||
				(strings.Contains("亥卯未", base) && check == "子") ||
				(strings.Contains("巳酉丑", base) && check == "午")
		}
		addIf(i, fromYz(isTaohua) || fromDz(isTaohua), "桃花")

		isYima := func(base, check string) bool {
			return (strings.Contains("申子辰", base) && check == "寅") ||
				(strings.Contains("寅午戌", base) && check == "申") ||
				(strings.Contains("亥卯未", base) && check == "巳") ||
				(strings.Contains("巳酉丑", base) && check == "亥")
		}
		addIf(i, fromYz(isYima) || fromDz(isYima), "驿马")

		isHuagai := func(base, check string) bool {
			return (strings.Contains("申子辰", base) && check == "辰") ||
				(strings.Contains("寅午戌", base) && check == "戌") ||
				(strings.Contains("亥卯未", base) && check == "未") ||
				(strings.Contains("巳酉丑", base) && check == "丑")
		}
		addIf(i, fromYz(isHuagai) || fromDz(isHuagai), "华盖")

		isJiangxing := func(base, check string) bool {
			return (strings.Contains("申子辰", base) && check == "子") ||
				(strings.Contains("寅午戌", base) && check == "午") ||
				(strings.Contains("亥卯未", base) && check == "卯") ||
				(strings.Contains("巳酉丑", base) && check == "酉")
		}
		addIf(i, fromYz(isJiangxing) || fromDz(isJiangxing), "将星")

		isJiesha := func(base, check string) bool {
			return (strings.Contains("申子辰", base) && check == "巳") ||
				(strings.Contains("寅午戌", base) && check == "亥") ||
				(strings.Contains("亥卯未", base) && check == "申") ||
				(strings.Contains("巳酉丑", base) && check == "寅")
		}
		addIf(i, fromYz(isJiesha) || fromDz(isJiesha), "劫煞")

		isWangshen := func(base, check string) bool {
			return (strings.Contains("申子辰", base) && check == "亥") ||
				(strings.Contains("寅午戌", base) && check == "巳") ||
				(strings.Contains("亥卯未", base) && check == "寅") ||
				(strings.Contains("巳酉丑", base) && check == "申")
		}
		addIf(i, fromYz(isWangshen) || fromDz(isWangshen), "亡神")
	}

	// ══════════════════════════════════════════════════════════
	// 第四组：年支 → 四柱地支（孤辰、寡宿、天喜、灾煞、福星贵人）
	// ══════════════════════════════════════════════════════════
	for i, z := range zhis {
		// 孤辰（亥子丑→寅, 寅卯辰→巳, 巳午未→申, 申酉戌→亥）
		isGuchen := func(base, check string) bool {
			return (strings.Contains("亥子丑", base) && check == "寅") ||
				(strings.Contains("寅卯辰", base) && check == "巳") ||
				(strings.Contains("巳午未", base) && check == "申") ||
				(strings.Contains("申酉戌", base) && check == "亥")
		}
		addIf(i, isGuchen(yz, z), "孤辰")

		// 寡宿（亥子丑→戌, 寅卯辰→丑, 巳午未→辰, 申酉戌→未）
		isGuasu := func(base, check string) bool {
			return (strings.Contains("亥子丑", base) && check == "戌") ||
				(strings.Contains("寅卯辰", base) && check == "丑") ||
				(strings.Contains("巳午未", base) && check == "辰") ||
				(strings.Contains("申酉戌", base) && check == "未")
		}
		addIf(i, isGuasu(yz, z), "寡宿")

		// 天喜（红鸾天喜：年支顺数到酉方向）
		tianxiMap := map[string]string{
			"子": "酉", "丑": "申", "寅": "未", "卯": "午",
			"辰": "巳", "巳": "辰", "午": "卯", "未": "寅",
			"申": "丑", "酉": "子", "戌": "亥", "亥": "戌",
		}
		if tianxiZhi, ok := tianxiMap[yz]; ok {
			addIf(i, z == tianxiZhi, "天喜")
		}

		// 灾煞（新增）：申子辰→午, 寅午戌→子, 亥卯未→酉, 巳酉丑→卯
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
			addIf(i, z == zaizhi, "灾煞")
		}
	}

	// 福星贵人（年干 → 四柱地支）
	// Fix: 庚→午（庚马，原错误为庚→巳）
	fuxingMap := map[string][]string{
		"甲": {"寅"},
		"乙": {"丑", "子"},
		"丙": {"子", "亥"},
		"丁": {"酉"},
		"戊": {"申"},
		"己": {"未", "午"},
		"庚": {"午"},        // Fix: 庚马→午
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
	// 天医 = 月支前一辰
	// ══════════════════════════════════════════════════════════
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
		tianYiZhi := zhiOrder[(mzIdx+11)%12]
		for i, z := range zhis {
			addIf(i, z == tianYiZhi, "天医")
		}
	}

	// ══════════════════════════════════════════════════════════
	// 第六组：干支自柱
	// Fix: 阴差阳错、日德、魁罡、十恶大败均仅查日柱
	// ══════════════════════════════════════════════════════════
	for i, gz := range ganZhis {
		if i == 2 {
			// 阴差阳错（仅日柱）
			yinchaSet := map[string]bool{
				"丙子": true, "丁丑": true, "戊寅": true, "辛卯": true, "壬辰": true, "癸巳": true,
				"丙午": true, "丁未": true, "戊申": true, "辛酉": true, "壬戌": true, "癸亥": true,
			}
			addIf(i, yinchaSet[gz], "阴差阳错")

			// 日德（仅日柱）：日干自坐禄地
			riDeSet := map[string]bool{
				"甲寅": true, "乙卯": true, "丙午": true,
				"庚申": true, "辛酉": true, "壬子": true,
			}
			addIf(i, riDeSet[gz], "日德")

			// 魁罡（仅日柱）
			kuiGangSet := map[string]bool{
				"庚辰": true, "庚戌": true, "壬辰": true, "壬戌": true, "戊戌": true,
			}
			addIf(i, kuiGangSet[gz], "魁罡")

			// 十恶大败（仅日柱）
			shiESet := map[string]bool{
				"甲辰": true, "乙巳": true, "丙申": true, "丁亥": true, "戊戌": true,
				"己丑": true, "庚辰": true, "辛巳": true, "壬申": true, "癸亥": true,
			}
			addIf(i, shiESet[gz], "十恶大败")
		}
		_ = gz
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
	// 第八组：天罗地网（全盘地支扫描）
	// Fix: 戌亥均标天罗，辰巳均标地网（原只标戌/辰）
	// ══════════════════════════════════════════════════════════
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
