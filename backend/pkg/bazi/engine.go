// Package bazi 实现八字命理计算引擎
// 基于 github.com/6tail/lunar-go 天文历法库，精确到秒级节气计算
package bazi

import (
	"container/list"
	"crypto/md5"
	"fmt"

	"github.com/6tail/lunar-go/LunarUtil"
	"github.com/6tail/lunar-go/calendar"
)

// 天干
var Gan = []string{"甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"}

// 地支
var Zhi = []string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}

// 五行名称
var WuxingName = []string{"木", "火", "土", "金", "水"}

// BaziResult 八字计算结果
type BaziResult struct {
	YearGan  string `json:"year_gan"`
	YearZhi  string `json:"year_zhi"`
	MonthGan string `json:"month_gan"`
	MonthZhi string `json:"month_zhi"`
	DayGan   string `json:"day_gan"`
	DayZhi   string `json:"day_zhi"`
	HourGan  string `json:"hour_gan"`
	HourZhi  string `json:"hour_zhi"`

	YearGanWuxing  string `json:"year_gan_wuxing"`
	YearZhiWuxing  string `json:"year_zhi_wuxing"`
	MonthGanWuxing string `json:"month_gan_wuxing"`
	MonthZhiWuxing string `json:"month_zhi_wuxing"`
	DayGanWuxing   string `json:"day_gan_wuxing"`
	DayZhiWuxing   string `json:"day_zhi_wuxing"`
	HourGanWuxing  string `json:"hour_gan_wuxing"`
	HourZhiWuxing  string `json:"hour_zhi_wuxing"`

	// 十神（天干与地支）
	YearGanShiShen  string `json:"year_gan_shishen"`
	MonthGanShiShen string `json:"month_gan_shishen"`
	DayGanShiShen   string `json:"day_gan_shishen"`
	HourGanShiShen  string `json:"hour_gan_shishen"`

	YearZhiShiShen  []string `json:"year_zhi_shishen"`
	MonthZhiShiShen []string `json:"month_zhi_shishen"`
	DayZhiShiShen   []string `json:"day_zhi_shishen"`
	HourZhiShiShen  []string `json:"hour_zhi_shishen"`

	// 十二长生（其实这是日主对当前地支的星运，保留原名以兼容前端）
	YearDiShi  string `json:"year_di_shi"`
	MonthDiShi string `json:"month_di_shi"`
	DayDiShi   string `json:"day_di_shi"`
	HourDiShi  string `json:"hour_di_shi"`

	// 星运（实际语义同 DiShi，前端规范化后可使用）
	YearXingYun  string `json:"year_xing_yun"`
	MonthXingYun string `json:"month_xing_yun"`
	DayXingYun   string `json:"day_xing_yun"`
	HourXingYun  string `json:"hour_xing_yun"`

	// 旬空（空亡）
	YearXunKong  string `json:"year_xun_kong"`
	MonthXunKong string `json:"month_xun_kong"`
	DayXunKong   string `json:"day_xun_kong"`
	HourXunKong  string `json:"hour_xun_kong"`

	// 神煞
	YearShenSha  []string `json:"year_shen_sha"`
	MonthShenSha []string `json:"month_shen_sha"`
	DayShenSha   []string `json:"day_shen_sha"`
	HourShenSha  []string `json:"hour_shen_sha"`

	// 地支藏干
	YearHideGan  []string `json:"year_hide_gan"`
	MonthHideGan []string `json:"month_hide_gan"`
	DayHideGan   []string `json:"day_hide_gan"`
	HourHideGan  []string `json:"hour_hide_gan"`

	// 纳音
	YearNaYin  string `json:"year_na_yin"`
	MonthNaYin string `json:"month_na_yin"`
	DayNaYin   string `json:"day_na_yin"`
	HourNaYin  string `json:"hour_na_yin"`

	Wuxing WuxingStats `json:"wuxing"`

	// 用神/忌神由 LLM 推断，此处保留字段兼容性，值为空
	Yongshen string `json:"yongshen"`
	Jishen   string `json:"jishen"`

	// 调候用神（基于《穷通宝鉴》查表精算）
	Tiaohou *TiaohouResult `json:"tiaohou"`

	Dayun         []DayunItem `json:"dayun"`
	StartYunSolar string      `json:"start_yun_solar"` // 例如："1995年4月5日 14:30"
	Gender        string      `json:"gender"`

	BirthYear  int `json:"birth_year"`
	BirthMonth int `json:"birth_month"`
	BirthDay   int `json:"birth_day"`
	BirthHour  int `json:"birth_hour"`

	// 真太阳时（经度修正后）
	TrueSolarHour   int `json:"true_solar_hour"`
	TrueSolarMinute int `json:"true_solar_minute"`

	ChartHash string `json:"chart_hash"`
}

