-- 006_add_applications.sql
-- Self-contained: ensures the shared updated_at trigger function exists, then
-- creates the applications registry table + development seed rows.
-- Idempotent: safe to re-run.

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE IF NOT EXISTS applications (
    id              BIGSERIAL PRIMARY KEY,
    name            TEXT        NOT NULL,
    env             TEXT        NOT NULL,
    bundle_js       TEXT        NOT NULL,
    style_url       TEXT,
    remote_origin   TEXT        NOT NULL,
    proxy_base_path TEXT        NOT NULL,
    base_path       TEXT        NOT NULL DEFAULT '/',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (name, env)
);

DROP TRIGGER IF EXISTS set_applications_updated_at ON applications;
CREATE TRIGGER set_applications_updated_at
    BEFORE UPDATE ON applications
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

INSERT INTO applications (name, env, bundle_js, style_url, remote_origin, proxy_base_path, base_path)
VALUES
    ('app',      'development', 'app',      'styles', 'http://localhost:8080', '/external-app',      '/'),
    ('settings', 'development', 'settings', 'styles', 'http://localhost:8081', '/external-settings', '/'),
    ('resident', 'development', 'resident', 'styles', 'http://localhost:8082', '/external-resident', '/')
ON CONFLICT (name, env) DO NOTHING;
