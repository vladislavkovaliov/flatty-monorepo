# Phase 1 Persistent State (PostgreSQL) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Перевести state (users, torrent tasks, lifecycle events, settings) из in-memory в PostgreSQL через repository-слой, без изменений Telegram UI, с fail-fast стартом и авто-skip DB-тестами.

**Architecture:** `handlers → TorrentService → TorrentRepository (Protocol)` с двумя реализациями: `PostgresRepository` (psycopg 3 async pool, default) и `InMemoryRepository` (unit tests). Схема — `app/db/schema.sql`, применяется идемпотентно при старте (`CREATE TABLE IF NOT EXISTS`), без Alembic. Пул создаётся в lifespan `app/main.py`, закрывается в `finally`.

**Tech Stack:** Python 3.12+, aiogram 3.x, psycopg 3 (`psycopg[binary,pool]`), psycopg_pool.AsyncConnectionPool, pytest/pytest-asyncio (asyncio_mode=auto), ruff, mypy (strict, `files = ["app"]`).

**Spec:**
- `thoughts/shared/designs/2026-10-04-phase1-persistent-state-design.md` (authoritative design decisions)
- `docs/features/010-persistent-state.md` (requirements)
- `docs/ARCHITECTURE.md` §8, `docs/CONFIGURATION.md`, `docs/DEVELOPMENT.md` (DoD, agent rules)

## Global Constraints

- DoD: `ruff check .`, `ruff format --check .`, `mypy app`, `pytest` — все четыре команды зелёные (финальная проверка, Task 10).
- Существующие 20 тестов (`tests/unit/*`, `tests/integration/test_bot_flow.py`) остаются зелёными и не удаляются.
- Telegram UI/handlers (`app/handlers/*`) не изменяются — они не знают о database (ADR-0001); всё через `TorrentService`.
- `FakeTorrentProvider` не изменяется — меняется только слой хранения задач.
- FSM остаётся на `MemoryStorage` (явное решение design §Approach).
- `TEST_MODE=true` обязателен (config проверяет); реальных загрузок нет.
- Схема: только `users`, `torrents`, `torrent_events`, `settings`. Таблицы `media_items` НЕТ (FEAT-013). Миграции: НЕТ (Alembic не добавляем).
- `UNIQUE (user_id, info_hash)` в `torrents` — duplicate detection как constraint.
- `DATABASE_URL` обязателен: отсутствует/БД недоступна → fail-fast с внятной ошибкой; НИКОГДА не логировать пароль (redact до `user:***@host`).
- Migration Phase 0 данных: НЕ НУЖНА (in-memory тестовые данные) — состояние явно.
- Git: Conventional Commits; никогда не коммитить `.env`/секреты.
- ruff: line-length 100, rules E,F,I,UP,B,SIM; формат кода в `.md` не форматируется (extend-exclude).
- Тесты БД: `tests/integration/test_db.py` против реального `DATABASE_URL`, авто-skip если URL не задан или БД недоступна; тесты чистят только свои строки (никаких TRUNCATE всей БД).

## Task Dependency / Parallel Map

```text
Task 0 (git baseline, строго первый, один)
  └─ Wave A (параллельно): Task 1 (deps) | Task 2 (config) | Task 3 (schema) | Task 4 (protocol+memory)
       └─ Wave B (параллельно): Task 5 (pool, ждёт 1+3) | Task 6 (service refactor, ждёт 4)
            └─ Wave C: Task 7 (PostgresRepository, ждёт 3+5)
                 └─ Wave D (параллельно): Task 8 (DI wiring, ждёт 2+5+6+7) | Task 9 (test_db.py, ждёт 7)
                      └─ Task 10 (docs + финальная верификация, ждёт всё)
```

---

### Task 0: git init + Phase 0 baseline commit

**Files:**
- Modify: `.gitignore`

**Interfaces:**
- Consumes: —
- Produces: git-репозиторий с чистым Phase 0 baseline; все последующие задачи коммитят поверх.

- [ ] **Step 1: Обновить `.gitignore`** (сейчас в нём только `.env` — Phase 0 оставил артефакты сборки незаигнорированными)

```gitignore
.env
__pycache__/
*.py[cod]
.coverage
htmlcov/
.ruff_cache/
.mypy_cache/
.pytest_cache/
.venv/
```

- [ ] **Step 2: git init и первый коммит**

```bash
git init
git add -A
git ls-files | grep -x ".env"   # ожидаемо: пусто (.env не staged)
git commit -m "chore: initial commit — Phase 0 (Telegram UI + fake queue)"
```

Если git падает на отсутствии identity — `git config user.name "<name>" && git config user.email "<email>"` (локально).

**Acceptance criteria:**
- [ ] `git log --oneline` показывает 1 коммит; `git status` чистый.
- [ ] `git ls-files` не содержит `.env`, `__pycache__`, `.coverage`, `.ruff_cache`.
- [ ] В коммите: `app/`, `tests/`, `docs/`, `pyproject.toml`, `requirements.txt`, `.env.example`.

---

### Task 1: psycopg dependency

**Files:**
- Modify: `pyproject.toml` (`[project].dependencies`)
- Modify: `requirements.txt`

**Interfaces:**
- Consumes: —
- Produces: установленные `psycopg` + `psycopg_pool` (импортуются в Task 5/7/9).

- [ ] **Step 1: Добавить зависимость в `pyproject.toml`**

```toml
dependencies = [
    "aiogram>=3.25,<4",
    "python-dotenv>=1.0",
    "psycopg[binary,pool]>=3.3,<4",
]
```

`pool` extra подтверждён: `requires_dist` содержит `psycopg-pool; extra == "pool"` (PyPI, psycopg 3.3.x). `binary` — wheels без компиляции libpq.

- [ ] **Step 2: Добавить в `requirements.txt`** (файл держит runtime-депы, комментарий «синхронизировать с pyproject» уже есть)

```text
aiogram>=3.25,<4
python-dotenv>=1.0
psycopg[binary,pool]>=3.3,<4
```

- [ ] **Step 3: Установить и проверить импорт**

```bash
pip install -r requirements.txt
python -c "import psycopg, psycopg_pool; print(psycopg.__version__)"
```

Ожидаемо: версия 3.3.x, без ошибок.

- [ ] **Step 4: Быстрая проверка, что ничего не сломали**

```bash
ruff check . && ruff format --check . && mypy app && pytest -q
```

Ожидаемо: 20 passed, остальное чисто.

- [ ] **Step 5: Commit**

```bash
git add pyproject.toml requirements.txt
git commit -m "build: add psycopg 3 dependency"
```

**Acceptance criteria:**
- [ ] `python -c "import psycopg, psycopg_pool"` работает.
- [ ] Все 20 существующих тестов зелёные.

---

### Task 2: config DATABASE_URL + .env.example + CONFIGURATION.md

**Files:**
- Modify: `app/config.py`
- Modify: `.env.example`
- Modify: `docs/CONFIGURATION.md`
- Test: `tests/unit/test_config.py` (новый)
- Local only (НЕ коммитится): `.env` — добавить реальный `DATABASE_URL`

**Interfaces:**
- Consumes: —
- Produces: `Config.database_url: str` (garantированно non-empty, схема `postgres*://`) — используется Task 8.

- [ ] **Step 1: Написать failing-тест** `tests/unit/test_config.py`

```python
from __future__ import annotations

import pytest

from app.config import load_config


def test_missing_database_url_raises(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setattr("app.config.load_dotenv", lambda: None)
    monkeypatch.setenv("BOT_TOKEN", "123:TEST")
    monkeypatch.setenv("TEST_MODE", "true")
    monkeypatch.delenv("DATABASE_URL", raising=False)
    with pytest.raises(RuntimeError, match="DATABASE_URL is not set"):
        load_config()


def test_invalid_scheme_raises(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setattr("app.config.load_dotenv", lambda: None)
    monkeypatch.setenv("BOT_TOKEN", "123:TEST")
    monkeypatch.setenv("TEST_MODE", "true")
    monkeypatch.setenv("DATABASE_URL", "mysql://user:pass@host/db")
    with pytest.raises(RuntimeError, match="postgresql://"):
        load_config()


def test_valid_config_carries_database_url(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setattr("app.config.load_dotenv", lambda: None)
    monkeypatch.setenv("BOT_TOKEN", "123:TEST")
    monkeypatch.setenv("TEST_MODE", "true")
    url = "postgresql://user:pass@192.168.1.35:5432/torrent_bot?sslmode=disable"
    monkeypatch.setenv("DATABASE_URL", url)
    config = load_config()
    assert config.database_url == url
    assert config.test_mode is True
```

