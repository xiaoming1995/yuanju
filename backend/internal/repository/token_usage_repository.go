package repository

import (
	"fmt"
	"log"
	"sort"
	"time"
	"yuanju/pkg/database"
)

type TokenUsageSummaryRow struct {
	UserID           string  `json:"user_id"`
	Email            string  `json:"email"`
	Nickname         string  `json:"nickname"`
	Model            string  `json:"model"`
	RequestCount     int     `json:"request_count"`
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	TotalTokens      int     `json:"total_tokens"`
	EstimatedCostCny float64 `json:"estimated_cost_cny"`
}

type TokenUsageDetailRow struct {
	ID               string    `json:"id"`
	CallType         string    `json:"call_type"`
	Model            string    `json:"model"`
	PromptTokens     int       `json:"prompt_tokens"`
	CompletionTokens int       `json:"completion_tokens"`
	TotalTokens      int       `json:"total_tokens"`
	ReasoningTokens  int       `json:"reasoning_tokens"`
	CacheHitTokens   int       `json:"cache_hit_tokens"`
	CacheMissTokens  int       `json:"cache_miss_tokens"`
	EstimatedCostCny float64   `json:"estimated_cost_cny"`
	CreatedAt        time.Time `json:"created_at"`
}

// CreateTokenUsageLog 写入一条 token 用量记录；totalTokens==0 时跳过
func CreateTokenUsageLog(userID *string, chartID *string, callType, model, providerID string,
	promptTokens, completionTokens, totalTokens, reasoningTokens, cacheHitTokens, cacheMissTokens int,
	inputContent, outputContent string) error {
	log.Printf("[TokenUsage] 写入调用: callType=%s userID=%v prompt=%d completion=%d total=%d reasoning=%d cacheHit=%d",
		callType, userID, promptTokens, completionTokens, totalTokens, reasoningTokens, cacheHitTokens)
	if totalTokens == 0 {
		log.Printf("[TokenUsage] total=0，跳过写入")
		return nil
	}
	var providerIDPtr *string
	if providerID != "" {
		providerIDPtr = &providerID
	}
	_, err := database.DB.Exec(`
		INSERT INTO token_usage_logs
			(user_id, chart_id, call_type, model, provider_id,
			 prompt_tokens, completion_tokens, total_tokens,
			 reasoning_tokens, cache_hit_tokens, cache_miss_tokens,
			 input_content, output_content)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		userID, chartID, callType, model, providerIDPtr,
		promptTokens, completionTokens, totalTokens,
		reasoningTokens, cacheHitTokens, cacheMissTokens,
		inputContent, outputContent,
	)
	if err != nil {
		return fmt.Errorf("CreateTokenUsageLog: %w", err)
	}
	return nil
}

// GetTokenUsageContent 按 id 查询单条调用的输入/输出内容
func GetTokenUsageContent(id string) (inputContent, outputContent string, err error) {
	err = database.DB.QueryRow(`
		SELECT COALESCE(input_content, ''), COALESCE(output_content, '')
		FROM token_usage_logs WHERE id = $1`, id,
	).Scan(&inputContent, &outputContent)
	if err != nil {
		return "", "", fmt.Errorf("GetTokenUsageContent: %w", err)
	}
	return inputContent, outputContent, nil
}

// GetTokenUsageSummary 按用户聚合 token 消耗，from/to 均为日期（含）。
// costFn(model, promptTokens, completionTokens) 用于计算预估费用；传 nil 则费用为 0。
func GetTokenUsageSummary(from, to time.Time, costFn func(string, int, int) float64) ([]TokenUsageSummaryRow, error) {
	toExcl := to.AddDate(0, 0, 1)
	rows, err := database.DB.Query(`
		SELECT
			u.id,
			u.email,
			COALESCE(u.nickname, '') AS nickname,
			COUNT(t.id)::int                           AS request_count,
			COALESCE(t.model, '')                      AS model,
			COALESCE(SUM(t.prompt_tokens), 0)::int     AS prompt_tokens,
			COALESCE(SUM(t.completion_tokens), 0)::int AS completion_tokens,
			COALESCE(SUM(t.total_tokens), 0)::int      AS total_tokens
		FROM users u
		JOIN token_usage_logs t ON t.user_id = u.id
		WHERE t.created_at >= $1 AND t.created_at < $2
		GROUP BY u.id, u.email, u.nickname, t.model
		ORDER BY u.id`,
		from, toExcl,
	)
	if err != nil {
		return nil, fmt.Errorf("GetTokenUsageSummary: %w", err)
	}
	defer rows.Close()

	var byModel []TokenUsageSummaryRow
	for rows.Next() {
		var r TokenUsageSummaryRow
		if err := rows.Scan(&r.UserID, &r.Email, &r.Nickname, &r.RequestCount,
			&r.Model, &r.PromptTokens, &r.CompletionTokens, &r.TotalTokens); err != nil {
			log.Printf("[TokenUsage] Scan 失败: %v", err)
			continue
		}
		if costFn != nil {
			r.EstimatedCostCny = costFn(r.Model, r.PromptTokens, r.CompletionTokens)
		}
		byModel = append(byModel, r)
	}

	type userMeta struct {
		totalTokens int
		rows        []TokenUsageSummaryRow
	}
	userMap := make(map[string]*userMeta)
	var userOrder []string
	for _, r := range byModel {
		if _, exists := userMap[r.UserID]; !exists {
			userMap[r.UserID] = &userMeta{}
			userOrder = append(userOrder, r.UserID)
		}
		userMap[r.UserID].totalTokens += r.TotalTokens
		userMap[r.UserID].rows = append(userMap[r.UserID].rows, r)
	}

	sort.Slice(userOrder, func(i, j int) bool {
		return userMap[userOrder[i]].totalTokens > userMap[userOrder[j]].totalTokens
	})

	var result []TokenUsageSummaryRow
	for _, uid := range userOrder {
		meta := userMap[uid]
		sort.Slice(meta.rows, func(i, j int) bool {
			return meta.rows[i].TotalTokens > meta.rows[j].TotalTokens
		})
		result = append(result, meta.rows...)
	}
	return result, nil
}

// GetTokenUsageDetail 查询单用户分页明细。
// model 传空字符串则不过滤模型；costFn 传 nil 则费用为 0。
func GetTokenUsageDetail(userID string, from, to time.Time, page, limit int, model string, costFn func(string, int, int) float64) (total int, items []TokenUsageDetailRow, err error) {
	toExcl := to.AddDate(0, 0, 1)
	offset := (page - 1) * limit

	if err = database.DB.QueryRow(`
		SELECT COUNT(*) FROM token_usage_logs
		WHERE user_id = $1 AND created_at >= $2 AND created_at < $3
		AND ($4 = '' OR COALESCE(model, '') = $4)`,
		userID, from, toExcl, model,
	).Scan(&total); err != nil {
		return 0, nil, fmt.Errorf("GetTokenUsageDetail count: %w", err)
	}

	rows, err := database.DB.Query(`
		SELECT id, call_type, COALESCE(model, ''), prompt_tokens, completion_tokens, total_tokens,
		       reasoning_tokens, cache_hit_tokens, cache_miss_tokens, created_at
		FROM token_usage_logs
		WHERE user_id = $1 AND created_at >= $2 AND created_at < $3
		AND ($4 = '' OR COALESCE(model, '') = $4)
		ORDER BY created_at DESC
		LIMIT $5 OFFSET $6`,
		userID, from, toExcl, model, limit, offset,
	)
	if err != nil {
		return 0, nil, fmt.Errorf("GetTokenUsageDetail query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var r TokenUsageDetailRow
		if err := rows.Scan(&r.ID, &r.CallType, &r.Model,
			&r.PromptTokens, &r.CompletionTokens, &r.TotalTokens,
			&r.ReasoningTokens, &r.CacheHitTokens, &r.CacheMissTokens, &r.CreatedAt); err != nil {
			log.Printf("[TokenUsage] Scan detail 失败: %v", err)
			continue
		}
		if costFn != nil {
			r.EstimatedCostCny = costFn(r.Model, r.PromptTokens, r.CompletionTokens)
		}
		items = append(items, r)
	}
	return total, items, nil
}
