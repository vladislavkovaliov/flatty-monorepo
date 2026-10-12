# Перенос torrent-download-bot в flatty-budget-monorepo

date: 2026-10-12
topic: "Migration of torrent-download-bot into apps/ of flatty-budget-monorepo"
status: draft

## Problem Statement

Перенести standalone-репозиторий `torrent-download-bot` (Python 3.12, aiogram + FastAPI + PostgreSQL) в `apps/torrent-download-bot` внутри Nx-монорепо `flatty-budget-monorepo`. Всё сопровождение приложения — `docs/`, `thoughts/`, `init.md` — переезжает вместе с ним и остаётся закоммиченным.

## Constraints

- Монорепо — npm workspaces + Nx 23; все текущие apps: JS/TS (React, NestJS) или Go (`go-api`). Python-app первой — прецедент есть в `go-api` (нет `package.json`, есть `project.json` на `nx:run-commands`).
- В `apps/*` npm workspaces папки без `package.json` игнорируются — `go-api` и `nginx-proxy` уже так живут. Python-app не сломает `npm install`.
- Корневой `.gitignore` монорепо содержит строку `thoughts` — она игнорирует **любой** каталог `thoughts` на любом уровне. Закоммиченные в боте `thoughts/` (designs/plans/ledgers, 9 файлов) иначе станут untracked. Нужна negation-строка.
- Корневой `.gitignore` уже покрывает `.env` / `.env.*` с исключением `!apps/*/.env.example` — шаблон бота переедет без правок gitignore.
- Порт **8080 занят** `go-api`. HTTP API бота (Swagger, `API_PORT`) по умолчанию слушает 8080 — конфликт, нужен новый порт.
- Postgres монорепо: `localhost:55000` (compose-mapping), БД `flatty`. Боту рекомендована **отдельная БД** `torrent_bot` (docs бота: design §Documentation, `.env.example`).
- `docker compose` в монорепо — только инфраструктура; приложения в dev бегут на хосте через Nx. Бот остаётся host-процессом, compose не трогаем (кроме, опц., создания БД).
- Skills: у бота **нет** каталога `skills/` (проверено по корню и по истории) — мигрировать под этим именем нечего. Монорепо свои skills держит в `.claude/skills` (Go/React-oriented) — не трогаем.
- Старый GitHub-репозиторий `vladislavkovaliov/torrent-download-bot` остаётся архивом с полной историей; новый код живёт только в монорепо.

## Approach

**Чистое копирование tracked-файлов** в `apps/torrent-download-bot/` без git-subtree/filter-repo.

Альтернативы отклонены:

- `git subtree add` — сохраняет историю, но засоряет монорепо foreign-историей и merge-коммитами; для бота с 5 последними коммитами и малой историей выгода ≈ 0.
- git submodule — двойной репозиторий, двойная когнитивная нагрузка; пользователь явно хочет apps-папку, а не submodule.
- История не нужна в монорепо: она доступна на GitHub старого репо, которое архивируется, а не удаляется.

Колокация: `docs/`, `thoughts/`, `init.md`, `.gitignore` остаются **внутри** `apps/torrent-download-bot/`. Прецедент: `apps/go-api/docs/` (swagger), `apps/react-launcher/README.md`, у каждого app свой `.gitignore`. Корневой `thoughts/` монорепо — его собственный, чужие артефакты туда не мержим (и он всё равно gitignored).

## Architecture

Целевая структура:

```text
flatty-budget-monorepo/
├── apps/
│   ├── go-api/                  # прецедент: non-JS app
│   └── torrent-download-bot/    # ← переезд
│       ├── app/                 # python-пакет (без правок кода)
│       ├── tests/
│       ├── docs/                # API, ARCHITECTURE, CONFIGURATION, decisions/, features/…
│       ├── thoughts/            # shared/designs, shared/plans, ledgers (остаётся tracked)
│       ├── init.md              # исходный 28KB-спекулятив
│       ├── README.md
│       ├── AGENTS.md            # ← новый, по конвенции монорепо
│       ├── project.json         # ← новый, Nx-таргеты (по образцу go-api)
│       ├── pyproject.toml
│       ├── requirements.txt
│       ├── .env.example         # ← правки: порт, DATABASE_URL монорепо
│       └── .gitignore           # ← из бота (venv, кэши, logs)
├── AGENTS.md                    # ← + строка в таблице Projects
├── .gitignore                   # ← + negation для apps/torrent-download-bot/thoughts
└── … (npm-часть не меняется)
```

## Components

### 1. Копирование файлов

Переезжают все **tracked** файлы бота (список из `git ls-files`), минус мусор:

- Копировать: `app/`, `tests/`, `docs/`, `thoughts/`, `init.md`, `README.md`, `pyproject.toml`, `requirements.txt`, `.env.example`, `.gitignore`
- **Не** копировать: `.git/`, `.venv/`, `__pycache__/`, `.mypy_cache/`, `.pytest_cache/`, `.ruff_cache/`, `.coverage`, `logs/`, `.env` (секреты! токен бота останется в старом репо/локально; в монорепо `.env` создаётся заново из `.env.example`)
- `rsync -a --exclude-from` или `git archive HEAD | tar -x` в целевую папку — надёжнее ручного `cp -r`

### 2. project.json (Nx) — по образцу go-api

`nx:run-commands` с `cwd: apps/torrent-download-bot`:

