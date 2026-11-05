package service

import (
	"strings"
	"testing"
	"yuanju/pkg/bazi"
)

func questionChart() *bazi.BaziResult {
	chart := dailyChart("甲", "子")
	chart.YearGan, chart.YearZhi = "庚", "午"
	chart.MonthGan, chart.MonthZhi = "丁", "亥"
	chart.HourGan, chart.HourZhi = "壬", "申"
	chart.Dayun = []bazi.DayunItem{{Gan: "庚", Zhi: "申", LiuNian: []bazi.LiuNianItem{{Year: 2026, GanZhi: "丙午"}}}}
	return chart
}

func TestTodayQuestionContextKeepsLayersAndCalendarSeparate(t *testing.T) {
	ctx, err := BuildTodayQuestionContext("2026-11-05", "own-chart", "今天下午参加二面，如何准备？", questionChart())
	if err != nil {
		t.Fatal(err)
	}
	if ctx.Date != "2026-11-05" || len(ctx.Layers) < 2 || ctx.Layers[len(ctx.Layers)-1].GanZhi != "癸未" || len(ctx.CalendarYi) == 0 {
		t.Fatalf("date/calendar lost: %+v", ctx)
	}
	layers := map[string]bool{}
	for _, evidence := range ctx.Evidence {
		layers[evidence.Layer] = true
		if (evidence.Source != "本人命盘" && evidence.Source != "本人命盘·冲合参考") || evidence.ID == "" {
			t.Fatalf("untrusted evidence: %+v", evidence)
		}
	}
	if !layers["大运"] || !layers["流年"] || !layers["流日"] {
		t.Fatalf("missing relevant layers: %+v", ctx.Evidence)
	}
	for _, evidence := range ctx.Evidence {
		if strings.Contains(evidence.Text, "通用日历列出") {
			t.Fatalf("calendar promoted to personal evidence: %+v", evidence)
		}
	}
}

func TestTodayQuestionTravelDoesNotPromoteCalendarYi(t *testing.T) {
	chart := dailyChart("甲", "子")
	ctx, err := BuildTodayQuestionContext("2026-11-16", "own-chart", "今天要坐高铁去外地参加会议", chart)
	if err != nil {
		t.Fatal(err)
	}
	if len(ctx.CalendarYi) == 0 {
		t.Fatal("calendar labels missing")
	}
	for _, e := range ctx.Evidence {
		if strings.Contains(e.Text, "通用日历") {
			t.Fatalf("general almanac promoted to personal evidence: %+v", e)
		}
	}
}

func TestTodayQuestionFallsBackToIndirectPersonalDirection(t *testing.T) {
	ctx, err := BuildTodayQuestionContext("2026-11-05", "own-chart", "今天要坐高铁去外地参加会议", dailyChart("甲", "子"))
	if err != nil {
		t.Fatal(err)
	}
	if len(ctx.Evidence) != 2 || ctx.Evidence[0].Direction != "neutral" || ctx.Evidence[1].Source != "本人命盘·整体参考" || ctx.Evidence[1].Layer != "流日" {
		t.Fatalf("missing indirect personal direction: %+v", ctx.Evidence)
	}
	answer, err := ValidateTodayQuestionAnswer(ctx, `{"level":"偏有利","summary":"当日整体方向可作间接参考，不能推断这次行程结果。","support_ids":["E02"],"caution_ids":[],"advice":["提前核对车次和到站时间","确认路线及会面地点"]}`)
	if err != nil || answer.Level != "偏有利" || !answer.Indirect || len(answer.Support) != 1 || answer.Support[0].Source != "本人命盘·整体参考" || len(answer.Relations) != 1 {
		t.Fatalf("indirect evidence should be traceable: %+v %v", answer, err)
	}
	if _, err := ValidateTodayQuestionAnswer(ctx, `{"level":"偏有利","summary":"今天出行比较合适，请提前核对时间。","support_ids":["E02"],"caution_ids":[],"advice":["提前核对车次和到站时间","确认路线及会面地点"]}`); err == nil {
		t.Fatal("indirect evidence presented as direct event conclusion")
	}
	if _, err := ValidateTodayQuestionAnswer(ctx, `{"level":"依据不足","summary":"当前没有与此次出行直接相关的个人方向依据，无法判断是否有利。","support_ids":[],"caution_ids":[],"advice":["提前核对车次和到站时间","确认路线及会面地点"]}`); err == nil {
		t.Fatal("old insufficient restriction still accepted")
	}
	if _, err := ValidateTodayQuestionAnswer(ctx, `{"level":"偏有利","summary":"今天出行有明确的个人支持依据，可以顺利到达。","support_ids":[],"caution_ids":[],"advice":["提前核对车次和到站时间","确认路线及会面地点"]}`); err == nil {
		t.Fatal("uncited favorability accepted")
	}
}

