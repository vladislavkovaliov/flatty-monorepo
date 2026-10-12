from __future__ import annotations

import pytest

from app.services.magnet_parser import MagnetParseError, parse_magnet

VALID = (
    "magnet:?xt=urn:btih:a1b2c3d4e5f60718293a4b5c6d7e8f9012345678"
    "&dn=The.Matrix.1999.1080p.BluRay"
    "&tr=udp%3A%2F%2Ftracker.example.org%3A1337%2Fannounce"
    "&xl=12884901888"
    "&xs=http%3A%2F%2Funrelated.example"
)


def test_valid_magnet() -> None:
    info = parse_magnet(VALID)
    assert info.raw == VALID
    assert info.info_hash == "a1b2c3d4e5f60718293a4b5c6d7e8f9012345678"
    assert info.display_name == "The.Matrix.1999.1080p.BluRay"
    assert info.trackers == ["udp://tracker.example.org:1337/announce"]
    assert info.size_bytes == 12884901888


def test_missing_dn() -> None:
    info = parse_magnet("magnet:?xt=urn:btih:" + "a" * 40)
    assert info.display_name is None
    assert info.trackers == []
    assert info.size_bytes is None


def test_multiple_trackers() -> None:
    raw = (
        "magnet:?xt=urn:btih:"
        + "b" * 40
        + "&tr=udp%3A%2F%2Fone%2Fannounce"
        + "&tr=udp%3A%2F%2Ftwo%2Fannounce"
    )
    info = parse_magnet(raw)
    assert info.trackers == ["udp://one/announce", "udp://two/announce"]


def test_url_encoded_dn() -> None:
    raw = "magnet:?xt=urn:btih:" + "c" * 40 + "&dn=The%20Matrix%20%281999%29"
    assert parse_magnet(raw).display_name == "The Matrix (1999)"


def test_missing_info_hash() -> None:
    with pytest.raises(MagnetParseError):
        parse_magnet("magnet:?dn=NoHashHere")


def test_invalid_scheme() -> None:
    with pytest.raises(MagnetParseError):
        parse_magnet("https://example.com/file.torrent")
    with pytest.raises(MagnetParseError):
        parse_magnet("just some plain text")


def test_xl_parameter() -> None:
    raw = "magnet:?xt=urn:btih:" + "d" * 40 + "&xl=987654321"
    assert parse_magnet(raw).size_bytes == 987654321
