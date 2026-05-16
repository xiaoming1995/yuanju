package repository

import (
	"database/sql"
	"encoding/json"

	"yuanju/internal/model"
	"yuanju/pkg/database"
)

// UpsertPolishedReport 创建或覆盖一条润色版报告。
// UNIQUE chart_id 约束 → 同一命盘只保留最新一份。
func UpsertPolishedReport(
	chartID, userSituation, content, modelName string,
	contentStructured *json.RawMessage,
	promptTokens, completionTokens, totalTokens int,
) (*model.PolishedReport, error) {
	report := &model.PolishedReport{}
	err := database.DB.QueryRow(
		`INSERT INTO ai_polished_reports
		   (chart_id, user_situation, content, content_structured, model,
		    prompt_tokens, completion_tokens, total_tokens, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		 ON CONFLICT (chart_id) DO UPDATE SET
		   user_situation = EXCLUDED.user_situation,
		   content        = EXCLUDED.content,
		   content_structured = EXCLUDED.content_structured,
		   model          = EXCLUDED.model,
		   prompt_tokens  = EXCLUDED.prompt_tokens,
		   completion_tokens = EXCLUDED.completion_tokens,
		   total_tokens   = EXCLUDED.total_tokens,
		   updated_at     = NOW()
		 RETURNING id, chart_id, user_situation, content, content_structured,
		           model, prompt_tokens, completion_tokens, total_tokens,
		           created_at, updated_at`,
		chartID, userSituation, content, contentStructured, modelName,
		promptTokens, completionTokens, totalTokens,
	).Scan(
		&report.ID, &report.ChartID, &report.UserSituation, &report.Content,
		&report.ContentStructured, &report.Model,
		&report.PromptTokens, &report.CompletionTokens, &report.TotalTokens,
		&report.CreatedAt, &report.UpdatedAt,
	)
	return report, err
}

// GetPolishedByChartID 读指定命盘的润色版。无记录返回 (nil, nil)。
func GetPolishedByChartID(chartID string) (*model.PolishedReport, error) {
	report := &model.PolishedReport{}
	err := database.DB.QueryRow(
		`SELECT id, chart_id, user_situation, content, content_structured,
		        model, prompt_tokens, completion_tokens, total_tokens,
		        created_at, updated_at
		 FROM ai_polished_reports WHERE chart_id = $1`, chartID,
	).Scan(
		&report.ID, &report.ChartID, &report.UserSituation, &report.Content,
		&report.ContentStructured, &report.Model,
		&report.PromptTokens, &report.CompletionTokens, &report.TotalTokens,
		&report.CreatedAt, &report.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return report, err
}