- **serve** — `.venv/bin/python -m app.main` (fail если venv не создан — понятная ошибка, bootstrap отдельно)
- **setup** — `python3 -m venv .venv && .venv/bin/pip install -r requirements.txt` (разовый, документируется в AGENTS.md)
- **test** — `.venv/bin/pytest`
- **lint** — `.venv/bin/ruff check .`
- **typecheck** — `.venv/bin/mypy app`

`tags`: `["scope:backend", "type:bot"]` (не `type:api` — это polling-бот + localhost API).

Отдельного плагена для Python в nx.json **не добавляем** — run-commands достаточно.

### 3. AGENTS.md приложения

Новый файл по шаблону `apps/go-api/AGENTS.md`: порт, стек, таблица команд `nx serve torrent-download-bot` и т.д., заметки про `TEST_MODE=true`, fail-fast конфиг, тесты без БД skip'аются.

### 4. Корневой AGENTS.md

Добавить строку в таблицу Projects:

`| torrent-download-bot | Telegram-бот торрент-загрузок | Python, aiogram, FastAPI | 8090 | [→](apps/torrent-download-bot/AGENTS.md) |`

### 5. Порты и env

Занятые порты: 80 (nginx), 3000/3001 (nest), 8080 (go-api), 8081 (settings), 8082 (resident + redpanda console), 9000 (launcher), 55000 (postgres host-mapping).

**Боту отдаём 8090.** Правки `.env.example`:

```text
BOT_TOKEN=123456:REPLACE_WITH_TOKEN_FROM_BOTFATHER
TEST_MODE=true
DATABASE_URL=postgresql://postgres:password@localhost:55000/torrent_bot?sslmode=disable
LOG_LEVEL=INFO
API_HOST=127.0.0.1
API_PORT=8090
API_USER_ID=1
```

Никаких изменений в коде `config.py` не нужно — всё уже env-driven. Root `.env.dev`/`.env.prod` не трогаем: compose-переменные бота не нужны (не dockerized).

### 6. БД torrent_bot

- Добавить в `db.sql` (или в начало `scripts/init-db.sh`) создание БД: `CREATE DATABASE torrent_bot;` — инициализация через существующий `init-db` контейнер.
- Схему бота накатывает сам (`app/db/schema.sql` при старте — проверить при выполнении; если миграция ручная, задокументировать в AGENTS.md).

### 7. thoughts — снять gitignore-конфликт

В корневой `.gitignore` рядом со строкой `thoughts` добавить исключение:

```gitignore
thoughts
!apps/torrent-download-bot/thoughts
```

Паттерн `thoughts` матчит только пути с basename `thoughts`; один negation на каталог достаточно, содержимое внутри матчится заново. Проверка после переноса: `git status` в монорепо показывает 9 файлов thoughts как tracked/добавленные.

### 8. Судьба старого репозитория

- Локальную копию `~/temp/torrent-download-bot` **не удалять** — она остаётся рабочей до верификации переезда.
- На GitHub: Archive repository (ручная операция пользователя, инструкция в README-секции выполнения).
- Секрет `.env` старого репо (BOT_TOKEN) не переносить; пользователь создаёт `apps/torrent-download-bot/.env` заново.

## Data Flow

Не меняется. Приложение копируется целиком, внутренний флоу handlers → services → repositories/providers не затрагивается.

Изменяется только окружение запуска:

```text
(nx serve torrent-download-bot)
        │
        ▼
apps/torrent-download-bot/.venv/bin/python -m app.main
        │  load_dotenv()  → apps/torrent-download-bot/.env
        ▼
Telegram polling (aiogram) + FastAPI на :8090 ──► postgres localhost:55000 / torrent_bot
```

## Error Handling

Миграция — разовая операция; стратегия отката:

- Старый репозиторий не трогаем до полной верификации → откат = продолжать работать в старом месте.
- Копирование делаем `rsync`/`git archive` (идемпотентно, можно повторить).
- Проверки после копирования: `git status` чистый по мусору, `nx show project torrent-download-bot` видит проект, `nx test` / `nx lint` проходят, бот стартует с новым `.env`.
- Конфликт портов ловим на старте (Address already in use) —就是因为 этого порт меняется заранее, а не по факту.

## Testing Strategy

Приложение уже имеет 49 тестов (unit + integration, DB-тесты skip'аются без PostgreSQL). Верификация переезда:

1. `git ls-files` старого репо == файлы в `apps/torrent-download-bot` (кроме исключённого мусора)
2. `npm install` в корне монорепо не ломается (workspaces игнорирует app без package.json)
3. `nx show project torrent-download-bot` — проект виден Nx
4. `nx lint torrent-download-bot` / `nx typecheck torrent-download-bot` — зелёные
5. `nx test torrent-download-bot` — 49 тестов, skip DB при недоступной БД
6. `docker compose up postgres` + создать torrent_bot + `apps/torrent-download-bot/.env` → `nx serve torrent-download-bot` — бот отвечает в Telegram, Swagger на `http://127.0.0.1:8090/docs`
7. `git status` монорепо: thoughts бота закоммичены (не игнорируются), мусора нет

## Open Questions

- **История git**: пользователь не просил сохранять — решение: не сохранять, архивировать старый репо. Если захочет историю внутри монорепо — это отдельная задача subtree.
- **Skills**: каталога skills у бота нет. Если под «skill» имелись в виду `.claude/skills` монорепо — они уже общие, ничего не мигрируем.
- **CI**: в монорепо есть `.github/workflows` — добавлять ли job для Python-app? По умолчанию **нет** (YAGNI; в workflows старого репо тоже не было настройки CI — уточнить при выполнении, если workflows бота существуют).
