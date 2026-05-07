package bazi

import "strings"

// EventSignal 单个事件信号
type EventSignal struct {
	Type     string `json:"type"`     // 婚恋/事业/财运_得/财运_损/健康/迁变/喜神临运
	Evidence string `json:"evidence"` // 命理证据描述
}

// YearSignals 某流年的信号集合
type YearSignals struct {
	Year        int           `json:"year"`
	Age         int           `json:"age"`
	GanZhi      string        `json:"gan_zhi"`
	DayunGanZhi string        `json:"dayun_gan_zhi"`
	Signals     []EventSignal `json:"signals"`
}

// ─── 关系表 ───────────────────────────────────────────────────────────────────

// 六冲
var sixChong = map[string]string{
	"子": "午", "午": "子",
	"丑": "未", "未": "丑",
	"寅": "申", "申": "寅",
	"卯": "酉", "酉": "卯",
	"辰": "戌", "戌": "辰",
	"巳": "亥", "亥": "巳",
}

// 六合
var sixHe = map[string]string{
	"子": "丑", "丑": "子",
	"寅": "亥", "亥": "寅",
	"卯": "戌", "戌": "卯",
	"辰": "酉", "酉": "辰",
	"巳": "申", "申": "巳",
	"午": "未", "未": "午",
}

// 地支相刑（单向）：key 刑 value
var sixXing = map[string]string{
	"寅": "巳", "巳": "申", "申": "寅", // 无礼之刑
	"丑": "戌", "戌": "未", "未": "丑", // 无恩之刑
	"子": "卯", "卯": "子",             // 无义之刑
	// 自刑（辰午酉亥）：在健康检测中特别处理
}

// 自刑支
var selfXing = map[string]bool{"辰": true, "午": true, "酉": true, "亥": true}

// 三合局：每组三支 → 对应五行
var sanheGroups = [4][3]string{
	{"申", "子", "辰"}, // 水局
	{"寅", "午", "戌"}, // 火局
	{"亥", "卯", "未"}, // 木局
	{"巳", "酉", "丑"}, // 金局
}

// 五行相克：克者 → 被克者
var wxKe = map[string]string{
	"mu": "tu", "huo": "jin", "tu": "shui", "jin": "mu", "shui": "huo",
}

// 五行相生：生者 → 被生者
var wxSheng = map[string]string{
	"mu": "huo", "huo": "tu", "tu": "jin", "jin": "shui", "shui": "mu",
}

var ganWuxing = map[string]string{
	"甲": "mu", "乙": "mu", "丙": "huo", "丁": "huo",
	"戊": "tu", "己": "tu", "庚": "jin", "辛": "jin",
	"壬": "shui", "癸": "shui",
}

var zhiWuxing = map[string]string{
	"子": "shui", "亥": "shui",
	"寅": "mu", "卯": "mu",
	"巳": "huo", "午": "huo",
	"申": "jin", "酉": "jin",
	"辰": "tu", "戌": "tu", "丑": "tu", "未": "tu",
}

// ─── 辅助函数 ─────────────────────────────────────────────────────────────────

