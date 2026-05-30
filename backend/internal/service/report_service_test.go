package service

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"text/template"

	"yuanju/internal/model"
	"yuanju/pkg/bazi"
)

func TestCachedDayunSummaryToStreamItemReturnsCachedItem(t *testing.T) {
	themes := json.RawMessage(`["学业突破","贵人扶持"]`)
	years := json.RawMessage(`[{"year":2020,"gan_zhi":"庚子","narrative":"流年顺遂"}]`)
	cached := &model.AIDayunSummary{
		DayunIndex:  2,
		DayunGanZhi: "乙卯",
		Themes:      &themes,
		Summary:     "早年学习有助力，后段适合稳扎稳打。",
		Years:       &years,
	}

	item, ok := cachedDayunSummaryToStreamItem(cached, "甲寅")

	if !ok {
		t.Fatalf("expected valid cached summary to be usable")
	}
	if !item.Cached {
		t.Fatalf("expected item to be marked cached")
	}
	if item.DayunIndex != 2 {
		t.Fatalf("unexpected dayun index: %d", item.DayunIndex)
	}
	if item.GanZhi != "乙卯" {
		t.Fatalf("expected cached gan-zhi to be preserved, got %q", item.GanZhi)
	}
	if item.Summary != cached.Summary {
		t.Fatalf("unexpected summary: %q", item.Summary)
	}
	if got := strings.Join(item.Themes, ","); got != "学业突破,贵人扶持" {
		t.Fatalf("unexpected themes: %q", got)
	}
}

func TestBuildBaziPrompt_ExcludesCelebritySectionAndPersonaChapter(t *testing.T) {
	result := &bazi.BaziResult{
		YearGan:         "甲",
		YearZhi:         "子",
		MonthGan:        "丙",
		MonthZhi:        "寅",
		DayGan:          "戊",
		DayZhi:          "辰",
		HourGan:         "庚",
		HourZhi:         "午",
		YearGanWuxing:   "木",
		YearZhiWuxing:   "水",
		MonthGanWuxing:  "火",
		MonthZhiWuxing:  "木",
		DayGanWuxing:    "土",
		DayZhiWuxing:    "土",
		HourGanWuxing:   "金",
		HourZhiWuxing:   "火",
		YearGanShiShen:  "七杀",
		MonthGanShiShen: "偏印",
		HourGanShiShen:  "食神",
		YearZhiShiShen:  []string{"正财"},
		MonthZhiShiShen: []string{"偏印"},
		DayZhiShiShen:   []string{"比肩"},
		HourZhiShiShen:  []string{"正印"},
		YearDiShi:       "胎",
		MonthDiShi:      "长生",
		DayDiShi:        "冠带",
		HourDiShi:       "临官",
		YearXunKong:     "戌亥",
		MonthXunKong:    "申酉",
		DayXunKong:      "午未",
		HourXunKong:     "辰巳",
		YearHideGan:     []string{"癸"},
		MonthHideGan:    []string{"甲", "丙", "戊"},
		DayHideGan:      []string{"戊", "乙", "癸"},
		HourHideGan:     []string{"丁", "己"},
		YearNaYin:       "海中金",
		MonthNaYin:      "炉中火",
		DayNaYin:        "大林木",
		HourNaYin:       "路旁土",
		Wuxing:          bazi.WuxingStats{Mu: 2, Huo: 2, Tu: 2, Jin: 1, Shui: 1, MuPct: 25, HuoPct: 25, TuPct: 25, JinPct: 12.5, ShuiPct: 12.5},
		Yongshen:        "火土",
		Jishen:          "水木",
		StartYunSolar:   "2000年1月1日 00:00",
		Gender:          "male",
		Dayun:           []bazi.DayunItem{{Gan: "辛", Zhi: "卯", StartAge: 3, StartYear: 2000, GanShiShen: "伤官", ZhiShiShen: "正官", DiShi: "沐浴"}},
		YearShenSha:     []string{"天乙贵人"},
		MonthShenSha:    []string{"文昌"},
		DayShenSha:      []string{"华盖"},
		HourShenSha:     []string{"桃花"},
	}

	prompt := buildBaziPrompt(result)

	if strings.Contains(prompt, "名人参考库") {
		t.Fatalf("prompt should not include celebrity reference section")
	}
	if strings.Contains(prompt, "命理分身") {
		t.Fatalf("prompt should not include persona chapter instructions")
	}
}

