# Phase 0 Bot UI (FEAT-001…009) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Реализовать полный пользовательский flow Telegram-бота (FEAT-001…009): `/start` → magnet → категория → качество → подтверждение → тестовая очередь → `/downloads` с управлением — без реальных торрент-загрузок.

**Architecture:** Один процесс aiogram 3.x. Слои: `handlers` (только команды/FSM/рендер) → `TorrentService` (валидация, duplicate по `info_hash`, ownership, orchestration) → `TorrentProvider` Protocol → `FakeTorrentProvider` (in-memory, симуляция прогресса). Состояние пользователя — только FSM на `MemoryStorage`, глобальных переменных нет. БД нет (FEAT-010 позже).

**Tech Stack:** Python 3.12+, aiogram 3.x, python-dotenv, pytest + pytest-asyncio, ruff, mypy (strict).

**Specs (аргументы плана):**
- Дизайн: `thoughts/shared/designs/2026-10-04-phase0-bot-ui-design.md`
- Features: `docs/features/001-start-menu.md` … `009-user-isolation.md` (+ `018-duplicate-detection.md` для алгоритма дублей)
- Контракты: `docs/API.md`, `docs/ARCHITECTURE.md`, `docs/TESTING.md`, `docs/DEVELOPMENT.md`

---

## Global Constraints

- **Scope = только FEAT-001…009.** Запрещено: qBittorrent, NAS, реальные downloads, запись/перемещение файлов, media organizer, Lampa (DEVELOPMENT §2).
- **`TEST_MODE=true` обязателен.** При `false` бот на Phase 0 не стартует (config поднимает ошибку).
- **Runtime-зависимости только:** `aiogram` (3.x, `<4`), `python-dotenv`. Dev: pytest, pytest-asyncio, pytest-cov, ruff, mypy. Новых зависимостей не добавлять (DEVELOPMENT §5).
- **Python 3.12+**, FSM на `MemoryStorage`, БД нет.
- **Слои обязательны (ADR-0001):** бизнес-логика только в `TorrentService`; handlers не вызывают provider напрямую; handlers не содержат бизнес-логики торрентов (DEVELOPMENT §3).
- **Дубликаты — по `info_hash`** (ADR-0002, FEAT-018), **изоляция — по `user_id`** (FEAT-009).
- **Callback data — короткие строки формата `namespace:action:значение`**, без magnet/JSON/секретов (ARCHITECTURE §7, API.md §4).
- **DoD (DEVELOPMENT §9):** `ruff check .`, `ruff format --check .`, `mypy app`, `pytest` — все успешно.
- Логирование через `logging`, не `print`; не логировать `BOT_TOKEN` (DEVELOPMENT §6).
- Секреты (`.env`, токен) не коммитить (DEVELOPMENT §7). Репозиторий сейчас не git — шаги коммитов в плане нет.
- Тесты не требуют внешней инфраструктуры (TESTING §общие правила).

### Интерпретации спецификации (зафиксированы, не обсуждаются при исполнении)

1. **`torrent:quality:unknown`** — четвёртая кнопка FEAT-004 («Не определять»). Whitelist API.md §4 задаёт форму `torrent:quality:<value>`; значение `unknown` следует тому же шаблону. Строки таблицы API.md §4 дописываются в Task 19.
2. **Per-task callbacks** (FEAT-007/008): `torrent:open|pause|resume|cancel|delete:<task_id>`, `torrent:list:back`, плюс `menu:downloads|settings|help` (FEAT-001). Тот же формат коротких строк, id — int. Дописываются в API.md §4 в Task 19.
3. **Scope дубликата = `(user_id, info_hash)`.** FEAT-018 в Phase 1 переходит в unique constraint `info_hash + user_id`; глобальный дубль раскрыл бы чужую задачу пользователю (конфликт с FEAT-009). Одинаковый magnet у другого пользователя → своя отдельная задача.
4. **`torrent:confirm:back` («Изменить»)** возвращает на экран категории (FEAT-003), т.е. к блоку magnet-информации с вопросом «Куда сохранить?» — так доступны для правки и категория, и качество; качество выбирается заново на следующем шаге. «Предыдущий шаг» FEAT-005 трактуется как предыдущий экран выбора, а не только качество.
5. **`cancel` — операция только сервиса** (ставит `CANCELLED`): контракт `TorrentProvider` в API.md §2 не содержит `cancel`, менять контракт на Phase 0 нельзя; FEAT-011 добавит метод в provider. Провайдер-ошибки при pause/resume/add/delete → статус `FAILED` + сообщение (FEAT-008).
6. **Симуляция прогресса:** `FakeTorrentProvider.get_status` двигает жизненный цикл (`QUEUED` → `DOWNLOADING` → `COMPLETED`) на каждом чтении; `TorrentService._sync` добавляет `+12.5%` за каждый рендер в состоянии `DOWNLOADING`, при `COMPLETED` → `100.0`. Наблюдать прогресс вручную — повторными `/downloads` (Task 19 это описывает).
7. **aiogram v3:** хендлер без state-фильтра работает в *любом* состоянии — поэтому все wizard-колбэки получают явный state-фильтр (`TorrentWizard.<state>` первым аргументом декоратора), а `torrent:confirm:cancel` — без фильтра, но с проверкой `state.get_state() is not None`.
8. **Порядок роутеров и fallback:** по исходникам aiogram (`Router._propagate_event`) сначала проверяются свои хендлеры роутера, затем `sub_routers` по порядку включения. Поэтому: (а) командные роутеры включаются **до** `magnet` (иначе `/downloads` съест `F.text`-хендлер); (б) catch-all на неизвестные callback — отдельный роутер, включённый **последним** (регистрация `@dp.callback_query()` прямо на диспетчере перехватила бы всё до роутеров).
9. **Дупликация `_edit`** (~8 строк) в `magnet.py` и `downloads.py` — намеренная: дополнительный файл утилит не входит в утверждённый список путей.
10. **`FORMAT_SIZE`** живёт в `app/models/torrent.py` (общий для FEAT-003-экрана и детального вида FEAT-007).
11. **`TorrentTask.provider_id`** — добавлено к минимуму из ARCHITECTURE §4 (поле нужно, чтобы хранить id, возвращённый `provider.add_magnet`, для FEAT-011).

---

## File Structure

| Файл | Ответственность | Задача |
|---|---|---|
| `pyproject.toml` | tool-config: ruff, mypy strict, pytest (`pythonpath=["."]`, `asyncio_mode="auto"`), `[project]` metadata | T1 |
| `requirements.txt` | все зависимости (runtime + dev), единственная команда установки | T1 |
| `.env.example` | `BOT_TOKEN`, `TEST_MODE=true` (без секретов) | T1 |
| `app/config.py` | `Config`, `load_config()`: dotenv, валидация `BOT_TOKEN` и `TEST_MODE=true` | T2 |
| `app/models/media.py` | enum `MediaType`, `Quality` + константы отображения (`SAVE_PATHS`, `MEDIA_LABELS`, `MEDIA_EMOJI`, `QUALITY_LABELS`) | T3 |
| `app/models/torrent.py` | enum `TorrentStatus`, dataclass `TorrentTask`, `format_size()` | T4 |
| `app/services/magnet_parser.py` | `MagnetInfo`, `MagnetParseError`, `parse_magnet()` — чистая функция без Telegram | T5 |
| `app/providers/base.py` | `TorrentProvider` Protocol (API.md §2) | T6 |
| `app/providers/fake.py` | `FakeTorrentProvider` — in-memory, без сети/файловой системы | T7 |
| `app/services/torrent_service.py` | create/list/get/pause/resume/cancel/delete; duplicate по `(user_id, info_hash)`; ownership; `_sync` | T8 |
| `app/keyboards/main.py` | главное меню (📥/⚙️/ℹ️) | T9 |
| `app/keyboards/torrent.py` | `category_keyboard`, `quality_keyboard`, `confirm_keyboard` | T10 |
| `app/keyboards/downloads.py` | `downloads_keyboard(tasks)`, `detail_keyboard(task)` | T11 |
| `app/handlers/start.py` | `/start` | T12 |
| `app/handlers/settings.py` | заглушки `menu:settings`, `menu:help` | T13 |
| `app/handlers/magnet.py` | wizard: текст-magnet → категория → качество → подтверждение → `create` (FEAT-002/003/004/005/006/018) | T14 |
| `app/handlers/downloads.py` | `/downloads`, `menu:downloads`, детальный вид, pause/resume/cancel/delete, ownership (FEAT-007/008/009) | T15 |
| `app/bot.py` | `create_dispatcher(service, test_mode)`: DI, порядок роутеров, fallback, error handler | T16 |
| `app/main.py` | `async def main()`: logging, config, wiring, `start_polling` | T17 |
| `tests/unit/test_magnet_parser.py` | 7 обязательных кейсов parser | T5 |
| `tests/unit/test_fake_provider.py` | no-network + жизненный цикл | T7 |
| `tests/unit/test_torrent_service.py` | create/duplicate/ownership/переходы статусов/прогресс | T8 |
| `tests/integration/test_bot_flow.py` | полный flow через `dp.feed_raw_update` + recording session | T18 |
| `docs/TESTING.md` | раздел «Ручное тестирование в Telegram» + синхронизация доков | T19 |

`app/__init__.py` и `tests/__init__.py` **не нужны**: Python namespace packages + `pythonpath = ["."]` в pytest-конфиге.

---

## Dependency-Ordered Task Breakdown (параллельное исполнение)

```text
T1  scaffolding (pyproject, requirements, .env.example)
 ├─ T2  config                     T3  models/media ── T4  models/torrent ─┬─ T6  providers/base ── T7 providers/fake
 │                              │                                        │        │                    │
 ├─ T5  magnet_parser + tests ───┼────────────────────────────────────────┼────────┼────────────────────┤
 ├─ T9  keyboards/main ── T12 handlers/start                             │        │                    │
 └─ T13 handlers/settings       T10 keyboards/torrent                    T11 keyboards/downloads     │
                                T15 handlers/downloads ◄───────────────── T8  torrent_service + tests ◄┘
T14 handlers/magnet ◄── (T3, T5, T8, T10)
T16 bot ◄── (T12, T13, T14, T15)
T17 main ◄── (T2, T7, T8, T16)
T18 integration tests ◄── T17
T19 TESTING.md manual section (+ API.md §4, статусы features) ◄── T18
```

**Волны (внутри волны задачи независимы, можно диспетчировать параллельно):**

| Волна | Задачи |
|---|---|
| W1 | T1 |
| W2 | T2, T3, T5, T9, T13 |
| W3 | T4, T10, T12 |
| W4 | T6, T11 |
| W5 | T7 |
| W6 | T8 |
| W7 | T14, T15 |
| W8 | T16 |
| W9 | T17 |
| W10 | T18 |
| W11 | T19 |

После каждой задачи: `ruff check . && ruff format . && mypy app && pytest` (pytest — с T5; до T5 прогонять только lint/mypy). Если `ruff check` ругается только на порядок импортов (I001) — `ruff check --fix .` и повторить. Финальный гейт — в конце документа.

---

## Testing Matrix (обязательные тесты)

| Требование | Тест | Задача |
|---|---|---|
| magnet: valid magnet | `test_valid_magnet` | T5 |
| magnet: missing dn | `test_missing_dn` | T5 |
| magnet: multiple trackers | `test_multiple_trackers` | T5 |
| magnet: URL encoded dn | `test_url_encoded_dn` | T5 |
| magnet: missing info hash | `test_missing_info_hash` (raises `MagnetParseError`) | T5 |
| magnet: invalid scheme | `test_invalid_scheme` (raises `MagnetParseError`) | T5 |
| magnet: xl parameter | `test_xl_parameter` | T5 |
| service: create | `test_create_makes_queued_task` | T8 |
| service: duplicate by info_hash | `test_duplicate_by_info_hash_same_user`, `test_different_info_hash_creates_separate_tasks`, `test_other_user_can_create_same_hash` | T8 |
| service: ownership / изоляция | `test_user_isolation_reads`, `test_ownership_enforced_on_operations` | T8 |
| service: переходы статусов | `test_status_transitions`, `test_progress_reaches_completed` | T8 |
| fake provider без сети | `test_fake_provider_never_touches_network` (блок DNS/connect), `test_fake_provider_simulates_lifecycle` | T7 |
| integration: полный flow | `test_full_flow_start_to_downloads` (`/start → magnet → category → quality → confirmation → queue → /downloads`) | T18 |
| integration: ошибки/robustness | `test_invalid_magnet_gets_friendly_error`, `test_unknown_callback_is_ignored` | T18 |

