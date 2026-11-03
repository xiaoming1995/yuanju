package database_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"yuanju/pkg/database"
)

func TestUserChartQueryCountMigrationBackfillsExistingCharts(t *testing.T) {
	db, cleanup := spinUpPG(t)
	defer cleanup()
	restore := swapGlobalDB(t, db)
	defer restore()

	dir := t.TempDir()
	entries, err := os.ReadDir("migrations")
	if err != nil {
		t.Fatal(err)
	}
	var nextMigration []byte
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}
		data, err := os.ReadFile(filepath.Join("migrations", name))
		if err != nil {
			t.Fatal(err)
		}
		if name == "00032_user_chart_query_count.sql" {
			nextMigration = data
			continue
		}
		if name < "00032" {
			if err := os.WriteFile(filepath.Join(dir, name), data, 0600); err != nil {
				t.Fatal(err)
			}
		}
	}
	if len(nextMigration) == 0 {
		t.Fatal("missing chart count migration")
	}
	if _, err := database.MigrateFromDir(database.ModeApply, dir); err != nil {
		t.Fatal(err)
	}
	var userID string
	if err := db.QueryRow(`INSERT INTO users (email,password_hash) VALUES ('chart-count-legacy@example.com','hash') RETURNING id`).Scan(&userID); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if _, err := db.Exec(`INSERT INTO bazi_charts (user_id,birth_year,birth_month,birth_day,birth_hour,gender) VALUES ($1,1990,1,1,1,'male')`, userID); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(`INSERT INTO bazi_charts (birth_year,birth_month,birth_day,birth_hour,gender) VALUES (1990,1,1,1,'male')`); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "00032_user_chart_query_count.sql"), nextMigration, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := database.MigrateFromDir(database.ModeApply, dir); err != nil {
		t.Fatal(err)
	}
	var count int64
	if err := db.QueryRow(`SELECT chart_query_count FROM users WHERE id=$1`, userID).Scan(&count); err != nil || count != 2 {
		t.Fatalf("backfilled count=%d err=%v", count, err)
	}
	var deletedChartID string
	if err := db.QueryRow(`INSERT INTO user_activity_events (user_id,event_type,channel,object_id)
		VALUES ($1,'chart_created','web',gen_random_uuid()) RETURNING object_id`, userID).Scan(&deletedChartID); err != nil {
		t.Fatal(err)
	}
	historyMigration, err := os.ReadFile(filepath.Join("migrations", "00033_user_chart_query_history.sql"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "00033_user_chart_query_history.sql"), historyMigration, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := database.MigrateFromDir(database.ModeApply, dir); err != nil {
		t.Fatal(err)
	}
	var historyCount, deletedCount int
	if err := db.QueryRow(`SELECT COUNT(*), COUNT(*) FILTER (WHERE c.id IS NULL)
		FROM user_chart_query_history h LEFT JOIN bazi_charts c ON c.id=h.chart_id WHERE h.user_id=$1`, userID).Scan(&historyCount, &deletedCount); err != nil || historyCount != 3 || deletedCount != 1 {
		t.Fatalf("history count=%d deleted=%d err=%v", historyCount, deletedCount, err)
	}
	if err := db.QueryRow(`SELECT chart_query_count FROM users WHERE id=$1`, userID).Scan(&count); err != nil || count != 3 {
		t.Fatalf("aligned count=%d err=%v", count, err)
	}
	var retained bool
	if err := db.QueryRow(`SELECT EXISTS (SELECT 1 FROM user_chart_query_history WHERE chart_id=$1 AND user_id=$2)`, deletedChartID, userID).Scan(&retained); err != nil || !retained {
		t.Fatalf("recent deleted query missing: retained=%t err=%v", retained, err)
	}
}