func TestBuildBaziPrompt_UsesSystemMingGeAsPrimarySource(t *testing.T) {
	result := &bazi.BaziResult{
		YearGan:         "甲",
		YearZhi:         "子",
		MonthGan:        "丙",
		MonthZhi:        "寅",
		DayGan:          "戊",
		DayZhi:          "辰",
		HourGan:         "庚",
		HourZhi:         "午",
		YearGanWuxing:   "木",
		YearZhiWuxing:   "水",
		MonthGanWuxing:  "火",
		MonthZhiWuxing:  "木",
		DayGanWuxing:    "土",
		DayZhiWuxing:    "土",
		HourGanWuxing:   "金",
		HourZhiWuxing:   "火",
		YearGanShiShen:  "七杀",
		MonthGanShiShen: "偏印",
		HourGanShiShen:  "食神",
		YearZhiShiShen:  []string{"正财"},
		MonthZhiShiShen: []string{"偏印"},
		DayZhiShiShen:   []string{"比肩"},
		HourZhiShiShen:  []string{"正印"},
		YearDiShi:       "胎",
		MonthDiShi:      "长生",
		DayDiShi:        "冠带",
		HourDiShi:       "临官",
		YearXunKong:     "戌亥",
		MonthXunKong:    "申酉",
		DayXunKong:      "午未",
		HourXunKong:     "辰巳",
		YearHideGan:     []string{"癸"},
		MonthHideGan:    []string{"甲", "丙", "戊"},
		DayHideGan:      []string{"戊", "乙", "癸"},
		HourHideGan:     []string{"丁", "己"},
		YearNaYin:       "海中金",
		MonthNaYin:      "炉中火",
		DayNaYin:        "大林木",
		HourNaYin:       "路旁土",
		Wuxing:          bazi.WuxingStats{Mu: 2, Huo: 2, Tu: 2, Jin: 1, Shui: 1, MuPct: 25, HuoPct: 25, TuPct: 25, JinPct: 12.5, ShuiPct: 12.5},
		Yongshen:        "火土",
		Jishen:          "水木",
		StartYunSolar:   "2000年1月1日 00:00",
		Gender:          "male",
		Dayun:           []bazi.DayunItem{{Gan: "辛", Zhi: "卯", StartAge: 3, StartYear: 2000, GanShiShen: "伤官", ZhiShiShen: "正官", DiShi: "沐浴"}},
		YearShenSha:     []string{"天乙贵人"},
		MonthShenSha:    []string{"文昌"},
		DayShenSha:      []string{"华盖"},
		HourShenSha:     []string{"桃花"},
		MingGe:          "正官格",
		MingGeDesc:      "月令官星得气，格局以正官为主。",
	}

	prompt := buildBaziPrompt(result)

	for _, want := range []string{
		"[系统定格结果]",
		"主格：正官格",
		"不得重新改判格名",
		"【格局解释 — 以系统定格为准】",
		"若局中同时存在其它明显结构，必须明确写出“兼带某某倾向”或“局中亦见某某气象”一句",
		"格局模块用于解释主格，不再重新决定格局名称",
		"开头必须先写系统主格",
		"若局中兼象明显，必须显式写出“兼带某某倾向”或“局中亦见某某气象”",
		"兼带某某倾向",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("expected prompt to contain %q", want)
		}
	}

	for _, unwanted := range []string{
		"【格局评分 — 权重25票】",
		"严格按 System Prompt 中的【格局判断规则】公式执行",
		"有透干者以透干十神定格",
	} {
		if strings.Contains(prompt, unwanted) {
			t.Fatalf("prompt should not contain %q", unwanted)
		}
	}
}

