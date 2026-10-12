from __future__ import annotations

from dataclasses import dataclass
from urllib.parse import parse_qs, urlparse


class MagnetParseError(ValueError):
    """Magnet-ссылка не прошла валидацию (FEAT-002 Error Handling)."""


@dataclass(frozen=True)
class MagnetInfo:
    raw: str
    info_hash: str
    display_name: str | None
    trackers: list[str]
    size_bytes: int | None


def parse_magnet(raw: str) -> MagnetInfo:
    parsed = urlparse(raw.strip())
    if parsed.scheme != "magnet":
        raise MagnetParseError(f"unsupported scheme: {parsed.scheme!r}")

    params = parse_qs(parsed.query)

    info_hash = ""
    for xt in params.get("xt", []):
        if xt.startswith("urn:btih:"):
            info_hash = xt.removeprefix("urn:btih:").strip().lower()
            break
    if not info_hash:
        raise MagnetParseError("magnet has no info hash (xt=urn:btih:...)")

    display_name = params["dn"][0] if "dn" in params else None
    trackers = list(params.get("tr", []))

    size_bytes: int | None = None
    xl_values = params.get("xl", [])
    if xl_values:
        try:
            size_bytes = int(xl_values[0])
        except ValueError:
            size_bytes = None

    return MagnetInfo(
        raw=raw,
        info_hash=info_hash,
        display_name=display_name,
        trackers=trackers,
        size_bytes=size_bytes,
    )
