from __future__ import annotations

import itertools
from dataclasses import dataclass

from app.models.torrent import TorrentStatus


@dataclass
class _FakeRecord:
    status: TorrentStatus = TorrentStatus.QUEUED
    ticks: int = 0


class FakeTorrentProvider:
    """In-memory провайдер для TEST_MODE: без сети, без файловой системы, без qBittorrent."""

    def __init__(self) -> None:
        self._torrents: dict[str, _FakeRecord] = {}
        self._ids = itertools.count(1)

    async def add_magnet(self, magnet: str, save_path: str) -> str:
        torrent_id = f"fake-{next(self._ids)}"
        self._torrents[torrent_id] = _FakeRecord()
        return torrent_id

    async def get_status(self, torrent_id: str) -> TorrentStatus:
        record = self._torrents[torrent_id]
        if record.status not in (TorrentStatus.QUEUED, TorrentStatus.DOWNLOADING):
            return record.status
        record.ticks += 1
        if record.status is TorrentStatus.QUEUED and record.ticks >= 2:
            record.status = TorrentStatus.DOWNLOADING
        elif record.status is TorrentStatus.DOWNLOADING and record.ticks >= 8:
            record.status = TorrentStatus.COMPLETED
        return record.status

    async def pause(self, torrent_id: str) -> None:
        self._torrents[torrent_id].status = TorrentStatus.PAUSED

    async def resume(self, torrent_id: str) -> None:
        self._torrents[torrent_id].status = TorrentStatus.DOWNLOADING

    async def delete(self, torrent_id: str, delete_files: bool = False) -> None:
        del self._torrents[torrent_id]