func TestBuildBaziPrompt_DegradesWhenMingGeMissing(t *testing.T) {
	result := &bazi.BaziResult{
		YearGan:         "甲",
		YearZhi:         "子",
		MonthGan:        "丙",
		MonthZhi:        "寅",
		DayGan:          "戊",
		DayZhi:          "辰",
		HourGan:         "庚",
		HourZhi:         "午",
		YearGanWuxing:   "木",
		YearZhiWuxing:   "水",
		MonthGanWuxing:  "火",
		MonthZhiWuxing:  "木",
		DayGanWuxing:    "土",
		DayZhiWuxing:    "土",
		HourGanWuxing:   "金",
		HourZhiWuxing:   "火",
		YearGanShiShen:  "七杀",
		MonthGanShiShen: "偏印",
		HourGanShiShen:  "食神",
		YearZhiShiShen:  []string{"正财"},
		MonthZhiShiShen: []string{"偏印"},
		DayZhiShiShen:   []string{"比肩"},
		HourZhiShiShen:  []string{"正印"},
		YearDiShi:       "胎",
		MonthDiShi:      "长生",
		DayDiShi:        "冠带",
		HourDiShi:       "临官",
		YearXunKong:     "戌亥",
		MonthXunKong:    "申酉",
		DayXunKong:      "午未",
		HourXunKong:     "辰巳",
		YearHideGan:     []string{"癸"},
		MonthHideGan:    []string{"甲", "丙", "戊"},
		DayHideGan:      []string{"戊", "乙", "癸"},
		HourHideGan:     []string{"丁", "己"},
		YearNaYin:       "海中金",
		MonthNaYin:      "炉中火",
		DayNaYin:        "大林木",
		HourNaYin:       "路旁土",
		Wuxing:          bazi.WuxingStats{Mu: 2, Huo: 2, Tu: 2, Jin: 1, Shui: 1, MuPct: 25, HuoPct: 25, TuPct: 25, JinPct: 12.5, ShuiPct: 12.5},
		Yongshen:        "火土",
		Jishen:          "水木",
		StartYunSolar:   "2000年1月1日 00:00",
		Gender:          "male",
		Dayun:           []bazi.DayunItem{{Gan: "辛", Zhi: "卯", StartAge: 3, StartYear: 2000, GanShiShen: "伤官", ZhiShiShen: "正官", DiShi: "沐浴"}},
		YearShenSha:     []string{"天乙贵人"},
		MonthShenSha:    []string{"文昌"},
		DayShenSha:      []string{"华盖"},
		HourShenSha:     []string{"桃花"},
	}

	prompt := buildBaziPrompt(result)

	if strings.Contains(prompt, "[系统定格结果]") {
		t.Fatalf("prompt should omit system mingge block when MingGe is missing")
	}
	if !strings.Contains(prompt, "【格局评分 — 权重25票】") {
		t.Fatalf("prompt should keep legacy geju scoring path when MingGe is missing")
	}
}

func TestBuildBaziPrompt_ReadabilityDepthConstraints(t *testing.T) {
	result := bazi.Calculate(1996, 2, 8, 20, "male", false, 0, "solar", false)
	prompt := buildBaziPrompt(result)

	for _, want := range []string{
		"500-800字",
		"精简版：每章约80-120字",
		"专业版：每章约220-350字",
		"结论、命理依据、现实表现、建议",
		"术语出现后必须紧跟白话解释",
		"印星、官杀、食伤、财星、用神、忌神、调候、格局",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("expected prompt to contain readability constraint %q", want)
		}
	}
	if strings.Contains(prompt, "写一段整体分析（300-500字）") {
		t.Fatalf("prompt should no longer keep the terse 300-500 character analysis limit")
	}
}

