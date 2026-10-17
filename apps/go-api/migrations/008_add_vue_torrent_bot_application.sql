-- 008_add_vue_torrent_bot_application.sql
-- Registers the Vue micro-frontend (apps/vue-torrent-bot) in the applications
-- registry so react-entrypoint proxies /external-vue-torrent-bot/* to it.
-- Idempotent: safe to re-run.

-- Drop the pre-rename row that may have been registered manually (admin UI), so the
-- registry keeps a single canonical entry. Without this, the host would keep proxying
-- /external-vue-remote/* and the new /external-vue-torrent-bot/* would 404.
DELETE FROM applications WHERE name IN ('vue_remote', 'vue-remote');

INSERT INTO applications (name, env, bundle_js, style_url, remote_origin, proxy_base_path, base_path)
VALUES
    ('vue-torrent-bot', 'development', 'vue-torrent-bot', 'styles', 'http://localhost:8084', '/external-vue-torrent-bot', '/')
ON CONFLICT (name, env) DO NOTHING;
