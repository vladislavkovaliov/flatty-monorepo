# QBittorrentProvider (FEAT-011, Phase 2) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Реальный торрент-провайдер поверх qBittorrent Web API: `TEST_MODE=false` впервые валиден, настоящие загрузки/прогресс/скорости/ETA через тот же Telegram UI; `provider_id` персистится в БД и переживает рестарт.

**Architecture:** Новый `app/providers/qbittorrent.py` реализует тот же Protocol `TorrentProvider`; выбор провайдера в `app/main.py` по `config.test_mode` (Fake ↔ QBittorrent). Контракт расширяется: `get_status → get_info -> TorrentInfo` (статус + прогресс 0–100 + скорости + ETA). Схема: колонка `torrents.provider_id`; session-only `dict _provider_ids` удаляется, при рестарте magnet перерегистрируется, если `provider_id` пуст. Отмена = `delete(delete_files=False)` + `CANCELLED`. `SAVE_PATHS` переезжают в конфиг (`SAVE_PATH_MOVIES`/`SAVE_PATH_TV`).

**Tech Stack:** Python 3.12+, `qbittorrentapi` (синхронный, вызовы через `asyncio.to_thread`), aiogram 3, FastAPI, PostgreSQL (psycopg 3), pytest/pytest-asyncio (`asyncio_mode=auto`), ruff, mypy (strict, `files = ["app"]`).

**Spec:**
- `thoughts/shared/designs/2026-10-13-qbittorrent-provider-design.md` (authoritative design, status: validated)
- `docs/features/011-qbittorrent-provider.md` (чекбоксы приёмки)
- `docs/DEVELOPMENT.md` §7 (Conventional Commits), §9 (Definition of Done)

## Global Constraints

- DoD: `ruff check .`, `ruff format --check .`, `mypy app`, `pytest` — все четыре зелёные (финальная проверка, Task 10).
- Существующие 49 тестов не удаляются (DEVELOPMENT §3.8); ломающиеся адаптируются (Task 4).
- Handlers не меняются, кроме `app/handlers/magnet.py` (забирает путь у сервиса вместо `SAVE_PATHS` из media.py — Task 6).
- Пароль qBittorrent никогда не логируется и не попадает в `repr(Config)`.
- `TEST_MODE=true` остаётся дефолтом и полностью рабочим (Fake); интеграционные тесты остаются на Fake; живой qBittorrent — только ручной сценарий (docs/TESTING.md).
- ruff: line-length 100, rules E,F,I,UP,B,SIM; mypy strict; код в `.md` не форматируется.
- Git: Conventional Commits; `.env`/секреты никогда не коммитить.
- Новые зависимости синхронизировать в обоих файлах: `pyproject.toml` `[project].dependencies` и `requirements.txt`.
- Не добавлять: cancel-метод в Protocol, фоновый воркер, категории/лимиты qBittorrent, API keys (design §Open Questions — вне FEAT-011).

## Task Dependency / Parallel Map

```text
Wave A (параллельно): Task 1 (deps qbittorrentapi) | Task 2 (config: гейт/qbittorrent/save paths + config-тесты)
Wave B (параллельно): Task 3 (контракт get_info: models + base + fake) | Task 5 (schema + repos: provider_id)
Wave C (параллельно): Task 4 (починка тестов под get_info, ждёт 3) | Task 7 (QBittorrentProvider + unit-тесты, ждёт 1+3)
Wave D:               Task 6 (torrent_service: persistence/cancel/get_info/save_paths, ждёт 2+3+5)
Wave E:               Task 8 (main.py wiring, ждёт 2+7)
Wave F:               Task 9 (docs, ждёт всё)
Final:                Task 10 (полная верификация, ждёт всё)
```

## Technical Findings (из чтения кода — обязательны к пониманию)

