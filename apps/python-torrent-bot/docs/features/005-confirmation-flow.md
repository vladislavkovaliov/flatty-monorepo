# FEAT-005 — Confirmation Flow

## Status

Implemented

## Goal

Показать сводку будущей загрузки и дать пользователю явно подтвердить, изменить или отменить.

## User Story

As a user, I want to review category, quality and destination path before the download is queued, so mistakes are caught early.

## Scope

- Шаг «Готово к добавлению» со сводкой.
- Кнопки: `⬇️ Добавить в очередь`, `✏️ Изменить`, `❌ Отмена`.
- Callbacks: `torrent:confirm:add`, `torrent:confirm:cancel`, `torrent:confirm:back`.
- Уведомление о TEST MODE.

## Out of Scope

- Сама постановка в очередь (FEAT-006).

## User Flow

```text
Качество выбрано
  → Сводка
  → Добавить / Изменить / Отмена
  → Fake Torrent Service (FEAT-006)
```

## UI

```text
✅ Готово к добавлению

🎬 Название:
The Matrix

📁 Категория:
Фильм

🎞 Качество:
1080p

💾 Путь:
Movies/

🧪 TEST MODE

Файл реально скачиваться НЕ будет.
```

Кнопки: `⬇️ Добавить в очередь`, `✏️ Изменить`, `❌ Отмена`.

Путь зависит от категории: Фильм → `Movies/`, Сериал → `TV/`.

## Domain Changes

Нет (использует `MediaType`, `Quality`).

## API / Interfaces

Keyboards: `app/keyboards/torrent.py`. FSM: `TorrentWizard.confirming`.

## Storage Changes

Нет (данные — в FSM).

## Security

- Не подставлять пользовательские строки в пути без валидации.
- Callback только из whitelist.

## Error Handling

- Отмена → сброс FSM, сообщение об отмене.
- «Изменить» → возврат на предыдущий шаг (`torrent:confirm:back`).

## Tests

- `confirm:add` продолжает flow.
- `confirm:cancel` сбрасывает состояние.
- `confirm:back` возвращает на предыдущий шаг.

## Dependencies

- FEAT-004
- FEAT-006 (следующий шаг)

## Configuration

`TEST_MODE` — сообщение о тестовом режиме.

## Implementation Tasks

- [x] Сводка (название, категория, качество, путь)
- [x] Кнопки add / cancel / back
- [x] Уведомление TEST MODE
- [x] Обработка callbacks

## Acceptance Criteria

- [x] Сводка содержит все 4 поля + TEST MODE notice
- [x] Все три кнопки работают
- [x] Путь соответствует категории

## Related Documentation

- `docs/CONFIGURATION.md`
- `docs/API.md`
