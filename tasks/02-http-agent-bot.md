# Задача 02. Бот-агент: HTTP endpoint + agentic tool-calling

> Тема: **LLM agentic tool-calling** + MCP-клиент + FastAPI.
> Результат: HTTP-сервис `POST /chat`, который на естественном вопросе о расходах
> сам решает, какие инструменты MCP вызвать, и отвечает числом/текстом.

> ⚠️ **Правка пути.** В этом файле везде используется реальное имя Nx-приложения
> `apps/python-agent` (папки `apps/python-agent` **не существует**). Если
> где-то ниже встретишь `python-expense-agent` — это опечатка, в коде пиши
> `python-agent`.

---

## Цель обучения

К концу задачи ты должен уметь:

- объяснить разницу между «классическим RAG» и «агентным tool-calling» и почему для
  числовых агрегатов выбран второй;
- написать агентный цикл: вопрос → модель просит инструмент → вызов → ответ;
- использовать **MCP-клиент** (`mcp` SDK) как источник инструментов, а не хардкодить их;
- поднимать FastAPI-сервис внутри того же проекта;
- аккуратно обрабатывать таймауты, лимит итераций и ошибки инструментов.

## Почему не «классический RAG»

RAG (эмбеддинги + поиск похожих кусков) хорошо ищет **релевантный текст**. Но на вопрос
«сколько потрачено в январе 2023?» нужен точный `SUM`, а не «похожий по смыслу текст».
Поэтому здесь модель не «вспоминает», а **вызывает инструмент**, который делает
SQL-запрос. Это надёжнее и проверяемее. RAG при желании добавляется позже отдельным
инструментом (семантический поиск по `description`), но в базовой задаче его нет.

---

## Архитектура

```
HTTP POST /chat
   │
   ▼
FastAPI  ──►  Agent loop
                 │  1. собрать messages + список tools (из MCP)
                 │  2. Ollama.chat(model, messages, tools)
                 │  3. если есть tool_calls → вызвать через MCP-клиент
                 │     → добавить результат в messages → назад к 2
                 │  4. если tool_calls нет → вернуть финальный текст
                 ▼
             MCP client ──(stdio)──► app.mcp_server (задача 01)
                                          │
                                          ▼
                                     PostgreSQL
```

Ключевая идея: **инструменты не дублируются**. Их единственный источник — MCP-сервер.
Бот узнаёт о них через протокол (`list_tools`), а не импортирует репо-функции напрямую.

---

## Предварительные условия

- Задача 01 выполнена: `apps/python-agent/app/mcp_server.py` стартует по stdio,
  инструменты отвечают.
- Ollama доступен на `http://192.168.1.85:11434` и есть модель с поддержкой `tools`
  (`qwen2.5:7b`, `llama3.1:8b`, `mistral`).

Проверка, что модель умеет tools:

```bash
curl -s http://192.168.1.85:11434/api/chat -d '{
  "model": "qwen2.5:7b",
  "messages": [{"role": "user", "content": "what is 2+2?"}],
  "tools": [{
    "type": "function",
    "function": {
      "name": "add",
      "description": "Add two numbers",
      "parameters": {"type": "object", "properties": {"a": {"type": "integer"}, "b": {"type": "integer"}}, "required": ["a","b"]}
    }
  }],
  "stream": false
}'
```

Если в ответе есть `message.tool_calls` — всё хорошо.

---

## Что уже есть в репозитории (не переделывай)

Прочитай эти файлы перед стартом — часть работы за тебя уже сделана в задаче 01:

| Файл | Что там |
|------|---------|
| `apps/python-agent/app/config.py` | `Config` (frozen dataclass) + `load_config()`. Уже читает `DATABASE_URL`, `OLLAMA_BASE_URL`, `OLLAMA_MODEL`, `API_HOST`, `API_PORT`, `LOG_LEVEL`. Поля `cfg.ollama_base_url` (host) и `cfg.ollama` (имя модели) — готовы к использованию. |
| `apps/python-agent/app/db/connection.py` | `Pool = ConnectionPool[Connection]`, `create_pool(url)`. |
| `apps/python-agent/app/db/expenses.py` | 5 репо-функций: `get_monthly_totals`, `list_categories`, `list_expenses`, `get_month_averages`, `get_year_months`. `Decimal`→`float` уже сделан через `_money()`. |
| `apps/python-agent/app/db/types.py` | TypedDict'ы строк (`MonthlyTotalRow`, `ExpenseRow`, …). |
| `apps/python-agent/app/mcp_server.py` | `build_mcp_server(pool)` + `if __name__ == "__main__"` (stdio). |
| `apps/python-agent/requirements.txt` | **Уже содержит** `mcp`, `ollama`, `fastapi`, `uvicorn` и dev-зависимости (`pytest`, `pytest-asyncio`, `httpx`, `ruff`, `mypy`). Новых зависимостей ставить не нужно. |
| `apps/python-agent/.env` | `DATABASE_URL`, `OLLAMA_BASE_URL`, `OLLAMA_MODEL=qwen2.5:7b`, `API_HOST=127.0.0.1`, `API_PORT=8100`. |

**Два момента, которые надо поправить перед началом:**

1. **`project.json`** — таргеты `mcp` и `serve` указывают на несуществующую папку
   `apps/python-expense-agent`. Исправь `cwd` на `apps/python-agent` (см. Шаг 6).

2. **MCP-сервер регистрирует только один инструмент.** Сейчас в
   `build_mcp_server(pool)` есть `get_monthly_totals` и ресурс `resource://categories`,
   но нет `list_expenses` / `get_month_averages` / `get_year_months` — хотя
   репо-функции для них уже написаны. Для ручной проверки (Шаг 7) нужен минимум
   `list_expenses`, иначе вопросы «покажи последние 5 трат» не сработают. См. Шаг 0.

---

## Шаг 0. Дополни MCP-сервер инструментами (5 минут)

Открой `apps/python-agent/app/mcp_server.py`. Внутри `build_mcp_server(pool)` после
`get_monthly_totals` добавь ещё инструменты тем же декоратором `@mcp.tool()`. Все
репо-функции уже импортированы как `repo`:

```python
    @mcp.tool()
    def list_expenses(
        resident_location_id: int,
        year: int | None = None,
        month: int | None = None,
        category_id: int | None = None,
        limit: int = 50,
        offset: int = 0,
    ) -> list[ExpenseRow]:
        """Вернуть список трат объекта (с фильтрами по году/месяцу/категории)."""

        return repo.list_expenses(
            pool=pool,
            resident_location_id=resident_location_id,
            year=year,
            month=month,
            category_id=category_id,
            limit=limit,
            offset=offset,
        )
```

`ExpenseRow` уже импортирован сверху (`from app.db.types import ... ExpenseRow`). По
желанию так же зарегистрируй `get_month_averages` и `get_year_months` — репо-функции
готовы, сигнатура у них `(pool, resident_location_id)`.

Проверь, что сервер по-прежнему стартует:

```bash
cd apps/python-agent
.venv/bin/python -m app.mcp_server   # зависнет — это нормально (stdio-сервер)
# Ctrl+C
```

> Совет: вызывай инструменты и проверяй через MCP Inspector или клиент из задачи 01,
> прежде чем идти дальше. Агентный цикл сложнее дебажить, чем «голый» инструмент.

---

## Шаг 1. Клиент Ollama (`app/agent/ollama_client.py`)

Создай пакет `app/agent/` (пустой `__init__.py`) и файл `ollama_client.py`.

Используй официальный пакет `ollama` (он уже в `requirements.txt`):

```python
from ollama import AsyncClient

from app.config import Config


def build_ollama_client(cfg: Config) -> AsyncClient:
    """Клиент Ollama; host и модель — из Config (не хардкодить)."""
    return AsyncClient(host=cfg.ollama_base_url)
```

Требования:

- host берётся из `cfg.ollama_base_url`, модель — из `cfg.ollama`;
- **`temperature` передаётся через `options={"temperature": 0.0}`**, а не как отдельный
  аргумент — у `AsyncClient.chat()` нет параметра `temperature`;
- таймаут и понятная ошибка, если Ollama недоступна (это удобнее сделать в цикле,
  Шаг 4).

Проверка (временный скрипт или `python -c`):

```bash
cd apps/python-agent
.venv/bin/python -c "
import asyncio
from app.config import load_config
from app.agent.ollama_client import build_ollama_client

async def main():
    cfg = load_config()
    client = build_ollama_client(cfg)
    print(cfg)
    r = await client.chat(model=cfg.ollama, messages=[{'role':'user','content':'2+2?'}], options={'temperature': 0.0})
    print(r.message.content)

asyncio.run(main())
"
```

- [x] Простой `chat` без инструментов возвращает текст.

---

## Шаг 2. MCP-клиент (`app/agent/mcp_client.py`)

Подключись к своему MCP-серверу как клиент. Используй **высокоуровневый** `Client`
(он внутри поднимает subprocess и сессию):

```python
import sys
from pathlib import Path

from mcp import Client, StdioServerParameters

# app/agent/mcp_client.py → parents[0]=agent, [1]=app, [2]=python-agent
PROJECT_ROOT = Path(__file__).resolve().parents[2]


def build_server_params() -> StdioServerParameters:
    """Параметры запуска MCP-сервера (задача 01) как subprocess по stdio."""
    return StdioServerParameters(
        command=sys.executable,          # тот же интерпретатор venv, что и у бота
        args=["-m", "app.mcp_server"],
        cwd=str(PROJECT_ROOT),           # чтобы нашлись .env и пакет app
    )
```

Почему именно так:

- `command=sys.executable` — не хардкодим путь к `.venv/bin/python`, берём тот
  интерпретатор, под которым работает сам бот;
- `cwd=PROJECT_ROOT` — иначе `python -m app.mcp_server` не найдёт пакет `app`, а
  `load_dotenv()` не увидит `.env`.

Высокоуровневое подключение (в lifespan, не на каждый запрос — см. Шаг 5):

```python
async with Client(params) as mcp:
    tools = await mcp.list_tools()          # ListToolsResult → поле .tools
    result = await mcp.call_tool(
        "get_monthly_totals",
        {"resident_location_id": 1, "year": 2023, "month": 1},
    )
```

**Реальная сигнатура MCP 2.x** (проверено по установленной версии):

- `Client(server, *, raise_exceptions=False, ...)` — `server` может быть
  `StdioServerParameters`; контекст-менеджер делает connect/close сам;
- `Client.list_tools() -> ListToolsResult` — список инструментов лежит в `.tools`
  (каждый — `mcp.types.Tool`);
- `Client.call_tool(name, arguments) -> CallToolResult` — у результата есть
  `.content` (список блоков) и `.structured_content` (распарсенный объект).

Если хочешь полный контроль — низкоуровневый вариант (`ClientSession` +
`stdio_client`) работает так же, но `Client` проще и для этой задачи достаточно.

- [x] Временный скрипт: подключиться, напечатать имена инструментов, вызвать один
      (`get_monthly_totals`) и увидеть список сумм.

---

## Шаг 3. Конвертация схем (`app/agent/tools.py`)

Ollama ждёт инструменты в формате:

```json
{
  "type": "function",
  "function": {
    "name": "get_monthly_totals",
    "description": "...",
    "parameters": { "type": "object", "properties": { ... } }
  }
}
```

MCP отдаёт `Tool` с полями `name`, `description`, **`input_schema`** (обрати внимание:
snake_case, не `inputSchema` — это поле реальной модели MCP 2.x). Напиши функцию
конвертации:

```python
from mcp.types import Tool


def mcp_tools_to_ollama(tools: list[Tool]) -> list[dict]:
    """MCP Tool → список инструментов в формате Ollama chat(tools=...)."""
    return [
        {
            "type": "function",
            "function": {
                "name": tool.name,
                "description": tool.description or "",
                "parameters": tool.input_schema
                or {"type": "object", "properties": {}},
            },
        }
        for tool in tools
    ]
```

