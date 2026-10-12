from __future__ import annotations

from app.models.torrent import TorrentStatus
from app.providers.qbittorrent import _extract_info_hash, _map_state


def test_map_state_downloading() -> None:
    states = ("metaDL", "downloading", "forcedDL", "stalledDL", "checkingDL", "moving")
    for state in states:
        assert _map_state(state, 10.0) is TorrentStatus.DOWNLOADING


def test_map_state_queued() -> None:
    assert _map_state("queuedDL", 0.0) is TorrentStatus.QUEUED


def test_map_state_paused() -> None:
    assert _map_state("pausedDL", 50.0) is TorrentStatus.PAUSED
    assert _map_state("stoppedDL", 50.0) is TorrentStatus.PAUSED


def test_map_state_completed() -> None:
    for state in ("pausedUP", "stoppedUP", "uploading", "stalledUP", "forcedUP", "checkingUP"):
        assert _map_state(state, 100.0) is TorrentStatus.COMPLETED


def test_map_state_failed() -> None:
    for state in ("error", "missingFiles", "unknown"):
        assert _map_state(state, 0.0) is TorrentStatus.FAILED


def test_map_state_progress_overrides() -> None:
    assert _map_state("downloading", 100.0) is TorrentStatus.COMPLETED


def test_extract_info_hash() -> None:
    magnet = "magnet:?xt=urn:btih:ABCDEF0123456789ABCDEF0123456789ABCDEF01&dn=Name"
    assert _extract_info_hash(magnet) == "abcdef0123456789abcdef0123456789abcdef01"


def test_extract_info_hash_without_btih() -> None:
    assert _extract_info_hash("magnet:?dn=Name") == ""
