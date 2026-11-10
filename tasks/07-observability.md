# Задача 07. Observability: трейсы и метрики агента

> Тема: **OpenTelemetry** для LLM-агента — видеть, что происходит внутри одного запроса.
> Результат: каждый `/chat` даёт трейс в Jaeger: вызов LLM, каждый tool-call, ошибки,
> задержки; плюс метрики.

---

## Цель обучения

К концу задачи ты должен уметь:

- инструментировать async-приложение через OpenTelemetry Python SDK;
- строить **иерархию спанов**: `agent.run` → `llm.chat` / `tool.call`;
- добавлять полезные атрибуты (модель, инструмент, длительность, статус);
- различать tracing и metrics и знать, что куда;
- не утекать PII в атрибуты.

## Контекст проекта

В репозитории уже есть OTel: go-api шлёт трейсы в Jaeger (OTLP HTTP
`http://localhost:4318/v1/traces`). Есть `tracing.Init()`/`Shutdown()` в Go — посмотри
как образец. Мы сделаем то же на Python, чтобы трейсы агента были рядом с трейсами
бэкенда.

---

## Шаг 0. Зависимости

```
opentelemetry-api
opentelemetry-sdk
opentelemetry-exporter-otlp-proto-http
opentelemetry-instrumentation-fastapi
opentelemetry-instrumentation-httpx
```

Проверь, что Jaeger поднят (`docker-compose` / `podman-compose`) и принимает OTLP.

- [ ] Jaeger UI открывается (обычно `http://localhost:16686`).

## Шаг 1. Инициализация (`app/observability.py`)

```python
from opentelemetry import trace
from opentelemetry.sdk.resources import Resource
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter

def init_tracing(service_name: str = "python-expense-agent") -> None:
    provider = TracerProvider(resource=Resource.create({"service.name": service_name}))
    provider.add_span_processor(
        BatchSpanProcessor(OTLPSpanExporter(endpoint="http://localhost:4318/v1/traces"))
    )
    trace.set_tracer_provider(provider)
```

Требования:

- endpoint из конфига (`OTEL_EXPORTER_OTLP_TRACES_ENDPOINT`), как в go-api;
- `BatchSpanProcessor` (не Simple) — не блокировать запросы;
- graceful shutdown (flush при остановке) — по образцу `tracing.Shutdown()`;
- если экспортёр недоступен — приложение **не должно падать** (лог + продолжаем).

- [ ] Стартует без Jaeger (просто без трейсов).

## Шаг 2. Авто-инструментация FastAPI

```python
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor
FastAPIInstrumentor.instrument_app(app)
```

Плюс `HTTPXClientInstrumentor` — чтобы видеть исходящие запросы к Ollama.

- [ ] В Jaeger появляется трейс на `GET /health` и `POST /chat`.

## Шаг 3. Ручные спаны агента

Оберни ключевые операции. Ориентир:

```python
tracer = trace.get_tracer("expense-agent")

async def run_agent(...):
    with tracer.start_as_current_span("agent.run") as span:
        span.set_attribute("agent.question.length", len(question))
        ...
        with tracer.start_as_current_span("llm.chat") as llm_span:
            llm_span.set_attribute("llm.model", MODEL)
            response = await client.chat(...)
            llm_span.set_attribute("llm.tool_calls", len(response.message.tool_calls))
        for call in response.message.tool_calls:
            with tracer.start_as_current_span("tool.call") as tool_span:
                tool_span.set_attribute("tool.name", call.function.name)
                try:
                    result = await session.call_tool(...)
                    tool_span.set_attribute("tool.status", "ok")
                except Exception as e:
                    tool_span.record_exception(e)
                    tool_span.set_status(trace.Status(trace.StatusCode.ERROR, str(e)))
```

Требования:

- спан `agent.run` — корень запроса;
- `llm.chat` на **каждый** вызов модели (их может быть несколько за один запрос);
- `tool.call` на каждый инструмент, с именем и статусом;
- ошибки: `record_exception` + `set_status(ERROR)`;
- **не пиши PII** (полные описания трат, токены) в атрибуты; пиши длину/имя/счётчики.

- [ ] В Jaeger у одного `/chat` видно дерево: `agent.run` → несколько `llm.chat` и
      `tool.call`.

## Шаг 4. Метрики

Tracing отвечает «что произошло в этом запросе», metrics — «как в среднем».
Добавь счётчики/гистограммы:

- `agent.requests` (counter, теги: `status`);
- `llm.latency` (histogram, тег: `model`);
- `tool.latency` (histogram, тег: `tool`);
- `agent.steps` (histogram — сколько итераций занял запрос).

Экспорт — на выбор: OTLP metrics или Prometheus (`/metrics`). Не обязательно поднимать
полный стек; достаточно, чтобы метрики были видны.

- [ ] После нескольких запросов метрики содержат данные.
- [ ] Видно, какой инструмент самый медленный.

## Шаг 5. Проверка

1. Сделай 3–5 запросов к `/chat` (разные вопросы).
2. Открой Jaeger → сервис `python-expense-agent`.
3. Найди трейс с несколькими `llm.chat` (там, где были tool-calls).
4. Спровоцируй ошибку (останови Ollama / передай плохой аргумент) → проверь ERROR-спан.

- [ ] Трейсы читаемы, имена спанов осмысленны.
- [ ] Ошибка видна в трейсе, а не только в логах.
- [ ] В атрибутах нет сырых пользовательских данных.

## Критерии готовности

- [ ] `nx lint` / `nx typecheck` — чисто.
- [ ] Трейс на каждый `/chat` с вложенными `llm.chat` и `tool.call`.
- [ ] Ошибки инструмента/LLM помечают спан ERROR.
- [ ] Есть метрики задержек (LLM и tools).
- [ ] Приложение работает без Jaeger.
- [ ] Shutdown не теряет последние спаны (flush).

## Подсказки и грабли

- **Контекст в asyncio.** `start_as_current_span` работает с contextvars; при создании
  фоновых задач используй `trace.use_span(..., end_on_exit=False)` или передавай
  контекст явно, иначе спаны «оторвутся».
- **Блокирующий экспортёр.** `SimpleSpanProcessor` тормозит запросы — только `Batch`.
- **PII.** Не клади `question` целиком, `description`, суммы пользователя в атрибуты —
  это утечка. Длины и имена — ок.
- **Кардинальность.** Не делай тегом `session_id`/`question` — взорвёшь метрики.
- **Дубли FastAPI-инструментации.** Инструментируй один раз при старте.
- **Sampling.** На старте `AlwaysOn` ок; в проде — вероятностный сэмплинг.

## Что почитать

- OpenTelemetry Python: TracerProvider, spans, attributes, status.
- OTLP HTTP exporter, BatchSpanProcessor.
- Jaeger UI: поиск по сервису, дерево спанов.
- Разница traces / metrics / logs.

## Бонус (по желанию)

- Prometheus + Grafana дашборд (latency p50/p95, error rate).
- LLM-специфичные атрибуты: prompt/completion tokens, стоимость.
- Langfuse / Arize Phoenix как «LLM observability» поверх OTel.
- Трейс-корреляция: прокидывать `trace_id` в ответ `/chat` для поддержки.