1. **Session-only provider_id.** `TorrentService._provider_ids: dict[int, str]` (`app/services/torrent_service.py:40`) + комментарий `ponytail:` про FEAT-011 (`:37-39`). `_ensure_provider` (`:177-183`) держит карту и пишет `task.provider_id` только в памяти. Удаляем словарь; персист — через `repository.update_task(task)` (`app/repositories/postgres.py:130`, `memory.py:41` — сигнатура не меняется).
2. **`_task_from_row` хардкодит `provider_id=""`** (`postgres.py`, комментарий про design §Components) — читать колонку после Task 5. `TorrentTask.provider_id: str = ""` уже есть (`app/models/torrent.py`), `ponytail:`-комментарий про колонку — удалить.
3. **Фейковый прогресс в сервисе.** `_PROGRESS_STEP = 12.5` (`torrent_service.py:15`) и `task.progress = min(99.0, ...)` в `_sync` (`:221`) — удаляется; прогресс теперь приходит из `TorrentInfo` (Fake сам симулирует: см. Task 3).
4. **Cancel не трогает провайдер** (`torrent_service.py:143-151`): только статус. Design: `cancel` → `provider.delete(provider_id, delete_files=False)` (если `provider_id` непуст) → `CANCELLED`. `delete` (`:158`) уже вызывает `provider.delete` — не трогаем.
5. **`SAVE_PATHS`** — `app/models/media.py:18` (`MOVIE→"Movies/"`, `SERIES→"TV/"`); используются в `torrent_service.py:181` и `app/handlers/magnet.py:14,62`. Переезжают в Config (`save_path_movies`/`save_path_tv`, env `SAVE_PATH_MOVIES`/`SAVE_PATH_TV`); сервис принимает `save_paths` dict, магнит-хендлер берёт путь через новый `TorrentService.save_path_for(media_type)`.
6. **Config-гейт.** `app/config.py:26-30` — RuntimeError при `test_mode=false`; удаляется. `Config` — frozen dataclass без `__repr__` → пароль виден в repr; добавляем маскированный `__repr__`. Существующие config-тесты (`tests/unit/test_config.py`) используют `TEST_MODE=true` — остаются зелёными, добавляются новые кейсы гейта.
7. **Битые после rename тесты:** `tests/unit/test_fake_provider.py` (3 вызова `get_status`), `tests/unit/test_torrent_service.py` (моки `get_status`, ожидания `+12.5%`). Остальные unit/integration через Fake пройдут, если Fake корректно адаптирован.
8. **mypy + qbittorrentapi.** Проверить наличие `py.typed` в wheel; если нет — override в `[tool.mypy]` `ignore_missing_imports` для `qbittorrentapi` (fastapi/uvicorn/py.typed прецедент — см. план 2026-10-04, TF §6).
9. **v5 rename pause→stop.** Библиотека alias'ит `torrents_stop`/`torrents_start`; на маппинге состояний учитываем `stoppedDL`/`stoppedUP` (design §Error Handling).

---

### Task 1: зависимость qbittorrentapi

**Files:**
- Modify: `pyproject.toml` (`[project].dependencies`)
- Modify: `requirements.txt`

**Interfaces:**
- Consumes: —
- Produces: установленный `qbittorrentapi` (импортируется в Task 7).

- [ ] **Step 1: Добавить runtime-зависимость**

`pyproject.toml`:
```toml
dependencies = [
    "aiogram>=3.25,<4",
    "python-dotenv>=1.0",
    "psycopg[binary,pool]>=3.3,<4",
    "fastapi>=0.115,<1",
    "uvicorn>=0.54,<1",
    "qbittorrentapi>=8,<10",
]
```

`requirements.txt` — синхронно в секции runtime (комментарии «синхронизировать» уже есть).

- [ ] **Step 2: Установить и проверить импорт + py.typed**

```bash
pip install -r requirements.txt
python -c "import qbittorrentapi; print(qbittorrentapi.__version__)"
python -c "import qbittorrentapi, pathlib; print(pathlib.Path(qbittorrentapi.__file__).parent / 'py.typed')"
```

Если `py.typed` отсутствует — записать TODO в Task 10 (mypy override).

- [ ] **Step 3: Прогнать существующие проверки**

```bash
ruff check . && ruff format --check . && mypy app && pytest -q
```

Ожидаемо: 49 passed, линтеры/типы чистые.

- [ ] **Step 4: Commit**

```bash
git add pyproject.toml requirements.txt
git commit -m "build: add qbittorrentapi dependency"
```

**Acceptance criteria:**
- [ ] `import qbittorrentapi` работает; 49 тестов зелёные.

---

### Task 2: config — снять гейт, qbittorrent_*, save paths

**Files:**
- Modify: `app/config.py`
- Modify: `tests/unit/test_config.py`

