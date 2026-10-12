# FEAT-010 — Persistent State

## Status

Implemented

## Goal

Перевести состояние пользователей и torrent tasks из in-memory в database, не переписывая Telegram UI.

> **Phase gate:** не начинать, пока пользователь явно не переведёт проект на Phase 1.

## User Story

As a user, I want my downloads and settings to survive a bot restart, so nothing is lost.

## Scope

- Database + entities: `users`, `torrents`, `torrent_events`, `settings` (`media_items` — отложено в FEAT-013).
- Persistent FSM storage — НЕ в Phase 1: FSM остаётся `MemoryStorage` (design §Approach).

## Out of Scope

- Реальные загрузки (Phase 2+).

## User Flow

```text
State change
  → Application Service
  → Database
```

## UI

Не меняется — только источник данных.

## Domain Changes

Маппинг `TorrentTask` → таблица `torrents`; история бизнес-состояния в `torrent_events`.

## API / Interfaces

Repository/storage слой за сервисами; UI не знает о database напрямую.

## Storage Changes

- Database обязательна.
- **История бизнес-состояния хранится в database, а не в Redis.**

## Security

- Секреты (`DATABASE_URL`, `REDIS_PASSWORD`) — только в `.env`, никогда в git и логах.

## Error Handling

Сбои БД → понятная ошибка пользователю, без падения бота.

## Tests

Unit + integration с тестовой БД; не требовать внешней инфраструктуры в unit tests.

## Dependencies

- FEAT-001 — FEAT-009 (MVP zone).

## Configuration

`DATABASE_URL`, опционально `REDIS_URL`.

## Implementation Tasks

- [x] Схема database (`app/db/schema.sql`: users, torrents, torrent_events, settings)
- [x] Repository layer (protocol `app/repositories/__init__.py` / `PostgresRepository` / `InMemoryRepository`)
- [x] Connection pool + fail-fast + redaction (`app/db/pool.py`), `DATABASE_URL` в config
- [x] DI/lifespan wiring (`app/main.py`)
- [x] Тесты (unit: config/pool/service; integration: `tests/integration/test_db.py` против реального PostgreSQL)
- [ ] Persistent FSM — отложено (NOT в Phase 1, design §Approach)

## Acceptance Criteria

- [x] Состояние переживает рестарт
- [x] Telegram UI не переписан
- [x] History хранится в database

## Related Documentation

- `docs/ARCHITECTURE.md`
- `docs/ROADMAP.md`
