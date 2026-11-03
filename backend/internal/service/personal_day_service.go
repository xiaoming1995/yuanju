package service

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"yuanju/pkg/bazi"

	"github.com/6tail/lunar-go/calendar"
)

var ErrInvalidPersonalDayDate = errors.New("日期无效或超出支持范围（1900—2100年）")

type PersonalDayActivity struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Tip    string `json:"tip"`
	Detail string `json:"detail"`
}

type PersonalDayLuck struct {
	Available bool   `json:"available"`
	Level     string `json:"level,omitempty"`
	Element   string `json:"element,omitempty"`
	Direction string `json:"direction,omitempty"`
	Color     string `json:"color,omitempty"`
	Reason    string `json:"reason"`
}

type PersonalDayResult struct {
	Date             string                `json:"date"`
	LunarFull        string                `json:"lunar_full"`
	YearGanZhi       string                `json:"year_gan_zhi"`
	MonthGanZhi      string                `json:"month_gan_zhi"`
	DayGanZhi        string                `json:"day_gan_zhi"`
	DayGan           string                `json:"day_gan"`
	DayZhi           string                `json:"day_zhi"`
	CalendarYi       []string              `json:"calendar_yi"`
	CalendarJi       []string              `json:"calendar_ji"`
	StemElement      string                `json:"stem_element,omitempty"`
	BranchElement    string                `json:"branch_element,omitempty"`
	StemDirection    string                `json:"stem_direction,omitempty"`
	BranchDirection  string                `json:"branch_direction,omitempty"`
	Status           string                `json:"status"`
	Label            string                `json:"label"`
	Summary          string                `json:"summary"`
	Evidence         []string              `json:"evidence"`
	Activities       []PersonalDayActivity `json:"activities"`
	ActivityGuidance []PersonalDayGuidance `json:"activity_guidance"`
	Luck             PersonalDayLuck       `json:"luck"`
	RuleVersion      string                `json:"rule_version"`
	Favorable        string                `json:"favorable"`
	Adverse          string                `json:"adverse"`
	ChartID          string                `json:"chart_id"`
	ChartName        string                `json:"chart_name"`
	ChartDay         string                `json:"chart_day"`
}

var dayStemElement = map[string]string{"甲": "木", "乙": "木", "丙": "火", "丁": "火", "戊": "土", "己": "土", "庚": "金", "辛": "金", "壬": "水", "癸": "水"}
var dayBranchElement = map[string]string{"子": "水", "丑": "土", "寅": "木", "卯": "木", "辰": "土", "巳": "火", "午": "火", "未": "土", "申": "金", "酉": "金", "戌": "土", "亥": "水"}
var dayBranchOpposite = map[string]string{"子": "午", "丑": "未", "寅": "申", "卯": "酉", "辰": "戌", "巳": "亥", "午": "子", "未": "丑", "申": "寅", "酉": "卯", "戌": "辰", "亥": "巳"}
var elementLuck = map[string]struct{ direction, color string }{
	"木": {"东方", "绿色系"}, "火": {"南方", "红色系"}, "土": {"中央", "黄色系"},
	"金": {"西方", "白色系"}, "水": {"北方", "蓝黑色系"},
}

