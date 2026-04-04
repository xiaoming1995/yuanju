package database

import (
	"database/sql"
	"log"
	"yuanju/configs"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Connect() {
	var err error
	DB, err = sql.Open("postgres", configs.AppConfig.DatabaseURL)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatalf("数据库 Ping 失败: %v", err)
	}

	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)

	log.Println("✅ 数据库连接成功")
}

func Migrate() {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		email VARCHAR(255) UNIQUE NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		nickname VARCHAR(100),
		created_at TIMESTAMPTZ DEFAULT NOW(),
		updated_at TIMESTAMPTZ DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS bazi_charts (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id UUID REFERENCES users(id) ON DELETE SET NULL,
		birth_year INTEGER NOT NULL,
		birth_month INTEGER NOT NULL,
		birth_day INTEGER NOT NULL,
		birth_hour INTEGER NOT NULL,
		gender VARCHAR(10) NOT NULL DEFAULT 'male',
		year_gan VARCHAR(10),
		year_zhi VARCHAR(10),
		month_gan VARCHAR(10),
		month_zhi VARCHAR(10),
		day_gan VARCHAR(10),
		day_zhi VARCHAR(10),
		hour_gan VARCHAR(10),
		hour_zhi VARCHAR(10),
		wuxing JSONB,
		dayun JSONB,
		yongshen VARCHAR(20),
		jishen VARCHAR(20),
		chart_hash VARCHAR(64) UNIQUE,
		created_at TIMESTAMPTZ DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS ai_reports (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		chart_id UUID REFERENCES bazi_charts(id) ON DELETE CASCADE,
		content TEXT NOT NULL,
		model VARCHAR(50),
		created_at TIMESTAMPTZ DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_bazi_charts_user_id ON bazi_charts(user_id);
	CREATE INDEX IF NOT EXISTS idx_bazi_charts_hash ON bazi_charts(chart_hash);
	CREATE INDEX IF NOT EXISTS idx_ai_reports_chart_id ON ai_reports(chart_id);

	CREATE TABLE IF NOT EXISTS admins (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		email VARCHAR(255) UNIQUE NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		name VARCHAR(100),
		created_at TIMESTAMPTZ DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS llm_providers (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name VARCHAR(100) NOT NULL,
		type VARCHAR(50) NOT NULL,
		base_url VARCHAR(500) NOT NULL,
		model VARCHAR(100) NOT NULL,
		api_key_encrypted TEXT NOT NULL,
		active BOOLEAN NOT NULL DEFAULT false,
		created_at TIMESTAMPTZ DEFAULT NOW(),
		updated_at TIMESTAMPTZ DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS ai_requests_log (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		chart_id UUID REFERENCES bazi_charts(id) ON DELETE SET NULL,
		provider_id UUID REFERENCES llm_providers(id) ON DELETE SET NULL,
		model VARCHAR(100),
		duration_ms INTEGER,
		status VARCHAR(20) NOT NULL DEFAULT 'success',
		error_msg TEXT,
		created_at TIMESTAMPTZ DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_ai_requests_log_created ON ai_requests_log(created_at);
	CREATE INDEX IF NOT EXISTS idx_ai_requests_log_provider ON ai_requests_log(provider_id);

	CREATE TABLE IF NOT EXISTS celebrity_records (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name VARCHAR(255) NOT NULL,
		gender VARCHAR(10),
		traits TEXT,
		career VARCHAR(255),
		active BOOLEAN NOT NULL DEFAULT true,
		created_at TIMESTAMPTZ DEFAULT NOW(),
		updated_at TIMESTAMPTZ DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_celebrity_records_active ON celebrity_records(active);
	`

	if _, err := DB.Exec(schema); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 增量迁移：为 ai_reports 表新增 content_structured 字段（JSONB，历史兼容）
	alterSQL := `ALTER TABLE ai_reports ADD COLUMN IF NOT EXISTS content_structured JSONB;`
	if _, err := DB.Exec(alterSQL); err != nil {
		log.Fatalf("增量迁移失败 (content_structured): %v", err)
	}

	// 增量迁移：chart_hash 从全局唯一改为每用户唯一（此段历史增量旧锁已被 resource-based-bazi 废弃取消，防止启动重试添加时遭现存无约束重复记录报错）
	chartHashMigrations := []string{
		`ALTER TABLE bazi_charts DROP CONSTRAINT IF EXISTS bazi_charts_chart_hash_key;`,
		`DROP INDEX IF EXISTS idx_bazi_charts_hash;`,
	}
	for _, migSQL := range chartHashMigrations {
		if _, err := DB.Exec(migSQL); err != nil {
			log.Fatalf("增量迁移失败 (chart_hash isolation): %v\nSQL: %s", err, migSQL)
		}
	}

	// 增量迁移 (resource-based-bazi)：切底解除基于 chart_hash 的独立限制，让每次查命均落位新快照记录
	resourceUnbindMigrations := []string{
		`ALTER TABLE bazi_charts DROP CONSTRAINT IF EXISTS bazi_charts_chart_hash_user_id_key;`,
		`ALTER TABLE bazi_charts DROP CONSTRAINT IF EXISTS bazi_charts_chart_hash_key;`,
		`DROP INDEX IF EXISTS idx_bazi_charts_hash_user;`,
		`DROP INDEX IF EXISTS idx_bazi_charts_hash;`,
	}
	for _, sql := range resourceUnbindMigrations {
		if _, err := DB.Exec(sql); err != nil {
			log.Fatalf("增量迁移失败 (resource unbind migrations): %v\nSQL: %s", err, sql)
		}
	}

	log.Println("✅ 数据库迁移完成")
}

