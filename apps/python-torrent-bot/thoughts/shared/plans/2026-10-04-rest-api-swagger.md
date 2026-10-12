# REST API + Swagger UI (Phase 1.5) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Добавить публичный локальный REST API (FastAPI + Swagger UI) поверх существующего `TorrentService`, параллельно с Telegram-polling, без переписывания бизнес-слоя.

**Architecture:** Новый слой `app/api/` (`schemas.py` + `routes.py` + `app.py`) — третий клиент `TorrentService` наряду с handlers. FastAPI-приложение собирается фабрикой `create_api_app(service, repository, api_user_id)`; в `app/main.py` uvicorn-сервер и `dp.start_polling` идут рядом через `asyncio.gather(..., return_exceptions=True)`. Актор API — `API_USER_ID` из config; ownership-check сервиса даёт ровно telegram-поведение. Состояние в той же PostgreSQL — задачи общие с ботом.

**Tech Stack:** Python 3.12+, `fastapi>=0.115,<1`, `uvicorn>=0.54,<1` (runtime), `httpx>=0.27` (dev, нужен TestClient), Pydantic v2, pytest/pytest-asyncio (`asyncio_mode=auto`), ruff, mypy (strict, `files = ["app"]`).

**Spec:**
- `thoughts/shared/designs/2026-10-04-rest-api-swagger-design.md` (authoritative design)
- `init.md` §19 (шаблон feature-документа), §25 (ADR-шаблон)
- `docs/DEVELOPMENT.md` §7 (Conventional Commits), §9 (Definition of Done)

## Global Constraints

- DoD: `ruff check .`, `ruff format --check .`, `mypy app`, `pytest` — все четыре зелёные (финальная проверка, Task 9).
- Существующие 36 тестов остаются зелёными и не удаляются (DEVELOPMENT §3.8).
- `TorrentService` НЕ переписывается; Telegram-слой НЕ переписывается. **Единственное разрешённое расширение** — Task 4: метод `TorrentService.register_user` + одна строка в `cmd_start` (без этого инструкция «нажми /start» из design §Error Handling была бы нерабочей — см. Technical Findings; `create_dispatcher` и тесты бота при этом не меняются).
- `FakeTorrentProvider` не изменяется.
- API слушает **только localhost** (`127.0.0.1`/`localhost`/`::1`), без аутентификации — валидация в config fail-fast (design §Constraints).
- Секреты (`DATABASE_URL`, `BOT_TOKEN`) не попадают в ответы API и логи; `log_config=None` у uvicorn — не перебивать наш `logging.basicConfig`.
- `TEST_MODE=true` обязателен (config проверяет); реальных загрузок нет.
- ruff: line-length 100, rules E,F,I,UP,B,SIM; код в `.md` не форматируется (`extend-exclude`).
- Git: Conventional Commits; `.env`/секреты никогда не коммитить.
- Новые зависимости синхронизировать в обоих файлах: `pyproject.toml` `[project].dependencies` и `requirements.txt` (runtime ↔ dev).

## Task Dependency / Parallel Map

```text
Wave A (параллельно): Task 1 (deps) | Task 2 (config + .env.example + CONFIGURATION.md + config-тесты)
                      | Task 3 (repository: ensure_user/user_exists)
Wave B (параллельно): Task 4 (register_user + /start, ждёт 3) | Task 5 (api/schemas.py)
Wave C:               Task 6 (api/routes.py + api/app.py, ждёт 4+5)
Wave D (параллельно): Task 7 (main.py wiring, ждёт 6) | Task 8 (tests/integration/test_api.py, ждёт 6)
Final:                Task 9 (docs + финальная верификация, ждёт всё)
```

## Technical Findings (из чтения кода — обязательны к пониманию)

1. **Сервис не сигнализирует неверный статус исключением.** `TorrentService.pause/resume` при неверном статусе возвращают задачу **без изменений** (`app/services/torrent_service.py:110-114, 127-128`) — по возвращаемому значению это НЕ отличить от успеха (`pause` успешный тоже возвращает статус `paused`). Поэтому роуты обязаны **предпроверять** статус через `service.get()` ДО вызова мутации и отдавать 409. `cancel` вообще не отклоняет неверный статус (всегда ставит `CANCELLED`) — 409 «уже отменена» добавляется на уровне роута. `delete` возвращает `False` и для «не найдено», и для сбоя provider — маппится в 404 (accepted simplification, `ponytail:`).
2. **Signal-хендлеры: конфликт реален, но разрешается штатно.** aiogram 3.31 (`start_polling(handle_signals=True)` по умолчанию) ставит `loop.add_signal_handler(SIGINT/SIGTERM)`; uvicorn 0.54 в `Server.serve()` делает `capture_signals()` через `signal.signal` (перетирает C-уровневый хендлер aiogram), а при выходе **восстанавливает прежние хендлеры и пересылает сигнал** (`signal.raise_signal`) — сигнал доходит до aiogram'а и polling останавливается штатно. Guard `config.install_signal_handlers = False` в uvicorn 0.54 **не существует** (проверено по исходникам 0.54.0) — вместо него: (а) в `asyncio.gather` первым аргументом идёт `dp.start_polling` (aiogram ставит хендлеры раньше uvicorn — порядок создания задач в gather детерминирован), (б) `server.should_exit = True` в `finally` polling-обёртки, (в) ручная проверка Ctrl+C в Task 9.
3. **`/start` НЕ создаёт строку в `users`.** `_ensure_user` вызывается только из `insert_task`/`set_setting` PostgresRepository; `cmd_start` трогает лишь FSM. Без правки (Task 4) инструкция «нажми /start в боте» из design была бы нерабочей петлёй (400 → /start → снова 400). Решение: `ensure_user` + `register_user` + строка в `cmd_start` — минимальное расширение, `create_dispatcher(service, test_mode)` и `tests/integration/test_bot_flow.py` не меняются (aiogram берёт `torrent_service` из dispatcher data).
4. **Строки вместо Enum в JSON.** `MediaType`/`Quality`/`TorrentStatus` — `StrEnum`: Pydantic сериализует их значения как plain JSON-строки (design §Components «строки вместо Enum» выполняется), при этом в OpenAPI появляются enum-подсказки, а input-валидация даёт бесплатный 422. Входные `media_type`/`quality` валидируются enum'ом; неверное значение → 422 стандартным FastAPI-обработчиком.
5. **pytest.** `asyncio_mode = "auto"` — API-тесты пишутся `async def` (фикстуры тоже async из-за сидирования `ensure_user`); TestClient из fastapi синхронный и вызывается внутри async-теста без проблем (портал в отдельном потоке). Никакого `pytest-asyncio`-конфига менять не нужно.
6. **mypy strict.** `uvicorn` и `fastapi` имеют `py.typed` (проверено: uvicorn 0.54 wheel содержит `py.typed`) — override для mypy не нужен. `uvicorn.Config(app, host=..., port=..., log_config=None)` — `log_config=None` отключает dictConfig, наша `basicConfig`+файловый хендлер сохраняются (проверено по исходникам config.py 0.54.0).

---

### Task 1: fastapi/uvicorn/httpx dependencies

**Files:**
- Modify: `pyproject.toml` (`[project].dependencies`)
- Modify: `requirements.txt`

**Interfaces:**
- Consumes: —
- Produces: установленные `fastapi`, `uvicorn`, `httpx` (импортируются в Task 5–8).

- [ ] **Step 1: Добавить runtime-зависимости в `pyproject.toml`**

```toml
dependencies = [
    "aiogram>=3.25,<4",
    "python-dotenv>=1.0",
    "psycopg[binary,pool]>=3.3,<4",
    "fastapi>=0.115,<1",
    "uvicorn>=0.54,<1",
]
```