---

## Задачи

---

### Task 1: Scaffolding — pyproject, requirements, .env.example

**Files:**
- Create: `pyproject.toml`
- Create: `requirements.txt`
- Create: `.env.example`

**Depends on:** —
**Interfaces:**
- Produces: конфигурация инструментов (`ruff`, `mypy strict`, `pytest asyncio_mode=auto`, `pythonpath=["."]`), команда установки `pip install -r requirements.txt`.

**Spec:** DEVELOPMENT §4 (Python Tooling), §5 (Dependencies); TESTING §Команды.

- [ ] **Шаг 1: Создать `pyproject.toml`**

```toml
[project]
name = "torrent-download-bot"
version = "0.1.0"
description = "Telegram bot for torrent downloads (Phase 0: UI + fake queue)"
requires-python = ">=3.12"
dependencies = [
    "aiogram>=3.25,<4",
    "python-dotenv>=1.0",
]

[tool.ruff]
line-length = 100
target-version = "py312"

[tool.ruff.lint]
select = ["E", "F", "I", "UP", "B", "SIM"]

[tool.mypy]
python_version = "3.12"
strict = true
files = ["app"]

[tool.pytest.ini_options]
testpaths = ["tests"]
pythonpath = ["."]
asyncio_mode = "auto"
```

- [ ] **Шаг 2: Создать `requirements.txt`** (единственная команда установки; синхронизировать с `dependencies` выше)

```text
# runtime — синхронизировать с [project].dependencies в pyproject.toml
aiogram>=3.25,<4
python-dotenv>=1.0

# dev (DEVELOPMENT §5)
pytest>=8
pytest-asyncio>=0.24
pytest-cov>=5
ruff>=0.6
mypy>=1.11
```

- [ ] **Шаг 3: Создать `.env.example`** (секретов нет, `.env` не коммитится)

```text
BOT_TOKEN=123456:REPLACE_WITH_TOKEN_FROM_BOTFATHER
TEST_MODE=true
```

- [ ] **Шаг 4: Установить зависимости и проверить инструменты**

```bash
python3 -m venv .venv && source .venv/bin/activate
pip install -r requirements.txt
ruff check . && ruff format . && mypy app
```
Expected: все команды выходят с кодом 0 (`mypy` — «no issues» на пустом `app` — каталог `app/` ещё не создан, mypy допустимо запустить после Task 2; до этого достаточно `ruff check . && ruff format .`). `pytest` до T5 не запускать (exit code 5 «no tests collected»). Если `ruff check` предлагает автоисправления — `ruff check --fix .`.

**Acceptance Criteria:**
- [ ] `ruff`, `pytest`-конфиг, `mypy` настроены в `pyproject.toml`; `asyncio_mode = "auto"`.
- [ ] Установка — одной командой `pip install -r requirements.txt`.
- [ ] В `.env.example` нет реальных секретов.

---

### Task 2: `app/config.py`

**Files:**
- Create: `app/config.py`

**Depends on:** T1
**Interfaces:**
- Produces: `Config` (frozen dataclass: `bot_token: str`, `test_mode: bool`), `load_config() -> Config` — используется в T17 (`main.py`).

**Spec:** FEAT-001 Configuration (`BOT_TOKEN`); DEVELOPMENT §2 (`TEST_MODE=true` критически важно).

- [ ] **Шаг 1: Создать `app/config.py`**

```python
from __future__ import annotations

import os
from dataclasses import dataclass

from dotenv import load_dotenv


@dataclass(frozen=True)
class Config:
    bot_token: str
    test_mode: bool


def load_config() -> Config:
    load_dotenv()
    token = os.getenv("BOT_TOKEN")
    if not token:
        raise RuntimeError(
            "BOT_TOKEN is not set. Copy .env.example to .env and fill BOT_TOKEN."
        )
    test_mode = os.getenv("TEST_MODE", "true").strip().lower() in {"1", "true", "yes"}
    if not test_mode:
        raise RuntimeError(
            "Phase 0 requires TEST_MODE=true: real providers (qBittorrent/NAS) "
            "are not implemented yet."
        )
    return Config(bot_token=token, test_mode=test_mode)
```

- [ ] **Шаг 2: Проверки**

```bash
ruff check . && ruff format . && mypy app
```
Expected: PASS.

**Acceptance Criteria:**
- [ ] Без `BOT_TOKEN` — понятная ошибка, не stacktrace из глубины бота.
- [ ] `TEST_MODE=false` → бот отказывается стартовать (DEVELOPMENT §2).
- [ ] Лint/type чистые; токен нигде не логируется.

---

### Task 3: `app/models/media.py`

**Files:**
- Create: `app/models/media.py`

**Depends on:** T1
**Interfaces:**
- Produces:
  - `MediaType(str, Enum)`: `MOVIE="movie"`, `SERIES="series"`
  - `Quality(str, Enum)`: `UHD_2160="2160p"`, `FHD_1080="1080p"`, `HD_720="720p"`, `UNKNOWN="unknown"`
  - `SAVE_PATHS: dict[MediaType, str]` — `MOVIE→"Movies/"`, `SERIES→"TV/"` (FEAT-005 «Путь зависит от категории», конфиг-контракт API.md §6)
  - `MEDIA_LABELS: dict[MediaType, str]` — `"Фильм"` / `"Сериал"` (FEAT-003/005 UI)
  - `MEDIA_EMOJI: dict[MediaType, str]` — `"🎬"` / `"📺"` (FEAT-007 UI)
  - `QUALITY_LABELS: dict[Quality, str]` — `"2160p / 4K"`, `"1080p"`, `"720p"`, `"Не определять"` (FEAT-004 кнопки, FEAT-005 сводка, FEAT-007 список)

**Spec:** FEAT-003 Domain Changes; FEAT-004 Domain Changes; ARCHITECTURE §4.

- [ ] **Шаг 1: Создать `app/models/media.py`**

```python
from __future__ import annotations

from enum import Enum


class MediaType(str, Enum):
    MOVIE = "movie"
    SERIES = "series"


class Quality(str, Enum):
    UHD_2160 = "2160p"
    FHD_1080 = "1080p"
    HD_720 = "720p"
    UNKNOWN = "unknown"


SAVE_PATHS: dict[MediaType, str] = {
    MediaType.MOVIE: "Movies/",
    MediaType.SERIES: "TV/",
}

MEDIA_LABELS: dict[MediaType, str] = {
    MediaType.MOVIE: "Фильм",
    MediaType.SERIES: "Сериал",
}

MEDIA_EMOJI: dict[MediaType, str] = {
    MediaType.MOVIE: "🎬",
    MediaType.SERIES: "📺",
}

QUALITY_LABELS: dict[Quality, str] = {
    Quality.UHD_2160: "2160p / 4K",
    Quality.FHD_1080: "1080p",
    Quality.HD_720: "720p",
    Quality.UNKNOWN: "Не определять",
}
```

- [ ] **Шаг 2: Проверки** — `ruff check . && ruff format . && mypy app` → PASS.

**Acceptance Criteria:**
- [ ] Используются enum, а не произвольные строки (FEAT-003/004 AC).
- [ ] `SAVE_PATHS` покрывает обе категории: `Movies/`, `TV/`.

---

### Task 4: `app/models/torrent.py`

**Files:**
- Create: `app/models/torrent.py`

**Depends on:** T3
**Interfaces:**
- Consumes: `MediaType`, `Quality` (T3).
- Produces:
  - `TorrentStatus(str, Enum)`: `QUEUED/DOWNLOADING/PAUSED/COMPLETED/FAILED/CANCELLED` (ARCHITECTURE §4)
  - `TorrentTask` dataclass — поля из ARCHITECTURE §4 + `provider_id: str = ""` (интерпретация №11)
  - `format_size(size_bytes: int | None) -> str` — UI-хелпер (FEAT-003 «не определён», FEAT-007 размер)

**Spec:** FEAT-006 Domain Changes; ARCHITECTURE §4; FEAT-003 UI (размер).

- [ ] **Шаг 1: Создать `app/models/torrent.py`**

```python
from __future__ import annotations

from dataclasses import dataclass, field
from datetime import datetime, timezone
from enum import Enum

from app.models.media import MediaType, Quality


class TorrentStatus(str, Enum):
    QUEUED = "queued"
    DOWNLOADING = "downloading"
    PAUSED = "paused"
    COMPLETED = "completed"
    FAILED = "failed"
    CANCELLED = "cancelled"


@dataclass
class TorrentTask:
    id: int
    user_id: int
    magnet: str
    info_hash: str
    name: str
    media_type: MediaType
    quality: Quality
    status: TorrentStatus = TorrentStatus.QUEUED
    progress: float = 0.0
    size_bytes: int | None = None
    download_speed: int = 0
    upload_speed: int = 0
    eta_seconds: int | None = None
    created_at: datetime = field(default_factory=lambda: datetime.now(timezone.utc))
    completed_at: datetime | None = None
    # ponytail: provider_id сверх минимума ARCHITECTURE §4 — нужен для FEAT-011 (qBittorrent)
    provider_id: str = ""


def format_size(size_bytes: int | None) -> str:
    """Человекочитаемый размер для UI (FEAT-003/007). None -> «не определён»."""
    if size_bytes is None:
        return "не определён"
    value = float(size_bytes)
    for unit in ("B", "KiB", "MiB", "GiB", "TiB"):
        if value < 1024 or unit == "TiB":
            if unit == "B":
                return f"{int(value)} B"
            return f"{value:.1f} {unit}"
        value /= 1024
    return f"{value:.1f} TiB"
```

- [ ] **Шаг 2: Проверки** — `ruff check . && ruff format . && mypy app` → PASS.

**Acceptance Criteria:**
- [ ] Все поля `TorrentTask` из ARCHITECTURE §4 присутствуют.
- [ ] `format_size(None)` → `"не определён"`, `format_size(12884901888)` → `"12.0 GiB"`.

---

### Task 5: `app/services/magnet_parser.py` + unit-тесты (TDD)

**Files:**
- Create: `app/services/magnet_parser.py`
- Test: `tests/unit/test_magnet_parser.py`

**Depends on:** T1
**Interfaces:**
- Produces:
  - `MagnetInfo` (frozen dataclass): `raw: str`, `info_hash: str`, `display_name: str | None`, `trackers: list[str]`, `size_bytes: int | None`
  - `MagnetParseError(ValueError)`
  - `parse_magnet(raw: str) -> MagnetInfo` — чистая функция, без Telegram-кода (API.md §1)
- Потребитель: T14 (handler), T8 (service принимает `MagnetInfo`).

**Spec:** FEAT-002 целиком; API.md §1; TESTING §Минимальные тесты Magnet Parser.

- [ ] **Шаг 1: Написать failing-тесты — `tests/unit/test_magnet_parser.py` (все 7 обязательных кейсов)**

```python
from __future__ import annotations

import pytest

from app.services.magnet_parser import MagnetParseError, parse_magnet

VALID = (
    "magnet:?xt=urn:btih:a1b2c3d4e5f60718293a4b5c6d7e8f9012345678"
    "&dn=The.Matrix.1999.1080p.BluRay"
    "&tr=udp%3A%2F%2Ftracker.example.org%3A1337%2Fannounce"
    "&xl=12884901888"
    "&xs=http%3A%2F%2Funrelated.example"
)


def test_valid_magnet() -> None:
    info = parse_magnet(VALID)
    assert info.raw == VALID
    assert info.info_hash == "a1b2c3d4e5f60718293a4b5c6d7e8f9012345678"
    assert info.display_name == "The.Matrix.1999.1080p.BluRay"
    assert info.trackers == ["udp://tracker.example.org:1337/announce"]
    assert info.size_bytes == 12884901888


def test_missing_dn() -> None:
    info = parse_magnet("magnet:?xt=urn:btih:" + "a" * 40)
    assert info.display_name is None
    assert info.trackers == []
    assert info.size_bytes is None


def test_multiple_trackers() -> None:
    raw = (
        "magnet:?xt=urn:btih:" + "b" * 40
        + "&tr=udp%3A%2F%2Fone%2Fannounce"
        + "&tr=udp%3A%2F%2Ftwo%2Fannounce"
    )
    info = parse_magnet(raw)
    assert info.trackers == ["udp://one/announce", "udp://two/announce"]


def test_url_encoded_dn() -> None:
    raw = "magnet:?xt=urn:btih:" + "c" * 40 + "&dn=The%20Matrix%20%281999%29"
    assert parse_magnet(raw).display_name == "The Matrix (1999)"


def test_missing_info_hash() -> None:
    with pytest.raises(MagnetParseError):
        parse_magnet("magnet:?dn=NoHashHere")


def test_invalid_scheme() -> None:
    with pytest.raises(MagnetParseError):
        parse_magnet("https://example.com/file.torrent")
    with pytest.raises(MagnetParseError):
        parse_magnet("just some plain text")


def test_xl_parameter() -> None:
    raw = "magnet:?xt=urn:btih:" + "d" * 40 + "&xl=987654321"
    assert parse_magnet(raw).size_bytes == 987654321
```

