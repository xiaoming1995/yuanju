package repository

import (
	"database/sql"
	"fmt"
	"strconv"
	"yuanju/pkg/database"
)

const (
	SettingRegistrationEnabled        = "registration_enabled"
	SettingMiniappRegistrationEnabled = "miniapp_registration_enabled"
	SettingArticlesModuleEnabled      = "articles_module_enabled"
	SettingTodayQuestionDailyLimit    = "today_question_daily_limit"
)

const MaxTodayQuestionDailyLimit = 100

func GetTodayQuestionDailyLimit() (int, error) {
	var raw string
	err := database.DB.QueryRow(`SELECT value FROM system_settings WHERE key=$1`, SettingTodayQuestionDailyLimit).Scan(&raw)
	if err == sql.ErrNoRows {
		return 1, nil
	}
	if err != nil {
		return 0, err
	}
	limit, err := strconv.Atoi(raw)
	if err != nil || limit < 1 || limit > MaxTodayQuestionDailyLimit {
		return 0, fmt.Errorf("invalid today question daily limit %q", raw)
	}
	return limit, nil
}

func SetTodayQuestionDailyLimit(limit int) error {
	if limit < 1 || limit > MaxTodayQuestionDailyLimit {
		return fmt.Errorf("today question daily limit out of range: %d", limit)
	}
	_, err := database.DB.Exec(`INSERT INTO system_settings (key, value, updated_at)
		VALUES ($1,$2,NOW()) ON CONFLICT (key) DO UPDATE SET value=EXCLUDED.value, updated_at=NOW()`,
		SettingTodayQuestionDailyLimit, strconv.Itoa(limit))
	return err
}

func GetBoolSetting(key string, defaultValue bool) (bool, error) {
	var raw string
	err := database.DB.QueryRow(`SELECT value FROM system_settings WHERE key=$1`, key).Scan(&raw)
	if err == sql.ErrNoRows {
		return defaultValue, nil
	}
	if err != nil {
		return defaultValue, err
	}
	switch raw {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return defaultValue, fmt.Errorf("invalid bool setting %s=%q", key, raw)
	}
}

func IsArticleModuleEnabled() (bool, error) {
	return GetBoolSetting(SettingArticlesModuleEnabled, true)
}

func SetBoolSetting(key string, value bool) error {
	raw := "false"
	if value {
		raw = "true"
	}
	_, err := database.DB.Exec(
		`INSERT INTO system_settings (key, value, updated_at)
		 VALUES ($1, $2, NOW())
		 ON CONFLICT (key) DO UPDATE SET value=EXCLUDED.value, updated_at=NOW()`,
		key, raw,
	)
	return err
}
