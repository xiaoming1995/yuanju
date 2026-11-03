-- +goose Up
ALTER TABLE users ADD COLUMN last_login_at TIMESTAMPTZ;
ALTER TABLE users ADD COLUMN last_login_channel VARCHAR(20);
ALTER TABLE users ADD COLUMN last_active_at TIMESTAMPTZ;
CREATE INDEX idx_users_last_login_at ON users(last_login_at DESC);
CREATE INDEX idx_users_last_active_at ON users(last_active_at DESC);

CREATE TABLE user_activity_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_type VARCHAR(40) NOT NULL,
    channel VARCHAR(20) NOT NULL,
    object_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_user_activity_events_user_time ON user_activity_events(user_id, created_at DESC, id DESC);
CREATE INDEX idx_user_activity_events_created ON user_activity_events(created_at);

CREATE TABLE admin_user_audit_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_id UUID NOT NULL,
    target_user_id UUID NOT NULL,
    action VARCHAR(30) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_admin_user_audit_target_time ON admin_user_audit_events(target_user_id, created_at DESC, id DESC);
CREATE INDEX idx_admin_user_audit_created ON admin_user_audit_events(created_at);

-- +goose Down
DROP TABLE IF EXISTS admin_user_audit_events;
DROP TABLE IF EXISTS user_activity_events;
DROP INDEX IF EXISTS idx_users_last_active_at;
DROP INDEX IF EXISTS idx_users_last_login_at;
ALTER TABLE users DROP COLUMN IF EXISTS last_active_at;
ALTER TABLE users DROP COLUMN IF EXISTS last_login_channel;
ALTER TABLE users DROP COLUMN IF EXISTS last_login_at;
