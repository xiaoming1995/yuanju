package repository

import (
	"database/sql"
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
		`SELECT id, name, type, base_url, model, api_key_encrypted, active, created_at, updated_at
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
			&p.APIKeyEncrypted, &p.Active, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		providers = append(providers, p)
	}
	return providers, nil
}

func GetActiveLLMProvider() (*model.LLMProvider, error) {
	p := &model.LLMProvider{}
	err := database.DB.QueryRow(
		`SELECT id, name, type, base_url, model, api_key_encrypted, active, created_at, updated_at
		 FROM llm_providers WHERE active = true LIMIT 1`,
	).Scan(&p.ID, &p.Name, &p.Type, &p.BaseURL, &p.Model,
		&p.APIKeyEncrypted, &p.Active, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return p, err
}

func CreateLLMProvider(name, typ, baseURL, modelName, encryptedKey string) (*model.LLMProvider, error) {
	p := &model.LLMProvider{}
	err := database.DB.QueryRow(
		`INSERT INTO llm_providers (name, type, base_url, model, api_key_encrypted)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, name, type, base_url, model, api_key_encrypted, active, created_at, updated_at`,
		name, typ, baseURL, modelName, encryptedKey,
	).Scan(&p.ID, &p.Name, &p.Type, &p.BaseURL, &p.Model,
		&p.APIKeyEncrypted, &p.Active, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

func UpdateLLMProvider(id, name, baseURL, modelName, encryptedKey string) error {
	_, err := database.DB.Exec(
		`UPDATE llm_providers
		 SET name=$1, base_url=$2, model=$3, api_key_encrypted=$4, updated_at=NOW()
		 WHERE id=$5`,
		name, baseURL, modelName, encryptedKey, id,
	)
	return err
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
	database.DB.Exec(
		`INSERT INTO ai_requests_log (chart_id, provider_id, model, duration_ms, status, error_msg)
		 VALUES (NULLIF($1,''), NULLIF($2,''), $3, $4, $5, NULLIF($6,''))`,
		chartID, providerID, model, durationMs, status, errorMsg,
	)
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

