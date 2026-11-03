package repository

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"yuanju/internal/model"
	"yuanju/pkg/database"

	"github.com/lib/pq"
)

var ErrWeChatIdentityConflict = errors.New("微信身份已关联其他账号")
var ErrWeChatMergeConflict = errors.New("当前账号无法绑定该网页账号")

func GetUserByWeChatIdentity(appID, openID string) (*model.User, error) {
	var userID string
	err := database.DB.QueryRow(`SELECT user_id FROM wechat_user_identities WHERE app_id=$1 AND open_id=$2`, appID, openID).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return GetUserByID(userID)
}

// CreateUserWithWeChatIdentity keeps the user and identity atomic. A concurrent
// login that wins the unique identity key is returned after this transaction rolls back.
func CreateUserWithWeChatIdentity(ctx context.Context, appID, openID string) (*model.User, bool, error) {
	number, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return nil, false, err
	}
	nickname := fmt.Sprintf("缘友%06d", number.Int64())
	tx, err := database.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, false, err
	}
	defer tx.Rollback()

	var userID string
	if err := tx.QueryRowContext(ctx,
		`INSERT INTO users (nickname, source) VALUES ($1, 'miniapp_wechat') RETURNING id`, nickname,
	).Scan(&userID); err != nil {
		return nil, false, err
	}
	var linkedID string
	err = tx.QueryRowContext(ctx,
		`INSERT INTO wechat_user_identities (app_id, open_id, user_id)
		 VALUES ($1, $2, $3) ON CONFLICT (app_id, open_id) DO NOTHING RETURNING user_id`,
		appID, openID, userID,
	).Scan(&linkedID)
	if errors.Is(err, sql.ErrNoRows) {
		if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
			return nil, false, rollbackErr
		}
		user, lookupErr := GetUserByWeChatIdentity(appID, openID)
		return user, false, lookupErr
	}
	if err != nil {
		return nil, false, err
	}
	if err := RecordUserActivityTx(ctx, tx, userID, "register", "miniapp", nil); err != nil {
		return nil, false, err
	}
	if err := tx.Commit(); err != nil {
		return nil, false, err
	}
	user, err := GetUserByID(userID)
	return user, true, err
}

