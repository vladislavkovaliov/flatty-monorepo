from __future__ import annotations

from aiogram import F, Router
from aiogram.types import CallbackQuery

router = Router(name="settings")


@router.callback_query(F.data == "menu:settings")
async def on_settings(callback: CallbackQuery) -> None:
    await callback.answer("⚙️ Настройки появятся позже")


@router.callback_query(F.data == "menu:help")
async def on_help(callback: CallbackQuery) -> None:
    await callback.answer(
        "Отправьте magnet-ссылку → выберите категорию и качество → подтвердите. "
        "Все загрузки: /downloads"
    )
