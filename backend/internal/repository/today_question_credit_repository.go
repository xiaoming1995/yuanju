package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
	"yuanju/pkg/database"

	"github.com/lib/pq"
)

var ErrTodayQuestionCreditKeyConflict = errors.New("credit idempotency key belongs to another grant")

type TodayQuestionCreditEntry struct {
	ID           string    `json:"id"`
	Kind         string    `json:"kind"`
	Delta        int       `json:"delta"`
	BalanceAfter int       `json:"balance_after"`
	Reason       string    `json:"reason"`
	AdminID      string    `json:"admin_id,omitempty"`
	RequestID    string    `json:"request_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type TodayQuestionCreditAccount struct {
	Balance int                        `json:"balance"`
	Entries []TodayQuestionCreditEntry `json:"entries"`
}

func lockTodayQuestionUserTx(tx *sql.Tx, userID string) error {
	var lockResult any
	return tx.QueryRow(`SELECT pg_advisory_xact_lock(hashtext($1::text), 0)`, userID).Scan(&lockResult)
}

func todayQuestionCreditBalanceTx(tx *sql.Tx, userID string) (int, error) {
	var balance int
	err := tx.QueryRow(`SELECT COALESCE(SUM(delta),0)::int FROM today_question_credit_entries WHERE user_id=$1`, userID).Scan(&balance)
	if balance < 0 {
		return 0, fmt.Errorf("negative today question credit balance for user %s", userID)
	}
	return balance, err
}

func refundTodayQuestionCreditTx(tx *sql.Tx, userID, requestID string) error {
	var exists bool
	if err := tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM today_question_credit_entries WHERE request_id=$1 AND kind='refund')`, requestID).Scan(&exists); err != nil {
		return err
	}
	if exists {
		return nil
	}
	balance, err := todayQuestionCreditBalanceTx(tx, userID)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`INSERT INTO today_question_credit_entries
		(user_id,request_id,kind,delta,balance_after,reason)
		VALUES ($1,$2,'refund',1,$3,'AI 技术失败或处理超时退回')`, userID, requestID, balance+1)
	return err
}

func releaseStaleTodayQuestionRequestsTx(tx *sql.Tx, userID string) error {
	rows, err := tx.Query(`UPDATE today_question_requests SET status='error', quota_charged=FALSE,
		error_msg='处理超时，已退回今日问事次数', completed_at=NOW()
		WHERE user_id=$1 AND quota_charged AND status='processing'
		AND created_at < NOW() - INTERVAL '10 minutes' RETURNING id,quota_source`, userID)
	if err != nil {
		return err
	}
	creditIDs := []string{}
	for rows.Next() {
		var id, source string
		if err := rows.Scan(&id, &source); err != nil {
			rows.Close()
			return err
		}
		if source == "credit" {
			creditIDs = append(creditIDs, id)
		}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()
	for _, id := range creditIDs {
		if err := refundTodayQuestionCreditTx(tx, userID, id); err != nil {
			return err
		}
	}
	return nil
}

func GetTodayQuestionCreditAccount(userID string) (*TodayQuestionCreditAccount, error) {
	tx, err := database.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if err := lockTodayQuestionUserTx(tx, userID); err != nil {
		return nil, err
	}
	var exists bool
	if err := tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE id=$1)`, userID).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrAdminUserMissing
	}
	if err := releaseStaleTodayQuestionRequestsTx(tx, userID); err != nil {
		return nil, err
	}
	account := &TodayQuestionCreditAccount{Entries: []TodayQuestionCreditEntry{}}
	if err := tx.QueryRow(`SELECT COALESCE(SUM(delta),0)::int FROM today_question_credit_entries WHERE user_id=$1`, userID).Scan(&account.Balance); err != nil {
		return nil, err
	}
	rows, err := tx.Query(`SELECT id,kind,delta,balance_after,reason,COALESCE(admin_id::text,''),
		COALESCE(request_id::text,''),created_at FROM today_question_credit_entries
		WHERE user_id=$1 ORDER BY created_at DESC,id DESC LIMIT 20`, userID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var entry TodayQuestionCreditEntry
		if err := rows.Scan(&entry.ID, &entry.Kind, &entry.Delta, &entry.BalanceAfter, &entry.Reason, &entry.AdminID, &entry.RequestID, &entry.CreatedAt); err != nil {
			rows.Close()
			return nil, err
		}
		account.Entries = append(account.Entries, entry)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()
	return account, tx.Commit()
}

func GrantTodayQuestionCredits(userID, adminID string, amount int, reason, idempotencyKey string) (*TodayQuestionCreditEntry, error) {
	tx, err := database.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if err := lockTodayQuestionUserTx(tx, userID); err != nil {
		return nil, err
	}
	var exists bool
	if err := tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE id=$1)`, userID).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrAdminUserMissing
	}
	if err := releaseStaleTodayQuestionRequestsTx(tx, userID); err != nil {
		return nil, err
	}
	entry := &TodayQuestionCreditEntry{}
	var keyUser string
	err = tx.QueryRow(`SELECT id,user_id,delta,balance_after,reason,admin_id::text,created_at
		FROM today_question_credit_entries WHERE idempotency_key=$1`, idempotencyKey).Scan(
		&entry.ID, &keyUser, &entry.Delta, &entry.BalanceAfter, &entry.Reason, &entry.AdminID, &entry.CreatedAt)
	if err == nil {
		if keyUser != userID || entry.Delta != amount || entry.Reason != reason || entry.AdminID != adminID {
			return nil, ErrTodayQuestionCreditKeyConflict
		}
		entry.Kind = "grant"
		return entry, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}
	balance, err := todayQuestionCreditBalanceTx(tx, userID)
	if err != nil {
		return nil, err
	}
	entry.Kind, entry.Delta, entry.BalanceAfter, entry.Reason, entry.AdminID = "grant", amount, balance+amount, reason, adminID
	if err := tx.QueryRow(`INSERT INTO today_question_credit_entries
		(user_id,admin_id,kind,delta,balance_after,reason,idempotency_key)
		VALUES ($1,$2,'grant',$3,$4,$5,$6) RETURNING id,created_at`,
		userID, adminID, amount, entry.BalanceAfter, reason, idempotencyKey).Scan(&entry.ID, &entry.CreatedAt); err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.Constraint == "today_question_credit_entries_idempotency_key_key" {
			return nil, ErrTodayQuestionCreditKeyConflict
		}
		return nil, err
	}
	return entry, tx.Commit()
}
