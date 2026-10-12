# python-torrent-bot

Telegram-бот для торрент-загрузок: отправили magnet-ссылку — получили задачу в очереди,
список загрузок с управлением (пауза / продолжить / отмена / удалить).
Слоёная архитектура: handlers → services → repositories/providers. Сейчас реализованы
Phase 0 (Telegram UI) и Phase 1 (PostgreSQL): загрузки имитирует `FakeTorrentProvider`,
реальный протокол с qBittorrent — отдельная будущая фаза.

## Требования

- Python **3.12+**
- PostgreSQL — только для реального запуска (тесты без недоступной БД пропускаются)
- Токен Telegram-бота от [@BotFather](https://t.me/BotFather)

## Установка и запуск

```bash
# из корня монорепо
nx setup python-torrent-bot        # .venv + зависимости (разово)
cp apps/python-torrent-bot/.env.example apps/python-torrent-bot/.env
# вписать BOT_TOKEN; DATABASE_URL уже указывает на torrent_bot в Postgres монорепо (:55000)
nx serve python-torrent-bot
```

- `DATABASE_URL` — в `.env.example` уже указан общий Postgres монорепо
  (`localhost:55000/torrent_bot`, БД создаёт `init-db`). Без него бот не стартует (fail-fast).
- Пароль БД в логах маскируется; логи также пишутся в `logs/bot.log` (ротация 2 МБ × 3).

## HTTP API

Локальный REST API с Swagger UI: `http://127.0.0.1:8090/docs`
(OpenAPI-схема — `/openapi.json`). Работает от имени пользователя `API_USER_ID`
(см. [docs/CONFIGURATION.md](docs/CONFIGURATION.md)); без аутентификации — слушает
только localhost. Перед первым `POST /api/tasks` нажмите `/start` в боте.
Контракт: [docs/API.md](docs/API.md) §7.

## Тестирование

### Автотесты

```bash
nx test python-torrent-bot   # 49 тестов: unit + integration; DB-тесты skip'аются, если PostgreSQL недоступен
nx lint python-torrent-bot   # ruff
nx typecheck python-torrent-bot  # mypy
```

### Ручное тестирование в Telegram

Полный сценарий (21 шаг) — [docs/TESTING.md](docs/TESTING.md). Краткий чек-лист при `TEST_MODE=true`:

1. `/start` → главное меню (Мои загрузки / Настройки / Помощь).
2. Отправьте magnet (пример):
   ```text
   magnet:?xt=urn:btih:a1b2c3d4e5f60718293a4b5c6d7e8f9012345678&dn=The.Matrix.1999.1080p.BluRay&tr=udp%3A%2F%2Ftracker.example.org%3A1337%2Fannounce&tr=udp%3A%2F%2Fopen.demonii.com%3A1337%2Fannounce&xl=12884901888
   ```
3. Экран «🔗 Torrent найден» → `🎬 Фильм` → `1080p`.
4. В сводке нажмите `⬇️ Добавить в очередь` (повтор того же magnet даёт «Такая загрузка уже есть»).
5. `/downloads` → задача `#001` со статусом `🕓 В очереди`.
6. Повторяйте `/downloads` — прогресс растёт до `✅ Завершено`.
7. `⏸ Пауза` / `▶️ Продолжить` / `❌ Отменить` / `🗑 Удалить` — статусы меняются.
8. `Ctrl+C`, снова `nx serve python-torrent-bot`, `/downloads` — задача сохранилась (персистентность).

## Структура проекта

```text
app/
├── main.py          # точка входа: конфиг, БД, polling
├── bot.py           # создание бота/диспетчера
├── config.py        # env-переменные
├── handlers/        # Telegram handlers (commands, callbacks)
├── keyboards/       # inline-клавиатуры
├── services/        # бизнес-логика (magnet parser, torrent service)
├── providers/       # FakeTorrentProvider (заглушка), базовый интерфейс
├── repositories/    # доступ к БД (Postgres / InMemory)
├── models/          # domain-модели и enum
└── db/              # пул psycopg, schema.sql
tests/
├── unit/            # без внешней инфраструктуры
└── integration/     # bot flow + PostgreSQL (skip без БД)
docs/                # документация и ADR
```

## Документация

- [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) — development contract
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) — архитектура
- [docs/TESTING.md](docs/TESTING.md) — тестирование
- [docs/CONFIGURATION.md](docs/CONFIGURATION.md) — переменные окружения
- [docs/ROADMAP.md](docs/ROADMAP.md) — этапы
- [docs/decisions/](docs/decisions/) — ADR