`load_dotenv` замокан, чтобы `.env` исполнителя не вливал `DATABASE_URL` в тест.

- [ ] **Step 2: Запустить — убедиться, что падает**

```bash
pytest tests/unit/test_config.py -v
```

Ожидаемо: FAIL (`Config.__init__() missing 1 required positional argument` / `DATABASE_URL is not set`).

- [ ] **Step 3: Изменить `app/config.py`**

```python
@dataclass(frozen=True)
class Config:
    bot_token: str
    test_mode: bool
    database_url: str


def load_config() -> Config:
    load_dotenv()
    token = os.getenv("BOT_TOKEN")
    if not token:
        raise RuntimeError("BOT_TOKEN is not set. Copy .env.example to .env and fill BOT_TOKEN.")
    test_mode = os.getenv("TEST_MODE", "true").strip().lower() in {"1", "true", "yes"}
    if not test_mode:
        raise RuntimeError(
            "Phase 0 requires TEST_MODE=true: real providers (qBittorrent/NAS) "
            "are not implemented yet."
        )
    database_url = os.getenv("DATABASE_URL", "").strip()
    if not database_url:
        raise RuntimeError(
            "DATABASE_URL is not set. Copy .env.example to .env and fill DATABASE_URL "
            "(format: postgresql://user:pass@host:5432/dbname?sslmode=disable)."
        )
    if not database_url.startswith(("postgresql://", "postgres://")):
        raise RuntimeError(
            "DATABASE_URL must start with postgresql:// — other URL schemes are not supported."
        )
    return Config(bot_token=token, test_mode=test_mode, database_url=database_url)
```

Валидация формата намеренно минимальна — design: «format accepted as-is» (libpq URL принимается psycopg без изменений).

- [ ] **Step 4: Запустить — убедиться, что зелёные**

```bash
pytest tests/unit/test_config.py -v
```

Ожидаемо: 3 passed.

- [ ] **Step 5: `.env.example`** — добавить строку (и комментарий про отдельную БД):

```dotenv
BOT_TOKEN=123456:REPLACE_WITH_TOKEN_FROM_BOTFATHER
TEST_MODE=true
# Рекомендуется отдельная БД (например torrent_bot) вместо public — design §Documentation
DATABASE_URL=postgresql://postgres:password@192.168.1.35:5432/torrent_bot?sslmode=disable
```

- [ ] **Step 6: Локальный `.env` (НЕ коммитится)** — добавить строку `DATABASE_URL=...` с реальными кредами пользователя. Если реального URL ещё нет — спросить пользователя; без него DB-тесты (Task 9) будут skip'аться, а ручной persistence-тест (Task 10) не пройти.

- [ ] **Step 7: `docs/CONFIGURATION.md`** — в секцию «Текущая конфигурация (MVP)» добавить:

```text
DATABASE_URL=postgresql://user:pass@host:5432/dbname?sslmode=disable
```

и абзац: `DATABASE_URL` обязателен (fail-fast при старте), формат — libpq URL (принимается как есть), пароль никогда не логируется (redaction до `user:***@host`), рекомендуется отдельная БД `torrent_bot`. Убрать `DATABASE_URL=` из «Будущая конфигурация» (переменная уже не будущая).

- [ ] **Step 8: Прогон + commit**

```bash
ruff check . && ruff format . && mypy app && pytest -q
git add app/config.py .env.example docs/CONFIGURATION.md tests/unit/test_config.py
git commit -m "feat: require DATABASE_URL in config"
```

**Acceptance criteria:**
- [ ] `pytest tests/unit/test_config.py` — 3 passed; полный `pytest` — все зелёные.
- [ ] Без `DATABASE_URL` бот падает сразу с сообщением «DATABASE_URL is not set…».
- [ ] `.env` не попал в git (`git ls-files | grep -x ".env"` пусто).

---

### Task 3: schema.sql

**Files:**
- Create: `app/db/__init__.py` (пустой)
- Create: `app/db/schema.sql`

**Interfaces:**
- Consumes: —
- Produces: SQL-скрипт, применяемый Task 5 `apply_schema()`; таблицы `users`, `torrents`, `torrent_events`, `settings`.

- [ ] **Step 1: Создать `app/db/__init__.py`** (пустой файл — пакет).

- [ ] **Step 2: Создать `app/db/schema.sql`**

```sql
-- ponytail: идемпотентный schema.sql вместо Alembic (4 таблицы);
-- ceiling: нет миграций схемы — upgrade path: Alembic, когда схема начнёт часто меняться.

CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    telegram_user_id BIGINT NOT NULL UNIQUE,
    username TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS torrents (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    magnet TEXT NOT NULL,
    info_hash TEXT NOT NULL,
    name TEXT NOT NULL,
    media_type TEXT NOT NULL,
    quality TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'queued',
    progress DOUBLE PRECISION NOT NULL DEFAULT 0,
    size_bytes BIGINT,
    download_speed BIGINT NOT NULL DEFAULT 0,
    upload_speed BIGINT NOT NULL DEFAULT 0,
    eta_seconds BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at TIMESTAMPTZ,
    -- duplicate detection как constraint (design §Components)
    UNIQUE (user_id, info_hash)
);

CREATE TABLE IF NOT EXISTS torrent_events (
    id BIGSERIAL PRIMARY KEY,
    torrent_id BIGINT NOT NULL REFERENCES torrents(id) ON DELETE CASCADE,
    event TEXT NOT NULL,
    payload TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_torrent_events_torrent_id ON torrent_events (torrent_id);

-- таблица и repository есть; UI настроек нет (не в Phase 1)
CREATE TABLE IF NOT EXISTS settings (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    PRIMARY KEY (user_id, key)
);
```

Замечания, обязательные к пониманию исполнителем:
- Колонки ровно по design §Components (плюс два индекса). `media_items` НЕ создаётся. Колонки `provider_id` в `torrents` НЕТ по design — `provider_id` это session state (см. Task 6).
- Дополнительный индекс на `torrents(user_id)` не нужен: `UNIQUE (user_id, info_hash)` уже создаёт B-tree по `user_id` (prefix). Индекс на `torrent_events(torrent_id)` нужен — FK не индексируется в PG автоматически.
- Прогресс-тики НЕ пишутся в `torrent_events` — только lifecycle (created/status_changed/deleted).

- [ ] **Step 3: Синтаксическая проверка** (если доступна psql-цепочка — не обязательна, полноценная проверка в Task 9):

```bash
python -c "print(open('app/db/schema.sql').read()[:80])"
```

- [ ] **Step 4: Commit**

```bash
git add app/db/__init__.py app/db/schema.sql
git commit -m "feat: add database schema"
```

**Acceptance criteria:**
- [ ] Схема содержит ровно 4 таблицы, `UNIQUE (user_id, info_hash)`, никаких `media_items`.
- [ ] Каждый `CREATE ... IF NOT EXISTS` — повторное применение безопасно (идемпотентность проверяется в Task 9 `test_schema_applied_idempotent`).

---

### Task 4: repository protocol + InMemoryRepository

**Files:**
- Create: `app/repositories/__init__.py` (протокол)
- Create: `app/repositories/memory.py`

**Interfaces:**
- Consumes: `app/models/torrent.py::TorrentTask`, `TorrentStatus`.
- Produces (используют Task 6/7/9):

```python
class TorrentRepository(Protocol):
    async def find_by_hash(self, user_id: int, info_hash: str) -> TorrentTask | None: ...
    async def insert_task(self, task: TorrentTask) -> TorrentTask | None: ...
    async def get_task(self, user_id: int, task_id: int) -> TorrentTask | None: ...
    async def list_tasks(self, user_id: int) -> list[TorrentTask]: ...
    async def update_task(self, task: TorrentTask) -> None: ...
    async def delete_task(self, task_id: int) -> None: ...
    async def add_event(self, torrent_id: int, event: str, payload: str | None = None) -> None: ...
    async def get_setting(self, user_id: int, key: str) -> str | None: ...
    async def set_setting(self, user_id: int, key: str, value: str) -> None: ...
```