func TestTodayQuestionWithoutReliableDirectionReturnsLimitedReference(t *testing.T) {
	chart := dailyChart("甲", "子")
	chart.NatalAssessment.Fuyi.Yongshen = ""
	ctx, err := BuildTodayQuestionContext("2026-11-05", "own-chart", "今天要坐高铁去外地", chart)
	if err != nil || len(ctx.Evidence) != 1 || ctx.Evidence[0].Direction != "neutral" || hasTodayQuestionDirectionalEvidence(ctx.Evidence) {
		t.Fatalf("invalid fuyi should have no directional evidence: %+v %v", ctx, err)
	}
	answer, err := ValidateTodayQuestionAnswer(ctx, `{"level":"参考有限","summary":"本人扶抑喜忌资料不完整，无法判断这次行程的有利程度。","support_ids":[],"caution_ids":[],"advice":["提前核对车次和到站时间","确认路线及会面地点"]}`)
	if err != nil || answer.Level != "参考有限" || len(answer.Relations) != 1 {
		t.Fatalf("limited reference rejected: %+v %v", answer, err)
	}
}

func TestTodayQuestionTransitionDayRemainsUnknown(t *testing.T) {
	chart := questionChart()
	chart.Dayun[0].LiuNian[0].IsTransition = true
	chart.Dayun[0].LiuNian[0].TransMonth = 11
	chart.Dayun[0].LiuNian[0].TransDay = 5
	chart.Dayun[0].LiuNian[0].PrevDayun = "己未"
	ctx, err := BuildTodayQuestionContext("2026-11-05", "own-chart", "今天有一场重要面试", chart)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(ctx.Unknown, ""), "大运交接") {
		t.Fatalf("transition not disclosed: %+v", ctx.Unknown)
	}
	for _, e := range ctx.Evidence {
		if e.Layer == "大运" {
			t.Fatalf("transition day guessed: %+v", e)
		}
	}
	months, _, err := bazi.CalcLiuYue(2026, chart.DayGan)
	if err != nil {
		t.Fatal(err)
	}
	boundary := months[9].StartDate
	ctx, err = BuildTodayQuestionContext(boundary, "own-chart", "今天有一场重要面试", chart)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(ctx.Unknown, ""), "流月节气交界") {
		t.Fatalf("month boundary not disclosed: %+v", ctx.Unknown)
	}
	ctx, err = BuildTodayQuestionContext(months[0].StartDate, "own-chart", "今天有一场重要面试", chart)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(ctx.Unknown, ""), "立春节气交界") {
		t.Fatalf("year boundary not disclosed: %+v", ctx.Unknown)
	}
	for _, e := range ctx.Evidence {
		if e.Layer == "大运" || e.Layer == "流年" {
			t.Fatalf("year boundary guessed: %+v", e)
		}
	}
}