func BindWeChatIdentity(ctx context.Context, userID, appID, openID string) error {
	res, err := database.DB.ExecContext(ctx,
		`INSERT INTO wechat_user_identities (app_id, open_id, user_id)
		 VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`, appID, openID, userID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 1 {
		return nil
	}
	var linkedID string
	err = database.DB.QueryRowContext(ctx,
		`SELECT user_id FROM wechat_user_identities WHERE app_id=$1 AND open_id=$2`, appID, openID,
	).Scan(&linkedID)
	if err != nil {
		return ErrWeChatIdentityConflict
	}
	if linkedID != userID {
		return ErrWeChatIdentityConflict
	}
	return nil
}

// MergeMiniappIntoWebUser moves the miniapp account's records to the verified
// web account and makes the web user the owner of the WeChat identity.
func MergeMiniappIntoWebUser(ctx context.Context, miniappID, webID, appID string) error {
	if miniappID == webID {
		return nil
	}
	tx, err := database.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Lock in a stable order so concurrent link attempts cannot merge the same
	// source into different accounts or deadlock on the two users.
	rows, err := tx.QueryContext(ctx,
		`SELECT id, email IS NULL, source, disabled_at IS NOT NULL FROM users
		 WHERE id IN ($1, $2) ORDER BY id FOR UPDATE`, miniappID, webID)
	if err != nil {
		return err
	}
	type account struct {
		emailMissing, disabled bool
		source                 string
	}
	accounts := map[string]account{}
	for rows.Next() {
		var id string
		var value account
		if err = rows.Scan(&id, &value.emailMissing, &value.source, &value.disabled); err != nil {
			break
		}
		accounts[id] = value
	}
	if err == nil {
		err = rows.Err()
	}
	rows.Close()
	if err != nil {
		return err
	}
	miniapp, miniappExists := accounts[miniappID]
	web, webExists := accounts[webID]
	if !miniappExists || !webExists || miniapp.disabled || web.disabled ||
		!miniapp.emailMissing || miniapp.source != "miniapp_wechat" || web.emailMissing {
		return ErrWeChatMergeConflict
	}

	var identityCount int
	if err := tx.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM wechat_user_identities WHERE app_id=$1 AND user_id=$2`,
		appID, miniappID).Scan(&identityCount); err != nil {
		return err
	}
	if identityCount != 1 {
		return ErrWeChatMergeConflict
	}
	if err := tx.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM wechat_user_identities WHERE app_id=$1 AND user_id=$2`,
		appID, webID).Scan(&identityCount); err != nil {
		return err
	}
	if identityCount != 0 {
		return ErrWeChatIdentityConflict
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE users AS web SET avatar_path = miniapp.avatar_path
		 FROM users AS miniapp WHERE web.id=$2 AND miniapp.id=$1
		 AND web.avatar_path='' AND miniapp.avatar_path<>''`, miniappID, webID); err != nil {
		return err
	}

	statements := []string{
		`UPDATE user_activity_events SET user_id=$2 WHERE user_id=$1`,
		`UPDATE user_chart_query_history SET user_id=$2 WHERE user_id=$1`,
		`UPDATE users AS web SET last_login_channel=CASE WHEN miniapp.last_login_at IS NOT NULL AND (web.last_login_at IS NULL OR miniapp.last_login_at > web.last_login_at) THEN miniapp.last_login_channel ELSE web.last_login_channel END, last_login_at=GREATEST(web.last_login_at, miniapp.last_login_at), last_active_at=GREATEST(web.last_active_at, miniapp.last_active_at), chart_query_count=web.chart_query_count+miniapp.chart_query_count FROM users AS miniapp WHERE web.id=$2 AND miniapp.id=$1`,
		`UPDATE bazi_charts SET user_id=$2 WHERE user_id=$1`,
		`UPDATE compatibility_readings SET user_id=$2 WHERE user_id=$1`,
		`UPDATE token_usage_logs SET user_id=$2 WHERE user_id=$1`,
		`UPDATE article_original_clicks SET user_id=$2 WHERE user_id=$1`,
		`INSERT INTO article_favorites (user_id, article_id, created_at)
		 SELECT $2, article_id, created_at FROM article_favorites WHERE user_id=$1
		 ON CONFLICT (user_id, article_id) DO NOTHING`,
		`INSERT INTO user_export_brand (user_id, title, footer_text, logo_path, watermark_mode, watermark_text, updated_at, logo_mode)
		 SELECT $2, title, footer_text, logo_path, watermark_mode, watermark_text, updated_at, logo_mode
		 FROM user_export_brand WHERE user_id=$1 ON CONFLICT (user_id) DO NOTHING`,
		`INSERT INTO token_usage_logs_monthly
		 (user_id, model, year_month, call_count, prompt_tokens, completion_tokens, reasoning_tokens, total_tokens, aggregated_at)
		 SELECT $2, model, year_month, call_count, prompt_tokens, completion_tokens, reasoning_tokens, total_tokens, aggregated_at
		 FROM token_usage_logs_monthly WHERE user_id=$1
		 ON CONFLICT (user_id, model, year_month) DO UPDATE SET
		 call_count=token_usage_logs_monthly.call_count+EXCLUDED.call_count,
		 prompt_tokens=token_usage_logs_monthly.prompt_tokens+EXCLUDED.prompt_tokens,
		 completion_tokens=token_usage_logs_monthly.completion_tokens+EXCLUDED.completion_tokens,
		 reasoning_tokens=token_usage_logs_monthly.reasoning_tokens+EXCLUDED.reasoning_tokens,
		 total_tokens=token_usage_logs_monthly.total_tokens+EXCLUDED.total_tokens,
		 aggregated_at=GREATEST(token_usage_logs_monthly.aggregated_at, EXCLUDED.aggregated_at)`,
	}
	for _, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement, miniappID, webID); err != nil {
			return err
		}
	}
	// The web account's explicit selection wins if both accounts have one.
	if _, err := tx.ExecContext(ctx, `INSERT INTO user_personal_charts (user_id, chart_id)
		SELECT $2, chart_id FROM user_personal_charts WHERE user_id=$1
		ON CONFLICT (user_id) DO NOTHING`, miniappID, webID); err != nil {
		return err
	}
	for _, table := range []string{"article_favorites", "token_usage_logs_monthly"} {
		if _, err := tx.ExecContext(ctx, `DELETE FROM `+table+` WHERE user_id=$1`, miniappID); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE wechat_user_identities SET user_id=$2 WHERE app_id=$3 AND user_id=$1`,
		miniappID, webID, appID); err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrWeChatIdentityConflict
		}
		return err
	}
	if err := RecordUserActivityTx(ctx, tx, webID, "account_linked", "miniapp", nil); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM users WHERE id=$1`, miniappID); err != nil {
		return err
	}
	return tx.Commit()
}
