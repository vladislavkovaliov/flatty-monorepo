# ADR-0002: Provider-абстракция, Fake-провайдер на MVP

## Status

Accepted

## Context

Нужно проверить UX и workflow на MVP, но реальный торрент-клиент (qBittorrent) подключать рано: запрещены реальные downloads, доступ к NAS и запись файлов (`TEST_MODE=true`). При этом архитектура не должна блокировать подключение qBittorrent позже.

## Decision

- Ввести интерфейс `TorrentProvider` (Protocol): `add_magnet`, `get_status`, `pause`, `resume`, `delete`.
- На MVP реализовать `FakeTorrentProvider` (`app/providers/fake.py`): без HTTP, без qBittorrent, без NAS, без реальных downloads; симулирует очередь и прогресс.
- В будущем — `QBittorrentProvider` под тем же интерфейсом.
- Telegram handlers и application layer никогда не обращаются к qBittorrent напрямую.

## Consequences

Плюсы:

- MVP полностью работает без внешней инфраструктуры; тесты не требуют torrent client.
- Подключение qBittorrent = новая реализация интерфейса, UI не меняется.

Минусы:

- Симуляция может не покрыть некоторые реальные кейсы (лимиты, ошибки API) — их придётся добавлять при FEAT-011.
