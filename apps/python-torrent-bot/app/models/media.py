from __future__ import annotations

from enum import StrEnum


class MediaType(StrEnum):
    MOVIE = "movie"
    SERIES = "series"


class Quality(StrEnum):
    UHD_2160 = "2160p"
    FHD_1080 = "1080p"
    HD_720 = "720p"
    UNKNOWN = "unknown"


SAVE_PATHS: dict[MediaType, str] = {
    MediaType.MOVIE: "Movies/",
    MediaType.SERIES: "TV/",
}

MEDIA_LABELS: dict[MediaType, str] = {
    MediaType.MOVIE: "Фильм",
    MediaType.SERIES: "Сериал",
}

MEDIA_EMOJI: dict[MediaType, str] = {
    MediaType.MOVIE: "🎬",
    MediaType.SERIES: "📺",
}

QUALITY_LABELS: dict[Quality, str] = {
    Quality.UHD_2160: "2160p / 4K",
    Quality.FHD_1080: "1080p",
    Quality.HD_720: "720p",
    Quality.UNKNOWN: "Не определять",
}