**Interfaces:**
- Consumes: —
- Produces: `Config.qbittorrent_url/username/password`, `Config.save_path_movies/save_path_tv` + `Config.save_paths` (dict property); `TEST_MODE=false` валиден при заданном URL. Потребители: Task 6 (save_paths), Task 7/8 (qbittorrent_*).

- [ ] **Step 1: Расширить `Config`**

Новые поля (frozen dataclass):
```python
qbittorrent_url: str = ""
qbittorrent_username: str = ""
qbittorrent_password: str = ""
save_path_movies: str = "Movies/"
save_path_tv: str = "TV/"
```

Property (мобильность `MediaType` → путь для Task 6):
```python
@property
def save_paths(self) -> dict[MediaType, str]:
    return {MediaType.MOVIE: self.save_path_movies, MediaType.SERIES: self.save_path_tv}
```

`__repr__` вручную: все поля, кроме `qbittorrent_password` → `"***"` (frozen dataclass auto-repr больше не используется).

- [ ] **Step 2: Переписать гейт**

Удалить блок RuntimeError «Phase 0 requires TEST_MODE=true» (`config.py:26-30`). Вместо него, после чтения `test_mode`:
```python
if not test_mode and not qbittorrent_url:
    raise RuntimeError(
        "QBITTORRENT_URL is required when TEST_MODE=false "
        "(Web API base URL, e.g. http://192.168.1.50:8080)."
    )
```
Username/password опциональны (LAN без auth). Env: `QBITTORRENT_URL`, `QBITTORRENT_USERNAME`, `QBITTORRENT_PASSWORD`, `SAVE_PATH_MOVIES`, `SAVE_PATH_TV` (дефолты — текущие значения SAVE_PATHS).

- [ ] **Step 3: Новые тесты в `tests/unit/test_config.py`**

- `test_mode=false` без `QBITTORRENT_URL` → RuntimeError (match="QBITTORRENT_URL")
- `test_mode=false` с URL → `config.test_mode is False`, url сохранён
- `repr(config)` не содержит значения пароля (заданный `QBITTORRENT_PASSWORD`), содержит `"***"`
- дефолты save paths == `"Movies/"` / `"TV/"`; explicit env перекрывает

- [ ] **Step 4: Прогон**

```bash
pytest tests/unit/test_config.py -q && ruff check . && mypy app
```

- [ ] **Step 5: Commit**

```bash
git add app/config.py tests/unit/test_config.py
git commit -m "feat(config): allow TEST_MODE=false with QBITTORRENT_URL, add save path envs"
```

**Acceptance criteria:**
- [ ] `TEST_MODE=false` без URL падает; с URL — проходит; пароль маскирован в repr; дефолты путей совпадают с бывшим `SAVE_PATHS`.

---

### Task 3: контракт `get_info` — TorrentInfo + Protocol + Fake

**Files:**
- Modify: `app/models/torrent.py` (новый `TorrentInfo`)
- Modify: `app/providers/base.py`
- Modify: `app/providers/fake.py`

**Interfaces:**
- Consumes: `TorrentStatus` (models)
- Produces: `TorrentInfo(status, progress, download_speed, upload_speed, eta_seconds)`; Protocol `get_info(torrent_id) -> TorrentInfo` вместо `get_status`; FakeTorrentProvider с собственной симуляцией прогресса. Потребители: Task 4, 6, 7.

- [ ] **Step 1: Добавить `TorrentInfo` в `app/models/torrent.py`**

```python
@dataclass
class TorrentInfo:
    status: TorrentStatus
    progress: float = 0.0          # 0–100
    download_speed: int = 0        # B/s
    upload_speed: int = 0          # B/s
    eta_seconds: int | None = None
```

- [ ] **Step 2: Заменить метод в Protocol (`base.py`)**

`async def get_status(...) -> TorrentStatus` → `async def get_info(self, torrent_id: str) -> TorrentInfo: ...`. Импорт `TorrentInfo` вместо/вместе с `TorrentStatus` (TorrentStatus в Protocol больше не нужен — убрать).

- [ ] **Step 3: Адаптировать Fake (`fake.py`)**

