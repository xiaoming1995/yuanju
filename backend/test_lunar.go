package main

import (
	"fmt"
	"github.com/6tail/lunar-go/calendar"
)

func main() {
	// 2023年闰2月的情况（农历2023年没有闰2月，2023年闰2月是公历2023年）
	// 让我们用一个有闰月的年份，例如 2020 年闰 4 月
	lunar := calendar.NewLunar(2020, -4, 15, 12, 0, 0)
	solar := lunar.GetSolar()
	fmt.Printf("Lunar -> Solar: %d-%d-%d\n", solar.GetYear(), solar.GetMonth(), solar.GetDay())

	// 再看看是否可以通过 Solar 拿回 Lunar 的闰月信息
	l := solar.GetLunar()
	fmt.Printf("Solar -> Lunar: month=%d, math.Abs=%d\n", l.GetMonth(), l.GetMonth()) // 结果通常还是返回正数，但可通过另外的方法判断
}
