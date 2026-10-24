// Generate the fixed chart shown in the mini program preview from the same engine as the web result page.
// From backend/: go run ../miniapp/tools/generate_demo_chart.go ../miniapp/data/demo_chart.js
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"yuanju/pkg/bazi"
)

type exampleInput struct {
	DateTime string `json:"dateTime"`
	Gender   string `json:"gender"`
}

type examplePillar struct {
	Label      string   `json:"label"`
	Gan        string   `json:"gan"`
	Zhi        string   `json:"zhi"`
	GanWuxing  string   `json:"ganWuxing"`
	ZhiWuxing  string   `json:"zhiWuxing"`
	GanTone    string   `json:"ganTone"`
	ZhiTone    string   `json:"zhiTone"`
	GanShiShen string   `json:"ganShiShen"`
	HideGan    []string `json:"hideGan"`
	ZhiShiShen []string `json:"zhiShiShen"`
	XingYun    string   `json:"xingYun"`
	DiShi      string   `json:"diShi"`
	XunKong    string   `json:"xunKong"`
	NaYin      string   `json:"naYin"`
	ShenSha    []string `json:"shenSha"`
}

type wuxingRow struct {
	Label   string  `json:"label"`
	Tone    string  `json:"tone"`
	Count   int     `json:"count"`
	Percent float64 `json:"percent"`
}

type examplePersonalDay struct {
	DayBranch    string `json:"dayBranch"`
	FuyiYongshen string `json:"fuyiYongshen"`
	FuyiJishen   string `json:"fuyiJishen"`
}

type exampleChart struct {
	Input        exampleInput            `json:"input"`
	Pillars      []examplePillar         `json:"pillars"`
	WuxingRows   []wuxingRow             `json:"wuxingRows"`
	WuxingTotal  int                     `json:"wuxingTotal"`
	MingGe       string                  `json:"mingGe"`
	MingGeDesc   string                  `json:"mingGeDesc"`
	StemGuidance *bazi.NatalStemGuidance `json:"stemGuidance"`
	PersonalDay  examplePersonalDay      `json:"personalDay"`
}

func tone(element string) string {
	switch element {
	case "木":
		return "wood"
	case "火":
		return "fire"
	case "土":
		return "earth"
	case "金":
		return "metal"
	case "水":
		return "water"
	default:
		return "neutral"
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: generate_demo_chart.go <output-js>")
		os.Exit(2)
	}

	result := bazi.Calculate(1990, 1, 1, 1, "male", false, 0, "solar", false)
	if result.NatalAssessment == nil {
		panic("sample chart is missing natal assessment")
	}
	chart := exampleChart{
		Input:        exampleInput{DateTime: "1990-01-01 01:00", Gender: "male"},
		MingGe:       result.MingGe,
		MingGeDesc:   result.MingGeDesc,
		StemGuidance: result.NatalAssessment.StemGuidance,
		PersonalDay: examplePersonalDay{
			DayBranch:    result.DayZhi,
			FuyiYongshen: result.NatalAssessment.Fuyi.Yongshen,
			FuyiJishen:   result.NatalAssessment.Fuyi.Jishen,
		},
		Pillars: []examplePillar{
			{"年柱", result.YearGan, result.YearZhi, result.YearGanWuxing, result.YearZhiWuxing, tone(result.YearGanWuxing), tone(result.YearZhiWuxing), result.YearGanShiShen, result.YearHideGan, result.YearZhiShiShen, result.YearXingYun, result.YearDiShi, result.YearXunKong, result.YearNaYin, result.YearShenSha},
			{"月柱", result.MonthGan, result.MonthZhi, result.MonthGanWuxing, result.MonthZhiWuxing, tone(result.MonthGanWuxing), tone(result.MonthZhiWuxing), result.MonthGanShiShen, result.MonthHideGan, result.MonthZhiShiShen, result.MonthXingYun, result.MonthDiShi, result.MonthXunKong, result.MonthNaYin, result.MonthShenSha},
			{"日柱", result.DayGan, result.DayZhi, result.DayGanWuxing, result.DayZhiWuxing, tone(result.DayGanWuxing), tone(result.DayZhiWuxing), result.DayGanShiShen, result.DayHideGan, result.DayZhiShiShen, result.DayXingYun, result.DayDiShi, result.DayXunKong, result.DayNaYin, result.DayShenSha},
			{"时柱", result.HourGan, result.HourZhi, result.HourGanWuxing, result.HourZhiWuxing, tone(result.HourGanWuxing), tone(result.HourZhiWuxing), result.HourGanShiShen, result.HourHideGan, result.HourZhiShiShen, result.HourXingYun, result.HourDiShi, result.HourXunKong, result.HourNaYin, result.HourShenSha},
		},
		WuxingTotal: result.Wuxing.Total,
	}
	for _, item := range []struct {
		label string
		count int
	}{
		{"木", result.Wuxing.Mu},
		{"火", result.Wuxing.Huo},
		{"土", result.Wuxing.Tu},
		{"金", result.Wuxing.Jin},
		{"水", result.Wuxing.Shui},
	} {
		percent := 0.0
		if result.Wuxing.Total > 0 {
			percent = float64(item.count) * 100 / float64(result.Wuxing.Total)
		}
		chart.WuxingRows = append(chart.WuxingRows, wuxingRow{item.label, tone(item.label), item.count, percent})
	}

	data, err := json.MarshalIndent(chart, "", "  ")
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(os.Args[1], append([]byte("module.exports = "), data...), 0644); err != nil {
		panic(err)
	}
}
