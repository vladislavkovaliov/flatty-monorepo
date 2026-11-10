# Задача 02. Бот-агент: HTTP endpoint + agentic tool-calling

> Тема: **LLM agentic tool-calling** + MCP-клиент + FastAPI.
> Результат: HTTP-сервис `POST /chat`, который на естественном вопросе о расходах
> сам решает, какие инструменты MCP вызвать, и отвечает числом/текстом.

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
Бот узнаёт о них через протокол.

## Предварительные условия

- Задача 01 выполнена: MCP-сервер отвечает, инструменты работают.
- Ollama доступен и есть модель с поддержкой `tools` (`qwen2.5:7b`, `llama3.1:8b`, ...).

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

## Шаг 1. Клиент Ollama (`app/agent/ollama_client.py`)

Используй официальный пакет `ollama` (не raw HTTP — так чище и это тоже часть практики).

```python
from ollama import AsyncClient

client = AsyncClient(host="http://192.168.1.85:11434")
response = await client.chat(
    model="qwen2.5:7b",
    messages=[{"role": "user", "content": "..."}],
    tools=[...],          # JSON-schema инструментов
)
calls = response.message.tool_calls
```

Требования:

- host и модель берутся из `Config` (`.env`), не хардкодятся;
- `temperature=0` (или низкая) — для воспроизводимости (пригодится в evals);
- таймаут и понятная ошибка, если Ollama недоступна.

- [ ] Простой `chat` без инструментов возвращает текст.

## Шаг 2. MCP-клиент (`app/agent/mcp_client.py`)

Подключись к своему MCP-серверу как клиент. Два уровня API:

**Высокоуровневый** (v2):

```python
from mcp import Client, StdioServerParameters

params = StdioServerParameters(
    command="apps/python-expense-agent/.venv/bin/python",
    args=["-m", "app.mcp_server"],
)
async with Client(params) as client:
    tools = await client.list_tools()
    result = await client.call_tool("get_monthly_totals", {"resident_location_id": 1, "year": 2023, "month": 1})
```

**Низкоуровневый** (`ClientSession` + `stdio_client`) — если нужен полный контроль:

```python
from mcp import ClientSession, StdioServerParameters
from mcp.client.stdio import stdio_client

async with stdio_client(params) as (read, write):
    async with ClientSession(read, write) as session:
        await session.initialize()
        tools = await session.list_tools()
        result = await session.call_tool("get_monthly_totals", arguments={...})
```

Требования:

- **долгоживущее соединение**: открывай сессию один раз на старте приложения и
  переиспользуй (а не поднимай subprocess на каждый запрос);
- результат инструмента нужно превратить в строку для сообщения модели
  (`result.content` / `result.structured_content`);
- корректное закрытие при shutdown.

- [ ] Временный скрипт: подключиться, напечатать имена инструментов, вызвать один.

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

MCP отдаёт `Tool` с полями `name`, `description`, `inputSchema`. Напиши функцию
`mcp_tools_to_ollama(tools)`, которая их конвертирует.

Требования:

- не терять `description` (именно она помогает модели выбрать инструмент);
- не терять `inputSchema` (иначе модель не поймёт аргументы);
- покрыть юнит-тестом на одном примере.

- [ ] Конвертация даёт валидный список для `client.chat(tools=...)`.

## Шаг 4. Агентный цикл (`app/agent/loop.py`)

Сердце задачи. Псевдокод:

```python
async def run_agent(session, question: str, system_prompt: str) -> AgentResult:
    tools = mcp_tools_to_ollama(await session.list_tools())
    messages = [
        {"role": "system", "content": system_prompt},
        {"role": "user", "content": question},
    ]
    used: list[str] = []

    for _ in range(MAX_STEPS):                # защита от бесконечного цикла
        response = await ollama.chat(model=MODEL, messages=messages, tools=tools)
        calls = response.message.tool_calls
        if not calls:
            return AgentResult(answer=response.message.content or "", used_tools=used)

        messages.append(response.message)     # assistant с tool_calls
        for call in calls:
            used.append(call.function.name)
            result = await session.call_tool(call.function.name, dict(call.function.arguments))
            messages.append({
                "role": "tool",
                "tool_name": call.function.name,
                "content": tool_result_to_text(result),
            })

    return AgentResult(answer="Не удалось получить ответ за отведённое число шагов", used_tools=used)
```

Требования:

- **лимит итераций** (`MAX_STEPS`, например 6) — иначе модель может зациклиться;
- неизвестный инструмент / ошибка инструмента → вернуть модели понятную ошибку,
  а не падать;
- `AgentResult` содержит `answer` и `used_tools` (список вызванных инструментов) —
  это понадобится в evals;
