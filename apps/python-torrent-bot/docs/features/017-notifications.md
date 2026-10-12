# FEAT-017 — Notifications

## Status

Planned

## Goal

Сообщать пользователю о значимых событиях загрузки (завершение, ошибка) через Telegram.

> **Phase gate:** реализуется вместе с Background Worker (FEAT-016).

## User Story

As a user, I want to be notified when my download finishes or fails, so I don't have to poll `/downloads`.

## Scope

- Уведомления о: завершении загрузки, ошибке/`FAILED`.
- Отправка от имени worker'а, не из handlers.

## Out of Scope

- Частые обновления прогресса (запрещено редактировать message каждую секунду).

## User Flow

```text
Background Worker
  → Событие (completed / failed)
  → Telegram notification пользователю-владельцу
```

## UI

Отдельное сообщение пользователю.

## Domain Changes

События берутся из `torrent_events` (FEAT-010).

## API / Interfaces

Notification sender поверх бота; вызывается worker'ом.

## Storage Changes

Нет.

## Security

- Отправлять только владельцу задачи (ownership).
- Не включать секреты в текст уведомлений.

## Error Handling

Ошибка отправки → лог, не прерывать цикл worker'а.

## Tests

Unit: событие → корректный текст/получатель (mock sender).

## Dependencies

- FEAT-016
- FEAT-009

## Configuration

Нет.

## Implementation Tasks

- [ ] Шаблоны уведомлений
- [ ] Интеграция с worker'ом
- [ ] Тесты

## Acceptance Criteria

- [ ] Завершённая загрузка порождает уведомление владельцу
- [ ] Нет спама частыми обновлениями

## Related Documentation

- `docs/ARCHITECTURE.md`
