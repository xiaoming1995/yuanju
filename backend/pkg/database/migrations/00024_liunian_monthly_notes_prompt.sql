-- +goose Up
-- +goose StatementBegin

-- Extend the maintained liunian prompt so each cached annual report also stores
-- 12 monthly notes. Existing reports remain valid; users can regenerate a year
-- to fill monthly_notes.

UPDATE ai_prompts
SET content = REPLACE(
	REPLACE(
		content,
		'请为命主详细批断【{{.TargetYear}} {{.TargetYearGanZhi}}流年】运程（流年干十神={{.TargetYearGanShiShen}} 支十神={{.TargetYearZhiShiShen}}）。',
		'请为命主详细批断【{{.TargetYear}} {{.TargetYearGanZhi}}流年】运程（流年干十神={{.TargetYearGanShiShen}} 支十神={{.TargetYearZhiShiShen}}）。

本年 12 个流月信息（用于生成月度注意点，请按 index 原样返回）：
{{range .LiuYue}}- index={{.Index}} {{.MonthLabel}} {{.MonthName}} {{.GanZhi}}（{{.JieQiName}}，{{.StartDate}} 至 {{.EndDate}}，干十神={{.GanShiShen}}，支十神={{.ZhiShiShen}}）
{{end}}'
	),
	'  "advice": "年度锦囊（一句话点睛）"
}',
	'  "advice": "年度锦囊（一句话点睛）",
  "monthly_notes": [
    {
      "index": 0,
      "month_label": "2月",
      "liuyue_name": "寅月",
      "gan_zhi": "庚寅",
      "summary": "围绕本月整体节奏写 50-80 字，说明主要机会、压力和处理原则。",
      "career": "事业财运注意点，20-40字。",
      "romance": "感情桃花注意点，20-40字。",
      "health": "健康风险注意点，20-40字。"
    }
  ]
}
monthly_notes 必须包含 12 条，index 从 0 到 11，与上方流月信息一一对应；month_label、liuyue_name、gan_zhi 必须使用上方给定值，不要自行推算。'
),
updated_at = NOW()
WHERE module = 'liunian'
  AND content NOT LIKE '%monthly_notes%'
  AND content LIKE '%{{.TargetYearGanZhi}}流年%'
  AND content LIKE '%"advice": "年度锦囊（一句话点睛）"%';

-- +goose StatementEnd