Семантика контракта:
- `user_id` везде — это **telegram user id** (как в handlers). Postgres-реализация сама мапит его на `users.id` ( surrogate key ) через JOIN; InMemory хранит как есть.
- `insert_task` → `None` означает конфликт `UNIQUE(user_id, info_hash)` (аналог `ON CONFLICT DO NOTHING`); вызывающий (service) в этом случае возвращает существующую задачу.
- `get_task`/`list_tasks` фильтруют по владельцу (FEAT-009: чужое = «не найдено»).

- [ ] **Step 1: Создать `app/repositories/__init__.py`**

```python
from __future__ import annotations

from typing import Protocol

from app.models.torrent import TorrentTask


class TorrentRepository(Protocol):
    """Хранение состояния за TorrentService (design §Architecture); handlers не видят БД."""

    async def find_by_hash(self, user_id: int, info_hash: str) -> TorrentTask | None: ...
    async def insert_task(self, task: TorrentTask) -> TorrentTask | None: ...
    async def get_task(self, user_id: int, task_id: int) -> TorrentTask | None: ...
    async def list_tasks(self, user_id: int) -> list[TorrentTask]: ...
    async def update_task(self, task: TorrentTask) -> None: ...
    async def delete_task(self, task_id: int) -> None: ...
    async def add_event(self, torrent_id: int, event: str, payload: str | None = None) -> None: ...
    async def get_setting(self, user_id: int, key: str) -> str | None: ...
    async def set_setting(self, user_id: int, key: str, value: str) -> None: ...
```

- [ ] **Step 2: Создать `app/repositories/memory.py`**

```python
from __future__ import annotations

import itertools

from app.models.torrent import TorrentTask


class InMemoryRepository:
    """Тестовый double: тот же контракт, что и PostgresRepository, без инфраструктуры."""

    def __init__(self) -> None:
        self._tasks: dict[int, TorrentTask] = {}
        self._ids = itertools.count(1)
        self.events: list[tuple[int, str, str | None]] = []
        self._settings: dict[tuple[int, str], str] = {}

    async def find_by_hash(self, user_id: int, info_hash: str) -> TorrentTask | None:
        for task in self._tasks.values():
            if task.user_id == user_id and task.info_hash == info_hash:
                return task
        return None

    async def insert_task(self, task: TorrentTask) -> TorrentTask | None:
        if await self.find_by_hash(task.user_id, task.info_hash) is not None:
            return None  # как ON CONFLICT (user_id, info_hash) DO NOTHING
        task.id = next(self._ids)
        self._tasks[task.id] = task
        return task

    async def get_task(self, user_id: int, task_id: int) -> TorrentTask | None:
        task = self._tasks.get(task_id)
        if task is None or task.user_id != user_id:
            return None
        return task

    async def list_tasks(self, user_id: int) -> list[TorrentTask]:
        return [t for t in self._tasks.values() if t.user_id == user_id]

    async def update_task(self, task: TorrentTask) -> None:
        # объекты хранятся по ссылке — мутации уже видны; совпадает с контрактом «персист»
        return None

    async def delete_task(self, task_id: int) -> None:
        self._tasks.pop(task_id, None)

    async def add_event(self, torrent_id: int, event: str, payload: str | None = None) -> None:
        self.events.append((torrent_id, event, payload))

    async def get_setting(self, user_id: int, key: str) -> str | None:
        return self._settings.get((user_id, key))

    async def set_setting(self, user_id: int, key: str, value: str) -> None:
        self._settings[(user_id, key)] = value
```

Отдельный тест-файл для InMemory НЕ пишем (YAGNI): контракт покрывается тестами сервиса (Task 6, где все прежние 20 тестов бегут именно на этом double) и тестами Postgres (Task 9).

- [ ] **Step 3: Проверка**

```bash
ruff check . && ruff format . && mypy app && pytest -q
```

Ожидаемо: всё зелёное (новый код ещё никем не используется, но mypy/ruff его видят).

- [ ] **Step 4: Commit**

```bash
git add app/repositories/
git commit -m "feat: add repository protocol and in-memory implementation"
```

**Acceptance criteria:**
- [ ] `mypy app` strict чистый (Protocol с `...` телами).
- [ ] `pytest` — прежние 20 зелёные.

---

### Task 5: pool + schema apply (app/db/pool.py)

**Files:**
- Create: `app/db/pool.py`
- Test: `tests/unit/test_db_pool.py` (новый)

**Interfaces:**
- Consumes: Task 1 (`psycopg_pool`), Task 3 (`app/db/schema.sql`).
- Produces (используют Task 8/9):

```python
def redact_database_url(url: str) -> str: ...
async def create_pool(database_url: str) -> AsyncConnectionPool: ...   # fail-fast при недоступности
async def apply_schema(pool: AsyncConnectionPool) -> None: ...         # идемпотентно
```

- [ ] **Step 1: Написать failing-тест** `tests/unit/test_db_pool.py`

```python
from __future__ import annotations

from app.db.pool import redact_database_url


def test_redact_password() -> None:
    url = "postgresql://postgres:s3cret@192.168.1.35:5432/torrent_bot?sslmode=disable"
    redacted = redact_database_url(url)
    assert "s3cret" not in redacted
    assert redacted == "postgresql://postgres:***@192.168.1.35:5432/torrent_bot?sslmode=disable"


def test_redact_without_password_is_noop() -> None:
    assert redact_database_url("postgresql://localhost/torrent_bot") == "postgresql://localhost/torrent_bot"
```

- [ ] **Step 2: Запустить — убедиться, что падает**

```bash
pytest tests/unit/test_db_pool.py -v
```

Ожидаемо: FAIL (`ModuleNotFoundError: app.db.pool`).

- [ ] **Step 3: Создать `app/db/pool.py`**

```python
from __future__ import annotations

import re
from pathlib import Path

from psycopg_pool import AsyncConnectionPool

_SCHEMA_PATH = Path(__file__).with_name("schema.sql")
# DATABASE_URL содержит пароль — логировать/показывать только в отредактированном виде (design §Error Handling)
_PASSWORD_RE = re.compile(r"://([^:/@]+):[^@/]+@")


def redact_database_url(url: str) -> str:
    return _PASSWORD_RE.sub(r"://\1:***@", url)


async def create_pool(database_url: str) -> AsyncConnectionPool:
    pool = AsyncConnectionPool(
        database_url,
        min_size=1,
        max_size=5,
        open=False,
        kwargs={"connect_timeout": 5},
    )
    try:
        await pool.open(wait=True, timeout=10)
    except Exception as exc:
        await pool.close()
        raise RuntimeError(
            f"Cannot reach PostgreSQL at {redact_database_url(database_url)}: {exc}"
        ) from exc
    return pool


async def apply_schema(pool: AsyncConnectionPool) -> None:
    script = _SCHEMA_PATH.read_text(encoding="utf-8")
    statements = [s.strip() for s in script.split(";") if s.strip()]
    try:
        async with pool.connection() as conn:
            for statement in statements:
                await conn.execute(statement)
    except Exception as exc:
        # URL уже прошёл проверку в create_pool; в сообщении его нет — пароль не утекает
        raise RuntimeError(f"Cannot apply database schema: {exc}") from exc
```

Пояснения:
- `kwargs={"connect_timeout": 5}` — kwargs прокидываются в `connect()` (API `AsyncConnectionPool`), ограничивает каждую попытку соединения.
- Схема режется по `;` и исполняется посказательно внутри одного `pool.connection()` — одна транзакция PG: при ошибке DDL откатывается целиком. (Всё-в-одном `conn.execute(script)` тоже работает без параметров в psycopg 3, но сплит детерминирован и не зависит от протокола биндинга.)
- `create_pool` при недоступной БД кидает `RuntimeError` с **redacted** URL — это и есть fail-fast (Task 8).

- [ ] **Step 4: Запустить — зелёные**

```bash
pytest tests/unit/test_db_pool.py -v
```