func TestParseMarkdownToStructured_ExcludesPersonaChapter(t *testing.T) {
	md := strings.Join([]string{
		"## 【喜用神】",
		"火土",
		"",
		"## 【忌神】",
		"水木",
		"",
		"## 【命理摘要】",
		"稳中见锋",
		"",
		"## 【命局分析总览】",
		"整体分析",
		"",
		"## 【性格特质-精简版】",
		"性格简版",
		"",
		"## 【性格特质-专业版】",
		"性格专业版",
		"",
		"## 【感情运势-精简版】",
		"感情简版",
		"",
		"## 【感情运势-专业版】",
		"感情专业版",
		"",
		"## 【事业财运-精简版】",
		"事业简版",
		"",
		"## 【事业财运-专业版】",
		"事业专业版",
		"",
		"## 【健康提示-精简版】",
		"健康简版",
		"",
		"## 【健康提示-专业版】",
		"健康专业版",
		"",
		"## 【大运走势-精简版】",
		"大运简版",
		"",
		"## 【大运走势-专业版】",
		"大运专业版",
		"",
		"## 【命理分身-精简版】",
		"命理分身简版",
		"",
		"## 【命理分身-专业版】",
		"命理分身专业版",
	}, "\n")

	parsed, brief := ParseMarkdownToStructured(md)

	if parsed == nil {
		t.Fatalf("expected structured report")
	}
	if len(parsed.Chapters) != 5 {
		t.Fatalf("expected 5 chapters without persona section, got %d", len(parsed.Chapters))
	}
	for _, chapter := range parsed.Chapters {
		if chapter.Title == "命理分身" {
			t.Fatalf("expected persona chapter to be ignored")
		}
	}
	if strings.Contains(brief, "命理分身") {
		t.Fatalf("brief content should not include persona section")
	}
}

func TestParseAIReportContent_PrefersMarkdownStructured(t *testing.T) {
	md := strings.Join([]string{
		"## 【喜用神】",
		"火土",
		"",
		"## 【忌神】",
		"水木",
		"",
		"## 【命理摘要】",
		"稳中见锋",
		"",
		"## 【命局分析总览】",
		"此命以【正官格】立局。",
		"",
		"## 【性格特质-精简版】",
		"性格简版",
		"",
		"## 【性格特质-专业版】",
		"性格专业版",
		"",
		"## 【感情运势-精简版】",
		"感情简版",
		"",
		"## 【感情运势-专业版】",
		"感情专业版",
		"",
		"## 【事业财运-精简版】",
		"事业简版",
		"",
		"## 【事业财运-专业版】",
		"事业专业版",
		"",
		"## 【健康提示-精简版】",
		"健康简版",
		"",
		"## 【健康提示-专业版】",
		"健康专业版",
		"",
		"## 【大运走势-精简版】",
		"大运简版",
		"",
		"## 【大运走势-专业版】",
		"大运专业版",
	}, "\n")

	parsed, brief, contentStructured := parseAIReportContent(md, "")

	if parsed == nil {
		t.Fatalf("expected parsed structured report")
	}
	if parsed.Analysis.Logic != "此命以【正官格】立局。" {
		t.Fatalf("unexpected analysis logic: %q", parsed.Analysis.Logic)
	}
	if contentStructured == nil {
		t.Fatalf("expected markdown content to populate content_structured")
	}

	var stored structuredReport
	if err := json.Unmarshal(*contentStructured, &stored); err != nil {
		t.Fatalf("expected valid structured json: %v", err)
	}
	if stored.Analysis.Logic != parsed.Analysis.Logic {
		t.Fatalf("expected stored structured logic to match parsed logic")
	}
	if !strings.Contains(brief, "【命局概要】\n稳中见锋") {
		t.Fatalf("expected brief content to include summary, got %q", brief)
	}
}

