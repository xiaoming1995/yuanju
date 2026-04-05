package bazi

import (
	"strings"
	"testing"
	"time"
)

// TestCalcLiuYue_2026_甲日主 验证 2026 年甲日主的 12 个流月干支
func TestCalcLiuYue_2026_甲日主(t *testing.T) {
	items, currentIndex, err := CalcLiuYue(2026, "甲")
	if err != nil {
		t.Fatalf("CalcLiuYue 返回错误：%v", err)
	}
	if len(items) != 12 {
		t.Fatalf("期望 12 个流月，实际 %d 个", len(items))
	}

	// 2026 年天干为丙（丙午年），五虎遁：丙/辛年寅月天干从庚开始
	// 寅月=庚寅, 卯月=辛卯, 辰月=壬辰, 巳月=癸巳, 午月=甲午, 未月=乙未
	// 申月=丙申, 酉月=丁酉, 戌月=戊戌, 亥月=己亥, 子月=庚子, 丑月=辛丑
	expected := []string{"庚寅", "辛卯", "壬辰", "癸巳", "甲午", "乙未", "丙申", "丁酉", "戊戌", "己亥", "庚子", "辛丑"}
	for i, want := range expected {
		if items[i].GanZhi != want {
			t.Errorf("index=%d 期望干支 %s，实际 %s", i, want, items[i].GanZhi)
		}
	}

	// 节气名称验证
	if items[0].JieQiName != "立春" {
		t.Errorf("寅月(index=0) 节气应为立春，实际 %s", items[0].JieQiName)
	}
	if items[11].JieQiName != "小寒" {
		t.Errorf("丑月(index=11) 节气应为小寒，实际 %s", items[11].JieQiName)
	}

	// 月名验证
	if items[0].MonthName != "寅月" {
		t.Errorf("index=0 月名应为寅月，实际 %s", items[0].MonthName)
	}

	// start_date 格式验证
	for _, item := range items {
		if len(item.StartDate) != 10 || item.StartDate[4] != '-' {
			t.Errorf("index=%d StartDate 格式异常：%s", item.Index, item.StartDate)
		}
		if len(item.EndDate) != 10 || item.EndDate[4] != '-' {
			t.Errorf("index=%d EndDate 格式异常：%s", item.Index, item.EndDate)
		}
	}

	// 十神非空验证（甲日主=比肩，月支字面为辰，地支十神应为正财之类）
	for _, item := range items {
		if item.GanShiShen == "" {
			t.Errorf("index=%d GanShiShen 不应为空", item.Index)
		}
		if item.ZhiShiShen == "" {
			t.Errorf("index=%d ZhiShiShen 不应为空", item.Index)
		}
	}

	t.Logf("current_month_index=%d", currentIndex)
	for _, item := range items {
		t.Logf("[%d] %s %s %s %s-%s 天干十神:%s 地支十神:%s",
			item.Index, item.JieQiName, item.MonthName, item.GanZhi,
			item.StartDate, item.EndDate, item.GanShiShen, item.ZhiShiShen)
	}
}

// TestCalcLiuYue_CurrentMonthIndex 验证当前流月 index 与今天节气是否匹配
func TestCalcLiuYue_CurrentMonthIndex(t *testing.T) {
	currentYear := time.Now().Year()
	items, currentIndex, err := CalcLiuYue(currentYear, "甲")
	if err != nil {
		t.Fatalf("CalcLiuYue 返回错误：%v", err)
	}

	// 正常情况下，当前年份查询 currentIndex 应在 0-11 范围内
	// （极端情况：在 1月立春前查询当年，此时属上一年丑月，返回 -1 亦合法）
	if currentIndex < -1 || currentIndex > 11 {
		t.Errorf("currentIndex 超出范围：%d", currentIndex)
	}

	if currentIndex >= 0 {
		item := items[currentIndex]
		t.Logf("今天(%s)所处流月：[%d] %s %s %s (%s~%s)",
			time.Now().Format("2006-01-02"),
			currentIndex, item.JieQiName, item.MonthName, item.GanZhi,
			item.StartDate, item.EndDate)

		// 验证今天在该流月的起止区间内
		today := time.Now().Format("2006-01-02")
		if today < item.StartDate {
			t.Errorf("今天(%s)在流月起始日(%s)之前，index 计算有误", today, item.StartDate)
		}
		if today > item.EndDate {
			t.Errorf("今天(%s)超过流月结束日(%s)，index 计算有误", today, item.EndDate)
		}
	} else {
		t.Logf("当前年份=%d，今天不在该命理年范围内（立春前），currentIndex=-1（合法）", currentYear)
	}
}

// TestCalcLiuYue_ChouchueCrossYear 验证丑月结束日期跨年
func TestCalcLiuYue_ChouchueCrossYear(t *testing.T) {
	for _, year := range []int{2024, 2025, 2026} {
		items, _, err := CalcLiuYue(year, "甲")
		if err != nil {
			t.Fatalf("年份 %d CalcLiuYue 错误：%v", year, err)
		}
		chou := items[11] // 丑月 index=11
		// 丑月从小寒开始（1月），结束于次年立春前一天（也是1-2月）
		// 所以 EndDate 年份应等于 year+1
		if !strings.HasPrefix(chou.EndDate, string(rune('0'+((year+1)/1000%10)))+
			string(rune('0'+((year+1)/100%10)))+
			string(rune('0'+((year+1)/10%10)))+
			string(rune('0'+(year+1)%10))) {
			t.Errorf("年份 %d 丑月 EndDate=%s，期望年份为 %d", year, chou.EndDate, year+1)
		}
		t.Logf("年份 %d 丑月：%s (%s ~ %s)", year, chou.GanZhi, chou.StartDate, chou.EndDate)
	}
}

// TestCalcLiuYue_InvalidInput 验证非法输入
func TestCalcLiuYue_InvalidInput(t *testing.T) {
	_, _, err := CalcLiuYue(1800, "甲")
	if err == nil {
		t.Error("年份 1800 应返回错误")
	}

	_, _, err = CalcLiuYue(2026, "X")
	if err == nil {
		t.Error("非法天干 'X' 应返回错误")
	}

	_, _, err = CalcLiuYue(2026, "")
	if err == nil {
		t.Error("空天干 应返回错误")
	}
}
