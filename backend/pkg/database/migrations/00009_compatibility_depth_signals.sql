-- +goose Up
-- +goose StatementBegin
ALTER TABLE compatibility_readings
    ADD COLUMN IF NOT EXISTS score_explanations JSONB NOT NULL DEFAULT '[]'::jsonb;

ALTER TABLE compatibility_evidences
    ADD COLUMN IF NOT EXISTS perspective VARCHAR(40),
    ADD COLUMN IF NOT EXISTS actor VARCHAR(30),
    ADD COLUMN IF NOT EXISTS target VARCHAR(30),
    ADD COLUMN IF NOT EXISTS related_sources JSONB NOT NULL DEFAULT '[]'::jsonb;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE compatibility_evidences
    DROP COLUMN IF EXISTS related_sources,
    DROP COLUMN IF EXISTS target,
    DROP COLUMN IF EXISTS actor,
    DROP COLUMN IF EXISTS perspective;

ALTER TABLE compatibility_readings
    DROP COLUMN IF EXISTS score_explanations;
-- +goose StatementEnd
