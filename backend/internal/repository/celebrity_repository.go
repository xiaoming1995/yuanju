package repository

import (
	"log"
	"yuanju/internal/model"
	"yuanju/pkg/database"
)

// CreateCelebrity 创建新名人记录
func CreateCelebrity(name, gender, traits, career string, active bool) (*model.CelebrityRecord, error) {
	c := &model.CelebrityRecord{}
	err := database.DB.QueryRow(
		`INSERT INTO celebrity_records (name, gender, traits, career, active)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, name, gender, traits, career, active, created_at, updated_at`,
		name, gender, traits, career, active,
	).Scan(&c.ID, &c.Name, &c.Gender, &c.Traits, &c.Career, &c.Active, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		log.Printf("[CreateCelebrity] Error: %v", err)
	}
	return c, err
}

// ListCelebrities 获取所有的名人记录
func ListCelebrities(onlyActive bool) ([]model.CelebrityRecord, error) {
	query := `SELECT id, name, gender, traits, career, active, created_at, updated_at
	          FROM celebrity_records`
	if onlyActive {
		query += ` WHERE active = true`
	}
	query += ` ORDER BY created_at DESC`

	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var celebs []model.CelebrityRecord
	for rows.Next() {
		var c model.CelebrityRecord
		if err := rows.Scan(
			&c.ID, &c.Name, &c.Gender, &c.Traits, &c.Career,
			&c.Active, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, err
		}
		celebs = append(celebs, c)
	}
	return celebs, nil
}

// UpdateCelebrity 更新名人记录
func UpdateCelebrity(id, name, gender, traits, career string, active bool) error {
	_, err := database.DB.Exec(
		`UPDATE celebrity_records
		 SET name=$1, gender=$2, traits=$3, career=$4, active=$5, updated_at=NOW()
		 WHERE id=$6`,
		name, gender, traits, career, active, id,
	)
	return err
}

// DeleteCelebrity 删除名人记录
func DeleteCelebrity(id string) error {
	_, err := database.DB.Exec(`DELETE FROM celebrity_records WHERE id=$1`, id)
	return err
}
