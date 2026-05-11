package model

import (
	"encoding/json"
	"time"
)

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Nickname     string    `json:"nickname"`
	CreatedAt    time.Time `json:"created_at"`
}

type BaziChart struct {
	ID         string      `json:"id"`
	UserID     *string     `json:"user_id,omitempty"`
	BirthYear  int         `json:"birth_year"`
	BirthMonth int         `json:"birth_month"`
	BirthDay   int         `json:"birth_day"`
	BirthHour  int         `json:"birth_hour"`
	Gender     string      `json:"gender"`
	YearGan    string      `json:"year_gan"`
	YearZhi    string      `json:"year_zhi"`
	MonthGan   string      `json:"month_gan"`
	MonthZhi   string      `json:"month_zhi"`
	DayGan     string      `json:"day_gan"`
	DayZhi     string      `json:"day_zhi"`
	HourGan    string      `json:"hour_gan"`
	HourZhi    string      `json:"hour_zhi"`
	Wuxing     interface{} `json:"wuxing"`
	Dayun      interface{} `json:"dayun"`
	Yongshen   string      `json:"yongshen"`
	Jishen     string      `json:"jishen"`
	ChartHash    string      `json:"chart_hash"`
	CalendarType string      `json:"calendar_type"` // "solar" 或 "lunar"
	IsLeapMonth  bool        `json:"is_leap_month"` // 农历闰月标识
	CreatedAt    time.Time   `json:"created_at"`
}

type AIReport struct {
	ID                string           `json:"id"`
	ChartID           string           `json:"chart_id"`
	Content           string           `json:"content"`
	ContentStructured *json.RawMessage `json:"content_structured,omitempty"`
	Model             string           `json:"model"`
	CreatedAt         time.Time        `json:"created_at"`
}

type WuxingData struct {
	Jin     int     `json:"jin"`
	Mu      int     `json:"mu"`
	Shui    int     `json:"shui"`
	Huo     int     `json:"huo"`
	Tu      int     `json:"tu"`
	Total   int     `json:"total"`
	JinPct  float64 `json:"jin_pct"`
	MuPct   float64 `json:"mu_pct"`
	ShuiPct float64 `json:"shui_pct"`
	HuoPct  float64 `json:"huo_pct"`
	TuPct   float64 `json:"tu_pct"`
}

type LiuNianItem struct {
	Year       int    `json:"year"`
	Age        int    `json:"age"`
	GanZhi     string `json:"gan_zhi"`
	GanShiShen string `json:"gan_shishen"` // 流年干对日主十神
	ZhiShiShen string `json:"zhi_shishen"` // 流年支对日主十神
}

type DayunItem struct {
	Index      int           `json:"index"`
	Gan        string        `json:"gan"`
	Zhi        string        `json:"zhi"`
	StartAge   int           `json:"start_age"`
	StartYear  int           `json:"start_year"`
	EndYear    int           `json:"end_year"`
	GanShiShen string        `json:"gan_shishen"`
	ZhiShiShen string        `json:"zhi_shishen"`
	DiShi      string        `json:"di_shi"`
	LiuNian    []LiuNianItem `json:"liu_nian"`
}

// ShenshaAnnotation 神煞注解（存储详细说明文案）
type ShenshaAnnotation struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Polarity    string    `json:"polarity"`   // ji / xiong / zhong
	Category    string    `json:"category"`   // 贵人系/桃花系/凶煞系/...
	ShortDesc   string    `json:"short_desc"` // 一句话简介
	Description string    `json:"description"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// AIPastEvents 过往年份事件推算报告（命盘级缓存）
type AIPastEvents struct {
	ID                string           `json:"id"`
	ChartID           string           `json:"chart_id"`
	ContentStructured *json.RawMessage `json:"content_structured"`
	Model             string           `json:"model"`
	CreatedAt         time.Time        `json:"created_at"`
}

// AIDayunSummary 单段大运 AI 总结（按 chart_id + dayun_index 缓存）
type AIDayunSummary struct {
	ID          string           `json:"id"`
	ChartID     string           `json:"chart_id"`
	DayunIndex  int              `json:"dayun_index"`
	DayunGanZhi string           `json:"dayun_ganzhi"`
	Themes      *json.RawMessage `json:"themes"`
	Summary     string           `json:"summary"`
	Model       string           `json:"model"`
	CreatedAt   time.Time        `json:"created_at"`
}

// DayunSummaryTemplateData 单段大运 AI prompt 模板上下文
type DayunSummaryTemplateData struct {
	Gender         string
	DayGan         string
	NatalSummary   string
	YongshenInfo   string
	StrengthDetail string
	DayunInfo      string // 当前大运："大运 戊戌 30-39岁（2025-2034年）[偏财/伤官]"
	HuaheTag       string // 合化标签（若有）
	YearsData      string // 仅这段大运 10 年的 signals JSON
	LifeStageHint  string // 读书期/跨界期 prompt 提示（按 youngRatio 三档）
}

// PastEventsTemplateData past_events Prompt 模板所需的数据上下文
type PastEventsTemplateData struct {
	Gender         string // 性别（男/女）
	DayGan         string // 日干
	NatalSummary   string // 原局概要
	YearsData      string // JSON 格式的年份信号列表（每条信号含 polarity / source 字段）
	DayunList      string // 大运列表文字描述（每行一条，含干支/十神/起止年龄）
	YongshenInfo   string // 用神/忌神描述（如"用神：火、土 / 忌神：水、木"），缺失时为空串
	GejuSummary    string // 原局格局描述（无格局信息时为空串）
	DayunHuahe     string // 大运合化标签（多行换行拼接），无则为空串
	StrengthDetail string // 加权身强弱评分明细（如"中和(评分2): 月令同气加分5"）
}