func personalDayLuck(stemElement, branchElement, stemDirection, branchDirection, favorable string, climateConflict, clash bool) PersonalDayLuck {
	unavailable := func(reason string) PersonalDayLuck { return PersonalDayLuck{Reason: reason} }
	if stemElement == "" || branchElement == "" || stemDirection == "" || branchDirection == "" || favorable == "" {
		return unavailable("命盘喜忌依据不足，暂无明确参考。")
	}
	baseline := func(context string) PersonalDayLuck {
		directions, colors := []string{}, []string{}
		seen := map[rune]bool{}
		for _, element := range favorable {
			if seen[element] {
				continue
			}
			mapping, ok := elementLuck[string(element)]
			if !ok {
				return unavailable("命盘喜用五行资料不完整，暂无明确参考。")
			}
			seen[element] = true
			directions = append(directions, mapping.direction)
			colors = append(colors, mapping.color)
		}
		return PersonalDayLuck{Available: true, Level: "baseline", Element: favorable,
			Direction: strings.Join(directions, "、"), Color: strings.Join(colors, "、"),
			Reason: fmt.Sprintf("%s；以下仅是本人命盘喜用%s的传统五行对应，不代表今日适宜。", context, favorable)}
	}
	element := ""
	basis := ""
	if stemDirection == "favorable" {
		element = stemElement
		basis = "日干"
	}
	if branchDirection == "favorable" {
		if element != "" && element != branchElement {
			return baseline("今日干支涉及两个不同的喜用五行，无法选出唯一的每日参考")
		}
		element = branchElement
		if basis == "" {
			basis = "日支主气"
		} else {
			basis = "日干与日支主气"
		}
	}
	if element == "" {
		return baseline("今日干支未落入明确喜用五行")
	}
	mapping, ok := elementLuck[element]
	if !ok {
		return unavailable("五行资料不完整，暂无明确参考。")
	}
	result := PersonalDayLuck{Available: true, Level: "aligned", Element: element, Direction: mapping.direction, Color: mapping.color,
		Reason: fmt.Sprintf("今日%s的%s属本人扶抑喜用；方位与色彩按传统五行对应，仅供参考。", basis, element)}
	if climateConflict || clash || stemDirection == "adverse" || branchDirection == "adverse" {
		conflicts := []string{}
		if stemDirection == "adverse" || branchDirection == "adverse" {
			conflicts = append(conflicts, "另一干支五行属忌神")
		}
		if climateConflict {
			conflicts = append(conflicts, "扶抑与调候方向不同")
		}
		if clash {
			conflicts = append(conflicts, "日支与本人日支相冲")
		}
		result.Level = "qualified"
		result.Reason = fmt.Sprintf("今日%s的%s属本人扶抑喜用，可参考其传统五行对应；但%s，仅为局部参考，不代表整日适宜。", basis, element, strings.Join(conflicts, "、"))
	}
	return result
}

func personalDayDirection(element, favorable, adverse string) string {
	if strings.Contains(favorable, element) {
		return "favorable"
	}
	if strings.Contains(adverse, element) {
		return "adverse"
	}
	return "neutral"
}

func validElementSet(s string) bool {
	if s == "" {
		return false
	}
	for _, ch := range s {
		if !strings.ContainsRune("木火土金水", ch) {
			return false
		}
	}
	return true
}

