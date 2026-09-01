package service

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"text/template"
	"time"

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

func TestBuildLiunianGenderPromptContextFemaleUsesFuXingRule(t *testing.T) {
	genderLabel, relationshipRule, guardRule := buildLiunianGenderPromptContext("female")

	if genderLabel != "女命" {
		t.Fatalf("gender label = %q, want 女命", genderLabel)
	}
	for _, want := range []string{"官杀", "夫星"} {
		if !strings.Contains(relationshipRule, want) {
			t.Fatalf("female relationship rule missing %q: %s", want, relationshipRule)
		}
	}
	for _, forbidden := range []string{"男命以财为妻星", "妻星"} {
		if !strings.Contains(guardRule, forbidden) {
			t.Fatalf("female guard rule should explicitly forbid %q: %s", forbidden, guardRule)
		}
	}
}

func TestPrependLiunianGenderGuardKeepsFemaleContextWithLegacyPrompt(t *testing.T) {
	data := model.LiunianTemplateData{
		NatalAnalysisLogic:     "原局分析",
		GenderLabel:            "女命",
		DayGan:                 "壬",
		RelationshipStarRule:   "女命婚恋以官杀为夫星。",
		GenderGuardRule:        "严禁出现男命以财为妻星。",
		CurrentDayunGanZhi:     "丙子",
		CurrentDayunGanShiShen: "偏财",
		CurrentDayunZhiShiShen: "劫财",
		TargetYear:             2026,
		TargetYearGanZhi:       "丙午",
		TargetYearGanShiShen:   "偏财",
		TargetYearZhiShiShen:   "正财",
	}
	tpl := `该用户的原局分析：
{{.NatalAnalysisLogic}}

请为他详细批断【{{.TargetYear}} {{.TargetYearGanZhi}}流年】运程。`
	parsed, err := template.New("legacy_liunian").Parse(tpl)
	if err != nil {
		t.Fatal(err)
	}
	var rendered strings.Builder
	if err := parsed.Execute(&rendered, data); err != nil {
		t.Fatal(err)
	}

	prompt := prependLiunianGenderGuard(rendered.String(), data)
	for _, want := range []string{"性别：女命", "日主：壬", "女命婚恋以官杀为夫星", "严禁出现男命以财为妻星"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q:\n%s", want, prompt)
		}
	}
}

func TestValidateLiunianGenderConsistencyRejectsOppositeGenderTerms(t *testing.T) {
	if err := validateLiunianGenderConsistency("female", `{"romance":"男命以财为妻星，桃花运明显。"}`); err == nil {
		t.Fatal("expected female report with male spouse-star wording to be rejected")
	}
	if err := validateLiunianGenderConsistency("male", `{"romance":"女命以官杀为夫星，关系推进。"}`); err == nil {
		t.Fatal("expected male report with female spouse-star wording to be rejected")
	}
	if err := validateLiunianGenderConsistency("female", `{"romance":"女命本年宜看官杀与夫妻宫互动。"}`); err != nil {
		t.Fatalf("expected valid female report, got %v", err)
	}
}