func TestTodayQuestionAnswerRejectsFabricationAndProbability(t *testing.T) {
	ctx := &TodayQuestionContext{Date: "2026-11-05", ChartID: "own-chart", Category: "interview", Evidence: []TodayQuestionEvidence{
		{ID: "E01", Layer: "大运", Source: "本人命盘", Direction: "favorable", Text: "工作信号支持"},
		{ID: "E02", Layer: "流日", Source: "本人命盘", Direction: "review", Text: "表达细节需留意"},
	}}
	valid := `{"level":"信号交错","summary":"工作安排有支持信号，表达细节仍需准备。","support_ids":["E01"],"caution_ids":["E02"],"advice":["整理具体项目案例","核对岗位要求与时间"]}`
	answer, err := ValidateTodayQuestionAnswer(ctx, valid)
	if err != nil || len(answer.Support) != 1 || len(answer.Caution) != 1 {
		t.Fatalf("valid answer rejected: %+v %v", answer, err)
	}
	fourAdvice := strings.Replace(valid, `"核对岗位要求与时间"]`, `"核对岗位要求与时间","演练一遍自我介绍","提前确认到场路线"]`, 1)
	answer, err = ValidateTodayQuestionAnswer(ctx, fourAdvice)
	if err != nil || len(answer.Advice) != 4 {
		t.Fatalf("four safe preparation suggestions rejected: %+v %v", answer, err)
	}
	for _, raw := range []string{
		strings.Replace(valid, "E01", "E99", 1),
		strings.Replace(valid, "工作安排有支持信号", "成功率 78%", 1),
		strings.Replace(valid, "整理具体项目案例", "保证录用", 1),
		strings.Replace(valid, `"support_ids":["E01"]`, `"support_ids":["E01","E01"]`, 1),
	} {
		if _, err := ValidateTodayQuestionAnswer(ctx, raw); err == nil {
			t.Fatalf("invalid answer accepted: %s", raw)
		}
	}
	if ValidTodayQuestion("短") {
		t.Fatal("invalid question accepted")
	}
	if ValidTodayQuestion("今天应该买入股票吗？") {
		t.Fatal("unsupported financial question accepted")
	}
	if ValidTodayQuestion("今天想卖出股票吗？") {
		t.Fatal("sell recommendation question accepted")
	}
}

func TestTodayQuestionInfersEvidenceFromQuestion(t *testing.T) {
	for question, want := range map[string]string{
		"今天下午有一场面试":                  "interview",
		"帮我看看今天的投资运如何":               "finance",
		"今天适合做财务规划吗":                 "finance",
		"今天去外地坐高铁开会":                 "travel",
		"今天要出一趟远门需要注意什么吗":            "travel",
		"今天要去异地出差，途中要注意什么":           "travel",
		"今天乘动车去外地办事顺利吗":              "travel",
		"今天与客户沟通方案":                  "work",
		"今天想去谈一个合作不知道会不会成功":          "work",
		"今天约会见面顺利吗":                  "social",
		"我想约一个心仪的女孩子，约了很久了不知道能不能约得上": "social",
		"今天想向喜欢的人发出邀约":               "social",
		"今天有个心仪的工作面试":                "interview",
		"今天能做成这件事吗":                  "general",
	} {
		if got := inferTodayQuestionCategory(question); got != want {
			t.Errorf("%q: got %s, want %s", question, got, want)
		}
	}
	trip, err := BuildTodayQuestionContext("2026-11-06", "own-chart", "今天要出一趟远门需要注意什么吗", questionChart())
	if err != nil {
		t.Fatal(err)
	}
	if trip.Category != "travel" || len(trip.Evidence) == 0 || strings.Contains(strings.Join(trip.Unknown, ""), "未匹配") {
		t.Fatalf("colloquial trip should use relevant personal evidence: %+v", trip)
	}
	clashFound := false
	for _, evidence := range trip.Evidence {
		if evidence.Layer == "流年" && strings.Contains(evidence.Text, "地支六冲") {
			clashFound = true
		}
	}
	if !clashFound {
		t.Fatalf("missing 2026 午 year versus natal 子 day branch clash: %+v", trip.Evidence)
	}
	finance, err := BuildTodayQuestionContext("2026-11-06", "own-chart", "帮我看看今天的投资运如何", dailyChart("丙", "子"))
	if err != nil || finance.Category != "finance" {
		t.Fatalf("finance question not classified: %+v %v", finance, err)
	}
	financeEvidence := false
	for _, evidence := range finance.Evidence {
		if evidence.Source == "本人命盘" && evidence.Layer == "流日" && strings.Contains(evidence.Text, "投资研究与财务规划") {
			financeEvidence = true
		}
	}
	if !financeEvidence {
		t.Fatalf("finance-specific personal evidence missing: %+v", finance.Evidence)
	}
	invitation, err := BuildTodayQuestionContext("2026-11-05", "own-chart", "我想约一个心仪的女孩子，约了很久了不知道能不能约得上", questionChart())
	if err != nil || invitation.Category != "social" || len(invitation.Evidence) == 0 || strings.Contains(strings.Join(invitation.Unknown, ""), "未匹配") {
		t.Fatalf("romantic invitation should use personal social evidence: %+v, %v", invitation, err)
	}
	cooperationChart := dailyChart("丙", "子")
	cooperationChart.NatalAssessment.Fuyi.Jishen = "木火"
	cooperationChart.YearGan, cooperationChart.YearZhi = "乙", "亥"
	cooperationChart.MonthGan, cooperationChart.MonthZhi = "丙", "戌"
	cooperationChart.HourGan, cooperationChart.HourZhi = "甲", "午"
	cooperationChart.Dayun = []bazi.DayunItem{{Gan: "癸", Zhi: "未", LiuNian: []bazi.LiuNianItem{{Year: 2026, GanZhi: "丙午"}}}}
	cooperation, err := BuildTodayQuestionContext("2026-11-05", "own-chart", "今天想去谈一个合作不知道会不会成功", cooperationChart)
	if err != nil || cooperation.Category != "work" || len(cooperation.Evidence) == 0 || strings.Contains(strings.Join(cooperation.Unknown, ""), "未匹配") {
		t.Fatalf("cooperation discussion should use work evidence: %+v, %v", cooperation, err)
	}
	ctx, err := BuildTodayQuestionContext("2026-11-05", "own-chart", "今天能做成这件事吗", questionChart())
	if err != nil {
		t.Fatal(err)
	}
	if ctx.Category != "general" || len(ctx.Evidence) == 0 || ctx.Evidence[0].Source != "本人命盘·冲合参考" || !todayQuestionIsIndirect(ctx.Evidence) || !strings.Contains(strings.Join(ctx.Unknown, ""), "间接参考") {
		t.Fatalf("generic question lost personal overall reference: %+v", ctx)
	}
}

