// Package bazi - 命格（格局）推算引擎
// 基于命师 6 规则取格（月禄/月刃 → 候选 → 食伤同透 → 通根强度 → 地支气势 → 杂气兜底）
// Spec: docs/superpowers/specs/2026-05-17-mingge-professional-rules-design.md
package bazi

// 地支藏干表（按主气→中气→余气排列，主气在首位）
var zhiHideGanFull = map[string][]string{
	"子": {"癸"},
	"丑": {"己", "癸", "辛"},
	"寅": {"甲", "丙", "戊"},
	"卯": {"乙"},
	"辰": {"戊", "乙", "癸"},
	"巳": {"丙", "庚", "戊"},
	"午": {"丁", "己"},
	"未": {"己", "丁", "乙"},
	"申": {"庚", "壬", "戊"},
	"酉": {"辛"},
	"戌": {"戊", "辛", "丁"},
	"亥": {"壬", "甲"},
}

// 三合局：地支三合五行
// 寅午戌→火，巳酉丑→金，申子辰→水，亥卯未→木
var sanHeJu = []struct {
	zhis   [3]string
	wuxing string
}{
	{[3]string{"寅", "午", "戌"}, "火"},
	{[3]string{"巳", "酉", "丑"}, "金"},
	{[3]string{"申", "子", "辰"}, "水"},
	{[3]string{"亥", "卯", "未"}, "木"},
}

// 三会局：地支三会五行
// 寅卯辰→木，巳午未→火，申酉戌→金，亥子丑→水
var sanHuiJu = []struct {
	zhis   [3]string
	wuxing string
}{
	{[3]string{"寅", "卯", "辰"}, "木"},
	{[3]string{"巳", "午", "未"}, "火"},
	{[3]string{"申", "酉", "戌"}, "金"},
	{[3]string{"亥", "子", "丑"}, "水"},
}

// wuxingMainGan 五行对应的阳干代表（用于合局取格神）
var wuxingMainGan = map[string]string{
	"木": "甲",
	"火": "丙",
	"土": "戊",
	"金": "庚",
	"水": "壬",
}

// wuxingKe 五行相克：木→土、土→水、水→火、火→金、金→木
var wuxingKe = map[string]string{
	"木": "土",
	"土": "水",
	"水": "火",
	"火": "金",
	"金": "木",
}

// isKeWuxing 判断 attacker 五行是否克 defender 五行
func isKeWuxing(attacker, defender string) bool {
	return wuxingKe[attacker] == defender
}

// linGuanZhi 日干 → 临官地支（月禄格判定用）
var linGuanZhi = map[string]string{
	"甲": "寅",
	"乙": "卯",
	"丙": "巳",
	"丁": "午",
	"戊": "巳",
	"己": "午",
	"庚": "申",
	"辛": "酉",
	"壬": "亥",
	"癸": "子",
}

// diWangZhi 日干 → 帝旺地支（月刃格判定用，仅阳干）
var diWangZhi = map[string]string{
	"甲": "卯",
	"丙": "午",
	"戊": "午",
	"庚": "酉",
	"壬": "子",
}

// yangGans 阳干集合（月刃格判定用）
var yangGans = map[string]bool{
	"甲": true,
	"丙": true,
	"戊": true,
	"庚": true,
	"壬": true,
}

// ganWuxing 天干→五行
var ganWuxingMap = map[string]string{
	"甲": "木", "乙": "木",
	"丙": "火", "丁": "火",
	"戊": "土", "己": "土",
	"庚": "金", "辛": "金",
	"壬": "水", "癸": "水",
}

