from __future__ import annotations

import asyncio
import logging
import logging.handlers
from pathlib import Path

import uvicorn
from aiogram import Bot, Dispatcher

from app.api.app import create_api_app
from app.bot import create_dispatcher
from app.config import load_config
from app.db.pool import apply_schema, create_pool, redact_database_url
from app.providers.fake import FakeTorrentProvider
from app.repositories.postgres import PostgresRepository
from app.services.torrent_service import TorrentService

logger = logging.getLogger(__name__)


async def _run_polling(dp: Dispatcher, bot: Bot, server: uvicorn.Server) -> None:
    try:
        await dp.start_polling(bot)
    finally:
        # polling завершился (штатно или с ошибкой) — гасим API-сервер, иначе gather ждёт вечно
        server.should_exit = True


async def _run_api(server: uvicorn.Server) -> None:
    try:
        await server.serve()
    except SystemExit as exc:
        # uvicorn sys.exit(STARTUP_FAILURE) при неудачном bind (занятый порт) —
        # SystemExit не Exception, ловим явно, иначе роняет весь процесс
        logger.error("API server exited with code %s; bot polling continues", exc.code)
    except Exception:
        # падение сервера НЕ роняет бота: логируем, polling продолжается (design §Error Handling)
        logger.error("API server failed; bot polling continues", exc_info=True)


async def main() -> None:
    config = load_config()
    format_ = "%(asctime)s %(levelname)s %(name)s: %(message)s"
    logging.basicConfig(level=config.log_level, format=format_)
    Path("logs").mkdir(exist_ok=True)
    file_handler = logging.handlers.RotatingFileHandler(
        "logs/bot.log", maxBytes=2_000_000, backupCount=3, encoding="utf-8"
    )
    file_handler.setLevel(config.log_level)
    file_handler.setFormatter(logging.Formatter(format_))
    logging.getLogger().addHandler(file_handler)
    if config.log_level == "DEBUG":
        # DEBUG не должен заливать логи polling-внутренностями
        logging.getLogger("aiogram").setLevel(logging.INFO)
        logging.getLogger("psycopg").setLevel(logging.INFO)
    # fail fast: нет DATABASE_URL -> RuntimeError из load_config; БД недоступна ->
    # RuntimeError из create_pool (redacted URL).
    # Никакого тихого memory-режима (design §Error Handling)
    pool = await create_pool(config.database_url)
    try:
        await apply_schema(pool)
        repository = PostgresRepository(pool)
        service = TorrentService(FakeTorrentProvider(), repository)
        dp: Dispatcher = create_dispatcher(service, config.test_mode)
        bot = Bot(token=config.bot_token)
        api_app = create_api_app(service, repository, config.api_user_id)
        server = uvicorn.Server(
            uvicorn.Config(api_app, host=config.api_host, port=config.api_port, log_config=None)
        )
        logger.info(
            "bot started (TEST_MODE=%s, db=%s, api=http://%s:%s/docs)",
            config.test_mode,
            redact_database_url(config.database_url),
            config.api_host,
            config.api_port,
        )
        # Порядок gather обязателен: aiogram первым ставит signal-хендлеры
        # (loop.add_signal_handler), uvicorn вторым перехватывает их через signal.signal
        # и при выходе восстанавливает + пересылает сигнал (uvicorn Server.capture_signals):
        # Ctrl+C гасит API, затем polling.
        # return_exceptions: падение одного не отменяет другой (design §Error Handling).
        results = await asyncio.gather(
            _run_polling(dp, bot, server),
            _run_api(server),
            return_exceptions=True,
        )
        for result in results:
            if isinstance(result, BaseException):
                raise result  # всплытие ошибки polling (design: «и наоборот» — но бот первичен)
    finally:
        await pool.close()


if __name__ == "__main__":
    asyncio.run(main())
