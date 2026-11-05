-- +goose Up
CREATE TABLE today_question_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    chart_id UUID NOT NULL,
    question_date DATE NOT NULL,
    category VARCHAR(30) NOT NULL,
    question TEXT NOT NULL,
    context_data JSONB NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'processing'
        CHECK (status IN ('processing', 'success', 'insufficient', 'error')),
    answer_data JSONB,
    ai_system_prompt TEXT NOT NULL DEFAULT '',
    ai_user_prompt TEXT NOT NULL DEFAULT '',
    ai_output TEXT NOT NULL DEFAULT '',
    provider_id UUID REFERENCES llm_providers(id) ON DELETE SET NULL,
    model VARCHAR(100) NOT NULL DEFAULT '',
    duration_ms INT NOT NULL DEFAULT 0,
    prompt_tokens INT NOT NULL DEFAULT 0,
    completion_tokens INT NOT NULL DEFAULT 0,
    total_tokens INT NOT NULL DEFAULT 0,
    reasoning_tokens INT NOT NULL DEFAULT 0,
    cache_hit_tokens INT NOT NULL DEFAULT 0,
    cache_miss_tokens INT NOT NULL DEFAULT 0,
    estimated_cost_cny NUMERIC(12, 6) NOT NULL DEFAULT 0,
    error_msg TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at TIMESTAMPTZ
);
CREATE INDEX idx_today_question_requests_created ON today_question_requests (created_at DESC, id DESC);
CREATE INDEX idx_today_question_requests_user_created ON today_question_requests (user_id, created_at DESC);
CREATE INDEX idx_today_question_requests_date_status ON today_question_requests (question_date DESC, status);

-- +goose Down
DROP TABLE IF EXISTS today_question_requests;
