# CONFIGURATION

Env variables, defaults, production configuration.

---

## Правила

- Использовать `.env`.
- `.env` никогда не должен попадать в git.
- Обязательно создать `.env.example`.
- Никогда не логировать секреты: `BOT_TOKEN`, `QBITTORRENT_PASSWORD`, `NAS_PASSWORD`, `REDIS_PASSWORD`, `DATABASE_PASSWORD`.

---

## Текущая конфигурация (MVP)

```text
BOT_TOKEN=change_me

APP_ENV=development
LOG_LEVEL=INFO

TEST_MODE=true
DATABASE_URL=postgresql://user:pass@host:5432/dbname?sslmode=disable
API_HOST=127.0.0.1
API_PORT=8080
API_USER_ID=1
```

`DATABASE_URL` обязателен: без него бот падает сразу при старте (fail-fast).
Формат — libpq URL (принимается psycopg как есть, валидация только на схему
`postgresql://`). Пароль никогда не логируется (redaction до `user:***@host`).
Рекомендуется выделить отдельную БД (например `torrent_bot`) вместо `public`.

`LOG_LEVEL` — `DEBUG|INFO|WARNING|ERROR|CRITICAL`, default `INFO`; логи также
пишутся в `logs/bot.log` (ротация 2 МБ × 3).

`TEST_MODE=true` (FakeTorrentProvider):

- запрещает реальные torrent downloads;
- запрещает доступ к NAS и запись файлов;
- запрещает перемещение медиа;
- разрешает только симуляцию загрузки.

`TEST_MODE=false` (FEAT-011, QBittorrentProvider) требует `QBITTORRENT_URL`
(fail-fast: "QBITTORRENT_URL is required when TEST_MODE=false"). Username/password
опциональны (LAN без auth). Пути `SAVE_PATH_MOVIES`/`SAVE_PATH_TV` передаются
qBittorrent как есть — они должны быть видимы с его стороны (NFS/relative его конфига).

```text
QBITTORRENT_URL=http://192.168.1.50:8080
QBITTORRENT_USERNAME=
QBITTORRENT_PASSWORD=
SAVE_PATH_MOVIES=Movies/
SAVE_PATH_TV=TV/
```

`API_HOST`/`API_PORT`/`API_USER_ID` (FEAT-020) — адрес HTTP API (Swagger UI: `http://127.0.0.1:8080/docs`).
`API_HOST` принимает только localhost (`127.0.0.1`/`localhost`/`::1`) — API без аутентификации,
fail-fast при попытке слушать наружу. `API_USER_ID` — telegram id пользователя, **от имени которого
работает API**: create ставит его владельцем задачи, list/get/pause/resume/cancel/delete проверяют
ownership через него (поведение = Telegram). Несовпадение с вашим реальным telegram id = пустой
список задач. Перед первым `POST /api/tasks` пользователь должен нажать `/start` в боте
(регистрация в `bot_users`).

---

## Будущая конфигурация

```text
DOWNLOAD_PATH=/mnt/dns320/Downloads
MOVIES_PATH=/mnt/dns320/Movies
TV_PATH=/mnt/dns320/TV

REDIS_URL=redis://localhost:6379/0
```

Эти переменные вводятся только соответствующими feature (FEAT-012 и др.) — не раньше.
