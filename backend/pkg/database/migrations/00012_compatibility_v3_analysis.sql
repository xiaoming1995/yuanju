-- +goose Up

-- Compatibility v3: rename rationale comment + add overall_score column.
-- The new 100-point scoring formula stores per-module breakdown in the
-- existing dimension_scores JSONB column with new keys (zodiac/nayin/
-- day_pillar/eight_chars); analysis_version='v3' marks the new format.

ALTER TABLE compatibility_readings
	ADD COLUMN IF NOT EXISTS overall_score INTEGER NOT NULL DEFAULT 0;

COMMENT ON COLUMN compatibility_readings.analysis_version IS
	'v1/v2 = legacy 4-dim evidence-weighted scoring (attraction/stability/communication/practicality); v3 = zodiac/nayin/day_pillar/eight_chars 100-point classical formula';
COMMENT ON COLUMN compatibility_readings.overall_score IS
	'v3 only: total 0–100 score = sum of 4 module scores stored in dimension_scores JSONB. For v1/v2 records the column is 0.';
