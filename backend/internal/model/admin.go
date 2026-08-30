package model

import (
	"encoding/json"
	"time"
)

// Admin 管理员账号（与普通用户完全隔离）
type Admin struct {
	ID           string    `db:"id" json:"id"`
	Email        string    `db:"email" json:"email"`
	PasswordHash string    `db:"password_hash" json:"-"`
	Name         string    `db:"name" json:"name"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

// LLMProvider LLM 服务提供商配置
type LLMProvider struct {
	ID              string    `db:"id" json:"id"`
	Name            string    `db:"name" json:"name"`
	Type            string    `db:"type" json:"type"`
	BaseURL         string    `db:"base_url" json:"base_url"`
	Model           string    `db:"model" json:"model"`
	APIKeyEncrypted string    `db:"api_key_encrypted" json:"-"`
	APIKeyPreview   string    `db:"api_key_preview" json:"api_key_preview,omitempty"`
	APIKeyMasked    string    `db:"-" json:"api_key_masked,omitempty"`
	ThinkingEnabled bool      `db:"thinking_enabled" json:"thinking_enabled"`
	InputPriceCny   float64   `db:"input_price_cny" json:"input_price_cny"`
	OutputPriceCny  float64   `db:"output_price_cny" json:"output_price_cny"`
	Active          bool      `db:"active" json:"active"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time `db:"updated_at" json:"updated_at"`
}

// AIRequestLog AI 调用日志记录
type AIRequestLog struct {
	ID           string    `db:"id" json:"id"`
	ChartID      string    `db:"chart_id" json:"chart_id,omitempty"`
	ProviderID   string    `db:"provider_id" json:"provider_id,omitempty"`
	ProviderName string    `db:"-" json:"provider_name,omitempty"` // JOIN 查询填充
	Model        string    `db:"model" json:"model"`
	DurationMs   int       `db:"duration_ms" json:"duration_ms"`
	Status       string    `db:"status" json:"status"` // "success" | "error"
	ErrorMsg     string    `db:"error_msg" json:"error_msg,omitempty"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

// AdminChartRecord 管理后台：起盘明细记录
type AdminChartRecord struct {
	ID                 string           `db:"id" json:"id"`
	UserID             *string          `db:"user_id" json:"user_id,omitempty"`
	UserEmail          *string          `db:"user_email" json:"user_email,omitempty"`
	BirthYear          int              `db:"birth_year" json:"birth_year"`
	BirthMonth         int              `db:"birth_month" json:"birth_month"`
	BirthDay           int              `db:"birth_day" json:"birth_day"`
	BirthHour          int              `db:"birth_hour" json:"birth_hour"`
	Gender             string           `db:"gender" json:"gender"`
	YearGan            string           `db:"year_gan" json:"year_gan"`
	YearZhi            string           `db:"year_zhi" json:"year_zhi"`
	MonthGan           string           `db:"month_gan" json:"month_gan"`
	MonthZhi           string           `db:"month_zhi" json:"month_zhi"`
	DayGan             string           `db:"day_gan" json:"day_gan"`
	DayZhi             string           `db:"day_zhi" json:"day_zhi"`
	HourGan            string           `db:"hour_gan" json:"hour_gan"`
	HourZhi            string           `db:"hour_zhi" json:"hour_zhi"`
	Yongshen           string           `db:"yongshen" json:"yongshen"`
	Jishen             string           `db:"jishen" json:"jishen"`
	AIResult           *string          `db:"ai_result" json:"ai_result"`
	AIResultStructured *json.RawMessage `db:"ai_result_structured" json:"ai_result_structured"`
	CreatedAt          time.Time        `db:"created_at" json:"created_at"`
}

// AdminPastEventsDayunRecord 管理后台：过往推算大运段缓存记录
type AdminPastEventsDayunRecord struct {
	ID               string           `json:"id"`
	ChartID          string           `json:"chart_id"`
	DayunIndex       int              `json:"dayun_index"`
	DayunGanZhi      string           `json:"dayun_ganzhi"`
	Themes           *json.RawMessage `json:"themes"`
	Summary          string           `json:"summary"`
	Years            *json.RawMessage `json:"years"`
	Model            string           `json:"model"`
	AlgorithmVersion string           `json:"algorithm_version"`
	CreatedAt        time.Time        `json:"created_at"`
}

// CelebrityRecord 名人八字信息记录
type CelebrityRecord struct {
	ID        string    `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Gender    string    `db:"gender" json:"gender"`
	Traits    string    `db:"traits" json:"traits"`
	Career    string    `db:"career" json:"career"`
	Active    bool      `db:"active" json:"active"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// MingGeHistoricalFigure 是按命格归类、经过人工审核的古人映照内容。
// 它与 legacy celebrity_records 分离，不能作为 AI 自动生成内容直接发布。
type MingGeHistoricalFigure struct {
	ID                   string    `json:"id"`
	MingGe               string    `json:"ming_ge"`
	FigureName           string    `json:"figure_name"`
	Era                  string    `json:"era"`
	Identity             string    `json:"identity"`
	HistoricalMemory     string    `json:"historical_memory"`
	TurningPoint         string    `json:"turning_point"`
	TurningPointYear     string    `json:"turning_point_year"`
	SourceTitle          string    `json:"source_title"`
	SourceURL            string    `json:"source_url"`
	BirthDataPrecision   string    `json:"birth_data_precision"`
	BaziVerificationNote string    `json:"bazi_verification_note,omitempty"`
	DayunPeriod          string    `json:"dayun_period,omitempty"`
	DayunExplanation     string    `json:"dayun_explanation,omitempty"`
	ShowDayun            bool      `json:"show_dayun"`
	DisplayOrder         int       `json:"display_order"`
	ReviewStatus         string    `json:"review_status"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// AdminCompatListItem 后台合盘明细列表行
type AdminCompatListItem struct {
	ID                string    `json:"id"`
	UserEmail         *string   `json:"user_email"`
	SelfName          string    `json:"self_name"`
	PartnerName       string    `json:"partner_name"`
	OverallScore      int       `json:"overall_score"`
	OverallLevel      string    `json:"overall_level"`
	RelationshipStage string    `json:"relationship_stage"`
	PrimaryQuestion   string    `json:"primary_question"`
	AnalysisVersion   string    `json:"analysis_version"`
	CreatedAt         time.Time `json:"created_at"`
}

// PredefinedProviders 预设的 Provider 类型（用于前端下拉）
var PredefinedProviders = []map[string]string{
	{"type": "deepseek", "name": "DeepSeek V4 Flash", "base_url": "https://api.deepseek.com", "model": "deepseek-v4-flash"},
	{"type": "deepseek", "name": "DeepSeek V4 Pro（思考）", "base_url": "https://api.deepseek.com", "model": "deepseek-v4-pro"},
	{"type": "openai", "name": "OpenAI", "base_url": "https://api.openai.com", "model": "gpt-4o-mini"},
	{"type": "kimi", "name": "Kimi K2.5（月之暗面）", "base_url": "https://api.moonshot.cn/v1", "model": "kimi-k2.5"},
	{"type": "qwen", "name": "阿里 Qwen", "base_url": "https://dashscope.aliyuncs.com/compatible-mode", "model": "qwen-plus"},
	{"type": "claude", "name": "Anthropic Claude", "base_url": "https://api.anthropic.com", "model": "claude-3-5-sonnet-20241022"},
	{"type": "gemini", "name": "Google Gemini", "base_url": "https://generativelanguage.googleapis.com/v1beta/openai", "model": "gemini-2.0-flash"},
	{"type": "custom", "name": "自定义", "base_url": "", "model": ""},
}
