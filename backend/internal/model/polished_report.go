package model

import (
	"encoding/json"
	"time"
)

// PolishedReport 润色版命理解读
// 一个 chart 对应至多 1 条记录（UNIQUE chart_id），重生覆盖。
type PolishedReport struct {
	ID                string           `json:"id"`
	ChartID           string           `json:"chart_id"`
	UserSituation     string           `json:"user_situation"`
	Content           string           `json:"content"`
	ContentStructured *json.RawMessage `json:"content_structured,omitempty"`
	Model             string           `json:"model"`
	PromptTokens      int              `json:"prompt_tokens"`
	CompletionTokens  int              `json:"completion_tokens"`
	TotalTokens       int              `json:"total_tokens"`
	CreatedAt         time.Time        `json:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at"`
}
