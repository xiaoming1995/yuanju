package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"yuanju/pkg/database"
)

type ActivityEvent struct {
	ID        string    `json:"id"`
	Type      string    `json:"event_type"`
	Channel   string    `json:"channel"`
	ObjectID  *string   `json:"object_id"`
	CreatedAt time.Time `json:"created_at"`
}

type AdminUserAuditEvent struct {
	ID        string    `json:"id"`
	AdminID   string    `json:"admin_id"`
	AdminName string    `json:"admin_name"`
	Action    string    `json:"action"`
	CreatedAt time.Time `json:"created_at"`
}

var activityTypes = map[string]bool{
	"register": true, "login": true, "account_linked": true,
	"chart_created": true, "compatibility_created": true,
	"bazi_report_generated": true, "compatibility_report_generated": true,
	"nickname_updated": true, "avatar_updated": true,
}
var activityChannels = map[string]bool{"web": true, "miniapp": true}

// RecordUserActivity serializes updates to one user's snapshot and event in one transaction.
func RecordUserActivity(ctx context.Context, userID, eventType, channel string, objectID *string) error {
	tx, err := database.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := RecordUserActivityTx(ctx, tx, userID, eventType, channel, objectID); err != nil {
		return err
	}
	return tx.Commit()
}

func RecordUserActivityTx(ctx context.Context, tx *sql.Tx, userID, eventType, channel string, objectID *string) error {
	if !activityTypes[eventType] || !activityChannels[channel] {
		return fmt.Errorf("invalid activity event: %s/%s", eventType, channel)
	}
	if eventType == "login" {
		result, err := tx.ExecContext(ctx, `UPDATE users SET last_login_at=NOW(), last_login_channel=$2, last_active_at=NOW() WHERE id=$1`, userID, channel)
		if err != nil {
			return err
		}
		n, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if n == 0 {
			return sql.ErrNoRows
		}
	} else {
		result, err := tx.ExecContext(ctx, `UPDATE users SET last_active_at=NOW() WHERE id=$1`, userID)
		if err != nil {
			return err
		}
		n, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if n == 0 {
			return sql.ErrNoRows
		}
	}
	_, err := tx.ExecContext(ctx, `INSERT INTO user_activity_events (user_id,event_type,channel,object_id) VALUES ($1,$2,$3,$4)`, userID, eventType, channel, objectID)
	return err
}

func ListUserActivity(userID string, page, limit int) ([]ActivityEvent, int, error) {
	if page < 1 || limit < 1 || limit > 100 {
		return nil, 0, errors.New("invalid pagination")
	}
	var total int
	if err := database.DB.QueryRow(`SELECT COUNT(*) FROM user_activity_events WHERE user_id=$1 AND created_at>=NOW()-INTERVAL '90 days'`, userID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := database.DB.Query(`SELECT id,event_type,channel,object_id::text,created_at FROM user_activity_events WHERE user_id=$1 AND created_at>=NOW()-INTERVAL '90 days' ORDER BY created_at DESC,id DESC LIMIT $2 OFFSET $3`, userID, limit, (page-1)*limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := []ActivityEvent{}
	for rows.Next() {
		var item ActivityEvent
		if err := rows.Scan(&item.ID, &item.Type, &item.Channel, &item.ObjectID, &item.CreatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func ListAdminUserAudit(userID string, page, limit int) ([]AdminUserAuditEvent, error) {
	if page < 1 || limit < 1 || limit > 100 {
		return nil, errors.New("invalid pagination")
	}
	rows, err := database.DB.Query(`SELECT e.id,e.admin_id,COALESCE(NULLIF(a.name,''),a.email,e.admin_id::text),e.action,e.created_at FROM admin_user_audit_events e LEFT JOIN admins a ON a.id=e.admin_id WHERE e.target_user_id=$1 AND e.created_at>=NOW()-INTERVAL '180 days' ORDER BY e.created_at DESC,e.id DESC LIMIT $2 OFFSET $3`, userID, limit, (page-1)*limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []AdminUserAuditEvent{}
	for rows.Next() {
		var item AdminUserAuditEvent
		if err := rows.Scan(&item.ID, &item.AdminID, &item.AdminName, &item.Action, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func CleanUserActivity(ctx context.Context, cutoff time.Time) (int64, error) {
	result, err := database.DB.ExecContext(ctx, `DELETE FROM user_activity_events WHERE created_at < $1`, cutoff)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
func CleanAdminUserAudit(ctx context.Context, cutoff time.Time) (int64, error) {
	result, err := database.DB.ExecContext(ctx, `DELETE FROM admin_user_audit_events WHERE created_at < $1`, cutoff)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
