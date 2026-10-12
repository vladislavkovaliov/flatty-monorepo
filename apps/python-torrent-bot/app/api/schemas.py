from __future__ import annotations

from datetime import datetime

from pydantic import BaseModel, ConfigDict

from app.models.media import MediaType, Quality
from app.models.torrent import TorrentStatus


class TaskCreateIn(BaseModel):
    magnet: str
    media_type: MediaType
    quality: Quality


class TorrentTaskOut(BaseModel):
    model_config = ConfigDict(from_attributes=True)

    id: int
    user_id: int
    magnet: str
    info_hash: str
    name: str
    media_type: MediaType
    quality: Quality
    status: TorrentStatus
    progress: float
    size_bytes: int | None
    download_speed: int
    upload_speed: int
    eta_seconds: int | None
    created_at: datetime
    completed_at: datetime | None


class ErrorOut(BaseModel):
    detail: str
