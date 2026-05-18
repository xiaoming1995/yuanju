-- +goose Up
-- syntax error: missing table name on purpose
SELECT * FROM ;

-- +goose Down
SELECT 1;
