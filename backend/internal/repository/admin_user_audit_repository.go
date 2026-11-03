package repository

import (
	"context"
	"database/sql"
	"errors"

	"yuanju/internal/model"
	"yuanju/pkg/database"
)

var ErrAdminUserMissing = errors.New("用户不存在")

func CreateUserByAdminAudited(ctx context.Context, adminID, email, hash, nickname string) (*model.User, error) {
	tx, err := database.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	user := &model.User{}
	err = tx.QueryRowContext(ctx, `INSERT INTO users (email,password_hash,nickname,source) VALUES ($1,$2,$3,'admin_created') RETURNING id,email,nickname,source,created_at`, email, hash, nickname).Scan(&user.ID, &user.Email, &user.Nickname, &user.Source, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	if err := recordAdminUserAuditTx(ctx, tx, adminID, user.ID, "created"); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return user, nil
}

func SetUserDisabledAudited(ctx context.Context, adminID, userID string, disabled bool) error {
	tx, err := database.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `UPDATE users SET disabled_at=CASE WHEN $2 THEN NOW() ELSE NULL END WHERE id=$1 AND (disabled_at IS NULL)=$2`, userID, disabled)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		var exists bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id=$1)`, userID).Scan(&exists); err != nil {
			return err
		}
		if exists {
			return nil
		}
		return ErrAdminUserMissing
	}
	action := "enabled"
	if disabled {
		action = "disabled"
	}
	if err := recordAdminUserAuditTx(ctx, tx, adminID, userID, action); err != nil {
		return err
	}
	return tx.Commit()
}

func ResetUserPasswordAudited(ctx context.Context, adminID, userID, hash string) error {
	tx, err := database.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `UPDATE users SET password_hash=$2 WHERE id=$1 AND email IS NOT NULL`, userID, hash)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrAdminUserMissing
	}
	if err := recordAdminUserAuditTx(ctx, tx, adminID, userID, "password_reset"); err != nil {
		return err
	}
	return tx.Commit()
}

func DeleteUserAudited(ctx context.Context, adminID, userID string) error {
	tx, err := database.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `DELETE FROM users WHERE id=$1`, userID)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrAdminUserMissing
	}
	if err := recordAdminUserAuditTx(ctx, tx, adminID, userID, "deleted"); err != nil {
		return err
	}
	return tx.Commit()
}

func recordAdminUserAuditTx(ctx context.Context, tx *sql.Tx, adminID, userID, action string) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO admin_user_audit_events (admin_id,target_user_id,action) VALUES ($1,$2,$3)`, adminID, userID, action)
	return err
}

type AdminUserDetail struct {
	AdminUserRow
	AIRequestsThisMonth int `json:"ai_requests_this_month"`
}

func GetAdminUserDetail(id string) (*AdminUserDetail, error) {
	detail := &AdminUserDetail{}
	err := database.DB.QueryRow(`SELECT u.id,COALESCE(u.email,''),COALESCE(u.nickname,''),COALESCE(u.source,'self_registered'),u.created_at,u.disabled_at,u.last_login_at,COALESCE(u.last_login_channel,''),u.last_active_at,
	    (SELECT COUNT(*) FROM bazi_charts WHERE user_id=u.id),u.chart_query_count,(SELECT COUNT(*) FROM compatibility_readings WHERE user_id=u.id),
        (SELECT COUNT(*) FROM token_usage_logs WHERE user_id=u.id AND created_at>=DATE_TRUNC('month',NOW()))
	    FROM users u WHERE u.id=$1`, id).Scan(&detail.ID, &detail.Email, &detail.Nickname, &detail.Source, &detail.CreatedAt, &detail.DisabledAt, &detail.LastLoginAt, &detail.LastLoginChannel, &detail.LastActiveAt, &detail.ChartCount, &detail.ChartQueryCount, &detail.CompatCount, &detail.AIRequestsThisMonth)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return detail, err
}
