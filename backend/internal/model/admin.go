package model

import "time"

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
	APIKeyMasked    string    `db:"-" json:"api_key_masked,omitempty"`
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
	ID         string    `db:"id" json:"id"`
	UserID     *string   `db:"user_id" json:"user_id,omitempty"`
	UserEmail  *string   `db:"user_email" json:"user_email,omitempty"`
	BirthYear  int       `db:"birth_year" json:"birth_year"`
	BirthMonth int       `db:"birth_month" json:"birth_month"`
	BirthDay   int       `db:"birth_day" json:"birth_day"`
	BirthHour  int       `db:"birth_hour" json:"birth_hour"`
	Gender     string    `db:"gender" json:"gender"`
	YearGan    string    `db:"year_gan" json:"year_gan"`
	YearZhi    string    `db:"year_zhi" json:"year_zhi"`
	MonthGan   string    `db:"month_gan" json:"month_gan"`
	MonthZhi   string    `db:"month_zhi" json:"month_zhi"`
	DayGan     string    `db:"day_gan" json:"day_gan"`
	DayZhi     string    `db:"day_zhi" json:"day_zhi"`
	HourGan    string    `db:"hour_gan" json:"hour_gan"`
	HourZhi    string    `db:"hour_zhi" json:"hour_zhi"`
	Yongshen   string    `db:"yongshen" json:"yongshen"`
	Jishen     string    `db:"jishen" json:"jishen"`
	AIResult   *string   `db:"ai_result" json:"ai_result"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
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

// PredefinedProviders 预设的 Provider 类型（用于前端下拉）
var PredefinedProviders = []map[string]string{
	{"type": "deepseek", "name": "DeepSeek Chat (V3)", "base_url": "https://api.deepseek.com", "model": "deepseek-chat"},
	{"type": "deepseek-reasoner", "name": "DeepSeek Reasoner (R1)", "base_url": "https://api.deepseek.com", "model": "deepseek-reasoner"},
	{"type": "openai", "name": "OpenAI", "base_url": "https://api.openai.com", "model": "gpt-4o-mini"},
	{"type": "kimi", "name": "Kimi K2.5（月之暗面）", "base_url": "https://api.moonshot.cn/v1", "model": "kimi-k2.5"},
	{"type": "qwen", "name": "阿里 Qwen", "base_url": "https://dashscope.aliyuncs.com/compatible-mode", "model": "qwen-plus"},
	{"type": "claude", "name": "Anthropic Claude", "base_url": "https://api.anthropic.com", "model": "claude-3-5-sonnet-20241022"},
	{"type": "gemini", "name": "Google Gemini", "base_url": "https://generativelanguage.googleapis.com/v1beta/openai", "model": "gemini-2.0-flash"},
	{"type": "custom", "name": "自定义", "base_url": "", "model": ""},
}
