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
        await provider.get_info(torrent_id)
        await provider.pause(torrent_id)
        assert (await provider.get_info(torrent_id)).status is TorrentStatus.PAUSED
        await provider.resume(torrent_id)
        await provider.delete(torrent_id)

    asyncio.run(exercise())


def test_fake_provider_simulates_lifecycle() -> None:
    async def run() -> list[TorrentStatus]:
        provider = FakeTorrentProvider()
        torrent_id = await provider.add_magnet("magnet:?xt=urn:btih:" + "b" * 40, "Movies/")
        return [(await provider.get_info(torrent_id)).status for _ in range(10)]

    seen = asyncio.run(run())
    assert seen[0] is TorrentStatus.QUEUED
    assert TorrentStatus.DOWNLOADING in seen
    assert seen[-1] is TorrentStatus.COMPLETED


def test_fake_provider_progress_reaches_100() -> None:
    async def run() -> list[float]:
        provider = FakeTorrentProvider()
        torrent_id = await provider.add_magnet("magnet:?xt=urn:btih:" + "c" * 40, "Movies/")
        return [(await provider.get_info(torrent_id)).progress for _ in range(10)]

    progresses = asyncio.run(run())
    assert progresses[-1] == 100.0
    active = progresses[:8]
    assert all(b >= a for a, b in zip(active, active[1:], strict=False))


def test_fake_provider_pause_freezes_progress() -> None:
    async def run() -> float:
        provider = FakeTorrentProvider()
        torrent_id = await provider.add_magnet("magnet:?xt=urn:btih:" + "d" * 40, "Movies/")
        await provider.get_info(torrent_id)
        await provider.get_info(torrent_id)
        await provider.pause(torrent_id)
        frozen = (await provider.get_info(torrent_id)).progress
        await provider.get_info(torrent_id)
        return frozen

    frozen = asyncio.run(run())
    assert 0.0 < frozen < 100.0


def test_fake_provider_download_speed_only_when_downloading() -> None:
    async def run():
        provider = FakeTorrentProvider()
        torrent_id = await provider.add_magnet("magnet:?xt=urn:btih:" + "e" * 40, "Movies/")
        infos = [await provider.get_info(torrent_id) for _ in range(10)]
        return infos

    infos = asyncio.run(run())
    downloading = [i for i in infos if i.status is TorrentStatus.DOWNLOADING]
    assert downloading
    assert all(i.download_speed > 0 and i.eta_seconds is not None for i in downloading)
    assert infos[0].download_speed == 0
    assert infos[-1].status is TorrentStatus.COMPLETED
    assert infos[-1].download_speed == 0
