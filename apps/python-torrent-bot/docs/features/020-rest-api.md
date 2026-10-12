# FEAT-020 — REST API + Swagger UI

## Status

Implemented

## Goal

Дать пользователю интерфейс для UI вне Telegram: публичный локальный HTTP API
с автодокументацией Swagger, поверх уже существующего `TorrentService`.

## User Story

As a user, I want to manage my downloads through a local HTTP API with Swagger docs,
so I can build my own UI outside Telegram.

## Scope

- FastAPI + uvicorn, параллельно с Telegram-polling (один процесс, `asyncio.gather`).
- Эндпоинты `/api/tasks`: list/get/create/pause/resume/cancel/delete.
- Swagger UI (`/docs`) и OpenAPI (`/openapi.json`) из коробки.
- `API_HOST`/`API_PORT`/`API_USER_ID` в config; регистрация пользователя на `/start`.

## Out of Scope

- Аутентификация (API слушает только localhost — ADR-0004).
- WebSocket (см. «Future» ниже).
- Изменения `TorrentService`/Telegram UI (кроме регистрации на /start).

## User Flow

```text
HTTP request → FastAPI route (валидация) → TorrentService → repository/provider → JSON
```

Задача, созданная через API, видна в `/downloads` бота, и наоборот (общая PostgreSQL).

## UI

Не меняется. Swagger UI: `http://127.0.0.1:8080/docs`.

## Domain Changes

Нет новых доменных моделей; API-схемы (`TaskCreateIn`, `TorrentTaskOut`, `ErrorOut`)
отражают существующие `TorrentTask`/enum'ы (StrEnum → plain strings в JSON).

## API / Interfaces

`app/api/`: `create_api_app(service, repository, api_user_id) -> FastAPI`;
роуты поверх `TorrentService`; repository получил `ensure_user`/`user_exists`.

## Storage Changes

Нет (та же PostgreSQL). Строка `users` для `API_USER_ID` создаётся при `/start`.

## Security

- Без аутентификации → только localhost, fail-fast в config (API_HOST).
- Секреты не попадают в ответы/логи.

## Error Handling

- 422 — валидация Pydantic; 400 — невалидный magnet / пользователь не нажал /start;
  404 — не найдено/чужое; 409 — дубль (detail = существующая задача) и неверный статус;
  500 — неожиданная ошибка (лог с traceback, тело `{"detail": "internal error"}`).
- Падение uvicorn не роняет polling (логируется); завершение polling гасит сервер,
  ошибка polling всплывает.

## Tests

`tests/integration/test_api.py` (TestClient + InMemoryRepository): флоу,
дубль→409, 404, 422, 400×2, 409 неверный статус, `/openapi.json`, `/docs`;
`tests/unit/test_config.py` — дефолты/парсинг API-переменных.

## Dependencies

- FEAT-001 — FEAT-010 (сервис, репозитории, PostgreSQL).

## Configuration

`API_HOST` (default `127.0.0.1`, только localhost), `API_PORT` (default `8080`),
`API_USER_ID` (default `1`) — актор API.

## Implementation Tasks

- [x] Зависимости (fastapi, uvicorn; httpx dev)
- [x] Config: API_HOST/API_PORT/API_USER_ID + fail-fast localhost
- [x] Repository: `ensure_user`/`user_exists`; `/start` регистрирует пользователя
- [x] `app/api/`: schemas + routes + `create_api_app`
- [x] `app/main.py`: uvicorn параллельно с `dp.start_polling` (gather, return_exceptions)
- [x] Тесты: `tests/integration/test_api.py`, `tests/unit/test_config.py` (+4)
- [ ] Future: `WS /api/ws` события задач — отдельная задача, REST не меняется

## Acceptance Criteria

- [x] `GET/POST /api/tasks`, pause/resume/cancel/delete отвечают кодами design §Endpoints
- [x] Swagger UI доступен на `/docs`, схема на `/openapi.json`
- [x] Задачи общие с ботом (общая БД, ownership = API_USER_ID)
- [x] Бот и API работают одновременно; падение одного не убивает другое молча
- [x] `ruff`/`mypy`/`pytest` зелёные

## Related Documentation

- `docs/API.md` §7
- `docs/decisions/0004-http-api.md`
- `docs/CONFIGURATION.md`
- `docs/ROADMAP.md`

## Future: `WS /api/ws` события задач

WebSocket для live-обновлений прогресса добавляется позже отдельной задачей;
REST-контракт не меняется.
