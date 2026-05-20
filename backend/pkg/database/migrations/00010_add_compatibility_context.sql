-- +goose Up

ALTER TABLE compatibility_readings
  ADD COLUMN IF NOT EXISTS relationship_stage TEXT NOT NULL DEFAULT 'general',
  ADD COLUMN IF NOT EXISTS primary_question TEXT NOT NULL DEFAULT 'general';

CREATE INDEX IF NOT EXISTS idx_compatibility_readings_context
  ON compatibility_readings(user_id, relationship_stage, primary_question);
