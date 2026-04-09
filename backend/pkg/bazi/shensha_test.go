package bazi

import (
	"fmt"
	"testing"
)

func TestShenShaOutputAll(t *testing.T) {
	cases := []struct {
		name string
		y, m, d, h int
	}{
		{"2000年2月29日 午时", 2000, 2, 29, 11},
		{"1995年10月12日 午时", 1995, 10, 12, 11},
		{"1996年2月8日 戌时", 1996, 2, 8, 20},
	}
	for _, c := range cases {
		result := Calculate(c.y, c.m, c.d, c.h, "male", false, 120.0, "solar", false)
		fmt.Printf("=== %s ===\n", c.name)
		fmt.Printf("  四柱: %s%s %s%s %s%s %s%s\n",
			result.YearGan, result.YearZhi,
			result.MonthGan, result.MonthZhi,
			result.DayGan, result.DayZhi,
			result.HourGan, result.HourZhi,
		)
		labels := []string{"年柱", "月柱", "日柱", "时柱"}
		pillars := []string{
			result.YearGan + result.YearZhi,
			result.MonthGan + result.MonthZhi,
			result.DayGan + result.DayZhi,
			result.HourGan + result.HourZhi,
		}
		shenshArrays := [][]string{result.YearShenSha, result.MonthShenSha, result.DayShenSha, result.HourShenSha}
		for i, ss := range shenshArrays {
			fmt.Printf("  %s(%s): %v\n", labels[i], pillars[i], ss)
		}
		fmt.Println()
	}
}
