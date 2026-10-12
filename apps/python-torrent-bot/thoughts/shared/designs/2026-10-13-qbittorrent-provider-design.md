# Design: QBittorrentProvider (FEAT-011, Phase 2)

date: 2026-10-13
topic: "QBittorrentProvider — реальный торрент-провайдер"
status: validated

## Problem Statement

MVP работает только в TEST_MODE с FakeTorrentProvider: загрузки имитируются, прогресс фейковый (+12.5% за рендер), рестарт процесса теряет связь задач с провайдером. Нужен реальный провайдер поверх qBittorrent Web API, чтобы через Telegram UI запускались настоящие загрузки с реальными прогрессом/скоростью/ETA и управлением (пауза/продолжить/отмена/удалить).

Целевая инфраструктура: qBittorrent (Docker, linuxserver) в LXC на Proxmox, скачивает на NAS через NFS-монтирование. Боту доступ к NAS не нужен — только HTTP к Web API.

## Constraints

- ADR-0002: handlers и сервис работают только через Protocol `TorrentProvider`; qBittorrent — деталь провайдера
- Процессуальный гейт ROADMAP снят: пользователь явно перевёл проект на Phase 2 («давай через провайдер»)
- `TEST_MODE=true` остаётся дефолтом и полностью рабочим (Fake); `false` впервые становится валидным
- Пароль qBittorrent никогда не логируется; Web API — только LAN
- Интеграция с живым qBittorrent — вне CI (ручной сценарий в docs/TESTING.md)
- mypy strict, ruff, pytest — остаются зелёными

## Approach

Реализуем `QBittorrentProvider` на библиотеке `qbittorrentapi` (зрелый клиент Web API: авто-relogin, enum состояний, поддержка v4.1+). Библиотека синхронная (requests) — вызовы оборачиваем в `asyncio.to_thread`.

Рассматривали голый `httpx` + Web API v2 — отклонили: ручной cookie-login, re-login, разбор 20+ состояний и версий API = больше кода без выигрыша.

FTP/SMB-вариант «положить .torrent в папку на NAS» отклонён: watch-dir не поддерживает magnet, не даёт обратной связи по прогрессу и убивает уже готовые FEAT-007/008.

## Architecture

```
Telegram handler → TorrentService → TorrentProvider (Protocol)
                                        ├─ FakeTorrentProvider   (TEST_MODE=true)
                                        └─ QBittorrentProvider   (TEST_MODE=false)
                                                └─ qbittorrentapi.Client → Web API qBittorrent
                                                                            └─ save_path → NFS → NAS
```

Выбор провайдера — в `app/main.py` по `config.test_mode`. Хендлеры не меняются.

## Components

### QBittorrentProvider (новый: `app/providers/qbittorrent.py`)

Реализует Protocol:

- `add_magnet(magnet, save_path)` → `torrents_add(urls=magnet, save_path=...)`; возвращает infohash (`"Ok."` → найти хеш по magnet)
- `get_info(torrent_id)` → `torrents_info(torrent_hashes=...)`; маппит в `TorrentInfo`
- `pause/resume` → `torrents_pause/stop` + `torrents_start` (алиасы resume в v5)
- `delete(torrent_id, delete_files)` → `torrents_delete(delete_files=...)`

### Маппинг состояний qBittorrent → TorrentStatus

- `metaDL`, `downloading`, `forcedDL`, `stalledDL`, `checkingDL`, `moving`, `allocating` → `DOWNLOADING`
- `queuedDL` → `QUEUED`
- `pausedDL`, `stoppedDL` → `PAUSED`
- `pausedUP`, `stoppedUP`, `uploading`, `stalledUP`, `forcedUP`, `checkingUP` → `COMPLETED` (progress ≥ 1.0 дополнительно гарантирует COMPLETED)
- `error`, `missingFiles` → `FAILED`
- неизвестные → `FAILED` (логировать)

### Расширение контракта: `get_status` → `get_info`

Protocol вместо `get_status -> TorrentStatus` возвращает `get_info(torrent_id) -> TorrentInfo`:

- `status: TorrentStatus`
- `progress: float` (0–100)
- `download_speed: int` (B/s)
- `upload_speed: int` (B/s)
- `eta_seconds: int | None`

Новая модель в `app/models/torrent.py`. FakeTorrentProvider адаптируется (существующая тик-симуляция + корректные поля). `TorrentService._sync` пишет реальные значения в уже существующие колонки `progress/download_speed/upload_speed/eta_seconds`; фейковый `+12.5%` удаляется.

### Cancel = delete без файлов

