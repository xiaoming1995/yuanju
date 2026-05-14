package service

import (
	"encoding/json"
	"strings"
	"testing"

	"yuanju/internal/model"
	"yuanju/pkg/bazi"
)

func TestCachedDayunSummaryToStreamItemReturnsCachedItem(t *testing.T) {
	themes := json.RawMessage(`["学业突破","贵人扶持"]`)
	cached := &model.AIDayunSummary{
		DayunIndex:  2,
		DayunGanZhi: "乙卯",
		Themes:      &themes,
		Summary:     "早年学习有助力，后段适合稳扎稳打。",
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
