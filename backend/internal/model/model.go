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
	Polarity    string    `json:"polarity"`    // ji / xiong / zhong
	Description string    `json:"description"` // 详细说明（命理书级别）
	UpdatedAt   time.Time `json:"updated_at"`
}
