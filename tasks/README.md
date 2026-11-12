# Задачи: LLM-агент над данными расходов

Учебный трек из трёх задач. Цель — руками пройти путь «данные → инструменты (MCP) →
агент с tool-calling → проверка (evals)» и при этом прокачать Python.

Темы:

1. **LLM + RAG** (в нашем варианте — agentic tool-calling вместо классического RAG).
2. **Свой MCP-сервер** над данными расходов.
3. **Evals** — как проверять, что агент отвечает правильно.

Решение — **Python** (для практики языка). Всё живёт в одном новом Nx-приложении
`apps/python-expense-agent`, по конвенциям `apps/python-torrent-bot`.

> Задачи написаны как упражнения: в них есть цели, требования, критерии готовности и
> подсказки, но **нет готового решения**. Код пишешь ты — иначе смысла нет.

---

## Данные проекта

В БД `flatty` (PostgreSQL, порт `55000`) есть два «слоя» данных:

- **`expenses`** — отдельные траты:
  `id`, `user_id`, `resident_location_id`, `category_id`, `amount NUMERIC(12,2)`,
  `description`, `month` (1–12), `year`, `created_at`, `updated_at`.
- **spending (агрегаты)**:
  - `expense_monthly_totals` — `total_spent` по `(resident_location_id, month, year)`;
  - `expense_monthly_averages` — `average_amount`, `expense_count` по тому же ключу.
- **`categories`** — справочник категорий (`utilities`, `additional`, `electricity`,
  `internet_tv`, `mts_1`, `mts_2`, `parking_milya`, `parking_kamova`).
- **`resident_locations`** — объекты/квартиры, к которым привязаны траты.

Подключение по умолчанию:

```
DATABASE_URL=postgres://postgres:password@localhost:55000/flatty?sslmode=disable
```

Данные можно поднять локально:

```bash
podman-compose -f docker-compose.yml --env-file .env.dev up -d postgres init-db
```

## Инфраструктура LLM

Ollama уже поднят в сети: `http://192.168.1.85:11434`.

Для tool-calling нужна модель, которая умеет `tools`. Подойдут `qwen2.5:7b`,
`llama3.1:8b`, `mistral`. Модель `qwen2.5-coder` для tool-calling хуже — она про код.
Модель задаём через `.env`, чтобы можно было менять без правок кода.

## Маршрут

Базовый трек (обязательный):

```
01-mcp-expense-server.md   →  02-http-agent-bot.md  →  03-evals-harness.md
   (инструменты)                 (агент поверх них)        (проверка агента)
```

Рекомендую идти строго по порядку: каждая задача опирается на предыдущую.

Дополнительные задачи (по интересу, после базы):

| Файл | Тема | Опирается на |
|---|---|---|
| `04-rag-tool.md` | RAG как MCP-инструмент: семантический поиск по `description` | 01, 02 |
| `05-streaming-memory.md` | SSE-стриминг и память диалога | 02 |
| `07-observability.md` | OpenTelemetry: трейсы и метрики агента | 02 |
| `08-guardrails.md` | Безопасность: read-only, allow-list, prompt-injection, adversarial evals | 02, 03 |
| `09-model-comparison.md` | Сравнение моделей и отчёт по evals | 03 |

> Номера 06 и 10 намеренно пропущены — это задел на будущее (auth/мультитенантность
> и Telegram-интерфейс).

## Инженерная гигиена монорепы (11–15)

Отдельный трек — не про LLM-агента, а про сам репозиторий. Это слабые места,
найденные при архитектурном разборе перед собеседованиями на tech-lead /
solution-architect. В каждой задаче — «зачем» и «как» вплоть до конкретных файлов
и правок. Код правишь ты.

| Файл | Тема |
|---|---|
| `11-git-hygiene.md` | Чистый коммит-лог: conventional commits + интерактивный rebase |
| `12-module-boundaries.md` | Контроль границ модулей: `depConstraints` + `@nx/enforce-module-boundaries` |
| `13-db-migrations.md` | Миграции БД: goose вместо ручного slice |
| `14-docs-drift.md` | Дрейф документации: `react-wrapper`, dead-код |
| `15-tooling-robustness.md` | Хрупкий тулинг: staticcheck, PATH, venv, TypeScript-шим |

## Что должно получиться в итоге

```
apps/python-expense-agent/
├── app/
│   ├── config.py            # настройки из .env (как в python-torrent-bot)
│   ├── db/                  # psycopg3: подключение + репозитории
│   ├── tools/               # функции-инструменты над данными
│   ├── mcp_server.py        # MCP-сервер (задача 1)
│   ├── agent/               # Ollama + MCP-клиент + агентный цикл (задача 2)
│   ├── api/                 # FastAPI POST /chat (задача 2)
│   └── main.py              # запуск HTTP API (задача 2)
├── tests/                   # evals + юнит-тесты (задача 3)
├── pyproject.toml
├── requirements.txt
├── .env.example
└── project.json             # Nx-таргеты: setup/serve/mcp/test/lint/typecheck
```

## Общие правила

- **Не коммить секреты.** `.env` в `.gitignore`, в репозиторий — только `.env.example`.
- **Маленькие коммиты** по шагам задачи: после каждого пункта «готово» — коммит.
- **`Decimal` ≠ `float`.** `NUMERIC` из psycopg3 приходит как `Decimal`; `json.dumps`
  его не сериализует. Приводи к `float` (или `str`) на границе инструментов.
- **Типизация.** В проекте `mypy strict` — приучайся писать аннотации сразу.
- **Сначала руками, потом автоматизация.** Каждый пункт проверяй вручную (curl / MCP
  Inspector / opencode), и только потом пиши тест.

## Предварительные требования

- Python 3.12
- Доступ к PostgreSQL `flatty` (локально через podman-compose)
- Ollama с моделью, поддерживающей `tools`
- Свободный порт `8100` (8080 — go-api, 8090 — torrent-bot, 3000 — nest-graphql)