type WuxingStats struct {
	Mu   int `json:"mu"`
	Huo  int `json:"huo"`
	Tu   int `json:"tu"`
	Jin  int `json:"jin"`
	Shui int `json:"shui"`

	Total   int     `json:"total"`
	MuPct   float64 `json:"mu_pct"`
	HuoPct  float64 `json:"huo_pct"`
	TuPct   float64 `json:"tu_pct"`
	JinPct  float64 `json:"jin_pct"`
	ShuiPct float64 `json:"shui_pct"`
}

type LiuNianItem struct {
	Year         int    `json:"year"`
	Age          int    `json:"age"`
	GanZhi       string `json:"gan_zhi"`
	GanShiShen   string `json:"gan_shishen"`
	ZhiShiShen   string `json:"zhi_shishen"`
	IsTransition bool   `json:"is_transition"`
	TransMonth   int    `json:"trans_month,omitempty"`
	TransDay     int    `json:"trans_day,omitempty"`
	PrevDayun    string `json:"prev_dayun,omitempty"`
}

type DayunItem struct {
	Index      int           `json:"index"`
	Gan        string        `json:"gan"`
	Zhi        string        `json:"zhi"`
	StartAge   int           `json:"start_age"`
	StartYear  int           `json:"start_year"`
	EndYear    int           `json:"end_year"`
	GanShiShen string        `json:"gan_shishen"`
	ZhiShiShen string        `json:"zhi_shishen"`
	DiShi      string        `json:"di_shi"`
	LiuNian    []LiuNianItem `json:"liu_nian"`
}

