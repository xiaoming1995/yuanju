package repository

import (
	"database/sql"
	"encoding/json"
	"yuanju/internal/model"
	"yuanju/pkg/database"
)

func GetPastEvents(chartID string) (*model.AIPastEvents, error) {
	r := &model.AIPastEvents{}
	err := database.DB.QueryRow(
		`SELECT id, chart_id, content_structured, model, created_at FROM ai_past_events WHERE chart_id = $1`,
		chartID,
	).Scan(&r.ID, &r.ChartID, &r.ContentStructured, &r.Model, &r.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return r, err
}

func CreatePastEvents(chartID string, contentStructured *json.RawMessage, modelName string) (*model.AIPastEvents, error) {
	r := &model.AIPastEvents{}
	err := database.DB.QueryRow(
		`INSERT INTO ai_past_events (chart_id, content_structured, model)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (chart_id) DO UPDATE
		 SET content_structured = EXCLUDED.content_structured, model = EXCLUDED.model, created_at = NOW()
		 RETURNING id, chart_id, content_structured, model, created_at`,
		chartID, contentStructured, modelName,
	).Scan(&r.ID, &r.ChartID, &r.ContentStructured, &r.Model, &r.CreatedAt)
	return r, err
}

func DeletePastEvents(chartID string) error {
	_, err := database.DB.Exec(`DELETE FROM ai_past_events WHERE chart_id = $1`, chartID)
	return err
}
