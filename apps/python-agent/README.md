# python-agent (MCP expense-agent)

MCP-сервер для агрегатов по расходам (`expense-agent`). Использует MCP **2.3.0**
(`mcp>=2.0,<3`).

## Структура

| Файл | Роль |
|------|------|
| `app/mcp_server.py` | Сборка сервера `MCPServer`, регистрация tools/resources, `main` |
| `app/db/expenses.py` | 5 репо-функций (SQL к базе) |
| `app/db/connection.py` | `Pool = ConnectionPool[Connection]` + `create_pool` |
| `app/db/types.py` | TypedDict'ы строк (`CategoryRow`, `MonthlyTotalRow`, …) |
| `app/config.py` | Загрузка `DATABASE_URL` из `.env` |

Все функции `app/db/expenses.py` принимают `pool: Pool` первым аргументом.

## Tools

Инструменты — вычисляемые/фильтруемые запросы. Регистрируются декоратором `@mcp.tool()`:

- `get_monthly_totals(resident_location_id, year=None, month=None)`
- `list_expenses(resident_location_id, year=None, month=None, category_id=None, limit=50, offset=0)`
- `get_month_averages(resident_location_id)`
- `get_year_months(resident_location_id)`

## Resources

Ресурсы — это **read-only данные по URI** (справочники, конфигурация), в отличие от tools
(вызываемых действий). Справочник категорий логичнее отдавать как ресурс, а не как tool.

### Как добавить `@mcp.resource` для категорий (сделай сам)

В `app/mcp_server.py`, внутри `build_mcp_server(pool)`, после блока tools добавь:

```python
from app.db.types import CategoryRow  # добавить в импорт сверху

    @mcp.resource(
        uri="resource://categories",
        name="categories",
        description="Справочник категорий расходов (id, name, description).",
    )
    def categories() -> list[CategoryRow]:
        """Статические справочные данные — возвращаются по URI, не вызываются."""
        return repo.list_categories(pool)
```

Что происходит:

- `@mcp.resource(uri=..., name=..., description=...)` — сигнатура MCP 2.3.0:
  `resource(uri, *, name=None, title=None, description=None, mime_type=None, ...)`.
- Декорированная функция вызывается **без аргументов** и возвращает данные.
- Возвращаемый тип определяет способ отдачи:
  - `str` → отдаётся как текст;
  - `bytes` → бинарный контент;
  - любой другой объект (список, dict) → сериализуется в **JSON**.
  Поэтому `list[CategoryRow]` (список dict'ов) отдаётся как JSON-массив без доп. кода.
- Для текстового ресурса можно указать `mime_type` (например `"text/plain"` /
  `"application/json"`), иначе MCP подставит тип по умолчанию.

### Проверка

После добавления ресурса в Inspector (и в любом MCP-клиенте) должен появиться
ресурс `resource://categories` рядом с инструментами. Чтение ресурса вернёт JSON-массив
категорий.

## Запуск

```bash
cd apps/python-agent
.venv/bin/python -m app.mcp_server
```

`mcp.run()` без аргументов использует stdio-транспорт. Для HTTP:

```python
mcp.run(transport="streamable-http", host="127.0.0.1", port=8000)
```

## Проверка типов

```bash
.venv/bin/mypy
```
