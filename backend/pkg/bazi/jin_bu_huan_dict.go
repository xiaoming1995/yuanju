package bazi

// ────────────────────────────────────────────────────────────────
// 数据结构
// ────────────────────────────────────────────────────────────────

// DayunRule 梁湘润《子平教材讲义》金不换大运——调候用神合表
// 每条记录对应一个日干+月支组合
type DayunRule struct {
	GanXi []string `json:"gan_xi"` // 调候喜天干（前5年依据）
	GanJi []string `json:"gan_ji"` // 调候忌天干（前5年依据）
	ZhiXi []string `json:"zhi_xi"` // 金不换喜地支（后5年依据）
	ZhiJi []string `json:"zhi_ji"` // 金不换忌地支（后5年依据）
	Note  string   `json:"note"`   // 特注（如"身强"、"天折"等）
	Verse string   `json:"verse"`  // 金不换原诗
}

// JinBuHuanResult 挂载到 DayunItem 上的计算结果
// 天干管前5年、地支管后5年，分别独立评级
type JinBuHuanResult struct {
	QianLevel string `json:"qian_level"` // 前5年评级（天干决定）
	QianDesc  string `json:"qian_desc"`  // 前5年说明
	HouLevel  string `json:"hou_level"`  // 后5年评级（地支决定）
	HouDesc   string `json:"hou_desc"`   // 后5年说明
	Verse     string `json:"verse"`      // 金不换原诗
}

// ────────────────────────────────────────────────────────────────
// 查询 & 计算
// ────────────────────────────────────────────────────────────────

// GetDayunRule 查询合表规则
func GetDayunRule(dayGan, monthZhi string) *DayunRule {
	key := dayGan + "_" + monthZhi
	if rule, ok := dayunRuleDict[key]; ok {
		return &rule
	}
	return nil
}

// contains 检查 slice 中是否包含指定字符串（用于天干精确匹配）
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// 地支五行分组（用于方向性描述的展开匹配）
var waterWoodZhi = map[string]bool{"亥": true, "子": true, "寅": true, "卯": true}
var metalWaterZhi = map[string]bool{"申": true, "酉": true, "亥": true, "子": true}

// matchesZhiItem 判断合表中单条地支条目是否命中大运地支。
// 合表中除真实地支字符外，还存在方向性描述（如"逆行水木"、"顺运"），
// 需结合命主顺/逆排信息（isShunXing）展开判断。
func matchesZhiItem(item, dayunZhi string, isShunXing bool) bool {
	switch item {
	case "顺运", "顺行":
		return isShunXing
	case "逆运", "逆行":
		return !isShunXing
	case "逆行水木":
		return !isShunXing && waterWoodZhi[dayunZhi]
	case "逆行金水":
		return !isShunXing && metalWaterZhi[dayunZhi]
	case "不拘顺逆", "顺逆不拘", "不忌顺逆运":
		// 无特殊方向喜忌，不触发吉/凶
		return false
	default:
		return item == dayunZhi
	}
}

// matchesZhiList 遍历 ZhiXi 或 ZhiJi，判断是否有任一条目命中大运地支
func matchesZhiList(list []string, dayunZhi string, isShunXing bool) bool {
	for _, item := range list {
		if matchesZhiItem(item, dayunZhi, isShunXing) {
			return true
		}
	}
	return false
}

// CalcJinBuHuanDayun 计算某柱大运的金不换评价
// 前5年：大运天干 vs 合表调候喜忌天干
// 后5年：大运地支 vs 合表金不换喜忌地支（含方向性描述展开）
// isShunXing：命主是否顺排（男阳年/女阴年为顺）
func CalcJinBuHuanDayun(dayGan, monthZhi, dayunGan, dayunZhi string, isShunXing bool) *JinBuHuanResult {
	rule := GetDayunRule(dayGan, monthZhi)
	if rule == nil {
		return nil
	}

	// ── 前5年：天干评级 ──────────────────────────────────
	qianLevel := "平"
	qianDesc := ""

	if contains(rule.GanXi, dayunGan) {
		qianLevel = "吉"
		qianDesc = dayunGan + "为调候喜用天干，前5年顺遂吉利。"
	} else if contains(rule.GanJi, dayunGan) {
		qianLevel = "凶"
		qianDesc = dayunGan + "为调候忌神天干，前5年多阻碍不顺。"
	} else {
		qianDesc = dayunGan + "不在调候喜忌天干之列，前5年中平论之。"
	}

	// ── 后5年：地支评级 ──────────────────────────────────
	houLevel := "平"
	houDesc := ""

	if matchesZhiList(rule.ZhiXi, dayunZhi, isShunXing) {
		houLevel = "吉"
		houDesc = dayunZhi + "为金不换喜用地支，后5年大运亨通。"
	} else if matchesZhiList(rule.ZhiJi, dayunZhi, isShunXing) {
		houLevel = "凶"
		houDesc = dayunZhi + "为金不换忌神地支，后5年运势不利。"
	} else {
		houDesc = dayunZhi + "不在金不换喜忌地支之列，后5年中平论之。"
	}

	// 附加特注信息
	if rule.Note != "" {
		if qianLevel != "平" {
			qianDesc += "（" + rule.Note + "）"
		}
		if houLevel != "平" {
			houDesc += "（" + rule.Note + "）"
		}
	}

	return &JinBuHuanResult{
		QianLevel: qianLevel,
		QianDesc:  qianDesc,
		HouLevel:  houLevel,
		HouDesc:   houDesc,
		Verse:     rule.Verse,
	}
}

