package bazi

import (
	"fmt"
	"strings"
)

// ─── 信号 Source 常量 ────────────────────────────────────────────────────────

const (
	SourceShensha  = "神煞"
	SourceZhuwei   = "柱位互动"
	SourceHehua    = "合化"
	SourceKongwang = "空亡"
	SourceXing     = "刑"
	SourceHui      = "会"
	SourceFuyin    = "伏吟"
	SourceYongshen = "用神基底"
)

// 信号 Polarity 常量
const (
	PolarityJi      = "吉"
	PolarityXiong   = "凶"
	PolarityNeutral = "中性"
)

// YoungAgeCutoff 读书期与成人期边界年龄
// age < YoungAgeCutoff → 启用读书期语义分支（学业 / 性格 / 同窗等）
const YoungAgeCutoff = 18

// 读书期 Type 常量（替代成人期的 财运/事业/婚恋 系列）
const (
	TypeXueYeZiYuan    = "学业_资源" // 财星透干
	TypeXueYeJingZheng = "学业_竞争" // 比劫透干
	TypeXueYeYaLi      = "学业_压力" // 官杀透干
	TypeXueYeGuiRen    = "学业_贵人" // 印星透干
	TypeXueYeCaiYi     = "学业_才艺" // 食伤透干
	TypeXingGeQingYi   = "性格_情谊" // 合日支 / 桃花 / 同窗情
	TypeXingGePanNi    = "性格_叛逆" // 冲日支 / 情绪波动
)

// TypeJuShiZhong 三合/三会忌神局极凶标星信号
const TypeJuShiZhong = "局势_重"

