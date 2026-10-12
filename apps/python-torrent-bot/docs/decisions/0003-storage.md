# ADR-0003: Конфигурационные пути хранения, без vendor-specific logic

## Status

Accepted

## Context

Целевая инфраструктура: torrent client на Proxmox, NAS (D-Link DNS-320) смонтирована через SMB/NFS в `/mnt/dns320/` с директориями `Downloads/`, `Movies/`, `TV/`. Код, привязанный к конкретной модели NAS, усложнит замену хранилища и тестирование.

## Decision

- Весь доступ к storage — только через configuration: `DOWNLOAD_PATH`, `MOVIES_PATH`, `TV_PATH`.
- Application не содержит D-Link-specific (и вообще vendor-specific) logic.
- Структура библиотеки — контракт: `Movies/<Title> (<Year>)/`, `TV/<Show>/Season NN/`.
- Перемещение в media library только после 100% завершения torrent.
- При работе с filesystem обязателна защита от `../../` и абсолютных путей из пользовательских данных.

## Consequences

Плюсы:

- Хранилище меняется правкой `.env`, не кода.
- Локальная разработка возможна на любой директории.

Минусы:

- Нужно валидировать пути в одном месте (storage abstraction), а не рассчитывать на «правильные» константы.
