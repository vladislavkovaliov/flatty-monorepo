# FEAT-012 — NAS Storage

## Status

Planned

## Goal

Хранить загрузки и медиатеку на NAS (D-Link DNS-320) через mount на Proxmox, с абстракцией путей.

> **Phase gate:** не начинать, пока пользователь явно не переведёт проект на Phase 3.

## User Story

As a user, I want finished files stored on my NAS, so they persist independently of the bot host.

## Scope

- NAS mount (SMB/NFS) на Proxmox: `/mnt/dns320/`.
- Директории: `Downloads/`, `Movies/`, `TV/`.
- Storage abstraction за конфигурацией.
- Disk space checks.

## Out of Scope

- Организация/переименование медиа (FEAT-013).

## User Flow

```text
qBittorrent → /mnt/dns320/Downloads → (FEAT-013) Movies/ | TV/
```

## UI

Ошибки нехватки места — сообщением пользователю.

## Domain Changes

Нет.

## API / Interfaces

Работа только через `DOWNLOAD_PATH` / `MOVIES_PATH` / `TV_PATH`.

## Storage Changes

- Application **не должна содержать D-Link-specific logic**.
- Path validation: защита от `../../` и абсолютных путей из пользовательских данных.

## Security

- `NAS_PASSWORD` не логировать, не коммитить.

## Error Handling

Mount недоступен / мало места → понятная ошибка, без падения.

## Tests

Unit: path abstraction, disk space checks (mocked).

## Dependencies

- FEAT-011

## Configuration

`DOWNLOAD_PATH=/mnt/dns320/Downloads`, `MOVIES_PATH=/mnt/dns320/Movies`, `TV_PATH=/mnt/dns320/TV`.

## Implementation Tasks

- [ ] Storage abstraction
- [ ] Disk space checks
- [ ] Path security validation
- [ ] Тесты

## Acceptance Criteria

- [ ] Пути настраиваются только через env
- [ ] Нет vendor-specific logic в коде
- [ ] Path traversal исключён

## Related Documentation

- `docs/CONFIGURATION.md`
- `docs/DEPLOYMENT.md`
