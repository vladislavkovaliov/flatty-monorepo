from __future__ import annotations

import contextlib
import logging

from aiogram import F, Router
from aiogram.exceptions import TelegramBadRequest
from aiogram.fsm.context import FSMContext
from aiogram.fsm.state import State, StatesGroup
from aiogram.types import CallbackQuery, InlineKeyboardMarkup, Message

from app.keyboards.main import main_menu_keyboard
from app.keyboards.torrent import category_keyboard, confirm_keyboard, quality_keyboard
from app.models.media import MEDIA_LABELS, QUALITY_LABELS, SAVE_PATHS, MediaType, Quality
from app.models.torrent import TorrentStatus, format_size
from app.services.magnet_parser import MagnetInfo, MagnetParseError, parse_magnet
from app.services.torrent_service import TorrentService

logger = logging.getLogger(__name__)

router = Router(name="magnet")


class TorrentWizard(StatesGroup):
    choosing_category = State()
    choosing_quality = State()
    confirming = State()


_QUALITY_BY_CALLBACK: dict[str, Quality] = {
    "2160": Quality.UHD_2160,
    "1080": Quality.FHD_1080,
    "720": Quality.HD_720,
    "unknown": Quality.UNKNOWN,
}

INVALID_MAGNET_TEXT = (
    "❌ Не удалось распознать magnet-ссылку.\n\nОтправьте ссылку вида:\nmagnet:?xt=urn:btih:..."
)


def _info_text(info: MagnetInfo) -> str:
    name = info.display_name or "не определено"
    return (
        "🔗 Torrent найден\n\n"
        f"🎬 Название:\n{name}\n\n"
        f"🔑 Hash:\n{info.info_hash[:10]}...\n\n"
        f"📡 Trackers:\n{len(info.trackers)}\n\n"
        f"💾 Размер:\n{format_size(info.size_bytes)}\n\n"
        "Куда сохранить?"
    )


def _summary_text(
    info: MagnetInfo, media_type: MediaType, quality: Quality, test_mode: bool
) -> str:
    text = (
        "✅ Готово к добавлению\n\n"
        f"🎬 Название:\n{info.display_name or 'не определяло'}\n\n"
        f"📁 Категория:\n{MEDIA_LABELS[media_type]}\n\n"
        f"🎞 Качество:\n{QUALITY_LABELS[quality]}\n\n"
        f"💾 Путь:\n{SAVE_PATHS[media_type]}\n"
    )
    if test_mode:
        text += "\n🧪 TEST MODE\n\nФайл реально скачиваться НЕ будет."
    return text


async def _edit(callback: CallbackQuery, text: str, keyboard: InlineKeyboardMarkup) -> None:
    # намеренная дупликация с downloads.py — отдельный файл утилит вне скоупа
    if not isinstance(callback.message, Message):
        return
    with contextlib.suppress(TelegramBadRequest):
        await callback.message.edit_text(text, reply_markup=keyboard)


@router.message(F.text)
async def on_text(message: Message, state: FSMContext) -> None:
    try:
        info = parse_magnet(message.text or "")
    except MagnetParseError:
        user_id = message.from_user.id if message.from_user else 0
        logger.info("invalid magnet from user_id=%s", user_id)
        await state.clear()
        await message.answer(INVALID_MAGNET_TEXT)
        return
    logger.debug("magnet received: %s", info.info_hash)
    await state.set_state(TorrentWizard.choosing_category)
    await state.update_data(magnet=info)
    await message.answer(_info_text(info), reply_markup=category_keyboard())


@router.callback_query(TorrentWizard.choosing_category, F.data.startswith("torrent:category:"))
async def on_category(callback: CallbackQuery, state: FSMContext) -> None:
    value = (callback.data or "").rsplit(":", 1)[-1]
    try:
        media_type = MediaType(value)
    except ValueError:
        await callback.answer("Неизвестная категория", show_alert=True)
        return
    await state.update_data(media_type=media_type)
    await state.set_state(TorrentWizard.choosing_quality)
    await callback.answer()
    await _edit(callback, "Выберите качество:", quality_keyboard())


@router.callback_query(TorrentWizard.choosing_quality, F.data.startswith("torrent:quality:"))
async def on_quality(callback: CallbackQuery, state: FSMContext, test_mode: bool) -> None:
    value = (callback.data or "").rsplit(":", 1)[-1]
    quality = _QUALITY_BY_CALLBACK.get(value)
    if quality is None:
        await callback.answer("Неизвестное качество", show_alert=True)
        return
    data = await state.get_data()
    info: MagnetInfo = data["magnet"]
    media_type: MediaType = data["media_type"]
    await state.update_data(quality=quality)
    await state.set_state(TorrentWizard.confirming)
    await callback.answer()
    await _edit(callback, _summary_text(info, media_type, quality, test_mode), confirm_keyboard())


@router.callback_query(TorrentWizard.confirming, F.data == "torrent:confirm:add")
async def on_confirm_add(
    callback: CallbackQuery, state: FSMContext, torrent_service: TorrentService
) -> None:
    data = await state.get_data()
    info: MagnetInfo = data["magnet"]
    media_type: MediaType = data["media_type"]
    quality: Quality = data["quality"]
    task, created = await torrent_service.create(
        user_id=callback.from_user.id,
        magnet=info,
        media_type=media_type,
        quality=quality,
    )
    await state.clear()
    if task.status is TorrentStatus.FAILED:
        text = "⚠️ Не удалось добавить загрузку. Попробуйте позже."
    elif created:
        text = f"✅ Добавлено в очередь\n\n#{task.id:03d} {task.name}\n\nСмотрите /downloads"
    else:
        text = f"ℹ️ Такая загрузка уже есть\n\n#{task.id:03d} {task.name}\n\nСмотрите /downloads"
    await callback.answer()
    await _edit(callback, text, main_menu_keyboard())


@router.callback_query(TorrentWizard.confirming, F.data == "torrent:confirm:back")
async def on_confirm_back(callback: CallbackQuery, state: FSMContext) -> None:
    data = await state.get_data()
    info: MagnetInfo = data["magnet"]
    # «Изменить» → назад к экрану категории (интерпретация №4)
    await state.set_state(TorrentWizard.choosing_category)
    await callback.answer()
    await _edit(callback, _info_text(info), category_keyboard())


@router.callback_query(F.data == "torrent:confirm:cancel")
async def on_cancel(callback: CallbackQuery, state: FSMContext) -> None:
    if await state.get_state() is None:
        await callback.answer()
        return
    await state.clear()
    await callback.answer()
    await _edit(callback, "❌ Добавление отменено.", main_menu_keyboard())
