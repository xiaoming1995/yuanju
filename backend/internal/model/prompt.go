package model

import (
	"encoding/json"
	"time"
)

// AIPrompt 系统 Prompt 模板管理
type AIPrompt struct {
	ID          string    `db:"id" json:"id"`
	Module      string    `db:"module" json:"module"` // 例如 "liunian", "natal"
	Content     string    `db:"content" json:"content"`
	Description string    `db:"description" json:"description"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

// AILiunianReport 流年运势精批报告
type AILiunianReport struct {
	ID                string           `db:"id" json:"id"`
	ChartID           string           `db:"chart_id" json:"chart_id"`
	TargetYear        int              `db:"target_year" json:"target_year"`
	DayunGanzhi       string           `db:"dayun_ganzhi" json:"dayun_ganzhi"`
	ContentStructured *json.RawMessage `db:"content_structured" json:"content_structured"`
	Model             string           `db:"model" json:"model"`
	CreatedAt         time.Time        `db:"created_at" json:"created_at"`
}

// LiunianTemplateData 流年 Prompt 模板所需的数据上下文
type LiunianTemplateData struct {
	// 原局分析总结
	NatalAnalysisLogic string
	
	// 当前大运信息
	CurrentDayunGanZhi      string
	CurrentDayunGanShiShen  string
	CurrentDayunZhiShiShen  string
	
	// 目标流年信息
	TargetYear            int
	TargetYearGanZhi      string
	TargetYearGanShiShen  string
	TargetYearZhiShiShen  string
}
