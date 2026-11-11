# python-agent (MCP expense-agent)

MCP server for expense aggregates (`expense-agent`). Uses MCP **2.3.0**
(`mcp>=2.0,<3`).

## Structure

| File | Role |
|------|------|
| `app/mcp_server.py` | Builds the `MCPServer`, registers tools/resources, `main` entrypoint |
| `app/db/expenses.py` | 5 repo functions (SQL against the database) |
| `app/db/connection.py` | `Pool = ConnectionPool[Connection]` + `create_pool` |
| `app/db/types.py` | Row TypedDicts (`CategoryRow`, `MonthlyTotalRow`, …) |
| `app/config.py` | Loads config from `.env` |

All `app/db/expenses.py` functions take `pool: Pool` as their first argument.

## Tools

Tools are computed/filtered queries, registered with the `@mcp.tool()` decorator:

- `get_monthly_totals(resident_location_id, year=None, month=None)`
- `list_expenses(resident_location_id, category_id=None, month=None, year=None, limit=5, offset=0)`
- `get_month_averages(resident_location_id)`
- `get_year_months(resident_location_id)`
- `list_categories()` — expense categories (`id`, `name`, `description`)

## Resources

Resources are **read-only data exposed by URI** (reference data, configuration),
unlike tools (invoked actions). The category reference is exposed as a resource:

- `resource://categories` — expense categories reference (`id`, `name`, `description`),
  returned as a JSON array.

## Run

The server requires a valid `.env` — `load_config()` validates `DATABASE_URL`,
`OLLAMA_MODEL`, and `OLLAMA_BASE_URL` on startup, even for pure-MCP (stdio) usage.

```bash
cd apps/python-agent
.venv/bin/python -m app.mcp_server
```

`mcp.run()` with no arguments uses the stdio transport. For HTTP:

```python
mcp.run(transport="streamable-http", host="127.0.0.1", port=8000)
```

## Inspector

```bash
npx @modelcontextprotocol/inspector .venv/bin/python -m app.mcp_server
```

The web UI opens at http://localhost:5173 by default.

## Environment variables

| Variable | Required | Default | Description |
|----------|:--------:|---------|-------------|
| `DATABASE_URL` | yes | — | PostgreSQL URL (`postgresql://user:pass@host:5432/db?sslmode=disable`) |
| `OLLAMA_MODEL` | yes | — | Ollama model name |
| `OLLAMA_BASE_URL` | yes | — | Ollama base URL |
| `LOG_LEVEL` | no | `INFO` | One of `DEBUG`/`INFO`/`WARNING`/`ERROR`/`CRITICAL` |
| `API_HOST` | no | `127.0.0.1` | HTTP API bind address (localhost only) |
| `API_PORT` | no | `8080` | HTTP API port |

> Copy `.env.example` to `.env` and fill in the three required variables.

## Lint & type-check

```bash
.venv/bin/ruff check .   # lint (ruff, line length 100, rules E/F/I/UP/B/SIM)
.venv/bin/mypy app       # type-check (mypy, strict)
```

## Setup

```bash
nx setup python-agent     # once: creates .venv and installs requirements.txt
```
