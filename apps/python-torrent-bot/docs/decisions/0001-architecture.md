# ADR-0001: Слоистая архитектура вместо монолитного Telegram-скрипта

## Status

Accepted

## Context

Проект можно реализовать двумя способами: как один большой Telegram-скрипт (handlers делают всё: парсинг, qBittorrent, файлы) или как слоистую систему. Монолит быстрее стартует, но делает невозможными замену torrent-клиента, добавление database и NAS без полного переписывания UI.

## Decision

Реализовать целевую архитектуру из четырёх слоёв:

- **Telegram (handlers/UI)** — только команды, сообщения, inline keyboard, FSM, отображение состояния, ввод. Запрещено: обращения к qBittorrent, работа с NAS, перемещение файлов, бизнес-логика торрентов.
- **Application Services** — бизнес-логика: создание torrent task, валидация, duplicate detection, управление состоянием, вызов provider, orchestration.
- **Domain** — модели, enum, бизнес-сущности, бизнес-правила.
- **Providers** — изоляция внешних систем (`FakeTorrentProvider`, `QBittorrentProvider`).

## Consequences

Плюсы:

- Telegram UI и бизнес-слой не переписываются при следующих расширениях (qBittorrent, NAS, Lampa).
- Тестируемость: services/providers тестятся без Telegram и внешней инфраструктуры.
- Единый development contract в `docs/DEVELOPMENT.md`.

Минусы:

- Больше файлов и слоёв на старте, чем в «одном скрипте».
