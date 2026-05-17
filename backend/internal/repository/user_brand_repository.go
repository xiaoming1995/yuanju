package repository

import (
	"database/sql"
	"yuanju/internal/model"
	"yuanju/pkg/database"
)

// GetExportBrand returns the user's brand row, or a zero-value struct
// with empty strings and WatermarkMode="none" if no row exists.
func GetExportBrand(userID string) (model.ExportBrand, error) {
	var b model.ExportBrand
	b.UserID = userID
	err := database.DB.QueryRow(`
		SELECT title, footer_text, logo_path, watermark_mode, watermark_text, updated_at
		FROM user_export_brand WHERE user_id = $1`, userID).Scan(
		&b.Title, &b.FooterText, &b.LogoPath, &b.WatermarkMode, &b.WatermarkText, &b.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		b.WatermarkMode = "none"
		return b, nil
	}
	return b, err
}

// UpsertExportBrandText writes title/footer/watermark fields. Does NOT touch logo_path.
func UpsertExportBrandText(userID, title, footerText, watermarkMode, watermarkText string) error {
	_, err := database.DB.Exec(`
		INSERT INTO user_export_brand (user_id, title, footer_text, watermark_mode, watermark_text)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id) DO UPDATE SET
			title = EXCLUDED.title,
			footer_text = EXCLUDED.footer_text,
			watermark_mode = EXCLUDED.watermark_mode,
			watermark_text = EXCLUDED.watermark_text,
			updated_at = NOW()`,
		userID, title, footerText, watermarkMode, watermarkText)
	return err
}

// UpdateExportBrandLogo sets logo_path, creating the row if it doesn't exist.
// Returns the previous logo_path (empty if none) so caller can delete the old file.
// Uses an explicit two-statement transaction for clarity over a RETURNING trick.
func UpdateExportBrandLogo(userID, logoPath string) (oldPath string, err error) {
	tx, err := database.DB.Begin()
	if err != nil {
		return "", err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	err = tx.QueryRow(`SELECT logo_path FROM user_export_brand WHERE user_id = $1`, userID).Scan(&oldPath)
	if err == sql.ErrNoRows {
		oldPath = ""
		err = nil
	} else if err != nil {
		return "", err
	}

	if _, err = tx.Exec(`
		INSERT INTO user_export_brand (user_id, logo_path)
		VALUES ($1, $2)
		ON CONFLICT (user_id) DO UPDATE SET
			logo_path = EXCLUDED.logo_path,
			updated_at = NOW()`,
		userID, logoPath); err != nil {
		return "", err
	}

	if err = tx.Commit(); err != nil {
		return "", err
	}
	return oldPath, nil
}

// ClearExportBrandLogo sets logo_path to empty, returns previous value for file cleanup.
func ClearExportBrandLogo(userID string) (oldPath string, err error) {
	tx, err := database.DB.Begin()
	if err != nil {
		return "", err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	err = tx.QueryRow(`SELECT logo_path FROM user_export_brand WHERE user_id = $1`, userID).Scan(&oldPath)
	if err == sql.ErrNoRows {
		oldPath = ""
		err = nil
		// nothing to clear
		if cerr := tx.Commit(); cerr != nil {
			return "", cerr
		}
		return "", nil
	} else if err != nil {
		return "", err
	}

	if _, err = tx.Exec(`UPDATE user_export_brand SET logo_path = '', updated_at = NOW() WHERE user_id = $1`, userID); err != nil {
		return "", err
	}

	if err = tx.Commit(); err != nil {
		return "", err
	}
	return oldPath, nil
}
