// Generate static date labels for the UI preview; this is not a runtime API.
// From backend/: go run ../miniapp/tools/generate_demo_calendar.go ../miniapp/data/calendar.js
package main

import (
	"container/list"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/6tail/lunar-go/calendar"
)

type dayLabel struct {
	Lunar       string   `json:"lunar"`
	LunarFull   string   `json:"lunarFull"`
	JieQi       string   `json:"jieQi,omitempty"`
	DayGanZhi   string   `json:"dayGanZhi"`
	YearGanZhi  string   `json:"yearGanZhi"`
	MonthGanZhi string   `json:"monthGanZhi"`
	JianChu     string   `json:"jianChu"`
	ActivityYi  []string `json:"activityYi"`
	ActivityJi  []string `json:"activityJi"`
}

var activityLabels = map[string]bool{
	"入宅": true, "移徙": true, "嫁娶": true,
	"开市": true, "出行": true, "动土": true,
}

func selectedLabels(items *list.List, includeGeneralJi bool) []string {
	labels := []string{}
	for item := items.Front(); item != nil; item = item.Next() {
		label, ok := item.Value.(string)
		if ok && (activityLabels[label] || includeGeneralJi && label == "诸事不宜") {
			labels = append(labels, label)
		}
	}
	return labels
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: generate_demo_calendar.go <output-js>")
		os.Exit(2)
	}
	days := make(map[string]dayLabel)
	for year := 2025; year <= 2027; year++ {
		for month := 1; month <= 12; month++ {
			for day := 1; day <= time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.UTC).Day(); day++ {
				solar := calendar.NewSolar(year, month, day, 12, 0, 0)
				lunar := solar.GetLunar()
				monthText := lunar.GetMonthInChinese() + "月"
				dayText := lunar.GetDayInChinese()
				label := dayText
				if dayText == "初一" {
					label = monthText
				}
				key := fmt.Sprintf("%04d-%02d-%02d", year, month, day)
				days[key] = dayLabel{
					Lunar:       label,
					LunarFull:   monthText + dayText,
					JieQi:       lunar.GetJieQi(),
					DayGanZhi:   lunar.GetDayInGanZhi(),
					YearGanZhi:  lunar.GetYearInGanZhiByLiChun(),
					MonthGanZhi: lunar.GetMonthInGanZhi(),
					JianChu:     lunar.GetZhiXing(),
					ActivityYi:  selectedLabels(lunar.GetDayYi(), false),
					ActivityJi:  selectedLabels(lunar.GetDayJi(), true),
				}
			}
		}
	}
	data, err := json.Marshal(days)
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(os.Args[1], append([]byte("module.exports = "), data...), 0644); err != nil {
		panic(err)
	}
}