// EventSignal 单个事件信号
type EventSignal struct {
	Type     string `json:"type"`               // 婚恋_合/冲/变 | 事业 | 财运_得/损 | 健康 | 迁变 | 喜神临运 | 综合变动 | 用神基底 | 大运合化 | 伏吟 | 反吟
	Evidence string `json:"evidence"`           // 命理证据描述
	Polarity string `json:"polarity,omitempty"` // 吉/凶/中性
	Source   string `json:"source,omitempty"`   // 来源标签
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

// 六害（穿破）：key 与 value 互害
var sixHai = map[string]string{
	"子": "未", "未": "子",
	"丑": "午", "午": "丑",
	"寅": "巳", "巳": "寅",
	"卯": "辰", "辰": "卯",
	"申": "亥", "亥": "申",
	"酉": "戌", "戌": "酉",
}

// 地支六合化神（双向）：pair → 化出五行(pinyin)
var zhiLiuheHua = map[[2]string]string{
	{"子", "丑"}: "tu", {"丑", "子"}: "tu",
	{"寅", "亥"}: "mu", {"亥", "寅"}: "mu",
	{"卯", "戌"}: "huo", {"戌", "卯"}: "huo",
	{"辰", "酉"}: "jin", {"酉", "辰"}: "jin",
	{"巳", "申"}: "shui", {"申", "巳"}: "shui",
	{"午", "未"}: "huo", {"未", "午"}: "huo",
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

// 三会局：每组三支 → 对应五行
var sanhuiGroups = [4][3]string{
	{"寅", "卯", "辰"}, // 木会
	{"巳", "午", "未"}, // 火会
	{"申", "酉", "戌"}, // 金会
	{"亥", "子", "丑"}, // 水会
}

// 三刑全局组合：寅巳申、丑未戌
var sanxingGroups = [2][3]string{
	{"寅", "巳", "申"},
	{"丑", "未", "戌"},
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

// pinyin 五行 → 中文五行
var wxPinyin2CN = map[string]string{
	"mu": "木", "huo": "火", "tu": "土", "jin": "金", "shui": "水",
}

// 天干五合（组合双向）→ 化神五行（pinyin）
var ganWuhe = map[[2]string]string{
	{"甲", "己"}: "tu", {"己", "甲"}: "tu",
	{"乙", "庚"}: "jin", {"庚", "乙"}: "jin",
	{"丙", "辛"}: "shui", {"辛", "丙"}: "shui",
	{"丁", "壬"}: "mu", {"壬", "丁"}: "mu",
	{"戊", "癸"}: "huo", {"癸", "戊"}: "huo",
}

// 地支藏干主气
var zhiZhuQi = map[string]string{
	"子": "癸", "丑": "己", "寅": "甲", "卯": "乙",
	"辰": "戊", "巳": "丙", "午": "丁", "未": "己",
	"申": "庚", "酉": "辛", "戌": "戊", "亥": "壬",
}

// 旬空表：日柱所在旬 → 旬空二支
// 60甲子分6旬，每旬10日，旬空 = 该旬中"未配天干"的两支
var xunkongTable = map[string][2]string{
	// 甲子旬（甲子-癸酉）→ 戌亥空
	"甲子": {"戌", "亥"}, "乙丑": {"戌", "亥"}, "丙寅": {"戌", "亥"}, "丁卯": {"戌", "亥"}, "戊辰": {"戌", "亥"},
	"己巳": {"戌", "亥"}, "庚午": {"戌", "亥"}, "辛未": {"戌", "亥"}, "壬申": {"戌", "亥"}, "癸酉": {"戌", "亥"},
	// 甲戌旬（甲戌-癸未）→ 申酉空
	"甲戌": {"申", "酉"}, "乙亥": {"申", "酉"}, "丙子": {"申", "酉"}, "丁丑": {"申", "酉"}, "戊寅": {"申", "酉"},
	"己卯": {"申", "酉"}, "庚辰": {"申", "酉"}, "辛巳": {"申", "酉"}, "壬午": {"申", "酉"}, "癸未": {"申", "酉"},
	// 甲申旬（甲申-癸巳）→ 午未空
	"甲申": {"午", "未"}, "乙酉": {"午", "未"}, "丙戌": {"午", "未"}, "丁亥": {"午", "未"}, "戊子": {"午", "未"},
	"己丑": {"午", "未"}, "庚寅": {"午", "未"}, "辛卯": {"午", "未"}, "壬辰": {"午", "未"}, "癸巳": {"午", "未"},
	// 甲午旬（甲午-癸卯）→ 辰巳空
	"甲午": {"辰", "巳"}, "乙未": {"辰", "巳"}, "丙申": {"辰", "巳"}, "丁酉": {"辰", "巳"}, "戊戌": {"辰", "巳"},
	"己亥": {"辰", "巳"}, "庚子": {"辰", "巳"}, "辛丑": {"辰", "巳"}, "壬寅": {"辰", "巳"}, "癸卯": {"辰", "巳"},
	// 甲辰旬（甲辰-癸丑）→ 寅卯空
	"甲辰": {"寅", "卯"}, "乙巳": {"寅", "卯"}, "丙午": {"寅", "卯"}, "丁未": {"寅", "卯"}, "戊申": {"寅", "卯"},
	"己酉": {"寅", "卯"}, "庚戌": {"寅", "卯"}, "辛亥": {"寅", "卯"}, "壬子": {"寅", "卯"}, "癸丑": {"寅", "卯"},
	// 甲寅旬（甲寅-癸亥）→ 子丑空
	"甲寅": {"子", "丑"}, "乙卯": {"子", "丑"}, "丙辰": {"子", "丑"}, "丁巳": {"子", "丑"}, "戊午": {"子", "丑"},
	"己未": {"子", "丑"}, "庚申": {"子", "丑"}, "辛酉": {"子", "丑"}, "壬戌": {"子", "丑"}, "癸亥": {"子", "丑"},
}

// 神煞白名单：神煞名 → 默认 (Polarity, Type)
type shenshaMeta struct {
	Polarity string
	Type     string
	Hint     string
	IsHeavy  bool // 重煞标记：羊刃/白虎/岁破/丧门/吊客/灾煞/劫煞/亡神
}

var shenshaWhitelist = map[string]shenshaMeta{
	"天乙贵人": {PolarityJi, "事业", "天乙贵人临运，主贵人扶持、化险为夷", false},
	"天德贵人": {PolarityJi, "事业", "天德贵人临运，主逢凶化吉、福德深厚", false},
	"月德贵人": {PolarityJi, "事业", "月德贵人临运，主性情温润、得长辈助力", false},
	"天德合":  {PolarityJi, "事业", "天德合临运，福德相合，吉中藏吉", false},
	"月德合":  {PolarityJi, "事业", "月德合临运，温和有助", false},
	"德秀贵人": {PolarityJi, "事业", "德秀贵人临运，文采才华彰显", false},
	"将星":   {PolarityJi, "事业", "将星临运，主统御能力增强、领导机会", false},
	"福星贵人": {PolarityJi, "事业", "福星贵人临运，主进财得福", false},
	"文昌贵人": {PolarityJi, "事业", "文昌贵人临运，主考试、文书、晋升顺遂", false},
	"国印贵人": {PolarityJi, "事业", "国印贵人临运，主权印职位", false},
	"金舆贵人": {PolarityJi, "事业", "金舆贵人临运，主人际财禄", false},
	"太极贵人": {PolarityJi, "事业", "太极贵人临运，主始终如一、完成大事", false},
	"天厨贵人": {PolarityJi, "财运_得", "天厨贵人临运，主衣食丰盈", false},
	"天医":   {PolarityJi, "健康", "天医临运，主疾病减损、医疗顺遂", false},
	"天喜":   {PolarityJi, "婚恋_合", "天喜临运，主喜庆之事", false},

	"羊刃": {PolarityXiong, "健康", "羊刃临运，宜防开刀、血光、车祸", true},
	"白虎": {PolarityXiong, "健康", "白虎临运，主孝服、突发伤痛或意外", true},
	"丧门": {PolarityXiong, "健康", "丧门临运，宜防家中长辈丧事或孝服", true},
	"吊客": {PolarityXiong, "健康", "吊客临运，主家事操心、孝服之忧", true},
	"勾绞": {PolarityXiong, "综合变动", "勾绞临运，主纷争口舌、官司缠绕", false},
	"亡神": {PolarityXiong, "综合变动", "亡神临运，宜防破耗、失去", true},
	"劫煞": {PolarityXiong, "综合变动", "劫煞临运，宜防意外破财或人事冲突", true},
	"灾煞": {PolarityXiong, "健康", "灾煞临运，宜防突发灾患", true},
	"流霞": {PolarityXiong, "健康", "流霞临运，女命防产厄、男命防伤灾", false},
	"墓门": {PolarityXiong, "综合变动", "墓门临运，主郁闷、消沉、收敛", false},

	"红艳": {PolarityNeutral, "婚恋_合", "红艳临运，主桃花、感情绚烂", false},
	"桃花": {PolarityNeutral, "婚恋_合", "桃花临运，人缘异性缘旺", false},
	"驿马": {PolarityNeutral, "迁变", "驿马临运，主奔波、出行、变动", false},
	"华盖": {PolarityNeutral, "迁变", "华盖临运，主清高、宗教、艺术、孤独", false},
	"孤辰": {PolarityNeutral, "婚恋_合", "孤辰临运，宜防孤独情绪", false},
	"寡宿": {PolarityNeutral, "婚恋_合", "寡宿临运，宜防孤独情绪", false},
	"词馆": {PolarityJi, "事业", "词馆临运，主文笔、口才、学问", false},
	"禄神": {PolarityJi, "财运_得", "禄神临运，主稳定收入、职位之禄", false},
}

// ─── 辅助函数 ─────────────────────────────────────────────────────────────────

// hideGanMainZhong 取藏干数组前两项（主气+中气），余气不计
func hideGanMainZhong(hideGan []string) []string {
	if len(hideGan) <= 2 {
		return hideGan
	}
	return hideGan[:2]
}

// collectYongshenPositions 收集原局用神/忌神覆盖的干支位置列表
// 地支匹配：本气五行 OR 藏干主气+中气五行，任一命中即计入
func collectYongshenPositions(natal *BaziResult) (yongPos, jiPos []string) {
	if natal == nil || (natal.Yongshen == "" && natal.Jishen == "") {
		return nil, nil
	}
	cn2pin := map[string]string{"木": "mu", "火": "huo", "土": "tu", "金": "jin", "水": "shui"}
	yongSet := map[string]bool{}
	jiSet := map[string]bool{}
	for _, ch := range natal.Yongshen {
		if p, ok := cn2pin[string(ch)]; ok {
			yongSet[p] = true
		}
	}
	for _, ch := range natal.Jishen {
		if p, ok := cn2pin[string(ch)]; ok {
			jiSet[p] = true
		}
	}
	for _, g := range []struct{ label, gan string }{
		{"年干", natal.YearGan}, {"月干", natal.MonthGan},
		{"日干", natal.DayGan}, {"时干", natal.HourGan},
	} {
		if g.gan == "" {
			continue
		}
		wx := ganWuxing[g.gan]
		if yongSet[wx] {
			yongPos = append(yongPos, g.label)
		}
		if jiSet[wx] {
			jiPos = append(jiPos, g.label)
		}
	}
	for _, b := range []struct {
		label   string
		zhi     string
		hideGan []string
	}{
		{"年支", natal.YearZhi, natal.YearHideGan},
		{"月支", natal.MonthZhi, natal.MonthHideGan},
		{"日支", natal.DayZhi, natal.DayHideGan},
		{"时支", natal.HourZhi, natal.HourHideGan},
	} {
		if b.zhi == "" {
			continue
		}
		yongHit, jiHit := yongSet[zhiWuxing[b.zhi]], jiSet[zhiWuxing[b.zhi]]
		for _, hg := range hideGanMainZhong(b.hideGan) {
			hgWx := ganWuxing[hg]
			if yongSet[hgWx] {
				yongHit = true
			}
			if jiSet[hgWx] {
				jiHit = true
			}
		}
		if yongHit {
			yongPos = append(yongPos, b.label)
		}
		if jiHit {
			jiPos = append(jiPos, b.label)
		}
	}
	return yongPos, jiPos
}

// pillarWeightLabel 返回被冲克宫位的权重标注文字（日柱最重，月柱次之，其余不标）
func pillarWeightLabel(pos string) string {
	switch pos {
	case "日干", "日支":
		return "（日柱宫位，权重较重）"
	case "月干", "月支":
		return "（月柱宫位，权重次之）"
	default:
		return ""
	}
}

// juGroup 三合/三会局组定义
type juGroup struct {
	branches [3]string
	wx       string // 局五行（pinyin）
	kind     string // "三合" 或 "三会"
}

// allJuGroups 所有三合/三会局
var allJuGroups = []juGroup{
	// 三合
	{[3]string{"申", "子", "辰"}, "shui", "三合"},
	{[3]string{"寅", "午", "戌"}, "huo", "三合"},
	{[3]string{"亥", "卯", "未"}, "mu", "三合"},
	{[3]string{"巳", "酉", "丑"}, "jin", "三合"},
	// 三会
	{[3]string{"寅", "卯", "辰"}, "mu", "三会"},
	{[3]string{"巳", "午", "未"}, "huo", "三会"},
	{[3]string{"申", "酉", "戌"}, "jin", "三会"},
	{[3]string{"亥", "子", "丑"}, "shui", "三会"},
}

// collectJuShiSignals 检测流年地支（结合大运+原局）是否补全三合/三会局。
// 三支全齐（matchCount=2）才算局成；局五行=用神且克忌神→吉；局五行=忌神→极凶（TypeJuShiZhong）。
func collectJuShiSignals(natal *BaziResult, lnZhi, dyZhi string) []EventSignal {
	if natal == nil || lnZhi == "" {
		return nil
	}
	if natal.Yongshen == "" && natal.Jishen == "" {
		return nil
	}

	existingZhi := []string{natal.YearZhi, natal.MonthZhi, natal.DayZhi, natal.HourZhi}
	if dyZhi != "" {
		existingZhi = append(existingZhi, dyZhi)
	}

	cn2pin := map[string]string{"木": "mu", "火": "huo", "土": "tu", "金": "jin", "水": "shui"}
	yongSet := map[string]bool{}
	jiSet := map[string]bool{}
	for _, ch := range natal.Yongshen {
		if p, ok := cn2pin[string(ch)]; ok {
			yongSet[p] = true
		}
	}
	for _, ch := range natal.Jishen {
		if p, ok := cn2pin[string(ch)]; ok {
			jiSet[p] = true
		}
	}

	var sigs []EventSignal
	for _, g := range allJuGroups {
		// 确认 lnZhi 在该组中
		lnIdx := -1
		for i, z := range g.branches {
			if z == lnZhi {
				lnIdx = i
				break
			}
		}
		if lnIdx < 0 {
			continue
		}
		// 另外两支必须都在 existingZhi 中（三支全齐）
		other := make([]string, 0, 2)
		for i, z := range g.branches {
			if i != lnIdx {
				other = append(other, z)
			}
		}
		if !containsStr(existingZhi, other[0]) || !containsStr(existingZhi, other[1]) {
			continue
		}

		localWx := g.wx
		juName := string(g.branches[0]) + string(g.branches[1]) + string(g.branches[2])
		localWxCN := wxPinyin2CN[localWx]

		if yongSet[localWx] {
			keTarget := wxKe[localWx]
			if jiSet[keTarget] {
				jiCN := wxPinyin2CN[keTarget]
				sigs = append(sigs, EventSignal{
					Type:     "综合变动",
					Evidence: fmt.Sprintf("流年%s补全%s%s%s局，%s势力大增，克制忌神%s，用神赢，应期吉", lnZhi, juName, g.kind, localWxCN, localWxCN, jiCN),
					Polarity: PolarityJi,
					Source:   SourceZhuwei,
				})
			}
		} else if jiSet[localWx] {
			sigs = append(sigs, EventSignal{
				Type:     TypeJuShiZhong,
				Evidence: fmt.Sprintf("★流年%s补全%s%s%s局，忌神势力极强，用神承压，应期极凶", lnZhi, juName, g.kind, localWxCN),
				Polarity: PolarityXiong,
				Source:   SourceZhuwei,
			})
		}
	}
	return sigs
}

// yingqiTagged 内部用：带合并元数据的应期信号
type yingqiTagged struct {
	EventSignal
	srcLabel string // "流年" 或 "大运"
	hitPos   string // 被命中的原局位置，如"日支"
	hitKind  string // 交互类型："冲"/"刑"/"穿"（仅这三类参与双冲合并）
}

// collectYingqiSignals 检测流年/大运干支与原局用神/忌神位置的刑冲克合穿五种交互
// 每条命中独立输出 EventSignal，极性由交互类型与用神/忌神位置联合决定
// 大运+流年同类型双冲同一用神位时自动合并为叠加信号
func collectYingqiSignals(natal *BaziResult, lnGan, lnZhi, dyGan, dyZhi string) []EventSignal {
	yongPos, jiPos := collectYongshenPositions(natal)
	if len(yongPos) == 0 && len(jiPos) == 0 {
		return nil
	}
	var tagged []yingqiTagged
	addTagged := func(sig EventSignal, src, pos, kind string) {
		tagged = append(tagged, yingqiTagged{sig, src, pos, kind})
	}
	// sigs is kept for non-tagged appends (合化类 etc.)
	var extraSigs []EventSignal

	posGanVal := map[string]string{
		"年干": natal.YearGan, "月干": natal.MonthGan,
		"日干": natal.DayGan, "时干": natal.HourGan,
	}
	posZhiVal := map[string]string{
		"年支": natal.YearZhi, "月支": natal.MonthZhi,
		"日支": natal.DayZhi, "时支": natal.HourZhi,
	}

	cn2pin := map[string]string{"木": "mu", "火": "huo", "土": "tu", "金": "jin", "水": "shui"}
	yongSet := map[string]bool{}
	jiSet := map[string]bool{}
	for _, ch := range natal.Yongshen {
		if p, ok := cn2pin[string(ch)]; ok {
			yongSet[p] = true
		}
	}
	for _, ch := range natal.Jishen {
		if p, ok := cn2pin[string(ch)]; ok {
			jiSet[p] = true
		}
	}

	// allNatalGan 原局四柱天干，供五行流通检查用
	allNatalGan := []string{natal.YearGan, natal.MonthGan, natal.DayGan, natal.HourGan}

	// checkGan 检测一个天干与原局用神/忌神天干位的克、五合交互
	checkGan := func(inGan, inLabel string) {
		if inGan == "" {
			return
		}
		inWx := ganWuxing[inGan]
		// 3.3 天干克
		// 若流年/大运天干本身属于用神五行，则不对用神位产生克害凶信号
		if !yongSet[inWx] {
			for _, pos := range yongPos {
				pg, ok := posGanVal[pos]
				if !ok || pg == "" {
					continue
				}
				if wxKe[inWx] == ganWuxing[pg] {
					yongWx := ganWuxing[pg]
					// 五行流通检查：原局天干存在 M 使得 inWx生M 且 M生yongWx
					hasLiutong := false
					for _, ng := range allNatalGan {
						if ng == "" {
							continue
						}
						m := ganWuxing[ng]
						if wxSheng[inWx] == m && wxSheng[m] == yongWx {
							hasLiutong = true
							break
						}
					}
					if hasLiutong {
						addTagged(EventSignal{
							Type:     "综合变动",
							Evidence: fmt.Sprintf("%s%s克原局%s%s（用神位），五行流通，克势转化，力度减弱%s", inLabel, inGan, pos, pg, pillarWeightLabel(pos)),
							Polarity: PolarityNeutral, Source: SourceZhuwei,
						}, inLabel, pos, "克")
					} else {
						addTagged(EventSignal{
							Type:     "综合变动",
							Evidence: fmt.Sprintf("%s%s克原局%s%s（用神位），用神受克，应期凶%s", inLabel, inGan, pos, pg, pillarWeightLabel(pos)),
							Polarity: PolarityXiong, Source: SourceZhuwei,
						}, inLabel, pos, "克")
					}
				}
			}
		}
		for _, pos := range jiPos {
			pg, ok := posGanVal[pos]
			if !ok || pg == "" {
				continue
			}
			if wxKe[inWx] == ganWuxing[pg] {
				addTagged(EventSignal{
					Type:     "综合变动",
					Evidence: fmt.Sprintf("%s%s克原局%s%s（忌神位），忌神受制，应期吉%s", inLabel, inGan, pos, pg, pillarWeightLabel(pos)),
					Polarity: PolarityJi, Source: SourceZhuwei,
				}, inLabel, pos, "克")
			}
		}
		// 3.7 天干五合
		for _, pos := range append(append([]string{}, yongPos...), jiPos...) {
			pg, ok := posGanVal[pos]
			if !ok || pg == "" {
				continue
			}
			huaWx, hasHe := ganWuhe[[2]string{inGan, pg}]
			if !hasHe {
				continue
			}
			hua := zhiWuxing[natal.MonthZhi] == huaWx || zhiWuxing[dyZhi] == huaWx
			huaWxCN := wxPinyin2CN[huaWx]
			isYongPos := containsStr(yongPos, pos)
			if hua {
				if yongSet[huaWx] {
					extraSigs = append(extraSigs, EventSignal{
						Type:     "综合变动",
						Evidence: fmt.Sprintf("%s%s合原局%s%s，化%s属用神，合化成立，应期吉", inLabel, inGan, pos, pg, huaWxCN),
						Polarity: PolarityJi, Source: SourceHehua,
					})
				} else if jiSet[huaWx] {
					extraSigs = append(extraSigs, EventSignal{
						Type:     "综合变动",
						Evidence: fmt.Sprintf("%s%s合原局%s%s，化%s属忌神，合化成立，应期凶", inLabel, inGan, pos, pg, huaWxCN),
						Polarity: PolarityXiong, Source: SourceHehua,
					})
				}
			} else if isYongPos {
				extraSigs = append(extraSigs, EventSignal{
					Type:     "综合变动",
					Evidence: fmt.Sprintf("%s%s合住原局%s%s（用神位），合而不化，用神被锁，应期凶", inLabel, inGan, pos, pg),
					Polarity: PolarityXiong, Source: SourceHehua,
				})
			} else {
				extraSigs = append(extraSigs, EventSignal{
					Type:     "综合变动",
					Evidence: fmt.Sprintf("%s%s合住原局%s%s（忌神位），合而不化，忌神被锁，应期吉", inLabel, inGan, pos, pg),
					Polarity: PolarityJi, Source: SourceHehua,
				})
			}
		}
	}

	// checkZhi 检测一个地支与原局用神/忌神地支位的冲、刑、穿、六合交互
	checkZhi := func(inZhi, inLabel string) {
		if inZhi == "" {
			return
		}
		// 流年/大运地支本身的五行：若属用神则不对用神位产生冲克刑穿凶信号
		inZhiWx := zhiWuxing[inZhi]
		// 3.4 地支冲
		if !yongSet[inZhiWx] {
			for _, pos := range yongPos {
				if pz, ok := posZhiVal[pos]; ok && pz != "" {
					if c, ok2 := sixChong[inZhi]; ok2 && c == pz {
						addTagged(EventSignal{
							Type:     "综合变动",
							Evidence: fmt.Sprintf("%s%s冲原局%s%s（用神位），用神受冲，应期力度强，应期凶%s", inLabel, inZhi, pos, pz, pillarWeightLabel(pos)),
							Polarity: PolarityXiong, Source: SourceZhuwei,
						}, inLabel, pos, "冲")
					}
				}
			}
		}
		for _, pos := range jiPos {
			if pz, ok := posZhiVal[pos]; ok && pz != "" {
				if c, ok2 := sixChong[inZhi]; ok2 && c == pz {
					addTagged(EventSignal{
						Type:     "综合变动",
						Evidence: fmt.Sprintf("%s%s冲原局%s%s（忌神位），忌神受冲，应期力度强，应期吉%s", inLabel, inZhi, pos, pz, pillarWeightLabel(pos)),
						Polarity: PolarityJi, Source: SourceZhuwei,
					}, inLabel, pos, "冲")
				}
			}
		}
		// 3.5 地支刑
		for _, set := range []struct {
			poses  []string
			isYong bool
		}{{yongPos, true}, {jiPos, false}} {
			// 若流年/大运地支本身属用神，跳过对用神位的刑害凶信号
			if set.isYong && yongSet[inZhiWx] {
				continue
			}
			for _, pos := range set.poses {
				pz, ok := posZhiVal[pos]
				if !ok || pz == "" {
					continue
				}
				xingHit := false
				if x, ok2 := sixXing[inZhi]; ok2 && x == pz {
					xingHit = true
				}
				if selfXing[inZhi] && inZhi == pz {
					xingHit = true
				}
				if !xingHit {
					continue
				}
				posName := "用神"
				pol := PolarityXiong
				if !set.isYong {
					posName = "忌神"
					pol = PolarityJi
				}
				addTagged(EventSignal{
					Type:     "综合变动",
					Evidence: fmt.Sprintf("%s%s刑原局%s%s（%s位），%s位受刑，应期力度强，应期%s%s", inLabel, inZhi, pos, pz, posName, posName, map[bool]string{true: "凶", false: "吉"}[set.isYong], pillarWeightLabel(pos)),
					Polarity: pol, Source: SourceXing,
				}, inLabel, pos, "刑")
			}
		}
		// 3.6 地支穿（六害）
		if !yongSet[inZhiWx] {
			for _, pos := range yongPos {
				if pz, ok := posZhiVal[pos]; ok && pz != "" {
					if hai, ok2 := sixHai[inZhi]; ok2 && hai == pz {
						addTagged(EventSignal{
							Type:     "综合变动",
							Evidence: fmt.Sprintf("%s%s穿害原局%s%s（用神位），用神受穿，应期力度中，应期凶%s", inLabel, inZhi, pos, pz, pillarWeightLabel(pos)),
							Polarity: PolarityXiong, Source: SourceZhuwei,
						}, inLabel, pos, "穿")
					}
				}
			}
		}
		for _, pos := range jiPos {
			if pz, ok := posZhiVal[pos]; ok && pz != "" {
				if hai, ok2 := sixHai[inZhi]; ok2 && hai == pz {
					addTagged(EventSignal{
						Type:     "综合变动",
						Evidence: fmt.Sprintf("%s%s穿害原局%s%s（忌神位），忌神受穿，应期力度中，应期吉%s", inLabel, inZhi, pos, pz, pillarWeightLabel(pos)),
						Polarity: PolarityJi, Source: SourceZhuwei,
					}, inLabel, pos, "穿")
				}
			}
		}
		// 3.7 地支六合
		for _, set := range []struct {
			poses  []string
			isYong bool
		}{{yongPos, true}, {jiPos, false}} {
			for _, pos := range set.poses {
				pz, ok := posZhiVal[pos]
				if !ok || pz == "" {
					continue
				}
				he, ok2 := sixHe[inZhi]
				if !ok2 || he != pz {
					continue
				}
				huaWx, ok3 := zhiLiuheHua[[2]string{inZhi, pz}]
				if !ok3 {
					continue
				}
				hua := zhiWuxing[natal.MonthZhi] == huaWx
				huaWxCN := wxPinyin2CN[huaWx]
				if hua {
					if yongSet[huaWx] {
						extraSigs = append(extraSigs, EventSignal{
							Type:     "综合变动",
							Evidence: fmt.Sprintf("%s%s六合原局%s%s，化%s属用神，应期吉", inLabel, inZhi, pos, pz, huaWxCN),
							Polarity: PolarityJi, Source: SourceHehua,
						})
					} else if jiSet[huaWx] {
						extraSigs = append(extraSigs, EventSignal{
							Type:     "综合变动",
							Evidence: fmt.Sprintf("%s%s六合原局%s%s，化%s属忌神，应期凶", inLabel, inZhi, pos, pz, huaWxCN),
							Polarity: PolarityXiong, Source: SourceHehua,
						})
					}
				} else {
					posName := "用神"
					pol := PolarityXiong
					if !set.isYong {
						posName = "忌神"
						pol = PolarityJi
					}
					extraSigs = append(extraSigs, EventSignal{
						Type:     "综合变动",
						Evidence: fmt.Sprintf("%s%s六合原局%s%s（%s位），合而不化，%s被锁，应期力度弱，应期%s%s", inLabel, inZhi, pos, pz, posName, posName, map[bool]string{true: "凶", false: "吉"}[set.isYong], pillarWeightLabel(pos)),
						Polarity: pol, Source: SourceHehua,
					})
				}
			}
		}
	}

	checkGan(lnGan, "流年")
	checkZhi(lnZhi, "流年")
	checkGan(dyGan, "大运")
	checkZhi(dyZhi, "大运")

	// 大运+流年双重命中同一用神位合并（仅冲/刑/穿三类）
	merged := make([]bool, len(tagged))
	var result []EventSignal
	for i := 0; i < len(tagged); i++ {
		if merged[i] {
			continue
		}
		a := tagged[i]
		// 只对冲/刑/穿寻找对应的大运/流年配对
		if a.hitKind != "冲" && a.hitKind != "刑" && a.hitKind != "穿" {
			result = append(result, a.EventSignal)
			continue
		}
		pairIdx := -1
		for j := i + 1; j < len(tagged); j++ {
			b := tagged[j]
			if !merged[j] && b.srcLabel != a.srcLabel &&
				b.hitPos == a.hitPos && b.hitKind == a.hitKind && b.Polarity == a.Polarity {
				pairIdx = j
				break
			}
		}
		if pairIdx >= 0 {
			merged[pairIdx] = true
			result = append(result, EventSignal{
				Type:     a.Type,
				Evidence: fmt.Sprintf("大运流年双重命中：%s；大运流年双重命中，力度倍增%s", a.Evidence, pillarWeightLabel(a.hitPos)),
				Polarity: a.Polarity,
				Source:   a.Source,
			})
		} else {
			result = append(result, a.EventSignal)
		}
	}
	return append(result, extraSigs...)
}

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
		hasLn := false
		for _, z := range group {
			if z == lnZhi {
				hasLn = true
				break
			}
		}
		if !hasLn {
			continue
		}
		matchCount := 0
		for _, g := range group {
			if g != lnZhi && containsStr(existingZhi, g) {
				matchCount++
			}
		}
		if matchCount >= 1 {
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
	return false, ""
}

// isSanhuiTriggered 判断 lnZhi 是否触发三会（与给定支集合中已有两支凑成三会局）
func isSanhuiTriggered(lnZhi string, existingZhi []string) (triggered bool, localWx string) {
	for _, group := range sanhuiGroups {
		hasLn := false
		for _, z := range group {
			if z == lnZhi {
				hasLn = true
				break
			}
		}
		if !hasLn {
			continue
		}
		matchCount := 0
		for _, g := range group {
			if g != lnZhi && containsStr(existingZhi, g) {
				matchCount++
			}
		}
		if matchCount >= 2 {
			// 三会需要齐全（流年补足最后一支，原局/大运至少含其余两支）
			switch group {
			case [3]string{"寅", "卯", "辰"}:
				localWx = "mu"
			case [3]string{"巳", "午", "未"}:
				localWx = "huo"
			case [3]string{"申", "酉", "戌"}:
				localWx = "jin"
			case [3]string{"亥", "子", "丑"}:
				localWx = "shui"
			}
			return true, localWx
		}
	}
	return false, ""
}

// isSanxingTriggered 判断三刑全局是否凑齐（流年地支 + 既有支集合）
func isSanxingTriggered(lnZhi string, existingZhi []string) (triggered bool, kind string) {
	for _, group := range sanxingGroups {
		hasLn := false
		for _, z := range group {
			if z == lnZhi {
				hasLn = true
				break
			}
		}
		if !hasLn {
			continue
		}
		matchCount := 0
		for _, g := range group {
			if g != lnZhi && containsStr(existingZhi, g) {
				matchCount++
			}
		}
		if matchCount >= 2 {
			return true, group[0] + group[1] + group[2]
		}
	}
	return false, ""
}

// isFuyin 判断是否伏吟（干支完全相同）
func isFuyin(lnGan, lnZhi, pillarGan, pillarZhi string) bool {
	return pillarGan != "" && pillarZhi != "" && lnGan == pillarGan && lnZhi == pillarZhi
}

// isFanyin 判断是否反吟（天干相克 + 地支六冲）
func isFanyin(lnGan, lnZhi, pillarGan, pillarZhi string) bool {
	if pillarGan == "" || pillarZhi == "" {
		return false
	}
	lnGanWx := ganWuxing[lnGan]
	pillarGanWx := ganWuxing[pillarGan]
	if !((wxKe[lnGanWx] == pillarGanWx) || (wxKe[pillarGanWx] == lnGanWx)) {
		return false
	}
	if c, ok := sixChong[lnZhi]; ok && c == pillarZhi {
		return true
	}
	return false
}

// getXunkong 按日柱（日干+日支）查表返回旬空二支
func getXunkong(dayGan, dayZhi string) (string, string) {
	gz := dayGan + dayZhi
	if pair, ok := xunkongTable[gz]; ok {
		return pair[0], pair[1]
	}
	return "", ""
}

// isTiaohouXi 判断某天干是否为调候喜神（在 tiaohou.Expected 中）
func isTiaohouXi(gan string, tiaohou *TiaohouResult) bool {
	if tiaohou == nil {
		return false
	}
	return containsStr(tiaohou.Expected, gan)
}

// getYongshenBaseline 根据流年天干五行 vs 命主用神/忌神，返回基底色 polarity 与说明
// 用神/忌神格式为中文五行串联（如 "水木"、"金土火"）；缺失时降级到调候喜神
func getYongshenBaseline(natal *BaziResult, lnGan string) (polarity, evidence string) {
	if natal == nil {
		return "", ""
	}
	lnWxPin := ganWuxing[lnGan]
	lnWxCN, ok := wxPinyin2CN[lnWxPin]
	if !ok {
		return "", ""
	}

	if natal.Yongshen != "" || natal.Jishen != "" {
		isYong := strings.Contains(natal.Yongshen, lnWxCN)
		isJi := strings.Contains(natal.Jishen, lnWxCN)
		switch {
		case isYong && !isJi:
			return PolarityJi, fmt.Sprintf("流年%s（%s）属命主用神（%s），全年定调偏吉", lnGan, lnWxCN, natal.Yongshen)
		case !isYong && isJi:
			return PolarityXiong, fmt.Sprintf("流年%s（%s）属命主忌神（%s），全年定调偏凶", lnGan, lnWxCN, natal.Jishen)
		case !isYong && !isJi:
			return PolarityNeutral, fmt.Sprintf("流年%s（%s）非用神亦非忌神（用神：%s／忌神：%s），全年定调中性", lnGan, lnWxCN, natal.Yongshen, natal.Jishen)
		}
	}

	// 降级：调候喜神
	if natal.Tiaohou != nil && len(natal.Tiaohou.Expected) > 0 {
		if containsStr(natal.Tiaohou.Expected, lnGan) {
			return PolarityJi, fmt.Sprintf("流年%s为调候喜神（综合用神信息缺失，按调候参考），全年定调偏吉", lnGan)
		}
		return PolarityNeutral, ""
	}

	return "", ""
}

// applyPolarity 给信号填写 Polarity；signalSelf 为信号本身倾向（如"吉/凶/中性"），baseline 为基底色
// 规则：
//   - 神煞/伏吟/反吟自带的强 polarity 优先
//   - 普通事件信号：若与 baseline 同向则继承 baseline；若与 baseline 冲突，evidence 中已注明转折则保留 baseline；否则取 signalSelf
//   - baseline 为空时直接采用 signalSelf
func applyPolarity(sig *EventSignal, signalSelf, baseline string, source string) {
	sig.Source = source
	switch source {
	case SourceShensha, SourceFuyin:
		// 神煞与伏吟反吟自带强 polarity，由调用方在 signalSelf 给定，直接采用
		sig.Polarity = signalSelf
		return
	}
	switch {
	case baseline == "":
		sig.Polarity = signalSelf
	case signalSelf == "":
		sig.Polarity = baseline
	case signalSelf == baseline:
		sig.Polarity = baseline
	case signalSelf == PolarityNeutral:
		sig.Polarity = baseline
	case baseline == PolarityNeutral:
		sig.Polarity = signalSelf
	default:
		// 冲突：以 baseline 为主（用神基底优先），但保留 signalSelf 的存在感由 evidence 表达
		sig.Polarity = baseline
	}
}

// dayMasterStrengthLevel 加权身强弱评分（5 档），返回档位、评分与说明明细
// 权重规则：
//   - 月支与日干关系（得令）：×5
//   - 其余地支本气与日干关系（得地）：×3
//   - 藏干透出 + 天干生扶/克泄（得势）：×2
//
// 返回档位：vstrong / strong / neutral / weak / vweak
func dayMasterStrengthLevel(natal *BaziResult) (level string, score int, detail string) {
	if natal == nil {
		return "neutral", 0, ""
	}
	dayWx := ganWuxing[natal.DayGan]
	var details []string

	// 得令（月支） ×5
	mzWx := zhiWuxing[natal.MonthZhi]
	switch {
	case mzWx == dayWx:
		score += 5
		details = append(details, "月令同气+5")
	case wxSheng[mzWx] == dayWx:
		score += 5
		details = append(details, "月令生我+5")
	case wxKe[mzWx] == dayWx:
		score -= 5
		details = append(details, "月令克我-5")
	case wxSheng[dayWx] == mzWx:
		score -= 4
		details = append(details, "月令泄我-4")
	case wxKe[dayWx] == mzWx:
		score -= 2
		details = append(details, "月令受我克-2")
	}

	// 得地（其他三支本气） ×3
	for label, zhi := range map[string]string{"年支": natal.YearZhi, "日支": natal.DayZhi, "时支": natal.HourZhi} {
		zWx := zhiWuxing[zhi]
		switch {
		case zWx == dayWx:
			score += 3
			details = append(details, label+"同气+3")
		case wxSheng[zWx] == dayWx:
			score += 3
			details = append(details, label+"生我+3")
		case wxKe[zWx] == dayWx:
			score -= 3
			details = append(details, label+"克我-3")
		case wxSheng[dayWx] == zWx:
			score -= 2
			details = append(details, label+"泄我-2")
		case wxKe[dayWx] == zWx:
			score -= 1
			details = append(details, label+"受我克-1")
		}
	}

	// 得势（其余天干 + 藏干透出）×2
	for label, gan := range map[string]string{"年干": natal.YearGan, "月干": natal.MonthGan, "时干": natal.HourGan} {
		gWx := ganWuxing[gan]
		switch {
		case gWx == dayWx:
			score += 2
			details = append(details, label+"同气+2")
		case wxSheng[gWx] == dayWx:
			score += 2
			details = append(details, label+"生我+2")
		case wxKe[gWx] == dayWx:
			score -= 2
			details = append(details, label+"克我-2")
		case wxSheng[dayWx] == gWx:
			score -= 1
			details = append(details, label+"泄我-1")
		case wxKe[dayWx] == gWx:
			score -= 1
			details = append(details, label+"受我克-1")
		}
	}

	// 藏干透出（地支主气透出天干），加权 +1
	allTianGan := []string{natal.YearGan, natal.MonthGan, natal.DayGan, natal.HourGan}
	for label, zhi := range map[string]string{"年支": natal.YearZhi, "月支": natal.MonthZhi, "日支": natal.DayZhi, "时支": natal.HourZhi} {
		zhuQi := zhiZhuQi[zhi]
		if zhuQi == "" {
			continue
		}
		zhuQiWx := ganWuxing[zhuQi]
		// 主气是否透出到任一天干
		透出 := false
		for _, g := range allTianGan {
			if g == zhuQi {
				透出 = true
				break
			}
		}
		if !透出 {
			continue
		}
		switch {
		case zhuQiWx == dayWx:
			score += 1
			details = append(details, label+"藏干"+zhuQi+"透出+1")
		case wxSheng[zhuQiWx] == dayWx:
			score += 1
			details = append(details, label+"藏干"+zhuQi+"透出生我+1")
		case wxKe[zhuQiWx] == dayWx:
			score -= 1
			details = append(details, label+"藏干"+zhuQi+"透出克我-1")
		}
	}

	cfg := GetAlgoConfig()
	thresholds := cfg.ShenStrengthThresholds
	if thresholds.VStrong == 0 && thresholds.Strong == 0 && thresholds.Weak == 0 && thresholds.VWeak == 0 {
		thresholds = DefaultShenStrengthThresholds
	}

	switch {
	case score >= thresholds.VStrong:
		level = "vstrong"
	case score >= thresholds.Strong:
		level = "strong"
	case score <= thresholds.VWeak:
		level = "vweak"
	case score <= thresholds.Weak:
		level = "weak"
	default:
		level = "neutral"
	}

	detail = fmt.Sprintf("评分%d, 明细: %s", score, strings.Join(details, "/"))
	return
}

// dayMasterStrength 兼容旧签名，返回三档（strong/weak/neutral）
// 内部由 dayMasterStrengthLevel 折算
func dayMasterStrength(natal *BaziResult) string {
	level, _, _ := dayMasterStrengthLevel(natal)
	switch level {
	case "vstrong", "strong":
		return "strong"
	case "vweak", "weak":
		return "weak"
	default:
		return "neutral"
	}
}

// ─── 大运合化 ─────────────────────────────────────────────────────────────────

// HuaheStatus 大运合化状态
type HuaheStatus struct {
	Triggered bool   // 是否成立合化
	Combined  bool   // 是否构成五合（无论是否化）
	HuashenWx string // 化神五行（pinyin），合化成立时有值
	HuashenCN string // 化神五行（中文）
	Evidence  string // 描述文字
}

// detectDayunHuahe 检测某条大运天干与日干的合化情况
// 化神成立条件（合而成局）：
// 1. 大运天干与日干构成天干五合
// 2. 月支或大运地支提供化神五行根气（本气=化神）
// 3. 原局其他天干无强力反克化神（克化神且五行为化神之克者，至少一处）
func detectDayunHuahe(natal *BaziResult, dyGan, dyZhi string) HuaheStatus {
	if natal == nil || dyGan == "" {
		return HuaheStatus{}
	}
	huashen, ok := ganWuhe[[2]string{natal.DayGan, dyGan}]
	if !ok {
		return HuaheStatus{}
	}
	huashenCN := wxPinyin2CN[huashen]
	status := HuaheStatus{
		Combined:  true,
		HuashenWx: huashen,
		HuashenCN: huashenCN,
	}

	// 化神根气：月支本气 == 化神 OR 大运地支本气 == 化神
	monthZhiWx := zhiWuxing[natal.MonthZhi]
	dyZhiWx := zhiWuxing[dyZhi]
	hasRoot := (monthZhiWx == huashen) || (dyZhiWx == huashen)

	// 反克检测：化神之克者出现在原局其余天干（去日干外）
	huashenKeBy := ""
	for ke, target := range wxKe {
		if target == huashen {
			huashenKeBy = ke
			break
		}
	}
	hasFanke := false
	for _, g := range []string{natal.YearGan, natal.MonthGan, natal.HourGan} {
		if ganWuxing[g] == huashenKeBy {
			hasFanke = true
			break
		}
	}

	if hasRoot && !hasFanke {
		status.Triggered = true
		status.Evidence = fmt.Sprintf("大运%s%s与日干%s构成%s合，化神五行为%s；月支或大运地支提供根气，原局无强反克，合化成立",
			dyGan, dyZhi, natal.DayGan,
			natal.DayGan+dyGan, huashenCN)
	} else {
		// 合而不化
		reason := ""
		if !hasRoot {
			reason = "缺化神根气"
		} else if hasFanke {
			reason = "原局有反克"
		}
		status.Evidence = fmt.Sprintf("大运%s%s与日干%s构成%s合，但%s，合而不化，日干被合住但能量未转",
			dyGan, dyZhi, natal.DayGan, natal.DayGan+dyGan, reason)
	}
	return status
}

// ─── 神煞接入 ─────────────────────────────────────────────────────────────────

// getYearShensha 计算流年神煞（复用 GetDayunShenSha 公式，传入流年代替大运）
// 公式以原局年/月/日柱为基准，对目标天干地支查询，故对流年同样适用
func getYearShensha(natal *BaziResult, lnGan, lnZhi string) []string {
	if natal == nil {
		return nil
	}
	return GetDayunShenSha(natal.YearGan, natal.YearZhi, natal.MonthZhi, natal.DayGan, natal.DayZhi, lnGan, lnZhi)
}

// shenshaSignal 将一个神煞名转换为 EventSignal（白名单内才输出，外部已注明 baseline）
func shenshaSignal(name, baseline string) (EventSignal, bool) {
	meta, ok := shenshaWhitelist[name]
	if !ok {
		return EventSignal{}, false
	}
	sig := EventSignal{
		Type:     meta.Type,
		Evidence: meta.Hint,
		Source:   SourceShensha,
		Polarity: meta.Polarity,
	}
	// 神煞自带 polarity 优先；若与 baseline 显著不同，evidence 注明
	if baseline != "" && meta.Polarity != PolarityNeutral && meta.Polarity != baseline {
		sig.Evidence += "（与流年基底色" + baseline + "方向不同，需结合命主自身把握）"
	}
	return sig, true
}

// ─── 核心信号检测 ──────────────────────────────────────────────────────────────

// GetYearEventSignals 计算某一流年激活的事件信号
// age：流年虚岁；当 age < YoungAgeCutoff 时启用读书期语义重映射
func GetYearEventSignals(natal *BaziResult, lnGan, lnZhi, dayunGanZhi, gender string, age int) []EventSignal {
	if natal == nil {
		return nil
	}
	var signals []EventSignal
	seen := map[string]bool{}
	add := func(typ, evidence, signalSelf, source string) {
		if seen[typ] {
			return
		}
		seen[typ] = true
		sig := EventSignal{Type: typ, Evidence: evidence}
		applyPolarity(&sig, signalSelf, "", source)
		signals = append(signals, sig)
	}

	// 大运干支拆解（提前，供应期位置信号使用）
	var dyGan, dyZhi string
	dyRunes := []rune(dayunGanZhi)
	if len(dyRunes) >= 2 {
		dyGan = string(dyRunes[0])
		dyZhi = string(dyRunes[1])
	}

	// ── 应期位置信号（Layer 0：刑冲克合穿破原局用神/忌神位 + 三合/三会局势力）──
	layer0Sigs := collectYingqiSignals(natal, lnGan, lnZhi, dyGan, dyZhi)
	layer0Sigs = append(layer0Sigs, collectJuShiSignals(natal, lnZhi, dyZhi)...)
	layer0HasXiong := false
	for _, s := range layer0Sigs {
		if s.Polarity == PolarityXiong {
			layer0HasXiong = true
			break
		}
	}
	signals = append(signals, layer0Sigs...)
	layer0End := len(signals) // Layer 0 ends here; Layer 4+ signals begin after

	// addP: 添加信号，极性由 signalSelf 独立决定
	addP := func(typ, evidence, signalSelf, source string) {
		if seen[typ] {
			return
		}
		seen[typ] = true
		sig := EventSignal{Type: typ, Evidence: evidence}
		applyPolarity(&sig, signalSelf, "", source)
		signals = append(signals, sig)
	}
	_ = add // 保留兼容引用以避免未使用警告

	dayGan := natal.DayGan
	dayZhi := natal.DayZhi
	yearGan := natal.YearGan
	yearZhi := natal.YearZhi
	monthGan := natal.MonthGan
	monthZhi := natal.MonthZhi
	hourGan := natal.HourGan
	hourZhi := natal.HourZhi

	lnWx := ganWuxing[lnGan]
	dayWx := ganWuxing[dayGan]
	shishen := GetShiShen(dayGan, lnGan)
	strengthLevel, _, _ := dayMasterStrengthLevel(natal)
	// 兼容性身强弱（粗粒度）
	strength := dayMasterStrength(natal)
	_ = strengthLevel

	// 读书期判定（age < 18 时启用学业 / 性格语义重映射）
	isYoung := age > 0 && age < YoungAgeCutoff

	dyShishen := GetShiShen(dayGan, dyGan)
	dyZhiShishen := GetZhiShiShen(dayGan, dyZhi)

	existingZhi := []string{yearZhi, monthZhi, dayZhi, hourZhi}
	if dyZhi != "" {
		existingZhi = append(existingZhi, dyZhi)
	}

	// ── 大运合化检测 ─────────────────────────────────────────────────────────
	if dyGan != "" {
		hh := detectDayunHuahe(natal, dyGan, dyZhi)
		if hh.Combined {
			t := "大运合化"
			if !hh.Triggered {
				t = "综合变动"
			}
			addP(t, hh.Evidence, PolarityNeutral, SourceHehua)
		}
	}

	// ── 调候喜神透干 ────────────────────────────────────────────────────────
	if isTiaohouXi(lnGan, natal.Tiaohou) {
		addP("喜神临运", lnGan+"为调候喜神（《穷通宝鉴》所取），该年喜神透干，全局运势有明显助力", PolarityJi, SourceYongshen)
	}

	// ── 大运地支 × 流年地支 关系 ──────────────────────────────────────────
	if dyZhi != "" {
		if chong, ok := sixChong[dyZhi]; ok && chong == lnZhi {
			addP("综合变动", "大运地支"+dyZhi+"与流年地支"+lnZhi+"相冲，大运流年地支双冲，本年为重大事件高发年，各类信号均被放大", PolarityXiong, SourceZhuwei)
		}
		if he, ok := sixHe[dyZhi]; ok && he == lnZhi {
			addP("综合变动", "大运地支"+dyZhi+"与流年地支"+lnZhi+"六合，大运流年地支相合，能量聚合，该年事件易有正向突破", PolarityJi, SourceZhuwei)
		}
		if chong, ok := sixChong[dyZhi]; ok && chong == dayZhi {
			if isYoung {
				addP(TypeXingGePanNi, "大运地支"+dyZhi+"冲日支"+dayZhi+"（自我宫位），少年期家庭关系紧张、自我意识觉醒，情绪波动较大", PolarityXiong, SourceZhuwei)
			} else {
				addP("婚恋_冲", "大运地支"+dyZhi+"冲日支"+dayZhi+"（夫妻宫），整个大运期间感情宫位持续震动，本年流年触发更易产生感情重大变化", PolarityXiong, SourceZhuwei)
			}
		}
		if he, ok := sixHe[dyZhi]; ok && he == dayZhi {
			if isYoung {
				addP(TypeXingGeQingYi, "大运地支"+dyZhi+"合住日支"+dayZhi+"（自我宫位），少年期同窗情谊深厚，性格塑造期人际和谐", PolarityJi, SourceZhuwei)
			} else {
				addP("婚恋_合", "大运地支"+dyZhi+"合住日支"+dayZhi+"（夫妻宫），大运期间感情宫位被激活，该年感情事件易有进展", PolarityJi, SourceZhuwei)
			}
		}
	}

	// ── 婚恋信号（财官透干） ─────────────────────────────────────────────────
	dyFinanceDouble := (dyShishen == "偏财" || dyShishen == "正财") &&
		(dyZhiShishen == "偏财" || dyZhiShishen == "正财")
	dyOfficialDouble := (dyShishen == "正官" || dyShishen == "七杀") &&
		(dyZhiShishen == "正官" || dyZhiShishen == "七杀")

	// 男财/女官透干 → 婚恋_合：成人期专属（少年期不输出，由学业/性格分支独立覆盖）
	if !isYoung {
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
				addP("婚恋_合", evidence, PolarityJi, SourceZhuwei)
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
				addP("婚恋_合", evidence, PolarityJi, SourceZhuwei)
			}
		}
	}

	// 流年地支与日支六合 / 六冲
	if he, ok := sixHe[lnZhi]; ok && he == dayZhi {
		if isYoung {
			addP(TypeXingGeQingYi, "流年地支"+lnZhi+"与日支"+dayZhi+"（自我宫位）六合，少年期同窗情谊深、心意相通", PolarityJi, SourceZhuwei)
		} else {
			addP("婚恋_合", "流年地支"+lnZhi+"与日支"+dayZhi+"（夫妻宫）六合，感情宫位被激活", PolarityJi, SourceZhuwei)
		}
	}
	if chong, ok := sixChong[lnZhi]; ok && chong == dayZhi {
		if isYoung {
			addP(TypeXingGePanNi, "流年地支"+lnZhi+"冲日支"+dayZhi+"（自我宫位），少年期情绪波动 / 自我意识与家庭关系紧张", PolarityXiong, SourceZhuwei)
		} else {
			addP("婚恋_冲", "流年地支"+lnZhi+"冲日支"+dayZhi+"（夫妻宫），感情宫位受震动，关系或有重大变化", PolarityXiong, SourceZhuwei)
		}
	}

	// 桃花临命（以基础规则；shensha 引擎中更精细的桃花会另行 add）
	if isTaohuaBase(yearZhi, lnZhi) || isTaohuaBase(dayZhi, lnZhi) {
		if isYoung {
			addP(TypeXingGeQingYi, "流年地支"+lnZhi+"为桃花星临命，少年期人缘旺 / 异性缘萌动 / 同窗喜事多", PolarityNeutral, SourceZhuwei)
		} else {
			addP("婚恋_合", "流年地支"+lnZhi+"为桃花星临命，人缘异性缘大旺", PolarityNeutral, SourceZhuwei)
		}
	}

	// 三合局引动夫妻宫
	if ok, sanheWx := isSanheTriggered(lnZhi, existingZhi); ok {
		var targetWx string
		if gender == "male" {
			targetWx = wxKe[dayWx]
		} else {
			for w, kTarget := range wxKe {
				if kTarget == dayWx {
					targetWx = w
					break
				}
			}
		}
		if sanheWx == targetWx {
			if isYoung {
				addP(TypeXingGeQingYi, "流年地支"+lnZhi+"引动三合局（"+wxPinyin2CN[sanheWx]+"），少年期同窗友情或异性缘渐显", PolarityJi, SourceHui)
			} else {
				addP("婚恋_合", "流年地支"+lnZhi+"引动三合局（"+wxPinyin2CN[sanheWx]+"），感情星五行局成，婚恋机遇显著增强", PolarityJi, SourceHui)
			}
		} else {
			addP("综合变动", "流年地支"+lnZhi+"引动三合"+wxPinyin2CN[sanheWx]+"局，能量聚合", PolarityNeutral, SourceHui)
		}
	}

	// 三会局引动
	if ok, sanhuiWx := isSanhuiTriggered(lnZhi, existingZhi); ok {
		var targetWx string
		if gender == "male" {
			targetWx = wxKe[dayWx]
		} else {
			for w, kTarget := range wxKe {
				if kTarget == dayWx {
					targetWx = w
					break
				}
			}
		}
		if sanhuiWx == targetWx {
			if isYoung {
				addP(TypeXingGeQingYi, "流年地支"+lnZhi+"引动三会"+wxPinyin2CN[sanhuiWx]+"局，少年期人际格局剧变，性格塑造期人际能量极强", PolarityJi, SourceHui)
			} else {
				addP("婚恋_合", "流年地支"+lnZhi+"引动三会"+wxPinyin2CN[sanhuiWx]+"局，能量极猛，感情/官财动象强烈", PolarityJi, SourceHui)
			}
		} else {
			addP("综合变动", "流年地支"+lnZhi+"引动三会"+wxPinyin2CN[sanhuiWx]+"局，能量场剧烈，本年事件力度强", PolarityNeutral, SourceHui)
		}
	}

	// ── 事业信号 ───────────────────────────────────────────────────────────
	isOfficialStar := shishen == "正官" || shishen == "七杀"
	isSealStar := shishen == "正印" || shishen == "偏印"

	if isOfficialStar {
		if isYoung {
			switch strength {
			case "weak":
				addP(TypeXueYeYaLi, lnGan+"透干为"+shishen+"，少年身弱遇官杀，考试 / 升学 / 老师管教带来较大学业压力", PolarityXiong, SourceZhuwei)
			case "strong":
				addP(TypeXueYeYaLi, lnGan+"透干为"+shishen+"，少年身旺，官星显达，班干 / 学校职务 / 突出表现的机会", PolarityJi, SourceZhuwei)
			default:
				addP(TypeXueYeYaLi, lnGan+"透干为"+shishen+"，少年期官星临运，学业上有规则约束或重大考核", PolarityNeutral, SourceZhuwei)
			}
			// 官杀双叠：成人期映射至"事业贯穿"，少年期不再独立加注（已由 学业_压力 表达）
		} else {
			switch strength {
			case "weak":
				addP("事业", lnGan+"透干为"+shishen+"，日主身弱，官杀压力较大，事业有阻力或竞争加剧", PolarityXiong, SourceZhuwei)
			case "strong":
				addP("事业", lnGan+"透干为"+shishen+"，日主身旺，官星有力，事业晋升或仕途机遇来临", PolarityJi, SourceZhuwei)
			default:
				addP("事业", lnGan+"透干为"+shishen+"，官星临运，事业格局有变动", PolarityNeutral, SourceZhuwei)
			}
			dyHasOfficial := dyShishen == "正官" || dyShishen == "七杀" || dyZhiShishen == "正官" || dyZhiShishen == "七杀"
			if dyHasOfficial && !seen["事业_叠"] {
				seen["事业_叠"] = true
				ev := "大运流年官杀双叠（大运" + dyGan + "/" + dyZhi + "，流年" + lnGan + "），仕途压力或机遇贯穿整年"
				if dyOfficialDouble {
					ev = "大运" + dayunGanZhi + "干支均为官杀，整个大运官杀场域强劲，流年再叠，本年事业变动力度极大"
				}
				signals = append(signals, EventSignal{Type: "事业", Evidence: ev, Polarity: PolarityNeutral, Source: SourceZhuwei})
			}
		}
	}

	if isSealStar {
		if isYoung {
			addP(TypeXueYeGuiRen, lnGan+"透干为"+shishen+"，少年期印星护身，得师长指点 / 学习方法突破 / 资格认证机会", PolarityJi, SourceZhuwei)
		} else {
			addP("事业", lnGan+"透干为"+shishen+"，印星护身生扶日主，利于考试晋升、资格认证或获贵人提携", PolarityJi, SourceZhuwei)
		}
	}

	// ── 财运信号 ───────────────────────────────────────────────────────────
	isFinanceStar := shishen == "偏财" || shishen == "正财"
	if isFinanceStar {
		// 财星 vs 用忌神：若财星五行为忌神，polarity 翻为凶
		caiWx := wxKe[dayWx]
		caiWxCN := wxPinyin2CN[caiWx]
		caiIsJi := natal != nil && strings.Contains(natal.Jishen, caiWxCN)
		if isYoung {
			if caiIsJi {
				addP(TypeXueYeZiYuan, lnGan+"透干为"+shishen+"，少年期财星为忌，家庭经济波动或物质过盛分散学习注意力，宜守心向学", PolarityXiong, SourceZhuwei)
			} else if strength == "weak" {
				addP(TypeXueYeZiYuan, lnGan+"透干为"+shishen+"，少年期家境/零用钱有信号但身弱难任，宜量力规划学习投入", PolarityNeutral, SourceZhuwei)
			} else {
				addP(TypeXueYeZiYuan, lnGan+"透干为"+shishen+"，少年期家境改善 / 物质条件提升，可适度规划学习投入", PolarityJi, SourceZhuwei)
			}
			// 双叠：少年期不输出"婚恋年"或"财运爆发"语义
		} else {
			if caiIsJi {
				addP("财运_得", lnGan+"透干为"+shishen+"，但财星五行（"+caiWxCN+"）为命主忌神，财来财去/破耗，宜守不宜攻", PolarityXiong, SourceZhuwei)
			} else {
				if strength == "weak" {
					addP("财运_得", lnGan+"透干为"+shishen+"，财星现身，但日主身弱财多身弱，宜量力而为", PolarityNeutral, SourceZhuwei)
				} else {
					addP("财运_得", lnGan+"透干为"+shishen+"，财星透出，财运有望提升，宜主动把握进财机会", PolarityJi, SourceZhuwei)
				}
			}
			dyHasFinance := dyShishen == "偏财" || dyShishen == "正财" || dyZhiShishen == "偏财" || dyZhiShishen == "正财"
			if dyHasFinance && !seen["财运_叠"] {
				seen["财运_叠"] = true
				ev := "大运流年财星双叠（大运" + dyGan + "/" + dyZhi + "，流年" + lnGan + "），财运进项力度强，但需防比劫争财"
				if dyFinanceDouble {
					ev = "大运" + dayunGanZhi + "干支均为财星，整个大运财星旺盛，流年财星再叠，财运爆发年"
				}
				ev2pol := PolarityJi
				if caiIsJi {
					ev2pol = PolarityXiong
					ev += "（财为忌神，大叠反主大破耗）"
				}
				signals = append(signals, EventSignal{Type: "财运_得", Evidence: ev, Polarity: ev2pol, Source: SourceZhuwei})
			}
		}
	}
	if shishen == "比肩" || shishen == "劫财" {
		if isYoung {
			pol := PolarityXiong
			ev := lnGan + "透干为" + shishen + "，少年期同学竞争 / 友谊摩擦显著，宜以平常心相处"
			if strength == "weak" {
				pol = PolarityJi
				ev = lnGan + "透干为" + shishen + "，少年身弱遇比劫，得同伴帮扶 / 团体支持，宜借力同盟"
			}
			addP(TypeXueYeJingZheng, ev, pol, SourceZhuwei)
		} else {
			// 身弱时比劫为帮身（吉）；身强时为夺财（凶）
			pol := PolarityXiong
			ev := lnGan + "透干为" + shishen + "，比劫争财，财运有损耗风险，投资需谨慎"
			if strength == "weak" {
				pol = PolarityJi
				ev = lnGan + "透干为" + shishen + "，日主身弱，比劫帮身有助，宜借力同盟"
			}
			addP("财运_损", ev, pol, SourceZhuwei)
		}
	}
	if shishen == "食神" || shishen == "伤官" {
		if isYoung {
			pol := PolarityJi
			ev := lnGan + "透干为" + shishen + "，少年期才艺特长 / 表达欲展现，宜参与兴趣活动"
			if strength == "weak" {
				pol = PolarityXiong
				ev = lnGan + "透干为" + shishen + "，少年身弱遇食伤，过度投入兴趣致分心 / 操劳，宜量力而行"
			}
			addP(TypeXueYeCaiYi, ev, pol, SourceZhuwei)
		} else {
			// 身强食伤洩秀（吉）；身弱食伤洩气（凶）
			pol := PolarityJi
			ev := lnGan + "透干为" + shishen + "，食伤生财，技艺才华有望变现，适合创业或副业尝试"
			if strength == "weak" {
				pol = PolarityXiong
				ev = lnGan + "透干为" + shishen + "，日主身弱，食伤洩气过度反主操劳损耗，宜量力而行"
			}
			addP("财运_得", ev, pol, SourceZhuwei)
		}
	}

	// ── 健康信号 ───────────────────────────────────────────────────────────
	if wxKe[lnWx] == dayWx {
		addP("健康", "流年天干"+lnGan+"（"+wxPinyin2CN[lnWx]+"）克制日干"+dayGan+"（"+wxPinyin2CN[dayWx]+"），日主元气受损，需注意身体健康", PolarityXiong, SourceZhuwei)
	}
	if chong, ok := sixChong[lnZhi]; ok && chong == dayZhi {
		addP("健康", "流年地支"+lnZhi+"冲日支"+dayZhi+"，日柱受冲，体力精神有下滑风险", PolarityXiong, SourceZhuwei)
	}
	if chong, ok := sixChong[lnZhi]; ok && chong == yearZhi {
		addP("健康", "流年地支"+lnZhi+"冲年支"+yearZhi+"，岁破临命，需防突发意外或家庭变故", PolarityXiong, SourceZhuwei)
	}
	if xing, ok := sixXing[lnZhi]; ok && xing == dayZhi {
		addP("健康", "流年地支"+lnZhi+"刑日支"+dayZhi+"，地支相刑，易有手术、伤病或官非之虞", PolarityXiong, SourceXing)
	}
	if selfXing[lnZhi] && lnZhi == dayZhi {
		addP("健康", "流年地支"+lnZhi+"与日支同支自刑，精神压力较大，需防积劳成疾", PolarityXiong, SourceXing)
	}

	// ── 三刑全局 ────────────────────────────────────────────────────────────
	if ok, kind := isSanxingTriggered(lnZhi, existingZhi); ok {
		ev := "流年补足" + kind + "三刑全局，主官非/手术/伤病/纠葛"
		if kind == "丑未戌" {
			ev = "流年补足丑未戌三刑（无恩之刑），主家庭/事业纠葛"
		}
		// 复用 健康 槽位时若已 seen，则改 Type 为 综合变动
		if seen["健康"] {
			signals = append(signals, EventSignal{Type: "综合变动", Evidence: ev, Polarity: PolarityXiong, Source: SourceXing})
		} else {
			addP("健康", ev, PolarityXiong, SourceXing)
		}
	}

	// ── 迁变信号 ────────────────────────────────────────────────────────────
	if isYimaBase(yearZhi, lnZhi) || isYimaBase(dayZhi, lnZhi) {
		addP("迁变", "流年地支"+lnZhi+"为驿马星，主奔波变动、出行迁移或职位调动", PolarityNeutral, SourceZhuwei)
	}
	if chong, ok := sixChong[lnZhi]; ok && chong == yearZhi {
		// 既有"健康(岁破)"信号已 add；此处补"迁变"角度
		if !seen["迁变"] {
			addP("迁变", "流年地支"+lnZhi+"冲年支"+yearZhi+"，岁破之年，人生格局易有较大转变", PolarityXiong, SourceZhuwei)
		}
	}

	// ── 流年与年/月/时柱互动（合冲，事业/根基/晚景） ───────────────────────
	// 与年支：六合 → 综合变动（家族/根基喜事）；冲已在岁破处理
	if he, ok := sixHe[lnZhi]; ok && he == yearZhi {
		if !seen["综合变动"] {
			addP("综合变动", "流年地支"+lnZhi+"合年支"+yearZhi+"（祖荫/根基），家族/根基方面易有正向事件", PolarityJi, SourceZhuwei)
		}
	}
	// 与月支：六冲 → 事业 / 学业方向变动；六合 → 行业 / 师承关系融洽
	if chong, ok := sixChong[lnZhi]; ok && chong == monthZhi {
		if isYoung {
			addP(TypeXueYeYaLi, "流年地支"+lnZhi+"冲月柱"+monthGan+monthZhi+"（提纲），少年期学业方向 / 学校 / 重要科目调整压力增大", PolarityXiong, SourceZhuwei)
		} else if !seen["事业"] {
			addP("事业", "流年地支"+lnZhi+"冲月柱"+monthGan+monthZhi+"（提纲），易有行业/职位变动", PolarityXiong, SourceZhuwei)
		} else {
			signals = append(signals, EventSignal{
				Type: "事业", Evidence: "流年地支" + lnZhi + "冲月柱" + monthGan + monthZhi + "（提纲），易有行业/职位变动",
				Polarity: PolarityXiong, Source: SourceZhuwei,
			})
		}
	}
	if he, ok := sixHe[lnZhi]; ok && he == monthZhi {
		if isYoung {
			addP(TypeXueYeGuiRen, "流年地支"+lnZhi+"合月支"+monthZhi+"（提纲），少年期师生关系融洽 / 学习方向得力", PolarityJi, SourceZhuwei)
		} else if !seen["事业"] {
			addP("事业", "流年地支"+lnZhi+"合月支"+monthZhi+"（提纲），行业/职业关系易有融洽进展", PolarityJi, SourceZhuwei)
		}
	}
	// 与时支：六合 → 子女/晚景喜事；六冲 → 子女/晚景动象
	if he, ok := sixHe[lnZhi]; ok && he == hourZhi {
		signals = append(signals, EventSignal{
			Type: "综合变动", Evidence: "流年地支" + lnZhi + "合时柱" + hourGan + hourZhi + "（子女/晚景宫）",
			Polarity: PolarityJi, Source: SourceZhuwei,
		})
	}
	if chong, ok := sixChong[lnZhi]; ok && chong == hourZhi {
		signals = append(signals, EventSignal{
			Type: "综合变动", Evidence: "流年地支" + lnZhi + "冲时柱" + hourGan + hourZhi + "（子女/晚景宫），子女或晚景方面有动象",
			Polarity: PolarityXiong, Source: SourceZhuwei,
		})
	}

	// ── 伏吟 / 反吟 ─────────────────────────────────────────────────────────
	pillars := []struct{ label, gan, zhi string }{
		{"年柱", yearGan, yearZhi},
		{"月柱", monthGan, monthZhi},
		{"日柱", dayGan, dayZhi},
		{"时柱", hourGan, hourZhi},
	}
	if dyGan != "" {
		pillars = append(pillars, struct{ label, gan, zhi string }{"大运", dyGan, dyZhi})
	}
	for _, p := range pillars {
		if isFuyin(lnGan, lnZhi, p.gan, p.zhi) {
			ev := fmt.Sprintf("流年%s%s伏吟%s%s%s，主同类事件重现/旧事重提", lnGan, lnZhi, p.label, p.gan, p.zhi)
			signals = append(signals, EventSignal{Type: "伏吟", Evidence: ev, Polarity: PolarityXiong, Source: SourceFuyin})
		}
		if isFanyin(lnGan, lnZhi, p.gan, p.zhi) {
			ev := fmt.Sprintf("流年%s%s反吟%s%s%s（天克地冲），主剧烈变动", lnGan, lnZhi, p.label, p.gan, p.zhi)
			signals = append(signals, EventSignal{Type: "反吟", Evidence: ev, Polarity: PolarityXiong, Source: SourceFuyin})
		}
	}

	// ── 空亡（按日柱旬空） ─────────────────────────────────────────────────
	xk1, xk2 := getXunkong(dayGan, dayZhi)
	if xk1 != "" {
		if lnZhi == xk1 || lnZhi == xk2 {
			signals = append(signals, EventSignal{
				Type:     "综合变动",
				Evidence: fmt.Sprintf("流年地支%s落日柱旬空（%s%s空），事件虚而不实/过而不留，本年其他信号力度减半", lnZhi, xk1, xk2),
				Polarity: PolarityNeutral,
				Source:   SourceKongwang,
			})
			// 在已有信号 evidence 上注明降权
			for i := range signals {
				if signals[i].Source == SourceKongwang || signals[i].Source == SourceYongshen {
					continue
				}
				if !strings.Contains(signals[i].Evidence, "受空亡影响") {
					signals[i].Evidence += "（受空亡影响，力度减半）"
				}
			}
		}
		if dyZhi != "" && (dyZhi == xk1 || dyZhi == xk2) {
			signals = append(signals, EventSignal{
				Type:     "综合变动",
				Evidence: fmt.Sprintf("大运地支%s落日柱旬空（%s%s空），整段大运能量打折", dyZhi, xk1, xk2),
				Polarity: PolarityNeutral,
				Source:   SourceKongwang,
			})
		}
	}

	// ── 神煞接入（含重煞/轻煞分级）────────────────────────────────────────────
	type shenshaSigWithMeta struct {
		sig     EventSignal
		isHeavy bool
	}
	shenshaList := getYearShensha(natal, lnGan, lnZhi)
	var shenshaSigs []shenshaSigWithMeta
	for _, name := range shenshaList {
		ssSig, ok := shenshaSignal(name, "")
		if !ok {
			continue
		}
		meta := shenshaWhitelist[name]
		shenshaSigs = append(shenshaSigs, shenshaSigWithMeta{ssSig, meta.IsHeavy})
	}
	// 检测本年是否存在重煞凶信号
	hasHeavyXiong := false
	for _, ss := range shenshaSigs {
		if ss.isHeavy && ss.sig.Polarity == PolarityXiong {
			hasHeavyXiong = true
			break
		}
	}
	// 重煞出现时，轻煞吉信号追加降级注释（Polarity 不变）
	for i := range shenshaSigs {
		if hasHeavyXiong && !shenshaSigs[i].isHeavy && shenshaSigs[i].sig.Polarity == PolarityJi {
			shenshaSigs[i].sig.Evidence += "（本年有重煞，此信号仅作参考）"
		}
		signals = append(signals, shenshaSigs[i].sig)
	}

	// ── 少年期：将神煞输出的成人期 Type 重映射至学业 / 性格 ─────────────────────
	if isYoung {
		for i := range signals {
			if signals[i].Source != SourceShensha {
				continue
			}
			switch signals[i].Type {
			case "婚恋_合", "婚恋_冲", "婚恋_变":
				signals[i].Type = TypeXingGeQingYi
			case "财运_得", "财运_损":
				signals[i].Type = TypeXueYeZiYuan
			case "事业":
				if signals[i].Polarity == PolarityXiong {
					signals[i].Type = TypeXueYeYaLi
				} else {
					signals[i].Type = TypeXueYeGuiRen
				}
			}
		}
	}

	// ── 合冲并见对消 ───────────────────────────────────────────────────────
	if seen["婚恋_冲"] && seen["婚恋_合"] {
		filtered := signals[:0]
		var ev []string
		for _, s := range signals {
			if s.Type == "婚恋_合" || s.Type == "婚恋_冲" {
				ev = append(ev, s.Evidence)
				continue
			}
			filtered = append(filtered, s)
		}
		signals = append(filtered, EventSignal{
			Type:     "婚恋_变",
			Evidence: "感情合冲交织：" + strings.Join(ev, "；") + "，方向不定，事件张力大",
			Polarity: PolarityXiong,
			Source:   SourceZhuwei,
		})
	}

	// ── Layer 0 凶压制 Layer 4 吉（末尾后处理）──────────────────────────────
	// Layer 0 有凶信号时，过滤掉 Layer 4（i >= layer0End）中 Source=柱位互动 且 Polarity=吉 的信号
	if layer0HasXiong {
		kept := make([]EventSignal, 0, len(signals))
		for i, s := range signals {
			if i >= layer0End && s.Source == SourceZhuwei && s.Polarity == PolarityJi {
				continue
			}
			kept = append(kept, s)
		}
		signals = kept
	}

	return signals
}

// GetAllYearSignals 批量扫描全部流年事件信号（含过往与未来）
func GetAllYearSignals(result *BaziResult, gender string, currentYear, minAge int) []YearSignals {
	var out []YearSignals
	for _, dy := range result.Dayun {
		dyGanZhi := dy.Gan + dy.Zhi
		for _, ln := range dy.LiuNian {
			if ln.Age < minAge {
				continue
			}
			lnRunes := []rune(ln.GanZhi)
			if len(lnRunes) < 2 {
				continue
			}
			sigs := GetYearEventSignals(result, string(lnRunes[0]), string(lnRunes[1]), dyGanZhi, gender, ln.Age)
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

// CollectDayunHuaheMap 汇总命盘中所有大运的合化标签 → map[gz]label，供 Service 层按大运 lookup
func CollectDayunHuaheMap(natal *BaziResult) map[string]string {
	out := map[string]string{}
	if natal == nil {
		return out
	}
	for _, dy := range natal.Dayun {
		hh := detectDayunHuahe(natal, dy.Gan, dy.Zhi)
		if !hh.Combined {
			continue
		}
		gz := dy.Gan + dy.Zhi
		if hh.Triggered {
			out[gz] = "合化成立(" + hh.HuashenCN + ")"
		} else {
			out[gz] = "合而不化"
		}
	}
	return out
}

// CollectDayunHuaheLines 汇总命盘中所有大运的合化标签（供 Prompt 使用）
func CollectDayunHuaheLines(natal *BaziResult) []string {
	if natal == nil {
		return nil
	}
	var lines []string
	for _, dy := range natal.Dayun {
		hh := detectDayunHuahe(natal, dy.Gan, dy.Zhi)
		if hh.Combined {
			tag := "合而不化"
			if hh.Triggered {
				tag = "合化成立(" + hh.HuashenCN + ")"
			}
			lines = append(lines, fmt.Sprintf("第%d步大运 %s%s（%d-%d岁）：%s",
				dy.Index, dy.Gan, dy.Zhi, dy.StartAge, dy.StartAge+9, tag))
		}
	}
	return lines
}

// GetStrengthDetail 公开身强弱明细（供 Prompt 与 service 使用）
func GetStrengthDetail(natal *BaziResult) (level string, score int, detail string) {
	return dayMasterStrengthLevel(natal)
}
