package bazi

import (
	"fmt"
	"testing"
)

func TestShenShaOutputTestCase2(t *testing.T) {
	// 1995年10月12日 午时（11点）
	result := Calculate(1995, 10, 12, 11, "male", false, 120.0, "solar", false)
	fmt.Println("=== 1995年10月12日 午时 ===")
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
}

func TestShenShaOutputTestCase1(t *testing.T) {
	// 2000年2月29日 午时（11点）庚辰 戊寅 丁巳 丙午
	result := Calculate(2000, 2, 29, 11, "male", false, 120.0, "solar", false)
	fmt.Println("=== 2000年2月29日 午时 ===")
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
}
