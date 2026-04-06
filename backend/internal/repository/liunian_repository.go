package repository

import (
	"database/sql"
	"encoding/json"
	"yuanju/internal/model"
	"yuanju/pkg/database"
)

// GetLiunianReport 获取某一流年报告（缓存）
func GetLiunianReport(chartID string, targetYear int) (*model.AILiunianReport, error) {
	report := &model.AILiunianReport{}
	err := database.DB.QueryRow(
		`SELECT id, chart_id, target_year, dayun_ganzhi, content_structured, model, created_at 
		 FROM ai_liunian_reports 
		 WHERE chart_id = $1 AND target_year = $2`,
		chartID, targetYear,
	).Scan(&report.ID, &report.ChartID, &report.TargetYear, &report.DayunGanzhi, &report.ContentStructured, &report.Model, &report.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return report, err
}

// CreateLiunianReport 存入新的流年报告
func CreateLiunianReport(chartID string, targetYear int, dayunGanzhi string, contentStructured *json.RawMessage, modelName string) (*model.AILiunianReport, error) {
	report := &model.AILiunianReport{}
	err := database.DB.QueryRow(
		`INSERT INTO ai_liunian_reports (chart_id, target_year, dayun_ganzhi, content_structured, model)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (chart_id, target_year) DO UPDATE 
		 SET content_structured = EXCLUDED.content_structured, model = EXCLUDED.model
		 RETURNING id, chart_id, target_year, dayun_ganzhi, content_structured, model, created_at`,
		chartID, targetYear, dayunGanzhi, contentStructured, modelName,
	).Scan(&report.ID, &report.ChartID, &report.TargetYear, &report.DayunGanzhi, &report.ContentStructured, &report.Model, &report.CreatedAt)

	return report, err
}

// GetLiunianReportsByChartID 获取某排盘下所有的流年批断记录
func GetLiunianReportsByChartID(chartID string) ([]model.AILiunianReport, error) {
	rows, err := database.DB.Query(
		`SELECT id, chart_id, target_year, dayun_ganzhi, content_structured, model, created_at
		 FROM ai_liunian_reports
		 WHERE chart_id = $1
		 ORDER BY target_year ASC, created_at DESC`,
		chartID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []model.AILiunianReport
	for rows.Next() {
		var r model.AILiunianReport
		if err := rows.Scan(&r.ID, &r.ChartID, &r.TargetYear, &r.DayunGanzhi, &r.ContentStructured, &r.Model, &r.CreatedAt); err != nil {
			return nil, err
		}
		reports = append(reports, r)
	}
	return reports, nil
}

// DeleteLiunianReportByID 单独删除某一流年报告缓存
func DeleteLiunianReportByID(id string) error {
	_, err := database.DB.Exec(`DELETE FROM ai_liunian_reports WHERE id = $1`, id)
	return err
}