func TestParseAIReportContent_FallsBackToLegacyJSON(t *testing.T) {
	raw := `{"yongshen":"火土","jishen":"水木","report":"旧版摘要"}`

	parsed, brief, contentStructured := parseAIReportContent(raw, raw)

	if parsed == nil {
		t.Fatalf("expected parsed legacy report")
	}
	if parsed.Yongshen != "火土" || parsed.Jishen != "水木" {
		t.Fatalf("expected legacy yongshen/jishen to be preserved, got %q / %q", parsed.Yongshen, parsed.Jishen)
	}
	if brief != "旧版摘要" {
		t.Fatalf("expected legacy report content, got %q", brief)
	}
	if contentStructured != nil {
		t.Fatalf("expected legacy json path to keep content_structured nil")
	}
}

// ── Dayun summary prompt: 喜忌十神 注入分支渲染 ──────────────────────────

// 测试模板片段：与 report_service.go::GenerateDayunSummariesStream 的 promptTpl
// 注入块保持同步；任何 prompt 行为变更需同步更新这个 fixture。
// 这里仅渲染本次新加的 ShishenConfidence 相关 3 个分支，不渲染完整 prompt。
const shishenInjectionTplFixture = `{{if eq .ShishenConfidence "hard"}}本命喜十神：{{range $i, $s := .FavorableShishen}}{{if $i}}、{{end}}{{$s}}{{end}}；本命忌十神：{{range $i, $s := .AdverseShishen}}{{if $i}}、{{end}}{{$s}}{{end}}（强势二元判定，请以此为流年吉凶主轴）
{{else if eq .ShishenConfidence "medium"}}本命偏向喜十神：{{range $i, $s := .FavorableShishen}}{{if $i}}、{{end}}{{$s}}{{end}}；偏忌十神：{{range $i, $s := .AdverseShishen}}{{if $i}}、{{end}}{{$s}}{{end}}（中等强度，调候/格局可微调）
{{else if eq .ShishenConfidence "soft"}}本命喜忌不显（身强弱中和），{{if .TiaohouSummary}}以调候用神 {{.TiaohouSummary}} 为主{{else}}以调候为主{{end}}，AI 自行从年度 evidence 判断{{end}}`

func renderShishenInjection(t *testing.T, data model.DayunSummaryTemplateData) string {
	t.Helper()
	tmpl, err := template.New("shishen_inject").Parse(shishenInjectionTplFixture)
	if err != nil {
		t.Fatalf("template parse failed: %v", err)
	}
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("template execute failed: %v", err)
	}
	return buf.String()
}

func TestDayunSummaryPrompt_HardConfidence_EmitsExplicitLists(t *testing.T) {
	out := renderShishenInjection(t, model.DayunSummaryTemplateData{
		ShishenConfidence: bazi.ShishenConfHard,
		FavorableShishen:  []string{"食神", "伤官", "偏财"},
		AdverseShishen:    []string{"比肩", "劫财"},
	})
	if !strings.Contains(out, "本命喜十神：食神、伤官、偏财") {
		t.Errorf("hard band should list favorable shishen explicitly; got: %s", out)
	}
	if !strings.Contains(out, "本命忌十神：比肩、劫财") {
		t.Errorf("hard band should list adverse shishen explicitly; got: %s", out)
	}
	if !strings.Contains(out, "强势二元判定") {
		t.Errorf("hard band should carry the '强势二元判定' hint; got: %s", out)
	}
}

func TestDayunSummaryPrompt_MediumConfidence_EmitsSoftenedLists(t *testing.T) {
	out := renderShishenInjection(t, model.DayunSummaryTemplateData{
		ShishenConfidence: bazi.ShishenConfMedium,
		FavorableShishen:  []string{"偏印", "正印", "比肩"},
		AdverseShishen:    []string{"正官", "七杀"},
	})
	if !strings.Contains(out, "本命偏向喜十神：偏印、正印、比肩") {
		t.Errorf("medium band should use '偏向' wording; got: %s", out)
	}
	if !strings.Contains(out, "中等强度") {
		t.Errorf("medium band should carry the '中等强度' hint; got: %s", out)
	}
}