// Calculate 计算八字四柱
// year, month, day: 公历年月日
// hour: 北京时间小时（0-23）
// gender: "male" | "female"
// isEarlyZishi: true=早子时（23:00 属前一天）
// longitude: 出生地经度，用于真太阳时修正，0 表示不修正（按北京时间）
func Calculate(year, month, day, hour int, gender string, isEarlyZishi bool, longitude float64) *BaziResult {
	// 真太阳时修正：北京时间基于东经 120°，每差 1° 差 4 分钟
	trueSolarMinuteOffset := 0
	if longitude != 0 {
		trueSolarMinuteOffset = int((longitude - 120.0) * 4)
	}

	// 计算真太阳时的小时和分钟
	totalMinutes := hour*60 + trueSolarMinuteOffset
	// 处理越界
	for totalMinutes < 0 {
		totalMinutes += 24 * 60
	}
	totalMinutes = totalMinutes % (24 * 60)
	tsHour := totalMinutes / 60
	tsMinute := totalMinutes % 60

	// 早子时：23:00 属前一日子时
	calcHour := tsHour
	calcMinute := tsMinute
	calcDay := day
	calcMonth := month
	calcYear := year
	if isEarlyZishi && hour == 23 {
		// 日柱按前一日，时辰仍为子（0）
		solar0 := calendar.NewSolar(year, month, day, 0, 0, 0)
		solar0 = solar0.NextDay(-1)
		calcYear = solar0.GetYear()
		calcMonth = solar0.GetMonth()
		calcDay = solar0.GetDay()
		calcHour = 23
		calcMinute = 0
	}

	// 构造 Solar 并获取 Lunar/EightChar
	solar := calendar.NewSolar(calcYear, calcMonth, calcDay, calcHour, calcMinute, 0)
	lunar := solar.GetLunar()
	bz := lunar.GetEightChar()

	// 提取天干地支
	yearGan := bz.GetYearGan()
	yearZhi := bz.GetYearZhi()
	monthGan := bz.GetMonthGan()
	monthZhi := bz.GetMonthZhi()
	dayGan := bz.GetDayGan()
	dayZhi := bz.GetDayZhi()
	hourGan := bz.GetTimeGan()
	hourZhi := bz.GetTimeZhi()

	// 五行
	yearGanWx := LunarUtil.WU_XING_GAN[yearGan]
	yearZhiWx := LunarUtil.WU_XING_ZHI[yearZhi]
	monthGanWx := LunarUtil.WU_XING_GAN[monthGan]
	monthZhiWx := LunarUtil.WU_XING_ZHI[monthZhi]
	dayGanWx := LunarUtil.WU_XING_GAN[dayGan]
	dayZhiWx := LunarUtil.WU_XING_ZHI[dayZhi]
	hourGanWx := LunarUtil.WU_XING_GAN[hourGan]
	hourZhiWx := LunarUtil.WU_XING_ZHI[hourZhi]

	// 五行统计
	wuxing := calcWuxingFromNames(
		yearGanWx, yearZhiWx,
		monthGanWx, monthZhiWx,
		dayGanWx, dayZhiWx,
		hourGanWx, hourZhiWx,
	)

	// 本地极强推理：推算初级喜用神
	yongshen, jishen := inferNativeYongshen(dayGanWx, wuxing)

	// 藏干
	yearHideGan := bz.GetYearHideGan()
	monthHideGan := bz.GetMonthHideGan()
	dayHideGan := bz.GetDayHideGan()
	hourHideGan := bz.GetTimeHideGan()

	// 纳音
	yearNaYin := bz.GetYearNaYin()
	monthNaYin := bz.GetMonthNaYin()
	dayNaYin := bz.GetDayNaYin()
	hourNaYin := bz.GetTimeNaYin()

	// 大运（1=男命顺排，0=女命逆排）
	genderCode := 1
	if gender == "female" {
		genderCode = 0
	}
	var yun *calendar.Yun
	if genderCode == 1 || genderCode == 0 {
		// 使用 Sect 2 (按3天等于1年，1天等于4个月的传统比例折算)，对主流排盘软件保持严格对齐
		yun = bz.GetYunBySect(genderCode, 2)
	} else {
		yun = bz.GetYunBySect(1, 2) // fallback to male
	}
	startSolar := yun.GetStartSolar()
	startYunStr := fmt.Sprintf("%d年%d月%d日 %02d:%02d交运",
		startSolar.GetYear(), startSolar.GetMonth(), startSolar.GetDay(),
		startSolar.GetHour(), startSolar.GetMinute())

	daYunArr := yun.GetDaYun()
	dayunItems := make([]DayunItem, 0, len(daYunArr))
	prevDayunGanzhi := ""
	for _, dy := range daYunArr {
		gz := dy.GetGanZhi()
		if len([]rune(gz)) < 2 {
			continue
		}
		runes := []rune(gz)
		gan := string(runes[0])
		zhi := string(runes[1])

		lns := dy.GetLiuNian()
		lnItems := make([]LiuNianItem, 0, 10)
		for i, ln := range lns {
			if i >= 10 {
				break
			}
			lngz := ln.GetGanZhi()
			if len([]rune(lngz)) < 2 {
				continue
			}
			lnr := []rune(lngz)
			lnGan := string(lnr[0])
			lnZhi := string(lnr[1])

			isTrans := false
			tm, td := 0, 0
			pd := ""
			if i == 0 {
				isTrans = true
				tm = startSolar.GetMonth()
				td = startSolar.GetDay()
				pd = prevDayunGanzhi
			}

			lnItems = append(lnItems, LiuNianItem{
				Year:         ln.GetYear(),
				Age:          ln.GetAge(),
				GanZhi:       lngz,
				GanShiShen:   GetShiShen(dayGan, lnGan),
				ZhiShiShen:   GetZhiShiShen(dayGan, lnZhi),
				IsTransition: isTrans,
				TransMonth:   tm,
				TransDay:     td,
				PrevDayun:    pd,
			})
		}

		dayunItems = append(dayunItems, DayunItem{
			Index:      dy.GetIndex(),
			Gan:        gan,
			Zhi:        zhi,
			StartAge:   dy.GetStartAge(),
			StartYear:  dy.GetStartYear(),
			EndYear:    dy.GetStartYear() + 9,
			GanShiShen: GetShiShen(dayGan, gan),
			ZhiShiShen: GetZhiShiShen(dayGan, zhi),
			DiShi:      GetDiShi(dayGan, zhi),
			LiuNian:    lnItems,
		})
		prevDayunGanzhi = gz
	}

	// Hash（基于原始输入，保持幂等性）
	hash := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%d-%d-%d-%d-%s", year, month, day, hour, gender))))

	// 调用自研的核心神煞引擎
	shensha := GetPillarsShenSha(yearGan, yearZhi, monthGan, monthZhi, dayGan, dayZhi, hourGan, hourZhi)

	res := &BaziResult{
		YearGan:  yearGan,
		YearZhi:  yearZhi,
		MonthGan: monthGan,
		MonthZhi: monthZhi,
		DayGan:   dayGan,
		DayZhi:   dayZhi,
		HourGan:  hourGan,
		HourZhi:  hourZhi,

		YearGanWuxing:  yearGanWx,
		YearZhiWuxing:  yearZhiWx,
		MonthGanWuxing: monthGanWx,
		MonthZhiWuxing: monthZhiWx,
		DayGanWuxing:   dayGanWx,
		DayZhiWuxing:   dayZhiWx,
		HourGanWuxing:  hourGanWx,
		HourZhiWuxing:  hourZhiWx,

		YearGanShiShen:  bz.GetYearShiShenGan(),
		MonthGanShiShen: bz.GetMonthShiShenGan(),
		DayGanShiShen:   bz.GetDayShiShenGan(),
		HourGanShiShen:  bz.GetTimeShiShenGan(),

		YearZhiShiShen:  listToSlice(bz.GetYearShiShenZhi()),
		MonthZhiShiShen: listToSlice(bz.GetMonthShiShenZhi()),
		DayZhiShiShen:   listToSlice(bz.GetDayShiShenZhi()),
		HourZhiShiShen:  listToSlice(bz.GetTimeShiShenZhi()),

		YearDiShi:  bz.GetYearDiShi(),
		MonthDiShi: bz.GetMonthDiShi(),
		DayDiShi:   bz.GetDayDiShi(),
		HourDiShi:  bz.GetTimeDiShi(),

		YearXingYun:  bz.GetYearDiShi(),
		MonthXingYun: bz.GetMonthDiShi(),
		DayXingYun:   bz.GetDayDiShi(),
		HourXingYun:  bz.GetTimeDiShi(),

		YearXunKong:  bz.GetYearXunKong(),
		MonthXunKong: bz.GetMonthXunKong(),
		DayXunKong:   bz.GetDayXunKong(),
		HourXunKong:  bz.GetTimeXunKong(),

		YearHideGan:  yearHideGan,
		MonthHideGan: monthHideGan,
		DayHideGan:   dayHideGan,
		HourHideGan:  hourHideGan,

		YearShenSha:  shensha[0],
		MonthShenSha: shensha[1],
		DayShenSha:   shensha[2],
		HourShenSha:  shensha[3],

		YearNaYin:  yearNaYin,
		MonthNaYin: monthNaYin,
		DayNaYin:   dayNaYin,
		HourNaYin:  hourNaYin,

		Wuxing: wuxing,

		Yongshen: yongshen,
		Jishen:   jishen,

		StartYunSolar: startYunStr,
		Dayun:         dayunItems,
		Gender:        gender,

		BirthYear:  year,
		BirthMonth: month,
		BirthDay:   day,
		BirthHour:  hour,

		TrueSolarHour:   tsHour,
		TrueSolarMinute: tsMinute,

		ChartHash: hash,
	}

	// 计算调候用神
	res.Tiaohou = calcTiaohou(res)

	return res
}

