package repository

import (
	"database/sql"
	"encoding/json"
	"yuanju/internal/model"
	"yuanju/pkg/database"
)

// ---- 用户 ----

func CreateUser(email, passwordHash, nickname string) (*model.User, error) {
	user := &model.User{}
	err := database.DB.QueryRow(
		`INSERT INTO users (email, password_hash, nickname) VALUES ($1, $2, $3)
		 RETURNING id, email, nickname, created_at`,
		email, passwordHash, nickname,
	).Scan(&user.ID, &user.Email, &user.Nickname, &user.CreatedAt)
	return user, err
}

func GetUserByEmail(email string) (*model.User, error) {
	user := &model.User{}
	err := database.DB.QueryRow(
		`SELECT id, email, password_hash, nickname, created_at FROM users WHERE email=$1`,
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Nickname, &user.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

func GetUserByID(id string) (*model.User, error) {
	user := &model.User{}
	err := database.DB.QueryRow(
		`SELECT id, email, nickname, created_at FROM users WHERE id=$1`, id,
	).Scan(&user.ID, &user.Email, &user.Nickname, &user.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

// ---- 八字命盘 ----

func CreateChart(chart *model.BaziChart) (*model.BaziChart, error) {
	wuxingJSON, _ := json.Marshal(chart.Wuxing)
	dayunJSON, _ := json.Marshal(chart.Dayun)

	result := &model.BaziChart{}
	var userID sql.NullString
	if chart.UserID != nil {
		userID = sql.NullString{String: *chart.UserID, Valid: true}
	}

	err := database.DB.QueryRow(`
		INSERT INTO bazi_charts
		(user_id, birth_year, birth_month, birth_day, birth_hour, gender,
		 year_gan, year_zhi, month_gan, month_zhi, day_gan, day_zhi, hour_gan, hour_zhi,
		 wuxing, dayun, yongshen, jishen, chart_hash, calendar_type, is_leap_month)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21)
		RETURNING id, chart_hash, created_at`,
		userID, chart.BirthYear, chart.BirthMonth, chart.BirthDay, chart.BirthHour, chart.Gender,
		chart.YearGan, chart.YearZhi, chart.MonthGan, chart.MonthZhi,
		chart.DayGan, chart.DayZhi, chart.HourGan, chart.HourZhi,
		wuxingJSON, dayunJSON, chart.Yongshen, chart.Jishen, chart.ChartHash,
		chart.CalendarType, chart.IsLeapMonth,
	).Scan(&result.ID, &result.ChartHash, &result.CreatedAt)

	if err != nil {
		return nil, err
	}
	result.YearGan = chart.YearGan
	result.YearZhi = chart.YearZhi
	result.MonthGan = chart.MonthGan
	result.MonthZhi = chart.MonthZhi
	result.DayGan = chart.DayGan
	result.DayZhi = chart.DayZhi
	result.HourGan = chart.HourGan
	result.HourZhi = chart.HourZhi
	result.Wuxing = chart.Wuxing
	result.Dayun = chart.Dayun
	result.Yongshen = chart.Yongshen
	result.Jishen = chart.Jishen
	result.BirthYear = chart.BirthYear
	result.BirthMonth = chart.BirthMonth
	result.BirthDay = chart.BirthDay
	result.BirthHour = chart.BirthHour
	result.Gender = chart.Gender
	result.CalendarType = chart.CalendarType
	result.IsLeapMonth = chart.IsLeapMonth
	return result, nil
}

func GetChartByHash(hash, userID string) (*model.BaziChart, error) {
	var id, chartHash string
	row := database.DB.QueryRow(
		`SELECT id, chart_hash FROM bazi_charts WHERE chart_hash=$1 AND user_id=$2`,
		hash, userID,
	)
	err := row.Scan(&id, &chartHash)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &model.BaziChart{ID: id, ChartHash: chartHash}, nil
}

func GetChartByID(id string) (*model.BaziChart, error) {
	chart := &model.BaziChart{}
	var wuxingJSON, dayunJSON []byte
	var userID sql.NullString

	err := database.DB.QueryRow(`
		SELECT id, user_id, birth_year, birth_month, birth_day, birth_hour, gender,
		       year_gan, year_zhi, month_gan, month_zhi, day_gan, day_zhi, hour_gan, hour_zhi,
		       wuxing, dayun, yongshen, jishen, chart_hash,
		       COALESCE(calendar_type, 'solar') AS calendar_type,
		       COALESCE(is_leap_month, false) AS is_leap_month,
		       created_at
		FROM bazi_charts WHERE id=$1`, id,
	).Scan(
		&chart.ID, &userID,
		&chart.BirthYear, &chart.BirthMonth, &chart.BirthDay, &chart.BirthHour, &chart.Gender,
		&chart.YearGan, &chart.YearZhi, &chart.MonthGan, &chart.MonthZhi,
		&chart.DayGan, &chart.DayZhi, &chart.HourGan, &chart.HourZhi,
		&wuxingJSON, &dayunJSON,
		&chart.Yongshen, &chart.Jishen, &chart.ChartHash,
		&chart.CalendarType, &chart.IsLeapMonth,
		&chart.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if userID.Valid {
		chart.UserID = &userID.String
	}
	json.Unmarshal(wuxingJSON, &chart.Wuxing)
	json.Unmarshal(dayunJSON, &chart.Dayun)
	return chart, nil
}

func GetChartsByUserID(userID string, limit, offset int) ([]*model.BaziChart, error) {
	rows, err := database.DB.Query(`
		SELECT id, birth_year, birth_month, birth_day, birth_hour, gender,
		       year_gan, year_zhi, month_gan, month_zhi, day_gan, day_zhi, hour_gan, hour_zhi,
		       yongshen, jishen, chart_hash,
		       COALESCE(calendar_type, 'solar') AS calendar_type,
		       COALESCE(is_leap_month, false) AS is_leap_month,
		       created_at
		FROM bazi_charts WHERE user_id=$1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var charts []*model.BaziChart
	for rows.Next() {
		c := &model.BaziChart{}
		if err := rows.Scan(
			&c.ID, &c.BirthYear, &c.BirthMonth, &c.BirthDay, &c.BirthHour, &c.Gender,
			&c.YearGan, &c.YearZhi, &c.MonthGan, &c.MonthZhi,
			&c.DayGan, &c.DayZhi, &c.HourGan, &c.HourZhi,
			&c.Yongshen, &c.Jishen, &c.ChartHash,
			&c.CalendarType, &c.IsLeapMonth,
			&c.CreatedAt,
		); err != nil {
			return nil, err
		}
		charts = append(charts, c)
	}
	return charts, nil
}

func UpdateChartYongshenJishen(chartID, yongshen, jishen string) error {
	_, err := database.DB.Exec(
		`UPDATE bazi_charts SET yongshen=$1, jishen=$2 WHERE id=$3`,
		yongshen, jishen, chartID,
	)
	return err
}

// ---- AI 报告 ----

func CreateReport(chartID, content, modelName string, contentStructured *json.RawMessage) (*model.AIReport, error) {
	report := &model.AIReport{}
	err := database.DB.QueryRow(
		`INSERT INTO ai_reports (chart_id, content, model, content_structured) VALUES ($1,$2,$3,$4)
		 RETURNING id, chart_id, content, model, created_at, content_structured`,
		chartID, content, modelName, contentStructured,
	).Scan(&report.ID, &report.ChartID, &report.Content, &report.Model, &report.CreatedAt, &report.ContentStructured)
	return report, err
}

func GetReportByChartID(chartID string) (*model.AIReport, error) {
	report := &model.AIReport{}
	err := database.DB.QueryRow(
		`SELECT id, chart_id, content, model, created_at, content_structured
		 FROM ai_reports WHERE chart_id=$1
		 ORDER BY created_at DESC LIMIT 1`, chartID,
	).Scan(&report.ID, &report.ChartID, &report.Content, &report.Model, &report.CreatedAt, &report.ContentStructured)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return report, err
}

// DeleteAllReports 清空所有 AI 报告缓存
func DeleteAllReports() (int64, error) {
	result, err := database.DB.Exec(`DELETE FROM ai_reports`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// DeleteReportByChartID 清除指定命盘的 AI 报告缓存
func DeleteReportByChartID(chartID string) (int64, error) {
	result, err := database.DB.Exec(`DELETE FROM ai_reports WHERE chart_id=$1`, chartID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