func TestDayunSummaryPrompt_SoftConfidence_FallsBackToTiaohou(t *testing.T) {
	out := renderShishenInjection(t, model.DayunSummaryTemplateData{
		ShishenConfidence: bazi.ShishenConfSoft,
		TiaohouSummary:    "丙、丁火",
	})
	if !strings.Contains(out, "喜忌不显") {
		t.Errorf("soft band should declare '喜忌不显'; got: %s", out)
	}
	if !strings.Contains(out, "调候用神 丙、丁火") {
		t.Errorf("soft band should reference tiaohou summary; got: %s", out)
	}
	if strings.Contains(out, "本命喜十神") || strings.Contains(out, "本命忌十神") {
		t.Errorf("soft band must NOT emit shishen lists; got: %s", out)
	}
}

func TestDayunSummaryPrompt_SoftConfidence_OmitsTiaohouSentenceWhenAbsent(t *testing.T) {
	out := renderShishenInjection(t, model.DayunSummaryTemplateData{
		ShishenConfidence: bazi.ShishenConfSoft,
		TiaohouSummary:    "",
	})
	if !strings.Contains(out, "喜忌不显") {
		t.Errorf("soft band w/o tiaohou should still declare '喜忌不显'; got: %s", out)
	}
	if strings.Contains(out, "调候用神 ") {
		t.Errorf("soft band w/o tiaohou should not show empty tiaohou label; got: %s", out)
	}
}

// ── computeAutoGenDayunIndexes 测试 ──────────────────────────────────

// helper for building DayunItem slices in tests; StartYear 由 birthYear + StartAge 推算，
// 用于模拟"起运正好踩在生日整数年"的常规命盘。需要 StartYear 与 birthYear+StartAge
// 偏移一年的边界场景时，请直接构造 []bazi.DayunItem 字面量。
func mkDayuns(birthYear int, starts ...int) []bazi.DayunItem {
	out := make([]bazi.DayunItem, len(starts))
	for i, s := range starts {
		out[i] = bazi.DayunItem{Index: i + 1, StartAge: s, StartYear: birthYear + s}
	}
	return out
}

func TestComputeAutoGenDayunIndexes_MidLifeUser(t *testing.T) {
	// 1995 生，2026 年 → 31 岁
	dayuns := mkDayuns(1995, 0, 9, 19, 29, 39, 49, 59, 69, 79)
	got := computeAutoGenDayunIndexesAt(dayuns, 2026)
	want := []int{1, 2, 3, 4} // 含当前段 (StartYear=2024 ≤ 2026)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("31岁 命主 expected %v, got %v", want, got)
	}
}

func TestComputeAutoGenDayunIndexes_VeryYoungUser(t *testing.T) {
	// 2020 生，2026 年 → 6 岁
	dayuns := mkDayuns(2020, 0, 9, 19, 29, 39, 49, 59, 69, 79)
	got := computeAutoGenDayunIndexesAt(dayuns, 2026)
	want := []int{1} // 只有 dayun 1 起始年 ≤ 2026
	if !reflect.DeepEqual(got, want) {
		t.Errorf("6岁 命主 expected %v, got %v", want, got)
	}
}

func TestComputeAutoGenDayunIndexes_ElderlyUser(t *testing.T) {
	// 1950 生，2026 年 → 76 岁
	dayuns := mkDayuns(1950, 0, 9, 19, 29, 39, 49, 59, 69, 79)
	got := computeAutoGenDayunIndexesAt(dayuns, 2026)
	want := []int{1, 2, 3, 4, 5, 6, 7, 8} // 第 9 段 StartYear=2029 > 2026，排除
	if !reflect.DeepEqual(got, want) {
		t.Errorf("76岁 命主 expected %v, got %v", want, got)
	}
}

