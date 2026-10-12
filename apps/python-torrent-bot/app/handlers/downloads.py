from __future__ import annotations

import contextlib
import logging
from collections.abc import Awaitable, Callable

from aiogram import F, Router
from aiogram.exceptions import TelegramBadRequest
from aiogram.filters import Command
from aiogram.types import CallbackQuery, InlineKeyboardMarkup, Message

from app.keyboards.downloads import detail_keyboard, downloads_keyboard
from app.keyboards.main import main_menu_keyboard
from app.models.media import MEDIA_EMOJI, QUALITY_LABELS
from app.models.torrent import TorrentStatus, TorrentTask, format_size
from app.services.torrent_service import TorrentService

logger = logging.getLogger(__name__)

router = Router(name="downloads")

_LIST_HEADER = "📥 Мои загрузки"
_EMPTY_TEXT = "📥 Мои загрузки\n\nУ вас пока нет загрузок."

_STATUS_LINE: dict[TorrentStatus, str] = {
    TorrentStatus.QUEUED: "🕓 В очереди",
    TorrentStatus.PAUSED: "⏸ На паузе",
    TorrentStatus.COMPLETED: "✅ Завершено",
    TorrentStatus.FAILED: "❌ Ошибка",
    TorrentStatus.CANCELLED: "🚫 Отменено",
}

_STATUS_WORD: dict[TorrentStatus, str] = {
    TorrentStatus.QUEUED: "🕓 Queued",
    TorrentStatus.DOWNLOADING: "⏳ Downloading",
    TorrentStatus.PAUSED: "⏸ Paused",
    TorrentStatus.COMPLETED: "✅ Completed",
    TorrentStatus.FAILED: "❌ Failed",
    TorrentStatus.CANCELLED: "🚫 Cancelled",
}


def _status_line(task: TorrentTask) -> str:
    if task.status is TorrentStatus.DOWNLOADING:
        return f"⏳ {task.progress:.0f}%"
    return _STATUS_LINE[task.status]


def _progress_bar(progress: float) -> str:
    filled = int(progress / 100 * 16)
    return "█" * filled + "░" * (16 - filled) + f" {progress:.0f}%"


def _detail_text(task: TorrentTask) -> str:
    parts: list[str] = [_STATUS_WORD[task.status], "", task.name]
    if task.progress > 0 or task.status in (
        TorrentStatus.DOWNLOADING,
        TorrentStatus.COMPLETED,
    ):
        parts += ["", _progress_bar(task.progress), "", f"{task.progress:.0f}%"]
    if task.status is TorrentStatus.DOWNLOADING:
        parts += ["", "Speed: TEST", "ETA: TEST"]
    if task.size_bytes is not None:
        parts += ["", f"💾 {format_size(task.size_bytes)}"]
    parts += ["", "🧪 TEST MODE"]
    return "\n".join(parts)


async def _render_list(service: TorrentService, user_id: int) -> tuple[str, InlineKeyboardMarkup]:
    tasks = await service.list(user_id)
    if not tasks:
        return _EMPTY_TEXT, main_menu_keyboard()
    blocks = [
        f"#{task.id:03d}\n"
        f"{MEDIA_EMOJI[task.media_type]} {task.name}\n"
        f"{QUALITY_LABELS[task.quality]}\n"
        f"{_status_line(task)}"
        for task in tasks
    ]
    return f"{_LIST_HEADER}\n\n" + "\n\n".join(blocks), downloads_keyboard(tasks)


async def _edit(callback: CallbackQuery, text: str, keyboard: InlineKeyboardMarkup) -> None:
    # намеренная дупликация с magnet.py — отдельный файл утилит вне скоупа
    if not isinstance(callback.message, Message):
        return
    with contextlib.suppress(TelegramBadRequest):
        await callback.message.edit_text(text, reply_markup=keyboard)


def _parse_id(data: str) -> int | None:
    try:
        return int(data.rsplit(":", 1)[1])
    except (IndexError, ValueError):
        return None


@router.message(Command("downloads"))
async def cmd_downloads(message: Message, torrent_service: TorrentService) -> None:
    user_id = message.from_user.id if message.from_user else 0
    text, keyboard = await _render_list(torrent_service, user_id)
    await message.answer(text, reply_markup=keyboard)


@router.callback_query(F.data == "menu:downloads")
async def on_menu_downloads(callback: CallbackQuery, torrent_service: TorrentService) -> None:
    text, keyboard = await _render_list(torrent_service, callback.from_user.id)
    await callback.answer()
    await _edit(callback, text, keyboard)


@router.callback_query(F.data.startswith("torrent:open:"))
async def on_open(callback: CallbackQuery, torrent_service: TorrentService) -> None:
    task_id = _parse_id(callback.data or "")
    task = (
        await torrent_service.get(callback.from_user.id, task_id) if task_id is not None else None
    )
    if task is None:
        logger.info("task %s not found for user_id=%s", task_id, callback.from_user.id)
        await callback.answer("Задача не найдена", show_alert=True)
        return
    await callback.answer()
    await _edit(callback, _detail_text(task), detail_keyboard(task))


@router.callback_query(F.data == "torrent:list:back")
async def on_back_to_list(callback: CallbackQuery, torrent_service: TorrentService) -> None:
    text, keyboard = await _render_list(torrent_service, callback.from_user.id)
    await callback.answer()
    await _edit(callback, text, keyboard)


async def _apply_and_render(
    callback: CallbackQuery,
    torrent_service: TorrentService,
    operation: Callable[[int, int], Awaitable[TorrentTask | None]],
    success_text: str,
) -> None:
    task_id = _parse_id(callback.data or "")
    task = await operation(callback.from_user.id, task_id) if task_id is not None else None
    if task is None:
        # FEAT-009: чужая/несуществующая — «не найдено», без раскрытия
        await callback.answer("Задача не найдена", show_alert=True)
        return
    if task.status is TorrentStatus.FAILED:
        await callback.answer("⚠️ Ошибка операции", show_alert=True)
    else:
        await callback.answer(success_text)
    text, keyboard = await _render_list(torrent_service, callback.from_user.id)
    await _edit(callback, text, keyboard)


@router.callback_query(F.data.startswith("torrent:pause:"))
async def on_pause(callback: CallbackQuery, torrent_service: TorrentService) -> None:
    await _apply_and_render(callback, torrent_service, torrent_service.pause, "⏸ На паузе")


@router.callback_query(F.data.startswith("torrent:resume:"))
async def on_resume(callback: CallbackQuery, torrent_service: TorrentService) -> None:
    await _apply_and_render(callback, torrent_service, torrent_service.resume, "▶️ Продолжено")


@router.callback_query(F.data.startswith("torrent:cancel:"))
async def on_cancel_task(callback: CallbackQuery, torrent_service: TorrentService) -> None:
    await _apply_and_render(callback, torrent_service, torrent_service.cancel, "❌ Отменено")


@router.callback_query(F.data.startswith("torrent:delete:"))
async def on_delete_task(callback: CallbackQuery, torrent_service: TorrentService) -> None:
    task_id = _parse_id(callback.data or "")
    removed = (
        await torrent_service.delete(callback.from_user.id, task_id)
        if task_id is not None
        else False
    )
    if not removed:
        await callback.answer("Задача не найдена", show_alert=True)
        return
    await callback.answer("🗑 Удалено")
    text, keyboard = await _render_list(torrent_service, callback.from_user.id)
    await _edit(callback, text, keyboard)
