# FEAT-013 — Media Organizer

## Status

Planned

## Goal

После завершения загрузки автоматически организовывать файлы в медиатеку.

> **Phase gate:** не начинать, пока пользователь явно не переведёт проект на Phase 4.

## User Story

As a user, I want finished downloads renamed and moved into Movies/ or TV/ automatically, so my library stays clean.

## Scope

- Service `MediaOrganizer`:
  - определить media type;
  - определить movie/series;
  - определить season;
  - нормализовать название;
  - создать директории;
  - переместить файлы.

## Out of Scope

- Lampa (FEAT-019).

## User Flow

```text
100% complete → Media Organizer → Movies/ | TV/
```

**Не перемещать файл в media library до завершения torrent.**

## UI

Уведомление о завершении/результате организации (связь с FEAT-017).

## Domain Changes

Сущность `media_items` (FEAT-010).

## API / Interfaces

Вызывается background worker'ом (FEAT-016), не из Telegram handlers.

## Storage Changes

Создание директорий и перемещение файлов на NAS.

## Security

- Нормализация имён: защита от `../../`, спецсимволов.
- Не доверять имени файла из торрент-метаданных.

## Error Handling

Не удалось определить/переместить → оставить в `Downloads/`, зафиксировать ошибку.

## Tests

Unit: определение типа, сезона, нормализация имён (на фикстурах).

## Dependencies

- FEAT-012
- FEAT-016

## Configuration

`MOVIES_PATH`, `TV_PATH`.

## Implementation Tasks

- [ ] Detection (movie/series/season)
- [ ] Renaming + folder creation
- [ ] Move logic + completion processing
- [ ] Тесты на фикстурах

## Acceptance Criteria

- [ ] Файлы перемещаются только после 100%
- [ ] Структура директорий соответствует примерам
- [ ] Path security соблюдена

## Related Documentation

- `docs/ARCHITECTURE.md`
- `docs/features/014-movie-library.md`
- `docs/features/015-series-library.md`
