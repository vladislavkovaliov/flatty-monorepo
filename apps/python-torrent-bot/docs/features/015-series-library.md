# FEAT-015 — Series Library

## Status

Planned

## Goal

Хранить сериалы в канонической структуре `TV/` с разбивкой по сезонам.

> **Phase gate:** следует за FEAT-013 (Phase 4).

## User Story

As a user, I want episodes organized as `TV/<Show>/Season NN/<Show> - SnnEnn.ext`, so my library is browsable.

## Scope

- Целевая структура для `MediaType.SERIES`.
- Определение show и season, создание директорий.

## Out of Scope

- Фильмы (FEAT-014), Lampa (FEAT-019).

## User Flow

```text
Завершённый сериал
  → Media Organizer (show + season detection)
  → TV/<Show>/Season NN/<Show> - SnnEnn.mkv
```

## UI

Нет отдельного UI на этом этапе.

## Domain Changes

`media_items` c типом series + season/episode.

## API / Interfaces

`MediaOrganizer` (FEAT-013).

## Storage Changes

```text
TV/
└── Breaking Bad/
    ├── Season 01/
    │   ├── Breaking Bad - S01E01.mkv
    │   └── Breaking Bad - S01E02.mkv
    └── Season 02/
```

## Security

Нормализация имён; path traversal исключён.

## Error Handling

Не удалось определить сезон → ошибка в лог, файл не теряется.

## Tests

Unit: detection сезона/эпизода на фикстурах.

## Dependencies

- FEAT-013
- FEAT-012

## Configuration

`TV_PATH`.

## Implementation Tasks

- [ ] Show/season detection
- [ ] Episode renaming
- [ ] Тесты

## Acceptance Criteria

- [ ] Эпизоды попадают в `TV/<Show>/Season NN/`
- [ ] Имена в формате `SnnEnn`

## Related Documentation

- `docs/ARCHITECTURE.md`