- Rename `get_status` → `get_info`, возврат `TorrentInfo`.
- Тик-симуляция статусов не меняется: QUEUED → DOWNLOADING на 2-м тике, → COMPLETED на 8-м; pause/resume — как сейчас.
- Прогресс (сервис больше не накручивает): активные (QUEUED/DOWNLOADING) — `min(ticks * 12.5, 99.9)`; COMPLETED — `100.0`; PAUSED/FAILED/CANCELLED — текущее значение (при pause замораживается).
- Скорости: DOWNLOADING → `download_speed=1_048_576`, `upload_speed=131_072`, `eta_seconds=42`; иначе `0/0/None`.

- [ ] **Step 4: Прогон (ожидаемо красный до Task 4)**

```bash
mypy app
```

Ожидаемо: mypy по `app/` зелёный (сервис ещё зовёт `get_status`? — нет: сервис типизирован по Protocol только структурно; проверить компиляцию `mypy app` — если сервис зовёт `provider.get_status`, mypy ругается → это нормально, Task 4+6 чинят; допустимо пропустить до Task 6, если mypy падает только на сервисе — тогда в этом таске проверяем `mypy app/providers app/models`).

- [ ] **Step 5: Commit**

```bash
git add app/models/torrent.py app/providers/base.py app/providers/fake.py
git commit -m "feat(providers): extend TorrentProvider contract with get_info -> TorrentInfo"
```

**Acceptance criteria:**
- [ ] Protocol возвращает `TorrentInfo`; Fake детерминированно двигает статусы (2/8 тики) и прогресс.

---

### Task 4: починка сломанных тестов под `get_info`

**Files:**
- Modify: `tests/unit/test_fake_provider.py`
- Modify: `tests/unit/test_torrent_service.py`

**Interfaces:**
- Consumes: Task 3 (`get_info`, `TorrentInfo`, Fake-прогресс)
- Produces: зелёные unit-тесты Fake и сервиса.

- [ ] **Step 1: `test_fake_provider.py`**

- Все `await provider.get_status(...)` → `await provider.get_info(...)`, проверки через `.status`.
- Новые проверки: прогресс растёт от 0 к 100 (монотонно в QUEUED/DOWNLOADING); `COMPLETED` → `progress == 100.0`; pause фиксирует прогресс.

- [ ] **Step 2: `test_torrent_service.py`**

- Моки/стабы провайдера: вместо `get_status → TorrentStatus` возвращать `get_info → TorrentInfo(...)`.
- Убрать ожидания `+12.5%` от сервиса: ожидаемый `task.progress` = то, что вернул мок.
- Кейс «провайдер бросил исключение → FAILED» оставить (теперь на `get_info`).

- [ ] **Step 3: Прогон**

```bash
pytest tests/unit/test_fake_provider.py tests/unit/test_torrent_service.py -q
```

Ожидаемо: зелёные (интеграционные ещё могут падать из-за сервиса — их чинит Task 6; если падают уже сейчас из-за Fake-прогресса — дописать ожидания в `tests/integration/test_bot_flow.py` здесь же минимально).

- [ ] **Step 4: Commit**

```bash
git add tests/unit/test_fake_provider.py tests/unit/test_torrent_service.py
git commit -m "test: adapt fake provider and service tests to get_info contract"
```

**Acceptance criteria:**
- [ ] Unit-тесты Fake и сервиса зелёные на новом контракте.

---

### Task 5: схема + репозитории — колонка `provider_id`

**Files:**
- Modify: `app/db/schema.sql`
- Modify: `app/repositories/postgres.py`
- Modify: `app/repositories/memory.py` (проверка, без изменений интерфейса)
- Modify: `app/models/torrent.py` (удалить `ponytail:`-комментарий про колонку)

**Interfaces:**
- Consumes: `TorrentTask.provider_id` (уже есть)
- Produces: колонка `torrents.provider_id TEXT`; `insert_task` пишет и `_task_from_row` читает `provider_id`; `update_task` персистит `provider_id` (сигнатуры без изменений). Потребитель: Task 6.

- [ ] **Step 1: `schema.sql` — идемпотентная миграция**

В `CREATE TABLE torrents` добавить `provider_id TEXT` (пусто = NULL/`""` допустимо) + после таблицы, по паттерну IF NOT EXISTS схемы:
```sql
ALTER TABLE torrents ADD COLUMN IF NOT EXISTS provider_id TEXT;
```

- [ ] **Step 2: `postgres.py`**

