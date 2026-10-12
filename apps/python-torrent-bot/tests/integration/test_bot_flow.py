from __future__ import annotations

from collections.abc import AsyncGenerator
from typing import Any

from aiogram import Bot, Dispatcher
from aiogram.client.session.base import BaseSession
from aiogram.methods import AnswerCallbackQuery

from app.bot import create_dispatcher
from app.models.media import MediaType, Quality
from app.models.torrent import TorrentStatus
from app.providers.fake import FakeTorrentProvider
from app.repositories.memory import InMemoryRepository
from app.services.torrent_service import TorrentService

MAGNET = (
    "magnet:?xt=urn:btih:a1b2c3d4e5f60718293a4b5c6d7e8f9012345678"
    "&dn=The.Matrix.1999.1080p.BluRay"
    "&tr=udp%3A%2F%2Ftracker.example.org%3A1337%2Fannounce"
    "&tr=udp%3A%2F%2Fopen.demonii.com%3A1337%2Fannounce"
    "&xl=12884901888"
)


class RecordingSession(BaseSession):
    """Перехватывает вызовы Telegram API: тесты ходят в сеть только через этот объект."""

    def __init__(self) -> None:
        super().__init__()
        self.methods: list[Any] = []

    async def close(self) -> None:
        return None

    async def make_request(self, bot: Bot, method: Any, timeout: int | None = None) -> Any:
        self.methods.append(method)
        return None

    async def stream_content(
        self,
        url: str,
        headers: dict[str, Any] | None = None,
        timeout: int = 30,
        chunk_size: int = 65536,
        raise_for_status: bool = True,
    ) -> AsyncGenerator[bytes, None]:
        yield b""


def _message_update(update_id: int, text: str) -> dict[str, Any]:
    return {
        "update_id": update_id,
        "message": {
            "message_id": update_id,
            "date": 1790000000,
            "chat": {"id": 1, "type": "private"},
            "from": {"id": 1, "is_bot": False, "first_name": "Tester"},
            "text": text,
        },
    }


def _callback_update(update_id: int, data: str) -> dict[str, Any]:
    return {
        "update_id": update_id,
        "callback_query": {
            "id": str(update_id),
            "from": {"id": 1, "is_bot": False, "first_name": "Tester"},
            "chat_instance": "test-chat-instance",
            "data": data,
            "message": {
                "message_id": 100,
                "date": 1790000000,
                "chat": {"id": 1, "type": "private"},
                "from": {"id": 42, "is_bot": True, "first_name": "Bot"},
                "text": "ui",
            },
        },
    }


def _last_text(session: RecordingSession) -> str:
    for method in reversed(session.methods):
        text = getattr(method, "text", None)
        if isinstance(text, str):
            return text
    return ""


def _last_buttons(session: RecordingSession) -> list[tuple[str, str | None]]:
    for method in reversed(session.methods):
        markup = getattr(method, "reply_markup", None)
        if markup is not None and hasattr(markup, "inline_keyboard"):
            return [
                (button.text, button.callback_data)
                for row in markup.inline_keyboard
                for button in row
            ]
    return []


def _setup() -> tuple[Dispatcher, RecordingSession, Bot, TorrentService]:
    service = TorrentService(FakeTorrentProvider(), InMemoryRepository())
    dp = create_dispatcher(service=service, test_mode=True)
    session = RecordingSession()
    bot = Bot(token="42:TEST", session=session)
    return dp, session, bot, service


async def _wizard_to_confirm(
    dp: Dispatcher, bot: Bot, session: RecordingSession, start_id: int
) -> int:
    uid = start_id
    await dp.feed_raw_update(bot, _message_update(uid, MAGNET))
    uid += 1
    assert "🔗 Torrent найден" in _last_text(session)

    await dp.feed_raw_update(bot, _callback_update(uid, "torrent:category:movie"))
    uid += 1
    assert "Выберите качество" in _last_text(session)

    await dp.feed_raw_update(bot, _callback_update(uid, "torrent:quality:1080"))
    uid += 1
    assert "Готово к добавлению" in _last_text(session)
    assert "🧪 TEST MODE" in _last_text(session)
    assert "💾 Путь:\nMovies/" in _last_text(session)
    return uid


async def test_full_flow_start_to_downloads() -> None:
    dp, session, bot, service = _setup()

    # 1. /start (FEAT-001)
    await dp.feed_raw_update(bot, _message_update(1, "/start"))
    assert "🤖 Torrent Bot" in _last_text(session)
    assert _last_buttons(session) == [
        ("📥 Мои загрузки", "menu:downloads"),
        ("⚙️ Настройки", "menu:settings"),
        ("ℹ️ Помощь", "menu:help"),
    ]

    # 2-4. magnet -> категория -> качество (FEAT-003/004/005)
    next_id = await _wizard_to_confirm(dp, bot, session, 2)

    # 5. подтверждение -> очередь (FEAT-006)
    await dp.feed_raw_update(bot, _callback_update(next_id, "torrent:confirm:add"))
    next_id += 1
    assert "✅ Добавлено в очередь" in _last_text(session)
    tasks = await service.list(1)
    assert len(tasks) == 1
    assert tasks[0].status is TorrentStatus.QUEUED
    assert tasks[0].media_type is MediaType.MOVIE
    assert tasks[0].quality is Quality.FHD_1080
    assert tasks[0].progress == 0.0

    # 6. /downloads (FEAT-007)
    await dp.feed_raw_update(bot, _message_update(next_id, "/downloads"))
    next_id += 1
    text = _last_text(session)
    assert "#001" in text
    assert "The.Matrix.1999.1080p.BluRay" in text
    assert "1080p" in text

    # 7. повторный magnet -> дубликат (FEAT-018)
    next_id = await _wizard_to_confirm(dp, bot, session, next_id)
    await dp.feed_raw_update(bot, _callback_update(next_id, "torrent:confirm:add"))
    assert "уже есть" in _last_text(session)
    assert len(await service.list(1)) == 1


async def test_invalid_magnet_gets_friendly_error() -> None:
    dp, session, bot, _service = _setup()
    await dp.feed_raw_update(bot, _message_update(1, "просто обычный текст"))
    assert "Не удалось распознать" in _last_text(session)


async def test_unknown_callback_is_ignored() -> None:
    dp, session, bot, _service = _setup()
    await dp.feed_raw_update(bot, _callback_update(1, "foo:bar"))
    assert isinstance(session.methods[-1], AnswerCallbackQuery)
