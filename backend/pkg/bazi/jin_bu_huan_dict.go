package bazi

// JinBuHuanRule 金不换大运歌规则
type JinBuHuanRule struct {
	Verse          string             `json:"verse"`           // 原诗
	GoodDirections []string           `json:"good_directions"` // 喜行方向，如 "南方火"
	BadDirections  []string           `json:"bad_directions"`  // 忌行方向，如 "西方金"
	SpecificZhi    map[string]JBHEval `json:"specific_zhi"`    // 特定地支精确评价（可选）
}

// JBHEval 针对某个方位/地支的细粒度评分
type JBHEval struct {
	Level   string `json:"level"`   // "大吉" / "吉" / "平" / "凶" / "大凶"
	Keyword string `json:"keyword"` // 关键词，如 "发财" / "剥丧"
	Desc    string `json:"desc"`    // 简短描述
}

// JinBuHuanResult 挂载到 DayunItem 上的计算结果
type JinBuHuanResult struct {
	Level   string `json:"level"`
	Keyword string `json:"keyword"`
	Text    string `json:"text"`
}

// GetZhiDirection 将地支映射到三会方位
func GetZhiDirection(zhi string) string {
	switch zhi {
	case "寅", "卯", "辰":
		return "东方木"
	case "巳", "午", "未":
		return "南方火"
	case "申", "酉", "戌":
		return "西方金"
	case "亥", "子", "丑":
		return "北方水"
	default:
		return ""
	}
}

// GetJinBuHuanRule 查询金不换规则
func GetJinBuHuanRule(dayGan, monthZhi string) *JinBuHuanRule {
	key := dayGan + "_" + monthZhi
	if rule, ok := jinBuHuanDict[key]; ok {
		return &rule
	}
	return nil
}

// CalcJinBuHuanDayun 计算某柱大运的金不换评价
func CalcJinBuHuanDayun(dayGan, monthZhi, dayunZhi string) *JinBuHuanResult {
	rule := GetJinBuHuanRule(dayGan, monthZhi)
	if rule == nil {
		return nil
	}

	// 优先匹配特定地支
	if eval, ok := rule.SpecificZhi[dayunZhi]; ok {
		return &JinBuHuanResult{
			Level:   eval.Level,
			Keyword: eval.Keyword,
			Text:    eval.Desc,
		}
	}

	// 按方向匹配
	dir := GetZhiDirection(dayunZhi)
	if dir == "" {
		return nil
	}

	for _, gd := range rule.GoodDirections {
		if gd == dir {
			return &JinBuHuanResult{
				Level:   "吉",
				Keyword: "顺运",
				Text:    rule.Verse,
			}
		}
	}
	for _, bd := range rule.BadDirections {
		if bd == dir {
			return &JinBuHuanResult{
				Level:   "凶",
				Keyword: "逆运",
				Text:    rule.Verse,
			}
		}
	}

	return &JinBuHuanResult{
		Level:   "平",
		Keyword: "平运",
		Text:    rule.Verse,
	}
}

