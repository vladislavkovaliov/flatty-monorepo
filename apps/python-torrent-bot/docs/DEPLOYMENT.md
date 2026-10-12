# DEPLOYMENT

Local запуск, Docker, Proxmox, NAS, production deployment.

---

## Текущий статус

MVP — только локальный запуск. Docker / Proxmox / NAS deployment относятся к будущим фазам (Roadmap Phase 3, Phase 6) и реализовываться не будут, пока проект явно не переведён на следующую фазу.

---

## Локальный запуск (MVP)

1. Python 3.12+.
2. Скопировать `.env.example` в `.env`, заполнить `BOT_TOKEN`.
3. `TEST_MODE=true` — оставить по умолчанию.
4. Установить зависимости, запустить приложение.

---

## Production (будущее)

Из структуры проекта:

- `Dockerfile`
- `docker-compose.yml`

Целевая инфраструктура:

```text
Proxmox
  ├── torrent client (qBittorrent)
  └── NAS mount (SMB/NFS) → /mnt/dns320/
        ├── Downloads/
        ├── Movies/
        └── TV/
```

NAS: D-Link DNS-320. Application не должна содержать D-Link-specific logic — только переменные окружения `DOWNLOAD_PATH`, `MOVIES_PATH`, `TV_PATH`.

См. также: `docs/ROADMAP.md` (Phase 3 — NAS, Phase 6 — Production).
