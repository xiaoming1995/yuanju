package bazi

import (
	"fmt"
	"time"

	"github.com/6tail/lunar-go/calendar"
)

// LiuYueItem 流月数据结构
type LiuYueItem struct {
	Index      int    `json:"index"`
	MonthName  string `json:"month_name"`
	GanZhi     string `json:"gan_zhi"`
	GanShiShen string `json:"gan_shishen"`
	ZhiShiShen string `json:"zhi_shishen"`
	JieQiName  string `json:"jie_qi_name"`
	StartDate  string `json:"start_date"` // YYYY-MM-DD
	EndDate    string `json:"end_date"`   // YYYY-MM-DD（含）
}

// liuYueJieQi 12个月令对应节气（寅月=立春，依序）
var liuYueJieQi = []string{
	"立春", "惊蛰", "清明", "立夏", "芒种", "小暑",
	"立秋", "白露", "寒露", "立冬", "大雪", "小寒",
}

// liuYueMonthNames 流月地支月名
var liuYueMonthNames = []string{
	"寅月", "卯月", "辰月", "巳月", "午月", "未月",
	"申月", "酉月", "戌月", "亥月", "子月", "丑月",
}

// wuHuDunOffset 五虎遁：年干→寅月天干在 Gan[] 中的偏移量（0-indexed）
// 规则：甲己丙起，乙庚戊起，丙辛庚起，丁壬壬起，戊癸甲起
var wuHuDunOffset = map[string]int{
	"甲": 2, "己": 2, // 丙寅起
	"乙": 4, "庚": 4, // 戊寅起
	"丙": 6, "辛": 6, // 庚寅起
	"丁": 8, "壬": 8, // 壬寅起
	"戊": 0, "癸": 0, // 甲寅起
}

// validGanSet 合法天干集合（用于参数校验）
var validGanSet = map[string]bool{
	"甲": true, "乙": true, "丙": true, "丁": true, "戊": true,
	"己": true, "庚": true, "辛": true, "壬": true, "癸": true,
}

// CalcLiuYue 计算指定流年公历年的 12 个流月数据
//
// year:   流年公历年份，范围 1900-2200
// dayGan: 命主日主天干（单个汉字），用于计算十神
//
// 返回：
//   - []LiuYueItem：12 个流月（寅月→丑月）
//   - int：当前流月 index（-1 表示今天不在该命理年范围内）
//   - error
func CalcLiuYue(year int, dayGan string) ([]LiuYueItem, int, error) {
	if year < 1900 || year > 2200 {
		return nil, -1, fmt.Errorf("年份超出范围（1900-2200）：%d", year)
	}
	if !validGanSet[dayGan] {
		return nil, -1, fmt.Errorf("无效天干：%s", dayGan)
	}

	// 取当年 6月1日对应的农历年（必然在该命理年「立春X年→立春X+1年」范围内）
	solar := calendar.NewSolar(year, 6, 1, 0, 0, 0)
	jieQiTable := solar.GetLunar().GetJieQiTable()

	// 丑月（index=11）结束日期跨年，需要次年立春日期
	nextSolar := calendar.NewSolar(year+1, 6, 1, 0, 0, 0)
	nextJieQiTable := nextSolar.GetLunar().GetJieQiTable()

	// 计算流年天干（用于五虎遁月干推算）
	// 甲子=1984: (1984-4)%10=0→甲; ((year-4)%10+10)%10 防负数
	yearGanIndex := ((year - 4) % 10 + 10) % 10
	yearGan := Gan[yearGanIndex]
	offset := wuHuDunOffset[yearGan]

	// 关键：小寒（丑月）在 lunar-go 节气表中按公历年存储，
	// jieQiTable["小寒"] 返回当年 1 月的小寒（如 2026-01-05），
	// 但丑月实际起始在命理年末（次年1月），必须从 nextJieQiTable 取。
	xiaoHanSolar := nextJieQiTable["小寒"]

	// 预先生成 12 个月的起始 Solar（用于统一计算 end_date 和当前月 index）
	monthStartSolars := make([]*calendar.Solar, 12)
	for i := 0; i < 11; i++ {
		monthStartSolars[i] = jieQiTable[liuYueJieQi[i]]
	}
	monthStartSolars[11] = xiaoHanSolar // 丑月从次年小寒开始

	items := make([]LiuYueItem, 12)
	for i := 0; i < 12; i++ {
		// 月干：五虎遁从寅月开始，每月顺推1位天干
		monthGan := Gan[(i+offset)%10]
		// 月支：寅=Zhi[2]，每月顺推1位地支
		monthZhi := Zhi[(i+2)%12]
		ganZhi := monthGan + monthZhi

		// 节气起始日期
		startSolar := monthStartSolars[i]
		startDate := fmt.Sprintf("%04d-%02d-%02d",
			startSolar.GetYear(), startSolar.GetMonth(), startSolar.GetDay())

		// 节气结束日期（下一个月起始前一天）
		var endDate string
		if i < 11 {
			// 下一个月从 monthStartSolars[i+1] 起始
			endSolar := monthStartSolars[i+1].NextDay(-1)
			endDate = fmt.Sprintf("%04d-%02d-%02d",
				endSolar.GetYear(), endSolar.GetMonth(), endSolar.GetDay())
		} else {
			// 丑月：结束于次年立春前一天
			nextLiChunSolar := nextJieQiTable["立春"]
			endSolar := nextLiChunSolar.NextDay(-1)
			endDate = fmt.Sprintf("%04d-%02d-%02d",
				endSolar.GetYear(), endSolar.GetMonth(), endSolar.GetDay())
		}

		items[i] = LiuYueItem{
			Index:      i,
			MonthName:  liuYueMonthNames[i],
			GanZhi:     ganZhi,
			GanShiShen: GetShiShen(dayGan, monthGan),
			ZhiShiShen: GetZhiShiShen(dayGan, monthZhi),
			JieQiName:  liuYueJieQi[i],
			StartDate:  startDate,
			EndDate:    endDate,
		}
	}

	// ─── 计算当前流月 index ───────────────────────────────────────────
	// 判断今天是否在该命理年范围内（立春 year ≤ 今天 < 立春 year+1）
	liChun := jieQiTable["立春"]
	nextLiChunForCheck := nextJieQiTable["立春"]
	today := time.Now()
	todayDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.Local)
	liChunDate := time.Date(liChun.GetYear(), time.Month(liChun.GetMonth()), liChun.GetDay(), 0, 0, 0, 0, time.Local)
	nextLiChunDate := time.Date(nextLiChunForCheck.GetYear(), time.Month(nextLiChunForCheck.GetMonth()), nextLiChunForCheck.GetDay(), 0, 0, 0, 0, time.Local)

	if todayDate.Before(liChunDate) || !todayDate.Before(nextLiChunDate) {
		// 今天不在该命理年范围内，返回 -1（前端不高亮）
		return items, -1, nil
	}

	// 用修正后的 monthStartSolars（含2027小寒）从丑月往前找当前月
	currentIndex := 0
	for i := 11; i >= 0; i-- {
		s := monthStartSolars[i]
		sd := time.Date(s.GetYear(), time.Month(s.GetMonth()), s.GetDay(), 0, 0, 0, 0, time.Local)
		if !todayDate.Before(sd) {
			currentIndex = i
			break
		}
	}

	return items, currentIndex, nil
}
