package service

import (
	"context"
	"database/sql"
	"testing"
	"time"
	"yuanju/internal/model"
	"yuanju/internal/repository"
	"yuanju/pkg/database"

	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	tcpg "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func withArticleCollectionSchedulerDB(t *testing.T) *sql.DB {
	t.Helper()
	ctx := context.Background()
	pg, err := tcpg.Run(ctx,
		"postgres:16-alpine",
		tcpg.WithDatabase("yuanju_article_scheduler_test"),
		tcpg.WithUsername("test"),
		tcpg.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}
	t.Cleanup(func() { _ = pg.Terminate(ctx) })

	dsn, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("conn string: %v", err)
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Fatalf("ping db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	origDB := database.DB
	database.DB = db
	t.Cleanup(func() { database.DB = origDB })

	if _, err := database.Migrate(database.ModeApply); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestArticleCollectionDueUsesMinuteInterval(t *testing.T) {
	lastRun := time.Now().Add(-59 * time.Minute)
	cfg := &model.ArticleCollectionConfig{
		ScheduleInterval: 60,
		LastRunAt:        &lastRun,
	}
	if articleCollectionDue(cfg) {
		t.Fatal("collection should not be due before configured minute interval")
	}

	lastRun = time.Now().Add(-61 * time.Minute)
	cfg.LastRunAt = &lastRun
	if !articleCollectionDue(cfg) {
		t.Fatal("collection should be due after configured minute interval")
	}
}

func TestArticleCollectionDueAllowsOneMinuteInterval(t *testing.T) {
	lastRun := time.Now().Add(-30 * time.Second)
	cfg := &model.ArticleCollectionConfig{
		ScheduleInterval: 1,
		LastRunAt:        &lastRun,
	}
	if articleCollectionDue(cfg) {
		t.Fatal("collection should not be due before the one minute interval")
	}

	lastRun = time.Now().Add(-90 * time.Second)
	cfg.LastRunAt = &lastRun
	if !articleCollectionDue(cfg) {
		t.Fatal("collection should be due after the one minute interval")
	}
}

func TestArticleCollectionDueFallsBackToLegacyFrequency(t *testing.T) {
	lastRun := time.Now().Add(-8 * 24 * time.Hour)
	cfg := &model.ArticleCollectionConfig{
		Frequency: "weekly",
		LastRunAt: &lastRun,
	}
	if !articleCollectionDue(cfg) {
		t.Fatal("collection should fall back to weekly frequency when minute interval is missing")
	}
}

func TestScheduledArticleCollectionSkipsWhenModuleDisabled(t *testing.T) {
	db := withArticleCollectionSchedulerDB(t)
	if err := repository.SetBoolSetting(repository.SettingArticlesModuleEnabled, false); err != nil {
		t.Fatalf("disable article module: %v", err)
	}
	if _, err := repository.UpdateArticleCollectionConfig(model.ArticleCollectionConfig{
		Enabled:                    true,
		Frequency:                  "daily",
		ScheduleInterval:           1,
		MaxResultsPerRun:           2,
		SearchPageMin:              1,
		SearchPageMax:              1,
		AutoPublishMinQualityScore: 90,
		AutoPublishRequiresBody:    true,
		AutoPublishMaxPerRun:       1,
	}); err != nil {
		t.Fatalf("update collection config: %v", err)
	}
	if _, err := repository.CreateArticleKeyword("八字", true); err != nil {
		t.Fatalf("create keyword: %v", err)
	}

	provider := &fakeArticleSearchProvider{results: map[string][]ArticleSearchResult{}}
	runScheduledArticleCollection(context.Background(), provider)

	if len(provider.queries) != 0 {
		t.Fatalf("provider queries=%v, want none while module disabled", provider.queries)
	}
	var taskCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM article_collection_tasks`).Scan(&taskCount); err != nil {
		t.Fatalf("count collection tasks: %v", err)
	}
	if taskCount != 0 {
		t.Fatalf("collection task count=%d, want 0", taskCount)
	}
	var lastRun sql.NullTime
	if err := db.QueryRow(`SELECT last_run_at FROM article_collection_config ORDER BY updated_at DESC LIMIT 1`).Scan(&lastRun); err != nil {
		t.Fatalf("read last run: %v", err)
	}
	if lastRun.Valid {
		t.Fatalf("last_run_at=%v, want null while module disabled", lastRun.Time)
	}
}