func TestGenerateLiunianReportReturnsCachedReportByDefault(t *testing.T) {
	origGet := getLiunianReport
	defer func() { getLiunianReport = origGet }()

	raw := json.RawMessage(`{"career":"已缓存"}`)
	getCalled := false
	getLiunianReport = func(chartID string, targetYear int) (*model.AILiunianReport, error) {
		getCalled = true
		if chartID != "chart-1" || targetYear != 2026 {
			t.Fatalf("unexpected cache key chart=%s year=%d", chartID, targetYear)
		}
		return &model.AILiunianReport{
			ID:                "report-1",
			ChartID:           chartID,
			TargetYear:        targetYear,
			ContentStructured: &raw,
		}, nil
	}

	report, cached, err := GenerateLiunianReport("chart-1", 2026, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if !getCalled {
		t.Fatal("expected cache lookup")
	}
	if !cached {
		t.Fatal("expected cached=true")
	}
	if report == nil || report.ID != "report-1" {
		t.Fatalf("unexpected report: %+v", report)
	}
}

func TestBuildLiunianLiuYueTemplateDataReturnsTwelveMonths(t *testing.T) {
	months := buildLiunianLiuYueTemplateData(2026, "甲")

	if len(months) != 12 {
		t.Fatalf("expected 12 liuyue months, got %d", len(months))
	}
	first := months[0]
	if first.Index != 0 || first.MonthLabel != "2月" || first.MonthName != "寅月" || first.GanZhi == "" {
		t.Fatalf("unexpected first month: %+v", first)
	}
	last := months[11]
	if last.Index != 11 || last.MonthLabel != "1月" || last.MonthName != "丑月" || last.GanZhi == "" {
		t.Fatalf("unexpected last month: %+v", last)
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

func TestBuildBaziPromptIncludesGongJiaContext(t *testing.T) {
	result := &bazi.BaziResult{
		YearGan:         "甲",
		YearZhi:         "子",
		MonthGan:        "甲",
		MonthZhi:        "寅",
		DayGan:          "庚",
		DayZhi:          "午",
		HourGan:         "戊",
		HourZhi:         "申",
		YearGanWuxing:   "木",
		YearZhiWuxing:   "水",
		MonthGanWuxing:  "木",
		MonthZhiWuxing:  "木",
		DayGanWuxing:    "金",
		DayZhiWuxing:    "火",
		HourGanWuxing:   "土",
		HourZhiWuxing:   "金",
		YearGanShiShen:  "偏财",
		MonthGanShiShen: "偏财",
		HourGanShiShen:  "偏印",
		YearZhiShiShen:  []string{"伤官"},
		MonthZhiShiShen: []string{"偏财"},
		DayZhiShiShen:   []string{"正官"},
		HourZhiShiShen:  []string{"比肩"},
		YearDiShi:       "死",
		MonthDiShi:      "绝",
		DayDiShi:        "沐浴",
		HourDiShi:       "临官",
		YearXunKong:     "戌亥",
		MonthXunKong:    "子丑",
		DayXunKong:      "戌亥",
		HourXunKong:     "寅卯",
		YearHideGan:     []string{"癸"},
		MonthHideGan:    []string{"甲", "丙", "戊"},
		DayHideGan:      []string{"丁", "己"},
		HourHideGan:     []string{"庚", "壬", "戊"},
		YearNaYin:       "海中金",
		MonthNaYin:      "大溪水",
		DayNaYin:        "路旁土",
		HourNaYin:       "大驿土",
		Wuxing:          bazi.WuxingStats{Mu: 3, Huo: 1, Tu: 1, Jin: 2, Shui: 1},
		Gender:          "male",
	}
	bazi.EnsureGongJia(result)

	prompt := buildBaziPrompt(result)

	shenshaIndex := strings.Index(prompt, "[神煞]")
	gongJiaIndex := strings.Index(prompt, "[原局夹拱]")
	dayunIndex := strings.Index(prompt, "[大运序列]")
	if shenshaIndex == -1 {
		t.Fatalf("expected prompt to contain [神煞]")
	}
	if gongJiaIndex == -1 {
		t.Fatalf("expected prompt to contain [原局夹拱]")
	}
	if dayunIndex == -1 {
		t.Fatalf("expected prompt to contain [大运序列]")
	}
	if !(shenshaIndex < gongJiaIndex && gongJiaIndex < dayunIndex) {
		t.Fatalf("expected block order [神煞] -> [原局夹拱] -> [大运序列], got indexes %d, %d, %d", shenshaIndex, gongJiaIndex, dayunIndex)
	}
	if got := strings.Count(prompt, "[原局夹拱]"); got != 1 {
		t.Fatalf("expected [原局夹拱] exactly once, got %d", got)
	}
	if strings.Contains(prompt, "[神煞]\n年柱：无 | 月柱：无 | 日柱：无 | 时柱：无\n\n\n[原局夹拱]") {
		t.Fatalf("prompt should not contain an extra blank line before [原局夹拱]")
	}
	if !strings.Contains(prompt, "拱神煞=天乙贵人") {
		t.Fatalf("expected prompt to include GongJia shensha")
	}
	for _, want := range []string{
		"[原局夹拱]",
		"年月夹丑",
		"暗藏虚支",
		"不改原局五行、用神或命格",
		"天乙贵人",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("expected prompt to contain %q", want)
		}
	}
}

func TestBuildBaziPromptOmitsGongJiaContextWhenEmpty(t *testing.T) {
	result := &bazi.BaziResult{
		YearGan:         "甲",
		YearZhi:         "子",
		MonthGan:        "乙",
		MonthZhi:        "丑",
		DayGan:          "庚",
		DayZhi:          "午",
		HourGan:         "戊",
		HourZhi:         "申",
		YearGanWuxing:   "木",
		YearZhiWuxing:   "水",
		MonthGanWuxing:  "木",
		MonthZhiWuxing:  "土",
		DayGanWuxing:    "金",
		DayZhiWuxing:    "火",
		HourGanWuxing:   "土",
		HourZhiWuxing:   "金",
		YearGanShiShen:  "偏财",
		MonthGanShiShen: "正财",
		HourGanShiShen:  "偏印",
		YearZhiShiShen:  []string{"伤官"},
		MonthZhiShiShen: []string{"正印"},
		DayZhiShiShen:   []string{"正官"},
		HourZhiShiShen:  []string{"比肩"},
		YearDiShi:       "死",
		MonthDiShi:      "墓",
		DayDiShi:        "沐浴",
		HourDiShi:       "临官",
		YearXunKong:     "戌亥",
		MonthXunKong:    "子丑",
		DayXunKong:      "戌亥",
		HourXunKong:     "寅卯",
		YearHideGan:     []string{"癸"},
		MonthHideGan:    []string{"己", "癸", "辛"},
		DayHideGan:      []string{"丁", "己"},
		HourHideGan:     []string{"庚", "壬", "戊"},
		YearNaYin:       "海中金",
		MonthNaYin:      "海中金",
		DayNaYin:        "路旁土",
		HourNaYin:       "大驿土",
		Wuxing:          bazi.WuxingStats{Mu: 2, Huo: 1, Tu: 2, Jin: 2, Shui: 1},
		Gender:          "male",
	}

	prompt := buildBaziPrompt(result)

	if strings.Contains(prompt, "[原局夹拱]") {
		t.Fatalf("prompt should omit [原局夹拱] when GongJia is empty")
	}
	if !strings.Contains(prompt, "[神煞]\n年柱：无 | 月柱：无 | 日柱：无 | 时柱：无\n\n[大运序列]") {
		t.Fatalf("prompt should keep sensible separation between [神煞] and [大运序列] when GongJia is empty")
	}
}

func TestFormatGongJiaSummaryUsesFallbackForEmptyHideGan(t *testing.T) {
	result := &bazi.BaziResult{
		GongJia: []bazi.GongJiaItem{{
			Source:     "year_month",
			SourceZhis: []string{"子", "寅"},
			VirtualZhi: "丑",
			HideGan:    nil,
			ShiShen:    []string{"正印"},
			ShenSha:    []string{"天乙贵人"},
		}},
	}

	summary := formatGongJiaSummary(result)

	if !strings.HasPrefix(summary, "[原局夹拱]\n") {
		t.Fatalf("summary should not start with a leading blank line, got %q", summary)
	}
	if !strings.Contains(summary, "藏干无") {
		t.Fatalf("empty HideGan should render as 无, got %q", summary)
	}
}

func TestLoadOrCalculateResultBackfillsGongJiaSnapshot(t *testing.T) {
	cached := bazi.BaziResult{
		YearGan:           "甲",
		YearZhi:           "子",
		MonthGan:          "甲",
		MonthZhi:          "寅",
		DayGan:            "庚",
		DayZhi:            "午",
		HourGan:           "戊",
		HourZhi:           "申",
		Yongshen:          "土金",
		Jishen:            "火木",
		ShishenConfidence: bazi.ShishenConfHard,
		FavorableShishen:  []string{"偏印", "正印"},
		AdverseShishen:    []string{"正官", "七杀"},
		GongJia:           nil,
		TenGodRelation:    &bazi.TenGodRelationMatrix{},
		YearGanShiShen:    "偏财",
		MonthGanShiShen:   "偏财",
		HourGanShiShen:    "偏印",
		YearZhiShiShen:    []string{"伤官"},
		MonthZhiShiShen:   []string{"偏财"},
		DayZhiShiShen:     []string{"正官"},
		HourZhiShiShen:    []string{"比肩"},
		YearHideGan:       []string{"癸"},
		MonthHideGan:      []string{"甲", "丙", "戊"},
		DayHideGan:        []string{"丁", "己"},
		HourHideGan:       []string{"庚", "壬", "戊"},
	}
	raw, err := json.Marshal(cached)
	if err != nil {
		t.Fatalf("marshal cached result: %v", err)
	}

	originalGet := getChartResultJSON
	originalSave := saveChartResultJSON
	t.Cleanup(func() {
		getChartResultJSON = originalGet
		saveChartResultJSON = originalSave
	})

	getChartResultJSON = func(chartID string) ([]byte, error) {
		if chartID != "chart-gongjia-backfill" {
			t.Fatalf("unexpected chart id for get: %s", chartID)
		}
		return raw, nil
	}
	var saved []byte
	saveChartResultJSON = func(chartID string, resultJSON []byte) error {
		if chartID != "chart-gongjia-backfill" {
			t.Fatalf("unexpected chart id for save: %s", chartID)
		}
		saved = append([]byte(nil), resultJSON...)
		return nil
	}

	result, err := LoadOrCalculateResult(&model.BaziChart{ID: "chart-gongjia-backfill"})
	if err != nil {
		t.Fatalf("LoadOrCalculateResult returned error: %v", err)
	}
	if len(result.GongJia) == 0 {
		t.Fatalf("expected result to be backfilled with GongJia")
	}
	if len(saved) == 0 {
		t.Fatalf("expected upgraded cached snapshot to be persisted when only GongJia is backfilled")
	}
	var persisted bazi.BaziResult
	if err := json.Unmarshal(saved, &persisted); err != nil {
		t.Fatalf("saved snapshot should be valid JSON: %v", err)
	}
	if persisted.ShishenConfidence != bazi.ShishenConfHard {
		t.Fatalf("ShishenConfidence should be preserved, got %q", persisted.ShishenConfidence)
	}
	if len(persisted.GongJia) == 0 {
		t.Fatalf("saved snapshot should contain backfilled gong_jia")
	}
}

func TestLoadOrCalculateResultBackfillsVehicleRoadSnapshot(t *testing.T) {
	cached := bazi.Calculate(1995, 10, 12, 12, "male", false, 0, "solar", false)
	cached.NatalAssessment = nil
	cached.VehicleProfile = nil
	cached.DayunRoadmap = nil
	raw, err := json.Marshal(cached)
	if err != nil {
		t.Fatalf("marshal cached result: %v", err)
	}

	originalGet := getChartResultJSON
	originalSave := saveChartResultJSON
	t.Cleanup(func() {
		getChartResultJSON = originalGet
		saveChartResultJSON = originalSave
	})

	getChartResultJSON = func(chartID string) ([]byte, error) {
		if chartID != "chart-vehicle-road-backfill" {
			t.Fatalf("unexpected chart id for get: %s", chartID)
		}
		return raw, nil
	}
	var saved []byte
	saveChartResultJSON = func(chartID string, resultJSON []byte) error {
		if chartID != "chart-vehicle-road-backfill" {
			t.Fatalf("unexpected chart id for save: %s", chartID)
		}
		saved = append([]byte(nil), resultJSON...)
		return nil
	}

	result, err := LoadOrCalculateResult(&model.BaziChart{ID: "chart-vehicle-road-backfill"})
	if err != nil {
		t.Fatalf("LoadOrCalculateResult returned error: %v", err)
	}
	if result.VehicleProfile == nil {
		t.Fatalf("expected result to be backfilled with vehicle_profile")
	}
	if result.NatalAssessment == nil || result.NatalAssessment.Version != bazi.NatalAssessmentVersion {
		t.Fatalf("expected result to be backfilled with natal assessment, got %+v", result.NatalAssessment)
	}
	if len(result.DayunRoadmap) != len(result.Dayun) {
		t.Fatalf("expected dayun_roadmap length %d, got %d", len(result.Dayun), len(result.DayunRoadmap))
	}
	if len(saved) == 0 {
		t.Fatalf("expected upgraded cached snapshot to be persisted")
	}
	var persisted bazi.BaziResult
	if err := json.Unmarshal(saved, &persisted); err != nil {
		t.Fatalf("saved snapshot should be valid JSON: %v", err)
	}
	if persisted.VehicleProfile == nil {
		t.Fatalf("saved snapshot should contain vehicle_profile")
	}
	if persisted.NatalAssessment == nil || persisted.NatalAssessment.Version != bazi.NatalAssessmentVersion {
		t.Fatalf("saved snapshot should contain current natal assessment, got %+v", persisted.NatalAssessment)
	}
	if len(persisted.DayunRoadmap) != len(persisted.Dayun) {
		t.Fatalf("saved snapshot should contain dayun_roadmap")
	}
}

func TestLoadOrCalculateResultBackfillsWealthProfileSnapshot(t *testing.T) {
	cached := bazi.Calculate(1995, 10, 12, 12, "male", false, 0, "solar", false)
	cached.WealthProfile = nil
	raw, err := json.Marshal(cached)
	if err != nil {
		t.Fatalf("marshal cached result: %v", err)
	}

	originalGet := getChartResultJSON
	originalSave := saveChartResultJSON
	t.Cleanup(func() {
		getChartResultJSON = originalGet
		saveChartResultJSON = originalSave
	})

	getChartResultJSON = func(chartID string) ([]byte, error) {
		if chartID != "chart-wealth-backfill" {
			t.Fatalf("unexpected chart id for get: %s", chartID)
		}
		return raw, nil
	}
	var saved []byte
	saveChartResultJSON = func(chartID string, resultJSON []byte) error {
		if chartID != "chart-wealth-backfill" {
			t.Fatalf("unexpected chart id for save: %s", chartID)
		}
		saved = append([]byte(nil), resultJSON...)
		return nil
	}

	result, err := LoadOrCalculateResult(&model.BaziChart{ID: "chart-wealth-backfill"})
	if err != nil {
		t.Fatalf("LoadOrCalculateResult returned error: %v", err)
	}
	if result.WealthProfile == nil || result.WealthProfile.Version != bazi.WealthProfileVersion {
		t.Fatalf("expected result to be backfilled with current wealth profile, got %+v", result.WealthProfile)
	}
	if len(saved) == 0 {
		t.Fatalf("expected upgraded cached snapshot to be persisted")
	}
	var persisted bazi.BaziResult
	if err := json.Unmarshal(saved, &persisted); err != nil {
		t.Fatalf("saved snapshot should be valid JSON: %v", err)
	}
	if persisted.WealthProfile == nil || persisted.WealthProfile.Version != bazi.WealthProfileVersion {
		t.Fatalf("saved snapshot should contain current wealth profile, got %+v", persisted.WealthProfile)
	}
}

func TestLoadOrCalculateResultBackfillsMingGeSnapshot(t *testing.T) {
	cached := bazi.Calculate(1995, 10, 12, 12, "male", false, 0, "solar", false)
	expectedVehicleScore := cached.VehicleProfile.Score
	cached.MingGe = ""
	cached.MingGeDesc = ""
	raw, err := json.Marshal(cached)
	if err != nil {
		t.Fatalf("marshal cached result: %v", err)
	}

	originalGet := getChartResultJSON
	originalSave := saveChartResultJSON
	t.Cleanup(func() {
		getChartResultJSON = originalGet
		saveChartResultJSON = originalSave
	})
	getChartResultJSON = func(chartID string) ([]byte, error) { return raw, nil }
	var saved []byte
	saveChartResultJSON = func(_ string, resultJSON []byte) error {
		saved = append([]byte(nil), resultJSON...)
		return nil
	}

	result, err := LoadOrCalculateResult(&model.BaziChart{ID: "chart-mingge-backfill"})
	if err != nil {
		t.Fatalf("LoadOrCalculateResult returned error: %v", err)
	}
	if result.MingGe == "" || result.MingGeDesc == "" {
		t.Fatalf("expected Ming Ge fields to be backfilled, got %+v", result)
	}
	if result.VehicleProfile == nil || result.VehicleProfile.Score != expectedVehicleScore {
		t.Fatal("Ming Ge backfill must preserve unrelated snapshot data")
	}
	var persisted bazi.BaziResult
	if err := json.Unmarshal(saved, &persisted); err != nil {
		t.Fatalf("saved snapshot should be valid JSON: %v", err)
	}
	if persisted.MingGe != result.MingGe || persisted.MingGeDesc != result.MingGeDesc {
		t.Fatalf("saved snapshot missing Ming Ge backfill: %+v", persisted)
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

func TestBuildBaziPromptIncludesVehicleRoadContext(t *testing.T) {
	result := bazi.Calculate(1995, 10, 12, 12, "male", false, 0, "solar", false)
	prompt := buildBaziPrompt(result)

	for _, want := range []string{
		"[命盘座驾与大运路况-算法精算]",
		"命盘座驾：",
		"座驾等级表示原局基础层次：调候急需优先；无急需时以扶抑为基线，再看日干调候成格、主格结构、制化与流通",
		"日干调候可用性=resolved（得分=12，所需=甲、壬，透=甲，藏=甲、壬）",
		"日干调候成格=formed/high（基础加成=24）",
		"主格结构=偏印格/partial",
		"解释边界：日干调候天透地藏成格代表高格基础",
		"寒热调候=偏燥/seasonal_partial",
		"扶抑喜用=水金",
		"当前路况：",
		"财富结构等级表示原局对钱财资源的显露、承载、流通和守成能力",
		"不是现实资产、收入、社会阶层或投资建议",
		"财富结构：",
		"财富结构依据：",
		"财富表达边界",
		"禁止写保证发财、资产规模、收益时点或投资建议",
		"不得重新打分、改判等级",
		"禁止写必富、必败、注定、阶层高低",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("expected prompt to contain %q", want)
		}
	}
	if result.VehicleProfile == nil || result.VehicleProfile.Grade == "" {
		t.Fatalf("test fixture should contain vehicle profile: %+v", result.VehicleProfile)
	}
	if result.WealthProfile == nil || result.WealthProfile.Grade == "" {
		t.Fatalf("test fixture should contain wealth profile: %+v", result.WealthProfile)
	}
	if !strings.Contains(prompt, result.VehicleProfile.VehicleType) {
		t.Fatalf("expected prompt to include vehicle type %q", result.VehicleProfile.VehicleType)
	}
	if !strings.Contains(prompt, result.WealthProfile.WealthType) {
		t.Fatalf("expected prompt to include wealth type %q", result.WealthProfile.WealthType)
	}
}

func TestFormatDayunNatalAssessmentSeparatesHighFoundationAndPrimaryPattern(t *testing.T) {
	result := bazi.Calculate(1995, 10, 12, 12, "male", false, 0, "solar", false)
	context := formatDayunNatalAssessment(result)
	for _, want := range []string{
		"日干调候天透地藏成格，为高格基础",
		"所需甲、壬",
		"透甲",
		"藏甲、壬",
		"主格结构偏印格/partial",
		"扶抑喜用水金",
		"不替代主格结构、制化或扶抑结论",
	} {
		if !strings.Contains(context, want) {
			t.Fatalf("expected dayun natal context to contain %q, got %s", want, context)
		}
	}
}

func TestNatalPromptIncludesSharedPriorityYongshen(t *testing.T) {
	result := bazi.Calculate(1996, 2, 8, 20, "male", false, 0, "solar", false)
	context := formatDayunNatalAssessment(result)
	if !strings.Contains(context, "扶抑喜用金土火") || !strings.Contains(context, "共同优先用神：火（寒热调候 + 扶抑）") {
		t.Fatalf("expected complete Fuyi and shared fire priority in dayun context, got %s", context)
	}
	prompt := buildBaziPrompt(result)
	if !strings.Contains(prompt, "共同优先用神：火（寒热调候 + 扶抑）") {
		t.Fatalf("expected report prompt to include shared fire priority, got %s", prompt)
	}
}

func TestBuildBaziPromptOmitsVehicleRoadContextWhenMissing(t *testing.T) {
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
		YearHideGan:     []string{"癸"},
		MonthHideGan:    []string{"甲", "丙", "戊"},
		DayHideGan:      []string{"戊", "乙", "癸"},
		HourHideGan:     []string{"丁", "己"},
		Wuxing:          bazi.WuxingStats{Mu: 2, Huo: 2, Tu: 2, Jin: 1, Shui: 1},
		Yongshen:        "火土",
		Jishen:          "水木",
		Gender:          "male",
		Dayun:           []bazi.DayunItem{{Index: 1, Gan: "辛", Zhi: "卯", StartAge: 3, StartYear: 2000, EndYear: 2009, GanShiShen: "伤官", ZhiShiShen: "正官", DiShi: "沐浴"}},
	}

	prompt := buildBaziPrompt(result)
	if strings.Contains(prompt, "[命盘座驾与大运路况-算法精算]") {
		t.Fatalf("prompt should omit vehicle-road context when structured fields are missing")
	}
	if strings.Contains(prompt, "重新生成座驾") || strings.Contains(prompt, "自行判断座驾") {
		t.Fatalf("prompt should not ask AI to invent missing vehicle-road labels")
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

func TestDayunSummaryPrompt_YearNarrativeRequiresAIRewriteFromEvidence(t *testing.T) {
	prompt := dayunSummaryPromptTpl

	for _, want := range []string{
		"AI 润色参考",
		"fallback_narrative",
		"把命理依据翻译成普通用户能听懂的人话",
		"不得照抄 fallback_narrative",
		"不要照抄 evidence 原文",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("expected prompt to contain %q", want)
		}
	}
	for _, bad := range []string{
		"必须点名当年关键干支事件",
		"引用上方 evidence 已有的命理术语",
	} {
		if strings.Contains(prompt, bad) {
			t.Fatalf("prompt should no longer force technical citation %q", bad)
		}
	}
}

func TestDayunSummaryPrompt_YearNarrativeMustLeadWithPlainLanguage(t *testing.T) {
	prompt := dayunSummaryPromptTpl

	for _, want := range []string{
		"第一句必须先写现实场景",
		"禁止以「流年」「伏吟」「反吟」「三合」「三会」「六合」「天克地冲」",
		"先讲用户能感知到的现实变化，再补一句命理依据",
		"不要写成命理术语清单",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("expected prompt to enforce plain-language lead %q", want)
		}
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

func TestFillBlankYearNarratives_GenericGeneratedTextUsesEvidenceAlignedFallback(t *testing.T) {
	parsed := []parsedYearAI{
		{Year: 2024, GanZhi: "甲辰", Narrative: "甲辰年整体会有变化，感情和生活节奏需要稳一点。"},
	}
	signals := []bazi.YearSignals{
		{
			Year:   2024,
			Age:    32,
			GanZhi: "甲辰",
			Signals: []bazi.EventSignal{
				{
					Type:     "婚恋_冲",
					Evidence: "流年地支辰冲日支戌，感情关系、居住状态或合作边界受触动",
					Polarity: bazi.PolarityXiong,
					Source:   bazi.SourceZhuwei,
				},
			},
		},
	}

	out := fillBlankYearNarratives(parsed, signals, 1)
	if len(out) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(out))
	}
	if out[0].Narrative == parsed[0].Narrative {
		t.Fatalf("generic generated text should not override evidence-aligned fallback: %q", out[0].Narrative)
	}
	for _, want := range []string{"波动", "亲密关系"} {
		if !strings.Contains(out[0].Narrative, want) {
			t.Fatalf("expected fallback narrative to explain %q, got: %s", want, out[0].Narrative)
		}
	}
	if strings.Contains(out[0].Narrative, "从命理依据看") {
		t.Fatalf("fallback narrative should be readable body text, got: %s", out[0].Narrative)
	}
}

func TestFillBlankYearNarratives_HumanizedAINarrativePreserved(t *testing.T) {
	parsed := []parsedYearAI{
		{Year: 1997, GanZhi: "丁丑", Narrative: "这一年同学朋友的影响会更明显，身边有人愿意帮忙，但也容易因为比较、资源分配或意见不同产生小摩擦。遇到计划变化时，先确认对方承诺和具体安排，不要只凭一句话就当成定局。长辈或老师的提醒可以帮你缓和局面，适合把注意力放在学习节奏和日常关系的稳定上。"},
	}
	signals := []bazi.YearSignals{
		{
			Year:   1997,
			Age:    3,
			GanZhi: "丁丑",
			Signals: []bazi.EventSignal{
				{
					Type:     bazi.TypeXueYeJingZheng,
					Evidence: "丁透干为劫财，少年期间同学竞争 / 友谊摩擦显著，宜以平常心相处",
					Polarity: bazi.PolarityNeutral,
					Source:   bazi.SourceZhuwei,
				},
				{
					Type:     "夹拱",
					Evidence: "原局年柱与日柱天干相同，地支隔位夹拱，主意料之外的人相助",
					Polarity: bazi.PolarityNeutral,
					Source:   bazi.SourceGongJia,
				},
				{
					Type:     "综合变动",
					Evidence: "流年地支丑落日柱旬空，事件虚而不实/过而不留",
					Polarity: bazi.PolarityNeutral,
					Source:   bazi.SourceKongwang,
				},
			},
		},
	}

	out := fillBlankYearNarratives(parsed, signals, 1)
	if len(out) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(out))
	}
	if out[0].Narrative != parsed[0].Narrative {
		t.Fatalf("humanized AI narrative should be preserved, got: %s", out[0].Narrative)
	}
}

