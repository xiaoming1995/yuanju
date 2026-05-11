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
	RequestCount     int     `json:"request_count"`
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	TotalTokens      int     `json:"total_tokens"`
	EstimatedCostCny float64 `json:"estimated_cost_cny"`
}

type summaryByModel struct {
	userID, email, nickname             string
	model                               string
	requestCount                        int
	promptTokens, completionTokens, totalTokens int
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
	CreatedAt        time.Time `json:"created_at"`
}

// CreateTokenUsageLog 写入一条 token 用量记录；totalTokens==0 时跳过
func CreateTokenUsageLog(userID *string, chartID *string, callType, model, providerID string, promptTokens, completionTokens, totalTokens, reasoningTokens, cacheHitTokens, cacheMissTokens int) error {
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
			 reasoning_tokens, cache_hit_tokens, cache_miss_tokens)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		userID, chartID, callType, model, providerIDPtr,
		promptTokens, completionTokens, totalTokens,
		reasoningTokens, cacheHitTokens, cacheMissTokens,
	)
	if err != nil {
		return fmt.Errorf("CreateTokenUsageLog: %w", err)
	}
	return nil
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

	var byModel []summaryByModel
	for rows.Next() {
		var r summaryByModel
		if err := rows.Scan(&r.userID, &r.email, &r.nickname, &r.requestCount,
			&r.model, &r.promptTokens, &r.completionTokens, &r.totalTokens); err != nil {
			log.Printf("[TokenUsage] Scan 失败: %v", err)
			continue
		}
		byModel = append(byModel, r)
	}

	type entry struct {
		row  TokenUsageSummaryRow
		cost float64
	}
	userMap := make(map[string]*entry)
	var userOrder []string

	for _, r := range byModel {
		e, exists := userMap[r.userID]
		if !exists {
			e = &entry{row: TokenUsageSummaryRow{
				UserID:   r.userID,
				Email:    r.email,
				Nickname: r.nickname,
			}}
			userMap[r.userID] = e
			userOrder = append(userOrder, r.userID)
		}
		e.row.RequestCount += r.requestCount
		e.row.PromptTokens += r.promptTokens
		e.row.CompletionTokens += r.completionTokens
		e.row.TotalTokens += r.totalTokens
		if costFn != nil {
			e.cost += costFn(r.model, r.promptTokens, r.completionTokens)
		}
	}

	result := make([]TokenUsageSummaryRow, 0, len(userOrder))
	for _, uid := range userOrder {
		e := userMap[uid]
		e.row.EstimatedCostCny = e.cost
		result = append(result, e.row)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].TotalTokens > result[j].TotalTokens
	})

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
		SELECT id, call_type, COALESCE(model, ''), prompt_tokens, completion_tokens, total_tokens,
		       reasoning_tokens, cache_hit_tokens, cache_miss_tokens, created_at
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
			&r.PromptTokens, &r.CompletionTokens, &r.TotalTokens,
			&r.ReasoningTokens, &r.CacheHitTokens, &r.CacheMissTokens, &r.CreatedAt); err != nil {
			log.Printf("[TokenUsage] Scan detail 失败: %v", err)
			continue
		}
		items = append(items, r)
	}
	return total, items, nil
}
