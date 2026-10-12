from __future__ import annotations

from aiogram.types import InlineKeyboardButton, InlineKeyboardMarkup

from app.models.torrent import TorrentStatus, TorrentTask


def downloads_keyboard(tasks: list[TorrentTask]) -> InlineKeyboardMarkup:
    rows: list[list[InlineKeyboardButton]] = []
    for task in tasks:
        prefix = f"#{task.id:03d}"
        row = [
            InlineKeyboardButton(
                text=f"{prefix} 📊 Открыть", callback_data=f"torrent:open:{task.id}"
            )
        ]
        if task.status in (TorrentStatus.QUEUED, TorrentStatus.DOWNLOADING):
            row += [
                InlineKeyboardButton(
                    text=f"{prefix} ⏸ Пауза",
                    callback_data=f"torrent:pause:{task.id}",
                ),
                InlineKeyboardButton(
                    text=f"{prefix} ❌ Отменить",
                    callback_data=f"torrent:cancel:{task.id}",
                ),
            ]
        elif task.status is TorrentStatus.PAUSED:
            row += [
                InlineKeyboardButton(
                    text=f"{prefix} ▶️ Продолжить",
                    callback_data=f"torrent:resume:{task.id}",
                ),
                InlineKeyboardButton(
                    text=f"{prefix} ❌ Отменить",
                    callback_data=f"torrent:cancel:{task.id}",
                ),
            ]
        else:
            row.append(
                InlineKeyboardButton(
                    text=f"{prefix} 🗑 Удалить",
                    callback_data=f"torrent:delete:{task.id}",
                )
            )
        rows.append(row)
    return InlineKeyboardMarkup(inline_keyboard=rows)


def detail_keyboard(task: TorrentTask) -> InlineKeyboardMarkup:
    markup = downloads_keyboard([task])
    rows = list(markup.inline_keyboard)
    rows.append([InlineKeyboardButton(text="◀️ К списку", callback_data="torrent:list:back")])
    return InlineKeyboardMarkup(inline_keyboard=rows)
