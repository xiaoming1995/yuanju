package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"
	"yuanju/pkg/database"
)

var ErrTodayQuestionDailyLimit = errors.New("today question daily limit reached")

// ReserveTodayQuestionRequest atomically checks the user's Beijing-day quota.
// The user-level advisory lock also serializes cross-day credit spending and grants.
func ReserveTodayQuestionRequest(userID, chartID, date, category, question string, contextData []byte) (string, int, error) {
	tx, err := database.DB.Begin()
	if err != nil {
		return "", 0, err
	}
	defer tx.Rollback()
	if err := lockTodayQuestionUserTx(tx, userID); err != nil {
		return "", 0, err
	}
	if err := releaseStaleTodayQuestionRequestsTx(tx, userID); err != nil {
		return "", 0, err
	}
	limit := 1
	var raw string
	err = tx.QueryRow(`SELECT value FROM system_settings WHERE key=$1`, SettingTodayQuestionDailyLimit).Scan(&raw)
	if err != nil && err != sql.ErrNoRows {
		return "", 0, err
	}
	if err == nil {
		limit, err = strconv.Atoi(raw)
		if err != nil || limit < 1 || limit > MaxTodayQuestionDailyLimit {
			return "", 0, fmt.Errorf("invalid today question daily limit %q", raw)
		}
	}
	var used int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM today_question_requests
		WHERE user_id=$1 AND question_date=$2 AND quota_charged AND quota_source='daily' AND status <> 'error'`, userID, date).Scan(&used); err != nil {
		return "", 0, err
	}
	source := "daily"
	creditBalance := 0
	if used >= limit {
		source = "credit"
		creditBalance, err = todayQuestionCreditBalanceTx(tx, userID)
		if err != nil {
			return "", 0, err
		}
		if creditBalance < 1 {
			return "", limit, ErrTodayQuestionDailyLimit
		}
	}
	var id string
	if err := tx.QueryRow(`INSERT INTO today_question_requests
		(user_id, chart_id, question_date, category, question, context_data, quota_charged, quota_source)
		VALUES ($1,$2,$3,$4,$5,$6,TRUE,$7) RETURNING id`,
		userID, chartID, date, category, question, contextData, source).Scan(&id); err != nil {
		return "", 0, err
	}
	if source == "credit" {
		if _, err := tx.Exec(`INSERT INTO today_question_credit_entries
			(user_id,request_id,kind,delta,balance_after)
			VALUES ($1,$2,'spend',-1,$3)`, userID, id, creditBalance-1); err != nil {
			return "", 0, err
		}
	}
	if err := tx.Commit(); err != nil {
		return "", 0, err
	}
	return id, limit, nil
}

type TodayQuestionRecord struct {
	ID               string          `json:"id"`
	UserID           string          `json:"user_id"`
	UserEmail        string          `json:"user_email"`
	UserNickname     string          `json:"user_nickname"`
	ChartID          string          `json:"chart_id"`
	QuestionDate     string          `json:"question_date"`
	Category         string          `json:"category"`
	Question         string          `json:"question"`
	Status           string          `json:"status"`
	ProviderName     string          `json:"provider_name"`
	Model            string          `json:"model"`
	DurationMs       int             `json:"duration_ms"`
	PromptTokens     int             `json:"prompt_tokens"`
	CompletionTokens int             `json:"completion_tokens"`
	TotalTokens      int             `json:"total_tokens"`
	ReasoningTokens  int             `json:"reasoning_tokens"`
	CacheHitTokens   int             `json:"cache_hit_tokens"`
	CacheMissTokens  int             `json:"cache_miss_tokens"`
	EstimatedCostCny float64         `json:"estimated_cost_cny"`
	ErrorMsg         string          `json:"error_msg"`
	CreatedAt        time.Time       `json:"created_at"`
	CompletedAt      sql.NullTime    `json:"-"`
	ContextData      json.RawMessage `json:"context_data,omitempty"`
	AnswerData       json.RawMessage `json:"answer_data,omitempty"`
	AISystemPrompt   string          `json:"ai_system_prompt,omitempty"`
	AIUserPrompt     string          `json:"ai_user_prompt,omitempty"`
	AIOutput         string          `json:"ai_output,omitempty"`
}

type TodayQuestionCompletion struct {
	Status           string
	AnswerData       []byte
	AISystemPrompt   string
	AIUserPrompt     string
	AIOutput         string
	ProviderID       string
	Model            string
	DurationMs       int
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	ReasoningTokens  int
	CacheHitTokens   int
	CacheMissTokens  int
	EstimatedCostCny float64
	ErrorMsg         string
}

func CompleteTodayQuestionRequest(id string, result TodayQuestionCompletion) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var userID string
	if err := tx.QueryRow(`SELECT user_id FROM today_question_requests WHERE id=$1`, id).Scan(&userID); err != nil {
		return err
	}
	if err := lockTodayQuestionUserTx(tx, userID); err != nil {
		return err
	}
	var chartID, source string
	if err := tx.QueryRow(`SELECT chart_id,quota_source FROM today_question_requests WHERE id=$1 AND status='processing' FOR UPDATE`, id).Scan(&chartID, &source); err != nil {
		return err
	}
	var providerID any
	if result.ProviderID != "" {
		providerID = result.ProviderID
	}
	var answerData any
	if len(result.AnswerData) > 0 {
		answerData = result.AnswerData
	}
	var tokenUsageID any
	aiUserPrompt, aiOutput := result.AIUserPrompt, result.AIOutput
	if result.TotalTokens > 0 {
		var linkedID string
		err := tx.QueryRow(`INSERT INTO token_usage_logs
			(user_id, chart_id, call_type, model, provider_id, prompt_tokens, completion_tokens,
			 total_tokens, reasoning_tokens, cache_hit_tokens, cache_miss_tokens, input_content, output_content)
			VALUES ($1,$2,'today_question',$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) RETURNING id`,
			userID, chartID, result.Model, providerID, result.PromptTokens, result.CompletionTokens,
			result.TotalTokens, result.ReasoningTokens, result.CacheHitTokens, result.CacheMissTokens,
			result.AIUserPrompt, result.AIOutput).Scan(&linkedID)
		if err != nil {
			return err
		}
		tokenUsageID = linkedID
		aiUserPrompt, aiOutput = "", ""
	}
	updated, err := tx.Exec(`UPDATE today_question_requests SET
		status=$2::text, answer_data=$3, ai_system_prompt=$4, ai_user_prompt=$5, ai_output=$6,
		provider_id=$7, model=$8, duration_ms=$9, token_usage_id=$10,
		estimated_cost_cny=$11, error_msg=$12, completed_at=now(),
		quota_charged=CASE WHEN $2::text='error' THEN FALSE ELSE quota_charged END
		WHERE id=$1 AND status='processing'`,
		id, result.Status, answerData, result.AISystemPrompt, aiUserPrompt, aiOutput,
		providerID, result.Model, result.DurationMs, tokenUsageID,
		result.EstimatedCostCny, result.ErrorMsg)
	if err != nil {
		return err
	}
	count, err := updated.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return sql.ErrNoRows
	}
	if result.Status == "error" && source == "credit" {
		if err := refundTodayQuestionCreditTx(tx, userID, id); err != nil {
			return err
		}
	}
	return tx.Commit()
}

const todayQuestionListWhere = ` FROM today_question_requests q
	JOIN users u ON u.id=q.user_id
	LEFT JOIN llm_providers p ON p.id=q.provider_id
	LEFT JOIN token_usage_logs t ON t.id=q.token_usage_id
	WHERE ($1='' OR q.question_date >= NULLIF($1,'')::date)
	AND ($2='' OR q.question_date <= NULLIF($2,'')::date)
	AND ($3='' OR q.status=$3)
	AND ($4='' OR u.email ILIKE '%' || $4 || '%' OR u.nickname ILIKE '%' || $4 || '%' OR u.id::text=$4)`

func ListTodayQuestionRequests(from, to, status, user string, page, limit int) (int, []TodayQuestionRecord, error) {
	var total int
	if err := database.DB.QueryRow(`SELECT count(*)`+todayQuestionListWhere, from, to, status, user).Scan(&total); err != nil {
		return 0, nil, fmt.Errorf("count today question requests: %w", err)
	}
	rows, err := database.DB.Query(`SELECT q.id, q.user_id, COALESCE(u.email,''), COALESCE(u.nickname,''),
		q.chart_id, q.question_date::text, q.category, q.question, q.status, COALESCE(p.name,''),
		q.model, q.duration_ms, COALESCE(t.prompt_tokens,q.prompt_tokens), COALESCE(t.completion_tokens,q.completion_tokens), COALESCE(t.total_tokens,q.total_tokens),
		COALESCE(t.reasoning_tokens,q.reasoning_tokens), COALESCE(t.cache_hit_tokens,q.cache_hit_tokens), COALESCE(t.cache_miss_tokens,q.cache_miss_tokens), q.estimated_cost_cny,
		q.error_msg, q.created_at, q.completed_at`+todayQuestionListWhere+`
		ORDER BY q.created_at DESC, q.id DESC LIMIT $5 OFFSET $6`,
		from, to, status, user, limit, (page-1)*limit)
	if err != nil {
		return 0, nil, fmt.Errorf("list today question requests: %w", err)
	}
	defer rows.Close()
	items := []TodayQuestionRecord{}
	for rows.Next() {
		var item TodayQuestionRecord
		if err := rows.Scan(&item.ID, &item.UserID, &item.UserEmail, &item.UserNickname,
			&item.ChartID, &item.QuestionDate, &item.Category, &item.Question, &item.Status, &item.ProviderName,
			&item.Model, &item.DurationMs, &item.PromptTokens, &item.CompletionTokens, &item.TotalTokens,
			&item.ReasoningTokens, &item.CacheHitTokens, &item.CacheMissTokens, &item.EstimatedCostCny,
			&item.ErrorMsg, &item.CreatedAt, &item.CompletedAt); err != nil {
			return 0, nil, fmt.Errorf("scan today question request: %w", err)
		}
		items = append(items, item)
	}
	return total, items, rows.Err()
}

func GetTodayQuestionRequest(id string) (*TodayQuestionRecord, error) {
	var item TodayQuestionRecord
	var contextData, answerData []byte
	err := database.DB.QueryRow(`SELECT q.id, q.user_id, COALESCE(u.email,''), COALESCE(u.nickname,''),
		q.chart_id, q.question_date::text, q.category, q.question, q.status, COALESCE(p.name,''),
		q.model, q.duration_ms, COALESCE(t.prompt_tokens,q.prompt_tokens), COALESCE(t.completion_tokens,q.completion_tokens), COALESCE(t.total_tokens,q.total_tokens),
		COALESCE(t.reasoning_tokens,q.reasoning_tokens), COALESCE(t.cache_hit_tokens,q.cache_hit_tokens), COALESCE(t.cache_miss_tokens,q.cache_miss_tokens), q.estimated_cost_cny,
		q.error_msg, q.created_at, q.completed_at, q.context_data, q.answer_data,
		q.ai_system_prompt, COALESCE(NULLIF(q.ai_user_prompt,''),t.input_content,''),
		COALESCE(NULLIF(q.ai_output,''),t.output_content,'')
		FROM today_question_requests q JOIN users u ON u.id=q.user_id
		LEFT JOIN llm_providers p ON p.id=q.provider_id
		LEFT JOIN token_usage_logs t ON t.id=q.token_usage_id WHERE q.id=$1`, id).Scan(
		&item.ID, &item.UserID, &item.UserEmail, &item.UserNickname, &item.ChartID,
		&item.QuestionDate, &item.Category, &item.Question, &item.Status, &item.ProviderName,
		&item.Model, &item.DurationMs, &item.PromptTokens, &item.CompletionTokens, &item.TotalTokens,
		&item.ReasoningTokens, &item.CacheHitTokens, &item.CacheMissTokens, &item.EstimatedCostCny,
		&item.ErrorMsg, &item.CreatedAt, &item.CompletedAt, &contextData, &answerData,
		&item.AISystemPrompt, &item.AIUserPrompt, &item.AIOutput)
	if err != nil {
		return nil, err
	}
	item.ContextData, item.AnswerData = contextData, answerData
	return &item, nil
}
