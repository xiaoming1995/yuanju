package repository

import (
	"fmt"
	"log"
	"time"
	"yuanju/pkg/database"
)

type TokenUsageSummaryRow struct {
	UserID           string `json:"user_id"`
	Email            string `json:"email"`
	Nickname         string `json:"nickname"`
	RequestCount     int    `json:"request_count"`
	PromptTokens     int    `json:"prompt_tokens"`
	CompletionTokens int    `json:"completion_tokens"`
	TotalTokens      int    `json:"total_tokens"`
}

type TokenUsageDetailRow struct {
	ID               string    `json:"id"`
	CallType         string    `json:"call_type"`
	Model            string    `json:"model"`
	PromptTokens     int       `json:"prompt_tokens"`
	CompletionTokens int       `json:"completion_tokens"`
	TotalTokens      int       `json:"total_tokens"`
	CreatedAt        time.Time `json:"created_at"`
}

// CreateTokenUsageLog 写入一条 token 用量记录；totalTokens==0 时跳过
func CreateTokenUsageLog(userID *string, chartID *string, callType, model, providerID string, promptTokens, completionTokens, totalTokens int) error {
	if totalTokens == 0 {
		return nil
	}
	var providerIDPtr *string
	if providerID != "" {
		providerIDPtr = &providerID
	}
	_, err := database.DB.Exec(`
		INSERT INTO token_usage_logs
			(user_id, chart_id, call_type, model, provider_id, prompt_tokens, completion_tokens, total_tokens)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		userID, chartID, callType, model, providerIDPtr,
		promptTokens, completionTokens, totalTokens,
	)
	if err != nil {
		return fmt.Errorf("CreateTokenUsageLog: %w", err)
	}
	return nil
}

// GetTokenUsageSummary 按用户聚合 token 消耗，from/to 均为日期（含）
func GetTokenUsageSummary(from, to time.Time) ([]TokenUsageSummaryRow, error) {
	toExcl := to.AddDate(0, 0, 1)
	rows, err := database.DB.Query(`
		SELECT
			u.id,
			u.email,
			COALESCE(u.nickname, '') AS nickname,
			COUNT(t.id)::int                           AS request_count,
			COALESCE(SUM(t.prompt_tokens), 0)::int     AS prompt_tokens,
			COALESCE(SUM(t.completion_tokens), 0)::int AS completion_tokens,
			COALESCE(SUM(t.total_tokens), 0)::int      AS total_tokens
		FROM users u
		JOIN token_usage_logs t ON t.user_id = u.id
		WHERE t.created_at >= $1 AND t.created_at < $2
		GROUP BY u.id, u.email, u.nickname
		ORDER BY total_tokens DESC`,
		from, toExcl,
	)
	if err != nil {
		return nil, fmt.Errorf("GetTokenUsageSummary: %w", err)
	}
	defer rows.Close()

	var result []TokenUsageSummaryRow
	for rows.Next() {
		var r TokenUsageSummaryRow
		if err := rows.Scan(&r.UserID, &r.Email, &r.Nickname, &r.RequestCount,
			&r.PromptTokens, &r.CompletionTokens, &r.TotalTokens); err != nil {
			log.Printf("[TokenUsage] Scan 失败: %v", err)
			continue
		}
		result = append(result, r)
	}
	return result, nil
}

// GetTokenUsageDetail 查询单用户分页明细
func GetTokenUsageDetail(userID string, from, to time.Time, page, limit int) (total int, items []TokenUsageDetailRow, err error) {
	toExcl := to.AddDate(0, 0, 1)
	offset := (page - 1) * limit

	if err = database.DB.QueryRow(`
		SELECT COUNT(*) FROM token_usage_logs
		WHERE user_id = $1 AND created_at >= $2 AND created_at < $3`,
		userID, from, toExcl,
	).Scan(&total); err != nil {
		return 0, nil, fmt.Errorf("GetTokenUsageDetail count: %w", err)
	}

	rows, err := database.DB.Query(`
		SELECT id, call_type, COALESCE(model, ''), prompt_tokens, completion_tokens, total_tokens, created_at
		FROM token_usage_logs
		WHERE user_id = $1 AND created_at >= $2 AND created_at < $3
		ORDER BY created_at DESC
		LIMIT $4 OFFSET $5`,
		userID, from, toExcl, limit, offset,
	)
	if err != nil {
		return 0, nil, fmt.Errorf("GetTokenUsageDetail query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var r TokenUsageDetailRow
		if err := rows.Scan(&r.ID, &r.CallType, &r.Model,
			&r.PromptTokens, &r.CompletionTokens, &r.TotalTokens, &r.CreatedAt); err != nil {
			log.Printf("[TokenUsage] Scan detail 失败: %v", err)
			continue
		}
		items = append(items, r)
	}
	return total, items, nil
}
