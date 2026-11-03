package repository

import (
	"context"
	"database/sql"
	"errors"
	"yuanju/pkg/database"
)

var ErrProfileUserMissing = errors.New("用户不存在")

func UpdateUserNickname(ctx context.Context, userID, nickname string, channels ...string) error {
	tx, err := database.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx,
		`UPDATE users SET nickname=$2, updated_at=NOW() WHERE id=$1`, userID, nickname)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrProfileUserMissing
	}
	channel := "web"
	if len(channels) > 0 && channels[0] == "miniapp" {
		channel = "miniapp"
	}
	if err := RecordUserActivityTx(ctx, tx, userID, "nickname_updated", channel, nil); err != nil {
		return err
	}
	return tx.Commit()
}

// UpdateUserAvatar returns the previous generated path for file cleanup.
func UpdateUserAvatar(ctx context.Context, userID, avatarPath string, channels ...string) (string, error) {
	tx, err := database.DB.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	var previous string
	err = tx.QueryRowContext(ctx,
		`SELECT avatar_path FROM users WHERE id=$1 FOR UPDATE`, userID).Scan(&previous)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrProfileUserMissing
	}
	if err != nil {
		return "", err
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE users SET avatar_path=$2, updated_at=NOW() WHERE id=$1`, userID, avatarPath); err != nil {
		return "", err
	}
	channel := "web"
	if len(channels) > 0 && channels[0] == "miniapp" {
		channel = "miniapp"
	}
	if err := RecordUserActivityTx(ctx, tx, userID, "avatar_updated", channel, nil); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	return previous, nil
}
