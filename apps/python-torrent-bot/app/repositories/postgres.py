from __future__ import annotations

from collections.abc import Mapping
from typing import Any

from psycopg import AsyncCursor
from psycopg.rows import dict_row
from psycopg_pool import AsyncConnectionPool

from app.models.media import MediaType, Quality
from app.models.torrent import TorrentStatus, TorrentTask

_TASK_COLUMNS = """
    t.id, t.magnet, t.info_hash, t.name, t.media_type, t.quality, t.status,
    t.progress, t.size_bytes, t.download_speed, t.upload_speed, t.eta_seconds,
    t.created_at, t.completed_at, u.telegram_user_id
"""


def _task_from_row(row: Mapping[str, Any]) -> TorrentTask:
    return TorrentTask(
        id=row["id"],
        user_id=row["telegram_user_id"],
        magnet=row["magnet"],
        info_hash=row["info_hash"],
        name=row["name"],
        media_type=MediaType(row["media_type"]),
        quality=Quality(row["quality"]),
        status=TorrentStatus(row["status"]),
        progress=float(row["progress"]),
        size_bytes=row["size_bytes"],
        download_speed=int(row["download_speed"]),
        upload_speed=int(row["upload_speed"]),
        eta_seconds=row["eta_seconds"],
        created_at=row["created_at"],
        completed_at=row["completed_at"],
        # колонки нет по design (§Components) — session state заполняет TorrentService
        provider_id="",
    )


class PostgresRepository:
    def __init__(self, pool: AsyncConnectionPool) -> None:
        self._pool = pool

    @staticmethod
    async def _ensure_user(cur: AsyncCursor[dict[str, Any]], telegram_user_id: int) -> int:
        await cur.execute(
            """
            INSERT INTO users (telegram_user_id) VALUES (%s)
            ON CONFLICT (telegram_user_id)
            DO UPDATE SET username = COALESCE(users.username, EXCLUDED.username)
            RETURNING id
            """,
            (telegram_user_id,),
        )
        row = await cur.fetchone()
        assert row is not None
        return int(row["id"])

    async def find_by_hash(self, user_id: int, info_hash: str) -> TorrentTask | None:
        async with self._pool.connection() as conn, conn.cursor(row_factory=dict_row) as cur:
            await cur.execute(
                f"SELECT {_TASK_COLUMNS} FROM torrents t "
                "JOIN users u ON u.id = t.user_id "
                "WHERE u.telegram_user_id = %s AND t.info_hash = %s",
                (user_id, info_hash),
            )
            row = await cur.fetchone()
        return _task_from_row(row) if row is not None else None

    async def insert_task(self, task: TorrentTask) -> TorrentTask | None:
        async with self._pool.connection() as conn, conn.cursor(row_factory=dict_row) as cur:
            user_pk = await self._ensure_user(cur, task.user_id)
            await cur.execute(
                """
                    INSERT INTO torrents (
                        user_id, magnet, info_hash, name, media_type, quality,
                        status, progress, size_bytes, download_speed, upload_speed,
                        eta_seconds, created_at, completed_at
                    ) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
                    ON CONFLICT (user_id, info_hash) DO NOTHING
                    RETURNING id
                    """,
                (
                    user_pk,
                    task.magnet,
                    task.info_hash,
                    task.name,
                    str(task.media_type),
                    str(task.quality),
                    str(task.status),
                    task.progress,
                    task.size_bytes,
                    task.download_speed,
                    task.upload_speed,
                    task.eta_seconds,
                    task.created_at,
                    task.completed_at,
                ),
            )
            inserted = await cur.fetchone()
            if inserted is None:
                return None
            task.id = int(inserted["id"])
        return await self.get_task(task.user_id, task.id)

    async def get_task(self, user_id: int, task_id: int) -> TorrentTask | None:
        async with self._pool.connection() as conn, conn.cursor(row_factory=dict_row) as cur:
            await cur.execute(
                f"SELECT {_TASK_COLUMNS} FROM torrents t "
                "JOIN users u ON u.id = t.user_id "
                "WHERE u.telegram_user_id = %s AND t.id = %s",
                (user_id, task_id),
            )
            row = await cur.fetchone()
        return _task_from_row(row) if row is not None else None

    async def list_tasks(self, user_id: int) -> list[TorrentTask]:
        async with self._pool.connection() as conn, conn.cursor(row_factory=dict_row) as cur:
            await cur.execute(
                f"SELECT {_TASK_COLUMNS} FROM torrents t "
                "JOIN users u ON u.id = t.user_id "
                "WHERE u.telegram_user_id = %s ORDER BY t.id",
                (user_id,),
            )
            rows = await cur.fetchall()
        return [_task_from_row(row) for row in rows]

    async def update_task(self, task: TorrentTask) -> None:
        async with self._pool.connection() as conn, conn.cursor() as cur:
            await cur.execute(
                """
                    UPDATE torrents SET
                        magnet = %s, name = %s, status = %s, progress = %s,
                        size_bytes = %s, download_speed = %s, upload_speed = %s,
                        eta_seconds = %s, completed_at = %s
                    WHERE id = %s
                    """,
                (
                    task.magnet,
                    task.name,
                    str(task.status),
                    task.progress,
                    task.size_bytes,
                    task.download_speed,
                    task.upload_speed,
                    task.eta_seconds,
                    task.completed_at,
                    task.id,
                ),
            )

    async def delete_task(self, task_id: int) -> None:
        async with self._pool.connection() as conn, conn.cursor() as cur:
            await cur.execute("DELETE FROM torrents WHERE id = %s", (task_id,))

    async def add_event(self, torrent_id: int, event: str, payload: str | None = None) -> None:
        async with self._pool.connection() as conn, conn.cursor() as cur:
            await cur.execute(
                "INSERT INTO torrent_events (torrent_id, event, payload) VALUES (%s, %s, %s)",
                (torrent_id, event, payload),
            )

    async def get_setting(self, user_id: int, key: str) -> str | None:
        async with self._pool.connection() as conn, conn.cursor() as cur:
            await cur.execute(
                "SELECT s.value FROM settings s "
                "JOIN users u ON u.id = s.user_id "
                "WHERE u.telegram_user_id = %s AND s.key = %s",
                (user_id, key),
            )
            row = await cur.fetchone()
        return row[0] if row is not None else None

    async def set_setting(self, user_id: int, key: str, value: str) -> None:
        async with self._pool.connection() as conn, conn.cursor(row_factory=dict_row) as cur:
            user_pk = await self._ensure_user(cur, user_id)
            await cur.execute(
                "INSERT INTO settings (user_id, key, value) VALUES (%s, %s, %s) "
                "ON CONFLICT (user_id, key) DO UPDATE SET value = EXCLUDED.value",
                (user_pk, key, value),
            )

    async def ensure_user(self, telegram_user_id: int) -> None:
        async with self._pool.connection() as conn, conn.cursor(row_factory=dict_row) as cur:
            await self._ensure_user(cur, telegram_user_id)

    async def user_exists(self, telegram_user_id: int) -> bool:
        async with self._pool.connection() as conn, conn.cursor() as cur:
            await cur.execute(
                "SELECT 1 FROM users WHERE telegram_user_id = %s", (telegram_user_id,)
            )
            return await cur.fetchone() is not None
