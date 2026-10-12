from __future__ import annotations

from typing import Protocol

from app.models.torrent import TorrentInfo


class TorrentProvider(Protocol):
    """Контракт провайдера (docs/API.md §2). Handlers никогда не работают с ним напрямую."""

    async def add_magnet(self, magnet: str, save_path: str) -> str: ...

    async def get_info(self, torrent_id: str) -> TorrentInfo: ...

    async def pause(self, torrent_id: str) -> None: ...

    async def resume(self, torrent_id: str) -> None: ...

    async def delete(self, torrent_id: str, delete_files: bool = False) -> None: ...
