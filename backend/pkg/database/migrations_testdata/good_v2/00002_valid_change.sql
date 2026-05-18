-- +goose Up
CREATE TABLE test_bar (
    id SERIAL PRIMARY KEY,
    foo_id INT REFERENCES test_foo(id)
);

-- +goose Down
DROP TABLE IF EXISTS test_bar;
