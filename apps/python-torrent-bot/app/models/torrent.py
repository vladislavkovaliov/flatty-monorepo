from __future__ import annotations

from dataclasses import dataclass, field
from datetime import UTC, datetime
from enum import StrEnum

from app.models.media import MediaType, Quality


class TorrentStatus(StrEnum):
    QUEUED = "queued"
    DOWNLOADING = "downloading"
    PAUSED = "paused"
    COMPLETED = "completed"
    FAILED = "failed"
    CANCELLED = "cancelled"


@dataclass
class TorrentTask:
    id: int
    user_id: int
    magnet: str
    info_hash: str
    name: str
    media_type: MediaType
    quality: Quality
    status: TorrentStatus = TorrentStatus.QUEUED
    progress: float = 0.0
    size_bytes: int | None = None
    download_speed: int = 0
    upload_speed: int = 0
    eta_seconds: int | None = None
    created_at: datetime = field(default_factory=lambda: datetime.now(UTC))
    completed_at: datetime | None = None
    # ponytail: provider_id сверх минимума ARCHITECTURE §4 — нужен для FEAT-011 (qBittorrent)
    provider_id: str = ""


def format_size(size_bytes: int | None) -> str:
    """Человекочитаемый размер для UI (FEAT-003/007). None -> «не определён»."""
    if size_bytes is None:
        return "не определён"
    value = float(size_bytes)
    for unit in ("B", "KiB", "MiB", "GiB", "TiB"):
        if value < 1024 or unit == "TiB":
            if unit == "B":
                return f"{int(value)} B"
            return f"{value:.1f} {unit}"
        value /= 1024
    return f"{value:.1f} TiB"