- `insert_task`: INSERT включает `provider_id` (значение из task, обычно `""` → писать как есть или `NULL` если пусто — консистентно с чтением).
- `_task_from_row`: `provider_id=row["provider_id"] or ""` вместо хардкода `""`.
- `update_task`: убедиться, что UPDATE пишет `provider_id` (добавить в список колонок, если его нет).

- [ ] **Step 3: `memory.py` + models**

- Memory-репозиторий хранит `TorrentTask` целиком — `update_task` уже сохраняет `provider_id`; проверить, что нигде не делается копия без поля.
- В `app/models/torrent.py` удалить устаревший `ponytail:`-комментарий над `provider_id` (колонка теперь есть).

- [ ] **Step 4: Прогон (integration нужен postgres :55000)**

```bash
docker compose up postgres init-db   # если ещё не поднято
pytest -q
```

Ожидаемо: 49 passed (DB-тесты подняты, если Postgres доступен; иначе skip — допустимо).

- [ ] **Step 5: Commit**

```bash
git add app/db/schema.sql app/repositories/postgres.py app/models/torrent.py
git commit -m "feat(db): persist torrents.provider_id column"
```

**Acceptance criteria:**
- [ ] Свежая и существующая БД получают колонку; `provider_id` пишется/читается/обновляется обоими репозиториями.

---

### Task 6: TorrentService — persistence, cancel, get_info, save paths

**Files:**
- Modify: `app/services/torrent_service.py`
- Modify: `app/handlers/magnet.py`
- Modify: `tests/unit/test_torrent_service.py` (добить кейсы), при необходимости `tests/integration/test_bot_flow.py`

**Interfaces:**
- Consumes: Task 2 (`config.save_paths` — через main, Task 8; дефолты — константа внутри сервиса до того), Task 3 (`get_info`), Task 5 (`update_task` с provider_id)
- Produces: сервис без `_provider_ids` и `_PROGRESS_STEP`; `_ensure_provider` персистит `task.provider_id` через `update_task`; рестарт-сценарий: пустой `provider_id` + нетерминальный статус → повторный `add_magnet`; `cancel` удаляет торрент без файлов; `save_path_for()` для хендлеров; `_sync` пишет `TorrentInfo` в task.

- [ ] **Step 1: Конструктор — save paths**

`TorrentService.__init__`: новый параметр `save_paths: dict[MediaType, str] | None = None`; внутри `self._save_paths = save_paths or {MediaType.MOVIE: "Movies/", MediaType.SERIES: "TV/"}` (default — чтобы существующие тесты/вызовы не ломались; main передаст config.save_paths в Task 8). Убрать импорт `SAVE_PATHS` из media.py.

Публичный метод:
```python
def save_path_for(self, media_type: MediaType) -> str:
    return self._save_paths[media_type]
```

- [ ] **Step 2: `_ensure_provider` — персист вместо dict**

Удалить `self._provider_ids` (`:40`) и `pop` в `delete` (`:165`). Новая логика:
```python
if task.provider_id:
    return
task.provider_id = await self._provider.add_magnet(task.magnet, self._save_paths[task.media_type])
await self._repository.update_task(task)
```
Вызывается из `create` (уже есть) и `_sync` (уже есть) — после рестарта задача с пустым `provider_id` и нетерминальным статусом получает magnet заново (дедуп по info_hash на стороне qBittorrent). Терминальные статусы в `_sync` не дёргают `_ensure_provider` — проверить порядок проверок в `_sync`.

- [ ] **Step 3: `_sync` — реальные данные из `TorrentInfo`**

```python
info = await self._provider.get_info(task.provider_id)
task.status = info.status
task.progress = info.progress
task.download_speed = info.download_speed
task.upload_speed = info.upload_speed
task.eta_seconds = info.eta_seconds
```
Удалить `_PROGRESS_STEP` (`:15`) и ручной инкремент (`:221`). Прочую семантику `_sync` (TERMINAL, completed_at, update_task) — не трогать.

- [ ] **Step 4: `cancel` — удаление без файлов**

В `cancel` после проверки владельца/статуса, если `task.provider_id` непуст: `await self._provider.delete(task.provider_id, delete_files=False)` (исключение провайдера — как у delete: log + False/без изменений, по текущему паттерну сервиса). Затем `_change(task, CANCELLED)` как сейчас.

- [ ] **Step 5: `magnet.py` — путь из сервиса**

