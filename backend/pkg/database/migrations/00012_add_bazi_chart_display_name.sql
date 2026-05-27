-- +goose Up

ALTER TABLE bazi_charts
ADD COLUMN IF NOT EXISTS display_name TEXT;
