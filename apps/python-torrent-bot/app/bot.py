from __future__ import annotations

import logging

from aiogram import Dispatcher, Router
from aiogram.types import CallbackQuery, ErrorEvent

from app.handlers import downloads, magnet, settings, start
from app.services.torrent_service import TorrentService

logger = logging.getLogger(__name__)


def _fallback_router() -> Router:
    """Последний роутер: гасит неизвестные callback (design: игнор без падения).

    Свой хендлер диспетчера тут нельзя — по propagate_event сначала проверяются
    свои хендлеры, потом sub_routers (интерпретация №8).
    """
    fallback = Router(name="fallback")

    @fallback.callback_query()
    async def unknown_callback(callback: CallbackQuery) -> None:
        logger.debug("ignored unknown callback data=%r", callback.data)
        await callback.answer()

    return fallback


def _include(dp: Dispatcher, router: Router) -> None:
    # aiogram не даёт включать один router в два Dispatcher (в тестах create_dispatcher
    # зовётся многократно) — снимаем привязку предыдущего родителя перед включением
    parent = router.parent_router
    if parent is not None:
        parent.sub_routers.remove(router)
        router._parent_router = None
    dp.include_router(router)


def create_dispatcher(service: TorrentService, test_mode: bool) -> Dispatcher:
    # порядок обязателен: командные роутеры ДО magnet (иначе /downloads съест F.text)
    dp = Dispatcher(torrent_service=service, test_mode=test_mode)
    _include(dp, start.router)
    _include(dp, downloads.router)
    _include(dp, settings.router)
    _include(dp, magnet.router)
    dp.include_router(_fallback_router())

    @dp.error()
    async def on_error(event: ErrorEvent) -> None:
        logger.error("unhandled error: %s", event.exception, exc_info=event.exception)
        update = event.update
        try:
            if update.message is not None:
                await update.message.answer("⚠️ Произошла внутренняя ошибка. Попробуйте ещё раз.")
            elif update.callback_query is not None:
                await update.callback_query.answer("⚠️ Ошибка", show_alert=True)
        except Exception:
            logger.error("failed to send error reply", exc_info=True)

    return dp
