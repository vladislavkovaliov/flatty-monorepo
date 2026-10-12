# FEAT-003 — Media Type Selection

## Status

Implemented

## Goal

Позволить пользователю указать тип контента (Фильм / Сериал), чтобы будущая организация файлов знала целевую папку.

## User Story

As a user, I want to choose whether this download is a movie or a series, so files end up in the right library.

## Scope

- Шаг wizard «Куда сохранить?» после показа magnet-информации.
- Кнопки: 🎬 Фильм, 📺 Сериал, ❌ Отмена.
- Enum `MediaType`.
- Callback `torrent:category:movie` / `torrent:category:series`.

## Out of Scope

- Выбор качества (FEAT-004).
- Реальное сохранение файлов.

## User Flow

```text
Magnet info показано
  → «Куда сохранить?»
  → Фильм / Сериал / Отмена
  → Выбор качества (FEAT-004)
```

## UI

```text
🔗 Torrent найден

🎬 Название:
The.Matrix.1999.1080p.BluRay

🔑 Hash:
A1B2C3D4E5...

📡 Trackers:
3

💾 Размер:
не определён

Куда сохранить?
```

Кнопки: `🎬 Фильм`, `📺 Сериал`, `❌ Отмена`.

## Domain Changes

```python
class MediaType(str, Enum):
    MOVIE = "movie"
    SERIES = "series"
```

## API / Interfaces

Keyboards: `app/keyboards/torrent.py`. Шаг в FSM `TorrentWizard.choosing_category`.

## Storage Changes

Нет (данные — в FSM).

## Security

- Callback data короткие, только из whitelist значений.
- Не доверять callback data от пользователя — валидировать.

## Error Handling

Неизвестный callback → игнорировать/ответ ошибкой, не падать. Отмена → сброс FSM.

## Tests

- Callback категории маппится в `MediaType`.
- Отмена сбрасывает состояние.

## Dependencies

- FEAT-001
- FEAT-002

## Configuration

Нет.

## Implementation Tasks

- [x] Enum `MediaType`
- [x] Кнопки категории
- [x] FSM состояние `choosing_category`
- [x] Обработка callback и отмены

## Acceptance Criteria

- [x] После magnet показывается блок информации и вопрос «Куда сохранить?»
- [x] Кнопки Фильм/Сериал/Отмена работают
- [x] Используется enum, не произвольные строки

## Related Documentation

- `docs/ARCHITECTURE.md`
- `docs/API.md`
