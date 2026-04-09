package repository

import (
	"yuanju/pkg/database"
)

// AlgoConfigRow 算法参数表记录
type AlgoConfigRow struct {
	Key         string
	Value       string
	Description string
}

// GetAllAlgoConfig 读取全部算法参数
func GetAllAlgoConfig() ([]AlgoConfigRow, error) {
	rows, err := database.DB.Query(
		`SELECT key, value, COALESCE(description, '') FROM algo_config ORDER BY key`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []AlgoConfigRow
	for rows.Next() {
		var r AlgoConfigRow
		if err := rows.Scan(&r.Key, &r.Value, &r.Description); err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}

// UpsertAlgoConfig 写入或更新单条算法参数
func UpsertAlgoConfig(key, value, description string) error {
	_, err := database.DB.Exec(
		`INSERT INTO algo_config (key, value, description)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (key) DO UPDATE
		 SET value = EXCLUDED.value,
		     description = EXCLUDED.description,
		     updated_at = NOW()`,
		key, value, description,
	)
	return err
}
