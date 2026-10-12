# FEAT-011 — qBittorrent Provider

## Status

Planned

## Goal

Подключить реальный торрент-клиент через `TorrentProvider`, не меняя Telegram UI и service layer.

> **Phase gate:** не начинать, пока пользователь явно не переведёт проект на Phase 2.

## User Story

As a user, I want my confirmed magnet to actually download via qBittorrent, with real progress, speed and ETA.

## Scope

- `app/providers/qbittorrent.py` — `QBittorrentProvider`.
- Операции: add magnet, status, pause, resume, cancel, delete.
- Отображение progress / speed / ETA.

## Out of Scope

- NAS storage (FEAT-012), media organizer (FEAT-013).

## User Flow

```text
Magnet → Parse → category → quality → create task
  → TorrentProvider → QBittorrentProvider → qBittorrent Web API
```

## UI

Статусы/прогресс в `/downloads` — с реальными значениями вместо `TEST`.

## Domain Changes

`TorrentStatus` маппится из состояний qBittorrent.

## API / Interfaces

```text
TorrentService → TorrentProvider → QBittorrentProvider → qBittorrent Web API
```

Telegram layer не знает деталей qBittorrent.

## Storage Changes

Задачи хранятся в database (FEAT-010).

## Security

- `QBITTORRENT_PASSWORD` не логировать.
- Не доверять данным пользователя при формировании запросов.

## Error Handling

Недоступность qBittorrent → статус `FAILED` + понятное сообщение.

## Tests

- Unit: маппинг статусов, контракт provider.
- Интеграционные тесты с реальным qBittorrent — вне CI.

## Dependencies

- FEAT-006 (интерфейс provider)
- FEAT-010 (database)

## Configuration

`QBITTORRENT_URL`, `QBITTORRENT_USERNAME`, `QBITTORRENT_PASSWORD`, `TEST_MODE=false`.

## Implementation Tasks

- [ ] `QBittorrentProvider`
- [ ] Маппинг статусов/прогресса
- [ ] Интеграция с `TorrentService`
- [ ] Тесты

## Acceptance Criteria

- [ ] Через UI запускается реальная загрузка
- [ ] Прогресс/скорость/ETA реальные
- [ ] Handlers не обращаются к qBittorrent напрямую

## Related Documentation

- `docs/API.md`
- `docs/ARCHITECTURE.md`