// 格局说明文字字典
var minggeDescDict = map[string]string{
	"正官格": "正官格是八字中最受推崇的格局之一。月令透出正官，代表命主为人正直守规、重视名誉与责任，凡事依法循理，不走偏途。正官格之人仕途有望，适合在体制内或正规企业发展，易得领导赏识与贵人扶持。日主身强者，官星有力，功名显达；日主身弱需印星护持，方能承受官星之压。",
	"七杀格": "七杀格气势凌厉，月令透出七杀，代表命主天生具备魄力与果断，行事干练强悍，敢于迎难而上。七杀格之人适合军警、外科、竞技、创业等高压领域，能独当一面。七杀格需有制化，以食神制杀最佳，称食神制杀格，否则杀气太重，容易招惹是非或内心压力极大。日主强旺能驾驭七杀，则一生奋斗有成。",
	"食神格": "食神格是八字中最温和吉祥的格局之一。月令透出食神，代表命主聪明机智、富有才艺与口福，为人仁慈宽厚，善于享受生活。食神格之人往往多才多艺，适合餐饮、艺术、教育、写作等领域。食神格不喜见枭神（偏印）夺食，否则才艺受损、运势起伏。整体而言，食神格命局温润，衣食无忧，人缘极佳。",
	"伤官格": "伤官格是八字中最具个性的格局。月令透出伤官，代表命主才华横溢、思维活跃，具有超强的创造力与表达欲，但也往往不服管束、傲气十足。伤官格之人适合艺术、科技、法律、演艺等需要个人风格的领域。伤官见官（正官）在命中会产生冲突，轻则影响仕途，重则招惹官非，需格外留意。伤官配印，则才华得以发挥，成就斐然。",
	"正财格": "正财格月令透出正财，代表命主勤俭务实、踏实肯干，重视物质积累，凡事稳扎稳打。正财格之人一生财源稳定，不求横财暴富，靠双手积累财富，适合实业、金融、会计等稳健型行业。正财格命主对感情亦认真，婚姻较为稳固。身强财旺者，富贵双全；身弱财重则为财所累，需以比劫帮扶。",
	"偏财格": "偏财格月令透出偏财，代表命主广结人缘、善于交际，财来财往、左右逢源，具有天生的经商嗅觉。偏财格之人适合商业、销售、投资、娱乐等流动性强的领域，有机会获得意外之财或横财。偏财格也主与父亲缘分特殊。身强财旺者，财运亨通；身弱偏财重则易破财散财，需量力而行。",
	"正印格": "正印格月令透出正印，代表命主仁慈善良、为人厚道，学识丰富，重视精神修养与文化积累。正印格之人适合学术、教育、医疗、宗教等领域，易得长辈、师长的庇护与提携。正印格之人内心稳重，处变不惊。正印格最忌财星破印，若命中财旺克印，则学业受阻，贵人缘薄；印星有力者，名誉与才学俱佳。",
	"偏印格": "偏印格月令透出偏印（枭神），代表命主思维独特、聪慧多思，常有异于常人的见解与创意，偏向研究型或艺术型方向。偏印格之人适合研究、策划、哲学、艺术、心理学等领域，但有时想法过多、执行力不足。偏印格最忌夺食（克食神），若命中食神被偏印所克，则口福受损、才艺难展。合理配置用神，则灵思泉涌，自成一派。",
	"建禄格": "建禄格即月令透出比肩，日主之气聚于月令，身强有力。建禄格之人自立心强，独立自主，做事靠自身努力，不依赖他人。建禄格命主意志坚定，韧性十足，适合自主创业或独立执业。建禄格不喜比劫过多争财，需有官杀来疏导旺气。身强用官杀或财星，则功名利禄皆可期；整体而言是奋斗型命格，努力必有回报。",
	"月禄格": "月禄格即月支临官，日主之气聚于月令，身强有力。月禄格之人自立心强，独立自主，做事靠自身努力，不依赖他人。月禄格命主意志坚定，韧性十足，适合自主创业或独立执业。月禄格不喜比劫过多争财，需有官杀来疏导旺气。身强用官杀或财星，则功名利禄皆可期；整体而言是奋斗型命格，努力必有回报。",
	"月刃格": "月刃格即月令透出劫财，日主之气极旺，有锋芒毕露之势。月刃格之人性格刚烈、争强好胜，凡事力争上游，具有极强的竞争意识和执行力。月刃格需有官杀制刃，方能化刚为用，成就一番事业；无制则容易冲动行事，招惹纷争。月刃格适合武职、竞技、外科、法律等需要魄力与决断的领域，是一种需要驾驭的强力命格。",
	"杂气格": "此命局四柱透干关系较为复杂，未能归入月令正格体系。格局以杂气论之，命局吉凶需结合用神、大运综合研判，不可单以格局定论。建议参考调候用神与五行分布，以喜用之五行为行运依据，综合判断人生走势。",
}

// shiShenToGeName 十神 → 格名（含建禄/月刃特殊映射）
func shiShenToGeName(shishen string) string {
	switch shishen {
	case "比肩":
		return "建禄格"
	case "劫财":
		return "月刃格"
	default:
		return shishen + "格"
	}
}

// detectSanHeHui 检测四柱地支是否构成三合局或三会局
// 返回合局五行（木/火/土/金/水），未命中返回空串
func detectSanHeHui(zhis []string) string {
	zhiSet := make(map[string]bool)
	for _, z := range zhis {
		zhiSet[z] = true
	}
	// 三合优先
	for _, he := range sanHeJu {
		if zhiSet[he.zhis[0]] && zhiSet[he.zhis[1]] && zhiSet[he.zhis[2]] {
			return he.wuxing
		}
	}
	// 三会次之
	for _, hui := range sanHuiJu {
		if zhiSet[hui.zhis[0]] && zhiSet[hui.zhis[1]] && zhiSet[hui.zhis[2]] {
			return hui.wuxing
		}
	}
	return ""
}

