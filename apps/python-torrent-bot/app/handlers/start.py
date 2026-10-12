from __future__ import annotations

from aiogram import Router
from aiogram.filters import Command
from aiogram.fsm.context import FSMContext
from aiogram.types import Message

from app.keyboards.main import main_menu_keyboard
from app.services.torrent_service import TorrentService

router = Router(name="start")

START_TEXT = "🤖 Torrent Bot\n\nОтправьте magnet-ссылку, чтобы создать новую загрузку."


@router.message(Command("start"))
async def cmd_start(message: Message, state: FSMContext, torrent_service: TorrentService) -> None:
    user_id = message.from_user.id if message.from_user else 0
    await torrent_service.register_user(user_id)
    await state.clear()
    await message.answer(START_TEXT, reply_markup=main_menu_keyboard())