// wuxingNameToKey 将五行名称映射为统计 key
func wuxingNameToKey(name string) string {
	switch name {
	case "木":
		return "mu"
	case "火":
		return "huo"
	case "土":
		return "tu"
	case "金":
		return "jin"
	case "水":
		return "shui"
	}
	return ""
}

// calcWuxingFromNames 统计八个干支的五行分布
func calcWuxingFromNames(names ...string) WuxingStats {
	counts := map[string]int{"mu": 0, "huo": 0, "tu": 0, "jin": 0, "shui": 0}
	for _, n := range names {
		k := wuxingNameToKey(n)
		if k != "" {
			counts[k]++
		}
	}
	total := 8
	return WuxingStats{
		Mu: counts["mu"], Huo: counts["huo"], Tu: counts["tu"],
		Jin: counts["jin"], Shui: counts["shui"],
		Total:   total,
		MuPct:   float64(counts["mu"]) / float64(total) * 100,
		HuoPct:  float64(counts["huo"]) / float64(total) * 100,
		TuPct:   float64(counts["tu"]) / float64(total) * 100,
		JinPct:  float64(counts["jin"]) / float64(total) * 100,
		ShuiPct: float64(counts["shui"]) / float64(total) * 100,
	}
}

// listToSlice 辅佐函数将 lunar-go 的 list 转化为 string slice
func listToSlice(l *list.List) []string {
	if l == nil {
		return []string{}
	}
	res := make([]string, 0, l.Len())
	for e := l.Front(); e != nil; e = e.Next() {
		res = append(res, e.Value.(string))
	}
	return res
}

// inferNativeYongshen 引擎层内联极简用神统计算法（无须 LLM 的极速版）
// 它仅仅依赖基础常识：命中自身五行与生助自己的五行比例是否越界（>40% 为身强极性判定）
func inferNativeYongshen(dayGanWx string, stats WuxingStats) (yongshen, jishen string) {
	var helpPct float64
	var helpElements, opposeElements string
	switch dayGanWx {
	case "木":
		helpPct = stats.MuPct + stats.ShuiPct
		helpElements = "水木"
		opposeElements = "金土火"
	case "火":
		helpPct = stats.HuoPct + stats.MuPct
		helpElements = "木火"
		opposeElements = "水金土"
	case "土":
		helpPct = stats.TuPct + stats.HuoPct
		helpElements = "火土"
		opposeElements = "木水金"
	case "金":
		helpPct = stats.JinPct + stats.TuPct
		helpElements = "土金"
		opposeElements = "火木水"
	case "水":
		helpPct = stats.ShuiPct + stats.JinPct
		helpElements = "金水"
		opposeElements = "土火木"
	}

	// 如果生助的属性超过 40%，按极简算法算作身强，喜克/泄/耗
	if helpPct >= 40.0 {
		return opposeElements, helpElements
	}
	// 否则身弱，喜生/助
	return helpElements, opposeElements
}


