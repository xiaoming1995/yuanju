package repository

import (
	"database/sql"
	"encoding/json"
	"yuanju/internal/model"
	"yuanju/pkg/database"
)

func CountChartsByUserID(userID string) (int, error) {
	var count int
	err := database.DB.QueryRow(`SELECT COUNT(*) FROM bazi_charts WHERE user_id = $1`, userID).Scan(&count)
	return count, err
}

func CountAIReportsByUserID(userID string) (int, error) {
	var count int
	err := database.DB.QueryRow(`
		SELECT COUNT(ar.id)
		FROM ai_reports ar
		JOIN bazi_charts bc ON bc.id = ar.chart_id
		WHERE bc.user_id = $1`, userID).Scan(&count)
	return count, err
}

func CountCompatibilityReadingsByUserID(userID string) (int, error) {
	var count int
	err := database.DB.QueryRow(`SELECT COUNT(*) FROM compatibility_readings WHERE user_id = $1`, userID).Scan(&count)
	return count, err
}

func ListRecentChartsForProfile(userID string, limit int) ([]model.UserProfileChartSummary, error) {
	rows, err := database.DB.Query(`
		SELECT id, birth_year, birth_month, birth_day, birth_hour, gender,
		       year_gan, year_zhi, month_gan, month_zhi, day_gan, day_zhi, hour_gan, hour_zhi,
		       COALESCE(yongshen, ''), COALESCE(jishen, ''), created_at
		FROM bazi_charts
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []model.UserProfileChartSummary{}
	for rows.Next() {
		var item model.UserProfileChartSummary
		if err := rows.Scan(
			&item.ID, &item.BirthYear, &item.BirthMonth, &item.BirthDay, &item.BirthHour, &item.Gender,
			&item.YearGan, &item.YearZhi, &item.MonthGan, &item.MonthZhi, &item.DayGan, &item.DayZhi, &item.HourGan, &item.HourZhi,
			&item.Yongshen, &item.Jishen, &item.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func ListRecentCompatibilityForProfile(userID string, limit int) ([]model.UserProfileCompatibilitySummary, error) {
	rows, err := database.DB.Query(`
		SELECT id, overall_level, summary_tags, created_at
		FROM compatibility_readings
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []model.UserProfileCompatibilitySummary{}
	for rows.Next() {
		var item model.UserProfileCompatibilitySummary
		var rawTags []byte
		if err := rows.Scan(&item.ID, &item.OverallLevel, &rawTags, &item.CreatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(rawTags, &item.SummaryTags)
		selfName, partnerName, err := compatibilityParticipantNames(item.ID)
		if err != nil {
			return nil, err
		}
		item.SelfName = selfName
		item.PartnerName = partnerName
		out = append(out, item)
	}
	return out, rows.Err()
}

func compatibilityParticipantNames(readingID string) (string, string, error) {
	rows, err := database.DB.Query(
		`SELECT role, display_name FROM compatibility_participants WHERE reading_id = $1`,
		readingID,
	)
	if err == sql.ErrNoRows {
		return "", "", nil
	}
	if err != nil {
		return "", "", err
	}
	defer rows.Close()

	selfName, partnerName := "", ""
	for rows.Next() {
		var role, displayName string
		if err := rows.Scan(&role, &displayName); err != nil {
			return "", "", err
		}
		switch role {
		case "self":
			selfName = displayName
		case "partner":
			partnerName = displayName
		}
	}
	return selfName, partnerName, rows.Err()
}
