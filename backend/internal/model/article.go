package model

import (
	"encoding/json"
	"time"
)

type ArticleStatus string

const (
	ArticleStatusCandidate ArticleStatus = "candidate"
	ArticleStatusPublished ArticleStatus = "published"
	ArticleStatusRejected  ArticleStatus = "rejected"
	ArticleStatusTakenDown ArticleStatus = "taken_down"
	ArticleStatusDeleted   ArticleStatus = "deleted"
)

type CollectionTaskStatus string

const (
	CollectionTaskStatusPending   CollectionTaskStatus = "pending"
	CollectionTaskStatusRunning   CollectionTaskStatus = "running"
	CollectionTaskStatusSucceeded CollectionTaskStatus = "succeeded"
	CollectionTaskStatusFailed    CollectionTaskStatus = "failed"
)

type CollectionItemStatus string

const (
	CollectionItemStatusInserted  CollectionItemStatus = "inserted"
	CollectionItemStatusDuplicate CollectionItemStatus = "duplicate"
	CollectionItemStatusFailed    CollectionItemStatus = "failed"
	CollectionItemStatusSkipped   CollectionItemStatus = "skipped"
)

type ArticleAIStatus string

const (
	ArticleAIStatusPending   ArticleAIStatus = "pending"
	ArticleAIStatusSucceeded ArticleAIStatus = "succeeded"
	ArticleAIStatusFailed    ArticleAIStatus = "failed"
)

type ArticleBodyFetchStatus string

const (
	ArticleBodyFetchStatusPending              ArticleBodyFetchStatus = "pending"
	ArticleBodyFetchStatusSucceeded            ArticleBodyFetchStatus = "succeeded"
	ArticleBodyFetchStatusManual               ArticleBodyFetchStatus = "manual"
	ArticleBodyFetchStatusSogouVerifyRequired  ArticleBodyFetchStatus = "sogou_verify_required"
	ArticleBodyFetchStatusSogouRedirectMissing ArticleBodyFetchStatus = "sogou_redirect_missing"
	ArticleBodyFetchStatusWechatNoJSContent    ArticleBodyFetchStatus = "wechat_no_js_content"
	ArticleBodyFetchStatusWechatVideoPage      ArticleBodyFetchStatus = "wechat_video_page"
	ArticleBodyFetchStatusWechatAntispider     ArticleBodyFetchStatus = "wechat_antispider"
	ArticleBodyFetchStatusTimeout              ArticleBodyFetchStatus = "timeout"
	ArticleBodyFetchStatusHTTPError            ArticleBodyFetchStatus = "http_error"
	ArticleBodyFetchStatusFailed               ArticleBodyFetchStatus = "failed"
)

type ArticleCategory struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	SortOrder int       `json:"sort_order"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ArticleTag struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ArticleKeyword struct {
	ID        string    `json:"id"`
	Keyword   string    `json:"keyword"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ArticleAIAnalysis struct {
	OneSentenceSummary string   `json:"one_sentence_summary"`
	KeyPoints          []string `json:"key_points"`
	TargetReaders      []string `json:"target_readers"`
	RelatedTopics      []string `json:"related_topics"`
	SuggestedTags      []string `json:"suggested_tags"`
	TitlePattern       string   `json:"title_pattern"`
	OpeningStyle       string   `json:"opening_style"`
	StructureOutline   []string `json:"structure_outline"`
	ExpressionStyle    []string `json:"expression_style"`
	RewriteAngles      []string `json:"rewrite_angles"`
}

type ArticleQualityReason struct {
	Type    string `json:"type"`
	Points  int    `json:"points"`
	Message string `json:"message"`
}

type Article struct {
	ID                 string                 `json:"id"`
	Title              string                 `json:"title"`
	SourceName         string                 `json:"source_name"`
	OriginalURL        string                 `json:"original_url"`
	CanonicalURLHash   string                 `json:"canonical_url_hash"`
	CoverURL           string                 `json:"cover_url"`
	PublishedAtSource  *time.Time             `json:"published_at_source,omitempty"`
	SearchSnippet      string                 `json:"search_snippet"`
	Summary            string                 `json:"summary"`
	AIAnalysis         *ArticleAIAnalysis     `json:"ai_analysis,omitempty"`
	AIAnalysisRaw      json.RawMessage        `json:"-"`
	AIStatus           ArticleAIStatus        `json:"ai_status"`
	AIErrorMsg         string                 `json:"ai_error_msg,omitempty"`
	CategoryID         *string                `json:"category_id,omitempty"`
	Category           *ArticleCategory       `json:"category,omitempty"`
	Tags               []ArticleTag           `json:"tags,omitempty"`
	Status             ArticleStatus          `json:"status"`
	ViewCount          int                    `json:"view_count"`
	OriginalClickCount int                    `json:"original_click_count"`
	QualityScore       int                    `json:"quality_score"`
	QualityReasons     []ArticleQualityReason `json:"quality_reasons,omitempty"`
	QualityReasonsRaw  json.RawMessage        `json:"-"`
	FullTextAuthorized bool                   `json:"full_text_authorized"`
	BodyContent        string                 `json:"body_content,omitempty"`
	BodyFetchStatus    ArticleBodyFetchStatus `json:"body_fetch_status"`
	BodyFetchError     string                 `json:"body_fetch_error,omitempty"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
	PublishedAt        *time.Time             `json:"published_at,omitempty"`
	TakenDownAt        *time.Time             `json:"taken_down_at,omitempty"`
	DeletedAt          *time.Time             `json:"deleted_at,omitempty"`
}

type ArticleCollectionTask struct {
	ID             string               `json:"id"`
	TriggerType    string               `json:"trigger_type"`
	Status         CollectionTaskStatus `json:"status"`
	StartedAt      time.Time            `json:"started_at"`
	FinishedAt     *time.Time           `json:"finished_at,omitempty"`
	KeywordCount   int                  `json:"keyword_count"`
	FoundCount     int                  `json:"found_count"`
	InsertedCount  int                  `json:"inserted_count"`
	DuplicateCount int                  `json:"duplicate_count"`
	FailedCount    int                  `json:"failed_count"`
	ErrorMsg       string               `json:"error_msg,omitempty"`
	CreatedAt      time.Time            `json:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at"`
}