Ожидаемо: 2 passed.

- [ ] **Step 5: Прогон + commit**

```bash
ruff check . && ruff format . && mypy app && pytest -q
git add app/db/pool.py tests/unit/test_db_pool.py
git commit -m "feat: add connection pool with fail-fast and schema apply"
```

**Acceptance criteria:**
- [ ] `redact_database_url` вырезает пароль; тесты зелёные.
- [ ] `mypy app` чистый (типы `AsyncConnectionPool`).
- [ ] Нигде в коде/логах не выводится сырая строка `DATABASE_URL`.

---

### Task 6: TorrentService → repository (рефакторинг)

**Files:**
- Modify: `app/services/torrent_service.py` (полная переработка внутренностей, публичные сигнатуры методов БЕЗ ИЗМЕНЕНИЙ)
- Modify: `tests/unit/test_torrent_service.py` (fixture: + repository)
- Modify: `tests/integration/test_bot_flow.py` (только `_setup()` — + repository)
- Test additions в `tests/unit/test_torrent_service.py`

**Interfaces:**
- Consumes: Task 4 (`TorrentRepository`, `InMemoryRepository`).
- Produces: `TorrentService(provider: TorrentProvider, repository: TorrentRepository)` — использует Task 8; публичные методы `create/list/get/pause/resume/cancel/delete` без изменений (handlers не меняются).

Ключевые изменения поведения:
1. `self._tasks`/`self._next_id` удаляются — всё через repository.
2. Duplicate: `find_by_hash` → при конфликте `insert_task` вернёт `None` → вернуть существующую (`created=False`).
3. Lifecycle-события: `created`, `status_changed` (payload JSON `{"from": ..., "to": ...}`), `deleted` — НЕ на каждый тик прогресса.
4. **`provider_id` — session state**: колонки в БД нет (design §Components), `FakeTorrentProvider` хранит состояние в памяти процесса. Поэтому сервис держит карту `task_id → provider_id` на время сессии: при загрузке задачи из БД (рестарт) провайдер повторно регистрирует magnet ОДИН раз на задачу за жизнь процесса. Иначе после рестарта `provider.get_status()` падал бы по KeyError и все задачи становились FAILED (ручной persistence-тест из Task 10 требует, чтобы задача «пережила» рестарт).

- [ ] **Step 1: Добавить новые тесты в `tests/unit/test_torrent_service.py`** (сначала красные)

Заменить fixture:

```python
import pytest

from app.providers.fake import FakeTorrentProvider
from app.repositories.memory import InMemoryRepository
from app.services.torrent_service import TorrentService

@pytest.fixture
def repo() -> InMemoryRepository:
    return InMemoryRepository()


@pytest.fixture
def service(repo: InMemoryRepository) -> TorrentService:
    return TorrentService(FakeTorrentProvider(), repo)
```

Добавить в конец файла:

```python
async def test_provider_reregistered_after_restart(repo: InMemoryRepository) -> None:
    # «рестарт»: задача уже в хранилище, новый сервис + новый (пустой) provider
    seeded = TorrentTask(
        id=0,
        user_id=1,
        magnet=f"magnet:?xt=urn:btih:{HASH_A}",
        info_hash=HASH_A,
        name="Restored",
        media_type=MediaType.MOVIE,
        quality=Quality.FHD_1080,
    )
    assert await repo.insert_task(seeded) is not None
    service = TorrentService(FakeTorrentProvider(), repo)
    tasks = await service.list(1)
    assert [t.name for t in tasks] == ["Restored"]
    assert tasks[0].provider_id.startswith("fake-")


async def test_status_changes_record_lifecycle_events(
    service: TorrentService, repo: InMemoryRepository
) -> None:
    task, _ = await service.create(
        user_id=1, magnet=_magnet(), media_type=MediaType.MOVIE, quality=Quality.FHD_1080
    )
    await service.pause(1, task.id)
    await service.cancel(1, task.id)
    await service.delete(1, task.id)
    kinds = [event for _tid, event, _payload in repo.events]
    assert kinds == ["created", "status_changed", "status_changed", "deleted"]
```

Импорт `TorrentTask` в тест-файле добавить (`from app.models.torrent import TorrentTask` — `TorrentStatus` уже импортирован).

- [ ] **Step 2: Запустить — красные**

```bash
pytest tests/unit/test_torrent_service.py -v
```

Ожидаемо: FAIL — `TorrentService.__init__() takes 1 positional argument but 3 were given` / старые тесты падают из-за нового обязательного параметра.

- [ ] **Step 3: Переписать `app/services/torrent_service.py`**