func TestComputeAutoGenDayunIndexes_BoundaryAtStartYear(t *testing.T) {
	// 1996 生，2026 年，某段 StartYear 正好等于 currentYear
	dayuns := mkDayuns(1996, 0, 10, 20, 30, 40)
	got := computeAutoGenDayunIndexesAt(dayuns, 2026)
	want := []int{1, 2, 3, 4} // dayun 4 StartYear=2026 等于 currentYear=2026，包含
	if !reflect.DeepEqual(got, want) {
		t.Errorf("边界 currentYear==StartYear expected %v, got %v", want, got)
	}
}

func TestComputeAutoGenDayunIndexes_FutureBirth(t *testing.T) {
	// 防御性边界：BirthYear > CurrentYear（不可能但代码不应崩）
	dayuns := mkDayuns(2030, 0, 9, 19)
	got := computeAutoGenDayunIndexesAt(dayuns, 2026)
	if len(got) != 0 {
		t.Errorf("未来出生命主 expected empty, got %v", got)
	}
}

func TestComputeAutoGenDayunIndexes_AtFirstYearOfNewDayun(t *testing.T) {
	// 起运月日导致 StartYear 比 birthYear+StartAge 早一年（公元年提前跨进新段，
	// 但命主还没满 StartAge）。判定应以 StartYear ≤ currentYear 为准，否则
	// 前端把这段标 loading=true、后端却跳过推送 → 死锁转圈。
	dayuns := []bazi.DayunItem{
		{Index: 1, StartAge: 2, StartYear: 1996},
		{Index: 2, StartAge: 12, StartYear: 2006},
		{Index: 3, StartAge: 22, StartYear: 2016},
		{Index: 4, StartAge: 32, StartYear: 2026},
		{Index: 5, StartAge: 42, StartYear: 2036},
	}
	got := computeAutoGenDayunIndexesAt(dayuns, 2026)
	want := []int{1, 2, 3, 4}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("命主踏入新大运起始年 expected %v, got %v", want, got)
	}
}

func TestComputeAutoGenDayunIndexes_EmptyDayuns(t *testing.T) {
	got := computeAutoGenDayunIndexesAt([]bazi.DayunItem{}, 2026)
	if len(got) != 0 {
		t.Errorf("空 dayun 列表 expected empty, got %v", got)
	}
}

func TestFillBlankYearNarratives_EmptyNarrativeGetsFallback(t *testing.T) {
	parsed := []parsedYearAI{
		{Year: 2020, GanZhi: "庚子", Narrative: ""},
	}
	signals := []bazi.YearSignals{
		{Year: 2020, Age: 25, GanZhi: "庚子", DayunGanZhi: "甲寅",
			Signals: []bazi.EventSignal{
				{Type: "用神基底", Source: bazi.SourceYongshen, Polarity: bazi.PolarityJi},
			}},
	}
	out := fillBlankYearNarratives(parsed, signals, 1)
	if len(out) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(out))
	}
	if out[0].Narrative == "" {
		t.Error("blank AI narrative should be filled by fallback")
	}
	if !strings.Contains(out[0].Narrative, "庚子") {
		t.Errorf("fallback should reference ganzhi; got %q", out[0].Narrative)
	}
}

func TestFillBlankYearNarratives_ValidAIPreserved(t *testing.T) {
	parsed := []parsedYearAI{
		{Year: 2020, GanZhi: "庚子", Narrative: "庚子年食神高透，事业稳步推进。"},
	}
	signals := []bazi.YearSignals{
		{Year: 2020, Age: 25, GanZhi: "庚子",
			Signals: []bazi.EventSignal{
				{Type: "事业", Evidence: "食神高透", Polarity: bazi.PolarityJi, Source: "天干"},
			}},
	}
	out := fillBlankYearNarratives(parsed, signals, 1)
	if out[0].Narrative != "庚子年食神高透，事业稳步推进。" {
		t.Errorf("valid AI narrative should be preserved verbatim; got %q", out[0].Narrative)
	}
}

