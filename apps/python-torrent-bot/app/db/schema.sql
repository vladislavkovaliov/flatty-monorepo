-- ponytail: идемпотентный schema.sql вместо Alembic (4 таблицы);
-- ceiling: нет миграций схемы — upgrade path: Alembic, когда схема начнёт часто меняться.

CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    telegram_user_id BIGINT NOT NULL UNIQUE,
    username TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS torrents (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    magnet TEXT NOT NULL,
    info_hash TEXT NOT NULL,
    name TEXT NOT NULL,
    media_type TEXT NOT NULL,
    quality TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'queued',
    progress DOUBLE PRECISION NOT NULL DEFAULT 0,
    size_bytes BIGINT,
    download_speed BIGINT NOT NULL DEFAULT 0,
    upload_speed BIGINT NOT NULL DEFAULT 0,
    eta_seconds BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at TIMESTAMPTZ,
    -- duplicate detection как constraint (design §Components)
    UNIQUE (user_id, info_hash)
);

CREATE TABLE IF NOT EXISTS torrent_events (
    id BIGSERIAL PRIMARY KEY,
    torrent_id BIGINT NOT NULL REFERENCES torrents(id) ON DELETE CASCADE,
    event TEXT NOT NULL,
    payload TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_torrent_events_torrent_id ON torrent_events (torrent_id);

-- таблица и repository есть; UI настроек нет (не в Phase 1)
CREATE TABLE IF NOT EXISTS settings (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    PRIMARY KEY (user_id, key)
);