- [ ] **Шаг 2: Запустить, убедиться что падает**

```bash
pytest tests/unit/test_magnet_parser.py -v
```
Expected: FAIL — `ModuleNotFoundError: No module named 'app.services.magnet_parser'`.

- [ ] **Шаг 3: Реализовать `app/services/magnet_parser.py`**

```python
from __future__ import annotations

from dataclasses import dataclass
from urllib.parse import parse_qs, urlparse


class MagnetParseError(ValueError):
    """Magnet-ссылка не прошла валидацию (FEAT-002 Error Handling)."""


@dataclass(frozen=True)
class MagnetInfo:
    raw: str
    info_hash: str
    display_name: str | None
    trackers: list[str]
    size_bytes: int | None


def parse_magnet(raw: str) -> MagnetInfo:
    parsed = urlparse(raw.strip())
    if parsed.scheme != "magnet":
        raise MagnetParseError(f"unsupported scheme: {parsed.scheme!r}")

    params = parse_qs(parsed.query)

    info_hash = ""
    for xt in params.get("xt", []):
        if xt.startswith("urn:btih:"):
            info_hash = xt.removeprefix("urn:btih:").strip().lower()
            break
    if not info_hash:
        raise MagnetParseError("magnet has no info hash (xt=urn:btih:...)")

    display_name = params["dn"][0] if "dn" in params else None
    trackers = list(params.get("tr", []))

    size_bytes: int | None = None
    xl_values = params.get("xl", [])
    if xl_values:
        try:
            size_bytes = int(xl_values[0])
        except ValueError:
            size_bytes = None

    return MagnetInfo(
        raw=raw,
        info_hash=info_hash,
        display_name=display_name,
        trackers=trackers,
        size_bytes=size_bytes,
    )
```

Примечания: `parse_qs` сам делает URL decoding (поэтому `dn`/`tr` уже декодированы); неизвестные параметры (`xs` и др.) игнорируются без падения; формат hash не валидируется сверх непустоты (тесты FEAT-002 требуют только наличия) — «валидация схемы и info hash» выполнена как scheme-проверка + presence-check; `raw` сохраняется исходным (FEAT-002 AC).

- [ ] **Шаг 4: Запустить тесты**

```bash
pytest tests/unit/test_magnet_parser.py -v
```
Expected: 7 passed.

- [ ] **Шаг 5: Лint/типы** — `ruff check . && ruff format . && mypy app` → PASS.

**Acceptance Criteria (FEAT-002 AC):**
- [ ] Все 7 обязательных тестов проходят.
- [ ] Парсер не падает на неизвестных параметрах; исходный magnet в `raw`.
- [ ] `ruff`, `mypy`, `pytest` чистые.

---

### Task 6: `app/providers/base.py`

**Files:**
- Create: `app/providers/base.py`

**Depends on:** T4
**Interfaces:**
- Consumes: `TorrentStatus` (T4).
- Produces: `TorrentProvider` Protocol — точная сигнатура API.md §2:

```python
class TorrentProvider(Protocol):
    async def add_magnet(self, magnet: str, save_path: str) -> str: ...
    async def get_status(self, torrent_id: str) -> TorrentStatus: ...
    async def pause(self, torrent_id: str) -> None: ...
    async def resume(self, torrent_id: str) -> None: ...
    async def delete(self, torrent_id: str, delete_files: bool = False) -> None: ...
```

Потребители: T7 (fake реализует структурно), T8 (`TorrentService.__init__`).

**Spec:** API.md §2; ARCHITECTURE §2 (Providers); FEAT-006 API/Interfaces.

- [ ] **Шаг 1: Создать `app/providers/base.py`**

```python
from __future__ import annotations

from typing import Protocol

from app.models.torrent import TorrentStatus


class TorrentProvider(Protocol):
    """Контракт провайдера (docs/API.md §2). Handlers никогда не работают с ним напрямую."""

    async def add_magnet(self, magnet: str, save_path: str) -> str: ...

    async def get_status(self, torrent_id: str) -> TorrentStatus: ...

    async def pause(self, torrent_id: str) -> None: ...

    async def resume(self, torrent_id: str) -> None: ...

    async def delete(self, torrent_id: str, delete_files: bool = False) -> None: ...
```

- [ ] **Шаг 2: Проверки** — `ruff check . && ruff format . && mypy app` → PASS.

**Acceptance Criteria:**
- [ ] Сигнатуры 1-в-1 совпадают с API.md §2 (минус `cancel` — интерпретация №5).
- [ ] Наследование не требуется — Protocol, структурная типизация.

---

### Task 7: `app/providers/fake.py` + unit-тесты (TDD)

**Files:**
- Create: `app/providers/fake.py`
- Test: `tests/unit/test_fake_provider.py`

**Depends on:** T6
**Interfaces:**
- Consumes: `TorrentProvider` (T6), `TorrentStatus` (T4).
- Produces: `FakeTorrentProvider` — реализует Protocol; жизненный цикл: `QUEUED` (тикает 1–2 чтения) → `DOWNLOADING` (до 8-го чтения) → `COMPLETED`; `pause`/`resume`/`delete` — мгновенные. Никаких HTTP/файловых операций.

**Spec:** FEAT-006 (FakeTorrentProvider: не делает HTTP, не подключается к qBittorrent, не обращается к NAS, симулирует очередь и прогресс); TESTING §общие правила.

- [ ] **Шаг 1: Написать failing-тесты — `tests/unit/test_fake_provider.py`**

```python
from __future__ import annotations

import asyncio
import socket

from app.models.torrent import TorrentStatus
from app.providers.fake import FakeTorrentProvider


def test_fake_provider_never_touches_network(monkeypatch) -> None:
    def _blocked_dns(*args, **kwargs):
        raise AssertionError("DNS lookup attempted")

    def _blocked_connect(self, *args, **kwargs):
        raise AssertionError("socket.connect attempted")

    async def exercise() -> None:
        # патчи ставим внутри корутины: event loop уже создан asyncio.run
        monkeypatch.setattr(socket, "getaddrinfo", _blocked_dns)
        monkeypatch.setattr(socket.socket, "connect", _blocked_connect)
        provider = FakeTorrentProvider()
        torrent_id = await provider.add_magnet("magnet:?xt=urn:btih:" + "a" * 40, "Movies/")
        await provider.get_status(torrent_id)
        await provider.pause(torrent_id)
        assert await provider.get_status(torrent_id) is TorrentStatus.PAUSED
        await provider.resume(torrent_id)
        await provider.delete(torrent_id)

    asyncio.run(exercise())


def test_fake_provider_simulates_lifecycle() -> None:
    async def run() -> list[TorrentStatus]:
        provider = FakeTorrentProvider()
        torrent_id = await provider.add_magnet("magnet:?xt=urn:btih:" + "b" * 40, "Movies/")
        return [await provider.get_status(torrent_id) for _ in range(10)]

    seen = asyncio.run(run())
    assert seen[0] is TorrentStatus.QUEUED
    assert TorrentStatus.DOWNLOADING in seen
    assert seen[-1] is TorrentStatus.COMPLETED
```

- [ ] **Шаг 2: Запустить, убедиться что падает**

```bash
pytest tests/unit/test_fake_provider.py -v
```
Expected: FAIL — `ModuleNotFoundError: No module named 'app.providers.fake'`.

- [ ] **Шаг 3: Реализовать `app/providers/fake.py`**

```python
from __future__ import annotations

import itertools
from dataclasses import dataclass

from app.models.torrent import TorrentStatus


@dataclass
class _FakeRecord:
    status: TorrentStatus = TorrentStatus.QUEUED
    ticks: int = 0


class FakeTorrentProvider:
    """In-memory провайдер для TEST_MODE: без сети, без файловой системы, без qBittorrent."""

    def __init__(self) -> None:
        self._torrents: dict[str, _FakeRecord] = {}
        self._ids = itertools.count(1)

    async def add_magnet(self, magnet: str, save_path: str) -> str:
        torrent_id = f"fake-{next(self._ids)}"
        self._torrents[torrent_id] = _FakeRecord()
        return torrent_id

    async def get_status(self, torrent_id: str) -> TorrentStatus:
        record = self._torrents[torrent_id]
        if record.status not in (TorrentStatus.QUEUED, TorrentStatus.DOWNLOADING):
            return record.status
        record.ticks += 1
        if record.status is TorrentStatus.QUEUED and record.ticks >= 2:
            record.status = TorrentStatus.DOWNLOADING
        elif record.status is TorrentStatus.DOWNLOADING and record.ticks >= 8:
            record.status = TorrentStatus.COMPLETED
        return record.status

    async def pause(self, torrent_id: str) -> None:
        self._torrents[torrent_id].status = TorrentStatus.PAUSED

    async def resume(self, torrent_id: str) -> None:
        self._torrents[torrent_id].status = TorrentStatus.DOWNLOADING

    async def delete(self, torrent_id: str, delete_files: bool = False) -> None:
        del self._torrents[torrent_id]
```

- [ ] **Шаг 4: Запустить тесты**

```bash
pytest tests/unit/test_fake_provider.py -v
```
Expected: 2 passed.

- [ ] **Шаг 5: Лint/типы** — `ruff check . && ruff format . && mypy app` → PASS (mypy проверит структурное соответствие Protocol в T8, здесь — сам файл).

**Acceptance Criteria (FEAT-006 AC):**
- [ ] Никаких сетевых запросов и файловых операций (тест блокирует `socket.getaddrinfo` и `socket.connect`).
- [ ] Очередь/прогресс симулируются детерминированно (без таймеров — воспроизводимо в тестах).
- [ ] `ruff`, `mypy`, `pytest` чистые.

---

### Task 8: `app/services/torrent_service.py` + unit-тесты (TDD)

**Files:**
- Create: `app/services/torrent_service.py`
- Test: `tests/unit/test_torrent_service.py`

**Depends on:** T4, T5, T7
**Interfaces:**
- Consumes: `MagnetInfo` (T5), `TorrentProvider` (T6), `TorrentStatus`/`TorrentTask` (T4), `SAVE_PATHS`/`MediaType`/`Quality` (T3), `FakeTorrentProvider` (только в тестах).
- Produces (точные сигнатуры — по ним пишутся T14/T15):

```python
class TorrentService:
    def __init__(self, provider: TorrentProvider) -> None: ...
    async def create(self, *, user_id: int, magnet: MagnetInfo,
                     media_type: MediaType, quality: Quality) -> tuple[TorrentTask, bool]
        # (task, created): created=False → показать существующий дубликат (FEAT-018)
    async def list(self, user_id: int) -> list[TorrentTask]
    async def get(self, user_id: int, task_id: int) -> TorrentTask | None
    async def pause(self, user_id: int, task_id: int) -> TorrentTask | None
    async def resume(self, user_id: int, task_id: int) -> TorrentTask | None
    async def cancel(self, user_id: int, task_id: int) -> TorrentTask | None
    async def delete(self, user_id: int, task_id: int) -> bool
```

Контракт ошибок: `None`/`False` = «задача не найдена или чужая» (FEAT-009 — не раскрывать существование); `task.status is FAILED` = ошибка provider (FEAT-008); никакого Telegram-кода (API.md §3).

**Spec:** FEAT-006, FEAT-008, FEAT-009, FEAT-018; API.md §3, §5.

- [ ] **Шаг 1: Написать failing-тесты — `tests/unit/test_torrent_service.py`**