`uvicorn>=0.54` — нижняя граница по версии, чьё поведение `capture_signals` (restore + re-raise) проверено по исходникам (см. Technical Findings §2). Базовый uvicorn без `[standard]`: uvloop/httptools не нужны (Ponytail: YAGNI).

- [ ] **Step 2: Добавить dev-зависимость в `requirements.txt`** (файл держит runtime и dev, комментарии «синхронизировать» уже есть)

```text
# runtime — синхронизировать с [project].dependencies в pyproject.toml
aiogram>=3.25,<4
python-dotenv>=1.0
psycopg[binary,pool]>=3.3,<4
fastapi>=0.115,<1
uvicorn>=0.54,<1

# dev (DEVELOPMENT §5)
pytest>=8
pytest-asyncio>=0.24
pytest-cov>=5
ruff>=0.6
mypy>=1.11
httpx>=0.27
```

- [ ] **Step 3: Установить и проверить импорты**

```bash
pip install -r requirements.txt
python -c "import fastapi, uvicorn, httpx; print(fastapi.__version__, uvicorn.__version__, httpx.__version__)"
```

Ожидаемо: версии выводятся без ошибок (fastapi 0.115+, uvicorn 0.54+, httpx 0.27+).

- [ ] **Step 4: Быстрая проверка, что ничего не сломали**

```bash
ruff check . && ruff format --check . && mypy app && pytest -q
```

Ожидаемо: 36 passed, линтеры/типы чистые.

- [ ] **Step 5: Commit**

```bash
git add pyproject.toml requirements.txt
git commit -m "build: add fastapi, uvicorn and httpx dependencies"
```

**Acceptance criteria:**
- [ ] `import fastapi, uvicorn, httpx` работает; все 36 существующих тестов зелёные.

---

### Task 2: config API_HOST/API_PORT/API_USER_ID + .env.example + CONFIGURATION.md + config-тесты (TDD)

**Files:**
- Modify: `app/config.py`
- Modify: `.env.example`
- Modify: `docs/CONFIGURATION.md`
- Test: `tests/unit/test_config.py` (4 новых теста)

**Interfaces:**
- Consumes: —
- Produces (использует Task 7): `Config.api_host: str` (только localhost), `Config.api_port: int` (1–65535), `Config.api_user_id: int` (> 0) — гарантированно провалидированы.

Стиль — как у существующих полей: frozen dataclass, `os.getenv` + `.strip()`, ошибки через `RuntimeError` с понятным сообщением (как `LOG_LEVEL`). Валидация localhost — security check: API без аутентификации, design §Constraints («слушает только localhost»).

- [ ] **Step 1: Написать failing-тесты в `tests/unit/test_config.py`** (добавить в конец файла, существующие не трогать)

```python
def test_api_defaults(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setattr("app.config.load_dotenv", lambda: None)
    monkeypatch.setenv("BOT_TOKEN", "123:TEST")
    monkeypatch.setenv("TEST_MODE", "true")
    monkeypatch.setenv("DATABASE_URL", "postgresql://user:pass@host:5432/db")
    monkeypatch.delenv("API_HOST", raising=False)
    monkeypatch.delenv("API_PORT", raising=False)
    monkeypatch.delenv("API_USER_ID", raising=False)
    config = load_config()
    assert config.api_host == "127.0.0.1"
    assert config.api_port == 8080
    assert config.api_user_id == 1


def test_api_explicit_values(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setattr("app.config.load_dotenv", lambda: None)
    monkeypatch.setenv("BOT_TOKEN", "123:TEST")
    monkeypatch.setenv("TEST_MODE", "true")
    monkeypatch.setenv("DATABASE_URL", "postgresql://user:pass@host:5432/db")
    monkeypatch.setenv("API_HOST", "localhost")
    monkeypatch.setenv("API_PORT", "9090")
    monkeypatch.setenv("API_USER_ID", "42")
    config = load_config()
    assert (config.api_host, config.api_port, config.api_user_id) == ("localhost", 9090, 42)


def test_invalid_api_port_raises(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setattr("app.config.load_dotenv", lambda: None)
    monkeypatch.setenv("BOT_TOKEN", "123:TEST")
    monkeypatch.setenv("TEST_MODE", "true")
    monkeypatch.setenv("DATABASE_URL", "postgresql://user:pass@host:5432/db")
    monkeypatch.setenv("API_PORT", "eighty")
    with pytest.raises(RuntimeError, match="API_PORT"):
        load_config()


def test_non_localhost_api_host_raises(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setattr("app.config.load_dotenv", lambda: None)
    monkeypatch.setenv("BOT_TOKEN", "123:TEST")
    monkeypatch.setenv("TEST_MODE", "true")
    monkeypatch.setenv("DATABASE_URL", "postgresql://user:pass@host:5432/db")
    monkeypatch.setenv("API_HOST", "0.0.0.0")
    with pytest.raises(RuntimeError, match="API_HOST"):
        load_config()
```

- [ ] **Step 2: Запустить — убедиться, что падает**

```bash
pytest tests/unit/test_config.py -v
```

Ожидаемо: 4 новых FAIL (`Config.__init__() missing ... argument` — `api_host` и т.д.), 5 прежних PASS.

- [ ] **Step 3: Изменить `app/config.py`**

Добавить поля в dataclass:

```python
@dataclass(frozen=True)
class Config:
    bot_token: str
    test_mode: bool
    database_url: str
    log_level: str
    api_host: str
    api_port: int
    api_user_id: int
```

Добавить в `load_config()` перед `return Config(...)` и расширить return:

```python
    api_host = (os.getenv("API_HOST") or "127.0.0.1").strip()
    if api_host not in {"127.0.0.1", "localhost", "::1"}:
        raise RuntimeError(
            f"API_HOST must be a localhost address — the HTTP API has no auth "
            f"(got {api_host!r})."
        )
    api_port_raw = (os.getenv("API_PORT") or "8080").strip()
    try:
        api_port = int(api_port_raw)
    except ValueError:
        raise RuntimeError(f"API_PORT must be an integer (got {api_port_raw!r}).") from None
    if not 1 <= api_port <= 65535:
        raise RuntimeError(f"API_PORT must be between 1 and 65535 (got {api_port}).")
    api_user_id_raw = (os.getenv("API_USER_ID") or "1").strip()
    try:
        api_user_id = int(api_user_id_raw)
    except ValueError:
        raise RuntimeError(f"API_USER_ID must be an integer (got {api_user_id_raw!r}).") from None
    if api_user_id <= 0:
        raise RuntimeError(f"API_USER_ID must be a positive Telegram user id (got {api_user_id}).")
    return Config(
        bot_token=token,
        test_mode=test_mode,
        database_url=database_url,
        log_level=log_level,
        api_host=api_host,
        api_port=api_port,
        api_user_id=api_user_id,
    )
```

- [ ] **Step 4: Запустить — зелёные**

```bash
pytest tests/unit/test_config.py -v
```

Ожидаемо: 9 passed.

- [ ] **Step 5: `.env.example`** — добавить три строки с комментарием:

```dotenv
BOT_TOKEN=123456:REPLACE_WITH_TOKEN_FROM_BOTFATHER
TEST_MODE=true
# Рекомендуется отдельная БД (например torrent_bot) вместо public — design §Documentation
DATABASE_URL=postgresql://postgres:password@192.168.1.35:5432/torrent_bot?sslmode=disable
LOG_LEVEL=INFO  # DEBUG|INFO|WARNING|ERROR|CRITICAL
# HTTP API (FEAT-020): только localhost, без аутентификации; API_USER_ID — telegram id владельца API
API_HOST=127.0.0.1
API_PORT=8080
API_USER_ID=1
```

