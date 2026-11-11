# Задача 01. MCP-сервер над данными расходов

> Тема: **Model Context Protocol (MCP)** + прямой доступ к PostgreSQL из Python.
> Результат: работающий MCP-сервер, который умеет отвечать на вопросы о расходах,
> и который можно подключить к opencode (или любому MCP-клиенту).

---

## Цель обучения

К концу задачи ты должен уметь:

- объяснить, что такое MCP, зачем нужны **tools**, и как работает транспорт `stdio`;
- написать MCP-сервер на Python (`mcp` SDK) с типизированными инструментами;
- подключить его к MCP-клиенту (opencode) и вызвать инструмент руками;
- работать с psycopg3: параметризованные запросы, пул соединений, `Decimal → float`;
- отделять «доступ к данным» (репозитории) от «интерфейса» (MCP).

## Что такое MCP в двух словах

MCP — это JSON-RPC-протокол, которым LLM-приложение (клиент) общается с сервером,
отдающим **инструменты**, **ресурсы** и **промпты**. Нас интересуют инструменты:
ты описываешь функцию (`name`, `description`, параметры), клиент показывает её модели,
модель решает вызвать — клиент вызывает — результат возвращается модели.

Транспорт по умолчанию — `stdio`: клиент запускает процесс сервера и общается через
stdin/stdout. Именно так opencode запускает локальные MCP-серверы.

> В MCP Python SDK **v2** высокоуровневый серверный класс называется `MCPServer`
> (в v1 он назывался `FastMCP`). Импорт v1 удалён, а не помечен deprecated.
> Сверься с установленной версией: `pip index versions mcp` / `pip show mcp`.
> Сниппеты ниже — для v2; для v1 замени `from mcp.server import MCPServer` на
> `from mcp.server.fastmcp import FastMCP`.

Полезные ссылки:

- MCP Python SDK: https://github.com/modelcontextprotocol/python-sdk
- Спецификация: https://modelcontextprotocol.io
- psycopg3: https://www.psycopg.org/psycopg3/docs/

---

## Шаг 0. Каркас приложения

Создай `apps/python-expense-agent` по образцу `apps/python-torrent-bot`.

**`project.json`** (Nx-таргеты; обрати внимание на дополнительный `mcp`):

```json
{
  "name": "python-expense-agent",
  "$schema": "../../node_modules/nx/schemas/project-schema.json",
  "projectType": "application",
  "sourceRoot": "apps/python-expense-agent/app",
  "targets": {
    "setup": {
      "executor": "nx:run-commands",
      "options": {
        "command": "python3 -m venv .venv && .venv/bin/pip install -r requirements.txt",
        "cwd": "apps/python-expense-agent"
      }
    },
    "mcp": {
      "executor": "nx:run-commands",
      "options": {
        "command": ".venv/bin/python -m app.mcp_server",
        "cwd": "apps/python-expense-agent"
      }
    },
    "serve": {
      "executor": "nx:run-commands",
      "options": {
        "command": ".venv/bin/python -m app.main",
        "cwd": "apps/python-expense-agent"
      }
    },
    "test": {
      "executor": "nx:run-commands",
      "options": { "command": ".venv/bin/pytest", "cwd": "apps/python-expense-agent" }
    },
    "lint": {
      "executor": "nx:run-commands",
      "options": { "command": ".venv/bin/ruff check .", "cwd": "apps/python-expense-agent" }
    },
    "typecheck": {
      "executor": "nx:run-commands",
      "options": { "command": ".venv/bin/mypy app", "cwd": "apps/python-expense-agent" }
    }
  },
  "tags": ["scope:backend", "type:agent"]
}
```

**`requirements.txt`** (версию `mcp` сверь с актуальной):

```
psycopg[binary,pool]>=3.3,<4
python-dotenv>=1.0
mcp>=2.0,<3
ollama>=0.4,<1
fastapi>=0.115,<1
uvicorn[standard]>=0.54,<1

# dev
pytest>=8
pytest-asyncio>=0.24
httpx>=0.27
ruff>=0.6
mypy>=1.11
```

**`pyproject.toml`** — скопируй настройки `ruff`/`mypy`/`pytest` из
`apps/python-torrent-bot/pyproject.toml`, поменяй `name`.

**`.env.example`**:

```
DATABASE_URL=postgres://postgres:password@localhost:55000/flatty?sslmode=disable
OLLAMA_BASE_URL=http://192.168.1.85:11434
OLLAMA_MODEL=qwen2.5:7b
API_HOST=127.0.0.1
API_PORT=8100
LOG_LEVEL=INFO
```

Сделай `cp .env.example .env`, `.gitignore` — с `.env`, `.venv/`, `__pycache__/`.
Запусти `nx setup python-expense-agent`.

- [ ] `nx setup python-expense-agent` проходит без ошибок.

