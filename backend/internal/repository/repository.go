package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"
	"yuanju/internal/model"
	"yuanju/pkg/database"
)

// ---- 用户 ----

func CreateUser(email, passwordHash, nickname string) (*model.User, error) {
	return CreateUserWithSource(email, passwordHash, nickname, "self_registered")
}

func CreateUserWithSource(email, passwordHash, nickname, source string) (*model.User, error) {
	user := &model.User{}
	err := database.DB.QueryRow(
		`INSERT INTO users (email, password_hash, nickname, source) VALUES ($1, $2, $3, $4)
		 RETURNING id, email, nickname, source, created_at`,
		email, passwordHash, nickname, source,
	).Scan(&user.ID, &user.Email, &user.Nickname, &user.Source, &user.CreatedAt)
	return user, err
}

func CreateRegisteredUser(email, passwordHash, nickname string) (*model.User, error) {
	tx, err := database.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	user := &model.User{}
	err = tx.QueryRow(`INSERT INTO users (email,password_hash,nickname,source) VALUES ($1,$2,$3,'self_registered') RETURNING id,email,nickname,source,created_at`, email, passwordHash, nickname).Scan(&user.ID, &user.Email, &user.Nickname, &user.Source, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	if err = RecordUserActivityTx(context.Background(), tx, user.ID, "register", "web", nil); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return user, nil
}

func GetUserByEmail(email string) (*model.User, error) {
	user := &model.User{}
	err := database.DB.QueryRow(
		`SELECT id, email, password_hash, nickname,
		 COALESCE('/static/uploads/' || NULLIF(avatar_path, ''), ''),
		 COALESCE(source, 'self_registered'), created_at, disabled_at FROM users WHERE email=$1`,
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Nickname, &user.AvatarURL, &user.Source, &user.CreatedAt, &user.DisabledAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

func GetUserByID(id string) (*model.User, error) {
	user := &model.User{}
	err := database.DB.QueryRow(
		`SELECT id, COALESCE(email, ''), COALESCE(nickname, ''),
		 COALESCE('/static/uploads/' || NULLIF(avatar_path, ''), ''),
		 COALESCE(source, 'self_registered'), created_at, disabled_at FROM users WHERE id=$1`, id,
	).Scan(&user.ID, &user.Email, &user.Nickname, &user.AvatarURL, &user.Source, &user.CreatedAt, &user.DisabledAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

func UpdateUserPassword(userID, passwordHash string) error {
	_, err := database.DB.Exec(`UPDATE users SET password_hash=$1 WHERE id=$2`, passwordHash, userID)
	return err
}

func SetUserDisabled(userID string, disabled bool) error {
	if disabled {
		_, err := database.DB.Exec(`UPDATE users SET disabled_at=NOW() WHERE id=$1`, userID)
		return err
	}
	_, err := database.DB.Exec(`UPDATE users SET disabled_at=NULL WHERE id=$1`, userID)
	return err
}

func DeleteUser(userID string) error {
	_, err := database.DB.Exec(`DELETE FROM users WHERE id=$1`, userID)
	return err
}

// ---- 八字干支 ----

func CreateChart(chart *model.BaziChart) (*model.BaziChart, error) {
	return CreateChartWithChannel(chart, "web")
}

func CreateChartWithChannel(chart *model.BaziChart, channel string) (*model.BaziChart, error) {
	return createChartWithChannel(chart, channel, false)
}

// CreatePersonalChartWithChannel saves a new chart and switches the user's
// personal binding in one transaction. An error leaves the old binding intact.
func CreatePersonalChartWithChannel(chart *model.BaziChart, channel string) (*model.BaziChart, error) {
	if chart == nil || chart.UserID == nil || *chart.UserID == "" {
		return nil, errors.New("本人命盘必须属于已登录用户")
	}
	return createChartWithChannel(chart, channel, true)
}

func createChartWithChannel(chart *model.BaziChart, channel string, bindPersonal bool) (*model.BaziChart, error) {
	wuxingJSON, _ := json.Marshal(chart.Wuxing)
	dayunJSON, _ := json.Marshal(chart.Dayun)
	tx, err := database.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if bindPersonal {
		var owner string
		if err := tx.QueryRow(`SELECT id FROM users WHERE id=$1 FOR UPDATE`, *chart.UserID).Scan(&owner); err != nil {
			return nil, err
		}
	}

	result := &model.BaziChart{}
	var userID sql.NullString
	if chart.UserID != nil {
		userID = sql.NullString{String: *chart.UserID, Valid: true}
	}

	err = tx.QueryRow(`
		INSERT INTO bazi_charts
		(user_id, birth_year, birth_month, birth_day, birth_hour, gender, display_name,
		 year_gan, year_zhi, month_gan, month_zhi, day_gan, day_zhi, hour_gan, hour_zhi,
		 wuxing, dayun, yongshen, jishen, chart_hash, calendar_type, is_leap_month)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22)
		RETURNING id, chart_hash, created_at`,
		userID, chart.BirthYear, chart.BirthMonth, chart.BirthDay, chart.BirthHour, chart.Gender, chart.DisplayName,
		chart.YearGan, chart.YearZhi, chart.MonthGan, chart.MonthZhi,
		chart.DayGan, chart.DayZhi, chart.HourGan, chart.HourZhi,
		wuxingJSON, dayunJSON, chart.Yongshen, chart.Jishen, chart.ChartHash,
		chart.CalendarType, chart.IsLeapMonth,
	).Scan(&result.ID, &result.ChartHash, &result.CreatedAt)

	if err != nil {
		return nil, err
	}
	if chart.UserID != nil {
		if _, err := tx.Exec(`INSERT INTO user_chart_query_history (chart_id,user_id,queried_at) VALUES ($1,$2,$3)`, result.ID, *chart.UserID, result.CreatedAt); err != nil {
			return nil, err
		}
		if _, err := tx.Exec(`UPDATE users SET chart_query_count=chart_query_count+1 WHERE id=$1`, *chart.UserID); err != nil {
			return nil, err
		}
		if err := RecordUserActivityTx(context.Background(), tx, *chart.UserID, "chart_created", channel, &result.ID); err != nil {
			return nil, err
		}
		if bindPersonal {
			if _, err := tx.Exec(`INSERT INTO user_personal_charts (user_id, chart_id) VALUES ($1,$2)
				ON CONFLICT (user_id) DO UPDATE SET chart_id=EXCLUDED.chart_id, updated_at=NOW()`, *chart.UserID, result.ID); err != nil {
				return nil, err
			}
		}
	}
	if err := tx.Commit(); err != nil {
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
	result.DisplayName = chart.DisplayName
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
		       COALESCE(display_name, '') AS display_name,
		       year_gan, year_zhi, month_gan, month_zhi, day_gan, day_zhi, hour_gan, hour_zhi,
		       wuxing, dayun, yongshen, jishen, chart_hash,
		       COALESCE(calendar_type, 'solar') AS calendar_type,
		       COALESCE(is_leap_month, false) AS is_leap_month,
		       created_at
		FROM bazi_charts WHERE id=$1`, id,
	).Scan(
		&chart.ID, &userID,
		&chart.BirthYear, &chart.BirthMonth, &chart.BirthDay, &chart.BirthHour, &chart.Gender,
		&chart.DisplayName,
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
		       COALESCE(display_name, '') AS display_name,
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
			&c.DisplayName,
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

// CountChartsGenderByUserID 返回用户干支总数及男女命分布（供历史页分页与统计卡使用）
// 注：CountChartsByUserID（user_profile_repository.go）只返回总数，这里多了性别聚合
func CountChartsGenderByUserID(userID string) (total, male, female int, err error) {
	err = database.DB.QueryRow(`
		SELECT COUNT(*),
		       COUNT(*) FILTER (WHERE gender='male'),
		       COUNT(*) FILTER (WHERE gender='female')
		FROM bazi_charts WHERE user_id=$1`,
		userID,
	).Scan(&total, &male, &female)
	return
}

func UpdateChartDisplayName(chartID, displayName string) error {
	_, err := database.DB.Exec(
		`UPDATE bazi_charts SET display_name=$1 WHERE id=$2`,
		displayName, chartID,
	)
	return err
}

func UpdateChartYongshenJishen(chartID, yongshen, jishen string) error {
	_, err := database.DB.Exec(
		`UPDATE bazi_charts SET yongshen=$1, jishen=$2 WHERE id=$3`,
		yongshen, jishen, chartID,
	)
	return err
}

// GetChartResultJSON 读取干支的完整 BaziResult 快照（result_json）
// 返回 nil 表示该干支尚未存快照（旧干支需懒加载补回）
func GetChartResultJSON(chartID string) ([]byte, error) {
	var raw sql.NullString
	err := database.DB.QueryRow(
		`SELECT result_json::text FROM bazi_charts WHERE id=$1`, chartID,
	).Scan(&raw)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if !raw.Valid || raw.String == "" || raw.String == "null" {
		return nil, nil
	}
	return []byte(raw.String), nil
}

// SaveChartResultJSON 写入干支的 BaziResult 快照（懒加载或新建干支时调用）
func SaveChartResultJSON(chartID string, resultJSON []byte) error {
	_, err := database.DB.Exec(
		`UPDATE bazi_charts SET result_json=$1 WHERE id=$2`,
		string(resultJSON), chartID,
	)
	return err
}

// DeleteChart 删除干支及其级联内容；独立查询历史仅留下时间和已删除状态。
// 计费日志由 ON DELETE SET NULL 保留。返回删除条数。
func DeleteChart(chartID string) (int64, error) {
	result, err := database.DB.Exec(`DELETE FROM bazi_charts WHERE id=$1`, chartID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// ---- AI 报告 ----

func CreateReport(chartID, content, modelName string, contentStructured *json.RawMessage) (*model.AIReport, error) {
	return CreateReportWithChannel(chartID, content, modelName, contentStructured, "web")
}

func CreateReportWithChannel(chartID, content, modelName string, contentStructured *json.RawMessage, channel string) (*model.AIReport, error) {
	tx, err := database.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	report := &model.AIReport{}
	err = tx.QueryRow(
		`INSERT INTO ai_reports (chart_id, content, model, content_structured, algorithm_version)
		 VALUES ($1,$2,$3,$4,$5)
		 RETURNING id, chart_id, content, model, created_at, content_structured`,
		chartID, content, modelName, contentStructured, CurrentAlgorithmVersion,
	).Scan(&report.ID, &report.ChartID, &report.Content, &report.Model, &report.CreatedAt, &report.ContentStructured)
	if err != nil {
		return nil, err
	}
	var userID sql.NullString
	if err := tx.QueryRow(`SELECT user_id FROM bazi_charts WHERE id=$1`, chartID).Scan(&userID); err != nil {
		return nil, err
	}
	if userID.Valid {
		if err := RecordUserActivityTx(context.Background(), tx, userID.String, "bazi_report_generated", channel, &report.ID); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return report, nil
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

// DeleteReportByChartID 清除指定干支的 AI 报告缓存
func DeleteReportByChartID(chartID string) (int64, error) {
	result, err := database.DB.Exec(`DELETE FROM ai_reports WHERE chart_id=$1`, chartID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// DeleteAIReportsOlderThan 删除 created_at 早于 cutoff 的 ai_reports 行。
// 返回删除条数。
func DeleteAIReportsOlderThan(cutoff time.Time) (int64, error) {
	res, err := database.DB.Exec(`DELETE FROM ai_reports WHERE created_at < $1`, cutoff)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
