# FEAT-002 — Magnet Parser

## Status

Implemented

## Goal

Надёжно извлекать метаданные из magnet-ссылки без потери информации и без падений на нестандартных входах.

## User Story

As a user, I want to paste a magnet link and see its name, hash, trackers and size, so I know what I'm downloading.

## Scope

- Сервис `app/services/magnet_parser.py`.
- Модель `MagnetInfo`.
- Извлечение `xt` / `btih` / `dn` / `tr` / `xl` и остальных query parameters.
- Валидация схемы и info hash.

## Out of Scope

- Запрос размера у трекеров/parsers (только `xl`, если есть).
- Создание torrent task.

## User Flow

```text
Пользователь отправляет magnet
  → Parser
  → MagnetInfo
  → Показать информацию (FEAT-003 UI)
```

## UI

Владелец UI — следующие feature; parser только возвращает данные.

## Domain Changes

```python
@dataclass
class MagnetInfo:
    raw: str
    info_hash: str
    display_name: str | None
    trackers: list[str]
    size_bytes: int | None
```

## API / Interfaces

`parse_magnet(raw: str) -> MagnetInfo` (или эквивалент) — чистая функция/сервис, без Telegram-кода.

## Storage Changes

Нет.

## Security

- Не доверять `dn` (название torrent) — выводить как есть, без интерпретации.
- Не выполнять и не декодировать что-либо, кроме URL decoding.

## Error Handling

- Не `magnet:?` scheme → ошибка валидации.
- Нет info hash → ошибка валидации.
- Отсутствие `dn`, `tr`, `xl` → валидный результат с `None` / пустым списком.
- Неизвестные параметры → не падать, сохранять информацию.

## Tests

Обязательные unit tests:

- [ ] valid magnet
- [ ] missing dn
- [ ] multiple trackers
- [ ] URL encoded dn
- [ ] missing info hash
- [ ] invalid scheme
- [ ] xl parameter

## Dependencies

Нет.

## Configuration

Нет.

## Implementation Tasks

- [x] Модель `MagnetInfo`
- [x] Парсинг query parameters + URL decode
- [x] Валидация scheme и info hash
- [x] Unit tests

## Acceptance Criteria

- [x] Все обязательные тесты parser проходят
- [x] Парсер не падает на неизвестных параметрах
- [x] Исходный magnet сохраняется в `raw`
- [x] `ruff`, `mypy`, `pytest` чистые

## Related Documentation

- `docs/API.md`
- `docs/TESTING.md`