// BuildPersonalDay uses civil dates in Asia/Shanghai and noon, matching the
// static miniapp calendar generator. It is a limited interpretive rule set.
func BuildPersonalDay(date string, chart *bazi.BaziResult) (*PersonalDayResult, error) {
	zone, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return nil, err
	}
	parsed, err := time.ParseInLocation("2006-01-02", date, zone)
	if err != nil || parsed.Year() < 1900 || parsed.Year() > 2100 || parsed.Format("2006-01-02") != date {
		return nil, ErrInvalidPersonalDayDate
	}
	solar := calendar.NewSolar(parsed.Year(), int(parsed.Month()), parsed.Day(), 12, 0, 0)
	lunar := solar.GetLunar()
	ganZhi := lunar.GetDayInGanZhi()
	runes := []rune(ganZhi)
	if len(runes) != 2 {
		return nil, errors.New("无法计算当日干支")
	}
	gan, zhi := string(runes[0]), string(runes[1])
	output := &PersonalDayResult{
		Date: date, LunarFull: lunar.GetMonthInChinese() + "月" + lunar.GetDayInChinese(),
		YearGanZhi: lunar.GetYearInGanZhiByLiChun(), MonthGanZhi: lunar.GetMonthInGanZhi(),
		DayGanZhi: ganZhi, DayGan: gan, DayZhi: zhi, Status: "insufficient", Label: "依据不足",
		Summary:  "当前命盘缺少有效的扶抑喜忌，暂不提供个人事项判断。",
		Evidence: []string{}, Activities: []PersonalDayActivity{}, RuleVersion: "personal-day-v4",
		Luck: PersonalDayLuck{Reason: "命盘喜忌依据不足，暂无明确参考。"},
	}
	yi, ji := lunar.GetDayYi(), lunar.GetDayJi()
	output.CalendarYi, output.CalendarJi = dayListLabels(yi), dayListLabels(ji)
	activity := personalDayActivityContext{dayGan: gan, dayZhi: zhi, yi: yi, ji: ji}
	output.ActivityGuidance = buildPersonalDayGuidance(activity)
	if chart == nil || chart.NatalAssessment == nil || dayStemElement[chart.DayGan] == "" || dayBranchElement[chart.DayZhi] == "" {
		return output, nil
	}
	fuyi := chart.NatalAssessment.Fuyi
	if !validElementSet(fuyi.Yongshen) || !validElementSet(fuyi.Jishen) {
		return output, nil
	}
	for _, e := range fuyi.Yongshen {
		if strings.ContainsRune(fuyi.Jishen, e) {
			return output, nil
		}
	}
	output.Favorable, output.Adverse, output.ChartDay = fuyi.Yongshen, fuyi.Jishen, chart.DayGan+chart.DayZhi
	ganElement, zhiElement := dayStemElement[gan], dayBranchElement[zhi]
	gd, zd := personalDayDirection(ganElement, fuyi.Yongshen, fuyi.Jishen), personalDayDirection(zhiElement, fuyi.Yongshen, fuyi.Jishen)
	output.StemElement, output.BranchElement = ganElement, zhiElement
	output.StemDirection, output.BranchDirection = gd, zd
	ganGod, zhiGod := bazi.GetShiShen(chart.DayGan, gan), bazi.GetZhiShiShen(chart.DayGan, zhi)
	output.Evidence = append(output.Evidence,
		fmt.Sprintf("日干「%s」属%s，对%s日主为%s；扶抑方向：%s。", gan, ganElement, chart.DayGan, ganGod, directionLabel(gd)),
		fmt.Sprintf("日支「%s」主气属%s，对%s日主为%s；扶抑方向：%s（仅按主气）。", zhi, zhiElement, chart.DayGan, zhiGod, directionLabel(zd)))
	clash := dayBranchOpposite[zhi] == chart.DayZhi
	if clash {
		output.Evidence = append(output.Evidence, fmt.Sprintf("当日日支「%s」与命盘日支「%s」相冲；重要安排可另行核对。", zhi, chart.DayZhi))
	}
	climateConflict := false
	conflictElements := map[string]bool{}
	required := chart.NatalAssessment.Tiaohou.Thermal.RequiredElements
	for _, element := range []string{ganElement, zhiElement} {
		if strings.Contains(fuyi.Jishen, element) && strings.Contains(required, element) {
			climateConflict = true
			conflictElements[element] = true
		}
	}
	if climateConflict {
		output.Evidence = append(output.Evidence, "扶抑与调候对此五行的方向不同，需结合完整命盘阅读。")
	}
	output.Luck = personalDayLuck(ganElement, zhiElement, gd, zd, fuyi.Yongshen, climateConflict, clash)
	activity.valid = true
	activity.ganElement, activity.zhiElement = ganElement, zhiElement
	activity.ganDirection, activity.zhiDirection = gd, zd
	activity.ganGod, activity.zhiGod = ganGod, zhiGod
	activity.clash = clash
	activity.conflictElements = conflictElements
	output.ActivityGuidance = buildPersonalDayGuidance(activity)
	switch {
	case climateConflict:
		output.Status, output.Label, output.Summary = "mixed", "关系复杂", "扶抑与调候方向不同，需看完整命盘依据。"
	case (gd == "favorable" && zd == "adverse") || (gd == "adverse" && zd == "favorable"):
		output.Status, output.Label, output.Summary = "mixed", "关系复杂", "日干与日支主气方向相反，需分别看依据。"
	case clash:
		output.Status, output.Label, output.Summary = "caution", "需留意", "当日日支与命盘日支相冲，重要安排请核对现实条件。"
	case gd == "adverse" && zd == "adverse":
		output.Status, output.Label, output.Summary = "caution", "需留意", "日干与日支主气均落入扶抑忌神，安排事项请核对现实条件。"
	case gd == "favorable" && zd == "favorable":
		output.Status, output.Label, output.Summary = "supportive", "偏合", "当日日干与日支主气都与扶抑喜用同向，可作日程参考。"
	case gd == "favorable" && zd == "neutral":
		output.Status, output.Label = "undetermined", "日干偏合"
		output.Summary = fmt.Sprintf("日干%s（%s）属喜用；日支%s（%s主气）未落入喜忌。", gan, ganElement, zhi, zhiElement)
	case gd == "neutral" && zd == "favorable":
		output.Status, output.Label = "undetermined", "日支偏合"
		output.Summary = fmt.Sprintf("日支%s（%s主气）属喜用；日干%s（%s）未落入喜忌。", zhi, zhiElement, gan, ganElement)
	case gd == "adverse" && zd == "neutral":
		output.Status, output.Label = "undetermined", "日干需留意"
		output.Summary = fmt.Sprintf("日干%s（%s）属忌神；日支%s（%s主气）未落入喜忌。", gan, ganElement, zhi, zhiElement)
	case gd == "neutral" && zd == "adverse":
		output.Status, output.Label = "undetermined", "日支需留意"
		output.Summary = fmt.Sprintf("日支%s（%s主气）属忌神；日干%s（%s）未落入喜忌。", zhi, zhiElement, gan, ganElement)
	case gd == "neutral" && zd == "neutral":
		output.Status, output.Label = "undetermined", "均无明显喜忌"
		output.Summary = fmt.Sprintf("日干%s（%s）和日支%s（%s主气）都未落入喜忌。", gan, ganElement, zhi, zhiElement)
	default:
		output.Status, output.Label, output.Summary = "undetermined", "分别看干支", "日干与日支主气未形成一致方向，请看下方依据。"
	}
	if output.Status == "mixed" && climateConflict {
		return output, nil
	}
	gods := []string{ganGod, zhiGod}
	for _, god := range gods {
		switch god {
		case "正财", "偏财", "比肩", "劫财":
			appendActivity(output, PersonalDayActivity{"finance", "财务安排", "先核对预算", "当日十神涉及财务或资源分配，可核对预算、支出和合同条款；不构成投资判断。"})
		case "正官", "七杀":
			appendActivity(output, PersonalDayActivity{"work", "工作安排", "核对责任", "当日十神涉及规则或责任，可核对任务边界和截止时间。"})
		case "食神", "伤官":
			appendActivity(output, PersonalDayActivity{"communication", "表达交付", "检查表达", "当日十神涉及表达或输出，可检查沟通内容与交付细节。"})
		}
	}
	if clash {
		appendActivity(output, PersonalDayActivity{"schedule", "重要安排", "预留余地", "当日日支与命盘日支相冲，重要安排可再核对时间、人员与实际条件。"})
	}
	return output, nil
}

func directionLabel(direction string) string {
	switch direction {
	case "favorable":
		return "喜用"
	case "adverse":
		return "忌神"
	default:
		return "中性"
	}
}

func appendActivity(day *PersonalDayResult, activity PersonalDayActivity) {
	for _, existing := range day.Activities {
		if existing.ID == activity.ID {
			return
		}
	}
	day.Activities = append(day.Activities, activity)
}
