package repository

import (
	"database/sql"
	"errors"
	"yuanju/internal/model"
	"yuanju/pkg/database"
)

var ErrPersonalChartNotOwned = errors.New("干支不存在或不属于当前账号")

func GetPersonalChart(userID string) (*model.BaziChart, error) {
	var chartID string
	err := database.DB.QueryRow(`SELECT p.chart_id FROM user_personal_charts p
        JOIN bazi_charts c ON c.id=p.chart_id AND c.user_id=p.user_id
        WHERE p.user_id=$1`, userID).Scan(&chartID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return GetChartByID(chartID)
}

func SetPersonalChart(userID, chartID string) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	// Match the account-merge lock order: user first, then chart.
	var owner string
	if err := tx.QueryRow(`SELECT id FROM users WHERE id=$1 FOR UPDATE`, userID).Scan(&owner); err != nil {
		return err
	}
	var selected string
	if err := tx.QueryRow(`SELECT id FROM bazi_charts WHERE id=$1 AND user_id=$2 FOR UPDATE`, chartID, userID).Scan(&selected); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrPersonalChartNotOwned
		}
		return err
	}
	if _, err := tx.Exec(`INSERT INTO user_personal_charts (user_id, chart_id) VALUES ($1,$2)
		ON CONFLICT (user_id) DO UPDATE SET chart_id=EXCLUDED.chart_id, updated_at=NOW()`, userID, chartID); err != nil {
		return err
	}
	return tx.Commit()
}

func ClearPersonalChart(userID string) error {
	_, err := database.DB.Exec(`DELETE FROM user_personal_charts WHERE user_id=$1`, userID)
	return err
}