- [ ] **Step 6: `docs/CONFIGURATION.md`** — в секцию «Текущая конфигурация (MVP)” добавить в блок env и абзац:

```text
API_HOST=127.0.0.1
API_PORT=8080
API_USER_ID=1
```

```markdown
`API_HOST`/`API_PORT`/`API_USER_ID` (FEAT-020) — адрес HTTP API (Swagger UI: `http://127.0.0.1:8080/docs`).
`API_HOST` принимает только localhost (`127.0.0.1`/`localhost`/`::1`) — API без аутентификации,
fail-fast при попытке слушать наружу. `API_USER_ID` — telegram id пользователя, **от имени которого
работает API**: create ставит его владельцем задачи, list/get/pause/resume/cancel/delete проверяют
ownership через него (поведение = Telegram). Несовпадение с вашим реальным telegram id = пустой
список задач. Перед первым `POST /api/tasks` пользователь должен нажать `/start` в боте
(регистрация в `users`).
```

- [ ] **Step 7: Прогон + commit**

```bash
ruff check . && ruff format . && mypy app && pytest -q
git add app/config.py .env.example docs/CONFIGURATION.md tests/unit/test_config.py
git commit -m "feat: add API_HOST/API_PORT/API_USER_ID config"
```

**Acceptance criteria:**
- [ ] `pytest tests/unit/test_config.py` — 9 passed; полный `pytest` — все зелёные.
- [ ] `API_HOST=0.0.0.0` → fail-fast с сообщением про localhost; `API_PORT=eighty` → RuntimeError.
- [ ] `.env` не попал в git (`git ls-files | grep -x ".env"` пусто).

---

### Task 3: repository — ensure_user / user_exists (protocol + memory + postgres)

**Files:**
- Modify: `app/repositories/__init__.py` (протокол, +2 метода)
- Modify: `app/repositories/memory.py` (InMemoryRepository, +2 метода, `_users`)
- Modify: `app/repositories/postgres.py` (PostgresRepository, +2 метода)

**Interfaces:**
- Consumes: —
- Produces (используют Task 4 и 6):

```python
class TorrentRepository(Protocol):
    ...
    async def ensure_user(self, telegram_user_id: int) -> None: ...
    async def user_exists(self, telegram_user_id: int) -> bool: ...
```

Семантика: `ensure_user` — upsert строки `users` (как существующий приватный `_ensure_user`); `user_exists` — read-only проверка существования (API create → 400, пока пользователь не зарегистрирован). Отдельного тест-файла для репозиториев НЕ пишем (YAGNI): контракт покрывается API-тестами Task 8 (400/флоу на InMemory) и существующим `tests/integration/test_db.py` для Postgres-пути (Postgres-методы — 3-строчные обёртки над уже покрытыми конструкциями).

- [ ] **Step 1: Добавить 2 метода в протокол `app/repositories/__init__.py`** (после `set_setting`)

```python
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
    async def ensure_user(self, telegram_user_id: int) -> None: ...
    async def user_exists(self, telegram_user_id: int) -> bool: ...
```

- [ ] **Step 2: `app/repositories/memory.py`** — в `__init__` добавить `self._users: set[int] = set()`; в `insert_task` (перед `return task`) и в `set_setting` добавить `self._users.add(task.user_id)` / `self._users.add(user_id)` (зеркалит `_ensure_user`); добавить методы:

```python
    async def ensure_user(self, telegram_user_id: int) -> None:
        self._users.add(telegram_user_id)

    async def user_exists(self, telegram_user_id: int) -> bool:
        return telegram_user_id in self._users
```

- [ ] **Step 3: `app/repositories/postgres.py`** — добавить методы (повторяют контекст-менеджмент соседних методов):

```python
    async def ensure_user(self, telegram_user_id: int) -> None:
        async with self._pool.connection() as conn, conn.cursor() as cur:
            await self._ensure_user(cur, telegram_user_id)

    async def user_exists(self, telegram_user_id: int) -> bool:
        async with self._pool.connection() as conn, conn.cursor() as cur:
            await cur.execute(
                "SELECT 1 FROM users WHERE telegram_user_id = %s", (telegram_user_id,)
            )
            return await cur.fetchone() is not None
```

- [ ] **Step 4: Проверка + commit**

```bash
ruff check . && ruff format . && mypy app && pytest -q
git add app/repositories/
git commit -m "feat: add user registration and lookup to repository"
```

Ожидаемо: всё зелёное (mypy видит, что обе реализации покрывают протокол — структурная совместимость проверяется в Task 7 при передаче `PostgresRepository` в `create_api_app`).

**Acceptance criteria:**
- [ ] `mypy app` strict чистый; `pytest` — прежние 36 зелёные.
- [ ] Протокол, memory и postgres содержат оба новых метода с одинаковой сигнатурой.

---

### Task 4: `TorrentService.register_user` + регистрация на /start

**Files:**
- Modify: `app/services/torrent_service.py` (+1 публичный метод)
- Modify: `app/handlers/start.py` (+1 параметр, +1 вызов)

**Interfaces:**
- Consumes: Task 3 (`TorrentRepository.ensure_user`).
- Produces (использует Task 6/8): `TorrentService.register_user(user_id: int) -> None` — создаёт строку `users` при `/start`, чтобы POST /api/tasks не отвечал 400 навсегда (см. Technical Findings §3). Это **минимальное разрешённое расширение** сервиса/Telegram-слоя (Global Constraints).

`create_dispatcher(service, test_mode)` НЕ меняется: aiogram берает `torrent_service` из dispatcher data; `tests/integration/test_bot_flow.py` не меняется (фид `/start` продолжает работать — `register_user` на InMemoryRepository просто пишет в set).

- [ ] **Step 1: Добавить метод в `app/services/torrent_service.py`** (после `__init__`, перед `create`)

```python
    async def register_user(self, user_id: int) -> None:
        """Регистрация при /start: строка users нужна API до первого create (FEAT-020)."""
        await self._repo.ensure_user(user_id)
```

- [ ] **Step 2: Изменить `app/handlers/start.py`**

```python
from __future__ import annotations

from aiogram import Router
from aiogram.filters import Command
from aiogram.fsm.context import FSMContext
from aiogram.types import Message

from app.keyboards.main import main_menu_keyboard
from app.services.torrent_service import TorrentService

router = Router(name="start")

START_TEXT = "🤖 Torrent Bot\n\nОтправьте magnet-ссылку, чтобы создать новую загрузку."


@router.message(Command("start"))
async def cmd_start(message: Message, state: FSMContext, torrent_service: TorrentService) -> None:
    user_id = message.from_user.id if message.from_user else 0
    await torrent_service.register_user(user_id)
    await state.clear()
    await message.answer(START_TEXT, reply_markup=main_menu_keyboard())
