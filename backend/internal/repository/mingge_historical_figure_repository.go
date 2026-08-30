package repository

import (
	"fmt"
	"net/url"
	"strings"
	"yuanju/internal/model"
	"yuanju/pkg/database"
)

const historicalFigureColumns = `
	id, ming_ge, figure_name, era, identity, historical_memory,
	turning_point, turning_point_year, source_title, source_url,
	birth_data_precision, bazi_verification_note, dayun_period,
	dayun_explanation, show_dayun, display_order, review_status,
	created_at, updated_at`

func normalizeHistoricalFigure(figure *model.MingGeHistoricalFigure) {
	figure.MingGe = strings.TrimSpace(figure.MingGe)
	figure.FigureName = strings.TrimSpace(figure.FigureName)
	figure.Era = strings.TrimSpace(figure.Era)
	figure.Identity = strings.TrimSpace(figure.Identity)
	figure.HistoricalMemory = strings.TrimSpace(figure.HistoricalMemory)
	figure.TurningPoint = strings.TrimSpace(figure.TurningPoint)
	figure.TurningPointYear = strings.TrimSpace(figure.TurningPointYear)
	figure.SourceTitle = strings.TrimSpace(figure.SourceTitle)
	figure.SourceURL = strings.TrimSpace(figure.SourceURL)
	figure.BirthDataPrecision = strings.TrimSpace(figure.BirthDataPrecision)
	figure.BaziVerificationNote = strings.TrimSpace(figure.BaziVerificationNote)
	figure.DayunPeriod = strings.TrimSpace(figure.DayunPeriod)
	figure.DayunExplanation = strings.TrimSpace(figure.DayunExplanation)
	figure.ReviewStatus = strings.TrimSpace(figure.ReviewStatus)
	if figure.BirthDataPrecision == "" {
		figure.BirthDataPrecision = "unknown"
	}
	if figure.ReviewStatus == "" {
		figure.ReviewStatus = "draft"
	}
}

// ValidateMingGeHistoricalFigure ensures public historical content has editorial provenance.
func ValidateMingGeHistoricalFigure(figure *model.MingGeHistoricalFigure) error {
	if figure == nil {
		return fmt.Errorf("古人映照内容不能为空")
	}
	normalizeHistoricalFigure(figure)
	for label, value := range map[string]string{
		"命格": figure.MingGe, "人物": figure.FigureName, "时代": figure.Era,
		"身份": figure.Identity, "历史印记": figure.HistoricalMemory,
		"资料标题": figure.SourceTitle, "资料链接": figure.SourceURL,
	} {
		if value == "" {
			return fmt.Errorf("%s不能为空", label)
		}
	}
	parsed, err := url.ParseRequestURI(figure.SourceURL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") {
		return fmt.Errorf("资料链接必须是有效的 http 或 https 地址")
	}
	switch figure.BirthDataPrecision {
	case "unknown", "date_only", "exact_hour":
	default:
		return fmt.Errorf("出生资料精度必须为 unknown、date_only 或 exact_hour")
	}
	switch figure.ReviewStatus {
	case "draft", "reviewed", "published", "archived":
	default:
		return fmt.Errorf("审核状态必须为 draft、reviewed、published 或 archived")
	}
	if figure.ShowDayun {
		if figure.BirthDataPrecision != "exact_hour" || figure.BaziVerificationNote == "" ||
			figure.TurningPoint == "" || figure.TurningPointYear == "" ||
			figure.DayunPeriod == "" || figure.DayunExplanation == "" {
			return fmt.Errorf("展示大运需提供精确时辰、命盘核验、人生转折、转折年份、大运阶段与呼应说明")
		}
	}
	return nil
}