type ArticleCollectionTaskItem struct {
	ID                string                 `json:"id"`
	TaskID            string                 `json:"task_id"`
	ArticleID         *string                `json:"article_id,omitempty"`
	OriginalURL       string                 `json:"original_url"`
	Keyword           string                 `json:"keyword"`
	Status            CollectionItemStatus   `json:"status"`
	ErrorMsg          string                 `json:"error_msg,omitempty"`
	ArticleTitle      string                 `json:"article_title,omitempty"`
	SourceName        string                 `json:"source_name,omitempty"`
	CoverURL          string                 `json:"cover_url,omitempty"`
	BodyFetchStatus   ArticleBodyFetchStatus `json:"body_fetch_status,omitempty"`
	AIStatus          ArticleAIStatus        `json:"ai_status,omitempty"`
	QualityScore      int                    `json:"quality_score"`
	QualityReasons    []ArticleQualityReason `json:"quality_reasons,omitempty"`
	SkipReason        string                 `json:"skip_reason,omitempty"`
	AutoPublished     bool                   `json:"auto_published"`
	AutoPublishReason string                 `json:"auto_publish_reason,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

type ArticleAuditEvent struct {
	ID         string        `json:"id"`
	ArticleID  string        `json:"article_id"`
	AdminID    string        `json:"admin_id"`
	Action     string        `json:"action"`
	FromStatus ArticleStatus `json:"from_status"`
	ToStatus   ArticleStatus `json:"to_status"`
	Note       string        `json:"note,omitempty"`
	CreatedAt  time.Time     `json:"created_at"`
}

type ArticleOriginalClick struct {
	ID        string    `json:"id"`
	ArticleID string    `json:"article_id"`
	UserID    string    `json:"user_id"`
	ClickedAt time.Time `json:"clicked_at"`
}

type ArticleCollectionConfig struct {
	ID                         string     `json:"id"`
	Enabled                    bool       `json:"enabled"`
	Frequency                  string     `json:"frequency"`
	ScheduleInterval           int        `json:"schedule_interval_minutes"`
	MaxResultsPerRun           int        `json:"max_results_per_run"`
	SearchPageMin              int        `json:"search_page_min"`
	SearchPageMax              int        `json:"search_page_max"`
	AutoPublishEnabled         bool       `json:"auto_publish_enabled"`
	AutoPublishMinQualityScore int        `json:"auto_publish_min_quality_score"`
	AutoPublishRequiresBody    bool       `json:"auto_publish_requires_body"`
	AutoPublishMaxPerRun       int        `json:"auto_publish_max_per_run"`
	LastRunAt                  *time.Time `json:"last_run_at,omitempty"`
	UpdatedAt                  time.Time  `json:"updated_at"`
}

type ArticleQualityConfig struct {
	ID                    string    `json:"id"`
	Enabled               bool      `json:"quality_filter_enabled"`
	MinQualityScore       int       `json:"min_quality_score"`
	AllowWithoutBody      bool      `json:"allow_without_body"`
	BonusKeywords         []string  `json:"bonus_keywords"`
	SourceBlacklist       []string  `json:"source_blacklist"`
	PreferredSources      []string  `json:"preferred_sources"`
	AIQualityCheckEnabled bool      `json:"ai_quality_check_enabled"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type ArticleAIProvider struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Type            string    `json:"type"`
	BaseURL         string    `json:"base_url"`
	Model           string    `json:"model"`
	APIKeyEncrypted string    `json:"-"`
	APIKeyPreview   string    `json:"api_key_preview,omitempty"`
	APIKeyMasked    string    `json:"api_key_masked,omitempty"`
	Active          bool      `json:"active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type ArticleAIPrompt struct {
	ID          string    `json:"id"`
	Content     string    `json:"content"`
	Description string    `json:"description"`
	UpdatedAt   time.Time `json:"updated_at"`
}