```python
from __future__ import annotations

import pytest

from app.models.media import MediaType, Quality
from app.models.torrent import TorrentStatus
from app.providers.fake import FakeTorrentProvider
from app.services.magnet_parser import MagnetInfo
from app.services.torrent_service import TorrentService

HASH_A = "a" * 40
HASH_B = "b1" + "0" * 38


def _magnet(info_hash: str = HASH_A, name: str | None = "The Matrix") -> MagnetInfo:
    raw = f"magnet:?xt=urn:btih:{info_hash}" + (f"&dn={name}" if name else "")
    return MagnetInfo(
        raw=raw,
        info_hash=info_hash.lower(),
        display_name=name,
        trackers=[],
        size_bytes=None,
    )


@pytest.fixture
def service() -> TorrentService:
    return TorrentService(FakeTorrentProvider())


async def test_create_makes_queued_task(service: TorrentService) -> None:
    task, created = await service.create(
        user_id=1, magnet=_magnet(), media_type=MediaType.MOVIE, quality=Quality.FHD_1080
    )
    assert created is True
    assert task.status is TorrentStatus.QUEUED
    assert task.user_id == 1
    assert task.name == "The Matrix"
    assert task.provider_id.startswith("fake-")


async def test_duplicate_by_info_hash_same_user(service: TorrentService) -> None:
    first, created_first = await service.create(
        user_id=1, magnet=_magnet(), media_type=MediaType.MOVIE, quality=Quality.FHD_1080
    )
    second, created_second = await service.create(
        user_id=1, magnet=_magnet(), media_type=MediaType.SERIES, quality=Quality.HD_720
    )
    assert created_first is True
    assert created_second is False
    assert second.id == first.id


async def test_different_info_hash_creates_separate_tasks(service: TorrentService) -> None:
    first, _ = await service.create(
        user_id=1, magnet=_magnet(HASH_A), media_type=MediaType.MOVIE, quality=Quality.FHD_1080
    )
    second, created = await service.create(
        user_id=1, magnet=_magnet(HASH_B), media_type=MediaType.MOVIE, quality=Quality.FHD_1080
    )
    assert created is True
    assert second.id != first.id
    assert len(await service.list(1)) == 2


async def test_other_user_can_create_same_hash(service: TorrentService) -> None:
    # FEAT-009 + интерпретация №3: дубль scope = (user_id, info_hash)
    mine, _ = await service.create(
        user_id=1, magnet=_magnet(), media_type=MediaType.MOVIE, quality=Quality.FHD_1080
    )
    other, created = await service.create(
        user_id=2, magnet=_magnet(), media_type=MediaType.MOVIE, quality=Quality.FHD_1080
    )
    assert created is True
    assert other.id != mine.id


async def test_user_isolation_reads(service: TorrentService) -> None:
    task, _ = await service.create(
        user_id=1, magnet=_magnet(), media_type=MediaType.MOVIE, quality=Quality.FHD_1080
    )
    assert await service.list(2) == []
    assert await service.get(2, task.id) is None
    assert await service.get(1, task.id) is not None


async def test_ownership_enforced_on_operations(service: TorrentService) -> None:
    task, _ = await service.create(
        user_id=1, magnet=_magnet(), media_type=MediaType.MOVIE, quality=Quality.FHD_1080
    )
    assert await service.pause(2, task.id) is None
    assert await service.resume(2, task.id) is None
    assert await service.cancel(2, task.id) is None
    assert await service.delete(2, task.id) is False
    untouched = await service.get(1, task.id)
    assert untouched is not None
    assert untouched.status is TorrentStatus.QUEUED


async def test_status_transitions(service: TorrentService) -> None:
    task, _ = await service.create(
        user_id=1, magnet=_magnet(), media_type=MediaType.MOVIE, quality=Quality.FHD_1080
    )
    assert task.status is TorrentStatus.QUEUED

    for _ in range(3):
        task = (await service.list(1))[0]
        if task.status is not TorrentStatus.QUEUED:
            break
    assert task.status is TorrentStatus.DOWNLOADING

    paused = await service.pause(1, task.id)
    assert paused is not None and paused.status is TorrentStatus.PAUSED
    assert (await service.list(1))[0].status is TorrentStatus.PAUSED

    resumed = await service.resume(1, task.id)
    assert resumed is not None and resumed.status is TorrentStatus.DOWNLOADING

    cancelled = await service.cancel(1, task.id)
    assert cancelled is not None and cancelled.status is TorrentStatus.CANCELLED
    assert (await service.list(1))[0].status is TorrentStatus.CANCELLED

    assert await service.delete(1, task.id) is True
    assert await service.list(1) == []


async def test_progress_reaches_completed(service: TorrentService) -> None:
    task, _ = await service.create(
        user_id=1, magnet=_magnet(), media_type=MediaType.MOVIE, quality=Quality.FHD_1080
    )
    for _ in range(20):
        task = (await service.list(1))[0]
        if task.status is TorrentStatus.COMPLETED:
            break
    assert task.status is TorrentStatus.COMPLETED
    assert task.progress == 100.0
    assert task.completed_at is not None
```

- [ ] **Шаг 2: Запустить, убедиться что падает**

```bash
pytest tests/unit/test_torrent_service.py -v
```
Expected: FAIL — `ModuleNotFoundError: No module named 'app.services.torrent_service'`.

- [ ] **Шаг 3: Реализовать `app/services/torrent_service.py`**

```python
from __future__ import annotations

import logging
from datetime import datetime, timezone

from app.models.media import MediaType, Quality, SAVE_PATHS
from app.models.torrent import TorrentStatus, TorrentTask
from app.providers.base import TorrentProvider
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

    Без Telegram-кода (docs/API.md §3).
    """

    def __init__(self, provider: TorrentProvider) -> None:
        self._provider = provider
        self._tasks: dict[int, TorrentTask] = {}
        self._next_id = 1

    async def create(
        self,
        *,
        user_id: int,
        magnet: MagnetInfo,
        media_type: MediaType,
        quality: Quality,
    ) -> tuple[TorrentTask, bool]:
        existing = await self._find_existing(user_id, magnet.info_hash)
        if existing is not None:
            logger.info(
                "duplicate info_hash=%s user_id=%s task_id=%s",
                magnet.info_hash, user_id, existing.id,
            )
            return existing, False

        task = TorrentTask(
            id=self._next_id,
            user_id=user_id,
            magnet=magnet.raw,
            info_hash=magnet.info_hash,
            name=magnet.display_name or f"magnet-{magnet.info_hash[:8]}",
            media_type=media_type,
            quality=quality,
            size_bytes=magnet.size_bytes,
        )
        self._next_id += 1
        self._tasks[task.id] = task

        try:
            task.provider_id = await self._provider.add_magnet(
                task.magnet, SAVE_PATHS[media_type]
            )
        except Exception:
            logger.error("provider.add_magnet failed task_id=%s", task.id, exc_info=True)
            task.status = TorrentStatus.FAILED

        logger.info(
            "created task_id=%s user_id=%s info_hash=%s",
            task.id, user_id, magnet.info_hash,
        )
        return task, True

    async def list(self, user_id: int) -> list[TorrentTask]:
        tasks = [t for t in self._tasks.values() if t.user_id == user_id]
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
            task.status = TorrentStatus.FAILED
            return task
        task.status = TorrentStatus.PAUSED
        return task

    async def resume(self, user_id: int, task_id: int) -> TorrentTask | None:
        task = await self._owned(user_id, task_id)
        if task is None or task.status is not TorrentStatus.PAUSED:
            return task
        try:
            await self._provider.resume(task.provider_id)
        except Exception:
            logger.error("provider.resume failed task_id=%s", task.id, exc_info=True)
            task.status = TorrentStatus.FAILED
            return task
        task.status = TorrentStatus.DOWNLOADING
        return task

    async def cancel(self, user_id: int, task_id: int) -> TorrentTask | None:
        task = await self._owned(user_id, task_id)
        if task is None:
            return None
        # ponytail: cancel без вызова provider — в контракте API.md §2 метода нет;
        # upgrade path: FEAT-011 добавит cancel в TorrentProvider
        task.status = TorrentStatus.CANCELLED
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
        del self._tasks[task_id]
        logger.info("deleted task_id=%s user_id=%s", task_id, user_id)
        return True

    async def _find_existing(self, user_id: int, info_hash: str) -> TorrentTask | None:
        for task in self._tasks.values():
            if task.user_id == user_id and task.info_hash == info_hash:
                await self._sync(task)
                return task
        return None

    async def _owned(self, user_id: int, task_id: int) -> TorrentTask | None:
        task = self._tasks.get(task_id)
        if task is None or task.user_id != user_id:
            # FEAT-009: чужая/несуществующая — трактовать как «не найдено»
            return None
        await self._sync(task)
        return task

    async def _sync(self, task: TorrentTask) -> None:
        if task.status in _TERMINAL:
            return
        try:
            status = await self._provider.get_status(task.provider_id)
        except Exception:
            logger.error(
                "provider.get_status failed task_id=%s", task.id, exc_info=True
            )
            task.status = TorrentStatus.FAILED
            return
        task.status = status
        if status is TorrentStatus.COMPLETED:
            task.progress = 100.0
            task.completed_at = datetime.now(timezone.utc)
        elif status is TorrentStatus.DOWNLOADING:
            # ponytail: фейковый прогресс +12.5% за рендер; upgrade path — данные из provider
            task.progress = min(99.0, task.progress + _PROGRESS_STEP)
```

- [ ] **Шаг 4: Запустить тесты**

```bash
pytest tests/unit/test_torrent_service.py -v
```
Expected: 7 passed.

- [ ] **Шаг 5: Лint/типы** — `ruff check . && ruff format . && mypy app && pytest` → PASS (весь suite: parser + fake + service).

**Acceptance Criteria:**
- [ ] create / duplicate-by-info_hash / ownership / переходы статусов покрыты (Testing Matrix).
- [ ] Никакого Telegram-кода в сервисе; дубликат возвращается как `(existing, False)` — один асинхронный вызов, без гонки check-then-create (FEAT-018 Error Handling).
- [ ] Ошибки provider → `FAILED`, исключения не всплывают в handlers (FEAT-006/008).
- [ ] `ruff`, `mypy`, `pytest` чистые.

---

### Task 9: `app/keyboards/main.py`

**Files:**
- Create: `app/keyboards/main.py`

**Depends on:** T1
**Interfaces:**
- Produces: `main_menu_keyboard() -> InlineKeyboardMarkup` — кнопки `menu:downloads`, `menu:settings`, `menu:help`.
- Потребители: T12 (`/start`), T14 (экраны после confirm/cancel), T15 (пустой список).

**Spec:** FEAT-001 UI (кнопки «📥 Мои загрузки», «⚙️ Настройки», «ℹ️ Помощь»), Security (короткие callbacks без секретов).

- [ ] **Шаг 1: Создать `app/keyboards/main.py`**

```python
from __future__ import annotations

from aiogram.types import InlineKeyboardButton, InlineKeyboardMarkup


def main_menu_keyboard() -> InlineKeyboardMarkup:
    return InlineKeyboardMarkup(
        inline_keyboard=[
            [
                InlineKeyboardButton(
                    text="📥 Мои загрузки", callback_data="menu:downloads"
                )
            ],
            [
                InlineKeyboardButton(
                    text="⚙️ Настройки", callback_data="menu:settings"
                )
            ],
            [InlineKeyboardButton(text="ℹ️ Помощь", callback_data="menu:help")],
        ]
    )
```

- [ ] **Шаг 2: Проверки** — `ruff check . && ruff format . && mypy app` → PASS.

**Acceptance Criteria:**
- [ ] Ровно три кнопки из FEAT-001; callback — короткие строки, без секретов.

---

### Task 10: `app/keyboards/torrent.py`

**Files:**
- Create: `app/keyboards/torrent.py`

**Depends on:** T3
**Interfaces:**
- Consumes: `MEDIA_LABELS`, `QUALITY_LABELS`, `MediaType`, `Quality` (T3).
- Produces:
  - `category_keyboard()` — `torrent:category:movie`, `torrent:category:series`, `torrent:confirm:cancel`
  - `quality_keyboard()` — `torrent:quality:2160|1080|720|unknown` (FEAT-004, интерпретация №1)
  - `confirm_keyboard()` — `torrent:confirm:add|back|cancel`
- Потребитель: T14.

**Spec:** FEAT-003 UI, FEAT-004 UI, FEAT-005 UI; API.md §4.

- [ ] **Шаг 1: Создать `app/keyboards/torrent.py`**