Требования:

- не терять `description` (именно она помогает модели выбрать инструмент);
- не терять `input_schema` (иначе модель не поймёт аргументы);
- покрыть юнит-тестом на одном примере.

Минимальный тест `tests/test_tools.py`:

```python
from mcp.types import Tool

from app.agent.tools import mcp_tools_to_ollama


def test_converts_mcp_tool_to_ollama() -> None:
    tool = Tool(
        name="get_monthly_totals",
        description="Вернуть суммы трат по месяцам",
        input_schema={
            "type": "object",
            "properties": {"resident_location_id": {"type": "integer"}},
            "required": ["resident_location_id"],
        },
    )
    result = mcp_tools_to_ollama([tool])
    assert result[0] == {
        "type": "function",
        "function": {
            "name": "get_monthly_totals",
            "description": "Вернуть суммы трат по месяцам",
            "parameters": {
                "type": "object",
                "properties": {"resident_location_id": {"type": "integer"}},
                "required": ["resident_location_id"],
            },
        },
    }
```

Запуск:

```bash
cd apps/python-agent && .venv/bin/pytest tests/test_tools.py -q
```

- [ ] Конвертация даёт валидный список для `client.chat(tools=...)`.

---

## Шаг 4. Агентный цикл (`app/agent/loop.py`)

Сердце задачи. Два уровня API (`Client` и `ClientSession`) оба дают `list_tools()` и
`call_tool(name, arguments)`, поэтому цикл принимает любой объект с этими методами.

```python
import json
from dataclasses import dataclass, field
from typing import Any, Protocol

from ollama import AsyncClient

from app.agent.tools import mcp_tools_to_ollama

MAX_STEPS = 6

SYSTEM_PROMPT = (
    "Ты — помощник по личным расходам. Для любых цифр и фактов "
    "обязательно вызывай инструменты и никогда не выдумывай суммы. "
    "Если данных нет — так и скажи. Отвечай кратко и по-русски."
)


@dataclass
class AgentResult:
    answer: str
    used_tools: list[str] = field(default_factory=list)


class McpClient(Protocol):
    async def list_tools(self) -> Any: ...

    async def call_tool(self, name: str, arguments: dict | None = None) -> Any: ...


def tool_result_to_text(result: Any) -> str:
    """CallToolResult → строка для сообщения модели."""
    structured = getattr(result, "structured_content", None)
    if structured is not None:
        return json.dumps(structured, ensure_ascii=False, default=str)
    # fallback: склеить текстовые блоки из result.content
    parts: list[str] = []
    for block in getattr(result, "content", []) or []:
        text = getattr(block, "text", None)
        if text:
            parts.append(text)
    return "\n".join(parts)


async def run_agent(
    ollama: AsyncClient,
    mcp: McpClient,
    question: str,
    model: str,
    resident_location_id: int,
) -> AgentResult:
    tools_result = await mcp.list_tools()
    tools = mcp_tools_to_ollama(tools_result.tools)

    system_prompt = (
        SYSTEM_PROMPT
        + f"\nТекущий объект (resident_location_id): {resident_location_id}."
    )
    messages: list[dict] = [
        {"role": "system", "content": system_prompt},
        {"role": "user", "content": question},
    ]
    used: list[str] = []

    for _ in range(MAX_STEPS):                       # защита от бесконечного цикла
        response = await ollama.chat(
            model=model,
            messages=messages,
            tools=tools,
            options={"temperature": 0.0},
        )
        msg = response.message
        calls = msg.tool_calls or []
        if not calls:
            return AgentResult(answer=msg.content or "", used_tools=used)

        messages.append({                            # assistant с tool_calls
            "role": "assistant",
            "content": msg.content or "",
            "tool_calls": calls,
        })
        for call in calls:
            name = call.function.name
            used.append(name)
            try:
                arguments = dict(call.function.arguments or {})
                result = await mcp.call_tool(name, arguments)
                content = tool_result_to_text(result)
            except Exception as exc:                 # ошибку отдаём модели, не падаем
                content = f"Ошибка вызова инструмента {name}: {exc}"
            messages.append({
                "role": "tool",
                "tool_name": name,
                "content": content,
            })

    return AgentResult(
        answer="Не удалось получить ответ за отведённое число шагов",
        used_tools=used,
    )
```

