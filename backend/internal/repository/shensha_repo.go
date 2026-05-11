package repository

import (
	"database/sql"
	"yuanju/internal/model"
)

// GetAllShenshaAnnotations 返回全部神煞注解（按名称排序）
func GetAllShenshaAnnotations(db *sql.DB) ([]model.ShenshaAnnotation, error) {
	rows, err := db.Query(`
		SELECT id, name, polarity, category, short_desc, description, updated_at
		FROM shensha_annotations
		ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []model.ShenshaAnnotation
	for rows.Next() {
		var a model.ShenshaAnnotation
		if err := rows.Scan(&a.ID, &a.Name, &a.Polarity, &a.Category, &a.ShortDesc, &a.Description, &a.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, a)
	}
	return result, rows.Err()
}

// UpdateShenshaAnnotation 更新神煞注解（Admin 专用），支持 category / short_desc / description
func UpdateShenshaAnnotation(db *sql.DB, name, category, shortDesc, description string) error {
	result, err := db.Exec(`
		UPDATE shensha_annotations
		SET category = $1, short_desc = $2, description = $3, updated_at = NOW()
		WHERE name = $4
	`, category, shortDesc, description, name)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}
