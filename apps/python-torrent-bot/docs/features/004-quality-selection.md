# FEAT-004 — Quality Selection

## Status

Implemented

## Goal

Спросить желаемое качество контента до подтверждения загрузки.

## User Story

As a user, I want to pick a quality (or skip detection), so my download task records my preference.

## Scope

- Шаг «Выберите качество».
- Кнопки: 2160p / 4K, 1080p, 720p, Не определять.
- Enum `Quality`.
- Callback `torrent:quality:2160|1080|720`.

## Out of Scope

- Фильтрация файлов по качеству, парсинг качества из имени.

## User Flow

```text
Категория выбрана
  → Выберите качество
  → 2160p / 1080p / 720p / Не определять
  → Confirmation (FEAT-005)
```

## UI

```text
Выберите качество:
```

Кнопки: `2160p / 4K`, `1080p`, `720p`, `Не определять`.

## Domain Changes

```python
class Quality(str, Enum):
    UHD_2160 = "2160p"
    FHD_1080 = "1080p"
    HD_720 = "720p"
    UNKNOWN = "unknown"
```

«Не определять» → `Quality.UNKNOWN`.

## API / Interfaces

Keyboards: `app/keyboards/torrent.py`. FSM: `TorrentWizard.choosing_quality`.

## Storage Changes

Нет (данные — в FSM).

## Security

Валидировать callback; допустимы только значения из whitelist.

## Error Handling

Неизвестный callback → игнорировать. Отмена → сброс FSM.

## Tests

- Каждая кнопка маппится в правильный enum value.
- «Не определять» → `UNKNOWN`.

## Dependencies

- FEAT-003

## Configuration

Нет.

## Implementation Tasks

- [x] Enum `Quality`
- [x] Кнопки качества
- [x] FSM состояние `choosing_quality`
- [x] Обработка callback

## Acceptance Criteria

- [x] После выбора категории показывается выбор качества
- [x] 4 кнопки качества работают
- [x] Значения соответствуют enum

## Related Documentation

- `docs/ARCHITECTURE.md`
- `docs/API.md`