// jinBuHuanDict 金不换大运歌字典
// Key: "日干_月支"
var jinBuHuanDict = map[string]JinBuHuanRule{
	// ==================
	// 甲木
	// ==================
	"甲_子": {
		Verse:          "甲木生来值子宫，水多木漂怕金逢。运行火土多财禄，西北行来见祸凶。",
		GoodDirections: []string{"南方火"},
		BadDirections:  []string{"西方金", "北方水"},
		SpecificZhi: map[string]JBHEval{
			"午": {Level: "大吉", Keyword: "财禄丰盈", Desc: "南方火地暖局，甲木寒命逢火生发，财官两旺。"},
			"巳": {Level: "吉", Keyword: "火暖生发", Desc: "巳火长生之地，助力解冻，事业渐起。"},
			"申": {Level: "凶", Keyword: "金旺克身", Desc: "七杀当头，秋金肃杀，须防官灾破财。"},
			"酉": {Level: "大凶", Keyword: "金锐难挡", Desc: "正官过旺反成灾，身弱不胜官鬼。"},
		},
	},
	"甲_丑": {
		Verse:          "甲木丑月寒气重，调候须用丙火功。南方运地多亨泰，水冷金寒命不通。",
		GoodDirections: []string{"南方火", "东方木"},
		BadDirections:  []string{"北方水", "西方金"},
		SpecificZhi: map[string]JBHEval{
			"午": {Level: "大吉", Keyword: "暖局生辉", Desc: "丙火正位，冻木回春，功名利禄。"},
		},
	},
	"甲_寅": {
		Verse:          "甲木寅月当令旺，庚金劈甲引丁方。运喜西南成大器，北方水泛反凄凉。",
		GoodDirections: []string{"西方金", "南方火"},
		BadDirections:  []string{"北方水"},
		SpecificZhi: map[string]JBHEval{
			"酉": {Level: "大吉", Keyword: "金刀裁木", Desc: "庚金当令劈甲，栋梁之材得用，大贵。"},
			"午": {Level: "吉", Keyword: "木火通明", Desc: "甲木生火吐秀，文章显达。"},
		},
	},
	"甲_卯": {
		Verse:          "甲木卯月禄当权，独喜庚金来削坚。运到西方多得利，东方太旺反颠连。",
		GoodDirections: []string{"西方金"},
		BadDirections:  []string{"东方木"},
		SpecificZhi: map[string]JBHEval{
			"酉": {Level: "大吉", Keyword: "金来雕琢", Desc: "阳刃逢冲煞化权，有金制木成大器。"},
		},
	},
	"甲_辰": {
		Verse:          "甲木辰月土重重，喜用庚金壬水通。运行金水多发达，火土运中命不丰。",
		GoodDirections: []string{"西方金", "北方水"},
		BadDirections:  []string{"南方火"},
		SpecificZhi: map[string]JBHEval{
			"申": {Level: "吉", Keyword: "金水得源", Desc: "申金生水润木，源远流长。"},
		},
	},
	"甲_巳": {
		Verse:          "甲木巳月火正炎，须得癸水润根源。运逢北方财官显，南行火地反招愆。",
		GoodDirections: []string{"北方水"},
		BadDirections:  []string{"南方火"},
		SpecificZhi: map[string]JBHEval{
			"子": {Level: "大吉", Keyword: "癸水归源", Desc: "子水正位，木得水润，枯木逢春。"},
		},
	},
	"甲_午": {
		Verse:          "甲木午月火炎天，专取癸水为根源。运行北方多如意，南方火地见忧煎。",
		GoodDirections: []string{"北方水"},
		BadDirections:  []string{"南方火"},
		SpecificZhi: map[string]JBHEval{
			"子": {Level: "大吉", Keyword: "水济火炎", Desc: "水火既济大吉，甲木得润生机。"},
			"午": {Level: "大凶", Keyword: "火炎焦木", Desc: "午火重逢木焦，恐有灾厄。"},
		},
	},
	"甲_未": {
		Verse:          "甲木未月土燥干，癸水调候最为先。运喜北方水润泽，西南火土多灾难。",
		GoodDirections: []string{"北方水"},
		BadDirections:  []string{"南方火"},
		SpecificZhi: map[string]JBHEval{
			"亥": {Level: "大吉", Keyword: "水润生木", Desc: "亥水生木，枯木逢甘霖，万物复苏。"},
		},
	},
	"甲_申": {
		Verse:          "甲木申月值七杀，顺行南地可荣华。逆行须防伤克重，北方水旺亦堪嗟。",
		GoodDirections: []string{"南方火"},
		BadDirections:  []string{"北方水", "西方金"},
		SpecificZhi: map[string]JBHEval{
			"午": {Level: "大吉", Keyword: "火炼秋金", Desc: "丁火制杀，化敌为友，武贵之命。"},
		},
	},
	"甲_酉": {
		Verse:          "甲木酉月正官临，丁火伤官制杀金。运到南方多富贵，水乡金地反沉沦。",
		GoodDirections: []string{"南方火"},
		BadDirections:  []string{"北方水", "西方金"},
		SpecificZhi: map[string]JBHEval{
			"午": {Level: "大吉", Keyword: "火制金精", Desc: "南方火地制金，木得自由，名利双收。"},
		},
	},
	"甲_戌": {
		Verse:          "甲木戌月土重重，庚金劈甲建奇功。运行金水多清贵，火土相逢反受穷。",
		GoodDirections: []string{"西方金", "北方水"},
		BadDirections:  []string{"南方火"},
		SpecificZhi: map[string]JBHEval{
			"申": {Level: "吉", Keyword: "金水相生", Desc: "申金发水源头，疏土润木，有功名。"},
		},
	},
	"甲_亥": {
		Verse:          "甲木亥月正逢生，火暖金裁可成名。运行南方发福禄，北方水地总飘零。",
		GoodDirections: []string{"南方火", "西方金"},
		BadDirections:  []string{"北方水"},
		SpecificZhi: map[string]JBHEval{
			"午": {Level: "大吉", Keyword: "寒木向阳", Desc: "南火暖局，甲木得生机，鸿运当头。"},
		},
	},

	// ==================
	// 乙木
	// ==================
	"乙_子": {
		Verse:          "乙木子月水寒凝，丙火当先暖气生。运行南方多得意，水乡金地少光明。",
		GoodDirections: []string{"南方火"},
		BadDirections:  []string{"北方水", "西方金"},
	},
	"乙_丑": {
		Verse:          "乙木丑月寒冰冻，丙火解冻是正宗。南方运地春风暖，北地金乡黯淡中。",
		GoodDirections: []string{"南方火"},
		BadDirections:  []string{"北方水", "西方金"},
	},
	"乙_寅": {
		Verse:          "乙木寅月得生扶，丙火吐秀美如图。行运南方财官旺，北行水地反模糊。",
		GoodDirections: []string{"南方火"},
		BadDirections:  []string{"北方水"},
	},
	"乙_卯": {
		Verse:          "乙木卯月自身强，喜得金来削木方。运走西方逢富贵，东行太旺反招殃。",
		GoodDirections: []string{"西方金"},
		BadDirections:  []string{"东方木"},
	},
	"乙_辰": {
		Verse:          "乙木辰月土旺崇，癸水疏通丙火融。运行水火多亨泰，纯土金乡不称雄。",
		GoodDirections: []string{"北方水", "南方火"},
		BadDirections:  []string{"西方金"},
	},
	"乙_巳": {
		Verse:          "乙木巳月火正炎，癸水调候不可偏。运行北方多顺遂，南方火旺反忧煎。",
		GoodDirections: []string{"北方水"},
		BadDirections:  []string{"南方火"},
	},
	"乙_午": {
		Verse:          "乙木午月火炎炎，专用癸水作甘泉。运行北地多如意，再走南方苦万千。",
		GoodDirections: []string{"北方水"},
		BadDirections:  []string{"南方火"},
	},
	"乙_未": {
		Verse:          "乙木未月土又燥，癸水润泽最为好。运行北方水如意，火土运中多烦恼。",
		GoodDirections: []string{"北方水"},
		BadDirections:  []string{"南方火"},
	},
	"乙_申": {
		Verse:          "乙木申月杀星强，丙火制杀保安康。运走南方多吉利，北方水重反凄凉。",
		GoodDirections: []string{"南方火"},
		BadDirections:  []string{"北方水"},
	},
	"乙_酉": {
		Verse:          "乙木酉月正官星，癸水化煞丙暖明。南方运地多富贵，金水重逢命不亨。",
		GoodDirections: []string{"南方火"},
		BadDirections:  []string{"西方金"},
	},
	"乙_戌": {
		Verse:          "乙木戌月土重埋，癸水辛金并用来。运走金水多清贵，火土运中命不开。",
		GoodDirections: []string{"北方水", "西方金"},
		BadDirections:  []string{"南方火"},
	},
	"乙_亥": {
		Verse:          "乙木亥提水气深，丙火当先暖为真。运走南方多发达，水乡金地反沉沦。",
		GoodDirections: []string{"南方火"},
		BadDirections:  []string{"北方水"},
	},

	// ==================
	// 丙火
	// ==================
	"丙_子": {
		Verse:          "丙火子月水当权，壬水重重要土坚。运行南方多发达，北方水地总忧煎。",
		GoodDirections: []string{"南方火", "东方木"},
		BadDirections:  []string{"北方水"},
	},
	"丙_丑": {
		Verse:          "丙火丑月寒凝重，甲木生火壬水用。运行东南多亨通，西北金水反无功。",
		GoodDirections: []string{"东方木", "南方火"},
		BadDirections:  []string{"西方金", "北方水"},
	},
	"丙_寅": {
		Verse:          "丙火寅月木来生，壬水偏财格自成。运行北方财如海，南方火地反无情。",
		GoodDirections: []string{"北方水"},
		BadDirections:  []string{"南方火"},
	},
	"丙_卯": {
		Verse:          "丙火卯月印绶真，壬水偏财喜气临。运行北方多财禄，火乡木地反崎嵚。",
		GoodDirections: []string{"北方水"},
		BadDirections:  []string{"东方木"},
	},
	"丙_辰": {
		Verse:          "丙火辰月土伤官，壬水偏财喜相看。运走北方多快意，纯土运中反不欢。",
		GoodDirections: []string{"北方水"},
		BadDirections:  []string{"南方火"},
	},
	"丙_巳": {
		Verse:          "丙火巳月自临官，壬水制火最为先。运行北方源流远，南方火地反成灾。",
		GoodDirections: []string{"北方水"},
		BadDirections:  []string{"南方火"},
	},
	"丙_午": {
		Verse:          "丙火午月阳刃强，壬水庚金缺不良。运行西北多如意，南方东地步凄凉。",
		GoodDirections: []string{"西方金", "北方水"},
		BadDirections:  []string{"南方火"},
	},
	"丙_未": {
		Verse:          "丙火未月土燥干，壬水庚金两不闲。运走北方多发达，南方火旺反多难。",
		GoodDirections: []string{"北方水", "西方金"},
		BadDirections:  []string{"南方火"},
	},
	"丙_申": {
		Verse:          "丙火申月偏财临，戊土制水壬为真。运行土火多亨泰，金水运来反不宁。",
		GoodDirections: []string{"南方火"},
		BadDirections:  []string{"北方水"},
	},
	"丙_酉": {
		Verse:          "丙火酉月正财星，壬水相辉映碧空。运行东南多富贵，西北金水反飘零。",
		GoodDirections: []string{"东方木", "南方火"},
		BadDirections:  []string{"北方水"},
	},
	"丙_戌": {
		Verse:          "丙火戌月入火库，甲木疏通壬水辅。运行东方木火明，西北金水成碍阻。",
		GoodDirections: []string{"东方木"},
		BadDirections:  []string{"西方金", "北方水"},
	},
	"丙_亥": {
		Verse:          "丙火亥月正财乡，甲木生身火自强。运行东南多快意，西北金寒不相当。",
		GoodDirections: []string{"东方木", "南方火"},
		BadDirections:  []string{"西方金", "北方水"},
	},

	// ==================
	// 丁火
	// ==================
	"丁_子": {
		Verse:          "丁火子月甲木功，庚金劈甲引丁红。运行东方林木旺，水多北地反成空。",
		GoodDirections: []string{"东方木"},
		BadDirections:  []string{"北方水"},
	},
	"丁_丑": {
		Verse:          "丁火丑月寒极天，甲木庚金并用全。运行东南多暖意，西北金水苦难言。",
		GoodDirections: []string{"东方木", "南方火"},
		BadDirections:  []string{"西方金", "北方水"},
	},
	"丁_寅": {
		Verse:          "丁火寅月得长生，庚金劈甲引丁明。运行西方多得利，北方水旺反伤情。",
		GoodDirections: []string{"西方金"},
		BadDirections:  []string{"北方水"},
	},
	"丁_卯": {
		Verse:          "丁火卯月印旺真，庚金去乙引丁神。运行西方金为用，东方木旺反伤身。",
		GoodDirections: []string{"西方金"},
		BadDirections:  []string{"东方木"},
	},
	"丁_辰": {
		Verse:          "丁火辰月伤官旺，甲木生身庚做伴。运行东西皆可用，北方水地防灾患。",
		GoodDirections: []string{"东方木", "西方金"},
		BadDirections:  []string{"北方水"},
	},
	"丁_巳": {
		Verse:          "丁火巳月火正旺，甲木庚金不可忘。运行西方多富贵，南方太旺反招殃。",
		GoodDirections: []string{"西方金"},
		BadDirections:  []string{"南方火"},
	},
	"丁_午": {
		Verse:          "丁火午月劫财多，壬水庚金缺不可。运行西北多顺遂，东南火旺祸偏多。",
		GoodDirections: []string{"西方金", "北方水"},
		BadDirections:  []string{"南方火"},
	},
	"丁_未": {
		Verse:          "丁火未月财星旺，甲木壬水两相当。运行东北方有利，西南火地步凄凉。",
		GoodDirections: []string{"东方木", "北方水"},
		BadDirections:  []string{"南方火"},
	},
	"丁_申": {
		Verse:          "丁火申月财官全，甲木生身庚做权。运行东方林荫好，北方水泛反忧煎。",
		GoodDirections: []string{"东方木"},
		BadDirections:  []string{"北方水"},
	},
	"丁_酉": {
		Verse:          "丁火酉月正财星，甲木庚金互用精。运行东方多吉庆，北方水旺反多惊。",
		GoodDirections: []string{"东方木"},
		BadDirections:  []string{"北方水"},
	},
	"丁_戌": {
		Verse:          "丁火戌月火归库，甲木疏通庚金辅。运行东方发长久，北方水地成碍阻。",
		GoodDirections: []string{"东方木"},
		BadDirections:  []string{"北方水"},
	},
	"丁_亥": {
		Verse:          "丁火亥月正官临，甲木庚金两用心。运行东方多发达，北方水重反沉沦。",
		GoodDirections: []string{"东方木"},
		BadDirections:  []string{"北方水"},
	},

	// ==================
	// 戊土
	// ==================
	"戊_子": {
		Verse:          "戊土子月财星旺，丙火甲木两不忘。运行南方多发达，北方水重反凄凉。",
		GoodDirections: []string{"南方火"},
		BadDirections:  []string{"北方水"},
	},
	"戊_丑": {
		Verse:          "戊土丑月比肩强，丙火甲木正相当。运行南方多亨通，北方水地反不良。",
		GoodDirections: []string{"南方火", "东方木"},
		BadDirections:  []string{"北方水"},
	},
	"戊_寅": {
		Verse:          "戊土寅月杀星强，丙火暖身甲木当。运行南方多发越，西方金地反凄凉。",
		GoodDirections: []string{"南方火"},
		BadDirections:  []string{"西方金"},
	},
	"戊_卯": {
		Verse:          "戊土卯月官星临，丙火甲木喜气新。运走南方多顺遂，北方水旺反伤身。",
		GoodDirections: []string{"南方火"},
		BadDirections:  []string{"北方水"},
	},
	"戊_辰": {
		Verse:          "戊土辰月比肩重，甲木疏之丙火用。运行南方多快意，纯土运中反无功。",
		GoodDirections: []string{"南方火", "东方木"},
		BadDirections:  []string{"北方水"},
	},
	"戊_巳": {
		Verse:          "戊土巳月印绶生，甲木疏通癸水清。运行北方多财禄，火土运中命不亨。",
		GoodDirections: []string{"北方水", "东方木"},
		BadDirections:  []string{"南方火"},
	},
	"戊_午": {
		Verse:          "戊土午月火极炎，壬水甲木两般全。运行北方多得利，南方火旺反成灾。",
		GoodDirections: []string{"北方水"},
		BadDirections:  []string{"南方火"},
	},
	"戊_未": {
		Verse:          "戊土未月燥极天，癸水丙火要当先。运行北方多亨泰，南方火旺反忧煎。",
		GoodDirections: []string{"北方水"},
		BadDirections:  []string{"南方火"},
	},
	"戊_申": {
		Verse:          "戊土申月食神明，丙火生身癸水清。运行南方多发达，北方水重反飘零。",
		GoodDirections: []string{"南方火"},
		BadDirections:  []string{"北方水"},
	},
	"戊_酉": {
		Verse:          "戊土酉月伤官旺，丙火癸水不可忘。运行南方多快意，金水运中命不强。",
		GoodDirections: []string{"南方火"},
		BadDirections:  []string{"北方水"},
	},
	"戊_戌": {
		Verse:          "戊土戌月当权令，甲木丙火两相称。运行东南多发越，纯土运中不称情。",
		GoodDirections: []string{"东方木", "南方火"},
		BadDirections:  []string{"北方水"},
	},
	"戊_亥": {
		Verse:          "戊土亥月财星临，甲木丙火保安宁。运行南方多发达，北方水泛反飘零。",
		GoodDirections: []string{"南方火", "东方木"},
		BadDirections:  []string{"北方水"},
	},

	// ==================
	// 己土
	// ==================
	"己_子": {
		Verse:          "己土子月正财明，丙火甲木两相迎。运行南方多顺遂，北方水重反不宁。",
		GoodDirections: []string{"南方火"},
		BadDirections:  []string{"北方水"},
	},
	"己_丑": {
		Verse:          "己土丑月比肩扶，丙火暖身甲木疏。运行南方多如意，北方水地少欢娱。",
		GoodDirections: []string{"南方火"},
		BadDirections:  []string{"北方水"},
	},
	"己_寅": {
		Verse:          "己土寅月官杀旺，丙火庚金正相当。运行南方多福禄，北方水旺反凄凉。",
		GoodDirections: []string{"南方火"},
		BadDirections:  []string{"北方水"},
	},
	"己_卯": {
		Verse:          "己土卯月正官星，甲木癸水丙火称。运行南方多发达，水乡金地反伤情。",
		GoodDirections: []string{"南方火"},
		BadDirections:  []string{"北方水", "西方金"},
	},
	"己_辰": {
		Verse:          "己土辰月比肩强，丙火癸水甲木忙。运行南方多亨泰，纯土运中反不良。",
		GoodDirections: []string{"南方火"},
		BadDirections:  []string{"北方水"},
	},
	"己_巳": {
		Verse:          "己土巳月印星隆，癸水丙火两兼通。运行北方多财禄，南方火旺反不丰。",
		GoodDirections: []string{"北方水"},
		BadDirections:  []string{"南方火"},
	},
	"己_午": {
		Verse:          "己土午月火炎天，癸水丙火辛金全。运行北方多亨通，南方火旺苦万千。",
		GoodDirections: []string{"北方水", "西方金"},
		BadDirections:  []string{"南方火"},
	},
	"己_未": {
		Verse:          "己土未月当权令，癸水丙火必须用。运行北方多发越，火土运中反无功。",
		GoodDirections: []string{"北方水"},
		BadDirections:  []string{"南方火"},
	},
	"己_申": {
		Verse:          "己土申月食神旺，丙火癸水不可忘。运行南方多发达，北方水重难安康。",
		GoodDirections: []string{"南方火"},
		BadDirections:  []string{"北方水"},
	},
	"己_酉": {
		Verse:          "己土酉月伤官明，丙火癸水互相迎。运行南方多快意，北方水旺反飘零。",
		GoodDirections: []string{"南方火"},
		BadDirections:  []string{"北方水"},
	},
	"己_戌": {
		Verse:          "己土戌月比肩重，甲木丙火癸水用。运行东南多亨通，纯土运中反无功。",
		GoodDirections: []string{"东方木", "南方火"},
		BadDirections:  []string{"北方水"},
	},
	"己_亥": {
		Verse:          "己土亥月正财星，丙火甲木保安宁。运行南方多发达，北方水泛总飘零。",
		GoodDirections: []string{"南方火"},
		BadDirections:  []string{"北方水"},
	},

	// ==================
	// 庚金
	// ==================
	"庚_子": {
		Verse:          "庚金子月伤官旺，丁火甲木不可忘。运行南方多发达，北方水冷反凄凉。",
		GoodDirections: []string{"南方火"},
		BadDirections:  []string{"北方水"},
	},
	"庚_丑": {
		Verse:          "庚金丑月土生金，丙丁并用去寒侵。运行南方多快意，北方水重反沉沦。",
		GoodDirections: []string{"南方火"},
		BadDirections:  []string{"北方水"},
	},
	"庚_寅": {
		Verse:          "庚金寅月财官显，甲木壬水丙火全。运行南方多得利，北方水泛反不然。",
		GoodDirections: []string{"南方火"},
		BadDirections:  []string{"北方水"},
	},
	"庚_卯": {
		Verse:          "庚金卯月正财真，丁火甲庚配合神。运行南方多财禄，北方水旺反伤身。",
		GoodDirections: []string{"南方火"},
		BadDirections:  []string{"北方水"},
	},
	"庚_辰": {
		Verse:          "庚金辰月印绶生，甲木丁火壬水清。运行东南多亨泰，西北金水不称情。",
		GoodDirections: []string{"东方木", "南方火"},
		BadDirections:  []string{"北方水"},
	},
	"庚_巳": {
		Verse:          "庚金巳月正官方，壬水戊土配合良。运行北方多发达，南方火旺反招殃。",
		GoodDirections: []string{"北方水"},
		BadDirections:  []string{"南方火"},
	},
	"庚_午": {
		Verse:          "庚金午月杀星强，壬癸制火保安康。运行北方多顺遂，南方火地反忧伤。",
		GoodDirections: []string{"北方水"},
		BadDirections:  []string{"南方火"},
	},
	"庚_未": {
		Verse:          "庚金未月三伏天，丁火甲木两相全。运行东方多发越，南方火旺苦难言。",
		GoodDirections: []string{"东方木"},
		BadDirections:  []string{"南方火"},
	},
	"庚_申": {
		Verse:          "庚金申月建禄旺，丁火甲木不可忘。运行南方多富贵，北方水旺反凄凉。",
		GoodDirections: []string{"南方火"},
		BadDirections:  []string{"北方水"},
	},
	"庚_酉": {
		Verse:          "庚金酉月阳刃雄，丁火甲丙并用功。运行南方多发达，北方水冷命不通。",
		GoodDirections: []string{"南方火"},
		BadDirections:  []string{"北方水"},
	},
	"庚_戌": {
		Verse:          "庚金戌月土重重，甲木壬水两相通。运行东方多亨泰，火土运中反无功。",
		GoodDirections: []string{"东方木", "北方水"},
		BadDirections:  []string{"南方火"},
	},
	"庚_亥": {
		Verse:          "庚金亥月伤官临，丁丙火暖不可禁。运行南方多如意，北方水重反沉沦。",
		GoodDirections: []string{"南方火"},
		BadDirections:  []string{"北方水"},
	},

	// ==================
	// 辛金
	// ==================
	"辛_子": {
		Verse:          "辛金子月食神旺，丙火壬水两相当。运行南方多发达，北方水冷反凄凉。",
		GoodDirections: []string{"南方火"},
		BadDirections:  []string{"北方水"},
	},
	"辛_丑": {
		Verse:          "辛金丑月土生身，丙火暖局壬水润。运行南方多快意，北方水重反伤神。",
		GoodDirections: []string{"南方火"},
		BadDirections:  []string{"北方水"},
	},
	"辛_寅": {
		Verse:          "辛金寅月财官显，己壬庚配合有缘。运行南方多发达，北方水泛反忧煎。",
		GoodDirections: []string{"南方火"},
		BadDirections:  []string{"北方水"},
	},
	"辛_卯": {
		Verse:          "辛金卯月偏财旺，壬水甲木两相当。运行北方多亨泰，东方木地反不良。",
		GoodDirections: []string{"北方水"},
		BadDirections:  []string{"东方木"},
	},
	"辛_辰": {
		Verse:          "辛金辰月印旺崇，壬水甲木要相逢。运行北方多清贵，火土运中反不丰。",
		GoodDirections: []string{"北方水"},
		BadDirections:  []string{"南方火"},
	},
	"辛_巳": {
		Verse:          "辛金巳月正官方，壬水甲木保安康。运行北方多发达，南方火旺反招殃。",
		GoodDirections: []string{"北方水"},
		BadDirections:  []string{"南方火"},
	},
	"辛_午": {
		Verse:          "辛金午月杀星强，壬己癸水不可忘。运行北方多吉利，南方火重反凄凉。",
		GoodDirections: []string{"北方水"},
		BadDirections:  []string{"南方火"},
	},
	"辛_未": {
		Verse:          "辛金未月土燥干，壬庚甲木正相关。运行北方多发越，火土运中命不安。",
		GoodDirections: []string{"北方水"},
		BadDirections:  []string{"南方火"},
	},
	"辛_申": {
		Verse:          "辛金申月身旺强，壬水甲木正相当。运行东方多亨泰，北方水旺亦可量。",
		GoodDirections: []string{"东方木", "北方水"},
		BadDirections:  []string{"西方金"},
	},
	"辛_酉": {
		Verse:          "辛金酉月建禄旺，壬水甲木不可忘。运行东方多富贵，纯金运中反不良。",
		GoodDirections: []string{"东方木", "北方水"},
		BadDirections:  []string{"西方金"},
	},
	"辛_戌": {
		Verse:          "辛金戌月土重重，壬水甲木两相通。运行北方多清贵，土重金埋反不丰。",
		GoodDirections: []string{"北方水", "东方木"},
		BadDirections:  []string{"南方火"},
	},
	"辛_亥": {
		Verse:          "辛金亥月食神良，壬水丙火要相当。运行南方多快意，北方水冷反凄凉。",
		GoodDirections: []string{"南方火"},
		BadDirections:  []string{"北方水"},
	},

	// ==================
	// 壬水
	// ==================
	"壬_子": {
		Verse:          "壬水子月阳刃旺，戊土丙火不可忘。运行南方多发达，北方水重反凄凉。",
		GoodDirections: []string{"南方火"},
		BadDirections:  []string{"北方水"},
	},
	"壬_丑": {
		Verse:          "壬水丑月杂气生，丙火甲木两相迎。运行南方多亨通，北方水旺反飘零。",
		GoodDirections: []string{"南方火", "东方木"},
		BadDirections:  []string{"北方水"},
	},
	"壬_寅": {
		Verse:          "壬水寅月食神明，庚丙戊土配合精。运行南方多发越，北方水泛反无情。",
		GoodDirections: []string{"南方火"},
		BadDirections:  []string{"北方水"},
	},
	"壬_卯": {
		Verse:          "壬水卯月伤官旺，戊辛庚金不可忘。运行西方多得利，东方木旺反招殃。",
		GoodDirections: []string{"西方金"},
		BadDirections:  []string{"东方木"},
	},
	"壬_辰": {
		Verse:          "壬水辰月有库根，甲木庚金两用真。运行东方多亨泰，北方水泛反不伸。",
		GoodDirections: []string{"东方木"},
		BadDirections:  []string{"北方水"},
	},
	"壬_巳": {
		Verse:          "壬水巳月财星旺，辛庚癸水互相帮。运行西方多发达，南方火地反忧伤。",
		GoodDirections: []string{"西方金"},
		BadDirections:  []string{"南方火"},
	},
	"壬_午": {
		Verse:          "壬水午月正财星，癸庚辛金水最清。运行西方多富贵，南方火旺反飘零。",
		GoodDirections: []string{"西方金"},
		BadDirections:  []string{"南方火"},
	},
	"壬_未": {
		Verse:          "壬水未月土燥干，辛金甲木正相关。运行西方多亨通，火土运中反不安。",
		GoodDirections: []string{"西方金"},
		BadDirections:  []string{"南方火"},
	},
	"壬_申": {
		Verse:          "壬水申月印星隆，戊土丁火两兼通。运行南方多快意，北方金水反不丰。",
		GoodDirections: []string{"南方火"},
		BadDirections:  []string{"北方水"},
	},
	"壬_酉": {
		Verse:          "壬水酉月印绶真，甲庚并用见精神。运行东方多亨泰，西方金重反不伸。",
		GoodDirections: []string{"东方木"},
		BadDirections:  []string{"西方金"},
	},
	"壬_戌": {
		Verse:          "壬水戌月财官临，甲木丙火正相亲。运行东方多发达，北方水重反沉沦。",
		GoodDirections: []string{"东方木", "南方火"},
		BadDirections:  []string{"北方水"},
	},
	"壬_亥": {
		Verse:          "壬水亥月建禄强，戊土丙火用两当。运行南方多发越，北方水泛反凄凉。",
		GoodDirections: []string{"南方火"},
		BadDirections:  []string{"北方水"},
	},

	// ==================
	// 癸水
	// ==================
	"癸_子": {
		Verse:          "癸水子月劫财多，戊土丙火不可缺。运行南方多亨通，北方水重反成祸。",
		GoodDirections: []string{"南方火"},
		BadDirections:  []string{"北方水"},
	},
	"癸_丑": {
		Verse:          "癸水丑月杂气库，丙丁火暖不可负。运行南方多亨泰，北方水地反无助。",
		GoodDirections: []string{"南方火"},
		BadDirections:  []string{"北方水"},
	},
	"癸_寅": {
		Verse:          "癸水寅月伤官旺，辛丙配合最相当。运行南方多得利，北方水旺反凄凉。",
		GoodDirections: []string{"南方火", "西方金"},
		BadDirections:  []string{"北方水"},
	},
	"癸_卯": {
		Verse:          "癸水卯月食神清，庚辛发水不可停。运行西方多富贵，东方木旺反不宁。",
		GoodDirections: []string{"西方金"},
		BadDirections:  []string{"东方木"},
	},
	"癸_辰": {
		Verse:          "癸水辰月有库根，丙辛甲木并用真。运行南方多快意，北方水泛总伤身。",
		GoodDirections: []string{"南方火"},
		BadDirections:  []string{"北方水"},
	},
	"癸_巳": {
		Verse:          "癸水巳月正财隆，辛庚发水保安宁。运行西方多发达，南方火旺反成凶。",
		GoodDirections: []string{"西方金"},
		BadDirections:  []string{"南方火"},
	},
	"癸_午": {
		Verse:          "癸水午月正财星，庚辛壬癸配合精。运行西方多亨泰，南方火旺反飘零。",
		GoodDirections: []string{"西方金"},
		BadDirections:  []string{"南方火"},
	},
	"癸_未": {
		Verse:          "癸水未月土燥干，庚辛壬癸正相关。运行西方多发越，火土运中苦万般。",
		GoodDirections: []string{"西方金"},
		BadDirections:  []string{"南方火"},
	},
	"癸_申": {
		Verse:          "癸水申月印绶强，丁甲配合火暖香。运行南方多富贵，北方金水反凄凉。",
		GoodDirections: []string{"南方火", "东方木"},
		BadDirections:  []string{"北方水"},
	},
	"癸_酉": {
		Verse:          "癸水酉月正印星，辛丙配合最称情。运行南方多发达，西方金旺反不宁。",
		GoodDirections: []string{"南方火"},
		BadDirections:  []string{"西方金"},
	},
	"癸_戌": {
		Verse:          "癸水戌月官杀临，辛甲壬癸保安身。运行西方多清贵，火土运中反不伸。",
		GoodDirections: []string{"西方金"},
		BadDirections:  []string{"南方火"},
	},
	"癸_亥": {
		Verse:          "癸水亥月建禄强，庚辛戊丁用两当。运行南方多发越，北方水泛反凄凉。",
		GoodDirections: []string{"南方火"},
		BadDirections:  []string{"北方水"},
	},
}
