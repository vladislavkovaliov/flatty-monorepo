# FEAT-019 — Lampa Library

## Status

Planned

## Goal

Использовать стабильную медиатеку (`Movies/`, `TV/`) как источник для Lampa.

> **Phase gate:** только после стабилизации media library (FEAT-013 — FEAT-015). Lampa — downstream consumer.

## User Story

As a user, I want my organized library visible in Lampa, so I can watch downloaded content.

## Scope

- Интеграция Lampa со стабильной media library.

## Out of Scope

- Прямая связь `Telegram -> Lampa` — **запрещена**.

## User Flow

```text
Torrent → Media Organizer → Media Library → Lampa
```

## UI

Нет Telegram UI; Lampa потребляет файловую библиотеку.

## Domain Changes

Медиатека как контракт (структура из FEAT-014 / FEAT-015).

## API / Interfaces

Интеграция на уровне storage/library, не через Telegram handlers.

## Storage Changes

Нет (Lampa читает существующие `Movies/` и `TV/`).

## Security

Lampa не получает доступ к секретам бота.

## Error Handling

Ошибка интеграции не должна ломать загрузочный pipeline.

## Tests

Smoke-тест структуры библиотеки как контракта для Lampa.

## Dependencies

- FEAT-014
- FEAT-015

## Configuration

Пути к библиотеке (`MOVIES_PATH`, `TV_PATH`).

## Implementation Tasks

- [ ] Стабилизация структуры библиотеки
- [ ] Lampa integration
- [ ] Smoke tests контракта

## Acceptance Criteria

- [ ] Lampa видит организованную библиотеку
- [ ] Нет прямой связи Telegram → Lampa

## Related Documentation

- `docs/ARCHITECTURE.md`
- `docs/ROADMAP.md`
