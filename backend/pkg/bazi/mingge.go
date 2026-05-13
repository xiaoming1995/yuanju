// Package bazi - 命格（格局）推算引擎
// 基于七优先级透干取格法
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

// gansContains 检查某天干是否存在于天干列表中
func gansContains(gans []string, target string) bool {
	for _, g := range gans {
		if g == target {
			return true
		}
	}
	return false
}

// countGanInGans 统计某天干在天干列表中出现次数
func countGanInGans(gans []string, target string) int {
	count := 0
	for _, g := range gans {
		if g == target {
			count++
		}
	}
	return count
}

// wuxingScore 统计某五行在天干列表中的出现次数（五行力量）
func wuxingScore(gans []string, wx string) int {
	count := 0
	for _, g := range gans {
		if ganWuxingMap[g] == wx {
			count++
		}
	}
	return count
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

// DetectMingGe 按七优先级透干取格，返回 (格名, 说明文字)
func DetectMingGe(r *BaziResult) (name, desc string) {
	dayGan := r.DayGan
	monthGan := r.MonthGan
	monthZhi := r.MonthZhi

	// 四柱天干集合（用于透出检测）
	allGans := []string{r.YearGan, r.MonthGan, r.DayGan, r.HourGan}
	otherGans := []string{r.YearGan, r.DayGan, r.HourGan} // 除月干外

	// 四柱地支
	allZhis := []string{r.YearZhi, r.MonthZhi, r.DayZhi, r.HourZhi}
	otherZhis := []string{r.YearZhi, r.DayZhi, r.HourZhi} // 除月支外

	// 月支藏干（主气在首）
	monthHideGans := zhiHideGanFull[monthZhi]
	if len(monthHideGans) == 0 {
		// 兜底：使用引擎已计算的藏干
		monthHideGans = r.MonthHideGan
	}

	// ── 优先级 1：月支主气透月干（月柱自透）────────────────────
	if len(monthHideGans) > 0 {
		mainGan := monthHideGans[0] // 主气
		if mainGan == monthGan {
			ss := GetShiShen(dayGan, mainGan)
			geName := shiShenToGeName(ss)
			return geName, minggeDescDict[geName]
		}
	}

	// ── 优先级 2：月支任意藏干透它干（年/日/时干）────────────────
	type candidate struct {
		gan   string
		count int
		wxCnt int
	}
	var p2Candidates []candidate
	for _, hg := range monthHideGans {
		if gansContains(otherGans, hg) {
			cnt := countGanInGans(allGans, hg)
			wxCnt := wuxingScore(allGans, ganWuxingMap[hg])
			p2Candidates = append(p2Candidates, candidate{hg, cnt, wxCnt})
		}
	}
	if len(p2Candidates) == 1 {
		ss := GetShiShen(dayGan, p2Candidates[0].gan)
		geName := shiShenToGeName(ss)
		return geName, minggeDescDict[geName]
	} else if len(p2Candidates) > 1 {
		// 多透取次数最多者，次数相同取五行力量最强者
		best := p2Candidates[0]
		for _, c := range p2Candidates[1:] {
			if c.count > best.count || (c.count == best.count && c.wxCnt > best.wxCnt) {
				best = c
			}
		}
		ss := GetShiShen(dayGan, best.gan)
		geName := shiShenToGeName(ss)
		return geName, minggeDescDict[geName]
	}

	// ── 优先级 3：其它三柱地支主气透月干 ────────────────────────
	for _, oz := range otherZhis {
		hgs := zhiHideGanFull[oz]
		if len(hgs) == 0 {
			continue
		}
		mainGan := hgs[0]
		if mainGan == monthGan {
			ss := GetShiShen(dayGan, mainGan)
			geName := shiShenToGeName(ss)
			return geName, minggeDescDict[geName]
		}
	}

	// ── 优先级 4：其它地支任意藏干透任意天干 ─────────────────────
	var p4Candidates []candidate
	seenGan := make(map[string]bool)
	for _, oz := range otherZhis {
		hgs := zhiHideGanFull[oz]
		for _, hg := range hgs {
			if seenGan[hg] {
				continue
			}
			if gansContains(allGans, hg) {
				seenGan[hg] = true
				cnt := countGanInGans(allGans, hg)
				wxCnt := wuxingScore(allGans, ganWuxingMap[hg])
				p4Candidates = append(p4Candidates, candidate{hg, cnt, wxCnt})
			}
		}
	}
	if len(p4Candidates) == 1 {
		ss := GetShiShen(dayGan, p4Candidates[0].gan)
		geName := shiShenToGeName(ss)
		return geName, minggeDescDict[geName]
	} else if len(p4Candidates) > 1 {
		best := p4Candidates[0]
		for _, c := range p4Candidates[1:] {
			if c.count > best.count || (c.count == best.count && c.wxCnt > best.wxCnt) {
				best = c
			}
		}
		ss := GetShiShen(dayGan, best.gan)
		geName := shiShenToGeName(ss)
		return geName, minggeDescDict[geName]
	}

	// ── 优先级 5：（已在优先级2、4中处理多透逻辑，此处不重复）────

	// ── 优先级 6：无透格神，检测三合/三会局 ─────────────────────
	if wx := detectSanHeHui(allZhis); wx != "" {
		if repGan, ok := wuxingMainGan[wx]; ok {
			ss := GetShiShen(dayGan, repGan)
			geName := shiShenToGeName(ss)
			return geName, minggeDescDict[geName]
		}
	}

	// ── 优先级 7：统计全局干支十神频率，满4个以上取最高频者 ────────
	// 统计所有天干 + 各地支主气中各十神出现次数
	allSources := append([]string{}, allGans...)
	for _, z := range allZhis {
		hgs := zhiHideGanFull[z]
		if len(hgs) > 0 {
			allSources = append(allSources, hgs[0]) // 主气
		}
	}
	shishenCount := make(map[string]int)
	for _, g := range allSources {
		ss := GetShiShen(dayGan, g)
		if ss != "" {
			shishenCount[ss]++
		}
	}
	bestSS := ""
	bestCnt := 0
	for ss, cnt := range shishenCount {
		if cnt >= 4 && cnt > bestCnt {
			bestCnt = cnt
			bestSS = ss
		}
	}
	if bestSS != "" {
		geName := shiShenToGeName(bestSS)
		return geName, minggeDescDict[geName]
	}

	// 兜底
	return "杂气格", minggeDescDict["杂气格"]
}