## Шаг 1. Конфиг

`app/config.py` — по образцу `apps/python-torrent-bot/app/config.py`:

- `dataclass(frozen=True) Config` с полями `database_url`, `ollama_base_url`,
  `ollama_model`, `api_host`, `api_port`, `log_level`;
- `load_config()` читает `.env` через `python-dotenv`, делает fail-fast проверки
  (обязательные переменные, схема URL, диапазон порта).

- [x] `python -c "from app.config import load_config; print(load_config())"` работает.

## Шаг 2. Слой БД (`app/db/`)

Слой БД — это **чистые функции над SQL**, без MCP, без LLM и без HTTP. Их потом
обернут инструменты (шаг 3). Раздели файлы так:

```
app/db/
├── __init__.py
├── connection.py    # пул соединений (создаётся один раз)
├── types.py         # TypedDict-ы для результата
└── expenses.py      # сами запросы (репозитории)
```

### `app/db/connection.py` — пул соединений

Пул создаётся **один раз на процесс** и переиспользуется. Не создавай пул/соединение
внутри каждой функции — это дорого.

```python
from psycopg import Connection
from psycopg_pool import ConnectionPool


def create_pool(database_url: str) -> "ConnectionPool[Connection]":
    """Создать пул соединений. Вызывать один раз при старте."""
    return ConnectionPool(
        conninfo=database_url,
        min_size=1,
        max_size=5,
        timeout=5.0,   # сколько секунд ждать свободное соединение
        open=True,     # открыть сразу; можно open=False + pool.wait()
    )
```

Требования:

- сигнатура `create_pool(database_url: str) -> ConnectionPool`;
- `min_size=1`, `max_size` небольшой (5–10 хватит для учебного сервиса);
- закрытие — `pool.close()` при остановке (в MCP-сервере это делается в
  `finally`/`lifespan`, в HTTP — в `lifespan`);
- если mypy не принимает generic `ConnectionPool[Connection]`, используй
  `ConnectionPool` с комментарием `# type: ignore[type-arg]` — не отключай strict целиком.

> Для учебного варианта допустимо `psycopg.connect(database_url)` на каждый вызов,
> но пул — правильнее и пригодится в задаче 02 (долгоживущий HTTP-сервис).

### `app/db/types.py` — форма результата

Опиши, что возвращает каждая функция. Это и документация, и проверка для mypy.

```python
from typing import TypedDict


class CategoryRow(TypedDict):
    id: int
    name: str
    description: str | None


class ExpenseRow(TypedDict):
    id: int
    category_id: int
    category_name: str
    amount: float          # Decimal -> float на границе
    description: str | None
    month: int
    year: int


class MonthlyTotalRow(TypedDict):
    month: int
    year: int
    total_spent: float


class MonthlyAverageRow(TypedDict):
    month: int
    year: int
    average_amount: float
    expense_count: int


class YearMonthRow(TypedDict):
    year: int
    month: int
```

Зачем `TypedDict`, а не «просто dict»: инструменты и тесты будут знать точные ключи,
а mypy поймает опечатки вроде `r["total_snent"]`.

### `app/db/expenses.py` — общий паттерн

Все функции устроены одинаково:

```python
from decimal import Decimal

from psycopg.rows import dict_row

from app.db.types import MonthlyTotalRow


def _money(value: Decimal | float | int | None) -> float:
    """NUMERIC/Decimal -> float. None -> 0.0."""
    return float(value) if value is not None else 0.0


def get_monthly_totals(
    pool,                       # ConnectionPool
    resident_location_id: int,
    year: int | None = None,
    month: int | None = None,
) -> list[MonthlyTotalRow]:
    sql = """
        SELECT month, year, total_spent
        FROM expense_monthly_totals
        WHERE resident_location_id = %s
          AND (%s::int IS NULL OR year = %s)
          AND (%s::int IS NULL OR month = %s)
        ORDER BY year DESC, month DESC
    """
    with pool.connection() as conn:
        with conn.cursor(row_factory=dict_row) as cur:
            cur.execute(sql, (resident_location_id, year, year, month, month))
            rows = cur.fetchall()

    return [
        {
            "month": int(r["month"]),
            "year": int(r["year"]),
            "total_spent": _money(r["total_spent"]),
        }
        for r in rows
    ]
```

Разбор ключевых мест:

- `pool.connection()` — контекстный менеджер: берёт соединение из пула и возвращает
  его обратно; при выходе делает commit/rollback (для `SELECT` неважно);
- `conn.cursor(row_factory=dict_row)` — строки приходят как `dict`, а не tuple;
- `%s` — плейсхолдер psycopg (не `$1`, не `?`). Значения передаются **вторым
  аргументом** `execute`, поэтому SQL-инъекция невозможна;
