# FEAT-014 — Movie Library

## Status

Planned

## Goal

Хранить фильмы в канонической структуре `Movies/`.

> **Phase gate:** следует за FEAT-013 (Phase 4).

## User Story

As a user, I want my movies organized as `Movies/<Title> (<Year>)/<Title> (<Year>).<ext>`, so my library is browsable.

## Scope

- Целевая структура для `MediaType.MOVIE`.
- Создание директорий и размещение файлов.

## Out of Scope

- Сериалы (FEAT-015), Lampa (FEAT-019).

## User Flow

```text
Завершённый фильм
  → Media Organizer
  → Movies/<Title> (<Year>)/<Title> (<Year>).mkv
```

## UI

Нет отдельного UI на этом этапе (библиотека для Lampa).

## Domain Changes

`media_items` c типом movie.

## API / Interfaces

`MediaOrganizer` (FEAT-013).

## Storage Changes

```text
Movies/
└── The Matrix (1999)/
    └── The Matrix (1999).mkv
```

## Security

Имена из пользовательских/торрент-данных нормализуются; path traversal исключён.

## Error Handling

Конфликт имён / нет прав → ошибка в лог, файл не теряется.

## Tests

Unit: формирование пути для фильма.

## Dependencies

- FEAT-013
- FEAT-012

## Configuration

`MOVIES_PATH`.

## Implementation Tasks

- [ ] Movie path template
- [ ] Создание директорий
- [ ] Тесты

## Acceptance Criteria

- [ ] Фильм попадает в `Movies/<Title> (<Year>)/`
- [ ] Существующие файлы не перезаписываются молча

## Related Documentation

- `docs/ARCHITECTURE.md`
