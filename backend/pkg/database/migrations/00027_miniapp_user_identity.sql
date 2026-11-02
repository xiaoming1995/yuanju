-- +goose Up

ALTER TABLE users ALTER COLUMN email DROP NOT NULL;
ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;
ALTER TABLE users ADD CONSTRAINT users_email_password_together
    CHECK ((email IS NULL AND password_hash IS NULL) OR (email IS NOT NULL AND password_hash IS NOT NULL));

CREATE TABLE wechat_user_identities (
    app_id VARCHAR(64) NOT NULL,
    open_id VARCHAR(128) NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (app_id, open_id),
    UNIQUE (app_id, user_id)
);

INSERT INTO system_settings (key, value)
VALUES ('miniapp_registration_enabled', 'true')
ON CONFLICT (key) DO NOTHING;

-- +goose Down

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM users WHERE email IS NULL OR password_hash IS NULL) THEN
        RAISE EXCEPTION 'cannot roll back miniapp identity migration while email-less users exist';
    END IF;
END $$;
DELETE FROM system_settings WHERE key = 'miniapp_registration_enabled';
DROP TABLE IF EXISTS wechat_user_identities;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_password_together;
ALTER TABLE users ALTER COLUMN email SET NOT NULL;
ALTER TABLE users ALTER COLUMN password_hash SET NOT NULL;
