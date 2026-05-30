-- +goose Up

-- 用户禁用标记：NULL = 正常，非空 = 已禁用（禁用时间）
ALTER TABLE users ADD COLUMN IF NOT EXISTS disabled_at TIMESTAMPTZ;