- `(%s::int IS NULL OR year = %s)` — трюк «необязательный фильтр»: один статический
  SQL на все комбинации `year`/`month`. Параметр передаётся дважды;
- `Decimal -> float` — обязательно, иначе `json.dumps` упадёт;
- собираем новый dict явно (а не возвращаем строку как есть) — так mypy strict
  понимает тип и мы контролируем набор полей.

### Функции, которые нужно реализовать

| Функция | Возвращает | SQL-суть |
|---|---|---|
| `list_categories(pool)` | `list[CategoryRow]` | `SELECT id, name, description FROM categories ORDER BY name` |
| `list_expenses(pool, rl_id, year=None, month=None, category_id=None, limit=50, offset=0)` | `list[ExpenseRow]` | `expenses JOIN categories` + фильтры |
| `get_monthly_totals(pool, rl_id, year=None, month=None)` | `list[MonthlyTotalRow]` | `expense_monthly_totals` |
| `get_monthly_averages(pool, rl_id, year=None, month=None)` | `list[MonthlyAverageRow]` | `expense_monthly_averages` |
| `get_years_months(pool, rl_id)` | `list[YearMonthRow]` | `SELECT DISTINCT year, month FROM expenses ...` |

`list_expenses` — единственная с **динамическим** набором фильтров. Строй список
условий и список параметров, а не подставляй значения в SQL:

```python
def list_expenses(
    pool,
    resident_location_id: int,
    year: int | None = None,
    month: int | None = None,
    category_id: int | None = None,
    limit: int = 50,
    offset: int = 0,
) -> list[ExpenseRow]:
    where = ["e.resident_location_id = %s"]
    params: list[object] = [resident_location_id]

    if year is not None:
        where.append("e.year = %s")
        params.append(year)
    if month is not None:
        where.append("e.month = %s")
        params.append(month)
    if category_id is not None:
        where.append("e.category_id = %s")
        params.append(category_id)

    limit = max(1, min(limit, 100))   # защита от limit=100000 из LLM
    offset = max(0, offset)

    sql = f"""
        SELECT e.id, e.category_id, c.name AS category_name,
               e.amount, e.description, e.month, e.year
        FROM expenses e
        JOIN categories c ON c.id = e.category_id
        WHERE {' AND '.join(where)}
        ORDER BY e.year DESC, e.month DESC, e.id DESC
        LIMIT %s OFFSET %s
    """
    params.extend([limit, offset])
    # ... execute(sql, params) и сборка ExpenseRow, как в примере выше
```

> f-string здесь склеивает **только доверенные фрагменты** (`where` собирается в коде),
> а все пользовательские значения по-прежнему идут через `%s`. Никогда не вставляй
> `year`/`month`/`description` в f-string.

`get_years_months`:

```sql
SELECT DISTINCT year, month
FROM expenses
WHERE resident_location_id = %s
ORDER BY year DESC, month DESC
```

`get_monthly_averages` — по аналогии с totals, поля `average_amount`, `expense_count`.

### Требования ко слою

- **Только параметризованные запросы** (`%s`), никакой конкатенации пользовательских
  значений — защита от SQL-инъекций, раз аргументы приходят от LLM.
- `NUMERIC` → `float` на границе репозитория (через `_money`).
- Все функции типизированы, `mypy strict` проходит.
- Пустой результат — `[]`, а не `None`.
- `limit`/`offset` нормализуются внутри (clamp), а не доверяются вызывающему.
- Репозиторий **не знает** про пользователей и авторизацию: он принимает
  `resident_location_id`. Скоуп по `user_id` — задача уровня инструмента/API.
- Никаких `print` — только `logging` (и в MCP-сервере — в stderr, см. шаг 4).

### Проверка слоя

Быстрый временный скрипт (не коммить) или `psql`:

```python
from app.config import load_config
from app.db.connection import create_pool
from app.db.expenses import list_categories, get_monthly_totals

cfg = load_config()
pool = create_pool(cfg.database_url)
print(list_categories(pool))
print(get_monthly_totals(pool, 1, 2023))
pool.close()
```

- [x] Пул создаётся и закрывается без ошибок.
- [x] Каждая из 5 функций возвращает осмысленные данные на реальной БД.
- [x] `list_expenses` с разными комбинациями фильтров даёт разные выборки.
- [x] `limit=100000` фактически возвращает не больше 100 строк.
- [x] Все `total_spent`/`amount`/`average_amount` — `float`, а не `Decimal`.
- [x] Пустой месяц возвращает `[]`.

## Шаг 3. Инструменты (`app/tools/`)

Оберни репозитории в «инструменты» — функции с понятным именем, аннотациями и
**docstring** (описание инструмента попадает в схему для модели, поэтому пиши его
по-человечески, что и какие аргументы делает).

Пример формы (не готовое решение, а ориентир):

