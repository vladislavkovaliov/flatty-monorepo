from __future__ import annotations

from aiogram.types import InlineKeyboardButton, InlineKeyboardMarkup

from app.models.media import MEDIA_LABELS, QUALITY_LABELS, MediaType, Quality


def category_keyboard() -> InlineKeyboardMarkup:
    return InlineKeyboardMarkup(
        inline_keyboard=[
            [
                InlineKeyboardButton(
                    text=f"🎬 {MEDIA_LABELS[MediaType.MOVIE]}",
                    callback_data="torrent:category:movie",
                ),
                InlineKeyboardButton(
                    text=f"📺 {MEDIA_LABELS[MediaType.SERIES]}",
                    callback_data="torrent:category:series",
                ),
            ],
            [InlineKeyboardButton(text="❌ Отмена", callback_data="torrent:confirm:cancel")],
        ]
    )


def quality_keyboard() -> InlineKeyboardMarkup:
    return InlineKeyboardMarkup(
        inline_keyboard=[
            [
                InlineKeyboardButton(
                    text=QUALITY_LABELS[Quality.UHD_2160],
                    callback_data="torrent:quality:2160",
                ),
                InlineKeyboardButton(
                    text=QUALITY_LABELS[Quality.FHD_1080],
                    callback_data="torrent:quality:1080",
                ),
            ],
            [
                InlineKeyboardButton(
                    text=QUALITY_LABELS[Quality.HD_720],
                    callback_data="torrent:quality:720",
                ),
                InlineKeyboardButton(
                    text=QUALITY_LABELS[Quality.UNKNOWN],
                    callback_data="torrent:quality:unknown",
                ),
            ],
        ]
    )


def confirm_keyboard() -> InlineKeyboardMarkup:
    return InlineKeyboardMarkup(
        inline_keyboard=[
            [
                InlineKeyboardButton(
                    text="⬇️ Добавить в очередь",
                    callback_data="torrent:confirm:add",
                )
            ],
            [InlineKeyboardButton(text="✏️ Изменить", callback_data="torrent:confirm:back")],
            [InlineKeyboardButton(text="❌ Отмена", callback_data="torrent:confirm:cancel")],
        ]
    )