func TestFinanceRelationEvidenceUsesNatalDirection(t *testing.T) {
	chart := dailyChart("甲", "子")
	chart.NatalAssessment.Fuyi.Yongshen = "土金"
	got := []TodayQuestionEvidence{}
	add := func(layer, source, direction, value string) {
		got = append(got, TodayQuestionEvidence{Layer: layer, Source: source, Direction: direction, Text: value})
	}
	addQuestionRelationEvidence(add, "流月", "己丑", chart, "finance")
	if len(got) != 2 || got[0].Direction != "candidate" || got[1].Direction != "candidate" ||
		!strings.Contains(got[0].Text, "天干五合") || !strings.Contains(got[1].Text, "地支六合") ||
		!strings.Contains(got[0].Text, "己与本人日柱甲") || !strings.Contains(got[1].Text, "丑与本人日柱子") {
		t.Fatalf("missing traceable five/six-combination evidence: %+v", got)
	}
	got = nil
	chart.NatalAssessment.Fuyi.Yongshen = "水金"
	addQuestionRelationEvidence(add, "流月", "己丑", chart, "finance")
	if len(got) != 2 || got[0].Direction != "neutral" || got[1].Direction != "neutral" {
		t.Fatalf("neutral combination treated as investment support: %+v", got)
	}
	got = nil
	addQuestionRelationEvidence(add, "流年", "丙午", chart, "finance")
	if len(got) != 1 || got[0].Direction != "review" || !strings.Contains(got[0].Text, "午与本人日支子构成地支六冲") || !strings.Contains(got[0].Text, "实际亏损") {
		t.Fatalf("clash should be a qualified review signal: %+v", got)
	}
}