func TestRenderPastEventsStageOneNarrative_AIModeStillReturnsEvidenceFallback(t *testing.T) {
	ys := bazi.YearSignals{
		Year:   2024,
		Age:    32,
		GanZhi: "甲辰",
		Signals: []bazi.EventSignal{
			{
				Type:     "婚恋_冲",
				Evidence: "流年地支辰冲日支戌，感情关系、居住状态或合作边界受触动",
				Polarity: bazi.PolarityXiong,
				Source:   bazi.SourceZhuwei,
			},
		},
	}

	got := renderPastEventsStageOneNarrative("ai", ys)
	if got == "" {
		t.Fatal("ai mode should still expose deterministic evidence-aligned fallback narrative")
	}
	for _, want := range []string{"亲密关系", "居住状态", "合作关系"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected stage-one fallback to include %q, got: %s", want, got)
		}
	}
	if strings.Contains(got, "从命理依据看") {
		t.Fatalf("stage-one fallback should not use evidence-section wording: %s", got)
	}
}

func TestBuildPastEventsExportDataFiltersToGeneratedCachedSegments(t *testing.T) {
	validYears := json.RawMessage(`[
		{"year":2000,"ganzhi":"庚辰","narrative":"这一年学习和关系节奏更容易被外界推动，适合稳住安排。"},
		{"year":2001,"ganzhi":"辛巳","narrative":""}
	]`)
	themes := json.RawMessage(`["学业突破","同伴助力"]`)
	emptyYears := json.RawMessage(`[]`)
	invalidYears := json.RawMessage(`{"bad":true}`)
	chart := &model.BaziChart{
		ID: "chart-1", DisplayName: "测试命盘",
		BirthYear: 1990, BirthMonth: 1, BirthDay: 2, BirthHour: 12,
		Gender:  "male",
		YearGan: "庚", YearZhi: "午", MonthGan: "戊", MonthZhi: "寅",
		DayGan: "甲", DayZhi: "子", HourGan: "庚", HourZhi: "午",
	}
	result := &bazi.BaziResult{
		YearGan: "庚", YearZhi: "午", MonthGan: "戊", MonthZhi: "寅",
		DayGan: "甲", DayZhi: "子", HourGan: "庚", HourZhi: "午",
		Gender: "male", BirthYear: 1990, BirthMonth: 1, BirthDay: 2, BirthHour: 12,
		Dayun: []bazi.DayunItem{
			{
				Index: 2, Gan: "乙", Zhi: "卯", StartAge: 10, StartYear: 2000, EndYear: 2009,
				GanShiShen: "劫财", ZhiShiShen: "劫财",
				LiuNian: []bazi.LiuNianItem{
					{Year: 2000, Age: 10, GanZhi: "庚辰", GanShiShen: "七杀", ZhiShiShen: "偏财"},
					{Year: 2001, Age: 11, GanZhi: "辛巳", GanShiShen: "正官", ZhiShiShen: "食神"},
				},
			},
			{
				Index: 1, Gan: "甲", Zhi: "寅", StartAge: 0, StartYear: 1990, EndYear: 1999,
				GanShiShen: "比肩", ZhiShiShen: "比肩",
			},
		},
	}
	cached := []model.AIDayunSummary{
		{DayunIndex: 9, DayunGanZhi: "癸亥", Summary: "没有年份数据", Years: &emptyYears},
		{DayunIndex: 2, DayunGanZhi: "乙卯", Themes: &themes, Summary: "这步大运已经生成。", Years: &validYears, Model: "test-model", CreatedAt: time.Unix(100, 0)},
		{DayunIndex: 3, DayunGanZhi: "丙辰", Summary: "坏 JSON", Years: &invalidYears},
		{DayunIndex: 4, DayunGanZhi: "丁巳", Summary: "", Years: &validYears},
		{DayunIndex: 5, DayunGanZhi: "戊午", Summary: "无 years"},
	}

	got, err := BuildPastEventsExportData(chart, result, cached)
	if err != nil {
		t.Fatalf("BuildPastEventsExportData returned error: %v", err)
	}
	if got.Chart.ID != "chart-1" || got.Chart.DisplayName != "测试命盘" {
		t.Fatalf("unexpected chart context: %+v", got.Chart)
	}
	if len(got.Segments) != 1 {
		t.Fatalf("expected only one exportable segment, got %d: %+v", len(got.Segments), got.Segments)
	}
	seg := got.Segments[0]
	if seg.DayunIndex != 2 || seg.StartAge != 10 || seg.EndAge != 19 || seg.StartYear != 2000 || seg.EndYear != 2009 {
		t.Fatalf("unexpected segment metadata: %+v", seg)
	}
	if len(seg.Themes) != 2 || seg.Themes[0] != "学业突破" {
		t.Fatalf("unexpected themes: %+v", seg.Themes)
	}
	if len(seg.Years) != 1 {
		t.Fatalf("expected only non-empty generated year narrative, got %+v", seg.Years)
	}
	if seg.Years[0].Year != 2000 || seg.Years[0].Age != 10 || seg.Years[0].GanZhi != "庚辰" {
		t.Fatalf("unexpected year metadata: %+v", seg.Years[0])
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