```

- [ ] **Step 3: Проверка**

```bash
ruff check . && ruff format . && mypy app && pytest -q
```

Ожидаемо: всё зелёное (36 passed — включая `test_full_flow_start_to_downloads`, который фидит `/start`: `register_user` выполняется на InMemoryRepository без ошибок).

- [ ] **Step 4: Commit**

```bash
git add app/services/torrent_service.py app/handlers/start.py
git commit -m "feat: register user on /start for HTTP API"
```

**Acceptance criteria:**
- [ ] Все тесты зелёные; `app/bot.py` и `tests/integration/test_bot_flow.py` не изменены.
- [ ] `TorrentService` не знает о Telegram (метод доменный, без импортов aiogram).

---

### Task 5: `app/api/schemas.py` — Pydantic-модели

**Files:**
- Create: `app/api/schemas.py` (пакет namespace — `__init__.py` НЕ нужен, как у `app/handlers`)

**Interfaces:**
- Consumes: `app.models.media.MediaType/Quality`, `app.models.torrent.TorrentStatus/TorrentTask`.
- Produces (использует Task 6/8): `TaskCreateIn`, `TorrentTaskOut`, `ErrorOut`.

Дизайн-решения:
- `TorrentTaskOut` — все поля доменной модели, кроме `provider_id` (session state, наружу не протекает). `from_attributes=True` → валидация прямо из dataclass `TorrentTask` (`model_validate(task)`), без ручного маппинга.
- StrEnum в модели → в JSON уходят plain-строки (design §Components «строки вместо Enum»), в OpenAPI — enum-подсказки (см. Technical Findings §4).
- `TaskCreateIn` использует те же enum'ы → неверное значение `media_type`/`quality` = 422 стандартным FastAPI-обработчиком.

- [ ] **Step 1: Создать `app/api/schemas.py`**

```python
from __future__ import annotations

from datetime import datetime

from pydantic import BaseModel, ConfigDict

from app.models.media import MediaType, Quality
from app.models.torrent import TorrentStatus


class TaskCreateIn(BaseModel):
    magnet: str
    media_type: MediaType
    quality: Quality


class TorrentTaskOut(BaseModel):
    model_config = ConfigDict(from_attributes=True)

    id: int
    user_id: int
    magnet: str
    info_hash: str
    name: str
    media_type: MediaType
    quality: Quality
    status: TorrentStatus
    progress: float
    size_bytes: int | None
    download_speed: int
    upload_speed: int
    eta_seconds: int | None
    created_at: datetime
    completed_at: datetime | None


class ErrorOut(BaseModel):
    detail: str
```

- [ ] **Step 2: Проверка + commit**

```bash
ruff check . && ruff format . && mypy app && python -c "from app.api.schemas import TorrentTaskOut; print('ok')"
git add app/api/schemas.py
git commit -m "feat: add API pydantic schemas"
```

Ожидаемо: `ok`, линтеры/типы чистые.

**Acceptance criteria:**
- [ ] `mypy app` strict чистый; схема экспортируется; ни один существующий тест не изменился.

---

### Task 6: `app/api/routes.py` + `app/api/app.py` — маршруты и фабрика приложения

**Files:**
- Create: `app/api/routes.py`
- Create: `app/api/app.py`

**Interfaces:**
- Consumes: Task 3 (`user_exists`), Task 4 (`service.register_user` — контекст), Task 5 (схемы), `TorrentService` (create/list/get/pause/resume/cancel/delete), `parse_magnet` (ValueError → 400).
- Produces (используют Task 7/8): `create_api_app(service: TorrentService, repository: TorrentRepository, api_user_id: int) -> FastAPI`; `create_router(...) -> APIRouter`.

Ключевые решения (все по design §Endpoints + Technical Findings §1):
- **409 на неверный статус — предпроверкой**: `service.get()` → None → 404; статус не из разрешённого набора → 409; иначе мутация. Разрешённые наборы: pause — `{queued, downloading}` (как в сервисе), resume — `{paused}` (как в сервисе), cancel — всё, кроме уже `cancelled` (сам сервис cancel не отклоняет никогда — 409 добавлен на уровне роута по design-таблице).
- **500 — `_guard()`**: единственная точка с `try/except Exception` → `logger.error(..., exc_info=True)` + `HTTPException(500, "internal error")`. Это реализация требования «route-level try/except» один раз вместо семи копий (DRY); `HTTPException`/`MagnetParseError` наружу не проходят — они поднимаются вне `_guard`.
- **409-дубль**: `detail` = JSON существующей задачи («409 with existing task in detail»).
- **400**: два источника — невалидный magnet (`parse_magnet` → `ValueError` → detail = `str(exc)`) и несуществующий `API_USER_ID` (detail = «пользователь не найден — нажми /start в боте»). Порядок: magnet → пользователь → create.
- **DELETE**: `service.delete` → `False` → 404 (False и для «не найдено», и для сбоя provider — accepted simplification); успех → 204 без тела.

- [ ] **Step 1: Создать `app/api/routes.py`**

```python
from __future__ import annotations

import logging
from collections.abc import Awaitable
from typing import TypeVar

from fastapi import APIRouter, HTTPException, status

from app.api.schemas import ErrorOut, TaskCreateIn, TorrentTaskOut
from app.models.torrent import TorrentStatus, TorrentTask
from app.repositories import TorrentRepository
from app.services.magnet_parser import parse_magnet
from app.services.torrent_service import TorrentService

logger = logging.getLogger(__name__)

T = TypeVar("T")

_PAUSEABLE = (TorrentStatus.QUEUED, TorrentStatus.DOWNLOADING)


async def _guard(op: str, coro: Awaitable[T]) -> T:
    """Неожиданная ошибка сервиса → log + 500 без раскрытия внутренностей (design §Error Handling)."""
    try:
        return await coro
    except Exception:
        logger.error("api %s failed", op, exc_info=True)
        raise HTTPException(status_code=500, detail="internal error") from None


def create_router(
    service: TorrentService, repository: TorrentRepository, api_user_id: int
) -> APIRouter:
    router = APIRouter(prefix="/api/tasks", tags=["tasks"])

    def _out(task: TorrentTask) -> TorrentTaskOut:
        return TorrentTaskOut.model_validate(task)

    async def _owned(task_id: int) -> TorrentTask:
        task = await _guard("get", service.get(api_user_id, task_id))
        if task is None:
            raise HTTPException(status_code=404, detail="task not found")
        return task

    async def _mutated(op: str, task: TorrentTask | None) -> TorrentTask:
        if task is None:
            raise HTTPException(status_code=404, detail="task not found")
        return _out(task)

    @router.get("", response_model=list[TorrentTaskOut])
    async def list_tasks() -> list[TorrentTaskOut]:
        tasks = await _guard("list", service.list(api_user_id))
        return [_out(task) for task in tasks]

    @router.get(
        "/{task_id}",
        response_model=TorrentTaskOut,
        responses={404: {"model": ErrorOut}},
    )
    async def get_task(task_id: int) -> TorrentTaskOut:
        return _out(await _owned(task_id))

    @router.post(
        "",
        status_code=status.HTTP_201_CREATED,
        response_model=TorrentTaskOut,
        responses={
            400: {"model": ErrorOut},
            409: {"description": "Duplicate; detail содержит существующую задачу"},
        },
    )
    async def create_task(payload: TaskCreateIn) -> TorrentTaskOut:
        try:
            magnet = parse_magnet(payload.magnet)
        except ValueError as exc:
            raise HTTPException(status_code=400, detail=str(exc)) from None
        if not await _guard("user_exists", repository.user_exists(api_user_id)):
            raise HTTPException(
                status_code=400, detail="пользователь не найден — нажми /start в боте"
            )
        task, created = await _guard(
            "create",
            service.create(
                user_id=api_user_id,
                magnet=magnet,
                media_type=payload.media_type,
                quality=payload.quality,
            ),
        )
        if not created:
            raise HTTPException(status_code=409, detail=_out(task).model_dump(mode="json"))
        return _out(task)

    @router.post(
        "/{task_id}/pause",
        response_model=TorrentTaskOut,
        responses={404: {"model": ErrorOut}, 409: {"model": ErrorOut}},
    )
    async def pause_task(task_id: int) -> TorrentTaskOut:
        task = await _owned(task_id)
        # сервис при неверном статусе вернул бы задачу без изменений — ловим предпроверкой
        if task.status not in _PAUSEABLE:
            raise HTTPException(
                status_code=409, detail=f"cannot pause task in status {task.status}"
            )
        return await _mutated("pause", await _guard("pause", service.pause(api_user_id, task_id)))

    @router.post(
        "/{task_id}/resume",
        response_model=TorrentTaskOut,
        responses={404: {"model": ErrorOut}, 409: {"model": ErrorOut}},
    )
    async def resume_task(task_id: int) -> TorrentTaskOut:
        task = await _owned(task_id)
        if task.status is not TorrentStatus.PAUSED:
            raise HTTPException(
                status_code=409, detail=f"cannot resume task in status {task.status}"
            )
        return await _mutated(
            "resume", await _guard("resume", service.resume(api_user_id, task_id))
        )

    @router.post(
        "/{task_id}/cancel",
        response_model=TorrentTaskOut,
        responses={404: {"model": ErrorOut}, 409: {"model": ErrorOut}},
    )
    async def cancel_task(task_id: int) -> TorrentTaskOut:
        task = await _owned(task_id)
        # сервис cancel не отклоняет никогда — 409 только на уже отменённую (design §Components)
        if task.status is TorrentStatus.CANCELLED:
            raise HTTPException(status_code=409, detail="task already cancelled")
        return await _mutated(
            "cancel", await _guard("cancel", service.cancel(api_user_id, task_id))
        )

    @router.delete(
        "/{task_id}",
        status_code=status.HTTP_204_NO_CONTENT,
        responses={404: {"model": ErrorOut}},
    )
    async def delete_task(task_id: int) -> None:
        removed = await _guard("delete", service.delete(api_user_id, task_id))
        if not removed:
            # False = «не найдено»; сбой provider тоже сюда (ponytail: сервис различить не даёт)
            raise HTTPException(status_code=404, detail="task not found")

    return router