Ключевые детали (обязательно):

- **лимит итераций** (`MAX_STEPS = 6`) — иначе модель может зациклиться на повторном
  вызове одного инструмента;
- **неизвестный инструмент / ошибка инструмента** → результат-строку с текстом ошибки
  возвращаем модели (`role: "tool"`), а не роняем весь цикл;
- **формат tool-сообщения у Ollama** — отдельное сообщение с `role: "tool"` и полем
  **`tool_name`** (не `name` и не `function.name`). Рядом обязательно должно лежать
  assistant-сообщение с `tool_calls`, иначе модель «не увидит» результат;
- `resident_location_id` подмешиваем в системный промпт — модель не угадает его из
  вопроса, а инструментам он обязателен;
- `AgentResult` содержит `answer` и `used_tools` — это понадобится в evals (задача 03).

- [x] Вопрос «Сколько потрачено в январе 2023 по объекту 1?» → вызов
      `get_monthly_totals` → правильное число.

---

## Шаг 5. FastAPI (`app/api/app.py`)

Создай пакет `app/api/` (пустой `__init__.py`) и файл `app.py`.

MCP-сессия открывается **один раз** в `lifespan` (startup/shutdown), а не в хендлере —
иначе на каждый запрос будет подниматься новый subprocess.

```python
from contextlib import asynccontextmanager

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel

from app.agent.loop import run_agent
from app.agent.mcp_client import build_server_params
from app.agent.ollama_client import build_ollama_client
from app.config import load_config
from mcp import Client


class ChatRequest(BaseModel):
    message: str
    resident_location_id: int = 1


class ChatResponse(BaseModel):
    answer: str
    used_tools: list[str]


@asynccontextmanager
async def lifespan(app: FastAPI):
    cfg = load_config()
    app.state.model = cfg.ollama
    app.state.ollama = build_ollama_client(cfg)
    async with Client(build_server_params()) as mcp:   # connect на старте, close на стопе
        app.state.mcp = mcp
        yield


app = FastAPI(title="expense-agent", lifespan=lifespan)


@app.get("/health")
def health() -> dict[str, str]:
    return {"status": "ok"}


@app.post("/chat", response_model=ChatResponse)
async def chat(req: ChatRequest) -> ChatResponse:
    if not req.message.strip():
        raise HTTPException(status_code=400, detail="message is empty")

    try:
        result = await run_agent(
            app.state.ollama,
            app.state.mcp,
            req.message,
            app.state.model,
            req.resident_location_id,
        )
    except Exception as exc:          # сузь тип позже; главное — не 500 с непонятным трейсом
        raise HTTPException(status_code=503, detail=f"agent error: {exc}") from exc

    return ChatResponse(answer=result.answer, used_tools=result.used_tools)
```

Требования:

- MCP-сессия открывается в `lifespan`, а не в хендлере;
- `POST /chat` возвращает `answer` и `used_tools`;
- корректные коды ошибок: `400` на пустой вопрос, `503` если Ollama/агент недоступен;
- Swagger доступен на `/docs`.

---

## Шаг 6. `app/main.py` и правка `project.json`

`app/main.py`:

```python
import uvicorn

from app.config import load_config

if __name__ == "__main__":
    cfg = load_config()
    uvicorn.run("app.api.app:app", host=cfg.api_host, port=cfg.api_port)
```

Запуск сервиса:

```bash
nx setup python-agent   # один раз: .venv + зависимости (если ещё не делал)
nx serve python-agent   # HTTP API на http://127.0.0.1:8100/docs
```

- [x] `nx serve python-agent` поднимает сервис на `:8100`.
- [x] `curl -s localhost:8100/health` → `{"status":"ok"}`.
- [x] `curl` на `/chat` отвечает осмысленно.

