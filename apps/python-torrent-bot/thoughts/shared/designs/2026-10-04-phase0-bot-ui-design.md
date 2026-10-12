---
date: 2026-10-04
topic: "Phase 0 Telegram Bot UI (FEAT-001-009)"
status: validated
---

# Phase 0 — Telegram Bot UI без реальной загрузки

## Problem Statement

Реализовать Phase 0 (FEAT-001–009): полный пользовательский flow в Telegram — `/start`, парсинг magnet, выбор категории/качества, подтверждение, тестовая очередь, `/downloads` с управлением — без реальных торрент-загрузок, чтобы проверить UX и архитектурный workflow в реальном Telegram.

## Constraints

- `TEST_MODE=true`: запрещены подключение к qBittorrent, реальные downloads, доступ к NAS, запись/перемещение файлов.
- Runtime dependencies: только `aiogram`, `python-dotenv`.
- Python 3.12+, aiogram 3.x, FSM на `MemoryStorage` (без database — FEAT-010 позже).
- Слоистая архитектура по ADR-0001: handlers без бизнес-логики, всё через `TorrentService` + `TorrentProvider`.
- Дубликаты — по `info_hash` (ADR-0002, FEAT-018), изоляция по `user_id` (FEAT-009).
- DoD: `ruff check`, `ruff format --check`, `mypy app`, `pytest` — успешно.
- Источник требований: `docs/features/001-start-menu.md` … `009-user-isolation.md`, `docs/ARCHITECTURE.md`, `docs/API.md`.

## Approach

Один процесс aiogram-бота с in-memory очередью через `FakeTorrentProvider`. Никаких упрощений «прямо в хендлерах» — слои обязательны, чтобы Phase 2 (qBittorrent) не переписывал UI. Состояния пользователя — только FSM, глобальные переменные запрещены.

## Architecture

```text
Telegram
  → handlers (команды, FSM, рендер, keyboards)
    → TorrentService (валидация, duplicate detection, ownership, orchestration)
      → TorrentProvider (Protocol)
        → FakeTorrentProvider (in-memory, симуляция прогресса/статусов)
```

## Components

- `app/models/media.py` — enum `MediaType`, `Quality`
- `app/models/torrent.py` — enum `TorrentStatus`, dataclass/модель `TorrentTask`
- `app/services/magnet_parser.py` — `MagnetInfo`, валидация `magnet:?` + info hash, URL decode
- `app/services/torrent_service.py` — create/list/get/pause/resume/cancel/delete, duplicate по `info_hash`, фильтр по `user_id`
- `app/providers/base.py` — `TorrentProvider` Protocol
- `app/providers/fake.py` — in-memory реализация, симуляция прогресса, маркер TEST MODE
- `app/handlers/start.py`, `magnet.py`, `downloads.py`, `settings.py` (Настройки/Помощь — заглушки)
- `app/keyboards/main.py`, `torrent.py`, `downloads.py` — callbacks из whitelist `torrent:*`
- `app/config.py`, `app/bot.py`, `app/main.py` — env, сборка диспетчера, запуск
- `pyproject.toml`, `.env.example`, `requirements.txt`

## Data Flow

`/start` → пользователь шлёт magnet → парсер → показ метаданных (FEAT-003 UI) → категория → качество → сводка + TEST MODE notice → `confirm:add` → `TorrentService.create` (info_hash duplicate → показ существующего; иначе новый task с `user_id`) → fake queue → `/downloads` (только свои) → pause/resume/cancel/delete с ownership check.

## Error Handling

- Невалидный magnet (нет scheme/info hash) → дружелюбное сообщение, FSM сброс.
- Неизвестный callback → игнор без падения.
- Чужая/несуществующая задача → трактовать как «не найдено».
- Ошибки provider → статус `FAILED`, сообщение пользователю.
- Непредвиденное исключение в хендлере → лог error, ответ пользователю, бот жив.

## Testing Strategy

- Unit: 7 обязательных кейсов magnet parser (valid, missing dn, multi trackers, URL-encoded dn, missing hash, invalid scheme, xl); `TorrentService` — create/duplicate/ownership/status transitions; fake provider — без сетевых вызовов.
- Integration: полный flow `/start → magnet → category → quality → confirmation → queue → /downloads` (aiogram test machinery / мок-апдейты).
- Lint/type: `ruff check .`, `ruff format --check .`, `mypy app`.
- Manual: пошаговая инструкция тестирования в Telegram → `docs/TESTING.md`, раздел «Ручное тестирование».

## Open Questions

Нет. Объём = FEAT-001–009 из docs/.
