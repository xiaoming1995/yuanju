package database_test

import (
	"os"
	"strings"
	"testing"
)

func TestArticleMigrationDeclaresCoreTablesAndIndexes(t *testing.T) {
	sql, err := os.ReadFile("migrations/00016_article_inspiration_library.sql")
	if err != nil {
		t.Fatalf("read article migration: %v", err)
	}
	body := string(sql)
	for _, want := range []string{
		"CREATE TABLE IF NOT EXISTS article_categories",
		"CREATE TABLE IF NOT EXISTS article_tags",
		"CREATE TABLE IF NOT EXISTS article_keywords",
		"CREATE TABLE IF NOT EXISTS articles",
		"CREATE TABLE IF NOT EXISTS article_tag_links",
		"CREATE TABLE IF NOT EXISTS article_collection_tasks",
		"CREATE TABLE IF NOT EXISTS article_collection_task_items",
		"CREATE TABLE IF NOT EXISTS article_audit_events",
		"CREATE TABLE IF NOT EXISTS article_original_clicks",
		"CREATE TABLE IF NOT EXISTS article_favorites",
		"CREATE TABLE IF NOT EXISTS article_ai_providers",
		"CREATE TABLE IF NOT EXISTS article_ai_prompts",
		"CREATE INDEX IF NOT EXISTS idx_articles_status",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_articles_canonical_url_hash",
		"CREATE INDEX IF NOT EXISTS idx_article_tag_links_tag",
		"CREATE INDEX IF NOT EXISTS idx_article_collection_tasks_status",
		"CREATE INDEX IF NOT EXISTS idx_article_original_clicks_article_user",
		"full_text_authorized BOOLEAN NOT NULL DEFAULT false",
		"body_content TEXT",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("article migration should contain %q", want)
		}
	}
}

func TestArticleBodyFetchStatusMigrationDeclaresDiagnosticsColumns(t *testing.T) {
	sql, err := os.ReadFile("migrations/00017_article_body_fetch_status.sql")
	if err != nil {
		t.Fatalf("read article body status migration: %v", err)
	}
	body := string(sql)
	for _, want := range []string{
		"ADD COLUMN IF NOT EXISTS body_fetch_status VARCHAR(40) NOT NULL DEFAULT 'pending'",
		"ADD COLUMN IF NOT EXISTS body_fetch_error TEXT NOT NULL DEFAULT ''",
		"WHEN full_text_authorized = true AND COALESCE(body_content, '') <> '' THEN 'succeeded'",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("article body status migration should contain %q", want)
		}
	}
}

func TestArticleCollectionRandomPagesMigrationDeclaresConfigColumns(t *testing.T) {
	sql, err := os.ReadFile("migrations/00018_article_collection_random_pages.sql")
	if err != nil {
		t.Fatalf("read article collection random pages migration: %v", err)
	}
	body := string(sql)
	for _, want := range []string{
		"ADD COLUMN IF NOT EXISTS search_page_min INTEGER NOT NULL DEFAULT 1",
		"ADD COLUMN IF NOT EXISTS search_page_max INTEGER NOT NULL DEFAULT 5",
		"search_page_min = LEAST(GREATEST(1, COALESCE(search_page_min, 1)), 20)",
		"search_page_max = LEAST",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("article collection random pages migration should contain %q", want)
		}
	}
}

func TestArticleCollectionMinuteScheduleMigrationDeclaresIntervalColumn(t *testing.T) {
	sql, err := os.ReadFile("migrations/00019_article_collection_minute_schedule.sql")
	if err != nil {
		t.Fatalf("read article collection minute schedule migration: %v", err)
	}
	body := string(sql)
	for _, want := range []string{
		"ADD COLUMN IF NOT EXISTS schedule_interval_minutes INTEGER NOT NULL DEFAULT 1440",
		"WHEN frequency = 'weekly' THEN 10080",
		"LEAST(GREATEST(schedule_interval_minutes, 1), 10080)",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("article collection minute schedule migration should contain %q", want)
		}
	}
}

