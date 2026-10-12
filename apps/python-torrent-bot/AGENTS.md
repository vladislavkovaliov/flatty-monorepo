# python-torrent-bot

Telegram-бот для торрент-загрузок: magnet-ссылка → задача в очереди, список с управлением
(пауза / продолжить / отмена / удалить). Внутри — локальный HTTP API со Swagger.

- **Порт:** 8090 (HTTP API, только localhost, без аутентификации)
- **Стек:** Python 3.12, aiogram 3, FastAPI, PostgreSQL (psycopg 3)
- **Entry:** `python -m app.main`
- **Режим:** Phase 0/1 — `TEST_MODE=true` обязателен, реальный qBittorrent не подключён

## Commands

| Command | Description |
|---------|-------------|
| `nx setup python-torrent-bot` | Создать `.venv` и установить зависимости (разово) |
| `nx serve python-torrent-bot` | Запустить бота (polling + API на :8090) |
| `nx test python-torrent-bot` | pytest (unit + integration; DB-тесты skip без PostgreSQL) |
| `nx lint python-torrent-bot` | ruff check |
| `nx typecheck python-torrent-bot` | mypy (strict) |

## Configuration

Создайте `apps/python-torrent-bot/.env` из `.env.example`:

- `BOT_TOKEN` — от [@BotFather](https://t.me/BotFather)
- `DATABASE_URL` — общий Postgres монорепо: `postgresql://postgres:password@localhost:55000/torrent_bot?sslmode=disable`
  (БД `torrent_bot` создаётся автоматически через `docker compose up postgres init-db`)
- `API_PORT=8090` — порт свободен; 8080 занят go-api
- `TEST_MODE=true` — fail-fast, если false (реальные провайдеры не реализованы)

Подробности: [docs/CONFIGURATION.md](docs/CONFIGURATION.md).

## Testing

```bash
nx test python-torrent-bot   # pytest; integration с БД skip'ается, если PostgreSQL недоступен
nx lint python-torrent-bot
nx typecheck python-torrent-bot
```

Ручной сценарий в Telegram — [docs/TESTING.md](docs/TESTING.md).
Перед первым `POST /api/tasks` нажмите `/start` в боте. Swagger: `http://127.0.0.1:8090/docs`.

## Docs & Thoughts

Полная документация (architecture, ADR, features, roadmap) — [docs/](docs/).
Дизайны/планы сессий и continuity-ledgers — [thoughts/](thoughts/).
Исходный спекулятив проекта — [init.md](init.md).
