from __future__ import annotations

import os
from dataclasses import dataclass

from dotenv import load_dotenv


@dataclass(frozen=True)
class Config:
    log_level: str
    api_host: str
    api_port: int
    database_url: str
    ollama: str
    ollama_base_url: str


def load_config() -> Config:
    load_dotenv()

    ollama = os.getenv("OLLAMA_MODEL", "").strip()
    
    if not ollama:
        """ TODO: update .env and env.example by OLLAMA_MODEL """
        raise RuntimeError("ollama str is not set. Copy .env.example to .env "
        "and fill OLLAMA_MODEL.")
    
    ollama_base_url = os.getenv("OLLAMA_BASE_URL", "").strip()

    if not ollama_base_url:
        """ TODO: update .env and env.example by OLLAMA_BASE_URL """
        raise RuntimeError("ollama_base_url str is not set. Copy .env.example to .env "
        "and fill OLLAMA_BASE_URL.")
    

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
    
    return Config(
        database_url=database_url,
        log_level=log_level,
        api_host=api_host,
        api_port=api_port,
        ollama=ollama,
        ollama_base_url=ollama_base_url,
    )
