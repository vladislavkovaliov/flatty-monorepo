# FEAT-018 — Duplicate Detection

## Status

Planned

## Goal

Не создавать дублирующиеся torrent tasks: главный идентификатор — `info_hash`.

## User Story

As a user, I want to be shown my existing download when I paste the same magnet again, instead of creating a duplicate.

## Scope

- Алгоритм duplicate detection по `info_hash` в `TorrentService`.
- Показ существующего task при повторе.

## Out of Scope

- fuzzy-совпадения по названию (название **не** primary identity).

## User Flow

```text
info_hash
  → Существует task?
      YES → показать существующий task
      NO  → создать новый task
```

## UI

При дубликате — сообщение со ссылкой на существующую загрузку.

## Domain Changes

`TorrentTask.info_hash` — уникальный ключ (на уровне бизнес-правила, позже — DB constraint).

## API / Interfaces

`TorrentService.create` выполняет проверку перед вызовом provider.

## Storage Changes

На MVP — in-memory проверка; в Phase 1 — unique constraint на `info_hash` + `user_id`.

## Security

Сравнение по `info_hash`, не по данным от пользователя.

## Error Handling

Гонка двух одновременных добавлений → одна задача, второй показывается существующая.

## Tests

- Повторный magnet → существующий task.
- Разные info_hash → разные tasks.

## Dependencies

- FEAT-006 (алгоритм входит в контракт service layer)

## Configuration

Нет.

## Implementation Tasks

- [ ] Проверка по `info_hash` в сервисе
- [ ] UI-ответ при дубликате
- [ ] Unit tests

## Acceptance Criteria

- [ ] Одинаковый info_hash → одна задача
- [ ] Название не используется как идентификатор

## Related Documentation

- `docs/ARCHITECTURE.md`
- `docs/API.md`