```python
from __future__ import annotations

import json
import logging
from datetime import UTC, datetime

from app.models.media import SAVE_PATHS, MediaType, Quality
from app.models.torrent import TorrentStatus, TorrentTask
from app.providers.base import TorrentProvider
from app.repositories import TorrentRepository
from app.services.magnet_parser import MagnetInfo

logger = logging.getLogger(__name__)

_PROGRESS_STEP = 12.5
# PAUSED/COMPLETED/FAILED/CANCELLED не опрашиваются у provider (нет ложных переходов)
_TERMINAL = frozenset(
    {
        TorrentStatus.PAUSED,
        TorrentStatus.COMPLETED,
        TorrentStatus.FAILED,
        TorrentStatus.CANCELLED,
    }
)


class TorrentService:
    """Бизнес-логика очереди: валидация, duplicate по info_hash, ownership, вызовы provider.

    Хранение — за TorrentRepository (design §Architecture); handlers/Telegram не знают о БД.
    Без Telegram-кода (docs/API.md §3).
    """

    def __init__(self, provider: TorrentProvider, repository: TorrentRepository) -> None:
        self._provider = provider
        self._repo = repository
        # ponytail: provider_id — session state, колонки в БД нет (design §Components),
        # FakeTorrentProvider живёт в памяти процесса; upgrade path — FEAT-011 добавит
        # колонку torrents.provider_id и регистрацию в qBittorrent при загрузке из БД
        self._provider_ids: dict[int, str] = {}

    async def create(
        self,
        *,
        user_id: int,
        magnet: MagnetInfo,
        media_type: MediaType,
        quality: Quality,
    ) -> tuple[TorrentTask, bool]:
        existing = await self._repo.find_by_hash(user_id, magnet.info_hash)
        if existing is not None:
            logger.info(
                "duplicate info_hash=%s user_id=%s task_id=%s",
                magnet.info_hash,
                user_id,
                existing.id,
            )
            await self._sync(existing)
            return existing, False

        task = TorrentTask(
            id=0,
            user_id=user_id,
            magnet=magnet.raw,
            info_hash=magnet.info_hash,
            name=magnet.display_name or f"magnet-{magnet.info_hash[:8]}",
            media_type=media_type,
            quality=quality,
            size_bytes=magnet.size_bytes,
        )
        inserted = await self._repo.insert_task(task)
        if inserted is None:
            # гонка на UNIQUE(user_id, info_hash): кто-то вставил между find и insert
            existing = await self._repo.find_by_hash(user_id, magnet.info_hash)
            if existing is None:
                raise RuntimeError(f"torrent conflict resolved but row missing: {magnet.info_hash}")
            logger.info(
                "duplicate info_hash=%s user_id=%s task_id=%s",
                magnet.info_hash,
                user_id,
                existing.id,
            )
            return existing, False
        task = inserted

        try:
            await self._ensure_provider(task)
        except Exception:
            logger.error("provider.add_magnet failed task_id=%s", task.id, exc_info=True)
            await self._change(task, TorrentStatus.FAILED)
            return task, True

        await self._repo.add_event(task.id, "created")
        logger.info(
            "created task_id=%s user_id=%s info_hash=%s", task.id, user_id, magnet.info_hash
        )
        return task, True

    async def list(self, user_id: int) -> list[TorrentTask]:
        tasks = await self._repo.list_tasks(user_id)
        for task in tasks:
            await self._sync(task)
        return tasks

    async def get(self, user_id: int, task_id: int) -> TorrentTask | None:
        return await self._owned(user_id, task_id)

    async def pause(self, user_id: int, task_id: int) -> TorrentTask | None:
        task = await self._owned(user_id, task_id)
        if task is None or task.status not in (
            TorrentStatus.QUEUED,
            TorrentStatus.DOWNLOADING,
        ):
            return task
        try:
            await self._provider.pause(task.provider_id)
        except Exception:
            logger.error("provider.pause failed task_id=%s", task.id, exc_info=True)
            await self._change(task, TorrentStatus.FAILED)
            return task
        await self._change(task, TorrentStatus.PAUSED)
        return task

    async def resume(self, user_id: int, task_id: int) -> TorrentTask | None:
        task = await self._owned(user_id, task_id)
        if task is None or task.status is not TorrentStatus.PAUSED:
            return task
        try:
            await self._provider.resume(task.provider_id)
        except Exception:
            logger.error("provider.resume failed task_id=%s", task.id, exc_info=True)
            await self._change(task, TorrentStatus.FAILED)
            return task
        await self._change(task, TorrentStatus.DOWNLOADING)
        return task

    async def cancel(self, user_id: int, task_id: int) -> TorrentTask | None:
        task = await self._owned(user_id, task_id)
        if task is None:
            return None
        # ponytail: cancel без вызова provider — в контракте API.md §2 метода нет;
        # upgrade path: FEAT-011 добавит cancel в TorrentProvider
        await self._change(task, TorrentStatus.CANCELLED)
        logger.info("cancelled task_id=%s user_id=%s", task_id, user_id)
        return task

    async def delete(self, user_id: int, task_id: int) -> bool:
        task = await self._owned(user_id, task_id)
        if task is None:
            return False
        try:
            await self._provider.delete(task.provider_id)
        except Exception:
            logger.error("provider.delete failed task_id=%s", task.id, exc_info=True)
            return False
        # событие ДО удаления: torrent_events.torrent_id — FK с ON DELETE CASCADE
        await self._repo.add_event(task.id, "deleted")
        await self._repo.delete_task(task_id)
        self._provider_ids.pop(task_id, None)
        logger.info("deleted task_id=%s user_id=%s", task_id, user_id)
        return True

    async def _owned(self, user_id: int, task_id: int) -> TorrentTask | None:
        task = await self._repo.get_task(user_id, task_id)
        if task is None:
            # FEAT-009: чужая/несуществующая — трактовать как «не найдено»
            return None
        await self._sync(task)
        return task

    async def _ensure_provider(self, task: TorrentTask) -> None:
        """Регистрирует magnet у provider один раз за жизнь процесса (task_id → provider_id)."""
        provider_id = self._provider_ids.get(task.id)
        if provider_id is None:
            provider_id = await self._provider.add_magnet(task.magnet, SAVE_PATHS[task.media_type])
            self._provider_ids[task.id] = provider_id
        task.provider_id = provider_id

    async def _change(self, task: TorrentTask, new_status: TorrentStatus) -> None:
        if new_status is not task.status:
            await self._repo.add_event(
                task.id,
                "status_changed",
                json.dumps({"from": task.status, "to": new_status}),
            )
        task.status = new_status
        await self._repo.update_task(task)

    async def _sync(self, task: TorrentTask) -> None:
        try:
            await self._ensure_provider(task)
        except Exception:
            logger.error("provider.add_magnet failed task_id=%s", task.id, exc_info=True)
            await self._change(task, TorrentStatus.FAILED)
            return
        if task.status in _TERMINAL:
            return
        try:
            status = await self._provider.get_status(task.provider_id)
        except Exception:
            logger.error("provider.get_status failed task_id=%s", task.id, exc_info=True)
            await self._change(task, TorrentStatus.FAILED)
            return
        if status is not task.status:
            # lifecycle-событие только на смене статуса, НЕ на тике прогресса (design §Components)
            await self._repo.add_event(
                task.id, "status_changed", json.dumps({"from": task.status, "to": status})
            )
            task.status = status
        if status is TorrentStatus.COMPLETED:
            task.progress = 100.0
            task.completed_at = datetime.now(UTC)
        elif status is TorrentStatus.DOWNLOADING:
            # ponytail: фейковый прогресс +12.5% за рендер; upgrade path — данные из provider
            task.progress = min(99.0, task.progress + _PROGRESS_STEP)
        await self._repo.update_task(task)
```

- [ ] **Step 4: Обновить `tests/integration/test_bot_flow.py` — только `_setup()`**

```python
from app.repositories.memory import InMemoryRepository

def _setup() -> tuple[Dispatcher, RecordingSession, Bot, TorrentService]:
    service = TorrentService(FakeTorrentProvider(), InMemoryRepository())
    dp = create_dispatcher(service=service, test_mode=True)
    session = RecordingSession()
    bot = Bot(token="42:TEST", session=session)
    return dp, session, bot, service
```

Остальной код теста не меняется (публичные сигнатуры сервиса не изменились).

- [ ] **Step 5: Полный прогон — все тесты зелёные**

```bash
pytest -v
```

Ожидаемо: 20 прежних + 2 новых = 22 passed, 0 failed.
Если что-то красное — чинить в сервисе/тестах, НЕ удаляя существующие тесты (DEVELOPMENT §3.8).

- [ ] **Step 6: Прогон + commit**

```bash
ruff check . && ruff format . && mypy app && pytest -q
git add app/services/torrent_service.py tests/unit/test_torrent_service.py tests/integration/test_bot_flow.py
git commit -m "refactor: route torrent service through repository"
```

**Acceptance criteria:**
- [ ] `TorrentService` не содержит dict-хранилища (`grep -n "_tasks\|_next_id" app/services/torrent_service.py` → пусто).
- [ ] Все тесты зелёные (22), включая restart-тест (задача из «БД» получает provider_id) и lifecycle-тест (created → status_changed ×2 → deleted).
- [ ] Handlers (`app/handlers/*`) не изменены.
- [ ] `FakeTorrentProvider` не изменён.

---

### Task 7: PostgresRepository

**Files:**
- Create: `app/repositories/postgres.py`

**Interfaces:**
- Consumes: Task 3 (schema), Task 5 (`AsyncConnectionPool`), Task 4 (protocol).
- Produces: `PostgresRepository(pool: AsyncConnectionPool)` — реализует `TorrentRepository`; используется Task 8/9. Row-mapping строк БД ↔ `TorrentTask` живёт здесь же.

- [ ] **Step 1: Создать `app/repositories/postgres.py`**

