package repository

import (
	"database/sql"
	"log"
	"yuanju/internal/model"
	"yuanju/pkg/database"
)

// ====== Admin Repository ======

func CreateAdmin(email, passwordHash, name string) (*model.Admin, error) {
	admin := &model.Admin{}
	err := database.DB.QueryRow(
		`INSERT INTO admins (email, password_hash, name)
		 VALUES ($1, $2, $3)
		 RETURNING id, email, name, created_at`,
		email, passwordHash, name,
	).Scan(&admin.ID, &admin.Email, &admin.Name, &admin.CreatedAt)
	return admin, err
}

func GetAdminByEmail(email string) (*model.Admin, error) {
	admin := &model.Admin{}
	err := database.DB.QueryRow(
		`SELECT id, email, password_hash, name, created_at FROM admins WHERE email = $1`,
		email,
	).Scan(&admin.ID, &admin.Email, &admin.PasswordHash, &admin.Name, &admin.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return admin, err
}

// ====== LLM Provider Repository ======

func ListLLMProviders() ([]model.LLMProvider, error) {
	rows, err := database.DB.Query(
		`SELECT id, name, type, base_url, model, api_key_encrypted, api_key_preview, thinking_enabled, input_price_cny, output_price_cny, active, created_at, updated_at
		 FROM llm_providers ORDER BY created_at ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var providers []model.LLMProvider
	for rows.Next() {
		var p model.LLMProvider
		if err := rows.Scan(&p.ID, &p.Name, &p.Type, &p.BaseURL, &p.Model,
			&p.APIKeyEncrypted, &p.APIKeyPreview, &p.ThinkingEnabled, &p.InputPriceCny, &p.OutputPriceCny, &p.Active, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		providers = append(providers, p)
	}
	return providers, nil
}

func GetActiveLLMProvider() (*model.LLMProvider, error) {
	p := &model.LLMProvider{}
	err := database.DB.QueryRow(
		`SELECT id, name, type, base_url, model, api_key_encrypted, thinking_enabled, active, created_at, updated_at
		 FROM llm_providers WHERE active = true LIMIT 1`,
	).Scan(&p.ID, &p.Name, &p.Type, &p.BaseURL, &p.Model,
		&p.APIKeyEncrypted, &p.ThinkingEnabled, &p.Active, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return p, err
}

func CreateLLMProvider(name, typ, baseURL, modelName, encryptedKey, preview string, thinkingEnabled bool, inputPrice, outputPrice float64) (*model.LLMProvider, error) {
	p := &model.LLMProvider{}
	err := database.DB.QueryRow(
		`INSERT INTO llm_providers (name, type, base_url, model, api_key_encrypted, api_key_preview, thinking_enabled, input_price_cny, output_price_cny)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING id, name, type, base_url, model, api_key_encrypted, api_key_preview, thinking_enabled, input_price_cny, output_price_cny, active, created_at, updated_at`,
		name, typ, baseURL, modelName, encryptedKey, preview, thinkingEnabled, inputPrice, outputPrice,
	).Scan(&p.ID, &p.Name, &p.Type, &p.BaseURL, &p.Model,
		&p.APIKeyEncrypted, &p.APIKeyPreview, &p.ThinkingEnabled, &p.InputPriceCny, &p.OutputPriceCny, &p.Active, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

func UpdateLLMProvider(id, name, baseURL, modelName, encryptedKey, preview string, thinkingEnabled bool, inputPrice, outputPrice float64) error {
	_, err := database.DB.Exec(
		`UPDATE llm_providers
		 SET name=$1, base_url=$2, model=$3, api_key_encrypted=$4,
		     api_key_preview=CASE WHEN $5 != '' THEN $5 ELSE api_key_preview END,
		     thinking_enabled=$6, input_price_cny=$7, output_price_cny=$8, updated_at=NOW()
		 WHERE id=$9`,
		name, baseURL, modelName, encryptedKey, preview, thinkingEnabled, inputPrice, outputPrice, id,
	)
	return err
}

func GetPriceByModel(modelName string) (inputPrice, outputPrice float64, err error) {
	err = database.DB.QueryRow(
		`SELECT input_price_cny, output_price_cny FROM llm_providers WHERE model = $1 LIMIT 1`,
		modelName,
	).Scan(&inputPrice, &outputPrice)
	return
}

func ActivateLLMProvider(id string) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.Exec(`UPDATE llm_providers SET active=false`); err != nil {
		return err
	}
	if _, err = tx.Exec(`UPDATE llm_providers SET active=true, updated_at=NOW() WHERE id=$1`, id); err != nil {
		return err
	}
	return tx.Commit()
}

func DeleteLLMProvider(id string) (bool, error) {
	// 检查是否激活中
	var active bool
	err := database.DB.QueryRow(`SELECT active FROM llm_providers WHERE id=$1`, id).Scan(&active)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if active {
		return false, nil // 调用方检查返回 false
	}
	_, err = database.DB.Exec(`DELETE FROM llm_providers WHERE id=$1`, id)
	return true, err
}

func LLMProviderExists(id string) bool {
	var count int
	database.DB.QueryRow(`SELECT COUNT(*) FROM llm_providers WHERE id=$1`, id).Scan(&count)
	return count > 0
}

// ====== AI Request Log Repository ======

func CreateAIRequestLog(chartID, providerID, model string, durationMs int, status, errorMsg string) {
	// 在 Go 层将空字符串转为 nil，避免 NULLIF($1,'') 返回 text 而非 uuid 导致的类型不匹配错误
	var chartIDParam, providerIDParam interface{}
	if chartID != "" {
		chartIDParam = chartID
	}
	if providerID != "" {
		providerIDParam = providerID
	}
	var errMsgParam interface{}
	if errorMsg != "" {
		errMsgParam = errorMsg
	}

	_, err := database.DB.Exec(
		`INSERT INTO ai_requests_log (chart_id, provider_id, model, duration_ms, status, error_msg)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		chartIDParam, providerIDParam, model, durationMs, status, errMsgParam,
	)
	if err != nil {
		log.Printf("[CreateAIRequestLog] 写入日志失败: %v", err)
	}
}

// ListAIRequestLogs 分页查询 AI 调用日志，支持按 status 筛选
func ListAIRequestLogs(page, pageSize int, statusFilter string) ([]model.AIRequestLog, int, error) {
	offset := (page - 1) * pageSize

	var rows *sql.Rows
	var err error
	var total int

	if statusFilter != "" && statusFilter != "all" {
		err = database.DB.QueryRow(
			`SELECT COUNT(*) FROM ai_requests_log WHERE status = $1`, statusFilter,
		).Scan(&total)
		if err != nil {
			return nil, 0, err
		}
		rows, err = database.DB.Query(
			`SELECT l.id, COALESCE(l.chart_id::text,''), COALESCE(l.provider_id::text,''),
			        COALESCE(p.name,''), l.model, l.duration_ms, l.status,
			        COALESCE(l.error_msg,''), l.created_at
			 FROM ai_requests_log l
			 LEFT JOIN llm_providers p ON l.provider_id = p.id
			 WHERE l.status = $1
			 ORDER BY l.created_at DESC
			 LIMIT $2 OFFSET $3`,
			statusFilter, pageSize, offset,
		)
	} else {
		err = database.DB.QueryRow(`SELECT COUNT(*) FROM ai_requests_log`).Scan(&total)
		if err != nil {
			return nil, 0, err
		}
		rows, err = database.DB.Query(
			`SELECT l.id, COALESCE(l.chart_id::text,''), COALESCE(l.provider_id::text,''),
			        COALESCE(p.name,''), l.model, l.duration_ms, l.status,
			        COALESCE(l.error_msg,''), l.created_at
			 FROM ai_requests_log l
			 LEFT JOIN llm_providers p ON l.provider_id = p.id
			 ORDER BY l.created_at DESC
			 LIMIT $1 OFFSET $2`,
			pageSize, offset,
		)
	}
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []model.AIRequestLog
	for rows.Next() {
		var l model.AIRequestLog
		if err := rows.Scan(
			&l.ID, &l.ChartID, &l.ProviderID, &l.ProviderName,
			&l.Model, &l.DurationMs, &l.Status, &l.ErrorMsg, &l.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		logs = append(logs, l)
	}
	return logs, total, nil
}

// AILogDayStat 近 7 天某一天的调用统计
type AILogDayStat struct {
	Date          string  `json:"date"`
	Total         int     `json:"total"`
	SuccessCount  int     `json:"success_count"`
	ErrorCount    int     `json:"error_count"`
	AvgDurationMs float64 `json:"avg_duration_ms"`
}

// GetAILogsSummary 返回近 7 天按天分组的 AI 调用统计（无数据日期补零）
func GetAILogsSummary() ([]AILogDayStat, error) {
	rows, err := database.DB.Query(`
		SELECT
			TO_CHAR(d.day, 'YYYY-MM-DD') AS date,
			COUNT(l.id) AS total,
			COUNT(CASE WHEN l.status = 'success' THEN 1 END) AS success_count,
			COUNT(CASE WHEN l.status = 'error' THEN 1 END) AS error_count,
			COALESCE(AVG(l.duration_ms), 0) AS avg_duration_ms
		FROM (
			SELECT generate_series(
				CURRENT_DATE - INTERVAL '6 days',
				CURRENT_DATE,
				INTERVAL '1 day'
			)::date AS day
		) d
		LEFT JOIN ai_requests_log l ON DATE(l.created_at) = d.day
		GROUP BY d.day
		ORDER BY d.day ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []AILogDayStat
	for rows.Next() {
		var s AILogDayStat
		if err := rows.Scan(&s.Date, &s.Total, &s.SuccessCount, &s.ErrorCount, &s.AvgDurationMs); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, nil
}

// ====== Bazi Charts (Admin view) ======

// ListBaziCharts 分页查询所有用户的排盘记录
func ListBaziCharts(page, pageSize int) ([]model.AdminChartRecord, int, error) {
	offset := (page - 1) * pageSize

	// 统计总数
	var total int
	err := database.DB.QueryRow(`SELECT COUNT(*) FROM bazi_charts`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []model.AdminChartRecord{}, 0, nil
	}

	// 使用相关子查询替代 LEFT JOIN ai_reports，避免多条报告时产生重复行
	rows, err := database.DB.Query(`
		SELECT 
			c.id, c.user_id, u.email as user_email, 
			c.birth_year, c.birth_month, c.birth_day, c.birth_hour, 
			c.gender, c.year_gan, c.year_zhi, 
			c.month_gan, c.month_zhi, c.day_gan, c.day_zhi, c.hour_gan, c.hour_zhi,
			COALESCE(c.yongshen, '') as yongshen, 
			COALESCE(c.jishen, '') as jishen, 
			(SELECT content FROM ai_reports WHERE chart_id=c.id ORDER BY created_at DESC LIMIT 1) as ai_result,
			(SELECT content_structured FROM ai_reports WHERE chart_id=c.id ORDER BY created_at DESC LIMIT 1) as ai_result_structured,
			c.created_at
		FROM bazi_charts c
		LEFT JOIN users u ON c.user_id = u.id
		ORDER BY c.created_at DESC
		LIMIT $1 OFFSET $2
	`, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var charts []model.AdminChartRecord
	for rows.Next() {
		var r model.AdminChartRecord
		if err := rows.Scan(
			&r.ID, &r.UserID, &r.UserEmail,
			&r.BirthYear, &r.BirthMonth, &r.BirthDay, &r.BirthHour,
			&r.Gender, &r.YearGan, &r.YearZhi,
			&r.MonthGan, &r.MonthZhi, &r.DayGan, &r.DayZhi, &r.HourGan, &r.HourZhi,
			&r.Yongshen, &r.Jishen,
			&r.AIResult,
			&r.AIResultStructured,
			&r.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		charts = append(charts, r)
	}

	return charts, total, nil
}
