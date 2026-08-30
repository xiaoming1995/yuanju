package repository

import (
	"testing"
	"yuanju/internal/model"
)

func completeHistoricalFigure() model.MingGeHistoricalFigure {
	return model.MingGeHistoricalFigure{
		MingGe: "伤官格", FigureName: "测试人物", Era: "测试时代", Identity: "测试身份",
		HistoricalMemory: "测试历史印记", SourceTitle: "测试资料", SourceURL: "https://example.com/source",
		BirthDataPrecision: "unknown", ReviewStatus: "draft",
	}
}

func TestValidateMingGeHistoricalFigureAcceptsReviewedEditorialContent(t *testing.T) {
	figure := completeHistoricalFigure()
	figure.ReviewStatus = "published"
	if err := ValidateMingGeHistoricalFigure(&figure); err != nil {
		t.Fatalf("expected valid editorial content, got %v", err)
	}
}

func TestValidateMingGeHistoricalFigureRejectsInvalidDayunDisplay(t *testing.T) {
	figure := completeHistoricalFigure()
	figure.ShowDayun = true
	if err := ValidateMingGeHistoricalFigure(&figure); err == nil {
		t.Fatal("expected incomplete Dayun display data to be rejected")
	}
}

func TestValidateMingGeHistoricalFigureAcceptsVerifiedDayunDisplay(t *testing.T) {
	figure := completeHistoricalFigure()
	figure.BirthDataPrecision = "exact_hour"
	figure.BaziVerificationNote = "出生时辰与四柱经编辑复核。"
	figure.TurningPoint = "完成代表性工作。"
	figure.TurningPointYear = "100 年"
	figure.DayunPeriod = "甲子大运"
	figure.DayunExplanation = "该阶段与其人生转折形成呼应，不作为单一因果。"
	figure.ShowDayun = true
	figure.ReviewStatus = "published"
	if err := ValidateMingGeHistoricalFigure(&figure); err != nil {
		t.Fatalf("expected verified Dayun display data to pass, got %v", err)
	}
}

func TestValidateMingGeHistoricalFigureRejectsUnsafeSourceURL(t *testing.T) {
	figure := completeHistoricalFigure()
	figure.SourceURL = "javascript:alert(1)"
	if err := ValidateMingGeHistoricalFigure(&figure); err == nil {
		t.Fatal("expected invalid source URL to be rejected")
	}
}

func TestListPublishedMingGeHistoricalFiguresFiltersOrdersAndHidesUnsupportedDayun(t *testing.T) {
	withArticleRepositoryDB(t)
	for _, input := range []model.MingGeHistoricalFigure{
		{MingGe: "测试格", FigureName: "草稿人物", Era: "测试时代", Identity: "测试身份", HistoricalMemory: "不应公开", SourceTitle: "测试资料", SourceURL: "https://example.com/draft", ReviewStatus: "draft", DisplayOrder: 0},
		{MingGe: "测试格", FigureName: "第二人物", Era: "测试时代", Identity: "测试身份", HistoricalMemory: "第二条公开资料", SourceTitle: "测试资料", SourceURL: "https://example.com/second", ReviewStatus: "published", DisplayOrder: 20},
		{MingGe: "测试格", FigureName: "第一人物", Era: "测试时代", Identity: "测试身份", HistoricalMemory: "第一条公开资料", SourceTitle: "测试资料", SourceURL: "https://example.com/first", BaziVerificationNote: "不应公开", DayunPeriod: "不应公开", DayunExplanation: "不应公开", ReviewStatus: "published", DisplayOrder: 10},
		{MingGe: "测试格", FigureName: "第三人物", Era: "测试时代", Identity: "测试身份", HistoricalMemory: "超出最大数量", SourceTitle: "测试资料", SourceURL: "https://example.com/third", ReviewStatus: "published", DisplayOrder: 30},
	} {
		if _, err := CreateMingGeHistoricalFigure(input); err != nil {
			t.Fatalf("create test figure %q: %v", input.FigureName, err)
		}
	}

	figures, err := ListPublishedMingGeHistoricalFigures("测试格")
	if err != nil {
		t.Fatalf("list published figures: %v", err)
	}
	if len(figures) != 2 || figures[0].FigureName != "第一人物" || figures[1].FigureName != "第二人物" {
		t.Fatalf("expected two ordered published figures, got %+v", figures)
	}
	if figures[0].DayunPeriod != "" || figures[0].DayunExplanation != "" || figures[0].BaziVerificationNote != "" {
		t.Fatalf("unverified Dayun fields must be hidden from public data: %+v", figures[0])
	}
	other, err := ListPublishedMingGeHistoricalFigures("不存在的格")
	if err != nil {
		t.Fatalf("list another Ming Ge: %v", err)
	}
	if len(other) != 0 {
		t.Fatalf("Ming Ge filter should not leak other references: %+v", other)
	}
}
