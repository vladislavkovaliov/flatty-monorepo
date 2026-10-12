from __future__ import annotations

from aiogram.types import InlineKeyboardButton, InlineKeyboardMarkup


def main_menu_keyboard() -> InlineKeyboardMarkup:
    return InlineKeyboardMarkup(
        inline_keyboard=[
            [InlineKeyboardButton(text="📥 Мои загрузки", callback_data="menu:downloads")],
            [InlineKeyboardButton(text="⚙️ Настройки", callback_data="menu:settings")],
            [InlineKeyboardButton(text="ℹ️ Помощь", callback_data="menu:help")],
        ]
    )
