package repository

import (
	"database/sql"
	"encoding/json"
	"yuanju/internal/model"
	"yuanju/pkg/database"
)

// GetDayunSummary 按 chart + dayun_index 查询缓存
func GetDayunSummary(chartID string, dayunIndex int) (*model.AIDayunSummary, error) {
	r := &model.AIDayunSummary{}
	err := database.DB.QueryRow(
		`SELECT id, chart_id, dayun_index, dayun_ganzhi, themes, summary, model, created_at
		 FROM ai_dayun_summaries
		 WHERE chart_id = $1 AND dayun_index = $2`,
		chartID, dayunIndex,
	).Scan(&r.ID, &r.ChartID, &r.DayunIndex, &r.DayunGanZhi, &r.Themes, &r.Summary, &r.Model, &r.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return r, err
}

// ListDayunSummaries 按 chart 拉取所有已缓存大运 summary
func ListDayunSummaries(chartID string) ([]model.AIDayunSummary, error) {
	rows, err := database.DB.Query(
		`SELECT id, chart_id, dayun_index, dayun_ganzhi, themes, summary, model, created_at
		 FROM ai_dayun_summaries
		 WHERE chart_id = $1
		 ORDER BY dayun_index`,
		chartID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.AIDayunSummary
	for rows.Next() {
		var r model.AIDayunSummary
		if err := rows.Scan(&r.ID, &r.ChartID, &r.DayunIndex, &r.DayunGanZhi, &r.Themes, &r.Summary, &r.Model, &r.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// UpsertDayunSummary 写入或覆盖单段缓存
func UpsertDayunSummary(chartID string, dayunIndex int, dayunGanZhi string, themes *json.RawMessage, summary string, modelName string) error {
	_, err := database.DB.Exec(
		`INSERT INTO ai_dayun_summaries (chart_id, dayun_index, dayun_ganzhi, themes, summary, model)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (chart_id, dayun_index) DO UPDATE
		 SET dayun_ganzhi = EXCLUDED.dayun_ganzhi,
		     themes = EXCLUDED.themes,
		     summary = EXCLUDED.summary,
		     model = EXCLUDED.model,
		     created_at = NOW()`,
		chartID, dayunIndex, dayunGanZhi, themes, summary, modelName,
	)
	return err
}

// DeleteDayunSummariesByChart 删除某 chart 的所有大运 summary 缓存
func DeleteDayunSummariesByChart(chartID string) error {
	_, err := database.DB.Exec(`DELETE FROM ai_dayun_summaries WHERE chart_id = $1`, chartID)
	return err
}