```

- [ ] **Step 2: Создать `app/api/app.py`**

```python
from __future__ import annotations

from fastapi import FastAPI

from app.api.routes import create_router
from app.repositories import TorrentRepository
from app.services.torrent_service import TorrentService


def create_api_app(
    service: TorrentService, repository: TorrentRepository, api_user_id: int
) -> FastAPI:
    app = FastAPI(
        title="Torrent Download Bot API",
        version="0.1.0",
        docs_url="/docs",
        openapi_url="/openapi.json",
    )
    app.include_router(create_router(service, repository, api_user_id))
    return app
```

- [ ] **Step 3: Smoke-проверка + commit**

```bash
python -c "from app.api.app import create_api_app; print('ok')"
ruff check . && ruff format . && mypy app && pytest -q
git add app/api/routes.py app/api/app.py
git commit -m "feat: add REST API routes with Swagger app factory"
```

Ожидаемо: `ok`, всё зелёное (роуты ещё никем не вызваны — покрытие в Task 8).

**Acceptance criteria:**
- [ ] `mypy app` strict чистый; `ruff format --check` чистый.
- [ ] `create_api_app(service, repository, api_user_id)` собирает FastAPI с `/docs`, `/openapi.json` и префиксом `/api/tasks`.
- [ ] Каждый код из таблицы design §Endpoints отработан логикой роутов: 200/201/204/400/404/409/422/500.

---

### Task 7: wiring в `app/main.py` — uvicorn параллельно с polling

**Files:**
- Modify: `app/main.py`

**Interfaces:**
- Consumes: Task 2 (`config.api_host/api_port/api_user_id`), Task 6 (`create_api_app`), `dp.start_polling(bot)`, `uvicorn.Server.serve()`.
- Produces: рабочий `python -m app.main` = бот + API на `http://127.0.0.1:8080/docs`.

Семантика (design §Error Handling + requirement):
- сервер упал → залогировано (`logger.error(..., exc_info=True)`), **polling продолжает работать**;
- polling завершился (штатно или с ошибкой) → `server.should_exit = True` в `finally` → сервер гасится, исключение polling **всплывает** после завершения обоих (`raise result` после gather).
- `asyncio.gather(..., return_exceptions=True)`: падение одного не отменяет второго.
- **Порядок аргументов gather обязателен**: `dp.start_polling` первым — aiogram ставит свои signal-хендлеры до того, как uvicorn их переставит; при выходе uvicorn восстанавливает их и пересылает сигнал (Technical Findings §2). Guard `install_signal_handlers` не нужен и в uvicorn 0.54 не существует.
- `uvicorn.Config(..., log_config=None)` — не перебивать наш logging.

- [ ] **Step 1: Переписать `app/main.py`** (полный файл)

```python
from __future__ import annotations

import asyncio
import logging
import logging.handlers
from pathlib import Path

import uvicorn
from aiogram import Bot, Dispatcher

from app.api.app import create_api_app
from app.bot import create_dispatcher
from app.config import load_config
from app.db.pool import apply_schema, create_pool, redact_database_url
from app.providers.fake import FakeTorrentProvider
from app.repositories.postgres import PostgresRepository
from app.services.torrent_service import TorrentService

logger = logging.getLogger(__name__)


async def _run_polling(dp: Dispatcher, bot: Bot, server: uvicorn.Server) -> None:
    try:
        await dp.start_polling(bot)
    finally:
        # polling завершился (штатно или с ошибкой) — гасим API-сервер, иначе gather ждёт вечно
        server.should_exit = True


async def _run_api(server: uvicorn.Server) -> None:
    try:
        await server.serve()
    except Exception:
        # падение сервера НЕ роняет бота: логируем, polling продолжается (design §Error Handling)
        logger.error("API server failed; bot polling continues", exc_info=True)


async def main() -> None:
    config = load_config()
    format_ = "%(asctime)s %(levelname)s %(name)s: %(message)s"
    logging.basicConfig(level=config.log_level, format=format_)
    Path("logs").mkdir(exist_ok=True)
    file_handler = logging.handlers.RotatingFileHandler(
        "logs/bot.log", maxBytes=2_000_000, backupCount=3, encoding="utf-8"
    )
    file_handler.setLevel(config.log_level)
    file_handler.setFormatter(logging.Formatter(format_))
    logging.getLogger().addHandler(file_handler)
    if config.log_level == "DEBUG":
        # DEBUG не должен заливать логи polling-внутренностями
        logging.getLogger("aiogram").setLevel(logging.INFO)
        logging.getLogger("psycopg").setLevel(logging.INFO)
    # fail fast: нет DATABASE_URL -> RuntimeError из load_config; БД недоступна ->
    # RuntimeError из create_pool (redacted URL).
    # Никакого тихого memory-режима (design §Error Handling)
    pool = await create_pool(config.database_url)
    try:
        await apply_schema(pool)
        repository = PostgresRepository(pool)
        service = TorrentService(FakeTorrentProvider(), repository)
        dp: Dispatcher = create_dispatcher(service, config.test_mode)
        bot = Bot(token=config.bot_token)
        api_app = create_api_app(service, repository, config.api_user_id)
        server = uvicorn.Server(
            uvicorn.Config(api_app, host=config.api_host, port=config.api_port, log_config=None)
        )
        logger.info(
            "bot started (TEST_MODE=%s, db=%s, api=http://%s:%s/docs)",
            config.test_mode,
            redact_database_url(config.database_url),
            config.api_host,
            config.api_port,
        )
        # Порядок gather обязателен: aiogram первым ставит signal-хендлеры (loop.add_signal_handler),
        # uvicorn вторым перехватывает их через signal.signal и при выходе восстанавливает
        # + пересылает сигнал (uvicorn Server.capture_signals): Ctrl+C гасит API, затем polling.
        # return_exceptions: падение одного не отменяет другой (design §Error Handling).
        results = await asyncio.gather(
            _run_polling(dp, bot, server),
            _run_api(server),
            return_exceptions=True,
        )
        for result in results:
            if isinstance(result, BaseException):
                raise result  # всплытие ошибки polling (design: «и наоборот» — но бот первичен)
    finally:
        await pool.close()


if __name__ == "__main__":
    asyncio.run(main())
```

