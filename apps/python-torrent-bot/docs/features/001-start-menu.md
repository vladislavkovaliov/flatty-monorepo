# FEAT-001 — Start Menu

## Status

Implemented

## Goal

Дать пользователю точку входа в бота и понятные действия: отправить magnet, увидеть загрузки, настройки, помощь.

## User Story

As a user, I want to start the bot and see the main menu, so I know how to create a download.

## Scope

- Команда `/start`.
- Приветственное сообщение.
- Inline keyboard с тремя кнопками.

## Out of Scope

- Обработка magnet (FEAT-002).
- Реальная загрузка, qBittorrent, NAS.

## User Flow

```text
/start
  → Приветственное сообщение
  → Пользователь выбирает действие
```

## UI

Сообщение:

```text
🤖 Torrent Bot

Отправьте magnet-ссылку, чтобы создать новую загрузку.
```

Кнопки:

- 📥 Мои загрузки
- ⚙️ Настройки
- ℹ️ Помощь

## Domain Changes

Нет.

## API / Interfaces

Handler в `app/handlers/start.py`. Клавиатура — `app/keyboards/main.py`.

## Storage Changes

Нет.

## Security

Callback data кнопок короткие, без секретов.

## Error Handling

Базовая обработка ошибок aiogram; бот не падает на неизвестных сообщениях.

## Tests

- Интеграционный: `/start` показывает сообщение и кнопки.

## Dependencies

Нет.

## Configuration

`BOT_TOKEN`.

## Implementation Tasks

- [x] Handler `/start`
- [x] Приветственное сообщение
- [x] Main keyboard (3 кнопки)

## Acceptance Criteria

- [x] `/start` отвечает текстом из спецификации
- [x] Показаны кнопки: Мои загрузки, Настройки, Помощь
- [x] Тесты проходят, lint/type check чистые

## Related Documentation

- `docs/ARCHITECTURE.md`
- `docs/DEVELOPMENT.md`