```python
from __future__ import annotations

from aiogram.types import InlineKeyboardButton, InlineKeyboardMarkup

from app.models.media import MEDIA_LABELS, QUALITY_LABELS, MediaType, Quality


def category_keyboard() -> InlineKeyboardMarkup:
    return InlineKeyboardMarkup(
        inline_keyboard=[
            [
                InlineKeyboardButton(
                    text=f"🎬 {MEDIA_LABELS[MediaType.MOVIE]}",
                    callback_data="torrent:category:movie",
                ),
                InlineKeyboardButton(
                    text=f"📺 {MEDIA_LABELS[MediaType.SERIES]}",
                    callback_data="torrent:category:series",
                ),
            ],
            [
                InlineKeyboardButton(
                    text="❌ Отмена", callback_data="torrent:confirm:cancel"
                )
            ],
        ]
    )


def quality_keyboard() -> InlineKeyboardMarkup:
    return InlineKeyboardMarkup(
        inline_keyboard=[
            [
                InlineKeyboardButton(
                    text=QUALITY_LABELS[Quality.UHD_2160],
                    callback_data="torrent:quality:2160",
                ),
                InlineKeyboardButton(
                    text=QUALITY_LABELS[Quality.FHD_1080],
                    callback_data="torrent:quality:1080",
                ),
            ],
            [
                InlineKeyboardButton(
                    text=QUALITY_LABELS[Quality.HD_720],
                    callback_data="torrent:quality:720",
                ),
                InlineKeyboardButton(
                    text=QUALITY_LABELS[Quality.UNKNOWN],
                    callback_data="torrent:quality:unknown",
                ),
            ],
        ]
    )


def confirm_keyboard() -> InlineKeyboardMarkup:
    return InlineKeyboardMarkup(
        inline_keyboard=[
            [
                InlineKeyboardButton(
                    text="⬇️ Добавить в очередь",
                    callback_data="torrent:confirm:add",
                )
            ],
            [
                InlineKeyboardButton(
                    text="✏️ Изменить", callback_data="torrent:confirm:back"
                )
            ],
            [
                InlineKeyboardButton(
                    text="❌ Отмена", callback_data="torrent:confirm:cancel"
                )
            ],
        ]
    )
```

- [ ] **Шаг 2: Проверки** — `ruff check . && ruff format . && mypy app` → PASS.

**Acceptance Criteria:**
- [ ] Все callback из whitelist API.md §4 (+ `torrent:quality:unknown` — интерпретация №1).
- [ ] Тексты кнопок соответствуют UI из FEAT-003/004/005.

---

### Task 11: `app/keyboards/downloads.py`

**Files:**
- Create: `app/keyboards/downloads.py`

**Depends on:** T4
**Interfaces:**
- Consumes: `TorrentTask`, `TorrentStatus` (T4).
- Produces:
  - `downloads_keyboard(tasks: list[TorrentTask]) -> InlineKeyboardMarkup` — по строке на задачу: активные `Открыть/Пауза/Отменить`, на паузе `Открыть/Продолжить/Отменить`, завершённые `Открыть/Удалить`; в каждой кнопке префикс `#id` (иначе строки неотличимы)
  - `detail_keyboard(task: TorrentTask)` — те же кнопки + `◀️ К списку` → `torrent:list:back`
- Потребитель: T15.

**Spec:** FEAT-007 UI (кнопки), FEAT-008 (действия); интерпретация №2.

- [ ] **Шаг 1: Создать `app/keyboards/downloads.py`**

```python
from __future__ import annotations

from aiogram.types import InlineKeyboardButton, InlineKeyboardMarkup

from app.models.torrent import TorrentStatus, TorrentTask


def downloads_keyboard(tasks: list[TorrentTask]) -> InlineKeyboardMarkup:
    rows: list[list[InlineKeyboardButton]] = []
    for task in tasks:
        prefix = f"#{task.id:03d}"
        row = [
            InlineKeyboardButton(
                text=f"{prefix} 📊 Открыть", callback_data=f"torrent:open:{task.id}"
            )
        ]
        if task.status in (TorrentStatus.QUEUED, TorrentStatus.DOWNLOADING):
            row += [
                InlineKeyboardButton(
                    text=f"{prefix} ⏸ Пауза",
                    callback_data=f"torrent:pause:{task.id}",
                ),
                InlineKeyboardButton(
                    text=f"{prefix} ❌ Отменить",
                    callback_data=f"torrent:cancel:{task.id}",
                ),
            ]
        elif task.status is TorrentStatus.PAUSED:
            row += [
                InlineKeyboardButton(
                    text=f"{prefix} ▶️ Продолжить",
                    callback_data=f"torrent:resume:{task.id}",
                ),
                InlineKeyboardButton(
                    text=f"{prefix} ❌ Отменить",
                    callback_data=f"torrent:cancel:{task.id}",
                ),
            ]
        else:
            row.append(
                InlineKeyboardButton(
                    text=f"{prefix} 🗑 Удалить",
                    callback_data=f"torrent:delete:{task.id}",
                )
            )
        rows.append(row)
    return InlineKeyboardMarkup(inline_keyboard=rows)


def detail_keyboard(task: TorrentTask) -> InlineKeyboardMarkup:
    markup = downloads_keyboard([task])
    rows = list(markup.inline_keyboard)
    rows.append(
        [
            InlineKeyboardButton(
                text="◀️ К списку", callback_data="torrent:list:back"
            )
        ]
    )
    return InlineKeyboardMarkup(inline_keyboard=rows)
```

- [ ] **Шаг 2: Проверки** — `ruff check . && ruff format . && mypy app` → PASS.

**Acceptance Criteria:**
- [ ] Набор кнопок соответствует статусу задачи (FEAT-007 AC: кнопки по состоянию).
- [ ] Каждая кнопка однозначно указывает задачу (`#id` в тексте + id в callback).

---

### Task 12: `app/handlers/start.py`

**Files:**
- Create: `app/handlers/start.py`

**Depends on:** T9
**Interfaces:**
- Consumes: `main_menu_keyboard()` (T9).
- Produces: модуль `router` c хендлером `/start`; сбрасывает FSM (`state.clear()` — escape hatch из любого состояния).

**Spec:** FEAT-001 целиком (текст сообщения и 3 кнопки — точно из спецификации); Error Handling: бот не падает на неизвестных сообщениях (unknown text обрабатывает T14, unknown callback — fallback в T16).

- [ ] **Шаг 1: Создать `app/handlers/start.py`**

```python
from __future__ import annotations

from aiogram import Router
from aiogram.filters import Command
from aiogram.fsm.context import FSMContext
from aiogram.types import Message

from app.keyboards.main import main_menu_keyboard

router = Router(name="start")

START_TEXT = "🤖 Torrent Bot\n\nОтправьте magnet-ссылку, чтобы создать новую загрузку."


@router.message(Command("start"))
async def cmd_start(message: Message, state: FSMContext) -> None:
    await state.clear()
    await message.answer(START_TEXT, reply_markup=main_menu_keyboard())
```

- [ ] **Шаг 2: Проверки** — `ruff check . && ruff format . && mypy app` → PASS.

**Acceptance Criteria (FEAT-001 AC):**
- [ ] Текст 1-в-1 из спецификации; показаны кнопки «Мои загрузки», «Настройки», «Помощь».

---

### Task 13: `app/handlers/settings.py`

**Files:**
- Create: `app/handlers/settings.py`

**Depends on:** T1
**Interfaces:**
- Produces: модуль `router` с `menu:settings` и `menu:help` — заглушки (design: «Настройки/Помощь — заглушки»).

**Spec:** FEAT-001 (кнопки Настройки/Помощь), FEAT-001 Error Handling (bot не падает).

- [ ] **Шаг 1: Создать `app/handlers/settings.py`**

```python
from __future__ import annotations

from aiogram import F, Router
from aiogram.types import CallbackQuery

router = Router(name="settings")


@router.callback_query(F.data == "menu:settings")
async def on_settings(callback: CallbackQuery) -> None:
    await callback.answer("⚙️ Настройки появятся позже")


@router.callback_query(F.data == "menu:help")
async def on_help(callback: CallbackQuery) -> None:
    await callback.answer(
        "Отправьте magnet-ссылку → выберите категорию и качество → подтвердите. "
        "Все загрузки: /downloads"
    )
```

- [ ] **Шаг 2: Проверки** — `ruff check . && ruff format . && mypy app` → PASS.

**Acceptance Criteria:**
- [ ] Обе кнопки отвечают тостом, бот не падает; тексты ≤ 200 символов (ограничение Telegram).

---

### Task 14: `app/handlers/magnet.py` — wizard FEAT-002/003/004/005/006/018

**Files:**
- Create: `app/handlers/magnet.py`

**Depends on:** T3, T5, T8, T10
**Interfaces:**
- Consumes: `parse_magnet`/`MagnetInfo`/`MagnetParseError` (T5), `TorrentService.create` (T8), `category/quality/confirm_keyboard` (T10), `main_menu_keyboard` (T9), `format_size` (T4), `SAVE_PATHS`-константы через `_summary_text` (T3), DI: `torrent_service: TorrentService`, `test_mode: bool` (T16 положит в `Dispatcher`).
- Produces: модуль `router`; FSM-группа `TorrentWizard(choosing_category, choosing_quality, confirming)`; тексты экранов; идемпотентный helper `_edit`.

**Spec:** FEAT-002 (получение magnet), FEAT-003 (экран информации + категория), FEAT-004 (качество), FEAT-005 (сводка + TEST MODE + add/back/cancel), FEAT-006 (очередь после confirm — через `create`), FEAT-018 (дубликат → показ существующего), ARCHITECTURE §5 (FSM, MemoryStorage, без глобальных переменных), §7 (callback whitelist).

- [ ] **Шаг 1: Создать `app/handlers/magnet.py`**