// ────────────────────────────────────────────────────────────────
// 合表字典：10天干 × 12月支 = 120条
// 数据来源：梁湘润《子平教材讲义》第二级次
//           金不换大运——调候用神合表（一）（二）
// ────────────────────────────────────────────────────────────────

var dayunRuleDict = map[string]DayunRule{
	// ==================
	// 甲木
	// ==================
	"甲_子": {
		GanXi: []string{"丁"},
		GanJi: []string{"壬", "癸"},
		ZhiXi: []string{"逆行水木"},
		ZhiJi: []string{"午", "未"},
		Note:  "官杀",
		Verse: "甲木子月正官格，丁火为用庚劈甲。运行逆行水木美，午未忌行多蹉跎。",
	},
	"甲_丑": {
		GanXi: []string{"丁"},
		GanJi: []string{"壬", "辛", "癸"},
		ZhiXi: []string{"寅", "卯", "申", "酉"},
		ZhiJi: []string{"午", "未"},
		Note:  "",
		Verse: "",
	},
	"甲_寅": {
		GanXi: []string{"丙", "癸"},
		GanJi: []string{"戊", "己"},
		ZhiXi: []string{"逆行金水"},
		ZhiJi: []string{"巳", "午", "未"},
		Note:  "正官",
		Verse: "",
	},
	"甲_卯": {
		GanXi: []string{"丙", "癸"},
		GanJi: []string{"戊", "己"},
		ZhiXi: []string{"不拘顺逆"},
		ZhiJi: []string{},
		Note:  "",
		Verse: "",
	},
	"甲_辰": {
		GanXi: []string{"壬", "癸"},
		GanJi: []string{"庚"},
		ZhiXi: []string{"申", "酉"},
		ZhiJi: []string{},
		Note:  "",
		Verse: "",
	},
	"甲_巳": {
		GanXi: []string{"壬", "癸"},
		GanJi: []string{"庚"},
		ZhiXi: []string{"亥", "子"},
		ZhiJi: []string{"午", "未"},
		Note:  "财、杀",
		Verse: "",
	},
	"甲_午": {
		GanXi: []string{"壬", "丁", "癸"},
		GanJi: []string{"戊", "己"},
		ZhiXi: []string{"亥", "子"},
		ZhiJi: []string{"申", "酉"},
		Note:  "",
		Verse: "",
	},
	"甲_未": {
		GanXi: []string{"寅", "卯"},
		GanJi: []string{"申", "酉"},
		ZhiXi: []string{"寅", "卯", "逆运"},
		ZhiJi: []string{"申", "酉"},
		Note:  "",
		Verse: "",
	},
	"甲_申": {
		GanXi: []string{"丁"},
		GanJi: []string{"壬", "辛", "癸"},
		ZhiXi: []string{"亥", "子"},
		ZhiJi: []string{"午", "巳"},
		Note:  "杀旺",
		Verse: "",
	},
	"甲_酉": {
		GanXi: []string{"丙", "癸"},
		GanJi: []string{"戊", "己"},
		ZhiXi: []string{"顺运", "寅", "卯", "亥", "未"},
		ZhiJi: []string{},
		Note:  "",
		Verse: "",
	},
	"甲_戌": {
		GanXi: []string{"丙", "癸"},
		GanJi: []string{"戊", "己"},
		ZhiXi: []string{"顺运", "寅", "卯", "亥", "未"},
		ZhiJi: []string{"午"},
		Note:  "",
		Verse: "",
	},
	"甲_亥": {
		GanXi: []string{"丁", "癸"},
		GanJi: []string{"壬"},
		ZhiXi: []string{"顺运", "寅", "卯"},
		ZhiJi: []string{"午", "未"},
		Note:  "",
		Verse: "",
	},

	// ==================
	// 乙木
	// ==================
	"乙_子": {
		GanXi: []string{"丙"},
		GanJi: []string{"壬", "癸"},
		ZhiXi: []string{"戊", "申", "酉"},
		ZhiJi: []string{"亥", "丑"},
		Note:  "",
		Verse: "",
	},
	"乙_丑": {
		GanXi: []string{"丙"},
		GanJi: []string{"壬", "癸"},
		ZhiXi: []string{"巳", "午", "未", "寅", "卯"},
		ZhiJi: []string{"亥", "子"},
		Note:  "",
		Verse: "",
	},
	"乙_寅": {
		GanXi: []string{"丙"},
		GanJi: []string{"壬", "癸"},
		ZhiXi: []string{"丑", "午", "申", "酉"},
		ZhiJi: []string{"巳", "亥", "子"},
		Note:  "",
		Verse: "",
	},
	"乙_卯": {
		GanXi: []string{"丙"},
		GanJi: []string{"壬", "癸"},
		ZhiXi: []string{"卯", "子", "亥"},
		ZhiJi: []string{"巳", "午", "亥"},
		Note:  "",
		Verse: "",
	},
	"乙_辰": {
		GanXi: []string{"丙", "癸"},
		GanJi: []string{"壬"},
		ZhiXi: []string{"申", "酉"},
		ZhiJi: []string{},
		Note:  "",
		Verse: "",
	},
	"乙_巳": {
		GanXi: []string{"丙", "癸"},
		GanJi: []string{"壬"},
		ZhiXi: []string{"卯", "子", "亥"},
		ZhiJi: []string{"申", "酉"},
		Note:  "壽損",
		Verse: "",
	},
	"乙_午": {
		GanXi: []string{"壬"},
		GanJi: []string{"己", "庚"},
		ZhiXi: []string{"辰", "卯", "寅"},
		ZhiJi: []string{"申", "酉"},
		Note:  "災疾",
		Verse: "",
	},
	"乙_未": {
		GanXi: []string{"丙", "癸"},
		GanJi: []string{"壬"},
		ZhiXi: []string{"顺行"},
		ZhiJi: []string{},
		Note:  "",
		Verse: "",
	},
	"乙_申": {
		GanXi: []string{"丙", "癸"},
		GanJi: []string{"壬"},
		ZhiXi: []string{"亥", "子", "寅", "卯"},
		ZhiJi: []string{"午"},
		Note:  "",
		Verse: "",
	},
	"乙_酉": {
		GanXi: []string{"丙", "癸"},
		GanJi: []string{"壬"},
		ZhiXi: []string{"亥", "子", "寅", "卯"},
		ZhiJi: []string{"午"},
		Note:  "",
		Verse: "",
	},
	"乙_戌": {
		GanXi: []string{"丙", "癸"},
		GanJi: []string{"壬"},
		ZhiXi: []string{"亥", "子", "寅", "卯"},
		ZhiJi: []string{},
		Note:  "",
		Verse: "",
	},
	"乙_亥": {
		GanXi: []string{"丙"},
		GanJi: []string{"壬", "癸"},
		ZhiXi: []string{"顺行", "巳", "午"},
		ZhiJi: []string{"亥", "子"},
		Note:  "",
		Verse: "",
	},

	// ==================
	// 丙火
	// ==================
	"丙_子": {
		GanXi: []string{"丁"},
		GanJi: []string{"壬", "癸"},
		ZhiXi: []string{"逆行水木"},
		ZhiJi: []string{"午", "未"},
		Note:  "官杀",
		Verse: "丙火子月正官格，壬水当权用食伤。运行木火多快意，金水运来事不祥。",
	},
	"丙_丑": {
		GanXi: []string{"丁"},
		GanJi: []string{"壬", "辛", "癸"},
		ZhiXi: []string{"寅", "卯", "巳", "午"},
		ZhiJi: []string{"午", "未"},
		Note:  "喜木火",
		Verse: "",
	},
	"丙_寅": {
		GanXi: []string{"壬"},
		GanJi: []string{"戊", "己"},
		ZhiXi: []string{"巳", "午", "未"},
		ZhiJi: []string{"子"},
		Note:  "忌无根",
		Verse: "",
	},
	"丙_卯": {
		GanXi: []string{"壬", "癸"},
		GanJi: []string{"庚"},
		ZhiXi: []string{"不拘顺逆"},
		ZhiJi: []string{"午", "未"},
		Note:  "喜身旺",
		Verse: "",
	},
	"丙_辰": {
		GanXi: []string{"壬", "癸"},
		GanJi: []string{"庚"},
		ZhiXi: []string{"申", "酉"},
		ZhiJi: []string{},
		Note:  "财官",
		Verse: "",
	},
	"丙_巳": {
		GanXi: []string{"壬", "癸"},
		GanJi: []string{"庚"},
		ZhiXi: []string{"申", "酉"},
		ZhiJi: []string{"巳", "午"},
		Note:  "",
		Verse: "",
	},
	"丙_午": {
		GanXi: []string{"壬"},
		GanJi: []string{"戊", "己"},
		ZhiXi: []string{"亥", "子"},
		ZhiJi: []string{"巳", "午", "未"},
		Note:  "",
		Verse: "",
	},
	"丙_未": {
		GanXi: []string{"壬"},
		GanJi: []string{"戊", "己"},
		ZhiXi: []string{"申", "酉", "亥", "子"},
		ZhiJi: []string{"午", "未"},
		Note:  "",
		Verse: "",
	},
	"丙_申": {
		GanXi: []string{"壬", "丁"},
		GanJi: []string{"壬", "辛", "癸"},
		ZhiXi: []string{"寅", "卯"},
		ZhiJi: []string{"午", "巳"},
		Note:  "财、杀旺",
		Verse: "",
	},
	"丙_酉": {
		GanXi: []string{"壬"},
		GanJi: []string{"戊", "己"},
		ZhiXi: []string{"顺运", "寅", "卯", "亥", "未"},
		ZhiJi: []string{"午"},
		Note:  "",
		Verse: "",
	},
	"丙_戌": {
		GanXi: []string{"甲", "壬"},
		GanJi: []string{"戊", "己"},
		ZhiXi: []string{"顺运", "寅", "卯", "亥", "未"},
		ZhiJi: []string{"午"},
		Note:  "",
		Verse: "丙火戌月入火库，甲木疏通壬水辅。运行东方木火明，西北金水成碍阻。",
	},
	"丙_亥": {
		GanXi: []string{"甲", "壬"},
		GanJi: []string{"戊", "己"},
		ZhiXi: []string{"顺运", "寅", "卯"},
		ZhiJi: []string{"辰"},
		Note:  "",
		Verse: "",
	},

	// ==================
	// 丁火
	// ==================
	"丁_子": {
		GanXi: []string{"甲"},
		GanJi: []string{"壬", "癸"},
		ZhiXi: []string{"寅", "卯"},
		ZhiJi: []string{"申", "酉"},
		Note:  "",
		Verse: "丁火子月甲木功，庚金劈甲引丁红。运行东方林木旺，水多北地反成空。",
	},
	"丁_丑": {
		GanXi: []string{"甲"},
		GanJi: []string{"壬", "癸"},
		ZhiXi: []string{"寅", "卯", "巳", "午"},
		ZhiJi: []string{"申", "酉"},
		Note:  "",
		Verse: "",
	},
	"丁_寅": {
		GanXi: []string{"庚"},
		GanJi: []string{"壬"},
		ZhiXi: []string{"申", "酉"},
		ZhiJi: []string{"亥", "子"},
		Note:  "根深",
		Verse: "",
	},
	"丁_卯": {
		GanXi: []string{"庚"},
		GanJi: []string{"壬"},
		ZhiXi: []string{"申", "酉"},
		ZhiJi: []string{"亥", "子"},
		Note:  "根深",
		Verse: "",
	},
	"丁_辰": {
		GanXi: []string{"甲", "庚"},
		GanJi: []string{"壬", "癸"},
		ZhiXi: []string{"申", "酉"},
		ZhiJi: []string{"亥", "子"},
		Note:  "正官",
		Verse: "",
	},
	"丁_巳": {
		GanXi: []string{"甲", "庚"},
		GanJi: []string{"壬", "癸"},
		ZhiXi: []string{"申", "酉", "亥"},
		ZhiJi: []string{"午"},
		Note:  "",
		Verse: "",
	},
	"丁_午": {
		GanXi: []string{"壬"},
		GanJi: []string{"甲"},
		ZhiXi: []string{"亥", "子", "申", "酉"},
		ZhiJi: []string{"巳", "午"},
		Note:  "",
		Verse: "",
	},
	"丁_未": {
		GanXi: []string{"甲", "壬"},
		GanJi: []string{"戊"},
		ZhiXi: []string{"申", "酉"},
		ZhiJi: []string{"午", "未"},
		Note:  "七杀",
		Verse: "",
	},
	"丁_申": {
		GanXi: []string{"甲", "壬"},
		GanJi: []string{"庚"},
		ZhiXi: []string{"寅", "卯"},
		ZhiJi: []string{"申", "酉"},
		Note:  "财官",
		Verse: "",
	},
	"丁_酉": {
		GanXi: []string{"甲"},
		GanJi: []string{"壬"},
		ZhiXi: []string{"寅", "卯"},
		ZhiJi: []string{"申", "酉"},
		Note:  "",
		Verse: "",
	},
	"丁_戌": {
		GanXi: []string{"甲"},
		GanJi: []string{"壬"},
		ZhiXi: []string{"寅", "卯", "辰"},
		ZhiJi: []string{"申", "酉"},
		Note:  "",
		Verse: "",
	},
	"丁_亥": {
		GanXi: []string{"甲"},
		GanJi: []string{"壬"},
		ZhiXi: []string{"寅", "卯"},
		ZhiJi: []string{"申", "酉"},
		Note:  "",
		Verse: "",
	},

	// ==================
	// 戊土
	// ==================
	"戊_子": {
		GanXi: []string{"丙", "甲"},
		GanJi: []string{"戊"},
		ZhiXi: []string{"寅", "卯", "巳", "午"},
		ZhiJi: []string{"申", "酉", "亥"},
		Note:  "弱",
		Verse: "",
	},
	"戊_丑": {
		GanXi: []string{"丙", "甲"},
		GanJi: []string{"戊"},
		ZhiXi: []string{"寅", "卯", "巳", "午"},
		ZhiJi: []string{"申", "酉"},
		Note:  "混忌",
		Verse: "",
	},
	"戊_寅": {
		GanXi: []string{"丙"},
		GanJi: []string{"壬"},
		ZhiXi: []string{"巳", "午"},
		ZhiJi: []string{"子"},
		Note:  "忌杀旺",
		Verse: "",
	},
	"戊_卯": {
		GanXi: []string{"丙"},
		GanJi: []string{"壬"},
		ZhiXi: []string{"巳", "午"},
		ZhiJi: []string{"子"},
		Note:  "柔",
		Verse: "",
	},
	"戊_辰": {
		GanXi: []string{"丙", "癸"},
		GanJi: []string{"辛", "己"},
		ZhiXi: []string{"巳", "午"},
		ZhiJi: []string{"亥", "子"},
		Note:  "杀旺",
		Verse: "",
	},
	"戊_巳": {
		GanXi: []string{"壬"},
		GanJi: []string{"甲"},
		ZhiXi: []string{"申", "酉"},
		ZhiJi: []string{"寅", "卯"},
		Note:  "身强",
		Verse: "",
	},
	"戊_午": {
		GanXi: []string{"壬"},
		GanJi: []string{"甲", "丙"},
		ZhiXi: []string{"亥", "子"},
		ZhiJi: []string{"巳", "午"},
		Note:  "肩",
		Verse: "",
	},
	"戊_未": {
		GanXi: []string{"癸"},
		GanJi: []string{"丙", "甲"},
		ZhiXi: []string{"申", "酉", "亥", "子"},
		ZhiJi: []string{"午", "未"},
		Note:  "",
		Verse: "",
	},
	"戊_申": {
		GanXi: []string{"丙", "癸"},
		GanJi: []string{"壬"},
		ZhiXi: []string{"巳", "午"},
		ZhiJi: []string{"亥", "子"},
		Note:  "",
		Verse: "",
	},
	"戊_酉": {
		GanXi: []string{"丙", "癸"},
		GanJi: []string{"壬"},
		ZhiXi: []string{"巳", "午"},
		ZhiJi: []string{"亥", "子"},
		Note:  "财官",
		Verse: "",
	},
	"戊_戌": {
		GanXi: []string{"甲", "丙", "癸"},
		GanJi: []string{"壬"},
		ZhiXi: []string{"寅", "卯", "巳", "午"},
		ZhiJi: []string{"亥", "子"},
		Note:  "",
		Verse: "",
	},
	"戊_亥": {
		GanXi: []string{"甲", "丙"},
		GanJi: []string{"壬"},
		ZhiXi: []string{"寅", "卯", "巳", "午"},
		ZhiJi: []string{"申", "酉", "亥"},
		Note:  "",
		Verse: "",
	},

	// ==================
	// 己土
	// ==================
	"己_子": {
		GanXi: []string{"丙", "甲"},
		GanJi: []string{"戊"},
		ZhiXi: []string{"寅", "卯", "巳", "午"},
		ZhiJi: []string{"申", "酉"},
		Note:  "根深",
		Verse: "",
	},
	"己_丑": {
		GanXi: []string{"丙", "甲"},
		GanJi: []string{"戊"},
		ZhiXi: []string{"寅", "卯", "巳", "午"},
		ZhiJi: []string{"申", "酉"},
		Note:  "喜七杀",
		Verse: "",
	},
	"己_未": {
		GanXi: []string{"癸", "丙"},
		GanJi: []string{"乙"},
		ZhiXi: []string{"不拘顺逆"},
		ZhiJi: []string{"丑"},
		Note:  "",
		Verse: "",
	},
	"己_申": {
		GanXi: []string{"丙"},
		GanJi: []string{"甲"},
		ZhiXi: []string{"寅", "卯", "巳", "午"},
		ZhiJi: []string{"申", "戌"},
		Note:  "身强",
		Verse: "",
	},
	"己_酉": {
		GanXi: []string{"丙", "癸"},
		GanJi: []string{"辛"},
		ZhiXi: []string{"顺逆不拘"},
		ZhiJi: []string{"申", "戌"},
		Note:  "忌无根",
		Verse: "",
	},
	"己_戌": {
		GanXi: []string{"甲", "丙", "癸"},
		GanJi: []string{"庚"},
		ZhiXi: []string{"顺逆不拘"},
		ZhiJi: []string{"戌"},
		Note:  "",
		Verse: "",
	},
	"己_亥": {
		GanXi: []string{"丙", "甲"},
		GanJi: []string{"己"},
		ZhiXi: []string{"子", "丑"},
		ZhiJi: []string{"寅", "卯"},
		Note:  "忌无根",
		Verse: "",
	},
	"己_寅": {
		GanXi: []string{"丙"},
		GanJi: []string{"壬"},
		ZhiXi: []string{"巳", "午"},
		ZhiJi: []string{"亥"},
		Note:  "",
		Verse: "",
	},
	"己_卯": {
		GanXi: []string{"丙"},
		GanJi: []string{"壬"},
		ZhiXi: []string{"巳", "午"},
		ZhiJi: []string{"亥"},
		Note:  "",
		Verse: "",
	},
	"己_辰": {
		GanXi: []string{"丙"},
		GanJi: []string{"壬"},
		ZhiXi: []string{"巳", "午"},
		ZhiJi: []string{"亥", "子"},
		Note:  "",
		Verse: "",
	},
	"己_巳": {
		GanXi: []string{"癸"},
		GanJi: []string{"丙"},
		ZhiXi: []string{"申", "酉"},
		ZhiJi: []string{"午"},
		Note:  "",
		Verse: "",
	},
	"己_午": {
		GanXi: []string{"癸"},
		GanJi: []string{"丙", "甲"},
		ZhiXi: []string{"申", "酉", "亥", "子"},
		ZhiJi: []string{"午", "巳"},
		Note:  "",
		Verse: "",
	},

	// ==================
	// 庚金
	// ==================
	"庚_子": {
		GanXi: []string{"丁", "甲", "丙"},
		GanJi: []string{"癸"},
		ZhiXi: []string{"寅", "卯", "辰"},
		ZhiJi: []string{"午"},
		Note:  "喜财杀",
		Verse: "",
	},
	"庚_丑": {
		GanXi: []string{"丁", "甲", "丙"},
		GanJi: []string{"癸"},
		ZhiXi: []string{"寅", "卯", "巳", "午"},
		ZhiJi: []string{"未"},
		Note:  "喜木火",
		Verse: "",
	},
	"庚_寅": {
		GanXi: []string{"丙", "甲"},
		GanJi: []string{"辛"},
		ZhiXi: []string{"卯", "丑", "巳", "酉"},
		ZhiJi: []string{"子", "午"},
		Note:  "喜透土",
		Verse: "",
	},
	"庚_卯": {
		GanXi: []string{"丁"},
		GanJi: []string{"癸"},
		ZhiXi: []string{"午", "巳", "酉"},
		ZhiJi: []string{"子"},
		Note:  "",
		Verse: "",
	},
	"庚_辰": {
		GanXi: []string{"甲", "丁"},
		GanJi: []string{"癸"},
		ZhiXi: []string{"午"},
		ZhiJi: []string{"申", "酉"},
		Note:  "喜身旺",
		Verse: "",
	},
	"庚_巳": {
		GanXi: []string{"壬"},
		GanJi: []string{"庚"},
		ZhiXi: []string{"子", "卯", "寅", "亥"},
		ZhiJi: []string{"午"},
		Note:  "忌无根",
		Verse: "",
	},
	"庚_午": {
		GanXi: []string{"壬"},
		GanJi: []string{"丁"},
		ZhiXi: []string{"子"},
		ZhiJi: []string{"寅", "卯", "辰", "巳"},
		Note:  "有水",
		Verse: "",
	},
	"庚_未": {
		GanXi: []string{"丁", "甲"},
		GanJi: []string{"乙", "戊", "己"},
		ZhiXi: []string{"寅", "卯", "辰", "巳", "午"},
		ZhiJi: []string{"戌", "亥"},
		Note:  "忌土重",
		Verse: "",
	},
	"庚_申": {
		GanXi: []string{"丁", "甲"},
		GanJi: []string{"乙", "戊", "己"},
		ZhiXi: []string{"寅", "卯", "辰", "巳"},
		ZhiJi: []string{"酉"},
		Note:  "忌过旺",
		Verse: "",
	},
	"庚_酉": {
		GanXi: []string{"丁", "甲"},
		GanJi: []string{"辛"},
		ZhiXi: []string{"巳", "午"},
		ZhiJi: []string{"申"},
		Note:  "喜七杀",
		Verse: "",
	},
	"庚_戌": {
		GanXi: []string{"甲", "壬"},
		GanJi: []string{"癸"},
		ZhiXi: []string{"寅", "卯", "辰", "巳"},
		ZhiJi: []string{"酉", "辰"},
		Note:  "喜土",
		Verse: "",
	},
	"庚_亥": {
		GanXi: []string{"丁", "丙"},
		GanJi: []string{"癸"},
		ZhiXi: []string{"子", "辰", "巳", "午"},
		ZhiJi: []string{"卯", "寅"},
		Note:  "",
		Verse: "",
	},

	// ==================
	// 辛金
	// ==================
	"辛_子": {
		GanXi: []string{"丙", "壬"},
		GanJi: []string{"癸"},
		ZhiXi: []string{"寅", "卯", "巳", "午"},
		ZhiJi: []string{"亥", "子"},
		Note:  "喜丙丁",
		Verse: "",
	},
	"辛_丑": {
		GanXi: []string{"丙", "壬"},
		GanJi: []string{"癸"},
		ZhiXi: []string{"寅", "卯", "辰", "戌", "午"},
		ZhiJi: []string{"亥"},
		Note:  "喜丁",
		Verse: "",
	},
	"辛_寅": {
		GanXi: []string{"壬"},
		GanJi: []string{"辛"},
		ZhiXi: []string{"寅", "卯", "辰"},
		ZhiJi: []string{"午", "未"},
		Note:  "土多天",
		Verse: "",
	},
	"辛_卯": {
		GanXi: []string{"壬"},
		GanJi: []string{"丁"},
		ZhiXi: []string{"巳", "午"},
		ZhiJi: []string{"亥", "子"},
		Note:  "喜支坐",
		Verse: "",
	},
	"辛_辰": {
		GanXi: []string{"壬"},
		GanJi: []string{"丁"},
		ZhiXi: []string{"申", "酉"},
		ZhiJi: []string{"午", "未"},
		Note:  "",
		Verse: "",
	},
	"辛_巳": {
		GanXi: []string{"壬", "甲"},
		GanJi: []string{"丁", "戊"},
		ZhiXi: []string{"不忌顺逆运"},
		ZhiJi: []string{},
		Note:  "",
		Verse: "",
	},
	"辛_午": {
		GanXi: []string{"壬", "己"},
		GanJi: []string{"丁"},
		ZhiXi: []string{"申", "戌", "午"},
		ZhiJi: []string{},
		Note:  "喜财",
		Verse: "",
	},
	"辛_未": {
		GanXi: []string{"壬"},
		GanJi: []string{"申"},
		ZhiXi: []string{"寅", "卯", "巳", "午"},
		ZhiJi: []string{"申"},
		Note:  "忌金",
		Verse: "",
	},
	"辛_申": {
		GanXi: []string{"壬"},
		GanJi: []string{"辛"},
		ZhiXi: []string{"寅", "卯", "巳", "午"},
		ZhiJi: []string{"酉"},
		Note:  "酉",
		Verse: "",
	},
	"辛_酉": {
		GanXi: []string{"壬"},
		GanJi: []string{"辛"},
		ZhiXi: []string{"亥", "子", "寅", "卯"},
		ZhiJi: []string{"申", "酉"},
		Note:  "",
		Verse: "",
	},
	"辛_戌": {
		GanXi: []string{"壬"},
		GanJi: []string{"丁"},
		ZhiXi: []string{"不忌顺逆"},
		ZhiJi: []string{"子"},
		Note:  "",
		Verse: "",
	},
	"辛_亥": {
		GanXi: []string{"壬"},
		GanJi: []string{"丁"},
		ZhiXi: []string{"巳", "丑"},
		ZhiJi: []string{"子", "丑"},
		Note:  "",
		Verse: "",
	},

	// ==================
	// 壬水
	// ==================
	"壬_子": {
		GanXi: []string{"戊"},
		GanJi: []string{"壬", "癸"},
		ZhiXi: []string{"巳", "午", "未"},
		ZhiJi: []string{"亥", "子"},
		Note:  "喜丁",
		Verse: "",
	},
	"壬_丑": {
		GanXi: []string{"丙", "甲"},
		GanJi: []string{"壬"},
		ZhiXi: []string{"寅", "卯", "巳", "午"},
		ZhiJi: []string{"亥", "子"},
		Note:  "喜木火",
		Verse: "",
	},
	"壬_寅": {
		GanXi: []string{"庚", "丙"},
		GanJi: []string{"戊"},
		ZhiXi: []string{"巳", "午", "未"},
		ZhiJi: []string{"辰", "丑"},
		Note:  "喜透金",
		Verse: "",
	},
	"壬_卯": {
		GanXi: []string{"庚"},
		GanJi: []string{"戊"},
		ZhiXi: []string{"巳", "午", "申", "酉"},
		ZhiJi: []string{"丑"},
		Note:  "",
		Verse: "",
	},
	"壬_辰": {
		GanXi: []string{"甲", "庚"},
		GanJi: []string{"戊", "壬"},
		ZhiXi: []string{"巳", "午"},
		ZhiJi: []string{"辰", "丑"},
		Note:  "",
		Verse: "",
	},
	"壬_巳": {
		GanXi: []string{"壬", "辛"},
		GanJi: []string{"戊"},
		ZhiXi: []string{"申", "酉", "亥"},
		ZhiJi: []string{"巳", "午"},
		Note:  "身弱",
		Verse: "",
	},
	"壬_午": {
		GanXi: []string{"癸", "辛"},
		GanJi: []string{"戊"},
		ZhiXi: []string{"申", "酉", "亥"},
		ZhiJi: []string{"巳", "午"},
		Note:  "忌土",
		Verse: "",
	},
	"壬_未": {
		GanXi: []string{"辛", "甲"},
		GanJi: []string{"戊", "丙"},
		ZhiXi: []string{"申", "酉", "亥"},
		ZhiJi: []string{"巳", "午"},
		Note:  "",
		Verse: "",
	},
	"壬_申": {
		GanXi: []string{"戊", "丁"},
		GanJi: []string{"庚", "辛"},
		ZhiXi: []string{"巳", "午"},
		ZhiJi: []string{"亥", "子"},
		Note:  "身旺",
		Verse: "",
	},
	"壬_酉": {
		GanXi: []string{"甲"},
		GanJi: []string{"庚", "辛"},
		ZhiXi: []string{"巳", "午", "寅", "卯"},
		ZhiJi: []string{"亥", "子"},
		Note:  "",
		Verse: "",
	},
	"壬_戌": {
		GanXi: []string{"甲", "丙"},
		GanJi: []string{"戊"},
		ZhiXi: []string{"寅", "卯", "巳", "午"},
		ZhiJi: []string{"亥", "子"},
		Note:  "",
		Verse: "",
	},
	"壬_亥": {
		GanXi: []string{"戊", "丙"},
		GanJi: []string{"壬"},
		ZhiXi: []string{"寅", "卯", "巳", "午"},
		ZhiJi: []string{"亥", "子"},
		Note:  "旺",
		Verse: "",
	},

	// ==================
	// 癸水
	// ==================
	"癸_子": {
		GanXi: []string{"丙"},
		GanJi: []string{"壬", "癸"},
		ZhiXi: []string{"巳", "午", "未"},
		ZhiJi: []string{"亥", "子"},
		Note:  "旺",
		Verse: "",
	},
	"癸_丑": {
		GanXi: []string{"丙"},
		GanJi: []string{"壬"},
		ZhiXi: []string{"寅", "卯", "巳", "午"},
		ZhiJi: []string{"亥", "子"},
		Note:  "",
		Verse: "",
	},
	"癸_寅": {
		GanXi: []string{"庚", "辛"},
		GanJi: []string{"戊"},
		ZhiXi: []string{"申", "酉"},
		ZhiJi: []string{"巳", "午"},
		Note:  "",
		Verse: "",
	},
	"癸_卯": {
		GanXi: []string{"庚"},
		GanJi: []string{"戊", "己"},
		ZhiXi: []string{"申", "酉"},
		ZhiJi: []string{"辰", "丑"},
		Note:  "",
		Verse: "",
	},
	"癸_辰": {
		GanXi: []string{"丙", "辛"},
		GanJi: []string{"壬"},
		ZhiXi: []string{"巳", "午"},
		ZhiJi: []string{"亥", "子"},
		Note:  "",
		Verse: "",
	},
	"癸_巳": {
		GanXi: []string{"庚", "辛"},
		GanJi: []string{"戊"},
		ZhiXi: []string{"申", "酉", "亥"},
		ZhiJi: []string{"巳", "午"},
		Note:  "忌无根",
		Verse: "",
	},
	"癸_午": {
		GanXi: []string{"庚", "辛"},
		GanJi: []string{"戊", "己"},
		ZhiXi: []string{"申", "酉", "亥"},
		ZhiJi: []string{"巳", "午"},
		Note:  "",
		Verse: "",
	},
	"癸_未": {
		GanXi: []string{"庚"},
		GanJi: []string{"戊", "己"},
		ZhiXi: []string{"申", "酉", "亥"},
		ZhiJi: []string{"巳", "午"},
		Note:  "",
		Verse: "",
	},
	"癸_申": {
		GanXi: []string{"丁"},
		GanJi: []string{"庚", "辛"},
		ZhiXi: []string{"巳", "午", "寅", "卯"},
		ZhiJi: []string{"申", "酉"},
		Note:  "身旺",
		Verse: "",
	},
	"癸_酉": {
		GanXi: []string{"丙"},
		GanJi: []string{"庚", "辛"},
		ZhiXi: []string{"巳", "午", "寅", "卯"},
		ZhiJi: []string{"申", "酉"},
		Note:  "",
		Verse: "",
	},
	"癸_戌": {
		GanXi: []string{"辛", "丙"},
		GanJi: []string{"壬"},
		ZhiXi: []string{"寅", "卯", "巳", "午"},
		ZhiJi: []string{"亥", "子"},
		Note:  "",
		Verse: "",
	},
	"癸_亥": {
		GanXi: []string{"戊", "丙"},
		GanJi: []string{"壬"},
		ZhiXi: []string{"巳", "午", "寅", "卯"},
		ZhiJi: []string{"亥", "子"},
		Note:  "旺",
		Verse: "",
	},
}