```python
from __future__ import annotations

from typing import Any, Mapping

from psycopg import Cursor
from psycopg.rows import dict_row
from psycopg_pool import AsyncConnectionPool

from app.models.media import MediaType, Quality
from app.models.torrent import TorrentStatus, TorrentTask

_TASK_COLUMNS = """
    t.id, t.magnet, t.info_hash, t.name, t.media_type, t.quality, t.status,
    t.progress, t.size_bytes, t.download_speed, t.upload_speed, t.eta_seconds,
    t.created_at, t.completed_at, u.telegram_user_id
"""


def _task_from_row(row: Mapping[str, Any]) -> TorrentTask:
    return TorrentTask(
        id=row["id"],
        user_id=row["telegram_user_id"],
        magnet=row["magnet"],
        info_hash=row["info_hash"],
        name=row["name"],
        media_type=MediaType(row["media_type"]),
        quality=Quality(row["quality"]),
        status=TorrentStatus(row["status"]),
        progress=float(row["progress"]),
        size_bytes=row["size_bytes"],
        download_speed=int(row["download_speed"]),
        upload_speed=int(row["upload_speed"]),
        eta_seconds=row["eta_seconds"],
        created_at=row["created_at"],
        completed_at=row["completed_at"],
        # колонки нет по design (§Components) — session state заполняет TorrentService
        provider_id="",
    )


class PostgresRepository:
    def __init__(self, pool: AsyncConnectionPool) -> None:
        self._pool = pool

    @staticmethod
    async def _ensure_user(cur: Cursor[dict[str, Any]], telegram_user_id: int) -> int:
        await cur.execute(
            """
            INSERT INTO users (telegram_user_id) VALUES (%s)
            ON CONFLICT (telegram_user_id)
            DO UPDATE SET username = COALESCE(users.username, EXCLUDED.username)
            RETURNING id
            """,
            (telegram_user_id,),
        )
        row = await cur.fetchone()
        assert row is not None
        return int(row["id"])

    async def find_by_hash(self, user_id: int, info_hash: str) -> TorrentTask | None:
        async with self._pool.connection() as conn:
            async with conn.cursor(row_factory=dict_row) as cur:
                await cur.execute(
                    f"SELECT {_TASK_COLUMNS} FROM torrents t "
                    "JOIN users u ON u.id = t.user_id "
                    "WHERE u.telegram_user_id = %s AND t.info_hash = %s",
                    (user_id, info_hash),
                )
                row = await cur.fetchone()
        return _task_from_row(row) if row is not None else None

    async def insert_task(self, task: TorrentTask) -> TorrentTask | None:
        async with self._pool.connection() as conn:
            async with conn.cursor(row_factory=dict_row) as cur:
                user_pk = await self._ensure_user(cur, task.user_id)
                await cur.execute(
                    """
                    INSERT INTO torrents (
                        user_id, magnet, info_hash, name, media_type, quality,
                        status, progress, size_bytes, download_speed, upload_speed,
                        eta_seconds, created_at, completed_at
                    ) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
                    ON CONFLICT (user_id, info_hash) DO NOTHING
                    RETURNING id
                    """,
                    (
                        user_pk,
                        task.magnet,
                        task.info_hash,
                        task.name,
                        str(task.media_type),
                        str(task.quality),
                        str(task.status),
                        task.progress,
                        task.size_bytes,
                        task.download_speed,
                        task.upload_speed,
                        task.eta_seconds,
                        task.created_at,
                        task.completed_at,
                    ),
                )
                inserted = await cur.fetchone()
                if inserted is None:
                    return None
                task.id = int(inserted["id"])
        return await self.get_task(task.user_id, task.id)

    async def get_task(self, user_id: int, task_id: int) -> TorrentTask | None:
        async with self._pool.connection() as conn:
            async with conn.cursor(row_factory=dict_row) as cur:
                await cur.execute(
                    f"SELECT {_TASK_COLUMNS} FROM torrents t "
                    "JOIN users u ON u.id = t.user_id "
                    "WHERE u.telegram_user_id = %s AND t.id = %s",
                    (user_id, task_id),
                )
                row = await cur.fetchone()
        return _task_from_row(row) if row is not None else None

    async def list_tasks(self, user_id: int) -> list[TorrentTask]:
        async with self._pool.connection() as conn:
            async with conn.cursor(row_factory=dict_row) as cur:
                await cur.execute(
                    f"SELECT {_TASK_COLUMNS} FROM torrents t "
                    "JOIN users u ON u.id = t.user_id "
                    "WHERE u.telegram_user_id = %s ORDER BY t.id",
                    (user_id,),
                )
                rows = await cur.fetchall()
        return [_task_from_row(row) for row in rows]

    async def update_task(self, task: TorrentTask) -> None:
        async with self._pool.connection() as conn:
            async with conn.cursor() as cur:
                await cur.execute(
                    """
                    UPDATE torrents SET
                        magnet = %s, name = %s, status = %s, progress = %s,
                        size_bytes = %s, download_speed = %s, upload_speed = %s,
                        eta_seconds = %s, completed_at = %s
                    WHERE id = %s
                    """,
                    (
                        task.magnet,
                        task.name,
                        str(task.status),
                        task.progress,
                        task.size_bytes,
                        task.download_speed,
                        task.upload_speed,
                        task.eta_seconds,
                        task.completed_at,
                        task.id,
                    ),
                )

    async def delete_task(self, task_id: int) -> None:
        async with self._pool.connection() as conn:
            async with conn.cursor() as cur:
                await cur.execute("DELETE FROM torrents WHERE id = %s", (task_id,))

    async def add_event(self, torrent_id: int, event: str, payload: str | None = None) -> None:
        async with self._pool.connection() as conn:
            async with conn.cursor() as cur:
                await cur.execute(
                    "INSERT INTO torrent_events (torrent_id, event, payload) VALUES (%s, %s, %s)",
                    (torrent_id, event, payload),
                )

    async def get_setting(self, user_id: int, key: str) -> str | None:
        async with self._pool.connection() as conn:
            async with conn.cursor() as cur:
                await cur.execute(
                    "SELECT s.value FROM settings s "
                    "JOIN users u ON u.id = s.user_id "
                    "WHERE u.telegram_user_id = %s AND s.key = %s",
                    (user_id, key),
                )
                row = await cur.fetchone()
        return row[0] if row is not None else None

    async def set_setting(self, user_id: int, key: str, value: str) -> None:
        async with self._pool.connection() as conn:
            async with conn.cursor() as cur:
                user_pk = await self._ensure_user(cur, user_id)
                await cur.execute(
                    "INSERT INTO settings (user_id, key, value) VALUES (%s, %s, %s) "
                    "ON CONFLICT (user_id, key) DO UPDATE SET value = EXCLUDED.value",
                    (user_pk, key, value),
                )
```

Замечания:
- Enum'ы приводятся к строке явно (`str(task.status)` — StrEnum возвращает value) — детерминированная адаптация psycopg.
- Ownership-фильтр — в SQL (`WHERE u.telegram_user_id = %s`), join `users` даёт telegram id обратно (внутренний `users.id` — surrogate, наружу не протекает).
- Ошибки запросов НЕ глотаются: они всплывают в error-handler диспетчера (`app/bot.py::on_error` — сообщение пользователю, бот не падает; design §Error Handling).
- Если mypy жалуется на `psycopg` (нет `py.typed`/типов) — добавить в `pyproject.toml` override `[[tool.mypy.overrides]] module = "psycopg.*" ignore_missing_imports = true` (сначала убедиться, что реально нужен).

- [ ] **Step 2: Smoke-проверка импорта + линтеры**

```bash
python -c "from app.repositories.postgres import PostgresRepository; print('ok')"
ruff check . && ruff format . && mypy app && pytest -q
```

Ожидаемо: `ok`, всё зелёное.

- [ ] **Step 3: Commit**

```bash
git add app/repositories/postgres.py
git commit -m "feat: add postgres repository"
```

**Acceptance criteria:**
- [ ] `PostgresRepository` реализует все 9 методов протокола (mypy проверяет структурную совместимость — передать его в `TorrentService(FakeTorrentProvider(), PostgresRepository(pool))` типизированно должно компилироваться; можно временно проверить одной строкой в python-консоли с `typing.runtime_checkable` — не обязательно).
- [ ] `mypy app` strict чистый; прежние тесты зелёные (on InMemory).

---

### Task 8: DI wiring в main.py (lifespan, fail-fast)

**Files:**
- Modify: `app/main.py`
- Review: `app/bot.py` (изменений не требуется — `TorrentService` уже инжектится в `Dispatcher(torrent_service=...)`; repository живёт внутри сервиса)

**Interfaces:**
- Consumes: `load_config()` (Task 2), `create_pool/apply_schema/redact_database_url` (Task 5), `PostgresRepository` (Task 7), `TorrentService` (Task 6).
- Produces: рабочий entrypoint `python -m app.main`.

- [ ] **Step 1: Переписать `app/main.py`**

```python
from __future__ import annotations

import asyncio
import logging

from aiogram import Bot, Dispatcher

from app.bot import create_dispatcher
from app.config import load_config
from app.db.pool import apply_schema, create_pool, redact_database_url
from app.providers.fake import FakeTorrentProvider
from app.repositories.postgres import PostgresRepository
from app.services.torrent_service import TorrentService

logger = logging.getLogger(__name__)


async def main() -> None:
    logging.basicConfig(
        level=logging.INFO,
        format="%(asctime)s %(levelname)s %(name)s: %(message)s",
    )
    config = load_config()
    # fail fast: нет DATABASE_URL -> RuntimeError из load_config; БД недоступна ->
    # RuntimeError из create_pool (redacted URL). Никакого тихого memory-режима (design §Error Handling)
    pool = await create_pool(config.database_url)
    try:
        await apply_schema(pool)
        repository = PostgresRepository(pool)
        service = TorrentService(FakeTorrentProvider(), repository)
        dp: Dispatcher = create_dispatcher(service, config.test_mode)
        bot = Bot(token=config.bot_token)
        logger.info(
            "bot started (TEST_MODE=%s, db=%s)",
            config.test_mode,
            redact_database_url(config.database_url),
        )
        await dp.start_polling(bot)
    finally:
        await pool.close()


if __name__ == "__main__":
    asyncio.run(main())
```