`cancel` в Protocol не добавляем (YAGNI). `TorrentService.cancel` вызывает `provider.delete(provider_id, delete_files=False)`, затем ставит `CANCELLED`. Иначе реальный торрент продолжил бы качаться в qBittorrent.

### Персистентный provider_id

- Миграция схемы: `ALTER TABLE torrents ADD COLUMN provider_id TEXT` (schema.sql идемпотентный — колонка добавляется при старте; для уже созданных БД — ALTER в schema.sql через IF NOT EXISTS-паттерн, согласно существующей схеме миграций приложения)
- Убираем session-only `dict _provider_ids`; `_ensure_provider` пишет `provider_id` в задачу/БД
- Задача с `provider_id=NULL` и не-терминальным статусом после рестарта: `_ensure_provider` добавляет magnet заново (дедупликация по info_hash на стороне qBittorrent)
- Задача с `provider_id`, но отсутствующая в qBittorrent → `get_info` исключение → `FAILED` (текущая семантика)

### Config (`app/config.py`)

- Удаляем RuntimeError «Phase 0 requires TEST_MODE=true»
- Новые поля: `qbittorrent_url`, `qbittorrent_username`, `qbittorrent_password`
- Валидация: при `test_mode=false` — `qbittorrent_url` обязателен (username/password опциональны, qBittorrent может быть без auth в закрытой LAN); при `test_mode=true` — не требуются
- `__repr__`/логи: password/redaction

### Save paths

`SAVE_PATHS` (media_type → путь) переезжают в конфиг с env: `SAVE_PATH_MOVIES`, `SAVE_PATH_TV` (дефолты — текущие значения). Пути трактуются как видит их **qBittorrent**.

### Wiring (`app/main.py`)

```text
provider = FakeTorrentProvider() if config.test_mode else QBittorrentProvider(...)
```

### Зависимости

`qbittorrentapi` — в runtime dependencies (`requirements.txt` + `pyproject.toml`, синхронно). `httpx` остаётся dev-only.

### Docs

- `.env.example`, `docs/CONFIGURATION.md` (новые переменные, TEST_MODE=false впервые валиден)
- `docs/ROADMAP.md` — Phase 2 по мере выполнения
- `docs/TESTING.md` — ручной сценарий с реальным qBittorrent
- `docs/features/011-*.md` — статус/чекбоксы
- `AGENTS.md`/`README.md` — снять «TEST_MODE=true обязателен»
- `docs/API.md` §2 — обновить контракт (get_info)

## Data Flow

1. `/start` → пользователь шлёт magnet → wizard (без изменений)
2. `TorrentService.create` → `add_magnet(magnet, save_path)` → infohash сохраняется как `provider_id` в БД
3. Список/детали загрузок → `_sync` → `get_info(provider_id)` → статус/прогресс/скорости в БД и в ответ UI
4. Pause/Resume/Cancel/Delete → сервис → провайдер → qBittorrent; исключения провайдера → `FAILED` (delete → False → 404)

## Error Handling

- Любое исключение провайдера — как сейчас: log + `FAILED` (delete → `False`), хендлер не падает
- Недоступен Web API → задачи уходят в `FAILED` при следующем `_sync` (тонкий момент: реальный отказ API отличаем по исключению connection — логируем с URL без пароля)
- `add_magnet` вернул не `"Ok."` → исключение провайдера → задача `FAILED`
- qBittorrent v5 переименовал pause→stop: библиотека alias'ит; на маппинге состояний учитываем `stopped*`

## Testing Strategy

- `tests/unit/test_qbittorrent_mapping.py` — таблица состояний qBittorrent → TorrentStatus (+ progress≥100 → COMPLETED)
- `tests/unit/test_qbittorrent_provider.py` — мок `qbittorrentapi.Client`: add/pause/resume/delete/get_info, обработка ошибок, возврат хеша
- `tests/unit/test_fake_provider.py` / `test_torrent_service.py` — адаптация под `get_info`
- `tests/unit/test_config.py` — гейт: `test_mode=false` без URL → RuntimeError; с URL → ок; password не в repr
- `tests/integration/*` — остаются на Fake (зелёные); реальный qBittorrent — ручной прогон по TESTING.md
- Приёмка FEAT-011: реальная загрузка через UI, реальные прогресс/скорость/ETA, handlers не знают о qBittorrent

## Open Questions

Закрыты решениями:

- cancel ≠ отдельный метод провайдера → delete(delete_files=False)
- get_status → get_info (одна точка расширения вместо второго метода)
- провайдер на qbittorrentapi, не httpx

Остаётся за рамками (не блокирует FEAT-011): фоновый воркер опроса (FEAT-016), категории/лимиты qBittorrent, API key вместо пароля (qBittorrent 5.1+) — пароль достаточно.
