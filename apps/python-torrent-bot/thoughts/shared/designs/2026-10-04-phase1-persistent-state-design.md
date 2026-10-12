---
date: 2026-10-04
topic: "Phase 1 Persistent State (PostgreSQL)"
status: validated
---

# Phase 1 — Persistent State на PostgreSQL (FEAT-010)

## Problem Statement

Phase 0 работает, но всё состояние (users, torrent tasks) живёт в памяти: рестарт бота теряет связь task ↔ torrent, duplicate detection и ownership не гарантированы. Phase 1 переводит состояние в PostgreSQL по DATABASE_URL пользователя (`192.168.1.35:5432`), не переписывая Telegram UI.

## Constraints

- Источник требований: `docs/features/010-persistent-state.md`, `docs/ARCHITECTURE.md` §8, `docs/CONFIGURATION.md`.
- UI/handlers не должны знать о database — всё через repository за `TorrentService` (ADR-0001).
- История бизнес-состояния — в database, не в Redis.
- Unit tests не требуют внешней инфраструктуры; секреты (DATABASE_URL с паролем) не логируются и не коммитятся.
- `TEST_MODE=true` остаётся — реальных загрузок по-прежнему нет.
- DoD: `ruff check`, `ruff format --check`, `mypy app`, `pytest` — зелёные.

## Approach

**Драйвер: psycopg 3 (async, `psycopg[binary]`).** Обоснование: пользовательский URL (`postgresql://...?sslmode=disable`) — формат libpq, psycopg принимает его без изменений; async-first под aiogram; без ORM-слоя. Отклонил SQLAlchemy+Alembic — 4 таблицы не оправдывают ORM, mypy-сложность и второй слой абстракции (upgrade path: Alembic появится, когда схема начнёт часто меняться). Отклонил asyncpg — жёстче работа с sslmode-URL и ручной маппинг.

**Схема: `app/db/schema.sql`**, применяется идемпотентно при старте (`CREATE TABLE IF NOT EXISTS`). Alembic не тащим (см. ponytail-комментарий в плане).

**FSM остаётся на MemoryStorage** — wizard живёт секунды, Redis/DB-storage не нужен (явное сознательное сокращение против чек-листа Phase 1; легко добавить позже).

## Architecture

```text
handlers → TorrentService → Repository (interface)
                               ├── PostgresRepository (psycopg pool)   [default]
                               └── InMemoryRepository (test double)    [unit tests]
```

- Пул psycopg создаётся в startup, закрывается в shutdown; repository инжектится в `TorrentService`.
- `FakeTorrentProvider` остаётся — меняется только слой хранения задач.

## Components

- `app/db/__init__.py`, `app/db/pool.py` — пул, startup/shutdown lifecycle
- `app/db/schema.sql` — users, torrents, torrent_events, settings
- `app/repositories/__init__.py` — протокол/базовый интерфейс
- `app/repositories/postgres.py` — реализация на psycopg
- `app/repositories/memory.py` — in-memory double для unit/integration-без-БД
- `app/config.py` — `DATABASE_URL` (валидация при старте, fail fast)
- `app/main.py` — lifespan: pool create → schema apply → bot run
- Правки: `torrent_service.py` (работает через repository), `bot.py` (DI), `models/torrent.py` (маппинг строк ↔ enum)

### Схема (MVP-минимум FEAT-010)

- `users(id, telegram_user_id UNIQUE, username, created_at)`
- `torrents(id, user_id FK, magnet, info_hash, name, media_type, quality, status, progress, size_bytes, download_speed, upload_speed, eta_seconds, created_at, completed_at)` + **UNIQUE(user_id, info_hash)** — duplicate detection становится constraint'ом
- `torrent_events(id, torrent_id FK, event, payload, created_at)` — lifecycle-события (created/status changes), НЕ каждый тик прогресса
- `settings(user_id FK, key, value)` — таблица и repository, UI настройок нет (не в Phase 1)
- `media_items` — НЕ создаётся (FEAT-013)

## Data Flow

`create(user, magnet…)` → upsert user → `INSERT torrents … ON CONFLICT (user_id, info_hash)` → существующий? показать дубликат → event `created` → UI без изменений. Остальные операции (list/pause/resume/cancel/delete) — UPDATE + event.

## Error Handling

- `DATABASE_URL` отсутствует / БД недоступна при старте → fail fast с внятным сообщением (никакого тихого memory-режима — один режим работы).
- Ошибки запросов → лог error (с redacted URL), пользователю — сообщение, бот не падает.
- Логировать DATABASE_URL запрещено (содержит пароль) — редактировать до host/db.

## Testing Strategy

- **Unit (без инфраструктуры):** service + in-memory repository double — create/duplicate/ownership/status, все прежние 20 тестов остаются зелёными.
- **Integration:** `tests/integration/test_db.py` против реального `DATABASE_URL` (schema apply, CRUD, UNIQUE constraint, events) — **авто-skip, если БД недоступна**, чтобы `pytest` проходил и в CI.
- Прежний полный UI-flow тест остаётся герметичным (in-memory double).
- Verification: `ruff check .`, `ruff format --check .`, `mypy app`, `pytest`.

## Documentation & Git

- `git init` + первый коммит (Phase 0) перед изменением зависимостей — задача 0 плана.
- Обновить: `docs/CONFIGURATION.md` (DATABASE_URL), `docs/ARCHITECTURE.md` §8 (DB реализована), `docs/features/010` → In Progress/Implemented, `docs/TESTING.md` (DB-тесты), `.env.example`.
- Рекомендация пользователю (в инструкции): выделить отдельную БД (`torrent_bot`) вместо `public` — опционально.

## Open Questions

Нет. Технология (psycopg3) и объём согласованы с указанием пользователя «у меня БД».