- [ ] **Step 2: Проверки**

```bash
ruff check . && ruff format . && mypy app && pytest -q
```

Ожидаемо: всё зелёное.

Fail-fast вручную (ожидаемо RuntimeError с сообщением про DATABASE_URL, без трейсбека в БД):

```bash
env -u DATABASE_URL BOT_TOKEN=123:TEST python -m app.main
```

Ожидаемо: `RuntimeError: DATABASE_URL is not set...`, код выхода ≠ 0.

Недоступная БД (за < ~15 c, сообщение с `***` вместо пароля):

```bash
BOT_TOKEN=123:TEST DATABASE_URL='postgresql://user:wrongpass@127.0.0.1:1/postgres?sslmode=disable' python -m app.main
```

Ожидаемо: `RuntimeError: Cannot reach PostgreSQL at postgresql://user:***@127.0.0.1:1/postgres...`, `wrongpass` нигде не появляется.

- [ ] **Step 3: Commit**

```bash
git add app/main.py
git commit -m "feat: wire postgres repository at startup with fail-fast"
```

**Acceptance criteria:**
- [ ] Без `DATABASE_URL` / при недоступной БД процесс завершается сразу с понятным redacted-сообщением.
- [ ] В логе старта виден `db=postgresql://user:***@host/...` (без пароля).
- [ ] `pytest` зелёный.

---

### Task 9: tests/integration/test_db.py

**Files:**
- Create: `tests/integration/test_db.py`

**Interfaces:**
- Consumes: Task 3 (schema), Task 5 (`create_pool`, `apply_schema`), Task 7 (`PostgresRepository`).
- Produces: интеграционные проверки контракта против реального PostgreSQL; авто-skip без БД.

Правила безопасности: тесты чистят **только свои строки** (тестовые telegram id в диапазоне `990000001..990000099` — реальные Telegram id в это диапазон не попадают). TRUNCATE всей БД ЗАПРЕЩЁН (можно убить реальные данные пользователя). Никогда не печатать `DATABASE_URL` (skip-сообщения идут через redacted-сообщения `create_pool`).

- [ ] **Step 1: Создать `tests/integration/test_db.py`**

```python
from __future__ import annotations

import os
from collections.abc import AsyncGenerator

import pytest
from dotenv import load_dotenv
from psycopg_pool import AsyncConnectionPool

from app.db.pool import apply_schema, create_pool
from app.models.media import MediaType, Quality
from app.models.torrent import TorrentStatus, TorrentTask
from app.repositories.postgres import PostgresRepository

load_dotenv()
DATABASE_URL = os.getenv("DATABASE_URL", "").strip()

# NB: module-level pytestmark НЕ использовать — он skip'нул бы и
# test_unreachable_database_fails_fast. Skip живёт в fixture pool:
# тесты, зависящие от pool/repo, skip'аются; fail-fast тест бегает всегда.

# тестовые telegram id (см. header файла): чистятся каскадом через users
_TEST_USER_LO = 990000001
_TEST_USER_HI = 990000099


def _task(user_id: int, info_hash: str) -> TorrentTask:
    return TorrentTask(
        id=0,
        user_id=user_id,
        magnet=f"magnet:?xt=urn:btih:{info_hash}",
        info_hash=info_hash,
        name="DB Test",
        media_type=MediaType.MOVIE,
        quality=Quality.FHD_1080,
    )


@pytest.fixture(scope="session")
async def pool() -> AsyncGenerator[AsyncConnectionPool, None]:
    if not DATABASE_URL:
        pytest.skip("DATABASE_URL is not set")
    try:
        p = await create_pool(DATABASE_URL)
    except RuntimeError as exc:
        pytest.skip(f"database unreachable: {exc}")  # exc уже с redacted URL
    try:
        await apply_schema(p)
        yield p
    finally:
        await p.close()


@pytest.fixture
async def repo(pool: AsyncConnectionPool) -> AsyncGenerator[PostgresRepository, None]:
    async with pool.connection() as conn:
        await conn.execute(
            "DELETE FROM users WHERE telegram_user_id BETWEEN %s AND %s",
            (_TEST_USER_LO, _TEST_USER_HI),
        )  # каскадом уберёт torrents/events/settings этих пользователей
    yield PostgresRepository(pool)
    async with pool.connection() as conn:
        await conn.execute(
            "DELETE FROM users WHERE telegram_user_id BETWEEN %s AND %s",
            (_TEST_USER_LO, _TEST_USER_HI),
        )


async def test_unreachable_database_fails_fast() -> None:
    # НЕ skip: порт 1 на localhost отказывает мгновенно; нужен только psycopg
    with pytest.raises(RuntimeError, match="Cannot reach PostgreSQL"):
        await create_pool("postgresql://nobody@127.0.0.1:1/postgres?sslmode=disable")


async def test_schema_applied_idempotent(pool: AsyncConnectionPool) -> None:
    await apply_schema(pool)  # повторное применение — не падает (CREATE IF NOT EXISTS)
    async with pool.connection() as conn:
        cur = await conn.execute(
            "SELECT tablename FROM pg_tables WHERE schemaname = 'public'"
        )
        rows = await cur.fetchall()
    names = {row[0] for row in rows}
    assert {"users", "torrents", "torrent_events", "settings"} <= names
    assert "media_items" not in names


async def test_task_crud_roundtrip(repo: PostgresRepository) -> None:
    user = _TEST_USER_LO
    inserted = await repo.insert_task(_task(user, "c" * 40))
    assert inserted is not None and inserted.id > 0

    fetched = await repo.get_task(user, inserted.id)
    assert fetched is not None
    assert fetched.user_id == user
    assert fetched.name == "DB Test"
    assert fetched.status is TorrentStatus.QUEUED
    assert fetched.media_type is MediaType.MOVIE
    assert fetched.quality is Quality.FHD_1080
    assert fetched.progress == 0.0
    assert fetched.created_at is not None

    fetched.status = TorrentStatus.DOWNLOADING
    fetched.progress = 42.5
    fetched.completed_at = None
    await repo.update_task(fetched)

    again = await repo.get_task(user, inserted.id)
    assert again is not None
    assert again.status is TorrentStatus.DOWNLOADING
    assert again.progress == 42.5

    await repo.delete_task(inserted.id)
    assert await repo.get_task(user, inserted.id) is None


async def test_duplicate_user_hash_rejected(repo: PostgresRepository) -> None:
    first = await repo.insert_task(_task(_TEST_USER_LO + 1, "d" * 40))
    assert first is not None

    duplicate = await repo.insert_task(_task(_TEST_USER_LO + 1, "d" * 40))
    assert duplicate is None  # UNIQUE (user_id, info_hash)

    other_user = await repo.insert_task(_task(_TEST_USER_HI, "d" * 40))
    assert other_user is not None  # другой пользователь — тот же hash разрешён


async def test_event_inserted(repo: PostgresRepository, pool: AsyncConnectionPool) -> None:
    task = await repo.insert_task(_task(_TEST_USER_LO + 2, "e" * 40))
    assert task is not None
    await repo.add_event(task.id, "status_changed", '{"from": "queued", "to": "paused"}')
    async with pool.connection() as conn:
        cur = await conn.execute(
            "SELECT event, payload FROM torrent_events WHERE torrent_id = %s", (task.id,)
        )
        rows = await cur.fetchall()
    assert rows == [("status_changed", '{"from": "queued", "to": "paused"}')]


async def test_ownership_filter(repo: PostgresRepository) -> None:
    task = await repo.insert_task(_task(_TEST_USER_LO + 3, "f" * 40))
    assert task is not None
    assert await repo.get_task(_TEST_USER_HI, task.id) is None  # чужой = «не найдено»
    assert await repo.list_tasks(_TEST_USER_HI) == []
    assert len(await repo.list_tasks(_TEST_USER_LO + 3)) == 1


async def test_settings_roundtrip(repo: PostgresRepository) -> None:
    user = _TEST_USER_LO + 4
    assert await repo.get_setting(user, "notify") is None
    await repo.set_setting(user, "notify", "on")
    assert await repo.get_setting(user, "notify") == "on"
    await repo.set_setting(user, "notify", "off")
    assert await repo.get_setting(user, "notify") == "off"
```