---

## Шаг 7. Ручная проверка

Пройди 4–5 вопросов разного типа:

```bash
curl -s localhost:8100/chat -H 'content-type: application/json' \
  -d '{"message":"Сколько всего потрачено в январе 2023?","resident_location_id":1}'

curl -s localhost:8100/chat -H 'content-type: application/json' \
  -d '{"message":"Какая категория самая дорогая в 2023 году?","resident_location_id":1}'

curl -s localhost:8100/chat -H 'content-type: application/json' \
  -d '{"message":"Покажи последние 5 трат","resident_location_id":1}'

curl -s localhost:8100/chat -H 'content-type: application/json' \
  -d '{"message":"привет"}'
```

- [x] Ответы содержат правильные цифры (сверь с `psql` / клиентом из задачи 01).
- [x] `used_tools` соответствует вопросу (агрегат → `get_monthly_totals`, список →
      `list_expenses`).
- [x] Вопрос «привет» не вызывает инструменты и отвечает текстом.

---

## Критерии готовности

- [x] `nx lint python-agent` / `nx typecheck python-agent` — чисто.
- [x] Сервис стартует и корректно останавливается (MCP-сессия закрывается).
- [x] `/chat` отвечает на вопросы трёх типов: агрегат, список, «светская беседа».
- [x] Модель не выдумывает суммы: если данных нет, инструмент вернул пусто, и агент
      честно говорит об этом.
- [x] `used_tools` заполняется и логируется (в stderr, не в stdout).

---

## Подсказки и грабли

- **Модель без tool-calling.** Если `tool_calls` всегда `None`, модель их не
  поддерживает или схема невалидна. Проверь curl-ом из начала файла.
- **`temperature` не отдельным аргументом.** У `AsyncClient.chat()` его нет — только
  через `options={"temperature": 0.0}`.
- **Поле схемы — `input_schema`.** У MCP 2.x `Tool` называется `input_schema`, а не
  `inputSchema` (как в старых примерах/OpenAI-совместимых API).
- **Цикл без конца.** Всегда ставь `MAX_STEPS`. Модель может вызывать один инструмент
  повторно.
- **Формат `tool`-сообщения.** У Ollama результат инструмента — отдельное сообщение с
  `role: "tool"` и полем `tool_name`. Если ответ модели пустой/странный — проверь, что
  добавил и assistant-сообщение с `tool_calls`, и tool-сообщение с результатом.
- **`list_tools()` возвращает `ListToolsResult`.** Не `list[Tool]` — инструменты лежат
  в `.tools`.
- **Один subprocess на запрос — плохо.** Поднимай MCP-сессию в `lifespan`.
- **stdout снова.** MCP-сервер пишет протокол в stdout; логируй только в stderr
  (`logging`, а не `print`).
- **Инъекции.** Аргументы идут в SQL только через параметры `%s` — это уже сделано в
  задаче 01, не ломай.
- **`resident_location_id` не появится сам.** Его надо сообщить модели (системный
  промпт или приставка к вопросу), иначе она передаст инструменту мусор или `null`.

## Что почитать

- Ollama Python: tool calling (`AsyncClient.chat(tools=...)`, `message.tool_calls`,
  `options={"temperature": ...}`).
- MCP Python SDK: клиент (`Client`, `StdioServerParameters`, `call_tool`,
  `CallToolResult.structured_content`).
- FastAPI: `lifespan`, `BaseModel`, `response_model`.
- Про агентные циклы: «ReAct» / «tool use» — общая идея, без привязки к фреймворку.

## Бонус (по желанию)

- Стриминг ответа (`stream=true`) и SSE-endpoint.
- Память диалога (передавать `history` в запросе).
- Инструмент `search_expenses_by_description` на эмбеддингах (`nomic-embed-text`) —
  вот здесь уже появляется классический RAG как **дополнительный** инструмент.
- Ограничение: не отдавать модели `user_id`, если сервис многопользовательский.