```python
from __future__ import annotations

import logging

from aiogram import F, Router
from aiogram.exceptions import TelegramBadRequest
from aiogram.fsm.context import FSMContext
from aiogram.fsm.state import State, StatesGroup
from aiogram.types import CallbackQuery, InlineKeyboardMarkup, Message

from app.keyboards.main import main_menu_keyboard
from app.keyboards.torrent import category_keyboard, confirm_keyboard, quality_keyboard
from app.models.media import MEDIA_LABELS, QUALITY_LABELS, SAVE_PATHS, MediaType, Quality
from app.models.torrent import TorrentStatus, format_size
from app.services.magnet_parser import MagnetInfo, MagnetParseError, parse_magnet
from app.services.torrent_service import TorrentService

logger = logging.getLogger(__name__)

router = Router(name="magnet")


class TorrentWizard(StatesGroup):
    choosing_category = State()
    choosing_quality = State()
    confirming = State()


_QUALITY_BY_CALLBACK: dict[str, Quality] = {
    "2160": Quality.UHD_2160,
    "1080": Quality.FHD_1080,
    "720": Quality.HD_720,
    "unknown": Quality.UNKNOWN,
}

INVALID_MAGNET_TEXT = (
    "❌ Не удалось распознать magnet-ссылку.\n\n"
    "Отправьте ссылку вида:\n"
    "magnet:?xt=urn:btih:..."
)


def _info_text(info: MagnetInfo) -> str:
    name = info.display_name or "не определено"
    return (
        "🔗 Torrent найден\n\n"
        f"🎬 Название:\n{name}\n\n"
        f"🔑 Hash:\n{info.info_hash[:10]}...\n\n"
        f"📡 Trackers:\n{len(info.trackers)}\n\n"
        f"💾 Размер:\n{format_size(info.size_bytes)}\n\n"
        "Куда сохранить?"
    )


def _summary_text(
    info: MagnetInfo, media_type: MediaType, quality: Quality, test_mode: bool
) -> str:
    text = (
        "✅ Готово к добавлению\n\n"
        f"🎬 Название:\n{info.display_name or 'не определяло'}\n\n"
        f"📁 Категория:\n{MEDIA_LABELS[media_type]}\n\n"
        f"🎞 Качество:\n{QUALITY_LABELS[quality]}\n\n"
        f"💾 Путь:\n{SAVE_PATHS[media_type]}\n"
    )
    if test_mode:
        text += "\n🧪 TEST MODE\n\nФайл реально скачиваться НЕ будет."
    return text


async def _edit(
    callback: CallbackQuery, text: str, keyboard: InlineKeyboardMarkup
) -> None:
    # намеренная дупликация с downloads.py — отдельный файл утилит вне скоупа
    if not isinstance(callback.message, Message):
        return
    try:
        await callback.message.edit_text(text, reply_markup=keyboard)
    except TelegramBadRequest:
        pass


@router.message(F.text)
async def on_text(message: Message, state: FSMContext) -> None:
    try:
        info = parse_magnet(message.text or "")
    except MagnetParseError:
        user_id = message.from_user.id if message.from_user else 0
        logger.info("invalid magnet from user_id=%s", user_id)
        await state.clear()
        await message.answer(INVALID_MAGNET_TEXT)
        return
    await state.set_state(TorrentWizard.choosing_category)
    await state.update_data(magnet=info)
    await message.answer(_info_text(info), reply_markup=category_keyboard())


@router.callback_query(
    TorrentWizard.choosing_category, F.data.startswith("torrent:category:")
)
async def on_category(callback: CallbackQuery, state: FSMContext) -> None:
    value = (callback.data or "").rsplit(":", 1)[-1]
    try:
        media_type = MediaType(value)
    except ValueError:
        await callback.answer("Неизвестная категория", show_alert=True)
        return
    await state.update_data(media_type=media_type)
    await state.set_state(TorrentWizard.choosing_quality)
    await callback.answer()
    await _edit(callback, "Выберите качество:", quality_keyboard())


@router.callback_query(
    TorrentWizard.choosing_quality, F.data.startswith("torrent:quality:")
)
async def on_quality(
    callback: CallbackQuery, state: FSMContext, test_mode: bool
) -> None:
    value = (callback.data or "").rsplit(":", 1)[-1]
    quality = _QUALITY_BY_CALLBACK.get(value)
    if quality is None:
        await callback.answer("Неизвестное качество", show_alert=True)
        return
    data = await state.get_data()
    info: MagnetInfo = data["magnet"]
    media_type: MediaType = data["media_type"]
    await state.update_data(quality=quality)
    await state.set_state(TorrentWizard.confirming)
    await callback.answer()
    await _edit(
        callback, _summary_text(info, media_type, quality, test_mode), confirm_keyboard()
    )


@router.callback_query(TorrentWizard.confirming, F.data == "torrent:confirm:add")
async def on_confirm_add(
    callback: CallbackQuery, state: FSMContext, torrent_service: TorrentService
) -> None:
    data = await state.get_data()
    info: MagnetInfo = data["magnet"]
    media_type: MediaType = data["media_type"]
    quality: Quality = data["quality"]
    task, created = await torrent_service.create(
        user_id=callback.from_user.id,
        magnet=info,
        media_type=media_type,
        quality=quality,
    )
    await state.clear()
    if task.status is TorrentStatus.FAILED:
        text = "⚠️ Не удалось добавить загрузку. Попробуйте позже."
    elif created:
        text = (
            f"✅ Добавлено в очередь\n\n#{task.id:03d} {task.name}\n\n"
            "Смотрите /downloads"
        )
    else:
        text = (
            f"ℹ️ Такая загрузка уже есть\n\n#{task.id:03d} {task.name}\n\n"
            "Смотрите /downloads"
        )
    await callback.answer()
    await _edit(callback, text, main_menu_keyboard())


@router.callback_query(TorrentWizard.confirming, F.data == "torrent:confirm:back")
async def on_confirm_back(callback: CallbackQuery, state: FSMContext) -> None:
    data = await state.get_data()
    info: MagnetInfo = data["magnet"]
    # «Изменить» → назад к экрану категории (интерпретация №4)
    await state.set_state(TorrentWizard.choosing_category)
    await callback.answer()
    await _edit(callback, _info_text(info), category_keyboard())


@router.callback_query(F.data == "torrent:confirm:cancel")
async def on_cancel(callback: CallbackQuery, state: FSMContext) -> None:
    if await state.get_state() is None:
        await callback.answer()
        return
    await state.clear()
    await callback.answer()
    await _edit(callback, "❌ Добавление отменено.", main_menu_keyboard())
```

- [ ] **Шаг 2: Проверки**

```bash
ruff check . && ruff format . && mypy app && pytest
```
Expected: PASS (pytest — только существующие unit-тесты; полный flow появится в T18).

**Acceptance Criteria:**
- [ ] После magnet — блок информации (название, hash, trackers, размер) + вопрос «Куда сохранить?» (FEAT-003 AC).
- [ ] 4 кнопки качества, значения маппятся в enum, «Не определять» → `UNKNOWN` (FEAT-004 AC).
- [ ] Сводка содержит все 4 поля + TEST MODE notice (только при `test_mode=True`), путь по категории (FEAT-005 AC).
- [ ] `confirm:add` → `service.create`; дубликат → сообщение «уже есть» + `main_menu` (FEAT-018 AC).
- [ ] `confirm:cancel` сбрасывает FSM; `confirm:back` возвращает к категории (FEAT-005 AC).
- [ ] Невалидный magnet → дружелюбное сообщение + `state.clear()`; unknown callback не роняет бота (FEAT-002 Error Handling).
- [ ] Бизнес-логики в хендлере нет — только рендер/FSM/вызов `service.create`.

---

### Task 15: `app/handlers/downloads.py` — FEAT-007/008/009

**Files:**
- Create: `app/handlers/downloads.py`

**Depends on:** T4, T8, T11
**Interfaces:**
- Consumes: `TorrentService.list/get/pause/resume/cancel/delete` (T8), `downloads_keyboard`/`detail_keyboard` (T11), `main_menu_keyboard` (T9), `MEDIA_EMOJI`/`QUALITY_LABELS` (T3), `format_size` (T4), DI `torrent_service: TorrentService`.
- Produces: модуль `router` с хендлерами: `/downloads`, `menu:downloads`, `torrent:open:<id>`, `torrent:list:back`, `torrent:pause|resume|cancel|delete:<id>`.

**Spec:** FEAT-007 (список, рендер, кнопки, пустой список), FEAT-008 (4 операции, ошибки provider), FEAT-009 (ownership перед каждой операцией, «не найдено» вместо раскрытия), FEAT-006 UI (детальный вид с баром/TEST MODE).

- [ ] **Шаг 1: Создать `app/handlers/downloads.py`**

```python
from __future__ import annotations

import logging
from collections.abc import Awaitable, Callable

from aiogram import F, Router
from aiogram.exceptions import TelegramBadRequest
from aiogram.filters import Command
from aiogram.types import CallbackQuery, InlineKeyboardMarkup, Message

from app.keyboards.downloads import detail_keyboard, downloads_keyboard
from app.keyboards.main import main_menu_keyboard
from app.models.media import MEDIA_EMOJI, QUALITY_LABELS
from app.models.torrent import TorrentStatus, TorrentTask, format_size
from app.services.torrent_service import TorrentService

logger = logging.getLogger(__name__)

router = Router(name="downloads")

_LIST_HEADER = "📥 Мои загрузки"
_EMPTY_TEXT = "📥 Мои загрузки\n\nУ вас пока нет загрузок."

_STATUS_LINE: dict[TorrentStatus, str] = {
    TorrentStatus.QUEUED: "🕓 В очереди",
    TorrentStatus.PAUSED: "⏸ На паузе",
    TorrentStatus.COMPLETED: "✅ Завершено",
    TorrentStatus.FAILED: "❌ Ошибка",
    TorrentStatus.CANCELLED: "🚫 Отменено",
}

_STATUS_WORD: dict[TorrentStatus, str] = {
    TorrentStatus.QUEUED: "🕓 Queued",
    TorrentStatus.DOWNLOADING: "⏳ Downloading",
    TorrentStatus.PAUSED: "⏸ Paused",
    TorrentStatus.COMPLETED: "✅ Completed",
    TorrentStatus.FAILED: "❌ Failed",
    TorrentStatus.CANCELLED: "🚫 Cancelled",
}


def _status_line(task: TorrentTask) -> str:
    if task.status is TorrentStatus.DOWNLOADING:
        return f"⏳ {task.progress:.0f}%"
    return _STATUS_LINE[task.status]


def _progress_bar(progress: float) -> str:
    filled = int(progress / 100 * 16)
    return "█" * filled + "░" * (16 - filled) + f" {progress:.0f}%"


def _detail_text(task: TorrentTask) -> str:
    parts: list[str] = [_STATUS_WORD[task.status], "", task.name]
    if task.progress > 0 or task.status in (
        TorrentStatus.DOWNLOADING,
        TorrentStatus.COMPLETED,
    ):
        parts += ["", _progress_bar(task.progress), "", f"{task.progress:.0f}%"]
    if task.status is TorrentStatus.DOWNLOADING:
        parts += ["", "Speed: TEST", "ETA: TEST"]
    if task.size_bytes is not None:
        parts += ["", f"💾 {format_size(task.size_bytes)}"]
    parts += ["", "🧪 TEST MODE"]
    return "\n".join(parts)


async def _render_list(
    service: TorrentService, user_id: int
) -> tuple[str, InlineKeyboardMarkup]:
    tasks = await service.list(user_id)
    if not tasks:
        return _EMPTY_TEXT, main_menu_keyboard()
    blocks = [
        f"#{task.id:03d}\n"
        f"{MEDIA_EMOJI[task.media_type]} {task.name}\n"
        f"{QUALITY_LABELS[task.quality]}\n"
        f"{_status_line(task)}"
        for task in tasks
    ]
    return f"{_LIST_HEADER}\n\n" + "\n\n".join(blocks), downloads_keyboard(tasks)


async def _edit(
    callback: CallbackQuery, text: str, keyboard: InlineKeyboardMarkup
) -> None:
    # намеренная дупликация с magnet.py — отдельный файл утилит вне скоупа
    if not isinstance(callback.message, Message):
        return
    try:
        await callback.message.edit_text(text, reply_markup=keyboard)
    except TelegramBadRequest:
        pass


def _parse_id(data: str) -> int | None:
    try:
        return int(data.rsplit(":", 1)[1])
    except (IndexError, ValueError):
        return None


@router.message(Command("downloads"))
async def cmd_downloads(message: Message, torrent_service: TorrentService) -> None:
    user_id = message.from_user.id if message.from_user else 0
    text, keyboard = await _render_list(torrent_service, user_id)
    await message.answer(text, reply_markup=keyboard)


@router.callback_query(F.data == "menu:downloads")
async def on_menu_downloads(
    callback: CallbackQuery, torrent_service: TorrentService
) -> None:
    text, keyboard = await _render_list(torrent_service, callback.from_user.id)
    await callback.answer()
    await _edit(callback, text, keyboard)


@router.callback_query(F.data.startswith("torrent:open:"))
async def on_open(callback: CallbackQuery, torrent_service: TorrentService) -> None:
    task_id = _parse_id(callback.data or "")
    task = (
        await torrent_service.get(callback.from_user.id, task_id)
        if task_id is not None
        else None
    )
    if task is None:
        logger.info(
            "task %s not found for user_id=%s", task_id, callback.from_user.id
        )
        await callback.answer("Задача не найдена", show_alert=True)
        return
    await callback.answer()
    await _edit(callback, _detail_text(task), detail_keyboard(task))


@router.callback_query(F.data == "torrent:list:back")
async def on_back_to_list(
    callback: CallbackQuery, torrent_service: TorrentService
) -> None:
    text, keyboard = await _render_list(torrent_service, callback.from_user.id)
    await callback.answer()
    await _edit(callback, text, keyboard)


async def _apply_and_render(
    callback: CallbackQuery,
    torrent_service: TorrentService,
    operation: Callable[[int, int], Awaitable[TorrentTask | None]],
    success_text: str,
) -> None:
    task_id = _parse_id(callback.data or "")
    task = (
        await operation(callback.from_user.id, task_id)
        if task_id is not None
        else None
    )
    if task is None:
        # FEAT-009: чужая/несуществующая — «не найдено», без раскрытия
        await callback.answer("Задача не найдена", show_alert=True)
        return
    if task.status is TorrentStatus.FAILED:
        await callback.answer("⚠️ Ошибка операции", show_alert=True)
    else:
        await callback.answer(success_text)
    text, keyboard = await _render_list(torrent_service, callback.from_user.id)
    await _edit(callback, text, keyboard)


@router.callback_query(F.data.startswith("torrent:pause:"))
async def on_pause(callback: CallbackQuery, torrent_service: TorrentService) -> None:
    await _apply_and_render(
        callback, torrent_service, torrent_service.pause, "⏸ На паузе"
    )


@router.callback_query(F.data.startswith("torrent:resume:"))
async def on_resume(callback: CallbackQuery, torrent_service: TorrentService) -> None:
    await _apply_and_render(
        callback, torrent_service, torrent_service.resume, "▶️ Продолжено"
    )


@router.callback_query(F.data.startswith("torrent:cancel:"))
async def on_cancel_task(
    callback: CallbackQuery, torrent_service: TorrentService
) -> None:
    await _apply_and_render(
        callback, torrent_service, torrent_service.cancel, "❌ Отменено"
    )


@router.callback_query(F.data.startswith("torrent:delete:"))
async def on_delete_task(
    callback: CallbackQuery, torrent_service: TorrentService
) -> None:
    task_id = _parse_id(callback.data or "")
    removed = (
        await torrent_service.delete(callback.from_user.id, task_id)
        if task_id is not None
        else False
    )
    if not removed:
        await callback.answer("Задача не найдена", show_alert=True)
        return
    await callback.answer("🗑 Удалено")
    text, keyboard = await _render_list(torrent_service, callback.from_user.id)
    await _edit(callback, text, keyboard)
```

