from __future__ import annotations

import pytest

from app.config import load_config


def test_missing_database_url_raises(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setattr("app.config.load_dotenv", lambda: None)
    monkeypatch.setenv("BOT_TOKEN", "123:TEST")
    monkeypatch.setenv("TEST_MODE", "true")
    monkeypatch.delenv("DATABASE_URL", raising=False)
    with pytest.raises(RuntimeError, match="DATABASE_URL is not set"):
        load_config()


def test_invalid_scheme_raises(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setattr("app.config.load_dotenv", lambda: None)
    monkeypatch.setenv("BOT_TOKEN", "123:TEST")
    monkeypatch.setenv("TEST_MODE", "true")
    monkeypatch.setenv("DATABASE_URL", "mysql://user:pass@host/db")
    with pytest.raises(RuntimeError, match="postgresql://"):
        load_config()


def test_valid_config_carries_database_url(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setattr("app.config.load_dotenv", lambda: None)
    monkeypatch.setenv("BOT_TOKEN", "123:TEST")
    monkeypatch.setenv("TEST_MODE", "true")
    url = "postgresql://user:pass@192.168.1.35:5432/torrent_bot?sslmode=disable"
    monkeypatch.setenv("DATABASE_URL", url)
    config = load_config()
    assert config.database_url == url
    assert config.test_mode is True


def test_log_level_normalized(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setattr("app.config.load_dotenv", lambda: None)
    monkeypatch.setenv("BOT_TOKEN", "123:TEST")
    monkeypatch.setenv("TEST_MODE", "true")
    monkeypatch.setenv("DATABASE_URL", "postgresql://user:pass@host:5432/db")
    monkeypatch.setenv("LOG_LEVEL", "debug")
    assert load_config().log_level == "DEBUG"


def test_invalid_log_level_raises(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setattr("app.config.load_dotenv", lambda: None)
    monkeypatch.setenv("BOT_TOKEN", "123:TEST")
    monkeypatch.setenv("TEST_MODE", "true")
    monkeypatch.setenv("DATABASE_URL", "postgresql://user:pass@host:5432/db")
    monkeypatch.setenv("LOG_LEVEL", "NOPE")
    with pytest.raises(RuntimeError, match="LOG_LEVEL"):
        load_config()


def test_api_defaults(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setattr("app.config.load_dotenv", lambda: None)
    monkeypatch.setenv("BOT_TOKEN", "123:TEST")
    monkeypatch.setenv("TEST_MODE", "true")
    monkeypatch.setenv("DATABASE_URL", "postgresql://user:pass@host:5432/db")
    monkeypatch.delenv("API_HOST", raising=False)
    monkeypatch.delenv("API_PORT", raising=False)
    monkeypatch.delenv("API_USER_ID", raising=False)
    config = load_config()
    assert config.api_host == "127.0.0.1"
    assert config.api_port == 8080
    assert config.api_user_id == 1


def test_api_explicit_values(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setattr("app.config.load_dotenv", lambda: None)
    monkeypatch.setenv("BOT_TOKEN", "123:TEST")
    monkeypatch.setenv("TEST_MODE", "true")
    monkeypatch.setenv("DATABASE_URL", "postgresql://user:pass@host:5432/db")
    monkeypatch.setenv("API_HOST", "localhost")
    monkeypatch.setenv("API_PORT", "9090")
    monkeypatch.setenv("API_USER_ID", "42")
    config = load_config()
    assert (config.api_host, config.api_port, config.api_user_id) == ("localhost", 9090, 42)


def test_invalid_api_port_raises(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setattr("app.config.load_dotenv", lambda: None)
    monkeypatch.setenv("BOT_TOKEN", "123:TEST")
    monkeypatch.setenv("TEST_MODE", "true")
    monkeypatch.setenv("DATABASE_URL", "postgresql://user:pass@host:5432/db")
    monkeypatch.setenv("API_PORT", "eighty")
    with pytest.raises(RuntimeError, match="API_PORT"):
        load_config()


def test_non_localhost_api_host_raises(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setattr("app.config.load_dotenv", lambda: None)
    monkeypatch.setenv("BOT_TOKEN", "123:TEST")
    monkeypatch.setenv("TEST_MODE", "true")
    monkeypatch.setenv("DATABASE_URL", "postgresql://user:pass@host:5432/db")
    monkeypatch.setenv("API_HOST", "0.0.0.0")
    with pytest.raises(RuntimeError, match="API_HOST"):
        load_config()


def _base_env(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setattr("app.config.load_dotenv", lambda: None)
    monkeypatch.setenv("BOT_TOKEN", "123:TEST")
    monkeypatch.setenv("DATABASE_URL", "postgresql://user:pass@host:5432/db")
    monkeypatch.delenv("QBITTORRENT_URL", raising=False)
    monkeypatch.delenv("QBITTORRENT_USERNAME", raising=False)
    monkeypatch.delenv("QBITTORRENT_PASSWORD", raising=False)
    monkeypatch.delenv("SAVE_PATH_MOVIES", raising=False)
    monkeypatch.delenv("SAVE_PATH_TV", raising=False)


def test_test_mode_false_without_url_raises(monkeypatch: pytest.MonkeyPatch) -> None:
    _base_env(monkeypatch)
    monkeypatch.setenv("TEST_MODE", "false")
    with pytest.raises(RuntimeError, match="QBITTORRENT_URL"):
        load_config()


def test_test_mode_false_with_url_is_valid(monkeypatch: pytest.MonkeyPatch) -> None:
    _base_env(monkeypatch)
    monkeypatch.setenv("TEST_MODE", "false")
    monkeypatch.setenv("QBITTORRENT_URL", "http://192.168.1.50:8080")
    config = load_config()
    assert config.test_mode is False
    assert config.qbittorrent_url == "http://192.168.1.50:8080"


def test_repr_masks_password(monkeypatch: pytest.MonkeyPatch) -> None:
    _base_env(monkeypatch)
    monkeypatch.setenv("TEST_MODE", "true")
    monkeypatch.setenv("QBITTORRENT_PASSWORD", "hunter2secret")
    config = load_config()
    text = repr(config)
    assert "hunter2secret" not in text
    assert "***" in text


def test_save_paths_defaults(monkeypatch: pytest.MonkeyPatch) -> None:
    _base_env(monkeypatch)
    monkeypatch.setenv("TEST_MODE", "true")
    config = load_config()
    assert config.save_path_movies == "Movies/"
    assert config.save_path_tv == "TV/"


def test_save_paths_env_override(monkeypatch: pytest.MonkeyPatch) -> None:
    _base_env(monkeypatch)
    monkeypatch.setenv("TEST_MODE", "true")
    monkeypatch.setenv("SAVE_PATH_MOVIES", "/mnt/media/movies")
    monkeypatch.setenv("SAVE_PATH_TV", "/mnt/media/tv")
    config = load_config()
    assert config.save_path_movies == "/mnt/media/movies"
    assert config.save_path_tv == "/mnt/media/tv"
