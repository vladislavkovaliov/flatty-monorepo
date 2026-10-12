# ADR-0004: HTTP API на FastAPI, localhost без auth, API_USER_ID как актор

## Status

Accepted

## Context

Нужен UI вне Telegram поверх готового `TorrentService` (без Telegram-кода,
design Phase 1). Нужна автодокументация (пользователь пишет свой фронтенд).
Проект в TEST_MODE: реальных загрузок нет, слушать наружу опасно.

## Decision

- **FastAPI + uvicorn** (runtime-зависимости): Swagger UI и OpenAPI из коробки;
  альтернативы (Starlette вручную, aiohttp) отклонены — нет нормальной автодоки.
- **Только localhost, без аутентификации** — config fail-fast на не-localhost API_HOST.
- **`API_USER_ID` — актор API**: create ставит его владельцем, остальные операции
  передают его в сервис; ownership-проверка = ровно поведение Telegram (FEAT-009).
- **`/start` регистрирует пользователя** (`ensure_user`): до этого `users` создавался
  только вместе с первой задачей/настройкой, и «нажми /start» был бы нерабочей петлёй.
- **Параллельно с polling**: uvicorn `Server.serve()` и `dp.start_polling` в одном
  event loop через `asyncio.gather(..., return_exceptions=True)`; порядок аргументов
  (polling первым) обязателен из-за перехвата signal-хендлеров uvicorn'ом.
- **WebSocket отложен** — будущая отдельная задача, REST не меняется.

## Consequences

Плюсы:

- Swagger UI без дополнительной инфраструктуры; фронтенд пользователя кодогенерится из openapi.json.
- Общая PostgreSQL: задачи видны и в боте, и в API; бизнес-слой не переписан.

Минусы:

- Нет auth — нельзя слушать не-localhost (охраняется config-валидацией).
- `API_USER_ID` ≠ вашего telegram id = пустой список (документировано в CONFIGURATION.md).
- Сбой provider при delete маппится в 404 (сервис не различает причины False).
