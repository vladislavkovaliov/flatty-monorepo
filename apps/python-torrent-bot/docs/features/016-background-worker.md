# FEAT-016 — Background Worker

## Status

Planned

## Goal

Убрать опрос torrent-клиента из Telegram handlers: фоновый worker опрашивает состояние, обновляет database и шлёт уведомления.

> **Phase gate:** required для Phase 2+ (реальные загрузки).

## User Story

As a user, I want download status to update reliably without my messages spamming the torrent client.

## Scope

- Background worker:
  - получает torrent state;
  - обновляет database;
  - отправляет Telegram notifications;
  - обрабатывает завершённые torrents (триггерит Media Organizer).

## Out of Scope

- Обновление Telegram message каждую секунду (**запрещено**).

## User Flow

```text
Telegram Bot → Database → Background Worker → qBittorrent
```

## UI

Нет собственного UI; источник данных для `/downloads` — database.

## Domain Changes

`torrent_events` — история изменений.

## API / Interfaces

Worker читает provider/database, пишет в database; handlers читают database.

## Storage Changes

Database как единственный источник актуального состояния.

## Security

Секреты провайдера — только в env.

## Error Handling

Сбои опроса → retry/backoff, статус не сбрасывать ложно.

## Tests

Unit: цикл обновления состояния (mock provider + DB).

## Dependencies

- FEAT-010
- FEAT-011

## Configuration

Интервалы опроса, `DATABASE_URL`.

## Implementation Tasks

- [ ] Worker loop
- [ ] Обновление database из provider state
- [ ] Уведомления о завершении
- [ ] Тесты

## Acceptance Criteria

- [ ] Handlers не опрашивают qBittorrent
- [ ] Состояние в DB обновляется фоново
- [ ] Telegram messages не спамятся каждый тик

## Related Documentation

- `docs/ARCHITECTURE.md`