func scanHistoricalFigure(scanner interface{ Scan(...any) error }) (*model.MingGeHistoricalFigure, error) {
	figure := &model.MingGeHistoricalFigure{}
	err := scanner.Scan(
		&figure.ID, &figure.MingGe, &figure.FigureName, &figure.Era, &figure.Identity, &figure.HistoricalMemory,
		&figure.TurningPoint, &figure.TurningPointYear, &figure.SourceTitle, &figure.SourceURL,
		&figure.BirthDataPrecision, &figure.BaziVerificationNote, &figure.DayunPeriod,
		&figure.DayunExplanation, &figure.ShowDayun, &figure.DisplayOrder, &figure.ReviewStatus,
		&figure.CreatedAt, &figure.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return figure, nil
}

func CreateMingGeHistoricalFigure(figure model.MingGeHistoricalFigure) (*model.MingGeHistoricalFigure, error) {
	if err := ValidateMingGeHistoricalFigure(&figure); err != nil {
		return nil, err
	}
	query := `INSERT INTO mingge_historical_figures (
		ming_ge, figure_name, era, identity, historical_memory, turning_point, turning_point_year,
		source_title, source_url, birth_data_precision, bazi_verification_note, dayun_period,
		dayun_explanation, show_dayun, display_order, review_status
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
	RETURNING ` + historicalFigureColumns
	return scanHistoricalFigure(database.DB.QueryRow(query,
		figure.MingGe, figure.FigureName, figure.Era, figure.Identity, figure.HistoricalMemory,
		figure.TurningPoint, figure.TurningPointYear, figure.SourceTitle, figure.SourceURL,
		figure.BirthDataPrecision, figure.BaziVerificationNote, figure.DayunPeriod, figure.DayunExplanation,
		figure.ShowDayun, figure.DisplayOrder, figure.ReviewStatus,
	))
}

func UpdateMingGeHistoricalFigure(id string, figure model.MingGeHistoricalFigure) (*model.MingGeHistoricalFigure, error) {
	if err := ValidateMingGeHistoricalFigure(&figure); err != nil {
		return nil, err
	}
	query := `UPDATE mingge_historical_figures SET
		ming_ge=$1, figure_name=$2, era=$3, identity=$4, historical_memory=$5,
		turning_point=$6, turning_point_year=$7, source_title=$8, source_url=$9,
		birth_data_precision=$10, bazi_verification_note=$11, dayun_period=$12,
		dayun_explanation=$13, show_dayun=$14, display_order=$15, review_status=$16,
		updated_at=NOW()
		WHERE id=$17 RETURNING ` + historicalFigureColumns
	return scanHistoricalFigure(database.DB.QueryRow(query,
		figure.MingGe, figure.FigureName, figure.Era, figure.Identity, figure.HistoricalMemory,
		figure.TurningPoint, figure.TurningPointYear, figure.SourceTitle, figure.SourceURL,
		figure.BirthDataPrecision, figure.BaziVerificationNote, figure.DayunPeriod, figure.DayunExplanation,
		figure.ShowDayun, figure.DisplayOrder, figure.ReviewStatus, id,
	))
}

func ListMingGeHistoricalFigures() ([]model.MingGeHistoricalFigure, error) {
	rows, err := database.DB.Query(`SELECT ` + historicalFigureColumns + ` FROM mingge_historical_figures ORDER BY ming_ge, display_order, created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectHistoricalFigures(rows)
}

func ListPublishedMingGeHistoricalFigures(mingGe string) ([]model.MingGeHistoricalFigure, error) {
	rows, err := database.DB.Query(`SELECT `+historicalFigureColumns+` FROM mingge_historical_figures
		WHERE ming_ge=$1 AND review_status='published'
		ORDER BY display_order, created_at DESC LIMIT 2`, strings.TrimSpace(mingGe))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	figures, err := collectHistoricalFigures(rows)
	if err != nil {
		return nil, err
	}
	for index := range figures {
		if !figures[index].ShowDayun {
			figures[index].DayunPeriod = ""
			figures[index].DayunExplanation = ""
			figures[index].BaziVerificationNote = ""
		}
	}
	return figures, nil
}

func collectHistoricalFigures(rows interface {
	Next() bool
	Scan(...any) error
	Err() error
}) ([]model.MingGeHistoricalFigure, error) {
	figures := make([]model.MingGeHistoricalFigure, 0)
	for rows.Next() {
		figure, err := scanHistoricalFigure(rows)
		if err != nil {
			return nil, err
		}
		figures = append(figures, *figure)
	}
	return figures, rows.Err()
}

func ArchiveMingGeHistoricalFigure(id string) error {
	_, err := database.DB.Exec(`UPDATE mingge_historical_figures SET review_status='archived', updated_at=NOW() WHERE id=$1`, id)
	return err
}
