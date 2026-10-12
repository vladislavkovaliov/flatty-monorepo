# FEAT-009 — User Isolation

## Status

Implemented

## Goal

Гарантировать, что каждый Telegram-пользователь видит и управляет только своими загрузками.

## User Story

As a user, I want my downloads to be private and non-manageable by others, so nobody can touch my queue.

## Scope

- Фильтрация задач по `user_id` во всех чтениях.
- Проверка ownership перед каждой операцией (открыть, пауза, resume, cancel, delete).

## Out of Scope

- Полноценная auth/roles (не требуется: Telegram user id — источник идентичности).

## User Flow

```text
Любое действие с загрузкой
  → owner_id == current user_id ?
      YES → выполнить
      NO  → отклонить
```

## UI

Нет отдельного UI — встроено в FEAT-007 / FEAT-008.

## Domain Changes

`TorrentTask.user_id` — обязательное поле, заполняется из Telegram update.

## API / Interfaces

`TorrentService`: все выборки и операции принимают/проверяют `user_id`.

## Storage Changes

Любое будущее хранилище обязано поддерживать фильтрацию по `user_id`.

## Security

- Никогда не выводить чужие `info_hash`, названия, пути.
- Callback не должен позволять манипулировать чужими id (проверять владельца на сервере, а не доверять callback).

## Error Handling

Чужая/несуществующая задача → трактовать как «не найдено» (не раскрывать существование).

## Tests

- Пользователь A не видит задачи пользователя B.
- Операция A над задачей B отклоняется.

## Dependencies

- FEAT-007

## Configuration

Нет.

## Implementation Tasks

- [x] Фильтрация по `user_id` в сервисе
- [x] Ownership check в каждой операции
- [x] Тесты на изоляцию

## Acceptance Criteria

- [x] `/downloads` показывает только свои задачи
- [x] Каждая операция проверяет владельца
- [x] Тесты изоляции проходят

## Related Documentation

- `docs/API.md`
- `docs/DEVELOPMENT.md` (Security)
