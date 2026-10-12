# ARCHITECTURE

Компоненты, зависимости, boundaries, data flow, provider architecture.

---

## 1. Главный архитектурный принцип

Проект нельзя реализовывать как один большой Telegram-скрипт.

Целевая архитектура:

```text
                    Telegram
                       |
                       v
              Telegram Bot / UI
                       |
                       v
              Application Services
                       |
          +------------+------------+
          |                         |
          v                         v
       Database              Torrent Provider
                                    |
                       +------------+------------+
                       |                         |
                       v                         v
                Fake Provider             qBittorrent
                                                |
                                               v
                                           NAS Storage
                                                |
                              +-----------------+----------------+
                              |                                  |
                              v                                  v
                           Movies                               TV
                              |                                  |
                              +-----------------+----------------+
                                                |
                                                v
                                              Lampa
```

---

## 2. Ответственность компонентов

### Telegram (handlers / UI)

Отвечает только за:

- команды;
- сообщения;
- inline keyboard;
- FSM;
- отображение состояния;
- получение пользовательского ввода.

Telegram handlers **не должны** напрямую:

- обращаться к qBittorrent;
- работать с NAS;
- перемещать файлы;
- содержать бизнес-логику торрент-загрузок.

### Application Services

Отвечают за бизнес-логику:

- создание torrent task;
- проверку данных;
- duplicate detection;
- управление состоянием;
- вызов provider;
- orchestration.

### Domain

Содержит:

- модели;
- enum;
- бизнес-сущности;
- бизнес-правила.

### Providers

Изолируют внешние системы. Минимум:

- `FakeTorrentProvider`
- `QBittorrentProvider`

Telegram и application layer не должны знать внутренние детали qBittorrent.

### Storage

Отвечает за хранение файлов.

Production:

```text
Proxmox
   |
   v
NAS mount
   |
   +--> Downloads
   +--> Movies
   +--> TV
```

---

## 3. MVP Flow

```text
/start
  → Пользователь отправляет magnet
  → Magnet Parser
  → Показать информацию
  → Выбор: Фильм / Сериал
  → Выбор качества
  → Подтверждение
  → Fake Torrent Service
  → Test Queue
  → /downloads
```

---

## 4. Domain Models

Использовать enum вместо произвольных строк.

```python
class MediaType(str, Enum):
    MOVIE = "movie"
    SERIES = "series"

class Quality(str, Enum):
    UHD_2160 = "2160p"
    FHD_1080 = "1080p"
    HD_720 = "720p"
    UNKNOWN = "unknown"

class TorrentStatus(str, Enum):
    QUEUED = "queued"
    DOWNLOADING = "downloading"
    PAUSED = "paused"
    COMPLETED = "completed"
    FAILED = "failed"
    CANCELLED = "cancelled"
```

`TorrentTask` должен содержать минимум:

- `id: int`
- `user_id: int`
- `magnet: str`
- `info_hash: str`
- `name: str`
- `media_type: MediaType`
- `quality: Quality`
- `status: TorrentStatus`
- `progress: float`
- `size_bytes: int | None`
- `download_speed: int`
- `upload_speed: int`
- `eta_seconds: int | None`
- `created_at: datetime`
- `completed_at: datetime | None`

---

## 5. FSM

Использовать aiogram FSM:

```python
class TorrentWizard(StatesGroup):
    choosing_category = State()
    choosing_quality = State()
    confirming = State()
```

- Нельзя использовать глобальные Python variables для хранения состояния пользователя.
- На MVP можно использовать `MemoryStorage`.
- Persistent FSM storage — НЕ в Phase 1 (design §Approach, см. §8).

---

## 6. Duplicate Detection

Главным идентификатором torrent является **info_hash**. Не использовать название как primary identity.

```text
info_hash
  → Существует task?
      YES → показать существующий task
      NO  → создать новый task
```

---

## 7. Callback Data

Callback data должна быть короткой.

Разрешено:

```text
torrent:category:movie
torrent:category:series
torrent:quality:2160
torrent:quality:1080
torrent:quality:720
torrent:confirm:add
torrent:confirm:cancel
torrent:confirm:back
```

Не помещать в callback:

- полный magnet;
- большие JSON;
- секреты;
- большие metadata objects.

