from __future__ import annotations

import os
from dataclasses import dataclass

from dotenv import load_dotenv


@dataclass(frozen=True)
class Config:
    bot_token: str
    test_mode: bool
    database_url: str
    log_level: str
    api_host: str
    api_port: int
    api_user_id: int


def load_config() -> Config:
    load_dotenv()
    token = os.getenv("BOT_TOKEN")
    if not token:
        raise RuntimeError("BOT_TOKEN is not set. Copy .env.example to .env and fill BOT_TOKEN.")
    test_mode = os.getenv("TEST_MODE", "true").strip().lower() in {"1", "true", "yes"}
    if not test_mode:
        raise RuntimeError(
            "Phase 0 requires TEST_MODE=true: real providers (qBittorrent/NAS) "
            "are not implemented yet."
        )
    database_url = os.getenv("DATABASE_URL", "").strip()
    if not database_url:
        raise RuntimeError(
            "DATABASE_URL is not set. Copy .env.example to .env and fill DATABASE_URL "
            "(format: postgresql://user:pass@host:5432/dbname?sslmode=disable)."
        )
    if not database_url.startswith(("postgresql://", "postgres://")):
        raise RuntimeError(
            "DATABASE_URL must start with postgresql:// — other URL schemes are not supported."
        )
    log_level = (os.getenv("LOG_LEVEL") or "INFO").strip().upper()
    if log_level not in {"DEBUG", "INFO", "WARNING", "ERROR", "CRITICAL"}:
        raise RuntimeError(
            f"LOG_LEVEL must be one of DEBUG/INFO/WARNING/ERROR/CRITICAL (got {log_level!r})."
        )
    api_host = (os.getenv("API_HOST") or "127.0.0.1").strip()
    if api_host not in {"127.0.0.1", "localhost", "::1"}:
        raise RuntimeError(
            f"API_HOST must be a localhost address — the HTTP API has no auth (got {api_host!r})."
        )
    api_port_raw = (os.getenv("API_PORT") or "8080").strip()
    try:
        api_port = int(api_port_raw)
    except ValueError:
        raise RuntimeError(f"API_PORT must be an integer (got {api_port_raw!r}).") from None
    if not 1 <= api_port <= 65535:
        raise RuntimeError(f"API_PORT must be between 1 and 65535 (got {api_port}).")
    api_user_id_raw = (os.getenv("API_USER_ID") or "1").strip()
    try:
        api_user_id = int(api_user_id_raw)
    except ValueError:
        raise RuntimeError(f"API_USER_ID must be an integer (got {api_user_id_raw!r}).") from None
    if api_user_id <= 0:
        raise RuntimeError(f"API_USER_ID must be a positive Telegram user id (got {api_user_id}).")
    return Config(
        bot_token=token,
        test_mode=test_mode,
        database_url=database_url,
        log_level=log_level,
        api_host=api_host,
        api_port=api_port,
        api_user_id=api_user_id,
    )