func TestArticleCollectionQualityScoringMigrationDeclaresColumnsAndConfig(t *testing.T) {
	sql, err := os.ReadFile("migrations/00020_article_collection_quality_scoring.sql")
	if err != nil {
		t.Fatalf("read article quality scoring migration: %v", err)
	}
	body := string(sql)
	for _, want := range []string{
		"ADD COLUMN IF NOT EXISTS quality_score INTEGER NOT NULL DEFAULT 0",
		"ADD COLUMN IF NOT EXISTS quality_reasons JSONB NOT NULL DEFAULT '[]'::jsonb",
		"ADD COLUMN IF NOT EXISTS skip_reason TEXT NOT NULL DEFAULT ''",
		"CREATE TABLE IF NOT EXISTS article_quality_config",
		"quality_filter_enabled BOOLEAN NOT NULL DEFAULT false",
		"min_quality_score INTEGER NOT NULL DEFAULT 60",
		"source_blacklist TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[]",
		"CREATE INDEX IF NOT EXISTS idx_articles_quality_score",
		"CREATE INDEX IF NOT EXISTS idx_article_collection_task_items_quality",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("article quality scoring migration should contain %q", want)
		}
	}
}

func TestArticleAutoPublishByQualityMigrationDeclaresColumns(t *testing.T) {
	sql, err := os.ReadFile("migrations/00021_article_auto_publish_by_quality.sql")
	if err != nil {
		t.Fatalf("read article auto publish migration: %v", err)
	}
	body := string(sql)
	for _, want := range []string{
		"ADD COLUMN IF NOT EXISTS auto_publish_enabled BOOLEAN NOT NULL DEFAULT false",
		"ADD COLUMN IF NOT EXISTS auto_publish_min_quality_score INTEGER NOT NULL DEFAULT 90",
		"ADD COLUMN IF NOT EXISTS auto_publish_requires_body BOOLEAN NOT NULL DEFAULT true",
		"ADD COLUMN IF NOT EXISTS auto_publish_max_per_run INTEGER NOT NULL DEFAULT 3",
		"ADD COLUMN IF NOT EXISTS auto_published BOOLEAN NOT NULL DEFAULT false",
		"ADD COLUMN IF NOT EXISTS auto_publish_reason TEXT NOT NULL DEFAULT ''",
		"CREATE INDEX IF NOT EXISTS idx_article_collection_task_items_auto_published",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("article auto publish migration should contain %q", want)
		}
	}
}

func TestArticleModuleEnabledSettingMigrationDeclaresDefault(t *testing.T) {
	sql, err := os.ReadFile("migrations/00022_article_module_enabled_setting.sql")
	if err != nil {
		t.Fatalf("read article module setting migration: %v", err)
	}
	body := string(sql)
	for _, want := range []string{
		"CREATE TABLE IF NOT EXISTS system_settings",
		"VALUES ('articles_module_enabled', 'true')",
		"ON CONFLICT (key) DO NOTHING",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("article module setting migration should contain %q", want)
		}
	}
}

func TestMingGeHistoricalFigureMigrationsDeclareReviewAndSeedData(t *testing.T) {
	schema, err := os.ReadFile("migrations/00025_mingge_historical_figures.sql")
	if err != nil {
		t.Fatalf("read historical figure schema migration: %v", err)
	}
	if !strings.Contains(string(schema), "mingge_historical_figures") || !strings.Contains(string(schema), "review_status") {
		t.Fatalf("historical figure schema migration missing table or review status: %s", schema)
	}
	seed, err := os.ReadFile("migrations/00026_seed_mingge_historical_figures.sql")
	if err != nil {
		t.Fatalf("read historical figure seed migration: %v", err)
	}
	if !strings.Contains(string(seed), "伤官格") || !strings.Contains(string(seed), "show_dayun") {
		t.Fatalf("historical figure seed migration missing reviewed reference data: %s", seed)
	}
}
