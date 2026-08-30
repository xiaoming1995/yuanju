-- +goose Up

CREATE TABLE IF NOT EXISTS mingge_historical_figures (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	ming_ge VARCHAR(32) NOT NULL,
	figure_name VARCHAR(128) NOT NULL,
	era VARCHAR(128) NOT NULL,
	identity VARCHAR(255) NOT NULL,
	historical_memory TEXT NOT NULL,
	turning_point TEXT NOT NULL DEFAULT '',
	turning_point_year VARCHAR(64) NOT NULL DEFAULT '',
	source_title VARCHAR(255) NOT NULL,
	source_url TEXT NOT NULL,
	birth_data_precision VARCHAR(20) NOT NULL DEFAULT 'unknown'
		CHECK (birth_data_precision IN ('unknown', 'date_only', 'exact_hour')),
	bazi_verification_note TEXT NOT NULL DEFAULT '',
	dayun_period VARCHAR(128) NOT NULL DEFAULT '',
	dayun_explanation TEXT NOT NULL DEFAULT '',
	show_dayun BOOLEAN NOT NULL DEFAULT false,
	display_order INTEGER NOT NULL DEFAULT 0,
	review_status VARCHAR(20) NOT NULL DEFAULT 'draft'
		CHECK (review_status IN ('draft', 'reviewed', 'published', 'archived')),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	UNIQUE (ming_ge, figure_name)
);

CREATE INDEX IF NOT EXISTS idx_mingge_historical_figures_public
	ON mingge_historical_figures (ming_ge, review_status, display_order, created_at);
