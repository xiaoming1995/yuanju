package repository_test

import (
	"context"
	"testing"
	"time"

	"yuanju/internal/model"
	"yuanju/internal/repository"
	"yuanju/pkg/database"
)

func TestUserActivityTimelineFiltersMergeAndDeletion(t *testing.T) {
	db, cleanup := spinUpPG(t)
	defer cleanup()
	restore := swapGlobalDB(t, db)
	defer restore()
	if _, err := database.Migrate(database.ModeApply); err != nil {
		t.Fatal(err)
	}
	var webID, miniID, adminID string
	if err := db.QueryRow(`INSERT INTO users (email,password_hash,source) VALUES ('activity-web@example.com','hash','self_registered') RETURNING id`).Scan(&webID); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`INSERT INTO users (source) VALUES ('miniapp_wechat') RETURNING id`).Scan(&miniID); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`INSERT INTO admins (email,password_hash,name) VALUES ('activity-admin@example.com','hash','Admin') RETURNING id`).Scan(&adminID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO wechat_user_identities (app_id,open_id,user_id) VALUES ('activity-app','openid',$1)`, miniID); err != nil {
		t.Fatal(err)
	}

	users, total, err := repository.ListAdminUsersFiltered(1, 20, "", "", "", "never", "created_desc")
	if err != nil || total != 2 || len(users) != 2 || users[0].LastLoginAt != nil {
		t.Fatalf("initial never users=%+v total=%d err=%v", users, total, err)
	}
	if err := repository.RecordUserActivity(context.Background(), webID, "login", "web", nil); err != nil {
		t.Fatal(err)
	}
	if err := repository.RecordUserActivity(context.Background(), miniID, "login", "miniapp", nil); err != nil {
		t.Fatal(err)
	}
	if err := repository.RecordUserActivity(context.Background(), miniID, "nickname_updated", "miniapp", nil); err != nil {
		t.Fatal(err)
	}
	users, total, err = repository.ListAdminUsersFiltered(1, 20, "", "", "", "never", "created_desc")
	if err != nil || total != 0 || len(users) != 0 {
		t.Fatalf("never after login: %d %+v %v", total, users, err)
	}
	events, total, err := repository.ListUserActivity(miniID, 1, 1)
	if err != nil || total != 2 || len(events) != 1 || events[0].Type != "nickname_updated" {
		t.Fatalf("timeline %+v total=%d err=%v", events, total, err)
	}
	secondPage, secondTotal, err := repository.ListUserActivity(miniID, 2, 1)
	if err != nil || secondTotal != 2 || len(secondPage) != 1 || secondPage[0].Type != "login" {
		t.Fatalf("timeline page two %+v total=%d err=%v", secondPage, secondTotal, err)
	}
	if _, err := db.Exec(`UPDATE users SET chart_query_count=CASE WHEN id=$1 THEN 2 ELSE 1 END WHERE id IN ($1,$2)`, miniID, webID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO user_chart_query_history (chart_id,user_id,queried_at) VALUES (gen_random_uuid(),$1,NOW())`, miniID); err != nil {
		t.Fatal(err)
	}

	if err := repository.MergeMiniappIntoWebUser(context.Background(), miniID, webID, "activity-app"); err != nil {
		t.Fatal(err)
	}
	queryHistory, queryTotal, err := repository.ListAdminUserChartQueries(1, 20, webID, "", "")
	if err != nil || queryTotal != 1 || len(queryHistory) != 1 || !queryHistory[0].Deleted {
		t.Fatalf("merged query history %+v total=%d err=%v", queryHistory, queryTotal, err)
	}
	events, total, err = repository.ListUserActivity(webID, 1, 20)
	if err != nil || total != 4 || len(events) != 4 || events[0].Type != "account_linked" {
		t.Fatalf("merged timeline %+v total=%d err=%v", events, total, err)
	}
	detail, err := repository.GetAdminUserDetail(webID)
	if err != nil || detail == nil || detail.LastLoginAt == nil || detail.LastActiveAt == nil || detail.LastLoginChannel != "miniapp" || detail.ChartQueryCount != 3 {
		t.Fatalf("detail %+v err=%v", detail, err)
	}
	old := time.Now().AddDate(0, 0, -100)
	if _, err := db.Exec(`UPDATE user_activity_events SET created_at=$1 WHERE user_id=$2`, old, webID); err != nil {
		t.Fatal(err)
	}
	if n, err := repository.CleanUserActivity(context.Background(), time.Now().AddDate(0, 0, -90)); err != nil || n != 4 {
		t.Fatalf("cleanup n=%d err=%v", n, err)
	}
	detail, err = repository.GetAdminUserDetail(webID)
	if err != nil || detail.LastLoginAt == nil || detail.LastActiveAt == nil {
		t.Fatalf("snapshots lost: %+v %v", detail, err)
	}
	chart, err := repository.CreateChartWithChannel(&model.BaziChart{UserID: &webID, BirthYear: 1990, BirthMonth: 1, BirthDay: 1, BirthHour: 1, Gender: "male", ChartHash: "activity-chart"}, "web")
	if err != nil || chart == nil {
		t.Fatalf("create chart: %+v %v", chart, err)
	}
	secondChart, err := repository.CreateChartWithChannel(&model.BaziChart{UserID: &webID, BirthYear: 1990, BirthMonth: 1, BirthDay: 1, BirthHour: 1, Gender: "male", ChartHash: "activity-chart"}, "web")
	if err != nil || secondChart == nil {
		t.Fatalf("repeat chart: %+v %v", secondChart, err)
	}
	if _, err := repository.DeleteChart(secondChart.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.CreateChartWithChannel(&model.BaziChart{BirthYear: 1990, BirthMonth: 1, BirthDay: 1, BirthHour: 1, Gender: "male", ChartHash: "guest-chart"}, "web"); err != nil {
		t.Fatal(err)
	}
	detail, err = repository.GetAdminUserDetail(webID)
	if err != nil || detail.ChartQueryCount != 5 || detail.ChartCount != 1 {
		t.Fatalf("query count after repeat/delete/guest: %+v %v", detail, err)
	}
	queryHistory, queryTotal, err = repository.ListAdminUserChartQueries(1, 20, webID, "", "")
	if err != nil || queryTotal != 3 || len(queryHistory) != 3 {
		t.Fatalf("query history after repeat/delete/guest: %+v total=%d err=%v", queryHistory, queryTotal, err)
	}
	deleted, active := 0, 0
	for _, record := range queryHistory {
		if record.Deleted {
			deleted++
			if record.ChartID != nil {
				t.Fatalf("deleted query exposes chart ID: %+v", record)
			}
		} else {
			active++
		}
	}
	if deleted != 2 || active != 1 {
		t.Fatalf("query history states: %+v", queryHistory)
	}
	pageTwo, pageTotal, err := repository.ListAdminUserChartQueries(2, 1, webID, "", "")
	if err != nil || pageTotal != 3 || len(pageTwo) != 1 {
		t.Fatalf("query history pagination: %+v total=%d err=%v", pageTwo, pageTotal, err)
	}
	users, _, err = repository.ListAdminUsersFiltered(1, 20, webID, "", "", "", "created_desc")
	if err != nil || len(users) != 1 || users[0].ChartQueryCount != 5 || users[0].ChartCount != 1 {
		t.Fatalf("admin list query count: %+v %v", users, err)
	}
	report, err := repository.CreateReportWithChannel(chart.ID, "test report", "test-model", nil, "web")
	if err != nil || report == nil {
		t.Fatalf("create report: %+v %v", report, err)
	}
	if err := repository.UpdateUserNickname(context.Background(), webID, "新昵称", "web"); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.UpdateUserAvatar(context.Background(), webID, "user-avatars/test.png", "web"); err != nil {
		t.Fatal(err)
	}
	events, total, err = repository.ListUserActivity(webID, 1, 20)
	if err != nil || total != 5 || events[0].Type != "avatar_updated" || events[1].Type != "nickname_updated" || events[2].Type != "bazi_report_generated" || events[2].ObjectID == nil || *events[2].ObjectID != report.ID || events[3].Type != "chart_created" || events[4].Type != "chart_created" {
		t.Fatalf("core events %+v total=%d err=%v", events, total, err)
	}
	if err := repository.SetUserDisabledAudited(context.Background(), adminID, webID, true); err != nil {
		t.Fatal(err)
	}
	if err := repository.SetUserDisabledAudited(context.Background(), adminID, webID, false); err != nil {
		t.Fatal(err)
	}
	auditEvents, err := repository.ListAdminUserAudit(webID, 1, 20)
	if err != nil || len(auditEvents) != 2 || auditEvents[0].Action != "enabled" || auditEvents[0].AdminName != "Admin" {
		t.Fatalf("audit before delete %+v %v", auditEvents, err)
	}
	if err := repository.DeleteUserAudited(context.Background(), adminID, webID); err != nil {
		t.Fatal(err)
	}
	queryHistory, queryTotal, err = repository.ListAdminUserChartQueries(1, 20, webID, "", "")
	if err != nil || queryTotal != 0 || len(queryHistory) != 0 {
		t.Fatalf("deleted user retained query history: %+v total=%d err=%v", queryHistory, queryTotal, err)
	}
	events, total, err = repository.ListUserActivity(webID, 1, 20)
	if err != nil || total != 0 || len(events) != 0 {
		t.Fatalf("deleted events %+v %d %v", events, total, err)
	}
	var action string
	if err := db.QueryRow(`SELECT action FROM admin_user_audit_events WHERE target_user_id=$1 ORDER BY created_at DESC,id DESC LIMIT 1`, webID).Scan(&action); err != nil || action != "deleted" {
		t.Fatalf("audit %q %v", action, err)
	}
	if _, err := db.Exec(`UPDATE admin_user_audit_events SET created_at=$1 WHERE target_user_id=$2`, time.Now().AddDate(0, 0, -190), webID); err != nil {
		t.Fatal(err)
	}
	if n, err := repository.CleanAdminUserAudit(context.Background(), time.Now().AddDate(0, 0, -180)); err != nil || n != 3 {
		t.Fatalf("audit cleanup n=%d err=%v", n, err)
	}
}
