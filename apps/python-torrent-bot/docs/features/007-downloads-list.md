# FEAT-007 — Downloads List

## Status

Implemented

## Goal

Показать пользователю список его загрузок с состоянием и базовыми действиями.

## User Story

As a user, I want to see all my downloads with status and progress, so I can monitor them.

## Scope

- Команда `/downloads`.
- Список задач текущего пользователя.
- Отображение статуса, прогресса, скорости, ETA, размера.
- Inline actions для активных и завершённых задач.

## Out of Scope

- Реальное управление qBittorrent (FEAT-008 управляет через provider, но реальный клиент — позже).
- Пагинация/поиск (пока не требуется).

## User Flow

```text
/downloads
  → Список задач пользователя
  → Открыть / Пауза / Отменить  (активные)
  → Открыть / Удалить           (завершённые)
```

## UI

```text
📥 Мои загрузки

#001
🎬 The Matrix
1080p
⏳ 35%

#002
📺 Breaking Bad
1080p
✅ Завершено
```

Для активных: `📊 Открыть`, `⏸ Пауза`, `❌ Отменить`.
Для завершённых: `📊 Открыть`, `🗑 Удалить из списка`.

## Domain Changes

Нет (использует `TorrentTask`, `TorrentStatus`).

## API / Interfaces

Handler `app/handlers/downloads.py`, keyboards `app/keyboards/downloads.py`. Данные — через `TorrentService`, не напрямую из provider.

## Storage Changes

In-memory на MVP; контракт — только задачи текущего `user_id`.

## Security

- Показывать только загрузки текущего Telegram user (FEAT-009).
- Callback actions валидировать, проверять ownership.

## Error Handling

- Пустой список → понятное сообщение.
- Ошибка чтения состояния → сообщение об ошибке, не падать.

## Tests

- Список содержит только задачи пользователя.
- Отображаются правильные статусы и кнопки для активных/завершённых.

## Dependencies

- FEAT-006
- FEAT-009

## Configuration

Нет.

## Implementation Tasks

- [x] `/downloads` handler
- [x] Рендер списка
- [x] Кнопки действий
- [x] Тесты изоляции и отображения

## Acceptance Criteria

- [x] `/downloads` показывает только свои загрузки
- [x] Статус/прогресс отображаются по формату из спецификации
- [x] Кнопки соответствуют состоянию задачи

## Related Documentation

- `docs/ARCHITECTURE.md`
- `docs/API.md`