```python
def get_monthly_totals(
    resident_location_id: int,
    year: int | None = None,
    month: int | None = None,
) -> list[dict[str, object]]:
    """Вернуть суммы трат по месяцам для объекта.

    Args:
      resident_location_id: id объекта (квартиры).
      year: год, если нужно отфильтровать.
      month: месяц 1-12, если нужно отфильтровать.
    """
    ...
```

Требования к инструментам:

- возвращают **JSON-сериализуемые** структуры (`dict`/`list`/`str`/`int`/`float`);
- понятные сообщения об ошибках (например, объект не найден → пустой список или
  понятная ошибка, но не 500-стектрейс в модель);
- не содержат бизнес-логики сверх «сходи в БД и верни данные».

- [x] Ручной вызов каждого инструмента из Python возвращает ожидаемое.

## Шаг 4. MCP-сервер (`app/mcp_server.py`)

Собери MCP-сервер, который публикует инструменты из шага 3.

Ориентир для v2:

```python
from mcp.server import MCPServer

mcp = MCPServer("expense-agent")

@mcp.tool()
def get_monthly_totals(
    resident_location_id: int,
    year: int | None = None,
    month: int | None = None,
) -> list[dict[str, object]]:
    """Вернуть суммы трат по месяцам для объекта."""
    ...

if __name__ == "__main__":
    mcp.run()          # stdio по умолчанию
```

Требования:

- имена и описания инструментов — на английском (так стабильнее для моделей), а
  значения/данные могут быть любыми;
- запуск без аргументов = `stdio`;
- проверь, что процесс не печатает в **stdout** ничего лишнего (stdout занят
  протоколом — логи только в stderr!). Это классическая ловушка.

Проверка сервера — через **MCP Inspector**:

```bash
npx @modelcontextprotocol/inspector .venv/bin/python -m app.mcp_server
```

- [x] Inspector видит все инструменты.
- [x] Через Inspector инструмент возвращает корректные данные.

## Шаг 5. Подключение к opencode

Добавь сервер в `opencode.json` в корне репозитория (рядом с существующим конфигом):

```json
{
  "mcp": {
    "expense-agent": {
      "type": "local",
      "enabled": true,
      "command": "apps/python-agent/.venv/bin/python",
      "args": ["-m", "app.mcp_server"],
      "environment": {
        "DATABASE_URL": "postgres://postgres:password@localhost:55000/flatty?sslmode=disable"
      }
    }
  }
}
```

> Используй абсолютный путь к `.venv/bin/python` или запускай с `cwd` приложения —
> иначе модуль `app` не найдётся. Проверь, как opencode передаёт рабочую директорию.

Перезапусти opencode и проверь, что он видит инструменты. Затем спроси что-нибудь
вроде: «Сколько потрачено в январе 2023 по объекту 1?» — агент должен вызвать
`get_monthly_totals` и ответить числом.

- [x] opencode видит инструменты.
- [x] Ответ на вопрос содержит правильное число.

## Критерии готовности

- [x] `nx lint python-agent` — чисто.
- [x] `nx typecheck python-agent` — чисто (`mypy strict`).
- [x] `nx setup` и запуск MCP работают с нуля по инструкции.
- [ ] Все 5 инструментов доступны в Inspector.
- [x] opencode через MCP отвечает на 2–3 разных вопроса о расходах.
- [x] В `.env.example` нет реальных секретов, `.env` в `.gitignore`.

## Подсказки и грабли

- **stdout vs stderr.** MCP stdio использует stdout для JSON-RPC. Любой `print(...)`
  в коде сервера сломает протокол. Логируй в stderr (`logging` с handler в stderr).
- **`Decimal`.** `psycopg` возвращает `NUMERIC` как `Decimal`. `json.dumps(Decimal)`
  падает — конвертируй.
- **Дефолтные значения в схеме.** `int | None = None` даёт в JSON Schema
  `"anyOf"`/nullable. Некоторые модели путаются; можно делать отдельные инструменты
  без необязательных параметров, если качество страдает.
- **Идентификаторы.** Не доверяй «магическим» id от модели: пусть она сначала вызовет
  `get_years_months`/`list_categories`, потом уточняет.
- **Версия SDK.** Если `from mcp.server import MCPServer` не импортируется — у тебя v1,
  используй `FastMCP`. Зафиксируй версию в `requirements.txt`.

## Что почитать

- MCP Python SDK README и `docs/` (сервер, клиент, транспорты).
- MCP Inspector — как отлаживать сервер без агента.
- psycopg3: connection pool, параметризация, работа с `Decimal`.

## Бонус (по желанию)

- Добавь транспорт `streamable-http` (`mcp.run(transport="streamable-http", ...)`) и
  подключи opencode по URL вместо subprocess.
- Добавь `@mcp.resource`, отдающий справочник категорий.
- Напиши юнит-тесты на репозитории (без MCP) — они пригодятся в задаче 03.
