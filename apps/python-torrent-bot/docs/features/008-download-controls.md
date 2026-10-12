# FEAT-008 — Download Controls

## Status

Implemented

## Goal

Дать пользователю управление загрузками: pause, resume, cancel, delete — через service layer и provider abstraction.

## User Story

As a user, I want to pause, resume, cancel or delete my downloads, so I stay in control of the queue.

## Scope

- Действия: `pause`, `resume`, `cancel`, `delete`.
- Обработка callback из списка загрузок.
- Обновление статуса задачи после операции.

## Out of Scope

- Реальный qBittorrent (операции идут через `TorrentProvider`, на MVP — fake).

## User Flow

```text
/downloads
  → Выбор задачи
  → Пауза / Продолжить / Отменить / Удалить
  → Проверка ownership
  → TorrentService → Provider
  → Обновление статуса в UI
```

## UI

Кнопки из FEAT-007: `⏸ Пауза`, `❌ Отменить`, `🗑 Удалить из списка`, `📊 Открыть`.

## Domain Changes

Переходы `TorrentStatus`: `downloading ↔ paused`, `→ cancelled`, удаление из списка.

## API / Interfaces

`TorrentService` операции поверх `TorrentProvider.pause/resume/delete`. Handlers не вызывают provider напрямую.

## Storage Changes

Нет (in-memory на MVP).

## Security

- Перед каждой операцией проверять **ownership** (FEAT-009).
- Валидировать callback data.

## Error Handling

- Ошибка provider → сообщение пользователю, статус не менять ложно.
- Попытка операции над чужой/несуществующей задачей → отказ.

## Tests

- Pause/resume/cancel/delete меняют статус корректно.
- Операция над чужой задачей отклоняется.

## Dependencies

- FEAT-007
- FEAT-006

## Configuration

`TEST_MODE`.

## Implementation Tasks

- [x] Callback handlers управления
- [x] Сервисные операции
- [x] Ownership checks
- [x] Тесты переходов статусов

## Acceptance Criteria

- [x] Все 4 операции работают в UI
- [x] Статус обновляется после операции
- [x] Ownership проверяется перед операцией

## Related Documentation

- `docs/API.md`
- `docs/features/009-user-isolation.md`