// DetectMingGe 按命师 6 规则取格
//
// 规则优先级：
//   0. 月禄/月刃 special case
//   1. 收集天干非比劫候选
//   4. 食伤同透 → 强制立伤官格
//   3. 通根强度排序 → 取最强者立格
//   6. 候选都无根 / 无候选 → 地支气势全土且为日干财 → 立财格
//   5. 兜底 → 杂气格
//
// Spec: docs/superpowers/specs/2026-05-17-mingge-professional-rules-design.md
func DetectMingGe(r *BaziResult) (name, desc string) {
	dayGan := r.DayGan

	// ── 规则 0: 月禄/月刃 special case ───────────────────────
	if linGuanZhi[dayGan] == r.MonthZhi {
		return "月禄格", minggeDescDict["月禄格"]
	}
	if yangGans[dayGan] && diWangZhi[dayGan] == r.MonthZhi {
		return "月刃格", minggeDescDict["月刃格"]
	}

	// ── 规则 1: 收集天干非比劫候选 ─────────────────────────
	type cand struct {
		gan     string
		shishen string
	}
	var candidates []cand
	for _, g := range []string{r.YearGan, r.MonthGan, r.HourGan} {
		ss := GetShiShen(dayGan, g)
		if ss == "比肩" || ss == "劫财" || ss == "" {
			continue
		}
		candidates = append(candidates, cand{g, ss})
	}

	// ── 规则 4: 食伤同透 → 强制立伤官格 ────────────────────
	hasFood, hasInjury := false, false
	for _, c := range candidates {
		if c.shishen == "食神" {
			hasFood = true
		}
		if c.shishen == "伤官" {
			hasInjury = true
		}
	}
	if hasFood && hasInjury {
		return "伤官格", minggeDescDict["伤官格"]
	}

	// ── 规则 3: 通根强度排序，取最强 ──────────────────────
	if len(candidates) > 0 {
		bestIdx := 0
		bestRoot := rootStrength(candidates[0].gan, r)
		for i := 1; i < len(candidates); i++ {
			root := rootStrength(candidates[i].gan, r)
			if root > bestRoot {
				bestRoot = root
				bestIdx = i
			}
		}
		if bestRoot > 0 {
			geName := shiShenToGeName(candidates[bestIdx].shishen)
			return geName, minggeDescDict[geName]
		}
	}

	// ── 规则 6: 地支气势全土且为日干财 → 立财格 ──────────
	zhiMains := []string{
		zhiHideGanFull[r.YearZhi][0],
		zhiHideGanFull[r.MonthZhi][0],
		zhiHideGanFull[r.DayZhi][0],
		zhiHideGanFull[r.HourZhi][0],
	}
	firstWx := ganWuxingMap[zhiMains[0]]
	allSameWx := firstWx != ""
	for _, mg := range zhiMains[1:] {
		if ganWuxingMap[mg] != firstWx {
			allSameWx = false
			break
		}
	}
	if allSameWx {
		dayWx := ganWuxingMap[dayGan]
		if isKeWuxing(dayWx, firstWx) {
			monthMainGan := zhiHideGanFull[r.MonthZhi][0]
			ss := GetShiShen(dayGan, monthMainGan)
			geName := shiShenToGeName(ss)
			return geName, minggeDescDict[geName]
		}
	}

	// ── 规则 5: 兜底 ────────────────────────────────────
	return "杂气格", minggeDescDict["杂气格"]
}

// rootStrength 计算 gan 在 r 四柱地支藏干中能找到的最强单根分
//
// 评分表：
//   月支主气=6, 月支中气=5, 月支余气=4
//   他支主气=3, 他支中气=2, 他支余气=1
//   无根=0
//
// 返回 4 柱中找到的最大分。如果 gan 同时在月支和他支扎根，取月支根（分高）。
func rootStrength(gan string, r *BaziResult) int {
	posZhis := []struct {
		isMonth bool
		zhi     string
	}{
		{true, r.MonthZhi},
		{false, r.YearZhi},
		{false, r.DayZhi},
		{false, r.HourZhi},
	}

	best := 0
	for _, pz := range posZhis {
		hgs := zhiHideGanFull[pz.zhi]
		for i, hg := range hgs {
			if hg != gan {
				continue
			}
			var s int
			if pz.isMonth {
				s = 6 - i // 主气=6, 中气=5, 余气=4
			} else {
				s = 3 - i // 主气=3, 中气=2, 余气=1
			}
			if s > best {
				best = s
			}
		}
	}
	return best
}
