# ROADMAP

Этапы развития проекта.

---

## Зона MVP

**FEAT-001 — FEAT-009** являются первоначальной зоной MVP.

FEAT-010 и выше не реализовывать, пока пользователь явно не переведёт проект на следующий этап.

---

## Phase 0 — Telegram UI (MVP)

- [ ] Start menu
- [ ] Magnet parser
- [ ] Movie / Series
- [ ] Quality
- [ ] Confirmation
- [ ] Fake queue
- [ ] Downloads list
- [ ] Download controls
- [ ] Tests
- [ ] Documentation

## Phase 1 — Persistent State

- [ ] Database
- [ ] Users
- [ ] Torrents
- [ ] Settings
- [ ] Persistent FSM

## Phase 1.5 — REST API (Swagger)

- [x] FastAPI + Swagger UI (`/docs`)
- [x] `/api/tasks`: list/get/create/pause/resume/cancel/delete
- [x] `API_USER_ID` как актор, localhost без auth

## Phase 2 — qBittorrent

- [ ] Provider
- [ ] Add magnet
- [ ] Status
- [ ] Pause
- [ ] Resume
- [ ] Cancel
- [ ] Delete
- [ ] Progress
- [ ] Speed
- [ ] ETA

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
