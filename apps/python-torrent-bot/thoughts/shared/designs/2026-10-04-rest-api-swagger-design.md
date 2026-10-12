date: 2026-10-04
topic: "REST API + Swagger UI (Phase 1.5)"
status: validated
---

# REST API + Swagger UI — Design

## Problem Statement

Пользователю нужен интерфейс для UI вне Telegram: публичный HTTP API с автодокументацией Swagger. Фронтенд пишется пользователем отдельно; бот продолжает работать как раньше. В спеке `init.md` HTTP API не описан — это новая feature (020).

## Constraints

- Только REST сейчас; WebSocket — будущая отдельная задача, но дизайн не должен ему мешать.
- Без аутентификации, сервер слушает **только localhost** (`127.0.0.1`) — решение пользователя.
- `TorrentService` и Telegram-слой переписывать нельзя (constraint ROADMAP: расширения без переписывания бизнес-слоя).
- Тесты/линтеры должны оставаться зелёными: `ruff check .`, `ruff format --check .`, `mypy app`, `pytest`.
- Секреты (`DATABASE_URL`, токены) не попадают в ответы API и логи (существующее маскирование сохраняется).

## Approach

**FastAPI + uvicorn** (новые runtime-зависимости `fastapi`, `uvicorn`; dev: `httpx` для TestClient).

- Swagger UI из коробки: `http://127.0.0.1:8080/docs`, схема `/openapi.json`.
- Async-совместим с текущим event loop: uvicorn запускается параллельно с `dp.start_polling` через `asyncio.gather` в одном `main()`.
- Альтернативы отклонены: Starlette вручную (Swagger пришлось бы собирать самим), aiohttp (нет нормального OpenAPI).

## Architecture

Новый слой `app/api/` рядом с `app/handlers/`; оба слоя — клиенты одного `TorrentService` (уже Telegram-free):

- `app/api/schemas.py` — Pydantic-модели входа/выхода и ошибок
- `app/api/routes.py` — маршруты `/api/...`
- `app/api/app.py` — сборка FastAPI-приложения (docs, openapi, роутер)
- `app/main.py` — после инициализации БД/сервиса: `asyncio.gather(start_polling, uvicorn_server.serve())`

## Components

### Config (`.env`)

- `API_HOST` (default `127.0.0.1`), `API_PORT` (default `8080`), `API_USER_ID` (default `1`).
- `API_USER_ID` — пользователь, **от имени которого работает API**: create ставит его как владельца, list/get/mutations передают его в сервис (ownership-check = ровно то поведение, что в Telegram). Несовпадение с твоим Telegram id = пустой список; документируем в CONFIGURATION.md.

### Endpoints

| Метод | Путь | Сервис | Успех |
|---|---|---|---|
| GET | `/api/tasks` | `list(api_user_id)` | 200, список задач |
| GET | `/api/tasks/{id}` | `get(api_user_id, id)` | 200 / 404 |
| POST | `/api/tasks` | `create(user_id=api_user_id, magnet, media_type, quality)` | 201 / 409 дубль / 422 валидация / 400 нет пользователя |
| POST | `/api/tasks/{id}/pause` | `pause` | 200 / 404 / 409 неверный статус |
| POST | `/api/tasks/{id}/resume` | `resume` | 200 / 404 / 409 |
| POST | `/api/tasks/{id}/cancel` | `cancel` | 200 / 404 / 409 |
| DELETE | `/api/tasks/{id}` | `delete` | 204 / 404 |

- Схемы: `TaskCreateIn {magnet, media_type, quality}`, `TorrentTaskOut` (все поля доменной модели, строки вместо Enum), `ErrorOut {detail}`.
- Ошибки: валидация Pydantic → 422 (стандарт FastAPI); «пользователь не найден» при create → 400; внутренняя → 500 + `logger.error(..., exc_info=True)`.

### WebSocket (future)

В feature-docе раздел «Future: `WS /api/ws` события задач» — добавляется позже как отдельная задача, REST не меняется.

## Data Flow

`HTTP request → FastAPI route (валидация) → TorrentService → repository/provider → JSON-ответ`. Состояние в PostgreSQL общее с ботом: задача, созданная через API, видна в `/downloads`, и наоборот.

## Error Handling

- Сервисные ошибки маппятся в HTTP-коды (404/409/400), не раскрывая внутренности (traceback — только в лог).
- Если `users` пуст (никто не жал `/start`) — create отвечает 400 с инструкцией.
- Падение uvicorn-сервера не роняет бота: `gather` с `return_exceptions`, ошибка логируется, polling продолжается (и наоборот).

## Testing Strategy

- `tests/integration/test_api.py`: TestClient + `InMemoryRepository` — флоу create→list→pause→resume→cancel→delete, дубль → 409, 404, 422, `GET /openapi.json` → 200, `/docs` → 200.
- `tests/unit/test_config.py`: +2 (API_HOST/PORT/API_USER_ID дефолты и парсинг).
- Финальная верификация: `ruff check . && ruff format --check . && mypy app && pytest`.

## Docs

- `docs/features/020-rest-api.md` (шаблон §19; Status: Implemented по завершении).
- `docs/decisions/0004-http-api.md`: FastAPI, localhost-only без auth, API_USER_ID как актор, WebSocket отложен.
- `docs/CONFIGURATION.md`: 3 новые переменные; `docs/ROADMAP.md`: запись «Phase 1.5 — REST API»; `README.md`: секция «API» (адрес Swagger).
- `docs/API.md`: раздел «HTTP API» со ссылкой на `/docs`.

## Open Questions

Нет — решения приняты пользователем: Swagger обязателен, фронт свой, auth = localhost, WebSocket отложен, `API_USER_ID` остаётся.
