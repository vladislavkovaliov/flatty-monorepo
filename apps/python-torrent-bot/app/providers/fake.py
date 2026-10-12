from __future__ import annotations

import itertools
from dataclasses import dataclass

from app.models.torrent import TorrentInfo, TorrentStatus

_DOWNLOAD_SPEED = 1_048_576  # 1 MiB/s — фейковая скорость для UI
_UPLOAD_SPEED = 131_072  # 128 KiB/s
_ETA_SECONDS = 42
_PROGRESS_STEP = 12.5


@dataclass
class _FakeRecord:
    status: TorrentStatus = TorrentStatus.QUEUED
    ticks: int = 0
    progress: float = 0.0


class FakeTorrentProvider:
    """In-memory провайдер для TEST_MODE: без сети, без файловой системы, без qBittorrent."""

    def __init__(self) -> None:
        self._torrents: dict[str, _FakeRecord] = {}
        self._ids = itertools.count(1)

    async def add_magnet(self, magnet: str, save_path: str) -> str:
        torrent_id = f"fake-{next(self._ids)}"
        self._torrents[torrent_id] = _FakeRecord()
        return torrent_id

    async def get_info(self, torrent_id: str) -> TorrentInfo:
        record = self._torrents[torrent_id]
        if record.status in (TorrentStatus.QUEUED, TorrentStatus.DOWNLOADING):
            record.ticks += 1
            if record.status is TorrentStatus.QUEUED and record.ticks >= 2:
                record.status = TorrentStatus.DOWNLOADING
            elif record.status is TorrentStatus.DOWNLOADING and record.ticks >= 8:
                record.status = TorrentStatus.COMPLETED
            record.progress = min(record.ticks * _PROGRESS_STEP, 99.9)
        if record.status is TorrentStatus.COMPLETED:
            record.progress = 100.0
            return TorrentInfo(status=record.status, progress=100.0)
        if record.status is TorrentStatus.DOWNLOADING:
            return TorrentInfo(
                status=record.status,
                progress=record.progress,
                download_speed=_DOWNLOAD_SPEED,
                upload_speed=_UPLOAD_SPEED,
                eta_seconds=_ETA_SECONDS,
            )
        return TorrentInfo(status=record.status, progress=record.progress)

    async def pause(self, torrent_id: str) -> None:
        self._torrents[torrent_id].status = TorrentStatus.PAUSED

    async def resume(self, torrent_id: str) -> None:
        self._torrents[torrent_id].status = TorrentStatus.DOWNLOADING

    async def delete(self, torrent_id: str, delete_files: bool = False) -> None:
        self._torrents.pop(torrent_id, None)  # идемпотентно: cancel уже удалил
