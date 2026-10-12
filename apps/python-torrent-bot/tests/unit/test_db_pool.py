from __future__ import annotations

from app.db.pool import redact_database_url


def test_redact_password() -> None:
    url = "postgresql://postgres:s3cret@192.168.1.35:5432/torrent_bot?sslmode=disable"
    redacted = redact_database_url(url)
    assert "s3cret" not in redacted
    assert redacted == "postgresql://postgres:***@192.168.1.35:5432/torrent_bot?sslmode=disable"


def test_redact_without_password_is_noop() -> None:
    assert (
        redact_database_url("postgresql://localhost/torrent_bot")
        == "postgresql://localhost/torrent_bot"
    )
