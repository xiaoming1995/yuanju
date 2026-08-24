-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS article_categories (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	name VARCHAR(100) NOT NULL,
	slug VARCHAR(120) NOT NULL UNIQUE,
	sort_order INTEGER NOT NULL DEFAULT 0,
	active BOOLEAN NOT NULL DEFAULT true,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS article_tags (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	name VARCHAR(100) NOT NULL,
	slug VARCHAR(120) NOT NULL UNIQUE,
	active BOOLEAN NOT NULL DEFAULT true,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS article_keywords (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	keyword VARCHAR(100) NOT NULL UNIQUE,
	active BOOLEAN NOT NULL DEFAULT true,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS articles (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	title VARCHAR(500) NOT NULL,
	source_name VARCHAR(255) NOT NULL DEFAULT '',
	original_url TEXT NOT NULL,
	canonical_url_hash VARCHAR(64) NOT NULL,
	cover_url TEXT NOT NULL DEFAULT '',
	published_at_source TIMESTAMPTZ,
	search_snippet TEXT NOT NULL DEFAULT '',
	summary TEXT NOT NULL DEFAULT '',
	ai_analysis JSONB,
	ai_status VARCHAR(20) NOT NULL DEFAULT 'pending',
	ai_error_msg TEXT NOT NULL DEFAULT '',
	category_id UUID REFERENCES article_categories(id) ON DELETE SET NULL,
	status VARCHAR(20) NOT NULL DEFAULT 'candidate',
	view_count INTEGER NOT NULL DEFAULT 0,
	original_click_count INTEGER NOT NULL DEFAULT 0,
	full_text_authorized BOOLEAN NOT NULL DEFAULT false,
	body_content TEXT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	published_at TIMESTAMPTZ,
	taken_down_at TIMESTAMPTZ,
	deleted_at TIMESTAMPTZ,
	CONSTRAINT chk_articles_status CHECK (status IN ('candidate', 'published', 'rejected', 'taken_down', 'deleted')),
	CONSTRAINT chk_articles_ai_status CHECK (ai_status IN ('pending', 'succeeded', 'failed'))
);

CREATE TABLE IF NOT EXISTS article_tag_links (
	article_id UUID NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
	tag_id UUID NOT NULL REFERENCES article_tags(id) ON DELETE CASCADE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	PRIMARY KEY (article_id, tag_id)
);

CREATE TABLE IF NOT EXISTS article_collection_tasks (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	trigger_type VARCHAR(30) NOT NULL DEFAULT 'manual',
	status VARCHAR(20) NOT NULL DEFAULT 'pending',
	started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	finished_at TIMESTAMPTZ,
	keyword_count INTEGER NOT NULL DEFAULT 0,
	found_count INTEGER NOT NULL DEFAULT 0,
	inserted_count INTEGER NOT NULL DEFAULT 0,
	duplicate_count INTEGER NOT NULL DEFAULT 0,
	failed_count INTEGER NOT NULL DEFAULT 0,
	error_msg TEXT NOT NULL DEFAULT '',
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT chk_article_collection_tasks_status CHECK (status IN ('pending', 'running', 'succeeded', 'failed'))
);

CREATE TABLE IF NOT EXISTS article_collection_task_items (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	task_id UUID NOT NULL REFERENCES article_collection_tasks(id) ON DELETE CASCADE,
	article_id UUID REFERENCES articles(id) ON DELETE SET NULL,
	original_url TEXT NOT NULL DEFAULT '',
	keyword VARCHAR(100) NOT NULL DEFAULT '',
	status VARCHAR(20) NOT NULL,
	error_msg TEXT NOT NULL DEFAULT '',
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT chk_article_collection_task_items_status CHECK (status IN ('inserted', 'duplicate', 'failed', 'skipped'))
);

CREATE TABLE IF NOT EXISTS article_audit_events (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	article_id UUID NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
	admin_id UUID REFERENCES admins(id) ON DELETE SET NULL,
	action VARCHAR(30) NOT NULL,
	from_status VARCHAR(20) NOT NULL,
	to_status VARCHAR(20) NOT NULL,
	note TEXT NOT NULL DEFAULT '',
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS article_original_clicks (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	article_id UUID NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
	user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	clicked_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS article_favorites (
	user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	article_id UUID NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	PRIMARY KEY (user_id, article_id)
);

CREATE TABLE IF NOT EXISTS article_collection_config (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	enabled BOOLEAN NOT NULL DEFAULT false,
	frequency VARCHAR(20) NOT NULL DEFAULT 'daily',
	max_results_per_run INTEGER NOT NULL DEFAULT 20,
	last_run_at TIMESTAMPTZ,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT chk_article_collection_config_frequency CHECK (frequency IN ('daily', 'weekly'))
);

CREATE TABLE IF NOT EXISTS article_ai_providers (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	name VARCHAR(100) NOT NULL,
	type VARCHAR(50) NOT NULL,
	base_url VARCHAR(500) NOT NULL,
	model VARCHAR(100) NOT NULL,
	api_key_encrypted TEXT NOT NULL,
	api_key_preview VARCHAR(40) NOT NULL DEFAULT '',
	active BOOLEAN NOT NULL DEFAULT false,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS article_ai_prompts (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	content TEXT NOT NULL,
	description VARCHAR(255) NOT NULL DEFAULT '',
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO article_collection_config (enabled, frequency, max_results_per_run)
SELECT false, 'daily', 20
WHERE NOT EXISTS (SELECT 1 FROM article_collection_config);

INSERT INTO article_ai_prompts (content, description)
SELECT '请基于文章标题、来源、发布时间、公开搜索摘要、平台摘要、原文链接、后台分类和标签，输出资讯阅读辅助与仿写拆解 JSON。不要编造原文正文细节。', '资讯文章 AI 分析默认 Prompt'
WHERE NOT EXISTS (SELECT 1 FROM article_ai_prompts);

CREATE INDEX IF NOT EXISTS idx_article_categories_active_sort ON article_categories(active, sort_order, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_article_tags_active ON article_tags(active, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_article_keywords_active ON article_keywords(active, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_articles_status ON articles(status);
CREATE INDEX IF NOT EXISTS idx_articles_category ON articles(category_id);
CREATE INDEX IF NOT EXISTS idx_articles_source_publish_time ON articles(published_at_source DESC NULLS LAST);
CREATE INDEX IF NOT EXISTS idx_articles_view_count ON articles(view_count DESC);
CREATE UNIQUE INDEX IF NOT EXISTS idx_articles_canonical_url_hash ON articles(canonical_url_hash);
CREATE INDEX IF NOT EXISTS idx_articles_published_sort ON articles(status, published_at DESC NULLS LAST, published_at_source DESC NULLS LAST, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_article_tag_links_tag ON article_tag_links(tag_id, article_id);
CREATE INDEX IF NOT EXISTS idx_article_collection_tasks_status ON article_collection_tasks(status, started_at DESC);
CREATE INDEX IF NOT EXISTS idx_article_collection_task_items_task ON article_collection_task_items(task_id, status);
CREATE INDEX IF NOT EXISTS idx_article_audit_events_article ON article_audit_events(article_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_article_original_clicks_article_user ON article_original_clicks(article_id, user_id, clicked_at DESC);
CREATE INDEX IF NOT EXISTS idx_article_original_clicks_user ON article_original_clicks(user_id, clicked_at DESC);
CREATE UNIQUE INDEX IF NOT EXISTS idx_article_ai_providers_active ON article_ai_providers(active) WHERE active;

-- +goose StatementEnd
