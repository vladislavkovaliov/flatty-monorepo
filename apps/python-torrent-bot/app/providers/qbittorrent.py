from __future__ import annotations

import asyncio
import contextlib
import logging

from qbittorrentapi import Client
from qbittorrentapi.exceptions import NotFound404Error

from app.models.torrent import TorrentInfo, TorrentStatus

logger = logging.getLogger(__name__)

_DOWNLOADING = {
    "metaDL",
    "downloading",
    "forcedDL",
    "stalledDL",
    "checkingDL",
    "moving",
    "allocating",
}
_QUEUED = {"queuedDL"}
_PAUSED = {"pausedDL", "stoppedDL"}
_COMPLETED = {"pausedUP", "stoppedUP", "uploading", "stalledUP", "forcedUP", "checkingUP"}
_FAILED = {"error", "missingFiles", "unknown"}


def _extract_info_hash(magnet: str) -> str:
    marker = "urn:btih:"
    idx = magnet.find(marker)
    if idx == -1:
        return ""
    rest = magnet[idx + len(marker) :]
    for sep in ("&",):
        rest = rest.split(sep, 1)[0]
    return rest.lower()


def _map_state(state: str, progress: float) -> TorrentStatus:
    if state in _FAILED:
        return TorrentStatus.FAILED
    if state in _COMPLETED or progress >= 100.0:
        return TorrentStatus.COMPLETED
    if state in _PAUSED:
        return TorrentStatus.PAUSED
    if state in _QUEUED:
        return TorrentStatus.QUEUED
    if state in _DOWNLOADING:
        return TorrentStatus.DOWNLOADING
    return TorrentStatus.DOWNLOADING


class QBittorrentProvider:
    """Реальный провайдер: синхронный qbittorrent-api через asyncio.to_thread."""

    def __init__(self, host: str, port: int, username: str, password: str) -> None:
        self._client = Client(host=host, port=port, username=username, password=password)

    async def add_magnet(self, magnet: str, save_path: str) -> str:
        info_hash = _extract_info_hash(magnet)

        def _add() -> None:
            self._client.torrents_add(urls=magnet, save_path=save_path)

        await asyncio.to_thread(_add)
        await self._wait_for_torrent(info_hash)
        logger.info("qbittorrent add_magnet info_hash=%s", info_hash)
        return info_hash

    async def _wait_for_torrent(self, info_hash: str) -> None:
        def _exists() -> bool:
            return bool(self._client.torrents_info(torrent_hashes=info_hash))

        for _ in range(20):  # ponytail: до 10с на появление торрента
            if await asyncio.to_thread(_exists):
                return
            await asyncio.sleep(0.5)
        logger.warning("qbittorrent torrent not visible yet info_hash=%s", info_hash)

    async def get_info(self, torrent_id: str) -> TorrentInfo:
        def _get() -> tuple[str, float, int, int, int] | None:
            try:
                t = self._client.torrents_info(torrent_hashes=torrent_id)[0]
            except IndexError:
                return None
            return str(t.state), float(t.progress), int(t.eta), int(t.dlspeed), int(t.upspeed)

        raw = await asyncio.to_thread(_get)
        if raw is None:
            raise NotFound404Error
        state, progress_raw, eta, dlspeed, upspeed = raw
        progress = progress_raw * 100.0
        return TorrentInfo(
            status=_map_state(state, progress),
            progress=progress,
            download_speed=dlspeed,
            upload_speed=upspeed,
            eta_seconds=None if eta < 0 else eta,
        )

    async def pause(self, torrent_id: str) -> None:
        await asyncio.to_thread(self._client.torrents_pause, torrents_hashes=torrent_id)

    async def resume(self, torrent_id: str) -> None:
        await asyncio.to_thread(self._client.torrents_resume, torrents_hashes=torrent_id)

    async def delete(self, torrent_id: str, delete_files: bool = False) -> None:
        def _delete() -> None:
            with contextlib.suppress(NotFound404Error):  # уже удалён — идемпотентно
                self._client.torrents_delete(delete_files=delete_files, torrents_hashes=torrent_id)

        await asyncio.to_thread(_delete)
        logger.info("qbittorrent delete info_hash=%s delete_files=%s", torrent_id, delete_files)
