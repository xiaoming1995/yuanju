-- +goose Up

ALTER TABLE compatibility_readings
  ADD COLUMN IF NOT EXISTS consulting_assessment JSONB NOT NULL DEFAULT '{}'::jsonb;

ALTER TABLE compatibility_evidences
  ADD COLUMN IF NOT EXISTS evidence_key TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_compatibility_evidences_key
  ON compatibility_evidences(reading_id, evidence_key);