- [ ] **Step 2: Проверки**

```bash
ruff check . && ruff format . && mypy app && pytest -q
```

Ожидаемо: всё зелёное. Fail-fast без DATABASE_URL (как в Phase 1, Task 8):

```bash
env -u DATABASE_URL BOT_TOKEN=123:TEST python -m app.main
```

Ожидаемо: `RuntimeError: DATABASE_URL is not set...`, код выхода ≠ 0 (API тоже не поднимается — порядок: сначала БД).

- [ ] **Step 3: Commit**

```bash
git add app/main.py
git commit -m "feat: serve REST API alongside bot polling"
```

**Acceptance criteria:**
- [ ] `pytest` зелёный; `mypy app` чистый.
- [ ] В логе старта видно `api=http://127.0.0.1:8080/docs`.
- [ ] Ручной чек (Task 9): бот и `/docs` живут одновременно; Ctrl+C останавливает оба; упавший порт (занят) логируется, бот продолжает работать.

---

### Task 8: `tests/integration/test_api.py`

**Files:**
- Create: `tests/integration/test_api.py`

**Interfaces:**
- Consumes: Task 3 (`InMemoryRepository.ensure_user`), Task 5/6 (схемы, `create_api_app`), `FakeTorrentProvider`, `TorrentService`.
- Produces: 8 тестов, покрывающих design §Testing Strategy (флоу, дубль→409, 404, 422, 400×2, 409 неверный статус, openapi/docs).

Тесты — `async def` (auto-mode), фикстуры async (нужен `await repo.ensure_user`); TestClient синхронный — вызывается внутри async-теста как есть. `with TestClient(app)` — без lifespan (его нет и не нужен: БД уже в репозитории-дубле).

- [ ] **Step 1: Создать `tests/integration/test_api.py`**

```python
from __future__ import annotations

from collections.abc import AsyncGenerator

import pytest
from fastapi.testclient import TestClient

from app.api.app import create_api_app
from app.providers.fake import FakeTorrentProvider
from app.repositories.memory import InMemoryRepository
from app.services.torrent_service import TorrentService

_API_USER = 1
_MAGNET = "magnet:?xt=urn:btih:" + "ab" * 20 + "&dn=Test.Torrent"
_MAGNET_OTHER = "magnet:?xt=urn:btih:" + "cd" * 20 + "&dn=Other"


@pytest.fixture
async def repo() -> AsyncGenerator[InMemoryRepository, None]:
    r = InMemoryRepository()
    await r.ensure_user(_API_USER)  # «пользователь нажал /start»
    yield r


@pytest.fixture
async def client(repo: InMemoryRepository) -> AsyncGenerator[TestClient, None]:
    service = TorrentService(FakeTorrentProvider(), repo)
    with TestClient(create_api_app(service, repo, _API_USER)) as c:
        yield c


async def test_create_list_flow(client: TestClient) -> None:
    created = client.post(
        "/api/tasks",
        json={"magnet": _MAGNET, "media_type": "movie", "quality": "1080p"},
    )
    assert created.status_code == 201
    body = created.json()
    assert body["id"] == 1
    assert body["user_id"] == _API_USER
    assert body["status"] == "queued"  # StrEnum -> plain string в JSON
    assert body["media_type"] == "movie"
    assert body["quality"] == "1080p"

    listed = client.get("/api/tasks")
    assert listed.status_code == 200
    assert [t["id"] for t in listed.json()] == [1]

    single = client.get("/api/tasks/1")
    assert single.status_code == 200
    assert single.json()["name"] == "Test.Torrent"


async def test_pause_resume_cancel_delete_flow(client: TestClient) -> None:
    task_id = client.post(
        "/api/tasks",
        json={"magnet": _MAGNET, "media_type": "movie", "quality": "1080p"},
    ).json()["id"]

    paused = client.post(f"/api/tasks/{task_id}/pause")
    assert paused.status_code == 200
    assert paused.json()["status"] == "paused"

    resumed = client.post(f"/api/tasks/{task_id}/resume")
    assert resumed.status_code == 200
    assert resumed.json()["status"] == "downloading"

    cancelled = client.post(f"/api/tasks/{task_id}/cancel")
    assert cancelled.status_code == 200
    assert cancelled.json()["status"] == "cancelled"

    deleted = client.delete(f"/api/tasks/{task_id}")
    assert deleted.status_code == 204
    assert deleted.content == b""

    assert client.get(f"/api/tasks/{task_id}").status_code == 404


async def test_duplicate_returns_409_with_existing_task(client: TestClient) -> None:
    first = client.post(
        "/api/tasks",
        json={"magnet": _MAGNET, "media_type": "movie", "quality": "1080p"},
    )
    assert first.status_code == 201

    duplicate = client.post(
        "/api/tasks",
        json={"magnet": _MAGNET, "media_type": "series", "quality": "720p"},
    )
    assert duplicate.status_code == 409
    detail = duplicate.json()["detail"]
    assert detail["id"] == first.json()["id"]
    assert detail["media_type"] == "movie"  # существующая задача, а не новая


async def test_wrong_state_mutations_409(client: TestClient) -> None:
    task_id = client.post(
        "/api/tasks",
        json={"magnet": _MAGNET, "media_type": "movie", "quality": "1080p"},
    ).json()["id"]

    assert client.post(f"/api/tasks/{task_id}/pause").status_code == 200
    assert client.post(f"/api/tasks/{task_id}/pause").status_code == 409  # уже на паузе

    assert client.post(f"/api/tasks/{task_id}/resume").status_code == 200
    assert client.post(f"/api/tasks/{task_id}/resume").status_code == 409  # не на паузе

    assert client.post(f"/api/tasks/{task_id}/cancel").status_code == 200
    assert client.post(f"/api/tasks/{task_id}/cancel").status_code == 409  # уже отменена


async def test_get_and_delete_404(client: TestClient) -> None:
    assert client.get("/api/tasks/999").status_code == 404
    assert client.delete("/api/tasks/999").status_code == 404


async def test_create_422_validation(client: TestClient) -> None:
    invalid_enum = client.post(
        "/api/tasks",
        json={"magnet": _MAGNET, "media_type": "book", "quality": "1080p"},
    )
    assert invalid_enum.status_code == 422

    missing_field = client.post("/api/tasks", json={"magnet": _MAGNET})
    assert missing_field.status_code == 422


async def test_create_400_bad_magnet(client: TestClient) -> None:
    response = client.post(
        "/api/tasks",
        json={"magnet": "http://example.com/file.torrent", "media_type": "movie", "quality": "1080p"},
    )
    assert response.status_code == 400
    assert "scheme" in response.json()["detail"]


async def test_create_400_when_api_user_unknown() -> None:
    repo = InMemoryRepository()  # никто не нажал /start
    service = TorrentService(FakeTorrentProvider(), repo)
    client = TestClient(create_api_app(service, repo, _API_USER))
    response = client.post(
        "/api/tasks",
        json={"magnet": _MAGNET, "media_type": "movie", "quality": "1080p"},
    )
    assert response.status_code == 400
    assert "/start" in response.json()["detail"]


async def test_openapi_and_docs(client: TestClient) -> None:
    schema = client.get("/openapi.json")
    assert schema.status_code == 200
    assert "/api/tasks" in schema.json()["paths"]
    assert client.get("/docs").status_code == 200
```

- [ ] **Step 2: Запустить**

```bash
pytest tests/integration/test_api.py -v
```

Ожидаемо: 8 passed. Если что-то красное — чинить в `app/api/*` (НЕ в тестах, если ошибка не в самом тесте), никогда не удалять тесты.