Убрать `SAVE_PATHS` из импорта; `f"💾 Путь:\n{SAVE_PATHS[media_type]}\n"` → `torrent_service.save_path_for(media_type)` (хендлер уже получает сервис из dispatcher data).

- [ ] **Step 6: Тесты**

`test_torrent_service.py`:
- `_ensure_provider` вызывает `add_magnet` один раз и `update_task` пишет provider_id; повторный `_sync` не добавляет magnet снова (задача с provider_id).
- Задача с пустым provider_id («после рестарта») → `add_magnet` вызывается.
- `cancel` вызывает `provider.delete(id, delete_files=False)` и ставит CANCELLED.
- `_sync` копирует поля `TorrentInfo` в task (включая progress от мока, не `+12.5`).

Интеграционные (если падали): минимальные правки ожиданий прогресса (теперь из Fake).

- [ ] **Step 7: Прогон**

```bash
pytest -q && ruff check . && ruff format --check . && mypy app
```

Ожидаемо: все зелёные.

- [ ] **Step 8: Commit**

```bash
git add app/services/torrent_service.py app/handlers/magnet.py tests/
git commit -m "feat(service): persist provider_id, cancel deletes torrent, sync from get_info"
```

**Acceptance criteria:**
- [ ] Нет `_provider_ids`/`_PROGRESS_STEP`; рестарт с пустым provider_id перерегистрирует magnet; cancel удаляет торрент без файлов; сервис пишет реальные/фейковые `TorrentInfo`-поля.

---

### Task 7: QBittorrentProvider + маппинг состояний

**Files:**
- Create: `app/providers/qbittorrent.py`
- Create: `tests/unit/test_qbittorrent_mapping.py`
- Create: `tests/unit/test_qbittorrent_provider.py`
- Modify: `pyproject.toml` (mypy override для `qbittorrentapi`, если нет py.typed — по итогу Task 1 Step 2)

**Interfaces:**
- Consumes: Task 1 (`qbittorrentapi`), Task 3 (`TorrentInfo`, Protocol)
- Produces: `QBittorrentProvider(url, username, password)` — `add_magnet`/`get_info`/`pause`/`resume`/`delete`; экспортируемая функция маппинга состояний. Потребитель: Task 8.

- [ ] **Step 1: Маппинг состояний (модульная функция, тестируемая без клиента)**

```python
_STATE_MAP: dict[str, TorrentStatus] = {
    "metadl": TorrentStatus.DOWNLOADING, "downloading": ..., "forceddl": ...,
    "stalleddl": ..., "checkingdl": ..., "moving": ..., "allocating": ...,
    "queueddl": TorrentStatus.QUEUED,
    "pauseddl": ..., "stoppeddl": ... → PAUSED,
    "pausedup": ..., "stoppedup": ..., "uploading": ..., "stalledup": ...,
    "forcedup": ..., "checkingup": ... → COMPLETED,
    "error": ..., "missingfiles": ... → FAILED,
}

def map_torrent_state(state: str, progress: float) -> TorrentStatus:
    if progress >= 100.0:
        return TorrentStatus.COMPLETED
    status = _STATE_MAP.get(state.lower())
    if status is None:
        logger.warning("qbittorrent: unknown state %r → FAILED", state)
        return TorrentStatus.FAILED
    return status
```

(Полный список статей — design §Маппинг; сравнение `state` case-insensitive.)

- [ ] **Step 2: `QBittorrentProvider`**

- `__init__(self, url: str, username: str = "", password: str = "")`: `self._client = qbittorrentapi.Client(url=url, username=username, password=password)` (клиент не соединяется до первого запроса). Клиент можно внедрить/подменить в тестах (`client: Client | None = None` параметр — сем для тестов).
- Каждый вызов Web API — через `asyncio.to_thread(...)` (библиотека синхронная).
- `add_magnet`: `torrents_add(urls=magnet, save_path=save_path)`; результат != `"Ok."` → `raise RuntimeError(...)` (без пароля в тексте). Иначе извлечь infohash из magnet regex'ом `xt=urn:btih:([0-9A-Fa-f]{40}|[A-Za-z2-7]{32})` и вернуть его как `torrent_id`; regex не совпал → RuntimeError.
- `get_info`: `torrents_info(torrent_hashes=torrent_id)`; пустой список → `LookupError` (сервис пометит FAILED); иначе первый торрент: `TorrentInfo(status=map_torrent_state(t.state, t.progress), progress=t.progress * 100, download_speed=t.dlspeed, upload_speed=t.upspeed, eta_seconds=None if t.eta in (-1, 8640000) else int(t.eta))`.
- `pause`: `torrents_pause(torrent_hashes=...)`; `resume`: `torrents_start(...)` (v5 stop/start alias — на стороне библиотеки, TF §9).
- `delete`: `torrents_delete(torrent_hashes=..., delete_files=delete_files)`.
- Исключения не глотаем — сервис маппит в FAILED.