Поведение после pause/resume/cancel/delete: текущее сообщение перерисовывается в **список** (единый путь рендера, обновлённый статус сразу виден) — зафиксировано здесь и вручном тестировании (T19).

- [ ] **Шаг 2: Проверки**

```bash
ruff check . && ruff format . && mypy app && pytest
```
Expected: PASS.

**Acceptance Criteria:**
- [ ] `/downloads` показывает только задачи текущего `user_id` (FEAT-007/009 AC).
- [ ] Статус/прогресс по формату спецификации; кнопки соответствуют состоянию задачи (FEAT-007 AC).
- [ ] Все 4 операции (pause/resume/cancel/delete) работают; статус обновляется после операции; ownership проверяется до операции (FEAT-008/009 AC).
- [ ] Пустой список → понятное сообщение; чужая задача → «Задача не найдена» (FEAT-007/009 Error Handling).
- [ ] Handlers вызывают только `TorrentService`, не provider (ARCHITECTURE §2).

---

### Task 16: `app/bot.py` — сборка диспетчера

**Files:**
- Create: `app/bot.py`

**Depends on:** T12, T13, T14, T15
**Interfaces:**
- Consumes: `router` из всех четырёх handler-модулей; `TorrentService` (T8).
- Produces: `create_dispatcher(service: TorrentService, test_mode: bool) -> Dispatcher` — DI через kwargs конструктора (`Dispatcher(torrent_service=..., test_mode=...)`, значения автоматически инжектируются в хендлеры по имени — aiogram Dependency Injection); порядок роутеров; fallback-роутер; error handler.

**Spec:** design §Error Handling (неизвестный callback → игнор; исключение → log + ответ + бот жив); DEVELOPMENT §6 (logging); интерпретация №8 (порядок роутеров).

- [ ] **Шаг 1: Создать `app/bot.py`**

```python
from __future__ import annotations

import logging

from aiogram import Dispatcher, Router
from aiogram.types import CallbackQuery, ErrorEvent

from app.handlers import downloads, magnet, settings, start
from app.services.torrent_service import TorrentService

logger = logging.getLogger(__name__)


def _fallback_router() -> Router:
    """Последний роутер: гасит неизвестные callback (design: игнор без падения).

    Свой хендлер диспетчера тут нельзя — по propagate_event сначала проверяются
    свои хендлеры, потом sub_routers (интерпретация №8).
    """
    fallback = Router(name="fallback")

    @fallback.callback_query()
    async def unknown_callback(callback: CallbackQuery) -> None:
        logger.debug("ignored unknown callback data=%r", callback.data)
        await callback.answer()

    return fallback


def create_dispatcher(service: TorrentService, test_mode: bool) -> Dispatcher:
    # порядок обязателен: командные роутеры ДО magnet (иначе /downloads съест F.text)
    dp = Dispatcher(torrent_service=service, test_mode=test_mode)
    dp.include_router(start.router)
    dp.include_router(downloads.router)
    dp.include_router(settings.router)
    dp.include_router(magnet.router)
    dp.include_router(_fallback_router())

    @dp.error()
    async def on_error(event: ErrorEvent) -> None:
        logger.error(
            "unhandled error: %s", event.exception, exc_info=event.exception
        )
        update = event.update
        try:
            if update.message is not None:
                await update.message.answer(
                    "⚠️ Произошла внутренняя ошибка. Попробуйте ещё раз."
                )
            elif update.callback_query is not None:
                await update.callback.answer("⚠️ Ошибка", show_alert=True)
        except Exception:
            logger.error("failed to send error reply", exc_info=True)

    return dp
```

- [ ] **Шаг 2: Проверки** — `ruff check . && ruff format . && mypy app && pytest` → PASS.

**Acceptance Criteria:**
- [ ] `torrent_service` и `test_mode` доступны всем хендлерам через инъекцию по имени.
- [ ] Неизвестный callback не падает и не висит в спиннере (fallback отвечает `callback.answer()`).
- [ ] Любое исключение в хендлере логируется, пользователю уходит ответ, бот продолжает работать (design §Error Handling).

---

### Task 17: `app/main.py` — точка входа

**Files:**
- Create: `app/main.py`

**Depends on:** T2, T7, T8, T16
**Interfaces:**
- Consumes: `load_config()` (T2), `FakeTorrentProvider` (T7), `TorrentService` (T8), `create_dispatcher` (T16).
- Produces: запуск — `python -m app.main`.

**Spec:** design §Approach (один процесс, in-memory очередь); DEVELOPMENT §6 (logging, не print).

- [ ] **Шаг 1: Создать `app/main.py`**

```python
from __future__ import annotations

import asyncio
import logging

from aiogram import Bot, Dispatcher

from app.bot import create_dispatcher
from app.config import load_config
from app.providers.fake import FakeTorrentProvider
from app.services.torrent_service import TorrentService

logger = logging.getLogger(__name__)


async def main() -> None:
    logging.basicConfig(
        level=logging.INFO,
        format="%(asctime)s %(levelname)s %(name)s: %(message)s",
    )
    config = load_config()
    service = TorrentService(FakeTorrentProvider())
    dp: Dispatcher = create_dispatcher(service, config.test_mode)
    bot = Bot(token=config.bot_token)
    logger.info("bot started (TEST_MODE=%s)", config.test_mode)
    await dp.start_polling(bot)


if __name__ == "__main__":
    asyncio.run(main())
```

- [ ] **Шаг 2: Проверки** — `ruff check . && ruff format . && mypy app && pytest` → PASS.
- [ ] **Шаг 3 (Smoke, без Telegram):** `python -c "from app.main import main; print('ok')"` → `ok` (импорты/сборка без ошибок; сетевой запуск — в T19 вручную).

**Acceptance Criteria:**
- [ ] Один процесс: config → fake provider → service → dispatcher → polling.
- [ ] Запуск — `python -m app.main`; токен только из env.

---

### Task 18: `tests/integration/test_bot_flow.py`

**Files:**
- Create: `tests/integration/test_bot_flow.py`

**Depends on:** T17
**Interfaces:**
- Consumes: `create_dispatcher` (T16), `TorrentService` + `FakeTorrentProvider` (T7/T8), `TorrentStatus` (T4); aiogram `Dispatcher.feed_raw_update` + собственный `RecordingSession` (перехват всех API-вызовов, никакой сети — aiogram не поставляет MockedBot в wheel, `tests/mocked_bot.py` живёт только в репозитории aiogram).
- Produces: 3 теста — полный flow, невалидный magnet, неизвестный callback.

**Spec:** TESTING §Integration tests (полный flow `/start → magnet → category → quality → confirmation → queue → /downloads`); FEAT-001 Tests; FEAT-006 Tests (confirmation → queue); дизайн §Testing Strategy.

- [ ] **Шаг 1: Написать `tests/integration/test_bot_flow.py`**

```python
from __future__ import annotations

from collections.abc import AsyncGenerator
from typing import Any

from aiogram import Bot, Dispatcher
from aiogram.client.session.base import BaseSession
from aiogram.methods import AnswerCallbackQuery, EditMessageText, SendMessage

from app.bot import create_dispatcher
from app.models.media import MediaType, Quality
from app.models.torrent import TorrentStatus
from app.providers.fake import FakeTorrentProvider
from app.services.torrent_service import TorrentService

MAGNET = (
    "magnet:?xt=urn:btih:a1b2c3d4e5f60718293a4b5c6d7e8f9012345678"
    "&dn=The.Matrix.1999.1080p.BluRay"
    "&tr=udp%3A%2F%2Ftracker.example.org%3A1337%2Fannounce"
    "&tr=udp%3A%2F%2Fopen.demonii.com%3A1337%2Fannounce"
    "&xl=12884901888"
)


class RecordingSession(BaseSession):
    """Перехватывает вызовы Telegram API: тесты ходят в сеть только через этот объект."""

    def __init__(self) -> None:
        super().__init__()
        self.methods: list[Any] = []

    async def close(self) -> None:
        return None

    async def make_request(
        self, bot: Bot, method: Any, timeout: int | None = None
    ) -> Any:
        self.methods.append(method)
        return None

    async def stream_content(
        self,
        url: str,
        headers: dict[str, Any] | None = None,
        timeout: int = 30,
        chunk_size: int = 65536,
        raise_for_status: bool = True,
    ) -> AsyncGenerator[bytes, None]:
        yield b""


def _message_update(update_id: int, text: str) -> dict[str, Any]:
    return {
        "update_id": update_id,
        "message": {
            "message_id": update_id,
            "date": 1790000000,
            "chat": {"id": 1, "type": "private"},
            "from": {"id": 1, "is_bot": False, "first_name": "Tester"},
            "text": text,
        },
    }


def _callback_update(update_id: int, data: str) -> dict[str, Any]:
    return {
        "update_id": update_id,
        "callback_query": {
            "id": str(update_id),
            "from": {"id": 1, "is_bot": False, "first_name": "Tester"},
            "chat_instance": "test-chat-instance",
            "data": data,
            "message": {
                "message_id": 100,
                "date": 1790000000,
                "chat": {"id": 1, "type": "private"},
                "from": {"id": 42, "is_bot": True, "first_name": "Bot"},
                "text": "ui",
            },
        },
    }


def _last_text(session: RecordingSession) -> str:
    for method in reversed(session.methods):
        text = getattr(method, "text", None)
        if isinstance(text, str):
            return text
    return ""


def _last_buttons(session: RecordingSession) -> list[tuple[str, str | None]]:
    for method in reversed(session.methods):
        markup = getattr(method, "reply_markup", None)
        if markup is not None and hasattr(markup, "inline_keyboard"):
            return [
                (button.text, button.callback_data)
                for row in markup.inline_keyboard
                for button in row
            ]
    return []


def _setup() -> tuple[Dispatcher, RecordingSession, Bot, TorrentService]:
    service = TorrentService(FakeTorrentProvider())
    dp = create_dispatcher(service=service, test_mode=True)
    session = RecordingSession()
    bot = Bot(token="42:TEST", session=session)
    return dp, session, bot, service


async def _wizard_to_confirm(
    dp: Dispatcher, bot: Bot, session: RecordingSession, start_id: int
) -> int:
    uid = start_id
    await dp.feed_raw_update(bot, _message_update(uid, MAGNET))
    uid += 1
    assert "🔗 Torrent найден" in _last_text(session)

    await dp.feed_raw_update(bot, _callback_update(uid, "torrent:category:movie"))
    uid += 1
    assert "Выберите качество" in _last_text(session)

    await dp.feed_raw_update(bot, _callback_update(uid, "torrent:quality:1080"))
    uid += 1
    assert "Готово к добавлению" in _last_text(session)
    assert "🧪 TEST MODE" in _last_text(session)
    assert "💾 Путь:\nMovies/" in _last_text(session)
    return uid


async def test_full_flow_start_to_downloads() -> None:
    dp, session, bot, service = _setup()

    # 1. /start (FEAT-001)
    await dp.feed_raw_update(bot, _message_update(1, "/start"))
    assert "🤖 Torrent Bot" in _last_text(session)
    assert _last_buttons(session) == [
        ("📥 Мои загрузки", "menu:downloads"),
        ("⚙️ Настройки", "menu:settings"),
        ("ℹ️ Помощь", "menu:help"),
    ]

    # 2-4. magnet -> категория -> качество (FEAT-003/004/005)
    next_id = await _wizard_to_confirm(dp, bot, session, 2)

    # 5. подтверждение -> очередь (FEAT-006)
    await dp.feed_raw_update(bot, _callback_update(next_id, "torrent:confirm:add"))
    next_id += 1
    assert "✅ Добавлено в очередь" in _last_text(session)
    tasks = await service.list(1)
    assert len(tasks) == 1
    assert tasks[0].status is TorrentStatus.QUEUED
    assert tasks[0].media_type is MediaType.MOVIE
    assert tasks[0].quality is Quality.FHD_1080
    assert tasks[0].progress == 0.0

    # 6. /downloads (FEAT-007)
    await dp.feed_raw_update(bot, _message_update(next_id, "/downloads"))
    next_id += 1
    text = _last_text(session)
    assert "#001" in text
    assert "The.Matrix.1999.1080p.BluRay" in text
    assert "1080p" in text

    # 7. повторный magnet -> дубликат (FEAT-018)
    next_id = await _wizard_to_confirm(dp, bot, session, next_id)
    await dp.feed_raw_update(bot, _callback_update(next_id, "torrent:confirm:add"))
    assert "уже есть" in _last_text(session)
    assert len(await service.list(1)) == 1


async def test_invalid_magnet_gets_friendly_error() -> None:
    dp, session, bot, _service = _setup()
    await dp.feed_raw_update(bot, _message_update(1, "просто обычный текст"))
    assert "Не удалось распознать" in _last_text(session)


async def test_unknown_callback_is_ignored() -> None:
    dp, session, bot, _service = _setup()
    await dp.feed_raw_update(bot, _callback_update(1, "foo:bar"))
    assert isinstance(session.methods[-1], AnswerCallbackQuery)
```