Замечания к ожидаемому поведению (проверь при отладке):
- pause сразу после create: `_sync` внутри сервиса может успеть перевести `queued → downloading` (FakeTorrentProvider: 2-й tick) — оба статуса pauseable, предпроверка пропустит.
- resume ставит `downloading` (provider.resume → DOWNLOADING), cancel ставит `cancelled` — минимум get_status-ticks до `completed` (8) в тестах не набирается.

- [ ] **Step 3: Полный прогон + commit**

```bash
pytest -q
git add tests/integration/test_api.py
git commit -m "test: add REST API integration tests"
```

**Acceptance criteria:**
- [ ] Полный `pytest` — 36 прежних + 4 config + 8 API = 48 passed, 0 failed.
- [ ] Флоу create→list→pause→resume→cancel→delete, дубль→409 (detail = существующая задача), 404, 422, 400×2 (magnet и /start), 409 неверный статус, `/openapi.json` 200, `/docs` 200 — проверены.

---

### Task 9: Docs + ручной чек + финальная верификация

**Files:**
- Create: `docs/features/020-rest-api.md`
- Create: `docs/decisions/0004-http-api.md`
- Modify: `docs/ROADMAP.md` (Phase 1.5 + «Порядок развития»)
- Modify: `README.md` (секция HTTP API + счётчик тестов)
- Modify: `docs/API.md` (§7 HTTP API)

**Interfaces:**
- Consumes: Tasks 1–8 (весь реализованный функционал).
- Produces: документация по шаблонам init.md §19/§25, зелёный DoD, feature = Implemented.

- [ ] **Step 1: `docs/features/020-rest-api.md`** (шаблон init.md §19; Status ставим `Implemented` в Step 6, когда DoD зелёный — до этого `In Progress`)

```markdown
# FEAT-020 — REST API + Swagger UI

## Status

Implemented

## Goal

Дать пользователю интерфейс для UI вне Telegram: публичный локальный HTTP API
с автодокументацией Swagger, поверх уже существующего `TorrentService`.

## User Story

As a user, I want to manage my downloads through a local HTTP API with Swagger docs,
so I can build my own UI outside Telegram.

## Scope

- FastAPI + uvicorn, параллельно с Telegram-polling (один процесс, `asyncio.gather`).
- Эндпоинты `/api/tasks`: list/get/create/pause/resume/cancel/delete.
- Swagger UI (`/docs`) и OpenAPI (`/openapi.json`) из коробки.
- `API_HOST`/`API_PORT`/`API_USER_ID` в config; регистрация пользователя на `/start`.

## Out of Scope

- Аутентификация (API слушает только localhost — ADR-0004).
- WebSocket (см. «Future» ниже).
- Изменения `TorrentService`/Telegram UI (кроме регистрации на /start).

## User Flow

```text
HTTP request → FastAPI route (валидация) → TorrentService → repository/provider → JSON
```

Задача, созданная через API, видна в `/downloads` бота, и наоборот (общая PostgreSQL).

## UI

Не меняется. Swagger UI: `http://127.0.0.1:8080/docs`.

## Domain Changes

Нет новых доменных моделей; API-схемы (`TaskCreateIn`, `TorrentTaskOut`, `ErrorOut`)
отражают существующие `TorrentTask`/enum'ы (StrEnum → plain strings в JSON).

## API / Interfaces

`app/api/`: `create_api_app(service, repository, api_user_id) -> FastAPI`;
роуты поверх `TorrentService`; repository получил `ensure_user`/`user_exists`.

## Storage Changes

Нет (та же PostgreSQL). Строка `users` для `API_USER_ID` создаётся при `/start`.

## Security

- Без аутентификации → только localhost, fail-fast в config (API_HOST).
- Секреты не попадают в ответы/логи.

## Error Handling

- 422 — валидация Pydantic; 400 — невалидный magnet / пользователь не нажал /start;
  404 — не найдено/чужое; 409 — дубль (detail = существующая задача) и неверный статус;
  500 — неожиданная ошибка (лог с traceback, тело `{"detail": "internal error"}`).
- Падение uvicorn не роняет polling (логируется); завершение polling гасит сервер,
  ошибка polling всплывает.

## Tests

`tests/integration/test_api.py` (TestClient + InMemoryRepository): флоу,
дубль→409, 404, 422, 400×2, 409 неверный статус, `/openapi.json`, `/docs`;
`tests/unit/test_config.py` — дефолты/парсинг API-переменных.

## Dependencies

- FEAT-001 — FEAT-010 (сервис, репозитории, PostgreSQL).

## Configuration

`API_HOST` (default `127.0.0.1`, только localhost), `API_PORT` (default `8080`),
`API_USER_ID` (default `1`) — актор API.

## Implementation Tasks

- [x] Зависимости (fastapi, uvicorn; httpx dev)
- [x] Config: API_HOST/API_PORT/API_USER_ID + fail-fast localhost
- [x] Repository: `ensure_user`/`user_exists`; `/start` регистрирует пользователя
- [x] `app/api/`: schemas + routes + `create_api_app`
- [x] `app/main.py`: uvicorn параллельно с `dp.start_polling` (gather, return_exceptions)
- [x] Тесты: `tests/integration/test_api.py`, `tests/unit/test_config.py` (+4)
- [ ] Future: `WS /api/ws` события задач — отдельная задача, REST не меняется

## Acceptance Criteria

- [x] `GET/POST /api/tasks`, pause/resume/cancel/delete отвечают кодами design §Endpoints
- [x] Swagger UI доступен на `/docs`, схема на `/openapi.json`
- [x] Задачи общие с ботом (общая БД, ownership = API_USER_ID)
- [x] Бот и API работают одновременно; падение одного не убивает другое молча
- [x] `ruff`/`mypy`/`pytest` зелёные

## Related Documentation

- `docs/API.md` §7
- `docs/decisions/0004-http-api.md`
- `docs/CONFIGURATION.md`
- `docs/ROADMAP.md`

## Future: `WS /api/ws` события задач

WebSocket для live-обновлений прогресса добавляется позже отдельной задачей;
REST-контракт не меняется.
```

- [ ] **Step 2: `docs/decisions/0004-http-api.md`** (ADR-шаблон init.md §25)

```markdown
# ADR-0004: HTTP API на FastAPI, localhost без auth, API_USER_ID как актор

## Status

Accepted

## Context

Нужен UI вне Telegram поверх готового `TorrentService` (без Telegram-кода,
design Phase 1). Нужна автодокументация (пользователь пишет свой фронтенд).
Проект в TEST_MODE: реальных загрузок нет, слушать наружу опасно.

## Decision

- **FastAPI + uvicorn** (runtime-зависимости): Swagger UI и OpenAPI из коробки;
  альтернативы (Starlette вручную, aiohttp) отклонены — нет нормальной автодоки.
- **Только localhost, без аутентификации** — config fail-fast на не-localhost API_HOST.
- **`API_USER_ID` — актор API**: create ставит его владельцем, остальные операции
  передают его в сервис; ownership-проверка = ровно поведение Telegram (FEAT-009).
- **`/start` регистрирует пользователя** (`ensure_user`): до этого `users` создавался
  только вместе с первой задачей/настройкой, и «нажми /start» был бы нерабочей петлёй.
- **Параллельно с polling**: uvicorn `Server.serve()` и `dp.start_polling` в одном
  event loop через `asyncio.gather(..., return_exceptions=True)`; порядок аргументов
  (polling первым) обязателен из-за перехвата signal-хендлеров uvicorn'ом.
- **WebSocket отложен** — будущая отдельная задача, REST не меняется.

## Consequences

Плюсы:

- Swagger UI без дополнительной инфраструктуры; фронтенд пользователя кодогенерится из openapi.json.
- Общая PostgreSQL: задачи видны и в боте, и в API; бизнес-слой не переписан.

Минусы:

- Нет auth — нельзя слушать не-localhost (охраняется config-валидацией).
- `API_USER_ID` ≠ вашего telegram id = пустой список (документировано в CONFIGURATION.md).
- Сбой provider при delete маппится в 404 (сервис не различает причины False).
```

- [ ] **Step 3: `docs/ROADMAP.md`** — после секции `## Phase 1 — Persistent State` вставить:

```markdown
## Phase 1.5 — REST API (Swagger)

- [x] FastAPI + Swagger UI (`/docs`)
- [x] `/api/tasks`: list/get/create/pause/resume/cancel/delete
- [x] `API_USER_ID` как актор, localhost без auth
```

и в «Порядок развития» после строки `→ Persistent State` добавить `→ REST API`.

- [ ] **Step 4: `README.md`**:

1. После секции «Установка и запуск» добавить:

```markdown
## HTTP API

Локальный REST API с Swagger UI: `http://127.0.0.1:8080/docs`
(OpenAPI-схема — `/openapi.json`). Работает от имени пользователя `API_USER_ID`
(см. [docs/CONFIGURATION.md](docs/CONFIGURATION.md)); без аутентификации — слушает
только localhost. Перед первым `POST /api/tasks` нажмите `/start` в боте.
Контракт: [docs/API.md](docs/API.md) §7.
```

2. Обновить счётчик в «Тестирование → Автотесты»: `pytest            # 48 тестов: ...`
   (подставить фактическое число из `pytest -q`).

- [ ] **Step 5: `docs/API.md`** — добавить раздел в конец файла:

```markdown
---

## 7. HTTP API

FastAPI-приложение `app/api/` поверх того же `TorrentService` (FEAT-020, ADR-0004).

| Метод | Путь | Успех | Ошибки |
|---|---|---|---|
| GET | `/api/tasks` | 200, список задач | — |
| GET | `/api/tasks/{id}` | 200 | 404 |
| POST | `/api/tasks` | 201 | 400 magnet/пользователь, 409 дубль, 422 валидация |
| POST | `/api/tasks/{id}/pause` | 200 | 404, 409 неверный статус |
| POST | `/api/tasks/{id}/resume` | 200 | 404, 409 |
| POST | `/api/tasks/{id}/cancel` | 200 | 404, 409 (уже отменена) |
| DELETE | `/api/tasks/{id}` | 204 | 404 |

- Актор — `API_USER_ID`; ownership = ровно поведение Telegram (§5).
- Swagger UI: `/docs`, схема: `/openapi.json`. Только localhost, без аутентификации.
- Неожиданная ошибка → 500 `{"detail": "internal error"}` + traceback в лог.
- WebSocket (`/api/ws`) — будущая задача, REST-контракт не изменится.
```

- [ ] **Step 6: Финальная верификация (все четыре команды обязательны, DEVELOPMENT §9)**

```bash
ruff check .
ruff format --check .
mypy app
pytest
```

Ожидаемо: без ошибок; `pytest` — 48 passed, exit code 0. Если красное — чинить и перезапускать все четыре до зелёного; никогда не «чинить» удалением тестов.

- [ ] **Step 7: Ручной чек** (требует реального `.env` с `BOT_TOKEN`+`DATABASE_URL`; если их нет — запросить у пользователя):

1. `python -m app.main` → в логе `bot started (..., api=http://127.0.0.1:8080/docs)`.
2. `curl -s http://127.0.0.1:8080/openapi.json | head` → JSON; `curl -s -o /dev/null -w "%{http_code}" http://127.0.0.1:8080/docs` → `200`.
3. `curl -s -X POST http://127.0.0.1:8080/api/tasks -H 'Content-Type: application/json' -d '{"magnet":"magnet:?xt=urn:btih:a1b2c3d4e5f60718293a4b5c6d7e8f9012345678&dn=Test","media_type":"movie","quality":"1080p"}'` → 201 (если 400 — нажать `/start` в боте и повторить).
4. Задача видна в `/downloads` бота (общая БД).
5. `Ctrl+C` → процесс завершается, в логе видны shutdown uvicorn и polling; повторный запарт работает.
6. (опционально) Занять порт `lsof -i :8080` перед стартом → в логе `API server failed; bot polling continues`, бот продолжает работать.

- [ ] **Step 8: Feature status + commit**

`docs/features/020-rest-api.md` → `Status: Implemented` (если Step 6 зелёный; иначе `In Progress` и не коммитить как готовое).

```bash
git add docs/features/020-rest-api.md docs/decisions/0004-http-api.md docs/ROADMAP.md README.md docs/API.md
git commit -m "docs: document FEAT-020 REST API"
```

**Acceptance criteria:**
- [ ] `ruff check .` / `ruff format --check .` / `mypy app` / `pytest` — все зелёные (48 passed).
- [ ] Feature 020: `Implemented`, критерии отмечены; ADR-0004 на месте.
- [ ] README содержит секцию «HTTP API»; ROADMAP — Phase 1.5; API.md — §7.
- [ ] Ручной чек пройден: `/docs` отвечает, задача через API видна в боте, Ctrl+C останавливает оба.

---

## Self-Review Notes (для ревьюера плана)

- **Spec coverage:** design §Components (app/api-слои, config 3 поля, эндпоинты, схемы) — Tasks 2/5/6; §Data Flow (общая БД, актор API_USER_ID) — Tasks 6/7; §Error Handling (400/404/409/422/500, gather-семантика) — Tasks 6/7; §Testing Strategy (test_api + test_config, финальная верификация) — Tasks 2/8/9; §Docs (features/020, ADR-0004, CONFIGURATION, ROADMAP, README, API.md) — Tasks 2/9; WebSocket-future — feature-doc §Future.
- **Отклонения от чернового списка waves (осознанные):**
  - Repository `ensure_user`/`user_exists` вынесены в Wave A (независимая задача), регистрация на `/start` — в Wave B: без этого «нажми /start» из design нерабочий (см. Technical Findings §3).
  - Тесты config живут в Task 2 (TDD: красный → зелёный, как в Phase 1 Task 2), а не в Wave D — тот же контент, правильный порядок.
  - Guard `config.install_signal_handlers = False` НЕ включён: в uvicorn 0.54 такой опции нет, а `capture_signals` корректно восстанавливает хендлеры aiogram и пересылает сигнал (исходники 0.54.0 проверены) — вместо guard'а: порядок аргументов gather + `should_exit` + ручной Ctrl+C чек.
- **409 для cancel:** сервис cancel не отклоняет никогда — 409 только на уже `cancelled` (минимальное отклонение, сохраняющее Telegram-поведение для остальных статусов); отдельно задокументировано в routes.py.
- **Delete False→404:** сервис не различает «не найдено» и сбой provider — accepted simplification (ponytail-комментарий в routes.py).
- **Type consistency:** `ensure_user`/`user_exists` — протокол (Task 3) = memory/postgres (Task 3) = вызовы `register_user` (Task 4) и `repository.user_exists` (Task 6) = сидирование в тестах (Task 8). `create_api_app(service, repository, api_user_id)` — определение (Task 6) = вызовы в main (Task 7) и тестах (Task 8). `Config.api_*` (Task 2) = main (Task 7). `_PAUSEABLE`/статусы = поведение сервиса (pause: QUEUED/DOWNLOADING, resume: PAUSED).
- **Счётчик тестов:** 36 + 4 + 8 = 48; в README Step 4 указана формулировка «подставить фактическое число» — если executor добавит тестов, число берётся из `pytest -q`.