- [ ] **Step 3: `test_qbittorrent_mapping.py`**

Таблицей: каждая строка design §Маппинг (state → expected), неизвестный state → FAILED, `progress=100` при любом state → COMPLETED.

- [ ] **Step 4: `test_qbittorrent_provider.py` (мок `Client`)**

- `add_magnet`: `"Ok."` → возврат хеша (magnet с hex-hash); не `"Ok."` → RuntimeError; regex-промах → RuntimeError.
- `get_info`: мок `torrents_info` с объектом-torrent (SimpleNamespace: state/progress/dlspeed/upspeed/eta) → корректный `TorrentInfo`; progress 0.42 → 42.0; eta=-1 → None; пустой список → LookupError.
- `pause/resume/delete`: проверка вызова клиента с нужными аргументами (включая `delete_files`).
- Пароль не появляется в `str(provider)` / сообщениях ошибок.

- [ ] **Step 5: Прогон**

```bash
pytest tests/unit/test_qbittorrent_mapping.py tests/unit/test_qbittorrent_provider.py -q && ruff check . && mypy app
```

- [ ] **Step 6: Commit**

```bash
git add app/providers/qbittorrent.py tests/unit/test_qbittorrent_mapping.py tests/unit/test_qbittorrent_provider.py pyproject.toml
git commit -m "feat(providers): add QBittorrentProvider with state mapping"
```

**Acceptance criteria:**
- [ ] Маппинг покрывает design-таблицу; провайдер покрыт юнит-тестами на моке; ни один сетевой вызов не идёт в тестах.

---

### Task 8: wiring в `app/main.py`

**Files:**
- Modify: `app/main.py`

**Interfaces:**
- Consumes: Task 2 (`Config.qbittorrent_*`, `config.save_paths`), Task 7 (`QBittorrentProvider`), FakeTorrentProvider (существующий)
- Produces: выбор провайдера по `config.test_mode`; сервис получает `save_paths=config.save_paths`.

- [ ] **Step 1: Выбор провайдера**

```python
provider: TorrentProvider = (
    FakeTorrentProvider()
    if config.test_mode
    else QBittorrentProvider(
        config.qbittorrent_url,
        config.qbittorrent_username,
        config.qbittorrent_password,
    )
)
```
(тип-аннотация/присвоение — по текущему стилю main.py; `TorrentProvider` Protocol импортируется для читаемости, если не нарушает mypy structural usage.)

- [ ] **Step 2: Передать save paths в сервис**

`TorrentService(repo, provider, save_paths=config.save_paths)` — по текущей сигнатуре сборки сервиса в main.

- [ ] **Step 3: Прогон**

```bash
pytest -q && ruff check . && mypy app
```

Плюс смоук без qBittorrent: `TEST_MODE=true` бот стартует как раньше (ручной `nx serve` по желанию).

- [ ] **Step 4: Commit**

```bash
git add app/main.py
git commit -m "feat(main): select provider by TEST_MODE, wire save paths from config"
```

**Acceptance criteria:**
- [ ] `TEST_MODE=true` → Fake (поведение не изменилось); `TEST_MODE=false` + URL → `QBittorrentProvider`; handlers не знают о выборе.

---

### Task 9: docs

**Files:**
- Modify: `.env.example`
- Modify: `docs/CONFIGURATION.md`
- Modify: `docs/TESTING.md`
- Modify: `docs/ROADMAP.md`
- Modify: `docs/API.md` (§2 — контракт провайдера)
- Modify: `docs/features/011-qbittorrent-provider.md` (статус/чекбоксы)
- Modify: `AGENTS.md`, `README.md` (снять «TEST_MODE=true обязателен / реальный qBittorrent не подключён»)

