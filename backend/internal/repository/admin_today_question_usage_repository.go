package repository

import (
	"fmt"
	"sort"
	"yuanju/pkg/database"

	"github.com/lib/pq"
)

type AdminTodayQuestionUsage struct {
	Date           string `json:"date"`
	DailyLimit     int    `json:"daily_limit"`
	Used           int    `json:"used"`
	FreeUsed       int    `json:"free_used"`
	Processing     int    `json:"processing"`
	FreeProcessing int    `json:"free_processing"`
	CreditBalance  int    `json:"credit_balance"`
	Remaining      int    `json:"remaining"`
}

// Only users on the current page are included in these aggregates.
func populateAdminTodayQuestionUsage(users []AdminUserRow) error {
	if len(users) == 0 {
		return nil
	}
	ids := make([]string, 0, len(users))
	byID := make(map[string]*AdminUserRow, len(users))
	for i := range users {
		ids = append(ids, users[i].ID)
		byID[users[i].ID] = &users[i]
	}
	if err := releaseStaleTodayQuestionsForAdminPage(ids); err != nil {
		return err
	}
	limit, err := GetTodayQuestionDailyLimit()
	if err != nil {
		return err
	}
	rows, err := database.DB.Query(`WITH page_users AS (
		SELECT unnest($1::uuid[]) AS id
	), request_usage AS (
		SELECT user_id,
			COUNT(*) FILTER (WHERE quota_charged AND status IN ('success','insufficient')) AS used,
			COUNT(*) FILTER (WHERE quota_charged AND quota_source='daily' AND status IN ('success','insufficient')) AS free_used,
			COUNT(*) FILTER (WHERE quota_charged AND status='processing') AS processing,
			COUNT(*) FILTER (WHERE quota_charged AND quota_source='daily' AND status='processing') AS free_processing
		FROM today_question_requests
		WHERE user_id=ANY($1::uuid[])
			AND question_date=(NOW() AT TIME ZONE 'Asia/Shanghai')::date
		GROUP BY user_id
	), credits AS (
		SELECT user_id, SUM(delta)::int AS balance
		FROM today_question_credit_entries
		WHERE user_id=ANY($1::uuid[])
		GROUP BY user_id
	)
	SELECT p.id::text, (NOW() AT TIME ZONE 'Asia/Shanghai')::date::text,
		COALESCE(q.used,0), COALESCE(q.free_used,0),
		COALESCE(q.processing,0), COALESCE(q.free_processing,0), COALESCE(c.balance,0)
	FROM page_users p
	LEFT JOIN request_usage q ON q.user_id=p.id
	LEFT JOIN credits c ON c.user_id=p.id`, pq.Array(ids))
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		var usage AdminTodayQuestionUsage
		if err := rows.Scan(&id, &usage.Date, &usage.Used, &usage.FreeUsed,
			&usage.Processing, &usage.FreeProcessing, &usage.CreditBalance); err != nil {
			return err
		}
		if usage.CreditBalance < 0 {
			return fmt.Errorf("negative today question credit balance for user %s", id)
		}
		usage.DailyLimit = limit
		freeRemaining := limit - usage.FreeUsed - usage.FreeProcessing
		if freeRemaining < 0 {
			freeRemaining = 0
		}
		usage.Remaining = freeRemaining + usage.CreditBalance
		byID[id].TodayQuestion = usage
	}
	return rows.Err()
}

// Reconcile only expired reservations found on this page before displaying balances.
func releaseStaleTodayQuestionsForAdminPage(ids []string) error {
	rows, err := database.DB.Query(`SELECT DISTINCT user_id::text FROM today_question_requests
		WHERE user_id=ANY($1::uuid[]) AND quota_charged AND status='processing'
		AND created_at < NOW() - INTERVAL '10 minutes'`, pq.Array(ids))
	if err != nil {
		return err
	}
	staleUsers := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		staleUsers = append(staleUsers, id)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()
	if len(staleUsers) == 0 {
		return nil
	}
	sort.Strings(staleUsers)
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, id := range staleUsers {
		if err := lockTodayQuestionUserTx(tx, id); err != nil {
			return err
		}
		if err := releaseStaleTodayQuestionRequestsTx(tx, id); err != nil {
			return err
		}
	}
	return tx.Commit()
}
