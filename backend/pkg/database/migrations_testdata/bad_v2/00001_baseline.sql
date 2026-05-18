-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS test_foo (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS test_seed_table (
    key VARCHAR(50) PRIMARY KEY,
    value TEXT
);
INSERT INTO test_seed_table (key, value) VALUES ('foo', 'bar') ON CONFLICT (key) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- baseline does not support rollback
SELECT 1;