- [ ] **Step 2: Запустить без БД — убедиться в graceful skip**

```bash
env -u DATABASE_URL pytest tests/integration/test_db.py -v
```

Ожидаемо: `test_unreachable_database_fails_fast` PASSED; остальные SKIPPED (`DATABASE_URL is not set`); общий `pytest` — exit code 0.

- [ ] **Step 3: Запустить с реальным `DATABASE_URL`** (строка уже в `.env` после Task 2)

```bash
pytest tests/integration/test_db.py -v
```

Ожидаемо: все passed (включая идемпотентность схемы и `UNIQUE`).

- [ ] **Step 4: Полный прогон + commit**

```bash
pytest -q
git add tests/integration/test_db.py
git commit -m "test: add database integration tests"
```

**Acceptance criteria:**
- [ ] Без `DATABASE_URL`: 1 passed (fail-fast) + skip'ы, `pytest` exit 0.
- [ ] С `DATABASE_URL`: все тесты файла passed.
- [ ] Повторный `apply_schema` не падает; `media_items` отсутствует; дубль `(user_id, info_hash)` → `None`, другой юзер → вставлен; события пишутся; ownership фильтрует.
- [ ] В выводе тестов/логах нигде нет пароля из `DATABASE_URL`.

---

### Task 10: Docs + manual persistence check + финальная верификация

**Files:**
- Modify: `docs/ARCHITECTURE.md` (§8)
- Modify: `docs/features/010-persistent-state.md` (статус и scope)
- Modify: `docs/TESTING.md` (структура, раздел DB-тестов, ручной persistence-чек)

**Interfaces:**
- Consumes: Tasks 1–9 (весь реализованный функционал).
- Produces: обновлённая документация, зелёный DoD, feature = Implemented.

- [ ] **Step 1: `docs/ARCHITECTURE.md` §8** — заменить секцию «Database (будущее, FEAT-010)» на:

```markdown
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
```

- [ ] **Step 2: `docs/features/010-persistent-state.md`**:
  - `Status` → `Implemented` (после Step 4, когда DoD зелёный — поставить в конце задачи).
  - `Scope` поправить по design: `media_items` → «отложено в FEAT-013»; «Persistent FSM storage» → «NOT в Phase 1: FSM остаётся MemoryStorage (design §Approach)».
  - `Implementation Tasks` — отметить сделанное, убрать `media_items`/`Persistent FSM` из чек-листа (или пометить «отложено»), добавить: схема, repository layer (protocol/postgres/memory), pool + fail-fast, интеграционные DB-тесты.
  - `Acceptance Criteria` — отметить `[x]`: «Состояние переживает рестарт» (ручной чек, Step 5), «Telegram UI не переписан», «History хранится в database».

- [ ] **Step 3: `docs/TESTING.md`**:

1. Обновить дерево структуры:

```text
tests/
├── unit/
│   ├── test_magnet_parser.py
│   ├── test_fake_provider.py
│   ├── test_torrent_service.py
│   ├── test_config.py
│   └── test_db_pool.py
├── integration/
│   ├── test_bot_flow.py
│   └── test_db.py
└── fixtures/
```

2. Новый раздел после «Integration tests»:

```markdown
## Database tests (FEAT-010)

`tests/integration/test_db.py` — против реального PostgreSQL по `DATABASE_URL` из `.env`.

- Без `DATABASE_URL` (или при недоступной БД) тесты автоматически skip'аются —
  `pytest` остаётся зелёным в CI. Fail-fast проверка (`test_unreachable_database_fails_fast`)
  бегает всегда.
- Тесты чистят только свои строки (тестовые telegram id 990000001..990000099,
  каскадный DELETE через `users`). TRUNCATE всей БД не используется —
  не запускайте DB-тесты против production-БД с важными данными; лучше выделить
  отдельную БД (например torrent_bot).
- Пароль из DATABASE_URL не попадает в вывод тестов (redaction в app/db/pool.py).
```

3. Обновить «Подготовку» ручного тестирования: в п.2 (`.env`) добавить заполнение
   `DATABASE_URL` (формат из `.env.example`), в п.4 — ожидание лога
   `bot started (TEST_MODE=True, db=postgresql://user:***@...)`.
   В п.20 (остановка) — упомянуть, что без `DATABASE_URL` бот не стартует (fail-fast).

4. Новый ручной шаг — **проверка персистентности** (после п.19, перед «Критерий успеха»):

```markdown
21. **Персистентность (FEAT-010).** Создайте любую задачу (пп. 3–8),
    убедитесь, что `/downloads` её показывает. Остановите бота `Ctrl+C`,
    запустите снова `python -m app.main`. Отправьте `/downloads`.
    Ожидаемо: задача всё ещё в списке с тем же номером `#001` и названием
    (статус после рестарта может временно вернуться к «🕓 В очереди» —
    FakeTorrentProvider живёт в памяти процесса, прогресс имитируется заново).
    Задачи, помеченные завершёнными/отменёнными/удалёнными до рестарта,
    сохраняют статус. Дубликат (пп. 9) после рестарта тоже срабатывает —
    UNIQUE(user_id, info_hash) теперь в БД.
```

В «Критерий успеха» добавить: «Шаг 21 пройден; в логе старта виден redacted URL БД; без DATABASE_URL бот падает с понятной ошибкой».

- [ ] **Step 4: Финальная верификация (все четыре команды обязательны, DEVELOPMENT §4)**

```bash
ruff check .
ruff format --check .
mypy app
pytest
```

Ожидаемо: без ошибок; `pytest` — все passed, DB-тесты либо passed (БД доступна), либо честно skipped (БД недоступна), exit code 0.

Если что-то красное — чинить и перезапускать все четыре до зелёного. Никогда не «чинить» удалением тестов.

- [ ] **Step 5: Ручной persistence-чек** (см. новый шаг 21 TESTING.md): бот стартует → задача → рестарт → `/downloads` показывает задачу. Если `DATABASE_URL` не заполнен — запросить у пользователя.

- [ ] **Step 6: Feature status + commit**

`docs/features/010-persistent-state.md` → `Status: Implemented` (если Step 4 зелёный; иначе `In Progress` и не коммитить как готовое).

```bash
git add docs/ARCHITECTURE.md docs/features/010-persistent-state.md docs/TESTING.md
git commit -m "docs: document FEAT-010 persistent state"
```

**Acceptance criteria:**
- [ ] `ruff check .` / `ruff format --check .` / `mypy app` / `pytest` — все зелёные.
- [ ] `docs/ARCHITECTURE.md` §8 описывает реализованное (не «будущее»).
- [ ] Feature 010: Status `Implemented`, критерии отмечены.
- [ ] `docs/TESTING.md` содержит раздел DB-тестов и ручной persistence-шаг.
- [ ] Ручной чек «рестарт → /downloads показывает задачу» пройден.

---

## Self-Review Notes (для ревьюера плана)

- **Spec coverage:** design §Components (4 таблицы, repository-троица, config, lifespan, schema.sql), §Data Flow (upsert user → ON CONFLICT → duplicate → event), §Error Handling (fail-fast, redaction, error-handler бота), §Testing Strategy (20 старых зелёные + DB-тесты со skip), §Documentation & Git (Task 0, Task 10, `.env.example`) — покрыты Task 1–10.
- **Явные отклонения от design-списка «правки»:** `app/bot.py` не меняется (DI уже есть через Dispatcher kwargs); row-mapping живёт в `app/repositories/postgres.py`, а не в `app/models/torrent.py` (StrEnum-поля и так round-trip'ятся, отдельный слой в models был бы мёртвым кодом).
- **provider_id:** колонки нет (design), но «состояние переживает рестарт» требует, чтобы задачи не падали в FAILED — решено session-map'ой в сервисе (Task 6), FakeTorrentProvider не тронут.
- **Type consistency:** сигнатуры `TorrentRepository` (Task 4) = реализация (Task 7) = вызовы в сервисе (Task 6) = фикстуры тестов (Task 6/9).