- системный промпт: коротко объясни роль («ты помощник по личным расходам; для цифр
  всегда используй инструменты; не выдумывай суммы; отвечай по-русски»).

- [ ] Вопрос «Сколько потрачено в январе 2023 по объекту 1?» → вызов
      `get_monthly_totals` → правильное число.

## Шаг 5. FastAPI (`app/api/`, `app/main.py`)

```python
# app/api/app.py
from fastapi import FastAPI
from pydantic import BaseModel

class ChatRequest(BaseModel):
    message: str
    resident_location_id: int = 1

class ChatResponse(BaseModel):
    answer: str
    used_tools: list[str]

app = FastAPI(title="expense-agent")

@app.get("/health")
def health() -> dict[str, str]:
    return {"status": "ok"}

@app.post("/chat", response_model=ChatResponse)
async def chat(req: ChatRequest) -> ChatResponse:
    ...
```

Требования:

- MCP-сессия открывается в `lifespan` (startup/shutdown), а не в хендлере;
- `POST /chat` возвращает `answer` и `used_tools`;
- корректные коды ошибок (400 на пустой вопрос, 503 если Ollama недоступна);
- Swagger доступен на `/docs`.

`app/main.py`:

```python
import uvicorn
from app.config import load_config

if __name__ == "__main__":
    cfg = load_config()
    uvicorn.run("app.api.app:app", host=cfg.api_host, port=cfg.api_port)
```

- [ ] `nx serve python-expense-agent` поднимает сервис на `:8100`.
- [ ] `curl -s localhost:8100/health` → `{"status":"ok"}`.
- [ ] `curl` на `/chat` отвечает осмысленно.

## Шаг 6. Ручная проверка

Пройди 4–5 вопросов разного типа:

```bash
curl -s localhost:8100/chat -H 'content-type: application/json' \
  -d '{"message":"Сколько всего потрачено в январе 2023?","resident_location_id":1}'

curl -s localhost:8100/chat -H 'content-type: application/json' \
  -d '{"message":"Какая категория самая дорогая в 2023 году?","resident_location_id":1}'

curl -s localhost:8100/chat -H 'content-type: application/json' \
  -d '{"message":"Покажи последние 5 трат","resident_location_id":1}'
```

- [ ] Ответы содержат правильные цифры (сверь с `psql`).
- [ ] `used_tools` соответствует вопросу (агрегат → totals/averages, список → list_expenses).
- [ ] Вопрос «привет» не вызывает инструменты и отвечает текстом.

## Критерии готовности

- [ ] `nx lint` / `nx typecheck` — чисто.
- [ ] Сервис стартует и корректно останавливается (MCP-сессия закрывается).
- [ ] `/chat` отвечает на вопросы трёх типов: агрегат, список, «светская беседа».
- [ ] Модель не выдумывает суммы: если данных нет, инструмент вернул пусто, и агент
      честно говорит об этом.
- [ ] `used_tools` заполняется и логируется (в stderr).

## Подсказки и грабли

- **Модель без tool-calling.** Если `tool_calls` всегда `None`, модель их не
  поддерживает или схема невалидна. Проверь curl-ом из начала файла.
- **Цикл без конца.** Всегда ставь `MAX_STEPS`. Модель может вызывать один инструмент
  повторно.
- **Формат `tool`-сообщения.** У Ollama результат инструмента — отдельное сообщение с
  `role: "tool"`. Если ответ модели пустой/странный — проверь, что ты добавил и
  assistant-сообщение с `tool_calls`, и tool-сообщение с результатом.
- **Один subprocess на запрос — плохо.** Поднимай MCP-сессию в `lifespan`.
- **stdout снова.** MCP-сервер пишет протокол в stdout; не логируй туда.
- **Инъекции.** Аргументы идут в SQL только через параметры `%s` — это уже сделано в
  задаче 01, не ломай.

## Что почитать

- Ollama Python: tool calling (`AsyncClient.chat(tools=...)`, `message.tool_calls`).
- MCP Python SDK: клиент (`Client`, `ClientSession`, `call_tool`).
- FastAPI: `lifespan`, `BaseModel`, `response_model`.
- Про агентные циклы: «ReAct» / «tool use» — общая идея, без привязки к фреймворку.

## Бонус (по желанию)

- Стриминг ответа (`stream=true`) и SSE-endpoint.
- Память диалога (передавать `history` в запросе).
- Инструмент `search_expenses_by_description` на эмбеддингах (`nomic-embed-text`) —
  вот здесь уже появляется классический RAG как **дополнительный** инструмент.
- Ограничение: не отдавать модели `user_id`, если сервис многопользовательский.
