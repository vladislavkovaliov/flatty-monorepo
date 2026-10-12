from __future__ import annotations

import os
from collections.abc import AsyncGenerator

import pytest
from dotenv import load_dotenv
from psycopg_pool import AsyncConnectionPool

from app.db.pool import apply_schema, create_pool
from app.models.media import MediaType, Quality
from app.models.torrent import TorrentStatus, TorrentTask
from app.repositories.postgres import PostgresRepository

load_dotenv()
DATABASE_URL = os.getenv("DATABASE_URL", "").strip()

# NB: module-level pytestmark НЕ использовать — он skip'нул бы и
# test_unreachable_database_fails_fast. Skip живёт в fixture pool:
# тесты, зависящие от pool/repo, skip'аются; fail-fast тест бегает всегда.

# тестовые telegram id (см. header файла): чистятся каскадом через users
_TEST_USER_LO = 990000001
_TEST_USER_HI = 990000099


def _task(user_id: int, info_hash: str) -> TorrentTask:
    return TorrentTask(
        id=0,
        user_id=user_id,
        magnet=f"magnet:?xt=urn:btih:{info_hash}",
        info_hash=info_hash,
        name="DB Test",
        media_type=MediaType.MOVIE,
        quality=Quality.FHD_1080,
    )


@pytest.fixture(scope="session")
async def pool() -> AsyncGenerator[AsyncConnectionPool, None]:
    if not DATABASE_URL:
        pytest.skip("DATABASE_URL is not set")
    try:
        p = await create_pool(DATABASE_URL)
    except RuntimeError as exc:
        pytest.skip(f"database unreachable: {exc}")  # exc уже с redacted URL
    try:
        await apply_schema(p)
        yield p
    finally:
        await p.close()


@pytest.fixture
async def repo(pool: AsyncConnectionPool) -> AsyncGenerator[PostgresRepository, None]:
    async with pool.connection() as conn:
        await conn.execute(
            "DELETE FROM users WHERE telegram_user_id BETWEEN %s AND %s",
            (_TEST_USER_LO, _TEST_USER_HI),
        )  # каскадом уберёт torrents/events/settings этих пользователей
    yield PostgresRepository(pool)
    async with pool.connection() as conn:
        await conn.execute(
            "DELETE FROM users WHERE telegram_user_id BETWEEN %s AND %s",
            (_TEST_USER_LO, _TEST_USER_HI),
        )


async def test_unreachable_database_fails_fast() -> None:
    # НЕ skip: порт 1 на localhost отказывает мгновенно; нужен только psycopg
    with pytest.raises(RuntimeError, match="Cannot reach PostgreSQL"):
        await create_pool("postgresql://nobody@127.0.0.1:1/postgres?sslmode=disable")


async def test_schema_applied_idempotent(pool: AsyncConnectionPool) -> None:
    await apply_schema(pool)  # повторное применение — не падает (CREATE IF NOT EXISTS)
    async with pool.connection() as conn:
        cur = await conn.execute("SELECT tablename FROM pg_tables WHERE schemaname = 'public'")
        rows = await cur.fetchall()
    names = {row[0] for row in rows}
    assert {"users", "torrents", "torrent_events", "settings"} <= names
    assert "media_items" not in names


async def test_task_crud_roundtrip(repo: PostgresRepository) -> None:
    user = _TEST_USER_LO
    inserted = await repo.insert_task(_task(user, "c" * 40))
    assert inserted is not None and inserted.id > 0

    fetched = await repo.get_task(user, inserted.id)
    assert fetched is not None
    assert fetched.user_id == user
    assert fetched.name == "DB Test"
    assert fetched.status is TorrentStatus.QUEUED
    assert fetched.media_type is MediaType.MOVIE
    assert fetched.quality is Quality.FHD_1080
    assert fetched.progress == 0.0
    assert fetched.created_at is not None

    fetched.status = TorrentStatus.DOWNLOADING
    fetched.progress = 42.5
    fetched.completed_at = None
    await repo.update_task(fetched)

    again = await repo.get_task(user, inserted.id)
    assert again is not None
    assert again.status is TorrentStatus.DOWNLOADING
    assert again.progress == 42.5

    await repo.delete_task(inserted.id)
    assert await repo.get_task(user, inserted.id) is None


async def test_duplicate_user_hash_rejected(repo: PostgresRepository) -> None:
    first = await repo.insert_task(_task(_TEST_USER_LO + 1, "d" * 40))
    assert first is not None

    duplicate = await repo.insert_task(_task(_TEST_USER_LO + 1, "d" * 40))
    assert duplicate is None  # UNIQUE (user_id, info_hash)

    other_user = await repo.insert_task(_task(_TEST_USER_HI, "d" * 40))
    assert other_user is not None  # другой пользователь — тот же hash разрешён


async def test_event_inserted(repo: PostgresRepository, pool: AsyncConnectionPool) -> None:
    task = await repo.insert_task(_task(_TEST_USER_LO + 2, "e" * 40))
    assert task is not None
    await repo.add_event(task.id, "status_changed", '{"from": "queued", "to": "paused"}')
    async with pool.connection() as conn:
        cur = await conn.execute(
            "SELECT event, payload FROM torrent_events WHERE torrent_id = %s", (task.id,)
        )
        rows = await cur.fetchall()
    assert rows == [("status_changed", '{"from": "queued", "to": "paused"}')]


async def test_ownership_filter(repo: PostgresRepository) -> None:
    task = await repo.insert_task(_task(_TEST_USER_LO + 3, "f" * 40))
    assert task is not None
    assert await repo.get_task(_TEST_USER_HI, task.id) is None  # чужой = «не найдено»
    assert await repo.list_tasks(_TEST_USER_HI) == []
    assert len(await repo.list_tasks(_TEST_USER_LO + 3)) == 1


async def test_settings_roundtrip(repo: PostgresRepository) -> None:
    user = _TEST_USER_LO + 4
    assert await repo.get_setting(user, "notify") is None
    await repo.set_setting(user, "notify", "on")
    assert await repo.get_setting(user, "notify") == "on"
    await repo.set_setting(user, "notify", "off")
    assert await repo.get_setting(user, "notify") == "off"