func TestFinanceContextIncludesOnlyKnownLayerRelations(t *testing.T) {
	chart := questionChart()
	ctx, err := BuildTodayQuestionContext("2026-11-06", "own-chart", "今天的投资运如何", chart)
	if err != nil {
		t.Fatal(err)
	}
	foundYearClash := false
	for _, e := range ctx.Evidence {
		if e.Source == "本人命盘·冲合参考" && e.Layer == "流年" && strings.Contains(e.Text, "午与本人日支子构成地支六冲") {
			foundYearClash = true
		}
	}
	if !foundYearClash {
		t.Fatalf("finance question missed known year clash: %+v", ctx.Evidence)
	}
	months, _, err := bazi.CalcLiuYue(2026, chart.DayGan)
	if err != nil {
		t.Fatal(err)
	}
	boundary, err := BuildTodayQuestionContext(months[0].StartDate, "own-chart", "今天的投资运如何", chart)
	if err != nil {
		t.Fatal(err)
	}
	unknown := strings.Join(boundary.Unknown, "；")
	if !strings.Contains(unknown, "流年与大运暂不判断") {
		t.Fatalf("year boundary not disclosed: %+v", boundary.Unknown)
	}
	for _, e := range boundary.Evidence {
		if e.Source == "本人命盘·冲合参考" && (e.Layer == "流年" || e.Layer == "大运") {
			t.Fatalf("uncertain solar-term layer produced relation evidence: %+v", e)
		}
	}
}

func TestFinanceDayClashSurvivesMissingFuyi(t *testing.T) {
	chart := dailyChart("甲", "丑")
	chart.NatalAssessment.Fuyi.Yongshen = ""
	chart.NatalAssessment.Fuyi.Jishen = ""
	ctx, err := BuildTodayQuestionContext("2026-11-05", "own-chart", "今天的投资运如何", chart)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, e := range ctx.Evidence {
		if e.Layer == "流日" && e.Source == "本人命盘·冲合参考" && e.Direction == "review" && strings.Contains(e.Text, "未与本人日支丑构成地支六冲") {
			found = true
		}
	}
	if !found {
		t.Fatalf("known personal day clash skipped with missing Fuyi: %+v", ctx.Evidence)
	}
}

func TestEveryTodayQuestionCategoryChecksPersonalRelations(t *testing.T) {
	for _, question := range []string{
		"今天参加面试如何准备", "今天与客户沟通如何准备", "今天坐高铁出行要注意什么",
		"今天约朋友见面合适吗", "今天的投资运如何", "今天想做一件重要的事",
	} {
		ctx, err := BuildTodayQuestionContext("2026-11-06", "own-chart", question, questionChart())
		if err != nil {
			t.Fatalf("%q: %v", question, err)
		}
		foundClash, foundDayCheck := false, false
		for _, e := range ctx.Evidence {
			if e.Source == "本人命盘·冲合参考" && e.Layer == "流年" && strings.Contains(e.Text, "午与本人日支子构成地支六冲") {
				foundClash = true
			}
			if e.Source == "本人命盘·冲合参考" && e.Layer == "流日" {
				foundDayCheck = true
			}
		}
		if !foundClash || !foundDayCheck {
			t.Errorf("%q missed shared personal relations: %+v", question, ctx.Evidence)
		}
		if ctx.Category == "general" && !todayQuestionIsIndirect(ctx.Evidence) {
			t.Errorf("unknown topic promoted a relation to direct event evidence: %+v", ctx.Evidence)
		}
	}
}

func TestCitedRelationIsNotDuplicatedInResult(t *testing.T) {
	ctx := &TodayQuestionContext{Category: "interview", Evidence: []TodayQuestionEvidence{
		{ID: "E01", Layer: "流日", Source: "本人命盘·冲合参考", Direction: "candidate", Text: "流日与本人日柱天干五合，整体参考"},
		{ID: "E02", Layer: "流日", Source: "本人命盘", Direction: "review", Text: "面试表达仍需准备"},
	}}
	answer, err := ValidateTodayQuestionAnswer(ctx, `{"level":"信号交错","summary":"整体冲合可作间接参考，面试表达仍需实际准备。","support_ids":["E01"],"caution_ids":["E02"],"advice":["整理自己的项目案例","核对岗位要求和时间"]}`)
	if err != nil || len(answer.Relations) != 0 || len(answer.Support) != 1 || answer.Indirect {
		t.Fatalf("cited relation duplicated or marked whole answer indirect: %+v %v", answer, err)
	}
}
