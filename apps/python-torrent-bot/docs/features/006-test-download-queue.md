# FEAT-006 — Test Download Queue

## Status

Implemented

## Goal

Создать очередь загрузок, работающую полностью в TEST_MODE без реальных торрентов — проверить UX и архитектурный workflow.

## User Story

As a user, I want to confirm a download and see it appear in a simulated queue, so I can test the whole flow safely.

## Scope

- `app/providers/fake.py` — `FakeTorrentProvider`.
- Интерфейс `TorrentProvider` (Protocol).
- `app/services/torrent_service.py` — создание task, валидация, вызов provider.
- Симуляция прогресса загрузки.

## Out of Scope

- qBittorrent, NAS, реальные downloads, запись файлов.

## User Flow

```text
Confirmation → Fake Torrent Service → Test Queue → /downloads
```

## UI

Пример отображения симулируемой загрузки:

```text
⏳ Downloading

The Matrix

██████░░░░░░░░░░ 35%

35%

Speed: TEST
ETA: TEST

🧪 TEST MODE
```

## Domain Changes

- `TorrentStatus` enum: `queued`, `downloading`, `paused`, `completed`, `failed`, `cancelled`.
- `TorrentTask`: `id`, `user_id`, `magnet`, `info_hash`, `name`, `media_type`, `quality`, `status`, `progress`, `size_bytes`, `download_speed`, `upload_speed`, `eta_seconds`, `created_at`, `completed_at`.

## API / Interfaces

```python
class TorrentProvider(Protocol):
    async def add_magnet(self, magnet: str, save_path: str) -> str: ...
    async def get_status(self, torrent_id: str) -> TorrentStatus: ...
    async def pause(self, torrent_id: str) -> None: ...
    async def resume(self, torrent_id: str) -> None: ...
    async def delete(self, torrent_id: str, delete_files: bool = False) -> None: ...
```

`FakeTorrentProvider`:

- не делает HTTP-запросов;
- не подключается к qBittorrent;
- не обращается к NAS;
- не создаёт реальные torrent downloads;
- симулирует очередь и прогресс.

`TorrentService` — без Telegram-кода.

## Storage Changes

На MVP — MemoryStorage / in-memory. Архитектура не должна мешать добавить database (FEAT-010).

## Security

- Не доверять данным пользователя (название и т.п.).
- Дубликаты определяются по `info_hash`.

## Error Handling

- Ошибки provider → статус `FAILED`, сообщение пользователю.
- Невалидная задача → отклонение до вызова provider.

## Tests

- Unit: `TorrentService` — создание task, валидация, duplicate detection.
- Unit: `FakeTorrentProvider` не делает сетевых вызовов.
- Integration: confirmation → queue.

## Dependencies

- FEAT-005
- FEAT-018 (duplicate detection — алгоритм по `info_hash`)

## Configuration

`TEST_MODE=true` — обязателен; при `false` на MVP реальные provider не подключаются.

## Implementation Tasks

- [x] Интерфейс `TorrentProvider` (Protocol)
- [x] `FakeTorrentProvider` с симуляцией прогресса
- [x] `TorrentService` (создание, валидация, duplicates)
- [x] Интеграция с confirmation flow
- [x] Тесты

## Acceptance Criteria

- [x] После подтверждения task появляется в очереди
- [x] Прогресс/статус симулируются, маркер TEST MODE виден
- [x] Никаких сетевых запросов и файловых операций
- [x] `ruff`, `mypy`, `pytest` чистые

## Related Documentation

- `docs/API.md`
- `docs/ARCHITECTURE.md`
- `docs/features/018-duplicate-detection.md`