func TestFillBlankYearNarratives_ValidatorWipedGetsFallback(t *testing.T) {
	// AI 写了"用神位受冲"但 evidence 没有"用神位" → validator 清空 → 兜底
	parsed := []parsedYearAI{
		{Year: 2020, GanZhi: "庚子", Narrative: "庚子年用神位受冲，运势波动。"},
	}
	signals := []bazi.YearSignals{
		{Year: 2020, Age: 25, GanZhi: "庚子",
			Signals: []bazi.EventSignal{
				{Type: "用神基底", Source: bazi.SourceYongshen, Polarity: bazi.PolarityXiong, Evidence: "日干受克"},
			}},
	}
	out := fillBlankYearNarratives(parsed, signals, 1)
	if out[0].Narrative == "庚子年用神位受冲，运势波动。" {
		t.Error("validator should have wiped the narrative")
	}
	if out[0].Narrative == "" {
		t.Error("wiped narrative should be replaced by fallback, not left empty")
	}
	if !strings.Contains(out[0].Narrative, "偏凶") {
		t.Errorf("xiong basis should produce 偏凶 fallback; got %q", out[0].Narrative)
	}
}


func TestExtractJSON(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"裸JSON", `{"a":1}`, `{"a":1}`},
		{"带空白裸JSON", "  \n{\"a\":1}\n ", `{"a":1}`},
		{"json围栏", "```json\n{\"a\":1}\n```", `{"a":1}`},
		{"裸围栏", "```\n{\"a\":1}\n```", `{"a":1}`},
		{"前置散文加围栏", "好的，以下是分析：\n```json\n{\"summary\":\"x\"}\n```", `{"summary":"x"}`},
		{"散文包裹裸JSON", "分析如下：{\"k\":\"v\"} 完毕", `{"k":"v"}`},
		{"无花括号", "no json here", "no json here"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := extractJSON(c.in); got != c.want {
				t.Errorf("extractJSON(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestStripTrailingCommas(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"对象尾逗号", `{"a":1,}`, `{"a":1}`},
		{"数组尾逗号", `[1,2,]`, `[1,2]`},
		{"数组后对象闭合前尾逗号", `{"k":[{"x":1}],}`, `{"k":[{"x":1}]}`},
		{"带空白换行的尾逗号", "{\"a\":1,\n  }", "{\"a\":1\n  }"},
		{"合法逗号不动", `{"a":1,"b":2}`, `{"a":1,"b":2}`},
		{"字符串内逗号不动", `{"a":"x, }","b":1}`, `{"a":"x, }","b":1}`},
		{"字符串内伪尾逗号不动", `{"a":"foo,]"}`, `{"a":"foo,]"}`},
		{"转义引号不误判", `{"a":"he said \"hi,\" ok","b":2}`, `{"a":"he said \"hi,\" ok","b":2}`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := stripTrailingCommas(c.in); got != c.want {
				t.Errorf("stripTrailingCommas(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

// 复现合盘报告解析失败的真实片段（personality_comparison 末尾尾逗号），
// 验证 stripTrailingCommas 修复后可被 encoding/json 正常解析。
func TestStripTrailingCommas_RealReportSnippet(t *testing.T) {
	bad := `{
  "personality_comparison": {
    "clash_points": [
      { "title": "节奏", "detail": "你追问，他后退。" }
    ],
  },
  "decision_advice": { "recommendation": "observe" }
}`
	repaired := stripTrailingCommas(bad)
	var out map[string]any
	if err := json.Unmarshal([]byte(repaired), &out); err != nil {
		t.Fatalf("修复后仍解析失败: %v\n修复结果:\n%s", err, repaired)
	}
}