Данные должны храниться в FSM или database.

---

## 8. Database (FEAT-010, реализовано)

- PostgreSQL через psycopg 3 async (`psycopg[binary,pool]`); пул создаётся при старте
  (`app/db/pool.py`), закрывается в `finally` `app/main.py`.
- Схема: `app/db/schema.sql`, применяется идемпотентно при старте
  (`CREATE TABLE IF NOT EXISTS`). Alembic не используется — upgrade path:
  Alembic, когда схема начнёт часто меняться.
- Таблицы: `users`, `torrents` (+ `UNIQUE (user_id, info_hash)` — duplicate detection
  как constraint), `torrent_events` (lifecycle-события: created/status_changed/deleted,
  НЕ тики прогресса), `settings`. `media_items` — FEAT-013.
- Repository pattern: `TorrentRepository` (protocol, `app/repositories/__init__.py`) →
  `PostgresRepository` (default) / `InMemoryRepository` (unit tests).
  Handlers не знают о БД — всё через `TorrentService` (ADR-0001).
- `DATABASE_URL` обязателен: fail-fast при отсутствии/недоступности БД;
  пароль никогда не логируется (redaction `user:***@host`).
- История бизнес-состояния — в `torrent_events`, не в Redis.
- FSM — `MemoryStorage`: wizard живёт секунды, DB/Redis-storage не нужен
  (design §Approach).
- `provider_id` задачи — session state процесса (колонки нет, design §Components);
  upgrade path — FEAT-011.

---

## 9. Background Worker (будущее, FEAT-016)

Когда появятся реальные загрузки, нельзя постоянно опрашивать qBittorrent из Telegram handlers.

Целевая архитектура:

```text
Telegram Bot → Database → Background Worker → qBittorrent
```

Worker:

- получает torrent state;
- обновляет database;
- отправляет Telegram notifications;
- обрабатывает завершённые torrents.

Не редактировать Telegram message каждую секунду.

---

## 10. qBittorrent (будущее, FEAT-011)

```text
TorrentService → TorrentProvider → QBittorrentProvider → qBittorrent Web API
```

Telegram layer не должен зависеть от конкретной реализации qBittorrent.

---

## 11. NAS (будущее, FEAT-012)

- NAS: D-Link DNS-320.
- Torrent client работает на Proxmox.
- NAS монтируется на Proxmox через SMB или NFS.

```text
/mnt/dns320/
├── Downloads/
├── Movies/
└── TV/
```

Application работает через configuration (`DOWNLOAD_PATH`, `MOVIES_PATH`, `TV_PATH`) и **не должна содержать D-Link-specific logic**.

---

## 12. Download Flow (после появления реального qBittorrent)

```text
Magnet → Parse → Select category → Select quality → Create task
  → qBittorrent → Downloads/ → 100% complete
  → Media Organizer → Movies/ | TV/
```

Не перемещать файл в media library до завершения torrent.

---

## 13. Media Organizer (будущее, FEAT-013)

Ответственность: определить media type, movie/series, season; нормализовать название; создать директории; переместить файлы.

Movie:

```text
Movies/
└── The Matrix (1999)/
    └── The Matrix (1999).mkv
```

Series:

```text
TV/
└── Breaking Bad/
    ├── Season 01/
    │   ├── Breaking Bad - S01E01.mkv
    │   └── Breaking Bad - S01E02.mkv
    └── Season 02/
```

---

## 14. Lampa (будущее, FEAT-019)

Lampa — downstream consumer. **Не делать прямую связь** `Telegram -> Lampa`.

```text
Torrent → Media Organizer → Media Library → Lampa
```

Сначала должна стабилизироваться media library. Только после этого — Lampa integration.

---

## 15. Final Development Principle

Главная задача MVP — проверить UX и архитектурный workflow. MVP не должен быть одноразовым прототипом.

Порядок развития:

```text
Telegram UI + Magnet Parser + Fake Queue
  → Persistent State
  → qBittorrent
  → NAS
  → Media Organizer
  → Movies / TV
  → Lampa
```

Любое следующее расширение должно добавляться так, чтобы уже написанный Telegram UI и бизнес-слой не пришлось переписывать с нуля.