**Interfaces:**
- Consumes: Task 1–8 (фактическое поведение)
- Produces: документация соответствует коду.

- [ ] **Step 1: `.env.example`**

Добавить (с комментариями): `QBITTORRENT_URL`, `QBITTORRENT_USERNAME`, `QBITTORRENT_PASSWORD`, `SAVE_PATH_MOVIES=Movies/`, `SAVE_PATH_TV=TV/`; обновить комментарий у `TEST_MODE` (false валиден при заданном URL; пути видит qBittorrent — NFS/relative его конфига).

- [ ] **Step 2: `docs/CONFIGURATION.md`**

Таблица новых переменных: когда обязательны, дефолты, «пароль не логируется», «Web API только LAN».

- [ ] **Step 3: `docs/TESTING.md` — ручной сценарий**

Новый раздел «Реальный qBittorrent (TEST_MODE=false)»: предусловия (qBittorrent Web UI доступен, права, save path существует у qBittorrent), шаги (set env → `nx serve` → magnet → реальный прогресс/скорость/ETA в `/downloads` → pause/resume/cancel(проверить в Web UI, что торрент без файлов удалён) → delete → рестарт бота: задача с provider_id продолжает отслеживать существующий торрент; задача с удалённым из qBittorrent → FAILED). Автотесты остаются на Fake — оговорка.

- [ ] **Step 4: `docs/API.md` §2**

Контракт `TorrentProvider`: `get_info -> TorrentInfo` (поля), cancel-семантика (delete без файлов), сноска про provider_id.

- [ ] **Step 5: `ROADMAP.md` / `features/011` / `AGENTS.md` / `README.md`**

- ROADMAP: Phase 2 / FEAT-011 отметить выполненное (по мере Tasks; здесь — финальная сверка).
- `features/011-*.md`: проставить чекбоксы реализованного; статус → Implemented (после Task 10).
- `AGENTS.md` (app): строка «Режим: Phase 0/1 — TEST_MODE=true обязателен» → «TEST_MODE=true (Fake) по умолчанию; false подключает qBittorrent»; строка конфига «TEST_MODE=true — fail-fast, если false» → новое описание.
- `README.md` (app): «реальный протокол с qBittorrent — отдельная будущая фаза» → факт; счётчик тестов обновить по итогу Task 10.

- [ ] **Step 6: Commit**

```bash
git add .env.example docs/ AGENTS.md README.md
git commit -m "docs: document qBittorrent provider, config and manual test scenario"
```

**Acceptance criteria:**
- [ ] Ни одного утверждения «TEST_MODE=false невозможен»; новый env описан везде, где описаны старые.

---

### Task 10: финальная верификация

**Files:** (без изменений, кроме добитых мелочей по итогам прогона)

- [ ] **Step 1: Полный DoD**

```bash
# из корня монорепо
nx test python-torrent-bot
nx lint python-torrent-bot
nx typecheck python-torrent-bot
```

Альтернатива (из `apps/python-torrent-bot`, `.venv`):
```bash
pytest -q && ruff check . && ruff format --check . && mypy app
```

При недоступном PostgreSQL поднять: `docker compose up postgres init-db` (порт 55000) и повторить — integration не должен skip'аться.

- [ ] **Step 2: Сверка объёма**

- [ ] Нет `get_status` в `app/` и `tests/` (grep).
- [ ] Нет `_provider_ids` / `_PROGRESS_STEP` / `ponytail:` про provider_id-словарь.
- [ ] `SAVE_PATHS` больше не импортируется из `app/models/media.py` (grep).
- [ ] README-счётчик тестов = фактическое число passed.

- [ ] **Step 3: Ручная приёмка (описание, выполнить при доступной инфраструктуре)**

- [ ] `TEST_MODE=true`: чек-лист README (Fake) — без регрессий.
- [ ] `TEST_MODE=false` + живой qBittorrent: сценарий docs/TESTING.md (реальная загрузка, прогресс/скорость/ETA, cancel удаляет торрент в Web UI, рестарт бота сохраняет связь).

- [ ] **Step 4: Commit (если были мелкие правки)**

```bash
git add -A && git commit -m "chore: final FEAT-011 verification fixes"
```

**Acceptance criteria:**
- [ ] Все четыре DoD-команды зелёные; ручные сценарии подписаны; features/011 переведён в Implemented.
