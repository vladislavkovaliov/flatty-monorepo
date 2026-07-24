CREATE TABLE user_settings (
    user_id     TEXT        PRIMARY KEY REFERENCES "user"(id) ON DELETE CASCADE,
    language    TEXT        NOT NULL DEFAULT 'en',
    theme       TEXT        NOT NULL DEFAULT 'system',
    timezone    TEXT        NOT NULL DEFAULT 'UTC+0',
    date_format TEXT        NOT NULL DEFAULT 'YYYY-MM-DD',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TRIGGER update_user_settings_updated_at
    BEFORE UPDATE ON user_settings
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();