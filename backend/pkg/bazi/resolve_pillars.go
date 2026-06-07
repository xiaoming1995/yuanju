package bazi

import "github.com/6tail/lunar-go/calendar"

const (
	resolveMinYear = 1900
	resolveMaxYear = 2030
)

const (
	resolveGan = "甲乙丙丁戊己庚辛壬癸"
	resolveZhi = "子丑寅卯辰巳午未申酉戌亥"
)

// Candidate 一个能产生目标四柱的候选公历日期
type Candidate struct {
	Year      int    `json:"year"`
	Month     int    `json:"month"`
	Day       int    `json:"day"`
	Hour      int    `json:"hour"`       // 时辰中点小时，供后续 Calculate 复现时柱
	LunarDate string `json:"lunar_date"` // 如「乙巳年六月初九」
	RefAge    int    `json:"ref_age"`    // referenceYear - Year，供用户按年龄辨识
}

// validGanZhi 判断 gz 是否为 60 甲子之一（阴阳同性配对）。
func validGanZhi(gz string) bool {
	r := []rune(gz)
	if len(r) != 2 {
		return false
	}
	gi := indexRune(resolveGan, r[0])
	zi := indexRune(resolveZhi, r[1])
	if gi < 0 || zi < 0 {
		return false
	}
	return gi%2 == zi%2
}

// indexRune 返回 c 在 s 中的 rune 序号，找不到返回 -1。
func indexRune(s string, c rune) int {
	for i, x := range []rune(s) {
		if x == c {
			return i
		}
	}
	return -1
}

// zhiMidpointHour 返回时支对应时辰的中点小时（子=0、丑=2 … 亥=22）。时支非法返回 -1。
func zhiMidpointHour(hourGZ string) int {
	r := []rune(hourGZ)
	if len(r) != 2 {
		return -1
	}
	zi := indexRune(resolveZhi, r[1])
	if zi < 0 {
		return -1
	}
	return zi * 2
}

// pillarsAt 用与 engine.go Calculate 完全一致的路径，取某公历时刻的四柱。
func pillarsAt(solar *calendar.Solar) (yearGZ, monthGZ, dayGZ, hourGZ string) {
	bz := solar.GetLunar().GetEightChar()
	yearGZ = bz.GetYearGan() + bz.GetYearZhi()
	monthGZ = bz.GetMonthGan() + bz.GetMonthZhi()
	dayGZ = bz.GetDayGan() + bz.GetDayZhi()
	hourGZ = bz.GetTimeGan() + bz.GetTimeZhi()
	return
}

// ResolvePillars 反查能产生目标四柱的公历日期。
// 4 个入参为干支字符串（如 "甲子"）；[minYear,maxYear] 会被夹到 [1900,2030]；
// referenceYear 用于计算候选参考年龄。非法/不自洽的四柱返回空切片。结果按年份升序。
func ResolvePillars(yearGZ, monthGZ, dayGZ, hourGZ string, minYear, maxYear, referenceYear int) []Candidate {
	out := []Candidate{}

	if !validGanZhi(yearGZ) || !validGanZhi(monthGZ) || !validGanZhi(dayGZ) || !validGanZhi(hourGZ) {
		return out
	}
	midHour := zhiMidpointHour(hourGZ)
	if midHour < 0 {
		return out
	}

	if minYear < resolveMinYear {
		minYear = resolveMinYear
	}
	if maxYear > resolveMaxYear {
		maxYear = resolveMaxYear
	}
	if minYear > maxYear {
		return out
	}

	start := calendar.NewSolar(minYear, 1, 1, midHour, 0, 0)
	end := calendar.NewSolar(maxYear, 12, 31, midHour, 0, 0)
	endJD := end.GetJulianDay()

	firstOffset := -1
	for i := 0; i < 60; i++ {
		s := start.NextDay(i)
		if _, _, d, _ := pillarsAt(s); d == dayGZ {
			firstOffset = i
			break
		}
	}
	if firstOffset < 0 {
		return out
	}

	for k := 0; ; k++ {
		s := start.NextDay(firstOffset + 60*k)
		if s.GetJulianDay() > endJD {
			break
		}
		lunar := s.GetLunar()
		bz := lunar.GetEightChar()
		y := bz.GetYearGan() + bz.GetYearZhi()
		mo := bz.GetMonthGan() + bz.GetMonthZhi()
		d := bz.GetDayGan() + bz.GetDayZhi()
		h := bz.GetTimeGan() + bz.GetTimeZhi()
		if y == yearGZ && mo == monthGZ && d == dayGZ && h == hourGZ {
			out = append(out, Candidate{
				Year:      s.GetYear(),
				Month:     s.GetMonth(),
				Day:       s.GetDay(),
				Hour:      midHour,
				LunarDate: lunar.GetYearInGanZhi() + "年" + lunar.GetMonthInChinese() + "月" + lunar.GetDayInChinese(),
				RefAge:    referenceYear - s.GetYear(),
			})
		}
	}
	return out
}
