# ROADMAP

Этапы развития проекта.

---

## Зона MVP

**FEAT-001 — FEAT-009** являются первоначальной зоной MVP.

FEAT-010 и выше не реализовывать, пока пользователь явно не переведёт проект на следующий этап.

---

## Phase 0 — Telegram UI (MVP)

- [x] Start menu
- [x] Magnet parser
- [x] Movie / Series
- [x] Quality
- [x] Confirmation
- [x] Fake queue
- [x] Downloads list
- [x] Download controls
- [x] Tests
- [x] Documentation

## Phase 1 — Persistent State

- [x] Database
- [x] Users
- [x] Torrents
- [ ] Settings
- [ ] Persistent FSM

## Phase 1.5 — REST API (Swagger)

- [x] FastAPI + Swagger UI (`/docs`)
- [x] `/api/tasks`: list/get/create/pause/resume/cancel/delete
- [x] `API_USER_ID` как актор, localhost без auth

## Phase 2 — qBittorrent

- [x] Provider
- [x] Add magnet
- [x] Status
- [x] Pause
- [x] Resume
- [x] Cancel
- [x] Delete
- [x] Progress
- [x] Speed
- [x] ETA

Реализовано 2026-10-13 (FEAT-011): `QBittorrentProvider` на `qbittorrentapi`,
`TEST_MODE=false` впервые валиден. Ручной E2E с живым qBittorrent — см. TESTING.md.

## Phase 3 — NAS

- [ ] NAS mount
- [ ] Downloads
- [ ] Movies
- [ ] TV
- [ ] Storage abstraction
- [ ] Disk space checks

## Phase 4 — Media Organizer

- [ ] Movie detection
- [ ] Series detection
- [ ] Season detection
- [ ] Renaming
- [ ] Folder creation
- [ ] Completion processing

## Phase 5 — Lampa

- [ ] Stable media library
- [ ] Lampa integration

## Phase 6 — Production

- [ ] Docker
- [ ] Health checks
- [ ] Monitoring
- [ ] Structured logging
- [ ] Backups
- [ ] Background worker
- [ ] CI/CD

---

## Порядок развития

```text
Telegram UI + Magnet Parser + Fake Queue
  → Persistent State
  → REST API
  → qBittorrent
  → NAS
  → Media Organizer
  → Movies / TV
  → Lampa
```

Расширения добавляются так, чтобы Telegram UI и бизнес-слой не пришлось переписывать с нуля.
