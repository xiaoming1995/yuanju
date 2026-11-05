-- +goose Up
ALTER TABLE today_question_requests
    ADD COLUMN quota_source VARCHAR(10) NOT NULL DEFAULT 'none'
    CHECK (quota_source IN ('none', 'daily', 'credit'));

UPDATE today_question_requests SET quota_source='daily' WHERE quota_charged;

CREATE TABLE today_question_credit_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    request_id UUID,
    admin_id UUID,
    kind VARCHAR(10) NOT NULL CHECK (kind IN ('grant', 'spend', 'refund')),
    delta INT NOT NULL CHECK (delta <> 0),
    balance_after INT NOT NULL CHECK (balance_after >= 0),
    reason TEXT NOT NULL DEFAULT '',
    idempotency_key UUID UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK ((kind='grant' AND delta>0 AND admin_id IS NOT NULL AND request_id IS NULL AND idempotency_key IS NOT NULL)
        OR (kind='spend' AND delta=-1 AND request_id IS NOT NULL AND admin_id IS NULL AND idempotency_key IS NULL)
        OR (kind='refund' AND delta=1 AND request_id IS NOT NULL AND admin_id IS NULL AND idempotency_key IS NULL))
);
CREATE INDEX idx_today_question_credit_entries_user_time
    ON today_question_credit_entries (user_id, created_at DESC, id DESC);
CREATE UNIQUE INDEX idx_today_question_credit_entries_request_kind
    ON today_question_credit_entries (request_id, kind) WHERE request_id IS NOT NULL;

-- +goose Down
DROP TABLE today_question_credit_entries;
ALTER TABLE today_question_requests DROP COLUMN quota_source;