func containsStr(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func isTaohuaBase(base, check string) bool {
	return (strings.Contains("申子辰", base) && check == "酉") ||
		(strings.Contains("寅午戌", base) && check == "卯") ||
		(strings.Contains("亥卯未", base) && check == "子") ||
		(strings.Contains("巳酉丑", base) && check == "午")
}

func isYimaBase(base, check string) bool {
	return (strings.Contains("申子辰", base) && check == "寅") ||
		(strings.Contains("寅午戌", base) && check == "申") ||
		(strings.Contains("亥卯未", base) && check == "巳") ||
		(strings.Contains("巳酉丑", base) && check == "亥")
}

// isSanheTriggered 判断 lnZhi 是否触发三合（与给定支集合中已有两支凑成三合局）
func isSanheTriggered(lnZhi string, existingZhi []string) (triggered bool, localWx string) {
	for _, group := range sanheGroups {
		for i, z := range group {
			if z != lnZhi {
				continue
			}
			// lnZhi 在此三合组，检查另外两支是否至少有一支在 existingZhi 中
			others := [2]string{group[(i+1)%3], group[(i+2)%3]}
			matchCount := 0
			for _, o := range others {
				if containsStr(existingZhi, o) {
					matchCount++
				}
			}
			if matchCount >= 1 {
				// 确定三合五行
				switch group {
				case [3]string{"申", "子", "辰"}:
					localWx = "shui"
				case [3]string{"寅", "午", "戌"}:
					localWx = "huo"
				case [3]string{"亥", "卯", "未"}:
					localWx = "mu"
				case [3]string{"巳", "酉", "丑"}:
					localWx = "jin"
				}
				return true, localWx
			}
		}
	}
	return false, ""
}

// isTiaohouXi 判断某天干是否为调候喜神（在 tiaohou.Expected 中）
func isTiaohouXi(gan string, tiaohou *TiaohouResult) bool {
	if tiaohou == nil {
		return false
	}
	return containsStr(tiaohou.Expected, gan)
}

// dayMasterStrength 粗估日主强弱：返回 "strong"/"weak"/"neutral"
// 简化算法：统计月支 + 其他天干地支中生我/比我 vs 克我/泄我的数量
func dayMasterStrength(natal *BaziResult) string {
	dayWx := ganWuxing[natal.DayGan]
	score := 0

	// 月支得令权重最大 (+3 / -3)
	mzWx := zhiWuxing[natal.MonthZhi]
	if mzWx == dayWx {
		score += 3 // 月令同气
	} else if wxSheng[mzWx] == dayWx {
		score += 3 // 月令生我
	} else if wxKe[mzWx] == dayWx {
		score -= 3 // 月令克我
	} else if wxSheng[dayWx] == mzWx {
		score -= 2 // 月令泄我
	}

	// 其余天干
	for _, g := range []string{natal.YearGan, natal.MonthGan, natal.HourGan} {
		gWx := ganWuxing[g]
		if gWx == dayWx {
			score++ // 比劫
		} else if wxSheng[gWx] == dayWx {
			score++ // 印生
		} else if wxKe[gWx] == dayWx {
			score-- // 官杀克
		} else if wxSheng[dayWx] == gWx {
			score-- // 食伤泄
		}
	}

	switch {
	case score >= 2:
		return "strong"
	case score <= -2:
		return "weak"
	default:
		return "neutral"
	}
}

// ─── 核心信号检测 ──────────────────────────────────────────────────────────────

// GetYearEventSignals 计算某一流年激活的事件信号
func GetYearEventSignals(natal *BaziResult, lnGan, lnZhi, dayunGanZhi, gender string) []EventSignal {
	var signals []EventSignal
	add := func(typ, evidence string) {
		signals = append(signals, EventSignal{Type: typ, Evidence: evidence})
	}

	dayGan := natal.DayGan
	dayZhi := natal.DayZhi
	yearZhi := natal.YearZhi
	monthZhi := natal.MonthZhi
	hourZhi := natal.HourZhi

	lnWx := ganWuxing[lnGan]
	dayWx := ganWuxing[dayGan]
	shishen := GetShiShen(dayGan, lnGan)
	strength := dayMasterStrength(natal)

	// 大运干支拆解
	var dyGan, dyZhi string
	dyRunes := []rune(dayunGanZhi)
	if len(dyRunes) >= 2 {
		dyGan = string(dyRunes[0])
		dyZhi = string(dyRunes[1])
	}
	dyShishen := GetShiShen(dayGan, dyGan)
	// 大运地支主气十神（权重高于天干，代表大运能量主体）
	dyZhiShishen := GetZhiShiShen(dayGan, dyZhi)

	// 用于三合的现有支集合（原局四支 + 大运支）
	existingZhi := []string{natal.YearZhi, monthZhi, dayZhi, hourZhi}
	if dyZhi != "" {
		existingZhi = append(existingZhi, dyZhi)
	}

	// ── ① 调候喜神透干 ────────────────────────────────────────────────────────
	if isTiaohouXi(lnGan, natal.Tiaohou) {
		add("喜神临运", lnGan+"为调候喜神（《穷通宝鉴》所取），该年喜神透干，全局运势有明显助力")
	}

	// ── ① 大运地支 × 流年地支 关系（最重要的事件触发机制）──────────────────
	if dyZhi != "" {
		// 大运地支 六冲 流年地支 → 大运流年双冲，重大事件触发年
		if chong, ok := sixChong[dyZhi]; ok && chong == lnZhi {
			add("综合变动", "大运地支"+dyZhi+"与流年地支"+lnZhi+"相冲，大运流年地支双冲，本年为重大事件高发年，各类信号均被放大")
		}
		// 大运地支 六合 流年地支 → 大运流年双合，能量聚合
		if he, ok := sixHe[dyZhi]; ok && he == lnZhi {
			add("综合变动", "大运地支"+dyZhi+"与流年地支"+lnZhi+"六合，大运流年地支相合，能量聚合，该年事件易有正向突破")
		}

		// 大运地支 六冲 日支（夫妻宫）→ 整个大运感情宫位处于震动底色
		if chong, ok := sixChong[dyZhi]; ok && chong == dayZhi {
			add("婚恋", "大运地支"+dyZhi+"冲日支"+dayZhi+"（夫妻宫），整个大运期间感情宫位持续震动，本年流年触发更易产生感情重大变化")
		}
		// 大运地支 六合 日支 → 整个大运感情宫位处于合住状态
		if he, ok := sixHe[dyZhi]; ok && he == dayZhi {
			add("婚恋", "大运地支"+dyZhi+"合住日支"+dayZhi+"（夫妻宫），大运期间感情宫位被激活，该年感情事件易有进展")
		}
	}

	// ── ② 婚恋信号 ───────────────────────────────────────────────────────────
	// 大运天干+地支是否均为财/官星（大运整体感情能量判断）
	dyFinanceDouble := (dyShishen == "偏财" || dyShishen == "正财") &&
		(dyZhiShishen == "偏财" || dyZhiShishen == "正财")
	dyOfficialDouble := (dyShishen == "正官" || dyShishen == "七杀") &&
		(dyZhiShishen == "正官" || dyZhiShishen == "七杀")

	if gender == "male" {
		if shishen == "偏财" || shishen == "正财" {
			label := map[string]string{"偏财": "女友", "正财": "妻星"}[shishen]
			evidence := lnGan + "透干为" + shishen + "（" + label + "象征）"
			if dyShishen == "偏财" || dyShishen == "正财" || dyZhiShishen == "偏财" || dyZhiShishen == "正财" {
				if dyFinanceDouble {
					evidence += "，大运" + dayunGanZhi + "干支均为财星，大运整体财星旺，流年再叠，感情婚恋年"
				} else {
					evidence += "，大运" + dyGan + "亦带财星气息，大运流年财星呼应"
				}
			}
			add("婚恋", evidence)
		}
	} else {
		if shishen == "正官" || shishen == "七杀" {
			label := map[string]string{"正官": "夫星", "七杀": "激情情感星"}[shishen]
			evidence := lnGan + "透干为" + shishen + "（" + label + "）"
			if dyShishen == "正官" || dyShishen == "七杀" || dyZhiShishen == "正官" || dyZhiShishen == "七杀" {
				if dyOfficialDouble {
					evidence += "，大运" + dayunGanZhi + "干支均为官星，大运整体官星旺，流年再叠，婚恋动象极强"
				} else {
					evidence += "，大运" + dyGan + "亦带官星气息，大运流年官星呼应"
				}
			}
			add("婚恋", evidence)
		}
	}

	// 夫妻宫（日支）六合 / 六冲
	if he, ok := sixHe[lnZhi]; ok && he == dayZhi {
		add("婚恋", "流年地支"+lnZhi+"与日支"+dayZhi+"（夫妻宫）六合，感情宫位被激活")
	}
	if chong, ok := sixChong[lnZhi]; ok && chong == dayZhi {
		add("婚恋", "流年地支"+lnZhi+"冲日支"+dayZhi+"（夫妻宫），感情宫位受震动，关系或有重大变化")
	}

	// 桃花临命
	if isTaohuaBase(yearZhi, lnZhi) || isTaohuaBase(dayZhi, lnZhi) {
		add("婚恋", "流年地支"+lnZhi+"为桃花星临命，人缘异性缘大旺")
	}

	// 三合局引动夫妻宫（lnZhi 与原局/大运支凑成三合，且其五行为财/官星五行）
	if ok, sanheWx := isSanheTriggered(lnZhi, existingZhi); ok {
		// 判断三合五行是否与感情星五行吻合
		var targetWx string
		if gender == "male" {
			// 财星五行 = 日主所克
			targetWx = wxKe[dayWx]
		} else {
			// 官星五行 = 克日主者
			for w, kTarget := range wxKe {
				if kTarget == dayWx {
					targetWx = w
					break
				}
			}
		}
		if sanheWx == targetWx {
			add("婚恋", "流年地支"+lnZhi+"引动三合局（"+sanheWx+"），感情星五行局成，婚恋机遇显著增强")
		}
	}

	// ── ③ 事业信号 ───────────────────────────────────────────────────────────
	isOfficialStar := shishen == "正官" || shishen == "七杀"
	isSealStar := shishen == "正印" || shishen == "偏印"

	if isOfficialStar {
		switch strength {
		case "weak":
			add("事业", lnGan+"透干为"+shishen+"，日主身弱，官杀压力较大，事业有阻力或竞争加剧")
		case "strong":
			add("事业", lnGan+"透干为"+shishen+"，日主身旺，官星有力，事业晋升或仕途机遇来临")
		default:
			add("事业", lnGan+"透干为"+shishen+"，官星临运，事业格局有变动")
		}
		// 大运干或地支也是官杀 → 双叠
		dyHasOfficial := dyShishen == "正官" || dyShishen == "七杀" || dyZhiShishen == "正官" || dyZhiShishen == "七杀"
		if dyHasOfficial {
			if dyOfficialDouble {
				add("事业", "大运"+dayunGanZhi+"干支均为官杀，整个大运官杀场域强劲，流年再叠，本年事业变动力度极大")
			} else {
				add("事业", "大运流年官杀双叠（大运"+dyGan+"/"+dyZhi+"，流年"+lnGan+"），仕途压力或机遇贯穿整年")
			}
		}
	}

	if isSealStar {
		add("事业", lnGan+"透干为"+shishen+"，印星护身生扶日主，利于考试晋升、资格认证或获贵人提携")
	}

	// 驿马动
	if isYimaBase(yearZhi, lnZhi) || isYimaBase(dayZhi, lnZhi) {
		add("事业", "流年地支"+lnZhi+"为驿马星，主奔波变动、出行迁移或职位调动")
	}

	// ── ④ 财运信号 ───────────────────────────────────────────────────────────
	isFinanceStar := shishen == "偏财" || shishen == "正财"
	if isFinanceStar {
		if strength == "weak" {
			add("财运_得", lnGan+"透干为"+shishen+"，财星现身，但日主身弱财多身弱，宜量力而为，切忌冒进")
		} else {
			add("财运_得", lnGan+"透干为"+shishen+"，财星透出，财运有望提升，宜主动把握进财机会")
		}
		// 大运干或地支也是财星 → 双叠
		dyHasFinance := dyShishen == "偏财" || dyShishen == "正财" || dyZhiShishen == "偏财" || dyZhiShishen == "正财"
		if dyHasFinance {
			if dyFinanceDouble {
				add("财运_得", "大运"+dayunGanZhi+"干支均为财星，整个大运财星旺盛，流年财星再叠，财运爆发年")
			} else {
				add("财运_得", "大运流年财星双叠（大运"+dyGan+"/"+dyZhi+"，流年"+lnGan+"），财运进项力度强，但需防比劫争财")
			}
		}
	}
	if shishen == "比肩" || shishen == "劫财" {
		add("财运_损", lnGan+"透干为"+shishen+"，比劫争财，财运有损耗风险，投资需谨慎")
	}
	// 流年食伤泄秀 → 有才华变现机会
	if shishen == "食神" || shishen == "伤官" {
		add("财运_得", lnGan+"透干为"+shishen+"，食伤生财，技艺才华有望变现，适合创业或副业尝试")
	}

	// ── ⑤ 健康信号 ───────────────────────────────────────────────────────────
	// 流年天干五行克日干五行
	if wxKe[lnWx] == dayWx {
		add("健康", "流年天干"+lnGan+"（"+lnWx+"）克制日干"+dayGan+"（"+dayWx+"），日主元气受损，需注意身体健康")
	}
	// 流年地支冲日支
	if chong, ok := sixChong[lnZhi]; ok && chong == dayZhi {
		add("健康", "流年地支"+lnZhi+"冲日支"+dayZhi+"，日柱受冲，体力精神有下滑风险")
	}
	// 流年地支冲年支（岁破）
	if chong, ok := sixChong[lnZhi]; ok && chong == yearZhi {
		add("健康", "流年地支"+lnZhi+"冲年支"+yearZhi+"，岁破临命，需防突发意外或家庭变故")
	}
	// 地支相刑（流年刑日支）
	if xing, ok := sixXing[lnZhi]; ok && xing == dayZhi {
		add("健康", "流年地支"+lnZhi+"刑日支"+dayZhi+"，地支相刑，易有手术、伤病或官非之虞")
	}
	// 自刑（流年地支自刑）
	if selfXing[lnZhi] && lnZhi == dayZhi {
		add("健康", "流年地支"+lnZhi+"与日支同支自刑，精神压力较大，需防积劳成疾")
	}

	// ── ⑥ 迁变信号 ───────────────────────────────────────────────────────────
	if isYimaBase(yearZhi, lnZhi) || isYimaBase(dayZhi, lnZhi) {
		add("迁变", "流年驿马星动，主环境变动、居所迁移或出行远游")
	}
	if chong, ok := sixChong[lnZhi]; ok && chong == yearZhi {
		add("迁变", "流年地支"+lnZhi+"冲年支"+yearZhi+"，岁破之年，人生格局易有较大转变")
	}

	return signals
}

// GetPastYearSignals 批量扫描过往流年事件信号
func GetPastYearSignals(result *BaziResult, gender string, currentYear, minAge int) []YearSignals {
	var out []YearSignals
	for _, dy := range result.Dayun {
		dyGanZhi := dy.Gan + dy.Zhi
		for _, ln := range dy.LiuNian {
			if ln.Age < minAge || ln.Year > currentYear {
				continue
			}
			lnRunes := []rune(ln.GanZhi)
			if len(lnRunes) < 2 {
				continue
			}
			sigs := GetYearEventSignals(result, string(lnRunes[0]), string(lnRunes[1]), dyGanZhi, gender)
			out = append(out, YearSignals{
				Year:        ln.Year,
				Age:         ln.Age,
				GanZhi:      ln.GanZhi,
				DayunGanZhi: dyGanZhi,
				Signals:     sigs,
			})
		}
	}
	return out
}
