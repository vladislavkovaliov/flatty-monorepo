from __future__ import annotations

from types import SimpleNamespace
from unittest.mock import MagicMock

import pytest
from qbittorrentapi.exceptions import NotFound404Error

from app.models.torrent import TorrentStatus
from app.providers.qbittorrent import QBittorrentProvider

HASH_A = "a" * 40


@pytest.fixture
def client() -> MagicMock:
    return MagicMock()


@pytest.fixture
def provider(client: MagicMock, monkeypatch: pytest.MonkeyPatch) -> QBittorrentProvider:
    monkeypatch.setattr("app.providers.qbittorrent.Client", lambda **_kwargs: client)
    return QBittorrentProvider(host="localhost", port=8080, username="u", password="p")


async def test_add_magnet_returns_hash(provider: QBittorrentProvider, client: MagicMock) -> None:
    client.torrents_info.return_value = [SimpleNamespace(info_hash=HASH_A.upper())]
    magnet = f"magnet:?xt=urn:btih:{HASH_A}&dn=Test"
    result = await provider.add_magnet(magnet, "Movies/")
    assert result == HASH_A
    client.torrents_add.assert_called_once_with(urls=magnet, save_path="Movies/")
    client.torrents_info.assert_called_with(torrent_hashes=HASH_A)


async def test_get_info_maps_fields(provider: QBittorrentProvider, client: MagicMock) -> None:
    client.torrents_info.return_value = [
        SimpleNamespace(
            state="downloading",
            progress=0.42,
            dlspeed=1_000_000,
            upspeed=50_000,
            eta=3600,
        )
    ]
    info = await provider.get_info(HASH_A)
    assert info.status is TorrentStatus.DOWNLOADING
    assert info.progress == 42.0
    assert info.download_speed == 1_000_000
    assert info.upload_speed == 50_000
    assert info.eta_seconds == 3600


async def test_get_info_negative_eta_none(provider: QBittorrentProvider, client: MagicMock) -> None:
    client.torrents_info.return_value = [
        SimpleNamespace(state="uploading", progress=1.0, dlspeed=0, upspeed=100, eta=-1)
    ]
    info = await provider.get_info(HASH_A)
    assert info.status is TorrentStatus.COMPLETED
    assert info.eta_seconds is None


async def test_get_info_missing_raises(provider: QBittorrentProvider, client: MagicMock) -> None:
    client.torrents_info.return_value = []
    with pytest.raises(NotFound404Error):
        await provider.get_info("missing")


async def test_pause_resume(provider: QBittorrentProvider, client: MagicMock) -> None:
    await provider.pause(HASH_A)
    client.torrents_pause.assert_called_once_with(torrents_hashes=HASH_A)
    await provider.resume(HASH_A)
    client.torrents_resume.assert_called_once_with(torrents_hashes=HASH_A)


async def test_delete_idempotent(provider: QBittorrentProvider, client: MagicMock) -> None:
    client.torrents_delete.side_effect = NotFound404Error()
    await provider.delete(HASH_A, delete_files=False)  # не должно бросить
    client.torrents_delete.assert_called_once_with(delete_files=False, torrents_hashes=HASH_A)