Замечания:
- `session.methods` содержит `SendMessage`/`EditMessageText`/`AnswerCallbackQuery` — ассерты по тексту и кнопкам идут от них; реальный network I/O отсутствует полностью.
- FSM `MemoryStorage` живёт внутри `dp` между `feed_raw_update` — шаги wizard проходят по одному user/chat.

- [ ] **Шаг 2: Запустить**

```bash
pytest tests/integration/test_bot_flow.py -v
```
Expected: 3 passed.

- [ ] **Шаг 3: Полный прогон + лint/типы**

```bash
ruff check . && ruff format . && mypy app && pytest
```
Expected: PASS.

**Acceptance Criteria:**
- [ ] Полный flow `/start → magnet → category → quality → confirmation → queue → /downloads` проходит через реальные роутеры/FSM (TESTING §Integration tests).
- [ ] Дубликат на UI показывает существующую задачу, очередь не растёт.
- [ ] Невалидный magnet и неизвестный callback не роняют диспетчер.
- [ ] Ни одного сетевого вызова (все API-вызовы перехвываются `RecordingSession`).

---

### Task 19: `docs/TESTING.md` — «Ручное тестирование в Telegram» + синхронизация доков

**Files:**
- Modify: `docs/TESTING.md` (добавить раздел, обновить дерево `tests/`)
- Modify: `docs/API.md` §4 (дописать callbacks)
- Modify: `docs/features/001…009` (Status → Implemented)

**Depends on:** T18
**Interfaces:**
- Consumes: результаты T1–T18.
- Produces: пошаговая инструкция ручного теста в Telegram.

**Spec:** дизайн §Testing Strategy («Manual: пошаговая инструкция тестирования в Telegram → docs/TESTING.md, раздел „Ручное тестирование“»); DEVELOPMENT §8/§9 (обновление документации и feature-файлов).

- [ ] **Шаг 1: Дописать в конец `docs/TESTING.md` раздел:**

````markdown
## Ручное тестирование в Telegram

Цель: проверить весь UI-flow (FEAT-001–009) в реальном Telegram при `TEST_MODE=true`.
Ни одна из операций не качает реальные торренты и не трогает файлы/сеть за пределами Telegram Bot API.

### Подготовка

1. Откройте [@BotFather](https://t.me/BotFather) → `/newbot` → задайте имя → получите токен вида `123456789:AA...`.
2. Создайте локальный `.env` (`.env` в git не попадает):
   ```bash
   cp .env.example .env
   # отредактируйте .env: BOT_TOKEN=<ваш токен>, TEST_MODE=true
   ```
3. Установите зависимости и запустите бота:
   ```bash
   python3 -m venv .venv && source .venv/bin/activate
   pip install -r requirements.txt
   python -m app.main
   ```
4. В терминале появится `bot started (TEST_MODE=True)` — оставьте процесс запущенным.

### Сценарий прохождения

Выполняйте шаги в указанном порядке в чате с ботом.

1. **Старт.** Отправьте `/start`.
   Ожидаемо: `🤖 Torrent Bot` + подпись «Отправьте magnet-ссылку…» и 3 кнопки: `📥 Мои загрузки`, `⚙️ Настройки`, `ℹ️ Помощь`.
2. **Заглушки меню.** Нажмите `⚙️ Настройки` → тост «Настройки появятся позже»; `ℹ️ Помощь` → тост с подсказкой. Бот не падает.
3. **Получение magnet (FEAT-002/003).** Отправьте одной строкой:
   ```text
   magnet:?xt=urn:btih:a1b2c3d4e5f60718293a4b5c6d7e8f9012345678&dn=The.Matrix.1999.1080p.BluRay&tr=udp%3A%2F%2Ftracker.example.org%3A1337%2Fannounce&tr=udp%3A%2F%2Fopen.demonii.com%3A1337%2Fannounce&xl=12884901888
   ```
   Ожидаемо: экран `🔗 Torrent найден` — Название `The.Matrix.1999.1080p.BluRay`, Hash `a1b2c3d4e5...`, Trackers `2`, Размер `12.0 GiB`, вопрос «Куда сохранить?», кнопки `🎬 Фильм` / `📺 Сериал` / `❌ Отмена`.
4. **Отмена (FEAT-003).** Нажмите `❌ Отмена` → «❌ Добавление отменено», меню из п.1. Повторите п.3.
5. **Категория → качество (FEAT-003/004).** Нажмите `🎬 Фильм` → «Выберите качество:» с кнопками `2160p / 4K`, `1080p`, `720p`, `Не определять`. Нажмите `1080p`.
6. **Сводка (FEAT-005).** Ожидаемо: `✅ Готово к добавлению`, поля Название/Категория `Фильм`/Качество `1080p`/Путь `Movies/`, блок `🧪 TEST MODE` «Файл реально скачиваться НЕ будет», кнопки `⬇️ Добавить в очередь`, `✏️ Изменить`, `❌ Отмена`.
7. **«Изменить».** Нажмите `✏️ Изменить` → вернулись на экран «Куда сохранить?». Выберите `📺 Сериал` → `720p` → в сводке Категория `Сериал`, Путь `TV/`. Нажмите `✏️ Изменить`, выберите обратно `🎬 Фильм` → `1080p`.
8. **Постановка в очередь (FEAT-006).** Нажмите `⬇️ Добавить в очередь`.
   Ожидаемо: «✅ Добавлено в очередь / #001 The.Matrix… / Смотрите /downloads» + главное меню. В логе терминала — `created task_id=1 ...`.
9. **Дубликат (FEAT-018).** Отправьте тот же magnet (п.3), пройдите wizard до подтверждения и нажмите `⬇️ Добавить в очередь`.
   Ожидаемо: «ℹ️ Такая загрузка уже есть / #001 …». В очереди по-прежнему **одна** задача (проверьте п.10).
10. **Список (FEAT-007).** Отправьте `/downloads`.
    Ожидаемо: `📥 Мои загрузки`, блок `#001` с `🎬 The.Matrix…`, `1080p`, статусом `🕓 В очереди`, и кнопки `#001 📊 Открыть`, `#001 ⏸ Пауза`, `#001 ❌ Отменить`.
11. **Симуляция прогресса (FEAT-006).** Повторяйте `/downloads` 3–8 раз.
    Ожидаемо: статус меняется `🕓 В очереди` → `⏳ 12%` → `⏳ 25%` → … (значения растут) → `✅ Завершено`; у завершённой задачи кнопки `#001 📊 Открыть` и `#001 🗑 Удалить`. Терминал: только логи сервиса, никаких обращений к qBittorrent/NAS.
12. **Детальный вид (FEAT-006/007).** Нажмите `#001 📊 Открыть`.
    Ожидаемо: строка статуса (`⏳ Downloading` и т.п.), название, прогресс-бар вида `█████░░░░░░░░░░ 35%`, строка процента, `Speed: TEST`, `ETA: TEST`, `🧪 TEST MODE`. Кнопки задачи + `◀️ К списку`.
13. **Назад.** Нажмите `◀️ К списку` → снова список.
14. **Пауза/продолжить (FEAT-008).** Для активной задачи: `⏸ Пауза` → тост «⏸ На паузе», в списке `⏸ На паузе`, кнопки сменяются на `▶️ Продолжить`; `▶️ Продолжить` → прогресс снова растёт.
15. **Отмена (FEAT-008).** `❌ Отменить` → тост «❌ Отменено», статус `🚫 Отменено`, кнопки: `📊 Открыть` + `🗑 Удалить`.
16. **Удаление (FEAT-008).** `🗑 Удалить` → тост «🗑 Удалено», задача исчезает. Удалите все задачи → `/downloads` показывает «У вас пока нет загрузок».
17. **Невалидный ввод (FEAT-002).** Отправьте `просто обычный текст`.
    Ожидаемо: «❌ Не удалось распознать magnet-ссылку…», бот жив, `/start` работает.
18. **Неизвестный/протухший callback (design §Error Handling).** Нажмите любую кнопку в **старом** сообщении (например, «Фильм» после завершения wizard). Ожидаемо: кнопка просто гаснет, ошибок в логе нет, бот продолжает отвечать.
19. **Изоляция пользователей (FEAT-009).** Откройте бота со второго Telegram-аккаунта (или попросите второго пользователя): `/downloads` → «У вас пока нет загрузок»; задачи первого пользователя не видны и недоступны.
20. **Остановка.** `Ctrl+C` в терминале — бот корректно останавливается.

### Критерий успеха

- Все 20 шагов пройдены без ошибок в терминале и без падений бота.
- В логах отсутствуют любые упоминания qBittorrent, NAS, записи файлов (TEST_MODE).
- Токен остался только в `.env` и нигде не залогирован.
````

- [ ] **Шаг 2: Обновить дерево тестов в шапке `docs/TESTING.md`** — заменить блок дерева на итоговое:

```text
tests/
├── unit/
│   ├── test_magnet_parser.py
│   ├── test_fake_provider.py
│   └── test_torrent_service.py
├── integration/
│   └── test_bot_flow.py
└── fixtures/
```

- [ ] **Шаг 3: Дописать callbacks в `docs/API.md` §4** (таблица «Callback Data Contract»), строки:

```text
| `torrent:quality:unknown` | «Не определять» (Quality.UNKNOWN) |
| `torrent:open:<id>` | открыть детальный вид задачи |
| `torrent:list:back` | вернуться к списку загрузок |
| `torrent:pause:<id>` / `torrent:resume:<id>` | пауза / продолжить |
| `torrent:cancel:<id>` / `torrent:delete:<id>` | отменить / удалить из списка |
| `menu:downloads` / `menu:settings` / `menu:help` | пункты главного меню (FEAT-001) |
```

- [ ] **Шаг 4: Перевести features в статус Implemented** — в `docs/features/001-start-menu.md` … `009-user-isolation.md` заменить в секции `Status` значение `Planned` → `Implemented`, и отметить `[x]` в `Implementation Tasks` / `Acceptance Criteria` каждого файла (только те пункты, которые действительно покрыты тестами из Testing Matrix).

- [ ] **Шаг 5: Финальный гейт**

```bash
ruff check .
ruff format --check .
mypy app
pytest
```
Expected: все четыре команды exit code 0.

**Acceptance Criteria:**
- [ ] В `docs/TESTING.md` есть раздел «Ручное тестирование в Telegram» с подготовкой (BotFather → `.env` → `python -m app.main`) и пошаговым проходом **каждого** экрана UI: `/start`, magnet-инфо, категория, качество, сводка+TEST MODE, очередь, `/downloads`, детальный вид, pause/resume/cancel/delete, дубликат, ошибки.
- [ ] `docs/API.md` §4 отражает все реально использованные callbacks.
- [ ] Features 001–009 — `Implemented`.
- [ ] Финальный гейт: `ruff check .`, `ruff format --check .`, `mypy app`, `pytest` — успешно.

---

## Финальная верификация (DoD)

```bash
ruff check .
ruff format --check .
mypy app
pytest
```

- [ ] Все четыре команды завершаются успешно (DEVELOPMENT §9).
- [ ] Существующие тесты не удалялись (DEVELOPMENT §3.8).
- [ ] Секретов в файлах нет: `.env` отсутствует в выдаче, токен только в `.env.example` как placeholder (DEVELOPMENT §10).
- [ ] Scope-check: в коде нет импортов `aiohttp`/`httpx`/`qbittorrent`/`smb` вне aiogram, нет `open(`/`Path.write` в `app/` (TEST_MODE, DEVELOPMENT §2).
