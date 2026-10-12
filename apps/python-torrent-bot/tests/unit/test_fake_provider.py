from __future__ import annotations

import asyncio
import socket

from app.models.torrent import TorrentStatus
from app.providers.fake import FakeTorrentProvider


def test_fake_provider_never_touches_network(monkeypatch) -> None:
    def _blocked_dns(*args, **kwargs):
        raise AssertionError("DNS lookup attempted")

    def _blocked_connect(self, *args, **kwargs):
        raise AssertionError("socket.connect attempted")

    async def exercise() -> None:
        # патчи ставим внутри корутины: event loop уже создан asyncio.run
        monkeypatch.setattr(socket, "getaddrinfo", _blocked_dns)
        monkeypatch.setattr(socket.socket, "connect", _blocked_connect)
        provider = FakeTorrentProvider()
        torrent_id = await provider.add_magnet("magnet:?xt=urn:btih:" + "a" * 40, "Movies/")
        await provider.get_status(torrent_id)
        await provider.pause(torrent_id)
        assert await provider.get_status(torrent_id) is TorrentStatus.PAUSED
        await provider.resume(torrent_id)
        await provider.delete(torrent_id)

    asyncio.run(exercise())


def test_fake_provider_simulates_lifecycle() -> None:
    async def run() -> list[TorrentStatus]:
        provider = FakeTorrentProvider()
        torrent_id = await provider.add_magnet("magnet:?xt=urn:btih:" + "b" * 40, "Movies/")
        return [await provider.get_status(torrent_id) for _ in range(10)]

    seen = asyncio.run(run())
    assert seen[0] is TorrentStatus.QUEUED
    assert TorrentStatus.DOWNLOADING in seen
    assert seen[-1] is TorrentStatus.COMPLETED
