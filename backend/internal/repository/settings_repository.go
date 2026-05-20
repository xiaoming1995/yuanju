package repository

import (
	"database/sql"
	"fmt"
	"yuanju/pkg/database"
)

const SettingRegistrationEnabled = "registration_enabled"

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
